package network

import (
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/inventory"
	"github.com/opd-ai/violence/pkg/quest"
	"github.com/sirupsen/logrus"
)

const (
	MinCoopPlayers = 2
	MaxCoopPlayers = 4
)

// CoopPlayerState tracks individual player state within a co-op session.
type CoopPlayerState struct {
	PlayerID  uint64
	EntityID  engine.Entity
	Inventory *inventory.Inventory
	Health    float64
	MaxHealth float64
	Armor     float64
	Active    bool // false if disconnected
	PosX      float64
	PosY      float64
	mu        sync.RWMutex
}

// CoopSession manages a 2-4 player cooperative game session.
type CoopSession struct {
	SessionID      string
	Players        map[uint64]*CoopPlayerState
	World          *engine.World
	QuestTracker   *quest.Tracker
	LevelSeed      uint64
	MaxPlayers     int
	mu             sync.RWMutex
	Started        bool
	LevelCompleted bool
	CreatedAt      time.Time
}

// NewCoopSession creates a new co-op session with specified max players (2-4).
func NewCoopSession(sessionID string, maxPlayers int, levelSeed uint64) (*CoopSession, error) {
	if maxPlayers < MinCoopPlayers || maxPlayers > MaxCoopPlayers {
		return nil, fmt.Errorf("invalid max players: %d (must be %d-%d)", maxPlayers, MinCoopPlayers, MaxCoopPlayers)
	}

	return &CoopSession{
		SessionID:    sessionID,
		Players:      make(map[uint64]*CoopPlayerState),
		World:        engine.NewWorld(),
		QuestTracker: quest.NewTracker(),
		LevelSeed:    levelSeed,
		MaxPlayers:   maxPlayers,
		CreatedAt:    time.Now(),
	}, nil
}

// AddPlayer adds a player to the co-op session.
// Returns error if session is full or player already exists.
func (s *CoopSession) AddPlayer(playerID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Players) >= s.MaxPlayers {
		return fmt.Errorf("session full: %d/%d players", len(s.Players), s.MaxPlayers)
	}

	if _, exists := s.Players[playerID]; exists {
		return fmt.Errorf("player %d already in session", playerID)
	}

	// Create player entity in world
	entityID := s.World.AddEntity()

	// Initialize player state with independent inventory
	playerState := &CoopPlayerState{
		PlayerID:  playerID,
		EntityID:  entityID,
		Inventory: inventory.NewInventory(),
		Health:    100.0,
		MaxHealth: 100.0,
		Armor:     0.0,
		Active:    true,
		PosX:      0.0,
		PosY:      0.0,
	}

	s.Players[playerID] = playerState

	logrus.WithFields(logrus.Fields{
		"system_name":  "coop_session",
		"session_id":   s.SessionID,
		"player_id":    playerID,
		"player_count": len(s.Players),
	}).Info("Player joined co-op session")

	return nil
}

// RemovePlayer removes a player from the session and marks them as inactive.
func (s *CoopSession) RemovePlayer(playerID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	playerState, exists := s.Players[playerID]
	if !exists {
		return fmt.Errorf("player %d not in session", playerID)
	}

	// Mark as inactive instead of deleting to preserve state for reconnect
	playerState.mu.Lock()
	playerState.Active = false
	playerState.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"system_name":  "coop_session",
		"session_id":   s.SessionID,
		"player_id":    playerID,
		"active_count": s.getActivePlayerCount(),
	}).Info("Player left co-op session")

	return nil
}

// GetPlayer returns the player state for a given player ID.
func (s *CoopSession) GetPlayer(playerID uint64) (*CoopPlayerState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	playerState, exists := s.Players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %d not found in session", playerID)
	}

	return playerState, nil
}

// GetActivePlayers returns all currently active players.
func (s *CoopSession) GetActivePlayers() []*CoopPlayerState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	active := make([]*CoopPlayerState, 0, len(s.Players))
	for _, p := range s.Players {
		p.mu.RLock()
		if p.Active {
			active = append(active, p)
		}
		p.mu.RUnlock()
	}

	return active
}

// GetPlayerCount returns the total number of players (active + inactive).
func (s *CoopSession) GetPlayerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Players)
}

// getActivePlayerCount returns count of active players (must hold lock).
func (s *CoopSession) getActivePlayerCount() int {
	count := 0
	for _, p := range s.Players {
		p.mu.RLock()
		if p.Active {
			count++
		}
		p.mu.RUnlock()
	}
	return count
}

// UpdateObjectiveProgress updates shared quest progress for all players.
// Co-op sessions share objective completion across all players.
func (s *CoopSession) UpdateObjectiveProgress(objectiveID string, amount int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.QuestTracker.UpdateProgress(objectiveID, amount)

	logrus.WithFields(logrus.Fields{
		"system_name":  "coop_session",
		"session_id":   s.SessionID,
		"objective_id": objectiveID,
		"amount":       amount,
	}).Debug("Objective progress updated")
}

// CompleteObjective marks a shared objective as complete.
func (s *CoopSession) CompleteObjective(objectiveID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.QuestTracker.Complete(objectiveID)

	logrus.WithFields(logrus.Fields{
		"system_name":  "coop_session",
		"session_id":   s.SessionID,
		"objective_id": objectiveID,
	}).Info("Objective completed")
}

// IsLevelComplete checks if all main objectives are complete.
func (s *CoopSession) IsLevelComplete() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	mainObjectives := s.QuestTracker.GetMainObjectives()
	return len(mainObjectives) == 0 // No incomplete main objectives
}

// Start begins the co-op session and generates level objectives.
func (s *CoopSession) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Started {
		return fmt.Errorf("session already started")
	}

	if len(s.Players) < MinCoopPlayers {
		return fmt.Errorf("not enough players: %d (minimum %d required)", len(s.Players), MinCoopPlayers)
	}

	// Generate procedural objectives from level seed
	s.QuestTracker.Generate(s.LevelSeed, 3)
	s.Started = true

	logrus.WithFields(logrus.Fields{
		"system_name":  "coop_session",
		"session_id":   s.SessionID,
		"player_count": len(s.Players),
		"level_seed":   s.LevelSeed,
	}).Info("Co-op session started")

	return nil
}

// CanStart returns true if session has enough players to start.
func (s *CoopSession) CanStart() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Players) >= MinCoopPlayers && len(s.Players) <= s.MaxPlayers
}

// IsFull returns true if session is at max player capacity.
func (s *CoopSession) IsFull() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Players) >= s.MaxPlayers
}

// UpdatePlayerPosition updates a player's position in the level.
func (s *CoopSession) UpdatePlayerPosition(playerID uint64, x, y float64) error {
	s.mu.RLock()
	playerState, exists := s.Players[playerID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("player %d not in session", playerID)
	}

	playerState.mu.Lock()
	playerState.PosX = x
	playerState.PosY = y
	playerState.mu.Unlock()

	return nil
}

// UpdatePlayerHealth updates a player's health value.
func (s *CoopSession) UpdatePlayerHealth(playerID uint64, health float64) error {
	s.mu.RLock()
	playerState, exists := s.Players[playerID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("player %d not in session", playerID)
	}

	playerState.mu.Lock()
	playerState.Health = health
	if playerState.Health < 0 {
		playerState.Health = 0
	}
	playerState.mu.Unlock()

	return nil
}

// SetGenre configures the session world and quest tracker for a genre.
func (s *CoopSession) SetGenre(genreID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.World.SetGenre(genreID)
	s.QuestTracker.SetGenre(genreID)
}

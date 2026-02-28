package network

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

const (
	MinFFAPlayers    = 2
	MaxFFAPlayers    = 8
	DefaultFragLimit = 25
	DefaultTimeLimit = 10 * time.Minute
	RespawnDelay     = 3 * time.Second
)

// FFAPlayerState tracks individual player state within an FFA match.
type FFAPlayerState struct {
	PlayerID    uint64
	EntityID    engine.Entity
	Frags       int
	Deaths      int
	Active      bool
	Dead        bool
	RespawnTime time.Time
	PosX        float64
	PosY        float64
	Health      float64
	MaxHealth   float64
	mu          sync.RWMutex
}

// SpawnPoint represents a player spawn location in the arena.
type SpawnPoint struct {
	X float64
	Y float64
}

// FFAMatch manages a free-for-all deathmatch session.
type FFAMatch struct {
	MatchID     string
	Players     map[uint64]*FFAPlayerState
	World       *engine.World
	FragLimit   int
	TimeLimit   time.Duration
	SpawnPoints []SpawnPoint
	Started     bool
	Finished    bool
	StartTime   time.Time
	WinnerID    uint64
	Seed        uint64
	mu          sync.RWMutex
}

// NewFFAMatch creates a new FFA deathmatch session.
func NewFFAMatch(matchID string, fragLimit int, timeLimit time.Duration, seed uint64) (*FFAMatch, error) {
	if fragLimit <= 0 {
		fragLimit = DefaultFragLimit
	}
	if timeLimit <= 0 {
		timeLimit = DefaultTimeLimit
	}

	logrus.WithFields(logrus.Fields{
		"match_id":   matchID,
		"frag_limit": fragLimit,
		"time_limit": timeLimit,
		"seed":       seed,
	}).Info("Creating FFA match")

	return &FFAMatch{
		MatchID:     matchID,
		Players:     make(map[uint64]*FFAPlayerState),
		World:       engine.NewWorld(),
		FragLimit:   fragLimit,
		TimeLimit:   timeLimit,
		SpawnPoints: make([]SpawnPoint, 0),
		Seed:        seed,
	}, nil
}

// AddPlayer adds a player to the FFA match.
func (m *FFAMatch) AddPlayer(playerID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Started {
		return fmt.Errorf("match already started")
	}

	if len(m.Players) >= MaxFFAPlayers {
		return fmt.Errorf("match is full (max %d players)", MaxFFAPlayers)
	}

	if _, exists := m.Players[playerID]; exists {
		return fmt.Errorf("player %d already in match", playerID)
	}

	player := &FFAPlayerState{
		PlayerID:  playerID,
		Active:    true,
		MaxHealth: 100.0,
		Health:    100.0,
	}

	m.Players[playerID] = player

	logrus.WithFields(logrus.Fields{
		"match_id":     m.MatchID,
		"player_id":    playerID,
		"player_count": len(m.Players),
	}).Info("Player joined FFA match")

	return nil
}

// RemovePlayer removes a player from the match.
func (m *FFAMatch) RemovePlayer(playerID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.Players[playerID]
	if !exists {
		return fmt.Errorf("player %d not in match", playerID)
	}

	player.mu.Lock()
	player.Active = false
	player.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"match_id":  m.MatchID,
		"player_id": playerID,
	}).Info("Player left FFA match")

	return nil
}

// GenerateSpawnPoints creates random spawn points using the match seed.
func (m *FFAMatch) GenerateSpawnPoints(count int, mapWidth, mapHeight float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rng := rand.New(rand.NewSource(int64(m.Seed)))
	m.SpawnPoints = make([]SpawnPoint, 0, count)

	for i := 0; i < count; i++ {
		spawn := SpawnPoint{
			X: rng.Float64() * mapWidth,
			Y: rng.Float64() * mapHeight,
		}
		m.SpawnPoints = append(m.SpawnPoints, spawn)
	}

	logrus.WithFields(logrus.Fields{
		"match_id":    m.MatchID,
		"spawn_count": count,
		"seed":        m.Seed,
	}).Info("Generated spawn points")
}

// GetRandomSpawnPoint returns a random spawn point using the match seed.
func (m *FFAMatch) GetRandomSpawnPoint(playerID uint64) (SpawnPoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.SpawnPoints) == 0 {
		return SpawnPoint{}, fmt.Errorf("no spawn points available")
	}

	// Use playerID + current time as seed for spawn selection
	rng := rand.New(rand.NewSource(int64(m.Seed + playerID + uint64(time.Now().UnixNano()))))
	idx := rng.Intn(len(m.SpawnPoints))

	return m.SpawnPoints[idx], nil
}

// StartMatch begins the FFA match.
func (m *FFAMatch) StartMatch() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Started {
		return fmt.Errorf("match already started")
	}

	if len(m.Players) < MinFFAPlayers {
		return fmt.Errorf("need at least %d players to start", MinFFAPlayers)
	}

	m.Started = true
	m.StartTime = time.Now()

	// Spawn all players at random spawn points
	if len(m.SpawnPoints) > 0 {
		for playerID, player := range m.Players {
			spawnIdx := int(playerID) % len(m.SpawnPoints)
			spawn := m.SpawnPoints[spawnIdx]
			player.mu.Lock()
			player.PosX = spawn.X
			player.PosY = spawn.Y
			player.Health = player.MaxHealth
			player.Dead = false
			player.mu.Unlock()
		}
	}

	logrus.WithFields(logrus.Fields{
		"match_id":     m.MatchID,
		"player_count": len(m.Players),
		"frag_limit":   m.FragLimit,
		"time_limit":   m.TimeLimit,
	}).Info("FFA match started")

	return nil
}

// OnPlayerKill registers a kill by one player of another.
func (m *FFAMatch) OnPlayerKill(killerID, victimID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Finished {
		return fmt.Errorf("match already finished")
	}

	killer, killerExists := m.Players[killerID]
	victim, victimExists := m.Players[victimID]

	if !killerExists {
		return fmt.Errorf("killer %d not in match", killerID)
	}
	if !victimExists {
		return fmt.Errorf("victim %d not in match", victimID)
	}

	// Update killer frags
	killer.mu.Lock()
	killer.Frags++
	currentFrags := killer.Frags
	killer.mu.Unlock()

	// Update victim deaths and mark as dead
	victim.mu.Lock()
	victim.Deaths++
	victim.Dead = true
	victim.RespawnTime = time.Now().Add(RespawnDelay)
	victim.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"match_id":  m.MatchID,
		"killer_id": killerID,
		"victim_id": victimID,
		"frags":     currentFrags,
	}).Info("Player kill registered")

	// Check win condition
	if currentFrags >= m.FragLimit {
		m.Finished = true
		m.WinnerID = killerID
		logrus.WithFields(logrus.Fields{
			"match_id":  m.MatchID,
			"winner_id": killerID,
			"frags":     currentFrags,
		}).Info("FFA match finished - frag limit reached")
	}

	return nil
}

// OnPlayerSuicide registers a player suicide (self-kill).
func (m *FFAMatch) OnPlayerSuicide(playerID uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Finished {
		return fmt.Errorf("match already finished")
	}

	player, exists := m.Players[playerID]
	if !exists {
		return fmt.Errorf("player %d not in match", playerID)
	}

	player.mu.Lock()
	player.Frags-- // Penalty for suicide
	player.Deaths++
	player.Dead = true
	player.RespawnTime = time.Now().Add(RespawnDelay)
	player.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"match_id":  m.MatchID,
		"player_id": playerID,
		"frags":     player.Frags,
	}).Info("Player suicide registered")

	return nil
}

// ProcessRespawns checks for players ready to respawn and respawns them.
func (m *FFAMatch) ProcessRespawns() []uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.Finished {
		return nil
	}

	now := time.Now()
	respawned := make([]uint64, 0)

	for playerID, player := range m.Players {
		player.mu.RLock()
		isDead := player.Dead
		canRespawn := !player.RespawnTime.IsZero() && now.After(player.RespawnTime)
		player.mu.RUnlock()

		if isDead && canRespawn {
			if err := m.RespawnPlayer(playerID); err == nil {
				respawned = append(respawned, playerID)
			}
		}
	}

	return respawned
}

// RespawnPlayer instantly respawns a player at a random spawn point.
func (m *FFAMatch) RespawnPlayer(playerID uint64) error {
	m.mu.RLock()
	player, exists := m.Players[playerID]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("player %d not in match", playerID)
	}

	spawn, err := m.GetRandomSpawnPoint(playerID)
	if err != nil {
		return fmt.Errorf("no spawn point available: %w", err)
	}

	player.mu.Lock()
	player.Dead = false
	player.PosX = spawn.X
	player.PosY = spawn.Y
	player.Health = player.MaxHealth
	player.RespawnTime = time.Time{}
	player.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"match_id":  m.MatchID,
		"player_id": playerID,
		"spawn_x":   spawn.X,
		"spawn_y":   spawn.Y,
	}).Info("Player respawned")

	return nil
}

// CheckTimeLimit checks if the time limit has been reached.
func (m *FFAMatch) CheckTimeLimit() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.Started || m.Finished {
		return false
	}

	elapsed := time.Since(m.StartTime)
	if elapsed >= m.TimeLimit {
		m.mu.RUnlock()
		m.mu.Lock()
		m.Finished = true
		m.WinnerID = m.getLeaderID()
		m.mu.Unlock()
		m.mu.RLock()

		logrus.WithFields(logrus.Fields{
			"match_id":  m.MatchID,
			"winner_id": m.WinnerID,
			"elapsed":   elapsed,
		}).Info("FFA match finished - time limit reached")

		return true
	}

	return false
}

// getLeaderID returns the player ID with the most frags (must be called with lock held).
func (m *FFAMatch) getLeaderID() uint64 {
	var leaderID uint64
	maxFrags := -1000 // Start low to account for negative scores from suicides

	for playerID, player := range m.Players {
		player.mu.RLock()
		frags := player.Frags
		player.mu.RUnlock()

		if frags > maxFrags {
			maxFrags = frags
			leaderID = playerID
		}
	}

	return leaderID
}

// GetPlayerStats returns the current stats for a player.
func (m *FFAMatch) GetPlayerStats(playerID uint64) (frags, deaths int, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.Players[playerID]
	if !exists {
		return 0, 0, fmt.Errorf("player %d not in match", playerID)
	}

	player.mu.RLock()
	defer player.mu.RUnlock()

	return player.Frags, player.Deaths, nil
}

// PlayerStats represents a player's statistics for the leaderboard.
type PlayerStats struct {
	PlayerID uint64
	Frags    int
	Deaths   int
}

// GetLeaderboard returns all players sorted by frags (descending).
func (m *FFAMatch) GetLeaderboard() []PlayerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	leaderboard := make([]PlayerStats, 0, len(m.Players))

	for playerID, player := range m.Players {
		player.mu.RLock()
		stats := PlayerStats{
			PlayerID: playerID,
			Frags:    player.Frags,
			Deaths:   player.Deaths,
		}
		player.mu.RUnlock()
		leaderboard = append(leaderboard, stats)
	}

	// Simple bubble sort by frags (descending)
	for i := 0; i < len(leaderboard); i++ {
		for j := i + 1; j < len(leaderboard); j++ {
			if leaderboard[j].Frags > leaderboard[i].Frags {
				leaderboard[i], leaderboard[j] = leaderboard[j], leaderboard[i]
			}
		}
	}

	return leaderboard
}

// IsFinished returns whether the match has ended.
func (m *FFAMatch) IsFinished() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Finished
}

// GetWinner returns the winner's player ID (0 if match not finished).
func (m *FFAMatch) GetWinner() uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.WinnerID
}

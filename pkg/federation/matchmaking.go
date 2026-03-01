// Package federation provides server federation and matchmaking.
package federation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// GameMode represents a game mode type.
type GameMode string

const (
	ModeCoop      GameMode = "coop"
	ModeFFA       GameMode = "ffa"
	ModeTeamDM    GameMode = "team-dm"
	ModeTerritory GameMode = "territory"
)

// QueueEntry represents a player in the matchmaking queue.
type QueueEntry struct {
	PlayerID  string
	Mode      GameMode
	Genre     string
	Region    Region
	EnqueueAt time.Time
}

// MatchResult represents the result of successful matchmaking.
type MatchResult struct {
	PlayerIDs     []string
	ServerAddress string
	ServerName    string
	Mode          GameMode
	Genre         string
}

// Matchmaker manages matchmaking queues and player grouping.
type Matchmaker struct {
	hub               *FederationHub
	queues            map[GameMode][]*QueueEntry
	mu                sync.Mutex
	ctx               context.Context
	cancel            context.CancelFunc
	matchInterval     time.Duration
	queueTimeout      time.Duration
	minPlayersPerMode map[GameMode]int
	maxPlayersPerMode map[GameMode]int
}

// NewMatchmaker creates a new matchmaker instance.
func NewMatchmaker(hub *FederationHub) *Matchmaker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Matchmaker{
		hub:           hub,
		queues:        make(map[GameMode][]*QueueEntry),
		ctx:           ctx,
		cancel:        cancel,
		matchInterval: 2 * time.Second,
		queueTimeout:  60 * time.Second,
		minPlayersPerMode: map[GameMode]int{
			ModeCoop:      2,
			ModeFFA:       2,
			ModeTeamDM:    2,
			ModeTerritory: 2,
		},
		maxPlayersPerMode: map[GameMode]int{
			ModeCoop:      4,
			ModeFFA:       8,
			ModeTeamDM:    16,
			ModeTerritory: 16,
		},
	}
}

// Start begins the matchmaking loop.
func (m *Matchmaker) Start() {
	go m.matchLoop()
}

// Stop halts the matchmaker.
func (m *Matchmaker) Stop() {
	m.cancel()
}

// Enqueue adds a player to the matchmaking queue.
func (m *Matchmaker) Enqueue(playerID string, mode GameMode, genre string, region Region) error {
	if playerID == "" {
		return fmt.Errorf("playerID is required")
	}
	if mode == "" {
		return fmt.Errorf("mode is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if player is already in queue
	for _, queue := range m.queues {
		for _, entry := range queue {
			if entry.PlayerID == playerID {
				return fmt.Errorf("player already in queue")
			}
		}
	}

	entry := &QueueEntry{
		PlayerID:  playerID,
		Mode:      mode,
		Genre:     genre,
		Region:    region,
		EnqueueAt: time.Now(),
	}

	m.queues[mode] = append(m.queues[mode], entry)

	logrus.WithFields(logrus.Fields{
		"player_id": playerID,
		"mode":      mode,
		"genre":     genre,
		"region":    region,
	}).Debug("player enqueued for matchmaking")

	return nil
}

// Dequeue removes a player from all queues.
func (m *Matchmaker) Dequeue(playerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for mode, queue := range m.queues {
		filtered := make([]*QueueEntry, 0, len(queue))
		for _, entry := range queue {
			if entry.PlayerID != playerID {
				filtered = append(filtered, entry)
			}
		}
		m.queues[mode] = filtered
	}
}

// GetQueueSize returns the number of players in queue for a mode.
func (m *Matchmaker) GetQueueSize(mode GameMode) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.queues[mode])
}

// matchLoop periodically attempts to create matches from queued players.
func (m *Matchmaker) matchLoop() {
	ticker := time.NewTicker(m.matchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.processMatches()
			m.cleanupExpiredEntries()
		}
	}
}

// processMatches attempts to create matches from queued players.
func (m *Matchmaker) processMatches() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for mode, queue := range m.queues {
		m.processModeQueue(mode, queue)
	}
}

// processModeQueue processes matches for a specific game mode.
func (m *Matchmaker) processModeQueue(mode GameMode, queue []*QueueEntry) {
	minPlayers := m.minPlayersPerMode[mode]
	maxPlayers := m.maxPlayersPerMode[mode]

	if len(queue) < minPlayers {
		return
	}

	groups := m.groupPlayers(queue)

	for key, group := range groups {
		m.processPlayerGroup(mode, key, group, minPlayers, maxPlayers)
	}
}

// processPlayerGroup attempts to create a match from a player group.
func (m *Matchmaker) processPlayerGroup(mode GameMode, key groupKey, group []*QueueEntry, minPlayers, maxPlayers int) {
	if len(group) < minPlayers {
		return
	}

	matchSize := len(group)
	if matchSize > maxPlayers {
		matchSize = maxPlayers
	}

	matchPlayers := group[:matchSize]

	server := m.findServer(mode, key.genre, key.region, matchSize)
	if server == nil {
		m.logNoServerAvailable(mode, key, matchSize)
		return
	}

	m.finalizeMatch(mode, matchPlayers, server)
}

// logNoServerAvailable logs when no server is available for a match.
func (m *Matchmaker) logNoServerAvailable(mode GameMode, key groupKey, matchSize int) {
	logrus.WithFields(logrus.Fields{
		"mode":   mode,
		"genre":  key.genre,
		"region": key.region,
		"size":   matchSize,
	}).Debug("no available server for match")
}

// finalizeMatch creates a match and logs the result.
func (m *Matchmaker) finalizeMatch(mode GameMode, matchPlayers []*QueueEntry, server *ServerAnnouncement) {
	result := m.createMatch(matchPlayers, server, mode)
	if result != nil {
		m.removeMatchedPlayers(mode, result.PlayerIDs)

		logrus.WithFields(logrus.Fields{
			"mode":        mode,
			"players":     len(result.PlayerIDs),
			"server":      result.ServerName,
			"server_addr": result.ServerAddress,
			"genre":       result.Genre,
		}).Info("match created")
	}
}

// groupKey identifies a player group by genre and region.
type groupKey struct {
	genre  string
	region Region
}

// groupPlayers groups queue entries by genre and region.
func (m *Matchmaker) groupPlayers(queue []*QueueEntry) map[groupKey][]*QueueEntry {
	groups := make(map[groupKey][]*QueueEntry)
	for _, entry := range queue {
		key := groupKey{
			genre:  entry.Genre,
			region: entry.Region,
		}
		groups[key] = append(groups[key], entry)
	}
	return groups
}

// findServer locates an available server for the match.
func (m *Matchmaker) findServer(mode GameMode, genre string, region Region, playerCount int) *ServerAnnouncement {
	query := &ServerQuery{
		Genre:  &genre,
		Region: &region,
	}

	servers := m.hub.queryServers(query)

	// Find server with enough capacity
	for _, server := range servers {
		availableSlots := server.MaxPlayers - server.Players
		if availableSlots >= playerCount {
			return server
		}
	}

	return nil
}

// createMatch creates a match result from players and server.
func (m *Matchmaker) createMatch(players []*QueueEntry, server *ServerAnnouncement, mode GameMode) *MatchResult {
	playerIDs := make([]string, len(players))
	for i, player := range players {
		playerIDs[i] = player.PlayerID
	}

	return &MatchResult{
		PlayerIDs:     playerIDs,
		ServerAddress: server.Address,
		ServerName:    server.Name,
		Mode:          mode,
		Genre:         server.Genre,
	}
}

// removeMatchedPlayers removes matched players from queue.
func (m *Matchmaker) removeMatchedPlayers(mode GameMode, playerIDs []string) {
	queue := m.queues[mode]
	filtered := make([]*QueueEntry, 0, len(queue))

	matchedSet := make(map[string]bool)
	for _, id := range playerIDs {
		matchedSet[id] = true
	}

	for _, entry := range queue {
		if !matchedSet[entry.PlayerID] {
			filtered = append(filtered, entry)
		}
	}

	m.queues[mode] = filtered
}

// cleanupExpiredEntries removes queue entries that have timed out.
func (m *Matchmaker) cleanupExpiredEntries() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for mode, queue := range m.queues {
		filtered := make([]*QueueEntry, 0, len(queue))
		for _, entry := range queue {
			if now.Sub(entry.EnqueueAt) <= m.queueTimeout {
				filtered = append(filtered, entry)
			} else {
				logrus.WithFields(logrus.Fields{
					"player_id": entry.PlayerID,
					"mode":      mode,
					"wait_time": now.Sub(entry.EnqueueAt),
				}).Debug("queue entry expired")
			}
		}
		m.queues[mode] = filtered
	}
}

// GetQueuedPlayers returns all players in queue for a mode.
func (m *Matchmaker) GetQueuedPlayers(mode GameMode) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue := m.queues[mode]
	playerIDs := make([]string, len(queue))
	for i, entry := range queue {
		playerIDs[i] = entry.PlayerID
	}
	return playerIDs
}

// IsQueued checks if a player is currently in any queue.
func (m *Matchmaker) IsQueued(playerID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, queue := range m.queues {
		for _, entry := range queue {
			if entry.PlayerID == playerID {
				return true
			}
		}
	}
	return false
}

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
	MinTeamPlayers       = 2
	MaxTeamPlayers       = 16
	DefaultTeamFragLimit = 50
	TeamRed              = 0
	TeamBlue             = 1
)

// TeamPlayerState tracks individual player state within a team match.
type TeamPlayerState struct {
	PlayerID    uint64
	EntityID    engine.Entity
	Team        int
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

// TeamScore tracks score and stats for a team.
type TeamScore struct {
	Team   int
	Frags  int
	Deaths int
	mu     sync.RWMutex
}

// TeamMatch manages a team deathmatch session.
type TeamMatch struct {
	MatchID     string
	Players     map[uint64]*TeamPlayerState
	Teams       map[int]*TeamScore
	World       *engine.World
	FragLimit   int
	TimeLimit   time.Duration
	SpawnPoints map[int][]SpawnPoint
	Started     bool
	Finished    bool
	StartTime   time.Time
	WinnerTeam  int
	Seed        uint64
	mu          sync.RWMutex
}

// NewTeamMatch creates a new team deathmatch session.
func NewTeamMatch(matchID string, fragLimit int, timeLimit time.Duration, seed uint64) (*TeamMatch, error) {
	if fragLimit <= 0 {
		fragLimit = DefaultTeamFragLimit
	}
	if timeLimit <= 0 {
		timeLimit = DefaultTimeLimit
	}

	logrus.WithFields(logrus.Fields{
		"match_id":   matchID,
		"frag_limit": fragLimit,
		"time_limit": timeLimit,
		"seed":       seed,
	}).Info("Creating team match")

	return &TeamMatch{
		MatchID: matchID,
		Players: make(map[uint64]*TeamPlayerState),
		Teams: map[int]*TeamScore{
			TeamRed:  {Team: TeamRed},
			TeamBlue: {Team: TeamBlue},
		},
		World:     engine.NewWorld(),
		FragLimit: fragLimit,
		TimeLimit: timeLimit,
		SpawnPoints: map[int][]SpawnPoint{
			TeamRed:  make([]SpawnPoint, 0),
			TeamBlue: make([]SpawnPoint, 0),
		},
		Seed:       seed,
		WinnerTeam: -1,
	}, nil
}

// AddPlayer adds a player to the team match.
func (m *TeamMatch) AddPlayer(playerID uint64, team int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Started {
		return fmt.Errorf("match already started")
	}

	if len(m.Players) >= MaxTeamPlayers {
		return fmt.Errorf("match is full (max %d players)", MaxTeamPlayers)
	}

	if team != TeamRed && team != TeamBlue {
		return fmt.Errorf("invalid team %d (must be %d or %d)", team, TeamRed, TeamBlue)
	}

	if _, exists := m.Players[playerID]; exists {
		return fmt.Errorf("player %d already in match", playerID)
	}

	player := &TeamPlayerState{
		PlayerID:  playerID,
		Team:      team,
		Active:    true,
		MaxHealth: 100.0,
		Health:    100.0,
	}

	m.Players[playerID] = player

	logrus.WithFields(logrus.Fields{
		"match_id":     m.MatchID,
		"player_id":    playerID,
		"team":         team,
		"player_count": len(m.Players),
	}).Info("Player joined team match")

	return nil
}

// RemovePlayer removes a player from the match.
func (m *TeamMatch) RemovePlayer(playerID uint64) error {
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
	}).Info("Player left team match")

	return nil
}

// GenerateSpawnPoints creates random spawn points for both teams using the match seed.
func (m *TeamMatch) GenerateSpawnPoints(countPerTeam int, mapWidth, mapHeight float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	rng := rand.New(rand.NewSource(int64(m.Seed)))

	// Red team spawns on left side
	m.SpawnPoints[TeamRed] = make([]SpawnPoint, 0, countPerTeam)
	for i := 0; i < countPerTeam; i++ {
		spawn := SpawnPoint{
			X: rng.Float64() * (mapWidth * 0.3),
			Y: rng.Float64() * mapHeight,
		}
		m.SpawnPoints[TeamRed] = append(m.SpawnPoints[TeamRed], spawn)
	}

	// Blue team spawns on right side
	m.SpawnPoints[TeamBlue] = make([]SpawnPoint, 0, countPerTeam)
	for i := 0; i < countPerTeam; i++ {
		spawn := SpawnPoint{
			X: (mapWidth * 0.7) + (rng.Float64() * (mapWidth * 0.3)),
			Y: rng.Float64() * mapHeight,
		}
		m.SpawnPoints[TeamBlue] = append(m.SpawnPoints[TeamBlue], spawn)
	}

	logrus.WithFields(logrus.Fields{
		"match_id":    m.MatchID,
		"spawn_count": countPerTeam,
		"seed":        m.Seed,
	}).Info("Generated team spawn points")
}

// GetRandomSpawnPoint returns a random spawn point for the player's team.
func (m *TeamMatch) GetRandomSpawnPoint(playerID uint64) (SpawnPoint, error) {
	m.mu.RLock()
	player, exists := m.Players[playerID]
	m.mu.RUnlock()

	if !exists {
		return SpawnPoint{}, fmt.Errorf("player %d not in match", playerID)
	}

	player.mu.RLock()
	team := player.Team
	player.mu.RUnlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	spawnPoints := m.SpawnPoints[team]
	if len(spawnPoints) == 0 {
		return SpawnPoint{}, fmt.Errorf("no spawn points for team %d", team)
	}

	rng := rand.New(rand.NewSource(int64(m.Seed + playerID + uint64(time.Now().UnixNano()))))
	idx := rng.Intn(len(spawnPoints))

	return spawnPoints[idx], nil
}

// StartMatch begins the team match.
func (m *TeamMatch) StartMatch() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Started {
		return fmt.Errorf("match already started")
	}

	if len(m.Players) < MinTeamPlayers {
		return fmt.Errorf("need at least %d players to start", MinTeamPlayers)
	}

	m.Started = true
	m.StartTime = time.Now()

	// Spawn all players at their team's spawn points
	for playerID, player := range m.Players {
		player.mu.RLock()
		team := player.Team
		player.mu.RUnlock()

		spawnPoints := m.SpawnPoints[team]
		if len(spawnPoints) > 0 {
			spawnIdx := int(playerID) % len(spawnPoints)
			spawn := spawnPoints[spawnIdx]
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
	}).Info("Team match started")

	return nil
}

// OnPlayerKill registers a kill by one player of another.
func (m *TeamMatch) OnPlayerKill(killerID, victimID uint64) error {
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

	killer.mu.Lock()
	killerTeam := killer.Team
	killer.Frags++
	killer.mu.Unlock()

	victim.mu.Lock()
	victimTeam := victim.Team
	victim.Deaths++
	victim.Dead = true
	victim.RespawnTime = time.Now().Add(RespawnDelay)
	victim.mu.Unlock()

	// Update team scores
	killerTeamScore := m.Teams[killerTeam]
	victimTeamScore := m.Teams[victimTeam]

	killerTeamScore.mu.Lock()
	killerTeamScore.Frags++
	currentTeamFrags := killerTeamScore.Frags
	killerTeamScore.mu.Unlock()

	victimTeamScore.mu.Lock()
	victimTeamScore.Deaths++
	victimTeamScore.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"match_id":    m.MatchID,
		"killer_id":   killerID,
		"killer_team": killerTeam,
		"victim_id":   victimID,
		"victim_team": victimTeam,
		"team_frags":  currentTeamFrags,
	}).Info("Player kill registered")

	// Check win condition
	if currentTeamFrags >= m.FragLimit {
		m.Finished = true
		m.WinnerTeam = killerTeam
		logrus.WithFields(logrus.Fields{
			"match_id":    m.MatchID,
			"winner_team": killerTeam,
			"frags":       currentTeamFrags,
		}).Info("Team match finished - frag limit reached")
	}

	return nil
}

// OnPlayerSuicide registers a player suicide (self-kill).
func (m *TeamMatch) OnPlayerSuicide(playerID uint64) error {
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
	team := player.Team
	player.Frags--
	player.Deaths++
	player.Dead = true
	player.RespawnTime = time.Now().Add(RespawnDelay)
	player.mu.Unlock()

	// Update team deaths
	teamScore := m.Teams[team]
	teamScore.mu.Lock()
	teamScore.Deaths++
	teamScore.mu.Unlock()

	logrus.WithFields(logrus.Fields{
		"match_id":  m.MatchID,
		"player_id": playerID,
		"team":      team,
		"frags":     player.Frags,
	}).Info("Player suicide registered")

	return nil
}

// ProcessRespawns checks for players ready to respawn and respawns them.
func (m *TeamMatch) ProcessRespawns() []uint64 {
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

// RespawnPlayer instantly respawns a player at a random team spawn point.
func (m *TeamMatch) RespawnPlayer(playerID uint64) error {
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
func (m *TeamMatch) CheckTimeLimit() bool {
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
		m.WinnerTeam = m.getLeadingTeam()
		m.mu.Unlock()
		m.mu.RLock()

		logrus.WithFields(logrus.Fields{
			"match_id":    m.MatchID,
			"winner_team": m.WinnerTeam,
			"elapsed":     elapsed,
		}).Info("Team match finished - time limit reached")

		return true
	}

	return false
}

// getLeadingTeam returns the team with the most frags (must be called with lock held).
func (m *TeamMatch) getLeadingTeam() int {
	redScore := m.Teams[TeamRed]
	blueScore := m.Teams[TeamBlue]

	redScore.mu.RLock()
	redFrags := redScore.Frags
	redScore.mu.RUnlock()

	blueScore.mu.RLock()
	blueFrags := blueScore.Frags
	blueScore.mu.RUnlock()

	if redFrags > blueFrags {
		return TeamRed
	}
	return TeamBlue
}

// GetPlayerStats returns the current stats for a player.
func (m *TeamMatch) GetPlayerStats(playerID uint64) (team, frags, deaths int, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.Players[playerID]
	if !exists {
		return 0, 0, 0, fmt.Errorf("player %d not in match", playerID)
	}

	player.mu.RLock()
	defer player.mu.RUnlock()

	return player.Team, player.Frags, player.Deaths, nil
}

// GetTeamScore returns the current score for a team.
func (m *TeamMatch) GetTeamScore(team int) (frags, deaths int, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	teamScore, exists := m.Teams[team]
	if !exists {
		return 0, 0, fmt.Errorf("invalid team %d", team)
	}

	teamScore.mu.RLock()
	defer teamScore.mu.RUnlock()

	return teamScore.Frags, teamScore.Deaths, nil
}

// TeamPlayerStats represents a player's statistics for the leaderboard.
type TeamPlayerStats struct {
	PlayerID uint64
	Team     int
	Frags    int
	Deaths   int
}

// GetLeaderboard returns all players sorted by team and then by frags (descending).
func (m *TeamMatch) GetLeaderboard() []TeamPlayerStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	leaderboard := make([]TeamPlayerStats, 0, len(m.Players))

	for playerID, player := range m.Players {
		player.mu.RLock()
		stats := TeamPlayerStats{
			PlayerID: playerID,
			Team:     player.Team,
			Frags:    player.Frags,
			Deaths:   player.Deaths,
		}
		player.mu.RUnlock()
		leaderboard = append(leaderboard, stats)
	}

	// Sort by team first, then by frags (descending)
	for i := 0; i < len(leaderboard); i++ {
		for j := i + 1; j < len(leaderboard); j++ {
			if leaderboard[j].Team < leaderboard[i].Team ||
				(leaderboard[j].Team == leaderboard[i].Team && leaderboard[j].Frags > leaderboard[i].Frags) {
				leaderboard[i], leaderboard[j] = leaderboard[j], leaderboard[i]
			}
		}
	}

	return leaderboard
}

// IsFinished returns whether the match has ended.
func (m *TeamMatch) IsFinished() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Finished
}

// GetWinner returns the winning team (-1 if match not finished).
func (m *TeamMatch) GetWinner() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.WinnerTeam
}

package network

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

const (
	// CaptureRadius is the distance within which players can capture a control point
	CaptureRadius = 5.0

	// CaptureRatePerTick is how much capture progress changes per tick when contested
	CaptureRatePerTick = 0.05 // 5% per tick = 1 second at 20 ticks/sec

	// DefaultScoreTickRate is how often teams score points for held control points
	DefaultScoreTickRate = time.Second

	// DefaultPointsPerTick is points earned per control point per tick
	DefaultPointsPerTick = 1
)

// ControlPointOwnership represents which team owns a control point
type ControlPointOwnership int

const (
	OwnershipNeutral ControlPointOwnership = -1
	OwnershipRed     ControlPointOwnership = TeamRed
	OwnershipBlue    ControlPointOwnership = TeamBlue
)

// ControlPoint represents a capturable objective in territory control mode.
type ControlPoint struct {
	ID              string
	EntityID        engine.Entity
	PosX            float64
	PosY            float64
	Owner           ControlPointOwnership
	CaptureProgress float64 // -1.0 (full red) to +1.0 (full blue), 0.0 is neutral
	LastTickTime    time.Time
	VisualStyle     string // Genre-specific visual style (altar/terminal/etc)
	mu              sync.RWMutex
}

// NewControlPoint creates a new control point at the specified position.
func NewControlPoint(id string, entityID engine.Entity, x, y float64) *ControlPoint {
	logrus.WithFields(logrus.Fields{
		"control_point_id": id,
		"entity_id":        entityID,
		"pos_x":            x,
		"pos_y":            y,
	}).Debug("Creating control point")

	return &ControlPoint{
		ID:              id,
		EntityID:        entityID,
		PosX:            x,
		PosY:            y,
		Owner:           OwnershipNeutral,
		CaptureProgress: 0.0,
		LastTickTime:    time.Now(),
		VisualStyle:     "generic", // Default style
	}
}

// GetOwner returns the current owner of the control point.
func (cp *ControlPoint) GetOwner() ControlPointOwnership {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.Owner
}

// GetCaptureProgress returns the current capture progress (-1.0 to +1.0).
func (cp *ControlPoint) GetCaptureProgress() float64 {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.CaptureProgress
}

// UpdateCapture processes capture logic based on nearby players.
// Returns true if ownership changed.
func (cp *ControlPoint) UpdateCapture(redCount, blueCount int) bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	oldOwner := cp.Owner

	// Determine capture direction based on player counts
	if redCount > blueCount {
		// Red team is capturing
		cp.CaptureProgress -= CaptureRatePerTick * float64(redCount-blueCount)
	} else if blueCount > redCount {
		// Blue team is capturing
		cp.CaptureProgress += CaptureRatePerTick * float64(blueCount-redCount)
	}
	// If counts are equal (contested), progress doesn't change

	// Clamp progress to [-1.0, +1.0]
	if cp.CaptureProgress < -1.0 {
		cp.CaptureProgress = -1.0
	} else if cp.CaptureProgress > 1.0 {
		cp.CaptureProgress = 1.0
	}

	// Update ownership based on progress thresholds
	// Use clear thresholds: owned at ±0.9, neutral between -0.5 and +0.5
	var newOwner ControlPointOwnership
	if cp.CaptureProgress <= -0.9 {
		newOwner = OwnershipRed
	} else if cp.CaptureProgress >= 0.9 {
		newOwner = OwnershipBlue
	} else {
		newOwner = OwnershipNeutral
	}

	if newOwner != oldOwner {
		cp.Owner = newOwner
		logrus.WithFields(logrus.Fields{
			"control_point_id": cp.ID,
			"old_owner":        oldOwner,
			"new_owner":        newOwner,
			"progress":         cp.CaptureProgress,
		}).Info("Control point ownership changed")
		return true
	}

	return false
}

// IsPlayerInRange checks if a player at (x, y) is within capture range.
func (cp *ControlPoint) IsPlayerInRange(x, y float64) bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	dx := cp.PosX - x
	dy := cp.PosY - y
	distance := math.Sqrt(dx*dx + dy*dy)

	return distance <= CaptureRadius
}

// SetVisualStyle sets the genre-specific visual style for the control point.
// Valid styles: "altar" (fantasy), "terminal" (scifi), "summoning-circle" (horror),
// "server-rack" (cyberpunk), "scrap-pile" (postapoc), "generic" (default)
func (cp *ControlPoint) SetVisualStyle(style string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.VisualStyle = style

	logrus.WithFields(logrus.Fields{
		"control_point_id": cp.ID,
		"visual_style":     style,
	}).Debug("Control point visual style updated")
}

// GetVisualStyle returns the current visual style of the control point.
func (cp *ControlPoint) GetVisualStyle() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.VisualStyle
}

// TerritoryMatch manages a territory control match.
type TerritoryMatch struct {
	MatchID       string
	Players       map[uint64]*TeamPlayerState
	Teams         map[int]*TeamScore
	ControlPoints map[string]*ControlPoint
	World         *engine.World
	ScoreLimit    int
	TimeLimit     time.Duration
	ScoreTickRate time.Duration
	PointsPerTick int
	SpawnPoints   map[int][]SpawnPoint
	Started       bool
	Finished      bool
	StartTime     time.Time
	LastScoreTick time.Time
	WinnerTeam    int
	Seed          uint64
	Genre         string // Current genre for control point visuals
	mu            sync.RWMutex
}

// NewTerritoryMatch creates a new territory control match.
func NewTerritoryMatch(matchID string, scoreLimit int, timeLimit time.Duration, seed uint64) (*TerritoryMatch, error) {
	if scoreLimit <= 0 {
		scoreLimit = 100 // Default: first to 100 points
	}
	if timeLimit <= 0 {
		timeLimit = DefaultTimeLimit
	}

	logrus.WithFields(logrus.Fields{
		"match_id":    matchID,
		"score_limit": scoreLimit,
		"time_limit":  timeLimit,
		"seed":        seed,
	}).Info("Creating territory match")

	return &TerritoryMatch{
		MatchID: matchID,
		Players: make(map[uint64]*TeamPlayerState),
		Teams: map[int]*TeamScore{
			TeamRed:  {Team: TeamRed},
			TeamBlue: {Team: TeamBlue},
		},
		ControlPoints: make(map[string]*ControlPoint),
		World:         engine.NewWorld(),
		ScoreLimit:    scoreLimit,
		TimeLimit:     timeLimit,
		ScoreTickRate: DefaultScoreTickRate,
		PointsPerTick: DefaultPointsPerTick,
		SpawnPoints: map[int][]SpawnPoint{
			TeamRed:  make([]SpawnPoint, 0),
			TeamBlue: make([]SpawnPoint, 0),
		},
		Seed:          seed,
		WinnerTeam:    -1,
		LastScoreTick: time.Now(),
	}, nil
}

// AddControlPoint adds a control point to the match.
func (m *TerritoryMatch) AddControlPoint(id string, x, y float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.ControlPoints[id]; exists {
		return fmt.Errorf("control point %s already exists", id)
	}

	entityID := m.World.AddEntity()
	cp := NewControlPoint(id, entityID, x, y)
	m.ControlPoints[id] = cp

	logrus.WithFields(logrus.Fields{
		"match_id":         m.MatchID,
		"control_point_id": id,
		"entity_id":        entityID,
		"pos_x":            x,
		"pos_y":            y,
	}).Info("Added control point to territory match")

	return nil
}

// AddPlayer adds a player to the territory match.
func (m *TerritoryMatch) AddPlayer(playerID uint64, team int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if team != TeamRed && team != TeamBlue {
		return fmt.Errorf("invalid team: %d", team)
	}

	if len(m.Players) >= MaxTeamPlayers {
		return fmt.Errorf("match is full")
	}

	if _, exists := m.Players[playerID]; exists {
		return fmt.Errorf("player %d already in match", playerID)
	}

	entityID := m.World.AddEntity()

	player := &TeamPlayerState{
		PlayerID:  playerID,
		EntityID:  entityID,
		Team:      team,
		Active:    true,
		Health:    100.0,
		MaxHealth: 100.0,
	}

	m.Players[playerID] = player

	logrus.WithFields(logrus.Fields{
		"match_id":  m.MatchID,
		"player_id": playerID,
		"team":      team,
		"entity_id": entityID,
	}).Info("Player added to territory match")

	return nil
}

// Start begins the territory match.
func (m *TerritoryMatch) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Started {
		return fmt.Errorf("match already started")
	}

	if len(m.Players) < MinTeamPlayers {
		return fmt.Errorf("need at least %d players to start", MinTeamPlayers)
	}

	if len(m.ControlPoints) == 0 {
		return fmt.Errorf("need at least one control point")
	}

	m.Started = true
	m.StartTime = time.Now()
	m.LastScoreTick = m.StartTime

	logrus.WithFields(logrus.Fields{
		"match_id":       m.MatchID,
		"player_count":   len(m.Players),
		"control_points": len(m.ControlPoints),
	}).Info("Territory match started")

	return nil
}

// ProcessCapture updates capture progress for all control points.
func (m *TerritoryMatch) ProcessCapture() {
	controlPoints, players := m.getMatchSnapshot()

	for _, cp := range controlPoints {
		redCount, blueCount := m.countPlayersNearPoint(cp, players)
		cp.UpdateCapture(redCount, blueCount)
	}
}

// getMatchSnapshot creates a snapshot of control points and players for processing.
func (m *TerritoryMatch) getMatchSnapshot() ([]*ControlPoint, []*TeamPlayerState) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	controlPoints := make([]*ControlPoint, 0, len(m.ControlPoints))
	for _, cp := range m.ControlPoints {
		controlPoints = append(controlPoints, cp)
	}
	players := make([]*TeamPlayerState, 0, len(m.Players))
	for _, p := range m.Players {
		players = append(players, p)
	}

	return controlPoints, players
}

// countPlayersNearPoint counts active players from each team near a control point.
func (m *TerritoryMatch) countPlayersNearPoint(cp *ControlPoint, players []*TeamPlayerState) (int, int) {
	redCount := 0
	blueCount := 0

	for _, p := range players {
		p.mu.RLock()
		active := p.Active && !p.Dead
		team := p.Team
		posX := p.PosX
		posY := p.PosY
		p.mu.RUnlock()

		if !active {
			continue
		}

		if cp.IsPlayerInRange(posX, posY) {
			if team == TeamRed {
				redCount++
			} else if team == TeamBlue {
				blueCount++
			}
		}
	}

	return redCount, blueCount
}

// ProcessScoring awards points to teams based on controlled points.
func (m *TerritoryMatch) ProcessScoring() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if now.Sub(m.LastScoreTick) < m.ScoreTickRate {
		return
	}

	m.LastScoreTick = now

	// Count controlled points per team
	redPoints := 0
	bluePoints := 0

	for _, cp := range m.ControlPoints {
		owner := cp.GetOwner()
		if owner == OwnershipRed {
			redPoints++
		} else if owner == OwnershipBlue {
			bluePoints++
		}
	}

	// Award points
	if redPoints > 0 {
		m.Teams[TeamRed].mu.Lock()
		m.Teams[TeamRed].Frags += redPoints * m.PointsPerTick
		newScore := m.Teams[TeamRed].Frags
		m.Teams[TeamRed].mu.Unlock()

		logrus.WithFields(logrus.Fields{
			"match_id":       m.MatchID,
			"team":           "red",
			"points_awarded": redPoints * m.PointsPerTick,
			"controlled_cps": redPoints,
			"total_score":    newScore,
		}).Debug("Awarded points to red team")
	}

	if bluePoints > 0 {
		m.Teams[TeamBlue].mu.Lock()
		m.Teams[TeamBlue].Frags += bluePoints * m.PointsPerTick
		newScore := m.Teams[TeamBlue].Frags
		m.Teams[TeamBlue].mu.Unlock()

		logrus.WithFields(logrus.Fields{
			"match_id":       m.MatchID,
			"team":           "blue",
			"points_awarded": bluePoints * m.PointsPerTick,
			"controlled_cps": bluePoints,
			"total_score":    newScore,
		}).Debug("Awarded points to blue team")
	}
}

// CheckWinCondition checks if a team has won.
func (m *TerritoryMatch) CheckWinCondition() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Finished {
		return true
	}

	// Check score limit
	m.Teams[TeamRed].mu.RLock()
	redScore := m.Teams[TeamRed].Frags
	m.Teams[TeamRed].mu.RUnlock()

	m.Teams[TeamBlue].mu.RLock()
	blueScore := m.Teams[TeamBlue].Frags
	m.Teams[TeamBlue].mu.RUnlock()

	if redScore >= m.ScoreLimit {
		m.Finished = true
		m.WinnerTeam = TeamRed
		logrus.WithFields(logrus.Fields{
			"match_id":    m.MatchID,
			"winner":      "red",
			"final_score": redScore,
		}).Info("Territory match ended - red team wins by score limit")
		return true
	}

	if blueScore >= m.ScoreLimit {
		m.Finished = true
		m.WinnerTeam = TeamBlue
		logrus.WithFields(logrus.Fields{
			"match_id":    m.MatchID,
			"winner":      "blue",
			"final_score": blueScore,
		}).Info("Territory match ended - blue team wins by score limit")
		return true
	}

	// Check time limit
	if m.TimeLimit > 0 && time.Since(m.StartTime) >= m.TimeLimit {
		m.Finished = true
		if redScore > blueScore {
			m.WinnerTeam = TeamRed
		} else if blueScore > redScore {
			m.WinnerTeam = TeamBlue
		} else {
			m.WinnerTeam = -1 // Tie
		}
		logrus.WithFields(logrus.Fields{
			"match_id":   m.MatchID,
			"winner":     m.WinnerTeam,
			"red_score":  redScore,
			"blue_score": blueScore,
		}).Info("Territory match ended - time limit reached")
		return true
	}

	return false
}

// GetTeamScore returns the score for a team.
func (m *TerritoryMatch) GetTeamScore(team int) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ts, exists := m.Teams[team]
	if !exists {
		return 0, fmt.Errorf("invalid team: %d", team)
	}

	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.Frags, nil
}

// genreToVisualStyle maps genre IDs to control point visual styles.
func genreToVisualStyle(genre string) string {
	switch genre {
	case "fantasy":
		return "altar"
	case "scifi":
		return "terminal"
	case "horror":
		return "summoning-circle"
	case "cyberpunk":
		return "server-rack"
	case "postapoc":
		return "scrap-pile"
	default:
		return "generic"
	}
}

// SetGenre sets the genre for the match and updates all control point visual styles.
func (m *TerritoryMatch) SetGenre(genre string) {
	m.mu.Lock()
	m.Genre = genre

	// Get visual style for this genre
	visualStyle := genreToVisualStyle(genre)

	// Update all existing control points
	controlPoints := make([]*ControlPoint, 0, len(m.ControlPoints))
	for _, cp := range m.ControlPoints {
		controlPoints = append(controlPoints, cp)
	}
	m.mu.Unlock()

	// Update visual styles (without holding match lock)
	for _, cp := range controlPoints {
		cp.SetVisualStyle(visualStyle)
	}

	logrus.WithFields(logrus.Fields{
		"match_id":     m.MatchID,
		"genre":        genre,
		"visual_style": visualStyle,
		"cp_count":     len(controlPoints),
	}).Info("Territory match genre updated")
}

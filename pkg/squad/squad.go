// Package squad provides squad-based AI and commands.
package squad

import (
	"math"

	"github.com/opd-ai/violence/pkg/ai"
	"github.com/opd-ai/violence/pkg/rng"
)

// BehaviorState represents the current behavior of squad members.
type BehaviorState int

const (
	// BehaviorFollow means squad members follow the leader.
	BehaviorFollow BehaviorState = iota
	// BehaviorHold means squad members hold their position.
	BehaviorHold
	// BehaviorAttack means squad members engage a specific target.
	BehaviorAttack
)

// Formation defines squad member positioning.
type Formation int

const (
	// FormationLine positions squad members in a horizontal line.
	FormationLine Formation = iota
	// FormationWedge positions squad members in a V formation.
	FormationWedge
	// FormationColumn positions squad members in a vertical line.
	FormationColumn
)

// SquadMember represents an AI-controlled squad companion.
type SquadMember struct {
	ID                                 string
	X, Y                               float64
	DirX, DirY                         float64
	Health                             float64
	MaxHealth                          float64
	Speed                              float64
	WeaponID                           string
	ClassID                            string
	Agent                              *ai.Agent
	BehaviorTree                       *ai.BehaviorTree
	HoldX, HoldY                       float64
	FormationOffsetX, FormationOffsetY float64
	TargetPlayerID                     uint64 // Human player target for follow/attack
}

// HumanPlayer represents a human player in co-op mode.
type HumanPlayer struct {
	PlayerID uint64
	Name     string
	X        float64
	Y        float64
	Health   float64
	Active   bool
}

// Squad manages a group of AI squad members.
type Squad struct {
	Members      []*SquadMember
	LeaderX      float64
	LeaderY      float64
	Behavior     BehaviorState
	Formation    Formation
	TargetX      float64
	TargetY      float64
	MaxMembers   int
	CurrentGenre string
	HumanPlayers []*HumanPlayer // Connected co-op players
}

// NewSquad creates a squad with default settings.
func NewSquad(maxMembers int) *Squad {
	if maxMembers < 1 {
		maxMembers = 3
	}
	return &Squad{
		Members:      []*SquadMember{},
		Behavior:     BehaviorFollow,
		Formation:    FormationWedge,
		MaxMembers:   maxMembers,
		CurrentGenre: "fantasy",
		HumanPlayers: []*HumanPlayer{},
	}
}

// AddMember adds a squad member with the specified class.
func (s *Squad) AddMember(id, classID, weaponID string, x, y float64, seed uint64) error {
	if len(s.Members) >= s.MaxMembers {
		return nil // Silently ignore if at max capacity
	}

	// Create AI agent using class archetype
	agent := ai.NewAgent(id, x, y)

	// Adjust stats based on class
	switch classID {
	case "grunt":
		agent.MaxHealth = 100
		agent.Health = 100
		agent.Speed = 0.035
		agent.Damage = 15
		agent.AttackRange = 10
	case "medic":
		agent.MaxHealth = 80
		agent.Health = 80
		agent.Speed = 0.04
		agent.Damage = 8
		agent.AttackRange = 12
	case "demo":
		agent.MaxHealth = 90
		agent.Health = 90
		agent.Speed = 0.03
		agent.Damage = 25
		agent.AttackRange = 8
	case "mystic":
		agent.MaxHealth = 70
		agent.Health = 70
		agent.Speed = 0.038
		agent.Damage = 20
		agent.AttackRange = 15
	default:
		agent.MaxHealth = 100
		agent.Health = 100
		agent.Speed = 0.035
	}

	member := &SquadMember{
		ID:           id,
		X:            x,
		Y:            y,
		DirX:         1,
		DirY:         0,
		Health:       agent.Health,
		MaxHealth:    agent.MaxHealth,
		Speed:        agent.Speed,
		WeaponID:     weaponID,
		ClassID:      classID,
		Agent:        agent,
		BehaviorTree: ai.NewBehaviorTree(),
	}

	s.Members = append(s.Members, member)
	s.updateFormation()
	return nil
}

// RemoveMember removes a squad member by ID.
func (s *Squad) RemoveMember(id string) {
	for i, m := range s.Members {
		if m.ID == id {
			s.Members = append(s.Members[:i], s.Members[i+1:]...)
			s.updateFormation()
			return
		}
	}
}

// Command issues a command to the squad.
func (s *Squad) Command(cmd string) {
	switch cmd {
	case "follow":
		s.Behavior = BehaviorFollow
	case "hold":
		s.Behavior = BehaviorHold
		for _, m := range s.Members {
			m.HoldX = m.X
			m.HoldY = m.Y
		}
	case "attack":
		s.Behavior = BehaviorAttack
	case "formation_line":
		s.Formation = FormationLine
		s.updateFormation()
	case "formation_wedge":
		s.Formation = FormationWedge
		s.updateFormation()
	case "formation_column":
		s.Formation = FormationColumn
		s.updateFormation()
	}
}

// SetTarget sets the attack target position for the squad.
func (s *Squad) SetTarget(x, y float64) {
	s.TargetX = x
	s.TargetY = y
}

// Update advances squad AI by one tick.
func (s *Squad) Update(leaderX, leaderY float64, tileMap [][]int, playerX, playerY float64, rngSeed uint64) {
	s.LeaderX = leaderX
	s.LeaderY = leaderY

	rng := rng.NewRNG(rngSeed)

	for _, member := range s.Members {
		switch s.Behavior {
		case BehaviorFollow:
			s.updateFollow(member, tileMap)
		case BehaviorHold:
			s.updateHold(member, tileMap)
		case BehaviorAttack:
			s.updateAttack(member, tileMap, playerX, playerY, rng)
		}

		// Sync member health with agent
		member.Health = member.Agent.Health
	}
}

// updateFollow makes the squad member follow the leader with formation offset.
func (s *Squad) updateFollow(member *SquadMember, tileMap [][]int) {
	// Check if following a specific human player
	followX := s.LeaderX
	followY := s.LeaderY

	if member.TargetPlayerID != 0 {
		for _, p := range s.HumanPlayers {
			if p.PlayerID == member.TargetPlayerID && p.Active {
				followX = p.X
				followY = p.Y
				break
			}
		}
	}

	targetX := followX + member.FormationOffsetX
	targetY := followY + member.FormationOffsetY

	dx := targetX - member.X
	dy := targetY - member.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 1.5 {
		// Use A* pathfinding to navigate toward formation position
		path := ai.FindPath(member.X, member.Y, targetX, targetY, tileMap)
		if len(path) > 1 {
			nextX := path[1].X
			nextY := path[1].Y
			dx = nextX - member.X
			dy = nextY - member.Y
			dist = math.Sqrt(dx*dx + dy*dy)
		}
	}

	if dist > 0.5 && dist > 0.01 {
		moveX := member.X + (dx/dist)*member.Speed
		moveY := member.Y + (dy/dist)*member.Speed
		if isWalkable(moveX, moveY, tileMap) {
			member.X = moveX
			member.Y = moveY
		}
		member.DirX = dx / dist
		member.DirY = dy / dist
	}

	// Update agent position
	member.Agent.X = member.X
	member.Agent.Y = member.Y
}

// updateHold makes the squad member hold their position.
func (s *Squad) updateHold(member *SquadMember, tileMap [][]int) {
	dx := member.HoldX - member.X
	dy := member.HoldY - member.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 0.1 && dist > 0.01 {
		moveX := member.X + (dx/dist)*member.Speed
		moveY := member.Y + (dy/dist)*member.Speed
		if isWalkable(moveX, moveY, tileMap) {
			member.X = moveX
			member.Y = moveY
		}
		member.DirX = dx / dist
		member.DirY = dy / dist
	}

	// Update agent position
	member.Agent.X = member.X
	member.Agent.Y = member.Y
}

// updateAttack makes the squad member engage the target.
func (s *Squad) updateAttack(member *SquadMember, tileMap [][]int, playerX, playerY float64, rng *rng.RNG) {
	ctx := &ai.Context{
		TileMap:     tileMap,
		PlayerX:     s.TargetX,
		PlayerY:     s.TargetY,
		CurrentTick: 0,
		RNG:         rng,
	}

	// Use behavior tree for combat AI
	member.BehaviorTree.Tick(member.Agent, ctx)

	// Sync position
	member.X = member.Agent.X
	member.Y = member.Agent.Y
	member.DirX = member.Agent.DirX
	member.DirY = member.Agent.DirY
}

// updateFormation calculates formation offsets for all squad members.
func (s *Squad) updateFormation() {
	count := len(s.Members)
	if count == 0 {
		return
	}

	spacing := 1.5

	for i, member := range s.Members {
		switch s.Formation {
		case FormationLine:
			// Horizontal line: -X, 0, +X
			offset := float64(i - count/2)
			member.FormationOffsetX = offset * spacing
			member.FormationOffsetY = -2.0

		case FormationWedge:
			// V formation: staggered diagonal
			row := i / 2
			side := i % 2
			xOffset := float64(row) * spacing
			if side == 0 {
				xOffset = -xOffset
			}
			member.FormationOffsetX = xOffset
			member.FormationOffsetY = -float64(row+1) * spacing

		case FormationColumn:
			// Vertical line behind leader
			member.FormationOffsetX = 0
			member.FormationOffsetY = -float64(i+1) * spacing
		}
	}
}

// GetMembers returns all squad members.
func (s *Squad) GetMembers() []*SquadMember {
	return s.Members
}

// GetBehavior returns the current squad behavior state.
func (s *Squad) GetBehavior() BehaviorState {
	return s.Behavior
}

// GetFormation returns the current formation type.
func (s *Squad) GetFormation() Formation {
	return s.Formation
}

// AddHumanPlayer registers a human player for squad command targeting.
func (s *Squad) AddHumanPlayer(playerID uint64, name string, x, y float64) {
	for _, p := range s.HumanPlayers {
		if p.PlayerID == playerID {
			p.Active = true
			p.X = x
			p.Y = y
			return
		}
	}
	s.HumanPlayers = append(s.HumanPlayers, &HumanPlayer{
		PlayerID: playerID,
		Name:     name,
		X:        x,
		Y:        y,
		Health:   100.0,
		Active:   true,
	})
}

// RemoveHumanPlayer marks a human player as inactive (disconnected).
func (s *Squad) RemoveHumanPlayer(playerID uint64) {
	for _, p := range s.HumanPlayers {
		if p.PlayerID == playerID {
			p.Active = false
			return
		}
	}
}

// UpdateHumanPlayer updates a human player's position and health.
func (s *Squad) UpdateHumanPlayer(playerID uint64, x, y, health float64) {
	for _, p := range s.HumanPlayers {
		if p.PlayerID == playerID {
			p.X = x
			p.Y = y
			p.Health = health
			return
		}
	}
}

// GetHumanPlayers returns all active human players.
func (s *Squad) GetHumanPlayers() []*HumanPlayer {
	active := []*HumanPlayer{}
	for _, p := range s.HumanPlayers {
		if p.Active {
			active = append(active, p)
		}
	}
	return active
}

// CommandTargetPlayer issues a command targeting a specific human player.
func (s *Squad) CommandTargetPlayer(cmd string, targetPlayerID uint64) {
	switch cmd {
	case "follow_player":
		s.Behavior = BehaviorFollow
		for _, m := range s.Members {
			m.TargetPlayerID = targetPlayerID
		}
	case "attack_player_target":
		// Attack what the target player is attacking
		s.Behavior = BehaviorAttack
		for _, m := range s.Members {
			m.TargetPlayerID = targetPlayerID
		}
	}
}

// SetGenre configures squad behavior for a genre.
func SetGenre(genreID string) {
	// Genre-specific behavior could be added here in the future
	ai.SetGenre(genreID)
}

// isWalkable checks if a position is on a walkable tile.
func isWalkable(x, y float64, tileMap [][]int) bool {
	if tileMap == nil || len(tileMap) == 0 || len(tileMap[0]) == 0 {
		return true
	}
	mapX := int(x)
	mapY := int(y)
	if mapY < 0 || mapY >= len(tileMap) || mapX < 0 || mapX >= len(tileMap[0]) {
		return false
	}
	tile := tileMap[mapY][mapX]
	return tile == 0 || tile == 2
}

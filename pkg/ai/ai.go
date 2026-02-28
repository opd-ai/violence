// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"math"

	"github.com/opd-ai/violence/pkg/rng"
)

// NodeStatus represents the return status of a behavior tree node.
type NodeStatus int

const (
	// StatusRunning indicates the node is still executing.
	StatusRunning NodeStatus = iota
	// StatusSuccess indicates the node succeeded.
	StatusSuccess
	// StatusFailure indicates the node failed.
	StatusFailure
)

// Node is the interface for all behavior tree nodes.
type Node interface {
	Tick(agent *Agent, ctx *Context) NodeStatus
}

// Selector executes children until one succeeds (OR logic).
type Selector struct {
	Children []Node
	current  int
}

// NewSelector creates a selector node.
func NewSelector(children ...Node) *Selector {
	return &Selector{Children: children}
}

// Tick executes children in order until one succeeds.
func (s *Selector) Tick(agent *Agent, ctx *Context) NodeStatus {
	for i := s.current; i < len(s.Children); i++ {
		status := s.Children[i].Tick(agent, ctx)
		if status == StatusRunning {
			s.current = i
			return StatusRunning
		}
		if status == StatusSuccess {
			s.current = 0
			return StatusSuccess
		}
	}
	s.current = 0
	return StatusFailure
}

// Sequence executes children until one fails (AND logic).
type Sequence struct {
	Children []Node
	current  int
}

// NewSequence creates a sequence node.
func NewSequence(children ...Node) *Sequence {
	return &Sequence{Children: children}
}

// Tick executes children in order until one fails.
func (s *Sequence) Tick(agent *Agent, ctx *Context) NodeStatus {
	for i := s.current; i < len(s.Children); i++ {
		status := s.Children[i].Tick(agent, ctx)
		if status == StatusRunning {
			s.current = i
			return StatusRunning
		}
		if status == StatusFailure {
			s.current = 0
			return StatusFailure
		}
	}
	s.current = 0
	return StatusSuccess
}

// Condition wraps a boolean function as a behavior tree node.
type Condition struct {
	Check func(*Agent, *Context) bool
}

// NewCondition creates a condition node.
func NewCondition(check func(*Agent, *Context) bool) *Condition {
	return &Condition{Check: check}
}

// Tick evaluates the condition.
func (c *Condition) Tick(agent *Agent, ctx *Context) NodeStatus {
	if c.Check(agent, ctx) {
		return StatusSuccess
	}
	return StatusFailure
}

// Action wraps an action function as a behavior tree node.
type Action struct {
	Execute func(*Agent, *Context) NodeStatus
}

// NewAction creates an action node.
func NewAction(execute func(*Agent, *Context) NodeStatus) *Action {
	return &Action{Execute: execute}
}

// Tick executes the action.
func (a *Action) Tick(agent *Agent, ctx *Context) NodeStatus {
	return a.Execute(agent, ctx)
}

// State represents an AI state.
type State int

const (
	// StateIdle means enemy is standing still.
	StateIdle State = iota
	// StatePatrol means enemy is following waypoints.
	StatePatrol
	// StateAlert means enemy heard something.
	StateAlert
	// StateChase means enemy is pursuing player.
	StateChase
	// StateStrafe means enemy is dodging.
	StateStrafe
	// StateCover means enemy is seeking cover.
	StateCover
	// StateRetreat means enemy is fleeing.
	StateRetreat
	// StateAttack means enemy is firing weapon.
	StateAttack
)

// Agent represents an AI-controlled enemy entity.
type Agent struct {
	ID                 string
	X, Y               float64
	DirX, DirY         float64
	Health, MaxHealth  float64
	Speed              float64
	AlertRadius        float64
	HearRadius         float64
	State              State
	TargetX, TargetY   float64
	PatrolWaypoints    []Waypoint
	PatrolIndex        int
	Cooldown           int
	StrafeDirection    float64
	ArchetypeID        string
	Damage             float64
	AttackRange        float64
	RetreatHealthRatio float64
}

// Waypoint represents a patrol destination.
type Waypoint struct {
	X, Y float64
}

// Context provides world state to behavior tree nodes.
type Context struct {
	TileMap      [][]int
	PlayerX      float64
	PlayerY      float64
	LastShotX    float64
	LastShotY    float64
	LastShotTick int
	CurrentTick  int
	RNG          *rng.RNG
}

// BehaviorTree represents an AI decision tree.
type BehaviorTree struct {
	Root Node
}

// NewBehaviorTree creates a behavior tree for FPS enemies.
func NewBehaviorTree() *BehaviorTree {
	// Build behavior tree: retreat if low health, otherwise attack if can see player,
	// chase if player heard, alert if gunshot heard, patrol or idle.
	root := NewSelector(
		// Retreat if health < 25%
		NewSequence(
			NewCondition(checkLowHealth),
			NewAction(actionRetreat),
		),
		// Attack if player in sight and range
		NewSequence(
			NewCondition(checkCanSeePlayer),
			NewCondition(checkInAttackRange),
			NewAction(actionAttack),
		),
		// Strafe if player in sight but not attacking
		NewSequence(
			NewCondition(checkCanSeePlayer),
			NewAction(actionStrafe),
		),
		// Chase if player in sight
		NewSequence(
			NewCondition(checkCanSeePlayer),
			NewAction(actionChase),
		),
		// Investigate if heard gunshot
		NewSequence(
			NewCondition(checkHeardGunshot),
			NewAction(actionAlert),
		),
		// Patrol waypoints
		NewAction(actionPatrol),
	)
	return &BehaviorTree{Root: root}
}

// Tick evaluates the behavior tree.
func (bt *BehaviorTree) Tick(agent *Agent, ctx *Context) {
	bt.Root.Tick(agent, ctx)
}

// Condition functions

func checkLowHealth(agent *Agent, ctx *Context) bool {
	return agent.Health < agent.MaxHealth*agent.RetreatHealthRatio
}

func checkCanSeePlayer(agent *Agent, ctx *Context) bool {
	return lineOfSight(agent.X, agent.Y, ctx.PlayerX, ctx.PlayerY, ctx.TileMap)
}

func checkInAttackRange(agent *Agent, ctx *Context) bool {
	dx := ctx.PlayerX - agent.X
	dy := ctx.PlayerY - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	return dist <= agent.AttackRange
}

func checkHeardGunshot(agent *Agent, ctx *Context) bool {
	if ctx.CurrentTick-ctx.LastShotTick > 60 {
		return false
	}
	dx := ctx.LastShotX - agent.X
	dy := ctx.LastShotY - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	return dist <= agent.HearRadius
}

// Action functions

func actionRetreat(agent *Agent, ctx *Context) NodeStatus {
	agent.State = StateRetreat
	// Move away from player
	dx := agent.X - ctx.PlayerX
	dy := agent.Y - ctx.PlayerY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.01 {
		return StatusRunning
	}
	moveX := agent.X + (dx/dist)*agent.Speed*1.5
	moveY := agent.Y + (dy/dist)*agent.Speed*1.5
	if isWalkable(moveX, moveY, ctx.TileMap) {
		agent.X = moveX
		agent.Y = moveY
	}
	return StatusRunning
}

func actionAttack(agent *Agent, ctx *Context) NodeStatus {
	agent.State = StateAttack
	if agent.Cooldown > 0 {
		agent.Cooldown--
		return StatusRunning
	}
	// Aim at player
	dx := ctx.PlayerX - agent.X
	dy := ctx.PlayerY - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > 0.01 {
		agent.DirX = dx / dist
		agent.DirY = dy / dist
	}
	// Fire weapon (cooldown 30 ticks)
	agent.Cooldown = 30
	return StatusSuccess
}

func actionStrafe(agent *Agent, ctx *Context) NodeStatus {
	agent.State = StateStrafe
	// Strafe perpendicular to player direction
	dx := ctx.PlayerX - agent.X
	dy := ctx.PlayerY - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.01 {
		return StatusRunning
	}
	// Perpendicular direction
	perpX := -dy / dist
	perpY := dx / dist
	// Flip strafe direction periodically
	if ctx.RNG.Intn(120) == 0 {
		agent.StrafeDirection = -agent.StrafeDirection
		if agent.StrafeDirection == 0 {
			agent.StrafeDirection = 1
		}
	}
	moveX := agent.X + perpX*agent.Speed*agent.StrafeDirection
	moveY := agent.Y + perpY*agent.Speed*agent.StrafeDirection
	if isWalkable(moveX, moveY, ctx.TileMap) {
		agent.X = moveX
		agent.Y = moveY
	}
	// Face player
	agent.DirX = dx / dist
	agent.DirY = dy / dist
	return StatusRunning
}

func actionChase(agent *Agent, ctx *Context) NodeStatus {
	agent.State = StateChase
	// Use A* pathfinding to navigate toward player
	path := FindPath(agent.X, agent.Y, ctx.PlayerX, ctx.PlayerY, ctx.TileMap)
	if len(path) > 1 {
		agent.TargetX = path[1].X
		agent.TargetY = path[1].Y
		dx := agent.TargetX - agent.X
		dy := agent.TargetY - agent.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > 0.01 {
			agent.DirX = dx / dist
			agent.DirY = dy / dist
			moveX := agent.X + agent.DirX*agent.Speed
			moveY := agent.Y + agent.DirY*agent.Speed
			if isWalkable(moveX, moveY, ctx.TileMap) {
				agent.X = moveX
				agent.Y = moveY
			}
		}
	}
	return StatusRunning
}

func actionAlert(agent *Agent, ctx *Context) NodeStatus {
	agent.State = StateAlert
	// Move toward last gunshot position
	dx := ctx.LastShotX - agent.X
	dy := ctx.LastShotY - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.5 {
		return StatusSuccess
	}
	if dist > 0.01 {
		agent.DirX = dx / dist
		agent.DirY = dy / dist
		moveX := agent.X + agent.DirX*agent.Speed
		moveY := agent.Y + agent.DirY*agent.Speed
		if isWalkable(moveX, moveY, ctx.TileMap) {
			agent.X = moveX
			agent.Y = moveY
		}
	}
	return StatusRunning
}

func actionPatrol(agent *Agent, ctx *Context) NodeStatus {
	agent.State = StatePatrol
	if len(agent.PatrolWaypoints) == 0 {
		agent.State = StateIdle
		return StatusRunning
	}
	wp := agent.PatrolWaypoints[agent.PatrolIndex]
	dx := wp.X - agent.X
	dy := wp.Y - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.5 {
		agent.PatrolIndex = (agent.PatrolIndex + 1) % len(agent.PatrolWaypoints)
		return StatusRunning
	}
	if dist > 0.01 {
		agent.DirX = dx / dist
		agent.DirY = dy / dist
		moveX := agent.X + agent.DirX*agent.Speed*0.5
		moveY := agent.Y + agent.DirY*agent.Speed*0.5
		if isWalkable(moveX, moveY, ctx.TileMap) {
			agent.X = moveX
			agent.Y = moveY
		}
	}
	return StatusRunning
}

// lineOfSight checks if there is unobstructed view between two points.
func lineOfSight(x1, y1, x2, y2 float64, tileMap [][]int) bool {
	if tileMap == nil || len(tileMap) == 0 || len(tileMap[0]) == 0 {
		return false
	}
	// DDA ray cast from (x1,y1) to (x2,y2)
	dx := x2 - x1
	dy := y2 - y1
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.01 {
		return true
	}
	steps := int(dist * 2)
	stepX := dx / float64(steps)
	stepY := dy / float64(steps)
	x, y := x1, y1
	for i := 0; i < steps; i++ {
		x += stepX
		y += stepY
		mapX := int(x)
		mapY := int(y)
		if mapY < 0 || mapY >= len(tileMap) || mapX < 0 || mapX >= len(tileMap[0]) {
			return false
		}
		if tileMap[mapY][mapX] == 1 {
			return false
		}
	}
	return true
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

// pathNode is used for A* pathfinding.
type pathNode struct {
	x, y   int
	g, h   float64
	parent *pathNode
}

func (n *pathNode) f() float64 {
	return n.g + n.h
}

// FindPath uses A* to find a path from start to goal.
func FindPath(x1, y1, x2, y2 float64, tileMap [][]int) []Waypoint {
	if tileMap == nil || len(tileMap) == 0 || len(tileMap[0]) == 0 {
		return []Waypoint{{X: x1, Y: y1}}
	}
	startX, startY := int(x1), int(y1)
	goalX, goalY := int(x2), int(y2)
	if startX < 0 || startY < 0 || startX >= len(tileMap[0]) || startY >= len(tileMap) {
		return []Waypoint{{X: x1, Y: y1}}
	}
	if goalX < 0 || goalY < 0 || goalX >= len(tileMap[0]) || goalY >= len(tileMap) {
		return []Waypoint{{X: x1, Y: y1}}
	}
	// A* implementation
	openSet := []*pathNode{{x: startX, y: startY, g: 0, h: heuristic(startX, startY, goalX, goalY)}}
	closedSet := make(map[int]bool)
	maxIter := 500
	for iter := 0; iter < maxIter && len(openSet) > 0; iter++ {
		// Find node with lowest f
		current := openSet[0]
		currentIdx := 0
		for i, node := range openSet {
			if node.f() < current.f() {
				current = node
				currentIdx = i
			}
		}
		// Remove from open set
		openSet = append(openSet[:currentIdx], openSet[currentIdx+1:]...)
		// Check if reached goal
		if current.x == goalX && current.y == goalY {
			return reconstructPath(current)
		}
		closedSet[current.y*len(tileMap[0])+current.x] = true
		// Check neighbors
		for _, dir := range []struct{ dx, dy int }{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} {
			nx, ny := current.x+dir.dx, current.y+dir.dy
			if ny < 0 || ny >= len(tileMap) || nx < 0 || nx >= len(tileMap[0]) {
				continue
			}
			if !isWalkable(float64(nx)+0.5, float64(ny)+0.5, tileMap) {
				continue
			}
			if closedSet[ny*len(tileMap[0])+nx] {
				continue
			}
			g := current.g + 1
			h := heuristic(nx, ny, goalX, goalY)
			neighbor := &pathNode{x: nx, y: ny, g: g, h: h, parent: current}
			// Check if already in open set
			found := false
			for i, node := range openSet {
				if node.x == nx && node.y == ny {
					if g < node.g {
						openSet[i] = neighbor
					}
					found = true
					break
				}
			}
			if !found {
				openSet = append(openSet, neighbor)
			}
		}
	}
	// No path found, return direct line
	return []Waypoint{{X: x1, Y: y1}, {X: x2, Y: y2}}
}

func heuristic(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}

func reconstructPath(node *pathNode) []Waypoint {
	path := []Waypoint{}
	for n := node; n != nil; n = n.parent {
		path = append([]Waypoint{{X: float64(n.x) + 0.5, Y: float64(n.y) + 0.5}}, path...)
	}
	return path
}

// Archetype defines enemy type characteristics.
type Archetype struct {
	ID                 string
	MaxHealth          float64
	Speed              float64
	Damage             float64
	AttackRange        float64
	AlertRadius        float64
	HearRadius         float64
	RetreatHealthRatio float64
}

var archetypes = map[string]Archetype{
	"fantasy_guard": {
		ID:                 "fantasy_guard",
		MaxHealth:          50,
		Speed:              0.03,
		Damage:             10,
		AttackRange:        8,
		AlertRadius:        10,
		HearRadius:         15,
		RetreatHealthRatio: 0.2,
	},
	"scifi_soldier": {
		ID:                 "scifi_soldier",
		MaxHealth:          60,
		Speed:              0.035,
		Damage:             12,
		AttackRange:        10,
		AlertRadius:        12,
		HearRadius:         18,
		RetreatHealthRatio: 0.25,
	},
	"horror_cultist": {
		ID:                 "horror_cultist",
		MaxHealth:          40,
		Speed:              0.025,
		Damage:             15,
		AttackRange:        6,
		AlertRadius:        8,
		HearRadius:         20,
		RetreatHealthRatio: 0.1,
	},
	"cyberpunk_drone": {
		ID:                 "cyberpunk_drone",
		MaxHealth:          45,
		Speed:              0.04,
		Damage:             10,
		AttackRange:        12,
		AlertRadius:        15,
		HearRadius:         12,
		RetreatHealthRatio: 0.3,
	},
	"postapoc_scavenger": {
		ID:                 "postapoc_scavenger",
		MaxHealth:          55,
		Speed:              0.032,
		Damage:             11,
		AttackRange:        7,
		AlertRadius:        9,
		HearRadius:         16,
		RetreatHealthRatio: 0.25,
	},
}

var currentGenre = "fantasy"

// SetGenre configures AI behaviors for a genre.
func SetGenre(genreID string) {
	currentGenre = genreID
}

// GetArchetype returns the archetype for the current genre.
func GetArchetype() Archetype {
	switch currentGenre {
	case "fantasy":
		return archetypes["fantasy_guard"]
	case "scifi":
		return archetypes["scifi_soldier"]
	case "horror":
		return archetypes["horror_cultist"]
	case "cyberpunk":
		return archetypes["cyberpunk_drone"]
	case "postapoc":
		return archetypes["postapoc_scavenger"]
	default:
		return archetypes["fantasy_guard"]
	}
}

// NewAgent creates an agent from archetype.
func NewAgent(id string, x, y float64) *Agent {
	arch := GetArchetype()
	return &Agent{
		ID:                 id,
		X:                  x,
		Y:                  y,
		DirX:               1,
		DirY:               0,
		Health:             arch.MaxHealth,
		MaxHealth:          arch.MaxHealth,
		Speed:              arch.Speed,
		AlertRadius:        arch.AlertRadius,
		HearRadius:         arch.HearRadius,
		State:              StatePatrol,
		Damage:             arch.Damage,
		AttackRange:        arch.AttackRange,
		RetreatHealthRatio: arch.RetreatHealthRatio,
		ArchetypeID:        arch.ID,
		StrafeDirection:    1,
	}
}

package ai

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/rng"
)

func TestBehaviorTree_Selector(t *testing.T) {
	agent := &Agent{Health: 50, MaxHealth: 100}
	ctx := &Context{}
	failAction := NewAction(func(a *Agent, c *Context) NodeStatus { return StatusFailure })
	successAction := NewAction(func(a *Agent, c *Context) NodeStatus { return StatusSuccess })
	selector := NewSelector(failAction, successAction)
	status := selector.Tick(agent, ctx)
	if status != StatusSuccess {
		t.Errorf("Selector should succeed when one child succeeds, got %v", status)
	}
}

func TestBehaviorTree_Sequence(t *testing.T) {
	agent := &Agent{Health: 50, MaxHealth: 100}
	ctx := &Context{}
	successAction := NewAction(func(a *Agent, c *Context) NodeStatus { return StatusSuccess })
	failAction := NewAction(func(a *Agent, c *Context) NodeStatus { return StatusFailure })
	sequence := NewSequence(successAction, failAction)
	status := sequence.Tick(agent, ctx)
	if status != StatusFailure {
		t.Errorf("Sequence should fail when one child fails, got %v", status)
	}
}

func TestBehaviorTree_Condition(t *testing.T) {
	agent := &Agent{Health: 20, MaxHealth: 100, RetreatHealthRatio: 0.25}
	ctx := &Context{}
	condition := NewCondition(checkLowHealth)
	status := condition.Tick(agent, ctx)
	if status != StatusSuccess {
		t.Errorf("Condition should succeed for low health, got %v", status)
	}
	agent.Health = 80
	status = condition.Tick(agent, ctx)
	if status != StatusFailure {
		t.Errorf("Condition should fail for high health, got %v", status)
	}
}

func TestCheckLowHealth(t *testing.T) {
	tests := []struct {
		name     string
		health   float64
		maxHP    float64
		ratio    float64
		expected bool
	}{
		{"low health", 10, 100, 0.25, true},
		{"exact threshold", 25, 100, 0.25, false},
		{"high health", 80, 100, 0.25, false},
		{"custom ratio", 15, 100, 0.2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{Health: tt.health, MaxHealth: tt.maxHP, RetreatHealthRatio: tt.ratio}
			ctx := &Context{}
			result := checkLowHealth(agent, ctx)
			if result != tt.expected {
				t.Errorf("checkLowHealth() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLineOfSight(t *testing.T) {
	tileMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 1, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}
	tests := []struct {
		name     string
		x1, y1   float64
		x2, y2   float64
		expected bool
	}{
		{"clear sight", 1.5, 1.5, 3.5, 1.5, true},
		{"blocked by wall", 1.5, 1.5, 3.5, 2.5, false},
		{"diagonal blocked", 1.5, 1.5, 3.5, 3.5, false},
		{"same position", 2.5, 2.5, 2.5, 2.5, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lineOfSight(tt.x1, tt.y1, tt.x2, tt.y2, tileMap)
			if result != tt.expected {
				t.Errorf("lineOfSight() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLineOfSight_NilMap(t *testing.T) {
	result := lineOfSight(1, 1, 2, 2, nil)
	if result != false {
		t.Errorf("lineOfSight with nil map should return false")
	}
}

func TestLineOfSight_EmptyMap(t *testing.T) {
	result := lineOfSight(1, 1, 2, 2, [][]int{})
	if result != false {
		t.Errorf("lineOfSight with empty map should return false")
	}
}

func TestIsWalkable(t *testing.T) {
	tileMap := [][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 2, 1},
	}
	tests := []struct {
		name     string
		x, y     float64
		expected bool
	}{
		{"empty tile", 1.5, 1.5, true},
		{"floor tile", 1.5, 2.5, true},
		{"wall tile", 0.5, 0.5, false},
		{"out of bounds", -1, -1, false},
		{"out of bounds high", 10, 10, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWalkable(tt.x, tt.y, tileMap)
			if result != tt.expected {
				t.Errorf("isWalkable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsWalkable_NilMap(t *testing.T) {
	result := isWalkable(1, 1, nil)
	if result != true {
		t.Errorf("isWalkable with nil map should return true (default)")
	}
}

func TestFindPath(t *testing.T) {
	tileMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 1, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}
	path := FindPath(1.5, 1.5, 3.5, 3.5, tileMap)
	if len(path) < 2 {
		t.Errorf("FindPath should return path with at least 2 waypoints")
	}
	// Check first waypoint is near start
	if math.Abs(path[0].X-1.5) > 1 || math.Abs(path[0].Y-1.5) > 1 {
		t.Errorf("First waypoint should be near start")
	}
	// Check last waypoint is near goal
	last := path[len(path)-1]
	if math.Abs(last.X-3.5) > 1 || math.Abs(last.Y-3.5) > 1 {
		t.Errorf("Last waypoint should be near goal")
	}
}

func TestFindPath_Blocked(t *testing.T) {
	tileMap := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 1, 0, 1},
		{1, 0, 1, 0, 1},
		{1, 0, 1, 0, 1},
		{1, 1, 1, 1, 1},
	}
	path := FindPath(1.5, 1.5, 3.5, 1.5, tileMap)
	// Should return direct line if no path found
	if len(path) != 2 {
		t.Errorf("Blocked path should return direct line with 2 waypoints, got %d", len(path))
	}
}

func TestFindPath_NilMap(t *testing.T) {
	path := FindPath(1, 1, 2, 2, nil)
	if len(path) != 1 {
		t.Errorf("FindPath with nil map should return start waypoint")
	}
}

func TestActionRetreat(t *testing.T) {
	agent := &Agent{
		X: 5, Y: 5,
		Health: 20, MaxHealth: 100,
		Speed: 0.05,
	}
	ctx := &Context{
		PlayerX: 7, PlayerY: 7,
		TileMap: [][]int{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	oldX, oldY := agent.X, agent.Y
	status := actionRetreat(agent, ctx)
	if status != StatusRunning {
		t.Errorf("actionRetreat should return Running")
	}
	if agent.State != StateRetreat {
		t.Errorf("actionRetreat should set state to Retreat")
	}
	// Check agent moved away from player
	distOld := math.Sqrt(math.Pow(oldX-ctx.PlayerX, 2) + math.Pow(oldY-ctx.PlayerY, 2))
	distNew := math.Sqrt(math.Pow(agent.X-ctx.PlayerX, 2) + math.Pow(agent.Y-ctx.PlayerY, 2))
	if distNew <= distOld {
		t.Errorf("Agent should move away from player during retreat")
	}
}

func TestActionAttack(t *testing.T) {
	agent := &Agent{
		X: 5, Y: 5,
		Cooldown: 0,
	}
	ctx := &Context{
		PlayerX: 7, PlayerY: 5,
	}
	status := actionAttack(agent, ctx)
	if status != StatusSuccess {
		t.Errorf("actionAttack should return Success")
	}
	if agent.State != StateAttack {
		t.Errorf("actionAttack should set state to Attack")
	}
	if agent.Cooldown != 30 {
		t.Errorf("actionAttack should set cooldown to 30, got %d", agent.Cooldown)
	}
	// Check agent is facing player
	expectedDirX := (ctx.PlayerX - agent.X) / 2.0
	if math.Abs(agent.DirX-expectedDirX) > 0.01 {
		t.Errorf("Agent should face player, dirX = %f, want %f", agent.DirX, expectedDirX)
	}
}

func TestActionAttack_Cooldown(t *testing.T) {
	agent := &Agent{
		X: 5, Y: 5,
		Cooldown: 10,
	}
	ctx := &Context{
		PlayerX: 7, PlayerY: 5,
	}
	status := actionAttack(agent, ctx)
	if status != StatusRunning {
		t.Errorf("actionAttack with cooldown should return Running")
	}
	if agent.Cooldown != 9 {
		t.Errorf("actionAttack should decrement cooldown, got %d", agent.Cooldown)
	}
}

func TestActionChase(t *testing.T) {
	tileMap := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	agent := &Agent{
		X: 1.5, Y: 1.5,
		Speed: 0.1,
	}
	ctx := &Context{
		PlayerX: 5.5, PlayerY: 1.5,
		TileMap: tileMap,
	}
	oldX := agent.X
	status := actionChase(agent, ctx)
	if status != StatusRunning {
		t.Errorf("actionChase should return Running")
	}
	if agent.State != StateChase {
		t.Errorf("actionChase should set state to Chase")
	}
	// Agent should move toward player
	if agent.X <= oldX {
		t.Errorf("Agent should move toward player")
	}
}

func TestActionPatrol(t *testing.T) {
	agent := &Agent{
		X:     1.5,
		Y:     1.5,
		Speed: 0.1,
		PatrolWaypoints: []Waypoint{
			{X: 2.5, Y: 1.5},
			{X: 2.5, Y: 2.5},
		},
		PatrolIndex: 0,
	}
	ctx := &Context{
		TileMap: [][]int{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
	}
	oldX := agent.X
	status := actionPatrol(agent, ctx)
	if status != StatusRunning {
		t.Errorf("actionPatrol should return Running")
	}
	if agent.State != StatePatrol {
		t.Errorf("actionPatrol should set state to Patrol")
	}
	// Agent should move toward first waypoint
	if agent.X <= oldX {
		t.Errorf("Agent should move toward waypoint")
	}
}

func TestActionPatrol_NoWaypoints(t *testing.T) {
	agent := &Agent{
		X:               1.5,
		Y:               1.5,
		PatrolWaypoints: []Waypoint{},
	}
	ctx := &Context{}
	status := actionPatrol(agent, ctx)
	if status != StatusRunning {
		t.Errorf("actionPatrol with no waypoints should return Running")
	}
	if agent.State != StateIdle {
		t.Errorf("actionPatrol with no waypoints should set state to Idle")
	}
}

func TestActionPatrol_WaypointCycle(t *testing.T) {
	agent := &Agent{
		X:     2.4,
		Y:     1.5,
		Speed: 0.1,
		PatrolWaypoints: []Waypoint{
			{X: 2.5, Y: 1.5},
			{X: 3.5, Y: 1.5},
		},
		PatrolIndex: 0,
	}
	ctx := &Context{
		TileMap: [][]int{{0, 0, 0, 0, 0}},
	}
	actionPatrol(agent, ctx)
	if agent.PatrolIndex != 1 {
		t.Errorf("Agent should advance to next waypoint")
	}
}

func TestCheckHeardGunshot(t *testing.T) {
	agent := &Agent{
		X: 5, Y: 5,
		HearRadius: 15,
	}
	ctx := &Context{
		LastShotX:    10,
		LastShotY:    5,
		LastShotTick: 100,
		CurrentTick:  110,
	}
	result := checkHeardGunshot(agent, ctx)
	if !result {
		t.Errorf("Agent should hear nearby gunshot")
	}
	// Too far
	ctx.LastShotX = 100
	result = checkHeardGunshot(agent, ctx)
	if result {
		t.Errorf("Agent should not hear distant gunshot")
	}
	// Too old
	ctx.LastShotX = 10
	ctx.LastShotTick = 10
	result = checkHeardGunshot(agent, ctx)
	if result {
		t.Errorf("Agent should not hear old gunshot")
	}
}

func TestNewBehaviorTree(t *testing.T) {
	bt := NewBehaviorTree()
	if bt.Root == nil {
		t.Errorf("NewBehaviorTree should create root node")
	}
}

func TestBehaviorTree_Tick(t *testing.T) {
	bt := NewBehaviorTree()
	agent := &Agent{
		X: 5, Y: 5,
		Health: 100, MaxHealth: 100,
		RetreatHealthRatio: 0.25,
		Speed:              0.05,
	}
	ctx := &Context{
		PlayerX: 7, PlayerY: 5,
		TileMap: [][]int{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		RNG: rng.NewRNG(12345),
	}
	// Should not panic
	bt.Tick(agent, ctx)
}

func TestArchetypes(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			SetGenre(genre)
			arch := GetArchetype()
			if arch.MaxHealth <= 0 {
				t.Errorf("Archetype %s should have positive max health", genre)
			}
			if arch.Speed <= 0 {
				t.Errorf("Archetype %s should have positive speed", genre)
			}
			if arch.Damage <= 0 {
				t.Errorf("Archetype %s should have positive damage", genre)
			}
		})
	}
}

func TestArchetypes_Distinctiveness(t *testing.T) {
	SetGenre("fantasy")
	fantasy := GetArchetype()
	SetGenre("scifi")
	scifi := GetArchetype()
	if fantasy.ID == scifi.ID {
		t.Errorf("Different genres should have different archetype IDs")
	}
}

func TestNewAgent(t *testing.T) {
	SetGenre("fantasy")
	agent := NewAgent("test-1", 5.5, 7.5)
	if agent.ID != "test-1" {
		t.Errorf("NewAgent should set ID")
	}
	if agent.X != 5.5 || agent.Y != 7.5 {
		t.Errorf("NewAgent should set position")
	}
	if agent.Health != agent.MaxHealth {
		t.Errorf("NewAgent should start at full health")
	}
	if agent.MaxHealth <= 0 {
		t.Errorf("NewAgent should have positive max health")
	}
	if agent.State != StatePatrol {
		t.Errorf("NewAgent should start in patrol state")
	}
}

func TestSetGenre(t *testing.T) {
	SetGenre("scifi")
	if currentGenre != "scifi" {
		t.Errorf("SetGenre should update current genre")
	}
}

func TestGetArchetype_Default(t *testing.T) {
	SetGenre("unknown")
	arch := GetArchetype()
	if arch.ID != "fantasy_guard" {
		t.Errorf("Unknown genre should default to fantasy archetype")
	}
}

func TestActionStrafe(t *testing.T) {
	agent := &Agent{
		X:               5,
		Y:               5,
		Speed:           0.1,
		StrafeDirection: 1,
	}
	ctx := &Context{
		PlayerX: 7,
		PlayerY: 5,
		TileMap: [][]int{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		RNG: rng.NewRNG(999),
	}
	oldY := agent.Y
	status := actionStrafe(agent, ctx)
	if status != StatusRunning {
		t.Errorf("actionStrafe should return Running")
	}
	if agent.State != StateStrafe {
		t.Errorf("actionStrafe should set state to Strafe")
	}
	// Agent should move perpendicular to player
	if agent.Y == oldY {
		t.Errorf("Agent should strafe perpendicular to player")
	}
}

func TestActionAlert(t *testing.T) {
	agent := &Agent{
		X:     1,
		Y:     1,
		Speed: 0.1,
	}
	ctx := &Context{
		LastShotX: 3,
		LastShotY: 1,
		TileMap: [][]int{
			{0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0},
		},
	}
	oldX := agent.X
	status := actionAlert(agent, ctx)
	if status != StatusRunning {
		t.Errorf("actionAlert should return Running")
	}
	if agent.State != StateAlert {
		t.Errorf("actionAlert should set state to Alert")
	}
	// Agent should move toward gunshot
	if agent.X <= oldX {
		t.Errorf("Agent should move toward gunshot location")
	}
}

func TestActionAlert_Arrived(t *testing.T) {
	agent := &Agent{
		X:     2.9,
		Y:     1,
		Speed: 0.1,
	}
	ctx := &Context{
		LastShotX: 3,
		LastShotY: 1,
		TileMap:   [][]int{{0, 0, 0, 0}},
	}
	status := actionAlert(agent, ctx)
	if status != StatusSuccess {
		t.Errorf("actionAlert should return Success when arrived")
	}
}

func BenchmarkBehaviorTree_Tick(b *testing.B) {
	bt := NewBehaviorTree()
	agent := NewAgent("bench", 5, 5)
	ctx := &Context{
		PlayerX: 10,
		PlayerY: 10,
		TileMap: make([][]int, 20),
		RNG:     rng.NewRNG(12345),
	}
	for i := range ctx.TileMap {
		ctx.TileMap[i] = make([]int, 20)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bt.Tick(agent, ctx)
	}
}

func BenchmarkLineOfSight(b *testing.B) {
	tileMap := make([][]int, 50)
	for i := range tileMap {
		tileMap[i] = make([]int, 50)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lineOfSight(5, 5, 45, 45, tileMap)
	}
}

func BenchmarkFindPath(b *testing.B) {
	tileMap := make([][]int, 30)
	for i := range tileMap {
		tileMap[i] = make([]int, 30)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindPath(1, 1, 28, 28, tileMap)
	}
}

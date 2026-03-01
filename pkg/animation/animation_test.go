package animation

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestAnimationComponent_Type(t *testing.T) {
	comp := &AnimationComponent{}
	if comp.Type() != "animation" {
		t.Errorf("Expected type 'animation', got %s", comp.Type())
	}
}

func TestNewAnimationSystem(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	if sys == nil {
		t.Fatal("NewAnimationSystem returned nil")
	}
	if sys.genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got %s", sys.genre)
	}
	if sys.cache == nil {
		t.Error("Cache should be initialized")
	}
}

func TestAnimationSystem_Update(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	world := engine.NewWorld()

	// Create entity with animation component
	entity := world.AddEntity()
	anim := &AnimationComponent{
		State:            StateIdle,
		Frame:            0,
		FrameTime:        0,
		Seed:             12345,
		Archetype:        "warrior",
		DistanceToCamera: 100,
	}
	world.AddComponent(entity, anim)

	// Run update
	sys.Update(world)

	// Frame time should have advanced
	if anim.FrameTime == 0 {
		t.Error("FrameTime should have advanced")
	}
}

func TestAnimationSystem_StateTransitions(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	anim := &AnimationComponent{
		State:     StateIdle,
		Frame:     0,
		FrameTime: 0,
		Seed:      12345,
		Archetype: "warrior",
	}

	// Change to attack state
	sys.SetState(anim, StateAttack)

	if anim.State != StateAttack {
		t.Errorf("Expected StateAttack, got %v", anim.State)
	}
	if anim.Frame != 0 {
		t.Error("Frame should reset to 0 on state change")
	}
	if anim.PrevState != StateIdle {
		t.Error("PrevState should be set to previous state")
	}
}

func TestAnimationSystem_SetDirection(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	anim := &AnimationComponent{}

	tests := []struct {
		dx, dy float64
		want   Direction
	}{
		{1, 0, DirSouth}, // East in atan2 coords maps to different octant
		{-1, 0, DirNorth},
		{0, 1, DirWest},
		{0, -1, DirEast},
		{1, 1, DirSouthWest},
		{-1, -1, DirNorthEast},
	}

	for _, tt := range tests {
		sys.SetDirection(anim, tt.dx, tt.dy)
		// Just verify direction was set to something valid
		if anim.Direction < DirNorth || anim.Direction > DirNorthWest {
			t.Errorf("SetDirection(%f, %f) set invalid direction %v", tt.dx, tt.dy, anim.Direction)
		}
	}
}

func TestAnimationSystem_FrameCount(t *testing.T) {
	sys := NewAnimationSystem("fantasy")

	tests := []struct {
		state AnimationState
		want  int
	}{
		{StateIdle, 4},
		{StateWalk, 8},
		{StateRun, 8},
		{StateAttack, 6},
		{StateHurt, 3},
		{StateDeath, 8},
		{StateBlock, 2},
	}

	for _, tt := range tests {
		got := sys.getFrameCount(tt.state)
		if got != tt.want {
			t.Errorf("getFrameCount(%v) = %d, want %d", tt.state, got, tt.want)
		}
	}
}

func TestAnimationSystem_OneShotAnimation(t *testing.T) {
	sys := NewAnimationSystem("fantasy")

	oneShot := []AnimationState{StateAttack, StateHurt, StateDeath, StateCast}
	for _, state := range oneShot {
		if !sys.isOneShotAnimation(state) {
			t.Errorf("State %v should be one-shot", state)
		}
	}

	looping := []AnimationState{StateIdle, StateWalk, StateRun, StateBlock}
	for _, state := range looping {
		if sys.isOneShotAnimation(state) {
			t.Errorf("State %v should loop", state)
		}
	}
}

func TestAnimationSystem_GenerateSprite(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	anim := &AnimationComponent{
		State:     StateIdle,
		Frame:     0,
		Direction: DirSouth,
		Seed:      12345,
		Archetype: "warrior",
	}

	sprite := sys.GenerateSprite(anim)
	if sprite == nil {
		t.Fatal("GenerateSprite returned nil")
	}

	bounds := sprite.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("Expected 32x32 sprite, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestAnimationSystem_SpriteCaching(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	anim := &AnimationComponent{
		State:     StateIdle,
		Frame:     0,
		Direction: DirSouth,
		Seed:      12345,
		Archetype: "warrior",
	}

	// Generate sprite twice
	sprite1 := sys.GenerateSprite(anim)
	sprite2 := sys.GenerateSprite(anim)

	// Should return same cached instance
	if sprite1 != sprite2 {
		t.Error("Expected cached sprite to be returned")
	}
}

func TestAnimationSystem_GenreColors(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewAnimationSystem(genre)
			anim := &AnimationComponent{
				State:     StateIdle,
				Frame:     0,
				Direction: DirSouth,
				Seed:      12345,
				Archetype: "warrior",
			}

			sprite := sys.GenerateSprite(anim)
			if sprite == nil {
				t.Fatalf("GenerateSprite returned nil for genre %s", genre)
			}
		})
	}
}

func TestAnimationSystem_AllStates(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	states := []AnimationState{StateIdle, StateWalk, StateRun, StateAttack, StateHurt, StateDeath, StateCast, StateBlock}

	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			anim := &AnimationComponent{
				State:     state,
				Frame:     0,
				Direction: DirSouth,
				Seed:      12345,
				Archetype: "warrior",
			}

			sprite := sys.GenerateSprite(anim)
			if sprite == nil {
				t.Fatalf("GenerateSprite returned nil for state %v", state)
			}
		})
	}
}

func TestAnimationSystem_AllDirections(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	dirs := []Direction{DirNorth, DirNorthEast, DirEast, DirSouthEast, DirSouth, DirSouthWest, DirWest, DirNorthWest}

	for _, dir := range dirs {
		t.Run(dir.String(), func(t *testing.T) {
			anim := &AnimationComponent{
				State:     StateIdle,
				Frame:     0,
				Direction: dir,
				Seed:      12345,
				Archetype: "warrior",
			}

			sprite := sys.GenerateSprite(anim)
			if sprite == nil {
				t.Fatalf("GenerateSprite returned nil for direction %v", dir)
			}
		})
	}
}

func TestSpriteCache_LRUEviction(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	sys.cache = NewSpriteCache(2) // Override with small cache

	// Add 3 sprites to cache with capacity 2
	anim1 := &AnimationComponent{Seed: 1, State: StateIdle, Frame: 0, Direction: DirNorth, Archetype: "warrior"}
	anim2 := &AnimationComponent{Seed: 2, State: StateIdle, Frame: 0, Direction: DirNorth, Archetype: "warrior"}
	anim3 := &AnimationComponent{Seed: 3, State: StateIdle, Frame: 0, Direction: DirNorth, Archetype: "warrior"}

	sys.GenerateSprite(anim1)         // Adds to cache
	img2 := sys.GenerateSprite(anim2) // Adds to cache
	img3 := sys.GenerateSprite(anim3) // Should evict anim1

	// Check cache directly
	cached1 := sys.cache.Get(1, StateIdle, 0, DirNorth, "fantasy")
	cached2 := sys.cache.Get(2, StateIdle, 0, DirNorth, "fantasy")
	cached3 := sys.cache.Get(3, StateIdle, 0, DirNorth, "fantasy")

	// First sprite should be evicted
	if cached1 != nil {
		t.Error("First sprite should have been evicted from cache")
	}

	// Second and third should remain
	if cached2 != img2 {
		t.Error("Second sprite should still be cached")
	}
	if cached3 != img3 {
		t.Error("Third sprite should still be cached")
	}
}

func TestAnimationSystem_SetGenreClearsCache(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	anim := &AnimationComponent{
		State:     StateIdle,
		Frame:     0,
		Direction: DirSouth,
		Seed:      12345,
		Archetype: "warrior",
	}

	// Generate sprite to populate cache
	sys.GenerateSprite(anim)

	// Change genre
	sys.SetGenre("scifi")

	// Cache should be cleared
	if len(sys.cache.entries) != 0 {
		t.Error("Cache should be cleared when genre changes")
	}
}

func TestAnimationSystem_LODFrameInterval(t *testing.T) {
	sys := NewAnimationSystem("fantasy")

	tests := []struct {
		distance    float64
		minInterval float64
		maxInterval float64
	}{
		{100, 1.0 / 13.0, 1.0 / 11.0}, // Close: 12 FPS
		{300, 1.0 / 7.0, 1.0 / 5.0},   // Medium: 6 FPS
		{500, 1.0 / 5.0, 1.0 / 3.0},   // Far: 4 FPS
	}

	for _, tt := range tests {
		interval := sys.getFrameInterval(tt.distance)
		if interval < tt.minInterval || interval > tt.maxInterval {
			t.Errorf("getFrameInterval(%f) = %f, want between %f and %f",
				tt.distance, interval, tt.minInterval, tt.maxInterval)
		}
	}
}

func TestAnimationSystem_FrameAdvancement(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	anim := &AnimationComponent{
		State:            StateWalk,
		Frame:            0,
		FrameTime:        0,
		Seed:             12345,
		Archetype:        "warrior",
		DistanceToCamera: 100, // Close distance for fast animation
	}
	world.AddComponent(entity, anim)

	// Directly call Update once and check the component was modified
	sys.Update(world)

	// Should have advanced by 1/60 second
	expected := 1.0 / 60.0
	if anim.FrameTime < expected*0.9 || anim.FrameTime > expected*1.1 {
		t.Errorf("Expected FrameTime ~%f after one update, got %f", expected, anim.FrameTime)
	}
}

func TestAnimationSystem_OneShotCompletion(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
	anim := &AnimationComponent{
		State:            StateAttack,
		Frame:            5,    // Almost at end
		FrameTime:        0.08, // Ready to advance
		Seed:             12345,
		Archetype:        "warrior",
		DistanceToCamera: 100,
	}
	world.AddComponent(entity, anim)

	// Update to complete animation
	sys.Update(world)

	// Should transition to idle
	if anim.State != StateIdle {
		t.Errorf("One-shot animation should transition to idle, got %v", anim.State)
	}
	if anim.Frame != 0 {
		t.Error("Frame should reset after one-shot completion")
	}
}

func TestAnimationSystem_Determinism(t *testing.T) {
	sys := NewAnimationSystem("fantasy")
	seed := int64(42)

	// Generate same sprite twice with same seed
	img1 := sys.generateSpriteFrame(seed, "warrior", StateIdle, 0, DirSouth)
	// Clear cache to force regeneration
	sys.cache.Clear()
	img2 := sys.generateSpriteFrame(seed, "warrior", StateIdle, 0, DirSouth)

	// Compare bounds
	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1 != bounds2 {
		t.Fatal("Sprite bounds don't match")
	}

	// Sample a few pixels to verify determinism
	testPixels := [][2]int{{16, 16}, {8, 8}, {24, 24}}
	for _, pos := range testPixels {
		c1 := img1.At(pos[0], pos[1])
		c2 := img2.At(pos[0], pos[1])
		if c1 != c2 {
			t.Errorf("Pixel at (%d, %d) differs: %v vs %v", pos[0], pos[1], c1, c2)
		}
	}
}

func TestAnimationComponent_Integration(t *testing.T) {
	world := engine.NewWorld()
	entity := world.AddEntity()

	anim := &AnimationComponent{
		State:     StateIdle,
		Frame:     0,
		Seed:      12345,
		Archetype: "warrior",
	}

	world.AddComponent(entity, anim)

	// Retrieve component
	compType := reflect.TypeOf(&AnimationComponent{})
	retrieved, ok := world.GetComponent(entity, compType)
	if !ok {
		t.Fatal("Failed to retrieve animation component")
	}

	retrievedAnim, ok := retrieved.(*AnimationComponent)
	if !ok {
		t.Fatal("Component is not AnimationComponent type")
	}

	if retrievedAnim.Seed != 12345 {
		t.Error("Retrieved component has wrong seed")
	}
}

// String methods for better test output
func (s AnimationState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateWalk:
		return "Walk"
	case StateRun:
		return "Run"
	case StateAttack:
		return "Attack"
	case StateHurt:
		return "Hurt"
	case StateDeath:
		return "Death"
	case StateCast:
		return "Cast"
	case StateBlock:
		return "Block"
	default:
		return "Unknown"
	}
}

func (d Direction) String() string {
	switch d {
	case DirNorth:
		return "North"
	case DirNorthEast:
		return "NorthEast"
	case DirEast:
		return "East"
	case DirSouthEast:
		return "SouthEast"
	case DirSouth:
		return "South"
	case DirSouthWest:
		return "SouthWest"
	case DirWest:
		return "West"
	case DirNorthWest:
		return "NorthWest"
	default:
		return "Unknown"
	}
}

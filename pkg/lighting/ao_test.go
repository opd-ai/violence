package lighting

import (
	"math"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestAOComponent_Creation(t *testing.T) {
	ao := NewAOComponent(2.0)

	if ao.SampleRadius != 2.0 {
		t.Errorf("Expected SampleRadius 2.0, got %f", ao.SampleRadius)
	}

	if !ao.CastsOcclusion {
		t.Error("Expected CastsOcclusion to be true by default")
	}

	if ao.Overall != 1.0 {
		t.Errorf("Expected Overall occlusion 1.0, got %f", ao.Overall)
	}

	if !ao.needsUpdate {
		t.Error("Expected needsUpdate to be true initially")
	}
}

func TestAOComponent_Invalidate(t *testing.T) {
	ao := NewAOComponent(2.0)
	ao.needsUpdate = false

	ao.Invalidate()

	if !ao.needsUpdate {
		t.Error("Expected needsUpdate to be true after Invalidate")
	}
}

func TestAOComponent_GetOcclusionAt(t *testing.T) {
	ao := NewAOComponent(2.0)
	ao.East = 1.0
	ao.NorthEast = 0.5
	ao.North = 0.0
	ao.NorthWest = 0.5
	ao.West = 1.0
	ao.SouthWest = 0.5
	ao.South = 0.0
	ao.SouthEast = 0.5

	tests := []struct {
		name     string
		angle    float64
		expected float64
		delta    float64
	}{
		{"East", 0.0, 1.0, 0.01},
		{"NorthEast", math.Pi / 4, 0.5, 0.01},
		{"North", math.Pi / 2, 0.0, 0.01},
		{"West", math.Pi, 1.0, 0.01},
		{"South", 3 * math.Pi / 2, 0.0, 0.01},
		{"Between E and NE", math.Pi / 8, 0.75, 0.05}, // Should interpolate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ao.GetOcclusionAt(tt.angle)
			if math.Abs(result-tt.expected) > tt.delta {
				t.Errorf("Expected %f, got %f (delta %f)", tt.expected, result, math.Abs(result-tt.expected))
			}
		})
	}
}

func TestAOComponent_Type(t *testing.T) {
	ao := NewAOComponent(2.0)
	if ao.Type() != "AmbientOcclusion" {
		t.Errorf("Expected Type 'AmbientOcclusion', got '%s'", ao.Type())
	}
}

func TestAOSystem_Creation(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewAOSystem(genre)

			if sys.genre != genre {
				t.Errorf("Expected genre %s, got %s", genre, sys.genre)
			}

			if sys.baseIntensity <= 0 || sys.baseIntensity > 1.0 {
				t.Errorf("Invalid baseIntensity: %f", sys.baseIntensity)
			}

			if sys.wallOcclusion <= 0 || sys.wallOcclusion > 1.0 {
				t.Errorf("Invalid wallOcclusion: %f", sys.wallOcclusion)
			}

			if sys.entityOcclusion < 0 || sys.entityOcclusion > 1.0 {
				t.Errorf("Invalid entityOcclusion: %f", sys.entityOcclusion)
			}
		})
	}
}

func TestAOSystem_SetGenre(t *testing.T) {
	sys := NewAOSystem("fantasy")
	originalIntensity := sys.baseIntensity

	sys.SetGenre("horror")

	if sys.genre != "horror" {
		t.Errorf("Expected genre 'horror', got '%s'", sys.genre)
	}

	// Horror should have different intensity than fantasy
	if sys.baseIntensity == originalIntensity {
		t.Error("Expected baseIntensity to change when genre changes")
	}
}

func TestAOSystem_Update(t *testing.T) {
	world := engine.NewWorld()
	sys := NewAOSystem("fantasy")

	// Create an entity with AO and position components
	entity := world.AddEntity()
	pos := &PositionComponent{X: 5.0, Y: 5.0}
	ao := NewAOComponent(2.0)

	world.AddComponent(entity, pos)
	world.AddComponent(entity, ao)

	// Initial update should process the entity
	initialUpdate := ao.needsUpdate
	if !initialUpdate {
		t.Error("Expected needsUpdate to be true initially")
	}

	// Run enough ticks to trigger update
	for i := 0; i < sys.updateInterval; i++ {
		sys.Update(world)
	}

	// After update, needsUpdate should be false
	aoComp, ok := world.GetComponent(entity, reflect.TypeOf(&AOComponent{}))
	if !ok {
		t.Fatal("Failed to retrieve AO component")
	}
	updatedAO := aoComp.(*AOComponent)

	if updatedAO.needsUpdate {
		t.Error("Expected needsUpdate to be false after update")
	}

	// Occlusion values should be set
	if updatedAO.Overall < 0 || updatedAO.Overall > 1.0 {
		t.Errorf("Invalid Overall occlusion: %f", updatedAO.Overall)
	}
}

func TestAOSystem_InvalidateAll(t *testing.T) {
	world := engine.NewWorld()
	sys := NewAOSystem("fantasy")

	// Create multiple entities with AO
	for i := 0; i < 3; i++ {
		entity := world.AddEntity()
		pos := &PositionComponent{X: float64(i), Y: float64(i)}
		ao := NewAOComponent(2.0)
		ao.needsUpdate = false // Mark as valid

		world.AddComponent(entity, pos)
		world.AddComponent(entity, ao)
	}

	// Invalidate all
	sys.InvalidateAll(world)

	// Check that all AO components are now invalid
	aoType := reflect.TypeOf(&AOComponent{})
	entities := world.Query(aoType)

	for _, entity := range entities {
		aoComp, ok := world.GetComponent(entity, aoType)
		if !ok {
			t.Fatal("Failed to retrieve AO component")
		}
		ao := aoComp.(*AOComponent)

		if !ao.needsUpdate {
			t.Error("Expected needsUpdate to be true after InvalidateAll")
		}
	}
}

func TestAOSystem_GenreSettings(t *testing.T) {
	tests := []struct {
		genre            string
		expectStrongAO   bool
		expectWeakEntity bool
	}{
		{"fantasy", true, true},
		{"horror", true, false}, // Horror has strong entity occlusion
		{"scifi", false, true},
		{"cyberpunk", true, false}, // Cyberpunk has moderate entity occlusion
		{"postapoc", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewAOSystem(tt.genre)

			if tt.expectStrongAO && sys.baseIntensity < 0.6 {
				t.Errorf("Expected strong AO for %s, got %f", tt.genre, sys.baseIntensity)
			}

			if !tt.expectStrongAO && sys.baseIntensity > 0.6 {
				t.Errorf("Expected weak AO for %s, got %f", tt.genre, sys.baseIntensity)
			}

			if tt.expectWeakEntity && sys.entityOcclusion > 0.35 {
				t.Errorf("Expected weak entity occlusion for %s, got %f", tt.genre, sys.entityOcclusion)
			}
		})
	}
}

func TestAOSystem_IsWall(t *testing.T) {
	sys := NewAOSystem("fantasy")

	tests := []struct {
		name     string
		x, y     float64
		expected bool
	}{
		{"Center of cell", 5.5, 5.5, false},
		{"Near left edge", 5.05, 5.5, true},
		{"Near right edge", 5.95, 5.5, true},
		{"Near top edge", 5.5, 5.05, true},
		{"Near bottom edge", 5.5, 5.95, true},
		{"Far from edges", 5.3, 5.7, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.isWall(tt.x, tt.y)
			if result != tt.expected {
				t.Errorf("isWall(%f, %f) = %v, expected %v", tt.x, tt.y, result, tt.expected)
			}
		})
	}
}

func TestAOSystem_SampleDirection(t *testing.T) {
	world := engine.NewWorld()
	sys := NewAOSystem("fantasy")

	entity := world.AddEntity()
	pos := &PositionComponent{X: 5.0, Y: 5.0}
	world.AddComponent(entity, pos)

	positionType := reflect.TypeOf(&PositionComponent{})
	colliderType := reflect.TypeOf(&ColliderComponent{})

	// Sample in a direction
	occlusion := sys.sampleDirection(world, entity, pos, 0.0, 2.0, positionType, colliderType)

	// Should return some occlusion value in valid range
	if occlusion < 0.0 || occlusion > 1.0 {
		t.Errorf("Invalid occlusion value: %f", occlusion)
	}
}

func TestAOSystem_CalculateOcclusion(t *testing.T) {
	world := engine.NewWorld()
	sys := NewAOSystem("fantasy")

	entity := world.AddEntity()
	pos := &PositionComponent{X: 5.0, Y: 5.0}
	ao := NewAOComponent(2.0)

	world.AddComponent(entity, pos)
	world.AddComponent(entity, ao)

	positionType := reflect.TypeOf(&PositionComponent{})
	colliderType := reflect.TypeOf(&ColliderComponent{})

	sys.calculateOcclusion(world, entity, pos, ao, positionType, colliderType)

	// Check that all directional values are set and valid
	directions := []float64{
		ao.North, ao.South, ao.East, ao.West,
		ao.NorthEast, ao.NorthWest, ao.SouthEast, ao.SouthWest,
		ao.Overall,
	}

	for i, val := range directions {
		if val < 0.0 || val > 1.0 {
			t.Errorf("Invalid occlusion value at index %d: %f", i, val)
		}
	}
}

func TestAOSystem_UpdateThrottle(t *testing.T) {
	world := engine.NewWorld()
	sys := NewAOSystem("fantasy")

	entity := world.AddEntity()
	pos := &PositionComponent{X: 5.0, Y: 5.0}
	ao := NewAOComponent(2.0)

	world.AddComponent(entity, pos)
	world.AddComponent(entity, ao)

	// First update should be skipped due to throttle
	sys.Update(world)

	aoComp, _ := world.GetComponent(entity, reflect.TypeOf(&AOComponent{}))
	updatedAO := aoComp.(*AOComponent)

	if !updatedAO.needsUpdate {
		t.Error("Expected update to be skipped on first tick")
	}

	// Update until interval is reached
	for i := 0; i < sys.updateInterval-1; i++ {
		sys.Update(world)
	}

	aoComp, _ = world.GetComponent(entity, reflect.TypeOf(&AOComponent{}))
	updatedAO = aoComp.(*AOComponent)

	if updatedAO.needsUpdate {
		t.Error("Expected update to occur after interval")
	}
}

// Benchmark AO calculation
func BenchmarkAOSystem_Update(b *testing.B) {
	world := engine.NewWorld()
	sys := NewAOSystem("fantasy")

	// Create 50 entities with AO
	for i := 0; i < 50; i++ {
		entity := world.AddEntity()
		pos := &PositionComponent{X: float64(i % 10), Y: float64(i / 10)}
		ao := NewAOComponent(2.0)

		world.AddComponent(entity, pos)
		world.AddComponent(entity, ao)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world)
	}
}

// Test coverage: ColliderComponent
func TestColliderComponent_Type(t *testing.T) {
	c := &ColliderComponent{Radius: 1.0, Solid: true}
	if c.Type() != "Collider" {
		t.Errorf("Expected Type 'Collider', got '%s'", c.Type())
	}
}

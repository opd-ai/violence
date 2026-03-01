package lighting

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// MockWorld is a simple wrapper around engine.World for testing.
func NewMockWorld() *engine.World {
	return engine.NewWorld()
}

func TestNewLightingSystem(t *testing.T) {
	sys := NewLightingSystem("fantasy")
	if sys == nil {
		t.Fatal("NewLightingSystem returned nil")
	}

	if sys.genre != "fantasy" {
		t.Errorf("expected genre = 'fantasy', got %s", sys.genre)
	}

	if sys.ambientIntensity == 0 {
		t.Error("ambient intensity should be set by genre preset")
	}
}

func TestGenreAmbientPresets(t *testing.T) {
	tests := []struct {
		genre           string
		expectIntensity bool
	}{
		{"fantasy", true},
		{"scifi", true},
		{"horror", true},
		{"cyberpunk", true},
		{"postapoc", true},
		{"unknown", true}, // Should fall back to default
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewLightingSystem(tt.genre)
			r, g, b, intensity := sys.GetAmbient()

			if tt.expectIntensity && intensity == 0 {
				t.Errorf("genre %s should have non-zero ambient intensity", tt.genre)
			}

			if r < 0 || r > 1 || g < 0 || g > 1 || b < 0 || b > 1 {
				t.Errorf("ambient color out of range: r=%f, g=%f, b=%f", r, g, b)
			}
		})
	}
}

func TestLightingSystemSetGenre(t *testing.T) {
	sys := NewLightingSystem("fantasy")

	// Get initial ambient
	r1, g1, b1, i1 := sys.GetAmbient()

	// Change genre
	sys.SetGenre("horror")
	r2, g2, b2, i2 := sys.GetAmbient()

	// Ambient should have changed
	if r1 == r2 && g1 == g2 && b1 == b2 && i1 == i2 {
		t.Error("SetGenre should change ambient lighting")
	}

	if sys.genre != "horror" {
		t.Errorf("expected genre = 'horror', got %s", sys.genre)
	}
}

func TestUpdateBasicLight(t *testing.T) {
	sys := NewLightingSystem("fantasy")
	world := NewMockWorld()

	// Create entity with light
	entity := world.AddEntity()
	preset := LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.6,
		B:         0.2,
		Flicker:   false,
	}
	light := NewLightComponent(preset, 42)
	world.AddComponent(entity, light)

	// Update system
	sys.Update(world)

	// Light should still be enabled
	if !light.Enabled {
		t.Error("light should remain enabled after update")
	}

	// Age should have increased
	if light.CurrentAge == 0 {
		t.Error("light age should increase after update")
	}
}

func TestUpdateTemporaryLight(t *testing.T) {
	sys := NewLightingSystem("fantasy")
	world := NewMockWorld()

	entity := world.AddEntity()
	preset := LightPreset{
		Name:      "flash",
		Radius:    3.0,
		Intensity: 1.0,
		R:         1.0,
		G:         1.0,
		B:         1.0,
		Flicker:   false,
	}
	light := NewTemporaryLight(preset, 42, 0.001) // Very short lifetime
	world.AddComponent(entity, light)

	// Update many times to exceed lifetime
	for i := 0; i < 100; i++ {
		sys.Update(world)
	}

	// Light should be disabled
	if light.Enabled {
		t.Error("temporary light should be disabled after lifetime expires")
	}
}

func TestUpdateAttachedLight(t *testing.T) {
	sys := NewLightingSystem("fantasy")
	world := NewMockWorld()

	entity := world.AddEntity()

	// Add position component
	pos := &PositionComponent{X: 10.0, Y: 20.0}
	world.AddComponent(entity, pos)

	// Add attached light with offset
	preset := LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.6,
		B:         0.2,
		Flicker:   false,
	}
	light := NewAttachedLight(preset, 42, 1.0, 2.0)
	world.AddComponent(entity, light)

	// Update system
	sys.Update(world)

	// Light position should match entity position + offset
	expectedX := 10.0 + 1.0
	expectedY := 20.0 + 2.0

	if math.Abs(light.X-expectedX) > 0.01 {
		t.Errorf("expected light.X = %f, got %f", expectedX, light.X)
	}

	if math.Abs(light.Y-expectedY) > 0.01 {
		t.Errorf("expected light.Y = %f, got %f", expectedY, light.Y)
	}
}

func TestUpdatePulsingLight(t *testing.T) {
	sys := NewLightingSystem("fantasy")
	world := NewMockWorld()

	entity := world.AddEntity()
	preset := LightPreset{
		Name:      "alarm",
		Radius:    4.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.2,
		B:         0.2,
		Flicker:   false,
	}
	light := NewPulsingLight(preset, 42, 1.0)
	world.AddComponent(entity, light)

	initialPhase := light.PulsePhase

	// Update multiple times
	for i := 0; i < 10; i++ {
		sys.Update(world)
	}

	// Pulse phase should have changed
	if light.PulsePhase == initialPhase {
		t.Error("pulse phase should update over time")
	}

	// Phase should wrap around 2π
	if light.PulsePhase < 0 || light.PulsePhase > 2*math.Pi {
		t.Errorf("pulse phase out of range: %f", light.PulsePhase)
	}
}

func TestCollectLights(t *testing.T) {
	sys := NewLightingSystem("fantasy")
	world := NewMockWorld()

	// Create multiple entities with lights
	for i := 0; i < 3; i++ {
		entity := world.AddEntity()
		preset := LightPreset{
			Name:      "torch",
			Radius:    5.0,
			Intensity: 0.8,
			R:         1.0,
			G:         0.6,
			B:         0.2,
			Flicker:   false,
		}
		light := NewLightComponent(preset, int64(i))
		light.X = float64(i * 10)
		light.Y = float64(i * 10)
		world.AddComponent(entity, light)
	}

	// Collect lights
	lights := sys.CollectLights(world)

	if len(lights) != 3 {
		t.Errorf("expected 3 lights, got %d", len(lights))
	}

	// Verify light properties
	for i, light := range lights {
		if light.Radius != 5.0 {
			t.Errorf("light %d: expected Radius = 5.0, got %f", i, light.Radius)
		}
	}
}

func TestCollectLightsSkipsDisabled(t *testing.T) {
	sys := NewLightingSystem("fantasy")
	world := NewMockWorld()

	// Create enabled light
	entity1 := world.AddEntity()
	preset := LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.6,
		B:         0.2,
		Flicker:   false,
	}
	light1 := NewLightComponent(preset, 1)
	world.AddComponent(entity1, light1)

	// Create disabled light
	entity2 := world.AddEntity()
	light2 := NewLightComponent(preset, 2)
	light2.Enabled = false
	world.AddComponent(entity2, light2)

	// Collect lights
	lights := sys.CollectLights(world)

	// Should only collect enabled light
	if len(lights) != 1 {
		t.Errorf("expected 1 enabled light, got %d", len(lights))
	}
}

func TestGetEffectiveIntensity(t *testing.T) {
	preset := LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.6,
		B:         0.2,
		Flicker:   false,
	}
	light := NewLightComponent(preset, 42)
	light.Intensity = 0.8

	// Test basic intensity
	intensity := GetEffectiveIntensity(light, 0)
	if intensity < 0.0 || intensity > 1.0 {
		t.Errorf("intensity out of range: %f", intensity)
	}

	// Test pulsing intensity varies
	light.Pulsing = true
	light.PulsePhase = 0 // sin(0) = 0
	intensity1 := GetEffectiveIntensity(light, 0)

	light.PulsePhase = math.Pi / 2 // sin(π/2) = 1
	intensity2 := GetEffectiveIntensity(light, 0)

	if math.Abs(intensity1-intensity2) < 0.01 {
		t.Errorf("pulsing light intensity should vary with phase: got %f and %f", intensity1, intensity2)
	}
}

func TestPositionComponentType(t *testing.T) {
	pos := &PositionComponent{}
	if pos.Type() != "Position" {
		t.Errorf("expected Type() = 'Position', got %s", pos.Type())
	}
}

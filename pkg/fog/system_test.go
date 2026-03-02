package fog

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	s := NewSystem("fantasy")
	if s == nil {
		t.Fatal("NewSystem returned nil")
	}
	if !s.IsEnabled() {
		t.Error("fog should be enabled by default")
	}
	if s.genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %q", s.genre)
	}
}

func TestGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			s := NewSystem(genre)
			if s.fogStart >= s.fogEnd {
				t.Errorf("%s: fogStart (%.1f) should be < fogEnd (%.1f)",
					genre, s.fogStart, s.fogEnd)
			}
			if s.fogDensity < 0.0 || s.fogDensity > 1.0 {
				t.Errorf("%s: fogDensity (%.2f) out of range [0.0-1.0]",
					genre, s.fogDensity)
			}
			if s.fogColor.R == 0 && s.fogColor.G == 0 && s.fogColor.B == 0 {
				t.Errorf("%s: fog color should not be pure black", genre)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	s := NewSystem("fantasy")
	originalColor := s.fogColor

	s.SetGenre("horror")
	if s.genre != "horror" {
		t.Errorf("expected genre 'horror', got %q", s.genre)
	}
	if s.fogColor == originalColor {
		t.Error("fog color should change when genre changes")
	}
}

func TestComputeFogDensity(t *testing.T) {
	s := NewSystem("fantasy")
	s.fogStart = 10.0
	s.fogEnd = 20.0
	s.fogDensity = 1.0

	tests := []struct {
		name     string
		distance float64
		wantMin  float64
		wantMax  float64
	}{
		{"before fog", 5.0, 0.0, 0.0},
		{"at fog start", 10.0, 0.0, 0.0},
		{"mid fog", 15.0, 0.2, 0.8},
		{"at fog end", 20.0, 0.99, 1.0},
		{"beyond fog", 30.0, 0.99, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.computeFogDensity(tt.distance)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("density %.3f not in range [%.3f, %.3f]",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFalloffTypes(t *testing.T) {
	falloffTypes := []string{"linear", "exponential", "exponential_squared"}

	for _, falloff := range falloffTypes {
		t.Run(falloff, func(t *testing.T) {
			s := NewSystem("fantasy")
			s.falloffType = falloff
			s.fogStart = 0.0
			s.fogEnd = 10.0
			s.fogDensity = 1.0

			// Test mid-range density
			density := s.computeFogDensity(5.0)
			if density < 0.0 || density > 1.0 {
				t.Errorf("%s: density %.3f out of range [0.0-1.0]", falloff, density)
			}

			// Density should increase with distance
			d1 := s.computeFogDensity(2.0)
			d2 := s.computeFogDensity(5.0)
			d3 := s.computeFogDensity(8.0)
			if !(d1 < d2 && d2 < d3) {
				t.Errorf("%s: density not monotonically increasing: %.3f, %.3f, %.3f",
					falloff, d1, d2, d3)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	world := engine.NewWorld()
	s := NewSystem("fantasy")
	s.SetCamera(0, 0)

	// Create entity with position
	e := world.AddEntity()
	world.AddComponent(e, &engine.Position{X: 10, Y: 0})

	// Run update
	s.Update(world)

	// Check fog component was added
	fogType := reflect.TypeOf(&Component{})
	fogComp, hasFog := world.GetComponent(e, fogType)
	if !hasFog {
		t.Fatal("fog component not added")
	}

	fog, ok := fogComp.(*Component)
	if !ok {
		t.Fatal("fog component has wrong type")
	}

	// Check distance calculation
	expectedDist := 10.0
	if fog.DistanceFromCamera != expectedDist {
		t.Errorf("distance = %.1f, want %.1f", fog.DistanceFromCamera, expectedDist)
	}

	// Check tint is set
	if fog.Tint[0] == 0 && fog.Tint[1] == 0 && fog.Tint[2] == 0 {
		t.Error("tint should not be zero")
	}
}

func TestSetCamera(t *testing.T) {
	world := engine.NewWorld()
	s := NewSystem("fantasy")
	fogType := reflect.TypeOf(&Component{})

	// Create entity at (10, 10)
	e := world.AddEntity()
	world.AddComponent(e, &engine.Position{X: 10, Y: 10})

	// Camera at origin
	s.SetCamera(0, 0)
	s.Update(world)
	fogComp1, _ := world.GetComponent(e, fogType)
	fog1 := fogComp1.(*Component)
	dist1 := fog1.DistanceFromCamera

	// Move camera closer
	s.SetCamera(5, 5)
	s.Update(world)
	fogComp2, _ := world.GetComponent(e, fogType)
	fog2 := fogComp2.(*Component)
	dist2 := fog2.DistanceFromCamera

	if dist2 >= dist1 {
		t.Errorf("distance should decrease when camera moves closer: %.1f >= %.1f",
			dist2, dist1)
	}
}

func TestSetEnabled(t *testing.T) {
	world := engine.NewWorld()
	s := NewSystem("fantasy")
	s.SetCamera(0, 0)
	fogType := reflect.TypeOf(&Component{})

	e := world.AddEntity()
	world.AddComponent(e, &engine.Position{X: 100, Y: 100})

	// Disable fog
	s.SetEnabled(false)
	s.Update(world)

	// Component should not be added when disabled
	_, hasFog := world.GetComponent(e, fogType)
	if hasFog {
		t.Error("fog component should not be added when system disabled")
	}

	// Re-enable and update
	s.SetEnabled(true)
	s.Update(world)

	_, hasFog = world.GetComponent(e, fogType)
	if !hasFog {
		t.Error("fog component should be added when system enabled")
	}
}

func TestSetParameters(t *testing.T) {
	s := NewSystem("fantasy")
	newColor := color.RGBA{100, 150, 200, 255}

	s.SetParameters(5.0, 15.0, 0.8, newColor)

	if s.fogStart != 5.0 {
		t.Errorf("fogStart = %.1f, want 5.0", s.fogStart)
	}
	if s.fogEnd != 15.0 {
		t.Errorf("fogEnd = %.1f, want 15.0", s.fogEnd)
	}
	if s.fogDensity != 0.8 {
		t.Errorf("fogDensity = %.1f, want 0.8", s.fogDensity)
	}
	if s.fogColor != newColor {
		t.Errorf("fogColor = %v, want %v", s.fogColor, newColor)
	}
}

func TestVisibilityThreshold(t *testing.T) {
	world := engine.NewWorld()
	s := NewSystem("horror") // Heavy fog for testing
	s.SetCamera(0, 0)
	fogType := reflect.TypeOf(&Component{})

	// Entity far away should be marked invisible
	e := world.AddEntity()
	world.AddComponent(e, &engine.Position{X: 100, Y: 100})

	s.Update(world)

	fogComp, _ := world.GetComponent(e, fogType)
	fog := fogComp.(*Component)
	if fog.FogDensity >= 0.95 && fog.Visible {
		t.Error("heavily obscured entity should be marked invisible")
	}
}

func BenchmarkUpdate(b *testing.B) {
	world := engine.NewWorld()
	s := NewSystem("fantasy")
	s.SetCamera(0, 0)

	// Create 100 entities
	for i := 0; i < 100; i++ {
		e := world.AddEntity()
		world.AddComponent(e, &engine.Position{
			X: float64(i % 10),
			Y: float64(i / 10),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update(world)
	}
}

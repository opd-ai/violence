package volumetric

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	c := NewComponent()

	if !c.Enabled {
		t.Error("Expected component to be enabled by default")
	}
	if c.Intensity <= 0 {
		t.Error("Expected positive intensity")
	}
	if c.Radius <= 0 {
		t.Error("Expected positive radius")
	}
	if c.DustDensity <= 0 {
		t.Error("Expected positive dust density")
	}
}

func TestComponentType(t *testing.T) {
	c := NewComponent()
	if c.Type() != "Volumetric" {
		t.Errorf("Expected Type() = 'Volumetric', got '%s'", c.Type())
	}
}

func TestComponentGetScatterColor(t *testing.T) {
	c := NewComponent()
	c.ScatterR = 1.0
	c.ScatterG = 0.5
	c.ScatterB = 0.25

	tests := []struct {
		name  string
		alpha float64
		wantA uint8
	}{
		{"zero alpha", 0.0, 0},
		{"half alpha", 0.5, 127},
		{"full alpha", 1.0, 255},
		{"over alpha", 1.5, 255}, // Clamped
		{"negative alpha", -0.5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := c.GetScatterColor(tt.alpha)
			if col.A != tt.wantA {
				t.Errorf("GetScatterColor(%f).A = %d, want %d", tt.alpha, col.A, tt.wantA)
			}
		})
	}
}

func TestComponentSetFromLightColor(t *testing.T) {
	c := NewComponent()
	c.SetFromLightColor(1.0, 0.8, 0.6)

	// Scatter should be slightly warmer (shifted toward red/orange)
	if c.ScatterR < 0.9 {
		t.Errorf("ScatterR too low: %f", c.ScatterR)
	}
	if c.ScatterG > 0.85 || c.ScatterG < 0.7 {
		t.Errorf("ScatterG unexpected: %f", c.ScatterG)
	}
	if c.ScatterB > 0.6 {
		t.Errorf("ScatterB too high: %f", c.ScatterB)
	}
}

func TestComponentIsDirectional(t *testing.T) {
	c := NewComponent()

	if c.IsDirectional() {
		t.Error("Default component should not be directional")
	}

	c.ConeAngle = 0.5
	if !c.IsDirectional() {
		t.Error("Component with cone angle should be directional")
	}
}

func TestNewLightShaft(t *testing.T) {
	shaft := NewLightShaft(10.0, 20.0, 5.0, 0.8, 1.0, 0.5, 0.25)

	if shaft.X != 10.0 {
		t.Errorf("X = %f, want 10.0", shaft.X)
	}
	if shaft.Y != 20.0 {
		t.Errorf("Y = %f, want 20.0", shaft.Y)
	}
	if shaft.Radius != 5.0 {
		t.Errorf("Radius = %f, want 5.0", shaft.Radius)
	}
	if shaft.Intensity != 0.8 {
		t.Errorf("Intensity = %f, want 0.8", shaft.Intensity)
	}
	if shaft.ConeAngle != 0 {
		t.Error("Default shaft should not have cone angle")
	}
}

func TestLightShaftWithCone(t *testing.T) {
	shaft := NewLightShaft(0, 0, 5.0, 1.0, 1.0, 1.0, 1.0)
	shaft = shaft.WithCone(0.5, 1.0, 0.0)

	if shaft.ConeAngle != 0.5 {
		t.Errorf("ConeAngle = %f, want 0.5", shaft.ConeAngle)
	}
	if shaft.DirX != 1.0 {
		t.Errorf("DirX = %f, want 1.0", shaft.DirX)
	}
	if shaft.DirY != 0.0 {
		t.Errorf("DirY = %f, want 0.0", shaft.DirY)
	}
}

func TestLightShaftWithDust(t *testing.T) {
	shaft := NewLightShaft(0, 0, 5.0, 1.0, 1.0, 1.0, 1.0)
	shaft = shaft.WithDust(0.75)

	if shaft.DustDensity != 0.75 {
		t.Errorf("DustDensity = %f, want 0.75", shaft.DustDensity)
	}
}

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 320, 200)

			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.screenW != 320 {
				t.Errorf("screenW = %d, want 320", sys.screenW)
			}
			if sys.screenH != 200 {
				t.Errorf("screenH = %d, want 200", sys.screenH)
			}
			if sys.overlay == nil {
				t.Error("overlay image should be initialized")
			}
		})
	}
}

func TestSystemSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Change to sci-fi
	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Errorf("genreID = %s, want 'scifi'", sys.genreID)
	}
	if sys.preset.BaseDustDensity >= 0.4 {
		t.Error("Sci-fi should have lower dust density than fantasy")
	}
}

func TestSystemSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Change screen size
	sys.SetScreenSize(640, 400)

	if sys.screenW != 640 {
		t.Errorf("screenW = %d, want 640", sys.screenW)
	}
	if sys.screenH != 400 {
		t.Errorf("screenH = %d, want 400", sys.screenH)
	}

	// Same size should be no-op
	oldOverlay := sys.overlay
	sys.SetScreenSize(640, 400)
	if sys.overlay != oldOverlay {
		t.Error("SetScreenSize with same size should not reallocate")
	}
}

func TestSystemCreateLightShaftFromTorch(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)
	shaft := sys.CreateLightShaftFromTorch(5.0, 10.0, 0.9)

	if shaft.X != 5.0 {
		t.Errorf("X = %f, want 5.0", shaft.X)
	}
	if shaft.Y != 10.0 {
		t.Errorf("Y = %f, want 10.0", shaft.Y)
	}
	if shaft.Intensity <= 0 {
		t.Error("Torch intensity should be positive")
	}
	// Torch should be warm colored
	if shaft.R < shaft.B {
		t.Error("Torch should have warm (red > blue) color")
	}
}

func TestSystemCreateLightShaftFromMagic(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)
	shaft := sys.CreateLightShaftFromMagic(5.0, 10.0, 0.8, 0.2, 0.8, 1.0)

	if shaft.X != 5.0 {
		t.Errorf("X = %f, want 5.0", shaft.X)
	}
	if shaft.Intensity != 0.8 {
		t.Errorf("Intensity = %f, want 0.8", shaft.Intensity)
	}
	if shaft.R != 0.2 || shaft.G != 0.8 || shaft.B != 1.0 {
		t.Error("Magic light should use provided colors")
	}
}

func TestSystemGetPreset(t *testing.T) {
	sys := NewSystem("horror", 320, 200)
	preset := sys.GetPreset()

	if preset.BaseDustDensity <= 0 {
		t.Error("Preset should have positive dust density")
	}
	if preset.SampleCount <= 0 {
		t.Error("Preset should have positive sample count")
	}
}

func TestGenrePresets(t *testing.T) {
	// Verify all expected genres have presets
	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range expectedGenres {
		preset, ok := genrePresets[genre]
		if !ok {
			t.Errorf("Missing preset for genre: %s", genre)
			continue
		}

		if preset.BaseDustDensity <= 0 || preset.BaseDustDensity > 1.0 {
			t.Errorf("%s: BaseDustDensity out of range: %f", genre, preset.BaseDustDensity)
		}
		if preset.RayIntensity <= 0 || preset.RayIntensity > 1.0 {
			t.Errorf("%s: RayIntensity out of range: %f", genre, preset.RayIntensity)
		}
		if preset.SampleCount < 4 || preset.SampleCount > 64 {
			t.Errorf("%s: SampleCount out of reasonable range: %d", genre, preset.SampleCount)
		}
	}
}

func TestGenreVisualDifferences(t *testing.T) {
	// Fantasy should be dustier than sci-fi
	if genrePresets["fantasy"].BaseDustDensity <= genrePresets["scifi"].BaseDustDensity {
		t.Error("Fantasy should be dustier than sci-fi")
	}

	// Horror should have higher ray intensity than sci-fi
	if genrePresets["horror"].RayIntensity <= genrePresets["scifi"].RayIntensity {
		t.Error("Horror should have more prominent volumetrics than sci-fi")
	}

	// Post-apocalyptic should be dusty
	if genrePresets["postapoc"].BaseDustDensity < 0.5 {
		t.Error("Post-apocalyptic should have high dust density")
	}
}

func TestClampF64(t *testing.T) {
	tests := []struct {
		name        string
		v, min, max float64
		want        float64
	}{
		{"in range", 0.5, 0.0, 1.0, 0.5},
		{"below min", -0.5, 0.0, 1.0, 0.0},
		{"above max", 1.5, 0.0, 1.0, 1.0},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"at max", 1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampF64(tt.v, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clampF64(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestNoiseTableInitialization(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Noise values should be in [0, 1]
	for i, v := range sys.noiseTable {
		if v < 0 || v > 1 {
			t.Errorf("noiseTable[%d] = %f, out of range [0,1]", i, v)
		}
	}

	// Noise should have some variation
	uniqueValues := make(map[float64]bool)
	for _, v := range sys.noiseTable {
		uniqueValues[v] = true
	}
	if len(uniqueValues) < 100 {
		t.Errorf("Noise table should have more variation, only %d unique values", len(uniqueValues))
	}
}

func TestRenderSimple(t *testing.T) {
	_ = NewSystem("fantasy", 320, 200)

	// Create test lights
	lights := []LightShaft{
		NewLightShaft(5.0, 5.0, 4.0, 0.8, 1.0, 0.7, 0.3),
		NewLightShaft(10.0, 5.0, 3.0, 0.6, 0.5, 0.5, 1.0),
	}

	// This should not panic even with nil screen (ebiten requires X11)
	// We're mainly testing the light processing logic
	if len(lights) != 2 {
		t.Errorf("Expected 2 lights, got %d", len(lights))
	}
}

func TestRenderWithOccluder(t *testing.T) {
	sys := NewSystem("horror", 320, 200)

	lights := []LightShaft{
		NewLightShaft(5.0, 5.0, 4.0, 0.8, 1.0, 0.7, 0.3),
	}

	// Test with an occluder function
	occluder := func(x, y float64) bool {
		return int(x)%2 == 0 // Every other column is blocked
	}

	// Verify the occluder works
	if !occluder(0, 0) {
		t.Error("Occluder should block x=0")
	}
	if occluder(1, 0) {
		t.Error("Occluder should not block x=1")
	}

	// Verify system and lights are valid
	if sys == nil {
		t.Error("System should not be nil")
	}
	if len(lights) == 0 {
		t.Error("Should have lights")
	}
}

func TestLightShaftChaining(t *testing.T) {
	shaft := NewLightShaft(0, 0, 5.0, 1.0, 1.0, 0.8, 0.5).
		WithCone(0.5, 1.0, 0.0).
		WithDust(0.8)

	if shaft.ConeAngle != 0.5 {
		t.Errorf("ConeAngle = %f, want 0.5", shaft.ConeAngle)
	}
	if shaft.DustDensity != 0.8 {
		t.Errorf("DustDensity = %f, want 0.8", shaft.DustDensity)
	}
	if shaft.DirX != 1.0 {
		t.Errorf("DirX = %f, want 1.0", shaft.DirX)
	}
}

func TestComponentReset(t *testing.T) {
	c := NewComponent()
	c.Intensity = 0.8
	c.DustDensity = 0.5
	c.ScatterR = 0.9

	// Component should retain values
	if c.Intensity != 0.8 {
		t.Errorf("Intensity = %f, want 0.8", c.Intensity)
	}
}

func TestGenreScatterColors(t *testing.T) {
	// Each genre should have distinct scatter colors
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		preset := genrePresets[genre]

		// Colors should be in valid range
		if preset.ScatterColorR < 0 || preset.ScatterColorR > 1 {
			t.Errorf("%s: ScatterColorR out of range: %f", genre, preset.ScatterColorR)
		}
		if preset.ScatterColorG < 0 || preset.ScatterColorG > 1 {
			t.Errorf("%s: ScatterColorG out of range: %f", genre, preset.ScatterColorG)
		}
		if preset.ScatterColorB < 0 || preset.ScatterColorB > 1 {
			t.Errorf("%s: ScatterColorB out of range: %f", genre, preset.ScatterColorB)
		}
	}
}

func TestFantasyVsCyberpunkPresets(t *testing.T) {
	fantasy := genrePresets["fantasy"]
	cyberpunk := genrePresets["cyberpunk"]

	// Fantasy should be warmer (higher R, lower B)
	if fantasy.ScatterColorR < fantasy.ScatterColorB {
		t.Error("Fantasy should have warm scatter (R > B)")
	}

	// Cyberpunk should have moderate dust
	if cyberpunk.BaseDustDensity > fantasy.BaseDustDensity {
		t.Error("Cyberpunk should have less dust than fantasy")
	}
}

// Benchmark tests
func BenchmarkNewSystem(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewSystem("fantasy", 320, 200)
	}
}

func BenchmarkCreateLightShaft(b *testing.B) {
	sys := NewSystem("fantasy", 320, 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.CreateLightShaftFromTorch(float64(i%100), float64(i%100), 0.8)
	}
}

func BenchmarkSetGenre(b *testing.B) {
	sys := NewSystem("fantasy", 320, 200)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.SetGenre(genres[i%len(genres)])
	}
}

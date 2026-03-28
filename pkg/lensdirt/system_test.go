package lensdirt

import (
	"image/color"
	"testing"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    int64
		screenW int
		screenH int
		wantNil bool
	}{
		{"fantasy genre", "fantasy", 12345, 320, 200, false},
		{"scifi genre", "scifi", 54321, 640, 480, false},
		{"horror genre", "horror", 11111, 320, 200, false},
		{"cyberpunk genre", "cyberpunk", 22222, 800, 600, false},
		{"postapoc genre", "postapoc", 33333, 320, 200, false},
		{"unknown genre defaults to fantasy", "unknown", 44444, 320, 200, false},
		{"empty genre defaults to fantasy", "", 55555, 320, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID, tt.seed, tt.screenW, tt.screenH)

			if (sys == nil) != tt.wantNil {
				t.Errorf("NewSystem() = %v, wantNil = %v", sys, tt.wantNil)
			}

			if sys != nil {
				if sys.screenW != tt.screenW {
					t.Errorf("screenW = %d, want %d", sys.screenW, tt.screenW)
				}
				if sys.screenH != tt.screenH {
					t.Errorf("screenH = %d, want %d", sys.screenH, tt.screenH)
				}
				if sys.pattern == nil {
					t.Error("pattern should not be nil after initialization")
				}
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	genres := []string{"scifi", "horror", "cyberpunk", "postapoc", "fantasy"}

	for _, genre := range genres {
		sys.SetGenre(genre)

		if sys.genreID != genre {
			t.Errorf("after SetGenre(%q), genreID = %q", genre, sys.genreID)
		}

		// Verify preset was updated
		preset := sys.GetPreset()
		expectedPreset, ok := genrePresets[genre]
		if !ok {
			expectedPreset = genrePresets["fantasy"]
		}

		if preset.DirtDensity != expectedPreset.DirtDensity {
			t.Errorf("preset.DirtDensity = %f, want %f", preset.DirtDensity, expectedPreset.DirtDensity)
		}
	}
}

func TestSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small screen", 160, 100},
		{"medium screen", 640, 480},
		{"large screen", 1920, 1080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.SetScreenSize(tt.width, tt.height)

			if sys.screenW != tt.width {
				t.Errorf("screenW = %d, want %d", sys.screenW, tt.width)
			}
			if sys.screenH != tt.height {
				t.Errorf("screenH = %d, want %d", sys.screenH, tt.height)
			}

			// Pattern should be regenerated
			if sys.pattern == nil {
				t.Error("pattern should not be nil after resize")
			}
		})
	}
}

func TestLightSourceManagement(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	// Test adding light sources
	light1 := NewLightSource(100, 100, 0.8, 50, color.RGBA{255, 200, 100, 255})
	light2 := NewLightSource(200, 150, 0.5, 30, color.RGBA{100, 150, 255, 255})

	sys.AddLightSource(light1)
	sys.AddLightSource(light2)

	sys.lightsMu.RLock()
	if len(sys.lights) != 2 {
		t.Errorf("expected 2 lights, got %d", len(sys.lights))
	}
	sys.lightsMu.RUnlock()

	// Test clearing lights
	sys.ClearLights()

	sys.lightsMu.RLock()
	if len(sys.lights) != 0 {
		t.Errorf("expected 0 lights after clear, got %d", len(sys.lights))
	}
	sys.lightsMu.RUnlock()

	// Test setting light sources
	lights := []LightSource{light1, light2}
	sys.SetLightSources(lights)

	sys.lightsMu.RLock()
	if len(sys.lights) != 2 {
		t.Errorf("expected 2 lights after SetLightSources, got %d", len(sys.lights))
	}
	sys.lightsMu.RUnlock()
}

func TestNewLightSource(t *testing.T) {
	tests := []struct {
		name          string
		screenX       float64
		screenY       float64
		intensity     float64
		radius        float64
		wantIntensity float64
	}{
		{"normal intensity", 100, 100, 0.5, 50, 0.5},
		{"clamped high intensity", 100, 100, 1.5, 50, 1.0},
		{"clamped low intensity", 100, 100, -0.5, 50, 0.0},
		{"zero intensity", 100, 100, 0, 50, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			light := NewLightSource(tt.screenX, tt.screenY, tt.intensity, tt.radius, color.RGBA{255, 255, 255, 255})

			if light.Intensity != tt.wantIntensity {
				t.Errorf("Intensity = %f, want %f", light.Intensity, tt.wantIntensity)
			}
			if light.ScreenX != tt.screenX {
				t.Errorf("ScreenX = %f, want %f", light.ScreenX, tt.screenX)
			}
			if light.ScreenY != tt.screenY {
				t.Errorf("ScreenY = %f, want %f", light.ScreenY, tt.screenY)
			}
		})
	}
}

func TestGeneratePattern(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	if sys.pattern == nil {
		t.Fatal("pattern should not be nil")
	}

	if len(sys.pattern.Specks) == 0 {
		t.Error("pattern should have specks")
	}

	// Verify specks are within bounds
	for i, speck := range sys.pattern.Specks {
		if speck.X < 0 || speck.X > 1 {
			t.Errorf("speck[%d].X = %f, out of [0,1] range", i, speck.X)
		}
		if speck.Y < 0 || speck.Y > 1 {
			t.Errorf("speck[%d].Y = %f, out of [0,1] range", i, speck.Y)
		}
		if speck.Size <= 0 {
			t.Errorf("speck[%d].Size = %f, should be positive", i, speck.Size)
		}
		if speck.BaseOpacity < 0 || speck.BaseOpacity > 1 {
			t.Errorf("speck[%d].BaseOpacity = %f, out of [0,1] range", i, speck.BaseOpacity)
		}
	}
}

func TestPatternDeterminism(t *testing.T) {
	seed := int64(42424242)

	// Generate two patterns with same seed
	sys1 := NewSystem("fantasy", seed, 320, 200)
	sys2 := NewSystem("fantasy", seed, 320, 200)

	if len(sys1.pattern.Specks) != len(sys2.pattern.Specks) {
		t.Fatalf("speck counts differ: %d vs %d", len(sys1.pattern.Specks), len(sys2.pattern.Specks))
	}

	for i := range sys1.pattern.Specks {
		s1 := sys1.pattern.Specks[i]
		s2 := sys2.pattern.Specks[i]

		if s1.X != s2.X || s1.Y != s2.Y {
			t.Errorf("speck[%d] positions differ: (%f,%f) vs (%f,%f)", i, s1.X, s1.Y, s2.X, s2.Y)
		}
		if s1.Size != s2.Size {
			t.Errorf("speck[%d] sizes differ: %f vs %f", i, s1.Size, s2.Size)
		}
		if s1.Shape != s2.Shape {
			t.Errorf("speck[%d] shapes differ: %d vs %d", i, s1.Shape, s2.Shape)
		}
	}
}

func TestGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset, ok := genrePresets[genre]
			if !ok {
				t.Fatalf("missing preset for genre %q", genre)
			}

			if preset.DirtDensity <= 0 || preset.DirtDensity > 1 {
				t.Errorf("DirtDensity = %f, should be in (0,1]", preset.DirtDensity)
			}
			if preset.SmudgeRatio < 0 || preset.SmudgeRatio > 1 {
				t.Errorf("SmudgeRatio = %f, should be in [0,1]", preset.SmudgeRatio)
			}
			if preset.StreakRatio < 0 || preset.StreakRatio > 1 {
				t.Errorf("StreakRatio = %f, should be in [0,1]", preset.StreakRatio)
			}
			if preset.IntensityScale <= 0 || preset.IntensityScale > 1 {
				t.Errorf("IntensityScale = %f, should be in (0,1]", preset.IntensityScale)
			}
			if preset.FalloffDistance <= 0 {
				t.Errorf("FalloffDistance = %f, should be positive", preset.FalloffDistance)
			}
		})
	}
}

func TestClamp01(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},
		{0.0, 0.0},
		{1.0, 1.0},
		{-0.5, 0.0},
		{1.5, 1.0},
		{-100, 0.0},
		{100, 1.0},
	}

	for _, tt := range tests {
		result := clamp01(tt.input)
		if result != tt.expected {
			t.Errorf("clamp01(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestComponentType(t *testing.T) {
	pattern := &DirtPattern{}
	if pattern.Type() != "DirtPattern" {
		t.Errorf("DirtPattern.Type() = %q, want %q", pattern.Type(), "DirtPattern")
	}

	light := &LightSource{}
	if light.Type() != "LightSource" {
		t.Errorf("LightSource.Type() = %q, want %q", light.Type(), "LightSource")
	}
}

func TestClearCache(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	// Create some cached sprites
	sys.getOrCreateSpeckSprite(ShapeCircle, 10, 255, 255, 255, 0)
	sys.getOrCreateSpeckSprite(ShapeSmudge, 15, 200, 180, 140, 0)

	sys.speckCacheMu.RLock()
	cacheSize := len(sys.speckCache)
	sys.speckCacheMu.RUnlock()

	if cacheSize == 0 {
		t.Fatal("cache should have entries before clear")
	}

	sys.ClearCache()

	sys.speckCacheMu.RLock()
	cacheSize = len(sys.speckCache)
	sys.speckCacheMu.RUnlock()

	if cacheSize != 0 {
		t.Errorf("cache should be empty after clear, got %d entries", cacheSize)
	}
}

func TestSpeckShapeConstants(t *testing.T) {
	// Verify shape constants are distinct
	shapes := []SpeckShape{ShapeCircle, ShapeSmudge, ShapeStreaky, ShapeDiffuse, ShapeHexagonal}
	seen := make(map[SpeckShape]bool)

	for _, shape := range shapes {
		if seen[shape] {
			t.Errorf("duplicate shape constant: %d", shape)
		}
		seen[shape] = true
	}
}

func TestCreateSpeckSpriteShapes(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	shapes := []SpeckShape{ShapeCircle, ShapeSmudge, ShapeStreaky, ShapeDiffuse, ShapeHexagonal}
	col := color.RGBA{R: 200, G: 180, B: 140, A: 255}

	for _, shape := range shapes {
		t.Run(shapeToString(shape), func(t *testing.T) {
			sprite := sys.createSpeckSprite(shape, 16, col)

			if sprite == nil {
				t.Error("sprite should not be nil")
				return
			}

			bounds := sprite.Bounds()
			if bounds.Dx() < 2 || bounds.Dy() < 2 {
				t.Errorf("sprite too small: %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func shapeToString(shape SpeckShape) string {
	switch shape {
	case ShapeCircle:
		return "circle"
	case ShapeSmudge:
		return "smudge"
	case ShapeStreaky:
		return "streaky"
	case ShapeDiffuse:
		return "diffuse"
	case ShapeHexagonal:
		return "hexagonal"
	default:
		return "unknown"
	}
}

func BenchmarkGeneratePattern(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sys := NewSystem("fantasy", int64(i), 320, 200)
		_ = sys.pattern
	}
}

func BenchmarkCreateSpeckSprite(b *testing.B) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	col := color.RGBA{R: 200, G: 180, B: 140, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.createSpeckSprite(ShapeCircle, 16, col)
	}
}

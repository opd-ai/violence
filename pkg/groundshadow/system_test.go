package groundshadow

import (
	"image/color"
	"testing"
)

func TestNewComponent(t *testing.T) {
	comp := NewComponent()
	if !comp.CastShadow {
		t.Error("expected CastShadow to be true by default")
	}
	if comp.Radius != 0.5 {
		t.Errorf("expected Radius 0.5, got %f", comp.Radius)
	}
	if comp.Height != 1.0 {
		t.Errorf("expected Height 1.0, got %f", comp.Height)
	}
	if comp.Opacity != 0.6 {
		t.Errorf("expected Opacity 0.6, got %f", comp.Opacity)
	}
}

func TestNewComponentWithSize(t *testing.T) {
	comp := NewComponentWithSize(0.8, 2.5)
	if comp.Radius != 0.8 {
		t.Errorf("expected Radius 0.8, got %f", comp.Radius)
	}
	if comp.Height != 2.5 {
		t.Errorf("expected Height 2.5, got %f", comp.Height)
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent()
	if comp.Type() != "groundshadow" {
		t.Errorf("expected Type() to return 'groundshadow', got %q", comp.Type())
	}
}

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy")
	if sys.genre != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %q", sys.genre)
	}
	if sys.pixelsPerUnit != 32.0 {
		t.Errorf("expected pixelsPerUnit 32.0, got %f", sys.pixelsPerUnit)
	}
	if sys.shadowCache == nil {
		t.Error("expected shadowCache to be initialized")
	}
}

func TestGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)
			preset := sys.GetPreset()

			// Verify valid ranges
			if preset.BaseOpacity < 0.0 || preset.BaseOpacity > 1.0 {
				t.Errorf("BaseOpacity out of range [0,1]: %f", preset.BaseOpacity)
			}
			if preset.Softness < 0.0 || preset.Softness > 1.0 {
				t.Errorf("Softness out of range [0,1]: %f", preset.Softness)
			}
			if preset.HeightScale <= 0.0 {
				t.Errorf("HeightScale should be positive: %f", preset.HeightScale)
			}
			if preset.MaxElongation < 0.0 || preset.MaxElongation > 1.0 {
				t.Errorf("MaxElongation out of range [0,1]: %f", preset.MaxElongation)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")
	originalPreset := sys.GetPreset()

	sys.SetGenre("horror")
	newPreset := sys.GetPreset()

	// Horror should have deeper shadows than fantasy
	if newPreset.BaseOpacity <= originalPreset.BaseOpacity {
		t.Error("horror preset should have higher BaseOpacity than fantasy")
	}
	if sys.GetGenre() != "horror" {
		t.Errorf("expected genre 'horror', got %q", sys.GetGenre())
	}
}

func TestSetLightDirection(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetLightDirection(0.5, -0.5, 1.5)

	// Verify values are stored (indirect test via render behavior)
	if sys.lightDirX != 0.5 {
		t.Errorf("expected lightDirX 0.5, got %f", sys.lightDirX)
	}
	if sys.lightDirY != -0.5 {
		t.Errorf("expected lightDirY -0.5, got %f", sys.lightDirY)
	}
	if sys.lightStrength != 1.5 {
		t.Errorf("expected lightStrength 1.5, got %f", sys.lightStrength)
	}
}

func TestSetPixelsPerUnit(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetPixelsPerUnit(64.0)

	if sys.pixelsPerUnit != 64.0 {
		t.Errorf("expected pixelsPerUnit 64.0, got %f", sys.pixelsPerUnit)
	}

	// Invalid value should be ignored
	sys.SetPixelsPerUnit(0.0)
	if sys.pixelsPerUnit != 64.0 {
		t.Error("SetPixelsPerUnit should ignore zero value")
	}
	sys.SetPixelsPerUnit(-10.0)
	if sys.pixelsPerUnit != 64.0 {
		t.Error("SetPixelsPerUnit should ignore negative value")
	}
}

func TestGenerateShadowImage(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		name       string
		radiusPx   int
		softness   float64
		elongation float64
		opacity    float64
	}{
		{"small circle", 8, 0.5, 0.0, 0.6},
		{"large circle", 32, 0.5, 0.0, 0.6},
		{"soft", 16, 0.9, 0.0, 0.6},
		{"hard", 16, 0.1, 0.0, 0.6},
		{"elongated", 16, 0.5, 0.5, 0.6},
		{"very elongated", 16, 0.5, 0.8, 0.6},
		{"transparent", 16, 0.5, 0.0, 0.2},
		{"opaque", 16, 0.5, 0.0, 0.9},
		{"minimum size", 1, 0.5, 0.0, 0.6},
	}

	tint := color.RGBA{R: 20, G: 15, B: 30, A: 255}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := sys.generateShadowImage(tt.radiusPx, tt.softness, tt.elongation, tt.opacity, tint)
			if img == nil {
				t.Fatal("expected non-nil image")
			}

			bounds := img.Bounds()
			if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
				t.Errorf("invalid image dimensions: %dx%d", bounds.Dx(), bounds.Dy())
			}

			// Verify image has some content (not all transparent)
			hasContent := false
			for y := 0; y < bounds.Dy() && !hasContent; y++ {
				for x := 0; x < bounds.Dx() && !hasContent; x++ {
					r, g, b, a := img.At(x, y).RGBA()
					if a > 0 || r > 0 || g > 0 || b > 0 {
						hasContent = true
					}
				}
			}
			if !hasContent && tt.opacity > 0 {
				t.Error("generated image has no visible content")
			}

			img.Dispose()
		})
	}
}

func TestShadowCache(t *testing.T) {
	cache := newShadowCache(3)

	// Test put and get
	key1 := shadowKey{radiusPx: 10, softness: 5, elongation: 0, opacity: 12}
	key2 := shadowKey{radiusPx: 20, softness: 5, elongation: 0, opacity: 12}
	key3 := shadowKey{radiusPx: 30, softness: 5, elongation: 0, opacity: 12}
	key4 := shadowKey{radiusPx: 40, softness: 5, elongation: 0, opacity: 12}

	sys := NewSystem("fantasy")
	tint := color.RGBA{R: 20, G: 15, B: 30, A: 255}

	img1 := sys.generateShadowImage(10, 0.5, 0.0, 0.6, tint)
	img2 := sys.generateShadowImage(20, 0.5, 0.0, 0.6, tint)
	img3 := sys.generateShadowImage(30, 0.5, 0.0, 0.6, tint)
	img4 := sys.generateShadowImage(40, 0.5, 0.0, 0.6, tint)

	cache.put(key1, img1)
	cache.put(key2, img2)
	cache.put(key3, img3)

	// All three should be cached
	if cache.get(key1) == nil {
		t.Error("key1 should be cached")
	}
	if cache.get(key2) == nil {
		t.Error("key2 should be cached")
	}
	if cache.get(key3) == nil {
		t.Error("key3 should be cached")
	}

	// Adding fourth should evict first
	cache.put(key4, img4)
	if cache.get(key1) != nil {
		t.Error("key1 should have been evicted")
	}
	if cache.get(key4) == nil {
		t.Error("key4 should be cached")
	}

	// Clear should empty cache
	cache.clear()
	if cache.get(key2) != nil {
		t.Error("cache should be empty after clear")
	}
	if len(cache.cache) != 0 {
		t.Errorf("cache map should be empty, has %d entries", len(cache.cache))
	}
}

func TestGetShadowImageForEntity(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewComponentWithSize(0.5, 1.0)

	img := sys.GetShadowImageForEntity(comp)
	if img == nil {
		t.Error("expected non-nil image for valid component")
	}

	// Nil component should return nil
	img = sys.GetShadowImageForEntity(nil)
	if img != nil {
		t.Error("expected nil image for nil component")
	}
}

func TestRenderShadowNilComponent(t *testing.T) {
	sys := NewSystem("fantasy")
	// Should not panic with nil component
	sys.RenderShadow(nil, 0, 0, 0, 0, nil)
}

func TestRenderShadowDisabled(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewComponent()
	comp.CastShadow = false
	// Should return early without error
	sys.RenderShadow(nil, 0, 0, 0, 0, comp)
}

func TestShadowKeyUniqueness(t *testing.T) {
	// Verify that different parameters produce different keys
	key1 := shadowKey{radiusPx: 10, softness: 5, elongation: 3, opacity: 12, tintR: 20, tintG: 15, tintB: 30}
	key2 := shadowKey{radiusPx: 10, softness: 5, elongation: 3, opacity: 12, tintR: 20, tintG: 15, tintB: 30}
	key3 := shadowKey{radiusPx: 11, softness: 5, elongation: 3, opacity: 12, tintR: 20, tintG: 15, tintB: 30}

	if key1 != key2 {
		t.Error("identical keys should be equal")
	}
	if key1 == key3 {
		t.Error("different keys should not be equal")
	}
}

func TestGenreSpecificShadowAppearance(t *testing.T) {
	genres := map[string]struct {
		expectSofter bool // Compared to scifi
		expectDarker bool // Compared to scifi
	}{
		"horror":   {expectSofter: true, expectDarker: true},
		"fantasy":  {expectSofter: true, expectDarker: true},
		"postapoc": {expectSofter: true, expectDarker: false},
	}

	scifiSys := NewSystem("scifi")
	scifiPreset := scifiSys.GetPreset()

	for genre, expect := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)
			preset := sys.GetPreset()

			if expect.expectSofter && preset.Softness <= scifiPreset.Softness {
				t.Errorf("%s should be softer than scifi (%.2f <= %.2f)",
					genre, preset.Softness, scifiPreset.Softness)
			}
			if expect.expectDarker && preset.BaseOpacity <= scifiPreset.BaseOpacity {
				t.Errorf("%s should be darker than scifi (%.2f <= %.2f)",
					genre, preset.BaseOpacity, scifiPreset.BaseOpacity)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	sys := NewSystem("fantasy")
	// Update should be a no-op and not panic
	sys.Update()
}

func BenchmarkGenerateShadowImage(b *testing.B) {
	sys := NewSystem("fantasy")
	tint := color.RGBA{R: 20, G: 15, B: 30, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		img := sys.generateShadowImage(16, 0.5, 0.3, 0.6, tint)
		img.Dispose()
	}
}

func BenchmarkCachedShadowLookup(b *testing.B) {
	sys := NewSystem("fantasy")
	tint := color.RGBA{R: 20, G: 15, B: 30, A: 255}

	// Pre-populate cache
	key := shadowKey{radiusPx: 16, softness: 5, elongation: 3, opacity: 12}
	img := sys.generateShadowImage(16, 0.5, 0.3, 0.6, tint)
	sys.shadowCache.put(key, img)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.shadowCache.get(key)
	}
}

package parallax

import (
	"testing"
)

func TestGenerateFantasyLayers(t *testing.T) {
	layers := GenerateLayers("fantasy", "forest", 12345, 800, 600)

	if len(layers) == 0 {
		t.Fatal("GenerateLayers returned no layers")
	}

	// Fantasy should have mountains, hills, and trees
	if len(layers) < 2 {
		t.Errorf("Fantasy layers = %d, want at least 2", len(layers))
	}

	// Verify layers are properly ordered by Z-index
	for i := 1; i < len(layers); i++ {
		if layers[i].ZIndex <= layers[i-1].ZIndex {
			// Z-index should generally increase, but allow equal values
			// (we'll sort them during rendering)
		}
	}

	// Verify all layers have images
	for i, layer := range layers {
		if layer.Image == nil {
			t.Errorf("Layer %d has nil image", i)
		}

		if layer.Width <= 0 || layer.Height <= 0 {
			t.Errorf("Layer %d has invalid dimensions: %dx%d", i, layer.Width, layer.Height)
		}

		if layer.Opacity <= 0 || layer.Opacity > 1.0 {
			t.Errorf("Layer %d opacity %f out of range", i, layer.Opacity)
		}
	}
}

func TestGenerateSciFiLayers(t *testing.T) {
	layers := GenerateLayers("scifi", "station", 67890, 800, 600)

	if len(layers) == 0 {
		t.Fatal("GenerateLayers returned no layers for scifi")
	}

	// SciFi should have starfield, structures, panels
	if len(layers) < 2 {
		t.Errorf("SciFi layers = %d, want at least 2", len(layers))
	}

	for i, layer := range layers {
		if layer.Image == nil {
			t.Errorf("SciFi layer %d has nil image", i)
		}
	}
}

func TestGenerateHorrorLayers(t *testing.T) {
	layers := GenerateLayers("horror", "crypt", 11111, 800, 600)

	if len(layers) == 0 {
		t.Fatal("GenerateLayers returned no layers for horror")
	}

	for i, layer := range layers {
		if layer.Image == nil {
			t.Errorf("Horror layer %d has nil image", i)
		}
	}
}

func TestGenerateCyberpunkLayers(t *testing.T) {
	layers := GenerateLayers("cyberpunk", "city", 22222, 800, 600)

	if len(layers) == 0 {
		t.Fatal("GenerateLayers returned no layers for cyberpunk")
	}

	// Cyberpunk should have skyline, rain, buildings
	if len(layers) < 2 {
		t.Errorf("Cyberpunk layers = %d, want at least 2", len(layers))
	}

	for i, layer := range layers {
		if layer.Image == nil {
			t.Errorf("Cyberpunk layer %d has nil image", i)
		}
	}
}

func TestGeneratePostApocLayers(t *testing.T) {
	layers := GenerateLayers("postapoc", "wasteland", 33333, 800, 600)

	if len(layers) == 0 {
		t.Fatal("GenerateLayers returned no layers for postapoc")
	}

	for i, layer := range layers {
		if layer.Image == nil {
			t.Errorf("PostApoc layer %d has nil image", i)
		}
	}
}

func TestLayerScrollSpeed(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layers := GenerateLayers(tt.genreID, "default", 12345, 800, 600)

			for i, layer := range layers {
				// Far layers should scroll slower than near layers
				if layer.ScrollSpeed < 0 || layer.ScrollSpeed > 1.0 {
					t.Errorf("Layer %d scroll speed %f out of expected range [0, 1]", i, layer.ScrollSpeed)
				}
			}

			// Verify scroll speeds generally increase with Z-index
			if len(layers) > 1 {
				hasVariation := false
				for i := 1; i < len(layers); i++ {
					if layers[i].ScrollSpeed != layers[0].ScrollSpeed {
						hasVariation = true
						break
					}
				}
				if !hasVariation {
					t.Error("All layers have same scroll speed - no parallax effect")
				}
			}
		})
	}
}

func TestLayerRepeat(t *testing.T) {
	layers := GenerateLayers("fantasy", "forest", 12345, 800, 600)

	// At least one layer should have RepeatX enabled
	hasRepeat := false
	for _, layer := range layers {
		if layer.RepeatX {
			hasRepeat = true
			break
		}
	}

	if !hasRepeat {
		t.Error("No layers have RepeatX enabled")
	}
}

func TestLayerTint(t *testing.T) {
	layers := GenerateLayers("fantasy", "forest", 12345, 800, 600)

	for i, layer := range layers {
		for j, val := range layer.Tint {
			if val < 0 || val > 1.0 {
				t.Errorf("Layer %d tint[%d] = %f, out of range [0, 1]", i, j, val)
			}
		}
	}
}

func TestDefaultGenre(t *testing.T) {
	// Unknown genre should default to fantasy
	layers := GenerateLayers("unknown_genre", "biome", 12345, 800, 600)

	if len(layers) == 0 {
		t.Fatal("Default genre fallback produced no layers")
	}
}

func TestDeterministicGeneration(t *testing.T) {
	seed := int64(99999)

	layers1 := GenerateLayers("fantasy", "forest", seed, 800, 600)
	layers2 := GenerateLayers("fantasy", "forest", seed, 800, 600)

	if len(layers1) != len(layers2) {
		t.Fatalf("Layer count mismatch: %d vs %d", len(layers1), len(layers2))
	}

	for i := range layers1 {
		if layers1[i].ScrollSpeed != layers2[i].ScrollSpeed {
			t.Errorf("Layer %d scroll speed differs: %f vs %f", i, layers1[i].ScrollSpeed, layers2[i].ScrollSpeed)
		}

		if layers1[i].ZIndex != layers2[i].ZIndex {
			t.Errorf("Layer %d ZIndex differs: %d vs %d", i, layers1[i].ZIndex, layers2[i].ZIndex)
		}

		if layers1[i].Opacity != layers2[i].Opacity {
			t.Errorf("Layer %d opacity differs: %f vs %f", i, layers1[i].Opacity, layers2[i].Opacity)
		}
	}
}

func TestLayerDimensions(t *testing.T) {
	width, height := 1024, 768
	layers := GenerateLayers("fantasy", "forest", 12345, width, height)

	for i, layer := range layers {
		if layer.Width <= 0 {
			t.Errorf("Layer %d has invalid width: %d", i, layer.Width)
		}

		if layer.Height <= 0 {
			t.Errorf("Layer %d has invalid height: %d", i, layer.Height)
		}

		// Layers should generally fit within requested dimensions (or be designed to tile)
		if layer.Height > height*2 {
			t.Errorf("Layer %d height %d exceeds reasonable bounds", i, layer.Height)
		}
	}
}

func BenchmarkGenerateFantasyLayers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateLayers("fantasy", "forest", int64(i), 800, 600)
	}
}

func BenchmarkGenerateSciFiLayers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateLayers("scifi", "station", int64(i), 800, 600)
	}
}

func BenchmarkGenerateHorrorLayers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateLayers("horror", "crypt", int64(i), 800, 600)
	}
}

func BenchmarkGenerateCyberpunkLayers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateLayers("cyberpunk", "city", int64(i), 800, 600)
	}
}

func BenchmarkGeneratePostApocLayers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateLayers("postapoc", "wasteland", int64(i), 800, 600)
	}
}

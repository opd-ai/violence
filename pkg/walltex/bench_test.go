package walltex

import (
	"image"
	"image/png"
	"os"
	"testing"
)

// TestVisualOutput generates sample texture images for visual inspection.
// This test is skipped in normal runs but useful for development.
func TestVisualOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual output test in short mode")
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		g := NewGenerator(genre)

		// Generate both primary and secondary material variants
		for variant := 0; variant < 2; variant++ {
			img := g.Generate(128, variant, 42)

			// Skip file writing in automated tests
			// Uncomment to generate actual PNG files for inspection:
			// filename := fmt.Sprintf("sample_%s_v%d.png", genre, variant)
			// saveImage(img, filename)

			if img == nil {
				t.Errorf("Failed to generate texture for %s variant %d", genre, variant)
			}
		}
	}
}

// Helper function to save images (not used in automated tests)
func saveImage(img image.Image, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// BenchmarkMaterialRendering benchmarks each material type separately.
func BenchmarkMaterialRendering(b *testing.B) {
	materials := []struct {
		name  string
		genre string
	}{
		{"stone_fantasy", "fantasy"},
		{"metal_scifi", "scifi"},
		{"wood_horror", "horror"},
		{"concrete_cyberpunk", "cyberpunk"},
		{"tech_scifi", "scifi"},
		{"organic_horror", "horror"},
		{"crystal_fantasy", "fantasy"},
	}

	for _, m := range materials {
		b.Run(m.name, func(b *testing.B) {
			g := NewGenerator(m.genre)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.Generate(64, i%2, uint64(i))
			}
		})
	}
}

// BenchmarkWithWeathering measures performance impact of weathering.
func BenchmarkWithWeathering(b *testing.B) {
	tests := []struct {
		name  string
		genre string
	}{
		{"low_weather_scifi", "scifi"},
		{"high_weather_postapoc", "postapoc"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			g := NewGenerator(tt.genre)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.Generate(64, 0, uint64(i))
			}
		})
	}
}

// BenchmarkSizes measures generation time across different texture sizes.
func BenchmarkSizes(b *testing.B) {
	sizes := []int{32, 64, 128, 256}
	g := NewGenerator("fantasy")

	for _, size := range sizes {
		b.Run(string(rune('0'+size/32)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				g.Generate(size, 0, uint64(i))
			}
		})
	}
}

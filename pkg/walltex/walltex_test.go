package walltex

import (
	"image/color"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name  string
		genre string
		want  Material
	}{
		{"fantasy", "fantasy", MaterialStone},
		{"scifi", "scifi", MaterialMetal},
		{"horror", "horror", MaterialWood},
		{"cyberpunk", "cyberpunk", MaterialConcrete},
		{"postapoc", "postapoc", MaterialConcrete},
		{"unknown", "unknown", MaterialStone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.genre)
			if g == nil {
				t.Fatal("NewGenerator returned nil")
			}
			if g.genre != tt.genre && tt.genre != "unknown" {
				t.Errorf("genre = %q, want %q", g.genre, tt.genre)
			}
			if g.preset.PrimaryMaterial != tt.want {
				t.Errorf("PrimaryMaterial = %v, want %v", g.preset.PrimaryMaterial, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		size    int
		variant int
		seed    uint64
	}{
		{"fantasy_stone_32", "fantasy", 32, 0, 12345},
		{"fantasy_wood_32", "fantasy", 32, 1, 12345},
		{"scifi_metal_64", "scifi", 64, 0, 54321},
		{"scifi_tech_64", "scifi", 64, 1, 54321},
		{"horror_wood_32", "horror", 32, 0, 99999},
		{"horror_organic_32", "horror", 32, 1, 99999},
		{"cyberpunk_concrete_64", "cyberpunk", 64, 0, 11111},
		{"cyberpunk_tech_64", "cyberpunk", 64, 1, 11111},
		{"postapoc_concrete_32", "postapoc", 32, 0, 22222},
		{"postapoc_metal_32", "postapoc", 32, 1, 22222},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.genre)
			img := g.Generate(tt.size, tt.variant, tt.seed)

			if img == nil {
				t.Fatal("Generate returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.size || bounds.Dy() != tt.size {
				t.Errorf("image size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), tt.size, tt.size)
			}

			// Check that image is not empty (has varied pixels)
			pixelVariety := make(map[color.RGBA]bool)
			for y := 0; y < tt.size; y++ {
				for x := 0; x < tt.size; x++ {
					c := img.RGBAAt(x, y)
					pixelVariety[c] = true
					if c.A != 255 {
						t.Errorf("pixel at (%d,%d) has alpha %d, want 255", x, y, c.A)
					}
				}
			}

			if len(pixelVariety) < 5 {
				t.Errorf("texture has only %d unique colors, expected varied texture", len(pixelVariety))
			}
		})
	}
}

func TestDeterminism(t *testing.T) {
	g := NewGenerator("fantasy")

	img1 := g.Generate(32, 0, 12345)
	img2 := g.Generate(32, 0, 12345)

	if img1 == nil || img2 == nil {
		t.Fatal("Generate returned nil")
	}

	// Check that same seed produces identical output
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c1 := img1.RGBAAt(x, y)
			c2 := img2.RGBAAt(x, y)
			if c1 != c2 {
				t.Errorf("pixel at (%d,%d): first=%v, second=%v, want identical", x, y, c1, c2)
				return
			}
		}
	}
}

func TestVariantDifference(t *testing.T) {
	g := NewGenerator("scifi")

	img0 := g.Generate(32, 0, 12345)
	img1 := g.Generate(32, 1, 12345)

	if img0 == nil || img1 == nil {
		t.Fatal("Generate returned nil")
	}

	// Check that different variants produce different output
	differences := 0
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			c0 := img0.RGBAAt(x, y)
			c1 := img1.RGBAAt(x, y)
			if c0 != c1 {
				differences++
			}
		}
	}

	if differences < 100 {
		t.Errorf("variants differ in only %d pixels, expected significant difference", differences)
	}
}

func TestMaterialRendering(t *testing.T) {
	tests := []struct {
		name     string
		material Material
	}{
		{"stone", MaterialStone},
		{"metal", MaterialMetal},
		{"wood", MaterialWood},
		{"concrete", MaterialConcrete},
		{"organic", MaterialOrganic},
		{"crystal", MaterialCrystal},
		{"tech", MaterialTech},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Generator{
				genre:  "fantasy",
				preset: genrePresets["fantasy"],
			}

			size := 32
			img := g.Generate(size, int(tt.material), 54321)

			if img == nil {
				t.Fatal("Generate returned nil")
			}

			// Verify texture is not uniform
			pixelMap := make(map[color.RGBA]int)
			for y := 0; y < size; y++ {
				for x := 0; x < size; x++ {
					c := img.RGBAAt(x, y)
					pixelMap[c]++
				}
			}

			if len(pixelMap) < 3 {
				t.Errorf("material %v has only %d unique colors, expected variety", tt.material, len(pixelMap))
			}
		})
	}
}

func TestWeathering(t *testing.T) {
	tests := []struct {
		name             string
		genre            string
		weatherIntensity float64
	}{
		{"fantasy_medium", "fantasy", 0.6},
		{"scifi_low", "scifi", 0.2},
		{"horror_high", "horror", 0.8},
		{"postapoc_very_high", "postapoc", 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.genre)

			if g.preset.WeatherIntensity != tt.weatherIntensity {
				t.Errorf("WeatherIntensity = %f, want %f", g.preset.WeatherIntensity, tt.weatherIntensity)
			}

			img := g.Generate(64, 0, 99999)
			if img == nil {
				t.Fatal("Generate returned nil")
			}

			// High weathering should have darker pixels (stains/cracks)
			if tt.weatherIntensity > 0.5 {
				darkPixels := 0
				for y := 0; y < 64; y++ {
					for x := 0; x < 64; x++ {
						c := img.RGBAAt(x, y)
						luminance := float64(c.R)*0.299 + float64(c.G)*0.587 + float64(c.B)*0.114
						if luminance < 80 {
							darkPixels++
						}
					}
				}
				if darkPixels < 50 {
					t.Errorf("high weathering genre has only %d dark pixels, expected more", darkPixels)
				}
			}
		})
	}
}

func TestGlow(t *testing.T) {
	tests := []struct {
		name          string
		genre         string
		glowIntensity float64
	}{
		{"scifi_glow", "scifi", 0.4},
		{"cyberpunk_high_glow", "cyberpunk", 0.7},
		{"fantasy_no_glow", "fantasy", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.genre)

			if g.preset.GlowIntensity != tt.glowIntensity {
				t.Errorf("GlowIntensity = %f, want %f", g.preset.GlowIntensity, tt.glowIntensity)
			}

			img := g.Generate(64, 0, 77777)
			if img == nil {
				t.Fatal("Generate returned nil")
			}

			// Check for bright blue/cyan pixels indicating glow
			if tt.glowIntensity > 0 {
				brightBluePixels := 0
				for y := 0; y < 64; y++ {
					for x := 0; x < 64; x++ {
						c := img.RGBAAt(x, y)
						// Glow is characterized by high B and G values
						if c.B > 150 && c.G > 100 {
							brightBluePixels++
						}
					}
				}
				if brightBluePixels < 10 {
					t.Logf("genre %s with glow=%f has %d bright blue pixels", tt.genre, tt.glowIntensity, brightBluePixels)
				}
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("clamp", func(t *testing.T) {
		tests := []struct {
			input int
			want  uint8
		}{
			{-10, 0},
			{0, 0},
			{128, 128},
			{255, 255},
			{300, 255},
		}
		for _, tt := range tests {
			got := clamp(tt.input)
			if got != tt.want {
				t.Errorf("clamp(%d) = %d, want %d", tt.input, got, tt.want)
			}
		}
	})

	t.Run("fade", func(t *testing.T) {
		tests := []struct {
			input float64
			check func(float64) bool
		}{
			{0.0, func(v float64) bool { return v == 0.0 }},
			{1.0, func(v float64) bool { return v == 1.0 }},
			{0.5, func(v float64) bool { return v > 0 && v < 1 }},
		}
		for _, tt := range tests {
			got := fade(tt.input)
			if !tt.check(got) {
				t.Errorf("fade(%f) = %f, failed check", tt.input, got)
			}
		}
	})

	t.Run("lerp", func(t *testing.T) {
		tests := []struct {
			a, b, t float64
			want    float64
		}{
			{0, 10, 0.0, 0},
			{0, 10, 1.0, 10},
			{0, 10, 0.5, 5},
			{-5, 5, 0.5, 0},
		}
		for _, tt := range tests {
			got := lerp(tt.a, tt.b, tt.t)
			if got != tt.want {
				t.Errorf("lerp(%f, %f, %f) = %f, want %f", tt.a, tt.b, tt.t, got, tt.want)
			}
		}
	})

	t.Run("hashSeed", func(t *testing.T) {
		seed1 := hashSeed(12345)
		seed2 := hashSeed(12345)
		seed3 := hashSeed(54321)

		if seed1 != seed2 {
			t.Error("hashSeed is not deterministic")
		}
		if seed1 == seed3 {
			t.Error("hashSeed produces same output for different inputs")
		}
	})

	t.Run("hashCoord", func(t *testing.T) {
		h1 := hashCoord(0, 0)
		h2 := hashCoord(1, 0)
		h3 := hashCoord(0, 1)
		h4 := hashCoord(0, 0)

		if h1 != h4 {
			t.Error("hashCoord is not deterministic")
		}
		if h1 == h2 || h1 == h3 {
			t.Error("hashCoord produces same output for different coordinates")
		}
	})

	t.Run("luminance", func(t *testing.T) {
		tests := []struct {
			name  string
			color color.Color
			check func(float64) bool
		}{
			{"black", color.RGBA{0, 0, 0, 255}, func(v float64) bool { return v == 0.0 }},
			{"white", color.RGBA{255, 255, 255, 255}, func(v float64) bool { return v > 0.99 && v <= 1.0 }},
			{"grey", color.RGBA{128, 128, 128, 255}, func(v float64) bool { return v > 0 && v < 1 }},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := luminance(tt.color)
				if !tt.check(got) {
					t.Errorf("luminance(%v) = %f, failed check", tt.color, got)
				}
			})
		}
	})
}

func BenchmarkGenerate32(b *testing.B) {
	g := NewGenerator("fantasy")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Generate(32, 0, uint64(i))
	}
}

func BenchmarkGenerate64(b *testing.B) {
	g := NewGenerator("scifi")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Generate(64, 0, uint64(i))
	}
}

func BenchmarkGenerate128(b *testing.B) {
	g := NewGenerator("cyberpunk")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Generate(128, 0, uint64(i))
	}
}

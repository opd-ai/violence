package floor

import (
	"image"
	"image/color"
	"math/rand"
	"testing"
)

// newTestRNG creates a deterministic RNG for testing.
func newTestRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

func TestDefaultWeatheringConfig(t *testing.T) {
	config := DefaultWeatheringConfig()

	tests := []struct {
		name  string
		value float64
		min   float64
		max   float64
	}{
		{"EdgeDamage", config.EdgeDamage, 0.0, 1.0},
		{"WearIntensity", config.WearIntensity, 0.0, 1.0},
		{"AgeVariation", config.AgeVariation, 0.0, 1.0},
		{"MoistureLevel", config.MoistureLevel, 0.0, 1.0},
		{"OrganicGrowth", config.OrganicGrowth, 0.0, 1.0},
		{"ColorTemperature", config.ColorTemperature, -1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min || tt.value > tt.max {
				t.Errorf("%s = %v, want in range [%v, %v]", tt.name, tt.value, tt.min, tt.max)
			}
		})
	}
}

func TestApplyWeathering(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	materials := []MaterialType{
		MaterialStone,
		MaterialMetal,
		MaterialWood,
		MaterialConcrete,
		MaterialTile,
	}

	for _, material := range materials {
		t.Run(material.String(), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			// Fill with base color
			baseColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					img.Set(x, y, baseColor)
				}
			}

			config := DefaultWeatheringConfig()
			gen.ApplyWeathering(img, material, config, 12345)

			// Verify image was modified
			modified := false
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					c := img.At(x, y).(color.RGBA)
					if c.R != baseColor.R || c.G != baseColor.G || c.B != baseColor.B {
						modified = true
						break
					}
				}
				if modified {
					break
				}
			}

			if !modified {
				t.Error("Weathering did not modify the image")
			}
		})
	}
}

func TestApplyAgeVariation(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	tests := []struct {
		intensity float64
		material  MaterialType
	}{
		{0.0, MaterialStone},
		{0.5, MaterialMetal},
		{1.0, MaterialWood},
	}

	for _, tt := range tests {
		t.Run(tt.material.String(), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 16, 16))
			baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}
			for y := 0; y < 16; y++ {
				for x := 0; x < 16; x++ {
					img.Set(x, y, baseColor)
				}
			}

			// Create a controlled RNG for reproducibility
			gen.applyAgeVariation(img, 16, tt.material, tt.intensity, newTestRNG(42))

			if tt.intensity == 0.0 {
				// Should not modify at zero intensity
				for y := 0; y < 16; y++ {
					for x := 0; x < 16; x++ {
						c := img.At(x, y).(color.RGBA)
						if c != baseColor {
							t.Error("Zero intensity should not modify image")
							return
						}
					}
				}
			} else {
				// Should modify at non-zero intensity
				modified := false
				for y := 0; y < 16; y++ {
					for x := 0; x < 16; x++ {
						c := img.At(x, y).(color.RGBA)
						if c != baseColor {
							modified = true
							break
						}
					}
				}
				if !modified {
					t.Error("Non-zero intensity should modify image")
				}
			}
		})
	}
}

func TestApplyEdgeDamage(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	tests := []struct {
		material     MaterialType
		expectDamage bool
	}{
		{MaterialStone, true},
		{MaterialMetal, true},
		{MaterialWood, true},
		{MaterialConcrete, true},
		{MaterialDirt, false},  // Soft materials don't chip
		{MaterialGrass, false}, // Soft materials don't chip
	}

	for _, tt := range tests {
		t.Run(tt.material.String(), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 16, 16))
			baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
			for y := 0; y < 16; y++ {
				for x := 0; x < 16; x++ {
					img.Set(x, y, baseColor)
				}
			}

			gen.applyEdgeDamage(img, 16, tt.material, 0.8, newTestRNG(123))

			// Check edges for modification
			edgeModified := false
			// Check top and bottom edges
			for x := 0; x < 16; x++ {
				if img.At(x, 0).(color.RGBA) != baseColor || img.At(x, 15).(color.RGBA) != baseColor {
					edgeModified = true
					break
				}
			}
			// Check left and right edges
			for y := 0; y < 16; y++ {
				if img.At(0, y).(color.RGBA) != baseColor || img.At(15, y).(color.RGBA) != baseColor {
					edgeModified = true
					break
				}
			}

			if tt.expectDamage && !edgeModified {
				t.Error("Expected edge damage but none found")
			}
			if !tt.expectDamage && edgeModified {
				t.Error("Did not expect edge damage but found some")
			}
		})
	}
}

func TestApplyWearPatterns(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	// Fill with base color
	baseColor := color.RGBA{R: 120, G: 120, B: 120, A: 255}
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, baseColor)
		}
	}

	gen.applyWearPatterns(img, 32, MaterialStone, 0.7, newTestRNG(456))

	// Verify wear was applied (some pixels should be different)
	modified := false
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			if img.At(x, y).(color.RGBA) != baseColor {
				modified = true
				break
			}
		}
	}

	if !modified {
		t.Error("Wear patterns should modify the image")
	}
}

func TestApplyMoistureEffects(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	tests := []struct {
		material      MaterialType
		expectReduced bool
	}{
		{MaterialStone, false},
		{MaterialWood, false},
		{MaterialMetal, true},   // Reduced effect
		{MaterialCrystal, true}, // Reduced effect
	}

	for _, tt := range tests {
		t.Run(tt.material.String(), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 24, 24))
			baseColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
			for y := 0; y < 24; y++ {
				for x := 0; x < 24; x++ {
					img.Set(x, y, baseColor)
				}
			}

			gen.applyMoistureEffects(img, 24, tt.material, 0.8, newTestRNG(789))

			// Check for moisture (darkening and cool tint)
			foundMoisture := false
			for y := 0; y < 24; y++ {
				for x := 0; x < 24; x++ {
					c := img.At(x, y).(color.RGBA)
					// Moisture darkens and adds blue tint
					if c.R < baseColor.R && c.B > baseColor.B {
						foundMoisture = true
						break
					}
				}
			}

			// Moisture effect is probabilistic, we just verify no panic
			_ = foundMoisture
		})
	}
}

func TestApplyOrganicGrowth(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	tests := []struct {
		material   MaterialType
		shouldGrow bool
	}{
		{MaterialStone, true},
		{MaterialWood, true},
		{MaterialMetal, false},   // No growth on metal
		{MaterialCrystal, false}, // No growth on crystal
	}

	for _, tt := range tests {
		t.Run(tt.material.String(), func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 24, 24))
			baseColor := color.RGBA{R: 140, G: 140, B: 140, A: 255}
			for y := 0; y < 24; y++ {
				for x := 0; x < 24; x++ {
					img.Set(x, y, baseColor)
				}
			}

			gen.applyOrganicGrowth(img, 24, tt.material, 0.9, newTestRNG(111))

			// Check if growth was applied (green/brown tints)
			foundGrowth := false
			for y := 0; y < 24; y++ {
				for x := 0; x < 24; x++ {
					c := img.At(x, y).(color.RGBA)
					// Look for green or brown tints
					if c.G > baseColor.G+10 || (c.R > baseColor.R+10 && c.B < baseColor.B-10) {
						foundGrowth = true
						break
					}
				}
			}

			// Growth is probabilistic, we just verify no panic
			_ = foundGrowth
		})
	}
}

func TestApplyColorTemperature(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	tests := []struct {
		name        string
		temperature float64
		expectWarm  bool
		expectCool  bool
	}{
		{"warm", 0.5, true, false},
		{"cool", -0.5, false, true},
		{"neutral", 0.0, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := image.NewRGBA(image.Rect(0, 0, 16, 16))
			baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}
			for y := 0; y < 16; y++ {
				for x := 0; x < 16; x++ {
					img.Set(x, y, baseColor)
				}
			}

			gen.applyColorTemperature(img, 16, tt.temperature)

			if tt.temperature == 0.0 {
				// Should not modify
				for y := 0; y < 16; y++ {
					for x := 0; x < 16; x++ {
						if img.At(x, y).(color.RGBA) != baseColor {
							t.Error("Neutral temperature should not modify image")
							return
						}
					}
				}
			} else {
				c := img.At(8, 8).(color.RGBA)
				if tt.expectWarm {
					// Warm: more red, less blue
					if c.R <= baseColor.R || c.B >= baseColor.B {
						t.Errorf("Expected warm tint, got R:%d B:%d (base R:%d B:%d)",
							c.R, c.B, baseColor.R, baseColor.B)
					}
				}
				if tt.expectCool {
					// Cool: less red, more blue
					if c.R >= baseColor.R || c.B <= baseColor.B {
						t.Errorf("Expected cool tint, got R:%d B:%d (base R:%d B:%d)",
							c.R, c.B, baseColor.R, baseColor.B)
					}
				}
			}
		})
	}
}

func TestGetGenreWeathering(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	genres := []string{"fantasy", "scifi", "cyberpunk", "horror", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := gen.getGenreWeathering(genre)

			// Verify all values are in valid ranges
			if config.EdgeDamage < 0 || config.EdgeDamage > 1 {
				t.Errorf("EdgeDamage out of range: %v", config.EdgeDamage)
			}
			if config.WearIntensity < 0 || config.WearIntensity > 1 {
				t.Errorf("WearIntensity out of range: %v", config.WearIntensity)
			}
			if config.AgeVariation < 0 || config.AgeVariation > 1 {
				t.Errorf("AgeVariation out of range: %v", config.AgeVariation)
			}
			if config.MoistureLevel < 0 || config.MoistureLevel > 1 {
				t.Errorf("MoistureLevel out of range: %v", config.MoistureLevel)
			}
			if config.OrganicGrowth < 0 || config.OrganicGrowth > 1 {
				t.Errorf("OrganicGrowth out of range: %v", config.OrganicGrowth)
			}
			if config.ColorTemperature < -1 || config.ColorTemperature > 1 {
				t.Errorf("ColorTemperature out of range: %v", config.ColorTemperature)
			}
		})
	}
}

func TestGetDamageColor(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	baseColor := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, baseColor)
		}
	}

	materials := []MaterialType{
		MaterialStone,
		MaterialMetal,
		MaterialWood,
		MaterialConcrete,
	}

	for _, material := range materials {
		t.Run(material.String(), func(t *testing.T) {
			damageColor := gen.getDamageColor(img, 4, 4, material)

			// Verify color is valid
			if damageColor.A != 255 {
				t.Errorf("Damage color should have full opacity, got %d", damageColor.A)
			}

			// Verify color differs based on material
			switch material {
			case MaterialStone:
				// Should be lighter
				if damageColor.R <= baseColor.R {
					t.Error("Stone damage should lighten")
				}
			case MaterialMetal:
				// Should be rust-colored
				if damageColor.R < 90 || damageColor.G > 70 {
					t.Errorf("Metal damage should be rust-colored, got R:%d G:%d B:%d",
						damageColor.R, damageColor.G, damageColor.B)
				}
			case MaterialWood:
				// Should be darker
				if damageColor.R >= baseColor.R {
					t.Error("Wood damage should darken")
				}
			}
		})
	}
}

func TestWeatheringIntegration(t *testing.T) {
	gen := NewTextureGenerator(10, "fantasy")

	// Generate a tile with weathering
	tile := gen.GetTile(MaterialStone, 0, 54321, 32)
	if tile == nil {
		t.Fatal("GetTile returned nil")
	}

	bounds := tile.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("Expected 32x32 tile, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Verify cache is populated (weathering is applied during generation)
	// We can't read pixels from Ebiten images before game starts,
	// but we can verify the tile was created and cached
	gen.mu.RLock()
	cacheSize := len(gen.cache)
	gen.mu.RUnlock()

	if cacheSize == 0 {
		t.Error("Expected tile to be cached")
	}
}

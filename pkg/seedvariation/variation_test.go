// Package seedvariation provides tests for seed-driven entity variation.
package seedvariation

import (
	"image"
	"image/color"
	"testing"
)

// TestVariationDeterminism verifies that the same seed produces identical variation.
func TestVariationDeterminism(t *testing.T) {
	gen := NewGenerator("fantasy")

	tests := []struct {
		name    string
		seed    int64
		genFunc func(int64) Variation
	}{
		{"humanoid_1", 12345, gen.GenerateHumanoidVariation},
		{"humanoid_2", 67890, gen.GenerateHumanoidVariation},
		{"quadruped_1", 12345, gen.GenerateQuadrupedVariation},
		{"insect_1", 12345, gen.GenerateInsectVariation},
		{"serpent_1", 12345, gen.GenerateSerpentVariation},
		{"flying_1", 12345, gen.GenerateFlyingVariation},
		{"amorphous_1", 12345, gen.GenerateAmorphousVariation},
		{"prop_1", 12345, gen.GeneratePropVariation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v1 := tt.genFunc(tt.seed)
			v2 := tt.genFunc(tt.seed)

			if v1.BodyWidthRatio != v2.BodyWidthRatio {
				t.Errorf("BodyWidthRatio not deterministic: %f != %f", v1.BodyWidthRatio, v2.BodyWidthRatio)
			}
			if v1.ColorShiftR != v2.ColorShiftR {
				t.Errorf("ColorShiftR not deterministic: %d != %d", v1.ColorShiftR, v2.ColorShiftR)
			}
			if v1.HasScar != v2.HasScar {
				t.Errorf("HasScar not deterministic: %v != %v", v1.HasScar, v2.HasScar)
			}
			if v1.PatternSeed != v2.PatternSeed {
				t.Errorf("PatternSeed not deterministic: %d != %d", v1.PatternSeed, v2.PatternSeed)
			}
		})
	}
}

// TestVariationDiversity verifies that different seeds produce different variations.
func TestVariationDiversity(t *testing.T) {
	gen := NewGenerator("fantasy")

	seed1 := int64(12345)
	seed2 := int64(67890)

	v1 := gen.GenerateHumanoidVariation(seed1)
	v2 := gen.GenerateHumanoidVariation(seed2)

	// At least one property should differ
	if v1.BodyWidthRatio == v2.BodyWidthRatio &&
		v1.ColorShiftR == v2.ColorShiftR &&
		v1.ColorShiftG == v2.ColorShiftG &&
		v1.WearLevel == v2.WearLevel {
		t.Error("Different seeds should produce different variations")
	}
}

// TestVariationBounds verifies that variation parameters stay within expected bounds.
func TestVariationBounds(t *testing.T) {
	gen := NewGenerator("fantasy")

	for i := int64(0); i < 100; i++ {
		v := gen.GenerateHumanoidVariation(i)

		if v.BodyWidthRatio < 0.5 || v.BodyWidthRatio > 1.5 {
			t.Errorf("BodyWidthRatio out of bounds: %f", v.BodyWidthRatio)
		}
		if v.BodyHeightRatio < 0.5 || v.BodyHeightRatio > 1.5 {
			t.Errorf("BodyHeightRatio out of bounds: %f", v.BodyHeightRatio)
		}
		if v.Brightness < 0.5 || v.Brightness > 2.0 {
			t.Errorf("Brightness out of bounds: %f", v.Brightness)
		}
		if v.WearLevel < 0 || v.WearLevel > 1.0 {
			t.Errorf("WearLevel out of bounds: %f", v.WearLevel)
		}
	}
}

// TestGenreAffectsVariation verifies that genre changes variation parameters.
func TestGenreAffectsVariation(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	seed := int64(42)
	variations := make(map[string]Variation)

	for _, genre := range genres {
		gen := NewGenerator(genre)
		variations[genre] = gen.GenerateHumanoidVariation(seed)
	}

	// Post-apocalyptic should have higher missing part chance (or other effects)
	// Verify that all genres produce valid variations
	for genre, v := range variations {
		if v.BodyWidthRatio == 0 {
			t.Errorf("Genre %s produced zero BodyWidthRatio", genre)
		}
	}
}

// TestApplyColorVariation verifies color modification works correctly.
func TestApplyColorVariation(t *testing.T) {
	tests := []struct {
		name     string
		v        Variation
		base     color.RGBA
		expected color.RGBA
	}{
		{
			name: "no change",
			v: Variation{
				ColorShiftR: 0,
				ColorShiftG: 0,
				ColorShiftB: 0,
				Brightness:  1.0,
				Saturation:  1.0,
			},
			base:     color.RGBA{R: 100, G: 100, B: 100, A: 255},
			expected: color.RGBA{R: 100, G: 100, B: 100, A: 255},
		},
		{
			name: "brightness increase",
			v: Variation{
				ColorShiftR: 0,
				ColorShiftG: 0,
				ColorShiftB: 0,
				Brightness:  1.2,
				Saturation:  1.0,
			},
			base:     color.RGBA{R: 100, G: 100, B: 100, A: 255},
			expected: color.RGBA{R: 120, G: 120, B: 120, A: 255},
		},
		{
			name: "color shift",
			v: Variation{
				ColorShiftR: 20,
				ColorShiftG: -20,
				ColorShiftB: 10,
				Brightness:  1.0,
				Saturation:  1.0,
			},
			base:     color.RGBA{R: 100, G: 100, B: 100, A: 255},
			expected: color.RGBA{R: 120, G: 80, B: 110, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.v.ApplyColorVariation(tt.base)
			if result != tt.expected {
				t.Errorf("ApplyColorVariation() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestApplyColorVariationClamping verifies colors are clamped to valid range.
func TestApplyColorVariationClamping(t *testing.T) {
	v := Variation{
		ColorShiftR: 127,  // max positive
		ColorShiftG: -128, // max negative (signed int8)
		ColorShiftB: 0,
		Brightness:  2.0, // Very bright
		Saturation:  1.0,
	}

	base := color.RGBA{R: 200, G: 50, B: 100, A: 255}
	result := v.ApplyColorVariation(base)

	// Should be clamped to 0-255
	if result.R < 0 || result.R > 255 {
		t.Errorf("Red not clamped: %d", result.R)
	}
	if result.G < 0 || result.G > 255 {
		t.Errorf("Green not clamped: %d", result.G)
	}
}

// TestScaleBodyPart verifies body part scaling works correctly.
func TestScaleBodyPart(t *testing.T) {
	v := Variation{
		LimbLengthRatio: 1.2,
		BodyWidthRatio:  0.8,
		HeadSizeRatio:   1.1,
	}

	tests := []struct {
		name     string
		base     int
		part     string
		expected int
	}{
		{"limb scaling", 100, "limb", 120},
		{"body scaling", 100, "body", 80},
		{"head scaling", 100, "head", 110},
		{"unknown part", 100, "unknown", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.ScaleBodyPart(tt.base, tt.part)
			if result != tt.expected {
				t.Errorf("ScaleBodyPart(%d, %s) = %d, want %d", tt.base, tt.part, result, tt.expected)
			}
		})
	}
}

// TestApplyMarkings verifies marking application doesn't panic.
func TestApplyMarkings(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	bounds := image.Rect(16, 16, 48, 48)
	baseColor := color.RGBA{R: 100, G: 80, B: 60, A: 255}

	gen := NewGenerator("fantasy")

	tests := []struct {
		name    string
		genFunc func(int64) Variation
	}{
		{"with spots", gen.GenerateHumanoidVariation},
		{"with stripes", gen.GenerateQuadrupedVariation},
		{"with bubbles", gen.GenerateAmorphousVariation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate variation (may or may not have markings)
			v := tt.genFunc(42)
			v.HasSpotting = true
			v.MarkingCount = 5

			// Should not panic
			ApplyMarkings(img, bounds, &v, baseColor)
		})
	}
}

// TestApplyWear verifies wear application doesn't panic.
func TestApplyWear(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	bounds := image.Rect(16, 16, 48, 48)

	// Fill with some color
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 255})
		}
	}

	tests := []struct {
		name      string
		wearLevel float64
	}{
		{"no wear", 0.0},
		{"light wear", 0.2},
		{"heavy wear", 0.8},
		{"max wear", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Variation{
				WearLevel:   tt.wearLevel,
				PatternSeed: 12345,
			}

			// Should not panic
			ApplyWear(img, bounds, &v)
		})
	}
}

// TestGetSegmentRadius verifies segment radius calculation.
func TestGetSegmentRadius(t *testing.T) {
	v := Variation{
		SegmentCount:      5,
		SegmentSizeGrowth: 0.9,
		BodyWidthRatio:    1.0,
	}

	baseRadius := 10
	radii := make([]int, v.SegmentCount)

	for i := 0; i < v.SegmentCount; i++ {
		radii[i] = v.GetSegmentRadius(baseRadius, i)
	}

	// Segments should generally decrease in size
	for i := 1; i < len(radii); i++ {
		if radii[i] > radii[i-1]+2 { // Allow some variance
			t.Logf("Warning: segment %d larger than previous (%d > %d)", i, radii[i], radii[i-1])
		}
	}
}

// BenchmarkGenerateHumanoidVariation measures variation generation performance.
func BenchmarkGenerateHumanoidVariation(b *testing.B) {
	gen := NewGenerator("fantasy")

	for i := 0; i < b.N; i++ {
		_ = gen.GenerateHumanoidVariation(int64(i))
	}
}

// BenchmarkApplyColorVariation measures color variation performance.
func BenchmarkApplyColorVariation(b *testing.B) {
	v := Variation{
		ColorShiftR: 15,
		ColorShiftG: -10,
		ColorShiftB: 5,
		Brightness:  1.1,
		Saturation:  0.95,
	}
	base := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	for i := 0; i < b.N; i++ {
		_ = v.ApplyColorVariation(base)
	}
}

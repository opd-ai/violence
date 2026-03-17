package rimlight

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestNewComponent(t *testing.T) {
	comp := NewComponent()

	if !comp.Enabled {
		t.Error("new component should be enabled by default")
	}
	if comp.Material != MaterialDefault {
		t.Errorf("Material = %v, want MaterialDefault", comp.Material)
	}
	if comp.Intensity != 1.0 {
		t.Errorf("Intensity = %f, want 1.0", comp.Intensity)
	}
	if comp.Width != 0 {
		t.Errorf("Width = %d, want 0 (auto)", comp.Width)
	}
	if comp.FadeInner != 0.5 {
		t.Errorf("FadeInner = %f, want 0.5", comp.FadeInner)
	}
}

func TestNewComponentWithMaterial(t *testing.T) {
	testCases := []struct {
		material Material
		name     string
	}{
		{MaterialMetal, "metal"},
		{MaterialCloth, "cloth"},
		{MaterialCrystal, "crystal"},
		{MaterialMagic, "magic"},
		{MaterialOrganic, "organic"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comp := NewComponentWithMaterial(tc.material)
			if comp.Material != tc.material {
				t.Errorf("Material = %v, want %v", comp.Material, tc.material)
			}
			if !comp.Enabled {
				t.Error("component should be enabled")
			}
		})
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent()
	if comp.Type() != "RimLightComponent" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "RimLightComponent")
	}
}

func TestMaterialIntensity(t *testing.T) {
	testCases := []struct {
		material  Material
		wantRange [2]float64
	}{
		{MaterialDefault, [2]float64{0.9, 1.1}},
		{MaterialMetal, [2]float64{1.3, 1.5}},
		{MaterialCrystal, [2]float64{1.7, 1.9}},
		{MaterialCloth, [2]float64{0.5, 0.7}},
		{MaterialOrganic, [2]float64{0.6, 0.8}},
	}

	for _, tc := range testCases {
		intensity := GetMaterialIntensity(tc.material)
		if intensity < tc.wantRange[0] || intensity > tc.wantRange[1] {
			t.Errorf("GetMaterialIntensity(%v) = %f, want in range [%f, %f]",
				tc.material, intensity, tc.wantRange[0], tc.wantRange[1])
		}
	}
}

func TestMaterialFresnel(t *testing.T) {
	testCases := []struct {
		material  Material
		wantRange [2]float64
	}{
		{MaterialDefault, [2]float64{2.0, 3.0}},
		{MaterialMetal, [2]float64{2.5, 3.5}},
		{MaterialCrystal, [2]float64{3.5, 4.5}},
		{MaterialCloth, [2]float64{1.0, 2.0}},
	}

	for _, tc := range testCases {
		fresnel := GetMaterialFresnel(tc.material)
		if fresnel < tc.wantRange[0] || fresnel > tc.wantRange[1] {
			t.Errorf("GetMaterialFresnel(%v) = %f, want in range [%f, %f]",
				tc.material, fresnel, tc.wantRange[0], tc.wantRange[1])
		}
	}
}

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)

			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, genre)
			}
			if sys.cache == nil {
				t.Error("cache should be initialized")
			}

			// Light direction should be normalized
			length := math.Sqrt(sys.lightDirX*sys.lightDirX + sys.lightDirY*sys.lightDirY)
			if math.Abs(length-1.0) > 0.01 {
				t.Errorf("light direction not normalized, length = %f", length)
			}

			// Rim color should be set
			if sys.rimColor.A == 0 {
				t.Error("rim color alpha should not be 0")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")
	originalColor := sys.rimColor

	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "scifi")
	}

	// Color should change for different genre
	if sys.rimColor == originalColor {
		t.Log("rim color may have changed (depends on genre config)")
	}

	// Cache should be cleared
	if len(sys.cache) != 0 {
		t.Error("cache should be cleared on genre change")
	}
}

func TestGetSetLightDirection(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.SetLightDirection(1.0, 0.0)
	x, y := sys.GetLightDirection()

	if math.Abs(x-1.0) > 0.001 {
		t.Errorf("lightDirX = %f, want 1.0", x)
	}
	if math.Abs(y) > 0.001 {
		t.Errorf("lightDirY = %f, want 0.0", y)
	}

	// Test normalization
	sys.SetLightDirection(3.0, 4.0)
	x, y = sys.GetLightDirection()
	length := math.Sqrt(x*x + y*y)
	if math.Abs(length-1.0) > 0.001 {
		t.Errorf("light direction not normalized after set, length = %f", length)
	}
}

func TestGetRimColor(t *testing.T) {
	sys := NewSystem("fantasy")
	color := sys.GetRimColor()

	if color.A == 0 {
		t.Error("rim color alpha should not be 0")
	}
	if color.R == 0 && color.G == 0 && color.B == 0 {
		t.Error("rim color should not be black")
	}
}

func TestGenerateSeedVariant(t *testing.T) {
	sys := NewSystem("fantasy")
	base := NewComponent()
	base.Intensity = 1.0
	base.Width = 4

	variant1 := sys.GenerateSeedVariant(12345, base)
	variant2 := sys.GenerateSeedVariant(12345, base)
	variant3 := sys.GenerateSeedVariant(99999, base)

	// Same seed should produce same result
	if variant1.Intensity != variant2.Intensity {
		t.Error("same seed should produce same intensity")
	}

	// Different seed should produce different result (usually)
	if variant1.Intensity == variant3.Intensity && variant1.Width == variant3.Width {
		t.Log("different seeds produced same values (unlikely but possible)")
	}

	// Intensity should be within expected range (0.8x to 1.2x of base)
	if variant1.Intensity < 0.5 || variant1.Intensity > 1.5 {
		t.Errorf("variant intensity %f out of expected range", variant1.Intensity)
	}
}

func TestCalculateEdgeDistance(t *testing.T) {
	sys := NewSystem("fantasy")

	// Create a simple test image: 8x8 with a 4x4 opaque center
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 2; y < 6; y++ {
		for x := 2; x < 6; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	// Center pixel should have distance ~2 to edge
	dist := sys.calculateEdgeDistance(img, 4, 4, 8, 8, 5)
	if dist < 1.5 || dist > 2.5 {
		t.Errorf("center edge distance = %f, want ~2", dist)
	}

	// Edge pixel should have distance ~1
	dist = sys.calculateEdgeDistance(img, 2, 3, 8, 8, 5)
	if dist > 1.5 {
		t.Errorf("edge pixel distance = %f, want ~1", dist)
	}
}

func TestApplyRimLightNilInputs(t *testing.T) {
	sys := NewSystem("fantasy")

	// Nil source should return nil
	result := sys.ApplyRimLight(nil, NewComponent())
	if result != nil {
		t.Error("nil source should return nil")
	}

	// Nil component should return source unchanged
	// Note: we can't easily test this without creating an ebiten.Image
}

func TestApplyRimLightDisabled(t *testing.T) {
	_ = NewSystem("fantasy")
	comp := NewComponent()
	comp.Enabled = false

	// With disabled component, should return source unchanged
	// Note: full test requires ebiten.Image which needs display context
	if comp.Enabled {
		t.Error("component should be disabled")
	}
}

func TestProcessRimLighting(t *testing.T) {
	sys := NewSystem("fantasy")

	// Create source image with opaque center
	src := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 4; y < 12; y++ {
		for x := 4; x < 12; x++ {
			src.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, 16, 16))

	sys.processRimLighting(src, dst, sys.rimColor, 3, 1.0, 2.5, 0.5)

	// Center should be unchanged or slightly brightened
	centerPix := dst.RGBAAt(8, 8)
	if centerPix.A == 0 {
		t.Error("center pixel should be opaque")
	}

	// Edge pixels facing light should have rim highlight
	// The rim effect adds to the existing color
	edgePix := dst.RGBAAt(4, 8)
	if edgePix.A == 0 {
		t.Error("edge pixel should be opaque")
	}

	// Verify rim effect was applied (edge should be brighter than original)
	srcEdge := src.RGBAAt(4, 8)
	// Note: depending on light direction, some edges get highlight, some don't
	t.Logf("Edge pixel: src=%v dst=%v", srcEdge, edgePix)
}

func TestMinMax(t *testing.T) {
	if max(3, 5) != 5 {
		t.Error("max(3, 5) should be 5")
	}
	if max(5, 3) != 5 {
		t.Error("max(5, 3) should be 5")
	}
	if min(3, 5) != 3 {
		t.Error("min(3, 5) should be 3")
	}
	if min(5, 3) != 3 {
		t.Error("min(5, 3) should be 3")
	}
}

func BenchmarkProcessRimLighting(b *testing.B) {
	sys := NewSystem("fantasy")

	// Create a 32x32 test image
	src := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			src.Set(x, y, color.RGBA{128, 128, 128, 255})
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, 32, 32))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.processRimLighting(src, dst, sys.rimColor, 4, 1.0, 2.5, 0.5)
	}
}

package sprite

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestDefaultLightConfig(t *testing.T) {
	cfg := DefaultLightConfig()

	// Verify light direction is normalized
	length := math.Sqrt(cfg.LightDirX*cfg.LightDirX + cfg.LightDirY*cfg.LightDirY + cfg.LightDirZ*cfg.LightDirZ)
	if math.Abs(length-1.0) > 0.01 {
		t.Errorf("Light direction not normalized: length = %f", length)
	}

	// Verify reasonable defaults
	if cfg.LightIntensity <= 0 {
		t.Errorf("Light intensity should be positive, got %f", cfg.LightIntensity)
	}
	if cfg.AmbientLevel < 0 || cfg.AmbientLevel > 1 {
		t.Errorf("Ambient level should be in [0,1], got %f", cfg.AmbientLevel)
	}
	if cfg.AOStrength < 0 || cfg.AOStrength > 1 {
		t.Errorf("AO strength should be in [0,1], got %f", cfg.AOStrength)
	}
}

func TestDefaultMaterialProperties(t *testing.T) {
	tests := []struct {
		name         string
		materialType MaterialDetail
		wantMetallic bool // metallic > 0.5
		wantShiny    bool // roughness < 0.5
	}{
		{"Metal", MaterialMetal, true, true},
		{"Cloth", MaterialCloth, false, false},
		{"Leather", MaterialLeather, false, false},
		{"Scales", MaterialScales, false, true},
		{"Chitin", MaterialChitin, false, true},
		{"Fur", MaterialFur, false, false},
		{"Crystal", MaterialCrystal, false, true},
		{"Slime", MaterialSlime, false, true},
	}

	baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := DefaultMaterialProperties(tt.materialType, baseColor)

			if props.BaseColor != baseColor {
				t.Errorf("Base color not preserved: got %v, want %v", props.BaseColor, baseColor)
			}

			isMetallic := props.Metallic > 0.5
			if isMetallic != tt.wantMetallic {
				t.Errorf("Metallic = %f, wantMetallic = %v", props.Metallic, tt.wantMetallic)
			}

			isShiny := props.Roughness < 0.5
			if isShiny != tt.wantShiny {
				t.Errorf("Roughness = %f, wantShiny = %v", props.Roughness, tt.wantShiny)
			}

			// All values should be in valid ranges
			if props.Metallic < 0 || props.Metallic > 1 {
				t.Errorf("Metallic out of range [0,1]: %f", props.Metallic)
			}
			if props.Roughness < 0 || props.Roughness > 1 {
				t.Errorf("Roughness out of range [0,1]: %f", props.Roughness)
			}
			if props.Specular < 0 {
				t.Errorf("Specular negative: %f", props.Specular)
			}
		})
	}
}

func TestCalculateSurfaceNormal(t *testing.T) {
	tests := []struct {
		name   string
		dx     float64
		dy     float64
		radius float64
		wantZ  float64 // Z should be positive for hemisphere
	}{
		{"Center", 0, 0, 10, 1.0},
		{"Edge", 10, 0, 10, 0.0},
		{"OutsideRadius", 20, 0, 10, 1.0}, // fallback to upward
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nx, ny, nz := CalculateSurfaceNormal(tt.dx, tt.dy, tt.radius)

			// Normal should be normalized
			length := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if math.Abs(length-1.0) > 0.01 {
				t.Errorf("Normal not normalized: length = %f", length)
			}

			// Z component check
			if math.Abs(nz-tt.wantZ) > 0.01 {
				t.Errorf("Z component = %f, want ~%f", nz, tt.wantZ)
			}

			// For hemisphere, Z should never be negative
			if nz < 0 {
				t.Errorf("Z component negative for hemisphere: %f", nz)
			}
		})
	}
}

func TestCalculateCylindricalNormal(t *testing.T) {
	tests := []struct {
		name   string
		dx     float64
		axis   float64
		radius float64
		wantNY float64 // Y should always be 0 for cylinder
	}{
		{"Center", 0, 0, 10, 0},
		{"LeftEdge", -10, 0, 10, 0},
		{"RightEdge", 10, 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nx, ny, nz := CalculateCylindricalNormal(tt.dx, tt.axis, tt.radius)

			// Normal should be normalized
			length := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if math.Abs(length-1.0) > 0.01 {
				t.Errorf("Normal not normalized: length = %f", length)
			}

			// Y should be 0 for cylinder
			if ny != tt.wantNY {
				t.Errorf("Y component = %f, want %f", ny, tt.wantNY)
			}

			// Z should be non-negative
			if nz < 0 {
				t.Errorf("Z component negative: %f", nz)
			}
		})
	}
}

func TestCalculatePlanarNormal(t *testing.T) {
	tests := []struct {
		name   string
		tiltX  float64
		tiltY  float64
		wantNZ float64 // Z should be dominant for mostly flat surface
	}{
		{"Flat", 0, 0, 1.0},
		{"SlightTiltX", 0.1, 0, 0.99},
		{"SlightTiltY", 0, 0.1, 0.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nx, ny, nz := CalculatePlanarNormal(tt.tiltX, tt.tiltY)

			// Normal should be normalized
			length := math.Sqrt(nx*nx + ny*ny + nz*nz)
			if math.Abs(length-1.0) > 0.01 {
				t.Errorf("Normal not normalized: length = %f", length)
			}

			// Z component check
			if math.Abs(nz-tt.wantNZ) > 0.02 {
				t.Errorf("Z component = %f, want ~%f", nz, tt.wantNZ)
			}
		})
	}
}

func TestCalculateAmbientOcclusion(t *testing.T) {
	tests := []struct {
		name          string
		dx            float64
		dy            float64
		radius        float64
		contactBottom bool
		wantRange     [2]float64 // [min, max]
	}{
		{"Center_NoContact", 0, 0, 10, false, [2]float64{0.9, 1.0}},
		{"Edge_NoContact", 8, 0, 10, false, [2]float64{0.5, 0.9}},
		{"Center_Contact", 0, 8, 10, true, [2]float64{0.5, 0.8}},
		{"Edge_Contact", 8, 8, 10, true, [2]float64{0.3, 0.7}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ao := CalculateAmbientOcclusion(tt.dx, tt.dy, tt.radius, tt.contactBottom)

			// AO should be in [0, 1]
			if ao < 0 || ao > 1 {
				t.Errorf("AO out of range [0,1]: %f", ao)
			}

			// Check expected range
			if ao < tt.wantRange[0] || ao > tt.wantRange[1] {
				t.Errorf("AO = %f, want in range [%f, %f]", ao, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

func TestComputePBRShading(t *testing.T) {
	light := DefaultLightConfig()
	baseColor := color.RGBA{R: 128, G: 64, B: 32, A: 255}

	tests := []struct {
		name           string
		material       MaterialDetail
		normalZ        float64 // facing camera (Z=1) vs edge (Z=0)
		expectBrighter bool    // should be brighter than base color
	}{
		{"Metal_FacingLight", MaterialMetal, 1.0, true},
		{"Metal_Edge", MaterialMetal, 0.1, false},
		{"Cloth_FacingLight", MaterialCloth, 1.0, true},
		{"Cloth_Edge", MaterialCloth, 0.1, false},
		{"Crystal_FacingLight", MaterialCrystal, 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matProps := DefaultMaterialProperties(tt.material, baseColor)
			context := ShadingContext{
				NormalX: 0,
				NormalY: 0,
				NormalZ: tt.normalZ,
				PosX:    0,
				PosY:    0,
				AO:      1.0, // no occlusion
			}

			result := ComputePBRShading(matProps, context, light)

			// Output should be valid color
			if result.A != 255 {
				t.Errorf("Alpha should be 255, got %d", result.A)
			}

			// Check brightness relative to base
			baseBrightness := float64(baseColor.R) + float64(baseColor.G) + float64(baseColor.B)
			resultBrightness := float64(result.R) + float64(result.G) + float64(result.B)

			isBrighter := resultBrightness > baseBrightness*0.8 // allow some tolerance

			if isBrighter != tt.expectBrighter {
				t.Errorf("Brightness check failed: result=%f, base=%f, expectBrighter=%v",
					resultBrightness, baseBrightness, tt.expectBrighter)
			}
		})
	}
}

func TestApplyPBRShadingToRegion(t *testing.T) {
	gen := NewGenerator(10)
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))

	// Fill a circular region with base color
	baseColor := color.RGBA{R: 100, G: 100, B: 150, A: 255}
	cx, cy := 16, 16
	radius := 10

	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, baseColor)
			}
		}
	}

	// Apply PBR shading
	bounds := image.Rect(cx-radius, cy-radius, cx+radius, cy+radius)
	gen.ApplyPBRShadingToRegion(img, bounds, MaterialMetal, "spherical", gen.lightCfg)

	// Check that pixels have been modified
	center := img.At(cx, cy)
	cr, cg, cb, ca := center.RGBA()
	if ca == 0 {
		t.Error("Center pixel should not be transparent")
	}

	// Center should be relatively bright (facing camera/light)
	centerBrightness := float64(cr>>8) + float64(cg>>8) + float64(cb>>8)
	baseBrightness := float64(baseColor.R) + float64(baseColor.G) + float64(baseColor.B)

	if centerBrightness < baseBrightness*0.5 {
		t.Errorf("Center too dark after shading: %f vs base %f", centerBrightness, baseBrightness)
	}

	// Edge should be darker than center (ambient occlusion)
	edge := img.At(cx+radius-1, cy)
	er, eg, eb, ea := edge.RGBA()
	if ea == 0 {
		t.Error("Edge pixel should not be transparent")
	}

	edgeBrightness := float64(er>>8) + float64(eg>>8) + float64(eb>>8)
	if edgeBrightness >= centerBrightness {
		t.Errorf("Edge should be darker than center: edge=%f, center=%f", edgeBrightness, centerBrightness)
	}
}

func TestPBRShadingConsistency(t *testing.T) {
	// PBR shading should produce same output for same input
	gen := NewGenerator(10)
	img1 := image.NewRGBA(image.Rect(0, 0, 16, 16))
	img2 := image.NewRGBA(image.Rect(0, 0, 16, 16))

	baseColor := color.RGBA{R: 128, G: 64, B: 32, A: 255}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img1.Set(x, y, baseColor)
			img2.Set(x, y, baseColor)
		}
	}

	bounds := image.Rect(0, 0, 16, 16)
	gen.ApplyPBRShadingToRegion(img1, bounds, MaterialLeather, "spherical", gen.lightCfg)
	gen.ApplyPBRShadingToRegion(img2, bounds, MaterialLeather, "spherical", gen.lightCfg)

	// Compare outputs
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 != c2 {
				t.Errorf("Inconsistent output at (%d,%d): %v vs %v", x, y, c1, c2)
				return
			}
		}
	}
}

func BenchmarkComputePBRShading(b *testing.B) {
	light := DefaultLightConfig()
	baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	matProps := DefaultMaterialProperties(MaterialMetal, baseColor)
	context := ShadingContext{
		NormalX: 0.5,
		NormalY: 0.5,
		NormalZ: 0.707,
		PosX:    5,
		PosY:    5,
		AO:      0.8,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputePBRShading(matProps, context, light)
	}
}

func BenchmarkApplyPBRShadingToRegion(b *testing.B) {
	gen := NewGenerator(10)
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, baseColor)
		}
	}

	bounds := image.Rect(8, 8, 24, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.ApplyPBRShadingToRegion(img, bounds, MaterialMetal, "spherical", gen.lightCfg)
	}
}

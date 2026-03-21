package sprite

import (
	"image/color"
	"math"
	"testing"
)

func TestDefaultNormalPerturbConfig(t *testing.T) {
	cfg := DefaultNormalPerturbConfig()

	if cfg.Intensity < 0 || cfg.Intensity > 1.0 {
		t.Errorf("Intensity out of expected range: %f", cfg.Intensity)
	}
	if cfg.Scale <= 0 {
		t.Errorf("Scale should be positive: %f", cfg.Scale)
	}
}

func TestNormalPerturbForMaterial(t *testing.T) {
	materials := []MaterialDetail{
		MaterialScales,
		MaterialFur,
		MaterialChitin,
		MaterialMembrane,
		MaterialMetal,
		MaterialCloth,
		MaterialLeather,
		MaterialCrystal,
		MaterialSlime,
	}

	for _, mat := range materials {
		cfg := NormalPerturbForMaterial(mat)
		if cfg.Intensity < 0 || cfg.Intensity > 1.0 {
			t.Errorf("Material %v: Intensity out of range: %f", mat, cfg.Intensity)
		}
		if cfg.Scale <= 0 {
			t.Errorf("Material %v: Scale should be positive: %f", mat, cfg.Scale)
		}
	}
}

func TestPerturbNormalPreservesApproximateLength(t *testing.T) {
	materials := []MaterialDetail{
		MaterialScales,
		MaterialFur,
		MaterialMetal,
		MaterialCloth,
	}

	seed := int64(12345)
	originalNormal := struct{ x, y, z float64 }{0.0, 0.0, 1.0}

	for _, mat := range materials {
		cfg := NormalPerturbForMaterial(mat)

		for x := 0; x < 10; x++ {
			for y := 0; y < 10; y++ {
				nx, ny, nz := PerturbNormal(
					originalNormal.x, originalNormal.y, originalNormal.z,
					x, y, mat, seed, cfg,
				)

				length := math.Sqrt(nx*nx + ny*ny + nz*nz)
				if math.Abs(length-1.0) > 0.01 {
					t.Errorf("Material %v at (%d,%d): normal length should be ~1.0, got %f", mat, x, y, length)
				}
			}
		}
	}
}

func TestPerturbNormalDeterministic(t *testing.T) {
	seed := int64(67890)
	cfg := NormalPerturbForMaterial(MaterialScales)

	// Same inputs should produce same outputs
	nx1, ny1, nz1 := PerturbNormal(0, 0, 1, 5, 5, MaterialScales, seed, cfg)
	nx2, ny2, nz2 := PerturbNormal(0, 0, 1, 5, 5, MaterialScales, seed, cfg)

	if nx1 != nx2 || ny1 != ny2 || nz1 != nz2 {
		t.Errorf("Perturbation not deterministic: (%f,%f,%f) vs (%f,%f,%f)",
			nx1, ny1, nz1, nx2, ny2, nz2)
	}
}

func TestPerturbNormalVariesWithPosition(t *testing.T) {
	seed := int64(11111)
	cfg := NormalPerturbForMaterial(MaterialScales)

	// Different positions should produce different perturbations
	nx1, ny1, nz1 := PerturbNormal(0, 0, 1, 0, 0, MaterialScales, seed, cfg)
	nx2, ny2, nz2 := PerturbNormal(0, 0, 1, 10, 10, MaterialScales, seed, cfg)

	if nx1 == nx2 && ny1 == ny2 && nz1 == nz2 {
		t.Error("Perturbation should vary with position")
	}
}

func TestPerturbNormalZeroIntensityNoChange(t *testing.T) {
	cfg := NormalPerturbConfig{
		Intensity:    0.0,
		Scale:        1.0,
		MaterialBias: 0.0,
	}

	originalX, originalY, originalZ := 0.3, 0.4, 0.866025 // Normalized
	nx, ny, nz := PerturbNormal(originalX, originalY, originalZ, 5, 5, MaterialMetal, 12345, cfg)

	if nx != originalX || ny != originalY || nz != originalZ {
		t.Errorf("Zero intensity should not change normal: (%f,%f,%f) -> (%f,%f,%f)",
			originalX, originalY, originalZ, nx, ny, nz)
	}
}

func TestPerturbScalesPattern(t *testing.T) {
	scale := 1.0
	seed := int64(99999)

	// Test scale pattern generates variation
	var lastX, lastY float64
	variationFound := false

	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			cfg := NormalPerturbConfig{Intensity: 1.0, Scale: scale}
			px, py, _ := PerturbNormal(0, 0, 1, x, y, MaterialScales, seed, cfg)

			if x > 0 && (px != lastX || py != lastY) {
				variationFound = true
			}
			lastX, lastY = px, py
		}
	}

	if !variationFound {
		t.Error("Scales pattern should produce spatial variation")
	}
}

func TestComputePBRShadingWithPerturbation(t *testing.T) {
	material := MaterialProperties{
		BaseColor: color.RGBA{R: 128, G: 64, B: 32, A: 255},
		Metallic:  0.0,
		Roughness: 0.5,
		Specular:  0.5,
	}

	context := ShadingContext{
		NormalX: 0,
		NormalY: 0,
		NormalZ: 1,
		PosX:    0,
		PosY:    0,
		AO:      1.0,
	}

	light := DefaultLightConfig()

	// Test that function returns valid color
	result := ComputePBRShadingWithPerturbation(
		material, context, light,
		MaterialLeather, 10, 10, 12345,
	)

	if result.A != 255 {
		t.Errorf("Result alpha should be 255, got %d", result.A)
	}

	// Color values should be valid (0-255)
	if result.R > 255 || result.G > 255 || result.B > 255 {
		t.Errorf("Invalid color values: R=%d G=%d B=%d", result.R, result.G, result.B)
	}
}

func TestPerturbationDiffersFromBasePBR(t *testing.T) {
	material := MaterialProperties{
		BaseColor: color.RGBA{R: 150, G: 150, B: 150, A: 255},
		Metallic:  0.9,
		Roughness: 0.3,
		Specular:  1.0,
	}

	context := ShadingContext{
		NormalX: 0,
		NormalY: 0,
		NormalZ: 1,
		PosX:    0,
		PosY:    0,
		AO:      1.0,
	}

	light := DefaultLightConfig()

	// Get base PBR result
	baseResult := ComputePBRShading(material, context, light)

	// Get perturbed result at multiple positions
	differentFound := false
	for x := 0; x < 20; x++ {
		for y := 0; y < 20; y++ {
			perturbedResult := ComputePBRShadingWithPerturbation(
				material, context, light,
				MaterialMetal, x, y, 12345,
			)

			// At least some positions should differ from base
			if perturbedResult.R != baseResult.R ||
				perturbedResult.G != baseResult.G ||
				perturbedResult.B != baseResult.B {
				differentFound = true
				break
			}
		}
		if differentFound {
			break
		}
	}

	if !differentFound {
		t.Error("Perturbation should produce different results than base PBR")
	}
}

func TestAllMaterialPerturbationsFunction(t *testing.T) {
	materials := []MaterialDetail{
		MaterialScales,
		MaterialFur,
		MaterialChitin,
		MaterialMembrane,
		MaterialMetal,
		MaterialCloth,
		MaterialLeather,
		MaterialCrystal,
		MaterialSlime,
	}

	seed := int64(54321)
	originalNormal := struct{ x, y, z float64 }{0.2, 0.3, 0.9327}

	for _, mat := range materials {
		cfg := NormalPerturbForMaterial(mat)

		// Should not panic and should produce valid normals
		for i := 0; i < 100; i++ {
			x := i % 10
			y := i / 10

			nx, ny, nz := PerturbNormal(
				originalNormal.x, originalNormal.y, originalNormal.z,
				x, y, mat, seed, cfg,
			)

			// Check for NaN
			if math.IsNaN(nx) || math.IsNaN(ny) || math.IsNaN(nz) {
				t.Errorf("Material %v produced NaN at (%d,%d)", mat, x, y)
			}

			// Check for Inf
			if math.IsInf(nx, 0) || math.IsInf(ny, 0) || math.IsInf(nz, 0) {
				t.Errorf("Material %v produced Inf at (%d,%d)", mat, x, y)
			}
		}
	}
}

func BenchmarkPerturbNormal(b *testing.B) {
	cfg := NormalPerturbForMaterial(MaterialScales)
	seed := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PerturbNormal(0, 0, 1, i%100, i/100, MaterialScales, seed, cfg)
	}
}

func BenchmarkComputePBRShadingWithPerturbation(b *testing.B) {
	material := MaterialProperties{
		BaseColor: color.RGBA{R: 128, G: 128, B: 128, A: 255},
		Metallic:  0.5,
		Roughness: 0.5,
		Specular:  0.5,
	}
	context := ShadingContext{
		NormalX: 0, NormalY: 0, NormalZ: 1,
		PosX: 0, PosY: 0, AO: 1.0,
	}
	light := DefaultLightConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputePBRShadingWithPerturbation(material, context, light, MaterialMetal, i%100, i/100, 12345)
	}
}

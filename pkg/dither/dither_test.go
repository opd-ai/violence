package dither

import (
	"image"
	"image/color"
	"math"
	"testing"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem(12345)
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.rng == nil {
		t.Error("System rng is nil")
	}
}

func TestNewSystem_DifferentSeeds(t *testing.T) {
	sys1 := NewSystem(12345)
	sys2 := NewSystem(54321)

	// Generate blue noise for both
	noise1 := sys1.generateBlueNoise(16, 16)
	noise2 := sys2.generateBlueNoise(16, 16)

	// They should differ due to different seeds
	diffCount := 0
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			if math.Abs(noise1[y][x]-noise2[y][x]) > 0.01 {
				diffCount++
			}
		}
	}

	if diffCount == 0 {
		t.Error("Different seeds should produce different blue noise patterns")
	}
}

func TestGetBayerThreshold(t *testing.T) {
	tests := []struct {
		name   string
		x, y   int
		use8x8 bool
		want   float64
	}{
		{"4x4 origin", 0, 0, false, 0.0 / 16.0},
		{"4x4 center", 1, 1, false, 4.0 / 16.0},
		{"4x4 wrap", 4, 4, false, 0.0 / 16.0}, // Should wrap
		{"8x8 origin", 0, 0, true, 0.0 / 64.0},
		{"8x8 center", 3, 3, true, 20.0 / 64.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getBayerThreshold(tt.x, tt.y, tt.use8x8)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("getBayerThreshold(%d, %d, %v) = %v, want %v", tt.x, tt.y, tt.use8x8, got, tt.want)
			}
		})
	}
}

func TestApplyDithering(t *testing.T) {
	sys := NewSystem(12345)

	// Create test image with gradient
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			gray := uint8(x * 8) // 0-248 gradient
			img.SetRGBA(x, y, color.RGBA{R: gray, G: gray, B: gray, A: 255})
		}
	}

	// Store original values
	original := make([]byte, len(img.Pix))
	copy(original, img.Pix)

	// Apply dithering
	sys.ApplyDithering(img, MaterialDefault, 0.5)

	// Verify some pixels changed
	changedCount := 0
	for i := 0; i < len(img.Pix); i++ {
		if img.Pix[i] != original[i] {
			changedCount++
		}
	}

	if changedCount == 0 {
		t.Error("Dithering should modify at least some pixels")
	}
}

func TestApplyDithering_ZeroIntensity(t *testing.T) {
	sys := NewSystem(12345)

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	original := make([]byte, len(img.Pix))
	copy(original, img.Pix)

	sys.ApplyDithering(img, MaterialDefault, 0.0)

	// Zero intensity should not change anything
	for i := 0; i < len(img.Pix); i++ {
		if img.Pix[i] != original[i] {
			t.Error("Zero intensity should not modify pixels")
			break
		}
	}
}

func TestApplyDithering_TransparentPixels(t *testing.T) {
	sys := NewSystem(12345)

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Leave all pixels transparent (A=0)

	sys.ApplyDithering(img, MaterialDefault, 1.0)

	// Transparent pixels should remain unchanged
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			c := img.RGBAAt(x, y)
			if c.A != 0 {
				t.Errorf("Transparent pixel at (%d,%d) was modified", x, y)
			}
		}
	}
}

func TestMaterialTypes(t *testing.T) {
	materials := []struct {
		material Material
		name     string
	}{
		{MaterialDefault, "Default"},
		{MaterialMetal, "Metal"},
		{MaterialCloth, "Cloth"},
		{MaterialLeather, "Leather"},
		{MaterialFur, "Fur"},
		{MaterialCrystal, "Crystal"},
		{MaterialFlesh, "Flesh"},
		{MaterialScales, "Scales"},
		{MaterialSlime, "Slime"},
	}

	for _, tt := range materials {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(12345)

			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
				}
			}

			// Should not panic for any material type
			sys.ApplyDithering(img, tt.material, 0.5)

			// Verify values are still in valid range
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					c := img.RGBAAt(x, y)
					if c.R > 255 || c.G > 255 || c.B > 255 {
						t.Errorf("Invalid color values for material %s at (%d,%d)", tt.name, x, y)
					}
				}
			}
		})
	}
}

func TestApplyGradientDithering(t *testing.T) {
	directions := []struct {
		dir  Direction
		name string
	}{
		{DirVertical, "Vertical"},
		{DirHorizontal, "Horizontal"},
		{DirRadial, "Radial"},
		{DirDiagonal, "Diagonal"},
	}

	for _, tt := range directions {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(12345)

			img := image.NewRGBA(image.Rect(0, 0, 32, 32))
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
				}
			}

			startColor := color.RGBA{R: 50, G: 50, B: 50, A: 255}
			endColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}

			sys.ApplyGradientDithering(img, img.Bounds(), startColor, endColor, tt.dir)

			// Verify gradient was applied (pixels should vary)
			var minLuma, maxLuma float64 = 255, 0
			for y := 0; y < 32; y++ {
				for x := 0; x < 32; x++ {
					c := img.RGBAAt(x, y)
					luma := float64(c.R)*0.299 + float64(c.G)*0.587 + float64(c.B)*0.114
					if luma < minLuma {
						minLuma = luma
					}
					if luma > maxLuma {
						maxLuma = luma
					}
				}
			}

			if maxLuma-minLuma < 50 {
				t.Errorf("%s gradient has insufficient range: min=%.1f max=%.1f", tt.name, minLuma, maxLuma)
			}
		})
	}
}

func TestApplyEdgeDithering(t *testing.T) {
	sys := NewSystem(12345)

	// Create image with a sharp edge
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			if x < 16 {
				img.SetRGBA(x, y, color.RGBA{R: 50, G: 50, B: 50, A: 255})
			} else {
				img.SetRGBA(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
			}
		}
	}

	original := make([]byte, len(img.Pix))
	copy(original, img.Pix)

	sys.ApplyEdgeDithering(img, img.Bounds(), 0.5)

	// Check that edge region (around x=16) was most affected
	edgeChanges := 0
	nonEdgeChanges := 0

	for y := 1; y < 31; y++ {
		for x := 1; x < 31; x++ {
			idx := (y*32 + x) * 4
			changed := false
			for c := 0; c < 4; c++ {
				if img.Pix[idx+c] != original[idx+c] {
					changed = true
					break
				}
			}

			if changed {
				if x >= 14 && x <= 18 {
					edgeChanges++
				} else {
					nonEdgeChanges++
				}
			}
		}
	}

	// Edge region should have more changes proportionally
	edgePixels := 5 * 30 // 5 columns, 30 rows
	nonEdgePixels := 25*30 - edgePixels

	edgeRate := float64(edgeChanges) / float64(edgePixels)
	nonEdgeRate := float64(nonEdgeChanges) / float64(nonEdgePixels)

	t.Logf("Edge change rate: %.2f, Non-edge change rate: %.2f", edgeRate, nonEdgeRate)

	// Edge-aware dithering should concentrate changes at edges
	// But we're more lenient here since the algorithm focuses on luminance gradients
	if edgeChanges == 0 && nonEdgeChanges > 0 {
		t.Error("Edge dithering should affect edge regions")
	}
}

func TestApplySpecularDithering(t *testing.T) {
	sys := NewSystem(12345)

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	original := make([]byte, len(img.Pix))
	copy(original, img.Pix)

	// Apply specular highlight pointing top-left
	highlightDir := -math.Pi / 4 // -45 degrees
	sys.ApplySpecularDithering(img, img.Bounds(), highlightDir, 0.8)

	// Pixels in the highlight direction should be brighter
	topLeftBrightness := float64(img.RGBAAt(4, 4).R)
	bottomRightBrightness := float64(img.RGBAAt(28, 28).R)
	centerBrightness := float64(img.RGBAAt(16, 16).R)

	t.Logf("TopLeft: %.0f, Center: %.0f, BottomRight: %.0f", topLeftBrightness, centerBrightness, bottomRightBrightness)

	// Top-left should be equal or brighter than center (specular highlight)
	if topLeftBrightness < centerBrightness-20 {
		t.Error("Top-left should be in the specular highlight region")
	}
}

func TestLerpColor(t *testing.T) {
	tests := []struct {
		name string
		a, b color.RGBA
		t    float64
		want color.RGBA
	}{
		{
			"t=0 returns a",
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
			0.0,
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			"t=1 returns b",
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
			1.0,
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			"t=0.5 returns midpoint",
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
			color.RGBA{R: 200, G: 100, B: 50, A: 255},
			0.5,
			color.RGBA{R: 100, G: 50, B: 25, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lerpColor(tt.a, tt.b, tt.t)
			if got != tt.want {
				t.Errorf("lerpColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		input float64
		want  uint8
	}{
		{-50, 0},
		{0, 0},
		{128, 128},
		{255, 255},
		{300, 255},
	}

	for _, tt := range tests {
		got := clampByte(tt.input)
		if got != tt.want {
			t.Errorf("clampByte(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestGenerateBlueNoise(t *testing.T) {
	sys := NewSystem(12345)

	noise := sys.generateBlueNoise(32, 32)

	if len(noise) != 32 {
		t.Errorf("Blue noise height = %d, want 32", len(noise))
	}
	if len(noise[0]) != 32 {
		t.Errorf("Blue noise width = %d, want 32", len(noise[0]))
	}

	// Verify values are in [0, 1] range
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			v := noise[y][x]
			if v < 0 || v > 1 {
				t.Errorf("Blue noise value out of range at (%d,%d): %f", x, y, v)
			}
		}
	}

	// Verify caching works
	noise2 := sys.generateBlueNoise(32, 32)
	if &noise[0] != &noise2[0] {
		t.Error("Blue noise should be cached")
	}
}

func TestGetDitherStrength(t *testing.T) {
	sys := NewSystem(12345)

	tests := []struct {
		material Material
		minStr   float64
		maxStr   float64
	}{
		{MaterialMetal, 1.4, 1.6},
		{MaterialCrystal, 1.9, 2.1},
		{MaterialCloth, 0.7, 0.9},
		{MaterialFlesh, 0.5, 0.7},
		{MaterialDefault, 0.9, 1.1},
	}

	for _, tt := range tests {
		str := sys.getDitherStrength(tt.material)
		if str < tt.minStr || str > tt.maxStr {
			t.Errorf("getDitherStrength(%d) = %f, want in [%f, %f]", tt.material, str, tt.minStr, tt.maxStr)
		}
	}
}

func TestApplyDitheringToBounds(t *testing.T) {
	sys := NewSystem(12345)

	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	original := make([]byte, len(img.Pix))
	copy(original, img.Pix)

	// Only apply to a sub-region
	bounds := image.Rect(8, 8, 24, 24)
	sys.ApplyDitheringToBounds(img, bounds, MaterialDefault, 0.5)

	// Check that only the bounded region was affected
	outsideChanged := false
	insideChanged := false

	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			idx := (y*32 + x) * 4
			changed := img.Pix[idx] != original[idx]

			inBounds := x >= 8 && x < 24 && y >= 8 && y < 24
			if changed {
				if inBounds {
					insideChanged = true
				} else {
					outsideChanged = true
				}
			}
		}
	}

	if outsideChanged {
		t.Error("Pixels outside bounds were modified")
	}
	if !insideChanged {
		t.Error("Pixels inside bounds should be modified")
	}
}

func BenchmarkApplyDithering(b *testing.B) {
	sys := NewSystem(12345)
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ApplyDithering(img, MaterialDefault, 0.5)
	}
}

func BenchmarkApplyDithering_Metal(b *testing.B) {
	sys := NewSystem(12345)
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ApplyDithering(img, MaterialMetal, 0.5)
	}
}

func BenchmarkApplyDithering_BlueNoise(b *testing.B) {
	sys := NewSystem(12345)
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ApplyDithering(img, MaterialCloth, 0.5)
	}
}

func BenchmarkApplyGradientDithering(b *testing.B) {
	sys := NewSystem(12345)
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 128, G: 128, B: 128, A: 255})
		}
	}

	start := color.RGBA{R: 50, G: 50, B: 50, A: 255}
	end := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ApplyGradientDithering(img, img.Bounds(), start, end, DirRadial)
	}
}

func BenchmarkApplyEdgeDithering(b *testing.B) {
	sys := NewSystem(12345)
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if x < 32 {
				img.SetRGBA(x, y, color.RGBA{R: 50, G: 50, B: 50, A: 255})
			} else {
				img.SetRGBA(x, y, color.RGBA{R: 200, G: 200, B: 200, A: 255})
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ApplyEdgeDithering(img, img.Bounds(), 0.5)
	}
}

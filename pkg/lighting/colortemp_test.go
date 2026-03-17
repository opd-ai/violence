package lighting

import (
	"image/color"
	"math"
	"testing"
)

func TestDefaultColorTempConfig(t *testing.T) {
	cfg := DefaultColorTempConfig()

	if !cfg.Enabled {
		t.Error("Default config should be enabled")
	}
	if cfg.MaxTintStrength <= 0 || cfg.MaxTintStrength > 1.0 {
		t.Errorf("MaxTintStrength should be in (0,1], got %f", cfg.MaxTintStrength)
	}
	if cfg.FalloffExponent <= 0 {
		t.Errorf("FalloffExponent should be positive, got %f", cfg.FalloffExponent)
	}
}

func TestNewColorTempSystem(t *testing.T) {
	cfg := DefaultColorTempConfig()
	sys := NewColorTempSystem(cfg)

	if sys == nil {
		t.Fatal("NewColorTempSystem returned nil")
	}
	if len(sys.lights) != 0 {
		t.Errorf("New system should have no lights, got %d", len(sys.lights))
	}
}

func TestColorTempSystem_AddLight(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	light := ColorTempLight{
		X:           10.0,
		Y:           20.0,
		Radius:      5.0,
		Intensity:   0.8,
		Temperature: TempWarmTorch,
		R:           1.0,
		G:           0.6,
		B:           0.2,
	}

	sys.AddLight(light)

	if len(sys.lights) != 1 {
		t.Errorf("Expected 1 light, got %d", len(sys.lights))
	}
	if sys.lights[0].X != 10.0 {
		t.Errorf("Light X should be 10.0, got %f", sys.lights[0].X)
	}
}

func TestColorTempSystem_ClearLights(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	sys.AddLight(ColorTempLight{X: 1, Y: 1, Radius: 5, Intensity: 1})
	sys.AddLight(ColorTempLight{X: 2, Y: 2, Radius: 5, Intensity: 1})

	if len(sys.lights) != 2 {
		t.Fatalf("Expected 2 lights before clear, got %d", len(sys.lights))
	}

	sys.ClearLights()

	if len(sys.lights) != 0 {
		t.Errorf("Expected 0 lights after clear, got %d", len(sys.lights))
	}
}

func TestInferTemperatureFromPreset(t *testing.T) {
	tests := []struct {
		name     string
		preset   LightPreset
		expected ColorTemperature
	}{
		{
			name:     "warm torch",
			preset:   LightPreset{Name: "torch", R: 1.0, G: 0.6, B: 0.2},
			expected: ColorTemperature(0.8), // 1.0 - 0.2 = 0.8
		},
		{
			name:     "cool blue",
			preset:   LightPreset{Name: "monitor", R: 0.3, G: 0.6, B: 1.0},
			expected: ColorTemperature(-0.7), // 0.3 - 1.0 = -0.7
		},
		{
			name:     "neutral white",
			preset:   LightPreset{Name: "generic", R: 1.0, G: 1.0, B: 1.0},
			expected: ColorTemperature(0.0), // 1.0 - 1.0 = 0.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferTemperatureFromPreset(tt.preset)
			if math.Abs(float64(got-tt.expected)) > 0.01 {
				t.Errorf("InferTemperatureFromPreset() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestInferTemperatureFromName(t *testing.T) {
	tests := []struct {
		lightType string
		minTemp   ColorTemperature
		maxTemp   ColorTemperature
	}{
		{"torch", TempWarmTorch - 0.1, TempWarmTorch + 0.1},
		{"magic_crystal", TempCoolMagic - 0.1, TempCoolMagic + 0.1},
		{"monitor", TempCoolMonitor - 0.1, TempCoolMonitor + 0.1},
		{"unknown_type", TempNeutral - 0.1, TempNeutral + 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.lightType, func(t *testing.T) {
			got := InferTemperatureFromName(tt.lightType)
			if got < tt.minTemp || got > tt.maxTemp {
				t.Errorf("InferTemperatureFromName(%s) = %v, want in range [%v, %v]",
					tt.lightType, got, tt.minTemp, tt.maxTemp)
			}
		})
	}
}

func TestColorTempSystem_CalculateTintAtPosition(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	// Add a warm torch light at origin
	sys.AddLight(ColorTempLight{
		X:           0,
		Y:           0,
		Radius:      10.0,
		Intensity:   1.0,
		Temperature: TempWarmTorch,
		R:           1.0,
		G:           0.6,
		B:           0.2,
	})

	// Position at the light center should have maximum tint
	tintAtCenter := sys.CalculateTintAtPosition(0, 0)
	if tintAtCenter.A == 0 {
		t.Error("Tint at light center should have non-zero alpha")
	}

	// Warm torch should have more red than blue
	if tintAtCenter.R <= tintAtCenter.B {
		t.Errorf("Warm light tint should have R > B, got R=%d, B=%d", tintAtCenter.R, tintAtCenter.B)
	}

	// Position far outside radius should have no tint
	tintOutside := sys.CalculateTintAtPosition(100, 100)
	if tintOutside.A != 0 {
		t.Errorf("Tint outside light radius should have alpha=0, got %d", tintOutside.A)
	}
}

func TestColorTempSystem_DisabledReturnsNoTint(t *testing.T) {
	cfg := DefaultColorTempConfig()
	cfg.Enabled = false
	sys := NewColorTempSystem(cfg)

	sys.AddLight(ColorTempLight{
		X:         0,
		Y:         0,
		Radius:    10.0,
		Intensity: 1.0,
	})

	tint := sys.CalculateTintAtPosition(0, 0)
	if tint.A != 0 {
		t.Errorf("Disabled system should return alpha=0, got %d", tint.A)
	}
}

func TestColorTempSystem_NoLightsReturnsNoTint(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	tint := sys.CalculateTintAtPosition(5, 5)
	if tint.A != 0 {
		t.Errorf("System with no lights should return alpha=0, got %d", tint.A)
	}
}

func TestApplyTintToColor(t *testing.T) {
	tests := []struct {
		name    string
		surface color.RGBA
		tint    color.RGBA
	}{
		{
			name:    "no tint (alpha=0)",
			surface: color.RGBA{R: 200, G: 200, B: 200, A: 255},
			tint:    color.RGBA{R: 255, G: 0, B: 0, A: 0},
		},
		{
			name:    "warm tint",
			surface: color.RGBA{R: 200, G: 200, B: 200, A: 255},
			tint:    color.RGBA{R: 255, G: 200, B: 100, A: 128},
		},
		{
			name:    "cool tint",
			surface: color.RGBA{R: 200, G: 200, B: 200, A: 255},
			tint:    color.RGBA{R: 100, G: 150, B: 255, A: 128},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyTintToColor(tt.surface, tt.tint)

			// Result should preserve surface alpha
			if result.A != tt.surface.A {
				t.Errorf("ApplyTintToColor should preserve alpha, got %d want %d", result.A, tt.surface.A)
			}

			// If tint alpha is 0, result should equal surface
			if tt.tint.A == 0 {
				if result != tt.surface {
					t.Errorf("With zero tint alpha, result should equal surface")
				}
			}
		})
	}
}

func TestColorTempSystem_TintFalloffWithDistance(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	sys.AddLight(ColorTempLight{
		X:           0,
		Y:           0,
		Radius:      10.0,
		Intensity:   1.0,
		Temperature: TempWarmTorch,
		R:           1.0,
		G:           0.6,
		B:           0.2,
	})

	// Tint should decrease with distance
	tintClose := sys.CalculateTintAtPosition(1, 0)
	tintMid := sys.CalculateTintAtPosition(5, 0)
	tintFar := sys.CalculateTintAtPosition(9, 0)

	if tintClose.A <= tintMid.A {
		t.Errorf("Closer position should have higher tint alpha: close=%d, mid=%d", tintClose.A, tintMid.A)
	}
	if tintMid.A <= tintFar.A {
		t.Errorf("Mid position should have higher tint alpha than far: mid=%d, far=%d", tintMid.A, tintFar.A)
	}
}

func TestColorTempSystem_MultipleLightsBlend(t *testing.T) {
	cfg := DefaultColorTempConfig()
	cfg.BlendMode = "additive"
	sys := NewColorTempSystem(cfg)

	// Add a warm light on the left
	sys.AddLight(ColorTempLight{
		X:           -5,
		Y:           0,
		Radius:      10.0,
		Intensity:   0.8,
		Temperature: TempWarmTorch,
		R:           1.0,
		G:           0.6,
		B:           0.2,
	})

	// Add a cool light on the right
	sys.AddLight(ColorTempLight{
		X:           5,
		Y:           0,
		Radius:      10.0,
		Intensity:   0.8,
		Temperature: TempCoolMagic,
		R:           0.4,
		G:           0.6,
		B:           1.0,
	})

	// Position between both lights should have contributions from both
	tintMiddle := sys.CalculateTintAtPosition(0, 0)
	if tintMiddle.A == 0 {
		t.Error("Middle position should have tint from both lights")
	}

	// Position near warm light should be warmer (more red)
	tintNearWarm := sys.CalculateTintAtPosition(-3, 0)
	// Position near cool light should be cooler (more blue)
	tintNearCool := sys.CalculateTintAtPosition(3, 0)

	// Warm position should have relatively more red
	warmRatio := float64(tintNearWarm.R) / float64(tintNearWarm.B+1)
	coolRatio := float64(tintNearCool.R) / float64(tintNearCool.B+1)

	if warmRatio <= coolRatio {
		t.Errorf("Position near warm light should have higher R/B ratio: warm=%f, cool=%f", warmRatio, coolRatio)
	}
}

func TestGetTemperatureForGenreLight(t *testing.T) {
	tests := []struct {
		genreID   string
		lightType string
		expected  ColorTemperature
	}{
		{"fantasy", "torch", TempWarmTorch},
		{"scifi", "monitor", TempCoolMonitor},
		{"fantasy", "unknown", TempWarmTorch}, // Genre default
		{"cyberpunk", "unknown", TempCoolMagic},
	}

	for _, tt := range tests {
		t.Run(tt.genreID+"_"+tt.lightType, func(t *testing.T) {
			got := GetTemperatureForGenreLight(tt.genreID, tt.lightType)
			if got != tt.expected {
				t.Errorf("GetTemperatureForGenreLight(%s, %s) = %v, want %v",
					tt.genreID, tt.lightType, got, tt.expected)
			}
		})
	}
}

func TestColorTempSystem_temperatureToRGB(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	// Warm temperature should have R > B
	warmR, warmG, warmB := sys.temperatureToRGB(TempWarmTorch)
	if warmR <= warmB {
		t.Errorf("Warm temperature should have R > B, got R=%f, B=%f", warmR, warmB)
	}
	if warmG >= warmR {
		t.Errorf("Warm temperature should have R >= G, got R=%f, G=%f", warmR, warmG)
	}

	// Cool temperature should have B > R
	coolR, _, coolB := sys.temperatureToRGB(TempCoolMagic)
	if coolB <= coolR {
		t.Errorf("Cool temperature should have B > R, got R=%f, B=%f", coolR, coolB)
	}

	// Neutral should have equal RGB
	neutR, neutG, neutB := sys.temperatureToRGB(TempNeutral)
	if neutR != 1.0 || neutG != 1.0 || neutB != 1.0 {
		t.Errorf("Neutral temperature should have R=G=B=1.0, got R=%f, G=%f, B=%f", neutR, neutG, neutB)
	}
}

func TestColorTempSystem_ApplyTintToImage(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	sys.AddLight(ColorTempLight{
		X:           2.0,
		Y:           2.0,
		Radius:      10.0,
		Intensity:   1.0,
		Temperature: TempWarmTorch,
		R:           1.0,
		G:           0.6,
		B:           0.2,
	})

	// Create a 4x4 gray image (RGBA)
	width, height := 4, 4
	pixels := make([]byte, width*height*4)
	for i := 0; i < len(pixels); i += 4 {
		pixels[i] = 200   // R
		pixels[i+1] = 200 // G
		pixels[i+2] = 200 // B
		pixels[i+3] = 255 // A
	}

	// Apply tint (world coords match pixel coords with scale=1)
	sys.ApplyTintToImage(pixels, width, height, 0, 0, 1.0)

	// Check that some pixels were modified
	modified := false
	for i := 0; i < len(pixels); i += 4 {
		if pixels[i] != 200 || pixels[i+1] != 200 || pixels[i+2] != 200 {
			modified = true
			break
		}
	}

	if !modified {
		t.Error("ApplyTintToImage should modify at least some pixels near the light")
	}
}

func TestColorTempSystem_SetConfig(t *testing.T) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	newCfg := ColorTempConfig{
		Enabled:         false,
		MaxTintStrength: 0.5,
		FalloffExponent: 1.5,
		BlendMode:       "average",
	}

	sys.SetConfig(newCfg)

	// Verify config was updated by checking behavior
	sys.AddLight(ColorTempLight{X: 0, Y: 0, Radius: 10, Intensity: 1})

	tint := sys.CalculateTintAtPosition(0, 0)
	if tint.A != 0 {
		t.Error("After disabling, system should return no tint")
	}
}

func BenchmarkCalculateTintAtPosition(b *testing.B) {
	sys := NewColorTempSystem(DefaultColorTempConfig())

	// Add several lights
	for i := 0; i < 10; i++ {
		sys.AddLight(ColorTempLight{
			X:           float64(i * 10),
			Y:           float64(i * 10),
			Radius:      15.0,
			Intensity:   0.8,
			Temperature: TempWarmTorch,
			R:           1.0,
			G:           0.6,
			B:           0.2,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.CalculateTintAtPosition(50, 50)
	}
}

func BenchmarkApplyTintToColor(b *testing.B) {
	surface := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	tint := color.RGBA{R: 255, G: 200, B: 100, A: 128}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyTintToColor(surface, tint)
	}
}

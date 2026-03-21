package bouncelight

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	comp := NewComponent()

	if comp == nil {
		t.Fatal("NewComponent returned nil")
	}

	if comp.Type() != "BounceLight" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "BounceLight")
	}

	if !comp.Enabled {
		t.Error("Component should be enabled by default")
	}

	if comp.Intensity != 0 {
		t.Errorf("Initial intensity = %f, want 0", comp.Intensity)
	}
}

func TestComponentAddContribution(t *testing.T) {
	tests := []struct {
		name      string
		r, g, b   float64
		intensity float64
		wantR     float64
		wantG     float64
		wantB     float64
	}{
		{
			name:      "red contribution",
			r:         1.0,
			g:         0.0,
			b:         0.0,
			intensity: 0.5,
			wantR:     0.5,
			wantG:     0.0,
			wantB:     0.0,
		},
		{
			name:      "green contribution",
			r:         0.0,
			g:         1.0,
			b:         0.0,
			intensity: 0.3,
			wantR:     0.0,
			wantG:     0.3,
			wantB:     0.0,
		},
		{
			name:      "white contribution",
			r:         1.0,
			g:         1.0,
			b:         1.0,
			intensity: 0.2,
			wantR:     0.2,
			wantG:     0.2,
			wantB:     0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent()
			comp.AddContribution(tt.r, tt.g, tt.b, tt.intensity)

			if comp.TintR != tt.wantR {
				t.Errorf("TintR = %f, want %f", comp.TintR, tt.wantR)
			}
			if comp.TintG != tt.wantG {
				t.Errorf("TintG = %f, want %f", comp.TintG, tt.wantG)
			}
			if comp.TintB != tt.wantB {
				t.Errorf("TintB = %f, want %f", comp.TintB, tt.wantB)
			}
			if comp.Intensity != tt.intensity {
				t.Errorf("Intensity = %f, want %f", comp.Intensity, tt.intensity)
			}
		})
	}
}

func TestComponentNormalize(t *testing.T) {
	comp := NewComponent()

	// Add multiple contributions
	comp.AddContribution(1.0, 0.0, 0.0, 0.5) // Red
	comp.AddContribution(0.0, 1.0, 0.0, 0.5) // Green

	comp.Normalize()

	// Should average to yellow-ish
	if comp.TintR < 0.4 || comp.TintR > 0.6 {
		t.Errorf("Normalized TintR = %f, expected ~0.5", comp.TintR)
	}
	if comp.TintG < 0.4 || comp.TintG > 0.6 {
		t.Errorf("Normalized TintG = %f, expected ~0.5", comp.TintG)
	}
	if comp.TintB > 0.1 {
		t.Errorf("Normalized TintB = %f, expected ~0", comp.TintB)
	}

	// Intensity should be clamped
	if comp.Intensity > 1.0 {
		t.Errorf("Intensity = %f, should be clamped to 1.0", comp.Intensity)
	}
}

func TestComponentReset(t *testing.T) {
	comp := NewComponent()
	comp.AddContribution(1.0, 1.0, 1.0, 1.0)
	comp.Normalize()

	comp.Reset()

	if comp.TintR != 0 || comp.TintG != 0 || comp.TintB != 0 {
		t.Error("Reset should clear tint values")
	}
	if comp.Intensity != 0 {
		t.Error("Reset should clear intensity")
	}
	if comp.ContributorCount != 0 {
		t.Error("Reset should clear contributor count")
	}
}

func TestComponentGetTint(t *testing.T) {
	comp := NewComponent()
	comp.TintR = 1.0
	comp.TintG = 0.5
	comp.TintB = 0.25
	comp.Intensity = 0.8

	tint := comp.GetTint(1.0)

	// Check that color is properly scaled by intensity
	expectedR := uint8(1.0 * 255 * 0.8)
	expectedG := uint8(0.5 * 255 * 0.8)
	expectedB := uint8(0.25 * 255 * 0.8)

	if tint.R != expectedR {
		t.Errorf("Tint R = %d, want %d", tint.R, expectedR)
	}
	if tint.G != expectedG {
		t.Errorf("Tint G = %d, want %d", tint.G, expectedG)
	}
	if tint.B != expectedB {
		t.Errorf("Tint B = %d, want %d", tint.B, expectedB)
	}
}

func TestBounceSurfaceContribution(t *testing.T) {
	tests := []struct {
		name           string
		surface        BounceSurface
		targetX        float64
		targetY        float64
		maxDist        float64
		wantZeroIntens bool
	}{
		{
			name: "close target",
			surface: BounceSurface{
				X: 0, Y: 0, R: 1.0, G: 0.0, B: 0.0,
				Reflectivity: 1.0, DirectLight: 1.0,
			},
			targetX:        1.0,
			targetY:        0.0,
			maxDist:        10.0,
			wantZeroIntens: false,
		},
		{
			name: "far target",
			surface: BounceSurface{
				X: 0, Y: 0, R: 1.0, G: 0.0, B: 0.0,
				Reflectivity: 1.0, DirectLight: 1.0,
			},
			targetX:        100.0,
			targetY:        100.0,
			maxDist:        10.0,
			wantZeroIntens: true, // Beyond max distance
		},
		{
			name: "same position",
			surface: BounceSurface{
				X: 5, Y: 5, R: 0.0, G: 1.0, B: 0.0,
				Reflectivity: 1.0, DirectLight: 1.0,
			},
			targetX:        5.0,
			targetY:        5.0,
			maxDist:        10.0,
			wantZeroIntens: false, // Should have high intensity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, g, b, intensity := tt.surface.BounceContribution(tt.targetX, tt.targetY, tt.maxDist)

			if tt.wantZeroIntens && intensity > 0 {
				t.Errorf("Expected zero intensity for far target, got %f", intensity)
			}

			if !tt.wantZeroIntens && intensity <= 0 {
				t.Errorf("Expected positive intensity for close target, got %f", intensity)
			}

			// Color should match surface when intensity > 0
			if intensity > 0 {
				if r != tt.surface.R || g != tt.surface.G || b != tt.surface.B {
					t.Errorf("Returned color (%f,%f,%f) doesn't match surface (%f,%f,%f)",
						r, g, b, tt.surface.R, tt.surface.G, tt.surface.B)
				}
			}
		})
	}
}

func TestNewBounceSurface(t *testing.T) {
	surf := NewBounceSurface(10.0, 20.0, 0.8, 0.5, 0.2, true)

	if surf.X != 10.0 || surf.Y != 20.0 {
		t.Errorf("Position = (%f,%f), want (10,20)", surf.X, surf.Y)
	}

	if surf.R != 0.8 || surf.G != 0.5 || surf.B != 0.2 {
		t.Errorf("Color = (%f,%f,%f), want (0.8,0.5,0.2)", surf.R, surf.G, surf.B)
	}

	if !surf.IsWall {
		t.Error("IsWall should be true")
	}

	if surf.Reflectivity != 0.5 {
		t.Errorf("Reflectivity = %f, want 0.5", surf.Reflectivity)
	}

	if surf.DirectLight != 1.0 {
		t.Errorf("DirectLight = %f, want 1.0", surf.DirectLight)
	}
}

func TestDisabledComponent(t *testing.T) {
	comp := NewComponent()
	comp.Enabled = false

	comp.AddContribution(1.0, 1.0, 1.0, 1.0)

	if comp.TintR != 0 || comp.TintG != 0 || comp.TintB != 0 {
		t.Error("Disabled component should not accept contributions")
	}

	if comp.Intensity != 0 {
		t.Error("Disabled component should have zero intensity")
	}
}

func TestZeroIntensityContribution(t *testing.T) {
	comp := NewComponent()

	comp.AddContribution(1.0, 1.0, 1.0, 0.0)

	if comp.ContributorCount != 0 {
		t.Error("Zero intensity should not count as contribution")
	}
}

func TestComponentClamp(t *testing.T) {
	comp := NewComponent()

	// Add excessive contributions that should exceed 1.0
	for i := 0; i < 10; i++ {
		comp.AddContribution(1.0, 1.0, 1.0, 0.5)
	}

	comp.Normalize()

	if comp.TintR > 1.0 || comp.TintG > 1.0 || comp.TintB > 1.0 {
		t.Error("Colors should be clamped to 1.0")
	}

	if comp.Intensity > 1.0 {
		t.Error("Intensity should be clamped to 1.0")
	}
}

func TestComponentTintWithStrength(t *testing.T) {
	comp := NewComponent()
	comp.TintR = 1.0
	comp.TintG = 1.0
	comp.TintB = 1.0
	comp.Intensity = 1.0

	// Full strength
	tint := comp.GetTint(1.0)
	if tint.R != 255 || tint.G != 255 || tint.B != 255 {
		t.Errorf("Full strength tint = (%d,%d,%d), want (255,255,255)",
			tint.R, tint.G, tint.B)
	}

	// Half strength
	tint = comp.GetTint(0.5)
	if tint.R < 125 || tint.R > 130 { // Allow for rounding
		t.Errorf("Half strength R = %d, want ~127", tint.R)
	}

	// Zero strength
	tint = comp.GetTint(0.0)
	if tint.R != 0 || tint.G != 0 || tint.B != 0 || tint.A != 0 {
		t.Error("Zero strength should produce zero tint")
	}
}

func TestContributionFalloff(t *testing.T) {
	surf := BounceSurface{
		X: 0, Y: 0, R: 1.0, G: 1.0, B: 1.0,
		Reflectivity: 1.0, DirectLight: 1.0,
	}

	maxDist := 10.0

	// Close point should have higher intensity than far point
	_, _, _, closeIntensity := surf.BounceContribution(1.0, 0.0, maxDist)
	_, _, _, farIntensity := surf.BounceContribution(5.0, 0.0, maxDist)

	if closeIntensity <= farIntensity {
		t.Errorf("Close intensity (%f) should be > far intensity (%f)",
			closeIntensity, farIntensity)
	}
}

func TestNormalizeEmptyComponent(t *testing.T) {
	comp := NewComponent()
	// Normalize with no contributions should not panic
	comp.Normalize()

	if comp.TintR != 0 || comp.TintG != 0 || comp.TintB != 0 {
		t.Error("Normalizing empty component should leave zeros")
	}
}

func TestComponentTypeString(t *testing.T) {
	comp := NewComponent()

	typeStr := comp.Type()
	if typeStr == "" {
		t.Error("Type() should return non-empty string")
	}
}

func TestTintAlphaChannel(t *testing.T) {
	comp := NewComponent()
	comp.TintR = 1.0
	comp.TintG = 1.0
	comp.TintB = 1.0
	comp.Intensity = 0.5

	tint := comp.GetTint(1.0)

	// Alpha should reflect intensity
	expectedAlpha := uint8(127) // 0.5 * 255 rounded
	if tint.A != expectedAlpha && tint.A != 128 {
		t.Errorf("Tint alpha = %d, want %d", tint.A, expectedAlpha)
	}
}

func TestBounceSurfaceNoDirectLight(t *testing.T) {
	surf := BounceSurface{
		X: 0, Y: 0, R: 1.0, G: 1.0, B: 1.0,
		Reflectivity: 1.0, DirectLight: 0.0, // No direct light
	}

	_, _, _, intensity := surf.BounceContribution(1.0, 0.0, 10.0)

	if intensity > 0 {
		t.Error("Surface with no direct light should not contribute")
	}
}

func TestBounceSurfaceZeroReflectivity(t *testing.T) {
	surf := BounceSurface{
		X: 0, Y: 0, R: 1.0, G: 1.0, B: 1.0,
		Reflectivity: 0.0, DirectLight: 1.0,
	}

	_, _, _, intensity := surf.BounceContribution(1.0, 0.0, 10.0)

	if intensity > 0 {
		t.Error("Surface with zero reflectivity should not contribute")
	}
}

func TestGetTintOverflow(t *testing.T) {
	comp := NewComponent()
	comp.TintR = 2.0 // Exceeds 1.0
	comp.TintG = 2.0
	comp.TintB = 2.0
	comp.Intensity = 2.0 // Exceeds 1.0

	tint := comp.GetTint(1.0)

	// GetTint should clamp to 255
	if tint.R > 255 || tint.G > 255 || tint.B > 255 || tint.A > 255 {
		t.Error("GetTint should clamp values to byte range")
	}
}

func BenchmarkAddContribution(b *testing.B) {
	comp := NewComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.AddContribution(0.8, 0.5, 0.3, 0.1)
	}
}

func BenchmarkBounceContribution(b *testing.B) {
	surf := BounceSurface{
		X: 0, Y: 0, R: 1.0, G: 0.5, B: 0.2,
		Reflectivity: 0.6, DirectLight: 0.8,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		surf.BounceContribution(5.0, 3.0, 10.0)
	}
}

func BenchmarkGetTint(b *testing.B) {
	comp := NewComponent()
	comp.TintR = 0.8
	comp.TintG = 0.5
	comp.TintB = 0.3
	comp.Intensity = 0.6

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.GetTint(1.0)
	}
}

func BenchmarkNormalize(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp := NewComponent()
		for j := 0; j < 10; j++ {
			comp.AddContribution(float64(j)*0.1, 0.5, 0.3, 0.1)
		}
		comp.Normalize()
	}
}

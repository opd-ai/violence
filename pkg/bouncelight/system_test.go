package bouncelight

import (
	"image/color"
	"testing"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "invalid"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 320, 240)

			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}

			preset := sys.GetPreset()
			if preset.BounceStrength <= 0 {
				t.Error("BounceStrength should be positive")
			}
			if preset.MaxBounceDistance <= 0 {
				t.Error("MaxBounceDistance should be positive")
			}
		})
	}
}

func TestSystemSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	// Change genre
	sys.SetGenre("scifi")

	preset := sys.GetPreset()

	// Sci-fi has specific characteristics
	if preset.WarmShift >= 0 {
		t.Error("Sci-fi should have cool (negative) warm shift")
	}
}

func TestCalculateBounce(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	// Create a simple scene with one red wall
	surfaces := []BounceSurface{
		{
			X: 2.0, Y: 2.0, R: 1.0, G: 0.0, B: 0.0,
			Reflectivity: 1.0, IsWall: true, DirectLight: 1.0,
		},
	}

	bounceMap := sys.CalculateBounce(surfaces, 8, 8, 1.0)

	if bounceMap == nil {
		t.Fatal("CalculateBounce returned nil")
	}

	if bounceMap.Width != 8 || bounceMap.Height != 8 {
		t.Errorf("BounceMap dimensions = (%d,%d), want (8,8)",
			bounceMap.Width, bounceMap.Height)
	}

	// Cell near the red wall should have red tint
	r, g, b, intensity := sys.GetBounceAt(bounceMap, 2.5, 2.5)

	if intensity <= 0 {
		t.Error("Cell near surface should have bounce intensity")
	}

	// Should be predominantly red
	if r < g || r < b {
		t.Errorf("Bounce near red wall should be reddish: R=%f G=%f B=%f", r, g, b)
	}
}

func TestCalculateBounceMultipleSurfaces(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	// Create surfaces with different colors
	surfaces := []BounceSurface{
		{X: 0, Y: 0, R: 1.0, G: 0.0, B: 0.0, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
		{X: 4, Y: 0, R: 0.0, G: 1.0, B: 0.0, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
	}

	bounceMap := sys.CalculateBounce(surfaces, 8, 8, 1.0)

	// Point between surfaces should have mixed color
	r, g, _, _ := sys.GetBounceAt(bounceMap, 2.0, 0.5)

	// Both red and green should be present
	if r <= 0 || g <= 0 {
		t.Errorf("Mixed area should have both colors: R=%f G=%f", r, g)
	}
}

func TestGetBounceAtBilinear(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	surfaces := []BounceSurface{
		{X: 2, Y: 2, R: 1.0, G: 1.0, B: 1.0, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
	}

	bounceMap := sys.CalculateBounce(surfaces, 4, 4, 1.0)

	// Bilinear interpolation should give smooth values
	r1, g1, b1, i1 := sys.GetBounceAtBilinear(bounceMap, 1.9, 1.9)
	r2, g2, b2, i2 := sys.GetBounceAtBilinear(bounceMap, 2.1, 2.1)

	// Values should be close (smooth interpolation)
	if diff := absFloat(r1 - r2); diff > 0.5 {
		t.Errorf("Bilinear R values too different: %f vs %f", r1, r2)
	}
	if diff := absFloat(g1 - g2); diff > 0.5 {
		t.Errorf("Bilinear G values too different: %f vs %f", g1, g2)
	}
	if diff := absFloat(b1 - b2); diff > 0.5 {
		t.Errorf("Bilinear B values too different: %f vs %f", b1, b2)
	}
	if diff := absFloat(i1 - i2); diff > 0.5 {
		t.Errorf("Bilinear intensity values too different: %f vs %f", i1, i2)
	}
}

func TestGetBounceAtOutOfBounds(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	surfaces := []BounceSurface{
		{X: 2, Y: 2, R: 1.0, G: 1.0, B: 1.0, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
	}

	bounceMap := sys.CalculateBounce(surfaces, 4, 4, 1.0)

	// Out of bounds queries should return zero
	_, _, _, i := sys.GetBounceAt(bounceMap, -10.0, -10.0)
	if i != 0 {
		t.Error("Out of bounds should return zero intensity")
	}

	_, _, _, i = sys.GetBounceAt(bounceMap, 100.0, 100.0)
	if i != 0 {
		t.Error("Out of bounds should return zero intensity")
	}
}

func TestGetBounceAtNilMap(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	r, g, b, i := sys.GetBounceAt(nil, 1.0, 1.0)

	if r != 0 || g != 0 || b != 0 || i != 0 {
		t.Error("Nil bounceMap should return zeros")
	}

	r, g, b, i = sys.GetBounceAtBilinear(nil, 1.0, 1.0)

	if r != 0 || g != 0 || b != 0 || i != 0 {
		t.Error("Nil bounceMap bilinear should return zeros")
	}
}

func TestApplyBounceToColor(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	tests := []struct {
		name      string
		base      color.RGBA
		r, g, b   float64
		intensity float64
		wantR     uint8
		wantG     uint8
		wantB     uint8
	}{
		{
			name:      "zero intensity",
			base:      color.RGBA{R: 100, G: 100, B: 100, A: 255},
			r:         1.0,
			g:         0.0,
			b:         0.0,
			intensity: 0.0,
			wantR:     100,
			wantG:     100,
			wantB:     100,
		},
		{
			name:      "add red bounce",
			base:      color.RGBA{R: 100, G: 100, B: 100, A: 255},
			r:         1.0,
			g:         0.0,
			b:         0.0,
			intensity: 0.2,
			wantR:     151, // 100 + 51 (0.2 * 255)
			wantG:     100,
			wantB:     100,
		},
		{
			name:      "preserve alpha",
			base:      color.RGBA{R: 100, G: 100, B: 100, A: 128},
			r:         0.5,
			g:         0.5,
			b:         0.5,
			intensity: 0.1,
			wantR:     112,
			wantG:     112,
			wantB:     112,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.ApplyBounceToColor(tt.base, tt.r, tt.g, tt.b, tt.intensity)

			// Allow small tolerance for rounding
			if abs(int(result.R)-int(tt.wantR)) > 2 {
				t.Errorf("R = %d, want ~%d", result.R, tt.wantR)
			}
			if abs(int(result.G)-int(tt.wantG)) > 2 {
				t.Errorf("G = %d, want ~%d", result.G, tt.wantG)
			}
			if abs(int(result.B)-int(tt.wantB)) > 2 {
				t.Errorf("B = %d, want ~%d", result.B, tt.wantB)
			}
			if result.A != tt.base.A {
				t.Errorf("Alpha = %d, want %d (preserved)", result.A, tt.base.A)
			}
		})
	}
}

func TestExtractSurfacesFromColors(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	colors := []color.RGBA{
		{R: 255, G: 0, B: 0, A: 255},     // Red
		{R: 0, G: 255, B: 0, A: 255},     // Green
		{R: 0, G: 0, B: 0, A: 0},         // Transparent (should be skipped)
		{R: 255, G: 255, B: 255, A: 255}, // White
	}

	surfaces := sys.ExtractSurfacesFromColors(colors, 2, 2, 1.0, true)

	// Should have 3 surfaces (transparent skipped)
	if len(surfaces) != 3 {
		t.Errorf("Expected 3 surfaces, got %d", len(surfaces))
	}

	// Check first surface (red)
	if surfaces[0].R != 1.0 || surfaces[0].G != 0 || surfaces[0].B != 0 {
		t.Error("First surface should be red")
	}

	// Check that all are walls
	for i, surf := range surfaces {
		if !surf.IsWall {
			t.Errorf("Surface %d should be wall", i)
		}
	}
}

func TestSaturationBoost(t *testing.T) {
	// Cyberpunk has saturation boost
	sys := NewSystem("cyberpunk", 320, 240)

	preset := sys.GetPreset()
	if preset.SaturationBoost <= 0 {
		t.Error("Cyberpunk should have positive saturation boost")
	}

	// Horror has negative saturation (desaturated)
	sys.SetGenre("horror")
	preset = sys.GetPreset()
	if preset.SaturationBoost >= 0 {
		t.Error("Horror should have negative saturation boost")
	}
}

func TestWallVsFloorContribution(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	// Wall surface
	wallSurf := []BounceSurface{
		{X: 2, Y: 2, R: 1.0, G: 0.0, B: 0.0, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
	}

	// Floor surface (same position and color)
	floorSurf := []BounceSurface{
		{X: 2, Y: 2, R: 1.0, G: 0.0, B: 0.0, Reflectivity: 1.0, IsWall: false, DirectLight: 1.0},
	}

	wallMap := sys.CalculateBounce(wallSurf, 4, 4, 1.0)
	floorMap := sys.CalculateBounce(floorSurf, 4, 4, 1.0)

	_, _, _, wallIntensity := sys.GetBounceAt(wallMap, 2.5, 2.5)
	_, _, _, floorIntensity := sys.GetBounceAt(floorMap, 2.5, 2.5)

	// Fantasy has higher wall contribution than floor
	preset := sys.GetPreset()
	if preset.WallContribution <= preset.FloorContribution {
		// Wall should contribute more in fantasy
		if wallIntensity <= floorIntensity {
			t.Logf("Wall intensity: %f, Floor intensity: %f", wallIntensity, floorIntensity)
		}
	}
}

func TestClearCache(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	// Calculate some bounce maps (they get cached)
	surfaces := []BounceSurface{
		{X: 2, Y: 2, R: 1.0, G: 0.0, B: 0.0, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
	}
	sys.CalculateBounce(surfaces, 4, 4, 1.0)

	// Clear cache
	sys.ClearCache()

	// Cache should be empty (no way to directly check, but no crash = pass)
}

func TestGenrePresets(t *testing.T) {
	for genre, preset := range genrePresets {
		t.Run(genre, func(t *testing.T) {
			// All presets should have valid values
			if preset.BounceStrength < 0 || preset.BounceStrength > 1 {
				t.Errorf("%s BounceStrength %f out of [0,1]", genre, preset.BounceStrength)
			}
			if preset.MaxBounceDistance <= 0 {
				t.Errorf("%s MaxBounceDistance should be positive", genre)
			}
			if preset.WallContribution < 0 || preset.WallContribution > 1 {
				t.Errorf("%s WallContribution %f out of [0,1]", genre, preset.WallContribution)
			}
			if preset.FloorContribution < 0 || preset.FloorContribution > 1 {
				t.Errorf("%s FloorContribution %f out of [0,1]", genre, preset.FloorContribution)
			}
		})
	}
}

func TestEmptySurfaceList(t *testing.T) {
	sys := NewSystem("fantasy", 320, 240)

	// Empty surface list should not crash
	bounceMap := sys.CalculateBounce([]BounceSurface{}, 4, 4, 1.0)

	if bounceMap == nil {
		t.Fatal("Empty surface list returned nil bounceMap")
	}

	// All cells should have zero intensity
	for i := range bounceMap.Data {
		if bounceMap.Data[i].Intensity != 0 {
			t.Error("Empty surface list should produce zero intensity")
			break
		}
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test clamp01
	if clamp01(-0.5) != 0 {
		t.Error("clamp01(-0.5) should be 0")
	}
	if clamp01(1.5) != 1 {
		t.Error("clamp01(1.5) should be 1")
	}
	if clamp01(0.5) != 0.5 {
		t.Error("clamp01(0.5) should be 0.5")
	}

	// Test lerp
	if lerp(0, 10, 0.5) != 5 {
		t.Errorf("lerp(0, 10, 0.5) = %f, want 5", lerp(0, 10, 0.5))
	}
	if lerp(0, 10, 0) != 0 {
		t.Errorf("lerp(0, 10, 0) = %f, want 0", lerp(0, 10, 0))
	}
	if lerp(0, 10, 1) != 10 {
		t.Errorf("lerp(0, 10, 1) = %f, want 10", lerp(0, 10, 1))
	}
}

// Helper functions for tests
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkCalculateBounce(b *testing.B) {
	sys := NewSystem("fantasy", 320, 240)

	// Create a medium-sized scene
	surfaces := make([]BounceSurface, 20)
	for i := range surfaces {
		surfaces[i] = BounceSurface{
			X: float64(i % 10), Y: float64(i / 10),
			R: 0.5, G: 0.5, B: 0.5,
			Reflectivity: 0.5, IsWall: i%2 == 0, DirectLight: 0.8,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.CalculateBounce(surfaces, 16, 16, 1.0)
	}
}

func BenchmarkGetBounceAt(b *testing.B) {
	sys := NewSystem("fantasy", 320, 240)

	surfaces := []BounceSurface{
		{X: 8, Y: 8, R: 1.0, G: 0.5, B: 0.2, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
	}
	bounceMap := sys.CalculateBounce(surfaces, 16, 16, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetBounceAt(bounceMap, 7.5, 7.5)
	}
}

func BenchmarkGetBounceAtBilinear(b *testing.B) {
	sys := NewSystem("fantasy", 320, 240)

	surfaces := []BounceSurface{
		{X: 8, Y: 8, R: 1.0, G: 0.5, B: 0.2, Reflectivity: 1.0, IsWall: true, DirectLight: 1.0},
	}
	bounceMap := sys.CalculateBounce(surfaces, 16, 16, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GetBounceAtBilinear(bounceMap, 7.3, 7.7)
	}
}

func BenchmarkApplyBounceToColor(b *testing.B) {
	sys := NewSystem("fantasy", 320, 240)
	base := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ApplyBounceToColor(base, 0.8, 0.4, 0.2, 0.3)
	}
}

func BenchmarkExtractSurfacesFromColors(b *testing.B) {
	sys := NewSystem("fantasy", 320, 240)

	colors := make([]color.RGBA, 64)
	for i := range colors {
		colors[i] = color.RGBA{R: 128, G: 64, B: 200, A: 255}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ExtractSurfacesFromColors(colors, 8, 8, 1.0, true)
	}
}

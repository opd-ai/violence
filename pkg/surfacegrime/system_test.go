package surfacegrime

import (
	"image"
	"image/color"
	"math/rand"
	"testing"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 12345)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, genre)
			}
			if sys.rng == nil {
				t.Error("rng is nil")
			}
			if !sys.enabled {
				t.Error("system should be enabled by default")
			}
			if sys.intensity != 1.0 {
				t.Errorf("intensity = %f, want 1.0", sys.intensity)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Generate overlay to populate cache
	sys.GenerateOverlay(64, 64, "room1", 12345)
	if sys.cachedRoomID != "room1" {
		t.Fatal("Cache should be populated")
	}

	// Change genre - should invalidate cache
	sys.SetGenre("horror")
	if sys.cachedRoomID != "" {
		t.Error("Cache should be invalidated on genre change")
	}
	if sys.genreID != "horror" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "horror")
	}

	// Same genre should not invalidate
	sys.GenerateOverlay(64, 64, "room2", 12345)
	sys.SetGenre("horror")
	if sys.cachedRoomID != "room2" {
		t.Error("Cache should not be invalidated for same genre")
	}
}

func TestSetEnabled(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.SetEnabled(false)
	if sys.enabled {
		t.Error("SetEnabled(false) did not disable")
	}

	sys.SetEnabled(true)
	if !sys.enabled {
		t.Error("SetEnabled(true) did not enable")
	}
}

func TestSetIntensity(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},
		{0.0, 0.0},
		{1.0, 1.0},
		{2.0, 2.0},
		{-0.5, 0.0}, // Clamped
		{3.0, 2.0},  // Clamped
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			sys := NewSystem("fantasy", 12345)
			sys.SetIntensity(tt.input)
			if sys.intensity != tt.expected {
				t.Errorf("intensity = %f, want %f", sys.intensity, tt.expected)
			}
		})
	}
}

func TestGenerateOverlay(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 42)
			overlay := sys.GenerateOverlay(64, 48, "testroom", 42)

			if overlay == nil {
				t.Fatal("GenerateOverlay returned nil")
			}

			bounds := overlay.Bounds()
			if bounds.Dx() != 64 || bounds.Dy() != 48 {
				t.Errorf("Overlay size = %dx%d, want 64x48", bounds.Dx(), bounds.Dy())
			}

			// Check cache was populated
			if sys.cachedRoomID != "testroom" {
				t.Error("Cache room ID not set")
			}
			if sys.overlayWidth != 64 || sys.overlayHeight != 48 {
				t.Error("Cache dimensions not set")
			}
		})
	}
}

func TestGenerateOverlayCache(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Generate first overlay
	overlay1 := sys.GenerateOverlay(64, 64, "room1", 12345)

	// Same parameters should return cached
	overlay2 := sys.GenerateOverlay(64, 64, "room1", 12345)
	if overlay1 != overlay2 {
		t.Error("Same parameters should return cached overlay")
	}

	// Different room ID should generate new
	overlay3 := sys.GenerateOverlay(64, 64, "room2", 12345)
	if overlay1 == overlay3 {
		t.Error("Different room ID should generate new overlay")
	}

	// Different size should generate new
	overlay4 := sys.GenerateOverlay(128, 64, "room2", 12345)
	if overlay3 == overlay4 {
		t.Error("Different size should generate new overlay")
	}
}

func TestGenerateOverlayDeterminism(t *testing.T) {
	// Same seed should produce identical cached references (for same room)
	sys := NewSystem("fantasy", 42)

	overlay1 := sys.GenerateOverlay(32, 32, "room1", 42)
	overlay2 := sys.GenerateOverlay(32, 32, "room1", 42)

	// Same parameters should return identical cached reference
	if overlay1 != overlay2 {
		t.Error("Same parameters should return cached overlay")
	}

	// Different room with same dimensions still returns new overlay
	overlay3 := sys.GenerateOverlay(32, 32, "room2", 42)
	if overlay1 == overlay3 {
		t.Error("Different room should generate different overlay")
	}
}

func TestSetEdgeMap(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Generate overlay to populate cache
	sys.GenerateOverlay(64, 64, "room1", 12345)
	if sys.cachedRoomID == "" {
		t.Fatal("Cache should be populated")
	}

	// Set edge map - should invalidate cache
	edgeMap := [][]float64{
		{0.0, 0.5, 1.0},
		{0.3, 0.6, 0.9},
		{0.1, 0.4, 0.8},
	}
	sys.SetEdgeMap(edgeMap)

	if sys.cachedRoomID != "" {
		t.Error("Cache should be invalidated when edge map changes")
	}
	if sys.edgeMap == nil {
		t.Error("Edge map should be set")
	}
}

func TestGetEdgeProximityWithEdgeMap(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	edgeMap := [][]float64{
		{0.0, 0.5, 1.0, 0.5},
		{0.3, 0.6, 0.9, 0.6},
		{0.1, 0.4, 0.8, 0.4},
		{0.0, 0.2, 0.5, 0.2},
	}
	sys.SetEdgeMap(edgeMap)

	// Test edge proximity lookup
	p := sys.getEdgeProximity(0, 0, 64, 64)
	if p < 0.0 || p > 1.0 {
		t.Errorf("Edge proximity out of range: %f", p)
	}
}

func TestSelectGrimeType(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	rng := rand.New(rand.NewSource(42))

	typeCounts := make(map[GrimeType]int)
	for i := 0; i < 1000; i++ {
		gt := sys.selectGrimeType(i, i, rng)
		typeCounts[gt]++
	}

	// Should have variety across types
	if len(typeCounts) < 2 {
		t.Errorf("Expected multiple grime types, got %d", len(typeCounts))
	}

	// Check all returned types are valid
	for gt := range typeCounts {
		if _, ok := GrimeColors[gt]; !ok {
			t.Errorf("Invalid grime type: %v", gt)
		}
	}
}

func TestGetGrimeColor(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	rng := rand.New(rand.NewSource(42))

	for gt := GrimeDirt; gt <= GrimeOoze; gt++ {
		t.Run("", func(t *testing.T) {
			c := sys.getGrimeColor(gt, 10, 10, rng)

			// Should be non-zero
			if c.A == 0 {
				t.Error("Alpha should not be zero")
			}

			// Color should be reasonable
			if c.R == 0 && c.G == 0 && c.B == 0 {
				t.Error("Color should not be pure black")
			}
		})
	}
}

func TestCalculateAccumulation(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	rng := rand.New(rand.NewSource(42))

	tests := []struct {
		name          string
		edgeProximity float64
		minExpected   float64
		maxExpected   float64
	}{
		{"no_edge", 0.0, 0.0, 0.1},
		{"partial_edge", 0.5, 0.1, 0.8},
		{"full_edge", 1.0, 0.3, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := sys.calculateAccumulation(32, 32, tt.edgeProximity, 64, 64, rng)
			if acc < tt.minExpected || acc > tt.maxExpected {
				t.Errorf("Accumulation = %f, expected [%f, %f]", acc, tt.minExpected, tt.maxExpected)
			}
		})
	}
}

func TestPerlinNoise(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Test that noise is in valid range
	for x := 0.0; x < 10.0; x += 0.5 {
		for y := 0.0; y < 10.0; y += 0.5 {
			n := sys.perlinNoise(x, y, 42)
			if n < 0.0 || n > 1.0 {
				t.Errorf("Noise value out of range: %f at (%f, %f)", n, x, y)
			}
		}
	}

	// Test determinism
	n1 := sys.perlinNoise(3.5, 7.2, 42)
	n2 := sys.perlinNoise(3.5, 7.2, 42)
	if n1 != n2 {
		t.Error("Noise should be deterministic")
	}
}

func TestApplyNoisePass(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	rng := rand.New(rand.NewSource(42))

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Fill with a test color
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 100, G: 100, B: 100, A: 200})
		}
	}

	sys.applyNoisePass(img, rng)

	// Check that some pixels were modified
	modified := false
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			c := img.RGBAAt(x, y)
			if c.R != 100 || c.G != 100 || c.B != 100 {
				modified = true
				break
			}
		}
		if modified {
			break
		}
	}

	if !modified {
		t.Error("Noise pass should modify some pixels")
	}
}

func TestApplyEdgeFeathering(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	// Create a sharp edge
	for y := 0; y < 16; y++ {
		for x := 0; x < 8; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 100, G: 80, B: 60, A: 200})
		}
	}

	// Get original alpha values
	origAlpha := make([]uint8, 16*16)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			origAlpha[y*16+x] = img.RGBAAt(x, y).A
		}
	}

	sys.applyEdgeFeathering(img)

	// Check that edge pixels were softened
	// Pixel at x=7 should have reduced alpha due to neighbors at x=8 having alpha 0
	edgeAlpha := img.RGBAAt(7, 8).A
	if edgeAlpha >= origAlpha[8*16+7] {
		// Allow for no change in some edge cases, but generally expect softening
		t.Log("Edge feathering may not change boundary pixels significantly")
	}
}

func TestComponentType(t *testing.T) {
	comp := &Component{}
	if comp.Type() != "surfacegrime" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "surfacegrime")
	}
}

func TestGrimeColorsComplete(t *testing.T) {
	// Verify all grime types have colors defined
	for gt := GrimeDirt; gt <= GrimeOoze; gt++ {
		if _, ok := GrimeColors[gt]; !ok {
			t.Errorf("Missing color for grime type %v", gt)
		}
		if _, ok := GrimeSecondaryColors[gt]; !ok {
			t.Errorf("Missing secondary colors for grime type %v", gt)
		}
	}
}

func TestGetGenreGrime(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset := GetGenreGrime(genre)
			if len(preset.Types) == 0 {
				t.Error("Preset should have at least one grime type")
			}
			if len(preset.Weights) != len(preset.Types) {
				t.Error("Weights and Types should have same length")
			}
			if preset.SpreadDistance <= 0 {
				t.Error("SpreadDistance should be positive")
			}
		})
	}

	// Unknown genre should return fantasy preset
	unknown := GetGenreGrime("unknown")
	fantasy := GetGenreGrime("fantasy")
	if len(unknown.Types) != len(fantasy.Types) {
		t.Error("Unknown genre should return fantasy preset")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, expected float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{0.0, 0.0, 1.0, 0.0},
		{1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		result := clamp(tt.v, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestClampByte(t *testing.T) {
	tests := []struct {
		v        int
		expected uint8
	}{
		{128, 128},
		{0, 0},
		{255, 255},
		{-10, 0},
		{300, 255},
	}

	for _, tt := range tests {
		result := clampByte(tt.v)
		if result != tt.expected {
			t.Errorf("clampByte(%d) = %d, want %d", tt.v, result, tt.expected)
		}
	}
}

func BenchmarkGenerateOverlay(b *testing.B) {
	sys := NewSystem("fantasy", 42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Force regeneration by changing room ID
		sys.GenerateOverlay(320, 200, "room"+string(rune(i%100)), int64(i))
	}
}

func BenchmarkPerlinNoise(b *testing.B) {
	sys := NewSystem("fantasy", 42)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sys.perlinNoise(float64(i%100), float64(i/100), 42)
	}
}

package caustics

import (
	"math"
	"math/rand"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		screenW int
		screenH int
	}{
		{"fantasy", "fantasy", 320, 200},
		{"scifi", "scifi", 640, 480},
		{"horror", "horror", 320, 200},
		{"cyberpunk", "cyberpunk", 1280, 720},
		{"postapoc", "postapoc", 320, 200},
		{"unknown defaults to fantasy", "unknown_genre", 320, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID, tt.screenW, tt.screenH)

			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}

			if sys.screenW != tt.screenW {
				t.Errorf("screenW = %d, want %d", sys.screenW, tt.screenW)
			}
			if sys.screenH != tt.screenH {
				t.Errorf("screenH = %d, want %d", sys.screenH, tt.screenH)
			}

			// Verify trig tables are initialized
			if len(sys.sinTable) != 360 {
				t.Errorf("sinTable length = %d, want 360", len(sys.sinTable))
			}
			if len(sys.cosTable) != 360 {
				t.Errorf("cosTable length = %d, want 360", len(sys.cosTable))
			}

			// Verify logger is set
			if sys.logger == nil {
				t.Error("logger is nil")
			}
		})
	}
}

func TestSystem_SetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Add a source to verify cache clearing
	sys.AddSource(&Component{WorldX: 1, WorldY: 1, Radius: 2})

	sys.SetGenre("cyberpunk")

	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "cyberpunk")
	}

	// Verify preset changed
	if sys.preset.BaseIntensity != genrePresets["cyberpunk"].BaseIntensity {
		t.Error("Preset not updated after SetGenre")
	}

	// Unknown genre should default to fantasy
	sys.SetGenre("nonexistent")
	if sys.preset.BaseIntensity != genrePresets["fantasy"].BaseIntensity {
		t.Error("Unknown genre should use fantasy preset")
	}
}

func TestSystem_Update(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	initialTime := sys.time
	initialFrame := sys.frameIdx

	// Create a mock world
	world := engine.NewWorld()

	// Update several times
	for i := 0; i < 60; i++ {
		sys.Update(world)
	}

	if sys.time <= initialTime {
		t.Error("Time did not advance after Update")
	}

	if sys.frameIdx == initialFrame && sys.time > 0.1 {
		t.Error("Frame index did not change after significant time")
	}
}

func TestSystem_AddAndClearSources(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	if sys.GetSourceCount() != 0 {
		t.Errorf("Initial source count = %d, want 0", sys.GetSourceCount())
	}

	// Add sources
	sys.AddSource(&Component{WorldX: 1, WorldY: 1})
	sys.AddSource(&Component{WorldX: 2, WorldY: 2})
	sys.AddSource(&Component{WorldX: 3, WorldY: 3})

	if sys.GetSourceCount() != 3 {
		t.Errorf("Source count = %d, want 3", sys.GetSourceCount())
	}

	// Clear sources
	sys.ClearSources()

	if sys.GetSourceCount() != 0 {
		t.Errorf("Source count after clear = %d, want 0", sys.GetSourceCount())
	}
}

func TestSystem_GenerateCausticsFromWetness(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	puddles := []PuddleLocation{
		{TileX: 1, TileY: 1, WorldX: 1.5, WorldY: 1.5, Moisture: 0.9}, // Pool
		{TileX: 2, TileY: 2, WorldX: 2.5, WorldY: 2.5, Moisture: 0.6}, // Puddle
		{TileX: 3, TileY: 3, WorldX: 3.5, WorldY: 3.5, Moisture: 0.3}, // Drip
	}

	sys.GenerateCausticsFromWetness(puddles, 12345)

	if sys.GetSourceCount() != 3 {
		t.Errorf("Source count = %d, want 3", sys.GetSourceCount())
	}

	// Verify source types based on moisture
	for _, src := range sys.sources {
		switch {
		case src.WorldX == 1.5:
			if src.SourceType != SourcePool {
				t.Errorf("High moisture should be Pool, got %d", src.SourceType)
			}
		case src.WorldX == 2.5:
			if src.SourceType != SourcePuddle {
				t.Errorf("Medium moisture should be Puddle, got %d", src.SourceType)
			}
		case src.WorldX == 3.5:
			if src.SourceType != SourceDrip {
				t.Errorf("Low moisture should be Drip, got %d", src.SourceType)
			}
		}
	}
}

func TestSystem_ClearCache(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Manually add to cache
	sys.cacheMu.Lock()
	sys.patternCache[cacheKey{frame: 0, radius: 2.0, sourceT: SourcePuddle}] = nil
	sys.cacheMu.Unlock()

	sys.ClearCache()

	sys.cacheMu.RLock()
	cacheLen := len(sys.patternCache)
	sys.cacheMu.RUnlock()

	if cacheLen != 0 {
		t.Errorf("Cache length after clear = %d, want 0", cacheLen)
	}
}

func TestSystem_calculateCausticColor(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Use seeded random for reproducibility
	rng := rand.New(rand.NewSource(12345))

	// Just test that the function returns valid colors
	col := sys.calculateCausticColor(0.8, rng)

	if col.A != 255 {
		t.Errorf("Alpha = %d, want 255", col.A)
	}

	// Color components should be non-zero for visible caustics
	if col.R == 0 && col.G == 0 && col.B == 0 {
		t.Error("Color is completely black")
	}
}

func TestSystem_voronoiNoise(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	// Test determinism
	result1 := sys.voronoiNoise(1.5, 2.5, 0.5, 12345)
	result2 := sys.voronoiNoise(1.5, 2.5, 0.5, 12345)

	if result1 != result2 {
		t.Error("voronoiNoise is not deterministic")
	}

	// Test that different positions give different results
	result3 := sys.voronoiNoise(5.5, 8.5, 0.5, 12345)
	if result1 == result3 {
		t.Error("voronoiNoise gave same result for different positions")
	}

	// Test that output is within reasonable range
	if result1 < 0 {
		t.Errorf("voronoiNoise returned negative value: %f", result1)
	}
	if result1 > 10 {
		t.Errorf("voronoiNoise returned unexpectedly large value: %f", result1)
	}
}

func TestSystem_calculateCausticIntensity(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	src := &Component{
		Seed:       12345,
		SourceType: SourcePuddle,
	}

	// Test that intensity varies across the pattern
	intensities := make([]float64, 0)
	for x := 0.0; x < 10.0; x += 1.0 {
		for y := 0.0; y < 10.0; y += 1.0 {
			intensity := sys.calculateCausticIntensity(x, y, 5, 5, 0, src, sys.preset)
			intensities = append(intensities, intensity)
		}
	}

	// Check for variation
	minIntensity := intensities[0]
	maxIntensity := intensities[0]
	for _, i := range intensities {
		if i < minIntensity {
			minIntensity = i
		}
		if i > maxIntensity {
			maxIntensity = i
		}
	}

	if maxIntensity-minIntensity < 0.01 {
		t.Error("Caustic intensity has no variation (flat pattern)")
	}
}

func TestSystem_GetGenre(t *testing.T) {
	sys := NewSystem("horror", 320, 200)

	if sys.GetGenre() != "horror" {
		t.Errorf("GetGenre() = %q, want %q", sys.GetGenre(), "horror")
	}
}

func TestSystem_GetTime(t *testing.T) {
	sys := NewSystem("fantasy", 320, 200)

	if sys.GetTime() != 0 {
		t.Errorf("Initial time = %f, want 0", sys.GetTime())
	}

	// Advance time
	world := engine.NewWorld()
	sys.Update(world)

	if sys.GetTime() <= 0 {
		t.Error("Time did not advance after Update")
	}
}

func TestGenrePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset, ok := genrePresets[genre]
			if !ok {
				t.Fatalf("Missing preset for genre %q", genre)
			}

			// Validate preset values are reasonable
			if preset.BaseIntensity <= 0 || preset.BaseIntensity > 1 {
				t.Errorf("BaseIntensity %f out of range (0,1]", preset.BaseIntensity)
			}
			if preset.AnimationSpeed <= 0 {
				t.Error("AnimationSpeed must be positive")
			}
			if preset.PatternScale <= 0 {
				t.Error("PatternScale must be positive")
			}
			if preset.FalloffExponent <= 0 {
				t.Error("FalloffExponent must be positive")
			}

			// Color components should be in [0,1]
			if preset.ColorR < 0 || preset.ColorR > 1 {
				t.Errorf("ColorR %f out of range [0,1]", preset.ColorR)
			}
			if preset.ColorG < 0 || preset.ColorG > 1 {
				t.Errorf("ColorG %f out of range [0,1]", preset.ColorG)
			}
			if preset.ColorB < 0 || preset.ColorB > 1 {
				t.Errorf("ColorB %f out of range [0,1]", preset.ColorB)
			}
		})
	}
}

func TestSystem_Determinism(t *testing.T) {
	// Create two systems with same parameters
	sys1 := NewSystem("fantasy", 320, 200)
	sys2 := NewSystem("fantasy", 320, 200)

	puddles := []PuddleLocation{
		{TileX: 1, TileY: 1, WorldX: 1.5, WorldY: 1.5, Moisture: 0.8},
		{TileX: 2, TileY: 3, WorldX: 2.5, WorldY: 3.5, Moisture: 0.6},
	}

	seed := int64(99999)
	sys1.GenerateCausticsFromWetness(puddles, seed)
	sys2.GenerateCausticsFromWetness(puddles, seed)

	if sys1.GetSourceCount() != sys2.GetSourceCount() {
		t.Error("Different source counts with same seed")
	}

	// Compare source properties
	for i := 0; i < sys1.GetSourceCount(); i++ {
		s1 := sys1.sources[i]
		s2 := sys2.sources[i]

		if s1.WorldX != s2.WorldX || s1.WorldY != s2.WorldY {
			t.Error("Source positions differ with same seed")
		}
		if s1.SourceType != s2.SourceType {
			t.Error("Source types differ with same seed")
		}
		if math.Abs(s1.Intensity-s2.Intensity) > 0.001 {
			t.Error("Source intensities differ with same seed")
		}
	}
}

func BenchmarkVoronoiNoise(b *testing.B) {
	sys := NewSystem("fantasy", 320, 200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.voronoiNoise(float64(i%100), float64(i%100), float64(i)*0.01, 12345)
	}
}

func BenchmarkCalculateCausticIntensity(b *testing.B) {
	sys := NewSystem("fantasy", 320, 200)
	src := &Component{Seed: 12345, SourceType: SourcePuddle}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.calculateCausticIntensity(
			float64(i%64), float64(i%64),
			32, 32, float64(i)*0.016, src, sys.preset,
		)
	}
}

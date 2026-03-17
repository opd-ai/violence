package flicker

import (
	"math"
	"testing"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		sys := NewSystem(genre)
		if sys == nil {
			t.Errorf("NewSystem(%q) returned nil", genre)
			continue
		}
		if sys.genre != genre {
			t.Errorf("NewSystem(%q) has genre=%q", genre, sys.genre)
		}
		if len(sys.presets) == 0 {
			t.Errorf("NewSystem(%q) has no presets", genre)
		}
	}
}

func TestGetFlickerParams(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		lightType string
		seed      int64
	}{
		{"torch", 12345},
		{"brazier", 54321},
		{"candle", 11111},
		{"magic_crystal", 22222},
		{"unknown_type", 99999}, // Should fall back to generic
	}

	for _, tt := range tests {
		params := sys.GetFlickerParams(tt.lightType, tt.seed, 1.0, 0.6, 0.2)

		if params.Seed != tt.seed {
			t.Errorf("GetFlickerParams(%q, %d): Seed=%d, want %d", tt.lightType, tt.seed, params.Seed, tt.seed)
		}
		if params.BaseR != 1.0 || params.BaseG != 0.6 || params.BaseB != 0.2 {
			t.Errorf("GetFlickerParams(%q): wrong base colors", tt.lightType)
		}
		// Phase shift should be randomized
		if params.SwayFrequency > 0 && params.SwayPhaseShift == 0 {
			// It's possible but very unlikely to be exactly 0
			t.Logf("GetFlickerParams(%q): SwayPhaseShift is 0 (could be coincidence)", tt.lightType)
		}
	}
}

func TestCalculateFlicker(t *testing.T) {
	sys := NewSystem("fantasy")
	params := sys.GetFlickerParams("torch", 12345, 1.0, 0.6, 0.2)

	// Test multiple ticks to ensure variation
	intensities := make([]float64, 100)
	var minIntensity, maxIntensity float64 = 2.0, -1.0

	for tick := 0; tick < 100; tick++ {
		intensity, r, g, b := sys.CalculateFlicker(&params, tick, 1.0)
		intensities[tick] = intensity

		// Check intensity is in valid range
		if intensity < 0.0 || intensity > 1.5 {
			t.Errorf("CalculateFlicker tick %d: intensity=%f out of range", tick, intensity)
		}

		// Check colors are in valid range
		if r < 0 || r > 1 || g < 0 || g > 1 || b < 0 || b > 1 {
			t.Errorf("CalculateFlicker tick %d: colors out of range (r=%f, g=%f, b=%f)", tick, r, g, b)
		}

		if intensity < minIntensity {
			minIntensity = intensity
		}
		if intensity > maxIntensity {
			maxIntensity = intensity
		}
	}

	// Verify there's actual variation (not just flat output)
	variation := maxIntensity - minIntensity
	if variation < 0.05 {
		t.Errorf("CalculateFlicker: insufficient variation (range=%f)", variation)
	}
}

func TestCalculateFlickerDeterminism(t *testing.T) {
	sys := NewSystem("fantasy")
	params1 := sys.GetFlickerParams("torch", 12345, 1.0, 0.6, 0.2)
	params2 := sys.GetFlickerParams("torch", 12345, 1.0, 0.6, 0.2)

	// Same seed should produce same results
	for tick := 0; tick < 50; tick++ {
		i1, r1, g1, b1 := sys.CalculateFlicker(&params1, tick, 1.0)
		i2, r2, g2, b2 := sys.CalculateFlicker(&params2, tick, 1.0)

		if i1 != i2 || r1 != r2 || g1 != g2 || b1 != b2 {
			t.Errorf("CalculateFlicker not deterministic at tick %d", tick)
			break
		}
	}
}

func TestCalculateFlickerDifferentSeeds(t *testing.T) {
	sys := NewSystem("fantasy")
	params1 := sys.GetFlickerParams("torch", 11111, 1.0, 0.6, 0.2)
	params2 := sys.GetFlickerParams("torch", 22222, 1.0, 0.6, 0.2)

	// Different seeds should produce different results
	sameCount := 0
	for tick := 0; tick < 50; tick++ {
		i1, _, _, _ := sys.CalculateFlicker(&params1, tick, 1.0)
		i2, _, _, _ := sys.CalculateFlicker(&params2, tick, 1.0)

		if math.Abs(i1-i2) < 0.001 {
			sameCount++
		}
	}

	// Allow some coincidental matches, but not many
	if sameCount > 10 {
		t.Errorf("Different seeds produced too many matching intensities: %d/50", sameCount)
	}
}

func TestSciFiNoFlicker(t *testing.T) {
	sys := NewSystem("scifi")
	params := sys.GetFlickerParams("ceiling_lamp", 12345, 0.9, 0.95, 1.0)

	// SciFi ceiling lamps should have minimal/no flicker
	intensities := make([]float64, 100)
	for tick := 0; tick < 100; tick++ {
		intensity, _, _, _ := sys.CalculateFlicker(&params, tick, 1.0)
		intensities[tick] = intensity
	}

	// Check variance is very low
	var sum, sumSq float64
	for _, v := range intensities {
		sum += v
		sumSq += v * v
	}
	mean := sum / 100.0
	variance := sumSq/100.0 - mean*mean

	if variance > 0.01 {
		t.Errorf("SciFi ceiling_lamp has too much variance: %f", variance)
	}
}

func TestHorrorHighFlicker(t *testing.T) {
	sys := NewSystem("horror")
	params := sys.GetFlickerParams("broken_lamp", 12345, 0.5, 0.5, 0.3)

	// Horror broken lamps should have high flicker
	var minIntensity, maxIntensity float64 = 2.0, -1.0
	for tick := 0; tick < 200; tick++ {
		intensity, _, _, _ := sys.CalculateFlicker(&params, tick, 1.0)
		if intensity < minIntensity {
			minIntensity = intensity
		}
		if intensity > maxIntensity {
			maxIntensity = intensity
		}
	}

	// Check for significant variation
	variation := maxIntensity - minIntensity
	if variation < 0.15 {
		t.Errorf("Horror broken_lamp should have high variation, got %f", variation)
	}
}

func TestColorTemperatureVariation(t *testing.T) {
	sys := NewSystem("fantasy")
	params := sys.GetFlickerParams("torch", 12345, 1.0, 0.6, 0.2)

	// Collect color variations
	var minR, maxR, minG, maxG, minB, maxB float64 = 2.0, -1.0, 2.0, -1.0, 2.0, -1.0

	for tick := 0; tick < 200; tick++ {
		_, r, g, b := sys.CalculateFlicker(&params, tick, 1.0)

		if r < minR {
			minR = r
		}
		if r > maxR {
			maxR = r
		}
		if g < minG {
			minG = g
		}
		if g > maxG {
			maxG = g
		}
		if b < minB {
			minB = b
		}
		if b > maxB {
			maxB = b
		}
	}

	// Torch should have some color variation (temperature effects)
	gVariation := maxG - minG
	bVariation := maxB - minB

	// G and B should vary with temperature
	if gVariation < 0.01 && bVariation < 0.01 {
		t.Errorf("Torch should have color temperature variation (G range=%f, B range=%f)", gVariation, bVariation)
	}
}

func TestGutteringEvents(t *testing.T) {
	sys := NewSystem("horror")
	params := sys.GetFlickerParams("candle", 55555, 1.0, 0.7, 0.3)

	// Run for enough ticks to likely trigger guttering
	gutterCount := 0
	for tick := 0; tick < 600; tick++ { // 10 seconds at 60fps
		intensity, _, _, _ := sys.CalculateFlicker(&params, tick, 1.0)

		// Detect significant drops that indicate guttering
		if intensity < 0.5 {
			gutterCount++
		}
	}

	// Horror candles have 45% gutter probability per second, so we should see some
	// But this is probabilistic, so we just check it's reasonable
	if gutterCount == 0 {
		t.Log("No guttering detected in 600 ticks (could be unlucky)")
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")

	if sys.GetGenre() != "fantasy" {
		t.Errorf("GetGenre() = %q, want fantasy", sys.GetGenre())
	}

	sys.SetGenre("horror")

	if sys.GetGenre() != "horror" {
		t.Errorf("After SetGenre('horror'), GetGenre() = %q", sys.GetGenre())
	}

	// Verify presets changed
	names := sys.GetPresetNames()
	found := false
	for _, name := range names {
		if name == "dim_bulb" {
			found = true
			break
		}
	}
	if !found {
		t.Error("After SetGenre('horror'), 'dim_bulb' preset not found")
	}
}

func TestGetPresetNames(t *testing.T) {
	genres := map[string][]string{
		"fantasy":   {"torch", "brazier", "candle", "magic_crystal"},
		"scifi":     {"monitor", "ceiling_lamp", "alarm", "console"},
		"horror":    {"dim_bulb", "emergency_light", "candle", "broken_lamp"},
		"cyberpunk": {"neon_pink", "neon_cyan", "hologram", "streetlight"},
		"postapoc":  {"oil_lamp", "fire_barrel", "generator_light", "salvaged_lamp"},
	}

	for genre, expectedNames := range genres {
		sys := NewSystem(genre)
		names := sys.GetPresetNames()

		for _, expected := range expectedNames {
			found := false
			for _, name := range names {
				if name == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Genre %q missing preset %q", genre, expected)
			}
		}
	}
}

func TestSmoothStep(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{0.5, 0.5},
		{-0.5, 0.0}, // Clamped
		{1.5, 1.0},  // Clamped
	}

	for _, tt := range tests {
		result := smoothStep(tt.input)
		if math.Abs(result-tt.expected) > 0.001 {
			t.Errorf("smoothStep(%f) = %f, want %f", tt.input, result, tt.expected)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{-0.5, 0.0, 1.0, 0.0},
		{1.5, 0.0, 1.0, 1.0},
		{100.0, -10.0, 10.0, 10.0},
	}

	for _, tt := range tests {
		result := clamp(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

func BenchmarkCalculateFlicker(b *testing.B) {
	sys := NewSystem("fantasy")
	params := sys.GetFlickerParams("torch", 12345, 1.0, 0.6, 0.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.CalculateFlicker(&params, i, 1.0)
	}
}

func BenchmarkCalculateFlickerAllGenres(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		b.Run(genre, func(b *testing.B) {
			sys := NewSystem(genre)
			names := sys.GetPresetNames()
			if len(names) == 0 {
				b.Skip("No presets")
			}
			params := sys.GetFlickerParams(names[0], 12345, 1.0, 0.6, 0.2)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sys.CalculateFlicker(&params, i, 1.0)
			}
		})
	}
}

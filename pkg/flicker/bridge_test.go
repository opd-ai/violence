package flicker

import (
	"testing"

	"github.com/opd-ai/violence/pkg/lighting"
)

func TestNewLightFlickerBridge(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	if bridge == nil {
		t.Fatal("NewLightFlickerBridge returned nil")
	}
	if bridge.sys == nil {
		t.Error("sys is nil")
	}
	if bridge.params == nil {
		t.Error("params map is nil")
	}
}

func TestLightFlickerBridgeSetGenre(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	// Add some cached params
	preset := lighting.LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0, G: 0.6, B: 0.2,
		Flicker: true,
	}
	light := lighting.NewPointLight(0, 0, preset, 12345)
	bridge.CalculateFlickerForLight(&light, 0)

	if len(bridge.params) == 0 {
		t.Error("Params should be cached after calculation")
	}

	// Change genre
	bridge.SetGenre("horror")

	// Cache should be cleared
	if len(bridge.params) != 0 {
		t.Error("Params cache should be cleared after genre change")
	}

	if bridge.sys.GetGenre() != "horror" {
		t.Errorf("Genre should be 'horror', got %q", bridge.sys.GetGenre())
	}
}

func TestCalculateFlickerForLight(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	preset := lighting.LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0, G: 0.6, B: 0.2,
		Flicker: true,
	}
	light := lighting.NewPointLight(0, 0, preset, 12345)

	// Calculate flicker at multiple ticks
	results := make([]FlickerResult, 60)
	for tick := 0; tick < 60; tick++ {
		results[tick] = bridge.CalculateFlickerForLight(&light, tick)

		// Verify values are in valid range
		if results[tick].Intensity < 0 || results[tick].Intensity > 1.5 {
			t.Errorf("tick %d: intensity out of range: %f", tick, results[tick].Intensity)
		}
		if results[tick].R < 0 || results[tick].R > 1 {
			t.Errorf("tick %d: R out of range: %f", tick, results[tick].R)
		}
		if results[tick].G < 0 || results[tick].G > 1 {
			t.Errorf("tick %d: G out of range: %f", tick, results[tick].G)
		}
		if results[tick].B < 0 || results[tick].B > 1 {
			t.Errorf("tick %d: B out of range: %f", tick, results[tick].B)
		}
	}

	// Verify there's variation
	minI, maxI := results[0].Intensity, results[0].Intensity
	for _, r := range results {
		if r.Intensity < minI {
			minI = r.Intensity
		}
		if r.Intensity > maxI {
			maxI = r.Intensity
		}
	}
	if maxI-minI < 0.01 {
		t.Errorf("No intensity variation detected (range: %f)", maxI-minI)
	}
}

func TestCalculateFlickerForLightComponent(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	preset := lighting.LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0, G: 0.6, B: 0.2,
		Flicker: true,
	}
	lc := lighting.NewLightComponent(preset, 12345)

	result := bridge.CalculateFlickerForLightComponent(lc, 0)

	if result.Intensity <= 0 || result.Intensity > 1.5 {
		t.Errorf("Intensity out of range: %f", result.Intensity)
	}
}

func TestLightFlickerBridgeCaching(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	preset := lighting.LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0, G: 0.6, B: 0.2,
		Flicker: true,
	}
	light := lighting.NewPointLight(0, 0, preset, 12345)

	// First calculation should create params
	bridge.CalculateFlickerForLight(&light, 0)

	params := bridge.GetParams(12345)
	if params == nil {
		t.Error("Params should be cached after calculation")
	}

	// Second calculation should reuse cached params
	bridge.CalculateFlickerForLight(&light, 1)
	params2 := bridge.GetParams(12345)

	if params != params2 {
		t.Error("Same params pointer should be returned for same seed")
	}
}

func TestLightFlickerBridgeClearCache(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	preset := lighting.LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0, G: 0.6, B: 0.2,
		Flicker: true,
	}
	light := lighting.NewPointLight(0, 0, preset, 12345)

	bridge.CalculateFlickerForLight(&light, 0)
	if len(bridge.params) == 0 {
		t.Error("Params should be cached")
	}

	bridge.ClearCache()
	if len(bridge.params) != 0 {
		t.Error("Cache should be empty after ClearCache")
	}
}

func BenchmarkCalculateFlickerForLight(b *testing.B) {
	bridge := NewLightFlickerBridge("fantasy")

	preset := lighting.LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0, G: 0.6, B: 0.2,
		Flicker: true,
	}
	light := lighting.NewPointLight(0, 0, preset, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.CalculateFlickerForLight(&light, i)
	}
}

func TestCalculateFlickerSimple(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	// Calculate flicker at multiple ticks
	results := make([]FlickerResult, 60)
	for tick := 0; tick < 60; tick++ {
		results[tick] = bridge.CalculateFlickerSimple(12345, tick, 0.9, 1.0, 0.8, 0.4)

		// Verify values are in valid range
		if results[tick].Intensity < 0 || results[tick].Intensity > 1.5 {
			t.Errorf("tick %d: intensity out of range: %f", tick, results[tick].Intensity)
		}
		if results[tick].R < 0 || results[tick].R > 1 {
			t.Errorf("tick %d: R out of range: %f", tick, results[tick].R)
		}
		if results[tick].G < 0 || results[tick].G > 1 {
			t.Errorf("tick %d: G out of range: %f", tick, results[tick].G)
		}
		if results[tick].B < 0 || results[tick].B > 1 {
			t.Errorf("tick %d: B out of range: %f", tick, results[tick].B)
		}
	}

	// Verify there's variation
	minI, maxI := results[0].Intensity, results[0].Intensity
	for _, r := range results {
		if r.Intensity < minI {
			minI = r.Intensity
		}
		if r.Intensity > maxI {
			maxI = r.Intensity
		}
	}
	if maxI-minI < 0.01 {
		t.Errorf("No intensity variation detected (range: %f)", maxI-minI)
	}
}

func TestCalculateFlickerSimpleCaching(t *testing.T) {
	bridge := NewLightFlickerBridge("fantasy")

	// First calculation should create params
	bridge.CalculateFlickerSimple(12345, 0, 0.9, 1.0, 0.8, 0.4)

	params := bridge.GetParams(12345)
	if params == nil {
		t.Error("Params should be cached after calculation")
	}

	// Second calculation should reuse cached params
	bridge.CalculateFlickerSimple(12345, 1, 0.9, 1.0, 0.8, 0.4)
	params2 := bridge.GetParams(12345)

	if params != params2 {
		t.Error("Same params pointer should be returned for same seed")
	}
}

func BenchmarkCalculateFlickerSimple(b *testing.B) {
	bridge := NewLightFlickerBridge("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bridge.CalculateFlickerSimple(int64(i%100), i, 0.9, 1.0, 0.8, 0.4)
	}
}

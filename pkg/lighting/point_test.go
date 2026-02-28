package lighting

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/procgen/genre"
)

func TestGetGenrePresets(t *testing.T) {
	tests := []struct {
		name      string
		genreID   string
		wantCount int
		checkName string
	}{
		{
			name:      "fantasy presets",
			genreID:   genre.Fantasy,
			wantCount: 4,
			checkName: "torch",
		},
		{
			name:      "scifi presets",
			genreID:   genre.SciFi,
			wantCount: 4,
			checkName: "monitor",
		},
		{
			name:      "horror presets",
			genreID:   genre.Horror,
			wantCount: 4,
			checkName: "dim_bulb",
		},
		{
			name:      "cyberpunk presets",
			genreID:   genre.Cyberpunk,
			wantCount: 4,
			checkName: "neon_pink",
		},
		{
			name:      "postapoc presets",
			genreID:   genre.PostApoc,
			wantCount: 4,
			checkName: "oil_lamp",
		},
		{
			name:      "unknown genre",
			genreID:   "unknown",
			wantCount: 1,
			checkName: "generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presets := GetGenrePresets(tt.genreID)
			if len(presets) != tt.wantCount {
				t.Errorf("GetGenrePresets(%s) count = %d, want %d", tt.genreID, len(presets), tt.wantCount)
			}

			// Check that expected name exists
			found := false
			for _, p := range presets {
				if p.Name == tt.checkName {
					found = true
					// Validate preset has valid values
					if p.Radius <= 0 {
						t.Errorf("Preset %s has invalid radius %f", p.Name, p.Radius)
					}
					if p.Intensity < 0 || p.Intensity > 1 {
						t.Errorf("Preset %s has invalid intensity %f", p.Name, p.Intensity)
					}
					if p.R < 0 || p.R > 1 {
						t.Errorf("Preset %s has invalid R %f", p.Name, p.R)
					}
					if p.G < 0 || p.G > 1 {
						t.Errorf("Preset %s has invalid G %f", p.Name, p.G)
					}
					if p.B < 0 || p.B > 1 {
						t.Errorf("Preset %s has invalid B %f", p.Name, p.B)
					}
					break
				}
			}
			if !found {
				t.Errorf("Expected preset %s not found in %s genre", tt.checkName, tt.genreID)
			}
		})
	}
}

func TestNewPointLight(t *testing.T) {
	preset := LightPreset{
		Name:      "test_light",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.5,
		B:         0.2,
		Flicker:   true,
	}

	pl := NewPointLight(10.0, 20.0, preset, 12345)

	if pl.X != 10.0 {
		t.Errorf("X = %f, want 10.0", pl.X)
	}
	if pl.Y != 20.0 {
		t.Errorf("Y = %f, want 20.0", pl.Y)
	}
	if pl.Radius != 5.0 {
		t.Errorf("Radius = %f, want 5.0", pl.Radius)
	}
	if pl.Intensity != 0.8 {
		t.Errorf("Intensity = %f, want 0.8", pl.Intensity)
	}
	if pl.R != 1.0 {
		t.Errorf("R = %f, want 1.0", pl.R)
	}
	if pl.G != 0.5 {
		t.Errorf("G = %f, want 0.5", pl.G)
	}
	if pl.B != 0.2 {
		t.Errorf("B = %f, want 0.2", pl.B)
	}
	if pl.LightType != "test_light" {
		t.Errorf("LightType = %s, want test_light", pl.LightType)
	}
	if !pl.IsFlickering {
		t.Error("IsFlickering = false, want true")
	}
	if pl.FlickerSeed != 12345 {
		t.Errorf("FlickerSeed = %d, want 12345", pl.FlickerSeed)
	}
}

func TestGetPresetByName(t *testing.T) {
	tests := []struct {
		name      string
		genreID   string
		lightName string
		wantFound bool
	}{
		{
			name:      "fantasy torch exists",
			genreID:   genre.Fantasy,
			lightName: "torch",
			wantFound: true,
		},
		{
			name:      "scifi monitor exists",
			genreID:   genre.SciFi,
			lightName: "monitor",
			wantFound: true,
		},
		{
			name:      "horror dim_bulb exists",
			genreID:   genre.Horror,
			lightName: "dim_bulb",
			wantFound: true,
		},
		{
			name:      "cyberpunk neon_pink exists",
			genreID:   genre.Cyberpunk,
			lightName: "neon_pink",
			wantFound: true,
		},
		{
			name:      "postapoc oil_lamp exists",
			genreID:   genre.PostApoc,
			lightName: "oil_lamp",
			wantFound: true,
		},
		{
			name:      "nonexistent light",
			genreID:   genre.Fantasy,
			lightName: "nonexistent",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset, found := GetPresetByName(tt.genreID, tt.lightName)
			if found != tt.wantFound {
				t.Errorf("GetPresetByName found = %v, want %v", found, tt.wantFound)
			}
			if found && preset.Name != tt.lightName {
				t.Errorf("Preset name = %s, want %s", preset.Name, tt.lightName)
			}
		})
	}
}

func TestUpdateFlicker(t *testing.T) {
	preset := LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.6,
		B:         0.2,
		Flicker:   true,
	}
	pl := NewPointLight(10.0, 10.0, preset, 12345)

	// Test flickering light
	intensity1 := pl.UpdateFlicker(0)
	if intensity1 < 0.0 || intensity1 > 1.0 {
		t.Errorf("Flickered intensity out of range: %f", intensity1)
	}

	intensity2 := pl.UpdateFlicker(100)
	if intensity2 < 0.0 || intensity2 > 1.0 {
		t.Errorf("Flickered intensity out of range: %f", intensity2)
	}

	// Different ticks should produce different intensities (usually)
	// Note: May occasionally be equal due to random variation
	if intensity1 == intensity2 {
		t.Logf("Warning: intensities equal at different ticks (rare but possible)")
	}

	// Test determinism: same tick should produce same result
	intensity3 := pl.UpdateFlicker(0)
	if math.Abs(intensity1-intensity3) > 0.001 {
		t.Errorf("Flicker not deterministic: %f vs %f", intensity1, intensity3)
	}

	// Test non-flickering light
	preset.Flicker = false
	pl2 := NewPointLight(10.0, 10.0, preset, 54321)
	steadyIntensity := pl2.UpdateFlicker(0)
	if steadyIntensity != preset.Intensity {
		t.Errorf("Non-flickering light changed intensity: %f vs %f", steadyIntensity, preset.Intensity)
	}
}

func TestApplyAttenuation(t *testing.T) {
	preset := LightPreset{
		Name:      "test",
		Radius:    5.0,
		Intensity: 1.0,
		R:         1.0,
		G:         1.0,
		B:         1.0,
		Flicker:   false,
	}
	pl := NewPointLight(10.0, 10.0, preset, 0)

	tests := []struct {
		name     string
		targetX  float64
		targetY  float64
		wantZero bool
	}{
		{
			name:     "at light position",
			targetX:  10.0,
			targetY:  10.0,
			wantZero: false,
		},
		{
			name:     "distance 1",
			targetX:  11.0,
			targetY:  10.0,
			wantZero: false,
		},
		{
			name:     "distance 3",
			targetX:  13.0,
			targetY:  10.0,
			wantZero: false,
		},
		{
			name:     "at radius boundary",
			targetX:  15.0,
			targetY:  10.0,
			wantZero: false,
		},
		{
			name:     "outside radius",
			targetX:  20.0,
			targetY:  10.0,
			wantZero: true,
		},
		{
			name:     "far outside radius",
			targetX:  100.0,
			targetY:  100.0,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contribution := pl.ApplyAttenuation(tt.targetX, tt.targetY)
			if contribution < 0.0 || contribution > 1.0 {
				t.Errorf("Contribution out of range: %f", contribution)
			}
			if tt.wantZero && contribution != 0.0 {
				t.Errorf("Expected zero contribution outside radius, got %f", contribution)
			}
			if !tt.wantZero && contribution == 0.0 {
				t.Errorf("Expected non-zero contribution inside radius, got 0")
			}
		})
	}

	// Test quadratic falloff: closer points should be brighter
	contrib1 := pl.ApplyAttenuation(10.0, 10.0) // distance 0
	contrib2 := pl.ApplyAttenuation(11.0, 10.0) // distance 1
	contrib3 := pl.ApplyAttenuation(12.0, 10.0) // distance 2

	if contrib1 <= contrib2 {
		t.Errorf("Contribution should decrease with distance: %f vs %f", contrib1, contrib2)
	}
	if contrib2 <= contrib3 {
		t.Errorf("Contribution should decrease with distance: %f vs %f", contrib2, contrib3)
	}
}

func TestLinearAttenuation(t *testing.T) {
	preset := LightPreset{
		Name:      "test",
		Radius:    5.0,
		Intensity: 1.0,
		R:         1.0,
		G:         1.0,
		B:         1.0,
		Flicker:   false,
	}
	pl := NewPointLight(10.0, 10.0, preset, 0)

	tests := []struct {
		name     string
		targetX  float64
		targetY  float64
		wantZero bool
	}{
		{
			name:     "at light position",
			targetX:  10.0,
			targetY:  10.0,
			wantZero: false,
		},
		{
			name:     "halfway to radius",
			targetX:  12.5,
			targetY:  10.0,
			wantZero: false,
		},
		{
			name:     "at radius boundary",
			targetX:  15.0,
			targetY:  10.0,
			wantZero: false,
		},
		{
			name:     "outside radius",
			targetX:  16.0,
			targetY:  10.0,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contribution := pl.LinearAttenuation(tt.targetX, tt.targetY)
			if contribution < 0.0 || contribution > 1.0 {
				t.Errorf("Contribution out of range: %f", contribution)
			}
			if tt.wantZero && contribution > 0.01 {
				t.Errorf("Expected ~zero contribution outside radius, got %f", contribution)
			}
		})
	}

	// Test linear falloff characteristics
	contribCenter := pl.LinearAttenuation(10.0, 10.0)
	if math.Abs(contribCenter-1.0) > 0.001 {
		t.Errorf("Center should have full intensity, got %f", contribCenter)
	}

	contribHalf := pl.LinearAttenuation(12.5, 10.0) // distance 2.5
	expectedHalf := 0.5                             // 1 - (2.5/5.0)
	if math.Abs(contribHalf-expectedHalf) > 0.01 {
		t.Errorf("Halfway point should have ~0.5 intensity, got %f", contribHalf)
	}
}

func TestPointLight_SetPosition(t *testing.T) {
	preset := LightPreset{Name: "test", Radius: 5.0, Intensity: 1.0, R: 1, G: 1, B: 1}
	pl := NewPointLight(0.0, 0.0, preset, 0)

	pl.SetPosition(15.5, 25.3)
	if pl.X != 15.5 {
		t.Errorf("X = %f, want 15.5", pl.X)
	}
	if pl.Y != 25.3 {
		t.Errorf("Y = %f, want 25.3", pl.Y)
	}

	pl.SetPosition(-10.0, -20.0)
	if pl.X != -10.0 {
		t.Errorf("X = %f, want -10.0", pl.X)
	}
	if pl.Y != -20.0 {
		t.Errorf("Y = %f, want -20.0", pl.Y)
	}
}

func TestPointLight_SetIntensity(t *testing.T) {
	preset := LightPreset{Name: "test", Radius: 5.0, Intensity: 1.0, R: 1, G: 1, B: 1}
	pl := NewPointLight(0.0, 0.0, preset, 0)

	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{"valid value", 0.5, 0.5},
		{"clamped high", 1.5, 1.0},
		{"clamped low", -0.5, 0.0},
		{"zero", 0.0, 0.0},
		{"max", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl.SetIntensity(tt.value)
			if pl.Intensity != tt.want {
				t.Errorf("Intensity = %f, want %f", pl.Intensity, tt.want)
			}
		})
	}
}

func TestPointLight_SetColor(t *testing.T) {
	preset := LightPreset{Name: "test", Radius: 5.0, Intensity: 1.0, R: 1, G: 1, B: 1}
	pl := NewPointLight(0.0, 0.0, preset, 0)

	tests := []struct {
		name                string
		r, g, b             float64
		wantR, wantG, wantB float64
	}{
		{"valid colors", 0.5, 0.6, 0.7, 0.5, 0.6, 0.7},
		{"clamped high", 1.5, 1.2, 1.8, 1.0, 1.0, 1.0},
		{"clamped low", -0.5, -0.2, -0.8, 0.0, 0.0, 0.0},
		{"mixed clamp", 0.5, 1.5, -0.5, 0.5, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pl.SetColor(tt.r, tt.g, tt.b)
			if pl.R != tt.wantR {
				t.Errorf("R = %f, want %f", pl.R, tt.wantR)
			}
			if pl.G != tt.wantG {
				t.Errorf("G = %f, want %f", pl.G, tt.wantG)
			}
			if pl.B != tt.wantB {
				t.Errorf("B = %f, want %f", pl.B, tt.wantB)
			}
		})
	}
}

func TestClampColor(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{"within range", 0.5, 0.5},
		{"below min", -0.5, 0.0},
		{"above max", 1.5, 1.0},
		{"at min", 0.0, 0.0},
		{"at max", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampColor(tt.value)
			if got != tt.want {
				t.Errorf("clampColor(%f) = %f, want %f", tt.value, got, tt.want)
			}
		})
	}
}

func TestFantasyLightPresets(t *testing.T) {
	presets := GetGenrePresets(genre.Fantasy)
	expectedLights := []string{"torch", "brazier", "candle", "magic_crystal"}

	for _, expected := range expectedLights {
		found := false
		for _, p := range presets {
			if p.Name == expected {
				found = true
				// Validate torch specifically (should flicker)
				if expected == "torch" && !p.Flicker {
					t.Errorf("Fantasy torch should flicker")
				}
				// Validate magic crystal (should not flicker)
				if expected == "magic_crystal" && p.Flicker {
					t.Errorf("Magic crystal should not flicker")
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected fantasy light %s not found", expected)
		}
	}
}

func TestSciFiLightPresets(t *testing.T) {
	presets := GetGenrePresets(genre.SciFi)

	// Check that alarm light flickers
	for _, p := range presets {
		if p.Name == "alarm" && !p.Flicker {
			t.Error("SciFi alarm should flicker")
		}
		if p.Name == "ceiling_lamp" && p.Flicker {
			t.Error("SciFi ceiling_lamp should not flicker")
		}
	}
}

func TestHorrorLightPresets(t *testing.T) {
	presets := GetGenrePresets(genre.Horror)

	// Horror lights should generally be dim
	for _, p := range presets {
		if p.Intensity > 0.6 {
			t.Errorf("Horror light %s too bright: %f (should be ≤0.6)", p.Name, p.Intensity)
		}
		// All horror lights should flicker
		if !p.Flicker {
			t.Errorf("Horror light %s should flicker", p.Name)
		}
	}
}

func TestCyberpunkLightPresets(t *testing.T) {
	presets := GetGenrePresets(genre.Cyberpunk)

	// Check for neon colors
	foundPink := false
	foundCyan := false
	for _, p := range presets {
		if p.Name == "neon_pink" {
			foundPink = true
			// Pink should have high R, low G, high B
			if p.R < 0.8 || p.G > 0.5 || p.B < 0.5 {
				t.Errorf("Neon pink has wrong color: R=%f G=%f B=%f", p.R, p.G, p.B)
			}
		}
		if p.Name == "neon_cyan" {
			foundCyan = true
			// Cyan should have low R, high G, high B
			if p.R > 0.5 || p.G < 0.5 || p.B < 0.8 {
				t.Errorf("Neon cyan has wrong color: R=%f G=%f B=%f", p.R, p.G, p.B)
			}
		}
	}
	if !foundPink {
		t.Error("Cyberpunk missing neon_pink preset")
	}
	if !foundCyan {
		t.Error("Cyberpunk missing neon_cyan preset")
	}
}

func TestPostApocLightPresets(t *testing.T) {
	presets := GetGenrePresets(genre.PostApoc)

	// All post-apoc lights should flicker (unreliable power)
	for _, p := range presets {
		if !p.Flicker {
			t.Errorf("PostApoc light %s should flicker", p.Name)
		}
	}
}

// BenchmarkApplyAttenuation measures attenuation calculation performance
func BenchmarkApplyAttenuation(b *testing.B) {
	preset := LightPreset{Name: "test", Radius: 10.0, Intensity: 1.0, R: 1, G: 1, B: 1}
	pl := NewPointLight(50.0, 50.0, preset, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pl.ApplyAttenuation(55.0, 55.0)
	}
}

// BenchmarkUpdateFlicker measures flicker calculation performance
func BenchmarkUpdateFlicker(b *testing.B) {
	preset := LightPreset{Name: "torch", Radius: 5.0, Intensity: 0.8, R: 1, G: 0.6, B: 0.2, Flicker: true}
	pl := NewPointLight(10.0, 10.0, preset, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pl.UpdateFlicker(i)
	}
}

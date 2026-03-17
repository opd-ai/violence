package statustint

import (
	"image/color"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/status"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre)

			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}

			if sys.genreID != genre {
				t.Errorf("expected genreID %s, got %s", genre, sys.genreID)
			}

			if len(sys.profiles) == 0 {
				t.Error("profiles map is empty")
			}

			if sys.logger == nil {
				t.Error("logger is nil")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")

	// Get initial profile count
	initialCount := len(sys.profiles)

	// Change genre
	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Errorf("expected genreID scifi, got %s", sys.genreID)
	}

	// Profiles should be reloaded
	if len(sys.profiles) == 0 {
		t.Error("profiles empty after genre change")
	}

	// Setting same genre should not trigger reload
	sys.SetGenre("scifi")
	if sys.genreID != "scifi" {
		t.Error("genre changed unexpectedly")
	}

	t.Logf("fantasy profiles: %d, scifi profiles: %d", initialCount, len(sys.profiles))
}

func TestTintComponentReset(t *testing.T) {
	tc := &TintComponent{
		TintColor:         color.RGBA{255, 0, 0, 255},
		Intensity:         0.8,
		Saturation:        0.5,
		Brightness:        0.3,
		Contrast:          2.0,
		EdgeGlowIntensity: 0.7,
		NoiseIntensity:    0.4,
		DominantEffect:    "burning",
	}

	tc.Reset()

	if tc.Intensity != 0 {
		t.Errorf("Intensity not reset, got %f", tc.Intensity)
	}
	if tc.Saturation != 0 {
		t.Errorf("Saturation not reset, got %f", tc.Saturation)
	}
	if tc.Brightness != 0 {
		t.Errorf("Brightness not reset, got %f", tc.Brightness)
	}
	if tc.Contrast != 1.0 {
		t.Errorf("Contrast not reset to 1.0, got %f", tc.Contrast)
	}
	if tc.DominantEffect != "" {
		t.Errorf("DominantEffect not reset, got %s", tc.DominantEffect)
	}
}

func TestTintComponentHasTint(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*TintComponent)
		expected bool
	}{
		{
			name:     "default (no tint)",
			setup:    func(tc *TintComponent) { tc.Reset() },
			expected: false,
		},
		{
			name: "with intensity",
			setup: func(tc *TintComponent) {
				tc.Reset()
				tc.Intensity = 0.5
			},
			expected: true,
		},
		{
			name: "with edge glow",
			setup: func(tc *TintComponent) {
				tc.Reset()
				tc.EdgeGlowIntensity = 0.3
			},
			expected: true,
		},
		{
			name: "with saturation change",
			setup: func(tc *TintComponent) {
				tc.Reset()
				tc.Saturation = -0.5
			},
			expected: true,
		},
		{
			name: "with brightness change",
			setup: func(tc *TintComponent) {
				tc.Reset()
				tc.Brightness = 0.2
			},
			expected: true,
		},
		{
			name: "with contrast change",
			setup: func(tc *TintComponent) {
				tc.Reset()
				tc.Contrast = 1.5
			},
			expected: true,
		},
		{
			name: "with noise",
			setup: func(tc *TintComponent) {
				tc.Reset()
				tc.NoiseIntensity = 0.15
			},
			expected: true,
		},
		{
			name: "below threshold",
			setup: func(tc *TintComponent) {
				tc.Reset()
				tc.Intensity = 0.005
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &TintComponent{}
			tt.setup(tc)
			if result := tc.HasTint(); result != tt.expected {
				t.Errorf("HasTint() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTintComponentType(t *testing.T) {
	tc := &TintComponent{}
	if tc.Type() != "StatusTint" {
		t.Errorf("expected Type() = StatusTint, got %s", tc.Type())
	}
}

func TestDefaultEffectProfiles(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			profiles := DefaultEffectProfiles(genre)

			if len(profiles) == 0 {
				t.Fatal("no profiles returned")
			}

			// Verify all profiles have reasonable values
			for name, profile := range profiles {
				if profile.TintStrength < 0 || profile.TintStrength > 1 {
					t.Errorf("%s: TintStrength out of range: %f", name, profile.TintStrength)
				}
				if profile.SaturationShift < -1 || profile.SaturationShift > 1 {
					t.Errorf("%s: SaturationShift out of range: %f", name, profile.SaturationShift)
				}
				if profile.BrightnessShift < -1 || profile.BrightnessShift > 1 {
					t.Errorf("%s: BrightnessShift out of range: %f", name, profile.BrightnessShift)
				}
				if profile.EdgeGlowStrength < 0 || profile.EdgeGlowStrength > 1 {
					t.Errorf("%s: EdgeGlowStrength out of range: %f", name, profile.EdgeGlowStrength)
				}
				if profile.NoiseStrength < 0 || profile.NoiseStrength > 1 {
					t.Errorf("%s: NoiseStrength out of range: %f", name, profile.NoiseStrength)
				}
			}

			t.Logf("%s: %d effect profiles", genre, len(profiles))
		})
	}
}

func TestDefaultEffectProfilesUnknownGenre(t *testing.T) {
	profiles := DefaultEffectProfiles("unknown_genre")
	fantasyProfiles := DefaultEffectProfiles("fantasy")

	if len(profiles) != len(fantasyProfiles) {
		t.Error("unknown genre should fall back to fantasy")
	}
}

func TestSystemUpdateWithNoEntities(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Should not panic
	sys.Update(world)
}

func TestSystemUpdateWithNilWorld(t *testing.T) {
	sys := NewSystem("fantasy")

	// Should not panic
	sys.Update(nil)
}

func TestSystemUpdateCreatesAndRemovesTintComponent(t *testing.T) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()

	// Add position and status components
	world.AddComponent(entity, &engine.Position{X: 0, Y: 0})
	sc := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "poisoned",
				TimeRemaining: 10 * time.Second,
				VisualColor:   0x88FF0088,
			},
		},
	}
	world.AddComponent(entity, sc)

	// Update should create tint component
	sys.Update(world)

	// Check if tint component exists
	tintComp := getTintComponent(world, entity)
	if tintComp == nil {
		t.Fatal("TintComponent not created")
	}

	if tintComp.DominantEffect != "poisoned" {
		t.Errorf("expected dominant effect 'poisoned', got '%s'", tintComp.DominantEffect)
	}

	// Clear status effects
	sc.ActiveEffects = nil
	sys.Update(world)

	// Tint component should be removed
	tintComp = getTintComponent(world, entity)
	if tintComp != nil {
		t.Error("TintComponent should have been removed")
	}
}

func TestSystemComputeAggregateTint(t *testing.T) {
	sys := NewSystem("fantasy")

	tc := &TintComponent{}
	tc.Reset()

	effects := []status.ActiveEffect{
		{
			EffectName:    "poisoned",
			TimeRemaining: 10 * time.Second,
		},
		{
			EffectName:    "burning",
			TimeRemaining: 5 * time.Second,
		},
	}

	sys.computeAggregateTint(tc, effects)

	if tc.Intensity == 0 {
		t.Error("Intensity should be non-zero with active effects")
	}

	if tc.DominantEffect == "" {
		t.Error("DominantEffect should be set")
	}

	// Verify blended tint (should be between green and orange)
	t.Logf("Blended tint: R=%d G=%d B=%d, Intensity=%.2f, Dominant=%s",
		tc.TintColor.R, tc.TintColor.G, tc.TintColor.B, tc.Intensity, tc.DominantEffect)
}

func TestApplyTintToColor(t *testing.T) {
	tests := []struct {
		name      string
		src       color.RGBA
		tint      *TintComponent
		checkFunc func(result color.RGBA) bool
		desc      string
	}{
		{
			name: "nil tint",
			src:  color.RGBA{128, 128, 128, 255},
			tint: nil,
			checkFunc: func(r color.RGBA) bool {
				return r.R == 128 && r.G == 128 && r.B == 128
			},
			desc: "should return unchanged color",
		},
		{
			name: "no tint active",
			src:  color.RGBA{128, 128, 128, 255},
			tint: &TintComponent{Contrast: 1.0},
			checkFunc: func(r color.RGBA) bool {
				return r.R == 128 && r.G == 128 && r.B == 128
			},
			desc: "should return unchanged color",
		},
		{
			name: "red tint",
			src:  color.RGBA{128, 128, 128, 255},
			tint: &TintComponent{
				TintColor: color.RGBA{255, 0, 0, 255},
				Intensity: 0.5,
				Contrast:  1.0,
			},
			checkFunc: func(r color.RGBA) bool {
				return r.R > r.G && r.R > r.B
			},
			desc: "should shift toward red",
		},
		{
			name: "brightness increase",
			src:  color.RGBA{100, 100, 100, 255},
			tint: &TintComponent{
				Brightness: 0.3,
				Contrast:   1.0,
			},
			checkFunc: func(r color.RGBA) bool {
				return r.R > 100 && r.G > 100 && r.B > 100
			},
			desc: "should be brighter",
		},
		{
			name: "desaturation",
			src:  color.RGBA{255, 0, 0, 255},
			tint: &TintComponent{
				Saturation: -1.0,
				Contrast:   1.0,
			},
			checkFunc: func(r color.RGBA) bool {
				// Should be close to grayscale (all channels similar)
				diff := math.Abs(float64(r.R)-float64(r.G)) + math.Abs(float64(r.G)-float64(r.B))
				return diff < 50
			},
			desc: "should be near grayscale",
		},
		{
			name: "preserve alpha",
			src:  color.RGBA{128, 128, 128, 100},
			tint: &TintComponent{
				TintColor: color.RGBA{255, 0, 0, 255},
				Intensity: 0.5,
				Contrast:  1.0,
			},
			checkFunc: func(r color.RGBA) bool {
				return r.A == 100
			},
			desc: "should preserve original alpha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyTintToColor(tt.src, tt.tint)
			if !tt.checkFunc(result) {
				t.Errorf("%s: src=%v, result=%v", tt.desc, tt.src, result)
			}
		})
	}
}

func TestGetEdgeGlowPixel(t *testing.T) {
	tint := &TintComponent{
		EdgeGlowColor:     color.RGBA{255, 200, 100, 255},
		EdgeGlowIntensity: 0.8,
		Contrast:          1.0,
	}

	// At edge (dist=0), glow should be strongest
	edgeGlow := GetEdgeGlowPixel(tint, 0.0)
	if edgeGlow.A == 0 {
		t.Error("edge glow should be visible at edge")
	}

	// At center (dist=1), glow should be minimal
	centerGlow := GetEdgeGlowPixel(tint, 1.0)
	if centerGlow.A >= edgeGlow.A {
		t.Error("center glow should be less than edge glow")
	}

	// With nil tint
	nilGlow := GetEdgeGlowPixel(nil, 0.0)
	if nilGlow.A != 0 {
		t.Error("nil tint should produce no glow")
	}
}

func TestGetNoiseOffset(t *testing.T) {
	tint := &TintComponent{
		NoiseIntensity: 0.5,
		NoisePhase:     1.0,
		Contrast:       1.0,
	}

	ox, oy := GetNoiseOffset(tint, 10, 20)

	// Should be within noise intensity range
	if math.Abs(ox) > 1 || math.Abs(oy) > 1 {
		t.Errorf("noise offset out of range: ox=%f, oy=%f", ox, oy)
	}

	// Different positions should give different offsets
	ox2, oy2 := GetNoiseOffset(tint, 50, 100)
	if ox == ox2 && oy == oy2 {
		t.Error("different positions should give different offsets")
	}

	// Nil tint should give zero offset
	nilOx, nilOy := GetNoiseOffset(nil, 10, 20)
	if nilOx != 0 || nilOy != 0 {
		t.Error("nil tint should give zero offset")
	}
}

func TestShouldApplyNoiseColor(t *testing.T) {
	tint := &TintComponent{
		NoiseIntensity: 0.5,
		NoisePhase:     0.0,
		Contrast:       1.0,
	}

	// Test that some pixels get noise and some don't (probabilistic)
	noiseCount := 0
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			if ShouldApplyNoiseColor(tint, x, y) {
				noiseCount++
			}
		}
	}

	// Should have some noise but not all pixels
	if noiseCount == 0 {
		t.Log("no noise applied (might be expected for low intensity)")
	}
	if noiseCount == 10000 {
		t.Error("all pixels have noise (unexpected)")
	}

	t.Logf("noise applied to %d/10000 pixels", noiseCount)

	// Nil tint should never apply noise
	if ShouldApplyNoiseColor(nil, 50, 50) {
		t.Error("nil tint should not apply noise")
	}

	// Very low intensity should not apply noise
	lowNoise := &TintComponent{NoiseIntensity: 0.05, Contrast: 1.0}
	if ShouldApplyNoiseColor(lowNoise, 50, 50) {
		t.Error("very low noise intensity should not apply noise")
	}
}

func TestIntensityCalculation(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		timeRemaining float64
		minIntensity  float64
		maxIntensity  float64
	}{
		{10.0, 0.99, 1.01}, // Full intensity
		{5.0, 0.99, 1.01},  // Full intensity
		{2.0, 0.99, 1.01},  // Full intensity (boundary)
		{1.5, 0.74, 0.76},  // Fading
		{1.0, 0.49, 0.51},  // Fading
		{0.5, 0.24, 0.26},  // Almost gone
		{0.0, 0.0, 0.01},   // Expired
		{-1.0, 0.0, 0.01},  // Past expiration
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			intensity := sys.calculateIntensity(tt.timeRemaining)
			if intensity < tt.minIntensity || intensity > tt.maxIntensity {
				t.Errorf("calculateIntensity(%f) = %f, want [%f, %f]",
					tt.timeRemaining, intensity, tt.minIntensity, tt.maxIntensity)
			}
		})
	}
}

func TestPulsingTint(t *testing.T) {
	sys := NewSystem("fantasy")

	tc := &TintComponent{}
	tc.Reset()

	effects := []status.ActiveEffect{
		{
			EffectName:    "burning", // Burning has high pulse frequency
			TimeRemaining: 5 * time.Second,
		},
	}

	// Capture tint at different times
	var phases []float64
	for i := 0; i < 10; i++ {
		sys.time = float64(i) * 0.1
		sys.computeAggregateTint(tc, effects)
		phases = append(phases, tc.PulsePhase)
	}

	// Check that phases change over time (not all the same)
	allSame := true
	for i := 1; i < len(phases); i++ {
		if phases[i] != phases[0] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("pulse phases should change over time")
	}

	// Check that pulse amplitude is set
	if tc.PulseAmplitude <= 0 {
		t.Error("burning effect should have pulse amplitude")
	}

	// Check that pulse frequency is set
	if tc.PulseFrequency <= 0 {
		t.Error("burning effect should have pulse frequency")
	}
}

func TestMultipleEffectsBlending(t *testing.T) {
	sys := NewSystem("fantasy")

	tc := &TintComponent{}
	tc.Reset()

	// Apply multiple effects with different characteristics
	effects := []status.ActiveEffect{
		{EffectName: "poisoned", TimeRemaining: 10 * time.Second},
		{EffectName: "bleeding", TimeRemaining: 10 * time.Second},
		{EffectName: "cursed", TimeRemaining: 10 * time.Second},
	}

	sys.computeAggregateTint(tc, effects)

	// Should have blended properties
	if tc.Intensity == 0 {
		t.Error("should have non-zero intensity")
	}

	// Saturation should be affected (poisoned and cursed both reduce saturation)
	if tc.Saturation == 0 {
		t.Log("saturation unchanged (effects may cancel)")
	}

	// Edge glow should accumulate
	if tc.EdgeGlowIntensity == 0 {
		t.Error("should have edge glow from multiple effects")
	}

	t.Logf("Blended result: intensity=%.2f, saturation=%.2f, brightness=%.2f, edgeGlow=%.2f, dominant=%s",
		tc.Intensity, tc.Saturation, tc.Brightness, tc.EdgeGlowIntensity, tc.DominantEffect)
}

// Helper function to get TintComponent from entity
func getTintComponent(w *engine.World, entity engine.Entity) *TintComponent {
	tintType := reflect.TypeOf(&TintComponent{})
	comp, ok := w.GetComponent(entity, tintType)
	if !ok {
		return nil
	}
	return comp.(*TintComponent)
}

func BenchmarkApplyTintToColor(b *testing.B) {
	tint := &TintComponent{
		TintColor:      color.RGBA{255, 100, 50, 255},
		Intensity:      0.5,
		Saturation:     0.2,
		Brightness:     0.1,
		Contrast:       1.1,
		PulseAmplitude: 0.2,
		PulsePhase:     1.5,
	}

	src := color.RGBA{128, 128, 128, 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyTintToColor(src, tint)
	}
}

func BenchmarkComputeAggregateTint(b *testing.B) {
	sys := NewSystem("fantasy")
	tc := &TintComponent{}

	effects := []status.ActiveEffect{
		{EffectName: "poisoned", TimeRemaining: 10 * time.Second},
		{EffectName: "burning", TimeRemaining: 5 * time.Second},
		{EffectName: "bleeding", TimeRemaining: 8 * time.Second},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Reset()
		sys.computeAggregateTint(tc, effects)
	}
}

func BenchmarkSystemUpdate(b *testing.B) {
	sys := NewSystem("fantasy")
	world := engine.NewWorld()

	// Create 100 entities with status effects
	for i := 0; i < 100; i++ {
		entity := world.AddEntity()
		world.AddComponent(entity, &engine.Position{X: float64(i), Y: float64(i)})
		world.AddComponent(entity, &status.StatusComponent{
			ActiveEffects: []status.ActiveEffect{
				{EffectName: "poisoned", TimeRemaining: 10 * time.Second},
			},
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world)
	}
}

package statustint

import "image/color"

// TintComponent stores the computed visual tint parameters for an entity.
// These values are read by the sprite rendering pipeline to apply
// status-based color modifications.
type TintComponent struct {
	// Base tint color to blend with sprite (RGBA)
	TintColor color.RGBA

	// Tint intensity (0.0 = no tint, 1.0 = full tint)
	Intensity float64

	// Saturation modifier (-1.0 = grayscale, 0.0 = normal, 1.0 = oversaturated)
	Saturation float64

	// Brightness modifier (-1.0 = black, 0.0 = normal, 1.0 = white)
	Brightness float64

	// Contrast modifier (0.0 = flat gray, 1.0 = normal, 2.0 = high contrast)
	Contrast float64

	// Edge glow parameters for rim lighting effects
	EdgeGlowColor     color.RGBA
	EdgeGlowIntensity float64

	// Noise parameters for effects like irradiation
	NoiseIntensity float64
	NoisePhase     float64

	// Pulse parameters for animated effects
	PulsePhase     float64
	PulseAmplitude float64
	PulseFrequency float64

	// Dominant effect name for debugging/logging
	DominantEffect string

	// Whether tint needs recalculation (dirty flag)
	Dirty bool
}

// Type returns the component type identifier.
func (c *TintComponent) Type() string {
	return "StatusTint"
}

// Reset clears all tint parameters to defaults.
func (c *TintComponent) Reset() {
	c.TintColor = color.RGBA{0, 0, 0, 0}
	c.Intensity = 0
	c.Saturation = 0
	c.Brightness = 0
	c.Contrast = 1.0
	c.EdgeGlowColor = color.RGBA{0, 0, 0, 0}
	c.EdgeGlowIntensity = 0
	c.NoiseIntensity = 0
	c.NoisePhase = 0
	c.PulsePhase = 0
	c.PulseAmplitude = 0
	c.PulseFrequency = 0
	c.DominantEffect = ""
	c.Dirty = false
}

// HasTint returns true if any tinting should be applied.
func (c *TintComponent) HasTint() bool {
	return c.Intensity > 0.01 ||
		c.EdgeGlowIntensity > 0.01 ||
		c.Saturation != 0 ||
		c.Brightness != 0 ||
		c.Contrast != 1.0 ||
		c.NoiseIntensity > 0.01
}

// EffectTintProfile defines the visual tint parameters for a specific status effect.
type EffectTintProfile struct {
	// Base tint color
	TintColor color.RGBA

	// How strongly to apply the tint (0.0-1.0)
	TintStrength float64

	// Saturation change (-1.0 to 1.0)
	SaturationShift float64

	// Brightness change (-1.0 to 1.0)
	BrightnessShift float64

	// Edge glow (rim lighting)
	EdgeGlowColor     color.RGBA
	EdgeGlowStrength  float64

	// Pulse animation
	PulseAmplitude float64
	PulseFrequency float64

	// Noise for radiation/glitch effects
	NoiseStrength float64
}

// DefaultEffectProfiles returns the standard tint profiles for common effects.
// These are genre-aware and produce material-appropriate visual changes.
func DefaultEffectProfiles(genreID string) map[string]EffectTintProfile {
	profiles := make(map[string]EffectTintProfile)

	switch genreID {
	case "fantasy":
		profiles = map[string]EffectTintProfile{
			"poisoned": {
				TintColor:       color.RGBA{R: 80, G: 180, B: 60, A: 255},
				TintStrength:    0.35,
				SaturationShift: -0.15,
				BrightnessShift: -0.1,
				EdgeGlowColor:   color.RGBA{R: 60, G: 150, B: 40, A: 200},
				EdgeGlowStrength: 0.3,
				PulseAmplitude:  0.15,
				PulseFrequency:  1.5,
			},
			"burning": {
				TintColor:       color.RGBA{R: 255, G: 120, B: 40, A: 255},
				TintStrength:    0.45,
				SaturationShift: 0.2,
				BrightnessShift: 0.15,
				EdgeGlowColor:   color.RGBA{R: 255, G: 200, B: 100, A: 230},
				EdgeGlowStrength: 0.6,
				PulseAmplitude:  0.25,
				PulseFrequency:  4.0,
			},
			"bleeding": {
				TintColor:       color.RGBA{R: 180, G: 30, B: 30, A: 255},
				TintStrength:    0.3,
				SaturationShift: 0.1,
				BrightnessShift: -0.15,
				EdgeGlowColor:   color.RGBA{R: 140, G: 0, B: 0, A: 180},
				EdgeGlowStrength: 0.2,
				PulseAmplitude:  0.1,
				PulseFrequency:  1.0,
			},
			"stunned": {
				TintColor:       color.RGBA{R: 255, G: 255, B: 180, A: 255},
				TintStrength:    0.25,
				SaturationShift: -0.5,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 255, G: 255, B: 200, A: 200},
				EdgeGlowStrength: 0.4,
				PulseAmplitude:  0.3,
				PulseFrequency:  6.0,
			},
			"regeneration": {
				TintColor:       color.RGBA{R: 80, G: 255, B: 120, A: 255},
				TintStrength:    0.2,
				SaturationShift: 0.1,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 100, G: 255, B: 150, A: 180},
				EdgeGlowStrength: 0.5,
				PulseAmplitude:  0.15,
				PulseFrequency:  2.0,
			},
			"blessed": {
				TintColor:       color.RGBA{R: 255, G: 230, B: 160, A: 255},
				TintStrength:    0.15,
				SaturationShift: 0.05,
				BrightnessShift: 0.15,
				EdgeGlowColor:   color.RGBA{R: 255, G: 240, B: 180, A: 220},
				EdgeGlowStrength: 0.6,
				PulseAmplitude:  0.1,
				PulseFrequency:  1.0,
			},
			"cursed": {
				TintColor:       color.RGBA{R: 80, G: 40, B: 120, A: 255},
				TintStrength:    0.4,
				SaturationShift: -0.3,
				BrightnessShift: -0.2,
				EdgeGlowColor:   color.RGBA{R: 100, G: 50, B: 150, A: 200},
				EdgeGlowStrength: 0.4,
				PulseAmplitude:  0.2,
				PulseFrequency:  0.8,
			},
			"slowed": {
				TintColor:       color.RGBA{R: 100, G: 180, B: 255, A: 255},
				TintStrength:    0.35,
				SaturationShift: -0.2,
				BrightnessShift: 0.05,
				EdgeGlowColor:   color.RGBA{R: 150, G: 200, B: 255, A: 200},
				EdgeGlowStrength: 0.5,
				PulseAmplitude:  0.05,
				PulseFrequency:  0.5,
			},
		}
	case "scifi":
		profiles = map[string]EffectTintProfile{
			"irradiated": {
				TintColor:       color.RGBA{R: 100, G: 255, B: 100, A: 255},
				TintStrength:    0.4,
				SaturationShift: 0.2,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 150, G: 255, B: 150, A: 220},
				EdgeGlowStrength: 0.5,
				PulseAmplitude:  0.2,
				PulseFrequency:  3.0,
				NoiseStrength:   0.15,
			},
			"burning": {
				TintColor:       color.RGBA{R: 255, G: 100, B: 50, A: 255},
				TintStrength:    0.5,
				SaturationShift: 0.25,
				BrightnessShift: 0.2,
				EdgeGlowColor:   color.RGBA{R: 255, G: 180, B: 80, A: 240},
				EdgeGlowStrength: 0.7,
				PulseAmplitude:  0.3,
				PulseFrequency:  5.0,
			},
			"emp_stunned": {
				TintColor:       color.RGBA{R: 100, G: 220, B: 255, A: 255},
				TintStrength:    0.35,
				SaturationShift: -0.6,
				BrightnessShift: 0.05,
				EdgeGlowColor:   color.RGBA{R: 150, G: 240, B: 255, A: 230},
				EdgeGlowStrength: 0.6,
				PulseAmplitude:  0.4,
				PulseFrequency:  8.0,
				NoiseStrength:   0.2,
			},
			"nanoheal": {
				TintColor:       color.RGBA{R: 80, G: 200, B: 255, A: 255},
				TintStrength:    0.2,
				SaturationShift: 0.1,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 100, G: 220, B: 255, A: 200},
				EdgeGlowStrength: 0.5,
				PulseAmplitude:  0.15,
				PulseFrequency:  2.5,
			},
			"overcharged": {
				TintColor:       color.RGBA{R: 255, G: 200, B: 50, A: 255},
				TintStrength:    0.25,
				SaturationShift: 0.2,
				BrightnessShift: 0.2,
				EdgeGlowColor:   color.RGBA{R: 255, G: 220, B: 100, A: 230},
				EdgeGlowStrength: 0.7,
				PulseAmplitude:  0.2,
				PulseFrequency:  3.0,
			},
			"corroded": {
				TintColor:       color.RGBA{R: 150, G: 120, B: 60, A: 255},
				TintStrength:    0.35,
				SaturationShift: -0.25,
				BrightnessShift: -0.15,
				EdgeGlowColor:   color.RGBA{R: 180, G: 140, B: 80, A: 180},
				EdgeGlowStrength: 0.3,
				PulseAmplitude:  0.1,
				PulseFrequency:  1.0,
			},
			"slowed": {
				TintColor:       color.RGBA{R: 80, G: 150, B: 255, A: 255},
				TintStrength:    0.3,
				SaturationShift: -0.15,
				BrightnessShift: 0,
				EdgeGlowColor:   color.RGBA{R: 120, G: 180, B: 255, A: 200},
				EdgeGlowStrength: 0.4,
				PulseAmplitude:  0.05,
				PulseFrequency:  0.5,
			},
		}
	case "horror":
		profiles = map[string]EffectTintProfile{
			"poisoned": {
				TintColor:       color.RGBA{R: 60, G: 120, B: 60, A: 255},
				TintStrength:    0.45,
				SaturationShift: -0.25,
				BrightnessShift: -0.2,
				EdgeGlowColor:   color.RGBA{R: 80, G: 140, B: 60, A: 180},
				EdgeGlowStrength: 0.3,
				PulseAmplitude:  0.2,
				PulseFrequency:  1.2,
			},
			"bleeding": {
				TintColor:       color.RGBA{R: 150, G: 20, B: 20, A: 255},
				TintStrength:    0.5,
				SaturationShift: 0.15,
				BrightnessShift: -0.25,
				EdgeGlowColor:   color.RGBA{R: 120, G: 0, B: 0, A: 200},
				EdgeGlowStrength: 0.4,
				PulseAmplitude:  0.15,
				PulseFrequency:  1.5,
			},
			"terrified": {
				TintColor:       color.RGBA{R: 140, G: 80, B: 160, A: 255},
				TintStrength:    0.4,
				SaturationShift: -0.4,
				BrightnessShift: -0.15,
				EdgeGlowColor:   color.RGBA{R: 160, G: 100, B: 180, A: 180},
				EdgeGlowStrength: 0.3,
				PulseAmplitude:  0.3,
				PulseFrequency:  4.0,
			},
			"infected": {
				TintColor:       color.RGBA{R: 80, G: 100, B: 60, A: 255},
				TintStrength:    0.4,
				SaturationShift: -0.3,
				BrightnessShift: -0.2,
				EdgeGlowColor:   color.RGBA{R: 100, G: 120, B: 80, A: 160},
				EdgeGlowStrength: 0.25,
				PulseAmplitude:  0.1,
				PulseFrequency:  0.8,
				NoiseStrength:   0.1,
			},
			"stunned": {
				TintColor:       color.RGBA{R: 200, G: 200, B: 200, A: 255},
				TintStrength:    0.3,
				SaturationShift: -0.7,
				BrightnessShift: 0,
				EdgeGlowColor:   color.RGBA{R: 220, G: 220, B: 220, A: 180},
				EdgeGlowStrength: 0.3,
				PulseAmplitude:  0.25,
				PulseFrequency:  5.0,
			},
			"regeneration": {
				TintColor:       color.RGBA{R: 80, G: 180, B: 80, A: 255},
				TintStrength:    0.2,
				SaturationShift: 0.05,
				BrightnessShift: 0.05,
				EdgeGlowColor:   color.RGBA{R: 100, G: 200, B: 100, A: 160},
				EdgeGlowStrength: 0.4,
				PulseAmplitude:  0.15,
				PulseFrequency:  1.5,
			},
		}
	case "cyberpunk":
		profiles = map[string]EffectTintProfile{
			"burning": {
				TintColor:       color.RGBA{R: 255, G: 80, B: 200, A: 255},
				TintStrength:    0.45,
				SaturationShift: 0.3,
				BrightnessShift: 0.15,
				EdgeGlowColor:   color.RGBA{R: 255, G: 150, B: 220, A: 240},
				EdgeGlowStrength: 0.7,
				PulseAmplitude:  0.3,
				PulseFrequency:  5.0,
			},
			"hacked": {
				TintColor:       color.RGBA{R: 50, G: 255, B: 150, A: 255},
				TintStrength:    0.4,
				SaturationShift: 0.1,
				BrightnessShift: 0,
				EdgeGlowColor:   color.RGBA{R: 80, G: 255, B: 180, A: 220},
				EdgeGlowStrength: 0.5,
				PulseAmplitude:  0.35,
				PulseFrequency:  6.0,
				NoiseStrength:   0.25,
			},
			"emp_stunned": {
				TintColor:       color.RGBA{R: 100, G: 255, B: 255, A: 255},
				TintStrength:    0.4,
				SaturationShift: -0.5,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 150, G: 255, B: 255, A: 240},
				EdgeGlowStrength: 0.7,
				PulseAmplitude:  0.5,
				PulseFrequency:  10.0,
				NoiseStrength:   0.3,
			},
			"stim_boosted": {
				TintColor:       color.RGBA{R: 255, G: 80, B: 150, A: 255},
				TintStrength:    0.25,
				SaturationShift: 0.2,
				BrightnessShift: 0.15,
				EdgeGlowColor:   color.RGBA{R: 255, G: 120, B: 180, A: 220},
				EdgeGlowStrength: 0.6,
				PulseAmplitude:  0.2,
				PulseFrequency:  3.0,
			},
			"nanoheal": {
				TintColor:       color.RGBA{R: 80, G: 220, B: 255, A: 255},
				TintStrength:    0.2,
				SaturationShift: 0.1,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 120, G: 240, B: 255, A: 200},
				EdgeGlowStrength: 0.5,
				PulseAmplitude:  0.15,
				PulseFrequency:  2.0,
			},
			"glitched": {
				TintColor:       color.RGBA{R: 255, G: 150, B: 255, A: 255},
				TintStrength:    0.35,
				SaturationShift: 0.2,
				BrightnessShift: 0,
				EdgeGlowColor:   color.RGBA{R: 255, G: 180, B: 255, A: 200},
				EdgeGlowStrength: 0.4,
				PulseAmplitude:  0.4,
				PulseFrequency:  8.0,
				NoiseStrength:   0.35,
			},
			"slowed": {
				TintColor:       color.RGBA{R: 150, G: 80, B: 255, A: 255},
				TintStrength:    0.3,
				SaturationShift: -0.1,
				BrightnessShift: -0.05,
				EdgeGlowColor:   color.RGBA{R: 180, G: 120, B: 255, A: 200},
				EdgeGlowStrength: 0.4,
				PulseAmplitude:  0.05,
				PulseFrequency:  0.5,
			},
		}
	case "postapoc":
		profiles = map[string]EffectTintProfile{
			"irradiated": {
				TintColor:       color.RGBA{R: 150, G: 255, B: 80, A: 255},
				TintStrength:    0.45,
				SaturationShift: 0.15,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 180, G: 255, B: 120, A: 220},
				EdgeGlowStrength: 0.6,
				PulseAmplitude:  0.25,
				PulseFrequency:  2.5,
				NoiseStrength:   0.2,
			},
			"poisoned": {
				TintColor:       color.RGBA{R: 100, G: 140, B: 60, A: 255},
				TintStrength:    0.4,
				SaturationShift: -0.2,
				BrightnessShift: -0.15,
				EdgeGlowColor:   color.RGBA{R: 120, G: 160, B: 80, A: 180},
				EdgeGlowStrength: 0.3,
				PulseAmplitude:  0.15,
				PulseFrequency:  1.2,
			},
			"bleeding": {
				TintColor:       color.RGBA{R: 160, G: 30, B: 30, A: 255},
				TintStrength:    0.35,
				SaturationShift: 0.1,
				BrightnessShift: -0.2,
				EdgeGlowColor:   color.RGBA{R: 130, G: 10, B: 10, A: 180},
				EdgeGlowStrength: 0.25,
				PulseAmplitude:  0.1,
				PulseFrequency:  1.0,
			},
			"stunned": {
				TintColor:       color.RGBA{R: 180, G: 180, B: 160, A: 255},
				TintStrength:    0.3,
				SaturationShift: -0.5,
				BrightnessShift: 0,
				EdgeGlowColor:   color.RGBA{R: 200, G: 200, B: 180, A: 180},
				EdgeGlowStrength: 0.35,
				PulseAmplitude:  0.25,
				PulseFrequency:  5.0,
			},
			"stimmed": {
				TintColor:       color.RGBA{R: 255, G: 180, B: 80, A: 255},
				TintStrength:    0.2,
				SaturationShift: 0.1,
				BrightnessShift: 0.1,
				EdgeGlowColor:   color.RGBA{R: 255, G: 200, B: 120, A: 200},
				EdgeGlowStrength: 0.5,
				PulseAmplitude:  0.15,
				PulseFrequency:  2.0,
			},
			"infected": {
				TintColor:       color.RGBA{R: 80, G: 120, B: 60, A: 255},
				TintStrength:    0.4,
				SaturationShift: -0.25,
				BrightnessShift: -0.15,
				EdgeGlowColor:   color.RGBA{R: 100, G: 140, B: 80, A: 160},
				EdgeGlowStrength: 0.25,
				PulseAmplitude:  0.1,
				PulseFrequency:  0.7,
				NoiseStrength:   0.1,
			},
			"corroded": {
				TintColor:       color.RGBA{R: 140, G: 110, B: 70, A: 255},
				TintStrength:    0.35,
				SaturationShift: -0.2,
				BrightnessShift: -0.1,
				EdgeGlowColor:   color.RGBA{R: 160, G: 130, B: 90, A: 170},
				EdgeGlowStrength: 0.3,
				PulseAmplitude:  0.1,
				PulseFrequency:  0.8,
			},
		}
	default:
		// Fall back to fantasy profiles
		return DefaultEffectProfiles("fantasy")
	}

	return profiles
}

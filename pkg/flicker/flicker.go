package flicker

import (
	"math"

	"github.com/opd-ai/violence/pkg/rng"
)

// FlickerParams defines the physics-based flicker behavior for a light source.
type FlickerParams struct {
	// Base light source identification
	LightType string  // e.g., "torch", "candle", "fire_barrel"
	Seed      int64   // For deterministic randomness
	BaseR     float64 // Base red component (0.0-1.0)
	BaseG     float64 // Base green component (0.0-1.0)
	BaseB     float64 // Base blue component (0.0-1.0)

	// Low-frequency sway parameters (main flame body movement)
	SwayFrequency  float64 // Hz (typically 0.3-0.8)
	SwayAmplitude  float64 // Intensity variation (typically 0.10-0.25)
	SwayPhaseShift float64 // Randomized phase offset

	// Mid-frequency oscillation parameters (rapid brightness fluctuations)
	OscillationFrequency float64 // Hz (typically 2-5)
	OscillationAmplitude float64 // Intensity variation (typically 0.05-0.15)

	// High-frequency turbulence parameters (visual shimmer)
	TurbulenceFrequency float64 // Hz (typically 15-30)
	TurbulenceAmplitude float64 // Intensity variation (typically 0.02-0.08)

	// Guttering event parameters (sudden dips)
	GutterProbability float64 // Chance per second of guttering event (typically 0.1-0.5)
	GutterDepth       float64 // How much intensity drops during gutter (typically 0.3-0.6)
	GutterDuration    float64 // Seconds the gutter lasts (typically 0.1-0.3)

	// Color temperature variation
	TempVariationRange float64 // How much color shifts with intensity (0.0-0.3)
	TempWarmBias       float64 // Bias toward warm (positive) or cool (negative)

	// Current gutter state (mutable)
	InGutter        bool
	GutterStartTick int
	GutterLength    int // ticks
}

// Type returns the component type identifier.
func (p *FlickerParams) Type() string {
	return "flicker_params"
}

// System manages realistic flame flicker calculations.
type System struct {
	genre   string
	presets map[string]FlickerParams
}

// NewSystem creates a flicker system with genre-specific presets.
func NewSystem(genre string) *System {
	sys := &System{
		genre:   genre,
		presets: make(map[string]FlickerParams),
	}
	sys.initializePresets()
	return sys
}

// initializePresets populates genre-specific flicker configurations.
func (s *System) initializePresets() {
	switch s.genre {
	case "fantasy":
		s.presets = map[string]FlickerParams{
			"torch": {
				SwayFrequency:        0.5,
				SwayAmplitude:        0.18,
				OscillationFrequency: 3.5,
				OscillationAmplitude: 0.10,
				TurbulenceFrequency:  20.0,
				TurbulenceAmplitude:  0.04,
				GutterProbability:    0.15,
				GutterDepth:          0.35,
				GutterDuration:       0.15,
				TempVariationRange:   0.15,
				TempWarmBias:         0.1,
			},
			"brazier": {
				SwayFrequency:        0.35,
				SwayAmplitude:        0.12,
				OscillationFrequency: 2.5,
				OscillationAmplitude: 0.08,
				TurbulenceFrequency:  15.0,
				TurbulenceAmplitude:  0.05,
				GutterProbability:    0.08,
				GutterDepth:          0.25,
				GutterDuration:       0.12,
				TempVariationRange:   0.12,
				TempWarmBias:         0.15,
			},
			"candle": {
				SwayFrequency:        0.8,
				SwayAmplitude:        0.25,
				OscillationFrequency: 5.0,
				OscillationAmplitude: 0.15,
				TurbulenceFrequency:  25.0,
				TurbulenceAmplitude:  0.06,
				GutterProbability:    0.25,
				GutterDepth:          0.50,
				GutterDuration:       0.20,
				TempVariationRange:   0.20,
				TempWarmBias:         0.05,
			},
			"magic_crystal": {
				SwayFrequency:        0.15,
				SwayAmplitude:        0.05,
				OscillationFrequency: 1.0,
				OscillationAmplitude: 0.03,
				TurbulenceFrequency:  8.0,
				TurbulenceAmplitude:  0.02,
				GutterProbability:    0.0,
				GutterDepth:          0.0,
				GutterDuration:       0.0,
				TempVariationRange:   0.25,
				TempWarmBias:         -0.2, // Magic is cooler
			},
		}
	case "scifi":
		s.presets = map[string]FlickerParams{
			"monitor": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 0.5,
				OscillationAmplitude: 0.02,
				TurbulenceFrequency:  0.0,
				TurbulenceAmplitude:  0.0,
				GutterProbability:    0.01,
				GutterDepth:          0.1,
				GutterDuration:       0.05,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
			"ceiling_lamp": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 0.0,
				OscillationAmplitude: 0.0,
				TurbulenceFrequency:  0.0,
				TurbulenceAmplitude:  0.0,
				GutterProbability:    0.0,
				GutterDepth:          0.0,
				GutterDuration:       0.0,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
			"alarm": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 2.0, // Strobe frequency
				OscillationAmplitude: 0.5, // Full on/off strobe
				TurbulenceFrequency:  0.0,
				TurbulenceAmplitude:  0.0,
				GutterProbability:    0.0,
				GutterDepth:          0.0,
				GutterDuration:       0.0,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
			"console": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 0.3,
				OscillationAmplitude: 0.05,
				TurbulenceFrequency:  0.0,
				TurbulenceAmplitude:  0.0,
				GutterProbability:    0.02,
				GutterDepth:          0.15,
				GutterDuration:       0.08,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
		}
	case "horror":
		s.presets = map[string]FlickerParams{
			"dim_bulb": {
				SwayFrequency:        0.3,
				SwayAmplitude:        0.20,
				OscillationFrequency: 4.0,
				OscillationAmplitude: 0.18,
				TurbulenceFrequency:  18.0,
				TurbulenceAmplitude:  0.08,
				GutterProbability:    0.40, // Very unreliable
				GutterDepth:          0.60,
				GutterDuration:       0.30,
				TempVariationRange:   0.20,
				TempWarmBias:         0.0,
			},
			"emergency_light": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 1.5,
				OscillationAmplitude: 0.40, // Pulsing
				TurbulenceFrequency:  10.0,
				TurbulenceAmplitude:  0.10,
				GutterProbability:    0.30,
				GutterDepth:          0.50,
				GutterDuration:       0.25,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
			"candle": {
				SwayFrequency:        0.6,
				SwayAmplitude:        0.30,
				OscillationFrequency: 4.5,
				OscillationAmplitude: 0.20,
				TurbulenceFrequency:  22.0,
				TurbulenceAmplitude:  0.08,
				GutterProbability:    0.45, // Very unstable
				GutterDepth:          0.65,
				GutterDuration:       0.25,
				TempVariationRange:   0.25,
				TempWarmBias:         0.05,
			},
			"broken_lamp": {
				SwayFrequency:        0.2,
				SwayAmplitude:        0.15,
				OscillationFrequency: 8.0, // Rapid electrical flicker
				OscillationAmplitude: 0.35,
				TurbulenceFrequency:  30.0,
				TurbulenceAmplitude:  0.12,
				GutterProbability:    0.50,
				GutterDepth:          0.70,
				GutterDuration:       0.35,
				TempVariationRange:   0.15,
				TempWarmBias:         -0.1,
			},
		}
	case "cyberpunk":
		s.presets = map[string]FlickerParams{
			"neon_pink": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 0.0,
				OscillationAmplitude: 0.0,
				TurbulenceFrequency:  0.0,
				TurbulenceAmplitude:  0.0,
				GutterProbability:    0.05, // Occasional flicker
				GutterDepth:          0.3,
				GutterDuration:       0.1,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
			"neon_cyan": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 0.0,
				OscillationAmplitude: 0.0,
				TurbulenceFrequency:  0.0,
				TurbulenceAmplitude:  0.0,
				GutterProbability:    0.05,
				GutterDepth:          0.3,
				GutterDuration:       0.1,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
			"hologram": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 0.2,
				OscillationAmplitude: 0.08,
				TurbulenceFrequency:  5.0,
				TurbulenceAmplitude:  0.03,
				GutterProbability:    0.10,
				GutterDepth:          0.4,
				GutterDuration:       0.15,
				TempVariationRange:   0.10,
				TempWarmBias:         -0.1,
			},
			"streetlight": {
				SwayFrequency:        0.0,
				SwayAmplitude:        0.0,
				OscillationFrequency: 0.0,
				OscillationAmplitude: 0.0,
				TurbulenceFrequency:  0.0,
				TurbulenceAmplitude:  0.0,
				GutterProbability:    0.02,
				GutterDepth:          0.2,
				GutterDuration:       0.08,
				TempVariationRange:   0.0,
				TempWarmBias:         0.0,
			},
		}
	case "postapoc":
		s.presets = map[string]FlickerParams{
			"oil_lamp": {
				SwayFrequency:        0.6,
				SwayAmplitude:        0.22,
				OscillationFrequency: 4.0,
				OscillationAmplitude: 0.15,
				TurbulenceFrequency:  18.0,
				TurbulenceAmplitude:  0.06,
				GutterProbability:    0.30,
				GutterDepth:          0.50,
				GutterDuration:       0.20,
				TempVariationRange:   0.18,
				TempWarmBias:         0.15,
			},
			"fire_barrel": {
				SwayFrequency:        0.4,
				SwayAmplitude:        0.15,
				OscillationFrequency: 2.8,
				OscillationAmplitude: 0.12,
				TurbulenceFrequency:  14.0,
				TurbulenceAmplitude:  0.05,
				GutterProbability:    0.20,
				GutterDepth:          0.40,
				GutterDuration:       0.18,
				TempVariationRange:   0.15,
				TempWarmBias:         0.20,
			},
			"generator_light": {
				SwayFrequency:        0.1,
				SwayAmplitude:        0.08,
				OscillationFrequency: 6.0,
				OscillationAmplitude: 0.25,
				TurbulenceFrequency:  25.0,
				TurbulenceAmplitude:  0.10,
				GutterProbability:    0.35,
				GutterDepth:          0.55,
				GutterDuration:       0.25,
				TempVariationRange:   0.10,
				TempWarmBias:         -0.05,
			},
			"salvaged_lamp": {
				SwayFrequency:        0.25,
				SwayAmplitude:        0.12,
				OscillationFrequency: 5.5,
				OscillationAmplitude: 0.20,
				TurbulenceFrequency:  20.0,
				TurbulenceAmplitude:  0.08,
				GutterProbability:    0.40,
				GutterDepth:          0.60,
				GutterDuration:       0.28,
				TempVariationRange:   0.12,
				TempWarmBias:         0.0,
			},
		}
	default:
		s.presets = map[string]FlickerParams{
			"generic": {
				SwayFrequency:        0.4,
				SwayAmplitude:        0.15,
				OscillationFrequency: 3.0,
				OscillationAmplitude: 0.10,
				TurbulenceFrequency:  15.0,
				TurbulenceAmplitude:  0.04,
				GutterProbability:    0.15,
				GutterDepth:          0.35,
				GutterDuration:       0.15,
				TempVariationRange:   0.10,
				TempWarmBias:         0.0,
			},
		}
	}
}

// GetFlickerParams returns flicker parameters for a light type, initialized with seed.
func (s *System) GetFlickerParams(lightType string, seed int64, baseR, baseG, baseB float64) FlickerParams {
	preset, found := s.presets[lightType]
	if !found {
		preset = s.presets["generic"]
		if preset.SwayFrequency == 0 && preset.OscillationFrequency == 0 {
			// No generic preset, create default
			preset = FlickerParams{
				SwayFrequency:        0.4,
				SwayAmplitude:        0.15,
				OscillationFrequency: 3.0,
				OscillationAmplitude: 0.10,
				TurbulenceFrequency:  15.0,
				TurbulenceAmplitude:  0.04,
				GutterProbability:    0.15,
				GutterDepth:          0.35,
				GutterDuration:       0.15,
				TempVariationRange:   0.10,
				TempWarmBias:         0.0,
			}
		}
	}

	// Initialize instance-specific fields
	params := preset
	params.LightType = lightType
	params.Seed = seed
	params.BaseR = baseR
	params.BaseG = baseG
	params.BaseB = baseB

	// Randomize phase shift for variety among multiple lights
	r := rng.NewRNG(uint64(seed))
	params.SwayPhaseShift = r.Float64() * 2.0 * math.Pi

	return params
}

// CalculateFlicker computes intensity and color at a given tick.
// Returns: intensity (0.0-1.0), r, g, b (0.0-1.0)
func (s *System) CalculateFlicker(params *FlickerParams, tick int, baseIntensity float64) (intensity, r, g, b float64) {
	// Convert tick to time (60 FPS assumed)
	t := float64(tick) / 60.0

	// Start with base intensity
	intensity = baseIntensity

	// Create deterministic RNG for this tick
	tickRNG := rng.NewRNG(uint64(params.Seed + int64(tick)))

	// 1. Low-frequency sway (main flame body movement)
	if params.SwayFrequency > 0 && params.SwayAmplitude > 0 {
		swayPhase := 2.0*math.Pi*params.SwayFrequency*t + params.SwayPhaseShift
		sway := math.Sin(swayPhase) * params.SwayAmplitude
		intensity += sway * baseIntensity
	}

	// 2. Mid-frequency oscillation (rapid brightness fluctuations)
	if params.OscillationFrequency > 0 && params.OscillationAmplitude > 0 {
		oscPhase := 2.0 * math.Pi * params.OscillationFrequency * t
		oscillation := math.Sin(oscPhase) * params.OscillationAmplitude
		// Add slight noise to oscillation
		noiseScale := tickRNG.Float64()*0.3 - 0.15
		oscillation *= (1.0 + noiseScale)
		intensity += oscillation * baseIntensity
	}

	// 3. High-frequency turbulence (visual shimmer)
	if params.TurbulenceFrequency > 0 && params.TurbulenceAmplitude > 0 {
		// Use simplex-like noise approximation via multiple sine waves
		turbPhase1 := 2.0 * math.Pi * params.TurbulenceFrequency * t
		turbPhase2 := 2.0 * math.Pi * params.TurbulenceFrequency * 1.3 * t
		turbPhase3 := 2.0 * math.Pi * params.TurbulenceFrequency * 0.7 * t
		turbulence := (math.Sin(turbPhase1) + 0.5*math.Sin(turbPhase2) + 0.3*math.Sin(turbPhase3)) / 1.8
		turbulence *= params.TurbulenceAmplitude
		intensity += turbulence * baseIntensity
	}

	// 4. Guttering events (sudden dips)
	intensity = s.applyGuttering(params, tick, intensity, tickRNG)

	// Clamp intensity to valid range
	intensity = clamp(intensity, 0.0, 1.2) // Allow slight over-bright for flare effect

	// 5. Calculate color temperature variation
	r, g, b = s.calculateColorTemperature(params, intensity, baseIntensity)

	return intensity, r, g, b
}

// applyGuttering handles sudden intensity drops (air current disruption).
func (s *System) applyGuttering(params *FlickerParams, tick int, intensity float64, tickRNG *rng.RNG) float64 {
	if params.GutterProbability <= 0 || params.GutterDepth <= 0 {
		return intensity
	}

	// Check if currently in a gutter event
	if params.InGutter {
		ticksElapsed := tick - params.GutterStartTick
		if ticksElapsed < params.GutterLength {
			// Calculate gutter shape (quick drop, gradual recovery)
			progress := float64(ticksElapsed) / float64(params.GutterLength)
			// Use smooth step for natural recovery
			gutterFactor := 1.0 - params.GutterDepth*smoothStep(1.0-progress)
			return intensity * gutterFactor
		}
		// Gutter ended
		params.InGutter = false
	}

	// Check if new gutter event should start
	// Probability per tick (convert per-second to per-tick at 60 FPS)
	perTickProb := params.GutterProbability / 60.0
	if tickRNG.Float64() < perTickProb {
		params.InGutter = true
		params.GutterStartTick = tick
		params.GutterLength = int(params.GutterDuration * 60.0) // Convert seconds to ticks

		// Apply immediate gutter drop
		return intensity * (1.0 - params.GutterDepth)
	}

	return intensity
}

// calculateColorTemperature adjusts RGB based on intensity variation.
// Hotter (higher intensity) shifts toward blue-white, cooler shifts toward red-orange.
func (s *System) calculateColorTemperature(params *FlickerParams, currentIntensity, baseIntensity float64) (r, g, b float64) {
	r = params.BaseR
	g = params.BaseG
	b = params.BaseB

	if params.TempVariationRange <= 0 {
		return r, g, b
	}

	// Calculate deviation from base intensity
	deviation := currentIntensity - baseIntensity // Positive = hotter, negative = cooler

	// Normalize deviation by amplitude range
	normalizedDev := clamp(deviation/(baseIntensity*0.5), -1.0, 1.0)

	// Apply warm bias
	normalizedDev += params.TempWarmBias

	// Temperature shift
	tempShift := normalizedDev * params.TempVariationRange

	if tempShift > 0 {
		// Hotter: shift toward white/blue (increase G and B relative to R)
		g += tempShift * 0.15
		b += tempShift * 0.20
	} else {
		// Cooler: shift toward red/orange (decrease G and B relative to R)
		g += tempShift * 0.20 // Subtracts because tempShift is negative
		b += tempShift * 0.30
	}

	// Clamp colors
	r = clamp(r, 0.0, 1.0)
	g = clamp(g, 0.0, 1.0)
	b = clamp(b, 0.0, 1.0)

	return r, g, b
}

// smoothStep provides smooth interpolation (Hermite curve).
func smoothStep(t float64) float64 {
	t = clamp(t, 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

// clamp restricts a value to [min, max].
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// SetGenre changes the active genre and reinitializes presets.
func (s *System) SetGenre(genre string) {
	s.genre = genre
	s.initializePresets()
}

// GetGenre returns the current genre.
func (s *System) GetGenre() string {
	return s.genre
}

// GetPresetNames returns all available light type names for the current genre.
func (s *System) GetPresetNames() []string {
	names := make([]string, 0, len(s.presets))
	for name := range s.presets {
		names = append(names, name)
	}
	return names
}

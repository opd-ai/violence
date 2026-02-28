package lighting

import (
	"math"

	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
)

// PointLight represents a light source at a specific location.
// It extends the base Light type with additional metadata and configuration.
type PointLight struct {
	Light               // Embedded base light
	LightType    string // Type identifier (torch, lamp, monitor, fire, etc.)
	IsFlickering bool   // Whether light intensity fluctuates
	FlickerSeed  int64  // Seed for deterministic flicker pattern
}

// LightPreset defines genre-specific lighting configurations.
type LightPreset struct {
	Name      string
	Radius    float64
	Intensity float64
	R, G, B   float64
	Flicker   bool
}

// GetGenrePresets returns all light source definitions for a genre.
func GetGenrePresets(genreID string) []LightPreset {
	switch genreID {
	case genre.Fantasy:
		return []LightPreset{
			{"torch", 5.0, 0.8, 1.0, 0.6, 0.2, true},
			{"brazier", 7.0, 1.0, 1.0, 0.5, 0.1, true},
			{"candle", 2.5, 0.4, 1.0, 0.8, 0.3, true},
			{"magic_crystal", 6.0, 0.9, 0.4, 0.6, 1.0, false},
		}
	case genre.SciFi:
		return []LightPreset{
			{"monitor", 3.0, 0.6, 0.4, 0.7, 1.0, false},
			{"ceiling_lamp", 8.0, 1.0, 0.9, 0.95, 1.0, false},
			{"alarm", 4.0, 0.8, 1.0, 0.2, 0.2, true},
			{"console", 3.5, 0.5, 0.3, 0.8, 0.6, false},
		}
	case genre.Horror:
		return []LightPreset{
			{"dim_bulb", 4.0, 0.4, 0.6, 0.6, 0.4, true},
			{"emergency_light", 5.0, 0.5, 0.8, 0.2, 0.2, true},
			{"candle", 2.0, 0.3, 1.0, 0.7, 0.3, true},
			{"broken_lamp", 3.0, 0.35, 0.5, 0.5, 0.3, true},
		}
	case genre.Cyberpunk:
		return []LightPreset{
			{"neon_pink", 6.0, 0.9, 1.0, 0.2, 0.8, false},
			{"neon_cyan", 6.0, 0.9, 0.2, 0.8, 1.0, false},
			{"hologram", 4.0, 0.7, 0.5, 0.7, 1.0, true},
			{"streetlight", 10.0, 0.8, 0.9, 0.9, 0.7, false},
		}
	case genre.PostApoc:
		return []LightPreset{
			{"oil_lamp", 4.0, 0.6, 1.0, 0.6, 0.2, true},
			{"fire_barrel", 6.0, 0.9, 1.0, 0.4, 0.1, true},
			{"generator_light", 5.0, 0.7, 0.8, 0.8, 0.6, true},
			{"salvaged_lamp", 3.5, 0.5, 0.7, 0.6, 0.4, true},
		}
	default:
		return []LightPreset{
			{"generic", 5.0, 0.7, 1.0, 1.0, 1.0, false},
		}
	}
}

// NewPointLight creates a point light from a preset.
func NewPointLight(x, y float64, preset LightPreset, seed int64) PointLight {
	return PointLight{
		Light: Light{
			X:         x,
			Y:         y,
			Radius:    preset.Radius,
			Intensity: preset.Intensity,
			R:         preset.R,
			G:         preset.G,
			B:         preset.B,
		},
		LightType:    preset.Name,
		IsFlickering: preset.Flicker,
		FlickerSeed:  seed,
	}
}

// GetPresetByName finds a light preset by name from a genre's presets.
func GetPresetByName(genreID, name string) (LightPreset, bool) {
	presets := GetGenrePresets(genreID)
	for _, p := range presets {
		if p.Name == name {
			return p, true
		}
	}
	return LightPreset{}, false
}

// UpdateFlicker modifies light intensity based on flicker pattern.
// tick is the current game tick for deterministic animation.
// Returns the modified intensity value.
func (pl *PointLight) UpdateFlicker(tick int) float64 {
	if !pl.IsFlickering {
		return pl.Intensity
	}

	// Create deterministic RNG from seed and tick
	r := rng.NewRNG(uint64(pl.FlickerSeed + int64(tick/10)))

	// Generate flicker variation: ±15% of base intensity
	variation := r.Float64()*0.3 - 0.15
	flickeredIntensity := pl.Intensity * (1.0 + variation)

	// Clamp to valid range
	if flickeredIntensity < 0.0 {
		flickeredIntensity = 0.0
	}
	if flickeredIntensity > 1.0 {
		flickeredIntensity = 1.0
	}

	return flickeredIntensity
}

// ApplyAttenuation calculates light contribution at a point.
// Uses quadratic falloff: intensity / (1 + distance²)
func (pl *PointLight) ApplyAttenuation(targetX, targetY float64) float64 {
	dx := targetX - pl.X
	dy := targetY - pl.Y
	distSq := dx*dx + dy*dy
	radiusSq := pl.Radius * pl.Radius

	// No contribution outside radius
	if distSq > radiusSq {
		return 0.0
	}

	// Quadratic attenuation
	return pl.Intensity / (1.0 + distSq)
}

// LinearAttenuation calculates light contribution with linear falloff.
// intensity * (1 - distance/radius)
func (pl *PointLight) LinearAttenuation(targetX, targetY float64) float64 {
	dx := targetX - pl.X
	dy := targetY - pl.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// No contribution outside radius
	if dist > pl.Radius {
		return 0.0
	}

	// Linear attenuation
	falloff := 1.0 - (dist / pl.Radius)
	return pl.Intensity * falloff
}

// SetPosition updates the light's location.
func (pl *PointLight) SetPosition(x, y float64) {
	pl.X = x
	pl.Y = y
}

// SetIntensity updates the base light intensity.
func (pl *PointLight) SetIntensity(intensity float64) {
	if intensity < 0.0 {
		pl.Intensity = 0.0
	} else if intensity > 1.0 {
		pl.Intensity = 1.0
	} else {
		pl.Intensity = intensity
	}
}

// SetColor updates the light color components.
func (pl *PointLight) SetColor(r, g, b float64) {
	pl.R = clampColor(r)
	pl.G = clampColor(g)
	pl.B = clampColor(b)
}

// clampColor restricts color component to [0.0, 1.0].
func clampColor(value float64) float64 {
	if value < 0.0 {
		return 0.0
	}
	if value > 1.0 {
		return 1.0
	}
	return value
}

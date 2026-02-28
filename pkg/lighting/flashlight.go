package lighting

import (
	"math"

	"github.com/opd-ai/violence/pkg/procgen/genre"
)

// ConeLight represents a directional cone-shaped light source.
// Used for flashlights, torches, headlamps, and other forward-facing illumination.
type ConeLight struct {
	X, Y           float64 // Position
	DirX, DirY     float64 // Direction vector (normalized)
	Radius         float64 // Maximum reach distance
	Angle          float64 // Cone half-angle in radians
	Intensity      float64 // Light intensity [0.0-1.0]
	R, G, B        float64 // Color components [0.0-1.0]
	IsActive       bool    // Whether light is currently on
	FlashlightType string  // Genre-specific variant name
}

// FlashlightPreset defines genre-specific flashlight configurations.
type FlashlightPreset struct {
	Name      string
	Radius    float64
	Angle     float64 // Half-angle in radians
	Intensity float64
	R, G, B   float64
}

// GetFlashlightPreset returns the genre-appropriate flashlight configuration.
func GetFlashlightPreset(genreID string) FlashlightPreset {
	switch genreID {
	case genre.Fantasy:
		// Torch: warm orange glow, wide cone, moderate reach
		return FlashlightPreset{
			Name:      "torch",
			Radius:    8.0,
			Angle:     math.Pi / 4, // 45 degrees
			Intensity: 0.9,
			R:         1.0,
			G:         0.6,
			B:         0.2,
		}
	case genre.SciFi:
		// Headlamp: bright white, narrow focused beam, long reach
		return FlashlightPreset{
			Name:      "headlamp",
			Radius:    12.0,
			Angle:     math.Pi / 6, // 30 degrees
			Intensity: 1.0,
			R:         0.9,
			G:         0.95,
			B:         1.0,
		}
	case genre.Horror:
		// Dim flashlight: weak yellowish, narrow beam, short reach
		return FlashlightPreset{
			Name:      "flashlight",
			Radius:    6.0,
			Angle:     math.Pi / 8, // 22.5 degrees
			Intensity: 0.6,
			R:         0.8,
			G:         0.7,
			B:         0.5,
		}
	case genre.Cyberpunk:
		// Glow-rod: cyan neon, medium cone, moderate reach
		return FlashlightPreset{
			Name:      "glow_rod",
			Radius:    9.0,
			Angle:     math.Pi / 5, // 36 degrees
			Intensity: 0.85,
			R:         0.3,
			G:         0.8,
			B:         1.0,
		}
	case genre.PostApoc:
		// Salvaged lamp: dim warm, wide cone, short reach
		return FlashlightPreset{
			Name:      "salvaged_lamp",
			Radius:    7.0,
			Angle:     math.Pi / 3.5, // ~51 degrees
			Intensity: 0.7,
			R:         0.9,
			G:         0.7,
			B:         0.4,
		}
	default:
		// Generic flashlight
		return FlashlightPreset{
			Name:      "flashlight",
			Radius:    10.0,
			Angle:     math.Pi / 6,
			Intensity: 0.8,
			R:         1.0,
			G:         1.0,
			B:         1.0,
		}
	}
}

// NewConeLight creates a cone light from a preset.
func NewConeLight(x, y, dirX, dirY float64, preset FlashlightPreset) ConeLight {
	// Normalize direction
	length := math.Sqrt(dirX*dirX + dirY*dirY)
	if length > 0.001 {
		dirX /= length
		dirY /= length
	} else {
		dirX = 1.0
		dirY = 0.0
	}

	return ConeLight{
		X:              x,
		Y:              y,
		DirX:           dirX,
		DirY:           dirY,
		Radius:         preset.Radius,
		Angle:          preset.Angle,
		Intensity:      preset.Intensity,
		R:              preset.R,
		G:              preset.G,
		B:              preset.B,
		IsActive:       true,
		FlashlightType: preset.Name,
	}
}

// SetDirection updates the cone's facing direction.
func (cl *ConeLight) SetDirection(dirX, dirY float64) {
	// Normalize direction vector
	length := math.Sqrt(dirX*dirX + dirY*dirY)
	if length > 0.001 {
		cl.DirX = dirX / length
		cl.DirY = dirY / length
	}
}

// SetPosition updates the cone light's location.
func (cl *ConeLight) SetPosition(x, y float64) {
	cl.X = x
	cl.Y = y
}

// Toggle switches the light on/off.
func (cl *ConeLight) Toggle() {
	cl.IsActive = !cl.IsActive
}

// SetActive explicitly sets the light state.
func (cl *ConeLight) SetActive(active bool) {
	cl.IsActive = active
}

// ApplyConeAttenuation calculates light contribution at a target point.
// Returns 0.0 if the point is outside the cone or beyond radius.
// Uses combined distance and angle attenuation.
func (cl *ConeLight) ApplyConeAttenuation(targetX, targetY float64) float64 {
	if !cl.IsActive {
		return 0.0
	}

	// Vector from light to target
	toTargetX := targetX - cl.X
	toTargetY := targetY - cl.Y
	dist := math.Sqrt(toTargetX*toTargetX + toTargetY*toTargetY)

	// Outside radius: no contribution
	if dist > cl.Radius || dist < 0.001 {
		return 0.0
	}

	// Normalize target vector
	toTargetX /= dist
	toTargetY /= dist

	// Calculate angle between cone direction and target
	// dot product = cos(angle)
	dotProduct := cl.DirX*toTargetX + cl.DirY*toTargetY
	angleToTarget := math.Acos(clamp(dotProduct, -1.0, 1.0))

	// Outside cone angle: no contribution
	if angleToTarget > cl.Angle {
		return 0.0
	}

	// Distance attenuation: quadratic falloff
	distAttenuation := 1.0 / (1.0 + dist*dist)

	// Angular attenuation: brighter in center, dimmer at edges
	// 1.0 at center, 0.0 at edge
	angleAttenuation := 1.0 - (angleToTarget / cl.Angle)

	// Combine attenuations
	totalAttenuation := distAttenuation * angleAttenuation
	return cl.Intensity * totalAttenuation
}

// GetContributionAsPointLight converts cone light to equivalent point light at position.
// Useful for adding flashlight to sector light map as a simplified point source.
func (cl *ConeLight) GetContributionAsPointLight() Light {
	return Light{
		X:         cl.X,
		Y:         cl.Y,
		Radius:    cl.Radius,
		Intensity: cl.Intensity * 0.7, // Reduce for point approximation
		R:         cl.R,
		G:         cl.G,
		B:         cl.B,
	}
}

// IsPointInCone checks if a point lies within the cone's illumination area.
func (cl *ConeLight) IsPointInCone(targetX, targetY float64) bool {
	if !cl.IsActive {
		return false
	}

	// Vector from light to target
	toTargetX := targetX - cl.X
	toTargetY := targetY - cl.Y
	distSq := toTargetX*toTargetX + toTargetY*toTargetY

	// Outside radius
	if distSq > cl.Radius*cl.Radius {
		return false
	}

	dist := math.Sqrt(distSq)
	if dist < 0.001 {
		return true // At light position
	}

	// Normalize target vector
	toTargetX /= dist
	toTargetY /= dist

	// Calculate angle
	dotProduct := cl.DirX*toTargetX + cl.DirY*toTargetY
	angleToTarget := math.Acos(clamp(dotProduct, -1.0, 1.0))

	return angleToTarget <= cl.Angle
}

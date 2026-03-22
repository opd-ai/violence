package volumetric

import "image/color"

// Component stores volumetric lighting data for a light source.
// This enables per-light volumetric effects with configurable parameters.
type Component struct {
	// Enabled controls whether this light produces volumetric rays
	Enabled bool

	// Intensity scales the volumetric effect strength [0.0-1.0]
	Intensity float64

	// Radius is the maximum distance light rays travel
	Radius float64

	// DustDensity controls how much light scatters in the air [0.0-1.0]
	DustDensity float64

	// ScatterColor tints the volumetric rays (dust/fog color)
	ScatterR, ScatterG, ScatterB float64

	// FalloffExponent controls how quickly rays fade with distance
	FalloffExponent float64

	// ConeAngle limits rays to a cone (0 = omnidirectional, >0 = spotlight)
	ConeAngle float64

	// ConeDirectionX, ConeDirectionY for directional lights
	ConeDirectionX, ConeDirectionY float64
}

// Type implements engine.Component.
func (c *Component) Type() string {
	return "Volumetric"
}

// NewComponent creates a volumetric component with default values.
func NewComponent() *Component {
	return &Component{
		Enabled:         true,
		Intensity:       0.5,
		Radius:          5.0,
		DustDensity:     0.3,
		ScatterR:        1.0,
		ScatterG:        0.95,
		ScatterB:        0.85,
		FalloffExponent: 2.0,
		ConeAngle:       0, // Omnidirectional by default
	}
}

// GetScatterColor returns the scatter tint as an RGBA color.
func (c *Component) GetScatterColor(alpha float64) color.RGBA {
	a := alpha
	if a > 1.0 {
		a = 1.0
	}
	if a < 0 {
		a = 0
	}
	return color.RGBA{
		R: uint8(c.ScatterR * 255),
		G: uint8(c.ScatterG * 255),
		B: uint8(c.ScatterB * 255),
		A: uint8(a * 255),
	}
}

// SetFromLightColor configures scatter color based on light source color.
// Applies a warm dust tint shift for realism.
func (c *Component) SetFromLightColor(r, g, b float64) {
	// Scatter color is slightly warmer than source (dust absorption)
	c.ScatterR = clampF64(r*1.05, 0, 1)
	c.ScatterG = clampF64(g*0.98, 0, 1)
	c.ScatterB = clampF64(b*0.90, 0, 1)
}

// IsDirectional returns true if this is a spotlight/cone light.
func (c *Component) IsDirectional() bool {
	return c.ConeAngle > 0
}

// clampF64 clamps a float64 to [min, max].
func clampF64(v, minV, maxV float64) float64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

// LightShaft represents a single volumetric light source for rendering.
type LightShaft struct {
	// World position
	X, Y float64

	// Light properties
	Intensity float64 // Overall brightness [0.0-1.0]
	Radius    float64 // Maximum ray distance

	// Color
	R, G, B float64

	// Volumetric parameters
	DustDensity     float64
	FalloffExponent float64

	// Directional properties (0 cone = omnidirectional)
	ConeAngle float64
	DirX, DirY float64
}

// NewLightShaft creates a light shaft from position and color.
func NewLightShaft(x, y, radius, intensity, r, g, b float64) LightShaft {
	return LightShaft{
		X:               x,
		Y:               y,
		Intensity:       intensity,
		Radius:          radius,
		R:               r,
		G:               g,
		B:               b,
		DustDensity:     0.3,
		FalloffExponent: 2.0,
		ConeAngle:       0,
	}
}

// WithCone makes this a directional spotlight.
func (s LightShaft) WithCone(angle, dirX, dirY float64) LightShaft {
	s.ConeAngle = angle
	s.DirX = dirX
	s.DirY = dirY
	return s
}

// WithDust sets custom dust density.
func (s LightShaft) WithDust(density float64) LightShaft {
	s.DustDensity = density
	return s
}

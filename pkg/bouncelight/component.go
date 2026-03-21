package bouncelight

import "image/color"

// Component stores bounce lighting data for a tile or surface.
// This is transient data recalculated per-frame or per-room-load.
type Component struct {
	// TintR, TintG, TintB are the accumulated bounce color [0.0-1.0]
	TintR, TintG, TintB float64

	// Intensity is the total bounce light intensity [0.0-1.0]
	Intensity float64

	// ContributorCount tracks how many surfaces contributed
	ContributorCount int

	// Enabled allows selective disable per-tile
	Enabled bool
}

// Type implements engine.Component.
func (c *Component) Type() string {
	return "BounceLight"
}

// NewComponent creates a bounce light component with default values.
func NewComponent() *Component {
	return &Component{
		TintR:     0,
		TintG:     0,
		TintB:     0,
		Intensity: 0,
		Enabled:   true,
	}
}

// GetTint returns the bounce tint as an RGBA color with given strength.
// The alpha channel encodes the intensity for blending.
func (c *Component) GetTint(strength float64) color.RGBA {
	s := strength * c.Intensity
	if s > 1.0 {
		s = 1.0
	}
	return color.RGBA{
		R: uint8(c.TintR * 255 * s),
		G: uint8(c.TintG * 255 * s),
		B: uint8(c.TintB * 255 * s),
		A: uint8(s * 255),
	}
}

// AddContribution adds a bounce contribution from a nearby surface.
// color is the surface color, intensity is the contribution strength.
func (c *Component) AddContribution(r, g, b, intensity float64) {
	if intensity <= 0 || !c.Enabled {
		return
	}

	// Weighted additive blending
	c.TintR += r * intensity
	c.TintG += g * intensity
	c.TintB += b * intensity
	c.Intensity += intensity
	c.ContributorCount++
}

// Normalize finalizes the bounce values after all contributions.
// Should be called once all nearby surfaces have contributed.
func (c *Component) Normalize() {
	if c.ContributorCount == 0 || c.Intensity == 0 {
		return
	}

	// Average the color contributions
	c.TintR /= c.Intensity
	c.TintG /= c.Intensity
	c.TintB /= c.Intensity

	// Clamp final intensity
	if c.Intensity > 1.0 {
		c.Intensity = 1.0
	}

	// Clamp colors
	if c.TintR > 1.0 {
		c.TintR = 1.0
	}
	if c.TintG > 1.0 {
		c.TintG = 1.0
	}
	if c.TintB > 1.0 {
		c.TintB = 1.0
	}
}

// Reset clears bounce data for recalculation.
func (c *Component) Reset() {
	c.TintR = 0
	c.TintG = 0
	c.TintB = 0
	c.Intensity = 0
	c.ContributorCount = 0
}

// BounceSurface represents a surface that can contribute bounce light.
type BounceSurface struct {
	// World position
	X, Y float64

	// Surface color (normalized 0-1)
	R, G, B float64

	// How much light this surface reflects (albedo)
	Reflectivity float64

	// Whether this is a wall (vertical) or floor (horizontal)
	IsWall bool

	// Illumination level from direct lighting
	DirectLight float64
}

// NewBounceSurface creates a surface with default reflectivity.
func NewBounceSurface(x, y, r, g, b float64, isWall bool) BounceSurface {
	return BounceSurface{
		X:            x,
		Y:            y,
		R:            r,
		G:            g,
		B:            b,
		Reflectivity: 0.5, // Default 50% reflection
		IsWall:       isWall,
		DirectLight:  1.0,
	}
}

// BounceContribution calculates this surface's contribution to a target point.
// Returns r, g, b, intensity values.
func (s *BounceSurface) BounceContribution(targetX, targetY, maxDist float64) (r, g, b, intensity float64) {
	// Calculate distance
	dx := targetX - s.X
	dy := targetY - s.Y
	distSq := dx*dx + dy*dy

	if distSq > maxDist*maxDist {
		return 0, 0, 0, 0
	}

	// Inverse-square falloff with minimum distance to avoid division by zero
	minDistSq := 1.0
	if distSq < minDistSq {
		distSq = minDistSq
	}

	// Calculate intensity with inverse-square falloff
	normalizedDist := distSq / (maxDist * maxDist)
	falloff := 1.0 - normalizedDist
	if falloff < 0 {
		falloff = 0
	}

	// Apply reflectivity and direct lighting
	intensity = falloff * s.Reflectivity * s.DirectLight

	// Return the surface color weighted by intensity
	return s.R, s.G, s.B, intensity
}

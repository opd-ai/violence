package damagedir

import "image/color"

// Component stores directional damage indicator state.
// Each active damage indicator represents one hit from a specific direction.
type Component struct {
	// Direction in radians relative to screen center (0 = right, PI/2 = top)
	Direction float64

	// Intensity controls the visibility (0.0-1.0), scaled by damage amount
	Intensity float64

	// Lifetime remaining in seconds before this indicator is removed
	Lifetime float64

	// MaxLifetime is the total duration for fade calculation
	MaxLifetime float64

	// Color is the indicator tint, set by genre
	Color color.RGBA

	// ArcWidth in radians (how wide the damage arc spans)
	ArcWidth float64

	// EdgeDepth is how far the vignette extends from the screen edge (pixels)
	EdgeDepth float64
}

// Type returns the component type identifier for ECS registration.
func (c *Component) Type() string {
	return "damagedir.Component"
}

// GetAlpha returns the current alpha based on lifetime fade.
func (c *Component) GetAlpha() float64 {
	if c.MaxLifetime <= 0 {
		return 0
	}
	// Fade out over lifetime
	progress := 1.0 - (c.Lifetime / c.MaxLifetime)
	// Use ease-out curve for smoother fade
	fade := 1.0 - (progress * progress)
	return c.Intensity * fade
}

// IsExpired returns true if this indicator should be removed.
func (c *Component) IsExpired() bool {
	return c.Lifetime <= 0
}

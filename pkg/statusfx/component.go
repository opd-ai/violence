// Package statusfx provides visual effects for status conditions on entities.
package statusfx

// VisualComponent stores visual effect state for an entity with status effects.
type VisualComponent struct {
	Effects []EffectVisual
}

// EffectVisual represents the visual parameters for a single active status effect.
type EffectVisual struct {
	Name        string
	Color       uint32  // RGBA color from status effect
	Intensity   float64 // 0.0 to 1.0, pulsates over time
	ParticleAge float64 // Time accumulator for particle emission
}

// Type returns the component type name.
func (c *VisualComponent) Type() string {
	return "StatusFXVisual"
}

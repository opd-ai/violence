// Package telegraph provides visual attack telegraphing for enemies.
package telegraph

import (
	"image/color"
)

// Component holds attack telegraph state for an entity.
// Pure data - all logic in system.
type Component struct {
	// Telegraph state
	Active         bool
	ChargeProgress float64 // 0.0 to 1.0 - how far through the telegraph
	TelegraphTime  float64 // Total time to telegraph in seconds
	AttackType     string  // "melee", "ranged", "aoe", "charge"

	// Visual settings
	PrimaryColor    color.RGBA
	SecondaryColor  color.RGBA
	IndicatorRadius float64 // Radius of telegraph indicator
	IndicatorAlpha  float64 // Current alpha for pulsing effect

	// Position (updated from entity position)
	X, Y float64

	// Particle emission
	EmitParticles  bool
	ParticleCount  int
	ParticleSpread float64

	// Screen effects
	ScreenShake    float64 // Intensity of screen shake on attack execute
	FlashIntensity float64 // Flash intensity on attack execute
}

// Type implements Component interface.
func (c *Component) Type() string {
	return "telegraph"
}

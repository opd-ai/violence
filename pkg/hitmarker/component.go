// Package hitmarker provides visual hit confirmation feedback at screen center.
// When the player damages an enemy, a brief marker appears at the crosshair
// to confirm the hit registered, essential for responsive combat feel.
package hitmarker

import (
	"image/color"
)

// HitType represents different categories of hits for visual differentiation.
type HitType int

const (
	// HitNormal is a standard damage hit.
	HitNormal HitType = iota
	// HitCritical is a critical/high-damage hit.
	HitCritical
	// HitKill is the killing blow on an enemy.
	HitKill
	// HitHeadshot is a precision headshot (ranged).
	HitHeadshot
	// HitWeakpoint is damage to a weak point.
	HitWeakpoint
)

// Component stores hit marker state for rendering.
type Component struct {
	// Active indicates whether a hit marker is currently showing.
	Active bool

	// HitType determines the visual style of the marker.
	HitType HitType

	// Age is time since the hit marker was triggered (seconds).
	Age float64

	// Duration is total display time before fade (seconds).
	Duration float64

	// Scale is the current size multiplier (for pop animation).
	Scale float64

	// Alpha is the current opacity (0-1).
	Alpha float64

	// Rotation is the current rotation angle (radians).
	Rotation float64

	// Color is the marker color (varies by hit type and genre).
	Color color.RGBA

	// SecondaryColor is for gradient/accent effects.
	SecondaryColor color.RGBA

	// Intensity scales the visual effect (higher for bigger hits).
	Intensity float64

	// DamageValue is the damage dealt (for intensity scaling).
	DamageValue int

	// ScreenX is the X position where marker appears (usually screen center).
	ScreenX float64

	// ScreenY is the Y position where marker appears (usually screen center).
	ScreenY float64
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "hitmarker"
}

// NewComponent creates a default hit marker component.
func NewComponent() *Component {
	return &Component{
		Active:         false,
		HitType:        HitNormal,
		Age:            0,
		Duration:       0.25, // Quick flash
		Scale:          1.0,
		Alpha:          1.0,
		Rotation:       0,
		Color:          color.RGBA{255, 255, 255, 255},
		SecondaryColor: color.RGBA{200, 200, 200, 255},
		Intensity:      1.0,
		DamageValue:    0,
		ScreenX:        0,
		ScreenY:        0,
	}
}

// Trigger activates the hit marker with given parameters.
func (c *Component) Trigger(hitType HitType, damageValue int, screenX, screenY float64) {
	c.Active = true
	c.HitType = hitType
	c.Age = 0
	c.DamageValue = damageValue
	c.ScreenX = screenX
	c.ScreenY = screenY

	// Scale intensity based on damage and hit type
	c.Intensity = 1.0
	if damageValue > 50 {
		c.Intensity = 1.5
	}
	if damageValue > 100 {
		c.Intensity = 2.0
	}

	// Hit type modifiers
	switch hitType {
	case HitCritical:
		c.Duration = 0.35
		c.Intensity *= 1.5
	case HitKill:
		c.Duration = 0.5
		c.Intensity *= 2.0
	case HitHeadshot:
		c.Duration = 0.4
		c.Intensity *= 1.75
	case HitWeakpoint:
		c.Duration = 0.3
		c.Intensity *= 1.25
	default:
		c.Duration = 0.25
	}

	// Reset animation state
	c.Scale = 0.5 // Start small for pop-in
	c.Alpha = 1.0
	c.Rotation = 0
}

// Reset clears the hit marker state.
func (c *Component) Reset() {
	c.Active = false
	c.Age = 0
	c.Scale = 1.0
	c.Alpha = 1.0
	c.Rotation = 0
	c.Intensity = 1.0
}

// Package damagenumber provides floating combat text for damage feedback.
package damagenumber

import (
	"image/color"
)

// Component marks an entity as having floating damage text.
type Component struct {
	Value      int
	DamageType string
	IsCritical bool
	IsHeal     bool
	X          float64
	Y          float64
	VelocityY  float64
	Lifetime   float64
	Age        float64
	Scale      float64
	Alpha      float64
	Color      color.RGBA
}

// Type returns the component type identifier for ECS registration.
func (c *Component) Type() string { return "DamageNumber" }

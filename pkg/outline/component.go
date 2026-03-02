// Package outline provides sprite silhouette rendering for visual clarity.
package outline

import "image/color"

// Component stores outline rendering configuration for an entity.
type Component struct {
	Enabled   bool
	Color     color.RGBA
	Thickness int
	Glow      bool
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "OutlineComponent"
}

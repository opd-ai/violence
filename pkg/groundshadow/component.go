// Package groundshadow provides entity ground-contact shadow rendering.
// Ground shadows are soft elliptical shadows beneath entities that visually
// anchor them to the floor plane, creating depth and realism.
package groundshadow

// Component holds ground shadow rendering parameters for an entity.
// This component should be attached to any entity that needs a grounding
// shadow beneath it (players, enemies, NPCs, dropped items).
type Component struct {
	// CastShadow controls whether this entity renders a ground shadow.
	CastShadow bool

	// Radius is the base shadow radius in world units. Larger entities
	// get larger shadows. Default is 0.5 for human-sized entities.
	Radius float64

	// Height is the entity's vertical offset from the ground in world units.
	// Higher entities cast larger, more diffuse shadows.
	Height float64

	// Opacity is the base shadow darkness [0.0-1.0]. Genre presets modify this.
	Opacity float64

	// Elongation controls shadow ellipse aspect ratio [0.0-1.0].
	// 0.0 = perfect circle, 1.0 = highly elongated toward light.
	Elongation float64

	// OffsetX and OffsetY shift the shadow center relative to entity feet.
	// Used for light-direction-based shadow offset.
	OffsetX float64
	OffsetY float64
}

// Type returns the component type identifier for ECS registration.
func (c *Component) Type() string {
	return "groundshadow"
}

// NewComponent creates a ground shadow component with default values.
func NewComponent() *Component {
	return &Component{
		CastShadow: true,
		Radius:     0.5,
		Height:     1.0,
		Opacity:    0.6,
		Elongation: 0.0,
		OffsetX:    0.0,
		OffsetY:    0.0,
	}
}

// NewComponentWithSize creates a ground shadow component for a specific entity size.
func NewComponentWithSize(radius, height float64) *Component {
	return &Component{
		CastShadow: true,
		Radius:     radius,
		Height:     height,
		Opacity:    0.6,
		Elongation: 0.0,
		OffsetX:    0.0,
		OffsetY:    0.0,
	}
}

package crosshair

// Component stores crosshair state for an entity (typically the player).
type Component struct {
	// AimX and AimY represent the aim direction vector (normalized)
	AimX, AimY float64

	// WeaponType determines crosshair style ("melee", "ranged", "magic")
	WeaponType string

	// Visible controls whether the crosshair is rendered
	Visible bool

	// Range is the distance at which the crosshair is drawn from the entity
	Range float64

	// Scale multiplier for crosshair size
	Scale float64

	// Color tint for crosshair (R, G, B, A in 0-1 range)
	ColorR, ColorG, ColorB, ColorA float64

	// SwayOffsetX and SwayOffsetY apply weapon sway offset to crosshair position (pixels)
	SwayOffsetX, SwayOffsetY float64
}

// Type implements the engine.Component interface.
func (c *Component) Type() string {
	return "crosshair.Component"
}

// NewComponent creates a crosshair component with default values.
func NewComponent() *Component {
	return &Component{
		AimX:        1.0,
		AimY:        0.0,
		WeaponType:  "melee",
		Visible:     true,
		Range:       3.0,
		Scale:       1.0,
		ColorR:      1.0,
		ColorG:      1.0,
		ColorB:      1.0,
		ColorA:      0.8,
		SwayOffsetX: 0.0,
		SwayOffsetY: 0.0,
	}
}

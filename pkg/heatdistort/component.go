package heatdistort

// HeatSourceType identifies the type of heat-emitting source.
type HeatSourceType int

const (
	// HeatTorch represents a standard torch or wall-mounted flame.
	HeatTorch HeatSourceType = iota
	// HeatBrazier represents a larger standing fire source.
	HeatBrazier
	// HeatLava represents molten lava or magma surfaces.
	HeatLava
	// HeatExplosion represents temporary explosive heat.
	HeatExplosion
	// HeatPlasma represents sci-fi plasma vents or energy sources.
	HeatPlasma
	// HeatRadiation represents radioactive heat sources.
	HeatRadiation
	// HeatMagic represents magical fire or energy effects.
	HeatMagic
)

// Component stores heat distortion parameters for an entity.
type Component struct {
	// SourceType identifies what kind of heat source this is.
	SourceType HeatSourceType

	// Intensity controls the strength of the distortion effect (0.0-1.0).
	Intensity float64

	// Radius is the world-space radius of the heat influence zone.
	Radius float64

	// ScreenX and ScreenY are the screen-space coordinates of the heat source.
	// These are updated each frame by the system based on world position.
	ScreenX float64
	ScreenY float64

	// Visible indicates whether this heat source is currently on-screen.
	Visible bool

	// WavePhase tracks the animation phase for this specific source.
	WavePhase float64

	// Lifetime remaining for temporary sources (like explosions).
	// Permanent sources use -1.
	Lifetime float64

	// TintR, TintG, TintB provide optional color tinting of distorted areas.
	TintR, TintG, TintB float64
}

// Type returns the component type identifier for ECS.
func (c *Component) Type() string {
	return "heatdistort.Component"
}

// NewComponent creates a heat distortion component with default settings.
func NewComponent(sourceType HeatSourceType) *Component {
	c := &Component{
		SourceType: sourceType,
		Intensity:  0.5,
		Radius:     2.0,
		Visible:    true,
		WavePhase:  0,
		Lifetime:   -1, // Permanent by default
		TintR:      1.0,
		TintG:      1.0,
		TintB:      1.0,
	}

	// Apply source-type-specific defaults
	switch sourceType {
	case HeatTorch:
		c.Intensity = 0.4
		c.Radius = 1.5
		c.TintR, c.TintG, c.TintB = 1.1, 0.95, 0.85
	case HeatBrazier:
		c.Intensity = 0.6
		c.Radius = 2.5
		c.TintR, c.TintG, c.TintB = 1.15, 0.9, 0.8
	case HeatLava:
		c.Intensity = 0.8
		c.Radius = 4.0
		c.TintR, c.TintG, c.TintB = 1.2, 0.85, 0.7
	case HeatExplosion:
		c.Intensity = 1.0
		c.Radius = 5.0
		c.Lifetime = 0.5 // Half-second burst
		c.TintR, c.TintG, c.TintB = 1.3, 0.8, 0.6
	case HeatPlasma:
		c.Intensity = 0.5
		c.Radius = 2.0
		c.TintR, c.TintG, c.TintB = 0.9, 0.95, 1.2
	case HeatRadiation:
		c.Intensity = 0.6
		c.Radius = 3.5
		c.TintR, c.TintG, c.TintB = 0.95, 1.1, 0.9
	case HeatMagic:
		c.Intensity = 0.55
		c.Radius = 2.0
		c.TintR, c.TintG, c.TintB = 1.0, 0.9, 1.15
	}

	return c
}

// SetScreenPosition updates the screen-space position of the heat source.
func (c *Component) SetScreenPosition(x, y float64) {
	c.ScreenX = x
	c.ScreenY = y
}

// SetVisible marks the heat source as visible or hidden.
func (c *Component) SetVisible(visible bool) {
	c.Visible = visible
}

// UpdatePhase advances the wave animation phase.
func (c *Component) UpdatePhase(deltaTime, frequency float64) {
	c.WavePhase += deltaTime * frequency
}

// UpdateLifetime decrements lifetime for temporary sources.
// Returns true if the source has expired.
func (c *Component) UpdateLifetime(deltaTime float64) bool {
	if c.Lifetime < 0 {
		return false // Permanent source
	}
	c.Lifetime -= deltaTime
	return c.Lifetime <= 0
}

// IsActive returns whether the heat source should produce distortion.
func (c *Component) IsActive() bool {
	return c.Visible && c.Intensity > 0 && (c.Lifetime < 0 || c.Lifetime > 0)
}

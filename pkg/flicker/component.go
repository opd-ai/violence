package flicker

// Component stores flicker state for an entity's light source.
// This is attached to entities that have flickering light sources.
type Component struct {
	// Light source identification
	LightType string // Type of light (e.g., "torch", "candle")
	Seed      int64  // Deterministic seed for this instance

	// Cached flicker parameters (initialized from System)
	Params FlickerParams

	// Current computed values (updated each frame)
	CurrentIntensity float64 // Current intensity multiplier
	CurrentR         float64 // Current red component
	CurrentG         float64 // Current green component
	CurrentB         float64 // Current blue component

	// Whether flicker is enabled
	Enabled bool
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "flicker"
}

// NewComponent creates a flicker component with default values.
func NewComponent(lightType string, seed int64) *Component {
	return &Component{
		LightType:        lightType,
		Seed:             seed,
		CurrentIntensity: 1.0,
		CurrentR:         1.0,
		CurrentG:         1.0,
		CurrentB:         1.0,
		Enabled:          true,
	}
}

// Initialize sets up the component with parameters from the system.
func (c *Component) Initialize(sys *System, baseR, baseG, baseB float64) {
	c.Params = sys.GetFlickerParams(c.LightType, c.Seed, baseR, baseG, baseB)
}

// SetEnabled turns flicker on or off.
func (c *Component) SetEnabled(enabled bool) {
	c.Enabled = enabled
}

// GetIntensityMultiplier returns the current intensity factor for the light.
func (c *Component) GetIntensityMultiplier() float64 {
	if !c.Enabled {
		return 1.0
	}
	return c.CurrentIntensity
}

// GetColorModulation returns the current RGB modulation factors.
func (c *Component) GetColorModulation() (r, g, b float64) {
	if !c.Enabled {
		return 1.0, 1.0, 1.0
	}
	return c.CurrentR, c.CurrentG, c.CurrentB
}

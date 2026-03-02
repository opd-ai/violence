// Package fog provides atmospheric depth rendering with fog-of-war and distance cueing.
package fog

// Component marks an entity as affected by atmospheric fog rendering.
// Transient data - not serialized, rebuilt each frame from position.
type Component struct {
	DistanceFromCamera float64    // Updated each frame by system
	FogDensity         float64    // 0.0 = no fog, 1.0 = fully obscured
	Tint               [3]float64 // RGB color tint from fog [0.0-1.0]
	Visible            bool       // False if completely obscured by fog
}

// Type implements engine.Component.
func (c *Component) Type() string {
	return "fog"
}

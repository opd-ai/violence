package dmgfx

// DamageVisualComponent stores active damage-type visual effects on an entity.
// Pure data component - no methods with logic.
type DamageVisualComponent struct {
	ActiveEffects []ActiveEffect
}

// Type returns the component type identifier.
func (d *DamageVisualComponent) Type() string {
	return "DamageVisualComponent"
}

// ActiveEffect represents a single visual effect from damage.
type ActiveEffect struct {
	DamageTypeName string  // Name of damage type (Fire, Ice, etc.)
	Intensity      float64 // Visual intensity (0.0-1.0)
	Duration       float64 // Remaining seconds
	MaxDuration    float64 // Original duration for fade calculation
}

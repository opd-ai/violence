// Package damagestate provides visual damage state rendering for entities.
package damagestate

// Component tracks visual damage state for an entity.
type Component struct {
	CurrentHP    float64
	MaxHP        float64
	DamageLevel  int     // 0=pristine, 1=scratched, 2=wounded, 3=critical
	WoundPattern int64   // Seed for wound placement consistency
	LastDamageX  float64 // Last damage direction X for wound placement
	LastDamageY  float64 // Last damage direction Y for wound placement
	DirtyCache   bool    // Flag to regenerate overlay
}

// Type implements Component interface.
func (c *Component) Type() string {
	return "damagestate"
}

// UpdateDamage recalculates damage level based on HP ratio.
func (c *Component) UpdateDamage() {
	ratio := c.CurrentHP / c.MaxHP
	oldLevel := c.DamageLevel

	switch {
	case ratio >= 0.75:
		c.DamageLevel = 0 // Pristine
	case ratio >= 0.5:
		c.DamageLevel = 1 // Light damage
	case ratio >= 0.25:
		c.DamageLevel = 2 // Moderate damage
	default:
		c.DamageLevel = 3 // Critical damage
	}

	if oldLevel != c.DamageLevel {
		c.DirtyCache = true
	}
}

package projectile

// DamageType represents different types of damage that can have resistances.
type DamageType int

const (
	DamagePhysical DamageType = iota
	DamageFire
	DamageIce
	DamageLightning
	DamagePoison
	DamageHoly
	DamageShadow
	DamageArcane
)

// DamageTypeNames provides string names for damage types.
var DamageTypeNames = map[DamageType]string{
	DamagePhysical:  "Physical",
	DamageFire:      "Fire",
	DamageIce:       "Ice",
	DamageLightning: "Lightning",
	DamagePoison:    "Poison",
	DamageHoly:      "Holy",
	DamageShadow:    "Shadow",
	DamageArcane:    "Arcane",
}

// ResistanceComponent stores entity resistances to different damage types.
// Pure data component - no methods with logic.
type ResistanceComponent struct {
	Resistances map[DamageType]float64 // 0.0 = no resistance, 0.5 = 50% reduction, 1.0 = immune, -0.5 = 50% weakness
}

// Type returns the component type identifier.
func (r *ResistanceComponent) Type() string {
	return "ResistanceComponent"
}

// NewResistanceComponent creates a new resistance component with default values.
func NewResistanceComponent() *ResistanceComponent {
	return &ResistanceComponent{
		Resistances: make(map[DamageType]float64),
	}
}

// CalculateDamage applies resistance to incoming damage.
func CalculateDamage(baseDamage float64, damageType DamageType, resistances map[DamageType]float64) float64 {
	if resistances == nil {
		return baseDamage
	}
	
	resistance, exists := resistances[damageType]
	if !exists {
		return baseDamage
	}
	
	// resistance of 0.5 = 50% damage reduction
	// resistance of -0.5 = 50% damage increase (weakness)
	return baseDamage * (1.0 - resistance)
}

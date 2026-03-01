package hazard

// HazardComponent marks an entity as an environmental hazard.
type HazardComponent struct {
	Type             Type
	State            State
	Timer            float64
	CycleDuration    float64
	ActiveDuration   float64
	ChargeDuration   float64
	CooldownDuration float64
	Damage           int
	StatusEffect     string
	Persistent       bool
	Triggered        bool
	Width            float64
	Height           float64
	Color            uint32
}

// PositionComponent stores entity world position.
type PositionComponent struct {
	X, Y float64
}

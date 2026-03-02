// Package corpse implements persistent corpse rendering and management.
package corpse

// CorpseComponent stores corpse visual state for dead entities.
type CorpseComponent struct {
	Corpses []Corpse
}

// Type implements engine.Component.
func (c *CorpseComponent) Type() string {
	return "CorpseComponent"
}

// Corpse represents a single dead entity's persistent visual remains.
type Corpse struct {
	X, Y       float64
	Seed       int64
	EntityType string
	Subtype    string
	Angle      float64
	Opacity    float64
	Age        float64
	MaxAge     float64
	GenreID    string
	Size       int
	HasLoot    bool
	DeathType  DeathType
	BloodPool  bool
	Frame      int
}

// DeathType categorizes how the entity died for visual variety.
type DeathType int

const (
	DeathNormal DeathType = iota
	DeathBurn
	DeathFreeze
	DeathElectric
	DeathAcid
	DeathExplosion
	DeathSlash
	DeathCrush
	DeathDisintegrate
)

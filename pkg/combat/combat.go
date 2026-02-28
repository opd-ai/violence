// Package combat handles damage calculation and combat events.
package combat

import (
	"math"
)

// DamageType represents different damage categories.
type DamageType string

const (
	DamagePhysical  DamageType = "physical"
	DamageFire      DamageType = "fire"
	DamagePlasma    DamageType = "plasma"
	DamageEnergy    DamageType = "energy"
	DamageExplosive DamageType = "explosive"
)

// DamageEvent represents a single damage instance.
type DamageEvent struct {
	Source  uint64
	Target  uint64
	Amount  float64
	DmgType DamageType
	PosX    float64
	PosY    float64
}

// DamageResult contains the outcome of a damage application.
type DamageResult struct {
	HealthDamage float64
	ArmorDamage  float64
	Killed       bool
	DirectionX   float64
	DirectionY   float64
}

// System manages combat state and difficulty scaling.
type System struct {
	armorAbsorption float64
	gibThreshold    float64
	difficulty      float64
	genreID         string
}

// NewSystem creates a combat system with default settings.
func NewSystem() *System {
	return &System{
		armorAbsorption: 0.5,
		gibThreshold:    -50.0,
		difficulty:      1.0,
		genreID:         "fantasy",
	}
}

// SetDifficulty adjusts combat scaling (0.5 = easy, 1.0 = normal, 1.5 = hard).
func (s *System) SetDifficulty(scale float64) {
	s.difficulty = scale
}

// SetGenre configures combat visuals/audio for a genre.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	switch genreID {
	case "fantasy":
		s.armorAbsorption = 0.5
	case "scifi":
		s.armorAbsorption = 0.6
	case "horror":
		s.armorAbsorption = 0.4
	case "cyberpunk":
		s.armorAbsorption = 0.55
	case "postapoc":
		s.armorAbsorption = 0.45
	default:
		s.armorAbsorption = 0.5
	}
}

// ApplyDamage calculates damage with armor absorption and returns result.
func (s *System) ApplyDamage(currentHealth, currentArmor, rawDamage, targetX, targetY, sourceX, sourceY float64) DamageResult {
	result := DamageResult{}

	damage := rawDamage * s.difficulty

	if currentArmor > 0 {
		absorbed := currentArmor * s.armorAbsorption
		if absorbed > damage {
			result.ArmorDamage = damage / s.armorAbsorption
			result.HealthDamage = 0
		} else {
			result.ArmorDamage = currentArmor
			result.HealthDamage = damage - absorbed
		}
	} else {
		result.HealthDamage = damage
	}

	newHealth := currentHealth - result.HealthDamage
	result.Killed = newHealth <= 0

	dx := targetX - sourceX
	dy := targetY - sourceY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > 0.001 {
		result.DirectionX = dx / dist
		result.DirectionY = dy / dist
	}

	return result
}

// ShouldGib returns true if death damage was severe enough for gibs.
func (s *System) ShouldGib(finalHealth float64) bool {
	return finalHealth < s.gibThreshold
}

// ScaleDamage applies difficulty modifier to raw damage.
func (s *System) ScaleDamage(baseDamage float64) float64 {
	return baseDamage * s.difficulty
}

// global system instance (package-level for convenience)
var globalSystem = NewSystem()

// Apply processes a damage event using the global system.
func Apply(e DamageEvent) DamageResult {
	return globalSystem.ApplyDamage(0, 0, e.Amount, 0, 0, 0, 0)
}

// SetGenre configures combat parameters for a genre.
func SetGenre(genreID string) {
	globalSystem.SetGenre(genreID)
}

// SetDifficulty adjusts combat difficulty scaling.
func SetDifficulty(scale float64) {
	globalSystem.SetDifficulty(scale)
}

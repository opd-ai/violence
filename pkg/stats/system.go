package stats

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages stat-based calculations and applies stat bonuses to entities.
type System struct {
	logger *logrus.Entry
}

// NewSystem creates a new stat allocation system.
func NewSystem() *System {
	return &System{
		logger: logrus.WithFields(logrus.Fields{
			"system": "stats",
		}),
	}
}

// Update applies stat bonuses to entity attributes each frame.
// This system modifies health, damage, and other derived stats based on allocated attributes.
func (s *System) Update(w *engine.World) {
	// Query all entities with StatAllocationComponent
	statType := reflect.TypeOf((*StatAllocationComponent)(nil))
	entities := w.Query(statType)

	for _, entity := range entities {
		statComp, ok := w.GetComponent(entity, statType)
		if !ok {
			continue
		}

		stats, ok := statComp.(*StatAllocationComponent)
		if !ok || stats.Attributes == nil {
			continue
		}

		// Apply vitality bonus to max health (if health component exists)
		healthType := reflect.TypeOf((*HealthComponent)(nil))
		if healthComp, ok := w.GetComponent(entity, healthType); ok {
			if health, ok := healthComp.(*HealthComponent); ok {
				baseHealth := health.BaseMaxHealth
				if baseHealth == 0 {
					baseHealth = 100 // Default base health
					health.BaseMaxHealth = baseHealth
				}
				health.MaxHealth = int(float64(baseHealth) * stats.Attributes.GetMaxHealthBonus())

				// Cap current health if it exceeds new max
				if health.Current > health.MaxHealth {
					health.Current = health.MaxHealth
				}
			}
		}
	}
}

// HealthComponent is a minimal health component for stat integration.
type HealthComponent struct {
	Current       int
	MaxHealth     int
	BaseMaxHealth int // Base health before stat bonuses
}

// ApplyDamageBonus applies strength or intelligence bonuses to damage calculations.
func (s *System) ApplyDamageBonus(w *engine.World, entity engine.Entity, baseDamage float64, damageType string) float64 {
	statType := reflect.TypeOf((*StatAllocationComponent)(nil))
	statComp, ok := w.GetComponent(entity, statType)
	if !ok {
		return baseDamage
	}

	stats, ok := statComp.(*StatAllocationComponent)
	if !ok || stats.Attributes == nil {
		return baseDamage
	}

	switch damageType {
	case "melee", "physical":
		return baseDamage * stats.Attributes.GetMeleeDamageBonus()
	case "magic", "tech", "skill":
		return baseDamage * stats.Attributes.GetSkillPowerBonus()
	default:
		return baseDamage
	}
}

// ApplyAccuracyBonus applies dexterity bonuses to accuracy checks.
func (s *System) ApplyAccuracyBonus(w *engine.World, entity engine.Entity, baseAccuracy float64) float64 {
	statType := reflect.TypeOf((*StatAllocationComponent)(nil))
	statComp, ok := w.GetComponent(entity, statType)
	if !ok {
		return baseAccuracy
	}

	stats, ok := statComp.(*StatAllocationComponent)
	if !ok || stats.Attributes == nil {
		return baseAccuracy
	}

	return baseAccuracy * stats.Attributes.GetAccuracyBonus()
}

// RollCritical determines if an attack is a critical hit based on luck.
func (s *System) RollCritical(w *engine.World, entity engine.Entity, randomValue float64) bool {
	statType := reflect.TypeOf((*StatAllocationComponent)(nil))
	statComp, ok := w.GetComponent(entity, statType)
	if !ok {
		return false
	}

	stats, ok := statComp.(*StatAllocationComponent)
	if !ok || stats.Attributes == nil {
		return false
	}

	return randomValue < stats.Attributes.GetCriticalChance()
}

// RollDodge determines if an attack is dodged based on dexterity.
func (s *System) RollDodge(w *engine.World, entity engine.Entity, randomValue float64) bool {
	statType := reflect.TypeOf((*StatAllocationComponent)(nil))
	statComp, ok := w.GetComponent(entity, statType)
	if !ok {
		return false
	}

	stats, ok := statComp.(*StatAllocationComponent)
	if !ok || stats.Attributes == nil {
		return false
	}

	return randomValue < stats.Attributes.GetDodgeChance()
}

// GetAttackSpeed returns the attack speed multiplier based on dexterity.
func (s *System) GetAttackSpeed(w *engine.World, entity engine.Entity) float64 {
	statType := reflect.TypeOf((*StatAllocationComponent)(nil))
	statComp, ok := w.GetComponent(entity, statType)
	if !ok {
		return 1.0
	}

	stats, ok := statComp.(*StatAllocationComponent)
	if !ok || stats.Attributes == nil {
		return 1.0
	}

	return stats.Attributes.GetAttackSpeedBonus()
}

// GetLootQualityModifier returns the loot quality bonus based on luck.
func (s *System) GetLootQualityModifier(w *engine.World, entity engine.Entity) float64 {
	statType := reflect.TypeOf((*StatAllocationComponent)(nil))
	statComp, ok := w.GetComponent(entity, statType)
	if !ok {
		return 0.0
	}

	stats, ok := statComp.(*StatAllocationComponent)
	if !ok || stats.Attributes == nil {
		return 0.0
	}

	return stats.Attributes.GetLootQualityBonus()
}

// Package stats implements character stat allocation and attribute bonuses.
package stats

import (
	"fmt"
	"sync"
)

// Stat represents a character attribute that can be increased through stat points.
type Stat string

const (
	// StatStrength affects melee damage and weapon carry capacity
	StatStrength Stat = "strength"
	// StatDexterity affects accuracy, dodge chance, and attack speed
	StatDexterity Stat = "dexterity"
	// StatIntelligence affects skill power, hack success, and magic damage
	StatIntelligence Stat = "intelligence"
	// StatVitality affects max health and health regeneration
	StatVitality Stat = "vitality"
	// StatLuck affects critical hit chance, loot quality, and rare encounters
	StatLuck Stat = "luck"
)

// BaseValues defines starting stat values for all characters
const (
	BaseStrength     = 10
	BaseDexterity    = 10
	BaseIntelligence = 10
	BaseVitality     = 10
	BaseLuck         = 10
)

// Attributes holds a character's stat values and available points.
type Attributes struct {
	Strength          int
	Dexterity         int
	Intelligence      int
	Vitality          int
	Luck              int
	UnallocatedPoints int
	mu                sync.RWMutex
}

// NewAttributes creates an Attributes instance with base stats.
func NewAttributes() *Attributes {
	return &Attributes{
		Strength:          BaseStrength,
		Dexterity:         BaseDexterity,
		Intelligence:      BaseIntelligence,
		Vitality:          BaseVitality,
		Luck:              BaseLuck,
		UnallocatedPoints: 0,
	}
}

// Get returns the current value of a stat.
func (a *Attributes) Get(stat Stat) int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	switch stat {
	case StatStrength:
		return a.Strength
	case StatDexterity:
		return a.Dexterity
	case StatIntelligence:
		return a.Intelligence
	case StatVitality:
		return a.Vitality
	case StatLuck:
		return a.Luck
	default:
		return 0
	}
}

// Allocate increases a stat by 1 point if unallocated points are available.
// Returns error if insufficient points or invalid stat.
func (a *Attributes) Allocate(stat Stat) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.UnallocatedPoints <= 0 {
		return fmt.Errorf("no unallocated stat points available")
	}

	switch stat {
	case StatStrength:
		a.Strength++
	case StatDexterity:
		a.Dexterity++
	case StatIntelligence:
		a.Intelligence++
	case StatVitality:
		a.Vitality++
	case StatLuck:
		a.Luck++
	default:
		return fmt.Errorf("invalid stat: %s", stat)
	}

	a.UnallocatedPoints--
	return nil
}

// AddPoints grants unallocated stat points (typically from leveling up).
func (a *Attributes) AddPoints(amount int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.UnallocatedPoints += amount
}

// GetUnallocatedPoints returns the number of unallocated stat points.
func (a *Attributes) GetUnallocatedPoints() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.UnallocatedPoints
}

// GetMeleeDamageBonus calculates melee damage multiplier from strength.
// Formula: 1.0 + (strength - base) * 0.02
func (a *Attributes) GetMeleeDamageBonus() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 1.0 + float64(a.Strength-BaseStrength)*0.02
}

// GetAccuracyBonus calculates accuracy multiplier from dexterity.
// Formula: 1.0 + (dexterity - base) * 0.015
func (a *Attributes) GetAccuracyBonus() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 1.0 + float64(a.Dexterity-BaseDexterity)*0.015
}

// GetSkillPowerBonus calculates skill power multiplier from intelligence.
// Formula: 1.0 + (intelligence - base) * 0.025
func (a *Attributes) GetSkillPowerBonus() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 1.0 + float64(a.Intelligence-BaseIntelligence)*0.025
}

// GetMaxHealthBonus calculates max health bonus from vitality.
// Formula: base_hp * (1.0 + (vitality - base) * 0.05)
func (a *Attributes) GetMaxHealthBonus() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 1.0 + float64(a.Vitality-BaseVitality)*0.05
}

// GetCriticalChance calculates critical hit chance from luck.
// Formula: 5% + (luck - base) * 0.5%
func (a *Attributes) GetCriticalChance() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 0.05 + float64(a.Luck-BaseLuck)*0.005
}

// GetDodgeChance calculates dodge chance from dexterity.
// Formula: (dexterity - base) * 0.3%
func (a *Attributes) GetDodgeChance() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return float64(a.Dexterity-BaseDexterity) * 0.003
}

// GetAttackSpeedBonus calculates attack speed multiplier from dexterity.
// Formula: 1.0 + (dexterity - base) * 0.01
func (a *Attributes) GetAttackSpeedBonus() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 1.0 + float64(a.Dexterity-BaseDexterity)*0.01
}

// GetLootQualityBonus calculates loot quality bonus from luck.
// Formula: (luck - base) * 1%
func (a *Attributes) GetLootQualityBonus() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return float64(a.Luck-BaseLuck) * 0.01
}

// Reset resets all stats to base values and refunds all points.
func (a *Attributes) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()

	totalAllocated := (a.Strength - BaseStrength) +
		(a.Dexterity - BaseDexterity) +
		(a.Intelligence - BaseIntelligence) +
		(a.Vitality - BaseVitality) +
		(a.Luck - BaseLuck)

	a.Strength = BaseStrength
	a.Dexterity = BaseDexterity
	a.Intelligence = BaseIntelligence
	a.Vitality = BaseVitality
	a.Luck = BaseLuck
	a.UnallocatedPoints += totalAllocated
}

// GetAll returns all stat values as a map for serialization or display.
func (a *Attributes) GetAll() map[string]int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]int{
		"strength":     a.Strength,
		"dexterity":    a.Dexterity,
		"intelligence": a.Intelligence,
		"vitality":     a.Vitality,
		"luck":         a.Luck,
		"unallocated":  a.UnallocatedPoints,
	}
}

// StatAllocationComponent is an ECS component that holds character attributes.
type StatAllocationComponent struct {
	Attributes *Attributes
}

// Type implements engine.Component interface.
func (c *StatAllocationComponent) Type() string {
	return "StatAllocation"
}

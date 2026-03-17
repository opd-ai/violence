// Package combat handles damage calculation and combat events.
package combat

import (
	"math/rand"
)

// PhaseTransition defines a boss phase with thresholds and mechanics.
type PhaseTransition struct {
	PhaseID          int
	HealthThreshold  float64 // Enter when health % drops below this
	DamageMultiplier float64
	SpeedMultiplier  float64
	AttackRate       float64
	AbilitySet       []string
	VisualEffect     string
	Enraged          bool
}

// BossPhaseComponent marks an entity as a boss with phase transitions.
type BossPhaseComponent struct {
	CurrentPhase       int
	Phases             []PhaseTransition
	LastTransitionTime float64
	TransitionCooldown float64
	IsTransitioning    bool
	TransitionProgress float64
	InitialMaxHealth   float64
	PhaseChangeCount   int
}

// Type returns the component type identifier.
func (c *BossPhaseComponent) Type() string {
	return "BossPhaseComponent"
}

// BossPhaseSystem manages boss phase transitions based on health.
type BossPhaseSystem struct {
	gameTime float64
}

// NewBossPhaseSystem creates the boss phase system.
func NewBossPhaseSystem() *BossPhaseSystem {
	return &BossPhaseSystem{}
}

// CreateBossPhases generates genre-appropriate boss phases.
func CreateBossPhases(genre string, rng *rand.Rand) []PhaseTransition {
	switch genre {
	case "fantasy":
		return createFantasyBossPhases(rng)
	case "scifi":
		return createSciFiBossPhases(rng)
	case "horror":
		return createHorrorBossPhases(rng)
	case "cyberpunk":
		return createCyberpunkBossPhases(rng)
	default:
		return createDefaultBossPhases(rng)
	}
}

func createFantasyBossPhases(rng *rand.Rand) []PhaseTransition {
	return []PhaseTransition{
		{
			PhaseID:          0,
			HealthThreshold:  1.0,
			DamageMultiplier: 1.0,
			SpeedMultiplier:  1.0,
			AttackRate:       1.0,
			AbilitySet:       []string{"basic_attack", "charge"},
			VisualEffect:     "none",
			Enraged:          false,
		},
		{
			PhaseID:          1,
			HealthThreshold:  0.66,
			DamageMultiplier: 1.2,
			SpeedMultiplier:  1.1,
			AttackRate:       1.2,
			AbilitySet:       []string{"basic_attack", "charge", "whirlwind"},
			VisualEffect:     "red_aura",
			Enraged:          false,
		},
		{
			PhaseID:          2,
			HealthThreshold:  0.33,
			DamageMultiplier: 1.5,
			SpeedMultiplier:  1.25,
			AttackRate:       1.5,
			AbilitySet:       []string{"basic_attack", "charge", "whirlwind", "ground_slam"},
			VisualEffect:     "dark_energy",
			Enraged:          true,
		},
	}
}

func createSciFiBossPhases(rng *rand.Rand) []PhaseTransition {
	return []PhaseTransition{
		{
			PhaseID:          0,
			HealthThreshold:  1.0,
			DamageMultiplier: 1.0,
			SpeedMultiplier:  1.0,
			AttackRate:       1.0,
			AbilitySet:       []string{"laser_burst", "shield_up"},
			VisualEffect:     "blue_shield",
			Enraged:          false,
		},
		{
			PhaseID:          1,
			HealthThreshold:  0.5,
			DamageMultiplier: 1.3,
			SpeedMultiplier:  1.15,
			AttackRate:       1.3,
			AbilitySet:       []string{"laser_burst", "plasma_missile", "teleport"},
			VisualEffect:     "energy_overload",
			Enraged:          false,
		},
		{
			PhaseID:          2,
			HealthThreshold:  0.2,
			DamageMultiplier: 1.8,
			SpeedMultiplier:  1.4,
			AttackRate:       2.0,
			AbilitySet:       []string{"laser_burst", "plasma_missile", "teleport", "orbital_strike"},
			VisualEffect:     "critical_damage",
			Enraged:          true,
		},
	}
}

func createHorrorBossPhases(rng *rand.Rand) []PhaseTransition {
	return []PhaseTransition{
		{
			PhaseID:          0,
			HealthThreshold:  1.0,
			DamageMultiplier: 1.0,
			SpeedMultiplier:  0.8,
			AttackRate:       0.8,
			AbilitySet:       []string{"slow_strike", "intimidate"},
			VisualEffect:     "shadow",
			Enraged:          false,
		},
		{
			PhaseID:          1,
			HealthThreshold:  0.7,
			DamageMultiplier: 1.4,
			SpeedMultiplier:  1.0,
			AttackRate:       1.1,
			AbilitySet:       []string{"slow_strike", "intimidate", "spawn_minions"},
			VisualEffect:     "blood_rage",
			Enraged:          false,
		},
		{
			PhaseID:          2,
			HealthThreshold:  0.4,
			DamageMultiplier: 1.7,
			SpeedMultiplier:  1.2,
			AttackRate:       1.4,
			AbilitySet:       []string{"slow_strike", "frenzy", "spawn_minions", "life_drain"},
			VisualEffect:     "mutated",
			Enraged:          true,
		},
		{
			PhaseID:          3,
			HealthThreshold:  0.15,
			DamageMultiplier: 2.0,
			SpeedMultiplier:  1.5,
			AttackRate:       2.0,
			AbilitySet:       []string{"frenzy", "spawn_minions", "life_drain", "death_curse"},
			VisualEffect:     "desperate",
			Enraged:          true,
		},
	}
}

func createCyberpunkBossPhases(rng *rand.Rand) []PhaseTransition {
	return []PhaseTransition{
		{
			PhaseID:          0,
			HealthThreshold:  1.0,
			DamageMultiplier: 1.0,
			SpeedMultiplier:  1.1,
			AttackRate:       1.0,
			AbilitySet:       []string{"smart_gun", "hack_attempt"},
			VisualEffect:     "neon_blue",
			Enraged:          false,
		},
		{
			PhaseID:          1,
			HealthThreshold:  0.6,
			DamageMultiplier: 1.25,
			SpeedMultiplier:  1.2,
			AttackRate:       1.3,
			AbilitySet:       []string{"smart_gun", "hack_attempt", "deploy_turret"},
			VisualEffect:     "glitch_effect",
			Enraged:          false,
		},
		{
			PhaseID:          2,
			HealthThreshold:  0.25,
			DamageMultiplier: 1.6,
			SpeedMultiplier:  1.35,
			AttackRate:       1.8,
			AbilitySet:       []string{"smart_gun", "deploy_turret", "cyber_overload", "emp_burst"},
			VisualEffect:     "red_alert",
			Enraged:          true,
		},
	}
}

func createDefaultBossPhases(rng *rand.Rand) []PhaseTransition {
	return []PhaseTransition{
		{
			PhaseID:          0,
			HealthThreshold:  1.0,
			DamageMultiplier: 1.0,
			SpeedMultiplier:  1.0,
			AttackRate:       1.0,
			AbilitySet:       []string{"basic_attack"},
			VisualEffect:     "none",
			Enraged:          false,
		},
		{
			PhaseID:          1,
			HealthThreshold:  0.5,
			DamageMultiplier: 1.3,
			SpeedMultiplier:  1.15,
			AttackRate:       1.3,
			AbilitySet:       []string{"basic_attack", "power_attack"},
			VisualEffect:     "glow",
			Enraged:          false,
		},
		{
			PhaseID:          2,
			HealthThreshold:  0.2,
			DamageMultiplier: 1.7,
			SpeedMultiplier:  1.3,
			AttackRate:       1.7,
			AbilitySet:       []string{"basic_attack", "power_attack", "ultimate"},
			VisualEffect:     "critical",
			Enraged:          true,
		},
	}
}

// GetCurrentPhaseData returns the active phase configuration.
func (c *BossPhaseComponent) GetCurrentPhaseData() *PhaseTransition {
	if c.CurrentPhase >= 0 && c.CurrentPhase < len(c.Phases) {
		return &c.Phases[c.CurrentPhase]
	}
	return nil
}

// ShouldTransition checks if boss should transition to next phase.
func (c *BossPhaseComponent) ShouldTransition(currentHealth, maxHealth, gameTime float64) bool {
	if c.IsTransitioning {
		return false
	}

	if gameTime-c.LastTransitionTime < c.TransitionCooldown {
		return false
	}

	healthPct := currentHealth / maxHealth
	nextPhase := c.CurrentPhase + 1

	if nextPhase >= len(c.Phases) {
		return false
	}

	return healthPct <= c.Phases[nextPhase].HealthThreshold
}

// StartTransition initiates a phase transition.
func (c *BossPhaseComponent) StartTransition(gameTime float64) {
	c.IsTransitioning = true
	c.TransitionProgress = 0.0
	c.LastTransitionTime = gameTime
}

// UpdateTransition advances the transition animation.
func (c *BossPhaseComponent) UpdateTransition(deltaTime float64) bool {
	if !c.IsTransitioning {
		return false
	}

	c.TransitionProgress += deltaTime * 2.0 // 0.5 second transition

	if c.TransitionProgress >= 1.0 {
		c.IsTransitioning = false
		c.TransitionProgress = 1.0
		c.CurrentPhase++
		c.PhaseChangeCount++
		return true
	}

	return false
}

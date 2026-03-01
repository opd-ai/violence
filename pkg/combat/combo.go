// Package combat - Combo system for chained weapon attacks
package combat

import (
	"math/rand"
)

// ComboState tracks the current position in a combo chain.
type ComboState int

const (
	ComboStateNone ComboState = iota
	ComboStateActive
	ComboStateBroken
)

// ComboStep defines one attack in a combo chain.
type ComboStep struct {
	Name          string
	DamageMul     float64 // Multiplier applied to base weapon damage
	SpeedMul      float64 // Attack speed multiplier
	RangeMul      float64 // Range multiplier
	KnockbackMul  float64 // Knockback force multiplier
	WindowStart   float64 // Input timing window start (seconds after previous hit)
	WindowEnd     float64 // Input timing window end (seconds after previous hit)
	StaggerChance float64 // Probability to stagger enemy (0-1)
	ParticleCount int     // Number of hit particles to spawn
	ScreenShake   float64 // Screen shake intensity
}

// ComboChain defines a sequence of attacks that can be performed.
type ComboChain struct {
	ID          string
	Name        string
	Steps       []ComboStep
	GenreID     string
	WeaponType  string // "melee", "ranged", "energy", "heavy"
	Description string
}

// ComboComponent tracks an entity's combo state.
type ComboComponent struct {
	ChainID       string
	CurrentStep   int
	State         ComboState
	TimeSinceHit  float64
	TotalHits     int
	TotalDamage   float64
	HighestCombo  int
	InputBuffer   bool // True if input received during window
	LastDirection [2]float64
}

// Type implements Component interface.
func (c *ComboComponent) Type() string {
	return "combo"
}

// DefaultChains returns genre-appropriate combo chains.
func DefaultChains(genreID string, rng *rand.Rand) []ComboChain {
	chains := []ComboChain{}

	switch genreID {
	case "fantasy":
		chains = append(chains, ComboChain{
			ID:          "sword_basic",
			Name:        "Blade Dance",
			GenreID:     genreID,
			WeaponType:  "melee",
			Description: "Swift three-hit sword combo",
			Steps: []ComboStep{
				{
					Name:          "Slash",
					DamageMul:     1.0,
					SpeedMul:      1.0,
					RangeMul:      1.0,
					KnockbackMul:  0.5,
					WindowStart:   0.0,
					WindowEnd:     0.6,
					StaggerChance: 0.1,
					ParticleCount: 3,
					ScreenShake:   0.02,
				},
				{
					Name:          "Cross Cut",
					DamageMul:     1.2,
					SpeedMul:      1.1,
					RangeMul:      1.1,
					KnockbackMul:  0.7,
					WindowStart:   0.1,
					WindowEnd:     0.5,
					StaggerChance: 0.2,
					ParticleCount: 5,
					ScreenShake:   0.03,
				},
				{
					Name:          "Finishing Thrust",
					DamageMul:     1.5,
					SpeedMul:      0.9,
					RangeMul:      1.3,
					KnockbackMul:  1.5,
					WindowStart:   0.15,
					WindowEnd:     0.55,
					StaggerChance: 0.4,
					ParticleCount: 8,
					ScreenShake:   0.05,
				},
			},
		})
		chains = append(chains, ComboChain{
			ID:          "magic_burst",
			Name:        "Arcane Barrage",
			GenreID:     genreID,
			WeaponType:  "energy",
			Description: "Rapid magical projectile combo",
			Steps: []ComboStep{
				{
					Name:          "Spark",
					DamageMul:     0.8,
					SpeedMul:      1.3,
					RangeMul:      1.0,
					KnockbackMul:  0.3,
					WindowStart:   0.0,
					WindowEnd:     0.4,
					StaggerChance: 0.05,
					ParticleCount: 4,
					ScreenShake:   0.01,
				},
				{
					Name:          "Bolt",
					DamageMul:     1.0,
					SpeedMul:      1.2,
					RangeMul:      1.2,
					KnockbackMul:  0.5,
					WindowStart:   0.05,
					WindowEnd:     0.35,
					StaggerChance: 0.1,
					ParticleCount: 6,
					ScreenShake:   0.02,
				},
				{
					Name:          "Blast",
					DamageMul:     1.8,
					SpeedMul:      0.8,
					RangeMul:      1.5,
					KnockbackMul:  1.2,
					WindowStart:   0.1,
					WindowEnd:     0.5,
					StaggerChance: 0.3,
					ParticleCount: 12,
					ScreenShake:   0.04,
				},
			},
		})

	case "scifi":
		chains = append(chains, ComboChain{
			ID:          "pulse_burst",
			Name:        "Pulse Cascade",
			GenreID:     genreID,
			WeaponType:  "energy",
			Description: "Energy weapon rapid fire sequence",
			Steps: []ComboStep{
				{
					Name:          "Pulse",
					DamageMul:     0.9,
					SpeedMul:      1.4,
					RangeMul:      1.0,
					KnockbackMul:  0.4,
					WindowStart:   0.0,
					WindowEnd:     0.3,
					StaggerChance: 0.08,
					ParticleCount: 3,
					ScreenShake:   0.015,
				},
				{
					Name:          "Double Pulse",
					DamageMul:     1.1,
					SpeedMul:      1.3,
					RangeMul:      1.1,
					KnockbackMul:  0.6,
					WindowStart:   0.03,
					WindowEnd:     0.25,
					StaggerChance: 0.15,
					ParticleCount: 6,
					ScreenShake:   0.025,
				},
				{
					Name:          "Overcharge",
					DamageMul:     2.0,
					SpeedMul:      0.7,
					RangeMul:      1.4,
					KnockbackMul:  1.8,
					WindowStart:   0.08,
					WindowEnd:     0.45,
					StaggerChance: 0.5,
					ParticleCount: 15,
					ScreenShake:   0.06,
				},
			},
		})

	case "horror":
		chains = append(chains, ComboChain{
			ID:          "brutal_assault",
			Name:        "Savage Frenzy",
			GenreID:     genreID,
			WeaponType:  "melee",
			Description: "Desperate close-quarters barrage",
			Steps: []ComboStep{
				{
					Name:          "Wild Swing",
					DamageMul:     1.1,
					SpeedMul:      1.2,
					RangeMul:      0.9,
					KnockbackMul:  0.6,
					WindowStart:   0.0,
					WindowEnd:     0.5,
					StaggerChance: 0.15,
					ParticleCount: 4,
					ScreenShake:   0.03,
				},
				{
					Name:          "Crushing Blow",
					DamageMul:     1.4,
					SpeedMul:      0.9,
					RangeMul:      1.0,
					KnockbackMul:  1.0,
					WindowStart:   0.12,
					WindowEnd:     0.55,
					StaggerChance: 0.3,
					ParticleCount: 7,
					ScreenShake:   0.04,
				},
				{
					Name:          "Execution",
					DamageMul:     2.2,
					SpeedMul:      0.7,
					RangeMul:      1.1,
					KnockbackMul:  2.0,
					WindowStart:   0.2,
					WindowEnd:     0.65,
					StaggerChance: 0.6,
					ParticleCount: 10,
					ScreenShake:   0.07,
				},
			},
		})

	case "cyberpunk":
		chains = append(chains, ComboChain{
			ID:          "tech_combo",
			Name:        "Cyber Assault",
			GenreID:     genreID,
			WeaponType:  "ranged",
			Description: "High-tech weapon combination attack",
			Steps: []ComboStep{
				{
					Name:          "Smart Shot",
					DamageMul:     1.0,
					SpeedMul:      1.3,
					RangeMul:      1.2,
					KnockbackMul:  0.4,
					WindowStart:   0.0,
					WindowEnd:     0.35,
					StaggerChance: 0.1,
					ParticleCount: 4,
					ScreenShake:   0.02,
				},
				{
					Name:          "EMP Burst",
					DamageMul:     1.3,
					SpeedMul:      1.0,
					RangeMul:      1.0,
					KnockbackMul:  0.8,
					WindowStart:   0.08,
					WindowEnd:     0.4,
					StaggerChance: 0.35,
					ParticleCount: 8,
					ScreenShake:   0.035,
				},
				{
					Name:          "Railgun Finish",
					DamageMul:     2.5,
					SpeedMul:      0.6,
					RangeMul:      2.0,
					KnockbackMul:  2.5,
					WindowStart:   0.15,
					WindowEnd:     0.6,
					StaggerChance: 0.7,
					ParticleCount: 20,
					ScreenShake:   0.08,
				},
			},
		})

	default: // postapoc or fallback
		chains = append(chains, ComboChain{
			ID:          "scrap_combo",
			Name:        "Scrap Beatdown",
			GenreID:     genreID,
			WeaponType:  "melee",
			Description: "Makeshift weapon combo",
			Steps: []ComboStep{
				{
					Name:          "Jab",
					DamageMul:     0.9,
					SpeedMul:      1.2,
					RangeMul:      0.95,
					KnockbackMul:  0.5,
					WindowStart:   0.0,
					WindowEnd:     0.5,
					StaggerChance: 0.12,
					ParticleCount: 3,
					ScreenShake:   0.025,
				},
				{
					Name:          "Smash",
					DamageMul:     1.3,
					SpeedMul:      0.95,
					RangeMul:      1.05,
					KnockbackMul:  0.9,
					WindowStart:   0.1,
					WindowEnd:     0.5,
					StaggerChance: 0.25,
					ParticleCount: 6,
					ScreenShake:   0.04,
				},
				{
					Name:          "Overhead Crush",
					DamageMul:     1.9,
					SpeedMul:      0.75,
					RangeMul:      1.15,
					KnockbackMul:  1.6,
					WindowStart:   0.18,
					WindowEnd:     0.6,
					StaggerChance: 0.45,
					ParticleCount: 9,
					ScreenShake:   0.06,
				},
			},
		})
	}

	return chains
}

// SelectChain picks an appropriate combo chain based on weapon type and distance.
func SelectChain(chains []ComboChain, weaponType string, rng *rand.Rand) *ComboChain {
	candidates := []ComboChain{}
	for _, chain := range chains {
		if chain.WeaponType == weaponType {
			candidates = append(candidates, chain)
		}
	}

	if len(candidates) == 0 {
		// Fallback to first available chain
		if len(chains) > 0 {
			return &chains[0]
		}
		return nil
	}

	// Select randomly if rng provided, otherwise first match
	idx := 0
	if rng != nil {
		idx = rng.Intn(len(candidates))
	}
	return &candidates[idx]
}

// ResetCombo resets a combo component to initial state.
func ResetCombo(combo *ComboComponent) {
	combo.State = ComboStateNone
	combo.CurrentStep = 0
	combo.TimeSinceHit = 0
	combo.InputBuffer = false
}

// AdvanceCombo moves to the next step in the combo chain.
func AdvanceCombo(combo *ComboComponent, chain *ComboChain) bool {
	if combo.CurrentStep >= len(chain.Steps) {
		return false
	}

	combo.CurrentStep++
	if combo.CurrentStep >= len(chain.Steps) {
		// Combo complete
		combo.State = ComboStateNone
		combo.CurrentStep = 0
		return false
	}

	combo.State = ComboStateActive
	combo.TimeSinceHit = 0
	combo.InputBuffer = false
	return true
}

// BreakCombo interrupts the current combo.
func BreakCombo(combo *ComboComponent) {
	combo.State = ComboStateBroken
	combo.TimeSinceHit = 0
	combo.InputBuffer = false
	// Track highest combo reached
	if combo.CurrentStep > combo.HighestCombo {
		combo.HighestCombo = combo.CurrentStep
	}
}

// Package combat - Defensive maneuvers (dodge, parry, block)
package combat

import (
	"math"
)

// DefenseType defines the type of defensive maneuver.
type DefenseType int

const (
	DefenseNone DefenseType = iota
	DefenseDodge
	DefenseParry
	DefenseBlock
)

// DefenseState represents the state of a defensive maneuver.
type DefenseState int

const (
	DefenseInactive DefenseState = iota
	DefenseWindup                // Startup frames before active
	DefenseActive                // Active defense window
	DefenseRecovery              // Post-action cooldown
)

// DefenseComponent stores defensive maneuver state.
type DefenseComponent struct {
	Type           DefenseType
	State          DefenseState
	StateTimer     float64
	StaminaCost    float64
	StaminaCurrent float64
	StaminaMax     float64
	StaminaRegen   float64

	// Dodge-specific
	DodgeDistance    float64
	DodgeDirection   float64
	DodgeIFrames     float64
	DodgeIFrameTimer float64
	DodgeVelocityX   float64
	DodgeVelocityY   float64

	// Parry-specific
	ParryWindowStart   float64
	ParryWindowEnd     float64
	ParryStunDuration  float64
	ParryCounterMul    float64
	ParryPerfectWindow float64

	// Block-specific
	BlockDamageReduction float64
	BlockStaminaDrain    float64
	BlockArc             float64

	// Cooldowns
	DodgeCooldown float64
	ParryCooldown float64
	BlockCooldown float64
	CooldownTimer float64

	// Buffs from successful defenses
	PerfectParryBuff  bool
	PerfectParryTimer float64
	BackstabWindow    bool
	BackstabTimer     float64
}

// DefensePreset holds genre-specific defense parameters.
type DefensePreset struct {
	DodgeDistance   float64
	DodgeIFrames    float64
	DodgeCost       float64
	ParryWindow     float64
	ParryPerfect    float64
	ParryCounterMul float64
	ParryCost       float64
	BlockReduction  float64
	BlockDrain      float64
	BlockCost       float64
	StaminaMax      float64
	StaminaRegen    float64
	DodgeCooldown   float64
	ParryCooldown   float64
	BlockCooldown   float64
}

// GetDefensePreset returns genre-appropriate defense parameters.
func GetDefensePreset(genreID string) DefensePreset {
	presets := map[string]DefensePreset{
		"fantasy": {
			DodgeDistance:   3.0,
			DodgeIFrames:    0.25,
			DodgeCost:       20.0,
			ParryWindow:     0.3,
			ParryPerfect:    0.1,
			ParryCounterMul: 2.0,
			ParryCost:       15.0,
			BlockReduction:  0.7,
			BlockDrain:      5.0,
			BlockCost:       10.0,
			StaminaMax:      100.0,
			StaminaRegen:    10.0,
			DodgeCooldown:   0.8,
			ParryCooldown:   0.5,
			BlockCooldown:   0.3,
		},
		"scifi": {
			DodgeDistance:   4.0,
			DodgeIFrames:    0.2,
			DodgeCost:       25.0,
			ParryWindow:     0.25,
			ParryPerfect:    0.08,
			ParryCounterMul: 1.8,
			ParryCost:       20.0,
			BlockReduction:  0.6,
			BlockDrain:      8.0,
			BlockCost:       12.0,
			StaminaMax:      120.0,
			StaminaRegen:    12.0,
			DodgeCooldown:   0.6,
			ParryCooldown:   0.6,
			BlockCooldown:   0.4,
		},
		"horror": {
			DodgeDistance:   2.5,
			DodgeIFrames:    0.3,
			DodgeCost:       30.0,
			ParryWindow:     0.35,
			ParryPerfect:    0.12,
			ParryCounterMul: 2.5,
			ParryCost:       25.0,
			BlockReduction:  0.5,
			BlockDrain:      10.0,
			BlockCost:       15.0,
			StaminaMax:      80.0,
			StaminaRegen:    8.0,
			DodgeCooldown:   1.0,
			ParryCooldown:   0.7,
			BlockCooldown:   0.5,
		},
		"cyberpunk": {
			DodgeDistance:   3.5,
			DodgeIFrames:    0.18,
			DodgeCost:       22.0,
			ParryWindow:     0.22,
			ParryPerfect:    0.07,
			ParryCounterMul: 1.5,
			ParryCost:       18.0,
			BlockReduction:  0.65,
			BlockDrain:      6.0,
			BlockCost:       10.0,
			StaminaMax:      110.0,
			StaminaRegen:    14.0,
			DodgeCooldown:   0.5,
			ParryCooldown:   0.5,
			BlockCooldown:   0.35,
		},
	}

	preset, ok := presets[genreID]
	if !ok {
		return presets["fantasy"]
	}
	return preset
}

// NewDefenseComponent creates a defense component with genre preset.
func NewDefenseComponent(genreID string) *DefenseComponent {
	preset := GetDefensePreset(genreID)
	return &DefenseComponent{
		Type:                 DefenseNone,
		State:                DefenseInactive,
		StaminaCurrent:       preset.StaminaMax,
		StaminaMax:           preset.StaminaMax,
		StaminaRegen:         preset.StaminaRegen,
		DodgeDistance:        preset.DodgeDistance,
		DodgeIFrames:         preset.DodgeIFrames,
		ParryWindowStart:     0.0,
		ParryWindowEnd:       preset.ParryWindow,
		ParryPerfectWindow:   preset.ParryPerfect,
		ParryCounterMul:      preset.ParryCounterMul,
		ParryStunDuration:    0.5,
		BlockDamageReduction: preset.BlockReduction,
		BlockStaminaDrain:    preset.BlockDrain,
		BlockArc:             math.Pi * 0.75,
		DodgeCooldown:        preset.DodgeCooldown,
		ParryCooldown:        preset.ParryCooldown,
		BlockCooldown:        preset.BlockCooldown,
	}
}

// CanDodge checks if dodge is available.
func (d *DefenseComponent) CanDodge() bool {
	if d.State != DefenseInactive {
		return false
	}
	if d.CooldownTimer > 0 {
		return false
	}
	preset := GetDefensePreset("fantasy")
	return d.StaminaCurrent >= preset.DodgeCost
}

// CanParry checks if parry is available.
func (d *DefenseComponent) CanParry() bool {
	if d.State != DefenseInactive {
		return false
	}
	if d.CooldownTimer > 0 {
		return false
	}
	preset := GetDefensePreset("fantasy")
	return d.StaminaCurrent >= preset.ParryCost
}

// CanBlock checks if block is available.
func (d *DefenseComponent) CanBlock() bool {
	preset := GetDefensePreset("fantasy")
	return d.StaminaCurrent >= preset.BlockCost
}

// IsInvulnerable checks if entity is currently invulnerable (during dodge i-frames).
func (d *DefenseComponent) IsInvulnerable() bool {
	return d.Type == DefenseDodge && d.DodgeIFrameTimer > 0
}

// IsBlocking checks if entity is actively blocking.
func (d *DefenseComponent) IsBlocking() bool {
	return d.Type == DefenseBlock && d.State == DefenseActive
}

// IsParrying checks if entity is in parry window.
func (d *DefenseComponent) IsParrying() bool {
	return d.Type == DefenseParry && d.State == DefenseActive
}

// IsPerfectParryWindow checks if entity is in perfect parry timing window.
func (d *DefenseComponent) IsPerfectParryWindow() bool {
	if d.Type != DefenseParry || d.State != DefenseActive {
		return false
	}
	return d.StateTimer <= d.ParryPerfectWindow
}

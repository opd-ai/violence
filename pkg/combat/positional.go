// Package combat - Positional advantage component
package combat

import "math"

// PositionalComponent stores entity facing direction and height.
type PositionalComponent struct {
	// Facing angle in radians (0 = right, π/2 = down, π = left, 3π/2 = up)
	FacingAngle float64
	// Height/elevation above ground (for verticality advantage)
	Height float64
	// Last update time for interpolation
	LastUpdate float64
}

// Type returns the component type identifier.
func (c *PositionalComponent) Type() string {
	return "PositionalComponent"
}

// SetFacingFromDirection updates facing angle from direction vector.
func (c *PositionalComponent) SetFacingFromDirection(dx, dy float64) {
	if math.Abs(dx) > 0.01 || math.Abs(dy) > 0.01 {
		c.FacingAngle = math.Atan2(dy, dx)
	}
}

// GetFacingVector returns unit vector for current facing.
func (c *PositionalComponent) GetFacingVector() (float64, float64) {
	return math.Cos(c.FacingAngle), math.Sin(c.FacingAngle)
}

// PositionalAdvantage defines positional combat modifiers.
type PositionalAdvantage int

const (
	AdvantageFrontal   PositionalAdvantage = iota // Normal attack from front
	AdvantageFlank                                // Attack from side (45°-135° off-axis)
	AdvantageBackstab                             // Attack from behind (within 45° of back)
	AdvantageElevation                            // Height advantage
)

// PositionalConfig holds angle thresholds and damage multipliers.
type PositionalConfig struct {
	BackstabAngle       float64 // Radians from directly behind to count as backstab
	FlankAngle          float64 // Radians from side to count as flank
	BackstabMultiplier  float64
	FlankMultiplier     float64
	ElevationThreshold  float64 // Height difference for elevation bonus
	ElevationMultiplier float64
}

// GetPositionalConfig returns genre-specific positional parameters.
func GetPositionalConfig(genreID string) PositionalConfig {
	configs := map[string]PositionalConfig{
		"fantasy": {
			BackstabAngle:       math.Pi / 4,     // 45° cone behind target
			FlankAngle:          math.Pi * 3 / 8, // ~67.5° cone on sides
			BackstabMultiplier:  2.0,
			FlankMultiplier:     1.5,
			ElevationThreshold:  2.0,
			ElevationMultiplier: 1.3,
		},
		"scifi": {
			BackstabAngle:       math.Pi / 3, // 60° - sensors reduce vulnerability
			FlankAngle:          math.Pi / 3,
			BackstabMultiplier:  1.8,
			FlankMultiplier:     1.4,
			ElevationThreshold:  3.0,
			ElevationMultiplier: 1.2,
		},
		"horror": {
			BackstabAngle:       math.Pi / 3,
			FlankAngle:          math.Pi / 2.5,
			BackstabMultiplier:  2.5, // Horror emphasizes vulnerability
			FlankMultiplier:     1.6,
			ElevationThreshold:  1.5,
			ElevationMultiplier: 1.25,
		},
		"cyberpunk": {
			BackstabAngle:       math.Pi / 3.5,
			FlankAngle:          math.Pi / 3,
			BackstabMultiplier:  2.2,
			FlankMultiplier:     1.5,
			ElevationThreshold:  2.5,
			ElevationMultiplier: 1.3,
		},
	}

	if cfg, ok := configs[genreID]; ok {
		return cfg
	}
	return configs["fantasy"]
}

// CalculatePositionalAdvantage determines advantage type and multiplier.
// attackerX/Y: attacker position
// targetX/Y: target position
// attackerPos: attacker PositionalComponent (can be nil if no facing data)
// targetPos: target PositionalComponent (required for facing calculation)
// cfg: genre config
func CalculatePositionalAdvantage(
	attackerX, attackerY, targetX, targetY float64,
	attackerPos, targetPos *PositionalComponent,
	cfg PositionalConfig,
) (PositionalAdvantage, float64) {
	if targetPos == nil {
		return AdvantageFrontal, 1.0
	}

	// Vector from target to attacker
	dx := attackerX - targetX
	dy := attackerY - targetY
	attackAngle := math.Atan2(dy, dx)

	// Angle difference between attack direction and target facing
	angleDiff := normalizeAngle(attackAngle - targetPos.FacingAngle)

	// Determine positional advantage
	multiplier := 1.0
	advantage := AdvantageFrontal

	// Check backstab (attack from behind)
	if math.Abs(angleDiff-math.Pi) < cfg.BackstabAngle {
		advantage = AdvantageBackstab
		multiplier = cfg.BackstabMultiplier
	} else if math.Abs(angleDiff-math.Pi/2) < cfg.FlankAngle || math.Abs(angleDiff+math.Pi/2) < cfg.FlankAngle {
		// Check flank (attack from sides)
		advantage = AdvantageFlank
		multiplier = cfg.FlankMultiplier
	}

	// Apply elevation bonus if both have height data
	if attackerPos != nil && targetPos != nil {
		heightDiff := attackerPos.Height - targetPos.Height
		if heightDiff >= cfg.ElevationThreshold {
			multiplier *= cfg.ElevationMultiplier
			if advantage == AdvantageFrontal {
				advantage = AdvantageElevation
			}
		}
	}

	return advantage, multiplier
}

// ApplyPositionalDamage calculates final damage with positional modifiers.
func ApplyPositionalDamage(baseDamage float64, advantage PositionalAdvantage, multiplier float64) float64 {
	return baseDamage * multiplier
}

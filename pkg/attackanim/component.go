// Package attackanim provides attack animation components for visual attack variety.
package attackanim

// Component stores attack animation state for an entity.
type Component struct {
	// Current attack animation state
	AttackState   AttackState // idle, windup, strike, recovery
	AnimationType string      // "melee_slash", "overhead_smash", "lunge", "ranged_charge", "spin_attack"

	// Timing
	StateTime    float64 // Time spent in current state
	WindupTime   float64 // Duration of windup phase
	StrikeTime   float64 // Duration of strike phase
	RecoveryTime float64 // Duration of recovery phase

	// Visual parameters
	RotationAngle float64 // Weapon/limb rotation during animation
	OffsetX       float64 // Body offset X during animation
	OffsetY       float64 // Body offset Y during animation
	SquashStretch float64 // Sprite squash/stretch factor

	// Attack direction
	TargetDirX float64 // Direction of attack
	TargetDirY float64 // Direction of attack

	// Animation intensity (0-1, affects visual exaggeration)
	Intensity float64
}

// AttackState represents the current animation phase.
type AttackState int

const (
	StateIdle     AttackState = iota // StateIdle is the idle attack state.
	StateWindup                      // StateWindup is the windup attack state.
	StateStrike                      // StateStrike is the strike attack state.
	StateRecovery                    // StateRecovery is the recovery attack state.
)

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "AttackAnimation"
}

// IsAnimating returns true if currently in an attack animation.
func (c *Component) IsAnimating() bool {
	return c.AttackState != StateIdle
}

// AnimationProgress returns normalized progress (0-1) through current state.
func (c *Component) AnimationProgress() float64 {
	switch c.AttackState {
	case StateWindup:
		if c.WindupTime > 0 {
			return c.StateTime / c.WindupTime
		}
	case StateStrike:
		if c.StrikeTime > 0 {
			return c.StateTime / c.StrikeTime
		}
	case StateRecovery:
		if c.RecoveryTime > 0 {
			return c.StateTime / c.RecoveryTime
		}
	}
	return 0
}

// Package weaponsway provides first-person weapon sway animation for realistic FPS feel.
// Weapon sway simulates the inertia and weight of held weapons, making movement
// and aiming feel organic rather than robotically locked to the camera.
package weaponsway

// Component holds weapon sway animation state for first-person view.
// It tracks current offset, velocity, and target positions to produce
// smooth, physically-motivated weapon movement.
type Component struct {
	// OffsetX is the current horizontal sway offset in screen pixels.
	OffsetX float64
	// OffsetY is the current vertical sway offset in screen pixels.
	OffsetY float64

	// VelocityX is the horizontal sway velocity for momentum simulation.
	VelocityX float64
	// VelocityY is the vertical sway velocity for momentum simulation.
	VelocityY float64

	// TargetX is the target horizontal offset (for smooth interpolation).
	TargetX float64
	// TargetY is the target vertical offset (for smooth interpolation).
	TargetY float64

	// BreathPhase tracks idle breathing animation phase (radians).
	BreathPhase float64
	// MovementPhase tracks movement bob animation phase (radians).
	MovementPhase float64

	// LastCameraYaw stores previous frame's camera yaw for delta calculation.
	LastCameraYaw float64
	// LastCameraPitch stores previous frame's camera pitch for delta calculation.
	LastCameraPitch float64

	// WeaponWeight affects sway magnitude (0.5 = light, 1.0 = medium, 2.0 = heavy).
	WeaponWeight float64
	// RecoverySpeed controls how fast the weapon returns to center (per second).
	RecoverySpeed float64

	// IsMoving indicates if the entity is currently moving (for movement sway).
	IsMoving bool
	// IsSprinting indicates if the entity is sprinting (increased sway).
	IsSprinting bool
	// IsAiming indicates ADS mode (reduced sway, faster recovery).
	IsAiming bool

	// Enabled allows disabling sway (e.g., during cutscenes).
	Enabled bool
}

// Type returns the ECS component type identifier.
func (c *Component) Type() string {
	return "weapon_sway"
}

// NewComponent creates a weapon sway component with default values.
func NewComponent() *Component {
	return &Component{
		OffsetX:       0,
		OffsetY:       0,
		VelocityX:     0,
		VelocityY:     0,
		TargetX:       0,
		TargetY:       0,
		BreathPhase:   0,
		MovementPhase: 0,
		WeaponWeight:  1.0,
		RecoverySpeed: 8.0,
		Enabled:       true,
	}
}

// SetWeaponWeight configures sway based on weapon type.
// Light weapons (pistols, knives): 0.5-0.7
// Medium weapons (rifles, SMGs): 1.0
// Heavy weapons (LMGs, launchers): 1.5-2.0
func (c *Component) SetWeaponWeight(weight float64) {
	if weight < 0.1 {
		weight = 0.1
	}
	if weight > 3.0 {
		weight = 3.0
	}
	c.WeaponWeight = weight
}

// GetSwayOffset returns the current combined sway offset (x, y) in screen pixels.
// This is the final value that should be applied to weapon rendering position.
func (c *Component) GetSwayOffset() (float64, float64) {
	if !c.Enabled {
		return 0, 0
	}
	return c.OffsetX, c.OffsetY
}

// Reset clears all sway state (e.g., on weapon switch or respawn).
func (c *Component) Reset() {
	c.OffsetX = 0
	c.OffsetY = 0
	c.VelocityX = 0
	c.VelocityY = 0
	c.TargetX = 0
	c.TargetY = 0
	c.BreathPhase = 0
	c.MovementPhase = 0
}

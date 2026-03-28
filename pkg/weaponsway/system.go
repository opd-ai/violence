package weaponsway

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// SwayConfig holds genre-specific sway parameters.
type SwayConfig struct {
	// TurnSwayMultiplier scales the sway from camera rotation (genre aesthetic).
	TurnSwayMultiplier float64
	// MovementBobAmplitude is the maximum movement bob offset in pixels.
	MovementBobAmplitude float64
	// MovementBobFrequency is the bob oscillation speed in Hz.
	MovementBobFrequency float64
	// BreathAmplitude is the idle breathing sway magnitude.
	BreathAmplitude float64
	// BreathFrequency is the idle breathing oscillation speed in Hz.
	BreathFrequency float64
	// MaxSwayOffset clamps the maximum sway distance from center.
	MaxSwayOffset float64
	// Damping reduces velocity per frame (0-1, higher = more damping).
	Damping float64
	// SpringStiffness controls return-to-center spring force.
	SpringStiffness float64
}

// System updates weapon sway components each frame.
type System struct {
	logger *logrus.Entry
	config SwayConfig
	genre  string
}

// NewSystem creates a weapon sway system with default configuration.
func NewSystem(genreID string) *System {
	s := &System{
		logger: logrus.WithFields(logrus.Fields{
			"system": "weaponsway",
		}),
		genre: genreID,
	}
	s.applyGenreConfig(genreID)
	return s
}

// SetGenre updates the sway configuration for a new genre.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
	s.applyGenreConfig(genreID)
}

// applyGenreConfig sets sway parameters based on genre aesthetic.
func (s *System) applyGenreConfig(genreID string) {
	switch genreID {
	case "fantasy":
		// Heavy, weighty weapons with dramatic sway
		s.config = SwayConfig{
			TurnSwayMultiplier:   2.5,
			MovementBobAmplitude: 6.0,
			MovementBobFrequency: 4.0,
			BreathAmplitude:      1.5,
			BreathFrequency:      0.8,
			MaxSwayOffset:        25.0,
			Damping:              0.85,
			SpringStiffness:      5.0,
		}
	case "scifi":
		// Lighter, more responsive futuristic weapons
		s.config = SwayConfig{
			TurnSwayMultiplier:   1.8,
			MovementBobAmplitude: 4.0,
			MovementBobFrequency: 5.0,
			BreathAmplitude:      0.8,
			BreathFrequency:      1.0,
			MaxSwayOffset:        18.0,
			Damping:              0.75,
			SpringStiffness:      7.0,
		}
	case "horror":
		// Shaky, nervous handling with exaggerated sway
		s.config = SwayConfig{
			TurnSwayMultiplier:   3.0,
			MovementBobAmplitude: 7.0,
			MovementBobFrequency: 3.5,
			BreathAmplitude:      2.5,
			BreathFrequency:      1.2,
			MaxSwayOffset:        30.0,
			Damping:              0.80,
			SpringStiffness:      4.0,
		}
	case "cyberpunk":
		// Snappy, responsive with quick recovery
		s.config = SwayConfig{
			TurnSwayMultiplier:   2.0,
			MovementBobAmplitude: 4.5,
			MovementBobFrequency: 5.5,
			BreathAmplitude:      1.0,
			BreathFrequency:      0.9,
			MaxSwayOffset:        20.0,
			Damping:              0.70,
			SpringStiffness:      8.0,
		}
	case "postapoc":
		// Rough, improvised weapons with irregular sway
		s.config = SwayConfig{
			TurnSwayMultiplier:   2.8,
			MovementBobAmplitude: 6.5,
			MovementBobFrequency: 3.8,
			BreathAmplitude:      1.8,
			BreathFrequency:      0.7,
			MaxSwayOffset:        28.0,
			Damping:              0.82,
			SpringStiffness:      4.5,
		}
	default:
		// Fallback to fantasy
		s.config = SwayConfig{
			TurnSwayMultiplier:   2.5,
			MovementBobAmplitude: 6.0,
			MovementBobFrequency: 4.0,
			BreathAmplitude:      1.5,
			BreathFrequency:      0.8,
			MaxSwayOffset:        25.0,
			Damping:              0.85,
			SpringStiffness:      5.0,
		}
	}
}

// Update processes all weapon sway components.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime
	if deltaTime <= 0 {
		return
	}

	compType := reflect.TypeOf(&Component{})
	for _, entity := range w.Query(compType) {
		comp, ok := w.GetComponent(entity, compType)
		if !ok {
			continue
		}
		sway := comp.(*Component)
		if !sway.Enabled {
			continue
		}
		s.updateSway(sway, deltaTime)
	}
}

// updateSway processes a single weapon sway component.
func (s *System) updateSway(sway *Component, dt float64) {
	// Apply turn-based sway (inertia from camera movement)
	s.applyTurnSway(sway, dt)

	// Apply movement bob if moving
	s.applyMovementBob(sway, dt)

	// Apply idle breathing sway
	s.applyBreathSway(sway, dt)

	// Apply physics: spring + damping
	s.applySpringPhysics(sway, dt)

	// Clamp to maximum offset
	s.clampOffset(sway)
}

// applyTurnSway adds sway from camera rotation.
// Weapon moves opposite to camera turn direction, simulating inertia.
func (s *System) applyTurnSway(sway *Component, dt float64) {
	// Turn sway impulse is applied externally via AddTurnImpulse
	// Here we just ensure the target follows physics
}

// applyMovementBob adds vertical/horizontal bob during movement.
func (s *System) applyMovementBob(sway *Component, dt float64) {
	if !sway.IsMoving {
		return
	}

	// Advance movement phase
	freq := s.config.MovementBobFrequency
	if sway.IsSprinting {
		freq *= 1.5
	}
	sway.MovementPhase += dt * freq * 2.0 * math.Pi

	// Normalize phase to prevent overflow
	if sway.MovementPhase > 2.0*math.Pi {
		sway.MovementPhase -= 2.0 * math.Pi
	}

	// Calculate bob offset
	amplitude := s.config.MovementBobAmplitude * sway.WeaponWeight
	if sway.IsSprinting {
		amplitude *= 1.6
	}
	if sway.IsAiming {
		amplitude *= 0.3 // Reduced bob when aiming
	}

	// Vertical bob (main up/down motion)
	verticalBob := math.Sin(sway.MovementPhase) * amplitude

	// Horizontal bob (slight side-to-side, half frequency of vertical)
	horizontalBob := math.Sin(sway.MovementPhase*0.5) * amplitude * 0.4

	// Add to target (spring physics will interpolate)
	sway.TargetY += verticalBob * dt * 30.0
	sway.TargetX += horizontalBob * dt * 30.0
}

// applyBreathSway adds subtle idle movement simulating breathing.
func (s *System) applyBreathSway(sway *Component, dt float64) {
	// Always advance breath phase (even when moving)
	sway.BreathPhase += dt * s.config.BreathFrequency * 2.0 * math.Pi

	// Normalize phase
	if sway.BreathPhase > 2.0*math.Pi {
		sway.BreathPhase -= 2.0 * math.Pi
	}

	// Breath only applies when not moving
	if sway.IsMoving {
		return
	}

	amplitude := s.config.BreathAmplitude
	if sway.IsAiming {
		amplitude *= 0.4 // Steadier when aiming
	}

	// Breath is primarily vertical with slight horizontal
	breathY := math.Sin(sway.BreathPhase) * amplitude
	breathX := math.Sin(sway.BreathPhase*0.7) * amplitude * 0.3

	// Gentle influence on target
	sway.TargetY += breathY * dt * 10.0
	sway.TargetX += breathX * dt * 10.0
}

// applySpringPhysics simulates spring-damper return to center.
func (s *System) applySpringPhysics(sway *Component, dt float64) {
	// Spring force pulls toward target (which decays toward 0)
	stiffness := s.config.SpringStiffness
	if sway.IsAiming {
		stiffness *= 1.8 // Faster recovery when aiming
	}

	// Calculate spring acceleration toward target
	accelX := (sway.TargetX - sway.OffsetX) * stiffness
	accelY := (sway.TargetY - sway.OffsetY) * stiffness

	// Apply acceleration to velocity
	sway.VelocityX += accelX * dt
	sway.VelocityY += accelY * dt

	// Apply damping
	damping := s.config.Damping
	sway.VelocityX *= math.Pow(damping, dt*60.0)
	sway.VelocityY *= math.Pow(damping, dt*60.0)

	// Integrate velocity into position
	sway.OffsetX += sway.VelocityX * dt * 60.0
	sway.OffsetY += sway.VelocityY * dt * 60.0

	// Decay target toward center (recovery)
	recoveryRate := sway.RecoverySpeed * dt
	if sway.IsAiming {
		recoveryRate *= 2.0
	}
	sway.TargetX *= math.Max(0, 1.0-recoveryRate)
	sway.TargetY *= math.Max(0, 1.0-recoveryRate)
}

// clampOffset ensures sway doesn't exceed maximum bounds.
func (s *System) clampOffset(sway *Component) {
	maxOffset := s.config.MaxSwayOffset
	if sway.IsAiming {
		maxOffset *= 0.4 // Tighter bounds when aiming
	}

	// Clamp position
	if sway.OffsetX > maxOffset {
		sway.OffsetX = maxOffset
	} else if sway.OffsetX < -maxOffset {
		sway.OffsetX = -maxOffset
	}

	if sway.OffsetY > maxOffset {
		sway.OffsetY = maxOffset
	} else if sway.OffsetY < -maxOffset {
		sway.OffsetY = -maxOffset
	}

	// Also clamp target
	if sway.TargetX > maxOffset {
		sway.TargetX = maxOffset
	} else if sway.TargetX < -maxOffset {
		sway.TargetX = -maxOffset
	}
	if sway.TargetY > maxOffset {
		sway.TargetY = maxOffset
	} else if sway.TargetY < -maxOffset {
		sway.TargetY = -maxOffset
	}
}

// AddTurnImpulse applies sway from camera rotation.
// deltaYaw and deltaPitch are the camera rotation deltas in radians.
// Call this from input handling when the camera rotates.
func (s *System) AddTurnImpulse(sway *Component, deltaYaw, deltaPitch float64) {
	if sway == nil || !sway.Enabled {
		return
	}

	// Convert to screen-space impulse
	// Yaw rotation creates horizontal sway (opposite direction = inertia)
	// Pitch rotation creates vertical sway (opposite direction = inertia)
	multiplier := s.config.TurnSwayMultiplier * sway.WeaponWeight

	if sway.IsAiming {
		multiplier *= 0.4 // Reduced sway when aiming
	}

	// Weapon lags behind camera movement (moves opposite initially)
	impulseX := -deltaYaw * multiplier * 100.0
	impulseY := -deltaPitch * multiplier * 80.0

	// Add to velocity for momentum
	sway.VelocityX += impulseX
	sway.VelocityY += impulseY

	// Add to target for spring physics
	sway.TargetX += impulseX * 0.3
	sway.TargetY += impulseY * 0.3
}

// SetMovementState updates the movement flags for appropriate sway behavior.
func (s *System) SetMovementState(sway *Component, isMoving, isSprinting bool) {
	if sway == nil {
		return
	}
	sway.IsMoving = isMoving
	sway.IsSprinting = isSprinting
}

// SetAimingState updates the aiming flag for reduced sway during ADS.
func (s *System) SetAimingState(sway *Component, isAiming bool) {
	if sway == nil {
		return
	}
	sway.IsAiming = isAiming
}

// GetConfig returns the current sway configuration (for debugging/tuning).
func (s *System) GetConfig() SwayConfig {
	return s.config
}

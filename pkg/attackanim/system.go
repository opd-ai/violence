// Package attackanim provides visual attack animation system.
package attackanim

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages attack animation state transitions and visual parameters.
type System struct {
	genreID string
	logger  *logrus.Entry
}

// NewSystem creates a new attack animation system.
func NewSystem(genreID string) *System {
	return &System{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "attackanim",
			"genre":  genreID,
		}),
	}
}

// Update advances attack animations and calculates visual parameters.
func (s *System) Update(w *engine.World) {
	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	deltaTime := 1.0 / 60.0

	for _, entity := range entities {
		comp := s.getComponent(w, entity)
		if comp == nil {
			continue
		}

		if comp.AttackState == StateIdle {
			continue
		}

		comp.StateTime += deltaTime

		// Transition states based on timing
		switch comp.AttackState {
		case StateWindup:
			if comp.StateTime >= comp.WindupTime {
				s.transitionToStrike(comp)
			} else {
				s.updateWindupVisuals(comp)
			}

		case StateStrike:
			if comp.StateTime >= comp.StrikeTime {
				s.transitionToRecovery(comp)
			} else {
				s.updateStrikeVisuals(comp)
			}

		case StateRecovery:
			if comp.StateTime >= comp.RecoveryTime {
				s.transitionToIdle(comp)
			} else {
				s.updateRecoveryVisuals(comp)
			}
		}
	}
}

// StartAttack initiates an attack animation.
func (s *System) StartAttack(w *engine.World, entity engine.Entity, animationType string, targetDirX, targetDirY, intensity float64) {
	comp := s.getOrCreateComponent(w, entity)

	comp.AnimationType = animationType
	comp.AttackState = StateWindup
	comp.StateTime = 0
	comp.TargetDirX = targetDirX
	comp.TargetDirY = targetDirY
	comp.Intensity = math.Max(0.5, math.Min(1.5, intensity))

	// Set timings based on attack type
	switch animationType {
	case "melee_slash":
		comp.WindupTime = 0.2
		comp.StrikeTime = 0.15
		comp.RecoveryTime = 0.25

	case "overhead_smash":
		comp.WindupTime = 0.4
		comp.StrikeTime = 0.1
		comp.RecoveryTime = 0.3

	case "lunge":
		comp.WindupTime = 0.25
		comp.StrikeTime = 0.2
		comp.RecoveryTime = 0.2

	case "ranged_charge":
		comp.WindupTime = 0.3
		comp.StrikeTime = 0.05
		comp.RecoveryTime = 0.15

	case "spin_attack":
		comp.WindupTime = 0.15
		comp.StrikeTime = 0.3
		comp.RecoveryTime = 0.25

	case "quick_jab":
		comp.WindupTime = 0.1
		comp.StrikeTime = 0.08
		comp.RecoveryTime = 0.15

	default:
		comp.WindupTime = 0.2
		comp.StrikeTime = 0.15
		comp.RecoveryTime = 0.2
	}

	s.logger.Debugf("started attack animation %s for entity %d", animationType, entity)
}

// StopAttack immediately ends the current attack animation.
func (s *System) StopAttack(w *engine.World, entity engine.Entity) {
	comp := s.getComponent(w, entity)
	if comp != nil {
		s.transitionToIdle(comp)
	}
}

func (s *System) transitionToStrike(comp *Component) {
	comp.AttackState = StateStrike
	comp.StateTime = 0
}

func (s *System) transitionToRecovery(comp *Component) {
	comp.AttackState = StateRecovery
	comp.StateTime = 0
}

func (s *System) transitionToIdle(comp *Component) {
	comp.AttackState = StateIdle
	comp.StateTime = 0
	comp.RotationAngle = 0
	comp.OffsetX = 0
	comp.OffsetY = 0
	comp.SquashStretch = 1.0
}

func (s *System) updateWindupVisuals(comp *Component) {
	progress := comp.AnimationProgress()
	easeProgress := s.easeInOut(progress)

	switch comp.AnimationType {
	case "melee_slash":
		// Pull back weapon at angle
		comp.RotationAngle = -45 * easeProgress * comp.Intensity
		comp.OffsetX = -comp.TargetDirX * 3 * easeProgress
		comp.OffsetY = -comp.TargetDirY * 3 * easeProgress
		comp.SquashStretch = 1.0 - 0.05*easeProgress

	case "overhead_smash":
		// Raise weapon overhead
		comp.RotationAngle = -90 * easeProgress * comp.Intensity
		comp.OffsetY = -5 * easeProgress
		comp.SquashStretch = 1.0 + 0.1*easeProgress

	case "lunge":
		// Crouch and pull back
		comp.OffsetX = -comp.TargetDirX * 5 * easeProgress
		comp.OffsetY = 2 * easeProgress
		comp.SquashStretch = 1.0 - 0.15*easeProgress
		comp.RotationAngle = -30 * easeProgress

	case "ranged_charge":
		// Draw back ranged weapon
		comp.OffsetX = -comp.TargetDirX * 4 * easeProgress
		comp.OffsetY = -comp.TargetDirY * 4 * easeProgress
		comp.SquashStretch = 1.0 - 0.08*easeProgress

	case "spin_attack":
		// Begin rotation
		comp.RotationAngle = -60 * easeProgress
		comp.OffsetX = math.Cos(comp.RotationAngle*math.Pi/180) * 2 * easeProgress
		comp.OffsetY = math.Sin(comp.RotationAngle*math.Pi/180) * 2 * easeProgress

	case "quick_jab":
		// Minimal windup
		comp.OffsetX = -comp.TargetDirX * 2 * easeProgress
		comp.OffsetY = -comp.TargetDirY * 2 * easeProgress
	}
}

func (s *System) updateStrikeVisuals(comp *Component) {
	progress := comp.AnimationProgress()
	easeProgress := s.easeOut(progress)

	switch comp.AnimationType {
	case "melee_slash":
		// Swing through
		comp.RotationAngle = -45 + 90*easeProgress*comp.Intensity
		comp.OffsetX = comp.TargetDirX * 4 * easeProgress
		comp.OffsetY = comp.TargetDirY * 4 * easeProgress
		comp.SquashStretch = 1.0 + 0.1*easeProgress

	case "overhead_smash":
		// Slam down
		comp.RotationAngle = -90 + 90*easeProgress*comp.Intensity
		comp.OffsetY = -5 + 8*easeProgress
		comp.SquashStretch = 1.1 - 0.2*easeProgress

	case "lunge":
		// Thrust forward
		comp.OffsetX = comp.TargetDirX * 8 * easeProgress
		comp.OffsetY = 2 - 2*easeProgress
		comp.SquashStretch = 0.85 + 0.25*easeProgress
		comp.RotationAngle = -30 + 30*easeProgress

	case "ranged_charge":
		// Release projectile
		comp.OffsetX = comp.TargetDirX * 2 * easeProgress
		comp.OffsetY = comp.TargetDirY * 2 * easeProgress
		comp.SquashStretch = 0.92 + 0.18*easeProgress

	case "spin_attack":
		// Full rotation
		totalRotation := -60 + 420*easeProgress
		comp.RotationAngle = totalRotation
		radius := 3.0
		comp.OffsetX = math.Cos(totalRotation*math.Pi/180) * radius
		comp.OffsetY = math.Sin(totalRotation*math.Pi/180) * radius

	case "quick_jab":
		// Fast thrust
		comp.OffsetX = comp.TargetDirX * 5 * easeProgress
		comp.OffsetY = comp.TargetDirY * 5 * easeProgress
	}
}

func (s *System) updateRecoveryVisuals(comp *Component) {
	progress := comp.AnimationProgress()
	easeProgress := s.easeInOut(progress)

	// Return to idle position from strike
	invProgress := 1.0 - easeProgress

	switch comp.AnimationType {
	case "melee_slash":
		comp.RotationAngle = 45 * invProgress
		comp.OffsetX = comp.TargetDirX * 4 * invProgress
		comp.OffsetY = comp.TargetDirY * 4 * invProgress
		comp.SquashStretch = 1.0 + 0.1*invProgress

	case "overhead_smash":
		comp.RotationAngle = 0
		comp.OffsetY = 3 * invProgress
		comp.SquashStretch = 0.9 + 0.1*easeProgress

	case "lunge":
		comp.OffsetX = comp.TargetDirX * 8 * invProgress
		comp.OffsetY = 0
		comp.SquashStretch = 1.0

	case "ranged_charge":
		comp.OffsetX = comp.TargetDirX * 2 * invProgress
		comp.OffsetY = comp.TargetDirY * 2 * invProgress
		comp.SquashStretch = 1.0

	case "spin_attack":
		comp.RotationAngle = 360 * invProgress
		comp.OffsetX = 0
		comp.OffsetY = 0

	case "quick_jab":
		comp.OffsetX = comp.TargetDirX * 5 * invProgress
		comp.OffsetY = comp.TargetDirY * 5 * invProgress
	}
}

func (s *System) easeInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

func (s *System) easeOut(t float64) float64 {
	return 1 - math.Pow(1-t, 3)
}

func (s *System) getComponent(w *engine.World, entity engine.Entity) *Component {
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if !ok {
		return nil
	}
	return comp.(*Component)
}

func (s *System) getOrCreateComponent(w *engine.World, entity engine.Entity) *Component {
	comp := s.getComponent(w, entity)
	if comp == nil {
		comp = &Component{
			AttackState:   StateIdle,
			SquashStretch: 1.0,
		}
		w.AddComponent(entity, comp)
	}
	return comp
}

// GetAnimationParams returns visual parameters for rendering.
func (s *System) GetAnimationParams(w *engine.World, entity engine.Entity) (offsetX, offsetY, rotation, squash float64, animating bool) {
	comp := s.getComponent(w, entity)
	if comp == nil || comp.AttackState == StateIdle {
		return 0, 0, 0, 1.0, false
	}
	return comp.OffsetX, comp.OffsetY, comp.RotationAngle, comp.SquashStretch, true
}

// Package motion provides organic, realistic animation through easing curves,
// squash-and-stretch, secondary motion, breathing idle, and weight-based movement.
package motion

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System applies organic motion effects to entities with velocity and position.
type System struct {
	logger *logrus.Entry
}

// NewSystem creates an organic motion system.
func NewSystem() *System {
	return &System{
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "motion",
			"package":     "motion",
		}),
	}
}

// Update applies easing, squash/stretch, secondary motion, and breathing to all entities.
func (s *System) Update(w *engine.World) {
	const deltaTime = 1.0 / 60.0

	motionType := reflect.TypeOf(&Component{})
	posType := reflect.TypeOf(&engine.Position{})
	velType := reflect.TypeOf(&engine.Velocity{})

	entities := w.Query(motionType)

	for _, ent := range entities {
		motionComp, found := w.GetComponent(ent, motionType)
		if !found {
			continue
		}
		motion := motionComp.(*Component)

		// Update breathing idle animation
		s.updateBreathing(motion, deltaTime)

		// Apply easing to velocity if entity has velocity component
		if velComp, found := w.GetComponent(ent, velType); found {
			vel := velComp.(*engine.Velocity)
			s.applyEasing(motion, vel, deltaTime)
		}

		// Update squash/stretch timing
		s.updateSquashStretch(motion, deltaTime)

		// Update secondary motion trails if entity has position
		if posComp, found := w.GetComponent(ent, posType); found {
			pos := posComp.(*engine.Position)
			s.updateSecondaryMotion(motion, pos, deltaTime)
		}
	}
}

// updateBreathing applies idle breathing animation.
func (s *System) updateBreathing(motion *Component, deltaTime float64) {
	if motion.BreathFrequency <= 0 {
		return
	}

	motion.BreathPhase += 2.0 * math.Pi * motion.BreathFrequency * deltaTime

	// Wrap phase to [0, 2π]
	if motion.BreathPhase > 2.0*math.Pi {
		motion.BreathPhase -= 2.0 * math.Pi
	}
}

// applyEasing smoothly transitions velocity using ease-in-out curves.
func (s *System) applyEasing(motion *Component, vel *engine.Velocity, deltaTime float64) {
	// Set target velocity from current velocity component
	motion.TargetVelocityX = vel.DX
	motion.TargetVelocityY = vel.DY

	// Compute acceleration factor based on mass (heavier = slower acceleration)
	massFactor := 1.0
	if motion.Mass > 0 {
		massFactor = 1.0 / math.Sqrt(motion.Mass)
	}

	easeSpeed := motion.EaseRate * massFactor * deltaTime

	// Ease-in-out using smooth lerp
	motion.EasedVelocityX = s.easeInOut(motion.EasedVelocityX, motion.TargetVelocityX, easeSpeed)
	motion.EasedVelocityY = s.easeInOut(motion.EasedVelocityY, motion.TargetVelocityY, easeSpeed)

	// Write eased velocity back to velocity component
	vel.DX = motion.EasedVelocityX
	vel.DY = motion.EasedVelocityY
}

// easeInOut applies smooth ease-in-out interpolation.
func (s *System) easeInOut(current, target, speed float64) float64 {
	diff := target - current

	if math.Abs(diff) < 0.001 {
		return target
	}

	// Smooth ease using exponential decay
	return current + diff*(1.0-math.Exp(-speed*10.0))
}

// updateSquashStretch manages squash/stretch timing and recovery.
func (s *System) updateSquashStretch(motion *Component, deltaTime float64) {
	motion.ImpactTime += deltaTime

	// Recover to normal proportions over 0.15 seconds
	recoveryRate := deltaTime / 0.15

	motion.SquashX += (1.0 - motion.SquashX) * recoveryRate
	motion.SquashY += (1.0 - motion.SquashY) * recoveryRate

	// Clamp to reasonable range
	motion.SquashX = clamp(motion.SquashX, 0.5, 1.5)
	motion.SquashY = clamp(motion.SquashY, 0.5, 1.5)
}

// updateSecondaryMotion updates trailing elements (tails, cloaks, hair).
func (s *System) updateSecondaryMotion(motion *Component, pos *engine.Position, deltaTime float64) {
	if motion.TrailLength <= 0 {
		return
	}

	// Initialize trail if not yet created
	if len(motion.TrailOffsetX) != motion.TrailLength {
		motion.TrailOffsetX = make([]float64, motion.TrailLength)
		motion.TrailOffsetY = make([]float64, motion.TrailLength)
		for i := 0; i < motion.TrailLength; i++ {
			motion.TrailOffsetX[i] = pos.X
			motion.TrailOffsetY[i] = pos.Y
		}
		motion.LastPositionX = pos.X
		motion.LastPositionY = pos.Y
		return
	}

	// Update trail segments - each follows the one ahead
	followRate := (1.0 - motion.TrailStiffness) * deltaTime * 10.0

	for i := motion.TrailLength - 1; i >= 0; i-- {
		var targetX, targetY float64
		if i == 0 {
			// First segment follows entity position
			targetX = pos.X
			targetY = pos.Y
		} else {
			// Other segments follow previous segment
			targetX = motion.TrailOffsetX[i-1]
			targetY = motion.TrailOffsetY[i-1]
		}

		// Smoothly interpolate
		motion.TrailOffsetX[i] += (targetX - motion.TrailOffsetX[i]) * followRate
		motion.TrailOffsetY[i] += (targetY - motion.TrailOffsetY[i]) * followRate
	}

	motion.LastPositionX = pos.X
	motion.LastPositionY = pos.Y
}

// TriggerImpact causes squash/stretch effect on landing or collision.
func (s *System) TriggerImpact(motion *Component, velocityMagnitude float64) {
	if motion.ImpactTime < 0.15 {
		return // Still recovering from previous impact
	}

	// Calculate squash amount based on velocity
	impactStrength := clamp(velocityMagnitude/100.0, 0.0, 1.0)

	// Squash vertically, expand horizontally (preserve volume)
	motion.SquashY = 1.0 - impactStrength*0.4
	motion.SquashX = 1.0 + impactStrength*0.3
	motion.ImpactTime = 0

	s.logger.WithFields(logrus.Fields{
		"impact_strength": impactStrength,
		"squash_y":        motion.SquashY,
		"squash_x":        motion.SquashX,
	}).Debug("triggered impact squash/stretch")
}

// GetBreathOffset returns Y-axis offset for breathing animation.
func (s *System) GetBreathOffset(motion *Component) float64 {
	if motion.BreathAmplitude <= 0 {
		return 0
	}
	return math.Sin(motion.BreathPhase) * motion.BreathAmplitude
}

// GetSquashStretch returns current scale factors for rendering.
func (s *System) GetSquashStretch(motion *Component) (scaleX, scaleY float64) {
	return motion.SquashX, motion.SquashY
}

// GetTrailSegment returns position of a trailing element segment.
func (s *System) GetTrailSegment(motion *Component, index int) (x, y float64, valid bool) {
	if index < 0 || index >= len(motion.TrailOffsetX) {
		return 0, 0, false
	}
	return motion.TrailOffsetX[index], motion.TrailOffsetY[index], true
}

// InitializeMotion sets up default motion parameters for an entity.
func InitializeMotion(mass float64, hasTrail bool) *Component {
	comp := &Component{
		SquashX:         1.0,
		SquashY:         1.0,
		EaseRate:        5.0,
		Mass:            mass,
		BreathFrequency: 0.2, // 0.2 breaths/sec = 1 breath per 5 seconds
		BreathAmplitude: 0.5, // Half pixel up/down
		Grounded:        true,
	}

	if hasTrail {
		comp.TrailLength = 5
		comp.TrailStiffness = 0.3
	}

	return comp
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

package weaponanim

import (
	"image/color"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

const (
	maxTrailPoints = 12
	trailFadeRate  = 3.0 // Points per second to age
)

// System updates weapon swing animations and manages motion trails.
type System struct {
	logger *logrus.Entry
	genre  string
}

// NewSystem creates a weapon animation system.
func NewSystem() *System {
	return &System{
		logger: logrus.WithFields(logrus.Fields{
			"system": "weaponanim",
		}),
		genre: "fantasy",
	}
}

// SetGenre configures genre-specific animation parameters.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
}

// Update implements the System interface.
func (s *System) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0

	it := w.QueryWithBitmask(engine.ComponentIDPosition)
	swingTriggerType := reflect.TypeOf(&SwingTriggerComponent{})
	weaponAnimType := reflect.TypeOf(&WeaponAnimComponent{})
	posType := reflect.TypeOf(&PositionComponent{})
	velType := reflect.TypeOf(&VelocityComponent{})

	for it.Next() {
		entity := it.Entity()

		// Check for pending swing triggers
		triggerComp, hasTrigger := w.GetComponent(entity, swingTriggerType)
		if hasTrigger {
			trigger := triggerComp.(*SwingTriggerComponent)
			if trigger.Pending {
				// Get facing direction from velocity or default
				facing := 0.0
				if velComp, hasVel := w.GetComponent(entity, velType); hasVel {
					vel := velComp.(*VelocityComponent)
					if vel.VX != 0 || vel.VY != 0 {
						facing = math.Atan2(vel.VY, vel.VX)
					}
				}

				// Start the swing animation
				startAngle, endAngle, duration := GetSwingParameters(SwingType(trigger.SwingType), facing)

				animComp, hasAnim := w.GetComponent(entity, weaponAnimType)
				if !hasAnim {
					animComp = &WeaponAnimComponent{
						Color: color.RGBA{R: 200, G: 200, B: 220, A: 200},
						Width: 3.0,
					}
					w.AddComponent(entity, animComp)
				}

				anim := animComp.(*WeaponAnimComponent)
				anim.Active = true
				anim.SwingType = SwingType(trigger.SwingType)
				anim.Progress = 0.0
				anim.Duration = duration
				anim.StartAngle = startAngle
				anim.EndAngle = endAngle
				anim.ArcRadius = 25.0 // Default weapon length
				anim.TrailPoints = nil

				trigger.Pending = false
			}
		}

		// Update existing animations
		animComp, ok := w.GetComponent(entity, weaponAnimType)
		if !ok {
			continue
		}

		anim := animComp.(*WeaponAnimComponent)
		if !anim.Active {
			continue
		}

		// Get entity position
		posComp, hasPos := w.GetComponent(entity, posType)
		if !hasPos {
			continue
		}
		pos := posComp.(*PositionComponent)

		s.updateAnimation(anim, pos, deltaTime)
		s.updateTrail(anim, pos, deltaTime)
	}
}

// updateAnimation progresses the swing animation.
func (s *System) updateAnimation(anim *WeaponAnimComponent, pos *PositionComponent, deltaTime float64) {
	if !anim.Active {
		return
	}

	anim.Progress += deltaTime / anim.Duration
	if anim.Progress >= 1.0 {
		anim.Progress = 1.0
		anim.Active = false
		anim.TrailPoints = nil
	}
}

// updateTrail maintains the motion trail behind the weapon.
func (s *System) updateTrail(anim *WeaponAnimComponent, pos *PositionComponent, deltaTime float64) {
	// Add new trail point at current tip position
	if anim.Active {
		tipX, tipY := anim.GetTipPosition(pos.X, pos.Y)
		newPoint := TrailPoint{
			X:        tipX,
			Y:        tipY,
			Age:      0.0,
			Rotation: anim.GetCurrentAngle(),
		}

		anim.TrailPoints = append(anim.TrailPoints, newPoint)
		if len(anim.TrailPoints) > maxTrailPoints {
			anim.TrailPoints = anim.TrailPoints[1:]
		}
	}

	// Age existing trail points
	surviving := make([]TrailPoint, 0, len(anim.TrailPoints))
	for i := range anim.TrailPoints {
		anim.TrailPoints[i].Age += deltaTime * trailFadeRate
		if anim.TrailPoints[i].Age < 1.0 {
			surviving = append(surviving, anim.TrailPoints[i])
		}
	}
	anim.TrailPoints = surviving
}

// PositionComponent is a minimal position component interface.
type PositionComponent struct {
	X, Y float64
}

func (p *PositionComponent) Type() string {
	return "position"
}

// VelocityComponent is a minimal velocity component interface.
type VelocityComponent struct {
	VX, VY float64
}

func (v *VelocityComponent) Type() string {
	return "velocity"
}

// SwingTriggerComponent signals pending weapon swings (defined in combat package).
type SwingTriggerComponent struct {
	SwingType int
	Pending   bool
}

func (s *SwingTriggerComponent) Type() string {
	return "swing_trigger"
}

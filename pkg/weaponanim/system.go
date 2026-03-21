package weaponanim

import (
	"image/color"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/common"
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
	deltaTime := common.DeltaTime

	swingTriggerType := reflect.TypeOf(&SwingTriggerComponent{})
	weaponAnimType := reflect.TypeOf(&WeaponAnimComponent{})
	posType := reflect.TypeOf(&PositionComponent{})
	velType := reflect.TypeOf(&VelocityComponent{})

	// Union of entities with active animations or pending swing triggers
	entitySet := make(map[engine.Entity]struct{})
	for _, e := range w.Query(weaponAnimType) {
		entitySet[e] = struct{}{}
	}
	for _, e := range w.Query(swingTriggerType) {
		entitySet[e] = struct{}{}
	}

	for entity := range entitySet {
		s.processPendingSwingTrigger(w, entity, swingTriggerType, weaponAnimType, velType)
		s.processActiveAnimation(w, entity, weaponAnimType, posType, deltaTime)
	}
}

// processPendingSwingTrigger checks for and initiates pending weapon swing animations.
func (s *System) processPendingSwingTrigger(w *engine.World, entity engine.Entity, swingTriggerType, weaponAnimType, velType reflect.Type) {
	triggerComp, hasTrigger := w.GetComponent(entity, swingTriggerType)
	if !hasTrigger {
		return
	}

	trigger := triggerComp.(*SwingTriggerComponent)
	if !trigger.Pending {
		return
	}

	facing := s.calculateFacingDirection(w, entity, velType)
	s.startSwingAnimation(w, entity, weaponAnimType, trigger, facing)
	trigger.Pending = false
}

// calculateFacingDirection determines entity facing from velocity or returns default.
func (s *System) calculateFacingDirection(w *engine.World, entity engine.Entity, velType reflect.Type) float64 {
	velComp, hasVel := w.GetComponent(entity, velType)
	if !hasVel {
		return 0.0
	}

	vel := velComp.(*VelocityComponent)
	if vel.VX != 0 || vel.VY != 0 {
		return math.Atan2(vel.VY, vel.VX)
	}
	return 0.0
}

// startSwingAnimation initializes a new weapon swing animation.
func (s *System) startSwingAnimation(w *engine.World, entity engine.Entity, weaponAnimType reflect.Type, trigger *SwingTriggerComponent, facing float64) {
	startAngle, endAngle, duration := GetSwingParameters(SwingType(trigger.SwingType), facing)

	anim := s.getOrCreateAnimComponent(w, entity, weaponAnimType)
	anim.Active = true
	anim.SwingType = SwingType(trigger.SwingType)
	anim.Progress = 0.0
	anim.Duration = duration
	anim.StartAngle = startAngle
	anim.EndAngle = endAngle
	anim.ArcRadius = 25.0
	anim.TrailPoints = nil
}

// getOrCreateAnimComponent retrieves existing or creates new weapon animation component.
func (s *System) getOrCreateAnimComponent(w *engine.World, entity engine.Entity, weaponAnimType reflect.Type) *WeaponAnimComponent {
	animComp, hasAnim := w.GetComponent(entity, weaponAnimType)
	if hasAnim {
		return animComp.(*WeaponAnimComponent)
	}

	newAnim := &WeaponAnimComponent{
		Color: color.RGBA{R: 200, G: 200, B: 220, A: 200},
		Width: 3.0,
	}
	w.AddComponent(entity, newAnim)
	return newAnim
}

// processActiveAnimation updates ongoing weapon swing animations.
func (s *System) processActiveAnimation(w *engine.World, entity engine.Entity, weaponAnimType, posType reflect.Type, deltaTime float64) {
	animComp, ok := w.GetComponent(entity, weaponAnimType)
	if !ok {
		return
	}

	anim := animComp.(*WeaponAnimComponent)

	posComp, hasPos := w.GetComponent(entity, posType)
	if !hasPos {
		return
	}
	pos := posComp.(*PositionComponent)

	if anim.Active {
		s.updateAnimation(anim, pos, deltaTime)
	}
	// Age trail points even when animation is inactive so they fade after completion.
	if anim.Active || len(anim.TrailPoints) > 0 {
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

// Type returns the component type identifier for ECS registration.
func (p *PositionComponent) Type() string {
	return "position"
}

// VelocityComponent is a minimal velocity component interface.
type VelocityComponent struct {
	VX, VY float64
}

// Type returns the component type identifier for ECS registration.
func (v *VelocityComponent) Type() string {
	return "velocity"
}

// SwingTriggerComponent signals pending weapon swings (defined in combat package).
type SwingTriggerComponent struct {
	SwingType int
	Pending   bool
}

// Type returns the component type identifier for ECS registration.
func (s *SwingTriggerComponent) Type() string {
	return "swing_trigger"
}

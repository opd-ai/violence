// Package combat - Positional advantage system
package combat

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// PositionalSystem manages entity facing and positional combat calculations.
type PositionalSystem struct {
	config   PositionalConfig
	genreID  string
	gameTime float64
}

// NewPositionalSystem creates the positional advantage system.
func NewPositionalSystem(genreID string) *PositionalSystem {
	return &PositionalSystem{
		config:  GetPositionalConfig(genreID),
		genreID: genreID,
	}
}

// Update runs positional logic each frame.
func (s *PositionalSystem) Update(w *engine.World) {
	deltaTime := 0.016 // Assume 60 FPS
	s.gameTime += deltaTime

	s.updateFacingFromVelocity(w, deltaTime)
}

// updateFacingFromVelocity updates entity facing based on movement direction.
func (s *PositionalSystem) updateFacingFromVelocity(w *engine.World, deltaTime float64) {
	// Query entities with both positional component and velocity
	entities := w.Query(
		reflect.TypeOf(&PositionalComponent{}),
		reflect.TypeOf(&engine.Velocity{}),
	)

	for _, e := range entities {
		posComp, _ := w.GetComponent(e, reflect.TypeOf(&PositionalComponent{}))
		velComp, _ := w.GetComponent(e, reflect.TypeOf(&engine.Velocity{}))

		pc := posComp.(*PositionalComponent)
		vel := velComp.(*engine.Velocity)

		// Update facing if moving
		speed := math.Sqrt(vel.DX*vel.DX + vel.DY*vel.DY)
		if speed > 0.1 {
			pc.SetFacingFromDirection(vel.DX, vel.DY)
			pc.LastUpdate = s.gameTime
		}
	}
}

// CalculateDamageMultiplier computes positional damage modifier for an attack.
func (s *PositionalSystem) CalculateDamageMultiplier(
	w *engine.World,
	attacker engine.Entity,
	target engine.Entity,
) float64 {
	// Get positions
	attackerPosComp, ok1 := w.GetComponent(attacker, reflect.TypeOf(&engine.Position{}))
	targetPosComp, ok2 := w.GetComponent(target, reflect.TypeOf(&engine.Position{}))
	if !ok1 || !ok2 {
		return 1.0
	}

	attackerPos := attackerPosComp.(*engine.Position)
	targetPos := targetPosComp.(*engine.Position)

	// Get positional components (facing data)
	var attackerPosData *PositionalComponent
	var targetPosData *PositionalComponent

	if comp, ok := w.GetComponent(attacker, reflect.TypeOf(&PositionalComponent{})); ok {
		attackerPosData = comp.(*PositionalComponent)
	}
	if comp, ok := w.GetComponent(target, reflect.TypeOf(&PositionalComponent{})); ok {
		targetPosData = comp.(*PositionalComponent)
	}

	// Calculate advantage
	advantage, multiplier := CalculatePositionalAdvantage(
		attackerPos.X, attackerPos.Y,
		targetPos.X, targetPos.Y,
		attackerPosData, targetPosData,
		s.config,
	)

	// Log significant advantages
	if advantage != AdvantageFrontal {
		logrus.WithFields(logrus.Fields{
			"system_name": "PositionalSystem",
			"attacker":    attacker,
			"target":      target,
			"advantage":   advantageString(advantage),
			"multiplier":  multiplier,
		}).Debug("Positional advantage applied")
	}

	return multiplier
}

// GetAdvantageForAttack returns both advantage type and multiplier (for UI feedback).
func (s *PositionalSystem) GetAdvantageForAttack(
	w *engine.World,
	attacker engine.Entity,
	target engine.Entity,
) (PositionalAdvantage, float64) {
	attackerPosComp, ok1 := w.GetComponent(attacker, reflect.TypeOf(&engine.Position{}))
	targetPosComp, ok2 := w.GetComponent(target, reflect.TypeOf(&engine.Position{}))
	if !ok1 || !ok2 {
		return AdvantageFrontal, 1.0
	}

	attackerPos := attackerPosComp.(*engine.Position)
	targetPos := targetPosComp.(*engine.Position)

	var attackerPosData *PositionalComponent
	var targetPosData *PositionalComponent

	if comp, ok := w.GetComponent(attacker, reflect.TypeOf(&PositionalComponent{})); ok {
		attackerPosData = comp.(*PositionalComponent)
	}
	if comp, ok := w.GetComponent(target, reflect.TypeOf(&PositionalComponent{})); ok {
		targetPosData = comp.(*PositionalComponent)
	}

	return CalculatePositionalAdvantage(
		attackerPos.X, attackerPos.Y,
		targetPos.X, targetPos.Y,
		attackerPosData, targetPosData,
		s.config,
	)
}

// SetGenre updates the config for a new genre.
func (s *PositionalSystem) SetGenre(genreID string) {
	s.config = GetPositionalConfig(genreID)
	s.genreID = genreID
}

// AddPositionalComponent adds facing/height data to an entity.
func (s *PositionalSystem) AddPositionalComponent(
	w *engine.World,
	entity engine.Entity,
	facingAngle float64,
	height float64,
) {
	comp := &PositionalComponent{
		FacingAngle: facingAngle,
		Height:      height,
		LastUpdate:  s.gameTime,
	}
	w.AddComponent(entity, comp)
}

// advantageString returns human-readable advantage name.
func advantageString(adv PositionalAdvantage) string {
	switch adv {
	case AdvantageFrontal:
		return "frontal"
	case AdvantageFlank:
		return "flank"
	case AdvantageBackstab:
		return "backstab"
	case AdvantageElevation:
		return "elevation"
	default:
		return "unknown"
	}
}

// IsBackstabAngle checks if angle qualifies as backstab without full calculation.
func (s *PositionalSystem) IsBackstabAngle(
	attackerX, attackerY float64,
	targetX, targetY float64,
	targetFacing float64,
) bool {
	dx := attackerX - targetX
	dy := attackerY - targetY
	attackAngle := math.Atan2(dy, dx)
	angleDiff := normalizeAngle(attackAngle - targetFacing)
	return math.Abs(angleDiff-math.Pi) < s.config.BackstabAngle
}

// IsFlankAngle checks if angle qualifies as flank.
func (s *PositionalSystem) IsFlankAngle(
	attackerX, attackerY float64,
	targetX, targetY float64,
	targetFacing float64,
) bool {
	dx := attackerX - targetX
	dy := attackerY - targetY
	attackAngle := math.Atan2(dy, dx)
	angleDiff := normalizeAngle(attackAngle - targetFacing)
	return math.Abs(angleDiff-math.Pi/2) < s.config.FlankAngle ||
		math.Abs(angleDiff+math.Pi/2) < s.config.FlankAngle
}

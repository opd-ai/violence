// Package combat - Telegraph system for enemy attack warnings
package combat

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// TelegraphSystem manages enemy attack telegraphs and execution.
type TelegraphSystem struct {
	genreID  string
	rng      *rng.RNG
	patterns []AttackPattern
	logger   *logrus.Entry
}

// NewTelegraphSystem creates a telegraph system.
func NewTelegraphSystem(genreID string, seed int64) *TelegraphSystem {
	r := rng.NewRNG(uint64(seed))
	return &TelegraphSystem{
		genreID:  genreID,
		rng:      r,
		patterns: DefaultPatterns(genreID, r),
		logger: logrus.WithFields(logrus.Fields{
			"system": "telegraph",
			"genre":  genreID,
		}),
	}
}

// Update processes all telegraph components.
func (s *TelegraphSystem) Update(w *engine.World, deltaTime float64) {
	// Query all entities with telegraph components
	it := w.QueryWithBitmask(engine.ComponentIDPosition, engine.ComponentIDHealth)

	posType := reflect.TypeOf(&engine.Position{})
	telegraphType := reflect.TypeOf(&TelegraphComponent{})
	healthType := reflect.TypeOf(&engine.Health{})

	for it.Next() {
		e := it.Entity()

		tComp, hasTelegraph := w.GetComponent(e, telegraphType)
		if !hasTelegraph {
			continue
		}

		telegraph := tComp.(*TelegraphComponent)
		pos, _ := w.GetComponent(e, posType)
		position := pos.(*engine.Position)

		health, _ := w.GetComponent(e, healthType)
		hp := health.(*engine.Health)

		// Dead entities don't attack
		if hp.Current <= 0 {
			telegraph.Phase = PhaseInactive
			continue
		}

		s.updateTelegraph(w, e, telegraph, position, deltaTime)
	}
}

func (s *TelegraphSystem) updateTelegraph(w *engine.World, entity engine.Entity, telegraph *TelegraphComponent, pos *engine.Position, deltaTime float64) {
	switch telegraph.Phase {
	case PhaseInactive:
		// System doesn't auto-trigger attacks - AI should initiate
		return

	case PhaseWindup:
		telegraph.PhaseTimer -= deltaTime
		if telegraph.PhaseTimer <= 0 {
			// Transition to active - deal damage
			telegraph.Phase = PhaseActive
			telegraph.PhaseTimer = telegraph.Pattern.ActiveTime
			telegraph.HasHit = false
			s.logger.WithFields(logrus.Fields{
				"entity":  entity,
				"pattern": telegraph.Pattern.Name,
			}).Debug("attack active")
		}

	case PhaseActive:
		if !telegraph.HasHit {
			// Check for targets in attack area and deal damage
			s.executeDamage(w, entity, telegraph, pos)
			telegraph.HasHit = true
		}

		telegraph.PhaseTimer -= deltaTime
		if telegraph.PhaseTimer <= 0 {
			telegraph.Phase = PhaseCooldown
			telegraph.PhaseTimer = telegraph.Pattern.CooldownTime
		}

	case PhaseCooldown:
		telegraph.PhaseTimer -= deltaTime
		if telegraph.PhaseTimer <= 0 {
			telegraph.Phase = PhaseInactive
			telegraph.PhaseTimer = 0
		}
	}
}

func (s *TelegraphSystem) executeDamage(w *engine.World, attacker engine.Entity, telegraph *TelegraphComponent, attackerPos *engine.Position) {
	// Find all potential targets
	it := w.QueryWithBitmask(engine.ComponentIDPosition, engine.ComponentIDHealth)

	posType := reflect.TypeOf(&engine.Position{})
	healthType := reflect.TypeOf(&engine.Health{})

	hitCount := 0

	for it.Next() {
		target := it.Entity()

		if target == attacker {
			continue // Don't hit self
		}

		tPosComp, _ := w.GetComponent(target, posType)
		targetPos := tPosComp.(*engine.Position)

		// Check if target is in attack area
		if !s.isInAttackArea(attackerPos, targetPos, telegraph) {
			continue
		}

		// Apply damage
		healthComp, hasHealth := w.GetComponent(target, healthType)
		if !hasHealth {
			continue
		}

		health := healthComp.(*engine.Health)

		// Calculate damage with direction
		dx := targetPos.X - attackerPos.X
		dy := targetPos.Y - attackerPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > 0.001 {
			_ = dx / dist // dirX for future knockback
			_ = dy / dist // dirY for future knockback
		}

		// Apply damage (simplified - in full system would use combat.System)
		health.Current -= int(telegraph.Pattern.Damage)
		if health.Current < 0 {
			health.Current = 0
		}

		hitCount++

		s.logger.WithFields(logrus.Fields{
			"attacker": attacker,
			"target":   target,
			"damage":   telegraph.Pattern.Damage,
			"pattern":  telegraph.Pattern.Name,
		}).Debug("telegraph hit")

		// TODO: Apply knockback based on telegraph.Pattern.KnockbackMul
		// TODO: Spawn hit particles
		// TODO: Play hit sound
	}

	if hitCount > 0 {
		s.logger.WithFields(logrus.Fields{
			"attacker": attacker,
			"hits":     hitCount,
			"pattern":  telegraph.Pattern.Name,
		}).Info("telegraph executed")
	}
}

func (s *TelegraphSystem) isInAttackArea(attackerPos, targetPos *engine.Position, telegraph *TelegraphComponent) bool {
	dx := targetPos.X - attackerPos.X
	dy := targetPos.Y - attackerPos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	switch telegraph.Pattern.Shape {
	case ShapeCone:
		// Check distance and angle
		if dist > telegraph.Pattern.Range {
			return false
		}
		// Angle to target
		angleToTarget := math.Atan2(dy, dx)
		// Angle of attack
		attackAngle := math.Atan2(telegraph.DirectionY, telegraph.DirectionX)
		// Angular difference
		angleDiff := math.Abs(normalizeAngle(angleToTarget - attackAngle))
		return angleDiff <= telegraph.Pattern.Angle/2

	case ShapeCircle:
		return dist <= telegraph.Pattern.Range

	case ShapeLine:
		// Distance along attack direction
		dotProduct := dx*telegraph.DirectionX + dy*telegraph.DirectionY
		if dotProduct < 0 || dotProduct > telegraph.Pattern.Range {
			return false
		}
		// Perpendicular distance from line
		perpDist := math.Abs(dx*telegraph.DirectionY - dy*telegraph.DirectionX)
		return perpDist <= telegraph.Pattern.Width/2

	case ShapeRing:
		// Between inner and outer radius
		innerRadius := telegraph.Pattern.Range - telegraph.Pattern.Width
		return dist >= innerRadius && dist <= telegraph.Pattern.Range
	}

	return false
}

func normalizeAngle(angle float64) float64 {
	for angle > math.Pi {
		angle -= 2 * math.Pi
	}
	for angle < -math.Pi {
		angle += 2 * math.Pi
	}
	return angle
}

// InitiateAttack starts an attack telegraph for an entity.
func (s *TelegraphSystem) InitiateAttack(w *engine.World, entity engine.Entity, targetX, targetY float64) bool {
	telegraphType := reflect.TypeOf(&TelegraphComponent{})
	posType := reflect.TypeOf(&engine.Position{})

	tComp, hasTelegraph := w.GetComponent(entity, telegraphType)
	if !hasTelegraph {
		return false
	}

	telegraph := tComp.(*TelegraphComponent)

	// Can't attack if not inactive
	if telegraph.Phase != PhaseInactive {
		return false
	}

	posComp, _ := w.GetComponent(entity, posType)
	pos := posComp.(*engine.Position)

	// Calculate direction to target
	dx := targetX - pos.X
	dy := targetY - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 0.001 {
		return false
	}

	// Select appropriate pattern based on distance
	pattern := SelectPattern(s.patterns, dist, s.rng)

	telegraph.Pattern = pattern
	telegraph.TargetX = targetX
	telegraph.TargetY = targetY
	telegraph.DirectionX = dx / dist
	telegraph.DirectionY = dy / dist
	telegraph.Phase = PhaseWindup
	telegraph.PhaseTimer = pattern.WindupTime
	telegraph.HasHit = false

	s.logger.WithFields(logrus.Fields{
		"entity":   entity,
		"pattern":  pattern.Name,
		"distance": dist,
	}).Debug("initiated attack")

	return true
}

// CanAttack checks if an entity can initiate an attack.
func (s *TelegraphSystem) CanAttack(w *engine.World, entity engine.Entity) bool {
	telegraphType := reflect.TypeOf(&TelegraphComponent{})
	tComp, hasTelegraph := w.GetComponent(entity, telegraphType)
	if !hasTelegraph {
		return false
	}

	telegraph := tComp.(*TelegraphComponent)
	return telegraph.Phase == PhaseInactive
}

// GetAttackProgress returns 0-1 progress through current attack, or -1 if inactive.
func (s *TelegraphSystem) GetAttackProgress(w *engine.World, entity engine.Entity) float64 {
	telegraphType := reflect.TypeOf(&TelegraphComponent{})
	tComp, hasTelegraph := w.GetComponent(entity, telegraphType)
	if !hasTelegraph {
		return -1
	}

	telegraph := tComp.(*TelegraphComponent)

	switch telegraph.Phase {
	case PhaseInactive:
		return -1
	case PhaseWindup:
		return 1.0 - (telegraph.PhaseTimer / telegraph.Pattern.WindupTime)
	case PhaseActive:
		return 1.0
	case PhaseCooldown:
		return 1.0 - (telegraph.PhaseTimer / telegraph.Pattern.CooldownTime)
	}

	return -1
}

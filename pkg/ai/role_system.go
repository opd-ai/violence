// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"fmt"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// EnemyRoleComponent marks an entity with a behavioral role.
type EnemyRoleComponent struct {
	Role   EnemyRole
	Config RoleConfig
}

// Type returns the component type identifier.
func (c *EnemyRoleComponent) Type() string {
	return "EnemyRoleComponent"
}

// SquadTacticsComponent tracks squad coordination state.
type SquadTacticsComponent struct {
	SquadID       string
	IsLeader      bool
	FlankSide     float64 // -1.0 or 1.0 for left/right flank
	LastAlertTime float64
}

// Type returns the component type identifier.
func (c *SquadTacticsComponent) Type() string {
	return "SquadTacticsComponent"
}

// PositionComponent stores entity position.
type PositionComponent struct {
	X, Y float64
}

// Type returns the component type identifier.
func (c *PositionComponent) Type() string {
	return "PositionComponent"
}

// HealthComponent stores entity health.
type HealthComponent struct {
	Current, Max float64
}

// Type returns the component type identifier.
func (c *HealthComponent) Type() string {
	return "HealthComponent"
}

// TargetComponent stores current target entity.
type TargetComponent struct {
	TargetID engine.Entity
	LastSeen float64
}

// Type returns the component type identifier.
func (c *TargetComponent) Type() string {
	return "TargetComponent"
}

// RoleBasedAISystem implements role-specific behaviors and squad tactics.
type RoleBasedAISystem struct {
	squads      map[string]*SquadTactics
	gameTime    float64
	updateTimer float64
}

// NewRoleBasedAISystem creates the AI system.
func NewRoleBasedAISystem() *RoleBasedAISystem {
	return &RoleBasedAISystem{
		squads: make(map[string]*SquadTactics),
	}
}

// Update runs role-based AI logic.
func (s *RoleBasedAISystem) Update(w *engine.World) {
	deltaTime := 0.016 // Assume 60 FPS
	s.gameTime += deltaTime
	s.updateTimer += deltaTime

	// Update squads every 0.2 seconds
	if s.updateTimer >= 0.2 {
		s.updateSquads(w)
		s.updateTimer = 0.0
	}

	// Update individual behaviors every frame
	s.updateRoleBehaviors(w, deltaTime)
}

func (s *RoleBasedAISystem) updateSquads(w *engine.World) {
	// Rebuild squad membership from components
	squadMembers := make(map[string][]engine.Entity)
	squadPositions := make(map[string]map[string][2]float64)

	entities := w.Query(
		reflect.TypeOf(&SquadTacticsComponent{}),
		reflect.TypeOf(&PositionComponent{}),
	)

	for _, e := range entities {
		squadComp, _ := w.GetComponent(e, reflect.TypeOf(&SquadTacticsComponent{}))
		posComp, _ := w.GetComponent(e, reflect.TypeOf(&PositionComponent{}))

		stc := squadComp.(*SquadTacticsComponent)
		pc := posComp.(*PositionComponent)

		squadMembers[stc.SquadID] = append(squadMembers[stc.SquadID], e)

		if squadPositions[stc.SquadID] == nil {
			squadPositions[stc.SquadID] = make(map[string][2]float64)
		}
		squadPositions[stc.SquadID][entityIDString(e)] = [2]float64{pc.X, pc.Y}
	}

	// Update each squad's tactics
	for squadID, members := range squadMembers {
		if len(members) == 0 {
			continue
		}

		squad := s.getOrCreateSquad(squadID)

		// Sync membership
		squad.Members = make([]string, len(members))
		for i, e := range members {
			squad.Members[i] = entityIDString(e)
		}

		// Find targets visible to squad
		visibleTargets := s.findVisibleTargets(w, members)

		// Update focus target if targets exist
		if len(visibleTargets) > 0 {
			squad.SelectFocusTarget(visibleTargets, nil)
			squad.RaiseAlert(0.3)
		} else {
			squad.DecayAlert(0.2)
		}

		// Update formation
		if positions, ok := squadPositions[squadID]; ok {
			targetPos := [2]float64{0, 0}
			if squad.FocusTargetID != "" {
				targetPos = s.getTargetPosition(w, squad.FocusTargetID)
			}
			squad.UpdateFormation(positions, targetPos, nil)
		}
	}
}

func (s *RoleBasedAISystem) updateRoleBehaviors(w *engine.World, deltaTime float64) {
	entities := w.Query(
		reflect.TypeOf(&EnemyRoleComponent{}),
		reflect.TypeOf(&PositionComponent{}),
		reflect.TypeOf(&HealthComponent{}),
	)

	for _, e := range entities {
		roleComp, _ := w.GetComponent(e, reflect.TypeOf(&EnemyRoleComponent{}))
		posComp, _ := w.GetComponent(e, reflect.TypeOf(&PositionComponent{}))
		healthComp, _ := w.GetComponent(e, reflect.TypeOf(&HealthComponent{}))

		rc := roleComp.(*EnemyRoleComponent)
		pc := posComp.(*PositionComponent)
		hc := healthComp.(*HealthComponent)

		// Get target if exists
		targetComp, hasTarget := w.GetComponent(e, reflect.TypeOf(&TargetComponent{}))
		var target *TargetComponent
		if hasTarget {
			target = targetComp.(*TargetComponent)
		}

		// Apply role-specific behavior
		s.applyRoleBehavior(w, e, rc, pc, hc, target, deltaTime)
	}
}

func (s *RoleBasedAISystem) applyRoleBehavior(w *engine.World, e engine.Entity, role *EnemyRoleComponent, pos *PositionComponent, health *HealthComponent, target *TargetComponent, deltaTime float64) {
	if target == nil || target.TargetID == 0 {
		return
	}

	targetPosComp, hasPos := w.GetComponent(target.TargetID, reflect.TypeOf(&PositionComponent{}))
	if !hasPos {
		return
	}
	targetPos := targetPosComp.(*PositionComponent)

	dx := targetPos.X - pos.X
	dy := targetPos.Y - pos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	healthPct := health.Current / health.Max

	// Role-specific decisions
	switch role.Role {
	case RoleTank:
		s.behaviorTank(pos, targetPos, dist, role.Config, deltaTime)
	case RoleRanged:
		s.behaviorRanged(pos, targetPos, dist, healthPct, role.Config, deltaTime)
	case RoleHealer:
		s.behaviorHealer(w, e, pos, healthPct, role.Config, deltaTime)
	case RoleAmbusher:
		s.behaviorAmbusher(pos, targetPos, dist, healthPct, role.Config, deltaTime)
	case RoleScout:
		s.behaviorScout(w, e, pos, targetPos, dist, role.Config, deltaTime)
	}
}

func (s *RoleBasedAISystem) behaviorTank(pos, targetPos *PositionComponent, dist float64, cfg RoleConfig, deltaTime float64) {
	// Aggressive advance toward target
	if dist > cfg.MinRange {
		// Move toward target
		logrus.WithFields(logrus.Fields{
			"system_name": "RoleBasedAISystem",
			"role":        "Tank",
			"action":      "advance",
			"distance":    dist,
		}).Trace("Tank advancing")
	}
}

func (s *RoleBasedAISystem) behaviorRanged(pos, targetPos *PositionComponent, dist, healthPct float64, cfg RoleConfig, deltaTime float64) {
	// Kite: maintain preferred range
	if dist < cfg.MinRange {
		// Too close, retreat
		logrus.WithFields(logrus.Fields{
			"system_name": "RoleBasedAISystem",
			"role":        "Ranged",
			"action":      "kite_away",
			"distance":    dist,
		}).Trace("Ranged kiting")
	} else if dist > cfg.MaxRange {
		// Too far, advance
		logrus.WithFields(logrus.Fields{
			"system_name": "RoleBasedAISystem",
			"role":        "Ranged",
			"action":      "advance",
			"distance":    dist,
		}).Trace("Ranged advancing")
	}
}

func (s *RoleBasedAISystem) behaviorHealer(w *engine.World, e engine.Entity, pos *PositionComponent, healthPct float64, cfg RoleConfig, deltaTime float64) {
	// Find wounded allies
	squadComp, hasSquad := w.GetComponent(e, reflect.TypeOf(&SquadTacticsComponent{}))
	if !hasSquad {
		return
	}

	stc := squadComp.(*SquadTacticsComponent)
	squad := s.getOrCreateSquad(stc.SquadID)

	// Look for wounded squad members
	for _, memberID := range squad.Members {
		if memberID == entityIDString(e) {
			continue
		}
		// In real impl, would check member health and heal
		logrus.WithFields(logrus.Fields{
			"system_name": "RoleBasedAISystem",
			"role":        "Healer",
			"action":      "support_check",
		}).Trace("Healer checking allies")
		break
	}
}

func (s *RoleBasedAISystem) behaviorAmbusher(pos, targetPos *PositionComponent, dist, healthPct float64, cfg RoleConfig, deltaTime float64) {
	// Wait for close approach, then burst
	if dist <= cfg.PreferredRange && dist > cfg.MinRange {
		logrus.WithFields(logrus.Fields{
			"system_name": "RoleBasedAISystem",
			"role":        "Ambusher",
			"action":      "ambush_ready",
			"distance":    dist,
		}).Trace("Ambusher ready")
	}
}

func (s *RoleBasedAISystem) behaviorScout(w *engine.World, e engine.Entity, pos, targetPos *PositionComponent, dist float64, cfg RoleConfig, deltaTime float64) {
	// Alert squad on player sight
	if cfg.AlertsOnPlayerSight && dist <= cfg.MaxRange {
		squadComp, hasSquad := w.GetComponent(e, reflect.TypeOf(&SquadTacticsComponent{}))
		if hasSquad {
			stc := squadComp.(*SquadTacticsComponent)
			squad := s.getOrCreateSquad(stc.SquadID)
			squad.RaiseAlert(0.5)

			logrus.WithFields(logrus.Fields{
				"system_name": "RoleBasedAISystem",
				"role":        "Scout",
				"action":      "alert_squad",
				"alert_level": squad.AlertLevel,
			}).Debug("Scout alerting squad")
		}
	}
}

func (s *RoleBasedAISystem) getOrCreateSquad(squadID string) *SquadTactics {
	if squad, exists := s.squads[squadID]; exists {
		return squad
	}
	squad := NewSquadTactics(squadID)
	s.squads[squadID] = squad
	return squad
}

func (s *RoleBasedAISystem) findVisibleTargets(w *engine.World, members []engine.Entity) []string {
	targetSet := make(map[string]bool)

	for _, e := range members {
		targetComp, hasTarget := w.GetComponent(e, reflect.TypeOf(&TargetComponent{}))
		if !hasTarget {
			continue
		}
		tc := targetComp.(*TargetComponent)
		if tc.TargetID != 0 {
			targetSet[entityIDString(tc.TargetID)] = true
		}
	}

	targets := make([]string, 0, len(targetSet))
	for t := range targetSet {
		targets = append(targets, t)
	}
	return targets
}

func (s *RoleBasedAISystem) getTargetPosition(w *engine.World, targetID string) [2]float64 {
	// In real impl, parse targetID and get position
	return [2]float64{0, 0}
}

func entityIDString(e engine.Entity) string {
	// Convert entity ID to string
	return fmt.Sprintf("entity-%d", e)
}

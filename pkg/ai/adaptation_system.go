// Package ai - System for adaptive enemy AI that learns from player behavior
package ai

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// AdaptiveAISystem observes player behavior and adjusts enemy AI accordingly.
type AdaptiveAISystem struct {
	genreID             string
	gameTime            float64
	observationInterval float64
	adaptationInterval  float64
	lastObservation     float64
	lastAdaptation      float64
	logger              *logrus.Entry
}

// NewAdaptiveAISystem creates an adaptive AI system.
func NewAdaptiveAISystem(genreID string) *AdaptiveAISystem {
	return &AdaptiveAISystem{
		genreID:             genreID,
		observationInterval: 2.0,  // Observe player every 2 seconds
		adaptationInterval:  10.0, // Adapt enemy AI every 10 seconds
		logger: logrus.WithFields(logrus.Fields{
			"system": "adaptive_ai",
			"genre":  genreID,
		}),
	}
}

// Update runs adaptive AI logic.
func (s *AdaptiveAISystem) Update(w *engine.World) {
	deltaTime := common.DeltaTime
	s.gameTime += deltaTime

	// Periodic player observation
	if s.gameTime-s.lastObservation >= s.observationInterval {
		s.observePlayer(w)
		s.lastObservation = s.gameTime
	}

	// Periodic AI adaptation
	if s.gameTime-s.lastAdaptation >= s.adaptationInterval {
		s.adaptEnemies(w)
		s.lastAdaptation = s.gameTime
	}
}

// observePlayer analyzes player entity and records behavior patterns.
func (s *AdaptiveAISystem) observePlayer(w *engine.World) {
	// Find player entity with profile component
	profileType := reflect.TypeOf(&PlayerProfileComponent{})
	posType := reflect.TypeOf(&engine.Position{})
	healthType := reflect.TypeOf(&engine.Health{})

	playerEntities := w.Query(profileType, posType, healthType)
	if len(playerEntities) == 0 {
		return
	}

	playerEntity := playerEntities[0]
	profileComp, _ := w.GetComponent(playerEntity, profileType)
	profile := profileComp.(*PlayerProfileComponent).Profile

	// Analyze current player state
	posComp, _ := w.GetComponent(playerEntity, posType)
	playerPos := posComp.(*engine.Position)

	// Find nearest enemy to determine engagement range
	nearestEnemyDist := s.findNearestEnemyDistance(w, playerPos)
	if nearestEnemyDist > 0 {
		profile.UpdateRange(nearestEnemyDist)
	}

	// Detect active weapon type (stub - would integrate with weapon system)
	// For now, infer from engagement distance
	tactic := s.inferTactic(profile, nearestEnemyDist)
	if tactic != TacticUnknown {
		obs := TacticObservation{
			Tactic:     tactic,
			Timestamp:  s.gameTime,
			Confidence: 0.8,
		}
		profile.RecordObservation(obs)

		s.logger.WithFields(logrus.Fields{
			"tactic":       tactic,
			"distance":     nearestEnemyDist,
			"observations": profile.ObservationCount,
		}).Debug("Recorded player tactic observation")
	}
}

// inferTactic infers player tactic from current state.
func (s *AdaptiveAISystem) inferTactic(profile *PlayerBehaviorProfile, engagementDist float64) PlayerTactic {
	if engagementDist <= 0 {
		return TacticUnknown
	}

	// Close range suggests melee
	if engagementDist < 3.0 {
		return TacticRushMelee
	}

	// Mid range
	if engagementDist < 8.0 {
		if profile.PrefersCover > 0.5 {
			return TacticCoverBased
		}
		return TacticUnknown
	}

	// Long range suggests ranged/kiting
	if engagementDist > 12.0 {
		return TacticKiteRanged
	}

	return TacticUnknown
}

// findNearestEnemyDistance finds distance to nearest enemy entity.
func (s *AdaptiveAISystem) findNearestEnemyDistance(w *engine.World, playerPos *engine.Position) float64 {
	posType := reflect.TypeOf(&engine.Position{})
	healthType := reflect.TypeOf(&engine.Health{})
	roleType := reflect.TypeOf(&EnemyRoleComponent{})

	enemies := w.Query(roleType, posType, healthType)

	minDist := -1.0
	for _, enemy := range enemies {
		enemyPosComp, _ := w.GetComponent(enemy, posType)
		enemyPos := enemyPosComp.(*engine.Position)

		dx := enemyPos.X - playerPos.X
		dy := enemyPos.Y - playerPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if minDist < 0 || dist < minDist {
			minDist = dist
		}
	}

	return minDist
}

// adaptEnemies updates enemy AI behavior based on learned player profile.
func (s *AdaptiveAISystem) adaptEnemies(w *engine.World) {
	// Get player profile
	profileType := reflect.TypeOf(&PlayerProfileComponent{})
	playerEntities := w.Query(profileType)
	if len(playerEntities) == 0 {
		return
	}

	profileComp, _ := w.GetComponent(playerEntities[0], profileType)
	playerProfile := profileComp.(*PlayerProfileComponent).Profile

	// Compute adaptation strategy
	adaptation := ComputeAdaptation(playerProfile)

	dominant := playerProfile.GetDominantTactic()
	s.logger.WithFields(logrus.Fields{
		"dominant_tactic":     dominant,
		"observations":        playerProfile.ObservationCount,
		"avg_range":           playerProfile.AverageEngagementRange,
		"melee_freq":          playerProfile.MeleeFrequency,
		"ranged_freq":         playerProfile.RangedFrequency,
		"range_multiplier":    adaptation.PreferredRangeMultiplier,
		"focus_fire_priority": adaptation.FocusFirePriority,
		"flanking_priority":   adaptation.FlankingPriority,
	}).Info("Adapting enemy AI to player tactics")

	// Apply adaptation to all enemy entities with roles
	roleType := reflect.TypeOf(&EnemyRoleComponent{})
	adaptType := reflect.TypeOf(&AdaptationComponent{})

	enemies := w.Query(roleType)
	for _, enemy := range enemies {
		// Get or create adaptation component
		adaptComp, hasAdapt := w.GetComponent(enemy, adaptType)
		if !hasAdapt {
			// Create new adaptation component
			newAdapt := &AdaptationComponent{
				CurrentAdaptation: adaptation,
				LastUpdateTime:    s.gameTime,
			}
			w.AddComponent(enemy, newAdapt)
		} else {
			// Update existing
			ac := adaptComp.(*AdaptationComponent)
			ac.CurrentAdaptation = adaptation
			ac.LastUpdateTime = s.gameTime
		}

		// Apply adaptation to role config
		roleComp, _ := w.GetComponent(enemy, roleType)
		rc := roleComp.(*EnemyRoleComponent)

		s.applyAdaptationToRole(rc, adaptation)
	}
}

// applyAdaptationToRole modifies role configuration based on adaptation.
func (s *AdaptiveAISystem) applyAdaptationToRole(role *EnemyRoleComponent, adapt Adaptation) {
	// Adjust range preferences
	role.Config.PreferredRange *= adapt.PreferredRangeMultiplier
	role.Config.MinRange *= adapt.PreferredRangeMultiplier * 0.8
	role.Config.MaxRange *= adapt.PreferredRangeMultiplier * 1.2

	// Adjust aggression and retreat
	if role.Config.RetreatHealthPct > 0 {
		role.Config.RetreatHealthPct = adapt.RetreatThreshold
	}

	// Adjust aggression level based on pursuit
	role.Config.AggressionLevel = lerp(role.Config.AggressionLevel, adapt.PursuitAggression, 0.3)

	// Apply tactical flags
	if adapt.UseCover && role.Role == RoleRanged {
		role.Config.UsesCover = true
	}

	if adapt.FlankingPriority > 0.7 && role.Role == RoleScout {
		role.Config.AlertsOnPlayerSight = true
	}

	s.logger.WithFields(logrus.Fields{
		"role":              role.Role,
		"preferred_range":   role.Config.PreferredRange,
		"retreat_threshold": role.Config.RetreatHealthPct,
		"uses_cover":        role.Config.UsesCover,
	}).Trace("Applied adaptation to enemy role")
}

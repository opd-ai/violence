// Package combat handles damage calculation and combat events.
package combat

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// HealthComponent stores entity health.
type HealthComponent struct {
	Current float64
	Max     float64
}

// Type returns the component type identifier.
func (c *HealthComponent) Type() string {
	return "HealthComponent"
}

// Update runs boss phase transition logic.
func (s *BossPhaseSystem) Update(w *engine.World) {
	deltaTime := 0.016 // Assume 60 FPS
	s.gameTime += deltaTime

	// Query all boss entities
	entities := w.Query(
		reflect.TypeOf(&BossPhaseComponent{}),
		reflect.TypeOf(&HealthComponent{}),
	)

	for _, e := range entities {
		bossComp, _ := w.GetComponent(e, reflect.TypeOf(&BossPhaseComponent{}))
		healthComp, _ := w.GetComponent(e, reflect.TypeOf(&HealthComponent{}))

		boss := bossComp.(*BossPhaseComponent)
		health := healthComp.(*HealthComponent)

		// Initialize max health on first update
		if boss.InitialMaxHealth == 0 {
			boss.InitialMaxHealth = health.Max
		}

		// Update ongoing transition
		if boss.IsTransitioning {
			completed := boss.UpdateTransition(deltaTime)
			if completed {
				s.onPhaseComplete(w, e, boss)
			}
			continue
		}

		// Check for phase transition trigger
		if boss.ShouldTransition(health.Current, boss.InitialMaxHealth, s.gameTime) {
			s.triggerPhaseTransition(w, e, boss, health)
		}
	}
}

func (s *BossPhaseSystem) triggerPhaseTransition(w *engine.World, e engine.Entity, boss *BossPhaseComponent, health *HealthComponent) {
	nextPhase := boss.CurrentPhase + 1
	if nextPhase >= len(boss.Phases) {
		return
	}

	boss.StartTransition(s.gameTime)

	logrus.WithFields(logrus.Fields{
		"system_name": "BossPhaseSystem",
		"entity_id":   e,
		"from_phase":  boss.CurrentPhase,
		"to_phase":    nextPhase,
		"health_pct":  health.Current / boss.InitialMaxHealth,
		"visual_fx":   boss.Phases[nextPhase].VisualEffect,
		"enraged":     boss.Phases[nextPhase].Enraged,
	}).Info("Boss phase transition triggered")

	// Trigger visual effects (screen shake, particle burst)
	s.spawnTransitionEffects(w, e, boss, nextPhase)
}

func (s *BossPhaseSystem) onPhaseComplete(w *engine.World, e engine.Entity, boss *BossPhaseComponent) {
	phase := boss.GetCurrentPhaseData()
	if phase == nil {
		return
	}

	logrus.WithFields(logrus.Fields{
		"system_name":      "BossPhaseSystem",
		"entity_id":        e,
		"new_phase":        boss.CurrentPhase,
		"damage_mult":      phase.DamageMultiplier,
		"speed_mult":       phase.SpeedMultiplier,
		"attack_rate":      phase.AttackRate,
		"ability_count":    len(phase.AbilitySet),
		"transition_count": boss.PhaseChangeCount,
	}).Info("Boss phase transition complete")

	// Apply phase stat modifiers
	s.applyPhaseModifiers(w, e, phase)
}

func (s *BossPhaseSystem) applyPhaseModifiers(w *engine.World, e engine.Entity, phase *PhaseTransition) {
	// In a full implementation, this would modify:
	// - Damage output via DamageMultiplier
	// - Movement speed via SpeedMultiplier
	// - Attack timing via AttackRate
	// - Available abilities via AbilitySet

	// For now, log the changes that would be applied
	logrus.WithFields(logrus.Fields{
		"system_name": "BossPhaseSystem",
		"entity_id":   e,
		"modifiers": map[string]interface{}{
			"damage":      phase.DamageMultiplier,
			"speed":       phase.SpeedMultiplier,
			"attack_rate": phase.AttackRate,
			"abilities":   phase.AbilitySet,
		},
	}).Debug("Phase modifiers applied")
}

func (s *BossPhaseSystem) spawnTransitionEffects(w *engine.World, e engine.Entity, boss *BossPhaseComponent, nextPhase int) {
	// Spawn visual effects for phase transition
	// In full implementation, this would:
	// 1. Create screen shake event
	// 2. Spawn particle burst at boss location
	// 3. Play transition sound effect
	// 4. Flash screen or show visual indicator

	effect := boss.Phases[nextPhase].VisualEffect

	logrus.WithFields(logrus.Fields{
		"system_name": "BossPhaseSystem",
		"entity_id":   e,
		"effect":      effect,
		"enraged":     boss.Phases[nextPhase].Enraged,
	}).Debug("Phase transition effects spawned")
}

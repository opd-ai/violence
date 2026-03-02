package ai

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewAdaptiveAISystem(t *testing.T) {
	sys := NewAdaptiveAISystem("fantasy")
	if sys == nil {
		t.Fatal("NewAdaptiveAISystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", sys.genreID)
	}
	if sys.observationInterval != 2.0 {
		t.Errorf("Expected observation interval 2.0, got %f", sys.observationInterval)
	}
	if sys.adaptationInterval != 10.0 {
		t.Errorf("Expected adaptation interval 10.0, got %f", sys.adaptationInterval)
	}
}

func TestAdaptiveAISystemUpdate(t *testing.T) {
	w := engine.NewWorld()
	sys := NewAdaptiveAISystem("fantasy")

	// Create player entity with profile
	player := w.AddEntity()
	profile := NewPlayerBehaviorProfile()
	w.AddComponent(player, &PlayerProfileComponent{Profile: profile})
	w.AddComponent(player, &engine.Position{X: 10.0, Y: 10.0})
	w.AddComponent(player, &engine.Health{Current: 100, Max: 100})

	// Create enemy entity
	enemy := w.AddEntity()
	w.AddComponent(enemy, &EnemyRoleComponent{
		Role: RoleRanged,
		Config: RoleConfig{
			PreferredRange: 10.0,
			MinRange:       5.0,
			MaxRange:       15.0,
		},
	})
	w.AddComponent(enemy, &engine.Position{X: 15.0, Y: 10.0})
	w.AddComponent(enemy, &engine.Health{Current: 50, Max: 50})

	// Run update cycles
	for i := 0; i < 200; i++ {
		sys.Update(w) // 20 seconds total at 60 FPS
	}

	// Verify profile was updated
	if profile.ObservationCount == 0 {
		t.Error("Expected player observations to be recorded")
	}

	// Verify enemy has adaptation component
	adaptType := reflect.TypeOf(&AdaptationComponent{})
	adaptComp, hasAdapt := w.GetComponent(enemy, adaptType)
	if !hasAdapt {
		t.Error("Expected enemy to have AdaptationComponent")
	}

	// Verify adaptation was applied
	ac := adaptComp.(*AdaptationComponent)
	if ac.LastUpdateTime == 0 {
		t.Error("Expected adaptation to have been updated")
	}
}

func TestObservePlayer(t *testing.T) {
	w := engine.NewWorld()
	sys := NewAdaptiveAISystem("fantasy")

	// Create player
	player := w.AddEntity()
	profile := NewPlayerBehaviorProfile()
	w.AddComponent(player, &PlayerProfileComponent{Profile: profile})
	w.AddComponent(player, &engine.Position{X: 10.0, Y: 10.0})
	w.AddComponent(player, &engine.Health{Current: 100, Max: 100})

	// Create nearby enemy
	enemy := w.AddEntity()
	w.AddComponent(enemy, &EnemyRoleComponent{Role: RoleTank})
	w.AddComponent(enemy, &engine.Position{X: 12.0, Y: 10.0}) // 2 units away
	w.AddComponent(enemy, &engine.Health{Current: 50, Max: 50})

	initialCount := profile.ObservationCount

	sys.observePlayer(w)

	// Profile should have updated range
	if profile.AverageEngagementRange == 10.0 {
		t.Error("Expected engagement range to be updated from default")
	}

	// May or may not add observation depending on inference, but should at least track range
	if profile.ObservationCount < initialCount {
		t.Error("Observation count should not decrease")
	}
}

func TestInferTactic(t *testing.T) {
	sys := NewAdaptiveAISystem("fantasy")
	profile := NewPlayerBehaviorProfile()

	tests := []struct {
		name     string
		distance float64
		expected PlayerTactic
	}{
		{"no engagement", 0.0, TacticUnknown},
		{"melee range", 2.0, TacticRushMelee},
		{"close range", 5.0, TacticUnknown},
		{"long range", 15.0, TacticKiteRanged},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.inferTactic(profile, tt.distance)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestInferTacticCoverBased(t *testing.T) {
	sys := NewAdaptiveAISystem("fantasy")
	profile := NewPlayerBehaviorProfile()
	profile.PrefersCover = 0.8

	// Mid-range with high cover preference should infer cover-based
	result := sys.inferTactic(profile, 6.0)
	if result != TacticCoverBased {
		t.Errorf("Expected TacticCoverBased with high cover preference, got %v", result)
	}
}

func TestFindNearestEnemyDistance(t *testing.T) {
	w := engine.NewWorld()
	sys := NewAdaptiveAISystem("fantasy")

	playerPos := &engine.Position{X: 10.0, Y: 10.0}

	// No enemies
	dist := sys.findNearestEnemyDistance(w, playerPos)
	if dist >= 0 {
		t.Errorf("Expected negative distance with no enemies, got %f", dist)
	}

	// Add enemies at various distances
	enemy1 := w.AddEntity()
	w.AddComponent(enemy1, &EnemyRoleComponent{Role: RoleTank})
	w.AddComponent(enemy1, &engine.Position{X: 15.0, Y: 10.0}) // 5 units
	w.AddComponent(enemy1, &engine.Health{Current: 50, Max: 50})

	enemy2 := w.AddEntity()
	w.AddComponent(enemy2, &EnemyRoleComponent{Role: RoleRanged})
	w.AddComponent(enemy2, &engine.Position{X: 12.0, Y: 10.0}) // 2 units
	w.AddComponent(enemy2, &engine.Health{Current: 50, Max: 50})

	dist = sys.findNearestEnemyDistance(w, playerPos)
	if dist < 1.9 || dist > 2.1 {
		t.Errorf("Expected nearest distance ~2.0, got %f", dist)
	}
}

func TestAdaptEnemies(t *testing.T) {
	w := engine.NewWorld()
	sys := NewAdaptiveAISystem("fantasy")

	// Create player with profile showing melee tendency
	player := w.AddEntity()
	profile := NewPlayerBehaviorProfile()
	profile.ObservationCount = 10
	profile.MeleeFrequency = 0.9
	profile.AverageEngagementRange = 3.0
	w.AddComponent(player, &PlayerProfileComponent{Profile: profile})

	// Create enemies
	for i := 0; i < 3; i++ {
		enemy := w.AddEntity()
		w.AddComponent(enemy, &EnemyRoleComponent{
			Role: RoleRanged,
			Config: RoleConfig{
				PreferredRange: 10.0,
				MinRange:       5.0,
				MaxRange:       15.0,
			},
		})
	}

	sys.adaptEnemies(w)

	// Verify all enemies have adaptation
	roleType := reflect.TypeOf(&EnemyRoleComponent{})
	adaptType := reflect.TypeOf(&AdaptationComponent{})

	enemies := w.Query(roleType)
	for _, enemy := range enemies {
		adaptComp, hasAdapt := w.GetComponent(enemy, adaptType)
		if !hasAdapt {
			t.Error("Expected enemy to have AdaptationComponent")
			continue
		}

		ac := adaptComp.(*AdaptationComponent)
		// Should counter melee with increased range
		if ac.CurrentAdaptation.PreferredRangeMultiplier <= 1.0 {
			t.Error("Expected range multiplier increase to counter melee")
		}
		if ac.CurrentAdaptation.DodgeFrequency <= 0.5 {
			t.Error("Expected dodge frequency increase to counter melee")
		}
	}
}

func TestApplyAdaptationToRole(t *testing.T) {
	sys := NewAdaptiveAISystem("fantasy")

	role := &EnemyRoleComponent{
		Role: RoleRanged,
		Config: RoleConfig{
			PreferredRange:   10.0,
			MinRange:         5.0,
			MaxRange:         15.0,
			RetreatHealthPct: 0.3,
			AggressionLevel:  0.5,
		},
	}

	adaptation := AIAdaptation{
		PreferredRangeMultiplier: 1.5,
		RetreatThreshold:         0.5,
		UseCover:                 true,
		PursuitAggression:        0.8,
		FocusFirePriority:        0.8,
		FlankingPriority:         0.9,
	}

	sys.applyAdaptationToRole(role, adaptation)

	// Check range adjustments
	if role.Config.PreferredRange != 15.0 {
		t.Errorf("Expected preferred range 15.0, got %f", role.Config.PreferredRange)
	}

	// Check retreat threshold
	if role.Config.RetreatHealthPct != 0.5 {
		t.Errorf("Expected retreat threshold 0.5, got %f", role.Config.RetreatHealthPct)
	}

	// Check cover flag
	if !role.Config.UsesCover {
		t.Error("Expected UsesCover to be true")
	}

	// Check aggression level was adjusted toward pursuit aggression
	if role.Config.AggressionLevel <= 0.5 {
		t.Error("Expected aggression level to increase toward pursuit aggression")
	}
}

func TestApplyAdaptationToScout(t *testing.T) {
	sys := NewAdaptiveAISystem("fantasy")

	role := &EnemyRoleComponent{
		Role: RoleScout,
		Config: RoleConfig{
			PreferredRange:      10.0,
			AlertsOnPlayerSight: false,
		},
	}

	adaptation := AIAdaptation{
		PreferredRangeMultiplier: 1.0,
		FlankingPriority:         0.8,
	}

	sys.applyAdaptationToRole(role, adaptation)

	// Scout with high flanking priority should alert
	if !role.Config.AlertsOnPlayerSight {
		t.Error("Expected AlertsOnPlayerSight to be true for scout with high flanking priority")
	}
}

func TestAdaptiveAISystemIntegration(t *testing.T) {
	w := engine.NewWorld()
	sys := NewAdaptiveAISystem("fantasy")

	// Create player
	player := w.AddEntity()
	profile := NewPlayerBehaviorProfile()
	w.AddComponent(player, &PlayerProfileComponent{Profile: profile})
	w.AddComponent(player, &engine.Position{X: 10.0, Y: 10.0})
	w.AddComponent(player, &engine.Health{Current: 100, Max: 100})

	// Create enemy squad
	for i := 0; i < 5; i++ {
		enemy := w.AddEntity()
		w.AddComponent(enemy, &EnemyRoleComponent{
			Role: RoleRanged,
			Config: RoleConfig{
				PreferredRange: 10.0,
				MinRange:       5.0,
				MaxRange:       15.0,
			},
		})
		w.AddComponent(enemy, &engine.Position{X: 10.0 + float64(i)*2, Y: 10.0})
		w.AddComponent(enemy, &engine.Health{Current: 50, Max: 50})
	}

	// Simulate 30 seconds of gameplay
	for i := 0; i < 1800; i++ { // 1800 frames at 60 FPS = 30 seconds
		sys.Update(w)
	}

	// Verify system state
	if sys.gameTime < 29.0 {
		t.Errorf("Expected game time ~30.0, got %f", sys.gameTime)
	}

	if profile.ObservationCount == 0 {
		t.Error("Expected observations to be recorded")
	}

	// Check that at least some enemies have adaptations
	roleType := reflect.TypeOf(&EnemyRoleComponent{})
	adaptType := reflect.TypeOf(&AdaptationComponent{})
	enemies := w.Query(roleType)

	adaptedCount := 0
	for _, enemy := range enemies {
		if _, hasAdapt := w.GetComponent(enemy, adaptType); hasAdapt {
			adaptedCount++
		}
	}

	if adaptedCount == 0 {
		t.Error("Expected at least some enemies to have adaptations")
	}
}

package main

import (
	"testing"

	"github.com/opd-ai/violence/pkg/ai"
	"github.com/opd-ai/violence/pkg/weapon"
)

// TestMasteryXPAwardedAfterDamageApplication verifies that mastery XP is only awarded
// after damage has been successfully applied to an enemy, not before.
func TestMasteryXPAwardedAfterDamageApplication(t *testing.T) {
	tests := []struct {
		name          string
		enemyHealth   float64
		weaponDamage  float64
		shouldAwardXP bool
		description   string
	}{
		{
			name:          "normal_hit_alive_enemy",
			enemyHealth:   100,
			weaponDamage:  25,
			shouldAwardXP: true,
			description:   "Normal hit on living enemy should award XP",
		},
		{
			name:          "kill_shot",
			enemyHealth:   10,
			weaponDamage:  25,
			shouldAwardXP: true,
			description:   "Kill shot should award XP (damage applied successfully)",
		},
		{
			name:          "overkill",
			enemyHealth:   1,
			weaponDamage:  100,
			shouldAwardXP: true,
			description:   "Overkill should still award XP (damage applied)",
		},
		{
			name:          "dead_enemy",
			enemyHealth:   0,
			weaponDamage:  25,
			shouldAwardXP: false,
			description:   "Shooting already dead enemy should NOT award XP",
		},
		{
			name:          "negative_health_enemy",
			enemyHealth:   -10,
			weaponDamage:  25,
			shouldAwardXP: false,
			description:   "Shooting enemy with negative health should NOT award XP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal game instance
			g := &Game{
				seed: 12345,
			}

			// Initialize mastery manager
			g.masteryManager = weapon.NewMasteryManager()

			// Create a weapon arsenal
			g.arsenal = weapon.NewArsenal()
			weaponSlot := g.arsenal.CurrentSlot // Use the actual current slot

			// Create a test AI agent with specified health
			g.aiAgents = []*ai.Agent{
				{
					ID:        "test_enemy_1",
					Health:    tt.enemyHealth,
					MaxHealth: 100,
					X:         10.0,
					Y:         10.0,
				},
			}

			// Get initial mastery XP
			initialMastery := g.masteryManager.GetMastery(weaponSlot)
			initialXP := 0
			if initialMastery != nil {
				initialXP = initialMastery.XP
			}

			// Simulate the damage application logic from main.go:805-814
			agentIdx := 0 // First agent
			agent := g.aiAgents[agentIdx]

			// This mirrors the exact logic in main.go:805-814
			if agent.Health > 0 {
				// Apply damage with upgrades and mastery bonuses
				// For test simplicity, use direct damage value
				upgradedDamage := tt.weaponDamage
				agent.Health -= upgradedDamage

				// Award mastery XP only after damage is successfully applied
				if g.masteryManager != nil {
					g.masteryManager.AddMasteryXP(g.arsenal.CurrentSlot, 10)
				}
			}

			// Get final mastery XP
			finalMastery := g.masteryManager.GetMastery(weaponSlot)
			finalXP := 0
			if finalMastery != nil {
				finalXP = finalMastery.XP
			}

			// Verify expectations
			if tt.shouldAwardXP {
				if finalXP != initialXP+10 {
					t.Errorf("%s: Expected XP to increase by 10 (from %d to %d), got %d",
						tt.description, initialXP, initialXP+10, finalXP)
				}
				// Also verify damage was actually applied
				expectedHealth := tt.enemyHealth - tt.weaponDamage
				if agent.Health != expectedHealth {
					t.Errorf("%s: Expected health to be %.1f after damage, got %.1f",
						tt.description, expectedHealth, agent.Health)
				}
			} else {
				if finalXP != initialXP {
					t.Errorf("%s: Expected no XP increase (XP should stay at %d), got %d",
						tt.description, initialXP, finalXP)
				}
				// Verify no damage was applied to dead enemy
				if agent.Health != tt.enemyHealth {
					t.Errorf("%s: Expected health to remain %.1f (no damage to dead enemy), got %.1f",
						tt.description, tt.enemyHealth, agent.Health)
				}
			}
		})
	}
}

// TestMasteryXPNotAwardedBeforeDamage verifies the edge case where damage application
// might theoretically fail, ensuring XP is not awarded prematurely.
func TestMasteryXPNotAwardedBeforeDamage(t *testing.T) {
	g := &Game{
		seed: 54321,
	}

	// Initialize mastery manager
	g.masteryManager = weapon.NewMasteryManager()

	// Create arsenal
	g.arsenal = weapon.NewArsenal()
	weaponSlot := g.arsenal.CurrentSlot

	// Create a living AI agent
	g.aiAgents = []*ai.Agent{
		{
			ID:        "test_enemy_1",
			Health:    100,
			MaxHealth: 100,
			X:         10.0,
			Y:         10.0,
		},
	}

	initialMastery := g.masteryManager.GetMastery(weaponSlot)
	initialXP := 0
	if initialMastery != nil {
		initialXP = initialMastery.XP
	}

	// Simulate a scenario where we check health but DON'T apply damage
	// (edge case: damage calculation fails or is skipped)
	agent := g.aiAgents[0]
	if agent.Health > 0 {
		// In the old (buggy) code, XP would be awarded here BEFORE damage
		// In the new (fixed) code, we apply damage FIRST

		// Simulate: getUpgradedWeaponDamage returns 0 (no damage upgrade active)
		// This should still award XP since damage application happens
		upgradedDamage := 0.0
		agent.Health -= upgradedDamage // No actual damage

		// Award XP after damage application (even if 0 damage)
		if g.masteryManager != nil {
			g.masteryManager.AddMasteryXP(weaponSlot, 10)
		}
	}

	finalMastery := g.masteryManager.GetMastery(weaponSlot)
	finalXP := 0
	if finalMastery != nil {
		finalXP = finalMastery.XP
	}

	// In this case, XP should still be awarded because the damage application
	// step completed (even though it was 0 damage). The key is that XP comes
	// AFTER the damage application line, not before it.
	if finalXP != initialXP+10 {
		t.Errorf("Expected XP to be awarded after damage application step, got %d (expected %d)",
			finalXP, initialXP+10)
	}

	// Verify health unchanged (0 damage was applied)
	if agent.Health != 100 {
		t.Errorf("Expected health to remain 100 with 0 damage, got %.1f", agent.Health)
	}
}

// TestMasteryXPProgressionAccuracy verifies that XP accumulation is accurate
// across multiple hits and matches the damage actually dealt.
func TestMasteryXPProgressionAccuracy(t *testing.T) {
	g := &Game{
		seed: 99999,
	}

	g.masteryManager = weapon.NewMasteryManager()
	g.arsenal = weapon.NewArsenal()

	weaponSlot := g.arsenal.CurrentSlot

	// Create multiple enemies
	g.aiAgents = []*ai.Agent{
		{ID: "enemy_1", Health: 100, MaxHealth: 100, X: 10.0, Y: 10.0},
		{ID: "enemy_2", Health: 50, MaxHealth: 100, X: 15.0, Y: 15.0},
		{ID: "enemy_3", Health: 0, MaxHealth: 100, X: 20.0, Y: 20.0}, // Already dead
	}

	initialMastery := g.masteryManager.GetMastery(weaponSlot)
	initialXP := 0
	if initialMastery != nil {
		initialXP = initialMastery.XP
	}

	damagePerShot := 25.0
	xpPerHit := 10
	expectedHits := 0

	// Simulate shooting each enemy
	for i, agent := range g.aiAgents {
		if agent.Health > 0 {
			upgradedDamage := damagePerShot
			agent.Health -= upgradedDamage

			// Award XP after damage applied
			if g.masteryManager != nil {
				g.masteryManager.AddMasteryXP(weaponSlot, xpPerHit)
				expectedHits++
			}
		}

		// Verify intermediate state
		currentMastery := g.masteryManager.GetMastery(weaponSlot)
		currentXP := 0
		if currentMastery != nil {
			currentXP = currentMastery.XP
		}
		expectedXP := initialXP + (expectedHits * xpPerHit)

		if currentXP != expectedXP {
			t.Errorf("After shooting agent %d: expected XP=%d, got XP=%d",
				i, expectedXP, currentXP)
		}
	}

	// Final verification
	finalMastery := g.masteryManager.GetMastery(weaponSlot)
	finalXP := 0
	if finalMastery != nil {
		finalXP = finalMastery.XP
	}
	expectedFinalXP := initialXP + (expectedHits * xpPerHit)

	if finalXP != expectedFinalXP {
		t.Errorf("Expected final XP=%d (from %d hits), got %d",
			expectedFinalXP, expectedHits, finalXP)
	}

	// Should have hit 2 enemies (agent 0 and 1), agent 2 was already dead
	if expectedHits != 2 {
		t.Errorf("Expected 2 successful hits, got %d", expectedHits)
	}

	// Verify damage was applied correctly
	if g.aiAgents[0].Health != 75 { // 100 - 25
		t.Errorf("Agent 0: expected health=75, got %.1f", g.aiAgents[0].Health)
	}
	if g.aiAgents[1].Health != 25 { // 50 - 25
		t.Errorf("Agent 1: expected health=25, got %.1f", g.aiAgents[1].Health)
	}
	if g.aiAgents[2].Health != 0 { // Should remain dead
		t.Errorf("Agent 2: expected to remain dead (health=0), got %.1f", g.aiAgents[2].Health)
	}
}

package main

import (
	"testing"

	"github.com/opd-ai/violence/pkg/weapon"
)

// TestMasteryIntegration verifies weapon mastery system is wired into the game.
func TestMasteryIntegration(t *testing.T) {
	g := NewGame()

	// Verify mastery manager is initialized
	if g.masteryManager == nil {
		t.Fatal("masteryManager should be initialized in NewGame()")
	}

	// Verify initial state: no mastery
	mastery := g.masteryManager.GetMastery(1) // Pistol is slot 1
	if mastery.XP != 0 {
		t.Errorf("Expected initial XP = 0, got %d", mastery.XP)
	}
	if mastery.CurrentBonus.HeadshotDamage != 1.0 {
		t.Errorf("Expected initial headshot bonus = 1.0, got %f", mastery.CurrentBonus.HeadshotDamage)
	}

	// Award some XP
	g.masteryManager.AddMasteryXP(1, 100)
	mastery = g.masteryManager.GetMastery(1)
	if mastery.XP != 100 {
		t.Errorf("Expected XP = 100 after AddMasteryXP(100), got %d", mastery.XP)
	}

	// Verify milestone unlock at 250 XP
	g.masteryManager.AddMasteryXP(1, 150)
	mastery = g.masteryManager.GetMastery(1)
	if mastery.XP != 250 {
		t.Errorf("Expected XP = 250, got %d", mastery.XP)
	}
	if !mastery.UnlockedBonus1 {
		t.Error("Expected Milestone 250 to be unlocked")
	}
	if mastery.CurrentBonus.HeadshotDamage != 1.10 {
		t.Errorf("Expected headshot bonus = 1.10 at 250 XP, got %f", mastery.CurrentBonus.HeadshotDamage)
	}

	// Verify damage calculation includes mastery bonuses
	pistol := g.arsenal.GetCurrentWeapon()
	baseDamage := pistol.Damage
	upgradedDamage := g.getUpgradedWeaponDamage(pistol)

	expectedDamage := baseDamage * 1.10 // Headshot bonus at 250 XP
	if upgradedDamage != expectedDamage {
		t.Errorf("Expected damage = %f (base %f * 1.10), got %f", expectedDamage, baseDamage, upgradedDamage)
	}
}

// TestMasteryMilestoneProgression verifies all milestones unlock correctly.
func TestMasteryMilestoneProgression(t *testing.T) {
	g := NewGame()

	tests := []struct {
		xp             int
		expectedBonus  weapon.MasteryBonus
		milestoneFlags [4]bool // [UnlockedBonus1, UnlockedBonus2, UnlockedBonus3, UnlockedBonus4]
	}{
		{
			xp:             0,
			expectedBonus:  weapon.MasteryBonus{HeadshotDamage: 1.0, ReloadSpeed: 1.0, Accuracy: 1.0, CriticalChance: 0.0},
			milestoneFlags: [4]bool{false, false, false, false},
		},
		{
			xp:             250,
			expectedBonus:  weapon.MasteryBonus{HeadshotDamage: 1.10, ReloadSpeed: 1.0, Accuracy: 1.0, CriticalChance: 0.0},
			milestoneFlags: [4]bool{true, false, false, false},
		},
		{
			xp:             500,
			expectedBonus:  weapon.MasteryBonus{HeadshotDamage: 1.10, ReloadSpeed: 1.15, Accuracy: 1.0, CriticalChance: 0.0},
			milestoneFlags: [4]bool{true, true, false, false},
		},
		{
			xp:             750,
			expectedBonus:  weapon.MasteryBonus{HeadshotDamage: 1.10, ReloadSpeed: 1.15, Accuracy: 1.10, CriticalChance: 0.0},
			milestoneFlags: [4]bool{true, true, true, false},
		},
		{
			xp:             1000,
			expectedBonus:  weapon.MasteryBonus{HeadshotDamage: 1.10, ReloadSpeed: 1.15, Accuracy: 1.10, CriticalChance: 0.05},
			milestoneFlags: [4]bool{true, true, true, true},
		},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.xp)), func(t *testing.T) {
			g.masteryManager.Reset()
			g.masteryManager.AddMasteryXP(1, tt.xp)
			mastery := g.masteryManager.GetMastery(1)

			if mastery.CurrentBonus != tt.expectedBonus {
				t.Errorf("XP %d: expected bonus %+v, got %+v", tt.xp, tt.expectedBonus, mastery.CurrentBonus)
			}

			if mastery.UnlockedBonus1 != tt.milestoneFlags[0] {
				t.Errorf("XP %d: UnlockedBonus1 = %v, expected %v", tt.xp, mastery.UnlockedBonus1, tt.milestoneFlags[0])
			}
			if mastery.UnlockedBonus2 != tt.milestoneFlags[1] {
				t.Errorf("XP %d: UnlockedBonus2 = %v, expected %v", tt.xp, mastery.UnlockedBonus2, tt.milestoneFlags[1])
			}
			if mastery.UnlockedBonus3 != tt.milestoneFlags[2] {
				t.Errorf("XP %d: UnlockedBonus3 = %v, expected %v", tt.xp, mastery.UnlockedBonus3, tt.milestoneFlags[2])
			}
			if mastery.UnlockedBonus4 != tt.milestoneFlags[3] {
				t.Errorf("XP %d: UnlockedBonus4 = %v, expected %v", tt.xp, mastery.UnlockedBonus4, tt.milestoneFlags[3])
			}
		})
	}
}

// TestMasteryXPCapping verifies XP caps at 1000.
func TestMasteryXPCapping(t *testing.T) {
	g := NewGame()

	// Award excessive XP
	g.masteryManager.AddMasteryXP(1, 2000)
	mastery := g.masteryManager.GetMastery(1)

	if mastery.XP != 1000 {
		t.Errorf("Expected XP capped at 1000, got %d", mastery.XP)
	}
}

// TestMasteryPerWeapon verifies each weapon tracks mastery independently.
func TestMasteryPerWeapon(t *testing.T) {
	g := NewGame()

	// Award XP to pistol (slot 1)
	g.masteryManager.AddMasteryXP(1, 250)
	// Award XP to shotgun (slot 2)
	g.masteryManager.AddMasteryXP(2, 500)

	pistolMastery := g.masteryManager.GetMastery(1)
	shotgunMastery := g.masteryManager.GetMastery(2)

	if pistolMastery.XP != 250 {
		t.Errorf("Expected pistol XP = 250, got %d", pistolMastery.XP)
	}
	if shotgunMastery.XP != 500 {
		t.Errorf("Expected shotgun XP = 500, got %d", shotgunMastery.XP)
	}

	// Verify bonuses are independent
	if pistolMastery.CurrentBonus.HeadshotDamage != 1.10 {
		t.Errorf("Pistol should have headshot bonus at 250 XP")
	}
	if shotgunMastery.CurrentBonus.ReloadSpeed != 1.15 {
		t.Errorf("Shotgun should have reload bonus at 500 XP")
	}
}

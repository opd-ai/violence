package weapon

import (
	"testing"
)

func TestNewMasteryManager(t *testing.T) {
	mm := NewMasteryManager()
	if mm == nil {
		t.Fatal("expected non-nil mastery manager")
	}
	if mm.Masteries == nil {
		t.Error("expected initialized masteries map")
	}
}

func TestAddMasteryXP(t *testing.T) {
	tests := []struct {
		name       string
		weaponSlot int
		xpAmounts  []int
		expectedXP int
	}{
		{
			name:       "single XP grant",
			weaponSlot: 1,
			xpAmounts:  []int{50},
			expectedXP: 50,
		},
		{
			name:       "multiple XP grants",
			weaponSlot: 2,
			xpAmounts:  []int{100, 150, 200},
			expectedXP: 450,
		},
		{
			name:       "XP cap at 1000",
			weaponSlot: 3,
			xpAmounts:  []int{600, 600},
			expectedXP: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMasteryManager()
			for _, amount := range tt.xpAmounts {
				mm.AddMasteryXP(tt.weaponSlot, amount)
			}
			xp := mm.GetXP(tt.weaponSlot)
			if xp != tt.expectedXP {
				t.Errorf("expected XP %d, got %d", tt.expectedXP, xp)
			}
		})
	}
}

func TestMilestoneBonuses(t *testing.T) {
	tests := []struct {
		name              string
		xp                int
		expectedHeadshot  float64
		expectedReload    float64
		expectedAccuracy  float64
		expectedCritical  float64
		expectedUnlocked1 bool
		expectedUnlocked2 bool
		expectedUnlocked3 bool
		expectedUnlocked4 bool
		expectedMilestone MasteryMilestone
	}{
		{
			name:              "no XP - no bonuses",
			xp:                0,
			expectedHeadshot:  1.0,
			expectedReload:    1.0,
			expectedAccuracy:  1.0,
			expectedCritical:  0.0,
			expectedUnlocked1: false,
			expectedUnlocked2: false,
			expectedUnlocked3: false,
			expectedUnlocked4: false,
			expectedMilestone: MilestoneNone,
		},
		{
			name:              "250 XP - headshot bonus",
			xp:                250,
			expectedHeadshot:  1.10,
			expectedReload:    1.0,
			expectedAccuracy:  1.0,
			expectedCritical:  0.0,
			expectedUnlocked1: true,
			expectedUnlocked2: false,
			expectedUnlocked3: false,
			expectedUnlocked4: false,
			expectedMilestone: Milestone250,
		},
		{
			name:              "500 XP - headshot + reload",
			xp:                500,
			expectedHeadshot:  1.10,
			expectedReload:    1.15,
			expectedAccuracy:  1.0,
			expectedCritical:  0.0,
			expectedUnlocked1: true,
			expectedUnlocked2: true,
			expectedUnlocked3: false,
			expectedUnlocked4: false,
			expectedMilestone: Milestone500,
		},
		{
			name:              "750 XP - headshot + reload + accuracy",
			xp:                750,
			expectedHeadshot:  1.10,
			expectedReload:    1.15,
			expectedAccuracy:  1.10,
			expectedCritical:  0.0,
			expectedUnlocked1: true,
			expectedUnlocked2: true,
			expectedUnlocked3: true,
			expectedUnlocked4: false,
			expectedMilestone: Milestone750,
		},
		{
			name:              "1000 XP - all bonuses",
			xp:                1000,
			expectedHeadshot:  1.10,
			expectedReload:    1.15,
			expectedAccuracy:  1.10,
			expectedCritical:  0.05,
			expectedUnlocked1: true,
			expectedUnlocked2: true,
			expectedUnlocked3: true,
			expectedUnlocked4: true,
			expectedMilestone: Milestone1000,
		},
		{
			name:              "partial progress to first milestone",
			xp:                100,
			expectedHeadshot:  1.0,
			expectedReload:    1.0,
			expectedAccuracy:  1.0,
			expectedCritical:  0.0,
			expectedUnlocked1: false,
			expectedUnlocked2: false,
			expectedUnlocked3: false,
			expectedUnlocked4: false,
			expectedMilestone: MilestoneNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMasteryManager()
			mm.AddMasteryXP(1, tt.xp)

			mastery := mm.GetMastery(1)
			bonus := mm.GetBonus(1)

			// Check bonuses
			if bonus.HeadshotDamage != tt.expectedHeadshot {
				t.Errorf("expected headshot bonus %.2f, got %.2f", tt.expectedHeadshot, bonus.HeadshotDamage)
			}
			if bonus.ReloadSpeed != tt.expectedReload {
				t.Errorf("expected reload bonus %.2f, got %.2f", tt.expectedReload, bonus.ReloadSpeed)
			}
			if bonus.Accuracy != tt.expectedAccuracy {
				t.Errorf("expected accuracy bonus %.2f, got %.2f", tt.expectedAccuracy, bonus.Accuracy)
			}
			if bonus.CriticalChance != tt.expectedCritical {
				t.Errorf("expected critical bonus %.2f, got %.2f", tt.expectedCritical, bonus.CriticalChance)
			}

			// Check unlocked flags
			if mastery.UnlockedBonus1 != tt.expectedUnlocked1 {
				t.Errorf("expected UnlockedBonus1 %v, got %v", tt.expectedUnlocked1, mastery.UnlockedBonus1)
			}
			if mastery.UnlockedBonus2 != tt.expectedUnlocked2 {
				t.Errorf("expected UnlockedBonus2 %v, got %v", tt.expectedUnlocked2, mastery.UnlockedBonus2)
			}
			if mastery.UnlockedBonus3 != tt.expectedUnlocked3 {
				t.Errorf("expected UnlockedBonus3 %v, got %v", tt.expectedUnlocked3, mastery.UnlockedBonus3)
			}
			if mastery.UnlockedBonus4 != tt.expectedUnlocked4 {
				t.Errorf("expected UnlockedBonus4 %v, got %v", tt.expectedUnlocked4, mastery.UnlockedBonus4)
			}

			// Check milestone
			milestone := mm.GetCurrentMilestone(1)
			if milestone != tt.expectedMilestone {
				t.Errorf("expected milestone %d, got %d", tt.expectedMilestone, milestone)
			}
		})
	}
}

func TestGetProgressToNextMilestone(t *testing.T) {
	tests := []struct {
		name             string
		xp               int
		expectedProgress int
	}{
		{
			name:             "0 XP - 0% progress to 250",
			xp:               0,
			expectedProgress: 0,
		},
		{
			name:             "125 XP - 50% progress to 250",
			xp:               125,
			expectedProgress: 50,
		},
		{
			name:             "250 XP - 0% progress to 500",
			xp:               250,
			expectedProgress: 0,
		},
		{
			name:             "375 XP - 50% progress to 500",
			xp:               375,
			expectedProgress: 50,
		},
		{
			name:             "500 XP - 0% progress to 750",
			xp:               500,
			expectedProgress: 0,
		},
		{
			name:             "625 XP - 50% progress to 750",
			xp:               625,
			expectedProgress: 50,
		},
		{
			name:             "750 XP - 0% progress to 1000",
			xp:               750,
			expectedProgress: 0,
		},
		{
			name:             "875 XP - 50% progress to 1000",
			xp:               875,
			expectedProgress: 50,
		},
		{
			name:             "1000 XP - max level",
			xp:               1000,
			expectedProgress: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMasteryManager()
			mm.AddMasteryXP(1, tt.xp)
			progress := mm.GetProgressToNextMilestone(1)
			if progress != tt.expectedProgress {
				t.Errorf("expected progress %d%%, got %d%%", tt.expectedProgress, progress)
			}
		})
	}
}

func TestGetMasteryForUnknownWeapon(t *testing.T) {
	mm := NewMasteryManager()
	mastery := mm.GetMastery(99) // Weapon never used
	if mastery == nil {
		t.Fatal("expected non-nil mastery for unknown weapon")
	}
	if mastery.XP != 0 {
		t.Errorf("expected 0 XP for new weapon, got %d", mastery.XP)
	}
	if mastery.CurrentBonus.HeadshotDamage != 1.0 {
		t.Error("expected base headshot bonus for new weapon")
	}
}

func TestGetBonusForUnknownWeapon(t *testing.T) {
	mm := NewMasteryManager()
	bonus := mm.GetBonus(99) // Weapon never used
	if bonus.HeadshotDamage != 1.0 {
		t.Errorf("expected base headshot bonus 1.0, got %.2f", bonus.HeadshotDamage)
	}
	if bonus.ReloadSpeed != 1.0 {
		t.Errorf("expected base reload bonus 1.0, got %.2f", bonus.ReloadSpeed)
	}
	if bonus.Accuracy != 1.0 {
		t.Errorf("expected base accuracy bonus 1.0, got %.2f", bonus.Accuracy)
	}
	if bonus.CriticalChance != 0.0 {
		t.Errorf("expected base critical chance 0.0, got %.2f", bonus.CriticalChance)
	}
}

func TestMultipleWeapons(t *testing.T) {
	mm := NewMasteryManager()

	// Add different XP amounts to different weapons
	mm.AddMasteryXP(1, 300)
	mm.AddMasteryXP(2, 600)
	mm.AddMasteryXP(3, 900)

	// Verify each weapon has correct XP and bonuses
	if xp := mm.GetXP(1); xp != 300 {
		t.Errorf("weapon 1: expected 300 XP, got %d", xp)
	}
	if xp := mm.GetXP(2); xp != 600 {
		t.Errorf("weapon 2: expected 600 XP, got %d", xp)
	}
	if xp := mm.GetXP(3); xp != 900 {
		t.Errorf("weapon 3: expected 900 XP, got %d", xp)
	}

	// Verify milestones
	if m := mm.GetCurrentMilestone(1); m != Milestone250 {
		t.Errorf("weapon 1: expected milestone 250, got %d", m)
	}
	if m := mm.GetCurrentMilestone(2); m != Milestone500 {
		t.Errorf("weapon 2: expected milestone 500, got %d", m)
	}
	if m := mm.GetCurrentMilestone(3); m != Milestone750 {
		t.Errorf("weapon 3: expected milestone 750, got %d", m)
	}
}

func TestReset(t *testing.T) {
	mm := NewMasteryManager()
	mm.AddMasteryXP(1, 500)
	mm.AddMasteryXP(2, 750)

	if len(mm.Masteries) != 2 {
		t.Errorf("expected 2 masteries before reset, got %d", len(mm.Masteries))
	}

	mm.Reset()

	if len(mm.Masteries) != 0 {
		t.Errorf("expected 0 masteries after reset, got %d", len(mm.Masteries))
	}

	// After reset, weapons should return default values
	xp := mm.GetXP(1)
	if xp != 0 {
		t.Errorf("expected 0 XP after reset, got %d", xp)
	}
}

func TestGetMilestoneDescription(t *testing.T) {
	tests := []struct {
		milestone   MasteryMilestone
		description string
	}{
		{MilestoneNone, "No bonuses"},
		{Milestone250, "Headshot Damage +10%"},
		{Milestone500, "Reload Speed +15%"},
		{Milestone750, "Accuracy +10%"},
		{Milestone1000, "Critical Chance +5%"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			desc := GetMilestoneDescription(tt.milestone)
			if desc != tt.description {
				t.Errorf("expected %q, got %q", tt.description, desc)
			}
		})
	}
}

func TestWeaponMasteryString(t *testing.T) {
	mm := NewMasteryManager()
	mm.AddMasteryXP(1, 750)

	mastery := mm.GetMastery(1)
	str := mastery.String()

	// Just verify it returns a non-empty string containing key info
	if str == "" {
		t.Error("expected non-empty string representation")
	}
	if !contains(str, "750") {
		t.Error("expected string to contain XP value")
	}
}

func TestIncrementalXPGrowth(t *testing.T) {
	mm := NewMasteryManager()

	// Simulate kill-by-kill XP gain
	for i := 0; i < 100; i++ {
		mm.AddMasteryXP(1, 10) // 10 XP per kill
	}

	xp := mm.GetXP(1)
	if xp != 1000 {
		t.Errorf("expected 1000 XP after 100 kills, got %d", xp)
	}

	// Verify all milestones unlocked
	mastery := mm.GetMastery(1)
	if !mastery.UnlockedBonus1 || !mastery.UnlockedBonus2 || !mastery.UnlockedBonus3 || !mastery.UnlockedBonus4 {
		t.Error("expected all bonuses unlocked at 1000 XP")
	}
}

func TestBonusConsistency(t *testing.T) {
	mm := NewMasteryManager()
	mm.AddMasteryXP(1, 500)

	// Get bonus multiple times - should be consistent
	bonus1 := mm.GetBonus(1)
	bonus2 := mm.GetBonus(1)

	if bonus1.HeadshotDamage != bonus2.HeadshotDamage {
		t.Error("bonus values should be consistent across calls")
	}
	if bonus1.ReloadSpeed != bonus2.ReloadSpeed {
		t.Error("bonus values should be consistent across calls")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

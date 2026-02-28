// Package weapon implements mastery progression for individual weapons.
package weapon

import "fmt"

// MasteryMilestone represents progression milestones.
type MasteryMilestone int

const (
	MilestoneNone MasteryMilestone = iota
	Milestone250                   // Headshot damage +10%
	Milestone500                   // Reload speed +15%
	Milestone750                   // Accuracy +10%
	Milestone1000                  // Critical chance +5%
)

// MasteryBonus represents passive bonuses at each milestone.
type MasteryBonus struct {
	HeadshotDamage float64 // Multiplicative bonus (1.0 = no bonus, 1.1 = +10%)
	ReloadSpeed    float64 // Multiplicative bonus (1.0 = no bonus, 1.15 = +15% faster)
	Accuracy       float64 // Multiplicative bonus (1.0 = no bonus, 1.1 = +10%)
	CriticalChance float64 // Additive percentage (0.0 = 0%, 0.05 = 5%)
}

// WeaponMastery tracks XP and bonuses for a single weapon.
type WeaponMastery struct {
	WeaponSlot     int
	XP             int
	CurrentBonus   MasteryBonus
	UnlockedBonus1 bool // 250 XP milestone
	UnlockedBonus2 bool // 500 XP milestone
	UnlockedBonus3 bool // 750 XP milestone
	UnlockedBonus4 bool // 1000 XP milestone
}

// MasteryManager tracks mastery for all weapons.
type MasteryManager struct {
	Masteries map[int]*WeaponMastery // Weapon slot -> mastery data
}

// NewMasteryManager creates a mastery manager for tracking weapon progression.
func NewMasteryManager() *MasteryManager {
	return &MasteryManager{
		Masteries: make(map[int]*WeaponMastery),
	}
}

// AddMasteryXP grants XP to a weapon and unlocks milestones.
func (mm *MasteryManager) AddMasteryXP(weaponSlot, amount int) {
	mastery, ok := mm.Masteries[weaponSlot]
	if !ok {
		mastery = &WeaponMastery{
			WeaponSlot:   weaponSlot,
			CurrentBonus: MasteryBonus{HeadshotDamage: 1.0, ReloadSpeed: 1.0, Accuracy: 1.0, CriticalChance: 0.0},
		}
		mm.Masteries[weaponSlot] = mastery
	}

	mastery.XP += amount

	// Cap at 1000 XP
	if mastery.XP > 1000 {
		mastery.XP = 1000
	}

	// Unlock bonuses at milestones
	mm.updateBonuses(mastery)
}

// updateBonuses recalculates bonuses based on current XP.
func (mm *MasteryManager) updateBonuses(mastery *WeaponMastery) {
	// Reset to base values
	mastery.CurrentBonus = MasteryBonus{
		HeadshotDamage: 1.0,
		ReloadSpeed:    1.0,
		Accuracy:       1.0,
		CriticalChance: 0.0,
	}

	// Apply bonuses at each milestone
	if mastery.XP >= 250 {
		mastery.UnlockedBonus1 = true
		mastery.CurrentBonus.HeadshotDamage = 1.10 // +10%
	}
	if mastery.XP >= 500 {
		mastery.UnlockedBonus2 = true
		mastery.CurrentBonus.ReloadSpeed = 1.15 // +15%
	}
	if mastery.XP >= 750 {
		mastery.UnlockedBonus3 = true
		mastery.CurrentBonus.Accuracy = 1.10 // +10%
	}
	if mastery.XP >= 1000 {
		mastery.UnlockedBonus4 = true
		mastery.CurrentBonus.CriticalChance = 0.05 // +5%
	}
}

// GetMastery returns mastery data for a weapon slot.
func (mm *MasteryManager) GetMastery(weaponSlot int) *WeaponMastery {
	mastery, ok := mm.Masteries[weaponSlot]
	if !ok {
		return &WeaponMastery{
			WeaponSlot:   weaponSlot,
			CurrentBonus: MasteryBonus{HeadshotDamage: 1.0, ReloadSpeed: 1.0, Accuracy: 1.0, CriticalChance: 0.0},
		}
	}
	return mastery
}

// GetBonus returns the current bonus for a weapon.
func (mm *MasteryManager) GetBonus(weaponSlot int) MasteryBonus {
	mastery := mm.GetMastery(weaponSlot)
	return mastery.CurrentBonus
}

// GetXP returns the current XP for a weapon.
func (mm *MasteryManager) GetXP(weaponSlot int) int {
	mastery := mm.GetMastery(weaponSlot)
	return mastery.XP
}

// GetCurrentMilestone returns the highest milestone unlocked for a weapon.
func (mm *MasteryManager) GetCurrentMilestone(weaponSlot int) MasteryMilestone {
	mastery := mm.GetMastery(weaponSlot)
	if mastery.XP >= 1000 {
		return Milestone1000
	}
	if mastery.XP >= 750 {
		return Milestone750
	}
	if mastery.XP >= 500 {
		return Milestone500
	}
	if mastery.XP >= 250 {
		return Milestone250
	}
	return MilestoneNone
}

// GetProgressToNextMilestone returns XP progress toward the next milestone (0-100).
func (mm *MasteryManager) GetProgressToNextMilestone(weaponSlot int) int {
	mastery := mm.GetMastery(weaponSlot)
	xp := mastery.XP

	// Determine which milestone range we're in
	var current, next int
	if xp < 250 {
		current = 0
		next = 250
	} else if xp < 500 {
		current = 250
		next = 500
	} else if xp < 750 {
		current = 500
		next = 750
	} else if xp < 1000 {
		current = 750
		next = 1000
	} else {
		return 100 // Max level
	}

	// Calculate percentage
	range_ := next - current
	progress := xp - current
	return (progress * 100) / range_
}

// GetMilestoneDescription returns human-readable description of a milestone bonus.
func GetMilestoneDescription(milestone MasteryMilestone) string {
	switch milestone {
	case Milestone250:
		return "Headshot Damage +10%"
	case Milestone500:
		return "Reload Speed +15%"
	case Milestone750:
		return "Accuracy +10%"
	case Milestone1000:
		return "Critical Chance +5%"
	default:
		return "No bonuses"
	}
}

// Reset clears all mastery data (for testing or new game).
func (mm *MasteryManager) Reset() {
	mm.Masteries = make(map[int]*WeaponMastery)
}

// String returns a summary of weapon mastery.
func (wm *WeaponMastery) String() string {
	return fmt.Sprintf("Weapon %d: %d XP (Bonuses: HS%.0f%% Reload%.0f%% Acc%.0f%% Crit%.0f%%)",
		wm.WeaponSlot,
		wm.XP,
		(wm.CurrentBonus.HeadshotDamage-1.0)*100,
		(wm.CurrentBonus.ReloadSpeed-1.0)*100,
		(wm.CurrentBonus.Accuracy-1.0)*100,
		wm.CurrentBonus.CriticalChance*100,
	)
}

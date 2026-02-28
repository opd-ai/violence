package upgrade

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/procgen/genre"
)

func TestNewUpgradeToken(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"zero tokens", 0},
		{"some tokens", 10},
		{"many tokens", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := NewUpgradeToken(tt.count)
			if token == nil {
				t.Fatal("NewUpgradeToken returned nil")
			}
			if token.Count != tt.count {
				t.Errorf("Count = %d, want %d", token.Count, tt.count)
			}
		})
	}
}

func TestUpgradeToken_Add(t *testing.T) {
	token := NewUpgradeToken(5)
	token.Add(3)
	if token.Count != 8 {
		t.Errorf("Count after Add(3) = %d, want 8", token.Count)
	}

	token.Add(0)
	if token.Count != 8 {
		t.Errorf("Count after Add(0) = %d, want 8", token.Count)
	}
}

func TestUpgradeToken_Spend(t *testing.T) {
	tests := []struct {
		name          string
		initialCount  int
		spendAmount   int
		wantSuccess   bool
		wantRemaining int
	}{
		{"sufficient tokens", 10, 5, true, 5},
		{"exact tokens", 10, 10, true, 0},
		{"insufficient tokens", 5, 10, false, 5},
		{"spend zero", 10, 0, true, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := NewUpgradeToken(tt.initialCount)
			success := token.Spend(tt.spendAmount)
			if success != tt.wantSuccess {
				t.Errorf("Spend() = %v, want %v", success, tt.wantSuccess)
			}
			if token.Count != tt.wantRemaining {
				t.Errorf("Count = %d, want %d", token.Count, tt.wantRemaining)
			}
		})
	}
}

func TestUpgradeToken_GetCount(t *testing.T) {
	token := NewUpgradeToken(42)
	if token.GetCount() != 42 {
		t.Errorf("GetCount() = %d, want 42", token.GetCount())
	}
}

func TestNewWeaponUpgrade(t *testing.T) {
	tests := []struct {
		name        string
		upgradeType UpgradeType
	}{
		{"damage upgrade", UpgradeDamage},
		{"fire rate upgrade", UpgradeFireRate},
		{"clip size upgrade", UpgradeClipSize},
		{"accuracy upgrade", UpgradeAccuracy},
		{"range upgrade", UpgradeRange},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upgrade := NewWeaponUpgrade(tt.upgradeType)
			if upgrade == nil {
				t.Fatal("NewWeaponUpgrade returned nil")
			}
			if upgrade.Type != tt.upgradeType {
				t.Errorf("Type = %v, want %v", upgrade.Type, tt.upgradeType)
			}
		})
	}
}

func TestNewWeaponUpgrade_DamageBonus(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeDamage)
	if upgrade.DamageMultiplier <= 1.0 {
		t.Errorf("DamageMultiplier = %f, should be > 1.0", upgrade.DamageMultiplier)
	}
}

func TestNewWeaponUpgrade_FireRateBonus(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeFireRate)
	if upgrade.FireRateModifier >= 1.0 {
		t.Errorf("FireRateModifier = %f, should be < 1.0 (faster)", upgrade.FireRateModifier)
	}
}

func TestNewWeaponUpgrade_ClipSizeBonus(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeClipSize)
	if upgrade.ClipSizeBonus <= 0 {
		t.Errorf("ClipSizeBonus = %d, should be > 0", upgrade.ClipSizeBonus)
	}
}

func TestNewWeaponUpgrade_AccuracyBonus(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeAccuracy)
	if upgrade.AccuracyBonus <= 0 {
		t.Errorf("AccuracyBonus = %f, should be > 0", upgrade.AccuracyBonus)
	}
}

func TestNewWeaponUpgrade_RangeBonus(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeRange)
	if upgrade.RangeBonus <= 0 {
		t.Errorf("RangeBonus = %f, should be > 0", upgrade.RangeBonus)
	}
}

func TestApplyWeaponStats_Damage(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeDamage)
	baseDamage := 10.0

	newDamage, _, _, _, _ := upgrade.ApplyWeaponStats(baseDamage, 1.0, 10, 0.0, 100.0)

	if newDamage <= baseDamage {
		t.Errorf("Upgraded damage = %f, should be > %f", newDamage, baseDamage)
	}
}

func TestApplyWeaponStats_FireRate(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeFireRate)
	baseFireRate := 10.0

	_, newFireRate, _, _, _ := upgrade.ApplyWeaponStats(10.0, baseFireRate, 10, 0.0, 100.0)

	if newFireRate >= baseFireRate {
		t.Errorf("Upgraded fire rate = %f, should be < %f (faster)", newFireRate, baseFireRate)
	}
}

func TestApplyWeaponStats_ClipSize(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeClipSize)
	baseClipSize := 10

	_, _, newClipSize, _, _ := upgrade.ApplyWeaponStats(10.0, 1.0, baseClipSize, 0.0, 100.0)

	if newClipSize <= baseClipSize {
		t.Errorf("Upgraded clip size = %d, should be > %d", newClipSize, baseClipSize)
	}
}

func TestApplyWeaponStats_Accuracy(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeAccuracy)
	baseSpread := 5.0

	_, _, _, newSpread, _ := upgrade.ApplyWeaponStats(10.0, 1.0, 10, baseSpread, 100.0)

	if newSpread >= baseSpread {
		t.Errorf("Upgraded spread = %f, should be < %f (more accurate)", newSpread, baseSpread)
	}
}

func TestApplyWeaponStats_Range(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeRange)
	baseRange := 100.0

	_, _, _, _, newRange := upgrade.ApplyWeaponStats(10.0, 1.0, 10, 0.0, baseRange)

	if newRange <= baseRange {
		t.Errorf("Upgraded range = %f, should be > %f", newRange, baseRange)
	}
}

func TestApplyWeaponStats_MultipleStats(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeDamage)

	damage, fireRate, clipSize, spread, weaponRange := upgrade.ApplyWeaponStats(10.0, 5.0, 12, 2.5, 80.0)

	// Damage upgrade should only affect damage
	if damage == 10.0 {
		t.Error("Damage should be modified")
	}
	if fireRate != 5.0 {
		t.Error("Fire rate should be unchanged")
	}
	if clipSize != 12 {
		t.Error("Clip size should be unchanged")
	}
	if spread != 2.5 {
		t.Error("Spread should be unchanged")
	}
	if weaponRange != 80.0 {
		t.Error("Range should be unchanged")
	}
}

func TestSetGenre(t *testing.T) {
	genres := []string{
		genre.Fantasy,
		genre.SciFi,
		genre.Cyberpunk,
		genre.Horror,
		genre.PostApoc,
	}

	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			upgrade := NewWeaponUpgrade(UpgradeDamage)
			upgrade.SetGenre(g)

			name := upgrade.GetGenreName()
			if name == "" {
				t.Error("Genre name should not be empty")
			}
			if name == "Unknown Upgrade" {
				t.Errorf("Genre name should be set for genre %s, got fallback", g)
			}
		})
	}
}

func TestGetGenreName_AllUpgradeTypes(t *testing.T) {
	upgradeTypes := []UpgradeType{
		UpgradeDamage,
		UpgradeFireRate,
		UpgradeClipSize,
		UpgradeAccuracy,
		UpgradeRange,
	}

	for _, upType := range upgradeTypes {
		t.Run("", func(t *testing.T) {
			upgrade := NewWeaponUpgrade(upType)
			upgrade.SetGenre(genre.Fantasy)

			name := upgrade.GetGenreName()
			if name == "" {
				t.Errorf("Genre name should not be empty for upgrade type %d", upType)
			}
		})
	}
}

func TestGetGenreName_Default(t *testing.T) {
	upgrade := NewWeaponUpgrade(UpgradeDamage)
	// Don't set genre

	name := upgrade.GetGenreName()
	if name == "" {
		t.Error("Should return default genre name when not set")
	}
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.weaponUpgrades == nil {
		t.Error("weaponUpgrades should be initialized")
	}
	if m.tokens == nil {
		t.Error("tokens should be initialized")
	}
	if m.tokens.GetCount() != 0 {
		t.Errorf("Initial token count = %d, want 0", m.tokens.GetCount())
	}
}

func TestManager_ApplyUpgrade(t *testing.T) {
	m := NewManager()
	m.tokens.Add(10)

	success := m.ApplyUpgrade("weapon1", UpgradeDamage, 5)
	if !success {
		t.Error("ApplyUpgrade should succeed with sufficient tokens")
	}
	if m.tokens.GetCount() != 5 {
		t.Errorf("Tokens after upgrade = %d, want 5", m.tokens.GetCount())
	}

	upgrades := m.GetUpgrades("weapon1")
	if len(upgrades) != 1 {
		t.Errorf("Upgrade count = %d, want 1", len(upgrades))
	}
	if upgrades[0] != UpgradeDamage {
		t.Errorf("Upgrade type = %v, want %v", upgrades[0], UpgradeDamage)
	}
}

func TestManager_ApplyUpgrade_InsufficientTokens(t *testing.T) {
	m := NewManager()
	m.tokens.Add(3)

	success := m.ApplyUpgrade("weapon1", UpgradeDamage, 5)
	if success {
		t.Error("ApplyUpgrade should fail with insufficient tokens")
	}
	if m.tokens.GetCount() != 3 {
		t.Errorf("Tokens should be unchanged = %d, want 3", m.tokens.GetCount())
	}

	upgrades := m.GetUpgrades("weapon1")
	if len(upgrades) != 0 {
		t.Errorf("No upgrades should be applied, got %d", len(upgrades))
	}
}

func TestManager_MultipleUpgrades(t *testing.T) {
	m := NewManager()
	m.tokens.Add(20)

	m.ApplyUpgrade("weapon1", UpgradeDamage, 5)
	m.ApplyUpgrade("weapon1", UpgradeFireRate, 5)
	m.ApplyUpgrade("weapon1", UpgradeClipSize, 5)

	upgrades := m.GetUpgrades("weapon1")
	if len(upgrades) != 3 {
		t.Errorf("Upgrade count = %d, want 3", len(upgrades))
	}
}

func TestManager_MultipleWeapons(t *testing.T) {
	m := NewManager()
	m.tokens.Add(30)

	m.ApplyUpgrade("weapon1", UpgradeDamage, 5)
	m.ApplyUpgrade("weapon2", UpgradeFireRate, 5)
	m.ApplyUpgrade("weapon3", UpgradeRange, 5)

	if len(m.GetUpgrades("weapon1")) != 1 {
		t.Error("weapon1 should have 1 upgrade")
	}
	if len(m.GetUpgrades("weapon2")) != 1 {
		t.Error("weapon2 should have 1 upgrade")
	}
	if len(m.GetUpgrades("weapon3")) != 1 {
		t.Error("weapon3 should have 1 upgrade")
	}
}

func TestManager_GetUpgrades_NoUpgrades(t *testing.T) {
	m := NewManager()

	upgrades := m.GetUpgrades("nonexistent")
	if upgrades != nil {
		t.Error("Should return nil for weapon with no upgrades")
	}
}

func TestManager_GetTokens(t *testing.T) {
	m := NewManager()
	m.tokens.Add(42)

	tokens := m.GetTokens()
	if tokens.GetCount() != 42 {
		t.Errorf("GetTokens().GetCount() = %d, want 42", tokens.GetCount())
	}
}

func TestManager_HasUpgrade(t *testing.T) {
	m := NewManager()
	m.tokens.Add(10)

	if m.HasUpgrade("weapon1", UpgradeDamage) {
		t.Error("Should not have upgrade before applying")
	}

	m.ApplyUpgrade("weapon1", UpgradeDamage, 5)

	if !m.HasUpgrade("weapon1", UpgradeDamage) {
		t.Error("Should have upgrade after applying")
	}
	if m.HasUpgrade("weapon1", UpgradeFireRate) {
		t.Error("Should not have unapplied upgrade")
	}
}

func TestApplyWeaponStats_NoChange(t *testing.T) {
	// Create upgrade with all bonuses at neutral values
	upgrade := &WeaponUpgrade{
		Type:             UpgradeDamage,
		DamageMultiplier: 1.0,
		FireRateModifier: 1.0,
		ClipSizeBonus:    0,
		AccuracyBonus:    0.0,
		RangeBonus:       0.0,
	}

	damage, fireRate, clipSize, spread, weaponRange := upgrade.ApplyWeaponStats(10.0, 5.0, 12, 2.5, 80.0)

	if math.Abs(damage-10.0) > 0.01 {
		t.Errorf("Damage = %f, want 10.0", damage)
	}
	if math.Abs(fireRate-5.0) > 0.01 {
		t.Errorf("FireRate = %f, want 5.0", fireRate)
	}
	if clipSize != 12 {
		t.Errorf("ClipSize = %d, want 12", clipSize)
	}
	if math.Abs(spread-2.5) > 0.01 {
		t.Errorf("Spread = %f, want 2.5", spread)
	}
	if math.Abs(weaponRange-80.0) > 0.01 {
		t.Errorf("Range = %f, want 80.0", weaponRange)
	}
}

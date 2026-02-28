// Package upgrade implements weapon upgrade mechanics and upgrade tokens.
package upgrade

import (
	"github.com/opd-ai/violence/pkg/procgen/genre"
)

// UpgradeType defines the type of weapon upgrade.
type UpgradeType int

const (
	UpgradeDamage UpgradeType = iota
	UpgradeFireRate
	UpgradeClipSize
	UpgradeAccuracy
	UpgradeRange
)

// UpgradeToken represents a collectible currency for upgrades.
type UpgradeToken struct {
	Count int
}

// WeaponUpgrade holds stat modifiers for a weapon upgrade.
type WeaponUpgrade struct {
	Type             UpgradeType
	DamageMultiplier float64 // 1.0 = no change, 1.2 = +20%
	FireRateModifier float64 // 0.8 = 20% faster (lower frames between shots)
	ClipSizeBonus    int
	AccuracyBonus    float64 // Reduces spread angle
	RangeBonus       float64
	genreName        string // Genre-specific display name
}

// NewUpgradeToken creates a token pool with initial count.
func NewUpgradeToken(count int) *UpgradeToken {
	return &UpgradeToken{Count: count}
}

// Add increases the token count.
func (ut *UpgradeToken) Add(amount int) {
	ut.Count += amount
}

// Spend decreases the token count. Returns false if insufficient tokens.
func (ut *UpgradeToken) Spend(amount int) bool {
	if ut.Count < amount {
		return false
	}
	ut.Count -= amount
	return true
}

// GetCount returns the current token count.
func (ut *UpgradeToken) GetCount() int {
	return ut.Count
}

// NewWeaponUpgrade creates a weapon upgrade of the given type.
func NewWeaponUpgrade(upgradeType UpgradeType) *WeaponUpgrade {
	upgrade := &WeaponUpgrade{
		Type:             upgradeType,
		DamageMultiplier: 1.0,
		FireRateModifier: 1.0,
		ClipSizeBonus:    0,
		AccuracyBonus:    0.0,
		RangeBonus:       0.0,
	}

	// Set default bonuses per upgrade type
	switch upgradeType {
	case UpgradeDamage:
		upgrade.DamageMultiplier = 1.25 // +25% damage
	case UpgradeFireRate:
		upgrade.FireRateModifier = 0.85 // 15% faster (fewer frames between shots)
	case UpgradeClipSize:
		upgrade.ClipSizeBonus = 5 // +5 rounds
	case UpgradeAccuracy:
		upgrade.AccuracyBonus = 0.2 // -20% spread
	case UpgradeRange:
		upgrade.RangeBonus = 15.0 // +15 units
	}

	return upgrade
}

// ApplyWeaponStats modifies weapon stats based on the upgrade.
// Takes current stats and returns modified stats.
func (wu *WeaponUpgrade) ApplyWeaponStats(damage, fireRate float64, clipSize int, spreadAngle, weaponRange float64) (float64, float64, int, float64, float64) {
	newDamage := damage * wu.DamageMultiplier
	newFireRate := fireRate * wu.FireRateModifier
	newClipSize := clipSize + wu.ClipSizeBonus
	newSpreadAngle := spreadAngle * (1.0 - wu.AccuracyBonus)
	newRange := weaponRange + wu.RangeBonus

	return newDamage, newFireRate, newClipSize, newSpreadAngle, newRange
}

// SetGenre sets genre-specific upgrade names.
func (wu *WeaponUpgrade) SetGenre(genreID string) {
	wu.genreName = getGenreUpgradeName(genreID, wu.Type)
}

// GetGenreName returns the genre-specific upgrade name.
func (wu *WeaponUpgrade) GetGenreName() string {
	if wu.genreName == "" {
		return getGenreUpgradeName(genre.Fantasy, wu.Type) // Default to fantasy
	}
	return wu.genreName
}

// getGenreUpgradeName returns genre-specific upgrade names.
func getGenreUpgradeName(genreID string, upgradeType UpgradeType) string {
	names := map[string]map[UpgradeType]string{
		genre.Fantasy: {
			UpgradeDamage:   "Enchantment of Power",
			UpgradeFireRate: "Enchantment of Haste",
			UpgradeClipSize: "Enchantment of Capacity",
			UpgradeAccuracy: "Enchantment of Precision",
			UpgradeRange:    "Enchantment of Reach",
		},
		genre.SciFi: {
			UpgradeDamage:   "Damage Calibration",
			UpgradeFireRate: "Rate-of-Fire Calibration",
			UpgradeClipSize: "Magazine Calibration",
			UpgradeAccuracy: "Targeting Calibration",
			UpgradeRange:    "Range Calibration",
		},
		genre.Cyberpunk: {
			UpgradeDamage:   "Damage Augmentation",
			UpgradeFireRate: "Fire-Rate Augmentation",
			UpgradeClipSize: "Capacity Augmentation",
			UpgradeAccuracy: "Accuracy Augmentation",
			UpgradeRange:    "Range Augmentation",
		},
		genre.Horror: {
			UpgradeDamage:   "Damage Modification",
			UpgradeFireRate: "Rate Modification",
			UpgradeClipSize: "Capacity Modification",
			UpgradeAccuracy: "Aim Modification",
			UpgradeRange:    "Range Modification",
		},
		genre.PostApoc: {
			UpgradeDamage:   "Damage Retrofit",
			UpgradeFireRate: "Fire-Rate Retrofit",
			UpgradeClipSize: "Magazine Retrofit",
			UpgradeAccuracy: "Accuracy Retrofit",
			UpgradeRange:    "Range Retrofit",
		},
	}

	if genreNames, ok := names[genreID]; ok {
		if name, ok := genreNames[upgradeType]; ok {
			return name
		}
	}

	// Fallback to generic names
	switch upgradeType {
	case UpgradeDamage:
		return "Damage Upgrade"
	case UpgradeFireRate:
		return "Fire Rate Upgrade"
	case UpgradeClipSize:
		return "Clip Size Upgrade"
	case UpgradeAccuracy:
		return "Accuracy Upgrade"
	case UpgradeRange:
		return "Range Upgrade"
	}
	return "Unknown Upgrade"
}

// Manager tracks applied upgrades per weapon.
type Manager struct {
	weaponUpgrades map[string][]UpgradeType // weapon ID -> list of applied upgrades
	tokens         *UpgradeToken
}

// NewManager creates an upgrade manager.
func NewManager() *Manager {
	return &Manager{
		weaponUpgrades: make(map[string][]UpgradeType),
		tokens:         NewUpgradeToken(0),
	}
}

// ApplyUpgrade applies an upgrade to a weapon.
func (m *Manager) ApplyUpgrade(weaponID string, upgradeType UpgradeType, cost int) bool {
	if !m.tokens.Spend(cost) {
		return false
	}

	if m.weaponUpgrades[weaponID] == nil {
		m.weaponUpgrades[weaponID] = []UpgradeType{}
	}
	m.weaponUpgrades[weaponID] = append(m.weaponUpgrades[weaponID], upgradeType)
	return true
}

// GetUpgrades returns all upgrades applied to a weapon.
func (m *Manager) GetUpgrades(weaponID string) []UpgradeType {
	return m.weaponUpgrades[weaponID]
}

// GetTokens returns the upgrade token pool.
func (m *Manager) GetTokens() *UpgradeToken {
	return m.tokens
}

// HasUpgrade checks if a weapon has a specific upgrade.
func (m *Manager) HasUpgrade(weaponID string, upgradeType UpgradeType) bool {
	upgrades := m.weaponUpgrades[weaponID]
	for _, u := range upgrades {
		if u == upgradeType {
			return true
		}
	}
	return false
}

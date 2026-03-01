// Package loot handles item drops and loot tables.
package loot

import (
	"github.com/opd-ai/violence/pkg/rng"
)

// Rarity defines item rarity tiers for secret rewards.
type Rarity int

const (
	RarityCommon Rarity = iota
	RarityUncommon
	RarityRare
	RarityLegendary
)

// Drop represents a single dropped item.
type Drop struct {
	ItemID string
	Chance float64
}

// LootTable defines a set of possible drops.
type LootTable struct {
	Drops []Drop
	rng   *rng.RNG
}

// SecretLootTable holds weighted rare items for secret rewards.
type SecretLootTable struct {
	Uncommon  []string // 30% chance
	Rare      []string // 50% chance
	Legendary []string // 20% chance
	rng       *rng.RNG
}

// NewLootTable creates an empty loot table.
func NewLootTable() *LootTable {
	return &LootTable{Drops: []Drop{}}
}

// NewLootTableWithRNG creates a loot table with deterministic RNG.
func NewLootTableWithRNG(r *rng.RNG) *LootTable {
	return &LootTable{
		Drops: []Drop{},
		rng:   r,
	}
}

// Roll selects drops from the loot table using deterministic RNG.
// Each drop has an independent chance to be included in the result.
// Returns a slice of item IDs that were successfully rolled.
func (lt *LootTable) Roll() []string {
	if len(lt.Drops) == 0 {
		return nil
	}

	// Use internal RNG if available, otherwise create one
	var localRNG *rng.RNG
	if lt.rng != nil {
		localRNG = lt.rng
	} else {
		localRNG = rng.NewRNG(0)
	}

	var result []string
	for _, drop := range lt.Drops {
		// Roll independently for each drop
		roll := localRNG.Float64()
		if roll < drop.Chance {
			result = append(result, drop.ItemID)
		}
	}

	return result
}

// RollWithSeed selects drops using a specific seed for deterministic results.
func (lt *LootTable) RollWithSeed(seed uint64) []string {
	if len(lt.Drops) == 0 {
		return nil
	}

	localRNG := rng.NewRNG(seed)

	var result []string
	for _, drop := range lt.Drops {
		roll := localRNG.Float64()
		if roll < drop.Chance {
			result = append(result, drop.ItemID)
		}
	}

	return result
}

// NewSecretLootTable creates a secret loot table with default rare items.
func NewSecretLootTable(r *rng.RNG) *SecretLootTable {
	return &SecretLootTable{
		Uncommon: []string{
			"ammo_bulk_bullets",
			"ammo_bulk_shells",
			"health_large",
			"armor_medium",
		},
		Rare: []string{
			"weapon_plasma",
			"weapon_super_shotgun",
			"ammo_rockets_pack",
			"armor_heavy",
			"health_mega",
		},
		Legendary: []string{
			"weapon_bfg",
			"weapon_railgun",
			"powerup_invulnerability",
			"powerup_quad_damage",
		},
		rng: r,
	}
}

// GenerateSecretReward returns a deterministic rare item based on seed.
// Returns the item ID and its rarity tier.
func (slt *SecretLootTable) GenerateSecretReward(seed uint64) (string, Rarity) {
	// Create deterministic RNG from seed
	localRNG := rng.NewRNG(seed)

	// Roll for rarity tier: 30% uncommon, 50% rare, 20% legendary
	roll := localRNG.Intn(100)

	if roll < 30 {
		// Uncommon (30%)
		if len(slt.Uncommon) == 0 {
			return "ammo_bullets", RarityCommon
		}
		idx := localRNG.Intn(len(slt.Uncommon))
		return slt.Uncommon[idx], RarityUncommon
	} else if roll < 80 {
		// Rare (50%)
		if len(slt.Rare) == 0 {
			return "health_large", RarityUncommon
		}
		idx := localRNG.Intn(len(slt.Rare))
		return slt.Rare[idx], RarityRare
	} else {
		// Legendary (20%)
		if len(slt.Legendary) == 0 {
			return "weapon_plasma", RarityRare
		}
		idx := localRNG.Intn(len(slt.Legendary))
		return slt.Legendary[idx], RarityLegendary
	}
}

// AddItem adds an item to a specific rarity tier.
func (slt *SecretLootTable) AddItem(item string, rarity Rarity) {
	switch rarity {
	case RarityUncommon:
		slt.Uncommon = append(slt.Uncommon, item)
	case RarityRare:
		slt.Rare = append(slt.Rare, item)
	case RarityLegendary:
		slt.Legendary = append(slt.Legendary, item)
	}
}

var currentGenre = "fantasy"

// SetGenre configures loot tables for a genre.
func SetGenre(genreID string) {
	currentGenre = genreID
}

// GetCurrentGenre returns the current global genre setting.
func GetCurrentGenre() string {
	return currentGenre
}

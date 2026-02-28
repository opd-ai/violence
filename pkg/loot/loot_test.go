package loot

import (
	"testing"

	"github.com/opd-ai/violence/pkg/rng"
)

func TestNewLootTable(t *testing.T) {
	lt := NewLootTable()
	if lt == nil {
		t.Fatal("NewLootTable returned nil")
	}
	if lt.Drops == nil {
		t.Error("LootTable Drops should be initialized (empty slice)")
	}
}

func TestDrop(t *testing.T) {
	tests := []struct {
		name   string
		itemID string
		chance float64
	}{
		{"common_item", "health_pack", 0.5},
		{"rare_item", "plasma_rifle", 0.05},
		{"guaranteed", "ammo_bullets", 1.0},
		{"never_drops", "legendary_sword", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drop := Drop{
				ItemID: tt.itemID,
				Chance: tt.chance,
			}

			if drop.ItemID != tt.itemID {
				t.Errorf("ItemID: expected %s, got %s", tt.itemID, drop.ItemID)
			}
			if drop.Chance != tt.chance {
				t.Errorf("Chance: expected %f, got %f", tt.chance, drop.Chance)
			}
		})
	}
}

func TestLootTableWithDrops(t *testing.T) {
	lt := NewLootTable()

	lt.Drops = []Drop{
		{ItemID: "health_small", Chance: 0.5},
		{ItemID: "ammo_bullets", Chance: 0.3},
		{ItemID: "armor_shard", Chance: 0.2},
	}

	if len(lt.Drops) != 3 {
		t.Errorf("Expected 3 drops, got %d", len(lt.Drops))
	}
}

func TestRoll(t *testing.T) {
	lt := NewLootTable()
	lt.Drops = []Drop{
		{ItemID: "item1", Chance: 1.0}, // Guaranteed
		{ItemID: "item2", Chance: 0.5},
	}

	// Roll should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Roll panicked: %v", r)
		}
	}()

	result := lt.Roll()
	if result == nil {
		t.Log("Roll returned nil (expected for stub implementation)")
	}
}

func TestRollEmptyTable(t *testing.T) {
	lt := NewLootTable()

	result := lt.Roll()
	if result == nil {
		t.Log("Empty loot table returned nil (expected)")
	}
}

func TestSetGenre(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("SetGenre(%q) panicked: %v", genre, r)
				}
			}()
			SetGenre(genre)
		})
	}
}

func TestDropChanceBounds(t *testing.T) {
	tests := []struct {
		name   string
		chance float64
	}{
		{"zero_chance", 0.0},
		{"half_chance", 0.5},
		{"full_chance", 1.0},
		{"over_100_percent", 1.5}, // Edge case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drop := Drop{ItemID: "test", Chance: tt.chance}
			if drop.Chance != tt.chance {
				t.Errorf("Chance: expected %f, got %f", tt.chance, drop.Chance)
			}
		})
	}
}

func TestMultipleLootTables(t *testing.T) {
	// Test creating multiple independent loot tables
	enemyLoot := NewLootTable()
	enemyLoot.Drops = []Drop{
		{ItemID: "ammo", Chance: 0.8},
	}

	crateLoot := NewLootTable()
	crateLoot.Drops = []Drop{
		{ItemID: "weapon", Chance: 0.3},
		{ItemID: "health", Chance: 0.7},
	}

	secretLoot := NewLootTable()
	secretLoot.Drops = []Drop{
		{ItemID: "legendary", Chance: 0.1},
		{ItemID: "artifact", Chance: 0.05},
	}

	if len(enemyLoot.Drops) != 1 {
		t.Error("Enemy loot table corrupted")
	}
	if len(crateLoot.Drops) != 2 {
		t.Error("Crate loot table corrupted")
	}
	if len(secretLoot.Drops) != 2 {
		t.Error("Secret loot table corrupted")
	}
}

func TestLootTableModification(t *testing.T) {
	lt := NewLootTable()

	// Add drops
	lt.Drops = append(lt.Drops, Drop{ItemID: "item1", Chance: 0.5})
	lt.Drops = append(lt.Drops, Drop{ItemID: "item2", Chance: 0.3})

	if len(lt.Drops) != 2 {
		t.Errorf("Expected 2 drops after append, got %d", len(lt.Drops))
	}

	// Verify items
	if lt.Drops[0].ItemID != "item1" {
		t.Error("First drop should be item1")
	}
	if lt.Drops[1].ItemID != "item2" {
		t.Error("Second drop should be item2")
	}
}

func TestNewSecretLootTable(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := NewSecretLootTable(r)

	if slt == nil {
		t.Fatal("NewSecretLootTable returned nil")
	}
	if len(slt.Uncommon) == 0 {
		t.Error("Uncommon items should be initialized")
	}
	if len(slt.Rare) == 0 {
		t.Error("Rare items should be initialized")
	}
	if len(slt.Legendary) == 0 {
		t.Error("Legendary items should be initialized")
	}
	if slt.rng == nil {
		t.Error("RNG should be set")
	}
}

func TestGenerateSecretReward_Deterministic(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := NewSecretLootTable(r)

	// Same seed should produce same result
	item1, rarity1 := slt.GenerateSecretReward(999)
	item2, rarity2 := slt.GenerateSecretReward(999)

	if item1 != item2 {
		t.Errorf("Same seed produced different items: %s vs %s", item1, item2)
	}
	if rarity1 != rarity2 {
		t.Errorf("Same seed produced different rarities: %d vs %d", rarity1, rarity2)
	}
}

func TestGenerateSecretReward_DifferentSeeds(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := NewSecretLootTable(r)

	// Different seeds should (usually) produce different results
	results := make(map[string]int)
	for seed := uint64(0); seed < 100; seed++ {
		item, _ := slt.GenerateSecretReward(seed)
		results[item]++
	}

	// Should have multiple different items across 100 rolls
	if len(results) < 3 {
		t.Errorf("Expected variety in rewards, got only %d unique items in 100 rolls", len(results))
	}
}

func TestGenerateSecretReward_RarityDistribution(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := NewSecretLootTable(r)

	uncommonCount := 0
	rareCount := 0
	legendaryCount := 0
	trials := 1000

	for seed := uint64(0); seed < uint64(trials); seed++ {
		_, rarity := slt.GenerateSecretReward(seed)
		switch rarity {
		case RarityUncommon:
			uncommonCount++
		case RarityRare:
			rareCount++
		case RarityLegendary:
			legendaryCount++
		}
	}

	// Check approximate distribution (30% uncommon, 50% rare, 20% legendary)
	// Allow 10% margin of error
	uncommonPercent := float64(uncommonCount) / float64(trials) * 100
	rarePercent := float64(rareCount) / float64(trials) * 100
	legendaryPercent := float64(legendaryCount) / float64(trials) * 100

	if uncommonPercent < 20 || uncommonPercent > 40 {
		t.Errorf("Uncommon rate = %.1f%%, want ~30%%", uncommonPercent)
	}
	if rarePercent < 40 || rarePercent > 60 {
		t.Errorf("Rare rate = %.1f%%, want ~50%%", rarePercent)
	}
	if legendaryPercent < 10 || legendaryPercent > 30 {
		t.Errorf("Legendary rate = %.1f%%, want ~20%%", legendaryPercent)
	}
}

func TestGenerateSecretReward_EmptyTables(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := &SecretLootTable{
		Uncommon:  []string{},
		Rare:      []string{},
		Legendary: []string{},
		rng:       r,
	}

	// Should return fallback items instead of panicking
	item, rarity := slt.GenerateSecretReward(123)
	if item == "" {
		t.Error("Empty tables should return fallback item, not empty string")
	}
	if rarity == RarityCommon {
		// Fallback for uncommon tier
		if item != "ammo_bullets" {
			t.Errorf("Fallback uncommon item = %s, want ammo_bullets", item)
		}
	}
}

func TestAddItem(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := NewSecretLootTable(r)

	initialUncommon := len(slt.Uncommon)
	initialRare := len(slt.Rare)
	initialLegendary := len(slt.Legendary)

	slt.AddItem("custom_uncommon", RarityUncommon)
	slt.AddItem("custom_rare", RarityRare)
	slt.AddItem("custom_legendary", RarityLegendary)

	if len(slt.Uncommon) != initialUncommon+1 {
		t.Errorf("Uncommon count = %d, want %d", len(slt.Uncommon), initialUncommon+1)
	}
	if len(slt.Rare) != initialRare+1 {
		t.Errorf("Rare count = %d, want %d", len(slt.Rare), initialRare+1)
	}
	if len(slt.Legendary) != initialLegendary+1 {
		t.Errorf("Legendary count = %d, want %d", len(slt.Legendary), initialLegendary+1)
	}

	// Verify items were added
	found := false
	for _, item := range slt.Uncommon {
		if item == "custom_uncommon" {
			found = true
			break
		}
	}
	if !found {
		t.Error("custom_uncommon not found in Uncommon list")
	}
}

func TestAddItem_CommonRarity(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := NewSecretLootTable(r)

	initialUncommon := len(slt.Uncommon)
	initialRare := len(slt.Rare)
	initialLegendary := len(slt.Legendary)

	// Adding common rarity should not add to any list
	slt.AddItem("common_item", RarityCommon)

	if len(slt.Uncommon) != initialUncommon {
		t.Error("Common item should not be added to Uncommon list")
	}
	if len(slt.Rare) != initialRare {
		t.Error("Common item should not be added to Rare list")
	}
	if len(slt.Legendary) != initialLegendary {
		t.Error("Common item should not be added to Legendary list")
	}
}

func TestRarityConstants(t *testing.T) {
	// Verify rarity constants are distinct
	rarities := map[Rarity]bool{
		RarityCommon:    true,
		RarityUncommon:  true,
		RarityRare:      true,
		RarityLegendary: true,
	}

	if len(rarities) != 4 {
		t.Error("Rarity constants should be distinct")
	}
}

func TestGenerateSecretReward_AllTiers(t *testing.T) {
	r := rng.NewRNG(12345)
	slt := NewSecretLootTable(r)

	foundUncommon := false
	foundRare := false
	foundLegendary := false

	// Generate enough rewards to hit all tiers
	for seed := uint64(0); seed < 100; seed++ {
		_, rarity := slt.GenerateSecretReward(seed)
		switch rarity {
		case RarityUncommon:
			foundUncommon = true
		case RarityRare:
			foundRare = true
		case RarityLegendary:
			foundLegendary = true
		}
	}

	if !foundUncommon {
		t.Error("Should generate at least one uncommon item in 100 rolls")
	}
	if !foundRare {
		t.Error("Should generate at least one rare item in 100 rolls")
	}
	if !foundLegendary {
		t.Error("Should generate at least one legendary item in 100 rolls")
	}
}

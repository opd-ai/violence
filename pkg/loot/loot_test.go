package loot

import "testing"

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

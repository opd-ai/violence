package loot

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
)

func TestNewLootDropSystem(t *testing.T) {
	sys := NewLootDropSystem(12345)
	if sys == nil {
		t.Fatal("NewLootDropSystem returned nil")
	}
	if sys.rng == nil {
		t.Error("System RNG should be initialized")
	}
	if sys.deathProcessed == nil {
		t.Error("Death tracking map should be initialized")
	}
}

func TestLootDropSystem_SetGenre(t *testing.T) {
	sys := NewLootDropSystem(12345)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys.SetGenre(genre)
			if sys.currentGenre != genre {
				t.Errorf("Genre = %s, want %s", sys.currentGenre, genre)
			}
		})
	}
}

func TestLootDropSystem_NoLootOnAlive(t *testing.T) {
	sys := NewLootDropSystem(12345)
	w := engine.NewWorld()

	// Create entity with loot drop but alive
	entity := w.AddEntity()
	table := NewLootTableWithRNG(rng.NewRNG(100))
	table.Drops = []Drop{{ItemID: "test_item", Chance: 1.0}}

	w.AddComponent(entity, &LootDropComponent{
		Table:      table,
		DropSeed:   100,
		DropChance: 1.0,
	})
	w.AddComponent(entity, &HealthComponent{Current: 50, Max: 100})
	w.AddComponent(entity, &PositionComponent{X: 10, Y: 10})

	// Run system
	sys.Update(w)

	// No loot items should be spawned
	entities := w.Query()
	lootCount := 0
	for _, e := range entities {
		if e != entity {
			lootCount++
		}
	}

	// Should only have the original entity
	if lootCount > 0 {
		t.Errorf("Loot spawned for alive entity, got %d extra entities", lootCount)
	}
}

func TestLootDropSystem_SpawnsLootOnDeath(t *testing.T) {
	sys := NewLootDropSystem(12345)
	w := engine.NewWorld()

	// Create entity that will die
	entity := w.AddEntity()
	table := NewLootTableWithRNG(rng.NewRNG(100))
	table.Drops = []Drop{
		{ItemID: "health_pack", Chance: 1.0},
		{ItemID: "ammo", Chance: 1.0},
	}

	w.AddComponent(entity, &LootDropComponent{
		Table:      table,
		DropSeed:   100,
		DropChance: 1.0,
	})
	w.AddComponent(entity, &HealthComponent{Current: 0, Max: 100}) // Dead
	w.AddComponent(entity, &PositionComponent{X: 10, Y: 10})

	// Track spawned loot
	spawnedItems := []string{}
	sys.SetLootSpawnCallback(func(itemID string, x, y float64, rarity Rarity) {
		spawnedItems = append(spawnedItems, itemID)
	})

	// Run system
	sys.Update(w)

	// Verify loot was spawned
	if len(spawnedItems) != 2 {
		t.Errorf("Expected 2 loot items spawned, got %d", len(spawnedItems))
	}
}

func TestLootDropSystem_DropChance(t *testing.T) {
	// Test with very low drop chance - should not drop
	sys := NewLootDropSystem(12345)
	w := engine.NewWorld()

	entity := w.AddEntity()
	table := NewLootTableWithRNG(rng.NewRNG(100))
	table.Drops = []Drop{{ItemID: "rare_item", Chance: 1.0}}

	w.AddComponent(entity, &LootDropComponent{
		Table:      table,
		DropSeed:   100,
		DropChance: 0.001, // Very low overall drop chance
	})
	w.AddComponent(entity, &HealthComponent{Current: 0, Max: 100})
	w.AddComponent(entity, &PositionComponent{X: 10, Y: 10})

	spawnedItems := []string{}
	sys.SetLootSpawnCallback(func(itemID string, x, y float64, rarity Rarity) {
		spawnedItems = append(spawnedItems, itemID)
	})

	// Run multiple times - with 0.1% chance, should almost never spawn
	successCount := 0
	for i := 0; i < 10; i++ {
		sys.deathProcessed = make(map[engine.Entity]bool) // Reset
		sys.Update(w)
		if len(spawnedItems) > 0 {
			successCount++
		}
		spawnedItems = []string{}
	}

	// With 0.1% chance across 10 trials, expect 0 or maybe 1 success
	if successCount > 2 {
		t.Errorf("Drop chance too high: %d/10 succeeded with 0.1%% chance", successCount)
	}
}

func TestLootDropSystem_DeterministicDrops(t *testing.T) {
	// Same seed should produce same drops
	w1 := engine.NewWorld()
	sys1 := NewLootDropSystem(12345)

	entity1 := w1.AddEntity()
	table1 := NewLootTableWithRNG(rng.NewRNG(100))
	table1.Drops = []Drop{
		{ItemID: "item1", Chance: 0.5},
		{ItemID: "item2", Chance: 0.5},
		{ItemID: "item3", Chance: 0.5},
	}

	w1.AddComponent(entity1, &LootDropComponent{
		Table:      table1,
		DropSeed:   999,
		DropChance: 1.0,
	})
	w1.AddComponent(entity1, &HealthComponent{Current: 0, Max: 100})
	w1.AddComponent(entity1, &PositionComponent{X: 10, Y: 10})

	var spawned1 []string
	sys1.SetLootSpawnCallback(func(itemID string, x, y float64, rarity Rarity) {
		spawned1 = append(spawned1, itemID)
	})
	sys1.Update(w1)

	// Second world with same seed
	w2 := engine.NewWorld()
	sys2 := NewLootDropSystem(12345)

	entity2 := w2.AddEntity()
	table2 := NewLootTableWithRNG(rng.NewRNG(100))
	table2.Drops = []Drop{
		{ItemID: "item1", Chance: 0.5},
		{ItemID: "item2", Chance: 0.5},
		{ItemID: "item3", Chance: 0.5},
	}

	w2.AddComponent(entity2, &LootDropComponent{
		Table:      table2,
		DropSeed:   999,
		DropChance: 1.0,
	})
	w2.AddComponent(entity2, &HealthComponent{Current: 0, Max: 100})
	w2.AddComponent(entity2, &PositionComponent{X: 10, Y: 10})

	var spawned2 []string
	sys2.SetLootSpawnCallback(func(itemID string, x, y float64, rarity Rarity) {
		spawned2 = append(spawned2, itemID)
	})
	sys2.Update(w2)

	// Should produce identical results
	if len(spawned1) != len(spawned2) {
		t.Errorf("Determinism failed: got %d and %d items", len(spawned1), len(spawned2))
	}

	for i := range spawned1 {
		if i >= len(spawned2) {
			break
		}
		if spawned1[i] != spawned2[i] {
			t.Errorf("Item %d mismatch: %s vs %s", i, spawned1[i], spawned2[i])
		}
	}
}

func TestLootDropSystem_NoDuplicateProcessing(t *testing.T) {
	sys := NewLootDropSystem(12345)
	w := engine.NewWorld()

	entity := w.AddEntity()
	table := NewLootTableWithRNG(rng.NewRNG(100))
	table.Drops = []Drop{{ItemID: "test_item", Chance: 1.0}}

	w.AddComponent(entity, &LootDropComponent{
		Table:      table,
		DropSeed:   100,
		DropChance: 1.0,
	})
	w.AddComponent(entity, &HealthComponent{Current: 0, Max: 100})
	w.AddComponent(entity, &PositionComponent{X: 10, Y: 10})

	spawnCount := 0
	sys.SetLootSpawnCallback(func(itemID string, x, y float64, rarity Rarity) {
		spawnCount++
	})

	// Run system multiple times
	sys.Update(w)
	sys.Update(w)
	sys.Update(w)

	// Should only spawn loot once
	if spawnCount != 1 {
		t.Errorf("Loot spawned %d times, want 1 (no duplicate processing)", spawnCount)
	}
}

func TestCreateEnemyLootTable(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	tiers := []int{1, 2, 3}

	for _, genre := range genres {
		for _, tier := range tiers {
			t.Run(genre+"_tier_"+string(rune('0'+tier)), func(t *testing.T) {
				table := CreateEnemyLootTable(genre, tier, 12345)
				if table == nil {
					t.Fatal("CreateEnemyLootTable returned nil")
				}
				if len(table.Drops) == 0 {
					t.Error("Enemy loot table should have drops")
				}

				// Higher tiers should have more/better drops
				if tier == 3 && len(table.Drops) < 3 {
					t.Error("Boss tier should have substantial loot table")
				}
			})
		}
	}
}

func TestCreateLootDropComponent(t *testing.T) {
	table := NewLootTable()
	comp := CreateLootDropComponent(table, 12345, 0.75)

	if comp == nil {
		t.Fatal("CreateLootDropComponent returned nil")
	}
	if comp.Table != table {
		t.Error("LootTable not set correctly")
	}
	if comp.DropSeed != 12345 {
		t.Errorf("DropSeed = %d, want 12345", comp.DropSeed)
	}
	if comp.DropChance != 0.75 {
		t.Errorf("DropChance = %f, want 0.75", comp.DropChance)
	}
}

func TestLootItemComponent_Type(t *testing.T) {
	comp := &LootItemComponent{}
	if comp.Type() != "LootItemComponent" {
		t.Errorf("Type() = %s, want LootItemComponent", comp.Type())
	}
}

func TestLootDropComponent_Type(t *testing.T) {
	comp := &LootDropComponent{}
	if comp.Type() != "LootDropComponent" {
		t.Errorf("Type() = %s, want LootDropComponent", comp.Type())
	}
}

func TestRarity_String(t *testing.T) {
	tests := []struct {
		rarity Rarity
		want   string
	}{
		{RarityCommon, "Common"},
		{RarityUncommon, "Uncommon"},
		{RarityRare, "Rare"},
		{RarityLegendary, "Legendary"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.rarity.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestLootDropSystem_LootItemDespawn(t *testing.T) {
	sys := NewLootDropSystem(12345)
	w := engine.NewWorld()

	// Create loot item with short lifetime
	lootEntity := w.AddEntity()
	w.AddComponent(lootEntity, &LootItemComponent{
		ItemID:      "test_item",
		Rarity:      RarityCommon,
		SpawnTime:   0.0,
		LifetimeMax: 1.0, // 1 second lifetime
	})

	// Advance time past lifetime
	for i := 0; i < 100; i++ { // 100 frames = ~1.6 seconds at 60fps
		sys.Update(w)
	}

	// Verify entity was removed
	entities := w.Query()
	for _, e := range entities {
		if e == lootEntity {
			t.Error("Loot item should have despawned after lifetime expired")
		}
	}
}

func TestGenreDropConfig_AllGenres(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := getDefaultGenreConfig(genre)

			if config.DefaultDropChance <= 0 || config.DefaultDropChance > 1 {
				t.Errorf("Invalid DefaultDropChance: %f", config.DefaultDropChance)
			}
			if config.DefaultLifetime <= 0 {
				t.Errorf("Invalid DefaultLifetime: %f", config.DefaultLifetime)
			}
			if len(config.CommonDrops) == 0 {
				t.Error("Genre should have common drops")
			}
			if len(config.UncommonDrops) == 0 {
				t.Error("Genre should have uncommon drops")
			}
			if len(config.RareDrops) == 0 {
				t.Error("Genre should have rare drops")
			}
			if len(config.LegendaryDrops) == 0 {
				t.Error("Genre should have legendary drops")
			}
		})
	}
}

func TestLootDropSystem_MultipleEntities(t *testing.T) {
	sys := NewLootDropSystem(12345)
	w := engine.NewWorld()

	// Create multiple dead entities with loot
	for i := 0; i < 5; i++ {
		entity := w.AddEntity()
		table := NewLootTableWithRNG(rng.NewRNG(uint64(100 + i)))
		table.Drops = []Drop{{ItemID: "test_item", Chance: 1.0}}

		w.AddComponent(entity, &LootDropComponent{
			Table:      table,
			DropSeed:   uint64(100 + i),
			DropChance: 1.0,
		})
		w.AddComponent(entity, &HealthComponent{Current: 0, Max: 100})
		w.AddComponent(entity, &PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
	}

	spawnCount := 0
	sys.SetLootSpawnCallback(func(itemID string, x, y float64, rarity Rarity) {
		spawnCount++
	})

	sys.Update(w)

	// Should spawn loot for all 5 entities
	if spawnCount != 5 {
		t.Errorf("Expected 5 loot spawns, got %d", spawnCount)
	}
}

func TestNewLootTableWithRNG(t *testing.T) {
	r := rng.NewRNG(12345)
	table := NewLootTableWithRNG(r)

	if table == nil {
		t.Fatal("NewLootTableWithRNG returned nil")
	}
	if table.rng != r {
		t.Error("RNG not set correctly")
	}
	if table.Drops == nil {
		t.Error("Drops should be initialized")
	}
}

func TestRollWithSeed(t *testing.T) {
	table := NewLootTable()
	table.Drops = []Drop{
		{ItemID: "item1", Chance: 1.0},
		{ItemID: "item2", Chance: 0.5},
		{ItemID: "item3", Chance: 0.5},
	}

	// Same seed should give same results
	result1 := table.RollWithSeed(999)
	result2 := table.RollWithSeed(999)

	if len(result1) != len(result2) {
		t.Errorf("Same seed produced different result counts: %d vs %d", len(result1), len(result2))
	}

	for i := range result1 {
		if i >= len(result2) {
			break
		}
		if result1[i] != result2[i] {
			t.Errorf("Item %d mismatch: %s vs %s", i, result1[i], result2[i])
		}
	}
}

func TestRoll_UsesInternalRNG(t *testing.T) {
	r := rng.NewRNG(12345)
	table := NewLootTableWithRNG(r)
	table.Drops = []Drop{
		{ItemID: "item1", Chance: 1.0},
		{ItemID: "item2", Chance: 0.5},
	}

	result := table.Roll()
	if result == nil {
		t.Fatal("Roll() returned nil")
	}

	// Should always include item1 (100% chance)
	found := false
	for _, item := range result {
		if item == "item1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Guaranteed item not in roll result")
	}
}

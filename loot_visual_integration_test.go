package main

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/loot"
)

// TestLootVisualSystemIntegration tests that loot visuals can be spawned and updated.
func TestLootVisualSystemIntegration(t *testing.T) {
	world := engine.NewWorld()
	visualSys := loot.NewVisualSystem("fantasy")

	world.AddSystem(visualSys)

	items := []struct {
		itemID string
		rarity loot.Rarity
		x, y   float64
		seed   int64
	}{
		{"health_potion", loot.RarityCommon, 10.5, 15.3, 12345},
		{"scroll_fireball", loot.RarityRare, 20.1, 25.7, 54321},
		{"sword_legendary", loot.RarityLegendary, 30.8, 35.2, 99999},
		{"gold_coins_100", loot.RarityCommon, 40.2, 45.9, 11111},
	}

	var entities []engine.Entity
	for _, item := range items {
		ent := loot.SpawnLootVisual(world, item.itemID, item.rarity, item.x, item.y, item.seed)
		entities = append(entities, ent)
	}

	if len(entities) != 4 {
		t.Fatalf("expected 4 entities, got %d", len(entities))
	}

	for i := 0; i < 10; i++ {
		world.Update()
	}

	t.Log("Loot visual system integration test passed")
}

// TestLootCategorizationIntegration verifies item categorization matches expected categories.
func TestLootCategorizationIntegration(t *testing.T) {
	tests := []struct {
		itemID   string
		expected loot.ItemCategory
	}{
		{"health_potion", loot.CategoryPotion},
		{"scroll_fireball", loot.CategoryScroll},
		{"weapon_enchanted", loot.CategoryWeapon},
		{"armor_plate", loot.CategoryArmor},
		{"gold_coins_100", loot.CategoryGold},
		{"circuit_board", loot.CategoryGear},
		{"artifact_ancient", loot.CategoryArtifact},
	}

	for _, tt := range tests {
		result := loot.CategorizeItem(tt.itemID)
		if result != tt.expected {
			t.Errorf("CategorizeItem(%s) = %v, want %v", tt.itemID, result, tt.expected)
		}
	}
}

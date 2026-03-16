package loot

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestVisualSystemCreation(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy genre", "fantasy"},
		{"scifi genre", "scifi"},
		{"cyberpunk genre", "cyberpunk"},
		{"horror genre", "horror"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vs := NewVisualSystem(tt.genreID)
			if vs == nil {
				t.Fatal("NewVisualSystem returned nil")
			}
			if vs.genreID != tt.genreID {
				t.Errorf("expected genreID %s, got %s", tt.genreID, vs.genreID)
			}
		})
	}
}

func TestVisualComponentCreation(t *testing.T) {
	vc := &VisualComponent{
		ItemID:   "health_potion",
		Category: CategoryPotion,
		Rarity:   RarityCommon,
		Seed:     12345,
	}

	if vc.Type() != "LootVisual" {
		t.Errorf("expected Type() to return 'LootVisual', got %s", vc.Type())
	}
}

func TestUpdateBobAndGlow(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	world := engine.NewWorld()

	ent := world.AddEntity()
	vc := &VisualComponent{
		ItemID:    "test_item",
		Category:  CategoryPotion,
		Rarity:    RarityCommon,
		BobPhase:  0,
		GlowPhase: 0,
	}
	world.AddComponent(ent, vc)

	vs.Update(world)

	if vc.BobPhase == 0 {
		t.Error("BobPhase should have been updated")
	}
	if vc.GlowPhase == 0 {
		t.Error("GlowPhase should have been updated")
	}
}

func TestUpdatePhaseWrapping(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	world := engine.NewWorld()

	ent := world.AddEntity()
	vc := &VisualComponent{
		ItemID:    "test_item",
		Category:  CategoryPotion,
		Rarity:    RarityLegendary,
		BobPhase:  2.0*math.Pi - 0.1,
		GlowPhase: 2.0*math.Pi - 0.1,
	}
	world.AddComponent(ent, vc)

	// Run multiple updates to ensure wrapping
	for i := 0; i < 100; i++ {
		vs.Update(world)
	}

	if vc.BobPhase > 2.0*math.Pi {
		t.Errorf("BobPhase should wrap at 2π, got %f", vc.BobPhase)
	}
	if vc.GlowPhase > 2.0*math.Pi {
		t.Errorf("GlowPhase should wrap at 2π, got %f", vc.GlowPhase)
	}
}

func TestUpdateSkipsCollectedItems(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	world := engine.NewWorld()

	ent := world.AddEntity()
	vc := &VisualComponent{
		ItemID:    "test_item",
		Category:  CategoryPotion,
		Rarity:    RarityCommon,
		Collected: true,
		BobPhase:  1.0,
		GlowPhase: 1.0,
	}
	world.AddComponent(ent, vc)

	originalBob := vc.BobPhase
	originalGlow := vc.GlowPhase

	vs.Update(world)

	if vc.BobPhase != originalBob {
		t.Error("Collected items should not update BobPhase")
	}
	if vc.GlowPhase != originalGlow {
		t.Error("Collected items should not update GlowPhase")
	}
}

func TestGeneratePotionSprite(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	img := vs.GenerateItemSprite("health_potion", CategoryPotion, RarityCommon, 12345, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for potion")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("expected 32x32 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateScrollSprite(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	img := vs.GenerateItemSprite("scroll_fireball", CategoryScroll, RarityRare, 54321, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for scroll")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("expected 32x32 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateWeaponSprite(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	img := vs.GenerateItemSprite("sword_legendary", CategoryWeapon, RarityLegendary, 99999, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for weapon")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("expected 32x32 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateArmorSprite(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	img := vs.GenerateItemSprite("shield_blessed", CategoryArmor, RarityUncommon, 11111, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for armor")
	}
}

func TestGenerateGoldSprite(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	img := vs.GenerateItemSprite("gold_coins_100", CategoryGold, RarityCommon, 22222, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for gold")
	}
}

func TestGenerateGearSprite(t *testing.T) {
	vs := NewVisualSystem("scifi")
	img := vs.GenerateItemSprite("circuit_board", CategoryGear, RarityRare, 33333, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for gear")
	}
}

func TestGenerateArtifactSprite(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	img := vs.GenerateItemSprite("relic_ancient", CategoryArtifact, RarityLegendary, 44444, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for artifact")
	}
}

func TestGenerateConsumableSprite(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	img := vs.GenerateItemSprite("bread", CategoryConsumable, RarityCommon, 55555, 32)

	if img == nil {
		t.Fatal("GenerateItemSprite returned nil for consumable")
	}
}

func TestGenerateVariousSizes(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	sizes := []int{16, 24, 32, 48, 64}

	for _, size := range sizes {
		img := vs.GenerateItemSprite("test_item", CategoryPotion, RarityCommon, 12345, size)
		if img == nil {
			t.Errorf("GenerateItemSprite returned nil for size %d", size)
			continue
		}

		bounds := img.Bounds()
		if bounds.Dx() != size || bounds.Dy() != size {
			t.Errorf("expected %dx%d image, got %dx%d", size, size, bounds.Dx(), bounds.Dy())
		}
	}
}

func TestDeterministicGeneration(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	seed := int64(99999)

	img1 := vs.GenerateItemSprite("test_item", CategoryPotion, RarityCommon, seed, 32)
	img2 := vs.GenerateItemSprite("test_item", CategoryPotion, RarityCommon, seed, 32)

	if img1 == nil || img2 == nil {
		t.Fatal("GenerateItemSprite returned nil")
	}

	bounds1 := img1.Bounds()
	bounds2 := img2.Bounds()

	if bounds1 != bounds2 {
		t.Errorf("expected same image bounds for deterministic generation: %v vs %v", bounds1, bounds2)
	}
}

func TestCategorizeItem(t *testing.T) {
	tests := []struct {
		itemID   string
		expected ItemCategory
	}{
		{"health_potion", CategoryPotion},
		{"mana_potion", CategoryPotion},
		{"elixir_strength", CategoryPotion},
		{"scroll_fireball", CategoryScroll},
		{"tome_wisdom", CategoryScroll},
		{"sword_iron", CategoryWeapon},
		{"axe_battle", CategoryWeapon},
		{"bow_longbow", CategoryWeapon},
		{"weapon_enchanted", CategoryWeapon},
		{"armor_plate", CategoryArmor},
		{"shield_wooden", CategoryArmor},
		{"helm_steel", CategoryArmor},
		{"gold_coins_100", CategoryGold},
		{"money_pouch", CategoryGold},
		{"circuit_board", CategoryGear},
		{"tech_module", CategoryGear},
		{"artifact_ancient", CategoryArtifact},
		{"relic_cursed", CategoryArtifact},
		{"enchanted_ring", CategoryArtifact},
		{"bread", CategoryConsumable},
		{"food_ration", CategoryConsumable},
		{"meat_jerky", CategoryConsumable},
		{"unknown_item", CategoryConsumable},
	}

	for _, tt := range tests {
		t.Run(tt.itemID, func(t *testing.T) {
			result := CategorizeItem(tt.itemID)
			if result != tt.expected {
				t.Errorf("CategorizeItem(%s) = %v, want %v", tt.itemID, result, tt.expected)
			}
		})
	}
}

func TestRarityAffectsVisuals(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	rarities := []Rarity{RarityCommon, RarityUncommon, RarityRare, RarityLegendary}

	for _, rarity := range rarities {
		img := vs.GenerateItemSprite("test_weapon", CategoryWeapon, rarity, 12345, 32)
		if img == nil {
			t.Errorf("failed to generate sprite for rarity %v", rarity)
		}
	}
}

func TestGenreSpecificRendering(t *testing.T) {
	genres := []string{"fantasy", "scifi", "cyberpunk", "horror"}

	for _, genre := range genres {
		vs := NewVisualSystem(genre)
		img := vs.GenerateItemSprite("gear_item", CategoryGear, RarityCommon, 12345, 32)
		if img == nil {
			t.Errorf("failed to generate sprite for genre %s", genre)
		}
	}
}

func TestUpdateWithNoEntities(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	world := engine.NewWorld()

	vs.Update(world)
}

func TestUpdateWithoutVisualComponent(t *testing.T) {
	vs := NewVisualSystem("fantasy")
	world := engine.NewWorld()
	world.AddEntity()

	vs.Update(world)
}

func BenchmarkGeneratePotion(b *testing.B) {
	vs := NewVisualSystem("fantasy")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		vs.GenerateItemSprite("health_potion", CategoryPotion, RarityCommon, int64(i), 32)
	}
}

func BenchmarkGenerateWeapon(b *testing.B) {
	vs := NewVisualSystem("fantasy")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		vs.GenerateItemSprite("sword", CategoryWeapon, RarityLegendary, int64(i), 32)
	}
}

func BenchmarkGenerateArtifact(b *testing.B) {
	vs := NewVisualSystem("fantasy")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		vs.GenerateItemSprite("relic", CategoryArtifact, RarityLegendary, int64(i), 32)
	}
}

func BenchmarkUpdate100Entities(b *testing.B) {
	vs := NewVisualSystem("fantasy")
	world := engine.NewWorld()

	for i := 0; i < 100; i++ {
		ent := world.AddEntity()
		vc := &VisualComponent{
			ItemID:   "test_item",
			Category: CategoryPotion,
			Rarity:   RarityCommon,
		}
		world.AddComponent(ent, vc)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vs.Update(world)
	}
}

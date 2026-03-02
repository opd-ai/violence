package equipment

import (
	"image/color"
	"testing"
)

func TestEquipmentComponent_Type(t *testing.T) {
	ec := &EquipmentComponent{}
	if ec.Type() != "equipment" {
		t.Errorf("Expected type 'equipment', got '%s'", ec.Type())
	}
}

func TestNewEquipmentSystem(t *testing.T) {
	sys := NewEquipmentSystem("fantasy")
	if sys == nil {
		t.Fatal("NewEquipmentSystem returned nil")
	}
	if sys.genre != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", sys.genre)
	}
	if sys.layerCache == nil {
		t.Error("layerCache not initialized")
	}
	if sys.poolBySize == nil {
		t.Error("poolBySize not initialized")
	}
}

func TestEquip(t *testing.T) {
	ec := &EquipmentComponent{}

	item := &Equipment{
		Slot:     SlotWeapon,
		Material: MaterialSteel,
		Rarity:   RarityRare,
		Seed:     12345,
		Name:     "Test Sword",
	}

	Equip(ec, item)

	if ec.Items[SlotWeapon] == nil {
		t.Fatal("Equipment not equipped")
	}
	if ec.Items[SlotWeapon].Name != "Test Sword" {
		t.Errorf("Expected 'Test Sword', got '%s'", ec.Items[SlotWeapon].Name)
	}
	if !ec.DirtyCache {
		t.Error("DirtyCache should be set to true after equip")
	}
}

func TestUnequip(t *testing.T) {
	ec := &EquipmentComponent{}

	item := &Equipment{
		Slot:     SlotHelmet,
		Material: MaterialIron,
		Rarity:   RarityCommon,
		Seed:     54321,
		Name:     "Test Helmet",
	}

	Equip(ec, item)
	ec.DirtyCache = false

	unequipped := Unequip(ec, SlotHelmet)

	if unequipped == nil {
		t.Fatal("Unequip returned nil")
	}
	if unequipped.Name != "Test Helmet" {
		t.Errorf("Expected 'Test Helmet', got '%s'", unequipped.Name)
	}
	if ec.Items[SlotHelmet] != nil {
		t.Error("Helmet should be nil after unequip")
	}
	if !ec.DirtyCache {
		t.Error("DirtyCache should be set to true after unequip")
	}
}

func TestGetEquipped(t *testing.T) {
	ec := &EquipmentComponent{}

	item := &Equipment{
		Slot:     SlotChest,
		Material: MaterialLeather,
		Rarity:   RarityUncommon,
		Seed:     99999,
		Name:     "Test Armor",
	}

	Equip(ec, item)

	retrieved := GetEquipped(ec, SlotChest)
	if retrieved == nil {
		t.Fatal("GetEquipped returned nil")
	}
	if retrieved.Name != "Test Armor" {
		t.Errorf("Expected 'Test Armor', got '%s'", retrieved.Name)
	}

	empty := GetEquipped(ec, SlotBoots)
	if empty != nil {
		t.Error("GetEquipped should return nil for empty slot")
	}
}

func TestEquipmentSystem_SetGenre(t *testing.T) {
	sys := NewEquipmentSystem("fantasy")

	sys.SetGenre("scifi")

	if sys.genre != "scifi" {
		t.Errorf("Expected genre 'scifi', got '%s'", sys.genre)
	}
}

func TestGetMaterialColor(t *testing.T) {
	sys := NewEquipmentSystem("fantasy")

	tests := []struct {
		material Material
		name     string
	}{
		{MaterialIron, "Iron"},
		{MaterialSteel, "Steel"},
		{MaterialMithril, "Mithril"},
		{MaterialLeather, "Leather"},
		{MaterialCloth, "Cloth"},
		{MaterialDragonscale, "Dragonscale"},
		{MaterialCrystal, "Crystal"},
		{MaterialNanofiber, "Nanofiber"},
		{MaterialBiotech, "Biotech"},
		{MaterialPlasma, "Plasma"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sys.getMaterialColor(tt.material)
			if c.A != 255 {
				t.Errorf("Expected alpha 255, got %d", c.A)
			}
		})
	}
}

func TestGetRarityAccent(t *testing.T) {
	sys := NewEquipmentSystem("fantasy")

	tests := []struct {
		rarity Rarity
		name   string
	}{
		{RarityCommon, "Common"},
		{RarityUncommon, "Uncommon"},
		{RarityRare, "Rare"},
		{RarityEpic, "Epic"},
		{RarityLegendary, "Legendary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sys.getRarityAccent(tt.rarity)
			if c.A != 255 {
				t.Errorf("Expected alpha 255, got %d", c.A)
			}
		})
	}
}

func TestBufferPooling(t *testing.T) {
	sys := NewEquipmentSystem("fantasy")

	sizeKey := 32 * 32 * 4

	buf1 := sys.acquireImageBuffer(sizeKey)
	if buf1 == nil {
		t.Fatal("acquireImageBuffer returned nil")
	}

	sys.releaseImageBuffer(buf1)

	buf2 := sys.acquireImageBuffer(sizeKey)
	if buf2 == nil {
		t.Fatal("acquireImageBuffer returned nil after release")
	}

	if buf1 != buf2 {
		t.Error("Buffer pooling should reuse released buffers")
	}
}

func TestMultipleEquipmentSlots(t *testing.T) {
	ec := &EquipmentComponent{}

	items := []*Equipment{
		{Slot: SlotWeapon, Name: "Sword", Seed: 1},
		{Slot: SlotHelmet, Name: "Helm", Seed: 2},
		{Slot: SlotChest, Name: "Plate", Seed: 3},
		{Slot: SlotLegs, Name: "Greaves", Seed: 4},
		{Slot: SlotBoots, Name: "Boots", Seed: 5},
		{Slot: SlotGloves, Name: "Gauntlets", Seed: 6},
		{Slot: SlotAccessory1, Name: "Ring1", Seed: 7},
		{Slot: SlotAccessory2, Name: "Ring2", Seed: 8},
	}

	for _, item := range items {
		Equip(ec, item)
	}

	for i, item := range items {
		equipped := GetEquipped(ec, item.Slot)
		if equipped == nil {
			t.Errorf("Slot %d should have equipment", i)
		} else if equipped.Name != item.Name {
			t.Errorf("Expected %s, got %s", item.Name, equipped.Name)
		}
	}
}

func TestEquipNilCases(t *testing.T) {
	ec := &EquipmentComponent{}
	item := &Equipment{Slot: SlotWeapon, Name: "Test"}

	Equip(nil, item)
	Equip(ec, nil)

	if ec.Items[SlotWeapon] != nil {
		t.Error("Equip with nil should not modify component")
	}

	result := Unequip(nil, SlotWeapon)
	if result != nil {
		t.Error("Unequip on nil component should return nil")
	}

	result = GetEquipped(nil, SlotWeapon)
	if result != nil {
		t.Error("GetEquipped on nil component should return nil")
	}
}

func TestEnchantedEquipmentGlowPhase(t *testing.T) {
	ec := &EquipmentComponent{}

	enchantedItem := &Equipment{
		Slot:         SlotWeapon,
		Material:     MaterialMithril,
		Rarity:       RarityLegendary,
		Enchanted:    true,
		EnchantColor: color.RGBA{255, 100, 255, 255},
		Seed:         11111,
		Name:         "Enchanted Blade",
	}
	Equip(ec, enchantedItem)

	if ec.GlowPhase != 0 {
		t.Error("Initial GlowPhase should be 0")
	}
}

func TestDamageStates(t *testing.T) {
	states := []DamageState{StatePristine, StateWorn, StateDamaged, StateBroken}

	for _, state := range states {
		item := &Equipment{
			Slot:        SlotChest,
			Material:    MaterialIron,
			Rarity:      RarityCommon,
			DamageState: state,
			Seed:        12345,
		}

		if item.DamageState != state {
			t.Errorf("DamageState not set correctly")
		}
	}
}

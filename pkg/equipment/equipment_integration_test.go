//go:build integration
// +build integration

package equipment

import (
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// This test requires a graphics context and should only run with -tags integration
func TestEquipmentSystem_Integration(t *testing.T) {
	sys := NewEquipmentSystem("fantasy")
	world := engine.NewWorld()

	entity := world.AddEntity()
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

	world.AddComponent(entity, ec)

	initialPhase := ec.GlowPhase
	sys.Update(world)

	if ec.GlowPhase <= initialPhase {
		t.Error("GlowPhase should increase for enchanted equipment")
	}
}

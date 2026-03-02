package itemicon_test

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/itemicon"
)

// This example demonstrates basic icon generation for common item types.
func ExampleIconSystem_GenerateIcon() {
	// Create icon system for fantasy genre with cache size of 200
	iconSys := itemicon.NewSystem("fantasy", 200)

	// Generate a legendary sword icon
	swordComp := &itemicon.ItemIconComponent{
		Seed:         12345,
		IconType:     "weapon",
		Rarity:       4, // legendary
		SubType:      "sword",
		IconSize:     48,
		BorderGlow:   true,
		EnchantLevel: 3,
		Durability:   1.0,
	}
	swordIcon := iconSys.GenerateIcon(swordComp)
	fmt.Printf("Sword icon: %dx%d\n", swordIcon.Bounds().Dx(), swordIcon.Bounds().Dy())

	// Generate a rare health potion icon
	potionComp := &itemicon.ItemIconComponent{
		Seed:       67890,
		IconType:   "consumable",
		Rarity:     2, // rare
		SubType:    "potion",
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}
	potionIcon := iconSys.GenerateIcon(potionComp)
	fmt.Printf("Potion icon: %dx%d\n", potionIcon.Bounds().Dx(), potionIcon.Bounds().Dy())

	// Generate an epic armor icon with wear
	armorComp := &itemicon.ItemIconComponent{
		Seed:       11111,
		IconType:   "armor",
		Rarity:     3, // epic
		SubType:    "",
		IconSize:   48,
		BorderGlow: true,
		Durability: 0.6, // 60% durability
	}
	armorIcon := iconSys.GenerateIcon(armorComp)
	fmt.Printf("Armor icon: %dx%d\n", armorIcon.Bounds().Dx(), armorIcon.Bounds().Dy())

	// Output:
	// Sword icon: 48x48
	// Potion icon: 48x48
	// Armor icon: 48x48
}

// This example shows how to render icons to the game screen.
func ExampleIconSystem_renderToScreen() {
	iconSys := itemicon.NewSystem("fantasy", 200)

	// Create a rare material icon
	comp := &itemicon.ItemIconComponent{
		Seed:       99999,
		IconType:   "material",
		Rarity:     2,
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}

	// Generate the icon
	icon := iconSys.GenerateIcon(comp)

	// In actual game code, you would render it like this:
	// screen := ebiten.NewImage(800, 600) // game screen
	// opts := &ebiten.DrawImageOptions{}
	// opts.GeoM.Translate(100, 100) // position on screen
	// screen.DrawImage(icon, opts)

	fmt.Printf("Material icon ready: %dx%d\n", icon.Bounds().Dx(), icon.Bounds().Dy())

	// Output:
	// Material icon ready: 48x48
}

// This example demonstrates how genre affects icon appearance.
func ExampleIconSystem_SetGenre() {
	// Fantasy icons have warm metal tones and golden enchantments
	fantasySys := itemicon.NewSystem("fantasy", 100)

	// Scifi icons have cool metal tones and cyan enchantments
	scifiSys := itemicon.NewSystem("scifi", 100)

	// Cyberpunk icons have dark metals and neon enchantments
	cyberpunkSys := itemicon.NewSystem("cyberpunk", 100)

	comp := &itemicon.ItemIconComponent{
		Seed:         12345,
		IconType:     "weapon",
		Rarity:       3,
		SubType:      "sword",
		IconSize:     48,
		BorderGlow:   true,
		EnchantLevel: 2,
		Durability:   1.0,
	}

	fantasyIcon := fantasySys.GenerateIcon(comp)
	scifiIcon := scifiSys.GenerateIcon(comp)
	cyberpunkIcon := cyberpunkSys.GenerateIcon(comp)

	fmt.Printf("Fantasy sword: %dx%d\n", fantasyIcon.Bounds().Dx(), fantasyIcon.Bounds().Dy())
	fmt.Printf("Scifi sword: %dx%d\n", scifiIcon.Bounds().Dx(), scifiIcon.Bounds().Dy())
	fmt.Printf("Cyberpunk sword: %dx%d\n", cyberpunkIcon.Bounds().Dx(), cyberpunkIcon.Bounds().Dy())

	// Output:
	// Fantasy sword: 48x48
	// Scifi sword: 48x48
	// Cyberpunk sword: 48x48
}

// This example shows cache behavior for repeated icon generation.
func ExampleIconSystem_caching() {
	iconSys := itemicon.NewSystem("fantasy", 100)

	comp := &itemicon.ItemIconComponent{
		Seed:       12345,
		IconType:   "weapon",
		Rarity:     2,
		SubType:    "sword",
		IconSize:   48,
		BorderGlow: true,
		Durability: 1.0,
	}

	// First call generates and caches
	icon1 := iconSys.GenerateIcon(comp)

	// Second call returns cached version
	icon2 := iconSys.GenerateIcon(comp)

	// Both pointers reference the same cached image
	same := icon1 == icon2
	fmt.Printf("Cached: %v\n", same)

	// Output:
	// Cached: true
}

// This example demonstrates rarity-based visual differentiation.
func ExampleIconSystem_rarityLevels() {
	iconSys := itemicon.NewSystem("fantasy", 100)

	rarities := []struct {
		level int
		name  string
	}{
		{0, "common"},
		{1, "uncommon"},
		{2, "rare"},
		{3, "epic"},
		{4, "legendary"},
	}

	for _, r := range rarities {
		comp := &itemicon.ItemIconComponent{
			Seed:       int64(r.level * 1000),
			IconType:   "weapon",
			Rarity:     r.level,
			SubType:    "sword",
			IconSize:   48,
			BorderGlow: true,
			Durability: 1.0,
		}

		icon := iconSys.GenerateIcon(comp)
		fmt.Printf("%s: %dx%d\n", r.name, icon.Bounds().Dx(), icon.Bounds().Dy())
	}

	// Output:
	// common: 48x48
	// uncommon: 48x48
	// rare: 48x48
	// epic: 48x48
	// legendary: 48x48
}

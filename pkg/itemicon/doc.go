// Package itemicon provides procedural item icon generation with rarity-based visual effects.
//
// The itemicon package generates visually distinctive icons for items in the game inventory and loot drops.
// Icons are procedurally generated based on item type, rarity, enchantment level, and durability, ensuring
// every item has a unique, recognizable visual representation.
//
// # Features
//
// - **Type-Specific Rendering**: Different icon styles for weapons, armor, consumables, materials, and quest items
// - **Rarity Visual Differentiation**: Color-coded borders and glows for common through legendary items
// - **Enchantment Effects**: Magical sparkles and glows for enchanted items
// - **Durability Visualization**: Wear and damage effects for items with low durability
// - **Genre-Aware Styling**: Metal colors, accents, and enchantment effects adapt to game genre (fantasy, scifi, cyberpunk, horror)
// - **LRU Caching**: Efficient caching system to avoid regenerating icons
//
// # Example Usage
//
//	// Create icon system
//	iconSys := itemicon.NewSystem("fantasy", 200)
//
//	// Create icon component for a legendary enchanted sword
//	comp := &itemicon.ItemIconComponent{
//		Seed:         12345,
//		IconType:     "weapon",
//		Rarity:       4, // legendary
//		SubType:      "sword",
//		IconSize:     48,
//		BorderGlow:   true,
//		EnchantLevel: 3,
//		Durability:   1.0,
//	}
//
//	// Generate icon
//	icon := iconSys.GenerateIcon(comp)
//
//	// Draw icon to screen
//	opts := &ebiten.DrawImageOptions{}
//	opts.GeoM.Translate(x, y)
//	screen.DrawImage(icon, opts)
package itemicon

// Package equipment provides visual rendering of equipped items on entity sprites.
//
// # Overview
//
// The equipment system allows entities to visibly wear armor, weapons, and accessories
// with genre-specific materials, rarity-based visual complexity, damage state weathering,
// and enchantment glow effects.
//
// # Equipment Slots
//
// Eight equipment slots are available:
//   - SlotWeapon: Main weapon (sword, gun, staff)
//   - SlotHelmet: Head protection
//   - SlotChest: Torso armor
//   - SlotLegs: Leg armor
//   - SlotBoots: Footwear
//   - SlotGloves: Hand protection
//   - SlotAccessory1: Ring, amulet, or other accessory
//   - SlotAccessory2: Second accessory slot
//
// # Materials
//
// Equipment can be crafted from various materials, each with distinct visual appearance:
//   - Fantasy: Iron, Steel, Mithril, Leather, Cloth, Dragonscale, Crystal
//   - Sci-Fi/Cyberpunk: Nanofiber, Biotech, Plasma
//
// # Rarity System
//
// Equipment rarity affects visual complexity and enchantment visibility:
//   - Common: Simple, basic appearance
//   - Uncommon: Slight visual enhancements
//   - Rare: Ornate details, accent colors
//   - Epic: Elaborate decorations, plumes, crests
//   - Legendary: Maximum detail, enchantment glow
//
// # Damage States
//
// Equipment degrades visually as it takes damage:
//   - StatePristine: Perfect condition
//   - StateWorn: Minor scratches and wear
//   - StateDamaged: Visible damage, multiple scratches
//   - StateBroken: Heavily damaged, near-unusable appearance
//
// # Usage Example
//
//	// Create equipment system
//	sys := equipment.NewEquipmentSystem("fantasy")
//	world.AddSystem(sys)
//
//	// Create entity with equipment component
//	entity := world.AddEntity()
//	ec := &equipment.EquipmentComponent{}
//	world.AddComponent(entity, ec)
//
//	// Equip a legendary enchanted sword
//	sword := &equipment.Equipment{
//	    Slot:         equipment.SlotWeapon,
//	    Material:     equipment.MaterialMithril,
//	    Rarity:       equipment.RarityLegendary,
//	    Enchanted:    true,
//	    EnchantColor: color.RGBA{100, 200, 255, 255},
//	    Seed:         12345,
//	    Name:         "Frostbite",
//	}
//	equipment.Equip(ec, sword)
//
//	// Render equipment onto entity sprite
//	sprite := sys.RenderEquipmentLayer(ec, baseSprite, direction)
//
// # Performance
//
// The equipment system uses:
//   - LRU caching for generated equipment sprites
//   - Memory pooling for image buffers
//   - Dirty cache tracking to avoid redundant generation
//   - Genre-aware sprite generation
//
// # Integration
//
// The equipment system integrates with:
//   - Animation system: Equipment renders in correct z-order with entity sprites
//   - Lighting system: Enchantment glow adds dynamic light
//   - Combat system: Damage state updates based on equipment durability
//   - Loot system: Procedurally generated equipment from drops
package equipment

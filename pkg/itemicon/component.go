// Package itemicon provides procedural item icon generation.
package itemicon

// ItemIconComponent stores visual representation data for items.
type ItemIconComponent struct {
	Seed         int64
	IconType     string  // "weapon", "armor", "consumable", "material", "quest"
	Rarity       int     // 0=common, 1=uncommon, 2=rare, 3=epic, 4=legendary
	SubType      string  // Specific item category (e.g., "sword", "potion", "ore")
	IconSize     int     // Pixel dimensions (32, 48, 64)
	BorderGlow   bool    // Whether to render rarity glow
	EnchantLevel int     // 0-5, adds visual effects
	Durability   float64 // 0.0-1.0, affects visual wear
}

// Type implements Component interface.
func (c *ItemIconComponent) Type() string {
	return "itemicon"
}

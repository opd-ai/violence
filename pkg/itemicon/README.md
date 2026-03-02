# Item Icon Generation System

Procedural item icon generation with rarity-based visual effects for inventory and loot display.

## Features

- **Type-Specific Rendering**: Different icon styles for weapons, armor, consumables, materials, and quest items
- **Rarity Visual Differentiation**: Color-coded borders and glows for common through legendary items
- **Enchantment Effects**: Magical sparkles and glows for enchanted items (levels 1-5)
- **Durability Visualization**: Wear and damage effects for items with low durability
- **Genre-Aware Styling**: Metal colors, accents, and enchantment effects adapt to game genre
- **LRU Caching**: Efficient caching system to avoid regenerating icons
- **Seed-Based Generation**: Deterministic icons for consistent appearance across sessions

## Usage

### Basic Icon Generation

```go
import "github.com/opd-ai/violence/pkg/itemicon"

// Create icon system for fantasy genre with cache size of 200
iconSys := itemicon.NewSystem("fantasy", 200)

// Create icon component for a legendary enchanted sword
comp := &itemicon.ItemIconComponent{
    Seed:         12345,
    IconType:     "weapon",
    Rarity:       4, // legendary (0-4)
    SubType:      "sword",
    IconSize:     48,
    BorderGlow:   true,
    EnchantLevel: 3, // magical level (0-5)
    Durability:   1.0, // full durability (0.0-1.0)
}

// Generate icon
icon := iconSys.GenerateIcon(comp)

// Draw to screen
opts := &ebiten.DrawImageOptions{}
opts.GeoM.Translate(x, y)
screen.DrawImage(icon, opts)
```

### Item Types

- **"weapon"**: Swords, axes, and generic weapons (subtypes: "sword", "axe")
- **"armor"**: Chest pieces and protective gear
- **"consumable"**: Potions and scrolls (subtypes: "potion", "scroll")
- **"material"**: Crafting materials with crystalline shards
- **"quest"**: Quest items with golden star icon

### Rarity Levels

| Level | Name      | Border Color     | Base Color Tint |
|-------|-----------|------------------|-----------------|
| 0     | Common    | Gray (100)       | Gray            |
| 1     | Uncommon  | Green (50,255,50)| Green           |
| 2     | Rare      | Blue (50,150,255)| Blue            |
| 3     | Epic      | Purple (200,50,255) | Purple       |
| 4     | Legendary | Gold (255,200,0) | Gold            |

### Genre-Specific Colors

**Fantasy** (default):
- Metal: Silver-gray (180, 180, 200)
- Accent: Golden brown (200, 150, 80)
- Enchant: Light blue-purple (200, 200, 255)

**Scifi**:
- Metal: Blue-gray steel (150, 170, 200)
- Accent: Cyan (100, 200, 255)
- Enchant: Bright cyan (100, 255, 255)

**Cyberpunk**:
- Metal: Dark gunmetal (100, 120, 140)
- Accent: Neon pink (255, 0, 128)
- Enchant: Neon magenta (255, 0, 255)

**Horror**:
- Metal: Tarnished metal (120, 100, 100)
- Accent: Blood red (180, 50, 50)
- Enchant: Sickly green (100, 255, 100)

### Advanced Features

#### Enchantment Sparkles
Setting `EnchantLevel` to 1-5 adds magical sparkle effects. Higher levels produce more sparkles.

#### Durability Wear
Setting `Durability` below 1.0 adds visual wear and scratches. Lower values show more damage.

#### Genre Switching
```go
iconSys.SetGenre("cyberpunk") // Clears cache and applies new colors
```

#### Caching
Icons with identical `Seed`, `IconType`, `Rarity`, `SubType`, `Size`, and `EnchantLevel` are cached. The cache uses LRU eviction when full.

## Integration

The system is integrated into the main game loop:

1. **Initialization** (main.go NewGame): Creates `itemIconSystem` with genre and cache size
2. **Genre Changes** (main.go changeGenre): Updates icon system genre when player changes world type
3. **Rendering**: Called when drawing inventory UI, loot drops, or item tooltips

## Performance

- **Generation**: ~0.1-0.5ms per icon (varies by complexity)
- **Cached Access**: <0.01ms per icon
- **Memory**: ~10KB per cached 48x48 icon
- **Cache Size**: Configurable (200 default = ~2MB)

## Testing

Run package tests with:
```bash
go test ./pkg/itemicon/...
```

Note: Tests require a display (GLFW/OpenGL) and will fail in headless environments.

## Architecture

- **component.go**: ECS component for item icon data
- **system.go**: Icon generation, caching, and rendering logic
- **doc.go**: Package documentation
- **system_test.go**: Comprehensive test suite
- **example_test.go**: Usage examples

## Visual Quality

Icons use shading, gradients, and highlights to maximize visual distinctiveness at small sizes (32-64px). Every pixel matters:
- Weapons show blade highlights and grip details
- Armor displays torso shape and shoulder pauldrons
- Potions render glass with liquid levels
- Materials show crystalline shard patterns
- Quest items display golden stars

Rarity affects both border glow and base color tint, ensuring items are instantly recognizable by rarity tier.

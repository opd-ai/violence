# Loot Visual System

## Overview

The Loot Visual System provides procedural sprite generation and rendering for dropped items in the game world. Each item type has a distinct visual representation with rarity-based effects and animations.

## Features

### Item Categories
- **Potions**: Flask-shaped bottles with colored liquids
- **Scrolls**: Rolled parchment with magical runes
- **Weapons**: Swords, axes, and other arms with metallic finishes
- **Armor**: Shields and armor pieces with boss details
- **Gold**: Coins and gold piles
- **Gear**: Mechanical/technological items
- **Artifacts**: Mystical items with complex patterns
- **Consumables**: Food and other consumable items

### Visual Effects
- **Bobbing Animation**: Items float and bob in the air
- **Glow Effect**: Pulsing glow based on rarity
- **Rarity-Based Appearance**:
  - Common: Basic materials (iron, plain colors)
  - Uncommon: Better materials (bronze, enhanced details)
  - Rare: Premium materials (silver, sparkles)
  - Legendary: Gold finish with enchantment glow

## Usage

### Spawning Loot Visuals

```go
import "github.com/opd-ai/violence/pkg/loot"

// Spawn a loot item in the world
entity := loot.SpawnLootVisual(
    world,              // *engine.World
    "health_potion",    // itemID
    loot.RarityRare,    // rarity
    10.5, 15.3,         // x, y position
    12345,              // seed for procedural generation
)
```

### System Integration

The visual system is registered with the ECS World and updates automatically:

```go
visualSystem := loot.NewVisualSystem("fantasy")
world.AddSystem(visualSystem)
```

### Rendering

The main rendering loop queries entities with `LootVisual` and `Position` components and draws them as sprites in world space.

## Architecture

### Components

- **VisualComponent**: Stores visual state (bob phase, glow phase, collected status)
- **PositionComponent**: Stores world position (shared with other systems)

### Visual Generation

- **Deterministic**: Same seed + item type = same sprite
- **Cached**: Uses LRU cache to avoid regeneration
- **Genre-Aware**: Visual style adapts to game genre (fantasy, scifi, cyberpunk, horror)

## Performance

- Sprite generation is seeded and cached
- Bob and glow animations run at 60 FPS
- Distance culling prevents rendering distant items
- Minimal per-frame allocation

## Testing

```bash
go test ./pkg/loot/... -v
go build ./pkg/loot
```

Test coverage includes:
- Sprite generation for all item categories
- Deterministic generation verification
- Animation update logic
- Rarity-based visual effects
- Genre-specific rendering

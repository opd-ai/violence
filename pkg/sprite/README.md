# Sprite Package

Procedural sprite generation system with shading, detail, and LRU caching.

## Overview

The sprite package provides a high-performance procedural sprite generation system that creates detailed 2D sprites with proper shading, highlights, and genre-specific theming. It replaces flat colored rectangles with rich, visually distinct sprites for all entity types in the game.

## Features

- **Procedural Generation**: Sprites are generated algorithmically with deterministic seeds
- **LRU Caching**: Automatically caches generated sprites with configurable capacity
- **Genre-Aware**: Colors and styles adapt to game genre (fantasy, scifi, horror, cyberpunk, postapoc)
- **Pixel-Perfect Detail**: Shading, gradients, and highlights at small sprite sizes
- **Animation Support**: Frame-based animation for torches, pickups, and effects
- **Zero Allocations**: Cached sprites eliminate per-frame allocations
- **Thread-Safe**: Concurrent access protected by RWMutex

## Supported Sprite Types

### Props (SpriteProp)
- `barrel` - Cylindrical barrel with wood grain and metal bands
- `crate` - Wooden crate with planks and corner reinforcements
- `table` - Table with perspective and legs
- `terminal` - Sci-fi terminal with screen and panel
- `bones` - Skeletal remains with scattered bones
- `plant` - Foliage with leaves and stem
- `pillar` - Stone pillar with weathering and capital
- `torch` - Animated torch with flickering flame
- `debris` - Scattered rubble
- `container` - Futuristic storage container with accent stripes

### Lore Items (SpriteLoreItem)
- `note` - Paper document with text lines
- `audiolog` - Recording device with LED
- `graffiti` - Wall art with geometric patterns
- `body` - Skeletal arrangement

### Destructibles (SpriteDestructible)
- `barrel` - Explosive barrel
- `crate` - Breakable crate

### Pickups (SpritePickup)
- `health` - Health pack with cross symbol
- `ammo` - Ammunition box
- `armor` - Armor plate with shield icon

### Projectiles (SpriteProjectile)
- `bullet` - Generic projectile

## Usage

```go
import "github.com/opd-ai/violence/pkg/sprite"

// Create generator with max 100 cached sprites
gen := sprite.NewGenerator(100)

// Set genre for themed colors
gen.SetGenre("fantasy")

// Get or generate a sprite
// Parameters: type, subtype, seed, frame, size
spriteImg := gen.GetSprite(sprite.SpriteProp, "barrel", 12345, 0, 64)

// Render sprite to screen
op := &ebiten.DrawImageOptions{}
op.GeoM.Translate(x, y)
screen.DrawImage(spriteImg, op)
```

## Performance

- **Cache Hit**: ~20 ns (pointer lookup)
- **Cache Miss**: ~500 μs (generation + caching)
- **Memory**: ~4 KB per 64x64 sprite
- **Typical Cache Size**: 100 sprites = ~400 KB

## Genre Theming

The generator adapts sprite colors to the current genre:

| Genre      | Wood        | Stone       | Foliage      |
|------------|-------------|-------------|--------------|
| Fantasy    | Brown oak   | Gray stone  | Green        |
| Sci-Fi     | Metal gray  | Light metal | Neon green   |
| Horror     | Dark decay  | Dark stone  | Withered     |
| Cyberpunk  | Black synth | Dark metal  | Neon pink    |
| Post-Apoc  | Weathered   | Concrete    | Green        |

## Shading Techniques

1. **Distance-based shading**: Darker on edges, lighter in center
2. **Wood grain**: Sine-wave noise for natural texture
3. **Plank variation**: Alternating rows for wooden crates
4. **Metallic highlights**: Specular highlights on metal bands
5. **Scanlines**: CRT-style scanlines on terminal screens
6. **Noise weathering**: Random noise for stone weathering
7. **Flame animation**: Multi-layer flame with flicker

## Integration

The sprite system integrates with:
- **Game**: Main game loop sets genre and retrieves sprites
- **Props**: Props manager provides prop types and positions
- **Lore**: Lore system provides lore item types
- **Destructibles**: Destructible objects use sprite rendering
- **Animation**: Animation ticker drives frame-based effects

## Testing

Test coverage: **96.9%**

```bash
go test ./pkg/sprite/... -cover
go test ./pkg/sprite/... -bench=.
```

## Future Enhancements

- Equipment overlays for character sprites
- Damage states for destructibles
- Particle effect sprites
- Environment-specific variations (wet, frozen, burning)
- Normal maps for lighting interaction

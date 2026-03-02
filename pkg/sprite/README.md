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

### Enemies (SpriteEnemy)

Enemies use genre-aware body plans with proper shading, animation, and visual variety:

**Humanoid Enemies** (role-based):
- `humanoid` - Generic humanoid enemy
- `tank` - Heavy armored melee fighter with shield
- `ranged` - Long-range combatant with rifle/bow
- `healer` - Support unit with healing symbol
- `ambusher` - Stealth attacker with crouch animation
- `scout` - Fast recon unit

**Non-Humanoid Creatures** (body plan templates):
- `quadruped` - Four-legged beasts with tail and distinct head
- `insect` - Multi-segmented arthropods with 6+ legs and antennae
- `serpent` - Sinuous snake-like creatures with wave motion
- `flying` - Winged creatures with hover animation and tail
- `amorphous` - Blob/slime entities with pulsing animation and multiple eyes

Each enemy type features:
- Genre-specific color palettes (fantasy, scifi, horror, cyberpunk, postapoc)
- Walk cycle animations (alternating leg/appendage movement)
- Attack animations (weapon swing, lunge, or ability casting)
- Proper shading with highlights and shadows
- Distinctive silhouettes for instant recognition
- Idle fidget animations for visual interest

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

// Get humanoid enemy sprite
enemySprite := gen.GetSprite(sprite.SpriteEnemy, "tank", 12345, 0, 64)

// Get quadruped creature sprite
beastSprite := gen.GetSprite(sprite.SpriteEnemy, "quadruped", 54321, 0, 64)

// Get animated serpent (frame-based animation)
for frame := 0; frame < 8; frame++ {
    serpentFrame := gen.GetSprite(sprite.SpriteEnemy, "serpent", 99999, frame, 64)
    // Render frame...
}

// Get prop sprite
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

| Genre      | Wood        | Stone       | Foliage      | Enemy Humanoid | Enemy Creature |
|------------|-------------|-------------|--------------|----------------|----------------|
| Fantasy    | Brown oak   | Gray stone  | Green        | Steel armor    | Brown/Green    |
| Sci-Fi     | Metal gray  | Light metal | Neon green   | Blue armor     | Metallic gray  |
| Horror     | Dark decay  | Dark stone  | Withered     | Dark robes     | Corrupted flesh|
| Cyberpunk  | Black synth | Dark metal  | Neon pink    | Black armor    | Neon mutant    |
| Post-Apoc  | Weathered   | Concrete    | Green        | Scrap armor    | Wasteland brown|

## Shading Techniques

### Enemy Sprites
1. **Body shading**: Gradient shading based on distance from center (cylindrical for limbs)
2. **Armor highlighting**: Metallic sheen on armor plates
3. **Limb animation**: Frame-based leg/arm movement with offset shading
4. **Weapon rendering**: Distinct materials (metal, wood, energy) with proper highlights
5. **Eye glow**: Bright accent colors for creature eyes
6. **Texture variation**: Scales, fur, skin rendered with subtle noise
7. **Shadow casting**: Darker shading on lower body parts for depth
8. **Wing transparency**: Alpha blending for flying creature wings

### Prop Sprites
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

Test coverage: **96.9%** (including comprehensive enemy sprite tests)

```bash
go test ./pkg/sprite/... -cover
go test ./pkg/sprite/... -bench=.
```

### Enemy Sprite Test Coverage
- All humanoid roles (tank, ranged, healer, ambusher, scout)
- All creature body plans (quadruped, insect, serpent, flying, amorphous)
- Animation frame generation and variety
- Genre-specific color palettes
- Seed-based determinism
- Sprite caching and LRU eviction

## Future Enhancements

- ~~Enemy sprite generation with body plan variety~~ ✅ **IMPLEMENTED**
- Equipment overlays for character sprites
- Damage states for destructibles and enemies
- Particle effect sprites
- Environment-specific variations (wet, frozen, burning)
- Normal maps for lighting interaction
- Boss-specific unique sprites with increased detail
- Hybrid creatures (mixing body plan templates)

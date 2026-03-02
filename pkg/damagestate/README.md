# Damage State Visualization System

The damage state system provides genre-specific visual feedback for entity damage, making combat feel more impactful and readable.

## Features

- **Progressive damage visualization**: Entities show increasing visible damage as HP decreases (pristine → light → moderate → critical)
- **Genre-specific aesthetics**:
  - **Fantasy**: Blood wounds and drips
  - **Sci-Fi**: Cracks, sparks, and glitch lines  
  - **Horror**: Gore splatter and decay
  - **Cyberpunk**: Neon glitches and circuit damage
  - **Post-Apocalyptic**: Rust patches and dirt streaks
- **Directional damage**: Wounds appear biased toward the direction of incoming damage
- **Cached rendering**: Damage overlays are cached per damage level and pattern for performance
- **Deterministic generation**: Seeded RNG ensures consistent wound patterns for the same entity

## Usage

### Component

Add a `damagestate.Component` to entities that should show damage:

```go
entity := world.AddEntity()
world.AddComponent(entity, &damagestate.Component{
    CurrentHP:    80,
    MaxHP:        100,
    WoundPattern: rng.Int63(),
    LastDamageX:  hitDirX,
    LastDamageY:  hitDirY,
})
```

### System Update

The system automatically updates damage levels based on HP ratio:

- **Level 0** (Pristine): HP ≥ 75%
- **Level 1** (Light): 50% ≤ HP < 75%
- **Level 2** (Moderate): 25% ≤ HP < 50%
- **Level 3** (Critical): HP < 25%

### Rendering

Apply damage overlay to entity sprites during rendering:

```go
baseSprite := spriteGenerator.GetSprite(...)
if comp, ok := getDamageStateComponent(entity); ok {
    sprite = damageStateSystem.RenderDamageOverlay(comp, baseSprite)
}
```

## Performance

- Overlay generation is deterministic and cached
- Cache invalidation only occurs when damage level changes
- Typical overhead: <1ms per entity per damage level transition
- Recommended for entities with HP systems (players, enemies, destructibles)

## Integration

The system is registered in `main.go` and updates automatically each frame. To use:

1. Add `damagestate.Component` to entities when they spawn
2. Update `CurrentHP` when entity takes damage
3. Set `LastDamageX/Y` to damage direction vector for directional wounds
4. Call `RenderDamageOverlay` when rendering entity sprites

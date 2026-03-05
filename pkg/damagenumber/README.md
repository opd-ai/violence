# Damage Number System

Floating combat text system for visual damage feedback.

## Overview

The damage number system displays floating numbers when entities take damage, providing immediate visual feedback for combat effectiveness. Numbers rise up, scale in, fade out, and are color-coded by damage type.

## Features

- **Animated Text**: Numbers rise upward with smooth scaling and fading
- **Damage Type Colors**: Visual distinction between physical, fire, ice, lightning, poison, etc.
- **Critical Hit Emphasis**: Larger, brighter, longer-lasting numbers with pulsing animation
- **Heal Indication**: Green numbers for healing effects
- **Genre Support**: Color schemes adapt to game genre (fantasy, scifi, cyberpunk, horror)
- **Performance**: Zero per-frame allocations, automatic cleanup after lifetime

## Usage

### Setup

```go
// In game initialization
damageNumberSystem := damagenumber.NewSystem(genreID)
world.AddSystem(damageNumberSystem)
```

### Spawning Damage Numbers

```go
import "github.com/opd-ai/violence/pkg/damagenumber"

// When damage is dealt
damagenumber.Spawn(
    world,
    damage,        // int: damage amount
    targetX,       // float64: world X position
    targetY,       // float64: world Y position
    "physical",    // string: damage type
    isCritical,    // bool: is critical hit?
    false,         // bool: is heal? (set true for healing)
)
```

### Rendering

```go
// In game render loop (after world entities, before UI)
damageNumberSystem.Render(world, screen, cameraX, cameraY)
```

## Damage Types and Colors

| Damage Type | Color | Example |
|------------|-------|---------|
| physical, kinetic, slash, pierce, blunt | Light Red | Sword hit |
| fire, burn, heat | Orange-Red | Fireball |
| ice, cold, frost | Cyan | Ice shard |
| lightning, electric, shock | Yellow | Lightning bolt |
| poison, toxic, acid | Yellow-Green | Poison DOT |
| dark, shadow, void | Purple | Shadow damage |
| holy, light, radiant | Pale Yellow | Holy smite |
| arcane, magic, mystic | Light Purple | Magic missile |
| heal (any type) | Green | Health potion |

## Animation Lifecycle

1. **Spawn (0-20%)**: Number scales from 0.5x to 1.0x
2. **Display (20-70%)**: Number rises at constant velocity, full opacity
3. **Fade (70-100%)**: Number fades to transparent, velocity decays

**Durations**:
- Normal: 1.5 seconds
- Critical: 2.0 seconds
- Heal: 1.5 seconds

**Rise Speed**:
- Normal: 40 units/sec
- Critical: 60 units/sec
- Heal: 30 units/sec

## Critical Hit Visual

Critical hits receive special treatment:
- 33% longer lifetime
- 50% faster rise speed
- Yellow color regardless of damage type
- Pulsing scale animation (±15% at 15 Hz)
- Exclamation mark suffix ("99!")

## Integration Example

```go
// In combat damage handler
func (g *Game) onEntityDamaged(targetEnt engine.Entity, damage int, damageType string, isCrit bool) {
    // Get target position from physics component
    pos := getEntityPosition(g.world, targetEnt)
    
    // Spawn damage number at target location
    damagenumber.Spawn(
        g.world,
        damage,
        pos.X,
        pos.Y - 10.0,  // Offset up slightly from entity center
        damageType,
        isCrit,
        false,
    )
    
    // ... rest of damage logic
}

// In heal handler
func (g *Game) onEntityHealed(targetEnt engine.Entity, healAmount int) {
    pos := getEntityPosition(g.world, targetEnt)
    
    damagenumber.Spawn(
        g.world,
        healAmount,
        pos.X,
        pos.Y - 10.0,
        "",     // Damage type unused for heals
        false,  // Never critical
        true,   // Is heal
    )
}
```

## Performance Characteristics

- **Update**: O(n) where n = active damage numbers (typically < 20)
- **Render**: O(n) text draws with camera culling
- **Memory**: ~200 bytes per active number
- **Allocations**: Zero per-frame (only at spawn)
- **Cleanup**: Automatic when Age >= Lifetime

## Testing

Run tests with:
```bash
go test -v ./pkg/damagenumber/...
```

Coverage target: ≥40% (display-dependent package)

## Architecture Notes

### Component

`Component` stores state for a single damage number:
- Value, damage type, critical/heal flags
- Position (X, Y) and velocity
- Animation state (age, lifetime, scale, alpha)
- Render color

### System

`System` manages animation and rendering:
- `Update(world)`: Animates all damage numbers, removes expired
- `Render(world, screen, camX, camY)`: Draws numbers to screen

### Spawn Function

`Spawn(world, ...)` creates and configures a new damage number entity with appropriate defaults based on damage type and flags.

## Future Enhancements

Potential improvements (not currently implemented):
- Font size scaling based on damage magnitude
- Damage number stacking/merging for rapid hits
- Horizontal spread for multiple numbers at same position
- Floating text for non-damage events (XP gain, level up, item pickup)
- Configurable font faces per genre
- Number outline/shadow for better readability

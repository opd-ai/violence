# Attack Trail System

Visual weapon attack trail rendering for dynamic combat feedback.

## Overview

The attacktrail package creates procedural visual trails that follow weapon attacks. Trails provide immediate visual feedback for slashes, thrusts, smashes, and other attack types, making combat more readable and satisfying.

## Features

- **Six Trail Types**: Slash, thrust, cleave, smash, spin, and projectile trails
- **Automatic Weapon Detection**: Helper functions classify weapons by name
- **Genre-Aware Colors**: Trail colors adapt to fantasy, sci-fi, horror, and cyberpunk themes
- **Smooth Animation**: Trails fade gracefully with configurable lifetimes
- **Performance**: Automatic lifecycle management and trail limiting per entity

## Trail Types

| Type | Visual | Best For |
|------|--------|----------|
| **Slash** | Curved arc | Swords, axes, daggers |
| **Thrust** | Linear pierce | Spears, rapiers, lances |
| **Cleave** | Wide sweeping arc | Greatswords, scythes |
| **Smash** | Radial impact burst | Hammers, maces, clubs |
| **Spin** | Full-circle rotation | Staffs, dual blades |
| **Projectile** | Motion streak | Arrows, thrown weapons |

## Quick Start

### 1. Register the System

```go
import "github.com/opd-ai/violence/pkg/attacktrail"

trailSystem := attacktrail.NewSystem("fantasy")
world.AddSystem(trailSystem)
```

### 2. Add Trail Component to Entities

```go
trailComp := attacktrail.NewTrailComponent(3) // Max 3 trails
world.AddComponent(entityID, trailComp)
```

### 3. Create Trails When Attacks Execute

**Option A: Automatic (Recommended)**

```go
attacktrail.AttachTrailToAttack(
    world, entityID,
    x, y, dirX, dirY, range_,
    "sword", rng, colorFunc,
)
```

**Option B: Manual**

```go
trail := attacktrail.CreateSlashTrail(
    x, y, angle, range, arc, width,
    color.RGBA{R: 200, G: 220, B: 255, A: 180},
)
trailComp.AddTrail(trail)
```

### 4. Render Trails Each Frame

```go
// In your Draw() method
trailSystem.Render(screen, world, cameraX, cameraY)
```

## Integration with Combat

The system is designed to hook into combat events:

```go
// When melee attack hits or swings
func onMeleeAttack(attacker Entity, target Entity, weapon *Weapon) {
    pos := getEntityPosition(attacker)
    dir := getAttackDirection(attacker, target)
    
    attacktrail.AttachTrailToAttack(
        world, attacker,
        pos.X, pos.Y, dir.X, dir.Y,
        weapon.Range, weapon.Type,
        rng, sys.GetWeaponTrailColor,
    )
}
```

## Customization

### Custom Trail Colors

```go
func myColorFunc(weaponName string, rng *rand.Rand) color.RGBA {
    return color.RGBA{R: 255, G: 100, B: 100, A: 200}
}
```

### Custom Trail Parameters

```go
trail := &attacktrail.Trail{
    Type:      attacktrail.TrailSlash,
    StartX:    x, StartY: y,
    Angle:     angle,
    Arc:       math.Pi / 2,
    Range:     100.0,
    Width:     5.0,
    Color:     myColor,
    Intensity: 1.0,
    MaxAge:    0.3,
    FadeStart: 0.1,
}
```

## Performance

- Trails automatically expire and are removed
- Per-entity trail limits prevent overdraw
- Vector rendering with batched draw calls
- Typical cost: <0.5ms for 50 active trails at 1080p

## Genre Styling

The system adapts trail appearance to game genre:

- **Fantasy**: Silver-blue steel trails
- **Sci-fi**: Plasma blue, laser red, energy green
- **Horror**: Dark crimson blood trails
- **Cyberpunk**: Neon magenta, cyan, yellow

Set genre during system creation:

```go
trailSystem := attacktrail.NewSystem("cyberpunk")
```

## Testing

Run tests:

```bash
go test ./pkg/attacktrail
```

Note: Some tests require a display and will be skipped in headless environments.

## Architecture

- **Component**: `TrailComponent` stores active trails per entity
- **System**: `System` updates trail lifecycle and handles rendering
- **Helpers**: Convenience functions for common use cases

## API Reference

See [doc.go](doc.go) for full API documentation.

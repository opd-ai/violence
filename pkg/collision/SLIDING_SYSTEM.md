# Terrain Sliding System

## Overview

The terrain sliding system provides smooth collision response for entities moving along walls and obstacles. Instead of stopping abruptly when hitting terrain, entities slide smoothly along surfaces, creating more fluid and player-friendly movement.

## What Changed

### New Files
- `pkg/collision/sliding.go` - Core sliding system implementation
- `pkg/collision/sliding_test.go` - Comprehensive test suite
- `pkg/collision/sliding_doc.go` - Package documentation
- `pkg/collision/example/main.go` - Usage example

### Modified Files
- `main.go` - Added sliding system field, initialization, and registration

## How It Works

When an entity with velocity would collide with terrain:

1. **Detect Collision**: Check if movement would result in collision
2. **Compute Normal**: Calculate the surface normal at collision point
3. **Project Velocity**: Project velocity onto surface tangent (perpendicular to normal)
4. **Apply Friction**: Reduce slide velocity based on friction coefficient
5. **Move Along Surface**: Apply the slide movement instead of stopping
6. **Iterate**: Repeat up to 4 times per frame to handle corners and complex geometry

This creates smooth movement along walls without abrupt stops or getting stuck.

## Components

### VelocityComponent
Stores entity velocity in units per second.

```go
type VelocityComponent struct {
    X, Y float64  // Velocity in world units/sec
}
```

### PositionComponent
Stores entity position in world coordinates.

```go
type PositionComponent struct {
    X, Y float64  // Position in world coordinates
}
```

### SlidingComponent
Configuration for sliding behavior.

```go
type SlidingComponent struct {
    Enabled           bool    // Toggle sliding on/off
    MaxSlideAngle     float64 // Max angle to allow sliding (radians)
    Friction          float64 // Friction coefficient (0.0-1.0)
    BounceOnSteep     bool    // Bounce off steep surfaces
    BounceRestitution float64 // Bounce elasticity (0.0-1.0)
}
```

### SlidingSystem
Processes entities and applies sliding movement each frame.

## Usage

Add the required components to entities that should slide:

```go
entity := world.AddEntity()

// Position
world.AddComponent(entity, &collision.PositionComponent{X: 100, Y: 100})

// Velocity
world.AddComponent(entity, &collision.VelocityComponent{X: 50, Y: 0})

// Collider
collider := collision.NewCircleCollider(100, 100, 10, 
    collision.LayerPlayer, collision.LayerAll)
world.AddComponent(entity, &collision.ColliderComponent{Collider: collider})

// Sliding configuration
world.AddComponent(entity, collision.NewSlidingComponent())
```

The system automatically processes these entities each frame.

## Configuration

Customize sliding behavior per entity:

```go
sliding := &collision.SlidingComponent{
    Enabled:           true,
    MaxSlideAngle:     math.Pi / 3,  // 60 degrees
    Friction:          0.2,           // 20% friction
    BounceOnSteep:     true,
    BounceRestitution: 0.5,           // 50% bounce
}
```

## Performance

- Uses spatial indexing for O(k) collision queries (k = nearby entities)
- Falls back to O(N) linear search if spatial index unavailable
- Limits to 4 iterations per frame to prevent runaway computation
- Most scenarios resolve in 1-2 iterations

## Integration Points

The sliding system is integrated into the game's ECS architecture:

1. **System Initialization** (`main.go:335`)
   ```go
   g.slidingSystem = collision.NewSlidingSystem(nil)
   ```

2. **Spatial Index Connection** (`main.go:348`)
   ```go
   g.slidingSystem.SetSpatialIndex(g.spatialSystem.GetGrid())
   ```

3. **World Registration** (`main.go:391`)
   ```go
   g.world.AddSystem(g.slidingSystem)
   ```

## Testing

Run the collision package tests:

```bash
go test ./pkg/collision/...
```

Run the example:

```bash
go run ./pkg/collision/example/main.go
```

## Player Experience

Before sliding system:
- Players bump into walls and stop completely
- Diagonal movement feels "sticky" near walls
- Easy to get stuck in corners
- Movement requires precise input

After sliding system:
- Players slide smoothly along walls
- Diagonal movement maintains momentum in valid direction
- Corners are navigated fluidly
- Movement feels responsive and forgiving

## Technical Details

### Collision Layers

The system respects collision layer masks:

```go
// Player slides against terrain and enemies
playerCollider := NewCircleCollider(x, y, r, 
    LayerPlayer, LayerTerrain|LayerEnemy)

// Projectiles ignore allies
projCollider := NewCircleCollider(x, y, r, 
    LayerProjectile, LayerAll^LayerPlayer)
```

### Multi-Iteration Sliding

Complex geometry requires multiple iterations:

```
Frame 1, Iteration 1: Move diagonally toward corner
    -> Collision detected with wall A
    -> Slide along wall A

Frame 1, Iteration 2: Remaining slide movement
    -> Collision detected with wall B  
    -> Slide along wall B

Result: Smooth corner navigation
```

### Friction Model

Friction reduces slide velocity:

```go
slideX, slideY := SlideVector(velocityX, velocityY, normalX, normalY)
frictionFactor := 1.0 - sliding.Friction
slideX *= frictionFactor  // Apply friction
slideY *= frictionFactor
```

Friction values:
- `0.0` = Frictionless (ice)
- `0.1` = Low friction (default, smooth sliding)
- `0.5` = High friction (significant slowdown)
- `1.0` = Full stop (no sliding)

## See Also

- `pkg/collision/collision.go` - Core collision detection
- `pkg/collision/system.go` - Collision ECS integration
- `pkg/spatial/` - Spatial indexing for fast queries
- `pkg/engine/` - ECS world and component management

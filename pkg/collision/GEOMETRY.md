# Collision Geometry Extraction System

## Overview

The Collision Geometry Extraction System provides precise, shape-aware collision detection for Violence. Instead of using coarse axis-aligned bounding boxes (AABBs) for all entities, this system extracts tight collision polygons from sprite data and generates attack-specific hit shapes that match visual representations.

## Architecture

### Core Components

1. **GeometryExtractor** - Extracts collision geometry from sprite alpha channels
2. **AttackShapeCache** - Pre-computed attack patterns for weapons and spells
3. **WeaponShapeGenerator** - Generates standard weapon attack shapes
4. **CollisionGeometrySystem** - Unified system managing all geometry operations

### Component Types

- **SpriteColliderComponent** - Stores polygon colliders extracted from entity sprites
- **AttackFrameComponent** - Stores attack hitboxes per animation frame

## Features

### Sprite-Based Collision Extraction

The system can extract convex hull collision polygons from any sprite:

```go
// Extract a collision polygon from a sprite
collider := collisionGeometry.ExtractSpriteCollider(sprite, x, y, 
    collision.LayerEnemy, collision.LayerPlayer)
```

Features:
- **Convex Hull Generation**: Converts sprite alpha channel to tight polygon
- **Douglas-Peucker Simplification**: Reduces polygon complexity for performance
- **Configurable Alpha Threshold**: Control which pixels are considered solid
- **Automatic Caching**: Avoids regenerating geometry for unchanged sprites

### Pre-Generated Attack Shapes

13 standard weapon/spell attack patterns are pre-generated:

| Shape Name | Type | Range | Arc/Spread | Use Case |
|------------|------|-------|------------|----------|
| sword_slash_h | Arc | 25 | 90° | Horizontal sword swing |
| sword_slash_v | Arc | 25 | 60° | Vertical sword slash |
| sword_thrust | Cone | 30 | 15° | Forward sword stab |
| axe_swing | Arc | 28 | 120° | Wide axe sweep |
| spear_thrust | Rectangle | 40x4 | - | Long spear attack |
| dagger_stab | Cone | 15 | 20° | Quick dagger strike |
| hammer_smash | Arc | 22 | 90° | Overhead hammer |
| whip_sweep | Arc | 45 | 180° | Long-range whip |
| fireball_impact | Square | 20x20 | - | AoE spell impact |
| lightning_beam | Rectangle | 50x6 | - | Beam spell |
| ice_cone | Cone | 35 | 60° | Cone spell |
| cleave | Arc | 32 | 120° | 2H weapon sweep |
| backstab | Cone | 12 | 10° | Precise backstab |

### Attack Shape Generation

Create custom attack shapes on the fly:

```go
// Weapon swing arc (90 degree horizontal slash)
arc := collision.GenerateAttackArc(playerX, playerY, 
    weaponRange, -0.785, 0.785, 10)

// Directional cone (thrust/beam)
cone := collision.GenerateConeShape(x, y, length, direction, spread)

// Rotated rectangle (slash/beam)
rect := collision.GenerateRectangleShape(x, y, width, height, rotation)
```

### Attack Frame Mapping

Link attack hitboxes to specific animation frames:

```go
attackComp := collision.NewAttackFrameComponent(
    "sword_slash_h", // Shape name
    25.0,            // Damage
    10.0,            // Knockback
)

// Set active hitbox for frame 3 (mid-swing)
attackComp.SetActiveFrame(3, hitCollider)
attackComp.UpdateFrame(3) // Sync with animation

// Check if current frame has active hitbox
if collider := attackComp.GetCurrentCollider(); collider != nil {
    // Process hits
}
```

## Integration Guide

### Basic Usage

The CollisionGeometrySystem is automatically initialized in the Game struct:

```go
// Access the system
sys := g.collisionGeometry

// Get a pre-generated attack shape
shape := sys.GetAttackShape("sword_slash_h")

// Create an attack collider at runtime
collider := sys.CreateAttackCollider(
    "axe_swing",           // Shape name
    playerX, playerY,      // Position
    dirX, dirY,            // Direction
    collision.LayerPlayer, // Layer
    collision.LayerEnemy,  // Mask
)
```

### Custom Attack Shapes

Register custom attack patterns:

```go
cache := sys.GetShapeCache()

customShape := &collision.AttackShape{
    Name: "custom_attack",
    Vertices: collision.GenerateAttackArc(0, 0, 35, -1.0, 1.0, 16),
}

cache.RegisterShape("custom_attack", customShape)
```

### Sprite Collision Updates

For entities with procedurally generated or changing sprites:

```go
comp := &collision.SpriteColliderComponent{
    Dirty: true,
}

// Update collision geometry when sprite changes
sys.UpdateSpriteCollider(
    comp,
    sprite,           // Current sprite image
    "entity_sprite_1", // Cache key
    entityX, entityY,  // Position
    collision.LayerEnemy,
    collision.LayerPlayer,
)

// Use the extracted colliders
broadphase := comp.BoundingBox  // AABB for quick rejection
narrowphase := comp.DetailedHull // Polygon for precise collision
```

## Performance Characteristics

### Sprite Extraction
- **Time Complexity**: O(n) where n = sprite pixel count
- **Convex Hull**: O(m log m) where m = solid pixel count
- **Simplification**: O(p²) where p = hull vertex count
- **Benchmark**: ~50-100µs for 32x32 sprite with caching

### Attack Shape Generation
- **Arc**: O(segments) - typically 8-16 segments
- **Cone**: O(1) - always 3 vertices
- **Rectangle**: O(1) - always 4 vertices
- **Benchmark**: <1µs per shape

### Collision Detection
- Uses existing collision.TestCollision() with polygon support
- SAT (Separating Axis Theorem) for polygon-polygon
- Optimized edge checks for polygon-circle
- Spatial partitioning compatible for broadphase

## Configuration

### Alpha Threshold

Control which sprite pixels are considered solid:

```go
extractor := sys.GetExtractor()
extractor.SetAlphaThreshold(200) // 0-255, default 128
```

### Simplification Tolerance

Control polygon simplification (lower = more detail, higher = fewer vertices):

```go
extractor.SetSimplifyEpsilon(3.0) // pixels, default 2.0
```

## Best Practices

### Entity Collision

1. **Static Entities**: Extract once, cache polygon
2. **Animated Entities**: Extract per-frame or use bounding box
3. **Small Sprites**: Use circle colliders (faster)
4. **Large/Irregular**: Use extracted polygons

### Attack Patterns

1. **Melee Weapons**: Use pre-generated arcs/cones
2. **Projectiles**: Use capsules for swept collision
3. **Spells**: Combine shapes (e.g., ring = outer - inner circle)
4. **AoE**: Start with circle, upgrade to polygon for complex shapes

### Performance

1. **Broadphase**: Always use bounding circles/AABBs first
2. **Cache**: Reuse extracted geometry, don't regenerate per frame
3. **Simplify**: Use higher epsilon for distant/small entities
4. **LOD**: Switch to simple shapes when entity is far from camera

## Example: Complete Attack Flow

```go
// 1. Get attack shape on weapon equip
shape := g.collisionGeometry.GetAttackShape("sword_slash_h")

// 2. Create attack frame component
attackFrames := collision.NewAttackFrameComponent("sword_slash_h", 25, 10)

// 3. Define active frames (frames 2-4 of 8-frame attack animation)
for frame := 2; frame <= 4; frame++ {
    collider := g.collisionGeometry.CreateAttackCollider(
        "sword_slash_h",
        g.camera.X, g.camera.Y,
        g.camera.DirX, g.camera.DirY,
        collision.LayerPlayer,
        collision.LayerEnemy,
    )
    attackFrames.SetActiveFrame(frame, collider)
}

// 4. During animation update
attackFrames.UpdateFrame(currentAnimFrame)

// 5. Check for hits
if hitbox := attackFrames.GetCurrentCollider(); hitbox != nil {
    hits := collision.QueryCollisions(g.world, hitbox)
    for _, enemy := range hits {
        // Apply damage, knockback, etc.
    }
}
```

## Technical Details

### Convex Hull Algorithm

Graham scan with polar angle sorting:
1. Find bottommost point
2. Sort remaining points by polar angle
3. Build hull with left-turn elimination
4. Time: O(n log n)

### Douglas-Peucker Simplification

Recursive line simplification:
1. Find point with maximum perpendicular distance
2. If distance > epsilon, split and recurse
3. Otherwise, replace segment with line
4. Preserves shape while reducing vertices

### Collision Layers

Standard layers for combat:
- **Player**: Player-controlled entities
- **Enemy**: AI-controlled hostiles
- **Projectile**: Bullets, arrows, fireballs
- **Terrain**: Walls, obstacles (static)
- **Environment**: Props, destructibles
- **Ethereal**: Pass-through entities
- **Interactive**: Doors, chests
- **Trigger**: Detection zones

## Future Enhancements

Potential extensions:
- [ ] Bitmap collision masks for pixel-perfect detection
- [ ] Oriented bounding boxes (OBB) for rotated AABBs
- [ ] Multi-part hitboxes for complex entities (body, head, limbs)
- [ ] Temporal collision (predict future collisions)
- [ ] Collision event callbacks
- [ ] Debug visualization overlay

## References

- SAT Algorithm: Separating Axis Theorem for convex polygon collision
- Graham Scan: Convex hull in 2D computational geometry
- Douglas-Peucker: Line simplification algorithm
- Capsule Collision: Swept sphere for moving objects

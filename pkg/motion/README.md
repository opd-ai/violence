# Organic Motion System

The organic motion system adds realistic, lifelike animation to Violence through:

1. **Acceleration/Deceleration Curves (Easing)** - Smooth velocity transitions instead of instant speed changes
2. **Squash and Stretch** - Entities deform on impacts, adding weight and physicality
3. **Secondary Motion** - Trailing elements (tails, cloaks, hair) lag behind with realistic follow-through
4. **Breathing Idle Animation** - Subtle vertical oscillation when stationary to prevent "frozen statue" appearance
5. **Weight-Based Movement** - Entity mass affects acceleration rates for distinct movement personalities

## Quick Start

```go
// Add motion component to any entity that should have organic movement
motionComp := motion.InitializeMotion(mass, hasTrail)
world.AddComponent(entity, motionComp)

// System automatically updates all motion components every frame
motionSystem.Update(world)
```

## Usage in Rendering

Rendering systems should apply motion transforms when drawing sprites:

```go
helper := motion.NewRenderHelper(motionSystem)

// Apply squash/stretch and breath offset
renderY, opts := helper.ApplyMotionTransform(motionComp, x, y, drawOpts)

// Draw with adjusted position and scale
screen.DrawImage(sprite, opts)
```

For entities with trails (tails, cloaks, etc.):

```go
positions := helper.GetTrailPositions(motionComp)
for i, pos := range positions {
    // Draw trail segment at pos.X, pos.Y
    // Later segments should be smaller/fainter
}
```

## Triggering Impact Effects

Call `TriggerImpact` on collision or landing:

```go
velocityMagnitude := math.Sqrt(vx*vx + vy*vy)
motionSystem.TriggerImpact(motionComp, velocityMagnitude)
```

The entity will squash vertically and stretch horizontally, then recover over ~0.15 seconds.

## Configuration

Motion parameters can be tuned per-entity:

```go
motion := &motion.Component{
    EaseRate:        8.0,  // Higher = faster acceleration (1-10)
    Mass:            3.0,  // Entity weight (affects acceleration)
    TrailLength:     5,    // Number of trailing segments
    TrailStiffness:  0.3,  // 0=loose, 1=rigid
    BreathFrequency: 0.2,  // Breaths per second
    BreathAmplitude: 0.5,  // Pixel offset range
}
```

## Performance

The system is designed for real-time 60 FPS performance:
- No allocations in hot paths
- Simple math operations (no complex physics)
- Pre-allocated trail segments
- Purely visual (doesn't affect gameplay)

Tested with 100+ animated entities maintaining 60+ FPS.

## Visual Realism Improvements

This system addresses the "LIFELESS ANIMATION" visual realism problem by:

1. **Eliminating robotic movement** - Easing curves make all motion feel natural and responsive
2. **Adding weight perception** - Squash/stretch and mass-based acceleration convey entity physicality
3. **Creating organic flow** - Secondary motion adds follow-through and drag
4. **Breaking static poses** - Breathing animation makes idle entities feel alive
5. **Differentiating entity types** - Mass variation creates distinct movement personalities

Before: Entities snap to full speed instantly, stop abruptly, and freeze completely when idle.
After: Smooth acceleration, natural deceleration, visible weight, and subtle life even when stationary.

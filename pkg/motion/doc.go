// Package motion provides organic, realistic animation through easing curves,
// squash-and-stretch, secondary motion, breathing idle, and weight-based movement.
//
// This system addresses the "LIFELESS ANIMATION" visual realism problem by adding:
//
// 1. **Acceleration/Deceleration Curves (Easing)**: Entities don't instantly reach
// full speed. Movement uses smooth ease-in-out curves that make motion feel natural
// and responsive rather than robotic. Heavier entities (higher mass) accelerate
// more slowly.
//
// 2. **Squash and Stretch**: On impacts, collisions, and landings, entities deform
// briefly - compressing vertically and expanding horizontally. This classic animation
// principle adds weight and physicality. The effect automatically recovers over ~0.15
// seconds.
//
// 3. **Secondary Motion**: Trailing elements (tails, cloaks, hair, chains) lag behind
// the main entity position, creating follow-through and drag. Each segment follows
// the one ahead with configurable stiffness. Loose trails (low stiffness) create
// fluid, organic movement.
//
// 4. **Breathing Idle Animation**: Stationary entities subtly move up and down in a
// sine wave pattern, simulating breathing. This prevents the "frozen statue" look
// and makes idle creatures feel alive. Frequency and amplitude are configurable.
//
// 5. **Weight-Appropriate Movement**: Entity mass affects acceleration rates. Heavy
// creatures can't change direction instantly. Light creatures respond quickly.
// This creates distinct movement personalities for different entity types.
//
// Usage:
//
// Add a motion.Component to any entity that should have organic movement:
//
//	motionComp := motion.InitializeMotion(mass, hasTrail)
//	world.AddComponent(entity, motionComp)
//
// The system automatically applies easing to entities with both motion.Component
// and engine.Velocity. It updates breathing animation, recovers squash/stretch,
// and updates trailing elements every frame.
//
// Rendering systems should use motion.RenderHelper to apply squash/stretch scaling
// and breath offset when drawing sprites:
//
//	helper := motion.NewRenderHelper(motionSystem)
//	renderY, opts := helper.ApplyMotionTransform(motionComp, x, y, drawOpts)
//
// To trigger impact effects (e.g., on landing or collision):
//
//	motionSystem.TriggerImpact(motionComp, velocityMagnitude)
//
// Performance:
//
// The system is designed for real-time use at 60 FPS:
// - No allocations in hot paths (Update loop)
// - Trail segments pre-allocated on initialization
// - Simple math operations (no complex physics)
// - All effects are purely visual (don't affect gameplay logic)
//
// The system maintains 60+ FPS even with hundreds of animated entities.
package motion

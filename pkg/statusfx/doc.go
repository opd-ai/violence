// Package statusfx provides visual effects for status conditions.
//
// This system renders glowing auras and particles around entities with active
// status effects (poison, burning, stunned, regeneration, etc.). Each effect
// type has a unique color defined by the status effect registry, and the
// system creates pulsating glows and periodic particle emissions to make
// status conditions immediately visible during combat.
//
// # Integration
//
// The system must be registered with the ECS World and called from the rendering
// pipeline:
//
// statusFXSystem := statusfx.NewSystem("fantasy", particleSystem)
// world.AddSystem(statusFXSystem)
//
// In the rendering loop:
//
// statusFXSystem.Render(screen, world, cameraX, cameraY)
//
// # Visual Behavior
//
// - Auras pulse at 3 Hz with 50% intensity variation
// - Particle emission rates vary by effect type (burning is fastest)
// - Effect intensity fades as remaining duration decreases
// - Distance-based alpha fade for entities far from camera
// - Multi-ring aura rendering for depth and visual impact
//
// # Performance
//
// The system only renders effects for entities within 20 tiles of the camera.
// Particle emission is throttled by effect type. Rendering cost scales with
// the number of entities with active status effects.
package statusfx

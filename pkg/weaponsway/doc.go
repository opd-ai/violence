// Package weaponsway provides first-person weapon sway animation for realistic FPS feel.
//
// Weapon sway simulates the inertia, weight, and organic movement of held weapons,
// making the first-person view feel alive and responsive rather than robotically
// locked to the camera. The system applies three types of motion:
//
//   - Turn sway: When the camera rotates, the weapon lags behind momentarily,
//     creating a sense of inertia. This makes aiming feel smooth and prevents
//     the weapon from feeling welded to the screen.
//
//   - Movement bob: When walking or running, the weapon gently bobs up and down
//     and side-to-side, simulating the natural motion of carrying a weapon.
//     Sprinting increases bob amplitude.
//
//   - Breath sway: When idle, subtle breathing motion keeps the weapon alive.
//     This prevents the static feeling of a frozen weapon when standing still.
//
// # Genre Variations
//
// Sway parameters adapt to the game's active genre:
//
//   - Fantasy: Heavy, weighty weapons with dramatic sway
//   - Sci-Fi: Lighter, more responsive futuristic weapons
//   - Horror: Shaky, nervous handling with exaggerated sway
//   - Cyberpunk: Snappy, responsive with quick recovery
//   - Post-Apocalyptic: Rough, improvised weapons with irregular sway
//
// # Integration
//
// The system integrates with the crosshair by syncing sway offset to the
// CrosshairComponent. The sway offset is applied to the crosshair rendering
// position, making the aiming reticle follow the weapon movement.
//
// # Component
//
// Add a Component to any entity that should have weapon sway (typically
// the player). The component tracks current sway state, velocity, and
// animation phases.
//
// # System
//
// The System updates all weapon sway components each frame. It must be
// registered with the ECS World and receives input via AddTurnImpulse
// when the camera rotates and SetMovementState when movement changes.
//
// # Physics Model
//
// The sway uses a spring-damper physics model:
//   - Target offset is set by turn impulses and movement/breath phases
//   - Spring stiffness pulls the weapon toward the target
//   - Damping prevents oscillation and creates smooth motion
//   - Recovery speed controls how fast the target decays to center
//
// This creates physically-motivated motion that feels natural.
package weaponsway

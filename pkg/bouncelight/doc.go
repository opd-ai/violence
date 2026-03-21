// Package bouncelight provides indirect illumination simulation for visual realism.
//
// Bounce lighting simulates how light reflects off surfaces and tints nearby areas.
// When light hits a colored surface (like a red wall), some of that color "bounces"
// onto adjacent surfaces, adding warmth and visual coherence to the scene.
//
// This is a computationally inexpensive approximation of global illumination that
// dramatically improves the realism of procedurally generated environments without
// requiring expensive ray tracing.
//
// # Key Features
//
//   - Surface-to-surface color bleeding based on material color
//   - Distance-based falloff using inverse-square law
//   - Genre-specific bounce intensity (horror = more, scifi = less)
//   - Light source color contribution to nearby surfaces
//   - Deterministic results based on room/level geometry
//   - Zero-allocation per-frame updates via pre-computed bounce maps
//
// # Visual Impact
//
// Without bounce lighting, a torch next to a red brick wall casts orange light on
// the floor, but the wall itself doesn't contribute any color to the scene. With
// bounce lighting enabled, the red wall subtly tints the floor with warm red tones,
// creating a cohesive, believable environment.
//
// # Integration
//
// The system processes wall and floor tiles to calculate bounce contributions,
// then provides tint values that the renderer applies during final composition.
// Register BounceSystem with the ECS World and call ApplyBounce() during rendering.
//
// # Performance
//
// Bounce maps are pre-computed per room/sector and cached. Per-frame cost is O(n)
// where n is visible tiles, with no allocations on the hot path.
package bouncelight

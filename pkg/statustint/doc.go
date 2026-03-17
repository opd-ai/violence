// Package statustint applies material-appropriate visual tinting to sprites
// based on active status effects.
//
// This enhancement addresses the KNOWN VISUAL REALISM PROBLEM:
// "Status effects rendered as material changes (frozen = blue tint + frost particles,
// burning = orange glow + smoke)."
//
// Instead of just rendering auras around entities (handled by statusfx), this system
// modifies the actual sprite colors to reflect the entity's afflicted state:
//
//   - Frozen/Slowed: Blue tint with ice crystallization, desaturation, frost edge overlay
//   - Burning: Orange/red warmth tint, brightness pulsing, ember glow at extremities
//   - Poisoned: Green tint with vein-like darkening, sickly pallor
//   - Bleeding: Red tint concentrating at wound points, dark blood overlay
//   - Irradiated: Green glow with brightness fluctuation, pixel noise
//   - Stunned/EMP: Desaturation with occasional flash frames
//   - Blessed/Buffed: Golden rim lighting, subtle brightness increase
//   - Cursed/Debuffed: Purple/dark tint, shadow encroachment
//
// The tinting is applied additively based on effect intensity, allowing multiple
// status effects to visually stack. Dominant effects (highest intensity) take
// priority for conflicting tints.
//
// Technical implementation:
//   - TintComponent stores computed tint parameters per entity
//   - System queries entities with StatusComponent and computes aggregate tint
//   - Tint parameters are exposed for sprite rendering pipeline to apply
//   - Uses color blending with intensity-weighted mixing
//   - Genre-aware: tint colors vary by game genre (fantasy, scifi, etc.)
//
// Integration:
//   - Register TintSystem in main.go system initialization
//   - Rendering code reads TintComponent to apply color modification
//   - Works alongside existing statusfx (auras) for layered visual feedback
package statustint

// Package decal provides persistent combat visual marks.
//
// The decal system creates long-lasting visual feedback from combat interactions.
// Blood splatters, scorch marks, slash trails, bullet holes, and magical burns
// persist on floors and walls, building environmental storytelling and reinforcing
// player actions.
//
// Features:
// - Genre-aware visual styles (fantasy blood, sci-fi energy, horror gore, cyberpunk neon)
// - Multiple decal types with procedural variation
// - Gradual opacity fade over time
// - LRU-cached procedural generation
// - Spatial culling for efficient rendering
//
// Integration:
// The system hooks into combat events to spawn decals automatically on hits and deaths.
// Decals are stored in game state and rendered as an overlay layer.
package decal

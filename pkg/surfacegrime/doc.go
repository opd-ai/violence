// Package surfacegrime provides procedural dirt, dust, and grime accumulation
// for realistic environmental weathering.
//
// This system adds visible grime textures that naturally accumulate in corners,
// along wall edges, and at floor/wall contact points. Unlike ambient occlusion
// which only darkens, surface grime adds actual colored texture representing
// dirt, dust, soot, mold, and other deposits that make environments look lived-in.
//
// # Visual Realism
//
// The grime system addresses the "artificial environments" problem by adding:
//   - Corner accumulation: Dirt piles up where walls meet floors
//   - Edge creep: Grime spreads along wall bases and ceiling joints
//   - Material-specific deposits: Stone gets mossy, metal gets sooty, wood gets moldy
//   - Genre-appropriate coloring: Fantasy gets natural earth tones, cyberpunk gets urban grime
//   - Procedural noise: Seed-based variation prevents repetitive patterns
//
// # Integration
//
// The system integrates with the rendering pipeline as a post-process layer
// applied to the framebuffer after wall/floor rendering but before sprites.
// It reads the edge AO map to identify accumulation zones and applies
// colored grime overlays with alpha blending.
//
// # Genre Presets
//
//   - Fantasy: Brown dirt, green moss, grey dust, cobwebs in corners
//   - Sci-Fi: Oil stains, carbon scoring, coolant leaks, minimal dirt
//   - Horror: Mold growth, dark stains, organic ooze, blood residue
//   - Cyberpunk: Urban grime, oil slicks, neon-tinted dust, rust
//   - Post-Apocalyptic: Heavy dust, ash, rust everywhere, debris
//
// # Performance
//
// The grime overlay is cached per-room and only regenerated when room geometry
// changes. The overlay uses 8-bit alpha for memory efficiency and applies via
// a single blended draw call per frame.
package surfacegrime

// Package lensdirt provides procedural lens dirt and smudge effects that
// appear when bright light sources are in view. This simulates camera optics
// imperfections like dust particles, fingerprints, and moisture that scatter
// light and create visible artifacts around bright sources.
//
// # Visual Effect
//
// Lens dirt manifests as radial smudges, dust specks, and diffused light
// patterns that become visible when bright light sources (torches, explosions,
// magic effects, sun) are on screen. The effect intensity scales with the
// brightness and proximity of light sources.
//
// # Genre Presets
//
// Different genres have distinct lens dirt characteristics:
//
//   - Fantasy: Warm amber dirt with organic smudge patterns, simulating
//     dusty dungeon environments and oil-lamp covered lenses.
//   - Sci-Fi: Clean lens with minimal blue-tinted dust, simulating
//     maintained camera equipment in sterile environments.
//   - Horror: Heavy dirt with green-tinted grime, simulating abandoned
//     equipment and decayed conditions.
//   - Cyberpunk: Neon-tinted streaks and rain residue, simulating
//     urban moisture and pollution on camera optics.
//   - Post-Apocalyptic: Heavy dust accumulation with sepia tint,
//     simulating sandstorm damage and neglected equipment.
//
// # Integration
//
// The system should be rendered as a screen-space overlay after all world
// rendering but before UI. It samples light source positions and creates
// localized dirt artifacts based on light intensity and screen position.
//
// Example usage:
//
//	lensDirtSys := lensdirt.NewSystem("fantasy", seed, screenW, screenH)
//	// During render loop:
//	lensDirtSys.SetLightSources(lights)
//	lensDirtSys.Render(screen)
package lensdirt

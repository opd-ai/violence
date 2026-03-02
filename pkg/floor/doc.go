// Package floor provides procedural floor tile variation and environmental detail
// for dungeons and procedurally generated levels.
//
// # Overview
//
// The floor package adds visual variety to floor tiles through procedural overlays
// including cracks, stains, debris, scorch marks, and genre-specific details like
// blood splatters (horror), graffiti (cyberpunk), and rust (scifi/postapoc).
//
// # Usage
//
// Create a floor detail system and generate details for a level:
//
//	sys := floor.NewSystem("fantasy", 32)
//	details := sys.GenerateFloorDetails(tiles, seed)
//
// Render detail overlays during level rendering:
//
//	for _, detail := range details {
//	    img := sys.RenderDetail(detail)
//	    // Draw img at tile position with blend mode
//	}
//
// # Detail Types
//
// - Crack: Floor cracks with branching patterns
// - Stain: Liquid stains with irregular edges
// - Debris: Small scattered debris pieces
// - Scorch: Burn marks with gradient falloff
// - Wear: Wear patterns from foot traffic
// - Graffiti: Simple marks and symbols (cyberpunk)
// - Blood: Blood splatter with droplets (horror)
// - Rust: Rust patterns (scifi/postapoc)
// - Corrode: Corrosion marks (scifi)
//
// # Genre-Specific Generation
//
// Each genre has appropriate detail types and densities:
//
// - Fantasy: Cracks, stains, wear, debris (density: 0.15)
// - Scifi: Cracks, scorch, wear, debris, corrode (density: 0.20)
// - Horror: Cracks, stains, blood, debris, wear (density: 0.25)
// - Cyberpunk: Scorch, stains, graffiti, wear, rust (density: 0.22)
// - Postapoc: All types except graffiti (density: 0.30)
//
// # Performance
//
// - Details are generated once per level using seeded RNG
// - Rendered sprites are cached with LRU eviction (max 500 entries)
// - Detail placement respects wall proximity for realistic variation
// - All rendering is deterministic based on seed
package floor

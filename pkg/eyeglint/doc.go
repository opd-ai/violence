// Package eyeglint provides procedural eye highlight rendering for creature sprites.
//
// Eyes without highlights look dead and flat. This package adds small, bright
// specular highlights to creature eyes that simulate the wet corneal surface
// catching light. The highlights:
//
//   - Are positioned consistently with the global light direction (top-left)
//   - Use size-appropriate highlight placement (larger eyes = larger offset)
//   - Include a subtle secondary reflection for added depth
//   - Animate slightly to create a living, watchful appearance
//
// Visual Effect:
//
// Before: Eyes are flat colored circles that look painted on
// After:  Eyes have wet highlights that make creatures appear alive and aware
//
// The system detects eye regions in sprites and adds highlights automatically.
// Eye detection uses color analysis (looking for typical eye colors) and
// positional heuristics (eyes are typically in upper portion of head region).
//
// Genre Presets:
//
//   - Fantasy: Warm golden highlights, mystical glow for magical creatures
//   - Sci-Fi: Sharp white highlights, occasional lens flare for cyborgs
//   - Horror: Dim sickly highlights, occasional red reflection
//   - Cyberpunk: Neon-tinted highlights, chrome reflection for augmented
//   - Post-Apocalyptic: Muted highlights, clouded appearance for mutations
//
// Integration:
//
//	sys := eyeglint.NewSystem("fantasy")
//	// After sprite generation, apply eye highlights:
//	enhanced := sys.ApplyEyeGlints(sprite, spriteType, seed)
//
// The system is lightweight and caches highlight patterns for performance.
package eyeglint

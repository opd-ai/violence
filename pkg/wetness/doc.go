// Package wetness provides surface wetness rendering for enhanced environmental realism.
//
// This package creates procedural puddles, wet surface highlights, and reflective
// water effects that add visual depth to floors and terrain. Wet surfaces are
// rendered with darkened base materials, specular highlights that catch light sources,
// and subtle distortion effects for deeper puddles.
//
// Visual Features:
//   - Procedural puddle shapes with organic, noise-based edges
//   - Wet surface darkening (15-25% darker than dry material)
//   - Specular highlights from light sources on wet surfaces
//   - Depth-based reflection intensity for puddles
//   - Genre-specific moisture density and color tinting
//
// Genre Presets:
//   - Fantasy: Moderate moisture, stone-colored puddles near fountains/drains
//   - Sci-Fi: Minimal moisture, oil-slick rainbow reflections on metal
//   - Horror: Heavy moisture, dark murky puddles with organic tint
//   - Cyberpunk: Rain-soaked streets, neon-reflecting puddles
//   - Post-Apoc: Strategic puddles, rust-tinged water, contaminated look
//
// Usage:
//
//	sys := wetness.NewSystem("cyberpunk", screenW, screenH)
//
//	// Generate wetness pattern for level
//	pattern := sys.GenerateWetnessPattern(tiles, seed)
//
//	// In render loop, after floor but before entities
//	sys.RenderWetness(screen, pattern, lights, cameraX, cameraY)
//
// The system integrates with the lighting system to create accurate specular
// reflections from point lights, torches, and environmental light sources.
package wetness

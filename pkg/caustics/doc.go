// Package caustics provides animated water caustics lighting for environmental realism.
//
// Caustics are the animated light patterns that appear on surfaces near water
// as light refracts through the water surface. This system generates and renders
// these patterns to add visual realism to wet environments.
//
// Features:
//   - Animated Voronoi-based caustic patterns that simulate light refraction
//   - Genre-specific presets (bright fantasy pools, murky horror swamps, neon cyberpunk puddles)
//   - Integration with the wetness system to place caustics near water sources
//   - Performance-optimized with pattern caching and LRU eviction
//   - Seed-based deterministic generation for reproducibility
//
// The system works by detecting nearby water sources (puddles, pools) and projecting
// animated caustic light patterns onto adjacent floor and wall surfaces. The animation
// uses multi-frequency wave functions to create organic, realistic light dancing.
//
// Usage:
//
//	sys := caustics.NewSystem("fantasy", screenWidth, screenHeight)
//	sys.GenerateCausticSources(wetnessPattern, seed)
//	// In render loop:
//	sys.Update(world) // Advance animation
//	sys.Render(screen, cameraX, cameraY)
//
// Genre effects:
//   - Fantasy: Bright blue-white caustics, slow gentle movement
//   - Sci-Fi: Sharp cyan patterns, fast precise movement
//   - Horror: Murky green-brown caustics, erratic unsettling motion
//   - Cyberpunk: Neon-tinted caustics that pick up nearby light colors
//   - Post-Apocalyptic: Dim yellow-brown patterns, sluggish irradiated water
package caustics

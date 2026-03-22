// Package volumetric provides volumetric light rendering for atmospheric light shafts.
//
// Volumetric lighting simulates visible light beams passing through dust particles,
// smoke, and atmospheric haze. This creates dramatic god-rays and light shafts that
// greatly enhance visual realism and atmosphere.
//
// The system supports:
//   - Radial light shafts emanating from point lights (torches, lamps)
//   - Directional god-rays from windows, skylights, or portals
//   - Genre-specific dust density and scatter color
//   - Performance-optimized ray marching with configurable sample count
//   - Shadow-aware occlusion (light rays blocked by walls)
//
// Genre Presets:
//   - Fantasy: Warm golden shafts from torches, heavy dust in dungeons
//   - Sci-Fi: Cool blue rays from terminals, minimal atmospheric dust
//   - Horror: Sickly pale rays, thick fog-like scatter
//   - Cyberpunk: Neon-tinted rays, moderate haze from pollution
//   - Post-Apocalyptic: Dusty orange/yellow shafts, heavy particle scatter
//
// Usage:
//
//	sys := volumetric.NewSystem(genreID, screenWidth, screenHeight)
//	sys.AddLightShaft(x, y, radius, intensity, colorR, colorG, colorB)
//	sys.Render(screen, cameraX, cameraY, dirX, dirY, lightSources)
//
// The system integrates with the existing lighting infrastructure and respects
// wall occlusion data to create realistic shadow cutoffs in light beams.
package volumetric

// Package heatdistort provides screen-space heat distortion effects near fire sources.
//
// The heat distortion system simulates the visual warping effect caused by hot air
// rising from fire sources, lava, explosions, and other heat-generating entities.
// This creates a realistic shimmer effect that makes fire sources appear to radiate
// heat and adds significant depth and realism to the game's visual presentation.
//
// # Technical Approach
//
// The system works by tracking heat source positions in screen space and applying
// a sinusoidal displacement to nearby pixels during post-processing. The displacement
// creates a rippling/warping effect that:
//   - Increases in intensity closer to the heat source
//   - Animates over time with a wave-like pattern
//   - Falls off smoothly with distance from the source
//   - Uses vertical bias (heat rises) while including horizontal ripples
//
// # Genre Presets
//
// Each genre has different heat distortion characteristics:
//   - Fantasy: Moderate distortion from torches and braziers, warm orange tint
//   - Sci-Fi: Subtle distortion from plasma vents and energy sources, blue shift
//   - Horror: Heavy distortion with erratic timing for unsettling effect
//   - Cyberpunk: Sharp, neon-tinted distortion from overheated tech
//   - Post-Apocalypse: Intense distortion from fires and radioactive sources
//
// # Integration
//
// The system is integrated as a post-processing step in the render pipeline.
// Heat sources are registered via the component system, and the distortion is
// applied after the main render but before final post-processing effects.
//
// # Performance
//
// The distortion effect is optimized to only process pixels within the heat
// influence radius. A spatial hash is used to quickly identify which heat
// sources affect which screen regions, and the sine calculations are cached
// per-frame to avoid redundant computation.
package heatdistort

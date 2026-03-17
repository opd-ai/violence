// Package flicker provides realistic flame/torch flickering with physics-based noise patterns.
//
// This system dramatically improves the visual realism of fire-based light sources
// (torches, braziers, candles, fire barrels) by replacing uniform random flicker
// with multi-frequency noise patterns that model actual fire behavior.
//
// Physics-based flame behavior:
//   - Low-frequency sway (0.3-0.8 Hz): The main flame body moves slowly with air currents
//   - Mid-frequency oscillation (2-5 Hz): Rapid brightness fluctuations as fuel burns
//   - High-frequency turbulence (15-30 Hz): Micro-variations creating visual shimmer
//   - Guttering events: Occasional sudden dips when air currents disrupt combustion
//   - Color temperature variation: Intensity affects color (hotter = bluer-white, cooler = redder)
//
// Genre presets:
//   - Fantasy: Strong flicker for torches/braziers/candles with warm orange tones
//   - Horror: Unstable, guttering flames with frequent near-extinguish events
//   - PostApoc: Unreliable oil lamps and fire barrels with sputtering behavior
//   - SciFi: Minimal flicker (holographic/stable power) but alarm lights strobe
//   - Cyberpunk: Neon doesn't flicker, but damaged equipment shows erratic patterns
//
// Integration:
//
//	The FlickerSystem is registered in the main game loop and processes all
//	light sources with flicker enabled. It modifies intensity and color
//	components each frame, with deterministic output based on seed + tick.
//
// Usage:
//
//	sys := flicker.NewSystem("fantasy")
//	params := sys.GetFlickerParams("torch", seed)
//	intensity, r, g, b := sys.CalculateFlicker(params, tick)
package flicker

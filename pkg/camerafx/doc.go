// Package camerafx provides camera effects for enhanced visual feedback.
//
// The camera effects system adds genre-aware screen shake, colored flashes,
// zoom effects, chromatic aberration, and vignette to improve combat feedback
// and player immersion. All effects automatically decay over time and scale
// based on the selected genre for authentic atmosphere.
//
// # Features
//
//   - Screen shake: Random directional shake with genre-specific intensity scaling
//   - Colored flash: Full-screen color flash with customizable RGBA values
//   - Zoom: Smooth camera zoom with interpolation to target level
//   - Chromatic aberration: RGB channel separation effect for impact moments
//   - Vignette: Genre-specific edge darkening for atmosphere
//
// # Usage
//
// Basic setup and triggering effects:
//
//	sys := camerafx.NewSystem("fantasy", 12345)
//
//	// Trigger screen shake on player hit
//	sys.TriggerShake(camerafx.Shake.Medium())
//
//	// Trigger colored flash on damage
//	r, g, b, a := camerafx.Flash.Red()
//	sys.TriggerFlash(r, g, b, a)
//
//	// Add chromatic aberration for heavy hits
//	sys.TriggerChromatic(0.5)
//
//	// Update every frame
//	sys.Update(deltaTime)
//
//	// Get current effects for rendering
//	shakeX, shakeY := sys.GetShakeOffset()
//	flashR, flashG, flashB, flashA := sys.GetFlashColor()
//	zoom := sys.GetZoom()
//
// # Genre Presets
//
// The system applies different effect intensities based on genre:
//
//   - Fantasy: Balanced shake, warm flashes, moderate effects
//   - Sci-Fi: Reduced shake, bright flashes, strong chromatic aberration
//   - Horror: Heavy shake, dim flashes, strong vignette for dread
//   - Cyberpunk: Sharp shake, neon flashes, extreme chromatic effects
//   - Post-Apocalyptic: Medium shake, desaturated effects, dusty atmosphere
//
// # Presets
//
// Convenient intensity presets for common scenarios:
//
// Shake intensities:
//
//	camerafx.Shake.Tiny()        // 0.5  - UI interaction
//	camerafx.Shake.Light()       // 1.5  - Small hit
//	camerafx.Shake.Medium()      // 3.0  - Normal attack
//	camerafx.Shake.Heavy()       // 6.0  - Critical hit
//	camerafx.Shake.Massive()     // 12.0 - Explosion
//	camerafx.Shake.Cataclysmic() // 20.0 - Boss death
//
// Flash colors:
//
//	camerafx.Flash.White()   // Generic impact
//	camerafx.Flash.Red()     // Damage/blood
//	camerafx.Flash.Orange()  // Fire/explosion
//	camerafx.Flash.Blue()    // Ice/energy
//	camerafx.Flash.Green()   // Poison/acid
//	camerafx.Flash.Purple()  // Magic/void
//	camerafx.Flash.Yellow()  // Lightning/holy
//
// # Performance
//
// The system is designed for 60 FPS gameplay:
//
//   - Update: ~3ns per frame
//   - TriggerShake: <1ns
//   - Zero allocations in hot paths
//   - All effects use smooth decay functions
package camerafx

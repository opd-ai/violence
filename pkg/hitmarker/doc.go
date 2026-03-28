// Package hitmarker provides visual hit confirmation feedback.
//
// When the player successfully damages an enemy, a brief marker appears at the
// crosshair position to confirm the hit registered. This is essential for
// responsive combat feel in action games, providing immediate visual confirmation
// that the attack connected, separate from damage numbers (which appear at the
// enemy location and can be missed in chaotic combat).
//
// # Hit Types
//
// The system supports multiple hit types with distinct visual styles:
//
//   - HitNormal: Standard damage hit (white crosshair marks)
//   - HitCritical: Critical/high-damage hit (gold/orange star burst)
//   - HitKill: Killing blow (red X mark, larger and longer duration)
//   - HitHeadshot: Precision shot (crosshair with ring and dot)
//   - HitWeakpoint: Weakpoint damage (diamond marker)
//
// # Genre Theming
//
// Each genre applies distinct color palettes:
//
//   - Fantasy: White/gold/red classic markers
//   - Sci-Fi: Blue/teal/orange tech-style markers
//   - Horror: Dark red/amber/blood-red ominous markers
//   - Cyberpunk: Cyan/magenta/pink neon markers
//   - Post-Apocalyptic: Tan/rust/olive muted markers
//
// # Animation
//
// Hit markers feature:
//
//   - Pop-in animation: Quick scale up with overshoot for impact
//   - Fade out: Smooth alpha fade in final 40% of duration
//   - Rotation: Subtle spin on kills/crits for emphasis
//   - Intensity scaling: Larger/brighter for higher damage
//
// # Usage
//
//	// Create system
//	sys := hitmarker.NewSystem("fantasy")
//	sys.SetScreenSize(320, 200)
//
//	// Create hit marker entity (one per player)
//	markerEnt := hitmarker.SpawnHitMarker(world)
//
//	// Trigger on damage dealt
//	hitmarker.TriggerHit(world, markerEnt, hitmarker.HitNormal, 25, screenCenterX, screenCenterY)
//
//	// Update each frame
//	sys.Update(world)
//
//	// Render to screen
//	sys.Render(world, screen)
//
// # Integration
//
// The hit marker system should be triggered whenever the player deals damage
// to an enemy. It works alongside (not replacing) existing feedback systems:
//
//   - Damage numbers: Show exact value at enemy position
//   - Impact bursts: Show hit effect at enemy position
//   - Muzzle flash: Show weapon fired
//   - Camera shake: Kinesthetic feedback
//   - Hit marker: Confirm hit at player's crosshair (this package)
//
// Together these systems provide comprehensive combat feedback.
package hitmarker

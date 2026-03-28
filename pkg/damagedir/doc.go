// Package damagedir provides directional damage indication via screen-edge vignettes.
//
// When the player takes damage, this system renders red arc vignettes on the screen
// edges corresponding to the direction the damage came from. This provides critical
// spatial awareness in FPS gameplay - players can immediately see which direction
// they're being attacked from without turning around.
//
// Features:
//   - Radial arc vignettes on screen edges showing damage direction
//   - Intensity scales with damage amount (heavier hits = more visible)
//   - Multiple concurrent damage indicators for multiple attackers
//   - Smooth fade-out animation for natural feel
//   - Genre-specific color theming (red for fantasy/horror, cyan for scifi/cyberpunk)
//   - Configurable arc width and falloff for visual customization
//
// The system tracks damage direction in world space and converts to screen-space
// arc positions. It supports up to 8 concurrent damage indicators before oldest
// are replaced.
//
// Integration:
//   - Call TriggerDamage(sourceX, sourceY, playerX, playerY, damage, playerAngle)
//     when player takes damage
//   - Call Render(screen) each frame to draw the indicators
//   - Call Update(w) each frame to animate fade-out
//
// Example:
//
//	sys := damagedir.NewSystem("fantasy")
//	// Player at (100,100) facing east (angle=0), hit by enemy at (80, 100) - damage from west
//	sys.TriggerDamage(80, 100, 100, 100, 25.0, 0.0)
//	// Renders a red arc on the left side of the screen
package damagedir

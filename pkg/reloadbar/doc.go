// Package reloadbar provides a visual weapon reload progress indicator.
//
// The reload bar system renders a progress bar below the crosshair when the
// player is reloading a weapon. This provides critical UI feedback so players
// know when they can fire again.
//
// Features:
//   - Smooth animated progress bar during reload
//   - Genre-specific visual styling (colors, borders, glow effects)
//   - Configurable position relative to crosshair
//   - Fade in/out transitions for polish
//   - Integration with weapon animation system
//   - Compact screen footprint (minimizes UI clutter)
//
// Usage:
//
//	sys := reloadbar.NewSystem("fantasy")
//	// In game update:
//	sys.SetReloadState(isReloading, progress, totalDuration)
//	// In draw:
//	sys.Render(screen, screenCenterX, screenCenterY)
package reloadbar

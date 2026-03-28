// Package objectivecompass provides compact screen-edge navigation indicators
// for active quest objectives.
//
// The objective compass system renders minimal, edge-hugging directional arrows
// that point toward off-screen quest objectives and points of interest. It
// integrates with the quest tracking system and supports genre-specific styling.
//
// Key features:
//   - Screen-edge placement to minimize viewport intrusion
//   - Distance-based scaling (closer objectives have larger indicators)
//   - Distance-based opacity fading (very distant objectives fade out)
//   - Pulsing animation for urgency indication
//   - Genre-specific color schemes and visual styles
//   - Support for multiple objective types (main, bonus, poi)
//   - Compact footprint (≤5% screen edge usage per indicator)
//
// Usage:
//
//	sys := objectivecompass.NewSystem(genreID)
//	sys.SetScreenSize(320, 200)
//	sys.SetPlayerPosition(playerX, playerY, playerAngle)
//
//	// Add objectives to track
//	sys.AddObjective("exit", objectivecompass.TypeMain, exitX, exitY)
//	sys.AddObjective("item", objectivecompass.TypeBonus, itemX, itemY)
//
//	// Update and render
//	sys.Update(deltaTime)
//	sys.Render(screen)
//
// The system automatically handles:
//   - Culling objectives that are on-screen (no indicator needed)
//   - Preventing indicator overlap at screen edges
//   - Smooth animations and transitions
//   - Priority-based rendering (main objectives over bonus)
package objectivecompass

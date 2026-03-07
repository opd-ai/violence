// Package entitylabel provides guaranteed text rendering for entity names and labels.
//
// This system solves the "UN-RENDERED TEXT" problem by ensuring that all entity
// names, tooltips, and labels are actually drawn on screen with proper fallback
// handling and visibility management.
//
// # Key Features
//
//   - Guaranteed text rendering: All text is drawn using bundled bitmap fonts with
//     fallback mechanisms. If primary font fails, fallback font is tried. If both
//     fail, a colored placeholder rectangle is drawn and error is logged.
//   - Proximity-based visibility: Labels only render within MaxDistance range,
//     with smooth alpha fading at the edge of visibility range.
//   - Type-specific styling: Enemies (red), NPCs (green), loot (yellow),
//     interactables (cyan), and bosses (orange) each have distinct colors, scales,
//     and visibility ranges.
//   - Layout manager integration: Uses ui.LayoutManager to prevent overlapping with
//     other UI elements like health bars and damage numbers.
//   - Distance-based detail: Far entities get no label, medium entities get faded
//     labels, close entities get full visibility. Boss and interactable labels
//     are always visible within range.
//   - Screen-edge clamping: Labels stay on-screen and do not clip at viewport edges.
//
// # Usage
//
// Create the system during game initialization:
//
//	labelSystem := entitylabel.NewSystem("fantasy")
//	world.AddSystem(labelSystem)
//
// Add label components to entities:
//
//	// Enemy
//	enemyEnt := world.AddEntity()
//	world.AddComponent(enemyEnt, &engine.Position{X: 10, Y: 10})
//	world.AddComponent(enemyEnt, entitylabel.NewEnemyLabel("Goblin"))
//
//	// Boss
//	bossEnt := world.AddEntity()
//	world.AddComponent(bossEnt, &engine.Position{X: 50, Y: 50})
//	world.AddComponent(bossEnt, entitylabel.NewBossLabel("Dragon Lord"))
//
//	// Loot
//	lootEnt := world.AddEntity()
//	world.AddComponent(lootEnt, &engine.Position{X: 20, Y: 20})
//	world.AddComponent(lootEnt, entitylabel.NewLootLabel("Health Potion"))
//
// Render labels in the main draw loop:
//
//	// Without layout manager (basic mode)
//	labelSystem.Render(world, screen, cameraX, cameraY)
//
//	// With layout manager (prevents overlap)
//	labelSystem.RenderWithLayout(world, screen, cameraX, cameraY, layoutManager)
//
// # Architecture
//
// The system follows Violence's ECS architecture:
//   - Component: Pure data struct with Type() method, no logic
//   - System: All rendering logic in Update() and Render() methods
//   - No external assets: Uses bundled basicfont.Face7x13
//   - Logging: Uses logrus with structured fields
//
// # Performance
//
// The system is optimized for 60+ FPS:
//   - Early distance culling: Entities beyond MaxDistance are skipped entirely
//   - Off-screen culling: Labels outside viewport (with margin) are not drawn
//   - Direct text rendering: No intermediate buffers for normal-scale text
//   - Scaled text pooling: Only boss/special labels use temporary images
//
// # Text Rendering Guarantee
//
// The system implements triple-layer fallback for text rendering:
//  1. Primary font (basicfont.Face7x13) with panic recovery
//  2. Fallback font (currently same, but separate for future customization)
//  3. Placeholder rectangle + error log if both fonts fail
//
// This ensures no text is silently invisible. If text cannot be drawn,
// the player sees a colored rectangle and developers see a logged warning.
package entitylabel

// Package automap provides an in-game auto-mapping system with exploration
// tracking, annotation support, and genre-themed minimap rendering.
//
// # Core Features
//
//   - Automatic exploration tracking via Reveal()
//   - Special markers (secrets, objectives, items) via AddAnnotation()
//   - Configurable minimap rendering with fog of war
//   - Genre-specific visual themes (fantasy, scifi, horror, etc.)
//
// # Basic Usage
//
//	// Create a map for a 100x100 level
//	m := automap.NewMap(100, 100)
//
//	// Reveal cells as player explores
//	m.Reveal(playerX, playerY)
//
//	// Add annotation for discovered secret
//	m.AddAnnotation(secretX, secretY, automap.AnnotationSecret)
//
//	// Render minimap in top-right corner
//	m.Render(screen, playerX, playerY, playerAngle, walls)
//
// # Advanced Rendering
//
// For custom minimap positioning and appearance, use RenderMinimap with
// a RenderConfig:
//
//	cfg := automap.RenderConfig{
//		X: 100, Y: 100,
//		Width: 150, Height: 150,
//		CellSize: 2.5,
//		PlayerX: playerX, PlayerY: playerY,
//		Opacity: 0.9,
//		ShowFogOfWar: true,
//	}
//	m.RenderMinimap(screen, cfg)
//
// # Genre Theming
//
// The package supports genre-specific visual themes. Set the global genre
// to customize automap colors and styles:
//
//	automap.SetGenre("scifi")  // Blue/cyan theme
//	automap.SetGenre("horror") // Dark red/purple theme
//
// # Thread Safety
//
// Map instances are not thread-safe. Genre settings use a global mutex
// and are safe for concurrent access across goroutines.
package automap

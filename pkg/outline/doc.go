// Package outline provides sprite silhouette rendering for visual clarity.
//
// The outline system addresses the POOR READABILITY visual problem by adding
// configurable sprite outlines that improve visual hierarchy and entity distinction
// during chaotic combat. Genre-specific color schemes ensure allies, enemies,
// and interactable objects are instantly recognizable.
//
// Key features:
//   - Per-pixel outline generation with distance-based falloff
//   - Genre-aware color palettes (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic)
//   - Optional glow effect with soft penumbra
//   - LRU caching for performance (zero allocations on cache hits)
//   - Memory-pooled image buffers to avoid GC pressure
//
// Usage:
//
//	sys := outline.NewSystem("fantasy")
//	world.AddSystem(sys)
//
//	// Add outline component to entities for automatic rendering
//	entity := world.AddEntity()
//	world.AddComponent(entity, &outline.Component{
//	    Enabled:   true,
//	    Color:     sys.GetEnemyColor(),
//	    Thickness: 2,
//	    Glow:      false,
//	})
//
//	// Generate outlined sprite manually
//	outlined := sys.GenerateOutline(sprite, color.RGBA{R: 255, G: 0, B: 0, A: 255}, 2, false)
package outline

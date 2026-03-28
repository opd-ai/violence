// Package floorreflect provides floor-based sprite reflections for visual realism.
//
// This system renders soft reflections of sprites (entities, props, particles) on
// wet, polished, or metallic floor surfaces. Reflections add significant visual
// depth and make environments feel more realistic and lived-in.
//
// Visual Features:
//   - Vertical flip of sprite images as reflections
//   - Soft fade-out with distance from floor contact point
//   - Surface-material-based reflectivity (metal > wet > polished stone > dry)
//   - Genre-specific reflection intensity and color tinting
//   - Light-source-aware reflection brightness
//   - Performance-optimized with reflection caching and LOD
//
// Genre Presets:
//   - Fantasy: Subtle reflections on wet stone, stronger in water areas
//   - Sci-Fi: Strong reflections on polished metal floors
//   - Horror: Dark murky reflections with distortion
//   - Cyberpunk: Neon-tinted reflections on rain-slicked streets
//   - Post-Apocalyptic: Partial reflections on puddles and oil slicks
//
// The reflection effect uses vertical sprite flipping with alpha fade and optional
// distortion noise for water/wet surfaces. Reflections respect floor tile material
// type and only render where the floor surface is reflective.
//
// Usage:
//
//	sys := floorreflect.NewSystem(genreID)
//	sys.SetReflectiveFloors(reflectiveTiles) // from wetness or floor system
//
//	// In render loop, after floor but before entities
//	for _, entity := range visibleEntities {
//	    if comp := world.GetComponent(entity, reflectType); comp != nil {
//	        sys.RenderReflection(screen, comp.(*Component), floorY)
//	    }
//	}
//
// The system integrates with the wetness package to identify reflective surfaces
// and with the lighting system for accurate reflection brightness.
package floorreflect

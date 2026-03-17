// Package rimlight provides directional edge lighting effects for sprites.
//
// Rim lighting creates a bright halo/edge highlight on the side of sprites
// facing away from the light source, simulating backlight and subsurface
// scattering effects that dramatically improve visual depth and separation.
//
// Key features:
//   - Directional rim highlights based on global light direction
//   - Fresnel-based edge intensity calculation for realistic falloff
//   - Material-specific rim intensity (metal reflects more, cloth less)
//   - Genre-specific rim colors (warm for fantasy, cool for sci-fi)
//   - Seed-based consistent results for deterministic rendering
//   - LRU cache for processed sprites to maintain 60+ FPS
//
// Unlike outline rendering (which adds a uniform border), rim lighting
// creates directional highlights that respond to the scene's light position,
// giving sprites a three-dimensional appearance.
//
// Integration:
//
// The system is registered automatically and applies rim lighting to all
// entities with RimLightComponent. Use EnableRimLight() helper to add
// the component with appropriate defaults for an entity type.
//
// Example usage:
//
//	// Add rim lighting to an entity
//	comp := rimlight.NewComponent()
//	comp.Material = rimlight.MaterialMetal
//	world.AddComponent(entityID, comp)
package rimlight

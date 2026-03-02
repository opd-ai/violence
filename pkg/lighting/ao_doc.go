// Package lighting provides ambient occlusion calculation for depth cues.
//
// # Ambient Occlusion System
//
// The ambient occlusion (AO) system calculates how occluded each entity is by nearby
// geometry and other entities. This creates depth cues and atmospheric shading that
// dramatically improves visual quality, especially at small sprite sizes.
//
// # Usage
//
// Basic setup:
//
//	world := engine.NewWorld()
//	aoSystem := lighting.NewAOSystem("fantasy")
//	world.AddSystem(aoSystem)
//
//	// Optional: connect to spatial index for performance
//	aoSystem.SetSpatialGrid(spatialSystem.GetGrid())
//
//	// Create entity with AO
//	entity := world.AddEntity()
//	pos := &lighting.PositionComponent{X: 10, Y: 10}
//	ao := lighting.NewAOComponent(2.0) // Sample radius
//	world.AddComponent(entity, pos)
//	world.AddComponent(entity, ao)
//
//	// System updates automatically
//	world.Update()
//
// # Genre-Specific Behavior
//
// Different genres have different AO intensity:
//
//   - Fantasy: Strong AO (0.7) — emphasizes cramped dungeon corridors
//   - Horror: Very strong AO (0.85) — deep shadows in every crevice
//   - Sci-fi: Moderate AO (0.5) — clean facilities with defined edges
//   - Cyberpunk: Strong AO (0.65) — cluttered urban environments
//   - Post-apocalyptic: Moderate AO (0.6) — dusty ambient with some depth
//
// # Performance
//
// The system is designed for real-time use:
//
//   - Updates are throttled (default: every 4 frames)
//   - Results are cached until invalidated
//   - Spatial indexing used for proximity queries when available
//   - Sample count and radius are tunable
//
// Invalidate AO when entities move or map changes:
//
//	ao.Invalidate()              // Single entity
//	aoSystem.InvalidateAll(world) // All entities
//
// # Rendering Integration
//
// Use AO values to darken sprites based on local occlusion:
//
//	aoComp, _ := world.GetComponent(entity, reflect.TypeOf(&lighting.AOComponent{}))
//	ao := aoComp.(*lighting.AOComponent)
//
//	// Overall occlusion for simple shading
//	brightness := ao.Overall // 0.0 (fully occluded) to 1.0 (no occlusion)
//	spriteColor.R = uint8(float64(spriteColor.R) * brightness)
//
//	// Directional occlusion for advanced shading
//	angleToLight := math.Atan2(lightY-entityY, lightX-entityX)
//	occlusion := ao.GetOcclusionAt(angleToLight)
//	// Apply directional darkening
//
// # Architecture
//
// The system follows Violence's ECS architecture:
//
//   - AOComponent: Pure data component storing occlusion factors
//   - AOSystem: System that processes entities and calculates occlusion
//   - No dependencies on rendering — AO values are consumed by render code
//
// # Technical Details
//
// Occlusion calculation:
//
//  1. For each entity with AOComponent
//  2. Cast rays in 8 directions (N, NE, E, SE, S, SW, W, NW)
//  3. Sample at multiple distances (0.5 to SampleRadius)
//  4. Check for wall collisions and nearby entities
//  5. Apply distance-based falloff (closer occluders = stronger AO)
//  6. Store occlusion factor per direction [0.0-1.0]
//
// Wall detection uses grid-based collision (positions near grid lines are walls).
// Entity detection uses spatial grid if available, else brute-force proximity check.
//
// # Coverage
//
// Test coverage target: ≥40% overall, ≥30% for display-dependent code.
package lighting

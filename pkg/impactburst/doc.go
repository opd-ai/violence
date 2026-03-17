// Package impactburst provides visually realistic impact burst rendering with
// shockwaves, material-specific debris, directional particle trails, and glow effects.
//
// This system addresses the WEAK FEEDBACK and UNCONVINCING MATERIALS visual
// realism problems by rendering combat impacts with:
//
//   - Multi-layer shockwave rings with inverse-square falloff
//   - Directional debris particles based on impact angle
//   - Material-specific visual signatures (metal sparks, stone chips, blood splatter)
//   - Genre-configurable color palettes and effect intensities
//   - Glow bloom effects for critical and magical impacts
//   - Light-source-consistent shading on all visual elements
//
// Integration:
//
//	impactSystem := impactburst.NewSystem("fantasy")
//	g.world.AddSystem(impactSystem)
//
// Impact spawning:
//
//	impactSystem.SpawnImpact(x, y, angle, impactburst.ImpactMelee, impactburst.MaterialMetal, 1.0)
//
// Rendering:
//
//	impactSystem.Render(screen, cameraX, cameraY, screenWidth, screenHeight)
package impactburst

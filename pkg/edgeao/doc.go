// Package edgeao provides edge ambient occlusion for environment geometry.
//
// Edge ambient occlusion darkens pixels near wall-floor junctions, corners,
// and other geometric edges to add visual depth and grounding. Unlike entity
// AO which tracks dynamic objects, edge AO is based on static level geometry.
//
// The system precomputes an AO map for the level, identifying:
//   - Wall-floor junctions (where floor meets wall base)
//   - Inside corners (L-shaped wall intersections)
//   - Outside corners (convex wall edges)
//   - Narrow passages (corridors with walls on both sides)
//
// AO values are computed using distance falloff from edges:
//   - Maximum darkness at the contact point
//   - Gradual falloff over configurable distance
//   - Genre-specific intensity and falloff curves
//
// Integration with the render pipeline:
//
//	edgeAO := edgeao.NewSystem(genreID)
//	edgeAO.BuildAOMap(levelTiles)
//
//	// During floor/ceiling rendering:
//	aoFactor := edgeAO.GetAO(worldX, worldY)
//	finalColor = applyAO(baseColor, aoFactor)
//
// Genre presets control the visual character:
//   - Fantasy: Strong AO in dungeon corners and alcoves
//   - Sci-Fi: Subtle edge definition on clean surfaces
//   - Horror: Deep shadows in corners and narrow passages
//   - Cyberpunk: Moderate AO with neon-lit edge contrast
//   - Post-Apoc: Dusty accumulation patterns at edges
package edgeao

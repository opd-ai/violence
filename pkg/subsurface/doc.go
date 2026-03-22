// Package subsurface provides subsurface scattering simulation for organic materials.
//
// Subsurface scattering (SSS) is the physical phenomenon where light penetrates a
// translucent surface, scatters internally, and exits at a different point. This
// creates the characteristic soft, glowing appearance of skin, wax, leaves, and
// other organic materials.
//
// Key improvements for visual realism:
//   - Light penetration: Light entering the material scatters and exits elsewhere,
//     creating soft color bleeding and warm undertones
//   - Depth-based color shift: Deeper penetration shifts colors toward warmer tones
//     (red/orange) as blue light is absorbed faster
//   - Edge translucency: Thin edges (ears, fingers, leaves) show more light transmission
//   - Material profiles: Different organic materials have distinct scattering properties
//     (skin is warm/red, leaves are green, wax is neutral)
//
// The system works by analyzing sprite silhouettes, computing thickness maps, and
// applying appropriate color shifts to simulate light traveling through the material.
// Integration is via ECS components attached to entities with organic materials.
//
// Material types supported:
//   - MaterialFlesh: Human/animal skin with strong red/warm scattering
//   - MaterialLeaf: Plant matter with green-biased scattering
//   - MaterialWax: Neutral warm scattering for candle/artificial materials
//   - MaterialSlime: Translucent gel with color preservation
//   - MaterialMembrane: Thin translucent tissue (wings, fins)
//
// Example usage:
//
//	entity := world.AddEntity()
//	world.AddComponent(entity, &subsurface.Component{
//	    Enabled:   true,
//	    Material:  subsurface.MaterialFlesh,
//	    Intensity: 1.0,
//	})
//
// The system processes entities during the render phase, applying SSS effects to
// sprites based on light direction and material properties.
package subsurface

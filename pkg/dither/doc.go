// Package dither provides procedural dithering for pixel art sprite rendering.
//
// Dithering creates smooth tonal transitions using strategic pixel patterns,
// which is essential for making pixel art feel three-dimensional and realistic
// at small sizes. This package implements multiple dithering algorithms:
//
//   - **Ordered (Bayer) Dithering**: Uses a threshold matrix for predictable,
//     artifact-free patterns ideal for gradients and lighting transitions.
//   - **Blue Noise Dithering**: Provides organic-looking patterns with less
//     visible structure, perfect for natural materials like skin and cloth.
//   - **Edge-Aware Dithering**: Concentrates dithering at tonal boundaries
//     to create smooth shading transitions without affecting flat color areas.
//
// # Material-Specific Dithering
//
// The system applies different dithering strategies per material type:
//   - Metals: High-contrast specular dithering for metallic sheen
//   - Cloth/Fur: Soft diffuse dithering for organic softness
//   - Crystal: Sparse noise dithering for magical sparkle
//   - Flesh/Leather: Medium-density pattern for natural gradients
//
// # Usage
//
// Dithering is automatically integrated into sprite generation via ApplyDithering:
//
//	ditherSys := dither.NewSystem(seed)
//	ditherSys.ApplyDithering(spriteImage, dither.MaterialMetal, 0.5)
//
// For fine-grained control over gradient regions:
//
//	ditherSys.ApplyGradientDithering(img, bounds, color1, color2, direction)
//
// # Integration
//
// This package integrates with pkg/sprite during sprite generation. The dithering
// is applied as a final pass after PBR shading to enhance tonal transitions.
package dither

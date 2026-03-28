// Package materialedge provides silhouette-edge effects for material recognition at small sprite sizes.
//
// Problem: Interior material patterns (scales, fur, chitin, etc.) vanish below ~32px.
// At typical gameplay distances (16-24px sprites), materials all look like flat fills.
//
// Solution: Apply high-contrast edge treatments that survive downscaling:
//   - Fur: Ragged silhouette edge with hair strands extending outward
//   - Chitin: Smooth edge with segmented glossy highlight bands
//   - Scales: Regular diamond/dot pattern at silhouette boundary
//   - Metal: Clean edge with single bright specular spot
//   - Slime: Irregular wobble edge with transparency drips
//   - Cloth: Soft, slightly wavy edge
//   - Membrane: Thin, translucent edge with vein highlights
//   - Leather: Organic grain visible at edge
//   - Crystal: Sharp faceted edge with bright refractions
//
// The edge treatment is the highest-contrast region of the sprite and communicates
// material faster than interior pixels at small sizes.
package materialedge

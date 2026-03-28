// Package emissive provides screen-space emissive glow rendering for light sources and magic effects.
//
// The emissive system adds realistic glow halos around entities that emit light - torches,
// projectiles, magical effects, and creature eyes. Unlike simple bloom post-processing,
// this system renders per-entity glow with proper color, intensity, and falloff.
//
// Features:
//   - Per-entity glow halos with configurable color and intensity
//   - Genre-specific glow styles (warm/organic fantasy, cool/sharp sci-fi, etc.)
//   - Distance-based glow scaling (closer = larger, more intense)
//   - Multiple glow layers for richer visual effect (core, inner halo, outer halo)
//   - Flicker integration for dynamic light sources
//   - Performance-optimized with sprite caching and LRU eviction
//
// Usage:
//
//	sys := emissive.NewSystem("fantasy")
//	world.AddComponent(torchEntity, &emissive.Component{
//	    Intensity: 1.0,
//	    Color:     color.RGBA{R: 255, G: 180, B: 80, A: 255},
//	    Radius:    24.0,
//	    Type:      emissive.TypeFlame,
//	})
//
// Genre Presets:
//   - Fantasy: Warm orange/gold glows, soft falloff, flame flicker
//   - Sci-Fi: Cool blue/cyan glows, sharp edges, steady emission
//   - Horror: Sickly green/purple glows, pulsing, unsettling
//   - Cyberpunk: Neon pink/cyan glows, intense, sharp
//   - Post-Apoc: Dim amber/red glows, flickering, sparse
package emissive

// Package audio provides procedurally generated sound effects, adaptive music layers,
// ambient soundscapes, and dynamic reverb effects for the Violence game engine.
//
// # Architecture
//
// The audio system consists of three main components:
//
//   - Engine: Manages music playback with adaptive intensity layers and 3D positioned sound effects
//   - ReverbCalculator: Computes room-based reverb effects from BSP level geometry
//   - Ambient: Generates genre-specific background soundscapes
//
// # Procedural Generation
//
// All audio is generated procedurally at runtime using deterministic algorithms.
// No pre-rendered audio files are used. Generation is seeded by:
//
//   - Sound name (for SFX and music)
//   - Genre ID (fantasy, scifi, horror, cyberpunk, postapoc)
//   - Layer index (for adaptive music intensity)
//
// # Genre System
//
// The audio system supports five game genres, each with distinct sonic characteristics:
//
//   - Fantasy: Orchestral instruments, medieval timbres, modal harmony
//   - SciFi: Synthetic waveforms, metallic resonances, futuristic textures
//   - Horror: Dissonant intervals, sub-bass drones, unsettling noise layers
//   - Cyberpunk: Electronic beats, glitch effects, industrial tones
//   - PostApoc: Percussive elements, distorted textures, sparse arrangements
//
// # Adaptive Music
//
// Music consists of a base layer plus up to 3 intensity layers. The Engine
// crossfades layer volumes based on gameplay intensity (0.0-1.0):
//
//   - Intensity 0.0: Base layer only
//   - Intensity 0.5: Base layer + partial layer 1
//   - Intensity 1.0: All layers at full volume
//
// Layer generation uses consistent seeds to ensure synchronization.
//
// # 3D Audio
//
// Sound effects use distance-based attenuation (inverse square law) and
// stereo panning based on listener position. Reverb parameters adapt
// dynamically to room geometry from BSP level data.
//
// # Usage Example
//
//	engine := audio.NewEngine()
//	engine.SetGenre("horror")
//	engine.SetListenerPosition(10.0, 5.0)
//
//	// Play adaptive music with medium intensity
//	engine.PlayMusic("combat_theme", 0.5)
//
//	// Play 3D positioned sound effect
//	engine.PlaySFX("explosion", 15.0, 8.0)
//
//	// Update intensity dynamically
//	engine.SetIntensity(0.9)
//
// # Thread Safety
//
// All Engine methods are safe for concurrent use via internal mutex protection.
package audio

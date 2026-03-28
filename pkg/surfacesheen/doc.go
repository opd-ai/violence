// Package surfacesheen provides dynamic surface reflection rendering based on
// material properties and nearby light sources.
//
// This system enhances visual realism by adding material-appropriate sheen
// effects to sprites and surfaces. Different materials (metal, wet, polished,
// organic) exhibit distinct reflection behaviors:
//
//   - Metallic surfaces: Directional specular highlights that track light sources
//   - Wet surfaces: Broad, soft reflections with slight distortion
//   - Polished surfaces: Sharp, mirror-like highlights
//   - Organic surfaces: Subtle subsurface-influenced glow
//
// The system supports multiple simultaneous light sources with color temperature,
// intensity falloff, and material interaction. Sheen intensity varies with
// distance from light sources using inverse-square falloff.
//
// Genre-specific presets adjust overall sheen intensity and color warmth:
//
//   - Fantasy: Warm golden sheens, moderate intensity
//   - Sci-Fi: Cool blue-tinted highlights, sharp specular
//   - Horror: Minimal sheen, sickly green-tinged wet surfaces
//   - Cyberpunk: Neon-tinted reflections, high contrast
//   - Post-Apocalyptic: Dusty, muted reflections with rust undertones
//
// Integration:
//
//	sheenSystem := surfacesheen.NewSystem("fantasy")
//	sheenSystem.Update(entities, deltaTime)
//	sheenSystem.Render(screen, cameraX, cameraY, lights)
//
// The system integrates with the lighting system to extract light positions
// and colors, and with the sprite/material systems to determine surface types.
package surfacesheen

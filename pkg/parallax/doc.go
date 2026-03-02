// Package parallax implements multi-layer parallax scrolling backgrounds
// for enhanced visual depth in procedurally generated environments.
//
// The parallax system generates and renders multiple background layers that
// move at different speeds relative to the camera, creating an illusion of
// depth. Each layer can be procedurally generated with genre-specific themes.
//
// Integration:
//   - Register ParallaxSystem in main game loop
//   - Layers auto-generate based on biome/genre
//   - Renders behind all entities and terrain
package parallax

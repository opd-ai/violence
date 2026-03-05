// Package particle provides enhanced particle rendering with varied shapes and effects.
//
// The particle renderer system replaces the placeholder 2x2 pixel particle rendering
// with visually distinct shapes based on particle properties:
//
// - Fast-moving particles with vertical velocity → Sparks (elongated streaks)
// - Slow-moving particles drifting upward → Smoke (soft, expanding clouds)
// - Red particles → Diamonds (blood splatter)
// - Bright yellow/white particles → Stars (muzzle flash, magic)
// - Medium-speed particles → Glow (explosions, fire)
// - Fast directional particles → Lines (motion trails)
// - Default → Circles (generic particles)
//
// Example integration in main game loop:
//
//	// In NewGame initialization:
//	g.particleRenderer = particle.NewRendererSystem()
//	g.world.AddSystem(g.particleRenderer)
//
//	// In render loop:
//	renderer := g.particleRenderer.GetRenderer()
//	for i := range particles {
//	    p := &particles[i]
//	    screenX := float32(worldToScreenX(p.X))
//	    screenY := float32(worldToScreenY(p.Y))
//	    renderer.RenderParticle(screen, p, screenX, screenY, genreID)
//	}
//
// The system automatically selects appropriate shapes based on particle velocity,
// color, and other properties, creating visually rich feedback without manual
// per-particle shape assignment.
package particle

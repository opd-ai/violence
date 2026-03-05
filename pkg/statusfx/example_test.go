package statusfx_test

import (
	"time"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/particle"
	"github.com/opd-ai/violence/pkg/status"
	"github.com/opd-ai/violence/pkg/statusfx"
)

// Example demonstrates basic usage of the status visual effects system.
func Example() {
	// Create particle system for status effect particles
	ps := particle.NewParticleSystem(1000, 42)

	// Create the status FX system
	sys := statusfx.NewSystem("fantasy", ps)

	// Create ECS world and entity
	w := engine.NewWorld()
	entity := w.AddEntity()

	// Add position component
	w.AddComponent(entity, &engine.Position{X: 10, Y: 10})

	// Add status effects
	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "burning",
				TimeRemaining: 5 * time.Second,
				VisualColor:   0xFF880088, // Orange-red
			},
			{
				EffectName:    "poisoned",
				TimeRemaining: 10 * time.Second,
				VisualColor:   0x88FF0088, // Green
			},
		},
	}
	w.AddComponent(entity, statusComp)

	// Update the system - this creates visual components
	sys.Update(w)

	// The system now manages pulsating auras and particle emissions
	// for all entities with active status effects.
	// In the game loop, call:
	//
	//	sys.Update(world)       // Each frame to sync visuals
	//	sys.Render(screen, ...)  // Each frame to draw effects
}

// Example_genreSpecific shows how different genres use the same system.
func Example_genreSpecific() {
	ps := particle.NewParticleSystem(500, 123)

	// Create system for cyberpunk genre
	sys := statusfx.NewSystem("cyberpunk", ps)

	w := engine.NewWorld()
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 5, Y: 5})

	// Apply cyberpunk-themed status effects
	statusComp := &status.StatusComponent{
		ActiveEffects: []status.ActiveEffect{
			{
				EffectName:    "hacked",
				TimeRemaining: 8 * time.Second,
				VisualColor:   0x00FF8888, // Cyan
			},
			{
				EffectName:    "glitched",
				TimeRemaining: 5 * time.Second,
				VisualColor:   0xFF88FF88, // Magenta
			},
		},
	}
	w.AddComponent(entity, statusComp)

	sys.Update(w)

	// Visual appearance matches the genre - same code, different visuals
}

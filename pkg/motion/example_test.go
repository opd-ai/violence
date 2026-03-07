package motion_test

import (
	"fmt"
	"math"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/motion"
)

// Example demonstrates organic motion on a moving entity.
func Example() {
	// Create world and motion system
	w := engine.NewWorld()
	sys := motion.NewSystem()

	// Create an entity with position, velocity, and motion
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})
	w.AddComponent(entity, &engine.Velocity{DX: 10.0, DY: 5.0})

	// Initialize motion with mass 2.0 and a trailing element (tail/cloak)
	motionComp := motion.InitializeMotion(2.0, true)
	w.AddComponent(entity, motionComp)

	// Simulate several frames
	for frame := 0; frame < 60; frame++ {
		sys.Update(w)
	}

	// Check that easing has smoothed the velocity
	fmt.Printf("Motion component initialized and running\n")
	fmt.Printf("Breath phase advanced: %.2f\n", motionComp.BreathPhase)
	fmt.Printf("Trail segments: %d\n", len(motionComp.TrailOffsetX))
	fmt.Printf("Squash recovery: X=%.2f Y=%.2f\n", motionComp.SquashX, motionComp.SquashY)

	// Output:
	// Motion component initialized and running
	// Breath phase advanced: 1.26
	// Trail segments: 5
	// Squash recovery: X=1.00 Y=1.00
}

// Example_impact demonstrates squash and stretch on impact.
func Example_impact() {
	sys := motion.NewSystem()
	motionComp := motion.InitializeMotion(1.5, false)

	// Set ImpactTime to allow impact
	motionComp.ImpactTime = 0.2

	// Trigger a high-velocity impact
	sys.TriggerImpact(motionComp, 80.0)

	scaleX, scaleY := sys.GetSquashStretch(motionComp)

	// Entity should be squashed vertically and stretched horizontally
	fmt.Printf("After impact: squashed=%.2f stretched=%.2f\n", scaleY, scaleX)

	// Output:
	// After impact: squashed=0.68 stretched=1.24
}

// Example_breathing demonstrates idle breathing animation.
func Example_breathing() {
	sys := motion.NewSystem()
	motionComp := &motion.Component{
		BreathFrequency: 0.2,
		BreathAmplitude: 1.5,
		BreathPhase:     math.Pi / 2, // Peak of breath
	}

	offset := sys.GetBreathOffset(motionComp)

	// At π/2, sine wave is at maximum (1.0)
	fmt.Printf("Breath offset at peak: %.1f pixels\n", offset)

	// Output:
	// Breath offset at peak: 1.5 pixels
}

// Example_trail demonstrates secondary motion for trailing elements.
func Example_trail() {
	w := engine.NewWorld()
	sys := motion.NewSystem()

	entity := w.AddEntity()
	pos := &engine.Position{X: 0, Y: 0}
	w.AddComponent(entity, pos)

	// Create motion with a 3-segment trail
	motionComp := &motion.Component{
		TrailLength:    3,
		TrailStiffness: 0.4,
	}
	w.AddComponent(entity, motionComp)

	// Initialize trail
	sys.Update(w)

	// Move entity rapidly
	pos.X = 100
	pos.Y = 50

	// Update several times
	for i := 0; i < 20; i++ {
		sys.Update(w)
	}

	// Trail segments lag behind main position
	x, y, _ := sys.GetTrailSegment(motionComp, 2)
	fmt.Printf("Main position: (%.0f, %.0f)\n", pos.X, pos.Y)
	fmt.Printf("Trail end lags: (%.0f, %.0f)\n", x, y)

	// Output:
	// Main position: (100, 50)
	// Trail end lags: (32, 16)
}

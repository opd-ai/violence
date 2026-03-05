// Package damagenumber provides floating combat text for damage feedback.
//
// # Overview
//
// This package implements an ECS-based damage number system that displays
// floating text when entities take damage or receive healing.
//
// # Basic Usage
//
//	// Initialize system
//	sys := damagenumber.NewSystem("fantasy")
//	world.AddSystem(sys)
//
//	// Spawn damage number
//	damagenumber.Spawn(world, 42, entityX, entityY, "physical", false, false)
//
//	// In render loop
//	sys.Render(world, screen, cameraX, cameraY)
//
// # Features
//
//   - Animated rise, scale-in, and fade-out
//   - Damage-type-specific colors
//   - Critical hit emphasis with pulsing
//   - Healing indication in green
//   - Zero per-frame allocations
//
// # Animation Timeline
//
//   - 0-20%: Scale from 0.5x to 1.0x
//   - 20-70%: Full display with upward motion
//   - 70-100%: Fade to transparent
//
// # Damage Type Colors
//
// Physical/Kinetic: Light Red
// Fire/Heat: Orange-Red
// Ice/Cold: Cyan
// Lightning/Electric: Yellow
// Poison/Toxic: Yellow-Green
// Dark/Shadow: Purple
// Holy/Light: Pale Yellow
// Arcane/Magic: Light Purple
// Healing: Green
//
// See README.md for detailed documentation.
package damagenumber

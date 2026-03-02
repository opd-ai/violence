package attacktrail_test

import (
	"image/color"
	"math/rand"

	"github.com/opd-ai/violence/pkg/attacktrail"
	"github.com/opd-ai/violence/pkg/engine"
)

// Example_basicUsage demonstrates how to integrate attack trails into a combat system.
func Example_basicUsage() {
	// Create ECS world and trail system
	world := engine.NewWorld()
	trailSystem := attacktrail.NewSystem("fantasy")
	world.AddSystem(trailSystem)

	// Create a player entity
	playerID := world.AddEntity()

	// When the player attacks, create a trail
	playerX, playerY := 100.0, 200.0
	dirX, dirY := 1.0, 0.0 // Attacking to the right
	weaponRange := 50.0

	// Define a color function (normally from the trail system)
	getColor := func(weaponName string, rng *rand.Rand) color.RGBA {
		return color.RGBA{R: 200, G: 220, B: 255, A: 180}
	}

	// Attach trail to attack
	rng := rand.New(rand.NewSource(12345))
	attacktrail.AttachTrailToAttack(
		world,
		playerID,
		playerX, playerY,
		dirX, dirY,
		weaponRange,
		"sword",
		rng,
		getColor,
	)

	// Update trail system (called each frame)
	world.Update()

	// Render trails (in your Draw method)
	// screen := ebiten.NewImage(800, 600)
	// trailSystem.Render(screen, world, cameraX, cameraY)
}

// Example_weaponTypes demonstrates different weapon trail types.
func Example_weaponTypes() {
	world := engine.NewWorld()
	entityID := world.AddEntity()

	x, y := 100.0, 100.0
	angle := 0.0

	// Create different trail types
	trailComp := attacktrail.NewTrailComponent(5)
	world.AddComponent(entityID, trailComp)

	// Slash trail for swords
	slashTrail := attacktrail.CreateSlashTrail(
		x, y, angle, 60.0, 1.57, 3.0, // 90-degree arc
		color.RGBA{R: 200, G: 220, B: 255, A: 180},
	)
	trailComp.AddTrail(slashTrail)

	// Thrust trail for spears
	thrustTrail := attacktrail.CreateThrustTrail(
		x, y, angle, 80.0, 2.5,
		color.RGBA{R: 180, G: 180, B: 200, A: 200},
	)
	trailComp.AddTrail(thrustTrail)

	// Smash trail for hammers
	smashTrail := attacktrail.CreateSmashTrail(
		x, y, 40.0, 5.0,
		color.RGBA{R: 255, G: 200, B: 100, A: 220},
	)
	trailComp.AddTrail(smashTrail)
}

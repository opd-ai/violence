// Example demonstrates the terrain sliding system for smooth wall collision.
package main

import (
	"fmt"
	"reflect"

	"github.com/opd-ai/violence/pkg/collision"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/spatial"
)

func main() {
	fmt.Println("=== Terrain Sliding System Example ===")

	// Create ECS world and spatial index
	world := engine.NewWorld()
	spatialGrid := spatial.NewGrid(64.0)
	slidingSystem := collision.NewSlidingSystem(spatialGrid)

	// Create player entity with sliding enabled
	player := createPlayer(world, 50, 50)
	fmt.Printf("Created player at (50, 50)\n")

	// Create vertical wall at x=60
	_ = createWall(world, spatialGrid, 60, 0, 10, 200)
	fmt.Printf("Created vertical wall at x=60\n")

	// Set player velocity - diagonal movement toward wall
	setVelocity(world, player, 100, 100)
	fmt.Printf("Player velocity: (100, 100) units/sec\n\n")

	// Simulate 5 frames of movement
	fmt.Println("Simulating 5 frames of movement:")
	for frame := 0; frame < 5; frame++ {
		// Update spatial index with current positions
		updateSpatialIndex(world, spatialGrid)

		// Get player position before update
		posBefore := getPosition(world, player)

		// Run sliding system
		slidingSystem.Update(world)

		// Get player position after update
		posAfter := getPosition(world, player)

		fmt.Printf("  Frame %d: (%.2f, %.2f) -> (%.2f, %.2f)\n",
			frame+1, posBefore.X, posBefore.Y, posAfter.X, posAfter.Y)
	}

	fmt.Println("\nResult:")
	finalPos := getPosition(world, player)
	fmt.Printf("  Final position: (%.2f, %.2f)\n", finalPos.X, finalPos.Y)
	fmt.Printf("  Player slid along wall in Y direction\n")
	fmt.Printf("  X position clamped near wall (collision boundary)\n")
}

func createPlayer(world *engine.World, x, y float64) engine.Entity {
	player := world.AddEntity()
	world.AddComponent(player, &collision.PositionComponent{X: x, Y: y})
	world.AddComponent(player, &collision.VelocityComponent{X: 0, Y: 0})

	collider := collision.NewCircleCollider(x, y, 5,
		collision.LayerPlayer,
		collision.LayerAll^collision.LayerPlayer)
	world.AddComponent(player, &collision.ColliderComponent{Collider: collider})

	sliding := collision.NewSlidingComponent()
	world.AddComponent(player, sliding)

	return player
}

func createWall(world *engine.World, grid *spatial.Grid, x, y, w, h float64) engine.Entity {
	wall := world.AddEntity()
	world.AddComponent(wall, &collision.PositionComponent{X: x, Y: y})

	collider := collision.NewAABBCollider(x, y, w, h,
		collision.LayerTerrain,
		collision.LayerAll)
	world.AddComponent(wall, &collision.ColliderComponent{Collider: collider})

	// Add to spatial index
	grid.Insert(wall, x+w/2, y+h/2)

	return wall
}

func setVelocity(world *engine.World, entity engine.Entity, vx, vy float64) {
	velComp, _ := world.GetComponent(entity, reflect.TypeOf(&collision.VelocityComponent{}))
	if vel, ok := velComp.(*collision.VelocityComponent); ok {
		vel.X = vx
		vel.Y = vy
	}
}

func getPosition(world *engine.World, entity engine.Entity) *collision.PositionComponent {
	posComp, _ := world.GetComponent(entity, reflect.TypeOf(&collision.PositionComponent{}))
	if pos, ok := posComp.(*collision.PositionComponent); ok {
		return pos
	}
	return &collision.PositionComponent{}
}

func updateSpatialIndex(world *engine.World, grid *spatial.Grid) {
	grid.Clear()
	posType := reflect.TypeOf(&collision.PositionComponent{})
	entities := world.Query(posType)

	for _, e := range entities {
		posComp, ok := world.GetComponent(e, posType)
		if !ok {
			continue
		}
		if pos, ok := posComp.(*collision.PositionComponent); ok {
			grid.Insert(e, pos.X, pos.Y)
		}
	}
}

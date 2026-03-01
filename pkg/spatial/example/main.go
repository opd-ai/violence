// Package main demonstrates practical usage of the spatial indexing system.
package main

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/spatial"
)

func main() {
	// Create world and spatial system
	world := engine.NewWorld()
	spatialSystem := spatial.NewSystem(64.0) // 64-unit cells

	// Register the spatial system
	world.AddSystem(spatialSystem)

	// Spawn player
	player := world.AddEntity()
	world.AddComponent(player, &engine.Position{X: 500.0, Y: 500.0})
	world.AddComponent(player, &engine.Health{Current: 100, Max: 100})

	// Spawn 100 random enemies across the map
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 100; i++ {
		enemy := world.AddEntity()
		x := rng.Float64() * 1000.0
		y := rng.Float64() * 1000.0
		world.AddComponent(enemy, &engine.Position{X: x, Y: y})
		world.AddComponent(enemy, &engine.Health{Current: 50, Max: 50})
	}

	// Update world (builds spatial index)
	world.Update()

	// Find enemies within attack range (50 units)
	playerPos := &engine.Position{X: 500.0, Y: 500.0}
	attackRange := 50.0

	nearbyEnemies := spatialSystem.QueryRadiusExact(world, playerPos.X, playerPos.Y, attackRange)

	fmt.Printf("Player at (%.1f, %.1f)\n", playerPos.X, playerPos.Y)
	fmt.Printf("Attack range: %.1f units\n", attackRange)
	fmt.Printf("Total entities: %d\n", spatialSystem.Count())
	fmt.Printf("Enemies in range: %d\n", len(nearbyEnemies)-1) // -1 for player

	// Find all entities in a screen region for rendering
	screenLeft, screenTop := 400.0, 400.0
	screenRight, screenBottom := 600.0, 600.0

	visibleEntities := spatialSystem.QueryBounds(screenLeft, screenTop, screenRight, screenBottom)

	fmt.Printf("\nScreen bounds: (%.0f,%.0f) to (%.0f,%.0f)\n", screenLeft, screenTop, screenRight, screenBottom)
	fmt.Printf("Entities to render: %d\n", len(visibleEntities))

	// Demonstrate performance benefit
	fmt.Printf("\nSpatial indexing active - queries are O(1) instead of O(n)\n")
	fmt.Printf("Grid cells occupied: %d\n", spatialSystem.CellCount())
}

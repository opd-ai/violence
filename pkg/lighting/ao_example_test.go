package lighting_test

import (
	"fmt"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/lighting"
)

// Example demonstrates basic ambient occlusion usage.
func Example_ambientOcclusion() {
	// Create a world and AO system
	world := engine.NewWorld()
	aoSystem := lighting.NewAOSystem("fantasy")

	// Register the system
	world.AddSystem(aoSystem)

	// Create an entity with position and AO components
	entity := world.AddEntity()

	pos := &lighting.PositionComponent{
		X: 10.0,
		Y: 10.0,
	}

	ao := lighting.NewAOComponent(2.0) // Sample radius of 2 units

	world.AddComponent(entity, pos)
	world.AddComponent(entity, ao)

	// Update the system to calculate occlusion
	for i := 0; i < 4; i++ { // Run enough ticks to trigger update
		aoSystem.Update(world)
	}

	// Retrieve the calculated AO values
	aoComp, _ := world.GetComponent(entity, reflect.TypeOf(&lighting.AOComponent{}))
	calculatedAO := aoComp.(*lighting.AOComponent)

	// Access directional occlusion values
	fmt.Printf("Overall occlusion: %.2f\n", calculatedAO.Overall)
	fmt.Printf("North occlusion: %.2f\n", calculatedAO.North)

	// Query occlusion at a specific angle (for rendering)
	angleEast := 0.0
	occlusionEast := calculatedAO.GetOcclusionAt(angleEast)
	fmt.Printf("Occlusion facing east: %.2f\n", occlusionEast)

	// Note: Lower values = more occluded (darker)
	//       Higher values = less occluded (lighter)
}

// Example_aoIntegration shows integration with rendering.
func Example_aoIntegration() {
	world := engine.NewWorld()
	aoSystem := lighting.NewAOSystem("horror") // Strong AO for horror genre

	world.AddSystem(aoSystem)

	// Create multiple entities
	for i := 0; i < 5; i++ {
		entity := world.AddEntity()

		pos := &lighting.PositionComponent{
			X: float64(i * 2),
			Y: float64(i * 2),
		}

		// Entities can have different sample radii
		ao := lighting.NewAOComponent(1.5)

		// Some entities don't cast occlusion (e.g., small particles)
		if i%2 == 0 {
			ao.CastsOcclusion = false
		}

		world.AddComponent(entity, pos)
		world.AddComponent(entity, ao)
	}

	// Run system updates
	for i := 0; i < 4; i++ {
		aoSystem.Update(world)
	}

	// In rendering code, use AO values to darken sprites
	aoType := reflect.TypeOf(&lighting.AOComponent{})
	entities := world.Query(aoType)

	for _, entity := range entities {
		aoComp, _ := world.GetComponent(entity, aoType)
		ao := aoComp.(*lighting.AOComponent)

		// Apply occlusion to sprite color
		// Example: multiply sprite RGB by ao.Overall
		brightness := ao.Overall // 0.0-1.0

		fmt.Printf("Entity %d brightness: %.2f\n", entity, brightness)
	}
}

// Example_aoInvalidation shows when to invalidate AO cache.
func Example_aoInvalidation() {
	world := engine.NewWorld()
	aoSystem := lighting.NewAOSystem("fantasy")
	world.AddSystem(aoSystem)

	// Create entities
	entity := world.AddEntity()
	pos := &lighting.PositionComponent{X: 5.0, Y: 5.0}
	ao := lighting.NewAOComponent(2.0)

	world.AddComponent(entity, pos)
	world.AddComponent(entity, ao)

	// Initial calculation
	for i := 0; i < 4; i++ {
		aoSystem.Update(world)
	}

	// Entity moves — invalidate its AO
	pos.X = 10.0
	pos.Y = 10.0
	ao.Invalidate()

	// Next update will recalculate
	for i := 0; i < 4; i++ {
		aoSystem.Update(world)
	}

	// Map changes significantly (e.g., wall destroyed) — invalidate all
	aoSystem.InvalidateAll(world)

	// All entities will recalculate on next update
	for i := 0; i < 4; i++ {
		aoSystem.Update(world)
	}

	fmt.Println("AO recalculated after invalidation")
}

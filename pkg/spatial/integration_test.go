package spatial

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// TestIntegration_SpatialWithECS demonstrates complete integration with the ECS.
func TestIntegration_SpatialWithECS(t *testing.T) {
	// Setup world with spatial system
	world := engine.NewWorld()
	spatialSys := NewSystem(32.0)
	world.AddSystem(spatialSys)

	// Create player entity at center
	player := world.AddEntity()
	world.AddComponent(player, &engine.Position{X: 100.0, Y: 100.0})
	world.AddComponent(player, &engine.Health{Current: 100, Max: 100})

	// Create nearby enemies
	enemy1 := world.AddEntity()
	world.AddComponent(enemy1, &engine.Position{X: 110.0, Y: 105.0})
	world.AddComponent(enemy1, &engine.Health{Current: 50, Max: 50})

	enemy2 := world.AddEntity()
	world.AddComponent(enemy2, &engine.Position{X: 95.0, Y: 98.0})
	world.AddComponent(enemy2, &engine.Health{Current: 50, Max: 50})

	// Create distant enemy
	enemy3 := world.AddEntity()
	world.AddComponent(enemy3, &engine.Position{X: 500.0, Y: 500.0})
	world.AddComponent(enemy3, &engine.Health{Current: 50, Max: 50})

	// Update world (including spatial system)
	world.Update()

	// Query nearby entities (within 20 units of player)
	nearby := spatialSys.QueryRadiusExact(world, 100.0, 100.0, 20.0)

	// Should find player and two nearby enemies
	if len(nearby) != 3 {
		t.Errorf("expected 3 nearby entities, got %d", len(nearby))
	}

	// Verify distant enemy is not in results
	foundPlayer, foundEnemy1, foundEnemy2, foundEnemy3 := false, false, false, false
	for _, e := range nearby {
		switch e {
		case player:
			foundPlayer = true
		case enemy1:
			foundEnemy1 = true
		case enemy2:
			foundEnemy2 = true
		case enemy3:
			foundEnemy3 = true
		}
	}

	if !foundPlayer {
		t.Error("player not found in nearby query")
	}
	if !foundEnemy1 {
		t.Error("enemy1 not found in nearby query")
	}
	if !foundEnemy2 {
		t.Error("enemy2 not found in nearby query")
	}
	if foundEnemy3 {
		t.Error("enemy3 should not be in nearby query")
	}
}

// TestIntegration_DynamicUpdates tests spatial index updates as entities move.
func TestIntegration_DynamicUpdates(t *testing.T) {
	world := engine.NewWorld()
	spatialSys := NewSystem(32.0)
	world.AddSystem(spatialSys)

	// Create entity
	e := world.AddEntity()
	pos := &engine.Position{X: 100.0, Y: 100.0}
	world.AddComponent(e, pos)

	// Initial update
	world.Update()

	// Query at initial position
	results := spatialSys.QueryRadius(100.0, 100.0, 10.0)
	if len(results) != 1 {
		t.Fatalf("expected 1 entity at initial position, got %d", len(results))
	}

	// Move entity far away
	pos.X = 500.0
	pos.Y = 500.0

	// Update world (spatial index should rebuild)
	world.Update()

	// Query at old position (should find nothing)
	results = spatialSys.QueryRadius(100.0, 100.0, 10.0)
	if len(results) != 0 {
		t.Errorf("expected 0 entities at old position, got %d", len(results))
	}

	// Query at new position (should find entity)
	results = spatialSys.QueryRadius(500.0, 500.0, 10.0)
	if len(results) != 1 {
		t.Errorf("expected 1 entity at new position, got %d", len(results))
	}
}

// TestIntegration_LargeWorldPerformance tests performance with many entities.
func TestIntegration_LargeWorldPerformance(t *testing.T) {
	world := engine.NewWorld()
	spatialSys := NewSystem(64.0)
	world.AddSystem(spatialSys)

	// Create 1000 entities across a 1000x1000 world
	for i := 0; i < 1000; i++ {
		e := world.AddEntity()
		world.AddComponent(e, &engine.Position{
			X: float64(i%100) * 10.0,
			Y: float64(i/100) * 10.0,
		})
	}

	// Update to build spatial index
	world.Update()

	if spatialSys.Count() != 1000 {
		t.Errorf("expected 1000 indexed entities, got %d", spatialSys.Count())
	}

	// Query should be fast even with 1000 entities
	results := spatialSys.QueryRadius(50.0, 50.0, 100.0)

	// Verify we're not getting all entities (spatial filtering working)
	if len(results) >= 1000 {
		t.Error("spatial filtering not working - got all or more entities")
	}

	// Should find some entities in this region
	if len(results) < 10 {
		t.Errorf("expected at least 10 entities near (50,50) with radius 100, got %d", len(results))
	}
}

// BenchmarkIntegration_WorldUpdate benchmarks the cost of spatial system integration.
func BenchmarkIntegration_WorldUpdate(b *testing.B) {
	world := engine.NewWorld()
	spatialSys := NewSystem(64.0)
	world.AddSystem(spatialSys)

	// Create 100 entities
	for i := 0; i < 100; i++ {
		e := world.AddEntity()
		world.AddComponent(e, &engine.Position{
			X: float64(i%10) * 10.0,
			Y: float64(i/10) * 10.0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		world.Update()
	}
}

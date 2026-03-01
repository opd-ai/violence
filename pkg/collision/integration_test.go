package collision

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

// TestSystemIntegration verifies collision system integrates with ECS.
func TestSystemIntegration(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem()

	// Add system to world
	w.AddSystem(sys)

	// Create two entities with colliders
	e1 := w.AddEntity()
	e2 := w.AddEntity()

	// Add circle colliders
	c1 := NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy)
	c2 := NewCircleCollider(1.5, 0, 1, LayerEnemy, LayerPlayer)

	AddColliderToEntity(w, e1, c1)
	AddColliderToEntity(w, e2, c2)

	// Verify colliders were added
	retrieved1 := GetEntityCollider(w, e1)
	if retrieved1 == nil {
		t.Fatal("Failed to retrieve collider from entity 1")
	}

	retrieved2 := GetEntityCollider(w, e2)
	if retrieved2 == nil {
		t.Fatal("Failed to retrieve collider from entity 2")
	}

	// Test collision between entities
	if !TestCollision(retrieved1, retrieved2) {
		t.Error("Expected entities to collide")
	}

	// Run system update (should not panic)
	w.Update()
}

// TestQueryCollisions verifies collision queries work with ECS.
func TestQueryCollisions(t *testing.T) {
	w := engine.NewWorld()

	// Create multiple entities
	e1 := w.AddEntity()
	e2 := w.AddEntity()
	e3 := w.AddEntity()

	// e1 and e2 are close, e3 is far away
	AddColliderToEntity(w, e1, NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy))
	AddColliderToEntity(w, e2, NewCircleCollider(1.5, 0, 1, LayerEnemy, LayerPlayer))
	AddColliderToEntity(w, e3, NewCircleCollider(100, 100, 1, LayerEnemy, LayerPlayer))

	// Query what e1 collides with
	e1Collider := GetEntityCollider(w, e1)
	hits := QueryCollisions(w, e1Collider)

	// Should hit e2 but not e3 (and not itself)
	foundE2 := false
	foundE3 := false
	for _, hit := range hits {
		if hit == e2 {
			foundE2 = true
		}
		if hit == e3 {
			foundE3 = true
		}
	}

	if !foundE2 {
		t.Error("Expected to find e2 in collision query")
	}
	if foundE3 {
		t.Error("Should not find e3 (too far away)")
	}
}

// TestQueryCollisionsInRadius verifies radius queries.
func TestQueryCollisionsInRadius(t *testing.T) {
	w := engine.NewWorld()

	// Create entities at various positions
	e1 := w.AddEntity()
	e2 := w.AddEntity()
	e3 := w.AddEntity()

	AddColliderToEntity(w, e1, NewCircleCollider(5, 0, 0.5, LayerEnemy, LayerPlayer))
	AddColliderToEntity(w, e2, NewCircleCollider(0, 5, 0.5, LayerEnemy, LayerPlayer))
	AddColliderToEntity(w, e3, NewCircleCollider(50, 50, 0.5, LayerEnemy, LayerPlayer))

	// Query 10 unit radius from origin
	hits := QueryCollisionsInRadius(w, 0, 0, 10, LayerPlayer)

	// Should find e1 and e2, but not e3
	if len(hits) < 2 {
		t.Errorf("Expected at least 2 hits in radius, got %d", len(hits))
	}

	foundE3 := false
	for _, hit := range hits {
		if hit == e3 {
			foundE3 = true
		}
	}
	if foundE3 {
		t.Error("Should not find e3 (outside radius)")
	}
}

// TestColliderUpdate verifies collider position updates.
func TestColliderUpdate(t *testing.T) {
	w := engine.NewWorld()
	e := w.AddEntity()

	collider := NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy)
	AddColliderToEntity(w, e, collider)

	// Update position
	UpdateColliderPosition(collider, 10, 10)

	retrieved := GetEntityCollider(w, e)
	if retrieved.X != 10 || retrieved.Y != 10 {
		t.Errorf("Expected position (10,10), got (%v,%v)", retrieved.X, retrieved.Y)
	}
}

// TestCapsuleUpdate verifies capsule collider updates.
func TestCapsuleUpdate(t *testing.T) {
	capsule := NewCapsuleCollider(0, 0, 10, 0, 1, LayerPlayer, LayerEnemy)

	// Update endpoints
	UpdateCapsuleCollider(capsule, 5, 5, 15, 5)

	if capsule.X != 5 || capsule.Y != 5 {
		t.Errorf("Expected start (5,5), got (%v,%v)", capsule.X, capsule.Y)
	}

	if capsule.X2 != 15 || capsule.Y2 != 5 {
		t.Errorf("Expected end (15,5), got (%v,%v)", capsule.X2, capsule.Y2)
	}
}

// TestLayerMaskingInECS verifies layer-based collision filtering in ECS context.
func TestLayerMaskingInECS(t *testing.T) {
	w := engine.NewWorld()

	// Player entity
	player := w.AddEntity()
	AddColliderToEntity(w, player, NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy|LayerTerrain))

	// Enemy entity
	enemy := w.AddEntity()
	AddColliderToEntity(w, enemy, NewCircleCollider(1, 0, 1, LayerEnemy, LayerPlayer))

	// Terrain entity (blocks all except ethereal)
	terrain := w.AddEntity()
	AddColliderToEntity(w, terrain, NewCircleCollider(0, 0, 1, LayerTerrain, LayerPlayer|LayerEnemy|LayerProjectile))

	// Ethereal entity (interacts with nothing)
	ethereal := w.AddEntity()
	AddColliderToEntity(w, ethereal, NewCircleCollider(0, 0, 1, LayerEthereal, LayerNone))

	playerCol := GetEntityCollider(w, player)
	enemyCol := GetEntityCollider(w, enemy)
	terrainCol := GetEntityCollider(w, terrain)
	etherealCol := GetEntityCollider(w, ethereal)

	// Player should collide with enemy and terrain
	if !TestCollision(playerCol, enemyCol) {
		t.Error("Player should collide with enemy")
	}
	if !TestCollision(playerCol, terrainCol) {
		t.Error("Player should collide with terrain")
	}

	// Ethereal should not collide with anything
	if TestCollision(etherealCol, playerCol) {
		t.Error("Ethereal should not collide with player")
	}
	if TestCollision(etherealCol, terrainCol) {
		t.Error("Ethereal should not collide with terrain")
	}
}

// BenchmarkECSCollisionQuery benchmarks collision queries in ECS context.
func BenchmarkECSCollisionQuery(b *testing.B) {
	w := engine.NewWorld()

	// Create 100 entities with colliders
	for i := 0; i < 100; i++ {
		e := w.AddEntity()
		x := float64(i % 10)
		y := float64(i / 10)
		AddColliderToEntity(w, e, NewCircleCollider(x, y, 0.5, LayerEnemy, LayerPlayer))
	}

	queryCol := NewCircleCollider(5, 5, 2, LayerPlayer, LayerEnemy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = QueryCollisions(w, queryCol)
	}
}

// BenchmarkRadiusQuery benchmarks radius-based queries.
func BenchmarkRadiusQuery(b *testing.B) {
	w := engine.NewWorld()

	// Create 100 entities
	for i := 0; i < 100; i++ {
		e := w.AddEntity()
		x := float64(i % 10)
		y := float64(i / 10)
		AddColliderToEntity(w, e, NewCircleCollider(x, y, 0.5, LayerEnemy, LayerPlayer))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = QueryCollisionsInRadius(w, 5, 5, 3, LayerPlayer)
	}
}

// TestComponentReflection verifies component type is correct for ECS queries.
func TestComponentReflection(t *testing.T) {
	w := engine.NewWorld()
	e := w.AddEntity()

	collider := NewCircleCollider(0, 0, 1, LayerPlayer, LayerEnemy)
	comp := &ColliderComponent{Collider: collider}
	w.AddComponent(e, comp)

	// Query using reflection
	colliderType := reflect.TypeOf(&ColliderComponent{})
	entities := w.Query(colliderType)

	if len(entities) != 1 {
		t.Fatalf("Expected 1 entity with ColliderComponent, got %d", len(entities))
	}

	if entities[0] != e {
		t.Error("Query returned wrong entity")
	}
}

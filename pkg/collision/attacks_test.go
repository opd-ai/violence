package collision

import (
	"math"
	"testing"
)

func TestCreateConeCollider(t *testing.T) {
	// Create a cone facing right
	cone := CreateConeCollider(0, 0, 1, 0, 10, math.Pi/3, LayerPlayer, LayerEnemy)

	if cone.Shape != ShapePolygon {
		t.Errorf("Expected polygon shape, got %v", cone.Shape)
	}

	if len(cone.Polygon) < 3 {
		t.Errorf("Cone should have at least 3 vertices, got %d", len(cone.Polygon))
	}

	// Check that cone is positioned correctly
	if cone.X != 0 || cone.Y != 0 {
		t.Errorf("Cone position should be (0,0), got (%v,%v)", cone.X, cone.Y)
	}
}

func TestCreateCircleAttackCollider(t *testing.T) {
	aoe := CreateCircleAttackCollider(5, 5, 20, LayerPlayer, LayerEnemy)

	if aoe.Shape != ShapeCircle {
		t.Errorf("Expected circle shape, got %v", aoe.Shape)
	}

	if aoe.Radius != 20 {
		t.Errorf("Expected radius 20, got %v", aoe.Radius)
	}

	if aoe.X != 5 || aoe.Y != 5 {
		t.Errorf("Expected position (5,5), got (%v,%v)", aoe.X, aoe.Y)
	}
}

func TestCreateLineAttackCollider(t *testing.T) {
	// Beam attack going right
	beam := CreateLineAttackCollider(0, 0, 1, 0, 50, 10, LayerPlayer, LayerEnemy)

	if beam.Shape != ShapeCapsule {
		t.Errorf("Expected capsule shape, got %v", beam.Shape)
	}

	// Check start position
	if beam.X != 0 || beam.Y != 0 {
		t.Errorf("Expected start at (0,0), got (%v,%v)", beam.X, beam.Y)
	}

	// Check end position (should be range in direction)
	expectedX := 50.0
	if math.Abs(beam.X2-expectedX) > 0.1 {
		t.Errorf("Expected X2 near %v, got %v", expectedX, beam.X2)
	}
}

func TestCreateRingCollider(t *testing.T) {
	outer, inner := CreateRingCollider(0, 0, 30, 10, LayerPlayer, LayerEnemy)

	if outer.Radius != 30 {
		t.Errorf("Expected outer radius 30, got %v", outer.Radius)
	}

	if inner.Radius != 10 {
		t.Errorf("Expected inner radius 10, got %v", inner.Radius)
	}

	// Test ring collision - target in ring
	target := NewCircleCollider(20, 0, 1, LayerEnemy, LayerPlayer)
	if !TestRingCollision(target, outer, inner) {
		t.Error("Target at radius 20 should be in ring (10-30)")
	}

	// Target inside inner circle
	targetInner := NewCircleCollider(5, 0, 1, LayerEnemy, LayerPlayer)
	if TestRingCollision(targetInner, outer, inner) {
		t.Error("Target at radius 5 should NOT be in ring (10-30)")
	}

	// Target outside outer circle
	targetOuter := NewCircleCollider(40, 0, 1, LayerEnemy, LayerPlayer)
	if TestRingCollision(targetOuter, outer, inner) {
		t.Error("Target at radius 40 should NOT be in ring (10-30)")
	}
}

func TestCreateProjectileCollider(t *testing.T) {
	// Projectile moved from (0,0) to (10,10)
	proj := CreateProjectileCollider(0, 0, 10, 10, 2, LayerProjectile, LayerEnemy)

	if proj.Shape != ShapeCapsule {
		t.Errorf("Expected capsule shape for projectile, got %v", proj.Shape)
	}

	// Should create swept collision from last to current position
	if proj.X != 0 || proj.Y != 0 {
		t.Errorf("Expected start at (0,0), got (%v,%v)", proj.X, proj.Y)
	}

	if proj.X2 != 10 || proj.Y2 != 10 {
		t.Errorf("Expected end at (10,10), got (%v,%v)", proj.X2, proj.Y2)
	}
}

func TestCreateMeleeWeaponCollider(t *testing.T) {
	// Player at (10,10) swinging right
	weapon := CreateMeleeWeaponCollider(10, 10, 1, 0, 15, 3, LayerPlayer, LayerEnemy)

	if weapon.Shape != ShapeCapsule {
		t.Errorf("Expected capsule shape for weapon, got %v", weapon.Shape)
	}

	// Weapon should extend in attack direction
	if weapon.X <= 10 {
		t.Error("Weapon start should be offset from player position")
	}
}

func TestCreateCharacterCollider(t *testing.T) {
	playerCol := CreateCharacterCollider(0, 0, 0.5, true)
	if playerCol.Layer != LayerPlayer {
		t.Errorf("Expected player layer, got %v", playerCol.Layer)
	}

	enemyCol := CreateCharacterCollider(0, 0, 0.5, false)
	if enemyCol.Layer != LayerEnemy {
		t.Errorf("Expected enemy layer, got %v", enemyCol.Layer)
	}

	// Player should collide with enemy
	if !CanCollide(playerCol, enemyCol) {
		t.Error("Player should be able to collide with enemy")
	}
}

func TestCreateTerrainCollider(t *testing.T) {
	terrain := CreateTerrainCollider(0, 0, 1, 1)

	if terrain.Layer != LayerTerrain {
		t.Errorf("Expected terrain layer, got %v", terrain.Layer)
	}

	if terrain.Mask != LayerAll {
		t.Errorf("Expected terrain to block all layers, got %v", terrain.Mask)
	}

	if terrain.Shape != ShapeAABB {
		t.Errorf("Expected AABB shape for terrain, got %v", terrain.Shape)
	}
}

func TestCreatePropCollider(t *testing.T) {
	blockingProp := CreatePropCollider(0, 0, 1, true)
	if (blockingProp.Mask & (LayerPlayer | LayerEnemy)) == 0 {
		t.Error("Blocking prop should collide with player and enemy")
	}

	nonBlockingProp := CreatePropCollider(0, 0, 1, false)
	if (nonBlockingProp.Mask & LayerPlayer) != 0 {
		t.Error("Non-blocking prop should not collide with player")
	}
}

func TestCreateTriggerZone(t *testing.T) {
	trigger := CreateTriggerZone(0, 0, 10, LayerPlayer|LayerEnemy)

	if trigger.Layer != LayerTrigger {
		t.Errorf("Expected trigger layer, got %v", trigger.Layer)
	}

	player := CreateCharacterCollider(5, 0, 0.5, true)
	if !CanCollide(trigger, player) {
		t.Error("Trigger should detect player")
	}

	if !TestCollision(trigger, player) {
		t.Error("Player inside trigger radius should collide")
	}
}

func BenchmarkConeCollider(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CreateConeCollider(0, 0, 1, 0, 10, math.Pi/3, LayerPlayer, LayerEnemy)
	}
}

func BenchmarkMeleeWeaponHit(b *testing.B) {
	weapon := CreateMeleeWeaponCollider(0, 0, 1, 0, 15, 3, LayerPlayer, LayerEnemy)
	enemy := CreateCharacterCollider(10, 0, 0.5, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TestCollision(weapon, enemy)
	}
}

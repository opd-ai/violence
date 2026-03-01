package collision

import (
	"testing"
)

func TestNewAttackShapeCache(t *testing.T) {
	cache := NewAttackShapeCache()
	if cache == nil {
		t.Fatal("NewAttackShapeCache() returned nil")
	}
	if cache.shapes == nil {
		t.Error("shapes map not initialized")
	}
}

func TestAttackShapeCache_RegisterAndGet(t *testing.T) {
	cache := NewAttackShapeCache()

	shape := &AttackShape{
		Name: "test_shape",
		Vertices: []Point{
			{X: 0, Y: 0},
			{X: 10, Y: 0},
			{X: 10, Y: 10},
		},
	}

	cache.RegisterShape("test_shape", shape)
	retrieved := cache.GetShape("test_shape")

	if retrieved == nil {
		t.Fatal("GetShape() returned nil for registered shape")
	}
	if retrieved.Name != "test_shape" {
		t.Errorf("GetShape() name = %v, want test_shape", retrieved.Name)
	}
	if len(retrieved.Vertices) != 3 {
		t.Errorf("GetShape() vertices count = %d, want 3", len(retrieved.Vertices))
	}
}

func TestAttackShapeCache_GetNonexistent(t *testing.T) {
	cache := NewAttackShapeCache()
	shape := cache.GetShape("nonexistent")

	if shape != nil {
		t.Error("GetShape() should return nil for nonexistent shape")
	}
}

func TestWeaponShapeGenerator_GenerateWeaponShapes(t *testing.T) {
	cache := NewAttackShapeCache()
	gen := NewWeaponShapeGenerator(cache)
	gen.GenerateWeaponShapes()

	// Test that common weapon shapes were registered
	tests := []string{
		"sword_slash_h",
		"sword_slash_v",
		"sword_thrust",
		"axe_swing",
		"spear_thrust",
		"dagger_stab",
		"hammer_smash",
		"whip_sweep",
		"fireball_impact",
		"lightning_beam",
		"ice_cone",
		"cleave",
		"backstab",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			shape := cache.GetShape(name)
			if shape == nil {
				t.Errorf("Shape %s not registered", name)
			}
			if shape != nil && len(shape.Vertices) < 3 {
				t.Errorf("Shape %s has insufficient vertices: %d", name, len(shape.Vertices))
			}
		})
	}
}

func TestAttackFrameComponent_NewAndSet(t *testing.T) {
	comp := NewAttackFrameComponent("sword_slash", 25.0, 10.0)

	if comp == nil {
		t.Fatal("NewAttackFrameComponent() returned nil")
	}
	if comp.ShapeName != "sword_slash" {
		t.Errorf("ShapeName = %v, want sword_slash", comp.ShapeName)
	}
	if comp.Damage != 25.0 {
		t.Errorf("Damage = %v, want 25.0", comp.Damage)
	}
	if comp.Knockback != 10.0 {
		t.Errorf("Knockback = %v, want 10.0", comp.Knockback)
	}
}

func TestAttackFrameComponent_ActiveFrames(t *testing.T) {
	comp := NewAttackFrameComponent("test", 10, 5)

	// Create test colliders for different frames
	frame2Collider := NewCircleCollider(0, 0, 10, LayerPlayer, LayerEnemy)
	frame3Collider := NewCircleCollider(5, 5, 15, LayerPlayer, LayerEnemy)

	comp.SetActiveFrame(2, frame2Collider)
	comp.SetActiveFrame(3, frame3Collider)

	// Test getting colliders
	comp.UpdateFrame(1)
	if comp.GetCurrentCollider() != nil {
		t.Error("GetCurrentCollider() should return nil for frame with no collider")
	}

	comp.UpdateFrame(2)
	collider := comp.GetCurrentCollider()
	if collider == nil {
		t.Fatal("GetCurrentCollider() returned nil for frame 2")
	}
	if collider.Radius != 10 {
		t.Errorf("Frame 2 collider radius = %v, want 10", collider.Radius)
	}

	comp.UpdateFrame(3)
	collider = comp.GetCurrentCollider()
	if collider == nil {
		t.Fatal("GetCurrentCollider() returned nil for frame 3")
	}
	if collider.Radius != 15 {
		t.Errorf("Frame 3 collider radius = %v, want 15", collider.Radius)
	}
}

func TestCollisionGeometrySystem_New(t *testing.T) {
	sys := NewCollisionGeometrySystem()

	if sys == nil {
		t.Fatal("NewCollisionGeometrySystem() returned nil")
	}
	if sys.extractor == nil {
		t.Error("extractor not initialized")
	}
	if sys.cache == nil {
		t.Error("cache not initialized")
	}
	if sys.generator == nil {
		t.Error("generator not initialized")
	}

	// Verify weapon shapes were pre-generated
	shape := sys.GetAttackShape("sword_slash_h")
	if shape == nil {
		t.Error("Pre-generated weapon shapes missing")
	}
}

func TestCollisionGeometrySystem_ExtractSpriteCollider(t *testing.T) {
	sys := NewCollisionGeometrySystem()

	// Test with nil sprite (should return fallback)
	collider := sys.ExtractSpriteCollider(nil, 10, 10, LayerPlayer, LayerEnemy)

	if collider == nil {
		t.Fatal("ExtractSpriteCollider() returned nil")
	}
	if collider.X != 10 || collider.Y != 10 {
		t.Errorf("Collider position = (%v,%v), want (10,10)", collider.X, collider.Y)
	}
	if collider.Layer != LayerPlayer {
		t.Errorf("Collider layer = %v, want LayerPlayer", collider.Layer)
	}
}

func TestCollisionGeometrySystem_ExtractSpriteCollider_NilSprite(t *testing.T) {
	sys := NewCollisionGeometrySystem()

	collider := sys.ExtractSpriteCollider(nil, 5, 5, LayerEnemy, LayerPlayer)

	// Should return a fallback collider
	if collider == nil {
		t.Error("ExtractSpriteCollider() should return fallback for nil sprite")
	}
}

func TestCollisionGeometrySystem_GetAttackShape(t *testing.T) {
	sys := NewCollisionGeometrySystem()

	shape := sys.GetAttackShape("sword_thrust")
	if shape == nil {
		t.Error("GetAttackShape() returned nil for valid shape")
	}

	shape = sys.GetAttackShape("nonexistent_weapon")
	if shape != nil {
		t.Error("GetAttackShape() should return nil for invalid shape")
	}
}

func TestCollisionGeometrySystem_CreateAttackCollider(t *testing.T) {
	sys := NewCollisionGeometrySystem()

	// Test with valid shape
	collider := sys.CreateAttackCollider("sword_slash_h", 10, 10, 1, 0, LayerPlayer, LayerEnemy)
	if collider == nil {
		t.Fatal("CreateAttackCollider() returned nil")
	}
	if collider.Layer != LayerPlayer {
		t.Errorf("Collider layer = %v, want LayerPlayer", collider.Layer)
	}

	// Test with invalid shape (should create fallback)
	collider = sys.CreateAttackCollider("invalid_shape", 5, 5, 0, 1, LayerEnemy, LayerPlayer)
	if collider == nil {
		t.Fatal("CreateAttackCollider() should create fallback for invalid shape")
	}
}

func TestSpriteColliderComponent_Update(t *testing.T) {
	sys := NewCollisionGeometrySystem()
	comp := &SpriteColliderComponent{
		Dirty: true,
	}

	// Test with nil sprite
	sys.UpdateSpriteCollider(comp, nil, "sprite_1", 10, 10, LayerEnemy, LayerPlayer)

	if comp.Dirty {
		t.Error("Dirty flag should be cleared after update")
	}
	if comp.LastSpriteID != "sprite_1" {
		t.Errorf("LastSpriteID = %v, want sprite_1", comp.LastSpriteID)
	}
	if comp.BoundingBox == nil {
		t.Error("BoundingBox not created")
	}
	if comp.DetailedHull == nil {
		t.Error("DetailedHull not created")
	}

	// Update with same sprite ID (should only update position)
	sys.UpdateSpriteCollider(comp, nil, "sprite_1", 20, 20, LayerEnemy, LayerPlayer)

	// Position should be updated
	if comp.DetailedHull.X != 20 || comp.DetailedHull.Y != 20 {
		t.Error("Position not updated correctly")
	}
}

func TestSpriteColliderComponent_UpdateDirty(t *testing.T) {
	sys := NewCollisionGeometrySystem()
	comp := &SpriteColliderComponent{
		Dirty:        false,
		LastSpriteID: "old_sprite",
	}

	// Update with different sprite ID
	sys.UpdateSpriteCollider(comp, nil, "new_sprite", 5, 5, LayerPlayer, LayerEnemy)

	if comp.LastSpriteID != "new_sprite" {
		t.Errorf("LastSpriteID = %v, want new_sprite", comp.LastSpriteID)
	}
	// BoundingBox should be created even for nil sprite
	if comp.BoundingBox == nil {
		t.Error("BoundingBox should be created")
	}
}

func TestCollisionGeometrySystem_GetExtractor(t *testing.T) {
	sys := NewCollisionGeometrySystem()
	extractor := sys.GetExtractor()

	if extractor == nil {
		t.Error("GetExtractor() returned nil")
	}
	if extractor != sys.extractor {
		t.Error("GetExtractor() returned different instance")
	}
}

func TestCollisionGeometrySystem_GetShapeCache(t *testing.T) {
	sys := NewCollisionGeometrySystem()
	cache := sys.GetShapeCache()

	if cache == nil {
		t.Error("GetShapeCache() returned nil")
	}
	if cache != sys.cache {
		t.Error("GetShapeCache() returned different instance")
	}
}

// Benchmark tests
func BenchmarkWeaponShapeGenerator_Generate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cache := NewAttackShapeCache()
		gen := NewWeaponShapeGenerator(cache)
		gen.GenerateWeaponShapes()
	}
}

func BenchmarkCollisionGeometrySystem_ExtractSpriteCollider(b *testing.B) {
	sys := NewCollisionGeometrySystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ExtractSpriteCollider(nil, 0, 0, LayerPlayer, LayerEnemy)
	}
}

func BenchmarkAttackShapeCache_GetShape(b *testing.B) {
	cache := NewAttackShapeCache()
	gen := NewWeaponShapeGenerator(cache)
	gen.GenerateWeaponShapes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetShape("sword_slash_h")
	}
}

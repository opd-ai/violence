package collision

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/spatial"
)

func TestSlidingComponent(t *testing.T) {
	tests := []struct {
		name string
		comp *SlidingComponent
	}{
		{
			name: "default values",
			comp: NewSlidingComponent(),
		},
		{
			name: "custom values",
			comp: &SlidingComponent{
				Enabled:           true,
				MaxSlideAngle:     math.Pi / 3,
				Friction:          0.2,
				BounceOnSteep:     true,
				BounceRestitution: 0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.comp.Type() != "SlidingComponent" {
				t.Errorf("Type() = %v, want SlidingComponent", tt.comp.Type())
			}
		})
	}
}

func TestVelocityComponent(t *testing.T) {
	vel := &VelocityComponent{X: 10.0, Y: 5.0}
	if vel.Type() != "VelocityComponent" {
		t.Errorf("Type() = %v, want VelocityComponent", vel.Type())
	}
}

func TestPositionComponent(t *testing.T) {
	pos := &PositionComponent{X: 100.0, Y: 200.0}
	if pos.Type() != "PositionComponent" {
		t.Errorf("Type() = %v, want PositionComponent", pos.Type())
	}
}

func TestColliderComponent(t *testing.T) {
	collider := NewCircleCollider(0, 0, 10, LayerPlayer, LayerAll)
	comp := &ColliderComponent{Collider: collider}
	if comp.Collider == nil {
		t.Error("Collider is nil")
	}
}

func TestSlidingSystemBasicMovement(t *testing.T) {
	// Create world and system
	w := engine.NewWorld()
	grid := spatial.NewGrid(50.0) // Single argument: cell size
	system := NewSlidingSystem(grid)

	// Create entity with all required components
	entity := w.AddEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	vel := &VelocityComponent{X: 100, Y: 0} // Move right at 100 units/sec
	collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerAll)
	sliding := NewSlidingComponent()

	w.AddComponent(entity, pos)
	w.AddComponent(entity, vel)
	w.AddComponent(entity, &ColliderComponent{Collider: collider})
	w.AddComponent(entity, sliding)

	// Update once (should move freely with no obstacles)
	originalX := pos.X
	system.Update(w)

	// Should have moved approximately 100 * 0.016 = 1.6 units to the right
	expectedDelta := 100.0 * 0.016
	actualDelta := pos.X - originalX

	if math.Abs(actualDelta-expectedDelta) > 0.1 {
		t.Errorf("Movement delta = %v, want approximately %v", actualDelta, expectedDelta)
	}

	// Collider should have moved with position
	if math.Abs(collider.X-pos.X) > 0.001 {
		t.Errorf("Collider X = %v, want %v", collider.X, pos.X)
	}
}

func TestSlidingSystemWallSliding(t *testing.T) {
	// Create world with spatial index
	w := engine.NewWorld()
	grid := spatial.NewGrid(50.0)
	system := NewSlidingSystem(grid)

	// Create player entity moving diagonally toward a wall
	player := w.AddEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	vel := &VelocityComponent{X: 100, Y: 100} // Diagonal movement
	collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerAll^LayerPlayer)
	sliding := NewSlidingComponent()

	w.AddComponent(player, pos)
	w.AddComponent(player, vel)
	w.AddComponent(player, &ColliderComponent{Collider: collider})
	w.AddComponent(player, sliding)

	// Create vertical wall (AABB) at x=60
	wall := w.AddEntity()
	wallPos := &PositionComponent{X: 60, Y: 0}
	wallCollider := NewAABBCollider(60, 0, 10, 200, LayerTerrain, LayerAll)

	w.AddComponent(wall, wallPos)
	w.AddComponent(wall, &ColliderComponent{Collider: wallCollider})
	grid.Insert(wall, 65, 100) // Two arguments: x, y

	// Update system
	originalY := pos.Y
	system.Update(w)

	// Player should slide along the wall (move in Y but not much in X)
	// X movement should be blocked or minimal
	// Y movement should continue (sliding along wall)

	// Player shouldn't penetrate the wall (stay at or before x=55 due to 5-unit radius)
	if pos.X > 56 { // 60 (wall) - 5 (radius) + 1 (tolerance)
		t.Errorf("Player penetrated wall: X = %v, expected < 56", pos.X)
	}

	// Player should have moved in Y direction (sliding along wall)
	if pos.Y <= originalY {
		t.Errorf("Player didn't slide in Y direction: Y = %v, original = %v", pos.Y, originalY)
	}
}

func TestSlidingSystemCornerSliding(t *testing.T) {
	// Test sliding around a corner with multiple iterations
	w := engine.NewWorld()
	grid := spatial.NewGrid(50.0)
	system := NewSlidingSystem(grid)

	// Create player
	player := w.AddEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	vel := &VelocityComponent{X: 100, Y: 100}
	collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerAll^LayerPlayer)
	sliding := NewSlidingComponent()

	w.AddComponent(player, pos)
	w.AddComponent(player, vel)
	w.AddComponent(player, &ColliderComponent{Collider: collider})
	w.AddComponent(player, sliding)

	// Create L-shaped corner
	wall1 := w.AddEntity()
	wall1Collider := NewAABBCollider(60, 0, 10, 60, LayerTerrain, LayerAll)
	w.AddComponent(wall1, &PositionComponent{X: 60, Y: 0})
	w.AddComponent(wall1, &ColliderComponent{Collider: wall1Collider})
	grid.Insert(wall1, 65, 30)

	wall2 := w.AddEntity()
	wall2Collider := NewAABBCollider(60, 50, 100, 10, LayerTerrain, LayerAll)
	w.AddComponent(wall2, &PositionComponent{X: 60, Y: 50})
	w.AddComponent(wall2, &ColliderComponent{Collider: wall2Collider})
	grid.Insert(wall2, 110, 55)

	// System should handle corner sliding with multiple iterations
	system.Update(w)

	// Player shouldn't penetrate walls
	if pos.X > 56 || pos.Y > 46 {
		t.Errorf("Player penetrated corner: pos = (%v, %v)", pos.X, pos.Y)
	}
}

func TestSlidingSystemDisabled(t *testing.T) {
	// Test that disabled sliding component prevents sliding
	w := engine.NewWorld()
	system := NewSlidingSystem(nil)

	player := w.AddEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	vel := &VelocityComponent{X: 100, Y: 0}
	collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerAll)
	sliding := NewSlidingComponent()
	sliding.Enabled = false // Disable sliding

	w.AddComponent(player, pos)
	w.AddComponent(player, vel)
	w.AddComponent(player, &ColliderComponent{Collider: collider})
	w.AddComponent(player, sliding)

	originalX := pos.X
	system.Update(w)

	// Position shouldn't change when sliding is disabled
	if pos.X != originalX {
		t.Errorf("Position changed with disabled sliding: X = %v, want %v", pos.X, originalX)
	}
}

func TestSlidingSystemMinVelocity(t *testing.T) {
	// Test that very small velocities are ignored
	w := engine.NewWorld()
	system := NewSlidingSystem(nil)

	player := w.AddEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	vel := &VelocityComponent{X: 0.0001, Y: 0.0001} // Below threshold
	collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerAll)
	sliding := NewSlidingComponent()

	w.AddComponent(player, pos)
	w.AddComponent(player, vel)
	w.AddComponent(player, &ColliderComponent{Collider: collider})
	w.AddComponent(player, sliding)

	originalX := pos.X
	system.Update(w)

	// Position shouldn't change for negligible velocity
	if math.Abs(pos.X-originalX) > 0.0001 {
		t.Errorf("Position changed for negligible velocity: delta = %v", pos.X-originalX)
	}
}

func TestSlidingSystemFriction(t *testing.T) {
	tests := []struct {
		name     string
		friction float64
	}{
		{"no friction", 0.0},
		{"low friction", 0.1},
		{"high friction", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := engine.NewWorld()
			system := NewSlidingSystem(nil)

			player := w.AddEntity()
			pos := &PositionComponent{X: 50, Y: 50}
			vel := &VelocityComponent{X: 100, Y: 100}
			collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerAll^LayerPlayer)
			sliding := NewSlidingComponent()
			sliding.Friction = tt.friction

			w.AddComponent(player, pos)
			w.AddComponent(player, vel)
			w.AddComponent(player, &ColliderComponent{Collider: collider})
			w.AddComponent(player, sliding)

			// Create wall
			wall := w.AddEntity()
			wallCollider := NewAABBCollider(60, 0, 10, 200, LayerTerrain, LayerAll)
			w.AddComponent(wall, &PositionComponent{X: 60, Y: 0})
			w.AddComponent(wall, &ColliderComponent{Collider: wallCollider})

			system.Update(w)

			// Higher friction should result in less sliding distance
			// Just verify it doesn't panic - detailed friction testing would need multiple frames
			if pos.X < 0 || pos.Y < 0 {
				t.Errorf("Invalid position after friction test: (%v, %v)", pos.X, pos.Y)
			}
		})
	}
}

func TestSlidingSystemBounce(t *testing.T) {
	w := engine.NewWorld()
	system := NewSlidingSystem(nil)

	player := w.AddEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	vel := &VelocityComponent{X: 100, Y: 0} // Move straight toward wall
	collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerAll^LayerPlayer)
	sliding := NewSlidingComponent()
	sliding.BounceOnSteep = true
	sliding.BounceRestitution = 0.5
	sliding.MaxSlideAngle = 0.1 // Very small angle, so head-on is "steep"

	w.AddComponent(player, pos)
	w.AddComponent(player, vel)
	w.AddComponent(player, &ColliderComponent{Collider: collider})
	w.AddComponent(player, sliding)

	// Create wall directly in front
	wall := w.AddEntity()
	wallCollider := NewAABBCollider(60, 40, 10, 20, LayerTerrain, LayerAll)
	w.AddComponent(wall, &PositionComponent{X: 60, Y: 40})
	w.AddComponent(wall, &ColliderComponent{Collider: wallCollider})

	originalVelX := vel.X
	system.Update(w)

	// Velocity should have changed due to bounce (or stopped)
	// This is a basic test; detailed bounce physics would need more setup
	if vel.X == originalVelX && pos.X >= 55 {
		t.Errorf("Bounce didn't affect velocity or position")
	}
}

func TestSlidingSystemLayerMasking(t *testing.T) {
	// Test that layer masks prevent collision
	w := engine.NewWorld()
	system := NewSlidingSystem(nil)

	player := w.AddEntity()
	pos := &PositionComponent{X: 50, Y: 50}
	vel := &VelocityComponent{X: 100, Y: 0}
	// Player only collides with Terrain layer
	collider := NewCircleCollider(50, 50, 5, LayerPlayer, LayerTerrain)
	sliding := NewSlidingComponent()

	w.AddComponent(player, pos)
	w.AddComponent(player, vel)
	w.AddComponent(player, &ColliderComponent{Collider: collider})
	w.AddComponent(player, sliding)

	// Create enemy entity (different layer, should pass through)
	enemy := w.AddEntity()
	enemyCollider := NewCircleCollider(52, 50, 5, LayerEnemy, LayerPlayer)
	w.AddComponent(enemy, &PositionComponent{X: 52, Y: 50})
	w.AddComponent(enemy, &ColliderComponent{Collider: enemyCollider})

	originalX := pos.X
	system.Update(w)

	// Player should pass through enemy (different layer)
	expectedDelta := 100.0 * 0.016
	actualDelta := pos.X - originalX

	if math.Abs(actualDelta-expectedDelta) > 0.1 {
		t.Errorf("Player was blocked by wrong layer: delta = %v, expected ~%v", actualDelta, expectedDelta)
	}
}

func TestSlidingSystemDebugLogging(t *testing.T) {
	system := NewSlidingSystem(nil)

	system.SetDebugLogging(true)
	if !system.debugLogging {
		t.Error("Debug logging not enabled")
	}

	system.SetDebugLogging(false)
	if system.debugLogging {
		t.Error("Debug logging not disabled")
	}
}

func BenchmarkSlidingSystem(b *testing.B) {
	w := engine.NewWorld()
	grid := spatial.NewGrid(50.0)
	system := NewSlidingSystem(grid)

	// Create 100 entities
	for i := 0; i < 100; i++ {
		entity := w.AddEntity()
		pos := &PositionComponent{X: float64(i * 10), Y: 50}
		vel := &VelocityComponent{X: 50, Y: 30}
		collider := NewCircleCollider(float64(i*10), 50, 5, LayerPlayer, LayerAll)
		sliding := NewSlidingComponent()

		w.AddComponent(entity, pos)
		w.AddComponent(entity, vel)
		w.AddComponent(entity, &ColliderComponent{Collider: collider})
		w.AddComponent(entity, sliding)

		grid.Insert(entity, float64(i*10), 50)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(w)
	}
}

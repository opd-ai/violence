// Package collision provides terrain sliding for smooth wall collision response.
//
// The sliding system replaces hard-stop collision detection with smooth
// sliding movement along surfaces. This prevents entities from getting stuck
// at walls and creates more fluid, player-friendly movement.
//
// # Architecture
//
// The sliding system follows the ECS pattern:
//   - VelocityComponent: stores entity velocity (X, Y in units/sec)
//   - PositionComponent: stores entity position (X, Y in world coordinates)
//   - ColliderComponent: wraps collision.Collider for shape and layer info
//   - SlidingComponent: configuration for sliding behavior
//   - SlidingSystem: processes entities and applies sliding movement
//
// # How It Works
//
// When an entity with velocity would collide with terrain:
//  1. Compute the collision normal (direction of surface)
//  2. Project velocity onto the surface tangent (perpendicular to normal)
//  3. Apply friction to the slide velocity
//  4. Move entity along the surface instead of stopping
//  5. Repeat up to maxIterations times to handle corners
//
// This creates smooth movement along walls instead of abrupt stops.
//
// # Usage
//
// Add the required components to entities that should slide:
//
//	entity := world.AddEntity()
//	world.AddComponent(entity, &collision.PositionComponent{X: 100, Y: 100})
//	world.AddComponent(entity, &collision.VelocityComponent{X: 50, Y: 0})
//	collider := collision.NewCircleCollider(100, 100, 10, collision.LayerPlayer, collision.LayerAll)
//	world.AddComponent(entity, &collision.ColliderComponent{Collider: collider})
//	world.AddComponent(entity, collision.NewSlidingComponent())
//
// The system will automatically process these entities each frame.
//
// # Configuration
//
// SlidingComponent supports several tuning parameters:
//
//   - Enabled: toggle sliding on/off per-entity
//   - MaxSlideAngle: surfaces steeper than this cause bounce/stop (radians)
//   - Friction: coefficient applied to slide velocity (0.0 = frictionless, 1.0 = full stop)
//   - BounceOnSteep: if true, bounce off steep surfaces instead of stopping
//   - BounceRestitution: elasticity of bounce (0.0 = no bounce, 1.0 = perfect elastic)
//
// Example custom configuration:
//
//	sliding := &collision.SlidingComponent{
//	    Enabled:           true,
//	    MaxSlideAngle:     math.Pi / 3,  // 60 degrees
//	    Friction:          0.2,           // 20% friction
//	    BounceOnSteep:     true,
//	    BounceRestitution: 0.5,           // 50% bounce
//	}
//
// # Performance
//
// The sliding system uses spatial indexing (via spatial.Grid) for efficient
// collision queries. Without spatial indexing, it falls back to linear search
// (O(N) per entity). With spatial indexing, collision queries are O(k) where
// k is the number of nearby entities.
//
// The system limits iterations per frame (default: 4) to prevent runaway
// computation on complex geometry. Most scenarios resolve in 1-2 iterations.
//
// # Integration
//
// The system is registered in main.go and updates automatically each frame:
//
//	g.slidingSystem = collision.NewSlidingSystem(nil)
//	g.slidingSystem.SetSpatialIndex(g.spatialSystem.GetGrid())
//	g.world.AddSystem(g.slidingSystem)
//
// The spatial index should be set after the spatial system is initialized
// to enable efficient collision queries.
//
// # Collision Layers
//
// The sliding system respects collision layers defined in collision.Layer.
// Entities only slide against surfaces they're configured to collide with:
//
//	// Player slides against terrain and enemies
//	playerCollider := NewCircleCollider(x, y, r, LayerPlayer, LayerTerrain|LayerEnemy)
//
//	// Projectiles ignore allies
//	projCollider := NewCircleCollider(x, y, r, LayerProjectile, LayerAll^LayerPlayer)
//
// # Examples
//
// Player sliding along a wall while moving diagonally:
//
//	// Player at (50, 50) moving northeast (100, 100) units/sec
//	// Wall at x=60 (vertical AABB)
//	// Result: Player slides upward along wall, X position clamped at ~55 (radius 5)
//
// Entity bouncing off a steep surface:
//
//	sliding := &SlidingComponent{
//	    MaxSlideAngle:     0.1,  // Very small angle
//	    BounceOnSteep:     true,
//	    BounceRestitution: 0.5,
//	}
//	// Entity moving straight at wall bounces back with 50% velocity
//
// Corner sliding with multiple iterations:
//
//	// Entity moving into L-shaped corner
//	// Iteration 1: slides along first wall
//	// Iteration 2: slides along second wall
//	// Result: smooth navigation around corner
package collision

// Package collision - Usage Examples
//
// The collision package provides a layer-based collision detection system with
// support for multiple shape types (circle, capsule, AABB, polygon).
//
// BASIC USAGE:
//
//	// Create colliders
//	player := collision.NewCircleCollider(0, 0, 0.5, collision.LayerPlayer, collision.LayerEnemy|collision.LayerTerrain)
//	enemy := collision.NewCircleCollider(10, 0, 0.5, collision.LayerEnemy, collision.LayerPlayer)
//
//	// Test collision
//	if collision.TestCollision(player, enemy) {
//	    // Handle collision
//	}
//
// LAYER MASKING:
//
// Layers define what category an entity belongs to:
//   - LayerPlayer: Player character
//   - LayerEnemy: Enemy NPCs
//   - LayerProjectile: Bullets, spells
//   - LayerTerrain: Walls, obstacles
//   - LayerEnvironment: Props, decorations
//   - LayerEthereal: Ghosts (pass through most layers)
//   - LayerInteractive: Doors, chests
//   - LayerTrigger: Trigger zones
//
// Masks define what layers an entity can interact with:
//
//	// Player collides with enemies and terrain
//	player := collision.NewCircleCollider(x, y, radius,
//	    collision.LayerPlayer,
//	    collision.LayerEnemy|collision.LayerTerrain)
//
//	// Projectile only hits enemies
//	projectile := collision.NewCircleCollider(x, y, radius,
//	    collision.LayerProjectile,
//	    collision.LayerEnemy)
//
// ATTACK SHAPES:
//
// Create precise collision shapes for attacks matching telegraph patterns:
//
//	// Cone attack (melee swipe)
//	cone := collision.CreateConeCollider(x, y, dirX, dirY, range_, angle,
//	    collision.LayerPlayer, collision.LayerEnemy)
//
//	// Circle AoE (explosion)
//	aoe := collision.CreateCircleAttackCollider(x, y, radius,
//	    collision.LayerPlayer, collision.LayerEnemy)
//
//	// Line attack (beam, charge)
//	beam := collision.CreateLineAttackCollider(x, y, dirX, dirY, range_, width,
//	    collision.LayerPlayer, collision.LayerEnemy)
//
//	// Ring attack (shockwave)
//	outer, inner := collision.CreateRingCollider(x, y, outerRadius, innerRadius,
//	    collision.LayerPlayer, collision.LayerEnemy)
//
// ECS INTEGRATION:
//
//	// Add collider to entity
//	w := engine.NewWorld()
//	e := w.AddEntity()
//	collider := collision.NewCircleCollider(x, y, radius, layer, mask)
//	collision.AddColliderToEntity(w, e, collider)
//
//	// Query collisions
//	hits := collision.QueryCollisions(w, playerCollider)
//	for _, entityID := range hits {
//	    // Process collision with entityID
//	}
//
//	// Radius query
//	nearby := collision.QueryCollisionsInRadius(w, x, y, radius, collision.LayerEnemy)
//
// MOVEMENT WITH SLIDING:
//
//	// Get collision normal
//	nx, ny := collision.GetCollisionNormal(player, wall)
//
//	// Compute slide vector
//	slideX, slideY := collision.SlideVector(velocityX, velocityY, nx, ny)
//
//	// Apply sliding movement instead of stopping
//	collision.UpdateColliderPosition(player, x+slideX, y+slideY)
//
// SHAPE TYPES:
//
//   - Circle: Best for characters, radial effects
//   - Capsule: Best for projectiles, beams, melee weapons
//   - AABB: Best for terrain tiles, rectangular objects
//   - Polygon: Best for irregular shapes, cone attacks
//
// PERFORMANCE:
//
// The collision system is optimized for real-time use:
//   - Layer masking prevents unnecessary tests
//   - Bounding circle broadphase for complex shapes
//   - Efficient geometric algorithms
//   - Zero allocations in hot paths (reuse colliders)
//
// Target: 60+ FPS with hundreds of active colliders.
package collision

// Package collision provides attack shape collider generation.
package collision

import (
	"math"
)

// CreateConeCollider creates a polygon collider matching a cone attack pattern.
// Used for melee swipes, flamethrowers, etc.
func CreateConeCollider(x, y, dirX, dirY, range_, angle float64, layer, mask Layer) *Collider {
	segments := 8
	vertices := make([]Point, segments+1)

	// Calculate angle from direction
	baseAngle := math.Atan2(dirY, dirX)
	halfAngle := angle / 2

	// First vertex at origin
	vertices[0] = Point{X: 0, Y: 0}

	// Arc vertices
	for i := 0; i < segments; i++ {
		t := -halfAngle + (float64(i)/float64(segments-1))*angle
		vx := math.Cos(baseAngle+t) * range_
		vy := math.Sin(baseAngle+t) * range_
		vertices[i+1] = Point{X: vx, Y: vy}
	}

	return NewPolygonCollider(x, y, vertices, layer, mask)
}

// CreateCircleAttackCollider creates a circular AoE collider.
// Used for slams, explosions, radial attacks.
func CreateCircleAttackCollider(x, y, range_ float64, layer, mask Layer) *Collider {
	return NewCircleCollider(x, y, range_, layer, mask)
}

// CreateLineAttackCollider creates a capsule collider for beam/charge attacks.
func CreateLineAttackCollider(x, y, dirX, dirY, range_, width float64, layer, mask Layer) *Collider {
	endX := x + dirX*range_
	endY := y + dirY*range_
	return NewCapsuleCollider(x, y, endX, endY, width/2, layer, mask)
}

// CreateRingCollider creates a ring (donut) shaped collider using two circles.
// Returns inner and outer colliders that must both be checked.
func CreateRingCollider(x, y, outerRange, innerRange float64, layer, mask Layer) (*Collider, *Collider) {
	outer := NewCircleCollider(x, y, outerRange, layer, mask)
	// Inner circle with inverted logic - entities must NOT be inside this
	inner := NewCircleCollider(x, y, innerRange, layer, LayerNone)
	return outer, inner
}

// TestRingCollision checks if a point/collider is in the ring (between inner and outer).
func TestRingCollision(target, outer, inner *Collider) bool {
	// Must be inside outer circle but NOT inside inner circle
	inOuter := TestCollision(target, outer)
	if !inOuter {
		return false
	}

	// Temporarily enable inner for test
	wasEnabled := inner.Enabled
	inner.Enabled = true
	inner.Mask = outer.Mask
	inInner := TestCollision(target, inner)
	inner.Enabled = wasEnabled
	inner.Mask = LayerNone

	return !inInner
}

// CreateProjectileCollider creates a capsule for moving projectiles.
// Swept collision from last position to current position.
func CreateProjectileCollider(lastX, lastY, currentX, currentY, radius float64, layer, mask Layer) *Collider {
	return NewCapsuleCollider(lastX, lastY, currentX, currentY, radius, layer, mask)
}

// CreateMeleeWeaponCollider creates a collider for a melee weapon swing.
// Approximates weapon as a rotating capsule.
func CreateMeleeWeaponCollider(playerX, playerY, dirX, dirY, weaponLength, weaponWidth float64, layer, mask Layer) *Collider {
	// Weapon extends from player position in attack direction
	startX := playerX + dirX*5 // Slight offset from player center
	startY := playerY + dirY*5
	endX := startX + dirX*weaponLength
	endY := startY + dirY*weaponLength
	return NewCapsuleCollider(startX, startY, endX, endY, weaponWidth/2, layer, mask)
}

// CreateCharacterCollider creates a standard circular collider for a character.
func CreateCharacterCollider(x, y, radius float64, isPlayer bool) *Collider {
	if isPlayer {
		return NewCircleCollider(x, y, radius, LayerPlayer, LayerEnemy|LayerTerrain|LayerEnvironment|LayerInteractive)
	}
	return NewCircleCollider(x, y, radius, LayerEnemy, LayerPlayer|LayerTerrain|LayerEnvironment)
}

// CreateTerrainCollider creates an AABB collider for terrain tiles.
func CreateTerrainCollider(x, y, w, h float64) *Collider {
	return NewAABBCollider(x, y, w, h, LayerTerrain, LayerAll)
}

// CreatePropCollider creates a collider for environment props.
func CreatePropCollider(x, y, radius float64, blocksMovement bool) *Collider {
	if blocksMovement {
		return NewCircleCollider(x, y, radius, LayerEnvironment, LayerPlayer|LayerEnemy)
	}
	// Non-blocking props don't collide with movement
	return NewCircleCollider(x, y, radius, LayerEnvironment, LayerProjectile)
}

// CreateTriggerZone creates a trigger collider that detects but doesn't block.
func CreateTriggerZone(x, y, radius float64, detectLayers Layer) *Collider {
	return NewCircleCollider(x, y, radius, LayerTrigger, detectLayers)
}

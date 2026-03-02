// Package collision provides terrain sliding for smooth movement along walls.
package collision

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/spatial"
	"github.com/sirupsen/logrus"
)

// VelocityComponent stores entity movement velocity.
type VelocityComponent struct {
	X, Y float64 // Current velocity in world units per second
}

// Type returns the component type identifier.
func (v *VelocityComponent) Type() string {
	return "VelocityComponent"
}

// PositionComponent stores entity position.
type PositionComponent struct {
	X, Y float64
}

// Type returns the component type identifier.
func (p *PositionComponent) Type() string {
	return "PositionComponent"
}

// SlidingComponent marks entities that should slide along terrain on collision.
type SlidingComponent struct {
	Enabled           bool    // Whether sliding is active
	MaxSlideAngle     float64 // Maximum angle (radians) to allow sliding (default: 90 degrees)
	Friction          float64 // Friction coefficient when sliding (0.0-1.0, default 0.1)
	BounceOnSteep     bool    // Whether to bounce off steep surfaces instead of stopping
	BounceRestitution float64 // Bounce coefficient (0.0-1.0, default 0.3)
}

// Type returns the component type identifier.
func (s *SlidingComponent) Type() string {
	return "SlidingComponent"
}

// NewSlidingComponent creates a sliding component with sensible defaults.
func NewSlidingComponent() *SlidingComponent {
	return &SlidingComponent{
		Enabled:           true,
		MaxSlideAngle:     math.Pi / 2, // 90 degrees
		Friction:          0.1,
		BounceOnSteep:     false,
		BounceRestitution: 0.3,
	}
}

// SlidingSystem applies smooth terrain sliding to entities with colliders.
type SlidingSystem struct {
	spatialIndex  *spatial.Grid
	maxIterations int     // Maximum slide iterations per frame
	minVelocity   float64 // Minimum velocity to consider for sliding
	debugLogging  bool
}

// NewSlidingSystem creates a sliding system.
func NewSlidingSystem(spatialIndex *spatial.Grid) *SlidingSystem {
	return &SlidingSystem{
		spatialIndex:  spatialIndex,
		maxIterations: 4,     // Up to 4 slide iterations per frame
		minVelocity:   0.001, // Ignore velocities below 0.001 units/sec
		debugLogging:  false,
	}
}

// SetSpatialIndex sets or updates the spatial index used for collision queries.
func (s *SlidingSystem) SetSpatialIndex(spatialIndex *spatial.Grid) {
	s.spatialIndex = spatialIndex
}

// Update applies sliding movement to all entities with the required components.
func (s *SlidingSystem) Update(w *engine.World) {
	deltaTime := 0.016 // Assume 60 FPS

	// Query entities with position, velocity, collider, and sliding components
	posType := reflect.TypeOf(&PositionComponent{})
	velType := reflect.TypeOf(&VelocityComponent{})
	colliderType := reflect.TypeOf(&ColliderComponent{})
	slidingType := reflect.TypeOf(&SlidingComponent{})

	entities := w.Query(posType, velType, colliderType, slidingType)

	for _, entity := range entities {
		posComp, _ := w.GetComponent(entity, posType)
		velComp, _ := w.GetComponent(entity, velType)
		colliderComp, _ := w.GetComponent(entity, colliderType)
		slidingComp, _ := w.GetComponent(entity, slidingType)

		pos := posComp.(*PositionComponent)
		vel := velComp.(*VelocityComponent)
		collider := colliderComp.(*ColliderComponent).Collider
		sliding := slidingComp.(*SlidingComponent)

		if !sliding.Enabled {
			continue
		}

		// Skip if velocity is negligible
		velMag := math.Sqrt(vel.X*vel.X + vel.Y*vel.Y)
		if velMag < s.minVelocity {
			continue
		}

		// Apply sliding movement with multiple iterations
		s.applySlidingMovement(w, entity, pos, vel, collider, sliding, deltaTime)
	}
}

// applySlidingMovement moves an entity with terrain sliding.
func (s *SlidingSystem) applySlidingMovement(
	w *engine.World,
	entity engine.Entity,
	pos *PositionComponent,
	vel *VelocityComponent,
	collider *Collider,
	sliding *SlidingComponent,
	deltaTime float64,
) {
	// Calculate desired movement
	moveX := vel.X * deltaTime
	moveY := vel.Y * deltaTime

	// Multi-iteration sliding to handle corners and complex geometry
	remainingX := moveX
	remainingY := moveY

	for iteration := 0; iteration < s.maxIterations; iteration++ {
		// Check if we still have movement to apply
		if math.Abs(remainingX) < 0.001 && math.Abs(remainingY) < 0.001 {
			break
		}

		// Try to move to target position
		targetX := pos.X + remainingX
		targetY := pos.Y + remainingY

		// Update collider position for test
		originalX, originalY := collider.X, collider.Y
		collider.X = targetX
		collider.Y = targetY

		// Find collisions
		collision := s.findFirstCollision(w, entity, collider)

		if collision == nil {
			// No collision, move freely
			pos.X = targetX
			pos.Y = targetY
			break
		}

		// Restore original collider position
		collider.X = originalX
		collider.Y = originalY

		// Get collision normal
		nx, ny := GetCollisionNormal(collider, collision)
		if nx == 0 && ny == 0 {
			// Unable to compute normal, stop movement
			if s.debugLogging {
				logrus.WithFields(logrus.Fields{
					"system": "SlidingSystem",
					"entity": entity,
				}).Debug("Unable to compute collision normal, stopping")
			}
			break
		}

		// Check if surface is too steep to slide
		velocityAngle := math.Atan2(remainingY, remainingX)
		normalAngle := math.Atan2(ny, nx)
		angleDiff := math.Abs(normalAngle - velocityAngle)
		if angleDiff > math.Pi {
			angleDiff = 2*math.Pi - angleDiff
		}

		if angleDiff > sliding.MaxSlideAngle {
			if sliding.BounceOnSteep {
				// Bounce off steep surface
				dot := remainingX*nx + remainingY*ny
				bounceX := remainingX - 2*dot*nx
				bounceY := remainingY - 2*dot*ny
				remainingX = bounceX * sliding.BounceRestitution
				remainingY = bounceY * sliding.BounceRestitution

				// Update velocity for bounce
				vel.X = (remainingX / deltaTime) * sliding.BounceRestitution
				vel.Y = (remainingY / deltaTime) * sliding.BounceRestitution
			}
			break
		}

		// Compute slide vector
		slideX, slideY := SlideVector(remainingX, remainingY, nx, ny)

		// Apply friction to slide
		frictionFactor := 1.0 - sliding.Friction
		slideX *= frictionFactor
		slideY *= frictionFactor

		// Move along the surface
		slideDistance := math.Sqrt(slideX*slideX + slideY*slideY)
		if slideDistance < 0.001 {
			// Slide distance too small, stop
			break
		}

		// Try sliding movement
		collider.X = pos.X + slideX
		collider.Y = pos.Y + slideY

		// Check if sliding movement also collides
		slideCollision := s.findFirstCollision(w, entity, collider)
		collider.X = originalX
		collider.Y = originalY

		if slideCollision == nil {
			// Sliding worked, apply it
			pos.X += slideX
			pos.Y += slideY

			// Update remaining movement (project onto slide direction)
			remainingDist := math.Sqrt(remainingX*remainingX + remainingY*remainingY)
			if remainingDist > 0 {
				slideNormX := slideX / slideDistance
				slideNormY := slideY / slideDistance
				projectedDist := remainingX*slideNormX + remainingY*slideNormY

				// Remaining movement is what we haven't slid yet
				remainingX = slideNormX * math.Max(0, projectedDist-slideDistance)
				remainingY = slideNormY * math.Max(0, projectedDist-slideDistance)
			} else {
				remainingX = 0
				remainingY = 0
			}
		} else {
			// Even sliding hits something, stop here
			break
		}
	}

	// Update collider to final position
	collider.X = pos.X
	collider.Y = pos.Y
}

// findFirstCollision finds the first colliding entity using spatial indexing.
func (s *SlidingSystem) findFirstCollision(w *engine.World, entity engine.Entity, collider *Collider) *Collider {
	if s.spatialIndex == nil {
		// Fallback to linear search if no spatial index
		return s.findFirstCollisionLinear(w, entity, collider)
	}

	// Get nearby entities from spatial grid
	bx, by, br := GetBoundingCircle(collider)
	nearby := s.spatialIndex.QueryRadius(bx, by, br*1.5) // 1.5x radius for safety margin

	colliderType := reflect.TypeOf(&ColliderComponent{})

	for _, nearbyEntity := range nearby {
		if nearbyEntity == entity {
			continue // Skip self
		}

		otherColliderComp, ok := w.GetComponent(nearbyEntity, colliderType)
		if !ok {
			continue
		}

		otherCollider := otherColliderComp.(*ColliderComponent).Collider
		if otherCollider == nil || !otherCollider.Enabled {
			continue
		}

		// Test collision
		if TestCollision(collider, otherCollider) {
			return otherCollider
		}
	}

	return nil
}

// findFirstCollisionLinear finds collisions by checking all entities (fallback).
func (s *SlidingSystem) findFirstCollisionLinear(w *engine.World, entity engine.Entity, collider *Collider) *Collider {
	colliderType := reflect.TypeOf(&ColliderComponent{})
	allEntities := w.Query(colliderType)

	for _, otherEntity := range allEntities {
		if otherEntity == entity {
			continue
		}

		otherColliderComp, _ := w.GetComponent(otherEntity, colliderType)
		otherCollider := otherColliderComp.(*ColliderComponent).Collider
		if otherCollider == nil || !otherCollider.Enabled {
			continue
		}

		if TestCollision(collider, otherCollider) {
			return otherCollider
		}
	}

	return nil
}

// SetDebugLogging enables or disables debug logging.
func (s *SlidingSystem) SetDebugLogging(enabled bool) {
	s.debugLogging = enabled
}

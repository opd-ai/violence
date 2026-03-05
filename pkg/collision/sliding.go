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
	moveX := vel.X * deltaTime
	moveY := vel.Y * deltaTime
	remainingX, remainingY := moveX, moveY

	for iteration := 0; iteration < s.maxIterations; iteration++ {
		if s.shouldStopMovement(remainingX, remainingY) {
			break
		}

		targetX, targetY := pos.X+remainingX, pos.Y+remainingY
		originalX, originalY := collider.X, collider.Y

		if s.tryDirectMovement(w, entity, pos, collider, targetX, targetY, originalX, originalY) {
			break
		}

		nx, ny := GetCollisionNormal(collider, s.findCollisionAt(w, entity, collider, targetX, targetY, originalX, originalY))
		if s.shouldStopOnInvalidNormal(entity, nx, ny) {
			break
		}

		if s.handleSteepSurface(sliding, &remainingX, &remainingY, vel, nx, ny, deltaTime) {
			break
		}

		slideX, slideY := s.computeFrictionalSlide(remainingX, remainingY, nx, ny, sliding)
		if s.shouldStopOnSmallSlide(slideX, slideY) {
			break
		}

		if !s.applySlidingIfPossible(w, entity, pos, collider, slideX, slideY, originalX, originalY, &remainingX, &remainingY) {
			break
		}
	}

	collider.X = pos.X
	collider.Y = pos.Y
}

// shouldStopMovement checks if remaining movement is negligible.
func (s *SlidingSystem) shouldStopMovement(remainingX, remainingY float64) bool {
	return math.Abs(remainingX) < 0.001 && math.Abs(remainingY) < 0.001
}

// tryDirectMovement attempts to move directly to target if no collision occurs.
func (s *SlidingSystem) tryDirectMovement(w *engine.World, entity engine.Entity, pos *PositionComponent, collider *Collider, targetX, targetY, originalX, originalY float64) bool {
	collider.X = targetX
	collider.Y = targetY

	if s.findFirstCollision(w, entity, collider) == nil {
		pos.X = targetX
		pos.Y = targetY
		return true
	}

	collider.X = originalX
	collider.Y = originalY
	return false
}

// findCollisionAt tests collision at a specific position and returns the collider.
func (s *SlidingSystem) findCollisionAt(w *engine.World, entity engine.Entity, collider *Collider, targetX, targetY, originalX, originalY float64) *Collider {
	collider.X = targetX
	collider.Y = targetY
	collision := s.findFirstCollision(w, entity, collider)
	collider.X = originalX
	collider.Y = originalY
	return collision
}

// shouldStopOnInvalidNormal checks if collision normal is invalid and logs if debugging.
func (s *SlidingSystem) shouldStopOnInvalidNormal(entity engine.Entity, nx, ny float64) bool {
	if nx == 0 && ny == 0 {
		if s.debugLogging {
			logrus.WithFields(logrus.Fields{
				"system": "SlidingSystem",
				"entity": entity,
			}).Debug("Unable to compute collision normal, stopping")
		}
		return true
	}
	return false
}

// handleSteepSurface checks surface angle and applies bounce or stops movement.
func (s *SlidingSystem) handleSteepSurface(sliding *SlidingComponent, remainingX, remainingY *float64, vel *VelocityComponent, nx, ny, deltaTime float64) bool {
	angleDiff := calculateNormalizedAngleDiff(*remainingX, *remainingY, nx, ny)

	if angleDiff > sliding.MaxSlideAngle {
		if sliding.BounceOnSteep {
			applyBounce(remainingX, remainingY, vel, nx, ny, sliding.BounceRestitution, deltaTime)
		}
		return true
	}
	return false
}

// calculateNormalizedAngleDiff computes the normalized angle difference between velocity and normal.
func calculateNormalizedAngleDiff(remainingX, remainingY, nx, ny float64) float64 {
	velocityAngle := math.Atan2(remainingY, remainingX)
	normalAngle := math.Atan2(ny, nx)
	angleDiff := math.Abs(normalAngle - velocityAngle)
	if angleDiff > math.Pi {
		angleDiff = 2*math.Pi - angleDiff
	}
	return angleDiff
}

// applyBounce computes bounce vector and updates remaining movement and velocity.
func applyBounce(remainingX, remainingY *float64, vel *VelocityComponent, nx, ny, restitution, deltaTime float64) {
	dot := (*remainingX)*nx + (*remainingY)*ny
	bounceX := *remainingX - 2*dot*nx
	bounceY := *remainingY - 2*dot*ny
	*remainingX = bounceX * restitution
	*remainingY = bounceY * restitution
	vel.X = (*remainingX / deltaTime) * restitution
	vel.Y = (*remainingY / deltaTime) * restitution
}

// computeFrictionalSlide calculates slide vector with friction applied.
func (s *SlidingSystem) computeFrictionalSlide(remainingX, remainingY, nx, ny float64, sliding *SlidingComponent) (float64, float64) {
	slideX, slideY := SlideVector(remainingX, remainingY, nx, ny)
	frictionFactor := 1.0 - sliding.Friction
	return slideX * frictionFactor, slideY * frictionFactor
}

// shouldStopOnSmallSlide checks if slide distance is too small to continue.
func (s *SlidingSystem) shouldStopOnSmallSlide(slideX, slideY float64) bool {
	slideDistance := math.Sqrt(slideX*slideX + slideY*slideY)
	return slideDistance < 0.001
}

// applySlidingIfPossible attempts sliding movement and updates remaining movement if successful.
func (s *SlidingSystem) applySlidingIfPossible(w *engine.World, entity engine.Entity, pos *PositionComponent, collider *Collider, slideX, slideY, originalX, originalY float64, remainingX, remainingY *float64) bool {
	collider.X = pos.X + slideX
	collider.Y = pos.Y + slideY

	slideCollision := s.findFirstCollision(w, entity, collider)
	collider.X = originalX
	collider.Y = originalY

	if slideCollision == nil {
		pos.X += slideX
		pos.Y += slideY
		updateRemainingMovement(remainingX, remainingY, slideX, slideY)
		return true
	}
	return false
}

// updateRemainingMovement projects remaining movement onto the slide direction.
func updateRemainingMovement(remainingX, remainingY *float64, slideX, slideY float64) {
	slideDistance := math.Sqrt(slideX*slideX + slideY*slideY)
	remainingDist := math.Sqrt((*remainingX)*(*remainingX) + (*remainingY)*(*remainingY))

	if remainingDist > 0 {
		slideNormX := slideX / slideDistance
		slideNormY := slideY / slideDistance
		projectedDist := (*remainingX)*slideNormX + (*remainingY)*slideNormY
		*remainingX = slideNormX * math.Max(0, projectedDist-slideDistance)
		*remainingY = slideNormY * math.Max(0, projectedDist-slideDistance)
	} else {
		*remainingX = 0
		*remainingY = 0
	}
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

// Package collision provides ECS collision system integration.
package collision

import (
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/spatial"
	"github.com/sirupsen/logrus"
)

// ColliderComponent is an ECS component wrapping collision data.
type ColliderComponent struct {
	Collider *Collider
}

// System handles collision detection and resolution.
type System struct {
	logger *logrus.Entry
}

// NewSystem creates a new collision system.
func NewSystem() *System {
	return &System{
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "collision",
		}),
	}
}

// Update processes all entities with colliders.
func (s *System) Update(w *engine.World) {
	// This system is passive - collision queries are done on-demand
	// by other systems (combat, movement, etc.)
	// We could add debug visualization here if needed
}

// QueryCollisions returns all colliders that intersect with the given collider.
func QueryCollisions(w *engine.World, collider *Collider) []engine.Entity {
	var results []engine.Entity

	colliderType := reflect.TypeOf(&ColliderComponent{})
	entities := w.Query(colliderType)

	for _, e := range entities {
		comp, ok := w.GetComponent(e, colliderType)
		if !ok {
			continue
		}

		cc, ok := comp.(*ColliderComponent)
		if !ok || cc.Collider == nil {
			continue
		}

		if TestCollision(collider, cc.Collider) {
			results = append(results, e)
		}
	}

	return results
}

// QueryCollisionsInRadius returns all colliders within a circular area.
// Uses bounding circle broadphase for efficiency.
func QueryCollisionsInRadius(w *engine.World, x, y, radius float64, layer Layer) []engine.Entity {
	var results []engine.Entity

	// Create query circle
	queryCollider := NewCircleCollider(x, y, radius, layer, LayerAll)

	colliderType := reflect.TypeOf(&ColliderComponent{})
	entities := w.Query(colliderType)

	for _, e := range entities {
		comp, ok := w.GetComponent(e, colliderType)
		if !ok {
			continue
		}

		cc, ok := comp.(*ColliderComponent)
		if !ok || cc.Collider == nil {
			continue
		}

		if TestCollision(queryCollider, cc.Collider) {
			results = append(results, e)
		}
	}

	return results
}

// UpdateColliderPosition updates a collider's position.
func UpdateColliderPosition(collider *Collider, x, y float64) {
	collider.X = x
	collider.Y = y
}

// UpdateCapsuleCollider updates a capsule collider's endpoints.
func UpdateCapsuleCollider(collider *Collider, x1, y1, x2, y2 float64) {
	if collider.Shape != ShapeCapsule {
		return
	}
	collider.X = x1
	collider.Y = y1
	collider.X2 = x2
	collider.Y2 = y2
}

// GetEntityCollider retrieves the collider component from an entity.
func GetEntityCollider(w *engine.World, e engine.Entity) *Collider {
	comp, ok := w.GetComponent(e, reflect.TypeOf(&ColliderComponent{}))
	if !ok {
		return nil
	}

	cc, ok := comp.(*ColliderComponent)
	if !ok {
		return nil
	}

	return cc.Collider
}

// AddColliderToEntity adds a collider component to an entity.
func AddColliderToEntity(w *engine.World, e engine.Entity, collider *Collider) {
	w.AddComponent(e, &ColliderComponent{Collider: collider})
}

// QueryCollisionsInRadiusSpatial returns all colliders within a circular area.
// Uses spatial indexing for O(1) broadphase when a spatial system is available.
// Falls back to linear search if no spatial system is registered.
func QueryCollisionsInRadiusSpatial(w *engine.World, spatialSys *spatial.System, x, y, radius float64, layer Layer) []engine.Entity {
	var results []engine.Entity

	// Create query circle
	queryCollider := NewCircleCollider(x, y, radius, layer, LayerAll)
	colliderType := reflect.TypeOf(&ColliderComponent{})

	var candidates []engine.Entity

	// Use spatial index if available for fast broadphase
	if spatialSys != nil {
		candidates = spatialSys.QueryRadius(x, y, radius)
	} else {
		// Fallback to linear search
		candidates = w.Query(colliderType)
	}

	// Narrow phase: precise collision testing
	for _, e := range candidates {
		comp, ok := w.GetComponent(e, colliderType)
		if !ok {
			continue
		}

		cc, ok := comp.(*ColliderComponent)
		if !ok || cc.Collider == nil {
			continue
		}

		if TestCollision(queryCollider, cc.Collider) {
			results = append(results, e)
		}
	}

	return results
}

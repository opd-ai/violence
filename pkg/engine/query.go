// Package engine provides ECS query functionality with bitmask-based archetype matching.
package engine

// ComponentID represents a unique bit position for a component type.
// Each component type is assigned a unique bit (0-63) for bitmask operations.
type ComponentID uint8

// Common component IDs (0-63 supported)
const (
	ComponentIDPosition ComponentID = iota
	ComponentIDVelocity
	ComponentIDHealth
	ComponentIDArmor
	ComponentIDInventory
	ComponentIDCamera
	ComponentIDInput
	ComponentIDSprite
	ComponentIDCollider
	ComponentIDAI
	ComponentIDWeapon
	ComponentIDProjectile
	ComponentIDLight
	ComponentIDSound
	ComponentIDAnimation
	ComponentIDParticle
)

// EntityIterator provides iteration over query results.
type EntityIterator struct {
	entities []Entity
	index    int
}

// Next advances to the next entity and returns true if available.
func (it *EntityIterator) Next() bool {
	it.index++
	return it.index < len(it.entities)
}

// Entity returns the current entity.
func (it *EntityIterator) Entity() Entity {
	if it.index < 0 || it.index >= len(it.entities) {
		return 0
	}
	return it.entities[it.index]
}

// Reset resets the iterator to the beginning.
func (it *EntityIterator) Reset() {
	it.index = -1
}

// HasNext returns true if there are more entities to iterate.
func (it *EntityIterator) HasNext() bool {
	return it.index+1 < len(it.entities)
}

// newEntityIterator creates an iterator from a slice of entities.
func newEntityIterator(entities []Entity) *EntityIterator {
	return &EntityIterator{
		entities: entities,
		index:    -1,
	}
}

// QueryWithBitmask returns an iterator over entities matching the component bitmask.
// Uses bitmask archetype matching: each bit represents a component type.
// Example: Query(ComponentIDPosition, ComponentIDVelocity) returns entities with both components.
func (w *World) QueryWithBitmask(componentIDs ...ComponentID) *EntityIterator {
	// Build query mask from component IDs
	var queryMask uint64
	for _, id := range componentIDs {
		if id < 64 {
			queryMask |= (1 << uint64(id))
		}
	}

	// Find matching entities
	var matched []Entity
	for entity, archetype := range w.archetypes {
		// Entity matches if it has all required component bits
		if archetype&queryMask == queryMask {
			matched = append(matched, entity)
		}
	}

	return newEntityIterator(matched)
}

// SetArchetype sets the component bitmask for an entity.
// This should be called whenever components are added or removed.
func (w *World) SetArchetype(e Entity, componentIDs ...ComponentID) {
	if w.archetypes == nil {
		w.archetypes = make(map[Entity]uint64)
	}

	var mask uint64
	for _, id := range componentIDs {
		if id < 64 {
			mask |= (1 << uint64(id))
		}
	}
	w.archetypes[e] = mask
}

// GetArchetype returns the component bitmask for an entity.
func (w *World) GetArchetype(e Entity) uint64 {
	if w.archetypes == nil {
		return 0
	}
	return w.archetypes[e]
}

// AddArchetypeComponent adds a component bit to an entity's archetype.
func (w *World) AddArchetypeComponent(e Entity, id ComponentID) {
	if w.archetypes == nil {
		w.archetypes = make(map[Entity]uint64)
	}
	if id < 64 {
		w.archetypes[e] |= (1 << uint64(id))
	}
}

// RemoveArchetypeComponent removes a component bit from an entity's archetype.
func (w *World) RemoveArchetypeComponent(e Entity, id ComponentID) {
	if w.archetypes == nil {
		return
	}
	if id < 64 {
		w.archetypes[e] &^= (1 << uint64(id))
	}
}

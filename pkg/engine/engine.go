// Package engine provides a minimal ECS (Entity-Component-System) framework.
package engine

// Entity is a unique identifier for a game entity.
type Entity uint64

// Component is a marker interface for ECS components.
type Component interface{}

// System processes entities each frame.
type System interface {
	Update(w *World)
}

// World holds all entities, components, and systems.
type World struct {
	nextID     Entity
	components map[Entity][]Component
	systems    []System
}

// NewWorld creates an empty World.
func NewWorld() *World {
	return &World{
		components: make(map[Entity][]Component),
	}
}

// AddEntity creates a new entity and returns its ID.
func (w *World) AddEntity() Entity {
	id := w.nextID
	w.nextID++
	return id
}

// AddComponent attaches a component to an entity.
func (w *World) AddComponent(e Entity, c Component) {
	w.components[e] = append(w.components[e], c)
}

// AddSystem registers a system for processing.
func (w *World) AddSystem(s System) {
	w.systems = append(w.systems, s)
}

// Update runs all registered systems.
func (w *World) Update() {
	for _, s := range w.systems {
		s.Update(w)
	}
}

// SetGenre configures the engine for a genre.
func SetGenre(genreID string) {}

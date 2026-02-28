// Package engine provides a minimal ECS (Entity-Component-System) framework.
package engine

import (
	"reflect"
)

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
	components map[Entity]map[reflect.Type]Component
	systems    []System
	genre      string
}

// NewWorld creates an empty World.
func NewWorld() *World {
	return &World{
		components: make(map[Entity]map[reflect.Type]Component),
		genre:      "fantasy",
	}
}

// AddEntity creates a new entity and returns its ID.
func (w *World) AddEntity() Entity {
	id := w.nextID
	w.nextID++
	w.components[id] = make(map[reflect.Type]Component)
	return id
}

// AddComponent attaches a component to an entity.
func (w *World) AddComponent(e Entity, c Component) {
	if c == nil {
		return
	}
	if w.components[e] == nil {
		w.components[e] = make(map[reflect.Type]Component)
	}
	w.components[e][reflect.TypeOf(c)] = c
}

// GetComponent retrieves a component from an entity.
// Returns the component and true if found, nil and false otherwise.
func (w *World) GetComponent(e Entity, componentType reflect.Type) (Component, bool) {
	entityComps, exists := w.components[e]
	if !exists {
		return nil, false
	}
	comp, found := entityComps[componentType]
	return comp, found
}

// HasComponent checks if an entity has a component of the given type.
func (w *World) HasComponent(e Entity, componentType reflect.Type) bool {
	entityComps, exists := w.components[e]
	if !exists {
		return false
	}
	_, found := entityComps[componentType]
	return found
}

// RemoveComponent removes a component from an entity.
func (w *World) RemoveComponent(e Entity, componentType reflect.Type) {
	if entityComps, exists := w.components[e]; exists {
		delete(entityComps, componentType)
	}
}

// RemoveEntity removes an entity and all its components.
func (w *World) RemoveEntity(e Entity) {
	delete(w.components, e)
}

// Query returns all entities that have all the specified component types.
func (w *World) Query(componentTypes ...reflect.Type) []Entity {
	var result []Entity
	for entity, comps := range w.components {
		hasAll := true
		for _, ct := range componentTypes {
			if _, found := comps[ct]; !found {
				hasAll = false
				break
			}
		}
		if hasAll {
			result = append(result, entity)
		}
	}
	return result
}

// AddSystem registers a system for processing.
func (w *World) AddSystem(s System) {
	w.systems = append(w.systems, s)
}

// Update runs all registered systems in order.
func (w *World) Update() {
	for _, s := range w.systems {
		s.Update(w)
	}
}

// SetGenre configures the world for a specific genre.
func (w *World) SetGenre(genreID string) {
	w.genre = genreID
}

// GetGenre returns the current genre.
func (w *World) GetGenre() string {
	return w.genre
}

var currentGenre = "fantasy"

// SetGenre configures the engine for a genre (global).
func SetGenre(genreID string) {
	currentGenre = genreID
}

// GetCurrentGenre returns the current global genre setting.
func GetCurrentGenre() string {
	return currentGenre
}

// Package destruct implements destructible environment objects.
package destruct

import (
	"sync"
)

// Destructible represents a destructible object in the world.
type Destructible struct {
	ID        string
	Health    float64
	MaxHealth float64
	Destroyed bool
	X, Y      float64
	Type      string
	DropItems []string
	mu        sync.RWMutex
}

// System manages destructible objects in a level.
type System struct {
	objects map[string]*Destructible
	mu      sync.RWMutex
}

// NewSystem creates a new destructible system.
func NewSystem() *System {
	return &System{
		objects: make(map[string]*Destructible),
	}
}

// NewDestructible creates a new destructible object.
func NewDestructible(id, objType string, health, x, y float64) *Destructible {
	return &Destructible{
		ID:        id,
		Health:    health,
		MaxHealth: health,
		Destroyed: false,
		X:         x,
		Y:         y,
		Type:      objType,
		DropItems: make([]string, 0),
	}
}

// Add adds a destructible object to the system.
func (s *System) Add(d *Destructible) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[d.ID] = d
}

// Remove removes a destructible object from the system.
func (s *System) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, id)
}

// Get retrieves a destructible object by ID.
func (s *System) Get(id string) (*Destructible, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.objects[id]
	return d, ok
}

// GetAll returns all destructible objects.
func (s *System) GetAll() []*Destructible {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Destructible, 0, len(s.objects))
	for _, d := range s.objects {
		result = append(result, d)
	}
	return result
}

// Damage applies damage to a destructible object.
func (d *Destructible) Damage(amount float64) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Destroyed {
		return false
	}

	d.Health -= amount
	if d.Health <= 0 {
		d.Health = 0
		d.Destroyed = true
		return true // Destroyed
	}
	return false
}

// Destroy immediately destroys the object.
func (d *Destructible) Destroy() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.Health = 0
	d.Destroyed = true
}

// Repair restores health to a destructible object.
func (d *Destructible) Repair(amount float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Destroyed {
		return
	}

	d.Health += amount
	if d.Health > d.MaxHealth {
		d.Health = d.MaxHealth
	}
}

// IsDestroyed returns whether the object is destroyed.
func (d *Destructible) IsDestroyed() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Destroyed
}

// GetHealth returns current health.
func (d *Destructible) GetHealth() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Health
}

// AddDropItem adds an item ID to the drop list.
func (d *Destructible) AddDropItem(itemID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.DropItems = append(d.DropItems, itemID)
}

// GetDropItems returns the drop item list.
func (d *Destructible) GetDropItems() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]string, len(d.DropItems))
	copy(result, d.DropItems)
	return result
}

// SetGenre configures destructible types for a genre.
func SetGenre(genreID string) {}

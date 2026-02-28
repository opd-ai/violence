// Package status manages status effects applied to entities.
package status

import "time"

// Effect represents a status effect that can be applied to an entity.
type Effect struct {
	Name          string
	Duration      time.Duration
	DamagePerTick float64
}

// Registry holds all known status effects.
type Registry struct {
	effects map[string]Effect
}

// NewRegistry creates a new status effect registry.
func NewRegistry() *Registry {
	return &Registry{effects: make(map[string]Effect)}
}

// Apply adds a status effect to an entity.
func (r *Registry) Apply(name string) {}

// Tick advances all active effects by one tick.
func (r *Registry) Tick() {}

var currentGenre = "fantasy"

// SetGenre configures status effects for a genre.
func SetGenre(genreID string) {
	currentGenre = genreID
}

// GetCurrentGenre returns the current global genre setting.
func GetCurrentGenre() string {
	return currentGenre
}

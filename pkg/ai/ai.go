// Package ai implements enemy artificial intelligence behaviors.
package ai

// Behavior represents an AI behavior type.
type Behavior int

const (
	Idle Behavior = iota
	Patrol
	Chase
	Attack
)

// Enemy represents an AI-controlled enemy.
type Enemy struct {
	X, Y     float64
	Health   float64
	Behavior Behavior
}

// Update advances the enemy AI by one tick.
func (e *Enemy) Update() {}

// SetGenre configures AI behaviors for a genre.
func SetGenre(genreID string) {}

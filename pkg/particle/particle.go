// Package particle implements a particle effects system.
package particle

// Particle represents a single particle.
type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
}

// Emitter spawns and manages particles.
type Emitter struct {
	Particles []Particle
	Rate      float64
}

// NewEmitter creates a particle emitter.
func NewEmitter(rate float64) *Emitter {
	return &Emitter{Rate: rate}
}

// Emit spawns new particles.
func (e *Emitter) Emit() {}

// Update advances all particles by one tick.
func (e *Emitter) Update() {}

// SetGenre configures particle effects for a genre.
func SetGenre(genreID string) {}

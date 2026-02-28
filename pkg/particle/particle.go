// Package particle implements a particle effects system with pooling and spatial culling.
package particle

import (
	"image/color"
	"math"
	"math/rand"
)

// Particle represents a single particle with position, velocity, lifetime, color, and size.
type Particle struct {
	X, Y       float64 // World position
	Z          float64 // Height (for 3D-like effects)
	VX, VY     float64 // Velocity
	VZ         float64 // Vertical velocity
	Life       float64 // Current life remaining (seconds)
	MaxLife    float64 // Initial life (for fade calculations)
	R, G, B, A uint8   // Color components
	Size       float64 // Particle size
	Active     bool    // Whether particle is alive
}

// ParticleSystem manages a pool of particles with spawn/update/cull lifecycle.
type ParticleSystem struct {
	particles []Particle
	poolSize  int
	nextIndex int
	rng       *rand.Rand
	genreID   string

	// Spatial culling bounds
	minX, maxX float64
	minY, maxY float64
}

// NewParticleSystem creates a particle system with a pre-allocated pool.
func NewParticleSystem(poolSize int, seed int64) *ParticleSystem {
	if poolSize <= 0 {
		poolSize = 1024 // default pool size
	}

	return &ParticleSystem{
		particles: make([]Particle, poolSize),
		poolSize:  poolSize,
		rng:       rand.New(rand.NewSource(seed)),
		minX:      -1000,
		maxX:      1000,
		minY:      -1000,
		maxY:      1000,
	}
}

// SetSpatialBounds defines culling boundaries for particles.
func (ps *ParticleSystem) SetSpatialBounds(minX, maxX, minY, maxY float64) {
	ps.minX = minX
	ps.maxX = maxX
	ps.minY = minY
	ps.maxY = maxY
}

// SetGenre configures particle behavior for a specific genre.
func (ps *ParticleSystem) SetGenre(genreID string) {
	ps.genreID = genreID
}

// Spawn creates a new particle from the pool.
func (ps *ParticleSystem) Spawn(x, y, z, vx, vy, vz, life, size float64, c color.RGBA) *Particle {
	// Find next available particle in pool
	startIndex := ps.nextIndex
	for {
		p := &ps.particles[ps.nextIndex]
		ps.nextIndex = (ps.nextIndex + 1) % ps.poolSize

		if !p.Active {
			// Reuse this particle
			p.X = x
			p.Y = y
			p.Z = z
			p.VX = vx
			p.VY = vy
			p.VZ = vz
			p.Life = life
			p.MaxLife = life
			p.R = c.R
			p.G = c.G
			p.B = c.B
			p.A = c.A
			p.Size = size
			p.Active = true
			return p
		}

		// Wrapped around without finding a free particle
		if ps.nextIndex == startIndex {
			return nil
		}
	}
}

// SpawnBurst creates multiple particles at once with randomized velocities.
func (ps *ParticleSystem) SpawnBurst(x, y, z float64, count int, speed, spread, life, size float64, c color.RGBA) {
	for i := 0; i < count; i++ {
		angle := ps.rng.Float64() * 2 * math.Pi
		velocity := speed * (0.5 + ps.rng.Float64()*0.5)

		vx := math.Cos(angle) * velocity
		vy := math.Sin(angle) * velocity
		vz := (ps.rng.Float64()*2 - 1) * spread

		ps.Spawn(x, y, z, vx, vy, vz, life, size, c)
	}
}

// Update advances all active particles by deltaTime seconds.
func (ps *ParticleSystem) Update(deltaTime float64) {
	for i := range ps.particles {
		p := &ps.particles[i]
		if !p.Active {
			continue
		}

		// Update position
		p.X += p.VX * deltaTime
		p.Y += p.VY * deltaTime
		p.Z += p.VZ * deltaTime

		// Update life
		p.Life -= deltaTime
		if p.Life <= 0 {
			p.Active = false
			continue
		}

		// Spatial culling
		if p.X < ps.minX || p.X > ps.maxX || p.Y < ps.minY || p.Y > ps.maxY {
			p.Active = false
			continue
		}

		// Fade alpha based on remaining life
		lifeFraction := p.Life / p.MaxLife
		p.A = uint8(float64(color.RGBA{p.R, p.G, p.B, 255}.A) * lifeFraction)
	}
}

// GetActiveParticles returns a slice of all currently active particles.
func (ps *ParticleSystem) GetActiveParticles() []Particle {
	active := make([]Particle, 0, ps.poolSize/4)
	for i := range ps.particles {
		if ps.particles[i].Active {
			active = append(active, ps.particles[i])
		}
	}
	return active
}

// GetActiveCount returns the number of currently active particles.
func (ps *ParticleSystem) GetActiveCount() int {
	count := 0
	for i := range ps.particles {
		if ps.particles[i].Active {
			count++
		}
	}
	return count
}

// Clear deactivates all particles.
func (ps *ParticleSystem) Clear() {
	for i := range ps.particles {
		ps.particles[i].Active = false
	}
	ps.nextIndex = 0
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

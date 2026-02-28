package particle

import "image/color"

// System wraps ParticleSystem with convenience methods for common operations.
type System struct {
	*ParticleSystem
}

// NewSystem creates a new particle system wrapper.
func NewSystem(poolSize int, seed int64) *System {
	return &System{
		ParticleSystem: NewParticleSystem(poolSize, seed),
	}
}

// SpawnPoint creates a single point particle.
func (s *System) SpawnPoint(x, y float64, c color.RGBA) *Particle {
	return s.Spawn(x, y, 0, 0, 0, 0, 1.0, 1.0, c)
}

// SpawnTrail creates a particle trail effect behind a moving object.
func (s *System) SpawnTrail(x, y, vx, vy float64, c color.RGBA) {
	s.Spawn(x, y, 0, vx*-0.1, vy*-0.1, 0, 0.5, 0.8, c)
}

// SpawnExplosion creates an explosion effect at the given position.
func (s *System) SpawnExplosion(x, y, intensity float64) {
	// Inner core - bright yellow/white
	coreColor := color.RGBA{R: 255, G: 255, B: 200, A: 255}
	s.SpawnBurst(x, y, 0, int(intensity*10), intensity*8, 2.0, 0.3, 3.0, coreColor)

	// Outer fire - orange/red
	fireColor := color.RGBA{R: 255, G: 100, B: 0, A: 255}
	s.SpawnBurst(x, y, 0, int(intensity*20), intensity*5, 1.5, 0.6, 2.0, fireColor)

	// Smoke - dark gray
	smokeColor := color.RGBA{R: 50, G: 50, B: 50, A: 200}
	s.SpawnBurst(x, y, 0, int(intensity*15), intensity*2, 0.5, 1.5, 2.5, smokeColor)
}

// SpawnMuzzleFlash creates a muzzle flash effect for weapon firing.
func (s *System) SpawnMuzzleFlash(x, y, angle float64) {
	flashColor := color.RGBA{R: 255, G: 220, B: 100, A: 255}
	s.SpawnBurst(x, y, 0, 8, 20.0, 1.0, 0.1, 1.5, flashColor)
}

// SpawnBlood creates a blood splatter effect.
func (s *System) SpawnBlood(x, y, vx, vy float64) {
	bloodColor := color.RGBA{R: 180, G: 0, B: 0, A: 255}
	s.SpawnBurst(x, y, 0, 15, 8.0, 0.5, 1.0, 1.2, bloodColor)
}

// SpawnSparks creates sparks for bullet impacts on metal.
func (s *System) SpawnSparks(x, y float64) {
	sparkColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}
	s.SpawnBurst(x, y, 0, 12, 15.0, 2.0, 0.4, 0.8, sparkColor)
}

// SpawnSmoke creates rising smoke particles.
func (s *System) SpawnSmoke(x, y float64) {
	smokeColor := color.RGBA{R: 100, G: 100, B: 100, A: 150}
	s.Spawn(x, y, 0, 0, 0, 2.0, 2.0, 2.0, smokeColor)
}

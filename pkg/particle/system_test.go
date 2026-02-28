package particle

import (
	"image/color"
	"testing"
)

func TestNewSystem(t *testing.T) {
	s := NewSystem(100, 12345)
	if s == nil {
		t.Fatal("NewSystem returned nil")
	}
	if s.ParticleSystem == nil {
		t.Fatal("ParticleSystem not initialized")
	}
}

func TestSystemSpawnPoint(t *testing.T) {
	s := NewSystem(10, 12345)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	p := s.SpawnPoint(10, 20, c)
	if p == nil {
		t.Fatal("SpawnPoint returned nil")
	}
	if !p.Active {
		t.Error("particle not active")
	}
	if p.X != 10 || p.Y != 20 {
		t.Errorf("position = (%f, %f), want (10, 20)", p.X, p.Y)
	}
	if p.VX != 0 || p.VY != 0 {
		t.Errorf("point particle has velocity (%f, %f), want (0, 0)", p.VX, p.VY)
	}
}

func TestSystemSpawnTrail(t *testing.T) {
	s := NewSystem(10, 12345)
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	initialCount := s.GetActiveCount()
	s.SpawnTrail(50, 60, 10, 5, c)

	if s.GetActiveCount() != initialCount+1 {
		t.Errorf("active count = %d, want %d", s.GetActiveCount(), initialCount+1)
	}

	particles := s.GetActiveParticles()
	if len(particles) == 0 {
		t.Fatal("no particles spawned")
	}

	p := particles[len(particles)-1]
	if p.X != 50 || p.Y != 60 {
		t.Errorf("trail position = (%f, %f), want (50, 60)", p.X, p.Y)
	}
}

func TestSystemSpawnExplosion(t *testing.T) {
	s := NewSystem(200, 12345)

	initialCount := s.GetActiveCount()
	s.SpawnExplosion(100, 100, 1.0)

	newCount := s.GetActiveCount()
	if newCount <= initialCount {
		t.Error("explosion did not spawn particles")
	}

	// Expect multiple particle layers (core + fire + smoke)
	particleCount := newCount - initialCount
	if particleCount < 30 {
		t.Errorf("explosion spawned %d particles, expected at least 30", particleCount)
	}
}

func TestSystemSpawnMuzzleFlash(t *testing.T) {
	s := NewSystem(50, 12345)

	initialCount := s.GetActiveCount()
	s.SpawnMuzzleFlash(25, 25, 0)

	newCount := s.GetActiveCount()
	if newCount <= initialCount {
		t.Error("muzzle flash did not spawn particles")
	}
}

func TestSystemSpawnBlood(t *testing.T) {
	s := NewSystem(50, 12345)

	initialCount := s.GetActiveCount()
	s.SpawnBlood(30, 30, 5, -2)

	newCount := s.GetActiveCount()
	if newCount <= initialCount {
		t.Error("blood splatter did not spawn particles")
	}
}

func TestSystemSpawnSparks(t *testing.T) {
	s := NewSystem(50, 12345)

	initialCount := s.GetActiveCount()
	s.SpawnSparks(40, 40)

	newCount := s.GetActiveCount()
	if newCount <= initialCount {
		t.Error("sparks did not spawn particles")
	}
}

func TestSystemSpawnSmoke(t *testing.T) {
	s := NewSystem(50, 12345)

	initialCount := s.GetActiveCount()
	s.SpawnSmoke(15, 15)

	newCount := s.GetActiveCount()
	if newCount != initialCount+1 {
		t.Errorf("smoke spawned %d particles, want 1", newCount-initialCount)
	}

	particles := s.GetActiveParticles()
	if len(particles) == 0 {
		t.Fatal("no smoke particle spawned")
	}

	// Smoke should rise (positive VZ)
	p := particles[len(particles)-1]
	if p.VZ <= 0 {
		t.Errorf("smoke VZ = %f, want positive (rising)", p.VZ)
	}
}

func TestOldEmitterAPI(t *testing.T) {
	// Test legacy Emitter API for backwards compatibility
	e := NewEmitter(10.0)
	if e == nil {
		t.Fatal("NewEmitter returned nil")
	}
	if e.Rate != 10.0 {
		t.Errorf("rate = %f, want 10.0", e.Rate)
	}

	// These are stubs but should not panic
	e.Emit()
	e.Update()
	SetGenre("fantasy")
}

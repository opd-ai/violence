package particle

import (
	"image/color"
	"testing"
)

func TestNewParticleSystem(t *testing.T) {
	tests := []struct {
		name     string
		poolSize int
		want     int
	}{
		{"default pool size", 0, 1024},
		{"custom pool size", 512, 512},
		{"large pool", 2048, 2048},
		{"negative pool size defaults", -1, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(tt.poolSize, 12345)
			if ps == nil {
				t.Fatal("NewParticleSystem returned nil")
			}
			if len(ps.particles) != tt.want {
				t.Errorf("poolSize = %d, want %d", len(ps.particles), tt.want)
			}
			if ps.rng == nil {
				t.Error("rng not initialized")
			}
		})
	}
}

func TestParticleSystemSpawn(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// Spawn a particle
	p := ps.Spawn(10, 20, 0, 1, 2, 0, 5.0, 2.0, c)
	if p == nil {
		t.Fatal("Spawn returned nil")
	}
	if !p.Active {
		t.Error("spawned particle not active")
	}
	if p.X != 10 || p.Y != 20 {
		t.Errorf("position = (%f, %f), want (10, 20)", p.X, p.Y)
	}
	if p.VX != 1 || p.VY != 2 {
		t.Errorf("velocity = (%f, %f), want (1, 2)", p.VX, p.VY)
	}
	if p.Life != 5.0 || p.MaxLife != 5.0 {
		t.Errorf("life = %f, maxLife = %f, want 5.0", p.Life, p.MaxLife)
	}
	if p.Size != 2.0 {
		t.Errorf("size = %f, want 2.0", p.Size)
	}
	if p.R != 255 || p.G != 0 || p.B != 0 {
		t.Errorf("color = (%d, %d, %d), want (255, 0, 0)", p.R, p.G, p.B)
	}
}

func TestParticleSystemSpawnPoolExhaustion(t *testing.T) {
	ps := NewParticleSystem(2, 12345)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// Fill the pool
	p1 := ps.Spawn(0, 0, 0, 0, 0, 0, 1, 1, c)
	p2 := ps.Spawn(1, 1, 0, 0, 0, 0, 1, 1, c)

	if p1 == nil || p2 == nil {
		t.Fatal("failed to spawn initial particles")
	}

	// Pool should be full now
	p3 := ps.Spawn(2, 2, 0, 0, 0, 0, 1, 1, c)
	if p3 != nil {
		t.Error("expected nil when pool exhausted, got particle")
	}

	// Deactivate one particle
	p1.Active = false

	// Should be able to spawn again
	p4 := ps.Spawn(3, 3, 0, 0, 0, 0, 1, 1, c)
	if p4 == nil {
		t.Error("expected particle after one freed, got nil")
	}
}

func TestParticleSystemSpawnBurst(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	c := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	ps.SpawnBurst(50, 50, 0, 20, 10.0, 2.0, 3.0, 1.5, c)

	count := ps.GetActiveCount()
	if count != 20 {
		t.Errorf("active count = %d, want 20", count)
	}

	// Verify particles have different velocities (randomized)
	particles := ps.GetActiveParticles()
	if len(particles) != 20 {
		t.Fatalf("GetActiveParticles returned %d, want 20", len(particles))
	}

	// Check at least some variety in velocities
	sameVelocity := true
	firstVX := particles[0].VX
	for _, p := range particles[1:] {
		if p.VX != firstVX {
			sameVelocity = false
			break
		}
	}
	if sameVelocity {
		t.Error("all particles have same velocity, expected randomization")
	}
}

func TestParticleSystemUpdate(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	p := ps.Spawn(0, 0, 0, 10, 5, 0, 2.0, 1.0, c)
	if p == nil {
		t.Fatal("Spawn failed")
	}

	// Update by 0.1 seconds
	ps.Update(0.1)

	// Check position changed
	expectedX := 10 * 0.1
	expectedY := 5 * 0.1
	if p.X != expectedX || p.Y != expectedY {
		t.Errorf("position = (%f, %f), want (%f, %f)", p.X, p.Y, expectedX, expectedY)
	}

	// Check life decreased
	expectedLife := 2.0 - 0.1
	if p.Life != expectedLife {
		t.Errorf("life = %f, want %f", p.Life, expectedLife)
	}

	// Particle should still be active
	if !p.Active {
		t.Error("particle deactivated too early")
	}
}

func TestParticleSystemUpdateLifeExpiration(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	p := ps.Spawn(0, 0, 0, 0, 0, 0, 1.0, 1.0, c)
	if p == nil {
		t.Fatal("Spawn failed")
	}

	// Update past life expectancy
	ps.Update(1.5)

	if p.Active {
		t.Error("particle still active after life expired")
	}
}

func TestParticleSystemUpdateSpatialCulling(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	ps.SetSpatialBounds(0, 100, 0, 100)
	c := color.RGBA{R: 100, G: 100, B: 100, A: 255}

	// Spawn particle that will move out of bounds
	p := ps.Spawn(50, 50, 0, 100, 0, 0, 10.0, 1.0, c)
	if p == nil {
		t.Fatal("Spawn failed")
	}

	// Update to move particle out of bounds
	ps.Update(1.0)

	if p.Active {
		t.Error("particle still active after moving out of bounds")
	}
}

func TestParticleSystemUpdateAlphaFade(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	p := ps.Spawn(0, 0, 0, 0, 0, 0, 2.0, 1.0, c)
	if p == nil {
		t.Fatal("Spawn failed")
	}

	initialAlpha := p.A

	// Update by half the life
	ps.Update(1.0)

	// Alpha should be approximately half (life fraction = 0.5)
	if p.A >= initialAlpha {
		t.Errorf("alpha not fading: initial=%d, current=%d", initialAlpha, p.A)
	}

	// Update to near end of life
	ps.Update(0.9)

	// Alpha should be very low
	if p.A > 50 {
		t.Errorf("alpha still high near end of life: %d", p.A)
	}
}

func TestParticleSystemSetSpatialBounds(t *testing.T) {
	ps := NewParticleSystem(10, 12345)

	ps.SetSpatialBounds(10, 20, 30, 40)

	if ps.minX != 10 || ps.maxX != 20 {
		t.Errorf("X bounds = (%f, %f), want (10, 20)", ps.minX, ps.maxX)
	}
	if ps.minY != 30 || ps.maxY != 40 {
		t.Errorf("Y bounds = (%f, %f), want (30, 40)", ps.minY, ps.maxY)
	}
}

func TestParticleSystemSetGenre(t *testing.T) {
	ps := NewParticleSystem(10, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			ps.SetGenre(genre)
			if ps.genreID != genre {
				t.Errorf("genreID = %s, want %s", ps.genreID, genre)
			}
		})
	}
}

func TestParticleSystemGetActiveParticles(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Initially empty
	active := ps.GetActiveParticles()
	if len(active) != 0 {
		t.Errorf("initial active count = %d, want 0", len(active))
	}

	// Spawn some particles
	ps.Spawn(0, 0, 0, 0, 0, 0, 1, 1, c)
	ps.Spawn(1, 1, 0, 0, 0, 0, 1, 1, c)
	ps.Spawn(2, 2, 0, 0, 0, 0, 1, 1, c)

	active = ps.GetActiveParticles()
	if len(active) != 3 {
		t.Errorf("active count = %d, want 3", len(active))
	}
}

func TestParticleSystemGetActiveCount(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	if count := ps.GetActiveCount(); count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	ps.Spawn(0, 0, 0, 0, 0, 0, 1, 1, c)
	ps.Spawn(1, 1, 0, 0, 0, 0, 1, 1, c)

	if count := ps.GetActiveCount(); count != 2 {
		t.Errorf("count = %d, want 2", count)
	}

	// Expire particles
	ps.Update(2.0)

	if count := ps.GetActiveCount(); count != 0 {
		t.Errorf("count after expiration = %d, want 0", count)
	}
}

func TestParticleSystemClear(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Spawn particles
	ps.Spawn(0, 0, 0, 0, 0, 0, 1, 1, c)
	ps.Spawn(1, 1, 0, 0, 0, 0, 1, 1, c)
	ps.Spawn(2, 2, 0, 0, 0, 0, 1, 1, c)

	if count := ps.GetActiveCount(); count != 3 {
		t.Fatalf("initial count = %d, want 3", count)
	}

	ps.Clear()

	if count := ps.GetActiveCount(); count != 0 {
		t.Errorf("count after clear = %d, want 0", count)
	}

	if ps.nextIndex != 0 {
		t.Errorf("nextIndex = %d, want 0 after clear", ps.nextIndex)
	}
}

func TestParticleSystemDeterminism(t *testing.T) {
	// Two systems with same seed should produce identical results
	ps1 := NewParticleSystem(100, 42)
	ps2 := NewParticleSystem(100, 42)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	ps1.SpawnBurst(0, 0, 0, 10, 5.0, 1.0, 1.0, 1.0, c)
	ps2.SpawnBurst(0, 0, 0, 10, 5.0, 1.0, 1.0, 1.0, c)

	particles1 := ps1.GetActiveParticles()
	particles2 := ps2.GetActiveParticles()

	if len(particles1) != len(particles2) {
		t.Fatalf("particle count mismatch: %d vs %d", len(particles1), len(particles2))
	}

	for i := range particles1 {
		if particles1[i].VX != particles2[i].VX || particles1[i].VY != particles2[i].VY {
			t.Errorf("particle %d velocity mismatch: (%f,%f) vs (%f,%f)",
				i, particles1[i].VX, particles1[i].VY, particles2[i].VX, particles2[i].VY)
		}
	}
}

func TestParticleSystem3DMovement(t *testing.T) {
	ps := NewParticleSystem(10, 12345)
	c := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	p := ps.Spawn(0, 0, 10, 0, 0, -5, 2.0, 1.0, c)
	if p == nil {
		t.Fatal("Spawn failed")
	}

	initialZ := p.Z

	ps.Update(0.5)

	expectedZ := initialZ + (-5 * 0.5)
	if p.Z != expectedZ {
		t.Errorf("Z position = %f, want %f", p.Z, expectedZ)
	}
}

func BenchmarkParticleSystemSpawn(b *testing.B) {
	ps := NewParticleSystem(10000, 12345)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Spawn(0, 0, 0, 1, 1, 0, 1.0, 1.0, c)
		if i%10000 == 9999 {
			ps.Clear()
		}
	}
}

func BenchmarkParticleSystemUpdate(b *testing.B) {
	ps := NewParticleSystem(1000, 12345)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	// Fill with particles
	for i := 0; i < 1000; i++ {
		ps.Spawn(float64(i), float64(i), 0, 1, 1, 0, 100.0, 1.0, c)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Update(0.016)
	}
}

func BenchmarkParticleSystemSpawnBurst(b *testing.B) {
	ps := NewParticleSystem(10000, 12345)
	c := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.SpawnBurst(0, 0, 0, 50, 10.0, 2.0, 1.0, 1.0, c)
		if i%100 == 99 {
			ps.Clear()
		}
	}
}

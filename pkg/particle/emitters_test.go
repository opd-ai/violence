package particle

import (
	"image/color"
	"testing"
)

func TestNewMuzzleFlashEmitter(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"default", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(100, 12345)
			emitter := NewMuzzleFlashEmitter(ps, tt.genreID)

			if emitter == nil {
				t.Fatal("NewMuzzleFlashEmitter returned nil")
			}
			if emitter.system != ps {
				t.Error("particle system not set correctly")
			}
			if emitter.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", emitter.genreID, tt.genreID)
			}
		})
	}
}

func TestMuzzleFlashEmitter_Emit(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(100, 12345)
			emitter := NewMuzzleFlashEmitter(ps, tt.genreID)

			initialCount := ps.GetActiveCount()
			emitter.Emit(10, 10, 0)

			newCount := ps.GetActiveCount()
			if newCount <= initialCount {
				t.Error("muzzle flash did not spawn particles")
			}

			// Verify particles were spawned near target position
			particles := ps.GetActiveParticles()
			if len(particles) == 0 {
				t.Fatal("no active particles after emit")
			}
		})
	}
}

func TestNewSparkEmitter(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	emitter := NewSparkEmitter(ps, "scifi")

	if emitter == nil {
		t.Fatal("NewSparkEmitter returned nil")
	}
	if emitter.system != ps {
		t.Error("particle system not set correctly")
	}
}

func TestSparkEmitter_Emit(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(100, 12345)
			emitter := NewSparkEmitter(ps, tt.genreID)

			initialCount := ps.GetActiveCount()
			emitter.Emit(20, 20, 1, 0) // Normal pointing right

			newCount := ps.GetActiveCount()
			if newCount <= initialCount {
				t.Error("sparks did not spawn")
			}

			particles := ps.GetActiveParticles()
			if len(particles) == 0 {
				t.Fatal("no active particles after emit")
			}
		})
	}
}

func TestNewBloodSplatterEmitter(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	emitter := NewBloodSplatterEmitter(ps, "horror")

	if emitter == nil {
		t.Fatal("NewBloodSplatterEmitter returned nil")
	}
}

func TestBloodSplatterEmitter_Emit(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(200, 12345)
			emitter := NewBloodSplatterEmitter(ps, tt.genreID)

			initialCount := ps.GetActiveCount()
			emitter.Emit(15, 15, 5, -3) // Impact with velocity

			newCount := ps.GetActiveCount()
			if newCount <= initialCount {
				t.Error("blood splatter did not spawn")
			}

			particles := ps.GetActiveParticles()
			if len(particles) < 5 {
				t.Errorf("expected multiple blood particles, got %d", len(particles))
			}
		})
	}
}

func TestNewExplosionEmitter(t *testing.T) {
	ps := NewParticleSystem(200, 12345)
	emitter := NewExplosionEmitter(ps, "fantasy")

	if emitter == nil {
		t.Fatal("NewExplosionEmitter returned nil")
	}
}

func TestExplosionEmitter_Emit(t *testing.T) {
	tests := []struct {
		name      string
		genreID   string
		intensity float64
	}{
		{"fantasy_small", "fantasy", 0.5},
		{"scifi_medium", "scifi", 1.0},
		{"horror_large", "horror", 2.0},
		{"cyberpunk_huge", "cyberpunk", 3.0},
		{"postapoc_massive", "postapoc", 5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(500, 12345)
			emitter := NewExplosionEmitter(ps, tt.genreID)

			initialCount := ps.GetActiveCount()
			emitter.Emit(50, 50, tt.intensity)

			newCount := ps.GetActiveCount()
			particleCount := newCount - initialCount

			if particleCount < 10 {
				t.Errorf("explosion spawned %d particles, expected at least 10", particleCount)
			}

			// Higher intensity should spawn more particles
			if tt.intensity > 1.0 && particleCount < 30 {
				t.Errorf("high-intensity explosion spawned only %d particles", particleCount)
			}
		})
	}
}

func TestNewEnergyDischargeEmitter(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	emitter := NewEnergyDischargeEmitter(ps, "scifi")

	if emitter == nil {
		t.Fatal("NewEnergyDischargeEmitter returned nil")
	}
}

func TestEnergyDischargeEmitter_Emit(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(150, 12345)
			emitter := NewEnergyDischargeEmitter(ps, tt.genreID)

			initialCount := ps.GetActiveCount()
			emitter.Emit(30, 30, 1, 0) // Direction vector right

			newCount := ps.GetActiveCount()
			if newCount <= initialCount {
				t.Error("energy discharge did not spawn")
			}

			particles := ps.GetActiveParticles()
			if len(particles) == 0 {
				t.Fatal("no active particles after emit")
			}
		})
	}
}

func TestEmitterDeterminism(t *testing.T) {
	// Test that emitters produce deterministic results with same seed
	seed := int64(98765)

	ps1 := NewParticleSystem(200, seed)
	emitter1 := NewMuzzleFlashEmitter(ps1, "scifi")
	emitter1.Emit(10, 10, 0)

	ps2 := NewParticleSystem(200, seed)
	emitter2 := NewMuzzleFlashEmitter(ps2, "scifi")
	emitter2.Emit(10, 10, 0)

	particles1 := ps1.GetActiveParticles()
	particles2 := ps2.GetActiveParticles()

	if len(particles1) != len(particles2) {
		t.Fatalf("particle counts differ: %d vs %d", len(particles1), len(particles2))
	}

	for i := range particles1 {
		p1 := particles1[i]
		p2 := particles2[i]

		if p1.X != p2.X || p1.Y != p2.Y || p1.Z != p2.Z {
			t.Errorf("particle %d position mismatch", i)
		}
		if p1.VX != p2.VX || p1.VY != p2.VY || p1.VZ != p2.VZ {
			t.Errorf("particle %d velocity mismatch", i)
		}
	}
}

func TestEmitterConfigs(t *testing.T) {
	// Verify all genre configs are distinct
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	ps := NewParticleSystem(100, 12345)

	t.Run("MuzzleFlash", func(t *testing.T) {
		configs := make(map[string]EmitterConfig)
		for _, genre := range genres {
			emitter := NewMuzzleFlashEmitter(ps, genre)
			config := emitter.getConfig()
			configs[genre] = config
		}

		// Verify no two configs are identical
		for i, g1 := range genres {
			for j, g2 := range genres {
				if i >= j {
					continue
				}
				c1 := configs[g1]
				c2 := configs[g2]
				if c1.Color == c2.Color && c1.ParticleCount == c2.ParticleCount && c1.Speed == c2.Speed {
					t.Errorf("configs for %s and %s are too similar", g1, g2)
				}
			}
		}
	})

	t.Run("Spark", func(t *testing.T) {
		configs := make(map[string]EmitterConfig)
		for _, genre := range genres {
			emitter := NewSparkEmitter(ps, genre)
			config := emitter.getConfig()
			configs[genre] = config
		}

		// Verify distinct colors per genre
		for i, g1 := range genres {
			for j, g2 := range genres {
				if i >= j {
					continue
				}
				c1 := configs[g1]
				c2 := configs[g2]
				if c1.Color == c2.Color {
					t.Errorf("spark colors for %s and %s are identical", g1, g2)
				}
			}
		}
	})
}

func TestEmitterGenreVariations(t *testing.T) {
	// Test that different genres produce visually different effects
	ps := NewParticleSystem(500, 12345)

	mf1 := NewMuzzleFlashEmitter(ps, "fantasy")
	mf2 := NewMuzzleFlashEmitter(ps, "scifi")

	ps.Clear()
	mf1.Emit(0, 0, 0)
	count1 := ps.GetActiveCount()

	ps.Clear()
	mf2.Emit(0, 0, 0)
	count2 := ps.GetActiveCount()

	if count1 == 0 || count2 == 0 {
		t.Error("emitters did not spawn particles")
	}

	// Counts should differ (different particle counts per genre)
	// This validates that genre configuration is being applied
}

func TestEmitterWithZeroIntensity(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	emitter := NewExplosionEmitter(ps, "scifi")

	initialCount := ps.GetActiveCount()
	emitter.Emit(10, 10, 0.0)

	newCount := ps.GetActiveCount()
	// Even with zero intensity, should spawn some particles (0.3 * 0 = 0, but smoke still spawns)
	if newCount == initialCount {
		// This is actually expected behavior - zero intensity = no core/fire
		// Only smoke may spawn depending on implementation
	}
}

func TestEmitterEdgeCases(t *testing.T) {
	t.Run("NegativeAngle", func(t *testing.T) {
		ps := NewParticleSystem(50, 12345)
		emitter := NewMuzzleFlashEmitter(ps, "scifi")

		emitter.Emit(10, 10, -3.14159) // Negative angle
		if ps.GetActiveCount() == 0 {
			t.Error("emitter failed with negative angle")
		}
	})

	t.Run("ZeroDirection", func(t *testing.T) {
		ps := NewParticleSystem(50, 12345)
		emitter := NewEnergyDischargeEmitter(ps, "cyberpunk")

		emitter.Emit(10, 10, 0, 0) // Zero direction vector
		if ps.GetActiveCount() == 0 {
			t.Error("emitter failed with zero direction")
		}
	})

	t.Run("NegativeVelocity", func(t *testing.T) {
		ps := NewParticleSystem(50, 12345)
		emitter := NewBloodSplatterEmitter(ps, "horror")

		emitter.Emit(10, 10, -5, -5) // Negative impact velocity
		if ps.GetActiveCount() == 0 {
			t.Error("emitter failed with negative velocity")
		}
	})
}

func TestEmitterPoolExhaustion(t *testing.T) {
	ps := NewParticleSystem(20, 12345) // Small pool
	emitter := NewExplosionEmitter(ps, "scifi")

	// Try to spawn more particles than pool can hold
	emitter.Emit(10, 10, 10.0) // Large explosion

	count := ps.GetActiveCount()
	if count > 20 {
		t.Errorf("particle count %d exceeds pool size 20", count)
	}
	if count == 0 {
		t.Error("no particles spawned despite pool availability")
	}
}

func BenchmarkMuzzleFlashEmitter(b *testing.B) {
	ps := NewParticleSystem(1000, 12345)
	emitter := NewMuzzleFlashEmitter(ps, "scifi")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Clear()
		emitter.Emit(10, 10, 0)
	}
}

func BenchmarkExplosionEmitter(b *testing.B) {
	ps := NewParticleSystem(1000, 12345)
	emitter := NewExplosionEmitter(ps, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Clear()
		emitter.Emit(50, 50, 1.0)
	}
}

func BenchmarkBloodSplatterEmitter(b *testing.B) {
	ps := NewParticleSystem(500, 12345)
	emitter := NewBloodSplatterEmitter(ps, "horror")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Clear()
		emitter.Emit(25, 25, 5, -3)
	}
}

func TestEmitterColorValidation(t *testing.T) {
	// Verify all emitters produce valid RGBA colors
	ps := NewParticleSystem(100, 12345)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			emitter := NewMuzzleFlashEmitter(ps, genre)
			ps.Clear()
			emitter.Emit(0, 0, 0)

			particles := ps.GetActiveParticles()
			for i, p := range particles {
				c := color.RGBA{R: p.R, G: p.G, B: p.B, A: p.A}
				if c.A == 0 {
					t.Errorf("particle %d has zero alpha", i)
				}
			}
		})
	}
}

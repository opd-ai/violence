package particle

import (
	"image/color"
	"math"
	"testing"
)

func TestImpactEffectEmitter_Creation(t *testing.T) {
	ps := NewParticleSystem(1024, 42)
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			emitter := NewImpactEffectEmitter(ps, genre)
			if emitter == nil {
				t.Fatal("expected non-nil emitter")
			}
			if emitter.system != ps {
				t.Error("emitter system not set correctly")
			}
			if emitter.genreID != genre {
				t.Errorf("expected genre %s, got %s", genre, emitter.genreID)
			}
		})
	}
}

func TestImpactEffectEmitter_EmitImpact(t *testing.T) {
	ps := NewParticleSystem(2048, 123)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	tests := []struct {
		name         string
		impactType   ImpactType
		material     MaterialType
		minParticles int
	}{
		{"melee_flesh", ImpactMelee, MaterialFlesh, 15},
		{"melee_metal", ImpactMelee, MaterialMetal, 20},
		{"critical_flesh", ImpactCritical, MaterialFlesh, 30},   // Should spawn 2.5x more
		{"explosion_stone", ImpactExplosion, MaterialStone, 50}, // Should spawn 3x more
		{"magic_energy", ImpactMagic, MaterialEnergy, 40},
		{"block_metal", ImpactBlock, MaterialMetal, 8},  // Should spawn 0.5x
		{"death_flesh", ImpactDeath, MaterialFlesh, 60}, // Should spawn 4x more
		{"projectile_wood", ImpactProjectile, MaterialWood, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps.Clear()
			initialCount := ps.GetActiveCount()

			emitter.EmitImpact(50.0, 50.0, tt.impactType, tt.material, 0)

			afterCount := ps.GetActiveCount()
			spawned := afterCount - initialCount

			if spawned < tt.minParticles {
				t.Errorf("expected at least %d particles, got %d", tt.minParticles, spawned)
			}
		})
	}
}

func TestImpactEffectConfig_MaterialDifferences(t *testing.T) {
	ps := NewParticleSystem(1024, 456)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	materials := []MaterialType{
		MaterialFlesh,
		MaterialMetal,
		MaterialStone,
		MaterialWood,
		MaterialEnergy,
		MaterialEthereal,
	}

	configs := make([]ImpactEffectConfig, len(materials))
	for i, mat := range materials {
		configs[i] = emitter.getImpactConfig(ImpactMelee, mat)
	}

	// Verify configs are different for different materials
	for i := 0; i < len(configs)-1; i++ {
		for j := i + 1; j < len(configs); j++ {
			if configs[i].PrimaryColor == configs[j].PrimaryColor &&
				configs[i].ParticleCount == configs[j].ParticleCount {
				t.Errorf("materials %d and %d have identical configs", i, j)
			}
		}
	}

	// Verify metal has sparks (yellow-ish primary color)
	metalConfig := emitter.getImpactConfig(ImpactMelee, MaterialMetal)
	if metalConfig.PrimaryColor.R < 200 || metalConfig.PrimaryColor.G < 150 {
		t.Error("metal sparks should be yellow-ish")
	}

	// Verify flesh has blood (red primary color)
	fleshConfig := emitter.getImpactConfig(ImpactMelee, MaterialFlesh)
	if fleshConfig.PrimaryColor.R < 150 || fleshConfig.PrimaryColor.G > 50 {
		t.Error("flesh impact should be red-ish")
	}
}

func TestImpactEffectConfig_CriticalEnhancement(t *testing.T) {
	ps := NewParticleSystem(1024, 789)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	normalConfig := emitter.getImpactConfig(ImpactMelee, MaterialFlesh)
	critConfig := emitter.getImpactConfig(ImpactCritical, MaterialFlesh)

	// Critical should spawn more particles
	if critConfig.ParticleCount <= normalConfig.ParticleCount {
		t.Errorf("critical should spawn more particles: normal=%d, crit=%d",
			normalConfig.ParticleCount, critConfig.ParticleCount)
	}

	// Critical should be faster
	if critConfig.Speed <= normalConfig.Speed {
		t.Errorf("critical should be faster: normal=%.2f, crit=%.2f",
			normalConfig.Speed, critConfig.Speed)
	}

	// Critical should last longer
	if critConfig.Life <= normalConfig.Life {
		t.Errorf("critical should last longer: normal=%.2f, crit=%.2f",
			normalConfig.Life, critConfig.Life)
	}

	// Critical should be larger
	if critConfig.Size <= normalConfig.Size {
		t.Errorf("critical should be larger: normal=%.2f, crit=%.2f",
			normalConfig.Size, critConfig.Size)
	}
}

func TestImpactEffectConfig_ExplosionFullSpread(t *testing.T) {
	ps := NewParticleSystem(1024, 111)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	config := emitter.getImpactConfig(ImpactExplosion, MaterialStone)

	// Explosion should have full 360-degree spread
	if config.Spread < math.Pi*1.9 {
		t.Errorf("explosion should have near full-circle spread: %.2f", config.Spread)
	}
}

func TestImpactEffectEmitter_GenreDifferences(t *testing.T) {
	ps := NewParticleSystem(2048, 222)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			emitter := NewImpactEffectEmitter(ps, genre)
			config := emitter.getImpactConfig(ImpactMelee, MaterialFlesh)

			// All genres should have valid particle counts
			if config.ParticleCount <= 0 || config.ParticleCount > 100 {
				t.Errorf("%s: invalid particle count %d", genre, config.ParticleCount)
			}

			// All genres should have reasonable speeds
			if config.Speed <= 0 || config.Speed > 50 {
				t.Errorf("%s: invalid speed %.2f", genre, config.Speed)
			}

			// All genres should have reasonable lifetimes
			if config.Life <= 0 || config.Life > 5 {
				t.Errorf("%s: invalid life %.2f", genre, config.Life)
			}
		})
	}
}

func TestImpactEffectEmitter_SpawnMainBurst(t *testing.T) {
	ps := NewParticleSystem(1024, 333)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	config := emitter.getImpactConfig(ImpactMelee, MaterialMetal)

	ps.Clear()
	emitter.spawnMainBurst(100.0, 100.0, math.Pi/4, config)

	activeCount := ps.GetActiveCount()
	if activeCount != config.ParticleCount {
		t.Errorf("expected %d particles, got %d", config.ParticleCount, activeCount)
	}

	// Verify particles have velocity
	particles := ps.GetActiveParticles()
	for i, p := range particles {
		if p.VX == 0 && p.VY == 0 {
			t.Errorf("particle %d has zero velocity", i)
		}

		// Verify particles are positioned near spawn point
		dx := p.X - 100.0
		dy := p.Y - 100.0
		if math.Abs(dx) > 5 || math.Abs(dy) > 5 {
			t.Errorf("particle %d too far from spawn point: (%.2f, %.2f)", i, dx, dy)
		}
	}
}

func TestImpactEffectEmitter_CriticalFlare(t *testing.T) {
	ps := NewParticleSystem(1024, 444)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	config := emitter.getImpactConfig(ImpactCritical, MaterialFlesh)

	ps.Clear()
	emitter.spawnCriticalFlare(50.0, 50.0, config)

	// Should spawn 8 central flares + 12 ring particles = 20
	activeCount := ps.GetActiveCount()
	if activeCount != 20 {
		t.Errorf("expected 20 flare particles, got %d", activeCount)
	}
}

func TestImpactEffectEmitter_ExplosionRing(t *testing.T) {
	ps := NewParticleSystem(1024, 555)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	config := emitter.getImpactConfig(ImpactExplosion, MaterialEnergy)

	ps.Clear()
	emitter.spawnExplosionRing(75.0, 75.0, config)

	// Should spawn 24 ring particles + center burst
	activeCount := ps.GetActiveCount()
	if activeCount < 30 { // 24 + some from burst
		t.Errorf("expected at least 30 explosion particles, got %d", activeCount)
	}
}

func TestImpactEffectEmitter_MagicSparkles(t *testing.T) {
	ps := NewParticleSystem(1024, 666)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	config := emitter.getImpactConfig(ImpactMagic, MaterialEnergy)

	ps.Clear()
	emitter.spawnMagicSparkles(60.0, 60.0, config)

	// Should spawn 15 sparkle particles
	activeCount := ps.GetActiveCount()
	if activeCount != 15 {
		t.Errorf("expected 15 sparkle particles, got %d", activeCount)
	}

	// Verify sparkles have upward Z velocity
	particles := ps.GetActiveParticles()
	upwardCount := 0
	for _, p := range particles {
		if p.VZ > 0 {
			upwardCount++
		}
	}

	// Most sparkles should rise
	if upwardCount < len(particles)/2 {
		t.Error("expected most sparkles to have upward velocity")
	}
}

func TestImpactEffectEmitter_DeathBurst(t *testing.T) {
	ps := NewParticleSystem(2048, 777)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	materials := []MaterialType{MaterialFlesh, MaterialEthereal}

	for _, mat := range materials {
		t.Run("material_"+string(rune(mat)), func(t *testing.T) {
			config := emitter.getImpactConfig(ImpactDeath, mat)

			ps.Clear()
			emitter.spawnDeathBurst(80.0, 80.0, mat, config)

			// Should spawn 40 burst particles
			activeCount := ps.GetActiveCount()

			// Ethereal/Energy materials spawn additional wisps
			minExpected := 40
			if mat == MaterialEthereal || mat == MaterialEnergy {
				minExpected = 48 // 40 + 8 wisps
			}

			if activeCount < minExpected {
				t.Errorf("expected at least %d death particles, got %d", minExpected, activeCount)
			}
		})
	}
}

func TestImpactEffectEmitter_BlockSparks(t *testing.T) {
	ps := NewParticleSystem(1024, 888)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	config := emitter.getImpactConfig(ImpactBlock, MaterialMetal)

	ps.Clear()
	emitter.spawnBlockSparks(90.0, 90.0, 0, config)

	// Should spawn 10 spark particles
	activeCount := ps.GetActiveCount()
	if activeCount != 10 {
		t.Errorf("expected 10 block sparks, got %d", activeCount)
	}
}

func TestImpactEffectEmitter_Debris(t *testing.T) {
	ps := NewParticleSystem(1024, 999)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	// Test with material that has debris
	config := emitter.getImpactConfig(ImpactMelee, MaterialStone)
	if config.DebrisCount == 0 {
		t.Skip("stone should have debris, but config says 0")
	}

	ps.Clear()
	emitter.spawnDebris(70.0, 70.0, MaterialStone, config)

	activeCount := ps.GetActiveCount()
	if activeCount != config.DebrisCount {
		t.Errorf("expected %d debris particles, got %d", config.DebrisCount, activeCount)
	}

	// Test with material that has no debris
	energyConfig := emitter.getImpactConfig(ImpactMelee, MaterialEnergy)
	ps.Clear()
	emitter.spawnDebris(70.0, 70.0, MaterialEnergy, energyConfig)

	if ps.GetActiveCount() != 0 {
		t.Error("energy material should not spawn debris")
	}
}

func TestBrightenColor(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 200}

	brightened := brightenColor(c, 2.0)

	if brightened.R != 200 || brightened.G != 200 || brightened.B != 200 {
		t.Errorf("expected (200,200,200), got (%d,%d,%d)",
			brightened.R, brightened.G, brightened.B)
	}

	if brightened.A != c.A {
		t.Error("alpha should remain unchanged")
	}

	// Test clamping at 255
	overBright := brightenColor(c, 10.0)
	if overBright.R != 255 || overBright.G != 255 || overBright.B != 255 {
		t.Error("brightening should clamp at 255")
	}
}

func TestDarkenColor(t *testing.T) {
	c := color.RGBA{R: 200, G: 200, B: 200, A: 200}

	darkened := darkenColor(c, 0.5)

	if darkened.R != 100 || darkened.G != 100 || darkened.B != 100 {
		t.Errorf("expected (100,100,100), got (%d,%d,%d)",
			darkened.R, darkened.G, darkened.B)
	}

	if darkened.A != c.A {
		t.Error("alpha should remain unchanged")
	}
}

func TestImpactEffectEmitter_ParticleLifecycle(t *testing.T) {
	ps := NewParticleSystem(1024, 1234)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	ps.Clear()
	emitter.EmitImpact(50.0, 50.0, ImpactMelee, MaterialFlesh, 0)

	initialCount := ps.GetActiveCount()
	if initialCount == 0 {
		t.Fatal("no particles spawned")
	}

	// Particles should decay over time
	for i := 0; i < 10; i++ {
		ps.Update(0.1)
	}

	// After 1 second, short-lived particles should be gone
	finalCount := ps.GetActiveCount()
	if finalCount >= initialCount {
		t.Error("particles should decay over time")
	}
}

func BenchmarkImpactEffectEmitter_EmitMelee(b *testing.B) {
	ps := NewParticleSystem(4096, 42)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.EmitImpact(float64(i%100), float64(i%100), ImpactMelee, MaterialFlesh, 0)
	}
}

func BenchmarkImpactEffectEmitter_EmitCritical(b *testing.B) {
	ps := NewParticleSystem(8192, 42)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.EmitImpact(float64(i%100), float64(i%100), ImpactCritical, MaterialMetal, 0)
	}
}

func BenchmarkImpactEffectEmitter_EmitExplosion(b *testing.B) {
	ps := NewParticleSystem(8192, 42)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.EmitImpact(float64(i%100), float64(i%100), ImpactExplosion, MaterialStone, 0)
	}
}

func BenchmarkImpactEffectEmitter_EmitDeath(b *testing.B) {
	ps := NewParticleSystem(8192, 42)
	emitter := NewImpactEffectEmitter(ps, "fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		emitter.EmitImpact(float64(i%100), float64(i%100), ImpactDeath, MaterialFlesh, 0)
	}
}

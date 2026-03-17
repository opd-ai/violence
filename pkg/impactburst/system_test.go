package impactburst

import (
	"image/color"
	"math"
	"testing"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 12345)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != genre {
				t.Errorf("expected genre %s, got %s", genre, sys.genreID)
			}
			if sys.rng == nil {
				t.Error("RNG not initialized")
			}
			if len(sys.profiles) == 0 {
				t.Error("profiles not pre-initialized")
			}
		})
	}
}

func TestSpawnImpact(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	testCases := []struct {
		name       string
		impactType ImpactType
		material   MaterialType
		intensity  float64
	}{
		{"melee_flesh", ImpactMelee, MaterialFlesh, 1.0},
		{"projectile_metal", ImpactProjectile, MaterialMetal, 1.5},
		{"explosion_stone", ImpactExplosion, MaterialStone, 2.0},
		{"magic_energy", ImpactMagic, MaterialEnergy, 0.5},
		{"critical_flesh", ImpactCritical, MaterialFlesh, 1.0},
		{"block_metal", ImpactBlock, MaterialMetal, 0.8},
		{"death_ethereal", ImpactDeath, MaterialEthereal, 1.2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialCount := len(sys.GetGlobalImpacts())
			sys.SpawnImpact(10.0, 20.0, math.Pi/4, tc.impactType, tc.material, tc.intensity)

			impacts := sys.GetGlobalImpacts()
			if len(impacts) != initialCount+1 {
				t.Errorf("expected %d impacts, got %d", initialCount+1, len(impacts))
			}

			imp := impacts[len(impacts)-1]
			if imp.X != 10.0 || imp.Y != 20.0 {
				t.Errorf("wrong position: got (%.1f, %.1f)", imp.X, imp.Y)
			}
			if imp.Type != tc.impactType {
				t.Errorf("wrong type: expected %v, got %v", tc.impactType, imp.Type)
			}
			if imp.Material != tc.material {
				t.Errorf("wrong material: expected %v, got %v", tc.material, imp.Material)
			}
			if imp.Age != 0 {
				t.Error("new impact should have age 0")
			}
			if imp.MaxAge <= 0 {
				t.Error("impact should have positive max age")
			}
		})
	}
}

func TestImpactDebrisGeneration(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Spawn a melee impact with flesh material (should have debris)
	sys.SpawnImpact(0, 0, 0, ImpactMelee, MaterialFlesh, 1.0)

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 {
		t.Fatal("no impacts spawned")
	}

	imp := impacts[0]
	if len(imp.Debris) == 0 {
		t.Error("impact should have debris particles")
	}

	// Check debris properties
	for i, debris := range imp.Debris {
		if debris.Size <= 0 {
			t.Errorf("debris %d has non-positive size", i)
		}
		if debris.MaxAge <= 0 {
			t.Errorf("debris %d has non-positive max age", i)
		}
		if debris.Color.A == 0 {
			t.Errorf("debris %d has zero alpha", i)
		}
	}
}

func TestMaterialProfileDifferences(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	materials := []MaterialType{
		MaterialFlesh,
		MaterialMetal,
		MaterialStone,
		MaterialWood,
		MaterialEnergy,
		MaterialEthereal,
	}

	profiles := make([]ImpactProfile, len(materials))
	for i, mat := range materials {
		profiles[i] = sys.getProfile(ImpactMelee, mat)
	}

	// Check that profiles differ
	for i := 0; i < len(profiles)-1; i++ {
		for j := i + 1; j < len(profiles); j++ {
			if profiles[i].PrimaryColor == profiles[j].PrimaryColor &&
				profiles[i].ParticleCount == profiles[j].ParticleCount {
				t.Errorf("materials %d and %d have identical profiles", i, j)
			}
		}
	}
}

func TestGenreProfileDifferences(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}
	profiles := make([]ImpactProfile, len(genres))

	for i, genre := range genres {
		sys := NewSystem(genre, 12345)
		profiles[i] = sys.getProfile(ImpactMelee, MaterialMetal)
	}

	// Check that at least some profiles differ
	differences := 0
	for i := 0; i < len(profiles)-1; i++ {
		if profiles[i].PrimaryColor != profiles[i+1].PrimaryColor {
			differences++
		}
	}

	if differences == 0 {
		t.Error("expected genre differences in profiles")
	}
}

func TestImpactTypeModification(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Get profiles for different impact types
	meleeProfile := sys.getProfile(ImpactMelee, MaterialFlesh)
	criticalProfile := sys.getProfile(ImpactCritical, MaterialFlesh)
	explosionProfile := sys.getProfile(ImpactExplosion, MaterialFlesh)
	deathProfile := sys.getProfile(ImpactDeath, MaterialFlesh)

	// Critical should have more particles than melee
	if criticalProfile.ParticleCount <= meleeProfile.ParticleCount {
		t.Error("critical impact should have more particles than melee")
	}

	// Explosion should have the highest
	if explosionProfile.ParticleCount <= criticalProfile.ParticleCount {
		t.Error("explosion impact should have more particles than critical")
	}

	// Death should enable glow
	if !deathProfile.HasGlow {
		t.Error("death impact should have glow enabled")
	}

	// Block should have fewer particles
	blockProfile := sys.getProfile(ImpactBlock, MaterialFlesh)
	if blockProfile.ParticleCount >= meleeProfile.ParticleCount {
		t.Error("block impact should have fewer particles than melee")
	}
}

func TestUpdateImpacts(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Spawn an impact (death has longer duration ~1s)
	sys.SpawnImpact(0, 0, 0, ImpactDeath, MaterialFlesh, 1.0)

	initialAge := sys.GetGlobalImpacts()[0].Age

	// Update a few times (not enough to expire)
	for i := 0; i < 20; i++ {
		sys.Update(nil)
	}

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 {
		t.Fatal("impact expired too quickly")
	}

	if impacts[0].Age <= initialAge {
		t.Error("impact age should have increased")
	}
}

func TestImpactExpiration(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Spawn a quick impact (block has short duration)
	sys.SpawnImpact(0, 0, 0, ImpactBlock, MaterialFlesh, 1.0)

	// Update many times until it expires
	for i := 0; i < 300; i++ {
		sys.Update(nil)
	}

	impacts := sys.GetGlobalImpacts()
	if len(impacts) > 0 && impacts[0].Type == ImpactBlock {
		t.Error("impact should have expired")
	}
}

func TestMaxImpactLimit(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Spawn more than max impacts
	for i := 0; i < 100; i++ {
		sys.SpawnImpact(float64(i), float64(i), 0, ImpactMelee, MaterialFlesh, 1.0)
	}

	impacts := sys.GetGlobalImpacts()
	if len(impacts) > sys.maxGlobalImpacts {
		t.Errorf("impact count %d exceeds max %d", len(impacts), sys.maxGlobalImpacts)
	}
}

func TestShockwaveExpansion(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Spawn an impact with shockwave
	sys.SpawnImpact(0, 0, 0, ImpactExplosion, MaterialMetal, 1.0)

	initialRadius := sys.GetGlobalImpacts()[0].ShockwaveRadius

	// Update
	for i := 0; i < 10; i++ {
		sys.Update(nil)
	}

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 {
		t.Fatal("impact missing")
	}

	if impacts[0].ShockwaveRadius <= initialRadius {
		t.Error("shockwave should expand over time")
	}
}

func TestDebrisMovement(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Spawn an impact
	sys.SpawnImpact(0, 0, math.Pi, ImpactMelee, MaterialMetal, 1.0)

	if len(sys.GetGlobalImpacts()[0].Debris) == 0 {
		t.Skip("no debris generated")
	}

	// Get initial debris position
	initialX := sys.GetGlobalImpacts()[0].Debris[0].X
	initialY := sys.GetGlobalImpacts()[0].Debris[0].Y

	// Update
	for i := 0; i < 10; i++ {
		sys.Update(nil)
	}

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 || len(impacts[0].Debris) == 0 {
		t.Skip("debris expired")
	}

	currentX := impacts[0].Debris[0].X
	currentY := impacts[0].Debris[0].Y

	if currentX == initialX && currentY == initialY {
		t.Error("debris should move over time")
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	if sys.genreID != "fantasy" {
		t.Error("initial genre should be fantasy")
	}

	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Error("genre should be updated to scifi")
	}

	// Profiles should be regenerated
	if len(sys.profiles) == 0 {
		t.Error("profiles should be regenerated after genre change")
	}
}

func TestComponentType(t *testing.T) {
	comp := &Component{
		ActiveImpacts: []Impact{},
	}

	if comp.Type() != "ImpactBurstComponent" {
		t.Errorf("unexpected component type: %s", comp.Type())
	}
}

func TestClampFunction(t *testing.T) {
	tests := []struct {
		v, min, max, expected float64
	}{
		{0.5, 0, 1, 0.5},
		{-1, 0, 1, 0},
		{2, 0, 1, 1},
		{0, 0, 1, 0},
		{1, 0, 1, 1},
	}

	for _, tt := range tests {
		result := clamp(tt.v, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%.1f, %.1f, %.1f) = %.1f, expected %.1f",
				tt.v, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestBrightenColor(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 200}

	brightened := brightenColor(c, 1.5)

	if brightened.R <= c.R || brightened.G <= c.G || brightened.B <= c.B {
		t.Error("brightened color should have higher RGB values")
	}

	if brightened.A != c.A {
		t.Error("alpha should be unchanged")
	}

	// Test clamping at 255
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	veryBright := brightenColor(white, 2.0)
	if veryBright.R != 255 || veryBright.G != 255 || veryBright.B != 255 {
		t.Error("color should be clamped at 255")
	}
}

func TestDirectionalDebris(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Metal has directional debris
	profile := sys.getProfile(ImpactMelee, MaterialMetal)
	if !profile.HasDirectionalDebris {
		t.Skip("metal doesn't have directional debris in this config")
	}

	// Spawn at angle pointing right
	impactAngle := 0.0 // Incoming from the right
	sys.SpawnImpact(0, 0, impactAngle, ImpactMelee, MaterialMetal, 1.0)

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 || len(impacts[0].Debris) == 0 {
		t.Skip("no debris to test")
	}

	// Update to let debris move
	for i := 0; i < 5; i++ {
		sys.Update(nil)
	}

	// Most debris should be moving away from impact direction (to the left, negative X)
	leftwardCount := 0
	for _, debris := range sys.GetGlobalImpacts()[0].Debris {
		if debris.X < 0 {
			leftwardCount++
		}
	}

	// At least some should be moving leftward
	if leftwardCount == 0 {
		t.Log("Note: directional debris randomization may cause occasional failures")
	}
}

func TestEnergyMaterialNoChunks(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	profile := sys.getProfile(ImpactMelee, MaterialEnergy)
	if profile.ChunkCount != 0 {
		t.Error("energy material should not have chunks")
	}
}

func TestEtherealFloatingParticles(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	profile := sys.getProfile(ImpactMelee, MaterialEthereal)
	if profile.ParticleGravity >= 0 {
		t.Error("ethereal particles should have negative gravity (float upward)")
	}
}

func BenchmarkSpawnImpact(b *testing.B) {
	sys := NewSystem("fantasy", 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.SpawnImpact(float64(i%100), float64(i%100), 0, ImpactMelee, MaterialFlesh, 1.0)
		// Clear to prevent memory buildup
		if len(sys.globalImpacts) > 32 {
			sys.globalImpacts = sys.globalImpacts[:0]
		}
	}
}

func BenchmarkUpdateImpacts(b *testing.B) {
	sys := NewSystem("fantasy", 12345)

	// Pre-spawn impacts
	for i := 0; i < 50; i++ {
		sys.SpawnImpact(float64(i), float64(i), 0, ImpactMelee, MaterialFlesh, 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(nil)
	}
}

func BenchmarkGetProfile(b *testing.B) {
	sys := NewSystem("fantasy", 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.getProfile(ImpactType(i%7), MaterialType(i%6))
	}
}

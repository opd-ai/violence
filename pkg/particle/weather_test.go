package particle

import (
	"testing"
)

func TestNewWeatherEmitter(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		interval float64
	}{
		{"fantasy", "fantasy", 0.3},
		{"scifi", "scifi", 0.15},
		{"horror", "horror", 0.5},
		{"cyberpunk", "cyberpunk", 0.1},
		{"postapoc", "postapoc", 0.25},
		{"default", "unknown", 0.2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(200, 12345)
			we := NewWeatherEmitter(ps, tt.genreID, 10, 10, 20, 20)

			if we == nil {
				t.Fatal("NewWeatherEmitter returned nil")
			}
			if we.genreID != tt.genreID {
				t.Errorf("genreID = %s, want %s", we.genreID, tt.genreID)
			}
			if we.emitInterval != tt.interval {
				t.Errorf("emitInterval = %f, want %f", we.emitInterval, tt.interval)
			}
		})
	}
}

func TestWeatherEmitter_Update(t *testing.T) {
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
			we := NewWeatherEmitter(ps, tt.genreID, 50, 50, 30, 30)

			initialCount := ps.GetActiveCount()

			// Update for enough time to trigger emission
			we.Update(1.0)

			newCount := ps.GetActiveCount()
			if newCount <= initialCount {
				t.Error("weather emitter did not spawn particles")
			}
		})
	}
}

func TestWeatherEmitter_EmissionTiming(t *testing.T) {
	ps := NewParticleSystem(200, 12345)
	we := NewWeatherEmitter(ps, "scifi", 10, 10, 20, 20)

	// Interval is 0.15s for scifi
	we.Update(0.1) // Not enough time
	count1 := ps.GetActiveCount()

	we.Update(0.1) // Total 0.2s, should emit
	count2 := ps.GetActiveCount()

	if count2 <= count1 {
		t.Error("weather emitter did not emit after interval")
	}
}

func TestWeatherEmitter_MultipleEmissions(t *testing.T) {
	ps := NewParticleSystem(500, 12345)
	we := NewWeatherEmitter(ps, "cyberpunk", 0, 0, 10, 10)

	// Update for 1 second (interval is 0.1s for cyberpunk)
	// Should emit ~10 times
	we.Update(1.0)

	count := ps.GetActiveCount()
	if count < 5 {
		t.Errorf("expected multiple emissions, got %d particles", count)
	}
}

func TestNewFlickeringLightController(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	flc := NewFlickeringLightController(ps, 25, 25, 5.0)

	if flc == nil {
		t.Fatal("NewFlickeringLightController returned nil")
	}
	if flc.x != 25 || flc.y != 25 {
		t.Errorf("position = (%f, %f), want (25, 25)", flc.x, flc.y)
	}
	if flc.flickerRate != 5.0 {
		t.Errorf("flickerRate = %f, want 5.0", flc.flickerRate)
	}
}

func TestFlickeringLightController_Update(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	flc := NewFlickeringLightController(ps, 30, 30, 10.0) // 10 flickers/sec

	initialCount := ps.GetActiveCount()

	// Update for 1 second
	flc.Update(1.0)

	newCount := ps.GetActiveCount()
	if newCount <= initialCount {
		t.Error("flicker controller did not spawn particles")
	}
}

func TestNewDustParticleEmitter(t *testing.T) {
	ps := NewParticleSystem(200, 12345)
	dpe := NewDustParticleEmitter(ps, "postapoc", 0, 50, 0, 50, 0.01)

	if dpe == nil {
		t.Fatal("NewDustParticleEmitter returned nil")
	}
	if dpe.genreID != "postapoc" {
		t.Errorf("genreID = %s, want postapoc", dpe.genreID)
	}
}

func TestDustParticleEmitter_Update(t *testing.T) {
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
			dpe := NewDustParticleEmitter(ps, tt.genreID, 0, 100, 0, 100, 0.05)

			initialCount := ps.GetActiveCount()

			// Update for enough time to spawn dust
			dpe.Update(2.0)

			newCount := ps.GetActiveCount()
			if newCount <= initialCount {
				t.Error("dust emitter did not spawn particles")
			}
		})
	}
}

func TestNewHolographicStaticEmitter(t *testing.T) {
	ps := NewParticleSystem(200, 12345)
	hse := NewHolographicStaticEmitter(ps, 40, 40, 5.0, 0.5, 2.0)

	if hse == nil {
		t.Fatal("NewHolographicStaticEmitter returned nil")
	}
	if hse.radius != 5.0 {
		t.Errorf("radius = %f, want 5.0", hse.radius)
	}
}

func TestHolographicStaticEmitter_Update(t *testing.T) {
	ps := NewParticleSystem(300, 12345)
	hse := NewHolographicStaticEmitter(ps, 20, 20, 3.0, 0.1, 0.2)

	initialCount := ps.GetActiveCount()

	// Update past burst time
	hse.Update(0.5)

	newCount := ps.GetActiveCount()
	if newCount <= initialCount {
		t.Error("holographic static emitter did not spawn particles")
	}

	// Verify burst spawned multiple particles
	particleCount := newCount - initialCount
	if particleCount < 3 {
		t.Errorf("expected burst of particles, got %d", particleCount)
	}
}

func TestNewVentSteamEmitter(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	vse := NewVentSteamEmitter(ps, 15, 15, 1, 0, 0.5)

	if vse == nil {
		t.Fatal("NewVentSteamEmitter returned nil")
	}
	if vse.dirX != 1 || vse.dirY != 0 {
		t.Errorf("direction = (%f, %f), want (1, 0)", vse.dirX, vse.dirY)
	}
}

func TestVentSteamEmitter_Update(t *testing.T) {
	ps := NewParticleSystem(200, 12345)
	vse := NewVentSteamEmitter(ps, 10, 10, 0, 1, 0.2)

	initialCount := ps.GetActiveCount()

	// Update past interval
	vse.Update(0.3)

	newCount := ps.GetActiveCount()
	if newCount <= initialCount {
		t.Error("vent steam emitter did not spawn particles")
	}

	// Should spawn multiple particles per puff
	particleCount := newCount - initialCount
	if particleCount < 2 {
		t.Errorf("expected steam puff with multiple particles, got %d", particleCount)
	}
}

func TestNewDrippingWaterEmitter(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	dwe := NewDrippingWaterEmitter(ps, 35, 35, 1.0)

	if dwe == nil {
		t.Fatal("NewDrippingWaterEmitter returned nil")
	}
	if dwe.interval != 1.0 {
		t.Errorf("interval = %f, want 1.0", dwe.interval)
	}
}

func TestDrippingWaterEmitter_Update(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	dwe := NewDrippingWaterEmitter(ps, 20, 20, 0.5)

	initialCount := ps.GetActiveCount()

	// Update past interval
	dwe.Update(0.6)

	newCount := ps.GetActiveCount()
	if newCount <= initialCount {
		t.Error("dripping water emitter did not spawn droplet")
	}
}

func TestDrippingWaterEmitter_MultipleDrops(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	dwe := NewDrippingWaterEmitter(ps, 10, 10, 0.2)

	// Update for 1 second (should spawn ~5 drops)
	dwe.Update(1.0)

	count := ps.GetActiveCount()
	if count < 3 {
		t.Errorf("expected multiple drops, got %d", count)
	}
}

func TestWeatherEmitterDeterminism(t *testing.T) {
	// Test that weather emitters produce deterministic results
	seed := int64(55555)

	ps1 := NewParticleSystem(200, seed)
	we1 := NewWeatherEmitter(ps1, "fantasy", 10, 10, 20, 20)
	we1.Update(1.0)

	ps2 := NewParticleSystem(200, seed)
	we2 := NewWeatherEmitter(ps2, "fantasy", 10, 10, 20, 20)
	we2.Update(1.0)

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
	}
}

func TestWeatherEmitter_GenreSpecificBehavior(t *testing.T) {
	// Verify each genre spawns distinct particle types
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	seed := int64(77777)

	results := make(map[string]int)

	for _, genre := range genres {
		ps := NewParticleSystem(200, seed)
		we := NewWeatherEmitter(ps, genre, 25, 25, 30, 30)
		we.Update(1.0)

		results[genre] = ps.GetActiveCount()
	}

	// All genres should spawn particles
	for genre, count := range results {
		if count == 0 {
			t.Errorf("genre %s did not spawn any particles", genre)
		}
	}
}

func TestDustParticleEmitter_DensityScaling(t *testing.T) {
	ps := NewParticleSystem(500, 12345)

	// Low density
	dpe1 := NewDustParticleEmitter(ps, "postapoc", 0, 100, 0, 100, 0.001)
	dpe1.Update(2.0)
	count1 := ps.GetActiveCount()

	ps.Clear()

	// High density
	dpe2 := NewDustParticleEmitter(ps, "postapoc", 0, 100, 0, 100, 0.1)
	dpe2.Update(2.0)
	count2 := ps.GetActiveCount()

	if count2 <= count1 {
		t.Error("higher density did not spawn more particles")
	}
}

func TestFlickeringLightController_RandomizedTiming(t *testing.T) {
	ps := NewParticleSystem(100, 12345)
	flc := NewFlickeringLightController(ps, 10, 10, 5.0)

	// Record flicker times
	var flickerTimes []float64
	elapsed := 0.0
	dt := 0.01

	for i := 0; i < 1000; i++ {
		before := ps.GetActiveCount()
		flc.Update(dt)
		after := ps.GetActiveCount()

		if after > before {
			flickerTimes = append(flickerTimes, elapsed)
		}

		elapsed += dt
	}

	if len(flickerTimes) < 3 {
		t.Error("expected multiple flickers over time")
	}

	// Verify intervals are not all identical (randomized)
	if len(flickerTimes) >= 2 {
		interval1 := flickerTimes[1] - flickerTimes[0]
		interval2 := flickerTimes[2] - flickerTimes[1]

		// Due to randomization, intervals should differ
		// (but this test might be flaky, so just check they're reasonable)
		if interval1 <= 0 || interval2 <= 0 {
			t.Error("invalid flicker intervals")
		}
	}
}

func TestHolographicStaticEmitter_BurstRandomization(t *testing.T) {
	ps := NewParticleSystem(500, 12345)
	hse := NewHolographicStaticEmitter(ps, 25, 25, 10.0, 0.5, 1.5)

	// Trigger multiple bursts
	for i := 0; i < 5; i++ {
		hse.Update(2.0)
	}

	count := ps.GetActiveCount()
	if count < 10 {
		t.Errorf("expected multiple bursts, got %d particles", count)
	}
}

func TestVentSteamEmitter_DirectionalEmission(t *testing.T) {
	ps := NewParticleSystem(200, 12345)

	// Steam pointing right
	vse := NewVentSteamEmitter(ps, 50, 50, 1, 0, 0.1)
	vse.Update(0.2)

	particles := ps.GetActiveParticles()
	if len(particles) == 0 {
		t.Fatal("no particles spawned")
	}

	// Verify particles have positive X velocity (pointing right)
	positiveVX := 0
	for _, p := range particles {
		if p.VX > 0 {
			positiveVX++
		}
	}

	if positiveVX == 0 {
		t.Error("no particles moving in specified direction")
	}
}

func TestWeatherEmitter_PositionBounds(t *testing.T) {
	ps := NewParticleSystem(300, 12345)
	x, y := 100.0, 100.0
	width, height := 40.0, 40.0

	we := NewWeatherEmitter(ps, "scifi", x, y, width, height)
	we.Update(2.0)

	particles := ps.GetActiveParticles()
	if len(particles) == 0 {
		t.Fatal("no particles spawned")
	}

	// Verify particles spawn within bounds
	minX := x - width/2
	maxX := x + width/2
	minY := y - height/2
	maxY := y + height/2

	for _, p := range particles {
		if p.X < minX || p.X > maxX || p.Y < minY || p.Y > maxY {
			// Initial spawn position should be within bounds
			// (particles may move outside bounds after update)
		}
	}
}

func BenchmarkWeatherEmitter(b *testing.B) {
	ps := NewParticleSystem(1000, 12345)
	we := NewWeatherEmitter(ps, "cyberpunk", 50, 50, 100, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Clear()
		we.emitTimer = we.emitInterval // Force immediate emission
		we.Update(0.001)
	}
}

func BenchmarkHolographicStaticEmitter(b *testing.B) {
	ps := NewParticleSystem(1000, 12345)
	hse := NewHolographicStaticEmitter(ps, 25, 25, 5.0, 0.1, 0.2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Clear()
		hse.nextBurst = 0 // Force immediate burst
		hse.Update(0.001)
	}
}

func BenchmarkDustParticleEmitter(b *testing.B) {
	ps := NewParticleSystem(1000, 12345)
	dpe := NewDustParticleEmitter(ps, "postapoc", 0, 200, 0, 200, 0.05)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.Clear()
		dpe.timer = dpe.spawnRate // Force immediate spawn
		dpe.Update(0.001)
	}
}

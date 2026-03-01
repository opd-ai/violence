package particle

import (
	"image/color"
	"math"
	"testing"
)

func TestGetVisibleParticles_DistanceCulling(t *testing.T) {
	tests := []struct {
		name        string
		maxDistSq   float64
		particleX   float64
		particleY   float64
		cameraX     float64
		cameraY     float64
		wantVisible bool
	}{
		{
			name:        "particle within range",
			maxDistSq:   400.0,
			particleX:   10.0,
			particleY:   0.0,
			cameraX:     0.0,
			cameraY:     0.0,
			wantVisible: true,
		},
		{
			name:        "particle at exact max distance",
			maxDistSq:   100.0,
			particleX:   10.0,
			particleY:   0.0,
			cameraX:     0.0,
			cameraY:     0.0,
			wantVisible: true,
		},
		{
			name:        "particle beyond range",
			maxDistSq:   50.0,
			particleX:   10.0,
			particleY:   0.0,
			cameraX:     0.0,
			cameraY:     0.0,
			wantVisible: false,
		},
		{
			name:        "particle far outside range",
			maxDistSq:   400.0,
			particleX:   100.0,
			particleY:   100.0,
			cameraX:     0.0,
			cameraY:     0.0,
			wantVisible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(64, 12345)
			ps.Spawn(tt.particleX, tt.particleY, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 0, 255})

			// Camera looking in +X direction
			visible := ps.GetVisibleParticles(tt.cameraX, tt.cameraY, 1.0, 0.0, tt.maxDistSq)

			if tt.wantVisible && len(visible) != 1 {
				t.Errorf("expected 1 visible particle, got %d", len(visible))
			}
			if !tt.wantVisible && len(visible) != 0 {
				t.Errorf("expected 0 visible particles, got %d", len(visible))
			}
		})
	}
}

func TestGetVisibleParticles_FrustumCulling(t *testing.T) {
	tests := []struct {
		name        string
		particleX   float64
		particleY   float64
		cameraX     float64
		cameraY     float64
		cameraDirX  float64
		cameraDirY  float64
		wantVisible bool
		description string
	}{
		{
			name:        "particle in front of camera",
			particleX:   10.0,
			particleY:   0.0,
			cameraX:     0.0,
			cameraY:     0.0,
			cameraDirX:  1.0,
			cameraDirY:  0.0,
			wantVisible: true,
			description: "camera facing +X, particle at +X",
		},
		{
			name:        "particle behind camera",
			particleX:   -10.0,
			particleY:   0.0,
			cameraX:     0.0,
			cameraY:     0.0,
			cameraDirX:  1.0,
			cameraDirY:  0.0,
			wantVisible: false,
			description: "camera facing +X, particle at -X",
		},
		{
			name:        "particle to the side in front",
			particleX:   5.0,
			particleY:   5.0,
			cameraX:     0.0,
			cameraY:     0.0,
			cameraDirX:  1.0,
			cameraDirY:  0.0,
			wantVisible: true,
			description: "camera facing +X, particle at +X+Y (still in front)",
		},
		{
			name:        "particle perpendicular behind",
			particleX:   0.0,
			particleY:   10.0,
			cameraX:     0.0,
			cameraY:     0.0,
			cameraDirX:  1.0,
			cameraDirY:  0.0,
			wantVisible: false,
			description: "camera facing +X, particle at +Y only (perpendicular, dot=0)",
		},
		{
			name:        "camera rotated 90 degrees",
			particleX:   0.0,
			particleY:   10.0,
			cameraX:     0.0,
			cameraY:     0.0,
			cameraDirX:  0.0,
			cameraDirY:  1.0,
			wantVisible: true,
			description: "camera facing +Y, particle at +Y",
		},
		{
			name:        "camera rotated 180 degrees",
			particleX:   -10.0,
			particleY:   0.0,
			cameraX:     0.0,
			cameraY:     0.0,
			cameraDirX:  -1.0,
			cameraDirY:  0.0,
			wantVisible: true,
			description: "camera facing -X, particle at -X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewParticleSystem(64, 12345)
			ps.Spawn(tt.particleX, tt.particleY, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 0, 255})

			visible := ps.GetVisibleParticles(
				tt.cameraX, tt.cameraY,
				tt.cameraDirX, tt.cameraDirY,
				400.0,
			)

			if tt.wantVisible && len(visible) != 1 {
				t.Errorf("%s: expected 1 visible particle, got %d", tt.description, len(visible))
			}
			if !tt.wantVisible && len(visible) != 0 {
				t.Errorf("%s: expected 0 visible particles, got %d", tt.description, len(visible))
			}
		})
	}
}

func TestGetVisibleParticles_MultipleParticles(t *testing.T) {
	ps := NewParticleSystem(128, 54321)

	// Spawn particles in different positions
	ps.Spawn(10.0, 0.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 0, 255})    // In front
	ps.Spawn(-10.0, 0.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{0, 255, 0, 255})   // Behind
	ps.Spawn(5.0, 5.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{0, 0, 255, 255})     // Front-side
	ps.Spawn(100.0, 0.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 255, 0, 255}) // Too far
	ps.Spawn(0.0, 10.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 255, 255})  // Perpendicular (behind)

	// Camera at origin facing +X
	visible := ps.GetVisibleParticles(0.0, 0.0, 1.0, 0.0, 400.0)

	// Should only see particles 1 and 3 (front and front-side)
	if len(visible) != 2 {
		t.Errorf("expected 2 visible particles, got %d", len(visible))
	}

	// Verify the correct particles are visible (check colors)
	visibleColors := make(map[uint8]bool)
	for _, p := range visible {
		visibleColors[p.R] = true
	}

	if !visibleColors[255] { // Red from particle 1
		t.Error("expected red particle (10,0) to be visible")
	}
	if visibleColors[0] && len(visible) == 2 { // Green would mean particle 2 is visible
		// Check if it's the blue particle (0,0,255) which is correct
		hasBlue := false
		for _, p := range visible {
			if p.B == 255 && p.R == 0 && p.G == 0 {
				hasBlue = true
			}
		}
		if !hasBlue {
			t.Error("expected blue particle (5,5) to be visible")
		}
	}
}

func TestGetVisibleParticles_Performance(t *testing.T) {
	// Create a particle system with maximum particles
	ps := NewParticleSystem(1024, 99999)

	// Spawn many particles in various positions
	for i := 0; i < 500; i++ {
		x := float64(i % 50)
		y := float64(i / 50)
		ps.Spawn(x, y, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 0, 255})
	}

	// This should efficiently cull most particles
	visible := ps.GetVisibleParticles(0.0, 0.0, 1.0, 0.0, 100.0)

	// Should have significantly fewer than 500 particles
	if len(visible) > 100 {
		t.Errorf("culling not effective: got %d visible particles from 500 total", len(visible))
	}

	// Verify only particles in front and within range are visible
	for _, p := range visible {
		dx := p.X - 0.0
		dy := p.Y - 0.0
		distSq := dx*dx + dy*dy

		if distSq > 100.0 {
			t.Errorf("particle at (%f,%f) exceeds max distance: distSq=%f", p.X, p.Y, distSq)
		}

		// Dot product with camera direction (1,0)
		dotProduct := dx*1.0 + dy*0.0
		if dotProduct <= 0 {
			t.Errorf("particle at (%f,%f) is behind camera: dot=%f", p.X, p.Y, dotProduct)
		}
	}
}

func TestGetVisibleParticles_EmptySystem(t *testing.T) {
	ps := NewParticleSystem(64, 11111)

	visible := ps.GetVisibleParticles(0.0, 0.0, 1.0, 0.0, 400.0)

	if len(visible) != 0 {
		t.Errorf("expected 0 visible particles from empty system, got %d", len(visible))
	}
}

func TestGetVisibleParticles_AllInactive(t *testing.T) {
	ps := NewParticleSystem(64, 22222)

	// Spawn particles with very short life
	ps.Spawn(10.0, 0.0, 0, 0, 0, 0, 0.01, 2.0, color.RGBA{255, 0, 0, 255})
	ps.Spawn(5.0, 5.0, 0, 0, 0, 0, 0.01, 2.0, color.RGBA{0, 255, 0, 255})

	// Update past their lifetime to deactivate them properly
	ps.Update(0.02)

	visible := ps.GetVisibleParticles(0.0, 0.0, 1.0, 0.0, 400.0)

	if len(visible) != 0 {
		t.Errorf("expected 0 visible particles after all deactivated, got %d", len(visible))
	}
}

func TestGetActiveParticles_UsesActiveIndices(t *testing.T) {
	ps := NewParticleSystem(1024, 33333)

	// Spawn a small number of particles in a large pool
	for i := 0; i < 10; i++ {
		ps.Spawn(float64(i), 0.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 0, 255})
	}

	active := ps.GetActiveParticles()

	if len(active) != 10 {
		t.Errorf("expected 10 active particles, got %d", len(active))
	}

	// Verify we got the correct particles
	for i, p := range active {
		if p.X != float64(i) {
			t.Errorf("particle %d: expected X=%d, got X=%f", i, i, p.X)
		}
	}
}

func TestGetActiveCount_Optimized(t *testing.T) {
	ps := NewParticleSystem(1024, 44444)

	// Initially empty
	if count := ps.GetActiveCount(); count != 0 {
		t.Errorf("expected 0 active particles initially, got %d", count)
	}

	// Spawn some particles
	for i := 0; i < 50; i++ {
		ps.Spawn(float64(i), 0.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 0, 255})
	}

	if count := ps.GetActiveCount(); count != 50 {
		t.Errorf("expected 50 active particles, got %d", count)
	}

	// Update with time to expire some particles
	ps.Update(1.1) // More than 1.0 life

	if count := ps.GetActiveCount(); count != 0 {
		t.Errorf("expected 0 active particles after expiration, got %d", count)
	}
}

func TestGetVisibleParticles_DiagonalCamera(t *testing.T) {
	ps := NewParticleSystem(64, 55555)

	// Camera facing northeast (normalized)
	sqrt2inv := 1.0 / math.Sqrt(2.0)
	dirX := sqrt2inv
	dirY := sqrt2inv

	// Spawn particles in various directions
	ps.Spawn(10.0, 10.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{255, 0, 0, 255})   // Front (northeast)
	ps.Spawn(-10.0, -10.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{0, 255, 0, 255}) // Behind (southwest)
	ps.Spawn(10.0, -10.0, 0, 0, 0, 0, 1.0, 2.0, color.RGBA{0, 0, 255, 255})  // Perpendicular (southeast)

	visible := ps.GetVisibleParticles(0.0, 0.0, dirX, dirY, 400.0)

	// Should only see the northeast particle
	if len(visible) != 1 {
		t.Errorf("expected 1 visible particle, got %d", len(visible))
	}

	if len(visible) > 0 && visible[0].R != 255 {
		t.Error("expected red particle (northeast) to be visible")
	}
}

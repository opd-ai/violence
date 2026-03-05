package particle

import (
	"math"
	"testing"
)

// TestDetermineShape_Logic tests shape determination logic without graphics
func TestDetermineShape_Logic(t *testing.T) {
	rs := &RenderSystem{}

	tests := []struct {
		name     string
		particle Particle
		expected ParticleShape
	}{
		{
			name: "fast spark - high speed with vertical",
			particle: Particle{
				VX: 100, VY: 50, VZ: 10,
				R: 255, G: 200, B: 50, A: 255,
				Size: 2, Active: true,
			},
			expected: ShapeSpark,
		},
		{
			name: "slow smoke - low speed upward",
			particle: Particle{
				VX: 5, VY: 3, VZ: -10,
				R: 100, G: 100, B: 100, A: 150,
				Size: 3, Active: true,
			},
			expected: ShapeSmoke,
		},
		{
			name: "blood diamond - red color",
			particle: Particle{
				VX: 20, VY: 15, VZ: 0,
				R: 180, G: 20, B: 20, A: 255,
				Size: 2, Active: true,
			},
			expected: ShapeDiamond,
		},
		{
			name: "bright star - yellow/white",
			particle: Particle{
				VX: 10, VY: 10, VZ: 0,
				R: 255, G: 220, B: 100, A: 255,
				Size: 2, Active: true,
			},
			expected: ShapeStar,
		},
		{
			name: "medium glow - medium speed",
			particle: Particle{
				VX: 50, VY: 40, VZ: 0,
				R: 255, G: 100, B: 0, A: 255,
				Size: 3, Active: true,
			},
			expected: ShapeGlow,
		},
		{
			name: "fast line - high speed no vertical",
			particle: Particle{
				VX: 70, VY: 60, VZ: 0,
				R: 100, G: 100, B: 100, A: 255,
				Size: 2, Active: true,
			},
			expected: ShapeLine,
		},
		{
			name: "default circle - slow generic",
			particle: Particle{
				VX: 10, VY: 10, VZ: 0,
				R: 100, G: 100, B: 100, A: 255,
				Size: 2, Active: true,
			},
			expected: ShapeCircle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shape := rs.DetermineShape(&tt.particle, "fantasy")
			if shape != tt.expected {
				t.Errorf("DetermineShape() = %v, want %v", shape, tt.expected)
			}
		})
	}
}

func TestRenderSystem_Creation_NoGraphics(t *testing.T) {
	rs := NewRenderSystem()
	if rs == nil {
		t.Fatal("NewRenderSystem returned nil")
	}

	if rs.glowCache == nil {
		t.Error("glowCache not initialized")
	}
	if rs.smokeCache == nil {
		t.Error("smokeCache not initialized")
	}
	if rs.sparkCache == nil {
		t.Error("sparkCache not initialized")
	}
	if rs.maxCacheSize != 20 {
		t.Errorf("maxCacheSize = %d, want 20", rs.maxCacheSize)
	}
}

func TestComponent_NoGraphics(t *testing.T) {
	comp := NewComponent(ShapeGlow, "cyberpunk")

	if comp == nil {
		t.Fatal("NewComponent returned nil")
	}

	if comp.PreferredShape != ShapeGlow {
		t.Errorf("PreferredShape = %v, want %v", comp.PreferredShape, ShapeGlow)
	}

	if comp.GenreID != "cyberpunk" {
		t.Errorf("GenreID = %s, want cyberpunk", comp.GenreID)
	}

	if comp.Type() != "ParticleRenderer" {
		t.Errorf("Type() = %s, want ParticleRenderer", comp.Type())
	}
}

func TestRendererSystem_NoGraphics(t *testing.T) {
	sys := NewRendererSystem()

	if sys == nil {
		t.Fatal("NewRendererSystem returned nil")
	}

	if sys.renderer == nil {
		t.Error("renderer not initialized")
	}

	if sys.Type() != "ParticleRenderer" {
		t.Errorf("Type() = %s, want ParticleRenderer", sys.Type())
	}

	renderer := sys.GetRenderer()
	if renderer == nil {
		t.Error("GetRenderer returned nil")
	}
	if renderer != sys.renderer {
		t.Error("GetRenderer returned different instance")
	}
}

func TestCacheManagement_NoGraphics(t *testing.T) {
	rs := NewRenderSystem()
	rs.maxCacheSize = 5

	// Populate caches
	for i := 0; i < 3; i++ {
		rs.glowCache[i*4] = nil // Just track keys, not actual images
	}

	if len(rs.glowCache) != 3 {
		t.Errorf("glowCache size = %d, want 3", len(rs.glowCache))
	}

	// Clear cache
	rs.ClearCache()

	if len(rs.glowCache) != 0 {
		t.Error("glowCache should be empty after clear")
	}
	if len(rs.smokeCache) != 0 {
		t.Error("smokeCache should be empty after clear")
	}
	if len(rs.sparkCache) != 0 {
		t.Error("sparkCache should be empty after clear")
	}
}

func BenchmarkDetermineShape_NoGraphics(b *testing.B) {
	rs := &RenderSystem{}
	p := Particle{
		VX: 50, VY: 40, VZ: 0,
		R: 255, G: 100, B: 0, A: 255,
		Size: 3, Active: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.DetermineShape(&p, "fantasy")
	}
}

func BenchmarkDetermineShape_AllTypes_NoGraphics(b *testing.B) {
	rs := &RenderSystem{}
	particles := []Particle{
		{VX: 100, VY: 50, VZ: 10, R: 255, G: 200, B: 50, A: 255, Size: 2, Active: true},
		{VX: 5, VY: 3, VZ: -10, R: 100, G: 100, B: 100, A: 150, Size: 3, Active: true},
		{VX: 20, VY: 15, VZ: 0, R: 180, G: 20, B: 20, A: 255, Size: 2, Active: true},
		{VX: 10, VY: 10, VZ: 0, R: 255, G: 220, B: 100, A: 255, Size: 2, Active: true},
		{VX: 50, VY: 40, VZ: 0, R: 255, G: 100, B: 0, A: 255, Size: 3, Active: true},
		{VX: 70, VY: 60, VZ: 0, R: 100, G: 100, B: 100, A: 255, Size: 2, Active: true},
		{VX: 10, VY: 10, VZ: 0, R: 100, G: 100, B: 100, A: 255, Size: 2, Active: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range particles {
			_ = rs.DetermineShape(&particles[j], "fantasy")
		}
	}
}

// Helper to get speed for testing
func getSpeed(vx, vy float64) float64 {
	return math.Sqrt(vx*vx + vy*vy)
}

func TestSpeed_Calculations(t *testing.T) {
	tests := []struct {
		vx, vy   float64
		minSpeed float64
	}{
		{100, 50, 80}, // Fast spark
		{5, 3, 0},     // Slow smoke
		{50, 40, 40},  // Medium glow
		{70, 60, 60},  // Fast line
	}

	for _, tt := range tests {
		speed := getSpeed(tt.vx, tt.vy)
		if speed < tt.minSpeed {
			t.Errorf("Speed(%v, %v) = %v, want >= %v", tt.vx, tt.vy, speed, tt.minSpeed)
		}
	}
}

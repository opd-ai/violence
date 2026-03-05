package particle

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
)

func TestRenderSystem_Creation(t *testing.T) {
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
}

func TestRenderSystem_DetermineShape(t *testing.T) {
	rs := NewRenderSystem()

	tests := []struct {
		name     string
		particle Particle
		expected ParticleShape
	}{
		{
			name: "fast spark",
			particle: Particle{
				VX: 100, VY: 50, VZ: 10,
				R: 255, G: 200, B: 50, A: 255,
				Size: 2,
			},
			expected: ShapeSpark,
		},
		{
			name: "slow smoke",
			particle: Particle{
				VX: 5, VY: 3, VZ: -10,
				R: 100, G: 100, B: 100, A: 150,
				Size: 3,
			},
			expected: ShapeSmoke,
		},
		{
			name: "blood diamond",
			particle: Particle{
				VX: 20, VY: 15, VZ: 0,
				R: 180, G: 20, B: 20, A: 255,
				Size: 2,
			},
			expected: ShapeDiamond,
		},
		{
			name: "bright star",
			particle: Particle{
				VX: 10, VY: 10, VZ: 0,
				R: 255, G: 220, B: 100, A: 255,
				Size: 2,
			},
			expected: ShapeStar,
		},
		{
			name: "medium glow",
			particle: Particle{
				VX: 50, VY: 40, VZ: 0,
				R: 255, G: 100, B: 0, A: 255,
				Size: 3,
			},
			expected: ShapeGlow,
		},
		{
			name: "fast line",
			particle: Particle{
				VX: 70, VY: 60, VZ: 0,
				R: 100, G: 100, B: 100, A: 255,
				Size: 2,
			},
			expected: ShapeLine,
		},
		{
			name: "default circle",
			particle: Particle{
				VX: 10, VY: 10, VZ: 0,
				R: 100, G: 100, B: 100, A: 255,
				Size: 2,
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

func TestRenderSystem_RenderParticle(t *testing.T) {
	rs := NewRenderSystem()
	screen := ebiten.NewImage(100, 100)

	tests := []struct {
		name     string
		particle Particle
		screenX  float32
		screenY  float32
	}{
		{
			name: "active circle particle",
			particle: Particle{
				X: 10, Y: 10, Z: 0,
				VX: 0, VY: 0, VZ: 0,
				Life: 1.0, MaxLife: 1.0,
				R: 255, G: 0, B: 0, A: 255,
				Size: 3, Active: true,
			},
			screenX: 50, screenY: 50,
		},
		{
			name: "fading particle",
			particle: Particle{
				X: 10, Y: 10, Z: 0,
				VX: 20, VY: 10, VZ: 0,
				Life: 0.3, MaxLife: 1.0,
				R: 100, G: 200, B: 100, A: 255,
				Size: 2, Active: true,
			},
			screenX: 60, screenY: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			rs.RenderParticle(screen, &tt.particle, tt.screenX, tt.screenY, "fantasy")
		})
	}
}

func TestRenderSystem_RenderInactiveParticle(t *testing.T) {
	rs := NewRenderSystem()
	screen := ebiten.NewImage(100, 100)

	p := Particle{
		Active: false,
		Life:   0,
	}

	// Should handle inactive particle gracefully
	rs.RenderParticle(screen, &p, 50, 50, "fantasy")
}

func TestRenderSystem_AllShapes(t *testing.T) {
	rs := NewRenderSystem()
	screen := ebiten.NewImage(200, 200)

	shapes := []ParticleShape{
		ShapeCircle, ShapeSquare, ShapeDiamond, ShapeStar,
		ShapeLine, ShapeGlow, ShapeSpark, ShapeSmoke,
	}

	for i, shape := range shapes {
		t.Run(shape.String(), func(t *testing.T) {
			p := Particle{
				X: 10, Y: 10, Z: 0,
				VX: 30, VY: 20, VZ: 0,
				Life: 1.0, MaxLife: 1.0,
				R: 255, G: 100, B: 50, A: 255,
				Size: 4, Active: true,
			}

			// Force specific shape by adjusting velocity
			switch shape {
			case ShapeSpark:
				p.VX, p.VY, p.VZ = 100, 50, 10
			case ShapeSmoke:
				p.VX, p.VY, p.VZ = 5, 3, -10
			case ShapeStar:
				p.R, p.G, p.B = 255, 220, 100
			case ShapeDiamond:
				p.R, p.G, p.B = 180, 20, 20
			case ShapeGlow:
				p.VX, p.VY = 50, 40
			case ShapeLine:
				p.VX, p.VY = 70, 60
			}

			screenX := float32(20 + i*20)
			screenY := float32(20 + i*20)

			// Should not panic for any shape
			rs.RenderParticle(screen, &p, screenX, screenY, "fantasy")
		})
	}
}

func TestRenderSystem_CacheManagement(t *testing.T) {
	rs := NewRenderSystem()

	// Request several glow textures
	for i := 8; i <= 32; i += 4 {
		img := rs.getGlowTexture(i)
		if img == nil {
			t.Errorf("getGlowTexture(%d) returned nil", i)
		}
	}

	// Check cache populated
	if len(rs.glowCache) == 0 {
		t.Error("glowCache should be populated")
	}

	// Clear cache
	rs.ClearCache()

	if len(rs.glowCache) != 0 {
		t.Error("glowCache should be empty after clear")
	}
}

func TestRenderSystem_CacheLimit(t *testing.T) {
	rs := NewRenderSystem()
	rs.maxCacheSize = 5

	// Request more textures than cache limit
	for i := 8; i <= 60; i += 4 {
		rs.getGlowTexture(i)
	}

	// Cache should be limited
	if len(rs.glowCache) > rs.maxCacheSize+1 {
		t.Errorf("glowCache size %d exceeds max %d", len(rs.glowCache), rs.maxCacheSize)
	}
}

func TestComponent_Creation(t *testing.T) {
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

func TestRendererSystem_Creation(t *testing.T) {
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
}

func TestRendererSystem_GetRenderer(t *testing.T) {
	sys := NewRendererSystem()
	renderer := sys.GetRenderer()

	if renderer == nil {
		t.Error("GetRenderer returned nil")
	}

	if renderer != sys.renderer {
		t.Error("GetRenderer returned different instance")
	}
}

func TestRendererSystem_Update(t *testing.T) {
	sys := NewRendererSystem()

	// Create a minimal world
	world := engine.NewWorld()

	// Update should not panic
	sys.Update(world)
}

// String returns shape name for debugging.
func (s ParticleShape) String() string {
	switch s {
	case ShapeCircle:
		return "Circle"
	case ShapeSquare:
		return "Square"
	case ShapeDiamond:
		return "Diamond"
	case ShapeStar:
		return "Star"
	case ShapeLine:
		return "Line"
	case ShapeGlow:
		return "Glow"
	case ShapeSpark:
		return "Spark"
	case ShapeSmoke:
		return "Smoke"
	default:
		return "Unknown"
	}
}

func BenchmarkRenderSystem_DetermineShape(b *testing.B) {
	rs := NewRenderSystem()
	p := Particle{
		VX: 50, VY: 40, VZ: 0,
		R: 255, G: 100, B: 0, A: 255,
		Size: 3, Active: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.DetermineShape(&p, "fantasy")
	}
}

func BenchmarkRenderSystem_RenderParticle(b *testing.B) {
	rs := NewRenderSystem()
	screen := ebiten.NewImage(800, 600)
	p := Particle{
		X: 10, Y: 10, Z: 0,
		VX: 30, VY: 20, VZ: 0,
		Life: 1.0, MaxLife: 1.0,
		R: 255, G: 100, B: 50, A: 255,
		Size: 4, Active: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rs.RenderParticle(screen, &p, 400, 300, "fantasy")
	}
}

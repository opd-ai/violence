package parallax

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem()

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}

	if sys.logger == nil {
		t.Error("System logger not initialized")
	}
}

func TestSystemUpdate(t *testing.T) {
	sys := NewSystem()

	// Update should not panic with nil
	sys.Update(nil)
}

func TestInitializeForWorld(t *testing.T) {
	sys := NewSystem()
	comp := NewComponent("fantasy", "forest", 12345)

	sys.InitializeForWorld(comp, 800, 600)

	if len(comp.Layers) == 0 {
		t.Error("InitializeForWorld did not create layers")
	}

	// Verify all layers have images
	for i, layer := range comp.Layers {
		if layer.Image == nil {
			t.Errorf("Layer %d has nil image after initialization", i)
		}
	}
}

func TestInitializeForWorldNilComponent(t *testing.T) {
	sys := NewSystem()

	// Should not panic with nil component
	sys.InitializeForWorld(nil, 800, 600)
}

func TestInitializeForWorldMultipleGenres(t *testing.T) {
	sys := NewSystem()

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		comp := NewComponent(genre, "default", 12345)
		sys.InitializeForWorld(comp, 800, 600)

		if len(comp.Layers) == 0 {
			t.Errorf("Genre %s produced no layers", genre)
		}
	}
}

func TestRenderNilComponent(t *testing.T) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)

	// Should not panic with disabled component
	comp := NewComponent("fantasy", "forest", 12345)
	comp.Enabled = false
	sys.Render(screen, comp)
}

func TestRenderEmptyLayers(t *testing.T) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)

	// Should not panic with no layers
	comp := NewComponent("fantasy", "forest", 12345)
	sys.Render(screen, comp)
}

func TestRenderWithLayers(t *testing.T) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)

	comp := NewComponent("fantasy", "forest", 12345)
	sys.InitializeForWorld(comp, 800, 600)
	comp.UpdateCamera(100, 50, 800, 600)

	// Should not panic when rendering
	sys.Render(screen, comp)
}

func TestRenderLayerSorting(t *testing.T) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)

	comp := NewComponent("fantasy", "forest", 12345)

	// Add layers in non-sorted order
	layer1 := &Layer{
		Image:       ebiten.NewImage(100, 100),
		ZIndex:      2,
		ScrollSpeed: 0.5,
		Opacity:     0.8,
		Width:       100,
		Height:      100,
		Tint:        [4]float64{1, 1, 1, 1},
	}

	layer2 := &Layer{
		Image:       ebiten.NewImage(100, 100),
		ZIndex:      0,
		ScrollSpeed: 0.2,
		Opacity:     0.6,
		Width:       100,
		Height:      100,
		Tint:        [4]float64{1, 1, 1, 1},
	}

	layer3 := &Layer{
		Image:       ebiten.NewImage(100, 100),
		ZIndex:      1,
		ScrollSpeed: 0.35,
		Opacity:     0.7,
		Width:       100,
		Height:      100,
		Tint:        [4]float64{1, 1, 1, 1},
	}

	comp.AddLayer(layer1)
	comp.AddLayer(layer2)
	comp.AddLayer(layer3)
	comp.UpdateCamera(0, 0, 800, 600)

	// Should render without panic (sorting happens internally)
	sys.Render(screen, comp)
}

func TestRenderWithRepeatingLayer(t *testing.T) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)

	comp := NewComponent("fantasy", "forest", 12345)

	layer := &Layer{
		Image:       ebiten.NewImage(200, 200),
		ScrollSpeed: 0.3,
		RepeatX:     true,
		Opacity:     0.8,
		ZIndex:      0,
		Width:       200,
		Height:      200,
		Tint:        [4]float64{1, 1, 1, 1},
	}

	comp.AddLayer(layer)
	comp.UpdateCamera(500, 300, 800, 600)

	// Should handle tiling without panic
	sys.Render(screen, comp)
}

func TestRenderWithCameraMovement(t *testing.T) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)

	comp := NewComponent("fantasy", "forest", 12345)
	sys.InitializeForWorld(comp, 800, 600)

	// Render at different camera positions
	positions := [][2]float64{
		{0, 0},
		{100, 50},
		{-50, -50},
		{1000, 500},
	}

	for _, pos := range positions {
		comp.UpdateCamera(pos[0], pos[1], 800, 600)
		sys.Render(screen, comp)
	}
}

func TestLayerTintApplication(t *testing.T) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)

	comp := NewComponent("fantasy", "forest", 12345)

	// Test layer with custom tint
	layer := &Layer{
		Image:       ebiten.NewImage(100, 100),
		ScrollSpeed: 0.5,
		Opacity:     0.9,
		ZIndex:      0,
		Width:       100,
		Height:      100,
		Tint:        [4]float64{0.8, 0.9, 1.0, 1.0},
	}

	comp.AddLayer(layer)
	comp.UpdateCamera(0, 0, 800, 600)

	// Should apply tint without panic
	sys.Render(screen, comp)
}

func BenchmarkSystemUpdate(b *testing.B) {
	sys := NewSystem()

	for i := 0; i < b.N; i++ {
		sys.Update(nil)
	}
}

func BenchmarkInitializeForWorld(b *testing.B) {
	sys := NewSystem()

	for i := 0; i < b.N; i++ {
		comp := NewComponent("fantasy", "forest", int64(i))
		sys.InitializeForWorld(comp, 800, 600)
	}
}

func BenchmarkRender(b *testing.B) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)
	comp := NewComponent("fantasy", "forest", 12345)
	sys.InitializeForWorld(comp, 800, 600)
	comp.UpdateCamera(100, 50, 800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Render(screen, comp)
	}
}

func BenchmarkRenderWithCameraUpdate(b *testing.B) {
	sys := NewSystem()
	screen := ebiten.NewImage(800, 600)
	comp := NewComponent("fantasy", "forest", 12345)
	sys.InitializeForWorld(comp, 800, 600)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.UpdateCamera(float64(i%1000), float64(i%600), 800, 600)
		sys.Render(screen, comp)
	}
}

package parallax

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		biomeID string
		seed    int64
	}{
		{"fantasy_forest", "fantasy", "forest", 12345},
		{"scifi_space", "scifi", "space", 67890},
		{"horror_dungeon", "horror", "dungeon", 11111},
		{"cyberpunk_city", "cyberpunk", "city", 22222},
		{"postapoc_wasteland", "postapoc", "wasteland", 33333},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent(tt.genreID, tt.biomeID, tt.seed)

			if comp == nil {
				t.Fatal("NewComponent returned nil")
			}

			if comp.GenreID != tt.genreID {
				t.Errorf("GenreID = %q, want %q", comp.GenreID, tt.genreID)
			}

			if comp.BiomeID != tt.biomeID {
				t.Errorf("BiomeID = %q, want %q", comp.BiomeID, tt.biomeID)
			}

			if comp.Seed != tt.seed {
				t.Errorf("Seed = %d, want %d", comp.Seed, tt.seed)
			}

			if !comp.Enabled {
				t.Error("Component should be enabled by default")
			}

			if comp.Layers == nil {
				t.Error("Layers slice should be initialized")
			}
		})
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent("fantasy", "forest", 12345)

	if comp.Type() != "parallax" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "parallax")
	}
}

func TestAddLayer(t *testing.T) {
	comp := NewComponent("fantasy", "forest", 12345)

	layer1 := &Layer{
		ScrollSpeed: 0.5,
		Opacity:     0.8,
		ZIndex:      1,
	}

	layer2 := &Layer{
		ScrollSpeed: 0.2,
		Opacity:     0.6,
		ZIndex:      0,
	}

	comp.AddLayer(layer1)
	comp.AddLayer(layer2)

	if len(comp.Layers) != 2 {
		t.Errorf("Layer count = %d, want 2", len(comp.Layers))
	}

	if comp.Layers[0] != layer1 {
		t.Error("First layer not stored correctly")
	}

	if comp.Layers[1] != layer2 {
		t.Error("Second layer not stored correctly")
	}
}

func TestUpdateCamera(t *testing.T) {
	comp := NewComponent("fantasy", "forest", 12345)

	comp.UpdateCamera(100.5, 200.7, 800, 600)

	if comp.CameraX != 100.5 {
		t.Errorf("CameraX = %f, want 100.5", comp.CameraX)
	}

	if comp.CameraY != 200.7 {
		t.Errorf("CameraY = %f, want 200.7", comp.CameraY)
	}

	if comp.ViewWidth != 800 {
		t.Errorf("ViewWidth = %d, want 800", comp.ViewWidth)
	}

	if comp.ViewHeight != 600 {
		t.Errorf("ViewHeight = %d, want 600", comp.ViewHeight)
	}
}

func TestLayerDefaults(t *testing.T) {
	layer := &Layer{
		ScrollSpeed: 0.5,
		RepeatX:     true,
		Opacity:     0.8,
		ZIndex:      1,
		Width:       1024,
		Height:      512,
		Tint:        [4]float64{1.0, 1.0, 1.0, 1.0},
	}

	if layer.ScrollSpeed != 0.5 {
		t.Errorf("ScrollSpeed = %f, want 0.5", layer.ScrollSpeed)
	}

	if !layer.RepeatX {
		t.Error("RepeatX should be true")
	}

	if layer.Opacity != 0.8 {
		t.Errorf("Opacity = %f, want 0.8", layer.Opacity)
	}

	if layer.ZIndex != 1 {
		t.Errorf("ZIndex = %d, want 1", layer.ZIndex)
	}

	if layer.Tint[0] != 1.0 || layer.Tint[1] != 1.0 || layer.Tint[2] != 1.0 || layer.Tint[3] != 1.0 {
		t.Errorf("Tint = %v, want [1 1 1 1]", layer.Tint)
	}
}

func TestComponentEnable(t *testing.T) {
	comp := NewComponent("fantasy", "forest", 12345)

	if !comp.Enabled {
		t.Error("Component should be enabled by default")
	}

	comp.Enabled = false

	if comp.Enabled {
		t.Error("Component should be disabled after setting to false")
	}
}

func TestMultipleLayersZOrdering(t *testing.T) {
	comp := NewComponent("fantasy", "forest", 12345)

	// Add layers in non-sorted order
	comp.AddLayer(&Layer{ZIndex: 2, ScrollSpeed: 0.4})
	comp.AddLayer(&Layer{ZIndex: 0, ScrollSpeed: 0.1})
	comp.AddLayer(&Layer{ZIndex: 1, ScrollSpeed: 0.25})

	if len(comp.Layers) != 3 {
		t.Fatalf("Expected 3 layers, got %d", len(comp.Layers))
	}

	// Layers should maintain insertion order
	if comp.Layers[0].ZIndex != 2 {
		t.Errorf("First layer ZIndex = %d, want 2", comp.Layers[0].ZIndex)
	}

	if comp.Layers[1].ZIndex != 0 {
		t.Errorf("Second layer ZIndex = %d, want 0", comp.Layers[1].ZIndex)
	}

	if comp.Layers[2].ZIndex != 1 {
		t.Errorf("Third layer ZIndex = %d, want 1", comp.Layers[2].ZIndex)
	}
}

func BenchmarkNewComponent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewComponent("fantasy", "forest", int64(i))
	}
}

func BenchmarkAddLayer(b *testing.B) {
	comp := NewComponent("fantasy", "forest", 12345)
	layer := &Layer{ScrollSpeed: 0.5, Opacity: 0.8, ZIndex: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.Layers = comp.Layers[:0] // Reset
		comp.AddLayer(layer)
	}
}

func BenchmarkUpdateCamera(b *testing.B) {
	comp := NewComponent("fantasy", "forest", 12345)

	for i := 0; i < b.N; i++ {
		comp.UpdateCamera(float64(i), float64(i*2), 800, 600)
	}
}

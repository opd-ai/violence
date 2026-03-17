package spritebatch

import (
	"image"
	"image/color"
	"testing"
)

func TestNewBatchItem(t *testing.T) {
	item := NewBatchItem(nil, 100, 200)

	if item.DstX != 100 {
		t.Errorf("Expected DstX=100, got %f", item.DstX)
	}
	if item.DstY != 200 {
		t.Errorf("Expected DstY=200, got %f", item.DstY)
	}
	if item.ScaleX != 1.0 || item.ScaleY != 1.0 {
		t.Errorf("Expected scale 1.0, got %f, %f", item.ScaleX, item.ScaleY)
	}
	if item.Alpha != 1.0 {
		t.Errorf("Expected alpha 1.0, got %f", item.Alpha)
	}
	if item.Layer != LayerEntity {
		t.Errorf("Expected LayerEntity, got %v", item.Layer)
	}
}

func TestBatchItemChaining(t *testing.T) {
	item := NewBatchItem(nil, 0, 0).
		WithScale(2.0, 3.0).
		WithRotation(1.5).
		WithAlpha(0.5).
		WithLayer(LayerEffect).
		WithZOrder(10.0).
		WithOrigin(0.0, 1.0).
		WithSourceRect(10, 20, 30, 40)

	if item.ScaleX != 2.0 || item.ScaleY != 3.0 {
		t.Errorf("Scale not set correctly: %f, %f", item.ScaleX, item.ScaleY)
	}
	if item.Rotation != 1.5 {
		t.Errorf("Rotation not set correctly: %f", item.Rotation)
	}
	if item.Alpha != 0.5 {
		t.Errorf("Alpha not set correctly: %f", item.Alpha)
	}
	if item.Layer != LayerEffect {
		t.Errorf("Layer not set correctly: %v", item.Layer)
	}
	if item.ZOrder != 10.0 {
		t.Errorf("ZOrder not set correctly: %f", item.ZOrder)
	}
	if item.OriginX != 0.0 || item.OriginY != 1.0 {
		t.Errorf("Origin not set correctly: %f, %f", item.OriginX, item.OriginY)
	}
	if item.SrcX != 10 || item.SrcY != 20 || item.SrcWidth != 30 || item.SrcHeight != 40 {
		t.Errorf("SourceRect not set correctly")
	}
}

func TestWithColor(t *testing.T) {
	item := NewBatchItem(nil, 0, 0).WithColor(0.5, 0.6, 0.7)

	if item.ColorR != 0.5 {
		t.Errorf("ColorR not set correctly: %f", item.ColorR)
	}
	if item.ColorG != 0.6 {
		t.Errorf("ColorG not set correctly: %f", item.ColorG)
	}
	if item.ColorB != 0.7 {
		t.Errorf("ColorB not set correctly: %f", item.ColorB)
	}
}

func TestLayerConstants(t *testing.T) {
	// Verify layer ordering
	if LayerFloor >= LayerDecal {
		t.Error("LayerFloor should be less than LayerDecal")
	}
	if LayerDecal >= LayerCorpse {
		t.Error("LayerDecal should be less than LayerCorpse")
	}
	if LayerCorpse >= LayerEntity {
		t.Error("LayerCorpse should be less than LayerEntity")
	}
	if LayerEntity >= LayerEffect {
		t.Error("LayerEntity should be less than LayerEffect")
	}
	if LayerEffect >= LayerProjectile {
		t.Error("LayerEffect should be less than LayerProjectile")
	}
	if LayerProjectile >= LayerUI {
		t.Error("LayerProjectile should be less than LayerUI")
	}
}

// createTestImage creates a simple test image without ebiten.
func createTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	return img
}

func TestBatchItemDefaults(t *testing.T) {
	item := NewBatchItem(nil, 50, 75)

	// Test all default values
	tests := []struct {
		name     string
		got      float64
		expected float64
	}{
		{"ScaleX", item.ScaleX, 1.0},
		{"ScaleY", item.ScaleY, 1.0},
		{"Rotation", item.Rotation, 0.0},
		{"OriginX", item.OriginX, 0.5},
		{"OriginY", item.OriginY, 0.5},
		{"ColorR", item.ColorR, 1.0},
		{"ColorG", item.ColorG, 1.0},
		{"ColorB", item.ColorB, 1.0},
		{"Alpha", item.Alpha, 1.0},
		{"ZOrder", item.ZOrder, 0.0},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: expected %f, got %f", tt.name, tt.expected, tt.got)
		}
	}

	// Test integer defaults
	if item.SrcX != 0 {
		t.Errorf("SrcX: expected 0, got %d", item.SrcX)
	}
	if item.SrcY != 0 {
		t.Errorf("SrcY: expected 0, got %d", item.SrcY)
	}
	if item.SrcWidth != 0 {
		t.Errorf("SrcWidth: expected 0, got %d", item.SrcWidth)
	}
	if item.SrcHeight != 0 {
		t.Errorf("SrcHeight: expected 0, got %d", item.SrcHeight)
	}
}

package spritebatch

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// BatchProvider is implemented by systems that can provide a sprite batch.
type BatchProvider interface {
	GetSpriteBatch() *System
}

// QueueIfBatching queues a sprite to the batch if batching is active,
// otherwise draws directly to screen. Returns true if batched.
func QueueIfBatching(batch *System, screen *ebiten.Image, item BatchItem) bool {
	if batch != nil && batch.IsActive() {
		batch.Queue(item)
		return true
	}

	// Fall back to direct drawing if no batch or not active
	if item.Source != nil && screen != nil {
		directDraw(screen, item)
	}
	return false
}

// QueueSpriteIfBatching is a convenience function for simple sprite draws.
func QueueSpriteIfBatching(batch *System, screen, source *ebiten.Image, x, y float64, layer Layer, zOrder float64) bool {
	item := NewBatchItem(source, x, y).
		WithLayer(layer).
		WithZOrder(zOrder)
	return QueueIfBatching(batch, screen, item)
}

// directDraw renders a single batch item directly to screen (fallback path).
func directDraw(screen *ebiten.Image, item BatchItem) {
	if item.Source == nil {
		return
	}

	opts := &ebiten.DrawImageOptions{}

	// Get source bounds
	bounds := item.Source.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	if item.SrcWidth > 0 && item.SrcHeight > 0 {
		srcW = float64(item.SrcWidth)
		srcH = float64(item.SrcHeight)
	}

	// Apply origin offset
	originX := srcW * item.OriginX
	originY := srcH * item.OriginY
	opts.GeoM.Translate(-originX, -originY)

	// Apply scale
	opts.GeoM.Scale(item.ScaleX, item.ScaleY)

	// Apply rotation
	if item.Rotation != 0 {
		opts.GeoM.Rotate(item.Rotation)
	}

	// Apply final position
	opts.GeoM.Translate(item.DstX, item.DstY)

	// Apply color multiply
	opts.ColorScale.Scale(float32(item.ColorR), float32(item.ColorG), float32(item.ColorB), float32(item.Alpha))

	screen.DrawImage(item.Source, opts)
}

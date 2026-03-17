package spritebatch

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Layer defines the rendering order for batched sprites.
type Layer int

const (
	// LayerFloor is for floor decals and details.
	LayerFloor Layer = iota
	// LayerDecal is for ground-level decals.
	LayerDecal
	// LayerCorpse is for corpse sprites.
	LayerCorpse
	// LayerEntity is for entity sprites (enemies, NPCs).
	LayerEntity
	// LayerEffect is for visual effects (particles, trails).
	LayerEffect
	// LayerProjectile is for projectile sprites.
	LayerProjectile
	// LayerUI is for in-world UI elements.
	LayerUI
)

// BatchItem represents a single sprite draw request.
type BatchItem struct {
	// Source is the source image to draw.
	Source *ebiten.Image
	// DstX is the destination X coordinate on screen.
	DstX float64
	// DstY is the destination Y coordinate on screen.
	DstY float64
	// SrcX is the source X coordinate within the source image.
	SrcX int
	// SrcY is the source Y coordinate within the source image.
	SrcY int
	// SrcWidth is the width of the source region.
	SrcWidth int
	// SrcHeight is the height of the source region.
	SrcHeight int
	// ScaleX is the horizontal scale factor.
	ScaleX float64
	// ScaleY is the vertical scale factor.
	ScaleY float64
	// Rotation is the rotation in radians.
	Rotation float64
	// OriginX is the rotation/scale origin X (0-1 relative to size).
	OriginX float64
	// OriginY is the rotation/scale origin Y (0-1 relative to size).
	OriginY float64
	// ColorR is the color multiply red component (0-1).
	ColorR float64
	// ColorG is the color multiply green component (0-1).
	ColorG float64
	// ColorB is the color multiply blue component (0-1).
	ColorB float64
	// Alpha is the opacity (0-1).
	Alpha float64
	// Layer determines rendering order.
	Layer Layer
	// ZOrder is the z-order within the layer (higher = drawn later).
	ZOrder float64
}

// NewBatchItem creates a batch item with default values.
func NewBatchItem(source *ebiten.Image, x, y float64) BatchItem {
	return BatchItem{
		Source:    source,
		DstX:      x,
		DstY:      y,
		SrcX:      0,
		SrcY:      0,
		SrcWidth:  0, // 0 means use full source image
		SrcHeight: 0,
		ScaleX:    1.0,
		ScaleY:    1.0,
		Rotation:  0,
		OriginX:   0.5,
		OriginY:   0.5,
		ColorR:    1.0,
		ColorG:    1.0,
		ColorB:    1.0,
		Alpha:     1.0,
		Layer:     LayerEntity,
		ZOrder:    0,
	}
}

// WithScale sets the scale factors.
func (b BatchItem) WithScale(sx, sy float64) BatchItem {
	b.ScaleX = sx
	b.ScaleY = sy
	return b
}

// WithRotation sets the rotation angle.
func (b BatchItem) WithRotation(rad float64) BatchItem {
	b.Rotation = rad
	return b
}

// WithColor sets the color multiply.
func (bi BatchItem) WithColor(r, g, b float64) BatchItem {
	bi.ColorR = r
	bi.ColorG = g
	bi.ColorB = b
	return bi
}

// WithAlpha sets the opacity.
func (b BatchItem) WithAlpha(a float64) BatchItem {
	b.Alpha = a
	return b
}

// WithLayer sets the rendering layer.
func (b BatchItem) WithLayer(l Layer) BatchItem {
	b.Layer = l
	return b
}

// WithZOrder sets the z-order within the layer.
func (b BatchItem) WithZOrder(z float64) BatchItem {
	b.ZOrder = z
	return b
}

// WithOrigin sets the rotation/scale origin.
func (b BatchItem) WithOrigin(ox, oy float64) BatchItem {
	b.OriginX = ox
	b.OriginY = oy
	return b
}

// WithSourceRect sets the source rectangle for atlas sprites.
func (b BatchItem) WithSourceRect(x, y, w, h int) BatchItem {
	b.SrcX = x
	b.SrcY = y
	b.SrcWidth = w
	b.SrcHeight = h
	return b
}

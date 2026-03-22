package subsurface

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// ApplySSSToEbitenImage applies SSS to an Ebiten image.
// This creates a new image with the effect applied.
func (s *System) ApplySSSToEbitenImage(src *ebiten.Image, mat Material, intensity float64) *ebiten.Image {
	if src == nil || intensity <= 0 {
		return src
	}

	bounds := src.Bounds()
	rgba := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.At(x, y)
			r, g, b, a := c.RGBA()
			rgba.SetRGBA(x, y, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	s.ApplySSS(rgba, mat, intensity)

	result := ebiten.NewImageFromImage(rgba)
	return result
}

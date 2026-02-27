// Package render provides the rendering pipeline for drawing frames.
package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/raycaster"
)

// Renderer manages the rendering pipeline.
type Renderer struct {
	Width       int
	Height      int
	framebuffer []byte
	raycaster   *raycaster.Raycaster
	palette     map[int]color.RGBA
	genreID     string
}

// NewRenderer creates a renderer with the given internal resolution.
func NewRenderer(width, height int, rc *raycaster.Raycaster) *Renderer {
	return &Renderer{
		Width:       width,
		Height:      height,
		framebuffer: make([]byte, width*height*4),
		raycaster:   rc,
		palette:     getDefaultPalette(),
		genreID:     "fantasy",
	}
}

// Render draws a frame to the given screen image.
// Calls raycaster, writes column data to framebuffer, blits to screen.
func (r *Renderer) Render(screen *ebiten.Image, posX, posY, dirX, dirY, pitch float64) {
	hits := r.raycaster.CastRays(posX, posY, dirX, dirY)

	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			idx := (y*r.Width + x) * 4
			var c color.RGBA

			if y < r.Height/2 {
				c = r.renderCeiling(x, y, posX, posY, dirX, dirY, pitch)
			} else if y > r.Height/2 {
				c = r.renderFloor(x, y, posX, posY, dirX, dirY, pitch)
			} else {
				c = r.palette[0]
			}

			if x < len(hits) {
				wallColor := r.renderWall(x, y, hits[x])
				if wallColor.A > 0 {
					c = wallColor
				}
			}

			r.framebuffer[idx] = c.R
			r.framebuffer[idx+1] = c.G
			r.framebuffer[idx+2] = c.B
			r.framebuffer[idx+3] = c.A
		}
	}

	img := ebiten.NewImageFromImageWithOptions(
		&frameImage{data: r.framebuffer, width: r.Width, height: r.Height},
		&ebiten.NewImageFromImageOptions{Unmanaged: true},
	)
	screen.DrawImage(img, nil)
}

// renderWall computes wall color for a given column and row.
func (r *Renderer) renderWall(x, y int, hit raycaster.RayHit) color.RGBA {
	if hit.Distance >= 1e30 || hit.WallType == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	lineHeight := int(float64(r.Height) / hit.Distance)
	drawStart := -lineHeight/2 + r.Height/2
	drawEnd := lineHeight/2 + r.Height/2

	if y < drawStart || y > drawEnd {
		return color.RGBA{0, 0, 0, 0}
	}

	baseColor := r.palette[hit.WallType]
	if hit.Side == 1 {
		baseColor.R = baseColor.R / 2
		baseColor.G = baseColor.G / 2
		baseColor.B = baseColor.B / 2
	}

	foggedColor := r.raycaster.ApplyFog(
		[3]float64{
			float64(baseColor.R) / 255.0,
			float64(baseColor.G) / 255.0,
			float64(baseColor.B) / 255.0,
		},
		hit.Distance,
	)

	return color.RGBA{
		R: uint8(foggedColor[0] * 255),
		G: uint8(foggedColor[1] * 255),
		B: uint8(foggedColor[2] * 255),
		A: 255,
	}
}

// renderFloor computes floor color for a given pixel.
func (r *Renderer) renderFloor(x, y int, posX, posY, dirX, dirY, pitch float64) color.RGBA {
	pixels := r.raycaster.CastFloorCeiling(y, posX, posY, dirX, dirY, pitch)
	if x >= len(pixels) {
		return r.palette[0]
	}

	baseColor := r.palette[2]
	foggedColor := r.raycaster.ApplyFog(
		[3]float64{
			float64(baseColor.R) / 255.0,
			float64(baseColor.G) / 255.0,
			float64(baseColor.B) / 255.0,
		},
		pixels[x].Distance,
	)

	return color.RGBA{
		R: uint8(foggedColor[0] * 255),
		G: uint8(foggedColor[1] * 255),
		B: uint8(foggedColor[2] * 255),
		A: 255,
	}
}

// renderCeiling computes ceiling color for a given pixel.
func (r *Renderer) renderCeiling(x, y int, posX, posY, dirX, dirY, pitch float64) color.RGBA {
	pixels := r.raycaster.CastFloorCeiling(r.Height-1-y, posX, posY, dirX, dirY, pitch)
	if x >= len(pixels) {
		return r.palette[0]
	}

	baseColor := r.palette[3]
	foggedColor := r.raycaster.ApplyFog(
		[3]float64{
			float64(baseColor.R) / 255.0,
			float64(baseColor.G) / 255.0,
			float64(baseColor.B) / 255.0,
		},
		pixels[x].Distance,
	)

	return color.RGBA{
		R: uint8(foggedColor[0] * 255),
		G: uint8(foggedColor[1] * 255),
		B: uint8(foggedColor[2] * 255),
		A: 255,
	}
}

// SetGenre configures the renderer for a genre.
func (r *Renderer) SetGenre(genreID string) {
	r.genreID = genreID
	r.palette = getPaletteForGenre(genreID)
	r.raycaster.SetGenre(genreID)
}

// getDefaultPalette returns the default color palette.
func getDefaultPalette() map[int]color.RGBA {
	return getPaletteForGenre("fantasy")
}

// getPaletteForGenre returns genre-specific color palette.
func getPaletteForGenre(genreID string) map[int]color.RGBA {
	switch genreID {
	case "fantasy":
		return map[int]color.RGBA{
			0: {20, 15, 30, 255},   // Sky/background
			1: {100, 80, 60, 255},  // Stone wall
			2: {40, 35, 30, 255},   // Floor
			3: {30, 25, 35, 255},   // Ceiling
			4: {120, 100, 80, 255}, // Alternate wall
		}
	case "scifi":
		return map[int]color.RGBA{
			0: {10, 15, 25, 255},    // Sky/background
			1: {80, 90, 100, 255},   // Metal hull
			2: {30, 35, 40, 255},    // Floor
			3: {25, 30, 35, 255},    // Ceiling
			4: {100, 110, 120, 255}, // Alternate wall
		}
	case "horror":
		return map[int]color.RGBA{
			0: {15, 5, 5, 255},    // Sky/background
			1: {80, 60, 50, 255},  // Decayed plaster
			2: {30, 20, 15, 255},  // Floor
			3: {25, 15, 10, 255},  // Ceiling
			4: {100, 70, 60, 255}, // Alternate wall
		}
	case "cyberpunk":
		return map[int]color.RGBA{
			0: {20, 10, 25, 255},   // Sky/background
			1: {90, 70, 100, 255},  // Neon-lit concrete
			2: {35, 30, 40, 255},   // Floor
			3: {30, 25, 35, 255},   // Ceiling
			4: {110, 80, 120, 255}, // Alternate wall
		}
	case "postapoc":
		return map[int]color.RGBA{
			0: {25, 20, 15, 255},  // Sky/background
			1: {100, 80, 60, 255}, // Rusted metal
			2: {40, 30, 25, 255},  // Floor
			3: {35, 25, 20, 255},  // Ceiling
			4: {120, 90, 70, 255}, // Alternate wall
		}
	default:
		return map[int]color.RGBA{
			0: {0, 0, 0, 255},
			1: {128, 128, 128, 255},
			2: {64, 64, 64, 255},
			3: {96, 96, 96, 255},
			4: {160, 160, 160, 255},
		}
	}
}

// frameImage implements image.Image for framebuffer blitting.
type frameImage struct {
	data   []byte
	width  int
	height int
}

func (f *frameImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (f *frameImage) Bounds() image.Rectangle {
	return image.Rect(0, 0, f.width, f.height)
}

func (f *frameImage) At(x, y int) color.Color {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return color.RGBA{0, 0, 0, 255}
	}
	idx := (y*f.width + x) * 4
	return color.RGBA{
		R: f.data[idx],
		G: f.data[idx+1],
		B: f.data[idx+2],
		A: f.data[idx+3],
	}
}

// Package outline provides sprite silhouette rendering for visual clarity.
package outline

import (
	"image"
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/pool"
	"github.com/sirupsen/logrus"
)

// System renders outlines around sprites for improved visual hierarchy.
type System struct {
	genreID       string
	outlineCache  map[cacheKey]*ebiten.Image
	playerColor   color.RGBA
	enemyColor    color.RGBA
	allyColor     color.RGBA
	neutralColor  color.RGBA
	interactColor color.RGBA
}

type cacheKey struct {
	imageID   uint64
	color     color.RGBA
	thickness int
	glow      bool
}

// NewSystem creates an outline rendering system.
func NewSystem(genreID string) *System {
	s := &System{
		genreID:      genreID,
		outlineCache: make(map[cacheKey]*ebiten.Image),
	}
	s.setGenreColors()
	return s
}

// setGenreColors configures outline colors based on genre aesthetics.
func (s *System) setGenreColors() {
	switch s.genreID {
	case "scifi":
		s.playerColor = color.RGBA{R: 100, G: 200, B: 255, A: 255}
		s.enemyColor = color.RGBA{R: 255, G: 80, B: 80, A: 255}
		s.allyColor = color.RGBA{R: 80, G: 255, B: 120, A: 255}
		s.neutralColor = color.RGBA{R: 200, G: 200, B: 200, A: 255}
		s.interactColor = color.RGBA{R: 255, G: 255, B: 100, A: 255}
	case "horror":
		s.playerColor = color.RGBA{R: 180, G: 180, B: 255, A: 255}
		s.enemyColor = color.RGBA{R: 200, G: 50, B: 50, A: 255}
		s.allyColor = color.RGBA{R: 150, G: 200, B: 150, A: 255}
		s.neutralColor = color.RGBA{R: 150, G: 150, B: 150, A: 255}
		s.interactColor = color.RGBA{R: 255, G: 200, B: 100, A: 255}
	case "cyberpunk":
		s.playerColor = color.RGBA{R: 0, G: 255, B: 255, A: 255}
		s.enemyColor = color.RGBA{R: 255, G: 0, B: 128, A: 255}
		s.allyColor = color.RGBA{R: 100, G: 255, B: 100, A: 255}
		s.neutralColor = color.RGBA{R: 180, G: 180, B: 200, A: 255}
		s.interactColor = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	case "postapoc":
		s.playerColor = color.RGBA{R: 200, G: 220, B: 255, A: 255}
		s.enemyColor = color.RGBA{R: 220, G: 80, B: 60, A: 255}
		s.allyColor = color.RGBA{R: 120, G: 200, B: 120, A: 255}
		s.neutralColor = color.RGBA{R: 180, G: 180, B: 160, A: 255}
		s.interactColor = color.RGBA{R: 240, G: 200, B: 80, A: 255}
	default: // fantasy
		s.playerColor = color.RGBA{R: 120, G: 180, B: 255, A: 255}
		s.enemyColor = color.RGBA{R: 255, G: 100, B: 100, A: 255}
		s.allyColor = color.RGBA{R: 100, G: 255, B: 100, A: 255}
		s.neutralColor = color.RGBA{R: 200, G: 200, B: 200, A: 255}
		s.interactColor = color.RGBA{R: 255, G: 220, B: 100, A: 255}
	}
}

// SetGenre updates the genre and reconfigures colors.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.setGenreColors()
	s.outlineCache = make(map[cacheKey]*ebiten.Image)
}

// Update processes entities and updates outline rendering state.
func (s *System) Update(w *engine.World) {
	outlineType := reflect.TypeOf(&Component{})
	entities := w.Query(outlineType)

	logrus.WithFields(logrus.Fields{
		"system_name":   "OutlineSystem",
		"entity_count":  len(entities),
		"cache_entries": len(s.outlineCache),
	}).Trace("Processing outline rendering")

	if len(s.outlineCache) > 200 {
		s.outlineCache = make(map[cacheKey]*ebiten.Image)
	}
}

// GenerateOutline creates an outlined version of a sprite.
func (s *System) GenerateOutline(src *ebiten.Image, outlineColor color.RGBA, thickness int, glow bool) *ebiten.Image {
	if src == nil {
		return nil
	}

	// Use image dimensions and color as cache key since we can't get pointer
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	imageID := uint64(w)<<32 | uint64(h)<<16 | uint64(outlineColor.R)<<8 | uint64(outlineColor.G)

	key := cacheKey{
		imageID:   imageID,
		color:     outlineColor,
		thickness: thickness,
		glow:      glow,
	}

	if cached, found := s.outlineCache[key]; found {
		return cached
	}

	rgba := pool.GlobalPools.Images.Get(w, h)
	defer pool.GlobalPools.Images.Put(rgba)

	src.ReadPixels(rgba.Pix)

	outlineRGBA := pool.GlobalPools.Images.Get(w, h)

	s.generateOutlinePixels(rgba, outlineRGBA, outlineColor, thickness, glow)

	result := ebiten.NewImageFromImage(outlineRGBA)
	pool.GlobalPools.Images.Put(outlineRGBA)

	s.outlineCache[key] = result
	return result
}

// generateOutlinePixels creates outline pixels around the sprite silhouette.
func (s *System) generateOutlinePixels(src, dst *image.RGBA, outlineColor color.RGBA, thickness int, glow bool) {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcIdx := (y*src.Stride + x*4)
			srcAlpha := src.Pix[srcIdx+3]

			if srcAlpha > 0 {
				dst.SetRGBA(x, y, src.RGBAAt(x, y))
				continue
			}

			hasOpaqueNeighbor := false
			minDist := float64(thickness + 1)

			for dy := -thickness; dy <= thickness; dy++ {
				for dx := -thickness; dx <= thickness; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}

					nx, ny := x+dx, y+dy
					if nx < 0 || nx >= w || ny < 0 || ny >= h {
						continue
					}

					nIdx := ny*src.Stride + nx*4
					nAlpha := src.Pix[nIdx+3]

					if nAlpha > 128 {
						dist := math.Sqrt(float64(dx*dx + dy*dy))
						if dist <= float64(thickness) {
							hasOpaqueNeighbor = true
							if dist < minDist {
								minDist = dist
							}
						}
					}
				}
			}

			if hasOpaqueNeighbor {
				alpha := uint8(255)
				if glow {
					falloff := 1.0 - (minDist / float64(thickness))
					alpha = uint8(float64(outlineColor.A) * math.Pow(falloff, 1.5))
				} else {
					falloff := 1.0 - (minDist / float64(thickness))
					alpha = uint8(float64(outlineColor.A) * falloff)
				}

				dst.SetRGBA(x, y, color.RGBA{
					R: outlineColor.R,
					G: outlineColor.G,
					B: outlineColor.B,
					A: alpha,
				})
			}
		}
	}
}

// GetPlayerColor returns the player outline color.
func (s *System) GetPlayerColor() color.RGBA {
	return s.playerColor
}

// GetEnemyColor returns the enemy outline color.
func (s *System) GetEnemyColor() color.RGBA {
	return s.enemyColor
}

// GetAllyColor returns the ally outline color.
func (s *System) GetAllyColor() color.RGBA {
	return s.allyColor
}

// GetNeutralColor returns the neutral entity outline color.
func (s *System) GetNeutralColor() color.RGBA {
	return s.neutralColor
}

// GetInteractColor returns the interactable object outline color.
func (s *System) GetInteractColor() color.RGBA {
	return s.interactColor
}

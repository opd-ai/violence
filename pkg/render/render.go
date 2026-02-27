// Package render provides the rendering pipeline for drawing frames.
package render

import "github.com/hajimehoshi/ebiten/v2"

// Renderer manages the rendering pipeline.
type Renderer struct {
	Width  int
	Height int
}

// NewRenderer creates a renderer with the given internal resolution.
func NewRenderer(width, height int) *Renderer {
	return &Renderer{Width: width, Height: height}
}

// Render draws a frame to the given screen image.
func (r *Renderer) Render(screen *ebiten.Image) {}

// SetGenre configures the renderer for a genre.
func SetGenre(genreID string) {}

// Package raycaster implements the core raycasting engine.
package raycaster

// Raycaster performs raycasting against a 2D map.
type Raycaster struct {
	FOV    float64
	Width  int
	Height int
}

// NewRaycaster creates a raycaster with the given field of view and resolution.
func NewRaycaster(fov float64, width, height int) *Raycaster {
	return &Raycaster{FOV: fov, Width: width, Height: height}
}

// CastRays casts all rays for a single frame and returns wall distances.
func (r *Raycaster) CastRays(posX, posY, dirX, dirY float64) []float64 {
	return make([]float64, r.Width)
}

// SetGenre configures raycaster parameters for a genre.
func SetGenre(genreID string) {}

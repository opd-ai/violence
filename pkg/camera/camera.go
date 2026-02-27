// Package camera manages the first-person camera.
package camera

// Camera represents the player's viewpoint.
type Camera struct {
	X, Y    float64
	DirX    float64
	DirY    float64
	FOV     float64
	Pitch   float64
	HeadBob float64
}

// NewCamera creates a camera with default settings.
func NewCamera(fov float64) *Camera {
	return &Camera{FOV: fov, DirX: 1}
}

// Update advances the camera state by one tick.
func (c *Camera) Update() {}

// SetGenre configures camera behavior for a genre.
func SetGenre(genreID string) {}

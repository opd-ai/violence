// Package lighting provides dynamic lighting calculations.
package lighting

// Light represents a point light source.
type Light struct {
	X, Y      float64
	Radius    float64
	Intensity float64
	R, G, B   float64
}

// Sector represents a lit area of the level.
type Sector struct {
	Lights []Light
}

// Calculate computes lighting for a sector.
func (s *Sector) Calculate() {}

// SetGenre configures lighting presets for a genre.
func SetGenre(genreID string) {}

// Package lighting provides dynamic lighting calculations.
package lighting

// Light represents a point light source.
type Light struct {
	X, Y      float64
	Radius    float64
	Intensity float64
	R, G, B   float64
}

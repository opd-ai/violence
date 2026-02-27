// Package props manages decorative props and level objects.
package props

// Prop represents a decorative prop in the game world.
type Prop struct {
	ID   string
	X, Y float64
	Name string
}

// Place adds a prop at the given position.
func Place(name string, x, y float64) *Prop {
	return &Prop{Name: name, X: x, Y: y}
}

// SetGenre configures prop sets for a genre.
func SetGenre(genreID string) {}

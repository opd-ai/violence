// Package class defines player character classes.
package class

const (
	Grunt  = "grunt"
	Medic  = "medic"
	Demo   = "demo"
	Mystic = "mystic"
)

// Class describes a character class and its base stats.
type Class struct {
	ID     string
	Name   string
	Health float64
	Speed  float64
}

// GetClass returns the class definition for the given ID.
func GetClass(id string) Class {
	return Class{ID: id}
}

// SetGenre configures available classes for a genre.
func SetGenre(genreID string) {}

// Package door implements keycard-locked doors.
package door

// Door represents a door in the game world.
type Door struct {
	ID       string
	Locked   bool
	Required string
}

// Keycard represents a keycard that can open doors.
type Keycard struct {
	Color string
}

// TryOpen attempts to open a door with the given keycard.
func TryOpen(d *Door, k *Keycard) bool {
	if !d.Locked {
		return true
	}
	return d.Required == k.Color
}

// SetGenre configures door/key themes for a genre.
func SetGenre(genreID string) {}

// Package mod provides mod loading and management.
package mod

// Mod represents a loaded game modification.
type Mod struct {
	Name    string
	Version string
	Path    string
}

// Loader manages loading and listing of mods.
type Loader struct {
	mods []Mod
}

// NewLoader creates a new mod loader.
func NewLoader() *Loader {
	return &Loader{}
}

// LoadMod loads a mod from the given path.
func (l *Loader) LoadMod(path string) error {
	return nil
}

// ListMods returns all loaded mods.
func (l *Loader) ListMods() []Mod {
	return l.mods
}

// SetGenre configures the mod system for a genre.
func SetGenre(genreID string) {}

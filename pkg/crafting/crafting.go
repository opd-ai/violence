// Package crafting provides item crafting and recipe management.
package crafting

// Recipe defines a crafting recipe.
type Recipe struct {
	Name    string
	Inputs  map[string]int
	Output  string
}

// Craft attempts to craft an item using the given recipe.
func Craft(r Recipe, available map[string]int) (string, bool) {
	for item, qty := range r.Inputs {
		if available[item] < qty {
			return "", false
		}
	}
	return r.Output, true
}

// SetGenre configures crafting recipes for a genre.
func SetGenre(genreID string) {}

// Package crafting provides scrap-to-ammo crafting and recipe management.
package crafting

// Recipe defines a crafting recipe.
type Recipe struct {
	ID        string
	Name      string
	Inputs    map[string]int // itemID -> quantity
	OutputID  string
	OutputQty int
}

var (
	genreRecipes = map[string][]Recipe{}
	currentGenre = "fantasy"
)

// Craft attempts to craft an item using the given recipe.
// Returns output item ID, output quantity, and success flag.
// Does NOT consume inputs - caller must handle that.
func Craft(r Recipe, available map[string]int) (string, int, bool) {
	if r.Inputs == nil || available == nil {
		return "", 0, false
	}
	for item, qty := range r.Inputs {
		if available[item] < qty {
			return "", 0, false
		}
	}
	return r.OutputID, r.OutputQty, true
}

// GetRecipes returns all recipes for current genre.
func GetRecipes() []Recipe {
	if recipes, ok := genreRecipes[currentGenre]; ok {
		return recipes
	}
	return getDefaultRecipes()
}

// GetRecipe returns recipe by ID for current genre.
func GetRecipe(id string) *Recipe {
	recipes := GetRecipes()
	for i := range recipes {
		if recipes[i].ID == id {
			return &recipes[i]
		}
	}
	return nil
}

// SetGenre configures crafting recipes for a genre.
func SetGenre(genreID string) {
	currentGenre = genreID
	if _, ok := genreRecipes[genreID]; !ok {
		genreRecipes[genreID] = getGenreRecipes(genreID)
	}
}

func getDefaultRecipes() []Recipe {
	return []Recipe{
		{ID: "bullets", Name: "Craft Bullets", Inputs: map[string]int{"scrap": 5}, OutputID: "bullets", OutputQty: 10},
		{ID: "shells", Name: "Craft Shells", Inputs: map[string]int{"scrap": 8}, OutputID: "shells", OutputQty: 5},
		{ID: "cells", Name: "Craft Energy Cells", Inputs: map[string]int{"scrap": 10}, OutputID: "cells", OutputQty: 10},
		{ID: "rockets", Name: "Craft Rockets", Inputs: map[string]int{"scrap": 15}, OutputID: "rockets", OutputQty: 2},
		{ID: "medkit", Name: "Craft Medkit", Inputs: map[string]int{"scrap": 12}, OutputID: "medkit", OutputQty: 1},
	}
}

func getGenreRecipes(genreID string) []Recipe {
	switch genreID {
	case "fantasy":
		return []Recipe{
			{ID: "arrows", Name: "Craft Arrows", Inputs: map[string]int{"bone_chips": 5}, OutputID: "arrows", OutputQty: 10},
			{ID: "bolts", Name: "Craft Bolts", Inputs: map[string]int{"bone_chips": 8}, OutputID: "bolts", OutputQty: 5},
			{ID: "mana", Name: "Craft Mana Crystals", Inputs: map[string]int{"bone_chips": 10}, OutputID: "mana", OutputQty: 10},
			{ID: "explosives", Name: "Craft Explosives", Inputs: map[string]int{"bone_chips": 15}, OutputID: "explosives", OutputQty: 2},
			{ID: "potion", Name: "Brew Potion", Inputs: map[string]int{"bone_chips": 12}, OutputID: "potion", OutputQty: 1},
		}
	case "scifi":
		return []Recipe{
			{ID: "bullets", Name: "Fabricate Bullets", Inputs: map[string]int{"circuit_boards": 5}, OutputID: "bullets", OutputQty: 10},
			{ID: "shells", Name: "Fabricate Shells", Inputs: map[string]int{"circuit_boards": 8}, OutputID: "shells", OutputQty: 5},
			{ID: "cells", Name: "Fabricate Energy Cells", Inputs: map[string]int{"circuit_boards": 10}, OutputID: "cells", OutputQty: 10},
			{ID: "rockets", Name: "Fabricate Rockets", Inputs: map[string]int{"circuit_boards": 15}, OutputID: "rockets", OutputQty: 2},
			{ID: "medkit", Name: "Synthesize Medkit", Inputs: map[string]int{"circuit_boards": 12}, OutputID: "medkit", OutputQty: 1},
		}
	case "horror":
		return []Recipe{
			{ID: "bullets", Name: "Assemble Bullets", Inputs: map[string]int{"flesh": 5}, OutputID: "bullets", OutputQty: 10},
			{ID: "shells", Name: "Assemble Shells", Inputs: map[string]int{"flesh": 8}, OutputID: "shells", OutputQty: 5},
			{ID: "cells", Name: "Condense Souls", Inputs: map[string]int{"flesh": 10}, OutputID: "cells", OutputQty: 10},
			{ID: "rockets", Name: "Bind Explosives", Inputs: map[string]int{"flesh": 15}, OutputID: "rockets", OutputQty: 2},
			{ID: "medkit", Name: "Stitch Medkit", Inputs: map[string]int{"flesh": 12}, OutputID: "medkit", OutputQty: 1},
		}
	case "cyberpunk":
		return []Recipe{
			{ID: "bullets", Name: "Print Bullets", Inputs: map[string]int{"data_shards": 5}, OutputID: "bullets", OutputQty: 10},
			{ID: "shells", Name: "Print Shells", Inputs: map[string]int{"data_shards": 8}, OutputID: "shells", OutputQty: 5},
			{ID: "cells", Name: "Charge Cells", Inputs: map[string]int{"data_shards": 10}, OutputID: "cells", OutputQty: 10},
			{ID: "rockets", Name: "Assemble Rockets", Inputs: map[string]int{"data_shards": 15}, OutputID: "rockets", OutputQty: 2},
			{ID: "medkit", Name: "Compile Medkit", Inputs: map[string]int{"data_shards": 12}, OutputID: "medkit", OutputQty: 1},
		}
	case "postapoc":
		return []Recipe{
			{ID: "bullets", Name: "Scavenge Bullets", Inputs: map[string]int{"salvage": 5}, OutputID: "bullets", OutputQty: 10},
			{ID: "shells", Name: "Scavenge Shells", Inputs: map[string]int{"salvage": 8}, OutputID: "shells", OutputQty: 5},
			{ID: "cells", Name: "Salvage Cells", Inputs: map[string]int{"salvage": 10}, OutputID: "cells", OutputQty: 10},
			{ID: "rockets", Name: "Jury-rig Rockets", Inputs: map[string]int{"salvage": 15}, OutputID: "rockets", OutputQty: 2},
			{ID: "medkit", Name: "Improvise Medkit", Inputs: map[string]int{"salvage": 12}, OutputID: "medkit", OutputQty: 1},
		}
	default:
		return getDefaultRecipes()
	}
}

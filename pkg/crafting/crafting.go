// Package crafting provides scrap-to-ammo crafting and recipe management.
package crafting

import (
	"fmt"
	"sync"
)

// Scrap represents a crafting material resource.
type Scrap struct {
	Type   string
	Amount int
}

// ScrapStorage manages scrap inventory.
type ScrapStorage struct {
	scrap map[string]int
	mu    sync.RWMutex
}

// NewScrapStorage creates an empty scrap storage.
func NewScrapStorage() *ScrapStorage {
	return &ScrapStorage{
		scrap: make(map[string]int),
	}
}

// Add adds scrap of a type.
func (s *ScrapStorage) Add(scrapType string, amount int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scrap[scrapType] += amount
}

// Remove removes scrap of a type. Returns true if successful.
func (s *ScrapStorage) Remove(scrapType string, amount int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.scrap[scrapType] < amount {
		return false
	}
	s.scrap[scrapType] -= amount
	if s.scrap[scrapType] == 0 {
		delete(s.scrap, scrapType)
	}
	return true
}

// Get returns the amount of a scrap type.
func (s *ScrapStorage) Get(scrapType string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scrap[scrapType]
}

// GetAll returns all scrap types and amounts.
func (s *ScrapStorage) GetAll() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]int, len(s.scrap))
	for k, v := range s.scrap {
		result[k] = v
	}
	return result
}

// Recipe defines a crafting recipe.
type Recipe struct {
	ID        string
	Name      string
	Inputs    map[string]int // itemID -> quantity
	OutputID  string
	OutputQty int
}

// CraftingMenu provides UI access to recipes and crafting.
type CraftingMenu struct {
	storage *ScrapStorage
	recipes []Recipe
	genreID string
	mu      sync.RWMutex
}

// NewCraftingMenu creates a crafting menu with the given scrap storage.
func NewCraftingMenu(storage *ScrapStorage, genreID string) *CraftingMenu {
	if storage == nil {
		storage = NewScrapStorage()
	}

	SetGenre(genreID)

	return &CraftingMenu{
		storage: storage,
		recipes: GetRecipes(),
		genreID: genreID,
	}
}

// GetAvailableRecipes returns recipes that can be crafted with current scrap.
func (m *CraftingMenu) GetAvailableRecipes() []Recipe {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := make([]Recipe, 0)
	scrapAmounts := m.storage.GetAll()

	for _, recipe := range m.recipes {
		canCraft := true
		for material, required := range recipe.Inputs {
			if scrapAmounts[material] < required {
				canCraft = false
				break
			}
		}
		if canCraft {
			available = append(available, recipe)
		}
	}

	return available
}

// GetAllRecipes returns all recipes for the current genre.
func (m *CraftingMenu) GetAllRecipes() []Recipe {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.recipes
}

// Craft attempts to craft a recipe by ID.
// Returns output item ID, quantity, and error.
func (m *CraftingMenu) Craft(recipeID string) (string, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find recipe
	var recipe *Recipe
	for i := range m.recipes {
		if m.recipes[i].ID == recipeID {
			recipe = &m.recipes[i]
			break
		}
	}
	if recipe == nil {
		return "", 0, fmt.Errorf("recipe not found: %s", recipeID)
	}

	// Check materials
	scrapAmounts := m.storage.GetAll()
	outputID, outputQty, success := Craft(*recipe, scrapAmounts)
	if !success {
		return "", 0, fmt.Errorf("insufficient materials for recipe: %s", recipeID)
	}

	// Consume materials
	for material, required := range recipe.Inputs {
		if !m.storage.Remove(material, required) {
			return "", 0, fmt.Errorf("failed to consume materials for recipe: %s", recipeID)
		}
	}

	return outputID, outputQty, nil
}

// GetScrapAmounts returns all scrap types and amounts.
func (m *CraftingMenu) GetScrapAmounts() map[string]int {
	return m.storage.GetAll()
}

// GetScrapNameForGenre returns the genre-specific scrap name.
func GetScrapNameForGenre(genreID string) string {
	switch genreID {
	case "fantasy":
		return "bone_chips"
	case "scifi":
		return "circuit_boards"
	case "horror":
		return "flesh"
	case "cyberpunk":
		return "data_shards"
	case "postapoc":
		return "salvage"
	default:
		return "scrap"
	}
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

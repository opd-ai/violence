package crafting

import "testing"

func TestCraft(t *testing.T) {
	tests := []struct {
		name          string
		recipe        Recipe
		available     map[string]int
		wantOutputID  string
		wantOutputQty int
		wantSuccess   bool
	}{
		{
			name: "successful craft",
			recipe: Recipe{
				ID:        "bullets",
				Name:      "Craft Bullets",
				Inputs:    map[string]int{"scrap": 5},
				OutputID:  "bullets",
				OutputQty: 10,
			},
			available:     map[string]int{"scrap": 10},
			wantOutputID:  "bullets",
			wantOutputQty: 10,
			wantSuccess:   true,
		},
		{
			name: "insufficient materials",
			recipe: Recipe{
				ID:        "bullets",
				Name:      "Craft Bullets",
				Inputs:    map[string]int{"scrap": 5},
				OutputID:  "bullets",
				OutputQty: 10,
			},
			available:     map[string]int{"scrap": 3},
			wantOutputID:  "",
			wantOutputQty: 0,
			wantSuccess:   false,
		},
		{
			name: "missing material",
			recipe: Recipe{
				ID:        "medkit",
				Name:      "Craft Medkit",
				Inputs:    map[string]int{"scrap": 12, "components": 3},
				OutputID:  "medkit",
				OutputQty: 1,
			},
			available:     map[string]int{"scrap": 20},
			wantOutputID:  "",
			wantOutputQty: 0,
			wantSuccess:   false,
		},
		{
			name: "exact materials",
			recipe: Recipe{
				ID:        "shells",
				Name:      "Craft Shells",
				Inputs:    map[string]int{"scrap": 8},
				OutputID:  "shells",
				OutputQty: 5,
			},
			available:     map[string]int{"scrap": 8},
			wantOutputID:  "shells",
			wantOutputQty: 5,
			wantSuccess:   true,
		},
		{
			name: "multiple inputs",
			recipe: Recipe{
				ID:        "rockets",
				Name:      "Craft Rockets",
				Inputs:    map[string]int{"scrap": 15, "fuel": 5},
				OutputID:  "rockets",
				OutputQty: 2,
			},
			available:     map[string]int{"scrap": 20, "fuel": 10},
			wantOutputID:  "rockets",
			wantOutputQty: 2,
			wantSuccess:   true,
		},
		{
			name: "nil inputs",
			recipe: Recipe{
				ID:        "test",
				Name:      "Test",
				Inputs:    nil,
				OutputID:  "test",
				OutputQty: 1,
			},
			available:     map[string]int{"scrap": 10},
			wantOutputID:  "",
			wantOutputQty: 0,
			wantSuccess:   false,
		},
		{
			name: "nil available",
			recipe: Recipe{
				ID:        "bullets",
				Name:      "Craft Bullets",
				Inputs:    map[string]int{"scrap": 5},
				OutputID:  "bullets",
				OutputQty: 10,
			},
			available:     nil,
			wantOutputID:  "",
			wantOutputQty: 0,
			wantSuccess:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputID, outputQty, success := Craft(tt.recipe, tt.available)
			if success != tt.wantSuccess {
				t.Errorf("Craft() success = %v, want %v", success, tt.wantSuccess)
			}
			if outputID != tt.wantOutputID {
				t.Errorf("Craft() outputID = %v, want %v", outputID, tt.wantOutputID)
			}
			if outputQty != tt.wantOutputQty {
				t.Errorf("Craft() outputQty = %v, want %v", outputQty, tt.wantOutputQty)
			}
		})
	}
}

func TestGetRecipes(t *testing.T) {
	// Set genre to populate cache
	SetGenre("fantasy")
	recipes := GetRecipes()
	if len(recipes) == 0 {
		t.Fatal("GetRecipes() returned empty slice")
	}
	// Should return fantasy recipes
	found := false
	for _, r := range recipes {
		if r.ID == "arrows" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetRecipes() should return fantasy-specific recipes")
	}
}

func TestGetRecipe(t *testing.T) {
	SetGenre("fantasy")
	recipe := GetRecipe("arrows")
	if recipe == nil {
		t.Fatal("GetRecipe(arrows) returned nil")
	}
	if recipe.ID != "arrows" {
		t.Errorf("GetRecipe(arrows).ID = %s, want arrows", recipe.ID)
	}

	recipe = GetRecipe("nonexistent")
	if recipe != nil {
		t.Errorf("GetRecipe(nonexistent) = %+v, want nil", recipe)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		genreID      string
		expectedID   string
		expectedName string
		scrapType    string
	}{
		{
			genreID:      "fantasy",
			expectedID:   "arrows",
			expectedName: "Craft Arrows",
			scrapType:    "bone_chips",
		},
		{
			genreID:      "scifi",
			expectedID:   "bullets",
			expectedName: "Fabricate Bullets",
			scrapType:    "circuit_boards",
		},
		{
			genreID:      "horror",
			expectedID:   "bullets",
			expectedName: "Assemble Bullets",
			scrapType:    "flesh",
		},
		{
			genreID:      "cyberpunk",
			expectedID:   "bullets",
			expectedName: "Print Bullets",
			scrapType:    "data_shards",
		},
		{
			genreID:      "postapoc",
			expectedID:   "bullets",
			expectedName: "Scavenge Bullets",
			scrapType:    "salvage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			SetGenre(tt.genreID)
			recipes := GetRecipes()
			if len(recipes) == 0 {
				t.Fatal("GetRecipes() returned empty slice")
			}
			found := false
			for _, r := range recipes {
				if r.ID == tt.expectedID {
					found = true
					if r.Name != tt.expectedName {
						t.Errorf("Recipe name = %s, want %s", r.Name, tt.expectedName)
					}
					if _, ok := r.Inputs[tt.scrapType]; !ok {
						t.Errorf("Recipe inputs should contain %s, got %+v", tt.scrapType, r.Inputs)
					}
					break
				}
			}
			if !found {
				t.Errorf("Genre %s should have recipe %s", tt.genreID, tt.expectedID)
			}
		})
	}
}

func TestGenreRecipeDistinctness(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	recipeNames := make(map[string]map[string]bool)

	for _, genre := range genres {
		SetGenre(genre)
		recipes := GetRecipes()
		recipeNames[genre] = make(map[string]bool)
		for _, r := range recipes {
			recipeNames[genre][r.Name] = true
		}
	}

	// Verify each genre has distinct recipe names (not all the same)
	baseline := recipeNames["fantasy"]
	allSame := true
	for _, genre := range genres[1:] {
		if len(recipeNames[genre]) != len(baseline) {
			allSame = false
			break
		}
		for name := range recipeNames[genre] {
			if !baseline[name] {
				allSame = false
				break
			}
		}
	}
	if allSame {
		t.Error("All genres have identical recipe names - genre differentiation not working")
	}
}

func TestGenreScrapTypes(t *testing.T) {
	scrapTypes := map[string]string{
		"fantasy":   "bone_chips",
		"scifi":     "circuit_boards",
		"horror":    "flesh",
		"cyberpunk": "data_shards",
		"postapoc":  "salvage",
	}

	for genre, expectedScrap := range scrapTypes {
		t.Run(genre, func(t *testing.T) {
			SetGenre(genre)
			recipes := GetRecipes()
			foundScrapType := false
			for _, r := range recipes {
				for scrap := range r.Inputs {
					if scrap == expectedScrap {
						foundScrapType = true
						break
					}
				}
				if foundScrapType {
					break
				}
			}
			if !foundScrapType {
				t.Errorf("Genre %s recipes should use %s scrap type", genre, expectedScrap)
			}
		})
	}
}

func TestDefaultRecipes(t *testing.T) {
	recipes := getDefaultRecipes()
	if len(recipes) != 5 {
		t.Errorf("getDefaultRecipes() returned %d recipes, want 5", len(recipes))
	}

	expectedIDs := []string{"bullets", "shells", "cells", "rockets", "medkit"}
	for _, id := range expectedIDs {
		found := false
		for _, r := range recipes {
			if r.ID == id {
				found = true
				if r.OutputID != id {
					t.Errorf("Recipe %s OutputID = %s, want %s", id, r.OutputID, id)
				}
				if r.Inputs["scrap"] <= 0 {
					t.Errorf("Recipe %s should require scrap > 0", id)
				}
				break
			}
		}
		if !found {
			t.Errorf("Default recipes missing %s", id)
		}
	}
}

func TestRecipeOutputQuantities(t *testing.T) {
	SetGenre("fantasy")
	recipes := GetRecipes()
	for _, r := range recipes {
		if r.OutputQty <= 0 {
			t.Errorf("Recipe %s has invalid OutputQty %d", r.ID, r.OutputQty)
		}
		for item, qty := range r.Inputs {
			if qty <= 0 {
				t.Errorf("Recipe %s has invalid input quantity for %s: %d", r.ID, item, qty)
			}
		}
	}
}

func TestUnknownGenreFallback(t *testing.T) {
	SetGenre("unknown_genre")
	recipes := GetRecipes()
	if len(recipes) == 0 {
		t.Fatal("Unknown genre should fall back to default recipes")
	}
	// Should return default recipes with "scrap" as input
	found := false
	for _, r := range recipes {
		if _, ok := r.Inputs["scrap"]; ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("Unknown genre fallback should use 'scrap' as input type")
	}
}

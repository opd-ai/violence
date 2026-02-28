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

// ScrapStorage tests

func TestNewScrapStorage(t *testing.T) {
	storage := NewScrapStorage()
	if storage == nil {
		t.Fatal("NewScrapStorage returned nil")
	}
	if storage.scrap == nil {
		t.Fatal("scrap map not initialized")
	}
}

func TestScrapStorage_AddAndGet(t *testing.T) {
	storage := NewScrapStorage()

	storage.Add("scrap", 10)
	if storage.Get("scrap") != 10 {
		t.Errorf("Get(scrap) = %d, want 10", storage.Get("scrap"))
	}

	storage.Add("scrap", 5)
	if storage.Get("scrap") != 15 {
		t.Errorf("Get(scrap) after second add = %d, want 15", storage.Get("scrap"))
	}

	// Get non-existent type
	if storage.Get("nonexistent") != 0 {
		t.Errorf("Get(nonexistent) = %d, want 0", storage.Get("nonexistent"))
	}
}

func TestScrapStorage_Remove(t *testing.T) {
	tests := []struct {
		name         string
		initialScrap map[string]int
		removeType   string
		removeAmount int
		wantSuccess  bool
		wantRemain   int
	}{
		{
			name:         "remove partial amount",
			initialScrap: map[string]int{"scrap": 10},
			removeType:   "scrap",
			removeAmount: 3,
			wantSuccess:  true,
			wantRemain:   7,
		},
		{
			name:         "remove exact amount",
			initialScrap: map[string]int{"scrap": 10},
			removeType:   "scrap",
			removeAmount: 10,
			wantSuccess:  true,
			wantRemain:   0,
		},
		{
			name:         "remove more than available",
			initialScrap: map[string]int{"scrap": 5},
			removeType:   "scrap",
			removeAmount: 10,
			wantSuccess:  false,
			wantRemain:   5,
		},
		{
			name:         "remove from non-existent type",
			initialScrap: map[string]int{},
			removeType:   "scrap",
			removeAmount: 5,
			wantSuccess:  false,
			wantRemain:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewScrapStorage()
			for scrapType, amount := range tt.initialScrap {
				storage.Add(scrapType, amount)
			}

			success := storage.Remove(tt.removeType, tt.removeAmount)
			if success != tt.wantSuccess {
				t.Errorf("Remove() = %v, want %v", success, tt.wantSuccess)
			}

			remaining := storage.Get(tt.removeType)
			if remaining != tt.wantRemain {
				t.Errorf("After Remove(), remaining = %d, want %d", remaining, tt.wantRemain)
			}
		})
	}
}

func TestScrapStorage_GetAll(t *testing.T) {
	storage := NewScrapStorage()
	storage.Add("scrap", 10)
	storage.Add("bone_chips", 5)
	storage.Add("circuit_boards", 3)

	all := storage.GetAll()
	if len(all) != 3 {
		t.Fatalf("GetAll() returned %d types, want 3", len(all))
	}

	expected := map[string]int{
		"scrap":          10,
		"bone_chips":     5,
		"circuit_boards": 3,
	}

	for scrapType, amount := range expected {
		if all[scrapType] != amount {
			t.Errorf("GetAll()[%s] = %d, want %d", scrapType, all[scrapType], amount)
		}
	}

	// Verify it's a copy, not the original map
	all["scrap"] = 999
	if storage.Get("scrap") == 999 {
		t.Error("GetAll() should return a copy, not the original map")
	}
}

// CraftingMenu tests

func TestNewCraftingMenu(t *testing.T) {
	storage := NewScrapStorage()
	storage.Add("bone_chips", 100)

	menu := NewCraftingMenu(storage, "fantasy")
	if menu == nil {
		t.Fatal("NewCraftingMenu returned nil")
	}
	if menu.storage == nil {
		t.Fatal("menu.storage is nil")
	}
	if len(menu.recipes) == 0 {
		t.Fatal("menu.recipes is empty")
	}
	if menu.genreID != "fantasy" {
		t.Errorf("menu.genreID = %s, want fantasy", menu.genreID)
	}
}

func TestNewCraftingMenu_NilStorage(t *testing.T) {
	menu := NewCraftingMenu(nil, "scifi")
	if menu == nil {
		t.Fatal("NewCraftingMenu returned nil")
	}
	if menu.storage == nil {
		t.Fatal("NewCraftingMenu should create storage if nil")
	}
}

func TestCraftingMenu_GetAllRecipes(t *testing.T) {
	menu := NewCraftingMenu(nil, "fantasy")
	recipes := menu.GetAllRecipes()

	if len(recipes) == 0 {
		t.Fatal("GetAllRecipes() returned empty slice")
	}

	// Should have fantasy recipes
	found := false
	for _, r := range recipes {
		if r.ID == "arrows" {
			found = true
			break
		}
	}
	if !found {
		t.Error("GetAllRecipes() should return fantasy-specific recipes")
	}
}

func TestCraftingMenu_GetAvailableRecipes(t *testing.T) {
	storage := NewScrapStorage()
	storage.Add("bone_chips", 10) // Enough for arrows (5) but not explosives (15)

	menu := NewCraftingMenu(storage, "fantasy")
	available := menu.GetAvailableRecipes()

	// Should be able to craft arrows and bolts
	canCraftArrows := false
	canCraftExplosives := false

	for _, r := range available {
		if r.ID == "arrows" {
			canCraftArrows = true
		}
		if r.ID == "explosives" {
			canCraftExplosives = true
		}
	}

	if !canCraftArrows {
		t.Error("Should be able to craft arrows with 10 bone_chips")
	}
	if canCraftExplosives {
		t.Error("Should NOT be able to craft explosives with only 10 bone_chips")
	}
}

func TestCraftingMenu_Craft(t *testing.T) {
	tests := []struct {
		name           string
		initialScrap   map[string]int
		recipeID       string
		wantOutputID   string
		wantOutputQty  int
		wantErr        bool
		wantScrapAfter int
	}{
		{
			name:           "successful craft",
			initialScrap:   map[string]int{"bone_chips": 10},
			recipeID:       "arrows",
			wantOutputID:   "arrows",
			wantOutputQty:  10,
			wantErr:        false,
			wantScrapAfter: 5, // 10 - 5 = 5
		},
		{
			name:           "insufficient materials",
			initialScrap:   map[string]int{"bone_chips": 3},
			recipeID:       "arrows",
			wantOutputID:   "",
			wantOutputQty:  0,
			wantErr:        true,
			wantScrapAfter: 3,
		},
		{
			name:           "recipe not found",
			initialScrap:   map[string]int{"bone_chips": 100},
			recipeID:       "nonexistent",
			wantOutputID:   "",
			wantOutputQty:  0,
			wantErr:        true,
			wantScrapAfter: 100,
		},
		{
			name:           "exact materials consumed",
			initialScrap:   map[string]int{"bone_chips": 15},
			recipeID:       "explosives",
			wantOutputID:   "explosives",
			wantOutputQty:  2,
			wantErr:        false,
			wantScrapAfter: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewScrapStorage()
			for scrapType, amount := range tt.initialScrap {
				storage.Add(scrapType, amount)
			}

			menu := NewCraftingMenu(storage, "fantasy")
			outputID, outputQty, err := menu.Craft(tt.recipeID)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Craft() error = %v, wantErr %v", err, tt.wantErr)
			}

			if outputID != tt.wantOutputID {
				t.Errorf("Craft() outputID = %s, want %s", outputID, tt.wantOutputID)
			}

			if outputQty != tt.wantOutputQty {
				t.Errorf("Craft() outputQty = %d, want %d", outputQty, tt.wantOutputQty)
			}

			// Check scrap was consumed correctly
			remaining := storage.Get("bone_chips")
			if remaining != tt.wantScrapAfter {
				t.Errorf("After Craft(), scrap = %d, want %d", remaining, tt.wantScrapAfter)
			}
		})
	}
}

func TestCraftingMenu_GetScrapAmounts(t *testing.T) {
	storage := NewScrapStorage()
	storage.Add("bone_chips", 10)
	storage.Add("salvage", 5)

	menu := NewCraftingMenu(storage, "fantasy")
	amounts := menu.GetScrapAmounts()

	if amounts["bone_chips"] != 10 {
		t.Errorf("GetScrapAmounts()[bone_chips] = %d, want 10", amounts["bone_chips"])
	}
	if amounts["salvage"] != 5 {
		t.Errorf("GetScrapAmounts()[salvage] = %d, want 5", amounts["salvage"])
	}
}

func TestGetScrapNameForGenre(t *testing.T) {
	tests := []struct {
		genreID  string
		wantName string
	}{
		{"fantasy", "bone_chips"},
		{"scifi", "circuit_boards"},
		{"horror", "flesh"},
		{"cyberpunk", "data_shards"},
		{"postapoc", "salvage"},
		{"unknown", "scrap"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			name := GetScrapNameForGenre(tt.genreID)
			if name != tt.wantName {
				t.Errorf("GetScrapNameForGenre(%s) = %s, want %s", tt.genreID, name, tt.wantName)
			}
		})
	}
}

func TestCraftingMenu_MultipleCrafts(t *testing.T) {
	storage := NewScrapStorage()
	storage.Add("bone_chips", 50)

	menu := NewCraftingMenu(storage, "fantasy")

	// Craft arrows twice
	_, _, err := menu.Craft("arrows")
	if err != nil {
		t.Fatalf("First craft failed: %v", err)
	}

	_, _, err = menu.Craft("arrows")
	if err != nil {
		t.Fatalf("Second craft failed: %v", err)
	}

	// Should have consumed 10 bone_chips (5+5)
	remaining := storage.Get("bone_chips")
	if remaining != 40 {
		t.Errorf("After 2 crafts, bone_chips = %d, want 40", remaining)
	}
}

func TestCraftingMenu_ConcurrentAccess(t *testing.T) {
	storage := NewScrapStorage()
	storage.Add("bone_chips", 1000)

	menu := NewCraftingMenu(storage, "fantasy")

	done := make(chan bool)

	// Multiple goroutines crafting
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				menu.Craft("arrows")
				menu.GetAvailableRecipes()
				menu.GetAllRecipes()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

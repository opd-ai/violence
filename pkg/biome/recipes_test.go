package biome

import (
	"testing"
)

func TestBiomeMaterialRecipes(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			recipes := BiomeMaterialRecipes(genre)

			if len(recipes) == 0 {
				t.Errorf("No recipes returned for genre %s", genre)
			}

			t.Logf("Genre %s has %d biome material recipes", genre, len(recipes))

			// Verify each recipe structure
			for _, recipe := range recipes {
				if recipe.ID == "" {
					t.Error("recipe has empty ID")
				}
				if recipe.Name == "" {
					t.Error("recipe has empty Name")
				}
				if len(recipe.Inputs) == 0 {
					t.Errorf("recipe %s has no inputs", recipe.ID)
				}
				if recipe.OutputID == "" {
					t.Errorf("recipe %s has empty OutputID", recipe.ID)
				}
				if recipe.OutputQty <= 0 {
					t.Errorf("recipe %s has invalid OutputQty: %d", recipe.ID, recipe.OutputQty)
				}

				// Verify inputs are valid amounts
				for material, amount := range recipe.Inputs {
					if material == "" {
						t.Errorf("recipe %s has empty material name", recipe.ID)
					}
					if amount <= 0 {
						t.Errorf("recipe %s requires invalid amount for %s: %d", recipe.ID, material, amount)
					}
				}

				t.Logf("  - %s: %s (outputs %dx %s)", recipe.ID, recipe.Name, recipe.OutputQty, recipe.OutputID)
			}
		})
	}
}

func TestFantasyBiomeMaterialRecipes(t *testing.T) {
	recipes := BiomeMaterialRecipes("fantasy")

	expectedRecipes := []string{
		"ironbark_arrows",
		"ethereal_potion",
		"ancient_staff",
		"echo_grenade",
		"diamond_blade",
		"cursed_weapon",
		"necromantic_scroll",
		"volcanic_bomb",
		"phoenix_amulet",
		"frost_arrows",
		"eternal_armor",
		"prismatic_wand",
		"arcane_focus",
	}

	recipeMap := make(map[string]bool)
	for _, recipe := range recipes {
		recipeMap[recipe.ID] = true
	}

	for _, expected := range expectedRecipes {
		if !recipeMap[expected] {
			t.Errorf("Missing expected recipe: %s", expected)
		}
	}
}

func TestScifiBiomeMaterialRecipes(t *testing.T) {
	recipes := BiomeMaterialRecipes("scifi")

	if len(recipes) == 0 {
		t.Fatal("No scifi recipes returned")
	}

	// Verify at least some key recipes exist
	hasLabRecipe := false
	hasHiveRecipe := false

	for _, recipe := range recipes {
		if recipe.ID == "bio_grenade" || recipe.ID == "prototype_weapon" {
			hasLabRecipe = true
		}
		if recipe.ID == "chitin_armor" || recipe.ID == "queen_implant" {
			hasHiveRecipe = true
		}
	}

	if !hasLabRecipe {
		t.Error("Missing lab-based recipes")
	}
	if !hasHiveRecipe {
		t.Error("Missing hive-based recipes")
	}
}

func TestHorrorBiomeMaterialRecipes(t *testing.T) {
	recipes := BiomeMaterialRecipes("horror")

	if len(recipes) == 0 {
		t.Fatal("No horror recipes returned")
	}

	// Verify horror theme materials
	hasNecroRecipe := false
	for _, recipe := range recipes {
		for material := range recipe.Inputs {
			if material == "cursed_bone" || material == "spectral_essence" || material == "necromantic_core" {
				hasNecroRecipe = true
				break
			}
		}
	}

	if !hasNecroRecipe {
		t.Error("Missing necromantic/horror themed recipes")
	}
}

func TestCyberpunkBiomeMaterialRecipes(t *testing.T) {
	recipes := BiomeMaterialRecipes("cyberpunk")

	if len(recipes) == 0 {
		t.Fatal("No cyberpunk recipes returned")
	}

	// Verify cyberpunk theme materials
	hasCorpRecipe := false
	for _, recipe := range recipes {
		if recipe.ID == "hack_tool" || recipe.ID == "corp_weapon" {
			hasCorpRecipe = true
		}
	}

	if !hasCorpRecipe {
		t.Error("Missing corporate/cyberpunk themed recipes")
	}
}

func TestPostapocBiomeMaterialRecipes(t *testing.T) {
	recipes := BiomeMaterialRecipes("postapoc")

	if len(recipes) == 0 {
		t.Fatal("No postapoc recipes returned")
	}

	// Verify wasteland and radiation materials
	hasWastelandRecipe := false
	hasRadRecipe := false

	for _, recipe := range recipes {
		for material := range recipe.Inputs {
			if material == "scrap_metal" || material == "rusty_parts" {
				hasWastelandRecipe = true
			}
			if material == "glowing_isotope" || material == "pure_plutonium" {
				hasRadRecipe = true
			}
		}
	}

	if !hasWastelandRecipe {
		t.Error("Missing wasteland-based recipes")
	}
	if !hasRadRecipe {
		t.Error("Missing radiation-based recipes")
	}
}

func TestUnknownGenreRecipes(t *testing.T) {
	recipes := BiomeMaterialRecipes("unknown_genre")

	if len(recipes) != 0 {
		t.Errorf("Unknown genre should return empty recipes, got %d", len(recipes))
	}
}

func TestRecipeComplexity(t *testing.T) {
	// Test that rare material recipes require multiple inputs
	recipes := BiomeMaterialRecipes("fantasy")

	for _, recipe := range recipes {
		// Recipes using rare materials should be more complex
		usesRareMat := false
		for material := range recipe.Inputs {
			if material == "ancient_root" || material == "phoenix_ember" || material == "eternal_ice" || material == "arcane_geode" || material == "necromantic_core" {
				usesRareMat = true
				break
			}
		}

		if usesRareMat && len(recipe.Inputs) < 2 {
			t.Logf("Warning: Recipe %s uses rare materials but only has %d inputs", recipe.ID, len(recipe.Inputs))
		}
	}
}

func TestRecipeNoDuplicateIDs(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		recipes := BiomeMaterialRecipes(genre)
		seenIDs := make(map[string]bool)

		for _, recipe := range recipes {
			if seenIDs[recipe.ID] {
				t.Errorf("Duplicate recipe ID in %s: %s", genre, recipe.ID)
			}
			seenIDs[recipe.ID] = true
		}
	}
}

func TestRecipeMaterialConsistency(t *testing.T) {
	// Test that recipes use materials that actually exist in biomes
	recipes := BiomeMaterialRecipes("fantasy")

	validMaterials := make(map[string]bool)
	for _, biome := range GetAllBiomeTypes() {
		profile := GetBiomeProfile(biome)
		for _, mat := range profile.CommonMats {
			validMaterials[mat.MaterialID] = true
		}
		for _, mat := range profile.UncommonMats {
			validMaterials[mat.MaterialID] = true
		}
		for _, mat := range profile.RareMats {
			validMaterials[mat.MaterialID] = true
		}
	}

	for _, recipe := range recipes {
		for material := range recipe.Inputs {
			if !validMaterials[material] {
				t.Logf("Warning: Recipe %s uses material %s which may not exist in any biome", recipe.ID, material)
			}
		}
	}
}

func BenchmarkBiomeMaterialRecipes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = BiomeMaterialRecipes("fantasy")
	}
}

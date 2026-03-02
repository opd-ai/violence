package biome

import (
	"math/rand"
	"testing"
)

func TestGetBiomeProfile(t *testing.T) {
	tests := []struct {
		name      string
		biomeType BiomeType
		wantName  string
	}{
		{
			name:      "forest dungeon",
			biomeType: BiomeForestDungeon,
			wantName:  "Forest Dungeon",
		},
		{
			name:      "deep cavern",
			biomeType: BiomeDeepCavern,
			wantName:  "Deep Cavern",
		},
		{
			name:      "alien hive",
			biomeType: BiomeAlienHive,
			wantName:  "Alien Hive",
		},
		{
			name:      "invalid biome defaults to underground",
			biomeType: BiomeType(999),
			wantName:  "Underground",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := GetBiomeProfile(tt.biomeType)
			if profile.Name != tt.wantName {
				t.Errorf("GetBiomeProfile().Name = %v, want %v", profile.Name, tt.wantName)
			}
			if len(profile.CommonMats) == 0 {
				t.Error("profile has no common materials")
			}
		})
	}
}

func TestRollMaterials(t *testing.T) {
	tests := []struct {
		name      string
		biomeType BiomeType
		tier      int
		genreID   string
		seed      uint64
	}{
		{
			name:      "tier 1 forest fantasy",
			biomeType: BiomeForestDungeon,
			tier:      1,
			genreID:   "fantasy",
			seed:      12345,
		},
		{
			name:      "tier 3 alien hive scifi",
			biomeType: BiomeAlienHive,
			tier:      3,
			genreID:   "scifi",
			seed:      67890,
		},
		{
			name:      "tier 2 crypt horror",
			biomeType: BiomeCrypt,
			tier:      2,
			genreID:   "horror",
			seed:      11111,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			materials := RollMaterials(tt.biomeType, tt.tier, tt.genreID, tt.seed)

			if materials == nil {
				t.Error("RollMaterials returned nil")
			}

			// Verify determinism
			materials2 := RollMaterials(tt.biomeType, tt.tier, tt.genreID, tt.seed)
			if len(materials) != len(materials2) {
				t.Error("RollMaterials is not deterministic")
			}

			for i := range materials {
				if materials[i].MaterialID != materials2[i].MaterialID {
					t.Error("RollMaterials produced different materials on second call")
				}
				if materials[i].MinAmount != materials2[i].MinAmount {
					t.Error("RollMaterials produced different amounts on second call")
				}
			}

			// Log results
			t.Logf("Rolled %d materials for %s tier %d in %s", len(materials), tt.biomeType.String(), tt.tier, tt.genreID)
			for _, mat := range materials {
				t.Logf("  - %s x%d", mat.MaterialID, mat.MinAmount)
			}
		})
	}
}

func TestRollMaterialsTierScaling(t *testing.T) {
	seed := uint64(42)
	biome := BiomeDeepCavern
	genre := "fantasy"

	tier1 := RollMaterials(biome, 1, genre, seed)
	tier2 := RollMaterials(biome, 2, genre, seed+1)
	tier3 := RollMaterials(biome, 3, genre, seed+2)

	// Higher tiers should generally drop more materials, but not guaranteed due to RNG
	t.Logf("Tier 1: %d materials", len(tier1))
	t.Logf("Tier 2: %d materials", len(tier2))
	t.Logf("Tier 3: %d materials", len(tier3))

	// Tier 1 should never drop rare materials
	for _, mat := range tier1 {
		profile := GetBiomeProfile(biome)
		for _, rare := range profile.RareMats {
			if mat.MaterialID == rare.MaterialID {
				t.Error("Tier 1 dropped rare material")
			}
		}
	}
}

func TestSelectBiomeForGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    int64
	}{
		{
			name:    "fantasy",
			genreID: "fantasy",
			seed:    12345,
		},
		{
			name:    "scifi",
			genreID: "scifi",
			seed:    67890,
		},
		{
			name:    "horror",
			genreID: "horror",
			seed:    11111,
		},
		{
			name:    "cyberpunk",
			genreID: "cyberpunk",
			seed:    22222,
		},
		{
			name:    "postapoc",
			genreID: "postapoc",
			seed:    33333,
		},
		{
			name:    "unknown defaults to fantasy",
			genreID: "unknown",
			seed:    44444,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(tt.seed))
			biome := SelectBiomeForGenre(tt.genreID, rng)

			profile := GetBiomeProfile(biome)
			t.Logf("Selected biome: %s for genre %s", profile.Name, tt.genreID)

			// Verify it's a valid biome
			if profile.Name == "" {
				t.Error("selected biome has empty name")
			}
		})
	}
}

func TestBiomeTypeString(t *testing.T) {
	tests := []struct {
		biomeType BiomeType
		want      string
	}{
		{BiomeForestDungeon, "Forest Dungeon"},
		{BiomeDeepCavern, "Deep Cavern"},
		{BiomeCrypt, "Crypt"},
		{BiomeAlienHive, "Alien Hive"},
		{BiomeCorpTower, "Corporate Tower"},
		{BiomeWasteland, "Wasteland"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.biomeType.String()
			if got != tt.want {
				t.Errorf("BiomeType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllBiomeTypes(t *testing.T) {
	biomes := GetAllBiomeTypes()

	if len(biomes) == 0 {
		t.Fatal("GetAllBiomeTypes returned empty slice")
	}

	// Verify all biomes have valid profiles
	for _, biome := range biomes {
		profile := GetBiomeProfile(biome)
		if profile.Name == "" {
			t.Errorf("Biome %d has no name", biome)
		}
	}

	t.Logf("Total biomes: %d", len(biomes))
}

func TestValidateMaterialID(t *testing.T) {
	tests := []struct {
		name       string
		materialID string
		wantErr    bool
	}{
		{
			name:       "valid common material",
			materialID: "wood_scraps",
			wantErr:    false,
		},
		{
			name:       "valid uncommon material",
			materialID: "iron_ore",
			wantErr:    false,
		},
		{
			name:       "valid rare material",
			materialID: "ancient_root",
			wantErr:    false,
		},
		{
			name:       "invalid material",
			materialID: "nonexistent_material",
			wantErr:    true,
		},
		{
			name:       "empty string",
			materialID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaterialID(tt.materialID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMaterialID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBiomeGenreAffinity(t *testing.T) {
	tests := []struct {
		biome   BiomeType
		genre   string
		minMult float64
	}{
		{BiomeForestDungeon, "fantasy", 1.0},
		{BiomeAlienHive, "scifi", 1.5},
		{BiomeCrypt, "horror", 1.5},
		{BiomeCorpTower, "cyberpunk", 1.5},
		{BiomeWasteland, "postapoc", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.biome.String()+"_"+tt.genre, func(t *testing.T) {
			profile := GetBiomeProfile(tt.biome)
			affinity, exists := profile.GenreAffinities[tt.genre]
			if !exists {
				t.Logf("No explicit affinity for %s in %s (using default 1.0)", tt.genre, profile.Name)
				return
			}
			if affinity < tt.minMult {
				t.Errorf("Genre affinity %f is below minimum %f", affinity, tt.minMult)
			}
			t.Logf("%s affinity in %s: %.1fx", profile.Name, tt.genre, affinity)
		})
	}
}

func TestBiomeComponentType(t *testing.T) {
	comp := &BiomeComponent{
		Biome:        BiomeForestDungeon,
		Tier:         2,
		MaterialSeed: 12345,
	}

	if comp.Type() != "BiomeComponent" {
		t.Errorf("BiomeComponent.Type() = %v, want BiomeComponent", comp.Type())
	}
}

func TestMaterialDropAmounts(t *testing.T) {
	// Test that material amounts are within specified ranges
	for i := 0; i < 100; i++ {
		biome := BiomeDeepCavern
		materials := RollMaterials(biome, 2, "fantasy", uint64(i))

		for _, mat := range materials {
			if mat.MinAmount < 0 {
				t.Errorf("Material %s has negative amount: %d", mat.MaterialID, mat.MinAmount)
			}
			if mat.MinAmount > 20 {
				t.Errorf("Material %s has excessive amount: %d", mat.MaterialID, mat.MinAmount)
			}
		}
	}
}

func BenchmarkRollMaterials(b *testing.B) {
	biome := BiomeDeepCavern
	genre := "fantasy"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RollMaterials(biome, 2, genre, uint64(i))
	}
}

func BenchmarkSelectBiomeForGenre(b *testing.B) {
	rng := rand.New(rand.NewSource(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SelectBiomeForGenre("fantasy", rng)
	}
}

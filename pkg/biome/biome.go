// Package biome provides biome-specific zone identification and material generation.
package biome

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// BiomeType defines different environmental biomes.
type BiomeType int

const (
	BiomeForestDungeon BiomeType = iota
	BiomeDeepCavern
	BiomeCrypt
	BiomeLavaShrine
	BiomeFrozenCaves
	BiomeCrystalMine
	BiomeAbandonedLab
	BiomeAlienHive
	BiomeCorpTower
	BiomeWasteland
	BiomeRadZone
	BiomeUnderground
)

// BiomeComponent marks a map region with a biome type.
type BiomeComponent struct {
	Biome        BiomeType
	Tier         int    // 1-3, affects material rarity
	MaterialSeed uint64 // Deterministic material generation
}

// Type returns the component type identifier.
func (c *BiomeComponent) Type() string {
	return "BiomeComponent"
}

// MaterialDrop represents a crafting material that can drop in this biome.
type MaterialDrop struct {
	MaterialID string
	Chance     float64
	MinAmount  int
	MaxAmount  int
}

// BiomeProfile defines materials and characteristics for a biome.
type BiomeProfile struct {
	Name            string
	CommonMats      []MaterialDrop
	UncommonMats    []MaterialDrop
	RareMats        []MaterialDrop
	GenreAffinities map[string]float64 // Multiplier per genre
}

var biomeProfiles = map[BiomeType]BiomeProfile{
	BiomeForestDungeon: {
		Name: "Forest Dungeon",
		CommonMats: []MaterialDrop{
			{MaterialID: "wood_scraps", Chance: 0.7, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "moss", Chance: 0.5, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "bark_fiber", Chance: 0.6, MinAmount: 1, MaxAmount: 4},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "ironbark", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "ethereal_sap", Chance: 0.2, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "ancient_root", Chance: 0.1, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"fantasy": 1.5, "horror": 1.2},
	},
	BiomeDeepCavern: {
		Name: "Deep Cavern",
		CommonMats: []MaterialDrop{
			{MaterialID: "stone_chunks", Chance: 0.8, MinAmount: 2, MaxAmount: 5},
			{MaterialID: "bat_guano", Chance: 0.4, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "cave_fungus", Chance: 0.5, MinAmount: 1, MaxAmount: 2},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "iron_ore", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "echo_crystals", Chance: 0.25, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "deep_diamond", Chance: 0.08, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"fantasy": 1.3, "scifi": 1.1},
	},
	BiomeCrypt: {
		Name: "Crypt",
		CommonMats: []MaterialDrop{
			{MaterialID: "bone_dust", Chance: 0.7, MinAmount: 1, MaxAmount: 4},
			{MaterialID: "grave_soil", Chance: 0.6, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "tattered_cloth", Chance: 0.5, MinAmount: 1, MaxAmount: 2},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "cursed_bone", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "spectral_essence", Chance: 0.25, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "necromantic_core", Chance: 0.12, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"fantasy": 1.4, "horror": 1.6},
	},
	BiomeLavaShrine: {
		Name: "Lava Shrine",
		CommonMats: []MaterialDrop{
			{MaterialID: "obsidian_shards", Chance: 0.6, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "sulfur", Chance: 0.7, MinAmount: 2, MaxAmount: 4},
			{MaterialID: "ash", Chance: 0.8, MinAmount: 1, MaxAmount: 5},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "fire_opal", Chance: 0.3, MinAmount: 1, MaxAmount: 1},
			{MaterialID: "volcanic_glass", Chance: 0.25, MinAmount: 1, MaxAmount: 2},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "phoenix_ember", Chance: 0.1, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"fantasy": 1.5},
	},
	BiomeFrozenCaves: {
		Name: "Frozen Caves",
		CommonMats: []MaterialDrop{
			{MaterialID: "ice_chunks", Chance: 0.7, MinAmount: 2, MaxAmount: 4},
			{MaterialID: "frost_moss", Chance: 0.5, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "frozen_bone", Chance: 0.4, MinAmount: 1, MaxAmount: 3},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "permafrost_crystal", Chance: 0.3, MinAmount: 1, MaxAmount: 1},
			{MaterialID: "winter_essence", Chance: 0.2, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "eternal_ice", Chance: 0.09, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"fantasy": 1.4, "scifi": 1.1},
	},
	BiomeCrystalMine: {
		Name: "Crystal Mine",
		CommonMats: []MaterialDrop{
			{MaterialID: "quartz_fragments", Chance: 0.8, MinAmount: 2, MaxAmount: 5},
			{MaterialID: "gem_dust", Chance: 0.6, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "ore_residue", Chance: 0.7, MinAmount: 1, MaxAmount: 4},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "prismatic_shard", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "mana_crystal", Chance: 0.25, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "arcane_geode", Chance: 0.11, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"fantasy": 1.6, "scifi": 1.3},
	},
	BiomeAbandonedLab: {
		Name: "Abandoned Lab",
		CommonMats: []MaterialDrop{
			{MaterialID: "circuit_scrap", Chance: 0.7, MinAmount: 1, MaxAmount: 4},
			{MaterialID: "chemical_vials", Chance: 0.6, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "wiring", Chance: 0.8, MinAmount: 2, MaxAmount: 5},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "bio_sample", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "data_chip", Chance: 0.25, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "research_prototype", Chance: 0.12, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"scifi": 1.6, "horror": 1.3},
	},
	BiomeAlienHive: {
		Name: "Alien Hive",
		CommonMats: []MaterialDrop{
			{MaterialID: "bio_mass", Chance: 0.8, MinAmount: 2, MaxAmount: 5},
			{MaterialID: "chitin_plates", Chance: 0.7, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "xenomorph_ichor", Chance: 0.6, MinAmount: 1, MaxAmount: 2},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "hive_resin", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "alien_pheromone", Chance: 0.2, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "queen_essence", Chance: 0.1, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"scifi": 1.7, "horror": 1.4},
	},
	BiomeCorpTower: {
		Name: "Corporate Tower",
		CommonMats: []MaterialDrop{
			{MaterialID: "eddies_cache", Chance: 0.7, MinAmount: 3, MaxAmount: 6},
			{MaterialID: "office_equipment", Chance: 0.6, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "security_chip", Chance: 0.5, MinAmount: 1, MaxAmount: 2},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "cyberware_parts", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "encrypted_data", Chance: 0.25, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "corp_secrets", Chance: 0.13, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"cyberpunk": 1.8},
	},
	BiomeWasteland: {
		Name: "Wasteland",
		CommonMats: []MaterialDrop{
			{MaterialID: "scrap_metal", Chance: 0.8, MinAmount: 2, MaxAmount: 5},
			{MaterialID: "rusty_parts", Chance: 0.7, MinAmount: 1, MaxAmount: 4},
			{MaterialID: "plastic_waste", Chance: 0.6, MinAmount: 1, MaxAmount: 3},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "intact_electronics", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "pre_war_material", Chance: 0.2, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "old_world_tech", Chance: 0.11, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"postapoc": 1.7, "scifi": 1.2},
	},
	BiomeRadZone: {
		Name: "Radiation Zone",
		CommonMats: []MaterialDrop{
			{MaterialID: "radioactive_debris", Chance: 0.7, MinAmount: 1, MaxAmount: 3},
			{MaterialID: "mutated_tissue", Chance: 0.6, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "contaminated_water", Chance: 0.5, MinAmount: 1, MaxAmount: 2},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "glowing_isotope", Chance: 0.3, MinAmount: 1, MaxAmount: 1},
			{MaterialID: "radiation_crystal", Chance: 0.25, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "pure_plutonium", Chance: 0.1, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"postapoc": 1.6, "scifi": 1.3},
	},
	BiomeUnderground: {
		Name: "Underground",
		CommonMats: []MaterialDrop{
			{MaterialID: "dirt", Chance: 0.8, MinAmount: 2, MaxAmount: 6},
			{MaterialID: "rock_chunks", Chance: 0.7, MinAmount: 1, MaxAmount: 4},
			{MaterialID: "roots", Chance: 0.5, MinAmount: 1, MaxAmount: 3},
		},
		UncommonMats: []MaterialDrop{
			{MaterialID: "mineral_vein", Chance: 0.3, MinAmount: 1, MaxAmount: 2},
			{MaterialID: "underground_spring_water", Chance: 0.2, MinAmount: 1, MaxAmount: 1},
		},
		RareMats: []MaterialDrop{
			{MaterialID: "ancient_artifact_fragment", Chance: 0.09, MinAmount: 1, MaxAmount: 1},
		},
		GenreAffinities: map[string]float64{"fantasy": 1.2, "scifi": 1.1, "postapoc": 1.2},
	},
}

// GetBiomeProfile returns the profile for a biome type.
func GetBiomeProfile(biomeType BiomeType) BiomeProfile {
	profile, exists := biomeProfiles[biomeType]
	if !exists {
		return biomeProfiles[BiomeUnderground]
	}
	return profile
}

// RollMaterials generates material drops for a biome based on tier and genre.
func RollMaterials(biomeType BiomeType, tier int, genreID string, seed uint64) []MaterialDrop {
	profile := GetBiomeProfile(biomeType)
	localRNG := rng.NewRNG(seed)

	affinity := 1.0
	if mult, ok := profile.GenreAffinities[genreID]; ok {
		affinity = mult
	}

	var results []MaterialDrop

	// Roll common materials
	for _, mat := range profile.CommonMats {
		adjustedChance := mat.Chance * affinity
		if tier >= 2 {
			adjustedChance *= 1.2
		}
		if tier >= 3 {
			adjustedChance *= 1.3
		}

		if localRNG.Float64() < adjustedChance {
			amount := mat.MinAmount
			if mat.MaxAmount > mat.MinAmount {
				amount += localRNG.Intn(mat.MaxAmount - mat.MinAmount + 1)
			}
			results = append(results, MaterialDrop{
				MaterialID: mat.MaterialID,
				Chance:     adjustedChance,
				MinAmount:  amount,
				MaxAmount:  amount,
			})
		}
	}

	// Roll uncommon materials
	for _, mat := range profile.UncommonMats {
		adjustedChance := mat.Chance * affinity
		if tier >= 2 {
			adjustedChance *= 1.3
		}
		if tier >= 3 {
			adjustedChance *= 1.5
		}

		if localRNG.Float64() < adjustedChance {
			amount := mat.MinAmount
			if mat.MaxAmount > mat.MinAmount {
				amount += localRNG.Intn(mat.MaxAmount - mat.MinAmount + 1)
			}
			results = append(results, MaterialDrop{
				MaterialID: mat.MaterialID,
				Chance:     adjustedChance,
				MinAmount:  amount,
				MaxAmount:  amount,
			})
		}
	}

	// Roll rare materials (only for tier 2+)
	if tier >= 2 {
		for _, mat := range profile.RareMats {
			adjustedChance := mat.Chance * affinity
			if tier >= 3 {
				adjustedChance *= 2.0
			}

			if localRNG.Float64() < adjustedChance {
				amount := mat.MinAmount
				if mat.MaxAmount > mat.MinAmount {
					amount += localRNG.Intn(mat.MaxAmount - mat.MinAmount + 1)
				}
				results = append(results, MaterialDrop{
					MaterialID: mat.MaterialID,
					Chance:     adjustedChance,
					MinAmount:  amount,
					MaxAmount:  amount,
				})
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"system_name":    "BiomeMaterialSystem",
		"biome":          profile.Name,
		"tier":           tier,
		"genre":          genreID,
		"material_count": len(results),
	}).Trace("Rolled biome materials")

	return results
}

// SelectBiomeForGenre chooses an appropriate biome type for a genre.
func SelectBiomeForGenre(genreID string, rngSource *rand.Rand) BiomeType {
	genreBiomes := map[string][]BiomeType{
		"fantasy": {
			BiomeForestDungeon,
			BiomeDeepCavern,
			BiomeCrypt,
			BiomeLavaShrine,
			BiomeFrozenCaves,
			BiomeCrystalMine,
		},
		"scifi": {
			BiomeAbandonedLab,
			BiomeAlienHive,
			BiomeDeepCavern,
			BiomeCrystalMine,
		},
		"horror": {
			BiomeCrypt,
			BiomeAbandonedLab,
			BiomeForestDungeon,
		},
		"cyberpunk": {
			BiomeCorpTower,
			BiomeAbandonedLab,
		},
		"postapoc": {
			BiomeWasteland,
			BiomeRadZone,
			BiomeAbandonedLab,
			BiomeUnderground,
		},
	}

	biomes, exists := genreBiomes[genreID]
	if !exists || len(biomes) == 0 {
		biomes = genreBiomes["fantasy"]
	}

	idx := rngSource.Intn(len(biomes))
	return biomes[idx]
}

// String returns the biome name.
func (bt BiomeType) String() string {
	profile := GetBiomeProfile(bt)
	return profile.Name
}

// GetAllBiomeTypes returns all available biome types.
func GetAllBiomeTypes() []BiomeType {
	return []BiomeType{
		BiomeForestDungeon,
		BiomeDeepCavern,
		BiomeCrypt,
		BiomeLavaShrine,
		BiomeFrozenCaves,
		BiomeCrystalMine,
		BiomeAbandonedLab,
		BiomeAlienHive,
		BiomeCorpTower,
		BiomeWasteland,
		BiomeRadZone,
		BiomeUnderground,
	}
}

// ValidateMaterialID checks if a material ID is valid for any biome.
func ValidateMaterialID(materialID string) error {
	for _, biomeType := range GetAllBiomeTypes() {
		profile := GetBiomeProfile(biomeType)
		for _, mat := range profile.CommonMats {
			if mat.MaterialID == materialID {
				return nil
			}
		}
		for _, mat := range profile.UncommonMats {
			if mat.MaterialID == materialID {
				return nil
			}
		}
		for _, mat := range profile.RareMats {
			if mat.MaterialID == materialID {
				return nil
			}
		}
	}
	return fmt.Errorf("unknown material ID: %s", materialID)
}

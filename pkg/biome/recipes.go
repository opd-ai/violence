// Package biome provides biome material recipe integration.
package biome

import (
	"github.com/opd-ai/violence/pkg/crafting"
)

// BiomeMaterialRecipes returns crafting recipes that use biome-specific materials.
func BiomeMaterialRecipes(genreID string) []crafting.Recipe {
	switch genreID {
	case "fantasy":
		return []crafting.Recipe{
			// Forest materials
			{
				ID:        "ironbark_arrows",
				Name:      "Craft Ironbark Arrows",
				Inputs:    map[string]int{"ironbark": 2, "wood_scraps": 5},
				OutputID:  "ironbark_arrows",
				OutputQty: 15,
			},
			{
				ID:        "ethereal_potion",
				Name:      "Brew Ethereal Potion",
				Inputs:    map[string]int{"ethereal_sap": 1, "moss": 3},
				OutputID:  "ethereal_potion",
				OutputQty: 2,
			},
			{
				ID:        "ancient_staff",
				Name:      "Forge Ancient Staff",
				Inputs:    map[string]int{"ancient_root": 1, "ironbark": 3, "mana_crystal": 1},
				OutputID:  "ancient_staff",
				OutputQty: 1,
			},
			// Cave materials
			{
				ID:        "echo_grenade",
				Name:      "Craft Echo Grenade",
				Inputs:    map[string]int{"echo_crystals": 1, "stone_chunks": 3},
				OutputID:  "echo_grenade",
				OutputQty: 3,
			},
			{
				ID:        "diamond_blade",
				Name:      "Forge Diamond Blade",
				Inputs:    map[string]int{"deep_diamond": 1, "iron_ore": 4},
				OutputID:  "diamond_blade",
				OutputQty: 1,
			},
			// Crypt materials
			{
				ID:        "cursed_weapon",
				Name:      "Forge Cursed Weapon",
				Inputs:    map[string]int{"cursed_bone": 3, "spectral_essence": 1},
				OutputID:  "cursed_weapon",
				OutputQty: 1,
			},
			{
				ID:        "necromantic_scroll",
				Name:      "Inscribe Necromantic Scroll",
				Inputs:    map[string]int{"necromantic_core": 1, "bone_dust": 5, "tattered_cloth": 2},
				OutputID:  "necromantic_scroll",
				OutputQty: 1,
			},
			// Lava materials
			{
				ID:        "volcanic_bomb",
				Name:      "Craft Volcanic Bomb",
				Inputs:    map[string]int{"obsidian_shards": 2, "sulfur": 4},
				OutputID:  "volcanic_bomb",
				OutputQty: 4,
			},
			{
				ID:        "phoenix_amulet",
				Name:      "Forge Phoenix Amulet",
				Inputs:    map[string]int{"phoenix_ember": 1, "fire_opal": 2, "volcanic_glass": 1},
				OutputID:  "phoenix_amulet",
				OutputQty: 1,
			},
			// Frozen materials
			{
				ID:        "frost_arrows",
				Name:      "Craft Frost Arrows",
				Inputs:    map[string]int{"permafrost_crystal": 1, "ice_chunks": 4},
				OutputID:  "frost_arrows",
				OutputQty: 20,
			},
			{
				ID:        "eternal_armor",
				Name:      "Forge Eternal Ice Armor",
				Inputs:    map[string]int{"eternal_ice": 1, "winter_essence": 2, "frozen_bone": 3},
				OutputID:  "eternal_armor",
				OutputQty: 1,
			},
			// Crystal materials
			{
				ID:        "prismatic_wand",
				Name:      "Craft Prismatic Wand",
				Inputs:    map[string]int{"prismatic_shard": 2, "quartz_fragments": 4},
				OutputID:  "prismatic_wand",
				OutputQty: 1,
			},
			{
				ID:        "arcane_focus",
				Name:      "Craft Arcane Focus",
				Inputs:    map[string]int{"arcane_geode": 1, "mana_crystal": 2, "gem_dust": 5},
				OutputID:  "arcane_focus",
				OutputQty: 1,
			},
		}
	case "scifi":
		return []crafting.Recipe{
			// Lab materials
			{
				ID:        "bio_grenade",
				Name:      "Synthesize Bio Grenade",
				Inputs:    map[string]int{"bio_sample": 2, "chemical_vials": 3},
				OutputID:  "bio_grenade",
				OutputQty: 3,
			},
			{
				ID:        "prototype_weapon",
				Name:      "Assemble Prototype Weapon",
				Inputs:    map[string]int{"research_prototype": 1, "circuit_scrap": 5, "data_chip": 1},
				OutputID:  "prototype_weapon",
				OutputQty: 1,
			},
			// Hive materials
			{
				ID:        "chitin_armor",
				Name:      "Craft Chitin Armor",
				Inputs:    map[string]int{"chitin_plates": 6, "hive_resin": 2},
				OutputID:  "chitin_armor",
				OutputQty: 1,
			},
			{
				ID:        "xenomorph_serum",
				Name:      "Synthesize Xenomorph Serum",
				Inputs:    map[string]int{"xenomorph_ichor": 3, "alien_pheromone": 1},
				OutputID:  "xenomorph_serum",
				OutputQty: 2,
			},
			{
				ID:        "queen_implant",
				Name:      "Craft Queen Implant",
				Inputs:    map[string]int{"queen_essence": 1, "bio_mass": 8},
				OutputID:  "queen_implant",
				OutputQty: 1,
			},
			// Cave materials (scifi variant)
			{
				ID:        "echo_scanner",
				Name:      "Build Echo Scanner",
				Inputs:    map[string]int{"echo_crystals": 2, "circuit_scrap": 4},
				OutputID:  "echo_scanner",
				OutputQty: 1,
			},
			{
				ID:        "diamond_drill",
				Name:      "Craft Diamond Drill",
				Inputs:    map[string]int{"deep_diamond": 1, "wiring": 3},
				OutputID:  "diamond_drill",
				OutputQty: 1,
			},
			// Crystal materials (scifi variant)
			{
				ID:        "crystal_battery",
				Name:      "Craft Crystal Battery",
				Inputs:    map[string]int{"mana_crystal": 3, "circuit_scrap": 2},
				OutputID:  "crystal_battery",
				OutputQty: 5,
			},
		}
	case "horror":
		return []crafting.Recipe{
			// Crypt materials
			{
				ID:        "bone_weapon",
				Name:      "Craft Bone Weapon",
				Inputs:    map[string]int{"cursed_bone": 4, "bone_dust": 6},
				OutputID:  "bone_weapon",
				OutputQty: 1,
			},
			{
				ID:        "spectral_trap",
				Name:      "Create Spectral Trap",
				Inputs:    map[string]int{"spectral_essence": 2, "grave_soil": 4},
				OutputID:  "spectral_trap",
				OutputQty: 3,
			},
			{
				ID:        "necro_device",
				Name:      "Assemble Necromantic Device",
				Inputs:    map[string]int{"necromantic_core": 1, "cursed_bone": 5, "spectral_essence": 2},
				OutputID:  "necro_device",
				OutputQty: 1,
			},
			// Lab materials
			{
				ID:        "mutagen_serum",
				Name:      "Brew Mutagen Serum",
				Inputs:    map[string]int{"bio_sample": 3, "chemical_vials": 2},
				OutputID:  "mutagen_serum",
				OutputQty: 2,
			},
			{
				ID:        "experimental_tool",
				Name:      "Salvage Experimental Tool",
				Inputs:    map[string]int{"research_prototype": 1, "wiring": 3},
				OutputID:  "experimental_tool",
				OutputQty: 1,
			},
		}
	case "cyberpunk":
		return []crafting.Recipe{
			// Corp Tower materials
			{
				ID:        "hack_tool",
				Name:      "Build Hacking Tool",
				Inputs:    map[string]int{"encrypted_data": 1, "security_chip": 3},
				OutputID:  "hack_tool",
				OutputQty: 1,
			},
			{
				ID:        "cyberware_upgrade",
				Name:      "Craft Cyberware Upgrade",
				Inputs:    map[string]int{"cyberware_parts": 3, "office_equipment": 2},
				OutputID:  "cyberware_upgrade",
				OutputQty: 1,
			},
			{
				ID:        "corp_weapon",
				Name:      "Assemble Corporate Weapon",
				Inputs:    map[string]int{"corp_secrets": 1, "cyberware_parts": 4, "security_chip": 2},
				OutputID:  "corp_weapon",
				OutputQty: 1,
			},
			// Lab materials
			{
				ID:        "tech_implant",
				Name:      "Craft Tech Implant",
				Inputs:    map[string]int{"data_chip": 2, "bio_sample": 1, "circuit_scrap": 3},
				OutputID:  "tech_implant",
				OutputQty: 1,
			},
		}
	case "postapoc":
		return []crafting.Recipe{
			// Wasteland materials
			{
				ID:        "scrap_armor",
				Name:      "Craft Scrap Armor",
				Inputs:    map[string]int{"scrap_metal": 8, "rusty_parts": 4},
				OutputID:  "scrap_armor",
				OutputQty: 1,
			},
			{
				ID:        "salvaged_gun",
				Name:      "Assemble Salvaged Gun",
				Inputs:    map[string]int{"intact_electronics": 2, "scrap_metal": 5, "rusty_parts": 3},
				OutputID:  "salvaged_gun",
				OutputQty: 1,
			},
			{
				ID:        "prewar_device",
				Name:      "Restore Pre-War Device",
				Inputs:    map[string]int{"pre_war_material": 2, "intact_electronics": 1},
				OutputID:  "prewar_device",
				OutputQty: 1,
			},
			{
				ID:        "tech_relic",
				Name:      "Repair Old World Tech",
				Inputs:    map[string]int{"old_world_tech": 1, "scrap_metal": 6, "intact_electronics": 2},
				OutputID:  "tech_relic",
				OutputQty: 1,
			},
			// RadZone materials
			{
				ID:        "rad_bomb",
				Name:      "Craft Radiation Bomb",
				Inputs:    map[string]int{"glowing_isotope": 2, "radioactive_debris": 4},
				OutputID:  "rad_bomb",
				OutputQty: 3,
			},
			{
				ID:        "crystal_power_core",
				Name:      "Craft Crystal Power Core",
				Inputs:    map[string]int{"radiation_crystal": 2, "intact_electronics": 1},
				OutputID:  "crystal_power_core",
				OutputQty: 1,
			},
			{
				ID:        "plutonium_cell",
				Name:      "Refine Plutonium Cell",
				Inputs:    map[string]int{"pure_plutonium": 1, "radiation_crystal": 1},
				OutputID:  "plutonium_cell",
				OutputQty: 3,
			},
		}
	default:
		return []crafting.Recipe{}
	}
}

// AddBiomeMaterialRecipes adds biome material recipes to a crafting menu.
func AddBiomeMaterialRecipes(menu *crafting.CraftingMenu, genreID string) {
	// Note: This would require modifying CraftingMenu to have an AddRecipe method
	// For now, recipes are added through GetRecipes integration
}

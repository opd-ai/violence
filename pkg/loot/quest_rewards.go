// Package loot handles quest-driven reward generation.
package loot

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/rng"
)

// QuestRewardTier defines reward quality tiers for quest completion.
type QuestRewardTier int

const (
	TierStandard QuestRewardTier = iota
	TierBonus
	TierPerfect
	TierLegendary
)

// QuestReward represents loot granted for completing a quest objective.
type QuestReward struct {
	ItemID      string
	Quantity    int
	Rarity      Rarity
	Tier        QuestRewardTier
	Description string
}

// QuestRewardGenerator generates context-aware loot based on quest objectives.
type QuestRewardGenerator struct {
	genreID        string
	standardItems  map[string][]string // objective type -> item pool
	bonusItems     map[string][]string
	legendaryItems []string
	rng            *rng.RNG
}

// NewQuestRewardGenerator creates a quest reward generator with genre-specific loot pools.
func NewQuestRewardGenerator(genreID string, seed uint64) *QuestRewardGenerator {
	g := &QuestRewardGenerator{
		genreID:        genreID,
		standardItems:  make(map[string][]string),
		bonusItems:     make(map[string][]string),
		legendaryItems: []string{},
		rng:            rng.NewRNG(seed),
	}
	g.initializeGenrePools()
	return g
}

// initializeGenrePools sets up genre-specific item pools.
func (g *QuestRewardGenerator) initializeGenrePools() {
	switch g.genreID {
	case "fantasy":
		g.standardItems["exit"] = []string{"gold_coins_100", "health_potion", "mana_potion"}
		g.standardItems["enemy"] = []string{"sword_steel", "bow_longbow", "armor_chainmail"}
		g.standardItems["item"] = []string{"artifact_crystal", "tome_ancient", "ring_power"}
		g.standardItems["destroy"] = []string{"explosive_charge", "hammer_warhammer", "staff_lightning"}
		g.standardItems["time"] = []string{"boots_speed", "amulet_endurance", "shield_tower"}
		g.standardItems["retrieve"] = []string{"bag_holding", "cloak_invisibility", "key_master"}
		g.standardItems["hostage"] = []string{"blessing_courage", "scroll_teleport", "rope_enchanted"}

		g.bonusItems["secret"] = []string{"weapon_legendary_sword", "armor_dragonscale", "ring_wishes"}
		g.bonusItems["enemy"] = []string{"weapon_enchanted", "armor_blessed", "potion_ultimate"}
		g.bonusItems["time"] = []string{"boots_winged", "hourglass_time", "crown_haste"}

		g.legendaryItems = []string{
			"excalibur", "aegis_shield", "helm_of_kings",
			"staff_of_archmage", "bow_of_legends", "gauntlets_of_titans",
		}

	case "scifi":
		g.standardItems["exit"] = []string{"credits_500", "medkit", "energy_cell"}
		g.standardItems["enemy"] = []string{"plasma_rifle", "combat_armor_mk2", "shield_generator"}
		g.standardItems["item"] = []string{"data_core_alpha", "tech_scanner", "AI_module"}
		g.standardItems["destroy"] = []string{"emp_grenade", "fusion_cutter", "hacking_tool_advanced"}
		g.standardItems["time"] = []string{"stim_pack_military", "exo_servos", "temporal_anchor"}
		g.standardItems["retrieve"] = []string{"containment_field", "teleporter_beacon", "universal_key"}
		g.standardItems["hostage"] = []string{"medic_drone", "rescue_beacon", "extraction_charges"}

		g.bonusItems["secret"] = []string{"plasma_cannon_prototype", "nanoweave_armor", "quantum_core"}
		g.bonusItems["enemy"] = []string{"laser_designator", "targeting_implant", "power_armor"}
		g.bonusItems["time"] = []string{"chrono_accelerator", "stasis_field", "phase_boots"}

		g.legendaryItems = []string{
			"antimatter_rifle", "singularity_generator", "nanite_swarm_controller",
			"zero_point_module", "quantum_entangler", "dark_matter_core",
		}

	case "horror":
		g.standardItems["exit"] = []string{"sanity_pills", "first_aid", "blessed_water"}
		g.standardItems["enemy"] = []string{"silver_bullets", "crucifix_blessed", "salt_circle_kit"}
		g.standardItems["item"] = []string{"cursed_tome", "soul_gem", "ritual_dagger"}
		g.standardItems["destroy"] = []string{"holy_fire", "banishment_seal", "purification_powder"}
		g.standardItems["time"] = []string{"adrenaline_shot", "warding_charm", "spirit_anchor"}
		g.standardItems["retrieve"] = []string{"containment_box", "binding_chains", "protective_sigil"}
		g.standardItems["hostage"] = []string{"smelling_salts", "escape_rope", "light_flare"}

		g.bonusItems["secret"] = []string{"necronomicon_page", "elder_sign", "blood_moon_relic"}
		g.bonusItems["enemy"] = []string{"exorcism_kit", "demon_slayer_blade", "ward_of_ancients"}
		g.bonusItems["time"] = []string{"time_loop_device", "prophetic_vision", "hourglass_of_dread"}

		g.legendaryItems = []string{
			"book_of_the_dead", "sword_of_exorcism", "crown_of_madness",
			"eye_of_nightmares", "heart_of_darkness", "void_shard",
		}

	case "cyberpunk":
		g.standardItems["exit"] = []string{"cred_chips_1000", "synth_medkit", "battery_pack"}
		g.standardItems["enemy"] = []string{"smart_gun", "reflex_boosters", "dermal_plating"}
		g.standardItems["item"] = []string{"datachip_encrypted", "neural_interface", "black_ice_breaker"}
		g.standardItems["destroy"] = []string{"virus_bomb", "EMP_mine", "decryption_suite"}
		g.standardItems["time"] = []string{"adrenaline_chip", "combat_stims", "sandevistan_basic"}
		g.standardItems["retrieve"] = []string{"stealth_cloak", "data_siphon", "master_passkey"}
		g.standardItems["hostage"] = []string{"trauma_team_beacon", "smoke_grenades", "grapple_line"}

		g.bonusItems["secret"] = []string{"mantis_blades", "gorilla_arms", "projectile_launcher"}
		g.bonusItems["enemy"] = []string{"smart_link_mk3", "subdermal_armor", "berserk_implant"}
		g.bonusItems["time"] = []string{"sandevistan_military", "kerenzikov_reflex", "speedware_ultra"}

		g.legendaryItems = []string{
			"railgun_prototype", "full_body_conversion", "AI_companion_chip",
			"netrunner_deck_apex", "monowire_legendary", "cybereyes_omega",
		}

	case "postapoc":
		g.standardItems["exit"] = []string{"scrap_metal_50", "dirty_water", "rations_canned"}
		g.standardItems["enemy"] = []string{"pipe_rifle", "scrap_armor", "rad_pills"}
		g.standardItems["item"] = []string{"pre_war_tech", "fuel_canister", "ammo_stockpile"}
		g.standardItems["destroy"] = []string{"molotov", "improvised_explosive", "breaching_charge"}
		g.standardItems["time"] = []string{"stimpak", "gas_mask", "radiation_suit"}
		g.standardItems["retrieve"] = []string{"lockpicks_professional", "rope_climber", "pry_bar"}
		g.standardItems["hostage"] = []string{"med_supplies", "flare_gun", "crowbar"}

		g.bonusItems["secret"] = []string{"power_armor_frame", "energy_weapon_cache", "vehicle_parts"}
		g.bonusItems["enemy"] = []string{"military_rifle", "combat_armor_prewar", "mutation_serum"}
		g.bonusItems["time"] = []string{"jet_chem", "turbo_implant", "time_dilation_serum"}

		g.legendaryItems = []string{
			"gauss_rifle", "enclave_power_armor", "mini_nuke_launcher",
			"plasma_defender", "experimental_serum", "vault_tech_armor",
		}

	default:
		g.genreID = "fantasy"
		g.initializeGenrePools()
	}
}

// GenerateReward creates a quest reward based on objective type and completion tier.
func (g *QuestRewardGenerator) GenerateReward(objectiveType string, tier QuestRewardTier, seed uint64) QuestReward {
	localRNG := rng.NewRNG(seed)

	reward := QuestReward{
		Tier: tier,
	}

	switch tier {
	case TierStandard:
		reward.Rarity = RarityCommon
		pool := g.standardItems[objectiveType]
		if len(pool) > 0 {
			reward.ItemID = pool[localRNG.Intn(len(pool))]
			reward.Quantity = 1
		} else {
			reward.ItemID = "generic_reward"
			reward.Quantity = 1
		}
		reward.Description = g.genreText(
			"Basic quest reward",
			"Standard mission compensation",
			"Minimal survival supplies",
			"Contract payment",
			"Scavenged goods",
		)

	case TierBonus:
		reward.Rarity = RarityUncommon
		pool := g.bonusItems[objectiveType]
		if len(pool) > 0 {
			reward.ItemID = pool[localRNG.Intn(len(pool))]
			reward.Quantity = 1 + localRNG.Intn(2) // 1-2 items
		} else {
			reward.ItemID = g.selectFromStandard(objectiveType, localRNG)
			reward.Quantity = 2
		}
		reward.Description = g.genreText(
			"Enhanced quest reward",
			"Bonus mission loot",
			"Survivor's cache",
			"Premium contract bonus",
			"Quality salvage",
		)

	case TierPerfect:
		reward.Rarity = RarityRare
		pool := g.bonusItems[objectiveType]
		if len(pool) > 0 {
			reward.ItemID = pool[localRNG.Intn(len(pool))]
			reward.Quantity = 2 + localRNG.Intn(2) // 2-3 items
		} else {
			reward.ItemID = g.selectFromStandard(objectiveType, localRNG)
			reward.Quantity = 3
		}
		reward.Description = g.genreText(
			"Rare hero's bounty",
			"Perfect execution reward",
			"Elite survivor package",
			"Flawless run bonus",
			"Premium haul",
		)

	case TierLegendary:
		reward.Rarity = RarityLegendary
		if len(g.legendaryItems) > 0 {
			reward.ItemID = g.legendaryItems[localRNG.Intn(len(g.legendaryItems))]
			reward.Quantity = 1
		} else {
			reward.ItemID = g.selectFromBonus(objectiveType, localRNG)
			reward.Quantity = 1
		}
		reward.Description = g.genreText(
			"Legendary artifact of power",
			"Prototype experimental tech",
			"Relic of unspeakable horror",
			"Black market ultra-rare",
			"Pre-war treasure",
		)
	}

	return reward
}

// GenerateMultipleRewards creates a set of rewards for completing multiple objectives.
func (g *QuestRewardGenerator) GenerateMultipleRewards(objectives []ObjectiveRewardSpec, seed uint64) []QuestReward {
	rewards := make([]QuestReward, 0, len(objectives))

	for i, spec := range objectives {
		objSeed := seed + uint64(i)*1000
		reward := g.GenerateReward(spec.Type, spec.Tier, objSeed)
		rewards = append(rewards, reward)
	}

	// Add bonus legendary if all objectives perfect
	allPerfect := true
	for _, spec := range objectives {
		if spec.Tier != TierPerfect && spec.Tier != TierLegendary {
			allPerfect = false
			break
		}
	}

	if allPerfect && len(objectives) >= 3 {
		legendaryReward := g.GenerateReward("perfect", TierLegendary, seed+999999)
		legendaryReward.Description = g.genreText(
			"Perfect quest completion - Legendary reward!",
			"Flawless mission execution - Prototype acquired!",
			"Complete mastery - Ancient horror's power!",
			"All objectives dominated - Black market legend!",
			"Total survival - Pre-war masterpiece!",
		)
		rewards = append(rewards, legendaryReward)
	}

	return rewards
}

// ObjectiveRewardSpec defines the parameters for generating a reward.
type ObjectiveRewardSpec struct {
	Type string          // Objective type string (e.g., "enemy", "exit")
	Tier QuestRewardTier // Reward quality tier
}

func (g *QuestRewardGenerator) selectFromStandard(objectiveType string, r *rng.RNG) string {
	pool := g.standardItems[objectiveType]
	if len(pool) > 0 {
		return pool[r.Intn(len(pool))]
	}
	return "generic_reward"
}

func (g *QuestRewardGenerator) selectFromBonus(objectiveType string, r *rng.RNG) string {
	pool := g.bonusItems[objectiveType]
	if len(pool) > 0 {
		return pool[r.Intn(len(pool))]
	}
	return g.selectFromStandard(objectiveType, r)
}

func (g *QuestRewardGenerator) genreText(fantasy, scifi, horror, cyberpunk, postapoc string) string {
	switch g.genreID {
	case "scifi":
		return scifi
	case "horror":
		return horror
	case "cyberpunk":
		return cyberpunk
	case "postapoc":
		return postapoc
	default:
		return fantasy
	}
}

// DetermineTierFromObjective calculates reward tier based on objective completion quality.
func DetermineTierFromObjective(isMain bool, progress, count int, timeElapsed, timeTarget float64) QuestRewardTier {
	if count == 0 {
		return TierStandard
	}
	completionRatio := float64(progress) / float64(count)

	if isMain {
		// Main objectives give standard or bonus
		if completionRatio >= 1.0 {
			return TierStandard
		}
		return TierStandard
	}

	// Bonus objectives scale with performance
	if timeTarget > 0 && timeElapsed > 0 {
		// Time-based objective
		if timeElapsed <= timeTarget*0.5 {
			return TierLegendary // Completed in half the time
		} else if timeElapsed <= timeTarget*0.75 {
			return TierPerfect
		} else if timeElapsed <= timeTarget {
			return TierBonus
		}
		return TierStandard
	}

	// Count-based objective
	if completionRatio >= 2.0 {
		return TierLegendary // Doubled the requirement
	} else if completionRatio >= 1.5 {
		return TierPerfect
	} else if completionRatio >= 1.0 {
		return TierBonus
	}

	return TierStandard
}

// GetRewardDescription returns a formatted description of the reward.
func (r *QuestReward) GetRewardDescription() string {
	rarityName := "Common"
	switch r.Rarity {
	case RarityUncommon:
		rarityName = "Uncommon"
	case RarityRare:
		rarityName = "Rare"
	case RarityLegendary:
		rarityName = "Legendary"
	}

	if r.Quantity > 1 {
		return fmt.Sprintf("[%s] %s x%d - %s", rarityName, r.ItemID, r.Quantity, r.Description)
	}
	return fmt.Sprintf("[%s] %s - %s", rarityName, r.ItemID, r.Description)
}

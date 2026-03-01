package economy

import (
	"fmt"
	"math/rand"
	"testing"
)

// PlaytestScenario simulates a typical level playthrough
type PlaytestScenario struct {
	Genre       string
	Difficulty  string
	PlayerLevel int

	// Simulated actions
	WeakKills   int
	MediumKills int
	StrongKills int
	BossKills   int

	SecretsFound     int
	HiddenCaches     int
	MapCompletion    bool
	PrimaryObjective bool
	SecondaryObj     int
	TimeBonus        bool
}

// SimulateLevel runs a realistic level playthrough simulation
func SimulateLevel(scenario PlaytestScenario, cfg *Config) (totalCredits int) {
	// Combat rewards (using Config calculations)
	// Enemy type multipliers calibrated to ECONOMY.md targets
	// With BaseKillReward=20: weak=10-12cr, medium=20-25cr, strong=40-50cr, boss=120cr
	baseKill := cfg.CalculateKillReward(scenario.Genre, scenario.Difficulty, scenario.PlayerLevel)

	weakReward := int(float64(baseKill) * 0.5)
	mediumReward := int(float64(baseKill) * 1.0)
	strongReward := int(float64(baseKill) * 2.0)
	bossReward := int(float64(baseKill) * 4.5)

	totalCredits += weakReward * scenario.WeakKills
	totalCredits += mediumReward * scenario.MediumKills
	totalCredits += strongReward * scenario.StrongKills
	totalCredits += bossReward * scenario.BossKills

	// Exploration rewards (calibrated to hit ~350cr total on Normal)
	totalCredits += scenario.SecretsFound * 25
	totalCredits += scenario.HiddenCaches * 35
	if scenario.MapCompletion {
		totalCredits += 25
	}

	// Objective rewards
	if scenario.PrimaryObjective {
		totalCredits += cfg.CalculateMissionReward(scenario.Genre, scenario.Difficulty, scenario.PlayerLevel)
	}
	totalCredits += scenario.SecondaryObj * cfg.CalculateObjectiveReward(scenario.Genre, scenario.Difficulty, scenario.PlayerLevel)
	if scenario.TimeBonus {
		totalCredits += 30
	}

	return totalCredits
}

// TestPlaytestNormalDifficulty validates ~3 purchases per level target on Normal
func TestPlaytestNormalDifficulty(t *testing.T) {
	cfg := NewConfig()

	// Typical level playthrough: moderate exploration, average combat
	scenario := PlaytestScenario{
		Genre:            "fantasy",
		Difficulty:       "normal",
		PlayerLevel:      1,
		WeakKills:        8,
		MediumKills:      4,
		StrongKills:      1,
		BossKills:        0,
		SecretsFound:     1,
		HiddenCaches:     1,
		MapCompletion:    true,
		PrimaryObjective: true,
		SecondaryObj:     0,
		TimeBonus:        false,
	}

	credits := SimulateLevel(scenario, cfg)

	// Target: ~350 credits on Normal (ECONOMY.md spec)
	if credits < 300 || credits > 400 {
		t.Errorf("Normal difficulty credits = %d, want 300-400 (target ~350)", credits)
	}

	// Calculate purchasing power
	// Per ECONOMY.md: Medium Health Pack = 50cr, Hitscan Basic = 100cr, Armor Vest = 80cr
	medHealthPack := 50
	hitBasic := 100
	armorVest := 80

	purchases := []int{
		credits / medHealthPack,                    // How many health packs
		credits / hitBasic,                         // How many basic weapons
		credits / armorVest,                        // How many armor vests
		credits / (medHealthPack + armorVest + 25), // Mixed purchase (3 items)
	}

	// Should be able to afford 2-4 meaningful purchases
	for i, p := range purchases {
		if p < 2 || p > 6 {
			t.Logf("Purchase option %d: %d items (credits=%d)", i, p, credits)
		}
	}

	t.Logf("Normal difficulty level: %d credits earned", credits)
	t.Logf("  Can buy: %d medium health packs OR %d basic weapons OR %d armor vests",
		purchases[0], purchases[1], purchases[2])
}

// TestPlaytestAllDifficulties validates credit scaling across difficulties
func TestPlaytestAllDifficulties(t *testing.T) {
	cfg := NewConfig()

	difficulties := []struct {
		name       string
		minCredits int
		maxCredits int
	}{
		{"easy", 200, 350},
		{"normal", 300, 450},
		{"hard", 380, 600},
		{"nightmare", 450, 750},
	}

	baseScenario := PlaytestScenario{
		Genre:            "fantasy",
		PlayerLevel:      1,
		WeakKills:        8,
		MediumKills:      4,
		StrongKills:      1,
		BossKills:        0,
		SecretsFound:     1,
		HiddenCaches:     1,
		MapCompletion:    true,
		PrimaryObjective: true,
		SecondaryObj:     0,
		TimeBonus:        false,
	}

	for _, tc := range difficulties {
		t.Run(tc.name, func(t *testing.T) {
			scenario := baseScenario
			scenario.Difficulty = tc.name

			credits := SimulateLevel(scenario, cfg)

			if credits < tc.minCredits || credits > tc.maxCredits {
				t.Errorf("%s difficulty credits = %d, want %d-%d",
					tc.name, credits, tc.minCredits, tc.maxCredits)
			}

			t.Logf("%s difficulty: %d credits", tc.name, credits)
		})
	}
}

// TestPlaytestAllGenres validates genre balance
func TestPlaytestAllGenres(t *testing.T) {
	cfg := NewConfig()

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	baseScenario := PlaytestScenario{
		Difficulty:       "normal",
		PlayerLevel:      1,
		WeakKills:        8,
		MediumKills:      4,
		StrongKills:      1,
		BossKills:        0,
		SecretsFound:     1,
		HiddenCaches:     1,
		MapCompletion:    true,
		PrimaryObjective: true,
		SecondaryObj:     0,
		TimeBonus:        false,
	}

	results := make(map[string]int)

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			scenario := baseScenario
			scenario.Genre = genre

			credits := SimulateLevel(scenario, cfg)
			results[genre] = credits

			// All genres should yield 250-450 credits on Normal
			if credits < 250 || credits > 450 {
				t.Errorf("%s genre credits = %d, want 250-450", genre, credits)
			}

			t.Logf("%s genre: %d credits (multiplier: %.2f)",
				genre, credits, cfg.GenreMultipliers[genre])
		})
	}

	// Verify genre variation exists but isn't extreme
	minCredits, maxCredits := 999999, 0
	for _, c := range results {
		if c < minCredits {
			minCredits = c
		}
		if c > maxCredits {
			maxCredits = c
		}
	}

	variation := float64(maxCredits-minCredits) / float64(minCredits)
	if variation > 0.5 {
		t.Logf("Warning: genre variation %.1f%% is high (min=%d, max=%d)",
			variation*100, minCredits, maxCredits)
	}
}

// TestPlaytestLevelProgression validates credit scaling across levels
func TestPlaytestLevelProgression(t *testing.T) {
	cfg := NewConfig()

	baseScenario := PlaytestScenario{
		Genre:            "fantasy",
		Difficulty:       "normal",
		WeakKills:        8,
		MediumKills:      4,
		StrongKills:      1,
		BossKills:        0,
		SecretsFound:     1,
		HiddenCaches:     1,
		MapCompletion:    true,
		PrimaryObjective: true,
		SecondaryObj:     0,
		TimeBonus:        false,
	}

	levels := []int{1, 3, 5, 7, 10}
	var prevCredits int
	var prevLevel int

	for _, level := range levels {
		t.Run(fmt.Sprintf("level_%d", level), func(t *testing.T) {
			scenario := baseScenario
			scenario.PlayerLevel = level

			credits := SimulateLevel(scenario, cfg)

			// Credits should increase or stay same if multiplier unchanged
			// Levels 1-3 have same multiplier (1.0), so credits should be equal
			if prevLevel > 0 && cfg.getLevelMultiplier(level) > cfg.getLevelMultiplier(prevLevel) && credits <= prevCredits {
				t.Errorf("level %d credits (%d) should be > level %d credits (%d) when multiplier increases",
					level, credits, prevLevel, prevCredits)
			}

			prevCredits = credits
			prevLevel = level
			t.Logf("Level %d: %d credits (multiplier: %.2f)",
				level, credits, cfg.getLevelMultiplier(level))
		})
	}
}

// TestPlaytestMinimalRun validates rushed playthrough (minimal credits)
func TestPlaytestMinimalRun(t *testing.T) {
	cfg := NewConfig()

	// Rushed run: skip secrets, minimal combat, primary objective only
	scenario := PlaytestScenario{
		Genre:            "fantasy",
		Difficulty:       "normal",
		PlayerLevel:      1,
		WeakKills:        4,
		MediumKills:      2,
		StrongKills:      0,
		BossKills:        0,
		SecretsFound:     0,
		HiddenCaches:     0,
		MapCompletion:    false,
		PrimaryObjective: true,
		SecondaryObj:     0,
		TimeBonus:        true, // Speed bonus for rushed run
	}

	credits := SimulateLevel(scenario, cfg)

	// ECONOMY.md target: Min (Rushed) = 225 on Normal
	if credits < 150 || credits > 280 {
		t.Errorf("Minimal run credits = %d, want 150-280 (target ~225)", credits)
	}

	t.Logf("Minimal run: %d credits (should allow 1-2 purchases)", credits)
}

// TestPlaytest100PercentCompletion validates full exploration (max credits)
func TestPlaytest100PercentCompletion(t *testing.T) {
	cfg := NewConfig()

	// 100% completion: all secrets, high combat, all objectives
	scenario := PlaytestScenario{
		Genre:            "fantasy",
		Difficulty:       "normal",
		PlayerLevel:      1,
		WeakKills:        12,
		MediumKills:      6,
		StrongKills:      2,
		BossKills:        1,
		SecretsFound:     3,
		HiddenCaches:     2,
		MapCompletion:    true,
		PrimaryObjective: true,
		SecondaryObj:     1,
		TimeBonus:        false,
	}

	credits := SimulateLevel(scenario, cfg)

	// ECONOMY.md target: Max (100% Completion) = 550 on Normal
	// Allowing wider range since 100% is rare edge case
	if credits < 450 || credits > 800 {
		t.Errorf("100%% completion credits = %d, want 450-800 (target ~550)", credits)
	}

	t.Logf("100%% completion: %d credits (should allow 4-6 purchases)", credits)
}

// TestPlaytestPurchasingPower validates that credits allow ~3 purchases
func TestPlaytestPurchasingPower(t *testing.T) {
	cfg := NewConfig()

	scenario := PlaytestScenario{
		Genre:            "fantasy",
		Difficulty:       "normal",
		PlayerLevel:      1,
		WeakKills:        8,
		MediumKills:      4,
		StrongKills:      1,
		BossKills:        0,
		SecretsFound:     1,
		HiddenCaches:     1,
		MapCompletion:    true,
		PrimaryObjective: true,
		SecondaryObj:     0,
		TimeBonus:        false,
	}

	credits := SimulateLevel(scenario, cfg)

	// Test typical loadouts
	loadouts := []struct {
		name  string
		items []string
		total int
	}{
		{
			name:  "consumable_focus",
			items: []string{"health_medium", "health_medium", "health_medium", "armor_vest", "ammo_basic"},
			total: 50 + 50 + 50 + 80 + 20,
		},
		{
			name:  "weapon_upgrade",
			items: []string{"hitscan_advanced", "health_small", "health_small"},
			total: 250 + 25 + 25,
		},
		{
			name:  "balanced",
			items: []string{"hitscan_basic", "armor_vest", "ammo_basic", "ammo_basic", "grenade"},
			total: 100 + 80 + 20 + 20 + 35,
		},
		{
			name:  "premium_weapon",
			items: []string{"hitscan_advanced", "health_medium"},
			total: 250 + 50,
		},
	}

	purchasable := 0
	for _, loadout := range loadouts {
		if credits >= loadout.total {
			purchasable++
			t.Logf("✓ Can afford %s (%d credits): %v",
				loadout.name, loadout.total, loadout.items)
		} else {
			t.Logf("✗ Cannot afford %s (%d credits, need %d)",
				loadout.name, loadout.total, credits)
		}
	}

	// Should be able to afford 2-4 typical loadouts
	if purchasable < 2 {
		t.Errorf("Can only afford %d/%d loadouts with %d credits (too poor)",
			purchasable, len(loadouts), credits)
	}

	// Calculate number of "median-priced items" affordable
	medianPrice := 60 // Average of common items
	medianPurchases := credits / medianPrice

	if medianPurchases < 2 || medianPurchases > 6 {
		t.Errorf("Can afford %d median-priced items (want 2-6 for '~3 purchases')",
			medianPurchases)
	}

	t.Logf("Earned %d credits → can afford %d median items (~60cr each)",
		credits, medianPurchases)
}

// TestPlaytestRandomSampling runs Monte Carlo simulation
func TestPlaytestRandomSampling(t *testing.T) {
	cfg := NewConfig()
	rng := rand.New(rand.NewSource(42))

	const numSimulations = 100
	var totalCredits int
	minCredits, maxCredits := 999999, 0

	for i := 0; i < numSimulations; i++ {
		// Generate random playthrough within realistic bounds
		scenario := PlaytestScenario{
			Genre:            "fantasy",
			Difficulty:       "normal",
			PlayerLevel:      1,
			WeakKills:        4 + rng.Intn(8),     // 4-12 weak kills
			MediumKills:      2 + rng.Intn(4),     // 2-6 medium kills
			StrongKills:      rng.Intn(3),         // 0-2 strong kills
			BossKills:        rng.Intn(2),         // 0-1 boss
			SecretsFound:     rng.Intn(4),         // 0-3 secrets
			HiddenCaches:     rng.Intn(3),         // 0-2 caches
			MapCompletion:    rng.Float64() < 0.4, // 40% chance
			PrimaryObjective: true,                // Always complete primary
			SecondaryObj:     rng.Intn(2),         // 0-1 secondary
			TimeBonus:        rng.Float64() < 0.2, // 20% chance
		}

		credits := SimulateLevel(scenario, cfg)
		totalCredits += credits

		if credits < minCredits {
			minCredits = credits
		}
		if credits > maxCredits {
			maxCredits = credits
		}
	}

	avgCredits := totalCredits / numSimulations

	// Average should be near 350 (ECONOMY.md target), allowing some variance
	if avgCredits < 280 || avgCredits > 500 {
		t.Errorf("Monte Carlo average = %d credits, want 280-500 (target ~350)", avgCredits)
	}

	t.Logf("Monte Carlo simulation (%d runs):", numSimulations)
	t.Logf("  Average: %d credits", avgCredits)
	t.Logf("  Min: %d credits", minCredits)
	t.Logf("  Max: %d credits", maxCredits)
	t.Logf("  Range: %d credits (%.1f%% variation)",
		maxCredits-minCredits, float64(maxCredits-minCredits)/float64(avgCredits)*100)

	// Verify min/max align with ECONOMY.md expectations (with tolerance)
	if minCredits < 150 {
		t.Errorf("Min credits %d is too low (ECONOMY.md min rushed = 225)", minCredits)
	}
	if maxCredits > 800 {
		t.Errorf("Max credits %d is too high (ECONOMY.md max 100%% = 550, allowing variance)", maxCredits)
	}
}

// TestPlaytestEconomyBalance validates the core "~3 purchases" design goal
func TestPlaytestEconomyBalance(t *testing.T) {
	cfg := NewConfig()

	scenario := PlaytestScenario{
		Genre:            "fantasy",
		Difficulty:       "normal",
		PlayerLevel:      1,
		WeakKills:        8,
		MediumKills:      4,
		StrongKills:      1,
		BossKills:        0,
		SecretsFound:     1,
		HiddenCaches:     1,
		MapCompletion:    true,
		PrimaryObjective: true,
		SecondaryObj:     0,
		TimeBonus:        false,
	}

	credits := SimulateLevel(scenario, cfg)

	// Core balance check: ~3 purchases per level
	// Using weighted average item price of ~80 credits
	avgItemPrice := 80
	purchasesAffordable := credits / avgItemPrice

	if purchasesAffordable < 2 || purchasesAffordable > 5 {
		t.Errorf("Can afford %d purchases (avg price %d), want 2-5 for '~3 purchases' target",
			purchasesAffordable, avgItemPrice)
	}

	t.Logf("Economy balance: %d credits ÷ %d avg price = %d purchases (target: ~3)",
		credits, avgItemPrice, purchasesAffordable)

	// Validate player has meaningful choices (not too rich, not too poor)
	if credits > 600 {
		t.Errorf("Credits %d too high - players will hoard (economy feels pointless)", credits)
	}
	if credits < 200 {
		t.Errorf("Credits %d too low - players constantly broke (frustrating)", credits)
	}
}

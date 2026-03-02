package loot

import (
	"testing"
)

func TestNewQuestRewardGenerator(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    uint64
	}{
		{"fantasy generator", "fantasy", 12345},
		{"scifi generator", "scifi", 67890},
		{"horror generator", "horror", 11111},
		{"cyberpunk generator", "cyberpunk", 22222},
		{"postapoc generator", "postapoc", 33333},
		{"default fallback", "unknown", 99999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewQuestRewardGenerator(tt.genreID, tt.seed)
			if gen == nil {
				t.Fatal("NewQuestRewardGenerator returned nil")
			}
			if gen.rng == nil {
				t.Error("Generator RNG is nil")
			}
			if len(gen.standardItems) == 0 {
				t.Error("Standard items not initialized")
			}
			if len(gen.bonusItems) == 0 {
				t.Error("Bonus items not initialized")
			}
			if len(gen.legendaryItems) == 0 {
				t.Error("Legendary items not initialized")
			}

			// Check genre fallback
			expectedGenre := tt.genreID
			if tt.genreID == "unknown" {
				expectedGenre = "fantasy"
			}
			if gen.genreID != expectedGenre {
				t.Errorf("Expected genreID %s, got %s", expectedGenre, gen.genreID)
			}
		})
	}
}

func TestGenerateReward(t *testing.T) {
	tests := []struct {
		name          string
		genreID       string
		objectiveType string
		tier          QuestRewardTier
		seed          uint64
		expectRarity  Rarity
	}{
		{
			name:          "standard exit reward",
			genreID:       "fantasy",
			objectiveType: "exit",
			tier:          TierStandard,
			seed:          1000,
			expectRarity:  RarityCommon,
		},
		{
			name:          "bonus enemy reward",
			genreID:       "scifi",
			objectiveType: "enemy",
			tier:          TierBonus,
			seed:          2000,
			expectRarity:  RarityUncommon,
		},
		{
			name:          "perfect time reward",
			genreID:       "horror",
			objectiveType: "time",
			tier:          TierPerfect,
			seed:          3000,
			expectRarity:  RarityRare,
		},
		{
			name:          "legendary reward",
			genreID:       "cyberpunk",
			objectiveType: "item",
			tier:          TierLegendary,
			seed:          4000,
			expectRarity:  RarityLegendary,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewQuestRewardGenerator(tt.genreID, tt.seed)
			reward := gen.GenerateReward(tt.objectiveType, tt.tier, tt.seed)

			if reward.ItemID == "" {
				t.Error("Reward has empty ItemID")
			}
			if reward.Quantity <= 0 {
				t.Errorf("Invalid quantity: %d", reward.Quantity)
			}
			if reward.Rarity != tt.expectRarity {
				t.Errorf("Expected rarity %v, got %v", tt.expectRarity, reward.Rarity)
			}
			if reward.Tier != tt.tier {
				t.Errorf("Expected tier %v, got %v", tt.tier, reward.Tier)
			}
			if reward.Description == "" {
				t.Error("Reward has empty description")
			}
		})
	}
}

func TestGenerateRewardDeterminism(t *testing.T) {
	gen1 := NewQuestRewardGenerator("fantasy", 12345)
	gen2 := NewQuestRewardGenerator("fantasy", 12345)

	seed := uint64(5555)
	reward1 := gen1.GenerateReward("enemy", TierBonus, seed)
	reward2 := gen2.GenerateReward("enemy", TierBonus, seed)

	if reward1.ItemID != reward2.ItemID {
		t.Errorf("Non-deterministic item generation: %s vs %s", reward1.ItemID, reward2.ItemID)
	}
	if reward1.Quantity != reward2.Quantity {
		t.Errorf("Non-deterministic quantity: %d vs %d", reward1.Quantity, reward2.Quantity)
	}
}

func TestGenerateMultipleRewards(t *testing.T) {
	gen := NewQuestRewardGenerator("fantasy", 77777)

	specs := []ObjectiveRewardSpec{
		{Type: "exit", Tier: TierStandard},
		{Type: "enemy", Tier: TierBonus},
		{Type: "time", Tier: TierPerfect},
	}

	rewards := gen.GenerateMultipleRewards(specs, 88888)

	if len(rewards) != len(specs) {
		t.Errorf("Expected %d rewards, got %d", len(specs), len(rewards))
	}

	for i, reward := range rewards {
		if reward.ItemID == "" {
			t.Errorf("Reward %d has empty ItemID", i)
		}
		if reward.Tier != specs[i].Tier {
			t.Errorf("Reward %d tier mismatch: expected %v, got %v", i, specs[i].Tier, reward.Tier)
		}
	}
}

func TestGenerateMultipleRewardsPerfectBonus(t *testing.T) {
	gen := NewQuestRewardGenerator("scifi", 11111)

	// All perfect tier objectives
	specs := []ObjectiveRewardSpec{
		{Type: "exit", Tier: TierPerfect},
		{Type: "enemy", Tier: TierPerfect},
		{Type: "time", Tier: TierPerfect},
	}

	rewards := gen.GenerateMultipleRewards(specs, 22222)

	// Should get 3 regular rewards + 1 legendary bonus
	if len(rewards) != 4 {
		t.Errorf("Expected 4 rewards (3 + legendary bonus), got %d", len(rewards))
	}

	// Last reward should be legendary
	lastReward := rewards[len(rewards)-1]
	if lastReward.Rarity != RarityLegendary {
		t.Errorf("Perfect completion bonus should be legendary, got %v", lastReward.Rarity)
	}
}

func TestGenerateMultipleRewardsNoBonus(t *testing.T) {
	gen := NewQuestRewardGenerator("horror", 33333)

	// Mixed tiers, not all perfect
	specs := []ObjectiveRewardSpec{
		{Type: "exit", Tier: TierStandard},
		{Type: "enemy", Tier: TierBonus},
		{Type: "time", Tier: TierPerfect},
	}

	rewards := gen.GenerateMultipleRewards(specs, 44444)

	// Should get exactly 3 rewards, no legendary bonus
	if len(rewards) != 3 {
		t.Errorf("Expected 3 rewards (no bonus), got %d", len(rewards))
	}
}

func TestDetermineTierFromObjective(t *testing.T) {
	tests := []struct {
		name         string
		isMain       bool
		progress     int
		count        int
		timeElapsed  float64
		timeTarget   float64
		expectedTier QuestRewardTier
	}{
		{
			name:         "main objective standard",
			isMain:       true,
			progress:     10,
			count:        10,
			timeElapsed:  0,
			timeTarget:   0,
			expectedTier: TierStandard,
		},
		{
			name:         "bonus objective exceeded 2x",
			isMain:       false,
			progress:     20,
			count:        10,
			timeElapsed:  0,
			timeTarget:   0,
			expectedTier: TierLegendary,
		},
		{
			name:         "bonus objective exceeded 1.5x",
			isMain:       false,
			progress:     15,
			count:        10,
			timeElapsed:  0,
			timeTarget:   0,
			expectedTier: TierPerfect,
		},
		{
			name:         "bonus objective met",
			isMain:       false,
			progress:     10,
			count:        10,
			timeElapsed:  0,
			timeTarget:   0,
			expectedTier: TierBonus,
		},
		{
			name:         "time objective half time",
			isMain:       false,
			progress:     1,
			count:        1,
			timeElapsed:  50,
			timeTarget:   100,
			expectedTier: TierLegendary,
		},
		{
			name:         "time objective 75% time",
			isMain:       false,
			progress:     1,
			count:        1,
			timeElapsed:  70,
			timeTarget:   100,
			expectedTier: TierPerfect,
		},
		{
			name:         "time objective within limit",
			isMain:       false,
			progress:     1,
			count:        1,
			timeElapsed:  90,
			timeTarget:   100,
			expectedTier: TierBonus,
		},
		{
			name:         "time objective exceeded",
			isMain:       false,
			progress:     1,
			count:        1,
			timeElapsed:  110,
			timeTarget:   100,
			expectedTier: TierStandard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tier := DetermineTierFromObjective(tt.isMain, tt.progress, tt.count, tt.timeElapsed, tt.timeTarget)
			if tier != tt.expectedTier {
				t.Errorf("Expected tier %v, got %v", tt.expectedTier, tier)
			}
		})
	}
}

func TestQuestRewardGetRewardDescription(t *testing.T) {
	tests := []struct {
		name     string
		reward   QuestReward
		contains string
	}{
		{
			name: "common single item",
			reward: QuestReward{
				ItemID:      "health_potion",
				Quantity:    1,
				Rarity:      RarityCommon,
				Description: "Basic healing",
			},
			contains: "Common",
		},
		{
			name: "uncommon multiple items",
			reward: QuestReward{
				ItemID:      "mana_potion",
				Quantity:    3,
				Rarity:      RarityUncommon,
				Description: "Enhanced supplies",
			},
			contains: "x3",
		},
		{
			name: "legendary item",
			reward: QuestReward{
				ItemID:      "excalibur",
				Quantity:    1,
				Rarity:      RarityLegendary,
				Description: "Sword of legend",
			},
			contains: "Legendary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.reward.GetRewardDescription()
			if desc == "" {
				t.Error("GetRewardDescription returned empty string")
			}
			if len(desc) < 10 {
				t.Errorf("Description too short: %s", desc)
			}
			// Basic validation that contains key info
			found := false
			if tt.contains != "" {
				for i := 0; i < len(desc)-len(tt.contains)+1; i++ {
					if desc[i:i+len(tt.contains)] == tt.contains {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Description missing expected content '%s': %s", tt.contains, desc)
				}
			}
		})
	}
}

func TestGenreSpecificItemPools(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			gen := NewQuestRewardGenerator(genre, 12345)

			// Check that standard pools are populated
			if len(gen.standardItems) == 0 {
				t.Error("Standard items empty")
			}

			// Check that bonus pools are populated
			if len(gen.bonusItems) == 0 {
				t.Error("Bonus items empty")
			}

			// Check that legendary items exist
			if len(gen.legendaryItems) == 0 {
				t.Error("Legendary items empty")
			}

			// Verify we can generate rewards
			reward := gen.GenerateReward("enemy", TierBonus, 54321)
			if reward.ItemID == "" {
				t.Error("Failed to generate reward for", genre)
			}
		})
	}
}

func TestRewardQuantityScaling(t *testing.T) {
	gen := NewQuestRewardGenerator("fantasy", 99999)

	standardReward := gen.GenerateReward("enemy", TierStandard, 1111)
	bonusReward := gen.GenerateReward("enemy", TierBonus, 2222)
	perfectReward := gen.GenerateReward("enemy", TierPerfect, 3333)

	// Higher tiers should generally give more items (or better quality)
	if standardReward.Quantity <= 0 {
		t.Error("Standard reward has invalid quantity")
	}
	if bonusReward.Quantity <= 0 {
		t.Error("Bonus reward has invalid quantity")
	}
	if perfectReward.Quantity <= 0 {
		t.Error("Perfect reward has invalid quantity")
	}

	// Verify rarity increases with tier
	if int(standardReward.Rarity) >= int(bonusReward.Rarity) {
		t.Error("Standard rarity should be lower than bonus")
	}
	if int(bonusReward.Rarity) >= int(perfectReward.Rarity) {
		t.Error("Bonus rarity should be lower than perfect")
	}
}

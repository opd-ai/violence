package loot

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewQuestLootSystem(t *testing.T) {
	system := NewQuestLootSystem("fantasy", 12345)
	if system == nil {
		t.Fatal("NewQuestLootSystem returned nil")
	}
	if system.generator == nil {
		t.Error("System generator is nil")
	}
	if system.logger == nil {
		t.Error("System logger is nil")
	}
	if !system.enabled {
		t.Error("System should be enabled by default")
	}
	if system.genreID != "fantasy" {
		t.Errorf("Expected genreID 'fantasy', got '%s'", system.genreID)
	}
}

func TestQuestLootSystemUpdate(t *testing.T) {
	system := NewQuestLootSystem("scifi", 67890)

	// Create a simple world for testing
	world := &engine.World{}

	// Update should not panic
	system.Update(world)

	// Disabled system should also not panic
	system.SetEnabled(false)
	system.Update(world)
}

func TestQuestLootSystemGrantRewardForObjective(t *testing.T) {
	system := NewQuestLootSystem("fantasy", 11111)

	tests := []struct {
		name          string
		objectiveType string
		isMain        bool
		progress      int
		count         int
		timeElapsed   float64
		timeTarget    float64
	}{
		{
			name:          "main objective completion",
			objectiveType: "exit",
			isMain:        true,
			progress:      1,
			count:         1,
			timeElapsed:   0,
			timeTarget:    0,
		},
		{
			name:          "bonus kill objective exceeded",
			objectiveType: "enemy",
			isMain:        false,
			progress:      30,
			count:         20,
			timeElapsed:   0,
			timeTarget:    0,
		},
		{
			name:          "speedrun objective perfect",
			objectiveType: "time",
			isMain:        false,
			progress:      1,
			count:         1,
			timeElapsed:   60,
			timeTarget:    180,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward := system.GrantRewardForObjective(
				tt.objectiveType,
				tt.isMain,
				tt.progress,
				tt.count,
				tt.timeElapsed,
				tt.timeTarget,
				22222,
			)

			if reward.ItemID == "" {
				t.Error("Reward has empty ItemID")
			}
			if reward.Quantity <= 0 {
				t.Errorf("Invalid quantity: %d", reward.Quantity)
			}
			if reward.Description == "" {
				t.Error("Reward has empty description")
			}
		})
	}
}

func TestQuestLootSystemGrantRewardsForMultipleObjectives(t *testing.T) {
	system := NewQuestLootSystem("horror", 33333)

	specs := []ObjectiveRewardSpec{
		{Type: "exit", Tier: TierStandard},
		{Type: "enemy", Tier: TierBonus},
		{Type: "secret", Tier: TierPerfect},
	}

	rewards := system.GrantRewardsForMultipleObjectives(specs, 44444)

	if len(rewards) < len(specs) {
		t.Errorf("Expected at least %d rewards, got %d", len(specs), len(rewards))
	}

	for i, reward := range rewards {
		if reward.ItemID == "" {
			t.Errorf("Reward %d has empty ItemID", i)
		}
	}
}

func TestQuestLootSystemSetEnabled(t *testing.T) {
	system := NewQuestLootSystem("cyberpunk", 55555)

	if !system.enabled {
		t.Error("System should start enabled")
	}

	system.SetEnabled(false)
	if system.enabled {
		t.Error("SetEnabled(false) did not disable system")
	}

	system.SetEnabled(true)
	if !system.enabled {
		t.Error("SetEnabled(true) did not enable system")
	}
}

func TestQuestLootSystemSetGenre(t *testing.T) {
	system := NewQuestLootSystem("fantasy", 66666)

	if system.genreID != "fantasy" {
		t.Error("Initial genre not set correctly")
	}

	system.SetGenre("scifi", 77777)
	if system.genreID != "scifi" {
		t.Error("SetGenre did not update genreID")
	}

	// Verify generator was reinitialized
	reward := system.GrantRewardForObjective("enemy", true, 1, 1, 0, 0, 88888)
	if reward.ItemID == "" {
		t.Error("Generator not working after SetGenre")
	}
}

func TestQuestLootComponentType(t *testing.T) {
	component := &QuestLootComponent{}
	if component.Type() != "QuestLoot" {
		t.Errorf("Expected component type 'QuestLoot', got '%s'", component.Type())
	}
}

func TestQuestLootSystemIntegration(t *testing.T) {
	// Test a full quest completion flow
	system := NewQuestLootSystem("fantasy", 99999)

	// Simulate completing multiple objectives with varying performance
	specs := []ObjectiveRewardSpec{
		{Type: "exit", Tier: TierStandard},  // Found exit
		{Type: "enemy", Tier: TierPerfect},  // Killed 1.5x enemies
		{Type: "time", Tier: TierLegendary}, // Speedrun in half time
		{Type: "secret", Tier: TierPerfect}, // Found all secrets
	}

	rewards := system.GrantRewardsForMultipleObjectives(specs, 11111)

	// Should get base rewards plus potential perfect completion bonus
	if len(rewards) < 4 {
		t.Errorf("Expected at least 4 rewards, got %d", len(rewards))
	}

	// Check that we got at least one legendary
	hasLegendary := false
	for _, reward := range rewards {
		if reward.Rarity == RarityLegendary {
			hasLegendary = true
			break
		}
	}
	if !hasLegendary {
		t.Error("Expected at least one legendary reward for excellent performance")
	}
}

func TestQuestLootSystemDeterminism(t *testing.T) {
	// Same seed and parameters should produce same rewards
	system1 := NewQuestLootSystem("fantasy", 12345)
	system2 := NewQuestLootSystem("fantasy", 12345)

	seed := uint64(99999)
	reward1 := system1.GrantRewardForObjective("enemy", false, 20, 10, 0, 0, seed)
	reward2 := system2.GrantRewardForObjective("enemy", false, 20, 10, 0, 0, seed)

	if reward1.ItemID != reward2.ItemID {
		t.Errorf("Non-deterministic rewards: %s vs %s", reward1.ItemID, reward2.ItemID)
	}
	if reward1.Quantity != reward2.Quantity {
		t.Errorf("Non-deterministic quantities: %d vs %d", reward1.Quantity, reward2.Quantity)
	}
}

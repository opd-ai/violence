// Package main demonstrates the quest-loot integration system.
package main

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/loot"
)

func main() {
	fmt.Println("=== Quest-Loot Integration System Demo ===")
	fmt.Println()

	// Create a quest reward system for a fantasy game
	system := loot.NewQuestLootSystem("fantasy", 12345)

	fmt.Println("1. Main Quest Objective Completed (Find Exit):")
	reward1 := system.GrantRewardForObjective("exit", true, 1, 1, 0, 0, 1000)
	fmt.Printf("   %s\n\n", reward1.GetRewardDescription())

	fmt.Println("2. Bonus Objective Completed (Kill 30/20 enemies - exceeded):")
	reward2 := system.GrantRewardForObjective("enemy", false, 30, 20, 0, 0, 2000)
	fmt.Printf("   %s\n\n", reward2.GetRewardDescription())

	fmt.Println("3. Speedrun Objective Perfect (60s / 180s target - legendary):")
	reward3 := system.GrantRewardForObjective("time", false, 1, 1, 60.0, 180.0, 3000)
	fmt.Printf("   %s\n\n", reward3.GetRewardDescription())

	fmt.Println("4. All Secrets Found (5/5):")
	reward4 := system.GrantRewardForObjective("secret", false, 5, 5, 0, 0, 4000)
	fmt.Printf("   %s\n\n", reward4.GetRewardDescription())

	fmt.Println("5. Perfect Quest Completion (all objectives at perfect/legendary tier):")
	specs := []loot.ObjectiveRewardSpec{
		{Type: "exit", Tier: loot.TierPerfect},
		{Type: "enemy", Tier: loot.TierPerfect},
		{Type: "time", Tier: loot.TierLegendary},
	}
	rewards := system.GrantRewardsForMultipleObjectives(specs, 5000)
	for i, r := range rewards {
		fmt.Printf("   Reward %d: %s\n", i+1, r.GetRewardDescription())
	}
	fmt.Println()

	// Demonstrate genre switching
	fmt.Println("6. Switching to Sci-Fi genre:")
	system.SetGenre("scifi", 67890)
	rewardScifi := system.GrantRewardForObjective("enemy", false, 40, 20, 0, 0, 6000)
	fmt.Printf("   %s\n\n", rewardScifi.GetRewardDescription())

	fmt.Println("=== Demo Complete ===")
}

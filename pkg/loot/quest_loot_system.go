// Package loot provides quest-loot integration system for ECS.
package loot

import (
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// QuestLootComponent tracks quest-related loot state for an entity.
type QuestLootComponent struct {
	PendingRewards []QuestReward
	TotalRewards   int
	LastQuestID    string
}

// Type returns the component type identifier.
func (c *QuestLootComponent) Type() string {
	return "QuestLoot"
}

// QuestLootSystem manages quest completion rewards in the ECS.
type QuestLootSystem struct {
	logger    *logrus.Entry
	generator *QuestRewardGenerator
	genreID   string
	enabled   bool
}

// NewQuestLootSystem creates a new quest-loot integration system.
func NewQuestLootSystem(genreID string, seed uint64) *QuestLootSystem {
	return &QuestLootSystem{
		logger: logrus.WithFields(logrus.Fields{
			"system": "quest_loot",
			"genre":  genreID,
		}),
		generator: NewQuestRewardGenerator(genreID, seed),
		genreID:   genreID,
		enabled:   true,
	}
}

// Update processes quest completion events and generates rewards.
func (s *QuestLootSystem) Update(w *engine.World) {
	if !s.enabled {
		return
	}

	// System processes quest completion events externally
	// This is a passive system that provides the API for quest→loot
}

// GrantRewardForObjective generates and grants a reward for completing a quest objective.
func (s *QuestLootSystem) GrantRewardForObjective(
	objectiveType string,
	isMain bool,
	progress, count int,
	timeElapsed, timeTarget float64,
	seed uint64,
) QuestReward {
	tier := DetermineTierFromObjective(isMain, progress, count, timeElapsed, timeTarget)
	reward := s.generator.GenerateReward(objectiveType, tier, seed)

	s.logger.WithFields(logrus.Fields{
		"objective_type": objectiveType,
		"tier":           tier,
		"item":           reward.ItemID,
		"quantity":       reward.Quantity,
		"rarity":         reward.Rarity,
	}).Info("Quest objective completed, reward granted")

	return reward
}

// GrantRewardsForMultipleObjectives generates rewards for completing multiple objectives.
func (s *QuestLootSystem) GrantRewardsForMultipleObjectives(specs []ObjectiveRewardSpec, seed uint64) []QuestReward {
	rewards := s.generator.GenerateMultipleRewards(specs, seed)

	s.logger.WithFields(logrus.Fields{
		"objective_count": len(specs),
		"reward_count":    len(rewards),
	}).Info("Multiple quest objectives completed")

	return rewards
}

// AddPendingReward adds a reward to an entity's pending queue.
func (s *QuestLootSystem) AddPendingReward(entity interface{}, reward QuestReward) {
	// This is a helper for external integration
	// In a full ECS, we'd iterate entities and add to QuestLootComponent
}

// SetEnabled enables or disables the quest loot system.
func (s *QuestLootSystem) SetEnabled(enabled bool) {
	s.enabled = enabled
	s.logger.WithField("enabled", enabled).Info("Quest loot system state changed")
}

// SetGenre updates the genre and reinitializes loot pools.
func (s *QuestLootSystem) SetGenre(genreID string, seed uint64) {
	s.genreID = genreID
	s.generator = NewQuestRewardGenerator(genreID, seed)
	s.logger.WithField("genre", genreID).Info("Quest loot genre updated")
}

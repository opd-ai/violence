// Package faction provides faction reputation tracking and relationship management.
//
// The faction system integrates with the economy (shop prices), AI behavior (NPC hostility),
// and quest system (quest availability and rewards) to create meaningful player choices.
//
// Example usage:
//
//	// Initialize faction system (done in main.go NewGame)
//	factionSys := faction.NewReputationSystem()
//	world.AddSystem(factionSys)
//
//	// Initialize player reputation
//	rep := faction.InitializePlayerReputation("fantasy")
//	world.AddComponent(playerEntity, rep)
//
//	// Modify reputation when player kills faction member
//	faction.ApplyEnemyKillReputation(factionSys, rep, faction.FactionMercenaries)
//
//	// Apply faction discount to shop prices
//	priceModifier := faction.GetFactionShopPriceModifier(rep, faction.FactionMercenaries)
//	shop.PurchaseWithModifier(itemID, credits, priceModifier)
//
//	// Check if NPC should be hostile
//	if faction.ShouldAttackPlayer(rep, npcFaction, rng) {
//	    // Engage combat
//	}
//
//	// Generate faction quest
//	quest := factionSys.GenerateFactionQuest(faction.FactionRebels, rng, "cyberpunk")
package faction

import (
	"math"
	"reflect"
	"sync"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// FactionID represents unique faction identifiers.
type FactionID string

const (
	FactionMercenaries FactionID = "mercenaries"
	FactionTechCorp    FactionID = "techcorp"
	FactionCult        FactionID = "cult"
	FactionRebels      FactionID = "rebels"
	FactionSyndicate   FactionID = "syndicate"
)

// Standing represents reputation levels with a faction.
type Standing int

const (
	StandingHated    Standing = -3
	StandingHostile  Standing = -2
	StandingUnfriend Standing = -1
	StandingNeutral  Standing = 0
	StandingFriendly Standing = 1
	StandingRespect  Standing = 2
	StandingExalted  Standing = 3
)

// String returns standing name.
func (s Standing) String() string {
	switch s {
	case StandingHated:
		return "Hated"
	case StandingHostile:
		return "Hostile"
	case StandingUnfriend:
		return "Unfriendly"
	case StandingNeutral:
		return "Neutral"
	case StandingFriendly:
		return "Friendly"
	case StandingRespect:
		return "Respected"
	case StandingExalted:
		return "Exalted"
	default:
		return "Unknown"
	}
}

// Faction represents a game faction with its properties.
type Faction struct {
	ID          FactionID
	Name        string
	Description string
	Enemies     []FactionID
	Allies      []FactionID
	GenreID     string
}

// ReputationComponent stores player reputation with factions.
type ReputationComponent struct {
	Scores map[FactionID]int
}

// Type returns component type identifier.
func (r *ReputationComponent) Type() string {
	return "ReputationComponent"
}

// FactionMemberComponent marks an entity as belonging to a faction.
type FactionMemberComponent struct {
	FactionID FactionID
	Rank      int
}

// Type returns component type identifier.
func (f *FactionMemberComponent) Type() string {
	return "FactionMemberComponent"
}

// ReputationSystem manages faction reputation and relationships.
type ReputationSystem struct {
	factions    map[FactionID]*Faction
	genreConfig map[string][]FactionID
	mu          sync.RWMutex
}

// NewReputationSystem creates a new faction reputation system.
func NewReputationSystem() *ReputationSystem {
	sys := &ReputationSystem{
		factions:    make(map[FactionID]*Faction),
		genreConfig: make(map[string][]FactionID),
	}
	sys.initializeFactions()
	return sys
}

func (s *ReputationSystem) initializeFactions() {
	s.factions[FactionMercenaries] = &Faction{
		ID:          FactionMercenaries,
		Name:        "Iron Brotherhood",
		Description: "Professional mercenaries who value contracts and coin",
		Enemies:     []FactionID{FactionCult},
		Allies:      []FactionID{FactionSyndicate},
		GenreID:     "any",
	}

	s.factions[FactionTechCorp] = &Faction{
		ID:          FactionTechCorp,
		Name:        "NeuroCorp",
		Description: "Tech conglomerate seeking cutting-edge artifacts",
		Enemies:     []FactionID{FactionRebels},
		Allies:      []FactionID{},
		GenreID:     "cyberpunk",
	}

	s.factions[FactionCult] = &Faction{
		ID:          FactionCult,
		Name:        "Crimson Covenant",
		Description: "Dark cultists pursuing forbidden knowledge",
		Enemies:     []FactionID{FactionMercenaries, FactionRebels},
		Allies:      []FactionID{},
		GenreID:     "horror",
	}

	s.factions[FactionRebels] = &Faction{
		ID:          FactionRebels,
		Name:        "The Uprising",
		Description: "Freedom fighters resisting oppression",
		Enemies:     []FactionID{FactionTechCorp, FactionCult},
		Allies:      []FactionID{},
		GenreID:     "any",
	}

	s.factions[FactionSyndicate] = &Faction{
		ID:          FactionSyndicate,
		Name:        "Shadow Syndicate",
		Description: "Criminal network controlling the underworld",
		Enemies:     []FactionID{},
		Allies:      []FactionID{FactionMercenaries},
		GenreID:     "any",
	}

	s.genreConfig["fantasy"] = []FactionID{FactionMercenaries, FactionCult, FactionRebels}
	s.genreConfig["cyberpunk"] = []FactionID{FactionTechCorp, FactionRebels, FactionSyndicate}
	s.genreConfig["horror"] = []FactionID{FactionCult, FactionMercenaries, FactionRebels}
	s.genreConfig["scifi"] = []FactionID{FactionTechCorp, FactionRebels, FactionSyndicate}
}

// GetActiveFactions returns factions available for a given genre.
func (s *ReputationSystem) GetActiveFactions(genreID string) []*Faction {
	s.mu.RLock()
	defer s.mu.RUnlock()

	factionIDs := s.genreConfig[genreID]
	if len(factionIDs) == 0 {
		factionIDs = s.genreConfig["fantasy"]
	}

	result := make([]*Faction, 0, len(factionIDs))
	for _, id := range factionIDs {
		if f, ok := s.factions[id]; ok {
			result = append(result, f)
		}
	}
	return result
}

// GetFaction retrieves faction by ID.
func (s *ReputationSystem) GetFaction(id FactionID) *Faction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.factions[id]
}

// GetStanding converts reputation score to standing level.
func GetStanding(score int) Standing {
	switch {
	case score >= 6000:
		return StandingExalted
	case score >= 3000:
		return StandingRespect
	case score >= 1000:
		return StandingFriendly
	case score >= -1000:
		return StandingNeutral
	case score >= -3000:
		return StandingUnfriend
	case score >= -6000:
		return StandingHostile
	default:
		return StandingHated
	}
}

// GetPriceModifier returns shop price multiplier based on standing.
func GetPriceModifier(standing Standing) float64 {
	switch standing {
	case StandingExalted:
		return 0.7
	case StandingRespect:
		return 0.85
	case StandingFriendly:
		return 0.95
	case StandingNeutral:
		return 1.0
	case StandingUnfriend:
		return 1.15
	case StandingHostile:
		return 1.35
	case StandingHated:
		return 1.6
	default:
		return 1.0
	}
}

// GetAggressionModifier returns AI hostility multiplier based on standing.
func GetAggressionModifier(standing Standing) float64 {
	switch standing {
	case StandingExalted, StandingRespect:
		return 0.0
	case StandingFriendly:
		return 0.3
	case StandingNeutral:
		return 1.0
	case StandingUnfriend:
		return 1.3
	case StandingHostile:
		return 1.6
	case StandingHated:
		return 2.0
	default:
		return 1.0
	}
}

// Update processes reputation changes and faction relationships.
func (s *ReputationSystem) Update(w *engine.World) {
	repType := reflect.TypeOf((*ReputationComponent)(nil))
	entities := w.Query(repType)

	for _, ent := range entities {
		repComp, ok := w.GetComponent(ent, repType)
		if !ok {
			continue
		}

		rep, ok := repComp.(*ReputationComponent)
		if !ok {
			continue
		}

		s.processReputationDecay(rep)
	}
}

func (s *ReputationSystem) processReputationDecay(rep *ReputationComponent) {
	for factionID, score := range rep.Scores {
		if score > 0 {
			rep.Scores[factionID] = int(math.Max(0, float64(score)-0.1))
		} else if score < 0 {
			rep.Scores[factionID] = int(math.Min(0, float64(score)+0.1))
		}
	}
}

// ModifyReputation changes player reputation with a faction and propagates to allies/enemies.
func (s *ReputationSystem) ModifyReputation(rep *ReputationComponent, factionID FactionID, delta int) {
	if rep.Scores == nil {
		rep.Scores = make(map[FactionID]int)
	}

	s.mu.RLock()
	faction := s.factions[factionID]
	s.mu.RUnlock()

	if faction == nil {
		return
	}

	oldScore := rep.Scores[factionID]
	newScore := clamp(oldScore+delta, -10000, 10000)
	rep.Scores[factionID] = newScore

	oldStanding := GetStanding(oldScore)
	newStanding := GetStanding(newScore)

	if oldStanding != newStanding {
		logrus.WithFields(logrus.Fields{
			"system":       "faction",
			"faction":      faction.Name,
			"old_standing": oldStanding.String(),
			"new_standing": newStanding.String(),
		}).Info("Faction standing changed")
	}

	spilloverAmount := int(float64(delta) * 0.25)
	for _, allyID := range faction.Allies {
		rep.Scores[allyID] = clamp(rep.Scores[allyID]+spilloverAmount, -10000, 10000)
	}

	penaltyAmount := int(float64(delta) * 0.5)
	for _, enemyID := range faction.Enemies {
		rep.Scores[enemyID] = clamp(rep.Scores[enemyID]-penaltyAmount, -10000, 10000)
	}
}

// GetReputation returns current reputation score with a faction.
func (s *ReputationSystem) GetReputation(rep *ReputationComponent, factionID FactionID) int {
	if rep.Scores == nil {
		return 0
	}
	return rep.Scores[factionID]
}

// CanInteract returns whether an entity can interact based on faction relations.
func (s *ReputationSystem) CanInteract(rep *ReputationComponent, factionID FactionID, minStanding Standing) bool {
	score := s.GetReputation(rep, factionID)
	return GetStanding(score) >= minStanding
}

// GenerateFactionQuest creates a faction-specific quest with reputation rewards.
func (s *ReputationSystem) GenerateFactionQuest(factionID FactionID, rng *rng.RNG, genreID string) *FactionQuest {
	s.mu.RLock()
	faction := s.factions[factionID]
	s.mu.RUnlock()

	if faction == nil {
		return nil
	}

	questTypes := []string{"eliminate", "retrieve", "escort", "sabotage"}
	questType := questTypes[rng.Intn(len(questTypes))]

	rewards := s.calculateQuestRewards(factionID, questType, rng)

	return &FactionQuest{
		FactionID:       factionID,
		Type:            questType,
		RepReward:       rewards.rep,
		CreditReward:    rewards.credits,
		RequireStanding: StandingNeutral,
	}
}

type questRewards struct {
	rep     int
	credits int
}

func (s *ReputationSystem) calculateQuestRewards(factionID FactionID, questType string, rng *rng.RNG) questRewards {
	baseRep := 200
	baseCredits := 500

	multiplier := 1.0
	switch questType {
	case "eliminate":
		multiplier = 1.2
	case "sabotage":
		multiplier = 1.5
	case "escort":
		multiplier = 0.8
	case "retrieve":
		multiplier = 1.0
	}

	variance := 0.8 + rng.Float64()*0.4

	return questRewards{
		rep:     int(float64(baseRep) * multiplier * variance),
		credits: int(float64(baseCredits) * multiplier * variance),
	}
}

// FactionQuest represents a quest offered by a faction.
type FactionQuest struct {
	FactionID       FactionID
	Type            string
	RepReward       int
	CreditReward    int
	RequireStanding Standing
}

// InitializePlayerReputation creates a reputation component for a new player.
func InitializePlayerReputation(genreID string) *ReputationComponent {
	rep := &ReputationComponent{
		Scores: make(map[FactionID]int),
	}

	allFactions := []FactionID{
		FactionMercenaries,
		FactionTechCorp,
		FactionCult,
		FactionRebels,
		FactionSyndicate,
	}

	for _, factionID := range allFactions {
		rep.Scores[factionID] = 0
	}

	return rep
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// GetFactionShopPriceModifier returns price modifier for a shop owned by a faction.
func GetFactionShopPriceModifier(rep *ReputationComponent, factionID FactionID) float64 {
	if rep == nil || rep.Scores == nil {
		return 1.0
	}
	score := rep.Scores[factionID]
	standing := GetStanding(score)
	return GetPriceModifier(standing)
}

// ShouldAttackPlayer determines if faction member should attack based on reputation.
func ShouldAttackPlayer(rep *ReputationComponent, factionID FactionID, rng *rng.RNG) bool {
	if rep == nil || rep.Scores == nil {
		return false
	}

	score := rep.Scores[factionID]
	standing := GetStanding(score)

	if standing >= StandingFriendly {
		return false
	}

	if standing == StandingHated || standing == StandingHostile {
		return true
	}

	aggressionChance := 0.0
	if standing == StandingUnfriend {
		aggressionChance = 0.2
	}

	return rng.Float64() < aggressionChance
}

// ApplyEnemyKillReputation modifies reputation when player kills a faction member.
func ApplyEnemyKillReputation(sys *ReputationSystem, rep *ReputationComponent, enemyFaction FactionID) {
	if sys == nil || rep == nil {
		return
	}
	sys.ModifyReputation(rep, enemyFaction, -150)
}

// ApplyQuestCompletionReputation rewards reputation for completing a faction quest.
func ApplyQuestCompletionReputation(sys *ReputationSystem, rep *ReputationComponent, quest *FactionQuest) {
	if sys == nil || rep == nil || quest == nil {
		return
	}
	sys.ModifyReputation(rep, quest.FactionID, quest.RepReward)
}

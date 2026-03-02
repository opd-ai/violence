package faction

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
)

func TestNewReputationSystem(t *testing.T) {
	sys := NewReputationSystem()
	if sys == nil {
		t.Fatal("NewReputationSystem returned nil")
	}

	if len(sys.factions) == 0 {
		t.Error("ReputationSystem should have initialized factions")
	}

	expectedFactions := []FactionID{
		FactionMercenaries,
		FactionTechCorp,
		FactionCult,
		FactionRebels,
		FactionSyndicate,
	}

	for _, fid := range expectedFactions {
		if sys.factions[fid] == nil {
			t.Errorf("Missing faction: %s", fid)
		}
	}
}

func TestGetStanding(t *testing.T) {
	tests := []struct {
		score    int
		expected Standing
	}{
		{7000, StandingExalted},
		{6000, StandingExalted},
		{5000, StandingRespect},
		{3000, StandingRespect},
		{2000, StandingFriendly},
		{1000, StandingFriendly},
		{0, StandingNeutral},
		{-500, StandingNeutral},
		{-2000, StandingUnfriend},
		{-4000, StandingHostile},
		{-7000, StandingHated},
	}

	for _, tt := range tests {
		result := GetStanding(tt.score)
		if result != tt.expected {
			t.Errorf("GetStanding(%d) = %v, want %v", tt.score, result, tt.expected)
		}
	}
}

func TestGetPriceModifier(t *testing.T) {
	tests := []struct {
		standing Standing
		expected float64
	}{
		{StandingExalted, 0.7},
		{StandingRespect, 0.85},
		{StandingFriendly, 0.95},
		{StandingNeutral, 1.0},
		{StandingUnfriend, 1.15},
		{StandingHostile, 1.35},
		{StandingHated, 1.6},
	}

	for _, tt := range tests {
		result := GetPriceModifier(tt.standing)
		if result != tt.expected {
			t.Errorf("GetPriceModifier(%v) = %v, want %v", tt.standing, result, tt.expected)
		}
	}
}

func TestGetAggressionModifier(t *testing.T) {
	tests := []struct {
		standing Standing
		min      float64
		max      float64
	}{
		{StandingExalted, 0.0, 0.0},
		{StandingRespect, 0.0, 0.0},
		{StandingFriendly, 0.3, 0.3},
		{StandingNeutral, 1.0, 1.0},
		{StandingUnfriend, 1.3, 1.3},
		{StandingHostile, 1.6, 1.6},
		{StandingHated, 2.0, 2.0},
	}

	for _, tt := range tests {
		result := GetAggressionModifier(tt.standing)
		if result < tt.min || result > tt.max {
			t.Errorf("GetAggressionModifier(%v) = %v, want between %v and %v",
				tt.standing, result, tt.min, tt.max)
		}
	}
}

func TestModifyReputation(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")

	initial := rep.Scores[FactionMercenaries]
	sys.ModifyReputation(rep, FactionMercenaries, 500)

	if rep.Scores[FactionMercenaries] != initial+500 {
		t.Errorf("Expected reputation %d, got %d", initial+500, rep.Scores[FactionMercenaries])
	}
}

func TestReputationPropagation(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")

	sys.ModifyReputation(rep, FactionMercenaries, 1000)

	syndicateRep := rep.Scores[FactionSyndicate]
	if syndicateRep <= 0 {
		t.Errorf("Allied faction should gain reputation, got %d", syndicateRep)
	}

	cultRep := rep.Scores[FactionCult]
	if cultRep >= 0 {
		t.Errorf("Enemy faction should lose reputation, got %d", cultRep)
	}
}

func TestReputationClamping(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")

	sys.ModifyReputation(rep, FactionMercenaries, 20000)
	if rep.Scores[FactionMercenaries] > 10000 {
		t.Errorf("Reputation should be clamped to 10000, got %d", rep.Scores[FactionMercenaries])
	}

	sys.ModifyReputation(rep, FactionCult, -20000)
	if rep.Scores[FactionCult] < -10000 {
		t.Errorf("Reputation should be clamped to -10000, got %d", rep.Scores[FactionCult])
	}
}

func TestGetActiveFactions(t *testing.T) {
	sys := NewReputationSystem()

	genres := []string{"fantasy", "cyberpunk", "horror", "scifi"}

	for _, genre := range genres {
		factions := sys.GetActiveFactions(genre)
		if len(factions) == 0 {
			t.Errorf("Genre %s should have active factions", genre)
		}

		for _, f := range factions {
			if f == nil {
				t.Errorf("Genre %s returned nil faction", genre)
			}
		}
	}
}

func TestCanInteract(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")

	sys.ModifyReputation(rep, FactionMercenaries, 2000)

	if !sys.CanInteract(rep, FactionMercenaries, StandingFriendly) {
		t.Error("Should allow interaction at friendly standing")
	}

	if sys.CanInteract(rep, FactionMercenaries, StandingExalted) {
		t.Error("Should not allow interaction requiring exalted standing")
	}
}

func TestGenerateFactionQuest(t *testing.T) {
	sys := NewReputationSystem()
	rngGen := rng.NewRNG(12345)

	quest := sys.GenerateFactionQuest(FactionMercenaries, rngGen, "fantasy")
	if quest == nil {
		t.Fatal("GenerateFactionQuest returned nil")
	}

	if quest.FactionID != FactionMercenaries {
		t.Errorf("Expected faction %s, got %s", FactionMercenaries, quest.FactionID)
	}

	if quest.RepReward <= 0 {
		t.Error("Quest should have positive reputation reward")
	}

	if quest.CreditReward <= 0 {
		t.Error("Quest should have positive credit reward")
	}

	validTypes := map[string]bool{
		"eliminate": true,
		"retrieve":  true,
		"escort":    true,
		"sabotage":  true,
	}

	if !validTypes[quest.Type] {
		t.Errorf("Invalid quest type: %s", quest.Type)
	}
}

func TestReputationComponent(t *testing.T) {
	rep := &ReputationComponent{
		Scores: make(map[FactionID]int),
	}

	if rep.Type() != "ReputationComponent" {
		t.Errorf("Component type mismatch, got %s", rep.Type())
	}

	rep.Scores[FactionMercenaries] = 500
	if rep.Scores[FactionMercenaries] != 500 {
		t.Error("Failed to store reputation score")
	}
}

func TestFactionMemberComponent(t *testing.T) {
	member := &FactionMemberComponent{
		FactionID: FactionCult,
		Rank:      5,
	}

	if member.Type() != "FactionMemberComponent" {
		t.Errorf("Component type mismatch, got %s", member.Type())
	}

	if member.FactionID != FactionCult {
		t.Error("FactionID not stored correctly")
	}
}

func TestGetFactionShopPriceModifier(t *testing.T) {
	rep := InitializePlayerReputation("fantasy")
	rep.Scores[FactionMercenaries] = 3500

	modifier := GetFactionShopPriceModifier(rep, FactionMercenaries)
	expected := GetPriceModifier(StandingRespect)

	if modifier != expected {
		t.Errorf("Expected modifier %v, got %v", expected, modifier)
	}

	nilModifier := GetFactionShopPriceModifier(nil, FactionMercenaries)
	if nilModifier != 1.0 {
		t.Errorf("Nil reputation should return 1.0, got %v", nilModifier)
	}
}

func TestShouldAttackPlayer(t *testing.T) {
	rngGen := rng.NewRNG(54321)
	rep := InitializePlayerReputation("fantasy")

	rep.Scores[FactionMercenaries] = 2000
	if ShouldAttackPlayer(rep, FactionMercenaries, rngGen) {
		t.Error("Friendly faction should not attack")
	}

	rep.Scores[FactionCult] = -7000
	if !ShouldAttackPlayer(rep, FactionCult, rngGen) {
		t.Error("Hated faction should always attack")
	}

	nilResult := ShouldAttackPlayer(nil, FactionMercenaries, rngGen)
	if nilResult {
		t.Error("Nil reputation should not trigger attack")
	}
}

func TestInitializePlayerReputation(t *testing.T) {
	rep := InitializePlayerReputation("fantasy")

	if rep == nil {
		t.Fatal("InitializePlayerReputation returned nil")
	}

	if rep.Scores == nil {
		t.Fatal("Scores map not initialized")
	}

	allFactions := []FactionID{
		FactionMercenaries,
		FactionTechCorp,
		FactionCult,
		FactionRebels,
		FactionSyndicate,
	}

	for _, fid := range allFactions {
		score, exists := rep.Scores[fid]
		if !exists {
			t.Errorf("Missing initial score for faction %s", fid)
		}
		if score != 0 {
			t.Errorf("Initial score should be 0, got %d for faction %s", score, fid)
		}
	}
}

func TestReputationSystemUpdate(t *testing.T) {
	sys := NewReputationSystem()
	world := engine.NewWorld()

	playerEnt := world.AddEntity()
	rep := InitializePlayerReputation("fantasy")
	rep.Scores[FactionMercenaries] = 100
	world.AddComponent(playerEnt, rep)

	sys.Update(world)

	repType := reflect.TypeOf((*ReputationComponent)(nil))
	updatedRepComp, ok := world.GetComponent(playerEnt, repType)
	if !ok {
		t.Fatal("Could not retrieve reputation component")
	}
	updatedRep := updatedRepComp.(*ReputationComponent)
	if updatedRep.Scores[FactionMercenaries] >= 100 {
		t.Error("Positive reputation should decay slightly")
	}
}

func TestStandingString(t *testing.T) {
	tests := []struct {
		standing Standing
		expected string
	}{
		{StandingExalted, "Exalted"},
		{StandingRespect, "Respected"},
		{StandingFriendly, "Friendly"},
		{StandingNeutral, "Neutral"},
		{StandingUnfriend, "Unfriendly"},
		{StandingHostile, "Hostile"},
		{StandingHated, "Hated"},
	}

	for _, tt := range tests {
		result := tt.standing.String()
		if result != tt.expected {
			t.Errorf("Standing.String() = %s, want %s", result, tt.expected)
		}
	}
}

func TestQuestRewardVariance(t *testing.T) {
	sys := NewReputationSystem()
	rngGen := rng.NewRNG(99999)

	quest1 := sys.GenerateFactionQuest(FactionMercenaries, rngGen, "fantasy")
	quest2 := sys.GenerateFactionQuest(FactionMercenaries, rngGen, "fantasy")

	if quest1.RepReward == quest2.RepReward && quest1.CreditReward == quest2.CreditReward {
		t.Error("Quest rewards should vary due to RNG")
	}
}

func TestFactionEnemiesAndAllies(t *testing.T) {
	sys := NewReputationSystem()

	merc := sys.GetFaction(FactionMercenaries)
	if merc == nil {
		t.Fatal("Failed to get mercenaries faction")
	}

	hasEnemy := false
	for _, enemy := range merc.Enemies {
		if enemy == FactionCult {
			hasEnemy = true
			break
		}
	}
	if !hasEnemy {
		t.Error("Mercenaries should be enemies with Cult")
	}

	hasAlly := false
	for _, ally := range merc.Allies {
		if ally == FactionSyndicate {
			hasAlly = true
			break
		}
	}
	if !hasAlly {
		t.Error("Mercenaries should be allies with Syndicate")
	}
}

func BenchmarkModifyReputation(b *testing.B) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.ModifyReputation(rep, FactionMercenaries, 10)
	}
}

func BenchmarkGetStanding(b *testing.B) {
	scores := []int{-8000, -4000, -1500, 0, 1500, 4000, 8000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetStanding(scores[i%len(scores)])
	}
}

func BenchmarkGenerateFactionQuest(b *testing.B) {
	sys := NewReputationSystem()
	rngGen := rng.NewRNG(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.GenerateFactionQuest(FactionMercenaries, rngGen, "fantasy")
	}
}

func TestApplyEnemyKillReputation(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")

	initial := rep.Scores[FactionMercenaries]
	ApplyEnemyKillReputation(sys, rep, FactionMercenaries)

	if rep.Scores[FactionMercenaries] >= initial {
		t.Error("Killing faction member should decrease reputation")
	}

	expectedDecrease := -150
	actual := rep.Scores[FactionMercenaries] - initial
	if actual != expectedDecrease {
		t.Errorf("Expected decrease around %d, got %d", expectedDecrease, actual)
	}
}

func TestApplyQuestCompletionReputation(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")
	rngGen := rng.NewRNG(42)

	quest := sys.GenerateFactionQuest(FactionRebels, rngGen, "fantasy")
	if quest == nil {
		t.Fatal("Failed to generate quest")
	}

	initial := rep.Scores[FactionRebels]
	ApplyQuestCompletionReputation(sys, rep, quest)

	increase := rep.Scores[FactionRebels] - initial
	if increase <= 0 {
		t.Error("Quest completion should increase reputation")
	}

	if increase != quest.RepReward {
		t.Errorf("Expected reputation increase %d, got %d", quest.RepReward, increase)
	}
}

func TestApplyEnemyKillReputationNil(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")

	ApplyEnemyKillReputation(nil, rep, FactionMercenaries)
	ApplyEnemyKillReputation(sys, nil, FactionMercenaries)
}

func TestApplyQuestCompletionReputationNil(t *testing.T) {
	sys := NewReputationSystem()
	rep := InitializePlayerReputation("fantasy")
	rngGen := rng.NewRNG(99)
	quest := sys.GenerateFactionQuest(FactionCult, rngGen, "horror")

	ApplyQuestCompletionReputation(nil, rep, quest)
	ApplyQuestCompletionReputation(sys, nil, quest)
	ApplyQuestCompletionReputation(sys, rep, nil)
}

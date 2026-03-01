// Package network integration tests verify matchmaking queue behavior.
package network

import (
	"testing"
	"time"
)

// TestMatchmakingQueue4Players verifies 4-player matchmaking queue functionality.
// This is the integration test required by PLAN.md task 1.
func TestMatchmakingQueue4Players(t *testing.T) {
	config := MatchmakingConfig{
		MinPlayers:     4,
		MaxWaitTime:    30 * time.Second,
		SkillTolerance: 200,
		RegionPriority: []string{"us-east", "us-west", "eu-west"},
	}

	// Create 4 players with varying skill levels within tolerance
	players := []Player{
		{ID: "player1", Name: "Alice", EloRating: 1200, Wins: 10, Losses: 8},
		{ID: "player2", Name: "Bob", EloRating: 1250, Wins: 12, Losses: 10},
		{ID: "player3", Name: "Charlie", EloRating: 1150, Wins: 8, Losses: 9},
		{ID: "player4", Name: "Diana", EloRating: 1300, Wins: 15, Losses: 7},
	}

	// Phase 1: Verify all players are within skill tolerance
	matched := MatchPlayersWithinSkillRange(players, config.SkillTolerance)
	if len(matched) != 4 {
		t.Errorf("expected all 4 players within skill range, got %d", len(matched))
	}

	// Phase 2: Balance players into two teams
	teamA, teamB := BalanceTeams(matched)

	if len(teamA) != 2 {
		t.Errorf("expected team A to have 2 players, got %d", len(teamA))
	}
	if len(teamB) != 2 {
		t.Errorf("expected team B to have 2 players, got %d", len(teamB))
	}

	// Phase 3: Verify teams are balanced (within 10% difference)
	balance := TeamBalanceDifference(teamA, teamB)
	if balance > 10.0 {
		t.Errorf("teams not balanced: %.2f%% difference (should be <10%%)", balance)
	}

	// Phase 4: Verify team composition
	avgA := AverageElo(teamA)
	avgB := AverageElo(teamB)

	t.Logf("Team A: %s (%.0f) + %s (%.0f) = avg %.0f",
		teamA[0].Name, float64(teamA[0].EloRating),
		teamA[1].Name, float64(teamA[1].EloRating),
		avgA)
	t.Logf("Team B: %s (%.0f) + %s (%.0f) = avg %.0f",
		teamB[0].Name, float64(teamB[0].EloRating),
		teamB[1].Name, float64(teamB[1].EloRating),
		avgB)

	// Both teams should have average Elo between 1150 and 1300
	if avgA < 1150 || avgA > 1300 {
		t.Errorf("team A average Elo out of expected range: %.0f", avgA)
	}
	if avgB < 1150 || avgB > 1300 {
		t.Errorf("team B average Elo out of expected range: %.0f", avgB)
	}
}

// TestMatchmakingQueueSkillFiltering tests that players outside tolerance are excluded.
func TestMatchmakingQueueSkillFiltering(t *testing.T) {
	config := MatchmakingConfig{
		MinPlayers:     4,
		MaxWaitTime:    30 * time.Second,
		SkillTolerance: 100, // Strict tolerance
		RegionPriority: []string{"us-east"},
	}

	// Create 6 players with wide skill range
	players := []Player{
		{ID: "player1", Name: "Newbie", EloRating: 800, Wins: 2, Losses: 10},
		{ID: "player2", Name: "Average1", EloRating: 1200, Wins: 10, Losses: 10},
		{ID: "player3", Name: "Average2", EloRating: 1220, Wins: 11, Losses: 9},
		{ID: "player4", Name: "Average3", EloRating: 1180, Wins: 9, Losses: 11},
		{ID: "player5", Name: "Average4", EloRating: 1250, Wins: 12, Losses: 8},
		{ID: "player6", Name: "Pro", EloRating: 1800, Wins: 50, Losses: 10},
	}

	// Filter players within skill range
	matched := MatchPlayersWithinSkillRange(players, config.SkillTolerance)

	// Should exclude newbie (800) and pro (1800)
	// Median is player3 at 1220, so tolerance ±100 = [1120, 1320]
	if len(matched) < 4 {
		t.Errorf("expected at least 4 players within tolerance, got %d", len(matched))
	}

	// Verify excluded players
	for _, p := range matched {
		if p.EloRating < 1100 || p.EloRating > 1350 {
			t.Errorf("player %s (Elo %d) should have been filtered out", p.Name, p.EloRating)
		}
	}

	// Verify newbie and pro are excluded
	foundNewbie := false
	foundPro := false
	for _, p := range matched {
		if p.Name == "Newbie" {
			foundNewbie = true
		}
		if p.Name == "Pro" {
			foundPro = true
		}
	}

	if foundNewbie {
		t.Error("newbie (Elo 800) should be filtered out")
	}
	if foundPro {
		t.Error("pro (Elo 1800) should be filtered out")
	}
}

// TestMatchmakingQueueLargePool tests balancing with larger player pool.
func TestMatchmakingQueueLargePool(t *testing.T) {
	// Create 8 players for a 4v4 match
	players := []Player{
		{ID: "p1", Name: "Player1", EloRating: 1100, Wins: 5, Losses: 5},
		{ID: "p2", Name: "Player2", EloRating: 1150, Wins: 6, Losses: 4},
		{ID: "p3", Name: "Player3", EloRating: 1200, Wins: 10, Losses: 10},
		{ID: "p4", Name: "Player4", EloRating: 1220, Wins: 11, Losses: 9},
		{ID: "p5", Name: "Player5", EloRating: 1250, Wins: 12, Losses: 8},
		{ID: "p6", Name: "Player6", EloRating: 1280, Wins: 13, Losses: 7},
		{ID: "p7", Name: "Player7", EloRating: 1300, Wins: 14, Losses: 6},
		{ID: "p8", Name: "Player8", EloRating: 1350, Wins: 16, Losses: 4},
	}

	// Balance into two teams
	teamA, teamB := BalanceTeams(players)

	// Should have 4 players each
	if len(teamA) != 4 {
		t.Errorf("expected team A to have 4 players, got %d", len(teamA))
	}
	if len(teamB) != 4 {
		t.Errorf("expected team B to have 4 players, got %d", len(teamB))
	}

	// Verify balance
	balance := TeamBalanceDifference(teamA, teamB)
	if balance > 10.0 {
		t.Errorf("8-player teams not balanced: %.2f%% difference", balance)
	}

	avgA := AverageElo(teamA)
	avgB := AverageElo(teamB)

	t.Logf("Team A (4 players): avg Elo %.0f", avgA)
	t.Logf("Team B (4 players): avg Elo %.0f", avgB)
}

// TestMatchmakingQueueMinPlayers tests behavior when minimum players not met.
func TestMatchmakingQueueMinPlayers(t *testing.T) {
	config := MatchmakingConfig{
		MinPlayers:     4,
		MaxWaitTime:    30 * time.Second,
		SkillTolerance: 200,
		RegionPriority: []string{"us-east"},
	}

	// Only 2 players in queue
	players := []Player{
		{ID: "player1", Name: "Alice", EloRating: 1200, Wins: 10, Losses: 8},
		{ID: "player2", Name: "Bob", EloRating: 1250, Wins: 12, Losses: 10},
	}

	// Match within skill range should return both players
	matched := MatchPlayersWithinSkillRange(players, config.SkillTolerance)
	if len(matched) != 2 {
		t.Errorf("expected 2 players within skill range, got %d", len(matched))
	}

	// But we can't form a 4-player match
	if len(matched) < config.MinPlayers {
		t.Logf("correctly identified insufficient players: %d < %d (min)", len(matched), config.MinPlayers)
	}
}

// TestMatchmakingQueueEloUpdates tests Elo calculation after match completion.
func TestMatchmakingQueueEloUpdates(t *testing.T) {
	// Create 4 players
	players := []Player{
		{ID: "p1", Name: "Alice", EloRating: 1200, Wins: 10, Losses: 8},
		{ID: "p2", Name: "Bob", EloRating: 1250, Wins: 12, Losses: 10},
		{ID: "p3", Name: "Charlie", EloRating: 1150, Wins: 8, Losses: 9},
		{ID: "p4", Name: "Diana", EloRating: 1300, Wins: 15, Losses: 7},
	}

	// Balance teams
	teamA, teamB := BalanceTeams(players)

	// Simulate team A winning
	avgA := int(AverageElo(teamA))
	avgB := int(AverageElo(teamB))

	winnerDelta, loserDelta := CalculateEloChange(avgA, avgB)

	// Apply Elo changes to team A (winners)
	for i := range teamA {
		teamA[i].EloRating += winnerDelta
		teamA[i].Wins++
	}

	// Apply Elo changes to team B (losers)
	for i := range teamB {
		teamB[i].EloRating += loserDelta // loserDelta is negative
		teamB[i].Losses++
	}

	// Verify Elo changes
	if winnerDelta <= 0 {
		t.Errorf("winner delta should be positive, got %d", winnerDelta)
	}
	if loserDelta >= 0 {
		t.Errorf("loser delta should be negative, got %d", loserDelta)
	}

	// Verify wins/losses updated
	for _, p := range teamA {
		if p.Wins < 1 {
			t.Errorf("team A player %s should have at least 1 win", p.Name)
		}
	}
	for _, p := range teamB {
		if p.Losses < 1 {
			t.Errorf("team B player %s should have at least 1 loss", p.Name)
		}
	}

	t.Logf("Match complete: Team A +%d Elo, Team B %d Elo", winnerDelta, loserDelta)
}

// TestMatchmakingQueueConsecutiveMatches tests multiple rounds of matchmaking.
func TestMatchmakingQueueConsecutiveMatches(t *testing.T) {
	// Create 8 players
	players := []Player{
		{ID: "p1", Name: "P1", EloRating: 1200, Wins: 0, Losses: 0},
		{ID: "p2", Name: "P2", EloRating: 1200, Wins: 0, Losses: 0},
		{ID: "p3", Name: "P3", EloRating: 1200, Wins: 0, Losses: 0},
		{ID: "p4", Name: "P4", EloRating: 1200, Wins: 0, Losses: 0},
		{ID: "p5", Name: "P5", EloRating: 1200, Wins: 0, Losses: 0},
		{ID: "p6", Name: "P6", EloRating: 1200, Wins: 0, Losses: 0},
		{ID: "p7", Name: "P7", EloRating: 1200, Wins: 0, Losses: 0},
		{ID: "p8", Name: "P8", EloRating: 1200, Wins: 0, Losses: 0},
	}

	// Run 10 matches, updating Elo each time
	for round := 0; round < 10; round++ {
		teamA, teamB := BalanceTeams(players)

		// Team A wins this round
		avgA := int(AverageElo(teamA))
		avgB := int(AverageElo(teamB))
		winDelta, loseDelta := CalculateEloChange(avgA, avgB)

		// Update ratings
		for i := range teamA {
			for j := range players {
				if players[j].ID == teamA[i].ID {
					players[j].EloRating += winDelta
					players[j].Wins++
				}
			}
		}

		for i := range teamB {
			for j := range players {
				if players[j].ID == teamB[i].ID {
					players[j].EloRating += loseDelta
					players[j].Losses++
				}
			}
		}
	}

	// After 10 rounds, Elo should have diverged
	minElo := 10000
	maxElo := 0
	for _, p := range players {
		if p.EloRating < minElo {
			minElo = p.EloRating
		}
		if p.EloRating > maxElo {
			maxElo = p.EloRating
		}
	}

	spread := maxElo - minElo
	if spread == 0 {
		t.Error("Elo ratings should have diverged after 10 matches")
	}

	t.Logf("After 10 matches: Elo spread = %d (min=%d, max=%d)", spread, minElo, maxElo)
}

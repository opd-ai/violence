package network

import (
	"math"
	"testing"
)

func TestCalculateEloChange(t *testing.T) {
	tests := []struct {
		name        string
		winnerElo   int
		loserElo    int
		wantWinMin  int // Minimum expected delta for winner
		wantWinMax  int // Maximum expected delta for winner
		wantLoseMin int // Maximum expected delta for loser (most negative)
		wantLoseMax int // Minimum expected delta for loser (least negative)
	}{
		{
			name:        "equal ratings",
			winnerElo:   1200,
			loserElo:    1200,
			wantWinMin:  15,
			wantWinMax:  17,
			wantLoseMin: -17,
			wantLoseMax: -15,
		},
		{
			name:        "underdog wins",
			winnerElo:   1000,
			loserElo:    1400,
			wantWinMin:  27,
			wantWinMax:  29,
			wantLoseMin: -29,
			wantLoseMax: -27,
		},
		{
			name:        "favorite wins",
			winnerElo:   1400,
			loserElo:    1000,
			wantWinMin:  3,
			wantWinMax:  5,
			wantLoseMin: -5,
			wantLoseMax: -3,
		},
		{
			name:        "small difference",
			winnerElo:   1250,
			loserElo:    1200,
			wantWinMin:  13,
			wantWinMax:  15,
			wantLoseMin: -15,
			wantLoseMax: -13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			winDelta, loseDelta := CalculateEloChange(tt.winnerElo, tt.loserElo)

			if winDelta < tt.wantWinMin || winDelta > tt.wantWinMax {
				t.Errorf("CalculateEloChange() winnerDelta = %v, want between %v and %v",
					winDelta, tt.wantWinMin, tt.wantWinMax)
			}

			if loseDelta < tt.wantLoseMin || loseDelta > tt.wantLoseMax {
				t.Errorf("CalculateEloChange() loserDelta = %v, want between %v and %v",
					loseDelta, tt.wantLoseMin, tt.wantLoseMax)
			}

			// Sum should be close to zero (rating is conserved)
			sumDelta := winDelta + loseDelta
			if sumDelta < -2 || sumDelta > 2 {
				t.Errorf("CalculateEloChange() sum of deltas = %v, expected near 0", sumDelta)
			}
		})
	}
}

func TestBalanceTeams(t *testing.T) {
	tests := []struct {
		name              string
		players           []Player
		wantMaxDifference float64 // Max acceptable % difference in avg Elo
	}{
		{
			name: "even skill levels",
			players: []Player{
				{ID: "p1", EloRating: 1200},
				{ID: "p2", EloRating: 1210},
				{ID: "p3", EloRating: 1190},
				{ID: "p4", EloRating: 1205},
			},
			wantMaxDifference: 2.0,
		},
		{
			name: "mixed skill levels",
			players: []Player{
				{ID: "p1", EloRating: 1500},
				{ID: "p2", EloRating: 900},
				{ID: "p3", EloRating: 1300},
				{ID: "p4", EloRating: 1100},
			},
			wantMaxDifference: 5.0,
		},
		{
			name: "wide skill gap",
			players: []Player{
				{ID: "p1", EloRating: 1800},
				{ID: "p2", EloRating: 800},
				{ID: "p3", EloRating: 1200},
				{ID: "p4", EloRating: 1200},
			},
			wantMaxDifference: 10.0,
		},
		{
			name: "odd number of players",
			players: []Player{
				{ID: "p1", EloRating: 1400},
				{ID: "p2", EloRating: 1200},
				{ID: "p3", EloRating: 1000},
			},
			wantMaxDifference: 25.0,
		},
		{
			name:              "two players",
			players:           []Player{{ID: "p1", EloRating: 1300}, {ID: "p2", EloRating: 1100}},
			wantMaxDifference: 100.0,
		},
		{
			name:              "one player",
			players:           []Player{{ID: "p1", EloRating: 1200}},
			wantMaxDifference: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			teamA, teamB := BalanceTeams(tt.players)

			// Check all players are assigned
			totalAssigned := len(teamA) + len(teamB)
			if totalAssigned != len(tt.players) {
				t.Errorf("BalanceTeams() assigned %v players, want %v", totalAssigned, len(tt.players))
			}

			// Check team sizes are balanced (differ by at most 1)
			sizeDiff := len(teamA) - len(teamB)
			if sizeDiff < -1 || sizeDiff > 1 {
				t.Errorf("BalanceTeams() team size difference = %v, want ±1", sizeDiff)
			}

			if len(tt.players) < 2 {
				return // Skip balance check for edge cases
			}

			// Check skill balance
			difference := TeamBalanceDifference(teamA, teamB)
			if difference > tt.wantMaxDifference {
				t.Errorf("BalanceTeams() skill difference = %.2f%%, want <= %.2f%%",
					difference, tt.wantMaxDifference)
				t.Logf("Team A avg: %.2f, Team B avg: %.2f",
					AverageElo(teamA), AverageElo(teamB))
			}
		})
	}
}

func TestMatchPlayersWithinSkillRange(t *testing.T) {
	tests := []struct {
		name      string
		players   []Player
		tolerance int
		wantCount int
		wantIDs   []string
	}{
		{
			name: "all within tolerance",
			players: []Player{
				{ID: "p1", EloRating: 1200},
				{ID: "p2", EloRating: 1250},
				{ID: "p3", EloRating: 1150},
			},
			tolerance: 200,
			wantCount: 3,
			wantIDs:   []string{"p1", "p2", "p3"},
		},
		{
			name: "filter outliers",
			players: []Player{
				{ID: "p1", EloRating: 1200},
				{ID: "p2", EloRating: 1250},
				{ID: "p3", EloRating: 1800}, // Too high
				{ID: "p4", EloRating: 600},  // Too low
			},
			tolerance: 100,
			wantCount: 2,
			wantIDs:   []string{"p1", "p2"},
		},
		{
			name: "tight tolerance",
			players: []Player{
				{ID: "p1", EloRating: 1200},
				{ID: "p2", EloRating: 1210},
				{ID: "p3", EloRating: 1260},
			},
			tolerance: 20,
			wantCount: 2,
			wantIDs:   []string{"p1", "p2"},
		},
		{
			name:      "empty list",
			players:   []Player{},
			tolerance: 200,
			wantCount: 0,
		},
		{
			name: "single player",
			players: []Player{
				{ID: "p1", EloRating: 1200},
			},
			tolerance: 200,
			wantCount: 1,
			wantIDs:   []string{"p1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := MatchPlayersWithinSkillRange(tt.players, tt.tolerance)

			if len(matched) != tt.wantCount {
				t.Errorf("MatchPlayersWithinSkillRange() count = %v, want %v",
					len(matched), tt.wantCount)
			}

			if tt.wantIDs != nil {
				matchedIDs := make(map[string]bool)
				for _, p := range matched {
					matchedIDs[p.ID] = true
				}

				for _, wantID := range tt.wantIDs {
					if !matchedIDs[wantID] {
						t.Errorf("MatchPlayersWithinSkillRange() missing player %v", wantID)
					}
				}
			}
		})
	}
}

func TestAverageElo(t *testing.T) {
	tests := []struct {
		name    string
		players []Player
		want    float64
	}{
		{
			name: "simple average",
			players: []Player{
				{EloRating: 1000},
				{EloRating: 1200},
				{EloRating: 1400},
			},
			want: 1200,
		},
		{
			name: "single player",
			players: []Player{
				{EloRating: 1500},
			},
			want: 1500,
		},
		{
			name:    "empty list",
			players: []Player{},
			want:    0,
		},
		{
			name: "mixed ratings",
			players: []Player{
				{EloRating: 800},
				{EloRating: 1600},
			},
			want: 1200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AverageElo(tt.players)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("AverageElo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTeamBalanceDifference(t *testing.T) {
	tests := []struct {
		name  string
		teamA []Player
		teamB []Player
		want  float64
	}{
		{
			name: "perfectly balanced",
			teamA: []Player{
				{EloRating: 1200},
				{EloRating: 1000},
			},
			teamB: []Player{
				{EloRating: 1100},
				{EloRating: 1100},
			},
			want: 0,
		},
		{
			name: "10% difference",
			teamA: []Player{
				{EloRating: 1200},
			},
			teamB: []Player{
				{EloRating: 1000},
			},
			want: 18.18, // (200/1100)*100 ≈ 18.18%
		},
		{
			name: "small difference",
			teamA: []Player{
				{EloRating: 1210},
				{EloRating: 1190},
			},
			teamB: []Player{
				{EloRating: 1205},
				{EloRating: 1195},
			},
			want: 0, // Both average 1200
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TeamBalanceDifference(tt.teamA, tt.teamB)
			if math.Abs(got-tt.want) > 0.5 {
				t.Errorf("TeamBalanceDifference() = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}

func TestMatchmakingConfigDefaults(t *testing.T) {
	// Test that default constants are sensible
	if DefaultEloRating != 1200 {
		t.Errorf("DefaultEloRating = %v, want 1200", DefaultEloRating)
	}
	if DefaultKFactor != 32 {
		t.Errorf("DefaultKFactor = %v, want 32", DefaultKFactor)
	}
	if DefaultSkillTolerance != 200 {
		t.Errorf("DefaultSkillTolerance = %v, want 200", DefaultSkillTolerance)
	}
}

// BenchmarkBalanceTeams measures team balancing performance
func BenchmarkBalanceTeams(b *testing.B) {
	players := make([]Player, 16)
	for i := range players {
		players[i] = Player{
			ID:        string(rune('A' + i)),
			EloRating: 1000 + i*50,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BalanceTeams(players)
	}
}

// BenchmarkCalculateEloChange measures Elo calculation performance
func BenchmarkCalculateEloChange(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CalculateEloChange(1200, 1400)
	}
}

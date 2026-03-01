// Package network provides Elo-based matchmaking and skill rating.
package network

import (
	"math"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultEloRating is the starting Elo rating for new players.
	DefaultEloRating = 1200
	// DefaultKFactor controls how much ratings change per game.
	DefaultKFactor = 32
	// DefaultSkillTolerance is the default max Elo difference for matchmaking.
	DefaultSkillTolerance = 200
)

// MatchmakingConfig defines matchmaking parameters.
type MatchmakingConfig struct {
	MinPlayers     int           // Minimum players to start match
	MaxWaitTime    time.Duration // Max time a player waits in queue
	SkillTolerance int           // Max Elo difference allowed (default: ±200)
	RegionPriority []string      // Preferred regions in order
}

// Player represents a player with skill rating.
type Player struct {
	ID        string
	Name      string
	EloRating int
	Wins      int
	Losses    int
}

// CalculateEloChange computes Elo rating changes after a match.
// Uses standard Elo formula: newRating = oldRating + K * (actual - expected)
// Returns delta for winner and loser (loser delta will be negative).
func CalculateEloChange(winnerElo, loserElo int) (winnerDelta, loserDelta int) {
	kFactor := float64(DefaultKFactor)

	// Calculate expected scores (probability of winning)
	expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserElo-winnerElo)/400.0))
	expectedLoser := 1.0 / (1.0 + math.Pow(10, float64(winnerElo-loserElo)/400.0))

	// Actual scores: winner = 1, loser = 0
	winnerDelta = int(math.Round(kFactor * (1.0 - expectedWinner)))
	loserDelta = int(math.Round(kFactor * (0.0 - expectedLoser)))

	return winnerDelta, loserDelta
}

// BalanceTeams creates two balanced teams from a list of players.
// Minimizes the difference in average Elo between teams.
// Returns teamA and teamB with roughly equal skill levels.
func BalanceTeams(players []Player) (teamA, teamB []Player) {
	if len(players) < 2 {
		return players, nil
	}

	// Sort players by Elo rating (highest to lowest)
	sorted := make([]Player, len(players))
	copy(sorted, players)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EloRating > sorted[j].EloRating
	})

	// Use greedy algorithm: assign each player to team with lower total Elo
	var sumA, sumB int
	teamA = make([]Player, 0, len(players)/2+1)
	teamB = make([]Player, 0, len(players)/2+1)

	for _, player := range sorted {
		if sumA <= sumB {
			teamA = append(teamA, player)
			sumA += player.EloRating
		} else {
			teamB = append(teamB, player)
			sumB += player.EloRating
		}
	}

	logrus.WithFields(logrus.Fields{
		"team_a_avg":  float64(sumA) / float64(len(teamA)),
		"team_b_avg":  float64(sumB) / float64(len(teamB)),
		"team_a_size": len(teamA),
		"team_b_size": len(teamB),
	}).Debug("teams balanced")

	return teamA, teamB
}

// MatchPlayersWithinSkillRange filters players within skill tolerance.
// Returns only players whose Elo is within tolerance of the median Elo.
func MatchPlayersWithinSkillRange(players []Player, tolerance int) []Player {
	if len(players) == 0 {
		return nil
	}

	// Calculate median Elo
	sorted := make([]Player, len(players))
	copy(sorted, players)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EloRating < sorted[j].EloRating
	})

	median := sorted[len(sorted)/2].EloRating

	// Filter players within tolerance
	matched := make([]Player, 0, len(players))
	for _, player := range players {
		diff := player.EloRating - median
		if diff < 0 {
			diff = -diff
		}
		if diff <= tolerance {
			matched = append(matched, player)
		}
	}

	return matched
}

// AverageElo calculates the average Elo rating of a player list.
func AverageElo(players []Player) float64 {
	if len(players) == 0 {
		return 0
	}

	sum := 0
	for _, p := range players {
		sum += p.EloRating
	}

	return float64(sum) / float64(len(players))
}

// TeamBalanceDifference calculates the percentage difference in average Elo.
// Returns 0-100 representing how balanced the teams are (0 = perfectly balanced).
func TeamBalanceDifference(teamA, teamB []Player) float64 {
	avgA := AverageElo(teamA)
	avgB := AverageElo(teamB)

	if avgA == 0 || avgB == 0 {
		return 0
	}

	diff := avgA - avgB
	if diff < 0 {
		diff = -diff
	}

	avg := (avgA + avgB) / 2
	return (diff / avg) * 100
}

// Package leaderboard provides local and federated score tracking.
package leaderboard

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// LeaderboardEntry represents a single leaderboard record.
type LeaderboardEntry struct {
	PlayerID   string
	PlayerName string
	Score      int64
	Stat       string // "kills", "wins", "high_score", etc.
	Period     string // "all_time", "weekly", "daily"
	UpdatedAt  time.Time
}

// Leaderboard manages score tracking and ranking.
type Leaderboard struct {
	db *sql.DB
}

// New creates a new leaderboard with SQLite persistence.
func New(dbPath string) (*Leaderboard, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	lb := &Leaderboard{db: db}

	if err := lb.createTables(); err != nil {
		db.Close()
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"db_path": dbPath,
	}).Info("leaderboard initialized")

	return lb, nil
}

// createTables initializes the database schema.
func (lb *Leaderboard) createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS leaderboard (
		player_id TEXT NOT NULL,
		player_name TEXT NOT NULL,
		stat TEXT NOT NULL,
		period TEXT NOT NULL,
		score INTEGER NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL,
		PRIMARY KEY (player_id, stat, period)
	);
	CREATE INDEX IF NOT EXISTS idx_leaderboard_stat_period_score 
		ON leaderboard(stat, period, score DESC);
	`

	if _, err := lb.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	return nil
}

// RecordScore updates or inserts a player's score for a given stat and period.
func (lb *Leaderboard) RecordScore(playerID, playerName, stat, period string, value int64) error {
	query := `
	INSERT INTO leaderboard (player_id, player_name, stat, period, score, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(player_id, stat, period) DO UPDATE SET
		score = excluded.score,
		player_name = excluded.player_name,
		updated_at = excluded.updated_at
	`

	now := time.Now()
	_, err := lb.db.Exec(query, playerID, playerName, stat, period, value, now)
	if err != nil {
		return fmt.Errorf("failed to record score: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"player_id":   playerID,
		"player_name": playerName,
		"stat":        stat,
		"period":      period,
		"score":       value,
	}).Debug("score recorded")

	return nil
}

// GetTop returns the top N entries for a given stat and period.
func (lb *Leaderboard) GetTop(stat, period string, limit int) ([]LeaderboardEntry, error) {
	query := `
	SELECT player_id, player_name, stat, period, score, updated_at
	FROM leaderboard
	WHERE stat = ? AND period = ?
	ORDER BY score DESC
	LIMIT ?
	`

	rows, err := lb.db.Query(query, stat, period, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top scores: %w", err)
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.PlayerID, &e.PlayerName, &e.Stat, &e.Period, &e.Score, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		entries = append(entries, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return entries, nil
}

// GetRank returns the rank (1-based) of a player for a given stat and period.
// Returns 0 if the player is not found.
func (lb *Leaderboard) GetRank(playerID, stat, period string) (int, error) {
	// First check if player exists
	var playerScore int64
	err := lb.db.QueryRow(
		`SELECT score FROM leaderboard WHERE player_id = ? AND stat = ? AND period = ?`,
		playerID, stat, period,
	).Scan(&playerScore)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Player not found
		}
		return 0, fmt.Errorf("failed to get player score: %w", err)
	}

	// Count how many players have higher scores
	query := `
	SELECT COUNT(*) + 1
	FROM leaderboard
	WHERE stat = ? AND period = ? AND score > ?
	`

	var rank int
	err = lb.db.QueryRow(query, stat, period, playerScore).Scan(&rank)
	if err != nil {
		return 0, fmt.Errorf("failed to get rank: %w", err)
	}

	return rank, nil
}

// IncrementScore adds a value to the player's current score.
func (lb *Leaderboard) IncrementScore(playerID, playerName, stat, period string, delta int64) error {
	// Get current score
	currentScore, err := lb.getScore(playerID, stat, period)
	if err != nil {
		return err
	}

	newScore := currentScore + delta
	return lb.RecordScore(playerID, playerName, stat, period, newScore)
}

// getScore retrieves the current score for a player.
func (lb *Leaderboard) getScore(playerID, stat, period string) (int64, error) {
	query := `
	SELECT score FROM leaderboard
	WHERE player_id = ? AND stat = ? AND period = ?
	`

	var score int64
	err := lb.db.QueryRow(query, playerID, stat, period).Scan(&score)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // No score yet
		}
		return 0, fmt.Errorf("failed to get score: %w", err)
	}

	return score, nil
}

// ClearPeriod removes all entries for a specific period (e.g., weekly reset).
func (lb *Leaderboard) ClearPeriod(period string) error {
	query := `DELETE FROM leaderboard WHERE period = ?`

	result, err := lb.db.Exec(query, period)
	if err != nil {
		return fmt.Errorf("failed to clear period: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logrus.WithFields(logrus.Fields{
		"period":        period,
		"rows_affected": rowsAffected,
	}).Info("leaderboard period cleared")

	return nil
}

// Close closes the database connection.
func (lb *Leaderboard) Close() error {
	if lb.db != nil {
		return lb.db.Close()
	}
	return nil
}

package leaderboard

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	lb, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	if lb.db == nil {
		t.Error("New() should initialize database connection")
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("New() should create database file")
	}
}

func TestNewInvalidPath(t *testing.T) {
	// Try to create database in non-existent directory with strict permissions
	_, err := New("/nonexistent/dir/test.db")
	if err == nil {
		t.Error("New() should error on invalid path")
	}
}

func TestRecordScore(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	tests := []struct {
		name       string
		playerID   string
		playerName string
		stat       string
		period     string
		score      int64
	}{
		{
			name:       "record kills",
			playerID:   "player1",
			playerName: "Alice",
			stat:       "kills",
			period:     "all_time",
			score:      100,
		},
		{
			name:       "record wins",
			playerID:   "player2",
			playerName: "Bob",
			stat:       "wins",
			period:     "weekly",
			score:      5,
		},
		{
			name:       "record high score",
			playerID:   "player3",
			playerName: "Charlie",
			stat:       "high_score",
			period:     "daily",
			score:      9999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lb.RecordScore(tt.playerID, tt.playerName, tt.stat, tt.period, tt.score)
			if err != nil {
				t.Errorf("RecordScore() error = %v", err)
			}

			// Verify score was recorded
			score, err := lb.getScore(tt.playerID, tt.stat, tt.period)
			if err != nil {
				t.Errorf("getScore() error = %v", err)
			}
			if score != tt.score {
				t.Errorf("getScore() = %v, want %v", score, tt.score)
			}
		})
	}
}

func TestRecordScoreUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	playerID := "player1"
	playerName := "Alice"
	stat := "kills"
	period := "all_time"

	// Record initial score
	if err := lb.RecordScore(playerID, playerName, stat, period, 100); err != nil {
		t.Fatalf("RecordScore() error = %v", err)
	}

	// Update score (should replace, not add)
	if err := lb.RecordScore(playerID, playerName, stat, period, 200); err != nil {
		t.Fatalf("RecordScore() update error = %v", err)
	}

	score, err := lb.getScore(playerID, stat, period)
	if err != nil {
		t.Fatalf("getScore() error = %v", err)
	}
	if score != 200 {
		t.Errorf("getScore() after update = %v, want 200", score)
	}
}

func TestGetTop(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Record scores for multiple players
	players := []struct {
		id    string
		name  string
		score int64
	}{
		{"p1", "Alice", 1000},
		{"p2", "Bob", 500},
		{"p3", "Charlie", 1500},
		{"p4", "Diana", 800},
		{"p5", "Eve", 1200},
	}

	for _, p := range players {
		if err := lb.RecordScore(p.id, p.name, "kills", "all_time", p.score); err != nil {
			t.Fatalf("RecordScore() error = %v", err)
		}
	}

	// Get top 3
	top, err := lb.GetTop("kills", "all_time", 3)
	if err != nil {
		t.Fatalf("GetTop() error = %v", err)
	}

	if len(top) != 3 {
		t.Fatalf("GetTop() returned %v entries, want 3", len(top))
	}

	// Verify ordering (descending)
	expectedOrder := []struct {
		id    string
		score int64
	}{
		{"p3", 1500},
		{"p5", 1200},
		{"p1", 1000},
	}

	for i, expected := range expectedOrder {
		if top[i].PlayerID != expected.id {
			t.Errorf("top[%d].PlayerID = %v, want %v", i, top[i].PlayerID, expected.id)
		}
		if top[i].Score != expected.score {
			t.Errorf("top[%d].Score = %v, want %v", i, top[i].Score, expected.score)
		}
	}
}

func TestGetTopEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Query empty leaderboard
	top, err := lb.GetTop("kills", "all_time", 10)
	if err != nil {
		t.Fatalf("GetTop() error = %v", err)
	}

	if len(top) != 0 {
		t.Errorf("GetTop() on empty leaderboard returned %v entries, want 0", len(top))
	}
}

func TestGetTopDifferentPeriods(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Record scores for different periods
	lb.RecordScore("p1", "Alice", "kills", "all_time", 1000)
	lb.RecordScore("p1", "Alice", "kills", "weekly", 100)
	lb.RecordScore("p2", "Bob", "kills", "all_time", 500)
	lb.RecordScore("p2", "Bob", "kills", "weekly", 200)

	// Get all_time top
	topAllTime, err := lb.GetTop("kills", "all_time", 10)
	if err != nil {
		t.Fatalf("GetTop(all_time) error = %v", err)
	}
	if len(topAllTime) != 2 {
		t.Errorf("GetTop(all_time) returned %v entries, want 2", len(topAllTime))
	}

	// Get weekly top
	topWeekly, err := lb.GetTop("kills", "weekly", 10)
	if err != nil {
		t.Fatalf("GetTop(weekly) error = %v", err)
	}
	if len(topWeekly) != 2 {
		t.Errorf("GetTop(weekly) returned %v entries, want 2", len(topWeekly))
	}

	// Verify scores are different
	if topWeekly[0].Score == topAllTime[0].Score {
		t.Error("Weekly and all_time scores should be different")
	}
}

func TestGetRank(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Record scores
	players := []struct {
		id    string
		score int64
		rank  int
	}{
		{"p1", 1000, 2}, // Second place
		{"p2", 500, 4},  // Fourth place
		{"p3", 1500, 1}, // First place
		{"p4", 800, 3},  // Third place
	}

	for _, p := range players {
		if err := lb.RecordScore(p.id, "Name", "kills", "all_time", p.score); err != nil {
			t.Fatalf("RecordScore() error = %v", err)
		}
	}

	// Verify ranks
	for _, p := range players {
		rank, err := lb.GetRank(p.id, "kills", "all_time")
		if err != nil {
			t.Errorf("GetRank(%s) error = %v", p.id, err)
			continue
		}
		if rank != p.rank {
			t.Errorf("GetRank(%s) = %v, want %v", p.id, rank, p.rank)
		}
	}
}

func TestGetRankNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	rank, err := lb.GetRank("nonexistent", "kills", "all_time")
	if err != nil {
		t.Errorf("GetRank() should not error for nonexistent player, got %v", err)
	}
	if rank != 0 {
		t.Errorf("GetRank() for nonexistent player = %v, want 0", rank)
	}
}

func TestIncrementScore(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	playerID := "p1"
	playerName := "Alice"
	stat := "kills"
	period := "all_time"

	// Increment from zero
	if err := lb.IncrementScore(playerID, playerName, stat, period, 10); err != nil {
		t.Fatalf("IncrementScore() error = %v", err)
	}

	score, _ := lb.getScore(playerID, stat, period)
	if score != 10 {
		t.Errorf("score after first increment = %v, want 10", score)
	}

	// Increment again
	if err := lb.IncrementScore(playerID, playerName, stat, period, 5); err != nil {
		t.Fatalf("IncrementScore() error = %v", err)
	}

	score, _ = lb.getScore(playerID, stat, period)
	if score != 15 {
		t.Errorf("score after second increment = %v, want 15", score)
	}

	// Negative increment (decrement)
	if err := lb.IncrementScore(playerID, playerName, stat, period, -3); err != nil {
		t.Fatalf("IncrementScore() error = %v", err)
	}

	score, _ = lb.getScore(playerID, stat, period)
	if score != 12 {
		t.Errorf("score after decrement = %v, want 12", score)
	}
}

func TestClearPeriod(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Record scores for different periods
	lb.RecordScore("p1", "Alice", "kills", "all_time", 1000)
	lb.RecordScore("p1", "Alice", "kills", "weekly", 100)
	lb.RecordScore("p2", "Bob", "kills", "all_time", 500)
	lb.RecordScore("p2", "Bob", "kills", "weekly", 200)

	// Clear weekly period
	if err := lb.ClearPeriod("weekly"); err != nil {
		t.Fatalf("ClearPeriod() error = %v", err)
	}

	// Verify weekly scores are gone
	weeklyTop, _ := lb.GetTop("kills", "weekly", 10)
	if len(weeklyTop) != 0 {
		t.Errorf("weekly leaderboard should be empty after clear, got %v entries", len(weeklyTop))
	}

	// Verify all_time scores remain
	allTimeTop, _ := lb.GetTop("kills", "all_time", 10)
	if len(allTimeTop) != 2 {
		t.Errorf("all_time leaderboard should have 2 entries, got %v", len(allTimeTop))
	}
}

func TestClearPeriodEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Clear non-existent period (should not error)
	if err := lb.ClearPeriod("monthly"); err != nil {
		t.Errorf("ClearPeriod() on empty period should not error, got %v", err)
	}
}

func TestMultipleStats(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	playerID := "p1"

	// Record different stats for same player
	lb.RecordScore(playerID, "Alice", "kills", "all_time", 1000)
	lb.RecordScore(playerID, "Alice", "wins", "all_time", 50)
	lb.RecordScore(playerID, "Alice", "high_score", "all_time", 99999)

	// Verify each stat is tracked independently
	killsTop, _ := lb.GetTop("kills", "all_time", 10)
	if len(killsTop) != 1 || killsTop[0].Score != 1000 {
		t.Errorf("kills stat not tracked correctly")
	}

	winsTop, _ := lb.GetTop("wins", "all_time", 10)
	if len(winsTop) != 1 || winsTop[0].Score != 50 {
		t.Errorf("wins stat not tracked correctly")
	}

	highScoreTop, _ := lb.GetTop("high_score", "all_time", 10)
	if len(highScoreTop) != 1 || highScoreTop[0].Score != 99999 {
		t.Errorf("high_score stat not tracked correctly")
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	lb, err := New(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Close should not error
	if err := lb.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Second close should not panic
	if err := lb.Close(); err == nil {
		// SQLite allows multiple closes, so this is expected
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist.db")

	// Create leaderboard and record scores
	lb1, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	lb1.RecordScore("p1", "Alice", "kills", "all_time", 1000)
	lb1.RecordScore("p2", "Bob", "kills", "all_time", 500)
	lb1.Close()

	// Reopen database
	lb2, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() reopening error = %v", err)
	}
	defer lb2.Close()

	// Verify scores persisted
	top, err := lb2.GetTop("kills", "all_time", 10)
	if err != nil {
		t.Fatalf("GetTop() after reopen error = %v", err)
	}

	if len(top) != 2 {
		t.Errorf("persisted leaderboard should have 2 entries, got %v", len(top))
	}

	if top[0].PlayerID != "p1" || top[0].Score != 1000 {
		t.Errorf("persisted score incorrect: got {%v, %v}, want {p1, 1000}",
			top[0].PlayerID, top[0].Score)
	}
}

// BenchmarkRecordScore measures score recording performance.
func BenchmarkRecordScore(b *testing.B) {
	tmpDir := b.TempDir()
	lb, err := New(filepath.Join(tmpDir, "bench.db"))
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.RecordScore("p1", "Alice", "kills", "all_time", int64(i))
	}
}

// BenchmarkGetTop measures top N query performance.
func BenchmarkGetTop(b *testing.B) {
	tmpDir := b.TempDir()
	lb, err := New(filepath.Join(tmpDir, "bench.db"))
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Populate with 1000 entries
	for i := 0; i < 1000; i++ {
		lb.RecordScore(fmt.Sprintf("p%d", i), "Name", "kills", "all_time", int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetTop("kills", "all_time", 10)
	}
}

// BenchmarkGetRank measures rank query performance.
func BenchmarkGetRank(b *testing.B) {
	tmpDir := b.TempDir()
	lb, err := New(filepath.Join(tmpDir, "bench.db"))
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer lb.Close()

	// Populate with 1000 entries
	for i := 0; i < 1000; i++ {
		lb.RecordScore(fmt.Sprintf("p%d", i), "Name", "kills", "all_time", int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lb.GetRank("p500", "kills", "all_time")
	}
}

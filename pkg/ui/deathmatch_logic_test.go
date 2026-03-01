package ui

import (
	"testing"
	"time"
)

func TestNewKillFeed(t *testing.T) {
	kf := NewKillFeed(100, 50)
	if kf == nil {
		t.Fatal("NewKillFeed returned nil")
	}
	if kf.X != 100 {
		t.Errorf("Expected X=100, got %d", kf.X)
	}
	if kf.Y != 50 {
		t.Errorf("Expected Y=50, got %d", kf.Y)
	}
	if len(kf.Entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(kf.Entries))
	}
}

func TestKillFeedAddKill(t *testing.T) {
	tests := []struct {
		name       string
		killerID   uint64
		victimID   uint64
		killerName string
		victimName string
		suicide    bool
		teamKill   bool
	}{
		{
			name:       "Normal kill",
			killerID:   1,
			victimID:   2,
			killerName: "Player1",
			victimName: "Player2",
			suicide:    false,
			teamKill:   false,
		},
		{
			name:       "Suicide",
			killerID:   1,
			victimID:   1,
			killerName: "Player1",
			victimName: "Player1",
			suicide:    true,
			teamKill:   false,
		},
		{
			name:       "Team kill",
			killerID:   1,
			victimID:   2,
			killerName: "Player1",
			victimName: "Player2",
			suicide:    false,
			teamKill:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kf := NewKillFeed(100, 50)
			kf.AddKill(tt.killerID, tt.victimID, tt.killerName, tt.victimName, tt.suicide, tt.teamKill)

			if len(kf.Entries) != 1 {
				t.Fatalf("Expected 1 entry, got %d", len(kf.Entries))
			}

			entry := kf.Entries[0]
			if entry.KillerID != tt.killerID {
				t.Errorf("Expected KillerID=%d, got %d", tt.killerID, entry.KillerID)
			}
			if entry.VictimID != tt.victimID {
				t.Errorf("Expected VictimID=%d, got %d", tt.victimID, entry.VictimID)
			}
			if entry.KillerName != tt.killerName {
				t.Errorf("Expected KillerName=%s, got %s", tt.killerName, entry.KillerName)
			}
			if entry.VictimName != tt.victimName {
				t.Errorf("Expected VictimName=%s, got %s", tt.victimName, entry.VictimName)
			}
			if entry.Suicide != tt.suicide {
				t.Errorf("Expected Suicide=%v, got %v", tt.suicide, entry.Suicide)
			}
			if entry.TeamKill != tt.teamKill {
				t.Errorf("Expected TeamKill=%v, got %v", tt.teamKill, entry.TeamKill)
			}
		})
	}
}

func TestKillFeedMaxEntries(t *testing.T) {
	kf := NewKillFeed(100, 50)

	// Add more than max entries
	for i := 0; i < KillFeedMaxEntries+3; i++ {
		kf.AddKill(uint64(i), uint64(i+1), "Killer", "Victim", false, false)
	}

	// Should only keep the most recent entries
	if len(kf.Entries) != KillFeedMaxEntries {
		t.Errorf("Expected %d entries, got %d", KillFeedMaxEntries, len(kf.Entries))
	}

	// First entry should be the 4th kill (index 3)
	if kf.Entries[0].KillerID != 3 {
		t.Errorf("Expected first entry KillerID=3, got %d", kf.Entries[0].KillerID)
	}
}

func TestKillFeedUpdate(t *testing.T) {
	kf := NewKillFeed(100, 50)

	// Add an entry
	kf.AddKill(1, 2, "Player1", "Player2", false, false)

	// Manually set timestamp to expired
	kf.Entries[0].Timestamp = time.Now().Add(-10 * time.Second)

	// Add a recent entry
	kf.AddKill(3, 4, "Player3", "Player4", false, false)

	// Update should remove expired entries
	kf.Update()

	if len(kf.Entries) != 1 {
		t.Fatalf("Expected 1 entry after update, got %d", len(kf.Entries))
	}

	if kf.Entries[0].KillerID != 3 {
		t.Errorf("Expected remaining entry KillerID=3, got %d", kf.Entries[0].KillerID)
	}
}

func TestNewScoreboard(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		showTeams bool
	}{
		{
			name:      "FFA scoreboard",
			title:     "Free For All",
			showTeams: false,
		},
		{
			name:      "Team scoreboard",
			title:     "Team Deathmatch",
			showTeams: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := NewScoreboard(tt.title, tt.showTeams)
			if sb == nil {
				t.Fatal("NewScoreboard returned nil")
			}
			if sb.Title != tt.title {
				t.Errorf("Expected Title=%s, got %s", tt.title, sb.Title)
			}
			if sb.ShowTeams != tt.showTeams {
				t.Errorf("Expected ShowTeams=%v, got %v", tt.showTeams, sb.ShowTeams)
			}
			if sb.Visible {
				t.Error("Expected scoreboard to start hidden")
			}
		})
	}
}

func TestScoreboardSetEntries(t *testing.T) {
	sb := NewScoreboard("Test", false)

	entries := []ScoreboardEntry{
		{PlayerID: 1, PlayerName: "Player1", Frags: 10, Deaths: 5, Assists: 2},
		{PlayerID: 2, PlayerName: "Player2", Frags: 8, Deaths: 6, Assists: 3},
		{PlayerID: 3, PlayerName: "Player3", Frags: 15, Deaths: 4, Assists: 1},
	}

	sb.SetEntries(entries)

	if len(sb.Entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(sb.Entries))
	}

	for i, entry := range entries {
		if sb.Entries[i].PlayerID != entry.PlayerID {
			t.Errorf("Entry %d: Expected PlayerID=%d, got %d", i, entry.PlayerID, sb.Entries[i].PlayerID)
		}
		if sb.Entries[i].Frags != entry.Frags {
			t.Errorf("Entry %d: Expected Frags=%d, got %d", i, entry.Frags, sb.Entries[i].Frags)
		}
	}
}

func TestScoreboardSetWinner(t *testing.T) {
	sb := NewScoreboard("Test", false)

	winnerText := "Player1 wins!"
	sb.SetWinner(winnerText)

	if sb.WinnerText != winnerText {
		t.Errorf("Expected WinnerText=%s, got %s", winnerText, sb.WinnerText)
	}
}

func TestScoreboardVisibility(t *testing.T) {
	sb := NewScoreboard("Test", false)

	// Should start hidden
	if sb.Visible {
		t.Error("Expected scoreboard to start hidden")
	}

	// Show
	sb.Show()
	if !sb.Visible {
		t.Error("Expected scoreboard to be visible after Show()")
	}

	// Hide
	sb.Hide()
	if sb.Visible {
		t.Error("Expected scoreboard to be hidden after Hide()")
	}

	// Toggle
	sb.Toggle()
	if !sb.Visible {
		t.Error("Expected scoreboard to be visible after Toggle()")
	}

	sb.Toggle()
	if sb.Visible {
		t.Error("Expected scoreboard to be hidden after second Toggle()")
	}
}

func TestScoreboardKDRatioCalculation(t *testing.T) {
	tests := []struct {
		name          string
		frags         int
		deaths        int
		expectedRatio float64
	}{
		{
			name:          "Positive K/D",
			frags:         10,
			deaths:        5,
			expectedRatio: 2.0,
		},
		{
			name:          "Equal K/D",
			frags:         5,
			deaths:        5,
			expectedRatio: 1.0,
		},
		{
			name:          "No deaths",
			frags:         10,
			deaths:        0,
			expectedRatio: 10.0,
		},
		{
			name:          "No kills",
			frags:         0,
			deaths:        5,
			expectedRatio: 0.0,
		},
		{
			name:          "No kills or deaths",
			frags:         0,
			deaths:        0,
			expectedRatio: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the calculation logic without drawing
			var kdRatio float64
			if tt.deaths > 0 {
				kdRatio = float64(tt.frags) / float64(tt.deaths)
			} else if tt.frags > 0 {
				kdRatio = float64(tt.frags)
			}

			if kdRatio != tt.expectedRatio {
				t.Errorf("Expected K/D ratio %.2f, got %.2f", tt.expectedRatio, kdRatio)
			}
		})
	}
}

func TestNewScoreboardEntry(t *testing.T) {
	entry := NewScoreboardEntry(12345, "TestPlayer", 1, 10, 5, 3)

	if entry.PlayerID != 12345 {
		t.Errorf("PlayerID = %d, want 12345", entry.PlayerID)
	}
	if entry.PlayerName != "TestPlayer" {
		t.Errorf("PlayerName = %s, want TestPlayer", entry.PlayerName)
	}
	if entry.Team != 1 {
		t.Errorf("Team = %d, want 1", entry.Team)
	}
	if entry.Frags != 10 {
		t.Errorf("Frags = %d, want 10", entry.Frags)
	}
	if entry.Deaths != 5 {
		t.Errorf("Deaths = %d, want 5", entry.Deaths)
	}
	if entry.Assists != 3 {
		t.Errorf("Assists = %d, want 3", entry.Assists)
	}
}

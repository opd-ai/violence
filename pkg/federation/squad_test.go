package federation

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSquad(t *testing.T) {
	squad := NewSquad("squad1", "Elite Squad", "ELIT", "player1", "Alice")

	if squad.ID != "squad1" {
		t.Errorf("expected ID squad1, got %s", squad.ID)
	}
	if squad.Name != "Elite Squad" {
		t.Errorf("expected name Elite Squad, got %s", squad.Name)
	}
	if squad.Tag != "ELIT" {
		t.Errorf("expected tag ELIT, got %s", squad.Tag)
	}
	if len(squad.Members) != 1 {
		t.Errorf("expected 1 member, got %d", len(squad.Members))
	}

	member, exists := squad.Members["player1"]
	if !exists {
		t.Fatal("founder not added to squad")
	}
	if member.PlayerName != "Alice" {
		t.Errorf("expected player name Alice, got %s", member.PlayerName)
	}
	if !member.IsLeader {
		t.Error("founder should be leader")
	}
}

func TestSquad_Invite(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Squad)
		invitee string
		wantErr error
	}{
		{
			name:    "valid invite",
			setup:   func(s *Squad) {},
			invitee: "player2",
			wantErr: nil,
		},
		{
			name: "invite existing member",
			setup: func(s *Squad) {
				s.Members["player2"] = SquadMember{PlayerID: "player2"}
			},
			invitee: "player2",
			wantErr: ErrAlreadyInSquad,
		},
		{
			name: "squad full",
			setup: func(s *Squad) {
				for i := 1; i < MaxSquadMembers; i++ {
					s.Members[string(rune('a'+i))] = SquadMember{PlayerID: string(rune('a' + i))}
				}
			},
			invitee: "player2",
			wantErr: ErrSquadFull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")
			tt.setup(squad)

			err := squad.Invite(tt.invitee)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

			if err == nil {
				if _, exists := squad.Invites[tt.invitee]; !exists {
					t.Error("invite not created")
				}
			}
		})
	}
}

func TestSquad_Accept(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*Squad)
		playerID   string
		playerName string
		wantErr    error
	}{
		{
			name: "valid accept",
			setup: func(s *Squad) {
				s.Invites["player2"] = SquadInvite{
					SquadID:  s.ID,
					PlayerID: "player2",
					SentAt:   time.Now(),
				}
			},
			playerID:   "player2",
			playerName: "Bob",
			wantErr:    nil,
		},
		{
			name:       "no invite",
			setup:      func(s *Squad) {},
			playerID:   "player2",
			playerName: "Bob",
			wantErr:    ErrNoInvite,
		},
		{
			name: "already member",
			setup: func(s *Squad) {
				s.Invites["player2"] = SquadInvite{SquadID: s.ID, PlayerID: "player2"}
				s.Members["player2"] = SquadMember{PlayerID: "player2"}
			},
			playerID:   "player2",
			playerName: "Bob",
			wantErr:    ErrAlreadyInSquad,
		},
		{
			name: "squad full on accept",
			setup: func(s *Squad) {
				s.Invites["player2"] = SquadInvite{SquadID: s.ID, PlayerID: "player2"}
				for i := 1; i < MaxSquadMembers; i++ {
					s.Members[string(rune('a'+i))] = SquadMember{PlayerID: string(rune('a' + i))}
				}
			},
			playerID:   "player2",
			playerName: "Bob",
			wantErr:    ErrSquadFull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")
			tt.setup(squad)

			err := squad.Accept(tt.playerID, tt.playerName)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

			if err == nil {
				if _, exists := squad.Members[tt.playerID]; !exists {
					t.Error("member not added to squad")
				}
				if _, exists := squad.Invites[tt.playerID]; exists {
					t.Error("invite not removed after accept")
				}
			}
		})
	}
}

func TestSquad_Leave(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Squad)
		playerID string
		wantErr  error
		check    func(*testing.T, *Squad)
	}{
		{
			name: "member leaves",
			setup: func(s *Squad) {
				s.Members["player2"] = SquadMember{PlayerID: "player2", PlayerName: "Bob"}
			},
			playerID: "player2",
			wantErr:  nil,
			check: func(t *testing.T, s *Squad) {
				if _, exists := s.Members["player2"]; exists {
					t.Error("member not removed")
				}
			},
		},
		{
			name:     "not in squad",
			setup:    func(s *Squad) {},
			playerID: "player2",
			wantErr:  ErrNotInSquad,
			check:    func(t *testing.T, s *Squad) {},
		},
		{
			name: "leader leaves, promotes another",
			setup: func(s *Squad) {
				s.Members["player2"] = SquadMember{PlayerID: "player2", PlayerName: "Bob"}
			},
			playerID: "player1",
			wantErr:  nil,
			check: func(t *testing.T, s *Squad) {
				if _, exists := s.Members["player1"]; exists {
					t.Error("leader not removed")
				}
				if !s.Members["player2"].IsLeader {
					t.Error("player2 should be promoted to leader")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")
			tt.setup(squad)

			err := squad.Leave(tt.playerID)
			if err != tt.wantErr {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}

			tt.check(t, squad)
		})
	}
}

func TestSquad_Getters(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")
	squad.Members["player2"] = SquadMember{PlayerID: "player2"}
	squad.Invites["player3"] = SquadInvite{PlayerID: "player3"}

	t.Run("GetMembers", func(t *testing.T) {
		members := squad.GetMembers()
		if len(members) != 2 {
			t.Errorf("expected 2 members, got %d", len(members))
		}
	})

	t.Run("GetMemberCount", func(t *testing.T) {
		count := squad.GetMemberCount()
		if count != 2 {
			t.Errorf("expected count 2, got %d", count)
		}
	})

	t.Run("IsMember", func(t *testing.T) {
		if !squad.IsMember("player1") {
			t.Error("player1 should be a member")
		}
		if squad.IsMember("player99") {
			t.Error("player99 should not be a member")
		}
	})

	t.Run("IsLeader", func(t *testing.T) {
		if !squad.IsLeader("player1") {
			t.Error("player1 should be leader")
		}
		if squad.IsLeader("player2") {
			t.Error("player2 should not be leader")
		}
	})

	t.Run("GetInvites", func(t *testing.T) {
		invites := squad.GetInvites()
		if len(invites) != 1 {
			t.Errorf("expected 1 invite, got %d", len(invites))
		}
	})
}

func TestSquadManager_CreateSquad(t *testing.T) {
	sm := NewSquadManager()

	squad, err := sm.CreateSquad("squad1", "Elite", "ELIT", "player1", "Alice")
	if err != nil {
		t.Fatalf("failed to create squad: %v", err)
	}
	if squad.ID != "squad1" {
		t.Errorf("expected ID squad1, got %s", squad.ID)
	}

	// Try creating duplicate
	_, err = sm.CreateSquad("squad1", "Elite2", "ELI2", "player2", "Bob")
	if err == nil {
		t.Error("expected error when creating duplicate squad")
	}
}

func TestSquadManager_GetSquad(t *testing.T) {
	sm := NewSquadManager()
	sm.CreateSquad("squad1", "Elite", "ELIT", "player1", "Alice")

	squad, err := sm.GetSquad("squad1")
	if err != nil {
		t.Fatalf("failed to get squad: %v", err)
	}
	if squad.Name != "Elite" {
		t.Errorf("expected name Elite, got %s", squad.Name)
	}

	_, err = sm.GetSquad("nonexistent")
	if err != ErrInvalidSquadID {
		t.Errorf("expected ErrInvalidSquadID, got %v", err)
	}
}

func TestSquadManager_DeleteSquad(t *testing.T) {
	sm := NewSquadManager()
	sm.CreateSquad("squad1", "Elite", "ELIT", "player1", "Alice")

	err := sm.DeleteSquad("squad1")
	if err != nil {
		t.Fatalf("failed to delete squad: %v", err)
	}

	_, err = sm.GetSquad("squad1")
	if err != ErrInvalidSquadID {
		t.Error("squad should be deleted")
	}

	err = sm.DeleteSquad("nonexistent")
	if err != ErrInvalidSquadID {
		t.Errorf("expected ErrInvalidSquadID, got %v", err)
	}
}

func TestSquadManager_ListSquads(t *testing.T) {
	sm := NewSquadManager()
	sm.CreateSquad("squad1", "Elite", "ELIT", "player1", "Alice")
	sm.CreateSquad("squad2", "Warriors", "WARR", "player2", "Bob")

	squads := sm.ListSquads()
	if len(squads) != 2 {
		t.Errorf("expected 2 squads, got %d", len(squads))
	}
}

func TestSquadManager_SaveLoad(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "squad_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override home directory for testing
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	sm := NewSquadManager()
	squad1, _ := sm.CreateSquad("squad1", "Elite", "ELIT", "player1", "Alice")
	squad1.Invite("player2")
	squad1.Accept("player2", "Bob")

	// Save
	if err := sm.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load into new manager
	sm2 := NewSquadManager()
	if err := sm2.Load(); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify loaded data
	squad, err := sm2.GetSquad("squad1")
	if err != nil {
		t.Fatalf("failed to get loaded squad: %v", err)
	}
	if squad.Name != "Elite" {
		t.Errorf("expected name Elite, got %s", squad.Name)
	}
	if len(squad.Members) != 2 {
		t.Errorf("expected 2 members, got %d", len(squad.Members))
	}
}

func TestSquadManager_LoadNonExistent(t *testing.T) {
	// Create temp directory for testing
	tmpDir, err := os.MkdirTemp("", "squad_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	sm := NewSquadManager()
	if err := sm.Load(); err != nil {
		t.Errorf("loading non-existent file should not error: %v", err)
	}
}

func TestSquad_ConcurrentAccess(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Test concurrent invites and accepts
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			playerID := string(rune('a' + idx))
			squad.Invite(playerID)
			squad.Accept(playerID, "Player")
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	// Should have at least some members (limited by MaxSquadMembers)
	if squad.GetMemberCount() < 1 {
		t.Error("concurrent operations failed")
	}
}

func TestSquad_MaxMembersLimit(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Fill to max
	for i := 1; i < MaxSquadMembers; i++ {
		playerID := string(rune('a' + i))
		squad.Invite(playerID)
		squad.Accept(playerID, "Player")
	}

	if squad.GetMemberCount() != MaxSquadMembers {
		t.Errorf("expected %d members, got %d", MaxSquadMembers, squad.GetMemberCount())
	}

	// Try to add one more
	err := squad.Invite("overflow")
	if err != ErrSquadFull {
		t.Error("should not allow invites when squad is full")
	}
}

func TestSquadManager_SaveInvalidPath(t *testing.T) {
	sm := NewSquadManager()
	sm.CreateSquad("squad1", "Test", "TEST", "player1", "Alice")

	// Set HOME to an invalid path
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/dev/null/invalid")
	defer os.Setenv("HOME", oldHome)

	// Attempting to save should fail gracefully
	err := sm.Save()
	if err == nil {
		t.Error("expected error when saving to invalid path")
	}
}

func TestSquad_EmptySquadOperations(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Leader leaves when they're the only member
	err := squad.Leave("player1")
	if err != nil {
		t.Errorf("failed to leave: %v", err)
	}

	if len(squad.Members) != 0 {
		t.Error("squad should be empty after leader leaves")
	}
}

func TestSquad_GetSquadSavePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "squad_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	sm := NewSquadManager()
	path, err := sm.getSquadSavePath()
	if err != nil {
		t.Fatalf("failed to get save path: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, ".violence", "squads", "squads.json")
	if path != expectedPath {
		t.Errorf("expected path %s, got %s", expectedPath, path)
	}

	// Verify directory was created
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("squad directory should be created")
	}
}

func TestSquad_GetStats_EmptySquad(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")
	squad.Leave("player1") // Empty the squad

	stats := squad.GetStats()

	if stats.MemberCount != 0 {
		t.Errorf("expected 0 members, got %d", stats.MemberCount)
	}
	if stats.TotalKills != 0 {
		t.Errorf("expected 0 total kills, got %d", stats.TotalKills)
	}
	if stats.TotalDeaths != 0 {
		t.Errorf("expected 0 total deaths, got %d", stats.TotalDeaths)
	}
	if stats.TotalWins != 0 {
		t.Errorf("expected 0 total wins, got %d", stats.TotalWins)
	}
	if stats.TotalPlayTime != 0 {
		t.Errorf("expected 0 total play time, got %v", stats.TotalPlayTime)
	}
}

func TestSquad_GetStats_SingleMember(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Update member stats
	err := squad.SetMemberStats("player1", 100, 50, 10, 5*time.Hour)
	if err != nil {
		t.Fatalf("failed to set member stats: %v", err)
	}

	stats := squad.GetStats()

	if stats.MemberCount != 1 {
		t.Errorf("expected 1 member, got %d", stats.MemberCount)
	}
	if stats.TotalKills != 100 {
		t.Errorf("expected 100 total kills, got %d", stats.TotalKills)
	}
	if stats.TotalDeaths != 50 {
		t.Errorf("expected 50 total deaths, got %d", stats.TotalDeaths)
	}
	if stats.TotalWins != 10 {
		t.Errorf("expected 10 total wins, got %d", stats.TotalWins)
	}
	if stats.TotalPlayTime != 5*time.Hour {
		t.Errorf("expected 5h total play time, got %v", stats.TotalPlayTime)
	}

	// Check averages (same as totals for single member)
	if stats.AvgKills != 100.0 {
		t.Errorf("expected 100.0 avg kills, got %f", stats.AvgKills)
	}
	if stats.AvgDeaths != 50.0 {
		t.Errorf("expected 50.0 avg deaths, got %f", stats.AvgDeaths)
	}
	if stats.AvgWins != 10.0 {
		t.Errorf("expected 10.0 avg wins, got %f", stats.AvgWins)
	}
	if stats.AvgPlayTime != 5*time.Hour {
		t.Errorf("expected 5h avg play time, got %v", stats.AvgPlayTime)
	}
}

func TestSquad_GetStats_MultipleMembers(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Add more members
	squad.Invite("player2")
	squad.Accept("player2", "Player2")
	squad.Invite("player3")
	squad.Accept("player3", "Player3")

	// Set stats for each member
	squad.SetMemberStats("player1", 100, 50, 10, 5*time.Hour)
	squad.SetMemberStats("player2", 80, 40, 8, 4*time.Hour)
	squad.SetMemberStats("player3", 60, 30, 6, 3*time.Hour)

	stats := squad.GetStats()

	if stats.MemberCount != 3 {
		t.Errorf("expected 3 members, got %d", stats.MemberCount)
	}
	if stats.TotalKills != 240 {
		t.Errorf("expected 240 total kills, got %d", stats.TotalKills)
	}
	if stats.TotalDeaths != 120 {
		t.Errorf("expected 120 total deaths, got %d", stats.TotalDeaths)
	}
	if stats.TotalWins != 24 {
		t.Errorf("expected 24 total wins, got %d", stats.TotalWins)
	}
	if stats.TotalPlayTime != 12*time.Hour {
		t.Errorf("expected 12h total play time, got %v", stats.TotalPlayTime)
	}

	// Check averages
	if stats.AvgKills != 80.0 {
		t.Errorf("expected 80.0 avg kills, got %f", stats.AvgKills)
	}
	if stats.AvgDeaths != 40.0 {
		t.Errorf("expected 40.0 avg deaths, got %f", stats.AvgDeaths)
	}
	if stats.AvgWins != 8.0 {
		t.Errorf("expected 8.0 avg wins, got %f", stats.AvgWins)
	}
	if stats.AvgPlayTime != 4*time.Hour {
		t.Errorf("expected 4h avg play time, got %v", stats.AvgPlayTime)
	}
}

func TestSquad_UpdateMemberStats(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Initial stats
	err := squad.SetMemberStats("player1", 100, 50, 10, 5*time.Hour)
	if err != nil {
		t.Fatalf("failed to set member stats: %v", err)
	}

	// Update with incremental values
	err = squad.UpdateMemberStats("player1", 20, 5, 2, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to update member stats: %v", err)
	}

	memberStats, err := squad.GetMemberStats("player1")
	if err != nil {
		t.Fatalf("failed to get member stats: %v", err)
	}

	if memberStats.TotalKills != 120 {
		t.Errorf("expected 120 kills, got %d", memberStats.TotalKills)
	}
	if memberStats.TotalDeaths != 55 {
		t.Errorf("expected 55 deaths, got %d", memberStats.TotalDeaths)
	}
	if memberStats.TotalWins != 12 {
		t.Errorf("expected 12 wins, got %d", memberStats.TotalWins)
	}
	if memberStats.PlayTime != 6*time.Hour {
		t.Errorf("expected 6h play time, got %v", memberStats.PlayTime)
	}
}

func TestSquad_UpdateMemberStats_NonExistent(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	err := squad.UpdateMemberStats("nonexistent", 10, 5, 1, 1*time.Hour)
	if err != ErrNotInSquad {
		t.Errorf("expected ErrNotInSquad, got %v", err)
	}
}

func TestSquad_SetMemberStats(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Set initial stats
	err := squad.SetMemberStats("player1", 100, 50, 10, 5*time.Hour)
	if err != nil {
		t.Fatalf("failed to set member stats: %v", err)
	}

	// Replace with new stats
	err = squad.SetMemberStats("player1", 200, 100, 20, 10*time.Hour)
	if err != nil {
		t.Fatalf("failed to set member stats: %v", err)
	}

	memberStats, err := squad.GetMemberStats("player1")
	if err != nil {
		t.Fatalf("failed to get member stats: %v", err)
	}

	if memberStats.TotalKills != 200 {
		t.Errorf("expected 200 kills, got %d", memberStats.TotalKills)
	}
	if memberStats.TotalDeaths != 100 {
		t.Errorf("expected 100 deaths, got %d", memberStats.TotalDeaths)
	}
	if memberStats.TotalWins != 20 {
		t.Errorf("expected 20 wins, got %d", memberStats.TotalWins)
	}
	if memberStats.PlayTime != 10*time.Hour {
		t.Errorf("expected 10h play time, got %v", memberStats.PlayTime)
	}
}

func TestSquad_SetMemberStats_NonExistent(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	err := squad.SetMemberStats("nonexistent", 100, 50, 10, 5*time.Hour)
	if err != ErrNotInSquad {
		t.Errorf("expected ErrNotInSquad, got %v", err)
	}
}

func TestSquad_GetMemberStats(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// New members start with zero stats
	memberStats, err := squad.GetMemberStats("player1")
	if err != nil {
		t.Fatalf("failed to get member stats: %v", err)
	}

	if memberStats.TotalKills != 0 {
		t.Errorf("expected 0 kills, got %d", memberStats.TotalKills)
	}
	if memberStats.TotalDeaths != 0 {
		t.Errorf("expected 0 deaths, got %d", memberStats.TotalDeaths)
	}
	if memberStats.TotalWins != 0 {
		t.Errorf("expected 0 wins, got %d", memberStats.TotalWins)
	}
	if memberStats.PlayTime != 0 {
		t.Errorf("expected 0 play time, got %v", memberStats.PlayTime)
	}
}

func TestSquad_GetMemberStats_NonExistent(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	_, err := squad.GetMemberStats("nonexistent")
	if err != ErrNotInSquad {
		t.Errorf("expected ErrNotInSquad, got %v", err)
	}
}

func TestSquad_StatsPersistence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "squad_stats_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create squad manager and squad
	sm := NewSquadManager()
	squad, err := sm.CreateSquad("squad1", "Test", "TEST", "player1", "Leader")
	if err != nil {
		t.Fatalf("failed to create squad: %v", err)
	}

	// Set stats
	squad.SetMemberStats("player1", 100, 50, 10, 5*time.Hour)

	// Save
	if err := sm.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load into new manager
	sm2 := NewSquadManager()
	if err := sm2.Load(); err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify stats persisted
	loadedSquad, err := sm2.GetSquad("squad1")
	if err != nil {
		t.Fatalf("failed to get squad: %v", err)
	}

	memberStats, err := loadedSquad.GetMemberStats("player1")
	if err != nil {
		t.Fatalf("failed to get member stats: %v", err)
	}

	if memberStats.TotalKills != 100 {
		t.Errorf("expected 100 kills, got %d", memberStats.TotalKills)
	}
	if memberStats.TotalDeaths != 50 {
		t.Errorf("expected 50 deaths, got %d", memberStats.TotalDeaths)
	}
	if memberStats.TotalWins != 10 {
		t.Errorf("expected 10 wins, got %d", memberStats.TotalWins)
	}
	if memberStats.PlayTime != 5*time.Hour {
		t.Errorf("expected 5h play time, got %v", memberStats.PlayTime)
	}
}

func TestSquad_ConcurrentStatsUpdate(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Concurrent stat updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			squad.UpdateMemberStats("player1", 1, 0, 0, 1*time.Second)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	memberStats, err := squad.GetMemberStats("player1")
	if err != nil {
		t.Fatalf("failed to get member stats: %v", err)
	}

	if memberStats.TotalKills != 10 {
		t.Errorf("expected 10 kills from concurrent updates, got %d", memberStats.TotalKills)
	}
	if memberStats.PlayTime != 10*time.Second {
		t.Errorf("expected 10s play time from concurrent updates, got %v", memberStats.PlayTime)
	}
}

func TestSquad_StatsAfterMemberLeaves(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Add member with stats
	squad.Invite("player2")
	squad.Accept("player2", "Player2")
	squad.SetMemberStats("player1", 100, 50, 10, 5*time.Hour)
	squad.SetMemberStats("player2", 80, 40, 8, 4*time.Hour)

	// Verify total stats
	stats := squad.GetStats()
	if stats.TotalKills != 180 {
		t.Errorf("expected 180 total kills, got %d", stats.TotalKills)
	}

	// Member leaves
	squad.Leave("player2")

	// Stats should reflect only remaining member
	stats = squad.GetStats()
	if stats.MemberCount != 1 {
		t.Errorf("expected 1 member, got %d", stats.MemberCount)
	}
	if stats.TotalKills != 100 {
		t.Errorf("expected 100 total kills after member leaves, got %d", stats.TotalKills)
	}
	if stats.TotalDeaths != 50 {
		t.Errorf("expected 50 total deaths after member leaves, got %d", stats.TotalDeaths)
	}
}

func TestSquad_ZeroStats(t *testing.T) {
	squad := NewSquad("squad1", "Test", "TEST", "player1", "Leader")

	// Don't set any stats - verify defaults
	stats := squad.GetStats()

	if stats.MemberCount != 1 {
		t.Errorf("expected 1 member, got %d", stats.MemberCount)
	}
	if stats.TotalKills != 0 {
		t.Errorf("expected 0 total kills, got %d", stats.TotalKills)
	}
	if stats.AvgKills != 0.0 {
		t.Errorf("expected 0.0 avg kills, got %f", stats.AvgKills)
	}
}

func TestSquad_GetTag(t *testing.T) {
	squad := NewSquad("squad1", "Test Squad", "ABCD", "player1", "Alice")

	tag := squad.GetTag()
	if tag != "ABCD" {
		t.Errorf("expected tag ABCD, got %s", tag)
	}
}

func TestSquad_SetTag(t *testing.T) {
	squad := NewSquad("squad1", "Test Squad", "OLD", "player1", "Alice")

	squad.SetTag("NEW")

	tag := squad.GetTag()
	if tag != "NEW" {
		t.Errorf("expected tag NEW, got %s", tag)
	}
}

func TestSquad_SetTag_TruncatesLongTag(t *testing.T) {
	squad := NewSquad("squad1", "Test Squad", "ORIG", "player1", "Alice")

	// Set a tag longer than MaxTagLength (4)
	squad.SetTag("TOOLONG")

	tag := squad.GetTag()
	if tag != "TOOL" {
		t.Errorf("expected truncated tag TOOL, got %s", tag)
	}
}

func TestSquad_SetTag_ExactlyFourCharacters(t *testing.T) {
	squad := NewSquad("squad1", "Test Squad", "OLD", "player1", "Alice")

	squad.SetTag("ABCD")

	tag := squad.GetTag()
	if tag != "ABCD" {
		t.Errorf("expected tag ABCD, got %s", tag)
	}
}

func TestSquad_SetTag_ThreeCharacters(t *testing.T) {
	squad := NewSquad("squad1", "Test Squad", "OLD", "player1", "Alice")

	squad.SetTag("ABC")

	tag := squad.GetTag()
	if tag != "ABC" {
		t.Errorf("expected tag ABC, got %s", tag)
	}
}

func TestSquad_SetTag_EmptyString(t *testing.T) {
	squad := NewSquad("squad1", "Test Squad", "OLD", "player1", "Alice")

	squad.SetTag("")

	tag := squad.GetTag()
	if tag != "" {
		t.Errorf("expected empty tag, got %s", tag)
	}
}

func TestNewSquad_TruncatesLongTag(t *testing.T) {
	// Create squad with tag longer than MaxTagLength
	squad := NewSquad("squad1", "Test Squad", "VERYLONGTAG", "player1", "Alice")

	if squad.Tag != "VERY" {
		t.Errorf("expected truncated tag VERY, got %s", squad.Tag)
	}
}

func TestSquad_GetName(t *testing.T) {
	squad := NewSquad("squad1", "Elite Squad", "ELIT", "player1", "Alice")

	name := squad.GetName()
	if name != "Elite Squad" {
		t.Errorf("expected name 'Elite Squad', got %s", name)
	}
}

func TestSquad_GetID(t *testing.T) {
	squad := NewSquad("squad1", "Elite Squad", "ELIT", "player1", "Alice")

	id := squad.GetID()
	if id != "squad1" {
		t.Errorf("expected ID 'squad1', got %s", id)
	}
}

func TestMaxTagLength_Constant(t *testing.T) {
	if MaxTagLength != 4 {
		t.Errorf("MaxTagLength should be 4, got %d", MaxTagLength)
	}
}

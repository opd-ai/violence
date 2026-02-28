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

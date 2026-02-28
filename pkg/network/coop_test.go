package network

import (
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/inventory"
)

func TestNewCoopSession(t *testing.T) {
	tests := []struct {
		name        string
		sessionID   string
		maxPlayers  int
		levelSeed   uint64
		expectError bool
	}{
		{"valid 2 players", "session1", 2, 12345, false},
		{"valid 3 players", "session2", 3, 67890, false},
		{"valid 4 players", "session3", 4, 11111, false},
		{"invalid too few", "session4", 1, 22222, true},
		{"invalid too many", "session5", 5, 33333, true},
		{"invalid zero", "session6", 0, 44444, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := NewCoopSession(tt.sessionID, tt.maxPlayers, tt.levelSeed)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if session.SessionID != tt.sessionID {
				t.Errorf("session ID = %s, want %s", session.SessionID, tt.sessionID)
			}

			if session.MaxPlayers != tt.maxPlayers {
				t.Errorf("max players = %d, want %d", session.MaxPlayers, tt.maxPlayers)
			}

			if session.LevelSeed != tt.levelSeed {
				t.Errorf("level seed = %d, want %d", session.LevelSeed, tt.levelSeed)
			}

			if session.Players == nil {
				t.Error("players map is nil")
			}

			if session.World == nil {
				t.Error("world is nil")
			}

			if session.QuestTracker == nil {
				t.Error("quest tracker is nil")
			}

			if session.Started {
				t.Error("session should not be started initially")
			}
		})
	}
}

func TestCoopSession_AddPlayer(t *testing.T) {
	tests := []struct {
		name        string
		maxPlayers  int
		playerIDs   []uint64
		expectError []bool
	}{
		{
			name:        "add single player",
			maxPlayers:  4,
			playerIDs:   []uint64{1},
			expectError: []bool{false},
		},
		{
			name:        "add multiple players",
			maxPlayers:  4,
			playerIDs:   []uint64{1, 2, 3},
			expectError: []bool{false, false, false},
		},
		{
			name:        "add duplicate player",
			maxPlayers:  4,
			playerIDs:   []uint64{1, 1},
			expectError: []bool{false, true},
		},
		{
			name:        "exceed max players",
			maxPlayers:  2,
			playerIDs:   []uint64{1, 2, 3},
			expectError: []bool{false, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := NewCoopSession("test", tt.maxPlayers, 12345)
			if err != nil {
				t.Fatalf("failed to create session: %v", err)
			}

			for i, playerID := range tt.playerIDs {
				err := session.AddPlayer(playerID)

				if tt.expectError[i] {
					if err == nil {
						t.Errorf("player %d: expected error but got none", playerID)
					}
				} else {
					if err != nil {
						t.Errorf("player %d: unexpected error: %v", playerID, err)
					}

					// Verify player was added correctly
					player, getErr := session.GetPlayer(playerID)
					if getErr != nil {
						t.Errorf("player %d: failed to get player: %v", playerID, getErr)
						continue
					}

					if player.PlayerID != playerID {
						t.Errorf("player ID = %d, want %d", player.PlayerID, playerID)
					}

					if !player.Active {
						t.Error("player should be active after adding")
					}

					if player.Inventory == nil {
						t.Error("player inventory is nil")
					}

					if player.Health != 100.0 {
						t.Errorf("player health = %f, want 100.0", player.Health)
					}

					if player.MaxHealth != 100.0 {
						t.Errorf("player max health = %f, want 100.0", player.MaxHealth)
					}
				}
			}
		})
	}
}

func TestCoopSession_RemovePlayer(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add players
	playerIDs := []uint64{1, 2, 3}
	for _, id := range playerIDs {
		if err := session.AddPlayer(id); err != nil {
			t.Fatalf("failed to add player %d: %v", id, err)
		}
	}

	tests := []struct {
		name        string
		playerID    uint64
		expectError bool
	}{
		{"remove existing player", 2, false},
		{"remove non-existent player", 999, true},
		{"remove already removed player", 2, false}, // Should succeed (already inactive)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := session.RemovePlayer(tt.playerID)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify player is marked inactive
			if !tt.expectError && tt.playerID != 999 {
				player, _ := session.GetPlayer(tt.playerID)
				if player != nil && player.Active {
					t.Error("player should be marked inactive after removal")
				}
			}
		})
	}
}

func TestCoopSession_GetActivePlayers(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add 3 players
	for i := uint64(1); i <= 3; i++ {
		if err := session.AddPlayer(i); err != nil {
			t.Fatalf("failed to add player %d: %v", i, err)
		}
	}

	// Initially all should be active
	active := session.GetActivePlayers()
	if len(active) != 3 {
		t.Errorf("active players = %d, want 3", len(active))
	}

	// Remove one player
	if err := session.RemovePlayer(2); err != nil {
		t.Fatalf("failed to remove player: %v", err)
	}

	// Now should have 2 active
	active = session.GetActivePlayers()
	if len(active) != 2 {
		t.Errorf("active players after removal = %d, want 2", len(active))
	}

	// Verify correct players are active
	activeIDs := make(map[uint64]bool)
	for _, p := range active {
		activeIDs[p.PlayerID] = true
	}

	if !activeIDs[1] || !activeIDs[3] || activeIDs[2] {
		t.Error("incorrect active player set")
	}
}

func TestCoopSession_ObjectiveProgress(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Generate objectives
	session.QuestTracker.Generate(12345, 3)

	if len(session.QuestTracker.Objectives) != 3 {
		t.Fatalf("expected 3 objectives, got %d", len(session.QuestTracker.Objectives))
	}

	objID := session.QuestTracker.Objectives[0].ID

	// Update progress
	session.UpdateObjectiveProgress(objID, 5)

	// Verify progress
	obj := session.QuestTracker.Objectives[0]
	if obj.Progress != 5 {
		t.Errorf("objective progress = %d, want 5", obj.Progress)
	}

	// Complete objective
	session.CompleteObjective(objID)

	if !session.QuestTracker.Objectives[0].Complete {
		t.Error("objective should be marked complete")
	}
}

func TestCoopSession_Start(t *testing.T) {
	tests := []struct {
		name         string
		playerCount  int
		expectError  bool
		errorMessage string
	}{
		{"enough players", 2, false, ""},
		{"not enough players", 1, true, "not enough players"},
		{"max players", 4, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := NewCoopSession("test", 4, 12345)
			if err != nil {
				t.Fatalf("failed to create session: %v", err)
			}

			// Add players
			for i := uint64(1); i <= uint64(tt.playerCount); i++ {
				if err := session.AddPlayer(i); err != nil {
					t.Fatalf("failed to add player %d: %v", i, err)
				}
			}

			// Try to start
			err = session.Start()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !session.Started {
				t.Error("session should be marked as started")
			}

			// Verify objectives were generated
			if len(session.QuestTracker.Objectives) == 0 {
				t.Error("no objectives generated")
			}

			// Try to start again (should fail)
			err = session.Start()
			if err == nil {
				t.Error("starting already started session should fail")
			}
		})
	}
}

func TestCoopSession_CanStart(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if session.CanStart() {
		t.Error("session with 0 players should not be startable")
	}

	// Add 1 player (not enough)
	session.AddPlayer(1)
	if session.CanStart() {
		t.Error("session with 1 player should not be startable")
	}

	// Add 2nd player (minimum)
	session.AddPlayer(2)
	if !session.CanStart() {
		t.Error("session with 2 players should be startable")
	}

	// Add more players
	session.AddPlayer(3)
	session.AddPlayer(4)
	if !session.CanStart() {
		t.Error("session with 4 players should be startable")
	}
}

func TestCoopSession_IsFull(t *testing.T) {
	session, err := NewCoopSession("test", 2, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	if session.IsFull() {
		t.Error("empty session should not be full")
	}

	session.AddPlayer(1)
	if session.IsFull() {
		t.Error("session with 1/2 players should not be full")
	}

	session.AddPlayer(2)
	if !session.IsFull() {
		t.Error("session with 2/2 players should be full")
	}
}

func TestCoopSession_UpdatePlayerPosition(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)

	tests := []struct {
		name        string
		playerID    uint64
		x           float64
		y           float64
		expectError bool
	}{
		{"valid update", 1, 10.5, 20.3, false},
		{"invalid player", 999, 5.0, 5.0, true},
		{"update again", 1, 100.0, 200.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := session.UpdatePlayerPosition(tt.playerID, tt.x, tt.y)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			player, _ := session.GetPlayer(tt.playerID)
			if player.PosX != tt.x || player.PosY != tt.y {
				t.Errorf("position = (%f, %f), want (%f, %f)", player.PosX, player.PosY, tt.x, tt.y)
			}
		})
	}
}

func TestCoopSession_UpdatePlayerHealth(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)

	tests := []struct {
		name        string
		playerID    uint64
		health      float64
		expected    float64
		expectError bool
	}{
		{"reduce health", 1, 50.0, 50.0, false},
		{"negative health clamped", 1, -10.0, 0.0, false},
		{"increase health", 1, 75.5, 75.5, false},
		{"invalid player", 999, 50.0, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := session.UpdatePlayerHealth(tt.playerID, tt.health)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			player, _ := session.GetPlayer(tt.playerID)
			if player.Health != tt.expected {
				t.Errorf("health = %f, want %f", player.Health, tt.expected)
			}
		})
	}
}

func TestCoopSession_IsLevelComplete(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Generate objectives with known seed
	session.QuestTracker.Generate(12345, 3)

	if session.IsLevelComplete() {
		t.Error("level should not be complete initially")
	}

	// Complete all main objectives
	for i := range session.QuestTracker.Objectives {
		session.QuestTracker.Objectives[i].Complete = true
	}

	if !session.IsLevelComplete() {
		t.Error("level should be complete after all objectives done")
	}
}

func TestCoopSession_IndependentInventories(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add two players
	session.AddPlayer(1)
	session.AddPlayer(2)

	player1, _ := session.GetPlayer(1)
	player2, _ := session.GetPlayer(2)

	// Add item to player 1's inventory
	player1.Inventory.Add(inventory.Item{ID: "ammo", Name: "Ammo", Qty: 10})

	// Verify player 2 doesn't have it
	if player2.Inventory.Has("ammo") {
		t.Error("player 2 should not have player 1's items")
	}

	// Add different item to player 2
	player2.Inventory.Add(inventory.Item{ID: "medkit", Name: "Medkit", Qty: 5})

	// Verify player 1 doesn't have it
	if player1.Inventory.Has("medkit") {
		t.Error("player 1 should not have player 2's items")
	}

	// Verify each has their own items
	if !player1.Inventory.Has("ammo") {
		t.Error("player 1 should still have ammo")
	}

	if !player2.Inventory.Has("medkit") {
		t.Error("player 2 should still have medkit")
	}
}

func TestCoopSession_SetGenre(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			session.SetGenre(genre)

			if session.World.GetGenre() != genre {
				t.Errorf("world genre = %s, want %s", session.World.GetGenre(), genre)
			}
		})
	}
}

func TestCoopSession_ConcurrentAccess(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add players concurrently
	done := make(chan bool, 4)

	for i := uint64(1); i <= 4; i++ {
		playerID := i
		go func() {
			err := session.AddPlayer(playerID)
			if err != nil && playerID <= 4 {
				t.Errorf("failed to add player %d: %v", playerID, err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for concurrent adds")
		}
	}

	// Verify all were added
	if session.GetPlayerCount() != 4 {
		t.Errorf("player count = %d, want 4", session.GetPlayerCount())
	}
}

func TestCoopPlayerState_ThreadSafety(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)

	// Update player state concurrently
	done := make(chan bool, 100)

	for i := 0; i < 50; i++ {
		go func() {
			session.UpdatePlayerPosition(1, 100.0, 200.0)
			done <- true
		}()
		go func() {
			session.UpdatePlayerHealth(1, 50.0)
			done <- true
		}()
	}

	// Wait for all updates
	for i := 0; i < 100; i++ {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for concurrent updates")
		}
	}

	// Verify no crashes and final state is consistent
	player, _ := session.GetPlayer(1)
	if player == nil {
		t.Fatal("player should still exist")
	}
}

func TestCoopSession_OnPlayerDeath(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)
	session.AddPlayer(2)

	tests := []struct {
		name        string
		playerID    uint64
		expectError bool
	}{
		{"valid player death", 1, false},
		{"invalid player", 999, true},
		{"death again", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := session.OnPlayerDeath(tt.playerID)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			player, _ := session.GetPlayer(tt.playerID)
			if !player.Dead {
				t.Error("player should be marked dead")
			}
			if player.Health != 0 {
				t.Errorf("dead player health = %f, want 0", player.Health)
			}
			if player.BleedoutEndTime.IsZero() {
				t.Error("bleedout timer should be set")
			}
		})
	}
}

func TestCoopSession_ProcessBleedouts(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)
	session.AddPlayer(2)
	session.AddPlayer(3)

	// Kill player 1 (should not be in respawn list yet)
	session.OnPlayerDeath(1)

	toRespawn := session.ProcessBleedouts()
	if len(toRespawn) != 0 {
		t.Errorf("should not have any players to respawn yet, got %d", len(toRespawn))
	}

	// Manually expire bleedout timer
	player1, _ := session.GetPlayer(1)
	player1.mu.Lock()
	player1.BleedoutEndTime = time.Now().Add(-1 * time.Second)
	player1.mu.Unlock()

	toRespawn = session.ProcessBleedouts()
	if len(toRespawn) != 1 {
		t.Errorf("should have 1 player to respawn, got %d", len(toRespawn))
	}
	if toRespawn[0] != 1 {
		t.Errorf("wrong player to respawn: got %d, want 1", toRespawn[0])
	}

	// Kill player 2 and expire immediately
	session.OnPlayerDeath(2)
	player2, _ := session.GetPlayer(2)
	player2.mu.Lock()
	player2.BleedoutEndTime = time.Now().Add(-1 * time.Second)
	player2.mu.Unlock()

	toRespawn = session.ProcessBleedouts()
	if len(toRespawn) != 2 {
		t.Errorf("should have 2 players to respawn, got %d", len(toRespawn))
	}
}

func TestCoopSession_RespawnPlayer(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add players at different positions
	session.AddPlayer(1)
	session.AddPlayer(2)
	session.AddPlayer(3)

	session.UpdatePlayerPosition(1, 0, 0)
	session.UpdatePlayerPosition(2, 100, 100)
	session.UpdatePlayerPosition(3, 200, 200)

	// Kill player 1
	session.OnPlayerDeath(1)

	// Respawn should succeed (teammates alive)
	err = session.RespawnPlayer(1)
	if err != nil {
		t.Fatalf("respawn failed: %v", err)
	}

	player1, _ := session.GetPlayer(1)
	if player1.Dead {
		t.Error("player should be alive after respawn")
	}
	if player1.Health != player1.MaxHealth {
		t.Errorf("health = %f, want %f", player1.Health, player1.MaxHealth)
	}
	if player1.Armor != 0 {
		t.Errorf("armor = %f, want 0", player1.Armor)
	}

	// Position should be near a teammate (player 2 or 3)
	if (player1.PosX != 100 || player1.PosY != 100) && (player1.PosX != 200 || player1.PosY != 200) {
		t.Errorf("respawn position (%f, %f) not near any teammate", player1.PosX, player1.PosY)
	}
}

func TestCoopSession_RespawnPlayer_NoTeammates(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)
	session.AddPlayer(2)

	// Kill both players
	session.OnPlayerDeath(1)
	session.OnPlayerDeath(2)

	// Respawn should fail (party wipe)
	err = session.RespawnPlayer(1)
	if err == nil {
		t.Error("expected error for party wipe, got none")
	}
}

func TestCoopSession_IsPartyWiped(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)
	session.AddPlayer(2)
	session.AddPlayer(3)

	if session.isPartyWiped() {
		t.Error("party should not be wiped initially")
	}

	// Kill two players
	session.OnPlayerDeath(1)
	session.OnPlayerDeath(2)

	if session.isPartyWiped() {
		t.Error("party should not be wiped with one alive")
	}

	// Kill last player
	session.OnPlayerDeath(3)

	if !session.isPartyWiped() {
		t.Error("party should be wiped with all dead")
	}
}

func TestCoopSession_RestartLevel(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)
	session.AddPlayer(2)
	session.Start()

	// Kill both players
	session.OnPlayerDeath(1)
	session.OnPlayerDeath(2)

	// Update positions and health
	session.UpdatePlayerPosition(1, 50, 50)
	session.UpdatePlayerPosition(2, 100, 100)

	// Restart level
	err = session.RestartLevel()
	if err != nil {
		t.Fatalf("restart failed: %v", err)
	}

	// Verify all players reset
	for _, playerID := range []uint64{1, 2} {
		player, _ := session.GetPlayer(playerID)
		if player.Dead {
			t.Errorf("player %d should be alive after restart", playerID)
		}
		if player.Health != player.MaxHealth {
			t.Errorf("player %d health = %f, want %f", playerID, player.Health, player.MaxHealth)
		}
		if player.PosX != 0 || player.PosY != 0 {
			t.Errorf("player %d position = (%f, %f), want (0, 0)", playerID, player.PosX, player.PosY)
		}
		if !player.BleedoutEndTime.IsZero() {
			t.Errorf("player %d bleedout timer should be cleared", playerID)
		}
	}

	if session.LevelCompleted {
		t.Error("level should not be completed after restart")
	}
}

func TestCoopSession_IsPlayerDead(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)

	// Player should be alive initially
	dead, err := session.IsPlayerDead(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dead {
		t.Error("player should be alive initially")
	}

	// Kill player
	session.OnPlayerDeath(1)

	dead, err = session.IsPlayerDead(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dead {
		t.Error("player should be dead after OnPlayerDeath")
	}

	// Invalid player
	_, err = session.IsPlayerDead(999)
	if err == nil {
		t.Error("expected error for invalid player")
	}
}

func TestCoopSession_FindNearestLivingTeammate(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)
	session.AddPlayer(2)
	session.AddPlayer(3)
	session.AddPlayer(4)

	session.UpdatePlayerPosition(1, 0, 0)
	session.UpdatePlayerPosition(2, 10, 10)   // Nearest
	session.UpdatePlayerPosition(3, 100, 100) // Far
	session.UpdatePlayerPosition(4, 200, 200) // Farther

	// Kill player 1
	session.OnPlayerDeath(1)

	x, y, found := session.findNearestLivingTeammate(1)
	if !found {
		t.Fatal("should find nearest teammate")
	}

	// Should be player 2's position (nearest)
	if x != 10 || y != 10 {
		t.Errorf("nearest position = (%f, %f), want (10, 10)", x, y)
	}

	// Kill player 2, should find player 3 now
	session.OnPlayerDeath(2)

	x, y, found = session.findNearestLivingTeammate(1)
	if !found {
		t.Fatal("should find next nearest teammate")
	}
	if x != 100 || y != 100 {
		t.Errorf("nearest position = (%f, %f), want (100, 100)", x, y)
	}
}

func TestCoopSession_RespawnPlayer_InvalidPlayer(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	err = session.RespawnPlayer(999)
	if err == nil {
		t.Error("expected error for invalid player")
	}
}

func TestCoopSession_BleedoutDuration(t *testing.T) {
	session, err := NewCoopSession("test", 4, 12345)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	session.AddPlayer(1)

	beforeDeath := time.Now()
	session.OnPlayerDeath(1)
	afterDeath := time.Now()

	player, _ := session.GetPlayer(1)

	// Bleedout end time should be approximately 10 seconds from now
	expectedEnd := beforeDeath.Add(BleedoutDuration)
	latestEnd := afterDeath.Add(BleedoutDuration)

	if player.BleedoutEndTime.Before(expectedEnd) || player.BleedoutEndTime.After(latestEnd.Add(100*time.Millisecond)) {
		t.Errorf("bleedout end time not in expected range: got %v", player.BleedoutEndTime)
	}
}

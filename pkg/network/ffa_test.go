package network

import (
	"testing"
	"time"
)

func TestNewFFAMatch(t *testing.T) {
	tests := []struct {
		name      string
		matchID   string
		fragLimit int
		timeLimit time.Duration
		seed      uint64
		wantFrags int
		wantTime  time.Duration
	}{
		{
			name:      "default limits",
			matchID:   "match1",
			fragLimit: 0,
			timeLimit: 0,
			seed:      12345,
			wantFrags: DefaultFragLimit,
			wantTime:  DefaultTimeLimit,
		},
		{
			name:      "custom limits",
			matchID:   "match2",
			fragLimit: 50,
			timeLimit: 5 * time.Minute,
			seed:      67890,
			wantFrags: 50,
			wantTime:  5 * time.Minute,
		},
		{
			name:      "negative limits use defaults",
			matchID:   "match3",
			fragLimit: -10,
			timeLimit: -1 * time.Second,
			seed:      99999,
			wantFrags: DefaultFragLimit,
			wantTime:  DefaultTimeLimit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := NewFFAMatch(tt.matchID, tt.fragLimit, tt.timeLimit, tt.seed)
			if err != nil {
				t.Fatalf("NewFFAMatch() error = %v", err)
			}
			if match.MatchID != tt.matchID {
				t.Errorf("MatchID = %v, want %v", match.MatchID, tt.matchID)
			}
			if match.FragLimit != tt.wantFrags {
				t.Errorf("FragLimit = %v, want %v", match.FragLimit, tt.wantFrags)
			}
			if match.TimeLimit != tt.wantTime {
				t.Errorf("TimeLimit = %v, want %v", match.TimeLimit, tt.wantTime)
			}
			if match.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", match.Seed, tt.seed)
			}
			if match.Started {
				t.Error("Started should be false initially")
			}
			if match.Finished {
				t.Error("Finished should be false initially")
			}
		})
	}
}

func TestFFAMatch_AddPlayer(t *testing.T) {
	tests := []struct {
		name        string
		playerCount int
		playerID    uint64
		wantErr     bool
	}{
		{
			name:        "first player",
			playerCount: 0,
			playerID:    1,
			wantErr:     false,
		},
		{
			name:        "second player",
			playerCount: 1,
			playerID:    2,
			wantErr:     false,
		},
		{
			name:        "max players",
			playerCount: MaxFFAPlayers - 1,
			playerID:    999,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)

			// Add existing players
			for i := 0; i < tt.playerCount; i++ {
				match.AddPlayer(uint64(100 + i))
			}

			err := match.AddPlayer(tt.playerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPlayer() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if _, exists := match.Players[tt.playerID]; !exists {
					t.Error("Player not added to match")
				}
			}
		})
	}
}

func TestFFAMatch_AddPlayer_Errors(t *testing.T) {
	t.Run("duplicate player", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
		match.AddPlayer(1)
		err := match.AddPlayer(1)
		if err == nil {
			t.Error("Expected error for duplicate player")
		}
	})

	t.Run("match full", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
		for i := 0; i < MaxFFAPlayers; i++ {
			match.AddPlayer(uint64(i + 1))
		}
		err := match.AddPlayer(999)
		if err == nil {
			t.Error("Expected error when match is full")
		}
	})

	t.Run("match already started", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
		match.AddPlayer(1)
		match.AddPlayer(2)
		match.StartMatch()
		err := match.AddPlayer(3)
		if err == nil {
			t.Error("Expected error when adding to started match")
		}
	})
}

func TestFFAMatch_RemovePlayer(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)

	err := match.RemovePlayer(1)
	if err != nil {
		t.Fatalf("RemovePlayer() error = %v", err)
	}

	player := match.Players[1]
	if player.Active {
		t.Error("Player should be inactive after removal")
	}
}

func TestFFAMatch_RemovePlayer_NotInMatch(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	err := match.RemovePlayer(999)
	if err == nil {
		t.Error("Expected error when removing non-existent player")
	}
}

func TestFFAMatch_GenerateSpawnPoints(t *testing.T) {
	tests := []struct {
		name      string
		count     int
		mapWidth  float64
		mapHeight float64
		seed      uint64
	}{
		{
			name:      "4 spawn points",
			count:     4,
			mapWidth:  1000.0,
			mapHeight: 1000.0,
			seed:      12345,
		},
		{
			name:      "8 spawn points",
			count:     8,
			mapWidth:  2000.0,
			mapHeight: 1500.0,
			seed:      67890,
		},
		{
			name:      "deterministic generation",
			count:     5,
			mapWidth:  800.0,
			mapHeight: 600.0,
			seed:      99999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, _ := NewFFAMatch("test", 25, 10*time.Minute, tt.seed)
			match.GenerateSpawnPoints(tt.count, tt.mapWidth, tt.mapHeight)

			if len(match.SpawnPoints) != tt.count {
				t.Errorf("Got %d spawn points, want %d", len(match.SpawnPoints), tt.count)
			}

			// Verify all spawn points are within bounds
			for i, sp := range match.SpawnPoints {
				if sp.X < 0 || sp.X > tt.mapWidth {
					t.Errorf("Spawn point %d X=%f out of bounds [0, %f]", i, sp.X, tt.mapWidth)
				}
				if sp.Y < 0 || sp.Y > tt.mapHeight {
					t.Errorf("Spawn point %d Y=%f out of bounds [0, %f]", i, sp.Y, tt.mapHeight)
				}
			}
		})
	}
}

func TestFFAMatch_GenerateSpawnPoints_Deterministic(t *testing.T) {
	seed := uint64(54321)
	match1, _ := NewFFAMatch("test1", 25, 10*time.Minute, seed)
	match2, _ := NewFFAMatch("test2", 25, 10*time.Minute, seed)

	match1.GenerateSpawnPoints(6, 1000.0, 1000.0)
	match2.GenerateSpawnPoints(6, 1000.0, 1000.0)

	if len(match1.SpawnPoints) != len(match2.SpawnPoints) {
		t.Fatal("Spawn point counts differ")
	}

	for i := range match1.SpawnPoints {
		if match1.SpawnPoints[i] != match2.SpawnPoints[i] {
			t.Errorf("Spawn point %d differs: %+v vs %+v", i, match1.SpawnPoints[i], match2.SpawnPoints[i])
		}
	}
}

func TestFFAMatch_GetRandomSpawnPoint(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)

	t.Run("no spawn points", func(t *testing.T) {
		_, err := match.GetRandomSpawnPoint(1)
		if err == nil {
			t.Error("Expected error when no spawn points available")
		}
	})

	match.GenerateSpawnPoints(4, 1000.0, 1000.0)

	t.Run("returns valid spawn point", func(t *testing.T) {
		spawn, err := match.GetRandomSpawnPoint(1)
		if err != nil {
			t.Fatalf("GetRandomSpawnPoint() error = %v", err)
		}
		if spawn.X < 0 || spawn.Y < 0 {
			t.Errorf("Invalid spawn point: %+v", spawn)
		}
	})
}

func TestFFAMatch_StartMatch(t *testing.T) {
	t.Run("successful start", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
		match.AddPlayer(1)
		match.AddPlayer(2)
		match.GenerateSpawnPoints(4, 1000.0, 1000.0)

		err := match.StartMatch()
		if err != nil {
			t.Fatalf("StartMatch() error = %v", err)
		}

		if !match.Started {
			t.Error("Match should be started")
		}
		if match.StartTime.IsZero() {
			t.Error("StartTime should be set")
		}

		// Verify players are spawned
		for _, player := range match.Players {
			if player.Dead {
				t.Error("Player should not be dead at start")
			}
			if player.Health != player.MaxHealth {
				t.Error("Player should have full health at start")
			}
		}
	})

	t.Run("insufficient players", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
		match.AddPlayer(1)

		err := match.StartMatch()
		if err == nil {
			t.Error("Expected error with insufficient players")
		}
	})

	t.Run("already started", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
		match.AddPlayer(1)
		match.AddPlayer(2)
		match.StartMatch()

		err := match.StartMatch()
		if err == nil {
			t.Error("Expected error when already started")
		}
	})
}

func TestFFAMatch_OnPlayerKill(t *testing.T) {
	match, _ := NewFFAMatch("test", 3, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.AddPlayer(3)
	match.GenerateSpawnPoints(4, 1000.0, 1000.0)
	match.StartMatch()

	t.Run("normal kill", func(t *testing.T) {
		err := match.OnPlayerKill(1, 2)
		if err != nil {
			t.Fatalf("OnPlayerKill() error = %v", err)
		}

		frags, _, _ := match.GetPlayerStats(1)
		if frags != 1 {
			t.Errorf("Killer should have 1 frag, got %d", frags)
		}

		_, vDeaths, _ := match.GetPlayerStats(2)
		if vDeaths != 1 {
			t.Errorf("Victim should have 1 death, got %d", vDeaths)
		}

		victim := match.Players[2]
		if !victim.Dead {
			t.Error("Victim should be marked as dead")
		}
		if victim.RespawnTime.IsZero() {
			t.Error("Victim respawn time should be set")
		}
	})

	t.Run("win by frag limit", func(t *testing.T) {
		// Player 1 needs 2 more kills to reach frag limit of 3
		match.OnPlayerKill(1, 3)
		match.OnPlayerKill(1, 2)

		if !match.IsFinished() {
			t.Error("Match should be finished after reaching frag limit")
		}
		if match.GetWinner() != 1 {
			t.Errorf("Winner should be player 1, got %d", match.GetWinner())
		}
	})
}

func TestFFAMatch_OnPlayerKill_Errors(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.StartMatch()

	t.Run("killer not in match", func(t *testing.T) {
		err := match.OnPlayerKill(999, 2)
		if err == nil {
			t.Error("Expected error for non-existent killer")
		}
	})

	t.Run("victim not in match", func(t *testing.T) {
		err := match.OnPlayerKill(1, 999)
		if err == nil {
			t.Error("Expected error for non-existent victim")
		}
	})

	t.Run("match already finished", func(t *testing.T) {
		match.Finished = true
		err := match.OnPlayerKill(1, 2)
		if err == nil {
			t.Error("Expected error when match is finished")
		}
	})
}

func TestFFAMatch_OnPlayerSuicide(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.StartMatch()

	err := match.OnPlayerSuicide(1)
	if err != nil {
		t.Fatalf("OnPlayerSuicide() error = %v", err)
	}

	frags, deaths, _ := match.GetPlayerStats(1)
	if frags != -1 {
		t.Errorf("Frags should be -1 after suicide, got %d", frags)
	}
	if deaths != 1 {
		t.Errorf("Deaths should be 1 after suicide, got %d", deaths)
	}

	player := match.Players[1]
	if !player.Dead {
		t.Error("Player should be marked as dead after suicide")
	}
}

func TestFFAMatch_OnPlayerSuicide_Errors(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.StartMatch()

	t.Run("player not in match", func(t *testing.T) {
		err := match.OnPlayerSuicide(999)
		if err == nil {
			t.Error("Expected error for non-existent player")
		}
	})

	t.Run("match already finished", func(t *testing.T) {
		match.Finished = true
		err := match.OnPlayerSuicide(1)
		if err == nil {
			t.Error("Expected error when match is finished")
		}
	})
}

func TestFFAMatch_ProcessRespawns(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.GenerateSpawnPoints(4, 1000.0, 1000.0)
	match.StartMatch()

	// Kill a player
	match.OnPlayerKill(1, 2)

	// Process immediately - should not respawn yet
	respawned := match.ProcessRespawns()
	if len(respawned) != 0 {
		t.Error("Should not respawn before delay expires")
	}

	// Manually set respawn time to past
	player := match.Players[2]
	player.mu.Lock()
	player.RespawnTime = time.Now().Add(-1 * time.Second)
	player.mu.Unlock()

	respawned = match.ProcessRespawns()
	if len(respawned) != 1 || respawned[0] != 2 {
		t.Errorf("Expected player 2 to respawn, got %v", respawned)
	}

	// Verify player is alive
	player.mu.RLock()
	isDead := player.Dead
	player.mu.RUnlock()

	if isDead {
		t.Error("Player should be alive after respawn")
	}
}

func TestFFAMatch_ProcessRespawns_FinishedMatch(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.StartMatch()
	match.Finished = true

	respawned := match.ProcessRespawns()
	if respawned != nil {
		t.Error("Should not process respawns in finished match")
	}
}

func TestFFAMatch_RespawnPlayer(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.GenerateSpawnPoints(4, 1000.0, 1000.0)
	match.StartMatch()

	// Kill a player
	match.OnPlayerKill(1, 2)

	err := match.RespawnPlayer(2)
	if err != nil {
		t.Fatalf("RespawnPlayer() error = %v", err)
	}

	player := match.Players[2]
	player.mu.RLock()
	defer player.mu.RUnlock()

	if player.Dead {
		t.Error("Player should be alive after respawn")
	}
	if player.Health != player.MaxHealth {
		t.Error("Player should have full health after respawn")
	}
	if !player.RespawnTime.IsZero() {
		t.Error("RespawnTime should be cleared")
	}
}

func TestFFAMatch_RespawnPlayer_Errors(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.StartMatch()

	t.Run("player not in match", func(t *testing.T) {
		err := match.RespawnPlayer(999)
		if err == nil {
			t.Error("Expected error for non-existent player")
		}
	})

	t.Run("no spawn points", func(t *testing.T) {
		match2, _ := NewFFAMatch("test2", 25, 10*time.Minute, 12345)
		match2.AddPlayer(1)
		match2.AddPlayer(2)
		match2.StartMatch()

		// Clear spawn points
		match2.SpawnPoints = nil

		err := match2.RespawnPlayer(1)
		if err == nil {
			t.Error("Expected error when no spawn points available")
		}
	})
}

func TestFFAMatch_CheckTimeLimit(t *testing.T) {
	t.Run("time limit not reached", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 1*time.Hour, 12345)
		match.AddPlayer(1)
		match.AddPlayer(2)
		match.StartMatch()

		if match.CheckTimeLimit() {
			t.Error("Time limit should not be reached")
		}
		if match.IsFinished() {
			t.Error("Match should not be finished")
		}
	})

	t.Run("time limit reached", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 1*time.Millisecond, 12345)
		match.AddPlayer(1)
		match.AddPlayer(2)
		match.AddPlayer(3)
		match.StartMatch()

		// Give player 2 some frags to be the leader
		match.OnPlayerKill(2, 1)
		match.OnPlayerKill(2, 3)

		time.Sleep(2 * time.Millisecond)

		if !match.CheckTimeLimit() {
			t.Error("Time limit should be reached")
		}
		if !match.IsFinished() {
			t.Error("Match should be finished")
		}
		if match.GetWinner() != 2 {
			t.Errorf("Winner should be player 2, got %d", match.GetWinner())
		}
	})

	t.Run("not started", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 1*time.Millisecond, 12345)
		if match.CheckTimeLimit() {
			t.Error("Should not check time limit for non-started match")
		}
	})

	t.Run("already finished", func(t *testing.T) {
		match, _ := NewFFAMatch("test", 25, 1*time.Millisecond, 12345)
		match.AddPlayer(1)
		match.AddPlayer(2)
		match.StartMatch()
		match.Finished = true

		if match.CheckTimeLimit() {
			t.Error("Should not check time limit for finished match")
		}
	})
}

func TestFFAMatch_GetPlayerStats(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.StartMatch()

	match.OnPlayerKill(1, 2)
	match.OnPlayerKill(1, 2)

	frags, deaths, err := match.GetPlayerStats(1)
	if err != nil {
		t.Fatalf("GetPlayerStats() error = %v", err)
	}
	if frags != 2 {
		t.Errorf("Frags = %d, want 2", frags)
	}
	if deaths != 0 {
		t.Errorf("Deaths = %d, want 0", deaths)
	}

	_ = deaths // Use the variable to avoid unused error

	vFrags, vDeaths, err := match.GetPlayerStats(2)
	if err != nil {
		t.Fatalf("GetPlayerStats() error = %v", err)
	}
	if vFrags != 0 {
		t.Errorf("Victim frags = %d, want 0", vFrags)
	}
	if vDeaths != 2 {
		t.Errorf("Victim deaths = %d, want 2", vDeaths)
	}
}

func TestFFAMatch_GetPlayerStats_NotInMatch(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	_, _, err := match.GetPlayerStats(999)
	if err == nil {
		t.Error("Expected error for non-existent player")
	}
}

func TestFFAMatch_GetLeaderboard(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)
	match.AddPlayer(1)
	match.AddPlayer(2)
	match.AddPlayer(3)
	match.StartMatch()

	// Create varied scores
	match.OnPlayerKill(1, 2) // Player 1: 1 frag
	match.OnPlayerKill(3, 2) // Player 3: 1 frag
	match.OnPlayerKill(3, 1) // Player 3: 2 frags
	match.OnPlayerKill(3, 2) // Player 3: 3 frags
	match.OnPlayerSuicide(2) // Player 2: -1 frags

	leaderboard := match.GetLeaderboard()

	if len(leaderboard) != 3 {
		t.Fatalf("Leaderboard length = %d, want 3", len(leaderboard))
	}

	// Verify sorted by frags descending
	if leaderboard[0].PlayerID != 3 || leaderboard[0].Frags != 3 {
		t.Errorf("First place should be player 3 with 3 frags, got player %d with %d frags",
			leaderboard[0].PlayerID, leaderboard[0].Frags)
	}
	if leaderboard[1].PlayerID != 1 || leaderboard[1].Frags != 1 {
		t.Errorf("Second place should be player 1 with 1 frag, got player %d with %d frags",
			leaderboard[1].PlayerID, leaderboard[1].Frags)
	}
	if leaderboard[2].PlayerID != 2 || leaderboard[2].Frags != -1 {
		t.Errorf("Third place should be player 2 with -1 frags, got player %d with %d frags",
			leaderboard[2].PlayerID, leaderboard[2].Frags)
	}
}

func TestFFAMatch_IsFinished(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)

	if match.IsFinished() {
		t.Error("New match should not be finished")
	}

	match.Finished = true
	if !match.IsFinished() {
		t.Error("Match should be finished")
	}
}

func TestFFAMatch_GetWinner(t *testing.T) {
	match, _ := NewFFAMatch("test", 25, 10*time.Minute, 12345)

	if match.GetWinner() != 0 {
		t.Error("Unfinished match should have no winner (0)")
	}

	match.Finished = true
	match.WinnerID = 42

	if match.GetWinner() != 42 {
		t.Errorf("Winner = %d, want 42", match.GetWinner())
	}
}

func TestFFAMatch_ConcurrentAccess(t *testing.T) {
	match, _ := NewFFAMatch("test", 100, 10*time.Minute, 12345)

	// Add players
	for i := uint64(1); i <= 4; i++ {
		match.AddPlayer(i)
	}
	match.GenerateSpawnPoints(8, 1000.0, 1000.0)
	match.StartMatch()

	// Simulate concurrent kills and respawns
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				match.OnPlayerKill(1, 2)
				match.OnPlayerKill(3, 4)
				match.GetLeaderboard()
				match.ProcessRespawns()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify match state is consistent
	leaderboard := match.GetLeaderboard()
	if len(leaderboard) != 4 {
		t.Errorf("Leaderboard should have 4 players, got %d", len(leaderboard))
	}
}

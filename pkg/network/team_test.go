package network

import (
	"testing"
	"time"
)

func TestNewTeamMatch(t *testing.T) {
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
			matchID:   "tm1",
			fragLimit: 0,
			timeLimit: 0,
			seed:      12345,
			wantFrags: DefaultTeamFragLimit,
			wantTime:  DefaultTimeLimit,
		},
		{
			name:      "custom limits",
			matchID:   "tm2",
			fragLimit: 100,
			timeLimit: 15 * time.Minute,
			seed:      67890,
			wantFrags: 100,
			wantTime:  15 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewTeamMatch(tt.matchID, tt.fragLimit, tt.timeLimit, tt.seed)
			if err != nil {
				t.Fatalf("NewTeamMatch() error = %v", err)
			}
			if m.MatchID != tt.matchID {
				t.Errorf("MatchID = %v, want %v", m.MatchID, tt.matchID)
			}
			if m.FragLimit != tt.wantFrags {
				t.Errorf("FragLimit = %v, want %v", m.FragLimit, tt.wantFrags)
			}
			if m.TimeLimit != tt.wantTime {
				t.Errorf("TimeLimit = %v, want %v", m.TimeLimit, tt.wantTime)
			}
			if m.Seed != tt.seed {
				t.Errorf("Seed = %v, want %v", m.Seed, tt.seed)
			}
			if m.WinnerTeam != -1 {
				t.Errorf("WinnerTeam = %v, want -1", m.WinnerTeam)
			}
			if len(m.Teams) != 2 {
				t.Errorf("Teams count = %v, want 2", len(m.Teams))
			}
		})
	}
}

func TestTeamMatch_AddPlayer(t *testing.T) {
	tests := []struct {
		name     string
		playerID uint64
		team     int
		wantErr  bool
	}{
		{
			name:     "add to red team",
			playerID: 1,
			team:     TeamRed,
			wantErr:  false,
		},
		{
			name:     "add to blue team",
			playerID: 2,
			team:     TeamBlue,
			wantErr:  false,
		},
		{
			name:     "invalid team",
			playerID: 3,
			team:     99,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
			err := m.AddPlayer(tt.playerID, tt.team)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddPlayer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if _, exists := m.Players[tt.playerID]; !exists {
					t.Errorf("player %d not added to match", tt.playerID)
				}
			}
		})
	}
}

func TestTeamMatch_AddPlayerErrors(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)

	// Add player successfully
	if err := m.AddPlayer(1, TeamRed); err != nil {
		t.Fatalf("AddPlayer() error = %v", err)
	}

	// Try to add same player again
	if err := m.AddPlayer(1, TeamBlue); err == nil {
		t.Error("AddPlayer() should fail for duplicate player")
	}

	// Start match
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 1000)
	m.StartMatch()

	// Try to add player after match started
	if err := m.AddPlayer(3, TeamRed); err == nil {
		t.Error("AddPlayer() should fail after match started")
	}
}

func TestTeamMatch_AddPlayerMaxPlayers(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)

	// Add max players
	for i := uint64(0); i < MaxTeamPlayers; i++ {
		team := int(i % 2)
		if err := m.AddPlayer(i, team); err != nil {
			t.Fatalf("AddPlayer(%d) error = %v", i, err)
		}
	}

	// Try to add one more
	if err := m.AddPlayer(999, TeamRed); err == nil {
		t.Error("AddPlayer() should fail when match is full")
	}
}

func TestTeamMatch_RemovePlayer(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)

	err := m.RemovePlayer(1)
	if err != nil {
		t.Fatalf("RemovePlayer() error = %v", err)
	}

	player := m.Players[1]
	if player.Active {
		t.Error("player should be inactive after removal")
	}

	// Try to remove non-existent player
	if err := m.RemovePlayer(999); err == nil {
		t.Error("RemovePlayer() should fail for non-existent player")
	}
}

func TestTeamMatch_GenerateSpawnPoints(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)

	m.GenerateSpawnPoints(4, 1000.0, 500.0)

	if len(m.SpawnPoints[TeamRed]) != 4 {
		t.Errorf("red team spawn points = %d, want 4", len(m.SpawnPoints[TeamRed]))
	}
	if len(m.SpawnPoints[TeamBlue]) != 4 {
		t.Errorf("blue team spawn points = %d, want 4", len(m.SpawnPoints[TeamBlue]))
	}

	// Verify red team spawns on left side (x < 300)
	for _, spawn := range m.SpawnPoints[TeamRed] {
		if spawn.X >= 300 {
			t.Errorf("red spawn X = %f, should be < 300", spawn.X)
		}
		if spawn.Y < 0 || spawn.Y > 500 {
			t.Errorf("red spawn Y = %f, should be in [0, 500]", spawn.Y)
		}
	}

	// Verify blue team spawns on right side (x > 700)
	for _, spawn := range m.SpawnPoints[TeamBlue] {
		if spawn.X <= 700 {
			t.Errorf("blue spawn X = %f, should be > 700", spawn.X)
		}
		if spawn.Y < 0 || spawn.Y > 500 {
			t.Errorf("blue spawn Y = %f, should be in [0, 500]", spawn.Y)
		}
	}
}

func TestTeamMatch_GetRandomSpawnPoint(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)

	// No spawn points yet
	if _, err := m.GetRandomSpawnPoint(1); err == nil {
		t.Error("GetRandomSpawnPoint() should fail with no spawn points")
	}

	m.GenerateSpawnPoints(4, 1000, 500)

	// Get spawn for red player
	spawn, err := m.GetRandomSpawnPoint(1)
	if err != nil {
		t.Fatalf("GetRandomSpawnPoint(1) error = %v", err)
	}
	if spawn.X >= 300 {
		t.Errorf("red player spawn X = %f, should be < 300", spawn.X)
	}

	// Get spawn for blue player
	spawn, err = m.GetRandomSpawnPoint(2)
	if err != nil {
		t.Fatalf("GetRandomSpawnPoint(2) error = %v", err)
	}
	if spawn.X <= 700 {
		t.Errorf("blue player spawn X = %f, should be > 700", spawn.X)
	}

	// Non-existent player
	if _, err := m.GetRandomSpawnPoint(999); err == nil {
		t.Error("GetRandomSpawnPoint() should fail for non-existent player")
	}
}

func TestTeamMatch_StartMatch(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)

	// Try to start with insufficient players
	if err := m.StartMatch(); err == nil {
		t.Error("StartMatch() should fail with insufficient players")
	}

	// Add players
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)

	// Start successfully
	if err := m.StartMatch(); err != nil {
		t.Fatalf("StartMatch() error = %v", err)
	}

	if !m.Started {
		t.Error("match should be started")
	}

	// Verify players spawned
	for _, player := range m.Players {
		if player.PosX == 0 && player.PosY == 0 {
			t.Error("player should be spawned at non-zero position")
		}
		if player.Dead {
			t.Error("player should not be dead at match start")
		}
		if player.Health != player.MaxHealth {
			t.Errorf("player health = %f, want %f", player.Health, player.MaxHealth)
		}
	}

	// Try to start again
	if err := m.StartMatch(); err == nil {
		t.Error("StartMatch() should fail when already started")
	}
}

func TestTeamMatch_OnPlayerKill(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 3, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Red kills blue
	if err := m.OnPlayerKill(1, 2); err != nil {
		t.Fatalf("OnPlayerKill() error = %v", err)
	}

	// Check killer stats
	team, frags, deaths, _ := m.GetPlayerStats(1)
	if team != TeamRed {
		t.Errorf("killer team = %d, want %d", team, TeamRed)
	}
	if frags != 1 {
		t.Errorf("killer frags = %d, want 1", frags)
	}
	if deaths != 0 {
		t.Errorf("killer deaths = %d, want 0", deaths)
	}

	// Check victim stats
	team, frags, deaths, _ = m.GetPlayerStats(2)
	if team != TeamBlue {
		t.Errorf("victim team = %d, want %d", team, TeamBlue)
	}
	if frags != 0 {
		t.Errorf("victim frags = %d, want 0", frags)
	}
	if deaths != 1 {
		t.Errorf("victim deaths = %d, want 1", deaths)
	}

	// Check team scores
	redFrags, redDeaths, _ := m.GetTeamScore(TeamRed)
	if redFrags != 1 {
		t.Errorf("red team frags = %d, want 1", redFrags)
	}
	if redDeaths != 0 {
		t.Errorf("red team deaths = %d, want 0", redDeaths)
	}

	blueFrags, blueDeaths, _ := m.GetTeamScore(TeamBlue)
	if blueFrags != 0 {
		t.Errorf("blue team frags = %d, want 0", blueFrags)
	}
	if blueDeaths != 1 {
		t.Errorf("blue team deaths = %d, want 1", blueDeaths)
	}

	// Check victim is dead
	victim := m.Players[2]
	victim.mu.RLock()
	isDead := victim.Dead
	victim.mu.RUnlock()
	if !isDead {
		t.Error("victim should be dead")
	}
}

func TestTeamMatch_OnPlayerKillWinCondition(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 2, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// First kill
	m.OnPlayerKill(1, 2)
	if m.IsFinished() {
		t.Error("match should not be finished after 1 kill")
	}

	// Second kill - reaches frag limit
	m.OnPlayerKill(1, 2)
	if !m.IsFinished() {
		t.Error("match should be finished after reaching frag limit")
	}
	if m.GetWinner() != TeamRed {
		t.Errorf("winner = %d, want %d", m.GetWinner(), TeamRed)
	}
}

func TestTeamMatch_OnPlayerSuicide(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Player suicides
	if err := m.OnPlayerSuicide(1); err != nil {
		t.Fatalf("OnPlayerSuicide() error = %v", err)
	}

	// Check player stats
	_, frags, deaths, _ := m.GetPlayerStats(1)
	if frags != -1 {
		t.Errorf("frags = %d, want -1 (suicide penalty)", frags)
	}
	if deaths != 1 {
		t.Errorf("deaths = %d, want 1", deaths)
	}

	// Check team deaths increased
	_, teamDeaths, _ := m.GetTeamScore(TeamRed)
	if teamDeaths != 1 {
		t.Errorf("team deaths = %d, want 1", teamDeaths)
	}

	// Check player is dead
	player := m.Players[1]
	player.mu.RLock()
	isDead := player.Dead
	player.mu.RUnlock()
	if !isDead {
		t.Error("player should be dead after suicide")
	}
}

func TestTeamMatch_ProcessRespawns(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Kill player
	m.OnPlayerKill(1, 2)

	// Immediately check - should not respawn yet
	respawned := m.ProcessRespawns()
	if len(respawned) != 0 {
		t.Error("player should not respawn before delay")
	}

	// Set respawn time to past
	player := m.Players[2]
	player.mu.Lock()
	player.RespawnTime = time.Now().Add(-1 * time.Second)
	player.mu.Unlock()

	// Process respawns
	respawned = m.ProcessRespawns()
	if len(respawned) != 1 {
		t.Errorf("respawned count = %d, want 1", len(respawned))
	}
	if respawned[0] != 2 {
		t.Errorf("respawned player = %d, want 2", respawned[0])
	}

	// Check player is alive
	player.mu.RLock()
	isDead := player.Dead
	health := player.Health
	player.mu.RUnlock()
	if isDead {
		t.Error("player should be alive after respawn")
	}
	if health != player.MaxHealth {
		t.Errorf("health = %f, want %f", health, player.MaxHealth)
	}
}

func TestTeamMatch_RespawnPlayer(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Mark player as dead
	player := m.Players[1]
	player.mu.Lock()
	player.Dead = true
	player.Health = 0
	player.mu.Unlock()

	// Respawn
	if err := m.RespawnPlayer(1); err != nil {
		t.Fatalf("RespawnPlayer() error = %v", err)
	}

	// Check player state
	player.mu.RLock()
	isDead := player.Dead
	health := player.Health
	posX := player.PosX
	posY := player.PosY
	player.mu.RUnlock()

	if isDead {
		t.Error("player should be alive after respawn")
	}
	if health != player.MaxHealth {
		t.Errorf("health = %f, want %f", health, player.MaxHealth)
	}
	if posX == 0 && posY == 0 {
		t.Error("player should be spawned at non-zero position")
	}
	// Verify red team spawn location
	if posX >= 300 {
		t.Errorf("red player spawn X = %f, should be < 300", posX)
	}

	// Try to respawn non-existent player
	if err := m.RespawnPlayer(999); err == nil {
		t.Error("RespawnPlayer() should fail for non-existent player")
	}
}

func TestTeamMatch_CheckTimeLimit(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 100*time.Millisecond, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Give red team some kills
	m.OnPlayerKill(1, 2)

	// Check before time limit
	if m.CheckTimeLimit() {
		t.Error("time limit should not be reached yet")
	}

	// Wait for time limit
	time.Sleep(150 * time.Millisecond)

	// Check after time limit
	if !m.CheckTimeLimit() {
		t.Error("time limit should be reached")
	}

	if !m.IsFinished() {
		t.Error("match should be finished")
	}

	if m.GetWinner() != TeamRed {
		t.Errorf("winner = %d, want %d (red team had more frags)", m.GetWinner(), TeamRed)
	}
}

func TestTeamMatch_GetPlayerStats(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)

	team, frags, deaths, err := m.GetPlayerStats(1)
	if err != nil {
		t.Fatalf("GetPlayerStats() error = %v", err)
	}
	if team != TeamRed {
		t.Errorf("team = %d, want %d", team, TeamRed)
	}
	if frags != 0 {
		t.Errorf("frags = %d, want 0", frags)
	}
	if deaths != 0 {
		t.Errorf("deaths = %d, want 0", deaths)
	}

	// Non-existent player
	if _, _, _, err := m.GetPlayerStats(999); err == nil {
		t.Error("GetPlayerStats() should fail for non-existent player")
	}
}

func TestTeamMatch_GetTeamScore(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)

	frags, deaths, err := m.GetTeamScore(TeamRed)
	if err != nil {
		t.Fatalf("GetTeamScore() error = %v", err)
	}
	if frags != 0 {
		t.Errorf("frags = %d, want 0", frags)
	}
	if deaths != 0 {
		t.Errorf("deaths = %d, want 0", deaths)
	}

	// Invalid team
	if _, _, err := m.GetTeamScore(99); err == nil {
		t.Error("GetTeamScore() should fail for invalid team")
	}
}

func TestTeamMatch_GetLeaderboard(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamRed)
	m.AddPlayer(3, TeamBlue)
	m.AddPlayer(4, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Simulate some kills
	m.OnPlayerKill(1, 3) // Red player 1 kills Blue player 3
	m.OnPlayerKill(1, 4) // Red player 1 kills Blue player 4
	m.OnPlayerKill(2, 3) // Red player 2 kills Blue player 3
	m.OnPlayerKill(4, 1) // Blue player 4 kills Red player 1

	leaderboard := m.GetLeaderboard()
	if len(leaderboard) != 4 {
		t.Fatalf("leaderboard length = %d, want 4", len(leaderboard))
	}

	// Verify sorting: team first, then frags
	// Expected order: Red team (sorted by frags), then Blue team (sorted by frags)
	expectedOrder := []struct {
		playerID uint64
		team     int
		frags    int
	}{
		{1, TeamRed, 2},
		{2, TeamRed, 1},
		{4, TeamBlue, 1},
		{3, TeamBlue, 0},
	}

	for i, expected := range expectedOrder {
		if leaderboard[i].PlayerID != expected.playerID {
			t.Errorf("leaderboard[%d].PlayerID = %d, want %d", i, leaderboard[i].PlayerID, expected.playerID)
		}
		if leaderboard[i].Team != expected.team {
			t.Errorf("leaderboard[%d].Team = %d, want %d", i, leaderboard[i].Team, expected.team)
		}
		if leaderboard[i].Frags != expected.frags {
			t.Errorf("leaderboard[%d].Frags = %d, want %d", i, leaderboard[i].Frags, expected.frags)
		}
	}
}

func TestTeamMatch_FinishedAfterKillErrors(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 1, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Finish match
	m.OnPlayerKill(1, 2)

	// Try actions after match finished
	if err := m.OnPlayerKill(1, 2); err == nil {
		t.Error("OnPlayerKill() should fail after match finished")
	}
	if err := m.OnPlayerSuicide(1); err == nil {
		t.Error("OnPlayerSuicide() should fail after match finished")
	}

	// ProcessRespawns should return nil
	respawned := m.ProcessRespawns()
	if respawned != nil {
		t.Error("ProcessRespawns() should return nil after match finished")
	}
}

func TestTeamMatch_InvalidPlayerKill(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Non-existent killer
	if err := m.OnPlayerKill(999, 2); err == nil {
		t.Error("OnPlayerKill() should fail for non-existent killer")
	}

	// Non-existent victim
	if err := m.OnPlayerKill(1, 999); err == nil {
		t.Error("OnPlayerKill() should fail for non-existent victim")
	}

	// Non-existent player suicide
	if err := m.OnPlayerSuicide(999); err == nil {
		t.Error("OnPlayerSuicide() should fail for non-existent player")
	}
}

func TestTeamMatch_Concurrency(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 100, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamRed)
	m.AddPlayer(3, TeamBlue)
	m.AddPlayer(4, TeamBlue)
	m.GenerateSpawnPoints(4, 1000, 500)
	m.StartMatch()

	// Concurrent kills
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			m.OnPlayerKill(1, 3)
			m.OnPlayerKill(2, 4)
			m.OnPlayerKill(3, 1)
			m.OnPlayerKill(4, 2)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify team scores are consistent
	redFrags, redDeaths, _ := m.GetTeamScore(TeamRed)
	blueFrags, blueDeaths, _ := m.GetTeamScore(TeamBlue)

	totalFrags := redFrags + blueFrags
	totalDeaths := redDeaths + blueDeaths

	if totalFrags != totalDeaths {
		t.Errorf("total frags %d != total deaths %d", totalFrags, totalDeaths)
	}
}

func TestTeamMatch_TeamColoredIndicators(t *testing.T) {
	m, _ := NewTeamMatch("tm1", 50, 10*time.Minute, 12345)
	m.AddPlayer(1, TeamRed)
	m.AddPlayer(2, TeamBlue)

	// Verify team assignments
	p1 := m.Players[1]
	p2 := m.Players[2]

	p1.mu.RLock()
	p1Team := p1.Team
	p1.mu.RUnlock()

	p2.mu.RLock()
	p2Team := p2.Team
	p2.mu.RUnlock()

	if p1Team != TeamRed {
		t.Errorf("player 1 team = %d, want %d", p1Team, TeamRed)
	}
	if p2Team != TeamBlue {
		t.Errorf("player 2 team = %d, want %d", p2Team, TeamBlue)
	}
}

package network

import (
	"testing"
	"time"
)

// Local types for testing without importing ui package (which requires display)
type testKillFeedEntry struct {
	KillerID   uint64
	VictimID   uint64
	KillerName string
	VictimName string
	Suicide    bool
	TeamKill   bool
}

type testScoreboardEntry struct {
	PlayerID   uint64
	PlayerName string
	Team       int
	Frags      int
	Deaths     int
}

// TestDeathmatchFFAIntegration simulates a full 4-player FFA match.
func TestDeathmatchFFAIntegration(t *testing.T) {
	// Create match
	match, err := NewFFAMatch("ffa-integration-1", 10, 5*time.Minute, 12345)
	if err != nil {
		t.Fatalf("Failed to create FFA match: %v", err)
	}

	// Track kills locally (simulating kill feed without importing ui)
	killFeed := make([]testKillFeedEntry, 0)

	// Add 4 players
	playerNames := map[uint64]string{
		1: "Player1",
		2: "Player2",
		3: "Player3",
		4: "Player4",
	}

	for playerID := range playerNames {
		if err := match.AddPlayer(playerID); err != nil {
			t.Fatalf("Failed to add player %d: %v", playerID, err)
		}
	}

	// Generate spawn points
	match.GenerateSpawnPoints(8, 100.0, 100.0)

	// Start match
	if err := match.StartMatch(); err != nil {
		t.Fatalf("Failed to start match: %v", err)
	}

	// Simulate combat sequence
	combatEvents := []struct {
		killerID uint64
		victimID uint64
		suicide  bool
	}{
		{1, 2, false}, // Player1 kills Player2 (1 frag)
		{3, 4, false}, // Player3 kills Player4 (1 frag)
		{1, 3, false}, // Player1 kills Player3 (2 frags)
		{2, 1, false}, // Player2 kills Player1 (1 frag, respawned)
		{4, 2, false}, // Player4 kills Player2 (1 frag, respawned)
		{1, 0, true},  // Player1 suicides (1 frag, respawned)
		{3, 4, false}, // Player3 kills Player4 (2 frags)
		{3, 1, false}, // Player3 kills Player1 (3 frags)
		{3, 2, false}, // Player3 kills Player2 (4 frags)
		{3, 4, false}, // Player3 kills Player4 (5 frags)
		{3, 1, false}, // Player3 kills Player1 (6 frags)
		{3, 2, false}, // Player3 kills Player2 (7 frags)
		{3, 4, false}, // Player3 kills Player4 (8 frags)
		{3, 1, false}, // Player3 kills Player1 (9 frags)
		{3, 2, false}, // Player3 kills Player2 (10 frags) - should win
	}

	for i, event := range combatEvents {
		time.Sleep(10 * time.Millisecond) // Simulate time passage

		if event.suicide {
			if err := match.OnPlayerSuicide(event.killerID); err != nil {
				t.Errorf("Event %d: Failed to register suicide: %v", i, err)
			}
			killFeed = append(killFeed, testKillFeedEntry{
				KillerID:   event.killerID,
				VictimID:   event.killerID,
				KillerName: playerNames[event.killerID],
				VictimName: playerNames[event.killerID],
				Suicide:    true,
			})
		} else {
			if err := match.OnPlayerKill(event.killerID, event.victimID); err != nil {
				t.Errorf("Event %d: Failed to register kill: %v", i, err)
			}
			killFeed = append(killFeed, testKillFeedEntry{
				KillerID:   event.killerID,
				VictimID:   event.victimID,
				KillerName: playerNames[event.killerID],
				VictimName: playerNames[event.victimID],
			})
		}

		// Process respawns
		match.ProcessRespawns()
	}

	// Check match finished
	if !match.IsFinished() {
		t.Error("Match should be finished after reaching frag limit")
	}

	// Check winner
	winner := match.GetWinner()
	if winner != 3 {
		t.Errorf("Expected winner to be Player3 (ID 3), got %d", winner)
	}

	// Verify frag counts
	frags, _, err := match.GetPlayerStats(3)
	if err != nil {
		t.Fatalf("Failed to get stats for winner: %v", err)
	}

	// Player3 should have 10 frags
	if frags != 10 {
		t.Errorf("Expected winner to have 10 frags, got %d", frags)
	}

	// Update scoreboard with final stats (simulate without importing ui)
	leaderboard := match.GetLeaderboard()
	scoreboardEntries := make([]testScoreboardEntry, len(leaderboard))
	for i, stats := range leaderboard {
		scoreboardEntries[i] = testScoreboardEntry{
			PlayerID:   stats.PlayerID,
			PlayerName: playerNames[stats.PlayerID],
			Frags:      stats.Frags,
			Deaths:     stats.Deaths,
		}
	}

	// Verify scoreboard has correct number of entries
	if len(scoreboardEntries) != 4 {
		t.Errorf("Expected 4 scoreboard entries, got %d", len(scoreboardEntries))
	}

	// Verify leaderboard is sorted by frags
	for i := 0; i < len(leaderboard)-1; i++ {
		if leaderboard[i].Frags < leaderboard[i+1].Frags {
			t.Errorf("Leaderboard not sorted: position %d has %d frags, position %d has %d frags",
				i, leaderboard[i].Frags, i+1, leaderboard[i+1].Frags)
		}
	}

	// Verify kill feed has entries
	if len(killFeed) == 0 {
		t.Error("Kill feed should have entries")
	}
}

// TestDeathmatchTeamIntegration simulates a 2v2 team match.
func TestDeathmatchTeamIntegration(t *testing.T) {
	// Create match
	match, err := NewTeamMatch("team-integration-1", 15, 5*time.Minute, 54321)
	if err != nil {
		t.Fatalf("Failed to create team match: %v", err)
	}

	// Track kills locally
	killFeed := make([]testKillFeedEntry, 0)

	// Add players to teams
	playerNames := map[uint64]string{
		1: "RedPlayer1",
		2: "RedPlayer2",
		3: "BluePlayer1",
		4: "BluePlayer2",
	}

	playerTeams := map[uint64]int{
		1: TeamRed,
		2: TeamRed,
		3: TeamBlue,
		4: TeamBlue,
	}

	for playerID, team := range playerTeams {
		if err := match.AddPlayer(playerID, team); err != nil {
			t.Fatalf("Failed to add player %d to team %d: %v", playerID, team, err)
		}
	}

	// Generate spawn points
	match.GenerateSpawnPoints(4, 100.0, 100.0)

	// Start match
	if err := match.StartMatch(); err != nil {
		t.Fatalf("Failed to start match: %v", err)
	}

	// Simulate combat sequence favoring Red team
	combatEvents := []struct {
		killerID uint64
		victimID uint64
		suicide  bool
		teamKill bool
	}{
		{1, 3, false, false}, // Red kills Blue (Red: 1)
		{2, 4, false, false}, // Red kills Blue (Red: 2)
		{1, 3, false, false}, // Red kills Blue (Red: 3)
		{3, 1, false, false}, // Blue kills Red (Blue: 1)
		{2, 3, false, false}, // Red kills Blue (Red: 4)
		{1, 4, false, false}, // Red kills Blue (Red: 5)
		{4, 2, false, false}, // Blue kills Red (Blue: 2)
		{1, 3, false, false}, // Red kills Blue (Red: 6)
		{2, 4, false, false}, // Red kills Blue (Red: 7)
		{1, 3, false, false}, // Red kills Blue (Red: 8)
		{2, 4, false, false}, // Red kills Blue (Red: 9)
		{1, 3, false, false}, // Red kills Blue (Red: 10)
		{2, 4, false, false}, // Red kills Blue (Red: 11)
		{1, 3, false, false}, // Red kills Blue (Red: 12)
		{2, 4, false, false}, // Red kills Blue (Red: 13)
		{1, 3, false, false}, // Red kills Blue (Red: 14)
		{2, 4, false, false}, // Red kills Blue (Red: 15) - Red wins
	}

	for i, event := range combatEvents {
		time.Sleep(10 * time.Millisecond)

		if event.suicide {
			if err := match.OnPlayerSuicide(event.killerID); err != nil {
				t.Errorf("Event %d: Failed to register suicide: %v", i, err)
			}
			killFeed = append(killFeed, testKillFeedEntry{
				KillerID:   event.killerID,
				VictimID:   event.killerID,
				KillerName: playerNames[event.killerID],
				VictimName: playerNames[event.killerID],
				Suicide:    true,
			})
		} else {
			if err := match.OnPlayerKill(event.killerID, event.victimID); err != nil {
				t.Errorf("Event %d: Failed to register kill: %v", i, err)
			}

			killerTeam := playerTeams[event.killerID]
			victimTeam := playerTeams[event.victimID]
			teamKill := killerTeam == victimTeam

			killFeed = append(killFeed, testKillFeedEntry{
				KillerID:   event.killerID,
				VictimID:   event.victimID,
				KillerName: playerNames[event.killerID],
				VictimName: playerNames[event.victimID],
				TeamKill:   teamKill,
			})
		}

		// Process respawns
		match.ProcessRespawns()
	}

	// Check match finished
	if !match.IsFinished() {
		t.Error("Match should be finished after reaching team frag limit")
	}

	// Check winner
	winner := match.GetWinner()
	if winner != TeamRed {
		t.Errorf("Expected Red team to win, got team %d", winner)
	}

	// Verify team scores
	redFrags, redDeaths, err := match.GetTeamScore(TeamRed)
	if err != nil {
		t.Fatalf("Failed to get Red team score: %v", err)
	}

	blueFrags, _, err := match.GetTeamScore(TeamBlue)
	if err != nil {
		t.Fatalf("Failed to get Blue team score: %v", err)
	}

	if redFrags != 15 {
		t.Errorf("Expected Red team to have 15 frags, got %d", redFrags)
	}

	if blueFrags >= redFrags {
		t.Errorf("Blue team should have fewer frags than Red team (Blue: %d, Red: %d)", blueFrags, redFrags)
	}

	if redDeaths != blueFrags {
		t.Errorf("Red deaths (%d) should equal Blue frags (%d)", redDeaths, blueFrags)
	}

	// Update scoreboard with final stats (simulate without importing ui)
	leaderboard := match.GetLeaderboard()
	scoreboardEntries := make([]testScoreboardEntry, len(leaderboard))
	for i, stats := range leaderboard {
		scoreboardEntries[i] = testScoreboardEntry{
			PlayerID:   stats.PlayerID,
			PlayerName: playerNames[stats.PlayerID],
			Team:       stats.Team,
			Frags:      stats.Frags,
			Deaths:     stats.Deaths,
		}
	}

	// Verify scoreboard entries
	if len(scoreboardEntries) != 4 {
		t.Errorf("Expected 4 scoreboard entries, got %d", len(scoreboardEntries))
	}

	// Verify team grouping in leaderboard (Red team first, then Blue)
	if scoreboardEntries[0].Team != TeamRed {
		t.Error("First leaderboard entry should be Red team")
	}
	if scoreboardEntries[1].Team != TeamRed {
		t.Error("Second leaderboard entry should be Red team")
	}
}

// TestDeathmatchTimeLimitIntegration tests time limit win condition.
func TestDeathmatchTimeLimitIntegration(t *testing.T) {
	// Create match with very short time limit
	match, err := NewFFAMatch("ffa-time-1", 100, 100*time.Millisecond, 99999)
	if err != nil {
		t.Fatalf("Failed to create FFA match: %v", err)
	}

	// Track kills locally
	killFeed := make([]testKillFeedEntry, 0)

	playerNames := map[uint64]string{
		1: "Player1",
		2: "Player2",
	}

	// Add players
	for playerID := range playerNames {
		if err := match.AddPlayer(playerID); err != nil {
			t.Fatalf("Failed to add player %d: %v", playerID, err)
		}
	}

	match.GenerateSpawnPoints(4, 100.0, 100.0)

	// Start match
	if err := match.StartMatch(); err != nil {
		t.Fatalf("Failed to start match: %v", err)
	}

	// Register a few kills
	match.OnPlayerKill(1, 2)
	killFeed = append(killFeed, testKillFeedEntry{
		KillerID:   1,
		VictimID:   2,
		KillerName: playerNames[1],
		VictimName: playerNames[2],
	})
	match.OnPlayerKill(1, 2)
	killFeed = append(killFeed, testKillFeedEntry{
		KillerID:   1,
		VictimID:   2,
		KillerName: playerNames[1],
		VictimName: playerNames[2],
	})

	// Wait for time limit
	time.Sleep(150 * time.Millisecond)

	// Check time limit
	if !match.CheckTimeLimit() {
		t.Error("Time limit should have been reached")
	}

	// Match should be finished
	if !match.IsFinished() {
		t.Error("Match should be finished after time limit")
	}

	// Winner should be player with most frags
	winner := match.GetWinner()
	if winner != 1 {
		t.Errorf("Expected Player1 to win on time, got player %d", winner)
	}

	// Update scoreboard (simulate without importing ui)
	leaderboard := match.GetLeaderboard()
	scoreboardEntries := make([]testScoreboardEntry, len(leaderboard))
	for i, stats := range leaderboard {
		scoreboardEntries[i] = testScoreboardEntry{
			PlayerID:   stats.PlayerID,
			PlayerName: playerNames[stats.PlayerID],
			Frags:      stats.Frags,
			Deaths:     stats.Deaths,
		}
	}
}

// TestDeathmatchRespawnCycle tests continuous respawn in combat.
func TestDeathmatchRespawnCycle(t *testing.T) {
	match, err := NewFFAMatch("ffa-respawn-1", 50, 5*time.Minute, 11111)
	if err != nil {
		t.Fatalf("Failed to create FFA match: %v", err)
	}

	// Track kills locally
	killFeed := make([]testKillFeedEntry, 0)

	playerNames := map[uint64]string{
		1: "Player1",
		2: "Player2",
	}

	for playerID := range playerNames {
		if err := match.AddPlayer(playerID); err != nil {
			t.Fatalf("Failed to add player %d: %v", playerID, err)
		}
	}

	match.GenerateSpawnPoints(4, 100.0, 100.0)

	if err := match.StartMatch(); err != nil {
		t.Fatalf("Failed to start match: %v", err)
	}

	// Kill player 2 multiple times
	for i := 0; i < 5; i++ {
		if err := match.OnPlayerKill(1, 2); err != nil {
			t.Errorf("Kill %d failed: %v", i, err)
		}
		killFeed = append(killFeed, testKillFeedEntry{
			KillerID:   1,
			VictimID:   2,
			KillerName: playerNames[1],
			VictimName: playerNames[2],
		})

		// Wait for respawn delay
		time.Sleep(RespawnDelay + 10*time.Millisecond)

		// Process respawns
		respawned := match.ProcessRespawns()
		if len(respawned) != 1 || respawned[0] != 2 {
			t.Errorf("Iteration %d: Expected player 2 to respawn, got %v", i, respawned)
		}
	}

	// Verify stats
	frags, _, err := match.GetPlayerStats(1)
	if err != nil {
		t.Fatalf("Failed to get Player1 stats: %v", err)
	}
	if frags != 5 {
		t.Errorf("Expected Player1 to have 5 frags, got %d", frags)
	}

	frags2, deaths2, err := match.GetPlayerStats(2)
	if err != nil {
		t.Fatalf("Failed to get Player2 stats: %v", err)
	}
	if deaths2 != 5 {
		t.Errorf("Expected Player2 to have 5 deaths, got %d", deaths2)
	}
	if frags2 != 0 {
		t.Errorf("Expected Player2 to have 0 frags, got %d", frags2)
	}
}

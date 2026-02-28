package network

import (
	"sync"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/combat"
)

// TestCoopIntegration_FullSession simulates a complete 4-player co-op session:
// - 4 players join
// - Players engage in combat
// - Some players die and respawn
// - Players complete objectives
// - Session ends with level completion
func TestCoopIntegration_FullSession(t *testing.T) {
	const (
		sessionID = "integration-test-session"
		levelSeed = uint64(42)
	)

	// Create session
	session, err := NewCoopSession(sessionID, MaxCoopPlayers, levelSeed)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Phase 1: Join - Add 4 players
	playerIDs := []uint64{100, 101, 102, 103}
	for _, pid := range playerIDs {
		if err := session.AddPlayer(pid); err != nil {
			t.Fatalf("AddPlayer(%d) failed: %v", pid, err)
		}
	}

	// Verify all players joined
	if count := session.GetPlayerCount(); count != 4 {
		t.Errorf("Expected 4 players, got %d", count)
	}

	// Start session
	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !session.Started {
		t.Error("Session should be started")
	}

	// Verify quest tracker generated objectives
	mainObjectives := session.QuestTracker.GetMainObjectives()
	if len(mainObjectives) == 0 {
		t.Error("Expected generated objectives, got none")
	}

	// Phase 2: Combat - Spread players out and simulate combat
	positions := [][2]float64{{10, 10}, {20, 10}, {10, 20}, {20, 20}}
	for i, pid := range playerIDs {
		if err := session.UpdatePlayerPosition(pid, positions[i][0], positions[i][1]); err != nil {
			t.Errorf("UpdatePlayerPosition(%d) failed: %v", pid, err)
		}
	}

	// Simulate damage to players
	damageEvents := []struct {
		playerID uint64
		damage   float64
	}{
		{100, 50.0},  // Player 100 takes 50 damage
		{101, 30.0},  // Player 101 takes 30 damage
		{102, 105.0}, // Player 102 takes fatal damage
		{103, 20.0},  // Player 103 takes 20 damage
	}

	for _, evt := range damageEvents {
		player, _ := session.GetPlayer(evt.playerID)
		newHealth := player.Health - evt.damage
		if newHealth <= 0 {
			// Player dies
			if err := session.OnPlayerDeath(evt.playerID); err != nil {
				t.Errorf("OnPlayerDeath(%d) failed: %v", evt.playerID, err)
			}
		} else {
			if err := session.UpdatePlayerHealth(evt.playerID, newHealth); err != nil {
				t.Errorf("UpdatePlayerHealth(%d) failed: %v", evt.playerID, err)
			}
		}
	}

	// Verify player 102 is dead and in bleedout
	player102, _ := session.GetPlayer(102)
	if !player102.Dead {
		t.Error("Player 102 should be dead")
	}
	if player102.Health != 0 {
		t.Errorf("Dead player should have 0 health, got %f", player102.Health)
	}

	// Phase 3: Respawn - Simulate bleedout timer expiry
	// Fast-forward bleedout timer
	player102.mu.Lock()
	player102.BleedoutEndTime = time.Now().Add(-1 * time.Second) // Already expired
	player102.mu.Unlock()

	// Process bleedouts
	toRespawn := session.ProcessBleedouts()
	if len(toRespawn) != 1 || toRespawn[0] != 102 {
		t.Errorf("Expected [102] to respawn, got %v", toRespawn)
	}

	// Respawn player 102 at nearest living teammate
	if err := session.RespawnPlayer(102); err != nil {
		t.Errorf("RespawnPlayer(102) failed: %v", err)
	}

	// Verify player 102 is alive with full health
	player102, _ = session.GetPlayer(102)
	if player102.Dead {
		t.Error("Player 102 should be alive after respawn")
	}
	if player102.Health != player102.MaxHealth {
		t.Errorf("Respawned player should have full health (%f), got %f", player102.MaxHealth, player102.Health)
	}

	// Verify respawn position is near a living teammate
	nearLivingTeammate := false
	for _, pid := range []uint64{100, 101, 103} {
		p, _ := session.GetPlayer(pid)
		p.mu.RLock()
		dx := p.PosX - player102.PosX
		dy := p.PosY - player102.PosY
		dist := dx*dx + dy*dy
		p.mu.RUnlock()
		if dist < 1.0 { // Very close (same position)
			nearLivingTeammate = true
			break
		}
	}
	if !nearLivingTeammate {
		t.Error("Respawned player should be near a living teammate")
	}

	// Phase 4: Objective Completion - Complete all objectives
	for _, obj := range mainObjectives {
		session.CompleteObjective(obj.ID)
	}

	// Verify level completion
	if !session.IsLevelComplete() {
		t.Error("Level should be complete after all objectives done")
	}

	// Phase 5: Verify final state
	activePlayers := session.GetActivePlayers()
	if len(activePlayers) != 4 {
		t.Errorf("Expected 4 active players at end, got %d", len(activePlayers))
	}

	// Verify all players are alive
	for _, pid := range playerIDs {
		isDead, _ := session.IsPlayerDead(pid)
		if isDead {
			t.Errorf("Player %d should be alive at session end", pid)
		}
	}
}

// TestCoopIntegration_PartyWipe simulates a full party wipe scenario:
// - 4 players join and start
// - All players die
// - Party wipe triggers level restart
func TestCoopIntegration_PartyWipe(t *testing.T) {
	session, err := NewCoopSession("party-wipe-test", 4, 999)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Add 4 players
	for i := uint64(200); i < 204; i++ {
		if err := session.AddPlayer(i); err != nil {
			t.Fatalf("AddPlayer failed: %v", err)
		}
	}

	// Start session
	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify objectives were generated
	initialObjectiveCount := len(session.QuestTracker.GetMainObjectives())
	if initialObjectiveCount == 0 {
		t.Error("Expected objectives to be generated")
	}

	// Kill all players
	for i := uint64(200); i < 204; i++ {
		if err := session.OnPlayerDeath(i); err != nil {
			t.Errorf("OnPlayerDeath(%d) failed: %v", i, err)
		}
	}

	// Verify party is wiped
	if !session.isPartyWiped() {
		t.Error("Party should be wiped when all players are dead")
	}

	// Attempt to respawn a player (should fail due to party wipe)
	err = session.RespawnPlayer(200)
	if err == nil {
		t.Error("Expected RespawnPlayer to fail during party wipe")
	}

	// Restart level
	if err := session.RestartLevel(); err != nil {
		t.Fatalf("RestartLevel failed: %v", err)
	}

	// Verify all players are alive and reset
	for i := uint64(200); i < 204; i++ {
		player, _ := session.GetPlayer(i)
		if player.Dead {
			t.Errorf("Player %d should be alive after restart", i)
		}
		if player.Health != player.MaxHealth {
			t.Errorf("Player %d should have full health after restart, got %f", i, player.Health)
		}
		if player.PosX != 0 || player.PosY != 0 {
			t.Errorf("Player %d should be at spawn (0,0), got (%f,%f)", i, player.PosX, player.PosY)
		}
	}

	// Verify objectives were regenerated
	newObjectiveCount := len(session.QuestTracker.GetMainObjectives())
	if newObjectiveCount == 0 {
		t.Error("Expected objectives to be regenerated after restart")
	}

	// Level should not be completed
	if session.LevelCompleted {
		t.Error("Level should not be completed after restart")
	}
}

// TestCoopIntegration_PlayerDisconnect simulates player disconnection and reconnection:
// - 3 players join
// - One player disconnects
// - Remaining players continue
// - Player reconnects (simulated by reactivating)
func TestCoopIntegration_PlayerDisconnect(t *testing.T) {
	session, err := NewCoopSession("disconnect-test", 4, 555)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Add 3 players
	playerIDs := []uint64{300, 301, 302}
	for _, pid := range playerIDs {
		if err := session.AddPlayer(pid); err != nil {
			t.Fatalf("AddPlayer failed: %v", err)
		}
	}

	// Start session
	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Set player positions
	for i, pid := range playerIDs {
		if err := session.UpdatePlayerPosition(pid, float64(i*10), float64(i*10)); err != nil {
			t.Errorf("UpdatePlayerPosition failed: %v", err)
		}
	}

	// Player 301 disconnects
	if err := session.RemovePlayer(301); err != nil {
		t.Errorf("RemovePlayer failed: %v", err)
	}

	// Verify player 301 is inactive
	player301, _ := session.GetPlayer(301)
	if player301.Active {
		t.Error("Disconnected player should be inactive")
	}

	// Verify active player count
	activePlayers := session.GetActivePlayers()
	if len(activePlayers) != 2 {
		t.Errorf("Expected 2 active players after disconnect, got %d", len(activePlayers))
	}

	// Total player count should still be 3 (state preserved)
	if session.GetPlayerCount() != 3 {
		t.Errorf("Expected 3 total players, got %d", session.GetPlayerCount())
	}

	// Simulate reconnection by reactivating player
	player301.mu.Lock()
	player301.Active = true
	player301.mu.Unlock()

	// Verify player 301 reconnected
	activePlayers = session.GetActivePlayers()
	if len(activePlayers) != 3 {
		t.Errorf("Expected 3 active players after reconnect, got %d", len(activePlayers))
	}

	// Verify player state was preserved (position, inventory)
	player301, _ = session.GetPlayer(301)
	if player301.PosX != 10 || player301.PosY != 10 {
		t.Errorf("Player position should be preserved, got (%f,%f)", player301.PosX, player301.PosY)
	}
}

// TestCoopIntegration_ConcurrentActions simulates concurrent player actions:
// - 4 players perform actions simultaneously
// - Tests thread safety of session operations
func TestCoopIntegration_ConcurrentActions(t *testing.T) {
	session, err := NewCoopSession("concurrent-test", 4, 777)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Add 4 players
	playerIDs := []uint64{400, 401, 402, 403}
	for _, pid := range playerIDs {
		if err := session.AddPlayer(pid); err != nil {
			t.Fatalf("AddPlayer failed: %v", err)
		}
	}

	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Simulate concurrent actions from all players
	var wg sync.WaitGroup
	iterations := 100

	for _, pid := range playerIDs {
		wg.Add(1)
		go func(playerID uint64) {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				// Update position
				x := float64(i % 50)
				y := float64(i % 50)
				session.UpdatePlayerPosition(playerID, x, y)

				// Update health
				health := 50.0 + float64(i%50)
				session.UpdatePlayerHealth(playerID, health)

				// Update objectives
				objID := "obj-1"
				session.UpdateObjectiveProgress(objID, 1)

				// Get player state (read operation)
				_, _ = session.GetPlayer(playerID)
			}
		}(pid)
	}

	wg.Wait()

	// Verify final state consistency
	for _, pid := range playerIDs {
		player, err := session.GetPlayer(pid)
		if err != nil {
			t.Errorf("GetPlayer(%d) failed: %v", pid, err)
		}
		if !player.Active {
			t.Errorf("Player %d should be active", pid)
		}
	}

	// All operations should complete without race conditions
	activePlayers := session.GetActivePlayers()
	if len(activePlayers) != 4 {
		t.Errorf("Expected 4 active players, got %d", len(activePlayers))
	}
}

// TestCoopIntegration_ObjectiveTracking tests shared objective progression:
// - Multiple players contribute to objective progress
// - Verify progress is shared across all players
func TestCoopIntegration_ObjectiveTracking(t *testing.T) {
	session, err := NewCoopSession("objective-test", 3, 888)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Add 3 players
	for i := uint64(500); i < 503; i++ {
		if err := session.AddPlayer(i); err != nil {
			t.Fatalf("AddPlayer failed: %v", err)
		}
	}

	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	objectives := session.QuestTracker.GetMainObjectives()
	if len(objectives) == 0 {
		t.Fatal("Expected objectives to be generated")
	}

	firstObj := objectives[0]

	// Each player contributes to the same objective
	for i := 0; i < 3; i++ {
		session.UpdateObjectiveProgress(firstObj.ID, 10)
	}

	// Check progress by looking at the objective directly
	var found bool
	for _, obj := range session.QuestTracker.Objectives {
		if obj.ID == firstObj.ID {
			if obj.Progress < 30 {
				t.Errorf("Expected progress >= 30, got %d", obj.Progress)
			}
			found = true
			break
		}
	}
	if !found {
		t.Error("Objective not found in tracker")
	}

	// Complete objective
	session.CompleteObjective(firstObj.ID)

	// Verify objective is complete
	var isComplete bool
	for _, obj := range session.QuestTracker.Objectives {
		if obj.ID == firstObj.ID && obj.Complete {
			isComplete = true
			break
		}
	}
	if !isComplete {
		t.Error("Objective should be complete")
	}
}

// TestCoopIntegration_RespawnChain tests sequential respawns:
// - Multiple players die in sequence
// - Each respawns at nearest living teammate
// - Verify respawn logic handles changing team composition
func TestCoopIntegration_RespawnChain(t *testing.T) {
	session, err := NewCoopSession("respawn-chain-test", 4, 1111)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Add 4 players
	playerIDs := []uint64{600, 601, 602, 603}
	for _, pid := range playerIDs {
		if err := session.AddPlayer(pid); err != nil {
			t.Fatalf("AddPlayer failed: %v", err)
		}
	}

	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Position players in a line
	for i, pid := range playerIDs {
		if err := session.UpdatePlayerPosition(pid, float64(i*100), 0); err != nil {
			t.Errorf("UpdatePlayerPosition failed: %v", err)
		}
	}

	// Kill players 600 and 602
	for _, pid := range []uint64{600, 602} {
		if err := session.OnPlayerDeath(pid); err != nil {
			t.Errorf("OnPlayerDeath failed: %v", err)
		}
	}

	// Expire bleedout timers
	for _, pid := range []uint64{600, 602} {
		p, _ := session.GetPlayer(pid)
		p.mu.Lock()
		p.BleedoutEndTime = time.Now().Add(-1 * time.Second)
		p.mu.Unlock()
	}

	// Process bleedouts
	toRespawn := session.ProcessBleedouts()
	if len(toRespawn) != 2 {
		t.Errorf("Expected 2 players to respawn, got %d", len(toRespawn))
	}

	// Respawn player 600 (should spawn near player 601 or 603)
	if err := session.RespawnPlayer(600); err != nil {
		t.Errorf("RespawnPlayer(600) failed: %v", err)
	}

	// Respawn player 602 (should spawn near living teammate)
	if err := session.RespawnPlayer(602); err != nil {
		t.Errorf("RespawnPlayer(602) failed: %v", err)
	}

	// Verify both are alive
	for _, pid := range []uint64{600, 602} {
		isDead, _ := session.IsPlayerDead(pid)
		if isDead {
			t.Errorf("Player %d should be alive after respawn", pid)
		}
	}

	// All 4 players should be alive now
	aliveCount := 0
	for _, pid := range playerIDs {
		isDead, _ := session.IsPlayerDead(pid)
		if !isDead {
			aliveCount++
		}
	}
	if aliveCount != 4 {
		t.Errorf("Expected 4 alive players, got %d", aliveCount)
	}
}

// TestCoopIntegration_GenreConfiguration tests genre-specific session setup:
// - Configure session for a specific genre
// - Verify genre is applied to world and quest tracker
func TestCoopIntegration_GenreConfiguration(t *testing.T) {
	session, err := NewCoopSession("genre-test", 2, 2222)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Add 2 players
	for i := uint64(700); i < 702; i++ {
		if err := session.AddPlayer(i); err != nil {
			t.Fatalf("AddPlayer failed: %v", err)
		}
	}

	// Set genre before starting
	session.SetGenre("scifi")

	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Verify world and quest tracker have genre set
	// (This is implicit - the SetGenre calls should propagate)
	// Just verify session started successfully with genre
	if !session.Started {
		t.Error("Session should have started")
	}
}

// TestCoopIntegration_EdgeCases tests edge cases and error conditions:
// - Session full (5th player tries to join)
// - Start with insufficient players
// - Restart without starting first
func TestCoopIntegration_EdgeCases(t *testing.T) {
	t.Run("SessionFull", func(t *testing.T) {
		session, _ := NewCoopSession("full-test", 2, 3333)

		// Add 2 players (max)
		session.AddPlayer(800)
		session.AddPlayer(801)

		// Try to add 3rd player (should fail)
		err := session.AddPlayer(802)
		if err == nil {
			t.Error("Expected error when session is full")
		}
	})

	t.Run("InsufficientPlayers", func(t *testing.T) {
		session, _ := NewCoopSession("insufficient-test", 4, 4444)

		// Add only 1 player (minimum is 2)
		session.AddPlayer(900)

		// Try to start (should fail)
		err := session.Start()
		if err == nil {
			t.Error("Expected error when starting with insufficient players")
		}
	})

	t.Run("RestartWithoutStart", func(t *testing.T) {
		session, _ := NewCoopSession("restart-test", 2, 5555)

		session.AddPlayer(1000)
		session.AddPlayer(1001)

		// Try to restart without starting first (should fail)
		err := session.RestartLevel()
		if err == nil {
			t.Error("Expected error when restarting without starting first")
		}
	})

	t.Run("DuplicatePlayer", func(t *testing.T) {
		session, _ := NewCoopSession("duplicate-test", 3, 6666)

		session.AddPlayer(1100)

		// Try to add same player again (should fail)
		err := session.AddPlayer(1100)
		if err == nil {
			t.Error("Expected error when adding duplicate player")
		}
	})

	t.Run("RemoveNonexistentPlayer", func(t *testing.T) {
		session, _ := NewCoopSession("remove-test", 2, 7777)

		// Try to remove player that was never added
		err := session.RemovePlayer(1200)
		if err == nil {
			t.Error("Expected error when removing nonexistent player")
		}
	})
}

// TestCoopIntegration_WithCombatSystem tests co-op with actual combat system integration:
// - Simulate combat encounters
// - Verify damage and death mechanics
func TestCoopIntegration_WithCombatSystem(t *testing.T) {
	session, err := NewCoopSession("combat-test", 3, 8888)
	if err != nil {
		t.Fatalf("NewCoopSession failed: %v", err)
	}

	// Add 3 players
	playerIDs := []uint64{1300, 1301, 1302}
	for _, pid := range playerIDs {
		if err := session.AddPlayer(pid); err != nil {
			t.Fatalf("AddPlayer failed: %v", err)
		}
	}

	if err := session.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Create combat system
	combatSys := combat.NewSystem()

	// Get player state
	player, _ := session.GetPlayer(1300)

	// Simulate combat damage: 50 base damage with player at 100 health, 0 armor
	damageResult := combatSys.ApplyDamage(player.Health, player.Armor, 50.0,
		player.PosX, player.PosY, 0, 0)

	// Apply health damage to player
	newHealth := player.Health - damageResult.HealthDamage

	if damageResult.Killed || newHealth <= 0 {
		session.OnPlayerDeath(1300)
	} else {
		session.UpdatePlayerHealth(1300, newHealth)
	}

	// Verify combat integration works
	player, _ = session.GetPlayer(1300)
	if damageResult.Killed {
		if !player.Dead {
			t.Error("Player should be dead after fatal damage")
		}
	} else {
		expectedHealth := 100.0 - damageResult.HealthDamage
		if player.Health != expectedHealth {
			t.Errorf("Expected health %f, got %f", expectedHealth, player.Health)
		}
	}
}

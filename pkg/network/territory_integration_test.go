package network

import (
	"testing"
	"time"
)

// TestTerritoryControlIntegration tests the complete territory control flow
func TestTerritoryControlIntegration(t *testing.T) {
	t.Run("full territory control match", func(t *testing.T) {
		// Create match with low score limit for fast test
		match, err := NewTerritoryMatch("territory_integration", 10, 1*time.Minute, 99999)
		if err != nil {
			t.Fatalf("NewTerritoryMatch() error = %v", err)
		}

		// Add 4 players (2v2)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamRed)
		match.AddPlayer(3, TeamBlue)
		match.AddPlayer(4, TeamBlue)

		// Add 3 control points
		match.AddControlPoint("cp_a", 0, 0)
		match.AddControlPoint("cp_b", 20, 0)
		match.AddControlPoint("cp_c", -20, 0)

		// Start match
		err = match.Start()
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		// Position players at control points
		// Red team captures cp_a and cp_c
		match.Players[1].mu.Lock()
		match.Players[1].PosX, match.Players[1].PosY = 0, 0
		match.Players[1].mu.Unlock()

		match.Players[2].mu.Lock()
		match.Players[2].PosX, match.Players[2].PosY = -20, 0
		match.Players[2].mu.Unlock()

		// Blue team captures cp_b
		match.Players[3].mu.Lock()
		match.Players[3].PosX, match.Players[3].PosY = 20, 0
		match.Players[3].mu.Unlock()

		match.Players[4].mu.Lock()
		match.Players[4].PosX, match.Players[4].PosY = 100, 100 // Far away
		match.Players[4].mu.Unlock()

		// Process capture for 20 ticks to fully capture points
		for i := 0; i < 20; i++ {
			match.ProcessCapture()
		}

		// Verify ownership
		if owner := match.ControlPoints["cp_a"].GetOwner(); owner != OwnershipRed {
			t.Errorf("cp_a owner = %v, want %v", owner, OwnershipRed)
		}
		if owner := match.ControlPoints["cp_b"].GetOwner(); owner != OwnershipBlue {
			t.Errorf("cp_b owner = %v, want %v", owner, OwnershipBlue)
		}
		if owner := match.ControlPoints["cp_c"].GetOwner(); owner != OwnershipRed {
			t.Errorf("cp_c owner = %v, want %v", owner, OwnershipRed)
		}

		// Force score tick to elapse
		match.mu.Lock()
		match.LastScoreTick = time.Now().Add(-2 * time.Second)
		match.mu.Unlock()

		// Process scoring
		match.ProcessScoring()

		// Red should have 2 points (2 CPs), Blue should have 1 point (1 CP)
		redScore, _ := match.GetTeamScore(TeamRed)
		blueScore, _ := match.GetTeamScore(TeamBlue)

		if redScore != 2 {
			t.Errorf("Red score = %d, want 2", redScore)
		}
		if blueScore != 1 {
			t.Errorf("Blue score = %d, want 1", blueScore)
		}

		// Simulate blue team losing cp_b to red
		match.Players[3].mu.Lock()
		match.Players[3].PosX, match.Players[3].PosY = 100, 100 // Move away
		match.Players[3].mu.Unlock()

		match.Players[2].mu.Lock()
		match.Players[2].PosX, match.Players[2].PosY = 20, 0 // Red captures cp_b
		match.Players[2].mu.Unlock()

		// Process capture for neutralization and re-capture
		for i := 0; i < 40; i++ {
			match.ProcessCapture()
		}

		// cp_b should now be red
		if owner := match.ControlPoints["cp_b"].GetOwner(); owner != OwnershipRed {
			t.Errorf("cp_b owner after recapture = %v, want %v", owner, OwnershipRed)
		}

		// Continue scoring until red team wins
		for i := 0; i < 10; i++ {
			match.mu.Lock()
			match.LastScoreTick = time.Now().Add(-2 * time.Second)
			match.mu.Unlock()

			match.ProcessScoring()

			if match.CheckWinCondition() {
				break
			}

			time.Sleep(10 * time.Millisecond)
		}

		// Verify red team won
		if !match.Finished {
			t.Error("Match should be finished")
		}
		if match.WinnerTeam != TeamRed {
			t.Errorf("WinnerTeam = %d, want %d", match.WinnerTeam, TeamRed)
		}

		finalRed, _ := match.GetTeamScore(TeamRed)
		if finalRed < 10 {
			t.Errorf("Final red score = %d, want >= 10", finalRed)
		}
	})

	t.Run("contested control point", func(t *testing.T) {
		match, _ := NewTerritoryMatch("contested", 100, 1*time.Minute, 88888)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp_contested", 0, 0)
		match.Start()

		// Both players at same control point
		match.Players[1].mu.Lock()
		match.Players[1].PosX, match.Players[1].PosY = 0, 0
		match.Players[1].mu.Unlock()

		match.Players[2].mu.Lock()
		match.Players[2].PosX, match.Players[2].PosY = 0, 0
		match.Players[2].mu.Unlock()

		// Process capture for many ticks
		for i := 0; i < 50; i++ {
			match.ProcessCapture()
		}

		// Control point should remain neutral
		cp := match.ControlPoints["cp_contested"]
		if owner := cp.GetOwner(); owner != OwnershipNeutral {
			t.Errorf("Contested CP owner = %v, want %v", owner, OwnershipNeutral)
		}

		// Progress should be near 0
		progress := cp.GetCaptureProgress()
		if progress < -0.1 || progress > 0.1 {
			t.Errorf("Contested CP progress = %f, want near 0.0", progress)
		}
	})

	t.Run("player advantage in capture", func(t *testing.T) {
		match, _ := NewTerritoryMatch("advantage", 100, 1*time.Minute, 77777)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamRed)
		match.AddPlayer(3, TeamBlue)
		match.AddControlPoint("cp_advantage", 0, 0)
		match.Start()

		// 2 red vs 1 blue
		match.Players[1].mu.Lock()
		match.Players[1].PosX, match.Players[1].PosY = 0, 0
		match.Players[1].mu.Unlock()

		match.Players[2].mu.Lock()
		match.Players[2].PosX, match.Players[2].PosY = 0, 0
		match.Players[2].mu.Unlock()

		match.Players[3].mu.Lock()
		match.Players[3].PosX, match.Players[3].PosY = 0, 0
		match.Players[3].mu.Unlock()

		// Process one tick
		match.ProcessCapture()

		// Progress should move toward red (negative)
		cp := match.ControlPoints["cp_advantage"]
		progress := cp.GetCaptureProgress()

		// With 2 red and 1 blue, net is 1 red player = -0.05 per tick
		if progress >= 0 {
			t.Errorf("Progress with red advantage = %f, want < 0", progress)
		}

		expectedProgress := -CaptureRatePerTick * 1.0 // 1 net red player
		if progress < expectedProgress-0.01 || progress > expectedProgress+0.01 {
			t.Errorf("Progress = %f, want ~%f", progress, expectedProgress)
		}
	})

	t.Run("time limit with equal scores", func(t *testing.T) {
		match, _ := NewTerritoryMatch("time_tie", 100, 1*time.Millisecond, 66666)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		// Set equal scores
		match.Teams[TeamRed].mu.Lock()
		match.Teams[TeamRed].Frags = 5
		match.Teams[TeamRed].mu.Unlock()

		match.Teams[TeamBlue].mu.Lock()
		match.Teams[TeamBlue].Frags = 5
		match.Teams[TeamBlue].mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		won := match.CheckWinCondition()

		if !won {
			t.Error("CheckWinCondition() should return true after time limit")
		}
		if match.WinnerTeam != -1 {
			t.Errorf("WinnerTeam = %d, want -1 (tie)", match.WinnerTeam)
		}
	})
}

// TestTerritoryControlScoring specifically tests the scoring system
func TestTerritoryControlScoring(t *testing.T) {
	t.Run("points earned per second per control point", func(t *testing.T) {
		match, _ := NewTerritoryMatch("scoring", 100, 10*time.Minute, 55555)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.AddControlPoint("cp2", 10, 10)
		match.AddControlPoint("cp3", -10, -10)
		match.Start()

		// Red owns all 3
		match.ControlPoints["cp1"].Owner = OwnershipRed
		match.ControlPoints["cp2"].Owner = OwnershipRed
		match.ControlPoints["cp3"].Owner = OwnershipRed

		initialRed, _ := match.GetTeamScore(TeamRed)

		// Force tick
		match.mu.Lock()
		match.LastScoreTick = time.Now().Add(-2 * time.Second)
		match.mu.Unlock()

		match.ProcessScoring()

		finalRed, _ := match.GetTeamScore(TeamRed)

		// Should earn 3 points (1 per CP)
		if finalRed != initialRed+3 {
			t.Errorf("Red score increased by %d, want 3", finalRed-initialRed)
		}
	})

	t.Run("no points for neutral control points", func(t *testing.T) {
		match, _ := NewTerritoryMatch("neutral_scoring", 100, 10*time.Minute, 44444)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		// CP remains neutral
		match.ControlPoints["cp1"].Owner = OwnershipNeutral

		initialRed, _ := match.GetTeamScore(TeamRed)
		initialBlue, _ := match.GetTeamScore(TeamBlue)

		match.mu.Lock()
		match.LastScoreTick = time.Now().Add(-2 * time.Second)
		match.mu.Unlock()

		match.ProcessScoring()

		finalRed, _ := match.GetTeamScore(TeamRed)
		finalBlue, _ := match.GetTeamScore(TeamBlue)

		if finalRed != initialRed || finalBlue != initialBlue {
			t.Error("Scores should not change for neutral control points")
		}
	})

	t.Run("score tick rate limits scoring frequency", func(t *testing.T) {
		match, _ := NewTerritoryMatch("tick_rate", 100, 10*time.Minute, 33333)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		match.ControlPoints["cp1"].Owner = OwnershipRed

		initialRed, _ := match.GetTeamScore(TeamRed)

		// Try to score immediately (should not work)
		match.ProcessScoring()

		afterFirstCall, _ := match.GetTeamScore(TeamRed)

		if afterFirstCall != initialRed {
			t.Error("Should not score before tick rate elapses")
		}

		// Wait for tick rate
		match.mu.Lock()
		match.LastScoreTick = time.Now().Add(-2 * time.Second)
		match.mu.Unlock()

		match.ProcessScoring()

		afterSecondCall, _ := match.GetTeamScore(TeamRed)

		if afterSecondCall == initialRed {
			t.Error("Should score after tick rate elapses")
		}
	})
}

package network

import (
	"testing"
	"time"
)

func TestNewControlPoint(t *testing.T) {
	tests := []struct {
		name string
		id   string
		x    float64
		y    float64
	}{
		{"center point", "cp_center", 0.0, 0.0},
		{"offset point", "cp_offset", 10.5, -5.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewControlPoint(tt.id, 1, tt.x, tt.y)

			if cp.ID != tt.id {
				t.Errorf("ID = %s, want %s", cp.ID, tt.id)
			}
			if cp.PosX != tt.x {
				t.Errorf("PosX = %f, want %f", cp.PosX, tt.x)
			}
			if cp.PosY != tt.y {
				t.Errorf("PosY = %f, want %f", cp.PosY, tt.y)
			}
			if cp.Owner != OwnershipNeutral {
				t.Errorf("Owner = %v, want %v", cp.Owner, OwnershipNeutral)
			}
			if cp.CaptureProgress != 0.0 {
				t.Errorf("CaptureProgress = %f, want 0.0", cp.CaptureProgress)
			}
		})
	}
}

func TestControlPointGetOwner(t *testing.T) {
	cp := NewControlPoint("test", 1, 0, 0)

	if owner := cp.GetOwner(); owner != OwnershipNeutral {
		t.Errorf("GetOwner() = %v, want %v", owner, OwnershipNeutral)
	}

	cp.Owner = OwnershipRed
	if owner := cp.GetOwner(); owner != OwnershipRed {
		t.Errorf("GetOwner() = %v, want %v", owner, OwnershipRed)
	}
}

func TestControlPointGetCaptureProgress(t *testing.T) {
	cp := NewControlPoint("test", 1, 0, 0)

	if progress := cp.GetCaptureProgress(); progress != 0.0 {
		t.Errorf("GetCaptureProgress() = %f, want 0.0", progress)
	}

	cp.CaptureProgress = 0.5
	if progress := cp.GetCaptureProgress(); progress != 0.5 {
		t.Errorf("GetCaptureProgress() = %f, want 0.5", progress)
	}
}

func TestControlPointUpdateCapture(t *testing.T) {
	tests := []struct {
		name            string
		initialProgress float64
		initialOwner    ControlPointOwnership
		redCount        int
		blueCount       int
		wantProgress    float64
		wantOwner       ControlPointOwnership
		wantChanged     bool
	}{
		{
			name:            "red captures neutral",
			initialProgress: 0.0,
			initialOwner:    OwnershipNeutral,
			redCount:        1,
			blueCount:       0,
			wantProgress:    -0.05,
			wantOwner:       OwnershipNeutral,
			wantChanged:     false,
		},
		{
			name:            "blue captures neutral",
			initialProgress: 0.0,
			initialOwner:    OwnershipNeutral,
			redCount:        0,
			blueCount:       1,
			wantProgress:    0.05,
			wantOwner:       OwnershipNeutral,
			wantChanged:     false,
		},
		{
			name:            "contested no change",
			initialProgress: 0.0,
			initialOwner:    OwnershipNeutral,
			redCount:        1,
			blueCount:       1,
			wantProgress:    0.0,
			wantOwner:       OwnershipNeutral,
			wantChanged:     false,
		},
		{
			name:            "red completes capture",
			initialProgress: -0.85,
			initialOwner:    OwnershipNeutral,
			redCount:        1,
			blueCount:       0,
			wantProgress:    -0.90,
			wantOwner:       OwnershipRed,
			wantChanged:     true,
		},
		{
			name:            "blue completes capture",
			initialProgress: 0.85,
			initialOwner:    OwnershipNeutral,
			redCount:        0,
			blueCount:       1,
			wantProgress:    0.90,
			wantOwner:       OwnershipBlue,
			wantChanged:     true,
		},
		{
			name:            "multiple red players",
			initialProgress: 0.0,
			initialOwner:    OwnershipNeutral,
			redCount:        3,
			blueCount:       0,
			wantProgress:    -0.15,
			wantOwner:       OwnershipNeutral,
			wantChanged:     false,
		},
		{
			name:            "progress clamped at -1.0",
			initialProgress: -0.98,
			initialOwner:    OwnershipRed,
			redCount:        2,
			blueCount:       0,
			wantProgress:    -1.0,
			wantOwner:       OwnershipRed,
			wantChanged:     false,
		},
		{
			name:            "progress clamped at 1.0",
			initialProgress: 0.98,
			initialOwner:    OwnershipBlue,
			redCount:        0,
			blueCount:       2,
			wantProgress:    1.0,
			wantOwner:       OwnershipBlue,
			wantChanged:     false,
		},
		{
			name:            "red loses point to neutral",
			initialProgress: -0.92,
			initialOwner:    OwnershipRed,
			redCount:        0,
			blueCount:       1,
			wantProgress:    -0.87,
			wantOwner:       OwnershipNeutral,
			wantChanged:     true,
		},
		{
			name:            "blue loses point to neutral",
			initialProgress: 0.92,
			initialOwner:    OwnershipBlue,
			redCount:        1,
			blueCount:       0,
			wantProgress:    0.87,
			wantOwner:       OwnershipNeutral,
			wantChanged:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := NewControlPoint("test", 1, 0, 0)
			cp.CaptureProgress = tt.initialProgress
			cp.Owner = tt.initialOwner

			changed := cp.UpdateCapture(tt.redCount, tt.blueCount)

			if changed != tt.wantChanged {
				t.Errorf("UpdateCapture() changed = %v, want %v", changed, tt.wantChanged)
			}

			// Allow small floating point error
			if progress := cp.GetCaptureProgress(); progress < tt.wantProgress-0.001 || progress > tt.wantProgress+0.001 {
				t.Errorf("CaptureProgress = %f, want %f", progress, tt.wantProgress)
			}

			if owner := cp.GetOwner(); owner != tt.wantOwner {
				t.Errorf("Owner = %v, want %v", owner, tt.wantOwner)
			}
		})
	}
}

func TestControlPointIsPlayerInRange(t *testing.T) {
	cp := NewControlPoint("test", 1, 0, 0)

	tests := []struct {
		name        string
		playerX     float64
		playerY     float64
		wantInRange bool
	}{
		{"at center", 0.0, 0.0, true},
		{"just inside", 4.9, 0.0, true},
		{"just outside", 5.1, 0.0, false},
		{"diagonal inside", 3.0, 3.0, true},   // sqrt(18) ≈ 4.24 < 5.0
		{"diagonal outside", 4.0, 4.0, false}, // sqrt(32) ≈ 5.66 > 5.0
		{"far away", 100.0, 100.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inRange := cp.IsPlayerInRange(tt.playerX, tt.playerY)
			if inRange != tt.wantInRange {
				t.Errorf("IsPlayerInRange(%f, %f) = %v, want %v",
					tt.playerX, tt.playerY, inRange, tt.wantInRange)
			}
		})
	}
}

func TestNewTerritoryMatch(t *testing.T) {
	tests := []struct {
		name       string
		matchID    string
		scoreLimit int
		timeLimit  time.Duration
		seed       uint64
	}{
		{"default limits", "match1", 100, 10 * time.Minute, 12345},
		{"custom limits", "match2", 200, 5 * time.Minute, 67890},
		{"zero uses defaults", "match3", 0, 0, 11111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := NewTerritoryMatch(tt.matchID, tt.scoreLimit, tt.timeLimit, tt.seed)
			if err != nil {
				t.Fatalf("NewTerritoryMatch() error = %v", err)
			}

			if match.MatchID != tt.matchID {
				t.Errorf("MatchID = %s, want %s", match.MatchID, tt.matchID)
			}

			if tt.scoreLimit > 0 && match.ScoreLimit != tt.scoreLimit {
				t.Errorf("ScoreLimit = %d, want %d", match.ScoreLimit, tt.scoreLimit)
			}

			if match.Seed != tt.seed {
				t.Errorf("Seed = %d, want %d", match.Seed, tt.seed)
			}

			if len(match.Teams) != 2 {
				t.Errorf("Teams count = %d, want 2", len(match.Teams))
			}

			if match.WinnerTeam != -1 {
				t.Errorf("WinnerTeam = %d, want -1", match.WinnerTeam)
			}
		})
	}
}

func TestTerritoryMatchAddControlPoint(t *testing.T) {
	match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)

	t.Run("add first control point", func(t *testing.T) {
		err := match.AddControlPoint("cp1", 10.0, 20.0)
		if err != nil {
			t.Fatalf("AddControlPoint() error = %v", err)
		}

		if len(match.ControlPoints) != 1 {
			t.Errorf("ControlPoints count = %d, want 1", len(match.ControlPoints))
		}

		cp := match.ControlPoints["cp1"]
		if cp == nil {
			t.Fatal("Control point cp1 not found")
		}
		if cp.PosX != 10.0 || cp.PosY != 20.0 {
			t.Errorf("Control point position = (%f, %f), want (10.0, 20.0)", cp.PosX, cp.PosY)
		}
	})

	t.Run("add duplicate control point fails", func(t *testing.T) {
		err := match.AddControlPoint("cp1", 0.0, 0.0)
		if err == nil {
			t.Error("AddControlPoint() expected error for duplicate, got nil")
		}
	})

	t.Run("add multiple control points", func(t *testing.T) {
		match.AddControlPoint("cp2", -10.0, -20.0)
		match.AddControlPoint("cp3", 5.0, 5.0)

		if len(match.ControlPoints) != 3 {
			t.Errorf("ControlPoints count = %d, want 3", len(match.ControlPoints))
		}
	})
}

func TestTerritoryMatchAddPlayer(t *testing.T) {
	match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)

	t.Run("add player to red team", func(t *testing.T) {
		err := match.AddPlayer(1, TeamRed)
		if err != nil {
			t.Fatalf("AddPlayer() error = %v", err)
		}

		if len(match.Players) != 1 {
			t.Errorf("Players count = %d, want 1", len(match.Players))
		}

		p := match.Players[1]
		if p.Team != TeamRed {
			t.Errorf("Player team = %d, want %d", p.Team, TeamRed)
		}
	})

	t.Run("add player to blue team", func(t *testing.T) {
		err := match.AddPlayer(2, TeamBlue)
		if err != nil {
			t.Fatalf("AddPlayer() error = %v", err)
		}

		if len(match.Players) != 2 {
			t.Errorf("Players count = %d, want 2", len(match.Players))
		}
	})

	t.Run("add duplicate player fails", func(t *testing.T) {
		err := match.AddPlayer(1, TeamBlue)
		if err == nil {
			t.Error("AddPlayer() expected error for duplicate, got nil")
		}
	})

	t.Run("add player with invalid team fails", func(t *testing.T) {
		err := match.AddPlayer(99, 5)
		if err == nil {
			t.Error("AddPlayer() expected error for invalid team, got nil")
		}
	})
}

func TestTerritoryMatchStart(t *testing.T) {
	t.Run("start with sufficient players and control points", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)

		err := match.Start()
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}

		if !match.Started {
			t.Error("Match should be started")
		}
	})

	t.Run("start without enough players fails", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
		match.AddPlayer(1, TeamRed)
		match.AddControlPoint("cp1", 0, 0)

		err := match.Start()
		if err == nil {
			t.Error("Start() expected error for insufficient players, got nil")
		}
	})

	t.Run("start without control points fails", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)

		err := match.Start()
		if err == nil {
			t.Error("Start() expected error for no control points, got nil")
		}
	})

	t.Run("start already started match fails", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		err := match.Start()
		if err == nil {
			t.Error("Start() expected error for already started match, got nil")
		}
	})
}

func TestTerritoryMatchProcessCapture(t *testing.T) {
	match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
	match.AddPlayer(1, TeamRed)
	match.AddPlayer(2, TeamBlue)
	match.AddControlPoint("cp1", 0, 0)
	match.Start()

	t.Run("capture with player in range", func(t *testing.T) {
		// Move red player to control point
		p := match.Players[1]
		p.mu.Lock()
		p.PosX = 0.0
		p.PosY = 0.0
		p.mu.Unlock()

		// Move blue player away
		p2 := match.Players[2]
		p2.mu.Lock()
		p2.PosX = 100.0
		p2.PosY = 100.0
		p2.mu.Unlock()

		match.ProcessCapture()

		cp := match.ControlPoints["cp1"]
		progress := cp.GetCaptureProgress()
		if progress >= 0 {
			t.Errorf("CaptureProgress = %f, expected negative (red capturing)", progress)
		}
	})

	t.Run("no capture when contested", func(t *testing.T) {
		cp := match.ControlPoints["cp1"]
		cp.CaptureProgress = 0.0

		// Both players at control point
		match.Players[1].mu.Lock()
		match.Players[1].PosX = 0.0
		match.Players[1].PosY = 0.0
		match.Players[1].mu.Unlock()

		match.Players[2].mu.Lock()
		match.Players[2].PosX = 0.0
		match.Players[2].PosY = 0.0
		match.Players[2].mu.Unlock()

		match.ProcessCapture()

		progress := cp.GetCaptureProgress()
		if progress != 0.0 {
			t.Errorf("CaptureProgress = %f, want 0.0 (contested)", progress)
		}
	})

	t.Run("dead players don't capture", func(t *testing.T) {
		cp := match.ControlPoints["cp1"]
		cp.CaptureProgress = 0.0

		// Red player at point but dead
		match.Players[1].mu.Lock()
		match.Players[1].PosX = 0.0
		match.Players[1].PosY = 0.0
		match.Players[1].Dead = true
		match.Players[1].mu.Unlock()

		// Blue player far away
		match.Players[2].mu.Lock()
		match.Players[2].PosX = 100.0
		match.Players[2].PosY = 100.0
		match.Players[2].mu.Unlock()

		match.ProcessCapture()

		progress := cp.GetCaptureProgress()
		if progress != 0.0 {
			t.Errorf("CaptureProgress = %f, want 0.0 (dead player)", progress)
		}
	})
}

func TestTerritoryMatchProcessScoring(t *testing.T) {
	match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
	match.AddPlayer(1, TeamRed)
	match.AddPlayer(2, TeamBlue)
	match.AddControlPoint("cp1", 0, 0)
	match.AddControlPoint("cp2", 10, 10)
	match.Start()

	t.Run("scoring before tick rate returns early", func(t *testing.T) {
		initialRed, _ := match.GetTeamScore(TeamRed)
		initialBlue, _ := match.GetTeamScore(TeamBlue)

		match.ProcessScoring()

		red, _ := match.GetTeamScore(TeamRed)
		blue, _ := match.GetTeamScore(TeamBlue)

		if red != initialRed || blue != initialBlue {
			t.Error("Scores should not change before tick rate elapsed")
		}
	})

	t.Run("scoring awards points for controlled points", func(t *testing.T) {
		// Set up controlled points
		match.ControlPoints["cp1"].Owner = OwnershipRed
		match.ControlPoints["cp2"].Owner = OwnershipBlue

		// Force tick time to pass
		match.mu.Lock()
		match.LastScoreTick = time.Now().Add(-2 * time.Second)
		match.mu.Unlock()

		initialRed, _ := match.GetTeamScore(TeamRed)
		initialBlue, _ := match.GetTeamScore(TeamBlue)

		match.ProcessScoring()

		red, _ := match.GetTeamScore(TeamRed)
		blue, _ := match.GetTeamScore(TeamBlue)

		if red != initialRed+1 {
			t.Errorf("Red score = %d, want %d (1 point for 1 CP)", red, initialRed+1)
		}
		if blue != initialBlue+1 {
			t.Errorf("Blue score = %d, want %d (1 point for 1 CP)", blue, initialBlue+1)
		}
	})

	t.Run("team with more CPs scores more", func(t *testing.T) {
		match.ControlPoints["cp1"].Owner = OwnershipRed
		match.ControlPoints["cp2"].Owner = OwnershipRed

		match.mu.Lock()
		match.LastScoreTick = time.Now().Add(-2 * time.Second)
		match.mu.Unlock()

		initialRed, _ := match.GetTeamScore(TeamRed)

		match.ProcessScoring()

		red, _ := match.GetTeamScore(TeamRed)

		if red != initialRed+2 {
			t.Errorf("Red score = %d, want %d (2 points for 2 CPs)", red, initialRed+2)
		}
	})
}

func TestTerritoryMatchCheckWinCondition(t *testing.T) {
	t.Run("red wins by score limit", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 10, 10*time.Minute, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		// Set red score to limit
		match.Teams[TeamRed].mu.Lock()
		match.Teams[TeamRed].Frags = 10
		match.Teams[TeamRed].mu.Unlock()

		won := match.CheckWinCondition()

		if !won {
			t.Error("CheckWinCondition() = false, want true")
		}
		if !match.Finished {
			t.Error("Match should be finished")
		}
		if match.WinnerTeam != TeamRed {
			t.Errorf("WinnerTeam = %d, want %d", match.WinnerTeam, TeamRed)
		}
	})

	t.Run("blue wins by score limit", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 10, 10*time.Minute, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		match.Teams[TeamBlue].mu.Lock()
		match.Teams[TeamBlue].Frags = 10
		match.Teams[TeamBlue].mu.Unlock()

		won := match.CheckWinCondition()

		if !won {
			t.Error("CheckWinCondition() = false, want true")
		}
		if match.WinnerTeam != TeamBlue {
			t.Errorf("WinnerTeam = %d, want %d", match.WinnerTeam, TeamBlue)
		}
	})

	t.Run("red wins by time limit", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 1*time.Millisecond, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		match.Teams[TeamRed].mu.Lock()
		match.Teams[TeamRed].Frags = 5
		match.Teams[TeamRed].mu.Unlock()

		match.Teams[TeamBlue].mu.Lock()
		match.Teams[TeamBlue].Frags = 3
		match.Teams[TeamBlue].mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		won := match.CheckWinCondition()

		if !won {
			t.Error("CheckWinCondition() = false, want true")
		}
		if match.WinnerTeam != TeamRed {
			t.Errorf("WinnerTeam = %d, want %d", match.WinnerTeam, TeamRed)
		}
	})

	t.Run("tie by time limit", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 1*time.Millisecond, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		match.Teams[TeamRed].mu.Lock()
		match.Teams[TeamRed].Frags = 5
		match.Teams[TeamRed].mu.Unlock()

		match.Teams[TeamBlue].mu.Lock()
		match.Teams[TeamBlue].Frags = 5
		match.Teams[TeamBlue].mu.Unlock()

		time.Sleep(10 * time.Millisecond)

		won := match.CheckWinCondition()

		if !won {
			t.Error("CheckWinCondition() = false, want true")
		}
		if match.WinnerTeam != -1 {
			t.Errorf("WinnerTeam = %d, want -1 (tie)", match.WinnerTeam)
		}
	})

	t.Run("no win condition met", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
		match.AddPlayer(1, TeamRed)
		match.AddPlayer(2, TeamBlue)
		match.AddControlPoint("cp1", 0, 0)
		match.Start()

		won := match.CheckWinCondition()

		if won {
			t.Error("CheckWinCondition() = true, want false")
		}
		if match.Finished {
			t.Error("Match should not be finished")
		}
	})
}

func TestTerritoryMatchGetTeamScore(t *testing.T) {
	match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)

	t.Run("get red team score", func(t *testing.T) {
		match.Teams[TeamRed].mu.Lock()
		match.Teams[TeamRed].Frags = 42
		match.Teams[TeamRed].mu.Unlock()

		score, err := match.GetTeamScore(TeamRed)
		if err != nil {
			t.Fatalf("GetTeamScore() error = %v", err)
		}
		if score != 42 {
			t.Errorf("GetTeamScore(TeamRed) = %d, want 42", score)
		}
	})

	t.Run("get blue team score", func(t *testing.T) {
		match.Teams[TeamBlue].mu.Lock()
		match.Teams[TeamBlue].Frags = 37
		match.Teams[TeamBlue].mu.Unlock()

		score, err := match.GetTeamScore(TeamBlue)
		if err != nil {
			t.Fatalf("GetTeamScore() error = %v", err)
		}
		if score != 37 {
			t.Errorf("GetTeamScore(TeamBlue) = %d, want 37", score)
		}
	})

	t.Run("invalid team returns error", func(t *testing.T) {
		_, err := match.GetTeamScore(999)
		if err == nil {
			t.Error("GetTeamScore(999) expected error, got nil")
		}
	})
}

// TestControlPointVisualStyle tests genre-flavored visual styles
func TestControlPointVisualStyle(t *testing.T) {
	t.Run("default visual style", func(t *testing.T) {
		cp := NewControlPoint("test", 1, 0, 0)

		style := cp.GetVisualStyle()
		if style != "generic" {
			t.Errorf("GetVisualStyle() = %s, want 'generic'", style)
		}
	})

	t.Run("set visual style", func(t *testing.T) {
		cp := NewControlPoint("test", 1, 0, 0)

		cp.SetVisualStyle("altar")

		style := cp.GetVisualStyle()
		if style != "altar" {
			t.Errorf("GetVisualStyle() = %s, want 'altar'", style)
		}
	})

	t.Run("set multiple styles", func(t *testing.T) {
		cp := NewControlPoint("test", 1, 0, 0)

		styles := []string{"altar", "terminal", "summoning-circle", "server-rack", "scrap-pile"}
		for _, style := range styles {
			cp.SetVisualStyle(style)
			got := cp.GetVisualStyle()
			if got != style {
				t.Errorf("After SetVisualStyle(%s), GetVisualStyle() = %s", style, got)
			}
		}
	})
}

// TestGenreToVisualStyle tests genre-to-visual-style mapping
func TestGenreToVisualStyle(t *testing.T) {
	tests := []struct {
		genre string
		want  string
	}{
		{"fantasy", "altar"},
		{"scifi", "terminal"},
		{"horror", "summoning-circle"},
		{"cyberpunk", "server-rack"},
		{"postapoc", "scrap-pile"},
		{"unknown", "generic"},
		{"", "generic"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			got := genreToVisualStyle(tt.genre)
			if got != tt.want {
				t.Errorf("genreToVisualStyle(%s) = %s, want %s", tt.genre, got, tt.want)
			}
		})
	}
}

// TestTerritoryMatchSetGenre tests SetGenre method
func TestTerritoryMatchSetGenre(t *testing.T) {
	t.Run("set genre updates all control points", func(t *testing.T) {
		match, _ := NewTerritoryMatch("genre_test", 100, 10*time.Minute, 123)
		match.AddControlPoint("cp1", 0, 0)
		match.AddControlPoint("cp2", 10, 10)
		match.AddControlPoint("cp3", -10, -10)

		match.SetGenre("fantasy")

		if match.Genre != "fantasy" {
			t.Errorf("Genre = %s, want 'fantasy'", match.Genre)
		}

		// All control points should have altar style
		for id, cp := range match.ControlPoints {
			style := cp.GetVisualStyle()
			if style != "altar" {
				t.Errorf("CP %s visual style = %s, want 'altar'", id, style)
			}
		}
	})

	t.Run("set different genres", func(t *testing.T) {
		genreTests := []struct {
			genre string
			style string
		}{
			{"scifi", "terminal"},
			{"horror", "summoning-circle"},
			{"cyberpunk", "server-rack"},
			{"postapoc", "scrap-pile"},
		}

		for _, tt := range genreTests {
			t.Run(tt.genre, func(t *testing.T) {
				match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)
				match.AddControlPoint("cp1", 0, 0)

				match.SetGenre(tt.genre)

				cp := match.ControlPoints["cp1"]
				style := cp.GetVisualStyle()
				if style != tt.style {
					t.Errorf("After SetGenre(%s), CP style = %s, want %s", tt.genre, style, tt.style)
				}
			})
		}
	})

	t.Run("set genre before adding control points", func(t *testing.T) {
		match, _ := NewTerritoryMatch("test", 100, 10*time.Minute, 123)

		match.SetGenre("cyberpunk")

		// Add control points after setting genre
		match.AddControlPoint("cp1", 0, 0)

		// New control point should have default style (not updated retroactively)
		cp := match.ControlPoints["cp1"]
		style := cp.GetVisualStyle()
		if style != "generic" {
			t.Errorf("CP style = %s, want 'generic' (default for new CP)", style)
		}

		// But we can update it by calling SetGenre again
		match.SetGenre("cyberpunk")
		style = cp.GetVisualStyle()
		if style != "server-rack" {
			t.Errorf("After second SetGenre, CP style = %s, want 'server-rack'", style)
		}
	})
}

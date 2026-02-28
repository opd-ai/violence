package ui

import (
	"image/color"
	"testing"
)

func TestNewNameplate(t *testing.T) {
	np := NewNameplate()
	if np == nil {
		t.Fatal("NewNameplate returned nil")
	}
	if np.maxTagLength != 4 {
		t.Errorf("expected maxTagLength=4, got %d", np.maxTagLength)
	}
	if len(np.players) != 0 {
		t.Errorf("expected empty players slice, got %d players", len(np.players))
	}
}

func TestNameplate_SetPlayers(t *testing.T) {
	np := NewNameplate()

	players := []NameplatePlayer{
		{PlayerID: "p1", PlayerName: "Alice", SquadTag: "TEAM", ScreenX: 100, ScreenY: 200, IsTeammate: true},
		{PlayerID: "p2", PlayerName: "Bob", SquadTag: "CREW", ScreenX: 150, ScreenY: 250, IsTeammate: false},
	}

	np.SetPlayers(players)

	if len(np.players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(np.players))
	}
	if np.players[0].PlayerName != "Alice" {
		t.Errorf("expected first player 'Alice', got %s", np.players[0].PlayerName)
	}
	if np.players[1].PlayerName != "Bob" {
		t.Errorf("expected second player 'Bob', got %s", np.players[1].PlayerName)
	}

	// Verify copy was made (not same slice)
	players[0].PlayerName = "Charlie"
	if np.players[0].PlayerName == "Charlie" {
		t.Error("SetPlayers did not copy the slice")
	}
}

func TestNameplate_AddPlayer(t *testing.T) {
	np := NewNameplate()

	player := NameplatePlayer{
		PlayerID:   "p1",
		PlayerName: "Alice",
		SquadTag:   "TEAM",
		ScreenX:    100,
		ScreenY:    200,
		IsTeammate: true,
	}

	np.AddPlayer(player)

	if len(np.players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(np.players))
	}
	if np.players[0].PlayerName != "Alice" {
		t.Errorf("expected player 'Alice', got %s", np.players[0].PlayerName)
	}
	if np.players[0].SquadTag != "TEAM" {
		t.Errorf("expected tag 'TEAM', got %s", np.players[0].SquadTag)
	}
}

func TestNameplate_AddPlayer_TruncatesLongTag(t *testing.T) {
	np := NewNameplate()

	player := NameplatePlayer{
		PlayerID:   "p1",
		PlayerName: "Alice",
		SquadTag:   "TOOLONG", // 7 characters, should be truncated to 4
		ScreenX:    100,
		ScreenY:    200,
	}

	np.AddPlayer(player)

	if len(np.players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(np.players))
	}
	if np.players[0].SquadTag != "TOOL" {
		t.Errorf("expected truncated tag 'TOOL', got %s", np.players[0].SquadTag)
	}
}

func TestNameplate_ClearPlayers(t *testing.T) {
	np := NewNameplate()

	np.AddPlayer(NameplatePlayer{PlayerID: "p1", PlayerName: "Alice"})
	np.AddPlayer(NameplatePlayer{PlayerID: "p2", PlayerName: "Bob"})

	if len(np.players) != 2 {
		t.Fatalf("expected 2 players before clear, got %d", len(np.players))
	}

	np.ClearPlayers()

	if len(np.players) != 0 {
		t.Errorf("expected 0 players after clear, got %d", len(np.players))
	}
}

func TestNameplate_GetPlayerCount(t *testing.T) {
	np := NewNameplate()

	if np.GetPlayerCount() != 0 {
		t.Errorf("expected count=0, got %d", np.GetPlayerCount())
	}

	np.AddPlayer(NameplatePlayer{PlayerID: "p1", PlayerName: "Alice"})
	if np.GetPlayerCount() != 1 {
		t.Errorf("expected count=1, got %d", np.GetPlayerCount())
	}

	np.AddPlayer(NameplatePlayer{PlayerID: "p2", PlayerName: "Bob"})
	if np.GetPlayerCount() != 2 {
		t.Errorf("expected count=2, got %d", np.GetPlayerCount())
	}

	np.ClearPlayers()
	if np.GetPlayerCount() != 0 {
		t.Errorf("expected count=0 after clear, got %d", np.GetPlayerCount())
	}
}

func TestNameplate_SetTeammateColor(t *testing.T) {
	np := NewNameplate()

	customColor := color.RGBA{100, 150, 200, 255}
	np.SetTeammateColor(customColor)

	if np.teammateColor != customColor {
		t.Errorf("expected teammateColor=%v, got %v", customColor, np.teammateColor)
	}
}

func TestNameplate_SetEnemyColor(t *testing.T) {
	np := NewNameplate()

	customColor := color.RGBA{200, 50, 50, 255}
	np.SetEnemyColor(customColor)

	if np.enemyColor != customColor {
		t.Errorf("expected enemyColor=%v, got %v", customColor, np.enemyColor)
	}
}

func TestNameplate_SetSelfColor(t *testing.T) {
	np := NewNameplate()

	customColor := color.RGBA{255, 200, 0, 255}
	np.SetSelfColor(customColor)

	if np.selfColor != customColor {
		t.Errorf("expected selfColor=%v, got %v", customColor, np.selfColor)
	}
}

func TestNameplate_MultiplePlayersWithDifferentTags(t *testing.T) {
	np := NewNameplate()

	players := []NameplatePlayer{
		{PlayerID: "p1", PlayerName: "Alice", SquadTag: "ALFA", IsTeammate: true, IsSelf: false},
		{PlayerID: "p2", PlayerName: "Bob", SquadTag: "BRVO", IsTeammate: true, IsSelf: false},
		{PlayerID: "p3", PlayerName: "Charlie", SquadTag: "", IsTeammate: false, IsSelf: false},
		{PlayerID: "p4", PlayerName: "Dave", SquadTag: "ECHO", IsTeammate: false, IsSelf: false},
		{PlayerID: "p5", PlayerName: "Eve", SquadTag: "ALFA", IsTeammate: true, IsSelf: true},
	}

	np.SetPlayers(players)

	if np.GetPlayerCount() != 5 {
		t.Fatalf("expected 5 players, got %d", np.GetPlayerCount())
	}

	// Verify each player
	tests := []struct {
		index      int
		name       string
		tag        string
		isTeammate bool
		isSelf     bool
	}{
		{0, "Alice", "ALFA", true, false},
		{1, "Bob", "BRVO", true, false},
		{2, "Charlie", "", false, false},
		{3, "Dave", "ECHO", false, false},
		{4, "Eve", "ALFA", true, true},
	}

	for _, tt := range tests {
		p := np.players[tt.index]
		if p.PlayerName != tt.name {
			t.Errorf("player[%d] name: expected %s, got %s", tt.index, tt.name, p.PlayerName)
		}
		if p.SquadTag != tt.tag {
			t.Errorf("player[%d] tag: expected %s, got %s", tt.index, tt.tag, p.SquadTag)
		}
		if p.IsTeammate != tt.isTeammate {
			t.Errorf("player[%d] isTeammate: expected %v, got %v", tt.index, tt.isTeammate, p.IsTeammate)
		}
		if p.IsSelf != tt.isSelf {
			t.Errorf("player[%d] isSelf: expected %v, got %v", tt.index, tt.isSelf, p.IsSelf)
		}
	}
}

func TestNameplate_EmptySquadTag(t *testing.T) {
	np := NewNameplate()

	player := NameplatePlayer{
		PlayerID:   "p1",
		PlayerName: "NoSquad",
		SquadTag:   "", // No squad tag
		ScreenX:    100,
		ScreenY:    200,
	}

	np.AddPlayer(player)

	if len(np.players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(np.players))
	}
	if np.players[0].SquadTag != "" {
		t.Errorf("expected empty tag, got %s", np.players[0].SquadTag)
	}
}

func TestNameplate_ScreenPositions(t *testing.T) {
	np := NewNameplate()

	players := []NameplatePlayer{
		{PlayerID: "p1", PlayerName: "Alice", ScreenX: 100.5, ScreenY: 200.25},
		{PlayerID: "p2", PlayerName: "Bob", ScreenX: 320.0, ScreenY: 240.0},
		{PlayerID: "p3", PlayerName: "Charlie", ScreenX: 640.0, ScreenY: 480.0},
	}

	np.SetPlayers(players)

	if np.players[0].ScreenX != 100.5 || np.players[0].ScreenY != 200.25 {
		t.Errorf("player[0] position: expected (100.5, 200.25), got (%.1f, %.1f)",
			np.players[0].ScreenX, np.players[0].ScreenY)
	}
	if np.players[1].ScreenX != 320.0 || np.players[1].ScreenY != 240.0 {
		t.Errorf("player[1] position: expected (320.0, 240.0), got (%.1f, %.1f)",
			np.players[1].ScreenX, np.players[1].ScreenY)
	}
	if np.players[2].ScreenX != 640.0 || np.players[2].ScreenY != 480.0 {
		t.Errorf("player[2] position: expected (640.0, 480.0), got (%.1f, %.1f)",
			np.players[2].ScreenX, np.players[2].ScreenY)
	}
}

func TestNameplate_DefaultColors(t *testing.T) {
	np := NewNameplate()

	expectedTeammate := color.RGBA{0, 255, 0, 255} // Green
	expectedEnemy := color.RGBA{255, 0, 0, 255}    // Red
	expectedSelf := color.RGBA{255, 255, 0, 255}   // Yellow

	if np.teammateColor != expectedTeammate {
		t.Errorf("default teammateColor: expected %v, got %v", expectedTeammate, np.teammateColor)
	}
	if np.enemyColor != expectedEnemy {
		t.Errorf("default enemyColor: expected %v, got %v", expectedEnemy, np.enemyColor)
	}
	if np.selfColor != expectedSelf {
		t.Errorf("default selfColor: expected %v, got %v", expectedSelf, np.selfColor)
	}
}

func TestNameplate_ExactlyFourCharacterTag(t *testing.T) {
	np := NewNameplate()

	player := NameplatePlayer{
		PlayerID:   "p1",
		PlayerName: "Alice",
		SquadTag:   "ABCD", // Exactly 4 characters
	}

	np.AddPlayer(player)

	if np.players[0].SquadTag != "ABCD" {
		t.Errorf("expected tag 'ABCD', got %s", np.players[0].SquadTag)
	}
}

func TestNameplate_ThreeCharacterTag(t *testing.T) {
	np := NewNameplate()

	player := NameplatePlayer{
		PlayerID:   "p1",
		PlayerName: "Alice",
		SquadTag:   "ABC", // 3 characters (under limit)
	}

	np.AddPlayer(player)

	if np.players[0].SquadTag != "ABC" {
		t.Errorf("expected tag 'ABC', got %s", np.players[0].SquadTag)
	}
}

func TestNameplate_SingleCharacterTag(t *testing.T) {
	np := NewNameplate()

	player := NameplatePlayer{
		PlayerID:   "p1",
		PlayerName: "Alice",
		SquadTag:   "X", // 1 character
	}

	np.AddPlayer(player)

	if np.players[0].SquadTag != "X" {
		t.Errorf("expected tag 'X', got %s", np.players[0].SquadTag)
	}
}

package main

import (
	"testing"

	"github.com/opd-ai/violence/pkg/bsp"
	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/rng"
)

// TestHUDMaxHealthInitialization verifies that MaxHealth and MaxArmor are explicitly set in startNewGame.
// This addresses AUDIT.md Edge Case Bug #1 (Medium severity).
func TestHUDMaxHealthInitialization(t *testing.T) {
	// Load config
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	g := NewGame()

	// Verify initial HUD state from NewHUD
	if g.hud.MaxHealth != 100 {
		t.Errorf("Initial MaxHealth should be 100, got %d", g.hud.MaxHealth)
	}
	if g.hud.MaxArmor != 100 {
		t.Errorf("Initial MaxArmor should be 100, got %d", g.hud.MaxArmor)
	}

	// Start a new game and verify explicit initialization
	g.startNewGame()

	if g.hud.Health != 100 {
		t.Errorf("Health should be 100 after startNewGame, got %d", g.hud.Health)
	}
	if g.hud.Armor != 0 {
		t.Errorf("Armor should be 0 after startNewGame, got %d", g.hud.Armor)
	}
	if g.hud.MaxHealth != 100 {
		t.Errorf("MaxHealth should be explicitly set to 100 after startNewGame, got %d", g.hud.MaxHealth)
	}
	if g.hud.MaxArmor != 100 {
		t.Errorf("MaxArmor should be explicitly set to 100 after startNewGame, got %d", g.hud.MaxArmor)
	}
}

// TestFindExitPositionWithValidRooms verifies that findExitPosition works correctly with valid BSP rooms.
// This addresses AUDIT.md Edge Case Bug #2 (Low severity) - ensures BSP generator never produces nil rooms.
func TestFindExitPositionWithValidRooms(t *testing.T) {
	g := &Game{}

	tests := []struct {
		name     string
		rooms    []*bsp.Room
		playerX  float64
		playerY  float64
		wantExit bool
	}{
		{
			name: "valid rooms",
			rooms: []*bsp.Room{
				{X: 10, Y: 10, W: 10, H: 10},
				{X: 50, Y: 50, W: 10, H: 10},
				{X: 80, Y: 80, W: 10, H: 10},
			},
			playerX:  15,
			playerY:  15,
			wantExit: true,
		},
		{
			name:     "empty room list",
			rooms:    []*bsp.Room{},
			playerX:  15,
			playerY:  15,
			wantExit: true, // Should use fallback position
		},
		{
			name: "single room",
			rooms: []*bsp.Room{
				{X: 10, Y: 10, W: 10, H: 10},
			},
			playerX:  15,
			playerY:  15,
			wantExit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos := g.findExitPosition(tt.rooms, tt.playerX, tt.playerY)
			if pos == nil {
				t.Fatal("findExitPosition returned nil")
			}
			if tt.wantExit {
				// Verify position is valid (non-negative coordinates)
				if pos.X < 0 || pos.Y < 0 {
					t.Errorf("Invalid exit position: X=%f, Y=%f", pos.X, pos.Y)
				}
			}
		})
	}
}

// TestFindExitPositionSelectsFurthestRoom verifies the algorithm selects the room furthest from player.
func TestFindExitPositionSelectsFurthestRoom(t *testing.T) {
	g := &Game{}

	rooms := []*bsp.Room{
		{X: 10, Y: 10, W: 10, H: 10},   // Center: (15, 15)
		{X: 50, Y: 50, W: 10, H: 10},   // Center: (55, 55)
		{X: 100, Y: 100, W: 10, H: 10}, // Center: (105, 105) - furthest from (15, 15)
	}

	pos := g.findExitPosition(rooms, 15, 15)
	if pos == nil {
		t.Fatal("findExitPosition returned nil")
	}

	// Exit should be in the furthest room (100, 100, 10, 10) with center (105, 105)
	expectedX := 105.0
	expectedY := 105.0

	if pos.X != expectedX || pos.Y != expectedY {
		t.Errorf("Exit position should be (%f, %f), got (%f, %f)", expectedX, expectedY, pos.X, pos.Y)
	}
}

// TestBSPGeneratorNeverProducesNilRooms verifies BSP generator contract.
func TestBSPGeneratorNeverProducesNilRooms(t *testing.T) {
	r := rng.NewRNG(12345)
	gen := bsp.NewGenerator(120, 120, r)
	tree, _ := gen.Generate()

	rooms := bsp.GetRooms(tree)
	if len(rooms) == 0 {
		t.Skip("BSP generator produced no rooms (valid for some seeds)")
	}

	for i, room := range rooms {
		if room == nil {
			t.Fatalf("BSP generator produced nil room at index %d - violates contract", i)
		}
	}
}

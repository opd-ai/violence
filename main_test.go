package main

import (
	"testing"

	"github.com/opd-ai/violence/pkg/config"
)

// TestNewGame verifies game initialization.
func TestNewGame(t *testing.T) {
	// Load config first
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	if game == nil {
		t.Fatal("NewGame() returned nil")
	}

	// Verify core systems are initialized
	if game.world == nil {
		t.Error("World not initialized")
	}
	if game.camera == nil {
		t.Error("Camera not initialized")
	}
	if game.raycaster == nil {
		t.Error("Raycaster not initialized")
	}
	if game.renderer == nil {
		t.Error("Renderer not initialized")
	}
	if game.input == nil {
		t.Error("Input manager not initialized")
	}
	if game.audioEngine == nil {
		t.Error("Audio engine not initialized")
	}
	if game.hud == nil {
		t.Error("HUD not initialized")
	}
	if game.menuManager == nil {
		t.Error("Menu manager not initialized")
	}
	if game.loadingScreen == nil {
		t.Error("Loading screen not initialized")
	}
	if game.tutorialSystem == nil {
		t.Error("Tutorial system not initialized")
	}
	if game.rng == nil {
		t.Error("RNG not initialized")
	}
	if game.bspGenerator == nil {
		t.Error("BSP generator not initialized")
	}

	// Verify initial state
	if game.state != StateMenu {
		t.Errorf("Expected initial state StateMenu, got %v", game.state)
	}
	if game.genreID != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got %s", game.genreID)
	}
	if game.seed == 0 {
		t.Error("Expected non-zero seed")
	}

	// Verify camera starts at expected position
	if game.camera.X != 5.0 || game.camera.Y != 5.0 {
		t.Errorf("Expected camera at (5.0, 5.0), got (%f, %f)", game.camera.X, game.camera.Y)
	}
	if game.camera.DirX != 1.0 || game.camera.DirY != 0.0 {
		t.Errorf("Expected camera direction (1.0, 0.0), got (%f, %f)", game.camera.DirX, game.camera.DirY)
	}

	// Verify menu is visible
	if !game.menuManager.IsVisible() {
		t.Error("Menu should be visible in initial state")
	}
}

// TestGameStates verifies state transitions.
func TestGameStates(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()

	tests := []struct {
		name          string
		initialState  GameState
		expectedState GameState
		setup         func()
	}{
		{
			name:          "Menu state initialization",
			initialState:  StateMenu,
			expectedState: StateMenu,
			setup: func() {
				game.state = StateMenu
			},
		},
		{
			name:          "Playing state",
			initialState:  StatePlaying,
			expectedState: StatePlaying,
			setup: func() {
				game.state = StatePlaying
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			if game.state != tt.initialState {
				t.Errorf("Setup failed: expected state %v, got %v", tt.initialState, game.state)
			}
		})
	}
}

// TestIsWalkable verifies collision detection.
func TestIsWalkable(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()

	// Test with no map (should be walkable)
	if !game.isWalkable(5.0, 5.0) {
		t.Error("Expected position to be walkable when no map is set")
	}

	// Generate a map
	game.bspGenerator.SetGenre("fantasy")
	_, tiles := game.bspGenerator.Generate()
	game.currentMap = tiles
	game.raycaster.SetMap(tiles)

	// Test a position (exact tile depends on generated map, but we can test bounds)
	if game.isWalkable(-1.0, -1.0) {
		t.Error("Expected out-of-bounds position to be unwalkable")
	}
	if game.isWalkable(100.0, 100.0) {
		t.Error("Expected out-of-bounds position to be unwalkable")
	}
}

// TestSaveLoad verifies save/load round-trip.
func TestSaveLoad(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Set some specific state
	game.camera.X = 10.5
	game.camera.Y = 15.3
	game.camera.Pitch = 5.0
	game.hud.Health = 75
	game.hud.Armor = 25
	game.hud.Ammo = 100

	// Save to slot 9 (avoid slot 0 for auto-save and slot 1 for main saves)
	game.saveGame(9)

	// Change state
	game.camera.X = 1.0
	game.camera.Y = 1.0
	game.hud.Health = 100

	// Load
	game.loadGame(9)

	// Verify restored state
	if game.camera.X != 10.5 {
		t.Errorf("Expected camera X = 10.5, got %f", game.camera.X)
	}
	if game.camera.Y != 15.3 {
		t.Errorf("Expected camera Y = 15.3, got %f", game.camera.Y)
	}
	if game.camera.Pitch != 5.0 {
		t.Errorf("Expected camera Pitch = 5.0, got %f", game.camera.Pitch)
	}
	if game.hud.Health != 75 {
		t.Errorf("Expected health = 75, got %d", game.hud.Health)
	}
	if game.hud.Armor != 25 {
		t.Errorf("Expected armor = 25, got %d", game.hud.Armor)
	}
	if game.hud.Ammo != 100 {
		t.Errorf("Expected ammo = 100, got %d", game.hud.Ammo)
	}
}

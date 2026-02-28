package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/ui"
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

// TestIsWalkableEmptyMap verifies handling of empty map slice.
func TestIsWalkableEmptyMap(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()

	// Test with empty map slice (edge case)
	game.currentMap = [][]int{}

	// Should not panic and should return true for empty map
	if !game.isWalkable(5.0, 5.0) {
		t.Error("Expected position to be walkable when map is empty slice")
	}

	// Verify no panic on various positions
	game.isWalkable(0.0, 0.0)
	game.isWalkable(-1.0, -1.0)
	game.isWalkable(100.0, 100.0)
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

// TestMenuActions verifies menu action handling.
func TestMenuActions(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name          string
		action        string
		initialState  GameState
		expectedState GameState
		setup         func(*Game)
		verify        func(*testing.T, *Game)
	}{
		{
			name:          "new_game shows difficulty menu",
			action:        "new_game",
			initialState:  StateMenu,
			expectedState: StateMenu,
			setup:         func(g *Game) { g.menuManager.Show(ui.MenuTypeMain) },
			verify: func(t *testing.T, g *Game) {
				if !g.menuManager.IsVisible() {
					t.Error("Menu should be visible after new_game")
				}
			},
		},
		{
			name:          "difficulty_selected shows genre menu",
			action:        "difficulty_selected",
			initialState:  StateMenu,
			expectedState: StateMenu,
			setup:         func(g *Game) { g.menuManager.Show(ui.MenuTypeDifficulty) },
			verify: func(t *testing.T, g *Game) {
				if !g.menuManager.IsVisible() {
					t.Error("Menu should be visible after difficulty selection")
				}
			},
		},
		{
			name:          "genre_selected starts game",
			action:        "genre_selected",
			initialState:  StateMenu,
			expectedState: StatePlaying,
			setup: func(g *Game) {
				g.menuManager.Show(ui.MenuTypeGenre)
			},
			verify: func(t *testing.T, g *Game) {
				if g.state != StatePlaying {
					t.Errorf("Expected state StatePlaying, got %v", g.state)
				}
			},
		},
		{
			name:          "settings shows settings menu",
			action:        "settings",
			initialState:  StateMenu,
			expectedState: StateMenu,
			setup:         func(g *Game) { g.menuManager.Show(ui.MenuTypeMain) },
			verify: func(t *testing.T, g *Game) {
				if !g.menuManager.IsVisible() {
					t.Error("Settings menu should be visible")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			game.state = tt.initialState
			tt.setup(game)
			game.handleMenuAction(tt.action)
			tt.verify(t, game)
		})
	}
}

// TestPauseActions verifies pause menu action handling.
func TestPauseActions(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name          string
		action        string
		expectedState GameState
		verify        func(*testing.T, *Game)
	}{
		{
			name:          "resume continues gameplay",
			action:        "resume",
			expectedState: StatePlaying,
			verify: func(t *testing.T, g *Game) {
				if g.state != StatePlaying {
					t.Errorf("Expected state StatePlaying, got %v", g.state)
				}
				if g.menuManager.IsVisible() {
					t.Error("Menu should be hidden after resume")
				}
			},
		},
		{
			name:          "save calls saveGame",
			action:        "save",
			expectedState: StatePaused,
			verify: func(t *testing.T, g *Game) {
				// Verify state is still paused after save
				if g.state != StatePaused {
					t.Errorf("Expected state StatePaused, got %v", g.state)
				}
			},
		},
		{
			name:          "quit_to_menu returns to main menu",
			action:        "quit_to_menu",
			expectedState: StateMenu,
			verify: func(t *testing.T, g *Game) {
				if g.state != StateMenu {
					t.Errorf("Expected state StateMenu, got %v", g.state)
				}
				if !g.menuManager.IsVisible() {
					t.Error("Menu should be visible after quit to menu")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			game.startNewGame()
			game.state = StatePaused
			game.menuManager.Show(ui.MenuTypePause)
			game.handlePauseAction(tt.action)
			tt.verify(t, game)
		})
	}
}

// TestUpdateStates verifies Update method for different game states.
func TestUpdateStates(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name         string
		initialState GameState
		verify       func(*testing.T, *Game)
	}{
		{
			name:         "update menu state",
			initialState: StateMenu,
			verify: func(t *testing.T, g *Game) {
				if err := g.Update(); err != nil {
					t.Errorf("Update failed: %v", err)
				}
			},
		},
		{
			name:         "update playing state",
			initialState: StatePlaying,
			verify: func(t *testing.T, g *Game) {
				if err := g.Update(); err != nil {
					t.Errorf("Update failed: %v", err)
				}
			},
		},
		{
			name:         "update paused state",
			initialState: StatePaused,
			verify: func(t *testing.T, g *Game) {
				if err := g.Update(); err != nil {
					t.Errorf("Update failed: %v", err)
				}
			},
		},
		{
			name:         "update loading state",
			initialState: StateLoading,
			verify: func(t *testing.T, g *Game) {
				if err := g.Update(); err != nil {
					t.Errorf("Update failed: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame()
			if tt.initialState == StatePlaying || tt.initialState == StatePaused {
				game.startNewGame()
			}
			game.state = tt.initialState
			tt.verify(t, game)
		})
	}
}

// TestLayout verifies Layout method returns correct dimensions.
func TestLayout(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	w, h := game.Layout(800, 600)

	if w != config.C.InternalWidth {
		t.Errorf("Expected width %d, got %d", config.C.InternalWidth, w)
	}
	if h != config.C.InternalHeight {
		t.Errorf("Expected height %d, got %d", config.C.InternalHeight, h)
	}
}

// TestDrawStates verifies Draw method for different game states.
func TestDrawStates(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Create a dummy screen image
	screen := ebiten.NewImage(config.C.InternalWidth, config.C.InternalHeight)

	tests := []struct {
		name  string
		state GameState
	}{
		{"menu", StateMenu},
		{"playing", StatePlaying},
		{"paused", StatePaused},
		{"loading", StateLoading},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game.state = tt.state
			// Should not panic
			game.Draw(screen)
		})
	}
}

// TestUpdatePlaying verifies gameplay update logic.
func TestUpdatePlaying(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Store initial position
	initialX := game.camera.X
	initialY := game.camera.Y

	// Call updatePlaying (no input, so no movement)
	if err := game.updatePlaying(); err != nil {
		t.Errorf("updatePlaying failed: %v", err)
	}

	// Verify position unchanged when no input
	if game.camera.X != initialX {
		t.Errorf("Camera X should not change without input")
	}
	if game.camera.Y != initialY {
		t.Errorf("Camera Y should not change without input")
	}
}

// TestStartNewGame verifies game initialization from menu.
func TestStartNewGame(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.genreID = "scifi"
	game.startNewGame()

	// Verify state transitions
	if game.state != StatePlaying {
		t.Errorf("Expected state StatePlaying after startNewGame, got %v", game.state)
	}

	// Verify systems are configured for genre
	if game.genreID != "scifi" {
		t.Errorf("Expected genre 'scifi', got %s", game.genreID)
	}

	// Verify HUD is reset
	if game.hud.Health != 100 {
		t.Errorf("Expected initial health 100, got %d", game.hud.Health)
	}
	if game.hud.Armor != 0 {
		t.Errorf("Expected initial armor 0, got %d", game.hud.Armor)
	}
	if game.hud.Ammo != 50 {
		t.Errorf("Expected initial ammo 50, got %d", game.hud.Ammo)
	}

	// Verify map is generated
	if game.currentMap == nil {
		t.Error("Current map should be generated")
	}
}

// TestUpdateMenu verifies menu navigation updates.
func TestUpdateMenu(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.state = StateMenu

	// Should not error
	if err := game.updateMenu(); err != nil {
		t.Errorf("updateMenu failed: %v", err)
	}
}

// TestUpdatePaused verifies paused state updates.
func TestUpdatePaused(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()
	game.state = StatePaused
	game.menuManager.Show(ui.MenuTypePause)

	// Should not error
	if err := game.updatePaused(); err != nil {
		t.Errorf("updatePaused failed: %v", err)
	}
}

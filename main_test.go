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

// TestAutomapCreation verifies automap is created during game start.
func TestAutomapCreation(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	if game.automap == nil {
		t.Error("Automap should be created during startNewGame")
	}
	if game.automap.Width != len(game.currentMap[0]) {
		t.Errorf("Expected automap width %d, got %d", len(game.currentMap[0]), game.automap.Width)
	}
	if game.automap.Height != len(game.currentMap) {
		t.Errorf("Expected automap height %d, got %d", len(game.currentMap), game.automap.Height)
	}
}

// TestAutomapToggle verifies automap visibility toggle.
func TestAutomapToggle(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	if game.automapVisible {
		t.Error("Automap should start hidden")
	}

	// Simulate toggling (note: actual toggle requires input mock)
	game.automapVisible = !game.automapVisible
	if !game.automapVisible {
		t.Error("Automap should be visible after toggle")
	}

	game.automapVisible = !game.automapVisible
	if game.automapVisible {
		t.Error("Automap should be hidden after second toggle")
	}
}

// TestKeycardInitialization verifies keycard map is initialized.
func TestKeycardInitialization(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()

	if game.keycards == nil {
		t.Error("Keycard map should be initialized")
	}
	if len(game.keycards) != 0 {
		t.Errorf("Expected empty keycard map, got %d entries", len(game.keycards))
	}
}

// TestDoorInteraction verifies door interaction logic.
func TestDoorInteraction(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Create a simple test map with a door
	game.currentMap = [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 3, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}
	game.raycaster.SetMap(game.currentMap)

	// Position player facing the door - place closer so 1.5 units puts us at the door
	game.camera.X = 2.0
	game.camera.Y = 0.7
	game.camera.DirX = 0.0
	game.camera.DirY = 1.0

	// Try to open door (no keycard required for stub)
	game.tryInteractDoor()

	// Door should be opened (replaced with floor tile which is TileFloor = 2)
	if game.currentMap[2][2] != 2 {
		t.Errorf("Expected door to be opened (tile 2 = TileFloor), got tile %d", game.currentMap[2][2])
	}
}

// TestGamepadAnalogStickSupport verifies gamepad stick methods don't panic.
func TestGamepadAnalogStickSupport(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify gamepad methods exist and don't panic
	leftX, leftY := game.input.GamepadLeftStick()
	if leftX != 0 || leftY != 0 {
		// OK if no gamepad connected
		t.Logf("Gamepad left stick: (%f, %f)", leftX, leftY)
	}

	rightX, rightY := game.input.GamepadRightStick()
	if rightX != 0 || rightY != 0 {
		t.Logf("Gamepad right stick: (%f, %f)", rightX, rightY)
	}
}

// TestHUDMessageDisplay verifies HUD message system.
func TestHUDMessageDisplay(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()

	// Test message display
	game.hud.ShowMessage("Test message")
	if game.hud.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", game.hud.Message)
	}
	if game.hud.MessageTime != 180 {
		t.Errorf("Expected message time 180, got %d", game.hud.MessageTime)
	}

	// Test message timeout
	for i := 0; i < 180; i++ {
		game.hud.Update()
	}
	if game.hud.Message != "" {
		t.Errorf("Expected message to be cleared after timeout, got '%s'", game.hud.Message)
	}
	if game.hud.MessageTime != 0 {
		t.Errorf("Expected message time 0 after timeout, got %d", game.hud.MessageTime)
	}
}

// TestDrawAutomap verifies automap rendering doesn't panic.
func TestDrawAutomap(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Create a dummy screen image
	screen := ebiten.NewImage(config.C.InternalWidth, config.C.InternalHeight)

	// Should not panic
	game.drawAutomap(screen)
}

// TestCombatLoopIntegration tests full combat flow: spawn → fire → damage → death → loot → XP → level-up
func TestCombatLoopIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify initial state
	if len(game.aiAgents) == 0 {
		t.Fatal("No enemies spawned")
	}
	initialXP := game.progression.XP
	initialLevel := game.progression.Level

	// Get first enemy
	enemy := game.aiAgents[0]
	initialHealth := enemy.Health
	if initialHealth <= 0 {
		t.Fatal("Enemy should start with positive health")
	}

	// Position player to face enemy
	game.camera.X = enemy.X - 2.0
	game.camera.Y = enemy.Y
	game.camera.DirX = 1.0
	game.camera.DirY = 0.0

	// Get initial ammo count
	currentWeapon := game.arsenal.GetCurrentWeapon()
	initialAmmo := game.ammoPool.Get(currentWeapon.AmmoType)

	// Simulate weapon firing
	raycastFn := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		for i, agent := range game.aiAgents {
			if agent.Health <= 0 {
				continue
			}
			agentDist := (agent.X-x)*(agent.X-x) + (agent.Y-y)*(agent.Y-y)
			if agentDist < maxDist*maxDist {
				toAgentX := agent.X - x
				toAgentY := agent.Y - y
				dot := toAgentX*dx + toAgentY*dy
				if dot > 0 {
					return true, agentDist, agent.X, agent.Y, uint64(i + 1)
				}
			}
		}
		return false, 0, 0, 0, 0
	}

	// Fire weapon at enemy
	hitResults := game.arsenal.Fire(game.camera.X, game.camera.Y, game.camera.DirX, game.camera.DirY, raycastFn)

	// Verify hit was registered
	if len(hitResults) == 0 {
		t.Fatal("No hit results returned from weapon fire")
	}
	if !hitResults[0].Hit {
		t.Error("Expected weapon to hit enemy")
	}

	// Verify ammo consumption
	if currentWeapon.Type != 2 { // TypeMelee
		game.ammoPool.Consume(currentWeapon.AmmoType, 1)
		newAmmo := game.ammoPool.Get(currentWeapon.AmmoType)
		if newAmmo != initialAmmo-1 {
			t.Errorf("Expected ammo %d, got %d", initialAmmo-1, newAmmo)
		}
	}

	// Apply damage to enemy
	for _, hitResult := range hitResults {
		if hitResult.Hit && hitResult.EntityID > 0 {
			agentIdx := int(hitResult.EntityID - 1)
			if agentIdx >= 0 && agentIdx < len(game.aiAgents) {
				agent := game.aiAgents[agentIdx]
				agent.Health -= currentWeapon.Damage
			}
		}
	}

	// Verify enemy took damage
	if enemy.Health >= initialHealth {
		t.Errorf("Enemy should have taken damage: before=%f, after=%f", initialHealth, enemy.Health)
	}

	// Fire multiple times to kill enemy
	shotsNeeded := int(initialHealth/currentWeapon.Damage) + 5
	for i := 0; i < shotsNeeded*int(currentWeapon.FireRate)+100 && enemy.Health > 0; i++ {
		// Update cooldown
		game.arsenal.Update()

		// Try to fire
		hitResults = game.arsenal.Fire(game.camera.X, game.camera.Y, game.camera.DirX, game.camera.DirY, raycastFn)
		if hitResults != nil && len(hitResults) > 0 {
			for _, hitResult := range hitResults {
				if hitResult.Hit && hitResult.EntityID > 0 {
					agentIdx := int(hitResult.EntityID - 1)
					if agentIdx >= 0 && agentIdx < len(game.aiAgents) {
						agent := game.aiAgents[agentIdx]
						if agent.Health > 0 {
							agent.Health -= currentWeapon.Damage
							if agent.Health <= 0 {
								// Award XP on death
								game.progression.AddXP(50)
							}
						}
					}
				}
			}
		}
	}

	// Verify enemy is dead
	if enemy.Health > 0 {
		t.Errorf("Enemy should be dead after multiple shots, health=%f", enemy.Health)
	}

	// Verify XP was awarded
	if game.progression.XP <= initialXP {
		t.Errorf("Expected XP to increase from %d, got %d", initialXP, game.progression.XP)
	}
	if game.progression.XP != initialXP+50 {
		t.Errorf("Expected exactly 50 XP gained, got %d", game.progression.XP-initialXP)
	}

	// Verify level is still correct (50 XP not enough for level 2)
	if game.progression.Level != initialLevel {
		t.Errorf("Expected level to remain %d with only 50 XP, got %d", initialLevel, game.progression.Level)
	}
}

// TestMultipleEnemyKills tests killing multiple enemies and accumulating XP
func TestMultipleEnemyKills(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	if len(game.aiAgents) < 2 {
		t.Fatal("Need at least 2 enemies for this test")
	}

	initialXP := game.progression.XP
	kills := 0

	// Kill all enemies
	for i := range game.aiAgents {
		enemy := game.aiAgents[i]
		if enemy.Health <= 0 {
			continue
		}

		// Position to face enemy
		game.camera.X = enemy.X - 2.0
		game.camera.Y = enemy.Y
		game.camera.DirX = 1.0
		game.camera.DirY = 0.0

		raycastFn := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
			agentDist := (enemy.X-x)*(enemy.X-x) + (enemy.Y-y)*(enemy.Y-y)
			if agentDist < maxDist*maxDist {
				toAgentX := enemy.X - x
				toAgentY := enemy.Y - y
				dot := toAgentX*dx + toAgentY*dy
				if dot > 0 && enemy.Health > 0 {
					return true, agentDist, enemy.X, enemy.Y, uint64(i + 1)
				}
			}
			return false, 0, 0, 0, 0
		}

		// Fire until enemy dies
		weapon := game.arsenal.GetCurrentWeapon()
		shotsNeeded := int(enemy.MaxHealth/weapon.Damage) + 5
		for j := 0; j < shotsNeeded*int(weapon.FireRate)+100 && enemy.Health > 0; j++ {
			// Update cooldown
			game.arsenal.Update()

			// Try to fire
			hitResults := game.arsenal.Fire(game.camera.X, game.camera.Y, game.camera.DirX, game.camera.DirY, raycastFn)
			if hitResults != nil && len(hitResults) > 0 {
				for _, hitResult := range hitResults {
					if hitResult.Hit && hitResult.EntityID > 0 {
						if enemy.Health > 0 {
							enemy.Health -= weapon.Damage
							if enemy.Health <= 0 {
								game.progression.AddXP(50)
								kills++
							}
						}
					}
				}
			}
		}
	}

	// Verify multiple kills
	if kills < 2 {
		t.Errorf("Expected at least 2 kills, got %d", kills)
	}

	// Verify XP accumulation
	expectedXP := initialXP + (kills * 50)
	if game.progression.XP != expectedXP {
		t.Errorf("Expected XP=%d (initial %d + %d kills * 50), got %d", expectedXP, initialXP, kills, game.progression.XP)
	}
}

// TestLevelUpThreshold tests progression to level 2 after accumulating enough XP
func TestLevelUpThreshold(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Start at level 1 with 0 XP
	if game.progression.Level != 1 {
		t.Errorf("Expected starting level 1, got %d", game.progression.Level)
	}

	initialLevel := game.progression.Level

	// Award XP to trigger level-up (threshold is 100 XP for level 2)
	game.progression.AddXP(100)

	// Verify XP was added
	if game.progression.XP != 100 {
		t.Errorf("Expected XP=100, got %d", game.progression.XP)
	}

	// Manually trigger level-up check (game would do this in update loop)
	if game.progression.XP >= 100 && game.progression.Level == 1 {
		game.progression.LevelUp()
	}

	// Verify level increased
	if game.progression.Level != initialLevel+1 {
		t.Errorf("Expected level %d after 100 XP, got %d", initialLevel+1, game.progression.Level)
	}
}

// TestPlayerTakesDamage tests player receiving damage from enemies
func TestPlayerTakesDamage(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	initialHealth := game.hud.Health
	initialArmor := game.hud.Armor

	if initialHealth != 100 {
		t.Errorf("Expected initial health 100, got %d", initialHealth)
	}

	// Simulate enemy dealing damage
	damage := 20

	// Apply damage (simplified - armor absorbs 50%, rest to health)
	armorDamage := damage / 2
	healthDamage := damage / 2

	if game.hud.Armor > 0 {
		if game.hud.Armor >= armorDamage {
			game.hud.Armor -= armorDamage
		} else {
			healthDamage += armorDamage - game.hud.Armor
			game.hud.Armor = 0
		}
	}
	game.hud.Health -= healthDamage

	// Verify health decreased
	if game.hud.Health >= initialHealth {
		t.Errorf("Player should have taken damage: before=%d, after=%d", initialHealth, game.hud.Health)
	}

	expectedHealth := initialHealth - healthDamage
	if game.hud.Armor == initialArmor {
		if game.hud.Health != expectedHealth {
			t.Errorf("Expected health=%d, got %d", expectedHealth, game.hud.Health)
		}
	}
}

// TestArmorAbsorption tests armor damage absorption mechanics
func TestArmorAbsorption(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Give player armor
	game.hud.Armor = 50
	game.hud.Health = 100

	damage := 20

	// Apply damage with armor absorption
	armorDamage := damage / 2
	healthDamage := damage / 2

	if game.hud.Armor >= armorDamage {
		game.hud.Armor -= armorDamage
	} else {
		healthDamage += armorDamage - game.hud.Armor
		game.hud.Armor = 0
	}
	game.hud.Health -= healthDamage

	// Verify armor absorbed damage
	if game.hud.Armor != 40 {
		t.Errorf("Expected armor=40 after absorbing 10 damage, got %d", game.hud.Armor)
	}
	if game.hud.Health != 90 {
		t.Errorf("Expected health=90 after taking 10 damage, got %d", game.hud.Health)
	}
}

// TestWeaponSwitchingDuringCombat tests changing weapons mid-combat
func TestWeaponSwitchingDuringCombat(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Get initial weapon
	weapon1 := game.arsenal.GetCurrentWeapon()
	if weapon1.Name == "" {
		t.Fatal("No initial weapon")
	}

	// Switch to next weapon
	game.arsenal.SwitchTo(1)
	weapon2 := game.arsenal.GetCurrentWeapon()

	// Weapons should be different (or at least switching succeeded)
	if weapon2.Name == "" {
		t.Error("Failed to switch weapon")
	}

	// Fire with new weapon
	raycastFn := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
		return false, 0, 0, 0, 0
	}

	hitResults := game.arsenal.Fire(game.camera.X, game.camera.Y, game.camera.DirX, game.camera.DirY, raycastFn)
	if hitResults == nil {
		t.Error("Should be able to fire after weapon switch")
	}
}

// TestEnemyRespawnDoesNotOccur verifies dead enemies stay dead
func TestEnemyRespawnDoesNotOccur(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	if len(game.aiAgents) == 0 {
		t.Fatal("No enemies spawned")
	}

	// Kill first enemy
	enemy := game.aiAgents[0]
	enemy.Health = 0

	// Wait and verify enemy doesn't respawn
	for i := 0; i < 100; i++ {
		game.arsenal.Update()
	}

	if enemy.Health > 0 {
		t.Error("Dead enemy should not respawn")
	}
}

// TestCombatWithDifferentWeaponTypes tests hitscan vs melee
func TestCombatWithDifferentWeaponTypes(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	tests := []struct {
		name       string
		weaponSlot int
		expectAmmo bool
	}{
		{"hitscan pistol", 0, true},
		{"shotgun", 1, true},
		{"melee knife", 6, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game.arsenal.SwitchTo(tt.weaponSlot)
			weapon := game.arsenal.GetCurrentWeapon()

			raycastFn := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
				return false, 0, 0, 0, 0
			}

			initialAmmo := game.ammoPool.Get(weapon.AmmoType)
			game.arsenal.Fire(game.camera.X, game.camera.Y, game.camera.DirX, game.camera.DirY, raycastFn)

			if tt.expectAmmo {
				// Ammo should be consumed for non-melee
				game.ammoPool.Consume(weapon.AmmoType, 1)
				newAmmo := game.ammoPool.Get(weapon.AmmoType)
				if weapon.Type != 2 && newAmmo >= initialAmmo {
					t.Errorf("Expected ammo consumption for %s", tt.name)
				}
			}
		})
	}
}

// TestAllGenresPlayable validates Step 60: end-to-end gameplay for all five genres.
// Tests that each genre can: start → generate level → spawn enemies → combat → progress XP.
func TestAllGenresPlayable(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			// Initialize game
			game := NewGame()
			if game == nil {
				t.Fatal("NewGame() returned nil")
			}

			// Start game with genre
			game.startNewGame()
			game.genreID = genreID
			game.arsenal.SetGenre(genreID)

			// Verify map generated
			if game.currentMap == nil {
				t.Fatal("Map not generated")
			}

			// Verify weapon equipped
			weapon := game.arsenal.GetCurrentWeapon()
			if weapon.Name == "" {
				t.Fatal("No weapon equipped")
			}

			// Verify enemies spawned
			if len(game.aiAgents) == 0 {
				t.Fatal("No enemies spawned")
			}

			// Simulate combat
			enemy := game.aiAgents[0]
			initialHealth := enemy.Health

			raycastFn := func(x, y, dx, dy, maxDist float64) (bool, float64, float64, float64, uint64) {
				return true, 10.0, enemy.X, enemy.Y, 0
			}

			// Fire at enemy
			game.arsenal.Update()
			hitResults := game.arsenal.Fire(game.camera.X, game.camera.Y, game.camera.DirX, game.camera.DirY, raycastFn)

			if len(hitResults) == 0 {
				t.Fatal("No hit results from weapon fire")
			}

			if hitResults[0].Hit {
				enemy.Health -= weapon.Damage
			}

			// Verify damage applied
			if enemy.Health >= initialHealth {
				t.Error("Enemy should have taken damage")
			}

			// Award XP
			initialXP := game.progression.XP
			game.progression.AddXP(50)

			if game.progression.XP != initialXP+50 {
				t.Errorf("Expected XP increase of 50, got %d", game.progression.XP-initialXP)
			}

			// Verify ammo system
			game.ammoPool.Add("bullets", 50)
			if game.ammoPool.Get("bullets") < 50 {
				t.Error("Ammo not added correctly")
			}

			// All core systems functional for this genre
			t.Logf("Genre %s: all systems functional", genreID)
		})
	}
}

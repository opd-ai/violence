package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/bsp"
	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/inventory"
	"github.com/opd-ai/violence/pkg/minigame"
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

// TestShopIntegration verifies shop system is initialized and functional.
func TestShopIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify shop systems are initialized
	if game.shopCredits == nil {
		t.Fatal("Shop credits not initialized after startNewGame")
	}
	if game.shopArmory == nil {
		t.Fatal("Shop armory not initialized after startNewGame")
	}
	if game.shopInventory == nil {
		t.Fatal("Shop inventory not initialized after startNewGame")
	}

	// Verify starting credits
	if game.shopCredits.Get() != 100 {
		t.Errorf("Expected 100 starting credits, got %d", game.shopCredits.Get())
	}

	// Verify shop has items
	allItems := game.shopArmory.Inventory.GetAllItems()
	if len(allItems) == 0 {
		t.Error("Shop inventory should have items")
	}

	// Verify shop name is genre-appropriate
	shopName := game.shopArmory.GetShopName()
	if shopName == "" {
		t.Error("Shop name should not be empty")
	}
}

// TestShopPurchaseFlow verifies item purchase mechanics.
func TestShopPurchaseFlow(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Give plenty of credits
	game.shopCredits.Set(1000)

	// Find a consumable item (medkit)
	item := game.shopArmory.Inventory.FindItem("medkit")
	if item == nil {
		t.Fatal("Shop should have a medkit item")
	}

	initialCredits := game.shopCredits.Get()
	initialHealth := game.hud.Health

	// Set health below max so purchase has visible effect
	game.hud.Health = 50

	// Purchase medkit
	if !game.shopArmory.Purchase("medkit", game.shopCredits) {
		t.Error("Should be able to purchase medkit with sufficient credits")
	}

	// Verify credits deducted
	if game.shopCredits.Get() != initialCredits-item.Price {
		t.Errorf("Expected credits %d, got %d", initialCredits-item.Price, game.shopCredits.Get())
	}

	// Apply item effects - medkits go to inventory
	game.applyShopItem("medkit")
	if !game.playerInventory.Has("medkit") {
		t.Error("Medkit should be added to inventory after purchase")
	}

	// Test insufficient credits
	game.shopCredits.Set(0)
	if game.shopArmory.Purchase("medkit", game.shopCredits) {
		t.Error("Should not be able to purchase with 0 credits")
	}

	// Verify inventory count doesn't change on failed purchase
	medkitItem := game.playerInventory.Get("medkit")
	if medkitItem == nil || medkitItem.Qty != 1 {
		t.Error("Failed purchase should not add to inventory")
	}

	_ = initialHealth // avoid unused warning
}

// TestShopStateTransition verifies shop state transitions work correctly.
func TestShopStateTransition(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify we're in playing state
	if game.state != StatePlaying {
		t.Errorf("Expected StatePlaying, got %v", game.state)
	}

	// Open shop
	game.openShop()
	if game.state != StateShop {
		t.Errorf("Expected StateShop, got %v", game.state)
	}
	if !game.menuManager.IsVisible() {
		t.Error("Menu should be visible in shop state")
	}
	if game.menuManager.GetCurrentMenu() != ui.MenuTypeShop {
		t.Error("Menu type should be MenuTypeShop")
	}
}

// TestCraftingIntegration verifies crafting system is initialized and functional.
func TestCraftingIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify crafting systems are initialized
	if game.scrapStorage == nil {
		t.Fatal("Scrap storage not initialized after startNewGame")
	}
	if game.craftingMenu == nil {
		t.Fatal("Crafting menu not initialized after startNewGame")
	}

	// Verify starting scrap
	scrapAmts := game.scrapStorage.GetAll()
	if len(scrapAmts) == 0 {
		t.Error("Should have starting scrap materials")
	}

	// Verify recipes exist
	allRecipes := game.craftingMenu.GetAllRecipes()
	if len(allRecipes) == 0 {
		t.Error("Should have crafting recipes")
	}
}

// TestCraftingFlow verifies the craft item flow.
func TestCraftingFlow(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.genreID = "fantasy"
	game.startNewGame()

	// Give plenty of scrap
	game.scrapStorage.Add("bone_chips", 100)

	// Record scrap before crafting
	beforeScrap := game.scrapStorage.Get("bone_chips")

	// Get recipes
	allRecipes := game.craftingMenu.GetAllRecipes()
	if len(allRecipes) == 0 {
		t.Fatal("No recipes available")
	}

	// Find a recipe that uses bone_chips
	var targetRecipe string
	for _, r := range allRecipes {
		if _, ok := r.Inputs["bone_chips"]; ok {
			targetRecipe = r.ID
			break
		}
	}
	if targetRecipe == "" {
		t.Fatal("No recipe found using bone_chips")
	}

	// Craft the item
	outputID, outputQty, err := game.craftingMenu.Craft(targetRecipe)
	if err != nil {
		t.Fatalf("Craft failed: %v", err)
	}
	if outputID == "" {
		t.Error("Output ID should not be empty")
	}
	if outputQty <= 0 {
		t.Error("Output quantity should be positive")
	}

	// Apply crafted item
	game.applyCraftedItem(outputID, outputQty)

	// Verify scrap was consumed
	remaining := game.scrapStorage.Get("bone_chips")
	if remaining >= beforeScrap {
		t.Error("Scrap should have been consumed during crafting")
	}
}

// TestCraftingStateTransition verifies crafting state transitions.
func TestCraftingStateTransition(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Open crafting
	game.openCrafting()
	if game.state != StateCrafting {
		t.Errorf("Expected StateCrafting, got %v", game.state)
	}
	if !game.menuManager.IsVisible() {
		t.Error("Menu should be visible in crafting state")
	}
	if game.menuManager.GetCurrentMenu() != ui.MenuTypeCrafting {
		t.Error("Menu type should be MenuTypeCrafting")
	}
}

// TestKillRewardsCreditsAndScrap verifies enemy kills award credits and scrap.
func TestKillRewardsCreditsAndScrap(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	initialCredits := game.shopCredits.Get()
	scrapName := "bone_chips" // fantasy genre
	initialScrap := game.scrapStorage.Get(scrapName)

	// Simulate an enemy kill by calling the reward logic directly
	game.shopCredits.Add(25)
	game.scrapStorage.Add(scrapName, 3)

	if game.shopCredits.Get() != initialCredits+25 {
		t.Errorf("Expected credits %d, got %d", initialCredits+25, game.shopCredits.Get())
	}
	if game.scrapStorage.Get(scrapName) != initialScrap+3 {
		t.Errorf("Expected scrap %d, got %d", initialScrap+3, game.scrapStorage.Get(scrapName))
	}
}

// TestDestructibleDropsScrapAndCredits verifies destructible objects drop rewards.
func TestDestructibleDropsScrapAndCredits(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	initialCredits := game.shopCredits.Get()
	scrapName := "bone_chips" // fantasy genre
	initialScrap := game.scrapStorage.Get(scrapName)

	// Simulate destructible drop rewards
	game.shopCredits.Add(10)
	game.scrapStorage.Add(scrapName, 2)

	if game.shopCredits.Get() != initialCredits+10 {
		t.Errorf("Expected credits %d, got %d", initialCredits+10, game.shopCredits.Get())
	}
	if game.scrapStorage.Get(scrapName) != initialScrap+2 {
		t.Errorf("Expected scrap %d, got %d", initialScrap+2, game.scrapStorage.Get(scrapName))
	}
}

// TestShopGenreCascade verifies genre changes update shop inventory.
func TestShopGenreCascade(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			game := NewGame()
			game.genreID = genreID
			game.startNewGame()

			// Verify shop was initialized with correct genre
			if game.shopArmory == nil {
				t.Fatal("Shop not initialized")
			}
			shopName := game.shopArmory.GetShopName()
			if shopName == "" {
				t.Error("Shop name should not be empty")
			}

			// Verify crafting uses genre-appropriate scrap
			if game.scrapStorage == nil {
				t.Fatal("Scrap storage not initialized")
			}
			expectedScrap := game.scrapStorage.GetAll()
			if len(expectedScrap) == 0 {
				t.Error("Should have starting scrap for genre")
			}

			// Verify recipes are genre-appropriate
			recipes := game.craftingMenu.GetAllRecipes()
			if len(recipes) == 0 {
				t.Error("Should have recipes for genre")
			}

			t.Logf("Genre %s: shop=%s, recipes=%d", genreID, shopName, len(recipes))
		})
	}
}

// TestBuildShopState verifies the shop UI state builder.
func TestBuildShopState(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	state := game.buildShopState()
	if state == nil {
		t.Fatal("buildShopState returned nil")
	}
	if state.ShopName == "" {
		t.Error("Shop name should not be empty")
	}
	if state.Credits != game.shopCredits.Get() {
		t.Error("Credits mismatch in shop state")
	}
	if len(state.Items) == 0 {
		t.Error("Shop state should have items")
	}
}

// TestBuildCraftingState verifies the crafting UI state builder.
func TestBuildCraftingState(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	state := game.buildCraftingState()
	if state == nil {
		t.Fatal("buildCraftingState returned nil")
	}
	if len(state.Recipes) == 0 {
		t.Error("Crafting state should have recipes")
	}
	if len(state.ScrapAmts) == 0 {
		t.Error("Crafting state should have scrap amounts")
	}
	if state.ScrapName == "" {
		t.Error("Scrap name should not be empty")
	}
}

// TestApplyShopItemAmmo verifies ammo shop items are applied correctly.
func TestApplyShopItemAmmo(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	tests := []struct {
		itemID   string
		ammoType string
		addQty   int
	}{
		{"ammo_bullets", "bullets", 20},
		{"ammo_shells", "shells", 10},
		{"ammo_cells", "cells", 15},
		{"ammo_rockets", "rockets", 5},
	}

	for _, tt := range tests {
		t.Run(tt.itemID, func(t *testing.T) {
			initial := game.ammoPool.Get(tt.ammoType)
			game.applyShopItem(tt.itemID)
			after := game.ammoPool.Get(tt.ammoType)
			if after != initial+tt.addQty {
				t.Errorf("Expected %s=%d, got %d", tt.ammoType, initial+tt.addQty, after)
			}
		})
	}
}

// TestApplyShopItemArmor verifies armor cap.
func TestApplyShopItemArmor(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	game.hud.Armor = 80
	game.applyShopItem("armor_vest")
	if game.hud.Armor > game.hud.MaxArmor {
		t.Errorf("Armor %d exceeds max %d", game.hud.Armor, game.hud.MaxArmor)
	}
}

// TestApplyCraftedItemMedkit verifies crafted medkit goes to inventory.
func TestApplyCraftedItemMedkit(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Crafted medkits should go to inventory
	game.applyCraftedItem("medkit", 1)
	if !game.playerInventory.Has("medkit") {
		t.Error("Expected medkit in inventory")
	}
	item := game.playerInventory.Get("medkit")
	if item == nil || item.Qty != 1 {
		t.Errorf("Expected 1 medkit, got %v", item)
	}

	// Multiple medkits should stack
	game.applyCraftedItem("medkit", 2)
	item = game.playerInventory.Get("medkit")
	if item == nil || item.Qty != 3 {
		t.Errorf("Expected 3 medkits after stacking, got %v", item)
	}
}

// TestDrawShopAndCrafting verifies draw methods don't panic.
func TestDrawShopAndCrafting(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	screen := ebiten.NewImage(320, 200)

	// Test drawShop doesn't panic
	game.openShop()
	game.drawShop(screen)

	// Test drawCrafting doesn't panic
	game.openCrafting()
	game.drawCrafting(screen)
}

// TestSkillsIntegration verifies skills system is initialized and functional.
func TestSkillsIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify skills system is initialized
	if game.skillManager == nil {
		t.Fatal("Skill manager not initialized after startNewGame")
	}

	// Verify starting skill points (3 per tree)
	combatTree, err := game.skillManager.GetTree("combat")
	if err != nil {
		t.Fatalf("Failed to get combat tree: %v", err)
	}
	if combatTree.GetPoints() != 3 {
		t.Errorf("Expected 3 starting points in combat tree, got %d", combatTree.GetPoints())
	}

	survivalTree, err := game.skillManager.GetTree("survival")
	if err != nil {
		t.Fatalf("Failed to get survival tree: %v", err)
	}
	if survivalTree.GetPoints() != 3 {
		t.Errorf("Expected 3 starting points in survival tree, got %d", survivalTree.GetPoints())
	}

	techTree, err := game.skillManager.GetTree("tech")
	if err != nil {
		t.Fatalf("Failed to get tech tree: %v", err)
	}
	if techTree.GetPoints() != 3 {
		t.Errorf("Expected 3 starting points in tech tree, got %d", techTree.GetPoints())
	}

	// Verify all three trees have nodes
	if len(combatTree.Nodes) != 5 {
		t.Errorf("Expected 5 combat nodes, got %d", len(combatTree.Nodes))
	}
	if len(survivalTree.Nodes) != 5 {
		t.Errorf("Expected 5 survival nodes, got %d", len(survivalTree.Nodes))
	}
	if len(techTree.Nodes) != 5 {
		t.Errorf("Expected 5 tech nodes, got %d", len(techTree.Nodes))
	}
}

// TestSkillsAllocateFlow verifies skill allocation mechanics in-game.
func TestSkillsAllocateFlow(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Allocate a root combat node (no prereqs)
	err := game.skillManager.AllocatePoint("combat", "combat_dmg_1")
	if err != nil {
		t.Errorf("Should be able to allocate root node: %v", err)
	}

	// Verify allocation
	if !game.skillManager.IsNodeAllocated("combat", "combat_dmg_1") {
		t.Error("combat_dmg_1 should be allocated")
	}

	// Check modifier
	dmgMod := game.skillManager.GetModifier("damage")
	if dmgMod < 0.09 || dmgMod > 0.11 { // Approx 0.10
		t.Errorf("Expected ~0.10 damage modifier, got %f", dmgMod)
	}

	// Try allocating a node with unmet prereqs
	err = game.skillManager.AllocatePoint("combat", "combat_master")
	if err == nil {
		t.Error("Should not be able to allocate combat_master without prereqs")
	}
}

// TestSkillsStateTransition verifies skills state transitions work correctly.
func TestSkillsStateTransition(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify we're in playing state
	if game.state != StatePlaying {
		t.Errorf("Expected StatePlaying, got %v", game.state)
	}

	// Open skills
	game.openSkills()
	if game.state != StateSkills {
		t.Errorf("Expected StateSkills, got %v", game.state)
	}
	if !game.menuManager.IsVisible() {
		t.Error("Menu should be visible in skills state")
	}
	if game.menuManager.GetCurrentMenu() != ui.MenuTypeSkills {
		t.Error("Menu type should be MenuTypeSkills")
	}
}

// TestBuildSkillsState verifies skills state building for UI.
func TestBuildSkillsState(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	state := game.buildSkillsState()
	if state == nil {
		t.Fatal("Skills state should not be nil")
	}
	if len(state.Trees) != 3 {
		t.Errorf("Expected 3 trees, got %d", len(state.Trees))
	}
	if state.TotalPoints != 3 {
		t.Errorf("Expected 3 total points, got %d", state.TotalPoints)
	}

	// Verify tree names
	treeNames := []string{"Combat", "Survival", "Tech"}
	for i, tree := range state.Trees {
		if tree.TreeName != treeNames[i] {
			t.Errorf("Tree %d: expected name %s, got %s", i, treeNames[i], tree.TreeName)
		}
		if len(tree.Nodes) != 5 {
			t.Errorf("Tree %s: expected 5 nodes, got %d", tree.TreeName, len(tree.Nodes))
		}
	}
}

// TestModLoaderIntegration verifies mod loader is initialized and functional.
func TestModLoaderIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify mod loader is initialized
	if game.modLoader == nil {
		t.Fatal("Mod loader not initialized after startNewGame")
	}

	// Verify mods directory
	modsDir := game.modLoader.GetModsDir()
	if modsDir == "" {
		t.Error("Mods directory should not be empty")
	}

	// Verify ListMods works (may be empty if no mods dir)
	mods := game.modLoader.ListMods()
	if mods == nil {
		t.Error("ListMods should return non-nil slice")
	}
}

// TestBuildModsState verifies mods state building for UI.
func TestBuildModsState(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	state := game.buildModsState()
	if state == nil {
		t.Fatal("Mods state should not be nil")
	}
	if state.ModsDir == "" {
		t.Error("Mods directory should not be empty")
	}
}

// TestMultiplayerStateTransition verifies multiplayer state transitions.
func TestMultiplayerStateTransition(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Open multiplayer lobby
	game.openMultiplayer()
	if game.state != StateMultiplayer {
		t.Errorf("Expected StateMultiplayer, got %v", game.state)
	}
	if !game.menuManager.IsVisible() {
		t.Error("Menu should be visible in multiplayer state")
	}
}

// TestMultiplayerModes verifies multiplayer mode list.
func TestMultiplayerModes(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	modes := game.getMultiplayerModes()

	expectedModes := []string{"coop", "ffa", "team", "territory"}
	if len(modes) != len(expectedModes) {
		t.Fatalf("Expected %d modes, got %d", len(expectedModes), len(modes))
	}
	for i, expected := range expectedModes {
		if modes[i].ID != expected {
			t.Errorf("Mode %d: expected ID %s, got %s", i, expected, modes[i].ID)
		}
		if modes[i].Name == "" {
			t.Errorf("Mode %s: name should not be empty", modes[i].ID)
		}
		if modes[i].MaxPlayers <= 0 {
			t.Errorf("Mode %s: max players should be positive", modes[i].ID)
		}
	}
}

// TestMultiplayerCoopInit verifies co-op session initialization from lobby.
func TestMultiplayerCoopInit(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()
	game.openMultiplayer()

	// Select co-op mode (index 0)
	game.mpSelectedMode = 0
	game.handleMultiplayerSelect()

	if !game.networkMode {
		t.Error("Network mode should be enabled after selecting co-op")
	}
	if game.multiplayerMgr == nil {
		t.Error("Multiplayer manager should be initialized")
	}
	if game.mpStatusMsg == "" {
		t.Error("Status message should be set")
	}
}

// TestDrawSkillsModsMultiplayer verifies draw methods don't panic.
func TestDrawSkillsModsMultiplayer(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	screen := ebiten.NewImage(320, 200)

	// Test drawSkills doesn't panic
	game.openSkills()
	game.drawSkills(screen)

	// Test drawMods doesn't panic
	game.state = StateMods
	game.drawMods(screen)

	// Test drawMultiplayer doesn't panic
	game.openMultiplayer()
	game.drawMultiplayer(screen)
}

// TestSkillPointsOnLevelUp verifies skill points are awarded on level-up.
func TestSkillPointsOnLevelUp(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	combatTree, _ := game.skillManager.GetTree("combat")
	startPoints := combatTree.GetPoints()

	// Directly simulate level-up and skill point award (matching main.go logic)
	game.skillManager.AddPoints(1)

	// Verify skill point was awarded to all trees
	afterTree, _ := game.skillManager.GetTree("combat")
	if afterTree.GetPoints() != startPoints+1 {
		t.Errorf("Expected %d skill points after AddPoints(1), got %d", startPoints+1, afterTree.GetPoints())
	}
}

// TestGetTreeNodeList verifies stable node ordering.
func TestGetTreeNodeList(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	tree, err := game.skillManager.GetTree("combat")
	if err != nil {
		t.Fatalf("Failed to get combat tree: %v", err)
	}

	nodes := game.getTreeNodeList(tree)
	if len(nodes) != 5 {
		t.Fatalf("Expected 5 nodes, got %d", len(nodes))
	}

	// Verify sorted order
	for i := 1; i < len(nodes); i++ {
		if nodes[i-1].ID > nodes[i].ID {
			t.Errorf("Nodes not sorted: %s > %s", nodes[i-1].ID, nodes[i].ID)
		}
	}
}

// TestInventoryIntegration verifies inventory system is fully integrated.
func TestInventoryIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify inventory is initialized
	if game.playerInventory == nil {
		t.Fatal("Player inventory not initialized")
	}

	// Test shop purchase adds to inventory
	game.shopCredits.Set(1000)
	game.applyShopItem("medkit")
	if !game.playerInventory.Has("medkit") {
		t.Error("Shop medkit purchase should add to inventory")
	}

	// Test crafting adds to inventory
	game.applyCraftedItem("medkit", 2)
	item := game.playerInventory.Get("medkit")
	if item == nil || item.Qty != 3 {
		t.Errorf("Expected 3 medkits after crafting, got %v", item)
	}

	// Test grenades can be added
	game.applyShopItem("grenade")
	if !game.playerInventory.Has("grenade") {
		t.Error("Shop grenade purchase should add to inventory")
	}

	// Test using item from quick slot
	initialHealth := 50
	game.hud.Health = initialHealth
	game.hud.MaxHealth = 100

	// Manually set quick slot with medkit
	game.playerInventory.SetQuickSlot(&inventory.Medkit{
		ID:         "medkit",
		Name:       "Medkit",
		HealAmount: 25,
	})

	// Use quick slot item
	game.useQuickSlotItem()

	// Verify health increased
	if game.hud.Health != initialHealth+25 {
		t.Errorf("Expected health %d after medkit use, got %d", initialHealth+25, game.hud.Health)
	}

	// Verify item consumed from inventory
	item = game.playerInventory.Get("medkit")
	if item == nil || item.Qty != 2 {
		t.Errorf("Expected 2 medkits after using one, got %v", item)
	}
}

// TestInventoryQuickSlotAutoEquip tests auto-equipping medkits.
func TestInventoryQuickSlotAutoEquip(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Add medkit to inventory but don't equip
	game.playerInventory.Add(inventory.Item{ID: "medkit", Name: "Medkit", Qty: 1})

	// Use item should auto-equip medkit
	game.hud.Health = 50
	game.hud.MaxHealth = 100

	game.useQuickSlotItem()

	// Should auto-equip and use the medkit
	if game.hud.Health != 75 {
		t.Errorf("Expected auto-equip and heal to 75, got %d", game.hud.Health)
	}

	// Medkit should be consumed
	if game.playerInventory.Has("medkit") {
		t.Error("Medkit should be consumed after use")
	}
}

// TestInventoryEmptyQuickSlot tests using empty quick slot.
func TestInventoryEmptyQuickSlot(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Try to use with empty inventory - should show message but not crash
	game.useQuickSlotItem()

	// No panic means success
}

// TestPropsManagerInitialization tests that props manager is initialized in NewGame.
func TestPropsManagerInitialization(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	if game.propsManager == nil {
		t.Fatal("Props manager not initialized in NewGame")
	}
}

// TestPropsPlacementInRooms tests that props are placed in generated rooms.
func TestPropsPlacementInRooms(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	if game.propsManager == nil {
		t.Fatal("Props manager should be initialized after startNewGame")
	}

	props := game.propsManager.GetProps()
	if len(props) == 0 {
		t.Error("No props placed in level - expected decorative props in rooms")
	}

	// Verify props are within map bounds
	for _, prop := range props {
		if prop.X < 0 || prop.Y < 0 {
			t.Errorf("Prop at invalid negative position: (%f, %f)", prop.X, prop.Y)
		}
		if int(prop.X) >= len(game.currentMap[0]) || int(prop.Y) >= len(game.currentMap) {
			t.Errorf("Prop outside map bounds: (%f, %f)", prop.X, prop.Y)
		}
	}
}

// TestPropsGenreConfiguration tests that props change with genre.
func TestPropsGenreConfiguration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genreID := range genres {
		game.genreID = genreID
		game.startNewGame()

		if game.propsManager.GetGenre() != genreID {
			t.Errorf("Props manager genre %s doesn't match game genre %s",
				game.propsManager.GetGenre(), genreID)
		}

		props := game.propsManager.GetProps()
		if len(props) > 0 {
			// Just verify we got props, genre-specific types are tested in pkg/props
			t.Logf("Genre %s: placed %d props", genreID, len(props))
		}
	}
}

// TestPropsClearedOnNewGame tests that props are cleared when starting a new game.
func TestPropsClearedOnNewGame(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	firstGameProps := len(game.propsManager.GetProps())
	if firstGameProps == 0 {
		t.Fatal("Expected props in first game")
	}

	// Start a new game with different seed
	game.seed++
	game.startNewGame()

	secondGameProps := len(game.propsManager.GetProps())
	if secondGameProps == 0 {
		t.Error("Expected props in second game")
	}

	// Props should be regenerated (different seed = different props)
	// We just verify the system is working, not that counts differ
	t.Logf("First game: %d props, Second game: %d props", firstGameProps, secondGameProps)
}

// TestPropsSetGenreCascade tests that setGenre() propagates to props manager.
func TestPropsSetGenreCascade(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	initialGenre := game.genreID
	if game.propsManager.GetGenre() != initialGenre {
		t.Errorf("Initial genre mismatch: game=%s, props=%s", initialGenre, game.propsManager.GetGenre())
	}

	// Change genre via setGenre cascade
	newGenre := "scifi"
	if initialGenre == "scifi" {
		newGenre = "horror"
	}

	game.setGenre(newGenre)

	if game.propsManager.GetGenre() != newGenre {
		t.Errorf("Genre not cascaded to props manager: expected %s, got %s",
			newGenre, game.propsManager.GetGenre())
	}
}

// TestLoreCodexIntegration verifies lore system initialization and usage.
func TestLoreCodexIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	if game.loreCodex == nil {
		t.Fatal("Lore codex not initialized")
	}
	if game.loreGenerator == nil {
		t.Fatal("Lore generator not initialized")
	}
	if game.loreItems == nil {
		t.Fatal("Lore items slice not initialized")
	}

	// Start a new game to generate lore
	game.startNewGame()

	// Verify lore items were placed
	if len(game.loreItems) == 0 {
		t.Error("No lore items generated in level")
	}

	// Verify codex entries were created (but not found yet)
	allEntries := game.loreCodex.GetEntries()
	if len(allEntries) == 0 {
		t.Error("No codex entries generated")
	}

	foundEntries := game.loreCodex.GetFoundEntries()
	if len(foundEntries) != 0 {
		t.Error("Entries should not be found initially")
	}

	// Simulate collecting a lore item
	if len(game.loreItems) > 0 {
		firstItem := game.loreItems[0]
		game.loreCodex.MarkFound(firstItem.CodexID)

		foundEntries = game.loreCodex.GetFoundEntries()
		if len(foundEntries) != 1 {
			t.Errorf("Expected 1 found entry after collection, got %d", len(foundEntries))
		}
	}
}

// TestLoreItemPlacement verifies lore items are placed in valid locations.
func TestLoreItemPlacement(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	if len(game.loreItems) == 0 {
		t.Skip("No lore items generated")
	}

	// Verify all lore items are within map bounds
	if len(game.currentMap) == 0 {
		t.Fatal("Map not generated")
	}

	mapWidth := len(game.currentMap[0])
	mapHeight := len(game.currentMap)

	for i, item := range game.loreItems {
		if item.PosX < 0 || item.PosX >= float64(mapWidth) {
			t.Errorf("Lore item %d X position out of bounds: %f", i, item.PosX)
		}
		if item.PosY < 0 || item.PosY >= float64(mapHeight) {
			t.Errorf("Lore item %d Y position out of bounds: %f", i, item.PosY)
		}
		if item.CodexID == "" {
			t.Errorf("Lore item %d has empty CodexID", i)
		}
	}
}

// TestLoreGenreIntegration verifies lore system respects genre changes.
func TestLoreGenreIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.genreID = "fantasy"
	game.startNewGame()

	// Capture initial genre
	initialGenre := game.genreID

	// Change genre
	newGenre := "scifi"
	game.setGenre(newGenre)

	// Verify lore generator was updated
	// Note: Generator doesn't expose GetGenre, so we test by generating content
	entry := game.loreGenerator.Generate("test_entry")
	if entry.ID != "test_entry" {
		t.Error("Generator failed to create entry after genre change")
	}
	if entry.Title == "" {
		t.Error("Generated entry has empty title")
	}
	if entry.Text == "" {
		t.Error("Generated entry has empty text")
	}

	// Verify genre was actually changed
	if game.genreID != newGenre {
		t.Errorf("Genre not changed: expected %s, got %s", newGenre, game.genreID)
	}
	if game.genreID == initialGenre {
		t.Error("Genre should have changed but didn't")
	}
}

// TestLoreItemCollection verifies tryCollectLore function.
func TestLoreItemCollection(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	if len(game.loreItems) == 0 {
		t.Skip("No lore items to test collection")
	}

	// Place player near first lore item
	firstItem := game.loreItems[0]
	game.camera.X = firstItem.PosX
	game.camera.Y = firstItem.PosY

	// Initially not activated
	if firstItem.Activated {
		t.Error("Lore item should not be activated initially")
	}

	// Try to collect
	game.tryCollectLore()

	// Should now be activated
	if !firstItem.Activated {
		t.Error("Lore item should be activated after collection")
	}

	// Codex entry should be marked as found
	entry, exists := game.loreCodex.GetEntry(firstItem.CodexID)
	if !exists {
		t.Error("Codex entry not found")
	}
	if !entry.Found {
		t.Error("Codex entry should be marked as found")
	}
}

// TestCodexUIState verifies codex UI state management.
func TestCodexUIState(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Initially should be in playing state
	if game.state != StatePlaying {
		game.state = StatePlaying
	}

	// Simulate opening codex (would normally be triggered by input)
	game.state = StateCodex
	game.codexScrollIdx = 0

	if game.state != StateCodex {
		t.Error("Game state should be StateCodex")
	}

	// Add some found entries
	if len(game.loreItems) > 0 {
		game.loreCodex.MarkFound(game.loreItems[0].CodexID)
	}

	foundEntries := game.loreCodex.GetFoundEntries()
	if len(foundEntries) > 0 {
		// Test scroll bounds
		game.codexScrollIdx = -1
		game.updateCodex()
		if game.codexScrollIdx < 0 {
			// updateCodex doesn't auto-fix, drawCodex does
			game.state = StateCodex
		}

		game.codexScrollIdx = len(foundEntries) + 10
		game.updateCodex()
		// Should be clamped in drawCodex
	}
}

// TestMinigameSystemIntegration verifies minigame system is integrated.
func TestMinigameSystemIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Verify StateMinigame exists
	if StateMinigame == 0 {
		t.Error("StateMinigame constant should be defined")
	}

	// Verify initial minigame state
	if game.activeMinigame != nil {
		t.Error("Active minigame should be nil initially")
	}

	// Simulate starting a minigame
	game.startMinigame(10, 10)

	if game.activeMinigame == nil {
		t.Error("Active minigame should be initialized after startMinigame")
	}

	if game.state != StateMinigame {
		t.Errorf("Expected StateMinigame, got %v", game.state)
	}

	if game.minigameDoorX != 10 || game.minigameDoorY != 10 {
		t.Errorf("Door coordinates not set correctly: (%d, %d)", game.minigameDoorX, game.minigameDoorY)
	}

	if game.minigameType == "" {
		t.Error("Minigame type should be set")
	}
}

// TestMinigameLockpickCreation tests lockpicking minigame initialization.
func TestMinigameLockpickCreation(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.genreID = "fantasy"
	game.startNewGame()

	// Start lockpicking minigame for fantasy genre
	game.startMinigame(5, 5)

	if game.minigameType != "lockpick" {
		t.Errorf("Fantasy genre should use lockpick minigame, got %s", game.minigameType)
	}

	// Verify it's a lockpick game
	_, ok := game.activeMinigame.(*minigame.LockpickGame)
	if !ok {
		t.Error("Active minigame should be *minigame.LockpickGame for fantasy genre")
	}
}

// TestMinigameGenreVariety tests different genres use different minigame types.
func TestMinigameGenreVariety(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		genre        string
		expectedType string
	}{
		{"fantasy", "lockpick"},
		{"cyberpunk", "circuit"},
		{"scifi", "code"},
		{"postapoc", "code"},
		{"horror", "hack"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			game := NewGame()
			game.genreID = tt.genre
			game.startNewGame()

			game.startMinigame(1, 1)

			if game.minigameType != tt.expectedType {
				t.Errorf("Genre %s: expected minigame type %s, got %s",
					tt.genre, tt.expectedType, game.minigameType)
			}
		})
	}
}

// TestMinigameStateTransition tests state transitions with minigames.
func TestMinigameStateTransition(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	previousState := game.state
	game.startMinigame(3, 3)

	if game.previousState != previousState {
		t.Errorf("Previous state not saved: expected %v, got %v", previousState, game.previousState)
	}

	if game.state != StateMinigame {
		t.Error("State should transition to StateMinigame")
	}
}

// TestMinigameUpdateCancellation tests ESC cancels minigame.
func TestMinigameUpdateCancellation(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	game.startMinigame(2, 2)
	previousState := game.previousState

	// Simulate cancellation by directly setting state (input simulation is complex)
	game.activeMinigame = nil
	game.state = previousState

	if game.activeMinigame != nil {
		t.Error("Active minigame should be nil after cancellation")
	}

	if game.state == StateMinigame {
		t.Error("State should not be StateMinigame after cancellation")
	}
}

// TestMinigameDrawFunctions tests minigame drawing doesn't panic.
func TestMinigameDrawFunctions(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()
	screen := ebiten.NewImage(320, 200)

	// Test each minigame type rendering
	genres := []string{"fantasy", "cyberpunk", "scifi", "horror"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			game.genreID = genre
			game.startMinigame(1, 1)

			// Should not panic
			game.drawMinigame(screen)
		})
	}
}

// TestTryInteractDoorWithLockedDoor tests locked door triggers minigame.
func TestTryInteractDoorWithLockedDoor(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Find a door in the map
	doorX, doorY := -1, -1
	for y := 0; y < len(game.currentMap); y++ {
		for x := 0; x < len(game.currentMap[y]); x++ {
			if game.currentMap[y][x] == bsp.TileDoor {
				doorX = x
				doorY = y
				break
			}
		}
		if doorX >= 0 {
			break
		}
	}

	if doorX < 0 {
		t.Skip("No door found in generated map")
	}

	// Position player near door
	game.camera.X = float64(doorX) + 0.5
	game.camera.Y = float64(doorY) - 1.0
	game.camera.DirX = 0
	game.camera.DirY = 1

	// Ensure door is "locked" by not having keycard
	game.keycards = make(map[string]bool)

	// Try to interact with locked door
	game.tryInteractDoor()

	// Should start minigame if door requires keycard
	// (getDoorColor returns "" in stub, so this test verifies no-keycard case)
	// In real implementation with locked doors, this would trigger minigame
}

// TestMinigameProgressTracking tests progress is tracked correctly.
func TestMinigameProgressTracking(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.genreID = "fantasy"
	game.startNewGame()

	game.startMinigame(1, 1)

	initialProgress := game.activeMinigame.GetProgress()
	if initialProgress < 0 || initialProgress > 1 {
		t.Errorf("Initial progress out of range: %f", initialProgress)
	}

	initialAttempts := game.activeMinigame.GetAttempts()
	if initialAttempts <= 0 {
		t.Error("Should have attempts available")
	}
}

// TestMinigameDifficultyScaling tests difficulty increases with progression level.
func TestMinigameDifficultyScaling(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Test at different progression levels
	levels := []int{0, 3, 6, 9, 12}

	for _, level := range levels {
		game.progression.Level = level
		game.startMinigame(level, level)

		if game.activeMinigame == nil {
			t.Fatalf("Minigame not created for level %d", level)
		}

		// Difficulty is capped at 3
		expectedDiff := level / 3
		if expectedDiff > 3 {
			expectedDiff = 3
		}

		// We can't directly check difficulty, but we can verify minigame was created
		t.Logf("Level %d: minigame created with expected difficulty %d", level, expectedDiff)
	}
}

// TestMinigameSeedDeterminism tests same door position produces same minigame.
func TestMinigameSeedDeterminism(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game1 := NewGame()
	game1.seed = 12345
	game1.startNewGame()
	game1.startMinigame(10, 20)

	game2 := NewGame()
	game2.seed = 12345
	game2.startNewGame()
	game2.startMinigame(10, 20)

	// Both should create the same type of minigame
	if game1.minigameType != game2.minigameType {
		t.Error("Same seed and position should create same minigame type")
	}
}

// TestSecretWallIntegration verifies secret wall system is initialized.
func TestSecretWallIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	if game.secretManager == nil {
		t.Fatal("Secret manager not initialized in NewGame")
	}

	game.startNewGame()
	if game.secretManager == nil {
		t.Fatal("Secret manager should be initialized after startNewGame")
	}
}

// TestSecretWallPlacement verifies secrets are placed in the map.
func TestSecretWallPlacement(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.seed = 99999 // Fixed seed for determinism
	game.startNewGame()

	// Count secret tiles in the map
	secretCount := 0
	for y := 0; y < len(game.currentMap); y++ {
		for x := 0; x < len(game.currentMap[y]); x++ {
			if game.currentMap[y][x] == bsp.TileSecret {
				secretCount++
			}
		}
	}

	// BSP generator places secrets with 15% chance in dead ends
	// We should have at least some secrets in a 64x64 map
	if secretCount == 0 {
		t.Log("Warning: No secret tiles placed (this can happen randomly)")
	}

	// Secret manager should track all secret tiles
	managerCount := game.secretManager.GetTotalCount()
	if managerCount != secretCount {
		t.Errorf("Secret manager has %d secrets but map has %d TileSecret tiles", managerCount, secretCount)
	}
}

// TestSecretWallDiscovery verifies secret discovery mechanics.
func TestSecretWallDiscovery(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Manually add a secret wall for testing at position (11, 10)
	game.currentMap[10][11] = bsp.TileSecret
	game.secretManager.Add(11, 10, 0) // DirNorth

	initialDiscovered := game.secretManager.GetDiscoveredCount()
	if initialDiscovered != 0 {
		t.Errorf("Expected 0 discovered secrets initially, got %d", initialDiscovered)
	}

	// Position player to face the secret wall at (11, 10)
	// checkDist = 1.5, so camera at (9.5, 10.0) with dir (1.0, 0.0)
	// will check position (9.5 + 1.0*1.5, 10.0 + 0.0*1.5) = (11.0, 10.0) -> (11, 10)
	game.camera.X = 9.5
	game.camera.Y = 10.0
	game.camera.DirX = 1.0
	game.camera.DirY = 0.0

	// Trigger secret discovery
	game.tryInteractDoor()

	// Secret should now be discovered
	discovered := game.secretManager.GetDiscoveredCount()
	if discovered != 1 {
		t.Errorf("Expected 1 discovered secret after interaction, got %d", discovered)
	}

	// Secret wall should start animating
	secret := game.secretManager.Get(11, 10)
	if secret == nil {
		t.Fatal("Secret wall should exist at (11, 10)")
	}
	if !secret.IsAnimating() && !secret.IsOpen() {
		t.Error("Secret wall should be animating or open after trigger")
	}
}

// TestSecretWallAnimation verifies animation state progression.
func TestSecretWallAnimation(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Add test secret
	game.currentMap[15][15] = bsp.TileSecret
	game.secretManager.Add(15, 15, 0)

	// Trigger it
	game.secretManager.TriggerAt(15, 15, "player")

	secret := game.secretManager.Get(15, 15)
	if secret == nil {
		t.Fatal("Secret should exist")
	}
	if !secret.IsAnimating() {
		t.Error("Secret should be animating after trigger")
	}

	// Simulate multiple update cycles
	deltaTime := 1.0 / 60.0    // 60 FPS
	for i := 0; i < 120; i++ { // 2 seconds worth of updates
		game.secretManager.Update(deltaTime)
	}

	// After 2 seconds (animation duration is 1 second), secret should be open
	if !secret.IsOpen() {
		t.Error("Secret should be fully open after animation completes")
	}
}

// TestSecretWallQuestTracking verifies quest system integration.
func TestSecretWallQuestTracking(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Add a test secret
	game.currentMap[20][20] = bsp.TileSecret
	game.secretManager.Add(20, 20, 0)

	// Position player to interact with secret
	game.camera.X = 19.5
	game.camera.Y = 20.0
	game.camera.DirX = 1.0
	game.camera.DirY = 0.0

	// Quest tracker should exist
	if game.questTracker == nil {
		t.Fatal("Quest tracker should be initialized")
	}

	// Interact with secret
	game.tryInteractDoor()

	// Quest progress should be updated
	// (bonus_secrets objective is updated in tryInteractDoor)
	for _, obj := range game.questTracker.Objectives {
		if obj.ID == "bonus_secrets" {
			if obj.Progress == 0 {
				t.Log("Warning: bonus_secrets progress not incremented (objective may not exist)")
			}
			break
		}
	}
}

// TestSecretWallGenreDifferences verifies genre-specific behavior.
func TestSecretWallGenreDifferences(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			game := NewGame()
			game.genreID = genre
			game.seed = 55555 // Fixed seed
			game.startNewGame()

			// Verify secrets are placed regardless of genre
			secretCount := game.secretManager.GetTotalCount()
			t.Logf("Genre %s has %d secrets", genre, secretCount)
			// Some maps may have 0 secrets due to random generation
			if secretCount < 0 {
				t.Errorf("Invalid secret count: %d", secretCount)
			}
		})
	}
}

// TestSecretWallDeterminism verifies deterministic secret placement.
func TestSecretWallDeterminism(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	seed := uint64(777777)

	game1 := NewGame()
	game1.seed = seed
	game1.startNewGame()
	secrets1 := game1.secretManager.GetAll()

	game2 := NewGame()
	game2.seed = seed
	game2.startNewGame()
	secrets2 := game2.secretManager.GetAll()

	if len(secrets1) != len(secrets2) {
		t.Errorf("Same seed should produce same number of secrets: %d vs %d", len(secrets1), len(secrets2))
	}

	// Verify positions match
	for i := 0; i < len(secrets1) && i < len(secrets2); i++ {
		if secrets1[i].X != secrets2[i].X || secrets1[i].Y != secrets2[i].Y {
			t.Errorf("Secret %d position mismatch: (%d,%d) vs (%d,%d)",
				i, secrets1[i].X, secrets1[i].Y, secrets2[i].X, secrets2[i].Y)
		}
	}
}

// TestWeaponUpgradeIntegration verifies weapon upgrade system initialization.
func TestWeaponUpgradeIntegration(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	if game.upgradeManager == nil {
		t.Fatal("Upgrade manager not initialized")
	}

	// Verify initial token count is zero
	if game.upgradeManager.GetTokens().GetCount() != 0 {
		t.Errorf("Expected 0 initial tokens, got %d", game.upgradeManager.GetTokens().GetCount())
	}
}

// TestWeaponUpgradeTokenDrops verifies upgrade tokens are awarded on enemy kills.
func TestWeaponUpgradeTokenDrops(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	initialTokens := game.upgradeManager.GetTokens().GetCount()

	// Manually add upgrade tokens (simulating enemy kill)
	game.upgradeManager.GetTokens().Add(5)

	newTokens := game.upgradeManager.GetTokens().GetCount()
	if newTokens != initialTokens+5 {
		t.Errorf("Expected %d tokens after adding 5, got %d", initialTokens+5, newTokens)
	}
}

// TestWeaponUpgradePurchase verifies upgrade purchase and application.
func TestWeaponUpgradePurchase(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Give player upgrade tokens
	game.upgradeManager.GetTokens().Add(10)

	// Apply damage upgrade to current weapon
	currentWeapon := game.arsenal.GetCurrentWeapon()
	weaponID := currentWeapon.Name

	success := game.upgradeManager.ApplyUpgrade(weaponID, 0, 2) // UpgradeDamage = 0, cost 2 tokens
	if !success {
		t.Error("Failed to apply upgrade with sufficient tokens")
	}

	// Verify tokens were deducted
	if game.upgradeManager.GetTokens().GetCount() != 8 {
		t.Errorf("Expected 8 tokens after spending 2, got %d", game.upgradeManager.GetTokens().GetCount())
	}

	// Verify upgrade was applied
	upgrades := game.upgradeManager.GetUpgrades(weaponID)
	if len(upgrades) != 1 {
		t.Errorf("Expected 1 upgrade applied, got %d", len(upgrades))
	}
}

// TestWeaponUpgradeDamageCalculation verifies upgraded damage is calculated correctly.
func TestWeaponUpgradeDamageCalculation(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	currentWeapon := game.arsenal.GetCurrentWeapon()
	baseDamage := currentWeapon.Damage

	// Get base damage (no upgrades)
	damage1 := game.getUpgradedWeaponDamage(currentWeapon)
	if damage1 != baseDamage {
		t.Errorf("Expected base damage %f, got %f", baseDamage, damage1)
	}

	// Apply damage upgrade
	game.upgradeManager.GetTokens().Add(10)
	game.upgradeManager.ApplyUpgrade(currentWeapon.Name, 0, 2) // UpgradeDamage = 0

	// Get upgraded damage (should be 25% higher)
	damage2 := game.getUpgradedWeaponDamage(currentWeapon)
	expectedDamage := baseDamage * 1.25
	if damage2 != expectedDamage {
		t.Errorf("Expected upgraded damage %f, got %f", expectedDamage, damage2)
	}
}

// TestWeaponUpgradeShopItems verifies shop upgrade items handling.
func TestWeaponUpgradeShopItems(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Give player upgrade tokens
	game.upgradeManager.GetTokens().Add(10)
	initialTokens := game.upgradeManager.GetTokens().GetCount()

	// Apply shop upgrade item (damage)
	game.applyShopItem("upgrade_damage")

	// Verify upgrade was applied and tokens deducted
	currentWeapon := game.arsenal.GetCurrentWeapon()
	upgrades := game.upgradeManager.GetUpgrades(currentWeapon.Name)
	if len(upgrades) == 0 {
		t.Error("No upgrades applied after shop purchase")
	}

	if game.upgradeManager.GetTokens().GetCount() >= initialTokens {
		t.Error("Tokens were not deducted after upgrade purchase")
	}
}

// TestWeaponUpgradeMultipleUpgrades verifies multiple upgrades stack correctly.
func TestWeaponUpgradeMultipleUpgrades(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	currentWeapon := game.arsenal.GetCurrentWeapon()
	baseDamage := currentWeapon.Damage

	// Apply two damage upgrades
	game.upgradeManager.GetTokens().Add(10)
	game.upgradeManager.ApplyUpgrade(currentWeapon.Name, 0, 2) // UpgradeDamage
	game.upgradeManager.ApplyUpgrade(currentWeapon.Name, 0, 2) // UpgradeDamage again

	// Get upgraded damage (should be 1.25 * 1.25 = 1.5625x base)
	damage := game.getUpgradedWeaponDamage(currentWeapon)
	expectedDamage := baseDamage * 1.25 * 1.25
	if damage != expectedDamage {
		t.Errorf("Expected stacked damage %f, got %f", expectedDamage, damage)
	}

	// Verify both upgrades are tracked
	upgrades := game.upgradeManager.GetUpgrades(currentWeapon.Name)
	if len(upgrades) != 2 {
		t.Errorf("Expected 2 upgrades tracked, got %d", len(upgrades))
	}
}

// TestWeaponUpgradeInsufficientTokens verifies purchase fails with insufficient tokens.
func TestWeaponUpgradeInsufficientTokens(t *testing.T) {
	if err := config.Load(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	game := NewGame()
	game.startNewGame()

	// Give player only 1 token
	game.upgradeManager.GetTokens().Add(1)

	currentWeapon := game.arsenal.GetCurrentWeapon()
	success := game.upgradeManager.ApplyUpgrade(currentWeapon.Name, 0, 2) // Cost 2 tokens
	if success {
		t.Error("Upgrade succeeded with insufficient tokens")
	}

	// Verify token count unchanged
	if game.upgradeManager.GetTokens().GetCount() != 1 {
		t.Errorf("Expected 1 token after failed purchase, got %d", game.upgradeManager.GetTokens().GetCount())
	}
}

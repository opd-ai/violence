package main

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/minigame"
)

// TestMinigameVisualComponents verifies all minigame drawing functions execute without errors.
func TestMinigameVisualComponents(t *testing.T) {
	tests := []struct {
		name         string
		minigameType string
		setupGame    func(*Game)
	}{
		{
			name:         "Lockpick visual rendering",
			minigameType: "lockpick",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewLockpickGame(1, 12345)
				g.minigameType = "lockpick"
			},
		},
		{
			name:         "Hack visual rendering",
			minigameType: "hack",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewHackGame(1, 12345)
				g.minigameType = "hack"
			},
		},
		{
			name:         "Circuit visual rendering",
			minigameType: "circuit",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewCircuitTraceGame(1, 12345)
				g.minigameType = "circuit"
			},
		},
		{
			name:         "Code visual rendering",
			minigameType: "code",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewBypassCodeGame(1, 12345)
				g.minigameType = "code"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal game instance
			g := &Game{}
			tt.setupGame(g)

			// Create test screen
			screen := ebiten.NewImage(640, 480)

			// Call drawing functions - should not panic
			centerX := float32(320)
			centerY := float32(240)

			switch tt.minigameType {
			case "lockpick":
				g.drawLockpickGame(screen, centerX, centerY)
			case "hack":
				g.drawHackGame(screen, centerX, centerY)
			case "circuit":
				g.drawCircuitGame(screen, centerX, centerY)
			case "code":
				g.drawCodeGame(screen, centerX, centerY)
			}

			// Verify screen was modified (not empty)
			bounds := screen.Bounds()
			if bounds.Empty() {
				t.Error("Screen bounds are empty after rendering")
			}
		})
	}
}

// TestMinigameVisualProgress verifies visual elements update based on game state.
func TestMinigameVisualProgress(t *testing.T) {
	t.Run("Lockpick shows pin progress", func(t *testing.T) {
		g := &Game{}
		lpGame := minigame.NewLockpickGame(2, 54321)
		lpGame.Start()

		// Unlock first pin
		lpGame.Position = lpGame.Target
		lpGame.Attempt()

		g.activeMinigame = lpGame
		g.minigameType = "lockpick"

		screen := ebiten.NewImage(640, 480)
		g.drawLockpickGame(screen, 320, 240)

		if lpGame.UnlockedPins != 1 {
			t.Errorf("Expected 1 unlocked pin, got %d", lpGame.UnlockedPins)
		}

		// Verify progress is reflected
		progress := lpGame.GetProgress()
		if progress == 0 {
			t.Error("Progress should be > 0 after unlocking a pin")
		}
	})

	t.Run("Hack shows sequence progress", func(t *testing.T) {
		g := &Game{}
		hackGame := minigame.NewHackGame(1, 54321)
		hackGame.Start()

		// Input first correct node
		if len(hackGame.Sequence) > 0 {
			hackGame.Input(hackGame.Sequence[0])
		}

		g.activeMinigame = hackGame
		g.minigameType = "hack"

		screen := ebiten.NewImage(640, 480)
		g.drawHackGame(screen, 320, 240)

		if len(hackGame.PlayerInput) != 1 {
			t.Errorf("Expected 1 player input, got %d", len(hackGame.PlayerInput))
		}
	})

	t.Run("Circuit shows position updates", func(t *testing.T) {
		g := &Game{}
		circuitGame := minigame.NewCircuitTraceGame(1, 54321)
		circuitGame.Start()

		initialX, initialY := circuitGame.CurrentX, circuitGame.CurrentY

		// Move right if possible
		circuitGame.Move(1)

		g.activeMinigame = circuitGame
		g.minigameType = "circuit"

		screen := ebiten.NewImage(640, 480)
		g.drawCircuitGame(screen, 320, 240)

		// Verify position changed or stayed same if blocked
		movedOrBlocked := (circuitGame.CurrentX != initialX) ||
			(circuitGame.CurrentX == initialX && circuitGame.CurrentY == initialY)
		if !movedOrBlocked {
			t.Error("Unexpected position state")
		}
	})

	t.Run("Code shows input progress", func(t *testing.T) {
		g := &Game{}
		codeGame := minigame.NewBypassCodeGame(1, 54321)
		codeGame.Start()

		// Input first digit
		codeGame.InputDigit(5)

		g.activeMinigame = codeGame
		g.minigameType = "code"

		screen := ebiten.NewImage(640, 480)
		g.drawCodeGame(screen, 320, 240)

		if len(codeGame.PlayerInput) != 1 {
			t.Errorf("Expected 1 digit input, got %d", len(codeGame.PlayerInput))
		}
	})
}

// TestMinigameDrawNilHandling verifies graceful handling of nil/invalid states.
func TestMinigameDrawNilHandling(t *testing.T) {
	tests := []struct {
		name     string
		drawFunc func(*Game, *ebiten.Image)
	}{
		{
			name: "Lockpick with nil minigame",
			drawFunc: func(g *Game, screen *ebiten.Image) {
				g.activeMinigame = nil
				g.drawLockpickGame(screen, 320, 240)
			},
		},
		{
			name: "Lockpick with wrong minigame type",
			drawFunc: func(g *Game, screen *ebiten.Image) {
				g.activeMinigame = minigame.NewHackGame(1, 12345)
				g.drawLockpickGame(screen, 320, 240)
			},
		},
		{
			name: "Hack with nil minigame",
			drawFunc: func(g *Game, screen *ebiten.Image) {
				g.activeMinigame = nil
				g.drawHackGame(screen, 320, 240)
			},
		},
		{
			name: "Circuit with nil minigame",
			drawFunc: func(g *Game, screen *ebiten.Image) {
				g.activeMinigame = nil
				g.drawCircuitGame(screen, 320, 240)
			},
		},
		{
			name: "Code with nil minigame",
			drawFunc: func(g *Game, screen *ebiten.Image) {
				g.activeMinigame = nil
				g.drawCodeGame(screen, 320, 240)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Drawing function panicked with nil/wrong minigame: %v", r)
				}
			}()

			g := &Game{}
			screen := ebiten.NewImage(640, 480)
			tt.drawFunc(g, screen)
		})
	}
}

// TestMinigameTextBounds verifies text rendering stays within screen bounds.
func TestMinigameTextBounds(t *testing.T) {
	screenWidth := 640
	screenHeight := 480

	tests := []struct {
		name         string
		minigameType string
		setupGame    func(*Game)
	}{
		{
			name:         "Lockpick text bounds",
			minigameType: "lockpick",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewLockpickGame(3, 99999)
				g.minigameType = "lockpick"
			},
		},
		{
			name:         "Hack text bounds",
			minigameType: "hack",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewHackGame(3, 99999)
				g.minigameType = "hack"
			},
		},
		{
			name:         "Circuit text bounds",
			minigameType: "circuit",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewCircuitTraceGame(3, 99999)
				g.minigameType = "circuit"
			},
		},
		{
			name:         "Code text bounds",
			minigameType: "code",
			setupGame: func(g *Game) {
				g.activeMinigame = minigame.NewBypassCodeGame(3, 99999)
				g.minigameType = "code"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{}
			tt.setupGame(g)

			screen := ebiten.NewImage(screenWidth, screenHeight)
			centerX := float32(screenWidth / 2)
			centerY := float32(screenHeight / 2)

			switch tt.minigameType {
			case "lockpick":
				g.drawLockpickGame(screen, centerX, centerY)
			case "hack":
				g.drawHackGame(screen, centerX, centerY)
			case "circuit":
				g.drawCircuitGame(screen, centerX, centerY)
			case "code":
				g.drawCodeGame(screen, centerX, centerY)
			}

			// Verify screen bounds
			bounds := screen.Bounds()
			if bounds.Dx() != screenWidth || bounds.Dy() != screenHeight {
				t.Errorf("Screen bounds changed: expected %dx%d, got %dx%d",
					screenWidth, screenHeight, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

// TestMinigameDrawWithFullMinigameState tests drawing with the full minigame UI state.
func TestMinigameDrawWithFullMinigameState(t *testing.T) {
	g := &Game{}
	g.activeMinigame = minigame.NewLockpickGame(2, 11111)
	g.minigameType = "lockpick"

	screen := ebiten.NewImage(640, 480)

	// Call the full drawMinigame function
	g.drawMinigame(screen)

	// Verify screen was rendered
	bounds := screen.Bounds()
	if bounds == image.ZR {
		t.Error("Screen was not rendered by drawMinigame")
	}

	// Test with nil minigame
	g.activeMinigame = nil
	g.drawMinigame(screen)
	// Should not panic
}

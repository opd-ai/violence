package ui

import (
	"image/color"
)

// Example: Interactive Main Menu
//
// This example shows how to create a main menu with interactive buttons
// and smooth transitions.

// CreateInteractiveMenu demonstrates creating an interactive menu system.
func CreateInteractiveMenu(screenWidth, screenHeight int) *InteractiveSystem {
	sys := NewInteractiveSystem()

	// Calculate centered menu position
	menuX := float32(screenWidth/2 - 100)
	startY := float32(screenHeight/2 - 100)
	buttonSpacing := float32(50)

	// Define theme colors
	idleColor := color.RGBA{R: 60, G: 60, B: 70, A: 255}
	hoverColor := color.RGBA{R: 80, G: 80, B: 100, A: 255}
	pressedColor := color.RGBA{R: 50, G: 50, B: 60, A: 255}
	focusedColor := color.RGBA{R: 100, G: 100, B: 140, A: 255}
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Create menu buttons
	menuItems := []struct {
		label   string
		onClick func()
	}{
		{"New Game", func() { /* Start new game */ }},
		{"Load Game", func() { /* Load saved game */ }},
		{"Settings", func() { /* Open settings */ }},
		{"Quit", func() { /* Exit game */ }},
	}

	for i, item := range menuItems {
		btn := &Button{
			X:            menuX,
			Y:            startY + float32(i)*buttonSpacing,
			Width:        200,
			Height:       40,
			Label:        item.label,
			ColorIdle:    idleColor,
			ColorHover:   hoverColor,
			ColorPressed: pressedColor,
			ColorFocused: focusedColor,
			TextColor:    textColor,
			OnClick:      item.onClick,
		}
		sys.AddButton(btn)
	}

	return sys
}

// CreateSettingsPanel demonstrates creating a slide-out settings panel.
func CreateSettingsPanel(screenWidth, screenHeight int) (*InteractiveSystem, *Panel) {
	sys := NewInteractiveSystem()

	// Create panel that slides in from the right
	panel := &Panel{
		X:           float32(screenWidth - 300),
		Y:           50,
		Width:       280,
		Height:      float32(screenHeight - 100),
		Visible:     false,
		BgColor:     color.RGBA{R: 40, G: 40, B: 50, A: 220},
		BorderColor: color.RGBA{R: 100, G: 100, B: 120, A: 255},
	}
	sys.AddPanel(panel)

	// Add settings buttons to the panel
	btnTheme := color.RGBA{R: 70, G: 70, B: 80, A: 255}
	btnHover := color.RGBA{R: 90, G: 90, B: 110, A: 255}
	btnPressed := color.RGBA{R: 60, G: 60, B: 70, A: 255}
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	settings := []string{"Video", "Audio", "Controls", "Back"}
	for i, label := range settings {
		btn := &Button{
			X:            panel.X + 20,
			Y:            panel.Y + 50 + float32(i*50),
			Width:        240,
			Height:       40,
			Label:        label,
			ColorIdle:    btnTheme,
			ColorHover:   btnHover,
			ColorPressed: btnPressed,
			TextColor:    textColor,
			OnClick: func() {
				if label == "Back" {
					sys.HidePanel(panel)
				}
			},
		}
		sys.AddButton(btn)
	}

	return sys, panel
}

// CreateInGameHUD demonstrates creating an interactive HUD with collapsible panels.
func CreateInGameHUD(screenWidth, screenHeight int) (*InteractiveSystem, *Panel, *Panel) {
	sys := NewInteractiveSystem()

	// Create inventory panel (left side)
	invPanel := &Panel{
		X:           10,
		Y:           float32(screenHeight/2 - 150),
		Width:       200,
		Height:      300,
		Visible:     false,
		BgColor:     color.RGBA{R: 30, G: 30, B: 40, A: 200},
		BorderColor: color.RGBA{R: 80, G: 80, B: 100, A: 255},
	}
	sys.AddPanel(invPanel)

	// Create quest panel (right side)
	questPanel := &Panel{
		X:           float32(screenWidth - 210),
		Y:           10,
		Width:       200,
		Height:      250,
		Visible:     true,
		BgColor:     color.RGBA{R: 30, G: 30, B: 40, A: 180},
		BorderColor: color.RGBA{R: 80, G: 80, B: 100, A: 255},
	}
	sys.AddPanel(questPanel)

	// Add toggle buttons for panels
	btnColor := color.RGBA{R: 60, G: 60, B: 70, A: 200}
	btnHover := color.RGBA{R: 80, G: 80, B: 90, A: 220}
	btnPressed := color.RGBA{R: 50, G: 50, B: 60, A: 200}
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Inventory toggle button
	invBtn := &Button{
		X:            10,
		Y:            10,
		Width:        60,
		Height:       30,
		Label:        "INV",
		ColorIdle:    btnColor,
		ColorHover:   btnHover,
		ColorPressed: btnPressed,
		TextColor:    textColor,
		OnClick: func() {
			if invPanel.Visible {
				sys.HidePanel(invPanel)
			} else {
				sys.ShowPanel(invPanel)
			}
		},
	}
	sys.AddButton(invBtn)

	return sys, invPanel, questPanel
}

// Example usage in game loop:
//
//	type Game struct {
//		interactiveSys *ui.InteractiveSystem
//		settingsPanel  *ui.Panel
//		// ... other fields
//	}
//
//	func (g *Game) Update() error {
//		mouseX, mouseY := ebiten.CursorPosition()
//		mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
//		g.interactiveSys.Update(mouseX, mouseY, mousePressed)
//
//		// Handle keyboard navigation
//		if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
//			// Cycle focus through buttons
//		}
//		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
//			g.interactiveSys.ShowPanel(g.settingsPanel)
//		}
//
//		return nil
//	}
//
//	func (g *Game) Draw(screen *ebiten.Image) {
//		// Draw game world...
//
//		// Draw interactive UI on top
//		g.interactiveSys.Draw(screen)
//	}

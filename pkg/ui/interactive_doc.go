// Package ui provides interactive UI polish system for Violence game.
//
// The InteractiveSystem adds hover states, press feedback, smooth transitions,
// and micro-animations to all UI elements, making menus and buttons feel
// physically responsive.
//
// # Features
//
//   - Hover state detection with smooth color transitions
//   - Press feedback with button depression animation
//   - Focus indicators for keyboard navigation
//   - Smooth panel slide/fade animations
//   - Configurable easing functions (cubic, quad, elastic)
//   - Sub-pixel animation for smooth 60fps transitions
//
// # Example Usage
//
//	// Create interactive system
//	interactiveSys := ui.NewInteractiveSystem()
//
//	// Create a button with hover/press states
//	btn := &ui.Button{
//		X: 100, Y: 100, Width: 120, Height: 40,
//		Label: "Start Game",
//		ColorIdle:    color.RGBA{R: 60, G: 60, B: 70, A: 255},
//		ColorHover:   color.RGBA{R: 80, G: 80, B: 100, A: 255},
//		ColorPressed: color.RGBA{R: 50, G: 50, B: 60, A: 255},
//		ColorFocused: color.RGBA{R: 100, G: 100, B: 140, A: 255},
//		TextColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
//		OnClick: func() {
//			// Start game logic
//		},
//	}
//	interactiveSys.AddButton(btn)
//
//	// Create a collapsible panel
//	panel := &ui.Panel{
//		X: 500, Y: 50, Width: 250, Height: 300,
//		Visible:     false,
//		BgColor:     color.RGBA{R: 40, G: 40, B: 50, A: 220},
//		BorderColor: color.RGBA{R: 100, G: 100, B: 120, A: 255},
//	}
//	interactiveSys.AddPanel(panel)
//
//	// In update loop
//	mouseX, mouseY := ebiten.CursorPosition()
//	mousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
//	interactiveSys.Update(mouseX, mouseY, mousePressed)
//
//	// In draw loop
//	interactiveSys.Draw(screen)
//
// # Transition Timings
//
//   - Button state transitions: 150ms (9 frames at 60fps)
//   - Panel show/hide: 200ms (12 frames at 60fps)
//   - All transitions use smooth easing curves
//
// # Integration with Existing UI
//
// The InteractiveSystem can wrap existing UI elements:
//
//	// Wrap menu items
//	for i, item := range menuManager.GetMenuItems() {
//		btn := &ui.Button{
//			X: 100, Y: float32(100 + i*50),
//			Width: 200, Height: 40,
//			Label: item,
//			ColorIdle: theme.ButtonIdle,
//			ColorHover: theme.ButtonHover,
//			ColorPressed: theme.ButtonPressed,
//			TextColor: theme.Text,
//			OnClick: func() {
//				menuManager.SelectItem(i)
//			},
//		}
//		interactiveSys.AddButton(btn)
//	}
//
// # Performance
//
// The system is optimized for minimal overhead:
//   - No allocations in hot paths
//   - Efficient state tracking with bit flags
//   - Batched color interpolation
//   - Transition culling when not visible
//
// # Accessibility
//
//   - Keyboard focus support with visible indicators
//   - High-contrast mode compatible (color configurable)
//   - Screen reader friendly (future: ARIA labels)
//   - Consistent 150ms timing for predictable interaction
package ui

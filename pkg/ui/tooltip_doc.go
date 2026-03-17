// Package ui provides a tooltip documentation.
//
// # Tooltip System
//
// The TooltipSystem addresses the UI/UX problem: "Tooltips must never cover the element
// they describe." It provides automatic screen-edge awareness and repositioning.
//
// # Features
//
//   - **Automatic positioning**: Tooltips choose the best position (above, below, left, right)
//     based on available screen space
//   - **Screen-edge awareness**: Tooltips automatically reposition to stay within screen bounds
//   - **No target coverage**: Tooltips detect and avoid covering their target element
//   - **Hover delay**: Configurable delay before showing (default 300ms) to prevent flicker
//   - **Fade-in animation**: Smooth opacity transition when appearing
//   - **Text wrapping**: Long text automatically wraps to multiple lines
//   - **Rounded corners**: Polished visual appearance with configurable corner radius
//
// # Usage
//
//	// Create system
//	config := ui.DefaultTooltipConfig()
//	tooltipSys := ui.NewTooltipSystem(screenWidth, screenHeight, config)
//
//	// Register a tooltip for a button
//	tooltipSys.RegisterTooltip(
//		"save-btn",           // Unique ID
//		"Save your progress", // Tooltip text
//		100, 200,            // Target element X, Y
//		80, 24,              // Target element W, H
//		ui.TooltipAuto,      // Auto-select best position
//	)
//
//	// In your update loop, call OnHover when mouse enters the element
//	if mouseInButton {
//		tooltipSys.OnHover("save-btn")
//	} else {
//		tooltipSys.OnLeave("save-btn")
//	}
//	tooltipSys.Update()
//
//	// In your draw loop (after all other UI)
//	tooltipSys.Render(screen)
//
// # Integration with Interactive UI
//
// The tooltip system works alongside the InteractiveSystem. Register tooltips for
// any interactive buttons:
//
//	btn := &ui.Button{X: 100, Y: 200, Width: 80, Height: 24, Label: "Save"}
//	interactiveUI.AddButton(btn)
//	tooltipSys.RegisterTooltip("save-btn", "Save progress to disk", 100, 200, 80, 24, ui.TooltipAuto)
//
// # Position Selection
//
// When using TooltipAuto, the system evaluates available space in each direction:
//   - TooltipAbove: Best when target is near bottom of screen
//   - TooltipBelow: Best when target is near top of screen
//   - TooltipLeft: Best when target is near right edge
//   - TooltipRight: Best when target is near left edge
//
// # Configuration
//
// Customize appearance via TooltipConfig:
//
//	config := ui.TooltipConfig{
//		BackgroundColor: color.RGBA{R: 30, G: 30, B: 35, A: 240},
//		BorderColor:     color.RGBA{R: 80, G: 80, B: 90, A: 255},
//		TextColor:       color.RGBA{R: 220, G: 220, B: 220, A: 255},
//		Padding:         6,
//		BorderWidth:     1.5,
//		CornerRadius:    3.0,
//		ShowDelay:       300 * time.Millisecond,
//		FadeInDuration:  100 * time.Millisecond,
//		MaxWidth:        200,
//		Offset:          8, // Distance from target
//	}
package ui

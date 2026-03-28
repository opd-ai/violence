// Package focusring provides an animated focus indicator and keyboard navigation system
// for UI elements. It renders a glowing, pulsing ring around the currently focused
// interactive element, with genre-specific color themes and smooth transition animations.
//
// The focus ring improves UI accessibility by:
//   - Clearly indicating which element has keyboard focus
//   - Supporting Tab/Shift+Tab cycling through focusable elements
//   - Supporting Arrow key navigation in spatial layouts
//   - Providing smooth animated transitions when focus changes
//   - Adapting visual style to match the current game genre
//
// Visual Features:
//   - Animated pulsing glow effect using sine-wave intensity modulation
//   - Rounded corner rendering that adapts to element shape
//   - Multi-layer ring with inner bright stroke and outer soft glow
//   - Smooth position interpolation when focus moves between elements
//   - Genre-specific color themes (fantasy=gold, scifi=cyan, etc.)
//
// Integration:
//   - Register focusable UI elements via AddFocusable()
//   - Call Update() each frame to process keyboard input and advance animations
//   - Call Draw() after other UI to render focus indicator on top
//   - System is registered in main.go and active by default
//
// Example usage:
//
//	frs := focusring.NewSystem()
//	frs.AddFocusable(&focusring.FocusableElement{
//	    ID:     "start_button",
//	    X:      100, Y: 200,
//	    Width:  200, Height: 40,
//	    TabIndex: 0,
//	})
//	frs.SetGenre("fantasy")
//
//	// In game loop:
//	frs.Update()
//	frs.Draw(screen)
package focusring

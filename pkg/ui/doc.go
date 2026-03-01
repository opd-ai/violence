// Package ui provides comprehensive user interface components for the VIOLENCE game,
// including HUD rendering, menu systems, chat overlays, and multiplayer UI elements.
//
// # Architecture
//
// The ui package is organized into several core components:
//
//   - HUD: Heads-up display showing health, armor, ammo, and weapon status
//   - MenuManager: Navigation system for game menus (main, pause, settings, etc.)
//   - ChatOverlay: Thread-safe in-game chat system with message history
//   - LoadingScreen: Procedurally generated loading screen with seed display
//   - Theme: Genre-specific color palettes for UI elements
//
// # Component Hierarchy
//
// The UI system follows a layered rendering approach:
//
//  1. Game world (rendered by other systems)
//  2. HUD overlay (health, ammo, messages)
//  3. Chat overlay (when visible)
//  4. Menu screens (when paused or in menu mode)
//  5. Loading screen (during level generation)
//
// # Thread Safety
//
// All UI components are designed for safe concurrent access:
//
//   - ChatOverlay uses sync.Mutex for all state modifications
//   - Theme changes should be coordinated through game state management
//   - Drawing methods read from synchronized state copies
//
// # Usage Example
//
//	// Create HUD
//	hud := ui.NewHUD(100, 50, 200, 1, "Pistol")
//	hud.Draw(screen, 800, 600)
//
//	// Create chat overlay
//	chat := ui.NewChatOverlay(10, 10, 400, 300)
//	chat.AddMessage("Player", "Hello!", time.Now().Unix())
//	chat.Show()
//	chat.Draw(screen)
//
//	// Create menu
//	menu := ui.NewMenuManager()
//	menu.Show(ui.MenuTypeMain)
//	menu.Draw(screen, 800, 600)
//
// # Integration Points
//
// The ui package integrates with:
//
//   - pkg/config: For user settings (video, audio, controls)
//   - pkg/input: For key binding configuration
//   - github.com/hajimehoshi/ebiten/v2: For rendering primitives
//   - golang.org/x/image/font/basicfont: For text rendering
//
// # Design Patterns
//
// The package follows these patterns:
//
//   - State encapsulation: All mutable state is private with accessor methods
//   - Synchronization: Mutex protection for concurrent access patterns
//   - Separation of concerns: Rendering logic separated from state management
//   - Deterministic rendering: UI responds to state, not time-based updates
package ui

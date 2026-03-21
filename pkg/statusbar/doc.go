// Package statusbar provides a compact HUD element for displaying active status effects.
//
// The status bar renders small icons for each active buff or debuff affecting the player,
// with radial cooldown indicators showing remaining duration. Icons are genre-colored and
// auto-hide when no effects are active to minimize screen real estate usage.
//
// # Architecture
//
// The package follows the ECS pattern:
//   - Component: StatusBarComponent stores display state and icon cache
//   - System: StatusBarSystem updates icons and renders to screen
//
// # Visual Design
//
// Each status effect icon is a 16x16 pixel procedurally generated symbol:
//   - Damage effects (fire, poison, bleed): downward arrows with effect color
//   - Healing effects: upward arrows with green tint
//   - Buff effects: upward chevrons with gold/effect color
//   - Debuff effects: downward chevrons with purple/effect color
//   - Stun effects: spiral/star symbol
//   - Slow effects: horizontal lines
//
// The radial cooldown uses a pie-chart style fill that decreases as duration expires.
// A subtle pulsing animation draws attention to expiring effects (< 3 seconds remaining).
//
// # Integration
//
// The system reads from the status.StatusComponent on the player entity and renders
// icons at a configurable screen position (default: below health bar, left side).
//
// Example usage:
//
//	statusBar := statusbar.NewSystem("fantasy")
//	world.AddSystem(statusBar)
//
//	// In render loop:
//	statusBar.Render(screen, world, playerEntity)
//
// # Genre Support
//
// All 5 genres are supported with appropriate color schemes:
//   - Fantasy: warm gold buffs, purple debuffs
//   - Sci-Fi: cyan buffs, orange debuffs
//   - Horror: sickly green buffs, blood red debuffs
//   - Cyberpunk: neon pink buffs, toxic green debuffs
//   - Post-Apocalyptic: rust orange buffs, sickly yellow debuffs
package statusbar

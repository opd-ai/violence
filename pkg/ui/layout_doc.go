// Package ui provides a spatial reservation system to prevent UI element overlap.
//
// The LayoutManager implements intelligent spatial reservation to solve the OVERLAPPING UI ELEMENTS problem.
// It prevents damage numbers, health bars, tooltips, and other UI elements from rendering on top of each other,
// dramatically improving visual clarity and information accessibility.
//
// # Core Features
//
//   - **Priority-based overlap resolution**: Critical elements (player health, active threats) take precedence
//     over ambient information (distant entity status, passive effects)
//   - **Damage number stacking**: Multiple damage numbers at the same location automatically stack vertically
//   - **Alternative positioning**: Lower-priority elements automatically reposition when conflicts occur
//   - **Screen bounds clamping**: All elements stay within visible screen area
//   - **Frame-based reset**: Layout state clears each frame for dynamic repositioning
//
// # Priority Levels
//
// The system uses four priority tiers:
//
//  1. PriorityAmbient: Background info, distant entity status (lowest priority)
//  2. PrioritySecondary: Mid-range entity info
//  3. PriorityImportant: Nearby entities, tooltips
//  4. PriorityCritical: Player vitals, active threats (highest priority)
//
// Higher-priority elements can displace lower-priority elements. Fixed elements (canMove=false) never move.
//
// # Usage Pattern
//
// At the start of each render frame:
//
//	layoutMgr.Clear() // Reset all reservations
//
// For each UI element to render:
//
//	x, y, visible := layoutMgr.Reserve(id, x, y, width, height, priority, canMove)
//	if visible {
//	    // Render at adjusted x, y position
//	}
//
// For damage numbers (automatic stacking):
//
//	x, y, visible := layoutMgr.ReserveDamageNumber(id, x, y, width, height, priority)
//	if visible {
//	    // Render at adjusted y position (stacked above others)
//	}
//
// # Integration
//
// The LayoutManager is integrated into:
//
// - damagenumber.System: RenderWithLayout method for stacking damage numbers
// - healthbar.System: RenderHealthBarsWithLayout method for distance-based priority
// - main.go: Instantiated in Game struct, cleared each frame before UI render
//
// # Performance
//
// The system uses O(n) conflict detection per element, where n is the number of currently visible UI elements.
// This is acceptable for typical UI counts (10-50 elements). For extreme scenarios (100+ simultaneous elements),
// lower-priority elements are suppressed rather than repositioned, maintaining frame rate.
//
// # Example Scenario
//
// Player attacks 5 enemies clustered together:
//
// - 5 health bars at similar screen positions
// - 10 damage numbers from rapid attacks
// - 3 status effect icons per enemy
//
// Without LayoutManager: 25+ overlapping elements create visual clutter.
// With LayoutManager: Health bars offset horizontally, damage numbers stack vertically,
// distant enemies show minimal UI - clean, readable display.
package ui

// Package proximityui provides distance-based UI detail levels to reduce visual clutter.
//
// The proximity UI system controls in-world indicator visibility based on entity
// distance from the camera. This addresses the "cluttered game view" problem where
// showing full detail (health bars, names, status icons) on every visible entity
// creates visual noise that obscures gameplay.
//
// # Detail Levels
//
// The system defines four detail levels that entities can display:
//
//   - DetailFull: All indicators visible (health bar, name, status icons, faction badge)
//   - DetailModerate: Health bar and name only
//   - DetailMinimal: Health bar only (when damaged)
//   - DetailNone: No in-world UI indicators
//
// # Distance Thresholds
//
// Default thresholds (configurable per genre):
//
//   - Adjacent (0-3 units): Full detail
//   - Near (3-8 units): Moderate detail
//   - Mid (8-15 units): Minimal detail
//   - Far (15+ units): No detail
//
// # Priority Overrides
//
// Certain entities always display at higher detail levels regardless of distance:
//
//   - Targeted entity: Always DetailFull
//   - Bosses: Always at least DetailModerate
//   - Players (in multiplayer): Always at least DetailMinimal
//   - Quest NPCs: Always at least DetailModerate
//
// # Visual Transitions
//
// Detail level changes use smooth fade transitions (not abrupt pop-in/out) to avoid
// jarring visual changes as entities move relative to the camera.
//
// # Integration
//
// The system provides detail level queries that healthbar, entitylabel, and other
// in-world UI systems can use to determine what to render:
//
//	level := proximitySystem.GetDetailLevel(entity, cameraX, cameraY)
//	if level >= proximityui.DetailModerate {
//	    renderHealthBar(entity)
//	}
//	if level >= proximityui.DetailFull {
//	    renderStatusIcons(entity)
//	}
//
// # Genre Adaptation
//
// Distance thresholds adapt to genre. Horror games have shorter visibility ranges
// to maintain tension. Sci-fi games have longer ranges due to open environments.
package proximityui

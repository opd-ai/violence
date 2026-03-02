// Package stats provides character stat allocation and attribute management.
//
// The stats package implements a traditional RPG stat allocation system where players
// distribute points among five core attributes: Strength, Dexterity, Intelligence,
// Vitality, and Luck. Each attribute affects specific gameplay mechanics:
//
// - Strength: Increases melee damage (2% per point above base)
// - Dexterity: Increases accuracy (1.5% per point), dodge chance (0.3% per point), and attack speed (1% per point)
// - Intelligence: Increases skill/magic damage (2.5% per point above base)
// - Vitality: Increases max health (5% per point above base)
// - Luck: Increases critical hit chance (0.5% per point) and loot quality (1% per point)
//
// The system integrates with the ECS architecture through StatAllocationComponent and System.
// The System applies stat bonuses each frame and provides helper methods for combat calculations.
//
// Players receive stat points on level up (3 points per level by default) and can allocate
// them freely. Stats can be reset to refund all allocated points.
package stats

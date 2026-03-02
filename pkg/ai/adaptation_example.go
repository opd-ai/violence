// Example: Using the Adaptive AI System
//
// The AdaptiveAISystem learns from player behavior and adjusts enemy AI accordingly.
// This creates dynamic difficulty that rewards tactical variety.
//
// SETUP:
//
// 1. Create the system (already done in main.go):
//    adaptiveAISystem := ai.NewAdaptiveAISystem("fantasy")
//
// 2. Register it with the world (already done in main.go):
//    world.AddSystem(adaptiveAISystem)
//
// 3. Add a PlayerProfileComponent to the player entity:
//    profile := ai.NewPlayerBehaviorProfile()
//    world.AddComponent(playerEntity, &ai.PlayerProfileComponent{Profile: profile})
//
// 4. Ensure enemies have EnemyRoleComponent (already done by role_system.go)
//
// HOW IT WORKS:
//
// - Every 2 seconds, the system observes the player:
//   * Records engagement range (distance to nearest enemy)
//   * Infers current tactic (melee rush, ranged kiting, cover-based, etc.)
//   * Updates the player's behavior profile
//
// - Every 10 seconds, the system adapts enemies:
//   * Analyzes the player's dominant tactic
//   * Computes counter-strategies
//   * Applies adaptations to all enemy roles
//
// PLAYER TACTICS DETECTED:
//
// - TacticRushMelee: Close-range combat (< 3 units)
// - TacticKiteRanged: Long-range combat (> 12 units)
// - TacticCoverBased: Mid-range with high cover usage
// - TacticStealthy: Low visibility, cautious approach
// - TacticExplosives: Heavy use of explosive weapons
// - TacticHitAndRun: Quick engagements followed by retreat
//
// AI COUNTER-ADAPTATIONS:
//
// vs Melee Rusher:
//   - Enemies spread out (harder to hit multiple)
//   - Increase preferred range (kite away)
//   - Higher dodge frequency
//
// vs Ranged Kiter:
//   - Enemies pursue more aggressively
//   - Focus fire on player
//   - Increased flanking
//   - More cover usage
//
// vs Cover User:
//   - Enemies flank more
//   - Spread to surround
//   - Increased alertness
//
// vs Stealth Player:
//   - Massively increased alertness
//   - Tighter formation (less spread)
//   - Focus fire when detected
//
// vs Explosives:
//   - Maximum spread formation
//   - Higher dodge frequency
//   - Adjusted engagement range
//
// vs Hit-and-Run:
//   - Persistent pursuit
//   - Lower retreat threshold
//   - Increased alertness
//
// PERFORMANCE:
//
// - Observes every 2 seconds (minimal overhead)
// - Adapts every 10 seconds (batch operation)
// - Lightweight profile updates (exponential moving average)
// - No frame-by-frame overhead
//
// The system creates emergent difficulty: players who rely on one tactic
// will find enemies adapting to counter it, encouraging tactical variety.

package ai

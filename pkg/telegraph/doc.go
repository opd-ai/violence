// Package telegraph provides visual attack telegraphing for enemies.
//
// The telegraph system displays clear visual indicators before enemy attacks,
// giving players reaction time and improving combat readability. Each attack
// type has a distinct visual style that matches the genre aesthetic.
//
// # Attack Types
//
// The system supports four attack telegraph types:
//
//   - melee: Sweeping arc indicator for close-range attacks
//   - ranged: Directional arrow for projectile attacks
//   - aoe: Expanding circle for area-of-effect attacks
//   - charge: Motion lines for charging/dash attacks
//
// # Usage
//
// Create a telegraph system during game initialization:
//
//	telegraphSystem := telegraph.NewSystem("fantasy", seed)
//	world.AddSystem(telegraphSystem)
//
// Add telegraph components to enemy entities:
//
//	telegraphComp := &telegraph.Component{
//	    Active:        false,
//	    TelegraphTime: 1.0,
//	    AttackType:    "melee",
//	}
//	world.AddComponent(enemyEntity, telegraphComp)
//
// Start a telegraph before an enemy attacks:
//
//	telegraphSystem.StartTelegraph(world, enemyEntity, "ranged", 0.8)
//
// The system will update the charge progress each frame and deactivate
// the telegraph when complete, signaling the AI to execute the attack.
//
// # Rendering
//
// Call Render during the game's draw phase to display active telegraphs:
//
//	telegraphSystem.Render(screen, world, cameraX, cameraY)
//
// Telegraph indicators are composited onto the screen with additive blending
// for a glowing effect. Colors are automatically chosen based on genre and
// attack type.
//
// # Genre Customization
//
// The system applies genre-specific color schemes:
//
//   - fantasy: Warm golds for melee, blues for ranged, reds for AoE
//   - scifi: Cyans and magentas with harder edges
//   - horror: Muted reds and greens for dread
//   - cyberpunk: Neon pinks, cyans, and purples
//
// # Integration with AI
//
// The telegraph system integrates with behavior trees in pkg/ai:
//
//	tree := ai.NewTelegraphBehaviorTree()
//	ctx := &ai.TelegraphAttackContext{
//	    Context:         baseContext,
//	    World:           world,
//	    TelegraphSystem: telegraphSystem,
//	}
//	tree.Tick(agent, ctx.Context)
//
// Enemies will automatically use telegraphs before attacking when possible.
package telegraph

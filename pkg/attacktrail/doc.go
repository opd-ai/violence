// Package attacktrail provides visual weapon attack trail rendering.
//
// # Overview
//
// The attacktrail package creates dynamic visual trails that follow weapon attacks,
// providing visual feedback for slashes, thrusts, smashes, and other attack types.
// Each trail is procedurally rendered with genre-appropriate colors, smooth fading,
// and weapon-specific animation patterns.
//
// # Trail Types
//
// The system supports six trail types:
//
//   - TrailSlash: Curved arc for sword swings and axe chops
//   - TrailThrust: Linear piercing for spears and rapiers
//   - TrailCleave: Wide sweeping arc for greatswords and scythes
//   - TrailSmash: Radial impact burst for hammers and maces
//   - TrailSpin: Full-circle rotation for staffs and dual blades
//   - TrailProjectile: Motion streak for thrown/fired projectiles
//
// # Integration
//
// Add a TrailComponent to entities that perform attacks:
//
//	trailComp := attacktrail.NewTrailComponent(3) // Max 3 trails
//	entity.Components = append(entity.Components, trailComp)
//
// Create trails when attacks execute:
//
//	trail := attacktrail.CreateSlashTrail(x, y, angle, range, arc, width, color)
//	trailComp.AddTrail(trail)
//
// Or use the helper for automatic weapon type detection:
//
//	attacktrail.AttachTrailToAttack(entity, x, y, dirX, dirY, range, "sword", rng, colorFunc)
//
// Register the system with the ECS world:
//
//	trailSystem := attacktrail.NewSystem("fantasy")
//	world.AddSystem(trailSystem)
//
// Render trails each frame:
//
//	trailSystem.Render(screen, entities, cameraX, cameraY)
//
// # Performance
//
// Trails use vector rendering with automatic fade-out and lifecycle management.
// Old trails are automatically removed when expired. Trail components limit
// the maximum number of simultaneous trails per entity to prevent overdraw.
//
// # Genre Styling
//
// Trail colors adapt to the game's genre:
//
//   - Fantasy: Silver-blue steel trails
//   - Sci-fi: Plasma blue, laser red, energy green
//   - Horror: Dark crimson blood trails
//   - Cyberpunk: Neon magenta, cyan, yellow
//
// Use GetWeaponTrailColor() to retrieve genre-appropriate colors.
package attacktrail

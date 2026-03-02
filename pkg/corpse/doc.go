// Package corpse provides persistent visual corpse rendering for dead entities.
//
// The corpse system creates genre-specific death visuals that fade over time,
// supporting multiple death types (burn, freeze, electric, acid, explosion, slash,
// crush, disintegrate) with procedural generation and LRU caching.
//
// Usage:
//
//	sys := corpse.NewSystem(200, "fantasy", seed)
//	corpses := make([]corpse.Corpse, 0)
//
//	// When an entity dies:
//	sys.SpawnCorpse(&corpses, x, y, seed, "enemy", "humanoid",
//	    corpse.DeathBurn, 64, hasLoot)
//
//	// Each frame:
//	sys.UpdateCorpses(&corpses, deltaTime)
//	sys.RenderAllCorpses(screen, corpses, cameraX, cameraY)
//
// Death types produce visually distinct corpses:
//   - DeathNormal: standard fallen body with blood pools
//   - DeathBurn: charred remains with embers
//   - DeathFreeze: icy corpse with frost crystals
//   - DeathElectric: scorched body with lightning marks
//   - DeathAcid: dissolved, melted remains
//   - DeathExplosion: scattered body parts with gore
//   - DeathSlash: deep cuts with blood trails
//   - DeathCrush: flattened corpse with wide blood splatter
//   - DeathDisintegrate: ash and dust particles
//
// Corpses fade over time based on MaxAge and are automatically removed when
// opacity drops below 5%. The system supports genre-specific colors for blood
// and corpse materials.
package corpse

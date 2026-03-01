# Implementation Gaps

## v1.0 — Core Engine + Playable Single-Player

### Asset Pipeline — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Procedural audio synthesis fully implemented in `pkg/audio`.
- **Implementation**: Runtime synthesis for music (genre-specific scales, tempos, harmonics) and SFX (gunshot, footstep, door, explosion, pickup, pain, reload) using deterministic algorithms. All audio generated via WAV synthesis with no embedded files.
- **Files**: `pkg/audio/audio.go`, `pkg/audio/sfx.go`, `pkg/audio/ambient.go`

### ECS Component Query API — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Bitmask-based archetype query system implemented.
- **Implementation**: `World.QueryWithBitmask()` with `EntityIterator`, `SetArchetype()`, component ID constants (0-63 bits), and archetype management functions.
- **Files**: `pkg/engine/query.go`, `pkg/engine/engine.go`

### Tile Grid Format — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Shared `TileMap` structure implemented in `pkg/level`.
- **Implementation**: `TileMap` with `[][]TileType` grid, `Get()`, `Set()`, `IsWalkable()`, `InBounds()` methods. Tile constants: `TileEmpty`, `TileWall`, `TileDoor`, `TileSecret`.
- **Files**: `pkg/level/tilemap.go`, `pkg/level/tilemap_test.go`

### Player Entity Definition — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Canonical player entity schema defined with factory function.
- **Implementation**: `World.NewPlayerEntity()` creates entity with `Position`, `Health`, `Armor`, `Inventory`, `Camera`, `Input` components. `IsPlayer()` validation method.
- **Files**: `pkg/engine/player.go`, `pkg/engine/player_test.go`

### Genre Audio and Visual Assets for Fantasy — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Genre-specific procedural generation parameters implemented for all five genres.
- **Implementation**: Audio synthesis with genre-specific scales (natural minor for fantasy), tempos (100 BPM), base notes (G3), and SFX parameters. Genre ambient soundscapes (dungeon echo, dripping water).
- **Files**: `pkg/audio/audio.go` (lines 420-441), `pkg/audio/ambient.go` (dungeon echo generation)

### Save Data Schema — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — JSON schema defined with all player state fields.
- **Implementation**: `SaveState` struct with `LevelSeed`, `PlayerPosition`, `Health`, `Armor`, `Inventory`, `DiscoveredTiles`, `CurrentObjective`, `CameraDirection` JSON fields.
- **Files**: `pkg/save/schema.go`, `pkg/save/schema_test.go`

### Gamepad Mapping Specification — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Default gamepad button mappings implemented and documented.
- **Implementation**: Button 0 (A/Cross) = Fire, Button 1 (B/Circle) = Interact, Button 2 (X/Square) = Automap, Button 4 (L1/LB) = NextWeapon, Button 5 (R1/RB) = PrevWeapon, Button 7 (Start) = Pause. Axis support for analog sticks.
- **Files**: `pkg/input/input.go` (lines 45-51, gamepad bindings)

### Test Infrastructure — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Comprehensive test infrastructure established across all packages.
- **Implementation**: Table-driven tests using standard `testing` package. Mock abstractions in `pkg/testutil` for Ebitengine dependencies. Package-specific test files with >40% coverage target. Integration tests for complex workflows.
- **Files**: `pkg/testutil/mocks.go`, `*_test.go` files throughout codebase

---

## v2.0 — Core Systems: Weapons, FPS AI, Keycards, All 5 Genres

### Procedural Weapon Sprite Generation
- **Gap**: No algorithm defined for generating weapon sprites at runtime; ROADMAP requires all visuals procedurally generated but no synthesis approach exists.
- **Impact**: Weapon visuals cannot be rendered without either breaking procedural generation policy or implementing sprite synthesis.
- **Resolution needed**: Define sprite synthesis approach using geometric primitives, noise functions, or parametric curves to generate weapon raise/lower/fire/reload frames deterministically from seed.

### Enemy Sprite Generation
- **Gap**: No algorithm defined for generating enemy visuals procedurally; requires definition of body part composition and animation frames.
- **Impact**: Enemies cannot be rendered; AI behavior exists but has no visual representation.
- **Resolution needed**: Define body part decomposition (head, torso, limbs) with parametric shapes; animation frames generated via joint angle interpolation; color palette per genre.

### Pathfinding Algorithm
- **Gap**: AI chase/patrol behaviors require pathfinding against BSP tile grid; no pathfinding algorithm specified.
- **Impact**: Enemies cannot navigate around obstacles to reach player; chase state will fail on non-trivial maps.
- **Resolution needed**: Implement A* pathfinding on tile grid; define tile traversability (walls blocked, doors conditionally blocked); cache paths for performance.

### Cover Detection Algorithm
- **Gap**: AI take-cover behavior requires identifying cover tiles; no algorithm to classify tiles as providing cover.
- **Impact**: Take-cover AI state cannot function; enemies will not seek defensive positions.
- **Resolution needed**: Define cover tile criteria (adjacent to wall, not visible from threat direction); implement cover scoring function for AI decision-making.

### Projectile Collision Broadphase
- **Gap**: High projectile counts may need spatial partitioning for efficient collision detection; current approach is O(n) entity iteration.
- **Impact**: Performance degradation with many simultaneous projectiles (rocket explosions, plasma spam).
- **Resolution needed**: Implement spatial hash or grid-based broadphase; evaluate necessity based on expected projectile density.

---

## v3.0 — Visual Polish: Textures, Lighting, Particles, Indoor Weather

### Sector Light Map Implementation — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Full sector-based lighting system implemented.
- **Implementation**: `SectorLightMap` with per-tile light storage, `AddLight()` with inverse-square falloff, `Calculate()` precomputation, genre-specific ambient levels (fantasy=0.3, horror=0.15, etc.).
- **Files**: `pkg/lighting/sector.go`, `pkg/lighting/sector_test.go`

### Flashlight Cone Implementation — ✓ RESOLVED (2026-03-01)
- **Status**: COMPLETE — Cone light (flashlight/torch/headlamp) fully implemented.
- **Implementation**: `AddFlashlight(x, y, dirX, dirY, coneAngle, radius, intensity)` with dot-product angle test, quadratic distance attenuation, angular falloff. Genre skinning via `SetGenre()`.
- **Files**: `pkg/lighting/sector.go` (lines 45-72, 186-238), `pkg/lighting/flashlight.go`

### Wall Texture Coordinate Calculation
- **Gap**: `RayHit` struct lacks texture coordinate (`TextureX`) for wall sampling; renderer cannot determine where on the wall texture to sample.
- **Impact**: Wall textures cannot be applied; only palette colors render.
- **Resolution needed**: Extend `RayHit` with `TextureX float64` computed as fractional part of exact wall hit position.

### Weather Emitter Genre Configurations
- **Gap**: `ParticleSystem` exists with spawn/update/cull lifecycle, but no `WeatherEmitter` or genre-specific particle configurations exist.
- **Impact**: Indoor weather atmosphere (drips, smoke, sparks, etc.) does not appear in any genre.
- **Resolution needed**: Create `WeatherEmitter` type with per-genre spawn configurations (rate, velocity, color, lifetime, spawn positions).

### Cyberpunk Neon Pulse Animated Texture
- **Gap**: Animated textures exist for fantasy, scifi, and horror but no "neon_pulse" pattern for cyberpunk is implemented.
- **Impact**: Cyberpunk walls lack the animated neon glow effect that distinguishes the genre.
- **Resolution needed**: Implement `generateNeonPulseFrame()` in `pkg/texture/animated.go` with magenta/cyan color cycling.

### Post-Apocalyptic Radiation Glow Animated Texture
- **Gap**: No "radiation_glow" animated texture pattern for postapoc genre.
- **Impact**: Postapoc lacks visual indicator for radiation hazard areas.
- **Resolution needed**: Implement `generateRadiationGlowFrame()` with green pulsing glow effect.

### Horror Static Burst Effect
- **Gap**: Post-processor has all standard effects but no static burst (brief full-screen noise) for horror genre.
- **Impact**: Horror atmosphere lacks the unsettling random static that contributes to psychological tension.
- **Resolution needed**: Add `ApplyStaticBurst()` to post-processor with configurable probability and duration; add to horror preset.

### Postapoc Film Scratch Effect
- **Gap**: No film scratch overlay effect exists for postapoc genre's "worn film" aesthetic.
- **Impact**: Postapoc lacks the scratched/damaged film look documented in roadmap.
- **Resolution needed**: Add `ApplyFilmScratches()` to post-processor with configurable scratch density and opacity.

### BSP-to-Reverb Integration
- **Gap**: `ReverbCalculator` accepts width/height but no integration exists to extract room dimensions from BSP level data.
- **Impact**: Reverb parameters are static; rooms of different sizes all have identical reverb.
- **Resolution needed**: Implement `SetRoomFromBSP(room *bsp.Room)` that extracts bounds and calls `SetRoomSize()`.

---

## v4.0 — Gameplay Expansion: Secrets, Upgrades, Squad AI, Storytelling

### Squad AI Formation Algorithm
- **Gap**: Squad companion AI requires formation positioning (line, wedge, column) but no algorithm is defined for calculating member offsets from leader position.
- **Impact**: Squad members will cluster at identical positions without formation logic.
- **Resolution needed**: Define formation shapes as offset arrays relative to leader facing direction; implement `GetFormationOffset(memberIndex, formationType, leaderDir)`.

### Procedural Text Generation Grammar
- **Gap**: Lore generation requires template grammar to produce coherent narrative text, but no grammar rules or word banks are defined.
- **Impact**: Generated lore text may be repetitive, incoherent, or lacking genre-appropriate vocabulary.
- **Resolution needed**: Design Markov chain or template-based grammar with genre-specific word banks; define sentence structure templates for notes, logs, and graffiti.

### Credit Economy Balance
- **Gap**: Initial credit values (kill=10, secret=50, objective=100) are placeholder estimates without playtesting validation.
- **Impact**: Progression curve may feel too fast or too slow; shop items may be trivially affordable or unreachable.
- **Resolution needed**: Playtest economy across multiple runs; adjust credit rewards and item prices to achieve ~3 shop purchases per level average.

---

## v5.0+ — Multiplayer, Social Features, Production Polish

### Key Exchange Protocol
- **Gap**: No ECDH key exchange implementation exists for establishing shared chat encryption keys between clients.
- **Impact**: Chat encryption relies on pre-shared keys; no secure mechanism for strangers to establish encrypted communication.
- **Resolution needed**: Implement ECDH key exchange during session join; derive AES key from shared secret.

### Mobile Input Mapping
- **Gap**: Touch controls for mobile builds (iOS/Android via gomobile) are undefined.
- **Impact**: Mobile builds will have no input mechanism; players cannot move, aim, or interact.
- **Resolution needed**: Design virtual joystick overlay for movement, touch-to-look for aiming, tap buttons for fire/interact/reload; document touch control layout.

### Federation Hub Hosting
- **Gap**: No specification for who hosts the federation hub server or how self-hosting works.
- **Impact**: Cross-server matchmaking cannot function without a central hub; single point of failure for federation.
- **Resolution needed**: Define hub protocol for self-hosting; consider distributed hash table approach for decentralized discovery.

### Mod Sandboxing
- **Gap**: Go plugins have full runtime access with no sandboxing mechanism to limit mod capabilities.
- **Impact**: Malicious mods could access filesystem, network, or compromise player security.
- **Resolution needed**: Evaluate WASM-based mod runtime as alternative to Go plugins; if Go plugins retained, document security risks and require mod signing.

### Profanity Word List
- **Gap**: No word list defined for the profanity filter; need localized lists for multiple languages.
- **Impact**: Profanity filter cannot function without content; players expecting filtered chat will see unfiltered content.
- **Resolution needed**: Compile word lists for supported languages (English, Spanish, German, French, Portuguese); implement list loading from config.

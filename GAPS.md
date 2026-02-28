# Implementation Gaps

## v1.0 — Core Engine + Playable Single-Player

### Asset Pipeline
- **Gap**: No procedural audio or texture generation strategy is defined; the roadmap specifies "zero external assets" and all assets must be procedurally generated at runtime, but no synthesis approach for audio is implemented.
- **Impact**: Audio engine (`pkg/audio`) cannot produce sound without a runtime synthesis approach. Texture rendering requires a procedural pixel generator. No pre-rendered, embedded, or bundled audio files (e.g., `.wav`, `.ogg`, `.mp3`) or image files (e.g., `.png`, `.jpg`) are permitted.
- **Resolution needed**: Implement runtime audio synthesis (e.g., generating tones, noise, and waveforms programmatically via deterministic algorithms seeded by `pkg/rng`). All audio and visual assets must be procedurally generated at runtime — embedding or bundling pre-made asset files is not an option.

### ECS Component Query API
- **Gap**: The roadmap specifies "Component registration and query API" but the current `pkg/engine` stub has no component query mechanism (e.g., querying all entities with a specific set of components).
- **Impact**: Systems cannot efficiently iterate over entities matching their required component signature, which is fundamental to ECS operation.
- **Resolution needed**: Define the query API — options include bitmask-based archetype queries, iterator-based component filters, or reflection-based type matching.

### Tile Grid Format
- **Gap**: The raycaster requires a tile grid to ray-march against, and the BSP generator produces rooms/corridors, but no shared tile grid data structure is defined between `pkg/bsp` and `pkg/raycaster`.
- **Impact**: The two core systems cannot integrate without an agreed-upon 2D grid format representing walls, floors, doors, and empty space.
- **Resolution needed**: Define a shared `TileMap` type (e.g., `[][]uint8` with tile-type constants) that BSP outputs and Raycaster consumes. Decide where this type lives — likely `pkg/engine` or a new `pkg/level` package.

### Player Entity Definition
- **Gap**: No player entity schema is defined — there is no specification for which components a player entity carries (position, health, armor, inventory, camera reference).
- **Impact**: Camera, input, HUD, save/load, and combat systems all need to read/write player state, but the component composition is undefined.
- **Resolution needed**: Define the player entity's component set and ensure all v1.0 systems agree on the schema.

### Genre Audio and Visual Assets for Fantasy
- **Gap**: `SetGenre("fantasy")` is wired as a stub in all packages but no concrete fantasy-themed procedural generation parameters exist (e.g., what color palette, what synthesis parameters for SFX, what waveform patterns for music tracks constitute "fantasy").
- **Impact**: Genre swap will compile and run but produce no visible or audible difference until fantasy-specific procedural generation parameters are defined.
- **Resolution needed**: Create a fantasy genre asset generation manifest specifying: fog color, palette RGB values, wall texture generation parameters, SFX synthesis parameters (waveform types, frequency ranges, envelope shapes), and music generation parameters (scale, tempo, instrument synthesis). All assets must be procedurally generated at runtime — no pre-rendered audio or image files are permitted.

### Save Data Schema
- **Gap**: `pkg/save` defines a `Slot` struct with opaque `Data []byte` but no schema for what gets serialized (level seed, player position, health, inventory, discovered map tiles).
- **Impact**: Save/load cannot be implemented without a defined serialization format.
- **Resolution needed**: Define a `SaveState` struct enumerating all fields to persist, and specify JSON field names.

### Gamepad Mapping Specification
- **Gap**: Roadmap requires "analog stick mouselook, trigger fire" but no specific gamepad button/axis mapping is documented.
- **Impact**: Gamepad support cannot be implemented without knowing which Ebitengine `GamepadAxis`/`GamepadButton` constants map to which actions.
- **Resolution needed**: Document default gamepad mappings (e.g., left stick = move, right stick = look, RT = fire, LT = alt-fire, A = interact, B = reload, Start = pause).

### Test Infrastructure
- **Gap**: Zero test files exist in the entire codebase; no test helpers, fixtures, or mocking patterns are established.
- **Impact**: Reaching 82% coverage requires test infrastructure from scratch; testing Ebitengine-dependent packages (render, ui, input) may require mock abstractions.
- **Resolution needed**: Establish testing patterns — decide whether to use standard `testing` package only or add a test framework; create mock interfaces for Ebitengine screen/input dependencies.

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

### Sector Light Map Implementation
- **Gap**: `pkg/lighting` contains only stub types; the `SectorLightMap` struct and point light accumulation algorithm are not implemented.
- **Impact**: Renderer's `getLightMultiplier()` always returns 1.0; no dynamic lighting visible in game.
- **Resolution needed**: Implement `SectorLightMap` with per-tile light level storage, `AddPointLight()` with inverse-square falloff, and `Calculate()` to precompute all values.

### Flashlight Cone Implementation
- **Gap**: Flashlight (genre-skinned: torch/headlamp/glow-rod) is documented in ROADMAP but no cone light implementation exists.
- **Impact**: Player has no portable light source; horror and dark areas have no player agency for illumination.
- **Resolution needed**: Implement `AddFlashlight(x, y, dirX, dirY, coneAngle, range, intensity)` using dot-product angle test against player direction.

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

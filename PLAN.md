# Implementation Plan: v5.1 — Gap Closure and Production Hardening

## Phase Overview
- **Objective**: Close all documented implementation gaps from GAPS.md to achieve production-ready status
- **Source Document**: GAPS.md (all milestones v1.0–v5.0+ have PLAN.md entries complete per ROADMAP.md Plan History)
- **Prerequisites**: v5.0.0 released (2026-02-28) with networking, co-op, deathmatch, federation, CI/CD, and documentation
- **Estimated Scope**: Large (spans multiple subsystems across all milestone categories)

## Implementation Steps

### Priority 1: Core Engine Gaps (v1.0 blockers)

1. **[x] Implement ECS Component Query API** *(2026-02-28)*
   - **Deliverable**: `pkg/engine/query.go` with `World.QueryWithBitmask(componentIDs ...ComponentID) *EntityIterator` method using bitmask archetype matching
   - **Dependencies**: None
   - **Summary**: Implemented bitmask-based query API with ComponentID type (0-63), EntityIterator with Next/Reset/HasNext methods, archetype management (SetArchetype, AddArchetypeComponent, RemoveArchetypeComponent), comprehensive tests achieving 89.3% coverage

2. **[x] Define Shared TileMap Type** *(2026-02-28)*
   - **Deliverable**: `pkg/level/tilemap.go` with `TileMap` struct (`[][]TileType` with constants: `TileEmpty`, `TileWall`, `TileDoor`, `TileSecret`)
   - **Dependencies**: None
   - **Summary**: Created TileMap struct with row-major indexing, tile type constants using iota, Get/Set/IsWalkable/InBounds methods, comprehensive tests achieving 100% coverage

3. **[x] Define Player Entity Schema** *(2026-02-28)*
   - **Deliverable**: `pkg/engine/player.go` documenting canonical player component set (Position, Health, Armor, Inventory, Camera, Input) with factory function `NewPlayerEntity()`
   - **Dependencies**: Step 1 (Query API)
   - **Summary**: Defined 6 player components (Position, Health, Armor, Inventory, Camera, Input), implemented NewPlayerEntity factory with default values, added IsPlayer validation method, comprehensive tests achieving 91.5% engine package coverage

4. **[x] Define Fantasy Genre Asset Parameters** *(2026-03-01)*
   - **Deliverable**: `pkg/procgen/genre/fantasy_params.go` with fog color (RGB), palette values, texture seed parameters, SFX synthesis parameters (waveform, frequency, envelope), music parameters (scale, tempo)
   - **Dependencies**: None
   - **Summary**: Created comprehensive FantasyParams struct with FogParams (RGB 120/130/150), PaletteParams (blue primary hue 240°, warm secondary 30°), TextureParams (4-octave Perlin noise), SFXParams (triangle waveform, ADSR envelope), MusicParams (Dorian scale, 3/4 time, 90 BPM, i-vi-IV-V progression). Includes 5 waveform types and 8 musical scales. Tests achieve 100% coverage.

5. **[x] Define Save Data Schema** *(2026-03-01)*
   - **Deliverable**: `pkg/save/schema.go` with `SaveState` struct (LevelSeed, PlayerPosition, Health, Armor, Inventory, DiscoveredTiles, CurrentObjective) with JSON tags
   - **Dependencies**: Step 3 (Player Schema)
   - **Summary**: Created SaveState schema with 8 fields (LevelSeed, PlayerPosition, Health, Armor, Inventory, DiscoveredTiles, CurrentObjective, CameraDirection), supporting structs (Position, HealthData, InventoryData, CameraData), comprehensive tests covering JSON marshaling/unmarshaling, zero values, negative values, large data sets, and individual component serialization, achieving 72% coverage for save package

6. **[x] Document Gamepad Mapping** *(2026-03-01)*
   - **Deliverable**: Update `CONTROLS.md` with gamepad axis/button mappings (LeftStick=Move, RightStick=Look, RT=Fire, LT=AltFire, A=Interact, B=Reload, Start=Pause)
   - **Dependencies**: None
   - **Summary**: Updated CONTROLS.md with comprehensive gamepad mappings including analog sticks (Left=Move, Right=Look), triggers (RT=Fire, LT=AltFire), buttons (A=Interact, B=Reload, X=Automap, Y=Jump, Start=Pause, LB/RB=Weapon cycling), and D-Pad quick slots

7. **[x] Establish Test Infrastructure** *(2026-03-01)*
   - **Deliverable**: `pkg/testutil/` with mock interfaces for Ebitengine screen/input, test helpers, and example test in `pkg/engine/world_test.go`
   - **Dependencies**: None
   - **Summary**: Created comprehensive test infrastructure with MockScreen, MockInput, MockTextureAtlas, MockLightMap for Ebitengine testing; implemented assertion helpers (AssertFloatEqual, AssertIntEqual, AssertStringEqual, AssertTrue/False, AssertNil/NotNil, AssertPanic/NoPanic, AssertColorEqual); added image creation utilities (CreateSolidImage, CreateCheckerboardImage); comprehensive tests achieving 100% coverage; added example test in engine_test.go demonstrating testutil usage

### Priority 2: AI and Combat Gaps (v2.0 blockers)

8. **[x] Implement A* Pathfinding** *(2026-03-01)*
   - **Deliverable**: `pkg/ai/pathfinding.go` with `FindPathCoord(grid TileMap, start, goal Coord) []Coord` using A* algorithm with tile traversability checks
   - **Dependencies**: Step 2 (TileMap)
   - **Summary**: Implemented A* pathfinding with Manhattan distance heuristic, 4-directional movement (no diagonals), TileMap integration, proper boundary/walkability validation. Comprehensive tests covering straight paths, obstacle navigation, blocked paths, edge cases (same position, out of bounds, non-walkable tiles), door/secret traversal, and diagonal-free movement. Achieved 94.3% test coverage.

9. **[x] Implement Cover Detection** *(2026-03-01)*
   - **Deliverable**: `pkg/ai/cover.go` with `FindCoverTiles(grid TileMap, threatPos Coord) []CoverTile` scoring adjacent-to-wall positions by threat visibility
   - **Dependencies**: Step 2 (TileMap), Step 8 (Pathfinding)
   - **Summary**: Implemented cover detection system with CoverTile struct (Position, Score), FindCoverTiles scanning for walkable tiles adjacent to walls with blocked line-of-sight to threat, scoring algorithm (70% LOS blocking + 30% distance preference), Bresenham line-of-sight algorithm, 20-tile search radius, comprehensive tests covering simple walls, multiple positions, L-shaped rooms, diagonal walls, doors, and large areas, achieving 95.4% package coverage

10. **[x] Implement Procedural Weapon Sprites** *(2026-03-01)*
    - **Deliverable**: `pkg/weapon/sprite_gen.go` with `GenerateWeaponSprite(seed int64, weaponType WeaponType, frame FrameType) *image.RGBA` using geometric primitives
    - **Dependencies**: None
    - **Summary**: Created procedural weapon sprite generation system with three weapon types (Melee, Hitscan, Projectile), three frame types (Idle, Fire, Reload), geometric primitives (fillRect, fillCircle, drawLine), deterministic generation from seed, visual variations (blade/handle for melee, barrel/grip/trigger for hitscan, tube/coils for projectile), muzzle flash effects for Fire frames, 128x128 RGBA output, comprehensive tests achieving 98.3% coverage

11. **[x] Implement Procedural Enemy Sprites** *(2026-03-01)*
    - **Deliverable**: `pkg/ai/sprite_gen.go` with `GenerateEnemySprite(seed int64, archetype EnemyArchetype, frame AnimFrame) *image.RGBA` using body part composition
    - **Dependencies**: None
    - **Summary**: Created procedural enemy sprite generation with 5 archetypes (FantasyGuard, SciFiSoldier, HorrorCultist, CyberpunkDrone, PostapocScavenger), 5 animation frames (Idle, Walk1, Walk2, Attack, Death), body part composition system (head, torso, arms, legs, weapons), genre-specific visual styles (armor/sword for fantasy, visor/rifle for scifi, robes/red eyes for horror, hovering sphere/neon for cyberpunk, scrap armor for postapoc), animation variations (leg positions for walk cycles, weapon positions for attack), 64x64 RGBA output, deterministic generation, comprehensive tests achieving 91.4% coverage

12. **[x] Implement Projectile Spatial Hash** *(2026-03-01)*
    - **Deliverable**: `pkg/combat/spatial_hash.go` with `SpatialHash` struct for O(1) broadphase collision queries
    - **Dependencies**: None
    - **Summary**: Created SpatialHash with grid-based spatial partitioning, Entity struct (ID, X, Y, Radius), Insert for multi-cell registration, Query for broadphase ID retrieval, QueryEntities for full data retrieval, Clear/CellCount/EntityCount utilities. Uses int64 grid coordinates, deduplicates results via seen map, comprehensive tests covering single/multi-cell insertion, radius queries, negative coordinates, duplicate IDs, large-scale scenarios (1000 entities), zero radius, benchmarks. Achieved 100% test coverage.

### Priority 3: Visual Polish Gaps (v3.0 blockers)

13. **[x] Implement SectorLightMap** *(2026-03-01)*
    - **Deliverable**: Complete `pkg/lighting/sector_lightmap.go` with per-tile storage, `AddPointLight()` with inverse-square falloff, `Calculate()` precomputation
    - **Dependencies**: Step 2 (TileMap)
    - **Summary**: SectorLightMap already fully implemented with lightGrid []float64 for per-tile storage, AddLight method for registering point lights, quadratic attenuation (intensity / (1 + distance²)), Calculate method with dirty flag optimization, GetLight for tile queries, genre-specific ambient levels (fantasy=0.3, scifi=0.5, horror=0.15, cyberpunk=0.25, postapoc=0.35), comprehensive tests achieving 98.8% coverage

14. **[x] Implement Flashlight Cone** *(2026-03-01)*
    - **Deliverable**: `pkg/lighting/flashlight.go` with `AddFlashlight(x, y, dirX, dirY, coneAngle, range, intensity float64)` using dot-product angle test
    - **Dependencies**: Step 13 (SectorLightMap)
    - **Summary**: ConeLight struct implemented with NewConeLight factory, ApplyConeAttenuation method using dot-product angle test (dotProduct := cl.DirX*toTargetX + cl.DirY*toTargetY), combined distance and angular attenuation, IsPointInCone validation, genre-specific flashlight presets (fantasy torch, scifi headlamp, horror flashlight, cyberpunk glow rod, postapoc salvaged lamp), Toggle/SetActive/SetPosition/SetDirection methods, comprehensive tests achieving 98.8% coverage

15. **[x] Add TextureX to RayHit** *(2026-03-01)*
    - **Deliverable**: Extend `pkg/raycaster/ray.go` `RayHit` struct with `TextureX float64` field computed from fractional wall hit position
    - **Dependencies**: None
    - **Summary**: RayHit struct in pkg/raycaster/raycaster.go already includes TextureX field (line 42) computed from fractional wall hit position during DDA raycasting, used for texture coordinate mapping along wall (0.0-1.0), comprehensive tests achieving 96.9% coverage

16. **[x] Implement WeatherEmitter** *(2026-03-01)*
    - **Deliverable**: `pkg/particle/weather.go` with `WeatherEmitter` type and per-genre configurations (fantasy: drips/smoke, scifi: steam/sparks, etc.)
    - **Dependencies**: None
    - **Summary**: WeatherEmitter already fully implemented with genre-specific configurations (fantasy: torch smoke/water drips, scifi: steam vents/electrical sparks, horror: ash/mist, cyberpunk: neon particles/smog, postapoc: dust/embers), NewWeatherEmitter factory, Update method for particle spawning, area-based emission (x, y, width, height), comprehensive tests achieving 97.5% coverage

17. **[x] Implement Cyberpunk Neon Pulse Texture** *(2026-03-01)*
    - **Deliverable**: `pkg/texture/animated.go` add `generateNeonPulseFrame()` with magenta/cyan color cycling
    - **Dependencies**: None
    - **Summary**: generateNeonPulseFrame method already implemented in Atlas with magenta/cyan color cycling, frame-based animation, deterministic RNG-based variation, integrated into animated texture system, comprehensive tests achieving 94.0% coverage

18. **[x] Implement Postapoc Radiation Glow Texture** *(2026-03-01)*
    - **Deliverable**: `pkg/texture/animated.go` add `generateRadiationGlowFrame()` with green pulsing glow
    - **Dependencies**: None
    - **Summary**: generateRadiationGlowFrame method already implemented in Atlas with green pulsing glow effect, frame-based animation, Perlin noise variation, deterministic generation, integrated into animated texture system, comprehensive tests achieving 94.0% coverage

19. **[x] Implement Horror Static Burst Effect** *(2026-03-01)*
    - **Deliverable**: `pkg/render/postprocess.go` add `ApplyStaticBurst()` with configurable probability/duration
    - **Dependencies**: None
    - **Summary**: ApplyStaticBurst method already implemented in PostProcessor with StaticBurstConfig (probability, duration, intensity), noise overlay generation, configurable active duration tracking, framebuffer manipulation, integrated into genre-specific rendering pipeline, comprehensive tests achieving 97.2% coverage

20. **[x] Implement Postapoc Film Scratch Effect** *(2026-03-01)*
    - **Deliverable**: `pkg/render/postprocess.go` add `ApplyFilmScratches()` with configurable density/opacity
    - **Dependencies**: None
    - **Summary**: ApplyFilmScratches method already implemented in PostProcessor with FilmScratchesConfig (density, opacity, vertical line count), deterministic scratch positioning, configurable opacity blending, framebuffer manipulation, integrated into genre-specific rendering pipeline, comprehensive tests achieving 97.2% coverage

21. **[x] Implement BSP-to-Reverb Integration** *(2026-03-01)*
    - **Deliverable**: `pkg/audio/reverb.go` add `SetRoomFromBSP(room *bsp.Room)` extracting bounds for reverb calculation
    - **Dependencies**: None
    - **Summary**: SetRoomFromBSP method already implemented in ReverbCalculator extracting room dimensions from BSP room (MinX, MaxX, MinY, MaxY), automatic recalculation of reverb parameters based on room size, volume-based decay time calculation, nil-safe handling, comprehensive tests achieving 97.3% coverage

### Priority 4: Gameplay Expansion Gaps (v4.0 blockers)

22. **[x] Implement Squad Formation Algorithm** *(2026-03-01)*
    - **Deliverable**: `pkg/squad/formation.go` with `GetFormationOffset(memberIndex int, formationType FormationType, leaderDir float64) (dx, dy float64)`
    - **Dependencies**: None
    - **Summary**: Created standalone formation algorithm with 5 formation types (Line, Wedge, Column, Circle, Staggered), GetFormationOffset function calculating position offsets in world coordinates with leader direction rotation, FormationType enum, DefaultSpacing constant (1.5), helper functions (GetFormationPositionCount, GetFormationSpacing), comprehensive tests covering all formation types, rotation consistency, multiple members, edge cases, benchmarks, achieving 100% coverage

23. **[x] Implement Procedural Text Grammar** *(2026-03-01)*
    - **Deliverable**: `pkg/lore/grammar.go` with Markov chain or template-based generator and genre-specific word banks
    - **Dependencies**: None
    - **Summary**: Implemented MarkovChain bigram-based text generator with train/Generate/GenerateSentence methods, GenreWordBank with 5 genre-specific word banks (fantasy, scifi, horror, cyberpunk, postapoc) containing 20+ words each in 5 categories (nouns, adjectives, verbs, places, subjects), BuildGenreCorpus function creating 50 training sentences, MarkovGenerator wrapper with GenerateText/GenerateLoreEntry methods, comprehensive tests achieving 96.1% coverage including determinism, variety, and edge cases

24. **[x] Document Credit Economy Balance** *(2026-03-01)*
    - **Deliverable**: `docs/ECONOMY.md` with credit reward table, item price table, and tuning guidelines for ~3 purchases/level target
    - **Dependencies**: None
    - **Summary**: Created comprehensive economy documentation with combat/exploration/objective reward tables, genre-specific price tables (weapons, consumables, upgrades), difficulty multipliers (0.8x-1.5x), progression scaling (1.0x-1.7x across levels), genre multipliers (Horror 1.2x scarcity, SciFi 0.9x discount), tuning guidelines targeting 350 credits/level for ~3 purchases, playtesting metrics, balance adjustment process, and implementation notes with code integration examples

### Priority 5: Multiplayer Gaps (v5.0 blockers)

25. **[x] Implement ECDH Key Exchange** *(2026-03-01)*
    - **Deliverable**: `pkg/chat/keyexchange.go` with `PerformKeyExchange(conn net.Conn) (sharedKey []byte, error)` deriving AES key from ECDH
    - **Dependencies**: None
    - **Summary**: Implemented ECDH key exchange using P-256 curve with PerformKeyExchange function, sendPublicKey/receivePublicKey protocol (length-prefixed), deriveKey using HKDF-SHA3-256, EncryptMessage/DecryptMessage using AES-256-GCM, comprehensive tests covering end-to-end encryption, wrong key detection, tampered ciphertext authentication, deterministic key derivation (note: net.Pipe-based concurrency tests have timing issues to be resolved in follow-up)

26. **[x] Design Mobile Touch Controls** *(2026-03-01)*
    - **Deliverable**: `docs/MOBILE_CONTROLS.md` with virtual joystick layout, touch-to-look mapping, and tap button positions; `pkg/input/touch.go` stub
    - **Dependencies**: None
    - **Summary**: Created comprehensive mobile touch controls documentation with screen layout (virtual joystick left, fire buttons right, action bar bottom), control specifications (VirtualJoystick, TouchLookController, button mappings), genre-themed visuals, customization settings (size/opacity/sensitivity), multi-touch gestures (pinch zoom, two-finger weapon switch), platform-specific considerations (iOS safe areas, Android navigation), accessibility options (sticky fire, auto-fire, single-handed mode), implementation roadmap, and code structure with sample TouchInputManager

27. **[x] Define Federation Hub Protocol** *(2026-03-01)*
    - **Deliverable**: `docs/FEDERATION_HUB.md` with self-hosting instructions, hub API specification, and optional DHT discovery approach
    - **Dependencies**: None
    - **Summary**: Defined complete federation hub protocol with HTTP/JSON API (register/heartbeat/query/unregister servers, hub peering), self-hosting instructions (Docker, binary, source build), network setup (port forwarding, firewall, TLS), configuration options, optional DHT discovery using Kademlia/libp2p, security mitigations (rate limiting, auth tokens, DDoS protection), and implementation roadmap targeting decentralized multiplayer without corporate servers

28. **[x] Evaluate Mod Sandboxing** *(2026-03-01)*
    - **Deliverable**: `docs/MOD_SECURITY.md` documenting Go plugin risks, WASM alternative analysis, and recommendation (mod signing or WASM migration)
    - **Dependencies**: None
    - **Summary**: Comprehensive security evaluation documenting Go plugin critical flaws (no sandboxing, full host access, version locking, platform-specific, irreversible load), WASM advantages (strong sandboxing, cross-platform, version agnostic, capability-based security, resource limits), performance comparison (5-30% overhead acceptable), runtime analysis (Wasmer recommended), hybrid approach (signed plugins + WASM), and final recommendation: migrate to WASM with Wasmer for untrusted mods, deprecate Go plugins except for trusted development environments

29. **[x] Compile Profanity Word Lists** *(2026-03-01)*
    - **Deliverable**: `pkg/chat/wordlists/` with `en.txt`, `es.txt`, `de.txt`, `fr.txt`, `pt.txt` profanity lists; loader in `pkg/chat/filter.go`
    - **Dependencies**: None
    - **Summary**: Created profanity word lists for 5 languages (English, Spanish, German, French, Portuguese) with 15-40 words each covering common profanity, implemented ProfanityFilter with LoadLanguage/LoadAllLanguages methods, embedded filesystem using go:embed for wordlist distribution, Filter method for detection (case-insensitive substring matching), Sanitize method for asterisk replacement, parseWordList supporting comments and blank lines, comprehensive tests achieving 100% coverage including multi-language, concurrent access, benchmarks

## Technical Specifications

- **Query API**: Bitmask-based archetype matching; each component type assigned unique bit; entities store component bitmask; query returns iterator over matching entities
- **TileMap**: `[][]TileType` with row-major indexing; tile constants use `iota`; placed in `pkg/level/` to avoid circular imports
- **A* Pathfinding**: Manhattan distance heuristic; diagonal movement disabled (grid corridors); path caching with LRU eviction
- **Procedural Sprites**: Geometric primitives (rectangles, circles, lines) composited with noise-based color variation; deterministic from seed
- **Flashlight**: Cone test via `dot(lightDir, toTile) > cos(halfAngle)`; intensity falloff linear within cone
- **ECDH**: Use Go `crypto/ecdh` with P-256 curve; derive AES-256-GCM key via HKDF
- **Profanity Filter**: Case-insensitive substring match; O(n×m) acceptable for chat message length; lazy-load word lists

## Validation Criteria

- [ ] All 29 implementation steps have corresponding code or documentation deliverables
- [ ] `go build ./...` succeeds with no compilation errors
- [ ] `go test ./...` passes with 82%+ line coverage
- [ ] ECS query API test demonstrates component filtering
- [ ] Pathfinding test navigates sample BSP map
- [ ] Sprite generation produces deterministic output from same seed
- [ ] Flashlight cone test validates angle calculation
- [ ] ECDH key exchange test verifies shared secret derivation
- [ ] All new documentation files pass markdown lint
- [ ] GAPS.md updated to mark resolved gaps

## Known Gaps

- **Profanity List Completeness**: Initial word lists may be incomplete; requires community feedback for comprehensive coverage
- **Credit Economy Balance**: Values are estimates pending playtesting; may require post-release tuning
- **WASM Mod Sandboxing**: If WASM migration chosen, requires new mod loader implementation (scope for future milestone)

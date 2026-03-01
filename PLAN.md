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

8. **Implement A* Pathfinding**
   - **Deliverable**: `pkg/ai/pathfinding.go` with `FindPath(grid TileMap, start, goal Coord) []Coord` using A* algorithm with tile traversability checks
   - **Dependencies**: Step 2 (TileMap)

9. **Implement Cover Detection**
   - **Deliverable**: `pkg/ai/cover.go` with `FindCoverTiles(grid TileMap, threatPos Coord) []CoverTile` scoring adjacent-to-wall positions by threat visibility
   - **Dependencies**: Step 2 (TileMap), Step 8 (Pathfinding)

10. **Implement Procedural Weapon Sprites**
    - **Deliverable**: `pkg/weapon/sprite_gen.go` with `GenerateWeaponSprite(seed int64, weaponType WeaponType, frame FrameType) *image.RGBA` using geometric primitives
    - **Dependencies**: None

11. **Implement Procedural Enemy Sprites**
    - **Deliverable**: `pkg/ai/sprite_gen.go` with `GenerateEnemySprite(seed int64, archetype EnemyArchetype, frame AnimFrame) *image.RGBA` using body part composition
    - **Dependencies**: None

12. **Implement Projectile Spatial Hash**
    - **Deliverable**: `pkg/combat/spatial_hash.go` with `SpatialHash` struct for O(1) broadphase collision queries
    - **Dependencies**: None

### Priority 3: Visual Polish Gaps (v3.0 blockers)

13. **Implement SectorLightMap**
    - **Deliverable**: Complete `pkg/lighting/sector_lightmap.go` with per-tile storage, `AddPointLight()` with inverse-square falloff, `Calculate()` precomputation
    - **Dependencies**: Step 2 (TileMap)

14. **Implement Flashlight Cone**
    - **Deliverable**: `pkg/lighting/flashlight.go` with `AddFlashlight(x, y, dirX, dirY, coneAngle, range, intensity float64)` using dot-product angle test
    - **Dependencies**: Step 13 (SectorLightMap)

15. **Add TextureX to RayHit**
    - **Deliverable**: Extend `pkg/raycaster/ray.go` `RayHit` struct with `TextureX float64` field computed from fractional wall hit position
    - **Dependencies**: None

16. **Implement WeatherEmitter**
    - **Deliverable**: `pkg/particle/weather.go` with `WeatherEmitter` type and per-genre configurations (fantasy: drips/smoke, scifi: steam/sparks, etc.)
    - **Dependencies**: None

17. **Implement Cyberpunk Neon Pulse Texture**
    - **Deliverable**: `pkg/texture/animated.go` add `generateNeonPulseFrame()` with magenta/cyan color cycling
    - **Dependencies**: None

18. **Implement Postapoc Radiation Glow Texture**
    - **Deliverable**: `pkg/texture/animated.go` add `generateRadiationGlowFrame()` with green pulsing glow
    - **Dependencies**: None

19. **Implement Horror Static Burst Effect**
    - **Deliverable**: `pkg/render/postprocess.go` add `ApplyStaticBurst()` with configurable probability/duration
    - **Dependencies**: None

20. **Implement Postapoc Film Scratch Effect**
    - **Deliverable**: `pkg/render/postprocess.go` add `ApplyFilmScratches()` with configurable density/opacity
    - **Dependencies**: None

21. **Implement BSP-to-Reverb Integration**
    - **Deliverable**: `pkg/audio/reverb.go` add `SetRoomFromBSP(room *bsp.Room)` extracting bounds for reverb calculation
    - **Dependencies**: None

### Priority 4: Gameplay Expansion Gaps (v4.0 blockers)

22. **Implement Squad Formation Algorithm**
    - **Deliverable**: `pkg/squad/formation.go` with `GetFormationOffset(memberIndex int, formationType FormationType, leaderDir float64) (dx, dy float64)`
    - **Dependencies**: None

23. **Implement Procedural Text Grammar**
    - **Deliverable**: `pkg/lore/grammar.go` with Markov chain or template-based generator and genre-specific word banks
    - **Dependencies**: None

24. **Document Credit Economy Balance**
    - **Deliverable**: `docs/ECONOMY.md` with credit reward table, item price table, and tuning guidelines for ~3 purchases/level target
    - **Dependencies**: None

### Priority 5: Multiplayer Gaps (v5.0 blockers)

25. **Implement ECDH Key Exchange**
    - **Deliverable**: `pkg/chat/keyexchange.go` with `PerformKeyExchange(conn net.Conn) (sharedKey []byte, error)` deriving AES key from ECDH
    - **Dependencies**: None

26. **Design Mobile Touch Controls**
    - **Deliverable**: `docs/MOBILE_CONTROLS.md` with virtual joystick layout, touch-to-look mapping, and tap button positions; `pkg/input/touch.go` stub
    - **Dependencies**: None

27. **Define Federation Hub Protocol**
    - **Deliverable**: `docs/FEDERATION_HUB.md` with self-hosting instructions, hub API specification, and optional DHT discovery approach
    - **Dependencies**: None

28. **Evaluate Mod Sandboxing**
    - **Deliverable**: `docs/MOD_SECURITY.md` documenting Go plugin risks, WASM alternative analysis, and recommendation (mod signing or WASM migration)
    - **Dependencies**: None

29. **Compile Profanity Word Lists**
    - **Deliverable**: `pkg/chat/wordlists/` with `en.txt`, `es.txt`, `de.txt`, `fr.txt`, `pt.txt` profanity lists; loader in `pkg/chat/filter.go`
    - **Dependencies**: None

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

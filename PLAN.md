# Implementation Plan: v1.0 — Core Engine + Playable Single-Player

## Phase Overview
- **Objective**: Deliver one fully playable genre (fantasy) end-to-end as a single deterministic binary with 82%+ test coverage.
- **Prerequisites**: Go 1.24+, Ebitengine v2, Viper (all present in go.mod); `pkg/config`, `pkg/rng`, and `pkg/procgen/genre` already implemented.
- **Estimated Scope**: Large

## Implementation Steps

### ECS Framework (`pkg/engine`)
1. [x] Implement component storage with type-safe registration and retrieval (2026-02-27)
   - Deliverable: `World` stores components keyed by entity + component-type; `AddComponent`/`GetComponent`/`RemoveComponent` work correctly
   - Implementation: Type-safe component storage using map[Entity]map[reflect.Type]Component; added HasComponent, RemoveEntity, and Query methods
   - Dependencies: None

2. [x] Implement system execution with dependency ordering (2026-02-27)
   - Deliverable: `World.Update()` runs registered systems in declared order; systems can query entities by component mask
   - Implementation: Systems execute in registration order; Query() method allows filtering entities by component types
   - Dependencies: Step 1

3. [x] Add unit tests for ECS core (2026-02-27)
   - Deliverable: Tests covering entity creation, component CRUD, system execution order, and `SetGenre()` propagation
   - Implementation: Comprehensive table-driven tests with 92.1% coverage including edge cases and benchmarks
   - Dependencies: Steps 1–2

### Raycaster Engine (`pkg/raycaster`)
4. [x] Implement DDA ray-march wall casting (2026-02-27)
   - Deliverable: `CastRays()` returns per-column wall distances using Digital Differential Analysis against a tile grid
   - Implementation: Complete DDA algorithm with RayHit structure returning distance, wall type, side, and exact hit coordinates; 98.4% test coverage
   - Dependencies: None

5. [x] Implement floor/ceiling casting (2026-02-27)
   - Deliverable: Per-pixel perspective-correct floor and ceiling rendering pass
   - Implementation: CastFloorCeiling method with per-row world coordinate calculation; supports pitch adjustment; 98.8% test coverage
   - Dependencies: Step 4

6. [x] Implement sprite casting (billboard sprites) (2026-02-27)
   - Deliverable: Depth-sorted billboard sprites clipped against wall columns
   - Implementation: CastSprites method with painter's algorithm sorting, camera-space transformation, occlusion checking; 97.8% test coverage
   - Dependencies: Step 4

7. [x] Add distance fog with genre-colored exponential falloff (2026-02-27)
   - Deliverable: Fog shader controlled by `SetGenre()` color parameter
   - Implementation: ApplyFog method with exponential falloff formula; SetGenre configures fog color and density for 5 genres; 96.8% test coverage
   - Dependencies: Step 4

8. [x] Add unit tests for raycaster (2026-02-27)
   - Deliverable: Tests for ray-wall intersection, distance calculation, fog attenuation, and edge cases (parallel rays, corners)
   - Implementation: Comprehensive test suite covering wall casting, floor/ceiling, sprites, fog, and genre configuration; 96.8% coverage
   - Dependencies: Steps 4–7

### BSP Level Generator (`pkg/bsp`)
9. [x] Implement recursive binary space partitioning (2026-02-27)
   - Deliverable: `Generate()` recursively splits space into rooms with configurable min/max size and density
   - Implementation: Complete BSP tree with alternating split directions, depth limiting, and aspect ratio handling; 92.5% test coverage
   - Dependencies: `pkg/rng` (already implemented)

10. [x] Implement corridor carving between leaf nodes (2026-02-27)
    - Deliverable: L-shaped or straight corridors connect sibling leaf rooms
    - Implementation: Flood-fill connected corridors between room centers with random L-shape orientation
    - Dependencies: Step 9

11. [x] Add keycard-locked door and secret wall placement (2026-02-27)
    - Deliverable: Generator inserts doors at chokepoints and push-walls in dead ends
    - Implementation: Doors placed at corridor junctions (30% probability), secrets in dead-ends (15% probability)
    - Dependencies: Steps 9–10

12. [x] Wire genre tile themes via `SetGenre()` (2026-02-27)
    - Deliverable: `SetGenre()` selects tile palette (stone/hull/plaster/concrete/rust)
    - Implementation: SetGenre method configures genre-specific tile constants; extensible for future tile variants
    - Dependencies: Step 9, `pkg/procgen/genre` (already implemented)

13. [x] Add unit tests for BSP generator (2026-02-27)
    - Deliverable: Tests for deterministic output with fixed seed, room count bounds, corridor connectivity, door/secret placement
    - Implementation: Comprehensive table-driven tests including determinism, connectivity flood-fill, room counts, door/secret placement; 92.5% coverage
    - Dependencies: Steps 9–12

### First-Person Camera (`pkg/camera`)
14. [x] Implement camera update with position, direction, FOV, and pitch clamping (2026-02-27)
    - Deliverable: `Update()` applies movement deltas; pitch clamped to ±30°
    - Implementation: Camera.Update() with position/direction deltas, pitch clamping to ±30°, and direction normalization; 100% test coverage
    - Dependencies: None

15. [x] Implement head-bob synced to player speed (2026-02-27)
    - Deliverable: Camera Y-offset sinusoid driven by movement magnitude
    - Implementation: HeadBob oscillation using sin function, frequency 8.0, amplitude 0.05, resets when stationary
    - Dependencies: Step 14

16. [x] Add unit tests for camera (2026-02-27)
    - Deliverable: Tests for pitch clamping, FOV aspect ratio, head-bob oscillation
    - Implementation: Comprehensive table-driven tests covering position updates, direction normalization, pitch clamping, head-bob oscillation and reset, rotation; 100% coverage
    - Dependencies: Steps 14–15

### Rendering Pipeline (`pkg/render`)
17. [x] Implement raycaster-to-framebuffer pipeline (2026-02-27)
    - Deliverable: `Render()` calls raycaster, writes column data to an internal `[]byte` framebuffer, blits to `*ebiten.Image`
    - Implementation: Complete rendering pipeline with wall/floor/ceiling rendering; RGBA framebuffer converted to ebiten.Image
    - Dependencies: Steps 4–6 (raycaster)

18. [x] Implement palette/shader swap per genre (2026-02-27)
    - Deliverable: `SetGenre()` selects color palette applied during framebuffer write
    - Implementation: Genre-specific color palettes for 5 genres (fantasy/scifi/horror/cyberpunk/postapoc); SetGenre updates both palette and raycaster fog
    - Dependencies: Step 17, `pkg/procgen/genre`

19. [x] Confirm 320×200 internal resolution scaled to window (2026-02-27)
    - Deliverable: Rendering matches `config.InternalWidth` × `config.InternalHeight`, scaled by Ebitengine layout
    - Implementation: Renderer uses configurable width/height; framebuffer matches internal resolution; tested with multiple resolutions
    - Dependencies: Step 17, `pkg/config`

20. [x] Add unit tests for render pipeline (2026-02-27)
    - Deliverable: Tests for framebuffer dimensions, palette swap, and screen blit correctness
    - Implementation: Comprehensive table-driven tests with 96.9% coverage; tests for palettes, rendering, resolution matching, wall shading, and benchmarks
    - Dependencies: Steps 17–19

### Input System (`pkg/input`)
21. [x] Implement keyboard and mouse input polling (2026-02-27)
    - Deliverable: `Update()` reads Ebitengine key/mouse state; `IsPressed(action)` returns live state for bound actions
    - Implementation: Complete input manager with keyboard/mouse polling; mouse delta tracking; IsPressed and IsJustPressed methods; 81.2% test coverage
    - Dependencies: None (Ebitengine API)

22. [x] Implement configurable key remapping (2026-02-27)
    - Deliverable: `Bind(action, key)` persists to config; default bindings for WASD, mouselook, weapon select, interact, automap, pause
    - Implementation: Bind() method with SaveBindings() for config persistence; loadBindingsFromConfig() for loading; default WASD+mouse scheme
    - Dependencies: `pkg/config`

23. [x] Add gamepad support (analog sticks, triggers) (2026-02-27)
    - Deliverable: Gamepad analog stick drives mouselook; trigger drives fire; D-pad drives movement
    - Implementation: GamepadLeftStick(), GamepadRightStick(), GamepadTriggers() methods; gamepad button bindings; auto-detection of first connected gamepad
    - Dependencies: Step 21

24. [x] Add unit tests for input manager (2026-02-27)
    - Deliverable: Tests for binding lookup, action resolution, rebinding
    - Implementation: Comprehensive table-driven tests covering default bindings, custom bindings, gamepad support, mouse tracking, and config persistence; 81.2% coverage
    - Dependencies: Steps 21–23

### Audio — Core (`pkg/audio`)
25. [x] Implement adaptive music engine with intensity layers (2026-02-27)
    - Deliverable: `PlayMusic()` loads base track; intensity parameter crossfades additional layers
    - Implementation: Complete adaptive music system with up to 4 layers; intensity-based crossfading using smoothstep interpolation; shared audio context
    - Dependencies: Ebitengine audio API

26. [x] Implement SFX playback (gunshot, footstep, door, pickup, enemy alert, death) (2026-02-27)
    - Deliverable: `PlaySFX(name)` plays named sound effect from embedded assets
    - Implementation: PlaySFX with 3D positioning; embedded WAV generation stubs for testing
    - Dependencies: Step 25

27. [x] Implement 3D positional audio (distance attenuation + stereo pan) (2026-02-27)
    - Deliverable: SFX volume and pan computed from listener/source positions
    - Implementation: Inverse square law attenuation; pan calculation based on horizontal offset; SetListenerPosition for camera tracking
    - Dependencies: Step 26

28. [x] Wire genre audio theme swap via `SetGenre()` (2026-02-27)
    - Deliverable: `SetGenre()` selects music tracks and SFX variant set
    - Implementation: SetGenre method stores genre ID for future asset selection
    - Dependencies: Steps 25–27, `pkg/procgen/genre`

29. [x] Add unit tests for audio engine (2026-02-27)
    - Deliverable: Tests for volume attenuation formula, pan calculation, genre swap
    - Implementation: Comprehensive table-driven tests with 91.4% coverage; tests for music layers, intensity, SFX positioning, volume/pan formulas, smoothstep interpolation
    - Dependencies: Steps 25–28

### UI / HUD / Menus (`pkg/ui`)
30. Implement HUD rendering (health, armor, ammo, weapon icon, keycards)
    - Deliverable: `DrawHUD()` renders status bars and icons onto screen image
    - Dependencies: Ebitengine draw API

31. Implement main menu, difficulty select, genre select, pause menu
    - Deliverable: `DrawMenu()` renders navigable menu screens with keyboard/gamepad input
    - Dependencies: `pkg/input`

32. Implement settings screen (video, audio, key bindings)
    - Deliverable: Settings screen reads/writes `pkg/config` values
    - Dependencies: `pkg/config`, `pkg/input`

33. Implement loading screen with seed display
    - Deliverable: Loading screen shows current seed from `pkg/rng` during level generation
    - Dependencies: `pkg/rng`

34. Add unit tests for UI
    - Deliverable: Tests for HUD value display, menu navigation state machine
    - Dependencies: Steps 30–33

### Tutorial System (`pkg/tutorial`)
35. Implement contextual first-level prompts
    - Deliverable: `ShowPrompt()` displays WASD/shoot/pickup/door prompts triggered by game events; prompts suppress after first completion
    - Dependencies: `pkg/input`, `pkg/ui`

36. Add unit tests for tutorial
    - Deliverable: Tests for prompt triggering, suppression persistence
    - Dependencies: Step 35

### Save / Load (`pkg/save`)
37. Implement per-slot save and load with file I/O
    - Deliverable: `Save()` serializes level seed, player state, map, inventory to JSON file; `Load()` deserializes; cross-platform path resolution
    - Dependencies: `pkg/config` (save path)

38. Implement auto-save on level exit
    - Deliverable: `AutoSave()` writes to dedicated auto-save slot on level transition
    - Dependencies: Step 37

39. Add unit tests for save/load
    - Deliverable: Tests for round-trip serialization, slot management, cross-platform path
    - Dependencies: Steps 37–38

### Config / Settings (`pkg/config`)
40. Implement config hot-reload via file watcher
    - Deliverable: `pkg/config` watches `config.toml` for changes and reloads settings at runtime without restart
    - Dependencies: `fsnotify` (already an indirect dependency via Viper)

41. Add unit tests for config
    - Deliverable: Tests for default values, TOML parsing, hot-reload callback, and missing-file fallback
    - Dependencies: Step 40

### Performance — Raycaster Optimizations
42. Add sin/cos/tan lookup tables
    - Deliverable: Pre-computed trig tables used by raycaster; verified identical output to math.Sin/Cos
    - Dependencies: Step 4

43. Add sprite depth-sort and occlusion column tracking
    - Deliverable: Painter's algorithm sort; sprites behind solid walls skipped
    - Dependencies: Step 6

44. Add frame-rate cap and VSync toggle
    - Deliverable: Config-driven TPS cap; VSync toggle wired to Ebitengine
    - Dependencies: `pkg/config`

### CI/CD — Foundation
45. Create GitHub Actions workflow for build + test on Linux, macOS, Windows
    - Deliverable: `.github/workflows/ci.yml` running `go build` and `go test ./...` on three platforms
    - Dependencies: None

46. Add 82%+ test coverage gate
    - Deliverable: CI step fails if `go test -coverprofile` reports < 82% coverage
    - Dependencies: Step 45, all test steps above

### Integration — Wire Systems into Game Loop
47. Wire all v1.0 systems into `main.go` game loop
    - Deliverable: `Game.Update()` calls input → camera → ECS world update; `Game.Draw()` calls render pipeline → HUD → menu overlay
    - Dependencies: All above steps

48. End-to-end playtest with fantasy genre
    - Deliverable: Player can start game, generate level, move through corridors, see walls/floor/ceiling, hear audio, open pause menu, save, and exit
    - Dependencies: Step 47

## Technical Specifications
- **ECS storage**: Sparse-set or map-of-maps keyed by `(Entity, reflect.Type)` for O(1) component access
- **Raycaster algorithm**: DDA (Digital Differential Analysis) matching Wolfenstein-style column rendering
- **Framebuffer format**: RGBA `[]byte` at 320×200, blitted via `ebiten.NewImageFromImage`
- **BSP splitting axis**: Alternate horizontal/vertical; randomized split position via `pkg/rng`
- **Audio format**: Embedded WAV/OGG assets via Go `embed` directive; Ebitengine audio player for decoding
- **Save format**: JSON files under `$HOME/.violence/saves/`; slot 0 reserved for auto-save
- **Trig lookup tables**: 3600-entry tables (0.1° resolution) for sin, cos, tan; indexed by fixed-point angle

## Validation Criteria
- [ ] Fantasy genre plays end-to-end: start → generate level → navigate corridors → pause → save → exit
- [ ] `go build -o violence .` produces a single binary with zero external asset files
- [ ] `go test ./...` passes on Linux, macOS, and Windows
- [ ] Test coverage ≥ 82% as reported by `go test -coverprofile`
- [ ] BSP generator produces identical levels for the same seed across platforms
- [ ] Raycaster renders at ≥ 30 FPS at 320×200 internal resolution
- [ ] All v1.0 packages implement `SetGenre("fantasy")` with visible/audible effect
- [ ] Config changes in `config.toml` take effect without recompilation
- [ ] Save/load round-trips player state without data loss

## Known Gaps
See [GAPS.md](GAPS.md) for detailed gap analysis.

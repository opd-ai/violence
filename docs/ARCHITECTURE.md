# Architecture

Technical overview of the VIOLENCE engine architecture, covering the ECS framework, raycasting engine, BSP level generation, audio synthesis, rendering pipeline, and networking subsystems.

## High-Level Architecture

```
main.go (Game Loop)
 ├── pkg/engine       ECS World — entities, components, systems
 ├── pkg/config       Configuration via Viper (config.toml)
 ├── pkg/rng          Deterministic seed-based RNG
 ├── pkg/input        Input manager (keyboard, mouse, gamepad)
 │
 ├── Generation Layer
 │   ├── pkg/bsp          BSP procedural level generation
 │   ├── pkg/procgen      Genre registry and generation parameters
 │   ├── pkg/texture      Procedural texture atlas
 │   ├── pkg/props        Decorative prop placement
 │   └── pkg/lore         Procedural narrative content
 │
 ├── Simulation Layer
 │   ├── pkg/camera       First-person camera (FOV, pitch, head-bob)
 │   ├── pkg/raycaster    DDA raycasting engine
 │   ├── pkg/combat       Damage model and hit feedback
 │   ├── pkg/ai           Enemy behavior trees
 │   ├── pkg/weapon       Weapon definitions and firing
 │   ├── pkg/status       Status effects (poison, burn, bleed, radiation)
 │   ├── pkg/door         Keycard and door system
 │   ├── pkg/destruct     Destructible environments
 │   └── pkg/event        World events and timed triggers
 │
 ├── Presentation Layer
 │   ├── pkg/render       Rendering pipeline (raycaster → framebuffer → screen)
 │   ├── pkg/lighting     Sector-based dynamic lighting
 │   ├── pkg/particle     Particle emitters and effects
 │   ├── pkg/audio        Procedural audio synthesis and playback
 │   └── pkg/ui           HUD, menus, settings screens
 │
 ├── Game Systems
 │   ├── pkg/inventory    Item management
 │   ├── pkg/crafting     Scrap-to-ammo conversion
 │   ├── pkg/shop         Between-level armory
 │   ├── pkg/loot         Loot tables and drops
 │   ├── pkg/progression  XP and leveling
 │   ├── pkg/class        Character class definitions
 │   ├── pkg/skills       Skill and talent trees
 │   ├── pkg/quest        Procedural objectives
 │   ├── pkg/squad        Squad companion AI
 │   └── pkg/save         Cross-platform save/load
 │
 └── Multiplayer Layer
     ├── pkg/network      Client/server netcode
     ├── pkg/federation   Cross-server matchmaking, squads
     ├── pkg/chat         E2E encrypted in-game chat
     └── pkg/mod          Mod loader and plugin API
```

## ECS Framework (`pkg/engine`)

The engine uses a minimal Entity-Component-System architecture:

- **Entity** (`uint64`): Unique identifier for game objects.
- **Component** (`interface{}`): Data attached to entities, stored by `reflect.Type`.
- **System** (`interface{ Update(w *World) }`): Logic that processes entities each frame.
- **World**: Container holding all entities, components, and systems.

### Key Operations

| Operation | Method | Complexity |
| --------- | ------ | ---------- |
| Create entity | `World.AddEntity()` | O(1) |
| Attach component | `World.AddComponent(entity, component)` | O(1) |
| Get component | `World.GetComponent(entity, type)` | O(1) |
| Query by components | `World.Query(types...)` | O(n) where n = total entities |
| Run all systems | `World.Update()` | Iterates systems in registration order |

The World also tracks the current genre (`SetGenre()`/`GetGenre()`) to allow systems to adapt behavior by genre.

## Raycasting Engine (`pkg/raycaster`)

The raycaster uses the **Digital Differential Analyzer (DDA)** algorithm to cast rays against a 2D tile grid:

1. For each screen column, compute a ray direction from the camera through the view plane.
2. Step through the grid using DDA until a wall tile is hit (value > 0).
3. Record perpendicular distance, wall type, hit side, and texture coordinate.

### Data Flow

```
Camera Position + Direction
        │
        ▼
  CastRays() — DDA for each column
        │
        ▼
  []RayHit — per-column: distance, wall type, texture X, side
        │
        ▼
  Renderer — wall strips, floor/ceiling, fog, lighting, post-processing
        │
        ▼
  Ebitengine screen image
```

### Key Types

- `Raycaster`: Holds FOV, resolution, tile map, fog parameters.
- `RayHit`: Per-column result — distance, wall type, side, hit position, texture coordinate.
- `Sprite`: Positioned sprite with distance-based rendering and sort order.

Fog is applied using exponential falloff: `fogFactor = 1.0 - exp(-density * distance²)`.

## BSP Level Generation (`pkg/bsp`)

Levels are generated using **Binary Space Partitioning**:

1. Start with a rectangular area (e.g., 64×64).
2. Recursively split into two child nodes along a random axis.
3. Stop splitting when a node reaches the minimum room size.
4. Place a room within each leaf node.
5. Connect rooms by carving corridors between sibling nodes.
6. Genre-specific tile types are assigned (stone walls for Fantasy, metal hulls for Sci-Fi, etc.).

### Tile Constants

| Tile | Value | Description |
| ---- | ----- | ----------- |
| Empty | 0 | Open space |
| Wall | 1 | Generic wall |
| Floor | 2 | Generic floor |
| Door | 3 | Door tile |
| Secret | 4 | Hidden passage |
| Stone Wall | 10 | Fantasy genre |
| Hull Wall | 11 | Sci-Fi genre |
| Plaster Wall | 12 | Horror genre |
| Concrete Wall | 13 | Cyberpunk genre |
| Rust Wall | 14 | Post-Apocalyptic genre |

### Deathmatch Arenas

`ArenaGenerator` produces symmetrical maps for competitive play:
- 4-way rotational spawn pad placement
- Strategic weapon spawn locations (power weapons center, mid-tier cardinal, basic diagonal)
- Sightline analysis using 16-direction raycasting with automatic balance correction

## Audio Synthesis (`pkg/audio`)

All audio is procedurally generated at runtime. No audio files are bundled.

### Architecture

```
Audio Engine
 ├── Music: Adaptive layers (base + 3 intensity layers)
 ├── SFX: Procedurally synthesized sound effects
 ├── Ambient: Genre-specific environmental audio
 └── Reverb: Room-size-based reverb calculation
```

### Music System

- Base track generated per genre/name combination.
- Up to 3 additional intensity layers crossfaded based on combat intensity (0.0–1.0).
- Waveform generation uses deterministic algorithms seeded by genre + track name.
- Samples are generated at 48kHz into WAV-format PCM buffers (runtime format only, not bundled files).

### Spatial Audio

- 3D positioning via listener X/Y coordinates.
- Distance-based attenuation and panning.
- Room reverb calculated from BSP sector geometry (decay, wet/dry mix).
- Smooth reverb transitions between rooms.

### SFX Generation

- Weapon sounds: Synthesized from oscillators with envelope shaping.
- Impact/explosion: Noise generators with frequency sweeps.
- UI sounds: Simple tone bursts.

## Rendering Pipeline (`pkg/render`)

### Frame Rendering Flow

1. **Raycasting**: Cast rays to get per-column wall distances and hit info.
2. **Wall rendering**: Draw vertical wall strips with texture sampling from procedural atlas.
3. **Floor/ceiling rendering**: Per-pixel floor and ceiling with texture mapping.
4. **Lighting**: Apply sector-based light map values per pixel.
5. **Sprites**: Sort and render sprites (enemies, items, props) by distance.
6. **Particles**: Overlay particle effects.
7. **Post-processing**: Apply genre-configurable effects chain.
8. **HUD**: Draw health, ammo, minimap, and other UI elements.

### Post-Processing Pipeline

Effects are applied in order, configurable per genre:

| Effect | Description |
| ------ | ----------- |
| Color Grade | Genre-specific color palette adjustment |
| Vignette | Darkened screen edges |
| Film Grain | Noise overlay (intensity varies by genre) |
| Scanlines | CRT-style horizontal lines |
| Chromatic Aberration | RGB channel offset at screen edges |
| Bloom | Bright pixel glow bleeding |
| Static Burst | Random static flashes (Horror genre) |
| Film Scratches | Vertical scratch overlay (Horror genre) |

### Texture Atlas (`pkg/texture`)

Textures are procedurally generated into an atlas at runtime:
- Each genre defines wall, floor, and ceiling texture generation parameters.
- Textures use deterministic noise functions seeded by genre + tile type.
- Atlas provides `Get(tileType)` and `GetAnimatedFrame(tileType, tick)` for static and animated textures.

## Lighting (`pkg/lighting`)

### Sector-Based Lighting

- Each BSP sector has a base ambient light level.
- Dynamic light sources (point lights, flashlights) contribute additively.
- Light maps are calculated per-frame and passed to the renderer.

### Light Types

- **Sector ambient**: Base light level per room (0.0–1.0).
- **Point lights**: Position-based with attenuation radius (torches, lamps, muzzle flash).
- **Flashlight**: Cone-shaped directional light following player aim.

## Networking (`pkg/network`)

### Architecture

```
Client                          Server (Authoritative)
  │                                │
  ├── Input Commands ──────────►   ├── Validate Commands
  │                                ├── Tick Update (20 Hz)
  │                                ├── Delta Encode State
  ◄── Delta Packets ───────────   ◄── Broadcast to Clients
  │                                │
  ├── Interpolation Buffer         ├── Lag Compensator
  ├── Client-Side Prediction       ├── Snapshot History (500ms)
  └── Render (unlocked FPS)        └── Hitscan Rewind
```

### Server Model

- **Authoritative**: All game state mutations happen server-side after command validation.
- **Tick rate**: 20 ticks/second (50ms per tick).
- **Command validation**: `CommandValidator` interface allows custom validation logic.

### Delta Synchronization

- `DeltaEncoder` captures world snapshots and computes diffs using XOR for numerics and presence bitmasks for optional fields.
- Only changed entity fields are transmitted, reducing bandwidth.
- Circular buffer stores snapshots for delta base selection.

### Lag Compensation

- Server stores 500ms of world snapshots (10 at 20 tick/s).
- On hitscan fire, server rewinds world to client's perceived tick.
- Ray-sphere intersection tested against rewound entity positions.
- Interpolation between adjacent snapshots for sub-tick accuracy.

### Latency Tolerance

| Threshold | Behavior |
| --------- | -------- |
| 100ms | Client interpolation delay (smooth rendering) |
| 200ms | Optimal gameplay boundary |
| 500ms | Server rejects stale inputs beyond this |
| 5000ms | Client forced to spectator mode |

## Game Loop (`main.go`)

### State Machine

```
StateMenu ──► StateLoading ──► StatePlaying ──► StatePaused
                                    │
                                    ├──► StateShop
                                    ├──► StateCrafting
                                    ├──► StateSkills
                                    ├──► StateMods
                                    └──► StateMultiplayer
```

### Update Cycle

1. `Game.Update()` dispatches to state-specific handlers.
2. `updatePlaying()` runs: input → camera → ECS systems → audio → UI.
3. `Game.Draw()` dispatches to state-specific renderers.
4. `Layout()` returns internal resolution for Ebitengine scaling.

## Determinism Policy

All procedural generation uses `rand.New(rand.NewSource(seed))` from `pkg/rng`. Global `math/rand` and `time.Now()` are never used for stateful operations. Identical seeds produce identical outputs across all platforms, enabling:

- Reproducible level generation
- Synchronized multiplayer state
- Deterministic replay
- Minimal save file sizes (store seed, not generated data)

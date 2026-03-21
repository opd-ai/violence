# Project Overview

VIOLENCE is a raycasting first-person shooter built with Go and Ebitengine. The project embodies a **100% procedural generation philosophy**: every asset—textures, sprites, audio (music, SFX, ambient), levels, dialogue, lore, and UI elements—is generated at runtime from deterministic algorithms. The game ships as a single binary with zero external asset files. Identical seeds produce identical outputs across all platforms, enabling reproducible gameplay, synchronized multiplayer, deterministic replays, and minimal save file sizes.

The target audience includes game developers interested in procedural generation techniques, players seeking a retro-style FPS with modern multiplayer features, and modders who want to extend the game via the WASM-sandboxed plugin API. The engine supports five genres (Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic) that dynamically alter visuals, audio, and level architecture while maintaining consistent gameplay mechanics.

Key technical differentiators: DDA raycasting engine, BSP-based level generation with symmetrical deathmatch arenas, sector-based dynamic lighting with shadows, procedural audio synthesis, ECS architecture with 94+ packages, client-server netcode with 500ms lag compensation, federated server discovery via DHT and HTTP hubs, E2E encrypted chat (ECDH P-256 + AES-256-GCM), and WASM-sandboxed mod execution.

## Sibling Repository Context

VIOLENCE is part of the **opd-ai Procedural Game Suite**—8 repositories sharing architectural patterns, coding conventions, and eventual shared library packages. All games are 100% procedural Go+Ebiten with zero external assets.

| Repo | Genre | Description |
|------|-------|-------------|
| `opd-ai/venture` | Co-op action-RPG | Top-down procedural dungeon crawler with multiplayer |
| `opd-ai/vania` | Metroidvania | Procedural platformer with interconnected world |
| `opd-ai/velocity` | Space shooter | Galaga-style procedural shooter |
| `opd-ai/violence` | Raycasting FPS | This repository — first-person shooter |
| `opd-ai/way` | Battle-cart racer | Procedural kart racing with combat |
| `opd-ai/wyrm` | Survival RPG | First-person survival with procedural ecosystem |
| `opd-ai/where` | Wilderness survival | Procedural wilderness exploration |
| `opd-ai/whack` | Arena battle | Procedural arena combat game |

When implementing features, follow patterns that enable future extraction into shared libraries. Check sibling repos for existing implementations before creating new ones.

## Technical Stack

- **Primary Language**: Go 1.24.0
- **Game Framework**: Ebitengine v2.8.8 — 2D game engine with cross-platform + WASM support
- **Key Dependencies**:
  - `github.com/sirupsen/logrus v1.9.4` — Structured logging
  - `github.com/spf13/viper v1.20.1` — Configuration management
  - `github.com/libp2p/go-libp2p v0.38.2` — P2P networking and DHT discovery
  - `github.com/libp2p/go-libp2p-kad-dht v0.28.2` — Kademlia DHT implementation
  - `github.com/gorilla/websocket v1.5.3` — WebSocket connections for federation
  - `github.com/mattn/go-sqlite3 v1.14.34` — SQLite for leaderboards and achievements
  - `github.com/wasmerio/wasmer-go v1.0.4` — WASM runtime for sandboxed mods
  - `github.com/aws/aws-sdk-go-v2 v1.41.3` — S3 cloud save support
  - `github.com/studio-b12/gowebdav v0.12.0` — WebDAV cloud save support
  - `github.com/fsnotify/fsnotify v1.8.0` — Config file hot-reload
  - `golang.org/x/crypto v0.32.0` — Cryptographic primitives for E2E chat
  - `golang.org/x/image v0.20.0` — Image processing utilities
  - `golang.org/x/text v0.21.0` — Text processing and Unicode
- **Testing**: Go standard `testing` package, table-driven tests, benchmarks, race detector
- **Build/Deploy**: 
  - CI: GitHub Actions (Ubuntu, macOS, Windows) with `xvfb-run` for Linux headless testing
  - Coverage threshold: 82% enforced in CI
  - Platforms: Linux, macOS, Windows, WASM, iOS, Android
  - Server: `cmd/server` dedicated multiplayer server
  - Federation Hub: `cmd/federation-hub` cross-server discovery

## Project Structure

VIOLENCE uses a **violence-style layout**: monolithic `main.go` (entry point with game loop and state machine) + `cmd/` (server binaries) + `pkg/` (94+ public library packages). This differs from venture-style (`cmd/client`, `cmd/server`, `cmd/mobile`) and vania-style (`internal/` private packages).

```
main.go                  Ebiten game loop, state machine, system initialization
config.toml              Default configuration (Viper)
cmd/
  server/                Dedicated multiplayer server
  federation-hub/        Cross-server matchmaking and DHT discovery hub
  mod-registry/          Mod distribution registry
pkg/
  engine/                ECS framework (Entity, Component, System, World)
  rng/                   Seed-based deterministic RNG (PCG algorithm)
  config/                Configuration via Viper with hot-reload
  input/                 Input manager (keyboard, mouse, gamepad, touch)
  pool/                  Memory pooling for zero-allocation hot paths
  
  # Generation Layer
  bsp/                   BSP procedural level generation, arena generator
  level/                 Level generation and tile-based maps
  procgen/genre/         Genre registry and SetGenre interface
  texture/               Procedural texture atlas
  walltex/               Enhanced wall texture generation
  sprite/                Procedural sprite generation
  props/                 Decorative prop placement
  decoration/            Room decoration, environmental storytelling
  lore/                  Procedural narrative content
  dialogue/              Procedurally generated NPC conversations
  biome/                 Biome-specific zone identification
  
  # Simulation Layer
  camera/                First-person camera (FOV, pitch, head-bob)
  raycaster/             DDA raycasting engine
  collision/             Collision detection with layer masking
  spatial/               Grid-based spatial indexing
  combat/                Damage model, combos, hit feedback, boss phases
  ai/                    Enemy behavior trees and adaptive AI
  weapon/                Weapon definitions, firing, mastery progression
  projectile/            Projectile simulation, spells
  status/                Status effects (poison, burn, bleed, radiation)
  door/                  Keycard and door system
  trap/                  Interactive trap mechanics
  hazard/                Environmental hazards
  destruct/              Destructible environments
  faction/               Faction reputation and relationships
  territory/             Dynamic faction territory control
  event/                 World events and timed triggers
  
  # Presentation Layer
  render/                Rendering pipeline, post-processing
  lighting/              Sector-based dynamic lighting with shadows
  fog/                   Atmospheric fog rendering
  particle/              Particle emitters and effects
  weather/               Environmental particle effects
  animation/             State-based sprite animation with LOD
  damagestate/           Visual damage state rendering
  decal/                 Persistent combat decals
  corpse/                Persistent corpse rendering
  outline/               Sprite silhouette rendering
  healthbar/             Overhead health bars and status icons
  telegraph/             Visual attack telegraphing
  attacktrail/           Visual weapon attack trails
  weaponanim/            Visual weapon attack animation
  dmgfx/                 Damage-type visual effects
  itemicon/              Procedural item icon generation
  floor/                 Procedural floor tile variation
  parallax/              Multi-layer parallax backgrounds
  feedback/              Visual and kinesthetic feedback
  flicker/               Physics-based torch/flame flicker
  audio/                 Procedural audio synthesis (music, SFX, ambient)
  ui/                    HUD, menus, settings screens
  
  # Game Systems
  inventory/             Item management
  equipment/             Visual rendering of equipped items
  crafting/              Scrap-to-ammo conversion
  shop/                  Between-level armory
  economy/               Configurable game economy and rewards
  loot/                  Loot tables and drops
  progression/           XP and leveling
  class/                 Character class definitions
  stats/                 Character stat allocation
  skills/                Skill and talent trees
  quest/                 Procedural objectives
  squad/                 Squad companion AI
  secret/                Push-wall secret discovery
  upgrade/               Weapon upgrade token system
  ammo/                  Ammo types and pools
  automap/               Fog-of-war automap
  minigame/              Hacking and lockpicking mini-games
  tutorial/              Context-sensitive tutorial prompts
  save/                  Cross-platform save/load, cloud sync
  
  # Multiplayer Layer
  network/               Client/server netcode, delta sync, lag compensation
  federation/            Cross-server matchmaking, DHT discovery
  chat/                  E2E encrypted in-game chat (ECDH + AES-GCM)
  leaderboard/           Local and federated score tracking
  achievements/          Local achievement tracking
  replay/                Deterministic replay recording and playback
  mod/                   Mod loader, plugin API, WASM sandboxing
  
  # Testing
  testutil/              Test helpers and mocks
  integration/           Integration test utilities
```

---

## ⚠️ CRITICAL: Complete Feature Integration (Zero Dangling Features)

**This is the single most important rule for this codebase.** Every feature, system, component, generator, and integration MUST be fully wired into the runtime. Dangling features are a maintenance burden, a source of frustration, and actively degrade code quality.

### The Dangling Feature Problem

In complex procedural game codebases, it is extremely common for features to be:
1. **Defined but never instantiated** — A system struct exists but is never created in `main()` or system registration
2. **Instantiated but never integrated** — A system runs but its output is never consumed by other systems
3. **Partially integrated** — A system works for one genre/theme but silently no-ops for others
4. **Tested in isolation but broken in context** — Unit tests pass but the system was never wired into the game loop

### Known Gaps (from GAPS.md)

The following gaps are documented and tracked. Do NOT add features that create similar patterns:

1. **Chat Key-Exchange Deadlock** (Gap 1): `pkg/chat/keyexchange.go` — `PerformKeyExchange` sends before receiving, causing deadlock on unbuffered connections
2. **Procedural Audio Untestable** (Gap 2): `pkg/audio/` — `generateLoop()` blocks indefinitely without context cancellation
3. **Federation/DHT Tests Unverifiable** (Gap 3): Tests start live libp2p hosts that require LAN gateway
4. **22 Packages Require X11** (Gap 4): Core packages panic without DISPLAY environment variable
5. **Config Watch Data Race** (Gap 5): Race between `reloadConfiguration()` and `viper.Reset()` in tests
6. **Profanity Filter Limited** (Gap 6): Basic l33t speak detection, no Unicode homoglyphs

### Mandatory Checks Before Adding or Modifying Any Feature

**Before writing ANY new code, verify the full integration chain:**

1. **Definition → Instantiation**: Is the struct/system created at runtime? Trace from `main()` through system registration.
2. **Instantiation → Registration**: Is the system registered with the ECS `World` via `AddSystem()`? Check `main.go` initialization.
3. **Registration → Update Loop**: Does the system's `Update(*World)` method get called when `World.Update()` runs?
4. **Update → Output**: Does the system produce outputs (components, events, state changes) that other systems consume?
5. **Output → Consumer**: Is there at least one other system that reads this system's output?
6. **Consumer → Player Effect**: Does the chain ultimately produce something visible, audible, or mechanically felt by the player?
7. **Genre Coverage**: Does the feature work for ALL 5 genres? Check `SetGenre()` switch statements.

If ANY link in this chain is missing, the feature is dangling. **Do not submit dangling features.**

### Specific Anti-Patterns to Reject

```go
// ❌ BAD: System defined but never added to the game world
type WeatherSystem struct { ... }
func (w *WeatherSystem) Update(world *engine.World) { ... }
// ...but NewWeatherSystem() is never called in main.go

// ✅ GOOD: System defined, instantiated, and registered
weatherSys := weather.NewSystem(seed)
world.AddSystem(weatherSys)
// AND renderSys uses weatherSys.GetCurrentWeather() for visual effects
```

```go
// ❌ BAD: Generator implements interface but never called outside tests
type CyberpunkTerrainGen struct { ... }
func (g *CyberpunkTerrainGen) Generate(seed int64, params map[string]interface{}) (interface{}, error) { ... }
// Only called in cyberpunk_terrain_test.go

// ✅ GOOD: Generator registered in genre dispatch
genRegistry["cyberpunk"] = &CyberpunkTerrainGen{}
// AND terrain := genRegistry[currentGenre].Generate(seed, params) is called in level init
```

```go
// ❌ BAD: SetGenre has incomplete genre coverage
func (s *MySystem) SetGenre(genreID string) {
    switch genreID {
    case genre.Fantasy:
        // configured
    case genre.SciFi:
        // configured
    // MISSING: Horror, Cyberpunk, PostApoc — will use zero values!
    }
}

// ✅ GOOD: All 5 genres handled with explicit default
func (s *MySystem) SetGenre(genreID string) {
    switch genreID {
    case genre.Fantasy:
        s.config = fantasyConfig
    case genre.SciFi:
        s.config = scifiConfig
    case genre.Horror:
        s.config = horrorConfig
    case genre.Cyberpunk:
        s.config = cyberpunkConfig
    case genre.PostApoc:
        s.config = postapocConfig
    default:
        s.config = fantasyConfig // explicit fallback
    }
}
```

### Seed Propagation Requirements

Seeds must flow through the full generation hierarchy. A common failure mode is accepting a seed but not propagating it to sub-generators:

```go
// ❌ BAD: Seed accepted but not propagated
func (g *LevelGenerator) Generate(seed uint64) *Level {
    rng := rng.NewRNG(seed)
    level := &Level{}
    
    // WRONG: enemyGen uses its own random seed
    level.Enemies = g.enemyGen.Generate()  // Uses internal/default seed!
    
    return level
}

// ✅ GOOD: Seed derived and propagated to all sub-generators
func (g *LevelGenerator) Generate(seed uint64) *Level {
    rng := rng.NewRNG(seed)
    level := &Level{}
    
    // Derive sub-seeds deterministically
    enemySeed := seed ^ 0x454E454D  // "ENEM"
    lootSeed := seed ^ 0x4C4F4F54   // "LOOT"
    propSeed := seed ^ 0x50524F50   // "PROP"
    
    level.Enemies = g.enemyGen.Generate(enemySeed)
    level.Loot = g.lootGen.Generate(lootSeed)
    level.Props = g.propGen.Generate(propSeed)
    
    return level
}
```

### Event Emitter/Listener Completeness

Every event emitted must have at least one listener. Every listener registered must receive events:

```go
// ❌ BAD: Event emitted with no listener
eventBus.Emit("player.levelup", playerID, newLevel)
// No eventBus.On("player.levelup", ...) anywhere in codebase

// ❌ BAD: Listener registered for event never fired
eventBus.On("boss.enraged", func(bossID string) {
    audio.PlayEnrageSound(bossID)
})
// But no code ever calls eventBus.Emit("boss.enraged", ...)

// ✅ GOOD: Event has matching emitter and listener(s)
// In progression system:
eventBus.Emit("player.levelup", playerID, newLevel)

// In UI system:
eventBus.On("player.levelup", func(id string, level int) {
    ui.ShowLevelUpAnimation(id, level)
})

// In audio system:
eventBus.On("player.levelup", func(id string, level int) {
    audio.PlayLevelUpSound()
})
```

### Integration Verification Checklist (run before every PR)

```bash
# Every constructor has at least one non-test caller
grep -rn 'func New' --include='*.go' | grep -v _test.go | head -20

# All TODOs are tracked in GAPS.md or ROADMAP.md
grep -rn 'TODO\|FIXME\|HACK\|XXX' --include='*.go' | grep -v _test.go

# No empty method bodies (excluding interface stubs)
grep -rn '{ }$\|{$' --include='*.go' -A1 | grep -v _test.go

# Verify SetGenre handles all genres
grep -rn 'func.*SetGenre' --include='*.go' | xargs grep -l 'switch'

# Check for event emitters without listeners
grep -rn 'Emit(' --include='*.go' | cut -d'"' -f2 | sort -u > /tmp/emitted.txt
grep -rn '\.On(' --include='*.go' | cut -d'"' -f2 | sort -u > /tmp/listened.txt
comm -23 /tmp/emitted.txt /tmp/listened.txt  # Shows emitted but not listened

# Verify all Generator implementations are registered
grep -rn 'type.*Generator struct' --include='*.go' | grep -v _test.go
# Cross-reference with generator registrations in main.go or init functions
```

### Documentation-Code Drift Prevention

When adding features, ensure documentation matches implementation:

1. **README.md**: Update if adding new CLI flags, config options, or user-facing features
2. **GAPS.md**: Add entry if the feature has known limitations
3. **docs/*.md**: Update relevant documentation for subsystem changes
4. **ROADMAP.md**: Mark completed items, add new planned work
5. **godoc comments**: Every exported type and function needs documentation

---

## Networking Best Practices (MANDATORY for all Go network code)

### Interface-Only Network Types (Hard Constraint)

When declaring network variables, ALWAYS use interface types. This enhances testability and flexibility.

| ❌ Never Use (Concrete Type) | ✅ Always Use (Interface Type) |
|---|----|
| `*net.UDPAddr` | `net.Addr` |
| `*net.IPAddr` | `net.Addr` |
| `*net.TCPAddr` | `net.Addr` |
| `*net.UDPConn` | `net.PacketConn` |
| `*net.TCPConn` | `net.Conn` |
| `*net.TCPListener` | `net.Listener` |

```go
// ✅ GOOD: Interface types everywhere
var addr net.Addr
var conn net.Conn
var listener net.Listener

// ❌ BAD: Concrete types
var addr *net.UDPAddr
var conn *net.TCPConn
var listener *net.TCPListener
```

**Note**: One exception exists in `pkg/network/gameserver.go:addClient()` which uses a type assertion for deadline setting. This pattern should be avoided in new code — use interface methods instead.

### High-Latency Network Design (200–5000ms)

All multiplayer networking code MUST be designed to function correctly under **200–5000ms round-trip latency**. The network layer already implements:

- **Server-authoritative model**: All game state mutations happen server-side after command validation
- **20 Hz tick rate**: Server processes commands and broadcasts deltas every 50ms
- **Delta synchronization**: `DeltaEncoder` transmits only changed entity fields using XOR and bitmasks
- **500ms lag compensation**: Server stores 10 snapshots for hitscan rewind
- **Interpolation buffer**: Client interpolates between known server states

#### Latency Thresholds and Behavior

| Threshold | Client Behavior | Server Behavior |
|-----------|-----------------|-----------------|
| 0–100ms | Optimal gameplay | Normal processing |
| 100–200ms | Interpolation delay increases | Normal processing |
| 200–500ms | Noticeable prediction | Commands accepted |
| 500–2000ms | Heavy prediction, rubber-banding possible | Stale commands rejected |
| 2000–5000ms | Minimal updates, spectator-like | Player marked AFK |
| >5000ms | Disconnected | Connection dropped |

#### Mandatory Design Principles for New Network Code

1. **Client-Side Prediction**: Simulate locally, reconcile when server state arrives. Never block game loop on network.

2. **No Synchronous RPC in Game Loops**: Never issue blocking network calls in `Update()` or `Draw()`.

```go
// ❌ BAD: Blocking RPC in game loop
func (g *Game) Update() error {
    state, err := g.server.GetWorldState()  // BLOCKS
    g.world = state
    return nil
}

// ✅ GOOD: Async receive with interpolation
func (g *Game) Update() error {
    select {
    case state := <-g.stateChannel:
        g.interpolator.PushServerState(state)
    default:
        // No new state — continue with prediction
    }
    g.world = g.interpolator.GetInterpolatedState(time.Now())
    return nil
}
```

3. **Generous Timeouts**: Minimum 10 seconds for connection timeouts.

```go
// ❌ BAD: Tight timeout drops satellite players
conn.SetReadDeadline(time.Now().Add(1 * time.Second))

// ✅ GOOD: Generous timeout for high-latency
conn.SetReadDeadline(time.Now().Add(10 * time.Second))
```

4. **Idempotent Messages**: Every message must be safe to process multiple times due to retransmission.

5. **Graceful Degradation**: At 5000ms latency, game remains playable — reduce update frequency, increase prediction windows.

6. **Heartbeat-Based Disconnect Detection**: Never disconnect on a single missed packet. Use sliding window (≥3 missed heartbeats).

```go
// ❌ BAD: Single packet timeout
if time.Since(lastPacket) > 1*time.Second {
    disconnect(player)
}

// ✅ GOOD: Sliding window heartbeat detection
const heartbeatInterval = 2 * time.Second
const missedHeartbeatThreshold = 3

if missedHeartbeats >= missedHeartbeatThreshold {
    disconnect(player)
}
```

#### Latency Budget Allocation (per frame at 60 FPS = 16.6ms)

| Task | Budget |
|------|--------|
| Input processing | ≤1ms |
| Local simulation / prediction | ≤4ms |
| State interpolation | ≤1ms |
| Network send (non-blocking enqueue) | ≤0.5ms |
| Rendering | ≤10ms |
| Network I/O goroutines | Independent (not counted against frame budget) |

---

## Code Assistance Guidelines

### 1. Deterministic Procedural Generation

All content generation MUST be deterministic and seed-based. The project uses `pkg/rng` with Go 1.22+ PCG algorithm.

```go
// ✅ GOOD: Use pkg/rng for all generation
import "github.com/opd-ai/violence/pkg/rng"

rng := rng.NewRNG(seed)
value := rng.Intn(100)
floatVal := rng.Float64()

// ✅ GOOD: Derived seeds for sub-generators (deterministic hierarchy)
terrainSeed := seed ^ 0x54455252  // "TERR"
enemySeed := seed ^ 0x454E454D    // "ENEM"
terrainRNG := rng.NewRNG(terrainSeed)
enemyRNG := rng.NewRNG(enemySeed)

// ❌ BAD: Global math/rand (non-deterministic, not thread-safe)
value := rand.Intn(100)

// ❌ BAD: Time-based seeding in generation code
rng := rand.New(rand.NewSource(time.Now().UnixNano()))
```

### 2. ECS Architecture (`pkg/engine`)

The engine uses a minimal Entity-Component-System pattern:

- **Entity** (`uint64`): Unique identifier for game objects
- **Component** (`interface{}`): Data attached to entities, stored by `reflect.Type`
- **System** (`interface{ Update(w *World) }`): Logic that processes entities
- **World**: Container holding all entities, components, and systems

#### Key ECS Operations

| Operation | Method | Complexity |
|-----------|--------|------------|
| Create entity | `World.AddEntity()` | O(1) |
| Attach component | `World.AddComponent(entity, component)` | O(1) |
| Get component | `World.GetComponent(entity, type)` | O(1) |
| Check component | `World.HasComponent(entity, type)` | O(1) |
| Remove component | `World.RemoveComponent(entity, type)` | O(1) |
| Remove entity | `World.RemoveEntity(entity)` | O(1) |
| Query by components | `World.Query(types...)` | O(n) where n = total entities |
| Run all systems | `World.Update()` | Iterates systems in registration order |

```go
// Creating entities and components
world := engine.NewWorld()
entity := world.AddEntity()
world.AddComponent(entity, &PositionComponent{X: 100, Y: 200})
world.AddComponent(entity, &HealthComponent{Current: 100, Max: 100})

// Querying entities by component type
posType := reflect.TypeOf(&PositionComponent{})
healthType := reflect.TypeOf(&HealthComponent{})
entities := world.Query(posType, healthType)

// System implementation
type MovementSystem struct{}
func (s *MovementSystem) Update(w *engine.World) {
    for _, e := range w.Query(posType, velocityType) {
        pos, _ := w.GetComponent(e, posType)
        vel, _ := w.GetComponent(e, velocityType)
        // Update position based on velocity
    }
}

// Genre-aware systems
world.SetGenre("scifi")
genre := world.GetGenre()  // "scifi"
```

#### ECS Best Practices

1. **Components are pure data**: No methods with game logic. Components may have helper methods for serialization or validation only.
2. **Systems contain all logic**: Query for entities with required components, process them, update component data.
3. **Never store entity references directly**: Use entity IDs (`uint64`). Entities may be removed between frames.
4. **Declare dependencies explicitly**: Systems should document which component types they require.
5. **Process in deterministic order**: Systems run in registration order. This matters for determinism.

### 3. Structured Logging (logrus)

Use structured logging with context fields throughout the codebase.

```go
// ✅ GOOD: Structured logging with context
logrus.WithFields(logrus.Fields{
    "system":   "terrain",
    "seed":     seed,
    "biome":    biomeType,
    "duration": elapsed,
}).Info("Terrain generation complete")

// ✅ GOOD: Error logging with context
logrus.WithFields(logrus.Fields{
    "system": "network",
    "player": playerID,
}).WithError(err).Error("Failed to process command")

// ❌ BAD: Unstructured logging
fmt.Printf("Generated terrain with seed %d\n", seed)
log.Println("terrain done")
```

Standard field names: `system`, `entity`, `player`, `seed`, `error`, `duration`, `count`, `genre`.

### 4. Performance Requirements

- Target 60 FPS on mid-range hardware
- Client memory budget: <500MB
- Use spatial partitioning (`pkg/spatial`) for entity queries over collections >100
- Cache generated sprites — never regenerate the same sprite twice per session
- Use object pooling (`pkg/pool`) for frequently allocated/deallocated objects
- Benchmark hot paths with `go test -bench=. -benchmem`

### 5. Zero External Assets

The single-binary philosophy means ALL content is generated at runtime:

- **Graphics**: Procedurally generated via pixel manipulation, shape primitives, noise functions
- **Audio**: Synthesized from oscillators, envelopes, and effects (`pkg/audio`)
- **Levels/Maps**: BSP, cellular automata, arena generators (`pkg/bsp`, `pkg/level`)
- **Items/NPCs/Quests**: Parameterized templates with seed-based variation
- **UI**: Built from code, not loaded from image files

**Never add asset files (PNG, WAV, OGG, JSON level files) to the repository.**

### 6. Error Handling

```go
// ✅ GOOD: Return errors, handle at call site
func GenerateTerrain(seed uint64) (*Terrain, error) {
    if seed == 0 {
        return nil, fmt.Errorf("terrain generation requires non-zero seed")
    }
    // ...
}

// ✅ GOOD: Log and recover gracefully in game systems
terrain, err := GenerateTerrain(seed)
if err != nil {
    logrus.WithError(err).Error("Terrain generation failed, using fallback")
    terrain = defaultTerrain()
}

// ❌ BAD: Panic in library/game code
if seed == 0 {
    panic("zero seed")  // Never panic in game logic
}
```

Panics are acceptable ONLY in `main()` for unrecoverable startup failures.

### 7. Genre System (`SetGenre` Convention)

When a package adapts to genre, implement `SetGenre(genreID string)`:

```go
import "github.com/opd-ai/violence/pkg/procgen/genre"

func (s *MySystem) SetGenre(genreID string) {
    switch genreID {
    case genre.Fantasy:   // "fantasy"
        s.config = fantasyConfig
    case genre.SciFi:     // "scifi"
        s.config = scifiConfig
    case genre.Horror:    // "horror"
        s.config = horrorConfig
    case genre.Cyberpunk: // "cyberpunk"
        s.config = cyberpunkConfig
    case genre.PostApoc:  // "postapoc"
        s.config = postapocConfig
    default:
        s.config = fantasyConfig
    }
}
```

#### Genre Effects by Subsystem

| Subsystem | Fantasy | Sci-Fi | Horror | Cyberpunk | Post-Apoc |
|-----------|---------|--------|--------|-----------|-----------|
| Wall Tiles | Stone (10) | Hull (11) | Plaster (12) | Concrete (13) | Rust (14) |
| Floor Tiles | Stone (20) | Hull (21) | Wood (22) | Concrete (23) | Dirt (24) |
| Color Grade | Warm | Cool blue | Desaturated | Neon tint | Sepia |
| Vignette | Moderate | Light | Heavy | Moderate | Heavy |
| Film Grain | None | None | Heavy | Light | Moderate |
| Bloom | Light | Moderate | None | Heavy | None |
| Music Style | Orchestral | Synth | Ambient horror | Synthwave | Industrial |
| Control Point | Altar | Terminal | Summoning circle | Server rack | Scrap pile |

### 8. Testing Patterns

#### Table-Driven Tests

```go
func TestDamageCalculation(t *testing.T) {
    tests := []struct {
        name     string
        baseDmg  int
        armor    int
        expected int
    }{
        {"no armor", 100, 0, 100},
        {"half armor", 100, 50, 50},
        {"full armor", 100, 100, 0},
        {"over armor", 100, 150, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateDamage(tt.baseDmg, tt.armor)
            if got != tt.expected {
                t.Errorf("CalculateDamage(%d, %d) = %d, want %d",
                    tt.baseDmg, tt.armor, got, tt.expected)
            }
        })
    }
}
```

#### Determinism Tests

```go
func TestGeneratorDeterminism(t *testing.T) {
    seed := uint64(12345)
    
    // Generate twice with same seed
    result1 := generator.Generate(seed)
    result2 := generator.Generate(seed)
    
    // Must be identical
    if !reflect.DeepEqual(result1, result2) {
        t.Errorf("Generator not deterministic: different results for same seed")
    }
}
```

#### Benchmark Hot Paths

```go
func BenchmarkRaycast(b *testing.B) {
    rc := raycaster.New(config)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        rc.CastRays(camera)
    }
}
```

---

## Cross-Repository Code Sharing Patterns

### Shared Pattern Catalog

When implementing features, follow these patterns for future extraction into shared packages:

| Pattern | Package Convention | Notes |
|---------|-------------------|-------|
| ECS core (World, Entity, Component, System) | `pkg/engine/` | System interface: `Update(w *World)` |
| Seed-based RNG | `pkg/rng/` | Use PCG algorithm, `uint64` seeds |
| Procedural generation framework | `pkg/procgen/` | Genre registry in `procgen/genre/` |
| Sprite/tile generation | `pkg/sprite/`, `pkg/texture/` | — |
| Audio synthesis | `pkg/audio/` | — |
| Input handling | `pkg/input/` | Keyboard, mouse, gamepad, touch |
| Camera systems | `pkg/camera/` | — |
| Particle systems | `pkg/particle/` | — |
| Save/load | `pkg/save/` | Cross-platform, cloud sync |
| Menu/UI framework | `pkg/ui/` | — |
| Configuration (viper) | `pkg/config/` | Hot-reload via fsnotify |
| Networking (multiplayer) | `pkg/network/`, `pkg/federation/` | — |
| Collision detection | `pkg/collision/` | Layer masking |

### Guidelines for Shareable Code

1. **Keep dependencies minimal**: Shared packages depend only on stdlib + Ebiten
2. **Use interfaces at boundaries**: Define interfaces for game-specific behavior
3. **Parameterize, don't specialize**: Generators accept parameters for any genre
4. **Same naming conventions across repos**: Consistent method signatures for future extraction
5. **Identical ECS interface**: `System.Update(w *World)` across all repos

### When Adding a Feature That Exists in a Sibling Repo

1. Check the sibling repo's implementation first
2. Use the same package structure and naming conventions
3. Match interface signatures for future extraction
4. If the sibling implementation has known issues (check its GAPS.md), fix them
5. Document divergences in ROADMAP.md with a note about future convergence

---

## Quality Standards

### Testing Requirements

- **Coverage**: ≥82% enforced in CI (≥40% per package target)
- **Table-driven tests** for all business logic and generation functions
- **Benchmarks** for hot-path code (rendering, physics, generation)
- **Race detection**: All tests must pass under `go test -race ./...`
- **Headless testing**: Linux CI uses `xvfb-run` for packages requiring X11

```bash
# Run all tests (Linux with xvfb)
xvfb-run -a -- go test -v -race -coverprofile=coverage.out ./...

# Run headless-safe tests only
go test -race -cover ./pkg/bsp/... ./pkg/raycaster/... ./pkg/texture/... \
  ./pkg/replay/... ./pkg/leaderboard/... ./pkg/achievements/... \
  ./pkg/quest/... ./pkg/rng/... ./pkg/config/... ./pkg/inventory/...
```

### Code Review Quality Gates

- Build success (client + server)
- All tests pass with `-race`
- `go vet ./...` clean
- `gofmt` compliance
- No new TODO/FIXME without corresponding GAPS.md entry
- Integration chain verified for new features

### Documentation Requirements

- Every exported type and function has a godoc comment
- README.md stays in sync with CLI flags and features
- GAPS.md updated when new gaps are discovered
- ROADMAP.md reflects current priorities
- AUDIT.md records functional audit findings

---

## Naming Conventions

- **Packages**: lowercase, single-word when possible (`engine`, `procgen`, `audio`, `render`)
- **Files**: snake_case (`terrain_generator.go`, `combat_system.go`)
- **Types**: PascalCase (`TerrainGenerator`, `CombatSystem`, `HealthComponent`)
- **Interfaces**: PascalCase, often `-er` suffix for single-method (`Generator`, `Renderer`)
- **Component types**: PascalCase (no suffix required, context is clear from usage)
- **System types**: PascalCase + "System" suffix when implementing `System` interface
- **Constants**: PascalCase for exported, camelCase for unexported
- **Seeds**: Always `uint64`, always named `seed` in function parameters
- **Genre IDs**: lowercase strings (`"fantasy"`, `"scifi"`, `"horror"`, `"cyberpunk"`, `"postapoc"`)

---

## GAPS.md and AUDIT.md Protocol

These repos use GAPS.md and AUDIT.md to track implementation gaps and audit findings.

### When Copilot Identifies a Potential Gap

1. Note it in your response
2. Suggest adding to GAPS.md with severity (Critical/High/Medium/Low)
3. Include file path and line number
4. Propose an actionable fix

### Gap Entry Format

```markdown
## Gap N: [Short Title]

- **Stated Goal**: What the feature claims to do
- **Current State**: What actually happens, with file:line reference
- **Impact**: User-facing or developer-facing consequences
- **Closing the Gap**: Specific fix with estimated effort
```

---

## Build and Run Commands

```bash
# Build client
go build -o violence .

# Run client
./violence

# Build and run dedicated server
go build -o violence-server ./cmd/server
./violence-server -port 7777 -log-level info

# Build and run federation hub
go build -o federation-hub ./cmd/federation-hub
./federation-hub -addr :8080 -log-level info

# Run tests with race detection (Linux)
xvfb-run -a -- go test -v -race -coverprofile=coverage.out ./...

# Run tests (macOS/Windows)
go test -v -race -coverprofile=coverage.out ./...

# Check coverage threshold
go tool cover -func=coverage.out | grep total

# Static analysis
go vet ./...
gofmt -l .

# WASM build
GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -v -o violence.wasm .
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" .
```

---

## Configuration

Configuration is loaded from `config.toml` (working directory or `$HOME/.violence/config.toml`) via Viper with hot-reload support.

```toml
WindowWidth = 1280
WindowHeight = 800
InternalWidth = 320        # Raycaster internal resolution
InternalHeight = 200
FOV = 66.0
MouseSensitivity = 1.0
MasterVolume = 0.8
MusicVolume = 0.7
SFXVolume = 0.8
DefaultGenre = "fantasy"   # fantasy, scifi, horror, cyberpunk, postapoc
VSync = true
FullScreen = false
MaxTPS = 60
FederationHubURL = ""      # Optional federation hub for server discovery
```

---

## Mod Development

Mods use the WASM-sandboxed plugin API (`pkg/mod`). See `docs/MODDING.md` and `docs/MODDING_WASM.md` for details.

```go
// Plugin interface
type Plugin interface {
    Load() error
    Unload() error
    Name() string
    Version() string
}

// Generator interface for custom procedural content
type Generator interface {
    Type() string
    Generate(seed int64, params map[string]interface{}) (interface{}, error)
}
```

**Mod determinism requirement**: All generators must use `rand.New(rand.NewSource(seed))` with the provided seed parameter. Never use `time.Now()` or global `math/rand`.

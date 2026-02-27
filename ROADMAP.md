# VIOLENCE

**Gameplay Style:** Raycasting First-Person Shooter (Wolfenstein-style corridors)

**Vision:** A fully-featured, procedurally-generated raycasting FPS where five thematic genres skin identical corridor-combat mechanics — from torchlit dungeons to ruined bunkers — with co-op/deathmatch multiplayer and 82%+ test coverage, shipped as a single deterministic binary. **100% of gameplay assets — including all audio, visual, and narrative/story-driven components — are procedurally generated at runtime using deterministic algorithms.** No pre-rendered, embedded, or bundled audio files (e.g., .mp3, .wav, .ogg), visual/image files (e.g., .png, .jpg, .svg, .gif), or static narrative content (e.g., hardcoded dialogue, pre-written cutscene scripts, fixed story arcs, embedded text assets) are permitted.

---

## Genre Support

Every system implements `SetGenre(genreID string)` to swap thematic presentation without changing gameplay logic.

| Genre ID    | Violence Manifestation                                                                                  |
|-------------|--------------------------------------------------------------------------------------------------------|
| `fantasy`   | Torchlit stone dungeons — castle corridors, magical barriers, enchanted weapons, rune-locked doors      |
| `scifi`     | Space station corridors — sliding bulkhead doors, laser/plasma weapons, viewport windows to space       |
| `horror`    | Abandoned asylum — flickering lights, blood trails, sanity effects, environmental jump-scares           |
| `cyberpunk` | Corporate tower — glass walls, holographic doors, energy weapons, hacking mini-games, security systems  |
| `postapoc`  | Ruined bunker — collapsed passages, improvised weapons, radiation hazards, scavenging scarcity          |

---

## Phased Milestones

### v1.0 — Core Engine + Playable Single-Player

*Goal: one fully playable genre (fantasy) end-to-end on all target platforms.*

#### ECS Framework
- Entity-Component-System core (entities, components, systems, world)
- Component registration and query API
- System execution order and dependency graph
- `SetGenre(genreID string)` interface required on every new system from day one

#### Seed-Based Deterministic RNG
- Seeded PRNG for reproducible level generation and loot
- Per-run seed storage for replays and bug reproduction

#### Input System
- WASD movement + mouselook (raw mouse input)
- Keyboard bindings for weapon select, use/interact, automap, pause
- Gamepad support (analog stick mouselook, trigger fire)
- Configurable key remapping (`pkg/input`)

#### Raycaster Engine *(violence-specific)*
- Wall casting: DDA ray-march against tile grid, column-height projection
- Floor/ceiling casting: per-pixel perspective-correct floor and ceiling
- Sprite casting: billboard sprites sorted by depth, depth-clipped against walls
- Distance fog: genre-colored exponential fog for atmosphere and perf
- Head-bob: camera Y-offset sinusoid synced to player speed

#### First-Person Camera System
- Field-of-view (FOV) and aspect-ratio settings
- Pitch clamping (no free-look beyond ±30°)
- Head-bob integration (see raycaster)

#### BSP Level Generator *(violence-specific)*
- Binary Space Partitioning recursive room/corridor splitting
- Room placement with configurable min/max size, density
- Corridor carving between leaf nodes
- Secret wall placement (push-walls that conceal rewards)
- Keycard-locked door insertion at chokepoints
- Per-genre tile theme assignment via `SetGenre()`

#### Rendering Pipeline
- Raycaster → framebuffer → Ebiten `*ebiten.Image` blit
- Palette/shader swap per genre via `SetGenre()`
- 320×200 internal resolution scaled to window

#### Audio — Core
- Adaptive music engine: base track + intensity layers, all procedurally generated at runtime via deterministic synthesis (no embedded or bundled audio files)
- SFX: gunshot, footstep, door, pickup, enemy alert, death — all procedurally synthesized at runtime from seed-driven algorithms
- 3D positional audio: distance attenuation + left/right pan for spatial awareness
- Genre audio theme swap via `SetGenre()` (`pkg/audio`)

#### UI / HUD / Menus
- HUD: health bar, armor bar, ammo counter, weapon icon, keycard inventory
- Main menu, difficulty select, genre select
- Pause menu with resume/save/quit
- Settings screen (video, audio, key bindings)
- Loading screen with seed display

#### Tutorial System
- First-level contextual prompts (WASD, shoot, pick up, open door)
- Suppressible after first completion

#### Save / Load
- Per-slot save: level seed, player state, discovered map, inventory
- Auto-save on level exit
- Cross-platform save path (`pkg/save`)

#### Config / Settings
- INI/TOML config file with hot-reload
- Resolution, FOV, mouse sensitivity, volume levels, genre default
- `pkg/config`

#### Performance — Raycaster Optimizations
- Sprite depth-sort (painter's algorithm)
- Lookup tables for sin/cos/tan
- Occlusion column tracking (skip sprites behind solid walls)
- Frame-rate cap and VSync toggle

#### CI/CD — Foundation
- GitHub Actions: build + test on Linux, macOS, Windows
- Single-binary `go build` with zero external assets (all assets procedurally generated at runtime; no embedded, bundled, or pre-rendered media files)
- 82%+ test coverage gate

---

### v2.0 — Core Systems: Weapons, FPS AI, Keycards, All 5 Genres

*Goal: all five genres playable with full combat loop and progression.*

#### Weapon System *(violence-specific)*
- Hitscan weapons: pistol, shotgun, chaingun — instant ray-cast hit detection
- Projectile weapons: rocket launcher, plasma gun — simulated in-world projectiles
- Melee backup: knife (silent), fist (desperation)
- Weapon wheel / quick-select (1–7 keys)
- Per-genre weapon skin swap via `SetGenre()` (pistol → blaster, shotgun → scatter-cannon, etc.)
- Weapon animation: raise, lower, fire, reload frames — all procedurally generated at runtime

#### Ammo System *(violence-specific)*
- Per-weapon ammo types: bullets, shells, cells, rockets
- Ammo pickups: boxes, backpacks, dropped from enemies
- Scarcity tuning per difficulty
- Ammo display on HUD

#### Keycard / Door System *(violence-specific)*
- Color-coded keycards (red, yellow, blue, plus genre variants: rune-key, access-card, biometric)
- Door types: standard swing, sliding bulkhead (scifi), portcullis (fantasy), shutter (postapoc), laser-barrier (cyberpunk/horror)
- Locked-door HUD feedback ("Need blue keycard")
- `SetGenre()` maps canonical color → genre door/key skin

#### Automap *(violence-specific)*
- Fog-of-war: tiles reveal on first sight
- Wall, door, locked-door, secret annotations
- Player position + facing indicator
- Overlay (Tab) and fullscreen modes

#### FPS Enemy AI
- Behavior tree nodes: patrol, idle, alert, chase, strafe, take-cover, retreat
- Line-of-sight + hearing (gunshot radius wakes nearby enemies)
- Enemy archetypes per genre (guard/soldier/cultist/drone/scavenger)
- `SetGenre()` swaps archetype visuals and audio, all procedurally generated at runtime

#### Combat System
- Damage model: health, armor absorption
- Hit feedback: screen flash, HUD damage indicators (directional)
- Enemy death states, gibs (difficulty-gated)
- Difficulty scaling: enemy HP, damage, count, alertness radius

#### Status Effects
- Poisoned (fantasy: cursed blade), Burning (scifi/cyberpunk: plasma splash), Bleeding (horror/postapoc: melee), Irradiated (postapoc: radiation zones)
- Per-effect HUD icon + screen tint
- `SetGenre()` remaps effect names and visuals

#### Loot / Drops
- Ammo drops from enemies (type scales to enemy weapon)
- Health packs: small (+10), large (+25), medkit (+50)
- Armor shards and vests
- Keycard drops from boss/captain enemies

#### Progression — XP / Leveling
- XP from kills, secrets found, objectives completed
- Level-up grants: max HP increase, max armor increase, ammo capacity increase
- Per-run progression (resets each run); persistent unlocks optional in v4.0

#### Character Classes — Starting Loadout
- Grunt: assault rifle + extra ammo
- Medic: pistol + extra health packs
- Demo: rocket launcher + less ammo
- Mystic: energy weapon (plasma pistol in base; genre-skinned) + passive ability charges that deal ranged AoE damage; lower max HP, higher energy ammo capacity
- Per-genre flavor names via `SetGenre()` (Grunt → Warrior/Marine/Survivor/Enforcer/Scavenger; Mystic → Mage/Psi-Ops/Occultist/Netrunner/Chemist)

#### Genre Integration — All 5
- All systems wired to `SetGenre()` for all five genre IDs
- Genre select on new-game screen
- Per-genre BSP tile themes, enemy rosters, weapon skins, audio, color palette

---

### v3.0 — Visual Polish: Textures, Lighting, Particles, Indoor Weather

*Goal: genre-distinct atmosphere; each genre feels visually unique.*

#### Texture Mapping
- Procedurally generated wall textures per genre (stone, hull, plaster, concrete, rust) — all textures synthesized at runtime from deterministic algorithms, no pre-rendered image files
- Animated textures: flickering torches (fantasy), blinking panels (scifi), dripping water (horror) — all procedurally generated at runtime
- Floor/ceiling texture support in raycaster floor-cast pass
- `SetGenre()` selects active texture generation parameters

#### Dynamic Lighting — Sector-Based
- Per-sector ambient light level
- Point light sources: torches, lamps, monitors, fires
- Flashlight: cone-shaped forward light, genre-skinned (torch/headlamp/glow-rod)
- `SetGenre()` sets base ambient level and light source parameters (all visuals procedurally generated, no bundled sprite image files)

#### Particles System
- Gunshot muzzle flash, bullet spark, blood splatter
- Explosion debris (rockets), energy discharge (plasma)
- Indoor weather particles: dripping water, dust motes, steam vents, ash fall
- `SetGenre()` routes particle themes

#### Indoor Weather / Atmosphere *(replaces outdoor weather)*
- Fantasy: dripping water, torch smoke, fog wisps
- Scifi: vent steam, coolant spray, hull breach sparks
- Horror: flickering lights (random duration), blood drip, mold spores
- Cyberpunk: holographic static, neon haze, electrical crackle
- Postapoc: dust particles, radiation shimmer, debris chunks
- All via `SetGenre()` + particle system

#### Genre Post-Processing Presets
- Fantasy: warm sepia vignette, film grain
- Scifi: cold blue scanline overlay, chromatic aberration
- Horror: desaturated green tint, heavy vignette, occasional static burst
- Cyberpunk: neon bloom, magenta/cyan channel split
- Postapoc: washed-out orange dust filter, scratches

#### Sound Polish
- Positional audio tuning pass: echo/reverb per room size (procedurally computed from level geometry)
- Genre ambient soundscapes procedurally synthesized at runtime (dungeon echo, station hum, hospital silence, server room drone, wind) — no pre-recorded audio loops or bundled sound files
- Reload sounds, empty-click, weapon pickup jingle — all procedurally generated via deterministic synthesis

---

### v4.0 — Gameplay Expansion: Secrets, Upgrades, Squad AI, Storytelling

*Goal: depth and replayability — skill expression, narrative texture, secrets.*

#### Secret Walls *(violence-specific)*
- Push-wall mechanic: use-key on marked wall segments slides them open
- Secret area rewards: rare ammo, powerful weapons, lore items
- Automap marks discovered secrets post-reveal
- BSP generator places secrets with configurable density

#### Weapon Upgrades / Mastery
- Upgrade tokens found on levels or bought at armory
- Upgrades: damage, fire rate, clip size, accuracy, range
- Weapon mastery XP track unlocks passive bonuses (headshot damage, reload speed)
- `SetGenre()` renames upgrades (enchantment/calibration/augmentation/modification/retrofit)

#### Survival Skills — Skill / Talent Trees
- Three trees: Combat, Survival, Tech
- Nodes: faster reload, armor efficiency, sprint stamina, stealth movement
- Points from leveling; persistent across a run

#### Inventory / Items
- Extended inventory: grenades, proximity mines, medkits (active use)
- Key items: keycards, quest items, collectible lore objects
- Item-use keybind with quick-use slot

#### Crafting — Ammo from Scrap
- Scrap drops from enemies and environment
- Crafting menu: convert scrap → ammo types, health packs, grenades
- `SetGenre()` renames scrap (bone chips/circuit boards/flesh/data shards/salvage)

#### Level Objectives / Quests
- Per-level objective procedurally generated at runtime: find exit, retrieve item, destroy target, rescue hostage — all quest content, dialogue, and descriptions deterministically generated from seed
- Objective tracker on HUD
- Bonus objectives for extra XP: secret count, kill count, speed run
- `SetGenre()` configures objective text generation parameters

#### Shop / Armory (Between Levels)
- Between-run armory screen: spend credits on weapons, ammo, upgrades, items
- Credits earned from kills, secrets, objectives
- `SetGenre()` skins armory (merchant tent / supply depot / black market / corpo shop / scrap trader)

#### Squad / Companion AI
- 1–3 squad members follow player, assist in combat
- Squad command: hold position, follow, attack target
- Squad member classes mirror player classes
- `SetGenre()` skins squad members per genre

#### Environmental Storytelling
- Lore notes, audio logs, graffiti, body arrangements placed by BSP generator — all content procedurally generated at runtime from deterministic algorithms (no pre-authored text, no embedded narrative assets)
- Contextual flavor text procedurally generated on approach using seed-driven text generation
- `SetGenre()` configures genre-appropriate generation parameters and templates

#### Collectible Logs / Books / Lore
- Collectible lore items with world backstory procedurally generated at runtime from deterministic algorithms — no pre-written or embedded text assets
- Codex screen accessible from pause menu
- `SetGenre()` configures lore voice and content generation parameters

#### Hacking / Lockpicking Mini-Games
- Cyberpunk doors: circuit trace hacking puzzle
- Fantasy doors (alternate): lockpick tension mini-game
- Scifi/postapoc doors: bypass code entry
- Skip option costs consumable item

#### Destructible Environments
- Breakable wall sections reveal hidden passages (reward secrets)
- Destructible barrels and crates (explosion chain)
- Debris blocks corridors temporarily
- `SetGenre()` sets material visuals (stone rubble/hull shards/plaster/glass/concrete), all procedurally generated

#### World Events / Timed Triggers
- Alarm trigger: enemies go to alert state, doors lock for N seconds
- Timed lockdown: find exit before timer expires
- Boss arena event: spawn wave on entry
- `SetGenre()` flavors event text and audio stings (all procedurally generated at runtime)

#### Props / Decoration
- Non-interactive sprites: barrels, crates, tables, terminals, bones, plants
- BSP generator places props at genre-appropriate density
- `SetGenre()` selects prop generation parameters (all props procedurally generated at runtime, no bundled sprite image files)

---

### v5.0+ — Multiplayer, Social Features, Production Polish

*Goal: co-op and deathmatch multiplayer; production-grade release.*

#### Networking — Client/Server Netcode
- Authoritative server model; target <200 ms for optimal play, up to 500 ms graceful degradation, 5000 ms maximum tolerated (spectator/reconnect fallback, matching venture's netcode range)
- Delta-state synchronization for player positions, projectiles, pickups
- Lag compensation for hitscan (server-side rewind)
- `pkg/network` portable from venture

#### Co-op Mode
- 2–4 player co-op: shared level, independent inventories, shared objective
- Respawn on teammate side after 10s bleed-out
- Squad commands extended to human players

#### Deathmatch Mode
- Free-for-all and team deathmatch
- Deathmatch-specific BSP maps (arena layout, weapon spawn pads)
- Kill feed, score HUD, end-of-match scoreboard

#### Territory Control (Deathmatch Variant)
- Control points in corridors; team holds region for score ticks
- Genre-flavored control point visuals

#### E2E Encrypted Chat
- In-game text chat with end-to-end encryption
- Server-side relay; no plaintext message storage
- Profanity filter toggle

#### Squads / Clans
- Squad groups of up to 8 players
- Squad chat channel, shared squad statistics
- Squad tag display in HUD nameplate

#### Federation / Cross-Server Matchmaking
- Federation protocol for cross-server player lookup
- Matchmaking across federated server pool
- `pkg/federation` portable from venture

#### CI/CD — Full Production
- Multi-platform builds: Linux (amd64/arm64), macOS (universal), Windows (amd64), WASM (browser), mobile (iOS/Android via gomobile)
- Docker image for dedicated server
- Binary signing (GPG + notarization on macOS)
- Release automation: tag → draft release + artifact upload

#### Documentation Suite
- `CHANGELOG.md` — semver changelog
- `CONTROLS.md` — full keybinding reference
- `FAQ.md` — common issues and answers
- `docs/` — architecture overview, genre system guide, modding guide

#### Mod Framework
- Plugin interface: custom enemy types, weapon definitions, genre themes
- Mod loader with conflict detection
- All mod content must be defined as procedural generation parameters or algorithms — no bundled PNG, WAV, or other pre-rendered asset files permitted. Mods extend or override generation rules, not static assets.
- `pkg/mod` API

#### Test Coverage
- 82%+ line coverage enforced in CI
- Integration tests: full level generate → play → exit loop
- Multiplayer integration tests: simulated 4-player session

---

## Excluded Venture Features

| Feature                   | Rationale                                                                              |
|---------------------------|----------------------------------------------------------------------------------------|
| Vehicles                  | Indoor corridor map geometry is too tight for vehicle navigation                       |
| Reputation / Alignment    | Kill-everything corridor FPS has no meaningful NPC faction state to track              |
| Fluid Dynamics            | Raycaster tile grid cannot represent free-flowing fluid volumes                        |
| Building / Housing        | Player-constructed structures have no place in procedural locked corridors             |
| Trading (player↔player)   | Replaced by the between-level armory shop; peer trading adds no corridor-FPS value     |
| Mail System               | No persistent world between sessions that would motivate asynchronous messaging        |
| Emotes                    | Third-person animations are invisible in a first-person view                           |

---

## Shared Infrastructure — Portable Venture Packages

The following packages are architecture-neutral and can be copied or imported directly from the venture codebase:

| Package                 | Venture Path             | Usage in Violence                                              |
|-------------------------|--------------------------|----------------------------------------------------------------|
| ECS framework           | `pkg/engine`             | Entity-Component-System core, identical API                    |
| Genre registry          | `pkg/procgen/genre`      | `SetGenre()` interface + genre ID constants                    |
| Seed-based RNG          | `pkg/rng`                | Deterministic PRNG for level gen and loot                      |
| Audio engine            | `pkg/audio`              | Adaptive music, SFX, 3D positional audio                       |
| Input manager           | `pkg/input`              | Key/mouse/gamepad bindings, remapping                          |
| Config / settings       | `pkg/config`             | INI/TOML config, hot-reload                                    |
| Save / load             | `pkg/save`               | Cross-platform save slots                                      |
| Networking              | `pkg/network`            | Client/server netcode, delta sync, lag compensation            |
| Federation              | `pkg/federation`         | Cross-server matchmaking and player lookup                     |
| Chat (E2E encrypted)    | `pkg/chat`               | Encrypted in-game text chat relay                              |
| Status effects          | `pkg/status`             | Effect registry, tick damage, HUD icons                        |
| Skill / talent trees    | `pkg/skills`             | Node graph, point allocation, passive bonuses                  |
| Lore / codex            | `pkg/lore`               | Collectible log storage, codex UI data model                   |
| Mod loader              | `pkg/mod`                | Plugin API, asset override, conflict detection                 |

---

## Plan History

- 2026-02-27 PLAN.md created for v1.0 — Core Engine + Playable Single-Player (archived to docs/)

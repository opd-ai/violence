# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

#### Documentation
- **GAPS.md Updated** — Audited and resolved all v1.0–v4.0 implementation gaps (2026-03-01)
  - Confirmed complete implementation of procedural asset pipeline, ECS query system, lighting, pathfinding, cover detection, squad formations, and text generation
  - Archived original GAPS.md to `docs/archive/GAPS_ORIGINAL_2026-03-01.md`
  - Replaced with minimal v5.0+ tracker focusing on multiplayer features (key exchange, mobile input, federation hub hosting, mod sandboxing, profanity filters)
  - **All core single-player and visual polish features are now implemented and tested**

## [5.0.0] - 2026-02-28

### Added

#### Networking — Client/Server Netcode (`pkg/network`)
- Authoritative game server with 20-tick/second update loop and command validation
- Delta-state synchronization with XOR-based entity field diffs and circular snapshot buffer
- Server-side lag compensation with world-state rewind and hitscan hit detection
- Latency tolerance: 100ms interpolation buffer, 500ms stale input threshold, 5000ms spectator fallback
- Comprehensive network unit tests (92.5% coverage)

#### Co-op Mode
- 2–4 player co-op session management with independent inventories and shared objectives
- Bleed-out respawn system (10s timer, respawn at nearest teammate, party wipe restart)
- Extended squad commands (Hold/Follow/Attack) targeting human players with command wheel UI
- Integration tests for full 4-player session lifecycle (94.5% coverage)

#### Deathmatch Mode
- Free-for-all deathmatch (2–8 players) with configurable frag/time limits
- Team deathmatch (2–16 players) with red/blue team assignment and team scoring
- Deathmatch-specific BSP arena generator with symmetrical spawns and sightline balancing
- Kill feed and scoreboard UI with K/D/A stats
- Integration tests for FFA and team matches

#### Territory Control
- Control point capture mechanics with progress bar, team ownership, and scoring
- Points-per-second scoring per held control point with configurable limits
- Genre-flavored control point visuals (altar, terminal, summoning-circle, server-rack, scrap-pile)

#### E2E Encrypted Chat (`pkg/chat`)
- In-game chat UI overlay with message history, input buffer, and scroll support
- Server-side relay with no plaintext storage (encrypted blobs only)
- Client-side profanity filter toggle with case-insensitive masking

#### Squads / Clans (`pkg/federation`)
- Squad group management (up to 8 members) with invite/accept/leave API and persistent storage
- Dedicated squad chat channel with shared AES-256 encryption key
- Aggregate squad statistics (kills, deaths, wins, play time)
- Configurable 4-character squad tag displayed in HUD nameplates

#### Federation / Cross-Server Matchmaking
- Federation protocol for server discovery via WebSocket announcements and REST queries
- Cross-server player lookup with automatic index updates
- Matchmaking queue supporting co-op, FFA, TDM, and territory modes
- Integration tests for multi-server federation

#### CI/CD
- Multi-platform build matrix: Linux (amd64/arm64), macOS (universal), Windows (amd64), WASM
- Mobile build targets: iOS (.ipa via gomobile), Android (.aar via gomobile)
- Docker image for dedicated server using distroless base
- GPG binary signing for Linux/Windows; macOS notarization workflow
- Release automation: tag-triggered draft release with signed artifacts

#### Documentation
- CHANGELOG.md following Keep a Changelog format
- CONTROLS.md keybinding reference for keyboard, mouse, and gamepad
- FAQ.md covering common issues, performance tuning, and multiplayer troubleshooting
- ARCHITECTURE.md covering ECS, raycaster, BSP, audio synthesis, and networking
- GENRE_SYSTEM.md explaining the SetGenre interface and genre registration
- MODDING.md explaining the plugin API, mod structure, and generator overrides

## [4.0.0] - 2026-02-27

### Added
- Destructible environments (`pkg/destruct`)
- Squad companion AI (`pkg/squad`)
- Procedurally generated quests and objectives (`pkg/quest`)
- Alarm and lockdown world event triggers (`pkg/event`)
- Boss arena generation in BSP (`pkg/bsp`)
- Crafting system with scrap-to-ammo conversion (`pkg/crafting`)
- Between-level armory shop (`pkg/shop`)
- Skill and talent trees (`pkg/skills`)
- Mod loader and plugin API (`pkg/mod`)

## [3.0.0] - 2026-02-26

### Added
- Procedural texture atlas generation (`pkg/texture`)
- Sector-based dynamic lighting with flashlights and point lights (`pkg/lighting`)
- Particle emitters and effects (`pkg/particle`)
- Post-processing pipeline (vignette, film grain, scanlines, chromatic aberration, bloom)
- Genre-specific visual styles across all rendering systems

## [2.0.0] - 2026-02-25

### Added
- Weapon definitions and firing mechanics (`pkg/weapon`)
- Ammo types and pools (`pkg/ammo`)
- Enemy behavior trees (`pkg/ai`)
- Damage model and hit feedback (`pkg/combat`)
- Status effects: poison, burn, bleed, radiation (`pkg/status`)
- Loot tables and drops (`pkg/loot`)
- XP and leveling progression (`pkg/progression`)
- Character class definitions (`pkg/class`)
- Five genre presets: Fantasy, Sci-Fi, Horror, Cyberpunk, Post-Apocalyptic

## [1.0.0] - 2026-02-24

### Added
- Core Ebitengine game loop with state machine (menu, playing, paused, loading)
- ECS framework (`pkg/engine`) with entity, component, and system abstractions
- DDA raycasting engine (`pkg/raycaster`) with fog and texture mapping
- BSP procedural level generator (`pkg/bsp`)
- Rendering pipeline with raycaster-to-framebuffer-to-screen flow (`pkg/render`)
- First-person camera with FOV, pitch, and head-bob (`pkg/camera`)
- Input manager for keyboard, mouse, and gamepad (`pkg/input`)
- Procedural audio engine with adaptive music layers and 3D positioning (`pkg/audio`)
- HUD, menus, and settings screens (`pkg/ui`)
- Context-sensitive tutorial prompts (`pkg/tutorial`)
- Keycard and door system (`pkg/door`)
- Fog-of-war automap (`pkg/automap`)
- Item inventory system (`pkg/inventory`)
- Save and load with cross-platform support (`pkg/save`)
- Seed-based deterministic RNG (`pkg/rng`)
- Configuration loading via Viper (`pkg/config`)
- Procedurally generated collectible lore and codex (`pkg/lore`)
- Hacking and lockpicking mini-games (`pkg/minigame`)
- Decorative prop placement (`pkg/props`)

[Unreleased]: https://github.com/opd-ai/violence/compare/v5.0.0...HEAD
[5.0.0]: https://github.com/opd-ai/violence/compare/v4.0.0...v5.0.0
[4.0.0]: https://github.com/opd-ai/violence/compare/v3.0.0...v4.0.0
[3.0.0]: https://github.com/opd-ai/violence/compare/v2.0.0...v3.0.0
[2.0.0]: https://github.com/opd-ai/violence/compare/v1.0.0...v2.0.0
[1.0.0]: https://github.com/opd-ai/violence/releases/tag/v1.0.0

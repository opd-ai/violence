# Implementation Plan: v5.0+ — Multiplayer, Social Features, Production Polish

## Phase Overview
- **Objective**: Deliver co-op and deathmatch multiplayer with production-grade networking, social features, and cross-platform release automation.
- **Source Document**: ROADMAP.md (v5.0+ section)
- **Prerequisites**: v1.0–v4.0 complete (Core Engine, Weapons/AI/Genres, Visual Polish, Gameplay Expansion all implemented)
- **Estimated Scope**: Large

## Implementation Steps

### Networking — Client/Server Netcode (`pkg/network`)
1. [x] Implement authoritative server model with tick-based game loop (2026-02-28)
   - **Deliverable**: `GameServer` struct with 20-tick/second update loop; client commands validated server-side before state mutation
   - **Dependencies**: None
   - **Completed**: Implemented `GameServer` with 20-tick/second update loop, command validation via `CommandValidator` interface, client connection management, graceful shutdown, and comprehensive tests achieving 89.6% coverage

2. [x] Implement delta-state synchronization (2026-02-28)
   - **Deliverable**: `DeltaEncoder`/`DeltaDecoder` that serialize only changed entity fields; baseline snapshot + delta packets reduce bandwidth
   - **Dependencies**: Step 1, `pkg/engine` ECS
   - **Completed**: Implemented `DeltaEncoder` with snapshot buffering (500ms/10 snapshots), `DeltaDecoder` with delta application, entity diff computation using XOR/presence bitmask, circular buffer for lag compensation, and comprehensive tests achieving 83.8% coverage

3. [x] Implement lag compensation for hitscan weapons (2026-02-28)
   - **Deliverable**: Server-side rewind system that reconstructs entity positions at client's perceived time; hit detection uses rewound state
   - **Dependencies**: Step 2
   - **Completed**: Implemented `LagCompensator` with snapshot history buffer (500ms/10 snapshots), `RewindWorld()` for tick-based rewind with interpolation, `PerformHitscan()` for lag-compensated hit detection using rewound positions, ray-sphere intersection, and comprehensive tests achieving 85.9% coverage

4. [x] Implement latency tolerance (200ms optimal, 500ms degraded, 5000ms spectator fallback) (2026-02-28)
   - **Deliverable**: Client interpolation buffer (100ms); server accepts inputs up to 500ms stale; >5000ms triggers spectator mode with reconnect prompt
   - **Dependencies**: Steps 1–3
   - **Completed**: Implemented `InterpolationBuffer` for client-side rendering (100ms delay/2 tick buffer), `LatencyMonitor` tracking RTT with spectator mode threshold (5000ms), server-side stale input rejection (`IsInputStale` at 500ms threshold), quality classification, and comprehensive tests achieving 100% coverage on all latency functions (86.8% package coverage)

5. [x] Add unit tests for network layer (2026-02-28)
   - **Deliverable**: Tests for delta encoding round-trip, lag compensation accuracy, latency edge cases
   - **Dependencies**: Steps 1–4
   - **Completed**: Added comprehensive unit tests covering delta encoding edge cases (entity diffs, deep copy, merge operations), lag compensation scenarios (empty snapshots, interpolation, raycasting), and gameserver command validation. Coverage increased from 86.6% to 92.5%.

### Co-op Mode (`pkg/network`)
6. [x] Implement 2–4 player co-op session management (2026-02-28)
   - **Deliverable**: `CoopSession` struct managing player join/leave, shared level state, independent inventories, shared objective progress
   - **Dependencies**: Steps 1–2
   - **Completed**: Implemented `CoopSession` with player management (2-4 players), `CoopPlayerState` with independent inventories, shared quest tracker for objectives, thread-safe concurrent access, player join/leave with inactive state preservation, position/health tracking, genre configuration, and comprehensive tests achieving 94.1% coverage

7. [x] Implement co-op respawn system (2026-02-28)
   - **Deliverable**: Dead players enter 10-second bleed-out timer; respawn at nearest living teammate; full party wipe restarts level
   - **Dependencies**: Step 6, `pkg/combat`
   - **Completed**: Implemented `OnPlayerDeath()` with 10-second bleedout timer, `ProcessBleedouts()` for expired timer detection, `RespawnPlayer()` at nearest living teammate position using distance calculation, `isPartyWiped()` detection, `RestartLevel()` for full party wipe with quest tracker reset, and comprehensive tests achieving 94.2% coverage (9 new test cases covering death, bleedout, respawn, party wipe, and level restart scenarios)

8. [x] Extend squad commands for human players (2026-02-28)
   - **Deliverable**: Hold/Follow/Attack commands target human teammates; command wheel UI shows connected players
   - **Dependencies**: Step 6, `pkg/squad`
   - **Completed**: Extended `Squad` to track human players via `HumanPlayer` struct with position/health tracking. Added `CommandTargetPlayer()` for player-targeted commands ("follow_player", "attack_player_target"). Squad members follow target player when `TargetPlayerID` is set. Implemented `CommandWheel` UI component with player selection, health bars, and navigation. Added comprehensive tests achieving 100% coverage on squad package

9. [x] Add integration tests for co-op mode (2026-02-28)
   - **Deliverable**: Simulated 4-player session test: join, combat, respawn, objective completion
   - **Dependencies**: Steps 6–8
   - **Completed**: Implemented comprehensive integration test suite (`coop_integration_test.go`) covering: (1) full 4-player session with join/combat/respawn/objective completion, (2) party wipe and level restart, (3) player disconnect/reconnect, (4) concurrent actions for thread safety, (5) shared objective tracking, (6) sequential respawn chain, (7) genre configuration, (8) edge cases (session full, insufficient players, duplicate adds), and (9) combat system integration. All 9 test scenarios pass with 94.5% package coverage

### Deathmatch Mode (`pkg/network`)
10. [x] Implement free-for-all deathmatch (2026-02-28)
    - **Deliverable**: `FFAMatch` struct with configurable frag limit, time limit, instant respawn at random spawn points
    - **Dependencies**: Steps 1–2
    - **Completed**: Implemented `FFAMatch` with player management (2-8 players), configurable frag/time limits, `OnPlayerKill()`/`OnPlayerSuicide()` for score tracking, `ProcessRespawns()` for 3-second respawn delay, deterministic spawn point generation using match seed, leaderboard sorting, win condition detection (frag limit or time limit), and comprehensive tests achieving 95.9% package coverage

11. [x] Implement team deathmatch (2026-02-28)
    - **Deliverable**: `TeamMatch` struct with team assignment, team-colored player indicators, team score tracking
    - **Dependencies**: Step 10
    - **Completed**: Implemented `TeamMatch` with player management (2-16 players), two-team system (red/blue), team assignment on join, team-specific spawn points (left/right side split), `OnPlayerKill()`/`OnPlayerSuicide()` with team score tracking, `GetTeamScore()` for team statistics, leaderboard sorted by team then frags, win condition based on team frag limit or time limit, and comprehensive tests (22 test cases) achieving 96.2% package coverage

12. [x] Generate deathmatch-specific BSP maps (2026-02-28)
    - **Deliverable**: Arena layout generator with symmetrical spawn pads, weapon spawn locations, sightline balancing
    - **Dependencies**: `pkg/bsp`
    - **Completed**: Implemented `ArenaGenerator` in `pkg/bsp/deathmatch.go` with: (1) symmetrical 4-way rotational spawn pad placement, (2) strategic weapon spawn locations (power weapons at center, mid-tier on cardinal directions, basic on diagonals), (3) tactical cover points in ring pattern, (4) sightline analysis using 16-direction raycasting, (5) automatic sightline balancing to prevent overpowered spawn positions, (6) genre-specific tile selection, (7) rounded arena corners for smooth movement, and comprehensive tests (13 test cases) achieving 94.5% package coverage

13. Implement kill feed and scoreboard UI
    - **Deliverable**: Real-time kill notifications; end-of-match scoreboard with K/D/A stats
    - **Dependencies**: Steps 10–11, `pkg/ui`

14. Add integration tests for deathmatch
    - **Deliverable**: Simulated 4-player FFA and 2v2 team matches with frag tracking validation
    - **Dependencies**: Steps 10–13

### Territory Control (Deathmatch Variant)
15. Implement control point capture mechanics
    - **Deliverable**: `ControlPoint` entity with capture progress bar, team ownership, score tick rate
    - **Dependencies**: Step 11

16. Implement territory control scoring
    - **Deliverable**: Teams earn points per second for each held control point; first to score limit wins
    - **Dependencies**: Step 15

17. Add genre-flavored control point visuals
    - **Deliverable**: `SetGenre()` selects control point visual style (altar/terminal/summoning-circle/server-rack/scrap-pile)
    - **Dependencies**: Step 15, `pkg/procgen/genre`

### E2E Encrypted Chat (`pkg/chat`)
18. Implement in-game chat UI overlay
    - **Deliverable**: Toggle-able chat window; message history; input field with send on Enter
    - **Dependencies**: `pkg/ui`

19. Implement server-side relay with no plaintext storage
    - **Deliverable**: Server relays encrypted blobs without decryption capability; messages encrypted client-side
    - **Dependencies**: Existing `pkg/chat` encryption

20. Implement profanity filter toggle
    - **Deliverable**: Client-side filter option that masks flagged words; toggled in settings
    - **Dependencies**: Step 18, `pkg/config`

21. Add unit tests for chat
    - **Deliverable**: Tests for encryption round-trip, relay without decryption, filter masking
    - **Dependencies**: Steps 18–20

### Squads / Clans (`pkg/federation`)
22. Implement squad group management
    - **Deliverable**: `Squad` struct with up to 8 members; invite/accept/leave API; persistent squad storage
    - **Dependencies**: `pkg/save`

23. Implement squad chat channel
    - **Deliverable**: Dedicated chat channel visible only to squad members; uses shared squad encryption key
    - **Dependencies**: Steps 18–19, Step 22

24. Implement squad statistics
    - **Deliverable**: Aggregate stats across squad members: total kills, wins, play time; displayed in squad info screen
    - **Dependencies**: Step 22

25. Implement squad tag display
    - **Deliverable**: Configurable 4-character tag shown in HUD nameplates above squad members
    - **Dependencies**: Step 22, `pkg/ui`

### Federation / Cross-Server Matchmaking (`pkg/federation`)
26. Implement federation protocol for server discovery
    - **Deliverable**: Servers announce to federation hub; clients query hub for available servers by region/genre/player count
    - **Dependencies**: Existing `pkg/federation`

27. Implement cross-server player lookup
    - **Deliverable**: Query player presence across federated servers; return server address if online
    - **Dependencies**: Step 26

28. Implement matchmaking queue
    - **Deliverable**: Players queue for mode (co-op/FFA/TDM/territory); matchmaker groups players and assigns server
    - **Dependencies**: Step 26

29. Add integration tests for federation
    - **Deliverable**: Simulated multi-server federation with player lookup and matchmaking
    - **Dependencies**: Steps 26–28

### CI/CD — Full Production
30. Create multi-platform build matrix
    - **Deliverable**: GitHub Actions workflow building for Linux (amd64/arm64), macOS (universal), Windows (amd64), WASM
    - **Dependencies**: Existing `.github/workflows/ci.yml`

31. Add mobile build targets (iOS/Android)
    - **Deliverable**: Gomobile-based build jobs producing `.ipa` and `.apk` artifacts
    - **Dependencies**: Step 30

32. Create Docker image for dedicated server
    - **Deliverable**: `Dockerfile` producing minimal server image; published to GitHub Container Registry
    - **Dependencies**: Step 1

33. Implement binary signing
    - **Deliverable**: GPG signing for Linux/Windows; notarization workflow for macOS
    - **Dependencies**: Step 30

34. Implement release automation
    - **Deliverable**: Git tag triggers draft release with all platform artifacts uploaded
    - **Dependencies**: Steps 30–33

### Documentation Suite
35. Create CHANGELOG.md with semver format
    - **Deliverable**: Changelog following Keep a Changelog format; auto-updated by release workflow
    - **Dependencies**: None

36. Create CONTROLS.md keybinding reference
    - **Deliverable**: Full keyboard/mouse/gamepad mapping documentation
    - **Dependencies**: `pkg/input`

37. Create FAQ.md
    - **Deliverable**: Common issues and answers; performance tuning; multiplayer troubleshooting
    - **Dependencies**: None

38. Create architecture documentation
    - **Deliverable**: `docs/ARCHITECTURE.md` covering ECS, raycaster, BSP, audio synthesis, networking
    - **Dependencies**: None

39. Create genre system guide
    - **Deliverable**: `docs/GENRE_SYSTEM.md` explaining `SetGenre()` interface and adding new genres
    - **Dependencies**: `pkg/procgen/genre`

40. Create modding guide
    - **Deliverable**: `docs/MODDING.md` explaining plugin API, mod structure, generation parameter overrides
    - **Dependencies**: `pkg/mod`

### Mod Framework (`pkg/mod`)
41. Define plugin interface for custom content
    - **Deliverable**: `Plugin` interface with `Load()`, `Unload()`, `GetGenerators()` methods
    - **Dependencies**: Existing `pkg/mod`

42. Implement mod loader with conflict detection
    - **Deliverable**: Load mods from `mods/` directory; detect conflicting generator overrides; report warnings
    - **Dependencies**: Step 41

43. Implement generation parameter override system
    - **Deliverable**: Mods register custom enemy types, weapon definitions, texture parameters as generation rules
    - **Dependencies**: Step 42

44. Add unit tests for mod framework
    - **Deliverable**: Tests for plugin loading, conflict detection, parameter override precedence
    - **Dependencies**: Steps 41–43

### Test Coverage
45. Audit and expand test coverage to 82%+
    - **Deliverable**: Coverage report; new tests for under-covered packages; all packages meet 82% threshold
    - **Dependencies**: All above steps

46. Create multiplayer integration test harness
    - **Deliverable**: Test framework that spins up server + N clients; simulates full session lifecycle
    - **Dependencies**: Steps 1–14

47. Automate coverage gate in CI
    - **Deliverable**: CI fails if overall coverage drops below 82%
    - **Dependencies**: Steps 45–46

## Technical Specifications
- **Tick rate**: Server runs at 20 ticks/second; clients render at unlocked FPS with interpolation
- **Delta encoding**: Entity state diffs using XOR for numerics, presence bitmask for optional fields
- **Lag compensation buffer**: Server stores 500ms of world snapshots (10 snapshots at 20 tick/s)
- **Interpolation delay**: Clients render 100ms behind server time for smooth movement
- **Squad size limit**: 8 players maximum per squad
- **Federation protocol**: JSON-over-WebSocket for server announcements; REST API for client queries
- **Encryption**: AES-256-GCM for chat (existing implementation); key exchange via ECDH
- **Docker base image**: `gcr.io/distroless/static-debian12` for minimal attack surface
- **Binary signing**: GPG key stored in GitHub Secrets; macOS uses `notarytool`
- **Mod format**: Mods are Go plugins (`.so` on Linux/macOS, `.dll` on Windows) implementing `Plugin` interface

## Validation Criteria
- [ ] 2–4 player co-op session runs end-to-end: join → play level → respawn → complete objective
- [ ] 4-player FFA deathmatch runs with accurate frag tracking
- [ ] Team deathmatch correctly tracks team scores
- [ ] Territory control points capture and score correctly
- [ ] Chat messages encrypt/decrypt correctly between clients
- [ ] Squad members see shared tag in HUD
- [ ] Federation matchmaking places players on available servers
- [ ] CI builds successfully on all 6 platform targets (Linux x2, macOS, Windows, WASM, mobile x2)
- [ ] Docker server image runs and accepts connections
- [ ] Release automation uploads signed artifacts for tagged commits
- [ ] Documentation covers all multiplayer and modding features
- [ ] Test coverage ≥ 82% across entire codebase
- [ ] Network code handles 500ms latency without gameplay breakage
- [ ] Mods load and override generation parameters correctly

## Known Gaps
- **Key exchange protocol**: No ECDH key exchange implementation exists for establishing shared chat encryption keys between clients
- **Mobile input mapping**: Touch controls for mobile builds are undefined; need touch-to-virtual-joystick mapping
- **Federation hub hosting**: No specification for who hosts the federation hub server; may need self-hosted option
- **Mod sandboxing**: Go plugins have full runtime access; no sandboxing mechanism to limit mod capabilities
- **Profanity word list**: No word list defined for profanity filter; need localized lists for multiple languages

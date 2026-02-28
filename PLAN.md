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

13. [x] Implement kill feed and scoreboard UI (2026-02-28)
    - **Deliverable**: Real-time kill notifications; end-of-match scoreboard with K/D/A stats
    - **Dependencies**: Steps 10–11, `pkg/ui`
    - **Completed**: Implemented `KillFeed` with real-time kill notifications (max 5 entries, 5s duration), support for suicide and team kill indicators, automatic entry expiration. Implemented `Scoreboard` with configurable FFA/team modes, K/D/A stat display, winner text, visibility toggle, team-colored player indicators. Added comprehensive unit tests for all UI logic components achieving high coverage on testable functions (rendering tests skipped due to display requirements)

14. [x] Add integration tests for deathmatch (2026-02-28)
    - **Deliverable**: Simulated 4-player FFA and 2v2 team matches with frag tracking validation
    - **Dependencies**: Steps 10–13
    - **Completed**: Implemented comprehensive integration test suite (`deathmatch_integration_test.go`) with 4 scenarios: (1) full 4-player FFA match with combat/respawn/frag limit win, (2) 2v2 team match with team score tracking and frag limit win, (3) time limit win condition test, and (4) continuous respawn cycle test. All tests pass, validating kill feed tracking, scoreboard updates, and proper match state management

### Territory Control (Deathmatch Variant)
15. [x] Implement control point capture mechanics (2026-02-28)
    - **Deliverable**: `ControlPoint` entity with capture progress bar, team ownership, score tick rate
    - **Dependencies**: Step 11
    - **Completed**: Implemented `ControlPoint` struct with position tracking, ownership (neutral/red/blue), capture progress (-1.0 to +1.0), capture radius (5.0 units), and `UpdateCapture()` processing player counts. Implemented `TerritoryMatch` managing control points, players, teams, score tracking, configurable score/time limits, `ProcessCapture()` for capture mechanics, `ProcessScoring()` awarding points per second for held control points (1 point per CP per tick), and `CheckWinCondition()` for score limit and time limit. Added comprehensive tests (30+ test cases) achieving 96.7% package coverage

16. [x] Implement territory control scoring (2026-02-28)
    - **Deliverable**: Teams earn points per second for each held control point; first to score limit wins
    - **Dependencies**: Step 15
    - **Completed**: Scoring system fully implemented in `TerritoryMatch.ProcessScoring()` - teams earn configurable points (default 1 point) per second per held control point, neutral points award no score, tick rate limits scoring frequency (default 1 second), win conditions check score limit and time limit. Added comprehensive integration tests covering full match flow, contested points, player advantage, scoring mechanics, and time limits. All tests pass with 96.7% package coverage

17. [x] Add genre-flavored control point visuals (2026-02-28)
    - **Deliverable**: `SetGenre()` selects control point visual style (altar/terminal/summoning-circle/server-rack/scrap-pile)
    - **Dependencies**: Step 15, `pkg/procgen/genre`
    - **Completed**: Added `VisualStyle` field to `ControlPoint` struct, `SetVisualStyle()`/`GetVisualStyle()` methods for individual control points, `genreToVisualStyle()` function mapping genre IDs to visual styles (fantasy→altar, scifi→terminal, horror→summoning-circle, cyberpunk→server-rack, postapoc→scrap-pile, default→generic), and `TerritoryMatch.SetGenre()` to update all control points in a match. Added comprehensive tests (18 test cases) covering default styles, style changes, genre mapping, and match-level genre updates. All tests pass with 96.7% package coverage

### E2E Encrypted Chat (`pkg/chat`)
18. [x] Implement in-game chat UI overlay (2026-02-28)
    - **Deliverable**: Toggle-able chat window; message history; input field with send on Enter
    - **Dependencies**: `pkg/ui`
    - **Completed**: Implemented `ChatOverlay` struct with visibility toggle, message history (max 100 messages, 10 visible), input buffer (max 200 chars), scroll support (PgUp/PgDn), position/size configuration. Added comprehensive tests covering all core functionality (visibility, message management, input buffer, scrolling, concurrent access) achieving 48.3% coverage on chat.go (Draw function excluded as it requires display)

19. [x] Implement server-side relay with no plaintext storage (2026-02-28)
    - **Deliverable**: Server relays encrypted blobs without decryption capability; messages encrypted client-side
    - **Dependencies**: Existing `pkg/chat` encryption
    - **Completed**: Implemented `RelayServer` with TCP-based encrypted message relay (no plaintext storage/decryption), `RelayClient` for client connections, line-based message protocol with proper delimiter handling, broadcast support ("all" recipient), per-client connection tracking, graceful shutdown, and `EncryptedMessage` struct for encrypted blob transmission. Added 10 comprehensive test scenarios including server lifecycle, client connections, encrypted message relay, broadcast messaging, E2E encryption with AES-256-GCM, multiple message handling, and no-plaintext-storage verification. All tests pass with 87.8% package coverage

20. [x] Implement profanity filter toggle (2026-02-28)
    - **Deliverable**: Client-side filter option that masks flagged words; toggled in settings
    - **Dependencies**: Step 18, `pkg/config`
    - **Completed**: Added `ProfanityFilter` boolean field to `Config` struct (defaults to `true`), implemented `FilterProfanity()` function with case-insensitive substring matching that replaces flagged words with asterisks of equal length, included minimal profanity word list (18 common words), added `AddProfanityWord()`, `ClearProfanityWords()`, and `SetProfanityWords()` for customization. Added 50+ comprehensive test cases covering filter enable/disable, encryption round-trip, relay integration, multiple occurrences, edge cases, case insensitivity, and length preservation. All tests pass with 89.1% package coverage

21. [x] Add unit tests for chat (2026-02-28)
    - **Deliverable**: Tests for encryption round-trip, relay without decryption, filter masking
    - **Dependencies**: Steps 18–20
    - **Completed**: Extended test suite with profanity filter integration tests including: encryption round-trip with filtering, relay integration (verifying server sees only encrypted blobs while clients filter after decryption), multiple occurrences, edge cases (unicode, numbers, special chars, substrings), case insensitivity, and length preservation. Total of 60+ test cases across all chat functionality with comprehensive coverage of encryption, relay, and filtering features. Package coverage: 89.1%

### Squads / Clans (`pkg/federation`)
22. [x] Implement squad group management (2026-02-28)
    - **Deliverable**: `Squad` struct with up to 8 members; invite/accept/leave API; persistent squad storage
    - **Dependencies**: `pkg/save`
    - **Completed**: Implemented `Squad` struct with member management (max 8 players), `SquadMember` with player info and leader designation, `Invite()`/`Accept()`/`Leave()` API for join flow, automatic leader promotion when leader leaves, `SquadManager` for managing multiple squads with `CreateSquad()`, `GetSquad()`, `DeleteSquad()`, `ListSquads()`, persistent storage via `Save()`/`Load()` to `~/.violence/squads/squads.json`, thread-safe concurrent access with mutex protection, and comprehensive tests achieving 96.3% coverage

23. [x] Implement squad chat channel (2026-02-28)
    - **Deliverable**: Dedicated chat channel visible only to squad members; uses shared squad encryption key
    - **Dependencies**: Steps 18–19, Step 22
    - **Completed**: Implemented `SquadChatChannel` with shared AES-256 encryption key generation, `NewSquadChatChannel()` for creating new channels and `NewSquadChatChannelWithKey()` for joining with existing key, `SendMessage()` encrypting and broadcasting to all squad members via relay server, `ReceiveMessages()` polling and decrypting messages using squad key (non-members cannot decrypt), `SquadChatManager` for managing multiple squad channels with create/join/remove operations, message history tracking, encryption key sharing for new members, and comprehensive tests (13 test scenarios including E2E integration) achieving 93.5% package coverage. Added `GetAddr()` method to `RelayServer` for test infrastructure

24. [x] Implement squad statistics (2026-02-28)
    - **Deliverable**: Aggregate stats across squad members: total kills, wins, play time; displayed in squad info screen
    - **Dependencies**: Step 22
    - **Completed**: Added `MemberStats` struct tracking individual member statistics (kills, deaths, wins, play time), `SquadStats` struct with aggregated totals and averages across all squad members, `GetStats()` method computing total and average statistics, `UpdateMemberStats()` for incremental stat updates, `SetMemberStats()` for replacing stats, `GetMemberStats()` for retrieving individual member stats, stats persistence via JSON serialization, thread-safe concurrent stat updates with mutex protection, and comprehensive tests (13 test scenarios covering empty squads, single/multiple members, incremental updates, persistence, concurrent access, member departures, and zero stats) achieving 94.4% package coverage

25. [x] Implement squad tag display (2026-02-28)
    - **Deliverable**: Configurable 4-character tag shown in HUD nameplates above squad members
    - **Dependencies**: Step 22, `pkg/ui`
    - **Completed**: Added `MaxTagLength` constant (4 chars) to `pkg/federation`, implemented `GetTag()`/`SetTag()`/`GetName()`/`GetID()` methods with auto-truncation, created `Nameplate` component in `pkg/ui` with `NameplatePlayer` struct for player info display, squad tag rendering above player names, color-coded nameplates (green teammates, red enemies, yellow self), semi-transparent backgrounds with borders, customizable colors via `SetTeammateColor()`/`SetEnemyColor()`/`SetSelfColor()`, and comprehensive tests achieving 100% coverage on squad tag logic and nameplate logic (drawing methods excluded due to display requirements). Documentation added to `docs/SQUAD_TAG_DISPLAY.md`

### Federation / Cross-Server Matchmaking (`pkg/federation`)
26. [x] Implement federation protocol for server discovery (2026-02-28)
    - **Deliverable**: Servers announce to federation hub; clients query hub for available servers by region/genre/player count
    - **Dependencies**: Existing `pkg/federation`
    - **Completed**: Implemented `FederationHub` with WebSocket-based server announcement system, REST API for client queries, automatic stale server cleanup (30s timeout), `ServerAnnouncer` for periodic heartbeats (10s interval), region-based filtering (7 regions: US East/West, EU East/West, Asia-Pacific, South America, Unknown), genre/player count query filters, and comprehensive tests with 92.8% package coverage. Added gorilla/websocket dependency for WebSocket support

27. [x] Implement cross-server player lookup (2026-02-28)
    - **Deliverable**: Query player presence across federated servers; return server address if online
    - **Dependencies**: Step 26
    - **Completed**: Implemented `PlayerLookupRequest`/`PlayerLookupResponse` structs, `handleLookup()` HTTP endpoint, `lookupPlayer()` for querying player presence, player index (`playerID -> serverName` mapping) in `FederationHub`, automatic index updates when servers announce player lists, player cleanup when servers go stale, `UpdatePlayerList()` method for `ServerAnnouncer`, and comprehensive tests (10 new test cases covering unit tests, HTTP endpoint, index updates, stale cleanup, and E2E integration) achieving 92.7% package coverage

28. [x] Implement matchmaking queue (2026-02-28)
    - **Deliverable**: Players queue for mode (co-op/FFA/TDM/territory); matchmaker groups players and assigns server
    - **Dependencies**: Step 26
    - **Completed**: Implemented `Matchmaker` with player queue management (enqueue/dequeue), game mode support (co-op/FFA/TDM/territory), configurable min/max players per mode, player grouping by genre/region, automatic match creation when sufficient players available, server capacity checking, 60-second queue timeout with automatic cleanup, 2-second match processing interval, and comprehensive tests (17 test scenarios covering enqueue/dequeue, grouping, matching, timeouts, server capacity, and integration) achieving 93.9% package coverage

29. [x] Add integration tests for federation (2026-02-28)
    - **Deliverable**: Simulated multi-server federation with player lookup and matchmaking
    - **Dependencies**: Steps 26–28
    - **Completed**: Implemented comprehensive integration test suite (`federation_integration_test.go`) with 5 test scenarios: (1) full multi-server federation test with 3 servers covering server query by region/genre/combined filters, player lookup across servers, cross-server player migration, matchmaking with FFA mode, server capacity filtering, and stale server cleanup, (2) matchmaking across all game modes (co-op/FFA/TDM/territory) with min player requirements, (3) concurrent server announcements testing hub stability with 10 simultaneous servers, (4) region-based matchmaking ensuring players match within their regions across 4 regions (US-East/West, EU-West, Asia-Pac). Added `SetInterval()` method to `ServerAnnouncer` and `SetCleanupInterval()`/`SetStaleTimeout()` methods to `FederationHub` for test configurability. All integration tests pass with 93.5% federation package coverage

### CI/CD — Full Production
30. [x] Create multi-platform build matrix (2026-02-28)
    - **Deliverable**: GitHub Actions workflow building for Linux (amd64/arm64), macOS (universal), Windows (amd64), WASM
    - **Dependencies**: Existing `.github/workflows/ci.yml`
    - **Completed**: Implemented `.github/workflows/build.yml` with build jobs for Linux (amd64/arm64 via cross-compilation), macOS (universal binary via lipo), Windows (amd64), and WASM (with wasm_exec.js loader). Added artifact uploads (30-day retention), build summary job, and comprehensive documentation in `docs/BUILD_MATRIX.md` covering platform targets, cross-compilation instructions, CGO requirements, and local build commands

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

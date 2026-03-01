# Implementation Plan: v6.0 — Gap Resolution & Test Coverage

## Phase Overview
- **Objective**: Resolve all documented implementation gaps from GAPS.md and achieve 82%+ test coverage gate
- **Source Document**: GAPS.md, ROADMAP.md (all v1.0-v5.0 milestones completed per CHANGELOG.md)
- **Prerequisites**: v5.0.0 released (2026-02-28)
- **Estimated Scope**: Large

## Implementation Steps

### 1. Fix Sector Light Map Implementation
- [x] **Completed 2026-03-01**: `pkg/lighting/sector.go` with `SectorLightMap` struct, `AddLight()` with inverse-square falloff, and `Calculate()` precomputation. Removed stub functions from `lighting.go`.
- **Deliverable**: `pkg/lighting/sector_lightmap.go` with `SectorLightMap` struct, `AddPointLight()` with inverse-square falloff, and `Calculate()` precomputation
- **Dependencies**: None

### 2. Implement Flashlight Cone Lighting
- [x] **Completed 2026-03-01**: Added `AddFlashlight(x, y, dirX, dirY, coneAngle, range, intensity)` method to `SectorLightMap` with dot-product angle test and quadratic attenuation. Full test coverage (99.1%).
- **Deliverable**: `AddFlashlight(x, y, dirX, dirY, coneAngle, range, intensity)` method using dot-product angle test against player direction
- **Dependencies**: Step 1 (SectorLightMap)

### 3. Add Wall Texture Coordinate to RayHit
- [x] **Completed 2026-03-01**: `TextureX` field added to `RayHit` struct in `pkg/raycaster/raycaster.go` (line 42). Calculated as fractional part of exact wall hit position using Y coordinate for vertical walls and X coordinate for horizontal walls (lines 168-176). Full test coverage with 3 test functions.
- **Deliverable**: Extend `RayHit` struct in `pkg/raycaster/types.go` with `TextureX float64` computed as fractional part of exact wall hit position
- **Dependencies**: None

### 4. Implement Weather Emitter Genre Configurations
- [x] **Completed 2026-03-01**: `WeatherEmitter` type implemented in `pkg/particle/weather.go` with per-genre spawn configurations (rate, velocity, color, lifetime, spawn positions). Includes 5 genre-specific emit methods. Full test coverage with 6 test functions.
- **Deliverable**: `WeatherEmitter` type in `pkg/particle/weather.go` with per-genre spawn configurations (rate, velocity, color, lifetime, spawn positions)
- **Dependencies**: None

### 5. Add Cyberpunk Neon Pulse Animated Texture
- [x] **Completed 2026-03-01**: `generateNeonPulseFrame()` implemented in `pkg/texture/animated.go` (lines 216-258) with magenta/cyan color cycling, horizontal bar pulsing, and scan line effects. Full test coverage.
- **Deliverable**: `generateNeonPulseFrame()` in `pkg/texture/animated.go` with magenta/cyan color cycling
- **Dependencies**: None

### 6. Add Post-Apocalyptic Radiation Glow Animated Texture
- [x] **Completed 2026-03-01**: `generateRadiationGlowFrame()` implemented in `pkg/texture/animated.go` (lines 260-296) with green-yellow pulsing glow effect, radial gradient, and organic shimmer. Full test coverage.
- **Deliverable**: `generateRadiationGlowFrame()` in `pkg/texture/animated.go` with green pulsing glow effect
- **Dependencies**: None

### 7. Add Horror Static Burst Post-Processing Effect
- [x] **Completed 2026-03-01**: `ApplyStaticBurst()` implemented in `pkg/render/postprocess.go` (lines 350-381) with configurable probability and duration. Integrated into horror preset (lines 523-528). Full test coverage with 5 test functions.
- **Deliverable**: `ApplyStaticBurst()` in post-processor with configurable probability and duration; integrate into horror preset
- **Dependencies**: None

### 8. Add Postapoc Film Scratch Post-Processing Effect
- [x] **Completed 2026-03-01**: `ApplyFilmScratches()` implemented in `pkg/render/postprocess.go` (lines 383-415) with configurable scratch density and length. Integrated into postapoc preset (lines 606-610). Full test coverage with 5 test functions.
- **Deliverable**: `ApplyFilmScratches()` in post-processor with configurable scratch density and opacity
- **Dependencies**: None

### 9. Implement BSP-to-Reverb Integration
- [x] **Completed 2026-03-01**: `SetRoomFromBSP(room *bsp.Room)` method implemented in `pkg/audio/reverb.go` (lines 38-44) that extracts bounds and calls `SetRoomSize()`. Full test coverage with 2 test functions.
- **Deliverable**: `SetRoomFromBSP(room *bsp.Room)` method in `pkg/audio/reverb.go` that extracts bounds and calls `SetRoomSize()`
- **Dependencies**: None

### 10. Define Squad AI Formation Algorithm
- [x] **Completed 2026-03-01**: Formation shapes implemented as `GetFormationOffset()` function in `pkg/squad/formation.go`. Supports Line, Wedge, Column, Circle, and Staggered formations with proper rotation based on leader direction. Full test coverage.
- **Deliverable**: Formation shapes as offset arrays in `pkg/squad/formation.go`; implement `GetFormationOffset(memberIndex, formationType, leaderDir)`
- **Dependencies**: None

### 11. Design Procedural Text Generation Grammar
- [x] **Completed 2026-03-01**: Markov chain generator implemented in `pkg/lore/grammar.go` with genre-specific word banks for all five genres. Supports sentence structure templates for notes, logs, and graffiti. Full test coverage with 15+ test functions.
- **Deliverable**: Markov chain or template-based grammar in `pkg/lore/grammar.go` with genre-specific word banks; sentence structure templates for notes, logs, and graffiti
- **Dependencies**: None

### 12. Implement ECDH Key Exchange for Chat Encryption
- [x] **Completed 2026-03-01**: ECDH key exchange implemented in `pkg/chat/keyexchange.go` using P-256 curve with HKDF-SHA3-256 for AES-256 key derivation. Full protocol implementation with send/receive public key functions. Test coverage verified.
- **Deliverable**: ECDH key exchange during session join in `pkg/chat/crypto.go`; derive AES key from shared secret
- **Dependencies**: None

### 13. Design Mobile Touch Control Layout
- **Deliverable**: Virtual joystick overlay for movement, touch-to-look for aiming, tap buttons for fire/interact/reload in `pkg/input/touch.go`; update `docs/MOBILE_CONTROLS.md`
- **Dependencies**: None

### 14. Define Federation Hub Self-Hosting Protocol
- **Deliverable**: Hub protocol specification in `docs/FEDERATION_HUB.md` with self-hosting instructions; evaluate distributed hash table approach for decentralized discovery
- **Dependencies**: None

### 15. Evaluate WASM-Based Mod Sandboxing
- **Deliverable**: WASM mod runtime implementation in `pkg/mod/wasm.go` as alternative to Go plugins; update `docs/MOD_SECURITY.md` with security model
- **Dependencies**: None

### 16. Compile Profanity Filter Word Lists
- **Deliverable**: Word lists for English, Spanish, German, French, Portuguese in `pkg/chat/wordlists/`; implement list loading from config
- **Dependencies**: None

### 17. Playtest and Balance Credit Economy
- **Deliverable**: Adjusted credit values in `pkg/economy/config.go` achieving ~3 shop purchases per level average; document tuning rationale in `docs/ECONOMY.md`
- **Dependencies**: None

### 18. Increase Root Package Test Coverage
- **Deliverable**: Additional tests for `main.go` and root package achieving 82%+ coverage; focus on game loop, state transitions, and integration paths
- **Dependencies**: Steps 1-17

### 19. Add cmd/server Test Coverage
- **Deliverable**: Unit tests for `cmd/server` achieving 82%+ coverage; mock network interfaces for isolation
- **Dependencies**: Step 18

## Technical Specifications
- All new code must implement `SetGenre(genreID string)` interface where applicable
- Lighting calculations use inverse-square falloff: `intensity / (distance² + 1)`
- Flashlight cone test: `dot(lightDir, tileDir) > cos(coneAngle/2)`
- ECDH uses P-256 curve with HKDF-SHA256 for key derivation
- WASM mod runtime uses Wazero for Go-native WebAssembly execution
- Mobile touch controls use Ebitengine's `TouchIDs()` and `TouchPosition()` APIs
- Profanity filter uses case-insensitive substring matching with Unicode normalization

## Validation Criteria
- [x] `pkg/lighting` tests pass with 99.1% coverage (exceeds 82%+ target)
- [x] Flashlight cone visibly illuminates forward area in all genres
- [x] Wall textures render correctly with `TextureX` coordinate
- [x] Weather particles spawn in all five genres
- [x] Cyberpunk neon pulse and postapoc radiation glow animate correctly
- [x] Horror static burst triggers randomly during gameplay
- [x] Postapoc film scratches visible on screen
- [x] Reverb parameters change based on BSP room size
- [x] Squad members maintain formation positions during movement
- [x] Generated lore text is coherent and genre-appropriate
- [x] Chat encryption establishes keys without pre-shared secrets
- [ ] Mobile build responds to touch input for movement and aiming
- [ ] Federation hub can be self-hosted following documentation
- [ ] WASM mods execute in sandboxed environment
- [ ] Profanity filter masks offensive words in chat
- [ ] Credit economy feels balanced across 10+ playtests
- [ ] `go test ./... -cover` reports 82%+ for all packages
- [ ] CI pipeline passes all checks on Linux, macOS, Windows

## Known Gaps
- None identified — this phase resolves all gaps documented in GAPS.md

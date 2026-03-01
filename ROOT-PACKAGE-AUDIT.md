# Audit: github.com/opd-ai/violence (Root Package)
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The root package contains the main game entry point (main.go, 3459 lines) implementing the complete VIOLENCE raycasting FPS game loop using ebiten/v2. The package has extensive integration test coverage (9 test files, 5426 lines) and achieves 60.9% test coverage, below the 65% target. Critical risks include incomplete error handling, missing error wrapping, several unimplemented TODOs in core gameplay loops, and absence of package documentation.

## Issues Found
- [x] high error — No error wrapping with %w format in entire codebase (`main.go:2499`, `main.go:2241`, `main.go:3410`, `main.go:3447`, `main.go:3457`)
- [x] high error — Error swallowed in mod loader without logging (`main.go:2455`)
- [x] high stub — Player death handling unimplemented in combat loop (`main.go:939`)
- [x] high stub — Boss wave spawning unimplemented in boss arena event (`main.go:994`)
- [x] med error — Unused return values swallowed in event audio generation (`main.go:989`)
- [x] med error — Multiple unused return values from weapon upgrade stats (`main.go:1619`)
- [x] med error — Unused variables in chat rendering (`main.go:2343-2344`)
- [x] med test — Test coverage at 60.9%, below 65% target (needs 4.1% improvement)
- [x] med doc — No package-level doc.go file
- [x] med doc — Only 3 of 78 exported methods have godoc comments (96% missing documentation)
- [x] med api — Uses bare `interface{}` type instead of `any` for multiplayerMgr field (`main.go:138`)
- [x] low stub — Lockdown failure handling unimplemented (`main.go:972`)
- [x] low stub — Enemy alert state during alarm unimplemented (`main.go:963`)
- [x] low stub — Inventory system population incomplete in save system (`main.go:2491`)
- [x] low stub — Level/XP display missing from HUD (`main.go:1014`)
- [x] low stub — Particle rendering marked as placeholder implementation (`main.go:2588`)
- [x] low stub — Prop rendering uses placeholder colored dots (`main.go:2668`)
- [x] low stub — Quest objective rendering uses placeholder dots (`main.go:2786`)
- [x] low error — Inconsistent logging: mixed use of log and logrus packages (`main.go:57`, `main.go:2241`, `main.go:3410`, `main.go:3447`)

## Test Coverage
60.9% (target: 65%)

**Test Files**:
- main_test.go (95,320 bytes) - comprehensive game state tests
- build_test.go - build validation
- chat_integration_test.go - chat system integration
- config_hotreload_test.go - configuration hot-reload
- federation_discovery_integration_test.go - federation discovery
- mastery_integration_test.go - weapon mastery system
- mastery_xp_timing_test.go - XP timing validation
- minigame_visual_test.go - minigame rendering
- audit_fixes_test.go - audit fix validation

**Race Detector**: PASS (no race conditions detected)
**go vet**: PASS (no issues)

## Dependencies

**External Dependencies**:
- `github.com/hajimehoshi/ebiten/v2` - game engine framework (justified: core game rendering)
- `github.com/sirupsen/logrus` - structured logging (justified: production logging)
- `golang.org/x/image/font/basicfont` - font rendering (justified: HUD/UI text)

**Internal Dependencies**: Imports all 47 pkg/* packages plus 2 cmd/* packages
- Creates tight coupling to entire codebase
- Main orchestrates: ai, ammo, audio, automap, bsp, camera, chat, class, combat, config, crafting, destruct, door, engine, event, federation, input, inventory, lighting, loot, lore, minigame, mod, network, particle, progression, props, quest, raycaster, render, rng, save, secret, shop, skills, squad, status, texture, tutorial, ui, upgrade, weapon

**Integration Surface**: 
- Main entry point for entire application
- Implements ebiten.Game interface
- Orchestrates all game systems via Update/Draw loop
- Highest integration risk in codebase

## Recommendations
1. **[HIGH PRIORITY]** Implement error wrapping with `fmt.Errorf("%w", err)` for all error propagation to enable proper error chain inspection and debugging
2. **[HIGH PRIORITY]** Complete the 4 critical gameplay stubs to avoid runtime crashes:
   - Player death handling (main.go:939)
   - Boss wave spawning (main.go:994)
   - Lockdown failure handling (main.go:972)
   - Enemy alert state propagation (main.go:963)
3. **[MEDIUM PRIORITY]** Add package doc.go explaining:
   - Game architecture and main game loop
   - State machine transitions
   - System initialization order
   - Integration patterns with pkg/* subsystems
4. **[MEDIUM PRIORITY]** Increase test coverage from 60.9% to 65%+ by:
   - Adding tests for uncovered state transitions
   - Testing error paths (save/load failures)
   - Testing minigame state transitions
   - Testing multiplayer mode switches
5. **[MEDIUM PRIORITY]** Standardize on single logging library (recommend logrus for structured logging throughout)
6. **[LOW PRIORITY]** Replace `interface{}` with `any` for Go 1.18+ compatibility (main.go:138)
7. **[LOW PRIORITY]** Add godoc comments to all 78 exported methods (currently only 3 documented)
8. **[LOW PRIORITY]** Handle unused return values properly:
   - Log event audio generation errors (main.go:989)
   - Check all weapon upgrade stat returns (main.go:1619)
   - Remove dead variables in chat rendering (main.go:2343-2344)

## Architecture Notes
The Game struct contains 52 fields organizing game state into versioned system groups (v2.0, v3.0, v4.0, v5.0+), suggesting incremental feature additions. The Update/Draw loop follows standard ebiten patterns with state-based routing. The main() function demonstrates proper initialization sequence with configuration loading, window setup, and hot-reload support.

**Positive Patterns**:
- Clean separation of update/draw logic per game state
- Proper ebiten.Game interface implementation
- Comprehensive integration test suite
- Race-free concurrent execution (verified with -race)
- Hot-reload configuration support

**Technical Debt**:
- 8 incomplete stub implementations
- Mixed logging libraries
- Minimal method documentation
- Below-target test coverage
- No error wrapping for diagnostics

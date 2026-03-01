# Audit: github.com/opd-ai/violence/pkg/bsp
**Date**: 2026-03-01
**Status**: Complete

## Summary
The `pkg/bsp` package provides Binary Space Partitioning for procedural level generation, including both standard dungeon/level generation (`bsp.go`) and deathmatch arena generation (`deathmatch.go`). The package is well-tested (94.4% coverage), race-free, and follows Go best practices. No critical issues found; some minor improvements suggested for hardcoded constants and edge case validation.

## Issues Found
- [ ] low documentation — Missing `doc.go` package documentation file (`bsp.go:1`)
- [ ] low api-design — Unexported `sightlineMap` field in `ArenaGenerator` could be documented or removed if unused (`deathmatch.go:32`)
- [ ] low api-design — Hardcoded magic numbers in `placeDoors` (30% chance) and `placeSecrets` (15% chance) should be constants (`bsp.go:330`, `bsp.go:369`)
- [ ] low api-design — Taylor series approximations `cosApprox`/`sinApprox` have limited accuracy; consider documenting accuracy bounds or using `math.Cos`/`math.Sin` (`deathmatch.go:359-386`)
- [ ] low error-handling — No validation of width/height parameters in `NewGenerator` and `NewArenaGenerator` (could panic on invalid dimensions) (`bsp.go:62`, `deathmatch.go:36`)
- [ ] low error-handling — `placeSecrets` performs defensive bounds checks but silently returns on invalid input rather than logging (`bsp.go:340-349`)
- [ ] low test-coverage — Test result discarded with `_ =` in `TestArenaGenerator_SightlineBalancing` (`deathmatch_test.go:278`)
- [ ] med concurrency — No mutex protection on Generator/ArenaGenerator fields; not safe for concurrent use if reused (currently appears single-use) (`bsp.go:44-53`, `deathmatch.go:22-33`)

## Test Coverage
94.4% (target: 65%)

**Coverage Details:**
- Comprehensive table-driven tests for both generators
- Determinism tests verify same seed produces same output
- Connectivity/sightline balance tests
- Genre-specific tile generation tests
- Edge case tests (small/large/narrow/tall arenas)
- Race detector: PASS

## Dependencies
**Internal:**
- `github.com/opd-ai/violence/pkg/procgen/genre` - genre constants (Fantasy, SciFi, Horror, etc.)
- `github.com/opd-ai/violence/pkg/rng` - deterministic random number generation

**External:** None (stdlib only)

**Imported By:**
- `main.go` - main game entry point
- `pkg/audio` - reverb/audio spatial calculations

**Dependency Health:** Clean, minimal dependencies. No circular imports detected.

## Recommendations
1. Add `doc.go` with package-level documentation explaining BSP algorithm and use cases
2. Add input validation to constructors (`NewGenerator`, `NewArenaGenerator`) to reject invalid dimensions (e.g., width/height <= 0 or too small for MinSize)
3. Extract magic numbers to named constants: `DoorPlacementChance = 30`, `SecretPlacementChance = 15`
4. Document concurrency safety (currently not thread-safe for reuse; generators appear designed for single-use)
5. Consider replacing Taylor series trig approximations with `math.Cos`/`math.Sin` for accuracy unless performance profiling shows bottleneck

# Audit: github.com/opd-ai/violence/pkg/squad
**Date**: 2026-03-01
**Status**: Complete

## Summary
Squad package manages AI-controlled squad members with formation patterns, behavior states, and co-op player integration. Overall code health is excellent with 100% test coverage and no race conditions. Minor issues identified relate to error handling consistency and API design clarity.

## Issues Found
- [ ] low documentation — Missing package-level `doc.go` file (`pkg/squad/`)
- [ ] low api-design — `AddMember` silently ignores errors when at max capacity instead of returning meaningful error (`squad.go:94`)
- [ ] med error-handling — `SetGenre` function is exported but only delegates to `ai.SetGenre` with no error handling or validation (`squad.go:431`)
- [ ] low api-design — Duplicate `Formation` type definitions: `Formation` in `squad.go:24` and `FormationType` in `formation.go:6` create naming confusion
- [ ] low documentation — `updateFormation` method comment doesn't explain the spacing calculations (`squad.go:310`)
- [ ] med concurrency — No mutex protection on `Squad` struct fields; concurrent access to `Members`, `HumanPlayers`, `Behavior`, or `Formation` could cause race conditions in multi-goroutine scenarios (`squad.go:63`)
- [ ] low api-design — `Update` method has unused `playerX, playerY` parameters when behavior is not `BehaviorAttack` (`squad.go:195`)
- [ ] low error-handling — `updateFollow` and `updateHold` don't handle pathfinding failures (`squad.go:217`, `squad.go:268`)

## Test Coverage
100.0% (target: 65%)

## Dependencies
**Internal dependencies:**
- `github.com/opd-ai/violence/pkg/ai` — AI agents and behavior trees
- `github.com/opd-ai/violence/pkg/rng` — Random number generation

**External dependencies:**
- `math` (stdlib) — Mathematical calculations for formations and positioning

**Integration points:**
- Combat system (via AI agent integration)
- Pathfinding system (A* via `ai.FindPath`)
- Network/multiplayer (human player coordination)

## Recommendations
1. Add mutex protection for concurrent access to Squad fields (methods are not goroutine-safe)
2. Consolidate `Formation` and `FormationType` into single type definition
3. Make `AddMember` return proper error when at max capacity instead of silently ignoring
4. Add package-level `doc.go` with usage examples
5. Consider validating/handling pathfinding edge cases in movement updates

# Audit: github.com/opd-ai/violence/pkg/engine
**Date**: 2026-03-01
**Status**: Complete

## Summary
The `pkg/engine` package provides a minimal ECS (Entity-Component-System) framework with bitmask-based archetype matching. Overall health is excellent with 91.5% test coverage and clean code. Minor documentation gaps and a consistency issue with dual global/instance genre tracking reduce maintainability.

## Issues Found
- [ ] low documentation — Missing `doc.go` for package-level overview
- [ ] low documentation — `newEntityIterator` is unexported but lacks justification comment (`query.go:59`)
- [ ] med api-design — Duplicate genre tracking with both instance (`World.genre`) and global (`currentGenre`) variables creates confusion (`engine.go:25`, `engine.go:129`)
- [ ] low concurrency — `World` struct has no documented concurrency safety guarantees; maps are not thread-safe if accessed concurrently
- [ ] low api-design — `RemoveEntity` does not clean up archetype bitmask entry (`engine.go:85`)

## Test Coverage
91.5% (target: 65%) ✓ EXCEEDS TARGET

## Dependencies
- `reflect` (standard library) — Used for reflection-based component type identification
- `math` (standard library) — Used for camera FOV calculations

No external dependencies. All imports justified and minimal.

## Recommendations
1. Add archetype cleanup to `RemoveEntity` to prevent memory leak in `w.archetypes` map
2. Consolidate genre tracking: remove either global `currentGenre` or instance `World.genre` to reduce API confusion
3. Document concurrency model: specify whether `World` is thread-safe or requires external synchronization
4. Add `doc.go` with package-level overview of ECS architecture and usage examples
5. Document why `newEntityIterator` is unexported (appears to be internal factory)

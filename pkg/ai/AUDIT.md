# Audit: github.com/opd-ai/violence/pkg/ai
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
Core AI package implementing behavior trees, pathfinding, cover system, and procedural sprite generation. Code is well-structured with excellent test coverage (91.4%), but suffers from concurrency-unsafe global state and lacks context-aware error handling. Critical for game AI behavior with high integration surface (imports: level, rng; imported by: combat, engine, network).

## Issues Found
- [x] high concurrency — Global mutable state `currentGenre` not protected by mutex (`ai.go:597`)
- [x] high concurrency — `SetGenre()` function modifies global state unsafely in concurrent environment (`ai.go:600`)
- [x] med api — `FindPath()` returns fallback path instead of empty slice, masking pathfinding failures (`ai.go:515`)
- [x] med api — `Selector.current` and `Sequence.current` fields are unexported but maintain state across Tick calls, potentially causing issues with concurrent BehaviorTree execution (`ai.go:30, 58`)
- [x] low documentation — Missing package-level `doc.go` file
- [x] low api — `archetypes` map is package-global and immutable, could be const or function-scoped (`ai.go:544`)
- [x] low performance — `FindPath()` uses linear search in openSet (O(n) per iteration), should use heap/priority queue (`ai.go:467-475`)
- [x] low performance — A* pathfinding has hardcoded `maxIter=500` which may be too low for large maps (`ai.go:465`)
- [x] low code-clarity — Magic numbers for animation offsets not documented (`ai.go:303`, `sprite_gen.go:81-88`)
- [x] low testing — No benchmark tests for pathfinding performance on large maps

## Test Coverage
91.4% (target: 65%) ✓ EXCEEDS TARGET

## Dependencies
- **Internal**: `pkg/level` (TileMap interface), `pkg/rng` (deterministic RNG)
- **External**: `image`, `image/color`, `math`, `math/rand` (stdlib only)
- **Imported By**: Combat system, game engine, network synchronization

## Recommendations
1. **CRITICAL**: Protect `currentGenre` with `sync.RWMutex` or refactor to context-based genre selection
2. **HIGH**: Make `FindPath()` return `[]Waypoint{}` on failure instead of fallback path, add error return
3. **MED**: Document behavior tree state management or add Reset() methods to Selector/Sequence
4. **MED**: Implement priority queue for A* openSet using `container/heap`
5. **LOW**: Add `doc.go` with package overview and usage examples

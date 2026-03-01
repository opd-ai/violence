# Audit: github.com/opd-ai/violence/pkg/progression
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The progression package manages player XP and leveling with a simple API. Implementation is minimalistic but has critical design flaws: incomplete leveling logic (no automatic level-up on XP threshold), global mutable state for genre configuration causing concurrency issues, and missing validation/bounds checking. Test coverage is good at 80%.

## Issues Found
- [ ] high API Design — `AddXP` does not trigger automatic level-up when XP threshold reached; manual `LevelUp()` call required breaks encapsulation (`progression.go:16-18`)
- [ ] high Concurrency Safety — Global mutable `currentGenre` variable not protected by mutex, causes race conditions in concurrent access (`progression.go:25`)
- [ ] med Stub/Incomplete Code — `SetGenre` function accepts genre string but does nothing with it (no validation, no curve configuration) (`progression.go:28-30`)
- [ ] med Error Handling — `AddXP` accepts negative values without validation or error return, can result in negative XP (`progression.go:16`)
- [ ] med API Design — Exported fields `XP` and `Level` allow external mutation bypassing business logic (`progression.go:5-7`)
- [ ] low Documentation — No `doc.go` file for package-level documentation
- [ ] low API Design — Missing `GetXP()` and `GetLevel()` accessor methods for encapsulated data access
- [ ] low API Design — No `XPForNextLevel()` method to query leveling thresholds
- [ ] low Error Handling — `LevelUp` allows calling beyond reasonable max level without bounds checking (`progression.go:21-23`)
- [ ] low Documentation — `SetGenre` purpose unclear without implementation details or comment explaining future curve logic (`progression.go:27-30`)

## Test Coverage
80.0% (target: 65%) ✓

## Dependencies
**Internal**: None (zero imports)
**External**: Standard library only (no external dependencies)
**Integration Points**: Used by `main.go` and `test/genre_cascade_test.go`

## Recommendations
1. Implement XP threshold logic in `AddXP()` to auto-level when thresholds reached; remove public `LevelUp()` or make it private
2. Remove global `currentGenre` variable OR protect with `sync.RWMutex` if multi-genre support needed
3. Make `XP` and `Level` fields private, add getter methods, validate inputs in `AddXP`
4. Implement or remove `SetGenre` stub function based on actual requirements
5. Add `doc.go` with package documentation and design rationale

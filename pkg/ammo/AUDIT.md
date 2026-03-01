# Audit: github.com/opd-ai/violence/pkg/ammo
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
Simple ammunition tracking system with Pool-based management. Code is well-tested (90% coverage) and race-free, but has concurrency safety gaps with mutable global state, negative value handling issues, and missing input validation.

## Issues Found
- [x] high concurrency — Global mutable state `currentGenre` not protected by mutex, creates race conditions in multi-goroutine access (`ammo.go:46`)
- [x] high validation — `Add()` accepts negative amounts allowing integer underflow attacks (`ammo.go:23`)
- [x] high validation — `Set()` accepts negative values, allowing invalid ammo counts (`ammo.go:42`)
- [x] med validation — `Consume()` accepts negative amounts, bypassing consumption logic (`ammo.go:28`)
- [x] med documentation — Missing package-level `doc.go` file for godoc browsing
- [x] med api-design — Exported `Pool.counts` map through unprotected direct field access in tests (`ammo_test.go:35`)
- [x] low documentation — `SetGenre()` function has no effect on ammo behavior, unclear purpose (`ammo.go:48`)
- [x] low testing — No benchmark tests for performance-critical `Consume()` operations
- [x] low error-handling — `Get()` returns 0 for missing keys, indistinguishable from legitimate zero values (`ammo.go:37`)

## Test Coverage
90.0% (target: 65%) ✓

## Dependencies
**Internal**: None  
**External**: None (stdlib only)  
**Integration Points**: Used by main.go and test/genre_cascade_test.go

## Recommendations
1. Add sync.RWMutex for `currentGenre` global state to prevent race conditions
2. Validate amounts in `Add()`, `Set()`, `Consume()` — reject negative values or document negative semantics
3. Add input validation to prevent integer overflow/underflow attacks
4. Create `doc.go` with package overview and usage examples
5. Consider returning error/bool from `Get()` to distinguish missing vs zero ammo

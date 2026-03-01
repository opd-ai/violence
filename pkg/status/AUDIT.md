# Audit: github.com/opd-ai/violence/pkg/status
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The status package manages status effects applied to entities with genre-based configuration. Code contains critical stub implementations and concurrency safety issues with global state. Test coverage meets target but tests only verify structure, not business logic.

## Issues Found
- [x] high stub/incomplete — Apply() method has empty body, no actual effect application (`status.go:24`)
- [x] high stub/incomplete — Tick() method has empty body, no effect progression logic (`status.go:27`)
- [x] high concurrency — Global variable `currentGenre` not protected by mutex, race condition risk (`status.go:29`)
- [x] med api-design — Registry.Apply() takes string but no validation or error return for unknown effects (`status.go:24`)
- [x] med api-design — Registry lacks method to register/lookup effects in the map (`status.go:14-16`)
- [x] med error-handling — No error return types on Apply/Tick, silent failures possible (`status.go:24,27`)
- [x] low documentation — Missing doc.go file for package overview
- [x] low api-design — Effect struct fields all exported but no validation methods (`status.go:7-11`)
- [x] low test — Tests only verify no-panic behavior, not actual effect application logic (`status_test.go:52-84`)
- [x] low api-design — SetGenre/GetCurrentGenre use global state instead of context-based config (`status.go:32,37`)

## Test Coverage
66.7% (target: 65%) ✓

## Dependencies
**Standard Library Only**:
- `time` — Duration tracking for status effects

**Integration Points**:
- Referenced by `main.go` — main game initialization
- Referenced by `test/genre_cascade_test.go` — genre system testing

**No External Dependencies**: Clean dependency footprint using only Go standard library.

## Recommendations
1. **CRITICAL**: Implement Apply() and Tick() method bodies to provide actual status effect functionality
2. **CRITICAL**: Add mutex protection for `currentGenre` global variable or refactor to instance-based configuration
3. **HIGH**: Add Registry methods for registering effects and retrieving active effects with proper error handling
4. **MEDIUM**: Add error returns to Apply() for validation failures and unknown effect names
5. **LOW**: Create doc.go with package-level documentation and usage examples

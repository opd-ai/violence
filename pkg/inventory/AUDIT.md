# Audit: github.com/opd-ai/violence/pkg/inventory
**Date**: 2026-03-01
**Status**: Complete

## Summary
The inventory package manages player item storage with support for active-use items and quick-slot functionality. Code quality is excellent with 94.6% test coverage, thread-safe operations via mutexes, and comprehensive error handling. Minor concerns include global mutable state for genre configuration and potential race-condition risk in getter methods returning pointer to internal data.

## Issues Found
- [ ] **low** API Design — `Get()` returns pointer to internal slice element, allowing external mutation (`inventory.go:204`)
- [ ] **low** Concurrency Safety — Global `currentGenre` variable is not protected by mutex for concurrent access (`inventory.go:312`)
- [ ] **low** Documentation — Missing package-level `doc.go` file for godoc
- [ ] **med** API Design — `SetGenre()` and `GetCurrentGenre()` use global mutable state instead of instance-based configuration (`inventory.go:315-322`)
- [ ] **low** Error Handling — `SetGenre()` has no validation or error return for invalid genre strings (`inventory.go:315`)

## Test Coverage
94.6% (target: 65%)

## Dependencies
**Standard Library Only:**
- `fmt` — Error formatting with context wrapping
- `sync` — RWMutex for thread-safe inventory operations

**Integration Points:**
- Used by `pkg/network` (coop player state)
- Referenced by `pkg/door` (keycard inventory pattern)
- Referenced by `pkg/shop` (shop inventory management)

## Recommendations
1. Change `Get()` to return value copy instead of pointer: `func (inv *Inventory) Get(id string) Item` to prevent external mutation
2. Replace global `currentGenre` with instance field on `Inventory` or remove if unused (currently not integrated with any inventory logic)
3. Add package-level `doc.go` with overview of inventory system and usage examples
4. Add genre validation in `SetGenre()` or document valid genre values
5. Consider adding `GetQuantity(id string) int` convenience method to avoid pointer access pattern

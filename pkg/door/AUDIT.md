# Audit: github.com/opd-ai/violence/pkg/door
**Date**: 2026-03-01
**Status**: Complete

## Summary
Package door implements keycard-locked doors with state-based animation and genre-specific theming. Code quality is excellent with 91.3% test coverage, no race conditions, and comprehensive table-driven tests. Minor issues found relate to global state management and missing doc.go file.

## Issues Found
- [ ] low documentation — No doc.go file for package overview (`door.go:1`)
- [ ] low design — Global variable `currentGenre` creates mutable package state, not concurrency-safe if modified during runtime (`door.go:194`)
- [ ] low design — `genreKeycardNames` and `genreDoorTypes` are package-level global maps, could be encapsulated in a genre configuration struct (`door.go:158-192`)
- [ ] med api — Legacy `TryOpen` function (door.go:227-232) bypasses DoorSystem abstraction, should be deprecated in favor of DoorSystem.TryOpen
- [ ] low documentation — `AnimationSpeed` field in Door struct is documented but never used in Update() method (`door.go:35`)
- [ ] low concurrency — KeycardInventory.keycards map is not concurrency-safe, no mutex protection if accessed from multiple goroutines (`door.go:40-42`)
- [ ] low concurrency — DoorSystem.doors slice has no concurrency protection for concurrent AddDoor/Update calls (`door.go:72-73`)

## Test Coverage
91.3% (target: 65%)

## Dependencies
Zero external dependencies - uses only Go standard library.

## Recommendations
1. Add doc.go file with package overview and usage examples
2. Deprecate legacy TryOpen function, document DoorSystem.TryOpen as canonical API
3. Consider refactoring genre configuration into a Genre struct to eliminate global mutable state
4. Add mutex protection to KeycardInventory and DoorSystem if concurrent access is expected
5. Document or remove unused AnimationSpeed field from Door struct

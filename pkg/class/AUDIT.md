# Audit: github.com/opd-ai/violence/pkg/class
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The class package defines player character classes with a simple struct and constants. The implementation is minimal and contains critical stub functionality in `GetClass()` that returns incomplete data. Package has adequate test coverage (66.7%) but lacks proper class data initialization and genre-specific class definitions despite having genre support.

## Issues Found
- [x] high stub — `GetClass()` returns incomplete Class with only ID set, missing Name/Health/Speed (`class.go:20`)
- [x] high concurrency — Global mutable state `currentGenre` not protected by mutex, potential race condition (`class.go:24`)
- [x] med api — `GetClass()` should return error for unknown class IDs instead of stub data (`class.go:20`)
- [x] med design — No class registry or data store, defeats purpose of `GetClass()` function (`class.go:20`)
- [x] med documentation — Package missing `doc.go` file for godoc overview
- [x] low design — Constants exported but no validation function to check valid class IDs
- [x] low design — `SetGenre()` accepts any string without validation or effect on `GetClass()` behavior (`class.go:27`)
- [x] low documentation — `Class` struct fields lack godoc comments (`class.go:12-17`)

## Test Coverage
66.7% (target: 65%) — PASS

## Dependencies
- No external dependencies (standard library only)
- Integration points: Likely used by player/progression systems (no imports found in quick scan)

## Recommendations
1. Implement proper class data registry with Name/Health/Speed values indexed by ID
2. Add mutex protection for `currentGenre` global variable or redesign as parameter
3. Return `(Class, error)` from `GetClass()` for unknown class IDs
4. Create `doc.go` with package overview
5. Implement genre-specific class data or remove unused genre functionality

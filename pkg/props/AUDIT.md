# Audit: github.com/opd-ai/violence/pkg/props
**Date**: 2026-03-01
**Status**: Complete

## Summary
The props package manages decorative game world objects with genre-specific placement logic. Code quality is excellent with 96.3% test coverage, comprehensive concurrency safety, and no critical issues. Legacy compatibility functions exist but are documented as such.

## Issues Found
- [ ] low API Design — Legacy function `SetGenre` has no-op implementation without documentation explaining deprecation (`props.go:290`)
- [ ] low Documentation — `PropPlant` type defined but unused in any genre templates (`props.go:20`)
- [ ] low Error Handling — `PlaceProps` returns empty slice silently when genre has no templates; could log warning (`props.go:162`)
- [ ] med API Design — `GetPropsByType` returns internal slice pointers instead of copies, inconsistent with `GetProps` (`props.go:260-269`)
- [ ] low Documentation — Package lacks `doc.go` file for godoc package overview

## Test Coverage
96.3% (target: 65%)

## Dependencies
**External**: 
- `github.com/opd-ai/violence/pkg/procgen/genre` (genre constants)
- `github.com/opd-ai/violence/pkg/rng` (deterministic random generation)

**Standard Library**: `sync` (concurrency safety)

**Integration**: No reverse dependencies found; package appears unused in current codebase

## Recommendations
1. Add godoc comment explaining `SetGenre` legacy function is no-op and users should use `Manager.SetGenre`
2. Make `GetPropsByType` return copies to match `GetProps` behavior and prevent accidental mutation
3. Add `doc.go` with package overview and usage examples
4. Either implement `PropPlant` in genre templates or remove unused constant
5. Log warning when `PlaceProps` called with uninitialized genre to aid debugging

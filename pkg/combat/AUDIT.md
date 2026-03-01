# Audit: github.com/opd-ai/violence/pkg/combat
**Date**: 2026-03-01
**Status**: Complete

## Summary
Core damage calculation and spatial collision detection package. Code quality is high with 100% test coverage, comprehensive table-driven tests, and benchmarks. The package has minimal dependencies (only `math` stdlib) and demonstrates strong Go practices. Primary concerns are global mutable state and missing package-level documentation file.

## Issues Found
- [ ] med API Design — Global mutable state via `globalSystem` variable exposes concurrency risks (`combat.go:124`)
- [ ] med API Design — Package-level functions `Apply`, `SetGenre`, `SetDifficulty` modify shared state unsafely (`combat.go:127-139`)
- [ ] low Documentation — Missing `doc.go` file for comprehensive package documentation
- [ ] low API Design — `Apply` function ignores `DamageEvent` position fields, uses hardcoded zeros (`combat.go:128`)
- [ ] low Error Handling — No validation for negative `cellSize` in `NewSpatialHash` (`spatial_hash.go:24`)
- [ ] low Error Handling — No validation for `SetDifficulty` parameter bounds (could be negative) (`combat.go:57`)

## Test Coverage
100.0% (target: 65%)

## Dependencies
**External**: None  
**Standard Library**: `math` only  
**Integration Points**: Designed as standalone combat calculation engine, likely consumed by game engine or entity systems (currently 0 internal importers detected)

## Recommendations
1. **Remove global state**: Deprecate package-level `Apply`, `SetGenre`, `SetDifficulty` functions in favor of instance-based `System` API to ensure thread-safety
2. **Add input validation**: Validate `cellSize > 0` in `NewSpatialHash` and `difficulty >= 0` in `SetDifficulty`
3. **Create doc.go**: Add package-level documentation explaining combat system architecture, damage calculation algorithm, and spatial hash usage patterns
4. **Fix Apply function**: Either use `DamageEvent` position fields or remove them from the struct to avoid API confusion
5. **Document concurrency**: Add godoc comments clarifying that `System` instances are not thread-safe and should not be shared across goroutines

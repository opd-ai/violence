# Audit: github.com/opd-ai/violence/pkg/lighting
**Date**: 2026-03-01
**Status**: Complete

## Summary
The lighting package provides dynamic lighting calculations for the game engine, including point lights, cone lights (flashlights), and sector-based light mapping. Code quality is excellent with 98.8% test coverage, comprehensive genre presets, and proper mathematical implementations. Two stub functions in `lighting.go` need implementation or removal, and some minor inconsistencies exist with genre constant usage.

## Issues Found
- [ ] high stub — Stub function `Calculate()` in Sector struct has empty body (`lighting.go:18`)
- [ ] high stub — Stub function `SetGenre()` has empty body (`lighting.go:21`)
- [ ] med api — `genreAmbientLevel()` uses hardcoded strings instead of genre package constants (`sector.go:176-189`)
- [ ] med api — Duplicate `clamp()` function definitions in `sector.go:141` and similar logic in `point.go:187` (`clampColor`)
- [ ] low docs — Missing package-level `doc.go` file
- [ ] low api — Exported `Sector` struct in `lighting.go` appears unused, replaced by `SectorLightMap`

## Test Coverage
98.8% (target: 65%)

## Dependencies
- `math` (standard library) — Mathematical operations for lighting calculations
- `github.com/opd-ai/violence/pkg/procgen/genre` — Genre constants for preset selection
- `github.com/opd-ai/violence/pkg/rng` — Deterministic RNG for flicker effects

No circular dependencies detected. External dependencies are minimal and justified.

## Recommendations
1. Remove or implement stub functions `Sector.Calculate()` and `SetGenre()` in `lighting.go:18,21`
2. Replace hardcoded genre strings in `genreAmbientLevel()` with `genre.Fantasy`, `genre.SciFi`, etc. constants
3. Consolidate duplicate `clamp()` utility functions into a single exported helper or use standard library alternative
4. Add package-level `doc.go` file with overview and usage examples
5. Consider deprecating unused exported `Sector` struct or documenting its purpose vs `SectorLightMap`

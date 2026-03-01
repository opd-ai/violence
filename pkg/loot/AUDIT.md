# Audit: github.com/opd-ai/violence/pkg/loot
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The loot package implements item drop mechanics with loot tables and secret reward generation. Core functionality is partially complete with `Roll()` method stubbed, but package has good test coverage (88%) and demonstrates proper use of deterministic RNG. Currently used only in main.go and test code, suggesting incomplete integration.

## Issues Found
- [x] high stub/incomplete — `Roll()` method returns only `nil`, core loot drop logic unimplemented (`loot.go:43`)
- [x] high concurrency — Global mutable state `currentGenre` not protected by mutex, race condition risk (`loot.go:118`)
- [x] med api-design — `SecretLootTable.rng` field is unused, constructor accepts RNG but uses new instance in `GenerateSecretReward()` (`loot.go:34`, `loot.go:77`)
- [x] med documentation — Package lacks `doc.go` file for comprehensive godoc overview
- [x] med error-handling — `Drop.Chance` field has no validation; values >1.0 or <0.0 accepted without error (`loot.go:21`)
- [x] low api-design — `Rarity` type lacks `String()` method for debugging/logging
- [x] low documentation — `Drop` struct fields lack godoc comments (`loot.go:19-22`)
- [x] low documentation — `Rarity` constants lack individual godoc comments (`loot.go:12-16`)
- [x] low api-design — No method to remove items from `SecretLootTable` after adding via `AddItem()`

## Test Coverage
88.0% (target: 65%) — EXCEEDS TARGET

## Dependencies
**Internal**: `pkg/rng` (proper use of project RNG)
**External**: None (standard library only)
**Importers**: `main.go`, `test/genre_cascade_test.go` (low integration suggests incomplete feature)

## Recommendations
1. **CRITICAL**: Implement `Roll()` method with proper drop chance calculation
2. **CRITICAL**: Add mutex protection for `currentGenre` global state or refactor to context-based design
3. Fix `SecretLootTable.rng` field usage inconsistency — either use injected RNG or remove field
4. Add validation for `Drop.Chance` field (0.0-1.0 range)
5. Create `doc.go` with package overview and usage examples

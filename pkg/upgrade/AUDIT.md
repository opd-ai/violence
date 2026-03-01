# Audit: github.com/opd-ai/violence/pkg/upgrade
**Date**: 2026-03-01
**Status**: Complete

## Summary
The upgrade package implements weapon upgrade mechanics with token-based currency and genre-specific naming. Code is well-structured with solid test coverage (86%), clean API design, and no concurrency issues. Minor documentation gap exists (missing doc.go).

## Issues Found
- [ ] low documentation — Missing doc.go file for package-level documentation
- [ ] low documentation — Unexported field `genreName` lacks godoc comment (`upgrade.go:32`)
- [ ] low api-design — UpgradeType constants lack godoc comments (`upgrade.go:12-17`)
- [ ] low error-handling — Manager.GetUpgrades returns nil for non-existent weapons; empty slice would be more idiomatic (`upgrade.go:202`)
- [ ] low testing — Missing negative test case for negative token amounts in Add/Spend

## Test Coverage
86.0% (target: 65%)

## Dependencies
- github.com/opd-ai/violence/pkg/procgen/genre (single external dependency, justified for genre-specific naming)

## Recommendations
1. Add doc.go file with package-level documentation and usage examples
2. Change Manager.GetUpgrades to return empty slice instead of nil for consistency
3. Add godoc comments for UpgradeType constants
4. Add edge case tests for negative values in token operations
5. Document unexported genreName field rationale or consider making it exported

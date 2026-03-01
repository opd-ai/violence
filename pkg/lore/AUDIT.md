# Audit: github.com/opd-ai/violence/pkg/lore
**Date**: 2026-03-01
**Status**: Complete

## Summary
Procedural narrative generation package implementing template-based and Markov chain text generation for in-game lore codex. Well-tested (96.1% coverage), thread-safe, with comprehensive godoc. Uses deprecated `strings.Title` and insecure `math/rand` for deterministic generation.

## Issues Found
- [ ] low documentation — Missing doc.go file for package-level examples
- [ ] low api_design — `strings.Title` deprecated in grammar.go:261, should use `cases.Title(language.English)` for consistency with lore.go
- [ ] low dependencies — `math/rand` used instead of `crypto/rand` (acceptable for deterministic procedural generation, not security-critical)

## Test Coverage
96.1% (target: 65%)

## Dependencies
**External**: `golang.org/x/text/cases`, `golang.org/x/text/language`
**Standard**: `fmt`, `math/rand`, `strings`, `sync`
**Integration**: No internal package imports detected (self-contained)

## Recommendations
1. Add doc.go with package overview and cross-referencing both generation approaches
2. Replace deprecated `strings.Title` with `cases.Title(language.English)` in grammar.go:261
3. Consider documenting that `math/rand` is intentionally used for deterministic generation (not a security issue)

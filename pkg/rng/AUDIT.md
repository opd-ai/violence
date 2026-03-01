# Audit: github.com/opd-ai/violence/pkg/rng
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The `pkg/rng` package provides a seed-based random number generator wrapper around `math/rand/v2`. The package is minimal, well-tested (100% coverage), and widely used (17+ import sites). However, it lacks concurrency safety mechanisms for shared RNG instances, which poses a critical race condition risk in concurrent usage scenarios.

## Issues Found
- [x] high concurrency — RNG struct not thread-safe; concurrent calls to Intn/Float64/Seed can cause data races (`rng.go:7-8,17-18,22-23,27-28`)
- [x] high concurrency — Seed() method can cause race when called concurrently with Intn/Float64 by reassigning g.r field (`rng.go:27-28`)
- [x] med api-design — Missing common RNG methods: IntRange(min,max), Shuffle, Perm, NormFloat64 limit utility (`rng.go:1-30`)
- [x] med documentation — No package-level doc.go file explaining thread-safety expectations and usage patterns (`rng.go:1`)
- [x] low documentation — NewRNG godoc doesn't explain the stream parameter derivation (seed^0xda3e39cb94b95bdb) (`rng.go:12-14`)
- [x] low api-design — Intn panics on n <= 0 (inherited from rand.IntN) but not documented in godoc (`rng.go:17-19`)
- [x] low testing — No test coverage for concurrent access patterns despite shared state vulnerability (`rng_test.go:1-137`)
- [x] low testing — Missing edge case tests: Intn(1), Intn with negative/zero values, nil pointer checks (`rng_test.go:1-137`)

## Test Coverage
100.0% (target: 65%) ✓

## Dependencies
**Standard Library Only:**
- `math/rand/v2` — Go's modern random number generator (PCG algorithm)

**Integration Points:**
- Used by 17+ packages: `pkg/ai`, `pkg/loot`, `pkg/quest`, `pkg/squad`, `pkg/props`, `pkg/lighting`, `pkg/federation`, `main.go`, etc.
- High integration surface suggests many potential concurrent usage sites

## Recommendations
1. **CRITICAL**: Add mutex protection for concurrent safety or document that instances are not thread-safe and must not be shared
2. Add package-level `doc.go` with thread-safety contract and example showing proper per-goroutine RNG usage pattern
3. Extend API with `IntRange(min, max int)`, `Shuffle(slice)`, and other common RNG utilities to reduce code duplication across consumers
4. Add benchmark tests for concurrent scenarios and document performance characteristics
5. Document Intn panic behavior for n <= 0 in godoc comments

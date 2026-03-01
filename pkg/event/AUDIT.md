# Audit: github.com/opd-ai/violence/pkg/event
**Date**: 2026-03-01
**Status**: Complete

## Summary
The event package implements a world event and trigger system with alarm triggers, timed lockdowns, and boss arena events. Code quality is excellent with 96.8% test coverage, proper concurrency safety, and comprehensive genre-based procedural generation. No critical issues found; only minor improvements suggested.

## Issues Found
- [ ] low API Design — Event and Trigger structs unused/incomplete (`event.go:19`, `event.go:28`)
- [ ] low Stub Code — Fire() function is no-op stub (`event.go:203`)
- [ ] low Documentation — Missing package-level doc.go file
- [ ] low Concurrency — Global currentGenre could use atomic.Value for lock-free reads (`event.go:206-222`)
- [ ] low Dependency — Using deprecated math/rand instead of math/rand/v2 (Go 1.22+) (`event.go:5`)

## Test Coverage
96.8% (target: 65%) ✓

## Dependencies
**External**: None (stdlib only)
**Internal**: None
**Stdlib**: math/rand, sync

**Integration Points**:
- Currently minimal external usage (1 import found in codebase)
- Designed for use by combat, AI, and progression systems
- Audio generation integrates with pkg/audio (via AudioSting type)

## Recommendations
1. **Remove or implement unused types** — Event and Trigger structs appear to be design artifacts; either implement them or remove to reduce API surface
2. **Implement or document Fire() stub** — Either implement event dispatch or mark as reserved for future use
3. **Add doc.go** — Include package overview with usage examples for godoc
4. **Upgrade to math/rand/v2** — Use newer deterministic API (crypto/rand if non-deterministic needed)
5. **Consider atomic.Value for genre** — Eliminates mutex overhead for high-frequency GetGenre() reads

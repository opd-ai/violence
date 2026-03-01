# Audit: github.com/opd-ai/violence/pkg/minigame
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
The minigame package provides four interactive mini-game implementations (HackGame, LockpickGame, CircuitTraceGame, BypassCodeGame) with genre-specific selection. Code is well-structured with excellent test coverage (90.2%), but has critical security and quality issues related to deprecated RNG, stub implementation, and missing documentation.

## Issues Found
- [x] high security — Using deprecated `math/rand` instead of `crypto/rand` or `math/rand/v2` for game generation (`minigame.go:5`, `minigame.go:30`, `minigame.go:118`, `minigame.go:222`, `minigame.go:365`)
- [x] high stub — `SetGenre()` function is empty stub with no implementation (`minigame.go:201`)
- [x] med documentation — Missing package-level `doc.go` file for godoc
- [x] med documentation — Exported fields in game structs lack godoc comments (`minigame.go:19-26`, `minigame.go:104-114`, `minigame.go:206-218`, `minigame.go:353-361`)
- [x] med error-handling — No validation or error returns for invalid difficulty values in constructors (`minigame.go:29`, `minigame.go:117`, `minigame.go:221`, `minigame.go:364`)
- [x] low documentation — `Input` method lacks godoc comment (`minigame.go:55`)
- [x] low documentation — `Advance` method lacks godoc comment (`minigame.go:143`)
- [x] low documentation — `Attempt` method lacks godoc comment (`minigame.go:155`)
- [x] low documentation — `Move` method lacks godoc comment (`minigame.go:269`)
- [x] low documentation — `InputDigit` method lacks godoc comment (`minigame.go:390`)
- [x] low documentation — `Clear` method lacks godoc comment (`minigame.go:430`)

## Test Coverage
90.2% (target: 65%) ✓ EXCEEDS TARGET

## Dependencies
- **Standard library**: `math/rand` (deprecated for non-cryptographic use in Go 1.20+)
- **Importers**: `main.go`, `main_test.go`, `minigame_visual_test.go`
- **Zero external dependencies** (good)

## Recommendations
1. **CRITICAL**: Replace `math/rand` with `math/rand/v2` throughout package (Go 1.22+) or document why legacy RNG is acceptable for game mechanics
2. **HIGH**: Implement `SetGenre()` function or remove if unnecessary (breaking API change)
3. **MEDIUM**: Add package-level `doc.go` with usage examples showing genre-specific mini-game selection
4. **MEDIUM**: Add godoc comments to all exported struct fields explaining their purpose and valid ranges
5. **LOW**: Add input validation to constructors (negative difficulty, nil seed handling)

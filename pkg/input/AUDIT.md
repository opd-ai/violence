# Audit: github.com/opd-ai/violence/pkg/input
**Date**: 2026-03-01
**Status**: Complete

## Summary
The input package provides comprehensive keyboard, mouse, gamepad, and touch input handling for the Violence game. Code quality is excellent with strong test coverage (83.8%), well-documented APIs, and clean separation of concerns across input devices. Minor issues include a stub function and potential concurrency concerns in the shared Manager state.

## Issues Found
- [ ] low stub — `SetGenre()` function is a stub with no implementation (`input.go:285-288`)
- [ ] med concurrency — Manager state (bindings, gamepadButtons, mouse tracking) not protected by mutex in concurrent access scenarios (`input.go:68-77`)
- [ ] low error-handling — `SaveBindings()` error from `config.Save()` not logged, caller may miss critical failures (`input.go:268`)
- [ ] low documentation — TouchInputManager.Update() lacks godoc comment explaining touch routing logic (`touch.go:233`)
- [ ] low api-design — VirtualJoystick.HandleTouch() hardcodes screen width threshold (300px) instead of using normalized coordinates (`touch.go:64`)

## Test Coverage
83.8% (target: 65%) ✓

**Coverage breakdown:**
- `input.go`: Comprehensive unit tests with table-driven approach, benchmarks included
- `touch.go`: All touch controls thoroughly tested including edge cases
- `touch_render.go`: All rendering styles and color schemes tested

**Strengths:**
- Table-driven tests for bindings and touch controls
- Tests cover all quadrants of joystick movement
- Screen size independence tests for touch buttons
- Benchmarks for performance-critical paths
- Genre-specific rendering validation

## Dependencies
**External:**
- `github.com/hajimehoshi/ebiten/v2` — Game engine (input APIs, rendering)
- `github.com/hajimehoshi/ebiten/v2/inpututil` — Input utilities for "just pressed" detection
- `github.com/hajimehoshi/ebiten/v2/vector` — Vector graphics rendering

**Internal:**
- `github.com/opd-ai/violence/pkg/config` — Configuration persistence for key bindings

**Integration points:**
- Used by: `main.go`, `pkg/ui`
- Provides input abstraction for entire game engine

## Recommendations
1. Add mutex protection to Manager for thread-safe concurrent access to bindings and state
2. Implement SetGenre() or remove if not needed for genre-specific input customization
3. Add structured logging for SaveBindings() failures to aid debugging
4. Replace hardcoded 300px threshold with normalized screen percentage in VirtualJoystick.HandleTouch()
5. Add godoc comments to TouchInputManager.Update() explaining touch event routing priority

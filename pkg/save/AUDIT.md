# Audit: github.com/opd-ai/violence/pkg/save
**Date**: 2026-03-01
**Status**: Complete

## Summary
The `pkg/save` package handles game state persistence with JSON serialization to platform-specific directories. The code demonstrates solid fundamentals with 72% test coverage (exceeding the 65% target), comprehensive error handling, and no race conditions. The package defines two overlapping state schemas (`GameState` and `SaveState`) that should be unified for consistency.

## Issues Found
- [ ] low API Design — Two competing state schemas (`GameState` in `save.go:25` and `SaveState` in `schema.go:6`) exist without documentation explaining their relationship or intended use cases
- [ ] low Documentation — Missing `doc.go` file for package-level documentation (package root)
- [ ] low Error Handling — `ListSlots()` silently continues on load errors without logging, potentially hiding corrupted save files (`save.go:210-212`)
- [ ] med Dependencies — Duplicate field definitions between `GameState.Player` and `SaveState.PlayerPosition`/`CameraDirection` suggest incomplete schema migration (`save.go:39-48`, `schema.go:11-30`)
- [ ] low API Design — `getSavePath()` and `getSlotPath()` are unexported but could be useful for debugging/testing extensions (consider exposing via test helpers) (`save.go:87`, `save.go:121`)
- [ ] low Error Handling — Version field hardcoded to "1.0" in `Save()` without validation during `Load()`, risking forward-compatibility issues (`save.go:146`)

## Test Coverage
72.0% (target: 65%) ✓

## Dependencies
**External**: None (stdlib only: `encoding/json`, `errors`, `fmt`, `os`, `path/filepath`, `runtime`, `time`)

**Integration Points**:
- Imported by `main.go:46` (primary consumer)
- Platform-specific save directories: Windows (`%APPDATA%\violence\saves`), Unix/macOS (`~/.violence/saves`)

## Recommendations
1. **Unify schemas** — Merge `GameState` and `SaveState` into a single canonical schema, or document the intended purpose of each (e.g., one for wire format, one for internal use)
2. **Add version validation** — Implement semver checking in `Load()` to reject incompatible save files gracefully
3. **Add doc.go** — Create package documentation explaining the save system architecture and schema choice
4. **Add error logging** — Log corrupted/unreadable saves in `ListSlots()` to aid debugging player-reported issues
5. **Consider backup** — Implement atomic writes with temp files + rename to prevent corruption on crashes during save

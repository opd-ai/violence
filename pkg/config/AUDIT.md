# Audit: github.com/opd-ai/violence/pkg/config
**Date**: 2026-03-01
**Status**: Complete

## Summary
Configuration management package handling TOML file loading, hot-reload via fsnotify, and thread-safe global config access. Overall health is good with strong test coverage (76.5%). Critical risk: global mutable state with potential race conditions if accessed without package-provided accessors.

## Issues Found
- [ ] med API Design — Global mutable variable `C` exported; clients could bypass thread-safe accessors (`config.go:34`)
- [ ] med Error Handling — `Save()` does not set `FederationHubURL` before writing, causing data loss on save/reload cycle (`config.go:86-107`)
- [ ] low Documentation — Missing godoc for package-level variables `mu`, `watcherMu`, `watcherActive`, etc. (`config.go:36-45`)
- [ ] low API Design — `Watch()` returns `(stop func(), err error)` but `err` is always `nil`; should return `func()` only (`config.go:114`)
- [ ] low Test Coverage — Missing test for `Save()` error handling when viper.WriteConfig() fails (`config_test.go`)
- [ ] low Concurrency Safety — `Watch()` callback could be invoked after `stop()` is called due to timing window in fsnotify handler (`config.go:128-155`)

## Test Coverage
76.5% (target: 65%) ✓

## Dependencies
**External:**
- `github.com/fsnotify/fsnotify` — File system event notifications for hot-reload
- `github.com/spf13/viper` — Configuration file parsing and management

**Internal:**
- None (leaf package)

**Integration Points:**
- Used by 2 packages in codebase (low coupling)
- Global state `C` accessible to all importers

## Recommendations
1. **HIGH**: Fix `Save()` to include `FederationHubURL` field (line 106) to prevent data loss
2. **MED**: Unexport `C` and force all access through `Get()`/`Set()` to prevent race conditions
3. **MED**: Fix `Watch()` callback timing issue by checking context before invoking callback
4. **LOW**: Simplify `Watch()` signature to return `func()` instead of `(func(), error)` since error is never returned
5. **LOW**: Add godoc comments for package-level synchronization primitives

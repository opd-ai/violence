# Audit: github.com/opd-ai/violence/pkg/ui
**Date**: 2026-03-01
**Status**: Needs Work

## Summary
UI package provides comprehensive HUD rendering, menu systems, chat overlays, and multiplayer UI components. Overall architecture is solid with good separation of concerns, but exhibits race condition risks in chat overlay, lacks complete test coverage (52.3% vs 65% target), and has no package-level documentation.

## Issues Found
- [ ] high concurrency — Race condition: `ChatOverlay.Messages` accessed without lock in `GetVisibleMessages()`, `Draw()`, and all scroll methods (`chat.go:152-172,175-249,133-149`)
- [ ] high concurrency — Race condition: `ChatOverlay.Visible`, `InputBuffer`, `CursorPosition`, `ScrollOffset` accessed without lock across multiple methods (`chat.go:57-80,106-131`)
- [ ] med documentation — Missing `doc.go` package documentation file (root of `pkg/ui/`)
- [ ] med testing — Test coverage at 52.3%, below 65% target (missing tests for shop, crafting, skills, mods, multiplayer UIs)
- [ ] med api-design — Global mutable state `currentTheme` accessed without synchronization (`ui.go:101,542`)
- [ ] low error-handling — `ApplySettingChange` and `ApplyKeyBinding` return errors but caller responsibility unclear (`ui.go:778,860`)
- [ ] low documentation — `getLoadingDots()` uses `ebiten.ActualTPS()` incorrectly - should use frame counter for animation cycle (`ui.go:965-979`)
- [ ] low api-design — `ChatOverlay` fields `Visible`, `Messages`, `InputBuffer` are exported but should be accessed via methods for encapsulation (`chat.go:28-39`)
- [ ] low api-design — `NameplatePlayer`, `ScoreboardEntry`, `ShopItem` etc. use exported fields instead of getters - violates encapsulation (`nameplate.go:14-22,deathmatch.go:105-113,ui.go:982-987`)

## Test Coverage
52.3% (target: 65%)

## Dependencies
External:
- `github.com/hajimehoshi/ebiten/v2` - Game engine (rendering, input)
- `golang.org/x/image/font/basicfont` - Font rendering

Internal:
- `pkg/config` - Configuration access
- `pkg/input` - Input action mappings

## Recommendations
1. **Fix race conditions in `ChatOverlay`**: Protect all shared state (`Messages`, `Visible`, `InputBuffer`, `ScrollOffset`) with mutex in all methods, not just `AddMessage()`
2. **Add package documentation**: Create `doc.go` describing UI architecture and component hierarchy
3. **Increase test coverage to 65%+**: Add tests for shop, crafting, skills, mods, and multiplayer UI rendering logic
4. **Thread-safe theme management**: Wrap `currentTheme` access with `sync.RWMutex` or make it immutable with atomic pointer swap
5. **Improve encapsulation**: Make struct fields private and expose via methods where mutation needs control

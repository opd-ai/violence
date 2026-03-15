# AUDIT — 2026-03-15

## Project Goals

VIOLENCE is a raycasting first-person shooter built with Go and Ebitengine. The project makes the following verifiable claims:

1. **Raycasting FPS engine** — DDA-based first-person renderer with FOV, pitch, and head-bob.
2. **100% procedural asset generation** — all audio, visuals, and narrative generated at runtime; no pre-authored asset files (`.mp3`, `.wav`, `.png`, etc.) are permitted.
3. **Deterministic RNG** — identical seeds produce identical outputs across all platforms.
4. **BSP level generation** — procedurally generated maps including arenas for deathmatch.
5. **E2E encrypted in-game chat** — ECDH P-256 key exchange with AES-256-GCM encryption.
6. **Multiplayer netcode** — co-op, deathmatch, and territory-control modes with lag compensation.
7. **Federation hub** — cross-server matchmaking and discovery via HTTP API.
8. **DHT decentralized discovery** — libp2p Kademlia DHT; claimed `<30s` bootstrap, `<5s` lookup.
9. **WASM mod sandboxing** — Wasmer runtime with capability-based security.
10. **Deterministic replay system** — binary format recording and playback.
11. **Leaderboards** — SQLite persistence with federated aggregation.
12. **Achievements** — 14 built-in achievements across four categories.
13. **Procedural dialogue, lore, and quests** — fully generated narrative content.
14. **82%+ test coverage threshold** — enforced in CI across Linux, macOS, Windows, WASM, iOS, Android.
15. **Dedicated server** — standalone binary (`cmd/server`) with configurable port and log level.
16. **Genre system** — registry with `SetGenre` interface; five built-in genres.
17. **Skill/talent trees, economy, faction reputation, territory control** — live simulation systems.

---

## Goal-Achievement Summary

| Goal | Status | Evidence |
|------|--------|----------|
| Raycasting FPS engine | ✅ Achieved | `pkg/raycaster/` tests pass; DDA algorithm verified |
| 100% procedural assets | ✅ Achieved | No `.mp3/.wav/.png/.jpg` files found; no `//go:embed` directives in production code |
| Deterministic RNG | ✅ Achieved | `pkg/rng/` tests pass with `-race` |
| BSP level generation | ✅ Achieved | `pkg/bsp/` tests pass with `-race` |
| E2E encrypted chat | ⚠️ Partial | Implementation exists; `TestPerformKeyExchange` deadlocks — confirmed hang at 30s timeout |
| Multiplayer netcode | ⚠️ Partial | `pkg/network/` built successfully; tests fail on headless runner (X11 missing) |
| Federation hub HTTP API | ✅ Achieved | `cmd/federation-hub/` builds successfully; HTTP handler implemented |
| DHT decentralized discovery | ⚠️ Partial | `pkg/federation/dht/` builds; tests timeout at 30s (UPnP/libp2p blocks) |
| WASM mod sandboxing | ✅ Achieved | `pkg/mod/` tests pass; Wasmer integration verified |
| Replay system | ✅ Achieved | `pkg/replay/` tests pass with `-race` |
| Leaderboards | ✅ Achieved | `pkg/leaderboard/` tests pass with `-race` |
| Achievements | ✅ Achieved | `pkg/achievements/` tests pass with `-race` |
| Procedural dialogue/lore/quests | ✅ Achieved | `pkg/dialogue/`, `pkg/lore/`, `pkg/quest/` all pass with `-race` |
| 82%+ test coverage in CI | ⚠️ Partial | 22 packages fail on headless runner (X11 panic); audio tests timeout; config has a race failure |
| Dedicated server binary | ✅ Achieved | `cmd/server` builds cleanly |
| Genre system | ✅ Achieved | `pkg/procgen/genre/` tests pass |
| Skills/economy/faction/territory | ⚠️ Partial | `pkg/skills`, `pkg/economy`, `pkg/faction` pass; `pkg/territory` and `pkg/squad` fail on headless |
| Procedural audio | ⚠️ Partial | `pkg/audio/` builds; all audio tests timeout at 30s (generation blocks) |

**Overall: 12/17 goals fully achieved (71%)**

---

## Findings

### CRITICAL

- [ ] **Chat key-exchange deadlock** — `pkg/chat/keyexchange.go:45` — `PerformKeyExchange` sends the local public key before reading the peer's key. Over an unbuffered synchronous connection (`net.Pipe()`), both goroutines block permanently at the first `conn.Write()` because neither reads until it finishes writing — a textbook send-send deadlock. `TestPerformKeyExchange` times out every run (`FAIL pkg/chat 30.049s`). On real TCP, kernel send-buffers mask the problem only while they have room; under load the same deadlock can manifest. E2E encrypted chat is documented as feature-complete.
  - **Remediation:** In `keyexchange.go`, launch the send in a goroutine before blocking on receive:
    ```go
    sendErr := make(chan error, 1)
    go func() { sendErr <- sendPublicKey(conn, ourPublicKey) }()
    peerPublicKeyBytes, err := receivePublicKey(conn)
    if err != nil { return nil, err }
    if err := <-sendErr; err != nil { return nil, err }
    ```
    This pipelines the send and receive, eliminating the deadlock regardless of connection type.
  - **Validation:** `go test -race -timeout 30s ./pkg/chat/...` passes.

- [ ] **Config watch data race** — `pkg/config/watch.go:76` — `go test -race ./pkg/config/...` reports a confirmed DATA RACE in `TestWatch_MultipleWatchers`: `reloadConfiguration()` (called from the viper file-watcher goroutine) writes to the global `C Config` via `viper.Unmarshal()` while the test concurrently calls `viper.Reset()`, which also mutates viper's internal state. The test does not stop the watcher before resetting viper, so the watcher goroutine and test goroutine race on viper's global singleton. In production this manifests as a race between hot-reload and any reader calling `viper.Unmarshal` concurrently.
  - **Remediation:** In `TestWatch_MultipleWatchers` (`pkg/config/config_test.go:627`), call the stop function returned by `Watch()` and add a short settle delay before `viper.Reset()`. Additionally, gate `reloadConfiguration()` behind a read of `watcherCtx.Done()` so that a cancelled watcher never calls `viper.Unmarshal` after `Reset()`.
  - **Validation:** `go test -race ./pkg/config/...` exits with code 0 and no race warnings.

---

### HIGH

- [ ] **Audio generation blocks indefinitely in tests** — `pkg/audio/` — All audio tests (`pkg/audio`) time out at 30s. The goroutine dump shows a goroutine blocked inside `generateLoop()` on another thread (stack unavailable), indicating the synchronous audio generation loop never terminates in the test context. Procedural audio is a primary differentiator and a core stated goal, yet zero audio test coverage is currently measured.
  - **Remediation:** Refactor `generateLoop()` to accept a `context.Context` for cancellation. In tests, pass a context with a 5s deadline: `ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second); defer cancel()`. This allows generation to be interrupted without hanging the test runner.
  - **Validation:** `go test -race -timeout 60s ./pkg/audio/...` passes.

- [ ] **Federation/DHT tests always timeout** — `pkg/federation/`, `pkg/federation/dht/` — Both packages time out at 30s. The goroutine dump reveals `goupnp/httpu` UDP reads and `go-libp2p-kad-dht` peer-loop goroutines that never terminate. The README claims `<30s bootstrap, <5s lookup` but these properties cannot be verified in test. The DHT node created in tests attempts real UPnP/SSDP multicast discovery which blocks indefinitely in environments without a LAN gateway.
  - **Remediation:** Add a test build tag `//go:build integration` to tests that start a live DHT node. Provide unit tests that mock the DHT interface. For the live tests, ensure `host.Close()` and `dht.Close()` are called in `defer` before the test timeout elapses, and pass a short `context.WithTimeout` to all bootstrap calls.
  - **Validation:** `go test -short -timeout 30s ./pkg/federation/...` passes; integration tests gated behind `-tags integration`.

- [ ] **22 packages untestable on headless runner** — multiple packages — The following packages all panic at init time with `glfw: X11: The DISPLAY environment variable is missing` because they import Ebitengine (which initialises GLFW in `init()`): `pkg/ai`, `pkg/combat`, `pkg/weapon`, `pkg/sprite`, `pkg/network`, `pkg/lighting`, `pkg/engine`, `pkg/render`, `pkg/ui`, `pkg/input`, `pkg/automap`, `pkg/territory`, `pkg/squad`, `pkg/collision`, `pkg/animation`, `pkg/audio`, `pkg/attacktrail`, `pkg/attackanim`, `pkg/camerafx`, `pkg/damagenumber`, `pkg/statusfx`, `pkg/playersprite`. These are the core simulation packages. The stated "82%+ coverage enforced in CI" is structurally unachievable in any headless CI runner without X11/Xvfb. No `//go:build !headless` guards or mock Ebitengine interfaces exist.
  - **Remediation:** (a) Add `Xvfb` to CI runner setup for the Linux headless job, or (b) introduce a `//go:build ebitengine` tag on files that directly reference `ebiten.Image` / `ebiten.DrawImageOptions`, so that logic-only tests can run without the graphics runtime. Option (a) is lower effort; option (b) is architecturally cleaner.
  - **Validation:** `go test -race -cover ./...` exits with code 0 on the CI Linux runner; `go tool cover` reports ≥82% total.

---

### MEDIUM

- [ ] **`PerformKeyExchange` lacks connection deadline / context support** — `pkg/chat/keyexchange.go:29` — Even after fixing the deadlock, `PerformKeyExchange` has no timeout parameter. A slow or malicious peer can hold a connection open indefinitely during the handshake phase. The existing `TestKeyExchangeTimeout` test works by setting a `net.Conn` deadline externally, but callers in production code do not set deadlines.
  - **Remediation:** Change the signature to `PerformKeyExchange(ctx context.Context, conn net.Conn) ([]byte, error)` and apply `conn.SetDeadline(time.Now().Add(10 * time.Second))` inside using the context deadline if set.
  - **Validation:** `go test -race -timeout 30s ./pkg/chat/...` passes; no blocking in `TestPerformKeyExchange`.

- [ ] **Profanity filter l33t-speak detection incomplete** — `pkg/chat/filter.go` — The wordlist generator produces basic single-character substitutions (a→4, e→3, etc.) but does not handle: double substitutions (`@@` for `aa`), Unicode homoglyphs (Cyrillic `а` for Latin `a`), or phonetic evasions. Acknowledged in the existing GAPS.md but not resolved. Chat content moderation is a stated safety feature.
  - **Remediation:** Extend `generateLeetSpeakVariants()` to include a Unicode normalisation step (`golang.org/x/text/unicode/norm`) before matching, and add a homoglyph-replacement map covering the most common Cyrillic/Greek lookalikes. Add test cases covering `sh1t`, `f@ck`, `а$$` (Cyrillic a), etc.
  - **Validation:** `go test -race ./pkg/chat/...` includes a `TestProfanityFilterVariants` case that catches the above examples.

- [ ] **54 dead/unreferenced exported functions** — various packages — `go-stats-generator` reports 54 unreferenced functions across the codebase. Dead exported functions cannot be verified for correctness, inflate the public API surface, and may mislead future contributors into depending on untested code paths.
  - **Remediation:** Run `go-stats-generator analyze . --format json | jq '.maintenance.dead_code'` to enumerate the list. For each: unexport it if not intended as a public API, or add a test that exercises the function.
  - **Validation:** `go-stats-generator analyze . --skip-tests | grep "Dead Code"` reports 0 unreferenced functions.

- [ ] **Deprecated public APIs still exported** — `pkg/mod/mod.go:316,333`, `pkg/status/status.go:130,184`, `pkg/ui/mod_browser.go:419,424` — Six functions/methods are marked `DEPRECATED` in their doc comments but remain exported. This bloats the public API and may be mistakenly used by mod authors.
  - **Remediation:** Add a `//go:deprecated` annotation (Go 1.21+) to each deprecated identifier so that `go vet` and IDEs surface the warning at call sites. Provide a migration guide in a `MIGRATION.md` or inline doc comment pointing to the replacement.
  - **Validation:** `go vet ./...` emits deprecation hints for callers of these functions.

- [ ] **High-complexity rendering functions** — `main.go:renderCombatEffects` (cyclomatic 21.8), `pkg/healthbar/system.go:RenderHealthBarsWithLayout` (18.4), `pkg/floor/weathering.go:applyEdgeDamage` (18.1) — Functions with cyclomatic complexity >15 are statistically correlated with latent bugs and are difficult to unit test.
  - **Remediation:** Decompose `renderCombatEffects` into `renderDamageNumbers()`, `renderHitMarkers()`, and `renderBloodEffects()`. Extract the layout calculation in `RenderHealthBarsWithLayout` into a `LayoutCalculator` helper type. Each extracted function should not exceed complexity 8.
  - **Validation:** `go-stats-generator analyze . --skip-tests | grep "High Complexity"` reports 0 functions with cyclomatic > 15.

---

### LOW

- [ ] **`main.go` monolith** — `main.go` — 6,481 lines, 262 functions, `NewGame()` is 349 lines, 90 package-level imports. This is architecturally expected for a game entry point but makes targeted testing and future refactoring expensive.
  - **Remediation:** Extract ECS system registration into `pkg/game/systems.go` and the `NewGame` initialisation sequence into `pkg/game/init.go`. Aim for the `main` package to have ≤50 direct package-level imports.
  - **Validation:** `go-stats-generator analyze . --skip-tests | grep "main"` shows average function length < 30 lines.

- [ ] **19,705 magic numbers** — codebase-wide — `go-stats-generator` reports 19,705 integer and float literals used directly in logic without named constants. Magic numbers impede readability and make tuning/balancing require grep hunts.
  - **Remediation:** Define named constants in domain packages (e.g., `pkg/combat/constants.go`) for tuning parameters like damage multipliers, speed limits, and FOV values. Focus first on constants that appear in multiple files.
  - **Validation:** After one pass, `go-stats-generator` should report a meaningful reduction in magic numbers in the targeted packages.

- [ ] **2.12% code duplication — 81 clone pairs** — `pkg/network/ffa.go:203-208` ≡ `pkg/network/team.go:249-254`; `pkg/fog/system.go:51-56` ≡ `pkg/fog/system.go:96-101`; `pkg/floor/texture.go:248-253` ≡ `pkg/parallax/generator.go:349-355` — Exact and renamed clone pairs indicate missed helper-function extractions. The network duplication is especially notable given that `ffa.go` and `team.go` share the same update logic.
  - **Remediation:** Extract the duplicated network logic in `ffa.go`/`team.go` into a shared `pkg/network/common.go` helper. Deduplicate the fog update pattern into a `pkg/fog/updateCells()` helper.
  - **Validation:** `go-stats-generator analyze . --skip-tests | grep "Duplication Ratio"` shows < 1.5%.

- [ ] **67 file naming violations, 116 identifier violations** — various — `go-stats-generator` identifies naming convention violations (e.g., `AIAdaptation` should be `aiAdaptation` per package-scope convention; `StateIdle` uses `acronym_conflict` pattern). One package name violation also reported.
  - **Remediation:** Correct the most impactful identifier names in public APIs (e.g., rename exported `AIAdaptation` to follow Go export conventions). File name violations can be corrected opportunistically during other work.
  - **Validation:** `go-stats-generator analyze . --skip-tests | grep "Naming Score"` improves from 0.99.

---

## Metrics Snapshot

| Metric | Value |
|--------|-------|
| Total Lines of Code | 44,459 |
| Total Files | 298 |
| Total Packages | 94 |
| Total Functions | 1,167 |
| Total Methods | 2,512 |
| Total Structs | 619 |
| Total Interfaces | 25 |
| Average Function Length | 14.5 lines |
| Longest Function (`NewGame`) | 349 lines |
| Functions > 50 lines | 125 (3.4%) |
| Functions > 100 lines | 14 (0.4%) |
| Average Cyclomatic Complexity | 3.8 |
| High Complexity Functions (>10) | 9 |
| Max Complexity (`renderCombatEffects`) | 21.8 |
| Documentation Coverage (overall) | 85.3% |
| Function Doc Coverage | 93.5% |
| Method Doc Coverage | 83.0% |
| Duplication Ratio | 2.12% (81 clone pairs, 1,728 lines) |
| Largest Clone | 59 lines |
| Dead/Unreferenced Functions | 54 |
| Magic Numbers | 19,705 |
| Complex Signatures (>5 params) | 300 |
| Deeply Nested Functions | 25 |
| Package Naming Violations | 1 |
| File Naming Violations | 67 |
| Identifier Violations | 116 |
| `go vet ./...` warnings | 0 |
| Confirmed Data Races (`-race`) | 1 (`pkg/config`) |
| Test Timeouts | 3 (`pkg/chat`, `pkg/audio`, `pkg/federation*`) |

### High-Risk Functions (Complexity > 15)

| Function | File | Lines | Complexity |
|----------|------|-------|------------|
| `renderCombatEffects` | `main.go` | 46 | 21.8 |
| `RenderHealthBarsWithLayout` | `pkg/healthbar/system.go` | 107 | 18.4 |
| `generateFlyingEnemy` | `pkg/sprite/sprite.go` | 86 | 18.1 |
| `applyEdgeDamage` | `pkg/floor/weathering.go` | 63 | 18.1 |
| `applyWearPatterns` | `pkg/floor/weathering.go` | 61 | 18.1 |
| `applyOrganicGrowth` | `pkg/floor/weathering.go` | 57 | 16.8 |
| `Update` | `pkg/ui/interactive.go` | 45 | 16.8 |
| `Reserve` | `pkg/ui/layout.go` | 76 | 16.6 |
| `DetermineTierFromObjective` | `pkg/loot/quest_rewards.go` | 36 | 16.3 |
| `renderProps` | `main.go` | 100 | 15.0 |

### Test Health by Package Category

| Category | Packages Passing | Packages Failing | Failure Reason |
|----------|-----------------|-----------------|----------------|
| Core procedural generation | 7/7 | 0/7 | — |
| Persistence / RPG systems | 8/8 | 0/8 | — |
| AI / rendering / simulation | 0/8 | 8/8 | X11 missing |
| Networking / federation | 0/3 | 3/3 | X11 / timeout |
| Chat encryption | 0/1 | 1/1 | Deadlock |
| Audio | 0/1 | 1/1 | Generation timeout |
| Config | 0/1 | 1/1 | Data race |

---

## Verification Commands

```bash
# Packages that pass today (no X11 required):
go test -race -timeout 60s \
  ./pkg/rng/... ./pkg/bsp/... ./pkg/raycaster/... ./pkg/texture/... \
  ./pkg/replay/... ./pkg/leaderboard/... ./pkg/achievements/... \
  ./pkg/quest/... ./pkg/inventory/... ./pkg/save/... \
  ./pkg/dialogue/... ./pkg/lore/... ./pkg/procgen/... \
  ./pkg/skills/... ./pkg/progression/... ./pkg/stats/... \
  ./pkg/crafting/... ./pkg/shop/... ./pkg/economy/... \
  ./pkg/faction/... ./pkg/minigame/... ./pkg/hazard/... \
  ./pkg/trap/... ./pkg/destruct/... ./pkg/secret/... \
  ./pkg/door/... ./pkg/level/... ./pkg/projectile/... \
  ./pkg/mod/... ./pkg/tutorial/...

# Confirm vet still clean:
go vet ./...

# Generate fresh metrics:
go-stats-generator analyze . --skip-tests

# Check race in config (currently fails):
go test -race -timeout 30s ./pkg/config/...

# Check chat deadlock (currently times out):
go test -race -timeout 30s ./pkg/chat/...
```

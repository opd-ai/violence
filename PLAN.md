# Implementation Plan: Achieving the 82% CI Coverage Gate and Eliminating Critical Blockers

## Project Context

- **What it does**: VIOLENCE is a raycasting first-person shooter built with Go and Ebitengine, featuring 100% procedurally generated assets (audio, visuals, narrative), deterministic RNG, multiplayer via libp2p DHT federation, E2E encrypted chat, WASM mod sandboxing, and a deterministic replay system.
- **Current goal**: Advance from 71% (12/17) to 100% goal achievement — beginning with the four items that block CI from ever reaching its stated 82% coverage threshold.
- **Estimated Scope**: Large (>15 items above threshold when counting all critical + high + medium findings)

---

## Goal-Achievement Status

| Stated Goal | Current Status | This Plan Addresses |
|---|---|---|
| Raycasting FPS engine | ✅ Achieved | No |
| 100% procedural asset generation | ✅ Achieved | No |
| Deterministic RNG | ✅ Achieved | No |
| BSP level generation | ✅ Achieved | No |
| E2E encrypted in-game chat | ⚠️ Partial — key-exchange deadlock | **Yes (Step 1)** |
| Multiplayer netcode | ⚠️ Partial — X11 headless panic | **Yes (Step 3)** |
| Federation hub HTTP API | ✅ Achieved | No |
| DHT decentralized discovery (<30 s bootstrap, <5 s lookup) | ⚠️ Partial — tests always timeout | **Yes (Step 4)** |
| WASM mod sandboxing | ✅ Achieved | No |
| Deterministic replay system | ✅ Achieved | No |
| Leaderboards | ✅ Achieved | No |
| Achievements | ✅ Achieved | No |
| Procedural dialogue / lore / quests | ✅ Achieved | No |
| 82%+ test coverage enforced in CI | ⚠️ Partial — structurally unreachable | **Yes (Steps 2, 3, 4, 5)** |
| Dedicated server binary | ✅ Achieved | No |
| Genre system | ✅ Achieved | No |
| Skills / economy / faction / territory simulation | ⚠️ Partial — territory & squad X11 panic | **Yes (Step 3)** |
| Procedural audio (core differentiator) | ⚠️ Partial — tests block indefinitely | **Yes (Step 2)** |

---

## Metrics Summary (go-stats-generator baseline — 2026-03-15)

| Metric | Value | Threshold | Status |
|---|---|---|---|
| Total LoC | 44,459 | — | — |
| Total packages | 94 | — | — |
| Total functions + methods | 3,679 | — | — |
| Functions with cyclomatic complexity > 9.0 | **17** | <5 = Small | **Large** |
| Duplication ratio | **2.12%** (81 clone pairs, 1,728 lines) | <3% = Small | Small |
| Doc coverage (overall) | **85.3%** | >75% = OK | OK |
| Doc coverage — methods | 83.0% | — | OK |
| Doc coverage — types | 83.6% | — | OK |
| Packages with confirmed test failures (headless) | **22** | 0 | ❌ |
| Packages with race detector failures | **1** (`pkg/config`) | 0 | ❌ |
| Packages whose tests time out | **3** (`pkg/audio`, `pkg/federation`, `pkg/federation/dht`) | 0 | ❌ |
| Deprecated exports without Go 1.21 annotation | **6** | 0 | ⚠️ |
| Clone pairs | 81 | <30 | ⚠️ |
| HACK/BUG inline comments needing follow-up | 7 | 0 | ⚠️ |

**Complexity hotspots on goal-critical paths:**

| Function | Package | File | Cyclomatic | Goal Impact |
|---|---|---|---|---|
| `renderCombatEffects` | `main` | `main.go:4565` | 16 | Rendering quality |
| `RenderHealthBarsWithLayout` | `healthbar` | `pkg/healthbar/system.go:517` | 13 | Rendering quality |
| `applyEdgeDamage` | `floor` | `pkg/floor/weathering.go:105` | 12 | Procedural generation |
| `applyWearPatterns` | `floor` | `pkg/floor/weathering.go:172` | 12 | Procedural generation |
| `Reserve` | `ui` | `pkg/ui/layout.go:94` | 12 | Headless-testable UI layer |
| `generateFlyingEnemy` | `sprite` | `pkg/sprite/sprite.go:1359` | 12 | Procedural assets |
| `applyOrganicGrowth` | `floor` | `pkg/floor/weathering.go:285` | 11 | Procedural generation |
| `DetermineTierFromObjective` | `loot` | `pkg/loot/quest_rewards.go:306` | 11 | Economy / progression |
| `Update` | `ui` | `pkg/ui/interactive.go:101` | 11 | UI coverage |

---

## Implementation Steps

Steps are ordered: **prerequisites first, then by descending impact on stated goals**.
Steps 1–5 address the critical and high-priority blockers that make the 82% gate structurally unachievable; Steps 6–9 address medium-priority items.

---

### Step 1: Fix Chat Key-Exchange Deadlock *(Critical — E2E Chat Goal)*

- **Deliverable**: `pkg/chat/keyexchange.go` — move `sendPublicKey` into a goroutine launched before the blocking `receivePublicKey` call, so both directions proceed concurrently.

  ```go
  // In PerformKeyExchange, replace sequential calls with:
  sendErr := make(chan error, 1)
  go func() { sendErr <- sendPublicKey(conn, ourPublicKey) }()
  peerPublicKeyBytes, err := receivePublicKey(conn)
  if err != nil { return nil, err }
  if err := <-sendErr; err != nil { return nil, err }
  ```

- **Dependencies**: None — this is the smallest independent fix.
- **Goal Impact**: Closes "E2E encrypted in-game chat" — the handshake is the only broken link; `EncryptMessage`/`DecryptMessage` already work correctly.
- **Acceptance**: `TestPerformKeyExchange` passes within its 30-second timeout; no goroutine leak.
- **Validation**:
  ```bash
  go test -race -timeout 30s ./pkg/chat/...
  ```

---

### Step 2: Make Procedural Audio Cancellable *(High — Audio Goal, CI Gate)*

- **Deliverable**: `pkg/audio/` — refactor `generateLoop()` (and every other long-running generation function in the package) to accept a `context.Context`. Check `ctx.Done()` between generation steps. Update all callers to pass a context. Add unit tests with a 5-second deadline context.

  Specifically:
  - Add `ctx context.Context` as the first parameter to `generateLoop` in `pkg/audio/audio.go`.
  - Add `select { case <-ctx.Done(): return; default: }` inside the generation loop's innermost hot loop.
  - Update tests to pass `context.WithTimeout(context.Background(), 5*time.Second)`.

- **Dependencies**: Step 1 (establishes passing test baseline).
- **Goal Impact**: Closes "procedural audio" gap; contributes measurable coverage for a core stated differentiator. Prevents synchronous stall on slow hardware before the first game frame.
- **Acceptance**: `go test -race -timeout 60s ./pkg/audio/...` passes (no timeout). Coverage for `pkg/audio` > 0%.
- **Validation**:
  ```bash
  go test -race -timeout 60s -cover ./pkg/audio/...
  ```

---

### Step 3: Unblock 22 Headless Packages in CI *(High — 82% Coverage Gate, Core Gameplay)*

This is the highest-leverage single step for the coverage goal — it unblocks `pkg/ai`, `pkg/combat`, `pkg/weapon`, `pkg/sprite`, `pkg/network`, `pkg/lighting`, `pkg/engine`, `pkg/render`, `pkg/ui`, `pkg/input`, `pkg/automap`, `pkg/territory`, `pkg/squad`, `pkg/collision`, `pkg/animation`, `pkg/attacktrail`, `pkg/attackanim`, `pkg/camerafx`, `pkg/damagenumber`, `pkg/statusfx`, `pkg/playersprite`, and audio simultaneously.

- **Deliverable** (low-effort path — recommended first): Update `.github/workflows/ci.yml` — add an Xvfb setup step before the `go test` invocation on the Linux runner:

  ```yaml
  - name: Install Xvfb
    run: sudo apt-get install -y xvfb

  - name: Run tests (Linux with virtual display)
    run: |
      Xvfb :99 -screen 0 1024x768x24 &
      export DISPLAY=:99
      go test -race -cover ./...
  ```

  For macOS runners, use the existing headless display support (no change needed — `ebiten` on macOS uses Metal, which does not require X11).

- **Dependencies**: Step 2 (audio no longer blocks); otherwise independent.
- **Goal Impact**: Enables CI to measure coverage across all 94 packages — prerequisite for the 82% gate. Directly unblocks territory and squad, completing the "Skills/economy/faction/territory" goal.
- **Acceptance**: `go test -race -cover ./...` on the CI Linux runner exits with code 0; `go tool cover` reports ≥82% total.
- **Validation**:
  ```bash
  # On CI runner with DISPLAY=:99 set:
  go test -race -cover ./... 2>&1 | grep -E "^(ok|FAIL|---)"
  go tool cover -func=coverage.out | tail -1
  ```

---

### Step 4: Fix Config Watch Data Race *(High — Reliability, CI Gate)*

- **Deliverable**: Two targeted changes:

  1. `pkg/config/watch.go` (near line 63): In `reloadConfiguration()`, add a guard before `viper.Unmarshal`:
     ```go
     if ctx.Err() != nil {
         return
     }
     ```
  2. `pkg/config/config_test.go` (near line 627, `TestWatch_MultipleWatchers`): Call the stop function returned by `Watch()` and add a settle delay before `viper.Reset()`:
     ```go
     stopFn()
     time.Sleep(50 * time.Millisecond)
     viper.Reset()
     ```

- **Dependencies**: None — independent of Steps 1–3.
- **Goal Impact**: Closes "Config Watch data race" — required for `-race` to pass on `pkg/config`, which is a CI gate dependency.
- **Acceptance**: `go test -race ./pkg/config/...` exits with code 0, no `WARNING: DATA RACE` output.
- **Validation**:
  ```bash
  go test -race -count=5 ./pkg/config/...
  ```

---

### Step 5: Gate DHT/Federation Tests with `integration` Build Tag *(High — DHT Goal)*

- **Deliverable**: `pkg/federation/` and `pkg/federation/dht/` — add `//go:build integration` to test files that start a live libp2p host or real UPnP/SSDP sockets. Provide lightweight unit tests (mock `DHT` interface) that exercise routing logic without live network I/O. Ensure all live-DHT tests call `host.Close()` and `dht.Close()` in `t.Cleanup`.

  Concretely:
  - Create `pkg/federation/dht/dht_interface.go` defining a `DHT` interface with `Bootstrap`, `FindPeer`, and `Close` methods.
  - Create `pkg/federation/dht/mock_dht_test.go` with a stub implementing the interface and basic unit tests.
  - Add `//go:build integration` to `pkg/federation/dht/dht_test.go` and `pkg/federation/federation_test.go`.
  - Update `.github/workflows/ci.yml` to add a separate `integration` job that runs `go test -tags integration ./pkg/federation/...` on a network-enabled runner.

- **Dependencies**: Step 3 (CI structure change makes adding a second job cheap).
- **Goal Impact**: Restores verifiability for the `<30 s bootstrap` and `<5 s lookup` claims. Prevents indefinite hangs from blocking the standard `go test ./...` run.
- **Acceptance**: `go test -short -timeout 30s ./pkg/federation/...` passes; integration tests visible as a separate CI job gated behind `-tags integration`.
- **Validation**:
  ```bash
  go test -short -timeout 30s ./pkg/federation/...
  go test -tags integration -timeout 120s ./pkg/federation/...  # on network-enabled runner
  ```

---

### Step 6: Add Go 1.21 Deprecation Annotations *(Medium — Mod API Safety)*

- **Deliverable**: Six deprecated exported identifiers updated across three files:
  - `pkg/mod/mod.go:316` — `PluginManager()`: add `// Deprecated: Use WASMLoader() instead. This function bypasses WASM sandboxing.`
  - `pkg/mod/mod.go:333` — `RegisterPlugin()`: add `// Deprecated: Use RegisterWASMModule() instead.`
  - `pkg/status/status.go:130,184` — two deprecated status methods: add `// Deprecated: Use <replacement> instead.`
  - `pkg/ui/mod_browser.go:419,424` — two deprecated navigation methods: add `// Deprecated: Use <replacement> instead.`

  Each deprecated comment must follow the Go doc convention: first line of the doc comment starts with `Deprecated:` so that `go vet`, `gopls`, and `staticcheck` surface it at call sites.

- **Dependencies**: None — purely additive documentation changes.
- **Goal Impact**: Protects the WASM mod sandboxing goal — prevents mod authors from accidentally using the unsafe `PluginManager` path. Addresses the mod API security concern noted in GAPS.md Gap 7.
- **Acceptance**: `go vet ./...` emits deprecation hints for any internal callers of these six identifiers.
- **Validation**:
  ```bash
  go vet ./...
  # Verify with gopls or staticcheck if available:
  staticcheck ./... 2>/dev/null | grep -i "deprecated" || true
  ```

---

### Step 7: Add Context + Deadline to `PerformKeyExchange` *(Medium — Chat Security)*

- **Deliverable**: `pkg/chat/keyexchange.go` — change signature to `PerformKeyExchange(ctx context.Context, conn net.Conn) ([]byte, error)`. Apply `conn.SetDeadline(time.Now().Add(10 * time.Second))` inside, overriding with the context deadline if shorter. Update all call sites.

- **Dependencies**: Step 1 (the deadlock fix must land first; this step hardens the already-fixed function).
- **Goal Impact**: Prevents a slow or malicious peer from holding a chat handshake open indefinitely. Addresses AUDIT.md §MEDIUM finding.
- **Acceptance**: `TestKeyExchangeTimeout` passes; `PerformKeyExchange` accepts a context; all existing chat tests pass.
- **Validation**:
  ```bash
  go test -race -timeout 30s ./pkg/chat/...
  go build ./...
  ```

---

### Step 8: Extend Profanity Filter for Unicode / Homoglyph Evasion *(Medium — Chat Safety)*

- **Deliverable**: `pkg/chat/filter.go` — three targeted changes:

  1. Apply `golang.org/x/text/unicode/norm.NFKD` normalisation to each candidate string before matching, collapsing Unicode homoglyphs.
  2. Extend `generateLeetSpeakVariants()` with a second-pass handling double-character patterns (`@@`, `**`, etc.).
  3. Add a homoglyph replacement map for the most common Cyrillic and Greek lookalikes (≥20 entries covering letters a, e, i, o, p, c, x, etc.).
  4. Add `TestProfanityFilterVariants` covering: `sh1t`, `f@ck`, `а$$` (Cyrillic а), `ρorn` (Greek ρ).

- **Dependencies**: Step 1 (requires chat tests to be passing as a baseline).
- **Goal Impact**: Closes GAPS.md Gap 6 — hardens the multiplayer content-safety posture for potential platform/storefront submission.
- **Acceptance**: `TestProfanityFilterVariants` passes; all four evasion examples are caught.
- **Validation**:
  ```bash
  go test -race -run TestProfanityFilter ./pkg/chat/...
  ```

---

### Step 9: Decompose High-Complexity Rendering Functions *(Medium — Testability, Maintainability)*

Targets the 17 functions above cyclomatic complexity 9.0. Prioritise the five that are on goal-critical paths and/or block coverage measurement.

- **Deliverable** (ordered by complexity / impact):

  1. `main.go:renderCombatEffects` (cyclomatic 16) → extract `renderDamageNumbers()`, `renderHitMarkers()`, `renderBloodEffects()` — each ≤8 cyclomatic. Keep in `main.go` or move to `pkg/render/effects.go`.
  2. `pkg/healthbar/system.go:RenderHealthBarsWithLayout` (13) → extract layout calculation into a `LayoutCalculator` helper type in the same file.
  3. `pkg/floor/weathering.go:applyEdgeDamage`, `applyWearPatterns`, `applyOrganicGrowth` (12, 12, 11) → extract shared sub-operations into named helpers within `weathering.go`.
  4. `pkg/ui/layout.go:Reserve` (12) → extract sub-cases into `reserveHorizontal()` / `reserveVertical()` helpers.

- **Dependencies**: Step 3 (Xvfb enables tests for `pkg/render` and `pkg/ui` to actually run and validate decomposition).
- **Goal Impact**: Reduces the functions-above-threshold count from 17 to ≤4; improves testability of the rendering pipeline; reduces statistical bug risk in the most complex paths.
- **Acceptance**: `go-stats-generator analyze . --skip-tests` reports ≤4 functions with cyclomatic complexity > 9.0.
- **Validation**:
  ```bash
  go-stats-generator analyze . --skip-tests --format json --sections functions | \
    python3 -c "
  import json,sys
  d=json.load(sys.stdin)
  hot=[f for f in d['functions'] if f['complexity']['cyclomatic']>9]
  print(f'{len(hot)} functions above threshold')
  for f in sorted(hot,key=lambda x:-x['complexity']['cyclomatic']):
      print(f'  {f[\"complexity\"][\"cyclomatic\"]}  {f[\"package\"]}.{f[\"name\"]}')
  "
  go build ./...
  go test -race ./...
  ```

---

## Dependency Graph

```
Step 4 (config race)   ──────────────────────────────────────┐
Step 1 (chat deadlock) ──┬──────────────────────────────────┐│
                         │                                   ││
Step 2 (audio ctx) ──────┤                                   ││
                         ▼                                   ││
Step 3 (Xvfb CI) ────────┬──► Step 9 (complexity)           ││
                         │                                   ││
Step 5 (DHT tags) ◄──────┘                                   ││
                                                             ││
Step 6 (deprecations)    ────────────────────────────────────┘│
                                                              │
Step 7 (chat ctx) ◄───────────────────────────────────────────┘
Step 8 (profanity) ◄─── Step 1
```

- Steps 1, 4, and 6 have no prerequisites — they can begin immediately and in parallel.
- Step 2 follows Step 1 (establishes a passing baseline before adding new tests).
- Step 3 follows Step 2 (audio must not time-out before Xvfb unblocks the full suite).
- Steps 5, 9 follow Step 3.
- Steps 7, 8 follow Step 1.

---

## Scope Calibration

| Category | Count | Classification |
|---|---|---|
| Functions above cyclomatic 9.0 | 17 | **Large** |
| Packages failing on headless | 22 | **Large** |
| Clone pairs | 81 (2.12% duplication ratio) | **Small** (below 3% threshold) |
| Doc coverage gap (exported symbols undocumented) | ~15% gap on methods/types | **Medium** |

**Overall plan scope: Large** — the dominant cost is the Xvfb/CI infra change (Step 3) and the audio refactor (Step 2). The chat fixes (Steps 1, 7, 8) are small in code size but high in strategic value.

---

## Deferred Items (out of scope for this plan)

The following are tracked in GAPS.md as intentionally deferred and are not addressed here:

- **Mod Marketplace** — requires player base first.
- **Mobile Store Publishing** — blocked on store policy review.
- **Cross-Save Sync** — blocked on backend infrastructure decision.
- **`main.go` monolith extraction** — low risk; plan separately as `pkg/game/` refactor after CI gate is green.
- **Magic number constants (19,705 literals)** — long-tail cleanup; address package-by-package during normal feature work.
- **Duplicate code elimination (81 pairs, 2.12%)** — below the 3% concern threshold; address opportunistically.

---

*Generated 2026-03-15 using `go-stats-generator v1.0.0` baseline. Re-run `go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation,packages,patterns` after each step to track regression against the thresholds above.*

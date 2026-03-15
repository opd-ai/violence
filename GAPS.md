# Implementation Gaps — 2026-03-15

> This document supersedes the previous GAPS.md (dated 2026-03-01).
> It is generated from a fresh functional audit. See `AUDIT.md` for full findings and metrics.

---

## Gap 1: Chat Key-Exchange Deadlock

- **Stated Goal**: "E2E encrypted in-game chat" using ECDH P-256 key exchange and AES-256-GCM encryption; documented as complete in v5.0.
- **Current State**: `pkg/chat/keyexchange.go:45` — `PerformKeyExchange` calls `sendPublicKey` before `receivePublicKey`. Both sides of a connection run the same function concurrently. Over any synchronous unbuffered connection (`net.Pipe()`, or a real TCP connection with exhausted send buffers), both goroutines block permanently at their first write because neither begins reading. `TestPerformKeyExchange` times out every run. The `EncryptMessage`/`DecryptMessage` primitives work correctly in isolation; the handshake is the broken link.
- **Impact**: E2E chat sessions cannot be established. Any game feature that depends on `PerformKeyExchange` (session initiation, chat room join) is non-functional on connections that don't have persistent OS-level send buffers large enough to absorb a 67-byte key message. On loaded servers, this is a production deadlock.
- **Closing the Gap**: In `pkg/chat/keyexchange.go`, move `sendPublicKey` into a goroutine launched before the blocking `receivePublicKey` call. Example fix in AUDIT.md §CRITICAL finding 1. Estimated effort: < 1 hour. Validation: `go test -race -timeout 30s ./pkg/chat/...` passes.

---

## Gap 2: Procedural Audio Untestable

- **Stated Goal**: "100% of gameplay assets are procedurally generated at runtime" — explicitly includes all audio (music, SFX, ambient). Claimed as a core differentiator.
- **Current State**: `pkg/audio/` builds successfully and is imported by the main game loop, but all test goroutines block indefinitely inside `generateLoop()`. Tests time out at 30s. Zero test coverage for the audio package is currently measurable. The procedural audio generation does not support cancellation.
- **Impact**: Correctness of procedural audio cannot be verified. Regressions in audio generation are undetectable by CI. The "deterministic: identical inputs produce identical outputs" guarantee is not tested for audio. On slow hardware or under load, the synchronous generation loop may cause the game to stall before the first frame renders.
- **Closing the Gap**: Refactor `generateLoop()` (and any other long-running audio generation functions) to accept a `context.Context` parameter and check `ctx.Done()` between generation steps. Add unit tests with a 5-second context deadline. Estimated effort: 2–4 hours.

---

## Gap 3: Federation and DHT Tests Unverifiable

- **Stated Goal**: "Federation hub enables cross-server matchmaking and discovery"; "DHT decentralised discovery with <30s bootstrap and <5s lookup".
- **Current State**: `pkg/federation/` and `pkg/federation/dht/` both time out at 30s. The `pkg/federation/dht` test starts a live libp2p host that attempts real UPnP/SSDP multicast over the local network; in a headless CI environment there is no LAN gateway and the UDP sockets block indefinitely. The `<30s bootstrap` and `<5s lookup` performance claims have no passing test to back them.
- **Impact**: DHT correctness and latency guarantees are unverified. A regression in bootstrap logic, peer routing, or lookup would not be caught by CI. The HTTP federation hub (`cmd/federation-hub`) builds and the handler code looks complete, but its integration with the DHT layer cannot be tested end-to-end in CI.
- **Closing the Gap**: (a) Gate tests that require a real network behind a `//go:build integration` tag and separate CI job. (b) Introduce a `DHT` interface in `pkg/federation/dht/` and provide a mock implementation for unit tests. (c) Ensure the live-DHT test properly closes the libp2p host and DHT node in a `t.Cleanup` function so goroutines don't leak across test runs.

---

## Gap 4: 22 Core Packages Cannot Be Tested Headlessly

- **Stated Goal**: "82%+ test coverage threshold enforced in CI" across Linux, macOS, Windows, WASM, iOS, and Android.
- **Current State**: 22 packages — including `pkg/ai`, `pkg/combat`, `pkg/weapon`, `pkg/sprite`, `pkg/network`, `pkg/lighting`, `pkg/engine`, `pkg/render`, `pkg/ui`, `pkg/input`, `pkg/automap`, `pkg/territory`, `pkg/squad`, `pkg/collision`, `pkg/animation`, `pkg/audio`, `pkg/attacktrail`, `pkg/attackanim`, `pkg/camerafx`, `pkg/damagenumber`, `pkg/statusfx`, `pkg/playersprite` — all panic at test init time with `glfw: X11: The DISPLAY environment variable is missing`. These are the most logic-dense packages in the project (combat, AI, rendering, input, networking). No `//go:build !headless` guards or mock Ebitengine interfaces exist. The 82% coverage claim is not achievable in any standard CI container without X11/Xvfb.
- **Impact**: Regressions in the combat model, AI behaviour trees, weapon firing logic, and rendering pipeline are entirely invisible to CI. The project's stated quality gate (82% coverage) is a paper goal. Core gameplay loops have no automated regression protection.
- **Closing the Gap**: Two approaches, ranked by effort:
  1. **Low effort**: Add `Xvfb` to the Linux CI workflow step (`apt-get install xvfb && Xvfb :99 -screen 0 1024x768x24 & export DISPLAY=:99`) before running `go test`. This unblocks all 22 packages without code changes.
  2. **High effort / more correct**: Introduce a `pkg/testutil/ebitenmock` package that provides stub implementations of the `ebiten.Image` interface, and add `//go:build ebitengine` guards to files that reference the Ebitengine graphics types directly. This separates business logic from rendering in tests.

---

## Gap 5: Config Watch Has a Confirmed Data Race

- **Stated Goal**: Configuration hot-reload via `pkg/config/Watch()` — callers can register a callback and receive config changes at runtime without restart.
- **Current State**: `go test -race ./pkg/config/...` exits with `--- FAIL: TestWatch_MultipleWatchers` and multiple `WARNING: DATA RACE` reports. The race is between `reloadConfiguration()` — called from the viper file-watcher goroutine — and `viper.Reset()` called in the test. While the race is partially a test-isolation issue, it also exposes that the config watcher goroutine does not honour its cancellation context before making viper calls, and that there is no synchronisation barrier between stopping the watcher and subsequent viper operations.
- **Impact**: In production, a config hot-reload that races with application startup or shutdown could corrupt the live `C Config` global, causing incorrect FOV values, wrong audio volumes, or network timeouts. The race detector failure means CI cannot be run with `-race` on `pkg/config`.
- **Closing the Gap**: In `TestWatch_MultipleWatchers`, call the stop function returned by `Watch()` and add a `time.Sleep(50 * time.Millisecond)` settle delay before `viper.Reset()`. In `reloadConfiguration()` (`watch.go:63`), check `ctx.Err() != nil` before calling `viper.Unmarshal` to prevent stale-context reloads after cancellation.

---

## Gap 6: Enhanced Profanity Filter (Variant Detection)

- **Stated Goal**: "Profanity filter framework with procedural wordlist generation" and variant detection for l33t speak.
- **Current State**: `pkg/chat/filter.go` implements basic single-character l33t substitutions (e→3, a→4, i→1, etc.). Unicode homoglyphs (Cyrillic `а` for Latin `a`), double substitutions (`@@`), and phonetic evasions are not handled. Acknowledged in the previous GAPS.md as "basic filter works, variant detection limited."
- **Impact**: Motivated bad actors can trivially bypass the filter with `sh1t`, `f@ck`, homoglyph substitutions, or Unicode variants. This undermines the stated content-safety posture for multiplayer chat, particularly relevant for any platform or storefront submission.
- **Closing the Gap**: (a) Apply Unicode normalisation (`golang.org/x/text/unicode/norm.NFKD`) to each candidate string before matching to collapse homoglyphs. (b) Extend the substitution map in `generateLeetSpeakVariants()` with a second-pass that handles double-character patterns. (c) Add a homoglyph table for the most common Cyrillic/Greek lookalikes. Estimated effort: 4–8 hours. Validation: new `TestProfanityFilterVariants` cases pass.

---

## Gap 7: Deprecated API Presence Without Migration Path

- **Stated Goal**: "Mod loader and plugin API" — WASM-based sandboxed mod execution is the documented standard.
- **Current State**: Three deprecated APIs remain exported without formal deprecation markers:
  - `pkg/mod/mod.go:316` — `PluginManager()` (returns unsafe Go plugin manager)
  - `pkg/mod/mod.go:333` — `RegisterPlugin()` (unsafe plugin registration)
  - `pkg/status/status.go:130,184` — two deprecated status system methods
  - `pkg/ui/mod_browser.go:419,424` — two deprecated navigation methods
  These are documented with `// DEPRECATED:` comments but do not use Go 1.21's `//go:deprecated` annotation that tools and IDEs recognise.
- **Impact**: Mod authors or contributors may accidentally use the deprecated `PluginManager` path, bypassing WASM sandboxing and introducing security vulnerabilities. IDEs and `go vet` will not warn callers.
- **Closing the Gap**: Add `// Deprecated: Use WASMLoader() instead.` (Go doc convention) and the `//go:deprecated` pragma on each deprecated exported identifier. Create a `MIGRATION.md` or inline the migration note in the deprecated function's doc comment.

---

## Deferred Items (v6.1+, intentionally out of scope)

The following items remain intentionally deferred from this audit and are tracked for future milestones:

- **Mod Marketplace**: Centralised mod distribution platform — requires player base first.
- **Mobile Store Publishing**: iOS/Android submission workflows — blocked on store policy review.
- **Cross-Save Sync**: Cloud save synchronisation across devices — requires backend infrastructure decision.

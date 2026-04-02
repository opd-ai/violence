# Ebitengine Game Audit Report

Generated: 2026-04-02T03:22:55Z
Repository: opd-ai/violence
Auditor: Copilot automated static analysis

## Executive Summary

- **Total Issues**: 42
- **Critical**: 6 — Crashes, game-breaking bugs, deadlocks
- **High**: 10 — Major functionality/UX problems
- **Medium**: 14 — Noticeable bugs, moderate impact
- **Low**: 4 — Minor issues, edge cases
- **Optimizations**: 4 — Performance improvements
- **Code Quality**: 4 — Maintainability concerns

This audit covers collision detection, UI architecture, input handling, rendering pipeline, state management, game logic, performance, asset management, Ebitengine-specific patterns, and code quality. All findings are based on static analysis only — no code was executed.

---

## Critical Issues

### [C-001] State Mutation in Draw() — HUD and StatusBar Update

- **Location**: `main.go:5538-5543`
- **Category**: Rendering / State Management
- **Description**: The `drawPlaying()` method (called from `Draw()`) invokes `g.hud.Update()` and `g.statusBarSystem.UpdatePlayerStatusBar()`, both of which mutate game state. Ebitengine's `Draw()` method should be read-only — state changes belong exclusively in `Update()`. Since Ebitengine may skip `Draw()` frames when the display is behind, or call `Draw()` at a different rate than `Update()`, these mutations create frame-rate-dependent behavior and state desynchronization.
- **Impact**: HUD state (health/ammo displays, message timers) ticks only when frames render. At 30 FPS, HUD updates half as fast as at 60 FPS. In multiplayer/replay, the non-deterministic update order breaks reproducibility.
- **Reproduction**:
  1. Run the game with `--max-tps 120` while display runs at 60 Hz.
  2. Observe that HUD message timers expire at half the expected rate.
- **Root Cause**: State-mutating calls placed in the rendering path instead of the update path.
- **Suggested Fix**: Move `g.hud.Update()` and `g.statusBarSystem.UpdatePlayerStatusBar()` into `updatePlaying()`.

### [C-002] Division by Zero in Camera Transform

- **Location**: `main.go:5775` and `main.go:7599`
- **Category**: Rendering / Math
- **Description**: Both `transformToCamera()` and `calculateLootTransform()` compute `invDet = 1.0 / (planeX*camera.DirY - camera.DirX*planeY)` without checking for a zero determinant. When the camera plane is parallel to the view direction (degenerate projection matrix), this produces `±Inf`, which propagates through all sprite/prop/loot screen position calculations, causing NaN corruption in rendering state.
- **Impact**: Game renders props and loot at invalid screen positions or crashes with NaN propagation. Triggers when FOV or camera direction reaches certain edge values.
- **Reproduction**:
  1. Set FOV to extreme values (approaching 0° or 180°).
  2. Observe sprite positions become corrupt.
- **Root Cause**: Missing guard against degenerate determinant.
- **Suggested Fix**: Add `if math.Abs(det) < 1e-10 { return 0, 0 }` guard before division.

### [C-003] Chat Relay Deadlock — Blocking I/O Under Read Lock

- **Location**: `pkg/chat/relay.go:176-211`
- **Category**: State Management / Deadlock
- **Description**: `routeMessage()` holds `rs.mu.RLock()` while calling `broadcastMessage()` → `sendMessage()` → `conn.Write()`. If a client's TCP send buffer is full, `Write()` blocks indefinitely while holding the read lock. Any goroutine attempting to acquire a write lock (e.g., `handleClient()` adding/removing clients) deadlocks. This is a classic lock-ordering violation.
- **Impact**: Chat relay server hangs completely. All connected clients lose chat functionality. If chat is on the game server goroutine, the entire server stalls.
- **Reproduction**:
  1. Connect multiple chat clients.
  2. One client stops reading from its connection (simulating a slow consumer).
  3. Another client sends a broadcast message.
  4. Server deadlocks when `Write()` blocks on the slow consumer's connection.
- **Root Cause**: I/O operations performed while holding a mutex.
- **Suggested Fix**: Copy the client map snapshot under lock, release lock, then perform I/O on the snapshot.

### [C-004] Collision Division by Zero in Slide Movement

- **Location**: `pkg/collision/sliding.go:280-286`
- **Category**: Collision / Floating-Point
- **Description**: `updateRemainingMovement()` computes `slideDistance = sqrt(slideX² + slideY²)` then divides by it at line 285-286 without checking for zero. When both slide components are zero (entity pressed flat against a wall corner), this produces NaN, corrupting the entity's remaining movement vector and potentially its position.
- **Impact**: Entity position becomes NaN after sliding into certain wall configurations, effectively removing the entity from the game world.
- **Reproduction**:
  1. Move player diagonally into an interior wall corner where both X and Y slide vectors cancel to zero.
  2. Collision system produces NaN position.
- **Root Cause**: Missing zero-distance guard in normalization.
- **Suggested Fix**: Add `if slideDistance < 1e-10 { *remainingX = 0; *remainingY = 0; return }` before normalization.

### [C-005] Raycast Returns First Agent, Not Closest

- **Location**: `main.go:2630-2647`
- **Category**: Collision / Logic
- **Description**: `createEnemyRaycastFunction()` iterates `g.aiAgents` and returns immediately upon finding the first agent in range with a positive dot product. It does not track minimum distance, so if agent at index 0 is 50 units away and agent at index 5 is 2 units away, the distant agent is hit instead of the closer one.
- **Impact**: Weapon fire inconsistently hits wrong enemies. Players shoot enemies they're not aiming at. Combat feels broken and unpredictable.
- **Reproduction**:
  1. Position multiple enemies in a line in front of the player.
  2. Fire at the nearest enemy.
  3. Observe the first enemy in the `aiAgents` slice (by index) takes damage instead.
- **Root Cause**: Early return in loop without tracking minimum distance.
- **Suggested Fix**: Track closest hit across the entire loop and return only after checking all agents.

### [C-006] Raycaster FOV Not Validated — Tangent Overflow

- **Location**: `pkg/raycaster/raycaster.go:20-27` and `pkg/raycaster/raycaster.go:51-52`
- **Category**: Rendering / Math
- **Description**: `NewRaycaster()` accepts any `fov` value without validation. At FOV=180°, `Tan(FOV * π / 360)` = `Tan(π/2)` = ±∞, producing infinite camera plane vectors. At FOV=0° or negative, the plane degenerates to zero. Both cases corrupt the entire raycasting pipeline.
- **Impact**: Rendering breaks completely at extreme FOV values. If FOV is configurable or modified by a mod, this is a crash vector.
- **Root Cause**: Missing input validation on constructor parameter.
- **Suggested Fix**: Clamp FOV to [10°, 170°] in `NewRaycaster()`.

---

## High Priority Issues

### [H-001] Hazard Damage Calculation — Armor Overflow Logic Error

- **Location**: `main.go:3487-3496`
- **Category**: Logic
- **Description**: When armor absorbs hazard damage and goes negative, the overflow calculation at line 3492 (`healthDamage = -g.hud.Armor`) captures the negative armor value after subtraction. While mathematically this happens to be the correct overflow amount, the code then sets armor to 0 at line 3493 *after* reading the overflow. More critically, the `else` branch at line 3495 applies `damage / 2` to health even when armor fully absorbed the hit, effectively double-counting damage — armor absorbs half and health takes half regardless.
- **Impact**: Players take incorrect damage from environmental hazards. Balance is skewed — hazards are more lethal than intended because both armor and health absorb full halves.
- **Suggested Fix**: Calculate armor absorption, overflow, and health damage in a single clear sequence before modifying any state.

### [H-002] Renderer Tick() Called in Draw Path

- **Location**: `main.go:5316`
- **Category**: Rendering / State
- **Description**: `setupRenderer()` calls `g.renderer.Tick()` which increments the renderer's internal frame counter. This is called from `drawPlaying()` (within `Draw()`), making the renderer's animation timing frame-rate dependent rather than tick-rate dependent.
- **Impact**: Texture animations, shader effects, and any time-based rendering run at display refresh rate instead of game tick rate. At 144 Hz, effects are 2.4x faster than at 60 Hz.
- **Suggested Fix**: Move `g.renderer.Tick()` to `updatePlaying()`.

### [H-003] Hardcoded 60 FPS Delta Time for Camera Effects

- **Location**: `main.go:2215`
- **Category**: Logic / Frame-Rate Dependency
- **Description**: `g.cameraFXSystem.Update(1.0 / 60.0)` passes a fixed 16.67ms delta regardless of actual frame rate. Ebitengine's TPS is configurable and defaults to 60 but can be changed.
- **Impact**: Camera shake, flash, and zoom effects have incorrect duration at non-60 TPS rates. At 30 TPS, effects last twice as long; at 120 TPS, half as long.
- **Suggested Fix**: Use `1.0 / float64(ebiten.TPS())` or pass `common.DeltaTime`.

### [H-004] Status Effect Duration Frame-Rate Dependent

- **Location**: `pkg/status/status.go:238`
- **Category**: Logic / Frame-Rate Dependency
- **Description**: Status effect `TimeRemaining` is decremented by a fixed `16 * time.Millisecond` per frame, assuming exactly 60 FPS. At different frame rates, effects expire at incorrect times.
- **Impact**: Poison, burn, and bleed effects last unpredictable durations. At 30 FPS they last ~2x longer than intended; at 120 FPS they last ~0.5x. Gameplay balance varies with hardware performance.
- **Suggested Fix**: Use actual elapsed time: `effect.TimeRemaining -= time.Since(lastUpdate)`.

### [H-005] Minimap Allocates 2D Array Every Frame

- **Location**: `main.go:7267-7274` and `main.go:7302-7309`
- **Category**: Performance
- **Description**: Both `drawAutomap()` and `drawCollapsibleAutomap()` allocate a new `walls [][]bool` 2D slice every frame. For a 64×64 map, this is 65 allocations per frame (1 outer + 64 inner slices), totaling ~3,900 allocations/second at 60 FPS.
- **Impact**: Significant GC pressure causing frame stutters, especially noticeable when minimap is visible. Memory allocator thrashing degrades overall performance.
- **Suggested Fix**: Pre-allocate the `walls` cache in the Game struct and reuse it each frame, only rebuilding when the map changes.

### [H-006] Polygon-Polygon Collision Missing SAT Edge Checks

- **Location**: `pkg/collision/collision.go:256-284`
- **Category**: Collision
- **Description**: The polygon-polygon collision test only checks vertex containment (whether vertices of polygon A lie inside polygon B and vice versa). It does not perform Separating Axis Theorem (SAT) edge-normal projection tests. This means two thin polygons can overlap without any vertex being inside the other — a known failure mode of vertex-only containment tests.
- **Impact**: Thin or elongated collision shapes (e.g., walls, projectile hitboxes) can pass through each other undetected. Entities tunnel through polygon obstacles.
- **Suggested Fix**: Implement proper SAT with edge normal projections, or add edge-to-edge intersection tests.

### [H-007] Particle System activeIndices Unbounded Growth

- **Location**: `pkg/particle/particle.go:94`
- **Category**: Performance / Memory Leak
- **Description**: `Spawn()` unconditionally appends the particle index to `activeIndices`. When particles are deactivated and their slots reused, the same index gets appended again. The `Update()` method rebuilds the slice each frame, but the append in `Spawn()` grows the underlying array capacity beyond the pool size over time.
- **Impact**: Memory leak proportional to total particle spawns over game lifetime. After thousands of particle cycles, the slice's backing array grows unbounded.
- **Suggested Fix**: In `Update()`, verify the active indices are rebuilt from scratch each frame and the old slice is released, or use a fixed-size bitset to track active particles.

### [H-008] Config Save() Uses Read Lock for Write Operations

- **Location**: `pkg/config/config.go:74-96`
- **Category**: State Management / Data Race
- **Description**: `Save()` acquires `mu.RLock()` but then calls `viper.Set()` (a write operation) 14 times and `viper.WriteConfig()` (file I/O). Multiple concurrent `Save()` calls can race on viper's internal state. Additionally, concurrent `Get()` calls also hold RLock, creating a read-write race on viper's map.
- **Impact**: Configuration corruption — saved config file may contain a mix of old and new values. Lost settings changes on concurrent access.
- **Suggested Fix**: Use `mu.Lock()` (write lock) in `Save()`.

### [H-009] Global Genre Variable Without Synchronization

- **Location**: `pkg/engine/engine.go:129-138`
- **Category**: State Management / Data Race
- **Description**: The package-level `currentGenre` string variable is read and written without any synchronization. `SetGenre()` and `GetCurrentGenre()` can race with each other and with any system reading the genre during initialization.
- **Impact**: Genre configuration may be inconsistent across systems during startup or genre changes, causing mismatched visual/audio/gameplay parameters.
- **Suggested Fix**: Use `sync.RWMutex` or `atomic.Value` to protect the global genre variable.

### [H-010] isWalkable Returns True for Nil Map

- **Location**: `main.go:3696-3699`
- **Category**: Collision / Logic
- **Description**: When `g.currentMap` is nil or empty, `isWalkable()` returns `true`, allowing the player to walk anywhere. This can occur during level transitions or if map generation fails.
- **Impact**: Player escapes level bounds, falls into void, or enters undefined map space. Can corrupt game state if other systems assume valid map coordinates.
- **Suggested Fix**: Return `false` when map is nil — prevent all movement until a valid map is loaded.

---

## Medium Priority Issues

### [M-001] Virtual Joystick Incorrect Normalization

- **Location**: `pkg/input/touch.go:50-55`
- **Category**: Input / Math
- **Description**: The joystick normalization uses `scale = 1.0 / (magnitude * 0.5)` where `magnitude` is already squared (`x*x + y*y`). The correct normalization would be `scale = 1.0 / math.Sqrt(magnitude)`. The current formula produces incorrect axis values that don't clamp to the unit circle properly.
- **Impact**: Mobile players experience erratic movement sensitivity. Diagonal input values exceed the expected [-1, 1] range.
- **Suggested Fix**: Replace with `scale := 1.0 / math.Sqrt(magnitude)`.

### [M-002] No Collision Tunneling Prevention (CCD)

- **Location**: `main.go:3635-3647`
- **Category**: Collision
- **Description**: `handleCollisionAndMovement()` only tests the final destination position and two axis-aligned fallback positions. No swept/continuous collision detection is performed. At high speeds (large `deltaX`, `deltaY`), the player can skip past thin walls entirely.
- **Impact**: Players clip through walls at high movement speeds or during speed boosts. Diagonal movement into corners creates tunneling opportunities.
- **Suggested Fix**: Implement stepped movement (subdivide delta into small increments) or swept AABB collision.

### [M-003] Mod Loading Errors Silently Ignored

- **Location**: `main.go:5116`
- **Category**: State Management / Error Handling
- **Description**: `g.modLoader.LoadMod(modPath)` return value is discarded with `_ =`. If a mod fails to load (corrupted WASM, invalid manifest, permission error), no error is logged and the player receives no feedback.
- **Impact**: Players install mods that silently fail. Debugging mod issues becomes impossible. Malformed mods could partially initialize, leaving the mod system in an inconsistent state.
- **Suggested Fix**: Log the error with `logrus.WithError(err).Warn("Failed to load mod")` and optionally display a toast notification.

### [M-004] PostProcessor Allocates Large Buffers Per Effect

- **Location**: `pkg/render/postprocess.go:167, 207, 227`
- **Category**: Performance
- **Description**: `ApplyChromaticAberration()` allocates `make([]byte, len(framebuffer))`, `extractBrightPixels()` allocates `make([]float64, len(framebuffer))`, and `blurBrightPixels()` allocates another `make([]float64, len(brightPixels))`. At 320×200 resolution, each allocation is 256KB. Three allocations per frame = 768KB/frame of garbage.
- **Impact**: GC pressure from post-processing causes micro-stutters. At higher resolutions, allocations grow quadratically.
- **Suggested Fix**: Pre-allocate scratch buffers in the `PostProcessor` struct and reuse them each frame.

### [M-005] Weapon Arsenal GetCurrentWeapon Missing Bounds Check

- **Location**: `pkg/weapon/weapon.go:276-277`
- **Category**: Logic / Panic Risk
- **Description**: `GetCurrentWeapon()` accesses `a.Weapons[a.CurrentSlot]` without validating that `CurrentSlot < len(Weapons)`. If `CurrentSlot` is corrupted (e.g., from deserialization or weapon switching edge cases), this panics with an index out of range.
- **Impact**: Game crash on invalid weapon slot access.
- **Suggested Fix**: Add bounds validation: `if a.CurrentSlot >= len(a.Weapons) { return a.Weapons[0] }`.

### [M-006] A* Pathfinding Uses Linear Search for Open Set

- **Location**: `pkg/ai/pathfinding.go:63-77, 80-85`
- **Category**: Performance
- **Description**: The A* open set is a plain slice. `findLowestFNodeCoord()` performs O(n) linear search per iteration. With `maxIter=1000` and growing open sets, this produces up to O(n²) behavior. Additionally, removing the current node at line 67 uses slice append which is O(n).
- **Impact**: AI pathfinding causes frame hitches when multiple enemies pathfind simultaneously on large maps. CPU spikes proportional to (enemies × map complexity).
- **Suggested Fix**: Replace the open set with a min-heap (`container/heap`).

### [M-007] Head Bob Phase Accumulates Without Wrapping

- **Location**: `pkg/camera/camera.go:62`
- **Category**: Logic / Floating-Point Precision
- **Description**: `c.headBobPhase += movementMagnitude * HeadBobFrequency` accumulates without ever wrapping to [0, 2π]. After extended gameplay, floating-point precision loss in `math.Sin()` causes jittery or incorrect head bob values.
- **Impact**: After many hours of continuous movement, head bob becomes visually glitchy. The phase value eventually loses sub-cycle precision.
- **Suggested Fix**: Add `c.headBobPhase = math.Mod(c.headBobPhase, 2*math.Pi)` after accumulation.

### [M-008] Network Coop Session Lock-Then-Access Pattern

- **Location**: `pkg/network/coop.go:367-389`
- **Category**: State Management / Race Condition
- **Description**: `RespawnPlayer()` reads `playerState` from the map under `RLock`, releases the lock, then accesses `playerState` fields. Between releasing and accessing, another goroutine could remove or modify the player entry.
- **Impact**: Potential use-after-free or nil pointer dereference during player respawn in co-op mode.
- **Suggested Fix**: Keep the lock held for the entire operation, or use per-player fine-grained locking.

### [M-009] Chat NewChatWithKey Panics Instead of Returning Error

- **Location**: `pkg/chat/chat.go:45-54`
- **Category**: Error Handling
- **Description**: `NewChatWithKey()` panics if the key length is not 32 bytes. In a production application, this should return an error rather than crash.
- **Impact**: Application crash if key derivation produces wrong-length output. Panic in library code is difficult to recover from.
- **Suggested Fix**: Change signature to return `(*Chat, error)` and return an error for invalid key lengths.

### [M-010] Point-in-Polygon Edge Case on Boundaries

- **Location**: `pkg/collision/collision.go:329-349` (via `pointInPolygonPooled`)
- **Category**: Collision / Boundary Condition
- **Description**: The ray-casting point-in-polygon algorithm uses strict `>` inequality for the Y-axis test. Points lying exactly on polygon edges produce inconsistent results depending on floating-point rounding.
- **Impact**: Collision detection fails intermittently at polygon edges. Attacks and movement at wall boundaries sometimes don't register correctly.
- **Suggested Fix**: Use `>=` for one boundary and `>` for the other (standard convention), or add epsilon tolerance.

### [M-011] Wall Rendering Overwrites Floor on Semi-Transparent Pixels

- **Location**: `pkg/render/render.go:122-127`
- **Category**: Rendering
- **Description**: The rendering pipeline checks `wallColor.A > 0` to decide whether to overwrite the floor/ceiling color. Any semi-transparent wall pixel (alpha 1-254) replaces the floor color entirely instead of blending, creating visual artifacts on transparent surfaces.
- **Impact**: Semi-transparent walls render as opaque, covering floor/ceiling detail underneath.
- **Suggested Fix**: Either require `wallColor.A == 255` for full overwrite, or implement proper alpha blending: `c = blend(floor, wall, wallColor.A/255.0)`.

### [M-012] Inventory Thread-Safety Workaround Creates Shallow Copy

- **Location**: `main.go:5128`
- **Category**: State Management
- **Description**: `convertInventoryToSaveItems()` attempts thread-safe access by re-assigning the items slice: `inv.Items = append([]inventory.Item{}, inv.Items...)`. This creates a shallow copy of the slice header but items containing pointers still share underlying data. The mutation of `inv.Items` itself is not protected by any lock.
- **Impact**: Save/load can capture partially modified inventory state. Items with pointer fields can be corrupted during serialization.
- **Suggested Fix**: Use the inventory's lock mechanism if available, or deep-copy items for serialization.

### [M-013] Layout() Ignores outsideWidth/outsideHeight Parameters

- **Location**: `main.go:8108-8118`
- **Category**: Ebitengine-Specific
- **Description**: The `Layout()` method receives `outsideWidth, outsideHeight` (actual window size) but passes `config.C.InternalWidth/Height` to UI systems instead. The UI layout manager never receives the actual window dimensions, so it cannot adapt to window resizing.
- **Impact**: UI positioning may not correctly account for window size changes if internal resolution differs from window size. Minimap and HUD elements use hardcoded offsets relative to internal resolution.
- **Suggested Fix**: Store and expose the outside dimensions for UI systems that need to know the actual window size.

### [M-014] BSP Recursion Depth Hardcoded to 10

- **Location**: `pkg/bsp/bsp.go:191-193`
- **Category**: Logic
- **Description**: BSP tree splitting depth is hardcoded to 10. For larger maps, this prevents proper room generation, leaving large undivided regions.
- **Impact**: Large maps have poor room variety and oversized empty spaces. Level quality degrades at higher map sizes.
- **Suggested Fix**: Derive max depth from map dimensions: `maxDepth = int(math.Log2(float64(maxDimension/minRoomSize)))`.

---

## Low Priority Issues

### [L-001] UI Button Edge Pixel Dead Zone

- **Location**: `pkg/ui/interactive.go:123-126`
- **Category**: UI
- **Description**: Button hit testing uses `float32` comparison without epsilon tolerance. Pixels at exact button boundaries may not register clicks due to float32 precision.
- **Impact**: 1-2 pixel dead zones at button edges; very rare user-visible impact.
- **Suggested Fix**: Add small epsilon or use integer comparison after rounding.

### [L-002] Color Lerp Unclamped Parameter

- **Location**: `pkg/ui/interactive.go:303-309`
- **Category**: UI / Math
- **Description**: `lerpColor()` does not clamp the `t` parameter to [0, 1]. If `t > 1.0`, the interpolated color values can exceed 255 and wrap around when cast to `uint8`.
- **Impact**: UI color transitions can produce flash artifacts if animation timing overshoots. Rare in practice.
- **Suggested Fix**: Add `t = math.Max(0, math.Min(1, t))` at function entry.

### [L-003] Variable Named g_ in Flicker Calculation

- **Location**: `main.go:6305-6310`
- **Category**: Code Quality
- **Description**: The variable `g_` is used to avoid shadowing the `g` receiver variable. While functionally correct, this is confusing naming that could lead to mistakes in future modifications.
- **Impact**: Code readability issue. No functional bug.
- **Suggested Fix**: Rename to `greenChannel` or `flickerG` for clarity.

### [L-004] Manhattan Distance Heuristic for Grid Pathfinding

- **Location**: `pkg/ai/pathfinding.go:25-32`
- **Category**: Logic
- **Description**: A* uses Manhattan distance heuristic, which is admissible for 4-directional movement. However, if diagonal movement is ever added, the heuristic becomes inadmissible and will produce suboptimal paths.
- **Impact**: No current impact. Future risk if diagonal movement is added without updating the heuristic.
- **Suggested Fix**: Document the 4-directional assumption, or switch to Chebyshev/Euclidean distance proactively.

---

## Performance Optimization Opportunities

### [P-001] Post-Processing Buffer Allocation Per Frame

- **Location**: `pkg/render/postprocess.go:167, 207, 227`
- **Current Impact**: ~768KB of garbage per frame at 320×200 resolution. At 60 FPS = ~45MB/s of allocations triggering frequent GC pauses.
- **Optimization**: Pre-allocate `originalBuf`, `brightPixelsBuf`, and `blurredBuf` in the `PostProcessor` struct during construction. Reuse each frame.
- **Expected Improvement**: Eliminate 3 allocations per frame. Reduce GC pause frequency by ~30% during post-processing.

### [P-002] Minimap Wall Array Allocation Per Frame

- **Location**: `main.go:7267-7274`, `main.go:7302-7309`
- **Current Impact**: 65 allocations per frame (1 outer + 64 inner slices) × 2 minimap functions = 130 allocations/frame when both minimaps are active.
- **Optimization**: Cache the `walls` 2D slice in the Game struct. Rebuild only when the map changes (level transitions).
- **Expected Improvement**: Eliminate ~130 allocations/frame. Measurable improvement in frame time consistency.

### [P-003] UI Update Method Allocations

- **Location**: `main.go:4256, 4294, 4299, 4414-4416, 4460-4469`
- **Current Impact**: Shop, crafting, and skills UI screens allocate new slices every frame they're active. `make([]ui.ShopItem, len(allItems))` and similar calls create garbage proportional to inventory/recipe/skill count.
- **Optimization**: Pre-allocate UI data slices and reuse them with length reset.
- **Expected Improvement**: Eliminate variable-size allocations in UI hot paths. Smoother frame times during menu interactions.

### [P-004] Box Blur in Bloom is O(n × radius²) Per Frame

- **Location**: `pkg/render/postprocess.go:225-250`
- **Current Impact**: Box blur iterates `width × height × (2×radius+1)²` pixels. At default radius=3, this is 49 texture reads per pixel = ~3.1M reads for 320×200 resolution per frame.
- **Optimization**: Replace box blur with two-pass separable blur (horizontal then vertical), reducing complexity from O(n × r²) to O(n × 2r).
- **Expected Improvement**: ~7x speedup for bloom post-processing at radius=3.

---

## Code Quality Observations

### [Q-001] Magic Numbers Throughout main.go

- **Location**: `main.go:592` (0.3), `main.go:594` (1024), `main.go:607` (64), `main.go:616` (64.0), `main.go:622` (100), `main.go:631` (2000), `main.go:674` (64), `main.go:695` (500), `main.go:702-703` (200)
- **Issue**: Numerous numeric literals without named constants. Makes tuning game parameters difficult and error-prone.
- **Suggestion**: Extract to named constants with descriptive names: `const ShadowAmbience = 0.3`, `const InitialParticleCapacity = 1024`, etc.

### [Q-002] main.go Exceeds 8000 Lines

- **Location**: `main.go` (8,360 lines)
- **Issue**: Single file contains the entire game loop, all state management, rendering dispatch, input handling, UI drawing, collision, AI updates, and system initialization. This monolithic structure makes navigation, code review, and testing extremely difficult.
- **Suggestion**: Extract logical sections into separate files within the main package: `game_update.go`, `game_draw.go`, `game_init.go`, `game_collision.go`, `game_ui.go`, etc.

### [Q-003] Inconsistent Error Handling Patterns

- **Location**: Various — `main.go:3306` (ignored audio error), `main.go:5116` (ignored mod error), `pkg/telegraph/system.go:259` (unchecked type assertion)
- **Issue**: Mix of `_ = err`, unchecked type assertions, and silent failures. No consistent error handling strategy across the codebase.
- **Suggestion**: Establish a project convention: always log errors at minimum, use checked type assertions with `ok` pattern, never discard errors from I/O or external system calls.

### [Q-004] Known Gaps Not Cross-Referenced in Code

- **Location**: GAPS.md documents 7 gaps, but the affected code locations have no `// GAP-N:` markers.
- **Issue**: Developers modifying code near known gaps may not realize they're in a problem area. No automated way to find gap-related code.
- **Suggestion**: Add `// GAP-1: Chat key-exchange deadlock` comments at the affected code locations, cross-referencing GAPS.md entries.

---

## Recommendations by Priority

### 1. Immediate Action Required

- **[C-001]**: Move HUD/StatusBar Update() calls from Draw() to Update() — frame-rate dependent state mutation
- **[C-002]**: Add zero-determinant guard in camera transform — NaN crash vector
- **[C-003]**: Fix chat relay lock-held-during-I/O — production deadlock
- **[C-004]**: Add zero-distance guard in sliding collision — NaN position corruption
- **[C-005]**: Fix raycast to return closest enemy, not first in slice — combat is broken
- **[C-006]**: Validate FOV range in raycaster constructor — tangent overflow

### 2. High Priority (Next Sprint)

- **[H-001]**: Fix hazard damage armor overflow calculation
- **[H-002]**: Move renderer Tick() to Update()
- **[H-003]**: Use dynamic delta time for camera effects
- **[H-004]**: Use real elapsed time for status effect durations
- **[H-005]**: Cache minimap wall arrays to eliminate per-frame allocation
- **[H-008]**: Fix config Save() to use write lock
- **[H-009]**: Synchronize global genre variable
- **[H-010]**: Return false from isWalkable when map is nil

### 3. Medium Priority (Backlog)

- **[M-001]**: Fix virtual joystick normalization for mobile
- **[M-002]**: Add stepped collision detection for tunneling prevention
- **[M-003]**: Log mod loading errors instead of silently discarding
- **[M-004]**: Pre-allocate post-processing scratch buffers
- **[M-005]**: Add bounds check in GetCurrentWeapon()
- **[M-006]**: Use heap-based open set for A* pathfinding
- **[M-011]**: Implement alpha blending for semi-transparent walls
- **[M-013]**: Pass actual window dimensions to UI layout systems

### 4. Technical Debt

- **[Q-001]**: Extract magic numbers to named constants
- **[Q-002]**: Split main.go into logical sub-files
- **[Q-003]**: Establish consistent error handling convention
- **[Q-004]**: Cross-reference GAPS.md entries in source code

---

## Testing Recommendations

### Critical Test Scenarios

1. **Camera transform edge cases**: Test `transformToCamera()` with zero and near-zero determinant values. Assert no NaN/Inf output.
2. **Raycast closest-enemy**: Create test with 3+ enemies at known positions; verify the closest is always hit first.
3. **Hazard damage with partial armor**: Test damage application when armor is less than half the incoming damage. Verify exact health/armor values.
4. **Status effect duration**: Run effect system at simulated 30, 60, and 120 FPS. Verify consistent real-time duration.
5. **Collision corner cases**: Position player at exact grid boundaries (e.g., x=5.0, y=3.0) and verify collision works correctly.
6. **Config concurrent access**: Run `Save()` and `Get()` concurrently with `-race` flag. Currently fails (GAPS.md Gap 5).

### Input Edge Cases

7. **Extreme FOV values**: Set FOV to 1°, 10°, 90°, 170°, 179° and verify rendering doesn't corrupt.
8. **Virtual joystick at max displacement**: Verify output is clamped to unit circle on all axes.
9. **Simultaneous UI and gameplay input**: Press menu key while firing — verify input is not consumed by both systems.

### Performance Benchmarks

10. **Post-processing allocation**: Benchmark `ApplyBloom` with and without pre-allocated buffers. Target: 0 allocations per call.
11. **Minimap rendering**: Benchmark `drawAutomap` with 64×64 and 128×128 maps. Target: 0 allocations per frame.
12. **A* pathfinding**: Benchmark with 10 concurrent pathfind requests on 64×64 map. Target: <1ms total per frame.
13. **Particle system lifecycle**: Spawn and expire 10,000 particles over 100 frames. Monitor `activeIndices` slice capacity.

---

## Audit Methodology Notes

### Analysis Approach

- **Static analysis only**: All findings are based on source code inspection without executing the codebase.
- **Line-by-line review** of `main.go` (8,360 lines) and 15 critical packages (`collision`, `ui`, `input`, `render`, `raycaster`, `lighting`, `camera`, `combat`, `weapon`, `ai`, `network`, `audio`, `engine`, `config`, `chat`).
- **Pattern matching** for known Ebitengine anti-patterns: state mutation in Draw(), Layout() ignoring parameters, frame-rate dependent logic, missing error returns.
- **Cross-reference** with GAPS.md to avoid duplicating known issues while identifying new ones.

### Areas Not Covered

- **Full pkg/ audit**: 94+ packages exist; only the 15 most critical were reviewed in depth. Packages like `decoration`, `lore`, `dialogue`, `quest`, `skills`, `economy` received cursory examination only.
- **WASM-specific paths**: Build tags and WASM-specific behavior were noted but not deeply audited.
- **Integration test coverage**: Existing tests were referenced for context but not run or verified.
- **Third-party dependency audit**: External dependencies (libp2p, wasmer-go, AWS SDK) were not audited for vulnerabilities.
- **Mobile/touch platforms**: Touch input was reviewed but physical device behavior could not be validated.

### Assumptions

- Ebitengine v2.8.8 semantics: `Update()` runs at fixed TPS, `Draw()` runs at display refresh rate, `Layout()` may be called at any time.
- Default TPS is 60 unless overridden by `ebiten.SetTPS()`.
- The `common.DeltaTime` value correctly tracks inter-frame time (not verified in this audit).
- `sync.Once.Do()` is idempotent and safe against double-close (verified: correct per Go specification).

### Limitations of Static Analysis

- **Race conditions**: Identified by pattern analysis but cannot be confirmed without runtime `-race` detector.
- **Performance impact**: Allocation counts estimated but actual frame time impact requires profiling.
- **Floating-point edge cases**: Division-by-zero and NaN paths identified structurally but triggering conditions may be rare in practice.
- **Dead code**: Some identified issues may be in code paths that are never reached in normal gameplay.

---

## Positive Observations

### Well-Implemented Patterns

1. **Object pooling (`pkg/pool/`)**: The pool package provides a well-designed memory pooling system with type-safe pools for commonly allocated types. Used effectively in collision detection to avoid per-frame allocations for polygon vertex slices.

2. **Spatial indexing (`pkg/spatial/`)**: Proper grid-based spatial partitioning is implemented and integrated with collision detection, avoiding O(n²) entity-vs-entity checks.

3. **Deterministic RNG (`pkg/rng/`)**: The PCG-based RNG package correctly uses `uint64` seeds and provides deterministic output. Seed derivation with XOR constants follows good practice for hierarchical generation.

4. **Renderer framebuffer reuse (`pkg/render/render.go:38,55`)**: The renderer correctly pre-allocates both the byte framebuffer and the `ebiten.Image`, using `WritePixels()` to avoid per-frame WebGL texture leaks in WASM.

5. **Clean ECS design (`pkg/engine/`)**: The Entity-Component-System architecture is minimal and well-structured. Component storage by `reflect.Type` provides O(1) access. System registration order provides deterministic update ordering.

6. **Genre system architecture**: The `SetGenre()` interface is consistently implemented across packages, with most implementations covering all 5 genres and providing explicit defaults.

7. **Structured logging**: Consistent use of `logrus.WithFields()` with standard field names (`system`, `entity`, `player`, `seed`) throughout the codebase.

8. **Configuration hot-reload (`pkg/config/`)**: Despite the known data race (Gap 5), the architecture of file-watching with callback notification is well-designed and uses context-based cancellation.

9. **Collision system layering (`pkg/collision/`)**: Layer and mask-based collision filtering is properly implemented, allowing selective collision detection between entity types.

10. **Save system atomic writes (`pkg/save/`)**: Save file writes use temporary files with atomic rename, preventing corruption from interrupted writes.

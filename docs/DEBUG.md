# UI & Gameplay Debugging Prompt

Generic reusable prompt for autonomous diagnosis and repair of visual, navigation, and interaction bugs in the Violence FPS (Go/Ebitengine raycasting game).

---

## Prompt

```
**Objective:** Diagnose and fix all bugs preventing correct [SYMPTOM AREA] in the Violence FPS (Go/Ebitengine raycasting game).

**Execution mode:** Autonomous action — discover bugs, implement fixes, and validate with `go build ./...` and targeted tests.

**Investigation methodology:**

For each subsystem under investigation, apply these checks in order:

### 1. Tile Type Consistency Audit

The BSP generator defines tile constants in `pkg/bsp/bsp.go`:
- Empty=0, Wall=1, Floor=2, Door=3, Secret=4
- Genre walls=10–14, Genre floors=20–24

Every file that classifies, compares, or branches on tile values must handle
ALL tile types correctly. Specifically check for:

- **`> 0` wall tests:** Any `tile > 0` or `tile != 0` treats floors, doors,
  and genre tiles as walls. Must use `raycaster.IsWallTile()` or an equivalent
  range check instead.
- **Exhaustive switch/case:** Any `switch tile` that lists only `TileWall`,
  `TileFloor`, `TileDoor`, `TileSecret` without handling genre wall (10–14)
  and genre floor (20–24) ranges will misclassify tiles when a genre is active.
- **Hardcoded `== bsp.TileFloor`:** Adjacency checks that compare against
  `bsp.TileFloor` (value 2) only will fail when the active genre uses floor
  values 20–24. Use `isWalkableTile()` or equivalent instead.
- **Palette / color map coverage:** Any `map[int]color.RGBA` or
  `map[int]string` indexed by tile type must include keys for genre-specific
  wall types (10–14) or fall back gracefully. Missing keys produce zero-value
  colors (transparent), causing visual holes.

Files to audit: `main.go`, `pkg/raycaster/*.go`, `pkg/render/*.go`,
`pkg/bsp/*.go`, `pkg/level/*.go`, and any file that reads `g.currentMap`.

### 2. Coordinate & Spawn Validation

- **Hardcoded world positions:** Any entity spawned at fixed `(x, y)`
  coordinates (enemies, destructibles, items, NPCs) will land inside walls
  unless those coordinates are guaranteed to be walkable floor tiles. Spawns
  must use BSP room data (`bsp.GetRooms()`) or include a walkability check.
- **Tile-center offset:** Player and entity positions should be at tile center
  (`float64(tileX) + 0.5`) to avoid edge/corner clipping. Verify any
  `rooms[i].X + rooms[i].W/2` includes the `+ 0.5` offset.
- **Interaction distance:** `getInteractionTileCoords()` projects a point
  along the look direction. The distance must be ≤ 1.0 tiles to target the
  adjacent tile; values > 1.0 overshoot and miss the intended target.

### 3. State-Dependent Walkability

- **Doors:** `TileDoor` (3) must block movement until opened via interaction.
  After interaction, the tile is replaced with `TileFloor` on `g.currentMap`
  and the raycaster map is updated via `raycaster.SetMap()`. Verify both the
  walkability check and the raycaster map are updated atomically.
- **Secrets:** `TileSecret` (4) is a wall until triggered. After trigger it
  becomes `TileFloor`. Same update pattern as doors.
- **Opened-door rendering:** After a door tile is converted to floor, the
  raycaster must see it as non-solid. Confirm `IsWallTile(2) == false`.

### 4. Render Pipeline Checks

- **Wall rendering guard:** `renderWall()` skips pixels where
  `hit.WallType == 0` (empty space). Verify this does not inadvertently skip
  valid wall types.
- **Texture name mapping:** `getWallTextureName()` maps `WallType` int to
  texture atlas keys. Must cover genre wall types (10–14) and door/secret
  types (3, 4). Missing mappings fall through to `"wall_1"`.
- **Floor/ceiling rendering:** `renderFloor()` / `renderCeiling()` sample
  textures named `"floor_main"` / `"ceiling_main"`. Verify these are generated
  by the texture atlas for each genre.
- **Palette fallback:** When a texture atlas lookup fails, the renderer uses
  `r.palette[hit.WallType]`. If the palette lacks a key for that wall type,
  the zero-value `color.RGBA{0,0,0,0}` is transparent and the pixel vanishes.
- **Automap drawing:** `drawAutomap()` must recognize genre walls (10–14) and
  genre floors (20–24) in addition to base tile types.

### 5. Camera & Movement

- **Direction vector initialization:** `camera.DirX` and `camera.DirY` must be
  nonzero (typically `DirX=1.0, DirY=0.0`). Zero direction = zero movement
  deltas regardless of input.
- **Collision radius:** `isWalkable()` tests 4 corners at `±playerRadius`
  (0.25). Verify the bounding box check is symmetric and doesn't allow
  single-axis wall penetration.
- **Slide-along-wall:** `handleCollisionAndMovement()` tries full move, then
  X-only, then Y-only. Confirm the fallback logic doesn't silently discard
  both axes when only one is blocked.

### 6. Input → Action Pipeline

- Verify `input.Manager.Update()` is called each frame before movement
  processing.
- Verify `input.IsPressed()` / `input.IsJustPressed()` return correct state
  for keyboard, mouse, and gamepad.
- Verify mouse delta is captured via `input.MouseDelta()` and applied to
  camera rotation with correct sign and sensitivity scaling.
- Verify `ebiten.SetCursorMode(ebiten.CursorModeCaptured)` is active during
  `StatePlaying` so that mouse delta is available.

### 7. HUD & Menu Rendering

- **HUD visibility:** Confirm HUD draw calls occur after world rendering but
  before screen present. Verify HUD elements use screen-space coordinates, not
  world-space.
- **Font rendering:** `text.Draw()` requires a valid `font.Face`. Verify
  `basicfont.Face7x13` or the configured font is non-nil.
- **Menu state transitions:** Each `GameState` must have both an `update*` and
  `draw*` handler in the main switch. A missing handler causes a blank screen
  or frozen input.
- **Settings persistence:** Verify `config.C` is reloaded after settings
  changes so that FOV, sensitivity, and resolution changes take effect.

---

**Key files to examine:**
- `main.go` — game loop, movement, collision, level init, spawn, draw dispatch
- `pkg/bsp/bsp.go` — BSP map generation, tile constants
- `pkg/raycaster/raycaster.go` — DDA raycasting, `IsWallTile()`, `SetMap()`
- `pkg/render/render.go` — wall/floor/ceiling rendering, palette, texture mapping
- `pkg/camera/camera.go` — camera state, direction, rotation, head-bob
- `pkg/input/input.go` — input manager, key/mouse/gamepad state
- `pkg/ui/*.go` — HUD, menus, settings screens
- `pkg/door/door.go` — door state machine
- `pkg/level/tilemap.go` — `TileType` enum (note: `level.TileType` is `uint8`
  vs `bsp` tile `int` — potential type mismatch)
- `pkg/texture/texture.go` — procedural texture atlas generation
- `pkg/lighting/*.go` — sector light map, flashlight cone

**Expected output:** For each bug found, provide:
1. File and line number
2. Root cause explanation
3. The code fix applied

**Success criteria:**
- `go build ./...` succeeds with no errors
- All existing tests pass (`go test ./...`)
- The specific [SYMPTOM AREA] behavior is resolved
- Genre-specific tile types are handled consistently across all subsystems
```

---

## Usage

1. Copy the prompt above.
2. Replace `[SYMPTOM AREA]` with the specific issue, e.g.:
   - "rendering (walls appear invisible or transparent)"
   - "player navigation (cannot move or gets stuck)"
   - "door interaction (doors don't open or block forever)"
   - "automap display (tiles shown in wrong colors)"
   - "enemy visibility (enemies don't appear or are inside walls)"
   - "HUD elements (health/ammo/minimap not drawing)"
   - "menu navigation (settings don't apply, screens blank)"
3. Paste into the AI assistant and let it execute autonomously.

## Common Bug Patterns

| Pattern | Symptom | Typical Location |
|---|---|---|
| `tile > 0` wall check | Floors rendered as walls, player trapped | `pkg/raycaster/raycaster.go` |
| Missing genre tile in `switch` | Feature works in fantasy, breaks in other genres | `main.go`, `pkg/render/render.go` |
| `== bsp.TileFloor` adjacency | Secrets/enemies misplaced under genre | `main.go` helper functions |
| Hardcoded spawn coords | Entities inside walls | `main.go` spawn functions |
| Palette missing key | Transparent/invisible walls | `pkg/render/render.go` |
| Interaction distance > 1.0 | Can't interact with adjacent tile | `main.go` interaction code |
| Door walkable before open | Player phases through closed doors | `main.go` `isWalkableTile()` |
| Zero direction vector | No movement despite input | `pkg/camera/camera.go` init |

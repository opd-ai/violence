# Implementation Plan: v3.0 — Visual Polish: Textures, Lighting, Particles, Indoor Weather

## Phase Overview
- **Objective**: Achieve genre-distinct visual atmosphere where each of the five genres feels visually unique through procedural textures, dynamic lighting, particle effects, and post-processing.
- **Source Document**: ROADMAP.md § v3.0 — Visual Polish
- **Prerequisites**: v2.0 complete (Weapons, FPS AI, Keycards, All 5 Genres); existing stubs in `pkg/texture`, `pkg/lighting`, `pkg/particle`, `pkg/render/postprocess.go`
- **Estimated Scope**: Medium (core visual systems exist; requires integration, expansion, and polish)

## Implementation Steps

### 1. Extend Sector-Based Lighting (`pkg/lighting`)
1. Implement `SectorLightMap` struct with per-tile light levels
   - **Deliverable**: `SectorLightMap` stores `[][]float64` grid of ambient + accumulated point light contributions per tile
   - **Dependencies**: None

2. Implement point light accumulation algorithm
   - **Deliverable**: `AddPointLight(x, y, radius, intensity, r, g, b)` contributes inverse-square falloff light to affected tiles; `Calculate()` precomputes all tile values
   - **Dependencies**: Step 1

3. Implement flashlight cone light source
   - **Deliverable**: `AddFlashlight(x, y, dirX, dirY, coneAngle, range, intensity)` adds cone-shaped forward light following player direction
   - **Dependencies**: Step 2

4. Wire genre-specific base ambient levels via `SetGenre()`
   - **Deliverable**: Fantasy=0.3, Scifi=0.5, Horror=0.15, Cyberpunk=0.25, Postapoc=0.35 base ambient values; `SetGenre()` selects preset
   - **Dependencies**: Step 1

5. Integrate `SectorLightMap` with render pipeline
   - **Deliverable**: `Renderer.SetLightMap(lightMap)` connects to existing `getLightMultiplier()` in `pkg/render/render.go`; walls/floors/ceilings modulated by lighting
   - **Dependencies**: Steps 1-4, `pkg/render`

6. Add unit tests for lighting calculations
   - **Deliverable**: Tests for point light falloff accuracy, flashlight cone bounds, ambient level per genre, integration with renderer
   - **Dependencies**: Steps 1-5

### 2. Wall Texture Sampling in Raycaster (`pkg/raycaster`, `pkg/render`)
7. Implement wall texture coordinate calculation
   - **Deliverable**: `RayHit` includes `TextureX float64` (0.0-1.0) representing hit position along wall surface
   - **Dependencies**: `pkg/raycaster` (existing)

8. Implement wall texture sampling in renderer
   - **Deliverable**: `renderWall()` samples from `atlas.Get("wall_N")` using `hit.TextureX` and vertical scanline position; falls back to palette color if no texture
   - **Dependencies**: Step 7, `pkg/texture`

9. Generate genre-specific wall texture variants
   - **Deliverable**: `Atlas.GenerateWallSet(genreID string)` creates wall_1 through wall_4 procedural textures per genre (stone/hull/plaster/concrete/rust); `SetGenre()` triggers regeneration
   - **Dependencies**: `pkg/texture` (existing)

10. Add unit tests for textured wall rendering
    - **Deliverable**: Tests for texture coordinate calculation, texture sampling at various hit positions, genre texture switching
    - **Dependencies**: Steps 7-9

### 3. Indoor Weather Particle Emitters (`pkg/particle`)
11. Implement `WeatherEmitter` for continuous ambient particles
    - **Deliverable**: `WeatherEmitter` continuously spawns particles at random ceiling/wall positions based on genre parameters (rate, velocity, color, lifetime)
    - **Dependencies**: `pkg/particle` (existing)

12. Implement fantasy weather: water drips, torch smoke, fog wisps
    - **Deliverable**: `SetGenre("fantasy")` configures: water drips (blue, downward, slow), torch smoke (orange→gray, upward, slow), fog wisps (white, lateral drift)
    - **Dependencies**: Step 11

13. Implement scifi weather: steam vents, coolant spray, hull breach sparks
    - **Deliverable**: `SetGenre("scifi")` configures: vent steam (white, upward bursts), coolant spray (cyan, lateral), sparks (yellow/orange, fast random)
    - **Dependencies**: Step 11

14. Implement horror weather: flickering particles, blood drips, mold spores
    - **Deliverable**: `SetGenre("horror")` configures: flicker (white→black, stationary, short life), blood drips (red, downward), spores (green, slow drift)
    - **Dependencies**: Step 11

15. Implement cyberpunk weather: holographic static, neon haze, electrical crackle
    - **Deliverable**: `SetGenre("cyberpunk")` configures: static (magenta/cyan, rapid flicker), neon haze (pink/blue, slow drift), crackle (white, fast zigzag)
    - **Dependencies**: Step 11

16. Implement postapoc weather: dust particles, radiation shimmer, debris chunks
    - **Deliverable**: `SetGenre("postapoc")` configures: dust (tan, slow lateral), shimmer (distortion effect), debris (brown, gravity-affected)
    - **Dependencies**: Step 11

17. Add unit tests for weather emitters
    - **Deliverable**: Tests for spawn rate per genre, particle velocity bounds, color correctness, genre switching
    - **Dependencies**: Steps 11-16

### 4. Animated Texture Integration (`pkg/texture`, `pkg/render`)
18. Wire animated texture playback to render pipeline
    - **Deliverable**: `Renderer.Tick` increments frame counter; `renderWall()` uses `atlas.GetAnimatedFrame(name, tick)` for animated wall segments
    - **Dependencies**: `pkg/texture/animated.go` (existing)

19. Generate genre-specific animated textures
    - **Deliverable**: Fantasy="flicker_torch", Scifi="blink_panel", Horror="drip_water", Cyberpunk="neon_pulse", Postapoc="radiation_glow"; `Atlas.GenerateGenreAnimations()` creates all five
    - **Dependencies**: `pkg/texture/animated.go` (existing)

20. Add unit tests for animated texture playback
    - **Deliverable**: Tests for frame progression, loop behavior, genre animation selection
    - **Dependencies**: Steps 18-19

### 5. Post-Processing Tuning Pass (`pkg/render`)
21. Add horror static burst effect
    - **Deliverable**: `ApplyStaticBurst()` occasionally (1% chance per frame) overlays noise on screen for 2-3 frames; triggered in horror genre preset
    - **Dependencies**: `pkg/render/postprocess.go` (existing)

22. Add postapoc film scratch effect
    - **Deliverable**: `ApplyFilmScratches()` draws vertical scratch lines at random positions with configurable density; added to postapoc preset
    - **Dependencies**: `pkg/render/postprocess.go` (existing)

23. Tune genre preset parameters for visual balance
    - **Deliverable**: Playtest each genre and adjust vignette/grain/bloom/aberration values for optimal atmosphere without obscuring gameplay
    - **Dependencies**: Steps 21-22

24. Add unit tests for new post-processing effects
    - **Deliverable**: Tests for static burst triggering, scratch line rendering, preset parameter ranges
    - **Dependencies**: Steps 21-23

### 6. Audio-Visual Synchronization (`pkg/audio`)
25. Expose room size to reverb calculator from BSP
    - **Deliverable**: `ReverbCalculator.SetRoomFromBSP(room *bsp.Room)` extracts width/height from room bounds and recalculates parameters
    - **Dependencies**: `pkg/audio/reverb.go` (existing), `pkg/bsp`

26. Implement per-room reverb transition
    - **Deliverable**: `Engine.UpdateReverb(currentRoomX, currentRoomY, levelBSP)` detects room changes and smoothly transitions reverb parameters over 0.5 seconds
    - **Dependencies**: Step 25

27. Add unit tests for spatial reverb
    - **Deliverable**: Tests for room size to decay mapping, transition smoothing, BSP integration
    - **Dependencies**: Steps 25-26

### 7. Integration and Wiring
28. Wire all v3.0 systems into game loop
    - **Deliverable**: `main.go` initializes `SectorLightMap`, `WeatherEmitter`, animated texture ticker; `Game.Update()` calls lighting recalculation, particle update, reverb update; `Game.Draw()` passes light map and particles to renderer
    - **Dependencies**: All above steps

29. Implement `SetGenre()` cascade for all v3.0 systems
    - **Deliverable**: Top-level `SetGenre(genreID)` propagates to `SectorLightMap`, `Atlas`, `WeatherEmitter`, `PostProcessor`, `AmbientSoundscape`, `ReverbCalculator`
    - **Dependencies**: Step 28

30. End-to-end visual playtest for each genre
    - **Deliverable**: Each genre visually distinct: Fantasy (warm, flickering torches, water drips), Scifi (cold blue, scanlines, steam), Horror (dark, static bursts, mold), Cyberpunk (neon bloom, electric crackle), Postapoc (orange dust, scratches)
    - **Dependencies**: Step 29

## Technical Specifications
- **Lighting algorithm**: Per-tile float64 light level (0.0=dark, 1.0=full); point lights use inverse-square falloff clamped at radius; flashlight uses dot-product cone test
- **Texture sampling**: Nearest-neighbor sampling with wrapping; wall texture coordinates computed as `hit.WallX - floor(hit.WallX)`
- **Particle pool**: Fixed 1024-particle pool with ring-buffer allocation; spatial culling at ±50 tiles from camera
- **Reverb parameters**: Decay 0.1-0.8, wet mix 0.0-0.5, dry mix 0.8-1.0; derived from room area (10x10=minimal, 50x50=maximum)
- **Post-processing order**: Color Grade → Vignette → Film Grain → Scanlines → Chromatic Aberration → Bloom → Static Burst → Film Scratches

## Validation Criteria
- [ ] Each genre produces visually distinct atmosphere (verifiable by screenshot comparison)
- [ ] Point lights illuminate surrounding tiles with visible falloff gradient
- [ ] Flashlight cone follows player direction and illuminates forward area
- [ ] Wall textures display correctly with perspective-correct vertical scaling
- [ ] Animated textures cycle smoothly at configured FPS
- [ ] Weather particles spawn continuously without pool exhaustion
- [ ] Genre switch changes all visual systems within 1 frame
- [ ] Reverb changes audibly when moving between small and large rooms
- [ ] Post-processing effects do not drop frame rate below 30 FPS at 320×200
- [ ] Test coverage remains ≥82% after all additions

## Known Gaps
- **Wall texture mipmap filtering**: Current implementation uses nearest-neighbor sampling which may cause aliasing at distance; bilinear filtering could be added in future optimization pass if visual artifacts are noticeable
- **Particle depth sorting**: Particles render in spawn order rather than depth-sorted; may cause visual artifacts when particles overlap at different distances; low priority given particle sizes and alpha blending

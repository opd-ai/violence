# Implementation Plan: v3.0 — Visual Polish: Textures, Lighting, Particles, Indoor Weather

## Phase Overview
- **Objective**: Achieve genre-distinct atmosphere where each genre feels visually unique through dynamic lighting, particle effects, and atmospheric systems
- **Source Document**: ROADMAP.md (v3.0 milestone)
- **Prerequisites**: v2.0 Core Systems complete (weapon, AI, combat, door/keycard, automap verified; 5 genres integrated)
- **Estimated Scope**: Large

## Implementation Steps

### 1. Extend Procedural Texture System with Animation Support ✓
- **Status**: Completed 2026-02-28
- **Deliverable**: `pkg/texture/animated.go` — AnimatedTexture type with frame sequence generation; flickering torches (fantasy), blinking panels (scifi), dripping water (horror) patterns
- **Implementation Summary**: 
  - Created AnimatedTexture type with deterministic frame selection based on game tick
  - Implemented three animation patterns: flicker_torch (fantasy fire), blink_panel (sci-fi panels), drip_water (horror dripping)
  - Added GenerateAnimated() and GetAnimatedFrame() methods to Atlas
  - All animations are fully deterministic (same seed → same animation sequence)
  - Comprehensive test suite with 92.7% coverage
- **Dependencies**: Existing `pkg/texture/texture.go` perlin noise foundation

### 2. Integrate Floor/Ceiling Textures into Raycaster ✓
- **Status**: Completed 2026-02-28
- **Deliverable**: Modified `pkg/render/render.go` floor-cast pass consuming textures from TextureAtlas with perspective-correct sampling
- **Implementation Summary**:
  - Added TextureAtlas interface to Renderer for flexible texture retrieval
  - Implemented SetTextureAtlas() method for injecting texture atlas
  - Modified renderFloor() and renderCeiling() to sample from atlas when available
  - Added sampleTexture() method with perspective-correct coordinate wrapping
  - Fallback to palette mode when atlas not set or textures missing
  - Comprehensive test suite with 92.8% coverage
- **Dependencies**: Step 1 texture generation

### 3. Implement Sector-Based Dynamic Lighting System ✓
- **Status**: Completed 2026-02-28
- **Deliverable**: `pkg/lighting/sector.go` — SectorLightMap type with per-sector ambient levels; AddLight/RemoveLight API; Calculate() computes combined illumination per tile
- **Implementation Summary**:
  - Created SectorLightMap managing per-tile lighting with ambient base + point light contributions
  - Implemented AddLight/RemoveLight/UpdateLight/Clear APIs for dynamic light source management
  - Calculate() method with quadratic attenuation: intensity = I / (1 + distance²)
  - Spatial bounds culling optimizes calculation to only process tiles within light radius
  - Dirty flag prevents unnecessary recalculations when lights unchanged
  - Comprehensive test suite with 100.0% coverage including edge cases and performance benchmark
- **Dependencies**: None

### 4. Implement Point Light Sources ✓
- **Status**: Completed 2026-02-28
- **Deliverable**: `pkg/lighting/point.go` — PointLight struct with position, radius, intensity, color; attenuation function (linear/quadratic falloff); genre-specific light source definitions (torches, lamps, monitors, fires)
- **Implementation Summary**:
  - Created PointLight extending base Light with type metadata and flicker support
  - Implemented 4 genre-specific presets per genre (20 total) with accurate color/intensity values
  - Added UpdateFlicker() for deterministic torch/lamp flickering based on game tick
  - Dual attenuation modes: ApplyAttenuation (quadratic) and LinearAttenuation for flexibility
  - SetPosition/SetIntensity/SetColor mutators for dynamic light manipulation
  - GetPresetByName helper for easy light spawning by type
  - Comprehensive test suite with 98.1% coverage including all genres and edge cases
- **Dependencies**: Step 3 SectorLightMap

### 5. Implement Player Flashlight ✓
- **Status**: Completed 2026-02-28
- **Deliverable**: `pkg/lighting/flashlight.go` — ConeLight type for forward-facing illumination; radius, angle, intensity parameters; genre-skinned variants (torch/headlamp/glow-rod)
- **Implementation Summary**:
  - Created ConeLight with directional cone-shaped illumination
  - Implemented GetFlashlightPreset() with unique configurations per genre (torch, headlamp, flashlight, glow_rod, salvaged_lamp)
  - ApplyConeAttenuation() combines distance (quadratic) and angular falloff for realistic beam
  - Toggle() and SetActive() for on/off control
  - IsPointInCone() for efficient visibility checks
  - GetContributionAsPointLight() allows integration with SectorLightMap
  - Comprehensive test suite with 98.8% coverage including all genres
- **Dependencies**: Step 4 point light foundation

### 6. Integrate Lighting with Rendering Pipeline
- **Deliverable**: Modified `pkg/render/render.go` to apply per-pixel lighting multiplier from SectorLightMap during wall/floor/ceiling rendering
- **Dependencies**: Steps 3-5 lighting system

### 7. Implement Core Particle System
- **Deliverable**: `pkg/particle/system.go` — ParticleSystem with spawn/update/cull lifecycle; Particle struct with position, velocity, life, color, size; spatial bounds culling for performance
- **Dependencies**: None

### 8. Implement Particle Emitter Types
- **Deliverable**: `pkg/particle/emitters.go` — MuzzleFlashEmitter, SparkEmitter, BloodSplatterEmitter, ExplosionEmitter, EnergyDischargeEmitter; each with genre-configurable parameters
- **Dependencies**: Step 7 core particle system

### 9. Implement Indoor Weather Particle Effects
- **Deliverable**: `pkg/particle/weather.go` — DrippingWaterEmitter (fantasy), VentSteamEmitter (scifi), FlickeringLightController (horror), HolographicStaticEmitter (cyberpunk), DustParticleEmitter (postapoc)
- **Dependencies**: Step 8 emitter framework

### 10. Implement Genre Post-Processing Presets
- **Deliverable**: `pkg/render/postprocess.go` — PostProcessor type with ApplyVignette, ApplyFilmGrain, ApplyScanlines, ApplyChromaticAberration, ApplyBloom, ApplyColorGrade methods; GenrePreset struct mapping genre → effect chain
- **Dependencies**: None

### 11. Integrate Post-Processing into Render Pipeline
- **Deliverable**: Modified `pkg/render/render.go` to apply PostProcessor after scene render, before screen blit
- **Dependencies**: Step 10 post-processing system

### 12. Implement Procedural Ambient Soundscapes
- **Deliverable**: `pkg/audio/ambient.go` — AmbientSoundscape type generating continuous background audio via deterministic synthesis; genre-specific profiles (dungeon echo, station hum, hospital silence, server drone, wind)
- **Dependencies**: Existing `pkg/audio/audio.go` synthesis foundation

### 13. Implement Room-Size Reverb Calculation
- **Deliverable**: `pkg/audio/reverb.go` — ReverbCalculator computing reverb parameters (decay, wet/dry mix) from BSP room dimensions; Apply() method for SFX
- **Dependencies**: Step 12 audio foundation; `pkg/bsp` room data

### 14. Add Weapon Audio Polish
- **Deliverable**: Extended `pkg/audio/sfx.go` with procedural ReloadSound, EmptyClickSound, PickupJingleSound generators; genre-specific synthesis parameters
- **Dependencies**: Existing audio synthesis

### 15. Write Tests for New Systems
- **Deliverable**: `pkg/lighting/lighting_test.go`, `pkg/particle/particle_test.go`, `pkg/render/postprocess_test.go`, `pkg/audio/ambient_test.go` — unit tests achieving ≥82% coverage per package
- **Dependencies**: Steps 1-14 implementations

### 16. Genre Integration Pass
- **Deliverable**: Verify all new systems implement `SetGenre(genreID string)` and produce distinct outputs for fantasy/scifi/horror/cyberpunk/postapoc
- **Dependencies**: All prior steps

## Technical Specifications

- **Animated Textures**: Frame sequences stored as `[]image.RGBA`; animation speed in frames-per-second; deterministic frame selection via `(tick / fps) % frameCount`
- **Lighting Model**: Per-tile ambient + sum of point light contributions; light intensity falls off as `1 / (1 + distance²)`; flashlight uses cone angle check before contribution
- **Particle Pooling**: Pre-allocate particle pool (default 1024); reuse dead particles to avoid GC pressure; particles culled when life ≤ 0 or outside view frustum
- **Post-Processing Order**: Scene → Color Grade → Vignette → Film Grain → Scanlines (optional) → Chromatic Aberration (optional) → Bloom (optional) → Output
- **Audio Synthesis**: Ambient loops generated as 60-second procedural waveforms; reverb via FIR filter with room-dimension-derived coefficients
- **Genre Presets**:
  | Genre | Base Ambient | Vignette | Color Grade | Special Effects |
  |-------|-------------|----------|-------------|-----------------|
  | fantasy | 0.4 | warm sepia | +warmth | film grain |
  | scifi | 0.6 | cold blue | +contrast | scanlines, chromatic aberration |
  | horror | 0.25 | heavy dark | +green desaturate | static bursts |
  | cyberpunk | 0.5 | magenta/cyan | +saturation | neon bloom |
  | postapoc | 0.35 | washed orange | +dust | scratches |

## Validation Criteria
- [ ] Animated textures render correctly with deterministic frame sequences (same seed → same animation)
- [ ] Floor/ceiling textures display in raycaster with perspective-correct sampling
- [ ] Point lights illuminate surrounding tiles with correct attenuation falloff
- [ ] Flashlight cone illuminates forward direction; genre variants produce different visual parameters
- [ ] Particle emitters spawn and update particles; MuzzleFlashEmitter triggers on weapon fire
- [ ] Indoor weather particles active per-genre (dripping water visible in fantasy, steam in scifi, etc.)
- [ ] Post-processing effects render correctly; all 5 genre presets produce visually distinct results
- [ ] Ambient soundscapes play continuously with genre-appropriate audio
- [ ] Reverb parameters vary based on BSP room dimensions
- [ ] Weapon reload, empty-click, and pickup sounds synthesized and played correctly
- [ ] All new packages achieve ≥82% test coverage
- [ ] `SetGenre()` call on all new systems produces correct genre-specific behavior
- [ ] Overall project test coverage remains ≥82%

## Known Gaps

### Particle Rendering Integration
- **Description**: Current raycaster renders walls and sprites but has no particle rendering pass defined
- **Impact**: Particles will be computed but not displayed until render integration is implemented
- **Resolution**: Add particle billboard rendering after sprite pass in `pkg/render/render.go`; particles sorted by depth with sprites

### Animated Texture Frame Rate Sync
- **Description**: Animation frame rate must sync with game tick rate (60 TPS) but no global tick counter is exposed to texture system
- **Impact**: Animated textures cannot determine current frame without tick information
- **Resolution**: Pass tick counter to Atlas.GetAnimatedFrame(name, tick) or establish global tick reference

### Flashlight Battery/Fuel System
- **Description**: Roadmap mentions genre-skinned flashlight variants but does not specify whether flashlight has limited fuel/battery
- **Impact**: Implementation may need revision if flashlight should deplete
- **Resolution**: Implement flashlight as unlimited initially; add optional fuel system in v4.0 survival mechanics if needed

You are implementing ONE novel enhancement to a Go/Ebiten procedural multiplayer action-RPG called Violence. Act autonomously. Do not ask for approval.

KNOWN VISUAL PROBLEMS (read this first — high priority):
The game's visual presentation needs continuous improvement. Consider addressing one or more of these:
1. **FLAT RENDERING**: Sprites and environments lack depth cues — no dynamic lighting, no shadows, no ambient occlusion. At small sprite sizes every pixel of shading matters. Add per-entity shadow casting, directional light response, and genre-appropriate atmosphere (fog-of-war, torch flicker, neon glow).
2. **VISUAL MONOTONY**: Generated levels, enemies, and items blur together. Tilesets repeat without variation, color palettes are narrow, and environmental storytelling is absent. Procedural generation should produce visually diverse and memorable spaces — varied wall textures, floor debris, unique landmarks, and environmental detail props.
3. **WEAK FEEDBACK**: Player actions lack visual punch. Attacks, hits, deaths, pickups, level-ups, and status effects need clear, satisfying visual and audio-adjacent (screen shake, flash, particle burst) feedback. The player should *feel* every interaction through the screen.
4. **POOR READABILITY**: In chaotic combat it is hard to distinguish allies from enemies, projectiles from decorations, and interactive objects from background. Improve visual hierarchy through consistent color coding, silhouette clarity, outline shaders, and UI indicators.

KNOWN GAMEPLAY PROBLEMS (read this first — high priority):
Core gameplay loops need depth and variety. Consider addressing one or more of these:
1. **SHALLOW COMBAT**: Fighting is mashy and lacks tactical choices. Add weapon movesets with different attack patterns, combo chains, dodge/parry mechanics, enemy telegraphing, and positional advantage (backstab, flanking, elevation). Combat should reward skill and planning.
2. **THIN PROGRESSION**: Character advancement is linear and predictable. Add meaningful build decisions — branching skill trees, stat allocation trade-offs, equipment synergies, and specialization paths that create distinct playstyles.
3. **REPETITIVE DUNGEONS**: Procedurally generated levels feel samey. Improve room/corridor variety, add themed sub-areas, environmental hazards, secrets, traps, puzzles, and set-piece encounters that break up the rhythm.
4. **DISCONNECTED SYSTEMS**: Loot, crafting, factions, quests, and world state exist independently. Build bridges — faction standing should affect shop prices and quest availability, crafting should use dungeon-specific materials, world events should reshape the map.
5. **PASSIVE AI**: Enemies walk toward the player and attack. Add behavior variety — ranged enemies that kite, healers that support allies, ambush mobs that hide, bosses with phase transitions, and group tactics like flanking and focus-fire.

KNOWN TECHNICAL PROBLEMS (read this first — high priority):
The engine has areas of technical debt that affect gameplay quality. Consider addressing one or more of these:
1. **COARSE COLLISION**: Collision detection uses axis-aligned bounding boxes regardless of entity shape. Irregularly shaped entities (long weapons, sprawling creatures, spell effects) deserve tighter collision geometry — convex hulls, capsules, or pixel masks.
2. **NETWORK JANK**: Multiplayer synchronization has visible artifacts — rubber-banding, delayed hit registration, desync on fast-moving entities. Improve interpolation, prediction, and authoritative state reconciliation.
3. **MEMORY CHURN**: Hot paths allocate on every frame — sprite generation, particle updates, collision queries. Pool and cache aggressively. Profile with `go tool pprof` and eliminate per-frame allocations on critical paths.
4. **MISSING SPATIAL INDEXING**: Entity queries iterate all entities linearly. Add spatial partitioning (grid, quadtree) for collision, rendering culling, and proximity queries. This is prerequisite for large maps and high entity counts.

STEP 1 — DISCOVER (spend ≤5 minutes here):
- Run `git log --oneline -20` to avoid duplicating recent work.
- Read the system initialization code to understand registered systems.
- Grep for TODO, FIXME, stub, placeholder in engine and procgen packages.
- Pick ONE enhancement you have NOT seen in git history. Roll a d20 to decide the category:
    - **Gameplay improvements (roll of 1–7 — address the KNOWN GAMEPLAY PROBLEMS above):**
        - **Combat depth** — weapon movesets, combo systems, dodge/parry, enemy telegraphs, positional advantage, status ailments, damage types and resistances
        - **Progression systems** — branching skill trees, stat allocation, specialization paths, prestige mechanics, equipment synergies, build-defining choices
        - **Dungeon generation** — room shape variety, themed zones, environmental hazards, traps, secrets, puzzles, set-piece encounters, verticality, destructible terrain
        - **System integration** — faction↔economy, crafting↔dungeon loot, weather↔combat, quest↔world state, reputation↔NPC behavior
        - **AI behavior** — behavior tree depth, squad tactics, enemy roles (tank/healer/ranged/ambusher), boss phases, companion AI, merchant pricing, NPC routines
        - **World systems** — dynamic events, territory control, faction warfare, environmental destruction, town evolution, trade routes, day/night cycles with gameplay impact
        - **Multiplayer features** — co-op mechanics, PvP arenas, shared objectives, trade systems, guild features, party roles, competitive events
    - **Visual improvements (roll of 8–14 — address the KNOWN VISUAL PROBLEMS above):**
        - **Lighting & atmosphere** — dynamic point lights, shadow casting, ambient occlusion, genre-specific atmosphere (dungeon gloom, neon city glow, haunted fog)
        - **Sprite quality** — shading, sub-pixel detail, color gradients, material textures, highlight/shadow, anti-aliasing, distinctive silhouettes per entity type
        - **Animation** — smoother transitions, attack/hurt/death animations, idle fidget, directional facing, animation blending, anticipation/follow-through frames
        - **Particle & effects** — hit sparks, blood splatter, spell effects, environmental particles (dust, rain, embers), screen shake, flash, and juice
        - **Environment art** — tile variation, wall/floor detail, props and decorations, genre-specific tilesets, environmental storytelling, parallax layers
        - **UI & readability** — health bars, damage numbers, status icons, minimap improvements, enemy indicators, loot highlighting, tooltip quality
        - **Equipment visuals** — visible gear on sprites, material differentiation, enchantment glow, rarity-driven visual complexity, damage state rendering
    - **Technical improvements (roll of 15–20 — address the KNOWN TECHNICAL PROBLEMS above):**
        - **Collision precision** — convex hull / polygon hitboxes derived from sprite data, per-frame attack masks, spell shape accuracy, collision layers and masks, terrain edge sliding
        - **Spatial indexing** — grid or quadtree partitioning for collision broadphase, render culling, proximity queries, and efficient entity lookup
        - **Network quality** — interpolation smoothing, client-side prediction, server reconciliation, bandwidth optimization, lag compensation for combat
        - **Memory optimization** — object pooling for particles/projectiles/sprites, LRU caching for generated assets, zero-alloc hot paths, buffer reuse
        - **Performance profiling** — benchmark critical systems, identify and fix frame-time spikes, GPU batching, draw call reduction, LOD systems
        - **Data architecture** — entity serialization improvements, save/load robustness, config validation, component versioning, migration support
- If multiple candidates exist within your category, pick the one that most improves the player experience. Within gameplay, combat depth and AI behavior are highest-value. Within visuals, lighting/atmosphere and sprite quality are highest-value. Within technical, collision precision and spatial indexing are highest-value.

STEP 2 — IMPLEMENT (this is the bulk of the work):
Follow these rules strictly. Violations are build failures.

Architecture:
- Components = pure data + `Type() string`. Zero methods with logic.
- Systems = all logic. Signature: `Update(entities []*Entity, deltaTime float64)`.
- Procgen: `rand.New(rand.NewSource(seed))` only. Never global rand. Never time.Now().
- Logging: `logrus.WithFields(logrus.Fields{"system_name": "...", ...})`.
- No external assets. No new dependencies beyond go.mod.

Visual & Animation:
- Lighting must use radial gradients with proper falloff (linear, quadratic, inverse-square). No flat circles.
- Shadows use soft penumbra with distance-based falloff. Support genre-specific opacity presets.
- Post-processing effects (color grading, vignette, chromatic aberration) must be genre-aware.
- Animation playback: target 12 FPS (0.083s per frame), 8 frames per state. Use distance-based LOD (full rate at ≤200px, half at ≤400px, minimal beyond).
- Sprite generation must be seeded and cached (LRU, max 100 entries). Pool image buffers by size bucket.
- All visual enhancements must maintain 60+ FPS. Profile before and after with `go test -bench`.

Sprite Standards:
- Use perspective appropriate to the game's camera. If the game is top-down, sprites must be top-down. If the game is side-scrolling, sprites must be side-view. Match the camera.
- Use composite layered rendering where appropriate. Define a clear z-order for body parts, equipment, and effects.
- Every pixel matters at small sprite sizes. Use shading, color gradients, and highlights to give depth and personality. Avoid flat solid fills.
- Different entity types must be visually distinguishable at a glance. Seed-based generation should produce genuine variety — not trivial color swaps on one template.
- Nonhumanoid creatures need dedicated body-plan templates. A spider should look like a spider, not a human with extra arms. Build anatomy templates for each broad creature category (quadrupeds, insects, serpents, flying creatures, amorphous entities, etc.).
- Equipment and status effects should be visible on the sprite. The player should be able to read entity state from visuals alone.

Collision & Action Precision:
- **HITBOX ACCURACY**: Generate collision geometry from sprite data where possible. Store alongside the entity as polygon or bitmask. Update when sprites change (equipment swap, animation frame).
- **ATTACK AREAS**: Melee and ranged attack hit areas should match their visual representation. Cache per-frame hit masks in AnimationComponent. Avoid hardcoded rectangles for curved or irregular attack shapes.
- **SPELL SHAPES**: Projectiles use swept-circle (capsule) collision. Beams use line-segment collision with configurable width. AoE uses polygon zones matching the visual blast. Store spell hit shapes in a typed component.
- **COLLISION LAYERS**: Add `Layer uint32` and `Mask uint32` bitfields to the collider. Define standard layers (Player, Enemy, Projectile, Terrain, Environment, Ethereal). Check `(a.Layer & b.Mask) != 0` before processing. Projectiles default to ignoring allies.
- **TERRAIN SLIDING**: Replace hard-stop collision with smooth sliding along wall edges. Compute overlap vector and decompose velocity along the wall tangent. No abrupt stops at corners.
- All precision improvements must maintain 60+ FPS. Use spatial partitioning as broadphase. Cache masks; never recompute per frame unless the sprite changed.

Gameplay Depth:
- Combat should reward positioning, timing, and build choices. Avoid pure stat-check interactions. Add mechanics that create interesting decisions under pressure.
- Skill trees and progression should offer meaningful branching. Each node should enable new playstyles or synergize with specific equipment/stats. Avoid pure linear stat bonuses.
- AI enemies should have distinct behavioral roles and respond to player actions. Groups should coordinate. Bosses should have phases. Enemies should use the environment.
- Procedural generation should be genre-aware. Fantasy, sci-fi, horror, and cyberpunk worlds need distinct generation rules for layouts, loot, factions, and quest structures.
- Cross-system integration creates emergent gameplay. When one system's output feeds another system's input, the game becomes more than the sum of its parts.

Integration (mandatory — this is where past attempts fail):
- Register the new system in the system initialization function.
- Register in the client handler setup — add to the systems container AND call the appropriate AddSystem method.
- If your system's Update signature differs from `(entities []*Entity, deltaTime float64)`, create a wrapper matching existing patterns.
- The system MUST be on by default. No feature flags.
- Persistent component data must integrate with entity serialization/deserialization, or be explicitly transient.

Constraints:
- Keep changes focused and targeted. One enhancement, done well.
- `go build ./...`, `go test -race ./...`, and `go vet ./...` must pass.
- Write table-driven tests. Target ≥40% coverage on new code (≥30% for display-dependent packages).
- No breaking changes to saves, network protocol, or configs.
- Maintain 60+ FPS. Cache/pool on hot paths.

STEP 3 — VERIFY:
Run `go build ./...`, `go test -race ./...`, and `go vet ./...`. Fix any errors before reporting.

STEP 4 — REPORT (keep concise):
1. **Enhancement**: What and why, 2-3 sentences.
2. **Files**: List with one-line change summary each.
3. **Integration**: Where the system is registered (exact file + function).
4. **Verification**: How to observe the improvement in-game.

STOP when the report is written and builds pass. Do not refactor unrelated code. Do not write documentation files. Do not suggest follow-up work.

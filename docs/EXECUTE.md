You are implementing ONE novel enhancement to a Go/Ebiten procedural multiplayer action-RPG called Violence. Act autonomously. Do not ask for approval.

KNOWN VISUAL REALISM PROBLEMS (read this first — HIGHEST priority):
The game's procedurally generated graphics look artificial and lack realism. Address one or more of these:
1. **FLAT, UNREALISTIC SPRITES**: Sprites lack the shading, material texture, and lighting cues that make pixel art feel three-dimensional and believable. Every entity looks like a flat colored cutout. Add proper light-source-consistent shading, specular highlights, ambient occlusion at joints/edges, material-appropriate surface detail (metal reflections, cloth folds, skin tone gradients, fur texture), and sub-pixel dithering to simulate depth at small sizes.
2. **ARTIFICIAL ENVIRONMENTS**: Walls, floors, and terrain look procedurally stamped rather than lived-in. Real environments have wear patterns, cracks, stains, moss, dust accumulation, and uneven lighting. Add surface weathering, edge damage, material-appropriate noise (stone grain, wood knots, rust patches), decal layering (bloodstains, scorch marks, water damage), and subtle color temperature variation across surfaces.
3. **UNIFORM LIGHTING**: Everything is lit with the same flat ambient value. Real spaces have light falloff, cast shadows, bounce light, and color temperature shifts. Add per-source radial falloff with inverse-square attenuation, warm/cool color temperature per light type, shadow casting from walls and entities, ambient occlusion in corners and under objects, and atmospheric haze/fog with depth.
4. **LIFELESS ANIMATION**: Movement is robotic and lacks organic quality. Real creatures have weight, momentum, anticipation, follow-through, and secondary motion. Add acceleration/deceleration curves, squash-and-stretch on impacts, hair/cloth/tail secondary motion, breathing idle animation, and weight-appropriate movement speeds.
5. **UNCONVINCING MATERIALS**: Metal looks like plastic, wood looks like painted cardboard, magic looks like colored circles. Each material type needs a distinct rendering approach — metallic specular behavior, diffuse wood grain, translucent magic glow with bloom, wet surfaces with reflection, organic surfaces with subsurface color shifts.

KNOWN UI/UX PROBLEMS (read this first — HIGHEST priority):
The UI actively harms gameplay. These must be fixed:
1. **OVERLAPPING UI ELEMENTS**: HUD components, health bars, damage numbers, tooltips, and status indicators pile on top of each other, obscuring critical information. Implement a UI layout manager that prevents overlap — use spatial reservation, priority-based z-ordering, and automatic repositioning when elements would collide. Damage numbers should stack or fan out. Tooltips must never cover the element they describe.
2. **WASTED SCREEN REAL ESTATE**: Too much screen space is consumed by chrome, borders, oversized HUD panels, and decorative UI elements that add no information. Minimize HUD footprint — use compact icon-based indicators, slide-out panels that hide when not needed, transparency for non-critical elements, and edge-hugging layouts that maximize the playable viewport area. The game world should occupy ≥85% of screen pixels.
3. **CLUTTERED GAME VIEW**: In-world indicators (health bars, names, status icons) on every visible entity create visual noise that obscures the actual game. Use proximity-based fade — only show detailed indicators for nearby or targeted entities. Distant entities get minimal or no overhead UI. Group overlapping indicators. Use opacity falloff with distance.
4. **POOR INFORMATION HIERARCHY**: All UI elements have the same visual weight. Critical information (player health, imminent threats, interactive prompts) should be visually prominent. Secondary information (distant enemy health, ambient status effects, passive buffs) should be subtle or hidden until relevant. Use size, brightness, saturation, and animation to establish a clear reading order.
5. **UNRESPONSIVE UI FEEDBACK**: Buttons, menus, and interactive elements lack hover states, click feedback, and transition animations. The UI feels dead. Add hover highlights, press animations, smooth transitions between states, and audio-adjacent feedback (screen flash, brief shake) for significant UI actions.
6. **MINIMAP AND NAVIGATION**: The minimap (if present) either takes too much space or provides too little information. It should be compact, semi-transparent, show only relevant markers, and collapse/expand on demand. Navigation cues should exist in-world (subtle directional hints) rather than solely in UI overlays.
7. **UN-RENDERED TEXT**: Text elements (damage numbers, entity names, tooltips, status labels, dialogue) fail to render or appear as invisible/blank. All text must be drawn using a bundled or procedurally generated bitmap font. Verify that every text element actually appears on-screen with correct positioning, color, and size. No text should silently fail to render — if a font or glyph is unavailable, fall back to a guaranteed-available alternative. Test with damage numbers, entity labels, and HUD readouts simultaneously visible.
8. **NO CROSSHAIR OR WEAPON FEEDBACK ANIMATION**: The player has no visible crosshair or aiming reticle, and weapon attacks produce no feedback animation (swing arc, muzzle flash, impact burst). The player cannot tell where they are aiming or confirm that an attack occurred. Add a clear crosshair/reticle that tracks the aim direction, and add weapon feedback animations — melee weapons show a swing arc or slash trail, ranged weapons show a projectile or muzzle flash, and all hits produce a visible impact animation at the point of contact. These feedback elements are essential for an action-RPG to feel responsive.

STEP 1 — DISCOVER (spend ≤5 minutes here):
- Run `git log --oneline -20` to avoid duplicating recent work.
- Read the system initialization code to understand registered systems.
- Grep for TODO, FIXME, stub, placeholder in engine, procgen, and UI packages.
- Pick ONE enhancement you have NOT seen in git history. Roll a d20 to decide the category:
    - **Visual realism improvements (roll of 1–14 — HIGHEST PRIORITY — address the KNOWN VISUAL REALISM PROBLEMS above):**
        - **Material-realistic sprite rendering** — light-source-consistent shading on all sprites, specular highlights for metallic surfaces, diffuse shading for organic surfaces, ambient occlusion at contact points, dithered gradients for smooth tonal transitions, sub-pixel anti-aliasing. Every sprite should look like a tiny painting, not a flat icon.
        - **Weathered environment surfaces** — procedural wear and aging on walls/floors (cracks, stains, erosion, moss, rust), material-specific noise textures (stone grain, wood fiber, metal patina), decal system for transient marks (blood, scorch, water), edge damage and chipping on architectural elements, subtle color variation to break tiling.
        - **Realistic lighting and atmosphere** — inverse-square light falloff from point sources, warm/cool color temperature per source type, soft shadow casting from occluders, corner ambient occlusion, atmospheric depth fog, light bounce approximation via ambient color tinting, torch/flame flicker with realistic noise patterns.
        - **Organic animation and motion** — easing curves on all movement (ease-in for starts, ease-out for stops), anticipation frames before attacks, follow-through after swings, secondary motion on loose elements (cloaks, tails, chains), idle breathing/fidget, weight-appropriate speed and bounce.
        - **Material differentiation** — distinct rendering pipelines per material class: metals get directional specular; wood gets grain-aligned shading; cloth gets soft diffuse folds; magic gets bloom, glow, and particle trails; liquids get transparency and surface distortion; organic surfaces get subsurface color shifting.
        - **Entity visual realism** — anatomically appropriate body plans per creature type (not all bipeds), visible muscle/joint structure in shading, equipment that conforms to body shape, damage states visible as sprite degradation, status effects rendered as material changes (frozen = blue tint + frost particles, burning = orange glow + smoke).
    - **UI/UX improvements (roll of 15–19 — HIGH PRIORITY — address the KNOWN UI/UX PROBLEMS above):**
        - **Overlap elimination** — implement UI spatial reservation system that prevents any two UI elements from occupying the same screen pixels. Damage numbers fan or stack. Tooltips reposition dynamically. Health bars offset when clustered. Priority system determines which element yields.
        - **Screen real estate maximization** — audit every HUD element for information density. Replace large panels with compact icon strips. Make panels collapsible or auto-hiding. Use transparency for non-focused UI. Push all chrome to screen edges. Target ≥85% viewport for game world.
        - **In-world indicator decluttering** — proximity-based detail levels for entity overhead UI. Targeted/nearby entities get full indicators; mid-range gets minimal; distant gets none. Aggregate overlapping indicators. Fade with distance. Player can configure detail threshold.
        - **Information hierarchy redesign** — size, brightness, and animation budget proportional to importance. Player vitals: large, bright, always visible. Active threats: highlighted, animated. Passive info: small, dim, or hidden until hovered/relevant. Implement a priority tier system for all UI elements.
        - **Interactive UI polish** — hover states, press feedback, smooth transitions (ease-in-out, 150ms), focus indicators, and micro-animations on all clickable/interactive elements. Menus slide, panels fade, buttons depress. Every interaction should feel physically responsive.
        - **Compact navigation UI** — small semi-transparent minimap with auto-collapse, in-world subtle directional cues, waypoint system that uses world-space markers rather than HUD overlays, fog-of-war on minimap matching explored state.
        - **Text rendering guarantee** — ensure all text elements (damage numbers, entity names, tooltips, HUD labels) actually render on screen using a bundled or procedurally generated bitmap font. Implement fallback font rendering, verify glyph availability, and test that no text is silently invisible.
        - **Crosshair and weapon feedback** — add a visible aiming reticle that tracks player aim direction, melee swing arc / slash trail animations, ranged weapon muzzle flash or projectile visuals, and impact burst animations at hit locations. Every attack must produce visible confirmation feedback.
    - **Technical improvements supporting visuals/UI (roll of 20 — lower priority):**
        - **Sprite render pipeline optimization** — batch draw calls by material type, cache composed sprites with LRU eviction, pool image buffers by size, eliminate per-frame allocations in render path. Visual improvements must not drop below 60 FPS.
        - **UI layout engine** — constraint-based layout solver for HUD elements, automatic reflow on window resize, spatial hash for overlap detection, dirty-flag rendering to avoid unnecessary redraws.

- If multiple candidates exist within your category, pick the one that most improves visual realism or UI/UX quality. Within visual realism: material-realistic sprite rendering, weathered environments, and realistic lighting are highest-value. Within UI/UX: overlap elimination, screen real estate maximization, text rendering guarantee, and crosshair/weapon feedback are highest-value.

STEP 2 — IMPLEMENT (this is the bulk of the work):
Follow these rules strictly. Violations are build failures.

Architecture:
- Components = pure data + `Type() string`. Zero methods with logic.
- Systems = all logic. Signature: `Update(entities []*Entity, deltaTime float64)`.
- Procgen: `rand.New(rand.NewSource(seed))` only. Never global rand. Never time.Now().
- Logging: `logrus.WithFields(logrus.Fields{"system_name": "...", ...})`.
- No external assets. No new dependencies beyond go.mod.

Visual Realism Standards:
- **Shading is mandatory**: No sprite may be rendered as flat solid fills. Every surface must show a light direction via highlight placement and shadow gradients. Use at least 3 tonal steps per color area (highlight, midtone, shadow).
- **Material identity**: A metal surface must read as metal (specular highlight, dark reflection edge). Wood must read as wood (grain lines, warm tones). Stone must read as stone (granular noise, cool tones, crack lines). The viewer should identify the material without labels.
- **Lighting consistency**: All sprites in a scene must share a consistent light direction. Shadow placement, highlight placement, and ambient occlusion must agree. No sprite should appear to be lit from a different angle than its neighbors.
- **Procedural weathering**: Generated environments must include wear — no pristine surfaces. Walls have edge damage. Floors have stains. Metal has patina or rust. Wood has grain and knot variation. The amount of weathering should be seed-driven and vary by location.
- **Depth through shading**: Use darker values at edges and contact points, lighter values on raised/exposed surfaces. Add subtle ambient occlusion where objects meet floors/walls. Use atmospheric perspective (slight desaturation and brightness shift) for distant elements.
- **Color temperature**: Light sources tint nearby surfaces. Torches add warmth (shift toward orange). Magical sources add their characteristic color. Ambient/indirect areas shift cooler (slight blue). This temperature variation adds enormous realism at low cost.
- Lighting must use radial gradients with proper falloff (inverse-square preferred, linear acceptable for performance). No flat circles.
- Shadows use soft penumbra with distance-based falloff. Support genre-specific opacity presets.
- Animation playback: target 12 FPS (0.083s per frame), 8 frames per state. Use distance-based LOD (full rate at ≤200px, half at ≤400px, minimal beyond).
- Sprite generation must be seeded and cached (LRU, max 100 entries). Pool image buffers by size bucket.
- All visual enhancements must maintain 60+ FPS. Profile before and after with `go test -bench`.

Sprite Standards:
- Use perspective appropriate to the game's camera. If the game is top-down, sprites must be top-down. If the game is side-scrolling, sprites must be side-view. Match the camera.
- Use composite layered rendering where appropriate. Define a clear z-order for body parts, equipment, and effects.
- Every pixel matters at small sprite sizes. Use shading, color gradients, highlights, and dithering to give depth, material identity, and personality. **Flat solid fills are prohibited.**
- Different entity types must be visually distinguishable at a glance. Seed-based generation should produce genuine variety — not trivial color swaps on one template.
- Nonhumanoid creatures need dedicated body-plan templates. A spider should look like a spider, not a human with extra arms. Build anatomy templates for each broad creature category (quadrupeds, insects, serpents, flying creatures, amorphous entities, etc.).
- Equipment and status effects should be visible on the sprite. The player should be able to read entity state from visuals alone.

UI/UX Standards:
- **No overlapping elements**: Implement collision detection for UI elements. If two elements would overlap, the lower-priority one must reposition, shrink, fade, or defer. Test with worst-case scenarios (many enemies clustered, multiple simultaneous damage numbers, tooltip near screen edge).
- **Maximize playable area**: The game world viewport must occupy at least 85% of total screen area. HUD elements hug edges and use minimal footprint. Panels that aren't actively needed must auto-hide or collapse to icons. No decorative borders or chrome that consume pixels without conveying information.
- **Distance-based detail**: In-world UI (health bars, names, indicators) must scale with relevance. Full detail only for targeted or adjacent entities. Mid-range entities get minimal indicators. Distant entities get none. Implement smooth fade transitions, not abrupt pop-in/out.
- **Information hierarchy**: Assign every UI element a priority tier (Critical / Important / Secondary / Ambient). Critical elements (player health, active threats) are always visible, large, and bright. Ambient elements (distant NPC status, passive buffs) are small, dim, and hidden until the player focuses on them.
- **Responsive interactions**: Every clickable/hoverable UI element must have distinct idle, hover, and pressed visual states. State transitions use easing (ease-out, 100-200ms). Focus/selection must be clearly indicated. No "dead" buttons that give no feedback on interaction.
- **Screen-edge awareness**: Tooltips, popup menus, and floating indicators must detect screen boundaries and reposition to stay fully visible. No element may be clipped by the screen edge.
- **Scalability**: UI must remain functional and non-overlapping at different window sizes and entity counts. Test at minimum with 20+ visible entities and 10+ simultaneous damage numbers.
- **Text must render**: Every text element must be visually confirmed to appear on screen. Use a bundled bitmap font or procedurally generated glyphs. If primary font fails, fall back to a guaranteed renderer. No silent text rendering failures — if text cannot be drawn, log a warning and draw a placeholder rectangle. Test damage numbers, entity labels, and HUD text simultaneously.
- **Crosshair and weapon feedback are mandatory**: The player must always have a visible aiming reticle. Every attack action must produce a visible feedback animation (swing arc, projectile trail, impact burst). The player must never be uncertain about where they are aiming or whether their attack registered. Crosshair style should adapt to weapon type (melee vs. ranged).

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
1. **Enhancement**: What and why, 2-3 sentences. Specifically describe how visual realism improved or how UI/UX issues were resolved.
2. **Files**: List with one-line change summary each.
3. **Integration**: Where the system is registered (exact file + function).
4. **Verification**: How to observe the improvement in-game — what should look more realistic, or what UI problem is now fixed.

STOP when the report is written and builds pass. Do not refactor unrelated code. Do not write documentation files. Do not suggest follow-up work.

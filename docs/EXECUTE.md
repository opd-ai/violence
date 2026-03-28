You are implementing ONE novel enhancement to a Go/Ebiten procedural raycasting FPS called Violence. Act autonomously. Do not ask for approval.

The overarching goal of every enhancement is to make in-game **objects, creatures, and humanoids** more **recognizable** (the player instantly identifies what something is), more **realistic** (entities look like plausible physical things, not programmer-art placeholders), and more **distinguishable** (no two entity categories look alike; variety within a category is genuine, not trivial recolors).

EXISTING VISUAL SYSTEMS (understand these before writing any code):
The codebase already has substantial visual infrastructure. Read the relevant packages before duplicating or conflicting with existing work.

| Package | Purpose | Key capabilities |
|---------|---------|-----------------|
| `pkg/sprite/` | Procedural sprite generation | 6 body-plan templates (humanoid, quadruped, insect, serpent, flying, amorphous); 9 material layers (scales, fur, chitin, membrane, metal, cloth, leather, crystal, slime); PBR shading (cylindrical/spherical); dithering; rim lighting; LRU cache |
| `pkg/texture/`, `pkg/walltex/` | Wall/floor texture generation | 7 material types (stone, metal, wood, concrete, organic, crystal, tech); normal mapping; weathering; genre presets with weather intensity, detail density, glow |
| `pkg/render/` | Raycasting render pipeline | DDA raycaster; framebuffer compositing; texture sampling; light/fog integration; post-processing; edge AO |
| `pkg/lighting/` | Atmospheric lighting | Point lights with falloff; shadow ray-tracing; exponential fog; color temperature; depth fade; ambient occlusion from occluders |
| `pkg/animation/` | State-based sprite animation | 8 animation states × 8 directions; LOD frame-rate scaling; sprite caching; archetype colors; state effects |
| `pkg/damagestate/` | Damage visualization | 5 damage levels; directional wound overlays; genre-themed damage colors |
| `pkg/equipment/` | Equipment rendering | 8 slots; 10 materials; 5 rarities; 4 damage states; layered compositing; enchantment glow |
| `pkg/surfacesheen/` | Material reflections | Genre-preset specular highlights; metal saturation; wet-surface boost; warmth shift |
| `pkg/outline/` | Entity silhouettes | Genre-colored outlines per entity type (player/enemy/ally/neutral/interact); thickness and glow |
| `pkg/healthbar/` | Overhead health bars | Genre-themed colors; status icons; proximity fade |
| `pkg/props/` | Environmental prop placement | 10 prop types with genre-specific spawn templates and weights |
| `pkg/decoration/` | Room decoration | 9 room types; 4 decoration categories; genre-tuned density |
| `pkg/itemicon/` | Item icon generation | Rarity-based complexity; enchantment glow; genre theming |
| `pkg/telegraph/` | Attack telegraphing | AOE/charge/cone indicators with genre colors |
| `pkg/particle/` | Particle effects | Emitters for combat, ambient, and status effects |
| `pkg/flicker/` | Flame/torch flicker | Multi-frequency noise; guttering events; color temperature variation |
| `pkg/procgen/genre/` | Genre registry | Constants: `Fantasy`, `SciFi`, `Horror`, `Cyberpunk`, `PostApoc` |

KNOWN RECOGNIZABILITY, REALISM & DISTINGUISHABILITY PROBLEMS:
These are the highest-priority visual issues. Each degrades the player's ability to identify, trust, or differentiate in-game entities.

1. **CREATURES LACK ANATOMICAL IDENTITY**
   Body-plan templates exist (humanoid, quadruped, insect, serpent, flying, amorphous) but produce silhouettes that are too similar at gameplay distances. A quadruped's profile barely differs from a squat humanoid. Insects read as blobs with lines. Improvements needed:
   - Exaggerate silhouette-defining features: long necks on serpents, wide wingspans on flyers, segmented thorax on insects, hunched mass on quadrupeds, irregular boundary on amorphous.
   - Add anatomical landmarks that survive downscaling: eyes (bright contrast dots), joints (dark creases), extremities (claws, antennae, tails) drawn at sprite-edge where they affect outline shape.
   - Ensure each body plan occupies a distinct aspect ratio and center-of-mass position so even a 16×16 thumbnail is identifiable.

2. **HUMANOIDS ARE INTERCHANGEABLE**
   All humanoid enemies share a single torso-arms-legs-head template. Role (tank, ranged, healer, ambusher, scout) is conveyed only by subtle color shifts. At a glance, every humanoid looks the same. Improvements needed:
   - Vary body proportions by role: tanks are wide-shouldered and stocky; scouts are lean and narrow; ranged units have exaggerated arms/weapon silhouettes.
   - Add role-specific silhouette markers: shield outline on tanks, staff/bow outline on ranged, hood/mask on ambushers, glowing hands on healers, pack/belt on scouts.
   - Ensure equipment rendering (`pkg/equipment/`) produces visible, role-consistent gear — heavy plate reads differently from light leather.

3. **OBJECTS LACK FUNCTION CUES**
   Props (barrels, crates, terminals, torches, etc.) are generated as generic rectangles or circles with color variation. The player cannot tell a barrel from a crate from a container at typical view distances. Improvements needed:
   - Give each prop type a distinctive silhouette: barrels are rounded with stave lines; crates are angular with plank edges; terminals have screen glow; torches have flame particle; pillars are tall and narrow.
   - Add material-appropriate surface detail: wood grain on crates, metal banding on barrels, cracked stone on pillars, flickering light on torches.
   - Loot pickups, destructibles, and interactable objects need a visual language (glow, outline, particle hint) that separates them from pure decoration at a glance.

4. **SEED VARIETY IS SUPERFICIAL**
   Different seeds for the same entity type produce color palette swaps but preserve identical shape, proportion, and feature placement. Two "different" goblins are the same goblin in different shirts. Improvements needed:
   - Seeds should modulate structural parameters: limb length ratios, head size, body width, stance, accessory presence/absence, asymmetric features (scars, missing parts, extra horns).
   - For creatures, seeds should vary the number and placement of features (eyes, horns, spines, spots, stripes) — not just their color.
   - For objects, seeds should vary wear level, accumulated damage, attached details (moss, cobwebs, dents, labels).

5. **MATERIALS DON'T READ AT SMALL SIZES**
   The 9 material layers (scales, fur, chitin, etc.) apply subtle procedural patterns that vanish below ~32px. At typical sprite sizes (16–24px), all materials look like flat fills with noise. Improvements needed:
   - Use high-contrast macro features for material identity: fur gets a ragged silhouette edge; chitin gets glossy highlight bands; scales get a regular dot/diamond grid; metal gets a single bright specular spot; slime gets transparency and wobble.
   - Prioritize silhouette-edge effects over interior texture at small sizes — the outline shape communicates material faster than interior pixels.
   - Ensure the 3-tone shading (highlight / midtone / shadow) uses material-appropriate tone mapping: warm for organic, cool for metal, saturated for crystal, muted for cloth.

6. **GENRE THEMES DON'T CHANGE ENTITY CHARACTER**
   Genre switching (`SetGenre`) changes color palettes and some density parameters but doesn't alter how creatures or objects fundamentally look and behave. A fantasy skeleton and a sci-fi skeleton are the same skeleton in different colors. Improvements needed:
   - Genre should influence body-plan selection and weighting: sci-fi favors mechanical/augmented humanoids; horror favors amorphous and insect types; fantasy favors quadrupeds and humanoids; cyberpunk favors humanoids with tech augmentations; post-apocalyptic favors mutated variants.
   - Genre should add signature visual elements: sci-fi entities get visor/antenna/circuit lines; horror entities get elongated limbs and asymmetry; cyberpunk entities get neon accent lines and chrome patches.
   - Props and objects should shift archetype with genre: fantasy torch → sci-fi light panel → horror candelabra → cyberpunk neon tube → post-apoc burning barrel.

7. **LIGHTING DOESN'T REVEAL FORM**
   The atmospheric lighting system (`pkg/lighting/`) calculates per-point light values, but these aren't used to modulate sprite shading in a way that reveals 3D form. Sprites are pre-shaded during generation with a fixed light direction, then uniformly dimmed/brightened by the lighting system. Improvements needed:
   - Sprite generation should produce a normal-mapped or hemisphere-shaded sprite that responds to the dominant light direction at render time.
   - At minimum, sprite brightness should vary across the sprite based on light source position — the side facing a torch should be brighter than the opposite side.
   - Ground contact shadows (simple dark ellipse beneath entities) are cheap and dramatically improve grounding and spatial reading.

STEP 1 — DISCOVER (spend ≤5 minutes):
- Run `git log --oneline -20` to see recent work and avoid duplicating it.
- Read the relevant package source to understand what already exists (see table above).
- Grep for TODO, FIXME, stub, placeholder in sprite, texture, render, lighting, and animation packages.
- Pick ONE enhancement from the problems listed above. Choose the one where you can make the most visible improvement with a focused, well-tested change. Prioritize:
    1. **(Highest)** Creature anatomical identity and humanoid role differentiation (#1, #2) — these produce the biggest instant improvement in recognizability.
    2. **(High)** Object function cues and seed-driven structural variety (#3, #4) — these improve distinguishability.
    3. **(Medium)** Material readability at small sizes and genre entity character (#5, #6) — these improve both realism and recognizability.
    4. **(Lower)** Lighting-driven form revelation (#7) — high-value but touches the render pipeline, which is more delicate.

STEP 2 — IMPLEMENT:
Follow these rules strictly. Violations are build failures.

Architecture:
- Components = pure data + `Type() string`. Zero methods with logic.
- Systems = all logic. Signature: `Update(w *engine.World)`.
- Procgen: `rng.NewRNG(seed)` only. Never global rand. Never `time.Now()` in generation code (it breaks determinism). `time.Now()` is fine for non-deterministic runtime operations like frame timestamps and profiling.
- Logging: `logrus.WithFields(logrus.Fields{"system": "...", ...})`.
- No external assets. No new dependencies beyond go.mod.

Recognizability Standards (what is it?):
- **Silhouette is king**: At 16×16px, each entity category must have a recognizably different outline shape. Humanoid ≠ quadruped ≠ insect ≠ serpent ≠ flyer ≠ amorphous. Test by squinting or blurring the sprite — if two types merge, they need more silhouette contrast.
- **Anatomical landmarks**: Every creature must have at least two high-contrast anchor points (e.g., eyes, joints, weapon tip) that survive downscaling and instantly communicate "this is alive/dangerous/interactive."
- **Functional affordance for objects**: The player must be able to tell what an object does from its shape alone. Containers look openable. Torches look like light sources. Terminals look interactive. Barrels look destructible. If a prop requires a label to identify, its sprite has failed.
- **Aspect ratio differentiation**: Tall-narrow (pillars, standing humanoids), wide-low (quadrupeds, crates), small-round (pickups, amorphous), elongated (serpents, weapons). Entity categories should cluster in different regions of aspect-ratio space.

Realism Standards (does it look like a plausible physical thing?):
- **3-tone minimum shading**: Every surface must show highlight, midtone, and shadow relative to a consistent light direction. No flat solid fills anywhere.
- **Material-appropriate rendering**: Metal = specular highlight + dark reflection edge. Fur = ragged edge + warm diffuse. Chitin = glossy bands + hard edge. Scales = regular pattern + iridescent highlight. Cloth = soft folds + matte diffuse. Slime = transparency + wobble highlight. The material must be identifiable without labels.
- **Weight and grounding**: Entities that touch the ground need a contact shadow (dark pixels at the base). Heavy entities are wider at the base. Light/flying entities have clear air gap. The sprite should communicate mass.
- **Consistent light direction**: All sprites in a scene agree on where light comes from. Highlight placement, shadow placement, and ambient occlusion are coherent. No sprite appears lit from a contradictory angle.
- **Organic imperfection**: Procedural generation must include asymmetry, wear, and irregularity. Perfect symmetry reads as artificial. Creatures have scars, uneven features, asymmetric accessories. Objects have dents, stains, chipped edges. The amount of imperfection is seed-driven.
- **Animation sells life**: Idle creatures breathe (subtle vertical oscillation). Walking has weight (acceleration on start, deceleration on stop). Attacks have anticipation and follow-through. Even 2-frame idle sway makes an entity feel alive vs. a static cutout.

Distinguishability Standards (can I tell these apart?):
- **Inter-category contrast**: Any two entity categories (e.g., humanoid vs. insect) must differ in at least 3 visual dimensions: silhouette shape, color palette range, movement pattern, size class, and material type.
- **Intra-category variety**: Seed-based generation for entities of the same type must produce visually distinct individuals. Two goblins with different seeds should differ in at least: body proportions, accessory presence, color distribution, and one structural feature (e.g., helmet vs. bare head, one-handed vs. two-handed weapon).
- **Role readability for humanoids**: Each humanoid role (tank, ranged, healer, ambusher, scout) must be identifiable from silhouette alone. Width, height, weapon shape, armor coverage, and posture must all encode role. A player who has seen one tank should recognize the next tank instantly, even with a different seed/color.
- **Genre-shifted identity**: The same logical entity type under different genres should look thematically different — not just recolored. A fantasy warrior and a sci-fi warrior should differ in silhouette (sword vs. rifle), material (plate vs. powered armor), and accent (cape vs. visor glow). Genre shifting should touch shape, not just hue.
- **Interactive vs. decorative distinction**: Objects the player can interact with (loot, doors, switches, destructibles) must be visually distinct from pure scenery. Use the outline system (`pkg/outline/`), subtle particle hints, or brightness contrast to separate interactive from decorative. The player should never miss a pickup because it looked like floor debris.

Genre Coverage (mandatory for all new features):
- Every `SetGenre` implementation must handle all 5 genres: `fantasy`, `scifi`, `horror`, `cyberpunk`, `postapoc`.
- Each genre must produce visually distinct results — not just palette swaps. Shape, proportion emphasis, material priority, and accent elements should all shift with genre.
- Include an explicit `default` case that falls back to `fantasy`.

Integration (mandatory — this is where past attempts fail):
- Register the new system in the system initialization function.
- Register in the client handler setup — add to the systems container AND call the appropriate AddSystem method.
- If your system's Update signature differs from `Update(w *engine.World)`, create a wrapper matching existing patterns.
- The system MUST be on by default. No feature flags.
- Persistent component data must integrate with entity serialization/deserialization, or be explicitly transient.
- Verify the full chain: Definition → Instantiation → Registration → Update Loop → Output → Consumer → Player Effect. If any link is missing, the feature is dangling.

Performance:
- All visual enhancements must maintain 60+ FPS. Profile before and after with `go test -bench`.
- Sprite generation must be seeded and cached (LRU, max 100 entries). Pool image buffers by size bucket.
- Animation playback: target 12 FPS (0.083s per frame), 8 frames per state. Use distance-based LOD (full rate at ≤200px, half at ≤400px, minimal beyond).

Constraints:
- Keep changes focused and targeted. One enhancement, done well.
- `go build ./...`, `go test -race ./...`, and `go vet ./...` must pass.
- Write table-driven tests. Target ≥40% coverage on new code (≥30% for display-dependent packages).
- No breaking changes to saves, network protocol, or configs.

STEP 3 — VERIFY:
Run `go build ./...`, `go test -race ./...`, and `go vet ./...`. Fix any errors before reporting.

STEP 4 — REPORT (keep concise):
1. **Enhancement**: Which recognizability/realism/distinguishability problem you addressed and what changed, 2-3 sentences.
2. **Before/After**: Describe what the entity or object looked like before and how it looks now. Be specific about silhouette, material, or variety changes.
3. **Files**: List with one-line change summary each.
4. **Integration**: Where the system is registered (exact file + function).
5. **Verification**: How to observe the improvement in-game — what should the player now be able to recognize or distinguish that they couldn't before.

STOP when the report is written and builds pass. Do not refactor unrelated code. Do not write documentation files. Do not suggest follow-up work.

# Implementation Plan: v4.0 — Gameplay Expansion: Secrets, Upgrades, Squad AI, Storytelling

## Phase Overview
- **Objective**: Add depth and replayability through secrets, weapon upgrades, squad AI companions, and procedurally generated environmental storytelling
- **Source Document**: ROADMAP.md (v4.0 milestone)
- **Prerequisites**: v3.0 Visual Polish complete (textures, lighting, particles, post-processing implemented)
- **Estimated Scope**: Large

## Implementation Steps

### 1. Secret Wall System (`pkg/secret/`) ✅
- **Deliverable**: `secret.go` with push-wall mechanic, `secret_test.go` with 85%+ coverage
- **Dependencies**: `pkg/bsp` (for secret placement), `pkg/automap` (for secret annotations)
- **Implementation**:
  - Define `SecretWall` struct with position, trigger state, slide direction, slide progress
  - Implement `Trigger(entityID)` that initiates wall slide animation
  - Add `Update(deltaTime)` for smooth wall movement over ~1 second
  - Integrate with BSP generator via `PlaceSecrets(density float64)` method
  - Add automap annotation type for discovered secrets
- **Completed**: 2026-02-28
  - Implemented `pkg/secret` with `SecretWall` struct and `Manager` for tracking all secrets
  - Trigger mechanics with `Trigger()` method that initiates animation
  - Update system with `Update(deltaTime)` using linear interpolation (98.1% test coverage)
  - `GetOffset()` for smooth rendering during slide animation (16 frames over 1 second)
  - BSP integration already existed via `TileSecret` constant and `placeSecrets()` method
  - Extended `pkg/automap` with `AnnotationType` and `AddAnnotation()` for secret markers (100% test coverage)

### 2. Secret Reward System ✅
- **Deliverable**: Extended `pkg/loot` with `SecretReward` type, rare item tables
- **Dependencies**: `pkg/secret` (trigger detection), `pkg/weapon` (rare weapons), `pkg/ammo` (rare ammo)
- **Implementation**:
  - Define `SecretLootTable` with weighted rare items (powerful weapons, bulk ammo, lore items)
  - Implement `GenerateSecretReward(seed uint64)` returning deterministic rare drops
  - Add secret-specific item tiers: uncommon (30%), rare (50%), legendary (20%)
- **Completed**: 2026-02-28
  - Implemented `SecretLootTable` with three rarity tiers (Uncommon, Rare, Legendary)
  - `GenerateSecretReward(seed)` uses deterministic RNG for reproducible rewards
  - Proper distribution: 30% uncommon, 50% rare, 20% legendary (verified via tests)
  - `AddItem()` method for dynamically extending loot tables
  - Comprehensive test coverage (91.3%) including determinism and distribution tests

### 3. Weapon Upgrade System (`pkg/upgrade/`) ✅
- **Deliverable**: `upgrade.go` with upgrade token mechanics, `upgrade_test.go` with 85%+ coverage
- **Dependencies**: `pkg/weapon` (upgrade application), `pkg/shop` (upgrade purchase)
- **Implementation**:
  - Define `UpgradeToken` as collectible currency
  - Create `WeaponUpgrade` struct with: damage multiplier, fire rate modifier, clip size bonus, accuracy bonus, range bonus
  - Implement `ApplyUpgrade(weapon *Weapon, upgrade UpgradeType)` modifying weapon stats
  - Add genre-specific upgrade names via `SetGenre()`: enchantment (fantasy), calibration (scifi), augmentation (cyberpunk), modification (horror), retrofit (postapoc)
- **Completed**: 2026-02-28
  - Implemented `UpgradeToken` for currency management with `Add()`, `Spend()`, `GetCount()`
  - Created `WeaponUpgrade` with 5 upgrade types: Damage, FireRate, ClipSize, Accuracy, Range
  - `ApplyWeaponStats()` applies multiplicative bonuses to weapon stats (avoids balance issues)
  - Genre-specific naming system for all 5 genres × 5 upgrade types = 25 unique names
  - `Manager` tracks per-weapon upgrades with `ApplyUpgrade()`, `GetUpgrades()`, `HasUpgrade()`
  - Test coverage: 86.0%

### 4. Weapon Mastery System ✅
- **Deliverable**: `mastery.go` tracking per-weapon XP, passive bonuses
- **Dependencies**: `pkg/weapon`, `pkg/progression` (XP events)
- **Implementation**:
  - Define `WeaponMastery` with XP track (0-1000 per weapon), milestone unlocks at 250/500/750/1000
  - Passive bonuses: headshot damage +10% (250), reload speed +15% (500), accuracy +10% (750), critical chance +5% (1000)
  - Implement `AddMasteryXP(weaponID, amount)` called on kills with that weapon
- **Completed**: 2026-02-28
  - Implemented `mastery.go` in `pkg/weapon` with `WeaponMastery` struct tracking XP and bonuses per weapon slot
  - `MasteryManager` manages all weapon masteries via `AddMasteryXP()` method
  - Milestone system with 4 unlockable bonuses at 250/500/750/1000 XP thresholds
  - `GetBonus()` returns multiplicative modifiers for headshot damage, reload speed, accuracy, and critical chance
  - `GetProgressToNextMilestone()` calculates percentage progress for UI display
  - Helper methods: `GetCurrentMilestone()`, `GetMilestoneDescription()`, `Reset()`
  - Test coverage: 100% (mastery.go), 98.1% overall package coverage

### 5. Survival Skills System (`pkg/skills/`) ✅
- **Deliverable**: Extended `skills.go` with three skill trees, `skills_test.go` with 85%+ coverage
- **Dependencies**: `pkg/progression` (level-up points)
- **Implementation**:
  - Define three trees: Combat (damage, reload, accuracy), Survival (health, armor, stamina), Tech (hacking, stealth, detection)
  - Each tree has 5 nodes with prerequisites forming a directed graph
  - Implement `AllocatePoint(treeID, nodeID)` validating prerequisites
  - Add passive effect application via `GetModifier(stat string) float64`
- **Completed**: 2026-02-28
  - Implemented `Manager` with three pre-configured skill trees (Combat, Survival, Tech)
  - Combat tree: 5 nodes for damage, reload speed, and accuracy bonuses with combat_master capstone
  - Survival tree: 5 nodes for max health, armor, stamina, health regen with survival_master capstone
  - Tech tree: 5 nodes for hacking, stealth, and detection bonuses with tech_master capstone
  - `AllocatePoint(treeID, nodeID)` validates prerequisites and point costs before allocation
  - `GetModifier(stat)` returns cumulative bonuses across all trees including mastery bonuses
  - Prerequisites form directed acyclic graphs (e.g., combat_master requires both combat_dmg_2 and combat_accuracy_1)
  - Test coverage: 98.5%

### 6. Extended Inventory System (`pkg/inventory/`) ✅
- **Deliverable**: Extended `inventory.go` with active-use items, quick-use slot
- **Dependencies**: `pkg/input` (use keybind)
- **Implementation**:
  - Add `ActiveItem` interface with `Use(user *Entity)` method
  - Implement concrete items: `Grenade`, `ProximityMine`, `Medkit`
  - Add `QuickSlot` for fast item access (key binding: Q)
  - Implement `UseQuickSlot()` consuming and applying quick-slot item
- **Completed**: 2026-02-28
  - Implemented `ActiveItem` interface with `Use(user *Entity) error`, `GetID()`, `GetName()` methods
  - Concrete items: `Grenade` (throwable explosive), `ProximityMine` (placeable trap), `Medkit` (healing with fixed or percentage heal)
  - `QuickSlot` with thread-safe `Set()`, `Get()`, `Clear()`, `IsEmpty()` operations
  - `UseQuickSlot(user)` validates item presence, uses item, consumes from inventory, and auto-clears when depleted
  - Thread-safe inventory operations using sync.RWMutex for concurrent access
  - Test coverage: 95.4% (21 test cases including concurrent access)

### 7. Crafting System (`pkg/crafting/`) ✅
- **Deliverable**: Extended `crafting.go` with scrap-to-ammo recipes
- **Dependencies**: `pkg/inventory` (scrap storage), `pkg/ammo` (ammo creation)
- **Implementation**:
  - Define `Scrap` resource type with genre-skinned names: bone chips (fantasy), circuit boards (scifi), flesh (horror), data shards (cyberpunk), salvage (postapoc)
  - Create `Recipe` struct mapping scrap amounts to output items
  - Implement `CraftingMenu` UI with available recipe list
  - Add scrap drops to loot tables (enemies, destructibles)
- **Completed**: 2026-02-28
  - Implemented `Scrap` struct and `ScrapStorage` for thread-safe scrap management
  - `ScrapStorage` with `Add()`, `Remove()`, `Get()`, `GetAll()` operations
  - `CraftingMenu` integrates storage with recipes for UI presentation
  - `GetAvailableRecipes()` filters recipes by current scrap amounts
  - `Craft(recipeID)` validates materials, consumes scrap, and returns crafted items
  - `GetScrapNameForGenre()` returns genre-specific scrap names
  - Genre-specific recipes already existed with 5 recipes per genre (arrows, bolts, mana, explosives, potion variants)
  - Test coverage: 97.8% (27 test cases including concurrent access)

### 8. Level Objectives System (`pkg/quest/`)
- **Deliverable**: Extended `quest.go` with procedural objective generation
- **Dependencies**: `pkg/bsp` (objective placement), `pkg/rng` (deterministic selection)
- **Implementation**:
  - Define objective types: `FindExit`, `RetrieveItem`, `DestroyTarget`, `RescueHostage`
  - Implement `GenerateObjective(seed, levelLayout)` selecting and placing objectives
  - Add objective tracker HUD element showing current/bonus objectives
  - Bonus objectives: secret count threshold, kill count, speed run timer

### 9. Shop/Armory System (`pkg/shop/`)
- **Deliverable**: Extended `shop.go` with between-level purchasing
- **Dependencies**: `pkg/weapon`, `pkg/ammo`, `pkg/upgrade`, `pkg/inventory`
- **Implementation**:
  - Define `Credit` currency earned from kills, secrets, objectives
  - Create `ShopInventory` with weapons, ammo, upgrades, consumables
  - Implement `Purchase(itemID)` deducting credits and adding item
  - Add genre-skinned shop UI: merchant tent (fantasy), supply depot (scifi), black market (horror), corpo shop (cyberpunk), scrap trader (postapoc)

### 10. Squad Companion AI (`pkg/squad/`)
- **Deliverable**: `squad.go` with 1-3 AI companions, `squad_test.go` with 85%+ coverage
- **Dependencies**: `pkg/ai` (behavior trees), `pkg/class` (squad member classes)
- **Implementation**:
  - Define `SquadMember` struct with position, health, weapon, class, behavior state
  - Implement behavior states: Follow (default), Hold Position, Attack Target
  - Add squad commands via radial menu or hotkeys (F1-F3)
  - Squad pathfinding uses existing A* with formation offsets
  - Squad members use player class archetypes (Grunt, Medic, Demo, Mystic)

### 11. Environmental Storytelling (`pkg/lore/`)
- **Deliverable**: Extended `lore.go` with procedural note/log generation
- **Dependencies**: `pkg/bsp` (placement), `pkg/rng` (text generation)
- **Implementation**:
  - Define `LoreItem` types: written notes, audio logs, graffiti, body arrangements
  - Implement `GenerateLoreText(seed, genreID, context)` producing deterministic narrative
  - Create genre-specific word banks and sentence templates
  - Add flavor text popup on approach using seed-driven generation

### 12. Collectible Codex System
- **Deliverable**: Extended `lore.go` with codex UI, collectible tracking
- **Dependencies**: `pkg/ui` (codex screen), `pkg/save` (persistence)
- **Implementation**:
  - Define `CodexEntry` struct with title, content, discovered flag
  - Implement codex screen accessible from pause menu (C key)
  - Track discovered entries across run (persistent in save data)
  - Generate world backstory entries procedurally per seed

### 13. Hacking/Lockpicking Mini-Games (`pkg/minigame/`)
- **Deliverable**: Extended `minigame.go` with genre-specific puzzles
- **Dependencies**: `pkg/input` (mini-game controls), `pkg/door` (lock bypass)
- **Implementation**:
  - Implement `CircuitTraceHacking` for cyberpunk (trace path through circuit)
  - Implement `LockpickTension` for fantasy (timing-based tension game)
  - Implement `BypassCodeEntry` for scifi/postapoc (simple code input)
  - Add skip option consuming `BypassTool` consumable item

### 14. Destructible Environments (`pkg/destruct/`)
- **Deliverable**: Extended `destruct.go` with breakable walls/objects
- **Dependencies**: `pkg/bsp` (destructible placement), `pkg/particle` (debris effects)
- **Implementation**:
  - Define `BreakableWall` struct revealing hidden passage on destruction
  - Implement `DestructibleObject` for barrels, crates with explosion chains
  - Add debris blocking corridors temporarily (cleared after N seconds)
  - Genre-specific debris materials: stone rubble, hull shards, plaster, glass, concrete

### 15. World Events System (`pkg/event/`)
- **Deliverable**: Extended `event.go` with timed triggers
- **Dependencies**: `pkg/ai` (enemy alert), `pkg/door` (lockdown)
- **Implementation**:
  - Define `AlarmTrigger` putting all enemies in alert state, locking doors temporarily
  - Implement `TimedLockdown` with countdown timer and escape objective
  - Add `BossArenaEvent` spawning enemy wave on room entry
  - Genre-flavored event text/audio stings (all procedurally generated)

### 16. Props/Decoration System (`pkg/props/`)
- **Deliverable**: `props.go` with decorative sprite placement, `props_test.go` with 85%+ coverage
- **Dependencies**: `pkg/bsp` (prop placement), `pkg/raycaster` (sprite rendering)
- **Implementation**:
  - Define `Prop` struct with position, sprite type, collision flag
  - Create genre-specific prop lists: barrels, crates, tables, terminals, bones, plants
  - Implement `PlaceProps(room *Room, density float64)` for BSP integration
  - All prop sprites procedurally generated via `pkg/texture` patterns

## Technical Specifications

- **All text content** (lore notes, quest descriptions, codex entries) must be procedurally generated at runtime from deterministic algorithms seeded by `pkg/rng`
- **Squad AI pathfinding** reuses existing A* implementation in `pkg/ai` with formation offset calculation
- **Skill tree structure** uses directed acyclic graph with prerequisite validation
- **Weapon upgrades** modify weapon stats multiplicatively to avoid balance issues
- **Weapon mastery** tracks per-weapon XP (0-1000) with milestones granting passive bonuses
- **Secret walls** use linear interpolation for smooth slide animation (16 frames over 1 second)
- **Mini-games** must support keyboard and gamepad input via `pkg/input` abstraction
- **Destructible walls** integrate with BSP tile grid by changing tile type on destruction
- **Credit economy** balanced around: kill=10, secret=50, objective=100, bonus=25

## Validation Criteria

- [ ] Secret walls slide open on use-key press and reveal reward spawns
- [ ] Weapon upgrades persist across level transitions and modify weapon behavior
- [x] Weapon mastery tracks XP per weapon and unlocks passive bonuses at milestones (250/500/750/1000 XP)
- [x] Skill tree UI allows point allocation with visual prerequisite indicators
- [ ] Squad companions follow player, respond to commands, and engage enemies
- [ ] Lore items display procedurally generated text unique to seed
- [ ] Crafting converts scrap to ammo/items with correct resource consumption
- [ ] Quest objectives appear on HUD and complete when conditions met
- [ ] Shop allows purchase of items between levels with credit deduction
- [ ] Mini-games block door access until completed or bypassed
- [ ] Destructible objects spawn debris particles and chain explosions
- [ ] World events trigger correctly (alarms, lockdowns, boss spawns)
- [ ] Props render as sprites at correct positions without collision issues
- [ ] All new systems implement `SetGenre(genreID string)` for theme swapping
- [ ] Test coverage for new packages ≥85%
- [ ] No embedded text assets — all narrative content generated from seed

## Known Gaps

- **Squad AI formation algorithm**: Need to define formation shapes (line, wedge, column) and offset calculation from leader position
- **Procedural text generation grammar**: Need to design template grammar for lore generation that produces coherent narrative without repetition
- **Credit economy balance**: Initial values provided but require playtesting to tune for satisfying progression curve

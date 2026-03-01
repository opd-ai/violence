// Package loot implements loot drop mechanics for the ECS.
package loot

import (
	"fmt"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// LootDropComponent marks an entity with a loot table to drop on death.
type LootDropComponent struct {
	Table      *LootTable
	DropSeed   uint64  // Seed for deterministic drop generation
	DropChance float64 // Overall chance to drop any loot (0.0-1.0)
}

// Type implements Component interface.
func (c *LootDropComponent) Type() string {
	return "LootDropComponent"
}

// HealthComponent stores entity health (must match engine definition).
type HealthComponent struct {
	Current, Max float64
}

// Type implements Component interface.
func (c *HealthComponent) Type() string {
	return "HealthComponent"
}

// PositionComponent stores entity position (must match engine definition).
type PositionComponent struct {
	X, Y float64
}

// Type implements Component interface.
func (c *PositionComponent) Type() string {
	return "PositionComponent"
}

// LootItemComponent marks an entity as a loot item that can be picked up.
type LootItemComponent struct {
	ItemID      string
	Rarity      Rarity
	SpawnTime   float64
	LifetimeMax float64 // Despawn after this many seconds (0 = never)
}

// Type implements Component interface.
func (c *LootItemComponent) Type() string {
	return "LootItemComponent"
}

// LootDropSystem handles loot generation on entity death.
type LootDropSystem struct {
	rng              *rng.RNG
	gameTime         float64
	deathProcessed   map[engine.Entity]bool // Track processed deaths to avoid duplicates
	onLootSpawned    func(itemID string, x, y float64, rarity Rarity)
	logger           *logrus.Entry
	genreConfig      GenreDropConfig
	currentGenre     string
	nextLootEntityID uint64
}

// GenreDropConfig defines genre-specific loot drop settings.
type GenreDropConfig struct {
	DefaultDropChance   float64
	DefaultLifetime     float64
	CommonDrops         []Drop
	UncommonDrops       []Drop
	RareDrops           []Drop
	LegendaryDrops      []Drop
	EnableVisualEffects bool
}

// NewLootDropSystem creates the loot drop system.
func NewLootDropSystem(seed int64) *LootDropSystem {
	return &LootDropSystem{
		rng:            rng.NewRNG(uint64(seed)),
		deathProcessed: make(map[engine.Entity]bool),
		logger: logrus.WithFields(logrus.Fields{
			"system": "loot_drop",
		}),
		genreConfig:  getDefaultGenreConfig("fantasy"),
		currentGenre: "fantasy",
	}
}

// SetGenre configures the system for a specific genre.
func (s *LootDropSystem) SetGenre(genreID string) {
	s.currentGenre = genreID
	s.genreConfig = getDefaultGenreConfig(genreID)
	s.logger.WithField("genre", genreID).Debug("Loot drop system genre set")
}

// SetLootSpawnCallback sets a callback for when loot is spawned.
func (s *LootDropSystem) SetLootSpawnCallback(fn func(itemID string, x, y float64, rarity Rarity)) {
	s.onLootSpawned = fn
}

// Update processes entity deaths and spawns loot.
func (s *LootDropSystem) Update(w *engine.World) {
	deltaTime := 0.016 // Assume 60 FPS
	s.gameTime += deltaTime

	// Clean up processed deaths older than 1 second
	if len(s.deathProcessed) > 1000 {
		s.deathProcessed = make(map[engine.Entity]bool)
	}

	// Query entities with health and loot drop components
	lootDropType := reflect.TypeOf((*LootDropComponent)(nil))
	healthType := reflect.TypeOf((*HealthComponent)(nil))
	positionType := reflect.TypeOf((*PositionComponent)(nil))

	entities := w.Query(lootDropType, healthType, positionType)

	for _, entity := range entities {
		// Skip if already processed
		if s.deathProcessed[entity] {
			continue
		}

		// Get components
		healthComp, ok := w.GetComponent(entity, healthType)
		if !ok {
			continue
		}
		health := healthComp.(*HealthComponent)

		// Check if entity is dead
		if health.Current > 0 {
			continue
		}

		// Mark as processed
		s.deathProcessed[entity] = true

		// Get loot drop component
		lootDropComp, ok := w.GetComponent(entity, lootDropType)
		if !ok {
			continue
		}
		lootDrop := lootDropComp.(*LootDropComponent)

		// Get position
		posComp, ok := w.GetComponent(entity, positionType)
		if !ok {
			continue
		}
		pos := posComp.(*PositionComponent)

		// Process loot drop
		s.processLootDrop(w, entity, lootDrop, pos.X, pos.Y)
	}

	// Update loot item lifetimes and despawn expired items
	s.updateLootItems(w, deltaTime)
}

func (s *LootDropSystem) processLootDrop(w *engine.World, deadEntity engine.Entity, lootDrop *LootDropComponent, x, y float64) {
	// Roll for overall drop chance
	if lootDrop.DropChance < 1.0 {
		roll := s.rng.Float64()
		if roll > lootDrop.DropChance {
			s.logger.WithFields(logrus.Fields{
				"entity": deadEntity,
				"x":      x,
				"y":      y,
			}).Debug("Loot drop chance failed")
			return
		}
	}

	// Generate loot items
	var droppedItems []string
	if lootDrop.Table != nil {
		droppedItems = lootDrop.Table.RollWithSeed(lootDrop.DropSeed + uint64(deadEntity))
	}

	if len(droppedItems) == 0 {
		s.logger.WithFields(logrus.Fields{
			"entity": deadEntity,
			"x":      x,
			"y":      y,
		}).Debug("No items rolled from loot table")
		return
	}

	// Spawn loot items
	s.logger.WithFields(logrus.Fields{
		"entity": deadEntity,
		"x":      x,
		"y":      y,
		"items":  droppedItems,
	}).Info("Spawning loot drops")

	for i, itemID := range droppedItems {
		// Determine rarity from item ID (simple heuristic)
		rarity := s.determineItemRarity(itemID)

		// Offset position slightly for multiple drops
		offsetX := x + float64(i%3-1)*0.3
		offsetY := y + float64(i/3)*0.3

		// Create loot item entity
		lootEntity := w.AddEntity()
		w.AddComponent(lootEntity, &PositionComponent{X: offsetX, Y: offsetY})
		w.AddComponent(lootEntity, &LootItemComponent{
			ItemID:      itemID,
			Rarity:      rarity,
			SpawnTime:   s.gameTime,
			LifetimeMax: s.genreConfig.DefaultLifetime,
		})

		// Call spawn callback if set
		if s.onLootSpawned != nil {
			s.onLootSpawned(itemID, offsetX, offsetY, rarity)
		}

		s.logger.WithFields(logrus.Fields{
			"item_id": itemID,
			"rarity":  rarity,
			"x":       offsetX,
			"y":       offsetY,
		}).Debug("Loot item spawned")
	}
}

func (s *LootDropSystem) determineItemRarity(itemID string) Rarity {
	// Check against genre-specific drop tables to determine rarity
	for _, drop := range s.genreConfig.LegendaryDrops {
		if drop.ItemID == itemID {
			return RarityLegendary
		}
	}
	for _, drop := range s.genreConfig.RareDrops {
		if drop.ItemID == itemID {
			return RarityRare
		}
	}
	for _, drop := range s.genreConfig.UncommonDrops {
		if drop.ItemID == itemID {
			return RarityUncommon
		}
	}
	return RarityCommon
}

func (s *LootDropSystem) updateLootItems(w *engine.World, deltaTime float64) {
	lootItemType := reflect.TypeOf((*LootItemComponent)(nil))
	entities := w.Query(lootItemType)

	for _, entity := range entities {
		lootItemComp, ok := w.GetComponent(entity, lootItemType)
		if !ok {
			continue
		}
		lootItem := lootItemComp.(*LootItemComponent)

		// Check if item should despawn
		if lootItem.LifetimeMax > 0 {
			age := s.gameTime - lootItem.SpawnTime
			if age >= lootItem.LifetimeMax {
				s.logger.WithFields(logrus.Fields{
					"entity":  entity,
					"item_id": lootItem.ItemID,
					"age":     age,
				}).Debug("Loot item despawned")
				w.RemoveEntity(entity)
			}
		}
	}
}

func getDefaultGenreConfig(genreID string) GenreDropConfig {
	switch genreID {
	case "fantasy":
		return GenreDropConfig{
			DefaultDropChance: 0.7,
			DefaultLifetime:   30.0,
			CommonDrops: []Drop{
				{ItemID: "health_small", Chance: 0.5},
				{ItemID: "ammo_arrows", Chance: 0.4},
				{ItemID: "gold_coins", Chance: 0.3},
			},
			UncommonDrops: []Drop{
				{ItemID: "health_medium", Chance: 0.3},
				{ItemID: "mana_potion", Chance: 0.25},
				{ItemID: "scroll_fireball", Chance: 0.15},
			},
			RareDrops: []Drop{
				{ItemID: "health_large", Chance: 0.15},
				{ItemID: "enchanted_sword", Chance: 0.1},
				{ItemID: "magic_ring", Chance: 0.08},
			},
			LegendaryDrops: []Drop{
				{ItemID: "legendary_artifact", Chance: 0.05},
				{ItemID: "divine_weapon", Chance: 0.03},
			},
			EnableVisualEffects: true,
		}
	case "scifi":
		return GenreDropConfig{
			DefaultDropChance: 0.65,
			DefaultLifetime:   25.0,
			CommonDrops: []Drop{
				{ItemID: "energy_cell", Chance: 0.5},
				{ItemID: "credits", Chance: 0.4},
				{ItemID: "nano_repair", Chance: 0.35},
			},
			UncommonDrops: []Drop{
				{ItemID: "plasma_ammo", Chance: 0.3},
				{ItemID: "shield_booster", Chance: 0.25},
				{ItemID: "tech_upgrade", Chance: 0.2},
			},
			RareDrops: []Drop{
				{ItemID: "pulse_rifle", Chance: 0.12},
				{ItemID: "cybernetic_implant", Chance: 0.1},
			},
			LegendaryDrops: []Drop{
				{ItemID: "alien_artifact", Chance: 0.04},
				{ItemID: "experimental_weapon", Chance: 0.03},
			},
			EnableVisualEffects: true,
		}
	case "horror":
		return GenreDropConfig{
			DefaultDropChance: 0.5,
			DefaultLifetime:   40.0,
			CommonDrops: []Drop{
				{ItemID: "bandages", Chance: 0.4},
				{ItemID: "shotgun_shells", Chance: 0.35},
				{ItemID: "flashlight_battery", Chance: 0.3},
			},
			UncommonDrops: []Drop{
				{ItemID: "first_aid_kit", Chance: 0.25},
				{ItemID: "incendiary_ammo", Chance: 0.2},
				{ItemID: "sanity_pills", Chance: 0.18},
			},
			RareDrops: []Drop{
				{ItemID: "chainsaw_fuel", Chance: 0.1},
				{ItemID: "blessed_weapon", Chance: 0.08},
			},
			LegendaryDrops: []Drop{
				{ItemID: "ancient_talisman", Chance: 0.03},
				{ItemID: "exorcism_kit", Chance: 0.02},
			},
			EnableVisualEffects: true,
		}
	case "cyberpunk":
		return GenreDropConfig{
			DefaultDropChance: 0.75,
			DefaultLifetime:   20.0,
			CommonDrops: []Drop{
				{ItemID: "eddies", Chance: 0.5},
				{ItemID: "stim_pack", Chance: 0.4},
				{ItemID: "ammo_smart", Chance: 0.35},
			},
			UncommonDrops: []Drop{
				{ItemID: "cyberware_mod", Chance: 0.3},
				{ItemID: "hacking_chip", Chance: 0.25},
				{ItemID: "armor_plate", Chance: 0.2},
			},
			RareDrops: []Drop{
				{ItemID: "iconic_weapon", Chance: 0.12},
				{ItemID: "legendary_cyberware", Chance: 0.1},
			},
			LegendaryDrops: []Drop{
				{ItemID: "prototype_tech", Chance: 0.05},
				{ItemID: "corp_secret", Chance: 0.03},
			},
			EnableVisualEffects: true,
		}
	case "postapoc":
		return GenreDropConfig{
			DefaultDropChance: 0.6,
			DefaultLifetime:   35.0,
			CommonDrops: []Drop{
				{ItemID: "scrap_metal", Chance: 0.5},
				{ItemID: "purified_water", Chance: 0.4},
				{ItemID: "dirty_ammo", Chance: 0.35},
			},
			UncommonDrops: []Drop{
				{ItemID: "canned_food", Chance: 0.3},
				{ItemID: "radiation_meds", Chance: 0.25},
				{ItemID: "salvaged_weapon", Chance: 0.2},
			},
			RareDrops: []Drop{
				{ItemID: "pre_war_tech", Chance: 0.1},
				{ItemID: "military_gear", Chance: 0.08},
			},
			LegendaryDrops: []Drop{
				{ItemID: "vault_tech_prototype", Chance: 0.04},
				{ItemID: "old_world_relic", Chance: 0.02},
			},
			EnableVisualEffects: true,
		}
	default:
		return getDefaultGenreConfig("fantasy")
	}
}

// CreateEnemyLootTable creates a standard enemy loot table for a genre.
func CreateEnemyLootTable(genreID string, enemyTier int, seed uint64) *LootTable {
	config := getDefaultGenreConfig(genreID)
	table := NewLootTableWithRNG(rng.NewRNG(seed))

	// Add drops based on enemy tier
	switch enemyTier {
	case 1: // Common enemies
		table.Drops = append(table.Drops, config.CommonDrops...)
		// Small chance for uncommon
		for _, drop := range config.UncommonDrops {
			table.Drops = append(table.Drops, Drop{
				ItemID: drop.ItemID,
				Chance: drop.Chance * 0.3,
			})
		}
	case 2: // Elite enemies
		table.Drops = append(table.Drops, config.CommonDrops...)
		table.Drops = append(table.Drops, config.UncommonDrops...)
		// Moderate chance for rare
		for _, drop := range config.RareDrops {
			table.Drops = append(table.Drops, Drop{
				ItemID: drop.ItemID,
				Chance: drop.Chance * 0.5,
			})
		}
	case 3: // Boss enemies
		table.Drops = append(table.Drops, config.UncommonDrops...)
		table.Drops = append(table.Drops, config.RareDrops...)
		table.Drops = append(table.Drops, config.LegendaryDrops...)
	default:
		table.Drops = append(table.Drops, config.CommonDrops...)
	}

	return table
}

// CreateLootDropComponent creates a loot drop component for an entity.
func CreateLootDropComponent(table *LootTable, seed uint64, dropChance float64) *LootDropComponent {
	return &LootDropComponent{
		Table:      table,
		DropSeed:   seed,
		DropChance: dropChance,
	}
}

// String returns a string representation for logging.
func (r Rarity) String() string {
	switch r {
	case RarityCommon:
		return "Common"
	case RarityUncommon:
		return "Uncommon"
	case RarityRare:
		return "Rare"
	case RarityLegendary:
		return "Legendary"
	default:
		return fmt.Sprintf("Rarity(%d)", r)
	}
}

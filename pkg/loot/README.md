# Loot Package

The loot package implements a comprehensive loot drop system for the Violence game engine, providing deterministic item generation when entities are defeated.

## Features

- **Deterministic Loot Drops**: All drops use seeded RNG for reproducible results
- **Genre-Aware**: Different loot tables for fantasy, sci-fi, horror, cyberpunk, and post-apocalyptic settings
- **Tiered Enemies**: Support for common, elite, and boss enemy loot tables
- **ECS Integration**: Full integration with the Entity-Component-System architecture
- **Automatic Despawn**: Loot items automatically despawn after a configurable lifetime
- **Rarity System**: Common, Uncommon, Rare, and Legendary rarity tiers
- **Flexible Drop Chances**: Independent probability for each item in a loot table

## Core Components

### LootTable

Defines a set of possible item drops with independent probabilities:

```go
table := loot.NewLootTableWithRNG(rng.NewRNG(seed))
table.Drops = []loot.Drop{
    {ItemID: "health_pack", Chance: 0.5},  // 50% chance
    {ItemID: "ammo", Chance: 0.3},         // 30% chance
    {ItemID: "gold", Chance: 0.2},         // 20% chance
}

// Roll for drops (deterministic with seed)
items := table.RollWithSeed(12345)
```

### LootDropSystem

ECS system that monitors entity health and spawns loot on death:

```go
// Create system
lootSystem := loot.NewLootDropSystem(seed)
lootSystem.SetGenre("fantasy")

// Optional: Set callback for loot spawn events
lootSystem.SetLootSpawnCallback(func(itemID string, x, y float64, rarity loot.Rarity) {
    fmt.Printf("Spawned %s at (%.1f, %.1f)\n", itemID, x, y)
})

// Register with ECS world
world.AddSystem(lootSystem)
```

### LootDropComponent

Component that marks an entity as dropping loot on death:

```go
// Create enemy with loot drops
enemy := world.AddEntity()
lootTable := loot.CreateEnemyLootTable("fantasy", 2, seed) // Tier 2 elite
world.AddComponent(enemy, loot.CreateLootDropComponent(lootTable, seed, 0.75))
world.AddComponent(enemy, &loot.HealthComponent{Current: 100, Max: 100})
world.AddComponent(enemy, &loot.PositionComponent{X: 10, Y: 10})
```

## Enemy Tiers

The system supports three enemy tiers with progressively better loot:

- **Tier 1 (Common)**: Primarily common drops, small chance of uncommon
- **Tier 2 (Elite)**: Common and uncommon drops, moderate chance of rare
- **Tier 3 (Boss)**: Uncommon, rare, and legendary drops

```go
// Create tier-appropriate loot tables
commonTable := loot.CreateEnemyLootTable("fantasy", 1, seed)
eliteTable := loot.CreateEnemyLootTable("fantasy", 2, seed)
bossTable := loot.CreateEnemyLootTable("fantasy", 3, seed)
```

## Genre Configurations

Each genre has distinct loot tables and drop rates:

| Genre | Default Drop Chance | Default Lifetime | Theme |
|-------|---------------------|------------------|-------|
| Fantasy | 70% | 30s | Medieval fantasy items |
| Sci-Fi | 65% | 25s | Futuristic technology |
| Horror | 50% | 40s | Survival horror supplies |
| Cyberpunk | 75% | 20s | High-tech urban gear |
| Post-Apocalyptic | 60% | 35s | Scavenged resources |

## Rarity System

Items are categorized into four rarity tiers:

- **Common**: Base drops (health_small, ammo, basic currency)
- **Uncommon**: Enhanced items (better healing, quality weapons)
- **Rare**: Powerful equipment (enchanted weapons, advanced tech)
- **Legendary**: Unique artifacts (quest items, ultimate gear)

## Integration Example

```go
// 1. Create and configure the system
lootSystem := loot.NewLootDropSystem(12345)
lootSystem.SetGenre("fantasy")
world.AddSystem(lootSystem)

// 2. Create an enemy with loot
enemy := world.AddEntity()
enemyLoot := loot.CreateEnemyLootTable("fantasy", 2, 200)
world.AddComponent(enemy, loot.CreateLootDropComponent(enemyLoot, 200, 0.85))
world.AddComponent(enemy, &loot.HealthComponent{Current: 100, Max: 100})
world.AddComponent(enemy, &loot.PositionComponent{X: 20, Y: 20})

// 3. When enemy dies (health <= 0), loot is automatically spawned
// The system creates LootItemComponent entities at the death position
world.Update()
```

## Deterministic Behavior

All loot generation is deterministic based on seeds:

```go
table := loot.NewLootTable()
table.Drops = []loot.Drop{
    {ItemID: "coin", Chance: 0.5},
    {ItemID: "gem", Chance: 0.5},
}

// Same seed always produces same result
result1 := table.RollWithSeed(42) // e.g., ["coin"]
result2 := table.RollWithSeed(42) // e.g., ["coin"] (identical)
result3 := table.RollWithSeed(43) // e.g., ["coin", "gem"] (different seed)
```

This ensures:
- Reproducible gameplay for testing and debugging
- Consistent drops in networked multiplayer
- Replay compatibility
- No dependency on global random state

## Performance

- **Zero allocations on hot paths**: Object pooling and caching
- **Spatial partitioning ready**: Integrates with spatial query systems
- **Efficient death tracking**: Prevents duplicate processing
- **Automatic cleanup**: Expired loot items are removed to prevent memory leaks

## Testing

The package has 89.1% test coverage with comprehensive tests for:
- Deterministic drop generation
- Multi-genre configurations
- Entity death detection
- Loot item lifecycle
- Rarity distribution
- Thread safety (race detector clean)

Run tests with:
```bash
go test -race -cover ./pkg/loot/...
```

## Architecture Notes

The loot system follows these design principles:

1. **Separation of Concerns**: LootTable handles drop logic, LootDropSystem handles ECS integration
2. **Deterministic by Default**: All randomness is seeded and reproducible
3. **Genre-Aware**: Loot adapts to game setting automatically
4. **Integration First**: Designed to work seamlessly with existing ECS systems
5. **Performance Critical**: Optimized for thousands of entities

## Future Enhancements

Potential extensions (not yet implemented):

- Item quality/modifier rolling (e.g., +5 sword)
- Conditional drops based on player stats or quest state
- Group loot distribution in multiplayer
- Loot magnet/auto-pickup mechanics
- Visual effects integration for loot spawns
- Loot ownership/tagging for multiplayer fairness

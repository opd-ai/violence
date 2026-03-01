# Economy Balance Guide

## Overview

The Violence credit economy is designed to create meaningful resource scarcity while allowing ~3 purchases per level. Credits are earned through combat, exploration, and objectives, then spent at vendors for weapons, items, and upgrades.

## Design Philosophy

- **Scarcity with Agency**: Players choose between hoarding for powerful purchases or buying consumables frequently
- **Risk-Reward Balance**: Higher difficulties yield more credits but increase resource consumption
- **Progressive Scaling**: Later levels offer better loot but require stronger equipment
- **Genre Variation**: Each genre has distinct economic balance reflecting its thematic feel

## Credit Reward Table

### Combat Rewards

| Enemy Archetype | Base Credits | Difficulty Multiplier |
|----------------|--------------|----------------------|
| Weak (FantasyGuard, PostapocScavenger) | 10-15 | Easy: 0.8x, Normal: 1.0x, Hard: 1.25x, Nightmare: 1.5x |
| Medium (SciFiSoldier, CyberpunkDrone) | 20-30 | Easy: 0.8x, Normal: 1.0x, Hard: 1.25x, Nightmare: 1.5x |
| Strong (HorrorCultist, Elite Variants) | 40-60 | Easy: 0.8x, Normal: 1.0x, Hard: 1.25x, Nightmare: 1.5x |
| Boss (Level Guardian) | 150-250 | Easy: 0.8x, Normal: 1.0x, Hard: 1.25x, Nightmare: 1.5x |

### Exploration Rewards

| Discovery Type | Credits | Notes |
|---------------|---------|-------|
| Secret Area Found | 25-50 | One-time per secret |
| Hidden Cache | 30-80 | Procedurally placed 1-3 per level |
| Map Completion Bonus | 50 | 100% room exploration |
| Automap Discovery | 5 | Per 10% map revealed |

### Objective Rewards

| Objective Type | Credits | Notes |
|---------------|---------|-------|
| Primary Objective Complete | 100 | Main level goal |
| Secondary Objective | 40-60 | Optional side tasks |
| Time Bonus (Speed Run) | 25-75 | Complete level <10 min |
| No Damage Bonus | 100 | Complete level with 100% health |
| Pacifist Bonus | 150 | Complete level avoiding combat (when possible) |

### Expected Credits Per Level

| Difficulty | Average Credits | Min (Rushed) | Max (100% Completion) |
|-----------|----------------|--------------|----------------------|
| Easy | 280 | 180 | 450 |
| Normal | 350 | 225 | 550 |
| Hard | 438 | 280 | 690 |
| Nightmare | 525 | 340 | 825 |

**Target**: ~350 credits/level on Normal allows 3-4 medium purchases or 1 major + 1 minor purchase.

## Item Price Table

### Weapons

| Weapon Type | Base Price | Genre Variants |
|------------|-----------|----------------|
| Melee Starter | 0 | Always available at spawn |
| Melee Upgraded | 150 | Fantasy: Enchanted Blade, SciFi: Plasma Cutter, Horror: Ritual Dagger, Cyberpunk: Mono-Blade, Postapoc: Chainsaw |
| Hitscan Basic | 100 | Fantasy: Crossbow, SciFi: Laser Pistol, Horror: Revolver, Cyberpunk: Smart Pistol, Postapoc: Pipe Rifle |
| Hitscan Advanced | 250 | Fantasy: Arcane Rifle, SciFi: Plasma Rifle, Horror: Shotgun, Cyberpunk: Rail Gun, Postapoc: Scrap Cannon |
| Projectile Basic | 120 | Fantasy: Fireball Staff, SciFi: Bolt Caster, Horror: Curse Launcher, Cyberpunk: Grenade Launcher, Postapoc: Molotov Thrower |
| Projectile Advanced | 280 | Fantasy: Lightning Rod, SciFi: Rocket Launcher, Horror: Soul Reaper, Cyberpunk: Missile Pod, Postapoc: Nuke Launcher |

### Consumables

| Item | Price | Effect | Stack Limit |
|------|-------|--------|-------------|
| Health Pack (Small) | 25 | +25 HP | 5 |
| Health Pack (Medium) | 50 | +50 HP | 5 |
| Health Pack (Large) | 100 | +100 HP (full heal) | 3 |
| Armor Shard | 30 | +20 Armor | 5 |
| Armor Vest | 80 | +50 Armor | 3 |
| Ammo Box (Basic) | 20 | +50 rounds | 10 |
| Ammo Box (Advanced) | 40 | +30 rounds (advanced weapons) | 5 |
| Grenade | 35 | Explosive damage AoE | 5 |
| Flashbang | 30 | Stun enemies 3s | 3 |
| Smoke Bomb | 25 | Concealment 5s | 3 |

### Upgrades (Permanent)

| Upgrade | Price | Effect | Stackable |
|---------|-------|--------|----------|
| Health Boost I | 100 | +25 Max HP | Yes (3x max) |
| Health Boost II | 200 | +50 Max HP | Yes (2x max) |
| Armor Boost I | 100 | +25 Max Armor | Yes (3x max) |
| Armor Boost II | 200 | +50 Max Armor | Yes (2x max) |
| Move Speed Boost | 150 | +10% movement | Yes (2x max) |
| Damage Boost | 200 | +15% weapon damage | Yes (2x max) |
| Reload Speed | 120 | -20% reload time | Yes (2x max) |
| Inventory Expansion | 80 | +2 item slots | Yes (5x max) |
| Flashlight Range | 90 | +30% light range | No |
| Map Reveal Radius | 70 | +50% automap discovery | No |

## Genre-Specific Economy Balance

### Fantasy Genre
- **Flavor**: Gold coins, bartering with merchants
- **Vendor Locations**: Taverns, blacksmith forges
- **Economic Feel**: Moderate scarcity, balanced combat/exploration rewards
- **Vendor Multiplier**: 1.0x (baseline prices)

### SciFi Genre
- **Flavor**: Credits, automated kiosks
- **Vendor Locations**: Medical bays, armories, supply depots
- **Economic Feel**: Generous credits, expensive advanced tech
- **Vendor Multiplier**: 0.9x (slight discount on basic items, premium on advanced)

### Horror Genre
- **Flavor**: Blood currency, forbidden relics
- **Vendor Locations**: Dark shrines, survivor camps
- **Economic Feel**: Extreme scarcity, high risk/reward
- **Vendor Multiplier**: 1.2x (20% markup to increase scarcity)

### Cyberpunk Genre
- **Flavor**: Crypto tokens, black market deals
- **Vendor Locations**: Neon markets, back alleys, data terminals
- **Economic Feel**: Moderate abundance, frequent small purchases
- **Vendor Multiplier**: 0.95x (slight discount to encourage spending)

### Postapocalyptic Genre
- **Flavor**: Scrap metal, bottlecaps, salvage
- **Vendor Locations**: Trader outposts, bunkers
- **Economic Feel**: Variable scarcity, heavy resource management
- **Vendor Multiplier**: 1.1x (10% markup reflecting harsh survival)

## Tuning Guidelines

### Achieving ~3 Purchases Per Level

**Target Budget**: 350 credits (Normal difficulty)

**Sample Loadouts**:
1. **Consumable Focus**: 3x Medium Health Pack (150cr) + 2x Ammo Box (40cr) + 1x Armor Vest (80cr) = 270cr, leaves 80cr buffer
2. **Weapon Upgrade**: 1x Hitscan Advanced (250cr) + 2x Small Health Pack (50cr) = 300cr, leaves 50cr buffer
3. **Balanced**: 1x Hitscan Basic (100cr) + 1x Armor Vest (80cr) + 3x Ammo Box (60cr) + 2x Grenade (70cr) = 310cr, leaves 40cr buffer
4. **Permanent Upgrade**: 1x Health Boost I (100cr) + 1x Damage Boost (200cr) = 300cr, leaves 50cr for consumables

### Balancing Feedback Loops

**If players hoard excessively** (>1000 credits banked):
- Reduce vendor prices 10-15%
- Increase high-tier weapon availability
- Add credit decay mechanic (optional, genre-specific)

**If players constantly broke** (<50 credits average):
- Increase combat rewards 15-20%
- Add more hidden caches
- Reduce consumable prices 10%
- Increase ammo drop rates (reduce ammo purchases)

**If economy feels pointless** (too easy to buy everything):
- Increase prices 20%
- Reduce credit rewards 15%
- Add more expensive legendary items
- Increase upgrade stack limits

### Progression Scaling (Multi-Level)

| Level Range | Price Scaling | Reward Scaling | Reasoning |
|-------------|---------------|----------------|-----------|
| 1-3 | 1.0x | 1.0x | Learning phase, stable economy |
| 4-6 | 1.15x | 1.2x | Mid-game, rewards outpace inflation slightly |
| 7-9 | 1.3x | 1.45x | Late-game, access to best gear |
| 10+ | 1.5x | 1.7x | End-game, power fantasy enabled |

**Implementation**: Multiply base prices/rewards by level-based scalar in vendor/loot systems.

### Difficulty Modifiers

| Difficulty | Credit Multiplier | Price Multiplier | Net Effect |
|-----------|------------------|------------------|------------|
| Easy | 0.8x | 0.9x | 12% less purchasing power (forgiving) |
| Normal | 1.0x | 1.0x | Baseline |
| Hard | 1.25x | 1.05x | 19% more purchasing power (need better gear) |
| Nightmare | 1.5x | 1.1x | 36% more purchasing power (essential for survival) |

**Reasoning**: Higher difficulties require better equipment, so reward multipliers outpace price increases.

## Testing and Validation

### Playtesting Metrics

Track these metrics during playtests:

1. **Average Credits/Level**: Should be ~350 on Normal
2. **Credits Banked**: Should stay <500 (not hoarding)
3. **Purchases/Level**: Target 2-4 transactions
4. **Vendor Visit Frequency**: Every 1-2 levels
5. **Item Type Distribution**: 60% consumables, 30% weapons, 10% upgrades

### Balance Adjustment Process

1. **Gather Data**: Log 20+ playthroughs across all difficulties
2. **Identify Outliers**: Flag levels with >50% deviation from target
3. **Isolate Variables**: Check if issue is rewards, prices, or difficulty scaling
4. **Incremental Changes**: Adjust by 10-15% max per iteration
5. **Re-test**: Validate changes don't break other difficulty tiers

### Red Flags

- **Player always buys same item**: Price too low or item too strong
- **Player never buys item**: Price too high or item useless
- **Credits cap consistently**: Increase high-value purchase options
- **Player dies frequently from resource starvation**: Increase consumable availability

## Implementation Notes

### Code Integration

Economy values should be centralized in configuration:

```go
// pkg/economy/config.go
type EconomyConfig struct {
    GenreMultiplier    float64
    DifficultyMultiplier float64
    LevelScalar        float64
    ItemPrices         map[string]int
    RewardRanges       map[string][2]int
}
```

### Data Files

Consider externalizing economy tables to TOML/JSON for live-tuning without recompilation:

```toml
[economy.rewards.combat]
weak_enemy = [10, 15]
medium_enemy = [20, 30]
strong_enemy = [40, 60]
boss_enemy = [150, 250]

[economy.prices.weapons]
melee_upgraded = 150
hitscan_basic = 100
hitscan_advanced = 250
```

### Runtime Adjustments

Implement analytics hooks:

```go
func RecordPurchase(playerID string, item string, price int)
func RecordCreditsEarned(playerID string, source string, amount int)
func GetAverageCreditsPerLevel() float64
func GetPurchaseFrequency() map[string]int
```

## Conclusion

This economy is designed for **meaningful scarcity** with **player agency**. The ~3 purchases/level target ensures players constantly make interesting choices without feeling starved. Genre and difficulty multipliers provide variety while maintaining balance.

**Next Steps**:
1. Implement economy config system in `pkg/economy/`
2. Add vendor transaction logging
3. Conduct 50+ hour playtesting across all genres/difficulties
4. Iterate based on telemetry data
5. Document post-launch tuning in CHANGELOG.md

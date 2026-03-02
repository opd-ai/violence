# Territory Control System

## Overview

The territory control system adds dynamic faction warfare and territorial conflicts to the Violence game. It divides the procedurally generated dungeon into controllable zones that factions can contest and capture, creating emergent gameplay through:

- **Territory Ownership**: Factions control specific rooms/zones in the dungeon
- **Dynamic Patrols**: Controlled territories spawn faction patrols
- **Battle Fronts**: Contested territories become active combat zones
- **Reinforcements**: Defending factions spawn reinforcements when territories are threatened
- **Control Transfer**: Territories change hands through sustained combat

## Architecture

### Components

1. **TerritoryComponent**: Marks entities within a specific territory
2. **PatrolComponent**: Controls patrol AI movement along predefined routes
3. **ReinforcementComponent**: Marks entities as temporary reinforcements
4. **PositionComponent**: Entity position for territory assignment

### System Integration

The `ControlSystem` integrates with:
- **Faction System**: Uses faction relationships for territory ownership
- **Combat System**: Tracks faction members via HealthComponent
- **Spatial Grid**: Efficient territory lookup via grid partitioning
- **BSP Generation**: Claims rooms as territories during level creation

## Usage

### Initialization

```go
// In NewGame() - main.go
g.factionSystem = faction.NewReputationSystem()
g.territorySystem = territory.NewControlSystem(64, 64, g.factionSystem)
g.world.AddSystem(g.territorySystem)
```

### Territory Claiming

```go
// During level population - claimTerritories()
activeFactions := g.factionSystem.GetActiveFactions(g.genreID)
for i := 1; i < len(rooms); i++ {
    factionID := activeFactions[rng.Intn(len(activeFactions))].ID
    g.territorySystem.ClaimRoom(room, factionID)
}
```

## Gameplay Impact

### Territory Control Mechanics

1. **Initial Ownership**: Rooms are assigned to factions during level generation
2. **Contestation**: When opposing faction members enter a territory, it becomes contested
3. **Control Points**: Contested territories lose control points when outnumbered
4. **Transfer**: At 0 control points, the territory transfers to the dominant faction
5. **Stabilization**: Uncontested territories slowly regenerate control points

### Patrol Spawning

- Territories spawn 0-2 patrols when uncontested
- Patrols follow randomized routes within their territory
- Spawn rate: ~1% chance per update (60 updates/sec) when conditions met
- Cooldown: 15 seconds after last battle before new patrols spawn

### Reinforcement System

- Reinforcements spawn when territories are contested
- 30-second cooldown between reinforcement waves
- Reinforcements are stronger than regular patrols (150 HP vs 100 HP)
- ~2% chance per update when territory is contested

### Battle Fronts

Battle fronts are territories actively being contested:
- Higher enemy density
- Continuous reinforcement spawning
- Potential for control transfer
- Strategic importance for faction dominance

## Performance Characteristics

### Spatial Partitioning

- 8x8 grid cells for territory lookup
- O(1) territory assignment per entity
- Minimal memory overhead per territory

### Update Costs

- Territory assignment: O(n) where n = entities with PositionComponent
- Contestation processing: O(m) where m = entities in contested territories
- Patrol spawning: O(k) where k = number of territories
- Patrol movement: O(p) where p = active patrols

## Observing in Game

1. **Territory Ownership**: Each room is controlled by a faction (Mercenaries, Rebels, Cult, etc.)
2. **Patrol Movement**: Watch for faction patrols following routes within rooms
3. **Battle Fronts**: Areas with multiple factions present become contested zones
4. **Reinforcements**: New enemies spawn when you enter contested territories
5. **Control Transfer**: Territories change ownership after sustained combat

## Future Enhancements

Potential expansions:
- Visual territory ownership indicators on automap
- Territory bonuses (resource generation, spawn rates)
- Strategic objectives (capture X territories for quest completion)
- Territory-based multiplayer modes (king of the hill, domination)
- Faction AI that actively seeks to expand territory
- Environmental changes based on controlling faction (decorations, lighting)

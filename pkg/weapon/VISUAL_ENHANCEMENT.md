# Weapon Visual Enhancement System

## Overview
The weapon visual enhancement system provides rich, material-based procedural weapon rendering with support for rarity tiers, damage states, and magical enchantments.

## Features

### Material System
Weapons are rendered with realistic material properties:
- **Metals**: Steel, Iron, Gold, Mithril, Obsidian - with metallic sheen and edge highlights
- **Organic**: Wood (with grain), Leather (with wrapping texture), Bone
- **Magical**: Crystal, Demonic - with unique visual properties

Each material has distinct base, highlight, and shadow colors, plus material-specific rendering techniques:
- Wood shows procedural grain patterns and occasional knots
- Leather displays diagonal wrapping with stitch marks
- Metals feature specular highlights and beveled edges

### Rarity Tiers
Five rarity levels with increasing visual complexity:
- **Common**: Basic weapon appearance
- **Uncommon**: Subtle edge glow
- **Rare**: Enhanced glow + pommel jewel
- **Epic**: Stronger effects + more energy coils (projectile weapons)
- **Legendary**: Maximum visual flair + golden accent

### Damage States
Weapons visually degrade based on durability:
- **Pristine** (≥80%): Clean, no wear
- **Scratched** (50-80%): Small surface scratches
- **Worn** (20-50%): Significant darkening, rust/tarnish spots
- **Broken** (<20%): Visible cracks, severe damage

### Enchantments
Six enchantment types with matching glow colors:
- **Fire**: Orange-red flames and muzzle flash
- **Ice**: Pale blue frost
- **Lightning**: Electric blue arcs
- **Poison**: Toxic green glow
- **Holy**: Golden radiance
- **Shadow**: Purple-black aura

Enchantments add:
- Subtle glow on weapon surface
- Sparkle particles during idle
- Enhanced muzzle flash/charge effects during fire frame

### Frame States
Three animation frames supported:
- **Idle**: Standard weapon appearance
- **Fire**: Muzzle flash, enchantment particles, energy charge
- **Reload**: (Future expansion)

## API

### Creating Enhanced Weapons
```go
import "github.com/opd-ai/violence/pkg/weapon"

// Define weapon visual spec
spec := weapon.WeaponVisualSpec{
    Type:        weapon.TypeMelee,
    Frame:       weapon.FrameIdle,
    Rarity:      weapon.RarityEpic,
    Damage:      weapon.DamagePristine,
    BladeMat:    weapon.MaterialMithril,
    HandleMat:   weapon.MaterialLeather,
    Seed:        12345,
    Enchantment: "lightning",
}

// Generate sprite
img := weapon.EnhancedGenerateWeaponSprite(spec)
```

### ECS Integration
The system integrates with the game's ECS architecture:

```go
// Add component to entity
vc := weapon.NewVisualComponent(weapon.TypeMelee, seed)
vc.SetRarity(weapon.RarityLegendary)
vc.SetMaterials(weapon.MaterialGold, weapon.MaterialLeather)
vc.SetEnchantment("fire")

// Component automatically regenerates sprite on changes
sprite := vc.GetSprite()

// Update damage state based on durability
weaponVisualSystem.UpdateWeaponDamage(vc, durability)
```

### Rendering
```go
// Simple rendering
weaponVisualSystem.RenderWeapon(screen, visualComp, x, y, scale)

// With rotation (for equipped weapons)
weaponVisualSystem.RenderWeaponWithRotation(
    screen, visualComp, x, y, rotation, scale)
```

## Performance

### Caching
- Sprites are cached per component and only regenerated when properties change
- Frame changes trigger regeneration (muzzle flash differs between idle/fire)
- Deterministic generation - same seed + spec = identical sprite

### Memory
- LRU sprite cache prevents unbounded memory growth
- Image buffers pooled by size for allocation efficiency
- Sprites automatically disposed when no longer cached

## Technical Details

### Shading Algorithm
Materials use gradient-based shading:
1. Calculate distance from geometric center (0-1 normalized)
2. Blend highlight/base/shadow colors based on distance
3. Apply depth-based darkening for 3D effect
4. Add material-specific details (grain, highlights, etc.)

### Weapon Types
- **Melee**: Tapered blade with crossguard and pommel
- **Hitscan**: Gun with barrel, receiver, grip, and trigger
- **Projectile**: Launcher with energy coils and charge effects

All types support full material and enchantment customization.

## Usage in Game
The weapon visual system is registered in `main.go` and updates automatically each frame. When weapons are:
- Picked up/equipped → Generate with appropriate rarity/materials
- Used in combat → Switch to fire frame for muzzle effects
- Damaged → Update damage state based on durability
- Enchanted → Add magical glow and particle effects

Sprites can be used in:
- Player HUD (equipped weapon display)
- Inventory UI (item icons)
- Character sprites (visible equipped weapons)
- Loot drops (on-ground item representation)

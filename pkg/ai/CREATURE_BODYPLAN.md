# Creature Body-Plan System

This enhancement adds dedicated anatomical templates for non-humanoid creatures, dramatically improving visual variety and readability in the game.

## What's New

Previously, all enemies used humanoid body templates (bipedal stance with two arms and two legs). This made spiders, snakes, slimes, and flying creatures visually indistinguishable from human guards and soldiers.

Now, the system includes **five distinct body plans**:

### 1. Quadrupeds (BodyPlanQuadruped)
Four-legged creatures with horizontal body orientation.
- **Wolf**: Gray fur, yellow eyes, pointed ears
- **Bear**: Brown, larger size, rounded ears
- **Lion**: Tawny coat, golden eyes, prominent mane
- **Hound**: Tan/brown hunting dog
- **Raptor**: Green scales, reptilian features

### 2. Insects (BodyPlanInsect)
Multi-legged arthropods with segmented bodies.
- **Spider**: 8 legs radiating from body, multiple red eyes
- **Beetle**: 6 legs, hard shell, green carapace
- **Mantis**: Elongated body, raptorial forelegs for attacks
- **Scorpion**: 8 legs, segmented tail with stinger
- **Ant**: 6 legs, compact segmented body

### 3. Serpents (BodyPlanSerpent)
Legless elongated creatures with serpentine motion.
- **Snake**: Green scales with belly stripe, forked tongue
- **Worm**: Thick-bodied, earth tones
- **Serpent**: Purple mystical variant
- **Lamia**: Hybrid serpent with upper humanoid features

### 4. Flying Creatures (BodyPlanFlying)
Winged entities with aerial poses.
- **Bat**: Large ears, membranous wings
- **Drake**: Horned dragon-like, larger wingspan
- **Harpy**: Feathered wings, talon feet
- **Wasp**: Striped abdomen, stinger, translucent wings

### 5. Amorphous (BodyPlanAmorphous)
Formless or semi-fluid entities.
- **Slime**: Green gelatinous blob, pulsating animation
- **Ooze**: Purple semi-transparent
- **Elemental**: Fire-like tendrils, orange glow
- **Wraith**: Ghostly wispy trails, ethereal appearance

## Usage

```go
import "github.com/opd-ai/violence/pkg/ai"

// Generate a spider sprite (8-legged insect body plan)
sprite := ai.GenerateCreatureSprite(seed, ai.CreatureSpider, ai.AnimFrameIdle)

// Check the body plan of a creature type
bodyPlan := ai.GetBodyPlan(ai.CreatureBat) // Returns BodyPlanFlying
```

## Animation Support

All body plans support standard animation frames:
- **AnimFrameIdle**: Standing/hovering/resting pose
- **AnimFrameWalk1/Walk2**: Movement cycle (leg motion, wing flap, slither, etc.)
- **AnimFrameAttack**: Attack pose (lunge, strike, extend stinger, etc.)
- **AnimFrameDeath**: Death animation

Each body plan adapts animation appropriately:
- Quadrupeds alternate leg pairs
- Insects shift leg positions radially
- Serpents undulate along S-curve
- Flying creatures flap wings
- Amorphous entities pulse and wobble

## Visual Benefits

1. **Instant Recognition**: Players can immediately identify creature type from silhouette
2. **Tactical Clarity**: Different body plans signal different behaviors and attack patterns
3. **Procedural Variety**: Seed-based generation within each type ensures no two spiders look identical
4. **Genre Appropriateness**: Fantasy dungeons have wolves and spiders; sci-fi has drones and elementals
5. **Combat Readability**: In chaotic multi-enemy fights, shape variety prevents visual blur

## Technical Details

- **Deterministic**: Same seed + creature type + frame = identical sprite
- **Lightweight**: Pure procedural generation, no asset files
- **Composable**: Body parts rendered in logical z-order (legs behind body, head on top)
- **Shaded**: Each creature uses appropriate color gradients and highlights for depth
- **Bounded**: All sprites fit within 64x64 canvas with consistent scaling

## Example Output

Run the showcase to generate sample sprites:
```bash
go run pkg/ai/examples/creature_showcase.go
```

This creates PNG files in `/tmp/` showing all creature types.

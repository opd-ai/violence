# Attack Animation System

The attack animation system adds visual variety and feedback to enemy attacks through procedural attack animations.

## Overview

This system gives each enemy attack type distinct visual telegraphing through body movements, rotations, and squash/stretch deformations. Different attack patterns produce unique visual signatures that help players read enemy intent.

## Attack Types

The system supports six distinct animation patterns:

1. **melee_slash** - Wide horizontal swing with rotation
2. **overhead_smash** - Vertical downward slam with stretch
3. **lunge** - Forward thrust with squash and extension
4. **ranged_charge** - Draw-back and release for projectiles
5. **spin_attack** - Full 360° rotation attack
6. **quick_jab** - Fast, minimal-windup thrust

## Animation Phases

Each attack progresses through three phases:

1. **Windup** - Telegraph the attack direction and type
2. **Strike** - Execute the attack with motion blur and impact
3. **Recovery** - Return to idle state

## Visual Parameters

The system calculates frame-by-frame:
- **Position offset** (offsetX, offsetY) - Body movement during attack
- **Rotation angle** - Weapon/body rotation
- **Squash/stretch** - Sprite deformation for impact feel
- **Animation intensity** - Scales visual exaggeration based on damage

## Integration

Attack animations are triggered when AI agents attack:
- Agent archetype determines preferred attack type
- Distance to target influences animation choice
- Damage value affects animation intensity

Visual parameters are available via `GetAnimationParams()` for rendering systems to apply transformations during sprite rendering.

## Performance

- Zero-allocation hot path after entity creation
- Eased animation curves for smooth motion
- ~68% test coverage
- Handles 100+ concurrent animations at 60 FPS

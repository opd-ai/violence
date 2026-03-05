// Package playersprite provides player and NPC sprite generation with equipment visibility.
package playersprite

import "github.com/hajimehoshi/ebiten/v2"

// Component stores player sprite visual state.
type Component struct {
	Class          string         // Player class (warrior, mage, rogue, tech, soldier, etc.)
	EquippedWeapon string         // Weapon type for visual rendering
	EquippedArmor  string         // Armor type for visual rendering
	CurrentFrame   int            // Animation frame index
	AnimState      AnimationState // Current animation state
	Seed           int64          // Deterministic generation seed
	Facing         int            // Direction: 0=down, 1=right, 2=up, 3=left
	CachedSprite   *ebiten.Image  // Cached rendered sprite
	DirtyFlag      bool           // True if sprite needs regeneration
}

// Type implements engine.Component.
func (c *Component) Type() string {
	return "PlayerSprite"
}

// AnimationState represents player animation states.
type AnimationState int

const (
	AnimIdle AnimationState = iota
	AnimWalk
	AnimAttack
	AnimHurt
	AnimDeath
	AnimDodge
	AnimCast
)

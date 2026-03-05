// Package weapon provides weapon visual enhancement components.
package weapon

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// VisualComponent stores enhanced weapon visual state.
type VisualComponent struct {
	Spec          WeaponVisualSpec
	CachedSprite  *ebiten.Image
	NeedsRegen    bool
	LastFrameType FrameType
}

// Type implements the Component interface.
func (vc *VisualComponent) Type() string {
	return "WeaponVisual"
}

// NewVisualComponent creates a weapon visual component.
func NewVisualComponent(weaponType WeaponType, seed int64) *VisualComponent {
	spec := WeaponVisualSpec{
		Type:        weaponType,
		Frame:       FrameIdle,
		Rarity:      RarityCommon,
		Damage:      DamagePristine,
		BladeMat:    MaterialSteel,
		HandleMat:   MaterialWood,
		Seed:        seed,
		Enchantment: "",
	}

	return &VisualComponent{
		Spec:          spec,
		CachedSprite:  nil,
		NeedsRegen:    true,
		LastFrameType: FrameIdle,
	}
}

// SetRarity updates weapon rarity and marks for regeneration.
func (vc *VisualComponent) SetRarity(rarity Rarity) {
	if vc.Spec.Rarity != rarity {
		vc.Spec.Rarity = rarity
		vc.NeedsRegen = true
	}
}

// SetDamageState updates weapon condition and marks for regeneration.
func (vc *VisualComponent) SetDamageState(damage DamageState) {
	if vc.Spec.Damage != damage {
		vc.Spec.Damage = damage
		vc.NeedsRegen = true
	}
}

// SetEnchantment updates weapon enchantment and marks for regeneration.
func (vc *VisualComponent) SetEnchantment(enchantment string) {
	if vc.Spec.Enchantment != enchantment {
		vc.Spec.Enchantment = enchantment
		vc.NeedsRegen = true
	}
}

// SetMaterials updates blade and handle materials and marks for regeneration.
func (vc *VisualComponent) SetMaterials(blade, handle Material) {
	if vc.Spec.BladeMat != blade || vc.Spec.HandleMat != handle {
		vc.Spec.BladeMat = blade
		vc.Spec.HandleMat = handle
		vc.NeedsRegen = true
	}
}

// SetFrame updates animation frame type.
func (vc *VisualComponent) SetFrame(frame FrameType) {
	if vc.Spec.Frame != frame {
		vc.Spec.Frame = frame
		vc.LastFrameType = vc.Spec.Frame
		vc.NeedsRegen = true
	}
}

// GetSprite returns the current cached sprite, regenerating if needed.
func (vc *VisualComponent) GetSprite() *ebiten.Image {
	if vc.NeedsRegen || vc.CachedSprite == nil {
		vc.regenerateSprite()
	}
	return vc.CachedSprite
}

// regenerateSprite creates a new sprite from current spec.
func (vc *VisualComponent) regenerateSprite() {
	if vc.CachedSprite != nil {
		vc.CachedSprite.Dispose()
	}

	rgba := EnhancedGenerateWeaponSprite(vc.Spec)
	vc.CachedSprite = ebiten.NewImageFromImage(rgba)
	vc.NeedsRegen = false
}

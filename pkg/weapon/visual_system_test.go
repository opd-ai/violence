package weapon

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
)

func TestVisualComponent_Creation(t *testing.T) {
	vc := NewVisualComponent(TypeMelee, 12345)

	if vc == nil {
		t.Fatal("NewVisualComponent returned nil")
	}

	if vc.Type() != "WeaponVisual" {
		t.Errorf("expected type 'WeaponVisual', got '%s'", vc.Type())
	}

	if !vc.NeedsRegen {
		t.Error("new component should need regeneration")
	}

	if vc.Spec.Type != TypeMelee {
		t.Error("weapon type not set correctly")
	}

	if vc.Spec.Seed != 12345 {
		t.Error("seed not set correctly")
	}
}

func TestVisualComponent_SetRarity(t *testing.T) {
	vc := NewVisualComponent(TypeMelee, 12345)
	vc.NeedsRegen = false

	vc.SetRarity(RarityLegendary)

	if vc.Spec.Rarity != RarityLegendary {
		t.Error("rarity not updated")
	}

	if !vc.NeedsRegen {
		t.Error("SetRarity should mark for regeneration")
	}

	// Setting same rarity should not mark for regen
	vc.NeedsRegen = false
	vc.SetRarity(RarityLegendary)

	if vc.NeedsRegen {
		t.Error("SetRarity with same value should not mark for regeneration")
	}
}

func TestVisualComponent_SetDamageState(t *testing.T) {
	vc := NewVisualComponent(TypeHitscan, 54321)
	vc.NeedsRegen = false

	vc.SetDamageState(DamageWorn)

	if vc.Spec.Damage != DamageWorn {
		t.Error("damage state not updated")
	}

	if !vc.NeedsRegen {
		t.Error("SetDamageState should mark for regeneration")
	}
}

func TestVisualComponent_SetEnchantment(t *testing.T) {
	vc := NewVisualComponent(TypeMelee, 99999)
	vc.NeedsRegen = false

	vc.SetEnchantment("fire")

	if vc.Spec.Enchantment != "fire" {
		t.Error("enchantment not updated")
	}

	if !vc.NeedsRegen {
		t.Error("SetEnchantment should mark for regeneration")
	}
}

func TestVisualComponent_SetMaterials(t *testing.T) {
	vc := NewVisualComponent(TypeMelee, 77777)
	vc.NeedsRegen = false

	vc.SetMaterials(MaterialGold, MaterialLeather)

	if vc.Spec.BladeMat != MaterialGold {
		t.Error("blade material not updated")
	}

	if vc.Spec.HandleMat != MaterialLeather {
		t.Error("handle material not updated")
	}

	if !vc.NeedsRegen {
		t.Error("SetMaterials should mark for regeneration")
	}
}

func TestVisualComponent_SetFrame(t *testing.T) {
	vc := NewVisualComponent(TypeHitscan, 11111)
	vc.NeedsRegen = false

	vc.SetFrame(FrameFire)

	if vc.Spec.Frame != FrameFire {
		t.Error("frame not updated")
	}

	if vc.LastFrameType != FrameFire {
		t.Error("last frame type not updated")
	}

	if !vc.NeedsRegen {
		t.Error("SetFrame should mark for regeneration")
	}
}

func TestVisualComponent_GetSprite(t *testing.T) {
	vc := NewVisualComponent(TypeMelee, 12345)

	sprite := vc.GetSprite()

	if sprite == nil {
		t.Fatal("GetSprite returned nil")
	}

	if vc.NeedsRegen {
		t.Error("GetSprite should clear NeedsRegen flag")
	}

	if vc.CachedSprite == nil {
		t.Error("GetSprite should populate CachedSprite")
	}

	// Getting sprite again should return same cached instance
	sprite2 := vc.GetSprite()
	if sprite != sprite2 {
		t.Error("GetSprite should return cached sprite when not marked for regen")
	}
}

func TestVisualComponent_Regeneration(t *testing.T) {
	vc := NewVisualComponent(TypeMelee, 12345)

	// Get initial sprite
	sprite1 := vc.GetSprite()

	// Change spec and verify regeneration
	vc.SetRarity(RarityEpic)
	sprite2 := vc.GetSprite()

	// Sprites should be different instances after regeneration
	if sprite1 == sprite2 {
		t.Error("regeneration should create new sprite instance")
	}
}

func TestVisualSystem_Update(t *testing.T) {
	sys := NewVisualSystem()
	vc := NewVisualComponent(TypeMelee, 12345)

	// Create a real engine.World and entity
	w := engine.NewWorld()

	ent := w.AddEntity()
	w.AddComponent(ent, vc)

	// Mark for regeneration
	vc.NeedsRegen = true

	// Update should trigger regeneration
	sys.Update(w)

	if vc.NeedsRegen {
		t.Error("Update should regenerate sprites marked for regen")
	}

	if vc.CachedSprite == nil {
		t.Error("Update should populate cached sprite")
	}
}

func TestVisualSystem_UpdateWithoutComponent(t *testing.T) {
	sys := NewVisualSystem()

	// Create world with entity but no weapon visual component
	w := engine.NewWorld()
	_ = w.AddEntity()

	// Should not crash with missing component
	sys.Update(w)
}

func TestVisualSystem_RenderWeapon(t *testing.T) {
	sys := NewVisualSystem()
	vc := NewVisualComponent(TypeMelee, 12345)

	screen := ebiten.NewImage(800, 600)
	defer screen.Dispose()

	// Should not crash
	sys.RenderWeapon(screen, vc, 400, 300, 1.0)

	// Test with nil component
	sys.RenderWeapon(screen, nil, 400, 300, 1.0)
}

func TestVisualSystem_RenderWeaponWithRotation(t *testing.T) {
	sys := NewVisualSystem()
	vc := NewVisualComponent(TypeHitscan, 54321)

	screen := ebiten.NewImage(800, 600)
	defer screen.Dispose()

	// Should not crash
	sys.RenderWeaponWithRotation(screen, vc, 400, 300, 0.5, 1.5)

	// Test with nil component
	sys.RenderWeaponWithRotation(screen, nil, 400, 300, 0.5, 1.5)
}

func TestVisualSystem_UpdateWeaponDamage(t *testing.T) {
	sys := NewVisualSystem()
	vc := NewVisualComponent(TypeMelee, 12345)

	tests := []struct {
		name       string
		durability float64
		wantState  DamageState
	}{
		{"pristine high durability", 1.0, DamagePristine},
		{"pristine low threshold", 0.8, DamagePristine},
		{"scratched", 0.7, DamageScratched},
		{"worn", 0.4, DamageWorn},
		{"broken", 0.1, DamageBroken},
		{"broken zero", 0.0, DamageBroken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys.UpdateWeaponDamage(vc, tt.durability)

			if vc.Spec.Damage != tt.wantState {
				t.Errorf("durability %.2f: expected %v, got %v",
					tt.durability, tt.wantState, vc.Spec.Damage)
			}
		})
	}

	// Test with nil component (should not crash)
	sys.UpdateWeaponDamage(nil, 0.5)
}

func TestVisualComponent_FullWorkflow(t *testing.T) {
	// Create component
	vc := NewVisualComponent(TypeMelee, 12345)

	// Configure weapon
	vc.SetRarity(RarityLegendary)
	vc.SetMaterials(MaterialMithril, MaterialLeather)
	vc.SetEnchantment("lightning")

	// Get sprite
	sprite := vc.GetSprite()
	if sprite == nil {
		t.Fatal("failed to generate sprite")
	}

	// Simulate weapon taking damage
	vc.SetDamageState(DamageScratched)

	// Sprite should regenerate
	sprite2 := vc.GetSprite()
	if sprite == sprite2 {
		t.Error("damage change should regenerate sprite")
	}

	// Change to fire frame
	vc.SetFrame(FrameFire)
	sprite3 := vc.GetSprite()
	if sprite2 == sprite3 {
		t.Error("frame change should regenerate sprite")
	}

	// Verify all settings preserved
	if vc.Spec.Rarity != RarityLegendary {
		t.Error("rarity not preserved")
	}
	if vc.Spec.BladeMat != MaterialMithril {
		t.Error("blade material not preserved")
	}
	if vc.Spec.Enchantment != "lightning" {
		t.Error("enchantment not preserved")
	}
}

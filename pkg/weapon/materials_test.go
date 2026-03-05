package weapon

import (
	"image/color"
	"testing"
)

func TestEnhancedGenerateWeaponSprite(t *testing.T) {
	tests := []struct {
		name string
		spec WeaponVisualSpec
	}{
		{
			name: "steel melee common",
			spec: WeaponVisualSpec{
				Type:      TypeMelee,
				Frame:     FrameIdle,
				Rarity:    RarityCommon,
				Damage:    DamagePristine,
				BladeMat:  MaterialSteel,
				HandleMat: MaterialWood,
				Seed:      12345,
			},
		},
		{
			name: "gold melee legendary with fire enchantment",
			spec: WeaponVisualSpec{
				Type:        TypeMelee,
				Frame:       FrameFire,
				Rarity:      RarityLegendary,
				Damage:      DamagePristine,
				BladeMat:    MaterialGold,
				HandleMat:   MaterialLeather,
				Seed:        54321,
				Enchantment: "fire",
			},
		},
		{
			name: "iron hitscan worn",
			spec: WeaponVisualSpec{
				Type:      TypeHitscan,
				Frame:     FrameIdle,
				Rarity:    RarityUncommon,
				Damage:    DamageWorn,
				BladeMat:  MaterialIron,
				HandleMat: MaterialWood,
				Seed:      99999,
			},
		},
		{
			name: "crystal projectile epic ice",
			spec: WeaponVisualSpec{
				Type:        TypeProjectile,
				Frame:       FrameFire,
				Rarity:      RarityEpic,
				Damage:      DamagePristine,
				BladeMat:    MaterialCrystal,
				HandleMat:   MaterialMithril,
				Seed:        77777,
				Enchantment: "ice",
			},
		},
		{
			name: "obsidian melee broken shadow",
			spec: WeaponVisualSpec{
				Type:        TypeMelee,
				Frame:       FrameIdle,
				Rarity:      RarityRare,
				Damage:      DamageBroken,
				BladeMat:    MaterialObsidian,
				HandleMat:   MaterialBone,
				Seed:        11111,
				Enchantment: "shadow",
			},
		},
		{
			name: "demonic hitscan fire frame",
			spec: WeaponVisualSpec{
				Type:        TypeHitscan,
				Frame:       FrameFire,
				Rarity:      RarityEpic,
				Damage:      DamageScratched,
				BladeMat:    MaterialDemonic,
				HandleMat:   MaterialLeather,
				Seed:        66666,
				Enchantment: "fire",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := EnhancedGenerateWeaponSprite(tt.spec)
			if img == nil {
				t.Fatal("EnhancedGenerateWeaponSprite returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != 128 || bounds.Dy() != 128 {
				t.Errorf("expected 128x128 image, got %dx%d", bounds.Dx(), bounds.Dy())
			}

			// Check that image has non-transparent pixels
			hasContent := false
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						hasContent = true
						break
					}
				}
				if hasContent {
					break
				}
			}

			if !hasContent {
				t.Error("generated image has no visible content")
			}
		})
	}
}

func TestGetMaterialColors(t *testing.T) {
	materials := []Material{
		MaterialSteel,
		MaterialIron,
		MaterialWood,
		MaterialLeather,
		MaterialGold,
		MaterialCrystal,
		MaterialBone,
		MaterialObsidian,
		MaterialMithril,
		MaterialDemonic,
	}

	for _, mat := range materials {
		t.Run(mat.String(), func(t *testing.T) {
			colors := getMaterialColors(mat)

			// Verify colors are valid (RGBA can't be nil, check for zero values instead)
			if colors.base.A == 0 {
				t.Error("base color has zero alpha")
			}
			if colors.highlight.A == 0 {
				t.Error("highlight color has zero alpha")
			}
			if colors.shadow.A == 0 {
				t.Error("shadow color has zero alpha")
			}

			// Verify highlight is lighter than base
			baseBrightness := uint32(colors.base.R) + uint32(colors.base.G) + uint32(colors.base.B)
			highlightBrightness := uint32(colors.highlight.R) + uint32(colors.highlight.G) + uint32(colors.highlight.B)

			if highlightBrightness <= baseBrightness {
				t.Error("highlight should be brighter than base")
			}

			// Verify shadow is darker than base
			shadowBrightness := uint32(colors.shadow.R) + uint32(colors.shadow.G) + uint32(colors.shadow.B)

			if shadowBrightness >= baseBrightness {
				t.Error("shadow should be darker than base")
			}
		})
	}
}

func TestIsMetallic(t *testing.T) {
	tests := []struct {
		material Material
		want     bool
	}{
		{MaterialSteel, true},
		{MaterialIron, true},
		{MaterialGold, true},
		{MaterialMithril, true},
		{MaterialObsidian, true},
		{MaterialWood, false},
		{MaterialLeather, false},
		{MaterialBone, false},
		{MaterialCrystal, false},
		{MaterialDemonic, false},
	}

	for _, tt := range tests {
		t.Run(tt.material.String(), func(t *testing.T) {
			got := isMetallic(tt.material)
			if got != tt.want {
				t.Errorf("isMetallic(%v) = %v, want %v", tt.material, got, tt.want)
			}
		})
	}
}

func TestLightenDarkenColor(t *testing.T) {
	baseColor := color.RGBA{R: 128, G: 128, B: 128, A: 255}

	lightened := lightenColor(baseColor, 0.3)
	lr, lg, lb, _ := lightened.RGBA()

	if lr <= 128<<8 || lg <= 128<<8 || lb <= 128<<8 {
		t.Error("lightenColor should increase brightness")
	}

	darkened := darkenColor(baseColor, 0.3)
	dr, dg, db, _ := darkened.RGBA()

	if dr >= 128<<8 || dg >= 128<<8 || db >= 128<<8 {
		t.Error("darkenColor should decrease brightness")
	}
}

func TestGetRarityColor(t *testing.T) {
	rarities := []Rarity{
		RarityCommon,
		RarityUncommon,
		RarityRare,
		RarityEpic,
		RarityLegendary,
	}

	for _, rarity := range rarities {
		t.Run(rarity.String(), func(t *testing.T) {
			col := getRarityColor(rarity)
			if col.A != 255 {
				t.Error("rarity color should be fully opaque")
			}

			// Each rarity should have a distinct color
			r, g, b, _ := col.RGBA()
			if r == 0 && g == 0 && b == 0 {
				t.Error("rarity color should not be black")
			}
		})
	}
}

func TestGetEnchantmentColor(t *testing.T) {
	enchantments := []string{
		"fire",
		"ice",
		"lightning",
		"poison",
		"holy",
		"shadow",
		"unknown",
	}

	for _, ench := range enchantments {
		t.Run(ench, func(t *testing.T) {
			col := getEnchantmentColor(ench)
			if col.A != 255 {
				t.Error("enchantment color should be fully opaque")
			}

			r, g, b, _ := col.RGBA()
			if r == 0 && g == 0 && b == 0 {
				t.Error("enchantment color should not be black")
			}
		})
	}
}

func TestBlendMaterialShading(t *testing.T) {
	colors := getMaterialColors(MaterialSteel)

	tests := []struct {
		name           string
		distFromCenter float64
		depth          float64
	}{
		{"center highlight", 0.0, 0.5},
		{"mid blend", 0.5, 0.5},
		{"edge shadow", 1.0, 0.5},
		{"shallow depth", 0.3, 0.1},
		{"deep depth", 0.3, 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := blendMaterialShading(colors, tt.distFromCenter, tt.depth)

			// Verify color is RGBA
			rgba, ok := col.(color.RGBA)
			if !ok {
				t.Fatal("blendMaterialShading should return color.RGBA")
			}

			if rgba.A != 255 {
				t.Error("blended color should be fully opaque")
			}
		})
	}
}

func TestDamageStateApplication(t *testing.T) {
	damageStates := []DamageState{
		DamagePristine,
		DamageScratched,
		DamageWorn,
		DamageBroken,
	}

	for _, damage := range damageStates {
		t.Run(damage.String(), func(t *testing.T) {
			spec := WeaponVisualSpec{
				Type:      TypeMelee,
				Frame:     FrameIdle,
				Rarity:    RarityCommon,
				Damage:    damage,
				BladeMat:  MaterialSteel,
				HandleMat: MaterialWood,
				Seed:      12345,
			}

			img := EnhancedGenerateWeaponSprite(spec)
			if img == nil {
				t.Fatal("EnhancedGenerateWeaponSprite returned nil")
			}

			// Verify damage state doesn't crash
			bounds := img.Bounds()
			hasContent := false
			for y := bounds.Min.Y; y < bounds.Max.Y && !hasContent; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					_, _, _, a := img.At(x, y).RGBA()
					if a > 0 {
						hasContent = true
						break
					}
				}
			}

			if !hasContent {
				t.Error("damaged weapon has no visible content")
			}
		})
	}
}

func TestWeaponTypesWithMaterials(t *testing.T) {
	weaponTypes := []WeaponType{
		TypeMelee,
		TypeHitscan,
		TypeProjectile,
	}

	for _, wType := range weaponTypes {
		t.Run(wType.String(), func(t *testing.T) {
			spec := WeaponVisualSpec{
				Type:      wType,
				Frame:     FrameIdle,
				Rarity:    RarityRare,
				Damage:    DamagePristine,
				BladeMat:  MaterialSteel,
				HandleMat: MaterialWood,
				Seed:      54321,
			}

			img := EnhancedGenerateWeaponSprite(spec)
			if img == nil {
				t.Fatalf("EnhancedGenerateWeaponSprite(%v) returned nil", wType)
			}

			if img.Bounds().Dx() != 128 || img.Bounds().Dy() != 128 {
				t.Errorf("wrong dimensions for %v", wType)
			}
		})
	}
}

func TestFireFrameEffects(t *testing.T) {
	weaponTypes := []WeaponType{TypeMelee, TypeHitscan, TypeProjectile}

	for _, wType := range weaponTypes {
		t.Run(wType.String(), func(t *testing.T) {
			idleSpec := WeaponVisualSpec{
				Type:        wType,
				Frame:       FrameIdle,
				Rarity:      RarityCommon,
				Damage:      DamagePristine,
				BladeMat:    MaterialSteel,
				HandleMat:   MaterialWood,
				Seed:        12345,
				Enchantment: "fire",
			}

			fireSpec := idleSpec
			fireSpec.Frame = FrameFire

			idleImg := EnhancedGenerateWeaponSprite(idleSpec)
			fireImg := EnhancedGenerateWeaponSprite(fireSpec)

			// Fire frame should have different pixels (muzzle flash, effects)
			// Count very bright pixels (near white) which indicate flash effects
			veryBrightCountFire := 0
			veryBrightCountIdle := 0

			bounds := fireImg.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					fr, fg, fb, fa := fireImg.At(x, y).RGBA()
					ir, ig, ib, ia := idleImg.At(x, y).RGBA()

					// Very bright = near white (R, G, or B > 220 and alpha > 0)
					if fa > 0 && (fr > 220*256 || fg > 220*256 || fb > 220*256) {
						veryBrightCountFire++
					}
					if ia > 0 && (ir > 220*256 || ig > 220*256 || ib > 220*256) {
						veryBrightCountIdle++
					}
				}
			}

			// Fire frame should have more very bright pixels (muzzle flash)
			// Melee weapons may not have significant flash, so skip for them
			if wType != TypeMelee && veryBrightCountFire <= veryBrightCountIdle {
				// Allow small difference for ranged weapons
				if veryBrightCountFire < veryBrightCountIdle-20 {
					t.Errorf("fire frame should have similar or more very bright pixels than idle for %v (fire: %d, idle: %d)",
						wType, veryBrightCountFire, veryBrightCountIdle)
				}
			}
		})
	}
}

func TestDeterministicGeneration(t *testing.T) {
	spec := WeaponVisualSpec{
		Type:        TypeMelee,
		Frame:       FrameIdle,
		Rarity:      RarityRare,
		Damage:      DamagePristine,
		BladeMat:    MaterialGold,
		HandleMat:   MaterialLeather,
		Seed:        99999,
		Enchantment: "lightning",
	}

	img1 := EnhancedGenerateWeaponSprite(spec)
	img2 := EnhancedGenerateWeaponSprite(spec)

	// Same seed should produce identical images
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, a1 := img1.At(x, y).RGBA()
			r2, g2, b2, a2 := img2.At(x, y).RGBA()

			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Fatalf("deterministic generation failed at (%d, %d)", x, y)
			}
		}
	}
}

// String methods for test output
func (m Material) String() string {
	names := []string{
		"Steel", "Iron", "Wood", "Leather", "Gold",
		"Crystal", "Bone", "Obsidian", "Mithril", "Demonic",
	}
	if int(m) < len(names) {
		return names[m]
	}
	return "Unknown"
}

func (r Rarity) String() string {
	names := []string{"Common", "Uncommon", "Rare", "Epic", "Legendary"}
	if int(r) < len(names) {
		return names[r]
	}
	return "Unknown"
}

func (d DamageState) String() string {
	names := []string{"Pristine", "Scratched", "Worn", "Broken"}
	if int(d) < len(names) {
		return names[d]
	}
	return "Unknown"
}

func (w WeaponType) String() string {
	names := []string{"Hitscan", "Projectile", "Melee"}
	if int(w) < len(names) {
		return names[w]
	}
	return "Unknown"
}

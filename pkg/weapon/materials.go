// Package weapon implements enhanced weapon visual generation with materials and rarity.
package weapon

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/violence/pkg/common"
)

// Material defines the visual appearance of weapon components.
type Material int

const (
	MaterialSteel Material = iota
	MaterialIron
	MaterialWood
	MaterialLeather
	MaterialGold
	MaterialCrystal
	MaterialBone
	MaterialObsidian
	MaterialMithril
	MaterialDemonic
)

// Rarity affects visual complexity and effects.
type Rarity int

const (
	RarityCommon Rarity = iota
	RarityUncommon
	RarityRare
	RarityEpic
	RarityLegendary
)

// DamageState represents weapon condition.
type DamageState int

const (
	DamagePristine DamageState = iota
	DamageScratched
	DamageWorn
	DamageBroken
)

// WeaponVisualSpec defines all visual parameters for weapon generation.
type WeaponVisualSpec struct {
	Type        WeaponType
	Frame       FrameType
	Rarity      Rarity
	Damage      DamageState
	BladeMat    Material
	HandleMat   Material
	Seed        int64
	Enchantment string // "fire", "ice", "lightning", "poison", "holy", "shadow", ""
}

// EnhancedGenerateWeaponSprite creates a visually rich weapon sprite.
func EnhancedGenerateWeaponSprite(spec WeaponVisualSpec) *image.RGBA {
	const size = 128
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	rng := rand.New(rand.NewSource(spec.Seed))

	// Generate base weapon with materials
	switch spec.Type {
	case TypeMelee:
		enhancedMeleeWeapon(img, rng, spec)
	case TypeHitscan:
		enhancedHitscanWeapon(img, rng, spec)
	case TypeProjectile:
		enhancedProjectileWeapon(img, rng, spec)
	}

	// Apply rarity effects
	applyRarityEffects(img, spec.Rarity, rng)

	// Apply enchantment glow
	if spec.Enchantment != "" {
		applyEnchantmentGlow(img, spec.Enchantment, spec.Frame, rng)
	}

	// Apply damage state
	applyDamageState(img, spec.Damage, rng)

	return img
}

// enhancedMeleeWeapon generates melee weapons with material shading.
func enhancedMeleeWeapon(img *image.RGBA, rng *rand.Rand, spec WeaponVisualSpec) {
	// Blade with material-specific rendering
	bladeColors := getMaterialColors(spec.BladeMat)
	handleColors := getMaterialColors(spec.HandleMat)

	// Blade shape (tapered from guard to tip)
	for y := 25; y < 88; y++ {
		progress := float64(y-25) / 63.0
		width := 16.0 - progress*12.0
		if width < 3 {
			width = 3
		}

		x1 := int(64 - width/2)
		x2 := int(64 + width/2)

		for x := x1; x <= x2; x++ {
			// Distance from blade center for shading
			distFromCenter := math.Abs(float64(x) - 64.0)
			normalizedDist := distFromCenter / (width / 2)

			// Apply material shading
			col := blendMaterialShading(bladeColors, normalizedDist, float64(y)/128.0)
			img.Set(x, y, col)
		}
	}

	// Edge highlight for metallic materials
	if isMetallic(spec.BladeMat) {
		highlightCol := lightenColor(bladeColors.base, 0.4)
		for y := 25; y < 88; y++ {
			progress := float64(y-25) / 63.0
			width := 16.0 - progress*12.0
			centerX := 64
			offsetX := int(width / 4)

			img.Set(centerX-offsetX, y, highlightCol)
			img.Set(centerX+offsetX, y, darkenColor(bladeColors.base, 0.2))
		}
	}

	// Guard (crossguard)
	guardY := 88
	common.FillRect(img, 35, guardY, 93, guardY+8, handleColors.base)
	// Guard detail with gradient
	for x := 35; x < 93; x++ {
		distFromCenter := math.Abs(float64(x) - 64.0)
		if distFromCenter < 4 {
			img.Set(x, guardY+1, lightenColor(handleColors.base, 0.3))
		}
	}

	// Handle with wood grain or leather wrapping
	handleY1 := 96
	handleY2 := 122
	if spec.HandleMat == MaterialLeather {
		drawLeatherWrap(img, 56, handleY1, 72, handleY2, handleColors, rng)
	} else if spec.HandleMat == MaterialWood {
		drawWoodGrain(img, 56, handleY1, 72, handleY2, handleColors, rng)
	} else {
		// Metal or other solid handle
		common.FillRect(img, 56, handleY1, 72, handleY2, handleColors.base)
		// Grip ridges
		for y := handleY1 + 4; y < handleY2; y += 6 {
			common.FillRect(img, 56, y, 72, y+2, darkenColor(handleColors.base, 0.15))
		}
	}

	// Pommel with decorative detail
	pommelY := 124
	pommelRadius := 7
	common.FillCircle(img, 64, pommelY, pommelRadius, handleColors.shadow)
	common.FillCircle(img, 64, pommelY, pommelRadius-1, handleColors.base)
	common.FillCircle(img, 64, pommelY, pommelRadius-3, handleColors.highlight)

	// Rarity-specific pommel jewel
	if spec.Rarity >= RarityRare {
		jewelCol := getRarityJewelColor(spec.Rarity)
		common.FillCircle(img, 64, pommelY, 3, jewelCol)
	}
}

// enhancedHitscanWeapon generates firearms with metallic sheen and details.
func enhancedHitscanWeapon(img *image.RGBA, rng *rand.Rand, spec WeaponVisualSpec) {
	metalColors := getMaterialColors(spec.BladeMat)
	gripColors := getMaterialColors(spec.HandleMat)

	// Barrel with cylindrical shading
	barrelY1 := 48
	barrelY2 := 62
	barrelX1 := 60
	barrelX2 := 115

	for y := barrelY1; y <= barrelY2; y++ {
		distFromCenter := math.Abs(float64(y) - float64(barrelY1+barrelY2)/2.0)
		normalizedDist := distFromCenter / (float64(barrelY2-barrelY1) / 2.0)
		col := blendMaterialShading(metalColors, normalizedDist, 0.3)

		for x := barrelX1; x <= barrelX2; x++ {
			img.Set(x, y, col)
		}
	}

	// Barrel highlight (top edge)
	for x := barrelX1; x <= barrelX2; x++ {
		img.Set(x, barrelY1+1, lightenColor(metalColors.base, 0.4))
		img.Set(x, barrelY1+2, lightenColor(metalColors.base, 0.2))
	}

	// Receiver (main body) with beveled edges
	receiverX1 := 42
	receiverX2 := 78
	receiverY1 := 54
	receiverY2 := 78

	common.FillRect(img, receiverX1, receiverY1, receiverX2, receiverY2, metalColors.base)

	// Beveled top edge
	for x := receiverX1; x <= receiverX2; x++ {
		img.Set(x, receiverY1, lightenColor(metalColors.base, 0.3))
		img.Set(x, receiverY1+1, lightenColor(metalColors.base, 0.15))
	}

	// Receiver side panels with rivet detail
	for y := receiverY1 + 8; y < receiverY2-4; y += 8 {
		common.FillCircle(img, receiverX1+4, y, 1, darkenColor(metalColors.base, 0.3))
		common.FillCircle(img, receiverX2-4, y, 1, darkenColor(metalColors.base, 0.3))
	}

	// Grip with texture
	gripX1 := 48
	gripX2 := 62
	gripY1 := 78
	gripY2 := 100

	if spec.HandleMat == MaterialWood {
		drawWoodGrain(img, gripX1, gripY1, gripX2, gripY2, gripColors, rng)
	} else if spec.HandleMat == MaterialLeather {
		drawLeatherWrap(img, gripX1, gripY1, gripX2, gripY2, gripColors, rng)
	} else {
		// Textured polymer grip
		drawGripPattern(img, gripX1, gripY1, gripX2, gripY2, gripColors)
	}

	// Trigger and guard
	common.DrawLine(img, 58, 74, 58, 82, lightenColor(metalColors.base, 0.1), 2)
	common.DrawLine(img, 58, 82, 66, 82, lightenColor(metalColors.base, 0.1), 2)
	common.FillRect(img, 60, 76, 63, 80, color.RGBA{R: 30, G: 30, B: 35, A: 255})

	// Front sight
	common.FillRect(img, 105, 44, 108, 48, lightenColor(metalColors.base, 0.2))

	// Muzzle flash (fire frame only)
	if spec.Frame == FrameFire {
		drawMuzzleFlash(img, barrelX2, (barrelY1+barrelY2)/2, spec.Enchantment, rng)
	}
}

// enhancedProjectileWeapon generates launchers with energy effects.
func enhancedProjectileWeapon(img *image.RGBA, rng *rand.Rand, spec WeaponVisualSpec) {
	bodyColors := getMaterialColors(spec.BladeMat)
	accentColors := getMaterialColors(spec.HandleMat)

	// Large barrel tube
	barrelY1 := 42
	barrelY2 := 72
	barrelX1 := 45
	barrelX2 := 118

	for y := barrelY1; y <= barrelY2; y++ {
		distFromCenter := math.Abs(float64(y) - float64(barrelY1+barrelY2)/2.0)
		normalizedDist := distFromCenter / (float64(barrelY2-barrelY1) / 2.0)
		col := blendMaterialShading(bodyColors, normalizedDist, 0.4)

		for x := barrelX1; x <= barrelX2; x++ {
			img.Set(x, y, col)
		}
	}

	// Barrel highlight
	for x := barrelX1; x <= barrelX2; x++ {
		img.Set(x, barrelY1+2, lightenColor(bodyColors.base, 0.3))
	}

	// Muzzle opening
	muzzleX := barrelX2
	muzzleY := (barrelY1 + barrelY2) / 2
	common.FillCircle(img, muzzleX, muzzleY, 14, bodyColors.shadow)
	common.FillCircle(img, muzzleX, muzzleY, 12, bodyColors.base)
	common.FillCircle(img, muzzleX, muzzleY, 8, color.RGBA{R: 20, G: 20, B: 25, A: 255})

	// Energy coils or tech details
	numCoils := 5 + int(spec.Rarity)
	for i := 0; i < numCoils; i++ {
		x := barrelX1 + 10 + i*12
		y := muzzleY

		// Coil housing
		common.FillCircle(img, x, y, 5, accentColors.shadow)
		common.FillCircle(img, x, y, 4, accentColors.base)

		// Energy glow inside
		if spec.Rarity >= RarityUncommon {
			glowCol := getEnchantmentColor(spec.Enchantment)
			common.FillCircle(img, x, y, 2, glowCol)
		}
	}

	// Grip underneath
	gripX1 := 52
	gripX2 := 70
	gripY1 := 72
	gripY2 := 94
	drawGripPattern(img, gripX1, gripY1, gripX2, gripY2, accentColors)

	// Stock/brace
	common.FillRect(img, 40, muzzleY-6, barrelX1, muzzleY+6, bodyColors.base)

	// Muzzle charge effect (fire frame)
	if spec.Frame == FrameFire {
		drawProjectileMuzzleCharge(img, muzzleX, muzzleY, spec.Enchantment, rng)
	}
}

// MaterialColors holds shading colors for a material.
type MaterialColors struct {
	base      color.RGBA
	highlight color.RGBA
	shadow    color.RGBA
}

// getMaterialColors returns color palette for material rendering.
func getMaterialColors(mat Material) MaterialColors {
	switch mat {
	case MaterialSteel:
		return MaterialColors{
			base:      color.RGBA{R: 160, G: 165, B: 175, A: 255},
			highlight: color.RGBA{R: 210, G: 215, B: 225, A: 255},
			shadow:    color.RGBA{R: 90, G: 95, B: 105, A: 255},
		}
	case MaterialIron:
		return MaterialColors{
			base:      color.RGBA{R: 130, G: 130, B: 140, A: 255},
			highlight: color.RGBA{R: 170, G: 170, B: 185, A: 255},
			shadow:    color.RGBA{R: 70, G: 70, B: 80, A: 255},
		}
	case MaterialWood:
		return MaterialColors{
			base:      color.RGBA{R: 120, G: 80, B: 50, A: 255},
			highlight: color.RGBA{R: 160, G: 110, B: 70, A: 255},
			shadow:    color.RGBA{R: 70, G: 45, B: 25, A: 255},
		}
	case MaterialLeather:
		return MaterialColors{
			base:      color.RGBA{R: 100, G: 70, B: 50, A: 255},
			highlight: color.RGBA{R: 130, G: 95, B: 70, A: 255},
			shadow:    color.RGBA{R: 60, G: 40, B: 30, A: 255},
		}
	case MaterialGold:
		return MaterialColors{
			base:      color.RGBA{R: 220, G: 180, B: 60, A: 255},
			highlight: color.RGBA{R: 255, G: 230, B: 120, A: 255},
			shadow:    color.RGBA{R: 150, G: 120, B: 30, A: 255},
		}
	case MaterialCrystal:
		return MaterialColors{
			base:      color.RGBA{R: 180, G: 200, B: 230, A: 255},
			highlight: color.RGBA{R: 230, G: 240, B: 255, A: 255},
			shadow:    color.RGBA{R: 100, G: 130, B: 180, A: 255},
		}
	case MaterialBone:
		return MaterialColors{
			base:      color.RGBA{R: 220, G: 210, B: 190, A: 255},
			highlight: color.RGBA{R: 245, G: 240, B: 230, A: 255},
			shadow:    color.RGBA{R: 150, G: 140, B: 120, A: 255},
		}
	case MaterialObsidian:
		return MaterialColors{
			base:      color.RGBA{R: 30, G: 25, B: 35, A: 255},
			highlight: color.RGBA{R: 80, G: 70, B: 90, A: 255},
			shadow:    color.RGBA{R: 10, G: 10, B: 15, A: 255},
		}
	case MaterialMithril:
		return MaterialColors{
			base:      color.RGBA{R: 200, G: 220, B: 240, A: 255},
			highlight: color.RGBA{R: 240, G: 250, B: 255, A: 255},
			shadow:    color.RGBA{R: 120, G: 140, B: 165, A: 255},
		}
	case MaterialDemonic:
		return MaterialColors{
			base:      color.RGBA{R: 140, G: 20, B: 30, A: 255},
			highlight: color.RGBA{R: 200, G: 50, B: 60, A: 255},
			shadow:    color.RGBA{R: 70, G: 10, B: 15, A: 255},
		}
	default:
		return getMaterialColors(MaterialSteel)
	}
}

// blendMaterialShading applies gradient shading based on geometry.
func blendMaterialShading(colors MaterialColors, distFromCenter, depth float64) color.Color {
	baseR, baseG, baseB, _ := colors.base.RGBA()
	highlightR, highlightG, highlightB, _ := colors.highlight.RGBA()
	shadowR, shadowG, shadowB, _ := colors.shadow.RGBA()

	// Depth gradient
	depthFactor := 1.0 - depth*0.2

	// Blend between highlight, base, and shadow
	var r, g, b uint32
	if distFromCenter < 0.3 {
		// Center highlight
		t := distFromCenter / 0.3
		r = uint32(float64(highlightR)*(1-t) + float64(baseR)*t)
		g = uint32(float64(highlightG)*(1-t) + float64(baseG)*t)
		b = uint32(float64(highlightB)*(1-t) + float64(baseB)*t)
	} else {
		// Edge shadow
		t := (distFromCenter - 0.3) / 0.7
		r = uint32(float64(baseR)*(1-t) + float64(shadowR)*t)
		g = uint32(float64(baseG)*(1-t) + float64(shadowG)*t)
		b = uint32(float64(baseB)*(1-t) + float64(shadowB)*t)
	}

	// Apply depth factor
	r = uint32(float64(r) * depthFactor)
	g = uint32(float64(g) * depthFactor)
	b = uint32(float64(b) * depthFactor)

	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: 255,
	}
}

// isMetallic checks if material should have metallic sheen.
func isMetallic(mat Material) bool {
	switch mat {
	case MaterialSteel, MaterialIron, MaterialGold, MaterialMithril, MaterialObsidian:
		return true
	default:
		return false
	}
}

// lightenColor increases color brightness.
func lightenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(clamp(float64(c.R) + 255*factor)),
		G: uint8(clamp(float64(c.G) + 255*factor)),
		B: uint8(clamp(float64(c.B) + 255*factor)),
		A: c.A,
	}
}

// darkenColor decreases color brightness.
func darkenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(clamp(float64(c.R) * (1 - factor))),
		G: uint8(clamp(float64(c.G) * (1 - factor))),
		B: uint8(clamp(float64(c.B) * (1 - factor))),
		A: c.A,
	}
}

// drawWoodGrain renders wood texture with grain lines.
func drawWoodGrain(img *image.RGBA, x1, y1, x2, y2 int, colors MaterialColors, rng *rand.Rand) {
	// Fill base color
	common.FillRect(img, x1, y1, x2, y2, colors.base)

	// Vertical grain lines
	numGrains := 4 + rng.Intn(4)
	for i := 0; i < numGrains; i++ {
		grainX := x1 + rng.Intn(x2-x1)
		grainCol := darkenColor(colors.base, 0.1+rng.Float64()*0.15)

		for y := y1; y < y2; y++ {
			// Wavy grain
			offset := int(math.Sin(float64(y)*0.3) * 1.5)
			img.Set(grainX+offset, y, grainCol)
		}
	}

	// Add knots occasionally
	if rng.Float64() < 0.3 {
		knotX := x1 + (x2-x1)/2 + rng.Intn(6) - 3
		knotY := y1 + (y2-y1)/2 + rng.Intn(6) - 3
		common.FillCircle(img, knotX, knotY, 2, colors.shadow)
	}
}

// drawLeatherWrap renders wrapped leather texture.
func drawLeatherWrap(img *image.RGBA, x1, y1, x2, y2 int, colors MaterialColors, rng *rand.Rand) {
	// Fill base
	common.FillRect(img, x1, y1, x2, y2, colors.base)

	// Diagonal wrap lines
	numWraps := (y2 - y1) / 6
	for i := 0; i < numWraps; i++ {
		y := y1 + i*6
		wrapCol := darkenColor(colors.base, 0.15)
		common.FillRect(img, x1, y, x2, y+2, wrapCol)

		// Stitch marks
		for x := x1 + 2; x < x2-2; x += 4 {
			img.Set(x, y+1, colors.shadow)
		}
	}
}

// drawGripPattern renders textured grip.
func drawGripPattern(img *image.RGBA, x1, y1, x2, y2 int, colors MaterialColors) {
	common.FillRect(img, x1, y1, x2, y2, colors.base)

	// Diamond pattern
	for y := y1; y < y2; y += 4 {
		for x := x1; x < x2; x += 4 {
			offset := 0
			if (y-y1)/4%2 == 1 {
				offset = 2
			}
			img.Set(x+offset, y, colors.shadow)
			img.Set(x+offset+1, y+1, colors.highlight)
		}
	}
}

// applyRarityEffects adds visual flair based on rarity.
func applyRarityEffects(img *image.RGBA, rarity Rarity, rng *rand.Rand) {
	if rarity < RarityRare {
		return
	}

	bounds := img.Bounds()
	rarityCol := getRarityColor(rarity)

	// Add subtle glow around weapon edges
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a > 0 && a < 65535 { // Semi-transparent edge pixels
				// Blend rarity color
				blendFactor := 0.3
				nr := uint8((float64(r>>8)*(1-blendFactor) + float64(rarityCol.R)*blendFactor))
				ng := uint8((float64(g>>8)*(1-blendFactor) + float64(rarityCol.G)*blendFactor))
				nb := uint8((float64(b>>8)*(1-blendFactor) + float64(rarityCol.B)*blendFactor))
				img.Set(x, y, color.RGBA{R: nr, G: ng, B: nb, A: uint8(a >> 8)})
			}
		}
	}
}

// getRarityColor returns the signature color for a rarity tier.
func getRarityColor(rarity Rarity) color.RGBA {
	switch rarity {
	case RarityUncommon:
		return color.RGBA{R: 80, G: 200, B: 80, A: 255}
	case RarityRare:
		return color.RGBA{R: 80, G: 120, B: 240, A: 255}
	case RarityEpic:
		return color.RGBA{R: 200, G: 80, B: 240, A: 255}
	case RarityLegendary:
		return color.RGBA{R: 255, G: 165, B: 0, A: 255}
	default:
		return color.RGBA{R: 200, G: 200, B: 200, A: 255}
	}
}

// getRarityJewelColor returns jewel color for pommel decoration.
func getRarityJewelColor(rarity Rarity) color.RGBA {
	switch rarity {
	case RarityRare:
		return color.RGBA{R: 100, G: 140, B: 255, A: 255}
	case RarityEpic:
		return color.RGBA{R: 220, G: 100, B: 255, A: 255}
	case RarityLegendary:
		return color.RGBA{R: 255, G: 200, B: 50, A: 255}
	default:
		return color.RGBA{R: 150, G: 150, B: 150, A: 255}
	}
}

// applyEnchantmentGlow adds magical effects for enchanted weapons.
func applyEnchantmentGlow(img *image.RGBA, enchantment string, frame FrameType, rng *rand.Rand) {
	enchantCol := getEnchantmentColor(enchantment)
	bounds := img.Bounds()

	// Subtle glow on weapon pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 32768 { // Opaque weapon pixels
				// Small chance of sparkle
				if rng.Float64() < 0.02 {
					img.Set(x, y, lightenColor(enchantCol, 0.5))
				}
			}
		}
	}

	// Magical particles during fire frame
	if frame == FrameFire {
		numParticles := 8 + rng.Intn(8)
		for i := 0; i < numParticles; i++ {
			px := bounds.Min.X + rng.Intn(bounds.Dx())
			py := bounds.Min.Y + rng.Intn(bounds.Dy())
			common.FillCircle(img, px, py, 1, enchantCol)
		}
	}
}

// getEnchantmentColor returns the color for an enchantment type.
func getEnchantmentColor(enchantment string) color.RGBA {
	switch enchantment {
	case "fire":
		return color.RGBA{R: 255, G: 120, B: 40, A: 255}
	case "ice":
		return color.RGBA{R: 120, G: 200, B: 255, A: 255}
	case "lightning":
		return color.RGBA{R: 200, G: 220, B: 255, A: 255}
	case "poison":
		return color.RGBA{R: 100, G: 255, B: 80, A: 255}
	case "holy":
		return color.RGBA{R: 255, G: 250, B: 200, A: 255}
	case "shadow":
		return color.RGBA{R: 120, G: 80, B: 160, A: 255}
	default:
		return color.RGBA{R: 200, G: 200, B: 255, A: 255}
	}
}

// applyDamageState adds wear and tear based on weapon condition.
func applyDamageState(img *image.RGBA, damage DamageState, rng *rand.Rand) {
	if damage == DamagePristine {
		return
	}

	bounds := img.Bounds()

	switch damage {
	case DamageScratched:
		// Add small scratches
		numScratches := 5 + rng.Intn(8)
		for i := 0; i < numScratches; i++ {
			x := bounds.Min.X + rng.Intn(bounds.Dx())
			y := bounds.Min.Y + rng.Intn(bounds.Dy())
			length := 3 + rng.Intn(5)
			angle := rng.Float64() * math.Pi * 2

			for j := 0; j < length; j++ {
				sx := x + int(float64(j)*math.Cos(angle))
				sy := y + int(float64(j)*math.Sin(angle))
				if sx >= bounds.Min.X && sx < bounds.Max.X && sy >= bounds.Min.Y && sy < bounds.Max.Y {
					r, g, b, a := img.At(sx, sy).RGBA()
					if a > 32768 {
						img.Set(sx, sy, color.RGBA{
							R: uint8((r >> 8) * 9 / 10),
							G: uint8((g >> 8) * 9 / 10),
							B: uint8((b >> 8) * 9 / 10),
							A: uint8(a >> 8),
						})
					}
				}
			}
		}

	case DamageWorn:
		// Significant wear - darken overall
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				if a > 32768 {
					img.Set(x, y, color.RGBA{
						R: uint8((r >> 8) * 8 / 10),
						G: uint8((g >> 8) * 8 / 10),
						B: uint8((b >> 8) * 8 / 10),
						A: uint8(a >> 8),
					})
				}
			}
		}

		// Add rust spots or tarnish
		numSpots := 8 + rng.Intn(10)
		rustCol := color.RGBA{R: 120, G: 60, B: 30, A: 255}
		for i := 0; i < numSpots; i++ {
			x := bounds.Min.X + rng.Intn(bounds.Dx())
			y := bounds.Min.Y + rng.Intn(bounds.Dy())
			common.FillCircle(img, x, y, 1+rng.Intn(2), rustCol)
		}

	case DamageBroken:
		// Cracks and severe damage
		numCracks := 3 + rng.Intn(4)
		for i := 0; i < numCracks; i++ {
			x := bounds.Min.X + rng.Intn(bounds.Dx())
			y := bounds.Min.Y + rng.Intn(bounds.Dy())
			length := 10 + rng.Intn(20)
			angle := rng.Float64() * math.Pi * 2

			for j := 0; j < length; j++ {
				cx := x + int(float64(j)*math.Cos(angle))
				cy := y + int(float64(j)*math.Sin(angle))
				if cx >= bounds.Min.X && cx < bounds.Max.X && cy >= bounds.Min.Y && cy < bounds.Max.Y {
					img.Set(cx, cy, color.RGBA{R: 20, G: 20, B: 20, A: 255})
				}
			}
		}
	}
}

// drawMuzzleFlash renders gun muzzle flash with enchantment color.
func drawMuzzleFlash(img *image.RGBA, x, y int, enchantment string, rng *rand.Rand) {
	flashCol := color.RGBA{R: 255, G: 240, B: 100, A: 255}
	if enchantment != "" {
		flashCol = getEnchantmentColor(enchantment)
	}

	// Central bright core
	common.FillCircle(img, x, y, 10, flashCol)
	common.FillCircle(img, x, y, 6, lightenColor(flashCol, 0.5))

	// Radial rays
	numRays := 6 + rng.Intn(4)
	for i := 0; i < numRays; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numRays)
		length := 12 + rng.Intn(8)
		ex := x + int(math.Cos(angle)*float64(length))
		ey := y + int(math.Sin(angle)*float64(length))
		common.DrawLine(img, x, y, ex, ey, flashCol, 2)
	}

	// Expanding ring
	common.FillCircle(img, x, y, 14, color.RGBA{
		R: flashCol.R,
		G: flashCol.G,
		B: flashCol.B,
		A: 128,
	})
}

// drawProjectileMuzzleCharge renders launcher charge effect.
func drawProjectileMuzzleCharge(img *image.RGBA, x, y int, enchantment string, rng *rand.Rand) {
	chargeCol := color.RGBA{R: 255, G: 200, B: 100, A: 255}
	if enchantment != "" {
		chargeCol = getEnchantmentColor(enchantment)
	}

	// Expanding rings
	for r := 20; r > 6; r -= 4 {
		alpha := uint8(255 * (20 - r) / 14)
		ringCol := color.RGBA{
			R: chargeCol.R,
			G: chargeCol.G,
			B: chargeCol.B,
			A: alpha,
		}
		common.FillCircle(img, x, y, r, ringCol)
	}

	// Bright center
	common.FillCircle(img, x, y, 8, lightenColor(chargeCol, 0.6))

	// Energy arcs
	numArcs := 4 + rng.Intn(4)
	for i := 0; i < numArcs; i++ {
		angle := rng.Float64() * 2 * math.Pi
		dist := 15 + rng.Intn(10)
		ex := x + int(math.Cos(angle)*float64(dist))
		ey := y + int(math.Sin(angle)*float64(dist))

		// Jagged arc
		steps := 5
		prevX, prevY := x, y
		for s := 1; s <= steps; s++ {
			t := float64(s) / float64(steps)
			nx := x + int((float64(ex-x))*t) + rng.Intn(6) - 3
			ny := y + int((float64(ey-y))*t) + rng.Intn(6) - 3
			common.DrawLine(img, prevX, prevY, nx, ny, lightenColor(chargeCol, 0.3), 1)
			prevX, prevY = nx, ny
		}
	}
}

// clamp restricts value to [0, 255].
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

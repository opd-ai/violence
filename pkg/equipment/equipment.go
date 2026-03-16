// Package equipment provides visual rendering of equipped items on entity sprites.
package equipment

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// Slot identifies where an equipment piece is worn.
type Slot int

const (
	SlotWeapon    Slot = iota // SlotWeapon is the primary weapon slot.
	SlotHelmet                // SlotHelmet covers head protection.
	SlotChest                 // SlotChest covers torso armor.
	SlotLegs                  // SlotLegs covers leg armor.
	SlotBoots                 // SlotBoots covers footwear.
	SlotGloves                // SlotGloves covers hand protection.
	SlotAccessory1            // SlotAccessory1 is the first accessory slot.
	SlotAccessory2            // SlotAccessory2 is the second accessory slot.
	SlotCount                 // SlotCount is the total number of equipment slots.
)

// Material defines the visual texture and shading of equipment.
type Material int

const (
	MaterialIron        Material = iota // MaterialIron is basic forged iron.
	MaterialSteel                       // MaterialSteel is refined steel alloy.
	MaterialMithril                     // MaterialMithril is lightweight magical metal.
	MaterialLeather                     // MaterialLeather is tanned animal hide.
	MaterialCloth                       // MaterialCloth is woven fabric.
	MaterialDragonscale                 // MaterialDragonscale is armored dragon hide.
	MaterialCrystal                     // MaterialCrystal is crystalline material.
	MaterialNanofiber                   // MaterialNanofiber is advanced synthetic fiber.
	MaterialBiotech                     // MaterialBiotech is organic living armor.
	MaterialPlasma                      // MaterialPlasma is energy-based material.
)

// Rarity affects visual complexity and enchantment effects.
type Rarity int

const (
	RarityCommon    Rarity = iota // RarityCommon is basic quality gear.
	RarityUncommon                // RarityUncommon is slightly enhanced gear.
	RarityRare                    // RarityRare is exceptional gear.
	RarityEpic                    // RarityEpic is powerful heroic gear.
	RarityLegendary               // RarityLegendary is the most powerful gear.
)

// DamageState represents equipment wear level.
type DamageState int

const (
	StatePristine DamageState = iota // StatePristine is undamaged condition.
	StateWorn                        // StateWorn is lightly used condition.
	StateDamaged                     // StateDamaged is significantly worn condition.
	StateBroken                      // StateBroken is non-functional condition.
)

// Equipment represents a single equipped item.
type Equipment struct {
	Slot          Slot
	Material      Material
	Rarity        Rarity
	DamageState   DamageState
	Enchanted     bool
	EnchantColor  color.RGBA
	Seed          int64
	Name          string
	VisualPattern int
}

// EquipmentComponent holds all equipped items for an entity.
type EquipmentComponent struct {
	Items       [SlotCount]*Equipment
	GlowPhase   float64
	DirtyCache  bool
	CachedLayer *ebiten.Image
}

// Type implements Component interface.
func (e *EquipmentComponent) Type() string {
	return "equipment"
}

// EquipmentSystem renders equipment onto entity sprites.
type EquipmentSystem struct {
	genre      string
	layerCache map[layerKey]*ebiten.Image
	cacheMu    sync.RWMutex
	poolBySize map[int][]*image.RGBA
	poolMu     sync.Mutex
	logger     *logrus.Entry
}

type layerKey struct {
	slot        Slot
	material    Material
	rarity      Rarity
	damageState DamageState
	enchanted   bool
	pattern     int
	seed        int64
	genre       string
}

// NewEquipmentSystem creates the equipment rendering system.
func NewEquipmentSystem(genre string) *EquipmentSystem {
	return &EquipmentSystem{
		genre:      genre,
		layerCache: make(map[layerKey]*ebiten.Image),
		poolBySize: make(map[int][]*image.RGBA),
		logger: logrus.WithFields(logrus.Fields{
			"system": "equipment",
		}),
	}
}

// Update processes equipment components and updates glow animations.
func (sys *EquipmentSystem) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0

	eqType := reflect.TypeOf(&EquipmentComponent{})
	entities := w.Query(eqType)

	for _, entity := range entities {
		compRaw, ok := w.GetComponent(entity, eqType)
		if !ok {
			continue
		}
		eq, ok := compRaw.(*EquipmentComponent)
		if !ok {
			continue
		}

		hasEnchanted := false
		for _, item := range eq.Items {
			if item != nil && item.Enchanted {
				hasEnchanted = true
				break
			}
		}

		if hasEnchanted {
			eq.GlowPhase += deltaTime * 2.0
			if eq.GlowPhase > 2.0*math.Pi {
				eq.GlowPhase -= 2.0 * math.Pi
			}
		}
	}
}

// RenderEquipmentLayer composites equipment onto an entity sprite.
func (sys *EquipmentSystem) RenderEquipmentLayer(eq *EquipmentComponent, baseSprite *ebiten.Image, entityDir int) *ebiten.Image {
	if eq == nil || baseSprite == nil {
		return baseSprite
	}

	bounds := baseSprite.Bounds()
	size := bounds.Dx()

	if eq.CachedLayer != nil && !eq.DirtyCache {
		result := ebiten.NewImageFromImage(baseSprite)
		op := &ebiten.DrawImageOptions{}
		result.DrawImage(eq.CachedLayer, op)
		return result
	}

	layer := ebiten.NewImage(size, size)

	renderOrder := []Slot{
		SlotBoots, SlotLegs, SlotChest, SlotGloves,
		SlotHelmet, SlotWeapon, SlotAccessory1, SlotAccessory2,
	}

	for _, slot := range renderOrder {
		item := eq.Items[slot]
		if item == nil {
			continue
		}

		itemLayer := sys.generateEquipmentSprite(item, size, entityDir, eq.GlowPhase)
		if itemLayer != nil {
			op := &ebiten.DrawImageOptions{}
			layer.DrawImage(itemLayer, op)
		}
	}

	eq.CachedLayer = layer
	eq.DirtyCache = false

	result := ebiten.NewImageFromImage(baseSprite)
	op := &ebiten.DrawImageOptions{}
	result.DrawImage(layer, op)

	return result
}

// generateEquipmentSprite creates the visual representation of an equipment piece.
func (sys *EquipmentSystem) generateEquipmentSprite(item *Equipment, size, direction int, glowPhase float64) *ebiten.Image {
	key := layerKey{
		slot:        item.Slot,
		material:    item.Material,
		rarity:      item.Rarity,
		damageState: item.DamageState,
		enchanted:   item.Enchanted,
		pattern:     item.VisualPattern,
		seed:        item.Seed,
		genre:       sys.genre,
	}

	sys.cacheMu.RLock()
	cached := sys.layerCache[key]
	sys.cacheMu.RUnlock()

	if cached != nil {
		if item.Enchanted {
			return sys.applyEnchantmentGlow(cached, item.EnchantColor, glowPhase)
		}
		return cached
	}

	rgba := sys.acquireImageBuffer(size * size * 4)
	rng := rand.New(rand.NewSource(item.Seed))

	baseColor := sys.getMaterialColor(item.Material)
	accentColor := sys.getRarityAccent(item.Rarity)

	switch item.Slot {
	case SlotWeapon:
		sys.renderWeapon(rgba, baseColor, accentColor, item, direction, rng)
	case SlotHelmet:
		sys.renderHelmet(rgba, baseColor, accentColor, item, direction, rng)
	case SlotChest:
		sys.renderChestArmor(rgba, baseColor, accentColor, item, direction, rng)
	case SlotLegs:
		sys.renderLegArmor(rgba, baseColor, accentColor, item, direction, rng)
	case SlotBoots:
		sys.renderBoots(rgba, baseColor, accentColor, item, direction, rng)
	case SlotGloves:
		sys.renderGloves(rgba, baseColor, accentColor, item, direction, rng)
	case SlotAccessory1, SlotAccessory2:
		sys.renderAccessory(rgba, baseColor, accentColor, item, direction, rng)
	}

	sys.applyDamageWeathering(rgba, item.DamageState, rng)

	ebitenImg := ebiten.NewImageFromImage(rgba)

	sys.cacheMu.Lock()
	sys.layerCache[key] = ebitenImg
	sys.cacheMu.Unlock()

	sys.releaseImageBuffer(rgba)

	if item.Enchanted {
		return sys.applyEnchantmentGlow(ebitenImg, item.EnchantColor, glowPhase)
	}

	return ebitenImg
}

// renderWeapon draws weapon sprite on entity's hand position.
func (sys *EquipmentSystem) renderWeapon(img *image.RGBA, base, accent color.RGBA, item *Equipment, dir int, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	weaponX, weaponY := cx+8, cy
	if dir >= 4 {
		weaponX = cx - 8
	}

	weaponLen := 12
	weaponThick := 2

	for y := weaponY - weaponLen; y < weaponY; y++ {
		for x := weaponX - weaponThick/2; x < weaponX+weaponThick/2; x++ {
			if x >= 0 && x < size && y >= 0 && y < size {
				shade := 0.7 + 0.3*float64(weaponY-y)/float64(weaponLen)
				r := uint8(float64(base.R) * shade)
				g := uint8(float64(base.G) * shade)
				b := uint8(float64(base.B) * shade)
				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}

	if item.Rarity >= RarityRare {
		gemY := weaponY - weaponLen/2
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				x, y := weaponX+dx, gemY+dy
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, accent)
				}
			}
		}
	}
}

// renderHelmet draws helmet on entity's head.
func (sys *EquipmentSystem) renderHelmet(img *image.RGBA, base, accent color.RGBA, item *Equipment, dir int, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2-6

	helmetRadius := 6

	for y := cy - helmetRadius; y <= cy; y++ {
		for x := cx - helmetRadius; x <= cx+helmetRadius; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < float64(helmetRadius) {
				if x >= 0 && x < size && y >= 0 && y < size {
					shade := 0.6 + 0.4*(1.0-dist/float64(helmetRadius))
					r := uint8(float64(base.R) * shade)
					g := uint8(float64(base.G) * shade)
					b := uint8(float64(base.B) * shade)
					img.Set(x, y, color.RGBA{r, g, b, 255})
				}
			}
		}
	}

	if item.Rarity >= RarityEpic {
		for i := 0; i < 3; i++ {
			px := cx + i - 1
			py := cy - helmetRadius - 2
			if px >= 0 && px < size && py >= 0 && py < size {
				img.Set(px, py, accent)
			}
		}
	}
}

// renderChestArmor draws chest armor on entity torso.
func (sys *EquipmentSystem) renderChestArmor(img *image.RGBA, base, accent color.RGBA, item *Equipment, dir int, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	chestW := 10
	chestH := 8

	for y := cy - 2; y < cy+chestH-2; y++ {
		for x := cx - chestW/2; x < cx+chestW/2; x++ {
			if x >= 0 && x < size && y >= 0 && y < size {
				distFromCenter := math.Abs(float64(x - cx))
				shade := 0.7 + 0.3*(1.0-distFromCenter/float64(chestW/2))

				r := uint8(float64(base.R) * shade)
				g := uint8(float64(base.G) * shade)
				b := uint8(float64(base.B) * shade)

				img.Set(x, y, color.RGBA{r, g, b, 255})
			}
		}
	}

	if item.Rarity >= RarityRare && item.VisualPattern%3 == 0 {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				x, y := cx+dx, cy+dy
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, accent)
				}
			}
		}
	}
}

// renderLegArmor draws leg armor.
func (sys *EquipmentSystem) renderLegArmor(img *image.RGBA, base, accent color.RGBA, item *Equipment, dir int, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	legY := cy + 6
	legW := 3
	legH := 6

	for _, offsetX := range []int{-4, 4} {
		for y := legY; y < legY+legH; y++ {
			for x := cx + offsetX - legW/2; x < cx+offsetX+legW/2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					shade := 0.65 + 0.2*float64(y-legY)/float64(legH)
					r := uint8(float64(base.R) * shade)
					g := uint8(float64(base.G) * shade)
					b := uint8(float64(base.B) * shade)
					img.Set(x, y, color.RGBA{r, g, b, 255})
				}
			}
		}
	}
}

// renderBoots draws boots at feet position.
func (sys *EquipmentSystem) renderBoots(img *image.RGBA, base, accent color.RGBA, item *Equipment, dir int, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bootY := cy + 12
	bootW := 4
	bootH := 3

	for _, offsetX := range []int{-4, 4} {
		for y := bootY; y < bootY+bootH; y++ {
			for x := cx + offsetX - bootW/2; x < cx+offsetX+bootW/2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, base)
				}
			}
		}
	}
}

// renderGloves draws gloves on hands.
func (sys *EquipmentSystem) renderGloves(img *image.RGBA, base, accent color.RGBA, item *Equipment, dir int, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	handY := cy
	for _, offsetX := range []int{-8, 8} {
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				x, y := cx+offsetX+dx, handY+dy
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, base)
				}
			}
		}
	}
}

// renderAccessory draws accessories.
func (sys *EquipmentSystem) renderAccessory(img *image.RGBA, base, accent color.RGBA, item *Equipment, dir int, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	orbY := cy - 8
	if item.Slot == SlotAccessory2 {
		orbY = cy - 10
	}

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			if dx*dx+dy*dy <= 4 {
				x, y := cx+dx, orbY+dy
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, accent)
				}
			}
		}
	}
}

// applyDamageWeathering adds wear based on damage state.
func (sys *EquipmentSystem) applyDamageWeathering(img *image.RGBA, state DamageState, rng *rand.Rand) {
	if state == StatePristine {
		return
	}

	size := img.Bounds().Dx()
	scratchCount := 0

	switch state {
	case StateWorn:
		scratchCount = 3
	case StateDamaged:
		scratchCount = 8
	case StateBroken:
		scratchCount = 15
	}

	darkColor := color.RGBA{30, 30, 30, 180}

	for i := 0; i < scratchCount; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)
		length := 2 + rng.Intn(4)

		for j := 0; j < length; j++ {
			px := x + j
			if px < size && y < size {
				current := img.At(px, y)
				if _, _, _, a := current.RGBA(); a > 0 {
					img.Set(px, y, darkColor)
				}
			}
		}
	}
}

// applyEnchantmentGlow adds magical glow effect to equipment.
func (sys *EquipmentSystem) applyEnchantmentGlow(base *ebiten.Image, glowColor color.RGBA, phase float64) *ebiten.Image {
	size := base.Bounds().Dx()
	result := ebiten.NewImage(size, size)

	op := &ebiten.DrawImageOptions{}
	result.DrawImage(base, op)

	glowIntensity := 0.5 + 0.3*math.Sin(phase)

	glowLayer := ebiten.NewImage(size, size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			_, _, _, a := base.At(x, y).RGBA()
			if a > 0 {
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						gx, gy := x+dx, y+dy
						if gx >= 0 && gx < size && gy >= 0 && gy < size {
							alpha := uint8(float64(glowColor.A) * glowIntensity * 0.6)
							glowLayer.Set(gx, gy, color.RGBA{
								R: glowColor.R,
								G: glowColor.G,
								B: glowColor.B,
								A: alpha,
							})
						}
					}
				}
			}
		}
	}

	op2 := &ebiten.DrawImageOptions{}
	op2.Blend = ebiten.BlendLighter
	result.DrawImage(glowLayer, op2)

	return result
}

// getMaterialColor returns base color for material type.
func (sys *EquipmentSystem) getMaterialColor(mat Material) color.RGBA {
	switch mat {
	case MaterialIron:
		return color.RGBA{140, 140, 150, 255}
	case MaterialSteel:
		return color.RGBA{180, 180, 190, 255}
	case MaterialMithril:
		return color.RGBA{200, 220, 240, 255}
	case MaterialLeather:
		return color.RGBA{120, 80, 50, 255}
	case MaterialCloth:
		return color.RGBA{160, 140, 120, 255}
	case MaterialDragonscale:
		return color.RGBA{100, 50, 50, 255}
	case MaterialCrystal:
		return color.RGBA{180, 220, 255, 255}
	case MaterialNanofiber:
		return color.RGBA{60, 70, 80, 255}
	case MaterialBiotech:
		return color.RGBA{80, 140, 100, 255}
	case MaterialPlasma:
		return color.RGBA{200, 100, 255, 255}
	default:
		return color.RGBA{128, 128, 128, 255}
	}
}

// getRarityAccent returns accent color based on rarity.
func (sys *EquipmentSystem) getRarityAccent(rarity Rarity) color.RGBA {
	switch rarity {
	case RarityCommon:
		return color.RGBA{200, 200, 200, 255}
	case RarityUncommon:
		return color.RGBA{100, 255, 100, 255}
	case RarityRare:
		return color.RGBA{100, 100, 255, 255}
	case RarityEpic:
		return color.RGBA{200, 100, 255, 255}
	case RarityLegendary:
		return color.RGBA{255, 180, 50, 255}
	default:
		return color.RGBA{255, 255, 255, 255}
	}
}

// acquireImageBuffer gets a pooled image buffer.
func (sys *EquipmentSystem) acquireImageBuffer(sizeKey int) *image.RGBA {
	sys.poolMu.Lock()
	defer sys.poolMu.Unlock()

	pool := sys.poolBySize[sizeKey]
	if len(pool) > 0 {
		img := pool[len(pool)-1]
		sys.poolBySize[sizeKey] = pool[:len(pool)-1]
		for i := range img.Pix {
			img.Pix[i] = 0
		}
		return img
	}

	return image.NewRGBA(image.Rect(0, 0, 32, 32))
}

// releaseImageBuffer returns buffer to pool.
func (sys *EquipmentSystem) releaseImageBuffer(img *image.RGBA) {
	sys.poolMu.Lock()
	defer sys.poolMu.Unlock()

	sizeKey := len(img.Pix)
	sys.poolBySize[sizeKey] = append(sys.poolBySize[sizeKey], img)
}

// SetGenre updates genre for equipment styling.
func (sys *EquipmentSystem) SetGenre(genre string) {
	if sys.genre != genre {
		sys.genre = genre
		sys.cacheMu.Lock()
		sys.layerCache = make(map[layerKey]*ebiten.Image)
		sys.cacheMu.Unlock()
		sys.logger.WithField("genre", genre).Info("Equipment cache cleared for genre change")
	}
}

// Equip adds an equipment piece to an entity.
func Equip(eq *EquipmentComponent, item *Equipment) {
	if eq == nil || item == nil {
		return
	}
	eq.Items[item.Slot] = item
	eq.DirtyCache = true
}

// Unequip removes equipment from a slot.
func Unequip(eq *EquipmentComponent, slot Slot) *Equipment {
	if eq == nil {
		return nil
	}
	item := eq.Items[slot]
	eq.Items[slot] = nil
	eq.DirtyCache = true
	return item
}

// GetEquipped retrieves equipment from a slot.
func GetEquipped(eq *EquipmentComponent, slot Slot) *Equipment {
	if eq == nil {
		return nil
	}
	return eq.Items[slot]
}

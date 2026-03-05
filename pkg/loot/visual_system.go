// Package loot provides visual rendering for dropped loot items.
package loot

import (
	"image/color"
	"math"
	"math/rand"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// VisualComponent marks an entity as having a visual loot representation.
type VisualComponent struct {
	ItemID    string
	Category  ItemCategory
	Rarity    Rarity
	Seed      int64
	BobPhase  float64
	GlowPhase float64
	Collected bool
	SpawnTime float64
}

func (vc *VisualComponent) Type() string { return "LootVisual" }

// ItemCategory defines the visual category of a loot item.
type ItemCategory int

const (
	CategoryPotion ItemCategory = iota
	CategoryScroll
	CategoryWeapon
	CategoryArmor
	CategoryGold
	CategoryGear
	CategoryArtifact
	CategoryConsumable
)

// VisualSystem handles rendering of loot items in the world.
type VisualSystem struct {
	genreID string
	logger  *logrus.Entry
}

// NewVisualSystem creates a loot visual rendering system.
func NewVisualSystem(genreID string) *VisualSystem {
	return &VisualSystem{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "loot_visual",
			"genre":       genreID,
		}),
	}
}

// Update animates loot visuals with bobbing and glow effects.
func (vs *VisualSystem) Update(w *engine.World) {
	visualType := reflect.TypeOf((*VisualComponent)(nil))

	// Query all entities that have the visual component
	entities := w.Query(visualType)

	for _, entity := range entities {
		comp, found := w.GetComponent(entity, visualType)
		if !found {
			continue
		}

		lv, ok := comp.(*VisualComponent)
		if !ok || lv.Collected {
			continue
		}

		// Use fixed deltaTime for consistent animation
		deltaTime := 1.0 / 60.0

		lv.BobPhase += deltaTime * 2.0
		if lv.BobPhase > 2.0*math.Pi {
			lv.BobPhase -= 2.0 * math.Pi
		}

		glowSpeed := 3.0
		switch lv.Rarity {
		case RarityLegendary:
			glowSpeed = 5.0
		case RarityRare:
			glowSpeed = 4.0
		case RarityUncommon:
			glowSpeed = 3.5
		}

		lv.GlowPhase += deltaTime * glowSpeed
		if lv.GlowPhase > 2.0*math.Pi {
			lv.GlowPhase -= 2.0 * math.Pi
		}
	}
}

// GenerateItemSprite creates a procedural sprite for a loot item.
func (vs *VisualSystem) GenerateItemSprite(itemID string, category ItemCategory, rarity Rarity, seed int64, size int) *ebiten.Image {
	rng := rand.New(rand.NewSource(seed))

	img := ebiten.NewImage(size, size)

	switch category {
	case CategoryPotion:
		vs.drawPotion(img, rarity, rng, size)
	case CategoryScroll:
		vs.drawScroll(img, rarity, rng, size)
	case CategoryWeapon:
		vs.drawWeapon(img, rarity, rng, size)
	case CategoryArmor:
		vs.drawArmor(img, rarity, rng, size)
	case CategoryGold:
		vs.drawGold(img, rarity, rng, size)
	case CategoryGear:
		vs.drawGear(img, rarity, rng, size)
	case CategoryArtifact:
		vs.drawArtifact(img, rarity, rng, size)
	case CategoryConsumable:
		vs.drawConsumable(img, rarity, rng, size)
	default:
		vs.drawGeneric(img, rarity, rng, size)
	}

	return img
}

// drawPotion renders a potion bottle with liquid and label.
func (vs *VisualSystem) drawPotion(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	baseColor := vs.getPotionColor(rng)
	glassColor := color.RGBA{200, 200, 220, 255}

	cx, cy := size/2, size/2
	bottleWidth := size / 3
	bottleHeight := size * 2 / 3
	neckHeight := size / 6

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy

			if dy < -bottleHeight/2 && dy > -bottleHeight/2-neckHeight {
				if common.Abs(dx) < bottleWidth/4 {
					img.Set(x, y, glassColor)
				}
			} else if dy >= -bottleHeight/2 && dy < bottleHeight/2 {
				widthAtY := bottleWidth / 2
				if dy > bottleHeight/4 {
					ratio := float64(dy-bottleHeight/4) / float64(bottleHeight/4)
					widthAtY = int(float64(bottleWidth/2) * (1.0 - ratio*0.3))
				}

				if common.Abs(dx) < widthAtY {
					liquidLevel := -bottleHeight/2 + bottleHeight/4
					if dy > liquidLevel {
						brightness := 1.0 - float64(common.Abs(dx))/float64(widthAtY)*0.3
						img.Set(x, y, applyShade(baseColor, brightness))
					} else {
						img.Set(x, y, glassColor)
					}

					if common.Abs(dx) == widthAtY-1 || dy == -bottleHeight/2 {
						img.Set(x, y, color.RGBA{100, 100, 120, 255})
					}
				}
			}
		}
	}

	if rarity >= RarityRare {
		vs.addSparkles(img, rng, size, 3+int(rarity))
	}
}

// drawScroll renders a rolled parchment with runes.
func (vs *VisualSystem) drawScroll(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	parchmentColor := color.RGBA{220, 200, 160, 255}
	shadowColor := color.RGBA{140, 120, 90, 255}

	cx, cy := size/2, size/2
	scrollWidth := size * 2 / 3
	scrollHeight := size / 2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy

			if common.Abs(dy) < scrollHeight {
				edgeDist := scrollWidth/2 - common.Abs(dx)
				if edgeDist > 0 {
					shade := 1.0 - float64(common.Abs(dy))/float64(scrollHeight)*0.2
					if edgeDist < 3 {
						shade *= 0.7
					}
					img.Set(x, y, applyShade(parchmentColor, shade))

					if common.Abs(dy) == scrollHeight-1 || edgeDist == 0 {
						img.Set(x, y, shadowColor)
					}
				}
			}
		}
	}

	runeCount := 4 + int(rarity)*2
	for i := 0; i < runeCount; i++ {
		rx := cx - scrollWidth/3 + rng.Intn(scrollWidth*2/3)
		ry := cy - scrollHeight/2 + rng.Intn(scrollHeight)
		runeSize := 2 + rng.Intn(2)

		for dy := -runeSize; dy <= runeSize; dy++ {
			for dx := -runeSize; dx <= runeSize; dx++ {
				if common.Abs(dx)+common.Abs(dy) < runeSize {
					img.Set(rx+dx, ry+dy, vs.getRuneColor(rarity))
				}
			}
		}
	}
}

// drawWeapon renders a stylized weapon icon.
func (vs *VisualSystem) drawWeapon(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	bladeColor := vs.getMetalColor(rarity, rng)
	handleColor := color.RGBA{80 + uint8(rng.Intn(40)), 50 + uint8(rng.Intn(30)), 30, 255}

	cx, cy := size/2, size/2
	bladeLen := size * 2 / 3
	bladeWidth := size / 6
	handleLen := size / 3

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy

			angle := math.Atan2(float64(dy), float64(dx))
			dist := math.Sqrt(float64(dx*dx + dy*dy))

			weaponAngle := -math.Pi / 4
			relAngle := angle - weaponAngle
			alignedX := dist * math.Cos(relAngle)
			alignedY := dist * math.Sin(relAngle)

			if alignedX >= 0 && alignedX < float64(bladeLen) && common.Abs(int(alignedY)) < bladeWidth {
				brightness := 1.0 - math.Abs(alignedY)/float64(bladeWidth)*0.4
				img.Set(x, y, applyShade(bladeColor, brightness))
			} else if alignedX >= float64(bladeLen) && alignedX < float64(bladeLen+handleLen) && common.Abs(int(alignedY)) < bladeWidth+2 {
				img.Set(x, y, handleColor)
			}
		}
	}

	if rarity >= RarityRare {
		vs.addEnchantmentGlow(img, rarity, size)
	}
}

// drawArmor renders a shield or armor piece.
func (vs *VisualSystem) drawArmor(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	metalColor := vs.getMetalColor(rarity, rng)

	cx, cy := size/2, size/2
	shieldRadius := size / 3

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy
			dist := math.Sqrt(float64(dx*dx + dy*dy))

			if dist < float64(shieldRadius) {
				angle := math.Atan2(float64(dy), float64(dx))
				brightness := 0.7 + 0.3*math.Cos(angle*4)
				edgeFade := 1.0 - dist/float64(shieldRadius)*0.3
				brightness *= edgeFade

				img.Set(x, y, applyShade(metalColor, brightness))

				if int(dist) == shieldRadius-1 {
					img.Set(x, y, color.RGBA{50, 50, 50, 255})
				}
			}
		}
	}

	bossRadius := size / 8
	for y := -bossRadius; y <= bossRadius; y++ {
		for x := -bossRadius; x <= bossRadius; x++ {
			if x*x+y*y < bossRadius*bossRadius {
				brightness := 1.2 - float64(x*x+y*y)/float64(bossRadius*bossRadius)*0.4
				img.Set(cx+x, cy+y, applyShade(metalColor, brightness))
			}
		}
	}

	if rarity >= RarityUncommon {
		vs.addSparkles(img, rng, size, 2+int(rarity))
	}
}

// drawGold renders coins or gold pile.
func (vs *VisualSystem) drawGold(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	goldColor := color.RGBA{255, 215, 0, 255}

	cx, cy := size/2, size/2
	coinCount := 3 + int(rarity)*2

	for i := 0; i < coinCount; i++ {
		offsetX := rng.Intn(size/3) - size/6
		offsetY := rng.Intn(size/3) - size/6
		coinRadius := size/8 + rng.Intn(size/16)

		for y := -coinRadius; y <= coinRadius; y++ {
			for x := -coinRadius; x <= coinRadius; x++ {
				if x*x+y*y < coinRadius*coinRadius {
					brightness := 1.0 - float64(x*x+y*y)/float64(coinRadius*coinRadius)*0.5
					brightness += 0.3 * math.Cos(float64(x)/float64(coinRadius)*math.Pi)

					px := cx + offsetX + x
					py := cy + offsetY + y
					if px >= 0 && px < size && py >= 0 && py < size {
						img.Set(px, py, applyShade(goldColor, brightness))
					}
				}
			}
		}
	}
}

// drawGear renders mechanical/gear items.
func (vs *VisualSystem) drawGear(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	gearColor := color.RGBA{120, 120, 140, 255}
	if vs.genreID == "cyberpunk" || vs.genreID == "scifi" {
		gearColor = color.RGBA{0, 180, 255, 255}
	}

	cx, cy := size/2, size/2
	radius := size / 3
	teeth := 8 + int(rarity)*2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			angle := math.Atan2(float64(dy), float64(dx))

			teethAngle := angle * float64(teeth) / (2 * math.Pi)
			teethPhase := teethAngle - math.Floor(teethAngle)
			toothHeight := 0.0
			if teethPhase < 0.5 {
				toothHeight = float64(radius) / 4
			}

			if dist < float64(radius)+toothHeight && dist > float64(radius)/2 {
				brightness := 1.0 - (dist-float64(radius)/2)/(float64(radius)/2+toothHeight)*0.4
				img.Set(x, y, applyShade(gearColor, brightness))
			}
		}
	}

	centerRadius := size / 6
	for y := -centerRadius; y <= centerRadius; y++ {
		for x := -centerRadius; x <= centerRadius; x++ {
			if x*x+y*y < centerRadius*centerRadius {
				img.Set(cx+x, cy+y, color.RGBA{40, 40, 50, 255})
			}
		}
	}
}

// drawArtifact renders mystical artifact with complex patterns.
func (vs *VisualSystem) drawArtifact(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	baseColor := vs.getArtifactColor(rng)

	cx, cy := size/2, size/2
	radius := size / 3

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			angle := math.Atan2(float64(dy), float64(dx))

			if dist < float64(radius) {
				pattern := math.Sin(angle*5) * math.Cos(dist*0.3)
				brightness := 0.6 + 0.4*pattern
				brightness *= 1.0 - dist/float64(radius)*0.3

				img.Set(x, y, applyShade(baseColor, brightness))
			}
		}
	}

	runeCircleRadius := radius * 2 / 3
	runeCount := 6 + int(rarity)*2
	for i := 0; i < runeCount; i++ {
		angle := float64(i) * 2 * math.Pi / float64(runeCount)
		rx := cx + int(float64(runeCircleRadius)*math.Cos(angle))
		ry := cy + int(float64(runeCircleRadius)*math.Sin(angle))

		runeSize := 2
		for dy := -runeSize; dy <= runeSize; dy++ {
			for dx := -runeSize; dx <= runeSize; dx++ {
				if dx*dx+dy*dy <= runeSize*runeSize {
					img.Set(rx+dx, ry+dy, vs.getRuneColor(rarity))
				}
			}
		}
	}

	vs.addSparkles(img, rng, size, 4+int(rarity)*2)
}

// drawConsumable renders food/consumable items.
func (vs *VisualSystem) drawConsumable(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	itemColor := color.RGBA{
		uint8(150 + rng.Intn(80)),
		uint8(100 + rng.Intn(80)),
		uint8(50 + rng.Intn(50)),
		255,
	}

	cx, cy := size/2, size/2
	itemRadius := size / 3

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := x - cx
			dy := y - cy
			dist := math.Sqrt(float64(dx*dx + dy*dy))

			if dist < float64(itemRadius) {
				brightness := 1.0 - dist/float64(itemRadius)*0.4
				brightness += 0.2 * math.Sin(float64(x+y)*0.5)
				img.Set(x, y, applyShade(itemColor, brightness))
			}
		}
	}
}

// drawGeneric renders a default item sprite.
func (vs *VisualSystem) drawGeneric(img *ebiten.Image, rarity Rarity, rng *rand.Rand, size int) {
	itemColor := color.RGBA{150, 150, 150, 255}

	cx, cy := size/2, size/2
	boxSize := size / 2

	for y := cy - boxSize/2; y < cy+boxSize/2; y++ {
		for x := cx - boxSize/2; x < cx+boxSize/2; x++ {
			if x >= 0 && x < size && y >= 0 && y < size {
				dx := common.Abs(x - cx)
				dy := common.Abs(y - cy)
				brightness := 1.0 - float64(dx+dy)/float64(boxSize)*0.3
				img.Set(x, y, applyShade(itemColor, brightness))
			}
		}
	}
}

// Helper functions

func (vs *VisualSystem) getPotionColor(rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{255, 50, 50, 255},   // Red (health)
		{50, 100, 255, 255},  // Blue (mana)
		{100, 255, 100, 255}, // Green (poison/antidote)
		{255, 200, 50, 255},  // Gold (buff)
		{200, 50, 255, 255},  // Purple (magic)
	}
	return colors[rng.Intn(len(colors))]
}

func (vs *VisualSystem) getMetalColor(rarity Rarity, rng *rand.Rand) color.RGBA {
	switch rarity {
	case RarityLegendary:
		return color.RGBA{255, 215, 0, 255} // Gold
	case RarityRare:
		return color.RGBA{192, 192, 192, 255} // Silver
	case RarityUncommon:
		return color.RGBA{205, 127, 50, 255} // Bronze
	default:
		return color.RGBA{120, 120, 130, 255} // Iron
	}
}

func (vs *VisualSystem) getRuneColor(rarity Rarity) color.RGBA {
	switch rarity {
	case RarityLegendary:
		return color.RGBA{255, 215, 0, 255}
	case RarityRare:
		return color.RGBA{200, 100, 255, 255}
	case RarityUncommon:
		return color.RGBA{100, 150, 255, 255}
	default:
		return color.RGBA{80, 80, 100, 255}
	}
}

func (vs *VisualSystem) getArtifactColor(rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{150, 50, 200, 255},  // Purple
		{50, 200, 200, 255},  // Cyan
		{200, 100, 50, 255},  // Orange
		{50, 255, 150, 255},  // Emerald
		{255, 100, 150, 255}, // Pink
	}
	return colors[rng.Intn(len(colors))]
}

func (vs *VisualSystem) addSparkles(img *ebiten.Image, rng *rand.Rand, size, count int) {
	sparkleColor := color.RGBA{255, 255, 255, 200}
	for i := 0; i < count; i++ {
		sx := rng.Intn(size)
		sy := rng.Intn(size)
		img.Set(sx, sy, sparkleColor)
		if sx > 0 {
			img.Set(sx-1, sy, sparkleColor)
		}
		if sx < size-1 {
			img.Set(sx+1, sy, sparkleColor)
		}
		if sy > 0 {
			img.Set(sx, sy-1, sparkleColor)
		}
		if sy < size-1 {
			img.Set(sx, sy+1, sparkleColor)
		}
	}
}

func (vs *VisualSystem) addEnchantmentGlow(img *ebiten.Image, rarity Rarity, size int) {
	glowColor := color.RGBA{100, 200, 255, 100}
	if rarity == RarityLegendary {
		glowColor = color.RGBA{255, 215, 0, 120}
	}

	cx, cy := size/2, size/2
	glowRadius := size / 2

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			existing := img.At(x, y)
			_, _, _, a := existing.RGBA()
			if a > 0 {
				dx := x - cx
				dy := y - cy
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < float64(glowRadius) {
					glowAlpha := uint8(float64(glowColor.A) * (1.0 - dist/float64(glowRadius)))
					overlayColor := color.RGBA{glowColor.R, glowColor.G, glowColor.B, glowAlpha}
					img.Set(x, y, blendColors(existing.(color.RGBA), overlayColor))
				}
			}
		}
	}
}

func applyShade(baseColor color.RGBA, brightness float64) color.RGBA {
	brightness = math.Max(0.0, math.Min(1.5, brightness))
	return color.RGBA{
		uint8(float64(baseColor.R) * brightness),
		uint8(float64(baseColor.G) * brightness),
		uint8(float64(baseColor.B) * brightness),
		baseColor.A,
	}
}

func blendColors(base, overlay color.RGBA) color.RGBA {
	alpha := float64(overlay.A) / 255.0
	return color.RGBA{
		uint8(float64(base.R)*(1-alpha) + float64(overlay.R)*alpha),
		uint8(float64(base.G)*(1-alpha) + float64(overlay.G)*alpha),
		uint8(float64(base.B)*(1-alpha) + float64(overlay.B)*alpha),
		base.A,
	}
}

// CategorizeItem determines the visual category from an item ID.
func CategorizeItem(itemID string) ItemCategory {
	// Heuristic categorization based on common naming patterns
	if contains(itemID, "potion") || contains(itemID, "elixir") || contains(itemID, "flask") {
		return CategoryPotion
	}
	if contains(itemID, "scroll") || contains(itemID, "tome") || contains(itemID, "book") {
		return CategoryScroll
	}
	if contains(itemID, "sword") || contains(itemID, "axe") || contains(itemID, "bow") || contains(itemID, "weapon") {
		return CategoryWeapon
	}
	if contains(itemID, "armor") || contains(itemID, "shield") || contains(itemID, "helm") || contains(itemID, "plate") {
		return CategoryArmor
	}
	if contains(itemID, "gold") || contains(itemID, "coin") || contains(itemID, "money") {
		return CategoryGold
	}
	if contains(itemID, "gear") || contains(itemID, "circuit") || contains(itemID, "chip") || contains(itemID, "tech") {
		return CategoryGear
	}
	if contains(itemID, "artifact") || contains(itemID, "relic") || contains(itemID, "enchanted") || contains(itemID, "blessed") {
		return CategoryArtifact
	}
	if contains(itemID, "food") || contains(itemID, "bread") || contains(itemID, "meat") {
		return CategoryConsumable
	}
	return CategoryConsumable
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

// SpawnLootVisual creates a loot visual entity and adds it to the world.
func SpawnLootVisual(world *engine.World, itemID string, rarity Rarity, x, y float64, seed int64) engine.Entity {
	category := CategorizeItem(itemID)

	ent := world.AddEntity()

	pos := &PositionComponent{X: x, Y: y}
	world.AddComponent(ent, pos)

	visual := &VisualComponent{
		ItemID:    itemID,
		Category:  category,
		Rarity:    rarity,
		Seed:      seed,
		BobPhase:  0,
		GlowPhase: 0,
		Collected: false,
		SpawnTime: 0,
	}
	world.AddComponent(ent, visual)

	return ent
}

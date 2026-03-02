// Package sprite provides procedural sprite generation with shading and detail.
package sprite

import (
	"container/list"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/pool"
)

// SpriteType identifies the category of sprite to generate.
type SpriteType int

const (
	SpriteEnemy SpriteType = iota
	SpriteProp
	SpriteLoreItem
	SpriteDestructible
	SpritePickup
	SpriteProjectile
)

// SpriteKey uniquely identifies a cached sprite.
type SpriteKey struct {
	Type    SpriteType
	Subtype string
	Seed    int64
	Frame   int
	Size    int
}

// CachedSprite stores a generated sprite with metadata.
type CachedSprite struct {
	Image     *ebiten.Image
	Key       SpriteKey
	AccessCnt int
}

// Generator creates procedural sprites with caching.
type Generator struct {
	cache      map[SpriteKey]*list.Element
	lruList    *list.List
	maxEntries int
	mu         sync.RWMutex
	genreID    string
}

// NewGenerator creates a sprite generator with LRU cache.
func NewGenerator(maxCacheEntries int) *Generator {
	return &Generator{
		cache:      make(map[SpriteKey]*list.Element),
		lruList:    list.New(),
		maxEntries: maxCacheEntries,
		genreID:    "fantasy",
	}
}

// SetGenre configures genre-specific sprite generation.
func (g *Generator) SetGenre(genreID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.genreID = genreID
	g.cache = make(map[SpriteKey]*list.Element)
	g.lruList = list.New()
}

// GetSprite retrieves or generates a sprite.
func (g *Generator) GetSprite(spriteType SpriteType, subtype string, seed int64, frame, size int) *ebiten.Image {
	key := SpriteKey{
		Type:    spriteType,
		Subtype: subtype,
		Seed:    seed,
		Frame:   frame,
		Size:    size,
	}

	g.mu.Lock()
	if elem, found := g.cache[key]; found {
		g.lruList.MoveToFront(elem)
		cached := elem.Value.(*CachedSprite)
		cached.AccessCnt++
		g.mu.Unlock()
		return cached.Image
	}
	g.mu.Unlock()

	img := g.generateSprite(spriteType, subtype, seed, frame, size)

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.lruList.Len() >= g.maxEntries {
		oldest := g.lruList.Back()
		if oldest != nil {
			g.lruList.Remove(oldest)
			delete(g.cache, oldest.Value.(*CachedSprite).Key)
		}
	}

	cached := &CachedSprite{
		Image:     img,
		Key:       key,
		AccessCnt: 1,
	}
	elem := g.lruList.PushFront(cached)
	g.cache[key] = elem

	return img
}

// generateSprite creates a new procedural sprite.
func (g *Generator) generateSprite(spriteType SpriteType, subtype string, seed int64, frame, size int) *ebiten.Image {
	rgba := pool.GlobalPools.Images.Get(size, size)
	rng := rand.New(rand.NewSource(seed))

	switch spriteType {
	case SpriteEnemy:
		g.generateEnemySprite(rgba, subtype, rng, frame)
	case SpriteProp:
		g.generatePropSprite(rgba, subtype, rng, frame)
	case SpriteLoreItem:
		g.generateLoreSprite(rgba, subtype, rng, frame)
	case SpriteDestructible:
		g.generateDestructibleSprite(rgba, subtype, rng, frame)
	case SpritePickup:
		g.generatePickupSprite(rgba, subtype, rng, frame)
	case SpriteProjectile:
		g.generateProjectileSprite(rgba, subtype, rng, frame)
	default:
		g.generateDefaultSprite(rgba, rng)
	}

	result := ebiten.NewImageFromImage(rgba)
	pool.GlobalPools.Images.Put(rgba)
	return result
}

// generatePropSprite creates prop sprites with genre-specific styling.
func (g *Generator) generatePropSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	switch subtype {
	case "barrel":
		g.drawBarrel(img, cx, cy, size, rng)
	case "crate":
		g.drawCrate(img, cx, cy, size, rng)
	case "table":
		g.drawTable(img, cx, cy, size, rng)
	case "terminal":
		g.drawTerminal(img, cx, cy, size, rng)
	case "bones":
		g.drawBones(img, cx, cy, size, rng)
	case "plant":
		g.drawPlant(img, cx, cy, size, rng)
	case "pillar":
		g.drawPillar(img, cx, cy, size, rng)
	case "torch":
		g.drawTorch(img, cx, cy, size, rng, frame)
	case "debris":
		g.drawDebris(img, cx, cy, size, rng)
	case "container":
		g.drawContainer(img, cx, cy, size, rng)
	default:
		g.drawCrate(img, cx, cy, size, rng)
	}
}

// drawBarrel renders a cylindrical barrel with wood grain and metal bands.
func (g *Generator) drawBarrel(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	baseColor := g.getGenreWoodColor()
	metalColor := color.RGBA{R: 80, G: 80, B: 90, A: 255}

	radius := size / 3
	height := size * 2 / 3

	for y := cy - height/2; y < cy+height/2; y++ {
		bulge := int(float64(radius) * (1.0 + 0.2*math.Sin(float64(y-cy+height/2)/float64(height)*math.Pi)))
		for x := cx - bulge; x < cx+bulge; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			dist := math.Abs(float64(x - cx))
			shade := 1.0 - dist/float64(bulge)*0.5

			woodGrain := math.Sin(float64(y)*0.3+rng.Float64()*0.5) * 0.1
			shade += woodGrain

			r := uint8(float64(baseColor.R) * shade)
			g := uint8(float64(baseColor.G) * shade)
			b := uint8(float64(baseColor.B) * shade)

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	bandY1 := cy - height/3
	bandY2 := cy + height/3
	for _, by := range []int{bandY1, bandY2} {
		for y := by - 2; y < by+2; y++ {
			for x := cx - radius; x < cx+radius; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, metalColor)
				}
			}
		}
	}
}

// drawCrate renders a wooden crate with planks and corner reinforcements.
func (g *Generator) drawCrate(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	baseColor := g.getGenreWoodColor()
	darkColor := color.RGBA{
		R: baseColor.R / 2,
		G: baseColor.G / 2,
		B: baseColor.B / 2,
		A: 255,
	}

	boxSize := size * 2 / 3
	x1, y1 := cx-boxSize/2, cy-boxSize/2
	x2, y2 := cx+boxSize/2, cy+boxSize/2

	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			distFromEdge := math.Min(math.Abs(dx), math.Abs(dy))
			shade := 0.7 + 0.3*distFromEdge/float64(boxSize/2)

			plankNoise := rng.Float64()*0.1 - 0.05
			shade += plankNoise

			r := uint8(math.Min(255, float64(baseColor.R)*shade))
			g := uint8(math.Min(255, float64(baseColor.G)*shade))
			b := uint8(math.Min(255, float64(baseColor.B)*shade))

			plankPhase := int(math.Floor(float64(y-y1) / 8))
			if plankPhase%2 == 0 {
				r = uint8(float64(r) * 0.9)
				g = uint8(float64(g) * 0.9)
				b = uint8(float64(b) * 0.9)
			}

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	for i := 0; i < 4; i++ {
		cornerX := x1
		cornerY := y1
		if i%2 == 1 {
			cornerX = x2 - 4
		}
		if i/2 == 1 {
			cornerY = y2 - 4
		}
		fillRect(img, cornerX, cornerY, cornerX+4, cornerY+4, darkColor)
	}

	for x := x1; x < x2; x++ {
		img.Set(x, y1, darkColor)
		img.Set(x, y2-1, darkColor)
	}
	for y := y1; y < y2; y++ {
		img.Set(x1, y, darkColor)
		img.Set(x2-1, y, darkColor)
	}
}

// drawTable renders a table sprite with perspective.
func (g *Generator) drawTable(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	woodColor := g.getGenreWoodColor()

	tableW := size * 3 / 4
	tableH := size / 4
	tableY := cy - size/6

	for y := tableY; y < tableY+tableH; y++ {
		for x := cx - tableW/2; x < cx+tableW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.8 + 0.2*float64(y-tableY)/float64(tableH)
				r := uint8(float64(woodColor.R) * shade)
				g := uint8(float64(woodColor.G) * shade)
				b := uint8(float64(woodColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	legColor := color.RGBA{R: woodColor.R / 2, G: woodColor.G / 2, B: woodColor.B / 2, A: 255}
	legPositions := [][2]int{
		{cx - tableW/2 + 4, tableY + tableH},
		{cx + tableW/2 - 8, tableY + tableH},
	}

	for _, legPos := range legPositions {
		lx, ly := legPos[0], legPos[1]
		fillRect(img, lx, ly, lx+4, ly+size/3, legColor)
	}
}

// drawTerminal renders a sci-fi terminal with screen and panel.
func (g *Generator) drawTerminal(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	panelColor := color.RGBA{R: 40, G: 45, B: 50, A: 255}
	screenColor := color.RGBA{R: 0, G: 100, B: 120, A: 255}
	glowColor := color.RGBA{R: 0, G: 200, B: 255, A: 255}

	termW := size * 2 / 3
	termH := size * 3 / 4
	x1, y1 := cx-termW/2, cy-termH/2

	fillRect(img, x1, y1, x1+termW, y1+termH, panelColor)

	screenX := x1 + termW/6
	screenY := y1 + termH/6
	screenW := termW * 2 / 3
	screenH := termH / 2

	for y := screenY; y < screenY+screenH; y++ {
		for x := screenX; x < screenX+screenW; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				scanline := math.Sin(float64(y)*0.5) * 0.15
				r := uint8(float64(screenColor.R) * (1.0 + scanline))
				g := uint8(float64(screenColor.G) * (1.0 + scanline))
				b := uint8(float64(screenColor.B) * (1.0 + scanline))
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	for i := 0; i < 3; i++ {
		lx := screenX + i*screenW/4 + 4
		ly := screenY + screenH/4
		fillCircle(img, lx, ly, 2, glowColor)
	}
}

// drawBones renders skeletal remains.
func (g *Generator) drawBones(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	boneColor := color.RGBA{R: 220, G: 210, B: 190, A: 255}

	for i := 0; i < 3; i++ {
		angle := rng.Float64() * 2 * math.Pi
		length := size/3 + rng.Intn(size/4)
		x1 := cx + int(math.Cos(angle)*float64(length/4))
		y1 := cy + int(math.Sin(angle)*float64(length/4))
		x2 := cx + int(math.Cos(angle)*float64(length))
		y2 := cy + int(math.Sin(angle)*float64(length))

		drawThickLine(img, x1, y1, x2, y2, 2, boneColor)
	}

	fillCircle(img, cx, cy, size/8, boneColor)
}

// drawPlant renders foliage with leaves.
func (g *Generator) drawPlant(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	leafColor := g.getGenreLeafColor()
	stemColor := color.RGBA{R: 60, G: 80, B: 40, A: 255}

	stemX := cx
	stemBot := cy + size/3
	stemTop := cy - size/3

	drawThickLine(img, stemX, stemBot, stemX, stemTop, 2, stemColor)

	for i := 0; i < 5; i++ {
		leafY := stemBot - i*size/7
		leafSize := 6 - i
		if leafSize < 2 {
			leafSize = 2
		}

		fillCircle(img, stemX-leafSize, leafY, leafSize, leafColor)
		fillCircle(img, stemX+leafSize, leafY, leafSize, leafColor)
	}
}

// drawPillar renders a stone pillar with weathering.
func (g *Generator) drawPillar(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	stoneColor := g.getGenreStoneColor()

	pillarW := size / 3
	pillarH := size * 3 / 4
	x1 := cx - pillarW/2
	y1 := cy - pillarH/2

	for y := y1; y < y1+pillarH; y++ {
		for x := x1; x < x1+pillarW; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			dx := float64(x - cx)
			shade := 1.0 - math.Abs(dx)/float64(pillarW)*0.5

			noise := (rng.Float64() - 0.5) * 0.2
			shade += noise

			r := uint8(math.Max(0, math.Min(255, float64(stoneColor.R)*shade)))
			g := uint8(math.Max(0, math.Min(255, float64(stoneColor.G)*shade)))
			b := uint8(math.Max(0, math.Min(255, float64(stoneColor.B)*shade)))

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	capH := pillarH / 10
	for y := y1; y < y1+capH; y++ {
		for x := x1 - 2; x < x1+pillarW+2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, stoneColor)
			}
		}
	}
}

// drawTorch renders an animated torch with flame.
func (g *Generator) drawTorch(img *image.RGBA, cx, cy, size int, rng *rand.Rand, frame int) {
	handleColor := color.RGBA{R: 80, G: 60, B: 40, A: 255}
	fireColor := color.RGBA{R: 255, G: 140, B: 0, A: 255}
	glowColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}

	handleW := 4
	handleH := size / 2
	handleX := cx - handleW/2
	handleY := cy

	fillRect(img, handleX, handleY, handleX+handleW, handleY+handleH, handleColor)

	flameY := cy - size/8
	flameFlicker := int(math.Sin(float64(frame)*0.2) * 3)
	flameH := size/4 + flameFlicker

	for i := flameH; i > 0; i-- {
		flameSize := 3 + i/3
		y := flameY - i
		c := fireColor
		if i > flameH/2 {
			c = glowColor
		}
		fillCircle(img, cx, y, flameSize, c)
	}
}

// drawDebris renders scattered rubble.
func (g *Generator) drawDebris(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	debrisColor := color.RGBA{R: 100, G: 95, B: 90, A: 255}

	for i := 0; i < 6; i++ {
		dx := rng.Intn(size/2) - size/4
		dy := rng.Intn(size/2) - size/4
		dsize := 3 + rng.Intn(5)

		shade := 0.6 + rng.Float64()*0.4
		r := uint8(float64(debrisColor.R) * shade)
		g := uint8(float64(debrisColor.G) * shade)
		b := uint8(float64(debrisColor.B) * shade)

		fillCircle(img, cx+dx, cy+dy, dsize, color.RGBA{R: r, G: g, B: b, A: 255})
	}
}

// drawContainer renders a futuristic storage container.
func (g *Generator) drawContainer(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	baseColor := color.RGBA{R: 180, G: 180, B: 190, A: 255}
	accentColor := color.RGBA{R: 255, G: 200, B: 0, A: 255}

	boxSize := size * 2 / 3
	x1, y1 := cx-boxSize/2, cy-boxSize/2

	fillRect(img, x1, y1, x1+boxSize, y1+boxSize, baseColor)

	for i := 0; i < 3; i++ {
		stripeY := y1 + i*boxSize/4 + boxSize/8
		fillRect(img, x1, stripeY, x1+boxSize, stripeY+2, accentColor)
	}

	handleW := boxSize / 6
	handleY := y1 + boxSize/3
	fillRect(img, cx-handleW/2, handleY, cx+handleW/2, handleY+4, color.RGBA{R: 100, G: 100, B: 110, A: 255})
}

// generateLoreSprite creates lore item sprites.
func (g *Generator) generateLoreSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	switch subtype {
	case "note":
		g.drawNote(img, cx, cy, size, rng)
	case "audiolog":
		g.drawAudioLog(img, cx, cy, size, rng)
	case "graffiti":
		g.drawGraffiti(img, cx, cy, size, rng)
	case "body":
		g.drawBodyArrangement(img, cx, cy, size, rng)
	default:
		g.drawNote(img, cx, cy, size, rng)
	}
}

// drawNote renders a paper document.
func (g *Generator) drawNote(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	paperColor := color.RGBA{R: 240, G: 235, B: 210, A: 255}
	textColor := color.RGBA{R: 40, G: 40, B: 50, A: 255}

	noteW := size / 2
	noteH := size * 2 / 3
	x1, y1 := cx-noteW/2, cy-noteH/2

	fillRect(img, x1, y1, x1+noteW, y1+noteH, paperColor)

	for i := 0; i < 5; i++ {
		lineY := y1 + noteH/6 + i*noteH/8
		fillRect(img, x1+4, lineY, x1+noteW-4, lineY+1, textColor)
	}
}

// drawAudioLog renders a recording device.
func (g *Generator) drawAudioLog(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	deviceColor := color.RGBA{R: 60, G: 70, B: 80, A: 255}
	ledColor := color.RGBA{R: 0, G: 255, B: 100, A: 255}

	deviceSize := size / 2
	x1, y1 := cx-deviceSize/2, cy-deviceSize/2

	fillRect(img, x1, y1, x1+deviceSize, y1+deviceSize, deviceColor)

	fillCircle(img, cx, cy-deviceSize/4, 3, ledColor)
}

// drawGraffiti renders wall art.
func (g *Generator) drawGraffiti(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	graffitiColor := color.RGBA{R: 255, G: 50, B: 50, A: 255}

	for i := 0; i < 4; i++ {
		angle := float64(i) * math.Pi / 2
		x := cx + int(math.Cos(angle)*float64(size/4))
		y := cy + int(math.Sin(angle)*float64(size/4))
		fillCircle(img, x, y, 3, graffitiColor)
	}

	fillCircle(img, cx, cy, size/6, graffitiColor)
}

// drawBodyArrangement renders skeletal arrangement.
func (g *Generator) drawBodyArrangement(img *image.RGBA, cx, cy, size int, rng *rand.Rand) {
	g.drawBones(img, cx, cy, size, rng)
}

// generateDestructibleSprite creates destructible object sprites.
func (g *Generator) generateDestructibleSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	switch subtype {
	case "barrel":
		g.drawBarrel(img, cx, cy, size, rng)
	case "crate":
		g.drawCrate(img, cx, cy, size, rng)
	default:
		g.drawCrate(img, cx, cy, size, rng)
	}
}

// generatePickupSprite creates item pickup sprites.
func (g *Generator) generatePickupSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	pulse := 1.0 + 0.1*math.Sin(float64(frame)*0.1)

	switch subtype {
	case "health":
		g.drawHealthPickup(img, cx, cy, size, pulse)
	case "ammo":
		g.drawAmmoPickup(img, cx, cy, size, pulse)
	case "armor":
		g.drawArmorPickup(img, cx, cy, size, pulse)
	default:
		g.drawHealthPickup(img, cx, cy, size, pulse)
	}
}

// drawHealthPickup renders a health pack with cross.
func (g *Generator) drawHealthPickup(img *image.RGBA, cx, cy, size int, pulse float64) {
	healthColor := color.RGBA{R: uint8(float64(255) * pulse), G: 50, B: 50, A: 255}

	crossW := size / 3
	crossH := size / 2

	fillRect(img, cx-crossW/2, cy-crossH/2, cx+crossW/2, cy+crossH/2, healthColor)
	fillRect(img, cx-crossH/2, cy-crossW/2, cx+crossH/2, cy+crossW/2, healthColor)
}

// drawAmmoPickup renders ammunition box.
func (g *Generator) drawAmmoPickup(img *image.RGBA, cx, cy, size int, pulse float64) {
	ammoColor := color.RGBA{R: uint8(float64(200) * pulse), G: uint8(float64(150) * pulse), B: 0, A: 255}

	boxSize := size / 2
	fillRect(img, cx-boxSize/2, cy-boxSize/2, cx+boxSize/2, cy+boxSize/2, ammoColor)
}

// drawArmorPickup renders armor plate.
func (g *Generator) drawArmorPickup(img *image.RGBA, cx, cy, size int, pulse float64) {
	armorColor := color.RGBA{R: uint8(float64(100) * pulse), G: uint8(float64(150) * pulse), B: uint8(float64(200) * pulse), A: 255}

	shieldRadius := size / 3
	fillCircle(img, cx, cy, shieldRadius, armorColor)
}

// generateProjectileSprite creates projectile sprites.
func (g *Generator) generateProjectileSprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bulletColor := color.RGBA{R: 255, G: 220, B: 100, A: 255}
	fillCircle(img, cx, cy, size/4, bulletColor)
}

// generateEnemySprite creates enemy sprites with body plan variety.
func (g *Generator) generateEnemySprite(img *image.RGBA, subtype string, rng *rand.Rand, frame int) {
	switch subtype {
	case "humanoid", "tank", "ranged", "healer", "ambusher", "scout":
		g.generateHumanoidEnemy(img, subtype, rng, frame)
	case "quadruped":
		g.generateQuadrupedEnemy(img, rng, frame)
	case "insect":
		g.generateInsectEnemy(img, rng, frame)
	case "serpent":
		g.generateSerpentEnemy(img, rng, frame)
	case "flying":
		g.generateFlyingEnemy(img, rng, frame)
	case "amorphous":
		g.generateAmorphousEnemy(img, rng, frame)
	default:
		g.generateHumanoidEnemy(img, subtype, rng, frame)
	}
}

// generateHumanoidEnemy creates genre-aware humanoid enemy sprites.
func (g *Generator) generateHumanoidEnemy(img *image.RGBA, role string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	var armorColor, accentColor, skinColor color.RGBA
	var weaponType int

	switch g.genreID {
	case "scifi":
		armorColor = color.RGBA{R: 40, G: 60, B: 80, A: 255}
		accentColor = color.RGBA{R: 80, G: 200, B: 240, A: 255}
		skinColor = color.RGBA{R: 100, G: 180, B: 255, A: 255}
		weaponType = 2
	case "horror":
		armorColor = color.RGBA{R: 60, G: 20, B: 20, A: 255}
		accentColor = color.RGBA{R: 255, G: 50, B: 50, A: 255}
		skinColor = color.RGBA{R: 180, G: 170, B: 160, A: 255}
		weaponType = 3
	case "cyberpunk":
		armorColor = color.RGBA{R: 30, G: 30, B: 35, A: 255}
		accentColor = color.RGBA{R: 255, G: 0, B: 128, A: 255}
		skinColor = color.RGBA{R: 0, G: 200, B: 255, A: 255}
		weaponType = 2
	case "postapoc":
		armorColor = color.RGBA{R: 100, G: 80, B: 60, A: 255}
		accentColor = color.RGBA{R: 110, G: 90, B: 70, A: 255}
		skinColor = color.RGBA{R: 190, G: 160, B: 140, A: 255}
		weaponType = 3
	default:
		armorColor = color.RGBA{R: 120, G: 120, B: 130, A: 255}
		accentColor = color.RGBA{R: 140, G: 140, B: 150, A: 255}
		skinColor = color.RGBA{R: 210, G: 180, B: 160, A: 255}
		weaponType = 1
	}

	leftLegY := cy + size/6
	rightLegY := cy + size/6
	leftArmY := cy - size/10
	rightArmY := cy - size/10
	bodyY := cy - size/8

	if frame%3 == 1 {
		leftLegY += 2
		rightLegY -= 2
	} else if frame%3 == 2 {
		leftLegY -= 2
		rightLegY += 2
	}

	if role == "ambusher" && frame == 0 {
		bodyY += 2
		leftArmY += 2
		rightArmY += 2
	}

	legW := size / 12
	legH := size / 4

	for y := leftLegY; y < leftLegY+legH; y++ {
		for x := cx - size/8 - legW/2; x < cx-size/8+legW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.7 + 0.3*float64(y-leftLegY)/float64(legH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	for y := rightLegY; y < rightLegY+legH; y++ {
		for x := cx + size/8 - legW/2; x < cx+size/8+legW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.7 + 0.3*float64(y-rightLegY)/float64(legH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	torsoW := size / 3
	torsoH := size / 3
	for y := bodyY; y < bodyY+torsoH; y++ {
		for x := cx - torsoW/2; x < cx+torsoW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dx := float64(x - cx)
				shade := 1.0 - math.Abs(dx)/float64(torsoW)*0.4
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	if role == "tank" || role == "healer" {
		accentY := bodyY + torsoH/3
		accentH := 2
		fillRect(img, cx-torsoW/3, accentY, cx+torsoW/3, accentY+accentH, accentColor)
	}

	armW := legW - 1
	armH := size / 4
	for y := leftArmY; y < leftArmY+armH; y++ {
		for x := cx - torsoW/2 - armW; x < cx-torsoW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.8 + 0.2*float64(y-leftArmY)/float64(armH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	attackOffset := 0
	if frame%4 == 3 {
		attackOffset = -3
	}

	for y := rightArmY + attackOffset; y < rightArmY+armH+attackOffset; y++ {
		for x := cx + torsoW/2; x < cx+torsoW/2+armW; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				shade := 0.8 + 0.2*float64(y-rightArmY-attackOffset)/float64(armH)
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	headRadius := size / 10
	for y := -headRadius; y <= headRadius; y++ {
		for x := -headRadius; x <= headRadius; x++ {
			if x*x+y*y <= headRadius*headRadius {
				px := cx + x
				py := bodyY - size/16 + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					dist := math.Sqrt(float64(x*x + y*y))
					shade := 1.0 - (dist / float64(headRadius) * 0.3)
					r := uint8(math.Min(255, float64(skinColor.R)*shade))
					g := uint8(math.Min(255, float64(skinColor.G)*shade))
					b := uint8(math.Min(255, float64(skinColor.B)*shade))
					img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}

	weaponColor := color.RGBA{R: 200, G: 200, B: 220, A: 255}
	if weaponType == 2 {
		weaponColor = color.RGBA{R: 60, G: 60, B: 70, A: 255}
	} else if weaponType == 3 {
		weaponColor = color.RGBA{R: 120, G: 100, B: 80, A: 255}
	}

	weaponX := cx + torsoW/2 + armW
	weaponY := rightArmY + armH/2 + attackOffset
	weaponLen := size / 5
	if weaponType == 2 {
		fillRect(img, weaponX, weaponY-1, weaponX+weaponLen, weaponY+1, weaponColor)
		fillCircle(img, weaponX+weaponLen, weaponY, 2, color.RGBA{R: 100, G: 100, B: 110, A: 255})
	} else {
		fillRect(img, weaponX, weaponY-1, weaponX+weaponLen, weaponY+1, weaponColor)
		if weaponType == 1 {
			fillCircle(img, weaponX+weaponLen, weaponY, 3, color.RGBA{R: 180, G: 180, B: 200, A: 255})
		}
	}

	if role == "tank" {
		shieldX := cx - torsoW/2 - armW - 2
		shieldY := leftArmY + armH/4
		shieldH := armH / 2
		for y := shieldY; y < shieldY+shieldH; y++ {
			for x := shieldX - 4; x < shieldX; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, color.RGBA{R: 160, G: 140, B: 100, A: 255})
				}
			}
		}
	}

	if role == "healer" {
		symbolX := cx - 2
		symbolY := bodyY + torsoH/2 - 2
		fillRect(img, symbolX, symbolY-4, symbolX+4, symbolY+4, accentColor)
		fillRect(img, symbolX-4, symbolY, symbolX+8, symbolY+4, accentColor)
	}

	// Apply material detail for visual richness
	armorBounds := image.Rect(cx-torsoW/2, bodyY, cx+torsoW/2, bodyY+torsoH)
	g.applyMaterialDetail(img, armorBounds, MaterialMetal, rng.Int63(), 1.0, armorColor)

	// Add skin texture to head
	headBounds := image.Rect(cx-headRadius, bodyY-size/8-headRadius, cx+headRadius, bodyY-size/8+headRadius)
	g.applyMaterialDetail(img, headBounds, MaterialLeather, rng.Int63(), 0.5, skinColor)
}

// generateQuadrupedEnemy creates four-legged creature sprites.
func (g *Generator) generateQuadrupedEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("quadruped", rng)
	darkColor := color.RGBA{
		R: bodyColor.R / 2,
		G: bodyColor.G / 2,
		B: bodyColor.B / 2,
		A: 255,
	}

	bodyW := size / 2
	bodyH := size / 5
	bodyY := cy - size/10

	for y := bodyY; y < bodyY+bodyH; y++ {
		for x := cx - bodyW/2; x < cx+bodyW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dy := float64(y - bodyY)
				shade := 0.7 + 0.3*(1.0-dy/float64(bodyH))
				r := uint8(float64(bodyColor.R) * shade)
				g := uint8(float64(bodyColor.G) * shade)
				b := uint8(float64(bodyColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	legOffsets := []int{0, 0, 0, 0}
	if frame%3 == 1 {
		legOffsets = []int{2, -2, -2, 2}
	} else if frame%3 == 2 {
		legOffsets = []int{-2, 2, 2, -2}
	}

	legPositions := [][2]int{
		{cx - bodyW/3, bodyY + bodyH},
		{cx - bodyW/6, bodyY + bodyH},
		{cx + bodyW/6, bodyY + bodyH},
		{cx + bodyW/3, bodyY + bodyH},
	}

	for i, pos := range legPositions {
		legX := pos[0]
		legY := pos[1] + legOffsets[i]
		legW := size / 16
		legH := size / 4

		for y := legY; y < legY+legH; y++ {
			for x := legX - legW/2; x < legX+legW/2; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, darkColor)
				}
			}
		}
	}

	headW := bodyW / 3
	headH := size / 5
	headX := cx + bodyW/2
	headY := bodyY - headH/2

	for y := headY; y < headY+headH; y++ {
		for x := headX; x < headX+headW; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dx := float64(x - headX)
				shade := 1.0 - dx/float64(headW)*0.3
				r := uint8(float64(bodyColor.R) * shade)
				g := uint8(float64(bodyColor.G) * shade)
				b := uint8(float64(bodyColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	eyeColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}
	fillCircle(img, headX+headW-3, headY+headH/3, 2, eyeColor)

	tailX := cx - bodyW/2
	tailY := bodyY + bodyH/2
	tailAngle := float64(frame%8) * math.Pi / 16
	tailLen := size / 4
	tailEndX := tailX - int(float64(tailLen)*math.Cos(tailAngle))
	tailEndY := tailY + int(float64(tailLen)*math.Sin(tailAngle))
	drawThickLine(img, tailX, tailY, tailEndX, tailEndY, 2, darkColor)

	// Apply fur texture to body
	bodyBounds := image.Rect(cx-bodyW/2, bodyY, cx+bodyW/2, bodyY+bodyH)
	g.applyMaterialDetail(img, bodyBounds, MaterialFur, rng.Int63(), 0.8, bodyColor)
}

// generateInsectEnemy creates multi-legged insect sprites.
func (g *Generator) generateInsectEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("insect", rng)
	darkColor := color.RGBA{
		R: bodyColor.R / 3,
		G: bodyColor.G / 3,
		B: bodyColor.B / 3,
		A: 255,
	}

	segmentCount := 3
	segmentW := size / 4
	segmentH := size / 6

	for i := 0; i < segmentCount; i++ {
		segY := cy - size/6 + i*segmentH
		segW := segmentW - i*2
		if segW < segmentW/2 {
			segW = segmentW / 2
		}

		for y := segY; y < segY+segmentH-2; y++ {
			for x := cx - segW/2; x < cx+segW/2; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					dx := float64(x - cx)
					shade := 1.0 - math.Abs(dx)/float64(segW)*0.5
					r := uint8(float64(bodyColor.R) * shade)
					g := uint8(float64(bodyColor.G) * shade)
					b := uint8(float64(bodyColor.B) * shade)
					img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}

	legCount := 6
	for i := 0; i < legCount; i++ {
		side := 1
		if i%2 == 0 {
			side = -1
		}
		segIdx := i / 2
		legY := cy - size/6 + segIdx*segmentH + segmentH/2

		legOffset := 0
		if frame%3 == 1 && i%2 == 0 {
			legOffset = 2
		} else if frame%3 == 2 && i%2 == 1 {
			legOffset = 2
		}

		legStartX := cx + side*segmentW/2
		legMidX := legStartX + side*(size/6+legOffset)
		legMidY := legY - size/12
		legEndX := legMidX + side*size/12
		legEndY := legY + size/8

		drawThickLine(img, legStartX, legY, legMidX, legMidY, 1, darkColor)
		drawThickLine(img, legMidX, legMidY, legEndX, legEndY, 1, darkColor)
	}

	headRadius := size / 8
	headY := cy - size/6 - 2
	fillCircle(img, cx, headY, headRadius, bodyColor)

	eyeColor := color.RGBA{R: 255, G: 50, B: 50, A: 255}
	fillCircle(img, cx-headRadius/2, headY, 2, eyeColor)
	fillCircle(img, cx+headRadius/2, headY, 2, eyeColor)

	antennaLen := size / 5
	antennaAngle := float64(frame%8) * math.Pi / 32
	leftAntennaX := cx - headRadius/2 - int(float64(antennaLen)*math.Sin(antennaAngle))
	leftAntennaY := headY - headRadius - int(float64(antennaLen)*math.Cos(antennaAngle))
	rightAntennaX := cx + headRadius/2 + int(float64(antennaLen)*math.Sin(antennaAngle))
	rightAntennaY := headY - headRadius - int(float64(antennaLen)*math.Cos(antennaAngle))

	drawThickLine(img, cx-headRadius/2, headY-headRadius, leftAntennaX, leftAntennaY, 1, darkColor)
	drawThickLine(img, cx+headRadius/2, headY-headRadius, rightAntennaX, rightAntennaY, 1, darkColor)

	// Apply chitin texture to all segments
	for i := 0; i < segmentCount; i++ {
		segY := cy - size/6 + i*segmentH
		segW := segmentW - i*2
		if segW < segmentW/2 {
			segW = segmentW / 2
		}
		segBounds := image.Rect(cx-segW/2, segY, cx+segW/2, segY+segmentH-2)
		g.applyMaterialDetail(img, segBounds, MaterialChitin, rng.Int63()+int64(i), 0.9, bodyColor)
	}
}

// generateSerpentEnemy creates snake-like creature sprites.
func (g *Generator) generateSerpentEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("serpent", rng)
	scaleColor := color.RGBA{
		R: uint8(math.Min(255, float64(bodyColor.R)*1.2)),
		G: uint8(math.Min(255, float64(bodyColor.G)*1.2)),
		B: uint8(math.Min(255, float64(bodyColor.B)*1.2)),
		A: 255,
	}

	segments := 8
	baseRadius := size / 8
	wavePhase := float64(frame) * 0.3

	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments)
		segY := cy - size/4 + int(t*float64(size)*0.8)
		waveOffset := int(math.Sin(t*math.Pi*2+wavePhase) * float64(size) / 6)
		segX := cx + waveOffset

		radius := baseRadius - i
		if radius < 2 {
			radius = 2
		}

		for y := -radius; y <= radius; y++ {
			for x := -radius; x <= radius; x++ {
				if x*x+y*y <= radius*radius {
					px := segX + x
					py := segY + y
					if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
						dist := math.Sqrt(float64(x*x + y*y))
						shade := 1.0 - (dist / float64(radius) * 0.4)
						r := uint8(math.Min(255, float64(bodyColor.R)*shade))
						g := uint8(math.Min(255, float64(bodyColor.G)*shade))
						b := uint8(math.Min(255, float64(bodyColor.B)*shade))
						img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
					}
				}
			}
		}

		if i%2 == 0 && i < segments-1 {
			fillCircle(img, segX-radius/2, segY, 1, scaleColor)
			fillCircle(img, segX+radius/2, segY, 1, scaleColor)
		}
	}

	headRadius := baseRadius + 2
	headY := cy - size/4
	waveOffset := int(math.Sin(wavePhase) * float64(size) / 6)
	headX := cx + waveOffset

	for y := -headRadius; y <= headRadius; y++ {
		for x := -headRadius; x <= headRadius; x++ {
			if x*x+y*y <= headRadius*headRadius {
				px := headX + x
				py := headY + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					dist := math.Sqrt(float64(x*x + y*y))
					shade := 1.2 - (dist / float64(headRadius) * 0.5)
					r := uint8(math.Min(255, float64(bodyColor.R)*shade))
					g := uint8(math.Min(255, float64(bodyColor.G)*shade))
					b := uint8(math.Min(255, float64(bodyColor.B)*shade))
					img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}

	eyeColor := color.RGBA{R: 255, G: 200, B: 0, A: 255}
	fillCircle(img, headX-headRadius/2, headY, 2, eyeColor)
	fillCircle(img, headX+headRadius/2, headY, 2, eyeColor)

	tongueLen := size / 8
	if frame%4 < 2 {
		tongueLen = size / 12
	}
	tongueColor := color.RGBA{R: 200, G: 50, B: 50, A: 255}
	drawThickLine(img, headX, headY+headRadius/2, headX, headY+headRadius/2+tongueLen, 1, tongueColor)

	// Apply scale texture to body segments
	fullBounds := image.Rect(cx-size/4, cy-size/4, cx+size/4, cy+size/2)
	g.applyMaterialDetail(img, fullBounds, MaterialScales, rng.Int63(), 1.0, bodyColor)
}

// generateFlyingEnemy creates winged creature sprites.
func (g *Generator) generateFlyingEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("flying", rng)
	wingColor := color.RGBA{
		R: bodyColor.R / 2,
		G: bodyColor.G / 2,
		B: bodyColor.B / 2,
		A: 200,
	}

	hoverOffset := 0
	if frame%4 < 2 {
		hoverOffset = -2
	} else {
		hoverOffset = 2
	}

	wingSpan := size / 2
	wingH := size / 4
	wingY := cy + hoverOffset - wingH/2

	wingAngle := 0.0
	if frame%4 < 2 {
		wingAngle = math.Pi / 6
	} else {
		wingAngle = -math.Pi / 6
	}

	for side := -1; side <= 1; side += 2 {
		wingCenterX := cx + side*size/8
		for y := 0; y < wingH; y++ {
			yOffset := float64(y) - float64(wingH)/2
			rotatedY := int(yOffset*math.Cos(wingAngle)) + wingY + wingH/2
			wingWidth := int(float64(wingSpan) * (1.0 - float64(y)/float64(wingH)))

			for x := 0; x < wingWidth; x++ {
				px := wingCenterX + side*x
				py := rotatedY
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					alpha := uint8(200 - uint8(float64(x)/float64(wingWidth)*150))
					img.Set(px, py, color.RGBA{R: wingColor.R, G: wingColor.G, B: wingColor.B, A: alpha})
				}
			}
		}
	}

	bodyW := size / 5
	bodyH := size / 3
	bodyY := cy + hoverOffset - bodyH/2

	for y := bodyY; y < bodyY+bodyH; y++ {
		for x := cx - bodyW/2; x < cx+bodyW/2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				dy := float64(y - bodyY)
				shade := 0.8 + 0.4*(1.0-math.Abs(dy-float64(bodyH)/2)/float64(bodyH))
				r := uint8(float64(bodyColor.R) * shade)
				g := uint8(float64(bodyColor.G) * shade)
				b := uint8(float64(bodyColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	headRadius := size / 9
	headY := bodyY - 2
	fillCircle(img, cx, headY, headRadius, bodyColor)

	eyeColor := color.RGBA{R: 255, G: 100, B: 0, A: 255}
	fillCircle(img, cx-headRadius/2, headY, 2, eyeColor)
	fillCircle(img, cx+headRadius/2, headY, 2, eyeColor)

	tailY := bodyY + bodyH
	tailLen := size / 5
	tailEndY := tailY + tailLen
	tailSway := int(math.Sin(float64(frame)*0.4) * 3)
	drawThickLine(img, cx, tailY, cx+tailSway, tailEndY, 2, bodyColor)

	// Apply membrane texture to wings
	wingBounds := image.Rect(cx-size/2, wingY, cx+size/2, wingY+wingH)
	g.applyMaterialDetail(img, wingBounds, MaterialMembrane, rng.Int63(), 0.6, wingColor)
}

// generateAmorphousEnemy creates blob/slime creature sprites.
func (g *Generator) generateAmorphousEnemy(img *image.RGBA, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	bodyColor := g.getCreatureColor("amorphous", rng)
	innerColor := color.RGBA{
		R: uint8(math.Min(255, float64(bodyColor.R)*1.3)),
		G: uint8(math.Min(255, float64(bodyColor.G)*1.3)),
		B: uint8(math.Min(255, float64(bodyColor.B)*1.3)),
		A: 255,
	}

	pulsePhase := float64(frame) * 0.2
	pulseAmount := 1.0 + math.Sin(pulsePhase)*0.15

	baseRadius := int(float64(size) / 3 * pulseAmount)

	blobPoints := 12
	radiusVariation := make([]float64, blobPoints)
	for i := 0; i < blobPoints; i++ {
		radiusVariation[i] = 0.8 + rng.Float64()*0.4
	}

	for y := -baseRadius; y <= baseRadius; y++ {
		for x := -baseRadius; x <= baseRadius; x++ {
			angle := math.Atan2(float64(y), float64(x))
			if angle < 0 {
				angle += 2 * math.Pi
			}
			pointIdx := int(angle / (2 * math.Pi / float64(blobPoints)))
			if pointIdx >= blobPoints {
				pointIdx = blobPoints - 1
			}

			maxDist := float64(baseRadius) * radiusVariation[pointIdx]
			dist := math.Sqrt(float64(x*x + y*y))

			if dist <= maxDist {
				px := cx + x
				py := cy + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					distRatio := dist / maxDist
					shade := 1.0 - distRatio*0.6

					c := bodyColor
					if distRatio < 0.3 {
						blend := distRatio / 0.3
						c.R = uint8(float64(innerColor.R)*(1-blend) + float64(bodyColor.R)*blend)
						c.G = uint8(float64(innerColor.G)*(1-blend) + float64(bodyColor.G)*blend)
						c.B = uint8(float64(innerColor.B)*(1-blend) + float64(bodyColor.B)*blend)
					}

					r := uint8(math.Min(255, float64(c.R)*shade))
					g := uint8(math.Min(255, float64(c.G)*shade))
					b := uint8(math.Min(255, float64(c.B)*shade))
					img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}

	eyeCount := 2 + rng.Intn(3)
	eyeColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	pupilColor := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	for i := 0; i < eyeCount; i++ {
		eyeAngle := float64(i) * 2 * math.Pi / float64(eyeCount)
		eyeDist := float64(baseRadius) / 3
		eyeX := cx + int(math.Cos(eyeAngle)*eyeDist)
		eyeY := cy + int(math.Sin(eyeAngle)*eyeDist) - baseRadius/4

		fillCircle(img, eyeX, eyeY, 4, eyeColor)
		fillCircle(img, eyeX, eyeY, 2, pupilColor)
	}

	if frame%8 < 4 {
		highlight1X := cx - baseRadius/3
		highlight1Y := cy - baseRadius/3
		fillCircle(img, highlight1X, highlight1Y, 3, color.RGBA{R: 255, G: 255, B: 255, A: 100})
	}
}

// getCreatureColor returns genre and creature-type specific colors.
func (g *Generator) getCreatureColor(creatureType string, rng *rand.Rand) color.RGBA {
	switch g.genreID {
	case "scifi":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 100, G: 120, B: 140, A: 255}
		case "insect":
			return color.RGBA{R: 80, G: 100, B: 120, A: 255}
		case "serpent":
			return color.RGBA{R: 60, G: 140, B: 160, A: 255}
		case "flying":
			return color.RGBA{R: 140, G: 120, B: 180, A: 255}
		case "amorphous":
			return color.RGBA{R: 0, G: 200, B: 150, A: 255}
		}
	case "horror":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 80, G: 60, B: 60, A: 255}
		case "insect":
			return color.RGBA{R: 60, G: 40, B: 40, A: 255}
		case "serpent":
			return color.RGBA{R: 90, G: 80, B: 70, A: 255}
		case "flying":
			return color.RGBA{R: 70, G: 50, B: 80, A: 255}
		case "amorphous":
			return color.RGBA{R: 100, G: 50, B: 80, A: 255}
		}
	case "cyberpunk":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 40, G: 40, B: 50, A: 255}
		case "insect":
			return color.RGBA{R: 255, G: 0, B: 128, A: 255}
		case "serpent":
			return color.RGBA{R: 0, G: 255, B: 200, A: 255}
		case "flying":
			return color.RGBA{R: 255, G: 100, B: 200, A: 255}
		case "amorphous":
			return color.RGBA{R: 128, G: 0, B: 255, A: 255}
		}
	case "postapoc":
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 120, G: 100, B: 80, A: 255}
		case "insect":
			return color.RGBA{R: 100, G: 80, B: 60, A: 255}
		case "serpent":
			return color.RGBA{R: 110, G: 90, B: 70, A: 255}
		case "flying":
			return color.RGBA{R: 90, G: 80, B: 70, A: 255}
		case "amorphous":
			return color.RGBA{R: 140, G: 100, B: 60, A: 255}
		}
	default:
		switch creatureType {
		case "quadruped":
			return color.RGBA{R: 140, G: 100, B: 60, A: 255}
		case "insect":
			return color.RGBA{R: 80, G: 120, B: 80, A: 255}
		case "serpent":
			return color.RGBA{R: 100, G: 140, B: 100, A: 255}
		case "flying":
			return color.RGBA{R: 120, G: 100, B: 140, A: 255}
		case "amorphous":
			return color.RGBA{R: 100, G: 200, B: 150, A: 255}
		}
	}
	return color.RGBA{R: 128, G: 128, B: 128, A: 255}
}

// generateDefaultSprite creates a fallback sprite.
func (g *Generator) generateDefaultSprite(img *image.RGBA, rng *rand.Rand) {
	size := img.Bounds().Dx()
	fillRect(img, 0, 0, size, size, color.RGBA{R: 128, G: 128, B: 128, A: 255})
}

// Genre-specific color helpers.

func (g *Generator) getGenreWoodColor() color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 100, G: 100, B: 110, A: 255}
	case "horror":
		return color.RGBA{R: 80, G: 60, B: 50, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 40, G: 40, B: 50, A: 255}
	case "postapoc":
		return color.RGBA{R: 110, G: 90, B: 70, A: 255}
	default:
		return color.RGBA{R: 139, G: 90, B: 43, A: 255}
	}
}

func (g *Generator) getGenreLeafColor() color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 0, G: 255, B: 150, A: 255}
	case "horror":
		return color.RGBA{R: 60, G: 80, B: 50, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 255, G: 0, B: 128, A: 255}
	default:
		return color.RGBA{R: 50, G: 150, B: 50, A: 255}
	}
}

func (g *Generator) getGenreStoneColor() color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 140, G: 150, B: 160, A: 255}
	case "horror":
		return color.RGBA{R: 80, G: 80, B: 90, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 60, G: 70, B: 80, A: 255}
	default:
		return color.RGBA{R: 120, G: 120, B: 130, A: 255}
	}
}

// Primitive drawing helpers.

func fillRect(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.Set(x, y, c)
				}
			}
		}
	}
}

func drawThickLine(img *image.RGBA, x1, y1, x2, y2, thickness int, c color.RGBA) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	for {
		for dt := -thickness / 2; dt <= thickness/2; dt++ {
			for dp := -thickness / 2; dp <= thickness/2; dp++ {
				px, py := x1+dt, y1+dp
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}

		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

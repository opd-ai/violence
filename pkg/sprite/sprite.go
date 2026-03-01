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
	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	rng := rand.New(rand.NewSource(seed))

	switch spriteType {
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

	return ebiten.NewImageFromImage(rgba)
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

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

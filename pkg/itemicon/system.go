// Package itemicon provides procedural item icon generation with rarity-based visual effects.
package itemicon

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/pool"
	"github.com/sirupsen/logrus"
)

// IconSystem generates and caches procedural item icons.
type IconSystem struct {
	cache   map[IconKey]*ebiten.Image
	mu      sync.RWMutex
	genre   string
	logger  *logrus.Entry
	maxSize int
}

// IconKey uniquely identifies a cached icon.
type IconKey struct {
	Seed         int64
	IconType     string
	Rarity       int
	SubType      string
	Size         int
	EnchantLevel int
}

// NewSystem creates an item icon generation system.
func NewSystem(genre string, maxCacheSize int) *IconSystem {
	return &IconSystem{
		cache:   make(map[IconKey]*ebiten.Image),
		genre:   genre,
		maxSize: maxCacheSize,
		logger: logrus.WithFields(logrus.Fields{
			"system": "itemicon",
			"genre":  genre,
		}),
	}
}

// SetGenre updates the genre setting.
func (s *IconSystem) SetGenre(genre string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.genre = genre
	s.cache = make(map[IconKey]*ebiten.Image)
}

// GenerateIcon creates or retrieves a cached item icon.
func (s *IconSystem) GenerateIcon(comp *ItemIconComponent) *ebiten.Image {
	key := IconKey{
		Seed:         comp.Seed,
		IconType:     comp.IconType,
		Rarity:       comp.Rarity,
		SubType:      comp.SubType,
		Size:         comp.IconSize,
		EnchantLevel: comp.EnchantLevel,
	}

	s.mu.RLock()
	if cached, found := s.cache[key]; found {
		s.mu.RUnlock()
		return cached
	}
	s.mu.RUnlock()

	icon := s.generateIcon(comp)

	s.mu.Lock()
	if len(s.cache) >= s.maxSize {
		for k := range s.cache {
			delete(s.cache, k)
			break
		}
	}
	s.cache[key] = icon
	s.mu.Unlock()

	return icon
}

// generateIcon renders a single item icon.
func (s *IconSystem) generateIcon(comp *ItemIconComponent) *ebiten.Image {
	size := comp.IconSize
	if size == 0 {
		size = 48
	}

	rgba := pool.GlobalPools.Images.Get(size, size)
	rng := rand.New(rand.NewSource(comp.Seed))

	switch comp.IconType {
	case "weapon":
		s.drawWeaponIcon(rgba, comp, rng)
	case "armor":
		s.drawArmorIcon(rgba, comp, rng)
	case "consumable":
		s.drawConsumableIcon(rgba, comp, rng)
	case "material":
		s.drawMaterialIcon(rgba, comp, rng)
	case "quest":
		s.drawQuestIcon(rgba, comp, rng)
	default:
		s.drawGenericIcon(rgba, comp, rng)
	}

	if comp.BorderGlow {
		s.addRarityBorder(rgba, comp.Rarity, size)
	}

	if comp.EnchantLevel > 0 {
		s.addEnchantmentEffect(rgba, comp.EnchantLevel, rng, size)
	}

	if comp.Durability < 1.0 {
		s.addWearEffect(rgba, comp.Durability, size)
	}

	result := ebiten.NewImageFromImage(rgba)
	pool.GlobalPools.Images.Put(rgba)
	return result
}

// drawWeaponIcon renders weapon-type item icons.
func (s *IconSystem) drawWeaponIcon(img *image.RGBA, comp *ItemIconComponent, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	baseColor := s.getRarityBaseColor(comp.Rarity)
	metalColor := s.getGenreMetalColor()
	handleColor := color.RGBA{R: 80, G: 50, B: 30, A: 255}

	switch comp.SubType {
	case "sword", "blade":
		bladeLen := size * 3 / 4
		bladeW := size / 8
		handleLen := size / 5

		for y := cy - bladeLen/2; y < cy+bladeLen/4; y++ {
			w := bladeW - int(math.Abs(float64(y-cy+bladeLen/2))/float64(bladeLen/2)*float64(bladeW/2))
			if w < 2 {
				w = 2
			}
			for x := cx - w/2; x < cx+w/2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					shade := 1.0 - math.Abs(float64(x-cx))/float64(w)*0.4
					r := uint8(float64(metalColor.R) * shade)
					g := uint8(float64(metalColor.G) * shade)
					b := uint8(float64(metalColor.B) * shade)
					img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}

		for x := cx - 1; x <= cx+1; x++ {
			for y := cy - bladeLen/2; y < cy+bladeLen/4; y++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					highlight := color.RGBA{R: 255, G: 255, B: 255, A: 180}
					img.Set(x, y, highlight)
				}
			}
		}

		guardY := cy + bladeLen/4
		fillRect(img, cx-size/4, guardY-2, cx+size/4, guardY+2, baseColor)

		for y := guardY; y < guardY+handleLen; y++ {
			for x := cx - bladeW/2; x < cx+bladeW/2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, handleColor)
				}
			}
		}

	case "axe":
		handleLen := size * 2 / 3
		for y := cy; y < cy+handleLen; y++ {
			for x := cx - 2; x < cx+2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, handleColor)
				}
			}
		}

		bladeW := size / 2
		bladeH := size / 3
		for y := cy - bladeH; y < cy+bladeH/3; y++ {
			for x := cx - bladeW/2; x < cx+bladeW/2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					dx := float64(x - cx)
					shade := 0.8 + 0.2*(1.0-math.Abs(dx)/float64(bladeW/2))
					r := uint8(float64(metalColor.R) * shade)
					g := uint8(float64(metalColor.G) * shade)
					b := uint8(float64(metalColor.B) * shade)
					img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}

	default:
		s.drawGenericWeapon(img, baseColor, metalColor, handleColor, rng)
	}
}

// drawGenericWeapon renders a simple generic weapon icon.
func (s *IconSystem) drawGenericWeapon(img *image.RGBA, base, metal, handle color.RGBA, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2
	for y := cy - size/3; y < cy+size/3; y++ {
		for x := cx - size/6; x < cx+size/6; x++ {
			if x >= 0 && x < size && y >= 0 && y < size {
				img.Set(x, y, metal)
			}
		}
	}
}

// drawArmorIcon renders armor-type item icons.
func (s *IconSystem) drawArmorIcon(img *image.RGBA, comp *ItemIconComponent, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	armorColor := s.getRarityBaseColor(comp.Rarity)
	accentColor := s.getGenreAccentColor()

	torsoW := size / 2
	torsoH := size / 2

	for y := cy - torsoH/2; y < cy+torsoH/2; y++ {
		for x := cx - torsoW/2; x < cx+torsoW/2; x++ {
			if x >= 0 && x < size && y >= 0 && y < size {
				dx := float64(x - cx)
				shade := 1.0 - math.Abs(dx)/float64(torsoW)*0.3
				r := uint8(float64(armorColor.R) * shade)
				g := uint8(float64(armorColor.G) * shade)
				b := uint8(float64(armorColor.B) * shade)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	shoulderY := cy - torsoH/2
	shoulderR := size / 8
	fillCircle(img, cx-torsoW/2, shoulderY, shoulderR, armorColor)
	fillCircle(img, cx+torsoW/2, shoulderY, shoulderR, armorColor)

	trimY1 := cy - torsoH/4
	trimY2 := cy + torsoH/4
	fillRect(img, cx-torsoW/3, trimY1, cx+torsoW/3, trimY1+2, accentColor)
	fillRect(img, cx-torsoW/3, trimY2, cx+torsoW/3, trimY2+2, accentColor)
}

// drawConsumableIcon renders potion/scroll/consumable icons.
func (s *IconSystem) drawConsumableIcon(img *image.RGBA, comp *ItemIconComponent, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	liquidColor := s.getRarityBaseColor(comp.Rarity)
	glassColor := color.RGBA{R: 180, G: 200, B: 220, A: 200}

	switch comp.SubType {
	case "potion":
		bottleW := size / 3
		bottleH := size * 2 / 3
		neckH := size / 6

		for y := cy + neckH; y < cy+bottleH/2; y++ {
			bulge := int(float64(bottleW/2) * (1.0 + 0.3*math.Sin(float64(y-cy-neckH)/float64(bottleH-neckH)*math.Pi)))
			for x := cx - bulge; x < cx+bulge; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, glassColor)
				}
			}
		}

		liquidLevel := cy + bottleH/2 - bottleH/4
		for y := liquidLevel; y < cy+bottleH/2-2; y++ {
			bulge := int(float64(bottleW/2) * (1.0 + 0.3*math.Sin(float64(y-cy-neckH)/float64(bottleH-neckH)*math.Pi)))
			for x := cx - bulge + 2; x < cx+bulge-2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, liquidColor)
				}
			}
		}

		neckW := bottleW / 3
		for y := cy; y < cy+neckH; y++ {
			for x := cx - neckW/2; x < cx+neckW/2; x++ {
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, glassColor)
				}
			}
		}

		fillRect(img, cx-neckW, cy-3, cx+neckW, cy+1, color.RGBA{R: 100, G: 60, B: 40, A: 255})

	case "scroll":
		scrollW := size * 2 / 3
		scrollH := size * 3 / 4
		paperColor := color.RGBA{R: 230, G: 215, B: 190, A: 255}

		fillRect(img, cx-scrollW/2, cy-scrollH/2, cx+scrollW/2, cy+scrollH/2, paperColor)

		for i := 0; i < 3; i++ {
			lineY := cy - scrollH/4 + i*scrollH/6
			fillRect(img, cx-scrollW/3, lineY, cx+scrollW/3, lineY+1, color.RGBA{R: 100, G: 80, B: 60, A: 180})
		}

		sealR := size / 8
		fillCircle(img, cx, cy+scrollH/3, sealR, liquidColor)

	default:
		fillCircle(img, cx, cy, size/3, liquidColor)
	}
}

// drawMaterialIcon renders crafting material icons.
func (s *IconSystem) drawMaterialIcon(img *image.RGBA, comp *ItemIconComponent, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	matColor := s.getRarityBaseColor(comp.Rarity)

	shardCnt := 3 + rng.Intn(3)
	for i := 0; i < shardCnt; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(shardCnt)
		x := cx + int(float64(size/6)*math.Cos(angle))
		y := cy + int(float64(size/6)*math.Sin(angle))
		shardSize := 3 + rng.Intn(4)

		for dy := -shardSize; dy < shardSize; dy++ {
			for dx := -shardSize; dx < shardSize; dx++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < size && py >= 0 && py < size {
					dist := math.Sqrt(float64(dx*dx + dy*dy))
					if dist < float64(shardSize) {
						shade := 1.0 - dist/float64(shardSize)*0.5
						r := uint8(math.Min(255, float64(matColor.R)*shade))
						g := uint8(math.Min(255, float64(matColor.G)*shade))
						b := uint8(math.Min(255, float64(matColor.B)*shade))
						img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
					}
				}
			}
		}
	}
}

// drawQuestIcon renders quest item icons.
func (s *IconSystem) drawQuestIcon(img *image.RGBA, comp *ItemIconComponent, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	questColor := color.RGBA{R: 255, G: 215, B: 0, A: 255}

	starR := size / 3
	points := 5
	for i := 0; i < points*2; i++ {
		angle := float64(i) * math.Pi / float64(points)
		r := starR
		if i%2 == 1 {
			r = starR / 2
		}
		x := cx + int(float64(r)*math.Cos(angle-math.Pi/2))
		y := cy + int(float64(r)*math.Sin(angle-math.Pi/2))

		nextI := (i + 1) % (points * 2)
		nextAngle := float64(nextI) * math.Pi / float64(points)
		nextR := starR
		if nextI%2 == 1 {
			nextR = starR / 2
		}
		nextX := cx + int(float64(nextR)*math.Cos(nextAngle-math.Pi/2))
		nextY := cy + int(float64(nextR)*math.Sin(nextAngle-math.Pi/2))

		drawLine(img, x, y, nextX, nextY, questColor)
	}

	fillCircle(img, cx, cy, starR/3, questColor)
}

// drawGenericIcon renders a fallback generic icon.
func (s *IconSystem) drawGenericIcon(img *image.RGBA, comp *ItemIconComponent, rng *rand.Rand) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	baseColor := s.getRarityBaseColor(comp.Rarity)
	fillRect(img, cx-size/3, cy-size/3, cx+size/3, cy+size/3, baseColor)

	outlineColor := color.RGBA{R: baseColor.R / 2, G: baseColor.G / 2, B: baseColor.B / 2, A: 255}
	drawRect(img, cx-size/3, cy-size/3, cx+size/3, cy+size/3, outlineColor)
}

// addRarityBorder adds a colored border based on rarity.
func (s *IconSystem) addRarityBorder(img *image.RGBA, rarity, size int) {
	borderColor := s.getRarityBorderColor(rarity)

	for i := 0; i < 2; i++ {
		drawRect(img, i, i, size-1-i, size-1-i, borderColor)
	}

	if rarity >= 3 {
		for x := 0; x < size; x++ {
			for y := 0; y < size; y++ {
				if x < 3 || x >= size-3 || y < 3 || y >= size-3 {
					existing := img.At(x, y)
					r, g, b, a := existing.RGBA()
					glowR := uint8((uint32(borderColor.R) + r>>8) / 2)
					glowG := uint8((uint32(borderColor.G) + g>>8) / 2)
					glowB := uint8((uint32(borderColor.B) + b>>8) / 2)
					glowA := uint8(math.Max(float64(borderColor.A), float64(a>>8)))
					img.Set(x, y, color.RGBA{R: glowR, G: glowG, B: glowB, A: glowA})
				}
			}
		}
	}
}

// addEnchantmentEffect adds magical sparkle effects for enchanted items.
func (s *IconSystem) addEnchantmentEffect(img *image.RGBA, level int, rng *rand.Rand, size int) {
	sparkleColor := s.getGenreEnchantColor()
	sparkleCount := level * 2

	for i := 0; i < sparkleCount; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)
		sparkleSize := 1 + rng.Intn(2)

		for dy := -sparkleSize; dy <= sparkleSize; dy++ {
			for dx := -sparkleSize; dx <= sparkleSize; dx++ {
				px, py := x+dx, y+dy
				if px >= 0 && px < size && py >= 0 && py < size {
					if dx*dx+dy*dy <= sparkleSize*sparkleSize {
						img.Set(px, py, sparkleColor)
					}
				}
			}
		}
	}
}

// addWearEffect adds visual damage/wear to items with low durability.
func (s *IconSystem) addWearEffect(img *image.RGBA, durability float64, size int) {
	if durability >= 1.0 {
		return
	}

	wearIntensity := 1.0 - durability
	darkColor := color.RGBA{R: 40, G: 40, B: 40, A: uint8(wearIntensity * 100)}

	scratchCount := int(wearIntensity * 10)
	for i := 0; i < scratchCount; i++ {
		x1, y1 := i*size/scratchCount, 0
		x2, y2 := x1+size/5, size
		drawLine(img, x1, y1, x2, y2, darkColor)
	}
}

// getRarityBaseColor returns the base color for an item based on rarity.
func (s *IconSystem) getRarityBaseColor(rarity int) color.RGBA {
	switch rarity {
	case 0:
		return color.RGBA{R: 150, G: 150, B: 150, A: 255}
	case 1:
		return color.RGBA{R: 100, G: 200, B: 100, A: 255}
	case 2:
		return color.RGBA{R: 80, G: 120, B: 255, A: 255}
	case 3:
		return color.RGBA{R: 180, G: 100, B: 255, A: 255}
	case 4:
		return color.RGBA{R: 255, G: 180, B: 50, A: 255}
	default:
		return color.RGBA{R: 200, G: 200, B: 200, A: 255}
	}
}

// getRarityBorderColor returns the border glow color based on rarity.
func (s *IconSystem) getRarityBorderColor(rarity int) color.RGBA {
	switch rarity {
	case 0:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	case 1:
		return color.RGBA{R: 50, G: 255, B: 50, A: 255}
	case 2:
		return color.RGBA{R: 50, G: 150, B: 255, A: 255}
	case 3:
		return color.RGBA{R: 200, G: 50, B: 255, A: 255}
	case 4:
		return color.RGBA{R: 255, G: 200, B: 0, A: 255}
	default:
		return color.RGBA{R: 150, G: 150, B: 150, A: 255}
	}
}

// getGenreMetalColor returns metal color based on genre.
func (s *IconSystem) getGenreMetalColor() color.RGBA {
	switch s.genre {
	case "scifi":
		return color.RGBA{R: 150, G: 170, B: 200, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 100, G: 120, B: 140, A: 255}
	case "horror":
		return color.RGBA{R: 120, G: 100, B: 100, A: 255}
	default:
		return color.RGBA{R: 180, G: 180, B: 200, A: 255}
	}
}

// getGenreAccentColor returns accent color based on genre.
func (s *IconSystem) getGenreAccentColor() color.RGBA {
	switch s.genre {
	case "scifi":
		return color.RGBA{R: 100, G: 200, B: 255, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 255, G: 0, B: 128, A: 255}
	case "horror":
		return color.RGBA{R: 180, G: 50, B: 50, A: 255}
	default:
		return color.RGBA{R: 200, G: 150, B: 80, A: 255}
	}
}

// getGenreEnchantColor returns enchantment sparkle color based on genre.
func (s *IconSystem) getGenreEnchantColor() color.RGBA {
	switch s.genre {
	case "scifi":
		return color.RGBA{R: 100, G: 255, B: 255, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 255, G: 0, B: 255, A: 255}
	case "horror":
		return color.RGBA{R: 100, G: 255, B: 100, A: 255}
	default:
		return color.RGBA{R: 200, G: 200, B: 255, A: 255}
	}
}

// fillRect fills a rectangle with the specified color.
func fillRect(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}

// drawRect draws a rectangle outline.
func drawRect(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	for x := x1; x < x2; x++ {
		if x >= 0 && x < img.Bounds().Dx() {
			if y1 >= 0 && y1 < img.Bounds().Dy() {
				img.Set(x, y1, c)
			}
			if y2-1 >= 0 && y2-1 < img.Bounds().Dy() {
				img.Set(x, y2-1, c)
			}
		}
	}
	for y := y1; y < y2; y++ {
		if y >= 0 && y < img.Bounds().Dy() {
			if x1 >= 0 && x1 < img.Bounds().Dx() {
				img.Set(x1, y, c)
			}
			if x2-1 >= 0 && x2-1 < img.Bounds().Dx() {
				img.Set(x2-1, y, c)
			}
		}
	}
}

// fillCircle fills a circle with the specified color.
func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				px, py := cx+x, cy+y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// drawLine draws a line between two points.
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.RGBA) {
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	sx, sy := 1, 1
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	for {
		if x1 >= 0 && x1 < img.Bounds().Dx() && y1 >= 0 && y1 < img.Bounds().Dy() {
			img.Set(x1, y1, c)
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

package corpse

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

// Generator creates procedural corpse visuals with caching.
type Generator struct {
	cache      map[cacheKey]*list.Element
	lruList    *list.List
	maxEntries int
	mu         sync.RWMutex
	genreID    string
}

type cacheKey struct {
	seed       int64
	entityType string
	deathType  DeathType
	frame      int
	size       int
	genre      string
}

type cachedCorpse struct {
	image     *ebiten.Image
	key       cacheKey
	accessCnt int
}

// NewGenerator creates a corpse visual generator with LRU cache.
func NewGenerator(maxCacheEntries int) *Generator {
	return &Generator{
		cache:      make(map[cacheKey]*list.Element),
		lruList:    list.New(),
		maxEntries: maxCacheEntries,
		genreID:    "fantasy",
	}
}

// SetGenre configures genre-specific corpse generation.
func (g *Generator) SetGenre(genreID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.genreID = genreID
	g.cache = make(map[cacheKey]*list.Element)
	g.lruList = list.New()
}

// GetCorpseImage retrieves or generates a corpse visual.
func (g *Generator) GetCorpseImage(seed int64, entityType string, deathType DeathType, frame, size int) *ebiten.Image {
	key := cacheKey{
		seed:       seed,
		entityType: entityType,
		deathType:  deathType,
		frame:      frame,
		size:       size,
		genre:      g.genreID,
	}

	g.mu.Lock()
	if elem, found := g.cache[key]; found {
		g.lruList.MoveToFront(elem)
		cached := elem.Value.(*cachedCorpse)
		cached.accessCnt++
		g.mu.Unlock()
		return cached.image
	}
	g.mu.Unlock()

	img := g.generateCorpse(seed, entityType, deathType, frame, size)

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.lruList.Len() >= g.maxEntries {
		oldest := g.lruList.Back()
		if oldest != nil {
			g.lruList.Remove(oldest)
			delete(g.cache, oldest.Value.(*cachedCorpse).key)
		}
	}

	cached := &cachedCorpse{
		image:     img,
		key:       key,
		accessCnt: 1,
	}
	elem := g.lruList.PushFront(cached)
	g.cache[key] = elem

	return img
}

// generateCorpse creates a corpse visual based on death type and entity.
func (g *Generator) generateCorpse(seed int64, entityType string, deathType DeathType, frame, size int) *ebiten.Image {
	rgba := pool.GlobalPools.Images.Get(size, size)
	rng := rand.New(rand.NewSource(seed + int64(frame)))

	switch deathType {
	case DeathBurn:
		g.generateBurnedCorpse(rgba, entityType, rng, frame)
	case DeathFreeze:
		g.generateFrozenCorpse(rgba, entityType, rng, frame)
	case DeathElectric:
		g.generateElectricCorpse(rgba, entityType, rng, frame)
	case DeathAcid:
		g.generateAcidCorpse(rgba, entityType, rng, frame)
	case DeathExplosion:
		g.generateExplodedCorpse(rgba, entityType, rng, frame)
	case DeathSlash:
		g.generateSlashedCorpse(rgba, entityType, rng, frame)
	case DeathCrush:
		g.generateCrushedCorpse(rgba, entityType, rng, frame)
	case DeathDisintegrate:
		g.generateDisintegratedCorpse(rgba, entityType, rng, frame)
	default:
		g.generateNormalCorpse(rgba, entityType, rng, frame)
	}

	result := ebiten.NewImageFromImage(rgba)
	pool.GlobalPools.Images.Put(rgba)
	return result
}

// generateNormalCorpse creates a standard fallen corpse.
func (g *Generator) generateNormalCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	baseColor := g.getCorpseColor(entityType)
	bloodColor := g.getBloodColor()

	heightRatio := 0.3
	width := size * 3 / 4
	height := int(float64(size) * heightRatio)

	for y := cy - height/2; y < cy+height/2; y++ {
		for x := cx - width/2; x < cx+width/2; x++ {
			if x < 0 || x >= size || y < 0 || y >= size {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := float64(width / 2)

			if dist < maxDist {
				shade := 1.0 - (dist / maxDist * 0.6)
				noise := rng.Float64()*0.15 - 0.075
				shade += noise

				r := uint8(math.Min(255, math.Max(0, float64(baseColor.R)*shade)))
				g := uint8(math.Min(255, math.Max(0, float64(baseColor.G)*shade)))
				b := uint8(math.Min(255, math.Max(0, float64(baseColor.B)*shade)))

				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	poolCount := 3 + rng.Intn(3)
	for i := 0; i < poolCount; i++ {
		px := cx + rng.Intn(width/2) - width/4
		py := cy + rng.Intn(height/2) - height/4
		pradius := 3 + rng.Intn(4)
		g.drawBloodPool(img, px, py, pradius, bloodColor, rng)
	}
}

// generateBurnedCorpse creates a charred corpse with embers.
func (g *Generator) generateBurnedCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	charColor := color.RGBA{R: 30, G: 25, B: 20, A: 255}
	emberColor := color.RGBA{R: 255, G: 100, B: 20, A: 255}

	heightRatio := 0.25
	width := size * 3 / 4
	height := int(float64(size) * heightRatio)

	for y := cy - height/2; y < cy+height/2; y++ {
		for x := cx - width/2; x < cx+width/2; x++ {
			if x < 0 || x >= size || y < 0 || y >= size {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := float64(width / 2)

			if dist < maxDist {
				shade := 1.0 - (dist / maxDist * 0.4)
				noise := rng.Float64()*0.2 - 0.1
				shade += noise

				r := uint8(math.Min(255, math.Max(0, float64(charColor.R)*shade)))
				g := uint8(math.Min(255, math.Max(0, float64(charColor.G)*shade)))
				b := uint8(math.Min(255, math.Max(0, float64(charColor.B)*shade)))

				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	emberCount := 5 + rng.Intn(8)
	for i := 0; i < emberCount; i++ {
		ex := cx + rng.Intn(width) - width/2
		ey := cy + rng.Intn(height) - height/2
		brightness := 0.5 + rng.Float64()*0.5
		r := uint8(float64(emberColor.R) * brightness)
		g := uint8(float64(emberColor.G) * brightness)
		b := uint8(float64(emberColor.B) * brightness)

		if ex >= 0 && ex < size && ey >= 0 && ey < size {
			img.Set(ex, ey, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
}

// generateFrozenCorpse creates an icy corpse with frost.
func (g *Generator) generateFrozenCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	iceColor := color.RGBA{R: 180, G: 220, B: 255, A: 255}
	frostColor := color.RGBA{R: 240, G: 250, B: 255, A: 255}

	heightRatio := 0.3
	width := size * 3 / 4
	height := int(float64(size) * heightRatio)

	for y := cy - height/2; y < cy+height/2; y++ {
		for x := cx - width/2; x < cx+width/2; x++ {
			if x < 0 || x >= size || y < 0 || y >= size {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := float64(width / 2)

			if dist < maxDist {
				shade := 0.7 + 0.3*(1.0-dist/maxDist)
				noise := rng.Float64()*0.1 - 0.05
				shade += noise

				r := uint8(math.Min(255, float64(iceColor.R)*shade))
				g := uint8(math.Min(255, float64(iceColor.G)*shade))
				b := uint8(math.Min(255, float64(iceColor.B)*shade))

				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	frostCount := 10 + rng.Intn(15)
	for i := 0; i < frostCount; i++ {
		fx := cx + rng.Intn(width) - width/2
		fy := cy + rng.Intn(height) - height/2
		if fx >= 0 && fx < size && fy >= 0 && fy < size {
			img.Set(fx, fy, frostColor)
		}
	}
}

// generateElectricCorpse creates a scorched corpse with electric marks.
func (g *Generator) generateElectricCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	scorchedColor := color.RGBA{R: 40, G: 35, B: 45, A: 255}
	boltColor := color.RGBA{R: 150, G: 200, B: 255, A: 255}

	heightRatio := 0.3
	width := size * 3 / 4
	height := int(float64(size) * heightRatio)

	for y := cy - height/2; y < cy+height/2; y++ {
		for x := cx - width/2; x < cx+width/2; x++ {
			if x < 0 || x >= size || y < 0 || y >= size {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := float64(width / 2)

			if dist < maxDist {
				shade := 1.0 - (dist / maxDist * 0.5)
				noise := rng.Float64()*0.15 - 0.075
				shade += noise

				r := uint8(math.Min(255, float64(scorchedColor.R)*shade))
				g := uint8(math.Min(255, float64(scorchedColor.G)*shade))
				b := uint8(math.Min(255, float64(scorchedColor.B)*shade))

				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	boltCount := 2 + rng.Intn(3)
	for i := 0; i < boltCount; i++ {
		startX := cx + rng.Intn(width/2) - width/4
		startY := cy + rng.Intn(height/2) - height/4
		g.drawLightningMark(img, startX, startY, boltColor, rng, size)
	}
}

// generateAcidCorpse creates a melted, dissolving corpse.
func (g *Generator) generateAcidCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	acidColor := color.RGBA{R: 100, G: 140, B: 50, A: 255}

	heightRatio := 0.2
	width := size * 3 / 4
	height := int(float64(size) * heightRatio)

	blobCount := 5 + rng.Intn(8)
	for i := 0; i < blobCount; i++ {
		bx := cx + rng.Intn(width) - width/2
		by := cy + rng.Intn(height) - height/4
		bradius := 5 + rng.Intn(8)

		for y := by - bradius; y < by+bradius; y++ {
			for x := bx - bradius; x < bx+bradius; x++ {
				if x < 0 || x >= size || y < 0 || y >= size {
					continue
				}

				dx := float64(x - bx)
				dy := float64(y - by)
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < float64(bradius) {
					shade := 1.0 - (dist / float64(bradius) * 0.6)
					r := uint8(math.Min(255, float64(acidColor.R)*shade))
					g := uint8(math.Min(255, float64(acidColor.G)*shade))
					b := uint8(math.Min(255, float64(acidColor.B)*shade))
					img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}
}

// generateExplodedCorpse creates scattered body parts.
func (g *Generator) generateExplodedCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	goreColor := g.getBloodColor()
	bodyColor := g.getCorpseColor(entityType)

	chunkCount := 6 + rng.Intn(6)
	for i := 0; i < chunkCount; i++ {
		angle := rng.Float64() * 2 * math.Pi
		distance := float64(5 + rng.Intn(size/3))
		px := cx + int(math.Cos(angle)*distance)
		py := cy + int(math.Sin(angle)*distance)
		chunkSize := 3 + rng.Intn(6)

		for y := py - chunkSize; y < py+chunkSize; y++ {
			for x := px - chunkSize; x < px+chunkSize; x++ {
				if x < 0 || x >= size || y < 0 || y >= size {
					continue
				}

				dx := float64(x - px)
				dy := float64(y - py)
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < float64(chunkSize) {
					shade := 1.0 - (dist / float64(chunkSize) * 0.5)
					useGore := rng.Float64() < 0.7
					baseCol := bodyColor
					if useGore {
						baseCol = goreColor
					}

					r := uint8(math.Min(255, float64(baseCol.R)*shade))
					g := uint8(math.Min(255, float64(baseCol.G)*shade))
					b := uint8(math.Min(255, float64(baseCol.B)*shade))
					img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
				}
			}
		}
	}
}

// generateSlashedCorpse creates a corpse with deep cuts.
func (g *Generator) generateSlashedCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	g.generateNormalCorpse(img, entityType, rng, frame)

	size := img.Bounds().Dx()
	cx, cy := size/2, size/2
	bloodColor := g.getBloodColor()

	slashCount := 2 + rng.Intn(3)
	for i := 0; i < slashCount; i++ {
		angle := rng.Float64() * math.Pi
		length := 10 + rng.Intn(size/3)
		width := 2 + rng.Intn(2)

		for t := 0; t < length; t++ {
			sx := cx + int(math.Cos(angle)*float64(t)) - length/2
			sy := cy + int(math.Sin(angle)*float64(t))

			for w := -width; w <= width; w++ {
				wx := sx + int(math.Cos(angle+math.Pi/2)*float64(w))
				wy := sy + int(math.Sin(angle+math.Pi/2)*float64(w))

				if wx >= 0 && wx < size && wy >= 0 && wy < size {
					img.Set(wx, wy, bloodColor)
				}
			}
		}
	}
}

// generateCrushedCorpse creates a flattened corpse.
func (g *Generator) generateCrushedCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	baseColor := g.getCorpseColor(entityType)
	bloodColor := g.getBloodColor()

	heightRatio := 0.15
	width := size * 4 / 5
	height := int(float64(size) * heightRatio)

	for y := cy - height/2; y < cy+height/2; y++ {
		for x := cx - width/2; x < cx+width/2; x++ {
			if x < 0 || x >= size || y < 0 || y >= size {
				continue
			}

			noise := rng.Float64()*0.3 - 0.15
			shade := 0.7 + noise

			r := uint8(math.Min(255, math.Max(0, float64(baseColor.R)*shade)))
			g := uint8(math.Min(255, math.Max(0, float64(baseColor.G)*shade)))
			b := uint8(math.Min(255, math.Max(0, float64(baseColor.B)*shade)))

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	splatRadius := width / 2
	for y := cy - splatRadius; y < cy+splatRadius; y++ {
		for x := cx - splatRadius; x < cx+splatRadius; x++ {
			if x < 0 || x >= size || y < 0 || y >= size {
				continue
			}

			dx := float64(x - cx)
			dy := float64(y - cy)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < float64(splatRadius) && rng.Float64() < 0.3 {
				img.Set(x, y, bloodColor)
			}
		}
	}
}

// generateDisintegratedCorpse creates ash/dust particles.
func (g *Generator) generateDisintegratedCorpse(img *image.RGBA, entityType string, rng *rand.Rand, frame int) {
	size := img.Bounds().Dx()
	cx, cy := size/2, size/2

	ashColor := color.RGBA{R: 100, G: 100, B: 110, A: 255}
	dustColor := color.RGBA{R: 140, G: 140, B: 150, A: 255}

	particleCount := 50 + rng.Intn(100)
	for i := 0; i < particleCount; i++ {
		angle := rng.Float64() * 2 * math.Pi
		distance := rng.Float64() * float64(size/3)
		px := cx + int(math.Cos(angle)*distance)
		py := cy + int(math.Sin(angle)*distance)
		psize := 1 + rng.Intn(3)

		useAsh := rng.Float64() < 0.6
		col := dustColor
		if useAsh {
			col = ashColor
		}

		for dy := 0; dy < psize; dy++ {
			for dx := 0; dx < psize; dx++ {
				x := px + dx
				y := py + dy
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, col)
				}
			}
		}
	}
}

func (g *Generator) drawBloodPool(img *image.RGBA, x, y, radius int, bloodColor color.RGBA, rng *rand.Rand) {
	size := img.Bounds().Dx()
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			px := x + dx
			py := y + dy

			if px < 0 || px >= size || py < 0 || py >= size {
				continue
			}

			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist < float64(radius) {
				opacity := 1.0 - (dist / float64(radius))
				if rng.Float64() < opacity*0.8 {
					img.Set(px, py, bloodColor)
				}
			}
		}
	}
}

func (g *Generator) drawLightningMark(img *image.RGBA, startX, startY int, boltColor color.RGBA, rng *rand.Rand, size int) {
	x, y := startX, startY
	steps := 5 + rng.Intn(8)

	for i := 0; i < steps; i++ {
		dx := rng.Intn(5) - 2
		dy := rng.Intn(5) - 2
		x += dx
		y += dy

		if x >= 0 && x < size && y >= 0 && y < size {
			img.Set(x, y, boltColor)
			if x+1 < size {
				img.Set(x+1, y, boltColor)
			}
		}
	}
}

func (g *Generator) getCorpseColor(entityType string) color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 90, G: 100, B: 110, A: 255}
	case "horror":
		return color.RGBA{R: 100, G: 90, B: 85, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 95, G: 95, B: 105, A: 255}
	case "postapoc":
		return color.RGBA{R: 110, G: 100, B: 90, A: 255}
	default:
		return color.RGBA{R: 120, G: 100, B: 90, A: 255}
	}
}

func (g *Generator) getBloodColor() color.RGBA {
	switch g.genreID {
	case "scifi":
		return color.RGBA{R: 100, G: 200, B: 100, A: 255}
	case "horror":
		return color.RGBA{R: 140, G: 0, B: 0, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 0, G: 255, B: 200, A: 255}
	case "postapoc":
		return color.RGBA{R: 130, G: 30, B: 20, A: 255}
	default:
		return color.RGBA{R: 150, G: 20, B: 20, A: 255}
	}
}

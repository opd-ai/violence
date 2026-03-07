package floor

import (
	"container/list"
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// TileKey uniquely identifies a cached floor tile texture.
type TileKey struct {
	Material MaterialType
	Variant  int
	Seed     int64
	Size     int
}

// CachedTile stores a generated floor tile with metadata.
type CachedTile struct {
	Image     *ebiten.Image
	Key       TileKey
	AccessCnt int
}

// TextureGenerator creates procedural floor tile textures with LRU caching.
type TextureGenerator struct {
	cache      map[TileKey]*list.Element
	lruList    *list.List
	maxEntries int
	mu         sync.RWMutex
	genreID    string
	weathering WeatheringConfig
}

// NewTextureGenerator creates a floor tile texture generator with caching.
func NewTextureGenerator(maxCacheEntries int, genreID string) *TextureGenerator {
	gen := &TextureGenerator{
		cache:      make(map[TileKey]*list.Element),
		lruList:    list.New(),
		maxEntries: maxCacheEntries,
		genreID:    genreID,
	}
	gen.weathering = gen.getGenreWeathering(genreID)
	return gen
}

// SetGenre updates the genre and clears cache.
func (g *TextureGenerator) SetGenre(genreID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.genreID = genreID
	g.weathering = g.getGenreWeathering(genreID)
	g.cache = make(map[TileKey]*list.Element)
	g.lruList = list.New()
}

// GetTile retrieves or generates a floor tile texture.
func (g *TextureGenerator) GetTile(material MaterialType, variant int, seed int64, size int) *ebiten.Image {
	key := TileKey{
		Material: material,
		Variant:  variant,
		Seed:     seed,
		Size:     size,
	}

	g.mu.Lock()
	if elem, found := g.cache[key]; found {
		g.lruList.MoveToFront(elem)
		cached := elem.Value.(*CachedTile)
		cached.AccessCnt++
		g.mu.Unlock()
		return cached.Image
	}
	g.mu.Unlock()

	img := g.generateTile(material, variant, seed, size)

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.lruList.Len() >= g.maxEntries {
		oldest := g.lruList.Back()
		if oldest != nil {
			old := oldest.Value.(*CachedTile)
			delete(g.cache, old.Key)
			g.lruList.Remove(oldest)
		}
	}

	cached := &CachedTile{
		Image:     img,
		Key:       key,
		AccessCnt: 1,
	}
	elem := g.lruList.PushFront(cached)
	g.cache[key] = elem

	return img
}

// generateTile creates a procedural floor tile texture.
func (g *TextureGenerator) generateTile(material MaterialType, variant int, seed int64, size int) *ebiten.Image {
	rng := rand.New(rand.NewSource(seed))
	rawImg := image.NewRGBA(image.Rect(0, 0, size, size))

	// Generate base color and pattern based on material
	switch material {
	case MaterialStone:
		g.drawStoneTile(rawImg, size, variant, rng)
	case MaterialMetal:
		g.drawMetalTile(rawImg, size, variant, rng)
	case MaterialWood:
		g.drawWoodTile(rawImg, size, variant, rng)
	case MaterialConcrete:
		g.drawConcreteTile(rawImg, size, variant, rng)
	case MaterialTile:
		g.drawCeramicTile(rawImg, size, variant, rng)
	case MaterialDirt:
		g.drawDirtTile(rawImg, size, variant, rng)
	case MaterialGrass:
		g.drawGrassTile(rawImg, size, variant, rng)
	case MaterialFlesh:
		g.drawFleshTile(rawImg, size, variant, rng)
	case MaterialCrystal:
		g.drawCrystalTile(rawImg, size, variant, rng)
	default:
		g.drawStoneTile(rawImg, size, variant, rng)
	}

	// Apply weathering effects to make surfaces look lived-in
	g.ApplyWeathering(rawImg, material, g.weathering, seed+1000)

	return ebiten.NewImageFromImage(rawImg)
}

// drawStoneTile creates a stone floor texture with natural variation.
func (g *TextureGenerator) drawStoneTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getStoneBaseColor(variant, rng)

	// Fill with base color and add noise
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			noise := rng.Float64()*0.15 - 0.075
			r := clamp(float64(baseColor.R) * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * (1.0 + noise))
			b := clamp(float64(baseColor.B) * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add veins and cracks
	veinCount := 2 + rng.Intn(3)
	for i := 0; i < veinCount; i++ {
		g.drawVein(img, size, baseColor, rng)
	}

	// Add subtle shading to create depth
	g.addRadialShading(img, size, 0.9, 1.1, rng)
}

// drawMetalTile creates a metallic floor texture with reflective properties.
func (g *TextureGenerator) drawMetalTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getMetalBaseColor(variant, rng)

	// Base fill with metallic gradient
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Subtle directional gradient for metallic look
			gradFactor := 0.9 + 0.2*(float64(x+y)/float64(size*2))
			noise := rng.Float64()*0.08 - 0.04

			r := clamp(float64(baseColor.R) * gradFactor * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * gradFactor * (1.0 + noise))
			b := clamp(float64(baseColor.B) * gradFactor * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add panel lines/rivets
	if size >= 16 {
		g.drawPanelLines(img, size, baseColor, rng)
	}

	// Add metallic highlights
	g.addMetallicHighlights(img, size, baseColor, rng)
}

// drawWoodTile creates a wood grain floor texture.
func (g *TextureGenerator) drawWoodTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getWoodBaseColor(variant, rng)

	// Determine grain direction
	vertical := rng.Float64() < 0.5

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Create wood grain pattern
			coord := float64(x)
			if vertical {
				coord = float64(y)
			}

			grain := math.Sin(coord*0.3+rng.Float64()*2.0) * 0.15
			noise := rng.Float64()*0.1 - 0.05

			brightness := 1.0 + grain + noise

			r := clamp(float64(baseColor.R) * brightness)
			gVal := clamp(float64(baseColor.G) * brightness)
			b := clamp(float64(baseColor.B) * brightness)

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add knots
	knotCount := rng.Intn(3)
	for i := 0; i < knotCount; i++ {
		g.drawWoodKnot(img, size, baseColor, rng)
	}
}

// drawConcreteTile creates a concrete floor texture.
func (g *TextureGenerator) drawConcreteTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getConcreteBaseColor(variant, rng)

	// Base fill with concrete texture
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			noise := rng.Float64()*0.2 - 0.1
			r := clamp(float64(baseColor.R) * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * (1.0 + noise))
			b := clamp(float64(baseColor.B) * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add aggregate spots
	aggregateCount := size * size / 20
	for i := 0; i < aggregateCount; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)
		brightness := 0.8 + rng.Float64()*0.4
		r := clamp(float64(baseColor.R) * brightness)
		gVal := clamp(float64(baseColor.G) * brightness)
		b := clamp(float64(baseColor.B) * brightness)
		img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
	}
}

// drawCeramicTile creates a ceramic/tile floor texture.
func (g *TextureGenerator) drawCeramicTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getTileBaseColor(variant, rng)

	// Smooth, glossy base
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			noise := rng.Float64()*0.05 - 0.025
			r := clamp(float64(baseColor.R) * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * (1.0 + noise))
			b := clamp(float64(baseColor.B) * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add grout lines
	g.drawGroutLines(img, size, baseColor, rng)

	// Add subtle shine
	g.addRadialShading(img, size, 0.95, 1.05, rng)
}

// drawDirtTile creates an earthen floor texture.
func (g *TextureGenerator) drawDirtTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getDirtBaseColor(variant, rng)

	// Rough, organic texture
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			noise := rng.Float64()*0.3 - 0.15
			r := clamp(float64(baseColor.R) * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * (1.0 + noise))
			b := clamp(float64(baseColor.B) * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add pebbles
	pebbleCount := size / 4
	for i := 0; i < pebbleCount; i++ {
		g.drawPebble(img, size, baseColor, rng)
	}
}

// drawGrassTile creates a grass floor texture.
func (g *TextureGenerator) drawGrassTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getGrassBaseColor(variant, rng)

	// Base grass color
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			noise := rng.Float64()*0.2 - 0.1
			r := clamp(float64(baseColor.R) * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * (1.0 + noise))
			b := clamp(float64(baseColor.B) * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add grass blades
	bladeCount := size / 2
	for i := 0; i < bladeCount; i++ {
		g.drawGrassBlade(img, size, baseColor, rng)
	}
}

// drawFleshTile creates an organic/fleshy floor texture (horror genre).
func (g *TextureGenerator) drawFleshTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := color.RGBA{R: 140, G: 80, B: 80, A: 255}

	// Organic base with pulsing variation
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			noise := rng.Float64()*0.2 - 0.1
			r := clamp(float64(baseColor.R) * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * (1.0 + noise))
			b := clamp(float64(baseColor.B) * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add vein network
	veinCount := 3 + rng.Intn(3)
	for i := 0; i < veinCount; i++ {
		g.drawBloodVein(img, size, rng)
	}
}

// drawCrystalTile creates a crystalline floor texture.
func (g *TextureGenerator) drawCrystalTile(img *image.RGBA, size, variant int, rng *rand.Rand) {
	baseColor := g.getCrystalBaseColor(variant, rng)

	// Faceted crystalline base
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			facet := math.Floor(float64(x+y) / 4.0)
			brightness := 0.8 + math.Mod(facet, 2.0)*0.4
			noise := rng.Float64()*0.1 - 0.05

			r := clamp(float64(baseColor.R) * brightness * (1.0 + noise))
			gVal := clamp(float64(baseColor.G) * brightness * (1.0 + noise))
			b := clamp(float64(baseColor.B) * brightness * (1.0 + noise))

			img.Set(x, y, color.RGBA{R: uint8(r), G: uint8(gVal), B: uint8(b), A: 255})
		}
	}

	// Add crystal highlights
	highlightCount := size / 8
	for i := 0; i < highlightCount; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)
		bright := uint8(200 + rng.Intn(55))
		img.Set(x, y, color.RGBA{R: bright, G: bright, B: bright, A: 255})
	}
}

// Helper drawing functions

func (g *TextureGenerator) drawVein(img *image.RGBA, size int, baseColor color.RGBA, rng *rand.Rand) {
	startX := rng.Intn(size)
	startY := rng.Intn(size)
	angle := rng.Float64() * math.Pi * 2
	length := float64(size) * (0.5 + rng.Float64()*0.5)

	veinColor := color.RGBA{
		R: uint8(clamp(float64(baseColor.R) * 0.85)),
		G: uint8(clamp(float64(baseColor.G) * 0.85)),
		B: uint8(clamp(float64(baseColor.B) * 0.85)),
		A: 255,
	}

	for t := 0.0; t < 1.0; t += 0.05 {
		x := startX + int(math.Cos(angle)*length*t)
		y := startY + int(math.Sin(angle)*length*t)
		if x >= 0 && x < size && y >= 0 && y < size {
			img.Set(x, y, veinColor)
		}
	}
}

func (g *TextureGenerator) drawPanelLines(img *image.RGBA, size int, baseColor color.RGBA, rng *rand.Rand) {
	lineColor := color.RGBA{
		R: uint8(clamp(float64(baseColor.R) * 0.7)),
		G: uint8(clamp(float64(baseColor.G) * 0.7)),
		B: uint8(clamp(float64(baseColor.B) * 0.7)),
		A: 255,
	}

	// Horizontal line
	y := size / 2
	for x := 0; x < size; x++ {
		img.Set(x, y, lineColor)
	}

	// Vertical line
	x := size / 2
	for y := 0; y < size; y++ {
		img.Set(x, y, lineColor)
	}
}

func (g *TextureGenerator) addMetallicHighlights(img *image.RGBA, size int, baseColor color.RGBA, rng *rand.Rand) {
	highlightCount := 3 + rng.Intn(5)
	for i := 0; i < highlightCount; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)
		brightness := 1.3 + rng.Float64()*0.4
		r := uint8(clamp(float64(baseColor.R) * brightness))
		gVal := uint8(clamp(float64(baseColor.G) * brightness))
		b := uint8(clamp(float64(baseColor.B) * brightness))
		img.Set(x, y, color.RGBA{R: r, G: gVal, B: b, A: 255})
	}
}

func (g *TextureGenerator) drawWoodKnot(img *image.RGBA, size int, baseColor color.RGBA, rng *rand.Rand) {
	centerX := rng.Intn(size)
	centerY := rng.Intn(size)
	radius := 2.0 + rng.Float64()*3.0

	knotColor := color.RGBA{
		R: uint8(clamp(float64(baseColor.R) * 0.6)),
		G: uint8(clamp(float64(baseColor.G) * 0.6)),
		B: uint8(clamp(float64(baseColor.B) * 0.6)),
		A: 255,
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < radius {
				falloff := 1.0 - (dist / radius)
				current := img.At(x, y).(color.RGBA)
				r := uint8(lerp(float64(current.R), float64(knotColor.R), falloff))
				gVal := uint8(lerp(float64(current.G), float64(knotColor.G), falloff))
				b := uint8(lerp(float64(current.B), float64(knotColor.B), falloff))
				img.Set(x, y, color.RGBA{R: r, G: gVal, B: b, A: 255})
			}
		}
	}
}

func (g *TextureGenerator) drawGroutLines(img *image.RGBA, size int, baseColor color.RGBA, rng *rand.Rand) {
	groutColor := color.RGBA{R: 180, G: 180, B: 180, A: 255}

	// Draw border grout
	for x := 0; x < size; x++ {
		img.Set(x, 0, groutColor)
		img.Set(x, size-1, groutColor)
	}
	for y := 0; y < size; y++ {
		img.Set(0, y, groutColor)
		img.Set(size-1, y, groutColor)
	}
}

func (g *TextureGenerator) drawPebble(img *image.RGBA, size int, baseColor color.RGBA, rng *rand.Rand) {
	x := rng.Intn(size)
	y := rng.Intn(size)
	pebbleSize := 1 + rng.Intn(2)

	pebbleColor := color.RGBA{
		R: uint8(clamp(float64(baseColor.R) * (0.7 + rng.Float64()*0.6))),
		G: uint8(clamp(float64(baseColor.G) * (0.7 + rng.Float64()*0.6))),
		B: uint8(clamp(float64(baseColor.B) * (0.7 + rng.Float64()*0.6))),
		A: 255,
	}

	for dy := 0; dy < pebbleSize; dy++ {
		for dx := 0; dx < pebbleSize; dx++ {
			px := x + dx
			py := y + dy
			if px < size && py < size {
				img.Set(px, py, pebbleColor)
			}
		}
	}
}

func (g *TextureGenerator) drawGrassBlade(img *image.RGBA, size int, baseColor color.RGBA, rng *rand.Rand) {
	x := rng.Intn(size)
	y := rng.Intn(size)
	length := 2 + rng.Intn(3)

	bladeColor := color.RGBA{
		R: uint8(clamp(float64(baseColor.R) * (0.6 + rng.Float64()*0.4))),
		G: uint8(clamp(float64(baseColor.G) * (1.0 + rng.Float64()*0.2))),
		B: uint8(clamp(float64(baseColor.B) * (0.6 + rng.Float64()*0.4))),
		A: 255,
	}

	for i := 0; i < length; i++ {
		py := y - i
		if py >= 0 && py < size {
			img.Set(x, py, bladeColor)
		}
	}
}

func (g *TextureGenerator) drawBloodVein(img *image.RGBA, size int, rng *rand.Rand) {
	startX := rng.Intn(size)
	startY := rng.Intn(size)
	angle := rng.Float64() * math.Pi * 2
	length := float64(size) * (0.4 + rng.Float64()*0.4)

	veinColor := color.RGBA{R: 100, G: 40, B: 40, A: 255}

	for t := 0.0; t < 1.0; t += 0.05 {
		x := startX + int(math.Cos(angle)*length*t)
		y := startY + int(math.Sin(angle)*length*t)
		if x >= 0 && x < size && y >= 0 && y < size {
			img.Set(x, y, veinColor)
		}
	}
}

func (g *TextureGenerator) addRadialShading(img *image.RGBA, size int, minBrightness, maxBrightness float64, rng *rand.Rand) {
	centerX := size / 2
	centerY := size / 2
	maxDist := math.Sqrt(float64(centerX*centerX + centerY*centerY))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)
			distFactor := dist / maxDist

			brightness := lerp(maxBrightness, minBrightness, distFactor)
			current := img.At(x, y).(color.RGBA)
			r := uint8(clamp(float64(current.R) * brightness))
			gVal := uint8(clamp(float64(current.G) * brightness))
			b := uint8(clamp(float64(current.B) * brightness))
			img.Set(x, y, color.RGBA{R: r, G: gVal, B: b, A: 255})
		}
	}
}

// Color palette functions

func (g *TextureGenerator) getStoneBaseColor(variant int, rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{R: 120, G: 120, B: 120, A: 255}, // Gray
		{R: 100, G: 90, B: 80, A: 255},   // Brown-gray
		{R: 80, G: 80, B: 90, A: 255},    // Blue-gray
		{R: 110, G: 100, B: 90, A: 255},  // Tan
	}
	return colors[variant%len(colors)]
}

func (g *TextureGenerator) getMetalBaseColor(variant int, rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{R: 160, G: 160, B: 170, A: 255}, // Steel
		{R: 140, G: 120, B: 100, A: 255}, // Bronze
		{R: 180, G: 180, B: 180, A: 255}, // Chrome
		{R: 120, G: 140, B: 150, A: 255}, // Aluminum
	}
	return colors[variant%len(colors)]
}

func (g *TextureGenerator) getWoodBaseColor(variant int, rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{R: 139, G: 90, B: 43, A: 255},   // Dark wood
		{R: 205, G: 133, B: 63, A: 255},  // Light wood
		{R: 160, G: 82, B: 45, A: 255},   // Mahogany
		{R: 210, G: 180, B: 140, A: 255}, // Tan wood
	}
	return colors[variant%len(colors)]
}

func (g *TextureGenerator) getConcreteBaseColor(variant int, rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{R: 150, G: 150, B: 150, A: 255},
		{R: 140, G: 140, B: 145, A: 255},
		{R: 160, G: 160, B: 155, A: 255},
	}
	return colors[variant%len(colors)]
}

func (g *TextureGenerator) getTileBaseColor(variant int, rng *rand.Rand) color.RGBA {
	genreColors := map[string][]color.RGBA{
		"fantasy":   {{R: 220, G: 210, B: 200, A: 255}, {R: 200, G: 180, B: 160, A: 255}},
		"scifi":     {{R: 200, G: 200, B: 210, A: 255}, {R: 180, G: 190, B: 200, A: 255}},
		"horror":    {{R: 160, G: 150, B: 150, A: 255}, {R: 140, G: 140, B: 145, A: 255}},
		"cyberpunk": {{R: 40, G: 40, B: 50, A: 255}, {R: 60, G: 60, B: 70, A: 255}},
		"postapoc":  {{R: 130, G: 130, B: 120, A: 255}, {R: 140, G: 135, B: 125, A: 255}},
	}

	if palette, ok := genreColors[g.genreID]; ok {
		return palette[variant%len(palette)]
	}
	return color.RGBA{R: 200, G: 200, B: 200, A: 255}
}

func (g *TextureGenerator) getDirtBaseColor(variant int, rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{R: 101, G: 67, B: 33, A: 255},  // Dark dirt
		{R: 139, G: 90, B: 43, A: 255},  // Medium dirt
		{R: 160, G: 110, B: 70, A: 255}, // Light dirt
	}
	return colors[variant%len(colors)]
}

func (g *TextureGenerator) getGrassBaseColor(variant int, rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{R: 60, G: 120, B: 50, A: 255},  // Dark grass
		{R: 80, G: 140, B: 70, A: 255},  // Medium grass
		{R: 100, G: 160, B: 90, A: 255}, // Light grass
	}
	return colors[variant%len(colors)]
}

func (g *TextureGenerator) getCrystalBaseColor(variant int, rng *rand.Rand) color.RGBA {
	colors := []color.RGBA{
		{R: 180, G: 200, B: 255, A: 255}, // Blue crystal
		{R: 255, G: 180, B: 200, A: 255}, // Pink crystal
		{R: 200, G: 255, B: 200, A: 255}, // Green crystal
		{R: 220, G: 220, B: 255, A: 255}, // Purple crystal
	}
	return colors[variant%len(colors)]
}

// Utility functions

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// getGenreWeathering returns weathering configuration appropriate for a genre.
func (g *TextureGenerator) getGenreWeathering(genreID string) WeatheringConfig {
	switch genreID {
	case "fantasy":
		// Ancient dungeons: heavy wear, moss, dampness
		return WeatheringConfig{
			EdgeDamage:       0.5,
			WearIntensity:    0.6,
			AgeVariation:     0.7,
			MoistureLevel:    0.4,
			OrganicGrowth:    0.5,
			ColorTemperature: -0.1, // Slightly cool (torchlight in distance)
		}
	case "scifi", "cyberpunk":
		// Worn industrial spaces: rust, corrosion, minimal organic growth
		return WeatheringConfig{
			EdgeDamage:       0.3,
			WearIntensity:    0.7,
			AgeVariation:     0.5,
			MoistureLevel:    0.3,
			OrganicGrowth:    0.1,
			ColorTemperature: 0.15, // Warm (artificial lighting)
		}
	case "horror":
		// Decaying environment: heavy organic growth, moisture, age
		return WeatheringConfig{
			EdgeDamage:       0.6,
			WearIntensity:    0.5,
			AgeVariation:     0.8,
			MoistureLevel:    0.6,
			OrganicGrowth:    0.7,
			ColorTemperature: -0.2, // Cool and sickly
		}
	case "postapoc":
		// Abandoned, weathered: everything shows age
		return WeatheringConfig{
			EdgeDamage:       0.7,
			WearIntensity:    0.6,
			AgeVariation:     0.9,
			MoistureLevel:    0.5,
			OrganicGrowth:    0.6,
			ColorTemperature: 0.1, // Dusty warm tones
		}
	default:
		return DefaultWeatheringConfig()
	}
}

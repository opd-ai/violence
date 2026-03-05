// Package decal implements procedural decal image generation.
package decal

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

// DecalKey uniquely identifies a cached decal.
type DecalKey struct {
	Type    DecalType
	Subtype int
	Seed    int64
	Size    int
}

// CachedDecal stores a generated decal with metadata.
type CachedDecal struct {
	Image     *ebiten.Image
	Key       DecalKey
	AccessCnt int
}

// Generator creates procedural decal images with caching.
type Generator struct {
	cache      map[DecalKey]*list.Element
	lruList    *list.List
	maxEntries int
	mu         sync.RWMutex
	genreID    string
	imagePool  *pool.ImagePool
}

// NewGenerator creates a decal generator with LRU cache.
func NewGenerator(maxCacheEntries int) *Generator {
	return &Generator{
		cache:      make(map[DecalKey]*list.Element),
		lruList:    list.New(),
		maxEntries: maxCacheEntries,
		genreID:    "fantasy",
		imagePool:  pool.NewImagePool(),
	}
}

// SetGenre configures genre-specific decal generation.
func (g *Generator) SetGenre(genreID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.genreID = genreID
	// Clear cache on genre change
	g.cache = make(map[DecalKey]*list.Element)
	g.lruList = list.New()
}

// GetDecal retrieves or generates a decal image.
func (g *Generator) GetDecal(decalType DecalType, subtype int, seed int64, size int) *ebiten.Image {
	key := DecalKey{
		Type:    decalType,
		Subtype: subtype,
		Seed:    seed,
		Size:    size,
	}

	g.mu.Lock()
	if elem, found := g.cache[key]; found {
		g.lruList.MoveToFront(elem)
		cached := elem.Value.(*CachedDecal)
		cached.AccessCnt++
		g.mu.Unlock()
		return cached.Image
	}
	g.mu.Unlock()

	img := g.generateDecal(decalType, subtype, seed, size)

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.lruList.Len() >= g.maxEntries {
		oldest := g.lruList.Back()
		if oldest != nil {
			g.lruList.Remove(oldest)
			oldCached := oldest.Value.(*CachedDecal)
			delete(g.cache, oldCached.Key)
		}
	}

	cached := &CachedDecal{
		Image:     img,
		Key:       key,
		AccessCnt: 1,
	}
	elem := g.lruList.PushFront(cached)
	g.cache[key] = elem

	return img
}

// generateDecal creates a decal image procedurally.
func (g *Generator) generateDecal(decalType DecalType, subtype int, seed int64, size int) *ebiten.Image {
	rng := rand.New(rand.NewSource(seed))

	// Get buffer from pool
	buf := g.imagePool.Get(size, size)
	defer g.imagePool.Put(buf)

	// Fill transparent
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			buf.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}

	switch decalType {
	case DecalBlood:
		g.drawBloodSplatter(buf, size, subtype, rng)
	case DecalScorch:
		g.drawScorchMark(buf, size, subtype, rng)
	case DecalSlash:
		g.drawSlashMark(buf, size, subtype, rng)
	case DecalBulletHole:
		g.drawBulletHole(buf, size, subtype, rng)
	case DecalMagicBurn:
		g.drawMagicBurn(buf, size, subtype, rng)
	case DecalAcid:
		g.drawAcidBurn(buf, size, subtype, rng)
	case DecalFreeze:
		g.drawFreezeMark(buf, size, subtype, rng)
	case DecalElectric:
		g.drawElectricScorch(buf, size, subtype, rng)
	}

	return ebiten.NewImageFromImage(buf)
}

// drawBloodSplatter creates blood splatter patterns.
func (g *Generator) drawBloodSplatter(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	baseColor := g.getBloodColor()
	centerX, centerY := size/2, size/2

	drawMainBloodBlob(buf, size, centerX, centerY, baseColor, rng)
	drawBloodStreaks(buf, size, centerX, centerY, subtype, baseColor, rng)
}

// drawMainBloodBlob renders the central blood splatter blob.
func drawMainBloodBlob(buf *image.RGBA, size, centerX, centerY int, baseColor color.RGBA, rng *rand.Rand) {
	blobRadius := float64(size) * (0.3 + rng.Float64()*0.2)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if col, ok := calculateBloodBlobPixel(x, y, centerX, centerY, blobRadius, baseColor, rng); ok {
				buf.Set(x, y, col)
			}
		}
	}
}

// calculateBloodBlobPixel computes the color for a blood blob pixel if within radius.
func calculateBloodBlobPixel(x, y, centerX, centerY int, blobRadius float64, baseColor color.RGBA, rng *rand.Rand) (color.RGBA, bool) {
	dx := float64(x - centerX)
	dy := float64(y - centerY)
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist >= blobRadius {
		return color.RGBA{}, false
	}

	edgeNoise := rng.Float64() * 0.3
	opacity := math.Max(0, math.Min(1, (blobRadius-dist)/blobRadius+edgeNoise))
	brightness := 0.6 + opacity*0.4

	return color.RGBA{
		R: uint8(float64(baseColor.R) * brightness),
		G: uint8(float64(baseColor.G) * brightness),
		B: uint8(float64(baseColor.B) * brightness),
		A: uint8(float64(baseColor.A) * opacity),
	}, true
}

// drawBloodStreaks renders radiating blood streaks from the center.
func drawBloodStreaks(buf *image.RGBA, size, centerX, centerY, subtype int, baseColor color.RGBA, rng *rand.Rand) {
	numStreaks := 3 + subtype*2

	for i := 0; i < numStreaks; i++ {
		angle := rng.Float64() * 2 * math.Pi
		length := 5.0 + rng.Float64()*float64(size)*0.3
		thickness := 1.0 + rng.Float64()*2.0

		drawSingleStreak(buf, size, centerX, centerY, angle, length, thickness, baseColor)
	}
}

// drawSingleStreak draws a single blood streak with gradual opacity falloff.
func drawSingleStreak(buf *image.RGBA, size, centerX, centerY int, angle, length, thickness float64, baseColor color.RGBA) {
	for t := 0.0; t < length; t++ {
		px := centerX + int(math.Cos(angle)*t)
		py := centerY + int(math.Sin(angle)*t)

		if px < 0 || px >= size || py < 0 || py >= size {
			continue
		}

		opacity := math.Max(0, 1.0-t/length)
		c := color.RGBA{
			R: uint8(float64(baseColor.R) * 0.7),
			G: uint8(float64(baseColor.G) * 0.7),
			B: uint8(float64(baseColor.B) * 0.7),
			A: uint8(float64(baseColor.A) * opacity),
		}

		applyStreakThickness(buf, size, px, py, int(thickness), c)
	}
}

// applyStreakThickness applies thickness to a streak pixel.
func applyStreakThickness(buf *image.RGBA, size, px, py, thickness int, c color.RGBA) {
	for dy := -thickness; dy <= thickness; dy++ {
		for dx := -thickness; dx <= thickness; dx++ {
			nx, ny := px+dx, py+dy
			if nx >= 0 && nx < size && ny >= 0 && ny < size {
				buf.Set(nx, ny, c)
			}
		}
	}
}

// drawScorchMark creates burn/scorch patterns.
func (g *Generator) drawScorchMark(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	centerX, centerY := size/2, size/2
	radius := float64(size) * (0.35 + rng.Float64()*0.15)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < radius {
				// Gradient from black center to orange edge
				t := dist / radius
				opacity := 1.0 - t

				var c color.RGBA
				if t < 0.4 {
					// Black center
					c = color.RGBA{R: 10, G: 10, B: 10, A: uint8(200 * opacity)}
				} else if t < 0.7 {
					// Dark brown/red
					c = color.RGBA{R: 60, G: 30, B: 10, A: uint8(180 * opacity)}
				} else {
					// Orange edge
					c = color.RGBA{R: 120, G: 60, B: 20, A: uint8(150 * opacity)}
				}

				buf.Set(x, y, c)
			}
		}
	}
}

// drawSlashMark creates slash/cut patterns.
func (g *Generator) drawSlashMark(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	baseColor := g.getBloodColor()
	centerX, centerY := size/2, size/2

	// Slash is elongated along one axis
	angle := float64(subtype) * math.Pi / 4
	length := float64(size) * 0.7
	thickness := 2.0 + rng.Float64()*3.0

	for t := -length / 2; t < length/2; t++ {
		px := centerX + int(math.Cos(angle)*t)
		py := centerY + int(math.Sin(angle)*t)

		if px >= 0 && px < size && py >= 0 && py < size {
			// Perpendicular direction for thickness
			perpAngle := angle + math.Pi/2

			for w := -thickness; w < thickness; w++ {
				nx := px + int(math.Cos(perpAngle)*w)
				ny := py + int(math.Sin(perpAngle)*w)

				if nx >= 0 && nx < size && ny >= 0 && ny < size {
					opacity := 1.0 - math.Abs(w)/thickness
					c := color.RGBA{
						R: uint8(float64(baseColor.R) * 0.8),
						G: uint8(float64(baseColor.G) * 0.8),
						B: uint8(float64(baseColor.B) * 0.8),
						A: uint8(float64(baseColor.A) * opacity * 0.9),
					}
					buf.Set(nx, ny, c)
				}
			}
		}
	}
}

// drawBulletHole creates bullet impact patterns.
func (g *Generator) drawBulletHole(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	centerX, centerY := size/2, size/2
	holeRadius := 2.0 + rng.Float64()*2.0
	crackRadius := float64(size) * 0.4

	// Dark hole center
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < holeRadius {
				c := color.RGBA{R: 20, G: 20, B: 20, A: 200}
				buf.Set(x, y, c)
			}
		}
	}

	// Radial cracks
	numCracks := 4 + subtype
	for i := 0; i < numCracks; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numCracks)
		angle += (rng.Float64() - 0.5) * 0.5

		length := holeRadius + rng.Float64()*crackRadius
		for t := holeRadius; t < length; t++ {
			px := centerX + int(math.Cos(angle)*t)
			py := centerY + int(math.Sin(angle)*t)

			if px >= 0 && px < size && py >= 0 && py < size {
				opacity := 1.0 - (t-holeRadius)/(length-holeRadius)
				c := color.RGBA{R: 40, G: 40, B: 40, A: uint8(120 * opacity)}
				buf.Set(px, py, c)
			}
		}
	}
}

// drawMagicBurn creates magical burn patterns.
func (g *Generator) drawMagicBurn(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	centerX, centerY := size/2, size/2
	radius := float64(size) * 0.4

	// Purple/blue magical burn
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < radius {
				t := dist / radius
				opacity := 1.0 - t

				c := color.RGBA{
					R: uint8(100 * (1 - t)),
					G: uint8(50 * (1 - t)),
					B: uint8(200 * opacity),
					A: uint8(180 * opacity),
				}
				buf.Set(x, y, c)
			}
		}
	}
}

// drawAcidBurn creates acid/corrosion patterns.
func (g *Generator) drawAcidBurn(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	centerX, centerY := size/2, size/2

	// Irregular acid splash
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			// Noise for irregular shape
			noise := rng.Float64()
			radius := float64(size) * (0.3 + noise*0.2)

			if dist < radius {
				opacity := 1.0 - dist/radius
				c := color.RGBA{
					R: uint8(100 * opacity),
					G: uint8(200 * opacity),
					B: uint8(50 * opacity),
					A: uint8(160 * opacity),
				}
				buf.Set(x, y, c)
			}
		}
	}
}

// drawFreezeMark creates ice/frost patterns.
func (g *Generator) drawFreezeMark(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	centerX, centerY := size/2, size/2

	// Ice crystal pattern
	numBranches := 6
	for i := 0; i < numBranches; i++ {
		angle := float64(i) * 2 * math.Pi / float64(numBranches)
		length := float64(size) * (0.3 + rng.Float64()*0.2)

		for t := 0.0; t < length; t++ {
			px := centerX + int(math.Cos(angle)*t)
			py := centerY + int(math.Sin(angle)*t)

			if px >= 0 && px < size && py >= 0 && py < size {
				opacity := 1.0 - t/length
				c := color.RGBA{
					R: uint8(200 * opacity),
					G: uint8(230 * opacity),
					B: 255,
					A: uint8(150 * opacity),
				}
				buf.Set(px, py, c)
			}
		}
	}
}

// drawElectricScorch creates electric burn patterns.
func (g *Generator) drawElectricScorch(buf *image.RGBA, size, subtype int, rng *rand.Rand) {
	centerX, centerY := size/2, size/2

	// Lightning-like branching pattern
	numBolts := 5 + subtype
	for i := 0; i < numBolts; i++ {
		angle := rng.Float64() * 2 * math.Pi
		length := float64(size) * (0.2 + rng.Float64()*0.3)

		px, py := float64(centerX), float64(centerY)
		for t := 0.0; t < length; t++ {
			// Add jitter
			angle += (rng.Float64() - 0.5) * 0.3
			px += math.Cos(angle)
			py += math.Sin(angle)

			ix, iy := int(px), int(py)
			if ix >= 0 && ix < size && iy >= 0 && iy < size {
				opacity := 1.0 - t/length
				c := color.RGBA{
					R: uint8(255 * opacity),
					G: uint8(255 * opacity),
					B: uint8(100 * opacity),
					A: uint8(180 * opacity),
				}
				buf.Set(ix, iy, c)
			}
		}
	}
}

// getBloodColor returns genre-appropriate blood color.
func (g *Generator) getBloodColor() color.RGBA {
	switch g.genreID {
	case "fantasy":
		return color.RGBA{R: 200, G: 0, B: 0, A: 255}
	case "scifi":
		return color.RGBA{R: 100, G: 255, B: 100, A: 255}
	case "horror":
		return color.RGBA{R: 120, G: 0, B: 0, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 255, G: 0, B: 100, A: 255}
	default:
		return color.RGBA{R: 200, G: 0, B: 0, A: 255}
	}
}

package floor

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages floor tile visual variation and detail overlays.
type System struct {
	genre       string
	tileSize    int
	detailCache map[cacheKey]*ebiten.Image
	logger      *logrus.Entry
}

type cacheKey struct {
	x, y       int
	detailType DetailType
	seed       int64
}

// NewSystem creates a floor detail system.
func NewSystem(genreID string, tileSize int) *System {
	return &System{
		genre:       genreID,
		tileSize:    tileSize,
		detailCache: make(map[cacheKey]*ebiten.Image),
		logger: logrus.WithFields(logrus.Fields{
			"system": "floor",
			"genre":  genreID,
		}),
	}
}

// SetGenre updates the genre for detail generation.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
	s.detailCache = make(map[cacheKey]*ebiten.Image)
	s.logger = s.logger.WithField("genre", genreID)
}

// Update processes floor detail components.
func (s *System) Update(w *engine.World) {
	// Floor details are static - no per-frame updates needed
}

// GenerateFloorDetails creates detail overlays for a dungeon level.
func (s *System) GenerateFloorDetails(tiles [][]int, seed int64) []*FloorDetailComponent {
	if len(tiles) == 0 || len(tiles[0]) == 0 {
		return nil
	}

	rng := rand.New(rand.NewSource(seed))
	details := make([]*FloorDetailComponent, 0, 128)

	// Determine detail density based on genre
	density := s.getDetailDensity()

	for y := range tiles {
		for x := range tiles[y] {
			// Only add details to floor tiles
			if !s.isFloorTile(tiles[y][x]) {
				continue
			}

			// Roll for detail placement
			if rng.Float64() > density {
				continue
			}

			detail := s.generateDetail(x, y, tiles, rng)
			if detail != nil {
				details = append(details, detail)
			}
		}
	}

	s.logger.WithFields(logrus.Fields{
		"details_generated": len(details),
		"seed":              seed,
	}).Debug("Floor details generated")

	return details
}

// generateDetail creates a single floor detail.
func (s *System) generateDetail(x, y int, tiles [][]int, rng *rand.Rand) *FloorDetailComponent {
	detailTypes := s.getGenreDetailTypes()
	if len(detailTypes) == 0 {
		return nil
	}

	dtype := detailTypes[rng.Intn(len(detailTypes))]
	intensity := 0.3 + rng.Float64()*0.7

	// Reduce intensity near room centers, increase near walls
	nearWall := s.isNearWall(x, y, tiles)
	if nearWall {
		intensity = math.Min(1.0, intensity*1.3)
	} else {
		intensity *= 0.7
	}

	return &FloorDetailComponent{
		X:          x,
		Y:          y,
		DetailType: dtype,
		Intensity:  intensity,
		Seed:       int64(x*1000 + y),
		GenreID:    s.genre,
	}
}

// RenderDetail generates or retrieves cached sprite for a floor detail.
func (s *System) RenderDetail(detail *FloorDetailComponent) *ebiten.Image {
	key := cacheKey{
		x:          detail.X,
		y:          detail.Y,
		detailType: detail.DetailType,
		seed:       detail.Seed,
	}

	if cached, ok := s.detailCache[key]; ok {
		return cached
	}

	img := s.generateDetailSprite(detail)
	s.detailCache[key] = img

	// Limit cache size
	if len(s.detailCache) > 500 {
		for k := range s.detailCache {
			delete(s.detailCache, k)
			break
		}
	}

	return img
}

// generateDetailSprite creates a procedural sprite for a floor detail.
func (s *System) generateDetailSprite(detail *FloorDetailComponent) *ebiten.Image {
	rng := rand.New(rand.NewSource(detail.Seed))
	size := s.tileSize

	rawImg := image.NewRGBA(image.Rect(0, 0, size, size))

	switch detail.DetailType {
	case DetailCrack:
		s.drawCrack(rawImg, size, detail.Intensity, rng)
	case DetailStain:
		s.drawStain(rawImg, size, detail.Intensity, rng)
	case DetailDebris:
		s.drawDebris(rawImg, size, detail.Intensity, rng)
	case DetailScorch:
		s.drawScorch(rawImg, size, detail.Intensity, rng)
	case DetailWear:
		s.drawWear(rawImg, size, detail.Intensity, rng)
	case DetailGraffiti:
		s.drawGraffiti(rawImg, size, detail.Intensity, rng)
	case DetailBlood:
		s.drawBlood(rawImg, size, detail.Intensity, rng)
	case DetailRust:
		s.drawRust(rawImg, size, detail.Intensity, rng)
	case DetailCorrode:
		s.drawCorrode(rawImg, size, detail.Intensity, rng)
	}

	return ebiten.NewImageFromImage(rawImg)
}

// drawCrack renders floor cracks.
func (s *System) drawCrack(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	centerX := size / 2
	centerY := size / 2

	// Main crack line
	angle := rng.Float64() * math.Pi * 2
	length := float64(size) * (0.5 + rng.Float64()*0.4)

	for t := 0.0; t < 1.0; t += 0.02 {
		x := centerX + int(math.Cos(angle)*length*t)
		y := centerY + int(math.Sin(angle)*length*t)

		if x < 0 || x >= size || y < 0 || y >= size {
			continue
		}

		// Add jitter
		x += rng.Intn(3) - 1
		y += rng.Intn(3) - 1

		if x >= 0 && x < size && y >= 0 && y < size {
			alpha := uint8(intensity * 160 * (1.0 - t*0.5))
			img.Set(x, y, color.RGBA{R: 40, G: 40, B: 40, A: alpha})

			// Side pixels for width
			if x+1 < size && rng.Float64() < 0.5 {
				img.Set(x+1, y, color.RGBA{R: 50, G: 50, B: 50, A: alpha / 2})
			}
		}
	}

	// Branch cracks
	if intensity > 0.6 && rng.Float64() < 0.4 {
		branchAngle := angle + (rng.Float64()-0.5)*math.Pi/2
		branchLen := length * 0.5
		for t := 0.0; t < 1.0; t += 0.03 {
			x := centerX + int(math.Cos(branchAngle)*branchLen*t)
			y := centerY + int(math.Sin(branchAngle)*branchLen*t)
			if x >= 0 && x < size && y >= 0 && y < size {
				alpha := uint8(intensity * 120 * (1.0 - t))
				img.Set(x, y, color.RGBA{R: 45, G: 45, B: 45, A: alpha})
			}
		}
	}
}

// drawStain renders liquid stains.
func (s *System) drawStain(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	centerX := size/2 + rng.Intn(size/4) - size/8
	centerY := size/2 + rng.Intn(size/4) - size/8

	radius := float64(size) * (0.25 + rng.Float64()*0.25)

	stainColor := s.getStainColor(rng)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < radius {
				// Irregular edges
				noise := rng.Float64() * 0.3
				falloff := 1.0 - (dist/radius)*noise

				if falloff > 0 {
					alpha := uint8(falloff * intensity * 180)
					r := stainColor.R
					g := stainColor.G
					b := stainColor.B
					img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: alpha})
				}
			}
		}
	}
}

// drawDebris renders small scattered debris.
func (s *System) drawDebris(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	debrisCount := int(intensity * 12)

	debrisColor := s.getDebrisColor(rng)

	for i := 0; i < debrisCount; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)
		debrisSize := 1 + rng.Intn(3)

		for dy := 0; dy < debrisSize; dy++ {
			for dx := 0; dx < debrisSize; dx++ {
				px := x + dx
				py := y + dy
				if px >= 0 && px < size && py >= 0 && py < size {
					alpha := uint8(intensity * 200)
					img.Set(px, py, color.RGBA{
						R: debrisColor.R,
						G: debrisColor.G,
						B: debrisColor.B,
						A: alpha,
					})
				}
			}
		}
	}
}

// drawScorch renders burn marks.
func (s *System) drawScorch(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	centerX := size / 2
	centerY := size / 2
	radius := float64(size) * (0.3 + rng.Float64()*0.2)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < radius {
				falloff := 1.0 - (dist / radius)
				alpha := uint8(falloff * intensity * 200)

				// Gradient from black to dark brown
				brightness := uint8(falloff * 40)
				img.Set(x, y, color.RGBA{
					R: brightness,
					G: brightness / 2,
					B: 0,
					A: alpha,
				})
			}
		}
	}
}

// drawWear renders wear patterns.
func (s *System) drawWear(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	for i := 0; i < size*size/8; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)

		if rng.Float64() < intensity {
			alpha := uint8(100 + rng.Intn(100))
			brightness := uint8(60 + rng.Intn(40))
			img.Set(x, y, color.RGBA{
				R: brightness,
				G: brightness,
				B: brightness,
				A: alpha,
			})
		}
	}
}

// drawGraffiti renders simple marks/symbols.
func (s *System) drawGraffiti(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	grafColor := s.getGraffitiColor(rng)

	// Simple line pattern
	startX := size / 4
	startY := size / 2
	endX := 3 * size / 4
	endY := size / 2

	steps := 20
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(float64(startX)*(1-t) + float64(endX)*t)
		y := int(float64(startY)*(1-t) + float64(endY)*t)

		// Add noise
		y += rng.Intn(5) - 2

		if x >= 0 && x < size && y >= 0 && y < size {
			alpha := uint8(intensity * 180)
			img.Set(x, y, color.RGBA{
				R: grafColor.R,
				G: grafColor.G,
				B: grafColor.B,
				A: alpha,
			})
		}
	}
}

// drawBlood renders blood splatter.
func (s *System) drawBlood(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	centerX := size/2 + rng.Intn(size/4) - size/8
	centerY := size/2 + rng.Intn(size/4) - size/8

	// Main pool
	radius := float64(size) * (0.2 + rng.Float64()*0.2)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < radius {
				falloff := 1.0 - (dist / radius)
				alpha := uint8(falloff * intensity * 200)
				img.Set(x, y, color.RGBA{R: 120, G: 0, B: 0, A: alpha})
			}
		}
	}

	// Splatter droplets
	droplets := int(intensity * 8)
	for i := 0; i < droplets; i++ {
		angle := rng.Float64() * math.Pi * 2
		dist := radius + rng.Float64()*float64(size)*0.3
		dx := centerX + int(math.Cos(angle)*dist)
		dy := centerY + int(math.Sin(angle)*dist)

		dropSize := 1 + rng.Intn(2)
		for py := 0; py < dropSize; py++ {
			for px := 0; px < dropSize; px++ {
				x := dx + px
				y := dy + py
				if x >= 0 && x < size && y >= 0 && y < size {
					img.Set(x, y, color.RGBA{R: 100, G: 0, B: 0, A: uint8(intensity * 180)})
				}
			}
		}
	}
}

// drawRust renders rust patterns.
func (s *System) drawRust(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	for i := 0; i < size*size/4; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)

		if rng.Float64() < intensity*0.8 {
			// Rust color variations
			r := uint8(140 + rng.Intn(40))
			g := uint8(60 + rng.Intn(30))
			b := uint8(20 + rng.Intn(20))
			alpha := uint8(120 + rng.Intn(100))

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: alpha})
		}
	}
}

// drawCorrode renders corrosion patterns.
func (s *System) drawCorrode(img *image.RGBA, size int, intensity float64, rng *rand.Rand) {
	for i := 0; i < size*size/3; i++ {
		x := rng.Intn(size)
		y := rng.Intn(size)

		if rng.Float64() < intensity*0.9 {
			// Greenish corrosion
			r := uint8(40 + rng.Intn(30))
			g := uint8(80 + rng.Intn(40))
			b := uint8(40 + rng.Intn(30))
			alpha := uint8(100 + rng.Intn(120))

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: alpha})
		}
	}
}

// Helper functions

func (s *System) getDetailDensity() float64 {
	densities := map[string]float64{
		"fantasy":   0.15,
		"scifi":     0.20,
		"horror":    0.25,
		"cyberpunk": 0.22,
		"postapoc":  0.30,
	}

	if d, ok := densities[s.genre]; ok {
		return d
	}
	return 0.15
}

func (s *System) getGenreDetailTypes() []DetailType {
	typeMap := map[string][]DetailType{
		"fantasy": {
			DetailCrack, DetailStain, DetailWear, DetailDebris,
		},
		"scifi": {
			DetailCrack, DetailScorch, DetailWear, DetailDebris, DetailCorrode,
		},
		"horror": {
			DetailCrack, DetailStain, DetailBlood, DetailDebris, DetailWear,
		},
		"cyberpunk": {
			DetailScorch, DetailStain, DetailGraffiti, DetailWear, DetailRust,
		},
		"postapoc": {
			DetailCrack, DetailStain, DetailDebris, DetailScorch, DetailRust, DetailBlood,
		},
	}

	if types, ok := typeMap[s.genre]; ok {
		return types
	}
	return []DetailType{DetailCrack, DetailStain, DetailWear}
}

func (s *System) getStainColor(rng *rand.Rand) color.RGBA {
	colors := map[string][]color.RGBA{
		"fantasy":   {{R: 80, G: 60, B: 40}, {R: 60, G: 80, B: 60}},
		"scifi":     {{R: 100, G: 100, B: 120}, {R: 80, G: 90, B: 100}},
		"horror":    {{R: 60, G: 50, B: 50}, {R: 90, G: 60, B: 60}},
		"cyberpunk": {{R: 120, G: 80, B: 140}, {R: 80, G: 120, B: 100}},
		"postapoc":  {{R: 70, G: 60, B: 50}, {R: 90, G: 80, B: 60}},
	}

	if palette, ok := colors[s.genre]; ok {
		return palette[rng.Intn(len(palette))]
	}
	return color.RGBA{R: 80, G: 80, B: 80, A: 255}
}

func (s *System) getDebrisColor(rng *rand.Rand) color.RGBA {
	colors := map[string][]color.RGBA{
		"fantasy":   {{R: 100, G: 90, B: 80}, {R: 120, G: 110, B: 100}},
		"scifi":     {{R: 140, G: 140, B: 150}, {R: 120, G: 120, B: 130}},
		"horror":    {{R: 80, G: 70, B: 70}, {R: 90, G: 80, B: 80}},
		"cyberpunk": {{R: 160, G: 160, B: 170}, {R: 140, G: 150, B: 160}},
		"postapoc":  {{R: 110, G: 100, B: 90}, {R: 130, G: 120, B: 100}},
	}

	if palette, ok := colors[s.genre]; ok {
		return palette[rng.Intn(len(palette))]
	}
	return color.RGBA{R: 100, G: 100, B: 100, A: 255}
}

func (s *System) getGraffitiColor(rng *rand.Rand) color.RGBA {
	colors := map[string][]color.RGBA{
		"fantasy":   {{R: 180, G: 160, B: 140}},
		"scifi":     {{R: 0, G: 180, B: 255}, {R: 255, G: 100, B: 0}},
		"horror":    {{R: 140, G: 100, B: 100}},
		"cyberpunk": {{R: 255, G: 0, B: 255}, {R: 0, G: 255, B: 200}},
		"postapoc":  {{R: 160, G: 140, B: 120}},
	}

	if palette, ok := colors[s.genre]; ok {
		return palette[rng.Intn(len(palette))]
	}
	return color.RGBA{R: 140, G: 140, B: 140, A: 255}
}

func (s *System) isFloorTile(tile int) bool {
	// Floor tiles are typically 2-9 in most dungeon generators
	return tile >= 2 && tile < 10
}

func (s *System) isNearWall(x, y int, tiles [][]int) bool {
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			ny := y + dy
			nx := x + dx
			if ny >= 0 && ny < len(tiles) && nx >= 0 && nx < len(tiles[0]) {
				if tiles[ny][nx] == 1 || tiles[ny][nx] >= 10 {
					return true
				}
			}
		}
	}
	return false
}

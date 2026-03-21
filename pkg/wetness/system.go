// Package wetness provides surface wetness rendering for environmental realism.
package wetness

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines wetness behavior for each genre.
type GenrePreset struct {
	BaseMoisture        float64 // Base moisture level (0-1)
	PuddleDensity       float64 // How often puddles form (0-1)
	SpecularStrength    float64 // Strength of specular highlights
	TintR, TintG, TintB float64 // Water tint color
	WallBias            float64 // How much puddles prefer wall-adjacent areas
	CornerBias          float64 // How much puddles prefer corners
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BaseMoisture:     0.3,
		PuddleDensity:    0.15,
		SpecularStrength: 0.6,
		TintR:            0.9, TintG: 0.95, TintB: 1.0,
		WallBias:   0.4,
		CornerBias: 0.6,
	},
	"scifi": {
		BaseMoisture:     0.15,
		PuddleDensity:    0.08,
		SpecularStrength: 0.9, // High reflectivity on metal floors
		TintR:            0.85, TintG: 0.9, TintB: 1.0,
		WallBias:   0.2,
		CornerBias: 0.3,
	},
	"horror": {
		BaseMoisture:     0.5,
		PuddleDensity:    0.35,
		SpecularStrength: 0.4,                           // Murky water reflects less
		TintR:            0.7, TintG: 0.75, TintB: 0.65, // Greenish murky tint
		WallBias:   0.5,
		CornerBias: 0.7,
	},
	"cyberpunk": {
		BaseMoisture:     0.45,
		PuddleDensity:    0.4,
		SpecularStrength: 0.85, // Neon reflects well
		TintR:            0.95, TintG: 0.95, TintB: 1.0,
		WallBias:   0.3,
		CornerBias: 0.4,
	},
	"postapoc": {
		BaseMoisture:     0.25,
		PuddleDensity:    0.2,
		SpecularStrength: 0.5,
		TintR:            0.85, TintG: 0.8, TintB: 0.7, // Rust-tinged
		WallBias:   0.5,
		CornerBias: 0.5,
	},
}

// System manages surface wetness rendering.
type System struct {
	genreID          string
	preset           GenrePreset
	screenW, screenH int
	tileSize         int
	cache            map[cacheKey]*ebiten.Image
	cacheLimit       int
	logger           *logrus.Entry
}

type cacheKey struct {
	moisture            float64
	depth               float64
	tintR, tintG, tintB float64
	size                int
}

// NewSystem creates a wetness rendering system.
func NewSystem(genreID string, screenW, screenH int) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	return &System{
		genreID:    genreID,
		preset:     preset,
		screenW:    screenW,
		screenH:    screenH,
		tileSize:   64, // Match floor tile size
		cache:      make(map[cacheKey]*ebiten.Image),
		cacheLimit: 100,
		logger: logrus.WithFields(logrus.Fields{
			"system": "wetness",
			"genre":  genreID,
		}),
	}
}

// SetGenre updates the genre preset.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}
	s.preset = preset
	s.cache = make(map[cacheKey]*ebiten.Image) // Clear cache on genre change
	s.logger = s.logger.WithField("genre", genreID)
}

// Update implements the System interface for ECS compatibility.
func (s *System) Update(w *engine.World) {
	// Wetness patterns are static after generation - no per-frame update needed
}

// GenerateWetnessPattern creates a wetness pattern for a level.
func (s *System) GenerateWetnessPattern(tiles [][]int, seed int64) *WetnessPattern {
	if len(tiles) == 0 || len(tiles[0]) == 0 {
		return nil
	}

	height := len(tiles)
	width := len(tiles[0])

	pattern := &WetnessPattern{
		Width:   width,
		Height:  height,
		Cells:   make([][]*Component, height),
		GenreID: s.genreID,
		Seed:    seed,
	}

	for y := range pattern.Cells {
		pattern.Cells[y] = make([]*Component, width)
	}

	rng := rand.New(rand.NewSource(seed))

	// First pass: base moisture using noise
	moistureMap := s.generateNoiseMap(width, height, seed)

	// Second pass: modify based on walls/corners
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if !s.isFloorTile(tiles[y][x]) {
				continue
			}

			baseMoisture := moistureMap[y][x] * s.preset.BaseMoisture

			// Bias toward walls
			wallFactor := s.countAdjacentWalls(tiles, x, y)
			wallBoost := float64(wallFactor) * s.preset.WallBias * 0.15
			baseMoisture += wallBoost

			// Bias toward corners
			if s.isCorner(tiles, x, y) {
				baseMoisture += s.preset.CornerBias * 0.2
			}

			// Random puddle chance
			if rng.Float64() < s.preset.PuddleDensity {
				baseMoisture += 0.3 + rng.Float64()*0.3
			}

			// Clamp
			baseMoisture = math.Min(1.0, baseMoisture)

			if baseMoisture > 0.1 {
				comp := NewComponent(x, y, baseMoisture, seed+int64(x*1000+y))
				comp.TintR = s.preset.TintR
				comp.TintG = s.preset.TintG
				comp.TintB = s.preset.TintB
				comp.SpecularIntensity = s.preset.SpecularStrength * (0.5 + baseMoisture*0.5)
				pattern.Cells[y][x] = comp
			}
		}
	}

	s.logger.WithFields(logrus.Fields{
		"width":  width,
		"height": height,
		"seed":   seed,
	}).Debug("Generated wetness pattern")

	return pattern
}

// generateNoiseMap creates a 2D noise map for organic puddle shapes.
func (s *System) generateNoiseMap(width, height int, seed int64) [][]float64 {
	noise := make([][]float64, height)
	for y := range noise {
		noise[y] = make([]float64, width)
	}

	rng := rand.New(rand.NewSource(seed ^ 0x574554)) // "WET" XOR

	// Multi-octave noise for organic shapes
	octaves := 3
	persistence := 0.5
	scale := 8.0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			value := 0.0
			amplitude := 1.0
			frequency := 1.0
			maxValue := 0.0

			for o := 0; o < octaves; o++ {
				// Simple value noise interpolation
				fx := float64(x) / scale * frequency
				fy := float64(y) / scale * frequency

				// Get corners
				x0 := int(fx) & 0xFF
				y0 := int(fy) & 0xFF
				x1 := (x0 + 1) & 0xFF
				y1 := (y0 + 1) & 0xFF

				// Fade curves
				sx := fx - math.Floor(fx)
				sy := fy - math.Floor(fy)
				sx = sx * sx * (3 - 2*sx) // Smoothstep
				sy = sy * sy * (3 - 2*sy)

				// Hash and interpolate
				rng.Seed(seed + int64(x0) + int64(y0)*256 + int64(o)*65536)
				n00 := rng.Float64()
				rng.Seed(seed + int64(x1) + int64(y0)*256 + int64(o)*65536)
				n10 := rng.Float64()
				rng.Seed(seed + int64(x0) + int64(y1)*256 + int64(o)*65536)
				n01 := rng.Float64()
				rng.Seed(seed + int64(x1) + int64(y1)*256 + int64(o)*65536)
				n11 := rng.Float64()

				nx0 := n00*(1-sx) + n10*sx
				nx1 := n01*(1-sx) + n11*sx
				nValue := nx0*(1-sy) + nx1*sy

				value += nValue * amplitude
				maxValue += amplitude

				amplitude *= persistence
				frequency *= 2
			}

			noise[y][x] = value / maxValue
		}
	}

	return noise
}

// countAdjacentWalls counts wall tiles adjacent to position.
func (s *System) countAdjacentWalls(tiles [][]int, x, y int) int {
	count := 0
	dirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, d := range dirs {
		nx, ny := x+d[0], y+d[1]
		if ny >= 0 && ny < len(tiles) && nx >= 0 && nx < len(tiles[0]) {
			if !s.isFloorTile(tiles[ny][nx]) {
				count++
			}
		}
	}

	return count
}

// isCorner checks if a floor tile is in a corner (two adjacent walls at 90 degrees).
func (s *System) isCorner(tiles [][]int, x, y int) bool {
	corners := [][2][2]int{
		{{-1, 0}, {0, -1}}, // Top-left
		{{1, 0}, {0, -1}},  // Top-right
		{{-1, 0}, {0, 1}},  // Bottom-left
		{{1, 0}, {0, 1}},   // Bottom-right
	}

	for _, corner := range corners {
		d1, d2 := corner[0], corner[1]
		nx1, ny1 := x+d1[0], y+d1[1]
		nx2, ny2 := x+d2[0], y+d2[1]

		wall1 := ny1 < 0 || ny1 >= len(tiles) || nx1 < 0 || nx1 >= len(tiles[0]) || !s.isFloorTile(tiles[ny1][nx1])
		wall2 := ny2 < 0 || ny2 >= len(tiles) || nx2 < 0 || nx2 >= len(tiles[0]) || !s.isFloorTile(tiles[ny2][nx2])

		if wall1 && wall2 {
			return true
		}
	}

	return false
}

// isFloorTile returns true if the tile value represents a walkable floor.
func (s *System) isFloorTile(tileValue int) bool {
	// Floor tiles are typically 0-9, walls are 10+
	return tileValue >= 0 && tileValue < 10
}

// RenderWetness draws wetness overlays on the screen.
func (s *System) RenderWetness(screen *ebiten.Image, pattern *WetnessPattern, lights []LightSource, cameraX, cameraY float64) {
	if pattern == nil {
		return
	}

	// Calculate visible tile range
	startTileX := int(cameraX) / s.tileSize
	startTileY := int(cameraY) / s.tileSize
	endTileX := startTileX + (s.screenW / s.tileSize) + 2
	endTileY := startTileY + (s.screenH / s.tileSize) + 2

	// Clamp to pattern bounds
	if startTileX < 0 {
		startTileX = 0
	}
	if startTileY < 0 {
		startTileY = 0
	}
	if endTileX > pattern.Width {
		endTileX = pattern.Width
	}
	if endTileY > pattern.Height {
		endTileY = pattern.Height
	}

	for y := startTileY; y < endTileY; y++ {
		for x := startTileX; x < endTileX; x++ {
			comp := pattern.GetComponentAt(x, y)
			if comp == nil || comp.Moisture < 0.1 {
				continue
			}

			screenX := float64(x*s.tileSize) - cameraX
			screenY := float64(y*s.tileSize) - cameraY

			// Skip if off-screen
			if screenX < -float64(s.tileSize) || screenY < -float64(s.tileSize) ||
				screenX > float64(s.screenW) || screenY > float64(s.screenH) {
				continue
			}

			s.renderWetTile(screen, comp, screenX, screenY, lights)
		}
	}
}

// renderWetTile draws a single wet tile overlay.
func (s *System) renderWetTile(screen *ebiten.Image, comp *Component, screenX, screenY float64, lights []LightSource) {
	// Get or create cached wetness sprite
	sprite := s.getWetSprite(comp)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(screenX, screenY)

	// Base alpha from moisture level
	baseAlpha := float32(comp.Moisture * 0.4)

	// Calculate specular from lights
	specular := s.calculateSpecular(comp, screenX, screenY, lights)

	// Combine alpha with specular boost
	totalAlpha := baseAlpha + float32(specular*0.3)
	if totalAlpha > 0.7 {
		totalAlpha = 0.7
	}

	opts.ColorScale.ScaleAlpha(totalAlpha)

	// Apply color tint
	opts.ColorScale.Scale(float32(comp.TintR), float32(comp.TintG), float32(comp.TintB), 1.0)

	screen.DrawImage(sprite, opts)
}

// getWetSprite returns a cached or newly generated wetness sprite.
func (s *System) getWetSprite(comp *Component) *ebiten.Image {
	// Quantize values for cache efficiency
	key := cacheKey{
		moisture: math.Round(comp.Moisture*10) / 10,
		depth:    math.Round(comp.Depth*10) / 10,
		tintR:    math.Round(comp.TintR*10) / 10,
		tintG:    math.Round(comp.TintG*10) / 10,
		tintB:    math.Round(comp.TintB*10) / 10,
		size:     s.tileSize,
	}

	if cached, ok := s.cache[key]; ok {
		return cached
	}

	// Generate new sprite
	sprite := s.generateWetSprite(comp)

	// Cache management
	if len(s.cache) >= s.cacheLimit {
		// Remove one random entry
		for k := range s.cache {
			delete(s.cache, k)
			break
		}
	}
	s.cache[key] = sprite

	return sprite
}

// generateWetSprite creates a wetness overlay sprite.
func (s *System) generateWetSprite(comp *Component) *ebiten.Image {
	size := s.tileSize
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	rng := rand.New(rand.NewSource(comp.Seed))

	// Base wet color (slightly darkening)
	baseR := uint8(20 * comp.TintR)
	baseG := uint8(30 * comp.TintG)
	baseB := uint8(40 * comp.TintB)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Calculate distance from center for radial falloff
			dx := float64(x - size/2)
			dy := float64(y - size/2)
			dist := math.Sqrt(dx*dx+dy*dy) / float64(size/2)

			// Noise-based alpha for organic edges
			noise := s.sampleNoise(x, y, comp.Seed)

			// Edge falloff with noise
			edgeFactor := 1.0 - dist
			if edgeFactor < 0 {
				edgeFactor = 0
			}
			edgeFactor = edgeFactor * edgeFactor // Quadratic falloff

			// Combine with noise for organic shape
			alpha := edgeFactor * (0.5 + noise*0.5) * comp.Moisture

			// Add specular highlight simulation (lighter spots)
			if comp.IsPuddle {
				// Puddles get specular highlight spots
				highlightNoise := s.sampleNoise(x*3, y*3, comp.Seed^0x5350) // "SP" XOR
				if highlightNoise > 0.7 {
					// Bright specular spot
					alpha *= 0.5 // Reduce base alpha where highlight will be
				}
			}

			// Apply dithering for smooth gradients
			ditherThreshold := (rng.Float64() - 0.5) * 0.1
			finalAlpha := alpha + ditherThreshold

			if finalAlpha > 0.05 {
				// Slight color variation
				r := baseR + uint8(rng.Intn(10))
				g := baseG + uint8(rng.Intn(10))
				b := baseB + uint8(rng.Intn(10))
				a := uint8(clamp(finalAlpha*255, 0, 255))

				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
			}
		}
	}

	// Add specular highlights for puddles
	if comp.IsPuddle {
		s.addSpecularHighlights(img, comp, rng)
	}

	return ebiten.NewImageFromImage(img)
}

// addSpecularHighlights adds bright specular spots to puddles.
func (s *System) addSpecularHighlights(img *image.RGBA, comp *Component, rng *rand.Rand) {
	size := s.tileSize
	numHighlights := 1 + int(comp.Depth*3)

	for i := 0; i < numHighlights; i++ {
		// Position highlights toward one side (simulating directional light)
		hx := size/4 + rng.Intn(size/2)
		hy := size/4 + rng.Intn(size/3)

		// Highlight size based on depth
		radius := 2 + int(comp.Depth*4)

		// Draw highlight with soft edges
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				px, py := hx+dx, hy+dy
				if px < 0 || py < 0 || px >= size || py >= size {
					continue
				}

				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist > float64(radius) {
					continue
				}

				// Soft falloff
				intensity := 1.0 - dist/float64(radius)
				intensity = intensity * intensity * comp.SpecularIntensity

				// Blend with existing color
				existing := img.RGBAAt(px, py)
				if existing.A < 10 {
					continue // Don't add highlights outside wet area
				}

				// Add white specular
				specR := uint8(clamp(float64(existing.R)+intensity*200, 0, 255))
				specG := uint8(clamp(float64(existing.G)+intensity*200, 0, 255))
				specB := uint8(clamp(float64(existing.B)+intensity*200, 0, 255))

				img.Set(px, py, color.RGBA{R: specR, G: specG, B: specB, A: existing.A})
			}
		}
	}
}

// sampleNoise returns a deterministic noise value for position.
func (s *System) sampleNoise(x, y int, seed int64) float64 {
	// Simple deterministic hash-based noise
	h := seed + int64(x)*374761393 + int64(y)*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h = h ^ (h >> 16)

	return float64(h&0xFFFF) / 65536.0
}

// calculateSpecular computes specular reflection from lights.
func (s *System) calculateSpecular(comp *Component, screenX, screenY float64, lights []LightSource) float64 {
	if len(lights) == 0 || !comp.IsPuddle {
		return 0
	}

	totalSpec := 0.0
	tileCenterX := screenX + float64(s.tileSize)/2
	tileCenterY := screenY + float64(s.tileSize)/2

	for _, light := range lights {
		dx := light.X - tileCenterX
		dy := light.Y - tileCenterY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > light.Radius {
			continue
		}

		// Distance falloff (inverse square)
		falloff := 1.0 - (dist / light.Radius)
		falloff = falloff * falloff

		// Specular calculation (simplified Phong)
		spec := falloff * light.Intensity * comp.SpecularIntensity * 0.5

		// Flicker modulation
		if light.FlickerAmount > 0 {
			flicker := 0.8 + 0.2*math.Sin(light.FlickerPhase*6.28)
			spec *= flicker
		}

		totalSpec += spec
	}

	return math.Min(totalSpec, 1.0)
}

// clamp constrains a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

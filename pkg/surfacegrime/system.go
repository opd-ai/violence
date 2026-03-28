package surfacegrime

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// System manages surface grime rendering for realistic environmental weathering.
type System struct {
	genreID string
	preset  GenreGrime
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry

	// Cached overlay for current room
	cachedRoomID  string
	cachedOverlay *ebiten.Image
	overlayWidth  int
	overlayHeight int

	// Edge map input (from edgeAO or level geometry)
	edgeMap [][]float64 // [y][x] edge proximity values 0.0-1.0

	// Configuration
	enabled   bool
	intensity float64 // Global intensity multiplier
}

// NewSystem creates a surface grime system with genre-appropriate settings.
func NewSystem(genreID string, seed int64) *System {
	s := &System{
		genreID:   genreID,
		seed:      seed,
		rng:       rand.New(rand.NewSource(seed)),
		enabled:   true,
		intensity: 1.0,
		logger: logrus.WithFields(logrus.Fields{
			"system": "surfacegrime",
		}),
	}
	s.applyGenrePreset(genreID)
	return s
}

// applyGenrePreset sets grime parameters based on genre.
func (s *System) applyGenrePreset(genreID string) {
	s.genreID = genreID
	s.preset = GetGenreGrime(genreID)
	s.logger.WithField("genre", genreID).Debug("Applied grime preset")
}

// SetGenre updates the genre and resets cached overlays.
func (s *System) SetGenre(genreID string) {
	if s.genreID == genreID {
		return
	}
	s.applyGenrePreset(genreID)
	s.cachedRoomID = ""
	s.cachedOverlay = nil
}

// SetEnabled toggles grime rendering on/off.
func (s *System) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// SetIntensity sets the global grime intensity multiplier.
func (s *System) SetIntensity(intensity float64) {
	s.intensity = clamp(intensity, 0.0, 2.0)
}

// SetEdgeMap provides edge proximity data for accumulation zones.
// Values should be 0.0 (no edge) to 1.0 (at edge/corner).
func (s *System) SetEdgeMap(edgeMap [][]float64) {
	s.edgeMap = edgeMap
	s.cachedRoomID = "" // Invalidate cache
}

// GenerateOverlay creates a grime overlay for the specified dimensions.
func (s *System) GenerateOverlay(width, height int, roomID string, seed int64) *ebiten.Image {
	// Check cache
	if s.cachedOverlay != nil && s.cachedRoomID == roomID &&
		s.overlayWidth == width && s.overlayHeight == height {
		return s.cachedOverlay
	}

	// Generate new overlay
	rng := rand.New(rand.NewSource(seed))
	rgba := image.NewRGBA(image.Rect(0, 0, width, height))

	// Generate grime for each pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get edge proximity (0.0 = no edge, 1.0 = at edge)
			edgeProximity := s.getEdgeProximity(x, y, width, height)

			// Skip if no edge nearby
			if edgeProximity < 0.05 {
				continue
			}

			// Calculate grime accumulation
			accumulation := s.calculateAccumulation(x, y, edgeProximity, width, height, rng)
			if accumulation < 0.05 {
				continue
			}

			// Select grime type based on weights
			grimeType := s.selectGrimeType(x, y, rng)

			// Get grime color with variation
			grimeColor := s.getGrimeColor(grimeType, x, y, rng)

			// Apply accumulation to alpha
			grimeColor.A = uint8(float64(grimeColor.A) * accumulation * s.intensity)

			// Set pixel with grime
			rgba.SetRGBA(x, y, grimeColor)
		}
	}

	// Apply noise pass for natural variation
	s.applyNoisePass(rgba, rng)

	// Apply edge feathering for smooth transitions
	s.applyEdgeFeathering(rgba)

	// Create ebiten image
	img := ebiten.NewImage(width, height)
	img.WritePixels(rgba.Pix)

	// Cache result
	s.cachedRoomID = roomID
	s.cachedOverlay = img
	s.overlayWidth = width
	s.overlayHeight = height

	s.logger.WithFields(logrus.Fields{
		"room_id": roomID,
		"width":   width,
		"height":  height,
	}).Debug("Generated grime overlay")

	return img
}

// getEdgeProximity returns edge proximity at a pixel location.
func (s *System) getEdgeProximity(x, y, width, height int) float64 {
	// If we have an edge map, use it
	if s.edgeMap != nil && len(s.edgeMap) > 0 {
		mapY := y * len(s.edgeMap) / height
		if mapY >= len(s.edgeMap) {
			mapY = len(s.edgeMap) - 1
		}
		if len(s.edgeMap[mapY]) > 0 {
			mapX := x * len(s.edgeMap[mapY]) / width
			if mapX >= len(s.edgeMap[mapY]) {
				mapX = len(s.edgeMap[mapY]) - 1
			}
			return s.edgeMap[mapY][mapX]
		}
	}

	// Fallback: simulate edges at screen borders and at regular intervals
	// This creates a generic "wall base" effect along the bottom of the screen
	edgeProximity := 0.0

	// Bottom edge (floor/wall junction)
	bottomDist := float64(height - y)
	if bottomDist < float64(s.preset.SpreadDistance) {
		edgeProximity = math.Max(edgeProximity, 1.0-bottomDist/float64(s.preset.SpreadDistance))
	}

	// Side edges
	leftDist := float64(x)
	rightDist := float64(width - x)
	if leftDist < float64(s.preset.SpreadDistance/2) {
		edgeProximity = math.Max(edgeProximity, (1.0-leftDist/float64(s.preset.SpreadDistance/2))*0.7)
	}
	if rightDist < float64(s.preset.SpreadDistance/2) {
		edgeProximity = math.Max(edgeProximity, (1.0-rightDist/float64(s.preset.SpreadDistance/2))*0.7)
	}

	return edgeProximity
}

// calculateAccumulation determines grime intensity at a pixel.
func (s *System) calculateAccumulation(x, y int, edgeProximity float64, width, height int, rng *rand.Rand) float64 {
	base := edgeProximity

	// Increase in corners (where two edges meet)
	isCorner := false
	if x < s.preset.SpreadDistance && y > height-s.preset.SpreadDistance {
		isCorner = true
	}
	if x > width-s.preset.SpreadDistance && y > height-s.preset.SpreadDistance {
		isCorner = true
	}
	if isCorner {
		base *= s.preset.CornerIntensity
	}

	// Add noise for natural variation
	noiseX := float64(x) * s.preset.NoiseScale
	noiseY := float64(y) * s.preset.NoiseScale
	noise := s.perlinNoise(noiseX, noiseY, int64(x*7919+y*104729)+s.seed)
	base *= 0.6 + noise*0.4

	// Random variation
	base *= 0.8 + rng.Float64()*0.4

	return clamp(base, 0.0, 1.0)
}

// selectGrimeType picks a grime type based on position and genre weights.
func (s *System) selectGrimeType(x, y int, rng *rand.Rand) GrimeType {
	if len(s.preset.Types) == 0 {
		return GrimeDirt
	}

	// Weight-based selection
	total := 0.0
	for _, w := range s.preset.Weights {
		total += w
	}

	roll := rng.Float64() * total
	cumulative := 0.0
	for i, w := range s.preset.Weights {
		cumulative += w
		if roll < cumulative && i < len(s.preset.Types) {
			return s.preset.Types[i]
		}
	}

	return s.preset.Types[0]
}

// getGrimeColor returns a color for the grime type with subtle variation.
func (s *System) getGrimeColor(grimeType GrimeType, x, y int, rng *rand.Rand) color.RGBA {
	baseColor := GrimeColors[grimeType]

	// Add variation
	secondaryColors := GrimeSecondaryColors[grimeType]
	if len(secondaryColors) > 0 && rng.Float64() < 0.4 {
		idx := rng.Intn(len(secondaryColors))
		secondary := secondaryColors[idx]
		// Blend with secondary color
		t := rng.Float64() * 0.5
		baseColor.R = uint8(float64(baseColor.R)*(1-t) + float64(secondary.R)*t)
		baseColor.G = uint8(float64(baseColor.G)*(1-t) + float64(secondary.G)*t)
		baseColor.B = uint8(float64(baseColor.B)*(1-t) + float64(secondary.B)*t)
	}

	// Subtle per-pixel variation
	variation := (rng.Float64() - 0.5) * 0.1
	baseColor.R = clampByte(int(float64(baseColor.R) * (1 + variation)))
	baseColor.G = clampByte(int(float64(baseColor.G) * (1 + variation)))
	baseColor.B = clampByte(int(float64(baseColor.B) * (1 + variation)))

	return baseColor
}

// applyNoisePass adds procedural noise for natural texture.
func (s *System) applyNoisePass(img *image.RGBA, rng *rand.Rand) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}

			// Add subtle brightness noise
			noiseVal := (rng.Float64() - 0.5) * 0.15
			c.R = clampByte(int(float64(c.R) * (1 + noiseVal)))
			c.G = clampByte(int(float64(c.G) * (1 + noiseVal)))
			c.B = clampByte(int(float64(c.B) * (1 + noiseVal)))

			img.SetRGBA(x, y, c)
		}
	}
}

// applyEdgeFeathering smooths grime edges for natural transitions.
func (s *System) applyEdgeFeathering(img *image.RGBA) {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Simple 3x3 blur on alpha channel only
	alphas := make([][]uint8, height)
	for y := 0; y < height; y++ {
		alphas[y] = make([]uint8, width)
		for x := 0; x < width; x++ {
			alphas[y][x] = img.RGBAAt(x, y).A
		}
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			if alphas[y][x] == 0 {
				continue
			}

			// Sample neighbors
			sum := 0
			count := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					sum += int(alphas[y+dy][x+dx])
					count++
				}
			}
			avg := sum / count

			c := img.RGBAAt(x, y)
			// Blend toward average (gentle feathering)
			c.A = uint8((int(c.A) + avg) / 2)
			img.SetRGBA(x, y, c)
		}
	}
}

// perlinNoise generates simple value noise for natural variation.
func (s *System) perlinNoise(x, y float64, seed int64) float64 {
	// Simple hash-based noise
	xi := int(math.Floor(x))
	yi := int(math.Floor(y))
	xf := x - float64(xi)
	yf := y - float64(yi)

	// Smooth interpolation
	u := xf * xf * (3 - 2*xf)
	v := yf * yf * (3 - 2*yf)

	// Hash corners
	n00 := s.hash(xi, yi, seed)
	n10 := s.hash(xi+1, yi, seed)
	n01 := s.hash(xi, yi+1, seed)
	n11 := s.hash(xi+1, yi+1, seed)

	// Bilinear interpolation
	nx0 := n00*(1-u) + n10*u
	nx1 := n01*(1-u) + n11*u
	return nx0*(1-v) + nx1*v
}

// hash generates a pseudo-random value from coordinates.
func (s *System) hash(x, y int, seed int64) float64 {
	h := seed + int64(x)*374761393 + int64(y)*668265263
	h = (h ^ (h >> 13)) * 1274126177
	h = h ^ (h >> 16)
	return float64(h&0xFFFF) / 65535.0
}

// Draw renders the grime overlay onto the target.
func (s *System) Draw(target *ebiten.Image, roomID string, seed int64) {
	if !s.enabled || s.intensity < 0.01 {
		return
	}

	width := target.Bounds().Dx()
	height := target.Bounds().Dy()

	overlay := s.GenerateOverlay(width, height, roomID, seed)
	if overlay == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	target.DrawImage(overlay, op)
}

// Update processes any time-dependent grime effects.
func (s *System) Update(deltaTime float64) {
	// Currently static - could add animated effects like dripping or spreading
}

// clamp restricts a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// clampByte restricts an int to byte range.
func clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

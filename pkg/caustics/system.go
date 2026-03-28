package caustics

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines caustic behavior for each genre.
type GenrePreset struct {
	// BaseIntensity scales overall caustic brightness.
	BaseIntensity float64

	// AnimationSpeed controls how fast the caustics animate (Hz).
	AnimationSpeed float64

	// PatternScale controls the size of caustic cells.
	PatternScale float64

	// WaveAmplitude controls how much the pattern shifts.
	WaveAmplitude float64

	// SecondaryWaveFreq adds complexity with a second wave.
	SecondaryWaveFreq float64

	// ColorR, ColorG, ColorB are the base caustic tint (0-1).
	ColorR, ColorG, ColorB float64

	// ColorVariation adds randomness to caustic colors.
	ColorVariation float64

	// FalloffExponent controls edge softness.
	FalloffExponent float64

	// JitterAmount adds irregularity to the animation.
	JitterAmount float64

	// PuddleRadius is default caustic radius for puddles.
	PuddleRadius float64

	// PoolRadius is default caustic radius for pools.
	PoolRadius float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BaseIntensity:     0.6,
		AnimationSpeed:    1.5,
		PatternScale:      12.0,
		WaveAmplitude:     3.0,
		SecondaryWaveFreq: 0.7,
		ColorR:            0.7, ColorG: 0.85, ColorB: 1.0, // Blue-white
		ColorVariation:  0.1,
		FalloffExponent: 2.0,
		JitterAmount:    0.1,
		PuddleRadius:    2.0,
		PoolRadius:      4.0,
	},
	"scifi": {
		BaseIntensity:     0.5,
		AnimationSpeed:    2.5,
		PatternScale:      8.0,
		WaveAmplitude:     2.0,
		SecondaryWaveFreq: 1.2,
		ColorR:            0.5, ColorG: 0.9, ColorB: 1.0, // Cyan
		ColorVariation:  0.05,
		FalloffExponent: 3.0, // Sharp edges
		JitterAmount:    0.05,
		PuddleRadius:    1.5,
		PoolRadius:      3.0,
	},
	"horror": {
		BaseIntensity:     0.45,
		AnimationSpeed:    0.8,
		PatternScale:      15.0,
		WaveAmplitude:     4.5,
		SecondaryWaveFreq: 0.3,
		ColorR:            0.6, ColorG: 0.7, ColorB: 0.5, // Murky green
		ColorVariation:  0.2,
		FalloffExponent: 1.5,
		JitterAmount:    0.25, // Erratic
		PuddleRadius:    2.5,
		PoolRadius:      5.0,
	},
	"cyberpunk": {
		BaseIntensity:     0.7,
		AnimationSpeed:    3.0,
		PatternScale:      10.0,
		WaveAmplitude:     2.5,
		SecondaryWaveFreq: 1.5,
		ColorR:            0.9, ColorG: 0.6, ColorB: 1.0, // Neon magenta
		ColorVariation:  0.3, // High variation for neon effects
		FalloffExponent: 2.5,
		JitterAmount:    0.15,
		PuddleRadius:    2.0,
		PoolRadius:      3.5,
	},
	"postapoc": {
		BaseIntensity:     0.35,
		AnimationSpeed:    0.6,
		PatternScale:      14.0,
		WaveAmplitude:     3.5,
		SecondaryWaveFreq: 0.4,
		ColorR:            0.8, ColorG: 0.7, ColorB: 0.4, // Murky yellow-brown
		ColorVariation:  0.15,
		FalloffExponent: 1.8,
		JitterAmount:    0.2,
		PuddleRadius:    1.8,
		PoolRadius:      4.0,
	},
}

// System manages water caustic generation and rendering.
type System struct {
	genreID  string
	preset   GenrePreset
	screenW  int
	screenH  int
	tileSize int

	// Animation state
	time     float64
	frameIdx int

	// Caustic sources in the level
	sources []*Component

	// Pattern cache for performance
	patternCache map[cacheKey]*ebiten.Image
	cacheLimit   int
	cacheMu      sync.RWMutex

	// Precomputed lookup tables
	sinTable []float64
	cosTable []float64

	logger *logrus.Entry
}

type cacheKey struct {
	frame   int
	radius  float64
	sourceT SourceType
}

// NewSystem creates a water caustics rendering system.
func NewSystem(genreID string, screenW, screenH int) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	// Precompute trig tables for performance
	tableSize := 360
	sinTable := make([]float64, tableSize)
	cosTable := make([]float64, tableSize)
	for i := 0; i < tableSize; i++ {
		angle := float64(i) * math.Pi / 180.0
		sinTable[i] = math.Sin(angle)
		cosTable[i] = math.Cos(angle)
	}

	return &System{
		genreID:      genreID,
		preset:       preset,
		screenW:      screenW,
		screenH:      screenH,
		tileSize:     64,
		sources:      make([]*Component, 0),
		patternCache: make(map[cacheKey]*ebiten.Image),
		cacheLimit:   50,
		sinTable:     sinTable,
		cosTable:     cosTable,
		logger: logrus.WithFields(logrus.Fields{
			"system": "caustics",
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
	s.ClearCache()
	s.logger = s.logger.WithField("genre", genreID)
}

// Update advances the caustic animation.
func (s *System) Update(w *engine.World) {
	dt := 1.0 / 60.0 // Assume 60 FPS
	s.time += dt * s.preset.AnimationSpeed
	s.frameIdx = int(s.time*10) % 100 // 100 animation frames
}

// ClearCache removes all cached patterns.
func (s *System) ClearCache() {
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()
	s.patternCache = make(map[cacheKey]*ebiten.Image)
}

// AddSource adds a caustic source to the system.
func (s *System) AddSource(comp *Component) {
	s.sources = append(s.sources, comp)
}

// ClearSources removes all caustic sources.
func (s *System) ClearSources() {
	s.sources = s.sources[:0]
}

// GenerateCausticsFromWetness creates caustic sources from a wetness pattern.
// This integrates with the wetness system to place caustics near puddles.
func (s *System) GenerateCausticsFromWetness(puddleLocations []PuddleLocation, seed int64) {
	s.ClearSources()
	rng := rand.New(rand.NewSource(seed))

	for _, puddle := range puddleLocations {
		// Determine source type based on moisture level
		var sourceType SourceType
		var radius float64
		if puddle.Moisture > 0.8 {
			sourceType = SourcePool
			radius = s.preset.PoolRadius
		} else if puddle.Moisture > 0.5 {
			sourceType = SourcePuddle
			radius = s.preset.PuddleRadius
		} else {
			sourceType = SourceDrip
			radius = s.preset.PuddleRadius * 0.5
		}

		// Calculate caustic color with variation
		col := s.calculateCausticColor(puddle.Moisture, rng)

		// Create caustic component
		comp := NewComponent(
			puddle.WorldX,
			puddle.WorldY,
			puddle.Moisture*s.preset.BaseIntensity,
			radius,
			rng.Float64()*math.Pi*2, // Random phase offset
			sourceType,
			col,
			seed^int64(puddle.TileX)^(int64(puddle.TileY)<<16),
		)
		s.AddSource(comp)
	}

	s.logger.WithField("count", len(s.sources)).Debug("Generated caustic sources")
}

// PuddleLocation represents a water source location for caustic generation.
type PuddleLocation struct {
	TileX, TileY int
	WorldX, WorldY float64
	Moisture float64
}

// calculateCausticColor determines the caustic tint with genre variation.
func (s *System) calculateCausticColor(moisture float64, rng *rand.Rand) color.RGBA {
	p := s.preset

	// Base color from preset
	r := p.ColorR
	g := p.ColorG
	b := p.ColorB

	// Add variation
	variation := p.ColorVariation
	r += (rng.Float64()*2 - 1) * variation
	g += (rng.Float64()*2 - 1) * variation
	b += (rng.Float64()*2 - 1) * variation

	// Clamp
	r = clampFloat(r, 0, 1)
	g = clampFloat(g, 0, 1)
	b = clampFloat(b, 0, 1)

	// Intensity based on moisture
	intensity := 0.5 + moisture*0.5

	return color.RGBA{
		R: uint8(r * intensity * 255),
		G: uint8(g * intensity * 255),
		B: uint8(b * intensity * 255),
		A: 255,
	}
}

// Render draws caustics onto the screen.
func (s *System) Render(screen *ebiten.Image, cameraX, cameraY float64) {
	if len(s.sources) == 0 {
		return
	}

	screenCenterX := float64(s.screenW) / 2
	screenCenterY := float64(s.screenH) / 2

	for _, src := range s.sources {
		// Calculate screen position
		worldDX := src.WorldX - cameraX
		worldDY := src.WorldY - cameraY

		// Simple perspective projection
		screenX := screenCenterX + worldDX*float64(s.tileSize)
		screenY := screenCenterY + worldDY*float64(s.tileSize)

		// Culling - skip if off screen
		radiusPx := src.Radius * float64(s.tileSize)
		if screenX+radiusPx < 0 || screenX-radiusPx > float64(s.screenW) ||
			screenY+radiusPx < 0 || screenY-radiusPx > float64(s.screenH) {
			continue
		}

		// Get or generate caustic pattern for this frame
		pattern := s.getPattern(src)
		if pattern == nil {
			continue
		}

		// Draw the caustic pattern
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX-radiusPx, screenY-radiusPx)
		opts.ColorScale.Scale(
			float32(src.Color.R)/255.0,
			float32(src.Color.G)/255.0,
			float32(src.Color.B)/255.0,
			float32(src.Intensity),
		)
		opts.Blend = ebiten.BlendSourceOver

		screen.DrawImage(pattern, opts)
	}
}

// getPattern retrieves or generates a caustic pattern.
func (s *System) getPattern(src *Component) *ebiten.Image {
	key := cacheKey{
		frame:   s.frameIdx,
		radius:  src.Radius,
		sourceT: src.SourceType,
	}

	s.cacheMu.RLock()
	if pattern, ok := s.patternCache[key]; ok {
		s.cacheMu.RUnlock()
		return pattern
	}
	s.cacheMu.RUnlock()

	// Generate new pattern
	pattern := s.generatePattern(src)

	s.cacheMu.Lock()
	// LRU eviction
	if len(s.patternCache) >= s.cacheLimit {
		// Remove oldest entry
		for k := range s.patternCache {
			delete(s.patternCache, k)
			break
		}
	}
	s.patternCache[key] = pattern
	s.cacheMu.Unlock()

	return pattern
}

// generatePattern creates a caustic pattern for a source.
func (s *System) generatePattern(src *Component) *ebiten.Image {
	radiusPx := int(src.Radius * float64(s.tileSize))
	if radiusPx < 4 {
		radiusPx = 4
	}
	size := radiusPx * 2

	img := image.NewRGBA(image.Rect(0, 0, size, size))

	p := s.preset
	centerX := float64(size) / 2
	centerY := float64(size) / 2
	time := s.time + src.Phase

	// Generate Voronoi-like caustic pattern
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Distance from center for falloff
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := float64(radiusPx)

			if dist > maxDist {
				continue
			}

			// Calculate caustic intensity at this point
			intensity := s.calculateCausticIntensity(
				float64(x), float64(y),
				centerX, centerY,
				time, src, p,
			)

			// Apply radial falloff
			falloff := 1.0 - math.Pow(dist/maxDist, p.FalloffExponent)
			intensity *= falloff

			// Clamp and apply
			if intensity > 0.01 {
				alpha := uint8(clampFloat(intensity, 0, 1) * 255)
				img.SetRGBA(x, y, color.RGBA{255, 255, 255, alpha})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// calculateCausticIntensity computes the caustic brightness at a point.
func (s *System) calculateCausticIntensity(x, y, cx, cy, time float64, src *Component, p GenrePreset) float64 {
	// Normalize coordinates to pattern space
	scale := p.PatternScale
	nx := (x - cx) / scale
	ny := (y - cy) / scale

	// Primary wave (creates the cell structure)
	wave1 := s.voronoiNoise(nx, ny, time, src.Seed)

	// Secondary wave (adds complexity)
	wave2 := s.voronoiNoise(
		nx*p.SecondaryWaveFreq+time*0.3,
		ny*p.SecondaryWaveFreq-time*0.2,
		time*0.5,
		src.Seed^0x5CA7,
	)

	// Combine waves
	combined := wave1*0.7 + wave2*0.3

	// Add jitter for organic feel
	if p.JitterAmount > 0 {
		jitter := math.Sin(time*17+nx*3)*math.Cos(time*13+ny*5) * p.JitterAmount
		combined += jitter
	}

	// Apply wave amplitude
	combined *= p.WaveAmplitude * 0.1

	// Convert to caustic brightness (bright lines between cells)
	// Higher values near cell edges
	caustic := 1.0 - combined
	caustic = math.Pow(caustic, 3) // Sharpen the caustic lines

	return caustic * p.BaseIntensity
}

// voronoiNoise generates Voronoi-based noise for caustic patterns.
func (s *System) voronoiNoise(x, y, time float64, seed int64) float64 {
	// Grid cell
	cellX := int(math.Floor(x))
	cellY := int(math.Floor(y))

	minDist := 10.0

	// Check neighboring cells
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			neighborX := cellX + dx
			neighborY := cellY + dy

			// Generate deterministic random point in cell
			cellSeed := seed ^ int64(neighborX*73856093) ^ int64(neighborY*19349663)
			rng := rand.New(rand.NewSource(cellSeed))

			// Point position with time-based animation
			pointX := float64(neighborX) + rng.Float64()
			pointY := float64(neighborY) + rng.Float64()

			// Animate the point
			animPhase := rng.Float64() * math.Pi * 2
			animSpeed := 0.5 + rng.Float64()*0.5
			pointX += math.Sin(time*animSpeed+animPhase) * 0.3
			pointY += math.Cos(time*animSpeed*0.7+animPhase) * 0.3

			// Distance to this point
			distX := x - pointX
			distY := y - pointY
			dist := math.Sqrt(distX*distX + distY*distY)

			if dist < minDist {
				minDist = dist
			}
		}
	}

	return minDist
}

// GetSourceCount returns the number of caustic sources.
func (s *System) GetSourceCount() int {
	return len(s.sources)
}

// GetGenre returns the current genre ID.
func (s *System) GetGenre() string {
	return s.genreID
}

// GetTime returns the current animation time.
func (s *System) GetTime() float64 {
	return s.time
}

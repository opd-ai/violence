package lensdirt

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines lens dirt characteristics per genre.
type GenrePreset struct {
	// DirtDensity controls how many specks are generated (0-1).
	DirtDensity float64

	// SmudgeRatio is the proportion of smudge vs dust specks (0-1).
	SmudgeRatio float64

	// StreakRatio is the proportion of streak patterns (0-1, from remaining after smudge).
	StreakRatio float64

	// BaseTint colors all dirt elements.
	BaseTint color.RGBA

	// IntensityScale multiplies overall dirt visibility.
	IntensityScale float64

	// FalloffDistance controls how far from lights dirt is visible (in screen pixels).
	FalloffDistance float64

	// EdgeDarkening adds vignette-like darkening to lens edges.
	EdgeDarkening float64

	// HexFlareEnabled adds hexagonal lens flare artifacts around bright lights.
	HexFlareEnabled bool

	// DiffuseHaloStrength controls soft glow halo around lights.
	DiffuseHaloStrength float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		DirtDensity:         0.6,
		SmudgeRatio:         0.3,
		StreakRatio:         0.1,
		BaseTint:            color.RGBA{R: 200, G: 180, B: 140, A: 255},
		IntensityScale:      0.4,
		FalloffDistance:     150,
		EdgeDarkening:       0.15,
		HexFlareEnabled:     false,
		DiffuseHaloStrength: 0.3,
	},
	"scifi": {
		DirtDensity:         0.25,
		SmudgeRatio:         0.1,
		StreakRatio:         0.05,
		BaseTint:            color.RGBA{R: 180, G: 200, B: 220, A: 255},
		IntensityScale:      0.25,
		FalloffDistance:     100,
		EdgeDarkening:       0.05,
		HexFlareEnabled:     true,
		DiffuseHaloStrength: 0.2,
	},
	"horror": {
		DirtDensity:         0.8,
		SmudgeRatio:         0.5,
		StreakRatio:         0.15,
		BaseTint:            color.RGBA{R: 150, G: 170, B: 140, A: 255},
		IntensityScale:      0.55,
		FalloffDistance:     180,
		EdgeDarkening:       0.25,
		HexFlareEnabled:     false,
		DiffuseHaloStrength: 0.4,
	},
	"cyberpunk": {
		DirtDensity:         0.5,
		SmudgeRatio:         0.2,
		StreakRatio:         0.4,
		BaseTint:            color.RGBA{R: 200, G: 180, B: 220, A: 255},
		IntensityScale:      0.35,
		FalloffDistance:     120,
		EdgeDarkening:       0.1,
		HexFlareEnabled:     true,
		DiffuseHaloStrength: 0.35,
	},
	"postapoc": {
		DirtDensity:         0.9,
		SmudgeRatio:         0.4,
		StreakRatio:         0.2,
		BaseTint:            color.RGBA{R: 180, G: 160, B: 120, A: 255},
		IntensityScale:      0.6,
		FalloffDistance:     200,
		EdgeDarkening:       0.3,
		HexFlareEnabled:     false,
		DiffuseHaloStrength: 0.45,
	},
}

// System manages lens dirt rendering.
type System struct {
	genreID string
	preset  GenrePreset
	logger  *logrus.Entry
	rng     *rand.Rand
	seed    int64

	screenW, screenH int

	// Cached dirt pattern
	pattern *DirtPattern

	// Light sources for current frame
	lights   []LightSource
	lightsMu sync.RWMutex

	// Pre-rendered sprites for each speck shape
	speckCache   map[speckCacheKey]*ebiten.Image
	speckCacheMu sync.RWMutex
	maxCache     int

	// Overlay buffer (reused each frame)
	overlay *ebiten.Image
}

type speckCacheKey struct {
	shape   SpeckShape
	size    int
	tintR   uint8
	tintG   uint8
	tintB   uint8
	rotStep int // rotation in 15-degree steps
}

// NewSystem creates a lens dirt system.
func NewSystem(genreID string, seed int64, screenW, screenH int) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	s := &System{
		genreID:    genreID,
		preset:     preset,
		logger:     logrus.WithFields(logrus.Fields{"system_name": "lensdirt", "genre": genreID}),
		rng:        rand.New(rand.NewSource(seed)),
		seed:       seed,
		screenW:    screenW,
		screenH:    screenH,
		speckCache: make(map[speckCacheKey]*ebiten.Image),
		maxCache:   64,
	}

	s.generatePattern()
	s.logger.Debug("Lens dirt system initialized")

	return s
}

// SetGenre updates the system for a new genre.
func (s *System) SetGenre(genreID string) {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}
	s.genreID = genreID
	s.preset = preset
	s.logger = s.logger.WithField("genre", genreID)

	// Regenerate pattern for new genre
	s.generatePattern()

	// Clear sprite cache
	s.speckCacheMu.Lock()
	s.speckCache = make(map[speckCacheKey]*ebiten.Image)
	s.speckCacheMu.Unlock()
}

// SetScreenSize updates the screen dimensions.
func (s *System) SetScreenSize(w, h int) {
	if w != s.screenW || h != s.screenH {
		s.screenW = w
		s.screenH = h
		s.overlay = nil // Force recreate
		s.generatePattern()
	}
}

// SetLightSources updates the current frame's light sources.
func (s *System) SetLightSources(lights []LightSource) {
	s.lightsMu.Lock()
	s.lights = lights
	s.lightsMu.Unlock()
}

// AddLightSource adds a single light source for the current frame.
func (s *System) AddLightSource(light LightSource) {
	s.lightsMu.Lock()
	s.lights = append(s.lights, light)
	s.lightsMu.Unlock()
}

// ClearLights removes all light sources.
func (s *System) ClearLights() {
	s.lightsMu.Lock()
	s.lights = s.lights[:0]
	s.lightsMu.Unlock()
}

// GetPreset returns the current genre preset.
func (s *System) GetPreset() GenrePreset {
	return s.preset
}

// generatePattern creates the procedural dirt layout.
func (s *System) generatePattern() {
	rng := rand.New(rand.NewSource(s.seed))

	// Calculate speck count based on density and screen area
	baseCount := int(float64(s.screenW*s.screenH) / 2000.0 * s.preset.DirtDensity)
	if baseCount < 10 {
		baseCount = 10
	}
	if baseCount > 200 {
		baseCount = 200
	}

	specks := make([]DirtSpeck, 0, baseCount)

	for i := 0; i < baseCount; i++ {
		speck := DirtSpeck{
			X:           rng.Float64(),
			Y:           rng.Float64(),
			Size:        4 + rng.Float64()*12,
			BaseOpacity: 0.3 + rng.Float64()*0.4,
			Rotation:    rng.Float64() * 2 * math.Pi,
		}

		// Determine shape based on genre ratios
		roll := rng.Float64()
		if roll < s.preset.SmudgeRatio {
			speck.Shape = ShapeSmudge
			speck.Size *= 1.5 // Smudges are larger
		} else if roll < s.preset.SmudgeRatio+s.preset.StreakRatio {
			speck.Shape = ShapeStreaky
			speck.Size *= 2.0 // Streaks are elongated
		} else if rng.Float64() < 0.1 {
			speck.Shape = ShapeDiffuse
			speck.Size *= 3.0 // Diffuse glows are large
			speck.BaseOpacity *= 0.5
		} else {
			speck.Shape = ShapeCircle
		}

		// Apply tint variation
		tintVar := 0.1
		speck.Tint = color.RGBA{
			R: uint8(clamp01(float64(s.preset.BaseTint.R)/255.0+rng.Float64()*tintVar*2-tintVar) * 255),
			G: uint8(clamp01(float64(s.preset.BaseTint.G)/255.0+rng.Float64()*tintVar*2-tintVar) * 255),
			B: uint8(clamp01(float64(s.preset.BaseTint.B)/255.0+rng.Float64()*tintVar*2-tintVar) * 255),
			A: 255,
		}

		specks = append(specks, speck)
	}

	// Add hexagonal flare elements if enabled
	if s.preset.HexFlareEnabled {
		for i := 0; i < 6; i++ {
			speck := DirtSpeck{
				X:           rng.Float64(),
				Y:           rng.Float64(),
				Size:        8 + rng.Float64()*8,
				BaseOpacity: 0.15 + rng.Float64()*0.15,
				Shape:       ShapeHexagonal,
				Rotation:    float64(i) * math.Pi / 3,
				Tint:        s.preset.BaseTint,
			}
			specks = append(specks, speck)
		}
	}

	s.pattern = &DirtPattern{
		Specks: specks,
		Width:  s.screenW,
		Height: s.screenH,
		Seed:   s.seed,
	}
}

// Render draws lens dirt effects based on visible light sources.
func (s *System) Render(screen *ebiten.Image) {
	if s.pattern == nil || len(s.pattern.Specks) == 0 {
		return
	}

	s.lightsMu.RLock()
	lights := s.lights
	s.lightsMu.RUnlock()

	if len(lights) == 0 {
		return
	}

	// Create/reuse overlay
	if s.overlay == nil || s.overlay.Bounds().Dx() != s.screenW || s.overlay.Bounds().Dy() != s.screenH {
		s.overlay = ebiten.NewImage(s.screenW, s.screenH)
	}
	s.overlay.Clear()

	// Render diffuse halos first (underneath specks)
	if s.preset.DiffuseHaloStrength > 0 {
		s.renderDiffuseHalos(lights)
	}

	// Render each speck based on nearby light intensity
	for _, speck := range s.pattern.Specks {
		s.renderSpeck(speck, lights)
	}

	// Apply edge darkening
	if s.preset.EdgeDarkening > 0 {
		s.applyEdgeDarkening()
	}

	// Composite overlay onto screen
	opts := &ebiten.DrawImageOptions{}
	opts.Blend = ebiten.BlendLighter // Additive blend for glow
	screen.DrawImage(s.overlay, opts)
}

// renderDiffuseHalos draws soft glow around each light source.
func (s *System) renderDiffuseHalos(lights []LightSource) {
	for _, light := range lights {
		if light.Intensity < 0.1 {
			continue
		}

		haloRadius := light.Radius * 2.0 * s.preset.DiffuseHaloStrength
		if haloRadius < 8 {
			continue
		}

		// Create or get cached halo sprite
		haloImg := s.getOrCreateHaloSprite(int(haloRadius), light.Color)

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(light.ScreenX-haloRadius, light.ScreenY-haloRadius)

		alpha := float32(light.Intensity * s.preset.DiffuseHaloStrength * 0.3)
		opts.ColorScale.ScaleAlpha(alpha)

		s.overlay.DrawImage(haloImg, opts)
	}
}

// renderSpeck draws a single dirt speck if illuminated.
func (s *System) renderSpeck(speck DirtSpeck, lights []LightSource) {
	// Calculate screen position
	screenX := speck.X * float64(s.screenW)
	screenY := speck.Y * float64(s.screenH)

	// Calculate total illumination from all lights
	var totalIllum float64
	var avgR, avgG, avgB float64
	var colorWeight float64

	for _, light := range lights {
		dx := screenX - light.ScreenX
		dy := screenY - light.ScreenY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist > s.preset.FalloffDistance {
			continue
		}

		// Inverse-square falloff
		falloff := 1.0 - (dist / s.preset.FalloffDistance)
		falloff = falloff * falloff // Quadratic falloff for softness

		illum := light.Intensity * falloff
		totalIllum += illum

		// Accumulate color contribution
		weight := illum
		avgR += float64(light.Color.R) / 255.0 * weight
		avgG += float64(light.Color.G) / 255.0 * weight
		avgB += float64(light.Color.B) / 255.0 * weight
		colorWeight += weight
	}

	if totalIllum < 0.05 {
		return
	}

	// Normalize color
	if colorWeight > 0 {
		avgR /= colorWeight
		avgG /= colorWeight
		avgB /= colorWeight
	}

	// Blend speck tint with light color
	finalR := uint8(clamp01(float64(speck.Tint.R)/255.0*0.6+avgR*0.4) * 255)
	finalG := uint8(clamp01(float64(speck.Tint.G)/255.0*0.6+avgG*0.4) * 255)
	finalB := uint8(clamp01(float64(speck.Tint.B)/255.0*0.6+avgB*0.4) * 255)

	// Get sprite for this speck
	sprite := s.getOrCreateSpeckSprite(speck.Shape, int(speck.Size), finalR, finalG, finalB, int(speck.Rotation*12/math.Pi))

	// Calculate final opacity
	opacity := speck.BaseOpacity * totalIllum * s.preset.IntensityScale
	if opacity > 0.8 {
		opacity = 0.8
	}

	opts := &ebiten.DrawImageOptions{}

	// Apply rotation around center
	hw := float64(sprite.Bounds().Dx()) / 2
	hh := float64(sprite.Bounds().Dy()) / 2
	opts.GeoM.Translate(-hw, -hh)
	opts.GeoM.Rotate(speck.Rotation)
	opts.GeoM.Translate(screenX, screenY)

	opts.ColorScale.ScaleAlpha(float32(opacity))

	s.overlay.DrawImage(sprite, opts)
}

// applyEdgeDarkening adds a subtle vignette to the dirt overlay.
func (s *System) applyEdgeDarkening() {
	if s.overlay == nil {
		return
	}

	centerX := float64(s.screenW) / 2
	centerY := float64(s.screenH) / 2
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	// We can't efficiently darken edges without reading pixels,
	// so we overlay a precomputed vignette instead
	vignetteImg := s.getOrCreateVignetteSprite()
	opts := &ebiten.DrawImageOptions{}
	opts.ColorScale.ScaleAlpha(float32(s.preset.EdgeDarkening))
	opts.Blend = ebiten.BlendSourceOver
	s.overlay.DrawImage(vignetteImg, opts)

	// Silence unused variable
	_ = maxDist
}

// getOrCreateSpeckSprite returns a cached or newly created speck sprite.
func (s *System) getOrCreateSpeckSprite(shape SpeckShape, size int, r, g, b uint8, rotStep int) *ebiten.Image {
	key := speckCacheKey{
		shape:   shape,
		size:    size,
		tintR:   r,
		tintG:   g,
		tintB:   b,
		rotStep: rotStep,
	}

	s.speckCacheMu.RLock()
	if cached, found := s.speckCache[key]; found {
		s.speckCacheMu.RUnlock()
		return cached
	}
	s.speckCacheMu.RUnlock()

	sprite := s.createSpeckSprite(shape, size, color.RGBA{R: r, G: g, B: b, A: 255})

	s.speckCacheMu.Lock()
	if len(s.speckCache) >= s.maxCache {
		// Evict one entry
		for k := range s.speckCache {
			delete(s.speckCache, k)
			break
		}
	}
	s.speckCache[key] = sprite
	s.speckCacheMu.Unlock()

	return sprite
}

// createSpeckSprite generates a dirt speck image.
func (s *System) createSpeckSprite(shape SpeckShape, size int, col color.RGBA) *ebiten.Image {
	if size < 2 {
		size = 2
	}

	// Different shapes have different aspect ratios
	var w, h int
	switch shape {
	case ShapeSmudge:
		w, h = size*2, size
	case ShapeStreaky:
		w, h = size/2+1, size*2
	default:
		w, h = size, size
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	centerX := float64(w) / 2
	centerY := float64(h) / 2

	switch shape {
	case ShapeCircle:
		s.drawCircleSpeck(img, centerX, centerY, float64(size)/2, col)
	case ShapeSmudge:
		s.drawSmudgeSpeck(img, centerX, centerY, float64(w)/2, float64(h)/2, col)
	case ShapeStreaky:
		s.drawStreakySpeck(img, centerX, centerY, float64(w)/2, float64(h)/2, col)
	case ShapeDiffuse:
		s.drawDiffuseSpeck(img, centerX, centerY, float64(size)/2, col)
	case ShapeHexagonal:
		s.drawHexagonalSpeck(img, centerX, centerY, float64(size)/2, col)
	default:
		s.drawCircleSpeck(img, centerX, centerY, float64(size)/2, col)
	}

	return ebiten.NewImageFromImage(img)
}

// drawCircleSpeck draws a simple dust particle.
func (s *System) drawCircleSpeck(img *image.RGBA, cx, cy, radius float64, col color.RGBA) {
	bounds := img.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > radius {
				continue
			}

			// Soft falloff from center
			alpha := 1.0 - (dist / radius)
			alpha = alpha * alpha // Quadratic falloff

			img.Set(x, y, color.RGBA{
				R: col.R,
				G: col.G,
				B: col.B,
				A: uint8(alpha * 255),
			})
		}
	}
}

// drawSmudgeSpeck draws an elongated fingerprint-like smear.
func (s *System) drawSmudgeSpeck(img *image.RGBA, cx, cy, radiusX, radiusY float64, col color.RGBA) {
	bounds := img.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dx := (float64(x) - cx) / radiusX
			dy := (float64(y) - cy) / radiusY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > 1 {
				continue
			}

			// Irregular edge using noise
			noise := math.Sin(float64(x)*0.5+float64(y)*0.3) * 0.15
			alpha := 1.0 - dist + noise
			if alpha < 0 {
				alpha = 0
			}
			alpha = alpha * alpha * 0.7 // Softer smudges

			img.Set(x, y, color.RGBA{
				R: col.R,
				G: col.G,
				B: col.B,
				A: uint8(clamp01(alpha) * 255),
			})
		}
	}
}

// drawStreakySpeck draws a rain-drop streak pattern.
func (s *System) drawStreakySpeck(img *image.RGBA, cx, cy, radiusX, radiusY float64, col color.RGBA) {
	bounds := img.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dx := (float64(x) - cx) / radiusX
			dy := (float64(y) - cy) / radiusY
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > 1 {
				continue
			}

			// Vertical streak falloff
			vertFade := 1.0 - math.Abs(dy)
			horizFade := 1.0 - math.Abs(dx)

			alpha := vertFade * horizFade * 0.6

			img.Set(x, y, color.RGBA{
				R: col.R,
				G: col.G,
				B: col.B,
				A: uint8(clamp01(alpha) * 255),
			})
		}
	}
}

// drawDiffuseSpeck draws a soft, diffused glow pattern.
func (s *System) drawDiffuseSpeck(img *image.RGBA, cx, cy, radius float64, col color.RGBA) {
	bounds := img.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > radius {
				continue
			}

			// Very soft gaussian-like falloff
			normalizedDist := dist / radius
			alpha := math.Exp(-normalizedDist * normalizedDist * 3)
			alpha *= 0.4 // Keep it subtle

			img.Set(x, y, color.RGBA{
				R: col.R,
				G: col.G,
				B: col.B,
				A: uint8(clamp01(alpha) * 255),
			})
		}
	}
}

// drawHexagonalSpeck draws a lens-flare hexagonal artifact.
func (s *System) drawHexagonalSpeck(img *image.RGBA, cx, cy, radius float64, col color.RGBA) {
	bounds := img.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy

			// Convert to hexagonal distance
			angle := math.Atan2(dy, dx)
			dist := math.Sqrt(dx*dx + dy*dy)

			// 6-fold symmetry
			hexAngle := math.Mod(angle+math.Pi, math.Pi/3) - math.Pi/6
			hexDist := dist / math.Cos(hexAngle)

			if hexDist > radius {
				continue
			}

			// Ring-like structure
			normalizedDist := hexDist / radius
			ringFade := math.Abs(normalizedDist - 0.7) // Peak at 70% radius
			alpha := math.Max(0, 1.0-ringFade*5) * 0.5

			img.Set(x, y, color.RGBA{
				R: col.R,
				G: col.G,
				B: col.B,
				A: uint8(clamp01(alpha) * 255),
			})
		}
	}
}

// getOrCreateHaloSprite returns a cached or newly created halo sprite.
func (s *System) getOrCreateHaloSprite(radius int, col color.RGBA) *ebiten.Image {
	// Simplified cache key for halos
	key := speckCacheKey{
		shape: ShapeDiffuse,
		size:  radius,
		tintR: col.R,
		tintG: col.G,
		tintB: col.B,
	}

	s.speckCacheMu.RLock()
	if cached, found := s.speckCache[key]; found {
		s.speckCacheMu.RUnlock()
		return cached
	}
	s.speckCacheMu.RUnlock()

	size := radius * 2
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := float64(radius)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - center
			dy := float64(y) - center
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > float64(radius) {
				continue
			}

			// Very soft gaussian falloff
			normalizedDist := dist / float64(radius)
			alpha := math.Exp(-normalizedDist * normalizedDist * 2)
			alpha *= 0.3 // Keep it subtle

			img.Set(x, y, color.RGBA{
				R: col.R,
				G: col.G,
				B: col.B,
				A: uint8(clamp01(alpha) * 255),
			})
		}
	}

	sprite := ebiten.NewImageFromImage(img)

	s.speckCacheMu.Lock()
	s.speckCache[key] = sprite
	s.speckCacheMu.Unlock()

	return sprite
}

// getOrCreateVignetteSprite returns a cached or newly created vignette.
func (s *System) getOrCreateVignetteSprite() *ebiten.Image {
	// Vignette is stored as a special entry
	key := speckCacheKey{
		shape: SpeckShape(99), // Special marker
		size:  s.screenW,
	}

	s.speckCacheMu.RLock()
	if cached, found := s.speckCache[key]; found {
		s.speckCacheMu.RUnlock()
		return cached
	}
	s.speckCacheMu.RUnlock()

	img := image.NewRGBA(image.Rect(0, 0, s.screenW, s.screenH))
	centerX := float64(s.screenW) / 2
	centerY := float64(s.screenH) / 2
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	for y := 0; y < s.screenH; y++ {
		for x := 0; x < s.screenW; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			normalizedDist := dist / maxDist
			// Only darken outer edges
			alpha := math.Pow(normalizedDist, 2) * 0.5

			img.Set(x, y, color.RGBA{
				R: 20,
				G: 15,
				B: 10,
				A: uint8(clamp01(alpha) * 255),
			})
		}
	}

	sprite := ebiten.NewImageFromImage(img)

	s.speckCacheMu.Lock()
	s.speckCache[key] = sprite
	s.speckCacheMu.Unlock()

	return sprite
}

// ClearCache removes all cached sprites.
func (s *System) ClearCache() {
	s.speckCacheMu.Lock()
	s.speckCache = make(map[speckCacheKey]*ebiten.Image)
	s.speckCacheMu.Unlock()
}

package emissive

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"reflect"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines emissive glow behavior per genre.
type GenrePreset struct {
	// FlameColor is the default flame/torch color.
	FlameColor color.RGBA
	// MagicColor is the default magic glow color.
	MagicColor color.RGBA
	// EyeColor is the default creature eye glow color.
	EyeColor color.RGBA
	// IntensityMult scales all glow intensities.
	IntensityMult float64
	// RadiusMult scales all glow radii.
	RadiusMult float64
	// FalloffBias adjusts edge softness (positive = softer).
	FalloffBias float64
	// FlickerIntensity controls flame flicker amount.
	FlickerIntensity float64
	// PulseIntensity controls magic pulse amount.
	PulseIntensity float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		FlameColor:       color.RGBA{R: 255, G: 180, B: 80, A: 255},
		MagicColor:       color.RGBA{R: 150, G: 100, B: 255, A: 255},
		EyeColor:         color.RGBA{R: 255, G: 200, B: 50, A: 255},
		IntensityMult:    1.0,
		RadiusMult:       1.0,
		FalloffBias:      0.0,
		FlickerIntensity: 0.25,
		PulseIntensity:   0.20,
	},
	"scifi": {
		FlameColor:       color.RGBA{R: 100, G: 200, B: 255, A: 255},
		MagicColor:       color.RGBA{R: 50, G: 255, B: 200, A: 255},
		EyeColor:         color.RGBA{R: 255, G: 100, B: 100, A: 255},
		IntensityMult:    0.9,
		RadiusMult:       0.85,
		FalloffBias:      0.3,
		FlickerIntensity: 0.05,
		PulseIntensity:   0.15,
	},
	"horror": {
		FlameColor:       color.RGBA{R: 200, G: 150, B: 100, A: 255},
		MagicColor:       color.RGBA{R: 100, G: 200, B: 100, A: 255},
		EyeColor:         color.RGBA{R: 200, G: 255, B: 200, A: 255},
		IntensityMult:    0.75,
		RadiusMult:       1.1,
		FalloffBias:      -0.2,
		FlickerIntensity: 0.35,
		PulseIntensity:   0.30,
	},
	"cyberpunk": {
		FlameColor:       color.RGBA{R: 255, G: 100, B: 200, A: 255},
		MagicColor:       color.RGBA{R: 0, G: 255, B: 255, A: 255},
		EyeColor:         color.RGBA{R: 255, G: 50, B: 150, A: 255},
		IntensityMult:    1.2,
		RadiusMult:       0.9,
		FalloffBias:      0.4,
		FlickerIntensity: 0.10,
		PulseIntensity:   0.25,
	},
	"postapoc": {
		FlameColor:       color.RGBA{R: 200, G: 100, B: 50, A: 255},
		MagicColor:       color.RGBA{R: 100, G: 255, B: 100, A: 255},
		EyeColor:         color.RGBA{R: 255, G: 150, B: 50, A: 255},
		IntensityMult:    0.7,
		RadiusMult:       0.8,
		FalloffBias:      -0.1,
		FlickerIntensity: 0.40,
		PulseIntensity:   0.15,
	},
}

// cacheKey uniquely identifies a cached glow sprite.
type cacheKey struct {
	colorR   uint8
	colorG   uint8
	colorB   uint8
	radius   int
	falloff  int // falloff * 10 as int
	glowType GlowType
}

// System renders emissive glow effects for marked entities.
type System struct {
	genreID  string
	preset   GenrePreset
	logger   *logrus.Entry
	rng      *rand.Rand
	tick     int
	screenW  int
	screenH  int
	cameraX  float64
	cameraY  float64
	cache    map[cacheKey]*ebiten.Image
	cacheMu  sync.RWMutex
	maxCache int
}

// NewSystem creates an emissive glow system.
func NewSystem(genreID string, seed int64) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	return &System{
		genreID:  genreID,
		preset:   preset,
		logger:   logrus.WithFields(logrus.Fields{"system_name": "emissive", "genre": genreID}),
		rng:      rand.New(rand.NewSource(seed)),
		cache:    make(map[cacheKey]*ebiten.Image),
		maxCache: 128,
		screenW:  320,
		screenH:  200,
	}
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
}

// SetScreenSize updates the screen dimensions for clipping.
func (s *System) SetScreenSize(w, h int) {
	s.screenW = w
	s.screenH = h
}

// SetCamera updates camera position for world-to-screen conversion.
func (s *System) SetCamera(x, y float64) {
	s.cameraX = x
	s.cameraY = y
}

// Update processes all emissive components and updates animation state.
func (s *System) Update(w *engine.World) {
	s.tick++
	deltaTime := 1.0 / 60.0

	compType := reflect.TypeOf(&Component{})
	posType := reflect.TypeOf(&engine.Position{})
	entities := w.Query(compType)

	for _, entity := range entities {
		compRaw, found := w.GetComponent(entity, compType)
		if !found {
			continue
		}
		comp, ok := compRaw.(*Component)
		if !ok || !comp.Enabled {
			continue
		}

		s.updateComponent(comp, deltaTime)

		if posRaw, found := w.GetComponent(entity, posType); found {
			if pos, ok := posRaw.(*engine.Position); ok {
				comp.ScreenX = pos.X - s.cameraX
				comp.ScreenY = pos.Y - s.cameraY
				dx := pos.X - s.cameraX
				dy := pos.Y - s.cameraY
				comp.Distance = math.Sqrt(dx*dx + dy*dy)
			}
		}
	}
}

// updateComponent animates pulse/flicker based on glow type.
func (s *System) updateComponent(comp *Component, deltaTime float64) {
	switch comp.GlowType {
	case TypeFlame:
		comp.PulsePhase += deltaTime * comp.PulseSpeed * 2 * math.Pi
		if comp.PulsePhase > 2*math.Pi {
			comp.PulsePhase -= 2 * math.Pi
		}
	case TypeMagic, TypeRadioactive:
		comp.PulsePhase += deltaTime * comp.PulseSpeed * 2 * math.Pi
		if comp.PulsePhase > 2*math.Pi {
			comp.PulsePhase -= 2 * math.Pi
		}
	case TypeElectric:
		comp.PulsePhase += deltaTime * comp.PulseSpeed * 4 * math.Pi
		if comp.PulsePhase > 2*math.Pi {
			comp.PulsePhase -= 2 * math.Pi
		}
	case TypeEye:
		comp.PulsePhase += deltaTime * comp.PulseSpeed * 2 * math.Pi
		if comp.PulsePhase > 2*math.Pi {
			comp.PulsePhase -= 2 * math.Pi
		}
	}
}

// Render draws all emissive glow effects to the screen.
func (s *System) Render(w *engine.World, screen *ebiten.Image) {
	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, entity := range entities {
		compRaw, found := w.GetComponent(entity, compType)
		if !found {
			continue
		}
		comp, ok := compRaw.(*Component)
		if !ok || !comp.Enabled {
			continue
		}

		s.renderGlow(screen, comp)
	}
}

// renderGlow draws a single glow effect.
func (s *System) renderGlow(screen *ebiten.Image, comp *Component) {
	effectiveIntensity := comp.Intensity * s.preset.IntensityMult
	effectiveRadius := comp.Radius * s.preset.RadiusMult

	distanceScale := 1.0
	if comp.Distance > 100 {
		distanceScale = 100 / comp.Distance
	}
	effectiveRadius *= distanceScale
	effectiveIntensity *= distanceScale

	if effectiveRadius < 4 || effectiveIntensity < 0.05 {
		return
	}

	if comp.ScreenX < -effectiveRadius || comp.ScreenX > float64(s.screenW)+effectiveRadius ||
		comp.ScreenY < -effectiveRadius || comp.ScreenY > float64(s.screenH)+effectiveRadius {
		return
	}

	modulatedIntensity := s.getModulatedIntensity(comp, effectiveIntensity)
	glowSprite := s.getOrCreateGlowSprite(comp, int(effectiveRadius))

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(
		comp.ScreenX-effectiveRadius,
		comp.ScreenY-effectiveRadius,
	)
	opts.ColorScale.ScaleAlpha(float32(modulatedIntensity))
	opts.Blend = ebiten.BlendLighter

	screen.DrawImage(glowSprite, opts)
}

// getModulatedIntensity applies flicker/pulse based on glow type.
func (s *System) getModulatedIntensity(comp *Component, baseIntensity float64) float64 {
	switch comp.GlowType {
	case TypeFlame:
		flicker := s.preset.FlickerIntensity
		noise := math.Sin(comp.PulsePhase) * 0.5
		noise += math.Sin(comp.PulsePhase*2.7) * 0.3
		noise += math.Sin(comp.PulsePhase*5.1) * 0.2
		return baseIntensity * (1.0 + noise*flicker)

	case TypeMagic, TypeRadioactive:
		pulse := s.preset.PulseIntensity
		mod := math.Sin(comp.PulsePhase) * 0.5
		return baseIntensity * (1.0 + mod*pulse)

	case TypeElectric:
		if s.rng.Float64() < 0.1 {
			return baseIntensity * (0.5 + s.rng.Float64())
		}
		return baseIntensity

	case TypeEye:
		pulse := 0.1
		mod := math.Sin(comp.PulsePhase) * 0.5
		return baseIntensity * (1.0 + mod*pulse)

	default:
		return baseIntensity
	}
}

// getOrCreateGlowSprite returns a cached or newly created glow sprite.
func (s *System) getOrCreateGlowSprite(comp *Component, radius int) *ebiten.Image {
	falloffInt := int(comp.FalloffPower * 10)
	key := cacheKey{
		colorR:   comp.Color.R,
		colorG:   comp.Color.G,
		colorB:   comp.Color.B,
		radius:   radius,
		falloff:  falloffInt,
		glowType: comp.GlowType,
	}

	s.cacheMu.RLock()
	if cached, found := s.cache[key]; found {
		s.cacheMu.RUnlock()
		return cached
	}
	s.cacheMu.RUnlock()

	sprite := s.createGlowSprite(comp, radius)

	s.cacheMu.Lock()
	if len(s.cache) >= s.maxCache {
		for k := range s.cache {
			delete(s.cache, k)
			break
		}
	}
	s.cache[key] = sprite
	s.cacheMu.Unlock()

	return sprite
}

// createGlowSprite generates a radial glow image.
func (s *System) createGlowSprite(comp *Component, radius int) *ebiten.Image {
	size := radius * 2
	if size < 4 {
		size = 4
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	centerX := float64(size) / 2
	centerY := float64(size) / 2

	falloffPower := comp.FalloffPower + s.preset.FalloffBias
	if falloffPower < 0.5 {
		falloffPower = 0.5
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)
			normalizedDist := dist / float64(radius)

			if normalizedDist > 1.0 {
				continue
			}

			falloff := math.Pow(1.0-normalizedDist, falloffPower)
			brightness := falloff * comp.CoreBrightness

			if normalizedDist < 0.3 {
				coreBoost := 1.0 + (0.3-normalizedDist)*0.5
				brightness *= coreBoost
			}

			if brightness > 1.0 {
				brightness = 1.0
			}

			alpha := uint8(brightness * 255)
			r := uint8(float64(comp.Color.R) * brightness)
			g := uint8(float64(comp.Color.G) * brightness)
			b := uint8(float64(comp.Color.B) * brightness)

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: alpha})
		}
	}

	return ebiten.NewImageFromImage(img)
}

// ClearCache removes all cached glow sprites.
func (s *System) ClearCache() {
	s.cacheMu.Lock()
	s.cache = make(map[cacheKey]*ebiten.Image)
	s.cacheMu.Unlock()
}

// GetPreset returns the current genre preset.
func (s *System) GetPreset() GenrePreset {
	return s.preset
}

package specsparkle

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

// GenrePreset defines sparkle appearance parameters for each genre.
type GenrePreset struct {
	// BaseIntensity scales all sparkle brightness [0.0-1.0].
	BaseIntensity float64
	// SpawnRate is sparkles per second per unit density.
	SpawnRate float64
	// LifetimeMin is minimum sparkle duration in seconds.
	LifetimeMin float64
	// LifetimeMax is maximum sparkle duration in seconds.
	LifetimeMax float64
	// SizeVariation adds random size variation [0.0-1.0].
	SizeVariation float64
	// ColorShiftAmount adds hue variation to sparkle color.
	ColorShiftAmount float64
	// MetalTint is the base color tint for metal sparkles.
	MetalTint color.RGBA
	// CrystalTint is the base color tint for crystal sparkles.
	CrystalTint color.RGBA
	// WetTint is the base color tint for wet surface sparkles.
	WetTint color.RGBA
	// StarShape uses star-shaped sparkles instead of circular.
	StarShape bool
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BaseIntensity:    0.85,
		SpawnRate:        2.5,
		LifetimeMin:      0.15,
		LifetimeMax:      0.4,
		SizeVariation:    0.3,
		ColorShiftAmount: 0.1,
		MetalTint:        color.RGBA{R: 255, G: 245, B: 220, A: 255}, // Warm gold
		CrystalTint:      color.RGBA{R: 220, G: 240, B: 255, A: 255}, // Cool blue
		WetTint:          color.RGBA{R: 255, G: 255, B: 255, A: 255}, // Pure white
		StarShape:        true,
	},
	"scifi": {
		BaseIntensity:    1.0,
		SpawnRate:        3.5,
		LifetimeMin:      0.1,
		LifetimeMax:      0.25,
		SizeVariation:    0.2,
		ColorShiftAmount: 0.05,
		MetalTint:        color.RGBA{R: 200, G: 220, B: 255, A: 255}, // Cool chrome
		CrystalTint:      color.RGBA{R: 100, G: 255, B: 200, A: 255}, // Tech green
		WetTint:          color.RGBA{R: 180, G: 220, B: 255, A: 255}, // Blue-white
		StarShape:        false,
	},
	"horror": {
		BaseIntensity:    0.5,
		SpawnRate:        1.0,
		LifetimeMin:      0.3,
		LifetimeMax:      0.6,
		SizeVariation:    0.5,
		ColorShiftAmount: 0.15,
		MetalTint:        color.RGBA{R: 200, G: 180, B: 160, A: 255}, // Tarnished
		CrystalTint:      color.RGBA{R: 180, G: 255, B: 180, A: 255}, // Sickly green
		WetTint:          color.RGBA{R: 200, G: 220, B: 200, A: 255}, // Murky
		StarShape:        true,
	},
	"cyberpunk": {
		BaseIntensity:    1.2,
		SpawnRate:        4.0,
		LifetimeMin:      0.08,
		LifetimeMax:      0.2,
		SizeVariation:    0.25,
		ColorShiftAmount: 0.2,
		MetalTint:        color.RGBA{R: 255, G: 100, B: 255, A: 255}, // Neon pink
		CrystalTint:      color.RGBA{R: 0, G: 255, B: 255, A: 255},   // Cyan
		WetTint:          color.RGBA{R: 255, G: 200, B: 255, A: 255}, // Pink-white
		StarShape:        false,
	},
	"postapoc": {
		BaseIntensity:    0.6,
		SpawnRate:        1.5,
		LifetimeMin:      0.2,
		LifetimeMax:      0.5,
		SizeVariation:    0.4,
		ColorShiftAmount: 0.1,
		MetalTint:        color.RGBA{R: 200, G: 170, B: 140, A: 255}, // Rusty
		CrystalTint:      color.RGBA{R: 200, G: 220, B: 180, A: 255}, // Dirty
		WetTint:          color.RGBA{R: 220, G: 200, B: 180, A: 255}, // Muddy
		StarShape:        true,
	},
}

// sparkleKey uniquely identifies a cached sparkle texture.
type sparkleKey struct {
	colorR    uint8
	colorG    uint8
	colorB    uint8
	size      int
	starShape bool
}

// System manages animated specular sparkle effects.
type System struct {
	mu       sync.RWMutex
	genreID  string
	preset   GenrePreset
	logger   *logrus.Entry
	rng      *rand.Rand
	seed     int64
	tick     int
	screenW  int
	screenH  int
	cameraX  float64
	cameraY  float64
	enabled  bool
	sparkles map[engine.Entity][]Sparkle
	cache    map[sparkleKey]*ebiten.Image
	maxCache int
}

// NewSystem creates an animated specular sparkle system.
func NewSystem(genreID string, seed int64) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	return &System{
		genreID:  genreID,
		preset:   preset,
		seed:     seed,
		rng:      rand.New(rand.NewSource(seed)),
		logger:   logrus.WithFields(logrus.Fields{"system_name": "specsparkle", "genre": genreID}),
		enabled:  true,
		screenW:  320,
		screenH:  200,
		sparkles: make(map[engine.Entity][]Sparkle),
		cache:    make(map[sparkleKey]*ebiten.Image),
		maxCache: 64,
	}
}

// SetGenre updates the system for a new genre.
func (s *System) SetGenre(genreID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}
	s.genreID = genreID
	s.preset = preset
	s.logger = s.logger.WithField("genre", genreID)
}

// SetScreenSize updates screen dimensions.
func (s *System) SetScreenSize(w, h int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.screenW = w
	s.screenH = h
}

// SetCamera updates camera position for world-to-screen conversion.
func (s *System) SetCamera(x, y float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cameraX = x
	s.cameraY = y
}

// SetEnabled toggles sparkle rendering globally.
func (s *System) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// Update processes all sparkle components and animates sparkles.
func (s *System) Update(w *engine.World) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return
	}

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

		// Try to get position from engine.Position component
		if posRaw, posFound := w.GetComponent(entity, posType); posFound {
			if pos, posOk := posRaw.(*engine.Position); posOk {
				dx := pos.X - s.cameraX
				dy := pos.Y - s.cameraY
				comp.Distance = math.Sqrt(dx*dx + dy*dy)
				// Convert to simple screen offset (proper conversion would use raycaster)
				// For now, use a simple projection approximation
				if comp.Distance > 0.1 {
					screenCenterX := float64(s.screenW) / 2.0
					screenCenterY := float64(s.screenH) / 2.0
					// Simple orthographic-like projection for nearby effects
					comp.ScreenX = screenCenterX + dx*20.0
					comp.ScreenY = screenCenterY + dy*20.0
				}
			}
		}

		s.updateEntitySparkles(entity, comp, deltaTime)
	}

	s.cleanupInactiveSparkles()
}

// updateEntitySparkles updates sparkle animation for a single entity.
func (s *System) updateEntitySparkles(entity engine.Entity, comp *Component, deltaTime float64) {
	sparkles := s.sparkles[entity]

	for i := range sparkles {
		if !sparkles[i].Active {
			continue
		}

		sparkles[i].Phase += sparkles[i].Speed * deltaTime
		if sparkles[i].Phase >= 1.0 {
			sparkles[i].Active = false
		}
	}

	if s.shouldSpawnSparkle(comp, deltaTime) {
		newSparkle := s.createSparkle(comp)
		found := false
		for i := range sparkles {
			if !sparkles[i].Active {
				sparkles[i] = newSparkle
				found = true
				break
			}
		}
		if !found && len(sparkles) < 32 {
			sparkles = append(sparkles, newSparkle)
		}
	}

	s.sparkles[entity] = sparkles
}

// shouldSpawnSparkle determines if a new sparkle should spawn this frame.
func (s *System) shouldSpawnSparkle(comp *Component, deltaTime float64) bool {
	spawnChance := s.preset.SpawnRate * comp.Density * deltaTime
	distanceFalloff := 1.0
	if comp.Distance > 5.0 {
		distanceFalloff = 5.0 / comp.Distance
	}
	return s.rng.Float64() < spawnChance*distanceFalloff
}

// createSparkle generates a new sparkle with randomized parameters.
func (s *System) createSparkle(comp *Component) Sparkle {
	lifetime := s.preset.LifetimeMin + s.rng.Float64()*(s.preset.LifetimeMax-s.preset.LifetimeMin)
	speed := 1.0 / lifetime

	sizeMult := 1.0 - s.preset.SizeVariation/2 + s.rng.Float64()*s.preset.SizeVariation

	baseColor := s.getMaterialColor(comp.Material)

	hueShift := (s.rng.Float64() - 0.5) * 2 * s.preset.ColorShiftAmount
	sparkleColor := s.shiftHue(baseColor, hueShift)

	return Sparkle{
		X:        s.rng.Float64(),
		Y:        s.rng.Float64(),
		Phase:    0.0,
		Speed:    speed,
		SizeMult: sizeMult,
		Color:    sparkleColor,
		Active:   true,
	}
}

// getMaterialColor returns the base sparkle color for a material type.
func (s *System) getMaterialColor(mat MaterialClass) color.RGBA {
	switch mat {
	case MaterialMetal:
		return s.preset.MetalTint
	case MaterialCrystal:
		return s.preset.CrystalTint
	case MaterialWet, MaterialGlass:
		return s.preset.WetTint
	case MaterialGold:
		return color.RGBA{R: 255, G: 220, B: 150, A: 255}
	case MaterialSilver:
		return color.RGBA{R: 220, G: 230, B: 255, A: 255}
	default:
		return s.preset.MetalTint
	}
}

// shiftHue applies a hue shift to a color.
func (s *System) shiftHue(c color.RGBA, shift float64) color.RGBA {
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255

	cMax := math.Max(r, math.Max(g, b))
	cMin := math.Min(r, math.Min(g, b))
	delta := cMax - cMin

	var h float64
	if delta < 0.001 {
		h = 0
	} else if cMax == r {
		h = math.Mod((g-b)/delta, 6.0) / 6.0
	} else if cMax == g {
		h = ((b-r)/delta + 2.0) / 6.0
	} else {
		h = ((r-g)/delta + 4.0) / 6.0
	}

	sat := 0.0
	if cMax > 0.001 {
		sat = delta / cMax
	}
	val := cMax

	h += shift
	for h < 0 {
		h += 1.0
	}
	for h >= 1.0 {
		h -= 1.0
	}

	return s.hsvToRGB(h, sat, val)
}

// hsvToRGB converts HSV to RGB color.
func (s *System) hsvToRGB(h, sat, val float64) color.RGBA {
	if sat < 0.001 {
		v := uint8(val * 255)
		return color.RGBA{R: v, G: v, B: v, A: 255}
	}

	h *= 6.0
	i := math.Floor(h)
	f := h - i
	p := val * (1 - sat)
	q := val * (1 - sat*f)
	t := val * (1 - sat*(1-f))

	var r, g, b float64
	switch int(i) % 6 {
	case 0:
		r, g, b = val, t, p
	case 1:
		r, g, b = q, val, p
	case 2:
		r, g, b = p, val, t
	case 3:
		r, g, b = p, q, val
	case 4:
		r, g, b = t, p, val
	case 5:
		r, g, b = val, p, q
	}

	return color.RGBA{R: uint8(r * 255), G: uint8(g * 255), B: uint8(b * 255), A: 255}
}

// cleanupInactiveSparkles removes dead sparkle lists.
func (s *System) cleanupInactiveSparkles() {
	for entity, sparkles := range s.sparkles {
		hasActive := false
		for _, sp := range sparkles {
			if sp.Active {
				hasActive = true
				break
			}
		}
		if !hasActive && len(sparkles) > 8 {
			delete(s.sparkles, entity)
		}
	}
}

// Draw renders all active sparkles to the target image.
func (s *System) Draw(target *ebiten.Image, w *engine.World) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return
	}

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

		sparkles := s.sparkles[entity]
		for _, sp := range sparkles {
			if !sp.Active {
				continue
			}
			s.drawSparkle(target, comp, &sp)
		}
	}
}

// drawSparkle renders a single sparkle to the target.
func (s *System) drawSparkle(target *ebiten.Image, comp *Component, sp *Sparkle) {
	brightness := s.getSparkleIntensity(sp.Phase)
	if brightness < 0.01 {
		return
	}

	screenX := comp.ScreenX + sp.X*comp.Width
	screenY := comp.ScreenY + sp.Y*comp.Height

	if screenX < -10 || screenX > float64(s.screenW)+10 ||
		screenY < -10 || screenY > float64(s.screenH)+10 {
		return
	}

	distanceFalloff := 1.0
	if comp.Distance > 3.0 {
		distanceFalloff = 3.0 / comp.Distance
		if distanceFalloff < 0.2 {
			distanceFalloff = 0.2
		}
	}

	finalBrightness := brightness * s.preset.BaseIntensity * comp.Intensity * distanceFalloff
	size := int(math.Ceil(comp.Size * sp.SizeMult * distanceFalloff))
	if size < 1 {
		size = 1
	}
	if size > 12 {
		size = 12
	}

	sparkleImg := s.getSparkleImage(sp.Color, size, s.preset.StarShape)
	if sparkleImg == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX-float64(size)/2, screenY-float64(size)/2)
	op.ColorScale.Scale(float32(finalBrightness), float32(finalBrightness), float32(finalBrightness), float32(finalBrightness))
	op.Blend = ebiten.BlendLighter

	target.DrawImage(sparkleImg, op)
}

// getSparkleIntensity calculates brightness based on animation phase.
func (s *System) getSparkleIntensity(phase float64) float64 {
	if phase < 0.5 {
		return math.Sin(phase * math.Pi)
	}
	return math.Sin(phase * math.Pi)
}

// getSparkleImage returns a cached or generated sparkle texture.
func (s *System) getSparkleImage(col color.RGBA, size int, starShape bool) *ebiten.Image {
	key := sparkleKey{
		colorR:    col.R,
		colorG:    col.G,
		colorB:    col.B,
		size:      size,
		starShape: starShape,
	}

	if img, ok := s.cache[key]; ok {
		return img
	}

	if len(s.cache) >= s.maxCache {
		for k := range s.cache {
			delete(s.cache, k)
			break
		}
	}

	img := s.generateSparkleImage(col, size, starShape)
	s.cache[key] = img
	return img
}

// generateSparkleImage creates a sparkle texture.
func (s *System) generateSparkleImage(col color.RGBA, size int, starShape bool) *ebiten.Image {
	if size < 1 {
		size = 1
	}

	rgba := image.NewRGBA(image.Rect(0, 0, size, size))
	center := float64(size) / 2.0

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) + 0.5 - center
			dy := float64(y) + 0.5 - center
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := center

			var alpha float64
			if starShape && size >= 3 {
				angle := math.Atan2(dy, dx)
				rays := 4.0
				rayFactor := math.Abs(math.Sin(angle * rays))
				adjustedDist := dist * (1.0 + (1.0-rayFactor)*0.5)
				alpha = 1.0 - adjustedDist/maxDist
			} else {
				alpha = 1.0 - dist/maxDist
			}

			if alpha > 0 {
				alpha = alpha * alpha
				if alpha > 1.0 {
					alpha = 1.0
				}

				coreIntensity := 1.0
				if dist < center*0.3 {
					coreIntensity = 1.5
				}

				r := uint8(clampByte(int(float64(col.R) * coreIntensity)))
				g := uint8(clampByte(int(float64(col.G) * coreIntensity)))
				b := uint8(clampByte(int(float64(col.B) * coreIntensity)))
				a := uint8(alpha * 255)

				rgba.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: a})
			}
		}
	}

	img := ebiten.NewImage(size, size)
	img.WritePixels(rgba.Pix)
	return img
}

// clampByte restricts an int to byte range.
func clampByte(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// AddSparkleRegion adds a sparkle effect to a screen region (not attached to entity).
func (s *System) AddSparkleRegion(x, y, width, height float64, mat MaterialClass) engine.Entity {
	return 0
}

// GetPreset returns the current genre preset for external configuration.
func (s *System) GetPreset() GenrePreset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.preset
}

package muzzleflash

import (
	"image/color"
	"math"
	"math/rand"
	"reflect"

	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages muzzle flash spawning, updating, and cleanup.
type System struct {
	genreID string
	rng     *rand.Rand
	logger  *logrus.Entry

	// Genre-specific flash style modifiers
	colorTint     color.RGBA
	intensityMult float64
	durationMult  float64
	emitLightMult float64
}

// NewSystem creates a muzzle flash system.
func NewSystem(genreID string, seed int64) *System {
	s := &System{
		genreID:       genreID,
		rng:           rand.New(rand.NewSource(seed)),
		intensityMult: 1.0,
		durationMult:  1.0,
		emitLightMult: 1.0,
		colorTint:     color.RGBA{R: 255, G: 255, B: 255, A: 255},
		logger: logrus.WithFields(logrus.Fields{
			"system": "muzzleflash",
			"genre":  genreID,
		}),
	}
	s.applyGenreStyle(genreID)
	return s
}

// SetGenre updates the genre-specific rendering style.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenreStyle(genreID)
	s.logger = s.logger.WithField("genre", genreID)
}

// applyGenreStyle sets visual parameters based on genre.
func (s *System) applyGenreStyle(genreID string) {
	switch genreID {
	case "fantasy":
		// Warmer, fire-like flashes
		s.colorTint = color.RGBA{R: 255, G: 240, B: 200, A: 255}
		s.intensityMult = 1.0
		s.durationMult = 1.1
		s.emitLightMult = 1.2

	case "scifi":
		// Cool blue/white energy flashes
		s.colorTint = color.RGBA{R: 200, G: 230, B: 255, A: 255}
		s.intensityMult = 1.2
		s.durationMult = 0.9
		s.emitLightMult = 1.5

	case "horror":
		// Muted, flickering flashes
		s.colorTint = color.RGBA{R: 255, G: 220, B: 180, A: 255}
		s.intensityMult = 0.8
		s.durationMult = 1.0
		s.emitLightMult = 0.7

	case "cyberpunk":
		// Neon-tinted, bright flashes
		s.colorTint = color.RGBA{R: 255, G: 180, B: 255, A: 255}
		s.intensityMult = 1.3
		s.durationMult = 0.8
		s.emitLightMult = 1.8

	case "postapoc":
		// Dirty, orange/brown flashes
		s.colorTint = color.RGBA{R: 255, G: 200, B: 150, A: 255}
		s.intensityMult = 0.9
		s.durationMult = 1.0
		s.emitLightMult = 1.0

	default:
		s.colorTint = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		s.intensityMult = 1.0
		s.durationMult = 1.0
		s.emitLightMult = 1.0
	}
}

// Update processes all muzzle flash components, aging and removing expired flashes.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime

	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, compType)
		if !ok {
			continue
		}

		mf := comp.(*Component)
		s.updateFlashes(mf, deltaTime)
	}
}

// updateFlashes ages all flashes and removes expired ones.
func (s *System) updateFlashes(mf *Component, deltaTime float64) {
	// Filter in place to avoid allocation
	n := 0
	for _, flash := range mf.ActiveFlashes {
		flash.Age += deltaTime
		if flash.Age < flash.Duration {
			mf.ActiveFlashes[n] = flash
			n++
		}
	}
	mf.ActiveFlashes = mf.ActiveFlashes[:n]
}

// SpawnFlash creates a new muzzle flash at the specified position.
func (s *System) SpawnFlash(w *engine.World, entity engine.Entity, x, y, angle float64, flashType string, intensity float64) {
	comp := s.getOrCreateComponent(w, entity)

	// Enforce max flashes limit
	if len(comp.ActiveFlashes) >= comp.MaxFlashes {
		// Remove oldest
		comp.ActiveFlashes = comp.ActiveFlashes[1:]
	}

	profile := GetProfile(flashType)

	// Apply genre modifiers
	duration := profile.Duration * s.durationMult
	finalIntensity := intensity * s.intensityMult

	// Generate slight random variation for organic feel
	angleVariation := (s.rng.Float64() - 0.5) * 0.1 // ±0.05 radians
	sizeVariation := 0.9 + s.rng.Float64()*0.2      // 0.9-1.1x

	// Tint colors by genre
	primaryColor := s.tintColor(profile.PrimaryColor)
	secondaryColor := s.tintColor(profile.SecondaryColor)

	flash := &Flash{
		X:              x,
		Y:              y,
		Angle:          angle + angleVariation,
		Age:            0,
		Duration:       duration,
		FlashType:      flashType,
		Intensity:      finalIntensity,
		PrimaryColor:   primaryColor,
		SecondaryColor: secondaryColor,
		Scale:          sizeVariation,
		EmitsLight:     profile.EmitsLight,
		LightIntensity: profile.LightIntensity * s.emitLightMult * finalIntensity,
		LightRadius:    profile.LightRadius,
	}

	comp.ActiveFlashes = append(comp.ActiveFlashes, flash)

	s.logger.WithFields(logrus.Fields{
		"entity":     entity,
		"flash_type": flashType,
		"x":          x,
		"y":          y,
		"intensity":  finalIntensity,
	}).Debug("spawned muzzle flash")
}

// tintColor applies genre color tint to a flash color.
func (s *System) tintColor(c color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8((int(c.R) * int(s.colorTint.R)) / 255),
		G: uint8((int(c.G) * int(s.colorTint.G)) / 255),
		B: uint8((int(c.B) * int(s.colorTint.B)) / 255),
		A: c.A,
	}
}

// getOrCreateComponent gets or creates a muzzle flash component for an entity.
func (s *System) getOrCreateComponent(w *engine.World, entity engine.Entity) *Component {
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if ok {
		return comp.(*Component)
	}

	newComp := NewComponent()
	w.AddComponent(entity, newComp)
	return newComp
}

// GetActiveFlashes returns all active flashes for an entity.
func (s *System) GetActiveFlashes(w *engine.World, entity engine.Entity) []*Flash {
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if !ok {
		return nil
	}
	return comp.(*Component).ActiveFlashes
}

// GetAllActiveFlashes returns all active flashes across all entities.
func (s *System) GetAllActiveFlashes(w *engine.World) []*Flash {
	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	var allFlashes []*Flash
	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, compType)
		if !ok {
			continue
		}
		mf := comp.(*Component)
		allFlashes = append(allFlashes, mf.ActiveFlashes...)
	}
	return allFlashes
}

// CollectLightSources returns light data for all active flashes (for lighting system integration).
func (s *System) CollectLightSources(w *engine.World) []LightSource {
	flashes := s.GetAllActiveFlashes(w)
	lights := make([]LightSource, 0, len(flashes))

	for _, flash := range flashes {
		if !flash.EmitsLight {
			continue
		}

		// Flash light fades over duration
		progress := flash.Age / flash.Duration
		fadeOut := 1.0 - progress*progress // Quadratic fadeout

		if fadeOut <= 0 {
			continue
		}

		lights = append(lights, LightSource{
			X:         flash.X,
			Y:         flash.Y,
			Radius:    flash.LightRadius,
			Intensity: flash.LightIntensity * fadeOut,
			Color:     flash.PrimaryColor,
		})
	}

	return lights
}

// LightSource represents light emitted by a muzzle flash.
type LightSource struct {
	X, Y      float64
	Radius    float64
	Intensity float64
	Color     color.RGBA
}

// FlashRenderData contains pre-computed rendering data for a flash.
type FlashRenderData struct {
	ScreenX, ScreenY float64
	Scale            float64
	Progress         float64 // 0-1, how far through the flash duration
	Alpha            float64 // Current opacity
	CoreRadius       float64
	OuterRadius      float64
	RayLength        float64
	RayAngles        []float64 // Pre-computed ray angles
	PrimaryColor     color.RGBA
	SecondaryColor   color.RGBA
	Profile          FlashProfile
}

// PrepareRenderData computes rendering data for a flash.
func (s *System) PrepareRenderData(flash *Flash, cameraX, cameraY float64, screenWidth, screenHeight int) *FlashRenderData {
	profile := GetProfile(flash.FlashType)

	// Calculate screen position
	dx := flash.X - cameraX
	dy := flash.Y - cameraY

	// Raycaster uses tile-based coordinates, convert to screen
	tileSize := 10.0 // pixels per tile unit
	screenX := float64(screenWidth)/2 + dx*tileSize
	screenY := float64(screenHeight)/2 + dy*tileSize

	// Progress through flash duration
	progress := flash.Age / flash.Duration
	if progress > 1.0 {
		progress = 1.0
	}

	// Flash opacity: bright start, quick fadeout
	alpha := 1.0 - progress*progress

	// Size: starts small, expands, then contracts
	sizeProgress := 1.0
	if progress < 0.3 {
		// Expand phase
		sizeProgress = progress / 0.3
	} else {
		// Contract phase
		sizeProgress = 1.0 - (progress-0.3)/0.7
	}
	sizeProgress = math.Max(0.3, sizeProgress)

	baseSize := profile.BaseSize * flash.Scale * flash.Intensity * sizeProgress
	coreRadius := baseSize * 0.3
	outerRadius := baseSize * 0.8

	// Compute ray angles
	var rayAngles []float64
	if profile.RayCount > 0 {
		rayAngles = make([]float64, profile.RayCount)
		for i := 0; i < profile.RayCount; i++ {
			baseAngle := flash.Angle + float64(i)*2*math.Pi/float64(profile.RayCount)
			if profile.RandomRays {
				// Add seeded random variation
				baseAngle += (s.rng.Float64() - 0.5) * 0.3
			}
			rayAngles[i] = baseAngle
		}
	}

	rayLength := baseSize * 1.5

	return &FlashRenderData{
		ScreenX:        screenX,
		ScreenY:        screenY,
		Scale:          flash.Scale,
		Progress:       progress,
		Alpha:          alpha,
		CoreRadius:     coreRadius,
		OuterRadius:    outerRadius,
		RayLength:      rayLength,
		RayAngles:      rayAngles,
		PrimaryColor:   flash.PrimaryColor,
		SecondaryColor: flash.SecondaryColor,
		Profile:        profile,
	}
}

// Type returns the system identifier.
func (s *System) Type() string {
	return "MuzzleFlashSystem"
}

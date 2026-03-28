// Package dustmote provides ambient floating particle effects for atmospheric realism.
// Dust motes drift lazily in the air, catching light beams to create a lived-in feel.
// Unlike weather particles which fall from above, dust motes float and swirl in place
// with Brownian motion, visible primarily when illuminated by nearby light sources.
package dustmote

import (
	"image/color"
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/sirupsen/logrus"
)

// MoteType represents different ambient particle types per genre.
type MoteType int

const (
	MoteTypeDust   MoteType = iota // Standard dust motes (fantasy, default)
	MoteTypeSpore                  // Fungal spores with slight glow (horror)
	MoteTypePollen                 // Golden pollen grains (fantasy outdoor)
	MoteTypeAsh                    // Fine ash particles (postapoc)
	MoteTypeDigit                  // Digital glitch particles (cyberpunk)
	MoteTypeClean                  // Minimal particles (scifi)
)

// GenrePreset defines mote behavior per genre.
type GenrePreset struct {
	MoteType       MoteType
	Density        float64 // Particles per unit area
	BaseSize       float64 // Average particle size in pixels
	SizeVariation  float64 // Size variation range
	Brightness     float64 // Base brightness (0-1)
	DriftSpeed     float64 // Horizontal drift speed
	TurbulenceAmp  float64 // Brownian motion amplitude
	LightResponse  float64 // How strongly particles respond to light (0-1)
	TwinkleRate    float64 // How often particles catch light
	ColorR         uint8
	ColorG         uint8
	ColorB         uint8
	GlowRadius     float64 // For spores/magic particles
	TrailLength    int     // For digital particles
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		MoteType:      MoteTypeDust,
		Density:       0.8,
		BaseSize:      1.5,
		SizeVariation: 0.8,
		Brightness:    0.5,
		DriftSpeed:    3.0,
		TurbulenceAmp: 2.5,
		LightResponse: 0.9,
		TwinkleRate:   0.15,
		ColorR:        200,
		ColorG:        190,
		ColorB:        170, // Warm dust color
		GlowRadius:    0,
		TrailLength:   0,
	},
	"scifi": {
		MoteType:      MoteTypeClean,
		Density:       0.15, // Clean environments
		BaseSize:      0.8,
		SizeVariation: 0.3,
		Brightness:    0.3,
		DriftSpeed:    1.5,
		TurbulenceAmp: 1.0,
		LightResponse: 0.6,
		TwinkleRate:   0.05,
		ColorR:        180,
		ColorG:        200,
		ColorB:        220, // Cool blue tint
		GlowRadius:    0,
		TrailLength:   0,
	},
	"horror": {
		MoteType:      MoteTypeSpore,
		Density:       0.6,
		BaseSize:      2.0,
		SizeVariation: 1.2,
		Brightness:    0.4,
		DriftSpeed:    2.0,
		TurbulenceAmp: 3.5, // More chaotic movement
		LightResponse: 0.7,
		TwinkleRate:   0.08,
		ColorR:        140,
		ColorG:        160,
		ColorB:        130, // Sickly greenish
		GlowRadius:    1.5,
		TrailLength:   0,
	},
	"cyberpunk": {
		MoteType:      MoteTypeDigit,
		Density:       0.4,
		BaseSize:      1.2,
		SizeVariation: 0.5,
		Brightness:    0.7,
		DriftSpeed:    5.0,
		TurbulenceAmp: 4.0,
		LightResponse: 0.5, // Less light-dependent
		TwinkleRate:   0.3, // More frequent flickers
		ColorR:        0,
		ColorG:        255,
		ColorB:        200, // Neon cyan
		GlowRadius:    2.0,
		TrailLength:   3,
	},
	"postapoc": {
		MoteType:      MoteTypeAsh,
		Density:       1.0, // Heavy particle density
		BaseSize:      1.8,
		SizeVariation: 1.0,
		Brightness:    0.35,
		DriftSpeed:    4.0,
		TurbulenceAmp: 3.0,
		LightResponse: 0.8,
		TwinkleRate:   0.1,
		ColorR:        120,
		ColorG:        110,
		ColorB:        100, // Gray ash
		GlowRadius:    0,
		TrailLength:   0,
	},
}

// Mote represents a single floating ambient particle.
type Mote struct {
	X, Y          float64 // World position
	VX, VY        float64 // Current velocity
	Phase         float64 // Brownian motion phase offset
	Size          float64 // Particle size
	BaseAlpha     uint8   // Base transparency
	TwinklePhase  float64 // Light catching animation phase
	LifePhase     float64 // For respawn cycling
	Active        bool
	LightLevel    float64 // Current illumination (0-1)
	Trail         []point // For digital particles
}

type point struct {
	X, Y float64
}

// System manages ambient floating dust mote particles.
type System struct {
	genreID string
	preset  GenrePreset
	logger  *logrus.Entry
	rng     *rand.Rand

	motes     []Mote
	maxMotes  int
	spawnArea float64 // World units around camera to spawn motes

	// Camera/view state
	cameraX, cameraY float64
	viewWidth        float64
	viewHeight       float64

	// Light sources for particle illumination
	lights     []LightSource
	lightMu    sync.RWMutex
	ambientLit float64 // Base ambient light level (0-1)

	// Animation timing
	time float64

	// Reusable overlay for rendering
	overlay   *ebiten.Image
	overlayMu sync.Mutex

	// Screen dimensions
	screenW, screenH int
}

// LightSource represents a point light that illuminates dust motes.
type LightSource struct {
	X, Y      float64 // World position
	Radius    float64 // Light falloff radius
	Intensity float64 // Light brightness (0-1)
	ColorR    uint8   // Tint
	ColorG    uint8
	ColorB    uint8
}

// NewSystem creates an ambient dust mote system.
func NewSystem(genreID string, seed int64, screenW, screenH int) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	// Calculate max motes based on screen size and density
	maxMotes := int(float64(screenW*screenH) * preset.Density / 500.0)
	if maxMotes < 50 {
		maxMotes = 50
	}
	if maxMotes > 500 {
		maxMotes = 500
	}

	sys := &System{
		genreID:    genreID,
		preset:     preset,
		logger:     logrus.WithField("system", "dustmote"),
		rng:        rand.New(rand.NewSource(seed)),
		motes:      make([]Mote, maxMotes),
		maxMotes:   maxMotes,
		spawnArea:  30.0, // World units around camera
		lights:     make([]LightSource, 0, 16),
		ambientLit: 0.1,
		screenW:    screenW,
		screenH:    screenH,
	}

	// Initialize all motes as inactive
	for i := range sys.motes {
		sys.motes[i].Active = false
		sys.motes[i].Trail = make([]point, 0, preset.TrailLength)
	}

	sys.logger.WithFields(logrus.Fields{
		"genre":    genreID,
		"maxMotes": maxMotes,
		"moteType": preset.MoteType,
	}).Debug("Dust mote system initialized")

	return sys
}

// SetGenre changes the genre and applies new presets.
func (s *System) SetGenre(genreID string) {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	s.genreID = genreID
	s.preset = preset

	// Clear existing motes for fresh spawn with new settings
	for i := range s.motes {
		s.motes[i].Active = false
	}

	s.logger.WithField("genre", genreID).Debug("Genre changed for dust mote system")
}

// SetCamera updates the camera position for mote spawning and culling.
func (s *System) SetCamera(x, y, viewWidth, viewHeight float64) {
	s.cameraX = x
	s.cameraY = y
	s.viewWidth = viewWidth
	s.viewHeight = viewHeight
}

// SetLights updates the light source list for illumination calculations.
func (s *System) SetLights(lights []LightSource) {
	s.lightMu.Lock()
	defer s.lightMu.Unlock()
	s.lights = lights
}

// AddLight adds a single light source.
func (s *System) AddLight(light LightSource) {
	s.lightMu.Lock()
	defer s.lightMu.Unlock()
	s.lights = append(s.lights, light)
}

// ClearLights removes all light sources.
func (s *System) ClearLights() {
	s.lightMu.Lock()
	defer s.lightMu.Unlock()
	s.lights = s.lights[:0]
}

// SetAmbientLight sets the base ambient illumination level.
func (s *System) SetAmbientLight(level float64) {
	if level < 0 {
		level = 0
	}
	if level > 1 {
		level = 1
	}
	s.ambientLit = level
}

// Update advances the dust mote simulation.
func (s *System) Update(deltaTime float64) {
	s.time += deltaTime

	// Spawn new motes to fill quota
	s.spawnMotes()

	// Update existing motes
	for i := range s.motes {
		m := &s.motes[i]
		if !m.Active {
			continue
		}

		// Update Brownian motion (turbulent drift)
		turbX := math.Sin(s.time*1.5+m.Phase) * s.preset.TurbulenceAmp
		turbY := math.Cos(s.time*1.2+m.Phase*1.3) * s.preset.TurbulenceAmp * 0.7

		// Add gentle drift
		driftX := math.Sin(s.time*0.3+m.Phase*0.5) * s.preset.DriftSpeed
		driftY := math.Cos(s.time*0.2+m.Phase*0.7) * s.preset.DriftSpeed * 0.5

		// Apply velocity with damping
		m.VX = m.VX*0.95 + (turbX+driftX)*0.05
		m.VY = m.VY*0.95 + (turbY+driftY)*0.05

		m.X += m.VX * deltaTime
		m.Y += m.VY * deltaTime

		// Update twinkle phase
		m.TwinklePhase += deltaTime * 3.0
		if m.TwinklePhase > math.Pi*2 {
			m.TwinklePhase -= math.Pi * 2
		}

		// Update life phase for respawn cycling
		m.LifePhase += deltaTime * 0.1
		if m.LifePhase > 1.0 {
			m.LifePhase = 0
		}

		// Calculate illumination from nearby lights
		m.LightLevel = s.calculateIllumination(m.X, m.Y)

		// Cull motes that drift too far from camera
		dx := m.X - s.cameraX
		dy := m.Y - s.cameraY
		if dx*dx+dy*dy > s.spawnArea*s.spawnArea*1.5 {
			m.Active = false
		}

		// Update trail for digital particles
		if s.preset.TrailLength > 0 && m.Active {
			m.Trail = append(m.Trail, point{m.X, m.Y})
			if len(m.Trail) > s.preset.TrailLength {
				m.Trail = m.Trail[1:]
			}
		}
	}
}

// spawnMotes fills the mote pool around the camera.
func (s *System) spawnMotes() {
	activeCount := 0
	for i := range s.motes {
		if s.motes[i].Active {
			activeCount++
		}
	}

	// Target density based on preset
	targetCount := int(float64(s.maxMotes) * s.preset.Density)
	if activeCount >= targetCount {
		return
	}

	// Spawn up to a few motes per frame to avoid hitching
	spawnCount := targetCount - activeCount
	if spawnCount > 5 {
		spawnCount = 5
	}

	for i := 0; i < spawnCount; i++ {
		s.spawnSingleMote()
	}
}

// spawnSingleMote creates one new mote near the camera.
func (s *System) spawnSingleMote() {
	// Find inactive slot
	var m *Mote
	for i := range s.motes {
		if !s.motes[i].Active {
			m = &s.motes[i]
			break
		}
	}
	if m == nil {
		return
	}

	// Spawn in random position around camera
	angle := s.rng.Float64() * math.Pi * 2
	dist := s.rng.Float64() * s.spawnArea
	m.X = s.cameraX + math.Cos(angle)*dist
	m.Y = s.cameraY + math.Sin(angle)*dist

	// Randomize properties
	m.Phase = s.rng.Float64() * math.Pi * 2
	m.Size = s.preset.BaseSize + (s.rng.Float64()-0.5)*2*s.preset.SizeVariation
	if m.Size < 0.5 {
		m.Size = 0.5
	}
	m.BaseAlpha = uint8(s.preset.Brightness * 255 * (0.7 + s.rng.Float64()*0.3))
	m.TwinklePhase = s.rng.Float64() * math.Pi * 2
	m.LifePhase = s.rng.Float64()
	m.VX = 0
	m.VY = 0
	m.LightLevel = 0
	m.Active = true

	// Reset trail for digital particles
	if s.preset.TrailLength > 0 {
		m.Trail = m.Trail[:0]
	}
}

// calculateIllumination computes how lit a position is based on nearby lights.
func (s *System) calculateIllumination(x, y float64) float64 {
	illumination := s.ambientLit

	s.lightMu.RLock()
	defer s.lightMu.RUnlock()

	for _, light := range s.lights {
		dx := x - light.X
		dy := y - light.Y
		distSq := dx*dx + dy*dy
		radiusSq := light.Radius * light.Radius

		if distSq < radiusSq {
			// Inverse-square falloff with soft edge
			falloff := 1.0 - math.Sqrt(distSq)/light.Radius
			falloff = falloff * falloff // Quadratic falloff
			illumination += light.Intensity * falloff
		}
	}

	if illumination > 1.0 {
		illumination = 1.0
	}
	return illumination
}

// Draw renders all visible dust motes to the screen.
func (s *System) Draw(screen *ebiten.Image) {
	halfW := float64(s.screenW) / 2
	halfH := float64(s.screenH) / 2
	worldScale := 10.0 // World units to screen pixels

	for i := range s.motes {
		m := &s.motes[i]
		if !m.Active {
			continue
		}

		// Calculate effective visibility
		visibility := m.LightLevel * s.preset.LightResponse
		if visibility < 0.05 {
			continue // Too dark to see
		}

		// Calculate twinkle effect (catching light)
		twinkleFactor := 1.0
		if s.rng.Float64() < s.preset.TwinkleRate {
			twinkleFactor = 1.0 + math.Sin(m.TwinklePhase)*0.8
		}

		// Calculate final alpha
		alpha := float64(m.BaseAlpha) * visibility * twinkleFactor
		if alpha > 255 {
			alpha = 255
		}
		if alpha < 5 {
			continue
		}

		// Project to screen space
		dx := m.X - s.cameraX
		dy := m.Y - s.cameraY
		screenX := float32(halfW + dx*worldScale)
		screenY := float32(halfH + dy*worldScale)

		// Cull off-screen
		margin := float32(20)
		if screenX < -margin || screenX > float32(s.screenW)+margin ||
			screenY < -margin || screenY > float32(s.screenH)+margin {
			continue
		}

		// Render based on mote type
		s.renderMote(screen, m, screenX, screenY, uint8(alpha))
	}
}

// renderMote draws a single mote based on its type.
func (s *System) renderMote(screen *ebiten.Image, m *Mote, x, y float32, alpha uint8) {
	size := float32(m.Size)
	preset := &s.preset

	// Apply light tinting
	s.lightMu.RLock()
	tintR, tintG, tintB := s.calculateLightTint(m.X, m.Y)
	s.lightMu.RUnlock()

	// Blend base color with light tint
	r := uint8((float64(preset.ColorR)*0.6 + float64(tintR)*0.4))
	g := uint8((float64(preset.ColorG)*0.6 + float64(tintG)*0.4))
	b := uint8((float64(preset.ColorB)*0.6 + float64(tintB)*0.4))

	c := color.RGBA{R: r, G: g, B: b, A: alpha}

	switch preset.MoteType {
	case MoteTypeDust, MoteTypeAsh, MoteTypeClean:
		// Simple filled circle for standard dust
		s.drawSoftCircle(screen, x, y, size, c)

	case MoteTypeSpore:
		// Glowing spore with soft halo
		s.drawSoftCircle(screen, x, y, size, c)
		if preset.GlowRadius > 0 {
			glowC := color.RGBA{R: r, G: g, B: b, A: uint8(float64(alpha) * 0.3)}
			s.drawSoftCircle(screen, x, y, size+float32(preset.GlowRadius), glowC)
		}

	case MoteTypePollen:
		// Golden pollen with warm glow
		s.drawSoftCircle(screen, x, y, size, c)
		warmC := color.RGBA{R: 255, G: 220, B: 100, A: uint8(float64(alpha) * 0.2)}
		s.drawSoftCircle(screen, x, y, size*1.5, warmC)

	case MoteTypeDigit:
		// Digital particle with trail
		s.drawDigitalMote(screen, m, x, y, c)
	}
}

// drawSoftCircle renders a particle as a soft-edged circle.
func (s *System) drawSoftCircle(screen *ebiten.Image, x, y, radius float32, c color.RGBA) {
	if radius < 1 {
		// Single pixel for tiny particles
		vector.DrawFilledRect(screen, x, y, 1, 1, c, false)
		return
	}

	// Draw filled circle with soft edge via multiple passes
	// Core (brightest)
	coreRadius := radius * 0.6
	vector.DrawFilledCircle(screen, x, y, coreRadius, c, false)

	// Middle ring (semi-transparent)
	midC := color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(float64(c.A) * 0.5)}
	vector.DrawFilledCircle(screen, x, y, radius*0.8, midC, false)

	// Outer halo (very transparent)
	outerC := color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(float64(c.A) * 0.2)}
	vector.DrawFilledCircle(screen, x, y, radius, outerC, false)
}

// drawDigitalMote renders a digital glitch particle with trail.
func (s *System) drawDigitalMote(screen *ebiten.Image, m *Mote, x, y float32, c color.RGBA) {
	worldScale := 10.0
	halfW := float64(s.screenW) / 2
	halfH := float64(s.screenH) / 2

	// Draw trail
	if len(m.Trail) > 1 {
		for i := 0; i < len(m.Trail)-1; i++ {
			t := float64(i) / float64(len(m.Trail))
			trailAlpha := uint8(float64(c.A) * t * 0.5)
			trailC := color.RGBA{R: c.R, G: c.G, B: c.B, A: trailAlpha}

			dx := m.Trail[i].X - s.cameraX
			dy := m.Trail[i].Y - s.cameraY
			tx := float32(halfW + dx*worldScale)
			ty := float32(halfH + dy*worldScale)

			vector.DrawFilledRect(screen, tx, ty, float32(m.Size)*0.5, float32(m.Size)*0.5, trailC, false)
		}
	}

	// Draw main particle as small rectangle (digital feel)
	vector.DrawFilledRect(screen, x-float32(m.Size)*0.5, y-float32(m.Size)*0.5, float32(m.Size), float32(m.Size), c, false)

	// Occasional glitch flicker - offset copy
	if s.rng.Float64() < 0.1 {
		offset := float32(s.rng.Float64()*4 - 2)
		glitchC := color.RGBA{R: 255, G: c.G, B: c.B, A: uint8(float64(c.A) * 0.5)}
		vector.DrawFilledRect(screen, x-float32(m.Size)*0.5+offset, y-float32(m.Size)*0.5, float32(m.Size), float32(m.Size), glitchC, false)
	}
}

// calculateLightTint returns the blended tint color from nearby lights.
func (s *System) calculateLightTint(x, y float64) (r, g, b uint8) {
	// Default warm ambient tint
	totalR, totalG, totalB := 220.0, 200.0, 180.0
	totalWeight := 0.1

	for _, light := range s.lights {
		dx := x - light.X
		dy := y - light.Y
		distSq := dx*dx + dy*dy
		radiusSq := light.Radius * light.Radius

		if distSq < radiusSq {
			falloff := 1.0 - math.Sqrt(distSq)/light.Radius
			weight := falloff * falloff * light.Intensity

			totalR += float64(light.ColorR) * weight
			totalG += float64(light.ColorG) * weight
			totalB += float64(light.ColorB) * weight
			totalWeight += weight
		}
	}

	r = uint8(totalR / totalWeight)
	g = uint8(totalG / totalWeight)
	b = uint8(totalB / totalWeight)
	return
}

// GetActiveCount returns the number of active motes.
func (s *System) GetActiveCount() int {
	count := 0
	for i := range s.motes {
		if s.motes[i].Active {
			count++
		}
	}
	return count
}

// Clear removes all active motes.
func (s *System) Clear() {
	for i := range s.motes {
		s.motes[i].Active = false
	}
}

// Type returns the component type identifier for ECS.
func (s *System) Type() string {
	return "DustMoteSystem"
}

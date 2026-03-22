package volumetric

import (
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines volumetric lighting behavior per genre.
type GenrePreset struct {
	// BaseDustDensity is the default dust density for the atmosphere
	BaseDustDensity float64

	// ScatterColorR/G/B is the atmospheric dust tint
	ScatterColorR, ScatterColorG, ScatterColorB float64

	// RayIntensity scales overall volumetric effect strength
	RayIntensity float64

	// SampleCount controls ray marching quality (more = better but slower)
	SampleCount int

	// FalloffExponent controls ray distance fade
	FalloffExponent float64

	// NoiseStrength adds procedural variation to rays
	NoiseStrength float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BaseDustDensity: 0.4,
		ScatterColorR:   1.0,
		ScatterColorG:   0.92,
		ScatterColorB:   0.75, // Warm golden dust
		RayIntensity:    0.6,
		SampleCount:     16,
		FalloffExponent: 2.0,
		NoiseStrength:   0.15,
	},
	"scifi": {
		BaseDustDensity: 0.15,
		ScatterColorR:   0.85,
		ScatterColorG:   0.95,
		ScatterColorB:   1.0, // Cool blue tint
		RayIntensity:    0.4,
		SampleCount:     12,
		FalloffExponent: 2.5,
		NoiseStrength:   0.05,
	},
	"horror": {
		BaseDustDensity: 0.55,
		ScatterColorR:   0.9,
		ScatterColorG:   0.9,
		ScatterColorB:   0.85, // Sickly pale
		RayIntensity:    0.7,
		SampleCount:     20,
		FalloffExponent: 1.5,
		NoiseStrength:   0.25,
	},
	"cyberpunk": {
		BaseDustDensity: 0.3,
		ScatterColorR:   0.95,
		ScatterColorG:   0.85,
		ScatterColorB:   1.0, // Slight magenta pollution
		RayIntensity:    0.5,
		SampleCount:     14,
		FalloffExponent: 2.2,
		NoiseStrength:   0.1,
	},
	"postapoc": {
		BaseDustDensity: 0.6,
		ScatterColorR:   1.0,
		ScatterColorG:   0.88,
		ScatterColorB:   0.65, // Dusty orange/yellow
		RayIntensity:    0.65,
		SampleCount:     18,
		FalloffExponent: 1.8,
		NoiseStrength:   0.3,
	},
}

// System manages volumetric light rendering.
type System struct {
	genreID string
	preset  GenrePreset
	logger  *logrus.Entry

	screenW, screenH int

	// Reusable overlay image
	overlay   *ebiten.Image
	overlayMu sync.Mutex

	// Ray cache for performance
	rayCache   []raySegment
	cacheValid bool

	// Noise lookup table for performance
	noiseTable [256]float64
}

// raySegment represents a cached ray for rendering.
type raySegment struct {
	screenX, screenY int
	intensity        float64
	r, g, b          float64
}

// NewSystem creates a volumetric lighting system.
func NewSystem(genreID string, screenW, screenH int) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	s := &System{
		genreID:  genreID,
		preset:   preset,
		screenW:  screenW,
		screenH:  screenH,
		overlay:  ebiten.NewImage(screenW, screenH),
		rayCache: make([]raySegment, 0, screenW*screenH/16),
		logger: logrus.WithFields(logrus.Fields{
			"system": "volumetric",
			"genre":  genreID,
		}),
	}

	// Initialize noise table for procedural variation
	s.initNoiseTable()

	return s
}

// initNoiseTable creates a deterministic noise lookup table.
func (s *System) initNoiseTable() {
	// Simple deterministic pseudo-random values
	seed := uint64(0xDEADBEEF)
	for i := range s.noiseTable {
		seed = seed*6364136223846793005 + 1442695040888963407
		s.noiseTable[i] = float64(seed&0xFFFF) / 65535.0
	}
}

// SetGenre updates the volumetric preset.
func (s *System) SetGenre(genreID string) {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}
	s.genreID = genreID
	s.preset = preset
	s.cacheValid = false
	s.logger = logrus.WithFields(logrus.Fields{
		"system": "volumetric",
		"genre":  genreID,
	})
}

// SetScreenSize updates the render target dimensions.
func (s *System) SetScreenSize(w, h int) {
	if w == s.screenW && h == s.screenH {
		return
	}
	s.screenW = w
	s.screenH = h
	s.overlayMu.Lock()
	s.overlay = ebiten.NewImage(w, h)
	s.overlayMu.Unlock()
	s.cacheValid = false
}

// Render draws volumetric light effects for the given light sources.
// cameraX, cameraY is the camera world position.
// dirX, dirY is the camera look direction (normalized).
// fov is the field of view in radians.
// lights contains all active light sources to render volumetric effects for.
// occluder is a function that returns true if a world position is blocked.
func (s *System) Render(
	screen *ebiten.Image,
	cameraX, cameraY float64,
	dirX, dirY float64,
	fov float64,
	lights []LightShaft,
	occluder func(worldX, worldY float64) bool,
) {
	if len(lights) == 0 {
		return
	}

	s.overlayMu.Lock()
	defer s.overlayMu.Unlock()

	// Clear overlay
	s.overlay.Clear()

	// Perpendicular vector for screen-space projection
	perpX := -dirY
	perpY := dirX

	// Half FOV tangent for screen mapping
	halfFOVTan := math.Tan(fov / 2)
	halfW := float64(s.screenW) / 2
	halfH := float64(s.screenH) / 2

	// Process each light
	for _, light := range lights {
		s.renderLightShaft(
			light,
			cameraX, cameraY,
			dirX, dirY,
			perpX, perpY,
			halfFOVTan, halfW, halfH,
			occluder,
		)
	}

	// Blend overlay onto screen with additive blending
	opts := &ebiten.DrawImageOptions{}
	opts.CompositeMode = ebiten.CompositeModeLighter
	screen.DrawImage(s.overlay, opts)
}

// renderLightShaft renders a single volumetric light source.
func (s *System) renderLightShaft(
	light LightShaft,
	cameraX, cameraY float64,
	dirX, dirY float64,
	perpX, perpY float64,
	halfFOVTan, halfW, halfH float64,
	occluder func(worldX, worldY float64) bool,
) {
	// Vector from camera to light
	toLightX := light.X - cameraX
	toLightY := light.Y - cameraY

	// Distance to light
	distToLight := math.Sqrt(toLightX*toLightX + toLightY*toLightY)
	if distToLight > light.Radius*3 {
		return // Too far for visible volumetrics
	}

	// Check if light is in front of camera
	dot := toLightX*dirX + toLightY*dirY
	if dot < 0.1 {
		return // Behind camera
	}

	// Project light center to screen
	screenDist := dot
	perpDot := toLightX*perpX + toLightY*perpY
	screenX := int(halfW + (perpDot/screenDist/halfFOVTan)*halfW)
	screenY := int(halfH) // Light sources are at eye level for simplicity

	// Calculate screen-space radius based on distance
	worldRadius := light.Radius
	screenRadius := int((worldRadius / screenDist / halfFOVTan) * halfW)
	if screenRadius < 10 {
		screenRadius = 10
	}
	if screenRadius > s.screenW {
		screenRadius = s.screenW
	}

	// Ray march parameters
	sampleCount := s.preset.SampleCount
	dustDensity := light.DustDensity * s.preset.BaseDustDensity
	intensity := light.Intensity * s.preset.RayIntensity

	// Render radial rays from light center
	rayCount := screenRadius / 2
	if rayCount < 8 {
		rayCount = 8
	}
	if rayCount > 64 {
		rayCount = 64
	}

	for ray := 0; ray < rayCount; ray++ {
		// Angle around light center
		angle := float64(ray) * 2 * math.Pi / float64(rayCount)

		// Add noise variation to angle
		noiseIdx := (ray * 17) & 0xFF
		angleNoise := (s.noiseTable[noiseIdx] - 0.5) * s.preset.NoiseStrength
		angle += angleNoise

		cosA := math.Cos(angle)
		sinA := math.Sin(angle)

		// March along ray
		for sample := 1; sample <= sampleCount; sample++ {
			// Distance along ray (0 to screenRadius)
			t := float64(sample) / float64(sampleCount)
			rayDist := t * float64(screenRadius)

			// Screen position
			sx := screenX + int(cosA*rayDist)
			sy := screenY + int(sinA*rayDist)

			// Bounds check
			if sx < 0 || sx >= s.screenW || sy < 0 || sy >= s.screenH {
				continue
			}

			// Convert screen position back to approximate world position for occlusion
			// Interpolate between camera and light position along the ray
			worldX := cameraX + (light.X-cameraX)*t
			worldY := cameraY + (light.Y-cameraY)*t

			// Check occlusion if available
			if occluder != nil && occluder(worldX, worldY) {
				continue // Ray blocked by wall
			}

			// Calculate intensity falloff
			falloff := 1.0 - math.Pow(t, s.preset.FalloffExponent)
			if falloff < 0 {
				falloff = 0
			}

			// Distance from camera falloff
			distFalloff := 1.0 / (1.0 + distToLight*0.1)

			// Dust scattering intensity
			scatter := dustDensity * falloff * intensity * distFalloff

			// Add ray noise variation
			noiseIdx2 := (sx*7 + sy*13) & 0xFF
			scatter *= 0.7 + s.noiseTable[noiseIdx2]*0.6

			// Clamp scatter
			if scatter > 1.0 {
				scatter = 1.0
			}
			if scatter < 0.01 {
				continue
			}

			// Calculate final color (light color * scatter color * intensity)
			r := light.R * s.preset.ScatterColorR * scatter
			g := light.G * s.preset.ScatterColorG * scatter
			b := light.B * s.preset.ScatterColorB * scatter

			// Draw pixel with additive blend
			s.setPixelAdditive(sx, sy, r, g, b, scatter)
		}
	}
}

// setPixelAdditive adds color to a pixel in the overlay.
func (s *System) setPixelAdditive(x, y int, r, g, b, a float64) {
	// Clamp values
	r = clampF64(r, 0, 1)
	g = clampF64(g, 0, 1)
	b = clampF64(b, 0, 1)
	a = clampF64(a, 0, 1)

	col := color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: uint8(a * 255),
	}

	// Draw a soft 3x3 kernel for smoother results
	s.overlay.Set(x, y, col)

	// Softer surrounding pixels
	softCol := color.RGBA{
		R: uint8(r * 127),
		G: uint8(g * 127),
		B: uint8(b * 127),
		A: uint8(a * 127),
	}
	if x > 0 {
		s.overlay.Set(x-1, y, softCol)
	}
	if x < s.screenW-1 {
		s.overlay.Set(x+1, y, softCol)
	}
	if y > 0 {
		s.overlay.Set(x, y-1, softCol)
	}
	if y < s.screenH-1 {
		s.overlay.Set(x, y+1, softCol)
	}
}

// RenderSimple draws volumetric effects without occlusion checking.
// Use this for performance when wall occlusion isn't critical.
func (s *System) RenderSimple(
	screen *ebiten.Image,
	cameraX, cameraY float64,
	dirX, dirY float64,
	fov float64,
	lights []LightShaft,
) {
	s.Render(screen, cameraX, cameraY, dirX, dirY, fov, lights, nil)
}

// CreateLightShaftFromTorch creates a volumetric light configured for a torch.
func (s *System) CreateLightShaftFromTorch(x, y, intensity float64) LightShaft {
	return LightShaft{
		X:               x,
		Y:               y,
		Intensity:       intensity * 0.8,
		Radius:          4.0,
		R:               1.0,
		G:               0.7,
		B:               0.3,
		DustDensity:     s.preset.BaseDustDensity * 1.2,
		FalloffExponent: s.preset.FalloffExponent,
	}
}

// CreateLightShaftFromMagic creates a volumetric light for magical sources.
func (s *System) CreateLightShaftFromMagic(x, y, intensity, r, g, b float64) LightShaft {
	return LightShaft{
		X:               x,
		Y:               y,
		Intensity:       intensity,
		Radius:          5.0,
		R:               r,
		G:               g,
		B:               b,
		DustDensity:     s.preset.BaseDustDensity * 0.8,
		FalloffExponent: s.preset.FalloffExponent * 0.8,
	}
}

// GetPreset returns the current genre preset for external configuration.
func (s *System) GetPreset() GenrePreset {
	return s.preset
}

package camerafx

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// System manages camera effects for visual feedback.
type System struct {
	genre        string
	rng          *rand.Rand
	shakeBase    float64
	flashBase    float64
	chromaBase   float64
	vignetteBase float64
	logger       *logrus.Entry
	component    *Component
}

// NewSystem creates a camera effects system.
func NewSystem(genreID string, seed int64) *System {
	s := &System{
		genre: genreID,
		rng:   rand.New(rand.NewSource(seed)),
		logger: logrus.WithFields(logrus.Fields{
			"system": "camerafx",
			"genre":  genreID,
		}),
		component: &Component{
			ZoomCurrent: 1.0,
			ZoomTarget:  1.0,
			ZoomSpeed:   2.0,
		},
	}
	s.applyGenrePreset(genreID)
	return s
}

// applyGenrePreset configures effect intensities based on genre.
func (s *System) applyGenrePreset(genreID string) {
	switch genreID {
	case "fantasy":
		s.shakeBase = 1.0
		s.flashBase = 0.8
		s.chromaBase = 0.3
		s.vignetteBase = 0.15
	case "scifi":
		s.shakeBase = 0.7
		s.flashBase = 1.0
		s.chromaBase = 0.6
		s.vignetteBase = 0.1
	case "horror":
		s.shakeBase = 1.3
		s.flashBase = 0.5
		s.chromaBase = 0.5
		s.vignetteBase = 0.35
	case "cyberpunk":
		s.shakeBase = 0.9
		s.flashBase = 1.2
		s.chromaBase = 0.8
		s.vignetteBase = 0.2
	case "postapoc":
		s.shakeBase = 1.1
		s.flashBase = 0.6
		s.chromaBase = 0.4
		s.vignetteBase = 0.25
	default:
		s.shakeBase = 1.0
		s.flashBase = 0.8
		s.chromaBase = 0.3
		s.vignetteBase = 0.15
	}
}

// SetGenre updates effect parameters for a new genre.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
	s.applyGenrePreset(genreID)
}

// Update processes camera effects each frame.
func (s *System) Update(deltaTime float64) {
	c := s.component

	// Update shake
	if c.ShakeIntensity > 0 {
		angle := s.rng.Float64() * math.Pi * 2.0
		magnitude := c.ShakeIntensity
		c.ShakeOffsetX = math.Cos(angle) * magnitude
		c.ShakeOffsetY = math.Sin(angle) * magnitude
		c.ShakeIntensity -= c.ShakeDecay * deltaTime
		if c.ShakeIntensity < 0.01 {
			c.ShakeIntensity = 0
			c.ShakeOffsetX = 0
			c.ShakeOffsetY = 0
		}
	}

	// Update flash
	if c.FlashAlpha > 0 {
		c.FlashAlpha -= c.FlashDecay * deltaTime
		if c.FlashAlpha < 0 {
			c.FlashAlpha = 0
		}
	}

	// Update zoom
	if math.Abs(c.ZoomCurrent-c.ZoomTarget) > 0.001 {
		diff := c.ZoomTarget - c.ZoomCurrent
		c.ZoomCurrent += diff * c.ZoomSpeed * deltaTime
	} else {
		c.ZoomCurrent = c.ZoomTarget
	}

	// Update chromatic aberration
	if c.ChromaticAberr > 0 {
		c.ChromaticAberr -= c.ChromaticDecay * deltaTime
		if c.ChromaticAberr < 0 {
			c.ChromaticAberr = 0
		}
	}
}

// TriggerShake adds screen shake with intensity scaled by genre.
func (s *System) TriggerShake(intensity float64) {
	c := s.component
	scaledIntensity := intensity * s.shakeBase
	if scaledIntensity > c.ShakeIntensity {
		c.ShakeIntensity = scaledIntensity
		c.ShakeDecay = scaledIntensity * 4.0
	}
}

// TriggerFlash adds a screen flash with color.
func (s *System) TriggerFlash(r, g, b, alpha float64) {
	c := s.component
	scaledAlpha := alpha * s.flashBase
	if scaledAlpha > c.FlashAlpha {
		c.FlashAlpha = scaledAlpha
		c.FlashR = r
		c.FlashG = g
		c.FlashB = b
		c.FlashDecay = scaledAlpha * 3.0
	}
}

// TriggerChromatic adds chromatic aberration effect.
func (s *System) TriggerChromatic(intensity float64) {
	c := s.component
	scaledIntensity := intensity * s.chromaBase
	if scaledIntensity > c.ChromaticAberr {
		c.ChromaticAberr = scaledIntensity
		c.ChromaticDecay = scaledIntensity * 2.0
	}
}

// SetZoom sets target camera zoom level.
func (s *System) SetZoom(zoom float64) {
	s.component.ZoomTarget = clamp(zoom, 0.5, 2.0)
}

// GetComponent returns the current camera effects component.
func (s *System) GetComponent() *Component {
	return s.component
}

// GetShakeOffset returns current shake displacement.
func (s *System) GetShakeOffset() (float64, float64) {
	return s.component.ShakeOffsetX, s.component.ShakeOffsetY
}

// GetFlashColor returns current flash color and alpha.
func (s *System) GetFlashColor() (r, g, b, a float64) {
	c := s.component
	return c.FlashR, c.FlashG, c.FlashB, c.FlashAlpha
}

// GetZoom returns current zoom level.
func (s *System) GetZoom() float64 {
	return s.component.ZoomCurrent
}

// GetChromaticAberration returns current aberration intensity.
func (s *System) GetChromaticAberration() float64 {
	return s.component.ChromaticAberr
}

// GetVignette returns base vignette strength for genre.
func (s *System) GetVignette() float64 {
	return s.vignetteBase
}

// ShakePresets provides common shake intensities.
type ShakePresets struct{}

// Shake provides convenient access to common camera shake intensity presets.
var Shake = ShakePresets{}

// Tiny returns a minimal shake intensity (0.5) for subtle feedback.
func (ShakePresets) Tiny() float64 { return 0.5 }

// Light returns a light shake intensity (1.5) for minor impacts.
func (ShakePresets) Light() float64 { return 1.5 }

// Medium returns a medium shake intensity (3.0) for standard hits.
func (ShakePresets) Medium() float64 { return 3.0 }

// Heavy returns a heavy shake intensity (6.0) for powerful impacts.
func (ShakePresets) Heavy() float64 { return 6.0 }

// Massive returns a massive shake intensity (12.0) for explosions.
func (ShakePresets) Massive() float64 { return 12.0 }

// Cataclysmic returns an extreme shake intensity (20.0) for catastrophic events.
func (ShakePresets) Cataclysmic() float64 { return 20.0 }

// FlashPresets provides common flash colors.
type FlashPresets struct{}

// Flash provides convenient access to common screen flash color presets.
var Flash = FlashPresets{}

// White returns a white flash (RGBA values) for bright impacts.
func (FlashPresets) White() (r, g, b, a float64) { return 1.0, 1.0, 1.0, 0.6 }

// Red returns a red flash (RGBA values) for damage taken.
func (FlashPresets) Red() (r, g, b, a float64) { return 1.0, 0.2, 0.2, 0.5 }

// Orange returns an orange flash (RGBA values) for fire effects.
func (FlashPresets) Orange() (r, g, b, a float64) { return 1.0, 0.6, 0.2, 0.5 }

// Blue returns a blue flash (RGBA values) for energy effects.
func (FlashPresets) Blue() (r, g, b, a float64) { return 0.2, 0.5, 1.0, 0.5 }

// Green returns a green flash (RGBA values) for healing effects.
func (FlashPresets) Green() (r, g, b, a float64) { return 0.3, 1.0, 0.3, 0.4 }

// Purple returns a purple flash (RGBA values) for magic effects.
func (FlashPresets) Purple() (r, g, b, a float64) { return 0.8, 0.3, 1.0, 0.5 }

// Yellow returns a yellow flash (RGBA values) for electric effects.
func (FlashPresets) Yellow() (r, g, b, a float64) { return 1.0, 1.0, 0.3, 0.5 }

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

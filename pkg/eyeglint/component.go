package eyeglint

import "image/color"

// Component stores eye glint rendering parameters for a sprite.
// This component tracks detected eye positions and glint animation state.
type Component struct {
	// EyePositions stores detected eye center coordinates [(x1,y1), (x2,y2), ...]
	EyePositions [][2]int

	// EyeSizes stores the radius of each detected eye
	EyeSizes []int

	// GlintPhase tracks animation phase for subtle glint movement [0.0-2π]
	GlintPhase float64

	// GlintIntensity scales the highlight brightness [0.0-1.0]
	GlintIntensity float64

	// Enabled controls whether glints render for this entity
	Enabled bool

	// CreatureType hints at eye style (humanoid, insect, serpent, etc.)
	CreatureType string
}

// Type returns the component type identifier for ECS registration.
func (c *Component) Type() string {
	return "eyeglint"
}

// NewComponent creates an eye glint component with default values.
func NewComponent() *Component {
	return &Component{
		EyePositions:   make([][2]int, 0, 4),
		EyeSizes:       make([]int, 0, 4),
		GlintPhase:     0.0,
		GlintIntensity: 0.8,
		Enabled:        true,
		CreatureType:   "humanoid",
	}
}

// AddEye registers an eye position for glint rendering.
func (c *Component) AddEye(x, y, radius int) {
	c.EyePositions = append(c.EyePositions, [2]int{x, y})
	c.EyeSizes = append(c.EyeSizes, radius)
}

// ClearEyes removes all registered eye positions.
func (c *Component) ClearEyes() {
	c.EyePositions = c.EyePositions[:0]
	c.EyeSizes = c.EyeSizes[:0]
}

// EyeCount returns the number of detected eyes.
func (c *Component) EyeCount() int {
	return len(c.EyePositions)
}

// GenrePreset defines eye glint appearance parameters for each genre.
type GenrePreset struct {
	// PrimaryColor is the main highlight color (usually bright white/cream)
	PrimaryColor color.RGBA
	// SecondaryColor is the smaller secondary reflection
	SecondaryColor color.RGBA
	// PrimarySize is the highlight size relative to eye radius [0.0-1.0]
	PrimarySize float64
	// SecondarySize is the secondary highlight size relative to eye
	SecondarySize float64
	// PrimaryOffset is the highlight position offset from center [0.0-1.0]
	PrimaryOffset float64
	// AnimationSpeed controls subtle glint movement (radians/sec)
	AnimationSpeed float64
	// AnimationAmplitude controls how much glint position varies
	AnimationAmplitude float64
	// GlowRadius adds soft glow around highlights (0 = no glow)
	GlowRadius float64
	// Saturation affects eye color vibrancy boost [0.0-1.0]
	Saturation float64
}

// DefaultGenrePresets returns preset configurations for all genres.
func DefaultGenrePresets() map[string]GenrePreset {
	return map[string]GenrePreset{
		"fantasy": {
			PrimaryColor:       color.RGBA{R: 255, G: 252, B: 240, A: 255},
			SecondaryColor:     color.RGBA{R: 255, G: 245, B: 220, A: 180},
			PrimarySize:        0.35,
			SecondarySize:      0.18,
			PrimaryOffset:      0.25,
			AnimationSpeed:     0.8,
			AnimationAmplitude: 0.03,
			GlowRadius:         1.2,
			Saturation:         0.15,
		},
		"scifi": {
			PrimaryColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			SecondaryColor:     color.RGBA{R: 200, G: 220, B: 255, A: 200},
			PrimarySize:        0.30,
			SecondarySize:      0.15,
			PrimaryOffset:      0.28,
			AnimationSpeed:     1.2,
			AnimationAmplitude: 0.02,
			GlowRadius:         0.8,
			Saturation:         0.10,
		},
		"horror": {
			PrimaryColor:       color.RGBA{R: 240, G: 230, B: 220, A: 220},
			SecondaryColor:     color.RGBA{R: 200, G: 180, B: 180, A: 140},
			PrimarySize:        0.28,
			SecondarySize:      0.12,
			PrimaryOffset:      0.22,
			AnimationSpeed:     0.4,
			AnimationAmplitude: 0.05,
			GlowRadius:         0.5,
			Saturation:         -0.1, // Desaturated
		},
		"cyberpunk": {
			PrimaryColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			SecondaryColor:     color.RGBA{R: 255, G: 100, B: 255, A: 180},
			PrimarySize:        0.32,
			SecondarySize:      0.20,
			PrimaryOffset:      0.30,
			AnimationSpeed:     1.5,
			AnimationAmplitude: 0.04,
			GlowRadius:         1.5,
			Saturation:         0.20,
		},
		"postapoc": {
			PrimaryColor:       color.RGBA{R: 240, G: 235, B: 220, A: 200},
			SecondaryColor:     color.RGBA{R: 200, G: 190, B: 170, A: 140},
			PrimarySize:        0.25,
			SecondarySize:      0.10,
			PrimaryOffset:      0.20,
			AnimationSpeed:     0.6,
			AnimationAmplitude: 0.02,
			GlowRadius:         0.3,
			Saturation:         -0.05,
		},
	}
}

// GetPreset returns the genre preset, falling back to fantasy if not found.
func GetPreset(genreID string) GenrePreset {
	presets := DefaultGenrePresets()
	if preset, ok := presets[genreID]; ok {
		return preset
	}
	return presets["fantasy"]
}

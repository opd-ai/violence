package floorreflect

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines reflection behavior per genre.
type GenrePreset struct {
	// DefaultReflectivity when no material is specified
	DefaultReflectivity float64

	// BaseIntensity multiplier for all reflections
	BaseIntensity float64

	// MaxReflectionHeight limits how tall reflections can be (performance)
	MaxReflectionHeight int

	// DistortionScale affects water/liquid distortion amount
	DistortionScale float64

	// DefaultTint applied to all reflections
	TintR, TintG, TintB float64

	// FadeExponent controls vertical fade curve (higher = sharper cutoff)
	FadeExponent float64

	// MinLightLevel for reflection visibility
	MinLightLevel float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		DefaultReflectivity: 0.25,
		BaseIntensity:       0.4,
		MaxReflectionHeight: 48,
		DistortionScale:     1.0,
		TintR:               0.95, TintG: 0.9, TintB: 0.85, // Warm torch tint
		FadeExponent:  1.5,
		MinLightLevel: 0.1,
	},
	"scifi": {
		DefaultReflectivity: 0.5,
		BaseIntensity:       0.6,
		MaxReflectionHeight: 64,
		DistortionScale:     0.3,
		TintR:               0.9, TintG: 0.95, TintB: 1.1, // Cool tech tint
		FadeExponent:  1.2,
		MinLightLevel: 0.15,
	},
	"horror": {
		DefaultReflectivity: 0.35,
		BaseIntensity:       0.45,
		MaxReflectionHeight: 40,
		DistortionScale:     1.5,
		TintR:               0.8, TintG: 0.85, TintB: 0.9, // Murky tint
		FadeExponent:  2.0,
		MinLightLevel: 0.05,
	},
	"cyberpunk": {
		DefaultReflectivity: 0.6,
		BaseIntensity:       0.7,
		MaxReflectionHeight: 72,
		DistortionScale:     0.8,
		TintR:               1.0, TintG: 0.9, TintB: 1.15, // Neon tint
		FadeExponent:  1.0,
		MinLightLevel: 0.2,
	},
	"postapoc": {
		DefaultReflectivity: 0.3,
		BaseIntensity:       0.35,
		MaxReflectionHeight: 36,
		DistortionScale:     1.2,
		TintR:               1.1, TintG: 0.95, TintB: 0.8, // Dusty warm tint
		FadeExponent:  1.8,
		MinLightLevel: 0.08,
	},
}

// System manages floor reflection rendering.
type System struct {
	genreID string
	preset  GenrePreset
	logger  *logrus.Entry

	// Reflective floor tiles (keyed by tileX*10000+tileY)
	reflectiveFloors map[int]FloorTileReflect

	// Screen dimensions for bounds checking
	screenWidth, screenHeight int

	// RNG for distortion effects
	rng  *rand.Rand
	seed int64

	// Pre-allocated reflection buffer for performance
	reflectionBuffer *image.RGBA

	// Frame counter for animated distortion
	frameCount int
}

// NewSystem creates a floor reflection system for the specified genre.
func NewSystem(genreID string) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	return &System{
		genreID:          genreID,
		preset:           preset,
		logger:           logrus.WithField("system", "floorreflect"),
		reflectiveFloors: make(map[int]FloorTileReflect),
		screenWidth:      320,
		screenHeight:     200,
		rng:              rand.New(rand.NewSource(42)),
		seed:             42,
		frameCount:       0,
	}
}

// SetScreenSize updates the screen dimensions for bounds checking.
func (s *System) SetScreenSize(width, height int) {
	s.screenWidth = width
	s.screenHeight = height

	// Resize reflection buffer if needed
	maxH := s.preset.MaxReflectionHeight
	if s.reflectionBuffer == nil || s.reflectionBuffer.Bounds().Dx() < width {
		s.reflectionBuffer = image.NewRGBA(image.Rect(0, 0, width, maxH))
	}
}

// SetGenre changes the genre and updates presets.
func (s *System) SetGenre(genreID string) {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}
	s.genreID = genreID
	s.preset = preset

	s.logger.WithFields(logrus.Fields{
		"genre":        genreID,
		"intensity":    preset.BaseIntensity,
		"reflectivity": preset.DefaultReflectivity,
	}).Debug("Floor reflection genre updated")
}

// SetSeed sets the random seed for deterministic distortion.
func (s *System) SetSeed(seed int64) {
	s.seed = seed
	s.rng = rand.New(rand.NewSource(seed))
}

// AddReflectiveFloor registers a floor tile as reflective.
func (s *System) AddReflectiveFloor(tileX, tileY int, material MaterialReflectivity, lightLevel float64) {
	key := tileX*10000 + tileY
	s.reflectiveFloors[key] = FloorTileReflect{
		TileX:      tileX,
		TileY:      tileY,
		Material:   material,
		LightLevel: lightLevel,
	}
}

// ClearReflectiveFloors removes all registered reflective floors.
func (s *System) ClearReflectiveFloors() {
	s.reflectiveFloors = make(map[int]FloorTileReflect)
}

// SetReflectiveFloorsFromWetness populates reflective floors from wetness data.
// This integrates with the wetness system.
func (s *System) SetReflectiveFloorsFromWetness(wetTiles map[int]float64, tileSize int) {
	s.ClearReflectiveFloors()

	for key, wetness := range wetTiles {
		tileX := key / 10000
		tileY := key % 10000

		// Scale material reflectivity by wetness amount
		var mat MaterialReflectivity
		if wetness > 0.7 {
			mat = ReflectWater
		} else if wetness > 0.4 {
			mat = ReflectWetStone
		} else if wetness > 0.1 {
			mat = MaterialReflectivity{
				Reflectivity: 0.2 * wetness,
				TintR:        0.95, TintG: 0.97, TintB: 1.0,
				Distortion: 0.05 * wetness,
				FadeRate:   1.5,
			}
		} else {
			continue // Not reflective enough
		}

		s.AddReflectiveFloor(tileX, tileY, mat, 0.5) // Default light level
	}
}

// IsFloorReflective checks if a world position has a reflective floor.
func (s *System) IsFloorReflective(worldX, worldY float64, tileSize int) (bool, MaterialReflectivity) {
	tileX := int(worldX) / tileSize
	tileY := int(worldY) / tileSize
	key := tileX*10000 + tileY

	if tile, ok := s.reflectiveFloors[key]; ok {
		return tile.Material.Reflectivity > 0.01, tile.Material
	}

	return false, ReflectNone
}

// GetFloorReflectivity returns the reflectivity at a world position.
func (s *System) GetFloorReflectivity(worldX, worldY float64, tileSize int) float64 {
	tileX := int(worldX) / tileSize
	tileY := int(worldY) / tileSize
	key := tileX*10000 + tileY

	if tile, ok := s.reflectiveFloors[key]; ok {
		return tile.Material.Reflectivity
	}

	return s.preset.DefaultReflectivity
}

// Update advances animation state.
func (s *System) Update() {
	s.frameCount++
}

// GenerateReflection creates a reflected sprite image.
func (s *System) GenerateReflection(source *image.RGBA, material MaterialReflectivity, lightLevel float64) *image.RGBA {
	if source == nil {
		return nil
	}

	bounds := source.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Limit reflection height for performance
	reflectHeight := height
	if reflectHeight > s.preset.MaxReflectionHeight {
		reflectHeight = s.preset.MaxReflectionHeight
	}

	// Create reflection image (vertically flipped)
	reflection := image.NewRGBA(image.Rect(0, 0, width, reflectHeight))

	// Calculate effective intensity
	intensity := s.preset.BaseIntensity * material.Reflectivity
	if lightLevel > 0 {
		intensity *= (0.5 + 0.5*lightLevel) // Brighter floors show stronger reflections
	}

	// Generate distortion seed for this frame
	distortSeed := s.seed ^ int64(s.frameCount/4) // Change every 4 frames

	for y := 0; y < reflectHeight; y++ {
		// Vertical fade: stronger at top (near floor contact), fades down
		fadeT := float64(y) / float64(reflectHeight)
		fade := math.Pow(1.0-fadeT, s.preset.FadeExponent)

		// Source Y is flipped: bottom of source maps to top of reflection
		sourceY := bounds.Max.Y - 1 - y
		if sourceY < bounds.Min.Y {
			continue
		}

		for x := 0; x < width; x++ {
			// Apply distortion for water/liquid surfaces
			sampleX := x
			if material.Distortion > 0 {
				distRng := rand.New(rand.NewSource(distortSeed + int64(x*1000+y)))
				distortOffset := int((distRng.Float64() - 0.5) * 2.0 * material.Distortion * s.preset.DistortionScale * 4.0)
				sampleX = x + distortOffset
				if sampleX < 0 {
					sampleX = 0
				}
				if sampleX >= width {
					sampleX = width - 1
				}
			}

			// Get source pixel
			srcColor := source.At(sampleX+bounds.Min.X, sourceY).(color.RGBA)

			// Skip fully transparent pixels
			if srcColor.A == 0 {
				continue
			}

			// Apply material tint
			tinted := tintColor(srcColor, material.TintR*s.preset.TintR,
				material.TintG*s.preset.TintG,
				material.TintB*s.preset.TintB)

			// Calculate final alpha with fade and intensity
			alpha := float64(tinted.A) / 255.0
			finalAlpha := alpha * fade * intensity

			if finalAlpha < 0.01 {
				continue
			}

			// Darken reflection slightly (reflections are never as bright as source)
			darken := 0.7
			reflection.Set(x, y, color.RGBA{
				R: uint8(float64(tinted.R) * darken),
				G: uint8(float64(tinted.G) * darken),
				B: uint8(float64(tinted.B) * darken),
				A: uint8(finalAlpha * 255),
			})
		}
	}

	return reflection
}

// RenderReflection draws a sprite's reflection to the screen.
func (s *System) RenderReflection(screen *ebiten.Image, comp *Component, screenX, screenY, tileSize int) {
	if !comp.Enabled || comp.SourceImage == nil {
		return
	}

	// Check if floor is reflective at entity position
	reflective, material := s.IsFloorReflective(comp.X, comp.Y, tileSize)
	if !reflective && s.preset.DefaultReflectivity < 0.05 {
		return
	}

	if !reflective {
		material = MaterialReflectivity{
			Reflectivity: s.preset.DefaultReflectivity,
			TintR:        s.preset.TintR, TintG: s.preset.TintG, TintB: s.preset.TintB,
			Distortion: 0.0,
			FadeRate:   1.0,
		}
	}

	// Get floor light level (simplified - could integrate with lighting system)
	lightLevel := 0.5
	key := int(comp.X)/tileSize*10000 + int(comp.Y)/tileSize
	if tile, ok := s.reflectiveFloors[key]; ok {
		lightLevel = tile.LightLevel
	}

	// Use intensity override if set
	if comp.IntensityOverride >= 0 {
		material.Reflectivity = comp.IntensityOverride
	}

	// Generate or use cached reflection
	reflection := s.GenerateReflection(comp.SourceImage, material, lightLevel)
	if reflection == nil {
		return
	}

	// Calculate reflection screen position (below sprite)
	bounds := comp.SourceImage.Bounds()
	reflectY := screenY + bounds.Dy() + int(comp.FloorOffset)

	// Bounds check
	if reflectY >= s.screenHeight || screenX+bounds.Dx() < 0 || screenX >= s.screenWidth {
		return
	}

	// Draw reflection
	ebitenImg := ebiten.NewImageFromImage(reflection)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(screenX), float64(reflectY))

	// Apply transparency blend
	opts.ColorScale.ScaleAlpha(float32(s.preset.BaseIntensity))

	screen.DrawImage(ebitenImg, opts)
}

// RenderReflections renders reflections for multiple entities efficiently.
func (s *System) RenderReflections(screen *ebiten.Image, entities []ReflectionEntity, tileSize int, cameraX, cameraY float64) {
	for _, entity := range entities {
		if entity.Component == nil || !entity.Component.Enabled {
			continue
		}

		// Calculate screen position
		screenX := int(entity.WorldX - cameraX)
		screenY := int(entity.WorldY - cameraY)

		// Update component position
		entity.Component.SetPosition(entity.WorldX, entity.WorldY)

		s.RenderReflection(screen, entity.Component, screenX, screenY, tileSize)
	}
}

// ReflectionEntity pairs a component with world position for batch rendering.
type ReflectionEntity struct {
	Component        *Component
	WorldX, WorldY   float64
	ScreenX, ScreenY int
}

// GetPreset returns the current genre preset (for testing/debugging).
func (s *System) GetPreset() GenrePreset {
	return s.preset
}

// GetGenreID returns the current genre.
func (s *System) GetGenreID() string {
	return s.genreID
}

// GetReflectiveFloorCount returns the number of registered reflective floors.
func (s *System) GetReflectiveFloorCount() int {
	return len(s.reflectiveFloors)
}

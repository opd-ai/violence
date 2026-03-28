package surfacesheen

import (
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines sheen appearance parameters for each genre.
type GenrePreset struct {
	// BaseIntensity scales all sheen effects [0.0-1.0].
	BaseIntensity float64

	// WarmthShift adjusts color temperature (-1.0 cool to 1.0 warm).
	WarmthShift float64

	// SpecularTightness controls highlight concentration [0.1-5.0].
	SpecularTightness float64

	// ReflectionTint adds a color tint to all reflections.
	ReflectionTint color.RGBA

	// WetSheenBoost multiplies wet surface reflections.
	WetSheenBoost float64

	// MetalSaturation controls how much metal reflects its own color [0.0-1.0].
	MetalSaturation float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BaseIntensity:     0.7,
		WarmthShift:       0.2,
		SpecularTightness: 1.5,
		ReflectionTint:    color.RGBA{R: 255, G: 245, B: 220, A: 255},
		WetSheenBoost:     1.2,
		MetalSaturation:   0.8,
	},
	"scifi": {
		BaseIntensity:     0.9,
		WarmthShift:       -0.3,
		SpecularTightness: 2.5,
		ReflectionTint:    color.RGBA{R: 200, G: 220, B: 255, A: 255},
		WetSheenBoost:     1.0,
		MetalSaturation:   0.9,
	},
	"horror": {
		BaseIntensity:     0.4,
		WarmthShift:       -0.1,
		SpecularTightness: 0.8,
		ReflectionTint:    color.RGBA{R: 180, G: 200, B: 180, A: 255},
		WetSheenBoost:     1.8,
		MetalSaturation:   0.5,
	},
	"cyberpunk": {
		BaseIntensity:     1.0,
		WarmthShift:       0.1,
		SpecularTightness: 3.0,
		ReflectionTint:    color.RGBA{R: 255, G: 180, B: 255, A: 255},
		WetSheenBoost:     1.3,
		MetalSaturation:   1.0,
	},
	"postapoc": {
		BaseIntensity:     0.5,
		WarmthShift:       0.3,
		SpecularTightness: 1.0,
		ReflectionTint:    color.RGBA{R: 200, G: 180, B: 150, A: 255},
		WetSheenBoost:     0.8,
		MetalSaturation:   0.6,
	},
}

// System manages surface sheen rendering for all entities.
type System struct {
	mu            sync.RWMutex
	genre         string
	preset        GenrePreset
	pixelsPerUnit float64
	logger        *logrus.Entry

	// Overlay image for additive blending
	overlay   *ebiten.Image
	overlayW  int
	overlayH  int
	overlayMu sync.Mutex

	// Light direction for consistent specular (normalized)
	lightDirX float64
	lightDirY float64
	lightDirZ float64
}

// NewSystem creates a surface sheen rendering system.
func NewSystem(genreID string) *System {
	s := &System{
		genre:         genreID,
		pixelsPerUnit: 32.0,
		lightDirX:     -0.5,
		lightDirY:     -0.7,
		lightDirZ:     0.5,
		logger: logrus.WithFields(logrus.Fields{
			"system": "surfacesheen",
		}),
	}
	s.applyGenrePreset(genreID)
	s.normalizeLight()
	return s
}

// normalizeLight ensures the light direction vector is unit length.
func (s *System) normalizeLight() {
	length := math.Sqrt(s.lightDirX*s.lightDirX + s.lightDirY*s.lightDirY + s.lightDirZ*s.lightDirZ)
	if length > 0 {
		s.lightDirX /= length
		s.lightDirY /= length
		s.lightDirZ /= length
	}
}

// applyGenrePreset configures sheen parameters based on genre.
func (s *System) applyGenrePreset(genreID string) {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
		s.logger.Warnf("unknown genre %s, using fantasy defaults", genreID)
	}
	s.preset = preset
	s.logger.Debugf("applied preset: intensity=%.2f, warmth=%.2f, tightness=%.2f",
		preset.BaseIntensity, preset.WarmthShift, preset.SpecularTightness)
}

// SetGenre updates sheen parameters for a new genre.
func (s *System) SetGenre(genreID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.genre = genreID
	s.applyGenrePreset(genreID)
}

// GetGenre returns the current genre ID.
func (s *System) GetGenre() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.genre
}

// ensureOverlay creates or resizes the overlay image.
func (s *System) ensureOverlay(width, height int) {
	s.overlayMu.Lock()
	defer s.overlayMu.Unlock()

	if s.overlay == nil || s.overlayW != width || s.overlayH != height {
		if s.overlay != nil {
			s.overlay.Dispose()
		}
		s.overlay = ebiten.NewImage(width, height)
		s.overlayW = width
		s.overlayH = height
	}
}

// CalculateSheenForEntity computes the sheen contribution for a single entity.
// Returns the sheen color and intensity to be blended additively.
func (s *System) CalculateSheenForEntity(
	comp *SheenComponent,
	entityX, entityY float64,
	entityRadius float64,
	lights []LightSource,
) (color.RGBA, float64) {
	s.mu.RLock()
	preset := s.preset
	s.mu.RUnlock()

	if comp == nil || comp.Intensity <= 0 {
		return color.RGBA{}, 0
	}

	// Accumulate sheen from all lights
	var totalR, totalG, totalB float64
	var totalIntensity float64

	for _, light := range lights {
		sheenR, sheenG, sheenB, intensity := s.calculateSingleLightSheen(
			comp, entityX, entityY, entityRadius, light, preset,
		)
		totalR += sheenR * intensity
		totalG += sheenG * intensity
		totalB += sheenB * intensity
		totalIntensity += intensity
	}

	if totalIntensity <= 0 {
		return color.RGBA{}, 0
	}

	// Normalize accumulated color
	totalR /= totalIntensity
	totalG /= totalIntensity
	totalB /= totalIntensity

	// Apply final intensity cap
	totalIntensity = math.Min(1.0, totalIntensity)

	return color.RGBA{
		R: clampUint8(totalR),
		G: clampUint8(totalG),
		B: clampUint8(totalB),
		A: 255,
	}, totalIntensity
}

// calculateSingleLightSheen computes sheen from one light source.
func (s *System) calculateSingleLightSheen(
	comp *SheenComponent,
	entityX, entityY float64,
	entityRadius float64,
	light LightSource,
	preset GenrePreset,
) (r, g, b, intensity float64) {
	// Direction from entity to light
	dx := light.X - entityX
	dy := light.Y - entityY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= 0 || dist > light.Radius*2 {
		return 0, 0, 0, 0
	}

	// Normalize direction
	dirX := dx / dist
	dirY := dy / dist

	// Distance falloff (inverse-square with cutoff)
	distFactor := light.Radius / (dist + light.Radius*0.1)
	distFactor = math.Min(1.0, distFactor*distFactor)

	// Calculate specular factor based on material
	specular := s.calculateSpecular(comp.Material, dirX, dirY, comp.Roughness, preset)

	// Material-specific color contribution
	sheenColor := s.calculateMaterialColor(comp, light, preset)

	// Apply wetness boost
	wetBoost := 1.0
	if comp.Wetness > 0 {
		wetBoost = 1.0 + comp.Wetness*preset.WetSheenBoost
	}

	// Final intensity calculation
	intensity = specular * distFactor * light.Intensity * comp.Intensity * preset.BaseIntensity * wetBoost

	return float64(sheenColor.R), float64(sheenColor.G), float64(sheenColor.B), intensity
}

// calculateSpecular computes the specular highlight factor.
func (s *System) calculateSpecular(
	material MaterialType,
	lightDirX, lightDirY float64,
	roughness float64,
	preset GenrePreset,
) float64 {
	// Calculate half-vector (simplified 2D)
	viewDirX := 0.0
	viewDirY := -1.0 // Assume camera looking down
	viewDirZ := 0.5

	halfX := lightDirX + viewDirX
	halfY := lightDirY + viewDirY
	halfZ := s.lightDirZ + viewDirZ
	halfLen := math.Sqrt(halfX*halfX + halfY*halfY + halfZ*halfZ)
	if halfLen > 0 {
		halfX /= halfLen
		halfY /= halfLen
		halfZ /= halfLen
	}

	// Normal points up (simplified for top-down)
	normalX := 0.0
	normalY := 0.0
	normalZ := 1.0

	// N dot H
	NdotH := math.Max(0, normalX*halfX+normalY*halfY+normalZ*halfZ)

	// Specular power based on roughness and material
	power := (1.0 - roughness) * 50.0 * preset.SpecularTightness
	if power < 1.0 {
		power = 1.0
	}

	// Material-specific adjustments
	switch material {
	case MaterialMetal:
		power *= 1.5
	case MaterialPolished:
		power *= 2.0
	case MaterialCrystal:
		power *= 1.8
	case MaterialWet:
		power *= 0.8 // Broader highlight
	case MaterialOrganic, MaterialCloth:
		power *= 0.5
	}

	return math.Pow(NdotH, power)
}

// calculateMaterialColor determines the sheen color based on material and light.
func (s *System) calculateMaterialColor(
	comp *SheenComponent,
	light LightSource,
	preset GenrePreset,
) color.RGBA {
	var r, g, b float64

	switch comp.Material {
	case MaterialMetal:
		// Metal reflects its own color mixed with light color
		sat := preset.MetalSaturation
		r = lerp(float64(light.Color.R), float64(comp.BaseColor.R), sat)
		g = lerp(float64(light.Color.G), float64(comp.BaseColor.G), sat)
		b = lerp(float64(light.Color.B), float64(comp.BaseColor.B), sat)

	case MaterialWet:
		// Wet surfaces reflect mostly light color with slight base tint
		r = float64(light.Color.R)*0.9 + float64(comp.BaseColor.R)*0.1
		g = float64(light.Color.G)*0.9 + float64(comp.BaseColor.G)*0.1
		b = float64(light.Color.B)*0.9 + float64(comp.BaseColor.B)*0.1

	case MaterialPolished:
		// Polished is mostly light color
		r = float64(light.Color.R)
		g = float64(light.Color.G)
		b = float64(light.Color.B)

	case MaterialCrystal:
		// Crystal adds prismatic effect
		r = float64(light.Color.R) * (1.0 + 0.2*math.Sin(float64(comp.BaseColor.R)/255.0*math.Pi))
		g = float64(light.Color.G) * (1.0 + 0.2*math.Sin(float64(comp.BaseColor.G)/255.0*math.Pi))
		b = float64(light.Color.B) * (1.0 + 0.2*math.Sin(float64(comp.BaseColor.B)/255.0*math.Pi))

	case MaterialOrganic:
		// Organic has soft subsurface influence
		r = lerp(float64(light.Color.R), float64(comp.BaseColor.R), 0.6)
		g = lerp(float64(light.Color.G), float64(comp.BaseColor.G), 0.6)
		b = lerp(float64(light.Color.B), float64(comp.BaseColor.B), 0.6)

	default:
		// Default blend
		r = lerp(float64(light.Color.R), float64(preset.ReflectionTint.R), 0.3)
		g = lerp(float64(light.Color.G), float64(preset.ReflectionTint.G), 0.3)
		b = lerp(float64(light.Color.B), float64(preset.ReflectionTint.B), 0.3)
	}

	// Apply genre warmth shift
	if preset.WarmthShift > 0 {
		r = math.Min(255, r*(1+preset.WarmthShift*0.15))
		b = b * (1 - preset.WarmthShift*0.1)
	} else if preset.WarmthShift < 0 {
		r = r * (1 + preset.WarmthShift*0.1)
		b = math.Min(255, b*(1-preset.WarmthShift*0.15))
	}

	return color.RGBA{
		R: clampUint8(r),
		G: clampUint8(g),
		B: clampUint8(b),
		A: 255,
	}
}

// RenderSheenOverlay renders sheen effects for multiple entities.
// screenX, screenY are entity screen positions.
// radiusPx are entity radii in screen pixels.
func (s *System) RenderSheenOverlay(
	screen *ebiten.Image,
	components []*SheenComponent,
	screenX, screenY []float64,
	radiusPx []float64,
	worldX, worldY []float64,
	lights []LightSource,
) {
	if len(components) == 0 || len(lights) == 0 {
		return
	}

	bounds := screen.Bounds()
	s.ensureOverlay(bounds.Dx(), bounds.Dy())

	s.overlayMu.Lock()
	s.overlay.Clear()
	s.overlayMu.Unlock()

	for i, comp := range components {
		if comp == nil || i >= len(screenX) || i >= len(screenY) || i >= len(radiusPx) {
			continue
		}

		worldRadius := 0.5 // Default world radius
		if i < len(worldX) && i < len(worldY) {
			sheenColor, intensity := s.CalculateSheenForEntity(
				comp, worldX[i], worldY[i], worldRadius, lights,
			)

			if intensity > 0.01 {
				s.drawSheenSpot(screenX[i], screenY[i], radiusPx[i], sheenColor, intensity)
			}
		}
	}

	// Composite overlay additively
	opts := &ebiten.DrawImageOptions{}
	opts.Blend = ebiten.BlendLighter
	s.overlayMu.Lock()
	screen.DrawImage(s.overlay, opts)
	s.overlayMu.Unlock()
}

// drawSheenSpot draws a radial sheen highlight at the specified position.
func (s *System) drawSheenSpot(x, y, radius float64, col color.RGBA, intensity float64) {
	s.overlayMu.Lock()
	defer s.overlayMu.Unlock()

	if s.overlay == nil {
		return
	}

	// Draw a small radial gradient
	halfSize := int(radius * 0.6)
	if halfSize < 2 {
		halfSize = 2
	}

	centerX := int(x)
	centerY := int(y - radius*0.3) // Offset up for top-down perspective

	bounds := s.overlay.Bounds()

	for dy := -halfSize; dy <= halfSize; dy++ {
		for dx := -halfSize; dx <= halfSize; dx++ {
			px := centerX + dx
			py := centerY + dy

			if px < bounds.Min.X || px >= bounds.Max.X || py < bounds.Min.Y || py >= bounds.Max.Y {
				continue
			}

			// Radial falloff
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			falloff := 1.0 - (dist / float64(halfSize))
			if falloff <= 0 {
				continue
			}

			// Quadratic falloff for sharper highlight
			falloff *= falloff

			alpha := falloff * intensity * 0.6

			// Get existing pixel and blend
			r := uint8(float64(col.R) * alpha)
			g := uint8(float64(col.G) * alpha)
			b := uint8(float64(col.B) * alpha)
			a := uint8(alpha * 255)

			s.overlay.Set(px, py, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}
}

// lerp performs linear interpolation between two values.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// clampUint8 clamps a float64 to valid uint8 range.
func clampUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

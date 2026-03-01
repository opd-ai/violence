// Package lighting provides shadow casting and rendering.
package lighting

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// ShadowCaster represents an entity that casts a shadow.
type ShadowCaster struct {
	X, Y         float64 // Position in world space
	Radius       float64 // Approximate size for shadow calculation
	Height       float64 // Vertical offset (affects shadow size/offset)
	Opacity      float64 // Shadow opacity [0.0-1.0]
	CastShadow   bool    // Whether this entity casts shadows
	ShadowOffset float64 // Shadow offset from entity center (genre-specific)
}

// ShadowSystem manages shadow rendering for all entities.
type ShadowSystem struct {
	width       int
	height      int
	genre       string
	shadowMap   *ebiten.Image
	softness    float64 // Penumbra softness [0.0-1.0]
	minOpacity  float64 // Minimum shadow opacity
	maxOpacity  float64 // Maximum shadow opacity
	falloffType string  // "linear", "quadratic", "inverse"
	logger      *logrus.Entry
}

// NewShadowSystem creates a shadow rendering system.
func NewShadowSystem(width, height int, genreID string) *ShadowSystem {
	s := &ShadowSystem{
		width:       width,
		height:      height,
		genre:       genreID,
		shadowMap:   ebiten.NewImage(width, height),
		softness:    0.5,
		minOpacity:  0.2,
		maxOpacity:  0.7,
		falloffType: "quadratic",
		logger: logrus.WithFields(logrus.Fields{
			"system": "shadow",
			"genre":  genreID,
		}),
	}
	s.applyGenrePreset(genreID)
	return s
}

// applyGenrePreset configures shadow parameters based on genre.
func (s *ShadowSystem) applyGenrePreset(genreID string) {
	switch genreID {
	case "fantasy":
		// Soft, medium shadows - torchlight and magic
		s.softness = 0.6
		s.minOpacity = 0.25
		s.maxOpacity = 0.65
		s.falloffType = "quadratic"
	case "scifi":
		// Harder shadows - artificial lighting
		s.softness = 0.3
		s.minOpacity = 0.3
		s.maxOpacity = 0.75
		s.falloffType = "linear"
	case "horror":
		// Very soft, deep shadows - atmospheric dread
		s.softness = 0.8
		s.minOpacity = 0.4
		s.maxOpacity = 0.85
		s.falloffType = "inverse"
	case "cyberpunk":
		// Hard, sharp shadows - neon contrast
		s.softness = 0.2
		s.minOpacity = 0.35
		s.maxOpacity = 0.8
		s.falloffType = "linear"
	case "postapoc":
		// Medium softness - diffuse daylight through dust
		s.softness = 0.5
		s.minOpacity = 0.3
		s.maxOpacity = 0.7
		s.falloffType = "quadratic"
	default:
		s.logger.Warnf("unknown genre %s, using fantasy defaults", genreID)
		// Apply fantasy defaults
		s.softness = 0.6
		s.minOpacity = 0.25
		s.maxOpacity = 0.65
		s.falloffType = "quadratic"
	}
	s.logger.Debugf("applied preset: softness=%.2f, opacity=[%.2f-%.2f], falloff=%s",
		s.softness, s.minOpacity, s.maxOpacity, s.falloffType)
}

// SetGenre updates shadow parameters for a new genre.
func (s *ShadowSystem) SetGenre(genreID string) {
	s.genre = genreID
	s.applyGenrePreset(genreID)
}

// RenderShadows draws shadows for all casters based on light sources.
func (s *ShadowSystem) RenderShadows(
	screen *ebiten.Image,
	casters []ShadowCaster,
	lights []Light,
	coneLights []ConeLight,
	cameraX, cameraY float64,
) {
	// Clear shadow map
	s.shadowMap.Clear()

	if len(lights) == 0 && len(coneLights) == 0 {
		return // No lights = no shadows
	}

	// Render shadow for each caster from each light
	for _, caster := range casters {
		if !caster.CastShadow {
			continue
		}

		// Process point lights
		for _, light := range lights {
			s.renderShadowFromLight(caster, light.X, light.Y, light.Intensity, cameraX, cameraY)
		}

		// Process cone lights (flashlights)
		for _, cone := range coneLights {
			if !cone.IsActive {
				continue
			}
			s.renderShadowFromLight(caster, cone.X, cone.Y, cone.Intensity, cameraX, cameraY)
		}
	}

	// Composite shadow map onto screen
	opts := &ebiten.DrawImageOptions{}
	opts.ColorScale.ScaleAlpha(0.8) // Overall shadow intensity
	screen.DrawImage(s.shadowMap, opts)
}

// renderShadowFromLight renders a single shadow from one light source.
func (s *ShadowSystem) renderShadowFromLight(
	caster ShadowCaster,
	lightX, lightY float64,
	lightIntensity float64,
	cameraX, cameraY float64,
) {
	// Vector from light to caster
	dx := caster.X - lightX
	dy := caster.Y - lightY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 0.1 {
		return // Caster is at light source
	}

	// Normalize direction
	dirX := dx / dist
	dirY := dy / dist

	// Calculate shadow parameters
	shadowLength := s.calculateShadowLength(dist, caster.Height, lightIntensity)
	shadowWidth := caster.Radius * 2.0 * (1.0 + shadowLength/10.0) // Widen with distance

	// Shadow end position (extends away from light)
	shadowEndX := caster.X + dirX*shadowLength
	shadowEndY := caster.Y + dirY*shadowLength

	// Calculate shadow opacity based on distance and light intensity
	opacity := s.calculateShadowOpacity(dist, lightIntensity)
	opacity *= caster.Opacity

	// Convert world coordinates to screen coordinates
	screenX := int((caster.X - cameraX) * 32) // Assuming 32 pixels per world unit
	screenY := int((caster.Y - cameraY) * 32)
	screenEndX := int((shadowEndX - cameraX) * 32)
	screenEndY := int((shadowEndY - cameraY) * 32)

	// Draw shadow ellipse with gradient (soft penumbra)
	s.drawSoftShadow(screenX, screenY, screenEndX, screenEndY, shadowWidth*32, opacity)
}

// calculateShadowLength determines shadow projection distance.
func (s *ShadowSystem) calculateShadowLength(distance, height, lightIntensity float64) float64 {
	// Shadow length increases with height and decreases with distance
	baseLength := height * 2.0
	distanceFactor := math.Max(0.5, 1.0/math.Sqrt(distance))
	intensityFactor := 1.0 / math.Max(0.3, lightIntensity)
	return baseLength * distanceFactor * intensityFactor
}

// calculateShadowOpacity determines shadow darkness based on distance and light.
func (s *ShadowSystem) calculateShadowOpacity(distance, lightIntensity float64) float64 {
	var attenuation float64
	switch s.falloffType {
	case "linear":
		attenuation = 1.0 / (1.0 + distance*0.1)
	case "quadratic":
		attenuation = 1.0 / (1.0 + distance*distance*0.01)
	case "inverse":
		attenuation = 1.0 / math.Max(1.0, distance)
	default:
		attenuation = 1.0 / (1.0 + distance*distance*0.01)
	}

	opacity := s.minOpacity + (s.maxOpacity-s.minOpacity)*attenuation*lightIntensity
	return clampF(opacity, s.minOpacity, s.maxOpacity)
}

// drawSoftShadow renders a shadow ellipse with gradient falloff.
func (s *ShadowSystem) drawSoftShadow(startX, startY, endX, endY int, width, opacity float64) {
	// Calculate midpoint and length
	midX := (startX + endX) / 2
	midY := (startY + endY) / 2
	length := math.Sqrt(float64((endX-startX)*(endX-startX) + (endY-startY)*(endY-startY)))

	if length < 1 {
		return
	}

	// Create shadow image
	shadowW := int(width) + int(s.softness*20)
	shadowH := int(length) + int(s.softness*20)
	if shadowW < 1 || shadowH < 1 {
		return
	}

	shadowImg := ebiten.NewImage(shadowW, shadowH)
	defer shadowImg.Dispose()

	// Draw gradient ellipse
	pixels := make([]byte, shadowW*shadowH*4)
	centerX := float64(shadowW) / 2
	centerY := float64(shadowH) / 2
	radiusX := width / 2
	radiusY := length / 2

	for y := 0; y < shadowH; y++ {
		for x := 0; x < shadowW; x++ {
			// Distance from ellipse center
			dx := (float64(x) - centerX) / radiusX
			dy := (float64(y) - centerY) / radiusY
			distFromCenter := math.Sqrt(dx*dx + dy*dy)

			// Gradient falloff for soft penumbra
			alpha := 0.0
			if distFromCenter < 1.0 {
				// Inside shadow core
				coreEdge := 1.0 - s.softness
				if distFromCenter < coreEdge {
					alpha = opacity
				} else {
					// Penumbra gradient
					t := (distFromCenter - coreEdge) / s.softness
					alpha = opacity * (1.0 - t*t) // Quadratic falloff
				}
			}

			idx := (y*shadowW + x) * 4
			pixels[idx] = 0                    // R
			pixels[idx+1] = 0                  // G
			pixels[idx+2] = 0                  // B
			pixels[idx+3] = uint8(alpha * 255) // A
		}
	}

	shadowImg.WritePixels(pixels)

	// Calculate rotation angle
	angle := math.Atan2(float64(endY-startY), float64(endX-startX))

	// Draw rotated shadow
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-centerX, -centerY)
	opts.GeoM.Rotate(angle)
	opts.GeoM.Translate(float64(midX), float64(midY))

	s.shadowMap.DrawImage(shadowImg, opts)
}

// GetShadowMap returns the current shadow map for debugging/visualization.
func (s *ShadowSystem) GetShadowMap() *ebiten.Image {
	return s.shadowMap
}

// Clear clears the shadow map.
func (s *ShadowSystem) Clear() {
	s.shadowMap.Clear()
}

// clampF restricts float64 value to [min, max] range.
func clampF(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Package lighting provides realistic atmospheric lighting with shadows and fog.
package lighting

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// AtmosphericConfig defines atmospheric lighting parameters.
type AtmosphericConfig struct {
	FogDensity        float64    // Fog thickness (0.0-1.0)
	FogColor          color.RGBA // Fog base color
	FogStartDistance  float64    // Distance where fog begins
	ShadowSoftness    float64    // Penumbra width (pixels)
	ShadowDarkness    float64    // Shadow opacity (0.0-1.0)
	OcclusionStrength float64    // Ambient occlusion intensity
	ColorTemperature  float64    // Overall temperature shift (-1.0 warm to +1.0 cool)
	DepthFadeStart    float64    // Distance where depth fade begins
	DepthFadeEnd      float64    // Distance where depth fade completes
	EnableShadows     bool       // Whether to cast shadows
	EnableFog         bool       // Whether to apply fog
	EnableOcclusion   bool       // Whether to apply AO
}

// AtmosphericLightingSystem manages realistic lighting with shadows and atmospheric effects.
type AtmosphericLightingSystem struct {
	config         AtmosphericConfig
	genre          string
	lightBuffer    []PointLight  // All active light sources
	shadowMap      *ebiten.Image // Precomputed shadow map
	shadowMapDirty bool
	occluders      []Occluder // Shadow-casting objects
	logger         *logrus.Entry
}

// Occluder represents an object that blocks light and casts shadows.
type Occluder struct {
	X, Y   float64 // Position
	Width  float64 // Bounding box width
	Height float64 // Bounding box height
	Type   OccluderType
}

// OccluderType defines shadow casting behavior.
type OccluderType int

const (
	OccluderWall   OccluderType = iota // Full shadow
	OccluderEntity                     // Soft shadow
	OccluderProp                       // Partial shadow
)

// NewAtmosphericLightingSystem creates a realistic lighting system.
func NewAtmosphericLightingSystem(genre string) *AtmosphericLightingSystem {
	sys := &AtmosphericLightingSystem{
		config:      getGenreAtmosphericConfig(genre),
		genre:       genre,
		lightBuffer: make([]PointLight, 0, 32),
		occluders:   make([]Occluder, 0, 128),
		logger: logrus.WithFields(logrus.Fields{
			"system": "atmospheric_lighting",
			"genre":  genre,
		}),
	}
	return sys
}

// getGenreAtmosphericConfig returns genre-specific atmospheric settings.
func getGenreAtmosphericConfig(genre string) AtmosphericConfig {
	switch genre {
	case "fantasy":
		return AtmosphericConfig{
			FogDensity:        0.35,
			FogColor:          color.RGBA{R: 60, G: 50, B: 40, A: 255},
			FogStartDistance:  8.0,
			ShadowSoftness:    3.0,
			ShadowDarkness:    0.70,
			OcclusionStrength: 0.40,
			ColorTemperature:  0.1, // Slightly warm (torchlight)
			DepthFadeStart:    15.0,
			DepthFadeEnd:      25.0,
			EnableShadows:     true,
			EnableFog:         true,
			EnableOcclusion:   true,
		}
	case "scifi":
		return AtmosphericConfig{
			FogDensity:        0.20,
			FogColor:          color.RGBA{R: 40, G: 50, B: 70, A: 255},
			FogStartDistance:  12.0,
			ShadowSoftness:    2.0,
			ShadowDarkness:    0.60,
			OcclusionStrength: 0.35,
			ColorTemperature:  -0.2, // Cool (artificial lighting)
			DepthFadeStart:    20.0,
			DepthFadeEnd:      35.0,
			EnableShadows:     true,
			EnableFog:         true,
			EnableOcclusion:   true,
		}
	case "horror":
		return AtmosphericConfig{
			FogDensity:        0.50,
			FogColor:          color.RGBA{R: 30, G: 35, B: 30, A: 255},
			FogStartDistance:  5.0,
			ShadowSoftness:    4.0,
			ShadowDarkness:    0.85,
			OcclusionStrength: 0.60,
			ColorTemperature:  -0.1, // Slightly cool (decay)
			DepthFadeStart:    8.0,
			DepthFadeEnd:      15.0,
			EnableShadows:     true,
			EnableFog:         true,
			EnableOcclusion:   true,
		}
	case "cyberpunk":
		return AtmosphericConfig{
			FogDensity:        0.30,
			FogColor:          color.RGBA{R: 70, G: 40, B: 80, A: 255},
			FogStartDistance:  10.0,
			ShadowSoftness:    2.5,
			ShadowDarkness:    0.65,
			OcclusionStrength: 0.30,
			ColorTemperature:  -0.3, // Cool (neon lighting)
			DepthFadeStart:    18.0,
			DepthFadeEnd:      30.0,
			EnableShadows:     true,
			EnableFog:         true,
			EnableOcclusion:   true,
		}
	case "postapoc":
		return AtmosphericConfig{
			FogDensity:        0.40,
			FogColor:          color.RGBA{R: 80, G: 70, B: 60, A: 255},
			FogStartDistance:  10.0,
			ShadowSoftness:    3.5,
			ShadowDarkness:    0.75,
			OcclusionStrength: 0.45,
			ColorTemperature:  0.2, // Warm (dusty atmosphere)
			DepthFadeStart:    12.0,
			DepthFadeEnd:      22.0,
			EnableShadows:     true,
			EnableFog:         true,
			EnableOcclusion:   true,
		}
	default:
		return AtmosphericConfig{
			FogDensity:        0.25,
			FogColor:          color.RGBA{R: 60, G: 60, B: 60, A: 255},
			FogStartDistance:  10.0,
			ShadowSoftness:    3.0,
			ShadowDarkness:    0.70,
			OcclusionStrength: 0.35,
			ColorTemperature:  0.0,
			DepthFadeStart:    15.0,
			DepthFadeEnd:      25.0,
			EnableShadows:     true,
			EnableFog:         true,
			EnableOcclusion:   true,
		}
	}
}

// SetGenre updates atmospheric configuration for a new genre.
func (s *AtmosphericLightingSystem) SetGenre(genre string) {
	s.genre = genre
	s.config = getGenreAtmosphericConfig(genre)
	s.shadowMapDirty = true
}

// RegisterLight adds a light source to the atmospheric system.
func (s *AtmosphericLightingSystem) RegisterLight(light PointLight) {
	s.lightBuffer = append(s.lightBuffer, light)
}

// ClearLights removes all registered lights.
func (s *AtmosphericLightingSystem) ClearLights() {
	s.lightBuffer = s.lightBuffer[:0]
}

// RegisterOccluder adds a shadow-casting object.
func (s *AtmosphericLightingSystem) RegisterOccluder(x, y, width, height float64, occType OccluderType) {
	s.occluders = append(s.occluders, Occluder{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Type:   occType,
	})
	s.shadowMapDirty = true
}

// ClearOccluders removes all shadow casters.
func (s *AtmosphericLightingSystem) ClearOccluders() {
	s.occluders = s.occluders[:0]
	s.shadowMapDirty = true
}

// CalculateLightingAtPoint computes final lighting at world position with all effects.
func (s *AtmosphericLightingSystem) CalculateLightingAtPoint(worldX, worldY, cameraX, cameraY float64) (r, g, b, a float64) {
	// Start with ambient light (from standard lighting system)
	totalR, totalG, totalB := 0.1, 0.1, 0.1 // Base ambient

	// Accumulate light from all sources
	for _, light := range s.lightBuffer {
		// Distance from light to point
		dx := worldX - light.X
		dy := worldY - light.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		// Skip if outside light radius
		if dist > light.Radius {
			continue
		}

		// Inverse-square attenuation
		attenuation := light.ApplyAttenuation(worldX, worldY)

		// Shadow casting
		shadowFactor := 1.0
		if s.config.EnableShadows {
			shadowFactor = s.calculateShadowFactor(light.X, light.Y, worldX, worldY)
		}

		// Apply light contribution with shadow
		contribution := attenuation * shadowFactor
		totalR += light.R * contribution
		totalG += light.G * contribution
		totalB += light.B * contribution
	}

	// Apply ambient occlusion in corners
	if s.config.EnableOcclusion {
		occlusionFactor := s.calculateOcclusionFactor(worldX, worldY)
		totalR *= occlusionFactor
		totalG *= occlusionFactor
		totalB *= occlusionFactor
	}

	// Apply atmospheric fog based on distance from camera
	if s.config.EnableFog {
		distFromCamera := math.Sqrt((worldX-cameraX)*(worldX-cameraX) + (worldY-cameraY)*(worldY-cameraY))
		fogFactor := s.calculateFogFactor(distFromCamera)

		fogR := float64(s.config.FogColor.R) / 255.0
		fogG := float64(s.config.FogColor.G) / 255.0
		fogB := float64(s.config.FogColor.B) / 255.0

		totalR = totalR*(1.0-fogFactor) + fogR*fogFactor
		totalG = totalG*(1.0-fogFactor) + fogG*fogFactor
		totalB = totalB*(1.0-fogFactor) + fogB*fogFactor
	}

	// Apply color temperature shift
	totalR, totalG, totalB = s.applyColorTemperature(totalR, totalG, totalB)

	// Apply depth-based desaturation
	distFromCamera := math.Sqrt((worldX-cameraX)*(worldX-cameraX) + (worldY-cameraY)*(worldY-cameraY))
	totalR, totalG, totalB, a = s.applyDepthFade(totalR, totalG, totalB, distFromCamera)

	// Clamp to valid range
	totalR = clampValue(totalR, 0.0, 1.0)
	totalG = clampValue(totalG, 0.0, 1.0)
	totalB = clampValue(totalB, 0.0, 1.0)
	a = clampValue(a, 0.0, 1.0)

	return totalR, totalG, totalB, a
}

// calculateShadowFactor determines how much light reaches a point through shadows.
// Returns 1.0 for fully lit, 0.0 for full shadow.
func (s *AtmosphericLightingSystem) calculateShadowFactor(lightX, lightY, targetX, targetY float64) float64 {
	// Ray from light to target
	rayDX := targetX - lightX
	rayDY := targetY - lightY
	rayLength := math.Sqrt(rayDX*rayDX + rayDY*rayDY)

	if rayLength < 0.001 {
		return 1.0 // Point is at light source
	}

	rayDirX := rayDX / rayLength
	rayDirY := rayDY / rayLength

	// Check intersection with occluders
	totalShadow := 0.0

	for _, occ := range s.occluders {
		if s.rayIntersectsOccluder(lightX, lightY, rayDirX, rayDirY, rayLength, occ) {
			// Calculate shadow strength based on occluder type
			shadowStrength := s.config.ShadowDarkness
			switch occ.Type {
			case OccluderWall:
				shadowStrength *= 1.0 // Full shadow
			case OccluderEntity:
				shadowStrength *= 0.7 // Softer shadow
			case OccluderProp:
				shadowStrength *= 0.5 // Partial shadow
			}

			// Apply penumbra softness based on distance
			distToOccluder := s.distanceToOccluder(lightX, lightY, occ)
			softnessFactor := 1.0 - clampValue(distToOccluder/s.config.ShadowSoftness, 0.0, 0.5)
			shadowStrength *= (0.5 + 0.5*softnessFactor)

			totalShadow += shadowStrength

			// Early exit if fully shadowed
			if totalShadow >= 1.0 {
				return 0.0
			}
		}
	}

	return clampValue(1.0-totalShadow, 0.0, 1.0)
}

// rayIntersectsOccluder checks if a ray intersects an axis-aligned bounding box.
func (s *AtmosphericLightingSystem) rayIntersectsOccluder(rayX, rayY, dirX, dirY, maxDist float64, occ Occluder) bool {
	// AABB bounds
	minX := occ.X - occ.Width/2
	maxX := occ.X + occ.Width/2
	minY := occ.Y - occ.Height/2
	maxY := occ.Y + occ.Height/2

	// Ray-AABB intersection (slab method)
	var tmin, tmax, tymin, tymax float64

	if dirX >= 0 {
		tmin = (minX - rayX) / dirX
		tmax = (maxX - rayX) / dirX
	} else {
		tmin = (maxX - rayX) / dirX
		tmax = (minX - rayX) / dirX
	}

	if dirY >= 0 {
		tymin = (minY - rayY) / dirY
		tymax = (maxY - rayY) / dirY
	} else {
		tymin = (maxY - rayY) / dirY
		tymax = (minY - rayY) / dirY
	}

	if (tmin > tymax) || (tymin > tmax) {
		return false
	}

	if tymin > tmin {
		tmin = tymin
	}
	if tymax < tmax {
		tmax = tymax
	}

	return tmin < maxDist && tmax > 0
}

// distanceToOccluder calculates distance from point to occluder edge.
func (s *AtmosphericLightingSystem) distanceToOccluder(x, y float64, occ Occluder) float64 {
	// Closest point on AABB to (x, y)
	closestX := clampValue(x, occ.X-occ.Width/2, occ.X+occ.Width/2)
	closestY := clampValue(y, occ.Y-occ.Height/2, occ.Y+occ.Height/2)

	dx := x - closestX
	dy := y - closestY
	return math.Sqrt(dx*dx + dy*dy)
}

// calculateOcclusionFactor computes ambient occlusion at a point.
// Returns darker values (0.0-1.0) in occluded areas.
func (s *AtmosphericLightingSystem) calculateOcclusionFactor(x, y float64) float64 {
	// Sample nearby occluders
	occlusionSamples := 8
	occlusionSum := 0.0
	sampleRadius := 2.0

	for i := 0; i < occlusionSamples; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(occlusionSamples)
		sampleX := x + math.Cos(angle)*sampleRadius
		sampleY := y + math.Sin(angle)*sampleRadius

		// Check if sample point is inside any occluder
		for _, occ := range s.occluders {
			if s.pointInsideOccluder(sampleX, sampleY, occ) {
				occlusionSum += 1.0
				break
			}
		}
	}

	occlusionRatio := occlusionSum / float64(occlusionSamples)
	return 1.0 - (occlusionRatio * s.config.OcclusionStrength)
}

// pointInsideOccluder checks if a point is within an occluder's bounds.
func (s *AtmosphericLightingSystem) pointInsideOccluder(x, y float64, occ Occluder) bool {
	return x >= occ.X-occ.Width/2 && x <= occ.X+occ.Width/2 &&
		y >= occ.Y-occ.Height/2 && y <= occ.Y+occ.Height/2
}

// calculateFogFactor computes fog intensity based on distance.
// Returns 0.0 for no fog, 1.0 for full fog.
func (s *AtmosphericLightingSystem) calculateFogFactor(distance float64) float64 {
	if distance < s.config.FogStartDistance {
		return 0.0
	}

	// Exponential fog falloff
	normalizedDist := (distance - s.config.FogStartDistance) / 10.0
	fogFactor := 1.0 - math.Exp(-s.config.FogDensity*normalizedDist)

	return clampValue(fogFactor, 0.0, 0.8) // Cap fog to preserve some visibility
}

// applyColorTemperature shifts RGB values toward warm or cool tones.
func (s *AtmosphericLightingSystem) applyColorTemperature(r, g, b float64) (float64, float64, float64) {
	temp := s.config.ColorTemperature

	if temp > 0 {
		// Cool shift (toward blue)
		r *= (1.0 - temp*0.2)
		g *= (1.0 - temp*0.1)
		b *= (1.0 + temp*0.3)
	} else if temp < 0 {
		// Warm shift (toward orange/red)
		temp = -temp
		r *= (1.0 + temp*0.3)
		g *= (1.0 + temp*0.1)
		b *= (1.0 - temp*0.2)
	}

	return r, g, b
}

// applyDepthFade applies atmospheric perspective (desaturation and fade with distance).
func (s *AtmosphericLightingSystem) applyDepthFade(r, g, b, distance float64) (float64, float64, float64, float64) {
	if distance < s.config.DepthFadeStart {
		return r, g, b, 1.0
	}

	// Fade factor from start to end distance
	fadeRange := s.config.DepthFadeEnd - s.config.DepthFadeStart
	fadeFactor := (distance - s.config.DepthFadeStart) / fadeRange
	fadeFactor = clampValue(fadeFactor, 0.0, 1.0)

	// Desaturate (move toward gray)
	avg := (r + g + b) / 3.0
	r = r*(1.0-fadeFactor*0.6) + avg*fadeFactor*0.6
	g = g*(1.0-fadeFactor*0.6) + avg*fadeFactor*0.6
	b = b*(1.0-fadeFactor*0.6) + avg*fadeFactor*0.6

	// Reduce alpha (fade out)
	alpha := 1.0 - fadeFactor*0.3

	return r, g, b, alpha
}

// GetConfig returns current atmospheric configuration.
func (s *AtmosphericLightingSystem) GetConfig() AtmosphericConfig {
	return s.config
}

// SetConfig updates atmospheric configuration.
func (s *AtmosphericLightingSystem) SetConfig(config AtmosphericConfig) {
	s.config = config
	s.shadowMapDirty = true
}

// clampValue restricts a value to [min, max].
func clampValue(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// GetLightCount returns number of active lights.
func (s *AtmosphericLightingSystem) GetLightCount() int {
	return len(s.lightBuffer)
}

// GetOccluderCount returns number of shadow casters.
func (s *AtmosphericLightingSystem) GetOccluderCount() int {
	return len(s.occluders)
}

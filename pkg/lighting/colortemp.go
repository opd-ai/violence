// Package lighting provides color temperature tinting for surfaces based on nearby light sources.
//
// This system implements realistic light color bleeding where warm lights (torches, fire)
// cast orange tints on nearby surfaces, and cool lights (magic, monitors) cast blue tints.
// The color temperature contribution uses inverse-square falloff and combines additively
// with multiple light sources to produce natural-looking color variation.
package lighting

import (
	"image/color"
	"math"
)

// ColorTemperature represents the warm/cool quality of a light source.
// Values range from -1.0 (cool blue) to +1.0 (warm orange).
type ColorTemperature float64

const (
	// TempCoolMagic is a cool magical blue-purple (-0.8)
	TempCoolMagic ColorTemperature = -0.8
	// TempCoolMonitor is a cool technological blue (-0.5)
	TempCoolMonitor ColorTemperature = -0.5
	// TempNeutral is neutral white (0.0)
	TempNeutral ColorTemperature = 0.0
	// TempWarmLamp is a slightly warm incandescent (+0.3)
	TempWarmLamp ColorTemperature = 0.3
	// TempWarmTorch is a warm torch orange (+0.6)
	TempWarmTorch ColorTemperature = 0.6
	// TempWarmFire is a hot fire red-orange (+0.9)
	TempWarmFire ColorTemperature = 0.9
)

// ColorTempConfig holds the configuration for color temperature tinting.
type ColorTempConfig struct {
	// Enabled controls whether temperature tinting is active
	Enabled bool
	// MaxTintStrength is the maximum color shift applied (0.0-1.0)
	MaxTintStrength float64
	// FalloffExponent controls how quickly tint fades with distance (2.0 = inverse-square)
	FalloffExponent float64
	// BlendMode determines how multiple lights combine ("additive" or "average")
	BlendMode string
}

// DefaultColorTempConfig returns sensible defaults for color temperature tinting.
func DefaultColorTempConfig() ColorTempConfig {
	return ColorTempConfig{
		Enabled:         true,
		MaxTintStrength: 0.35,
		FalloffExponent: 2.0,
		BlendMode:       "additive",
	}
}

// ColorTempLight represents a light source with color temperature for tinting calculations.
type ColorTempLight struct {
	X, Y        float64
	Radius      float64
	Intensity   float64
	Temperature ColorTemperature
	R, G, B     float64 // Base light color
}

// ColorTempSystem calculates and applies color temperature tinting to surfaces.
type ColorTempSystem struct {
	config ColorTempConfig
	lights []ColorTempLight
}

// NewColorTempSystem creates a color temperature tinting system.
func NewColorTempSystem(config ColorTempConfig) *ColorTempSystem {
	return &ColorTempSystem{
		config: config,
		lights: make([]ColorTempLight, 0, 32),
	}
}

// SetConfig updates the color temperature configuration.
func (s *ColorTempSystem) SetConfig(config ColorTempConfig) {
	s.config = config
}

// ClearLights removes all registered lights.
func (s *ColorTempSystem) ClearLights() {
	s.lights = s.lights[:0]
}

// AddLight registers a light source for temperature tinting.
func (s *ColorTempSystem) AddLight(light ColorTempLight) {
	s.lights = append(s.lights, light)
}

// AddLightFromPreset creates a ColorTempLight from a LightPreset with inferred temperature.
func (s *ColorTempSystem) AddLightFromPreset(x, y float64, preset LightPreset) {
	temp := InferTemperatureFromPreset(preset)
	s.lights = append(s.lights, ColorTempLight{
		X:           x,
		Y:           y,
		Radius:      preset.Radius,
		Intensity:   preset.Intensity,
		Temperature: temp,
		R:           preset.R,
		G:           preset.G,
		B:           preset.B,
	})
}

// InferTemperatureFromPreset determines color temperature from a light preset's color.
func InferTemperatureFromPreset(preset LightPreset) ColorTemperature {
	// Calculate warm/cool bias from RGB ratios
	// More red than blue = warm, more blue than red = cool
	warmBias := preset.R - preset.B

	// Scale to temperature range
	temp := ColorTemperature(warmBias)

	// Clamp to valid range
	if temp < -1.0 {
		temp = -1.0
	}
	if temp > 1.0 {
		temp = 1.0
	}

	return temp
}

// InferTemperatureFromName determines color temperature from light type name.
func InferTemperatureFromName(lightType string) ColorTemperature {
	switch lightType {
	// Warm sources
	case "torch", "brazier", "oil_lamp", "fire_barrel":
		return TempWarmTorch
	case "candle":
		return TempWarmLamp
	case "generator_light", "salvaged_lamp", "dim_bulb":
		return TempWarmLamp
	case "broken_lamp", "emergency_light":
		return TempWarmLamp

	// Cool sources
	case "magic_crystal":
		return TempCoolMagic
	case "monitor", "console", "hologram":
		return TempCoolMonitor
	case "neon_pink", "neon_cyan":
		return TempCoolMagic
	case "ceiling_lamp", "streetlight":
		return TempNeutral
	case "alarm":
		return TempWarmFire // Alarms are often red

	default:
		return TempNeutral
	}
}

// CalculateTintAtPosition computes the combined color temperature tint at a world position.
// Returns the tint color to be blended with surface colors.
func (s *ColorTempSystem) CalculateTintAtPosition(x, y float64) color.RGBA {
	if !s.config.Enabled || len(s.lights) == 0 {
		return color.RGBA{R: 128, G: 128, B: 128, A: 0} // Neutral, no tint
	}

	var totalR, totalG, totalB, totalWeight float64

	for _, light := range s.lights {
		// Calculate distance
		dx := x - light.X
		dy := y - light.Y
		distSq := dx*dx + dy*dy
		radiusSq := light.Radius * light.Radius

		// Skip if outside light radius
		if distSq > radiusSq {
			continue
		}

		// Calculate falloff
		dist := math.Sqrt(distSq)
		normalizedDist := dist / light.Radius
		falloff := math.Pow(1.0-normalizedDist, s.config.FalloffExponent)

		// Calculate contribution weight
		weight := light.Intensity * falloff

		// Get temperature-based tint color
		tintR, tintG, tintB := s.temperatureToRGB(light.Temperature)

		// Blend with light's base color for more accurate tinting
		tintR = (tintR + light.R) * 0.5
		tintG = (tintG + light.G) * 0.5
		tintB = (tintB + light.B) * 0.5

		totalR += tintR * weight
		totalG += tintG * weight
		totalB += tintB * weight
		totalWeight += weight
	}

	// No light contribution
	if totalWeight < 0.001 {
		return color.RGBA{R: 128, G: 128, B: 128, A: 0}
	}

	// Normalize based on blend mode
	var finalR, finalG, finalB float64
	if s.config.BlendMode == "average" {
		finalR = totalR / totalWeight
		finalG = totalG / totalWeight
		finalB = totalB / totalWeight
	} else {
		// Additive (clamped)
		finalR = math.Min(totalR, 1.0)
		finalG = math.Min(totalG, 1.0)
		finalB = math.Min(totalB, 1.0)
	}

	// Calculate alpha based on total weight (how much tinting to apply)
	alpha := math.Min(totalWeight*s.config.MaxTintStrength, 1.0)

	return color.RGBA{
		R: uint8(finalR * 255),
		G: uint8(finalG * 255),
		B: uint8(finalB * 255),
		A: uint8(alpha * 255),
	}
}

// temperatureToRGB converts a color temperature value to RGB components.
// Warm temperatures shift toward orange/red, cool temperatures toward blue.
func (s *ColorTempSystem) temperatureToRGB(temp ColorTemperature) (r, g, b float64) {
	t := float64(temp)

	if t > 0 {
		// Warm: shift toward orange (increase R, decrease B)
		r = 1.0
		g = 1.0 - t*0.3 // Slightly reduce green
		b = 1.0 - t*0.6 // Reduce blue more
	} else if t < 0 {
		// Cool: shift toward blue (decrease R, increase B)
		r = 1.0 + t*0.5 // Reduce red
		g = 1.0 + t*0.2 // Slightly reduce green
		b = 1.0
	} else {
		// Neutral white
		r, g, b = 1.0, 1.0, 1.0
	}

	// Clamp to valid range
	r = clampF(r, 0.0, 1.0)
	g = clampF(g, 0.0, 1.0)
	b = clampF(b, 0.0, 1.0)

	return r, g, b
}

// ApplyTintToColor blends a color temperature tint with a surface color.
// The tint alpha determines how much the temperature affects the surface.
func ApplyTintToColor(surface, tint color.RGBA) color.RGBA {
	if tint.A == 0 {
		return surface
	}

	// Calculate blend factor from tint alpha
	blend := float64(tint.A) / 255.0

	// Extract surface color
	sr := float64(surface.R) / 255.0
	sg := float64(surface.G) / 255.0
	sb := float64(surface.B) / 255.0

	// Extract tint color
	tr := float64(tint.R) / 255.0
	tg := float64(tint.G) / 255.0
	tb := float64(tint.B) / 255.0

	// Multiply blend (color grading style)
	// This preserves surface luminance while shifting hue toward tint
	finalR := sr*(1.0-blend) + sr*tr*blend
	finalG := sg*(1.0-blend) + sg*tg*blend
	finalB := sb*(1.0-blend) + sb*tb*blend

	return color.RGBA{
		R: uint8(clampF(finalR, 0.0, 1.0) * 255),
		G: uint8(clampF(finalG, 0.0, 1.0) * 255),
		B: uint8(clampF(finalB, 0.0, 1.0) * 255),
		A: surface.A,
	}
}

// ApplyTintToImage applies color temperature tinting to an entire image region.
// bounds specifies the world-space region covered by the image.
func (s *ColorTempSystem) ApplyTintToImage(pixels []byte, width, height int, worldX, worldY, scale float64) {
	if !s.config.Enabled {
		return
	}

	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			// Convert pixel position to world coordinates
			wx := worldX + float64(px)*scale
			wy := worldY + float64(py)*scale

			// Get tint at this position
			tint := s.CalculateTintAtPosition(wx, wy)

			// Skip if no tint
			if tint.A == 0 {
				continue
			}

			// Apply tint to pixel
			idx := (py*width + px) * 4
			if idx+3 >= len(pixels) {
				continue
			}

			surface := color.RGBA{
				R: pixels[idx],
				G: pixels[idx+1],
				B: pixels[idx+2],
				A: pixels[idx+3],
			}

			tinted := ApplyTintToColor(surface, tint)

			pixels[idx] = tinted.R
			pixels[idx+1] = tinted.G
			pixels[idx+2] = tinted.B
			// Preserve original alpha
		}
	}
}

// GetTemperatureForGenreLight returns the appropriate color temperature for a genre's light type.
func GetTemperatureForGenreLight(genreID, lightType string) ColorTemperature {
	// First try name-based inference
	temp := InferTemperatureFromName(lightType)
	if temp != TempNeutral {
		return temp
	}

	// Fall back to genre defaults
	switch genreID {
	case "fantasy":
		return TempWarmTorch // Fantasy defaults to warm torchlight
	case "scifi":
		return TempCoolMonitor // Sci-fi defaults to cool tech lighting
	case "horror":
		return TempWarmLamp // Horror uses dim warm lights
	case "cyberpunk":
		return TempCoolMagic // Cyberpunk uses cool neon
	case "postapoc":
		return TempWarmTorch // Post-apocalypse uses fire/oil lamps
	default:
		return TempNeutral
	}
}

// Note: clampF is defined in shadow.go within this package

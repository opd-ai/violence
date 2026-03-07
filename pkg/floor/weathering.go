package floor

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// WeatheringConfig controls the intensity and type of weathering applied to floor tiles.
type WeatheringConfig struct {
	EdgeDamage       float64 // 0.0-1.0: probability of edge chipping
	WearIntensity    float64 // 0.0-1.0: overall wear amount
	AgeVariation     float64 // 0.0-1.0: color variation from aging
	MoistureLevel    float64 // 0.0-1.0: water stains and dampness
	OrganicGrowth    float64 // 0.0-1.0: moss, algae, fungal growth
	ColorTemperature float64 // -1.0 to 1.0: warm (positive) to cool (negative) tint
}

// DefaultWeatheringConfig returns moderate weathering suitable for most environments.
func DefaultWeatheringConfig() WeatheringConfig {
	return WeatheringConfig{
		EdgeDamage:       0.3,
		WearIntensity:    0.4,
		AgeVariation:     0.5,
		MoistureLevel:    0.2,
		OrganicGrowth:    0.15,
		ColorTemperature: 0.0,
	}
}

// ApplyWeathering adds realistic aging and environmental effects to a floor tile texture.
func (g *TextureGenerator) ApplyWeathering(img *image.RGBA, material MaterialType, config WeatheringConfig, seed int64) {
	rng := rand.New(rand.NewSource(seed))
	size := img.Bounds().Dx()

	// Apply material-specific weathering in layers
	g.applyAgeVariation(img, size, material, config.AgeVariation, rng)
	g.applyEdgeDamage(img, size, material, config.EdgeDamage, rng)
	g.applyWearPatterns(img, size, material, config.WearIntensity, rng)
	g.applyMoistureEffects(img, size, material, config.MoistureLevel, rng)
	g.applyOrganicGrowth(img, size, material, config.OrganicGrowth, rng)
	g.applyColorTemperature(img, size, config.ColorTemperature)
}

// applyAgeVariation adds subtle color shifts from aging and oxidation.
func (g *TextureGenerator) applyAgeVariation(img *image.RGBA, size int, material MaterialType, intensity float64, rng *rand.Rand) {
	if intensity <= 0 {
		return
	}

	// Material-specific aging colors
	var ageColor color.RGBA
	switch material {
	case MaterialMetal:
		// Oxidation: rust tones
		ageColor = color.RGBA{R: 140, G: 80, B: 50, A: 255}
	case MaterialStone:
		// Weathering: darker, cooler tones
		ageColor = color.RGBA{R: 90, G: 95, B: 100, A: 255}
	case MaterialWood:
		// Aging: grayer, less saturation
		ageColor = color.RGBA{R: 110, G: 105, B: 95, A: 255}
	case MaterialConcrete:
		// Staining: yellowish-gray
		ageColor = color.RGBA{R: 130, G: 130, B: 110, A: 255}
	default:
		ageColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}

	// Apply age patches with perlin-like noise
	patchCount := int(intensity * 8)
	for i := 0; i < patchCount; i++ {
		centerX := rng.Intn(size)
		centerY := rng.Intn(size)
		radius := 4.0 + rng.Float64()*8.0
		patchIntensity := intensity * (0.5 + rng.Float64()*0.5)

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < radius {
					falloff := 1.0 - (dist / radius)
					// Irregular, organic falloff
					falloff *= (0.7 + rng.Float64()*0.3)
					alpha := falloff * patchIntensity * 0.4

					if alpha > 0 {
						current := img.At(x, y).(color.RGBA)
						r := uint8(lerp(float64(current.R), float64(ageColor.R), alpha))
						g := uint8(lerp(float64(current.G), float64(ageColor.G), alpha))
						b := uint8(lerp(float64(current.B), float64(ageColor.B), alpha))
						img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
					}
				}
			}
		}
	}
}

// applyEdgeDamage adds chipping and erosion along tile edges.
func (g *TextureGenerator) applyEdgeDamage(img *image.RGBA, size int, material MaterialType, intensity float64, rng *rand.Rand) {
	if intensity <= 0 {
		return
	}

	// Determine damage characteristics by material
	damageDepth := 2
	if material == MaterialMetal || material == MaterialCrystal {
		damageDepth = 1 // Harder materials chip less
	} else if material == MaterialDirt || material == MaterialGrass {
		return // Soft materials don't chip
	}

	// Damage along edges
	edges := []struct {
		isHorizontal bool
		position     int
	}{
		{false, 0},        // Left edge
		{false, size - 1}, // Right edge
		{true, 0},         // Top edge
		{true, size - 1},  // Bottom edge
	}

	for _, edge := range edges {
		if rng.Float64() > intensity {
			continue
		}

		damageCount := int(intensity * float64(size) * 0.4)
		for i := 0; i < damageCount; i++ {
			var x, y int
			if edge.isHorizontal {
				x = rng.Intn(size)
				y = edge.position
			} else {
				x = edge.position
				y = rng.Intn(size)
			}

			// Create irregular damage
			depth := rng.Intn(damageDepth) + 1
			width := 1 + rng.Intn(3)

			damageColor := g.getDamageColor(img, x, y, material)

			for d := 0; d < depth; d++ {
				for w := 0; w < width; w++ {
					var px, py int
					if edge.isHorizontal {
						px = x + w - width/2
						py = y + d*(2*edge.position/(size-1)-1) // Inward from edge
					} else {
						px = x + d*(2*edge.position/(size-1)-1)
						py = y + w - width/2
					}

					if px >= 0 && px < size && py >= 0 && py < size {
						img.Set(px, py, damageColor)
					}
				}
			}
		}
	}
}

// applyWearPatterns adds high-traffic wear marks and scuffing.
func (g *TextureGenerator) applyWearPatterns(img *image.RGBA, size int, material MaterialType, intensity float64, rng *rand.Rand) {
	if intensity <= 0 {
		return
	}

	// Wear appears as lighter/darker patches depending on material
	wearBrightness := 1.15
	if material == MaterialMetal {
		wearBrightness = 1.25 // Polished from foot traffic
	} else if material == MaterialWood {
		wearBrightness = 0.85 // Darkened from oils and dirt
	}

	// Create 1-3 wear paths across the tile
	pathCount := 1 + rng.Intn(3)
	for i := 0; i < pathCount; i++ {
		startX := rng.Intn(size)
		startY := rng.Intn(size)
		angle := rng.Float64() * math.Pi * 2
		length := float64(size) * (0.4 + rng.Float64()*0.5)
		width := 3.0 + rng.Float64()*4.0

		for t := 0.0; t < 1.0; t += 0.05 {
			centerX := startX + int(math.Cos(angle)*length*t)
			centerY := startY + int(math.Sin(angle)*length*t)

			// Wander slightly for organic path
			centerX += rng.Intn(3) - 1
			centerY += rng.Intn(3) - 1

			if centerX < 0 || centerX >= size || centerY < 0 || centerY >= size {
				continue
			}

			// Apply wear in radius around path center
			for dy := -int(width); dy <= int(width); dy++ {
				for dx := -int(width); dx <= int(width); dx++ {
					px := centerX + dx
					py := centerY + dy
					if px < 0 || px >= size || py < 0 || py >= size {
						continue
					}

					dist := math.Sqrt(float64(dx*dx + dy*dy))
					if dist > width {
						continue
					}

					falloff := 1.0 - (dist / width)
					wearStrength := intensity * falloff * (0.5 + rng.Float64()*0.5)

					if wearStrength > 0.1 {
						current := img.At(px, py).(color.RGBA)
						r := uint8(clamp(float64(current.R) * lerp(1.0, wearBrightness, wearStrength)))
						g := uint8(clamp(float64(current.G) * lerp(1.0, wearBrightness, wearStrength)))
						b := uint8(clamp(float64(current.B) * lerp(1.0, wearBrightness, wearStrength)))
						img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 255})
					}
				}
			}
		}
	}
}

// applyMoistureEffects adds water stains and dampness.
func (g *TextureGenerator) applyMoistureEffects(img *image.RGBA, size int, material MaterialType, intensity float64, rng *rand.Rand) {
	if intensity <= 0 {
		return
	}

	// Some materials don't show moisture
	if material == MaterialMetal || material == MaterialCrystal {
		intensity *= 0.3 // Reduced effect
	}

	// Add 1-3 moisture patches
	patchCount := 1 + rng.Intn(3)
	for i := 0; i < patchCount; i++ {
		centerX := rng.Intn(size)
		centerY := rng.Intn(size)
		radius := 5.0 + rng.Float64()*10.0

		// Moisture darkens and adds slight blue/green tint
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < radius {
					// Irregular moisture boundary
					falloff := 1.0 - (dist / radius)
					falloff *= (0.6 + rng.Float64()*0.4)
					moistStrength := intensity * falloff

					if moistStrength > 0.05 {
						current := img.At(x, y).(color.RGBA)
						// Darken
						darkenFactor := 1.0 - (moistStrength * 0.3)
						// Add cool tint
						coolTint := moistStrength * 0.15
						r := uint8(clamp(float64(current.R) * darkenFactor))
						g := uint8(clamp(float64(current.G)*darkenFactor + 255*coolTint*0.5))
						b := uint8(clamp(float64(current.B)*darkenFactor + 255*coolTint))
						img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
					}
				}
			}
		}
	}
}

// applyOrganicGrowth adds moss, algae, and fungal patches to appropriate materials.
func (g *TextureGenerator) applyOrganicGrowth(img *image.RGBA, size int, material MaterialType, intensity float64, rng *rand.Rand) {
	if intensity <= 0 {
		return
	}

	// Growth only on natural materials in moist environments
	if material == MaterialMetal || material == MaterialCrystal {
		return
	}
	if material == MaterialFlesh {
		intensity *= 2.0 // Accelerated decay
	}

	growthCount := int(intensity * 5)
	for i := 0; i < growthCount; i++ {
		centerX := rng.Intn(size)
		centerY := rng.Intn(size)
		radius := 2.0 + rng.Float64()*5.0

		// Determine growth color based on material and random variation
		var growthColor color.RGBA
		switch rng.Intn(3) {
		case 0: // Green moss
			growthColor = color.RGBA{R: 60, G: 100 + uint8(rng.Intn(40)), B: 50, A: 255}
		case 1: // Brown fungus
			growthColor = color.RGBA{R: 80 + uint8(rng.Intn(30)), G: 70, B: 50, A: 255}
		case 2: // Dark mold
			growthColor = color.RGBA{R: 50, G: 55, B: 50, A: 255}
		}

		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < radius {
					// Very irregular growth pattern
					noise := rng.Float64()
					if noise < 0.4 { // Only 40% coverage within radius
						continue
					}

					falloff := 1.0 - (dist / radius)
					growthStrength := intensity * falloff * noise

					if growthStrength > 0.3 {
						current := img.At(x, y).(color.RGBA)
						alpha := math.Min(growthStrength, 0.7)
						r := uint8(lerp(float64(current.R), float64(growthColor.R), alpha))
						g := uint8(lerp(float64(current.G), float64(growthColor.G), alpha))
						b := uint8(lerp(float64(current.B), float64(growthColor.B), alpha))
						img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
					}
				}
			}
		}
	}
}

// applyColorTemperature shifts overall color temperature (warm/cool lighting).
func (g *TextureGenerator) applyColorTemperature(img *image.RGBA, size int, temperature float64) {
	if temperature == 0 {
		return
	}

	// Clamp temperature to valid range
	temperature = math.Max(-1.0, math.Min(1.0, temperature))

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			current := img.At(x, y).(color.RGBA)

			var r, g, b float64
			if temperature > 0 {
				// Warm: increase red, slightly increase green, decrease blue
				r = float64(current.R) * (1.0 + temperature*0.15)
				g = float64(current.G) * (1.0 + temperature*0.05)
				b = float64(current.B) * (1.0 - temperature*0.1)
			} else {
				// Cool: decrease red, slightly increase green, increase blue
				r = float64(current.R) * (1.0 + temperature*0.1)
				g = float64(current.G) * (1.0 - temperature*0.05)
				b = float64(current.B) * (1.0 - temperature*0.15)
			}

			img.Set(x, y, color.RGBA{
				R: uint8(clamp(r)),
				G: uint8(clamp(g)),
				B: uint8(clamp(b)),
				A: 255,
			})
		}
	}
}

// getDamageColor returns an appropriate color for damaged material.
func (g *TextureGenerator) getDamageColor(img *image.RGBA, x, y int, material MaterialType) color.RGBA {
	size := img.Bounds().Dx()
	if x < 0 || x >= size || y < 0 || y >= size {
		return color.RGBA{R: 80, G: 80, B: 80, A: 255}
	}

	current := img.At(x, y).(color.RGBA)

	switch material {
	case MaterialStone:
		// Exposed lighter stone beneath
		return color.RGBA{
			R: uint8(clamp(float64(current.R) * 1.2)),
			G: uint8(clamp(float64(current.G) * 1.2)),
			B: uint8(clamp(float64(current.B) * 1.2)),
			A: 255,
		}
	case MaterialMetal:
		// Rust or darker oxidation
		return color.RGBA{R: 100, G: 60, B: 40, A: 255}
	case MaterialWood:
		// Splintered darker wood
		return color.RGBA{
			R: uint8(clamp(float64(current.R) * 0.6)),
			G: uint8(clamp(float64(current.G) * 0.6)),
			B: uint8(clamp(float64(current.B) * 0.6)),
			A: 255,
		}
	case MaterialConcrete:
		// Exposed aggregate (lighter)
		return color.RGBA{
			R: uint8(clamp(float64(current.R) * 1.15)),
			G: uint8(clamp(float64(current.G) * 1.15)),
			B: uint8(clamp(float64(current.B) * 1.15)),
			A: 255,
		}
	default:
		// Generic: slightly darker
		return color.RGBA{
			R: uint8(clamp(float64(current.R) * 0.8)),
			G: uint8(clamp(float64(current.G) * 0.8)),
			B: uint8(clamp(float64(current.B) * 0.8)),
			A: 255,
		}
	}
}

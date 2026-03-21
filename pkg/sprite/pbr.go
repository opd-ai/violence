// Package sprite provides physically-based rendering for procedural sprites.
//
// This enhancement implements a PBR (Physically-Based Rendering) shading system
// that dramatically improves the visual realism of all procedurally generated sprites.
//
// Key improvements:
//   - Light-source-consistent shading: All sprites share a consistent top-left light direction
//   - Specular highlights: Metallic surfaces show realistic reflective highlights
//   - Ambient occlusion: Contact points and edges are properly darkened
//   - Multi-tone shading: Each surface uses 3+ tonal steps for depth perception
//   - Material-specific properties: Metals, cloth, leather, scales, fur, and crystals
//     each have distinct reflection behaviors (metallic vs. diffuse)
//   - Hemisphere/cylindrical normals: Body parts use appropriate geometry for lighting
//
// The system is integrated into sprite generation and runs automatically.
// All enemy sprites (humanoid, quadruped, insect, serpent, flying, amorphous)
// and props (barrels, crates, pillars) now render with realistic shading.
package sprite

import (
	"image"
	"image/color"
	"math"
)

// LightConfig defines the lighting environment for sprite rendering.
type LightConfig struct {
	// Light direction (normalized vector)
	LightDirX, LightDirY, LightDirZ float64
	// Light color and intensity
	LightColor     color.RGBA
	LightIntensity float64
	// Ambient light level
	AmbientLevel float64
	// Ambient occlusion strength
	AOStrength float64
}

// DefaultLightConfig returns a standard top-left lighting setup.
func DefaultLightConfig() LightConfig {
	return LightConfig{
		LightDirX:      -0.5773, // Normalized (-1, -1, 1) for top-left-front light
		LightDirY:      -0.5773,
		LightDirZ:      0.5773,
		LightColor:     color.RGBA{R: 255, G: 250, B: 240, A: 255}, // Warm white
		LightIntensity: 1.0,
		AmbientLevel:   0.3,
		AOStrength:     0.6,
	}
}

// MaterialProperties defines surface properties for PBR shading.
type MaterialProperties struct {
	// Base surface color
	BaseColor color.RGBA
	// Metallic factor (0.0 = dielectric, 1.0 = metallic)
	Metallic float64
	// Roughness factor (0.0 = smooth/glossy, 1.0 = rough/matte)
	Roughness float64
	// Specular intensity for dielectrics
	Specular float64
}

// DefaultMaterialProperties returns standard material properties for different types.
func DefaultMaterialProperties(materialType MaterialDetail, baseColor color.RGBA) MaterialProperties {
	switch materialType {
	case MaterialMetal:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.9,
			Roughness: 0.3,
			Specular:  1.0,
		}
	case MaterialCloth:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.0,
			Roughness: 0.85,
			Specular:  0.2,
		}
	case MaterialLeather:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.0,
			Roughness: 0.65,
			Specular:  0.3,
		}
	case MaterialScales, MaterialChitin:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.1,
			Roughness: 0.4,
			Specular:  0.6,
		}
	case MaterialFur:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.0,
			Roughness: 0.95,
			Specular:  0.1,
		}
	case MaterialCrystal:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.0,
			Roughness: 0.05,
			Specular:  1.2,
		}
	case MaterialSlime:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.0,
			Roughness: 0.2,
			Specular:  0.8,
		}
	default:
		return MaterialProperties{
			BaseColor: baseColor,
			Metallic:  0.0,
			Roughness: 0.5,
			Specular:  0.5,
		}
	}
}

// ShadingContext holds per-pixel geometry information for shading calculations.
type ShadingContext struct {
	// Surface normal (normalized)
	NormalX, NormalY, NormalZ float64
	// Position relative to sprite center
	PosX, PosY float64
	// Ambient occlusion factor (0.0 = fully occluded, 1.0 = no occlusion)
	AO float64
}

// CalculateSurfaceNormal estimates surface normal from sprite geometry.
// For top-down sprites, we assume a sphere-like normal distribution.
func CalculateSurfaceNormal(dx, dy, radius float64) (float64, float64, float64) {
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > radius || radius == 0 {
		// Outside radius or invalid - use upward normal
		return 0, 0, 1
	}

	// Normalize to unit circle
	nx := dx / radius
	ny := dy / radius

	// Calculate Z component for hemisphere
	nz := math.Sqrt(math.Max(0, 1.0-nx*nx-ny*ny))

	// Normalize (should already be normalized, but ensure it)
	length := math.Sqrt(nx*nx + ny*ny + nz*nz)
	if length > 0 {
		nx /= length
		ny /= length
		nz /= length
	}

	return nx, ny, nz
}

// CalculateCylindricalNormal estimates normal for cylindrical shapes (limbs).
func CalculateCylindricalNormal(dx, axis, radius float64) (float64, float64, float64) {
	if radius == 0 {
		return 0, 0, 1
	}

	dist := math.Abs(dx)
	if dist > radius {
		// Outside cylinder
		if dx < 0 {
			return -1, 0, 0
		}
		return 1, 0, 0
	}

	// Normalize to unit cylinder
	nx := dx / radius

	// Calculate Z component for cylinder cross-section
	nz := math.Sqrt(math.Max(0, 1.0-nx*nx))

	// Normalize
	length := math.Sqrt(nx*nx + nz*nz)
	if length > 0 {
		nx /= length
		nz /= length
	}

	return nx, 0, nz
}

// CalculatePlanarNormal returns normal for flat surfaces with optional tilt.
func CalculatePlanarNormal(tiltX, tiltY float64) (float64, float64, float64) {
	// Start with upward normal
	nx := tiltX
	ny := tiltY
	nz := 1.0

	// Normalize
	length := math.Sqrt(nx*nx + ny*ny + nz*nz)
	if length > 0 {
		nx /= length
		ny /= length
		nz /= length
	}

	return nx, ny, nz
}

// CalculateAmbientOcclusion estimates occlusion based on distance to edges/contact points.
func CalculateAmbientOcclusion(dx, dy, radius float64, contactBottom bool) float64 {
	ao := 1.0

	// Distance-based occlusion (edges are darker)
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > radius*0.7 {
		edgeFactor := (dist - radius*0.7) / (radius * 0.3)
		ao -= edgeFactor * 0.3
	}

	// Contact point occlusion (bottom of sprites touch ground)
	if contactBottom && dy > radius*0.5 {
		contactFactor := (dy - radius*0.5) / (radius * 0.5)
		ao -= contactFactor * 0.4
	}

	return math.Max(0, math.Min(1, ao))
}

// ComputePBRShading calculates final pixel color using physically-based shading.
func ComputePBRShading(
	material MaterialProperties,
	context ShadingContext,
	light LightConfig,
) color.RGBA {
	// View direction (assume camera looking down for top-down game)
	viewX, viewY, viewZ := 0.0, 0.0, 1.0

	// Calculate diffuse lighting (Lambertian)
	NdotL := math.Max(0, context.NormalX*light.LightDirX+context.NormalY*light.LightDirY+context.NormalZ*light.LightDirZ)
	diffuse := NdotL * light.LightIntensity

	// Calculate specular lighting (Blinn-Phong)
	specular := 0.0
	if NdotL > 0 {
		// Half vector
		hx := light.LightDirX + viewX
		hy := light.LightDirY + viewY
		hz := light.LightDirZ + viewZ
		hlen := math.Sqrt(hx*hx + hy*hy + hz*hz)
		if hlen > 0 {
			hx /= hlen
			hy /= hlen
			hz /= hlen
		}

		// Specular intensity
		NdotH := math.Max(0, context.NormalX*hx+context.NormalY*hy+context.NormalZ*hz)

		// Roughness affects specular power (lower roughness = tighter highlight)
		specularPower := (1.0 - material.Roughness) * 100.0
		if specularPower < 2.0 {
			specularPower = 2.0
		}

		specular = math.Pow(NdotH, specularPower) * material.Specular * light.LightIntensity
	}

	// Fresnel effect (edges catch more specular light)
	NdotV := math.Max(0, context.NormalX*viewX+context.NormalY*viewY+context.NormalZ*viewZ)
	fresnel := math.Pow(1.0-NdotV, 5.0)
	specular += fresnel * material.Specular * 0.2

	// Combine ambient + diffuse + specular
	ambient := light.AmbientLevel

	// Apply ambient occlusion
	aoFactor := 1.0 - (1.0-context.AO)*light.AOStrength
	ambient *= aoFactor
	diffuse *= aoFactor

	// Total lighting
	totalLight := ambient + diffuse

	// Calculate final color
	var r, g, b float64

	if material.Metallic > 0.5 {
		// Metallic: base color affects specular color
		r = float64(material.BaseColor.R) * totalLight
		g = float64(material.BaseColor.G) * totalLight
		b = float64(material.BaseColor.B) * totalLight

		// Add colored specular
		r += float64(material.BaseColor.R) * specular * material.Metallic
		g += float64(material.BaseColor.G) * specular * material.Metallic
		b += float64(material.BaseColor.B) * specular * material.Metallic
	} else {
		// Dielectric: diffuse uses base color, specular is white-ish
		r = float64(material.BaseColor.R) * totalLight
		g = float64(material.BaseColor.G) * totalLight
		b = float64(material.BaseColor.B) * totalLight

		// Add white specular
		specularColor := float64(light.LightColor.R) * specular * (1.0 - material.Metallic)
		r += specularColor
		g += specularColor
		b += specularColor
	}

	// Clamp to valid range
	r = math.Max(0, math.Min(255, r))
	g = math.Max(0, math.Min(255, g))
	b = math.Max(0, math.Min(255, b))

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}
}

// ApplyPBRShadingToRegion applies physically-based shading to a sprite region.
func (g *Generator) ApplyPBRShadingToRegion(
	img *image.RGBA,
	bounds image.Rectangle,
	material MaterialDetail,
	geometryType string, // "spherical", "cylindrical", "planar"
	light LightConfig,
) {
	// Use the enhanced version with default seed
	g.ApplyPBRShadingToRegionWithSeed(img, bounds, material, geometryType, light, 0)
}

// ApplyPBRShadingToRegionWithSeed applies physically-based shading with normal perturbation.
// The seed parameter enables deterministic normal perturbation for micro-surface detail.
// When seed is 0, perturbation is disabled for backward compatibility.
func (g *Generator) ApplyPBRShadingToRegionWithSeed(
	img *image.RGBA,
	bounds image.Rectangle,
	material MaterialDetail,
	geometryType string,
	light LightConfig,
	seed int64,
) {
	cx := (bounds.Min.X + bounds.Max.X) / 2
	cy := (bounds.Min.Y + bounds.Max.Y) / 2
	radius := float64(bounds.Max.X-bounds.Min.X) / 2

	// Get perturbation config for this material (only if seed is provided)
	enablePerturbation := seed != 0
	var perturbCfg NormalPerturbConfig
	if enablePerturbation {
		perturbCfg = NormalPerturbForMaterial(material)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			// Skip transparent pixels
			existingColor := img.At(x, y)
			_, _, _, ea := existingColor.RGBA()
			if ea == 0 {
				continue
			}

			// Get base color
			er, eg, eb, _ := existingColor.RGBA()
			baseColor := color.RGBA{
				R: uint8(er >> 8),
				G: uint8(eg >> 8),
				B: uint8(eb >> 8),
				A: 255,
			}

			// Calculate relative position
			dx := float64(x - cx)
			dy := float64(y - cy)

			// Calculate surface normal based on geometry type
			var nx, ny, nz float64
			switch geometryType {
			case "spherical":
				nx, ny, nz = CalculateSurfaceNormal(dx, dy, radius)
			case "cylindrical":
				nx, ny, nz = CalculateCylindricalNormal(dx, 0, radius)
			case "planar":
				nx, ny, nz = CalculatePlanarNormal(0, 0)
			default:
				nx, ny, nz = CalculateSurfaceNormal(dx, dy, radius)
			}

			// Apply normal perturbation for micro-surface detail if enabled
			if enablePerturbation {
				nx, ny, nz = PerturbNormal(nx, ny, nz, x, y, material, seed, perturbCfg)
			}

			// Calculate ambient occlusion (assume bottom contact for Y > center)
			ao := CalculateAmbientOcclusion(dx, dy, radius, dy > 0)

			// Build shading context
			context := ShadingContext{
				NormalX: nx,
				NormalY: ny,
				NormalZ: nz,
				PosX:    dx,
				PosY:    dy,
				AO:      ao,
			}

			// Get material properties
			matProps := DefaultMaterialProperties(material, baseColor)

			// Compute shaded color
			shadedColor := ComputePBRShading(matProps, context, light)

			// Set pixel
			img.Set(x, y, shadedColor)
		}
	}
}

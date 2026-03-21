// Package sprite provides surface normal perturbation for enhanced material realism.
//
// The normal perturbation system adds micro-surface detail to PBR shading by
// perturbing surface normals based on material type. This creates the appearance
// of surface texture (scales, fur, cloth weave, metal scratches) in the lighting
// calculations, making materials look significantly more realistic.
//
// Unlike simple color variation, normal perturbation affects how light interacts
// with the surface - creating subtle shadows, highlights, and depth that match
// the expected properties of each material type.
package sprite

import (
	"image/color"
	"math"
	"math/rand"
)

// NormalPerturbConfig controls the strength and characteristics of normal perturbation.
type NormalPerturbConfig struct {
	// Intensity controls overall perturbation strength (0.0-1.0)
	Intensity float64
	// Scale controls the spatial frequency of detail (higher = smaller features)
	Scale float64
	// MaterialBias adjusts the perturbation direction preference per material
	MaterialBias float64
}

// DefaultNormalPerturbConfig returns standard perturbation settings.
func DefaultNormalPerturbConfig() NormalPerturbConfig {
	return NormalPerturbConfig{
		Intensity:    0.35,
		Scale:        1.0,
		MaterialBias: 0.0,
	}
}

// NormalPerturbForMaterial returns appropriate perturbation config for a material.
func NormalPerturbForMaterial(material MaterialDetail) NormalPerturbConfig {
	switch material {
	case MaterialScales:
		return NormalPerturbConfig{
			Intensity:    0.45, // Prominent scale edges
			Scale:        1.2,
			MaterialBias: 0.2,
		}
	case MaterialFur:
		return NormalPerturbConfig{
			Intensity:    0.30, // Soft fur texture
			Scale:        2.0,  // Fine strands
			MaterialBias: 0.3,  // Directional bias
		}
	case MaterialChitin:
		return NormalPerturbConfig{
			Intensity:    0.50, // Hard segments
			Scale:        0.8,
			MaterialBias: 0.1,
		}
	case MaterialMembrane:
		return NormalPerturbConfig{
			Intensity:    0.20, // Subtle veins
			Scale:        0.6,
			MaterialBias: 0.0,
		}
	case MaterialMetal:
		return NormalPerturbConfig{
			Intensity:    0.25, // Brushed texture
			Scale:        3.0,  // Fine scratches
			MaterialBias: 0.5,  // Anisotropic
		}
	case MaterialCloth:
		return NormalPerturbConfig{
			Intensity:    0.35, // Woven threads
			Scale:        2.5,
			MaterialBias: 0.0,
		}
	case MaterialLeather:
		return NormalPerturbConfig{
			Intensity:    0.30, // Grain texture
			Scale:        1.5,
			MaterialBias: 0.1,
		}
	case MaterialCrystal:
		return NormalPerturbConfig{
			Intensity:    0.40, // Sharp facets
			Scale:        0.5,  // Large facets
			MaterialBias: 0.0,
		}
	case MaterialSlime:
		return NormalPerturbConfig{
			Intensity:    0.15, // Smooth with ripples
			Scale:        0.8,
			MaterialBias: 0.0,
		}
	default:
		return DefaultNormalPerturbConfig()
	}
}

// PerturbNormal applies material-specific perturbation to a surface normal.
// Returns the perturbed normal (normalized).
func PerturbNormal(nx, ny, nz float64, x, y int, material MaterialDetail, seed int64, cfg NormalPerturbConfig) (float64, float64, float64) {
	if cfg.Intensity <= 0 {
		return nx, ny, nz
	}

	// Create deterministic RNG for this position
	posSeed := seed ^ (int64(x) * 73856093) ^ (int64(y) * 19349663)
	rng := rand.New(rand.NewSource(posSeed))

	// Get material-specific perturbation
	var perturbX, perturbY float64
	switch material {
	case MaterialScales:
		perturbX, perturbY = perturbScales(x, y, cfg.Scale, rng)
	case MaterialFur:
		perturbX, perturbY = perturbFur(x, y, cfg.Scale, cfg.MaterialBias, rng)
	case MaterialChitin:
		perturbX, perturbY = perturbChitin(x, y, cfg.Scale, rng)
	case MaterialMembrane:
		perturbX, perturbY = perturbMembrane(x, y, cfg.Scale, rng)
	case MaterialMetal:
		perturbX, perturbY = perturbMetal(x, y, cfg.Scale, cfg.MaterialBias, rng)
	case MaterialCloth:
		perturbX, perturbY = perturbCloth(x, y, cfg.Scale, rng)
	case MaterialLeather:
		perturbX, perturbY = perturbLeather(x, y, cfg.Scale, rng)
	case MaterialCrystal:
		perturbX, perturbY = perturbCrystal(x, y, cfg.Scale, rng)
	case MaterialSlime:
		perturbX, perturbY = perturbSlime(x, y, cfg.Scale, rng)
	default:
		perturbX, perturbY = perturbDefault(x, y, cfg.Scale, rng)
	}

	// Scale perturbation by intensity
	perturbX *= cfg.Intensity
	perturbY *= cfg.Intensity

	// Apply perturbation to normal (keeping it roughly unit length)
	newNx := nx + perturbX
	newNy := ny + perturbY
	// Reduce Z component slightly to compensate for tangent perturbation
	newNz := nz - math.Abs(perturbX+perturbY)*0.3

	// Renormalize
	length := math.Sqrt(newNx*newNx + newNy*newNy + newNz*newNz)
	if length > 0.001 {
		newNx /= length
		newNy /= length
		newNz /= length
	}

	return newNx, newNy, newNz
}

// perturbScales creates scale-edge normal perturbation.
func perturbScales(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	scaleSize := 4.0 / scale
	sx := float64(x) / scaleSize
	sy := float64(y) / scaleSize

	// Hexagonal scale pattern
	row := int(sy)
	offsetX := 0.0
	if row%2 == 1 {
		offsetX = 0.5
	}
	localX := math.Mod(sx+offsetX, 1.0)
	localY := math.Mod(sy, 1.0)

	// Perturbation points outward from scale center
	dx := localX - 0.5
	dy := localY - 0.5
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 0.01 {
		// Normal points outward at scale edges, inward at center
		edgeFactor := math.Min(1.0, dist*2.5)
		return dx * edgeFactor * 0.8, dy * edgeFactor * 0.8
	}
	return 0, 0
}

// perturbFur creates directional fur-strand normal perturbation.
func perturbFur(x, y int, scale, bias float64, rng *rand.Rand) (float64, float64) {
	// Fur strands aligned along a direction
	strandAngle := math.Pi/4 + bias*math.Pi/4
	perpX := math.Cos(strandAngle + math.Pi/2)
	perpY := math.Sin(strandAngle + math.Pi/2)

	perpDist := float64(x)*perpX + float64(y)*perpY
	strandPhase := math.Mod(perpDist*scale*0.5, 2.0)

	// Strands alternate normal direction
	perturbMag := (strandPhase - 1.0) * 0.6
	perturbMag += (rng.Float64() - 0.5) * 0.3 // Random variation

	return perpX * perturbMag, perpY * perturbMag
}

// perturbChitin creates segmented plate normal perturbation.
func perturbChitin(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	segmentSize := 6.0 / scale
	segmentY := math.Mod(float64(y), segmentSize)

	// Strong normal change at segment boundaries
	var perturbY float64
	if segmentY < 1.5 {
		// Upper edge - normal points up
		perturbY = -0.7 * (1.0 - segmentY/1.5)
	} else if segmentY > segmentSize-1.5 {
		// Lower edge - normal points down
		perturbY = 0.7 * (segmentY - (segmentSize - 1.5)) / 1.5
	}

	// Subtle X variation within plates
	perturbX := math.Sin(float64(x)*0.3*scale) * 0.15

	return perturbX, perturbY
}

// perturbMembrane creates vein-network normal perturbation.
func perturbMembrane(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	veinDensity := 8.0 / scale
	vx := math.Sin(float64(x) / veinDensity)
	vy := math.Cos(float64(y) / veinDensity)

	// Veins create subtle ridges
	veinPhase := vx*vy + 0.5
	perturbMag := math.Sin(veinPhase*math.Pi*2) * 0.4

	// Perpendicular to vein direction
	veinAngle := math.Atan2(vy, vx)
	perturbX := math.Cos(veinAngle+math.Pi/2) * perturbMag
	perturbY := math.Sin(veinAngle+math.Pi/2) * perturbMag

	return perturbX, perturbY
}

// perturbMetal creates anisotropic brushed-metal normal perturbation.
func perturbMetal(x, y int, scale, bias float64, rng *rand.Rand) (float64, float64) {
	// Anisotropic brush direction (mostly horizontal by default)
	brushAngle := bias * math.Pi / 2

	// Fine parallel scratches
	perpDist := float64(x)*math.Sin(brushAngle) - float64(y)*math.Cos(brushAngle)
	scratchPhase := math.Mod(perpDist*scale*0.8, 1.0)

	// Sharp scratch edges
	perturbMag := 0.0
	if scratchPhase < 0.3 {
		perturbMag = scratchPhase / 0.3 * 0.5
	} else if scratchPhase > 0.7 {
		perturbMag = -(1.0 - scratchPhase) / 0.3 * 0.5
	}

	// Add random micro-scratches
	perturbMag += (rng.Float64() - 0.5) * 0.2

	// Perpendicular to brush direction
	return math.Sin(brushAngle) * perturbMag, -math.Cos(brushAngle) * perturbMag
}

// perturbCloth creates woven-thread normal perturbation.
func perturbCloth(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	threadSize := 2.0 / scale
	warpPhase := math.Mod(float64(x), threadSize) / threadSize
	weftPhase := math.Mod(float64(y), threadSize) / threadSize

	var perturbX, perturbY float64

	// Warp threads (vertical) create X normal variation
	if warpPhase < 0.5 {
		perturbX = (0.5 - warpPhase) * 0.6
	} else {
		perturbX = (0.5 - warpPhase) * 0.6
	}

	// Weft threads (horizontal) create Y normal variation
	if weftPhase < 0.5 {
		perturbY = (0.5 - weftPhase) * 0.6
	} else {
		perturbY = (0.5 - weftPhase) * 0.6
	}

	// Over-under weave pattern
	if (warpPhase < 0.5) != (weftPhase < 0.5) {
		perturbX *= -1
		perturbY *= -1
	}

	return perturbX, perturbY
}

// perturbLeather creates organic grain normal perturbation.
func perturbLeather(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	// Organic grain pattern using multiple frequencies
	grainX := math.Sin(float64(x)*0.2*scale+rng.Float64()*6.28) * 0.3
	grainY := math.Cos(float64(y)*0.3*scale+rng.Float64()*6.28) * 0.3

	// Add pores (small depressions)
	poreX := math.Sin(float64(x)*0.8*scale) * math.Cos(float64(y)*0.9*scale) * 0.2
	poreY := math.Cos(float64(x)*0.7*scale) * math.Sin(float64(y)*0.6*scale) * 0.2

	return grainX + poreX, grainY + poreY
}

// perturbCrystal creates sharp facet normal perturbation.
func perturbCrystal(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	facetSize := 8.0 / scale
	fx := math.Floor(float64(x) / facetSize)
	fy := math.Floor(float64(y) / facetSize)

	// Each facet has a random tilt direction
	facetSeed := int64(fx)*73856093 ^ int64(fy)*19349663
	facetRng := rand.New(rand.NewSource(facetSeed))

	// Random facet normal tilt
	tiltAngle := facetRng.Float64() * math.Pi * 2
	tiltMag := 0.4 + facetRng.Float64()*0.3

	// Sharp edge at facet boundaries
	localX := math.Mod(float64(x), facetSize) / facetSize
	localY := math.Mod(float64(y), facetSize) / facetSize
	edgeDist := math.Min(math.Min(localX, 1-localX), math.Min(localY, 1-localY))

	if edgeDist < 0.15 {
		// At facet edge - strong normal change
		edgeFactor := (0.15 - edgeDist) / 0.15
		if localX < 0.15 {
			return -0.8 * edgeFactor, 0
		} else if localX > 0.85 {
			return 0.8 * edgeFactor, 0
		} else if localY < 0.15 {
			return 0, -0.8 * edgeFactor
		} else if localY > 0.85 {
			return 0, 0.8 * edgeFactor
		}
	}

	// Interior facet tilt
	return math.Cos(tiltAngle) * tiltMag, math.Sin(tiltAngle) * tiltMag
}

// perturbSlime creates smooth ripple normal perturbation.
func perturbSlime(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	// Concentric ripples
	rippleFreq := 0.15 * scale
	rippleX := float64(x) * rippleFreq
	rippleY := float64(y) * rippleFreq
	rippleDist := math.Sqrt(rippleX*rippleX + rippleY*rippleY)

	// Sine wave ripples
	ripplePhase := math.Sin(rippleDist * 2.0)
	rippleMag := ripplePhase * 0.3

	// Ripple normal points radially
	if rippleDist > 0.01 {
		return rippleX / rippleDist * rippleMag, rippleY / rippleDist * rippleMag
	}
	return 0, 0
}

// perturbDefault creates generic noise perturbation.
func perturbDefault(x, y int, scale float64, rng *rand.Rand) (float64, float64) {
	// Simple noise-based perturbation
	noiseX := math.Sin(float64(x)*0.3*scale) * 0.3
	noiseY := math.Cos(float64(y)*0.3*scale) * 0.3
	noiseX += (rng.Float64() - 0.5) * 0.2
	noiseY += (rng.Float64() - 0.5) * 0.2
	return noiseX, noiseY
}

// ComputePBRShadingWithPerturbation calculates shading with normal perturbation.
// This is the enhanced version of ComputePBRShading that applies material-specific
// micro-surface detail to create more realistic lighting.
func ComputePBRShadingWithPerturbation(
	material MaterialProperties,
	context ShadingContext,
	light LightConfig,
	materialType MaterialDetail,
	x, y int,
	seed int64,
) color.RGBA {
	// Get perturbation config for this material
	perturbCfg := NormalPerturbForMaterial(materialType)

	// Perturb the normal based on material type
	pnx, pny, pnz := PerturbNormal(
		context.NormalX, context.NormalY, context.NormalZ,
		x, y, materialType, seed, perturbCfg,
	)

	// Create modified context with perturbed normal
	perturbedContext := ShadingContext{
		NormalX: pnx,
		NormalY: pny,
		NormalZ: pnz,
		PosX:    context.PosX,
		PosY:    context.PosY,
		AO:      context.AO,
	}

	// Use standard PBR calculation with perturbed normal
	return ComputePBRShading(material, perturbedContext, light)
}

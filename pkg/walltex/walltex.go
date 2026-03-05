// Package walltex provides enhanced procedural wall texture generation with material-specific rendering.
package walltex

import (
	"image"
	"image/color"
	"math"

	"github.com/opd-ai/violence/pkg/rng"
)

// Material defines the physical material of a wall surface.
type Material int

const (
	MaterialStone Material = iota
	MaterialMetal
	MaterialWood
	MaterialConcrete
	MaterialOrganic
	MaterialCrystal
	MaterialTech
)

// GenrePreset defines material distribution and weathering for each genre.
type GenrePreset struct {
	PrimaryMaterial   Material
	SecondaryMaterial Material
	WeatherIntensity  float64 // 0.0-1.0, how much wear/damage to show
	DetailDensity     float64 // 0.0-1.0, density of cracks/stains/etc
	ColorVariation    float64 // 0.0-1.0, how much color varies within texture
	GlowIntensity     float64 // 0.0-1.0, for sci-fi/cyberpunk lighting effects
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		PrimaryMaterial:   MaterialStone,
		SecondaryMaterial: MaterialWood,
		WeatherIntensity:  0.6,
		DetailDensity:     0.5,
		ColorVariation:    0.3,
		GlowIntensity:     0.0,
	},
	"scifi": {
		PrimaryMaterial:   MaterialMetal,
		SecondaryMaterial: MaterialTech,
		WeatherIntensity:  0.2,
		DetailDensity:     0.4,
		ColorVariation:    0.2,
		GlowIntensity:     0.4,
	},
	"horror": {
		PrimaryMaterial:   MaterialWood,
		SecondaryMaterial: MaterialOrganic,
		WeatherIntensity:  0.8,
		DetailDensity:     0.7,
		ColorVariation:    0.4,
		GlowIntensity:     0.1,
	},
	"cyberpunk": {
		PrimaryMaterial:   MaterialConcrete,
		SecondaryMaterial: MaterialTech,
		WeatherIntensity:  0.5,
		DetailDensity:     0.6,
		ColorVariation:    0.3,
		GlowIntensity:     0.7,
	},
	"postapoc": {
		PrimaryMaterial:   MaterialConcrete,
		SecondaryMaterial: MaterialMetal,
		WeatherIntensity:  0.9,
		DetailDensity:     0.8,
		ColorVariation:    0.5,
		GlowIntensity:     0.0,
	},
}

// Generator creates enhanced wall textures with material-specific rendering.
type Generator struct {
	genre  string
	preset GenrePreset
}

// NewGenerator creates a wall texture generator for the specified genre.
func NewGenerator(genre string) *Generator {
	preset, ok := genrePresets[genre]
	if !ok {
		preset = genrePresets["fantasy"]
	}
	return &Generator{
		genre:  genre,
		preset: preset,
	}
}

// Generate creates a wall texture with the specified parameters.
// variant (0-3) determines which material mix to use.
// seed provides deterministic randomization.
func (g *Generator) Generate(size, variant int, seed uint64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	r := rng.NewRNG(seed)

	// Choose material based on variant
	material := g.preset.PrimaryMaterial
	if variant%2 == 1 {
		material = g.preset.SecondaryMaterial
	}

	// Generate base material pattern
	g.generateBaseMaterial(img, material, r)

	// Add depth and lighting
	g.applyNormalMapping(img, material, r)

	// Add weathering and detail
	g.applyWeathering(img, r)
	g.applyDetails(img, material, r)

	// Add genre-specific effects
	if g.preset.GlowIntensity > 0 {
		g.applyGlow(img, r)
	}

	return img
}

// GenerateWithMaterial creates a wall texture with explicit material and weathering control.
// This method is used by the WallTextureSystem for dynamic texture variation.
func (g *Generator) GenerateWithMaterial(size int, material Material, variant int, weathering float64, seed uint64) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	r := rng.NewRNG(seed)

	// Generate base material pattern
	g.generateBaseMaterial(img, material, r)

	// Add depth and lighting
	g.applyNormalMapping(img, material, r)

	// Apply custom weathering intensity
	if weathering > 0.1 {
		g.applyCustomWeathering(img, weathering, r)
	}

	// Add material-specific details
	g.applyDetails(img, material, r)

	// Add genre-specific effects (glow for tech materials)
	if (material == MaterialTech || material == MaterialCrystal) && g.preset.GlowIntensity > 0 {
		g.applyGlow(img, r)
	}

	return img
}

// applyCustomWeathering applies weathering with custom intensity.
func (g *Generator) applyCustomWeathering(img *image.RGBA, intensity float64, r *rng.RNG) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Scale crack and stain count by intensity
	numCracks := int(float64(w/16) * intensity)
	for i := 0; i < numCracks; i++ {
		crackSeed := uint64(i * 7919)
		startX := int(hashSeed(crackSeed) % uint64(w))
		startY := int(hashSeed(crackSeed+1) % uint64(h))
		length := 5 + int(hashSeed(crackSeed+2)%15)
		angle := float64(hashSeed(crackSeed+3)%360) * math.Pi / 180.0
		g.drawCrack(img, startX, startY, length, angle)
	}

	numStains := int(float64(w/32) * intensity)
	for i := 0; i < numStains; i++ {
		stainSeed := uint64(i*12347 + 999)
		stainX := int(hashSeed(stainSeed) % uint64(w))
		stainY := int(hashSeed(stainSeed+1) % uint64(h))
		stainSize := 3 + int(hashSeed(stainSeed+2)%8)
		g.drawStain(img, stainX, stainY, stainSize)
	}
}

// generateBaseMaterial creates the base material pattern.
func (g *Generator) generateBaseMaterial(img *image.RGBA, material Material, r *rng.RNG) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var col color.RGBA

			switch material {
			case MaterialStone:
				col = g.renderStone(x, y, w, h, r)
			case MaterialMetal:
				col = g.renderMetal(x, y, w, h, r)
			case MaterialWood:
				col = g.renderWood(x, y, w, h, r)
			case MaterialConcrete:
				col = g.renderConcrete(x, y, w, h, r)
			case MaterialOrganic:
				col = g.renderOrganic(x, y, w, h, r)
			case MaterialCrystal:
				col = g.renderCrystal(x, y, w, h, r)
			case MaterialTech:
				col = g.renderTech(x, y, w, h, r)
			default:
				col = g.renderStone(x, y, w, h, r)
			}

			img.Set(x, y, col)
		}
	}
}

// renderStone creates stone brick texture with mortar lines.
func (g *Generator) renderStone(x, y, w, h int, r *rng.RNG) color.RGBA {
	// Base stone color with variation
	baseR := 100 + int(r.Float64()*30*g.preset.ColorVariation)
	baseG := 90 + int(r.Float64()*25*g.preset.ColorVariation)
	baseB := 70 + int(r.Float64()*20*g.preset.ColorVariation)

	// Brick pattern (offset every other row)
	brickW, brickH := 32, 16
	offsetRow := (y / brickH) % 2
	brickX := x
	if offsetRow == 1 {
		brickX = (x + brickW/2) % w
	}

	// Mortar lines (darker)
	mortarThickness := 2
	atMortarH := (y % brickH) < mortarThickness
	atMortarV := (brickX % brickW) < mortarThickness

	if atMortarH || atMortarV {
		return color.RGBA{
			R: uint8(baseR / 2),
			G: uint8(baseG / 2),
			B: uint8(baseB / 2),
			A: 255,
		}
	}

	// Multi-octave noise for surface detail
	noise := g.octaveNoise(float64(x), float64(y), 3, r) * 0.3
	// Add some larger variation per brick
	brickSeed := uint64((x/brickW)*1000 + (y / brickH))
	brickNoise := float64(hashSeed(brickSeed)%100) / 100.0 * 0.2

	factor := 1.0 + noise + brickNoise

	return color.RGBA{
		R: clamp(int(float64(baseR) * factor)),
		G: clamp(int(float64(baseG) * factor)),
		B: clamp(int(float64(baseB) * factor)),
		A: 255,
	}
}

// renderMetal creates metal plating texture with panel seams.
func (g *Generator) renderMetal(x, y, w, h int, r *rng.RNG) color.RGBA {
	// Base metal color (bluish grey)
	baseR := 80 + int(r.Float64()*20*g.preset.ColorVariation)
	baseG := 90 + int(r.Float64()*20*g.preset.ColorVariation)
	baseB := 110 + int(r.Float64()*25*g.preset.ColorVariation)

	// Panel seams
	panelW, panelH := 48, 48
	seamW := 2
	atSeamH := (y % panelH) < seamW
	atSeamV := (x % panelW) < seamW

	if atSeamH || atSeamV {
		return color.RGBA{
			R: uint8(baseR / 3),
			G: uint8(baseG / 3),
			B: uint8(baseB / 3),
			A: 255,
		}
	}

	// Fine-grain metal texture
	noise := g.octaveNoise(float64(x), float64(y), 4, r) * 0.15

	// Brushed metal effect (horizontal striations)
	brushNoise := g.octaveNoise(float64(x)*0.1, float64(y)*2, 2, r) * 0.1

	factor := 1.0 + noise + brushNoise

	return color.RGBA{
		R: clamp(int(float64(baseR) * factor)),
		G: clamp(int(float64(baseG) * factor)),
		B: clamp(int(float64(baseB) * factor)),
		A: 255,
	}
}

// renderWood creates wood plank texture with grain.
func (g *Generator) renderWood(x, y, w, h int, r *rng.RNG) color.RGBA {
	// Base wood color (brown)
	baseR := 100 + int(r.Float64()*40*g.preset.ColorVariation)
	baseG := 70 + int(r.Float64()*30*g.preset.ColorVariation)
	baseB := 40 + int(r.Float64()*20*g.preset.ColorVariation)

	// Horizontal planks
	plankH := 24
	plankGap := 1
	atGap := (y % plankH) < plankGap
	if atGap {
		return color.RGBA{R: 30, G: 25, B: 15, A: 255}
	}

	// Wood grain (horizontal wavy lines)
	plankY := y / plankH
	plankSeed := uint64(plankY * 12345)
	grainPhase := float64(hashSeed(plankSeed)%360) * math.Pi / 180.0
	grainWave := math.Sin(float64(x)*0.3+grainPhase) * 0.15

	// Fine grain detail
	grainNoise := g.octaveNoise(float64(x)*0.5, float64(y%(plankH))*2, 3, r) * 0.2

	factor := 1.0 + grainWave + grainNoise

	return color.RGBA{
		R: clamp(int(float64(baseR) * factor)),
		G: clamp(int(float64(baseG) * factor)),
		B: clamp(int(float64(baseB) * factor)),
		A: 255,
	}
}

// renderConcrete creates concrete texture with subtle variation.
func (g *Generator) renderConcrete(x, y, w, h int, r *rng.RNG) color.RGBA {
	// Base concrete color (grey)
	baseR := 100 + int(r.Float64()*30*g.preset.ColorVariation)
	baseG := 100 + int(r.Float64()*30*g.preset.ColorVariation)
	baseB := 95 + int(r.Float64()*25*g.preset.ColorVariation)

	// Concrete has larger patches of color variation
	patchNoise := g.octaveNoise(float64(x)*0.05, float64(y)*0.05, 2, r) * 0.25
	// Fine aggregate texture
	fineNoise := g.octaveNoise(float64(x), float64(y), 4, r) * 0.15

	factor := 1.0 + patchNoise + fineNoise

	return color.RGBA{
		R: clamp(int(float64(baseR) * factor)),
		G: clamp(int(float64(baseG) * factor)),
		B: clamp(int(float64(baseB) * factor)),
		A: 255,
	}
}

// renderOrganic creates organic wall texture (flesh, vines, etc).
func (g *Generator) renderOrganic(x, y, w, h int, r *rng.RNG) color.RGBA {
	// Base organic color (reddish-brown or greenish)
	variant := hashSeed(uint64(x/16)) % 2
	var baseR, baseG, baseB int
	if variant == 0 {
		// Fleshy
		baseR = 120 + int(r.Float64()*30*g.preset.ColorVariation)
		baseG = 60 + int(r.Float64()*20*g.preset.ColorVariation)
		baseB = 60 + int(r.Float64()*20*g.preset.ColorVariation)
	} else {
		// Vine-covered
		baseR = 60 + int(r.Float64()*20*g.preset.ColorVariation)
		baseG = 100 + int(r.Float64()*30*g.preset.ColorVariation)
		baseB = 50 + int(r.Float64()*20*g.preset.ColorVariation)
	}

	// Organic irregular patterns
	noise1 := g.octaveNoise(float64(x)*0.3, float64(y)*0.3, 3, r) * 0.4
	noise2 := g.octaveNoise(float64(x)*0.15, float64(y)*0.15, 2, r) * 0.3

	factor := 1.0 + noise1 + noise2

	return color.RGBA{
		R: clamp(int(float64(baseR) * factor)),
		G: clamp(int(float64(baseG) * factor)),
		B: clamp(int(float64(baseB) * factor)),
		A: 255,
	}
}

// renderCrystal creates crystalline structure texture.
func (g *Generator) renderCrystal(x, y, w, h int, r *rng.RNG) color.RGBA {
	// Base crystal color (bluish or purplish)
	baseR := 100 + int(r.Float64()*40*g.preset.ColorVariation)
	baseG := 120 + int(r.Float64()*40*g.preset.ColorVariation)
	baseB := 150 + int(r.Float64()*50*g.preset.ColorVariation)

	// Angular facets
	facetSize := 16
	facetX := x / facetSize
	facetY := y / facetSize
	facetSeed := uint64(facetX*1000 + facetY)
	facetBrightness := float64(hashSeed(facetSeed)%100) / 100.0

	// Internal structure noise
	noise := g.octaveNoise(float64(x)*0.5, float64(y)*0.5, 3, r) * 0.2

	factor := 0.6 + facetBrightness*0.8 + noise

	return color.RGBA{
		R: clamp(int(float64(baseR) * factor)),
		G: clamp(int(float64(baseG) * factor)),
		B: clamp(int(float64(baseB) * factor)),
		A: 255,
	}
}

// renderTech creates high-tech paneling with circuit-like patterns.
func (g *Generator) renderTech(x, y, w, h int, r *rng.RNG) color.RGBA {
	// Base tech color (dark with slight blue tint)
	baseR := 40 + int(r.Float64()*20*g.preset.ColorVariation)
	baseG := 45 + int(r.Float64()*20*g.preset.ColorVariation)
	baseB := 55 + int(r.Float64()*25*g.preset.ColorVariation)

	// Panel grid
	panelSize := 32
	gridW := 1
	atGridH := (y % panelSize) < gridW
	atGridV := (x % panelSize) < gridW

	// Circuit traces (occasional bright lines)
	circuitSeed := hashSeed(uint64((x/8)*100 + (y / 8)))
	hasCircuit := (circuitSeed % 10) < 2

	if (atGridH || atGridV) && hasCircuit {
		return color.RGBA{R: 80, G: 150, B: 180, A: 255}
	} else if atGridH || atGridV {
		return color.RGBA{R: 60, G: 65, B: 75, A: 255}
	}

	// Fine surface detail
	noise := g.octaveNoise(float64(x), float64(y), 3, r) * 0.1

	factor := 1.0 + noise

	return color.RGBA{
		R: clamp(int(float64(baseR) * factor)),
		G: clamp(int(float64(baseG) * factor)),
		B: clamp(int(float64(baseB) * factor)),
		A: 255,
	}
}

// applyNormalMapping simulates lighting from a directional light source.
func (g *Generator) applyNormalMapping(img *image.RGBA, material Material, r *rng.RNG) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Light direction (from top-left)
	lightX, lightY := -0.5, -0.7

	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			// Sample neighboring pixels to estimate normal
			center := img.At(x, y)
			right := img.At(x+1, y)
			down := img.At(x, y+1)

			centerL := luminance(center)
			rightL := luminance(right)
			downL := luminance(down)

			// Gradient gives us surface normal direction
			dx := rightL - centerL
			dy := downL - centerL

			// Dot product with light direction
			dot := dx*lightX + dy*lightY

			// Apply subtle lighting adjustment
			adjustment := dot * 0.15

			cr, cg, cb, ca := center.RGBA()
			newR := clamp(int(float64(cr>>8) * (1.0 + adjustment)))
			newG := clamp(int(float64(cg>>8) * (1.0 + adjustment)))
			newB := clamp(int(float64(cb>>8) * (1.0 + adjustment)))

			img.Set(x, y, color.RGBA{
				R: uint8(newR),
				G: uint8(newG),
				B: uint8(newB),
				A: uint8(ca >> 8),
			})
		}
	}
}

// applyWeathering adds cracks, stains, and wear based on genre preset.
func (g *Generator) applyWeathering(img *image.RGBA, r *rng.RNG) {
	if g.preset.WeatherIntensity < 0.1 {
		return
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Cracks (thin dark lines)
	numCracks := int(float64(w/16) * g.preset.WeatherIntensity)
	for i := 0; i < numCracks; i++ {
		crackSeed := uint64(i * 7919)
		startX := int(hashSeed(crackSeed) % uint64(w))
		startY := int(hashSeed(crackSeed+1) % uint64(h))
		length := 5 + int(hashSeed(crackSeed+2)%15)
		angle := float64(hashSeed(crackSeed+3)%360) * math.Pi / 180.0

		g.drawCrack(img, startX, startY, length, angle)
	}

	// Stains (darker patches)
	numStains := int(float64(w/32) * g.preset.WeatherIntensity)
	for i := 0; i < numStains; i++ {
		stainSeed := uint64(i*12347 + 999)
		stainX := int(hashSeed(stainSeed) % uint64(w))
		stainY := int(hashSeed(stainSeed+1) % uint64(h))
		stainSize := 3 + int(hashSeed(stainSeed+2)%8)

		g.drawStain(img, stainX, stainY, stainSize)
	}
}

// drawCrack renders a crack line on the texture.
func (g *Generator) drawCrack(img *image.RGBA, x, y, length int, angle float64) {
	dx := math.Cos(angle)
	dy := math.Sin(angle)

	for i := 0; i < length; i++ {
		px := x + int(float64(i)*dx)
		py := y + int(float64(i)*dy)

		if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
			existing := img.At(px, py)
			r, g, b, a := existing.RGBA()
			img.Set(px, py, color.RGBA{
				R: uint8((r >> 8) * 6 / 10),
				G: uint8((g >> 8) * 6 / 10),
				B: uint8((b >> 8) * 6 / 10),
				A: uint8(a >> 8),
			})
		}
	}
}

// drawStain renders a circular stain on the texture.
func (g *Generator) drawStain(img *image.RGBA, cx, cy, radius int) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			dx := x - cx
			dy := y - cy
			dist := math.Sqrt(float64(dx*dx + dy*dy))

			if dist <= float64(radius) {
				falloff := 1.0 - (dist / float64(radius))
				existing := img.At(x, y)
				r, g, b, a := existing.RGBA()

				darkening := 0.7 + falloff*0.3
				img.Set(x, y, color.RGBA{
					R: uint8(float64(r>>8) * darkening),
					G: uint8(float64(g>>8) * darkening),
					B: uint8(float64(b>>8) * darkening),
					A: uint8(a >> 8),
				})
			}
		}
	}
}

// applyDetails adds fine detail elements based on material and genre.
func (g *Generator) applyDetails(img *image.RGBA, material Material, r *rng.RNG) {
	if g.preset.DetailDensity < 0.1 {
		return
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Add detail points (small scratches, holes, etc)
	numDetails := int(float64(w*h) * 0.001 * g.preset.DetailDensity)
	for i := 0; i < numDetails; i++ {
		detailSeed := uint64(i * 31337)
		x := int(hashSeed(detailSeed) % uint64(w))
		y := int(hashSeed(detailSeed+1) % uint64(h))
		detailType := hashSeed(detailSeed+2) % 3

		switch detailType {
		case 0: // Dark spot
			if x < w && y < h {
				existing := img.At(x, y)
				r, g, b, a := existing.RGBA()
				img.Set(x, y, color.RGBA{
					R: uint8((r >> 8) * 7 / 10),
					G: uint8((g >> 8) * 7 / 10),
					B: uint8((b >> 8) * 7 / 10),
					A: uint8(a >> 8),
				})
			}
		case 1: // Bright spot
			if x < w && y < h {
				existing := img.At(x, y)
				r, g, b, a := existing.RGBA()
				img.Set(x, y, color.RGBA{
					R: clamp(int(r>>8) + 15),
					G: clamp(int(g>>8) + 15),
					B: clamp(int(b>>8) + 15),
					A: uint8(a >> 8),
				})
			}
		case 2: // Small line
			length := 2 + int(hashSeed(detailSeed+3)%4)
			horizontal := (hashSeed(detailSeed+4) % 2) == 0
			for j := 0; j < length; j++ {
				px, py := x, y
				if horizontal {
					px += j
				} else {
					py += j
				}
				if px < w && py < h {
					existing := img.At(px, py)
					r, g, b, a := existing.RGBA()
					img.Set(px, py, color.RGBA{
						R: uint8((r >> 8) * 8 / 10),
						G: uint8((g >> 8) * 8 / 10),
						B: uint8((b >> 8) * 8 / 10),
						A: uint8(a >> 8),
					})
				}
			}
		}
	}
}

// applyGlow adds glowing elements for sci-fi/cyberpunk genres.
func (g *Generator) applyGlow(img *image.RGBA, r *rng.RNG) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Add glowing strips or panels
	numGlows := int(float64(w/64) * g.preset.GlowIntensity * 3)
	for i := 0; i < numGlows; i++ {
		glowSeed := uint64(i * 99991)
		glowX := int(hashSeed(glowSeed) % uint64(w))
		glowY := int(hashSeed(glowSeed+1) % uint64(h))
		glowW := 1 + int(hashSeed(glowSeed+2)%3)
		glowH := 4 + int(hashSeed(glowSeed+3)%8)

		// Glow color (cyan/blue for sci-fi, magenta/cyan for cyberpunk)
		glowColor := color.RGBA{
			R: uint8(50 + hashSeed(glowSeed+4)%100),
			G: uint8(150 + hashSeed(glowSeed+5)%50),
			B: uint8(200 + hashSeed(glowSeed+6)%55),
			A: 255,
		}

		for y := glowY; y < glowY+glowH && y < h; y++ {
			for x := glowX; x < glowX+glowW && x < w; x++ {
				existing := img.At(x, y)
				er, eg, eb, ea := existing.RGBA()

				// Blend the glow
				blend := 0.5
				img.Set(x, y, color.RGBA{
					R: uint8(float64(er>>8)*(1-blend) + float64(glowColor.R)*blend),
					G: uint8(float64(eg>>8)*(1-blend) + float64(glowColor.G)*blend),
					B: uint8(float64(eb>>8)*(1-blend) + float64(glowColor.B)*blend),
					A: uint8(ea >> 8),
				})
			}
		}
	}
}

// octaveNoise generates multi-octave Perlin-like noise.
func (g *Generator) octaveNoise(x, y float64, octaves int, r *rng.RNG) float64 {
	var total float64
	var amplitude float64 = 1.0
	var frequency float64 = 1.0
	var maxValue float64

	for i := 0; i < octaves; i++ {
		total += g.perlinNoise(x*frequency, y*frequency, r) * amplitude
		maxValue += amplitude
		amplitude *= 0.5
		frequency *= 2.0
	}

	return total / maxValue
}

// perlinNoise generates simplified Perlin noise.
func (g *Generator) perlinNoise(x, y float64, r *rng.RNG) float64 {
	x0 := math.Floor(x)
	y0 := math.Floor(y)
	dx := x - x0
	dy := y - y0

	u := fade(dx)
	v := fade(dy)

	ix0, iy0 := int(x0), int(y0)

	gx00, gy00 := g.gradient(ix0, iy0)
	gx10, gy10 := g.gradient(ix0+1, iy0)
	gx01, gy01 := g.gradient(ix0, iy0+1)
	gx11, gy11 := g.gradient(ix0+1, iy0+1)

	n00 := gx00*dx + gy00*dy
	n10 := gx10*(dx-1) + gy10*dy
	n01 := gx01*dx + gy01*(dy-1)
	n11 := gx11*(dx-1) + gy11*(dy-1)

	nx0 := lerp(n00, n10, u)
	nx1 := lerp(n01, n11, u)

	return lerp(nx0, nx1, v)
}

// gradient returns a deterministic gradient vector for grid point.
func (g *Generator) gradient(x, y int) (float64, float64) {
	h := hashCoord(x, y)
	angle := float64(h%360) * math.Pi / 180.0
	return math.Cos(angle), math.Sin(angle)
}

// Helper functions

func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

func hashCoord(x, y int) uint64 {
	const prime1 uint64 = 73856093
	const prime2 uint64 = 19349663
	return uint64(x)*prime1 ^ uint64(y)*prime2
}

func hashSeed(seed uint64) uint64 {
	seed ^= seed >> 33
	seed *= 0xff51afd7ed558ccd
	seed ^= seed >> 33
	seed *= 0xc4ceb9fe1a85ec53
	seed ^= seed >> 33
	return seed
}

func clamp(val int) uint8 {
	if val < 0 {
		return 0
	}
	if val > 255 {
		return 255
	}
	return uint8(val)
}

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	// Standard luminance formula
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
}

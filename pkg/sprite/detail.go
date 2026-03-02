// Package sprite provides procedural sprite generation with rich material detail.
package sprite

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// MaterialDetail represents a procedural material texture layer.
type MaterialDetail int

const (
	MaterialScales MaterialDetail = iota
	MaterialFur
	MaterialChitin
	MaterialMembrane
	MaterialMetal
	MaterialCloth
	MaterialLeather
	MaterialCrystal
	MaterialSlime
)

// DetailLayer applies procedural material detail to a sprite region.
type DetailLayer struct {
	Material   MaterialDetail
	Intensity  float64
	Seed       int64
	ColorShift color.RGBA
}

// applyMaterialDetail adds rich textural detail to a sprite region.
func (g *Generator) applyMaterialDetail(img *image.RGBA, bounds image.Rectangle, material MaterialDetail, seed int64, intensity float64, baseColor color.RGBA) {
	rng := rand.New(rand.NewSource(seed))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			existingColor := img.At(x, y)
			er, eg, eb, ea := existingColor.RGBA()
			if ea == 0 {
				continue
			}

			var detail float64
			switch material {
			case MaterialScales:
				detail = g.scalePattern(x, y, rng, intensity)
			case MaterialFur:
				detail = g.furPattern(x, y, rng, intensity)
			case MaterialChitin:
				detail = g.chitinPattern(x, y, rng, intensity)
			case MaterialMembrane:
				detail = g.membranePattern(x, y, rng, intensity)
			case MaterialMetal:
				detail = g.metalPattern(x, y, rng, intensity)
			case MaterialCloth:
				detail = g.clothPattern(x, y, rng, intensity)
			case MaterialLeather:
				detail = g.leatherPattern(x, y, rng, intensity)
			case MaterialCrystal:
				detail = g.crystalPattern(x, y, rng, intensity)
			case MaterialSlime:
				detail = g.slimePattern(x, y, rng, intensity)
			}

			r := uint8(er >> 8)
			g := uint8(eg >> 8)
			b := uint8(eb >> 8)

			r = uint8(math.Max(0, math.Min(255, float64(r)+detail*intensity*30)))
			g = uint8(math.Max(0, math.Min(255, float64(g)+detail*intensity*30)))
			b = uint8(math.Max(0, math.Min(255, float64(b)+detail*intensity*30)))

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: uint8(ea >> 8)})
		}
	}
}

// scalePattern generates reptilian/fish scale texture.
func (g *Generator) scalePattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	scaleSize := 4.0
	sx := float64(x) / scaleSize
	sy := float64(y) / scaleSize

	// Hexagonal scale grid
	row := int(sy)
	col := int(sx)
	offsetX := 0.0
	if row%2 == 1 {
		offsetX = 0.5
	}

	// Distance to scale center
	centerX := (float64(col) + offsetX) * scaleSize
	centerY := float64(row) * scaleSize
	dx := float64(x) - centerX
	dy := float64(y) - centerY
	dist := math.Sqrt(dx*dx + dy*dy)

	// Scale edge highlight
	scaleRadius := scaleSize / 2
	if dist < scaleRadius {
		edge := 1.0 - dist/scaleRadius
		highlight := math.Pow(edge, 3.0)
		return (highlight - 0.5) * 2.0
	}

	// Scale overlap shadow
	return -0.3
}

// furPattern generates fuzzy fur texture.
func (g *Generator) furPattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Directional fur strands
	strandAngle := math.Pi / 4
	perpDist := math.Abs(float64(x)*math.Cos(strandAngle) - float64(y)*math.Sin(strandAngle))
	strandPhase := math.Mod(perpDist, 2.0)

	noise := (rng.Float64() - 0.5) * 0.4
	return (strandPhase/2.0 - 0.5) + noise
}

// chitinPattern generates insect exoskeleton texture.
func (g *Generator) chitinPattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Segmented plates with edge highlights
	segmentSize := 6.0
	segmentY := math.Mod(float64(y), segmentSize)
	plateEdge := 0.0

	if segmentY < 1.0 {
		plateEdge = 1.0 - segmentY
	} else if segmentY > segmentSize-1.0 {
		plateEdge = segmentY - (segmentSize - 1.0)
	}

	// Subtle plate texture
	noiseX := math.Sin(float64(x) * 0.3)
	noiseY := math.Cos(float64(y) * 0.2)
	plateNoise := (noiseX + noiseY) * 0.15

	return plateEdge*0.8 + plateNoise
}

// membranePattern generates translucent wing/skin texture.
func (g *Generator) membranePattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Vein network
	veinDensity := 8.0
	vx := math.Sin(float64(x)/veinDensity) * veinDensity
	vy := math.Cos(float64(y)/veinDensity) * veinDensity
	veinDist := math.Sqrt(vx*vx + vy*vy)
	veinHighlight := 0.0
	if math.Mod(veinDist, veinDensity) < 1.0 {
		veinHighlight = 0.4
	}

	// Membrane wrinkles
	wrinkle := math.Sin(float64(x)*0.4+float64(y)*0.3) * 0.2

	return veinHighlight + wrinkle
}

// metalPattern generates polished metal texture.
func (g *Generator) metalPattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Anisotropic brushed metal effect
	brushDirection := float64(y)
	brushNoise := math.Sin(brushDirection*0.5+rng.Float64()*2*math.Pi) * 0.3

	// Specular highlights based on position variation
	posNoise := math.Abs(math.Sin(float64(x)*0.3)) * 0.5

	return brushNoise + posNoise
}

// clothPattern generates woven fabric texture.
func (g *Generator) clothPattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Woven thread grid
	threadSize := 2.0
	warpPhase := math.Mod(float64(x), threadSize) / threadSize
	weftPhase := math.Mod(float64(y), threadSize) / threadSize

	weave := 0.0
	if (warpPhase < 0.5 && weftPhase < 0.5) || (warpPhase >= 0.5 && weftPhase >= 0.5) {
		weave = 0.2
	} else {
		weave = -0.2
	}

	return weave
}

// leatherPattern generates hide/leather texture.
func (g *Generator) leatherPattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Organic grain variation
	grainX := math.Sin(float64(x)*0.2 + rng.Float64()*6.28)
	grainY := math.Cos(float64(y)*0.3 + rng.Float64()*6.28)
	grain := (grainX + grainY) * 0.15

	// Pores
	poreNoise := 0.0
	if rng.Float64() < 0.05 {
		poreNoise = -0.3
	}

	return grain + poreNoise
}

// crystalPattern generates crystalline facet texture.
func (g *Generator) crystalPattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Angular facets
	facetSize := 8.0
	fx := math.Floor(float64(x) / facetSize)
	fy := math.Floor(float64(y) / facetSize)

	// Facet orientation determines brightness
	facetSeed := int64(fx*1000 + fy)
	facetRng := rand.New(rand.NewSource(facetSeed))
	facetBrightness := facetRng.Float64()*2.0 - 1.0

	// Sharp facet edges
	edgeDist := math.Min(
		math.Mod(float64(x), facetSize),
		math.Mod(float64(y), facetSize),
	)
	edgeHighlight := 0.0
	if edgeDist < 1.0 {
		edgeHighlight = 0.6
	}

	return facetBrightness*0.5 + edgeHighlight
}

// slimePattern generates viscous fluid texture.
func (g *Generator) slimePattern(x, y int, rng *rand.Rand, intensity float64) float64 {
	// Flowing ripples
	ripplePhase := math.Sin(float64(x)*0.3) * math.Cos(float64(y)*0.25)
	ripple := ripplePhase * 0.3

	// Bubbles
	bubbleSeed := int64(x/6*1000 + y/6)
	bubbleRng := rand.New(rand.NewSource(bubbleSeed))
	if bubbleRng.Float64() < 0.08 {
		// Inside bubble
		bx := float64(x%6 - 3)
		by := float64(y%6 - 3)
		bubbleDist := math.Sqrt(bx*bx + by*by)
		if bubbleDist < 2.5 {
			return 0.8 - bubbleDist*0.2
		}
	}

	return ripple
}

// applyEquipmentOverlay renders visible equipment on a sprite.
func (g *Generator) applyEquipmentOverlay(img *image.RGBA, bounds image.Rectangle, equipmentType string, seed int64, baseColor color.RGBA) {
	rng := rand.New(rand.NewSource(seed))
	cx := (bounds.Min.X + bounds.Max.X) / 2
	cy := (bounds.Min.Y + bounds.Max.Y) / 2
	size := bounds.Max.X - bounds.Min.X

	switch equipmentType {
	case "helmet":
		g.drawHelmetOverlay(img, cx, cy-size/4, size/4, rng, baseColor)
	case "armor":
		g.drawArmorOverlay(img, cx, cy, size/3, rng, baseColor)
	case "shield":
		g.drawShieldOverlay(img, cx-size/3, cy, size/4, rng, baseColor)
	case "weapon":
		g.drawWeaponOverlay(img, cx+size/3, cy, size/3, rng, baseColor)
	case "aura":
		g.drawAuraOverlay(img, cx, cy, size/2, rng, baseColor)
	}
}

// drawHelmetOverlay renders helmet detail.
func (g *Generator) drawHelmetOverlay(img *image.RGBA, cx, cy, radius int, rng *rand.Rand, baseColor color.RGBA) {
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			if x*x+y*y <= radius*radius {
				px := cx + x
				py := cy + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					// Blend with existing pixel
					existing := img.At(px, py)
					er, eg, eb, ea := existing.RGBA()
					if ea > 0 {
						blend := 0.3
						r := uint8((float64(er>>8)*(1-blend) + float64(baseColor.R)*blend))
						g := uint8((float64(eg>>8)*(1-blend) + float64(baseColor.G)*blend))
						b := uint8((float64(eb>>8)*(1-blend) + float64(baseColor.B)*blend))
						img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: uint8(ea >> 8)})
					}
				}
			}
		}
	}
}

// drawArmorOverlay renders chest armor detail.
func (g *Generator) drawArmorOverlay(img *image.RGBA, cx, cy, size int, rng *rand.Rand, armorColor color.RGBA) {
	// Chest plate highlight
	for y := -size / 2; y <= size/2; y++ {
		for x := -size / 3; x <= size/3; x++ {
			px := cx + x
			py := cy + y
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				existing := img.At(px, py)
				_, _, _, ea := existing.RGBA()
				if ea > 0 && abs(x) < size/4 {
					// Center highlight
					highlight := uint8(float64(armorColor.R) * 1.2)
					img.Set(px, py, color.RGBA{R: highlight, G: highlight, B: highlight, A: 100})
				}
			}
		}
	}
}

// drawShieldOverlay renders shield equipment.
func (g *Generator) drawShieldOverlay(img *image.RGBA, cx, cy, size int, rng *rand.Rand, shieldColor color.RGBA) {
	// Shield outline
	for y := -size; y <= size; y++ {
		for x := -size / 2; x <= 0; x++ {
			if x*x+y*y/2 <= size*size/2 {
				px := cx + x
				py := cy + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					dist := math.Sqrt(float64(x*x + y*y/2))
					alpha := uint8(150 - dist*2)
					img.Set(px, py, color.RGBA{R: shieldColor.R, G: shieldColor.G, B: shieldColor.B, A: alpha})
				}
			}
		}
	}
}

// drawWeaponOverlay renders weapon equipment.
func (g *Generator) drawWeaponOverlay(img *image.RGBA, cx, cy, length int, rng *rand.Rand, weaponColor color.RGBA) {
	// Weapon blade/barrel
	thickness := 2
	for i := 0; i < length; i++ {
		for t := -thickness; t <= thickness; t++ {
			px := cx + i
			py := cy + t
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				// Gradient along length
				shade := 1.0 - float64(i)/float64(length)*0.3
				r := uint8(float64(weaponColor.R) * shade)
				g := uint8(float64(weaponColor.G) * shade)
				b := uint8(float64(weaponColor.B) * shade)
				img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: 200})
			}
		}
	}
}

// drawAuraOverlay renders magical aura effect.
func (g *Generator) drawAuraOverlay(img *image.RGBA, cx, cy, radius int, rng *rand.Rand, auraColor color.RGBA) {
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			dist := math.Sqrt(float64(x*x + y*y))
			if dist > float64(radius)*0.7 && dist < float64(radius) {
				px := cx + x
				py := cy + y
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					// Pulsing aura ring
					pulseIntensity := 1.0 - math.Abs(dist-float64(radius)*0.85)/(float64(radius)*0.15)
					alpha := uint8(pulseIntensity * 80)
					img.Set(px, py, color.RGBA{R: auraColor.R, G: auraColor.G, B: auraColor.B, A: alpha})
				}
			}
		}
	}
}

// applyStatusEffectOverlay renders visible status effects.
func (g *Generator) applyStatusEffectOverlay(img *image.RGBA, bounds image.Rectangle, statusType string, frame int) {
	cx := (bounds.Min.X + bounds.Max.X) / 2
	cy := (bounds.Min.Y + bounds.Max.Y) / 2
	size := bounds.Max.X - bounds.Min.X

	switch statusType {
	case "burning":
		g.drawBurningEffect(img, cx, cy, size, frame)
	case "frozen":
		g.drawFrozenEffect(img, bounds, frame)
	case "poisoned":
		g.drawPoisonedEffect(img, cx, cy, size, frame)
	case "shielded":
		g.drawShieldedEffect(img, cx, cy, size, frame)
	case "enraged":
		g.drawEnragedEffect(img, bounds, frame)
	}
}

// drawBurningEffect renders fire particles.
func (g *Generator) drawBurningEffect(img *image.RGBA, cx, cy, size, frame int) {
	rng := rand.New(rand.NewSource(int64(frame)))
	for i := 0; i < 5; i++ {
		px := cx + rng.Intn(size/2) - size/4
		py := cy + size/3 - (frame+i*3)%20
		if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
			intensity := 1.0 - float64((frame+i*3)%20)/20.0
			r := uint8(255 * intensity)
			g := uint8(150 * intensity)
			img.Set(px, py, color.RGBA{R: r, G: g, B: 0, A: uint8(intensity * 200)})
		}
	}
}

// drawFrozenEffect renders frost overlay.
func (g *Generator) drawFrozenEffect(img *image.RGBA, bounds image.Rectangle, frame int) {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				existing := img.At(x, y)
				_, _, _, ea := existing.RGBA()
				if ea > 0 {
					// Ice crystal pattern
					crystalNoise := math.Sin(float64(x)*0.8) * math.Cos(float64(y)*0.6)
					if crystalNoise > 0.5 {
						img.Set(x, y, color.RGBA{R: 180, G: 220, B: 255, A: 80})
					}
				}
			}
		}
	}
}

// drawPoisonedEffect renders poison drips.
func (g *Generator) drawPoisonedEffect(img *image.RGBA, cx, cy, size, frame int) {
	rng := rand.New(rand.NewSource(int64(frame)))
	for i := 0; i < 3; i++ {
		px := cx + rng.Intn(size/3) - size/6
		py := cy + (frame+i*5)%25
		if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
			img.Set(px, py, color.RGBA{R: 100, G: 255, B: 100, A: 180})
			img.Set(px, py+1, color.RGBA{R: 80, G: 200, B: 80, A: 120})
		}
	}
}

// drawShieldedEffect renders shield barrier.
func (g *Generator) drawShieldedEffect(img *image.RGBA, cx, cy, size, frame int) {
	radius := size/2 + 2
	hexPoints := 6
	for i := 0; i < hexPoints; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(hexPoints)
		x1 := cx + int(float64(radius)*math.Cos(angle))
		y1 := cy + int(float64(radius)*math.Sin(angle))
		x2 := cx + int(float64(radius)*math.Cos(angle+2.0*math.Pi/float64(hexPoints)))
		y2 := cy + int(float64(radius)*math.Sin(angle+2.0*math.Pi/float64(hexPoints)))

		// Draw hex edge
		steps := 20
		for s := 0; s <= steps; s++ {
			t := float64(s) / float64(steps)
			px := int(float64(x1)*(1-t) + float64(x2)*t)
			py := int(float64(y1)*(1-t) + float64(y2)*t)
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				pulse := uint8((math.Sin(float64(frame)*0.2)+1.0)*60 + 60)
				img.Set(px, py, color.RGBA{R: 100, G: 150, B: 255, A: pulse})
			}
		}
	}
}

// drawEnragedEffect renders rage aura.
func (g *Generator) drawEnragedEffect(img *image.RGBA, bounds image.Rectangle, frame int) {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				existing := img.At(x, y)
				er, eg, eb, ea := existing.RGBA()
				if ea > 0 {
					// Red tint pulse
					pulse := math.Sin(float64(frame)*0.3) * 0.3
					r := uint8(math.Min(255, float64(er>>8)*(1+pulse)))
					g := uint8(float64(eg >> 8))
					b := uint8(float64(eb >> 8))
					img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: uint8(ea >> 8)})
				}
			}
		}
	}
}

// abs returns absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

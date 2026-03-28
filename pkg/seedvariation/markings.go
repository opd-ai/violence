// Package seedvariation provides seed-driven structural variation for procedural entities.
package seedvariation

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// ApplyMarkings renders variation-determined markings (spots, stripes, scars) onto a sprite.
func ApplyMarkings(img *image.RGBA, bounds image.Rectangle, v *Variation, baseColor color.RGBA) {
	if v.HasSpotting {
		applySpots(img, bounds, v, baseColor)
	}
	if v.HasStripes {
		applyStripes(img, bounds, v, baseColor)
	}
	if v.HasScar {
		applyScar(img, bounds, v)
	}
	if v.HasHorns && v.HornCount > 0 {
		applyHornMarkings(img, bounds, v, baseColor)
	}
}

// applySpots renders colored spots/dots on the sprite.
func applySpots(img *image.RGBA, bounds image.Rectangle, v *Variation, baseColor color.RGBA) {
	rng := rand.New(rand.NewSource(v.PatternSeed))

	// Spot color is lighter or darker than base
	spotColor := lightenColor(baseColor, 0.3)
	if rng.Float64() < 0.5 {
		spotColor = darkenColor(baseColor, 0.3)
	}

	cx := (bounds.Min.X + bounds.Max.X) / 2
	cy := (bounds.Min.Y + bounds.Max.Y) / 2
	radiusX := (bounds.Max.X - bounds.Min.X) / 2
	radiusY := (bounds.Max.Y - bounds.Min.Y) / 2

	for i := 0; i < v.MarkingCount; i++ {
		// Random position within bounds (elliptical distribution)
		angle := rng.Float64() * 2 * math.Pi
		dist := rng.Float64() * 0.7 // Stay within 70% of radius
		spotX := cx + int(math.Cos(angle)*float64(radiusX)*dist)
		spotY := cy + int(math.Sin(angle)*float64(radiusY)*dist)

		// Spot size varies
		spotRadius := 1 + rng.Intn(3)

		// Draw spot
		for dy := -spotRadius; dy <= spotRadius; dy++ {
			for dx := -spotRadius; dx <= spotRadius; dx++ {
				if dx*dx+dy*dy <= spotRadius*spotRadius {
					px := spotX + dx
					py := spotY + dy
					if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
						existing := img.At(px, py)
						_, _, _, ea := existing.RGBA()
						if ea > 0 { // Only draw on existing pixels
							img.Set(px, py, spotColor)
						}
					}
				}
			}
		}
	}
}

// applyStripes renders stripe patterns along the body.
func applyStripes(img *image.RGBA, bounds image.Rectangle, v *Variation, baseColor color.RGBA) {
	rng := rand.New(rand.NewSource(v.PatternSeed + 1000))

	// Stripe color is contrasting
	stripeColor := darkenColor(baseColor, 0.4)
	if rng.Float64() < 0.3 {
		stripeColor = lightenColor(baseColor, 0.3)
	}

	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// Determine stripe direction based on body orientation
	vertical := height > width

	if v.MarkingCount <= 0 {
		return
	}
	stripeSpacing := height / v.MarkingCount
	if !vertical {
		stripeSpacing = width / v.MarkingCount
	}
	if stripeSpacing < 2 {
		stripeSpacing = 2
	}

	stripeWidth := 1 + rng.Intn(2)

	for i := 0; i < v.MarkingCount; i++ {
		if vertical {
			// Horizontal stripes
			stripeY := bounds.Min.Y + i*stripeSpacing + stripeSpacing/2

			for dy := 0; dy < stripeWidth; dy++ {
				py := stripeY + dy
				if py >= bounds.Max.Y {
					continue
				}
				for px := bounds.Min.X; px < bounds.Max.X; px++ {
					if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
						existing := img.At(px, py)
						_, _, _, ea := existing.RGBA()
						if ea > 0 {
							img.Set(px, py, stripeColor)
						}
					}
				}
			}
		} else {
			// Vertical stripes
			stripeX := bounds.Min.X + i*stripeSpacing + stripeSpacing/2

			for dx := 0; dx < stripeWidth; dx++ {
				px := stripeX + dx
				if px >= bounds.Max.X {
					continue
				}
				for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
					if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
						existing := img.At(px, py)
						_, _, _, ea := existing.RGBA()
						if ea > 0 {
							img.Set(px, py, stripeColor)
						}
					}
				}
			}
		}
	}
}

// applyScar renders a visible scar mark.
func applyScar(img *image.RGBA, bounds image.Rectangle, v *Variation) {
	rng := rand.New(rand.NewSource(v.PatternSeed + 2000))

	scarColor := color.RGBA{R: 180, G: 100, B: 100, A: 255}

	cx := (bounds.Min.X + bounds.Max.X) / 2
	cy := (bounds.Min.Y + bounds.Max.Y) / 2

	// Scar position varies
	scarX := cx + rng.Intn(5) - 2
	scarY := cy + rng.Intn(5) - 2

	// Draw diagonal scar line
	scarLen := 3 + rng.Intn(4)
	scarAngle := rng.Float64() * math.Pi

	for i := 0; i < scarLen; i++ {
		px := scarX + int(math.Cos(scarAngle)*float64(i))
		py := scarY + int(math.Sin(scarAngle)*float64(i))

		if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
			existing := img.At(px, py)
			_, _, _, ea := existing.RGBA()
			if ea > 0 {
				img.Set(px, py, scarColor)
				// Thicken scar slightly
				if py+1 < img.Bounds().Dy() {
					img.Set(px, py+1, scarColor)
				}
			}
		}
	}
}

// applyHornMarkings renders horn/spike base indicators.
func applyHornMarkings(img *image.RGBA, bounds image.Rectangle, v *Variation, baseColor color.RGBA) {
	rng := rand.New(rand.NewSource(v.PatternSeed + 3000))

	hornColor := darkenColor(baseColor, 0.5)

	cx := (bounds.Min.X + bounds.Max.X) / 2
	topY := bounds.Min.Y

	for i := 0; i < v.HornCount; i++ {
		// Position horns along top edge
		hornX := cx + (i-v.HornCount/2)*(bounds.Max.X-bounds.Min.X)/(v.HornCount+1)
		hornX += rng.Intn(3) - 1 // Slight randomization

		// Draw small triangle/spike
		hornHeight := 3 + rng.Intn(3)
		for dy := 0; dy < hornHeight; dy++ {
			py := topY - dy
			width := hornHeight - dy
			if width < 1 {
				width = 1
			}
			for dx := -width / 2; dx <= width/2; dx++ {
				px := hornX + dx
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, hornColor)
				}
			}
		}
	}
}

// ApplyExtraEyes renders additional eyes for creatures with HasExtraEyes.
func ApplyExtraEyes(img *image.RGBA, headCenterX, headCenterY, headRadius int, v *Variation) {
	if !v.HasExtraEyes || v.EyeCount <= 2 {
		return
	}

	rng := rand.New(rand.NewSource(v.PatternSeed + 4000))

	eyeColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}
	pupilColor := color.RGBA{R: 20, G: 20, B: 20, A: 255}

	// Add extra eyes beyond the base 2
	extraEyes := v.EyeCount - 2
	for i := 0; i < extraEyes; i++ {
		// Position extra eyes randomly around head
		angle := rng.Float64() * 2 * math.Pi
		dist := float64(headRadius) * (0.3 + rng.Float64()*0.4)
		eyeX := headCenterX + int(math.Cos(angle)*dist)
		eyeY := headCenterY + int(math.Sin(angle)*dist)

		// Smaller extra eyes
		eyeRadius := 1

		// Draw eye
		for dy := -eyeRadius; dy <= eyeRadius; dy++ {
			for dx := -eyeRadius; dx <= eyeRadius; dx++ {
				if dx*dx+dy*dy <= eyeRadius*eyeRadius {
					px := eyeX + dx
					py := eyeY + dy
					if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
						img.Set(px, py, eyeColor)
					}
				}
			}
		}

		// Pupil
		if eyeX >= 0 && eyeX < img.Bounds().Dx() && eyeY >= 0 && eyeY < img.Bounds().Dy() {
			img.Set(eyeX, eyeY, pupilColor)
		}
	}
}

// ApplyWear renders wear/damage effects based on WearLevel.
func ApplyWear(img *image.RGBA, bounds image.Rectangle, v *Variation) {
	if v.WearLevel <= 0.1 {
		return
	}

	rng := rand.New(rand.NewSource(v.PatternSeed + 5000))

	// Number of wear marks scales with wear level
	wearMarks := int(v.WearLevel * 10)

	for i := 0; i < wearMarks; i++ {
		// Random position
		px := bounds.Min.X + rng.Intn(bounds.Max.X-bounds.Min.X)
		py := bounds.Min.Y + rng.Intn(bounds.Max.Y-bounds.Min.Y)

		if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
			existing := img.At(px, py)
			er, eg, eb, ea := existing.RGBA()
			if ea == 0 {
				continue
			}

			// Darken/roughen the pixel
			r := uint8((er >> 8) * 8 / 10)
			g := uint8((eg >> 8) * 8 / 10)
			b := uint8((eb >> 8) * 8 / 10)

			img.Set(px, py, color.RGBA{R: r, G: g, B: b, A: uint8(ea >> 8)})

			// Spread wear to neighbors
			for _, delta := range [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}} {
				nx := px + delta[0]
				ny := py + delta[1]
				if nx >= 0 && nx < img.Bounds().Dx() && ny >= 0 && ny < img.Bounds().Dy() {
					if rng.Float64() < v.WearLevel {
						ne := img.At(nx, ny)
						ner, neg, neb, nea := ne.RGBA()
						if nea > 0 {
							img.Set(nx, ny, color.RGBA{
								R: uint8((ner >> 8) * 9 / 10),
								G: uint8((neg >> 8) * 9 / 10),
								B: uint8((neb >> 8) * 9 / 10),
								A: uint8(nea >> 8),
							})
						}
					}
				}
			}
		}
	}
}

// lightenColor creates a lighter version of the color.
func lightenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(c.R)*(1+factor))),
		G: uint8(math.Min(255, float64(c.G)*(1+factor))),
		B: uint8(math.Min(255, float64(c.B)*(1+factor))),
		A: c.A,
	}
}

// darkenColor creates a darker version of the color.
func darkenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * (1 - factor)),
		G: uint8(float64(c.G) * (1 - factor)),
		B: uint8(float64(c.B) * (1 - factor)),
		A: c.A,
	}
}

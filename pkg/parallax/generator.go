package parallax

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// GenerateLayers creates parallax layers for a given genre and biome.
func GenerateLayers(genreID, biomeID string, seed int64, width, height int) []*Layer {
	rng := rand.New(rand.NewSource(seed))

	switch genreID {
	case "fantasy":
		return generateFantasyLayers(rng, width, height)
	case "scifi":
		return generateSciFiLayers(rng, width, height)
	case "horror":
		return generateHorrorLayers(rng, width, height)
	case "cyberpunk":
		return generateCyberpunkLayers(rng, width, height)
	case "postapoc":
		return generatePostApocLayers(rng, width, height)
	default:
		return generateFantasyLayers(rng, width, height)
	}
}

// generateFantasyLayers creates fantasy-themed parallax layers.
func generateFantasyLayers(rng *rand.Rand, width, height int) []*Layer {
	layers := make([]*Layer, 0, 4)

	// Layer 0: Distant mountains
	mountains := generateMountains(rng, width, height/2,
		color.RGBA{R: 60, G: 70, B: 90, A: 180})
	layers = append(layers, &Layer{
		Image:       mountains,
		ScrollSpeed: 0.1,
		RepeatX:     true,
		Opacity:     0.6,
		ZIndex:      0,
		Width:       width,
		Height:      height / 2,
		Tint:        [4]float64{0.7, 0.75, 0.9, 1.0},
	})

	// Layer 1: Hills
	hills := generateHills(rng, width, height/3,
		color.RGBA{R: 80, G: 100, B: 70, A: 200})
	layers = append(layers, &Layer{
		Image:       hills,
		ScrollSpeed: 0.25,
		RepeatX:     true,
		Opacity:     0.75,
		ZIndex:      1,
		Width:       width,
		Height:      height / 3,
		Tint:        [4]float64{0.8, 0.85, 0.75, 1.0},
	})

	// Layer 2: Trees
	trees := generateTrees(rng, width, height/2,
		color.RGBA{R: 40, G: 80, B: 40, A: 220})
	layers = append(layers, &Layer{
		Image:       trees,
		ScrollSpeed: 0.5,
		RepeatX:     true,
		Opacity:     0.85,
		ZIndex:      2,
		Width:       width,
		Height:      height / 2,
		Tint:        [4]float64{0.85, 0.9, 0.8, 1.0},
	})

	return layers
}

// generateSciFiLayers creates sci-fi themed parallax layers.
func generateSciFiLayers(rng *rand.Rand, width, height int) []*Layer {
	layers := make([]*Layer, 0, 4)

	// Layer 0: Starfield
	starfield := generateStarfield(rng, width, height,
		color.RGBA{R: 200, G: 200, B: 255, A: 200})
	layers = append(layers, &Layer{
		Image:       starfield,
		ScrollSpeed: 0.05,
		RepeatX:     true,
		RepeatY:     true,
		Opacity:     0.7,
		ZIndex:      0,
		Width:       width,
		Height:      height,
		Tint:        [4]float64{0.8, 0.8, 1.0, 1.0},
	})

	// Layer 1: Distant structures
	structures := generateStructures(rng, width, height/2,
		color.RGBA{R: 100, G: 120, B: 140, A: 180})
	layers = append(layers, &Layer{
		Image:       structures,
		ScrollSpeed: 0.2,
		RepeatX:     true,
		Opacity:     0.65,
		ZIndex:      1,
		Width:       width,
		Height:      height / 2,
		Tint:        [4]float64{0.7, 0.8, 0.9, 1.0},
	})

	// Layer 2: Tech panels
	panels := generateTechPanels(rng, width, height/3,
		color.RGBA{R: 80, G: 100, B: 120, A: 220})
	layers = append(layers, &Layer{
		Image:       panels,
		ScrollSpeed: 0.4,
		RepeatX:     true,
		Opacity:     0.8,
		ZIndex:      2,
		Width:       width,
		Height:      height / 3,
		Tint:        [4]float64{0.75, 0.85, 0.95, 1.0},
	})

	return layers
}

// generateHorrorLayers creates horror-themed parallax layers.
func generateHorrorLayers(rng *rand.Rand, width, height int) []*Layer {
	layers := make([]*Layer, 0, 4)

	// Layer 0: Fog
	fog := generateFog(rng, width, height,
		color.RGBA{R: 80, G: 80, B: 90, A: 150})
	layers = append(layers, &Layer{
		Image:       fog,
		ScrollSpeed: 0.15,
		RepeatX:     true,
		RepeatY:     true,
		Opacity:     0.5,
		ZIndex:      0,
		Width:       width,
		Height:      height,
		Tint:        [4]float64{0.6, 0.6, 0.7, 1.0},
	})

	// Layer 1: Dead trees
	deadTrees := generateDeadTrees(rng, width, height/2,
		color.RGBA{R: 40, G: 35, B: 40, A: 200})
	layers = append(layers, &Layer{
		Image:       deadTrees,
		ScrollSpeed: 0.3,
		RepeatX:     true,
		Opacity:     0.7,
		ZIndex:      1,
		Width:       width,
		Height:      height / 2,
		Tint:        [4]float64{0.5, 0.5, 0.55, 1.0},
	})

	return layers
}

// generateCyberpunkLayers creates cyberpunk-themed parallax layers.
func generateCyberpunkLayers(rng *rand.Rand, width, height int) []*Layer {
	layers := make([]*Layer, 0, 4)

	// Layer 0: Neon skyline
	skyline := generateNeonSkyline(rng, width, height/2,
		color.RGBA{R: 255, G: 50, B: 150, A: 200})
	layers = append(layers, &Layer{
		Image:       skyline,
		ScrollSpeed: 0.15,
		RepeatX:     true,
		Opacity:     0.6,
		ZIndex:      0,
		Width:       width,
		Height:      height / 2,
		Tint:        [4]float64{1.0, 0.6, 0.8, 1.0},
	})

	// Layer 1: Rain streaks
	rain := generateRain(rng, width, height,
		color.RGBA{R: 150, G: 180, B: 200, A: 120})
	layers = append(layers, &Layer{
		Image:       rain,
		ScrollSpeed: 0.6,
		RepeatX:     true,
		RepeatY:     true,
		Opacity:     0.4,
		ZIndex:      1,
		Width:       width,
		Height:      height,
		Tint:        [4]float64{0.8, 0.9, 1.0, 1.0},
	})

	// Layer 2: Building fronts
	buildings := generateBuildings(rng, width, height*2/3,
		color.RGBA{R: 60, G: 70, B: 80, A: 220})
	layers = append(layers, &Layer{
		Image:       buildings,
		ScrollSpeed: 0.45,
		RepeatX:     true,
		Opacity:     0.85,
		ZIndex:      2,
		Width:       width,
		Height:      height * 2 / 3,
		Tint:        [4]float64{0.7, 0.75, 0.85, 1.0},
	})

	return layers
}

// generatePostApocLayers creates post-apocalyptic parallax layers.
func generatePostApocLayers(rng *rand.Rand, width, height int) []*Layer {
	layers := make([]*Layer, 0, 4)

	// Layer 0: Ruined cityscape
	ruins := generateRuins(rng, width, height/2,
		color.RGBA{R: 70, G: 65, B: 60, A: 180})
	layers = append(layers, &Layer{
		Image:       ruins,
		ScrollSpeed: 0.2,
		RepeatX:     true,
		Opacity:     0.65,
		ZIndex:      0,
		Width:       width,
		Height:      height / 2,
		Tint:        [4]float64{0.75, 0.7, 0.65, 1.0},
	})

	// Layer 1: Dust clouds
	dust := generateDust(rng, width, height,
		color.RGBA{R: 100, G: 90, B: 70, A: 100})
	layers = append(layers, &Layer{
		Image:       dust,
		ScrollSpeed: 0.35,
		RepeatX:     true,
		RepeatY:     true,
		Opacity:     0.5,
		ZIndex:      1,
		Width:       width,
		Height:      height,
		Tint:        [4]float64{0.8, 0.75, 0.65, 1.0},
	})

	return layers
}

// generateMountains creates a mountain silhouette.
func generateMountains(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Generate mountain peaks using noise
	peaks := make([]int, width)
	for x := 0; x < width; x++ {
		octave1 := math.Sin(float64(x)*0.02) * 0.4
		octave2 := math.Sin(float64(x)*0.05+rng.Float64()*10) * 0.3
		octave3 := rng.Float64() * 0.1
		noise := octave1 + octave2 + octave3
		peaks[x] = int(float64(height) * (0.3 + noise*0.4))
	}

	// Draw mountains with gradient shading
	for x := 0; x < width; x++ {
		peakY := peaks[x]
		for y := peakY; y < height; y++ {
			shade := 1.0 - float64(y-peakY)/float64(height-peakY)*0.5
			r := uint8(float64(baseColor.R) * shade)
			g := uint8(float64(baseColor.G) * shade)
			b := uint8(float64(baseColor.B) * shade)
			a := baseColor.A
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateHills creates rolling hills.
func generateHills(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	hillLine := make([]int, width)
	for x := 0; x < width; x++ {
		noise := math.Sin(float64(x)*0.03+rng.Float64()*5) * 0.4
		hillLine[x] = int(float64(height) * (0.5 + noise*0.3))
	}

	for x := 0; x < width; x++ {
		for y := hillLine[x]; y < height; y++ {
			shade := 1.0 - float64(y-hillLine[x])/float64(height)*0.3
			r := uint8(float64(baseColor.R) * shade)
			g := uint8(float64(baseColor.G) * shade)
			b := uint8(float64(baseColor.B) * shade)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: baseColor.A})
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateTrees creates a tree line silhouette.
func generateTrees(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Place trees at intervals
	for x := 0; x < width; x += 20 + rng.Intn(30) {
		treeHeight := height*2/3 + rng.Intn(height/3)
		treeWidth := 10 + rng.Intn(15)

		// Draw tree trunk
		trunkX := x + treeWidth/2
		for y := height - treeHeight; y < height; y++ {
			for tx := trunkX - 2; tx <= trunkX+2; tx++ {
				if tx >= 0 && tx < width {
					img.Set(tx, y, baseColor)
				}
			}
		}

		// Draw tree crown (triangle)
		crownTop := height - treeHeight
		for dy := 0; dy < treeHeight/2; dy++ {
			crownWidth := int(float64(treeWidth) * float64(dy) / float64(treeHeight/2))
			for dx := -crownWidth; dx <= crownWidth; dx++ {
				px := trunkX + dx
				py := crownTop + dy
				if px >= 0 && px < width && py >= 0 && py < height {
					img.Set(px, py, baseColor)
				}
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateStarfield creates a starfield pattern.
func generateStarfield(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Place stars randomly
	numStars := width * height / 200
	for i := 0; i < numStars; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)
		brightness := 0.3 + rng.Float64()*0.7

		r := uint8(float64(baseColor.R) * brightness)
		g := uint8(float64(baseColor.G) * brightness)
		b := uint8(float64(baseColor.B) * brightness)

		img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: baseColor.A})

		// Larger stars get a glow
		if brightness > 0.8 {
			img.Set(x-1, y, color.RGBA{R: r / 2, G: g / 2, B: b / 2, A: baseColor.A / 2})
			img.Set(x+1, y, color.RGBA{R: r / 2, G: g / 2, B: b / 2, A: baseColor.A / 2})
			img.Set(x, y-1, color.RGBA{R: r / 2, G: g / 2, B: b / 2, A: baseColor.A / 2})
			img.Set(x, y+1, color.RGBA{R: r / 2, G: g / 2, B: b / 2, A: baseColor.A / 2})
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateStructures creates distant building silhouettes.
func generateStructures(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	x := 0
	for x < width {
		buildingWidth := 30 + rng.Intn(50)
		buildingHeight := height/3 + rng.Intn(height*2/3)

		for bx := x; bx < x+buildingWidth && bx < width; bx++ {
			for by := height - buildingHeight; by < height; by++ {
				shade := 0.6 + rng.Float64()*0.2
				r := uint8(float64(baseColor.R) * shade)
				g := uint8(float64(baseColor.G) * shade)
				b := uint8(float64(baseColor.B) * shade)
				img.Set(bx, by, color.RGBA{R: r, G: g, B: b, A: baseColor.A})
			}
		}

		x += buildingWidth + rng.Intn(20)
	}

	return ebiten.NewImageFromImage(img)
}

// generateTechPanels creates tech panel patterns.
func generateTechPanels(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Draw horizontal tech lines
	for y := 10; y < height; y += 15 + rng.Intn(20) {
		for x := 0; x < width; x++ {
			if rng.Float64() > 0.3 {
				brightness := 0.7 + rng.Float64()*0.3
				r := uint8(float64(baseColor.R) * brightness)
				g := uint8(float64(baseColor.G) * brightness)
				b := uint8(float64(baseColor.B) * brightness)
				img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: baseColor.A})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateFog creates fog wisps.
func generateFog(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create fog blobs
	numBlobs := 10 + rng.Intn(10)
	for i := 0; i < numBlobs; i++ {
		cx := rng.Intn(width)
		cy := rng.Intn(height)
		radius := 30 + rng.Intn(50)

		for y := cy - radius; y < cy+radius; y++ {
			for x := cx - radius; x < cx+radius; x++ {
				if x < 0 || x >= width || y < 0 || y >= height {
					continue
				}

				dx := float64(x - cx)
				dy := float64(y - cy)
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < float64(radius) {
					opacity := (1.0 - dist/float64(radius)) * 0.3
					a := uint8(float64(baseColor.A) * opacity)
					img.Set(x, y, color.RGBA{
						R: baseColor.R,
						G: baseColor.G,
						B: baseColor.B,
						A: a,
					})
				}
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateDeadTrees creates dead tree silhouettes.
func generateDeadTrees(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x += 40 + rng.Intn(60) {
		treeHeight := height * 3 / 4

		// Draw twisted trunk
		for y := height - treeHeight; y < height; y++ {
			twist := int(math.Sin(float64(y)*0.1) * 3)
			trunkX := x + twist
			if trunkX >= 0 && trunkX < width {
				img.Set(trunkX, y, baseColor)
				img.Set(trunkX+1, y, baseColor)
			}
		}

		// Add bare branches
		branchY := height - treeHeight*2/3
		for i := 0; i < 3+rng.Intn(3); i++ {
			branchLen := 10 + rng.Intn(15)
			direction := 1
			if rng.Float64() > 0.5 {
				direction = -1
			}
			for bx := 0; bx < branchLen; bx++ {
				px := x + bx*direction
				py := branchY + i*15 + bx/3
				if px >= 0 && px < width && py >= 0 && py < height {
					img.Set(px, py, baseColor)
				}
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateNeonSkyline creates neon-lit buildings.
func generateNeonSkyline(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	x := 0
	for x < width {
		buildingWidth := 25 + rng.Intn(40)
		buildingHeight := height/4 + rng.Intn(height*3/4)

		// Building body
		darkColor := color.RGBA{R: 30, G: 30, B: 40, A: 255}
		for bx := x; bx < x+buildingWidth && bx < width; bx++ {
			for by := height - buildingHeight; by < height; by++ {
				img.Set(bx, by, darkColor)
			}
		}

		// Neon accents
		if rng.Float64() > 0.5 {
			neonY := height - buildingHeight + rng.Intn(buildingHeight/2)
			for bx := x; bx < x+buildingWidth && bx < width; bx++ {
				for ny := neonY; ny < neonY+3; ny++ {
					if ny < height {
						img.Set(bx, ny, baseColor)
					}
				}
			}
		}

		x += buildingWidth + 5 + rng.Intn(10)
	}

	return ebiten.NewImageFromImage(img)
}

// generateRain creates rain streaks.
func generateRain(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	numDrops := width * height / 400
	for i := 0; i < numDrops; i++ {
		x := rng.Intn(width)
		y := rng.Intn(height)
		dropLen := 8 + rng.Intn(12)

		for dy := 0; dy < dropLen; dy++ {
			py := y + dy
			if py < height {
				opacity := 1.0 - float64(dy)/float64(dropLen)
				a := uint8(float64(baseColor.A) * opacity)
				img.Set(x, py, color.RGBA{
					R: baseColor.R,
					G: baseColor.G,
					B: baseColor.B,
					A: a,
				})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// generateBuildings creates building fronts.
func generateBuildings(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	x := 0
	for x < width {
		buildingWidth := 40 + rng.Intn(60)
		drawSingleBuilding(img, x, buildingWidth, width, height, baseColor)
		drawBuildingWindows(img, x, buildingWidth, width, height, rng)
		x += buildingWidth + rng.Intn(15)
	}

	return ebiten.NewImageFromImage(img)
}

// drawSingleBuilding renders a single building silhouette with vertical gradient.
func drawSingleBuilding(img *image.RGBA, x, buildingWidth, width, height int, baseColor color.RGBA) {
	for bx := x; bx < x+buildingWidth && bx < width; bx++ {
		for by := 0; by < height; by++ {
			shade := 0.5 + 0.3*(1.0-float64(by)/float64(height))
			img.Set(bx, by, color.RGBA{
				R: uint8(float64(baseColor.R) * shade),
				G: uint8(float64(baseColor.G) * shade),
				B: uint8(float64(baseColor.B) * shade),
				A: baseColor.A,
			})
		}
	}
}

// drawBuildingWindows adds illuminated windows to a building.
func drawBuildingWindows(img *image.RGBA, x, buildingWidth, width, height int, rng *rand.Rand) {
	windowColor := color.RGBA{R: 200, G: 180, B: 100, A: 200}

	for wy := 20; wy < height-20; wy += 25 {
		for wx := x + 5; wx < x+buildingWidth-5; wx += 12 {
			if wx+8 < width {
				drawSingleWindow(img, wx, wy, height, windowColor, rng)
			}
		}
	}
}

// drawSingleWindow renders a single window with random lit pixels.
func drawSingleWindow(img *image.RGBA, wx, wy, height int, windowColor color.RGBA, rng *rand.Rand) {
	for wdy := 0; wdy < 15; wdy++ {
		for wdx := 0; wdx < 8; wdx++ {
			if wy+wdy < height && rng.Float64() > 0.3 {
				img.Set(wx+wdx, wy+wdy, windowColor)
			}
		}
	}
}

// generateRuins creates ruined building silhouettes.
func generateRuins(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	x := 0
	for x < width {
		ruinWidth := 30 + rng.Intn(50)
		ruinHeight := height/3 + rng.Intn(height*2/3)

		// Jagged top edge for destroyed look
		for bx := x; bx < x+ruinWidth && bx < width; bx++ {
			actualHeight := ruinHeight + rng.Intn(20) - 10
			for by := height - actualHeight; by < height; by++ {
				if rng.Float64() > 0.1 {
					shade := 0.5 + rng.Float64()*0.3
					r := uint8(float64(baseColor.R) * shade)
					g := uint8(float64(baseColor.G) * shade)
					b := uint8(float64(baseColor.B) * shade)
					img.Set(bx, by, color.RGBA{R: r, G: g, B: b, A: baseColor.A})
				}
			}
		}

		x += ruinWidth + 10 + rng.Intn(30)
	}

	return ebiten.NewImageFromImage(img)
}

// generateDust creates dust cloud overlay.
func generateDust(rng *rand.Rand, width, height int, baseColor color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create dust wisps
	numClouds := 5 + rng.Intn(8)
	for i := 0; i < numClouds; i++ {
		cx := rng.Intn(width)
		cy := rng.Intn(height)
		radiusX := 40 + rng.Intn(80)
		radiusY := 20 + rng.Intn(40)

		for y := cy - radiusY; y < cy+radiusY; y++ {
			for x := cx - radiusX; x < cx+radiusX; x++ {
				if x < 0 || x >= width || y < 0 || y >= height {
					continue
				}

				dx := float64(x-cx) / float64(radiusX)
				dy := float64(y-cy) / float64(radiusY)
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < 1.0 {
					opacity := (1.0 - dist) * 0.2
					a := uint8(float64(baseColor.A) * opacity)
					img.Set(x, y, color.RGBA{
						R: baseColor.R,
						G: baseColor.G,
						B: baseColor.B,
						A: a,
					})
				}
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

package dither

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// Material represents the material type for dither pattern selection.
type Material int

const (
	// MaterialDefault uses standard ordered dithering.
	MaterialDefault Material = iota
	// MaterialMetal uses high-contrast specular dithering patterns.
	MaterialMetal
	// MaterialCloth uses soft organic dithering patterns.
	MaterialCloth
	// MaterialLeather uses medium-density natural patterns.
	MaterialLeather
	// MaterialFur uses directional fuzzy dithering.
	MaterialFur
	// MaterialCrystal uses sparse sparkle patterns.
	MaterialCrystal
	// MaterialFlesh uses smooth gradient dithering.
	MaterialFlesh
	// MaterialScales uses regular hexagonal patterns.
	MaterialScales
	// MaterialSlime uses irregular blob patterns.
	MaterialSlime
)

// Direction represents gradient direction for directional dithering.
type Direction int

const (
	// DirVertical applies vertical gradient dithering (top to bottom).
	DirVertical Direction = iota
	// DirHorizontal applies horizontal gradient dithering (left to right).
	DirHorizontal
	// DirRadial applies radial gradient dithering (center to edge).
	DirRadial
	// DirDiagonal applies diagonal gradient dithering.
	DirDiagonal
)

// bayer4x4 is the standard 4x4 Bayer ordered dithering matrix.
var bayer4x4 = [4][4]float64{
	{0.0 / 16.0, 8.0 / 16.0, 2.0 / 16.0, 10.0 / 16.0},
	{12.0 / 16.0, 4.0 / 16.0, 14.0 / 16.0, 6.0 / 16.0},
	{3.0 / 16.0, 11.0 / 16.0, 1.0 / 16.0, 9.0 / 16.0},
	{15.0 / 16.0, 7.0 / 16.0, 13.0 / 16.0, 5.0 / 16.0},
}

// bayer8x8 is the 8x8 Bayer ordered dithering matrix for finer gradients.
var bayer8x8 = [8][8]float64{
	{0.0 / 64.0, 32.0 / 64.0, 8.0 / 64.0, 40.0 / 64.0, 2.0 / 64.0, 34.0 / 64.0, 10.0 / 64.0, 42.0 / 64.0},
	{48.0 / 64.0, 16.0 / 64.0, 56.0 / 64.0, 24.0 / 64.0, 50.0 / 64.0, 18.0 / 64.0, 58.0 / 64.0, 26.0 / 64.0},
	{12.0 / 64.0, 44.0 / 64.0, 4.0 / 64.0, 36.0 / 64.0, 14.0 / 64.0, 46.0 / 64.0, 6.0 / 64.0, 38.0 / 64.0},
	{60.0 / 64.0, 28.0 / 64.0, 52.0 / 64.0, 20.0 / 64.0, 62.0 / 64.0, 30.0 / 64.0, 54.0 / 64.0, 22.0 / 64.0},
	{3.0 / 64.0, 35.0 / 64.0, 11.0 / 64.0, 43.0 / 64.0, 1.0 / 64.0, 33.0 / 64.0, 9.0 / 64.0, 41.0 / 64.0},
	{51.0 / 64.0, 19.0 / 64.0, 59.0 / 64.0, 27.0 / 64.0, 49.0 / 64.0, 17.0 / 64.0, 57.0 / 64.0, 25.0 / 64.0},
	{15.0 / 64.0, 47.0 / 64.0, 7.0 / 64.0, 39.0 / 64.0, 13.0 / 64.0, 45.0 / 64.0, 5.0 / 64.0, 37.0 / 64.0},
	{63.0 / 64.0, 31.0 / 64.0, 55.0 / 64.0, 23.0 / 64.0, 61.0 / 64.0, 29.0 / 64.0, 53.0 / 64.0, 21.0 / 64.0},
}

// System applies dithering patterns to images for enhanced visual depth.
type System struct {
	rng          *rand.Rand
	blueNoise    [][]float64
	blueNoiseGen bool
}

// NewSystem creates a dithering system with the given seed.
func NewSystem(seed int64) *System {
	return &System{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// generateBlueNoise creates a blue noise pattern for organic dithering.
// Blue noise has minimal low-frequency content, appearing more natural.
func (s *System) generateBlueNoise(width, height int) [][]float64 {
	if s.blueNoiseGen && len(s.blueNoise) >= height {
		return s.blueNoise
	}

	noise := make([][]float64, height)
	for y := 0; y < height; y++ {
		noise[y] = make([]float64, width)
		for x := 0; x < width; x++ {
			// Combine multiple noise frequencies for blue noise approximation
			freq1 := math.Sin(float64(x)*0.7+float64(y)*0.3) * 0.3
			freq2 := math.Sin(float64(x)*1.3+float64(y)*0.9+float64(s.rng.Float64())) * 0.2
			freq3 := s.rng.Float64() * 0.5
			noise[y][x] = (freq1 + freq2 + freq3 + 1.0) * 0.5
			noise[y][x] = math.Max(0, math.Min(1, noise[y][x]))
		}
	}

	s.blueNoise = noise
	s.blueNoiseGen = true
	return noise
}

// getBayerThreshold returns the Bayer threshold at a given position.
func getBayerThreshold(x, y int, use8x8 bool) float64 {
	if use8x8 {
		return bayer8x8[y%8][x%8]
	}
	return bayer4x4[y%4][x%4]
}

// ApplyDithering applies material-appropriate dithering to an entire image.
// Intensity controls the strength of the dithering effect (0.0-1.0).
func (s *System) ApplyDithering(img *image.RGBA, material Material, intensity float64) {
	bounds := img.Bounds()
	s.ApplyDitheringToBounds(img, bounds, material, intensity)
}

// ApplyDitheringToBounds applies dithering to a specific region of an image.
func (s *System) ApplyDitheringToBounds(img *image.RGBA, bounds image.Rectangle, material Material, intensity float64) {
	if intensity <= 0 {
		return
	}
	intensity = math.Min(intensity, 1.0)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			c := img.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}

			ditherValue := s.getDitherValue(x, y, material)
			ditheredColor := s.applyDitherToColor(c, ditherValue, intensity, material)
			img.SetRGBA(x, y, ditheredColor)
		}
	}
}

// getDitherValue returns the dither threshold for a position based on material.
func (s *System) getDitherValue(x, y int, material Material) float64 {
	switch material {
	case MaterialMetal:
		// Metallic: Use sharp 4x4 Bayer for crisp specular patterns
		return getBayerThreshold(x, y, false)

	case MaterialCloth, MaterialFur:
		// Organic: Use blue noise for soft, natural transitions
		blueNoise := s.generateBlueNoise(64, 64)
		return blueNoise[y%64][x%64]

	case MaterialCrystal:
		// Crystal: Sparse high-contrast sparkle pattern
		threshold := getBayerThreshold(x, y, true)
		if threshold > 0.75 {
			return 1.0
		}
		return 0.0

	case MaterialSlime:
		// Slime: Irregular blob-like patterns
		blueNoise := s.generateBlueNoise(64, 64)
		baseVal := blueNoise[y%64][x%64]
		// Add blob clustering
		clusterX := float64(x%16) / 16.0
		clusterY := float64(y%16) / 16.0
		cluster := math.Sin(clusterX*math.Pi) * math.Sin(clusterY*math.Pi)
		return baseVal*0.6 + cluster*0.4

	case MaterialScales:
		// Scales: Hexagonal-ish regular pattern
		hexX := (x + (y/2)*3) % 6
		hexY := y % 6
		return bayer8x8[hexY%8][hexX%8]

	case MaterialFlesh, MaterialLeather:
		// Natural materials: 8x8 Bayer for smooth gradients
		return getBayerThreshold(x, y, true)

	default:
		// Default: Standard 4x4 Bayer dithering
		return getBayerThreshold(x, y, false)
	}
}

// applyDitherToColor modifies a color based on the dither threshold.
func (s *System) applyDitherToColor(c color.RGBA, ditherValue, intensity float64, material Material) color.RGBA {
	// Calculate luminance for gradient detection
	luma := float64(c.R)*0.299 + float64(c.G)*0.587 + float64(c.B)*0.114

	// Dithering is most effective at mid-tones
	// At pure black or white, there's nothing to dither
	midtoneFactor := 1.0 - math.Abs(luma-127.5)/127.5
	effectiveIntensity := intensity * midtoneFactor

	// Calculate dither offset
	ditherOffset := (ditherValue - 0.5) * effectiveIntensity * s.getDitherStrength(material)

	// Apply offset to color channels
	r := clampByte(float64(c.R) + ditherOffset*25)
	g := clampByte(float64(c.G) + ditherOffset*25)
	b := clampByte(float64(c.B) + ditherOffset*25)

	return color.RGBA{R: r, G: g, B: b, A: c.A}
}

// getDitherStrength returns the dithering strength multiplier for a material.
func (s *System) getDitherStrength(material Material) float64 {
	switch material {
	case MaterialMetal:
		return 1.5 // Strong for metallic sheen
	case MaterialCrystal:
		return 2.0 // Very strong for sparkle
	case MaterialCloth, MaterialFur:
		return 0.8 // Subtle for organics
	case MaterialFlesh:
		return 0.6 // Very subtle for skin
	case MaterialSlime:
		return 1.2 // Moderate for gooey effect
	default:
		return 1.0
	}
}

// ApplyGradientDithering applies dithering specifically along a gradient direction.
// This creates smooth transitions between two colors.
func (s *System) ApplyGradientDithering(
	img *image.RGBA,
	bounds image.Rectangle,
	startColor, endColor color.RGBA,
	direction Direction,
) {
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	if width <= 0 || height <= 0 {
		return
	}

	centerX := float64(bounds.Min.X + width/2)
	centerY := float64(bounds.Min.Y + height/2)
	maxDist := math.Sqrt(float64(width*width+height*height)) / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			existingAlpha := img.RGBAAt(x, y).A
			if existingAlpha == 0 {
				continue
			}

			// Calculate gradient position (0.0 to 1.0)
			var t float64
			switch direction {
			case DirVertical:
				t = float64(y-bounds.Min.Y) / float64(height)
			case DirHorizontal:
				t = float64(x-bounds.Min.X) / float64(width)
			case DirRadial:
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				t = math.Sqrt(dx*dx+dy*dy) / maxDist
			case DirDiagonal:
				t = (float64(x-bounds.Min.X)/float64(width) + float64(y-bounds.Min.Y)/float64(height)) / 2.0
			}
			t = math.Max(0, math.Min(1, t))

			// Get dither threshold
			threshold := getBayerThreshold(x, y, true)

			// Apply dithered interpolation
			ditheredT := t
			if t > threshold {
				ditheredT = math.Min(1.0, t+0.1)
			} else {
				ditheredT = math.Max(0.0, t-0.1)
			}

			// Interpolate colors
			blendedColor := lerpColor(startColor, endColor, ditheredT)
			blendedColor.A = existingAlpha
			img.SetRGBA(x, y, blendedColor)
		}
	}
}

// ApplyEdgeDithering applies dithering specifically at tonal boundaries.
// This detects edges in the image and concentrates dithering there.
func (s *System) ApplyEdgeDithering(img *image.RGBA, bounds image.Rectangle, intensity float64) {
	if intensity <= 0 {
		return
	}

	// First pass: detect edges using Sobel-like operator
	edgeMap := make([][]float64, bounds.Max.Y-bounds.Min.Y)
	for i := range edgeMap {
		edgeMap[i] = make([]float64, bounds.Max.X-bounds.Min.X)
	}

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			edgeStrength := s.detectEdge(img, x, y)
			edgeMap[y-bounds.Min.Y][x-bounds.Min.X] = edgeStrength
		}
	}

	// Second pass: apply dithering weighted by edge strength
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			c := img.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}

			ey := y - bounds.Min.Y
			ex := x - bounds.Min.X
			if ey >= 0 && ey < len(edgeMap) && ex >= 0 && ex < len(edgeMap[ey]) {
				edgeWeight := edgeMap[ey][ex]
				if edgeWeight > 0.1 {
					ditherValue := getBayerThreshold(x, y, true)
					effectiveIntensity := intensity * edgeWeight
					ditheredColor := s.applyDitherToColor(c, ditherValue, effectiveIntensity, MaterialDefault)
					img.SetRGBA(x, y, ditheredColor)
				}
			}
		}
	}
}

// detectEdge calculates edge strength at a pixel using luminance gradient.
func (s *System) detectEdge(img *image.RGBA, x, y int) float64 {
	// Get 3x3 neighborhood luminances
	var lumas [3][3]float64
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			c := img.RGBAAt(x+dx, y+dy)
			lumas[dy+1][dx+1] = float64(c.R)*0.299 + float64(c.G)*0.587 + float64(c.B)*0.114
		}
	}

	// Sobel operator for edge detection
	gx := -lumas[0][0] + lumas[0][2] - 2*lumas[1][0] + 2*lumas[1][2] - lumas[2][0] + lumas[2][2]
	gy := -lumas[0][0] - 2*lumas[0][1] - lumas[0][2] + lumas[2][0] + 2*lumas[2][1] + lumas[2][2]

	gradient := math.Sqrt(gx*gx + gy*gy)

	// Normalize to 0-1 range (gradient max is around 1020 for pure black/white edge)
	return math.Min(1.0, gradient/200.0)
}

// ApplySpecularDithering applies metallic specular highlight dithering.
// This creates the distinctive metallic sheen pattern.
func (s *System) ApplySpecularDithering(
	img *image.RGBA,
	bounds image.Rectangle,
	highlightDir float64, // Angle in radians for highlight direction
	intensity float64,
) {
	if intensity <= 0 {
		return
	}

	centerX := float64(bounds.Min.X+bounds.Max.X) / 2
	centerY := float64(bounds.Min.Y+bounds.Max.Y) / 2
	radius := float64(bounds.Max.X-bounds.Min.X) / 2

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			c := img.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}

			// Calculate position relative to center
			dx := float64(x) - centerX
			dy := float64(y) - centerY

			// Calculate angle-based specular factor
			pixelAngle := math.Atan2(dy, dx)
			angleDiff := math.Abs(pixelAngle - highlightDir)
			if angleDiff > math.Pi {
				angleDiff = 2*math.Pi - angleDiff
			}

			// Specular highlight falls off with angle difference
			specularFactor := math.Max(0, 1.0-angleDiff/(math.Pi/2))
			specularFactor *= specularFactor // Quadratic falloff for sharper highlight

			// Distance from center affects highlight intensity
			dist := math.Sqrt(dx*dx + dy*dy)
			distFactor := 1.0 - (dist / radius)
			distFactor = math.Max(0, distFactor)

			combinedFactor := specularFactor * distFactor

			if combinedFactor > 0.1 {
				// Apply strong dithering in highlight region
				ditherValue := getBayerThreshold(x, y, false)
				highlight := combinedFactor * intensity * ditherValue

				r := clampByte(float64(c.R) + highlight*80)
				g := clampByte(float64(c.G) + highlight*80)
				b := clampByte(float64(c.B) + highlight*80)

				img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: c.A})
			}
		}
	}
}

// lerpColor linearly interpolates between two colors.
func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

// clampByte clamps a float64 to valid byte range.
func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

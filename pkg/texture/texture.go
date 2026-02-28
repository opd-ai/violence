// Package texture provides procedurally generated texture atlas.
package texture

import (
	"image"
	"image/color"
	"math"

	"github.com/opd-ai/violence/pkg/rng"
)

// Atlas stores procedurally generated textures.
type Atlas struct {
	textures map[string]image.Image
	genre    string
	seed     uint64
}

// NewAtlas creates an empty texture atlas with the given seed.
func NewAtlas(seed uint64) *Atlas {
	return &Atlas{
		textures: make(map[string]image.Image),
		genre:    "fantasy",
		seed:     seed,
	}
}

// Generate procedurally generates a texture and adds it to the atlas.
// Size is the texture dimensions (typically 64, 128, or 256).
// Type determines the generation algorithm: "wall", "floor", "ceiling".
func (a *Atlas) Generate(name string, size int, textureType string) error {
	r := rng.NewRNG(a.seed ^ hashString(name))
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	switch textureType {
	case "wall":
		a.generateWallTexture(img, r)
	case "floor":
		a.generateFloorTexture(img, r)
	case "ceiling":
		a.generateCeilingTexture(img, r)
	default:
		a.generateWallTexture(img, r)
	}

	a.textures[name] = img
	return nil
}

// Get retrieves a texture by name.
func (a *Atlas) Get(name string) (image.Image, bool) {
	img, ok := a.textures[name]
	return img, ok
}

// SetGenre configures texture generation parameters for a genre.
func (a *Atlas) SetGenre(genreID string) {
	a.genre = genreID
}

// generateWallTexture creates a procedural wall texture based on genre.
func (a *Atlas) generateWallTexture(img *image.RGBA, r *rng.RNG) {
	bounds := img.Bounds()
	baseColor := a.getGenreBaseColor()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Multi-octave noise for surface detail
			noise := a.perlinNoise(float64(x)/16.0, float64(y)/16.0, r) * 0.5
			noise += a.perlinNoise(float64(x)/8.0, float64(y)/8.0, r) * 0.25
			noise += a.perlinNoise(float64(x)/4.0, float64(y)/4.0, r) * 0.125

			// Brick pattern for some genres
			if a.genre == "fantasy" || a.genre == "postapoc" {
				brickX := x % 32
				brickY := y % 16
				if brickX == 0 || brickY == 0 {
					noise -= 0.3
				}
			}

			c := a.applyNoise(baseColor, noise)
			img.Set(x, y, c)
		}
	}
}

// generateFloorTexture creates a procedural floor texture.
func (a *Atlas) generateFloorTexture(img *image.RGBA, r *rng.RNG) {
	bounds := img.Bounds()
	baseColor := a.getGenreFloorColor()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			noise := a.perlinNoise(float64(x)/12.0, float64(y)/12.0, r) * 0.4
			noise += a.perlinNoise(float64(x)/6.0, float64(y)/6.0, r) * 0.2

			c := a.applyNoise(baseColor, noise)
			img.Set(x, y, c)
		}
	}
}

// generateCeilingTexture creates a procedural ceiling texture.
func (a *Atlas) generateCeilingTexture(img *image.RGBA, r *rng.RNG) {
	bounds := img.Bounds()
	baseColor := a.getGenreCeilingColor()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			noise := a.perlinNoise(float64(x)/20.0, float64(y)/20.0, r) * 0.3

			c := a.applyNoise(baseColor, noise)
			img.Set(x, y, c)
		}
	}
}

// getGenreBaseColor returns the base wall color for the current genre.
func (a *Atlas) getGenreBaseColor() color.RGBA {
	switch a.genre {
	case "fantasy":
		return color.RGBA{R: 120, G: 100, B: 80, A: 255}
	case "scifi":
		return color.RGBA{R: 90, G: 100, B: 120, A: 255}
	case "horror":
		return color.RGBA{R: 100, G: 95, B: 90, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 80, G: 80, B: 90, A: 255}
	case "postapoc":
		return color.RGBA{R: 110, G: 100, B: 85, A: 255}
	default:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}
}

// getGenreFloorColor returns the base floor color for the current genre.
func (a *Atlas) getGenreFloorColor() color.RGBA {
	switch a.genre {
	case "fantasy":
		return color.RGBA{R: 80, G: 70, B: 60, A: 255}
	case "scifi":
		return color.RGBA{R: 60, G: 70, B: 80, A: 255}
	case "horror":
		return color.RGBA{R: 70, G: 65, B: 60, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 50, G: 50, B: 60, A: 255}
	case "postapoc":
		return color.RGBA{R: 90, G: 80, B: 70, A: 255}
	default:
		return color.RGBA{R: 70, G: 70, B: 70, A: 255}
	}
}

// getGenreCeilingColor returns the base ceiling color for the current genre.
func (a *Atlas) getGenreCeilingColor() color.RGBA {
	switch a.genre {
	case "fantasy":
		return color.RGBA{R: 60, G: 50, B: 40, A: 255}
	case "scifi":
		return color.RGBA{R: 50, G: 60, B: 70, A: 255}
	case "horror":
		return color.RGBA{R: 50, G: 45, B: 40, A: 255}
	case "cyberpunk":
		return color.RGBA{R: 40, G: 40, B: 50, A: 255}
	case "postapoc":
		return color.RGBA{R: 70, G: 60, B: 50, A: 255}
	default:
		return color.RGBA{R: 50, G: 50, B: 50, A: 255}
	}
}

// applyNoise applies noise value to base color.
func (a *Atlas) applyNoise(base color.RGBA, noise float64) color.RGBA {
	factor := 1.0 + noise
	return color.RGBA{
		R: clampUint8(float64(base.R) * factor),
		G: clampUint8(float64(base.G) * factor),
		B: clampUint8(float64(base.B) * factor),
		A: base.A,
	}
}

// perlinNoise generates simplified Perlin-like noise.
// This uses interpolated random gradients for texture variation.
func (a *Atlas) perlinNoise(x, y float64, r *rng.RNG) float64 {
	// Grid cell coordinates
	x0 := math.Floor(x)
	y0 := math.Floor(y)

	// Fractional offsets
	dx := x - x0
	dy := y - y0

	// Fade curves for smooth interpolation
	u := fade(dx)
	v := fade(dy)

	// Hash coordinates to get gradient indices
	ix0 := int(x0)
	iy0 := int(y0)

	// Generate corner gradients deterministically
	gx00, gy00 := a.gradient(ix0, iy0, r)
	gx10, gy10 := a.gradient(ix0+1, iy0, r)
	gx01, gy01 := a.gradient(ix0, iy0+1, r)
	gx11, gy11 := a.gradient(ix0+1, iy0+1, r)

	// Dot products with distance vectors
	n00 := dot2(gx00, gy00, dx, dy)
	n10 := dot2(gx10, gy10, dx-1, dy)
	n01 := dot2(gx01, gy01, dx, dy-1)
	n11 := dot2(gx11, gy11, dx-1, dy-1)

	// Bilinear interpolation
	nx0 := lerp(n00, n10, u)
	nx1 := lerp(n01, n11, u)

	return lerp(nx0, nx1, v)
}

// gradient returns a deterministic 2D gradient vector for grid point (x, y).
func (a *Atlas) gradient(x, y int, r *rng.RNG) (float64, float64) {
	// Hash grid coordinates
	h := hashCoord(x, y) ^ a.seed
	angle := float64(h%360) * math.Pi / 180.0
	return math.Cos(angle), math.Sin(angle)
}

// hashString converts a string to a uint64 hash.
func hashString(s string) uint64 {
	var h uint64 = 5381
	for _, c := range s {
		h = ((h << 5) + h) + uint64(c)
	}
	return h
}

// hashCoord generates a hash from grid coordinates.
func hashCoord(x, y int) uint64 {
	const prime1 uint64 = 73856093
	const prime2 uint64 = 19349663
	return uint64(x)*prime1 ^ uint64(y)*prime2
}

// fade applies smoothstep fade curve.
func fade(t float64) float64 {
	return t * t * t * (t*(t*6-15) + 10)
}

// lerp performs linear interpolation.
func lerp(a, b, t float64) float64 {
	return a + t*(b-a)
}

// dot2 computes 2D dot product.
func dot2(gx, gy, x, y float64) float64 {
	return gx*x + gy*y
}

// clampUint8 clamps a float64 to uint8 range.
func clampUint8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// getRNG returns a new RNG seeded from the atlas seed.
// Used for testing to get deterministic noise generation.
func (a *Atlas) getRNG() *rng.RNG {
	return rng.NewRNG(a.seed)
}

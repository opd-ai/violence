package rimlight

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/pool"
	"github.com/sirupsen/logrus"
)

// System applies directional rim lighting to sprites.
type System struct {
	genreID    string
	logger     *logrus.Entry
	cache      map[cacheKey]*ebiten.Image
	maxCache   int
	lightDirX  float64
	lightDirY  float64
	rimColor   color.RGBA
	magicColor color.RGBA
}

type cacheKey struct {
	imageID   uint64
	material  Material
	intensity float64
	width     int
}

// NewSystem creates a rim lighting system.
func NewSystem(genreID string) *System {
	s := &System{
		genreID:  genreID,
		cache:    make(map[cacheKey]*ebiten.Image),
		maxCache: 150,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "RimLightSystem",
			"package":     "rimlight",
		}),
	}
	s.setGenreConfig()
	return s
}

// setGenreConfig configures rim lighting parameters based on genre.
func (s *System) setGenreConfig() {
	switch s.genreID {
	case "scifi":
		// Cool blue rim, light from upper-right for tech feel
		s.lightDirX = 0.6
		s.lightDirY = -0.8
		s.rimColor = color.RGBA{R: 140, G: 200, B: 255, A: 255}
		s.magicColor = color.RGBA{R: 100, G: 220, B: 255, A: 255}
	case "horror":
		// Dim greenish rim, harsh angle
		s.lightDirX = -0.7
		s.lightDirY = -0.7
		s.rimColor = color.RGBA{R: 120, G: 150, B: 130, A: 255}
		s.magicColor = color.RGBA{R: 80, G: 200, B: 100, A: 255}
	case "cyberpunk":
		// Neon magenta/cyan rim
		s.lightDirX = 0.5
		s.lightDirY = -0.85
		s.rimColor = color.RGBA{R: 255, G: 100, B: 200, A: 255}
		s.magicColor = color.RGBA{R: 0, G: 255, B: 255, A: 255}
	case "postapoc":
		// Dusty orange rim, low sun angle
		s.lightDirX = 0.8
		s.lightDirY = -0.6
		s.rimColor = color.RGBA{R: 255, G: 180, B: 120, A: 255}
		s.magicColor = color.RGBA{R: 255, G: 150, B: 50, A: 255}
	default: // fantasy
		// Warm golden rim, classic top-left light
		s.lightDirX = -0.5
		s.lightDirY = -0.85
		s.rimColor = color.RGBA{R: 255, G: 240, B: 200, A: 255}
		s.magicColor = color.RGBA{R: 180, G: 150, B: 255, A: 255}
	}

	// Normalize light direction
	length := math.Sqrt(s.lightDirX*s.lightDirX + s.lightDirY*s.lightDirY)
	if length > 0 {
		s.lightDirX /= length
		s.lightDirY /= length
	}
}

// SetGenre updates the genre and reconfigures rim lighting.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.setGenreConfig()
	s.cache = make(map[cacheKey]*ebiten.Image)
}

// Update processes entities with rim light components.
func (s *System) Update(w *engine.World) {
	rimType := reflect.TypeOf(&Component{})
	entities := w.Query(rimType)

	s.logger.WithFields(logrus.Fields{
		"entity_count":  len(entities),
		"cache_entries": len(s.cache),
	}).Trace("Processing rim lighting")

	// Evict cache if too large
	if len(s.cache) > s.maxCache {
		s.cache = make(map[cacheKey]*ebiten.Image)
	}
}

// ApplyRimLight processes a sprite and returns a version with rim lighting applied.
func (s *System) ApplyRimLight(src *ebiten.Image, comp *Component) *ebiten.Image {
	if src == nil || comp == nil || !comp.Enabled {
		return src
	}

	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return src
	}

	// Generate cache key
	imageID := uint64(w)<<32 | uint64(h)<<16 | uint64(comp.Material)<<8
	key := cacheKey{
		imageID:   imageID,
		material:  comp.Material,
		intensity: comp.Intensity,
		width:     comp.Width,
	}

	if cached, found := s.cache[key]; found {
		return cached
	}

	// Read source pixels
	srcRGBA := pool.GlobalPools.Images.Get(w, h)
	defer pool.GlobalPools.Images.Put(srcRGBA)
	src.ReadPixels(srcRGBA.Pix)

	// Create output buffer
	dstRGBA := pool.GlobalPools.Images.Get(w, h)

	// Calculate rim width (auto-size based on sprite dimensions)
	rimWidth := comp.Width
	if rimWidth <= 0 {
		rimWidth = max(2, min(w, h)/8)
	}

	// Get effective rim color
	rimColor := s.getRimColor(comp)

	// Calculate effective intensity
	intensity := comp.Intensity * GetMaterialIntensity(comp.Material)
	fresnel := GetMaterialFresnel(comp.Material)

	// Apply rim lighting
	s.processRimLighting(srcRGBA, dstRGBA, rimColor, rimWidth, intensity, fresnel, comp.FadeInner)

	result := ebiten.NewImageFromImage(dstRGBA)
	pool.GlobalPools.Images.Put(dstRGBA)

	s.cache[key] = result
	return result
}

// getRimColor returns the appropriate rim color for a component.
func (s *System) getRimColor(comp *Component) color.RGBA {
	// Use component override if set
	if comp.Color.A > 0 {
		return comp.Color
	}

	// Magic material uses special color
	if comp.Material == MaterialMagic {
		return s.magicColor
	}

	return s.rimColor
}

// processRimLighting applies the rim lighting effect to the image.
func (s *System) processRimLighting(src, dst *image.RGBA, rimColor color.RGBA, rimWidth int, intensity, fresnel, fadeInner float64) {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	cx, cy := float64(w)/2.0, float64(h)/2.0

	// First pass: copy source and detect edges
	edgeMap := make([]float64, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*src.Stride + x*4
			srcAlpha := src.Pix[idx+3]

			// Copy source pixel
			dst.Pix[idx] = src.Pix[idx]
			dst.Pix[idx+1] = src.Pix[idx+1]
			dst.Pix[idx+2] = src.Pix[idx+2]
			dst.Pix[idx+3] = src.Pix[idx+3]

			if srcAlpha < 128 {
				continue
			}

			// Calculate edge distance (distance to nearest transparent pixel)
			edgeDist := s.calculateEdgeDistance(src, x, y, w, h, rimWidth)
			edgeMap[y*w+x] = edgeDist
		}
	}

	// Second pass: apply rim lighting based on edge distance and light direction
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*src.Stride + x*4
			if src.Pix[idx+3] < 128 {
				continue
			}

			edgeDist := edgeMap[y*w+x]
			if edgeDist >= float64(rimWidth) {
				continue
			}

			// Calculate normal from position relative to center
			nx := (float64(x) - cx) / cx
			ny := (float64(y) - cy) / cy

			// Normalize
			nlen := math.Sqrt(nx*nx + ny*ny)
			if nlen > 0 {
				nx /= nlen
				ny /= nlen
			}

			// Calculate rim intensity using Fresnel-like calculation
			// Rim is brightest where surface faces away from light (dot product with -light)
			dotLight := -(nx*s.lightDirX + ny*s.lightDirY)
			dotLight = math.Max(0, dotLight)

			// Fresnel falloff at edges
			edgeFactor := 1.0 - (edgeDist / float64(rimWidth))
			edgeFactor = math.Pow(edgeFactor, fadeInner*2.0)

			// Combined rim intensity
			rimIntensity := dotLight * edgeFactor * intensity
			rimIntensity = math.Pow(rimIntensity, 1.0/fresnel)
			rimIntensity = math.Min(1.0, rimIntensity)

			if rimIntensity < 0.05 {
				continue
			}

			// Blend rim color with source
			srcR := float64(src.Pix[idx])
			srcG := float64(src.Pix[idx+1])
			srcB := float64(src.Pix[idx+2])

			rimR := float64(rimColor.R)
			rimG := float64(rimColor.G)
			rimB := float64(rimColor.B)

			// Additive blend for rim highlight
			outR := srcR + rimR*rimIntensity*0.7
			outG := srcG + rimG*rimIntensity*0.7
			outB := srcB + rimB*rimIntensity*0.7

			dst.Pix[idx] = uint8(math.Min(255, outR))
			dst.Pix[idx+1] = uint8(math.Min(255, outG))
			dst.Pix[idx+2] = uint8(math.Min(255, outB))
		}
	}
}

// calculateEdgeDistance finds the distance to the nearest transparent pixel.
func (s *System) calculateEdgeDistance(img *image.RGBA, x, y, w, h, maxDist int) float64 {
	minDist := float64(maxDist + 1)

	for dy := -maxDist; dy <= maxDist; dy++ {
		for dx := -maxDist; dx <= maxDist; dx++ {
			nx, ny := x+dx, y+dy
			if nx < 0 || nx >= w || ny < 0 || ny >= h {
				// Treat out-of-bounds as transparent edge
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < minDist {
					minDist = dist
				}
				continue
			}

			idx := ny*img.Stride + nx*4
			if img.Pix[idx+3] < 128 {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < minDist {
					minDist = dist
				}
			}
		}
	}

	return minDist
}

// ApplyRimLightToImage is a convenience function for direct image processing.
func (s *System) ApplyRimLightToImage(img *image.RGBA, mat Material, intensity float64) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w == 0 || h == 0 {
		return
	}

	// Create temporary buffer
	temp := pool.GlobalPools.Images.Get(w, h)
	defer pool.GlobalPools.Images.Put(temp)

	rimWidth := max(2, min(w, h)/8)
	rimColor := s.rimColor
	if mat == MaterialMagic {
		rimColor = s.magicColor
	}

	effectiveIntensity := intensity * GetMaterialIntensity(mat)
	fresnel := GetMaterialFresnel(mat)

	s.processRimLighting(img, temp, rimColor, rimWidth, effectiveIntensity, fresnel, 0.5)

	// Copy back to original
	copy(img.Pix, temp.Pix)
}

// GetLightDirection returns the current rim light direction.
func (s *System) GetLightDirection() (x, y float64) {
	return s.lightDirX, s.lightDirY
}

// SetLightDirection sets a custom rim light direction (will be normalized).
func (s *System) SetLightDirection(x, y float64) {
	length := math.Sqrt(x*x + y*y)
	if length > 0 {
		s.lightDirX = x / length
		s.lightDirY = y / length
	}
}

// GetRimColor returns the current rim light color.
func (s *System) GetRimColor() color.RGBA {
	return s.rimColor
}

// GenerateSeedVariant creates a slightly varied rim light configuration.
func (s *System) GenerateSeedVariant(seed int64, baseComp *Component) *Component {
	rng := rand.New(rand.NewSource(seed))

	variant := NewComponent()
	variant.Enabled = true
	variant.Material = baseComp.Material

	// Slight intensity variation (±20%)
	variant.Intensity = baseComp.Intensity * (0.8 + rng.Float64()*0.4)

	// Slight width variation
	variant.Width = baseComp.Width
	if variant.Width > 2 {
		variant.Width += rng.Intn(3) - 1
	}

	variant.FadeInner = baseComp.FadeInner

	return variant
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

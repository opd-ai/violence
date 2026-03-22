package subsurface

import (
	"image"
	"image/color"
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/sirupsen/logrus"
)

// System applies subsurface scattering effects to organic sprites.
type System struct {
	genre           string
	lightDirX       float64
	lightDirY       float64
	lightDirZ       float64
	lightIntensity  float64
	ambientStrength float64
	thicknessCache  map[uint64]*thicknessMap
	maxCacheSize    int
}

// thicknessMap stores precomputed thickness values for a sprite.
type thicknessMap struct {
	width     int
	height    int
	thickness [][]float64
	edges     [][]bool
}

// NewSystem creates a new subsurface scattering system.
func NewSystem() *System {
	return &System{
		genre:           genre.Fantasy,
		lightDirX:       -0.5773,
		lightDirY:       -0.5773,
		lightDirZ:       0.5773,
		lightIntensity:  1.0,
		ambientStrength: 0.3,
		thicknessCache:  make(map[uint64]*thicknessMap),
		maxCacheSize:    100,
	}
}

// SetGenre updates the system's genre for genre-specific adjustments.
func (s *System) SetGenre(genreID string) {
	s.genre = genreID
	switch genreID {
	case genre.Fantasy:
		s.lightIntensity = 1.0
		s.ambientStrength = 0.3
	case genre.SciFi:
		s.lightIntensity = 0.8
		s.ambientStrength = 0.4
	case genre.Horror:
		s.lightIntensity = 0.6
		s.ambientStrength = 0.2
	case genre.Cyberpunk:
		s.lightIntensity = 0.9
		s.ambientStrength = 0.35
	case genre.PostApoc:
		s.lightIntensity = 1.1
		s.ambientStrength = 0.25
	default:
		s.lightIntensity = 1.0
		s.ambientStrength = 0.3
	}
}

// SetLightDirection sets the global light direction for SSS calculations.
func (s *System) SetLightDirection(x, y, z float64) {
	length := math.Sqrt(x*x + y*y + z*z)
	if length > 0 {
		s.lightDirX = x / length
		s.lightDirY = y / length
		s.lightDirZ = z / length
	}
}

// Update processes all entities with subsurface components.
func (s *System) Update(w *engine.World) {
	componentType := reflect.TypeOf(&Component{})
	entities := w.Query(componentType)

	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, componentType)
		if !ok {
			continue
		}

		sssComp, ok := comp.(*Component)
		if !ok || !sssComp.Enabled {
			continue
		}

		logrus.WithFields(logrus.Fields{
			"system": "subsurface",
			"entity": entity,
		}).Trace("Processing SSS for entity")
	}
}

// ApplySSS applies subsurface scattering to an image for a given material.
// This is the main entry point for sprite generation pipelines.
func (s *System) ApplySSS(img *image.RGBA, mat Material, intensity float64) {
	if intensity <= 0 {
		return
	}
	intensity = math.Min(intensity, 2.0)

	bounds := img.Bounds()
	profile := GetScatterProfile(mat)

	thickness := s.computeThickness(img, bounds)

	s.applyScattering(img, bounds, profile, thickness, intensity)
}

// ApplySSSToRegion applies subsurface scattering to a specific region.
func (s *System) ApplySSSToRegion(img *image.RGBA, bounds image.Rectangle, mat Material, intensity float64) {
	if intensity <= 0 {
		return
	}
	intensity = math.Min(intensity, 2.0)

	profile := GetScatterProfile(mat)
	thickness := s.computeThickness(img, bounds)

	s.applyScattering(img, bounds, profile, thickness, intensity)
}

// computeThickness calculates a thickness map from the sprite silhouette.
// Uses distance transform from edges to estimate material depth at each pixel.
func (s *System) computeThickness(img *image.RGBA, bounds image.Rectangle) *thicknessMap {
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	if width <= 0 || height <= 0 {
		return nil
	}

	tm := &thicknessMap{
		width:     width,
		height:    height,
		thickness: make([][]float64, height),
		edges:     make([][]bool, height),
	}

	for y := 0; y < height; y++ {
		tm.thickness[y] = make([]float64, width)
		tm.edges[y] = make([]bool, width)
	}

	alpha := make([][]uint8, height)
	for y := 0; y < height; y++ {
		alpha[y] = make([]uint8, width)
		for x := 0; x < width; x++ {
			px := x + bounds.Min.X
			py := y + bounds.Min.Y
			if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
				c := img.RGBAAt(px, py)
				alpha[y][x] = c.A
			}
		}
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			if alpha[y][x] > 0 {
				isEdge := alpha[y-1][x] == 0 || alpha[y+1][x] == 0 ||
					alpha[y][x-1] == 0 || alpha[y][x+1] == 0
				tm.edges[y][x] = isEdge
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if alpha[y][x] == 0 {
				continue
			}

			minDist := float64(width + height)
			for ey := 0; ey < height; ey++ {
				for ex := 0; ex < width; ex++ {
					if tm.edges[ey][ex] {
						dx := float64(x - ex)
						dy := float64(y - ey)
						dist := math.Sqrt(dx*dx + dy*dy)
						if dist < minDist {
							minDist = dist
						}
					}
				}
			}

			maxThickness := float64(min(width, height)) / 2.0
			if maxThickness > 0 {
				tm.thickness[y][x] = math.Min(minDist/maxThickness, 1.0)
			}
		}
	}

	return tm
}

// applyScattering applies the SSS effect based on thickness and material profile.
func (s *System) applyScattering(img *image.RGBA, bounds image.Rectangle, profile ScatterProfile, tm *thicknessMap, intensity float64) {
	if tm == nil {
		return
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if x < 0 || x >= img.Bounds().Dx() || y < 0 || y >= img.Bounds().Dy() {
				continue
			}

			c := img.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}

			localY := y - bounds.Min.Y
			localX := x - bounds.Min.X
			if localY < 0 || localY >= tm.height || localX < 0 || localX >= tm.width {
				continue
			}

			thickness := tm.thickness[localY][localX]
			isEdge := tm.edges[localY][localX]

			sssColor := s.computeSSSColor(c, profile, thickness, isEdge, intensity)
			img.SetRGBA(x, y, sssColor)
		}
	}
}

// computeSSSColor calculates the final color with SSS applied.
func (s *System) computeSSSColor(original color.RGBA, profile ScatterProfile, thickness float64, isEdge bool, intensity float64) color.RGBA {
	if thickness <= 0 {
		return original
	}

	thinnessFactor := 1.0 - thickness
	depthFactor := math.Pow(thickness, 0.5)

	scatterAmount := (1.0 - depthFactor) * intensity * s.lightIntensity

	if isEdge {
		scatterAmount *= (1.0 + profile.Translucency)
	}

	transmissionR := math.Exp(-profile.AbsorptionR * (1.0 - thickness) * profile.ScatterDistance)
	transmissionG := math.Exp(-profile.AbsorptionG * (1.0 - thickness) * profile.ScatterDistance)
	transmissionB := math.Exp(-profile.AbsorptionB * (1.0 - thickness) * profile.ScatterDistance)

	baseR := float64(original.R)
	baseG := float64(original.G)
	baseB := float64(original.B)

	scatterR := float64(profile.ScatterColor.R)
	scatterG := float64(profile.ScatterColor.G)
	scatterB := float64(profile.ScatterColor.B)

	sssR := baseR*transmissionR + scatterR*scatterAmount*thinnessFactor
	sssG := baseG*transmissionG + scatterG*scatterAmount*thinnessFactor*0.7
	sssB := baseB*transmissionB + scatterB*scatterAmount*thinnessFactor*0.5

	backscatter := profile.BackscatterStrength * thinnessFactor * intensity
	sssR += backscatter * scatterR * 0.2
	sssG += backscatter * scatterG * 0.15
	sssB += backscatter * scatterB * 0.1

	if isEdge && profile.Translucency > 0.3 {
		transIntensity := profile.Translucency * thinnessFactor * 0.5
		sssR = sssR*(1.0-transIntensity) + scatterR*transIntensity
		sssG = sssG*(1.0-transIntensity) + scatterG*transIntensity
		sssB = sssB*(1.0-transIntensity) + scatterB*transIntensity
	}

	return color.RGBA{
		R: clampByte(sssR),
		G: clampByte(sssG),
		B: clampByte(sssB),
		A: original.A,
	}
}

// RenderSSSDebug renders a debug visualization of the thickness map.
func (s *System) RenderSSSDebug(img *image.RGBA, bounds image.Rectangle) *image.RGBA {
	tm := s.computeThickness(img, bounds)
	if tm == nil {
		return img
	}

	debug := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			localY := y - bounds.Min.Y
			localX := x - bounds.Min.X
			if localY < 0 || localY >= tm.height || localX < 0 || localX >= tm.width {
				continue
			}

			thickness := tm.thickness[localY][localX]
			isEdge := tm.edges[localY][localX]

			if isEdge {
				debug.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			} else if thickness > 0 {
				v := uint8(thickness * 255)
				debug.SetRGBA(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
			}
		}
	}

	return debug
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

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package bouncelight

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines bounce lighting behavior per genre.
type GenrePreset struct {
	// BounceStrength controls overall indirect illumination intensity
	BounceStrength float64

	// MaxBounceDistance is the maximum distance for color bleeding (in world units)
	MaxBounceDistance float64

	// WallContribution is how much walls contribute bounce light
	WallContribution float64

	// FloorContribution is how much floors contribute bounce light
	FloorContribution float64

	// SaturationBoost increases color saturation of bounced light
	SaturationBoost float64

	// WarmShift adds warm tint to all bounced light (simulates multiple bounces)
	WarmShift float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BounceStrength:    0.35,
		MaxBounceDistance: 4.0,
		WallContribution:  0.6,
		FloorContribution: 0.3,
		SaturationBoost:   0.1,
		WarmShift:         0.15, // Torchlit dungeons are warm
	},
	"scifi": {
		BounceStrength:    0.25,
		MaxBounceDistance: 3.0,
		WallContribution:  0.4,
		FloorContribution: 0.2,
		SaturationBoost:   0.0,
		WarmShift:         -0.1, // Cool tech environments
	},
	"horror": {
		BounceStrength:    0.45, // Stronger bounce for atmosphere
		MaxBounceDistance: 5.0,
		WallContribution:  0.7,
		FloorContribution: 0.4,
		SaturationBoost:   -0.1, // Desaturated for dread
		WarmShift:         0.0,
	},
	"cyberpunk": {
		BounceStrength:    0.5, // Strong neon reflections
		MaxBounceDistance: 4.5,
		WallContribution:  0.5,
		FloorContribution: 0.6, // Wet floors reflect more
		SaturationBoost:   0.2, // Vivid neon colors
		WarmShift:         0.0,
	},
	"postapoc": {
		BounceStrength:    0.3,
		MaxBounceDistance: 3.5,
		WallContribution:  0.5,
		FloorContribution: 0.3,
		SaturationBoost:   -0.15, // Dusty, desaturated
		WarmShift:         0.1,   // Slightly warm from fires
	},
}

// System manages bounce lighting calculations and rendering.
type System struct {
	genreID string
	preset  GenrePreset
	logger  *logrus.Entry

	// Cached bounce map per room/sector
	bounceCache map[bounceCacheKey]*BounceMap
	cacheLimit  int

	// Screen dimensions for overlay rendering
	screenW, screenH int

	// Overlay image for blending (reused to avoid allocation)
	overlay *ebiten.Image
}

// bounceCacheKey identifies a unique bounce configuration.
type bounceCacheKey struct {
	roomID   int
	lightCfg uint64 // Hash of light positions
}

// BounceMap stores precomputed bounce values for a grid.
type BounceMap struct {
	Width, Height int
	TileSize      float64
	Data          []BounceCell
}

// BounceCell stores bounce data for a single grid cell.
type BounceCell struct {
	R, G, B   float64
	Intensity float64
}

// NewSystem creates a bounce lighting system for the specified genre.
func NewSystem(genreID string, screenW, screenH int) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	logger := logrus.WithFields(logrus.Fields{
		"system": "bouncelight",
		"genre":  genreID,
	})

	s := &System{
		genreID:     genreID,
		preset:      preset,
		logger:      logger,
		bounceCache: make(map[bounceCacheKey]*BounceMap),
		cacheLimit:  50,
		screenW:     screenW,
		screenH:     screenH,
	}

	logger.Debug("Bounce lighting system initialized")
	return s
}

// SetGenre updates the system for a new genre.
func (s *System) SetGenre(genreID string) {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}
	s.genreID = genreID
	s.preset = preset

	// Clear cache when genre changes
	s.bounceCache = make(map[bounceCacheKey]*BounceMap)

	s.logger = logrus.WithFields(logrus.Fields{
		"system": "bouncelight",
		"genre":  genreID,
	})
}

// GetPreset returns the current genre preset.
func (s *System) GetPreset() GenrePreset {
	return s.preset
}

// CalculateBounce computes bounce lighting for a grid of surfaces.
// surfaces are the contributing surfaces, gridW/gridH are grid dimensions,
// tileSize is the world-space size of each grid cell.
func (s *System) CalculateBounce(surfaces []BounceSurface, gridW, gridH int, tileSize float64) *BounceMap {
	bounceMap := &BounceMap{
		Width:    gridW,
		Height:   gridH,
		TileSize: tileSize,
		Data:     make([]BounceCell, gridW*gridH),
	}

	maxDist := s.preset.MaxBounceDistance

	// For each cell in the grid
	for gy := 0; gy < gridH; gy++ {
		for gx := 0; gx < gridW; gx++ {
			idx := gy*gridW + gx
			cell := &bounceMap.Data[idx]

			// World position of cell center
			wx := (float64(gx) + 0.5) * tileSize
			wy := (float64(gy) + 0.5) * tileSize

			// Accumulate contributions from all surfaces
			var totalR, totalG, totalB, totalIntensity float64

			for i := range surfaces {
				surf := &surfaces[i]

				// Get contribution from this surface
				r, g, b, intensity := surf.BounceContribution(wx, wy, maxDist)
				if intensity <= 0 {
					continue
				}

				// Apply wall/floor contribution factor
				if surf.IsWall {
					intensity *= s.preset.WallContribution
				} else {
					intensity *= s.preset.FloorContribution
				}

				totalR += r * intensity
				totalG += g * intensity
				totalB += b * intensity
				totalIntensity += intensity
			}

			// Normalize and store
			if totalIntensity > 0 {
				cell.R = totalR / totalIntensity
				cell.G = totalG / totalIntensity
				cell.B = totalB / totalIntensity
				cell.Intensity = totalIntensity * s.preset.BounceStrength

				// Apply saturation boost
				if s.preset.SaturationBoost != 0 {
					cell.R, cell.G, cell.B = s.adjustSaturation(
						cell.R, cell.G, cell.B,
						1.0+s.preset.SaturationBoost,
					)
				}

				// Apply warm shift
				if s.preset.WarmShift != 0 {
					cell.R += s.preset.WarmShift * 0.3
					cell.G += s.preset.WarmShift * 0.1
					cell.B -= s.preset.WarmShift * 0.2
				}

				// Clamp values
				cell.R = clamp01(cell.R)
				cell.G = clamp01(cell.G)
				cell.B = clamp01(cell.B)
				if cell.Intensity > 1.0 {
					cell.Intensity = 1.0
				}
			}
		}
	}

	return bounceMap
}

// GetBounceAt returns the bounce lighting at a world position.
func (s *System) GetBounceAt(bounceMap *BounceMap, worldX, worldY float64) (r, g, b, intensity float64) {
	if bounceMap == nil {
		return 0, 0, 0, 0
	}

	// Convert world position to grid coordinates
	gx := int(worldX / bounceMap.TileSize)
	gy := int(worldY / bounceMap.TileSize)

	// Bounds check
	if gx < 0 || gx >= bounceMap.Width || gy < 0 || gy >= bounceMap.Height {
		return 0, 0, 0, 0
	}

	idx := gy*bounceMap.Width + gx
	cell := &bounceMap.Data[idx]
	return cell.R, cell.G, cell.B, cell.Intensity
}

// GetBounceAtBilinear returns bilinearly interpolated bounce lighting.
func (s *System) GetBounceAtBilinear(bounceMap *BounceMap, worldX, worldY float64) (r, g, b, intensity float64) {
	if bounceMap == nil {
		return 0, 0, 0, 0
	}

	// Convert to fractional grid coordinates
	fx := worldX/bounceMap.TileSize - 0.5
	fy := worldY/bounceMap.TileSize - 0.5

	// Get integer coordinates and fractions
	gx0 := int(math.Floor(fx))
	gy0 := int(math.Floor(fy))
	gx1 := gx0 + 1
	gy1 := gy0 + 1

	fracX := fx - float64(gx0)
	fracY := fy - float64(gy0)

	// Sample four corners
	r00, g00, b00, i00 := s.sampleCell(bounceMap, gx0, gy0)
	r10, g10, b10, i10 := s.sampleCell(bounceMap, gx1, gy0)
	r01, g01, b01, i01 := s.sampleCell(bounceMap, gx0, gy1)
	r11, g11, b11, i11 := s.sampleCell(bounceMap, gx1, gy1)

	// Bilinear interpolation
	r = lerp(lerp(r00, r10, fracX), lerp(r01, r11, fracX), fracY)
	g = lerp(lerp(g00, g10, fracX), lerp(g01, g11, fracX), fracY)
	b = lerp(lerp(b00, b10, fracX), lerp(b01, b11, fracX), fracY)
	intensity = lerp(lerp(i00, i10, fracX), lerp(i01, i11, fracX), fracY)

	return r, g, b, intensity
}

// sampleCell samples a bounce map cell with bounds clamping.
func (s *System) sampleCell(bounceMap *BounceMap, gx, gy int) (r, g, b, intensity float64) {
	// Clamp to bounds
	if gx < 0 {
		gx = 0
	}
	if gx >= bounceMap.Width {
		gx = bounceMap.Width - 1
	}
	if gy < 0 {
		gy = 0
	}
	if gy >= bounceMap.Height {
		gy = bounceMap.Height - 1
	}

	idx := gy*bounceMap.Width + gx
	cell := &bounceMap.Data[idx]
	return cell.R, cell.G, cell.B, cell.Intensity
}

// ApplyBounceToColor blends bounce lighting into a surface color.
func (s *System) ApplyBounceToColor(baseColor color.RGBA, bounceR, bounceG, bounceB, bounceIntensity float64) color.RGBA {
	if bounceIntensity <= 0 {
		return baseColor
	}

	// Extract base color components
	br := float64(baseColor.R) / 255.0
	bg := float64(baseColor.G) / 255.0
	bb := float64(baseColor.B) / 255.0

	// Additive blend with bounce light (like real indirect illumination)
	br += bounceR * bounceIntensity
	bg += bounceG * bounceIntensity
	bb += bounceB * bounceIntensity

	// Clamp to valid range
	return color.RGBA{
		R: uint8(clamp01(br) * 255),
		G: uint8(clamp01(bg) * 255),
		B: uint8(clamp01(bb) * 255),
		A: baseColor.A,
	}
}

// RenderBounceOverlay renders a bounce map as a screen overlay.
// The overlay is blended additively with the scene.
func (s *System) RenderBounceOverlay(screen *ebiten.Image, bounceMap *BounceMap, cameraX, cameraY, scale float64) {
	if bounceMap == nil {
		return
	}

	// Create or reuse overlay image
	if s.overlay == nil || s.overlay.Bounds().Dx() != s.screenW || s.overlay.Bounds().Dy() != s.screenH {
		s.overlay = ebiten.NewImage(s.screenW, s.screenH)
	}
	s.overlay.Clear()

	tileScreenSize := int(bounceMap.TileSize * scale)
	if tileScreenSize < 1 {
		tileScreenSize = 1
	}

	// Calculate visible range
	startGX := int(cameraX / bounceMap.TileSize)
	startGY := int(cameraY / bounceMap.TileSize)
	tilesX := s.screenW/tileScreenSize + 2
	tilesY := s.screenH/tileScreenSize + 2

	// Render visible cells to overlay
	for dy := 0; dy < tilesY; dy++ {
		for dx := 0; dx < tilesX; dx++ {
			gx := startGX + dx
			gy := startGY + dy

			if gx < 0 || gx >= bounceMap.Width || gy < 0 || gy >= bounceMap.Height {
				continue
			}

			idx := gy*bounceMap.Width + gx
			cell := &bounceMap.Data[idx]

			if cell.Intensity < 0.01 {
				continue
			}

			// Calculate screen position
			screenX := int((float64(gx)*bounceMap.TileSize - cameraX) * scale)
			screenY := int((float64(gy)*bounceMap.TileSize - cameraY) * scale)

			// Draw colored rectangle for bounce contribution
			col := color.RGBA{
				R: uint8(cell.R * cell.Intensity * 128),
				G: uint8(cell.G * cell.Intensity * 128),
				B: uint8(cell.B * cell.Intensity * 128),
				A: uint8(cell.Intensity * 80), // Subtle alpha
			}

			for py := 0; py < tileScreenSize; py++ {
				for px := 0; px < tileScreenSize; px++ {
					sx := screenX + px
					sy := screenY + py
					if sx >= 0 && sx < s.screenW && sy >= 0 && sy < s.screenH {
						s.overlay.Set(sx, sy, col)
					}
				}
			}
		}
	}

	// Blend overlay additively
	opts := &ebiten.DrawImageOptions{}
	opts.Blend = ebiten.BlendSourceOver
	screen.DrawImage(s.overlay, opts)
}

// adjustSaturation modifies the saturation of an RGB color.
func (s *System) adjustSaturation(r, g, b, factor float64) (nr, ng, nb float64) {
	// Calculate luminance
	lum := 0.299*r + 0.587*g + 0.114*b

	// Interpolate between gray and color
	nr = lum + (r-lum)*factor
	ng = lum + (g-lum)*factor
	nb = lum + (b-lum)*factor

	return clamp01(nr), clamp01(ng), clamp01(nb)
}

// ExtractSurfacesFromColors creates bounce surfaces from a color grid.
// Useful for extracting surfaces from rendered wall/floor textures.
func (s *System) ExtractSurfacesFromColors(colors []color.RGBA, gridW, gridH int, tileSize float64, isWall bool) []BounceSurface {
	surfaces := make([]BounceSurface, 0, len(colors))

	for gy := 0; gy < gridH; gy++ {
		for gx := 0; gx < gridW; gx++ {
			idx := gy*gridW + gx
			if idx >= len(colors) {
				continue
			}

			col := colors[idx]
			if col.A == 0 {
				continue
			}

			surf := BounceSurface{
				X:            (float64(gx) + 0.5) * tileSize,
				Y:            (float64(gy) + 0.5) * tileSize,
				R:            float64(col.R) / 255.0,
				G:            float64(col.G) / 255.0,
				B:            float64(col.B) / 255.0,
				Reflectivity: 0.5,
				IsWall:       isWall,
				DirectLight:  1.0,
			}

			// Adjust reflectivity based on color brightness
			brightness := (surf.R + surf.G + surf.B) / 3.0
			surf.Reflectivity = 0.3 + brightness*0.4

			surfaces = append(surfaces, surf)
		}
	}

	return surfaces
}

// ClearCache clears the bounce map cache.
func (s *System) ClearCache() {
	s.bounceCache = make(map[bounceCacheKey]*BounceMap)
}

// Helper functions

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

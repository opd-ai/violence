// Package groundshadow provides entity ground-contact shadow rendering.
package groundshadow

import (
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines shadow appearance parameters for each genre.
type GenrePreset struct {
	BaseOpacity   float64    // Base shadow darkness [0.0-1.0]
	Softness      float64    // Penumbra gradient softness [0.0-1.0]
	ColorTint     color.RGBA // Shadow color tint (usually dark purple/blue)
	HeightScale   float64    // How much height affects shadow size
	MaxElongation float64    // Maximum elongation from light offset
}

// System renders ground-contact shadows beneath entities.
type System struct {
	mu            sync.RWMutex
	genre         string
	preset        GenrePreset
	pixelsPerUnit float64       // Screen pixels per world unit
	shadowCache   *shadowCache  // LRU cache for shadow images
	lightDirX     float64       // Normalized light direction X (for shadow offset)
	lightDirY     float64       // Normalized light direction Y
	lightStrength float64       // Light strength affects shadow offset magnitude
	logger        *logrus.Entry // Structured logger
}

// shadowCache provides LRU caching for precomputed shadow images.
type shadowCache struct {
	mu       sync.RWMutex
	cache    map[shadowKey]*ebiten.Image
	order    []shadowKey
	maxSize  int
	disposed []shadowKey // Track disposed keys for cleanup
}

// shadowKey uniquely identifies a shadow configuration for caching.
type shadowKey struct {
	radiusPx   int   // Shadow radius in pixels (quantized)
	softness   int   // Softness level (quantized to 10 steps)
	elongation int   // Elongation level (quantized to 10 steps)
	opacity    int   // Opacity level (quantized to 20 steps)
	tintR      uint8 // Color tint R
	tintG      uint8 // Color tint G
	tintB      uint8 // Color tint B
}

// NewSystem creates a ground shadow rendering system.
func NewSystem(genreID string) *System {
	s := &System{
		genre:         genreID,
		pixelsPerUnit: 32.0,
		shadowCache:   newShadowCache(64), // Cache up to 64 shadow variations
		lightDirX:     0.3,                // Default slight offset right/down
		lightDirY:     0.5,
		lightStrength: 1.0,
		logger: logrus.WithFields(logrus.Fields{
			"system": "groundshadow",
		}),
	}
	s.applyGenrePreset(genreID)
	return s
}

// newShadowCache creates an LRU cache for shadow images.
func newShadowCache(maxSize int) *shadowCache {
	return &shadowCache{
		cache:    make(map[shadowKey]*ebiten.Image),
		order:    make([]shadowKey, 0, maxSize),
		maxSize:  maxSize,
		disposed: make([]shadowKey, 0),
	}
}

// applyGenrePreset configures shadow parameters based on genre.
func (s *System) applyGenrePreset(genreID string) {
	switch genreID {
	case "fantasy":
		// Warm, medium-soft shadows for torchlit environments
		s.preset = GenrePreset{
			BaseOpacity:   0.55,
			Softness:      0.6,
			ColorTint:     color.RGBA{R: 20, G: 15, B: 30, A: 255},
			HeightScale:   1.2,
			MaxElongation: 0.4,
		}
	case "scifi":
		// Crisp, cool shadows for artificial lighting
		s.preset = GenrePreset{
			BaseOpacity:   0.5,
			Softness:      0.35,
			ColorTint:     color.RGBA{R: 10, G: 20, B: 35, A: 255},
			HeightScale:   1.0,
			MaxElongation: 0.3,
		}
	case "horror":
		// Deep, very soft shadows for atmospheric dread
		s.preset = GenrePreset{
			BaseOpacity:   0.7,
			Softness:      0.8,
			ColorTint:     color.RGBA{R: 15, G: 10, B: 20, A: 255},
			HeightScale:   1.5,
			MaxElongation: 0.6,
		}
	case "cyberpunk":
		// Hard-edged, high-contrast shadows for neon environments
		s.preset = GenrePreset{
			BaseOpacity:   0.6,
			Softness:      0.25,
			ColorTint:     color.RGBA{R: 30, G: 10, B: 40, A: 255},
			HeightScale:   0.9,
			MaxElongation: 0.35,
		}
	case "postapoc":
		// Medium shadows with dusty, warm tint
		s.preset = GenrePreset{
			BaseOpacity:   0.5,
			Softness:      0.55,
			ColorTint:     color.RGBA{R: 35, G: 25, B: 20, A: 255},
			HeightScale:   1.1,
			MaxElongation: 0.45,
		}
	default:
		// Default to fantasy preset
		s.preset = GenrePreset{
			BaseOpacity:   0.55,
			Softness:      0.6,
			ColorTint:     color.RGBA{R: 20, G: 15, B: 30, A: 255},
			HeightScale:   1.2,
			MaxElongation: 0.4,
		}
		s.logger.Warnf("unknown genre %q, using fantasy defaults", genreID)
	}
	s.logger.WithField("genre", genreID).Debug("applied genre preset")
}

// SetGenre updates shadow parameters for a new genre.
func (s *System) SetGenre(genreID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.genre == genreID {
		return
	}
	s.genre = genreID
	s.applyGenrePreset(genreID)
	// Clear cache when genre changes (shadow appearance differs)
	s.shadowCache.clear()
}

// SetLightDirection sets the dominant light direction for shadow offset.
// dx, dy should be normalized. Strength affects offset magnitude.
func (s *System) SetLightDirection(dx, dy, strength float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lightDirX = dx
	s.lightDirY = dy
	s.lightStrength = strength
}

// SetPixelsPerUnit configures the world-to-screen scale factor.
func (s *System) SetPixelsPerUnit(ppu float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ppu > 0 {
		s.pixelsPerUnit = ppu
	}
}

// RenderShadow draws a ground shadow for an entity at the given position.
// worldX, worldY: entity position in world coordinates
// cameraX, cameraY: camera position in world coordinates
// comp: the entity's ground shadow component
// screen: the target ebiten.Image to draw onto
func (s *System) RenderShadow(
	screen *ebiten.Image,
	worldX, worldY float64,
	cameraX, cameraY float64,
	comp *Component,
) {
	if comp == nil || !comp.CastShadow {
		return
	}

	s.mu.RLock()
	preset := s.preset
	ppu := s.pixelsPerUnit
	lightDirX := s.lightDirX
	lightDirY := s.lightDirY
	lightStrength := s.lightStrength
	s.mu.RUnlock()

	// Calculate shadow parameters
	shadowRadius := comp.Radius * (1.0 + comp.Height*preset.HeightScale*0.3)
	shadowRadiusPx := int(shadowRadius * ppu)
	if shadowRadiusPx < 2 {
		shadowRadiusPx = 2
	}

	// Calculate elongation from height and light direction
	elongation := comp.Elongation + comp.Height*0.1*preset.MaxElongation
	if elongation > preset.MaxElongation {
		elongation = preset.MaxElongation
	}

	// Calculate shadow offset from light direction
	offsetMag := comp.Height * lightStrength * 0.3
	offsetX := comp.OffsetX + lightDirX*offsetMag
	offsetY := comp.OffsetY + lightDirY*offsetMag

	// Calculate opacity (closer to ground = darker shadow)
	opacity := comp.Opacity * preset.BaseOpacity
	heightFade := 1.0 - comp.Height*0.15 // Higher entities have slightly lighter shadows
	if heightFade < 0.5 {
		heightFade = 0.5
	}
	opacity *= heightFade

	// Get or create cached shadow image
	key := shadowKey{
		radiusPx:   shadowRadiusPx,
		softness:   int(preset.Softness * 10),
		elongation: int(elongation * 10),
		opacity:    int(opacity * 20),
		tintR:      preset.ColorTint.R,
		tintG:      preset.ColorTint.G,
		tintB:      preset.ColorTint.B,
	}

	shadowImg := s.shadowCache.get(key)
	if shadowImg == nil {
		shadowImg = s.generateShadowImage(shadowRadiusPx, preset.Softness, elongation, opacity, preset.ColorTint)
		s.shadowCache.put(key, shadowImg)
	}

	// Convert world position to screen position
	screenX := (worldX - cameraX + offsetX) * ppu
	screenY := (worldY - cameraY + offsetY) * ppu

	// Draw shadow centered at entity feet
	imgBounds := shadowImg.Bounds()
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(
		screenX-float64(imgBounds.Dx())/2,
		screenY-float64(imgBounds.Dy())/2,
	)

	screen.DrawImage(shadowImg, opts)
}

// RenderShadows renders ground shadows for multiple entities.
// This is the batch version for efficiency.
func (s *System) RenderShadows(
	screen *ebiten.Image,
	positions [][2]float64, // [i][0]=worldX, [i][1]=worldY
	cameraX, cameraY float64,
	components []*Component,
) {
	if len(positions) != len(components) {
		s.logger.Error("positions and components length mismatch")
		return
	}

	for i, pos := range positions {
		s.RenderShadow(screen, pos[0], pos[1], cameraX, cameraY, components[i])
	}
}

// generateShadowImage creates a soft elliptical shadow image.
func (s *System) generateShadowImage(radiusPx int, softness, elongation, opacity float64, tint color.RGBA) *ebiten.Image {
	// Calculate image dimensions (add margin for soft edge)
	margin := int(float64(radiusPx) * softness * 0.5)
	if margin < 2 {
		margin = 2
	}

	// Elongation affects Y radius
	radiusX := radiusPx
	radiusY := int(float64(radiusPx) * (1.0 - elongation*0.5))
	if radiusY < 2 {
		radiusY = 2
	}

	width := (radiusX + margin) * 2
	height := (radiusY + margin) * 2

	// Prevent extremely large images
	if width > 256 {
		width = 256
	}
	if height > 256 {
		height = 256
	}

	pixels := make([]byte, width*height*4)
	centerX := float64(width) / 2
	centerY := float64(height) / 2
	radX := float64(radiusX)
	radY := float64(radiusY)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Normalized distance from ellipse center
			dx := (float64(x) - centerX) / radX
			dy := (float64(y) - centerY) / radY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Calculate alpha with gradient falloff
			alpha := 0.0
			coreRadius := 1.0 - softness*0.8
			if dist < coreRadius {
				// Core region: full opacity
				alpha = opacity
			} else if dist < 1.0 {
				// Transition region: linear gradient
				t := (dist - coreRadius) / (1.0 - coreRadius)
				// Quadratic falloff for smooth penumbra
				alpha = opacity * (1.0 - t*t)
			}
			// Outside dist >= 1.0: alpha stays 0

			// Apply tint color with calculated alpha
			idx := (y*width + x) * 4
			pixels[idx] = tint.R
			pixels[idx+1] = tint.G
			pixels[idx+2] = tint.B
			pixels[idx+3] = uint8(alpha * 255)
		}
	}

	img := ebiten.NewImage(width, height)
	img.WritePixels(pixels)
	return img
}

// Update is a no-op for the ECS system interface.
// Shadow rendering is triggered directly during the render pass.
func (s *System) Update() {
	// Ground shadows are rendered on demand, not updated per tick
}

// GetShadowImageForEntity returns the cached or generated shadow image for debugging.
func (s *System) GetShadowImageForEntity(comp *Component) image.Image {
	if comp == nil {
		return nil
	}

	s.mu.RLock()
	preset := s.preset
	ppu := s.pixelsPerUnit
	s.mu.RUnlock()

	shadowRadius := comp.Radius * (1.0 + comp.Height*preset.HeightScale*0.3)
	shadowRadiusPx := int(shadowRadius * ppu)
	if shadowRadiusPx < 2 {
		shadowRadiusPx = 2
	}

	elongation := comp.Elongation + comp.Height*0.1*preset.MaxElongation
	if elongation > preset.MaxElongation {
		elongation = preset.MaxElongation
	}

	opacity := comp.Opacity * preset.BaseOpacity
	heightFade := 1.0 - comp.Height*0.15
	if heightFade < 0.5 {
		heightFade = 0.5
	}
	opacity *= heightFade

	return s.generateShadowImage(shadowRadiusPx, preset.Softness, elongation, opacity, preset.ColorTint)
}

// shadow cache methods

func (c *shadowCache) get(key shadowKey) *ebiten.Image {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cache[key]
}

func (c *shadowCache) put(key shadowKey, img *ebiten.Image) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if already exists
	if _, exists := c.cache[key]; exists {
		return
	}

	// Evict oldest if at capacity
	if len(c.cache) >= c.maxSize {
		oldKey := c.order[0]
		c.order = c.order[1:]
		if oldImg, ok := c.cache[oldKey]; ok {
			oldImg.Dispose()
			delete(c.cache, oldKey)
		}
	}

	c.cache[key] = img
	c.order = append(c.order, key)
}

func (c *shadowCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, img := range c.cache {
		img.Dispose()
	}
	c.cache = make(map[shadowKey]*ebiten.Image)
	c.order = c.order[:0]
}

// GetGenre returns the current genre ID.
func (s *System) GetGenre() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.genre
}

// GetPreset returns the current genre preset (for testing).
func (s *System) GetPreset() GenrePreset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.preset
}

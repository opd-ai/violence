package damagedir

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// GenrePreset defines damage indicator appearance for each genre.
type GenrePreset struct {
	// BaseColor is the indicator color (usually a shade of red)
	BaseColor color.RGBA
	// ArcWidth is the angular width of the damage arc in radians
	ArcWidth float64
	// EdgeDepth is how far the vignette extends from screen edge
	EdgeDepth float64
	// FadeSpeed multiplier for how quickly indicators fade
	FadeSpeed float64
	// MaxIntensity caps the brightness of indicators
	MaxIntensity float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BaseColor:    color.RGBA{R: 200, G: 30, B: 30, A: 255},
		ArcWidth:     math.Pi / 3, // 60 degrees
		EdgeDepth:    80,
		FadeSpeed:    1.0,
		MaxIntensity: 0.8,
	},
	"scifi": {
		BaseColor:    color.RGBA{R: 255, G: 100, B: 80, A: 255},
		ArcWidth:     math.Pi / 4, // 45 degrees
		EdgeDepth:    60,
		FadeSpeed:    1.2,
		MaxIntensity: 0.7,
	},
	"horror": {
		BaseColor:    color.RGBA{R: 150, G: 0, B: 0, A: 255},
		ArcWidth:     math.Pi / 2.5, // 72 degrees
		EdgeDepth:    100,
		FadeSpeed:    0.8,
		MaxIntensity: 0.9,
	},
	"cyberpunk": {
		BaseColor:    color.RGBA{R: 255, G: 50, B: 100, A: 255},
		ArcWidth:     math.Pi / 4,
		EdgeDepth:    50,
		FadeSpeed:    1.5,
		MaxIntensity: 0.75,
	},
	"postapoc": {
		BaseColor:    color.RGBA{R: 180, G: 60, B: 40, A: 255},
		ArcWidth:     math.Pi / 3,
		EdgeDepth:    70,
		FadeSpeed:    0.9,
		MaxIntensity: 0.85,
	},
}

// System manages directional damage indicators.
type System struct {
	genreID       string
	preset        GenrePreset
	logger        *logrus.Entry
	indicators    []*Component
	maxIndicators int
	screenW       int
	screenH       int
	arcSprite     *ebiten.Image
}

// NewSystem creates a new directional damage indicator system.
func NewSystem(genreID string) *System {
	preset, ok := genrePresets[genreID]
	if !ok {
		preset = genrePresets["fantasy"]
	}

	s := &System{
		genreID:       genreID,
		preset:        preset,
		logger:        logrus.WithFields(logrus.Fields{"system_name": "damagedir", "genre": genreID}),
		indicators:    make([]*Component, 0, 8),
		maxIndicators: 8,
		screenW:       320,
		screenH:       200,
	}

	s.generateArcSprite()
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
	s.logger = s.logger.WithField("genre", genreID)
	s.generateArcSprite()
}

// SetScreenSize updates screen dimensions for rendering calculations.
func (s *System) SetScreenSize(w, h int) {
	if w != s.screenW || h != s.screenH {
		s.screenW = w
		s.screenH = h
		s.generateArcSprite()
	}
}

// generateArcSprite creates a cached radial gradient arc for efficient rendering.
func (s *System) generateArcSprite() {
	// Create a quarter-screen sized gradient for edge vignettes
	size := int(s.preset.EdgeDepth * 2)
	if size < 32 {
		size = 32
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	centerX := float64(size) / 2
	centerY := float64(size) / 2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)
			maxDist := float64(size) / 2

			// Radial falloff from edge (inverse - brighter at edge, fading inward)
			if dist > maxDist {
				continue
			}

			// Falloff from edge toward center
			edgeFactor := dist / maxDist
			// Use quadratic falloff for softer gradient
			brightness := edgeFactor * edgeFactor

			if brightness < 0.01 {
				continue
			}

			alpha := uint8(brightness * 255)
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: alpha})
		}
	}

	s.arcSprite = ebiten.NewImageFromImage(img)
}

// TriggerDamage adds a new damage direction indicator.
// sourceX, sourceY: world position of damage source
// playerX, playerY: world position of player
// damage: amount of damage (scales intensity)
// playerAngle: player's facing direction in radians
func (s *System) TriggerDamage(sourceX, sourceY, playerX, playerY, damage, playerAngle float64) {
	// Calculate direction from player to source in world space
	dx := sourceX - playerX
	dy := sourceY - playerY

	// Skip if source is at player position
	if math.Abs(dx) < 0.001 && math.Abs(dy) < 0.001 {
		return
	}

	// Get world-space angle to damage source
	worldAngle := math.Atan2(dy, dx)

	// Convert to screen-space by subtracting player's facing angle
	// This makes the indicator appear relative to where the player is looking
	screenAngle := worldAngle - playerAngle

	// Normalize to -PI to PI
	for screenAngle > math.Pi {
		screenAngle -= 2 * math.Pi
	}
	for screenAngle < -math.Pi {
		screenAngle += 2 * math.Pi
	}

	// Calculate intensity based on damage (scale 0-100 damage to 0.3-1.0 intensity)
	intensity := 0.3 + (math.Min(damage, 100.0)/100.0)*0.7
	if intensity > s.preset.MaxIntensity {
		intensity = s.preset.MaxIntensity
	}

	// Create indicator component
	indicator := &Component{
		Direction:   screenAngle,
		Intensity:   intensity,
		Lifetime:    1.5 / s.preset.FadeSpeed,
		MaxLifetime: 1.5 / s.preset.FadeSpeed,
		Color:       s.preset.BaseColor,
		ArcWidth:    s.preset.ArcWidth,
		EdgeDepth:   s.preset.EdgeDepth,
	}

	// Add to list, removing oldest if at capacity
	if len(s.indicators) >= s.maxIndicators {
		s.indicators = s.indicators[1:]
	}
	s.indicators = append(s.indicators, indicator)

	s.logger.WithFields(logrus.Fields{
		"direction": screenAngle,
		"intensity": intensity,
		"damage":    damage,
	}).Debug("Triggered damage direction indicator")
}

// Update processes all indicators, handling fade-out and removal.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime

	active := make([]*Component, 0, len(s.indicators))
	for _, ind := range s.indicators {
		ind.Lifetime -= deltaTime
		if !ind.IsExpired() {
			active = append(active, ind)
		}
	}
	s.indicators = active
}

// Render draws all active damage direction indicators to the screen.
func (s *System) Render(screen *ebiten.Image) {
	if len(s.indicators) == 0 || s.arcSprite == nil {
		return
	}

	screenCenterX := float64(s.screenW) / 2
	screenCenterY := float64(s.screenH) / 2

	for _, ind := range s.indicators {
		s.renderIndicator(screen, ind, screenCenterX, screenCenterY)
	}
}

// renderIndicator draws a single damage direction arc on the screen edge.
func (s *System) renderIndicator(screen *ebiten.Image, ind *Component, centerX, centerY float64) {
	alpha := ind.GetAlpha()
	if alpha < 0.01 {
		return
	}

	// Determine which screen edge(s) to render on based on direction
	// Direction: 0 = right, PI/2 = bottom, PI = left, -PI/2 = top
	// (note: screen Y increases downward)

	// Calculate position on screen edge
	dir := ind.Direction

	// Map direction to screen edge position
	// We render the indicator on the edge that the damage came FROM
	var edgeX, edgeY float64
	var rotation float64

	// Determine primary edge based on angle
	absDirX := math.Abs(math.Cos(dir))
	absDirY := math.Abs(math.Sin(dir))

	if absDirX > absDirY {
		// Left or right edge
		if math.Cos(dir) > 0 {
			// Right edge - damage from right
			edgeX = float64(s.screenW) - ind.EdgeDepth/2
			edgeY = centerY + math.Sin(dir)*centerY*0.8
			rotation = 0
		} else {
			// Left edge - damage from left
			edgeX = ind.EdgeDepth / 2
			edgeY = centerY + math.Sin(dir)*centerY*0.8
			rotation = math.Pi
		}
	} else {
		// Top or bottom edge
		if math.Sin(dir) > 0 {
			// Bottom edge - damage from below
			edgeX = centerX + math.Cos(dir)*centerX*0.8
			edgeY = float64(s.screenH) - ind.EdgeDepth/2
			rotation = math.Pi / 2
		} else {
			// Top edge - damage from above
			edgeX = centerX + math.Cos(dir)*centerX*0.8
			edgeY = ind.EdgeDepth / 2
			rotation = -math.Pi / 2
		}
	}

	// Clamp edge positions to valid screen coordinates
	edgeX = clamp(edgeX, ind.EdgeDepth/2, float64(s.screenW)-ind.EdgeDepth/2)
	edgeY = clamp(edgeY, ind.EdgeDepth/2, float64(s.screenH)-ind.EdgeDepth/2)

	// Draw the arc sprite at the edge position
	spriteW := float64(s.arcSprite.Bounds().Dx())
	spriteH := float64(s.arcSprite.Bounds().Dy())

	opts := &ebiten.DrawImageOptions{}

	// Center the sprite
	opts.GeoM.Translate(-spriteW/2, -spriteH/2)

	// Rotate to face inward from edge
	opts.GeoM.Rotate(rotation)

	// Move to edge position
	opts.GeoM.Translate(edgeX, edgeY)

	// Apply color tint and alpha
	r := float32(ind.Color.R) / 255.0
	g := float32(ind.Color.G) / 255.0
	b := float32(ind.Color.B) / 255.0
	opts.ColorScale.Scale(r, g, b, float32(alpha))

	// Use additive blending for glow effect
	opts.Blend = ebiten.BlendLighter

	screen.DrawImage(s.arcSprite, opts)

	// Draw additional arc segments for wider coverage
	arcHalfWidth := ind.ArcWidth / 2
	segments := 3
	for i := 1; i <= segments; i++ {
		offset := arcHalfWidth * float64(i) / float64(segments+1)

		for _, sign := range []float64{-1, 1} {
			segDir := dir + offset*sign
			segAlpha := alpha * (1.0 - float64(i)*0.25) // Fade outer segments

			if segAlpha < 0.05 {
				continue
			}

			var segX, segY float64
			var segRot float64

			segAbsDirX := math.Abs(math.Cos(segDir))
			segAbsDirY := math.Abs(math.Sin(segDir))

			if segAbsDirX > segAbsDirY {
				if math.Cos(segDir) > 0 {
					segX = float64(s.screenW) - ind.EdgeDepth/2
					segY = centerY + math.Sin(segDir)*centerY*0.8
					segRot = 0
				} else {
					segX = ind.EdgeDepth / 2
					segY = centerY + math.Sin(segDir)*centerY*0.8
					segRot = math.Pi
				}
			} else {
				if math.Sin(segDir) > 0 {
					segX = centerX + math.Cos(segDir)*centerX*0.8
					segY = float64(s.screenH) - ind.EdgeDepth/2
					segRot = math.Pi / 2
				} else {
					segX = centerX + math.Cos(segDir)*centerX*0.8
					segY = ind.EdgeDepth / 2
					segRot = -math.Pi / 2
				}
			}

			segX = clamp(segX, ind.EdgeDepth/2, float64(s.screenW)-ind.EdgeDepth/2)
			segY = clamp(segY, ind.EdgeDepth/2, float64(s.screenH)-ind.EdgeDepth/2)

			segOpts := &ebiten.DrawImageOptions{}
			segOpts.GeoM.Translate(-spriteW/2, -spriteH/2)
			segOpts.GeoM.Rotate(segRot)
			segOpts.GeoM.Translate(segX, segY)
			segOpts.ColorScale.Scale(r, g, b, float32(segAlpha))
			segOpts.Blend = ebiten.BlendLighter

			screen.DrawImage(s.arcSprite, segOpts)
		}
	}
}

// GetActiveCount returns the number of active damage indicators.
func (s *System) GetActiveCount() int {
	return len(s.indicators)
}

// Clear removes all active indicators.
func (s *System) Clear() {
	s.indicators = s.indicators[:0]
}

// clamp constrains a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

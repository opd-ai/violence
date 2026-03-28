// Package hitmarker provides visual hit confirmation feedback at screen center.
package hitmarker

import (
	"image"
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages hit marker animation and rendering.
type System struct {
	genreID string
	logger  *logrus.Entry

	// Genre-specific color schemes
	normalColor    color.RGBA
	criticalColor  color.RGBA
	killColor      color.RGBA
	headshotColor  color.RGBA
	weakpointColor color.RGBA

	// Cached marker images by type
	markerCache map[HitType]*ebiten.Image

	// Screen dimensions for centering
	screenWidth  int
	screenHeight int
}

// NewSystem creates a hit marker system with genre-specific theming.
func NewSystem(genreID string) *System {
	s := &System{
		genreID:      genreID,
		markerCache:  make(map[HitType]*ebiten.Image),
		screenWidth:  320,
		screenHeight: 200,
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "hitmarker",
			"genre":       genreID,
		}),
	}
	s.applyGenreTheme()
	return s
}

// SetScreenSize updates screen dimensions for marker positioning.
func (s *System) SetScreenSize(width, height int) {
	s.screenWidth = width
	s.screenHeight = height
}

// SetGenre updates the genre and applies new theming.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenreTheme()
	// Clear cache to regenerate with new colors
	s.markerCache = make(map[HitType]*ebiten.Image)
}

// applyGenreTheme sets genre-specific colors for hit markers.
func (s *System) applyGenreTheme() {
	switch s.genreID {
	case "cyberpunk":
		s.normalColor = color.RGBA{0, 255, 255, 255}    // Cyan
		s.criticalColor = color.RGBA{255, 0, 255, 255}  // Magenta
		s.killColor = color.RGBA{255, 50, 100, 255}     // Hot pink
		s.headshotColor = color.RGBA{255, 255, 0, 255}  // Yellow
		s.weakpointColor = color.RGBA{100, 255, 0, 255} // Lime
	case "horror":
		s.normalColor = color.RGBA{200, 50, 50, 255}    // Dark red
		s.criticalColor = color.RGBA{255, 100, 0, 255}  // Orange
		s.killColor = color.RGBA{180, 0, 0, 255}        // Blood red
		s.headshotColor = color.RGBA{255, 200, 0, 255}  // Amber
		s.weakpointColor = color.RGBA{200, 150, 0, 255} // Gold
	case "scifi":
		s.normalColor = color.RGBA{100, 200, 255, 255}    // Light blue
		s.criticalColor = color.RGBA{255, 150, 0, 255}    // Orange
		s.killColor = color.RGBA{255, 50, 50, 255}        // Red
		s.headshotColor = color.RGBA{0, 255, 200, 255}    // Teal
		s.weakpointColor = color.RGBA{200, 100, 255, 255} // Purple
	case "postapoc":
		s.normalColor = color.RGBA{200, 180, 100, 255}    // Tan
		s.criticalColor = color.RGBA{255, 150, 50, 255}   // Orange-brown
		s.killColor = color.RGBA{200, 50, 50, 255}        // Rust red
		s.headshotColor = color.RGBA{255, 200, 100, 255}  // Sand yellow
		s.weakpointColor = color.RGBA{150, 200, 100, 255} // Olive
	default: // fantasy
		s.normalColor = color.RGBA{255, 255, 255, 255}    // White
		s.criticalColor = color.RGBA{255, 200, 50, 255}   // Gold
		s.killColor = color.RGBA{255, 50, 50, 255}        // Red
		s.headshotColor = color.RGBA{255, 255, 100, 255}  // Bright yellow
		s.weakpointColor = color.RGBA{100, 255, 100, 255} // Green
	}
}

// getColorForHitType returns the appropriate color for a hit type.
func (s *System) getColorForHitType(hitType HitType) color.RGBA {
	switch hitType {
	case HitCritical:
		return s.criticalColor
	case HitKill:
		return s.killColor
	case HitHeadshot:
		return s.headshotColor
	case HitWeakpoint:
		return s.weakpointColor
	default:
		return s.normalColor
	}
}

// Update animates hit markers (scale pop, fade out).
func (s *System) Update(w *engine.World) {
	compType := reflect.TypeOf((*Component)(nil))
	entities := w.Query(compType)

	const deltaTime = 1.0 / 60.0

	for _, ent := range entities {
		comp, found := w.GetComponent(ent, compType)
		if !found {
			continue
		}

		hm, ok := comp.(*Component)
		if !ok || !hm.Active {
			continue
		}

		hm.Age += deltaTime

		if hm.Age >= hm.Duration {
			hm.Reset()
			continue
		}

		progress := hm.Age / hm.Duration

		// Pop-in animation: quick scale up then settle
		if progress < 0.2 {
			// Overshoot scale for impact
			hm.Scale = 0.5 + (progress/0.2)*0.8 // 0.5 -> 1.3
		} else if progress < 0.4 {
			// Settle back
			settleProgress := (progress - 0.2) / 0.2
			hm.Scale = 1.3 - settleProgress*0.3 // 1.3 -> 1.0
		} else {
			hm.Scale = 1.0
		}

		// Scale modifier for intensity
		hm.Scale *= (0.7 + 0.3*hm.Intensity)

		// Fade out in last 40%
		if progress > 0.6 {
			fadeProgress := (progress - 0.6) / 0.4
			hm.Alpha = 1.0 - fadeProgress
		} else {
			hm.Alpha = 1.0
		}

		// Slight rotation for kills/crits
		if hm.HitType == HitKill || hm.HitType == HitCritical {
			hm.Rotation = math.Sin(hm.Age*20.0) * 0.1 * (1.0 - progress)
		}

		// Update color with proper alpha
		baseColor := s.getColorForHitType(hm.HitType)
		hm.Color = baseColor
		hm.Color.A = uint8(hm.Alpha * 255)
	}
}

// Render draws active hit markers.
func (s *System) Render(w *engine.World, screen *ebiten.Image) {
	compType := reflect.TypeOf((*Component)(nil))
	entities := w.Query(compType)

	for _, ent := range entities {
		comp, found := w.GetComponent(ent, compType)
		if !found {
			continue
		}

		hm, ok := comp.(*Component)
		if !ok || !hm.Active || hm.Alpha <= 0 {
			continue
		}

		s.renderHitMarker(screen, hm)
	}
}

// renderHitMarker draws a single hit marker.
func (s *System) renderHitMarker(screen *ebiten.Image, hm *Component) {
	// Use screen center if position not specified
	centerX := hm.ScreenX
	centerY := hm.ScreenY
	if centerX == 0 && centerY == 0 {
		centerX = float64(s.screenWidth) / 2
		centerY = float64(s.screenHeight) / 2
	}

	// Get or generate marker image
	markerImg := s.getMarkerImage(hm.HitType)
	if markerImg == nil {
		return
	}

	// Draw with transformation
	opts := &ebiten.DrawImageOptions{}

	// Center the image
	bounds := markerImg.Bounds()
	halfW := float64(bounds.Dx()) / 2
	halfH := float64(bounds.Dy()) / 2

	opts.GeoM.Translate(-halfW, -halfH)

	// Apply rotation
	if hm.Rotation != 0 {
		opts.GeoM.Rotate(hm.Rotation)
	}

	// Apply scale
	opts.GeoM.Scale(hm.Scale, hm.Scale)

	// Move to position
	opts.GeoM.Translate(centerX, centerY)

	// Apply alpha and color tint
	r := float64(hm.Color.R) / 255.0
	g := float64(hm.Color.G) / 255.0
	b := float64(hm.Color.B) / 255.0
	opts.ColorScale.Scale(float32(r), float32(g), float32(b), float32(hm.Alpha))

	screen.DrawImage(markerImg, opts)
}

// getMarkerImage returns a cached or newly generated marker image.
func (s *System) getMarkerImage(hitType HitType) *ebiten.Image {
	if img, exists := s.markerCache[hitType]; exists {
		return img
	}

	img := s.generateMarkerImage(hitType)
	s.markerCache[hitType] = img
	return img
}

// generateMarkerImage creates the hit marker graphic.
func (s *System) generateMarkerImage(hitType HitType) *ebiten.Image {
	// Size varies by hit type
	size := 24
	if hitType == HitKill {
		size = 32
	} else if hitType == HitCritical || hitType == HitHeadshot {
		size = 28
	}

	img := ebiten.NewImage(size, size)
	centerF := float32(size) / 2

	// Draw on a temp RGBA for vector operations
	baseColor := s.getColorForHitType(hitType)

	switch hitType {
	case HitKill:
		// X mark for kills
		s.drawKillMarker(img, centerF, baseColor)
	case HitCritical:
		// Star burst for crits
		s.drawCritMarker(img, centerF, baseColor)
	case HitHeadshot:
		// Crosshair with dot for headshots
		s.drawHeadshotMarker(img, centerF, baseColor)
	case HitWeakpoint:
		// Diamond shape for weakpoint
		s.drawWeakpointMarker(img, centerF, baseColor)
	default:
		// Standard crosshair for normal hits
		s.drawNormalMarker(img, centerF, baseColor)
	}

	return img
}

// drawNormalMarker draws a simple crosshair hit marker.
func (s *System) drawNormalMarker(img *ebiten.Image, center float32, clr color.RGBA) {
	thickness := float32(2)
	armLength := center * 0.6
	gap := center * 0.2

	// Horizontal arms
	vector.DrawFilledRect(img, center-armLength, center-thickness/2, armLength-gap, thickness, clr, false)
	vector.DrawFilledRect(img, center+gap, center-thickness/2, armLength-gap, thickness, clr, false)

	// Vertical arms
	vector.DrawFilledRect(img, center-thickness/2, center-armLength, thickness, armLength-gap, clr, false)
	vector.DrawFilledRect(img, center-thickness/2, center+gap, thickness, armLength-gap, clr, false)

	// Add subtle shading - darker outlines
	outlineClr := color.RGBA{clr.R / 2, clr.G / 2, clr.B / 2, clr.A}
	outlineThick := float32(1)
	vector.StrokeRect(img, center-armLength-outlineThick, center-thickness/2-outlineThick,
		armLength-gap+outlineThick*2, thickness+outlineThick*2, outlineThick, outlineClr, false)
}

// drawKillMarker draws an X-shaped marker for kills.
func (s *System) drawKillMarker(img *ebiten.Image, center float32, clr color.RGBA) {
	armLength := center * 0.7
	thickness := float32(3)

	// Create a small image to compose the X
	rgba := image.NewRGBA(image.Rect(0, 0, int(center*2), int(center*2)))

	// Draw X diagonals manually
	for i := float32(-armLength); i <= armLength; i += 0.5 {
		// Main diagonal (top-left to bottom-right)
		x1 := int(center + i)
		y1 := int(center + i)
		for t := float32(-thickness / 2); t <= thickness/2; t++ {
			if x1 >= 0 && x1 < int(center*2) && int(float32(y1)+t) >= 0 && int(float32(y1)+t) < int(center*2) {
				rgba.Set(x1, int(float32(y1)+t), clr)
			}
		}

		// Anti-diagonal (top-right to bottom-left)
		x2 := int(center + i)
		y2 := int(center - i)
		for t := float32(-thickness / 2); t <= thickness/2; t++ {
			if x2 >= 0 && x2 < int(center*2) && int(float32(y2)+t) >= 0 && int(float32(y2)+t) < int(center*2) {
				rgba.Set(x2, int(float32(y2)+t), clr)
			}
		}
	}

	img.WritePixels(rgba.Pix)
}

// drawCritMarker draws a star-burst marker for critical hits.
func (s *System) drawCritMarker(img *ebiten.Image, center float32, clr color.RGBA) {
	// Draw 8-pointed star using lines
	outerRadius := center * 0.8
	innerRadius := center * 0.4
	points := 8

	for i := 0; i < points; i++ {
		angle := float64(i) * math.Pi / float64(points/2)
		nextAngle := float64(i+1) * math.Pi / float64(points/2)

		// Outer point
		ox := center + float32(math.Cos(angle)*float64(outerRadius))
		oy := center + float32(math.Sin(angle)*float64(outerRadius))

		// Inner point (between outer points)
		midAngle := angle + math.Pi/float64(points)
		ix := center + float32(math.Cos(midAngle)*float64(innerRadius))
		iy := center + float32(math.Sin(midAngle)*float64(innerRadius))

		// Next outer point
		nox := center + float32(math.Cos(nextAngle)*float64(outerRadius))
		noy := center + float32(math.Sin(nextAngle)*float64(outerRadius))

		// Draw lines
		vector.StrokeLine(img, ox, oy, ix, iy, 2, clr, false)
		vector.StrokeLine(img, ix, iy, nox, noy, 2, clr, false)
	}

	// Center dot
	vector.DrawFilledCircle(img, center, center, 3, clr, false)
}

// drawHeadshotMarker draws a precision crosshair for headshots.
func (s *System) drawHeadshotMarker(img *ebiten.Image, center float32, clr color.RGBA) {
	// Thin precision crosshair with center dot
	thickness := float32(1.5)
	armLength := center * 0.7
	gap := center * 0.15

	// Arms
	vector.DrawFilledRect(img, center-armLength, center-thickness/2, armLength-gap, thickness, clr, false)
	vector.DrawFilledRect(img, center+gap, center-thickness/2, armLength-gap, thickness, clr, false)
	vector.DrawFilledRect(img, center-thickness/2, center-armLength, thickness, armLength-gap, clr, false)
	vector.DrawFilledRect(img, center-thickness/2, center+gap, thickness, armLength-gap, clr, false)

	// Center dot (larger for headshot)
	vector.DrawFilledCircle(img, center, center, 3, clr, false)

	// Outer ring
	vector.StrokeCircle(img, center, center, center*0.5, 1, clr, false)
}

// drawWeakpointMarker draws a diamond marker for weakpoint hits.
func (s *System) drawWeakpointMarker(img *ebiten.Image, center float32, clr color.RGBA) {
	size := center * 0.6

	// Diamond vertices
	top := []float32{center, center - size}
	right := []float32{center + size, center}
	bottom := []float32{center, center + size}
	left := []float32{center - size, center}

	// Draw diamond outline
	vector.StrokeLine(img, top[0], top[1], right[0], right[1], 2, clr, false)
	vector.StrokeLine(img, right[0], right[1], bottom[0], bottom[1], 2, clr, false)
	vector.StrokeLine(img, bottom[0], bottom[1], left[0], left[1], 2, clr, false)
	vector.StrokeLine(img, left[0], left[1], top[0], top[1], 2, clr, false)

	// Center dot
	vector.DrawFilledCircle(img, center, center, 2, clr, false)
}

// TriggerHit is a convenience function to trigger a hit marker on an entity.
func TriggerHit(w *engine.World, ent engine.Entity, hitType HitType, damageValue int, screenX, screenY float64) {
	compType := reflect.TypeOf((*Component)(nil))
	comp, found := w.GetComponent(ent, compType)
	if !found {
		return
	}

	hm, ok := comp.(*Component)
	if !ok {
		return
	}

	hm.Trigger(hitType, damageValue, screenX, screenY)
}

// SpawnHitMarker creates a hit marker entity with component.
func SpawnHitMarker(w *engine.World) engine.Entity {
	ent := w.AddEntity()
	w.AddComponent(ent, NewComponent())
	return ent
}

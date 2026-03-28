package reloadbar

import (
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// GenreStyle defines genre-specific visual parameters for the reload bar.
type GenreStyle struct {
	BackgroundColor color.RGBA
	FillColor       color.RGBA
	BorderColor     color.RGBA
	GlowColor       color.RGBA
	TextColor       color.RGBA
	BarWidth        float32
	BarHeight       float32
	BorderWidth     float32
	CornerRadius    float32
	GlowIntensity   float32
	PulseSpeed      float64
	YOffset         float32 // Offset below crosshair center
}

// System manages reload bar rendering and updates.
type System struct {
	genreID string
	style   GenreStyle
	logger  *logrus.Entry

	// Current state (for non-ECS usage)
	isReloading   bool
	progress      float64
	fadeAlpha     float64
	totalDuration float64
	elapsedTime   float64
	pulsePhase    float64

	// Screen dimensions
	screenWidth  int
	screenHeight int
}

// NewSystem creates a reload bar system with genre-specific styling.
func NewSystem(genreID string) *System {
	s := &System{
		genreID:      genreID,
		screenWidth:  320,
		screenHeight: 200,
		logger: logrus.WithFields(logrus.Fields{
			"system":  "reloadbar",
			"package": "reloadbar",
		}),
	}
	s.applyGenreStyle()
	return s
}

// SetGenre updates the visual style for a different genre.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenreStyle()
}

// applyGenreStyle configures visual parameters based on genre.
func (s *System) applyGenreStyle() {
	switch s.genreID {
	case "cyberpunk":
		s.style = GenreStyle{
			BackgroundColor: color.RGBA{R: 20, G: 20, B: 40, A: 180},
			FillColor:       color.RGBA{R: 0, G: 255, B: 255, A: 255},
			BorderColor:     color.RGBA{R: 255, G: 0, B: 180, A: 255},
			GlowColor:       color.RGBA{R: 0, G: 255, B: 255, A: 100},
			TextColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
			BarWidth:        60,
			BarHeight:       6,
			BorderWidth:     1.5,
			CornerRadius:    2,
			GlowIntensity:   0.8,
			PulseSpeed:      8.0,
			YOffset:         25,
		}
	case "horror":
		s.style = GenreStyle{
			BackgroundColor: color.RGBA{R: 30, G: 10, B: 10, A: 200},
			FillColor:       color.RGBA{R: 180, G: 40, B: 40, A: 255},
			BorderColor:     color.RGBA{R: 100, G: 20, B: 20, A: 255},
			GlowColor:       color.RGBA{R: 180, G: 40, B: 40, A: 80},
			TextColor:       color.RGBA{R: 200, G: 180, B: 180, A: 255},
			BarWidth:        55,
			BarHeight:       5,
			BorderWidth:     1.0,
			CornerRadius:    0,
			GlowIntensity:   0.4,
			PulseSpeed:      3.0,
			YOffset:         22,
		}
	case "scifi":
		s.style = GenreStyle{
			BackgroundColor: color.RGBA{R: 10, G: 20, B: 40, A: 180},
			FillColor:       color.RGBA{R: 100, G: 180, B: 255, A: 255},
			BorderColor:     color.RGBA{R: 150, G: 200, B: 255, A: 255},
			GlowColor:       color.RGBA{R: 100, G: 180, B: 255, A: 90},
			TextColor:       color.RGBA{R: 220, G: 240, B: 255, A: 255},
			BarWidth:        65,
			BarHeight:       5,
			BorderWidth:     1.0,
			CornerRadius:    3,
			GlowIntensity:   0.6,
			PulseSpeed:      6.0,
			YOffset:         24,
		}
	case "postapoc":
		s.style = GenreStyle{
			BackgroundColor: color.RGBA{R: 40, G: 30, B: 20, A: 190},
			FillColor:       color.RGBA{R: 255, G: 150, B: 50, A: 255},
			BorderColor:     color.RGBA{R: 180, G: 100, B: 40, A: 255},
			GlowColor:       color.RGBA{R: 255, G: 150, B: 50, A: 70},
			TextColor:       color.RGBA{R: 230, G: 200, B: 150, A: 255},
			BarWidth:        50,
			BarHeight:       6,
			BorderWidth:     1.5,
			CornerRadius:    1,
			GlowIntensity:   0.3,
			PulseSpeed:      4.0,
			YOffset:         22,
		}
	default: // fantasy
		s.style = GenreStyle{
			BackgroundColor: color.RGBA{R: 30, G: 25, B: 40, A: 180},
			FillColor:       color.RGBA{R: 255, G: 200, B: 80, A: 255},
			BorderColor:     color.RGBA{R: 200, G: 150, B: 50, A: 255},
			GlowColor:       color.RGBA{R: 255, G: 200, B: 80, A: 80},
			TextColor:       color.RGBA{R: 255, G: 240, B: 200, A: 255},
			BarWidth:        55,
			BarHeight:       5,
			BorderWidth:     1.0,
			CornerRadius:    2,
			GlowIntensity:   0.5,
			PulseSpeed:      5.0,
			YOffset:         23,
		}
	}
}

// SetScreenSize updates screen dimensions for positioning.
func (s *System) SetScreenSize(width, height int) {
	s.screenWidth = width
	s.screenHeight = height
}

// SetReloadState updates the current reload progress (non-ECS mode).
func (s *System) SetReloadState(isReloading bool, progress, totalDuration float64) {
	s.isReloading = isReloading
	s.progress = progress
	s.totalDuration = totalDuration
}

// Update processes reload bar state and animations.
func (s *System) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0 // Assuming 60 TPS

	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, eid := range entities {
		comp, ok := w.GetComponent(eid, compType)
		if !ok {
			continue
		}
		rc := comp.(*Component)

		// Update progress
		rc.UpdateProgress(deltaTime)

		// Update fade
		if rc.IsReloading {
			rc.FadeAlpha = math.Min(1.0, rc.FadeAlpha+deltaTime*6.0)
		} else {
			rc.FadeAlpha = math.Max(0.0, rc.FadeAlpha-deltaTime*4.0)
		}
	}

	// Update non-ECS state
	s.pulsePhase += s.style.PulseSpeed * deltaTime
	if s.pulsePhase > 2*math.Pi {
		s.pulsePhase -= 2 * math.Pi
	}

	// Fade handling
	if s.isReloading {
		s.fadeAlpha = math.Min(1.0, s.fadeAlpha+deltaTime*6.0)
	} else {
		s.fadeAlpha = math.Max(0.0, s.fadeAlpha-deltaTime*4.0)
	}
}

// Render draws the reload bar at the specified screen position (centered).
func (s *System) Render(screen *ebiten.Image, centerX, centerY float32) {
	if s.fadeAlpha < 0.01 {
		return
	}

	// Position bar below crosshair
	barX := centerX - s.style.BarWidth/2
	barY := centerY + s.style.YOffset

	s.drawReloadBar(screen, barX, barY, float32(s.progress), float32(s.fadeAlpha))
}

// RenderForEntity draws the reload bar for a specific entity component.
func (s *System) RenderForEntity(screen *ebiten.Image, centerX, centerY float32, rc *Component) {
	if rc.FadeAlpha < 0.01 {
		return
	}

	barX := centerX - s.style.BarWidth/2
	barY := centerY + s.style.YOffset

	s.drawReloadBar(screen, barX, barY, float32(rc.Progress), float32(rc.FadeAlpha))
}

// drawReloadBar renders the actual reload bar graphics.
func (s *System) drawReloadBar(screen *ebiten.Image, x, y, progress, alpha float32) {
	width := s.style.BarWidth
	height := s.style.BarHeight

	// Apply pulse animation to progress bar
	pulseScale := float32(1.0 + 0.05*math.Sin(s.pulsePhase))

	// Glow effect (drawn first, behind bar)
	if s.style.GlowIntensity > 0 {
		glowCol := s.style.GlowColor
		glowCol.A = uint8(float32(glowCol.A) * alpha * s.style.GlowIntensity * pulseScale)
		glowExpand := float32(3)
		s.drawRoundedRect(screen, x-glowExpand, y-glowExpand,
			width+glowExpand*2, height+glowExpand*2,
			s.style.CornerRadius+glowExpand, glowCol)
	}

	// Background
	bgCol := s.style.BackgroundColor
	bgCol.A = uint8(float32(bgCol.A) * alpha)
	s.drawRoundedRect(screen, x, y, width, height, s.style.CornerRadius, bgCol)

	// Fill bar (progress)
	if progress > 0 {
		fillWidth := width * progress
		fillCol := s.style.FillColor
		fillCol.A = uint8(float32(fillCol.A) * alpha)

		// Add pulse brightness
		brightness := 1.0 + 0.15*float32(math.Sin(s.pulsePhase))
		fillCol.R = uint8(math.Min(255, float64(fillCol.R)*float64(brightness)))
		fillCol.G = uint8(math.Min(255, float64(fillCol.G)*float64(brightness)))
		fillCol.B = uint8(math.Min(255, float64(fillCol.B)*float64(brightness)))

		s.drawRoundedRect(screen, x, y, fillWidth, height, s.style.CornerRadius, fillCol)

		// Highlight line at fill edge
		if progress < 1.0 && progress > 0.05 {
			highlightCol := color.RGBA{R: 255, G: 255, B: 255, A: uint8(150 * alpha)}
			edgeX := x + fillWidth - 1
			vector.StrokeLine(screen, edgeX, y+1, edgeX, y+height-1, 1.0, highlightCol, false)
		}
	}

	// Border
	borderCol := s.style.BorderColor
	borderCol.A = uint8(float32(borderCol.A) * alpha)
	s.drawRoundedRectStroke(screen, x, y, width, height, s.style.CornerRadius, s.style.BorderWidth, borderCol)

	// Segment markers (every 25%)
	for i := 1; i < 4; i++ {
		markerX := x + width*float32(i)/4.0
		markerCol := s.style.BorderColor
		markerCol.A = uint8(float32(markerCol.A) * alpha * 0.5)
		vector.StrokeLine(screen, markerX, y+1, markerX, y+height-1, 0.5, markerCol, false)
	}
}

// drawRoundedRect draws a filled rounded rectangle.
func (s *System) drawRoundedRect(screen *ebiten.Image, x, y, width, height, radius float32, col color.RGBA) {
	if radius <= 0 {
		// Simple rectangle
		vector.DrawFilledRect(screen, x, y, width, height, col, false)
		return
	}

	// Clamp radius
	maxRadius := math.Min(float64(width), float64(height)) / 2
	if float64(radius) > maxRadius {
		radius = float32(maxRadius)
	}

	// Draw rounded corners using circles and rectangles
	// Main body (horizontal)
	vector.DrawFilledRect(screen, x+radius, y, width-radius*2, height, col, false)
	// Main body (vertical)
	vector.DrawFilledRect(screen, x, y+radius, width, height-radius*2, col, false)

	// Corner circles
	vector.DrawFilledCircle(screen, x+radius, y+radius, radius, col, false)
	vector.DrawFilledCircle(screen, x+width-radius, y+radius, radius, col, false)
	vector.DrawFilledCircle(screen, x+radius, y+height-radius, radius, col, false)
	vector.DrawFilledCircle(screen, x+width-radius, y+height-radius, radius, col, false)
}

// drawRoundedRectStroke draws a stroked rounded rectangle border.
func (s *System) drawRoundedRectStroke(screen *ebiten.Image, x, y, width, height, radius, strokeWidth float32, col color.RGBA) {
	if radius <= 0 {
		// Simple rectangle border
		vector.StrokeLine(screen, x, y, x+width, y, strokeWidth, col, false)
		vector.StrokeLine(screen, x+width, y, x+width, y+height, strokeWidth, col, false)
		vector.StrokeLine(screen, x+width, y+height, x, y+height, strokeWidth, col, false)
		vector.StrokeLine(screen, x, y+height, x, y, strokeWidth, col, false)
		return
	}

	// Clamp radius
	maxRadius := math.Min(float64(width), float64(height)) / 2
	if float64(radius) > maxRadius {
		radius = float32(maxRadius)
	}

	// Top edge
	vector.StrokeLine(screen, x+radius, y, x+width-radius, y, strokeWidth, col, false)
	// Bottom edge
	vector.StrokeLine(screen, x+radius, y+height, x+width-radius, y+height, strokeWidth, col, false)
	// Left edge
	vector.StrokeLine(screen, x, y+radius, x, y+height-radius, strokeWidth, col, false)
	// Right edge
	vector.StrokeLine(screen, x+width, y+radius, x+width, y+height-radius, strokeWidth, col, false)

	// Corner arcs (approximated with arc strokes)
	s.drawArcStroke(screen, x+radius, y+radius, radius, math.Pi, math.Pi*1.5, strokeWidth, col)
	s.drawArcStroke(screen, x+width-radius, y+radius, radius, math.Pi*1.5, math.Pi*2, strokeWidth, col)
	s.drawArcStroke(screen, x+radius, y+height-radius, radius, math.Pi*0.5, math.Pi, strokeWidth, col)
	s.drawArcStroke(screen, x+width-radius, y+height-radius, radius, 0, math.Pi*0.5, strokeWidth, col)
}

// drawArcStroke draws an arc stroke.
func (s *System) drawArcStroke(screen *ebiten.Image, cx, cy, radius float32, startAngle, endAngle float64, strokeWidth float32, col color.RGBA) {
	const segments = 8
	angleStep := (endAngle - startAngle) / float64(segments)

	for i := 0; i < segments; i++ {
		a1 := startAngle + float64(i)*angleStep
		a2 := startAngle + float64(i+1)*angleStep

		x1 := cx + radius*float32(math.Cos(a1))
		y1 := cy + radius*float32(math.Sin(a1))
		x2 := cx + radius*float32(math.Cos(a2))
		y2 := cy + radius*float32(math.Sin(a2))

		vector.StrokeLine(screen, x1, y1, x2, y2, strokeWidth, col, false)
	}
}

// IsActive returns whether the reload bar is currently visible.
func (s *System) IsActive() bool {
	return s.isReloading || s.fadeAlpha > 0.01
}

// GetProgress returns the current reload progress (0.0 to 1.0).
func (s *System) GetProgress() float64 {
	return s.progress
}

// GetStyle returns the current genre style (for testing).
func (s *System) GetStyle() GenreStyle {
	return s.style
}

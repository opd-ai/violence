// Package entitylabel provides entity name rendering with guaranteed text display.
package entitylabel

import (
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/ui"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// System handles entity label rendering with guaranteed text display.
type System struct {
	genre        string
	font         font.Face
	fallbackFont font.Face
	logger       *logrus.Entry
	textCache    map[string]*ebiten.Image
	maxCache     int
}

// NewSystem creates an entity label rendering system.
func NewSystem(genre string) *System {
	return &System{
		genre:        genre,
		font:         basicfont.Face7x13,
		fallbackFont: basicfont.Face7x13, // Same for now, but separate for future customization
		logger: logrus.WithFields(logrus.Fields{
			"system":  "entitylabel",
			"package": "entitylabel",
		}),
		textCache: make(map[string]*ebiten.Image),
		maxCache:  100,
	}
}

// Update processes entity labels (currently no per-frame updates needed).
func (s *System) Update(w *engine.World) {
	// No per-frame logic needed for static labels
	// Future: could add animation, pulsing effects, etc.
}

// SetGenre updates genre-specific settings.
func (s *System) SetGenre(genre string) {
	s.genre = genre
	s.applyGenreTheme()
}

// applyGenreTheme adjusts rendering based on genre.
func (s *System) applyGenreTheme() {
	// Future: load genre-specific fonts or styling
	// For now, all genres use basicfont.Face7x13
}

// Render draws entity labels to screen without layout manager.
func (s *System) Render(w *engine.World, screen *ebiten.Image, cameraX, cameraY float64) {
	s.RenderWithLayout(w, screen, cameraX, cameraY, nil)
}

// RenderWithLayout draws entity labels using layout manager to prevent overlap.
func (s *System) RenderWithLayout(w *engine.World, screen *ebiten.Image, cameraX, cameraY float64, layoutMgr *ui.LayoutManager) {
	labelType := reflect.TypeOf((*Component)(nil))
	posType := reflect.TypeOf((*engine.Position)(nil))

	entities := w.Query(labelType, posType)

	for _, eid := range entities {
		labelComp, ok := w.GetComponent(eid, labelType)
		if !ok {
			continue
		}
		label := labelComp.(*Component)

		posComp, ok := w.GetComponent(eid, posType)
		if !ok {
			continue
		}
		pos := posComp.(*engine.Position)

		// Calculate distance from camera
		dx := pos.X - cameraX
		dy := pos.Y - cameraY
		distance := math.Sqrt(dx*dx + dy*dy)

		// Skip if too far away
		if distance > label.MaxDistance {
			continue
		}

		// Skip if not always visible and too far
		if !label.AlwaysVisible && distance > label.MaxDistance*0.7 {
			continue
		}

		// Convert world position to screen position
		// Assuming top-down view with 16px per world unit
		const pixelsPerUnit = 16.0
		screenX := int((pos.X - cameraX) * pixelsPerUnit)
		screenY := int((pos.Y - cameraY) * pixelsPerUnit)

		// Apply vertical offset
		screenY += int(label.OffsetY)

		// Center the screen position
		bounds := screen.Bounds()
		centerX := bounds.Dx() / 2
		centerY := bounds.Dy() / 2
		finalX := centerX + screenX
		finalY := centerY + screenY

		// Skip if off-screen
		if finalX < -100 || finalX > bounds.Dx()+100 || finalY < -50 || finalY > bounds.Dy()+50 {
			continue
		}

		// Calculate alpha based on distance (fade at max range)
		alpha := 1.0
		if distance > label.MaxDistance*0.5 {
			fadeRange := label.MaxDistance * 0.5
			alpha = 1.0 - (distance-label.MaxDistance*0.5)/fadeRange
			alpha = math.Max(0.0, math.Min(1.0, alpha))
		}

		// Render the label
		s.renderLabel(screen, label, finalX, finalY, alpha, layoutMgr)
	}
}

// renderLabel draws a single label with text, background, and border.
func (s *System) renderLabel(screen *ebiten.Image, label *Component, x, y int, alpha float64, layoutMgr *ui.LayoutManager) {
	// Measure text
	textBounds := text.BoundString(s.font, label.Text)
	textWidth := float64(textBounds.Dx()) * label.Scale
	textHeight := float64(textBounds.Dy()) * label.Scale

	// Calculate background box dimensions
	padding := 4.0
	boxWidth := textWidth + padding*2
	boxHeight := textHeight + padding*2

	// Center the label horizontally
	boxX := float64(x) - boxWidth/2
	boxY := float64(y) - boxHeight/2

	// Check with layout manager for overlap
	if layoutMgr != nil {
		// Convert priority to layout manager priority
		var layoutPriority ui.Priority
		switch label.Priority {
		case 0:
			layoutPriority = ui.PrioritySecondary
		case 1:
			layoutPriority = ui.PriorityImportant
		case 2:
			layoutPriority = ui.PriorityCritical
		default:
			layoutPriority = ui.PriorityImportant
		}

		// Reserve space with layout manager
		adjustedX, adjustedY, _ := layoutMgr.Reserve(
			"label", float32(boxX), float32(boxY),
			float32(boxWidth), float32(boxHeight),
			layoutPriority,
			true, // canMove
		)
		boxX = float64(adjustedX)
		boxY = float64(adjustedY)
	}

	// Apply alpha to colors
	bgColor := label.BackgroundColor
	bgColor.A = uint8(float64(bgColor.A) * alpha)

	borderColor := label.BorderColor
	borderColor.A = uint8(float64(borderColor.A) * alpha)

	textColor := label.Color
	textColor.A = uint8(float64(textColor.A) * alpha)

	// Draw background box if enabled
	if label.ShowBackground {
		// Background fill
		vector.DrawFilledRect(
			screen,
			float32(boxX), float32(boxY),
			float32(boxWidth), float32(boxHeight),
			bgColor,
			false,
		)

		// Border
		vector.StrokeRect(
			screen,
			float32(boxX), float32(boxY),
			float32(boxWidth), float32(boxHeight),
			1.0,
			borderColor,
			false,
		)
	}

	// Draw text with fallback handling
	textX := int(boxX + padding)
	textY := int(boxY + padding + float64(textBounds.Dy()))

	if label.Scale == 1.0 {
		// Direct rendering at normal scale
		s.drawTextWithFallback(screen, label.Text, textX, textY, textColor)
	} else {
		// Scaled rendering via temporary image
		tmpImg := ebiten.NewImage(textBounds.Dx()+4, textBounds.Dy()+4)
		s.drawTextWithFallback(tmpImg, label.Text, 2, textBounds.Dy(), textColor)

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(label.Scale, label.Scale)
		opts.GeoM.Translate(float64(textX), float64(textY)-float64(textBounds.Dy())*label.Scale)
		opts.ColorScale.ScaleAlpha(float32(alpha))
		screen.DrawImage(tmpImg, opts)
	}
}

// drawTextWithFallback attempts to draw text with primary font, falls back if needed.
func (s *System) drawTextWithFallback(dst *ebiten.Image, str string, x, y int, clr color.Color) {
	// Try primary font
	defer func() {
		if r := recover(); r != nil {
			// Primary font failed, try fallback
			s.logger.WithFields(logrus.Fields{
				"text":  str,
				"error": r,
			}).Warn("Primary font rendering failed, using fallback")

			defer func() {
				if r2 := recover(); r2 != nil {
					// Fallback also failed, draw placeholder
					s.logger.WithFields(logrus.Fields{
						"text":  str,
						"error": r2,
					}).Error("Fallback font rendering failed, drawing placeholder")
					s.drawPlaceholder(dst, x, y, clr)
				}
			}()

			text.Draw(dst, str, s.fallbackFont, x, y, clr)
		}
	}()

	text.Draw(dst, str, s.font, x, y, clr)
}

// drawPlaceholder draws a simple rectangle when all font rendering fails.
func (s *System) drawPlaceholder(dst *ebiten.Image, x, y int, clr color.Color) {
	// Draw a small colored rectangle as absolute fallback
	r, g, b, a := clr.RGBA()
	placeholderColor := color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}

	vector.DrawFilledRect(
		dst,
		float32(x), float32(y-8),
		20, 10,
		placeholderColor,
		false,
	)
}

// ClearCache removes all cached text images.
func (s *System) ClearCache() {
	s.textCache = make(map[string]*ebiten.Image)
}

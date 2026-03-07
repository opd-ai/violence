package crosshair

import (
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System renders crosshairs for entities with crosshair components.
type System struct {
	logger  *logrus.Entry
	genreID string
}

// NewSystem creates a crosshair rendering system.
func NewSystem(genreID string) *System {
	return &System{
		logger: logrus.WithFields(logrus.Fields{
			"system": "crosshair",
			"genre":  genreID,
		}),
		genreID: genreID,
	}
}

// SetGenre updates the genre-specific rendering style.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.logger = s.logger.WithField("genre", genreID)
}

// Update implements the System interface.
func (s *System) Update(w *engine.World) {
	// Crosshair system is render-only, no update logic needed
}

// Render draws crosshairs for all entities with crosshair components.
func (s *System) Render(screen *ebiten.Image, w *engine.World, cameraX, cameraY float64, screenWidth, screenHeight int) {
	crosshairType := reflect.TypeOf(&Component{})
	positionType := reflect.TypeOf(&engine.Position{})

	entities := w.Query(crosshairType)

	for _, entityID := range entities {
		crosshairComp, ok := w.GetComponent(entityID, crosshairType)
		if !ok {
			continue
		}
		ch := crosshairComp.(*Component)

		if !ch.Visible {
			continue
		}

		posComp, ok := w.GetComponent(entityID, positionType)
		if !ok {
			continue
		}
		pos := posComp.(*engine.Position)

		s.renderCrosshair(screen, pos.X, pos.Y, ch, cameraX, cameraY, screenWidth, screenHeight)
	}
}

// renderCrosshair draws a single crosshair at the specified world position.
func (s *System) renderCrosshair(screen *ebiten.Image, entityX, entityY float64, ch *Component, cameraX, cameraY float64, screenWidth, screenHeight int) {
	// Calculate crosshair position in world space
	crosshairWorldX := entityX + ch.AimX*ch.Range
	crosshairWorldY := entityY + ch.AimY*ch.Range

	// Convert world position to screen space
	dx := crosshairWorldX - cameraX
	dy := crosshairWorldY - cameraY

	screenX := float32(screenWidth/2) + float32(dx*10)
	screenY := float32(screenHeight/2) + float32(dy*10)

	// Early exit if off-screen
	margin := float32(50)
	if screenX < -margin || screenX > float32(screenWidth)+margin ||
		screenY < -margin || screenY > float32(screenHeight)+margin {
		return
	}

	// Choose rendering style based on weapon type
	crosshairColor := color.RGBA{
		R: uint8(ch.ColorR * 255),
		G: uint8(ch.ColorG * 255),
		B: uint8(ch.ColorB * 255),
		A: uint8(ch.ColorA * 255),
	}

	switch ch.WeaponType {
	case "ranged":
		s.renderRangedCrosshair(screen, screenX, screenY, float32(ch.Scale), crosshairColor)
	case "magic":
		s.renderMagicCrosshair(screen, screenX, screenY, float32(ch.Scale), crosshairColor)
	case "melee":
		s.renderMeleeCrosshair(screen, screenX, screenY, float32(ch.Scale), crosshairColor)
	default:
		s.renderMeleeCrosshair(screen, screenX, screenY, float32(ch.Scale), crosshairColor)
	}
}

// renderMeleeCrosshair draws a simple dot or circle for melee weapons.
func (s *System) renderMeleeCrosshair(screen *ebiten.Image, x, y, scale float32, col color.RGBA) {
	radius := 3.0 * scale

	// Draw filled circle
	vector.DrawFilledCircle(screen, x, y, radius, col, false)

	// Draw outer ring for better visibility
	outerColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(float32(col.A) * 0.5)}
	vector.StrokeCircle(screen, x, y, radius+1, 1, outerColor, false)
}

// renderRangedCrosshair draws a precision crosshair for ranged weapons.
func (s *System) renderRangedCrosshair(screen *ebiten.Image, x, y, scale float32, col color.RGBA) {
	size := 8.0 * scale
	gap := 3.0 * scale
	thickness := 1.5 * scale

	// Draw four lines forming a crosshair with a gap in the center
	// Top line
	vector.StrokeLine(screen, x, y-gap-size, x, y-gap, thickness, col, false)
	// Bottom line
	vector.StrokeLine(screen, x, y+gap, x, y+gap+size, thickness, col, false)
	// Left line
	vector.StrokeLine(screen, x-gap-size, y, x-gap, y, thickness, col, false)
	// Right line
	vector.StrokeLine(screen, x+gap, y, x+gap+size, y, thickness, col, false)

	// Center dot
	vector.DrawFilledCircle(screen, x, y, 1.5*scale, col, false)
}

// renderMagicCrosshair draws a circular reticle for magic attacks.
func (s *System) renderMagicCrosshair(screen *ebiten.Image, x, y, scale float32, col color.RGBA) {
	radius := 6.0 * scale
	innerRadius := 4.0 * scale

	// Draw outer circle
	vector.StrokeCircle(screen, x, y, radius, 1.5*scale, col, false)

	// Draw inner circle with lower opacity
	innerColor := color.RGBA{R: col.R, G: col.G, B: col.B, A: uint8(float32(col.A) * 0.6)}
	vector.StrokeCircle(screen, x, y, innerRadius, 1.0*scale, innerColor, false)

	// Draw four corner marks at cardinal directions
	markLen := 3.0 * scale
	markOffset := radius + 2.0*scale
	for i := 0; i < 4; i++ {
		angle := float64(i) * math.Pi / 2.0
		startX := x + float32(math.Cos(angle)*float64(markOffset))
		startY := y + float32(math.Sin(angle)*float64(markOffset))
		endX := x + float32(math.Cos(angle)*float64(markOffset+markLen))
		endY := y + float32(math.Sin(angle)*float64(markOffset+markLen))
		vector.StrokeLine(screen, startX, startY, endX, endY, 1.5*scale, col, false)
	}
}

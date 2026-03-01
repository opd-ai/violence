package input

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// TouchRenderStyle defines visual style for touch controls
type TouchRenderStyle int

const (
	StyleDefault   TouchRenderStyle = iota
	StyleHorror                     // Rune circles with dark red accents
	StyleCyberpunk                  // Neon hexagons with cyan/magenta
	StylePostApoc                   // Rusted metal with orange glow
	StyleSciFi                      // Clean geometric with blue highlights
	StyleFantasy                    // Ornate circles with golden trim
)

// TouchRenderer handles rendering of touch controls with genre-specific styles
type TouchRenderer struct {
	style        TouchRenderStyle
	joystickImg  *ebiten.Image
	buttonImgs   map[string]*ebiten.Image
	overlayAlpha uint8
}

// NewTouchRenderer creates a new touch renderer with the specified style
func NewTouchRenderer(style TouchRenderStyle) *TouchRenderer {
	return &TouchRenderer{
		style:        style,
		buttonImgs:   make(map[string]*ebiten.Image),
		overlayAlpha: 128, // 50% transparency
	}
}

// SetStyle changes the visual style for touch controls
func (tr *TouchRenderer) SetStyle(style TouchRenderStyle) {
	tr.style = style
	// Clear cached images to force regeneration
	tr.joystickImg = nil
	tr.buttonImgs = make(map[string]*ebiten.Image)
}

// RenderJoystick draws the virtual joystick on screen
func (tr *TouchRenderer) RenderJoystick(screen *ebiten.Image, vj *VirtualJoystick) {
	if !vj.Active {
		return
	}

	// Draw outer circle (base)
	baseColor := tr.getBaseColor()
	tr.drawCircle(screen, vj.CenterX, vj.CenterY, vj.MaxRadius, baseColor)

	// Draw dead zone indicator
	deadColor := tr.getDeadZoneColor()
	tr.drawCircle(screen, vj.CenterX, vj.CenterY, vj.DeadZone, deadColor)

	// Draw knob (current position)
	knobColor := tr.getKnobColor()
	tr.drawCircle(screen, vj.KnobX, vj.KnobY, vj.MaxRadius*0.4, knobColor)

	// Add genre-specific accents
	tr.drawJoystickAccent(screen, vj)
}

// RenderButton draws a touch button on screen
func (tr *TouchRenderer) RenderButton(screen *ebiten.Image, btn *TouchButton, screenW, screenH int) {
	btnX := btn.X * float64(screenW)
	btnY := btn.Y * float64(screenH)

	// Draw button base
	baseColor := tr.getButtonColor(btn.Active)
	tr.drawButtonShape(screen, btnX, btnY, btn.Radius, baseColor)

	// Draw button accent based on genre
	tr.drawButtonAccent(screen, btnX, btnY, btn.Radius, btn.Active)
}

// drawCircle draws a filled circle at the specified position
func (tr *TouchRenderer) drawCircle(screen *ebiten.Image, x, y, radius float64, clr color.Color) {
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(radius), clr, false)
}

// drawButtonShape draws the button shape based on genre style
func (tr *TouchRenderer) drawButtonShape(screen *ebiten.Image, x, y, radius float64, clr color.Color) {
	switch tr.style {
	case StyleCyberpunk:
		tr.drawHexagon(screen, x, y, radius, clr)
	case StyleHorror:
		tr.drawRuneCircle(screen, x, y, radius, clr)
	default:
		tr.drawCircle(screen, x, y, radius, clr)
	}
}

// drawHexagon draws a hexagonal button shape
func (tr *TouchRenderer) drawHexagon(screen *ebiten.Image, centerX, centerY, radius float64, clr color.Color) {
	// Draw hexagon using 6 points
	points := make([]float32, 0, 14)
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		px := centerX + radius*math.Cos(angle)
		py := centerY + radius*math.Sin(angle)
		points = append(points, float32(px), float32(py))
	}
	// Close the path
	points = append(points, points[0], points[1])

	// Draw filled hexagon
	for i := 0; i < len(points)-2; i += 2 {
		vector.StrokeLine(screen, points[i], points[i+1], points[i+2], points[i+3], 3.0, clr, false)
	}
}

// drawRuneCircle draws a circular button with runic accents
func (tr *TouchRenderer) drawRuneCircle(screen *ebiten.Image, x, y, radius float64, clr color.Color) {
	// Draw outer ring
	tr.drawCircle(screen, x, y, radius, clr)
	// Draw inner ring for depth
	innerColor := color.RGBA{R: 20, G: 10, B: 10, A: tr.overlayAlpha}
	tr.drawCircle(screen, x, y, radius*0.7, innerColor)
}

// drawJoystickAccent adds genre-specific visual accents to the joystick
func (tr *TouchRenderer) drawJoystickAccent(screen *ebiten.Image, vj *VirtualJoystick) {
	switch tr.style {
	case StyleCyberpunk:
		// Draw neon glow rings
		glowColor := color.RGBA{R: 0, G: 255, B: 255, A: 64}
		vector.StrokeCircle(screen, float32(vj.CenterX), float32(vj.CenterY), float32(vj.MaxRadius), 2.0, glowColor, false)
	case StyleHorror:
		// Draw dark red pulse
		pulseColor := color.RGBA{R: 139, G: 0, B: 0, A: 96}
		vector.StrokeCircle(screen, float32(vj.CenterX), float32(vj.CenterY), float32(vj.MaxRadius*1.1), 3.0, pulseColor, false)
	case StylePostApoc:
		// Draw orange radiation glow
		glowColor := color.RGBA{R: 255, G: 140, B: 0, A: 80}
		vector.StrokeCircle(screen, float32(vj.CenterX), float32(vj.CenterY), float32(vj.MaxRadius), 2.5, glowColor, false)
	}
}

// drawButtonAccent adds genre-specific accents to buttons
func (tr *TouchRenderer) drawButtonAccent(screen *ebiten.Image, x, y, radius float64, active bool) {
	switch tr.style {
	case StyleCyberpunk:
		// Neon outline
		accentColor := color.RGBA{R: 255, G: 0, B: 255, A: 200}
		if active {
			accentColor = color.RGBA{R: 0, G: 255, B: 255, A: 255}
		}
		vector.StrokeCircle(screen, float32(x), float32(y), float32(radius), 3.0, accentColor, false)
	case StyleHorror:
		// Dark red glow
		accentColor := color.RGBA{R: 139, G: 0, B: 0, A: 128}
		if active {
			accentColor = color.RGBA{R: 220, G: 20, B: 60, A: 200}
		}
		vector.StrokeCircle(screen, float32(x), float32(y), float32(radius*1.1), 2.0, accentColor, false)
	}
}

// getBaseColor returns the base color for joystick background
func (tr *TouchRenderer) getBaseColor() color.Color {
	switch tr.style {
	case StyleHorror:
		return color.RGBA{R: 40, G: 20, B: 20, A: tr.overlayAlpha}
	case StyleCyberpunk:
		return color.RGBA{R: 20, G: 20, B: 60, A: tr.overlayAlpha}
	case StylePostApoc:
		return color.RGBA{R: 60, G: 40, B: 20, A: tr.overlayAlpha}
	case StyleSciFi:
		return color.RGBA{R: 20, G: 40, B: 80, A: tr.overlayAlpha}
	case StyleFantasy:
		return color.RGBA{R: 60, G: 50, B: 30, A: tr.overlayAlpha}
	default:
		return color.RGBA{R: 80, G: 80, B: 80, A: tr.overlayAlpha}
	}
}

// getDeadZoneColor returns the dead zone indicator color
func (tr *TouchRenderer) getDeadZoneColor() color.Color {
	switch tr.style {
	case StyleHorror:
		return color.RGBA{R: 20, G: 10, B: 10, A: tr.overlayAlpha / 2}
	case StyleCyberpunk:
		return color.RGBA{R: 10, G: 10, B: 30, A: tr.overlayAlpha / 2}
	default:
		return color.RGBA{R: 40, G: 40, B: 40, A: tr.overlayAlpha / 2}
	}
}

// getKnobColor returns the joystick knob color
func (tr *TouchRenderer) getKnobColor() color.Color {
	switch tr.style {
	case StyleHorror:
		return color.RGBA{R: 139, G: 0, B: 0, A: tr.overlayAlpha + 64}
	case StyleCyberpunk:
		return color.RGBA{R: 0, G: 255, B: 255, A: tr.overlayAlpha + 64}
	case StylePostApoc:
		return color.RGBA{R: 255, G: 140, B: 0, A: tr.overlayAlpha + 64}
	case StyleSciFi:
		return color.RGBA{R: 0, G: 150, B: 255, A: tr.overlayAlpha + 64}
	case StyleFantasy:
		return color.RGBA{R: 218, G: 165, B: 32, A: tr.overlayAlpha + 64}
	default:
		return color.RGBA{R: 150, G: 150, B: 150, A: tr.overlayAlpha + 64}
	}
}

// getButtonColor returns button color based on active state
func (tr *TouchRenderer) getButtonColor(active bool) color.Color {
	baseAlpha := calculateButtonAlpha(tr.overlayAlpha, active)
	return selectButtonColorByStyle(tr.style, active, baseAlpha)
}

// calculateButtonAlpha computes the alpha value for a button based on active state.
func calculateButtonAlpha(overlayAlpha uint8, active bool) uint8 {
	if active {
		return overlayAlpha + 64
	}
	return overlayAlpha
}

// selectButtonColorByStyle returns the appropriate color for a button based on style and state.
func selectButtonColorByStyle(style TouchRenderStyle, active bool, alpha uint8) color.Color {
	switch style {
	case StyleHorror:
		return getHorrorButtonColor(active, alpha)
	case StyleCyberpunk:
		return getCyberpunkButtonColor(active, alpha)
	case StylePostApoc:
		return getPostApocButtonColor(active, alpha)
	case StyleSciFi:
		return getSciFiButtonColor(active, alpha)
	case StyleFantasy:
		return getFantasyButtonColor(active, alpha)
	default:
		return getDefaultButtonColor(active, alpha)
	}
}

// getHorrorButtonColor returns horror-themed button color.
func getHorrorButtonColor(active bool, alpha uint8) color.Color {
	if active {
		return color.RGBA{R: 180, G: 20, B: 20, A: alpha}
	}
	return color.RGBA{R: 60, G: 20, B: 20, A: alpha}
}

// getCyberpunkButtonColor returns cyberpunk-themed button color.
func getCyberpunkButtonColor(active bool, alpha uint8) color.Color {
	if active {
		return color.RGBA{R: 255, G: 0, B: 255, A: alpha}
	}
	return color.RGBA{R: 40, G: 20, B: 80, A: alpha}
}

// getPostApocButtonColor returns post-apocalyptic-themed button color.
func getPostApocButtonColor(active bool, alpha uint8) color.Color {
	if active {
		return color.RGBA{R: 255, G: 160, B: 40, A: alpha}
	}
	return color.RGBA{R: 80, G: 60, B: 30, A: alpha}
}

// getSciFiButtonColor returns sci-fi-themed button color.
func getSciFiButtonColor(active bool, alpha uint8) color.Color {
	if active {
		return color.RGBA{R: 0, G: 200, B: 255, A: alpha}
	}
	return color.RGBA{R: 30, G: 60, B: 100, A: alpha}
}

// getFantasyButtonColor returns fantasy-themed button color.
func getFantasyButtonColor(active bool, alpha uint8) color.Color {
	if active {
		return color.RGBA{R: 255, G: 215, B: 0, A: alpha}
	}
	return color.RGBA{R: 100, G: 80, B: 50, A: alpha}
}

// getDefaultButtonColor returns default button color.
func getDefaultButtonColor(active bool, alpha uint8) color.Color {
	if active {
		return color.RGBA{R: 200, G: 200, B: 200, A: alpha}
	}
	return color.RGBA{R: 100, G: 100, B: 100, A: alpha}
}

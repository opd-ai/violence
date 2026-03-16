// Package ui provides interactive element rendering with hover states, press feedback, and smooth transitions.
package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// ElementState represents the interaction state of a UI element.
type ElementState int

const (
	StateIdle    ElementState = iota // StateIdle is the idle UI element state.
	StateHover                       // StateHover is the hovered UI element state.
	StatePressed                     // StatePressed is the pressed UI element state.
	StateFocused                     // StateFocused is the focused UI element state.
)

// TransitionConfig holds easing and timing parameters.
type TransitionConfig struct {
	Duration    int // frames
	EaseFunc    EaseFunc
	CurrentTime int
}

// EaseFunc defines an easing function: t in [0,1] -> value in [0,1].
type EaseFunc func(float64) float64

// Button represents an interactive button with state and transitions.
type Button struct {
	X, Y          float32
	Width, Height float32
	Label         string
	State         ElementState
	PrevState     ElementState
	Transition    TransitionConfig
	ColorIdle     color.RGBA
	ColorHover    color.RGBA
	ColorPressed  color.RGBA
	ColorFocused  color.RGBA
	TextColor     color.RGBA
	OnClick       func()
}

// Panel represents a collapsible panel with slide/fade animations.
type Panel struct {
	X, Y          float32
	Width, Height float32
	Visible       bool
	Collapsed     bool
	Transition    TransitionConfig
	BgColor       color.RGBA
	BorderColor   color.RGBA
	Contents      []PanelElement
}

// PanelElement is any renderable UI element within a panel.
type PanelElement interface {
	Draw(screen *ebiten.Image, offsetX, offsetY float32)
}

// InteractiveSystem manages all interactive UI elements with transition animations.
type InteractiveSystem struct {
	buttons []*Button
	panels  []*Panel
	focused *Button
}

// NewInteractiveSystem creates an interactive UI system.
func NewInteractiveSystem() *InteractiveSystem {
	return &InteractiveSystem{
		buttons: make([]*Button, 0, 32),
		panels:  make([]*Panel, 0, 8),
	}
}

// AddButton registers a button for rendering and interaction.
func (s *InteractiveSystem) AddButton(btn *Button) {
	if btn.Transition.Duration == 0 {
		btn.Transition.Duration = 9 // 150ms at 60fps
		btn.Transition.EaseFunc = EaseOutCubic
	}
	s.buttons = append(s.buttons, btn)
}

// AddPanel registers a panel for rendering.
func (s *InteractiveSystem) AddPanel(panel *Panel) {
	if panel.Transition.Duration == 0 {
		panel.Transition.Duration = 12 // 200ms at 60fps
		panel.Transition.EaseFunc = EaseInOutCubic
	}
	s.panels = append(s.panels, panel)
}

// Update processes hover detection and transitions.
func (s *InteractiveSystem) Update(mouseX, mouseY int, mousePressed bool) {
	for _, btn := range s.buttons {
		// Detect hover
		inBounds := float32(mouseX) >= btn.X && float32(mouseX) <= btn.X+btn.Width &&
			float32(mouseY) >= btn.Y && float32(mouseY) <= btn.Y+btn.Height

		btn.PrevState = btn.State

		// Update state based on interaction
		if inBounds {
			if mousePressed {
				btn.State = StatePressed
			} else if btn.State == StatePressed {
				// Button released over button - trigger click
				if btn.OnClick != nil {
					btn.OnClick()
				}
				btn.State = StateHover
			} else {
				btn.State = StateHover
			}
		} else {
			if btn == s.focused {
				btn.State = StateFocused
			} else {
				btn.State = StateIdle
			}
		}

		// Start transition if state changed
		if btn.State != btn.PrevState {
			btn.Transition.CurrentTime = 0
		}

		// Advance transition
		if btn.Transition.CurrentTime < btn.Transition.Duration {
			btn.Transition.CurrentTime++
		}
	}

	// Update panels
	for _, panel := range s.panels {
		if panel.Transition.CurrentTime < panel.Transition.Duration {
			panel.Transition.CurrentTime++
		}
	}
}

// Draw renders all interactive elements with smooth transitions.
func (s *InteractiveSystem) Draw(screen *ebiten.Image) {
	for _, panel := range s.panels {
		if panel.Visible || panel.Transition.CurrentTime > 0 {
			s.drawPanel(screen, panel)
		}
	}

	for _, btn := range s.buttons {
		s.drawButton(screen, btn)
	}
}

// drawButton renders a button with state-based colors and press animation.
func (s *InteractiveSystem) drawButton(screen *ebiten.Image, btn *Button) {
	// Calculate transition progress
	t := 0.0
	if btn.Transition.Duration > 0 {
		t = float64(btn.Transition.CurrentTime) / float64(btn.Transition.Duration)
	}
	if t > 1.0 {
		t = 1.0
	}
	easedT := btn.Transition.EaseFunc(t)

	// Interpolate color based on current and previous state
	currentColor := s.getButtonColor(btn, btn.State)
	prevColor := s.getButtonColor(btn, btn.PrevState)
	bgColor := lerpColor(prevColor, currentColor, easedT)

	// Press animation - button depresses slightly
	offsetY := float32(0)
	scaleY := float32(1.0)
	if btn.State == StatePressed {
		offsetY = 2 * float32(easedT)
		scaleY = 1.0 - 0.05*float32(easedT)
	} else if btn.PrevState == StatePressed {
		offsetY = 2 * float32(1.0-easedT)
		scaleY = 1.0 - 0.05*float32(1.0-easedT)
	}

	// Draw button background with scale
	btnY := btn.Y + offsetY
	btnHeight := btn.Height * scaleY
	vector.DrawFilledRect(screen, btn.X, btnY, btn.Width, btnHeight, bgColor, false)

	// Draw border
	borderColor := color.RGBA{R: 255, G: 255, B: 255, A: 200}
	if btn.State == StateFocused {
		borderColor = btn.ColorFocused
	}
	vector.StrokeRect(screen, btn.X, btnY, btn.Width, btnHeight, 1, borderColor, false)

	// Draw text centered
	face := basicfont.Face7x13
	textBounds := text.BoundString(face, btn.Label)
	textW := textBounds.Dx()
	textH := textBounds.Dy()
	textX := int(btn.X + btn.Width/2 - float32(textW)/2)
	textY := int(btnY + btnHeight/2 + float32(textH)/2)
	text.Draw(screen, btn.Label, face, textX, textY, btn.TextColor)
}

// drawPanel renders a panel with slide/fade transition.
func (s *InteractiveSystem) drawPanel(screen *ebiten.Image, panel *Panel) {
	// Calculate transition progress
	t := 0.0
	if panel.Transition.Duration > 0 {
		t = float64(panel.Transition.CurrentTime) / float64(panel.Transition.Duration)
	}
	if t > 1.0 {
		t = 1.0
	}
	easedT := panel.Transition.EaseFunc(t)

	// Visibility transition
	alpha := uint8(255)
	offsetX := float32(0)
	if panel.Visible {
		// Sliding in from right
		alpha = uint8(255 * easedT)
		offsetX = panel.Width * float32(1.0-easedT)
	} else {
		// Sliding out to right
		alpha = uint8(255 * (1.0 - easedT))
		offsetX = panel.Width * float32(easedT)
	}

	// Draw panel background with transition
	bgColor := panel.BgColor
	bgColor.A = alpha
	vector.DrawFilledRect(screen, panel.X+offsetX, panel.Y, panel.Width, panel.Height, bgColor, false)

	// Draw border
	borderColor := panel.BorderColor
	borderColor.A = alpha
	vector.StrokeRect(screen, panel.X+offsetX, panel.Y, panel.Width, panel.Height, 1, borderColor, false)

	// Draw contents if visible enough
	if alpha > 50 {
		for _, elem := range panel.Contents {
			elem.Draw(screen, panel.X+offsetX, panel.Y)
		}
	}
}

// getButtonColor returns the color for a given button state.
func (s *InteractiveSystem) getButtonColor(btn *Button, state ElementState) color.RGBA {
	switch state {
	case StateHover:
		return btn.ColorHover
	case StatePressed:
		return btn.ColorPressed
	case StateFocused:
		return btn.ColorFocused
	default:
		return btn.ColorIdle
	}
}

// SetFocus sets keyboard focus to a button.
func (s *InteractiveSystem) SetFocus(btn *Button) {
	if s.focused != nil && s.focused != btn {
		if s.focused.State == StateFocused {
			s.focused.State = StateIdle
		}
	}
	s.focused = btn
	if btn != nil {
		btn.State = StateFocused
	}
}

// ShowPanel triggers a panel to slide in.
func (s *InteractiveSystem) ShowPanel(panel *Panel) {
	panel.Visible = true
	panel.Transition.CurrentTime = 0
}

// HidePanel triggers a panel to slide out.
func (s *InteractiveSystem) HidePanel(panel *Panel) {
	panel.Visible = false
	panel.Transition.CurrentTime = 0
}

// lerpColor linearly interpolates between two colors.
func lerpColor(c1, c2 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c1.R)*(1-t) + float64(c2.R)*t),
		G: uint8(float64(c1.G)*(1-t) + float64(c2.G)*t),
		B: uint8(float64(c1.B)*(1-t) + float64(c2.B)*t),
		A: uint8(float64(c1.A)*(1-t) + float64(c2.A)*t),
	}
}

// Easing functions
// EaseOutCubic provides a smooth deceleration curve.
func EaseOutCubic(t float64) float64 {
	t = t - 1
	return t*t*t + 1
}

// EaseInOutCubic provides smooth acceleration and deceleration.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t = 2*t - 2
	return 1 + t*t*t/2
}

// EaseOutQuad provides gentler deceleration than cubic.
func EaseOutQuad(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

// EaseInOutQuad provides gentle acceleration and deceleration.
func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - 2*(1-t)*(1-t)
}

// EaseOutElastic provides a bounce effect (for special cases).
func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	p := 0.3
	return math.Pow(2, -10*t)*math.Sin((t-p/4)*(2*math.Pi)/p) + 1
}

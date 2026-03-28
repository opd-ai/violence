package focusring

import "image/color"

// FocusableElement represents a UI element that can receive keyboard focus.
type FocusableElement struct {
	// ID uniquely identifies this focusable element.
	ID string
	// X, Y position of the element's top-left corner.
	X, Y float32
	// Width, Height of the focusable area.
	Width, Height float32
	// TabIndex determines focus order when Tab is pressed (lower = earlier).
	TabIndex int
	// Enabled indicates whether this element can currently receive focus.
	Enabled bool
	// Group allows grouping elements for arrow key navigation.
	// Elements in the same group navigate spatially with arrow keys.
	Group string
	// OnFocus is called when this element receives focus.
	OnFocus func()
	// OnBlur is called when this element loses focus.
	OnBlur func()
	// OnActivate is called when Enter/Space is pressed while focused.
	OnActivate func()
}

// Type returns the component type identifier.
func (f *FocusableElement) Type() string {
	return "FocusableElement"
}

// FocusRingConfig holds visual configuration for the focus ring.
type FocusRingConfig struct {
	// RingColor is the primary color of the focus ring.
	RingColor color.RGBA
	// GlowColor is the outer glow color (usually semi-transparent).
	GlowColor color.RGBA
	// RingThickness is the stroke width of the inner ring.
	RingThickness float32
	// GlowRadius is the radius of the outer glow effect.
	GlowRadius float32
	// CornerRadius for rounded corners (0 = square).
	CornerRadius float32
	// PulseSpeed controls the animation speed (radians per frame).
	PulseSpeed float64
	// PulseIntensity controls how much the brightness varies (0-1).
	PulseIntensity float64
	// TransitionSpeed controls focus movement interpolation (0-1 per frame).
	TransitionSpeed float64
}

// FocusState holds the current focus ring animation state.
type FocusState struct {
	// FocusedID is the ID of the currently focused element.
	FocusedID string
	// CurrentX, CurrentY are the animated position (interpolated).
	CurrentX, CurrentY float32
	// CurrentW, CurrentH are the animated dimensions.
	CurrentW, CurrentH float32
	// TargetX, TargetY are the target position of the focused element.
	TargetX, TargetY float32
	// TargetW, TargetH are the target dimensions.
	TargetW, TargetH float32
	// PulsePhase is the current phase of the pulse animation.
	PulsePhase float64
	// TransitionProgress tracks focus transition animation (0-1).
	TransitionProgress float64
	// Visible indicates if the focus ring should be drawn.
	Visible bool
}

// GenrePreset holds genre-specific focus ring configuration.
type GenrePreset struct {
	Name           string
	RingColor      color.RGBA
	GlowColor      color.RGBA
	PulseSpeed     float64
	PulseIntensity float64
}

// DefaultGenrePresets returns focus ring presets for all supported genres.
func DefaultGenrePresets() map[string]GenrePreset {
	return map[string]GenrePreset{
		"fantasy": {
			Name:           "fantasy",
			RingColor:      color.RGBA{R: 255, G: 215, B: 0, A: 255}, // Gold
			GlowColor:      color.RGBA{R: 255, G: 215, B: 0, A: 80},  // Gold glow
			PulseSpeed:     0.08,
			PulseIntensity: 0.3,
		},
		"scifi": {
			Name:           "scifi",
			RingColor:      color.RGBA{R: 0, G: 255, B: 255, A: 255}, // Cyan
			GlowColor:      color.RGBA{R: 0, G: 255, B: 255, A: 80},  // Cyan glow
			PulseSpeed:     0.12,
			PulseIntensity: 0.4,
		},
		"horror": {
			Name:           "horror",
			RingColor:      color.RGBA{R: 180, G: 30, B: 30, A: 255}, // Blood red
			GlowColor:      color.RGBA{R: 180, G: 30, B: 30, A: 60},  // Red glow
			PulseSpeed:     0.05,
			PulseIntensity: 0.5,
		},
		"cyberpunk": {
			Name:           "cyberpunk",
			RingColor:      color.RGBA{R: 255, G: 0, B: 255, A: 255}, // Magenta
			GlowColor:      color.RGBA{R: 255, G: 0, B: 255, A: 80},  // Magenta glow
			PulseSpeed:     0.15,
			PulseIntensity: 0.35,
		},
		"postapoc": {
			Name:           "postapoc",
			RingColor:      color.RGBA{R: 255, G: 140, B: 0, A: 255}, // Orange
			GlowColor:      color.RGBA{R: 255, G: 140, B: 0, A: 70},  // Orange glow
			PulseSpeed:     0.06,
			PulseIntensity: 0.25,
		},
	}
}

// DefaultConfig returns the default focus ring configuration.
func DefaultConfig() FocusRingConfig {
	return FocusRingConfig{
		RingColor:       color.RGBA{R: 255, G: 215, B: 0, A: 255},
		GlowColor:       color.RGBA{R: 255, G: 215, B: 0, A: 80},
		RingThickness:   2.0,
		GlowRadius:      6.0,
		CornerRadius:    4.0,
		PulseSpeed:      0.08,
		PulseIntensity:  0.3,
		TransitionSpeed: 0.15,
	}
}

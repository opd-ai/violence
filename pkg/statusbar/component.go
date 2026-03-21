package statusbar

import (
	"image/color"
	"time"
)

// Component stores display state for the status effect icon bar.
type Component struct {
	// Icons holds cached icon data for each active effect.
	Icons []IconState

	// Position on screen (top-left of icon bar).
	X, Y float32

	// MaxIcons limits how many icons display at once.
	MaxIcons int

	// IconSize is the pixel dimension of each square icon.
	IconSize int

	// IconSpacing is pixels between icons.
	IconSpacing int

	// Visible controls whether the bar renders at all.
	Visible bool
}

// IconState represents a single status effect icon.
type IconState struct {
	// EffectName is the status effect identifier.
	EffectName string

	// DisplayName is the human-readable name for tooltips.
	DisplayName string

	// IconType determines which procedural icon to draw.
	IconType IconType

	// Color is the effect's display color (from status.Effect.VisualColor).
	Color color.RGBA

	// DurationRemaining is how much time is left.
	DurationRemaining time.Duration

	// TotalDuration is the original duration for radial progress.
	TotalDuration time.Duration

	// StackCount is how many stacks of this effect are active.
	StackCount int

	// IsExpiring is true when < 3 seconds remain (triggers pulse animation).
	IsExpiring bool

	// PulsePhase tracks animation state for expiring effects.
	PulsePhase float64
}

// IconType categorizes icons by visual style.
type IconType int

const (
	// IconDamage is for damage-over-time effects (fire, poison, bleed).
	IconDamage IconType = iota
	// IconHeal is for healing-over-time effects.
	IconHeal
	// IconBuff is for positive stat modifiers.
	IconBuff
	// IconDebuff is for negative stat modifiers.
	IconDebuff
	// IconStun is for stun/incapacitate effects.
	IconStun
	// IconSlow is for movement speed reductions.
	IconSlow
)

// Type returns the component type identifier for ECS.
func (c *Component) Type() string {
	return "statusbar"
}

// NewComponent creates a status bar component with default settings.
func NewComponent() *Component {
	return &Component{
		Icons:       make([]IconState, 0, 8),
		X:           4,
		Y:           50, // Below health bar area
		MaxIcons:    8,
		IconSize:    16,
		IconSpacing: 2,
		Visible:     true,
	}
}

// SetPosition updates the screen position of the status bar.
func (c *Component) SetPosition(x, y float32) {
	c.X = x
	c.Y = y
}

// ClearIcons removes all icon states.
func (c *Component) ClearIcons() {
	c.Icons = c.Icons[:0]
}

// AddIcon adds a new icon state to the bar.
func (c *Component) AddIcon(icon IconState) {
	if len(c.Icons) < c.MaxIcons {
		c.Icons = append(c.Icons, icon)
	}
}

// GetIconCount returns the number of active icons.
func (c *Component) GetIconCount() int {
	return len(c.Icons)
}

// HasEffect checks if an effect is currently displayed.
func (c *Component) HasEffect(effectName string) bool {
	for _, icon := range c.Icons {
		if icon.EffectName == effectName {
			return true
		}
	}
	return false
}

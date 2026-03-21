package proximityui

import "image/color"

// DetailLevel represents the amount of UI detail to show for an entity.
type DetailLevel int

const (
	// DetailNone shows no in-world UI indicators.
	DetailNone DetailLevel = iota
	// DetailMinimal shows only health bar when damaged.
	DetailMinimal
	// DetailModerate shows health bar and name.
	DetailModerate
	// DetailFull shows all indicators (health, name, status, faction).
	DetailFull
)

// String returns a human-readable name for the detail level.
func (d DetailLevel) String() string {
	switch d {
	case DetailNone:
		return "none"
	case DetailMinimal:
		return "minimal"
	case DetailModerate:
		return "moderate"
	case DetailFull:
		return "full"
	default:
		return "unknown"
	}
}

// Component stores proximity UI settings for an entity.
type Component struct {
	// PriorityOverride forces a minimum detail level regardless of distance.
	// Use for bosses, quest NPCs, targeted entities, etc.
	// -1 means no override (use distance-based calculation).
	PriorityOverride DetailLevel

	// IsTargeted marks this entity as currently targeted by the player.
	// Targeted entities always show full detail.
	IsTargeted bool

	// IsBoss marks this entity as a boss (always at least moderate detail).
	IsBoss bool

	// IsPlayer marks this entity as a player (multiplayer - always minimal).
	IsPlayer bool

	// IsQuestNPC marks this entity as quest-relevant (always moderate).
	IsQuestNPC bool

	// CurrentDetailLevel is the computed detail level after distance and priority.
	CurrentDetailLevel DetailLevel

	// TargetDetailLevel is the detail level we're transitioning toward.
	TargetDetailLevel DetailLevel

	// TransitionProgress is 0-1 for smooth fade between detail levels.
	TransitionProgress float64

	// FadeAlpha is the current opacity multiplier for this entity's UI (0-1).
	FadeAlpha float64

	// IndicatorColor can override the default indicator color for this entity.
	IndicatorColor *color.RGBA

	// LastDistance is the cached distance from camera (for fade calculations).
	LastDistance float64
}

// Type implements engine.Component interface.
func (c *Component) Type() string {
	return "proximityui"
}

// NewComponent creates a default proximity UI component.
func NewComponent() *Component {
	return &Component{
		PriorityOverride:   -1,
		CurrentDetailLevel: DetailModerate,
		TargetDetailLevel:  DetailModerate,
		TransitionProgress: 1.0,
		FadeAlpha:          1.0,
		LastDistance:       0,
	}
}

// NewBossComponent creates a proximity UI component for boss entities.
func NewBossComponent() *Component {
	c := NewComponent()
	c.IsBoss = true
	return c
}

// NewPlayerComponent creates a proximity UI component for player entities.
func NewPlayerComponent() *Component {
	c := NewComponent()
	c.IsPlayer = true
	return c
}

// NewQuestNPCComponent creates a proximity UI component for quest NPCs.
func NewQuestNPCComponent() *Component {
	c := NewComponent()
	c.IsQuestNPC = true
	return c
}

// SetTargeted marks or unmarks this entity as targeted.
func (c *Component) SetTargeted(targeted bool) {
	c.IsTargeted = targeted
	if targeted {
		c.TargetDetailLevel = DetailFull
	}
}

// GetEffectiveAlpha returns the alpha value for UI rendering,
// accounting for detail level transitions and distance fade.
func (c *Component) GetEffectiveAlpha() float64 {
	// During transitions, blend between old and new alpha
	if c.TransitionProgress < 1.0 {
		return c.FadeAlpha * c.TransitionProgress
	}
	return c.FadeAlpha
}

// ShouldShowHealthBar returns true if health bars should render at current detail.
func (c *Component) ShouldShowHealthBar() bool {
	return c.CurrentDetailLevel >= DetailMinimal
}

// ShouldShowName returns true if entity names should render at current detail.
func (c *Component) ShouldShowName() bool {
	return c.CurrentDetailLevel >= DetailModerate
}

// ShouldShowStatusIcons returns true if status icons should render at current detail.
func (c *Component) ShouldShowStatusIcons() bool {
	return c.CurrentDetailLevel >= DetailFull
}

// ShouldShowFactionBadge returns true if faction badges should render at current detail.
func (c *Component) ShouldShowFactionBadge() bool {
	return c.CurrentDetailLevel >= DetailFull
}

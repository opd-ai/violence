package toast

import "image/color"

// Priority levels for toast notifications.
type Priority int

const (
	// PriorityLow is for ambient information.
	PriorityLow Priority = iota
	// PriorityNormal is for standard notifications.
	PriorityNormal
	// PriorityHigh is for important events.
	PriorityHigh
	// PriorityCritical is for vital information.
	PriorityCritical
)

// NotificationType categorizes the notification for icon/styling.
type NotificationType string

const (
	// TypeItem indicates an item pickup notification.
	TypeItem NotificationType = "item"
	// TypeLevelUp indicates a level up notification.
	TypeLevelUp NotificationType = "levelup"
	// TypeAchievement indicates an achievement unlock.
	TypeAchievement NotificationType = "achievement"
	// TypeQuest indicates a quest update.
	TypeQuest NotificationType = "quest"
	// TypeLoot indicates loot acquisition.
	TypeLoot NotificationType = "loot"
	// TypeSkill indicates a skill unlock or upgrade.
	TypeSkill NotificationType = "skill"
	// TypeCurrency indicates currency gain.
	TypeCurrency NotificationType = "currency"
	// TypeDeath indicates a death event.
	TypeDeath NotificationType = "death"
	// TypeWarning indicates a warning message.
	TypeWarning NotificationType = "warning"
	// TypeInfo indicates general information.
	TypeInfo NotificationType = "info"
)

// Notification represents a single toast notification.
type Notification struct {
	ID          uint64           // Unique identifier
	Type        NotificationType // Category for styling
	Message     string           // Display text
	Priority    Priority         // Display priority
	Duration    float64          // Total display time in seconds
	Elapsed     float64          // Time since creation
	State       AnimState        // Current animation state
	StateTime   float64          // Time in current state
	ScreenX     float64          // Current X position
	ScreenY     float64          // Current Y position
	TargetX     float64          // Target X for animation
	TargetY     float64          // Target Y for animation
	Alpha       float64          // Current opacity (0-1)
	Scale       float64          // Current scale (for pulse effects)
	IconColor   color.RGBA       // Icon tint color
	TextColor   color.RGBA       // Text color
	BorderColor color.RGBA       // Border color
	BGColor     color.RGBA       // Background color
}

// AnimState represents the animation state of a notification.
type AnimState int

const (
	// StateEntering is the slide-in animation.
	StateEntering AnimState = iota
	// StateVisible is the stable display state.
	StateVisible
	// StateExiting is the fade-out animation.
	StateExiting
	// StateRemoved marks the notification for removal.
	StateRemoved
)

// Component marks an entity as capable of displaying toasts.
// This is typically added to the player entity.
type Component struct {
	Enabled bool // Whether toasts are enabled
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "toast"
}

// NewComponent creates a toast display component.
func NewComponent() *Component {
	return &Component{
		Enabled: true,
	}
}

// IsActive returns true if animation is still running.
func (n *Notification) IsActive() bool {
	return n.State != StateRemoved
}

// IsVisible returns true if notification should be rendered.
func (n *Notification) IsVisible() bool {
	return n.State == StateEntering || n.State == StateVisible || n.State == StateExiting
}

// GetAlpha returns the effective alpha for rendering.
func (n *Notification) GetAlpha() uint8 {
	return uint8(n.Alpha * 255)
}

// GetProgress returns the animation progress (0-1) for current state.
func (n *Notification) GetProgress() float64 {
	switch n.State {
	case StateEntering:
		const enterDur = 0.25
		if n.StateTime >= enterDur {
			return 1.0
		}
		return n.StateTime / enterDur
	case StateExiting:
		const exitDur = 0.3
		if n.StateTime >= exitDur {
			return 1.0
		}
		return n.StateTime / exitDur
	default:
		return 1.0
	}
}

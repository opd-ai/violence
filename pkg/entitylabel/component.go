// Package entitylabel provides overhead name/label rendering for entities.
package entitylabel

import (
	"image/color"
)

// Component stores label display data for an entity.
type Component struct {
	// Display text (entity name, description, etc.)
	Text string
	// Label color (defaults to white, can be color-coded by entity type)
	Color color.RGBA
	// Visibility range in world units
	MaxDistance float64
	// Whether to show even when at full health/idle
	AlwaysVisible bool
	// Priority for layout manager (0=low, 1=normal, 2=high)
	Priority int
	// Vertical offset from entity position (in pixels)
	OffsetY float64
	// Font scale multiplier
	Scale float64
	// Whether to show a background box
	ShowBackground bool
	// Background color
	BackgroundColor color.RGBA
	// Border color (if ShowBackground is true)
	BorderColor color.RGBA
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "EntityLabel"
}

// NewComponent creates a default entity label component.
func NewComponent(text string) *Component {
	return &Component{
		Text:            text,
		Color:           color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White
		MaxDistance:     15.0,                                       // 15 world units
		AlwaysVisible:   false,
		Priority:        1, // Normal priority
		OffsetY:         -20.0,
		Scale:           1.0,
		ShowBackground:  true,
		BackgroundColor: color.RGBA{R: 0, G: 0, B: 0, A: 180},       // Semi-transparent black
		BorderColor:     color.RGBA{R: 200, G: 200, B: 200, A: 255}, // Light gray
	}
}

// NewEnemyLabel creates a label for enemy entities (red text).
func NewEnemyLabel(name string) *Component {
	label := NewComponent(name)
	label.Color = color.RGBA{R: 255, G: 100, B: 100, A: 255} // Red
	label.MaxDistance = 12.0
	label.Priority = 1
	return label
}

// NewNPCLabel creates a label for friendly NPCs (green text).
func NewNPCLabel(name string) *Component {
	label := NewComponent(name)
	label.Color = color.RGBA{R: 100, G: 255, B: 100, A: 255} // Green
	label.MaxDistance = 15.0
	label.Priority = 1
	return label
}

// NewLootLabel creates a label for loot drops (yellow text).
func NewLootLabel(name string) *Component {
	label := NewComponent(name)
	label.Color = color.RGBA{R: 255, G: 255, B: 100, A: 255} // Yellow
	label.MaxDistance = 10.0
	label.Priority = 0 // Lower priority than entities
	label.Scale = 0.9
	return label
}

// NewInteractableLabel creates a label for interactable objects (cyan text).
func NewInteractableLabel(name string) *Component {
	label := NewComponent(name)
	label.Color = color.RGBA{R: 100, G: 255, B: 255, A: 255} // Cyan
	label.MaxDistance = 8.0
	label.Priority = 2 // Higher priority when close
	label.AlwaysVisible = true
	return label
}

// NewBossLabel creates a label for boss entities (orange text, larger).
func NewBossLabel(name string) *Component {
	label := NewComponent(name)
	label.Color = color.RGBA{R: 255, G: 140, B: 0, A: 255} // Orange
	label.MaxDistance = 25.0
	label.Priority = 2 // High priority
	label.Scale = 1.3
	label.AlwaysVisible = true
	label.BackgroundColor = color.RGBA{R: 50, G: 0, B: 0, A: 200} // Dark red background
	label.BorderColor = color.RGBA{R: 255, G: 140, B: 0, A: 255}
	return label
}

package healthbar

import "image/color"

// Component stores health bar display configuration for an entity.
type Component struct {
	Visible       bool
	Width         float32
	Height        float32
	OffsetY       float32
	ShowWhenFull  bool
	ThreatLevel   int
	LastDamageAge float64
	CustomColor   *color.RGBA
}

// Type implements engine.Component interface.
func (c *Component) Type() string {
	return "healthbar"
}

// StatusIconType identifies a status effect icon.
type StatusIconType int

const (
	IconPoison StatusIconType = iota
	IconBurn
	IconFreeze
	IconStun
	IconBleed
	IconRegen
	IconShield
	IconHaste
	IconSlow
	IconWeak
	IconBerserk
	IconInvisible
)

// StatusIcon represents a single status effect icon to display.
type StatusIcon struct {
	Type     StatusIconType
	Duration float64
	Stacks   int
	Color    color.RGBA
}

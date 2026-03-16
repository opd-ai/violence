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
	IconPoison    StatusIconType = iota // IconPoison is the poison status icon.
	IconBurn                            // IconBurn is the burning status icon.
	IconFreeze                          // IconFreeze is the frozen status icon.
	IconStun                            // IconStun is the stunned status icon.
	IconBleed                           // IconBleed is the bleeding status icon.
	IconRegen                           // IconRegen is the regeneration status icon.
	IconShield                          // IconShield is the shielded status icon.
	IconHaste                           // IconHaste is the haste status icon.
	IconSlow                            // IconSlow is the slowed status icon.
	IconWeak                            // IconWeak is the weakened status icon.
	IconBerserk                         // IconBerserk is the berserk status icon.
	IconInvisible                       // IconInvisible is the invisible status icon.
)

// StatusIcon represents a single status effect icon to display.
type StatusIcon struct {
	Type     StatusIconType
	Duration float64
	Stacks   int
	Color    color.RGBA
}

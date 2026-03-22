package threat

import "image/color"

// ThreatLevel indicates the intensity of a threat for visual prominence.
type ThreatLevel int

const (
	// ThreatNone indicates no active threat.
	ThreatNone ThreatLevel = iota
	// ThreatLow is for distant or passive threats.
	ThreatLow
	// ThreatMedium is for nearby hostile entities.
	ThreatMedium
	// ThreatHigh is for actively attacking enemies.
	ThreatHigh
	// ThreatCritical is for immediate danger (e.g., boss attacks).
	ThreatCritical
)

// Component holds threat indicator state for an entity.
type Component struct {
	// ThreatLevel determines visual prominence (size, brightness, animation speed).
	ThreatLevel ThreatLevel

	// ThreatDecay is time remaining until threat level decreases (seconds).
	ThreatDecay float64

	// LastDamageTime is when this entity last dealt damage to the player.
	LastDamageTime float64

	// AttackWindup indicates an attack is being prepared (for telegraphing).
	AttackWindup bool

	// WindupProgress is 0-1 representing telegraph fill.
	WindupProgress float64

	// PulsePhase is the current animation phase for threat pulse (0-2π).
	PulsePhase float64

	// BorderAlpha is the current alpha for the threat border (0-1).
	BorderAlpha float64

	// IsBoss indicates this is a boss entity (always shows threat at minimum level).
	IsBoss bool
}

// NewComponent creates a default threat component.
func NewComponent() *Component {
	return &Component{
		ThreatLevel: ThreatNone,
		ThreatDecay: 0,
		PulsePhase:  0,
		BorderAlpha: 0,
	}
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "threat"
}

// OffscreenIndicator represents a directional threat arrow for off-screen enemies.
type OffscreenIndicator struct {
	// Angle is the direction toward the threat (radians, 0 = right).
	Angle float64

	// Distance is how far the threat is (used for size scaling).
	Distance float64

	// ThreatLevel for visual intensity.
	ThreatLevel ThreatLevel

	// Alpha is current opacity (for fade in/out).
	Alpha float64

	// Color is the indicator color.
	Color color.RGBA

	// PulsePhase for animation.
	PulsePhase float64
}

// GenreStyle holds genre-specific visual parameters.
type GenreStyle struct {
	// PrimaryColor is the main threat indicator color.
	PrimaryColor color.RGBA

	// SecondaryColor is used for gradients and secondary elements.
	SecondaryColor color.RGBA

	// PulseSpeed is how fast the threat pulses (radians per second).
	PulseSpeed float64

	// BorderThickness is the width of the threat border.
	BorderThickness float32

	// GlowRadius is the size of the glow effect.
	GlowRadius float32

	// ArrowSize is the base size of off-screen indicators.
	ArrowSize float32

	// EdgePadding is distance from screen edge for off-screen indicators.
	EdgePadding float32
}

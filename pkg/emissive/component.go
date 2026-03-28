package emissive

import "image/color"

// GlowType defines the visual style of emission.
type GlowType int

const (
	// TypeFlame is a flickering fire/torch glow.
	TypeFlame GlowType = iota
	// TypeMagic is a pulsing magical glow.
	TypeMagic
	// TypeProjectile is a fast-moving projectile trail.
	TypeProjectile
	// TypeEye is creature eye glow.
	TypeEye
	// TypeNeon is a steady neon/tech glow.
	TypeNeon
	// TypeRadioactive is a pulsing toxic glow.
	TypeRadioactive
	// TypeElectric is crackling electric glow.
	TypeElectric
)

// Component marks an entity as emitting visible glow.
type Component struct {
	// Intensity controls overall brightness [0.0-2.0]. 1.0 = normal.
	Intensity float64

	// Color is the glow color (RGB). Alpha is ignored.
	Color color.RGBA

	// Radius is the base glow radius in screen pixels.
	Radius float64

	// GlowType determines glow behavior (flicker, pulse, steady).
	GlowType GlowType

	// CoreBrightness is the center brightness multiplier [0.0-1.0].
	CoreBrightness float64

	// FalloffPower controls edge softness. Higher = sharper edges.
	FalloffPower float64

	// PulseSpeed is cycles per second for pulsing glow types.
	PulseSpeed float64

	// PulsePhase is the current pulse phase [0.0-2π].
	PulsePhase float64

	// Enabled controls whether glow is rendered.
	Enabled bool

	// ScreenX, ScreenY are cached screen-space positions.
	ScreenX float64
	ScreenY float64

	// Distance from camera (for LOD and size scaling).
	Distance float64
}

// Type implements the ECS Component interface.
func (c *Component) Type() string {
	return "emissive.Component"
}

// NewComponent creates a glow component with default values.
func NewComponent(glowType GlowType, glowColor color.RGBA) *Component {
	return &Component{
		Intensity:      1.0,
		Color:          glowColor,
		Radius:         16.0,
		GlowType:       glowType,
		CoreBrightness: 0.9,
		FalloffPower:   2.0,
		PulseSpeed:     1.0,
		PulsePhase:     0.0,
		Enabled:        true,
	}
}

// NewFlameGlow creates a glow for fire/torch sources.
func NewFlameGlow() *Component {
	return &Component{
		Intensity:      1.0,
		Color:          color.RGBA{R: 255, G: 180, B: 80, A: 255},
		Radius:         20.0,
		GlowType:       TypeFlame,
		CoreBrightness: 0.95,
		FalloffPower:   1.8,
		PulseSpeed:     3.0,
		Enabled:        true,
	}
}

// NewMagicGlow creates a glow for magical effects.
func NewMagicGlow(magicColor color.RGBA) *Component {
	return &Component{
		Intensity:      1.2,
		Color:          magicColor,
		Radius:         24.0,
		GlowType:       TypeMagic,
		CoreBrightness: 0.85,
		FalloffPower:   1.5,
		PulseSpeed:     2.0,
		Enabled:        true,
	}
}

// NewProjectileGlow creates a glow for projectiles.
func NewProjectileGlow(projectileColor color.RGBA) *Component {
	return &Component{
		Intensity:      0.8,
		Color:          projectileColor,
		Radius:         12.0,
		GlowType:       TypeProjectile,
		CoreBrightness: 1.0,
		FalloffPower:   2.5,
		PulseSpeed:     0.0,
		Enabled:        true,
	}
}

// NewEyeGlow creates a glow for creature eyes.
func NewEyeGlow(eyeColor color.RGBA) *Component {
	return &Component{
		Intensity:      0.7,
		Color:          eyeColor,
		Radius:         8.0,
		GlowType:       TypeEye,
		CoreBrightness: 1.0,
		FalloffPower:   3.0,
		PulseSpeed:     0.5,
		Enabled:        true,
	}
}

// NewNeonGlow creates a steady neon/tech glow.
func NewNeonGlow(neonColor color.RGBA) *Component {
	return &Component{
		Intensity:      1.1,
		Color:          neonColor,
		Radius:         14.0,
		GlowType:       TypeNeon,
		CoreBrightness: 0.9,
		FalloffPower:   2.2,
		PulseSpeed:     0.0,
		Enabled:        true,
	}
}

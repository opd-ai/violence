// Package specsparkle provides animated specular sparkle effects for metallic,
// crystalline, and wet surfaces. Sparkles are small, bright highlights that
// appear and fade on reflective surfaces, simulating the glinting effect of
// light catching surface imperfections.
package specsparkle

import "image/color"

// MaterialClass identifies the type of reflective material.
type MaterialClass int

const (
	// MaterialMetal represents polished or brushed metal surfaces.
	MaterialMetal MaterialClass = iota
	// MaterialCrystal represents crystalline or gemstone surfaces.
	MaterialCrystal
	// MaterialWet represents water, puddles, or wet surfaces.
	MaterialWet
	// MaterialGlass represents glass or transparent surfaces.
	MaterialGlass
	// MaterialGold represents gold or brass surfaces (warm tint).
	MaterialGold
	// MaterialSilver represents silver or chrome surfaces (cool tint).
	MaterialSilver
)

// Component marks an entity or screen region as having sparkle effects.
type Component struct {
	// Material determines sparkle color and behavior.
	Material MaterialClass

	// Intensity scales sparkle brightness [0.0-1.0].
	Intensity float64

	// Density controls sparkle spawn frequency [0.0-1.0].
	Density float64

	// Size is the base sparkle size in pixels.
	Size float64

	// Enabled toggles sparkle rendering.
	Enabled bool

	// ScreenX, ScreenY is the screen position for world sparkles.
	ScreenX, ScreenY float64

	// Width, Height defines the sparkle spawn region (for surfaces).
	Width, Height float64

	// Distance from camera (for LOD and intensity falloff).
	Distance float64
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "specsparkle.Component"
}

// Sparkle represents a single animated sparkle point.
type Sparkle struct {
	// X, Y position within the spawn region [0.0-1.0].
	X, Y float64

	// Phase is the animation phase [0.0-1.0] where 0.5 is peak brightness.
	Phase float64

	// Speed controls animation rate.
	Speed float64

	// Size multiplier for this sparkle.
	SizeMult float64

	// Color tint for this sparkle.
	Color color.RGBA

	// Active indicates if sparkle is currently visible.
	Active bool
}

// NewComponent creates a sparkle component with sensible defaults.
func NewComponent(material MaterialClass) *Component {
	return &Component{
		Material:  material,
		Intensity: 0.8,
		Density:   0.5,
		Size:      3.0,
		Enabled:   true,
		Width:     32.0,
		Height:    32.0,
	}
}

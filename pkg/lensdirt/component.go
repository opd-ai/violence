package lensdirt

import (
	"image/color"
)

// LightSource represents a bright light that can trigger lens dirt visibility.
type LightSource struct {
	// ScreenX, ScreenY are screen-space coordinates of the light.
	ScreenX, ScreenY float64

	// Intensity controls how strongly this light triggers dirt visibility (0-1).
	Intensity float64

	// Color is the tint of this light source.
	Color color.RGBA

	// Radius is the effective radius of the light's influence on dirt.
	Radius float64
}

// NewLightSource creates a light source that triggers lens dirt.
func NewLightSource(screenX, screenY, intensity, radius float64, col color.RGBA) LightSource {
	return LightSource{
		ScreenX:   screenX,
		ScreenY:   screenY,
		Intensity: clamp01(intensity),
		Color:     col,
		Radius:    radius,
	}
}

// DirtSpeck represents a single dust/smudge element on the lens.
type DirtSpeck struct {
	// Position on lens (0-1 normalized coordinates).
	X, Y float64

	// Size of the speck (pixels).
	Size float64

	// Opacity at full illumination.
	BaseOpacity float64

	// Shape type for varied appearance.
	Shape SpeckShape

	// Rotation angle in radians.
	Rotation float64

	// Color tint for this speck.
	Tint color.RGBA
}

// SpeckShape defines the visual appearance of a dirt speck.
type SpeckShape int

const (
	// ShapeCircle is a simple round dust particle.
	ShapeCircle SpeckShape = iota
	// ShapeSmudge is an elongated fingerprint-like smear.
	ShapeSmudge
	// ShapeStreaky is a rain-drop streak pattern.
	ShapeStreaky
	// ShapeDiffuse is a soft, diffused glow pattern.
	ShapeDiffuse
	// ShapeHexagonal is a lens-flare hexagonal artifact.
	ShapeHexagonal
)

// DirtPattern stores the procedurally generated lens dirt layout.
type DirtPattern struct {
	// Specks are individual dirt elements.
	Specks []DirtSpeck

	// Width, Height are the pattern dimensions.
	Width, Height int

	// Seed used to generate this pattern.
	Seed int64
}

// Type returns the component type string.
func (d *DirtPattern) Type() string {
	return "DirtPattern"
}

// Type returns the component type string.
func (l *LightSource) Type() string {
	return "LightSource"
}

// clamp01 restricts a value to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

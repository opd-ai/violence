package caustics

import (
	"image/color"
)

// Component marks a location as having caustic light patterns.
// Attached to floor/wall entities near water sources.
type Component struct {
	// WorldX and WorldY are the world position of this caustic source.
	WorldX, WorldY float64

	// Intensity is the caustic brightness (0-1).
	Intensity float64

	// Radius is the effect radius in world units.
	Radius float64

	// Phase is the animation phase offset for variation.
	Phase float64

	// Color is the caustic tint color.
	Color color.RGBA

	// SourceType identifies what type of water caused this caustic.
	SourceType SourceType

	// Seed for deterministic pattern generation.
	Seed int64
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "caustics"
}

// SourceType identifies the type of water source generating caustics.
type SourceType int

const (
	// SourcePuddle is a small puddle creating tight caustics.
	SourcePuddle SourceType = iota

	// SourcePool is a larger water body with broader caustics.
	SourcePool

	// SourceStream is flowing water with directional caustics.
	SourceStream

	// SourceDrip is a dripping water source with small sporadic caustics.
	SourceDrip
)

// CausticPattern stores a precomputed caustic animation frame.
type CausticPattern struct {
	// Width and Height of the pattern in pixels.
	Width, Height int

	// Pixels contains the caustic intensity values (0-255).
	Pixels []byte

	// Frame is the animation frame index.
	Frame int
}

// NewComponent creates a caustic component for a water source.
func NewComponent(worldX, worldY, intensity, radius, phase float64, sourceType SourceType, col color.RGBA, seed int64) *Component {
	return &Component{
		WorldX:     worldX,
		WorldY:     worldY,
		Intensity:  clampFloat(intensity, 0, 1),
		Radius:     radius,
		Phase:      phase,
		Color:      col,
		SourceType: sourceType,
		Seed:       seed,
	}
}

// clampFloat clamps a float64 to a range.
func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

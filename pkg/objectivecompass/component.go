package objectivecompass

import "image/color"

// ObjectiveType categorizes objectives for styling and priority.
type ObjectiveType int

const (
	// TypeMain is a primary/required objective.
	TypeMain ObjectiveType = iota
	// TypeBonus is an optional/bonus objective.
	TypeBonus
	// TypePOI is a point of interest marker.
	TypePOI
	// TypeExit is a level exit marker.
	TypeExit
)

// Component stores compass indicator state for an objective.
type Component struct {
	// ID uniquely identifies this objective.
	ID string

	// ObjType determines styling and priority.
	ObjType ObjectiveType

	// WorldX, WorldY are the objective's world coordinates.
	WorldX float64
	WorldY float64

	// Completed marks whether this objective is done.
	Completed bool

	// ScreenAngle is the angle from screen center to the objective (radians).
	ScreenAngle float64

	// Distance from player to objective in world units.
	Distance float64

	// OnScreen indicates if the objective is currently visible on screen.
	OnScreen bool

	// Alpha is the current opacity (0-1).
	Alpha float64

	// Scale is the current size multiplier.
	Scale float64

	// PulsePhase tracks animation state.
	PulsePhase float64

	// EdgeX, EdgeY are the computed screen-edge position.
	EdgeX float32
	EdgeY float32
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "objectivecompass.Component"
}

// NewComponent creates a compass component with default values.
func NewComponent(id string, objType ObjectiveType, x, y float64) *Component {
	return &Component{
		ID:      id,
		ObjType: objType,
		WorldX:  x,
		WorldY:  y,
		Alpha:   1.0,
		Scale:   1.0,
	}
}

// GenreStyle defines genre-specific visual parameters.
type GenreStyle struct {
	// MainColor is the primary objective indicator color.
	MainColor color.RGBA
	// BonusColor is the bonus objective indicator color.
	BonusColor color.RGBA
	// POIColor is the point of interest indicator color.
	POIColor color.RGBA
	// ExitColor is the exit marker color.
	ExitColor color.RGBA
	// PulseSpeed controls animation rate.
	PulseSpeed float64
	// ArrowSize is the base indicator size.
	ArrowSize float32
	// EdgePadding is the distance from screen edge.
	EdgePadding float32
	// MinAlpha is the minimum opacity for distant objectives.
	MinAlpha float64
	// MaxDistance is the distance at which indicators reach MinAlpha.
	MaxDistance float64
	// GlowIntensity controls the glow effect strength.
	GlowIntensity float64
}

// DefaultGenreStyles returns genre presets.
func DefaultGenreStyles() map[string]GenreStyle {
	return map[string]GenreStyle{
		"fantasy": {
			MainColor:     color.RGBA{R: 255, G: 215, B: 0, A: 255},   // Gold
			BonusColor:    color.RGBA{R: 200, G: 180, B: 255, A: 220}, // Lavender
			POIColor:      color.RGBA{R: 100, G: 255, B: 150, A: 200}, // Mint
			ExitColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White
			PulseSpeed:    4.0,
			ArrowSize:     10.0,
			EdgePadding:   6.0,
			MinAlpha:      0.25,
			MaxDistance:   50.0,
			GlowIntensity: 0.4,
		},
		"scifi": {
			MainColor:     color.RGBA{R: 0, G: 255, B: 200, A: 255}, // Cyan
			BonusColor:    color.RGBA{R: 100, G: 200, B: 255, A: 220},
			POIColor:      color.RGBA{R: 255, G: 200, B: 100, A: 200},
			ExitColor:     color.RGBA{R: 0, G: 255, B: 100, A: 255}, // Bright green
			PulseSpeed:    6.0,
			ArrowSize:     8.0,
			EdgePadding:   8.0,
			MinAlpha:      0.3,
			MaxDistance:   60.0,
			GlowIntensity: 0.5,
		},
		"horror": {
			MainColor:     color.RGBA{R: 255, G: 100, B: 100, A: 255}, // Dull red
			BonusColor:    color.RGBA{R: 150, G: 100, B: 150, A: 200}, // Muted purple
			POIColor:      color.RGBA{R: 100, G: 100, B: 100, A: 180}, // Gray
			ExitColor:     color.RGBA{R: 200, G: 200, B: 200, A: 255}, // Dim white
			PulseSpeed:    2.5,
			ArrowSize:     10.0,
			EdgePadding:   5.0,
			MinAlpha:      0.15,
			MaxDistance:   40.0,
			GlowIntensity: 0.2,
		},
		"cyberpunk": {
			MainColor:     color.RGBA{R: 255, G: 0, B: 180, A: 255}, // Magenta
			BonusColor:    color.RGBA{R: 0, G: 255, B: 255, A: 220}, // Cyan
			POIColor:      color.RGBA{R: 255, G: 255, B: 0, A: 200}, // Yellow
			ExitColor:     color.RGBA{R: 0, G: 255, B: 150, A: 255}, // Neon green
			PulseSpeed:    8.0,
			ArrowSize:     9.0,
			EdgePadding:   7.0,
			MinAlpha:      0.35,
			MaxDistance:   55.0,
			GlowIntensity: 0.7,
		},
		"postapoc": {
			MainColor:     color.RGBA{R: 255, G: 150, B: 50, A: 255},  // Amber
			BonusColor:    color.RGBA{R: 180, G: 150, B: 100, A: 200}, // Tan
			POIColor:      color.RGBA{R: 150, G: 200, B: 100, A: 180}, // Olive
			ExitColor:     color.RGBA{R: 200, G: 180, B: 150, A: 255}, // Dusty white
			PulseSpeed:    3.5,
			ArrowSize:     11.0,
			EdgePadding:   6.0,
			MinAlpha:      0.2,
			MaxDistance:   45.0,
			GlowIntensity: 0.3,
		},
	}
}

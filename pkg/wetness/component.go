// Package wetness provides surface wetness rendering for environmental realism.
package wetness

// Component stores wetness state for a surface location.
type Component struct {
	// Grid position
	X, Y int

	// Moisture level (0.0 = dry, 1.0 = standing water)
	Moisture float64

	// Depth for puddles (0.0 = damp, 1.0 = deep puddle)
	Depth float64

	// Specular intensity multiplier for this wet surface
	SpecularIntensity float64

	// Tint color components (0.0-1.0) for contaminated/colored water
	TintR, TintG, TintB float64

	// Whether this is a puddle (vs just damp surface)
	IsPuddle bool

	// Seed for deterministic noise patterns
	Seed int64
}

// Type implements the engine.Component interface.
func (c *Component) Type() string {
	return "wetness.Component"
}

// NewComponent creates a wetness component with default values.
func NewComponent(x, y int, moisture float64, seed int64) *Component {
	isPuddle := moisture > 0.6
	depth := 0.0
	if isPuddle {
		depth = (moisture - 0.6) / 0.4 // Normalize to 0-1 for puddles
	}

	return &Component{
		X:                 x,
		Y:                 y,
		Moisture:          moisture,
		Depth:             depth,
		SpecularIntensity: 0.3 + moisture*0.7, // More wet = more specular
		TintR:             1.0,
		TintG:             1.0,
		TintB:             1.0,
		IsPuddle:          isPuddle,
		Seed:              seed,
	}
}

// WetnessPattern stores the wetness data for an entire level.
type WetnessPattern struct {
	Width, Height int
	Cells         [][]*Component // [y][x]
	GenreID       string
	Seed          int64
}

// GetMoistureAt returns the moisture level at a position, 0 if out of bounds.
func (p *WetnessPattern) GetMoistureAt(x, y int) float64 {
	if x < 0 || y < 0 || y >= len(p.Cells) || x >= len(p.Cells[0]) {
		return 0
	}
	if p.Cells[y][x] == nil {
		return 0
	}
	return p.Cells[y][x].Moisture
}

// GetComponentAt returns the wetness component at a position, nil if none.
func (p *WetnessPattern) GetComponentAt(x, y int) *Component {
	if x < 0 || y < 0 || y >= len(p.Cells) || x >= len(p.Cells[0]) {
		return nil
	}
	return p.Cells[y][x]
}

// LightSource represents a light for specular calculations.
type LightSource struct {
	X, Y          float64
	Radius        float64
	Intensity     float64
	R, G, B       float64 // Light color (0.0-1.0)
	IsWarm        bool    // Warm vs cool temperature hint
	FlickerPhase  float64 // For animated torch flicker
	FlickerAmount float64 // How much the light flickers (0-1)
}

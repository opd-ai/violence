package parallax

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Component stores parallax layer data for an entity or world.
type Component struct {
	Layers     []*Layer
	GenreID    string
	BiomeID    string
	Seed       int64
	Enabled    bool
	CameraX    float64
	CameraY    float64
	ViewWidth  int
	ViewHeight int
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "parallax"
}

// Layer represents a single parallax scrolling layer.
type Layer struct {
	Image       *ebiten.Image
	ScrollSpeed float64 // 0.0 = static, 1.0 = moves with camera, >1.0 = moves faster
	OffsetX     float64
	OffsetY     float64
	RepeatX     bool
	RepeatY     bool
	Opacity     float64
	ZIndex      int // Lower values render first (farther back)
	Width       int
	Height      int
	Tint        [4]float64 // RGBA multiplier for atmospheric effects
}

// NewComponent creates a parallax component with default settings.
func NewComponent(genreID, biomeID string, seed int64) *Component {
	return &Component{
		Layers:  make([]*Layer, 0, 4),
		GenreID: genreID,
		BiomeID: biomeID,
		Seed:    seed,
		Enabled: true,
	}
}

// AddLayer adds a parallax layer to the component.
func (c *Component) AddLayer(layer *Layer) {
	c.Layers = append(c.Layers, layer)
}

// UpdateCamera updates the camera position for parallax calculations.
func (c *Component) UpdateCamera(x, y float64, viewWidth, viewHeight int) {
	c.CameraX = x
	c.CameraY = y
	c.ViewWidth = viewWidth
	c.ViewHeight = viewHeight
}

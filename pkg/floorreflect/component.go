package floorreflect

import (
	"image"
	"image/color"
)

// Component stores reflection data for an entity.
type Component struct {
	// Source sprite image to reflect
	SourceImage *image.RGBA

	// World position of the entity
	X, Y float64

	// Reflection intensity override (0.0-1.0, -1 means use system default)
	IntensityOverride float64

	// Floor contact Y offset (distance from sprite bottom to floor)
	FloorOffset float64

	// Enable/disable reflection for this entity
	Enabled bool

	// Cached reflected image (for performance)
	cachedReflection *image.RGBA
	cachedSeed       uint64
	cacheValid       bool
}

// Type implements the engine.Component interface.
func (c *Component) Type() string {
	return "floorreflect"
}

// NewComponent creates a floor reflection component with default settings.
func NewComponent() *Component {
	return &Component{
		IntensityOverride: -1.0, // Use system default
		FloorOffset:       0.0,
		Enabled:           true,
		cacheValid:        false,
	}
}

// SetSourceImage updates the source sprite and invalidates the cache.
func (c *Component) SetSourceImage(img *image.RGBA) {
	if c.SourceImage != img {
		c.SourceImage = img
		c.cacheValid = false
	}
}

// SetPosition updates the entity world position.
func (c *Component) SetPosition(x, y float64) {
	c.X = x
	c.Y = y
}

// SetIntensity sets a custom reflection intensity for this entity.
func (c *Component) SetIntensity(intensity float64) {
	if intensity < 0 {
		c.IntensityOverride = -1.0
	} else if intensity > 1.0 {
		c.IntensityOverride = 1.0
	} else {
		c.IntensityOverride = intensity
	}
}

// InvalidateCache marks the cached reflection as stale.
func (c *Component) InvalidateCache() {
	c.cacheValid = false
}

// ReflectionBounds holds the screen-space rectangle for a reflection.
type ReflectionBounds struct {
	X, Y          int
	Width, Height int
	Alpha         float64
}

// MaterialReflectivity defines how reflective different floor materials are.
type MaterialReflectivity struct {
	// Base reflectivity (0.0-1.0)
	Reflectivity float64
	// Color tint applied to reflection
	TintR, TintG, TintB float64
	// Distortion amount for water/liquid surfaces
	Distortion float64
	// Fade distance multiplier (higher = reflection fades faster)
	FadeRate float64
}

// Material reflectivity presets
var (
	// ReflectMetal is highly reflective polished metal
	ReflectMetal = MaterialReflectivity{
		Reflectivity: 0.7,
		TintR:        1.0, TintG: 1.0, TintB: 1.05, // Slight cool tint
		Distortion: 0.0,
		FadeRate:   0.8,
	}

	// ReflectWetStone is wet stone with moderate reflectivity
	ReflectWetStone = MaterialReflectivity{
		Reflectivity: 0.4,
		TintR:        0.9, TintG: 0.95, TintB: 1.0,
		Distortion: 0.1,
		FadeRate:   1.2,
	}

	// ReflectWater is standing water with distortion
	ReflectWater = MaterialReflectivity{
		Reflectivity: 0.55,
		TintR:        0.85, TintG: 0.9, TintB: 1.0,
		Distortion: 0.25,
		FadeRate:   1.0,
	}

	// ReflectOilSlick is oil/chemical puddle
	ReflectOilSlick = MaterialReflectivity{
		Reflectivity: 0.6,
		TintR:        1.1, TintG: 0.9, TintB: 1.2, // Rainbow-ish
		Distortion: 0.15,
		FadeRate:   0.9,
	}

	// ReflectPolishedTile is clean polished tile
	ReflectPolishedTile = MaterialReflectivity{
		Reflectivity: 0.5,
		TintR:        1.0, TintG: 1.0, TintB: 1.0,
		Distortion: 0.0,
		FadeRate:   1.0,
	}

	// ReflectNone is non-reflective surface
	ReflectNone = MaterialReflectivity{
		Reflectivity: 0.0,
		TintR:        1.0, TintG: 1.0, TintB: 1.0,
		Distortion: 0.0,
		FadeRate:   2.0,
	}
)

// FloorTileReflect holds reflection data for a floor tile.
type FloorTileReflect struct {
	// Tile grid position
	TileX, TileY int

	// Material reflectivity properties
	Material MaterialReflectivity

	// Light contribution (from nearby light sources)
	LightLevel float64
}

// ReflectionData holds pre-computed reflection image and metadata.
type ReflectionData struct {
	// Reflected sprite image (vertically flipped, faded)
	Image *image.RGBA

	// Screen-space render position
	ScreenX, ScreenY int

	// Final alpha multiplier
	Alpha float64

	// Source entity bounds for culling
	SourceBounds image.Rectangle
}

// tintColor applies a color tint to a pixel.
func tintColor(c color.RGBA, tintR, tintG, tintB float64) color.RGBA {
	r := float64(c.R) * tintR
	g := float64(c.G) * tintG
	b := float64(c.B) * tintB

	// Clamp values
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: c.A,
	}
}

// lerpColor interpolates between two colors.
func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

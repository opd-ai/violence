// Package muzzleflash provides visual feedback for ranged weapon firing.
// Muzzle flashes spawn at the weapon barrel and provide immediate visual
// confirmation that a shot was fired, improving player feedback.
package muzzleflash

import "image/color"

// Flash represents an active muzzle flash effect.
type Flash struct {
	// Position in world space
	X, Y float64

	// Direction the flash is facing (radians)
	Angle float64

	// Time since spawn
	Age float64

	// Duration of this flash
	Duration float64

	// Type of flash ("bullet", "plasma", "energy", "fire", "magic")
	FlashType string

	// Intensity multiplier (1.0 = normal)
	Intensity float64

	// Primary color of the flash
	PrimaryColor color.RGBA

	// Secondary/glow color
	SecondaryColor color.RGBA

	// Size multiplier
	Scale float64

	// Whether this flash should emit light
	EmitsLight bool

	// Light intensity for dynamic lighting integration
	LightIntensity float64

	// Light radius for dynamic lighting
	LightRadius float64
}

// Component stores active muzzle flashes for an entity.
type Component struct {
	// ActiveFlashes holds all currently rendering flashes
	ActiveFlashes []*Flash

	// MaxFlashes limits concurrent flashes per entity (prevent spam)
	MaxFlashes int
}

// Type implements the engine.Component interface.
func (c *Component) Type() string {
	return "muzzleflash.Component"
}

// NewComponent creates a muzzle flash component with default values.
func NewComponent() *Component {
	return &Component{
		ActiveFlashes: make([]*Flash, 0, 8),
		MaxFlashes:    8,
	}
}

// FlashProfile defines visual parameters for a flash type.
type FlashProfile struct {
	// Duration in seconds
	Duration float64

	// Base size
	BaseSize float64

	// Number of rays/spikes
	RayCount int

	// Whether rays have random variation
	RandomRays bool

	// Primary flash color
	PrimaryColor color.RGBA

	// Glow/secondary color
	SecondaryColor color.RGBA

	// Core brightness multiplier
	CoreBrightness float64

	// Whether to emit light
	EmitsLight bool

	// Light intensity
	LightIntensity float64

	// Light radius
	LightRadius float64
}

// DefaultProfiles provides preset profiles for different weapon types.
var DefaultProfiles = map[string]FlashProfile{
	"bullet": {
		Duration:       0.06,
		BaseSize:       12.0,
		RayCount:       5,
		RandomRays:     true,
		PrimaryColor:   color.RGBA{R: 255, G: 220, B: 150, A: 255},
		SecondaryColor: color.RGBA{R: 255, G: 180, B: 80, A: 200},
		CoreBrightness: 2.0,
		EmitsLight:     true,
		LightIntensity: 1.5,
		LightRadius:    3.0,
	},
	"plasma": {
		Duration:       0.1,
		BaseSize:       16.0,
		RayCount:       0, // No rays, just glow
		RandomRays:     false,
		PrimaryColor:   color.RGBA{R: 100, G: 200, B: 255, A: 255},
		SecondaryColor: color.RGBA{R: 150, G: 230, B: 255, A: 180},
		CoreBrightness: 2.5,
		EmitsLight:     true,
		LightIntensity: 2.0,
		LightRadius:    4.0,
	},
	"energy": {
		Duration:       0.08,
		BaseSize:       14.0,
		RayCount:       4,
		RandomRays:     false,
		PrimaryColor:   color.RGBA{R: 180, G: 255, B: 180, A: 255},
		SecondaryColor: color.RGBA{R: 100, G: 255, B: 100, A: 180},
		CoreBrightness: 2.2,
		EmitsLight:     true,
		LightIntensity: 1.8,
		LightRadius:    3.5,
	},
	"fire": {
		Duration:       0.12,
		BaseSize:       18.0,
		RayCount:       7,
		RandomRays:     true,
		PrimaryColor:   color.RGBA{R: 255, G: 150, B: 50, A: 255},
		SecondaryColor: color.RGBA{R: 255, G: 100, B: 30, A: 200},
		CoreBrightness: 1.8,
		EmitsLight:     true,
		LightIntensity: 2.5,
		LightRadius:    5.0,
	},
	"magic": {
		Duration:       0.15,
		BaseSize:       20.0,
		RayCount:       6,
		RandomRays:     true,
		PrimaryColor:   color.RGBA{R: 200, G: 100, B: 255, A: 255},
		SecondaryColor: color.RGBA{R: 150, G: 50, B: 255, A: 180},
		CoreBrightness: 2.0,
		EmitsLight:     true,
		LightIntensity: 2.0,
		LightRadius:    4.5,
	},
	"shotgun": {
		Duration:       0.08,
		BaseSize:       20.0,
		RayCount:       8,
		RandomRays:     true,
		PrimaryColor:   color.RGBA{R: 255, G: 200, B: 120, A: 255},
		SecondaryColor: color.RGBA{R: 255, G: 160, B: 60, A: 220},
		CoreBrightness: 2.5,
		EmitsLight:     true,
		LightIntensity: 2.5,
		LightRadius:    4.0,
	},
	"laser": {
		Duration:       0.04,
		BaseSize:       10.0,
		RayCount:       0,
		RandomRays:     false,
		PrimaryColor:   color.RGBA{R: 255, G: 50, B: 50, A: 255},
		SecondaryColor: color.RGBA{R: 255, G: 150, B: 150, A: 200},
		CoreBrightness: 3.0,
		EmitsLight:     true,
		LightIntensity: 1.2,
		LightRadius:    2.5,
	},
}

// GetProfile returns a flash profile, defaulting to "bullet" if not found.
func GetProfile(flashType string) FlashProfile {
	if profile, ok := DefaultProfiles[flashType]; ok {
		return profile
	}
	return DefaultProfiles["bullet"]
}

package surfacesheen

import "image/color"

// MaterialType identifies the surface material for sheen calculation.
type MaterialType int

const (
	// MaterialMetal has high specular with colored reflections.
	MaterialMetal MaterialType = iota
	// MaterialWet has broad soft reflections.
	MaterialWet
	// MaterialPolished has sharp mirror-like highlights.
	MaterialPolished
	// MaterialOrganic has subtle subsurface glow.
	MaterialOrganic
	// MaterialCloth has minimal diffuse sheen.
	MaterialCloth
	// MaterialCrystal has prismatic colorful highlights.
	MaterialCrystal
	// MaterialDefault has standard moderate sheen.
	MaterialDefault
)

// String returns the material type name.
func (m MaterialType) String() string {
	switch m {
	case MaterialMetal:
		return "metal"
	case MaterialWet:
		return "wet"
	case MaterialPolished:
		return "polished"
	case MaterialOrganic:
		return "organic"
	case MaterialCloth:
		return "cloth"
	case MaterialCrystal:
		return "crystal"
	default:
		return "default"
	}
}

// Type returns the component type identifier for ECS.
func (m MaterialType) Type() string {
	return "surfacesheen.MaterialType"
}

// SheenComponent marks an entity as having surface sheen properties.
// Attach this component to entities that should exhibit material reflections.
type SheenComponent struct {
	// Material determines the reflection behavior.
	Material MaterialType

	// Intensity multiplier for the sheen effect [0.0-2.0].
	Intensity float64

	// BaseColor is the entity's primary surface color (used for tinting).
	BaseColor color.RGBA

	// Roughness affects highlight spread [0.0-1.0], lower = sharper.
	Roughness float64

	// Wetness adds additional wet surface reflection [0.0-1.0].
	Wetness float64
}

// Type returns the component type identifier for ECS.
func (c *SheenComponent) Type() string {
	return "surfacesheen.SheenComponent"
}

// NewSheenComponent creates a sheen component with sensible defaults.
func NewSheenComponent(material MaterialType, baseColor color.RGBA) *SheenComponent {
	return &SheenComponent{
		Material:  material,
		Intensity: 1.0,
		BaseColor: baseColor,
		Roughness: defaultRoughnessFor(material),
		Wetness:   0.0,
	}
}

// defaultRoughnessFor returns the typical roughness for a material type.
func defaultRoughnessFor(material MaterialType) float64 {
	switch material {
	case MaterialMetal:
		return 0.25
	case MaterialWet:
		return 0.15
	case MaterialPolished:
		return 0.05
	case MaterialOrganic:
		return 0.7
	case MaterialCloth:
		return 0.9
	case MaterialCrystal:
		return 0.1
	default:
		return 0.5
	}
}

// LightSource represents a light that can create sheen reflections.
type LightSource struct {
	// X, Y position in world space.
	X, Y float64

	// Color of the light.
	Color color.RGBA

	// Intensity of the light [0.0-2.0].
	Intensity float64

	// Radius of light influence (for falloff calculation).
	Radius float64
}

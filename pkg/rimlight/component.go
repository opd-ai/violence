package rimlight

import "image/color"

// Material types for rim light intensity calculation.
type Material int

const (
	// MaterialDefault provides standard rim lighting.
	MaterialDefault Material = iota
	// MaterialMetal has high reflectivity and strong rim.
	MaterialMetal
	// MaterialCloth has soft, diffuse rim lighting.
	MaterialCloth
	// MaterialLeather has moderate rim with slight sheen.
	MaterialLeather
	// MaterialOrganic has subsurface-style soft rim.
	MaterialOrganic
	// MaterialCrystal has very bright specular rim.
	MaterialCrystal
	// MaterialMagic has colored glow rim effect.
	MaterialMagic
)

// Component stores rim lighting configuration for an entity.
type Component struct {
	// Enabled controls whether rim lighting is applied.
	Enabled bool
	// Material determines rim intensity and behavior.
	Material Material
	// Intensity overrides material default (0.0-2.0, 1.0 = default).
	Intensity float64
	// Color overrides the rim light color (zero value = use genre default).
	Color color.RGBA
	// Width controls the rim edge width in pixels (0 = auto based on sprite size).
	Width int
	// FadeInner controls how quickly rim fades toward sprite interior (0.0-1.0).
	FadeInner float64
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "RimLightComponent"
}

// NewComponent creates a rim light component with default settings.
func NewComponent() *Component {
	return &Component{
		Enabled:   true,
		Material:  MaterialDefault,
		Intensity: 1.0,
		Width:     0, // Auto
		FadeInner: 0.5,
	}
}

// NewComponentWithMaterial creates a rim light component for a specific material.
func NewComponentWithMaterial(mat Material) *Component {
	comp := NewComponent()
	comp.Material = mat
	return comp
}

// GetMaterialIntensity returns the base rim intensity for a material type.
func GetMaterialIntensity(mat Material) float64 {
	switch mat {
	case MaterialMetal:
		return 1.4
	case MaterialCrystal:
		return 1.8
	case MaterialMagic:
		return 1.6
	case MaterialLeather:
		return 0.9
	case MaterialCloth:
		return 0.6
	case MaterialOrganic:
		return 0.7
	default:
		return 1.0
	}
}

// GetMaterialFresnel returns the Fresnel exponent for a material type.
// Higher values create tighter edge highlights.
func GetMaterialFresnel(mat Material) float64 {
	switch mat {
	case MaterialMetal:
		return 3.0
	case MaterialCrystal:
		return 4.0
	case MaterialMagic:
		return 2.0
	case MaterialCloth:
		return 1.5
	case MaterialOrganic:
		return 2.0
	default:
		return 2.5
	}
}

package subsurface

import "image/color"

// Material represents the organic material type for SSS calculation.
type Material int

const (
	// MaterialFlesh represents skin/muscle with strong red scattering.
	MaterialFlesh Material = iota
	// MaterialLeaf represents plant matter with green-biased scattering.
	MaterialLeaf
	// MaterialWax represents waxy surfaces with neutral warm scattering.
	MaterialWax
	// MaterialSlime represents translucent gel with color preservation.
	MaterialSlime
	// MaterialMembrane represents thin translucent tissue (wings, fins).
	MaterialMembrane
	// MaterialFruit represents fruit/vegetable with moderate scattering.
	MaterialFruit
	// MaterialBone represents bone/ivory with subtle yellow-white scattering.
	MaterialBone
)

// ScatterProfile defines material-specific scattering properties.
type ScatterProfile struct {
	// ScatterColor is the color bias introduced by scattering.
	// Light traveling through flesh shifts toward red; leaves toward green.
	ScatterColor color.RGBA
	// ScatterDistance is the mean free path in pixels before scattering.
	// Lower values = more opaque, higher = more translucent.
	ScatterDistance float64
	// Absorption rates for RGB (how quickly each component is absorbed).
	// Blue is typically absorbed fastest in flesh (high AbsorptionB).
	AbsorptionR, AbsorptionG, AbsorptionB float64
	// Translucency controls edge light transmission (0.0-1.0).
	Translucency float64
	// BackscatterStrength controls light bouncing back toward the viewer.
	BackscatterStrength float64
}

// Component stores subsurface scattering configuration for an entity.
type Component struct {
	// Enabled controls whether SSS is applied.
	Enabled bool
	// Material determines the scattering profile.
	Material Material
	// Intensity multiplier for SSS effect (0.0-2.0, 1.0 = default).
	Intensity float64
	// ThicknessOverride forces a specific thickness (0 = auto-compute from sprite).
	ThicknessOverride float64
	// ColorOverride replaces the material's scatter color if non-zero.
	ColorOverride color.RGBA
	// LightPenetration controls how far light penetrates (0.0-1.0).
	LightPenetration float64
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "SubsurfaceComponent"
}

// NewComponent creates a subsurface component with default settings.
func NewComponent() *Component {
	return &Component{
		Enabled:          true,
		Material:         MaterialFlesh,
		Intensity:        1.0,
		LightPenetration: 0.5,
	}
}

// NewComponentWithMaterial creates a subsurface component for a specific material.
func NewComponentWithMaterial(mat Material) *Component {
	comp := NewComponent()
	comp.Material = mat
	return comp
}

// GetScatterProfile returns the scattering properties for a material type.
func GetScatterProfile(mat Material) ScatterProfile {
	switch mat {
	case MaterialFlesh:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 255, G: 180, B: 160, A: 255},
			ScatterDistance:     8.0,
			AbsorptionR:         0.1,
			AbsorptionG:         0.5,
			AbsorptionB:         0.8,
			Translucency:        0.4,
			BackscatterStrength: 0.3,
		}
	case MaterialLeaf:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 140, G: 255, B: 140, A: 255},
			ScatterDistance:     6.0,
			AbsorptionR:         0.6,
			AbsorptionG:         0.1,
			AbsorptionB:         0.7,
			Translucency:        0.6,
			BackscatterStrength: 0.2,
		}
	case MaterialWax:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 255, G: 240, B: 210, A: 255},
			ScatterDistance:     10.0,
			AbsorptionR:         0.2,
			AbsorptionG:         0.3,
			AbsorptionB:         0.4,
			Translucency:        0.5,
			BackscatterStrength: 0.4,
		}
	case MaterialSlime:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 200, G: 255, B: 200, A: 255},
			ScatterDistance:     15.0,
			AbsorptionR:         0.3,
			AbsorptionG:         0.2,
			AbsorptionB:         0.3,
			Translucency:        0.8,
			BackscatterStrength: 0.5,
		}
	case MaterialMembrane:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 255, G: 220, B: 200, A: 255},
			ScatterDistance:     4.0,
			AbsorptionR:         0.2,
			AbsorptionG:         0.4,
			AbsorptionB:         0.5,
			Translucency:        0.7,
			BackscatterStrength: 0.2,
		}
	case MaterialFruit:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 255, G: 200, B: 150, A: 255},
			ScatterDistance:     7.0,
			AbsorptionR:         0.15,
			AbsorptionG:         0.4,
			AbsorptionB:         0.6,
			Translucency:        0.5,
			BackscatterStrength: 0.35,
		}
	case MaterialBone:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 255, G: 250, B: 230, A: 255},
			ScatterDistance:     5.0,
			AbsorptionR:         0.3,
			AbsorptionG:         0.35,
			AbsorptionB:         0.4,
			Translucency:        0.2,
			BackscatterStrength: 0.25,
		}
	default:
		return ScatterProfile{
			ScatterColor:        color.RGBA{R: 255, G: 200, B: 180, A: 255},
			ScatterDistance:     8.0,
			AbsorptionR:         0.2,
			AbsorptionG:         0.4,
			AbsorptionB:         0.6,
			Translucency:        0.4,
			BackscatterStrength: 0.3,
		}
	}
}

// GetMaterialName returns a human-readable name for the material.
func GetMaterialName(mat Material) string {
	switch mat {
	case MaterialFlesh:
		return "flesh"
	case MaterialLeaf:
		return "leaf"
	case MaterialWax:
		return "wax"
	case MaterialSlime:
		return "slime"
	case MaterialMembrane:
		return "membrane"
	case MaterialFruit:
		return "fruit"
	case MaterialBone:
		return "bone"
	default:
		return "unknown"
	}
}

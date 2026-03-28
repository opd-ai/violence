package surfacesheen

import (
	"image/color"
	"testing"
)

func TestMaterialTypeString(t *testing.T) {
	tests := []struct {
		material MaterialType
		expected string
	}{
		{MaterialMetal, "metal"},
		{MaterialWet, "wet"},
		{MaterialPolished, "polished"},
		{MaterialOrganic, "organic"},
		{MaterialCloth, "cloth"},
		{MaterialCrystal, "crystal"},
		{MaterialDefault, "default"},
		{MaterialType(99), "default"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := tt.material.String()
			if got != tt.expected {
				t.Errorf("MaterialType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestMaterialTypeType(t *testing.T) {
	m := MaterialMetal
	if got := m.Type(); got != "surfacesheen.MaterialType" {
		t.Errorf("MaterialType.Type() = %q, want %q", got, "surfacesheen.MaterialType")
	}
}

func TestSheenComponentType(t *testing.T) {
	c := &SheenComponent{}
	if got := c.Type(); got != "surfacesheen.SheenComponent" {
		t.Errorf("SheenComponent.Type() = %q, want %q", got, "surfacesheen.SheenComponent")
	}
}

func TestNewSheenComponent(t *testing.T) {
	tests := []struct {
		name              string
		material          MaterialType
		baseColor         color.RGBA
		expectedRoughness float64
	}{
		{
			name:              "metal component",
			material:          MaterialMetal,
			baseColor:         color.RGBA{R: 180, G: 180, B: 200, A: 255},
			expectedRoughness: 0.25,
		},
		{
			name:              "wet component",
			material:          MaterialWet,
			baseColor:         color.RGBA{R: 100, G: 100, B: 150, A: 255},
			expectedRoughness: 0.15,
		},
		{
			name:              "polished component",
			material:          MaterialPolished,
			baseColor:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
			expectedRoughness: 0.05,
		},
		{
			name:              "organic component",
			material:          MaterialOrganic,
			baseColor:         color.RGBA{R: 200, G: 150, B: 100, A: 255},
			expectedRoughness: 0.7,
		},
		{
			name:              "cloth component",
			material:          MaterialCloth,
			baseColor:         color.RGBA{R: 150, G: 50, B: 50, A: 255},
			expectedRoughness: 0.9,
		},
		{
			name:              "crystal component",
			material:          MaterialCrystal,
			baseColor:         color.RGBA{R: 200, G: 220, B: 255, A: 255},
			expectedRoughness: 0.1,
		},
		{
			name:              "default component",
			material:          MaterialDefault,
			baseColor:         color.RGBA{R: 128, G: 128, B: 128, A: 255},
			expectedRoughness: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewSheenComponent(tt.material, tt.baseColor)

			if comp.Material != tt.material {
				t.Errorf("Material = %v, want %v", comp.Material, tt.material)
			}
			if comp.BaseColor != tt.baseColor {
				t.Errorf("BaseColor = %v, want %v", comp.BaseColor, tt.baseColor)
			}
			if comp.Intensity != 1.0 {
				t.Errorf("Intensity = %v, want 1.0", comp.Intensity)
			}
			if comp.Roughness != tt.expectedRoughness {
				t.Errorf("Roughness = %v, want %v", comp.Roughness, tt.expectedRoughness)
			}
			if comp.Wetness != 0.0 {
				t.Errorf("Wetness = %v, want 0.0", comp.Wetness)
			}
		})
	}
}

func TestDefaultRoughnessFor(t *testing.T) {
	// Test that unknown material types return default roughness
	unknown := MaterialType(99)
	roughness := defaultRoughnessFor(unknown)
	if roughness != 0.5 {
		t.Errorf("defaultRoughnessFor(unknown) = %v, want 0.5", roughness)
	}
}

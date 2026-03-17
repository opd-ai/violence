package rimlight_test

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/rimlight"
)

// ExampleNewSystem demonstrates creating a rim lighting system for a genre.
func ExampleNewSystem() {
	// Create a rim lighting system for the fantasy genre
	sys := rimlight.NewSystem("fantasy")

	// Get the current light direction
	x, y := sys.GetLightDirection()
	fmt.Printf("Light direction: (%.2f, %.2f)\n", x, y)

	// Get the rim color
	color := sys.GetRimColor()
	fmt.Printf("Rim color: R=%d, G=%d, B=%d\n", color.R, color.G, color.B)
}

// ExampleNewComponent demonstrates creating rim light components.
func ExampleNewComponent() {
	// Create a default rim light component
	comp := rimlight.NewComponent()
	fmt.Printf("Enabled: %v\n", comp.Enabled)
	fmt.Printf("Intensity: %.1f\n", comp.Intensity)

	// Create a metal material component with strong rim
	metalComp := rimlight.NewComponentWithMaterial(rimlight.MaterialMetal)
	metalIntensity := rimlight.GetMaterialIntensity(metalComp.Material)
	fmt.Printf("Metal rim intensity: %.1f\n", metalIntensity)
}

// ExampleGetMaterialIntensity shows how different materials have different rim intensities.
func ExampleGetMaterialIntensity() {
	materials := []struct {
		name string
		mat  rimlight.Material
	}{
		{"Metal", rimlight.MaterialMetal},
		{"Crystal", rimlight.MaterialCrystal},
		{"Cloth", rimlight.MaterialCloth},
		{"Organic", rimlight.MaterialOrganic},
	}

	for _, m := range materials {
		intensity := rimlight.GetMaterialIntensity(m.mat)
		fmt.Printf("%s: %.1f\n", m.name, intensity)
	}
}

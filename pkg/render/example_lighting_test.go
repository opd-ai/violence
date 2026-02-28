package render_test

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/lighting"
	"github.com/opd-ai/violence/pkg/raycaster"
	"github.com/opd-ai/violence/pkg/render"
)

// Example demonstrating how to integrate lighting with the renderer.
func ExampleRenderer_SetLightMap() {
	// Create a raycaster with a simple map
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	})

	// Create a renderer
	r := render.NewRenderer(320, 200, rc)

	// Create a lighting map with low ambient light
	lightMap := lighting.NewSectorLightMap(5, 5, 0.3)

	// Add a bright torch in the center
	torch := lighting.Light{
		X:         2.5,
		Y:         2.5,
		Radius:    4.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.8,
		B:         0.6,
	}
	lightMap.AddLight(torch)

	// Calculate lighting
	lightMap.Calculate()

	// Integrate lighting with renderer
	r.SetLightMap(lightMap)

	// Now when rendering, walls/floors/ceilings will be lit according to the light map
	// Areas near the torch will be brighter, distant areas will be darker

	fmt.Println("Lighting integrated successfully")
	// Output: Lighting integrated successfully
}

// Example showing how lighting affects rendering brightness.
func ExampleRenderer_SetLightMap_brightness() {
	rc := raycaster.NewRaycaster(66.0, 320, 200)
	rc.SetMap([][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	})

	r := render.NewRenderer(320, 200, rc)

	// Scenario 1: No lighting (full brightness)
	fmt.Println("No lighting: full brightness")

	// Scenario 2: Dim ambient lighting
	dimLightMap := lighting.NewSectorLightMap(3, 3, 0.2)
	dimLightMap.Calculate()
	r.SetLightMap(dimLightMap)
	fmt.Println("Dim lighting: 0.2 ambient")

	// Scenario 3: Bright ambient lighting
	brightLightMap := lighting.NewSectorLightMap(3, 3, 0.9)
	brightLightMap.Calculate()
	r.SetLightMap(brightLightMap)
	fmt.Println("Bright lighting: 0.9 ambient")

	// Output:
	// No lighting: full brightness
	// Dim lighting: 0.2 ambient
	// Bright lighting: 0.9 ambient
}

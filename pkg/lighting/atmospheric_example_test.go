package lighting_test

import (
	"fmt"

	"github.com/opd-ai/violence/pkg/lighting"
)

// This example demonstrates creating an atmospheric lighting system for a fantasy dungeon
// with torch lights, wall shadows, and fog effects.
func ExampleAtmosphericLightingSystem_basic() {
	// Create atmospheric lighting system for fantasy genre
	sys := lighting.NewAtmosphericLightingSystem("fantasy")

	// Register a torch light source at dungeon entrance
	torchPreset, _ := lighting.GetPresetByName("fantasy", "torch")
	torchLight := lighting.NewPointLight(10.0, 10.0, torchPreset, 12345)
	sys.RegisterLight(torchLight)

	// Register a wall that casts shadows
	sys.RegisterOccluder(15.0, 10.0, 2.0, 8.0, lighting.OccluderWall)

	// Calculate lighting at a point near the torch
	r, g, b, alpha := sys.CalculateLightingAtPoint(12.0, 10.0, 10.0, 10.0)

	fmt.Printf("Lighting near torch: R=%.2f G=%.2f B=%.2f A=%.2f\n", r, g, b, alpha)
	// Output will show warm, bright lighting characteristic of torches
}

// This example shows how to configure atmospheric effects for different genres.
func ExampleAtmosphericLightingSystem_genreConfiguration() {
	// Horror genre: heavy fog, dark shadows, cool color temperature
	horrorSys := lighting.NewAtmosphericLightingSystem("horror")
	horrorConfig := horrorSys.GetConfig()
	fmt.Printf("Horror fog density: %.2f\n", horrorConfig.FogDensity)
	fmt.Printf("Horror shadow darkness: %.2f\n", horrorConfig.ShadowDarkness)
	fmt.Printf("Horror color temperature: %.2f\n", horrorConfig.ColorTemperature)

	// Sci-fi genre: light fog, clean shadows, cool color temperature
	scifiSys := lighting.NewAtmosphericLightingSystem("scifi")
	scifiConfig := scifiSys.GetConfig()
	fmt.Printf("Sci-fi fog density: %.2f\n", scifiConfig.FogDensity)
	fmt.Printf("Sci-fi color temperature: %.2f\n", scifiConfig.ColorTemperature)

	// Output:
	// Horror fog density: 0.50
	// Horror shadow darkness: 0.85
	// Horror color temperature: -0.10
	// Sci-fi fog density: 0.20
	// Sci-fi color temperature: -0.20
}

// This example demonstrates shadow casting from multiple occluder types.
func ExampleAtmosphericLightingSystem_shadows() {
	sys := lighting.NewAtmosphericLightingSystem("fantasy")
	sys.RegisterLight(lighting.PointLight{
		Light: lighting.Light{
			X: 0, Y: 0, Radius: 30, Intensity: 1.0,
			R: 1.0, G: 1.0, B: 1.0,
		},
	})

	// Different occluder types cast different shadow strengths
	sys.RegisterOccluder(10.0, 0.0, 2.0, 4.0, lighting.OccluderWall)   // Full shadow
	sys.RegisterOccluder(20.0, 0.0, 2.0, 4.0, lighting.OccluderEntity) // Soft shadow
	sys.RegisterOccluder(30.0, 0.0, 2.0, 4.0, lighting.OccluderProp)   // Partial shadow

	// Point behind wall (darkest)
	r1, _, _, _ := sys.CalculateLightingAtPoint(12.0, 0.0, 0, 0)

	// Point behind entity (medium darkness)
	r2, _, _, _ := sys.CalculateLightingAtPoint(22.0, 0.0, 0, 0)

	// Point behind prop (lightest shadow)
	r3, _, _, _ := sys.CalculateLightingAtPoint(32.0, 0.0, 0, 0)

	fmt.Printf("Wall shadow: %.2f\n", r1)
	fmt.Printf("Entity shadow: %.2f\n", r2)
	fmt.Printf("Prop shadow: %.2f\n", r3)
	// Wall shadow will be darkest, prop shadow lightest
}

// This example shows atmospheric fog effects at different distances.
func ExampleAtmosphericLightingSystem_fog() {
	sys := lighting.NewAtmosphericLightingSystem("fantasy")

	// Add a bright white light at origin
	sys.RegisterLight(lighting.PointLight{
		Light: lighting.Light{
			X: 0, Y: 0, Radius: 50, Intensity: 1.0,
			R: 1.0, G: 1.0, B: 1.0,
		},
	})

	// Sample lighting at increasing distances from camera
	for dist := 5.0; dist <= 25.0; dist += 10.0 {
		r, g, b, _ := sys.CalculateLightingAtPoint(dist, 0, 0, 0)
		avg := (r + g + b) / 3.0
		fmt.Printf("Distance %.0f: brightness %.2f\n", dist, avg)
	}
	// Output shows decreasing brightness as fog thickens with distance
}

// This example demonstrates color temperature effects.
func ExampleAtmosphericLightingSystem_colorTemperature() {
	// Warm lighting (torchlight, firelight)
	warmSys := lighting.NewAtmosphericLightingSystem("fantasy")
	warmConfig := warmSys.GetConfig()
	warmConfig.ColorTemperature = 0.5 // Warm shift
	warmSys.SetConfig(warmConfig)

	warmSys.RegisterLight(lighting.PointLight{
		Light: lighting.Light{
			X: 0, Y: 0, Radius: 20, Intensity: 1.0,
			R: 1.0, G: 1.0, B: 1.0,
		},
	})

	r, g, b, _ := warmSys.CalculateLightingAtPoint(5, 5, 0, 0)
	fmt.Printf("Warm lighting - R:%.2f G:%.2f B:%.2f (more red/orange)\n", r, g, b)

	// Cool lighting (moonlight, artificial light)
	coolSys := lighting.NewAtmosphericLightingSystem("scifi")
	coolConfig := coolSys.GetConfig()
	coolConfig.ColorTemperature = -0.5 // Cool shift
	coolSys.SetConfig(coolConfig)

	coolSys.RegisterLight(lighting.PointLight{
		Light: lighting.Light{
			X: 0, Y: 0, Radius: 20, Intensity: 1.0,
			R: 1.0, G: 1.0, B: 1.0,
		},
	})

	r2, g2, b2, _ := coolSys.CalculateLightingAtPoint(5, 5, 0, 0)
	fmt.Printf("Cool lighting - R:%.2f G:%.2f B:%.2f (more blue)\n", r2, g2, b2)
}

// This example shows how to use atmospheric lighting in a game render loop.
func ExampleAtmosphericLightingSystem_gameIntegration() {
	sys := lighting.NewAtmosphericLightingSystem("fantasy")

	// Setup phase: register static lights and occluders
	// (This would be done once per level)
	sys.RegisterLight(lighting.PointLight{
		Light: lighting.Light{
			X: 20, Y: 20, Radius: 15, Intensity: 0.8,
			R: 1.0, G: 0.7, B: 0.4, // Warm torch color
		},
	})

	// Register walls as occluders
	sys.RegisterOccluder(25, 20, 2, 10, lighting.OccluderWall)

	// Per-frame: calculate lighting for each visible sprite/tile
	cameraX, cameraY := 0.0, 0.0

	// For a sprite at position (15, 20)
	spriteX, spriteY := 15.0, 20.0
	r, g, b, alpha := sys.CalculateLightingAtPoint(spriteX, spriteY, cameraX, cameraY)

	// Apply lighting to sprite color
	baseR, baseG, baseB := 0.8, 0.6, 0.4 // Sprite's natural color
	litR := baseR * r
	litG := baseG * g
	litB := baseB * b

	fmt.Printf("Lit sprite color: R=%.2f G=%.2f B=%.2f A=%.2f\n", litR, litG, litB, alpha)

	// Clear lights and occluders for next frame if dynamic
	sys.ClearLights()
	sys.ClearOccluders()
}

// This example shows depth fade for atmospheric perspective.
func ExampleAtmosphericLightingSystem_depthFade() {
	sys := lighting.NewAtmosphericLightingSystem("fantasy")

	// Add bright light at origin
	sys.RegisterLight(lighting.PointLight{
		Light: lighting.Light{
			X: 0, Y: 0, Radius: 40, Intensity: 1.0,
			R: 1.0, G: 0.5, B: 0.3, // Orange light
		},
	})

	// Near object (within depth fade start)
	_, _, _, alphaNear := sys.CalculateLightingAtPoint(5, 0, 0, 0)

	// Mid-distance object (in fade range)
	_, _, _, alphaMid := sys.CalculateLightingAtPoint(18, 0, 0, 0)

	// Far object (past fade end)
	_, _, _, alphaFar := sys.CalculateLightingAtPoint(30, 0, 0, 0)

	fmt.Printf("Near alpha: %.2f\n", alphaNear)
	fmt.Printf("Mid alpha: %.2f\n", alphaMid)
	fmt.Printf("Far alpha: %.2f\n", alphaFar)
	// Demonstrates progressive fade-out with distance
}

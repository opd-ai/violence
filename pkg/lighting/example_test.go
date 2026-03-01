// Package lighting provides shadow system example usage.
package lighting_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/lighting"
)

// ExampleShadowSystem demonstrates basic shadow system usage.
func ExampleShadowSystem() {
	// Create shadow system for fantasy genre
	shadowSys := lighting.NewShadowSystem(320, 240, "fantasy")

	// Create shadow casters (entities/props)
	casters := []lighting.ShadowCaster{
		{
			X:          10.0,
			Y:          10.0,
			Radius:     0.5,
			Height:     1.8,
			Opacity:    0.75,
			CastShadow: true,
		},
	}

	// Create light sources
	lights := []lighting.Light{
		{
			X:         8.0,
			Y:         8.0,
			Radius:    10.0,
			Intensity: 1.0,
			R:         1.0,
			G:         0.8,
			B:         0.4,
		},
	}

	// Render shadows to screen
	screen := ebiten.NewImage(320, 240)
	shadowSys.RenderShadows(screen, casters, lights, nil, 0, 0)
}

// ExampleShadowSystem_genrePresets demonstrates genre-specific shadow styles.
func ExampleShadowSystem_genrePresets() {
	// Horror genre: deep, soft shadows
	horror := lighting.NewShadowSystem(320, 240, "horror")
	_ = horror

	// Cyberpunk genre: hard, sharp shadows
	cyber := lighting.NewShadowSystem(320, 240, "cyberpunk")
	_ = cyber

	// Fantasy genre: medium soft shadows
	fantasy := lighting.NewShadowSystem(320, 240, "fantasy")
	_ = fantasy
}

// TestShadowSystemIntegration verifies shadow rendering with multiple light types.
func TestShadowSystemIntegration(t *testing.T) {
	shadowSys := lighting.NewShadowSystem(320, 240, "fantasy")
	screen := ebiten.NewImage(320, 240)
	defer screen.Dispose()

	// Multiple casters
	casters := []lighting.ShadowCaster{
		{X: 5, Y: 5, Radius: 0.5, Height: 1.0, Opacity: 0.7, CastShadow: true},
		{X: 10, Y: 10, Radius: 0.6, Height: 1.5, Opacity: 0.8, CastShadow: true},
		{X: 15, Y: 5, Radius: 0.4, Height: 0.8, Opacity: 0.6, CastShadow: true},
	}

	// Point lights
	lights := []lighting.Light{
		{X: 3, Y: 3, Radius: 8, Intensity: 1.0, R: 1.0, G: 0.8, B: 0.4},
		{X: 12, Y: 12, Radius: 6, Intensity: 0.8, R: 1.0, G: 1.0, B: 1.0},
	}

	// Cone lights (flashlights)
	coneLights := []lighting.ConeLight{
		{
			X:         8,
			Y:         8,
			DirX:      1,
			DirY:      0,
			Radius:    10,
			Angle:     0.5,
			Intensity: 0.9,
			R:         1.0,
			G:         1.0,
			B:         1.0,
			IsActive:  true,
		},
	}

	// Render shadows
	shadowSys.RenderShadows(screen, casters, lights, coneLights, 0, 0)
}

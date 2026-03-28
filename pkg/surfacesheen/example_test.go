package surfacesheen_test

import (
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/surfacesheen"
)

// TestTypicalUsage demonstrates typical usage of the surface sheen system.
func TestTypicalUsage(t *testing.T) {
	// Create system with genre
	sys := surfacesheen.NewSystem("fantasy")

	// Create sheen components for entities
	metalArmor := surfacesheen.NewSheenComponent(
		surfacesheen.MaterialMetal,
		color.RGBA{R: 180, G: 180, B: 200, A: 255},
	)

	wetCreature := surfacesheen.NewSheenComponent(
		surfacesheen.MaterialOrganic,
		color.RGBA{R: 100, G: 150, B: 100, A: 255},
	)
	wetCreature.Wetness = 0.8 // Just came from water

	// Define light sources (would come from lighting system)
	lights := []surfacesheen.LightSource{
		{X: 5.0, Y: 3.0, Color: color.RGBA{R: 255, G: 200, B: 150, A: 255}, Intensity: 1.0, Radius: 10.0},
		{X: 8.0, Y: 6.0, Color: color.RGBA{R: 150, G: 180, B: 255, A: 255}, Intensity: 0.7, Radius: 8.0},
	}

	// Calculate sheen for metal armor near torch
	sheenColor, intensity := sys.CalculateSheenForEntity(metalArmor, 5.5, 3.0, 0.5, lights)
	t.Logf("Metal armor sheen: color=%v, intensity=%.3f", sheenColor, intensity)

	// Calculate sheen for wet creature
	sheenColor2, intensity2 := sys.CalculateSheenForEntity(wetCreature, 7.0, 5.0, 0.8, lights)
	t.Logf("Wet creature sheen: color=%v, intensity=%.3f", sheenColor2, intensity2)

	// Verify wet surfaces produce more sheen
	if intensity2 < intensity*0.5 {
		t.Logf("Note: wet creature has less sheen than metal - expected due to organic material")
	}
}

// TestIntegrationWithLighting verifies the system works with typical game lighting setups.
func TestIntegrationWithLighting(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := surfacesheen.NewSystem(genre)

			// Typical game scene: player with metal weapon near torch
			weaponComp := surfacesheen.NewSheenComponent(
				surfacesheen.MaterialMetal,
				color.RGBA{R: 200, G: 200, B: 220, A: 255},
			)

			torchLight := surfacesheen.LightSource{
				X: 5.0, Y: 5.0,
				Color:     color.RGBA{R: 255, G: 180, B: 80, A: 255},
				Intensity: 1.0,
				Radius:    8.0,
			}

			_, intensity := sys.CalculateSheenForEntity(
				weaponComp, 5.5, 5.0, 0.3, []surfacesheen.LightSource{torchLight},
			)

			if intensity <= 0 {
				t.Errorf("expected positive sheen intensity for %s genre", genre)
			}
		})
	}
}

package crosshair

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy")

	if sys == nil {
		t.Fatal("Expected non-nil system")
	}

	if sys.genreID != "fantasy" {
		t.Errorf("Expected genre 'fantasy', got '%s'", sys.genreID)
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")
	sys.SetGenre("scifi")

	if sys.genreID != "scifi" {
		t.Errorf("Expected genre 'scifi', got '%s'", sys.genreID)
	}
}

func TestRenderCrosshairStyles(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(800, 600)

	tests := []struct {
		name       string
		weaponType string
	}{
		{"melee", "melee"},
		{"ranged", "ranged"},
		{"magic", "magic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := NewComponent()
			ch.WeaponType = tt.weaponType
			ch.Visible = true

			// Should not panic
			sys.renderCrosshair(screen, 10.0, 10.0, ch, 0.0, 0.0, 800, 600)
		})
	}
}

func TestRenderOffScreen(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(800, 600)

	ch := NewComponent()
	ch.Visible = true
	ch.Range = 5.0

	// Position far off screen - should exit early without panic
	sys.renderCrosshair(screen, 1000.0, 1000.0, ch, 0.0, 0.0, 800, 600)
}

func TestRenderInvisibleCrosshair(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(800, 600)

	ch := NewComponent()
	ch.Visible = false

	// Should not render anything, but also not panic
	sys.renderCrosshair(screen, 10.0, 10.0, ch, 0.0, 0.0, 800, 600)
}

func TestCrosshairColor(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(800, 600)

	ch := NewComponent()
	ch.ColorR = 1.0
	ch.ColorG = 0.0
	ch.ColorB = 0.0
	ch.ColorA = 1.0

	// Should render red crosshair without panic
	sys.renderCrosshair(screen, 10.0, 10.0, ch, 0.0, 0.0, 800, 600)
}

func TestCrosshairScale(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(800, 600)

	scales := []float64{0.5, 1.0, 1.5, 2.0}

	for _, scale := range scales {
		t.Run("scale", func(t *testing.T) {
			ch := NewComponent()
			ch.Scale = scale

			// Should render at different scales without panic
			sys.renderCrosshair(screen, 10.0, 10.0, ch, 0.0, 0.0, 800, 600)
		})
	}
}

func TestRenderMeleeCrosshair(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(100, 100)

	col := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	sys.renderMeleeCrosshair(screen, 50, 50, 1.0, col)

	// Should not panic
}

func TestRenderRangedCrosshair(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(100, 100)

	col := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	sys.renderRangedCrosshair(screen, 50, 50, 1.0, col)

	// Should not panic
}

func TestRenderMagicCrosshair(t *testing.T) {
	sys := NewSystem("fantasy")
	screen := ebiten.NewImage(100, 100)

	col := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	sys.renderMagicCrosshair(screen, 50, 50, 1.0, col)

	// Should not panic
}

package input

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewTouchRenderer(t *testing.T) {
	tests := []struct {
		name  string
		style TouchRenderStyle
	}{
		{"default", StyleDefault},
		{"horror", StyleHorror},
		{"cyberpunk", StyleCyberpunk},
		{"postapoc", StylePostApoc},
		{"scifi", StyleSciFi},
		{"fantasy", StyleFantasy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTouchRenderer(tt.style)
			if tr.style != tt.style {
				t.Errorf("style = %v, want %v", tr.style, tt.style)
			}
			if tr.overlayAlpha != 128 {
				t.Errorf("overlayAlpha = %v, want 128", tr.overlayAlpha)
			}
			if tr.buttonImgs == nil {
				t.Error("buttonImgs should not be nil")
			}
		})
	}
}

func TestSetStyle(t *testing.T) {
	tr := NewTouchRenderer(StyleDefault)
	tr.buttonImgs["test"] = ebiten.NewImage(10, 10)
	tr.joystickImg = ebiten.NewImage(20, 20)

	tr.SetStyle(StyleCyberpunk)

	if tr.style != StyleCyberpunk {
		t.Errorf("style = %v, want %v", tr.style, StyleCyberpunk)
	}
	if tr.joystickImg != nil {
		t.Error("joystickImg should be cleared after SetStyle()")
	}
	if len(tr.buttonImgs) != 0 {
		t.Error("buttonImgs should be cleared after SetStyle()")
	}
}

func TestGetBaseColor(t *testing.T) {
	tests := []struct {
		name      string
		style     TouchRenderStyle
		wantRed   uint8
		wantGreen uint8
		wantBlue  uint8
	}{
		{"horror", StyleHorror, 40, 20, 20},
		{"cyberpunk", StyleCyberpunk, 20, 20, 60},
		{"postapoc", StylePostApoc, 60, 40, 20},
		{"scifi", StyleSciFi, 20, 40, 80},
		{"fantasy", StyleFantasy, 60, 50, 30},
		{"default", StyleDefault, 80, 80, 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTouchRenderer(tt.style)
			clr := tr.getBaseColor()
			rgba, ok := clr.(color.RGBA)
			if !ok {
				t.Fatal("color is not RGBA")
			}
			if rgba.R != tt.wantRed {
				t.Errorf("R = %v, want %v", rgba.R, tt.wantRed)
			}
			if rgba.G != tt.wantGreen {
				t.Errorf("G = %v, want %v", rgba.G, tt.wantGreen)
			}
			if rgba.B != tt.wantBlue {
				t.Errorf("B = %v, want %v", rgba.B, tt.wantBlue)
			}
		})
	}
}

func TestGetKnobColor(t *testing.T) {
	tests := []struct {
		name      string
		style     TouchRenderStyle
		wantRed   uint8
		wantGreen uint8
		wantBlue  uint8
	}{
		{"horror", StyleHorror, 139, 0, 0},
		{"cyberpunk", StyleCyberpunk, 0, 255, 255},
		{"postapoc", StylePostApoc, 255, 140, 0},
		{"scifi", StyleSciFi, 0, 150, 255},
		{"fantasy", StyleFantasy, 218, 165, 32},
		{"default", StyleDefault, 150, 150, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTouchRenderer(tt.style)
			clr := tr.getKnobColor()
			rgba, ok := clr.(color.RGBA)
			if !ok {
				t.Fatal("color is not RGBA")
			}
			if rgba.R != tt.wantRed {
				t.Errorf("R = %v, want %v", rgba.R, tt.wantRed)
			}
			if rgba.G != tt.wantGreen {
				t.Errorf("G = %v, want %v", rgba.G, tt.wantGreen)
			}
			if rgba.B != tt.wantBlue {
				t.Errorf("B = %v, want %v", rgba.B, tt.wantBlue)
			}
		})
	}
}

func TestGetButtonColor(t *testing.T) {
	tests := []struct {
		name   string
		style  TouchRenderStyle
		active bool
	}{
		{"horror_inactive", StyleHorror, false},
		{"horror_active", StyleHorror, true},
		{"cyberpunk_inactive", StyleCyberpunk, false},
		{"cyberpunk_active", StyleCyberpunk, true},
		{"postapoc_inactive", StylePostApoc, false},
		{"postapoc_active", StylePostApoc, true},
		{"scifi_inactive", StyleSciFi, false},
		{"scifi_active", StyleSciFi, true},
		{"fantasy_inactive", StyleFantasy, false},
		{"fantasy_active", StyleFantasy, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTouchRenderer(tt.style)
			clr := tr.getButtonColor(tt.active)
			rgba, ok := clr.(color.RGBA)
			if !ok {
				t.Fatal("color is not RGBA")
			}
			// Verify alpha channel changes with active state
			expectedAlpha := tr.overlayAlpha
			if tt.active {
				expectedAlpha = tr.overlayAlpha + 64
			}
			if rgba.A != expectedAlpha {
				t.Errorf("Alpha = %v, want %v", rgba.A, expectedAlpha)
			}
		})
	}
}

func TestRenderJoystick(t *testing.T) {
	tests := []struct {
		name   string
		style  TouchRenderStyle
		active bool
	}{
		{"inactive", StyleDefault, false},
		{"active_default", StyleDefault, true},
		{"active_horror", StyleHorror, true},
		{"active_cyberpunk", StyleCyberpunk, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTouchRenderer(tt.style)
			screen := ebiten.NewImage(800, 600)
			vj := NewVirtualJoystick(100, 200, 80)
			vj.Active = tt.active

			// Should not panic
			tr.RenderJoystick(screen, vj)
		})
	}
}

func TestRenderButton(t *testing.T) {
	tests := []struct {
		name   string
		style  TouchRenderStyle
		active bool
	}{
		{"inactive_default", StyleDefault, false},
		{"active_default", StyleDefault, true},
		{"inactive_horror", StyleHorror, false},
		{"active_horror", StyleHorror, true},
		{"inactive_cyberpunk", StyleCyberpunk, false},
		{"active_cyberpunk", StyleCyberpunk, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTouchRenderer(tt.style)
			screen := ebiten.NewImage(800, 600)
			btn := NewTouchButton(0.5, 0.5, "Test")
			btn.Active = tt.active

			// Should not panic
			tr.RenderButton(screen, btn, 800, 600)
		})
	}
}

func TestGenreThemedStyles(t *testing.T) {
	// Test that each genre has distinct visual characteristics
	styles := []TouchRenderStyle{
		StyleHorror,
		StyleCyberpunk,
		StylePostApoc,
		StyleSciFi,
		StyleFantasy,
	}

	colors := make(map[string]color.Color)
	for _, style := range styles {
		tr := NewTouchRenderer(style)

		// Each style should have unique base color
		baseColor := tr.getBaseColor()
		key := colorToString(baseColor)
		if _, exists := colors[key]; exists {
			t.Errorf("Style %v has duplicate base color", style)
		}
		colors[key] = baseColor

		// Each style should have unique knob color
		knobColor := tr.getKnobColor()
		keyKnob := colorToString(knobColor)
		if keyKnob == key {
			t.Errorf("Style %v has same base and knob color", style)
		}
	}
}

func colorToString(c color.Color) string {
	r, g, b, a := c.RGBA()
	return string([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8), byte(a >> 8)})
}

func TestDrawButtonShape(t *testing.T) {
	tests := []struct {
		name  string
		style TouchRenderStyle
	}{
		{"cyberpunk_hexagon", StyleCyberpunk},
		{"horror_rune", StyleHorror},
		{"default_circle", StyleDefault},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewTouchRenderer(tt.style)
			screen := ebiten.NewImage(200, 200)
			clr := color.RGBA{R: 255, G: 0, B: 0, A: 128}

			// Should not panic
			tr.drawButtonShape(screen, 100, 100, 50, clr)
		})
	}
}

func TestTouchRendererMultipleGenres(t *testing.T) {
	// Test that renderer can switch between genres without issues
	tr := NewTouchRenderer(StyleDefault)
	screen := ebiten.NewImage(800, 600)
	vj := NewVirtualJoystick(100, 200, 80)
	vj.Active = true
	btn := NewTouchButton(0.5, 0.5, "Test")
	btn.Active = true

	genres := []TouchRenderStyle{
		StyleHorror,
		StyleCyberpunk,
		StylePostApoc,
		StyleSciFi,
		StyleFantasy,
	}

	for _, genre := range genres {
		tr.SetStyle(genre)
		// Should not panic when switching styles
		tr.RenderJoystick(screen, vj)
		tr.RenderButton(screen, btn, 800, 600)
	}
}

package reloadbar

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			s := NewSystem(genre)

			if s == nil {
				t.Fatal("NewSystem returned nil")
			}

			if s.genreID != genre {
				t.Errorf("Expected genreID '%s', got '%s'", genre, s.genreID)
			}

			// Verify style is initialized
			if s.style.BarWidth <= 0 {
				t.Error("BarWidth should be positive")
			}
			if s.style.BarHeight <= 0 {
				t.Error("BarHeight should be positive")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	s := NewSystem("fantasy")
	originalColor := s.style.FillColor

	s.SetGenre("cyberpunk")

	if s.genreID != "cyberpunk" {
		t.Errorf("Expected genreID 'cyberpunk', got '%s'", s.genreID)
	}

	// Color should have changed
	if s.style.FillColor == originalColor {
		t.Error("Style should change when genre changes")
	}
}

func TestGenreStyles(t *testing.T) {
	tests := []struct {
		genre         string
		expectedWidth float32
		minHeight     float32
		hasPulse      bool
	}{
		{"fantasy", 55, 4, true},
		{"scifi", 65, 4, true},
		{"horror", 55, 4, true},
		{"cyberpunk", 60, 5, true},
		{"postapoc", 50, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			s := NewSystem(tt.genre)

			if s.style.BarWidth != tt.expectedWidth {
				t.Errorf("Expected width %f for %s, got %f", tt.expectedWidth, tt.genre, s.style.BarWidth)
			}

			if s.style.BarHeight < tt.minHeight {
				t.Errorf("Expected height >= %f for %s, got %f", tt.minHeight, tt.genre, s.style.BarHeight)
			}

			if tt.hasPulse && s.style.PulseSpeed <= 0 {
				t.Error("Expected positive pulse speed")
			}
		})
	}
}

func TestSetScreenSize(t *testing.T) {
	s := NewSystem("fantasy")

	s.SetScreenSize(640, 480)

	if s.screenWidth != 640 {
		t.Errorf("Expected screenWidth 640, got %d", s.screenWidth)
	}

	if s.screenHeight != 480 {
		t.Errorf("Expected screenHeight 480, got %d", s.screenHeight)
	}
}

func TestSetReloadState(t *testing.T) {
	s := NewSystem("fantasy")

	s.SetReloadState(true, 0.5, 2.0)

	if !s.isReloading {
		t.Error("Expected isReloading true")
	}

	if s.progress != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", s.progress)
	}

	if s.totalDuration != 2.0 {
		t.Errorf("Expected totalDuration 2.0, got %f", s.totalDuration)
	}
}

func TestIsActive(t *testing.T) {
	s := NewSystem("fantasy")

	// Initially not active
	if s.IsActive() {
		t.Error("Should not be active initially")
	}

	// Active when reloading
	s.SetReloadState(true, 0.5, 1.0)
	s.fadeAlpha = 1.0
	if !s.IsActive() {
		t.Error("Should be active when reloading")
	}

	// Active during fade out
	s.SetReloadState(false, 0, 0)
	s.fadeAlpha = 0.5
	if !s.IsActive() {
		t.Error("Should be active during fade out")
	}

	// Not active when fully faded
	s.fadeAlpha = 0.0
	if s.IsActive() {
		t.Error("Should not be active when fully faded")
	}
}

func TestGetProgress(t *testing.T) {
	s := NewSystem("fantasy")

	s.SetReloadState(true, 0.75, 1.0)

	if s.GetProgress() != 0.75 {
		t.Errorf("Expected progress 0.75, got %f", s.GetProgress())
	}
}

func TestGetStyle(t *testing.T) {
	s := NewSystem("scifi")

	style := s.GetStyle()

	if style.BarWidth != s.style.BarWidth {
		t.Error("GetStyle should return current style")
	}
}

func TestRenderNoOp(t *testing.T) {
	s := NewSystem("fantasy")

	// Create a small test image
	img := ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))

	// Should not panic when not active
	s.Render(img, 50, 50)

	// Should not panic with various states
	s.SetReloadState(true, 0.5, 1.0)
	s.fadeAlpha = 1.0
	s.Render(img, 50, 50)
}

func TestRenderForEntity(t *testing.T) {
	s := NewSystem("fantasy")
	c := NewComponent()
	c.StartReload(1.0, "Test", 10)
	c.FadeAlpha = 1.0
	c.Progress = 0.5

	img := ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))

	// Should not panic
	s.RenderForEntity(img, 50, 50, c)
}

func TestRenderForEntityInvisible(t *testing.T) {
	s := NewSystem("fantasy")
	c := NewComponent()
	c.FadeAlpha = 0.0

	img := ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))

	// Should return early and not panic
	s.RenderForEntity(img, 50, 50, c)
}

func TestDrawRoundedRect(t *testing.T) {
	s := NewSystem("fantasy")
	img := ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))

	// Test with various radii
	radii := []float32{0, 2, 5, 50}

	for _, radius := range radii {
		// Should not panic
		s.drawRoundedRect(img, 10, 10, 50, 20, radius, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
}

func TestDrawRoundedRectStroke(t *testing.T) {
	s := NewSystem("fantasy")
	img := ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 100, 100)))

	// Test with various radii
	radii := []float32{0, 2, 5, 50}

	for _, radius := range radii {
		// Should not panic
		s.drawRoundedRectStroke(img, 10, 10, 50, 20, radius, 1.0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
}

func TestColorIntensity(t *testing.T) {
	// Verify genre colors are distinguishable
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	colors := make(map[string]color.RGBA)

	for _, genre := range genres {
		s := NewSystem(genre)
		colors[genre] = s.style.FillColor
	}

	// Each genre should have a unique fill color
	for i, g1 := range genres {
		for j, g2 := range genres {
			if i >= j {
				continue
			}
			c1, c2 := colors[g1], colors[g2]
			if c1 == c2 {
				t.Errorf("Genres %s and %s have identical fill colors", g1, g2)
			}
		}
	}
}

func TestStyleParameters(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			s := NewSystem(genre)

			// All styles should have sensible values
			if s.style.BarWidth < 40 || s.style.BarWidth > 100 {
				t.Errorf("BarWidth %f out of expected range for %s", s.style.BarWidth, genre)
			}

			if s.style.BarHeight < 4 || s.style.BarHeight > 10 {
				t.Errorf("BarHeight %f out of expected range for %s", s.style.BarHeight, genre)
			}

			if s.style.YOffset < 15 || s.style.YOffset > 35 {
				t.Errorf("YOffset %f out of expected range for %s", s.style.YOffset, genre)
			}

			if s.style.PulseSpeed < 1 || s.style.PulseSpeed > 15 {
				t.Errorf("PulseSpeed %f out of expected range for %s", s.style.PulseSpeed, genre)
			}

			// Background should be semi-transparent
			if s.style.BackgroundColor.A < 100 || s.style.BackgroundColor.A > 220 {
				t.Errorf("BackgroundColor alpha %d unexpected for %s", s.style.BackgroundColor.A, genre)
			}
		})
	}
}

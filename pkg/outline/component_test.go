package outline

import (
	"image/color"
	"testing"
)

func TestComponent_Type(t *testing.T) {
	c := &Component{
		Enabled:   true,
		Color:     color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Thickness: 2,
		Glow:      false,
	}

	if got := c.Type(); got != "OutlineComponent" {
		t.Errorf("Component.Type() = %v, want OutlineComponent", got)
	}
}

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSystem(tt.genreID)
			if s == nil {
				t.Fatal("NewSystem returned nil")
			}
			if s.genreID != tt.genreID {
				t.Errorf("genreID = %v, want %v", s.genreID, tt.genreID)
			}
			if s.outlineCache == nil {
				t.Error("outlineCache is nil")
			}
		})
	}
}

func TestSystem_SetGenre(t *testing.T) {
	s := NewSystem("fantasy")

	s.SetGenre("scifi")

	if s.genreID != "scifi" {
		t.Errorf("genreID = %v, want scifi", s.genreID)
	}
	if len(s.outlineCache) != 0 {
		t.Error("cache was not reset after SetGenre")
	}
}

func TestSystem_GetColors(t *testing.T) {
	s := NewSystem("fantasy")

	tests := []struct {
		name   string
		getter func() color.RGBA
	}{
		{"GetPlayerColor", s.GetPlayerColor},
		{"GetEnemyColor", s.GetEnemyColor},
		{"GetAllyColor", s.GetAllyColor},
		{"GetNeutralColor", s.GetNeutralColor},
		{"GetInteractColor", s.GetInteractColor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.getter()
			if c.A == 0 {
				t.Errorf("%s returned zero alpha color", tt.name)
			}
		})
	}
}

func TestSystem_GenreColors(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			s := NewSystem(genre)

			// Verify all colors are set
			if s.playerColor.A == 0 {
				t.Error("playerColor has zero alpha")
			}
			if s.enemyColor.A == 0 {
				t.Error("enemyColor has zero alpha")
			}
			if s.allyColor.A == 0 {
				t.Error("allyColor has zero alpha")
			}
			if s.neutralColor.A == 0 {
				t.Error("neutralColor has zero alpha")
			}
			if s.interactColor.A == 0 {
				t.Error("interactColor has zero alpha")
			}
		})
	}
}

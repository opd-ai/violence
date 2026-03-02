package automap

import (
	"image/color"
	"testing"
)

func TestGetGenreTheme(t *testing.T) {
	tests := []struct {
		genre string
		check func(GenreTheme) bool
		desc  string
	}{
		{
			genre: "fantasy",
			check: func(th GenreTheme) bool { return th.Player.B > th.Player.R },
			desc:  "fantasy player should be bluish",
		},
		{
			genre: "scifi",
			check: func(th GenreTheme) bool { return th.Background.B > th.Background.R },
			desc:  "scifi background should be bluish",
		},
		{
			genre: "horror",
			check: func(th GenreTheme) bool { return th.Enemy.R > th.Enemy.G+th.Enemy.B },
			desc:  "horror enemies should be reddish",
		},
		{
			genre: "cyberpunk",
			check: func(th GenreTheme) bool { return th.EnemyBoss.R > 0 && th.EnemyBoss.B > 0 },
			desc:  "cyberpunk bosses should be magenta-ish",
		},
		{
			genre: "postapoc",
			check: func(th GenreTheme) bool { return th.Wall.R > th.Wall.B },
			desc:  "postapoc walls should be brownish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			theme := GetGenreTheme(tt.genre)
			if !tt.check(theme) {
				t.Errorf("%s failed", tt.desc)
			}

			if theme.Background.A == 0 {
				t.Error("background should not be fully transparent")
			}
			if theme.Player.A != 255 {
				t.Error("player marker should be fully opaque")
			}
		})
	}
}

func TestGetGenreThemeDefault(t *testing.T) {
	theme := GetGenreTheme("unknown_genre")

	if theme.Background.R == 0 && theme.Background.G == 0 && theme.Background.B == 0 {
		t.Error("default theme should have non-zero background color")
	}

	if theme.Player.A != 255 {
		t.Error("default theme player marker should be fully opaque")
	}
}

func TestRenderConfigDefaults(t *testing.T) {
	cfg := RenderConfig{
		X:           10,
		Y:           10,
		Width:       150,
		Height:      150,
		PlayerX:     10.0,
		PlayerY:     10.0,
		PlayerAngle: 0.0,
	}

	if cfg.CellSize != 0 {
		t.Errorf("default cell size should be 0 before render, got %f", cfg.CellSize)
	}

	if cfg.Opacity != 0 {
		t.Errorf("default opacity should be 0 before render, got %f", cfg.Opacity)
	}
}

func TestEnemyMarkerData(t *testing.T) {
	marker := EnemyMarker{
		X:         10.5,
		Y:         15.3,
		IsBoss:    true,
		IsHostile: true,
		HealthPct: 0.75,
	}

	if marker.X != 10.5 {
		t.Errorf("marker X = %f, want 10.5", marker.X)
	}
	if marker.Y != 15.3 {
		t.Errorf("marker Y = %f, want 15.3", marker.Y)
	}
	if !marker.IsBoss {
		t.Error("marker should be boss")
	}
	if marker.HealthPct != 0.75 {
		t.Errorf("marker HealthPct = %f, want 0.75", marker.HealthPct)
	}
}

func TestItemMarkerData(t *testing.T) {
	marker := ItemMarker{
		X:       5.5,
		Y:       7.2,
		IsQuest: true,
		IsRare:  false,
	}

	if marker.X != 5.5 {
		t.Errorf("marker X = %f, want 5.5", marker.X)
	}
	if marker.Y != 7.2 {
		t.Errorf("marker Y = %f, want 7.2", marker.Y)
	}
	if !marker.IsQuest {
		t.Error("marker should be quest item")
	}
	if marker.IsRare {
		t.Error("marker should not be rare")
	}
}

func TestGenreThemeColorConsistency(t *testing.T) {
	themes := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range themes {
		theme := GetGenreTheme(genre)

		if theme.Player == (color.RGBA{}) {
			t.Errorf("%s: player color should not be zero", genre)
		}
		if theme.Enemy == (color.RGBA{}) {
			t.Errorf("%s: enemy color should not be zero", genre)
		}
		if theme.Wall == (color.RGBA{}) {
			t.Errorf("%s: wall color should not be zero", genre)
		}

		if theme.EnemyBoss.A == 0 {
			t.Errorf("%s: boss marker should not be fully transparent", genre)
		}
		if theme.ItemRare.A == 0 {
			t.Errorf("%s: rare item marker should not be fully transparent", genre)
		}
	}
}

func TestGenreThemeUniqueness(t *testing.T) {
	fantasy := GetGenreTheme("fantasy")
	scifi := GetGenreTheme("scifi")
	horror := GetGenreTheme("horror")

	if fantasy.Player == scifi.Player {
		t.Error("fantasy and scifi should have different player colors")
	}

	if fantasy.Background == horror.Background {
		t.Error("fantasy and horror should have different background colors")
	}

	if scifi.Enemy == horror.Enemy {
		t.Error("scifi and horror should have different enemy colors")
	}
}

func TestRenderConfigStructure(t *testing.T) {
	cfg := RenderConfig{
		X:            100,
		Y:            200,
		Width:        300,
		Height:       400,
		CellSize:     5.0,
		PlayerX:      10.0,
		PlayerY:      20.0,
		PlayerAngle:  1.57,
		Walls:        make([][]bool, 10),
		Enemies:      make([]EnemyMarker, 0),
		Items:        make([]ItemMarker, 0),
		Opacity:      0.9,
		ShowFogOfWar: true,
	}

	if cfg.X != 100 {
		t.Errorf("cfg.X = %f, want 100", cfg.X)
	}
	if cfg.Width != 300 {
		t.Errorf("cfg.Width = %f, want 300", cfg.Width)
	}
	if cfg.CellSize != 5.0 {
		t.Errorf("cfg.CellSize = %f, want 5.0", cfg.CellSize)
	}
	if !cfg.ShowFogOfWar {
		t.Error("cfg.ShowFogOfWar should be true")
	}
}

func TestMultipleEnemyMarkers(t *testing.T) {
	markers := []EnemyMarker{
		{X: 10, Y: 10, IsBoss: false, HealthPct: 1.0},
		{X: 11, Y: 11, IsBoss: true, HealthPct: 0.5},
		{X: 12, Y: 12, IsBoss: false, HealthPct: 0.8},
	}

	if len(markers) != 3 {
		t.Errorf("expected 3 markers, got %d", len(markers))
	}

	bossCount := 0
	for _, m := range markers {
		if m.IsBoss {
			bossCount++
		}
		if m.HealthPct < 0 || m.HealthPct > 1 {
			t.Errorf("health percentage out of range: %f", m.HealthPct)
		}
	}

	if bossCount != 1 {
		t.Errorf("expected 1 boss, got %d", bossCount)
	}
}

func TestMultipleItemMarkers(t *testing.T) {
	markers := []ItemMarker{
		{X: 5, Y: 5, IsQuest: false, IsRare: false},
		{X: 6, Y: 6, IsQuest: true, IsRare: false},
		{X: 7, Y: 7, IsQuest: false, IsRare: true},
		{X: 8, Y: 8, IsQuest: true, IsRare: true},
	}

	if len(markers) != 4 {
		t.Errorf("expected 4 markers, got %d", len(markers))
	}

	questCount := 0
	rareCount := 0
	for _, m := range markers {
		if m.IsQuest {
			questCount++
		}
		if m.IsRare {
			rareCount++
		}
	}

	if questCount != 2 {
		t.Errorf("expected 2 quest items, got %d", questCount)
	}
	if rareCount != 2 {
		t.Errorf("expected 2 rare items, got %d", rareCount)
	}
}

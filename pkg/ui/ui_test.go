package ui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewHUD(t *testing.T) {
	hud := NewHUD()

	if hud.Health != 100 {
		t.Errorf("expected Health=100, got %d", hud.Health)
	}
	if hud.MaxHealth != 100 {
		t.Errorf("expected MaxHealth=100, got %d", hud.MaxHealth)
	}
	if hud.Armor != 0 {
		t.Errorf("expected Armor=0, got %d", hud.Armor)
	}
	if hud.MaxArmor != 100 {
		t.Errorf("expected MaxArmor=100, got %d", hud.MaxArmor)
	}
	if hud.Ammo != 50 {
		t.Errorf("expected Ammo=50, got %d", hud.Ammo)
	}
	if hud.MaxAmmo != 200 {
		t.Errorf("expected MaxAmmo=200, got %d", hud.MaxAmmo)
	}
	if hud.WeaponName != "Pistol" {
		t.Errorf("expected WeaponName='Pistol', got %s", hud.WeaponName)
	}
	if hud.theme == nil {
		t.Error("expected theme to be initialized")
	}
}

func TestDrawHUD_NilHUD(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	DrawHUD(screen, nil) // Should not panic
}

func TestDrawHUD_Rendering(t *testing.T) {
	tests := []struct {
		name   string
		hud    *HUD
		width  int
		height int
	}{
		{
			name: "full_health_armor_ammo",
			hud: &HUD{
				Health:     100,
				Armor:      100,
				Ammo:       200,
				MaxHealth:  100,
				MaxArmor:   100,
				MaxAmmo:    200,
				WeaponName: "Shotgun",
				Keycards:   [3]bool{true, true, true},
				theme:      getDefaultTheme(),
			},
			width:  640,
			height: 480,
		},
		{
			name: "half_health_no_armor",
			hud: &HUD{
				Health:     50,
				Armor:      0,
				Ammo:       25,
				MaxHealth:  100,
				MaxArmor:   100,
				MaxAmmo:    50,
				WeaponName: "Pistol",
				Keycards:   [3]bool{false, false, false},
				theme:      getDefaultTheme(),
			},
			width:  320,
			height: 200,
		},
		{
			name: "zero_health",
			hud: &HUD{
				Health:     0,
				Armor:      50,
				Ammo:       0,
				MaxHealth:  100,
				MaxArmor:   100,
				MaxAmmo:    100,
				WeaponName: "Rifle",
				Keycards:   [3]bool{true, false, true},
				theme:      getDefaultTheme(),
			},
			width:  800,
			height: 600,
		},
		{
			name: "over_max_health",
			hud: &HUD{
				Health:     150,
				Armor:      120,
				Ammo:       250,
				MaxHealth:  100,
				MaxArmor:   100,
				MaxAmmo:    200,
				WeaponName: "Plasma Gun",
				Keycards:   [3]bool{false, true, false},
				theme:      getDefaultTheme(),
			},
			width:  1024,
			height: 768,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(tt.width, tt.height)
			// DrawHUD should not panic
			DrawHUD(screen, tt.hud)

			// Verify screen dimensions match
			bounds := screen.Bounds()
			if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
				t.Errorf("expected screen size %dx%d, got %dx%d", tt.width, tt.height, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGenre(tt.genreID)
			if currentTheme == nil {
				t.Error("expected SetGenre to set currentTheme")
			}

			// Verify theme has valid colors
			if currentTheme.HealthColor.A == 0 {
				t.Error("expected HealthColor to have alpha > 0")
			}
			if currentTheme.ArmorColor.A == 0 {
				t.Error("expected ArmorColor to have alpha > 0")
			}
			if currentTheme.AmmoColor.A == 0 {
				t.Error("expected AmmoColor to have alpha > 0")
			}
		})
	}
}

func TestGetThemeForGenre(t *testing.T) {
	tests := []struct {
		genreID          string
		expectedHealthR  uint8
		expectedArmorR   uint8
		expectedKeycards int
	}{
		{"fantasy", 200, 100, 3},
		{"scifi", 50, 100, 3},
		{"horror", 150, 80, 3},
		{"cyberpunk", 255, 100, 3},
		{"postapoc", 180, 120, 3},
		{"unknown", 200, 100, 3},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			theme := getThemeForGenre(tt.genreID)

			if theme == nil {
				t.Fatal("expected theme to be non-nil")
			}

			if theme.HealthColor.R != tt.expectedHealthR {
				t.Errorf("expected HealthColor.R=%d, got %d", tt.expectedHealthR, theme.HealthColor.R)
			}

			if theme.ArmorColor.R != tt.expectedArmorR {
				t.Errorf("expected ArmorColor.R=%d, got %d", tt.expectedArmorR, theme.ArmorColor.R)
			}

			if len(theme.KeycardColors) != tt.expectedKeycards {
				t.Errorf("expected %d keycards, got %d", tt.expectedKeycards, len(theme.KeycardColors))
			}

			// Verify all colors have alpha set
			if theme.HealthColor.A != 255 {
				t.Error("expected HealthColor alpha to be 255")
			}
			if theme.ArmorColor.A != 255 {
				t.Error("expected ArmorColor alpha to be 255")
			}
			if theme.AmmoColor.A != 255 {
				t.Error("expected AmmoColor alpha to be 255")
			}
		})
	}
}

func TestDrawStatusBar(t *testing.T) {
	tests := []struct {
		name    string
		current int
		max     int
	}{
		{"full", 100, 100},
		{"half", 50, 100},
		{"empty", 0, 100},
		{"over_max", 150, 100},
		{"zero_max", 50, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(200, 100)
			fillColor := color.RGBA{255, 0, 0, 255}
			bgColor := color.RGBA{30, 30, 30, 255}
			borderColor := color.RGBA{128, 128, 128, 255}

			// Should not panic
			drawStatusBar(screen, 10, 10, 100, 20, tt.current, tt.max, fillColor, bgColor, borderColor)

			// Verify screen dimensions
			bounds := screen.Bounds()
			if bounds.Dx() != 200 || bounds.Dy() != 100 {
				t.Errorf("expected screen size 200x100, got %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestDrawKeycard(t *testing.T) {
	screen := ebiten.NewImage(100, 100)
	keycardColor := color.RGBA{255, 0, 0, 255}

	// Should not panic
	drawKeycard(screen, 10, 10, keycardColor)

	// Verify screen dimensions
	bounds := screen.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("expected screen size 100x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestDrawLabel(t *testing.T) {
	screen := ebiten.NewImage(200, 100)
	textColor := color.RGBA{255, 255, 255, 255}

	// Should not panic
	drawLabel(screen, 10, 20, "TEST", textColor)

	// Verify screen dimensions
	bounds := screen.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("expected screen size 200x100, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestHUD_ThemeInitialization(t *testing.T) {
	hud := &HUD{
		Health:    100,
		MaxHealth: 100,
		theme:     nil,
	}

	screen := ebiten.NewImage(640, 480)
	DrawHUD(screen, hud)

	// After DrawHUD, theme should be initialized
	if hud.theme == nil {
		t.Error("expected theme to be initialized after DrawHUD")
	}
}

func TestHUD_AllKeycards(t *testing.T) {
	hud := NewHUD()
	hud.Keycards = [3]bool{true, true, true}

	screen := ebiten.NewImage(640, 480)
	// Should not panic with all keycards enabled
	DrawHUD(screen, hud)
}

func TestHUD_NoKeycards(t *testing.T) {
	hud := NewHUD()
	hud.Keycards = [3]bool{false, false, false}

	screen := ebiten.NewImage(640, 480)
	// Should not panic with no keycards
	DrawHUD(screen, hud)
}

func TestHUD_VariousResolutions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"320x200", 320, 200},
		{"640x480", 640, 480},
		{"800x600", 800, 600},
		{"1024x768", 1024, 768},
		{"1920x1080", 1920, 1080},
	}

	hud := NewHUD()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(tt.width, tt.height)
			DrawHUD(screen, hud)

			bounds := screen.Bounds()
			if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
				t.Errorf("expected screen size %dx%d, got %dx%d", tt.width, tt.height, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func BenchmarkDrawHUD(b *testing.B) {
	screen := ebiten.NewImage(640, 480)
	hud := NewHUD()
	hud.Health = 75
	hud.Armor = 50
	hud.Ammo = 100
	hud.Keycards = [3]bool{true, false, true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawHUD(screen, hud)
	}
}

func BenchmarkSetGenre(b *testing.B) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SetGenre(genres[i%len(genres)])
	}
}

func BenchmarkNewHUD(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewHUD()
	}
}

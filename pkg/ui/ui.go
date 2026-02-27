// Package ui provides HUD rendering and menu management.
package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// HUD holds heads-up display state.
type HUD struct {
	Health     int
	Armor      int
	Ammo       int
	WeaponID   int
	Keycards   [3]bool // Red, Blue, Yellow
	MaxHealth  int
	MaxArmor   int
	MaxAmmo    int
	WeaponName string
	theme      *Theme
}

// Menu represents a UI menu screen.
type Menu struct {
	Title string
	Items []string
}

// Theme holds genre-specific UI colors.
type Theme struct {
	HealthColor   color.RGBA
	ArmorColor    color.RGBA
	AmmoColor     color.RGBA
	BarBG         color.RGBA
	BarBorder     color.RGBA
	TextColor     color.RGBA
	KeycardColors [3]color.RGBA // Red, Blue, Yellow
}

var currentTheme = getDefaultTheme()

// NewHUD creates a HUD with default values.
func NewHUD() *HUD {
	return &HUD{
		Health:     100,
		Armor:      0,
		Ammo:       50,
		WeaponID:   1,
		Keycards:   [3]bool{false, false, false},
		MaxHealth:  100,
		MaxArmor:   100,
		MaxAmmo:    200,
		WeaponName: "Pistol",
		theme:      currentTheme,
	}
}

// DrawHUD renders the HUD onto the screen.
// Layout: Bottom-left has health/armor bars, bottom-center has ammo/weapon, bottom-right has keycards.
func DrawHUD(screen *ebiten.Image, h *HUD) {
	if h == nil {
		return
	}
	if h.theme == nil {
		h.theme = currentTheme
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Bottom-left: Health and Armor bars
	drawStatusBar(screen, 10, screenHeight-60, 150, 20, h.Health, h.MaxHealth, h.theme.HealthColor, h.theme.BarBG, h.theme.BarBorder)
	drawLabel(screen, 10, screenHeight-65, "HEALTH", h.theme.TextColor)

	drawStatusBar(screen, 10, screenHeight-30, 150, 20, h.Armor, h.MaxArmor, h.theme.ArmorColor, h.theme.BarBG, h.theme.BarBorder)
	drawLabel(screen, 10, screenHeight-35, "ARMOR", h.theme.TextColor)

	// Bottom-center: Ammo and Weapon
	centerX := screenWidth / 2
	drawStatusBar(screen, centerX-75, screenHeight-30, 150, 20, h.Ammo, h.MaxAmmo, h.theme.AmmoColor, h.theme.BarBG, h.theme.BarBorder)
	drawLabel(screen, centerX-75, screenHeight-35, "AMMO", h.theme.TextColor)
	drawLabel(screen, centerX-75, screenHeight-10, h.WeaponName, h.theme.TextColor)

	// Bottom-right: Keycards
	keycardX := screenWidth - 100
	for i := 0; i < 3; i++ {
		if h.Keycards[i] {
			drawKeycard(screen, keycardX+float32(i*25), screenHeight-40, h.theme.KeycardColors[i])
		}
	}
	drawLabel(screen, keycardX, screenHeight-45, "KEYS", h.theme.TextColor)
}

// drawStatusBar renders a horizontal status bar.
func drawStatusBar(screen *ebiten.Image, x, y, width, height float32, current, max int, fillColor, bgColor, borderColor color.RGBA) {
	// Background
	vector.DrawFilledRect(screen, x, y, width, height, bgColor, false)

	// Fill based on percentage
	if max > 0 {
		fillWidth := width * float32(current) / float32(max)
		if fillWidth > width {
			fillWidth = width
		}
		if fillWidth > 0 {
			vector.DrawFilledRect(screen, x, y, fillWidth, height, fillColor, false)
		}
	}

	// Border
	vector.StrokeRect(screen, x, y, width, height, 1, borderColor, false)
}

// drawKeycard renders a small keycard icon.
func drawKeycard(screen *ebiten.Image, x, y float32, c color.RGBA) {
	vector.DrawFilledRect(screen, x, y, 20, 12, c, false)
	vector.StrokeRect(screen, x, y, 20, 12, 1, color.RGBA{255, 255, 255, 255}, false)
}

// drawLabel renders text at the given position.
func drawLabel(screen *ebiten.Image, x, y float32, label string, c color.RGBA) {
	face := basicfont.Face7x13
	text.Draw(screen, label, face, int(x), int(y), c)
}

// DrawMenu renders a menu onto the screen.
func DrawMenu(screen *ebiten.Image, m *Menu) {}

// SetGenre configures UI theme for a genre.
func SetGenre(genreID string) {
	currentTheme = getThemeForGenre(genreID)
}

// getDefaultTheme returns the default UI theme.
func getDefaultTheme() *Theme {
	return getThemeForGenre("fantasy")
}

// getThemeForGenre returns genre-specific UI theme.
func getThemeForGenre(genreID string) *Theme {
	switch genreID {
	case "fantasy":
		return &Theme{
			HealthColor:   color.RGBA{200, 50, 50, 255},
			ArmorColor:    color.RGBA{100, 150, 200, 255},
			AmmoColor:     color.RGBA{200, 180, 50, 255},
			BarBG:         color.RGBA{30, 30, 30, 255},
			BarBorder:     color.RGBA{150, 120, 80, 255},
			TextColor:     color.RGBA{220, 200, 160, 255},
			KeycardColors: [3]color.RGBA{{255, 50, 50, 255}, {50, 100, 255, 255}, {255, 220, 50, 255}},
		}
	case "scifi":
		return &Theme{
			HealthColor:   color.RGBA{50, 200, 100, 255},
			ArmorColor:    color.RGBA{100, 150, 255, 255},
			AmmoColor:     color.RGBA{200, 200, 50, 255},
			BarBG:         color.RGBA{20, 25, 30, 255},
			BarBorder:     color.RGBA{100, 150, 200, 255},
			TextColor:     color.RGBA{180, 220, 255, 255},
			KeycardColors: [3]color.RGBA{{255, 50, 50, 255}, {50, 150, 255, 255}, {255, 220, 50, 255}},
		}
	case "horror":
		return &Theme{
			HealthColor:   color.RGBA{150, 30, 30, 255},
			ArmorColor:    color.RGBA{80, 80, 100, 255},
			AmmoColor:     color.RGBA{150, 130, 40, 255},
			BarBG:         color.RGBA{15, 10, 10, 255},
			BarBorder:     color.RGBA{100, 60, 50, 255},
			TextColor:     color.RGBA{180, 150, 130, 255},
			KeycardColors: [3]color.RGBA{{200, 30, 30, 255}, {30, 60, 150, 255}, {200, 180, 30, 255}},
		}
	case "cyberpunk":
		return &Theme{
			HealthColor:   color.RGBA{255, 50, 150, 255},
			ArmorColor:    color.RGBA{100, 255, 200, 255},
			AmmoColor:     color.RGBA{255, 200, 50, 255},
			BarBG:         color.RGBA{20, 10, 25, 255},
			BarBorder:     color.RGBA{150, 80, 180, 255},
			TextColor:     color.RGBA{255, 100, 200, 255},
			KeycardColors: [3]color.RGBA{{255, 50, 100, 255}, {50, 150, 255, 255}, {255, 200, 50, 255}},
		}
	case "postapoc":
		return &Theme{
			HealthColor:   color.RGBA{180, 60, 40, 255},
			ArmorColor:    color.RGBA{120, 100, 80, 255},
			AmmoColor:     color.RGBA{180, 150, 60, 255},
			BarBG:         color.RGBA{25, 20, 15, 255},
			BarBorder:     color.RGBA{120, 90, 70, 255},
			TextColor:     color.RGBA{200, 170, 130, 255},
			KeycardColors: [3]color.RGBA{{220, 60, 40, 255}, {60, 100, 180, 255}, {220, 180, 60, 255}},
		}
	default:
		return &Theme{
			HealthColor:   color.RGBA{200, 50, 50, 255},
			ArmorColor:    color.RGBA{100, 150, 200, 255},
			AmmoColor:     color.RGBA{200, 180, 50, 255},
			BarBG:         color.RGBA{30, 30, 30, 255},
			BarBorder:     color.RGBA{128, 128, 128, 255},
			TextColor:     color.RGBA{220, 220, 220, 255},
			KeycardColors: [3]color.RGBA{{255, 50, 50, 255}, {50, 100, 255, 255}, {255, 220, 50, 255}},
		}
	}
}

// Package ui provides HUD rendering and menu management.
package ui

import "github.com/hajimehoshi/ebiten/v2"

// HUD holds heads-up display state.
type HUD struct {
	Health int
	Armor  int
	Ammo   int
}

// Menu represents a UI menu screen.
type Menu struct {
	Title string
	Items []string
}

// DrawHUD renders the HUD onto the screen.
func DrawHUD(screen *ebiten.Image, h *HUD) {}

// DrawMenu renders a menu onto the screen.
func DrawMenu(screen *ebiten.Image, m *Menu) {}

// SetGenre configures UI theme for a genre.
func SetGenre(genreID string) {}

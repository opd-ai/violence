// Package ui provides HUD rendering and menu management.
package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/config"
	"github.com/opd-ai/violence/pkg/input"
	"golang.org/x/image/font/basicfont"
)

// HUD holds heads-up display state.
type HUD struct {
	Health      int
	Armor       int
	Ammo        int
	WeaponID    int
	Keycards    [3]bool // Red, Blue, Yellow
	MaxHealth   int
	MaxArmor    int
	MaxAmmo     int
	WeaponName  string
	theme       *Theme
	Message     string
	MessageTime int
}

// MenuType represents different menu screens.
type MenuType int

const (
	MenuTypeMain MenuType = iota
	MenuTypeDifficulty
	MenuTypeGenre
	MenuTypePause
	MenuTypeSettings
	MenuTypeShop
	MenuTypeCrafting
	MenuTypeSkills
	MenuTypeMods
	MenuTypeMultiplayer
)

// DifficultyLevel represents game difficulty.
type DifficultyLevel int

const (
	DifficultyEasy DifficultyLevel = iota
	DifficultyNormal
	DifficultyHard
	DifficultyNightmare
)

// SettingsCategory represents different settings sections.
type SettingsCategory int

const (
	SettingsCategoryVideo SettingsCategory = iota
	SettingsCategoryAudio
	SettingsCategoryControls
)

// MenuManager manages menu screens and navigation.
type MenuManager struct {
	currentMenu      MenuType
	selectedIndex    int
	difficulty       DifficultyLevel
	selectedGenre    string
	visible          bool
	menuItems        map[MenuType][]string
	difficultyNames  []string
	genreNames       []string
	settingsCategory SettingsCategory
	settingsOptions  map[SettingsCategory][]string
	editingBinding   bool
	bindingAction    string
}

// LoadingScreen manages loading screen display state.
type LoadingScreen struct {
	visible bool
	seed    uint64
	message string
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
		Health:      100,
		Armor:       0,
		Ammo:        50,
		WeaponID:    1,
		Keycards:    [3]bool{false, false, false},
		MaxHealth:   100,
		MaxArmor:    100,
		MaxAmmo:     200,
		WeaponName:  "Pistol",
		theme:       currentTheme,
		Message:     "",
		MessageTime: 0,
	}
}

// ShowMessage displays a temporary message on the HUD.
func (h *HUD) ShowMessage(msg string) {
	h.Message = msg
	h.MessageTime = 180
}

// Update decrements the message timer.
func (h *HUD) Update() {
	if h.MessageTime > 0 {
		h.MessageTime--
	}
	if h.MessageTime == 0 {
		h.Message = ""
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

	// Center message
	if h.MessageTime > 0 && h.Message != "" {
		msgX := centerX - float32(len(h.Message)*7/2)
		drawLabel(screen, msgX, screenHeight-80, h.Message, h.theme.TextColor)
	}
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

// NewMenuManager creates a new menu manager.
func NewMenuManager() *MenuManager {
	mm := &MenuManager{
		currentMenu:      MenuTypeMain,
		selectedIndex:    0,
		difficulty:       DifficultyNormal,
		selectedGenre:    "fantasy",
		visible:          false,
		settingsCategory: SettingsCategoryVideo,
		editingBinding:   false,
		bindingAction:    "",
		difficultyNames: []string{
			"Easy",
			"Normal",
			"Hard",
			"Nightmare",
		},
		genreNames: []string{
			"fantasy",
			"scifi",
			"horror",
			"cyberpunk",
			"postapoc",
		},
		menuItems:       make(map[MenuType][]string),
		settingsOptions: make(map[SettingsCategory][]string),
	}
	mm.menuItems[MenuTypeMain] = []string{
		"New Game",
		"Load Game",
		"Settings",
		"Quit",
	}
	mm.menuItems[MenuTypeDifficulty] = mm.difficultyNames
	mm.menuItems[MenuTypeGenre] = []string{
		"Fantasy",
		"Sci-Fi",
		"Horror",
		"Cyberpunk",
		"Post-Apocalyptic",
	}
	mm.menuItems[MenuTypePause] = []string{
		"Resume",
		"Shop",
		"Settings",
		"Save Game",
		"Main Menu",
	}
	mm.menuItems[MenuTypeSettings] = []string{
		"Video",
		"Audio",
		"Controls",
		"Back",
	}
	mm.settingsOptions[SettingsCategoryVideo] = []string{
		"Resolution",
		"VSync",
		"Fullscreen",
		"FOV",
		"Back",
	}
	mm.settingsOptions[SettingsCategoryAudio] = []string{
		"Master Volume",
		"Music Volume",
		"SFX Volume",
		"Back",
	}
	mm.settingsOptions[SettingsCategoryControls] = []string{
		"Move Forward",
		"Move Backward",
		"Strafe Left",
		"Strafe Right",
		"Fire",
		"Interact",
		"Mouse Sensitivity",
		"Back",
	}
	return mm
}

// Show displays the menu.
func (mm *MenuManager) Show(menuType MenuType) {
	mm.currentMenu = menuType
	mm.selectedIndex = 0
	mm.visible = true
}

// Hide hides the menu.
func (mm *MenuManager) Hide() {
	mm.visible = false
}

// IsVisible returns true if the menu is visible.
func (mm *MenuManager) IsVisible() bool {
	return mm.visible
}

// MoveUp moves the selection up.
func (mm *MenuManager) MoveUp() {
	if mm.selectedIndex > 0 {
		mm.selectedIndex--
	} else {
		items := mm.menuItems[mm.currentMenu]
		mm.selectedIndex = len(items) - 1
	}
}

// MoveDown moves the selection down.
func (mm *MenuManager) MoveDown() {
	items := mm.menuItems[mm.currentMenu]
	if mm.selectedIndex < len(items)-1 {
		mm.selectedIndex++
	} else {
		mm.selectedIndex = 0
	}
}

// GetSelectedIndex returns the currently selected menu item index.
func (mm *MenuManager) GetSelectedIndex() int {
	return mm.selectedIndex
}

// GetCurrentMenu returns the current menu type.
func (mm *MenuManager) GetCurrentMenu() MenuType {
	return mm.currentMenu
}

// GetSelectedItem returns the currently selected menu item text.
func (mm *MenuManager) GetSelectedItem() string {
	items := mm.menuItems[mm.currentMenu]
	if mm.selectedIndex >= 0 && mm.selectedIndex < len(items) {
		return items[mm.selectedIndex]
	}
	return ""
}

// SelectDifficulty selects the current difficulty.
func (mm *MenuManager) SelectDifficulty() DifficultyLevel {
	if mm.currentMenu == MenuTypeDifficulty && mm.selectedIndex < len(mm.difficultyNames) {
		mm.difficulty = DifficultyLevel(mm.selectedIndex)
	}
	return mm.difficulty
}

// GetDifficulty returns the current difficulty.
func (mm *MenuManager) GetDifficulty() DifficultyLevel {
	return mm.difficulty
}

// SelectGenre selects the current genre.
func (mm *MenuManager) SelectGenre() string {
	if mm.currentMenu == MenuTypeGenre && mm.selectedIndex < len(mm.genreNames) {
		mm.selectedGenre = mm.genreNames[mm.selectedIndex]
	}
	return mm.selectedGenre
}

// GetSelectedGenre returns the currently selected genre.
func (mm *MenuManager) GetSelectedGenre() string {
	return mm.selectedGenre
}

// SetSettingsCategory sets the current settings category.
func (mm *MenuManager) SetSettingsCategory(category SettingsCategory) {
	mm.settingsCategory = category
	mm.selectedIndex = 0
}

// GetSettingsCategory returns the current settings category.
func (mm *MenuManager) GetSettingsCategory() SettingsCategory {
	return mm.settingsCategory
}

// IsEditingBinding returns true if waiting for key input.
func (mm *MenuManager) IsEditingBinding() bool {
	return mm.editingBinding
}

// StartEditingBinding enters binding edit mode for the selected action.
func (mm *MenuManager) StartEditingBinding(action string) {
	mm.editingBinding = true
	mm.bindingAction = action
}

// StopEditingBinding exits binding edit mode.
func (mm *MenuManager) StopEditingBinding() {
	mm.editingBinding = false
	mm.bindingAction = ""
}

// GetEditingAction returns the action currently being bound.
func (mm *MenuManager) GetEditingAction() string {
	return mm.bindingAction
}

// GetSettingsItems returns the menu items for the current context.
func (mm *MenuManager) GetSettingsItems() []string {
	if mm.currentMenu == MenuTypeSettings {
		if mm.selectedIndex < len(mm.menuItems[MenuTypeSettings])-1 {
			return mm.settingsOptions[mm.settingsCategory]
		}
		return mm.menuItems[MenuTypeSettings]
	}
	return mm.menuItems[mm.currentMenu]
}

// DrawMenu renders a menu onto the screen.
func DrawMenu(screen *ebiten.Image, mm *MenuManager) {
	if mm == nil || !mm.visible {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay
	overlay := color.RGBA{0, 0, 0, 180}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	// Handle settings screen specially
	if mm.currentMenu == MenuTypeSettings {
		drawSettingsScreen(screen, mm, screenWidth, screenHeight)
		return
	}

	// Get menu title
	title := mm.getMenuTitle()
	items := mm.menuItems[mm.currentMenu]

	// Calculate menu dimensions
	itemHeight := float32(30)
	titleHeight := float32(50)
	menuHeight := titleHeight + float32(len(items))*itemHeight
	menuY := (screenHeight - menuHeight) / 2

	// Draw title
	titleX := screenWidth / 2
	drawCenteredLabel(screen, titleX, menuY, title, currentTheme.TextColor)

	// Draw menu items
	for i, item := range items {
		itemY := menuY + titleHeight + float32(i)*itemHeight

		// Highlight selected item
		if i == mm.selectedIndex {
			highlightColor := color.RGBA{80, 80, 120, 200}
			highlightX := screenWidth/2 - 150
			vector.DrawFilledRect(screen, highlightX, itemY-5, 300, itemHeight-5, highlightColor, false)
		}

		// Draw item text
		itemColor := currentTheme.TextColor
		if i == mm.selectedIndex {
			itemColor = color.RGBA{255, 255, 255, 255}
		}
		drawCenteredLabel(screen, titleX, itemY+20, item, itemColor)
	}

	// Draw additional info for difficulty and genre menus
	if mm.currentMenu == MenuTypeDifficulty {
		infoY := menuY + menuHeight + 30
		drawCenteredLabel(screen, titleX, infoY, mm.getDifficultyDescription(), color.RGBA{180, 180, 180, 255})
	}
}

// getMenuTitle returns the title for the current menu.
func (mm *MenuManager) getMenuTitle() string {
	switch mm.currentMenu {
	case MenuTypeMain:
		return "VIOLENCE"
	case MenuTypeDifficulty:
		return "SELECT DIFFICULTY"
	case MenuTypeGenre:
		return "SELECT GENRE"
	case MenuTypePause:
		return "PAUSED"
	case MenuTypeSettings:
		return "SETTINGS"
	case MenuTypeShop:
		return "ARMORY"
	case MenuTypeCrafting:
		return "CRAFTING"
	default:
		return "MENU"
	}
}

// getDifficultyDescription returns a description of the selected difficulty.
func (mm *MenuManager) getDifficultyDescription() string {
	switch DifficultyLevel(mm.selectedIndex) {
	case DifficultyEasy:
		return "For beginners - Less damage, more items"
	case DifficultyNormal:
		return "Standard experience - Balanced gameplay"
	case DifficultyHard:
		return "For veterans - More damage, fewer items"
	case DifficultyNightmare:
		return "Extreme challenge - Brutal combat"
	default:
		return ""
	}
}

// drawCenteredLabel renders centered text at the given position.
func drawCenteredLabel(screen *ebiten.Image, x, y float32, label string, c color.RGBA) {
	face := basicfont.Face7x13
	// Approximate text width (7 pixels per character)
	textWidth := len(label) * 7
	offsetX := textWidth / 2
	text.Draw(screen, label, face, int(x)-offsetX, int(y), c)
}

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

// drawSettingsScreen renders the settings menu with category navigation.
func drawSettingsScreen(screen *ebiten.Image, mm *MenuManager, screenWidth, screenHeight float32) {
	titleX := screenWidth / 2
	titleY := float32(50)

	// Draw main title
	drawCenteredLabel(screen, titleX, titleY, "SETTINGS", currentTheme.TextColor)

	// Determine if we're in a category or main settings
	items := mm.menuItems[MenuTypeSettings]
	inCategory := false
	categoryTitle := ""

	// Check if a category is selected (not "Back")
	if mm.selectedIndex < len(items)-1 {
		inCategory = true
		switch mm.selectedIndex {
		case 0:
			items = mm.settingsOptions[SettingsCategoryVideo]
			categoryTitle = "VIDEO SETTINGS"
		case 1:
			items = mm.settingsOptions[SettingsCategoryAudio]
			categoryTitle = "AUDIO SETTINGS"
		case 2:
			items = mm.settingsOptions[SettingsCategoryControls]
			categoryTitle = "CONTROL SETTINGS"
		}
	}

	// Draw category title if in a category
	if inCategory {
		drawCenteredLabel(screen, titleX, titleY+40, categoryTitle, color.RGBA{180, 180, 200, 255})
	}

	// If editing a binding, show prompt
	if mm.editingBinding {
		promptY := screenHeight / 2
		drawCenteredLabel(screen, titleX, promptY, "Press a key to bind...", color.RGBA{255, 255, 100, 255})
		promptY += 30
		drawCenteredLabel(screen, titleX, promptY, "ESC to cancel", color.RGBA{180, 180, 180, 255})
		return
	}

	// Calculate menu dimensions
	itemHeight := float32(30)
	startY := titleY + 100

	// Draw menu items with values
	for i, item := range items {
		itemY := startY + float32(i)*itemHeight

		// Highlight selected item
		if i == mm.selectedIndex {
			highlightColor := color.RGBA{80, 80, 120, 200}
			highlightX := screenWidth/2 - 200
			vector.DrawFilledRect(screen, highlightX, itemY-5, 400, itemHeight-5, highlightColor, false)
		}

		// Draw item label
		itemColor := currentTheme.TextColor
		if i == mm.selectedIndex {
			itemColor = color.RGBA{255, 255, 255, 255}
		}

		// Draw label on the left
		labelX := screenWidth/2 - 150
		drawLabel(screen, labelX, itemY+20, item, itemColor)

		// Draw value on the right if not "Back"
		if item != "Back" && inCategory {
			valueX := screenWidth/2 + 50
			value := getSettingValue(mm, item)
			drawLabel(screen, valueX, itemY+20, value, itemColor)
		}
	}

	// Draw navigation hint at bottom
	hintY := screenHeight - 40
	drawCenteredLabel(screen, titleX, hintY, "Arrow keys to navigate, Enter to select, ESC to go back", color.RGBA{150, 150, 150, 255})
}

// getSettingValue returns the current value for a setting option.
func getSettingValue(mm *MenuManager, option string) string {
	switch option {
	case "Resolution":
		return fmt.Sprintf("%dx%d", config.C.WindowWidth, config.C.WindowHeight)
	case "VSync":
		if config.C.VSync {
			return "ON"
		}
		return "OFF"
	case "Fullscreen":
		if config.C.FullScreen {
			return "ON"
		}
		return "OFF"
	case "FOV":
		return fmt.Sprintf("%.0f", config.C.FOV)
	case "Master Volume":
		return fmt.Sprintf("%.0f%%", config.C.MasterVolume*100)
	case "Music Volume":
		return fmt.Sprintf("%.0f%%", config.C.MusicVolume*100)
	case "SFX Volume":
		return fmt.Sprintf("%.0f%%", config.C.SFXVolume*100)
	case "Mouse Sensitivity":
		return fmt.Sprintf("%.1f", config.C.MouseSensitivity)
	case "Move Forward", "Move Backward", "Strafe Left", "Strafe Right", "Fire", "Interact":
		return getKeyNameForAction(option)
	default:
		return ""
	}
}

// getKeyNameForAction returns the key name for a control action.
func getKeyNameForAction(option string) string {
	var action input.Action
	switch option {
	case "Move Forward":
		action = input.ActionMoveForward
	case "Move Backward":
		action = input.ActionMoveBackward
	case "Strafe Left":
		action = input.ActionStrafeLeft
	case "Strafe Right":
		action = input.ActionStrafeRight
	case "Fire":
		action = input.ActionFire
	case "Interact":
		action = input.ActionInteract
	default:
		return ""
	}

	if keyCode, ok := config.C.KeyBindings[string(action)]; ok {
		return ebiten.Key(keyCode).String()
	}

	// Return default if not in config
	switch option {
	case "Move Forward":
		return "W"
	case "Move Backward":
		return "S"
	case "Strafe Left":
		return "A"
	case "Strafe Right":
		return "D"
	case "Fire":
		return "SPACE"
	case "Interact":
		return "E"
	default:
		return ""
	}
}

// ApplySettingChange modifies a setting value and persists to config.
func ApplySettingChange(option string, increase bool) error {
	delta := -1.0
	if increase {
		delta = 1.0
	}

	switch option {
	case "Resolution":
		resolutions := [][2]int{
			{640, 400}, {800, 500}, {960, 600}, {1280, 800}, {1600, 1000}, {1920, 1200},
		}
		currentIdx := 2 // Default 960x600
		for i, res := range resolutions {
			if res[0] == config.C.WindowWidth && res[1] == config.C.WindowHeight {
				currentIdx = i
				break
			}
		}
		if increase && currentIdx < len(resolutions)-1 {
			currentIdx++
		} else if !increase && currentIdx > 0 {
			currentIdx--
		}
		config.C.WindowWidth = resolutions[currentIdx][0]
		config.C.WindowHeight = resolutions[currentIdx][1]

	case "VSync":
		config.C.VSync = !config.C.VSync

	case "Fullscreen":
		config.C.FullScreen = !config.C.FullScreen

	case "FOV":
		config.C.FOV += delta * 5
		if config.C.FOV < 50 {
			config.C.FOV = 50
		}
		if config.C.FOV > 120 {
			config.C.FOV = 120
		}

	case "Master Volume":
		config.C.MasterVolume += delta * 0.1
		if config.C.MasterVolume < 0 {
			config.C.MasterVolume = 0
		}
		if config.C.MasterVolume > 1 {
			config.C.MasterVolume = 1
		}

	case "Music Volume":
		config.C.MusicVolume += delta * 0.1
		if config.C.MusicVolume < 0 {
			config.C.MusicVolume = 0
		}
		if config.C.MusicVolume > 1 {
			config.C.MusicVolume = 1
		}

	case "SFX Volume":
		config.C.SFXVolume += delta * 0.1
		if config.C.SFXVolume < 0 {
			config.C.SFXVolume = 0
		}
		if config.C.SFXVolume > 1 {
			config.C.SFXVolume = 1
		}

	case "Mouse Sensitivity":
		config.C.MouseSensitivity += delta * 0.1
		if config.C.MouseSensitivity < 0.1 {
			config.C.MouseSensitivity = 0.1
		}
		if config.C.MouseSensitivity > 5.0 {
			config.C.MouseSensitivity = 5.0
		}
	}

	return config.Save()
}

// ApplyKeyBinding binds a key to a control action.
func ApplyKeyBinding(option string, key ebiten.Key) error {
	if config.C.KeyBindings == nil {
		config.C.KeyBindings = make(map[string]int)
	}

	var action input.Action
	switch option {
	case "Move Forward":
		action = input.ActionMoveForward
	case "Move Backward":
		action = input.ActionMoveBackward
	case "Strafe Left":
		action = input.ActionStrafeLeft
	case "Strafe Right":
		action = input.ActionStrafeRight
	case "Fire":
		action = input.ActionFire
	case "Interact":
		action = input.ActionInteract
	default:
		return nil
	}

	config.C.KeyBindings[string(action)] = int(key)
	return config.Save()
}

// NewLoadingScreen creates a new loading screen.
func NewLoadingScreen() *LoadingScreen {
	return &LoadingScreen{
		visible: false,
		seed:    0,
		message: "Loading...",
	}
}

// Show displays the loading screen with the given seed and message.
func (ls *LoadingScreen) Show(seed uint64, message string) {
	ls.visible = true
	ls.seed = seed
	if message != "" {
		ls.message = message
	} else {
		ls.message = "Loading..."
	}
}

// Hide hides the loading screen.
func (ls *LoadingScreen) Hide() {
	ls.visible = false
}

// IsVisible returns true if the loading screen is visible.
func (ls *LoadingScreen) IsVisible() bool {
	return ls.visible
}

// GetSeed returns the current seed.
func (ls *LoadingScreen) GetSeed() uint64 {
	return ls.seed
}

// SetMessage updates the loading message.
func (ls *LoadingScreen) SetMessage(message string) {
	ls.message = message
}

// DrawLoadingScreen renders the loading screen.
func DrawLoadingScreen(screen *ebiten.Image, ls *LoadingScreen) {
	if ls == nil || !ls.visible {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw full-screen overlay
	overlay := color.RGBA{0, 0, 0, 240}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	// Calculate center position
	centerX := screenWidth / 2
	centerY := screenHeight / 2

	// Draw title
	titleY := centerY - 60
	drawCenteredLabel(screen, centerX, titleY, "VIOLENCE", currentTheme.TextColor)

	// Draw loading message
	messageY := centerY
	drawCenteredLabel(screen, centerX, messageY, ls.message, color.RGBA{200, 200, 200, 255})

	// Draw seed display
	seedY := centerY + 40
	seedText := fmt.Sprintf("Seed: %d", ls.seed)
	drawCenteredLabel(screen, centerX, seedY, seedText, color.RGBA{150, 150, 200, 255})

	// Draw animated loading indicator (simple dots)
	indicatorY := centerY + 80
	dots := getLoadingDots()
	drawCenteredLabel(screen, centerX, indicatorY, dots, color.RGBA{180, 180, 180, 255})
}

// getLoadingDots returns animated loading dots based on frame count.
func getLoadingDots() string {
	// Simple animation cycle: ".", "..", "...", "...."
	frameCount := (ebiten.ActualTPS() * 0.25)
	cycle := int(frameCount) % 4
	switch cycle {
	case 0:
		return "."
	case 1:
		return ".."
	case 2:
		return "..."
	default:
		return "...."
	}
}

// ShopItem represents an item displayed in the shop UI.
type ShopItem struct {
	ID    string
	Name  string
	Price int
	Stock int // -1 = unlimited
}

// ShopState holds the shop display state for rendering.
type ShopState struct {
	ShopName string
	Items    []ShopItem
	Credits  int
	Selected int
}

// DrawShop renders the shop overlay screen.
func DrawShop(screen *ebiten.Image, state *ShopState) {
	if state == nil {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay
	overlay := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	centerX := screenWidth / 2

	// Draw shop title
	titleY := float32(30)
	drawCenteredLabel(screen, centerX, titleY, state.ShopName, color.RGBA{255, 220, 100, 255})

	// Draw credits
	creditsY := titleY + 25
	creditsText := fmt.Sprintf("Credits: %d", state.Credits)
	drawCenteredLabel(screen, centerX, creditsY, creditsText, color.RGBA{200, 200, 100, 255})

	// Draw items list
	itemHeight := float32(25)
	startY := creditsY + 30

	if len(state.Items) == 0 {
		drawCenteredLabel(screen, centerX, startY+20, "No items available", color.RGBA{150, 150, 150, 255})
	}

	for i, item := range state.Items {
		itemY := startY + float32(i)*itemHeight

		// Highlight selected item
		if i == state.Selected {
			highlightX := centerX - 180
			vector.DrawFilledRect(screen, highlightX, itemY-5, 360, itemHeight-2, color.RGBA{80, 80, 120, 200}, false)
		}

		// Item name
		nameColor := currentTheme.TextColor
		if i == state.Selected {
			nameColor = color.RGBA{255, 255, 255, 255}
		}
		nameX := centerX - 170
		drawLabel(screen, nameX, itemY+10, item.Name, nameColor)

		// Price
		priceText := fmt.Sprintf("%d cr", item.Price)
		priceX := centerX + 80
		priceColor := color.RGBA{200, 200, 100, 255}
		if item.Price > state.Credits {
			priceColor = color.RGBA{200, 80, 80, 255} // Red if can't afford
		}
		drawLabel(screen, priceX, itemY+10, priceText, priceColor)

		// Stock
		stockX := centerX + 140
		stockText := "∞"
		if item.Stock >= 0 {
			stockText = fmt.Sprintf("x%d", item.Stock)
		}
		stockColor := color.RGBA{150, 150, 150, 255}
		if item.Stock == 0 {
			stockColor = color.RGBA{200, 80, 80, 255}
		}
		drawLabel(screen, stockX, itemY+10, stockText, stockColor)
	}

	// Draw controls hint
	hintY := screenHeight - 40
	drawCenteredLabel(screen, centerX, hintY, "Up/Down: Select | Enter: Buy | ESC: Back", color.RGBA{150, 150, 150, 255})
}

// CraftingRecipe represents a recipe displayed in the crafting UI.
type CraftingRecipe struct {
	ID        string
	Name      string
	Inputs    map[string]int
	OutputQty int
	CanCraft  bool
}

// CraftingState holds the crafting display state for rendering.
type CraftingState struct {
	Recipes    []CraftingRecipe
	ScrapName  string
	ScrapAmts  map[string]int
	Selected   int
	LastResult string // Status message for last craft attempt
}

// DrawCrafting renders the crafting overlay screen.
func DrawCrafting(screen *ebiten.Image, state *CraftingState) {
	if state == nil {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay
	overlay := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	centerX := screenWidth / 2

	// Draw crafting title
	titleY := float32(30)
	drawCenteredLabel(screen, centerX, titleY, "CRAFTING", color.RGBA{100, 220, 255, 255})

	// Draw scrap inventory
	scrapY := titleY + 25
	for scrapType, amount := range state.ScrapAmts {
		scrapText := fmt.Sprintf("%s: %d", scrapType, amount)
		drawCenteredLabel(screen, centerX, scrapY, scrapText, color.RGBA{180, 180, 100, 255})
		scrapY += 18
	}

	// Draw recipes list
	itemHeight := float32(25)
	startY := scrapY + 15

	if len(state.Recipes) == 0 {
		drawCenteredLabel(screen, centerX, startY+20, "No recipes available", color.RGBA{150, 150, 150, 255})
	}

	for i, recipe := range state.Recipes {
		itemY := startY + float32(i)*itemHeight

		// Highlight selected recipe
		if i == state.Selected {
			highlightX := centerX - 180
			vector.DrawFilledRect(screen, highlightX, itemY-5, 360, itemHeight-2, color.RGBA{60, 90, 120, 200}, false)
		}

		// Recipe name and output quantity
		nameColor := currentTheme.TextColor
		if !recipe.CanCraft {
			nameColor = color.RGBA{120, 120, 120, 255} // Dim if can't craft
		} else if i == state.Selected {
			nameColor = color.RGBA{255, 255, 255, 255}
		}

		recipeText := fmt.Sprintf("%s (x%d)", recipe.Name, recipe.OutputQty)
		nameX := centerX - 170
		drawLabel(screen, nameX, itemY+10, recipeText, nameColor)

		// Input cost
		costX := centerX + 60
		for mat, qty := range recipe.Inputs {
			costText := fmt.Sprintf("%s: %d", mat, qty)
			costColor := color.RGBA{180, 180, 100, 255}
			if !recipe.CanCraft {
				costColor = color.RGBA{200, 80, 80, 255}
			}
			drawLabel(screen, costX, itemY+10, costText, costColor)
			break // Only show first material (recipes typically have one input)
		}
	}

	// Draw last result message
	if state.LastResult != "" {
		resultY := screenHeight - 65
		drawCenteredLabel(screen, centerX, resultY, state.LastResult, color.RGBA{100, 255, 100, 255})
	}

	// Draw controls hint
	hintY := screenHeight - 40
	drawCenteredLabel(screen, centerX, hintY, "Up/Down: Select | Enter: Craft | ESC: Back", color.RGBA{150, 150, 150, 255})
}

// Select handles menu item selection and returns an action string.
func (mm *MenuManager) Select() string {
	item := mm.GetSelectedItem()
	switch mm.currentMenu {
	case MenuTypeMain:
		switch item {
		case "New Game":
			return "new_game"
		case "Load Game":
			return "load_game"
		case "Settings":
			return "settings"
		case "Quit":
			return "quit"
		}
	case MenuTypeDifficulty:
		mm.SelectDifficulty()
		return "difficulty_selected"
	case MenuTypeGenre:
		mm.SelectGenre()
		return "genre_selected"
	case MenuTypePause:
		switch item {
		case "Resume":
			return "resume"
		case "Shop":
			return "shop"
		case "Settings":
			return "settings"
		case "Save Game":
			return "save"
		case "Main Menu":
			return "quit_to_menu"
		}
	case MenuTypeSettings:
		// Handle settings navigation
		return "settings_action"
	case MenuTypeShop:
		// Shop item selection handled by index
		return "shop_buy"
	case MenuTypeCrafting:
		// Crafting recipe selection handled by index
		return "craft_item"
	}
	return ""
}

// Back navigates back in the menu hierarchy.
func (mm *MenuManager) Back() {
	switch mm.currentMenu {
	case MenuTypeDifficulty, MenuTypeGenre, MenuTypeSettings:
		mm.Show(MenuTypeMain)
	case MenuTypePause:
		// Pause menu back should resume game
		mm.Hide()
	case MenuTypeShop, MenuTypeCrafting:
		// Return to pause menu from shop/crafting
		mm.Show(MenuTypePause)
	}
}

// DrawTutorial renders a tutorial prompt on the screen.
func DrawTutorial(screen *ebiten.Image, message string) {
	if message == "" {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay at bottom
	overlayHeight := float32(80)
	overlayY := screenHeight - overlayHeight
	overlay := color.RGBA{0, 0, 0, 180}
	vector.DrawFilledRect(screen, 0, overlayY, screenWidth, overlayHeight, overlay, false)

	// Draw border
	borderColor := color.RGBA{100, 100, 200, 255}
	vector.StrokeLine(screen, 0, overlayY, screenWidth, overlayY, 2, borderColor, false)

	// Draw tutorial message centered
	centerX := screenWidth / 2
	textY := overlayY + 30
	drawCenteredLabel(screen, centerX, textY, message, color.RGBA{255, 255, 200, 255})

	// Draw "Press any key to continue" hint
	hintY := overlayY + 55
	drawCenteredLabel(screen, centerX, hintY, "Press any key to dismiss", color.RGBA{150, 150, 150, 255})
}

// CommandWheelPlayer represents a player option in the command wheel.
type CommandWheelPlayer struct {
	PlayerID uint64
	Name     string
	Health   float64
	Active   bool
}

// CommandWheel manages the squad command wheel UI.
type CommandWheel struct {
	visible         bool
	selectedIndex   int
	players         []*CommandWheelPlayer
	selectedCommand string // "follow", "hold", "attack"
}

// NewCommandWheel creates a new command wheel.
func NewCommandWheel() *CommandWheel {
	return &CommandWheel{
		visible:         false,
		selectedIndex:   0,
		players:         []*CommandWheelPlayer{},
		selectedCommand: "",
	}
}

// Show displays the command wheel with the specified command.
func (cw *CommandWheel) Show(command string) {
	cw.visible = true
	cw.selectedCommand = command
	cw.selectedIndex = 0
}

// Hide hides the command wheel.
func (cw *CommandWheel) Hide() {
	cw.visible = false
}

// IsVisible returns true if the command wheel is visible.
func (cw *CommandWheel) IsVisible() bool {
	return cw.visible
}

// SetPlayers updates the list of available players.
func (cw *CommandWheel) SetPlayers(players []*CommandWheelPlayer) {
	cw.players = players
	if cw.selectedIndex >= len(players) {
		cw.selectedIndex = 0
	}
}

// MoveUp moves the selection up in the wheel.
func (cw *CommandWheel) MoveUp() {
	if len(cw.players) == 0 {
		return
	}
	if cw.selectedIndex > 0 {
		cw.selectedIndex--
	} else {
		cw.selectedIndex = len(cw.players) - 1
	}
}

// MoveDown moves the selection down in the wheel.
func (cw *CommandWheel) MoveDown() {
	if len(cw.players) == 0 {
		return
	}
	if cw.selectedIndex < len(cw.players)-1 {
		cw.selectedIndex++
	} else {
		cw.selectedIndex = 0
	}
}

// GetSelectedPlayerID returns the currently selected player ID.
func (cw *CommandWheel) GetSelectedPlayerID() uint64 {
	if cw.selectedIndex >= 0 && cw.selectedIndex < len(cw.players) {
		return cw.players[cw.selectedIndex].PlayerID
	}
	return 0
}

// GetCommand returns the current command.
func (cw *CommandWheel) GetCommand() string {
	return cw.selectedCommand
}

// DrawCommandWheel renders the command wheel overlay.
func DrawCommandWheel(screen *ebiten.Image, cw *CommandWheel) {
	if cw == nil || !cw.visible {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay
	overlay := color.RGBA{0, 0, 0, 160}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	// Calculate center position
	centerX := screenWidth / 2
	centerY := screenHeight / 2

	// Draw command title
	titleY := centerY - 120
	commandTitle := getCommandTitle(cw.selectedCommand)
	drawCenteredLabel(screen, centerX, titleY, commandTitle, color.RGBA{255, 255, 100, 255})

	// Draw instruction
	instructionY := titleY + 30
	drawCenteredLabel(screen, centerX, instructionY, "Select Target Player", color.RGBA{200, 200, 200, 255})

	// Draw player list
	if len(cw.players) == 0 {
		noPlayersY := centerY
		drawCenteredLabel(screen, centerX, noPlayersY, "No players available", color.RGBA{180, 180, 180, 255})
		return
	}

	itemHeight := float32(40)
	listStartY := centerY - float32(len(cw.players))*itemHeight/2

	for i, player := range cw.players {
		itemY := listStartY + float32(i)*itemHeight

		// Highlight selected item
		if i == cw.selectedIndex {
			highlightColor := color.RGBA{80, 120, 80, 220}
			highlightX := centerX - 200
			vector.DrawFilledRect(screen, highlightX, itemY-5, 400, itemHeight-5, highlightColor, false)
			vector.StrokeRect(screen, highlightX, itemY-5, 400, itemHeight-5, 2, color.RGBA{120, 200, 120, 255}, false)
		}

		// Draw player name
		nameColor := currentTheme.TextColor
		if !player.Active {
			nameColor = color.RGBA{100, 100, 100, 255}
		} else if i == cw.selectedIndex {
			nameColor = color.RGBA{255, 255, 255, 255}
		}

		nameX := centerX - 180
		nameY := itemY + 15
		playerLabel := fmt.Sprintf("%s (ID: %d)", player.Name, player.PlayerID)
		drawLabel(screen, nameX, nameY, playerLabel, nameColor)

		// Draw health bar
		if player.Active {
			barX := centerX - 180
			barY := itemY + 22
			barWidth := float32(360)
			barHeight := float32(8)

			healthPercent := player.Health / 100.0
			if healthPercent > 1.0 {
				healthPercent = 1.0
			}
			if healthPercent < 0 {
				healthPercent = 0
			}

			// Draw health bar background
			vector.DrawFilledRect(screen, barX, barY, barWidth, barHeight, color.RGBA{40, 40, 40, 255}, false)

			// Draw health bar fill
			fillWidth := barWidth * float32(healthPercent)
			healthColor := currentTheme.HealthColor
			if healthPercent < 0.3 {
				healthColor = color.RGBA{200, 50, 50, 255}
			} else if healthPercent < 0.6 {
				healthColor = color.RGBA{200, 150, 50, 255}
			}
			vector.DrawFilledRect(screen, barX, barY, fillWidth, barHeight, healthColor, false)

			// Draw border
			vector.StrokeRect(screen, barX, barY, barWidth, barHeight, 1, color.RGBA{100, 100, 100, 255}, false)
		} else {
			// Draw "Disconnected" status
			statusX := centerX + 100
			statusY := itemY + 15
			drawLabel(screen, statusX, statusY, "[DISCONNECTED]", color.RGBA{150, 50, 50, 255})
		}
	}

	// Draw controls hint at bottom
	hintY := screenHeight - 60
	drawCenteredLabel(screen, centerX, hintY, "↑/↓ to select, Enter to confirm, ESC to cancel", color.RGBA{150, 150, 150, 255})
}

// getCommandTitle returns the display title for a command.
func getCommandTitle(command string) string {
	switch command {
	case "follow":
		return "FOLLOW PLAYER"
	case "hold":
		return "HOLD POSITION"
	case "attack":
		return "ATTACK TARGET"
	default:
		return "SQUAD COMMAND"
	}
}

// SkillNode represents a skill node displayed in the skills UI.
type SkillNode struct {
	ID          string
	Name        string
	Description string
	Cost        int
	Allocated   bool
	Available   bool // Prerequisites met and enough points
}

// SkillTreeState holds the skill tree display state for rendering.
type SkillTreeState struct {
	TreeName string
	TreeID   string
	Nodes    []SkillNode
	Points   int
	Selected int
}

// SkillsState holds the entire skills screen state.
type SkillsState struct {
	Trees       []SkillTreeState
	ActiveTree  int // Index of the active tree tab
	Selected    int // Selected node within active tree
	TotalPoints int
}

// DrawSkills renders the skills overlay screen.
func DrawSkills(screen *ebiten.Image, state *SkillsState) {
	if state == nil {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay
	overlay := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	centerX := screenWidth / 2

	// Draw title
	titleY := float32(25)
	drawCenteredLabel(screen, centerX, titleY, "SKILL TREES", color.RGBA{255, 220, 100, 255})

	// Draw available points
	pointsY := titleY + 22
	pointsText := fmt.Sprintf("Available Points: %d", state.TotalPoints)
	drawCenteredLabel(screen, centerX, pointsY, pointsText, color.RGBA{180, 255, 180, 255})

	// Draw tree tabs
	tabNames := []string{"COMBAT", "SURVIVAL", "TECH"}
	tabY := pointsY + 25
	tabWidth := screenWidth / float32(len(tabNames))
	for i, name := range tabNames {
		tabX := float32(i)*tabWidth + tabWidth/2
		tabColor := color.RGBA{120, 120, 120, 255}
		if i == state.ActiveTree {
			tabColor = color.RGBA{255, 200, 50, 255}
		}
		drawCenteredLabel(screen, tabX, tabY, name, tabColor)
	}

	// Draw nodes for active tree
	if state.ActiveTree >= 0 && state.ActiveTree < len(state.Trees) {
		tree := state.Trees[state.ActiveTree]
		startY := tabY + 30
		nodeHeight := float32(28)

		for i, node := range tree.Nodes {
			y := startY + float32(i)*nodeHeight

			// Determine colors
			nameColor := color.RGBA{180, 180, 180, 255}
			statusText := ""
			if node.Allocated {
				nameColor = color.RGBA{100, 255, 100, 255}
				statusText = " [UNLOCKED]"
			} else if node.Available {
				nameColor = color.RGBA{255, 255, 200, 255}
				statusText = fmt.Sprintf(" [Cost: %d]", node.Cost)
			} else {
				nameColor = color.RGBA{100, 100, 100, 255}
				statusText = " [LOCKED]"
			}

			// Highlight selected
			if i == state.Selected {
				vector.DrawFilledRect(screen, 20, y-2, screenWidth-40, nodeHeight-4, color.RGBA{80, 80, 120, 150}, false)
			}

			// Draw node name and status
			nodeText := node.Name + statusText
			drawLabel(screen, 30, y+12, nodeText, nameColor)

			// Draw description for selected node
			if i == state.Selected {
				descY := y + 14
				drawLabel(screen, 30, descY+12, node.Description, color.RGBA{150, 150, 200, 255})
			}
		}
	}

	// Draw controls hint
	hintY := screenHeight - 40
	drawCenteredLabel(screen, centerX, hintY, "←/→ tree, ↑/↓ node, Enter allocate, ESC back", color.RGBA{150, 150, 150, 255})
}

// ModInfo represents a mod displayed in the mods UI.
type ModInfo struct {
	Name        string
	Version     string
	Description string
	Author      string
	Enabled     bool
}

// ModsState holds the mods screen display state.
type ModsState struct {
	Mods     []ModInfo
	ModsDir  string
	Selected int
}

// DrawMods renders the mods screen overlay.
func DrawMods(screen *ebiten.Image, state *ModsState) {
	if state == nil {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay
	overlay := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	centerX := screenWidth / 2

	// Draw title
	titleY := float32(25)
	drawCenteredLabel(screen, centerX, titleY, "MODS", color.RGBA{180, 100, 255, 255})

	// Draw mods directory
	dirY := titleY + 22
	dirText := fmt.Sprintf("Directory: %s", state.ModsDir)
	drawCenteredLabel(screen, centerX, dirY, dirText, color.RGBA{150, 150, 150, 255})

	if len(state.Mods) == 0 {
		emptyY := dirY + 40
		drawCenteredLabel(screen, centerX, emptyY, "No mods found", color.RGBA{150, 150, 150, 255})
	} else {
		startY := dirY + 30
		itemHeight := float32(30)
		for i, mod := range state.Mods {
			y := startY + float32(i)*itemHeight

			// Highlight selected
			if i == state.Selected {
				vector.DrawFilledRect(screen, 20, y-2, screenWidth-40, itemHeight-4, color.RGBA{80, 60, 120, 150}, false)
			}

			// Status indicator
			statusColor := color.RGBA{100, 255, 100, 255}
			statusText := "[ON]"
			if !mod.Enabled {
				statusColor = color.RGBA{255, 100, 100, 255}
				statusText = "[OFF]"
			}

			modText := fmt.Sprintf("%s v%s - %s %s", mod.Name, mod.Version, mod.Author, statusText)
			drawLabel(screen, 30, y+12, modText, statusColor)
		}
	}

	// Draw controls hint
	hintY := screenHeight - 40
	drawCenteredLabel(screen, centerX, hintY, "↑/↓ select, Enter toggle, ESC back", color.RGBA{150, 150, 150, 255})
}

// MultiplayerMode represents a multiplayer game mode.
type MultiplayerMode struct {
	ID          string
	Name        string
	Description string
	MaxPlayers  int
}

// MultiplayerState holds the multiplayer lobby display state.
type MultiplayerState struct {
	Modes      []MultiplayerMode
	Selected   int
	Connected  bool
	ServerAddr string
	StatusMsg  string
}

// DrawMultiplayer renders the multiplayer lobby screen.
func DrawMultiplayer(screen *ebiten.Image, state *MultiplayerState) {
	if state == nil {
		return
	}

	bounds := screen.Bounds()
	screenWidth := float32(bounds.Dx())
	screenHeight := float32(bounds.Dy())

	// Draw semi-transparent overlay
	overlay := color.RGBA{0, 0, 0, 200}
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, overlay, false)

	centerX := screenWidth / 2

	// Draw title
	titleY := float32(25)
	drawCenteredLabel(screen, centerX, titleY, "MULTIPLAYER", color.RGBA{100, 200, 255, 255})

	// Draw connection status
	statusY := titleY + 22
	statusColor := color.RGBA{255, 100, 100, 255}
	statusText := "Disconnected"
	if state.Connected {
		statusColor = color.RGBA{100, 255, 100, 255}
		statusText = fmt.Sprintf("Connected: %s", state.ServerAddr)
	}
	drawCenteredLabel(screen, centerX, statusY, statusText, statusColor)

	// Draw game modes
	modesY := statusY + 30
	drawCenteredLabel(screen, centerX, modesY, "GAME MODES", color.RGBA{200, 200, 200, 255})

	startY := modesY + 22
	itemHeight := float32(30)
	for i, mode := range state.Modes {
		y := startY + float32(i)*itemHeight

		// Highlight selected
		if i == state.Selected {
			vector.DrawFilledRect(screen, 20, y-2, screenWidth-40, itemHeight-4, color.RGBA{60, 80, 120, 150}, false)
		}

		modeText := fmt.Sprintf("%s (%d players max)", mode.Name, mode.MaxPlayers)
		nameColor := color.RGBA{200, 200, 255, 255}
		if i == state.Selected {
			nameColor = color.RGBA{255, 255, 255, 255}
		}
		drawLabel(screen, 30, y+12, modeText, nameColor)
		drawLabel(screen, 30, y+24, mode.Description, color.RGBA{150, 150, 180, 255})
	}

	// Draw status message
	if state.StatusMsg != "" {
		msgY := startY + float32(len(state.Modes))*itemHeight + 20
		drawCenteredLabel(screen, centerX, msgY, state.StatusMsg, color.RGBA{255, 255, 100, 255})
	}

	// Draw controls hint
	hintY := screenHeight - 40
	drawCenteredLabel(screen, centerX, hintY, "↑/↓ select, Enter join, ESC back", color.RGBA{150, 150, 150, 255})
}

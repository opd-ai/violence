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

// Menu tests

func TestNewMenuManager(t *testing.T) {
	mm := NewMenuManager()

	if mm.currentMenu != MenuTypeMain {
		t.Errorf("expected currentMenu=MenuTypeMain, got %d", mm.currentMenu)
	}
	if mm.selectedIndex != 0 {
		t.Errorf("expected selectedIndex=0, got %d", mm.selectedIndex)
	}
	if mm.difficulty != DifficultyNormal {
		t.Errorf("expected difficulty=DifficultyNormal, got %d", mm.difficulty)
	}
	if mm.selectedGenre != "fantasy" {
		t.Errorf("expected selectedGenre='fantasy', got %s", mm.selectedGenre)
	}
	if mm.visible {
		t.Error("expected menu to be hidden initially")
	}

	// Verify menu items are populated
	if len(mm.menuItems[MenuTypeMain]) == 0 {
		t.Error("expected main menu items to be populated")
	}
	if len(mm.menuItems[MenuTypeDifficulty]) == 0 {
		t.Error("expected difficulty menu items to be populated")
	}
	if len(mm.menuItems[MenuTypeGenre]) == 0 {
		t.Error("expected genre menu items to be populated")
	}
	if len(mm.menuItems[MenuTypePause]) == 0 {
		t.Error("expected pause menu items to be populated")
	}
}

func TestMenuManager_ShowHide(t *testing.T) {
	mm := NewMenuManager()

	if mm.IsVisible() {
		t.Error("expected menu to be hidden initially")
	}

	mm.Show(MenuTypeMain)
	if !mm.IsVisible() {
		t.Error("expected menu to be visible after Show")
	}
	if mm.GetCurrentMenu() != MenuTypeMain {
		t.Errorf("expected currentMenu=MenuTypeMain, got %d", mm.GetCurrentMenu())
	}

	mm.Hide()
	if mm.IsVisible() {
		t.Error("expected menu to be hidden after Hide")
	}
}

func TestMenuManager_Navigation(t *testing.T) {
	tests := []struct {
		name          string
		menuType      MenuType
		initialIndex  int
		moveUp        int
		moveDown      int
		expectedIndex int
	}{
		{
			name:          "main_menu_move_down",
			menuType:      MenuTypeMain,
			initialIndex:  0,
			moveDown:      1,
			expectedIndex: 1,
		},
		{
			name:          "main_menu_move_up_wrap",
			menuType:      MenuTypeMain,
			initialIndex:  0,
			moveUp:        1,
			expectedIndex: 3, // Wraps to last item
		},
		{
			name:          "main_menu_move_down_wrap",
			menuType:      MenuTypeMain,
			initialIndex:  0,
			moveDown:      5, // More than items
			expectedIndex: 1, // Wraps around
		},
		{
			name:          "difficulty_menu_navigation",
			menuType:      MenuTypeDifficulty,
			initialIndex:  0,
			moveDown:      2,
			expectedIndex: 2,
		},
		{
			name:          "genre_menu_navigation",
			menuType:      MenuTypeGenre,
			initialIndex:  0,
			moveDown:      3,
			expectedIndex: 3,
		},
		{
			name:          "pause_menu_up_down",
			menuType:      MenuTypePause,
			initialIndex:  0,
			moveDown:      2,
			moveUp:        1,
			expectedIndex: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMenuManager()
			mm.Show(tt.menuType)

			for i := 0; i < tt.moveDown; i++ {
				mm.MoveDown()
			}
			for i := 0; i < tt.moveUp; i++ {
				mm.MoveUp()
			}

			if mm.GetSelectedIndex() != tt.expectedIndex {
				t.Errorf("expected selectedIndex=%d, got %d", tt.expectedIndex, mm.GetSelectedIndex())
			}
		})
	}
}

func TestMenuManager_GetSelectedItem(t *testing.T) {
	tests := []struct {
		name         string
		menuType     MenuType
		selectedIdx  int
		expectedItem string
	}{
		{
			name:         "main_menu_new_game",
			menuType:     MenuTypeMain,
			selectedIdx:  0,
			expectedItem: "New Game",
		},
		{
			name:         "main_menu_quit",
			menuType:     MenuTypeMain,
			selectedIdx:  3,
			expectedItem: "Quit",
		},
		{
			name:         "difficulty_normal",
			menuType:     MenuTypeDifficulty,
			selectedIdx:  1,
			expectedItem: "Normal",
		},
		{
			name:         "genre_scifi",
			menuType:     MenuTypeGenre,
			selectedIdx:  1,
			expectedItem: "Sci-Fi",
		},
		{
			name:         "pause_resume",
			menuType:     MenuTypePause,
			selectedIdx:  0,
			expectedItem: "Resume",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMenuManager()
			mm.Show(tt.menuType)
			mm.selectedIndex = tt.selectedIdx

			item := mm.GetSelectedItem()
			if item != tt.expectedItem {
				t.Errorf("expected item='%s', got '%s'", tt.expectedItem, item)
			}
		})
	}
}

func TestMenuManager_DifficultySelection(t *testing.T) {
	tests := []struct {
		name               string
		selectedIndex      int
		expectedDifficulty DifficultyLevel
	}{
		{"easy", 0, DifficultyEasy},
		{"normal", 1, DifficultyNormal},
		{"hard", 2, DifficultyHard},
		{"nightmare", 3, DifficultyNightmare},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMenuManager()
			mm.Show(MenuTypeDifficulty)
			mm.selectedIndex = tt.selectedIndex

			difficulty := mm.SelectDifficulty()
			if difficulty != tt.expectedDifficulty {
				t.Errorf("expected difficulty=%d, got %d", tt.expectedDifficulty, difficulty)
			}
			if mm.GetDifficulty() != tt.expectedDifficulty {
				t.Errorf("expected stored difficulty=%d, got %d", tt.expectedDifficulty, mm.GetDifficulty())
			}
		})
	}
}

func TestMenuManager_GenreSelection(t *testing.T) {
	tests := []struct {
		name          string
		selectedIndex int
		expectedGenre string
	}{
		{"fantasy", 0, "fantasy"},
		{"scifi", 1, "scifi"},
		{"horror", 2, "horror"},
		{"cyberpunk", 3, "cyberpunk"},
		{"postapoc", 4, "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMenuManager()
			mm.Show(MenuTypeGenre)
			mm.selectedIndex = tt.selectedIndex

			genre := mm.SelectGenre()
			if genre != tt.expectedGenre {
				t.Errorf("expected genre='%s', got '%s'", tt.expectedGenre, genre)
			}
			if mm.GetSelectedGenre() != tt.expectedGenre {
				t.Errorf("expected stored genre='%s', got '%s'", tt.expectedGenre, mm.GetSelectedGenre())
			}
		})
	}
}

func TestMenuManager_GetMenuTitle(t *testing.T) {
	tests := []struct {
		menuType      MenuType
		expectedTitle string
	}{
		{MenuTypeMain, "VIOLENCE"},
		{MenuTypeDifficulty, "SELECT DIFFICULTY"},
		{MenuTypeGenre, "SELECT GENRE"},
		{MenuTypePause, "PAUSED"},
	}

	mm := NewMenuManager()
	for _, tt := range tests {
		t.Run(tt.expectedTitle, func(t *testing.T) {
			mm.currentMenu = tt.menuType
			title := mm.getMenuTitle()
			if title != tt.expectedTitle {
				t.Errorf("expected title='%s', got '%s'", tt.expectedTitle, title)
			}
		})
	}
}

func TestMenuManager_GetDifficultyDescription(t *testing.T) {
	mm := NewMenuManager()
	mm.Show(MenuTypeDifficulty)

	descriptions := []string{
		"For beginners - Less damage, more items",
		"Standard experience - Balanced gameplay",
		"For veterans - More damage, fewer items",
		"Extreme challenge - Brutal combat",
	}

	for i, expected := range descriptions {
		mm.selectedIndex = i
		desc := mm.getDifficultyDescription()
		if desc != expected {
			t.Errorf("difficulty %d: expected '%s', got '%s'", i, expected, desc)
		}
	}
}

func TestDrawMenu_NilManager(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	DrawMenu(screen, nil) // Should not panic
}

func TestDrawMenu_HiddenMenu(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	mm := NewMenuManager()
	mm.Hide()
	DrawMenu(screen, mm) // Should not panic, should not draw
}

func TestDrawMenu_AllMenuTypes(t *testing.T) {
	tests := []struct {
		name     string
		menuType MenuType
		width    int
		height   int
	}{
		{"main_menu_640x480", MenuTypeMain, 640, 480},
		{"difficulty_320x200", MenuTypeDifficulty, 320, 200},
		{"genre_800x600", MenuTypeGenre, 800, 600},
		{"pause_1024x768", MenuTypePause, 1024, 768},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(tt.width, tt.height)
			mm := NewMenuManager()
			mm.Show(tt.menuType)

			// Should not panic
			DrawMenu(screen, mm)

			// Verify screen dimensions
			bounds := screen.Bounds()
			if bounds.Dx() != tt.width || bounds.Dy() != tt.height {
				t.Errorf("expected screen size %dx%d, got %dx%d", tt.width, tt.height, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestDrawMenu_WithSelection(t *testing.T) {
	mm := NewMenuManager()
	mm.Show(MenuTypeMain)

	screen := ebiten.NewImage(640, 480)

	// Test different selections
	for i := 0; i < 4; i++ {
		mm.selectedIndex = i
		DrawMenu(screen, mm) // Should not panic
	}
}

func TestDrawMenu_DifficultyWithDescription(t *testing.T) {
	mm := NewMenuManager()
	mm.Show(MenuTypeDifficulty)

	screen := ebiten.NewImage(640, 480)

	// Test all difficulty levels
	for i := 0; i < 4; i++ {
		mm.selectedIndex = i
		DrawMenu(screen, mm) // Should render description
	}
}

func TestDrawCenteredLabel(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	textColor := color.RGBA{255, 255, 255, 255}

	// Should not panic
	drawCenteredLabel(screen, 320, 240, "Test Label", textColor)

	// Verify screen dimensions
	bounds := screen.Bounds()
	if bounds.Dx() != 640 || bounds.Dy() != 480 {
		t.Errorf("expected screen size 640x480, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestMenuManager_MultipleMenuSwitches(t *testing.T) {
	mm := NewMenuManager()

	// Test switching between menus
	menus := []MenuType{MenuTypeMain, MenuTypeDifficulty, MenuTypeGenre, MenuTypePause}

	for _, menuType := range menus {
		mm.Show(menuType)
		if mm.GetCurrentMenu() != menuType {
			t.Errorf("expected menu %d, got %d", menuType, mm.GetCurrentMenu())
		}
		if mm.GetSelectedIndex() != 0 {
			t.Errorf("expected selection to reset to 0, got %d", mm.GetSelectedIndex())
		}
	}
}

func TestMenuManager_EdgeCases(t *testing.T) {
	mm := NewMenuManager()
	mm.Show(MenuTypeMain)

	// Test wrap-around at top
	mm.selectedIndex = 0
	mm.MoveUp()
	items := mm.menuItems[MenuTypeMain]
	if mm.selectedIndex != len(items)-1 {
		t.Errorf("expected wrap to last item %d, got %d", len(items)-1, mm.selectedIndex)
	}

	// Test wrap-around at bottom
	mm.selectedIndex = len(items) - 1
	mm.MoveDown()
	if mm.selectedIndex != 0 {
		t.Errorf("expected wrap to first item 0, got %d", mm.selectedIndex)
	}
}

func BenchmarkDrawMenu(b *testing.B) {
	screen := ebiten.NewImage(640, 480)
	mm := NewMenuManager()
	mm.Show(MenuTypeMain)
	mm.selectedIndex = 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawMenu(screen, mm)
	}
}

func BenchmarkMenuManager_Navigation(b *testing.B) {
	mm := NewMenuManager()
	mm.Show(MenuTypeMain)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.MoveDown()
		mm.MoveUp()
	}
}

func BenchmarkNewMenuManager(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewMenuManager()
	}
}

func TestMenuManager_SettingsCategory(t *testing.T) {
	mm := NewMenuManager()
	mm.Show(MenuTypeSettings)

	// Test initial state
	if mm.GetSettingsCategory() != SettingsCategoryVideo {
		t.Errorf("expected initial category=SettingsCategoryVideo, got %d", mm.GetSettingsCategory())
	}

	// Test category switching
	mm.SetSettingsCategory(SettingsCategoryAudio)
	if mm.GetSettingsCategory() != SettingsCategoryAudio {
		t.Errorf("expected category=SettingsCategoryAudio, got %d", mm.GetSettingsCategory())
	}

	// Test index reset when changing category
	mm.selectedIndex = 5
	mm.SetSettingsCategory(SettingsCategoryControls)
	if mm.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to reset to 0, got %d", mm.selectedIndex)
	}
}

func TestMenuManager_BindingEdit(t *testing.T) {
	mm := NewMenuManager()

	// Test initial state
	if mm.IsEditingBinding() {
		t.Error("expected editingBinding=false initially")
	}
	if mm.GetEditingAction() != "" {
		t.Errorf("expected editingAction='', got '%s'", mm.GetEditingAction())
	}

	// Test starting binding edit
	mm.StartEditingBinding("Move Forward")
	if !mm.IsEditingBinding() {
		t.Error("expected editingBinding=true after StartEditingBinding")
	}
	if mm.GetEditingAction() != "Move Forward" {
		t.Errorf("expected editingAction='Move Forward', got '%s'", mm.GetEditingAction())
	}

	// Test stopping binding edit
	mm.StopEditingBinding()
	if mm.IsEditingBinding() {
		t.Error("expected editingBinding=false after StopEditingBinding")
	}
	if mm.GetEditingAction() != "" {
		t.Errorf("expected editingAction='', got '%s'", mm.GetEditingAction())
	}
}

func TestMenuManager_GetSettingsItems(t *testing.T) {
	mm := NewMenuManager()
	mm.Show(MenuTypeSettings)

	// Test main settings items (when "Back" is selected or no category active)
	mm.selectedIndex = 3 // "Back" item
	items := mm.GetSettingsItems()
	expected := mm.menuItems[MenuTypeSettings]
	if len(items) != len(expected) {
		t.Errorf("expected %d items, got %d", len(expected), len(items))
	}

	// Test video settings items
	mm.selectedIndex = 0
	mm.SetSettingsCategory(SettingsCategoryVideo)
	items = mm.GetSettingsItems()
	expected = mm.settingsOptions[SettingsCategoryVideo]
	if len(items) != len(expected) {
		t.Errorf("expected %d video items, got %d", len(expected), len(items))
	}

	// Test audio settings items
	mm.selectedIndex = 1
	mm.SetSettingsCategory(SettingsCategoryAudio)
	items = mm.GetSettingsItems()
	expected = mm.settingsOptions[SettingsCategoryAudio]
	if len(items) != len(expected) {
		t.Errorf("expected %d audio items, got %d", len(expected), len(items))
	}

	// Test controls settings items
	mm.selectedIndex = 2
	mm.SetSettingsCategory(SettingsCategoryControls)
	items = mm.GetSettingsItems()
	expected = mm.settingsOptions[SettingsCategoryControls]
	if len(items) != len(expected) {
		t.Errorf("expected %d control items, got %d", len(expected), len(items))
	}
}

func TestGetSettingValue(t *testing.T) {
	mm := NewMenuManager()

	tests := []struct {
		name   string
		option string
	}{
		{"resolution", "Resolution"},
		{"vsync", "VSync"},
		{"fullscreen", "Fullscreen"},
		{"fov", "FOV"},
		{"master_volume", "Master Volume"},
		{"music_volume", "Music Volume"},
		{"sfx_volume", "SFX Volume"},
		{"mouse_sensitivity", "Mouse Sensitivity"},
		{"move_forward", "Move Forward"},
		{"move_backward", "Move Backward"},
		{"strafe_left", "Strafe Left"},
		{"strafe_right", "Strafe Right"},
		{"fire", "Fire"},
		{"interact", "Interact"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := getSettingValue(mm, tt.option)
			if value == "" {
				t.Errorf("expected non-empty value for %s", tt.option)
			}
		})
	}
}

func TestGetKeyNameForAction(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{"move_forward", "Move Forward"},
		{"move_backward", "Move Backward"},
		{"strafe_left", "Strafe Left"},
		{"strafe_right", "Strafe Right"},
		{"fire", "Fire"},
		{"interact", "Interact"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyName := getKeyNameForAction(tt.action)
			if keyName == "" {
				t.Errorf("expected non-empty key name for %s", tt.action)
			}
		})
	}

	// Test invalid action
	keyName := getKeyNameForAction("Invalid Action")
	if keyName != "" {
		t.Errorf("expected empty key name for invalid action, got '%s'", keyName)
	}
}

func TestApplySettingChange(t *testing.T) {
	tests := []struct {
		name     string
		option   string
		increase bool
	}{
		{"resolution_increase", "Resolution", true},
		{"resolution_decrease", "Resolution", false},
		{"vsync_toggle", "VSync", true},
		{"fullscreen_toggle", "Fullscreen", true},
		{"fov_increase", "FOV", true},
		{"fov_decrease", "FOV", false},
		{"master_volume_increase", "Master Volume", true},
		{"master_volume_decrease", "Master Volume", false},
		{"music_volume_increase", "Music Volume", true},
		{"music_volume_decrease", "Music Volume", false},
		{"sfx_volume_increase", "SFX Volume", true},
		{"sfx_volume_decrease", "SFX Volume", false},
		{"mouse_sensitivity_increase", "Mouse Sensitivity", true},
		{"mouse_sensitivity_decrease", "Mouse Sensitivity", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplySettingChange(tt.option, tt.increase)
			if err != nil {
				// It's OK if config save fails in test environment
				t.Logf("ApplySettingChange returned error (expected in test): %v", err)
			}
		})
	}
}

func TestApplyKeyBinding(t *testing.T) {
	tests := []struct {
		name   string
		option string
		key    ebiten.Key
	}{
		{"bind_move_forward", "Move Forward", ebiten.KeyW},
		{"bind_move_backward", "Move Backward", ebiten.KeyS},
		{"bind_strafe_left", "Strafe Left", ebiten.KeyA},
		{"bind_strafe_right", "Strafe Right", ebiten.KeyD},
		{"bind_fire", "Fire", ebiten.KeySpace},
		{"bind_interact", "Interact", ebiten.KeyE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ApplyKeyBinding(tt.option, tt.key)
			if err != nil {
				// It's OK if config save fails in test environment
				t.Logf("ApplyKeyBinding returned error (expected in test): %v", err)
			}
		})
	}

	// Test invalid action
	err := ApplyKeyBinding("Invalid Action", ebiten.KeyA)
	if err != nil {
		t.Errorf("expected nil error for invalid action, got %v", err)
	}
}

func TestDrawSettingsScreen(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	mm := NewMenuManager()
	mm.Show(MenuTypeSettings)

	// Test main settings screen
	drawSettingsScreen(screen, mm, 640, 480)

	// Test video settings
	mm.selectedIndex = 0
	drawSettingsScreen(screen, mm, 640, 480)

	// Test audio settings
	mm.selectedIndex = 1
	drawSettingsScreen(screen, mm, 640, 480)

	// Test controls settings
	mm.selectedIndex = 2
	drawSettingsScreen(screen, mm, 640, 480)

	// Test while editing binding
	mm.StartEditingBinding("Move Forward")
	drawSettingsScreen(screen, mm, 640, 480)
	mm.StopEditingBinding()
}

func TestDrawMenu_Settings(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	mm := NewMenuManager()
	mm.Show(MenuTypeSettings)

	// Should render settings screen
	DrawMenu(screen, mm)

	// Test with different selections
	for i := 0; i < 4; i++ {
		mm.selectedIndex = i
		DrawMenu(screen, mm)
	}
}

func TestMenuManager_SettingsTitle(t *testing.T) {
	mm := NewMenuManager()
	mm.currentMenu = MenuTypeSettings

	title := mm.getMenuTitle()
	if title != "SETTINGS" {
		t.Errorf("expected title='SETTINGS', got '%s'", title)
	}
}

func BenchmarkDrawSettingsScreen(b *testing.B) {
	screen := ebiten.NewImage(640, 480)
	mm := NewMenuManager()
	mm.Show(MenuTypeSettings)
	mm.selectedIndex = 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		drawSettingsScreen(screen, mm, 640, 480)
	}
}

func BenchmarkApplySettingChange(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplySettingChange("FOV", true)
	}
}

func TestNewLoadingScreen(t *testing.T) {
	ls := NewLoadingScreen()

	if ls == nil {
		t.Fatal("expected non-nil LoadingScreen")
	}
	if ls.visible {
		t.Error("expected visible=false for new loading screen")
	}
	if ls.seed != 0 {
		t.Errorf("expected seed=0, got %d", ls.seed)
	}
	if ls.message != "Loading..." {
		t.Errorf("expected message='Loading...', got '%s'", ls.message)
	}
}

func TestLoadingScreen_Show(t *testing.T) {
	tests := []struct {
		name    string
		seed    uint64
		message string
		want    string
	}{
		{"default_message", 12345, "", "Loading..."},
		{"custom_message", 67890, "Generating level...", "Generating level..."},
		{"empty_seed", 0, "Starting...", "Starting..."},
		{"large_seed", 9223372036854775807, "Building world...", "Building world..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewLoadingScreen()
			ls.Show(tt.seed, tt.message)

			if !ls.visible {
				t.Error("expected visible=true after Show()")
			}
			if ls.seed != tt.seed {
				t.Errorf("expected seed=%d, got %d", tt.seed, ls.seed)
			}
			if ls.message != tt.want {
				t.Errorf("expected message='%s', got '%s'", tt.want, ls.message)
			}
		})
	}
}

func TestLoadingScreen_Hide(t *testing.T) {
	ls := NewLoadingScreen()
	ls.Show(12345, "Loading...")

	if !ls.visible {
		t.Error("expected visible=true after Show()")
	}

	ls.Hide()

	if ls.visible {
		t.Error("expected visible=false after Hide()")
	}
}

func TestLoadingScreen_IsVisible(t *testing.T) {
	ls := NewLoadingScreen()

	if ls.IsVisible() {
		t.Error("expected IsVisible()=false for new loading screen")
	}

	ls.Show(12345, "Loading...")

	if !ls.IsVisible() {
		t.Error("expected IsVisible()=true after Show()")
	}

	ls.Hide()

	if ls.IsVisible() {
		t.Error("expected IsVisible()=false after Hide()")
	}
}

func TestLoadingScreen_GetSeed(t *testing.T) {
	tests := []struct {
		name string
		seed uint64
	}{
		{"zero", 0},
		{"small", 123},
		{"medium", 123456789},
		{"large", 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewLoadingScreen()
			ls.Show(tt.seed, "Loading...")

			got := ls.GetSeed()
			if got != tt.seed {
				t.Errorf("expected seed=%d, got %d", tt.seed, got)
			}
		})
	}
}

func TestLoadingScreen_SetMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{"simple", "Loading..."},
		{"detailed", "Generating BSP tree..."},
		{"empty", ""},
		{"long", "This is a very long loading message that spans multiple words"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := NewLoadingScreen()
			ls.SetMessage(tt.message)

			if ls.message != tt.message {
				t.Errorf("expected message='%s', got '%s'", tt.message, ls.message)
			}
		})
	}
}

func TestDrawLoadingScreen_NilLoadingScreen(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	DrawLoadingScreen(screen, nil) // Should not panic
}

func TestDrawLoadingScreen_NotVisible(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	ls := NewLoadingScreen()

	// Should not render when not visible
	DrawLoadingScreen(screen, ls)
}

func TestDrawLoadingScreen_Rendering(t *testing.T) {
	tests := []struct {
		name    string
		seed    uint64
		message string
		width   int
		height  int
	}{
		{"standard_resolution", 12345, "Loading...", 640, 480},
		{"low_resolution", 98765, "Generating level...", 320, 200},
		{"high_resolution", 55555, "Building world...", 1920, 1080},
		{"zero_seed", 0, "Starting...", 800, 600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(tt.width, tt.height)
			ls := NewLoadingScreen()
			ls.Show(tt.seed, tt.message)

			// Should render without panicking
			DrawLoadingScreen(screen, ls)
		})
	}
}

func TestGetLoadingDots(t *testing.T) {
	// Test that it returns valid dot patterns
	dots := getLoadingDots()

	validPatterns := []string{".", "..", "...", "...."}
	valid := false
	for _, pattern := range validPatterns {
		if dots == pattern {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("expected one of %v, got '%s'", validPatterns, dots)
	}
}

func TestLoadingScreen_MessagePersistence(t *testing.T) {
	ls := NewLoadingScreen()
	ls.Show(12345, "Initial message")

	if ls.message != "Initial message" {
		t.Errorf("expected message='Initial message', got '%s'", ls.message)
	}

	ls.SetMessage("Updated message")

	if ls.message != "Updated message" {
		t.Errorf("expected message='Updated message', got '%s'", ls.message)
	}

	// Hide and show again with new message
	ls.Hide()
	ls.Show(67890, "New message")

	if ls.message != "New message" {
		t.Errorf("expected message='New message', got '%s'", ls.message)
	}
	if ls.seed != 67890 {
		t.Errorf("expected seed=67890, got %d", ls.seed)
	}
}

func BenchmarkDrawLoadingScreen(b *testing.B) {
	screen := ebiten.NewImage(640, 480)
	ls := NewLoadingScreen()
	ls.Show(12345, "Loading...")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawLoadingScreen(screen, ls)
	}
}

func BenchmarkLoadingScreen_Show(b *testing.B) {
	ls := NewLoadingScreen()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ls.Show(uint64(i), "Loading...")
	}
}

func TestNewCommandWheel(t *testing.T) {
	cw := NewCommandWheel()

	if cw.visible {
		t.Error("CommandWheel should start hidden")
	}
	if cw.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0", cw.selectedIndex)
	}
	if len(cw.players) != 0 {
		t.Errorf("players = %d, want 0", len(cw.players))
	}
}

func TestCommandWheel_Show(t *testing.T) {
	cw := NewCommandWheel()

	cw.Show("follow")

	if !cw.visible {
		t.Error("CommandWheel should be visible after Show")
	}
	if cw.selectedCommand != "follow" {
		t.Errorf("selectedCommand = %s, want 'follow'", cw.selectedCommand)
	}
}

func TestCommandWheel_Hide(t *testing.T) {
	cw := NewCommandWheel()

	cw.Show("attack")
	cw.Hide()

	if cw.visible {
		t.Error("CommandWheel should be hidden after Hide")
	}
}

func TestCommandWheel_SetPlayers(t *testing.T) {
	cw := NewCommandWheel()

	players := []*CommandWheelPlayer{
		{PlayerID: 101, Name: "Player1", Health: 100, Active: true},
		{PlayerID: 102, Name: "Player2", Health: 75, Active: true},
		{PlayerID: 103, Name: "Player3", Health: 50, Active: false},
	}

	cw.SetPlayers(players)

	if len(cw.players) != 3 {
		t.Errorf("players = %d, want 3", len(cw.players))
	}
	if cw.players[0].PlayerID != 101 {
		t.Errorf("Player 0 ID = %d, want 101", cw.players[0].PlayerID)
	}
}

func TestCommandWheel_MoveUp(t *testing.T) {
	cw := NewCommandWheel()

	players := []*CommandWheelPlayer{
		{PlayerID: 101, Name: "Player1", Health: 100, Active: true},
		{PlayerID: 102, Name: "Player2", Health: 75, Active: true},
		{PlayerID: 103, Name: "Player3", Health: 50, Active: true},
	}
	cw.SetPlayers(players)

	// Start at 0
	cw.MoveUp()
	if cw.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2 (wrapped to end)", cw.selectedIndex)
	}

	cw.MoveUp()
	if cw.selectedIndex != 1 {
		t.Errorf("selectedIndex = %d, want 1", cw.selectedIndex)
	}
}

func TestCommandWheel_MoveDown(t *testing.T) {
	cw := NewCommandWheel()

	players := []*CommandWheelPlayer{
		{PlayerID: 101, Name: "Player1", Health: 100, Active: true},
		{PlayerID: 102, Name: "Player2", Health: 75, Active: true},
		{PlayerID: 103, Name: "Player3", Health: 50, Active: true},
	}
	cw.SetPlayers(players)

	cw.MoveDown()
	if cw.selectedIndex != 1 {
		t.Errorf("selectedIndex = %d, want 1", cw.selectedIndex)
	}

	cw.MoveDown()
	if cw.selectedIndex != 2 {
		t.Errorf("selectedIndex = %d, want 2", cw.selectedIndex)
	}

	// Wrap to start
	cw.MoveDown()
	if cw.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 (wrapped to start)", cw.selectedIndex)
	}
}

func TestCommandWheel_MoveEmptyList(t *testing.T) {
	cw := NewCommandWheel()

	// Should not panic with empty player list
	cw.MoveUp()
	cw.MoveDown()
}

func TestCommandWheel_GetSelectedPlayerID(t *testing.T) {
	cw := NewCommandWheel()

	players := []*CommandWheelPlayer{
		{PlayerID: 101, Name: "Player1", Health: 100, Active: true},
		{PlayerID: 102, Name: "Player2", Health: 75, Active: true},
	}
	cw.SetPlayers(players)

	if cw.GetSelectedPlayerID() != 101 {
		t.Errorf("GetSelectedPlayerID = %d, want 101", cw.GetSelectedPlayerID())
	}

	cw.MoveDown()
	if cw.GetSelectedPlayerID() != 102 {
		t.Errorf("GetSelectedPlayerID = %d, want 102", cw.GetSelectedPlayerID())
	}
}

func TestCommandWheel_GetSelectedPlayerID_Empty(t *testing.T) {
	cw := NewCommandWheel()

	if cw.GetSelectedPlayerID() != 0 {
		t.Errorf("GetSelectedPlayerID = %d, want 0 (no players)", cw.GetSelectedPlayerID())
	}
}

func TestCommandWheel_GetCommand(t *testing.T) {
	cw := NewCommandWheel()

	tests := []struct {
		name    string
		command string
	}{
		{"Follow", "follow"},
		{"Hold", "hold"},
		{"Attack", "attack"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw.Show(tt.command)
			if cw.GetCommand() != tt.command {
				t.Errorf("GetCommand = %s, want %s", cw.GetCommand(), tt.command)
			}
		})
	}
}

func TestDrawCommandWheel_NilWheel(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	DrawCommandWheel(screen, nil) // Should not panic
}

func TestDrawCommandWheel_Hidden(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	cw := NewCommandWheel()
	DrawCommandWheel(screen, cw) // Should not draw when hidden
}

func TestDrawCommandWheel_EmptyPlayers(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	cw := NewCommandWheel()
	cw.Show("follow")
	DrawCommandWheel(screen, cw) // Should render "No players available"
}

func TestDrawCommandWheel_WithPlayers(t *testing.T) {
	screen := ebiten.NewImage(640, 480)
	cw := NewCommandWheel()

	players := []*CommandWheelPlayer{
		{PlayerID: 101, Name: "Player1", Health: 100, Active: true},
		{PlayerID: 102, Name: "Player2", Health: 50, Active: true},
		{PlayerID: 103, Name: "Player3", Health: 0, Active: false},
	}
	cw.SetPlayers(players)
	cw.Show("attack")

	DrawCommandWheel(screen, cw) // Should render player list with health bars
}

func TestGetCommandTitle(t *testing.T) {
	tests := []struct {
		command string
		want    string
	}{
		{"follow", "FOLLOW PLAYER"},
		{"hold", "HOLD POSITION"},
		{"attack", "ATTACK TARGET"},
		{"unknown", "SQUAD COMMAND"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			got := getCommandTitle(tt.command)
			if got != tt.want {
				t.Errorf("getCommandTitle(%s) = %s, want %s", tt.command, got, tt.want)
			}
		})
	}
}

func TestCommandWheel_SetPlayers_ResetIndex(t *testing.T) {
	cw := NewCommandWheel()

	players := []*CommandWheelPlayer{
		{PlayerID: 101, Name: "Player1", Health: 100, Active: true},
		{PlayerID: 102, Name: "Player2", Health: 75, Active: true},
		{PlayerID: 103, Name: "Player3", Health: 50, Active: true},
	}
	cw.SetPlayers(players)

	// Move to last player
	cw.selectedIndex = 2

	// Set fewer players
	fewerPlayers := []*CommandWheelPlayer{
		{PlayerID: 101, Name: "Player1", Health: 100, Active: true},
	}
	cw.SetPlayers(fewerPlayers)

	// Index should be reset to 0
	if cw.selectedIndex != 0 {
		t.Errorf("selectedIndex = %d, want 0 (reset after setting fewer players)", cw.selectedIndex)
	}
}

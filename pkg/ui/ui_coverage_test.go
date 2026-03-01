package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestHUD_ShowMessage tests the ShowMessage method.
func TestHUD_ShowMessage(t *testing.T) {
	hud := NewHUD()

	hud.ShowMessage("Test message")

	if hud.Message != "Test message" {
		t.Errorf("expected Message='Test message', got %s", hud.Message)
	}
	if hud.MessageTime != 180 {
		t.Errorf("expected MessageTime=180, got %d", hud.MessageTime)
	}
}

// TestHUD_Update tests the Update method.
func TestHUD_Update(t *testing.T) {
	tests := []struct {
		name         string
		initialTime  int
		initialMsg   string
		expectedTime int
		expectedMsg  string
		updateCount  int
	}{
		{
			name:         "message_countdown",
			initialTime:  5,
			initialMsg:   "Test",
			expectedTime: 4,
			expectedMsg:  "Test",
			updateCount:  1,
		},
		{
			name:         "message_expires",
			initialTime:  1,
			initialMsg:   "Expires",
			expectedTime: 0,
			expectedMsg:  "",
			updateCount:  1,
		},
		{
			name:         "no_message",
			initialTime:  0,
			initialMsg:   "",
			expectedTime: 0,
			expectedMsg:  "",
			updateCount:  1,
		},
		{
			name:         "multiple_updates",
			initialTime:  10,
			initialMsg:   "Test",
			expectedTime: 5,
			expectedMsg:  "Test",
			updateCount:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hud := NewHUD()
			hud.Message = tt.initialMsg
			hud.MessageTime = tt.initialTime

			for i := 0; i < tt.updateCount; i++ {
				hud.Update()
			}

			if hud.MessageTime != tt.expectedTime {
				t.Errorf("expected MessageTime=%d, got %d", tt.expectedTime, hud.MessageTime)
			}
			if hud.Message != tt.expectedMsg {
				t.Errorf("expected Message='%s', got '%s'", tt.expectedMsg, hud.Message)
			}
		})
	}
}

// TestDrawShop tests the DrawShop function.
func TestDrawShop(t *testing.T) {
	tests := []struct {
		name  string
		state *ShopState
	}{
		{
			name:  "nil_state",
			state: nil,
		},
		{
			name: "empty_shop",
			state: &ShopState{
				ShopName: "Empty Shop",
				Items:    []ShopItem{},
				Credits:  100,
				Selected: 0,
			},
		},
		{
			name: "shop_with_items",
			state: &ShopState{
				ShopName: "Armory",
				Credits:  500,
				Selected: 1,
				Items: []ShopItem{
					{ID: "pistol", Name: "Pistol", Price: 100, Stock: 5},
					{ID: "shotgun", Name: "Shotgun", Price: 300, Stock: 2},
					{ID: "ammo", Name: "Ammo Pack", Price: 50, Stock: -1},
				},
			},
		},
		{
			name: "cannot_afford_item",
			state: &ShopState{
				ShopName: "Expensive Shop",
				Credits:  50,
				Selected: 0,
				Items: []ShopItem{
					{ID: "expensive", Name: "Expensive Item", Price: 1000, Stock: 1},
				},
			},
		},
		{
			name: "out_of_stock",
			state: &ShopState{
				ShopName: "Depleted Shop",
				Credits:  200,
				Selected: 0,
				Items: []ShopItem{
					{ID: "sold_out", Name: "Sold Out Item", Price: 100, Stock: 0},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(640, 480)
			DrawShop(screen, tt.state) // Should not panic
		})
	}
}

// TestDrawCrafting tests the DrawCrafting function.
func TestDrawCrafting(t *testing.T) {
	tests := []struct {
		name  string
		state *CraftingState
	}{
		{
			name:  "nil_state",
			state: nil,
		},
		{
			name: "empty_crafting",
			state: &CraftingState{
				Recipes:    []CraftingRecipe{},
				ScrapName:  "Scrap",
				ScrapAmts:  map[string]int{},
				Selected:   0,
				LastResult: "",
			},
		},
		{
			name: "crafting_with_recipes",
			state: &CraftingState{
				ScrapName: "Metal",
				ScrapAmts: map[string]int{
					"Metal": 50,
					"Wood":  30,
				},
				Selected: 1,
				Recipes: []CraftingRecipe{
					{
						ID:        "ammo",
						Name:      "Ammo",
						Inputs:    map[string]int{"Metal": 10},
						OutputQty: 50,
						CanCraft:  true,
					},
					{
						ID:        "health",
						Name:      "Health Pack",
						Inputs:    map[string]int{"Wood": 20},
						OutputQty: 1,
						CanCraft:  true,
					},
				},
				LastResult: "",
			},
		},
		{
			name: "cannot_craft",
			state: &CraftingState{
				ScrapName: "Metal",
				ScrapAmts: map[string]int{"Metal": 5},
				Selected:  0,
				Recipes: []CraftingRecipe{
					{
						ID:        "weapon",
						Name:      "Weapon",
						Inputs:    map[string]int{"Metal": 100},
						OutputQty: 1,
						CanCraft:  false,
					},
				},
				LastResult: "Not enough materials",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(640, 480)
			DrawCrafting(screen, tt.state) // Should not panic
		})
	}
}

// TestMenuManager_Select tests the Select method.
func TestMenuManager_Select(t *testing.T) {
	tests := []struct {
		name           string
		menu           MenuType
		selectedIndex  int
		expectedAction string
	}{
		{
			name:           "main_menu_new_game",
			menu:           MenuTypeMain,
			selectedIndex:  0,
			expectedAction: "new_game",
		},
		{
			name:           "main_menu_quit",
			menu:           MenuTypeMain,
			selectedIndex:  3,
			expectedAction: "quit",
		},
		{
			name:           "difficulty_menu",
			menu:           MenuTypeDifficulty,
			selectedIndex:  0,
			expectedAction: "difficulty_selected",
		},
		{
			name:           "genre_menu",
			menu:           MenuTypeGenre,
			selectedIndex:  0,
			expectedAction: "genre_selected",
		},
		{
			name:           "pause_menu_resume",
			menu:           MenuTypePause,
			selectedIndex:  0,
			expectedAction: "resume",
		},
		{
			name:           "pause_menu_shop",
			menu:           MenuTypePause,
			selectedIndex:  1,
			expectedAction: "shop",
		},
		{
			name:           "settings_menu",
			menu:           MenuTypeSettings,
			selectedIndex:  0,
			expectedAction: "settings_action",
		},
		{
			name:           "shop_menu",
			menu:           MenuTypeShop,
			selectedIndex:  0,
			expectedAction: "shop_buy",
		},
		{
			name:           "crafting_menu",
			menu:           MenuTypeCrafting,
			selectedIndex:  0,
			expectedAction: "craft_item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMenuManager()
			mm.Show(tt.menu)
			mm.selectedIndex = tt.selectedIndex

			action := mm.Select()

			if action != tt.expectedAction {
				t.Errorf("expected action=%s, got %s", tt.expectedAction, action)
			}
		})
	}
}

// TestMenuManager_Back tests the Back method.
func TestMenuManager_Back(t *testing.T) {
	tests := []struct {
		name          string
		currentMenu   MenuType
		expectedMenu  MenuType
		expectVisible bool
	}{
		{
			name:          "difficulty_to_main",
			currentMenu:   MenuTypeDifficulty,
			expectedMenu:  MenuTypeMain,
			expectVisible: true,
		},
		{
			name:          "genre_to_main",
			currentMenu:   MenuTypeGenre,
			expectedMenu:  MenuTypeMain,
			expectVisible: true,
		},
		{
			name:          "settings_to_main",
			currentMenu:   MenuTypeSettings,
			expectedMenu:  MenuTypeMain,
			expectVisible: true,
		},
		{
			name:          "pause_hide",
			currentMenu:   MenuTypePause,
			expectedMenu:  MenuTypePause,
			expectVisible: false,
		},
		{
			name:          "shop_to_pause",
			currentMenu:   MenuTypeShop,
			expectedMenu:  MenuTypePause,
			expectVisible: true,
		},
		{
			name:          "crafting_to_pause",
			currentMenu:   MenuTypeCrafting,
			expectedMenu:  MenuTypePause,
			expectVisible: true,
		},
		{
			name:          "skills_to_pause",
			currentMenu:   MenuTypeSkills,
			expectedMenu:  MenuTypePause,
			expectVisible: true,
		},
		{
			name:          "mods_to_pause",
			currentMenu:   MenuTypeMods,
			expectedMenu:  MenuTypePause,
			expectVisible: true,
		},
		{
			name:          "multiplayer_to_pause",
			currentMenu:   MenuTypeMultiplayer,
			expectedMenu:  MenuTypePause,
			expectVisible: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mm := NewMenuManager()
			mm.Show(tt.currentMenu)

			mm.Back()

			if mm.currentMenu != tt.expectedMenu {
				t.Errorf("expected menu=%v, got %v", tt.expectedMenu, mm.currentMenu)
			}
			if mm.visible != tt.expectVisible {
				t.Errorf("expected visible=%v, got %v", tt.expectVisible, mm.visible)
			}
		})
	}
}

// TestDrawTutorial tests the DrawTutorial function.
func TestDrawTutorial(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "empty_message",
			message: "",
		},
		{
			name:    "basic_message",
			message: "Press WASD to move",
		},
		{
			name:    "long_message",
			message: "Use the mouse to look around and click to shoot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(640, 480)
			DrawTutorial(screen, tt.message) // Should not panic
		})
	}
}

// TestCommandWheel_IsVisible tests the IsVisible method.
func TestCommandWheel_IsVisible(t *testing.T) {
	cw := NewCommandWheel()

	if cw.IsVisible() {
		t.Error("expected IsVisible=false initially")
	}

	cw.Show("follow")

	if !cw.IsVisible() {
		t.Error("expected IsVisible=true after Show")
	}

	cw.Hide()

	if cw.IsVisible() {
		t.Error("expected IsVisible=false after Hide")
	}
}

// TestDrawSkills tests the DrawSkills function.
func TestDrawSkills(t *testing.T) {
	tests := []struct {
		name  string
		state *SkillsState
	}{
		{
			name:  "nil_state",
			state: nil,
		},
		{
			name: "empty_skills",
			state: &SkillsState{
				Trees:       []SkillTreeState{},
				ActiveTree:  0,
				Selected:    0,
				TotalPoints: 5,
			},
		},
		{
			name: "skills_with_trees",
			state: &SkillsState{
				ActiveTree:  1,
				Selected:    2,
				TotalPoints: 10,
				Trees: []SkillTreeState{
					{
						TreeName: "Combat",
						TreeID:   "combat",
						Points:   5,
						Selected: 0,
						Nodes: []SkillNode{
							{ID: "damage", Name: "Damage +10%", Description: "Increase damage", Cost: 1, Allocated: true, Available: false},
							{ID: "crit", Name: "Critical Hit", Description: "Enable crits", Cost: 2, Allocated: false, Available: true},
						},
					},
					{
						TreeName: "Survival",
						TreeID:   "survival",
						Points:   3,
						Selected: 2,
						Nodes: []SkillNode{
							{ID: "health", Name: "Health +20", Description: "Increase max health", Cost: 1, Allocated: true, Available: false},
							{ID: "armor", Name: "Armor +15", Description: "Increase max armor", Cost: 2, Allocated: false, Available: true},
							{ID: "regen", Name: "Regeneration", Description: "Slow health regen", Cost: 3, Allocated: false, Available: false},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(800, 600)
			DrawSkills(screen, tt.state) // Should not panic
		})
	}
}

// TestDrawMods tests the DrawMods function.
func TestDrawMods(t *testing.T) {
	tests := []struct {
		name  string
		state *ModsState
	}{
		{
			name:  "nil_state",
			state: nil,
		},
		{
			name: "empty_mods",
			state: &ModsState{
				Mods:     []ModInfo{},
				ModsDir:  "/mods",
				Selected: 0,
			},
		},
		{
			name: "mods_with_items",
			state: &ModsState{
				ModsDir:  "/home/user/.violence/mods",
				Selected: 1,
				Mods: []ModInfo{
					{Name: "Enhanced Graphics", Version: "1.0", Description: "Better visuals", Author: "Alice", Enabled: true},
					{Name: "New Weapons", Version: "2.1", Description: "Additional weapons", Author: "Bob", Enabled: false},
					{Name: "Sound Pack", Version: "1.5", Description: "New sounds", Author: "Charlie", Enabled: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(640, 480)
			DrawMods(screen, tt.state) // Should not panic
		})
	}
}

// TestDrawMultiplayer tests the DrawMultiplayer function.
func TestDrawMultiplayer(t *testing.T) {
	tests := []struct {
		name  string
		state *MultiplayerState
	}{
		{
			name:  "nil_state",
			state: nil,
		},
		{
			name: "disconnected",
			state: &MultiplayerState{
				Modes: []MultiplayerMode{
					{ID: "coop", Name: "Co-op", Description: "Team up", MaxPlayers: 4},
					{ID: "deathmatch", Name: "Deathmatch", Description: "FFA", MaxPlayers: 8},
				},
				Selected:   0,
				Connected:  false,
				ServerAddr: "",
				StatusMsg:  "",
			},
		},
		{
			name: "connected",
			state: &MultiplayerState{
				Modes: []MultiplayerMode{
					{ID: "coop", Name: "Co-op", Description: "Team up", MaxPlayers: 4},
					{ID: "tdm", Name: "Team Deathmatch", Description: "Team vs Team", MaxPlayers: 8},
				},
				Selected:   1,
				Connected:  true,
				ServerAddr: "192.168.1.1:8080",
				StatusMsg:  "Ready to join",
			},
		},
		{
			name: "with_status_message",
			state: &MultiplayerState{
				Modes: []MultiplayerMode{
					{ID: "coop", Name: "Co-op", Description: "Team up", MaxPlayers: 4},
				},
				Selected:   0,
				Connected:  false,
				ServerAddr: "",
				StatusMsg:  "Server not available",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			screen := ebiten.NewImage(640, 480)
			DrawMultiplayer(screen, tt.state) // Should not panic
		})
	}
}

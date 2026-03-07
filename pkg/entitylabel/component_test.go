package entitylabel

import (
	"image/color"
	"testing"
)

func TestNewComponent(t *testing.T) {
	label := NewComponent("Test Entity")

	if label.Text != "Test Entity" {
		t.Errorf("Expected text 'Test Entity', got %q", label.Text)
	}

	if label.MaxDistance != 15.0 {
		t.Errorf("Expected MaxDistance 15.0, got %f", label.MaxDistance)
	}

	if label.Scale != 1.0 {
		t.Errorf("Expected Scale 1.0, got %f", label.Scale)
	}

	if !label.ShowBackground {
		t.Error("Expected ShowBackground to be true")
	}
}

func TestNewEnemyLabel(t *testing.T) {
	label := NewEnemyLabel("Goblin")

	if label.Text != "Goblin" {
		t.Errorf("Expected text 'Goblin', got %q", label.Text)
	}

	expectedColor := color.RGBA{R: 255, G: 100, B: 100, A: 255}
	if label.Color != expectedColor {
		t.Errorf("Expected enemy color %v, got %v", expectedColor, label.Color)
	}

	if label.MaxDistance != 12.0 {
		t.Errorf("Expected MaxDistance 12.0 for enemy, got %f", label.MaxDistance)
	}
}

func TestNewNPCLabel(t *testing.T) {
	label := NewNPCLabel("Merchant")

	expectedColor := color.RGBA{R: 100, G: 255, B: 100, A: 255}
	if label.Color != expectedColor {
		t.Errorf("Expected NPC color %v, got %v", expectedColor, label.Color)
	}
}

func TestNewLootLabel(t *testing.T) {
	label := NewLootLabel("Gold Sword")

	expectedColor := color.RGBA{R: 255, G: 255, B: 100, A: 255}
	if label.Color != expectedColor {
		t.Errorf("Expected loot color %v, got %v", expectedColor, label.Color)
	}

	if label.Scale != 0.9 {
		t.Errorf("Expected Scale 0.9 for loot, got %f", label.Scale)
	}

	if label.Priority != 0 {
		t.Errorf("Expected Priority 0 for loot, got %d", label.Priority)
	}
}

func TestNewInteractableLabel(t *testing.T) {
	label := NewInteractableLabel("Door")

	expectedColor := color.RGBA{R: 100, G: 255, B: 255, A: 255}
	if label.Color != expectedColor {
		t.Errorf("Expected interactable color %v, got %v", expectedColor, label.Color)
	}

	if !label.AlwaysVisible {
		t.Error("Expected AlwaysVisible to be true for interactables")
	}

	if label.Priority != 2 {
		t.Errorf("Expected Priority 2 for interactables, got %d", label.Priority)
	}
}

func TestNewBossLabel(t *testing.T) {
	label := NewBossLabel("Dragon Lord")

	expectedColor := color.RGBA{R: 255, G: 140, B: 0, A: 255}
	if label.Color != expectedColor {
		t.Errorf("Expected boss color %v, got %v", expectedColor, label.Color)
	}

	if label.Scale != 1.3 {
		t.Errorf("Expected Scale 1.3 for boss, got %f", label.Scale)
	}

	if !label.AlwaysVisible {
		t.Error("Expected AlwaysVisible to be true for bosses")
	}

	if label.Priority != 2 {
		t.Errorf("Expected Priority 2 for bosses, got %d", label.Priority)
	}

	if label.MaxDistance != 25.0 {
		t.Errorf("Expected MaxDistance 25.0 for boss, got %f", label.MaxDistance)
	}
}

func TestComponentType(t *testing.T) {
	label := NewComponent("Test")
	if label.Type() != "EntityLabel" {
		t.Errorf("Expected Type 'EntityLabel', got %q", label.Type())
	}
}

func TestLabelColors(t *testing.T) {
	tests := []struct {
		name     string
		label    *Component
		expected color.RGBA
	}{
		{"Enemy", NewEnemyLabel(""), color.RGBA{R: 255, G: 100, B: 100, A: 255}},
		{"NPC", NewNPCLabel(""), color.RGBA{R: 100, G: 255, B: 100, A: 255}},
		{"Loot", NewLootLabel(""), color.RGBA{R: 255, G: 255, B: 100, A: 255}},
		{"Interactable", NewInteractableLabel(""), color.RGBA{R: 100, G: 255, B: 255, A: 255}},
		{"Boss", NewBossLabel(""), color.RGBA{R: 255, G: 140, B: 0, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.label.Color != tt.expected {
				t.Errorf("%s: Expected color %v, got %v", tt.name, tt.expected, tt.label.Color)
			}
		})
	}
}

func TestVisibilityRanges(t *testing.T) {
	tests := []struct {
		name     string
		label    *Component
		expected float64
	}{
		{"Enemy", NewEnemyLabel(""), 12.0},
		{"NPC", NewNPCLabel(""), 15.0},
		{"Loot", NewLootLabel(""), 10.0},
		{"Interactable", NewInteractableLabel(""), 8.0},
		{"Boss", NewBossLabel(""), 25.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.label.MaxDistance != tt.expected {
				t.Errorf("%s: Expected MaxDistance %f, got %f", tt.name, tt.expected, tt.label.MaxDistance)
			}
		})
	}
}

func TestPriorityLevels(t *testing.T) {
	loot := NewLootLabel("Item")
	if loot.Priority != 0 {
		t.Errorf("Loot should have priority 0, got %d", loot.Priority)
	}

	enemy := NewEnemyLabel("Enemy")
	if enemy.Priority != 1 {
		t.Errorf("Enemy should have priority 1, got %d", enemy.Priority)
	}

	boss := NewBossLabel("Boss")
	if boss.Priority != 2 {
		t.Errorf("Boss should have priority 2, got %d", boss.Priority)
	}

	interactable := NewInteractableLabel("Door")
	if interactable.Priority != 2 {
		t.Errorf("Interactable should have priority 2, got %d", interactable.Priority)
	}
}

package objectivecompass

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	comp := NewComponent("test_obj", TypeMain, 100.0, 200.0)

	if comp.ID != "test_obj" {
		t.Errorf("Expected ID 'test_obj', got %q", comp.ID)
	}
	if comp.ObjType != TypeMain {
		t.Errorf("Expected TypeMain, got %d", comp.ObjType)
	}
	if comp.WorldX != 100.0 {
		t.Errorf("Expected WorldX 100.0, got %f", comp.WorldX)
	}
	if comp.WorldY != 200.0 {
		t.Errorf("Expected WorldY 200.0, got %f", comp.WorldY)
	}
	if comp.Alpha != 1.0 {
		t.Errorf("Expected Alpha 1.0, got %f", comp.Alpha)
	}
	if comp.Scale != 1.0 {
		t.Errorf("Expected Scale 1.0, got %f", comp.Scale)
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent("test", TypeBonus, 0, 0)
	if comp.Type() != "objectivecompass.Component" {
		t.Errorf("Expected 'objectivecompass.Component', got %q", comp.Type())
	}
}

func TestDefaultGenreStyles(t *testing.T) {
	styles := DefaultGenreStyles()

	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range expectedGenres {
		style, ok := styles[genre]
		if !ok {
			t.Errorf("Missing style for genre %q", genre)
			continue
		}

		if style.ArrowSize <= 0 {
			t.Errorf("Genre %q has invalid ArrowSize: %f", genre, style.ArrowSize)
		}
		if style.PulseSpeed <= 0 {
			t.Errorf("Genre %q has invalid PulseSpeed: %f", genre, style.PulseSpeed)
		}
		if style.MaxDistance <= 0 {
			t.Errorf("Genre %q has invalid MaxDistance: %f", genre, style.MaxDistance)
		}
		if style.MainColor.A == 0 {
			t.Errorf("Genre %q has transparent MainColor", genre)
		}
	}
}

func TestObjectiveTypes(t *testing.T) {
	tests := []struct {
		objType ObjectiveType
		name    string
	}{
		{TypeMain, "main"},
		{TypeBonus, "bonus"},
		{TypePOI, "poi"},
		{TypeExit, "exit"},
	}

	for _, tt := range tests {
		comp := NewComponent(tt.name, tt.objType, 0, 0)
		if comp.ObjType != tt.objType {
			t.Errorf("Expected type %d for %s, got %d", tt.objType, tt.name, comp.ObjType)
		}
	}
}

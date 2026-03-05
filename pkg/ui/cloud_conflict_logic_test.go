package ui

import (
	"errors"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/save/cloud"
)

func TestConflictDialog_NewConflictDialog(t *testing.T) {
	dialog := NewConflictDialog()

	if dialog == nil {
		t.Fatal("NewConflictDialog returned nil")
	}
	if dialog.visible {
		t.Error("Dialog should not be visible initially")
	}
	if len(dialog.options) != 4 {
		t.Errorf("Expected 4 options, got %d", len(dialog.options))
	}
	if dialog.selectedIndex != 0 {
		t.Errorf("Expected selectedIndex 0, got %d", dialog.selectedIndex)
	}
}

func TestConflictDialog_ShowHide(t *testing.T) {
	dialog := NewConflictDialog()
	localMeta := cloud.SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now().Add(-1 * time.Hour),
		Genre:     "scifi",
	}
	cloudMeta := cloud.SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now(),
		Genre:     "scifi",
	}

	resolver := func(opt ConflictOption) error {
		return nil
	}

	dialog.Show(localMeta, cloudMeta, resolver)

	if !dialog.IsVisible() {
		t.Error("Dialog should be visible after Show()")
	}
	if dialog.localMeta.SlotID != localMeta.SlotID {
		t.Error("Local metadata not set correctly")
	}
	if dialog.cloudMeta.SlotID != cloudMeta.SlotID {
		t.Error("Cloud metadata not set correctly")
	}
	if dialog.onResolve == nil {
		t.Error("onResolve callback not set")
	}

	dialog.Hide()

	if dialog.IsVisible() {
		t.Error("Dialog should not be visible after Hide()")
	}
	if dialog.onResolve != nil {
		t.Error("onResolve callback should be cleared")
	}
}

func TestConflictDialog_NavigateDown(t *testing.T) {
	dialog := NewConflictDialog()
	dialog.visible = true
	dialog.selectedIndex = 0

	dialog.navigateDown()

	if dialog.selectedIndex != 1 {
		t.Errorf("Expected index 1, got %d", dialog.selectedIndex)
	}
}

func TestConflictDialog_NavigateDownWrap(t *testing.T) {
	dialog := NewConflictDialog()
	dialog.visible = true
	dialog.selectedIndex = 3

	dialog.navigateDown()

	if dialog.selectedIndex != 0 {
		t.Errorf("Expected index 0 (wrap), got %d", dialog.selectedIndex)
	}
}

func TestConflictDialog_NavigateUp(t *testing.T) {
	dialog := NewConflictDialog()
	dialog.visible = true
	dialog.selectedIndex = 1

	dialog.navigateUp()

	if dialog.selectedIndex != 0 {
		t.Errorf("Expected index 0, got %d", dialog.selectedIndex)
	}
}

func TestConflictDialog_NavigateUpWrap(t *testing.T) {
	dialog := NewConflictDialog()
	dialog.visible = true
	dialog.selectedIndex = 0

	dialog.navigateUp()

	if dialog.selectedIndex != 3 {
		t.Errorf("Expected index 3 (wrap), got %d", dialog.selectedIndex)
	}
}

func TestConflictDialog_ResolveKeepLocal(t *testing.T) {
	dialog := NewConflictDialog()
	localMeta := cloud.SaveMetadata{SlotID: 1, Genre: "fantasy"}
	cloudMeta := cloud.SaveMetadata{SlotID: 1, Genre: "scifi"}

	var resolvedOption ConflictOption
	resolver := func(opt ConflictOption) error {
		resolvedOption = opt
		return nil
	}

	dialog.Show(localMeta, cloudMeta, resolver)
	dialog.selectedIndex = 0

	err := dialog.handleSelection()
	if err != nil {
		t.Errorf("handleSelection returned error: %v", err)
	}
	if resolvedOption != ConflictKeepLocal {
		t.Errorf("Expected ConflictKeepLocal, got %v", resolvedOption)
	}
	if dialog.IsVisible() {
		t.Error("Dialog should be hidden after resolution")
	}
}

func TestConflictDialog_ResolveKeepCloud(t *testing.T) {
	dialog := NewConflictDialog()
	localMeta := cloud.SaveMetadata{SlotID: 2}
	cloudMeta := cloud.SaveMetadata{SlotID: 2}

	var resolvedOption ConflictOption
	resolver := func(opt ConflictOption) error {
		resolvedOption = opt
		return nil
	}

	dialog.Show(localMeta, cloudMeta, resolver)
	dialog.selectedIndex = 1

	err := dialog.handleSelection()
	if err != nil {
		t.Errorf("handleSelection returned error: %v", err)
	}
	if resolvedOption != ConflictKeepCloud {
		t.Errorf("Expected ConflictKeepCloud, got %v", resolvedOption)
	}
	if dialog.IsVisible() {
		t.Error("Dialog should be hidden after resolution")
	}
}

func TestConflictDialog_ResolveKeepBoth(t *testing.T) {
	dialog := NewConflictDialog()
	localMeta := cloud.SaveMetadata{SlotID: 3}
	cloudMeta := cloud.SaveMetadata{SlotID: 3}

	var resolvedOption ConflictOption
	resolver := func(opt ConflictOption) error {
		resolvedOption = opt
		return nil
	}

	dialog.Show(localMeta, cloudMeta, resolver)
	dialog.selectedIndex = 2

	err := dialog.handleSelection()
	if err != nil {
		t.Errorf("handleSelection returned error: %v", err)
	}
	if resolvedOption != ConflictKeepBoth {
		t.Errorf("Expected ConflictKeepBoth, got %v", resolvedOption)
	}
	if dialog.IsVisible() {
		t.Error("Dialog should be hidden after resolution")
	}
}

func TestConflictDialog_ResolveCancel(t *testing.T) {
	dialog := NewConflictDialog()
	localMeta := cloud.SaveMetadata{SlotID: 4}
	cloudMeta := cloud.SaveMetadata{SlotID: 4}

	var resolvedOption ConflictOption
	resolver := func(opt ConflictOption) error {
		resolvedOption = opt
		return nil
	}

	dialog.Show(localMeta, cloudMeta, resolver)
	dialog.selectedIndex = 3

	err := dialog.handleSelection()
	if err != nil {
		t.Errorf("handleSelection returned error: %v", err)
	}
	if resolvedOption != ConflictCancel {
		t.Errorf("Expected ConflictCancel, got %v", resolvedOption)
	}
	if dialog.IsVisible() {
		t.Error("Dialog should be hidden after cancellation")
	}
}

func TestConflictDialog_ResolveError(t *testing.T) {
	dialog := NewConflictDialog()
	localMeta := cloud.SaveMetadata{SlotID: 5}
	cloudMeta := cloud.SaveMetadata{SlotID: 5}

	expectedErr := errors.New("resolution failed")
	resolver := func(opt ConflictOption) error {
		return expectedErr
	}

	dialog.Show(localMeta, cloudMeta, resolver)

	err := dialog.handleSelection()

	if err == nil {
		t.Error("Expected error from handleSelection, got nil")
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestConflictDialog_NavigateWhenNotVisible(t *testing.T) {
	dialog := NewConflictDialog()
	dialog.visible = false
	initialIndex := dialog.selectedIndex

	dialog.navigateDown()

	if dialog.selectedIndex != initialIndex {
		t.Error("selectedIndex should not change when dialog not visible")
	}
}

func TestConflictDialog_HandleSelectionNoCallback(t *testing.T) {
	dialog := NewConflictDialog()
	dialog.visible = true
	dialog.onResolve = nil

	err := dialog.handleSelection()
	if err != nil {
		t.Errorf("handleSelection returned error: %v", err)
	}
	if dialog.IsVisible() {
		t.Error("Dialog should be hidden even without callback")
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "Zero time",
			input:    time.Time{},
			expected: "Never",
		},
		{
			name:     "Valid time",
			input:    time.Date(2026, 3, 5, 12, 30, 45, 0, time.UTC),
			expected: "2026-03-05 12:30:45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTime(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

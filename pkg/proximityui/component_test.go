package proximityui

import "testing"

func TestComponentInterface(t *testing.T) {
	// Verify Component satisfies expected interface with Type() method
	comp := NewComponent()

	typeStr := comp.Type()
	if typeStr != "proximityui" {
		t.Errorf("Component.Type() = %q, want %q", typeStr, "proximityui")
	}
}

func TestDetailLevelComparisons(t *testing.T) {
	// Test detail level ordering
	if DetailNone >= DetailMinimal {
		t.Error("DetailNone should be less than DetailMinimal")
	}
	if DetailMinimal >= DetailModerate {
		t.Error("DetailMinimal should be less than DetailModerate")
	}
	if DetailModerate >= DetailFull {
		t.Error("DetailModerate should be less than DetailFull")
	}
}

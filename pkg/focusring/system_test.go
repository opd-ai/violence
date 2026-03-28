package focusring

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem()
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.enabled != true {
		t.Error("System should be enabled by default")
	}
	if len(sys.elements) != 0 {
		t.Error("System should start with no elements")
	}
	if sys.currentGenre != "fantasy" {
		t.Errorf("Default genre should be fantasy, got %s", sys.currentGenre)
	}
}

func TestAddFocusable(t *testing.T) {
	sys := NewSystem()

	elem := &FocusableElement{
		ID:       "test_button",
		X:        100,
		Y:        200,
		Width:    150,
		Height:   40,
		TabIndex: 0,
		Enabled:  true,
	}
	sys.AddFocusable(elem)

	if len(sys.elements) != 1 {
		t.Errorf("Expected 1 element, got %d", len(sys.elements))
	}
	if _, exists := sys.elementMap["test_button"]; !exists {
		t.Error("Element not found in map")
	}
	if len(sys.tabOrder) != 1 {
		t.Errorf("Expected 1 element in tab order, got %d", len(sys.tabOrder))
	}
}

func TestAddFocusable_NilElement(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(nil)
	if len(sys.elements) != 0 {
		t.Error("Adding nil element should have no effect")
	}
}

func TestAddFocusable_EmptyID(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{ID: ""})
	if len(sys.elements) != 0 {
		t.Error("Adding element with empty ID should have no effect")
	}
}

func TestRemoveFocusable(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{ID: "btn1", Enabled: true})
	sys.AddFocusable(&FocusableElement{ID: "btn2", Enabled: true})

	sys.RemoveFocusable("btn1")

	if len(sys.elements) != 1 {
		t.Errorf("Expected 1 element after removal, got %d", len(sys.elements))
	}
	if _, exists := sys.elementMap["btn1"]; exists {
		t.Error("Removed element should not exist in map")
	}
}

func TestRemoveFocusable_ClearsFocus(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{
		ID:      "btn1",
		Enabled: true,
		X:       0, Y: 0, Width: 100, Height: 40,
	})
	sys.SetFocus("btn1")

	if sys.GetFocusedID() != "btn1" {
		t.Error("Focus should be on btn1")
	}

	sys.RemoveFocusable("btn1")

	if sys.GetFocusedID() != "" {
		t.Error("Focus should be cleared after removing focused element")
	}
}

func TestClearFocusables(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{ID: "btn1", Enabled: true})
	sys.AddFocusable(&FocusableElement{ID: "btn2", Enabled: true})
	sys.SetFocus("btn1")

	sys.ClearFocusables()

	if len(sys.elements) != 0 {
		t.Error("Elements should be cleared")
	}
	if len(sys.elementMap) != 0 {
		t.Error("Element map should be cleared")
	}
	if len(sys.tabOrder) != 0 {
		t.Error("Tab order should be cleared")
	}
	if sys.GetFocusedID() != "" {
		t.Error("Focus should be cleared")
	}
}

func TestSetFocus(t *testing.T) {
	sys := NewSystem()
	elem := &FocusableElement{
		ID:      "btn1",
		X:       50,
		Y:       100,
		Width:   200,
		Height:  50,
		Enabled: true,
	}
	sys.AddFocusable(elem)

	sys.SetFocus("btn1")

	if sys.GetFocusedID() != "btn1" {
		t.Errorf("Expected focused ID btn1, got %s", sys.GetFocusedID())
	}
	if !sys.state.Visible {
		t.Error("Focus ring should be visible")
	}
	if sys.state.TargetX != 50 || sys.state.TargetY != 100 {
		t.Error("Target position should match element position")
	}
}

func TestSetFocus_InvalidID(t *testing.T) {
	sys := NewSystem()
	sys.SetFocus("nonexistent")
	if sys.GetFocusedID() != "" {
		t.Error("Setting focus to invalid ID should have no effect")
	}
}

func TestSetFocus_DisabledElement(t *testing.T) {
	sys := NewSystem()
	elem := &FocusableElement{
		ID:      "btn1",
		Enabled: false,
	}
	sys.AddFocusable(elem)
	// Force add to map (AddFocusable sets Enabled=true by default)
	elem.Enabled = false
	sys.rebuildTabOrder()

	sys.SetFocus("btn1")
	if sys.GetFocusedID() == "btn1" {
		t.Error("Should not focus disabled element")
	}
}

func TestSetFocus_CallsCallbacks(t *testing.T) {
	sys := NewSystem()

	blurCalled := false
	focusCalled := false

	elem1 := &FocusableElement{
		ID:      "btn1",
		Enabled: true,
		OnBlur:  func() { blurCalled = true },
	}
	elem2 := &FocusableElement{
		ID:      "btn2",
		Enabled: true,
		OnFocus: func() { focusCalled = true },
	}
	sys.AddFocusable(elem1)
	sys.AddFocusable(elem2)

	sys.SetFocus("btn1")
	sys.SetFocus("btn2")

	if !blurCalled {
		t.Error("OnBlur should be called when focus leaves element")
	}
	if !focusCalled {
		t.Error("OnFocus should be called when element receives focus")
	}
}

func TestClearFocus(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{
		ID:      "btn1",
		Enabled: true,
		X:       0, Y: 0, Width: 100, Height: 40,
	})
	sys.SetFocus("btn1")

	sys.ClearFocus()

	if sys.GetFocusedID() != "" {
		t.Error("Focus should be cleared")
	}
	if sys.state.Visible {
		t.Error("Focus ring should not be visible")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		genre        string
		expectedR    uint8
		expectedName string
	}{
		{"fantasy", 255, "fantasy"},     // Gold
		{"scifi", 0, "scifi"},           // Cyan
		{"horror", 180, "horror"},       // Blood red
		{"cyberpunk", 255, "cyberpunk"}, // Magenta
		{"postapoc", 255, "postapoc"},   // Orange
		{"invalid", 255, "invalid"},     // Falls back to fantasy (gold)
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewSystem()
			sys.SetGenre(tt.genre)

			if sys.currentGenre != tt.genre {
				t.Errorf("Genre should be %s, got %s", tt.genre, sys.currentGenre)
			}
			if sys.config.RingColor.R != tt.expectedR {
				t.Errorf("Ring color R should be %d, got %d", tt.expectedR, sys.config.RingColor.R)
			}
		})
	}
}

func TestSetEnabled(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{
		ID:      "btn1",
		Enabled: true,
		X:       0, Y: 0, Width: 100, Height: 40,
	})
	sys.SetFocus("btn1")

	sys.SetEnabled(false)

	if sys.enabled {
		t.Error("System should be disabled")
	}
	if sys.state.Visible {
		t.Error("Focus ring should be hidden when system disabled")
	}
}

func TestTabOrder(t *testing.T) {
	sys := NewSystem()

	// Add elements in non-order
	sys.AddFocusable(&FocusableElement{ID: "btn3", TabIndex: 2, Enabled: true})
	sys.AddFocusable(&FocusableElement{ID: "btn1", TabIndex: 0, Enabled: true})
	sys.AddFocusable(&FocusableElement{ID: "btn2", TabIndex: 1, Enabled: true})

	expected := []string{"btn1", "btn2", "btn3"}
	for i, elem := range sys.tabOrder {
		if elem.ID != expected[i] {
			t.Errorf("Tab order[%d] should be %s, got %s", i, expected[i], elem.ID)
		}
	}
}

func TestUpdate_AdvancesPulse(t *testing.T) {
	sys := NewSystem()
	// Need at least one focusable element for Update to advance pulse
	sys.AddFocusable(&FocusableElement{
		ID:      "btn1",
		Enabled: true,
		X:       0, Y: 0, Width: 100, Height: 40,
	})
	sys.SetFocus("btn1")
	initialPhase := sys.state.PulsePhase

	sys.Update()

	if sys.state.PulsePhase <= initialPhase {
		t.Error("Pulse phase should advance on update")
	}
}

func TestUpdate_DisabledDoesNothing(t *testing.T) {
	sys := NewSystem()
	sys.SetEnabled(false)
	initialPhase := sys.state.PulsePhase

	sys.Update()

	if sys.state.PulsePhase != initialPhase {
		t.Error("Disabled system should not advance pulse")
	}
}

func TestUpdateElementPosition(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{
		ID: "btn1",
		X:  0, Y: 0,
		Width: 100, Height: 40,
		Enabled: true,
	})
	sys.SetFocus("btn1")

	sys.UpdateElementPosition("btn1", 50, 60, 200, 80)

	elem := sys.elementMap["btn1"]
	if elem.X != 50 || elem.Y != 60 || elem.Width != 200 || elem.Height != 80 {
		t.Error("Element position should be updated")
	}
	if sys.state.TargetX != 50 || sys.state.TargetY != 60 {
		t.Error("Target position should be updated for focused element")
	}
}

func TestSetElementEnabled(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{
		ID:      "btn1",
		Enabled: true,
		X:       0, Y: 0, Width: 100, Height: 40,
	})
	sys.AddFocusable(&FocusableElement{
		ID:      "btn2",
		Enabled: true,
		X:       0, Y: 50, Width: 100, Height: 40,
	})
	sys.SetFocus("btn1")

	sys.SetElementEnabled("btn1", false)

	if sys.elementMap["btn1"].Enabled {
		t.Error("Element should be disabled")
	}
	// Focus should move to next element
	if sys.GetFocusedID() != "btn2" {
		t.Errorf("Focus should move to btn2, got %s", sys.GetFocusedID())
	}
}

func TestFocusableElement_Type(t *testing.T) {
	elem := &FocusableElement{ID: "test"}
	if elem.Type() != "FocusableElement" {
		t.Errorf("Type() should return FocusableElement, got %s", elem.Type())
	}
}

func TestDefaultGenrePresets(t *testing.T) {
	presets := DefaultGenrePresets()

	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range expectedGenres {
		if _, exists := presets[genre]; !exists {
			t.Errorf("Missing preset for genre: %s", genre)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.RingThickness <= 0 {
		t.Error("RingThickness should be positive")
	}
	if config.GlowRadius <= 0 {
		t.Error("GlowRadius should be positive")
	}
	if config.PulseSpeed <= 0 {
		t.Error("PulseSpeed should be positive")
	}
	if config.TransitionSpeed <= 0 {
		t.Error("TransitionSpeed should be positive")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test easeOutCubic
	if easeOutCubic(0) != 0 {
		t.Error("easeOutCubic(0) should be 0")
	}
	if easeOutCubic(1) != 1 {
		t.Error("easeOutCubic(1) should be 1")
	}

	// Test lerp
	if lerp(0, 100, 0.5) != 50 {
		t.Error("lerp should interpolate correctly")
	}

	// Test absf
	if absf(-5) != 5 {
		t.Error("absf should return absolute value")
	}

	// Test minf
	if minf(10, 5) != 5 {
		t.Error("minf should return minimum")
	}
}

func TestSpatialNavigation(t *testing.T) {
	sys := NewSystem()

	// Create a grid of buttons
	//   [1] [2]
	//   [3] [4]
	sys.AddFocusable(&FocusableElement{
		ID: "btn1", X: 0, Y: 0, Width: 50, Height: 30, Enabled: true, TabIndex: 0,
	})
	sys.AddFocusable(&FocusableElement{
		ID: "btn2", X: 60, Y: 0, Width: 50, Height: 30, Enabled: true, TabIndex: 1,
	})
	sys.AddFocusable(&FocusableElement{
		ID: "btn3", X: 0, Y: 40, Width: 50, Height: 30, Enabled: true, TabIndex: 2,
	})
	sys.AddFocusable(&FocusableElement{
		ID: "btn4", X: 60, Y: 40, Width: 50, Height: 30, Enabled: true, TabIndex: 3,
	})

	sys.SetFocus("btn1")

	// Navigate right
	sys.focusSpatial(1, 0)
	if sys.GetFocusedID() != "btn2" {
		t.Errorf("Right from btn1 should focus btn2, got %s", sys.GetFocusedID())
	}

	// Navigate down
	sys.focusSpatial(0, 1)
	if sys.GetFocusedID() != "btn4" {
		t.Errorf("Down from btn2 should focus btn4, got %s", sys.GetFocusedID())
	}

	// Navigate left
	sys.focusSpatial(-1, 0)
	if sys.GetFocusedID() != "btn3" {
		t.Errorf("Left from btn4 should focus btn3, got %s", sys.GetFocusedID())
	}

	// Navigate up
	sys.focusSpatial(0, -1)
	if sys.GetFocusedID() != "btn1" {
		t.Errorf("Up from btn3 should focus btn1, got %s", sys.GetFocusedID())
	}
}

func TestFocusNextPrevious(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{ID: "btn1", Enabled: true, TabIndex: 0})
	sys.AddFocusable(&FocusableElement{ID: "btn2", Enabled: true, TabIndex: 1})
	sys.AddFocusable(&FocusableElement{ID: "btn3", Enabled: true, TabIndex: 2})

	// Test focusNext from nothing focused
	sys.focusNext()
	if sys.GetFocusedID() != "btn1" {
		t.Errorf("Expected btn1, got %s", sys.GetFocusedID())
	}

	// Test focusNext cycles
	sys.focusNext()
	if sys.GetFocusedID() != "btn2" {
		t.Errorf("Expected btn2, got %s", sys.GetFocusedID())
	}

	sys.focusNext()
	if sys.GetFocusedID() != "btn3" {
		t.Errorf("Expected btn3, got %s", sys.GetFocusedID())
	}

	// Wrap around
	sys.focusNext()
	if sys.GetFocusedID() != "btn1" {
		t.Errorf("Expected wrap to btn1, got %s", sys.GetFocusedID())
	}

	// Test focusPrevious
	sys.focusPrevious()
	if sys.GetFocusedID() != "btn3" {
		t.Errorf("Expected btn3, got %s", sys.GetFocusedID())
	}
}

func TestFocusPrevious_FromNothing(t *testing.T) {
	sys := NewSystem()
	sys.AddFocusable(&FocusableElement{ID: "btn1", Enabled: true, TabIndex: 0})
	sys.AddFocusable(&FocusableElement{ID: "btn2", Enabled: true, TabIndex: 1})

	sys.focusPrevious()
	if sys.GetFocusedID() != "btn2" {
		t.Errorf("Expected last element btn2, got %s", sys.GetFocusedID())
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem()
	for i := 0; i < 100; i++ {
		sys.AddFocusable(&FocusableElement{
			ID:       string(rune('A'+i%26)) + string(rune('0'+i/26)),
			X:        float32(i % 10 * 50),
			Y:        float32(i / 10 * 50),
			Width:    40,
			Height:   30,
			Enabled:  true,
			TabIndex: i,
		})
	}
	sys.SetFocus("A0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update()
	}
}

func BenchmarkSpatialNavigation(b *testing.B) {
	sys := NewSystem()
	for i := 0; i < 100; i++ {
		sys.AddFocusable(&FocusableElement{
			ID:       string(rune('A'+i%26)) + string(rune('0'+i/26)),
			X:        float32(i % 10 * 50),
			Y:        float32(i / 10 * 50),
			Width:    40,
			Height:   30,
			Enabled:  true,
			TabIndex: i,
		})
	}
	sys.SetFocus("A0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.focusSpatial(1, 0)
		sys.focusSpatial(0, 1)
		sys.focusSpatial(-1, 0)
		sys.focusSpatial(0, -1)
	}
}

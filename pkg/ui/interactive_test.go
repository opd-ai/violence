package ui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewInteractiveSystem(t *testing.T) {
	sys := NewInteractiveSystem()
	if sys == nil {
		t.Fatal("NewInteractiveSystem returned nil")
	}
	if sys.buttons == nil {
		t.Error("buttons slice not initialized")
	}
	if sys.panels == nil {
		t.Error("panels slice not initialized")
	}
}

func TestAddButton(t *testing.T) {
	sys := NewInteractiveSystem()
	btn := &Button{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 30,
		Label:  "Test",
	}
	sys.AddButton(btn)

	if len(sys.buttons) != 1 {
		t.Errorf("expected 1 button, got %d", len(sys.buttons))
	}
	if btn.Transition.Duration == 0 {
		t.Error("button transition duration not set")
	}
	if btn.Transition.EaseFunc == nil {
		t.Error("button ease function not set")
	}
}

func TestButtonHoverDetection(t *testing.T) {
	sys := NewInteractiveSystem()
	btn := &Button{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 30,
		Label:  "Test",
	}
	sys.AddButton(btn)

	// Mouse outside button
	sys.Update(5, 5, false)
	if btn.State != StateIdle {
		t.Errorf("expected StateIdle, got %v", btn.State)
	}

	// Mouse inside button
	sys.Update(50, 35, false)
	if btn.State != StateHover {
		t.Errorf("expected StateHover, got %v", btn.State)
	}

	// Mouse pressed inside button
	sys.Update(50, 35, true)
	if btn.State != StatePressed {
		t.Errorf("expected StatePressed, got %v", btn.State)
	}
}

func TestButtonClickTrigger(t *testing.T) {
	sys := NewInteractiveSystem()
	clicked := false
	btn := &Button{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 30,
		Label:  "Test",
		OnClick: func() {
			clicked = true
		},
	}
	sys.AddButton(btn)

	// Press button
	sys.Update(50, 35, true)
	if clicked {
		t.Error("click should not trigger while pressed")
	}

	// Release button
	sys.Update(50, 35, false)
	if !clicked {
		t.Error("click should trigger on release")
	}
}

func TestButtonTransition(t *testing.T) {
	sys := NewInteractiveSystem()
	btn := &Button{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 30,
		Label:  "Test",
	}
	sys.AddButton(btn)

	// Initial state
	if btn.Transition.CurrentTime != 0 {
		t.Error("initial transition time should be 0")
	}

	// Trigger state change
	sys.Update(50, 35, false) // Hover
	if btn.State != StateHover {
		t.Error("state should change to hover")
	}
	if btn.Transition.CurrentTime < 0 || btn.Transition.CurrentTime > 1 {
		t.Errorf("transition should start on state change, got %d", btn.Transition.CurrentTime)
	}

	// Advance transition
	for i := 0; i < btn.Transition.Duration; i++ {
		sys.Update(50, 35, false)
	}
	if btn.Transition.CurrentTime < btn.Transition.Duration {
		t.Errorf("transition should complete, got %d/%d", btn.Transition.CurrentTime, btn.Transition.Duration)
	}
}

func TestSetFocus(t *testing.T) {
	sys := NewInteractiveSystem()
	btn1 := &Button{X: 10, Y: 20, Width: 100, Height: 30, Label: "Button1"}
	btn2 := &Button{X: 10, Y: 60, Width: 100, Height: 30, Label: "Button2"}
	sys.AddButton(btn1)
	sys.AddButton(btn2)

	// Set focus to btn1
	sys.SetFocus(btn1)
	if btn1.State != StateFocused {
		t.Error("btn1 should be focused")
	}
	if sys.focused != btn1 {
		t.Error("system focused button should be btn1")
	}

	// Change focus to btn2
	sys.SetFocus(btn2)
	if btn2.State != StateFocused {
		t.Error("btn2 should be focused")
	}
	if btn1.State == StateFocused {
		t.Error("btn1 should lose focus")
	}
	if sys.focused != btn2 {
		t.Error("system focused button should be btn2")
	}
}

func TestAddPanel(t *testing.T) {
	sys := NewInteractiveSystem()
	panel := &Panel{
		X:      10,
		Y:      10,
		Width:  200,
		Height: 150,
	}
	sys.AddPanel(panel)

	if len(sys.panels) != 1 {
		t.Errorf("expected 1 panel, got %d", len(sys.panels))
	}
	if panel.Transition.Duration == 0 {
		t.Error("panel transition duration not set")
	}
	if panel.Transition.EaseFunc == nil {
		t.Error("panel ease function not set")
	}
}

func TestShowHidePanel(t *testing.T) {
	sys := NewInteractiveSystem()
	panel := &Panel{
		X:       10,
		Y:       10,
		Width:   200,
		Height:  150,
		Visible: false,
	}
	sys.AddPanel(panel)

	// Show panel
	sys.ShowPanel(panel)
	if !panel.Visible {
		t.Error("panel should be visible")
	}
	if panel.Transition.CurrentTime != 0 {
		t.Error("transition should reset on show")
	}

	// Hide panel
	sys.HidePanel(panel)
	if panel.Visible {
		t.Error("panel should be hidden")
	}
	if panel.Transition.CurrentTime != 0 {
		t.Error("transition should reset on hide")
	}
}

func TestLerpColor(t *testing.T) {
	c1 := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	c2 := color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// Test interpolation at 0
	result := lerpColor(c1, c2, 0.0)
	if result != c1 {
		t.Errorf("lerp at 0 should equal c1, got %v", result)
	}

	// Test interpolation at 1
	result = lerpColor(c1, c2, 1.0)
	if result != c2 {
		t.Errorf("lerp at 1 should equal c2, got %v", result)
	}

	// Test interpolation at 0.5
	result = lerpColor(c1, c2, 0.5)
	_ = color.RGBA{R: 127, G: 127, B: 127, A: 255} // Expected value for reference
	if result.R < 126 || result.R > 128 {
		t.Errorf("lerp at 0.5 should be ~127, got %v", result)
	}
}

func TestEaseFunctions(t *testing.T) {
	tests := []struct {
		name string
		ease EaseFunc
	}{
		{"EaseOutCubic", EaseOutCubic},
		{"EaseInOutCubic", EaseInOutCubic},
		{"EaseOutQuad", EaseOutQuad},
		{"EaseInOutQuad", EaseInOutQuad},
		{"EaseOutElastic", EaseOutElastic},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test boundary conditions
			v0 := tt.ease(0.0)
			if v0 < -0.01 || v0 > 0.01 {
				t.Errorf("%s(0) should be ~0, got %f", tt.name, v0)
			}

			v1 := tt.ease(1.0)
			if v1 < 0.99 || v1 > 1.01 {
				t.Errorf("%s(1) should be ~1, got %f", tt.name, v1)
			}

			// Test monotonicity for non-elastic functions
			if tt.name != "EaseOutElastic" {
				v0_5 := tt.ease(0.5)
				if v0_5 < 0 || v0_5 > 1 {
					t.Errorf("%s(0.5) should be in [0,1], got %f", tt.name, v0_5)
				}
			}
		})
	}
}

func TestGetButtonColor(t *testing.T) {
	sys := NewInteractiveSystem()
	btn := &Button{
		ColorIdle:    color.RGBA{R: 50, G: 50, B: 50, A: 255},
		ColorHover:   color.RGBA{R: 100, G: 100, B: 100, A: 255},
		ColorPressed: color.RGBA{R: 150, G: 150, B: 150, A: 255},
		ColorFocused: color.RGBA{R: 200, G: 200, B: 200, A: 255},
	}

	tests := []struct {
		state    ElementState
		expected color.RGBA
	}{
		{StateIdle, btn.ColorIdle},
		{StateHover, btn.ColorHover},
		{StatePressed, btn.ColorPressed},
		{StateFocused, btn.ColorFocused},
	}

	for _, tt := range tests {
		result := sys.getButtonColor(btn, tt.state)
		if result != tt.expected {
			t.Errorf("getButtonColor(%v) = %v, want %v", tt.state, result, tt.expected)
		}
	}
}

func TestDrawButton(t *testing.T) {
	sys := NewInteractiveSystem()
	btn := &Button{
		X:            10,
		Y:            20,
		Width:        100,
		Height:       30,
		Label:        "Test",
		State:        StateIdle,
		PrevState:    StateIdle,
		ColorIdle:    color.RGBA{R: 50, G: 50, B: 50, A: 255},
		ColorHover:   color.RGBA{R: 100, G: 100, B: 100, A: 255},
		ColorPressed: color.RGBA{R: 150, G: 150, B: 150, A: 255},
		ColorFocused: color.RGBA{R: 200, G: 200, B: 200, A: 255},
		TextColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
	sys.AddButton(btn)

	// Create test screen
	screen := ebiten.NewImage(320, 240)

	// Should not panic
	sys.drawButton(screen, btn)
}

func TestDrawPanel(t *testing.T) {
	sys := NewInteractiveSystem()
	panel := &Panel{
		X:           10,
		Y:           10,
		Width:       200,
		Height:      150,
		Visible:     true,
		BgColor:     color.RGBA{R: 50, G: 50, B: 50, A: 200},
		BorderColor: color.RGBA{R: 100, G: 100, B: 100, A: 255},
	}
	sys.AddPanel(panel)

	// Create test screen
	screen := ebiten.NewImage(320, 240)

	// Should not panic
	sys.drawPanel(screen, panel)
}

func TestDraw(t *testing.T) {
	sys := NewInteractiveSystem()

	// Add button
	btn := &Button{
		X:            10,
		Y:            20,
		Width:        100,
		Height:       30,
		Label:        "Test",
		ColorIdle:    color.RGBA{R: 50, G: 50, B: 50, A: 255},
		ColorHover:   color.RGBA{R: 100, G: 100, B: 100, A: 255},
		ColorPressed: color.RGBA{R: 150, G: 150, B: 150, A: 255},
		ColorFocused: color.RGBA{R: 200, G: 200, B: 200, A: 255},
		TextColor:    color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
	sys.AddButton(btn)

	// Add panel
	panel := &Panel{
		X:           10,
		Y:           10,
		Width:       200,
		Height:      150,
		Visible:     true,
		BgColor:     color.RGBA{R: 50, G: 50, B: 50, A: 200},
		BorderColor: color.RGBA{R: 100, G: 100, B: 100, A: 255},
	}
	sys.AddPanel(panel)

	// Create test screen
	screen := ebiten.NewImage(320, 240)

	// Should not panic
	sys.Draw(screen)
}

func TestButtonClickOutsideDoesNotTrigger(t *testing.T) {
	sys := NewInteractiveSystem()
	clicked := false
	btn := &Button{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 30,
		Label:  "Test",
		OnClick: func() {
			clicked = true
		},
	}
	sys.AddButton(btn)

	// Press inside, release outside
	sys.Update(50, 35, true)
	sys.Update(5, 5, false)

	if clicked {
		t.Error("click should not trigger when released outside button")
	}
}

func TestButtonStateTransitionSequence(t *testing.T) {
	sys := NewInteractiveSystem()
	btn := &Button{
		X:      10,
		Y:      20,
		Width:  100,
		Height: 30,
		Label:  "Test",
	}
	sys.AddButton(btn)

	// Idle -> Hover
	sys.Update(5, 5, false)
	if btn.State != StateIdle {
		t.Error("should start idle")
	}

	sys.Update(50, 35, false)
	if btn.State != StateHover {
		t.Error("should transition to hover")
	}

	// Hover -> Pressed
	sys.Update(50, 35, true)
	if btn.State != StatePressed {
		t.Error("should transition to pressed")
	}

	// Pressed -> Hover (on release)
	sys.Update(50, 35, false)
	if btn.State != StateHover {
		t.Error("should transition to hover on release")
	}

	// Hover -> Idle (move away)
	sys.Update(5, 5, false)
	if btn.State != StateIdle {
		t.Error("should transition to idle")
	}
}

func TestPanelTransitionProgress(t *testing.T) {
	sys := NewInteractiveSystem()
	panel := &Panel{
		X:       10,
		Y:       10,
		Width:   200,
		Height:  150,
		Visible: false,
	}
	sys.AddPanel(panel)

	// Show panel and advance transition
	sys.ShowPanel(panel)
	initialTime := panel.Transition.CurrentTime

	// Update multiple times
	for i := 0; i < 5; i++ {
		sys.Update(0, 0, false)
	}

	if panel.Transition.CurrentTime <= initialTime {
		t.Error("panel transition should progress")
	}
}

func TestMultipleButtons(t *testing.T) {
	sys := NewInteractiveSystem()

	// Add multiple buttons
	for i := 0; i < 5; i++ {
		btn := &Button{
			X:      10,
			Y:      float32(20 + i*40),
			Width:  100,
			Height: 30,
			Label:  "Button",
		}
		sys.AddButton(btn)
	}

	if len(sys.buttons) != 5 {
		t.Errorf("expected 5 buttons, got %d", len(sys.buttons))
	}

	// Hover over second button
	sys.Update(50, 75, false)

	// Check only second button is hovered
	hoveredCount := 0
	for _, btn := range sys.buttons {
		if btn.State == StateHover {
			hoveredCount++
		}
	}
	if hoveredCount != 1 {
		t.Errorf("expected 1 hovered button, got %d", hoveredCount)
	}
}

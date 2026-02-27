package input

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.bindings == nil {
		t.Error("bindings map not initialized")
	}
	if m.gamepadButtons == nil {
		t.Error("gamepadButtons map not initialized")
	}
}

func TestDefaultBindings(t *testing.T) {
	tests := []struct {
		name     string
		action   Action
		expected ebiten.Key
	}{
		{"move forward", ActionMoveForward, ebiten.KeyW},
		{"move backward", ActionMoveBackward, ebiten.KeyS},
		{"strafe left", ActionStrafeLeft, ebiten.KeyA},
		{"strafe right", ActionStrafeRight, ebiten.KeyD},
		{"turn left", ActionTurnLeft, ebiten.KeyLeft},
		{"turn right", ActionTurnRight, ebiten.KeyRight},
		{"fire", ActionFire, ebiten.KeySpace},
		{"interact", ActionInteract, ebiten.KeyE},
		{"automap", ActionAutomap, ebiten.KeyTab},
		{"pause", ActionPause, ebiten.KeyEscape},
		{"weapon 1", ActionWeapon1, ebiten.Key1},
		{"weapon 2", ActionWeapon2, ebiten.Key2},
		{"weapon 3", ActionWeapon3, ebiten.Key3},
		{"weapon 4", ActionWeapon4, ebiten.Key4},
		{"weapon 5", ActionWeapon5, ebiten.Key5},
		{"next weapon", ActionNextWeapon, ebiten.KeyQ},
		{"prev weapon", ActionPrevWeapon, ebiten.KeyZ},
	}

	m := NewManager()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := m.GetBinding(tt.action); got != tt.expected {
				t.Errorf("GetBinding(%q) = %v, want %v", tt.action, got, tt.expected)
			}
		})
	}
}

func TestBind(t *testing.T) {
	tests := []struct {
		name     string
		action   Action
		key      ebiten.Key
		expected ebiten.Key
	}{
		{"bind move forward to up arrow", ActionMoveForward, ebiten.KeyUp, ebiten.KeyUp},
		{"bind fire to left ctrl", ActionFire, ebiten.KeyControlLeft, ebiten.KeyControlLeft},
		{"bind interact to F", ActionInteract, ebiten.KeyF, ebiten.KeyF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			m.Bind(tt.action, tt.key)
			if got := m.GetBinding(tt.action); got != tt.expected {
				t.Errorf("after Bind, GetBinding(%q) = %v, want %v", tt.action, got, tt.expected)
			}
		})
	}
}

func TestGetBinding(t *testing.T) {
	tests := []struct {
		name     string
		action   Action
		expected ebiten.Key
	}{
		{"valid action", ActionMoveForward, ebiten.KeyW},
		{"invalid action", Action("nonexistent"), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			got := m.GetBinding(tt.action)
			if tt.action == "nonexistent" {
				if got != -1 {
					t.Errorf("GetBinding(%q) = %v, want -1", tt.action, got)
				}
			} else if got != tt.expected {
				t.Errorf("GetBinding(%q) = %v, want %v", tt.action, got, tt.expected)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	// Test that Update doesn't panic and initializes mouse tracking
	m := NewManager()
	m.Update()

	// After first update, deltas should be calculated on second update
	m.prevMouseX = 100
	m.prevMouseY = 100
	m.Update()

	// The deltas will depend on actual cursor position, just verify no panic
	_, _ = m.MouseDelta()
}

func TestMouseDelta(t *testing.T) {
	m := NewManager()

	// Initial state
	dx, dy := m.MouseDelta()
	if dx != 0 || dy != 0 {
		t.Errorf("initial MouseDelta() = (%v, %v), want (0, 0)", dx, dy)
	}

	// Simulate mouse movement
	m.prevMouseX = 100
	m.prevMouseY = 100
	m.mouseDeltaX = 10
	m.mouseDeltaY = -5

	dx, dy = m.MouseDelta()
	if dx != 10 || dy != -5 {
		t.Errorf("MouseDelta() = (%v, %v), want (10, -5)", dx, dy)
	}
}

func TestGamepadAxis(t *testing.T) {
	m := NewManager()

	// When no gamepad is connected, should return 0
	value := m.GamepadAxis(0)
	if value != 0 {
		t.Errorf("GamepadAxis(0) with no gamepad = %v, want 0", value)
	}
}

func TestGamepadLeftStick(t *testing.T) {
	m := NewManager()

	// When no gamepad is connected, should return (0, 0)
	x, y := m.GamepadLeftStick()
	if x != 0 || y != 0 {
		t.Errorf("GamepadLeftStick() with no gamepad = (%v, %v), want (0, 0)", x, y)
	}
}

func TestGamepadRightStick(t *testing.T) {
	m := NewManager()

	// When no gamepad is connected, should return (0, 0)
	x, y := m.GamepadRightStick()
	if x != 0 || y != 0 {
		t.Errorf("GamepadRightStick() with no gamepad = (%v, %v), want (0, 0)", x, y)
	}
}

func TestGamepadTriggers(t *testing.T) {
	m := NewManager()

	// When no gamepad is connected, should return (0, 0)
	left, right := m.GamepadTriggers()
	if left != 0 || right != 0 {
		t.Errorf("GamepadTriggers() with no gamepad = (%v, %v), want (0, 0)", left, right)
	}
}

func TestBindGamepadButton(t *testing.T) {
	m := NewManager()

	// Test binding a gamepad button
	m.BindGamepadButton(ActionFire, ebiten.GamepadButton10)

	// Verify the binding exists
	if btn, ok := m.gamepadButtons[ActionFire]; !ok || btn != ebiten.GamepadButton10 {
		t.Errorf("BindGamepadButton failed: expected button %v, got %v", ebiten.GamepadButton10, btn)
	}
}

func TestIsPressed(t *testing.T) {
	m := NewManager()

	// IsPressed should return false for unbound actions
	// Note: We can't actually test pressed keys without Ebitengine running
	result := m.IsPressed(Action("nonexistent"))
	if result {
		t.Error("IsPressed for nonexistent action should return false")
	}
}

func TestIsJustPressed(t *testing.T) {
	m := NewManager()

	// IsJustPressed should return false for unbound actions
	// Note: We can't actually test pressed keys without Ebitengine running
	result := m.IsJustPressed(Action("nonexistent"))
	if result {
		t.Error("IsJustPressed for nonexistent action should return false")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy genre", "fantasy"},
		{"scifi genre", "scifi"},
		{"horror genre", "horror"},
		{"cyberpunk genre", "cyberpunk"},
		{"postapoc genre", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SetGenre should not panic
			SetGenre(tt.genreID)
		})
	}
}

func TestSaveBindings(t *testing.T) {
	m := NewManager()

	// Change a binding
	m.Bind(ActionFire, ebiten.KeyControlLeft)

	// SaveBindings should not panic
	// Note: This may fail if config.toml is not writable, which is acceptable for unit test
	_ = m.SaveBindings()
}

func TestLoadBindingsFromConfig(t *testing.T) {
	m := NewManager()

	// Test that loadBindingsFromConfig doesn't panic with nil config
	m.loadBindingsFromConfig()

	// Verify default bindings are still intact
	if got := m.GetBinding(ActionMoveForward); got != ebiten.KeyW {
		t.Errorf("loadBindingsFromConfig affected default binding: got %v, want %v", got, ebiten.KeyW)
	}
}

func TestGamepadNoConnection(t *testing.T) {
	m := NewManager()
	m.gamepadID = -1

	// Test all gamepad methods when no gamepad is connected
	if val := m.GamepadAxis(0); val != 0 {
		t.Errorf("GamepadAxis with no gamepad = %v, want 0", val)
	}

	x, y := m.GamepadLeftStick()
	if x != 0 || y != 0 {
		t.Errorf("GamepadLeftStick with no gamepad = (%v, %v), want (0, 0)", x, y)
	}

	x, y = m.GamepadRightStick()
	if x != 0 || y != 0 {
		t.Errorf("GamepadRightStick with no gamepad = (%v, %v), want (0, 0)", x, y)
	}

	left, right := m.GamepadTriggers()
	if left != 0 || right != 0 {
		t.Errorf("GamepadTriggers with no gamepad = (%v, %v), want (0, 0)", left, right)
	}
}

func BenchmarkUpdate(b *testing.B) {
	m := NewManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Update()
	}
}

func BenchmarkIsPressed(b *testing.B) {
	m := NewManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.IsPressed(ActionMoveForward)
	}
}

func BenchmarkGetBinding(b *testing.B) {
	m := NewManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetBinding(ActionMoveForward)
	}
}

// Package input handles keyboard, mouse, and gamepad input.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/violence/pkg/config"
)

// Action represents a game action that can be bound to input.
type Action string

// Standard game actions
const (
	ActionMoveForward  Action = "move_forward"
	ActionMoveBackward Action = "move_backward"
	ActionStrafeLeft   Action = "strafe_left"
	ActionStrafeRight  Action = "strafe_right"
	ActionTurnLeft     Action = "turn_left"
	ActionTurnRight    Action = "turn_right"
	ActionFire         Action = "fire"
	ActionInteract     Action = "interact"
	ActionAutomap      Action = "automap"
	ActionPause        Action = "pause"
	ActionWeapon1      Action = "weapon_1"
	ActionWeapon2      Action = "weapon_2"
	ActionWeapon3      Action = "weapon_3"
	ActionWeapon4      Action = "weapon_4"
	ActionWeapon5      Action = "weapon_5"
	ActionNextWeapon   Action = "next_weapon"
	ActionPrevWeapon   Action = "prev_weapon"
)

// Manager tracks input state and key bindings.
type Manager struct {
	bindings       map[Action]ebiten.Key
	gamepadButtons map[Action]ebiten.GamepadButton
	prevMouseX     int
	prevMouseY     int
	mouseDeltaX    float64
	mouseDeltaY    float64
	gamepadID      ebiten.GamepadID
}

// NewManager creates a new input manager with default bindings.
func NewManager() *Manager {
	m := &Manager{
		bindings:       make(map[Action]ebiten.Key),
		gamepadButtons: make(map[Action]ebiten.GamepadButton),
		gamepadID:      -1,
	}
	m.setDefaultBindings()
	m.loadBindingsFromConfig()
	return m
}

// setDefaultBindings configures the default WASD + mouse control scheme.
func (m *Manager) setDefaultBindings() {
	m.bindings[ActionMoveForward] = ebiten.KeyW
	m.bindings[ActionMoveBackward] = ebiten.KeyS
	m.bindings[ActionStrafeLeft] = ebiten.KeyA
	m.bindings[ActionStrafeRight] = ebiten.KeyD
	m.bindings[ActionTurnLeft] = ebiten.KeyLeft
	m.bindings[ActionTurnRight] = ebiten.KeyRight
	m.bindings[ActionFire] = ebiten.KeySpace
	m.bindings[ActionInteract] = ebiten.KeyE
	m.bindings[ActionAutomap] = ebiten.KeyTab
	m.bindings[ActionPause] = ebiten.KeyEscape
	m.bindings[ActionWeapon1] = ebiten.Key1
	m.bindings[ActionWeapon2] = ebiten.Key2
	m.bindings[ActionWeapon3] = ebiten.Key3
	m.bindings[ActionWeapon4] = ebiten.Key4
	m.bindings[ActionWeapon5] = ebiten.Key5
	m.bindings[ActionNextWeapon] = ebiten.KeyQ
	m.bindings[ActionPrevWeapon] = ebiten.KeyZ

	// Gamepad button bindings
	m.gamepadButtons[ActionFire] = ebiten.GamepadButton0       // A/Cross
	m.gamepadButtons[ActionInteract] = ebiten.GamepadButton1   // B/Circle
	m.gamepadButtons[ActionAutomap] = ebiten.GamepadButton2    // X/Square
	m.gamepadButtons[ActionPause] = ebiten.GamepadButton7      // Start
	m.gamepadButtons[ActionNextWeapon] = ebiten.GamepadButton4 // L1/LB
	m.gamepadButtons[ActionPrevWeapon] = ebiten.GamepadButton5 // R1/RB
}

// loadBindingsFromConfig loads key bindings from config file.
func (m *Manager) loadBindingsFromConfig() {
	if config.C.KeyBindings == nil {
		return
	}
	for action, keyCode := range config.C.KeyBindings {
		m.bindings[Action(action)] = ebiten.Key(keyCode)
	}
}

// Update polls input devices and refreshes state.
func (m *Manager) Update() {
	// Update mouse delta
	mx, my := ebiten.CursorPosition()
	if m.prevMouseX != 0 || m.prevMouseY != 0 {
		m.mouseDeltaX = float64(mx - m.prevMouseX)
		m.mouseDeltaY = float64(my - m.prevMouseY)
	}
	m.prevMouseX = mx
	m.prevMouseY = my

	// Find first connected gamepad
	if m.gamepadID < 0 {
		gids := ebiten.AppendGamepadIDs(nil)
		if len(gids) > 0 {
			m.gamepadID = gids[0]
		}
	}
}

// IsPressed returns true if the named action is currently pressed.
func (m *Manager) IsPressed(action Action) bool {
	// Check keyboard
	if key, ok := m.bindings[action]; ok {
		if ebiten.IsKeyPressed(key) {
			return true
		}
	}

	// Check gamepad button
	if m.gamepadID >= 0 {
		if btn, ok := m.gamepadButtons[action]; ok {
			if ebiten.IsGamepadButtonPressed(m.gamepadID, btn) {
				return true
			}
		}
	}

	return false
}

// IsJustPressed returns true if the action was pressed this frame.
func (m *Manager) IsJustPressed(action Action) bool {
	// Check keyboard
	if key, ok := m.bindings[action]; ok {
		if inpututil.IsKeyJustPressed(key) {
			return true
		}
	}

	// Check gamepad button
	if m.gamepadID >= 0 {
		if btn, ok := m.gamepadButtons[action]; ok {
			if inpututil.IsGamepadButtonJustPressed(m.gamepadID, btn) {
				return true
			}
		}
	}

	return false
}

// MouseDelta returns mouse movement since last Update.
func (m *Manager) MouseDelta() (x, y float64) {
	return m.mouseDeltaX, m.mouseDeltaY
}

// GamepadAxis returns the value of the specified gamepad axis (-1.0 to 1.0).
func (m *Manager) GamepadAxis(axis int) float64 {
	if m.gamepadID < 0 {
		return 0
	}
	return ebiten.GamepadAxisValue(m.gamepadID, axis)
}

// GamepadLeftStick returns the left analog stick values.
func (m *Manager) GamepadLeftStick() (x, y float64) {
	if m.gamepadID < 0 {
		return 0, 0
	}
	return ebiten.GamepadAxisValue(m.gamepadID, 0), ebiten.GamepadAxisValue(m.gamepadID, 1)
}

// GamepadRightStick returns the right analog stick values.
func (m *Manager) GamepadRightStick() (x, y float64) {
	if m.gamepadID < 0 {
		return 0, 0
	}
	return ebiten.GamepadAxisValue(m.gamepadID, 2), ebiten.GamepadAxisValue(m.gamepadID, 3)
}

// GamepadTriggers returns the left and right trigger values (0.0 to 1.0).
func (m *Manager) GamepadTriggers() (left, right float64) {
	if m.gamepadID < 0 {
		return 0, 0
	}
	// Standard mapping: axis 4 = LT, axis 5 = RT
	return ebiten.GamepadAxisValue(m.gamepadID, 4), ebiten.GamepadAxisValue(m.gamepadID, 5)
}

// Bind maps an action name to a key.
func (m *Manager) Bind(action Action, key ebiten.Key) {
	m.bindings[action] = key
}

// SaveBindings persists current key bindings to config.
func (m *Manager) SaveBindings() error {
	if config.C.KeyBindings == nil {
		config.C.KeyBindings = make(map[string]int)
	}
	for action, key := range m.bindings {
		config.C.KeyBindings[string(action)] = int(key)
	}
	return config.Save()
}

// BindGamepadButton maps an action to a gamepad button.
func (m *Manager) BindGamepadButton(action Action, button ebiten.GamepadButton) {
	m.gamepadButtons[action] = button
}

// GetBinding returns the key bound to the action, or -1 if not bound.
func (m *Manager) GetBinding(action Action) ebiten.Key {
	if key, ok := m.bindings[action]; ok {
		return key
	}
	return -1
}

// SetGenre configures input defaults for a genre.
func SetGenre(genreID string) {
	// Genre-specific input configurations can be added here if needed
	// For now, all genres use the same default bindings
}

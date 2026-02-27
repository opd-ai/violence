// Package input handles keyboard, mouse, and gamepad input.
package input

// Manager tracks input state and key bindings.
type Manager struct {
	bindings map[string]string
}

// NewManager creates a new input manager.
func NewManager() *Manager {
	return &Manager{bindings: make(map[string]string)}
}

// Update polls input devices and refreshes state.
func (m *Manager) Update() {}

// IsPressed returns true if the named action is currently pressed.
func (m *Manager) IsPressed(action string) bool {
	return false
}

// Bind maps an action name to a key name.
func (m *Manager) Bind(action, key string) {
	m.bindings[action] = key
}

// SetGenre configures input defaults for a genre.
func SetGenre(genreID string) {}

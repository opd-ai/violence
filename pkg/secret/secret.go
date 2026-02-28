// Package secret implements push-wall mechanics and secret discovery.
package secret

import (
	"math"
)

const (
	// Animation constants
	AnimationDuration = 1.0 // seconds
	AnimationFrames   = 16  // frames over 1 second

	// Secret states
	StateIdle      = 0
	StateAnimating = 1
	StateOpen      = 2
)

// Direction represents cardinal directions for wall sliding.
type Direction int

const (
	DirNorth Direction = iota
	DirSouth
	DirEast
	DirWest
)

// SecretWall represents a push-wall that slides open when triggered.
type SecretWall struct {
	X, Y          int       // Grid position
	Direction     Direction // Slide direction
	State         int       // Animation state
	Progress      float64   // 0.0 to 1.0 during animation
	DiscoveredBy  string    // EntityID that discovered it
	RewardSpawned bool      // Track if reward was spawned
}

// NewSecretWall creates a secret wall at the given position.
func NewSecretWall(x, y int, dir Direction) *SecretWall {
	return &SecretWall{
		X:         x,
		Y:         y,
		Direction: dir,
		State:     StateIdle,
		Progress:  0.0,
	}
}

// Trigger initiates the wall slide animation.
// Returns true if the wall was triggered, false if already open/animating.
func (sw *SecretWall) Trigger(entityID string) bool {
	if sw.State != StateIdle {
		return false
	}

	sw.State = StateAnimating
	sw.DiscoveredBy = entityID
	return true
}

// Update advances the slide animation by deltaTime seconds.
// Returns true when animation completes.
func (sw *SecretWall) Update(deltaTime float64) bool {
	if sw.State != StateAnimating {
		return false
	}

	sw.Progress += deltaTime / AnimationDuration
	if sw.Progress >= 1.0 {
		sw.Progress = 1.0
		sw.State = StateOpen
		return true
	}
	return false
}

// GetOffset returns the current wall offset for rendering.
// Uses linear interpolation for smooth movement.
func (sw *SecretWall) GetOffset() (float64, float64) {
	if sw.State == StateIdle {
		return 0.0, 0.0
	}

	// Lerp from 0 to 1 tile in the slide direction
	offset := sw.Progress

	switch sw.Direction {
	case DirNorth:
		return 0.0, -offset
	case DirSouth:
		return 0.0, offset
	case DirEast:
		return offset, 0.0
	case DirWest:
		return -offset, 0.0
	}
	return 0.0, 0.0
}

// IsOpen returns true if the wall is fully open.
func (sw *SecretWall) IsOpen() bool {
	return sw.State == StateOpen
}

// IsAnimating returns true if the wall is currently sliding.
func (sw *SecretWall) IsAnimating() bool {
	return sw.State == StateAnimating
}

// Manager tracks all secret walls in a level.
type Manager struct {
	secrets map[int]*SecretWall // Key: y*width + x
	width   int
}

// NewManager creates a secret wall manager for a level.
func NewManager(width int) *Manager {
	return &Manager{
		secrets: make(map[int]*SecretWall),
		width:   width,
	}
}

// Add registers a secret wall at the given position.
func (m *Manager) Add(x, y int, dir Direction) {
	key := y*m.width + x
	m.secrets[key] = NewSecretWall(x, y, dir)
}

// Get retrieves the secret wall at the given position.
// Returns nil if no secret exists there.
func (m *Manager) Get(x, y int) *SecretWall {
	key := y*m.width + x
	return m.secrets[key]
}

// TriggerAt attempts to trigger a secret wall at the given position.
// Returns true if a secret was triggered.
func (m *Manager) TriggerAt(x, y int, entityID string) bool {
	secret := m.Get(x, y)
	if secret == nil {
		return false
	}
	return secret.Trigger(entityID)
}

// Update advances all animating secret walls.
// Returns the count of secrets that finished animating this frame.
func (m *Manager) Update(deltaTime float64) int {
	completed := 0
	for _, secret := range m.secrets {
		if secret.Update(deltaTime) {
			completed++
		}
	}
	return completed
}

// GetAll returns all secret walls in the level.
func (m *Manager) GetAll() []*SecretWall {
	result := make([]*SecretWall, 0, len(m.secrets))
	for _, secret := range m.secrets {
		result = append(result, secret)
	}
	return result
}

// GetDiscoveredCount returns the number of discovered secrets.
func (m *Manager) GetDiscoveredCount() int {
	count := 0
	for _, secret := range m.secrets {
		if secret.State != StateIdle {
			count++
		}
	}
	return count
}

// GetTotalCount returns the total number of secrets in the level.
func (m *Manager) GetTotalCount() int {
	return len(m.secrets)
}

// EaseInOut applies easing function for smoother animation.
// Not used in current linear implementation but available for future enhancement.
func EaseInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// SmoothStep provides Hermite interpolation (smoother than linear).
func SmoothStep(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

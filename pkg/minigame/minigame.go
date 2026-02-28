// Package minigame provides interactive mini-game interfaces.
package minigame

import (
	"math/rand"
)

// MiniGame is the interface for all mini-games.
type MiniGame interface {
	Start()
	Update() bool
	GetProgress() float64
	GetAttempts() int
}

// HackGame is a hacking mini-game.
// Player must match a sequence of nodes within time limit.
type HackGame struct {
	Complete    bool
	Progress    float64
	Sequence    []int
	PlayerInput []int
	Attempts    int
	MaxAttempts int
	Difficulty  int
}

// NewHackGame creates a new hacking minigame.
func NewHackGame(difficulty int, seed int64) *HackGame {
	rng := rand.New(rand.NewSource(seed))
	sequenceLength := 3 + difficulty
	sequence := make([]int, sequenceLength)
	for i := range sequence {
		sequence[i] = rng.Intn(6) // 0-5 nodes
	}

	return &HackGame{
		Sequence:    sequence,
		PlayerInput: make([]int, 0),
		Attempts:    0,
		MaxAttempts: 3,
		Difficulty:  difficulty,
	}
}

// Start begins the hacking mini-game.
func (h *HackGame) Start() {
	h.Progress = 0
	h.PlayerInput = make([]int, 0)
	h.Attempts = 0
	h.Complete = false
}

// Input adds a player node selection.
func (h *HackGame) Input(node int) bool {
	if h.Complete {
		return false
	}

	h.PlayerInput = append(h.PlayerInput, node)

	// Check if input matches sequence so far
	idx := len(h.PlayerInput) - 1
	if idx >= len(h.Sequence) || h.PlayerInput[idx] != h.Sequence[idx] {
		// Wrong input
		h.Attempts++
		h.PlayerInput = make([]int, 0)
		h.Progress = 0
		if h.Attempts >= h.MaxAttempts {
			h.Complete = true // Failed
		}
		return false
	}

	// Correct input
	h.Progress = float64(len(h.PlayerInput)) / float64(len(h.Sequence))

	if len(h.PlayerInput) == len(h.Sequence) {
		h.Complete = true // Success
		return true
	}

	return true
}

// Update advances the hacking game; returns true when finished.
func (h *HackGame) Update() bool {
	return h.Complete
}

// GetProgress returns completion progress (0.0 to 1.0).
func (h *HackGame) GetProgress() float64 {
	return h.Progress
}

// GetAttempts returns remaining attempts.
func (h *HackGame) GetAttempts() int {
	return h.MaxAttempts - h.Attempts
}

// LockpickGame is a lockpicking mini-game.
// Player must stop a moving pin at the correct position.
type LockpickGame struct {
	Complete     bool
	Progress     float64
	Position     float64
	Target       float64
	Speed        float64
	Tolerance    float64
	Pins         int
	UnlockedPins int
	Attempts     int
	MaxAttempts  int
}

// NewLockpickGame creates a new lockpicking minigame.
func NewLockpickGame(difficulty int, seed int64) *LockpickGame {
	rng := rand.New(rand.NewSource(seed))
	pins := 2 + difficulty

	return &LockpickGame{
		Pins:         pins,
		UnlockedPins: 0,
		Speed:        0.05 + float64(difficulty)*0.02,
		Tolerance:    0.1 - float64(difficulty)*0.02,
		Target:       0.3 + rng.Float64()*0.4, // 0.3-0.7
		Position:     0,
		Attempts:     0,
		MaxAttempts:  pins * 2,
	}
}

// Start begins the lockpicking mini-game.
func (l *LockpickGame) Start() {
	l.Position = 0
	l.UnlockedPins = 0
	l.Attempts = 0
	l.Complete = false
	l.Progress = 0
}

// Advance moves the lockpick position.
func (l *LockpickGame) Advance() {
	if l.Complete {
		return
	}

	l.Position += l.Speed
	if l.Position > 1.0 {
		l.Position = 0
	}
}

// Attempt tries to unlock current pin at current position.
func (l *LockpickGame) Attempt() bool {
	if l.Complete {
		return false
	}

	l.Attempts++
	distance := l.Position - l.Target
	if distance < 0 {
		distance = -distance
	}

	if distance <= l.Tolerance {
		// Success
		l.UnlockedPins++
		l.Progress = float64(l.UnlockedPins) / float64(l.Pins)
		l.Position = 0
		if l.UnlockedPins >= l.Pins {
			l.Complete = true
			return true
		}
		return true
	}

	// Failure
	if l.Attempts >= l.MaxAttempts {
		l.Complete = true // Failed
	}
	return false
}

// Update advances the lockpicking game; returns true when finished.
func (l *LockpickGame) Update() bool {
	return l.Complete
}

// GetProgress returns completion progress (0.0 to 1.0).
func (l *LockpickGame) GetProgress() float64 {
	return l.Progress
}

// GetAttempts returns remaining attempts.
func (l *LockpickGame) GetAttempts() int {
	return l.MaxAttempts - l.Attempts
}

// SetGenre configures mini-game themes for a genre.
func SetGenre(genreID string) {}

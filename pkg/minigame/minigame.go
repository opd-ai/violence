// Package minigame provides interactive mini-game interfaces.
package minigame

// MiniGame is the interface for all mini-games.
type MiniGame interface {
	Start()
	Update() bool
}

// HackGame is a hacking mini-game.
type HackGame struct {
	Complete bool
}

// Start begins the hacking mini-game.
func (h *HackGame) Start() {}

// Update advances the hacking game; returns true when finished.
func (h *HackGame) Update() bool {
	return h.Complete
}

// LockpickGame is a lockpicking mini-game.
type LockpickGame struct {
	Complete bool
}

// Start begins the lockpicking mini-game.
func (l *LockpickGame) Start() {}

// Update advances the lockpicking game; returns true when finished.
func (l *LockpickGame) Update() bool {
	return l.Complete
}

// SetGenre configures mini-game themes for a genre.
func SetGenre(genreID string) {}

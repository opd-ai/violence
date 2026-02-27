// Package tutorial provides in-game tutorial prompts.
package tutorial

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// PromptType identifies tutorial prompt categories.
type PromptType string

const (
	PromptMovement PromptType = "movement"
	PromptShoot    PromptType = "shoot"
	PromptPickup   PromptType = "pickup"
	PromptDoor     PromptType = "door"
	PromptAutomap  PromptType = "automap"
	PromptWeapon   PromptType = "weapon"
)

// Tutorial manages tutorial prompt state with suppression persistence.
type Tutorial struct {
	Active    bool
	Current   string
	Type      PromptType
	completed map[PromptType]bool
	mu        sync.RWMutex
	savePath  string
	genreID   string
}

// NewTutorial creates a new tutorial manager.
func NewTutorial() *Tutorial {
	home, _ := os.UserHomeDir()
	savePath := filepath.Join(home, ".violence", "tutorial_state.json")

	t := &Tutorial{
		completed: make(map[PromptType]bool),
		savePath:  savePath,
		genreID:   "fantasy",
	}

	t.loadState()
	return t
}

// ShowPrompt displays a tutorial prompt if not already completed.
// Returns true if prompt was shown, false if suppressed.
func (t *Tutorial) ShowPrompt(promptType PromptType, message string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.completed[promptType] {
		return false
	}

	t.Active = true
	t.Current = message
	t.Type = promptType
	return true
}

// Complete marks the current prompt as finished and suppresses future instances.
func (t *Tutorial) Complete() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Active && t.Type != "" {
		t.completed[t.Type] = true
		t.saveState()
	}

	t.Active = false
	t.Current = ""
	t.Type = ""
}

// Dismiss hides the current tutorial prompt without marking as completed.
func (t *Tutorial) Dismiss() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Active = false
	t.Current = ""
	t.Type = ""
}

// IsCompleted checks if a prompt type has been completed.
func (t *Tutorial) IsCompleted(promptType PromptType) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.completed[promptType]
}

// Reset clears all completion state (for testing or new game).
func (t *Tutorial) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.completed = make(map[PromptType]bool)
	t.Active = false
	t.Current = ""
	t.Type = ""
	t.saveState()
}

// loadState reads completion state from disk.
func (t *Tutorial) loadState() {
	data, err := os.ReadFile(t.savePath)
	if err != nil {
		return
	}

	var state map[string]bool
	if err := json.Unmarshal(data, &state); err != nil {
		return
	}

	for k, v := range state {
		t.completed[PromptType(k)] = v
	}
}

// saveState persists completion state to disk.
func (t *Tutorial) saveState() {
	state := make(map[string]bool)
	for k, v := range t.completed {
		state[string(k)] = v
	}

	data, err := json.Marshal(state)
	if err != nil {
		return
	}

	dir := filepath.Dir(t.savePath)
	os.MkdirAll(dir, 0o755)
	os.WriteFile(t.savePath, data, 0o644)
}

// SetGenre configures tutorial content for a genre.
func SetGenre(genreID string) {}

// GetMessage returns the appropriate message for a prompt type.
func GetMessage(promptType PromptType) string {
	messages := map[PromptType]string{
		PromptMovement: "Use WASD to move, mouse to look around",
		PromptShoot:    "Left-click or Left Trigger to shoot",
		PromptPickup:   "Press E to pick up items",
		PromptDoor:     "Press E to open doors",
		PromptAutomap:  "Press TAB to toggle the automap",
		PromptWeapon:   "Press 1-7 to switch weapons",
	}
	return messages[promptType]
}

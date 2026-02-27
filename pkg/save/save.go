// Package save handles game save and load functionality.
package save

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	AutoSaveSlot = 0
	MaxSlots     = 10
)

var (
	ErrInvalidSlot = errors.New("invalid save slot")
	ErrSlotEmpty   = errors.New("save slot is empty")
)

// GameState represents the complete serializable game state.
type GameState struct {
	Version   string    `json:"version"`
	Seed      int64     `json:"seed"`
	Timestamp time.Time `json:"timestamp"`
	Player    Player    `json:"player"`
	Map       Map       `json:"map"`
	Inventory Inventory `json:"inventory"`
	Genre     string    `json:"genre"`
}

// Player holds player state.
type Player struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	DirX   float64 `json:"dir_x"`
	DirY   float64 `json:"dir_y"`
	Pitch  float64 `json:"pitch"`
	Health int     `json:"health"`
	Armor  int     `json:"armor"`
	Ammo   int     `json:"ammo"`
}

// Map holds level map data.
type Map struct {
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Tiles  [][]int `json:"tiles"`
}

// Inventory holds player inventory items.
type Inventory struct {
	Items []Item `json:"items"`
}

// Item represents an inventory item.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Qty  int    `json:"qty"`
}

// Slot represents a save-game slot with metadata.
type Slot struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Genre     string    `json:"genre"`
	Seed      int64     `json:"seed"`
	Exists    bool      `json:"exists"`
}

// getSavePath returns the platform-specific save directory path.
func getSavePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	savePath := filepath.Join(home, ".violence", "saves")
	if err := os.MkdirAll(savePath, 0o755); err != nil {
		return "", fmt.Errorf("failed to create save directory: %w", err)
	}
	return savePath, nil
}

// getSlotPath returns the file path for a given slot.
func getSlotPath(slot int) (string, error) {
	if slot < 0 || slot >= MaxSlots {
		return "", ErrInvalidSlot
	}
	savePath, err := getSavePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(savePath, fmt.Sprintf("slot_%d.json", slot)), nil
}

// Save writes game state to the given slot.
func Save(slot int, state *GameState) error {
	if slot < 0 || slot >= MaxSlots {
		return ErrInvalidSlot
	}
	if state == nil {
		return errors.New("game state is nil")
	}

	slotPath, err := getSlotPath(slot)
	if err != nil {
		return err
	}

	state.Version = "1.0"
	state.Timestamp = time.Now()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	if err := os.WriteFile(slotPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	return nil
}

// Load reads game state from the given slot.
func Load(slot int) (*GameState, error) {
	if slot < 0 || slot >= MaxSlots {
		return nil, ErrInvalidSlot
	}

	slotPath, err := getSlotPath(slot)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(slotPath); os.IsNotExist(err) {
		return nil, ErrSlotEmpty
	}

	data, err := os.ReadFile(slotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read save file: %w", err)
	}

	var state GameState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal game state: %w", err)
	}

	return &state, nil
}

// AutoSave performs an automatic save to the reserved auto-save slot.
func AutoSave(state *GameState) error {
	return Save(AutoSaveSlot, state)
}

// ListSlots returns metadata for all save slots.
func ListSlots() ([]Slot, error) {
	slots := make([]Slot, MaxSlots)
	for i := 0; i < MaxSlots; i++ {
		slots[i].ID = i
		slots[i].Exists = false

		slotPath, err := getSlotPath(i)
		if err != nil {
			continue
		}

		if _, err := os.Stat(slotPath); os.IsNotExist(err) {
			continue
		}

		state, err := Load(i)
		if err != nil {
			continue
		}

		slots[i].Exists = true
		slots[i].Timestamp = state.Timestamp
		slots[i].Genre = state.Genre
		slots[i].Seed = state.Seed
	}
	return slots, nil
}

// DeleteSlot removes a save file for the given slot.
func DeleteSlot(slot int) error {
	if slot < 0 || slot >= MaxSlots {
		return ErrInvalidSlot
	}

	slotPath, err := getSlotPath(slot)
	if err != nil {
		return err
	}

	if _, err := os.Stat(slotPath); os.IsNotExist(err) {
		return ErrSlotEmpty
	}

	if err := os.Remove(slotPath); err != nil {
		return fmt.Errorf("failed to delete save file: %w", err)
	}

	return nil
}

// SetGenre configures the save system for a genre.
func SetGenre(genreID string) {
	// Genre-specific save behavior can be configured here if needed
}

// Package save handles game save and load functionality.
package save

// Slot represents a save-game slot.
type Slot struct {
	ID   int
	Name string
	Data []byte
}

// Save writes game state to the given slot.
func Save(slot int, data []byte) error {
	return nil
}

// Load reads game state from the given slot.
func Load(slot int) ([]byte, error) {
	return nil, nil
}

// AutoSave performs an automatic save to a reserved slot.
func AutoSave(data []byte) error {
	return nil
}

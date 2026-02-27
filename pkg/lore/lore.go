// Package lore manages the in-game lore codex.
package lore

// Entry represents a single lore codex entry.
type Entry struct {
	ID    string
	Title string
	Text  string
}

// Codex holds discovered lore entries.
type Codex struct {
	Entries []Entry
}

// NewCodex creates an empty codex.
func NewCodex() *Codex {
	return &Codex{}
}

// AddEntry adds a lore entry to the codex.
func (c *Codex) AddEntry(e Entry) {
	c.Entries = append(c.Entries, e)
}

// GetEntries returns all lore entries.
func (c *Codex) GetEntries() []Entry {
	return c.Entries
}

// SetGenre configures lore content for a genre.
func SetGenre(genreID string) {}

// Package loot handles item drops and loot tables.
package loot

// Drop represents a single dropped item.
type Drop struct {
	ItemID string
	Chance float64
}

// LootTable defines a set of possible drops.
type LootTable struct {
	Drops []Drop
}

// NewLootTable creates an empty loot table.
func NewLootTable() *LootTable {
	return &LootTable{}
}

// Roll selects drops from the loot table.
func (lt *LootTable) Roll() []string {
	return nil
}

// SetGenre configures loot tables for a genre.
func SetGenre(genreID string) {}

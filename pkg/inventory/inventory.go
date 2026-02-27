// Package inventory manages the player's item inventory.
package inventory

// Item represents an inventory item.
type Item struct {
	ID   string
	Name string
	Qty  int
}

// Inventory holds the player's items.
type Inventory struct {
	Items []Item
}

// NewInventory creates an empty inventory.
func NewInventory() *Inventory {
	return &Inventory{}
}

// Add places an item into the inventory.
func (inv *Inventory) Add(item Item) {
	inv.Items = append(inv.Items, item)
}

// Remove removes an item by ID.
func (inv *Inventory) Remove(id string) {}

// Use consumes or activates an item by ID.
func (inv *Inventory) Use(id string) {}

// SetGenre configures inventory rules for a genre.
func SetGenre(genreID string) {}

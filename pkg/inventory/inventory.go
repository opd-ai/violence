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
// If item already exists, increases quantity instead of adding duplicate.
func (inv *Inventory) Add(item Item) {
	if inv.Items == nil {
		inv.Items = []Item{}
	}
	for i := range inv.Items {
		if inv.Items[i].ID == item.ID {
			inv.Items[i].Qty += item.Qty
			return
		}
	}
	inv.Items = append(inv.Items, item)
}

// Remove removes an item by ID.
// Returns true if item was removed, false if not found.
func (inv *Inventory) Remove(id string) bool {
	if inv.Items == nil {
		return false
	}
	for i := range inv.Items {
		if inv.Items[i].ID == id {
			inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			return true
		}
	}
	return false
}

// Has checks if an item exists in inventory.
func (inv *Inventory) Has(id string) bool {
	if inv.Items == nil {
		return false
	}
	for i := range inv.Items {
		if inv.Items[i].ID == id {
			return true
		}
	}
	return false
}

// Get retrieves an item by ID.
// Returns nil if not found.
func (inv *Inventory) Get(id string) *Item {
	if inv.Items == nil {
		return nil
	}
	for i := range inv.Items {
		if inv.Items[i].ID == id {
			return &inv.Items[i]
		}
	}
	return nil
}

// Consume decreases item quantity by amount.
// Returns true if consumption succeeded, false if insufficient quantity or item not found.
func (inv *Inventory) Consume(id string, amount int) bool {
	if inv.Items == nil || amount <= 0 {
		return false
	}
	for i := range inv.Items {
		if inv.Items[i].ID == id {
			if inv.Items[i].Qty < amount {
				return false
			}
			inv.Items[i].Qty -= amount
			if inv.Items[i].Qty == 0 {
				inv.Items = append(inv.Items[:i], inv.Items[i+1:]...)
			}
			return true
		}
	}
	return false
}

// Use consumes or activates an item by ID.
// Returns true if item was used successfully.
func (inv *Inventory) Use(id string) bool {
	return inv.Consume(id, 1)
}

// Count returns total number of item types in inventory.
func (inv *Inventory) Count() int {
	if inv.Items == nil {
		return 0
	}
	return len(inv.Items)
}

// SetGenre configures inventory rules for a genre.
// Currently no genre-specific inventory rules.
func SetGenre(genreID string) {}

// Package shop implements the in-game shop and armory.
package shop

// Item represents an item for sale.
type Item struct {
	ID    string
	Name  string
	Price int
}

// Shop holds items available for purchase.
type Shop struct {
	Items []Item
}

// NewShop creates a shop with the given inventory.
func NewShop(items []Item) *Shop {
	return &Shop{Items: items}
}

// Buy purchases an item by ID, deducting from the player's currency.
func (s *Shop) Buy(id string, currency *int) bool {
	return false
}

// Sell sells an item, adding to the player's currency.
func (s *Shop) Sell(id string, currency *int) bool {
	return false
}

// SetGenre configures shop inventory for a genre.
func SetGenre(genreID string) {}

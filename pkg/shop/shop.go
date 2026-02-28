// Package shop implements the in-game shop and armory.
package shop

// Item represents an item for sale.
type Item struct {
	ID    string
	Name  string
	Price int
	Stock int // -1 = unlimited
}

// Shop holds items available for purchase.
type Shop struct {
	Items   []Item
	genreID string
}

// NewShop creates a shop with the given inventory.
func NewShop(items []Item) *Shop {
	return &Shop{Items: items, genreID: "fantasy"}
}

// NewArmory creates a shop with genre-appropriate default inventory.
func NewArmory(genreID string) *Shop {
	s := &Shop{genreID: genreID}
	s.Items = s.getDefaultItems()
	return s
}

func (s *Shop) getDefaultItems() []Item {
	switch s.genreID {
	case "scifi":
		return []Item{
			{ID: "ammo_bullets", Name: "Bullet Pack", Price: 50, Stock: -1},
			{ID: "ammo_cells", Name: "Energy Cells", Price: 75, Stock: -1},
			{ID: "ammo_rockets", Name: "Missile Pod", Price: 150, Stock: -1},
			{ID: "medkit", Name: "Med-Spray", Price: 100, Stock: -1},
			{ID: "armor_vest", Name: "Combat Armor", Price: 200, Stock: 3},
		}
	case "horror":
		return []Item{
			{ID: "ammo_bullets", Name: "Old Bullets", Price: 60, Stock: -1},
			{ID: "ammo_shells", Name: "Shotgun Shells", Price: 80, Stock: -1},
			{ID: "medkit", Name: "First Aid Kit", Price: 120, Stock: -1},
			{ID: "armor_vest", Name: "Kevlar Vest", Price: 250, Stock: 2},
		}
	case "cyberpunk":
		return []Item{
			{ID: "ammo_bullets", Name: "Smart Rounds", Price: 55, Stock: -1},
			{ID: "ammo_cells", Name: "Plasma Cells", Price: 70, Stock: -1},
			{ID: "medkit", Name: "Nano-Injector", Price: 90, Stock: -1},
			{ID: "armor_vest", Name: "Ballistic Weave", Price: 220, Stock: 3},
		}
	case "postapoc":
		return []Item{
			{ID: "ammo_bullets", Name: "Salvaged Ammo", Price: 70, Stock: -1},
			{ID: "ammo_shells", Name: "Scrap Shells", Price: 90, Stock: -1},
			{ID: "medkit", Name: "Stim Pack", Price: 110, Stock: -1},
			{ID: "armor_vest", Name: "Scrap Plate", Price: 180, Stock: 4},
		}
	default: // fantasy
		return []Item{
			{ID: "ammo_arrows", Name: "Quiver of Arrows", Price: 50, Stock: -1},
			{ID: "ammo_bolts", Name: "Crossbow Bolts", Price: 75, Stock: -1},
			{ID: "medkit", Name: "Healing Potion", Price: 100, Stock: -1},
			{ID: "armor_vest", Name: "Chainmail", Price: 200, Stock: 3},
		}
	}
}

// GetItem finds an item in the shop by ID.
func (s *Shop) GetItem(id string) *Item {
	if s.Items == nil {
		return nil
	}
	for i := range s.Items {
		if s.Items[i].ID == id {
			return &s.Items[i]
		}
	}
	return nil
}

// Buy purchases an item by ID, deducting from the player's currency.
func (s *Shop) Buy(id string, currency *int) bool {
	if s.Items == nil || currency == nil {
		return false
	}
	for i := range s.Items {
		if s.Items[i].ID == id {
			// Check price
			if *currency < s.Items[i].Price {
				return false
			}
			// Check stock
			if s.Items[i].Stock == 0 {
				return false
			}
			// Purchase
			*currency -= s.Items[i].Price
			if s.Items[i].Stock > 0 {
				s.Items[i].Stock--
			}
			return true
		}
	}
	return false
}

// Sell sells an item, adding to the player's currency.
func (s *Shop) Sell(id string, currency *int) bool {
	if currency == nil {
		return false
	}
	// Selling not implemented in armory (one-way shop)
	return false
}

// SetGenre configures shop inventory for a genre.
func SetGenre(genreID string) {}

// SetGenre on instance configures genre-specific inventory.
func (s *Shop) SetGenre(genreID string) {
	s.genreID = genreID
	s.Items = s.getDefaultItems()
}

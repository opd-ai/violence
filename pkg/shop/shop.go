// Package shop implements the in-game shop and armory.
package shop

import "sync"

// Credit represents the in-game currency.
type Credit struct {
	mu     sync.RWMutex
	amount int
}

// NewCredit creates a new credit balance.
func NewCredit(initial int) *Credit {
	return &Credit{amount: initial}
}

// Add increases credit balance (for kill/secret/objective rewards).
func (c *Credit) Add(amount int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.amount += amount
}

// Deduct decreases credit balance. Returns false if insufficient.
func (c *Credit) Deduct(amount int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.amount < amount {
		return false
	}
	c.amount -= amount
	return true
}

// Get returns current credit balance.
func (c *Credit) Get() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.amount
}

// Set sets the credit balance directly.
func (c *Credit) Set(amount int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.amount = amount
}

// ItemType categorizes shop items.
type ItemType int

const (
	ItemTypeAmmo ItemType = iota
	ItemTypeWeapon
	ItemTypeUpgrade
	ItemTypeConsumable
	ItemTypeArmor
)

// Item represents an item for sale.
type Item struct {
	ID    string
	Name  string
	Type  ItemType
	Price int
	Stock int // -1 = unlimited
}

// ShopInventory manages the complete shop catalog.
type ShopInventory struct {
	Weapons     []Item
	Ammo        []Item
	Upgrades    []Item
	Consumables []Item
	Armor       []Item
}

// GetAllItems returns all items across all categories.
func (inv ShopInventory) GetAllItems() []Item {
	all := make([]Item, 0)
	all = append(all, inv.Weapons...)
	all = append(all, inv.Ammo...)
	all = append(all, inv.Upgrades...)
	all = append(all, inv.Consumables...)
	all = append(all, inv.Armor...)
	return all
}

// FindItem searches all categories for an item by ID.
func (inv ShopInventory) FindItem(id string) *Item {
	all := inv.GetAllItems()
	for i := range all {
		if all[i].ID == id {
			return &all[i]
		}
	}
	return nil
}

// Shop holds items available for purchase.
type Shop struct {
	Inventory ShopInventory
	genreID   string
	shopName  string

	// Legacy items for backward compatibility
	Items []Item
}

// NewShop creates a shop with the given inventory.
func NewShop(items []Item) *Shop {
	return &Shop{Items: items, genreID: "fantasy"}
}

// NewArmory creates a shop with genre-appropriate default inventory.
func NewArmory(genreID string) *Shop {
	s := &Shop{genreID: genreID}
	s.shopName = s.getShopName()
	s.Inventory = s.getShopInventory()

	// Populate legacy Items for backward compat
	s.Items = s.Inventory.GetAllItems()
	return s
}

func (s *Shop) getShopName() string {
	switch s.genreID {
	case "scifi":
		return "Supply Depot"
	case "horror":
		return "Black Market"
	case "cyberpunk":
		return "Corpo Shop"
	case "postapoc":
		return "Scrap Trader"
	default:
		return "Merchant Tent"
	}
}

func (s *Shop) getShopInventory() ShopInventory {
	inv := ShopInventory{}

	switch s.genreID {
	case "scifi":
		inv.Ammo = []Item{
			{ID: "ammo_bullets", Name: "Bullet Pack", Type: ItemTypeAmmo, Price: 50, Stock: -1},
			{ID: "ammo_cells", Name: "Energy Cells", Type: ItemTypeAmmo, Price: 75, Stock: -1},
			{ID: "ammo_rockets", Name: "Missile Pod", Type: ItemTypeAmmo, Price: 150, Stock: -1},
		}
		inv.Weapons = []Item{
			{ID: "weapon_rifle", Name: "Pulse Rifle", Type: ItemTypeWeapon, Price: 500, Stock: 1},
			{ID: "weapon_shotgun", Name: "Scatter Gun", Type: ItemTypeWeapon, Price: 600, Stock: 1},
		}
		inv.Upgrades = []Item{
			{ID: "upgrade_damage", Name: "Damage Enhancer", Type: ItemTypeUpgrade, Price: 300, Stock: 3},
			{ID: "upgrade_firerate", Name: "Fire Rate Module", Type: ItemTypeUpgrade, Price: 250, Stock: 3},
		}
		inv.Consumables = []Item{
			{ID: "medkit", Name: "Med-Spray", Type: ItemTypeConsumable, Price: 100, Stock: -1},
			{ID: "grenade", Name: "Plasma Grenade", Type: ItemTypeConsumable, Price: 120, Stock: 5},
		}
		inv.Armor = []Item{
			{ID: "armor_vest", Name: "Combat Armor", Type: ItemTypeArmor, Price: 200, Stock: 3},
		}

	case "horror":
		inv.Ammo = []Item{
			{ID: "ammo_bullets", Name: "Old Bullets", Type: ItemTypeAmmo, Price: 60, Stock: -1},
			{ID: "ammo_shells", Name: "Shotgun Shells", Type: ItemTypeAmmo, Price: 80, Stock: -1},
		}
		inv.Weapons = []Item{
			{ID: "weapon_pistol", Name: "Revolver", Type: ItemTypeWeapon, Price: 400, Stock: 1},
			{ID: "weapon_shotgun", Name: "Double Barrel", Type: ItemTypeWeapon, Price: 550, Stock: 1},
		}
		inv.Upgrades = []Item{
			{ID: "upgrade_damage", Name: "Silver Rounds", Type: ItemTypeUpgrade, Price: 320, Stock: 2},
			{ID: "upgrade_accuracy", Name: "Laser Sight", Type: ItemTypeUpgrade, Price: 200, Stock: 3},
		}
		inv.Consumables = []Item{
			{ID: "medkit", Name: "First Aid Kit", Type: ItemTypeConsumable, Price: 120, Stock: -1},
			{ID: "flashbang", Name: "Flash Grenade", Type: ItemTypeConsumable, Price: 100, Stock: 5},
		}
		inv.Armor = []Item{
			{ID: "armor_vest", Name: "Kevlar Vest", Type: ItemTypeArmor, Price: 250, Stock: 2},
		}

	case "cyberpunk":
		inv.Ammo = []Item{
			{ID: "ammo_bullets", Name: "Smart Rounds", Type: ItemTypeAmmo, Price: 55, Stock: -1},
			{ID: "ammo_cells", Name: "Plasma Cells", Type: ItemTypeAmmo, Price: 70, Stock: -1},
		}
		inv.Weapons = []Item{
			{ID: "weapon_smg", Name: "Smart SMG", Type: ItemTypeWeapon, Price: 450, Stock: 1},
			{ID: "weapon_rifle", Name: "Cyber Rifle", Type: ItemTypeWeapon, Price: 600, Stock: 1},
		}
		inv.Upgrades = []Item{
			{ID: "upgrade_damage", Name: "Neuro-Amp", Type: ItemTypeUpgrade, Price: 280, Stock: 3},
			{ID: "upgrade_firerate", Name: "Reflex Boost", Type: ItemTypeUpgrade, Price: 240, Stock: 3},
			{ID: "upgrade_clipsize", Name: "Mag Expander", Type: ItemTypeUpgrade, Price: 220, Stock: 3},
		}
		inv.Consumables = []Item{
			{ID: "medkit", Name: "Nano-Injector", Type: ItemTypeConsumable, Price: 90, Stock: -1},
			{ID: "grenade", Name: "EMP Grenade", Type: ItemTypeConsumable, Price: 110, Stock: 5},
		}
		inv.Armor = []Item{
			{ID: "armor_vest", Name: "Ballistic Weave", Type: ItemTypeArmor, Price: 220, Stock: 3},
		}

	case "postapoc":
		inv.Ammo = []Item{
			{ID: "ammo_bullets", Name: "Salvaged Ammo", Type: ItemTypeAmmo, Price: 70, Stock: -1},
			{ID: "ammo_shells", Name: "Scrap Shells", Type: ItemTypeAmmo, Price: 90, Stock: -1},
		}
		inv.Weapons = []Item{
			{ID: "weapon_pipe", Name: "Pipe Gun", Type: ItemTypeWeapon, Price: 350, Stock: 1},
			{ID: "weapon_shotgun", Name: "Sawed-Off", Type: ItemTypeWeapon, Price: 500, Stock: 1},
		}
		inv.Upgrades = []Item{
			{ID: "upgrade_damage", Name: "Retrofit Kit", Type: ItemTypeUpgrade, Price: 260, Stock: 3},
			{ID: "upgrade_range", Name: "Barrel Extension", Type: ItemTypeUpgrade, Price: 200, Stock: 3},
		}
		inv.Consumables = []Item{
			{ID: "medkit", Name: "Stim Pack", Type: ItemTypeConsumable, Price: 110, Stock: -1},
			{ID: "molotov", Name: "Molotov Cocktail", Type: ItemTypeConsumable, Price: 80, Stock: 5},
		}
		inv.Armor = []Item{
			{ID: "armor_vest", Name: "Scrap Plate", Type: ItemTypeArmor, Price: 180, Stock: 4},
		}

	default: // fantasy
		inv.Ammo = []Item{
			{ID: "ammo_arrows", Name: "Quiver of Arrows", Type: ItemTypeAmmo, Price: 50, Stock: -1},
			{ID: "ammo_bolts", Name: "Crossbow Bolts", Type: ItemTypeAmmo, Price: 75, Stock: -1},
		}
		inv.Weapons = []Item{
			{ID: "weapon_sword", Name: "Longsword", Type: ItemTypeWeapon, Price: 400, Stock: 1},
			{ID: "weapon_bow", Name: "Longbow", Type: ItemTypeWeapon, Price: 450, Stock: 1},
		}
		inv.Upgrades = []Item{
			{ID: "upgrade_damage", Name: "Enchantment", Type: ItemTypeUpgrade, Price: 300, Stock: 3},
			{ID: "upgrade_accuracy", Name: "Steadying Charm", Type: ItemTypeUpgrade, Price: 250, Stock: 3},
		}
		inv.Consumables = []Item{
			{ID: "medkit", Name: "Healing Potion", Type: ItemTypeConsumable, Price: 100, Stock: -1},
			{ID: "bomb", Name: "Alchemical Bomb", Type: ItemTypeConsumable, Price: 130, Stock: 5},
		}
		inv.Armor = []Item{
			{ID: "armor_vest", Name: "Chainmail", Type: ItemTypeArmor, Price: 200, Stock: 3},
		}
	}

	return inv
}

func (s *Shop) getDefaultItems() []Item {
	return s.getShopInventory().GetAllItems()
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

// Purchase buys an item using Credits. Returns true if successful.
func (s *Shop) Purchase(itemID string, credits *Credit) bool {
	if credits == nil {
		return false
	}

	// Try to find in new inventory first
	item := s.Inventory.FindItem(itemID)
	if item == nil {
		// Fall back to legacy items
		item = s.GetItem(itemID)
	}

	if item == nil {
		return false
	}

	// Check stock
	if item.Stock == 0 {
		return false
	}

	// Deduct credits
	if !credits.Deduct(item.Price) {
		return false
	}

	// Update stock in both new and legacy structures
	s.decrementStock(itemID)

	return true
}

func (s *Shop) decrementStock(itemID string) {
	// Update in inventory categories
	s.updateStockInSlice(s.Inventory.Weapons, itemID)
	s.updateStockInSlice(s.Inventory.Ammo, itemID)
	s.updateStockInSlice(s.Inventory.Upgrades, itemID)
	s.updateStockInSlice(s.Inventory.Consumables, itemID)
	s.updateStockInSlice(s.Inventory.Armor, itemID)

	// Update in legacy Items
	s.updateStockInSlice(s.Items, itemID)
}

func (s *Shop) updateStockInSlice(items []Item, itemID string) {
	for i := range items {
		if items[i].ID == itemID && items[i].Stock > 0 {
			items[i].Stock--
			break
		}
	}
}

// GetShopName returns the genre-specific shop name.
func (s *Shop) GetShopName() string {
	if s.shopName == "" {
		s.shopName = s.getShopName()
	}
	return s.shopName
}

// SetGenre configures shop inventory for a genre.
func SetGenre(genreID string) {}

// SetGenre on instance configures genre-specific inventory.
func (s *Shop) SetGenre(genreID string) {
	s.genreID = genreID
	s.shopName = s.getShopName()
	s.Inventory = s.getShopInventory()
	s.Items = s.Inventory.GetAllItems() // Update legacy items
}

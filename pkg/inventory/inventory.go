// Package inventory manages the player's item inventory with active-use items.
package inventory

import (
	"fmt"
	"sync"
)

// Entity represents a game entity that can use items (player, NPC, etc).
type Entity struct {
	Health    float64
	MaxHealth float64
	X, Y      float64
}

// ActiveItem interface defines items that can be actively used.
type ActiveItem interface {
	Use(user *Entity) error
	GetID() string
	GetName() string
}

// Grenade is an explosive throwable item.
type Grenade struct {
	ID     string
	Name   string
	Damage float64
	Radius float64
}

func (g *Grenade) Use(user *Entity) error {
	if user == nil {
		return fmt.Errorf("cannot use grenade: nil user")
	}
	// Grenade effect handled by caller (spawn projectile/explosion)
	return nil
}

func (g *Grenade) GetID() string   { return g.ID }
func (g *Grenade) GetName() string { return g.Name }

// ProximityMine is a placeable explosive trap.
type ProximityMine struct {
	ID           string
	Name         string
	Damage       float64
	TriggerRange float64
}

func (p *ProximityMine) Use(user *Entity) error {
	if user == nil {
		return fmt.Errorf("cannot use proximity mine: nil user")
	}
	// Mine placement handled by caller (spawn mine entity at user position)
	return nil
}

func (p *ProximityMine) GetID() string   { return p.ID }
func (p *ProximityMine) GetName() string { return p.Name }

// Medkit is a healing consumable.
type Medkit struct {
	ID          string
	Name        string
	HealAmount  float64
	PercentHeal float64 // If > 0, heals percentage of max health instead of fixed amount
}

func (m *Medkit) Use(user *Entity) error {
	if user == nil {
		return fmt.Errorf("cannot use medkit: nil user")
	}

	healAmount := m.HealAmount
	if m.PercentHeal > 0 {
		healAmount = user.MaxHealth * m.PercentHeal
	}

	user.Health += healAmount
	if user.Health > user.MaxHealth {
		user.Health = user.MaxHealth
	}

	return nil
}

func (m *Medkit) GetID() string   { return m.ID }
func (m *Medkit) GetName() string { return m.Name }

// Item represents an inventory item.
type Item struct {
	ID   string
	Name string
	Qty  int
}

// Inventory holds the player's items.
type Inventory struct {
	Items     []Item
	QuickSlot *QuickSlot
	mu        sync.RWMutex
}

// QuickSlot holds an active item for fast access.
type QuickSlot struct {
	Item ActiveItem
	mu   sync.RWMutex
}

// NewQuickSlot creates an empty quick slot.
func NewQuickSlot() *QuickSlot {
	return &QuickSlot{}
}

// Set assigns an active item to the quick slot.
func (q *QuickSlot) Set(item ActiveItem) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Item = item
}

// Get returns the current quick slot item.
func (q *QuickSlot) Get() ActiveItem {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.Item
}

// Clear removes the item from the quick slot.
func (q *QuickSlot) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.Item = nil
}

// IsEmpty checks if the quick slot is empty.
func (q *QuickSlot) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.Item == nil
}

// NewInventory creates an empty inventory with an empty quick slot.
func NewInventory() *Inventory {
	return &Inventory{
		QuickSlot: NewQuickSlot(),
	}
}

// Add places an item into the inventory.
// If item already exists, increases quantity instead of adding duplicate.
func (inv *Inventory) Add(item Item) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

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
	inv.mu.Lock()
	defer inv.mu.Unlock()

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
	inv.mu.RLock()
	defer inv.mu.RUnlock()

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
	inv.mu.RLock()
	defer inv.mu.RUnlock()

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
	inv.mu.Lock()
	defer inv.mu.Unlock()

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
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	if inv.Items == nil {
		return 0
	}
	return len(inv.Items)
}

// SetQuickSlot assigns an active item to the quick slot.
func (inv *Inventory) SetQuickSlot(item ActiveItem) {
	if inv.QuickSlot == nil {
		inv.QuickSlot = NewQuickSlot()
	}
	inv.QuickSlot.Set(item)
}

// GetQuickSlot returns the current quick slot item.
func (inv *Inventory) GetQuickSlot() ActiveItem {
	if inv.QuickSlot == nil {
		return nil
	}
	return inv.QuickSlot.Get()
}

// UseQuickSlot uses the item in the quick slot and consumes it from inventory.
// Returns error if quick slot is empty, item not in inventory, or use fails.
func (inv *Inventory) UseQuickSlot(user *Entity) error {
	if inv.QuickSlot == nil || inv.QuickSlot.IsEmpty() {
		return fmt.Errorf("quick slot is empty")
	}

	item := inv.QuickSlot.Get()
	if item == nil {
		return fmt.Errorf("quick slot is empty")
	}

	// Check if item is in inventory
	if !inv.Has(item.GetID()) {
		inv.QuickSlot.Clear()
		return fmt.Errorf("item %s not in inventory", item.GetID())
	}

	// Use the item
	if err := item.Use(user); err != nil {
		return fmt.Errorf("failed to use item: %w", err)
	}

	// Consume from inventory
	if !inv.Consume(item.GetID(), 1) {
		return fmt.Errorf("failed to consume item %s", item.GetID())
	}

	// Clear quick slot if item is depleted
	if !inv.Has(item.GetID()) {
		inv.QuickSlot.Clear()
	}

	return nil
}

// SetGenre configures inventory rules for a genre.
// Currently no genre-specific inventory rules.
func SetGenre(genreID string) {}

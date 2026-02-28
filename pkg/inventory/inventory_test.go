package inventory

import "testing"

func TestNewInventory(t *testing.T) {
	inv := NewInventory()
	if inv == nil {
		t.Fatal("NewInventory returned nil")
	}
	if inv.Count() != 0 {
		t.Errorf("New inventory should be empty, got count %d", inv.Count())
	}
}

func TestInventoryAdd(t *testing.T) {
	tests := []struct {
		name      string
		items     []Item
		wantQty   map[string]int
		wantCount int
	}{
		{
			name:      "add single item",
			items:     []Item{{ID: "medkit", Name: "Medkit", Qty: 1}},
			wantQty:   map[string]int{"medkit": 1},
			wantCount: 1,
		},
		{
			name:      "add multiple different items",
			items:     []Item{{ID: "medkit", Name: "Medkit", Qty: 1}, {ID: "ammo", Name: "Ammo", Qty: 10}},
			wantQty:   map[string]int{"medkit": 1, "ammo": 10},
			wantCount: 2,
		},
		{
			name:      "add same item twice stacks quantity",
			items:     []Item{{ID: "medkit", Name: "Medkit", Qty: 1}, {ID: "medkit", Name: "Medkit", Qty: 2}},
			wantQty:   map[string]int{"medkit": 3},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInventory()
			for _, item := range tt.items {
				inv.Add(item)
			}
			if inv.Count() != tt.wantCount {
				t.Errorf("Count() = %d, want %d", inv.Count(), tt.wantCount)
			}
			for id, wantQty := range tt.wantQty {
				item := inv.Get(id)
				if item == nil {
					t.Errorf("Item %s not found", id)
					continue
				}
				if item.Qty != wantQty {
					t.Errorf("Item %s quantity = %d, want %d", id, item.Qty, wantQty)
				}
			}
		})
	}
}

func TestInventoryRemove(t *testing.T) {
	tests := []struct {
		name      string
		setup     []Item
		removeID  string
		wantFound bool
		wantCount int
	}{
		{
			name:      "remove existing item",
			setup:     []Item{{ID: "medkit", Name: "Medkit", Qty: 1}},
			removeID:  "medkit",
			wantFound: true,
			wantCount: 0,
		},
		{
			name:      "remove non-existent item",
			setup:     []Item{{ID: "medkit", Name: "Medkit", Qty: 1}},
			removeID:  "ammo",
			wantFound: false,
			wantCount: 1,
		},
		{
			name:      "remove from empty inventory",
			setup:     []Item{},
			removeID:  "medkit",
			wantFound: false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInventory()
			for _, item := range tt.setup {
				inv.Add(item)
			}
			found := inv.Remove(tt.removeID)
			if found != tt.wantFound {
				t.Errorf("Remove(%s) = %v, want %v", tt.removeID, found, tt.wantFound)
			}
			if inv.Count() != tt.wantCount {
				t.Errorf("Count() = %d, want %d", inv.Count(), tt.wantCount)
			}
		})
	}
}

func TestInventoryHas(t *testing.T) {
	inv := NewInventory()
	inv.Add(Item{ID: "medkit", Name: "Medkit", Qty: 1})
	inv.Add(Item{ID: "ammo", Name: "Ammo", Qty: 10})

	if !inv.Has("medkit") {
		t.Error("Has(medkit) = false, want true")
	}
	if !inv.Has("ammo") {
		t.Error("Has(ammo) = false, want true")
	}
	if inv.Has("nonexistent") {
		t.Error("Has(nonexistent) = true, want false")
	}
}

func TestInventoryGet(t *testing.T) {
	inv := NewInventory()
	inv.Add(Item{ID: "medkit", Name: "Medkit", Qty: 1})

	item := inv.Get("medkit")
	if item == nil {
		t.Fatal("Get(medkit) returned nil")
	}
	if item.ID != "medkit" || item.Name != "Medkit" || item.Qty != 1 {
		t.Errorf("Get(medkit) = %+v, want {ID:medkit Name:Medkit Qty:1}", item)
	}

	item = inv.Get("nonexistent")
	if item != nil {
		t.Errorf("Get(nonexistent) = %+v, want nil", item)
	}
}

func TestInventoryConsume(t *testing.T) {
	tests := []struct {
		name        string
		setup       Item
		consumeQty  int
		wantSuccess bool
		wantRemain  int
		wantRemoved bool
	}{
		{
			name:        "consume partial quantity",
			setup:       Item{ID: "ammo", Name: "Ammo", Qty: 10},
			consumeQty:  3,
			wantSuccess: true,
			wantRemain:  7,
			wantRemoved: false,
		},
		{
			name:        "consume exact quantity removes item",
			setup:       Item{ID: "ammo", Name: "Ammo", Qty: 10},
			consumeQty:  10,
			wantSuccess: true,
			wantRemain:  0,
			wantRemoved: true,
		},
		{
			name:        "consume more than available fails",
			setup:       Item{ID: "ammo", Name: "Ammo", Qty: 10},
			consumeQty:  15,
			wantSuccess: false,
			wantRemain:  10,
			wantRemoved: false,
		},
		{
			name:        "consume zero fails",
			setup:       Item{ID: "ammo", Name: "Ammo", Qty: 10},
			consumeQty:  0,
			wantSuccess: false,
			wantRemain:  10,
			wantRemoved: false,
		},
		{
			name:        "consume negative fails",
			setup:       Item{ID: "ammo", Name: "Ammo", Qty: 10},
			consumeQty:  -5,
			wantSuccess: false,
			wantRemain:  10,
			wantRemoved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInventory()
			inv.Add(tt.setup)
			success := inv.Consume(tt.setup.ID, tt.consumeQty)
			if success != tt.wantSuccess {
				t.Errorf("Consume() = %v, want %v", success, tt.wantSuccess)
			}
			if tt.wantRemoved {
				if inv.Has(tt.setup.ID) {
					t.Error("Item should be removed but still exists")
				}
			} else {
				item := inv.Get(tt.setup.ID)
				if item == nil {
					t.Fatal("Item was removed unexpectedly")
				}
				if item.Qty != tt.wantRemain {
					t.Errorf("Remaining quantity = %d, want %d", item.Qty, tt.wantRemain)
				}
			}
		})
	}
}

func TestInventoryUse(t *testing.T) {
	inv := NewInventory()
	inv.Add(Item{ID: "medkit", Name: "Medkit", Qty: 3})

	if !inv.Use("medkit") {
		t.Error("Use(medkit) failed")
	}
	item := inv.Get("medkit")
	if item == nil || item.Qty != 2 {
		t.Errorf("After Use(), quantity = %v, want 2", item)
	}

	if inv.Use("nonexistent") {
		t.Error("Use(nonexistent) should fail")
	}
}

func TestInventoryNilSafety(t *testing.T) {
	inv := NewInventory()
	inv.Items = nil

	if inv.Has("test") {
		t.Error("Has() on nil Items should return false")
	}
	if inv.Get("test") != nil {
		t.Error("Get() on nil Items should return nil")
	}
	if inv.Remove("test") {
		t.Error("Remove() on nil Items should return false")
	}
	if inv.Consume("test", 1) {
		t.Error("Consume() on nil Items should return false")
	}
	if inv.Count() != 0 {
		t.Error("Count() on nil Items should return 0")
	}

	// Add should initialize slice
	inv.Add(Item{ID: "test", Name: "Test", Qty: 1})
	if !inv.Has("test") {
		t.Error("Add() should initialize nil Items slice")
	}
}

func TestSetGenre(t *testing.T) {
	// SetGenre should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("horror")
	SetGenre("cyberpunk")
	SetGenre("postapoc")
}

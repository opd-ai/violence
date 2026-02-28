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

// ActiveItem tests

func TestGrenade_Use(t *testing.T) {
	grenade := &Grenade{
		ID:     "grenade",
		Name:   "Frag Grenade",
		Damage: 50.0,
		Radius: 5.0,
	}

	user := &Entity{Health: 100, MaxHealth: 100}
	err := grenade.Use(user)
	if err != nil {
		t.Fatalf("Grenade.Use() failed: %v", err)
	}

	// Test nil user
	err = grenade.Use(nil)
	if err == nil {
		t.Fatal("Grenade.Use(nil) should return error")
	}

	if grenade.GetID() != "grenade" {
		t.Errorf("GetID() = %s, want grenade", grenade.GetID())
	}
	if grenade.GetName() != "Frag Grenade" {
		t.Errorf("GetName() = %s, want Frag Grenade", grenade.GetName())
	}
}

func TestProximityMine_Use(t *testing.T) {
	mine := &ProximityMine{
		ID:           "mine",
		Name:         "Proximity Mine",
		Damage:       75.0,
		TriggerRange: 3.0,
	}

	user := &Entity{Health: 100, MaxHealth: 100}
	err := mine.Use(user)
	if err != nil {
		t.Fatalf("ProximityMine.Use() failed: %v", err)
	}

	// Test nil user
	err = mine.Use(nil)
	if err == nil {
		t.Fatal("ProximityMine.Use(nil) should return error")
	}

	if mine.GetID() != "mine" {
		t.Errorf("GetID() = %s, want mine", mine.GetID())
	}
	if mine.GetName() != "Proximity Mine" {
		t.Errorf("GetName() = %s, want Proximity Mine", mine.GetName())
	}
}

func TestMedkit_Use(t *testing.T) {
	tests := []struct {
		name       string
		medkit     *Medkit
		userHealth float64
		maxHealth  float64
		wantHealth float64
		wantErr    bool
	}{
		{
			name:       "fixed amount healing",
			medkit:     &Medkit{ID: "medkit", Name: "Medkit", HealAmount: 25.0},
			userHealth: 50.0,
			maxHealth:  100.0,
			wantHealth: 75.0,
			wantErr:    false,
		},
		{
			name:       "percentage healing",
			medkit:     &Medkit{ID: "medkit", Name: "Medkit", PercentHeal: 0.5},
			userHealth: 50.0,
			maxHealth:  100.0,
			wantHealth: 100.0,
			wantErr:    false,
		},
		{
			name:       "healing capped at max health",
			medkit:     &Medkit{ID: "medkit", Name: "Medkit", HealAmount: 100.0},
			userHealth: 80.0,
			maxHealth:  100.0,
			wantHealth: 100.0,
			wantErr:    false,
		},
		{
			name:       "nil user returns error",
			medkit:     &Medkit{ID: "medkit", Name: "Medkit", HealAmount: 25.0},
			userHealth: 0,
			maxHealth:  0,
			wantHealth: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var user *Entity
			if !tt.wantErr {
				user = &Entity{Health: tt.userHealth, MaxHealth: tt.maxHealth}
			}

			err := tt.medkit.Use(user)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Medkit.Use() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && user.Health != tt.wantHealth {
				t.Errorf("After Use(), health = %f, want %f", user.Health, tt.wantHealth)
			}
		})
	}

	medkit := &Medkit{ID: "medkit", Name: "Medkit", HealAmount: 25.0}
	if medkit.GetID() != "medkit" {
		t.Errorf("GetID() = %s, want medkit", medkit.GetID())
	}
	if medkit.GetName() != "Medkit" {
		t.Errorf("GetName() = %s, want Medkit", medkit.GetName())
	}
}

// QuickSlot tests

func TestNewQuickSlot(t *testing.T) {
	qs := NewQuickSlot()
	if qs == nil {
		t.Fatal("NewQuickSlot returned nil")
	}
	if !qs.IsEmpty() {
		t.Error("New quick slot should be empty")
	}
}

func TestQuickSlot_SetAndGet(t *testing.T) {
	qs := NewQuickSlot()
	grenade := &Grenade{ID: "grenade", Name: "Frag Grenade"}

	qs.Set(grenade)

	if qs.IsEmpty() {
		t.Error("Quick slot should not be empty after Set")
	}

	item := qs.Get()
	if item == nil {
		t.Fatal("Get() returned nil")
	}
	if item.GetID() != "grenade" {
		t.Errorf("Get().GetID() = %s, want grenade", item.GetID())
	}
}

func TestQuickSlot_Clear(t *testing.T) {
	qs := NewQuickSlot()
	grenade := &Grenade{ID: "grenade", Name: "Frag Grenade"}

	qs.Set(grenade)
	if qs.IsEmpty() {
		t.Error("Quick slot should not be empty after Set")
	}

	qs.Clear()
	if !qs.IsEmpty() {
		t.Error("Quick slot should be empty after Clear")
	}
	if qs.Get() != nil {
		t.Error("Get() should return nil after Clear")
	}
}

// Inventory + QuickSlot integration tests

func TestInventory_SetAndGetQuickSlot(t *testing.T) {
	inv := NewInventory()
	if inv.QuickSlot == nil {
		t.Fatal("Inventory should have QuickSlot initialized")
	}

	grenade := &Grenade{ID: "grenade", Name: "Frag Grenade"}
	inv.SetQuickSlot(grenade)

	item := inv.GetQuickSlot()
	if item == nil {
		t.Fatal("GetQuickSlot() returned nil")
	}
	if item.GetID() != "grenade" {
		t.Errorf("GetQuickSlot().GetID() = %s, want grenade", item.GetID())
	}
}

func TestInventory_UseQuickSlot(t *testing.T) {
	tests := []struct {
		name           string
		setupItems     []Item
		setupQuickSlot ActiveItem
		wantErr        bool
		wantHealth     float64
		wantItemQty    int
	}{
		{
			name:           "use medkit from quick slot",
			setupItems:     []Item{{ID: "medkit", Name: "Medkit", Qty: 2}},
			setupQuickSlot: &Medkit{ID: "medkit", Name: "Medkit", HealAmount: 25.0},
			wantErr:        false,
			wantHealth:     75.0,
			wantItemQty:    1,
		},
		{
			name:           "use last medkit removes from inventory",
			setupItems:     []Item{{ID: "medkit", Name: "Medkit", Qty: 1}},
			setupQuickSlot: &Medkit{ID: "medkit", Name: "Medkit", HealAmount: 25.0},
			wantErr:        false,
			wantHealth:     75.0,
			wantItemQty:    0,
		},
		{
			name:           "use grenade from quick slot",
			setupItems:     []Item{{ID: "grenade", Name: "Grenade", Qty: 3}},
			setupQuickSlot: &Grenade{ID: "grenade", Name: "Grenade", Damage: 50.0},
			wantErr:        false,
			wantHealth:     50.0,
			wantItemQty:    2,
		},
		{
			name:           "empty quick slot returns error",
			setupItems:     []Item{{ID: "medkit", Name: "Medkit", Qty: 1}},
			setupQuickSlot: nil,
			wantErr:        true,
			wantHealth:     50.0,
			wantItemQty:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInventory()
			for _, item := range tt.setupItems {
				inv.Add(item)
			}
			if tt.setupQuickSlot != nil {
				inv.SetQuickSlot(tt.setupQuickSlot)
			}

			user := &Entity{Health: 50.0, MaxHealth: 100.0}
			err := inv.UseQuickSlot(user)

			if (err != nil) != tt.wantErr {
				t.Fatalf("UseQuickSlot() error = %v, wantErr %v", err, tt.wantErr)
			}

			if user.Health != tt.wantHealth {
				t.Errorf("After UseQuickSlot(), health = %f, want %f", user.Health, tt.wantHealth)
			}

			if len(tt.setupItems) > 0 {
				itemID := tt.setupItems[0].ID
				if tt.wantItemQty == 0 {
					if inv.Has(itemID) {
						t.Errorf("Item %s should be removed from inventory", itemID)
					}
				} else {
					item := inv.Get(itemID)
					if item == nil {
						t.Fatal("Item should still be in inventory")
					}
					if item.Qty != tt.wantItemQty {
						t.Errorf("Item quantity = %d, want %d", item.Qty, tt.wantItemQty)
					}
				}
			}
		})
	}
}

func TestInventory_UseQuickSlot_ItemNotInInventory(t *testing.T) {
	inv := NewInventory()
	// Add grenade to quick slot but not to inventory
	grenade := &Grenade{ID: "grenade", Name: "Grenade", Damage: 50.0}
	inv.SetQuickSlot(grenade)

	user := &Entity{Health: 50.0, MaxHealth: 100.0}
	err := inv.UseQuickSlot(user)
	if err == nil {
		t.Fatal("UseQuickSlot() should fail when item not in inventory")
	}

	// Quick slot should be cleared
	if !inv.QuickSlot.IsEmpty() {
		t.Error("Quick slot should be cleared when item not in inventory")
	}
}

func TestInventory_UseQuickSlot_ClearsWhenDepleted(t *testing.T) {
	inv := NewInventory()
	inv.Add(Item{ID: "medkit", Name: "Medkit", Qty: 1})
	medkit := &Medkit{ID: "medkit", Name: "Medkit", HealAmount: 25.0}
	inv.SetQuickSlot(medkit)

	user := &Entity{Health: 50.0, MaxHealth: 100.0}
	err := inv.UseQuickSlot(user)
	if err != nil {
		t.Fatalf("UseQuickSlot() failed: %v", err)
	}

	// Item should be depleted and quick slot cleared
	if inv.Has("medkit") {
		t.Error("Medkit should be removed from inventory")
	}
	if !inv.QuickSlot.IsEmpty() {
		t.Error("Quick slot should be cleared when item is depleted")
	}
}

func TestInventory_ConcurrentAccess(t *testing.T) {
	inv := NewInventory()
	inv.Add(Item{ID: "medkit", Name: "Medkit", Qty: 100})

	// Test concurrent reads and writes don't cause data races
	done := make(chan bool)

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			inv.Has("medkit")
			inv.Get("medkit")
			inv.Count()
		}
		done <- true
	}()

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			inv.Add(Item{ID: "ammo", Name: "Ammo", Qty: 1})
			inv.Consume("ammo", 1)
		}
		done <- true
	}()

	<-done
	<-done
}

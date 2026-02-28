package shop

import (
	"testing"
)

func TestNewShop(t *testing.T) {
	items := []Item{
		{ID: "item1", Name: "Test Item", Price: 100},
	}
	shop := NewShop(items)
	if shop == nil {
		t.Fatal("NewShop returned nil")
	}
	if len(shop.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(shop.Items))
	}
}

func TestNewArmory(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			shop := NewArmory(genre)
			if shop == nil {
				t.Fatal("NewArmory returned nil")
			}
			if len(shop.Items) == 0 {
				t.Error("armory should have default items")
			}
			// Verify all items have required fields
			for i, item := range shop.Items {
				if item.ID == "" {
					t.Errorf("item %d has empty ID", i)
				}
				if item.Name == "" {
					t.Errorf("item %d has empty name", i)
				}
				if item.Price <= 0 {
					t.Errorf("item %d has invalid price %d", i, item.Price)
				}
			}
		})
	}
}

func TestShop_GetItem(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "ammo", Name: "Bullets", Price: 50},
		{ID: "medkit", Name: "Health Pack", Price: 100},
	})
	item := shop.GetItem("ammo")
	if item == nil {
		t.Fatal("GetItem returned nil for valid ID")
	}
	if item.ID != "ammo" {
		t.Errorf("expected ID ammo, got %s", item.ID)
	}
	if item.Price != 50 {
		t.Errorf("expected price 50, got %d", item.Price)
	}
}

func TestShop_GetItemNotFound(t *testing.T) {
	shop := NewShop([]Item{})
	item := shop.GetItem("missing")
	if item != nil {
		t.Error("GetItem should return nil for missing item")
	}
}

func TestShop_GetItemNilItems(t *testing.T) {
	shop := &Shop{}
	item := shop.GetItem("any")
	if item != nil {
		t.Error("GetItem should return nil for nil items")
	}
}

func TestShop_Buy(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "ammo", Name: "Bullets", Price: 50, Stock: -1},
	})
	currency := 100
	success := shop.Buy("ammo", &currency)
	if !success {
		t.Error("Buy should succeed with sufficient currency")
	}
	if currency != 50 {
		t.Errorf("expected currency 50 after purchase, got %d", currency)
	}
}

func TestShop_BuyInsufficientCurrency(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "armor", Name: "Armor", Price: 200, Stock: -1},
	})
	currency := 100
	success := shop.Buy("armor", &currency)
	if success {
		t.Error("Buy should fail with insufficient currency")
	}
	if currency != 100 {
		t.Errorf("currency should not change on failed purchase, got %d", currency)
	}
}

func TestShop_BuyOutOfStock(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "rare", Name: "Rare Item", Price: 50, Stock: 0},
	})
	currency := 100
	success := shop.Buy("rare", &currency)
	if success {
		t.Error("Buy should fail when out of stock")
	}
	if currency != 100 {
		t.Errorf("currency should not change on failed purchase, got %d", currency)
	}
}

func TestShop_BuyItemNotFound(t *testing.T) {
	shop := NewShop([]Item{})
	currency := 100
	success := shop.Buy("missing", &currency)
	if success {
		t.Error("Buy should fail for missing item")
	}
	if currency != 100 {
		t.Errorf("currency should not change, got %d", currency)
	}
}

func TestShop_BuyNilCurrency(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "item", Name: "Item", Price: 50},
	})
	success := shop.Buy("item", nil)
	if success {
		t.Error("Buy should fail with nil currency")
	}
}

func TestShop_BuyNilItems(t *testing.T) {
	shop := &Shop{}
	currency := 100
	success := shop.Buy("any", &currency)
	if success {
		t.Error("Buy should fail with nil items")
	}
}

func TestShop_BuyStockDecrements(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "limited", Name: "Limited Item", Price: 50, Stock: 3},
	})
	currency := 500
	shop.Buy("limited", &currency)
	item := shop.GetItem("limited")
	if item.Stock != 2 {
		t.Errorf("expected stock 2 after purchase, got %d", item.Stock)
	}
	shop.Buy("limited", &currency)
	shop.Buy("limited", &currency)
	item = shop.GetItem("limited")
	if item.Stock != 0 {
		t.Errorf("expected stock 0 after 3 purchases, got %d", item.Stock)
	}
	// Try to buy when stock is 0
	success := shop.Buy("limited", &currency)
	if success {
		t.Error("Buy should fail when stock reaches 0")
	}
}

func TestShop_BuyUnlimitedStock(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "unlimited", Name: "Unlimited Item", Price: 10, Stock: -1},
	})
	currency := 1000
	for i := 0; i < 50; i++ {
		success := shop.Buy("unlimited", &currency)
		if !success {
			t.Fatalf("Buy failed on iteration %d", i)
		}
	}
	item := shop.GetItem("unlimited")
	if item.Stock != -1 {
		t.Errorf("unlimited stock should remain -1, got %d", item.Stock)
	}
}

func TestShop_Sell(t *testing.T) {
	shop := NewShop([]Item{})
	currency := 100
	success := shop.Sell("item", &currency)
	if success {
		t.Error("Sell is not implemented, should return false")
	}
	if currency != 100 {
		t.Error("Sell should not modify currency")
	}
}

func TestShop_SellNilCurrency(t *testing.T) {
	shop := NewShop([]Item{})
	success := shop.Sell("item", nil)
	if success {
		t.Error("Sell with nil currency should fail")
	}
}

func TestShop_SetGenre(t *testing.T) {
	shop := NewArmory("fantasy")
	shop.SetGenre("scifi")
	if shop.genreID != "scifi" {
		t.Errorf("expected genreID scifi, got %s", shop.genreID)
	}
	// Items should be regenerated
	if len(shop.Items) == 0 {
		t.Error("SetGenre should regenerate items")
	}
	// Check for scifi-specific items
	found := false
	for _, item := range shop.Items {
		if item.Name == "Energy Cells" || item.Name == "Med-Spray" {
			found = true
			break
		}
	}
	if !found {
		t.Error("SetGenre should use genre-specific item names")
	}
}

func TestShop_GenreDistinctItems(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	itemNames := make(map[string]map[string]bool)
	for _, genre := range genres {
		itemNames[genre] = make(map[string]bool)
		shop := NewArmory(genre)
		for _, item := range shop.Items {
			itemNames[genre][item.Name] = true
		}
	}
	// Check that each genre has unique names
	for i, g1 := range genres {
		for j, g2 := range genres {
			if i >= j {
				continue
			}
			// Count overlap
			overlap := 0
			for name := range itemNames[g1] {
				if itemNames[g2][name] {
					overlap++
				}
			}
			// Should have minimal overlap (maybe medkit variants)
			if overlap > 1 {
				t.Errorf("genres %s and %s have too much overlap: %d items", g1, g2, overlap)
			}
		}
	}
}

func TestShop_DefaultGenre(t *testing.T) {
	shop := NewArmory("unknown_genre")
	// Should default to fantasy
	found := false
	for _, item := range shop.Items {
		if item.Name == "Healing Potion" || item.Name == "Chainmail" {
			found = true
			break
		}
	}
	if !found {
		t.Error("unknown genre should default to fantasy items")
	}
}

func TestCredit_NewCredit(t *testing.T) {
	c := NewCredit(500)
	if c == nil {
		t.Fatal("NewCredit returned nil")
	}
	if c.Get() != 500 {
		t.Errorf("expected 500 credits, got %d", c.Get())
	}
}

func TestCredit_Add(t *testing.T) {
	c := NewCredit(100)
	c.Add(50)
	if c.Get() != 150 {
		t.Errorf("expected 150 credits, got %d", c.Get())
	}
	c.Add(200)
	if c.Get() != 350 {
		t.Errorf("expected 350 credits, got %d", c.Get())
	}
}

func TestCredit_Deduct(t *testing.T) {
	c := NewCredit(100)
	success := c.Deduct(50)
	if !success {
		t.Error("Deduct should succeed with sufficient balance")
	}
	if c.Get() != 50 {
		t.Errorf("expected 50 credits, got %d", c.Get())
	}
}

func TestCredit_DeductInsufficient(t *testing.T) {
	c := NewCredit(50)
	success := c.Deduct(100)
	if success {
		t.Error("Deduct should fail with insufficient balance")
	}
	if c.Get() != 50 {
		t.Errorf("balance should remain 50, got %d", c.Get())
	}
}

func TestCredit_Set(t *testing.T) {
	c := NewCredit(100)
	c.Set(250)
	if c.Get() != 250 {
		t.Errorf("expected 250 credits, got %d", c.Get())
	}
}

func TestShopInventory_GetAllItems(t *testing.T) {
	inv := ShopInventory{
		Weapons:  []Item{{ID: "weapon1"}},
		Ammo:     []Item{{ID: "ammo1"}, {ID: "ammo2"}},
		Upgrades: []Item{{ID: "upgrade1"}},
	}
	all := inv.GetAllItems()
	if len(all) != 4 {
		t.Errorf("expected 4 items, got %d", len(all))
	}
}

func TestShopInventory_FindItem(t *testing.T) {
	inv := ShopInventory{
		Weapons: []Item{{ID: "weapon1", Name: "Sword"}},
		Ammo:    []Item{{ID: "ammo1", Name: "Arrows"}},
	}
	item := inv.FindItem("ammo1")
	if item == nil {
		t.Fatal("FindItem returned nil for valid ID")
	}
	if item.Name != "Arrows" {
		t.Errorf("expected Arrows, got %s", item.Name)
	}
}

func TestShopInventory_FindItemNotFound(t *testing.T) {
	inv := ShopInventory{}
	item := inv.FindItem("missing")
	if item != nil {
		t.Error("FindItem should return nil for missing item")
	}
}

func TestShop_Purchase(t *testing.T) {
	shop := NewArmory("fantasy")
	credits := NewCredit(1000)

	success := shop.Purchase("ammo_arrows", credits)
	if !success {
		t.Error("Purchase should succeed with sufficient credits")
	}
	if credits.Get() != 950 {
		t.Errorf("expected 950 credits remaining, got %d", credits.Get())
	}
}

func TestShop_PurchaseInsufficient(t *testing.T) {
	shop := NewArmory("scifi")
	credits := NewCredit(25)

	success := shop.Purchase("ammo_bullets", credits)
	if success {
		t.Error("Purchase should fail with insufficient credits")
	}
	if credits.Get() != 25 {
		t.Errorf("credits should remain 25, got %d", credits.Get())
	}
}

func TestShop_PurchaseOutOfStock(t *testing.T) {
	shop := NewShop([]Item{
		{ID: "rare", Name: "Rare", Price: 50, Stock: 0},
	})
	credits := NewCredit(100)

	success := shop.Purchase("rare", credits)
	if success {
		t.Error("Purchase should fail when out of stock")
	}
}

func TestShop_PurchaseItemNotFound(t *testing.T) {
	shop := NewArmory("horror")
	credits := NewCredit(500)

	success := shop.Purchase("missing_item", credits)
	if success {
		t.Error("Purchase should fail for non-existent item")
	}
	if credits.Get() != 500 {
		t.Error("credits should not change on failed purchase")
	}
}

func TestShop_PurchaseNilCredits(t *testing.T) {
	shop := NewArmory("cyberpunk")
	success := shop.Purchase("ammo_bullets", nil)
	if success {
		t.Error("Purchase should fail with nil credits")
	}
}

func TestShop_GetShopName(t *testing.T) {
	tests := []struct {
		genre string
		name  string
	}{
		{"fantasy", "Merchant Tent"},
		{"scifi", "Supply Depot"},
		{"horror", "Black Market"},
		{"cyberpunk", "Corpo Shop"},
		{"postapoc", "Scrap Trader"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			shop := NewArmory(tt.genre)
			name := shop.GetShopName()
			if name != tt.name {
				t.Errorf("expected %s, got %s", tt.name, name)
			}
		})
	}
}

func TestShop_InventoryCategories(t *testing.T) {
	shop := NewArmory("scifi")

	if len(shop.Inventory.Weapons) == 0 {
		t.Error("shop should have weapons")
	}
	if len(shop.Inventory.Ammo) == 0 {
		t.Error("shop should have ammo")
	}
	if len(shop.Inventory.Upgrades) == 0 {
		t.Error("shop should have upgrades")
	}
	if len(shop.Inventory.Consumables) == 0 {
		t.Error("shop should have consumables")
	}
	if len(shop.Inventory.Armor) == 0 {
		t.Error("shop should have armor")
	}
}

func TestShop_ItemTypes(t *testing.T) {
	shop := NewArmory("fantasy")

	// Check that items have correct types
	weaponFound := false
	upgradeFound := false

	for _, item := range shop.Inventory.Weapons {
		if item.Type == ItemTypeWeapon {
			weaponFound = true
		}
	}
	for _, item := range shop.Inventory.Upgrades {
		if item.Type == ItemTypeUpgrade {
			upgradeFound = true
		}
	}

	if !weaponFound {
		t.Error("weapons should have ItemTypeWeapon type")
	}
	if !upgradeFound {
		t.Error("upgrades should have ItemTypeUpgrade type")
	}
}

func TestShop_PurchaseStockDecrement(t *testing.T) {
	shop := NewArmory("postapoc")
	credits := NewCredit(5000)

	// Find a limited stock item
	initialStock := 0
	var itemID string
	for _, item := range shop.Inventory.Armor {
		if item.Stock > 0 {
			initialStock = item.Stock
			itemID = item.ID
			break
		}
	}

	if itemID == "" {
		t.Skip("no limited stock items found")
	}

	// Purchase the item
	shop.Purchase(itemID, credits)

	// Check stock decremented
	item := shop.Inventory.FindItem(itemID)
	if item.Stock != initialStock-1 {
		t.Errorf("expected stock %d, got %d", initialStock-1, item.Stock)
	}
}

func TestCredit_ConcurrentAccess(t *testing.T) {
	c := NewCredit(1000)
	done := make(chan bool)

	// Concurrent adds
	for i := 0; i < 10; i++ {
		go func() {
			c.Add(10)
			done <- true
		}()
	}

	// Concurrent deducts
	for i := 0; i < 5; i++ {
		go func() {
			c.Deduct(10)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// Final balance should be 1000 + (10*10) - (5*10) = 1050
	if c.Get() != 1050 {
		t.Errorf("expected 1050 credits after concurrent ops, got %d", c.Get())
	}
}

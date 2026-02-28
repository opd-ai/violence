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

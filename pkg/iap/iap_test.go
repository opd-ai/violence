package iap

import (
	"testing"
)

func TestStubProvider_ListProducts(t *testing.T) {
	provider := NewStubProvider()
	products, err := provider.ListProducts()
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
	if products != nil {
		t.Errorf("expected nil products, got %v", products)
	}
}

func TestStubProvider_GetProduct(t *testing.T) {
	provider := NewStubProvider()
	product, err := provider.GetProduct("test")
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
	if product != nil {
		t.Errorf("expected nil product, got %v", product)
	}
}

func TestStubProvider_Purchase(t *testing.T) {
	provider := NewStubProvider()
	purchase, err := provider.Purchase("test")
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
	if purchase != nil {
		t.Errorf("expected nil purchase, got %v", purchase)
	}
}

func TestStubProvider_RestorePurchases(t *testing.T) {
	provider := NewStubProvider()
	purchases, err := provider.RestorePurchases()
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
	if purchases != nil {
		t.Errorf("expected nil purchases, got %v", purchases)
	}
}

func TestStubProvider_VerifyPurchase(t *testing.T) {
	provider := NewStubProvider()
	err := provider.VerifyPurchase(&Purchase{})
	if err != ErrNotSupported {
		t.Errorf("expected ErrNotSupported, got %v", err)
	}
}

func TestGetProductByID_Found(t *testing.T) {
	product := GetProductByID("violence.cosmetic.developer_tip_small")
	if product == nil {
		t.Fatal("expected product, got nil")
	}
	if product.ID != "violence.cosmetic.developer_tip_small" {
		t.Errorf("expected ID %s, got %s", "violence.cosmetic.developer_tip_small", product.ID)
	}
	if !product.Cosmetic {
		t.Error("expected cosmetic product")
	}
}

func TestGetProductByID_NotFound(t *testing.T) {
	product := GetProductByID("nonexistent")
	if product != nil {
		t.Errorf("expected nil product, got %v", product)
	}
}

func TestValidateProduct_Valid(t *testing.T) {
	product := &Product{
		ID:       "test",
		Name:     "Test Product",
		Cosmetic: true,
	}
	if !ValidateProduct(product) {
		t.Error("expected valid product")
	}
}

func TestValidateProduct_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		product *Product
	}{
		{"nil product", nil},
		{"non-cosmetic", &Product{ID: "test", Name: "Test", Cosmetic: false}},
		{"empty ID", &Product{ID: "", Name: "Test", Cosmetic: true}},
		{"empty name", &Product{ID: "test", Name: "", Cosmetic: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if ValidateProduct(tt.product) {
				t.Error("expected invalid product")
			}
		})
	}
}

func TestCatalog_AllCosmetic(t *testing.T) {
	for i, product := range Catalog {
		if !product.Cosmetic {
			t.Errorf("product %d (%s) is not marked cosmetic", i, product.ID)
		}
		if product.ID == "" {
			t.Errorf("product %d has empty ID", i)
		}
		if product.Name == "" {
			t.Errorf("product %d has empty name", i)
		}
		if product.Price == "" {
			t.Errorf("product %d has empty price", i)
		}
	}
}

func TestCatalog_MinimumProducts(t *testing.T) {
	if len(Catalog) < 3 {
		t.Errorf("expected at least 3 products, got %d", len(Catalog))
	}
}

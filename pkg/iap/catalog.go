package iap

// Catalog defines the complete set of available cosmetic-only IAP products.
// All products are cosmetic with no gameplay impact to comply with app store policies.
var Catalog = []Product{
	{
		ID:          "violence.cosmetic.developer_tip_small",
		Name:        "Small Developer Tip",
		Description: "Support development with a small tip",
		Price:       "$0.99",
		Type:        ProductTypeConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"tip","amount":"small"}`,
	},
	{
		ID:          "violence.cosmetic.developer_tip_medium",
		Name:        "Medium Developer Tip",
		Description: "Support development with a medium tip",
		Price:       "$2.99",
		Type:        ProductTypeConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"tip","amount":"medium"}`,
	},
	{
		ID:          "violence.cosmetic.developer_tip_large",
		Name:        "Large Developer Tip",
		Description: "Support development with a large tip",
		Price:       "$4.99",
		Type:        ProductTypeConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"tip","amount":"large"}`,
	},
	{
		ID:          "violence.cosmetic.player_skin_crimson",
		Name:        "Crimson Player Skin",
		Description: "Exclusive red character skin (cosmetic only)",
		Price:       "$1.99",
		Type:        ProductTypeNonConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"skin","target":"player","color":"crimson"}`,
	},
	{
		ID:          "violence.cosmetic.player_skin_emerald",
		Name:        "Emerald Player Skin",
		Description: "Exclusive green character skin (cosmetic only)",
		Price:       "$1.99",
		Type:        ProductTypeNonConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"skin","target":"player","color":"emerald"}`,
	},
	{
		ID:          "violence.cosmetic.weapon_trail_fire",
		Name:        "Fire Weapon Trail",
		Description: "Flaming weapon trail effect (cosmetic only)",
		Price:       "$0.99",
		Type:        ProductTypeNonConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"trail","target":"weapon","effect":"fire"}`,
	},
	{
		ID:          "violence.cosmetic.weapon_trail_ice",
		Name:        "Ice Weapon Trail",
		Description: "Frozen weapon trail effect (cosmetic only)",
		Price:       "$0.99",
		Type:        ProductTypeNonConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"trail","target":"weapon","effect":"ice"}`,
	},
	{
		ID:          "violence.cosmetic.name_color_gold",
		Name:        "Gold Name Color",
		Description: "Display your name in gold (cosmetic only)",
		Price:       "$1.49",
		Type:        ProductTypeNonConsumable,
		Cosmetic:    true,
		Metadata:    `{"type":"name_color","color":"gold"}`,
	},
}

// GetProductByID retrieves a product from the catalog by ID.
func GetProductByID(productID string) *Product {
	for i := range Catalog {
		if Catalog[i].ID == productID {
			return &Catalog[i]
		}
	}
	return nil
}

// ValidateProduct checks if a product is valid for purchase.
func ValidateProduct(product *Product) bool {
	if product == nil {
		return false
	}
	return product.Cosmetic && product.ID != "" && product.Name != ""
}

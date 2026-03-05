// Package iap provides in-app purchase stubs for mobile store compliance.
// This package defines cosmetic-only purchases with no gameplay impact.
package iap

import (
	"errors"
	"time"
)

// ProductType categorizes in-app purchase products.
type ProductType string

const (
	// ProductTypeConsumable represents a consumable purchase (e.g., cosmetic tip).
	ProductTypeConsumable ProductType = "consumable"
	// ProductTypeNonConsumable represents a one-time purchase (e.g., skin pack).
	ProductTypeNonConsumable ProductType = "non_consumable"
)

// Product represents an in-app purchase product.
type Product struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       string      `json:"price"` // Formatted price (e.g., "$0.99")
	Type        ProductType `json:"type"`
	Cosmetic    bool        `json:"cosmetic"` // Must be true for compliance
	Metadata    string      `json:"metadata"` // JSON-encoded cosmetic data
}

// Purchase represents a completed purchase transaction.
type Purchase struct {
	TransactionID string    `json:"transaction_id"`
	ProductID     string    `json:"product_id"`
	Timestamp     time.Time `json:"timestamp"`
	Verified      bool      `json:"verified"`
	Receipt       []byte    `json:"receipt,omitempty"` // Platform receipt data
}

var (
	// ErrNotSupported indicates IAP is not available on this platform.
	ErrNotSupported = errors.New("in-app purchases not supported on this platform")
	// ErrPurchaseCancelled indicates the user cancelled the purchase.
	ErrPurchaseCancelled = errors.New("purchase cancelled by user")
	// ErrVerificationFailed indicates receipt verification failed.
	ErrVerificationFailed = errors.New("purchase verification failed")
	// ErrProductNotFound indicates the requested product does not exist.
	ErrProductNotFound = errors.New("product not found")
)

// Provider defines the interface for platform-specific IAP implementations.
type Provider interface {
	// ListProducts retrieves available products from the store.
	ListProducts() ([]Product, error)

	// GetProduct retrieves a single product by ID.
	GetProduct(productID string) (*Product, error)

	// Purchase initiates a purchase flow for the given product.
	Purchase(productID string) (*Purchase, error)

	// RestorePurchases restores previously purchased non-consumable items.
	RestorePurchases() ([]Purchase, error)

	// VerifyPurchase verifies a purchase receipt with the platform.
	VerifyPurchase(purchase *Purchase) error
}

// StubProvider is a no-op implementation for non-mobile platforms.
type StubProvider struct{}

// NewStubProvider creates a stub provider for testing and desktop builds.
func NewStubProvider() *StubProvider {
	return &StubProvider{}
}

// ListProducts returns an empty list for stub provider.
func (s *StubProvider) ListProducts() ([]Product, error) {
	return nil, ErrNotSupported
}

// GetProduct returns an error for stub provider.
func (s *StubProvider) GetProduct(productID string) (*Product, error) {
	return nil, ErrNotSupported
}

// Purchase returns an error for stub provider.
func (s *StubProvider) Purchase(productID string) (*Purchase, error) {
	return nil, ErrNotSupported
}

// RestorePurchases returns an empty list for stub provider.
func (s *StubProvider) RestorePurchases() ([]Purchase, error) {
	return nil, ErrNotSupported
}

// VerifyPurchase returns an error for stub provider.
func (s *StubProvider) VerifyPurchase(purchase *Purchase) error {
	return ErrNotSupported
}

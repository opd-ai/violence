// Package dht provides decentralized server discovery using LibP2P Kademlia DHT.
package dht

import (
	"errors"

	record "github.com/libp2p/go-libp2p-record"
)

const (
	// ViolenceNamespace is the DHT namespace for Violence server records
	ViolenceNamespace = "violence"
)

// ViolenceValidator validates Violence DHT records.
type ViolenceValidator struct{}

// Validate validates a DHT record value.
func (v ViolenceValidator) Validate(key string, value []byte) error {
	// Accept all violence/* keys - content validation happens at application level
	if len(value) == 0 {
		return errors.New("empty record value")
	}
	return nil
}

// Select chooses the best record when multiple values exist.
// Returns the index of the best record.
func (v ViolenceValidator) Select(key string, values [][]byte) (int, error) {
	if len(values) == 0 {
		return 0, errors.New("no values to select from")
	}
	// For now, always select the first (most recent in DHT)
	return 0, nil
}

// NewValidator creates a namespaced validator for Violence records.
func NewValidator() record.NamespacedValidator {
	validator := record.NamespacedValidator{
		ViolenceNamespace: ViolenceValidator{},
	}
	// Add required IPFS validators
	validator["pk"] = record.PublicKeyValidator{}
	validator["ipns"] = record.PublicKeyValidator{} // IPNS uses public key validation
	return validator
}

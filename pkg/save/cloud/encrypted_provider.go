package cloud

import (
	"context"
	"encoding/json"
)

// EncryptedProvider wraps any Provider to add encryption at rest.
type EncryptedProvider struct {
	backend  Provider
	password string
}

// NewEncryptedProvider creates a new encrypted provider wrapping an existing backend.
func NewEncryptedProvider(backend Provider, password string) *EncryptedProvider {
	return &EncryptedProvider{
		backend:  backend,
		password: password,
	}
}

// Upload encrypts save data before uploading to the backend.
func (e *EncryptedProvider) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	encrypted, err := Encrypt(data, e.password)
	if err != nil {
		return err
	}
	return e.backend.Upload(ctx, slotID, encrypted, metadata)
}

// Download retrieves and decrypts save data from the backend.
func (e *EncryptedProvider) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	encrypted, metadata, err := e.backend.Download(ctx, slotID)
	if err != nil {
		return nil, metadata, err
	}

	decrypted, err := Decrypt(encrypted, e.password)
	if err != nil {
		return nil, metadata, err
	}

	return decrypted, metadata, nil
}

// List returns metadata for all cloud saves (metadata is not encrypted).
func (e *EncryptedProvider) List(ctx context.Context) ([]SaveMetadata, error) {
	return e.backend.List(ctx)
}

// Delete removes a save file from cloud storage.
func (e *EncryptedProvider) Delete(ctx context.Context, slotID int) error {
	return e.backend.Delete(ctx, slotID)
}

// GetMetadata retrieves metadata without downloading the file.
func (e *EncryptedProvider) GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error) {
	return e.backend.GetMetadata(ctx, slotID)
}

// EncryptedMetadata wraps metadata with encryption info for storage.
type EncryptedMetadata struct {
	SaveMetadata
	Encrypted bool `json:"encrypted"`
}

// MarshalMetadata serializes metadata to JSON.
func MarshalMetadata(metadata SaveMetadata) ([]byte, error) {
	return json.Marshal(metadata)
}

// UnmarshalMetadata deserializes metadata from JSON.
func UnmarshalMetadata(data []byte) (SaveMetadata, error) {
	var metadata SaveMetadata
	err := json.Unmarshal(data, &metadata)
	return metadata, err
}

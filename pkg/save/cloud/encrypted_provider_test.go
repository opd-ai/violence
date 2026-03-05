package cloud

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockEncryptedProvider is a simple in-memory provider for testing encryption.
type mockEncryptedProvider struct {
	data     map[int][]byte
	metadata map[int]SaveMetadata
	fails    bool
}

func newMockEncryptedProvider() *mockEncryptedProvider {
	return &mockEncryptedProvider{
		data:     make(map[int][]byte),
		metadata: make(map[int]SaveMetadata),
	}
}

func (m *mockEncryptedProvider) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	if m.fails {
		return errors.New("mock upload failed")
	}
	m.data[slotID] = data
	m.metadata[slotID] = metadata
	return nil
}

func (m *mockEncryptedProvider) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	if m.fails {
		return nil, SaveMetadata{}, errors.New("mock download failed")
	}
	data, ok := m.data[slotID]
	if !ok {
		return nil, SaveMetadata{}, ErrNotFound
	}
	return data, m.metadata[slotID], nil
}

func (m *mockEncryptedProvider) List(ctx context.Context) ([]SaveMetadata, error) {
	if m.fails {
		return nil, errors.New("mock list failed")
	}
	var result []SaveMetadata
	for _, meta := range m.metadata {
		result = append(result, meta)
	}
	return result, nil
}

func (m *mockEncryptedProvider) Delete(ctx context.Context, slotID int) error {
	if m.fails {
		return errors.New("mock delete failed")
	}
	delete(m.data, slotID)
	delete(m.metadata, slotID)
	return nil
}

func (m *mockEncryptedProvider) GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error) {
	if m.fails {
		return SaveMetadata{}, errors.New("mock getmetadata failed")
	}
	meta, ok := m.metadata[slotID]
	if !ok {
		return SaveMetadata{}, ErrNotFound
	}
	return meta, nil
}

func TestEncryptedProviderUploadDownload(t *testing.T) {
	backend := newMockEncryptedProvider()
	password := "test-password-123"
	provider := NewEncryptedProvider(backend, password)

	ctx := context.Background()
	slotID := 1
	originalData := []byte("save file data with inventory and progress")
	metadata := SaveMetadata{
		SlotID:    slotID,
		Timestamp: time.Now(),
		Version:   "1.0",
		Genre:     "fantasy",
		Seed:      12345,
		Size:      int64(len(originalData)),
		Checksum:  "abc123",
	}

	err := provider.Upload(ctx, slotID, originalData, metadata)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	encryptedData := backend.data[slotID]
	if len(encryptedData) <= len(originalData) {
		t.Error("encrypted data should be larger than original due to overhead")
	}

	downloadedData, downloadedMeta, err := provider.Download(ctx, slotID)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if string(downloadedData) != string(originalData) {
		t.Error("downloaded data doesn't match original")
	}

	if downloadedMeta.SlotID != metadata.SlotID {
		t.Errorf("metadata mismatch: got slotID %d, want %d", downloadedMeta.SlotID, metadata.SlotID)
	}
}

func TestEncryptedProviderWrongPassword(t *testing.T) {
	backend := newMockEncryptedProvider()
	provider1 := NewEncryptedProvider(backend, "password1")
	provider2 := NewEncryptedProvider(backend, "password2")

	ctx := context.Background()
	slotID := 1
	data := []byte("secret data")
	metadata := SaveMetadata{SlotID: slotID}

	err := provider1.Upload(ctx, slotID, data, metadata)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	_, _, err = provider2.Download(ctx, slotID)
	if err != ErrDecryptionFailed {
		t.Errorf("expected ErrDecryptionFailed with wrong password, got: %v", err)
	}
}

func TestEncryptedProviderList(t *testing.T) {
	backend := newMockEncryptedProvider()
	provider := NewEncryptedProvider(backend, "password")

	ctx := context.Background()
	metadata1 := SaveMetadata{SlotID: 1, Genre: "fantasy"}
	metadata2 := SaveMetadata{SlotID: 2, Genre: "scifi"}

	provider.Upload(ctx, 1, []byte("data1"), metadata1)
	provider.Upload(ctx, 2, []byte("data2"), metadata2)

	list, err := provider.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 items, got %d", len(list))
	}
}

func TestEncryptedProviderDelete(t *testing.T) {
	backend := newMockEncryptedProvider()
	provider := NewEncryptedProvider(backend, "password")

	ctx := context.Background()
	slotID := 1
	metadata := SaveMetadata{SlotID: slotID}

	provider.Upload(ctx, slotID, []byte("data"), metadata)

	err := provider.Delete(ctx, slotID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, _, err = provider.Download(ctx, slotID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestEncryptedProviderGetMetadata(t *testing.T) {
	backend := newMockEncryptedProvider()
	provider := NewEncryptedProvider(backend, "password")

	ctx := context.Background()
	slotID := 1
	metadata := SaveMetadata{
		SlotID:  slotID,
		Genre:   "horror",
		Version: "2.0",
	}

	provider.Upload(ctx, slotID, []byte("data"), metadata)

	retrieved, err := provider.GetMetadata(ctx, slotID)
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if retrieved.Genre != metadata.Genre {
		t.Errorf("metadata genre mismatch: got %s, want %s", retrieved.Genre, metadata.Genre)
	}
}

func TestEncryptedProviderBackendFailures(t *testing.T) {
	backend := newMockEncryptedProvider()
	backend.fails = true
	provider := NewEncryptedProvider(backend, "password")

	ctx := context.Background()
	metadata := SaveMetadata{SlotID: 1}

	if err := provider.Upload(ctx, 1, []byte("data"), metadata); err == nil {
		t.Error("expected upload to fail with failing backend")
	}

	if _, err := provider.List(ctx); err == nil {
		t.Error("expected list to fail with failing backend")
	}

	if err := provider.Delete(ctx, 1); err == nil {
		t.Error("expected delete to fail with failing backend")
	}

	if _, err := provider.GetMetadata(ctx, 1); err == nil {
		t.Error("expected getmetadata to fail with failing backend")
	}
}

func TestMarshalUnmarshalMetadata(t *testing.T) {
	original := SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now().Truncate(time.Second),
		Version:   "1.0",
		Genre:     "cyberpunk",
		Seed:      99999,
		Size:      1024,
		Checksum:  "sha256sum",
	}

	data, err := MarshalMetadata(original)
	if err != nil {
		t.Fatalf("MarshalMetadata failed: %v", err)
	}

	recovered, err := UnmarshalMetadata(data)
	if err != nil {
		t.Fatalf("UnmarshalMetadata failed: %v", err)
	}

	if recovered.SlotID != original.SlotID || recovered.Genre != original.Genre {
		t.Error("metadata doesn't match after marshal/unmarshal")
	}
}

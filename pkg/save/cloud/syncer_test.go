package cloud

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockProvider implements Provider for testing.
type mockProvider struct {
	saves map[int]mockSave
}

type mockSave struct {
	data     []byte
	metadata SaveMetadata
}

func newMockProvider() *mockProvider {
	return &mockProvider{saves: make(map[int]mockSave)}
}

func (m *mockProvider) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	m.saves[slotID] = mockSave{data: data, metadata: metadata}
	return nil
}

func (m *mockProvider) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	save, ok := m.saves[slotID]
	if !ok {
		return nil, SaveMetadata{}, ErrNotFound
	}
	return save.data, save.metadata, nil
}

func (m *mockProvider) List(ctx context.Context) ([]SaveMetadata, error) {
	list := make([]SaveMetadata, 0, len(m.saves))
	for _, save := range m.saves {
		list = append(list, save.metadata)
	}
	return list, nil
}

func (m *mockProvider) Delete(ctx context.Context, slotID int) error {
	if _, ok := m.saves[slotID]; !ok {
		return ErrNotFound
	}
	delete(m.saves, slotID)
	return nil
}

func (m *mockProvider) GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error) {
	save, ok := m.saves[slotID]
	if !ok {
		return SaveMetadata{}, ErrNotFound
	}
	return save.metadata, nil
}

func TestSyncer_Upload(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	data := []byte("test save data")
	meta := SaveMetadata{SlotID: 1, Version: "1.0"}

	if err := syncer.Upload(context.Background(), 1, data, meta); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	saved, ok := provider.saves[1]
	if !ok {
		t.Fatal("Save not found in provider")
	}

	if string(saved.data) != string(data) {
		t.Errorf("Data mismatch: got %q, want %q", saved.data, data)
	}

	if saved.metadata.Checksum == "" {
		t.Error("Checksum not set")
	}

	if saved.metadata.Size != int64(len(data)) {
		t.Errorf("Size mismatch: got %d, want %d", saved.metadata.Size, len(data))
	}
}

func TestSyncer_Download(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	data := []byte("test save data")
	meta := SaveMetadata{
		SlotID:   1,
		Version:  "1.0",
		Checksum: computeChecksum(data),
		Size:     int64(len(data)),
	}

	provider.saves[1] = mockSave{data: data, metadata: meta}

	downloaded, downloadedMeta, err := syncer.Download(context.Background(), 1)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if string(downloaded) != string(data) {
		t.Errorf("Data mismatch: got %q, want %q", downloaded, data)
	}

	if downloadedMeta.SlotID != 1 {
		t.Errorf("SlotID mismatch: got %d, want 1", downloadedMeta.SlotID)
	}
}

func TestSyncer_DownloadChecksumMismatch(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	data := []byte("test save data")
	meta := SaveMetadata{
		SlotID:   1,
		Checksum: "invalid_checksum",
	}

	provider.saves[1] = mockSave{data: data, metadata: meta}

	_, _, err := syncer.Download(context.Background(), 1)
	if err == nil {
		t.Fatal("Expected checksum error, got nil")
	}

	if !errors.Is(err, errors.New("checksum mismatch")) {
		if err.Error() != "checksum mismatch" {
			t.Errorf("Wrong error: got %v", err)
		}
	}
}

func TestSyncer_SyncLocalNewer(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	now := time.Now()
	localData := []byte("local data")
	localMeta := SaveMetadata{
		SlotID:    1,
		Timestamp: now,
	}

	cloudMeta := SaveMetadata{
		SlotID:    1,
		Timestamp: now.Add(-1 * time.Hour),
	}
	provider.saves[1] = mockSave{metadata: cloudMeta}

	err := syncer.Sync(context.Background(), 1, localData, localMeta, KeepLocal)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	saved := provider.saves[1]
	if string(saved.data) != string(localData) {
		t.Error("Local data not uploaded")
	}
}

func TestSyncer_SyncCloudNewer(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	now := time.Now()
	localData := []byte("local data")
	localMeta := SaveMetadata{
		SlotID:    1,
		Timestamp: now.Add(-1 * time.Hour),
	}

	cloudMeta := SaveMetadata{
		SlotID:    1,
		Timestamp: now,
	}
	provider.saves[1] = mockSave{metadata: cloudMeta}

	err := syncer.Sync(context.Background(), 1, localData, localMeta, KeepLocal)
	if !errors.Is(err, ErrConflict) {
		t.Errorf("Expected conflict error, got %v", err)
	}
}

func TestSyncer_SyncNoCloud(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	localData := []byte("local data")
	localMeta := SaveMetadata{SlotID: 1, Timestamp: time.Now()}

	err := syncer.Sync(context.Background(), 1, localData, localMeta, KeepLocal)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if _, ok := provider.saves[1]; !ok {
		t.Error("Local data not uploaded")
	}
}

func TestSyncer_List(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	provider.saves[1] = mockSave{metadata: SaveMetadata{SlotID: 1}}
	provider.saves[2] = mockSave{metadata: SaveMetadata{SlotID: 2}}

	list, err := syncer.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(list) != 2 {
		t.Errorf("List length: got %d, want 2", len(list))
	}
}

func TestSyncer_Delete(t *testing.T) {
	provider := newMockProvider()
	syncer := NewSyncer(provider, 10)

	provider.saves[1] = mockSave{metadata: SaveMetadata{SlotID: 1}}

	if err := syncer.Delete(context.Background(), 1); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if _, ok := provider.saves[1]; ok {
		t.Error("Save still exists after delete")
	}
}

package cloud

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockS3Provider is a test implementation of Provider.
type mockS3Provider struct {
	saves map[int][]byte
	metas map[int]SaveMetadata
}

func newMockS3Provider() *mockS3Provider {
	return &mockS3Provider{
		saves: make(map[int][]byte),
		metas: make(map[int]SaveMetadata),
	}
}

func (m *mockS3Provider) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	m.saves[slotID] = append([]byte{}, data...)
	m.metas[slotID] = metadata
	return nil
}

func (m *mockS3Provider) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	data, ok := m.saves[slotID]
	if !ok {
		return nil, SaveMetadata{}, ErrNotFound
	}
	return append([]byte{}, data...), m.metas[slotID], nil
}

func (m *mockS3Provider) List(ctx context.Context) ([]SaveMetadata, error) {
	var result []SaveMetadata
	for _, meta := range m.metas {
		result = append(result, meta)
	}
	return result, nil
}

func (m *mockS3Provider) Delete(ctx context.Context, slotID int) error {
	if _, ok := m.saves[slotID]; !ok {
		return ErrNotFound
	}
	delete(m.saves, slotID)
	delete(m.metas, slotID)
	return nil
}

func (m *mockS3Provider) GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error) {
	meta, ok := m.metas[slotID]
	if !ok {
		return SaveMetadata{}, ErrNotFound
	}
	return meta, nil
}

func TestS3Config_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  S3Config
		wantErr bool
	}{
		{
			name: "valid minimal config",
			config: S3Config{
				Bucket: "test-bucket",
			},
			wantErr: false,
		},
		{
			name: "valid full config",
			config: S3Config{
				Region:          "us-east-1",
				Bucket:          "test-bucket",
				Endpoint:        "http://localhost:9000",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
			wantErr: false,
		},
		{
			name:    "missing bucket",
			config:  S3Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewS3Provider(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewS3Provider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestS3Provider_UploadDownload(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	testData := []byte("test save data")
	testMeta := SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Genre:     "fantasy",
		Seed:      12345,
		Size:      int64(len(testData)),
		Checksum:  "abc123",
	}

	err := mock.Upload(ctx, 1, testData, testMeta)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	data, meta, err := mock.Download(ctx, 1)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("Downloaded data = %q, want %q", data, testData)
	}
	if meta.SlotID != testMeta.SlotID {
		t.Errorf("SlotID = %d, want %d", meta.SlotID, testMeta.SlotID)
	}
	if meta.Version != testMeta.Version {
		t.Errorf("Version = %q, want %q", meta.Version, testMeta.Version)
	}
}

func TestS3Provider_DownloadNotFound(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	_, _, err := mock.Download(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Download error = %v, want ErrNotFound", err)
	}
}

func TestS3Provider_List(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	saves := []struct {
		slotID int
		data   []byte
		meta   SaveMetadata
	}{
		{
			slotID: 1,
			data:   []byte("save1"),
			meta: SaveMetadata{
				SlotID:  1,
				Version: "1.0.0",
			},
		},
		{
			slotID: 2,
			data:   []byte("save2"),
			meta: SaveMetadata{
				SlotID:  2,
				Version: "1.0.1",
			},
		},
	}

	for _, s := range saves {
		if err := mock.Upload(ctx, s.slotID, s.data, s.meta); err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
	}

	metas, err := mock.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(metas) != 2 {
		t.Errorf("List returned %d saves, want 2", len(metas))
	}
}

func TestS3Provider_Delete(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	testData := []byte("test data")
	testMeta := SaveMetadata{SlotID: 1}

	if err := mock.Upload(ctx, 1, testData, testMeta); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if err := mock.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, _, err := mock.Download(ctx, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Download after Delete error = %v, want ErrNotFound", err)
	}
}

func TestS3Provider_GetMetadata(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	testMeta := SaveMetadata{
		SlotID:  1,
		Version: "1.0.0",
		Genre:   "scifi",
		Seed:    54321,
	}

	if err := mock.Upload(ctx, 1, []byte("data"), testMeta); err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	meta, err := mock.GetMetadata(ctx, 1)
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if meta.SlotID != testMeta.SlotID {
		t.Errorf("SlotID = %d, want %d", meta.SlotID, testMeta.SlotID)
	}
	if meta.Version != testMeta.Version {
		t.Errorf("Version = %q, want %q", meta.Version, testMeta.Version)
	}
	if meta.Genre != testMeta.Genre {
		t.Errorf("Genre = %q, want %q", meta.Genre, testMeta.Genre)
	}
}

func TestS3Provider_GetMetadataNotFound(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	_, err := mock.GetMetadata(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetMetadata error = %v, want ErrNotFound", err)
	}
}

func TestS3Provider_KeyGeneration(t *testing.T) {
	cfg := S3Config{
		Bucket: "test-bucket",
	}
	provider, err := NewS3Provider(cfg)
	if err != nil {
		t.Fatalf("NewS3Provider failed: %v", err)
	}

	tests := []struct {
		slotID      int
		wantKey     string
		wantMetaKey string
	}{
		{1, "saves/slot-1.sav", "saves/slot-1.meta.json"},
		{10, "saves/slot-10.sav", "saves/slot-10.meta.json"},
		{999, "saves/slot-999.sav", "saves/slot-999.meta.json"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			key := provider.key(tt.slotID)
			if key != tt.wantKey {
				t.Errorf("key(%d) = %q, want %q", tt.slotID, key, tt.wantKey)
			}
			metaKey := provider.metadataKey(tt.slotID)
			if metaKey != tt.wantMetaKey {
				t.Errorf("metadataKey(%d) = %q, want %q", tt.slotID, metaKey, tt.wantMetaKey)
			}
		})
	}
}

func TestS3Provider_MultipleUploads(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	for i := 1; i <= 5; i++ {
		data := []byte("data for slot")
		meta := SaveMetadata{
			SlotID:  i,
			Version: "1.0.0",
		}
		if err := mock.Upload(ctx, i, data, meta); err != nil {
			t.Fatalf("Upload slot %d failed: %v", i, err)
		}
	}

	metas, err := mock.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(metas) != 5 {
		t.Errorf("List returned %d saves, want 5", len(metas))
	}
}

func TestS3Provider_OverwriteSlot(t *testing.T) {
	mock := newMockS3Provider()
	ctx := context.Background()

	data1 := []byte("first data")
	meta1 := SaveMetadata{
		SlotID:  1,
		Version: "1.0.0",
		Seed:    100,
	}
	if err := mock.Upload(ctx, 1, data1, meta1); err != nil {
		t.Fatalf("First upload failed: %v", err)
	}

	data2 := []byte("second data")
	meta2 := SaveMetadata{
		SlotID:  1,
		Version: "2.0.0",
		Seed:    200,
	}
	if err := mock.Upload(ctx, 1, data2, meta2); err != nil {
		t.Fatalf("Second upload failed: %v", err)
	}

	data, meta, err := mock.Download(ctx, 1)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if string(data) != string(data2) {
		t.Errorf("Downloaded data = %q, want %q", data, data2)
	}
	if meta.Version != "2.0.0" {
		t.Errorf("Version = %q, want 2.0.0", meta.Version)
	}
	if meta.Seed != 200 {
		t.Errorf("Seed = %d, want 200", meta.Seed)
	}
}

func TestS3Config_WithEndpoint(t *testing.T) {
	cfg := S3Config{
		Bucket:   "test-bucket",
		Endpoint: "http://localhost:9000",
		Region:   "us-east-1",
	}
	provider, err := NewS3Provider(cfg)
	if err != nil {
		t.Fatalf("NewS3Provider failed: %v", err)
	}
	if provider == nil {
		t.Error("Expected provider, got nil")
	}
	if provider.bucket != "test-bucket" {
		t.Errorf("bucket = %q, want test-bucket", provider.bucket)
	}
}

func TestS3Config_WithCredentials(t *testing.T) {
	cfg := S3Config{
		Bucket:          "test-bucket",
		AccessKeyID:     "test-key-id",
		SecretAccessKey: "test-secret-key",
		Region:          "us-west-2",
	}
	provider, err := NewS3Provider(cfg)
	if err != nil {
		t.Fatalf("NewS3Provider failed: %v", err)
	}
	if provider == nil {
		t.Error("Expected provider, got nil")
	}
}

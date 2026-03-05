package cloud

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// mockWebDAVClient mocks the WebDAV client for testing.
type mockWebDAVClient struct {
	files  map[string][]byte
	mkdirs map[string]bool
}

func newMockWebDAVClient() *mockWebDAVClient {
	return &mockWebDAVClient{
		files:  make(map[string][]byte),
		mkdirs: make(map[string]bool),
	}
}

func (m *mockWebDAVClient) Write(path string, data []byte, perm os.FileMode) error {
	m.files[path] = data
	return nil
}

func (m *mockWebDAVClient) Read(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, errors.New("404 not found")
	}
	return data, nil
}

func (m *mockWebDAVClient) ReadDir(path string) ([]os.FileInfo, error) {
	var files []os.FileInfo
	for key := range m.files {
		if len(key) > len(path) && key[:len(path)] == path {
			files = append(files, &mockFileInfo{name: key[len(path)+1:]})
		}
	}
	return files, nil
}

func (m *mockWebDAVClient) Remove(path string) error {
	delete(m.files, path)
	return nil
}

func (m *mockWebDAVClient) MkdirAll(path string, perm os.FileMode) error {
	m.mkdirs[path] = true
	return nil
}

// mockFileInfo implements os.FileInfo for testing.
type mockFileInfo struct {
	name string
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0o644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestNewWebDAVProvider(t *testing.T) {
	tests := []struct {
		name    string
		cfg     WebDAVConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     WebDAVConfig{URL: "https://example.com/dav", Username: "user", Password: "pass"},
			wantErr: false,
		},
		{
			name:    "missing URL",
			cfg:     WebDAVConfig{Username: "user", Password: "pass"},
			wantErr: true,
		},
		{
			name:    "custom base path",
			cfg:     WebDAVConfig{URL: "https://example.com/dav", BasePath: "/custom/saves"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewWebDAVProvider(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWebDAVProvider() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && provider == nil {
				t.Error("NewWebDAVProvider() returned nil provider")
			}
		})
	}
}

func TestWebDAVProvider_Upload(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	metadata := SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now(),
		Version:   "1.0",
		Genre:     "fantasy",
		Seed:      12345,
		Size:      1024,
		Checksum:  "abc123",
	}
	data := []byte("test save data")

	ctx := context.Background()
	err = provider.Upload(ctx, 1, data, metadata)
	if err != nil {
		t.Errorf("Upload() error = %v", err)
	}

	if len(mock.files) != 2 {
		t.Errorf("Upload() created %d files, want 2", len(mock.files))
	}
}

func TestWebDAVProvider_Download(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	metadata := SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now(),
		Version:   "1.0",
		Genre:     "fantasy",
		Seed:      12345,
		Size:      1024,
		Checksum:  "abc123",
	}
	uploadData := []byte("test save data")

	ctx := context.Background()
	err = provider.Upload(ctx, 1, uploadData, metadata)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	data, meta, err := provider.Download(ctx, 1)
	if err != nil {
		t.Errorf("Download() error = %v", err)
	}

	if string(data) != string(uploadData) {
		t.Errorf("Download() data = %v, want %v", data, uploadData)
	}

	if meta.SlotID != metadata.SlotID {
		t.Errorf("Download() metadata.SlotID = %v, want %v", meta.SlotID, metadata.SlotID)
	}
}

func TestWebDAVProvider_DownloadNotFound(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	ctx := context.Background()
	_, _, err = provider.Download(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Download() error = %v, want ErrNotFound", err)
	}
}

func TestWebDAVProvider_Delete(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	metadata := SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now(),
		Version:   "1.0",
		Genre:     "fantasy",
		Seed:      12345,
		Size:      1024,
		Checksum:  "abc123",
	}
	data := []byte("test save data")

	ctx := context.Background()
	err = provider.Upload(ctx, 1, data, metadata)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	err = provider.Delete(ctx, 1)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	if len(mock.files) != 0 {
		t.Errorf("Delete() left %d files, want 0", len(mock.files))
	}

	_, _, err = provider.Download(ctx, 1)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Download after Delete error = %v, want ErrNotFound", err)
	}
}

func TestWebDAVProvider_GetMetadata(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	metadata := SaveMetadata{
		SlotID:    1,
		Timestamp: time.Now(),
		Version:   "1.0",
		Genre:     "fantasy",
		Seed:      12345,
		Size:      1024,
		Checksum:  "abc123",
	}
	data := []byte("test save data")

	ctx := context.Background()
	err = provider.Upload(ctx, 1, data, metadata)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	meta, err := provider.GetMetadata(ctx, 1)
	if err != nil {
		t.Errorf("GetMetadata() error = %v", err)
	}

	if meta.SlotID != metadata.SlotID {
		t.Errorf("GetMetadata() SlotID = %v, want %v", meta.SlotID, metadata.SlotID)
	}
}

func TestWebDAVProvider_GetMetadataNotFound(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	ctx := context.Background()
	_, err = provider.GetMetadata(ctx, 999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("GetMetadata() error = %v, want ErrNotFound", err)
	}
}

func TestWebDAVProvider_List(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		metadata := SaveMetadata{
			SlotID:    i,
			Timestamp: time.Now(),
			Version:   "1.0",
			Genre:     "fantasy",
			Seed:      int64(12345 + i),
			Size:      1024,
			Checksum:  fmt.Sprintf("abc%d", i),
		}
		data := []byte(fmt.Sprintf("save data %d", i))

		err = provider.Upload(ctx, i, data, metadata)
		if err != nil {
			t.Fatalf("Upload() error = %v", err)
		}
	}

	metadatas, err := provider.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(metadatas) != 3 {
		t.Errorf("List() returned %d metadatas, want 3", len(metadatas))
	}
}

func TestWebDAVProvider_ListEmpty(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := newMockWebDAVClient()
	provider.client = mock

	ctx := context.Background()
	metadatas, err := provider.List(ctx)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(metadatas) != 0 {
		t.Errorf("List() returned %d metadatas, want 0", len(metadatas))
	}
}

func TestWebDAVProvider_UploadMkdirError(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := &mockWebDAVClientWithErrors{mkdirErr: errors.New("mkdir failed")}
	provider.client = mock

	metadata := SaveMetadata{SlotID: 1}
	data := []byte("test")

	ctx := context.Background()
	err = provider.Upload(ctx, 1, data, metadata)
	if err == nil {
		t.Error("Upload() expected error, got nil")
	}
}

func TestWebDAVProvider_UploadWriteError(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := &mockWebDAVClientWithErrors{writeErr: errors.New("write failed")}
	provider.client = mock

	metadata := SaveMetadata{SlotID: 1}
	data := []byte("test")

	ctx := context.Background()
	err = provider.Upload(ctx, 1, data, metadata)
	if err == nil {
		t.Error("Upload() expected error, got nil")
	}
}

// mockWebDAVClientWithErrors implements error injection for testing.
type mockWebDAVClientWithErrors struct {
	mkdirErr   error
	writeErr   error
	readErr    error
	readDirErr error
	removeErr  error
}

func (m *mockWebDAVClientWithErrors) MkdirAll(path string, perm os.FileMode) error {
	if m.mkdirErr != nil {
		return m.mkdirErr
	}
	return nil
}

func (m *mockWebDAVClientWithErrors) Write(path string, data []byte, perm os.FileMode) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	return nil
}

func (m *mockWebDAVClientWithErrors) Read(path string) ([]byte, error) {
	if m.readErr != nil {
		return nil, m.readErr
	}
	return nil, errors.New("404 not found")
}

func (m *mockWebDAVClientWithErrors) ReadDir(path string) ([]os.FileInfo, error) {
	if m.readDirErr != nil {
		return nil, m.readDirErr
	}
	return []os.FileInfo{}, nil
}

func (m *mockWebDAVClientWithErrors) Remove(path string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	return nil
}

func TestWebDAVProvider_DownloadReadError(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := &mockWebDAVClientWithErrors{readErr: errors.New("read failed")}
	provider.client = mock

	ctx := context.Background()
	_, _, err = provider.Download(ctx, 1)
	if err == nil {
		t.Error("Download() expected error, got nil")
	}
}

func TestWebDAVProvider_DownloadReadSaveError(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	metadata := SaveMetadata{SlotID: 1}
	data := []byte("test")

	mock := newMockWebDAVClient()
	provider.client = mock

	ctx := context.Background()
	err = provider.Upload(ctx, 1, data, metadata)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	mockErr := &mockWebDAVClientReadSaveError{mock: mock}
	provider.client = mockErr

	_, _, err = provider.Download(ctx, 1)
	if err == nil {
		t.Error("Download() expected error, got nil")
	}
}

func TestWebDAVProvider_ListReadDirError(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := &mockWebDAVClientWithErrors{readDirErr: errors.New("readdir failed")}
	provider.client = mock

	ctx := context.Background()
	_, err = provider.List(ctx)
	if err == nil {
		t.Error("List() expected error, got nil")
	}
}

func TestWebDAVProvider_DeleteRemoveError(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := &mockWebDAVClientWithErrors{removeErr: errors.New("remove failed")}
	provider.client = mock

	ctx := context.Background()
	err = provider.Delete(ctx, 1)
	if err == nil {
		t.Error("Delete() expected error, got nil")
	}
}

func TestWebDAVProvider_GetMetadataReadError(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := &mockWebDAVClientWithErrors{readErr: errors.New("read failed")}
	provider.client = mock

	ctx := context.Background()
	_, err = provider.GetMetadata(ctx, 1)
	if err == nil {
		t.Error("GetMetadata() expected error, got nil")
	}
}

func TestWebDAVProvider_GetMetadataInvalidJSON(t *testing.T) {
	provider, err := NewWebDAVProvider(WebDAVConfig{URL: "https://example.com/dav"})
	if err != nil {
		t.Fatalf("NewWebDAVProvider() error = %v", err)
	}

	mock := &mockWebDAVClientValidData{
		files: map[string][]byte{
			"/saves/slot-1.meta.json": []byte("invalid json"),
		},
	}
	provider.client = mock

	ctx := context.Background()
	_, err = provider.GetMetadata(ctx, 1)
	if err == nil {
		t.Error("GetMetadata() expected JSON error, got nil")
	}
}

// mockWebDAVClientValidData returns specific data for testing.
type mockWebDAVClientValidData struct {
	files map[string][]byte
}

func (m *mockWebDAVClientValidData) Read(path string) ([]byte, error) {
	data, ok := m.files[path]
	if !ok {
		return nil, errors.New("404 not found")
	}
	return data, nil
}

func (m *mockWebDAVClientValidData) Write(path string, data []byte, perm os.FileMode) error {
	return nil
}

func (m *mockWebDAVClientValidData) ReadDir(path string) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (m *mockWebDAVClientValidData) Remove(path string) error {
	return nil
}

func (m *mockWebDAVClientValidData) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"404 error", errors.New("404 not found"), true},
		{"not found error", errors.New("file not found"), true},
		{"Not Found error", errors.New("Not Found"), true},
		{"other error", errors.New("something else"), false},
		{"empty error", errors.New(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNotFoundError(tt.err)
			if got != tt.want {
				t.Errorf("isNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// mockWebDAVClientReadSaveError returns metadata but fails on save read.
type mockWebDAVClientReadSaveError struct {
	mock *mockWebDAVClient
}

func (m *mockWebDAVClientReadSaveError) Read(path string) ([]byte, error) {
	if strings.HasSuffix(path, ".meta.json") {
		return m.mock.Read(path)
	}
	return nil, errors.New("read save error")
}

func (m *mockWebDAVClientReadSaveError) Write(path string, data []byte, perm os.FileMode) error {
	return m.mock.Write(path, data, perm)
}

func (m *mockWebDAVClientReadSaveError) ReadDir(path string) ([]os.FileInfo, error) {
	return m.mock.ReadDir(path)
}

func (m *mockWebDAVClientReadSaveError) Remove(path string) error {
	return m.mock.Remove(path)
}

func (m *mockWebDAVClientReadSaveError) MkdirAll(path string, perm os.FileMode) error {
	return m.mock.MkdirAll(path, perm)
}

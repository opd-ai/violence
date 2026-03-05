package cloud

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// MockProvider is a mock cloud provider for testing.
type MockProvider struct {
	mu        sync.Mutex
	uploads   map[int][]byte
	failures  int
	failUntil int
}

// NewMockProvider creates a new mock provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		uploads: make(map[int][]byte),
	}
}

// Upload simulates cloud upload.
func (m *MockProvider) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failures < m.failUntil {
		m.failures++
		return errors.New("mock upload failure")
	}

	m.uploads[slotID] = data
	return nil
}

// Download simulates cloud download.
func (m *MockProvider) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, ok := m.uploads[slotID]
	if !ok {
		return nil, SaveMetadata{}, ErrNotFound
	}
	return data, SaveMetadata{SlotID: slotID}, nil
}

// List returns all stored saves.
func (m *MockProvider) List(ctx context.Context) ([]SaveMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]SaveMetadata, 0, len(m.uploads))
	for id := range m.uploads {
		result = append(result, SaveMetadata{SlotID: id})
	}
	return result, nil
}

// Delete removes a save.
func (m *MockProvider) Delete(ctx context.Context, slotID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.uploads, slotID)
	return nil
}

// GetMetadata retrieves metadata.
func (m *MockProvider) GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.uploads[slotID]; !ok {
		return SaveMetadata{}, ErrNotFound
	}
	return SaveMetadata{SlotID: slotID}, nil
}

// GetUploadCount returns the number of successful uploads.
func (m *MockProvider) GetUploadCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.uploads)
}

func TestSyncWorker_QueueUpload(t *testing.T) {
	provider := NewMockProvider()
	syncer := NewSyncer(provider, 10)
	worker := NewSyncWorker(syncer, time.Hour, 3)

	data := []byte("test save data")
	metadata := SaveMetadata{SlotID: 1}

	worker.QueueUpload(1, data, metadata)

	if worker.QueueLength() != 1 {
		t.Errorf("expected queue length 1, got %d", worker.QueueLength())
	}
}

func TestSyncWorker_ProcessQueue(t *testing.T) {
	provider := NewMockProvider()
	syncer := NewSyncer(provider, 10)
	worker := NewSyncWorker(syncer, 100*time.Millisecond, 3)

	data := []byte("test save data")
	metadata := SaveMetadata{SlotID: 1}

	worker.QueueUpload(1, data, metadata)
	worker.Start()

	time.Sleep(200 * time.Millisecond)
	worker.Stop()

	if provider.GetUploadCount() != 1 {
		t.Errorf("expected 1 upload, got %d", provider.GetUploadCount())
	}

	if worker.QueueLength() != 0 {
		t.Errorf("expected empty queue, got %d", worker.QueueLength())
	}
}

func TestSyncWorker_RetryLogic(t *testing.T) {
	provider := NewMockProvider()
	provider.failUntil = 2
	syncer := NewSyncer(provider, 10)
	worker := NewSyncWorker(syncer, 50*time.Millisecond, 5)
	worker.baseBackoff = 50 * time.Millisecond

	data := []byte("test save data")
	metadata := SaveMetadata{SlotID: 1}

	worker.QueueUpload(1, data, metadata)
	worker.Start()

	time.Sleep(500 * time.Millisecond)
	worker.Stop()

	if provider.GetUploadCount() != 1 {
		t.Errorf("expected 1 upload after retries, got %d", provider.GetUploadCount())
	}
}

func TestSyncWorker_MaxRetries(t *testing.T) {
	provider := NewMockProvider()
	provider.failUntil = 100
	syncer := NewSyncer(provider, 10)
	worker := NewSyncWorker(syncer, 50*time.Millisecond, 2)

	data := []byte("test save data")
	metadata := SaveMetadata{SlotID: 1}

	worker.QueueUpload(1, data, metadata)
	worker.Start()

	time.Sleep(400 * time.Millisecond)
	worker.Stop()

	if provider.GetUploadCount() != 0 {
		t.Errorf("expected 0 uploads (max retries exceeded), got %d", provider.GetUploadCount())
	}
}

func TestSyncWorker_MultipleOperations(t *testing.T) {
	provider := NewMockProvider()
	syncer := NewSyncer(provider, 10)
	worker := NewSyncWorker(syncer, 100*time.Millisecond, 3)

	for i := 1; i <= 5; i++ {
		data := []byte("save data")
		metadata := SaveMetadata{SlotID: i}
		worker.QueueUpload(i, data, metadata)
	}

	worker.Start()
	time.Sleep(200 * time.Millisecond)
	worker.Stop()

	if provider.GetUploadCount() != 5 {
		t.Errorf("expected 5 uploads, got %d", provider.GetUploadCount())
	}
}

func TestSyncWorker_StopCleansUp(t *testing.T) {
	provider := NewMockProvider()
	syncer := NewSyncer(provider, 10)
	worker := NewSyncWorker(syncer, time.Hour, 3)

	worker.Start()
	worker.Stop()

	done := make(chan bool)
	go func() {
		worker.wg.Wait()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("worker did not stop cleanly")
	}
}

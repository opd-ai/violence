// Package cloud provides cloud save synchronization.
package cloud

import (
	"context"
	"log"
	"sync"
	"time"
)

// SyncOperation represents a pending sync operation.
type SyncOperation struct {
	SlotID   int
	Data     []byte
	Metadata SaveMetadata
	Retries  int
}

// SyncWorker manages background save synchronization with retry and offline queue.
type SyncWorker struct {
	syncer       *Syncer
	syncInterval time.Duration
	maxRetries   int
	baseBackoff  time.Duration
	queue        []SyncOperation
	mu           sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NewSyncWorker creates a background sync worker.
func NewSyncWorker(syncer *Syncer, syncInterval time.Duration, maxRetries int) *SyncWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &SyncWorker{
		syncer:       syncer,
		syncInterval: syncInterval,
		maxRetries:   maxRetries,
		baseBackoff:  time.Second,
		queue:        make([]SyncOperation, 0),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start begins the background sync worker.
func (w *SyncWorker) Start() {
	w.wg.Add(1)
	go w.run()
}

// run executes the background sync loop.
func (w *SyncWorker) run() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.processQueue()
		}
	}
}

// QueueUpload adds an upload operation to the offline queue.
func (w *SyncWorker) QueueUpload(slotID int, data []byte, metadata SaveMetadata) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.queue = append(w.queue, SyncOperation{
		SlotID:   slotID,
		Data:     data,
		Metadata: metadata,
		Retries:  0,
	})
}

// processQueue attempts to sync all queued operations.
func (w *SyncWorker) processQueue() {
	w.mu.Lock()
	pending := make([]SyncOperation, len(w.queue))
	copy(pending, w.queue)
	w.queue = w.queue[:0]
	w.mu.Unlock()

	for _, op := range pending {
		w.processSyncOp(op)
	}
}

// processSyncOp processes a single sync operation with retry.
func (w *SyncWorker) processSyncOp(op SyncOperation) {
	ctx := w.ctx
	err := w.syncer.Upload(ctx, op.SlotID, op.Data, op.Metadata)
	if err != nil {
		w.handleRetry(op, err)
	}
}

// handleRetry handles retry logic with exponential backoff.
func (w *SyncWorker) handleRetry(op SyncOperation, err error) {
	if op.Retries >= w.maxRetries {
		log.Printf("sync failed after %d retries: %v", w.maxRetries, err)
		return
	}

	op.Retries++
	backoff := time.Duration(1<<uint(op.Retries-1)) * w.baseBackoff

	go func() {
		select {
		case <-w.ctx.Done():
			return
		case <-time.After(backoff):
			w.mu.Lock()
			w.queue = append(w.queue, op)
			w.mu.Unlock()
		}
	}()
}

// Stop gracefully stops the background worker.
func (w *SyncWorker) Stop() {
	w.cancel()
	w.wg.Wait()
}

// QueueLength returns the current queue size.
func (w *SyncWorker) QueueLength() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.queue)
}

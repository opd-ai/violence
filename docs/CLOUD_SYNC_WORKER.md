# Background Sync Worker

## Overview

The Background Sync Worker (`pkg/save/cloud/worker.go`) provides automatic cloud save synchronization with retry logic and offline queueing capabilities. It enables "set it and forget it" cloud save uploads with robust failure handling.

## Features

- **Background Synchronization**: Periodic automatic uploads on a configurable interval
- **Retry Logic**: Exponential backoff retry mechanism with configurable maximum attempts
- **Offline Queue**: Operations queued when offline and automatically retried when online
- **Graceful Shutdown**: Clean shutdown with wait for in-flight operations
- **Thread-Safe**: Concurrent-safe queue management using mutexes

## Usage

### Basic Setup

```go
import "github.com/opd-ai/violence/pkg/save/cloud"

// Create a cloud provider (S3, WebDAV, etc.)
provider := cloud.NewS3Provider(config)
syncer := cloud.NewSyncer(provider, 10)

// Create worker with 5-minute sync interval and max 3 retries
worker := cloud.NewSyncWorker(syncer, 5*time.Minute, 3)

// Start background worker
worker.Start()
defer worker.Stop()
```

### Queueing Uploads

```go
// Queue a save for background upload
data, _ := json.Marshal(gameState)
metadata := cloud.SaveMetadata{
    SlotID:    1,
    Timestamp: time.Now(),
    Genre:     "fps",
    Version:   "1.0",
}

worker.QueueUpload(1, data, metadata)
```

### Monitoring Queue

```go
// Check current queue size
queueSize := worker.QueueLength()
log.Printf("Pending uploads: %d", queueSize)
```

## Configuration

### Sync Interval
Time between automatic sync attempts. Recommended values:
- **High-frequency games**: 1-5 minutes
- **Normal gameplay**: 5-15 minutes  
- **Background apps**: 30+ minutes

### Max Retries
Maximum retry attempts before giving up. Recommended:
- **Critical saves**: 5-10 retries
- **Best-effort syncs**: 3-5 retries
- **Non-essential data**: 1-2 retries

### Base Backoff
Starting delay for exponential backoff (default: 1 second). Backoff progression:
- Retry 1: 1s
- Retry 2: 2s
- Retry 3: 4s
- Retry 4: 8s
- Retry 5: 16s

## Error Handling

Failed uploads are automatically retried with exponential backoff. After exceeding max retries, failures are logged and the operation is dropped. Monitor logs for persistent failures:

```
sync failed after 3 retries: connection timeout
```

## Thread Safety

All public methods are thread-safe and can be called from multiple goroutines. The internal queue uses mutex protection for concurrent access.

## Testing

Comprehensive test coverage in `worker_test.go`:
- Queue upload functionality
- Background processing
- Retry logic with exponential backoff
- Max retry enforcement
- Multiple concurrent operations
- Graceful shutdown

Run tests:
```bash
go test ./pkg/save/cloud/... -v -run TestSyncWorker
```

## Performance

- **Memory overhead**: ~100 bytes per queued operation
- **CPU usage**: Minimal (timer-based wakeup)
- **Network**: Batched uploads on sync interval
- **Latency**: Immediate queueing, delayed background upload

## Integration Example

```go
// Game save system with cloud sync
type SaveManager struct {
    localSaves *save.Manager
    syncWorker *cloud.SyncWorker
}

func (sm *SaveManager) SaveGame(slot int, state GameState) error {
    // Save locally first
    if err := sm.localSaves.Save(slot, state); err != nil {
        return err
    }
    
    // Queue for cloud sync (non-blocking)
    data, _ := json.Marshal(state)
    metadata := cloud.SaveMetadata{
        SlotID:    slot,
        Timestamp: time.Now(),
        Genre:     state.Genre,
        Version:   state.Version,
    }
    sm.syncWorker.QueueUpload(slot, data, metadata)
    
    return nil
}
```

## Implementation Details

### Architecture
- `SyncWorker`: Main worker struct with queue and sync loop
- `SyncOperation`: Queued operation with retry count
- Background goroutine: Ticker-based sync loop
- Retry goroutines: Per-operation backoff timers

### Cancellation
Uses `context.Context` for graceful shutdown. Cancel propagates to:
- Main sync loop
- Retry timers
- In-flight uploads (if provider supports context)

### Queue Ordering
FIFO (First In, First Out) processing. Retries are appended to queue end, ensuring fairness for new operations.

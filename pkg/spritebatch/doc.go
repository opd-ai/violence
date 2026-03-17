// Package spritebatch provides a batched sprite rendering system for efficient GPU usage.
//
// The SpriteBatchSystem collects draw requests during a frame and issues them as batched
// draw calls, reducing GPU state changes and improving performance. Sprites are grouped
// by source texture to maximize batching efficiency.
//
// Usage:
//
//	batcher := spritebatch.NewSystem()
//	batcher.Begin() // Call at start of frame
//	batcher.Queue(sprite, x, y, opts) // Queue sprite draws
//	batcher.End(screen) // Flush all batched draws
//
// Features:
//   - Groups sprites by source texture for efficient batching
//   - Pools DrawImageOptions to avoid per-frame allocations
//   - Supports z-ordering within batches via priority layers
//   - Thread-safe queuing for concurrent systems
//   - Automatic batch flushing when batch size exceeds threshold
package spritebatch

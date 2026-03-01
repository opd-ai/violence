// Package pool provides thread-safe object pooling for zero-allocation hot paths.
//
// The pool package addresses the KNOWN TECHNICAL PROBLEM of memory churn by providing
// reusable object pools for frequently allocated types. This eliminates per-frame
// allocations in critical game systems.
//
// # Pooled Types
//
// - EntitySlices: Query result buffers
// - Images: Sprite generation buffers
// - Polygons: Collision geometry vertices
// - ByteSlices: General-purpose buffers
// - Float64Slices: Vector operations
//
// # Usage
//
// Use the global pool instances for common patterns:
//
//	entities := pool.GlobalPools.EntitySlices.Get()
//	defer pool.GlobalPools.EntitySlices.Put(entities)
//
//	img := pool.GlobalPools.Images.Get(32, 32)
//	defer pool.GlobalPools.Images.Put(img)
//
// # Performance
//
// Benchmark results demonstrate significant allocation reduction:
//
//	BenchmarkImageWithPool      0 allocs/op  (vs 2 allocs/op without pooling)
//	BenchmarkEntitySlicePool    0 allocs/op
//	BenchmarkPolygonPool        0 allocs/op
//
// # Thread Safety
//
// All pools use sync.Pool internally and are safe for concurrent access.
// Pools are optimized for parallel workloads with minimal contention.
//
// # Memory Management
//
// Pools enforce size limits to prevent unbounded growth:
//   - Entity slices: max 1024 capacity
//   - Byte slices: max 65536 capacity
//   - Images: max 256x256 pixels
//   - Polygons: max 128 vertices
//
// Oversized objects are not returned to the pool and will be garbage collected normally.
package pool

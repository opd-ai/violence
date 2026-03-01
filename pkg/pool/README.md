# Pool Package

Memory pooling system for zero-allocation hot paths in the Violence game engine.

## Overview

The pool package provides thread-safe object pooling for frequently allocated types:
- Entity slices (query results)
- Image buffers (sprite generation)
- Polygon vertices (collision detection)
- Byte slices (general buffers)
- Float64 slices (vector operations)

## Performance Impact

Benchmark results show significant allocation reduction:

```
BenchmarkImageWithPool         1402772    854.3 ns/op    0 B/op    0 allocs/op
BenchmarkImageWithoutPool       750404   1459 ns/op   4160 B/op    2 allocs/op

BenchmarkEntitySliceWithPool  19541184    62.13 ns/op    0 B/op    0 allocs/op
BenchmarkConcurrentImagePool  16506722    69.98 ns/op    0 B/op    0 allocs/op
```

## Usage

### Global Pools

Use the global pool instances for common patterns:

```go
import "github.com/opd-ai/violence/pkg/pool"

// Entity queries
entities := pool.GlobalPools.EntitySlices.Get()
defer pool.GlobalPools.EntitySlices.Put(entities)

// Image generation
img := pool.GlobalPools.Images.Get(32, 32)
defer pool.GlobalPools.Images.Put(img)

// Collision polygons
poly := pool.GlobalPools.Polygons.Get()
defer pool.GlobalPools.Polygons.Put(poly)
```

### Custom Pools

Create specialized pools for specific use cases:

```go
imagePool := pool.NewImagePool()
img := imagePool.Get(64, 64)
// ... use image ...
imagePool.Put(img)
```

## Integration Points

The pool package is integrated into:

1. **ECS Queries** (`pkg/engine/query.go`): Entity queries use pooled slices. Call `Release()` on iterators when done.

2. **Sprite Generation** (`pkg/sprite/sprite.go`): Image buffers are pooled during procedural sprite generation.

3. **Collision Detection** (`pkg/collision/collision.go`): Polygon transformations use pooled vertex arrays.

## Thread Safety

All pools use `sync.Pool` internally and are safe for concurrent access.

## Memory Limits

Pools enforce size limits to prevent unbounded growth:
- Entity slices: max 1024 capacity
- Byte slices: max 65536 capacity  
- Images: max 256x256 pixels
- Polygons: max 128 vertices

Oversized objects are not returned to the pool.

## Profiling

Use the global profiler to monitor pool efficiency:

```go
import "github.com/opd-ai/violence/pkg/pool"

// Sample memory statistics
pool.GlobalProfiler.Sample()

// Get statistics
stats := pool.GlobalProfiler.GetStats()
fmt.Printf("Image pool hit rate: %.2f%%\n", stats.ImageHitRate * 100)
fmt.Printf("Allocation rate: %.2f KB/s\n", stats.AllocRateBytesS / 1024)
```

// Package pool provides generic object pooling for zero-allocation hot paths.
package pool

import (
	"image"
	"sync"
)

// EntitySlicePool pools []Entity slices to reduce query allocations.
type EntitySlicePool struct {
	pool sync.Pool
}

// NewEntitySlicePool creates an entity slice pool.
func NewEntitySlicePool(capacity int) *EntitySlicePool {
	return &EntitySlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				s := make([]uint64, 0, capacity)
				return &s
			},
		},
	}
}

// Get retrieves a slice from the pool.
func (p *EntitySlicePool) Get() *[]uint64 {
	s := p.pool.Get().(*[]uint64)
	*s = (*s)[:0]
	return s
}

// Put returns a slice to the pool.
func (p *EntitySlicePool) Put(s *[]uint64) {
	if s != nil && cap(*s) <= 1024 {
		p.pool.Put(s)
	}
}

// Float64SlicePool pools []float64 slices for vector operations.
type Float64SlicePool struct {
	pool sync.Pool
}

// NewFloat64SlicePool creates a float64 slice pool.
func NewFloat64SlicePool(capacity int) *Float64SlicePool {
	return &Float64SlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				s := make([]float64, 0, capacity)
				return &s
			},
		},
	}
}

// Get retrieves a slice from the pool.
func (p *Float64SlicePool) Get() *[]float64 {
	s := p.pool.Get().(*[]float64)
	*s = (*s)[:0]
	return s
}

// Put returns a slice to the pool.
func (p *Float64SlicePool) Put(s *[]float64) {
	if s != nil && cap(*s) <= 1024 {
		p.pool.Put(s)
	}
}

// ImagePool pools image.RGBA instances for sprite generation.
type ImagePool struct {
	pools map[[2]int]*sync.Pool
	mu    sync.RWMutex
}

// NewImagePool creates an image pool with size buckets.
func NewImagePool() *ImagePool {
	return &ImagePool{
		pools: make(map[[2]int]*sync.Pool),
	}
}

// Get retrieves an image from the pool, creating if necessary.
func (p *ImagePool) Get(width, height int) *image.RGBA {
	key := [2]int{width, height}

	p.mu.RLock()
	pool, exists := p.pools[key]
	p.mu.RUnlock()

	if !exists {
		p.mu.Lock()
		if _, exists := p.pools[key]; !exists {
			p.pools[key] = &sync.Pool{
				New: func() interface{} {
					return image.NewRGBA(image.Rect(0, 0, width, height))
				},
			}
		}
		pool = p.pools[key]
		p.mu.Unlock()
	}

	img := pool.Get().(*image.RGBA)
	// Clear the image
	for i := range img.Pix {
		img.Pix[i] = 0
	}
	return img
}

// Put returns an image to the pool.
func (p *ImagePool) Put(img *image.RGBA) {
	if img == nil {
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > 256 || height > 256 {
		return
	}

	key := [2]int{width, height}
	p.mu.RLock()
	pool, exists := p.pools[key]
	p.mu.RUnlock()

	if exists {
		pool.Put(img)
	}
}

// ByteSlicePool pools byte slices for buffer operations.
type ByteSlicePool struct {
	pool sync.Pool
}

// NewByteSlicePool creates a byte slice pool.
func NewByteSlicePool(capacity int) *ByteSlicePool {
	return &ByteSlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				s := make([]byte, 0, capacity)
				return &s
			},
		},
	}
}

// Get retrieves a slice from the pool.
func (p *ByteSlicePool) Get() *[]byte {
	s := p.pool.Get().(*[]byte)
	*s = (*s)[:0]
	return s
}

// Put returns a slice to the pool.
func (p *ByteSlicePool) Put(s *[]byte) {
	if s != nil && cap(*s) <= 65536 {
		p.pool.Put(s)
	}
}

// PolygonPool pools coordinate slices for collision detection.
type PolygonPool struct {
	pool sync.Pool
}

// NewPolygonPool creates a polygon pool.
func NewPolygonPool(vertexCapacity int) *PolygonPool {
	return &PolygonPool{
		pool: sync.Pool{
			New: func() interface{} {
				s := make([][2]float64, 0, vertexCapacity)
				return &s
			},
		},
	}
}

// Get retrieves a polygon from the pool.
func (p *PolygonPool) Get() *[][2]float64 {
	s := p.pool.Get().(*[][2]float64)
	*s = (*s)[:0]
	return s
}

// Put returns a polygon to the pool.
func (p *PolygonPool) Put(s *[][2]float64) {
	if s != nil && cap(*s) <= 128 {
		p.pool.Put(s)
	}
}

// GlobalPools provides singleton access to common pools.
var GlobalPools = struct {
	EntitySlices *EntitySlicePool
	Float64s     *Float64SlicePool
	Images       *ImagePool
	Bytes        *ByteSlicePool
	Polygons     *PolygonPool
}{
	EntitySlices: NewEntitySlicePool(256),
	Float64s:     NewFloat64SlicePool(128),
	Images:       NewImagePool(),
	Bytes:        NewByteSlicePool(4096),
	Polygons:     NewPolygonPool(16),
}

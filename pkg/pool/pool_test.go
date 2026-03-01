package pool

import (
	"image"
	"runtime"
	"sync"
	"testing"
)

func TestEntitySlicePool(t *testing.T) {
	pool := NewEntitySlicePool(10)

	// Test get and put
	s1 := pool.Get()
	if s1 == nil {
		t.Fatal("Get returned nil")
	}
	if len(*s1) != 0 {
		t.Errorf("Expected length 0, got %d", len(*s1))
	}

	// Add data
	*s1 = append(*s1, 1, 2, 3)

	// Return to pool
	pool.Put(s1)

	// Get again - should be cleared
	s2 := pool.Get()
	if len(*s2) != 0 {
		t.Errorf("Expected cleared slice, got length %d", len(*s2))
	}
}

func TestEntitySlicePoolConcurrent(t *testing.T) {
	pool := NewEntitySlicePool(20)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				s := pool.Get()
				*s = append(*s, uint64(id))
				pool.Put(s)
			}
		}(i)
	}

	wg.Wait()
}

func TestFloat64SlicePool(t *testing.T) {
	pool := NewFloat64SlicePool(16)

	s := pool.Get()
	*s = append(*s, 1.0, 2.0, 3.0)
	pool.Put(s)

	s2 := pool.Get()
	if len(*s2) != 0 {
		t.Errorf("Expected cleared slice, got length %d", len(*s2))
	}
}

func TestImagePool(t *testing.T) {
	pool := NewImagePool()

	// Get an image
	img := pool.Get(32, 32)
	if img == nil {
		t.Fatal("Get returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("Expected 32x32, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Modify image
	img.Pix[0] = 255

	// Return to pool
	pool.Put(img)

	// Get again - should be cleared
	img2 := pool.Get(32, 32)
	if img2.Pix[0] != 0 {
		t.Errorf("Expected cleared image, got pixel value %d", img2.Pix[0])
	}
}

func TestImagePoolDifferentSizes(t *testing.T) {
	pool := NewImagePool()

	sizes := [][2]int{
		{16, 16},
		{32, 32},
		{64, 64},
		{16, 32},
		{32, 16},
	}

	for _, size := range sizes {
		img := pool.Get(size[0], size[1])
		bounds := img.Bounds()
		if bounds.Dx() != size[0] || bounds.Dy() != size[1] {
			t.Errorf("Expected %dx%d, got %dx%d", size[0], size[1], bounds.Dx(), bounds.Dy())
		}
		pool.Put(img)
	}
}

func TestImagePoolOversized(t *testing.T) {
	pool := NewImagePool()

	// Oversized images should not be pooled
	img := pool.Get(512, 512)
	pool.Put(img)

	// Second get should create new instance since oversized was not pooled
	img2 := pool.Get(512, 512)
	if img == img2 {
		t.Error("Oversized image should not be pooled")
	}
}

func TestByteSlicePool(t *testing.T) {
	pool := NewByteSlicePool(1024)

	s := pool.Get()
	*s = append(*s, 1, 2, 3)
	pool.Put(s)

	s2 := pool.Get()
	if len(*s2) != 0 {
		t.Errorf("Expected cleared slice, got length %d", len(*s2))
	}
}

func TestPolygonPool(t *testing.T) {
	pool := NewPolygonPool(8)

	poly := pool.Get()
	*poly = append(*poly, [2]float64{0, 0}, [2]float64{1, 0}, [2]float64{1, 1})
	pool.Put(poly)

	poly2 := pool.Get()
	if len(*poly2) != 0 {
		t.Errorf("Expected cleared polygon, got length %d", len(*poly2))
	}
}

func TestGlobalPools(t *testing.T) {
	// Test that global pools are initialized
	if GlobalPools.EntitySlices == nil {
		t.Error("EntitySlices pool not initialized")
	}
	if GlobalPools.Float64s == nil {
		t.Error("Float64s pool not initialized")
	}
	if GlobalPools.Images == nil {
		t.Error("Images pool not initialized")
	}
	if GlobalPools.Bytes == nil {
		t.Error("Bytes pool not initialized")
	}
	if GlobalPools.Polygons == nil {
		t.Error("Polygons pool not initialized")
	}

	// Test usage
	s := GlobalPools.EntitySlices.Get()
	*s = append(*s, 1, 2, 3)
	GlobalPools.EntitySlices.Put(s)
}

// Benchmark tests
func BenchmarkEntitySlicePoolGet(b *testing.B) {
	pool := NewEntitySlicePool(256)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := pool.Get()
		pool.Put(s)
	}
}

func BenchmarkEntitySliceAllocate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]uint32, 0, 256)
		_ = s
	}
}

func BenchmarkImagePoolGet(b *testing.B) {
	pool := NewImagePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := pool.Get(32, 32)
		pool.Put(img)
	}
}

func BenchmarkImageAllocate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		_ = img
	}
}

func BenchmarkPolygonPoolGet(b *testing.B) {
	pool := NewPolygonPool(16)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		poly := pool.Get()
		pool.Put(poly)
	}
}

func BenchmarkPolygonAllocate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		poly := make([][2]float64, 0, 16)
		_ = poly
	}
}

// Memory allocation tests
func TestEntitySlicePoolNoAlloc(t *testing.T) {
	pool := NewEntitySlicePool(10)

	// Pre-warm the pool
	for i := 0; i < 10; i++ {
		s := pool.Get()
		pool.Put(s)
	}

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < 100; i++ {
		s := pool.Get()
		*s = append(*s, uint64(i))
		pool.Put(s)
	}

	runtime.ReadMemStats(&memAfter)

	allocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc
	if allocDiff > 10000 {
		t.Logf("Warning: Pool allocated %d bytes for 100 operations", allocDiff)
	}
}

func TestImagePoolNoAlloc(t *testing.T) {
	pool := NewImagePool()

	// Pre-warm the pool
	for i := 0; i < 5; i++ {
		img := pool.Get(32, 32)
		pool.Put(img)
	}

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < 50; i++ {
		img := pool.Get(32, 32)
		pool.Put(img)
	}

	runtime.ReadMemStats(&memAfter)

	allocDiff := memAfter.TotalAlloc - memBefore.TotalAlloc
	if allocDiff > 50000 {
		t.Logf("Warning: Pool allocated %d bytes for 50 operations", allocDiff)
	}
}

func TestPoolCapacityLimits(t *testing.T) {
	// Test that oversized slices are not returned to pool
	entityPool := NewEntitySlicePool(10)
	largeSlice := make([]uint64, 0, 2048)
	entityPool.Put(&largeSlice)

	// Verify pool still works with normal sizes
	s := entityPool.Get()
	if cap(*s) > 1024 {
		t.Error("Pool should not return oversized slice")
	}

	// Test byte pool limits
	bytePool := NewByteSlicePool(1024)
	largeBytes := make([]byte, 0, 100000)
	bytePool.Put(&largeBytes)

	b := bytePool.Get()
	if cap(*b) > 65536 {
		t.Error("Pool should not return oversized byte slice")
	}
}

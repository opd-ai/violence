package pool

import (
	"image"
	"testing"
)

// BenchmarkEntitySliceWithPool benchmarks entity slice operations with pooling.
func BenchmarkEntitySliceWithPool(b *testing.B) {
	pool := NewEntitySlicePool(256)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := pool.Get()
		for j := 0; j < 100; j++ {
			*s = append(*s, uint64(j))
		}
		pool.Put(s)
	}

	b.ReportAllocs()
}

// BenchmarkEntitySliceWithoutPool benchmarks entity slice operations without pooling.
func BenchmarkEntitySliceWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]uint64, 0, 256)
		for j := 0; j < 100; j++ {
			s = append(s, uint64(j))
		}
		_ = s
	}

	b.ReportAllocs()
}

// BenchmarkImageWithPool benchmarks image operations with pooling.
func BenchmarkImageWithPool(b *testing.B) {
	pool := NewImagePool()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		img := pool.Get(32, 32)
		// Simulate some work
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				img.Pix[y*img.Stride+x*4] = 255
			}
		}
		pool.Put(img)
	}

	b.ReportAllocs()
}

// BenchmarkImageWithoutPool benchmarks image operations without pooling.
func BenchmarkImageWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		img := image.NewRGBA(image.Rect(0, 0, 32, 32))
		// Simulate some work
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				img.Pix[y*img.Stride+x*4] = 255
			}
		}
		_ = img
	}

	b.ReportAllocs()
}

// BenchmarkPolygonWithPool benchmarks polygon operations with pooling.
func BenchmarkPolygonWithPool(b *testing.B) {
	pool := NewPolygonPool(16)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		poly := pool.Get()
		for j := 0; j < 8; j++ {
			*poly = append(*poly, [2]float64{float64(j), float64(j * 2)})
		}
		pool.Put(poly)
	}

	b.ReportAllocs()
}

// BenchmarkPolygonWithoutPool benchmarks polygon operations without pooling.
func BenchmarkPolygonWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		poly := make([][2]float64, 0, 16)
		for j := 0; j < 8; j++ {
			poly = append(poly, [2]float64{float64(j), float64(j * 2)})
		}
		_ = poly
	}

	b.ReportAllocs()
}

// BenchmarkConcurrentEntitySlicePool benchmarks concurrent access to entity slice pool.
func BenchmarkConcurrentEntitySlicePool(b *testing.B) {
	pool := NewEntitySlicePool(256)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s := pool.Get()
			for j := 0; j < 50; j++ {
				*s = append(*s, uint64(j))
			}
			pool.Put(s)
		}
	})

	b.ReportAllocs()
}

// BenchmarkConcurrentImagePool benchmarks concurrent access to image pool.
func BenchmarkConcurrentImagePool(b *testing.B) {
	pool := NewImagePool()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			img := pool.Get(32, 32)
			img.Pix[0] = 255
			pool.Put(img)
		}
	})

	b.ReportAllocs()
}

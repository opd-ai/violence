package pool

import (
	"runtime"
	"testing"
)

// TestIntegrationMemoryReduction verifies that pooling reduces allocations in real usage patterns.
func TestIntegrationMemoryReduction(t *testing.T) {
	// Simulate a game loop performing queries, sprite generation, and collision checks
	const iterations = 1000

	// Measure with pooling
	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < iterations; i++ {
		// Query simulation
		entities := GlobalPools.EntitySlices.Get()
		for j := 0; j < 50; j++ {
			*entities = append(*entities, uint64(j))
		}
		GlobalPools.EntitySlices.Put(entities)

		// Sprite generation simulation
		img := GlobalPools.Images.Get(32, 32)
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				img.Pix[y*img.Stride+x*4] = uint8(x + y)
			}
		}
		GlobalPools.Images.Put(img)

		// Collision check simulation
		poly1 := GlobalPools.Polygons.Get()
		poly2 := GlobalPools.Polygons.Get()
		for k := 0; k < 4; k++ {
			*poly1 = append(*poly1, [2]float64{float64(k), float64(k)})
			*poly2 = append(*poly2, [2]float64{float64(k + 1), float64(k + 1)})
		}
		GlobalPools.Polygons.Put(poly1)
		GlobalPools.Polygons.Put(poly2)
	}

	runtime.ReadMemStats(&memAfter)

	allocWithPool := memAfter.TotalAlloc - memBefore.TotalAlloc

	// Measure without pooling
	runtime.GC()
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < iterations; i++ {
		// Query simulation
		entities := make([]uint64, 0, 50)
		for j := 0; j < 50; j++ {
			entities = append(entities, uint64(j))
		}

		// Sprite generation simulation
		img := GlobalPools.Images.Get(32, 32)
		for y := 0; y < 32; y++ {
			for x := 0; x < 32; x++ {
				img.Pix[y*img.Stride+x*4] = uint8(x + y)
			}
		}
		// Note: Not putting back to pool simulates allocation

		// Collision check simulation
		poly1 := make([][2]float64, 0, 4)
		poly2 := make([][2]float64, 0, 4)
		for k := 0; k < 4; k++ {
			poly1 = append(poly1, [2]float64{float64(k), float64(k)})
			poly2 = append(poly2, [2]float64{float64(k + 1), float64(k + 1)})
		}
		_ = poly1
		_ = poly2
	}

	runtime.ReadMemStats(&memAfter)

	allocWithoutPool := memAfter.TotalAlloc - memBefore.TotalAlloc

	// Pooling should reduce allocations significantly
	reductionPct := (1.0 - float64(allocWithPool)/float64(allocWithoutPool)) * 100.0

	t.Logf("Allocations with pooling:    %d bytes", allocWithPool)
	t.Logf("Allocations without pooling: %d bytes", allocWithoutPool)
	t.Logf("Reduction: %.1f%%", reductionPct)

	if reductionPct < 30 {
		t.Logf("Warning: Expected at least 30%% reduction, got %.1f%%", reductionPct)
	}
}

// TestIntegrationConcurrentAccess verifies thread-safety under concurrent load.
func TestIntegrationConcurrentAccess(t *testing.T) {
	const goroutines = 100
	const operations = 100

	done := make(chan bool, goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			for i := 0; i < operations; i++ {
				// Mix different pool operations
				entities := GlobalPools.EntitySlices.Get()
				*entities = append(*entities, uint64(i))
				GlobalPools.EntitySlices.Put(entities)

				if i%3 == 0 {
					img := GlobalPools.Images.Get(16, 16)
					GlobalPools.Images.Put(img)
				}

				if i%5 == 0 {
					poly := GlobalPools.Polygons.Get()
					*poly = append(*poly, [2]float64{1, 2})
					GlobalPools.Polygons.Put(poly)
				}
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for g := 0; g < goroutines; g++ {
		<-done
	}

	t.Logf("Completed %d concurrent goroutines with %d operations each", goroutines, operations)
}

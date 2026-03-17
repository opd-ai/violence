package spritebatch

import (
	"sync"
	"testing"
)

func TestNewSystem(t *testing.T) {
	s := NewSystem()

	if s == nil {
		t.Fatal("NewSystem returned nil")
	}

	if s.active {
		t.Error("New system should not be active")
	}

	// Verify batches are initialized
	for i, batch := range s.batches {
		if batch == nil {
			t.Errorf("Batch %d not initialized", i)
		}
		if len(batch) != 0 {
			t.Errorf("Batch %d should be empty, has %d items", i, len(batch))
		}
	}
}

func TestBeginEnd(t *testing.T) {
	s := NewSystem()

	// Begin should activate
	s.Begin()
	if !s.active {
		t.Error("System should be active after Begin")
	}

	// End without screen should deactivate
	s.End(nil)
	if s.active {
		t.Error("System should not be active after End")
	}
}

func TestQueueWithoutBegin(t *testing.T) {
	s := NewSystem()

	// Should not panic, just log warning
	item := NewBatchItem(nil, 0, 0)
	s.Queue(item)

	// Verify item was not added
	total := s.GetQueuedCount()
	if total != 0 {
		t.Errorf("Expected 0 queued items, got %d", total)
	}
}

func TestQueueNilSource(t *testing.T) {
	s := NewSystem()
	s.Begin()

	// Queue with nil source should be ignored
	item := NewBatchItem(nil, 0, 0)
	s.Queue(item)

	total := s.GetQueuedCount()
	if total != 0 {
		t.Errorf("Nil source items should be ignored, got %d items", total)
	}

	s.End(nil)
}

func TestLayerValidation(t *testing.T) {
	s := NewSystem()
	s.Begin()

	// Note: We can't easily test with real ebiten.Image in unit tests
	// This test verifies the system doesn't panic with edge cases

	// Test invalid layer values (will be clamped to LayerEntity)
	item := BatchItem{
		Source: nil, // Will be ignored anyway
		Layer:  Layer(-1),
	}
	s.Queue(item)

	item.Layer = Layer(100)
	s.Queue(item)

	s.End(nil)
}

func TestGetStats(t *testing.T) {
	s := NewSystem()

	stats := s.GetStats()
	if stats.ItemsQueued != 0 {
		t.Errorf("Initial ItemsQueued should be 0, got %d", stats.ItemsQueued)
	}
	if stats.DrawCalls != 0 {
		t.Errorf("Initial DrawCalls should be 0, got %d", stats.DrawCalls)
	}
	if stats.BatchesProcessed != 0 {
		t.Errorf("Initial BatchesProcessed should be 0, got %d", stats.BatchesProcessed)
	}
}

func TestIsActive(t *testing.T) {
	s := NewSystem()

	if s.IsActive() {
		t.Error("Should not be active initially")
	}

	s.Begin()
	if !s.IsActive() {
		t.Error("Should be active after Begin")
	}

	s.End(nil)
	if s.IsActive() {
		t.Error("Should not be active after End")
	}
}

func TestConcurrentQueue(t *testing.T) {
	s := NewSystem()
	s.Begin()

	var wg sync.WaitGroup
	numGoroutines := 10
	itemsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < itemsPerGoroutine; j++ {
				// Queue with nil source (will be ignored but tests thread safety)
				item := NewBatchItem(nil, float64(j), float64(j))
				s.Queue(item)
			}
		}()
	}

	wg.Wait()
	s.End(nil)

	// Should not have any items since source was nil
	stats := s.GetStats()
	if stats.ItemsQueued != 0 {
		t.Errorf("Expected 0 items (nil sources), got %d", stats.ItemsQueued)
	}
}

func TestBeginClearsPreviousFrame(t *testing.T) {
	s := NewSystem()

	// First frame
	s.Begin()
	s.End(nil)

	// Second frame
	s.Begin()

	// Batches should be cleared
	for i, batch := range s.batches {
		if len(batch) != 0 {
			t.Errorf("Batch %d not cleared, has %d items", i, len(batch))
		}
	}

	s.End(nil)
}

func TestQueueSimple(t *testing.T) {
	s := NewSystem()
	s.Begin()

	// QueueSimple with nil source should not panic
	s.QueueSimple(nil, 10, 20, LayerEntity, 5.0)

	s.End(nil)
}

func TestQueueWithTransform(t *testing.T) {
	s := NewSystem()
	s.Begin()

	// QueueWithTransform with nil source should not panic
	s.QueueWithTransform(nil, 10, 20, 2.0, 2.0, 1.57, LayerEffect, 3.0)

	s.End(nil)
}

func TestQueueWithColor(t *testing.T) {
	s := NewSystem()
	s.Begin()

	// QueueWithColor with nil source should not panic
	s.QueueWithColor(nil, 10, 20, 1.0, 0.5, 0.0, 0.8, LayerDecal, 1.0)

	s.End(nil)
}

func TestUpdate(t *testing.T) {
	s := NewSystem()

	// Update is a no-op but should not panic
	s.Update(nil)
}

func TestMultipleBeginWarning(t *testing.T) {
	s := NewSystem()

	s.Begin()
	// Second Begin should warn but not crash
	s.Begin()

	s.End(nil)
}

func TestEndWithoutBegin(t *testing.T) {
	s := NewSystem()

	// End without Begin should warn but not crash
	s.End(nil)
}

func BenchmarkQueueItems(b *testing.B) {
	s := NewSystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Begin()
		for j := 0; j < 1000; j++ {
			item := NewBatchItem(nil, float64(j), float64(j))
			item.Layer = LayerEntity
			s.Queue(item)
		}
		s.End(nil)
	}
}

func BenchmarkNewBatchItem(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewBatchItem(nil, float64(i), float64(i)).
			WithScale(1.5, 1.5).
			WithRotation(0.5).
			WithAlpha(0.8).
			WithLayer(LayerEffect)
	}
}

package spritebatch

import (
	"testing"
)

func TestQueueIfBatching_NilBatch(t *testing.T) {
	item := NewBatchItem(nil, 100, 100)

	// Should return false with nil batch
	batched := QueueIfBatching(nil, nil, item)
	if batched {
		t.Error("Should return false when batch is nil")
	}
}

func TestQueueIfBatching_InactiveBatch(t *testing.T) {
	s := NewSystem()
	item := NewBatchItem(nil, 100, 100)

	// Should return false when batch not started
	batched := QueueIfBatching(s, nil, item)
	if batched {
		t.Error("Should return false when batch is inactive")
	}
}

func TestQueueIfBatching_ActiveBatch(t *testing.T) {
	s := NewSystem()
	s.Begin()

	item := NewBatchItem(nil, 100, 100)

	// Should return true when batch is active (even with nil source)
	batched := QueueIfBatching(s, nil, item)
	if !batched {
		t.Error("Should return true when batch is active")
	}

	s.End(nil)
}

func TestQueueSpriteIfBatching(t *testing.T) {
	s := NewSystem()
	s.Begin()

	// Should work with nil source
	batched := QueueSpriteIfBatching(s, nil, nil, 100, 200, LayerEntity, 5.0)
	if !batched {
		t.Error("Should return true when batch is active")
	}

	s.End(nil)
}

func TestDirectDraw_NilSource(t *testing.T) {
	// Should not panic with nil source
	item := NewBatchItem(nil, 0, 0)
	directDraw(nil, item)
}

func TestDirectDraw_NilScreen(t *testing.T) {
	// Should not panic with nil screen but valid source
	// Can't test with real ebiten.Image in unit test
	item := BatchItem{
		Source: nil, // We can't easily create a real ebiten.Image in tests
	}
	directDraw(nil, item)
}

func BenchmarkQueueIfBatching(b *testing.B) {
	s := NewSystem()
	item := NewBatchItem(nil, 100, 200).WithLayer(LayerEntity)

	for i := 0; i < b.N; i++ {
		s.Begin()
		for j := 0; j < 100; j++ {
			QueueIfBatching(s, nil, item)
		}
		s.End(nil)
	}
}

func BenchmarkDirectDraw_NilSource(b *testing.B) {
	item := NewBatchItem(nil, 100, 200)
	for i := 0; i < b.N; i++ {
		directDraw(nil, item)
	}
}

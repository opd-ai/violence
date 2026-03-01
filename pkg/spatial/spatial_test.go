package spatial

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestGrid_InsertAndQuery(t *testing.T) {
	grid := NewGrid(10.0)

	e1 := engine.Entity(1)
	e2 := engine.Entity(2)
	e3 := engine.Entity(3)

	grid.Insert(e1, 5.0, 5.0)
	grid.Insert(e2, 15.0, 15.0)
	grid.Insert(e3, 100.0, 100.0)

	if grid.Count() != 3 {
		t.Errorf("expected 3 entities, got %d", grid.Count())
	}

	// Query near e1
	results := grid.QueryRadius(5.0, 5.0, 5.0)
	if len(results) == 0 {
		t.Error("expected to find entities near (5,5)")
	}
	found := false
	for _, e := range results {
		if e == e1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("did not find e1 in query results")
	}

	// Query far away
	results = grid.QueryRadius(1000.0, 1000.0, 5.0)
	if len(results) != 0 {
		t.Errorf("expected no entities far away, got %d", len(results))
	}
}

func TestGrid_Update(t *testing.T) {
	grid := NewGrid(10.0)

	e1 := engine.Entity(1)
	grid.Insert(e1, 5.0, 5.0)

	// Move to new position
	grid.Update(e1, 25.0, 25.0)

	// Should find at new position
	results := grid.QueryRadius(25.0, 25.0, 5.0)
	found := false
	for _, e := range results {
		if e == e1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("entity not found at new position")
	}

	// Should not find at old position
	results = grid.QueryRadius(5.0, 5.0, 5.0)
	for _, e := range results {
		if e == e1 {
			t.Error("entity still found at old position")
		}
	}
}

func TestGrid_UpdateSameCell(t *testing.T) {
	grid := NewGrid(10.0)

	e1 := engine.Entity(1)
	grid.Insert(e1, 5.0, 5.0)

	// Move within same cell
	grid.Update(e1, 6.0, 6.0)

	if grid.Count() != 1 {
		t.Errorf("expected 1 entity after update, got %d", grid.Count())
	}

	results := grid.QueryRadius(6.0, 6.0, 5.0)
	if len(results) == 0 {
		t.Error("entity not found after same-cell update")
	}
}

func TestGrid_Remove(t *testing.T) {
	grid := NewGrid(10.0)

	e1 := engine.Entity(1)
	e2 := engine.Entity(2)

	grid.Insert(e1, 5.0, 5.0)
	grid.Insert(e2, 15.0, 15.0)

	grid.Remove(e1)

	if grid.Count() != 1 {
		t.Errorf("expected 1 entity after removal, got %d", grid.Count())
	}

	results := grid.QueryRadius(5.0, 5.0, 5.0)
	for _, e := range results {
		if e == e1 {
			t.Error("removed entity still in grid")
		}
	}
}

func TestGrid_QueryBounds(t *testing.T) {
	grid := NewGrid(10.0)

	e1 := engine.Entity(1)
	e2 := engine.Entity(2)
	e3 := engine.Entity(3)

	grid.Insert(e1, 5.0, 5.0)
	grid.Insert(e2, 15.0, 15.0)
	grid.Insert(e3, 100.0, 100.0)

	results := grid.QueryBounds(0.0, 0.0, 20.0, 20.0)

	if len(results) < 2 {
		t.Errorf("expected at least 2 entities in bounds, got %d", len(results))
	}

	foundE1, foundE2 := false, false
	for _, e := range results {
		if e == e1 {
			foundE1 = true
		}
		if e == e2 {
			foundE2 = true
		}
		if e == e3 {
			t.Error("e3 should not be in bounds")
		}
	}

	if !foundE1 || !foundE2 {
		t.Error("expected to find e1 and e2 in bounds")
	}
}

func TestGrid_QueryRadiusFiltered(t *testing.T) {
	grid := NewGrid(10.0)

	e1 := engine.Entity(1)
	e2 := engine.Entity(2)
	e3 := engine.Entity(3)

	grid.Insert(e1, 5.0, 5.0)
	grid.Insert(e2, 10.0, 10.0) // distance ~7.07
	grid.Insert(e3, 20.0, 20.0) // distance ~21.2

	positions := map[engine.Entity]*engine.Position{
		e1: {X: 5.0, Y: 5.0},
		e2: {X: 10.0, Y: 10.0},
		e3: {X: 20.0, Y: 20.0},
	}

	// Query with radius 8 from (5,5)
	results := grid.QueryRadiusFiltered(5.0, 5.0, 8.0, positions)

	foundE1, foundE2, foundE3 := false, false, false
	for _, e := range results {
		if e == e1 {
			foundE1 = true
		}
		if e == e2 {
			foundE2 = true
		}
		if e == e3 {
			foundE3 = true
		}
	}

	if !foundE1 {
		t.Error("e1 should be within radius")
	}
	if !foundE2 {
		t.Error("e2 should be within radius")
	}
	if foundE3 {
		t.Error("e3 should not be within radius")
	}
}

func TestGrid_Clear(t *testing.T) {
	grid := NewGrid(10.0)

	for i := 0; i < 100; i++ {
		grid.Insert(engine.Entity(i), float64(i), float64(i))
	}

	if grid.Count() != 100 {
		t.Errorf("expected 100 entities, got %d", grid.Count())
	}

	grid.Clear()

	if grid.Count() != 0 {
		t.Errorf("expected 0 entities after clear, got %d", grid.Count())
	}

	if grid.CellCount() != 0 {
		t.Errorf("expected 0 cells after clear, got %d", grid.CellCount())
	}
}

func TestGrid_CellCoord(t *testing.T) {
	grid := NewGrid(10.0)

	tests := []struct {
		coord    float64
		expected int64
	}{
		{0.0, 0},
		{5.0, 0},
		{9.9, 0},
		{10.0, 1},
		{15.0, 1},
		{-5.0, -1},
		{-10.0, -1},
		{-10.1, -2},
	}

	for _, tt := range tests {
		result := grid.cellCoord(tt.coord)
		if result != tt.expected {
			t.Errorf("cellCoord(%f) = %d, expected %d", tt.coord, result, tt.expected)
		}
	}
}

func TestGrid_MultipleEntitiesPerCell(t *testing.T) {
	grid := NewGrid(10.0)

	// Insert multiple entities in same cell
	for i := 0; i < 5; i++ {
		grid.Insert(engine.Entity(i), 5.0+float64(i)*0.1, 5.0)
	}

	results := grid.QueryRadius(5.0, 5.0, 5.0)
	if len(results) != 5 {
		t.Errorf("expected 5 entities in cell, got %d", len(results))
	}
}

func TestGrid_GetCellSize(t *testing.T) {
	grid := NewGrid(17.5)
	if grid.GetCellSize() != 17.5 {
		t.Errorf("expected cell size 17.5, got %f", grid.GetCellSize())
	}
}

func BenchmarkGrid_Insert(b *testing.B) {
	grid := NewGrid(32.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		grid.Insert(engine.Entity(i), float64(i%1000), float64(i%1000))
	}
}

func BenchmarkGrid_Update(b *testing.B) {
	grid := NewGrid(32.0)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		grid.Insert(engine.Entity(i), float64(i), float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := engine.Entity(i % 1000)
		grid.Update(e, float64(i%1000), float64(i%1000))
	}
}

func BenchmarkGrid_QueryRadius(b *testing.B) {
	grid := NewGrid(32.0)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		grid.Insert(engine.Entity(i), float64(i%100)*10, float64(i%100)*10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grid.QueryRadius(500.0, 500.0, 50.0)
	}
}

func BenchmarkGrid_QueryRadiusFiltered(b *testing.B) {
	grid := NewGrid(32.0)

	positions := make(map[engine.Entity]*engine.Position)
	for i := 0; i < 1000; i++ {
		e := engine.Entity(i)
		x := float64(i%100) * 10
		y := float64(i%100) * 10
		grid.Insert(e, x, y)
		positions[e] = &engine.Position{X: x, Y: y}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grid.QueryRadiusFiltered(500.0, 500.0, 50.0, positions)
	}
}

func BenchmarkGrid_QueryBounds(b *testing.B) {
	grid := NewGrid(32.0)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		grid.Insert(engine.Entity(i), float64(i%100)*10, float64(i%100)*10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grid.QueryBounds(400.0, 400.0, 600.0, 600.0)
	}
}

func TestGrid_Concurrent(t *testing.T) {
	grid := NewGrid(10.0)

	// Test basic concurrent safety (read/write locks should prevent panics)
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			grid.Insert(engine.Entity(i), float64(i), float64(i))
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			grid.QueryRadius(50.0, 50.0, 10.0)
		}
		done <- true
	}()

	<-done
	<-done

	// Should complete without panic
}

func TestGrid_LargeScale(t *testing.T) {
	grid := NewGrid(64.0)

	// Insert 10000 entities across a 1000x1000 map
	for i := 0; i < 10000; i++ {
		x := math.Mod(float64(i)*13.7, 1000.0)
		y := math.Mod(float64(i)*17.3, 1000.0)
		grid.Insert(engine.Entity(i), x, y)
	}

	if grid.Count() != 10000 {
		t.Errorf("expected 10000 entities, got %d", grid.Count())
	}

	// Query should be fast
	results := grid.QueryRadius(500.0, 500.0, 100.0)
	if len(results) == 0 {
		t.Error("expected some entities in query")
	}

	// Verify cells are distributed
	if grid.CellCount() < 10 {
		t.Error("expected cells to be distributed across the grid")
	}
}

package spatial

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestSystem_Update(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem(10.0)

	// Create entities with positions
	e1 := w.AddEntity()
	w.AddComponent(e1, &engine.Position{X: 5.0, Y: 5.0})

	e2 := w.AddEntity()
	w.AddComponent(e2, &engine.Position{X: 15.0, Y: 15.0})

	e3 := w.AddEntity()
	w.AddComponent(e3, &engine.Position{X: 100.0, Y: 100.0})

	// Update spatial index
	sys.Update(w)

	if sys.Count() != 3 {
		t.Errorf("expected 3 indexed entities, got %d", sys.Count())
	}
}

func TestSystem_QueryRadius(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem(10.0)

	e1 := w.AddEntity()
	w.AddComponent(e1, &engine.Position{X: 5.0, Y: 5.0})

	e2 := w.AddEntity()
	w.AddComponent(e2, &engine.Position{X: 15.0, Y: 15.0})

	e3 := w.AddEntity()
	w.AddComponent(e3, &engine.Position{X: 100.0, Y: 100.0})

	sys.Update(w)

	// Query near e1
	results := sys.QueryRadius(5.0, 5.0, 10.0)
	if len(results) == 0 {
		t.Error("expected to find entities near (5,5)")
	}

	found := false
	for _, e := range results {
		if e == e1 {
			found = true
		}
		if e == e3 {
			t.Error("e3 should not be in query results")
		}
	}
	if !found {
		t.Error("e1 not found in query")
	}
}

func TestSystem_QueryRadiusExact(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem(10.0)

	e1 := w.AddEntity()
	w.AddComponent(e1, &engine.Position{X: 5.0, Y: 5.0})

	e2 := w.AddEntity()
	w.AddComponent(e2, &engine.Position{X: 10.0, Y: 10.0})

	e3 := w.AddEntity()
	w.AddComponent(e3, &engine.Position{X: 20.0, Y: 20.0})

	sys.Update(w)

	// Query with exact radius 8.0 from (5,5)
	// e1 is at distance 0, e2 at ~7.07, e3 at ~21.2
	results := sys.QueryRadiusExact(w, 5.0, 5.0, 8.0)

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

func TestSystem_QueryBounds(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem(10.0)

	e1 := w.AddEntity()
	w.AddComponent(e1, &engine.Position{X: 5.0, Y: 5.0})

	e2 := w.AddEntity()
	w.AddComponent(e2, &engine.Position{X: 15.0, Y: 15.0})

	e3 := w.AddEntity()
	w.AddComponent(e3, &engine.Position{X: 100.0, Y: 100.0})

	sys.Update(w)

	results := sys.QueryBounds(0.0, 0.0, 20.0, 20.0)

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

	if !foundE1 || !foundE2 {
		t.Error("expected to find e1 and e2 in bounds")
	}
	if foundE3 {
		t.Error("e3 should not be in bounds")
	}
}

func TestSystem_NoPositionComponent(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem(10.0)

	// Entity without position component
	e1 := w.AddEntity()
	w.AddComponent(e1, &engine.Health{Current: 100, Max: 100})

	sys.Update(w)

	if sys.Count() != 0 {
		t.Errorf("expected 0 indexed entities, got %d", sys.Count())
	}
}

func TestSystem_CellCount(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem(10.0)

	// Add entities in different cells
	for i := 0; i < 5; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{X: float64(i * 20), Y: float64(i * 20)})
	}

	sys.Update(w)

	if sys.CellCount() < 2 {
		t.Error("expected cells to be distributed")
	}
}

func TestSystem_GetGrid(t *testing.T) {
	sys := NewSystem(10.0)
	grid := sys.GetGrid()

	if grid == nil {
		t.Error("GetGrid returned nil")
	}

	if grid.GetCellSize() != 10.0 {
		t.Errorf("expected cell size 10.0, got %f", grid.GetCellSize())
	}
}

func TestSystem_MultipleUpdates(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem(10.0)

	e1 := w.AddEntity()
	w.AddComponent(e1, &engine.Position{X: 5.0, Y: 5.0})

	sys.Update(w)
	count1 := sys.Count()

	// Move entity
	posComp, _ := w.GetComponent(e1, reflect.TypeOf(&engine.Position{}))
	pos := posComp.(*engine.Position)
	pos.X = 50.0
	pos.Y = 50.0

	sys.Update(w)
	count2 := sys.Count()

	if count1 != 1 || count2 != 1 {
		t.Errorf("entity count should remain 1, got %d then %d", count1, count2)
	}

	// Query at new position
	results := sys.QueryRadius(50.0, 50.0, 10.0)
	found := false
	for _, e := range results {
		if e == e1 {
			found = true
		}
	}
	if !found {
		t.Error("entity not found at new position")
	}
}

func BenchmarkSystem_Update(b *testing.B) {
	w := engine.NewWorld()
	sys := NewSystem(32.0)

	// Pre-populate world with entities
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w)
	}
}

func BenchmarkSystem_QueryRadius(b *testing.B) {
	w := engine.NewWorld()
	sys := NewSystem(32.0)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		})
	}

	sys.Update(w)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.QueryRadius(500.0, 500.0, 50.0)
	}
}

func BenchmarkSystem_QueryRadiusExact(b *testing.B) {
	w := engine.NewWorld()
	sys := NewSystem(32.0)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		})
	}

	sys.Update(w)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.QueryRadiusExact(w, 500.0, 500.0, 50.0)
	}
}

func BenchmarkSystem_QueryBounds(b *testing.B) {
	w := engine.NewWorld()
	sys := NewSystem(32.0)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		})
	}

	sys.Update(w)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.QueryBounds(400.0, 400.0, 600.0, 600.0)
	}
}

// Comparison benchmark: linear iteration vs spatial index
func BenchmarkLinearQuery(b *testing.B) {
	w := engine.NewWorld()

	// Pre-populate
	entities := make([]engine.Entity, 1000)
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		})
		entities[i] = e
	}

	queryX, queryY, radius := 500.0, 500.0, 50.0
	radiusSq := radius * radius

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var results []engine.Entity
		posType := reflect.TypeOf(&engine.Position{})
		for _, e := range entities {
			comp, ok := w.GetComponent(e, posType)
			if !ok {
				continue
			}
			pos := comp.(*engine.Position)
			dx := pos.X - queryX
			dy := pos.Y - queryY
			if dx*dx+dy*dy <= radiusSq {
				results = append(results, e)
			}
		}
	}
}

func BenchmarkSpatialQuery(b *testing.B) {
	w := engine.NewWorld()
	sys := NewSystem(32.0)

	// Pre-populate
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{
			X: float64(i%100) * 10,
			Y: float64(i/100) * 10,
		})
	}

	sys.Update(w)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.QueryRadiusExact(w, 500.0, 500.0, 50.0)
	}
}

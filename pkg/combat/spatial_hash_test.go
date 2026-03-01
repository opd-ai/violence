package combat

import (
	"testing"
)

func TestNewSpatialHash(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
	}{
		{"small cells", 1.0},
		{"medium cells", 10.0},
		{"large cells", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := NewSpatialHash(tt.cellSize)
			if sh == nil {
				t.Fatal("NewSpatialHash returned nil")
			}
			if sh.cellSize != tt.cellSize {
				t.Errorf("cellSize = %v, want %v", sh.cellSize, tt.cellSize)
			}
			if sh.grid == nil {
				t.Error("grid map is nil")
			}
			if sh.CellCount() != 0 {
				t.Errorf("CellCount = %d, want 0", sh.CellCount())
			}
		})
	}
}

func TestSpatialHashInsert(t *testing.T) {
	tests := []struct {
		name      string
		cellSize  float64
		entity    Entity
		wantCells int
	}{
		{
			name:      "single cell entity",
			cellSize:  10.0,
			entity:    Entity{ID: 1, X: 5.0, Y: 5.0, Radius: 1.0},
			wantCells: 1,
		},
		{
			name:      "multi-cell entity",
			cellSize:  10.0,
			entity:    Entity{ID: 2, X: 10.0, Y: 10.0, Radius: 5.0},
			wantCells: 4,
		},
		{
			name:      "large entity",
			cellSize:  5.0,
			entity:    Entity{ID: 3, X: 0.0, Y: 0.0, Radius: 10.0},
			wantCells: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := NewSpatialHash(tt.cellSize)
			sh.Insert(tt.entity)

			cellCount := sh.CellCount()
			if cellCount != tt.wantCells {
				t.Errorf("CellCount = %d, want %d", cellCount, tt.wantCells)
			}

			entityCount := sh.EntityCount()
			if entityCount < tt.wantCells {
				t.Errorf("EntityCount = %d, want >= %d", entityCount, tt.wantCells)
			}
		})
	}
}

func TestSpatialHashQuery(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
		entities []Entity
		queryX   float64
		queryY   float64
		queryR   float64
		wantIDs  []uint64
	}{
		{
			name:     "no entities",
			cellSize: 10.0,
			entities: []Entity{},
			queryX:   0.0,
			queryY:   0.0,
			queryR:   5.0,
			wantIDs:  []uint64{},
		},
		{
			name:     "single match",
			cellSize: 10.0,
			entities: []Entity{
				{ID: 1, X: 5.0, Y: 5.0, Radius: 1.0},
			},
			queryX:  5.0,
			queryY:  5.0,
			queryR:  2.0,
			wantIDs: []uint64{1},
		},
		{
			name:     "multiple matches",
			cellSize: 10.0,
			entities: []Entity{
				{ID: 1, X: 5.0, Y: 5.0, Radius: 1.0},
				{ID: 2, X: 6.0, Y: 6.0, Radius: 1.0},
				{ID: 3, X: 100.0, Y: 100.0, Radius: 1.0},
			},
			queryX:  5.0,
			queryY:  5.0,
			queryR:  3.0,
			wantIDs: []uint64{1, 2},
		},
		{
			name:     "no matches",
			cellSize: 10.0,
			entities: []Entity{
				{ID: 1, X: 100.0, Y: 100.0, Radius: 1.0},
			},
			queryX:  0.0,
			queryY:  0.0,
			queryR:  5.0,
			wantIDs: []uint64{},
		},
		{
			name:     "overlapping entities",
			cellSize: 5.0,
			entities: []Entity{
				{ID: 1, X: 0.0, Y: 0.0, Radius: 2.0},
				{ID: 2, X: 3.0, Y: 0.0, Radius: 2.0},
			},
			queryX:  1.5,
			queryY:  0.0,
			queryR:  1.0,
			wantIDs: []uint64{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := NewSpatialHash(tt.cellSize)
			for _, e := range tt.entities {
				sh.Insert(e)
			}

			results := sh.Query(tt.queryX, tt.queryY, tt.queryR)

			if len(results) != len(tt.wantIDs) {
				t.Errorf("Query returned %d results, want %d", len(results), len(tt.wantIDs))
			}

			resultMap := make(map[uint64]bool)
			for _, id := range results {
				resultMap[id] = true
			}

			for _, wantID := range tt.wantIDs {
				if !resultMap[wantID] {
					t.Errorf("Query missing expected ID %d", wantID)
				}
			}
		})
	}
}

func TestSpatialHashQueryEntities(t *testing.T) {
	tests := []struct {
		name       string
		cellSize   float64
		entities   []Entity
		queryX     float64
		queryY     float64
		queryR     float64
		wantCount  int
		checkFirst bool
		wantFirstX float64
	}{
		{
			name:       "single entity",
			cellSize:   10.0,
			entities:   []Entity{{ID: 1, X: 5.0, Y: 5.0, Radius: 1.0}},
			queryX:     5.0,
			queryY:     5.0,
			queryR:     2.0,
			wantCount:  1,
			checkFirst: true,
			wantFirstX: 5.0,
		},
		{
			name:     "multiple entities",
			cellSize: 10.0,
			entities: []Entity{
				{ID: 1, X: 5.0, Y: 5.0, Radius: 1.0},
				{ID: 2, X: 6.0, Y: 6.0, Radius: 1.0},
			},
			queryX:    5.5,
			queryY:    5.5,
			queryR:    2.0,
			wantCount: 2,
		},
		{
			name:      "no entities",
			cellSize:  10.0,
			entities:  []Entity{},
			queryX:    0.0,
			queryY:    0.0,
			queryR:    5.0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := NewSpatialHash(tt.cellSize)
			for _, e := range tt.entities {
				sh.Insert(e)
			}

			results := sh.QueryEntities(tt.queryX, tt.queryY, tt.queryR)

			if len(results) != tt.wantCount {
				t.Errorf("QueryEntities returned %d results, want %d", len(results), tt.wantCount)
			}

			if tt.checkFirst && len(results) > 0 {
				if results[0].X != tt.wantFirstX {
					t.Errorf("first entity X = %v, want %v", results[0].X, tt.wantFirstX)
				}
			}
		})
	}
}

func TestSpatialHashClear(t *testing.T) {
	sh := NewSpatialHash(10.0)
	sh.Insert(Entity{ID: 1, X: 5.0, Y: 5.0, Radius: 1.0})
	sh.Insert(Entity{ID: 2, X: 15.0, Y: 15.0, Radius: 1.0})

	if sh.CellCount() == 0 {
		t.Fatal("CellCount should be > 0 before Clear")
	}

	sh.Clear()

	if sh.CellCount() != 0 {
		t.Errorf("CellCount after Clear = %d, want 0", sh.CellCount())
	}

	if sh.EntityCount() != 0 {
		t.Errorf("EntityCount after Clear = %d, want 0", sh.EntityCount())
	}

	results := sh.Query(5.0, 5.0, 10.0)
	if len(results) != 0 {
		t.Errorf("Query after Clear returned %d results, want 0", len(results))
	}
}

func TestSpatialHashCellCoord(t *testing.T) {
	tests := []struct {
		name       string
		cellSize   float64
		worldCoord float64
		want       int64
	}{
		{"positive coordinate", 10.0, 25.5, 2},
		{"zero coordinate", 10.0, 0.0, 0},
		{"negative coordinate", 10.0, -15.5, -2},
		{"small cell size", 1.0, 5.5, 5},
		{"large cell size", 100.0, 250.0, 2},
		{"boundary case", 10.0, 10.0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sh := NewSpatialHash(tt.cellSize)
			got := sh.cellCoord(tt.worldCoord)
			if got != tt.want {
				t.Errorf("cellCoord(%v) = %v, want %v", tt.worldCoord, got, tt.want)
			}
		})
	}
}

func TestSpatialHashNegativeCoordinates(t *testing.T) {
	sh := NewSpatialHash(10.0)
	sh.Insert(Entity{ID: 1, X: -5.0, Y: -5.0, Radius: 1.0})
	sh.Insert(Entity{ID: 2, X: 5.0, Y: 5.0, Radius: 1.0})

	results := sh.Query(-5.0, -5.0, 2.0)
	if len(results) != 1 || results[0] != 1 {
		t.Errorf("Query at negative coords failed, got %v", results)
	}

	results = sh.Query(5.0, 5.0, 2.0)
	if len(results) != 1 || results[0] != 2 {
		t.Errorf("Query at positive coords failed, got %v", results)
	}
}

func TestSpatialHashDuplicateIDs(t *testing.T) {
	sh := NewSpatialHash(10.0)
	sh.Insert(Entity{ID: 1, X: 5.0, Y: 5.0, Radius: 1.0})
	sh.Insert(Entity{ID: 1, X: 6.0, Y: 6.0, Radius: 1.0})

	results := sh.Query(5.5, 5.5, 5.0)

	uniqueIDs := make(map[uint64]int)
	for _, id := range results {
		uniqueIDs[id]++
	}

	if count := uniqueIDs[1]; count != 1 {
		t.Errorf("ID 1 appeared %d times in results, want 1", count)
	}
}

func TestSpatialHashLargeScale(t *testing.T) {
	sh := NewSpatialHash(50.0)

	for i := uint64(0); i < 1000; i++ {
		x := float64(i%100) * 10.0
		y := float64(i/100) * 10.0
		sh.Insert(Entity{ID: i, X: x, Y: y, Radius: 5.0})
	}

	results := sh.Query(450.0, 45.0, 100.0)

	if len(results) == 0 {
		t.Error("Large scale query returned no results")
	}

	for _, id := range results {
		if id >= 1000 {
			t.Errorf("Query returned invalid ID %d", id)
		}
	}
}

func TestSpatialHashZeroRadius(t *testing.T) {
	sh := NewSpatialHash(10.0)
	sh.Insert(Entity{ID: 1, X: 5.0, Y: 5.0, Radius: 0.0})

	results := sh.Query(5.0, 5.0, 0.0)

	if len(results) != 1 || results[0] != 1 {
		t.Errorf("Query with zero radius failed, got %v", results)
	}
}

func BenchmarkSpatialHashInsert(b *testing.B) {
	sh := NewSpatialHash(10.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sh.Insert(Entity{
			ID:     uint64(i),
			X:      float64(i % 100),
			Y:      float64(i / 100),
			Radius: 1.0,
		})
	}
}

func BenchmarkSpatialHashQuery(b *testing.B) {
	sh := NewSpatialHash(10.0)
	for i := uint64(0); i < 1000; i++ {
		sh.Insert(Entity{
			ID:     i,
			X:      float64(i % 100),
			Y:      float64(i / 100),
			Radius: 1.0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sh.Query(50.0, 50.0, 20.0)
	}
}

func BenchmarkSpatialHashQueryEntities(b *testing.B) {
	sh := NewSpatialHash(10.0)
	for i := uint64(0); i < 1000; i++ {
		sh.Insert(Entity{
			ID:     i,
			X:      float64(i % 100),
			Y:      float64(i / 100),
			Radius: 1.0,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sh.QueryEntities(50.0, 50.0, 20.0)
	}
}

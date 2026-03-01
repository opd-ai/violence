package engine

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/testutil"
)

// Test components
type Velocity struct {
	DX, DY float64
}

type Name struct {
	Value string
}

// Test system
type TestSystem struct {
	UpdateCount int
}

func (s *TestSystem) Update(w *World) {
	s.UpdateCount++
}

func TestWorld_AddEntity(t *testing.T) {
	tests := []struct {
		name        string
		numEntities int
	}{
		{"single entity", 1},
		{"multiple entities", 5},
		{"many entities", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			entities := make([]Entity, tt.numEntities)
			for i := 0; i < tt.numEntities; i++ {
				entities[i] = w.AddEntity()
			}

			// Check that IDs are unique and sequential
			for i, e := range entities {
				if e != Entity(i) {
					t.Errorf("entity %d: got ID %d, want %d", i, e, i)
				}
			}
		})
	}
}

func TestWorld_AddComponent(t *testing.T) {
	tests := []struct {
		name      string
		component Component
	}{
		{"Position", &Position{X: 10.0, Y: 20.0}},
		{"Velocity", &Velocity{DX: 1.0, DY: -1.0}},
		{"Health", &Health{Current: 100, Max: 100}},
		{"Name", &Name{Value: "Player"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			e := w.AddEntity()
			w.AddComponent(e, tt.component)

			comp, found := w.GetComponent(e, reflect.TypeOf(tt.component))
			if !found {
				t.Fatalf("component not found")
			}
			if !reflect.DeepEqual(comp, tt.component) {
				t.Errorf("got %+v, want %+v", comp, tt.component)
			}
		})
	}
}

func TestWorld_AddComponent_NilComponent(t *testing.T) {
	w := NewWorld()
	e := w.AddEntity()
	w.AddComponent(e, nil)

	// Should not panic and should not add nil component
	if len(w.components[e]) != 0 {
		t.Errorf("nil component should not be added, got %d components", len(w.components[e]))
	}
}

func TestWorld_GetComponent(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*World, Entity)
		componentType reflect.Type
		wantFound     bool
	}{
		{
			name: "existing component",
			setup: func(w *World, e Entity) {
				w.AddComponent(e, &Position{X: 5.0, Y: 10.0})
			},
			componentType: reflect.TypeOf(&Position{}),
			wantFound:     true,
		},
		{
			name: "non-existent component",
			setup: func(w *World, e Entity) {
				w.AddComponent(e, &Position{X: 5.0, Y: 10.0})
			},
			componentType: reflect.TypeOf(&Velocity{}),
			wantFound:     false,
		},
		{
			name:          "component on non-existent entity",
			setup:         func(w *World, e Entity) {},
			componentType: reflect.TypeOf(&Position{}),
			wantFound:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			e := w.AddEntity()
			tt.setup(w, e)

			_, found := w.GetComponent(e, tt.componentType)
			if found != tt.wantFound {
				t.Errorf("got found=%v, want %v", found, tt.wantFound)
			}
		})
	}
}

func TestWorld_HasComponent(t *testing.T) {
	w := NewWorld()
	e := w.AddEntity()

	// Initially should not have any components
	if w.HasComponent(e, reflect.TypeOf(&Position{})) {
		t.Error("entity should not have Position component initially")
	}

	// Add component
	w.AddComponent(e, &Position{X: 1.0, Y: 2.0})

	// Should now have the component
	if !w.HasComponent(e, reflect.TypeOf(&Position{})) {
		t.Error("entity should have Position component after adding")
	}

	// Should not have other components
	if w.HasComponent(e, reflect.TypeOf(&Velocity{})) {
		t.Error("entity should not have Velocity component")
	}
}

func TestWorld_RemoveComponent(t *testing.T) {
	w := NewWorld()
	e := w.AddEntity()

	pos := &Position{X: 10.0, Y: 20.0}
	vel := &Velocity{DX: 1.0, DY: 1.0}

	w.AddComponent(e, pos)
	w.AddComponent(e, vel)

	// Verify both components exist
	if !w.HasComponent(e, reflect.TypeOf(pos)) {
		t.Fatal("Position component should exist")
	}
	if !w.HasComponent(e, reflect.TypeOf(vel)) {
		t.Fatal("Velocity component should exist")
	}

	// Remove Position
	w.RemoveComponent(e, reflect.TypeOf(pos))

	// Verify Position removed but Velocity remains
	if w.HasComponent(e, reflect.TypeOf(pos)) {
		t.Error("Position component should be removed")
	}
	if !w.HasComponent(e, reflect.TypeOf(vel)) {
		t.Error("Velocity component should still exist")
	}
}

func TestWorld_RemoveComponent_NonExistent(t *testing.T) {
	w := NewWorld()
	e := w.AddEntity()

	// Should not panic when removing non-existent component
	w.RemoveComponent(e, reflect.TypeOf(&Position{}))
}

func TestWorld_RemoveEntity(t *testing.T) {
	w := NewWorld()
	e1 := w.AddEntity()
	e2 := w.AddEntity()

	w.AddComponent(e1, &Position{X: 1.0, Y: 2.0})
	w.AddComponent(e2, &Position{X: 3.0, Y: 4.0})

	// Remove first entity
	w.RemoveEntity(e1)

	// Verify e1 removed
	if _, found := w.components[e1]; found {
		t.Error("entity e1 should be removed")
	}

	// Verify e2 still exists
	if _, found := w.components[e2]; !found {
		t.Error("entity e2 should still exist")
	}
}

func TestWorld_Query(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*World) []Entity
		queryTypes    []reflect.Type
		wantEntityIDs []int
	}{
		{
			name: "query single component",
			setup: func(w *World) []Entity {
				e1 := w.AddEntity()
				e2 := w.AddEntity()
				e3 := w.AddEntity()
				w.AddComponent(e1, &Position{X: 1.0, Y: 1.0})
				w.AddComponent(e2, &Position{X: 2.0, Y: 2.0})
				// e3 has no Position
				return []Entity{e1, e2, e3}
			},
			queryTypes:    []reflect.Type{reflect.TypeOf(&Position{})},
			wantEntityIDs: []int{0, 1},
		},
		{
			name: "query multiple components (AND)",
			setup: func(w *World) []Entity {
				e1 := w.AddEntity()
				e2 := w.AddEntity()
				e3 := w.AddEntity()
				w.AddComponent(e1, &Position{X: 1.0, Y: 1.0})
				w.AddComponent(e1, &Velocity{DX: 0.5, DY: 0.5})
				w.AddComponent(e2, &Position{X: 2.0, Y: 2.0})
				// e2 has Position but not Velocity
				w.AddComponent(e3, &Velocity{DX: 1.0, DY: 1.0})
				// e3 has Velocity but not Position
				return []Entity{e1, e2, e3}
			},
			queryTypes:    []reflect.Type{reflect.TypeOf(&Position{}), reflect.TypeOf(&Velocity{})},
			wantEntityIDs: []int{0},
		},
		{
			name: "query with no matches",
			setup: func(w *World) []Entity {
				e1 := w.AddEntity()
				w.AddComponent(e1, &Position{X: 1.0, Y: 1.0})
				return []Entity{e1}
			},
			queryTypes:    []reflect.Type{reflect.TypeOf(&Health{})},
			wantEntityIDs: []int{},
		},
		{
			name: "query empty world",
			setup: func(w *World) []Entity {
				return []Entity{}
			},
			queryTypes:    []reflect.Type{reflect.TypeOf(&Position{})},
			wantEntityIDs: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			entities := tt.setup(w)
			_ = entities // Keep reference for clarity

			result := w.Query(tt.queryTypes...)

			if len(result) != len(tt.wantEntityIDs) {
				t.Fatalf("got %d entities, want %d", len(result), len(tt.wantEntityIDs))
			}

			// Convert to map for easier checking (order doesn't matter)
			gotIDs := make(map[Entity]bool)
			for _, e := range result {
				gotIDs[e] = true
			}

			for _, wantID := range tt.wantEntityIDs {
				if !gotIDs[Entity(wantID)] {
					t.Errorf("missing entity %d in results", wantID)
				}
			}
		})
	}
}

func TestWorld_SystemExecution(t *testing.T) {
	tests := []struct {
		name            string
		numSystems      int
		numUpdates      int
		wantUpdateCount int
	}{
		{"single system, single update", 1, 1, 1},
		{"single system, multiple updates", 1, 5, 5},
		{"multiple systems, single update", 3, 1, 1},
		{"multiple systems, multiple updates", 3, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			systems := make([]*TestSystem, tt.numSystems)

			for i := 0; i < tt.numSystems; i++ {
				systems[i] = &TestSystem{}
				w.AddSystem(systems[i])
			}

			for i := 0; i < tt.numUpdates; i++ {
				w.Update()
			}

			for i, sys := range systems {
				if sys.UpdateCount != tt.wantUpdateCount {
					t.Errorf("system %d: got %d updates, want %d", i, sys.UpdateCount, tt.wantUpdateCount)
				}
			}
		})
	}
}

func TestWorld_SystemExecutionOrder(t *testing.T) {
	w := NewWorld()

	s1 := &TestSystem{}
	s2 := &TestSystem{}
	s3 := &TestSystem{}

	w.AddSystem(s1)
	w.AddSystem(s2)
	w.AddSystem(s3)

	w.Update()

	// All systems should have been called once
	if s1.UpdateCount != 1 || s2.UpdateCount != 1 || s3.UpdateCount != 1 {
		t.Errorf("systems not called: s1=%d, s2=%d, s3=%d", s1.UpdateCount, s2.UpdateCount, s3.UpdateCount)
	}

	// Call again to verify they all increment
	w.Update()
	if s1.UpdateCount != 2 || s2.UpdateCount != 2 || s3.UpdateCount != 2 {
		t.Errorf("systems not called correctly: s1=%d, s2=%d, s3=%d", s1.UpdateCount, s2.UpdateCount, s3.UpdateCount)
	}
}

func TestWorld_SetGenre(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		expected string
	}{
		{"fantasy", "fantasy", "fantasy"},
		{"scifi", "scifi", "scifi"},
		{"horror", "horror", "horror"},
		{"cyberpunk", "cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			w.SetGenre(tt.genreID)

			got := w.GetGenre()
			if got != tt.expected {
				t.Errorf("got genre %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWorld_SetGenre_Propagation(t *testing.T) {
	w := NewWorld()

	// Default genre should be fantasy
	if w.GetGenre() != "fantasy" {
		t.Errorf("default genre: got %q, want %q", w.GetGenre(), "fantasy")
	}

	// Change to scifi
	w.SetGenre("scifi")
	if w.GetGenre() != "scifi" {
		t.Errorf("after SetGenre: got %q, want %q", w.GetGenre(), "scifi")
	}

	// Create entities and verify genre persists
	e1 := w.AddEntity()
	w.AddComponent(e1, &Position{X: 1.0, Y: 1.0})

	if w.GetGenre() != "scifi" {
		t.Errorf("after adding entities: got %q, want %q", w.GetGenre(), "scifi")
	}
}

func TestWorld_ComponentReplacement(t *testing.T) {
	w := NewWorld()
	e := w.AddEntity()

	// Add initial position
	pos1 := &Position{X: 1.0, Y: 2.0}
	w.AddComponent(e, pos1)

	comp, _ := w.GetComponent(e, reflect.TypeOf(&Position{}))
	if comp.(*Position).X != 1.0 {
		t.Errorf("initial position X: got %f, want 1.0", comp.(*Position).X)
	}

	// Replace with new position
	pos2 := &Position{X: 10.0, Y: 20.0}
	w.AddComponent(e, pos2)

	comp, _ = w.GetComponent(e, reflect.TypeOf(&Position{}))
	if comp.(*Position).X != 10.0 {
		t.Errorf("replaced position X: got %f, want 10.0", comp.(*Position).X)
	}
}

// Benchmark tests
func BenchmarkWorld_AddEntity(b *testing.B) {
	w := NewWorld()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.AddEntity()
	}
}

func BenchmarkWorld_AddComponent(b *testing.B) {
	w := NewWorld()
	entities := make([]Entity, 1000)
	for i := range entities {
		entities[i] = w.AddEntity()
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e := entities[i%len(entities)]
		w.AddComponent(e, &Position{X: float64(i), Y: float64(i)})
	}
}

func BenchmarkWorld_Query(b *testing.B) {
	w := NewWorld()
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &Position{X: float64(i), Y: float64(i)})
		if i%2 == 0 {
			w.AddComponent(e, &Velocity{DX: 1.0, DY: 1.0})
		}
	}

	posType := reflect.TypeOf(&Position{})
	velType := reflect.TypeOf(&Velocity{})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w.Query(posType, velType)
	}
}

// Example test using testutil package for mocking and assertions
func TestWorld_WithTestUtil(t *testing.T) {
	tests := []struct {
		name           string
		setupEntity    func(*World) Entity
		expectedHealth float64
		expectedX      float64
		expectedY      float64
	}{
		{
			name: "player entity with full health",
			setupEntity: func(w *World) Entity {
				e := w.AddEntity()
				w.AddComponent(e, &Position{X: 10.0, Y: 20.0})
				w.AddComponent(e, &Health{Current: 100, Max: 100})
				return e
			},
			expectedHealth: 100.0,
			expectedX:      10.0,
			expectedY:      20.0,
		},
		{
			name: "damaged player entity",
			setupEntity: func(w *World) Entity {
				e := w.AddEntity()
				w.AddComponent(e, &Position{X: 5.5, Y: 15.5})
				w.AddComponent(e, &Health{Current: 50, Max: 100})
				return e
			},
			expectedHealth: 50.0,
			expectedX:      5.5,
			expectedY:      15.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			e := tt.setupEntity(w)

			// Use testutil assertions for cleaner test code
			pos, found := w.GetComponent(e, reflect.TypeOf(&Position{}))
			testutil.AssertTrue(t, found, "Position component should exist")
			testutil.AssertNotNil(t, pos, "Position component should not be nil")

			health, found := w.GetComponent(e, reflect.TypeOf(&Health{}))
			testutil.AssertTrue(t, found, "Health component should exist")
			testutil.AssertNotNil(t, health, "Health component should not be nil")

			// Verify component values
			posComp := pos.(*Position)
			testutil.AssertFloatEqual(t, posComp.X, tt.expectedX, 0.001, "Position X")
			testutil.AssertFloatEqual(t, posComp.Y, tt.expectedY, 0.001, "Position Y")

			healthComp := health.(*Health)
			testutil.AssertFloatEqual(t, float64(healthComp.Current), tt.expectedHealth, 0.001, "Health Current")
		})
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		expected string
	}{
		{"fantasy", "fantasy", "fantasy"},
		{"scifi", "scifi", "scifi"},
		{"horror", "horror", "horror"},
		{"cyberpunk", "cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGenre(tt.genreID)
			if got := GetCurrentGenre(); got != tt.expected {
				t.Errorf("GetCurrentGenre() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSetGenreDefault(t *testing.T) {
	// Reset to fantasy
	SetGenre("fantasy")
	if got := GetCurrentGenre(); got != "fantasy" {
		t.Errorf("Default genre = %v, want fantasy", got)
	}
}

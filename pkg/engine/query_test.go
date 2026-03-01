package engine

import "testing"

func TestQueryWithBitmask(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*World) []Entity
		queryIDs     []ComponentID
		wantEntities int
	}{
		{
			name: "empty world returns no entities",
			setupFunc: func(w *World) []Entity {
				return nil
			},
			queryIDs:     []ComponentID{ComponentIDPosition},
			wantEntities: 0,
		},
		{
			name: "single component query matches entities",
			setupFunc: func(w *World) []Entity {
				e1 := w.AddEntity()
				w.SetArchetype(e1, ComponentIDPosition)
				e2 := w.AddEntity()
				w.SetArchetype(e2, ComponentIDPosition)
				e3 := w.AddEntity()
				w.SetArchetype(e3, ComponentIDVelocity)
				return []Entity{e1, e2}
			},
			queryIDs:     []ComponentID{ComponentIDPosition},
			wantEntities: 2,
		},
		{
			name: "multiple component query matches only entities with all",
			setupFunc: func(w *World) []Entity {
				e1 := w.AddEntity()
				w.SetArchetype(e1, ComponentIDPosition, ComponentIDVelocity)
				e2 := w.AddEntity()
				w.SetArchetype(e2, ComponentIDPosition)
				e3 := w.AddEntity()
				w.SetArchetype(e3, ComponentIDVelocity)
				return []Entity{e1}
			},
			queryIDs:     []ComponentID{ComponentIDPosition, ComponentIDVelocity},
			wantEntities: 1,
		},
		{
			name: "query matches entities with superset of components",
			setupFunc: func(w *World) []Entity {
				e1 := w.AddEntity()
				w.SetArchetype(e1, ComponentIDPosition, ComponentIDVelocity, ComponentIDHealth)
				e2 := w.AddEntity()
				w.SetArchetype(e2, ComponentIDPosition, ComponentIDVelocity)
				return []Entity{e1, e2}
			},
			queryIDs:     []ComponentID{ComponentIDPosition, ComponentIDVelocity},
			wantEntities: 2,
		},
		{
			name: "no entities match query",
			setupFunc: func(w *World) []Entity {
				e1 := w.AddEntity()
				w.SetArchetype(e1, ComponentIDPosition)
				e2 := w.AddEntity()
				w.SetArchetype(e2, ComponentIDVelocity)
				return nil
			},
			queryIDs:     []ComponentID{ComponentIDHealth},
			wantEntities: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld()
			tt.setupFunc(w)

			it := w.QueryWithBitmask(tt.queryIDs...)
			count := 0
			for it.Next() {
				count++
			}

			if count != tt.wantEntities {
				t.Errorf("QueryWithBitmask() matched %d entities, want %d", count, tt.wantEntities)
			}
		})
	}
}

func TestEntityIterator(t *testing.T) {
	entities := []Entity{1, 2, 3, 4, 5}
	it := newEntityIterator(entities)

	t.Run("Next advances through all entities", func(t *testing.T) {
		count := 0
		for it.Next() {
			count++
			if it.Entity() != entities[count-1] {
				t.Errorf("Entity() = %v, want %v", it.Entity(), entities[count-1])
			}
		}
		if count != len(entities) {
			t.Errorf("iterated %d entities, want %d", count, len(entities))
		}
	})

	t.Run("Reset allows re-iteration", func(t *testing.T) {
		it.Reset()
		count := 0
		for it.Next() {
			count++
		}
		if count != len(entities) {
			t.Errorf("after Reset, iterated %d entities, want %d", count, len(entities))
		}
	})

	t.Run("HasNext returns correct values", func(t *testing.T) {
		it.Reset()
		if !it.HasNext() {
			t.Error("HasNext() = false, want true before first Next()")
		}
		for it.Next() {
			// iterate to end
		}
		if it.HasNext() {
			t.Error("HasNext() = true, want false after iteration complete")
		}
	})

	t.Run("empty iterator works correctly", func(t *testing.T) {
		emptyIt := newEntityIterator(nil)
		if emptyIt.Next() {
			t.Error("Next() on empty iterator returned true")
		}
		if emptyIt.HasNext() {
			t.Error("HasNext() on empty iterator returned true")
		}
	})
}

func TestArchetypeManagement(t *testing.T) {
	t.Run("SetArchetype sets correct bitmask", func(t *testing.T) {
		w := NewWorld()
		e := w.AddEntity()
		w.SetArchetype(e, ComponentIDPosition, ComponentIDVelocity)

		mask := w.GetArchetype(e)
		expectedMask := uint64(1<<ComponentIDPosition | 1<<ComponentIDVelocity)
		if mask != expectedMask {
			t.Errorf("GetArchetype() = %064b, want %064b", mask, expectedMask)
		}
	})

	t.Run("AddArchetypeComponent adds component bit", func(t *testing.T) {
		w := NewWorld()
		e := w.AddEntity()
		w.SetArchetype(e, ComponentIDPosition)
		w.AddArchetypeComponent(e, ComponentIDVelocity)

		mask := w.GetArchetype(e)
		expectedMask := uint64(1<<ComponentIDPosition | 1<<ComponentIDVelocity)
		if mask != expectedMask {
			t.Errorf("GetArchetype() = %064b, want %064b", mask, expectedMask)
		}
	})

	t.Run("RemoveArchetypeComponent removes component bit", func(t *testing.T) {
		w := NewWorld()
		e := w.AddEntity()
		w.SetArchetype(e, ComponentIDPosition, ComponentIDVelocity)
		w.RemoveArchetypeComponent(e, ComponentIDVelocity)

		mask := w.GetArchetype(e)
		expectedMask := uint64(1 << ComponentIDPosition)
		if mask != expectedMask {
			t.Errorf("GetArchetype() = %064b, want %064b", mask, expectedMask)
		}
	})

	t.Run("GetArchetype returns 0 for uninitialized entity", func(t *testing.T) {
		w := NewWorld()
		e := w.AddEntity()

		mask := w.GetArchetype(e)
		if mask != 0 {
			t.Errorf("GetArchetype() for new entity = %v, want 0", mask)
		}
	})

	t.Run("handles component IDs >= 64 gracefully", func(t *testing.T) {
		w := NewWorld()
		e := w.AddEntity()
		w.SetArchetype(e, ComponentID(64), ComponentID(128))

		mask := w.GetArchetype(e)
		if mask != 0 {
			t.Errorf("GetArchetype() with invalid IDs = %v, want 0", mask)
		}
	})
}

func TestQueryWithBitmaskFiltering(t *testing.T) {
	w := NewWorld()

	// Create entities with different component combinations
	e1 := w.AddEntity()
	w.SetArchetype(e1, ComponentIDPosition)

	e2 := w.AddEntity()
	w.SetArchetype(e2, ComponentIDPosition, ComponentIDVelocity)

	e3 := w.AddEntity()
	w.SetArchetype(e3, ComponentIDPosition, ComponentIDVelocity, ComponentIDHealth)

	e4 := w.AddEntity()
	w.SetArchetype(e4, ComponentIDHealth, ComponentIDArmor)

	e5 := w.AddEntity()
	w.SetArchetype(e5, ComponentIDPosition, ComponentIDHealth, ComponentIDArmor)

	tests := []struct {
		name     string
		queryIDs []ComponentID
		want     map[Entity]bool
	}{
		{
			name:     "position only",
			queryIDs: []ComponentID{ComponentIDPosition},
			want:     map[Entity]bool{e1: true, e2: true, e3: true, e5: true},
		},
		{
			name:     "position and velocity",
			queryIDs: []ComponentID{ComponentIDPosition, ComponentIDVelocity},
			want:     map[Entity]bool{e2: true, e3: true},
		},
		{
			name:     "health and armor",
			queryIDs: []ComponentID{ComponentIDHealth, ComponentIDArmor},
			want:     map[Entity]bool{e4: true, e5: true},
		},
		{
			name:     "position velocity and health",
			queryIDs: []ComponentID{ComponentIDPosition, ComponentIDVelocity, ComponentIDHealth},
			want:     map[Entity]bool{e3: true},
		},
		{
			name:     "no matches",
			queryIDs: []ComponentID{ComponentIDCamera, ComponentIDInput},
			want:     map[Entity]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			it := w.QueryWithBitmask(tt.queryIDs...)
			got := make(map[Entity]bool)
			for it.Next() {
				got[it.Entity()] = true
			}

			if len(got) != len(tt.want) {
				t.Errorf("QueryWithBitmask() matched %d entities, want %d", len(got), len(tt.want))
			}

			for e := range tt.want {
				if !got[e] {
					t.Errorf("QueryWithBitmask() missing expected entity %v", e)
				}
			}

			for e := range got {
				if !tt.want[e] {
					t.Errorf("QueryWithBitmask() returned unexpected entity %v", e)
				}
			}
		})
	}
}

func BenchmarkQueryWithBitmask(b *testing.B) {
	w := NewWorld()

	// Create 1000 entities with various component combinations
	for i := 0; i < 1000; i++ {
		e := w.AddEntity()
		switch i % 5 {
		case 0:
			w.SetArchetype(e, ComponentIDPosition, ComponentIDVelocity)
		case 1:
			w.SetArchetype(e, ComponentIDPosition, ComponentIDHealth)
		case 2:
			w.SetArchetype(e, ComponentIDPosition, ComponentIDVelocity, ComponentIDHealth)
		case 3:
			w.SetArchetype(e, ComponentIDHealth, ComponentIDArmor)
		case 4:
			w.SetArchetype(e, ComponentIDPosition)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it := w.QueryWithBitmask(ComponentIDPosition, ComponentIDVelocity)
		for it.Next() {
			_ = it.Entity()
		}
	}
}

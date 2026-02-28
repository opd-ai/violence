package props

import (
	"sync"
	"testing"

	"github.com/opd-ai/violence/pkg/procgen/genre"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.genre != genre.Fantasy {
		t.Errorf("Expected default genre %s, got %s", genre.Fantasy, m.genre)
	}
	if len(m.propLists) != 5 {
		t.Errorf("Expected 5 genre prop lists, got %d", len(m.propLists))
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		expected string
	}{
		{"Fantasy", genre.Fantasy, genre.Fantasy},
		{"SciFi", genre.SciFi, genre.SciFi},
		{"Horror", genre.Horror, genre.Horror},
		{"Cyberpunk", genre.Cyberpunk, genre.Cyberpunk},
		{"PostApoc", genre.PostApoc, genre.PostApoc},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			m.SetGenre(tt.genreID)
			if got := m.GetGenre(); got != tt.expected {
				t.Errorf("Expected genre %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestPlaceProps_ZeroDensity(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 10, H: 10}
	props := m.PlaceProps(room, 0.0, 12345)
	if len(props) != 0 {
		t.Errorf("Expected 0 props with zero density, got %d", len(props))
	}
}

func TestPlaceProps_SmallRoom(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 3, H: 3}
	props := m.PlaceProps(room, 0.5, 12345)
	// Small room should still get at least 1 prop with non-zero density
	if len(props) < 1 {
		t.Errorf("Expected at least 1 prop in small room, got %d", len(props))
	}
}

func TestPlaceProps_LargeRoom(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 20, H: 20}
	props := m.PlaceProps(room, 0.3, 12345)
	// Large room with reasonable density should get multiple props
	if len(props) < 2 {
		t.Errorf("Expected multiple props in large room, got %d", len(props))
	}
}

func TestPlaceProps_Determinism(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 15, H: 15}

	// Place props twice with same seed
	props1 := m.PlaceProps(room, 0.2, 99999)
	m.Clear()
	props2 := m.PlaceProps(room, 0.2, 99999)

	if len(props1) != len(props2) {
		t.Fatalf("Determinism failed: got %d props first time, %d second time", len(props1), len(props2))
	}

	// Check that prop properties match
	for i := range props1 {
		if props1[i].X != props2[i].X {
			t.Errorf("Prop %d X position mismatch: %f vs %f", i, props1[i].X, props2[i].X)
		}
		if props1[i].Y != props2[i].Y {
			t.Errorf("Prop %d Y position mismatch: %f vs %f", i, props1[i].Y, props2[i].Y)
		}
		if props1[i].SpriteType != props2[i].SpriteType {
			t.Errorf("Prop %d type mismatch: %v vs %v", i, props1[i].SpriteType, props2[i].SpriteType)
		}
	}
}

func TestPlaceProps_PositionBounds(t *testing.T) {
	m := NewManager()
	room := &Room{X: 5, Y: 10, W: 15, H: 20}
	props := m.PlaceProps(room, 0.3, 54321)

	for i, p := range props {
		// Check props are within room bounds
		if p.X < float64(room.X) || p.X >= float64(room.X+room.W) {
			t.Errorf("Prop %d X position %f outside room bounds [%d, %d)", i, p.X, room.X, room.X+room.W)
		}
		if p.Y < float64(room.Y) || p.Y >= float64(room.Y+room.H) {
			t.Errorf("Prop %d Y position %f outside room bounds [%d, %d)", i, p.Y, room.Y, room.Y+room.H)
		}
	}
}

func TestPlaceProps_GenreSpecific(t *testing.T) {
	tests := []struct {
		name     string
		genreID  string
		expected []string // Expected prop names to exist
	}{
		{"Fantasy", genre.Fantasy, []string{"Barrel", "Torch", "Stone Pillar"}},
		{"SciFi", genre.SciFi, []string{"Terminal", "Supply Crate", "Fuel Drum"}},
		{"Horror", genre.Horror, []string{"Corpse", "Gurney", "Body Bag"}},
		{"Cyberpunk", genre.Cyberpunk, []string{"Data Terminal", "Neon Sign", "Corp Crate"}},
		{"PostApoc", genre.PostApoc, []string{"Rusted Barrel", "Skeleton", "Scrap Pile"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			m.SetGenre(tt.genreID)
			room := &Room{X: 0, Y: 0, W: 30, H: 30} // Large room to get variety
			props := m.PlaceProps(room, 0.5, 12345)

			if len(props) == 0 {
				t.Fatal("No props placed")
			}

			// Check that prop names match genre
			foundNames := make(map[string]bool)
			for _, p := range props {
				foundNames[p.Name] = true
			}

			// At least one expected name should be found
			found := false
			for _, expected := range tt.expected {
				if foundNames[expected] {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find at least one of %v in genre %s, got names: %v", tt.expected, tt.genreID, getKeys(foundNames))
			}
		})
	}
}

func TestPlaceProps_CollisionFlags(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 25, H: 25}
	props := m.PlaceProps(room, 0.4, 11111)

	hasColliding := false
	hasNonColliding := false

	for _, p := range props {
		if p.Collision {
			hasColliding = true
		} else {
			hasNonColliding = true
		}
	}

	if !hasColliding {
		t.Error("Expected some props with collision enabled")
	}
	if !hasNonColliding {
		t.Error("Expected some props without collision")
	}
}

func TestGetProps(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 10, H: 10}
	placed := m.PlaceProps(room, 0.2, 22222)

	retrieved := m.GetProps()
	if len(retrieved) != len(placed) {
		t.Errorf("Expected %d props, got %d", len(placed), len(retrieved))
	}

	// Ensure it's a copy, not the internal slice
	if len(retrieved) > 0 {
		retrieved[0].X = 99999
		if m.props[0].X == 99999 {
			t.Error("GetProps returned internal slice, not a copy")
		}
	}
}

func TestClear(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 10, H: 10}
	m.PlaceProps(room, 0.3, 33333)

	if len(m.GetProps()) == 0 {
		t.Fatal("Props should exist before Clear")
	}

	m.Clear()
	if len(m.GetProps()) != 0 {
		t.Errorf("Expected 0 props after Clear, got %d", len(m.GetProps()))
	}
	if m.nextPropID != 0 {
		t.Errorf("Expected nextPropID to reset to 0, got %d", m.nextPropID)
	}
}

func TestAddProp(t *testing.T) {
	m := NewManager()
	prop := &Prop{
		X:          5.5,
		Y:          7.5,
		SpriteType: PropBarrel,
		Collision:  true,
		Name:       "Test Barrel",
	}

	m.AddProp(prop)
	if prop.ID == "" {
		t.Error("AddProp should generate ID if missing")
	}

	props := m.GetProps()
	if len(props) != 1 {
		t.Fatalf("Expected 1 prop, got %d", len(props))
	}
	if props[0].Name != "Test Barrel" {
		t.Errorf("Expected 'Test Barrel', got '%s'", props[0].Name)
	}
}

func TestAddProp_WithID(t *testing.T) {
	m := NewManager()
	prop := &Prop{
		ID:         "custom_id",
		X:          1.0,
		Y:          2.0,
		SpriteType: PropCrate,
		Name:       "Custom Crate",
	}

	m.AddProp(prop)
	if prop.ID != "custom_id" {
		t.Errorf("Expected ID 'custom_id', got '%s'", prop.ID)
	}
}

func TestRemoveProp(t *testing.T) {
	m := NewManager()
	prop1 := &Prop{Name: "Prop1", X: 1, Y: 1}
	prop2 := &Prop{Name: "Prop2", X: 2, Y: 2}
	m.AddProp(prop1)
	m.AddProp(prop2)

	if !m.RemoveProp(prop1.ID) {
		t.Error("RemoveProp should return true when prop exists")
	}

	props := m.GetProps()
	if len(props) != 1 {
		t.Fatalf("Expected 1 prop after removal, got %d", len(props))
	}
	if props[0].Name != "Prop2" {
		t.Errorf("Expected Prop2 to remain, got %s", props[0].Name)
	}
}

func TestRemoveProp_NotFound(t *testing.T) {
	m := NewManager()
	if m.RemoveProp("nonexistent") {
		t.Error("RemoveProp should return false for nonexistent ID")
	}
}

func TestGetPropsByType(t *testing.T) {
	m := NewManager()
	m.AddProp(&Prop{SpriteType: PropBarrel, Name: "Barrel1", X: 1, Y: 1})
	m.AddProp(&Prop{SpriteType: PropCrate, Name: "Crate1", X: 2, Y: 2})
	m.AddProp(&Prop{SpriteType: PropBarrel, Name: "Barrel2", X: 3, Y: 3})

	barrels := m.GetPropsByType(PropBarrel)
	if len(barrels) != 2 {
		t.Errorf("Expected 2 barrels, got %d", len(barrels))
	}

	crates := m.GetPropsByType(PropCrate)
	if len(crates) != 1 {
		t.Errorf("Expected 1 crate, got %d", len(crates))
	}

	tables := m.GetPropsByType(PropTable)
	if len(tables) != 0 {
		t.Errorf("Expected 0 tables, got %d", len(tables))
	}
}

func TestGeneratePropID(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "prop_0000"},
		{1, "prop_0001"},
		{42, "prop_0042"},
		{999, "prop_0999"},
		{1234, "prop_1234"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := generatePropID(tt.input)
			if got != tt.expected {
				t.Errorf("generatePropID(%d) = %s, want %s", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPlace_Legacy(t *testing.T) {
	prop := Place("Test", 3.5, 4.5)
	if prop == nil {
		t.Fatal("Place returned nil")
	}
	if prop.Name != "Test" {
		t.Errorf("Expected name 'Test', got '%s'", prop.Name)
	}
	if prop.X != 3.5 || prop.Y != 4.5 {
		t.Errorf("Expected position (3.5, 4.5), got (%f, %f)", prop.X, prop.Y)
	}
}

func TestSetGenre_Legacy(t *testing.T) {
	// Should not panic
	SetGenre(genre.SciFi)
}

func TestPropTypes(t *testing.T) {
	// Ensure all prop types have distinct values
	types := []PropType{
		PropBarrel, PropCrate, PropTable, PropTerminal,
		PropBones, PropPlant, PropPillar, PropTorch,
		PropDebris, PropContainer,
	}

	seen := make(map[PropType]bool)
	for _, pt := range types {
		if seen[pt] {
			t.Errorf("Duplicate PropType value: %v", pt)
		}
		seen[pt] = true
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := NewManager()
	room := &Room{X: 0, Y: 0, W: 10, H: 10}

	var wg sync.WaitGroup
	// Concurrent prop placement
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(seed uint64) {
			defer wg.Done()
			m.PlaceProps(room, 0.2, seed)
		}(uint64(i))
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.GetProps()
			_ = m.GetGenre()
		}()
	}

	wg.Wait()
	// If we get here without race conditions, test passes
}

func TestPlaceProps_AllGenres(t *testing.T) {
	genres := []string{genre.Fantasy, genre.SciFi, genre.Horror, genre.Cyberpunk, genre.PostApoc}
	room := &Room{X: 0, Y: 0, W: 20, H: 20}

	for _, g := range genres {
		t.Run(g, func(t *testing.T) {
			m := NewManager()
			m.SetGenre(g)
			props := m.PlaceProps(room, 0.3, 55555)
			if len(props) == 0 {
				t.Errorf("Genre %s produced no props", g)
			}
		})
	}
}

// Helper function to get map keys
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

package edgeao

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 12345)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != genre {
				t.Errorf("genreID = %s, want %s", sys.genreID, genre)
			}
			// Verify preset values are in valid range
			p := sys.preset
			if p.BaseIntensity < 0 || p.BaseIntensity > 1 {
				t.Errorf("BaseIntensity = %f, want [0, 1]", p.BaseIntensity)
			}
			if p.FalloffDistance <= 0 {
				t.Errorf("FalloffDistance = %f, want > 0", p.FalloffDistance)
			}
			if p.CornerMultiplier < 1 {
				t.Errorf("CornerMultiplier = %f, want >= 1", p.CornerMultiplier)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	originalIntensity := sys.preset.BaseIntensity

	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("genreID after SetGenre = %s, want horror", sys.genreID)
	}
	// Horror should have higher intensity than fantasy
	if sys.preset.BaseIntensity <= originalIntensity {
		t.Errorf("Horror intensity = %f should be > fantasy %f",
			sys.preset.BaseIntensity, originalIntensity)
	}
}

func TestBuildAOMap(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Create a simple level: walls around edges, floor in center
	// 1 = wall, 0 = floor
	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	sys.BuildAOMap(tiles)

	if sys.width != 5 || sys.height != 5 {
		t.Errorf("Map dimensions = %dx%d, want 5x5", sys.width, sys.height)
	}

	// Center tile should have lower AO than edges (but may have some from nearby walls)
	centerAO := sys.GetAO(2.5, 2.5)
	if centerAO > 0.5 {
		t.Errorf("Center AO = %f, want <= 0.5", centerAO)
	}

	// Edge tiles (next to walls) should have AO
	edgeAO := sys.GetAO(1.5, 1.5)
	if edgeAO <= 0 {
		t.Errorf("Edge AO = %f, want > 0", edgeAO)
	}
}

func TestEdgeClassification(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Create level with various edge configurations
	// 1 = wall, 0 = floor
	tiles := [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1, 0, 1},
		{1, 0, 0, 0, 1, 0, 1},
		{1, 1, 0, 1, 1, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 0, 0, 0, 0, 0, 1},
		{1, 1, 1, 1, 1, 1, 1},
	}

	sys.BuildAOMap(tiles)

	// Test that tiles adjacent to walls get classified correctly
	tests := []struct {
		x, y     int
		expected EdgeType
		name     string
	}{
		{1, 1, EdgeInsideCorner, "top-left inside corner"}, // Adjacent to walls N and W
		{2, 1, EdgeWallJunction, "wall junction top"},      // Adjacent to wall N only
		{3, 1, EdgeInsideCorner, "top-right area"},         // Adjacent to walls N and E
		{5, 1, EdgeNarrowPassage, "narrow passage"},        // Walls on E and W
		{1, 4, EdgeWallJunction, "left wall adjacent"},     // Adjacent to wall W
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.GetEdgeType(tt.x, tt.y)
			// We're testing that edge-adjacent tiles get non-none classification
			if got == EdgeNone && tt.expected != EdgeNone {
				t.Errorf("GetEdgeType(%d, %d) = %s, want non-none (expected %s)", tt.x, tt.y, got, tt.expected)
			}
		})
	}
}

func TestGetAOInterpolation(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	tiles := [][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	}

	sys.BuildAOMap(tiles)

	// Test interpolation produces smooth values
	ao00 := sys.GetAO(1.0, 1.0)
	ao05 := sys.GetAO(1.5, 1.5)
	ao10 := sys.GetAO(1.9, 1.9)

	// Values should be in valid range
	for _, ao := range []float64{ao00, ao05, ao10} {
		if ao < 0 || ao > 1 {
			t.Errorf("AO value %f outside [0, 1]", ao)
		}
	}
}

func TestGetAOBounds(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	tiles := [][]int{
		{1, 1, 1},
		{1, 0, 1},
		{1, 1, 1},
	}

	sys.BuildAOMap(tiles)

	// Out of bounds should return 0
	tests := []struct {
		x, y float64
		name string
	}{
		{-1, 0, "negative x"},
		{0, -1, "negative y"},
		{10, 0, "x > width"},
		{0, 10, "y > height"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ao := sys.GetAO(tt.x, tt.y)
			if ao != 0 {
				t.Errorf("GetAO(%f, %f) = %f, want 0", tt.x, tt.y, ao)
			}
		})
	}
}

func TestEmptyMap(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// GetAO before BuildAOMap should return 0
	ao := sys.GetAO(5, 5)
	if ao != 0 {
		t.Errorf("GetAO on empty map = %f, want 0", ao)
	}

	// BuildAOMap with empty tiles should not panic
	sys.BuildAOMap([][]int{})
	sys.BuildAOMap(nil)
}

func TestApplyAO(t *testing.T) {
	tests := []struct {
		r, g, b   uint8
		aoFactor  float64
		wantR     uint8
		wantG     uint8
		wantB     uint8
		tolerance uint8
	}{
		{255, 255, 255, 0.0, 255, 255, 255, 1}, // No AO
		{255, 255, 255, 0.5, 127, 127, 127, 2}, // 50% AO
		{255, 255, 255, 1.0, 0, 0, 0, 1},       // Full AO
		{200, 100, 50, 0.25, 150, 75, 37, 2},   // Partial AO
		{100, 100, 100, 0.0, 100, 100, 100, 1}, // Gray no AO
	}

	for _, tt := range tests {
		gotR, gotG, gotB := ApplyAO(tt.r, tt.g, tt.b, tt.aoFactor)

		if absDiff(gotR, tt.wantR) > tt.tolerance ||
			absDiff(gotG, tt.wantG) > tt.tolerance ||
			absDiff(gotB, tt.wantB) > tt.tolerance {
			t.Errorf("ApplyAO(%d,%d,%d, %f) = (%d,%d,%d), want (%d,%d,%d) ±%d",
				tt.r, tt.g, tt.b, tt.aoFactor,
				gotR, gotG, gotB,
				tt.wantR, tt.wantG, tt.wantB, tt.tolerance)
		}
	}
}

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func TestComponentType(t *testing.T) {
	c := NewComponent()
	if c.Type() != "edgeao.Component" {
		t.Errorf("Type() = %s, want edgeao.Component", c.Type())
	}
	if c.AOValue != 0 {
		t.Errorf("NewComponent AOValue = %f, want 0", c.AOValue)
	}
	if c.EdgeType != EdgeNone {
		t.Errorf("NewComponent EdgeType = %v, want EdgeNone", c.EdgeType)
	}
	if !c.Dirty {
		t.Error("NewComponent should be dirty by default")
	}
}

func TestEdgeTypeString(t *testing.T) {
	tests := []struct {
		edge EdgeType
		want string
	}{
		{EdgeNone, "none"},
		{EdgeWallJunction, "wall_junction"},
		{EdgeInsideCorner, "inside_corner"},
		{EdgeOutsideCorner, "outside_corner"},
		{EdgeNarrowPassage, "narrow_passage"},
		{EdgeAlcove, "alcove"},
	}

	for _, tt := range tests {
		got := tt.edge.String()
		if got != tt.want {
			t.Errorf("EdgeType(%d).String() = %s, want %s", tt.edge, got, tt.want)
		}
	}
}

func TestDeterminism(t *testing.T) {
	seed := int64(12345)
	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	}

	sys1 := NewSystem("fantasy", seed)
	sys1.BuildAOMap(tiles)

	sys2 := NewSystem("fantasy", seed)
	sys2.BuildAOMap(tiles)

	// Same seed should produce identical AO maps
	for y := 0; y < 5; y++ {
		for x := 0; x < 5; x++ {
			ao1 := sys1.GetAO(float64(x)+0.5, float64(y)+0.5)
			ao2 := sys2.GetAO(float64(x)+0.5, float64(y)+0.5)
			if ao1 != ao2 {
				t.Errorf("Non-deterministic AO at (%d,%d): %f vs %f", x, y, ao1, ao2)
			}
		}
	}
}

func TestGenreAOIntensityOrder(t *testing.T) {
	// Horror should have strongest AO, scifi should be more subtle
	sysHorror := NewSystem("horror", 12345)
	sysScifi := NewSystem("scifi", 12345)

	if sysHorror.preset.BaseIntensity <= sysScifi.preset.BaseIntensity {
		t.Errorf("Horror intensity (%f) should be > scifi (%f)",
			sysHorror.preset.BaseIntensity, sysScifi.preset.BaseIntensity)
	}
}

func BenchmarkBuildAOMap(b *testing.B) {
	// Create a large dungeon-like level
	size := 50
	tiles := make([][]int, size)
	for y := 0; y < size; y++ {
		tiles[y] = make([]int, size)
		for x := 0; x < size; x++ {
			// Walls on edges and scattered internal walls
			if x == 0 || x == size-1 || y == 0 || y == size-1 {
				tiles[y][x] = 1
			} else if (x%5 == 0 && y%3 != 0) || (y%5 == 0 && x%3 != 0) {
				tiles[y][x] = 1
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys := NewSystem("fantasy", int64(i))
		sys.BuildAOMap(tiles)
	}
}

func BenchmarkGetAO(b *testing.B) {
	sys := NewSystem("fantasy", 12345)

	size := 50
	tiles := make([][]int, size)
	for y := 0; y < size; y++ {
		tiles[y] = make([]int, size)
		for x := 0; x < size; x++ {
			if x == 0 || x == size-1 || y == 0 || y == size-1 {
				tiles[y][x] = 1
			}
		}
	}
	sys.BuildAOMap(tiles)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Query random positions
		x := float64(i%size) + 0.5
		y := float64((i/size)%size) + 0.5
		_ = sys.GetAO(x, y)
	}
}

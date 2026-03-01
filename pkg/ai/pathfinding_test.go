package ai

import (
	"testing"

	"github.com/opd-ai/violence/pkg/level"
)

func TestFindPathCoord_StraightLine(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		start      Coord
		goal       Coord
		expectLen  int
		expectPath []Coord
	}{
		{
			name:       "horizontal path",
			width:      5,
			height:     3,
			start:      Coord{X: 0, Y: 1},
			goal:       Coord{X: 4, Y: 1},
			expectLen:  5,
			expectPath: []Coord{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {4, 1}},
		},
		{
			name:       "vertical path",
			width:      3,
			height:     5,
			start:      Coord{X: 1, Y: 0},
			goal:       Coord{X: 1, Y: 4},
			expectLen:  5,
			expectPath: []Coord{{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}},
		},
		{
			name:       "single step",
			width:      3,
			height:     3,
			start:      Coord{X: 1, Y: 1},
			goal:       Coord{X: 2, Y: 1},
			expectLen:  2,
			expectPath: []Coord{{1, 1}, {2, 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := level.NewTileMap(tt.width, tt.height)
			path := FindPathCoord(*grid, tt.start, tt.goal)

			if len(path) != tt.expectLen {
				t.Errorf("Expected path length %d, got %d", tt.expectLen, len(path))
			}

			if len(path) != len(tt.expectPath) {
				return
			}

			for i := range path {
				if path[i].X != tt.expectPath[i].X || path[i].Y != tt.expectPath[i].Y {
					t.Errorf("Path mismatch at index %d: got (%d,%d), want (%d,%d)",
						i, path[i].X, path[i].Y, tt.expectPath[i].X, tt.expectPath[i].Y)
				}
			}
		})
	}
}

func TestFindPathCoord_AroundObstacles(t *testing.T) {
	// Create a grid with a wall
	// . . . . .
	// . . W . .
	// . . W . .
	// . . W . .
	// . . . . .
	grid := level.NewTileMap(5, 5)
	grid.Set(2, 1, level.TileWall)
	grid.Set(2, 2, level.TileWall)
	grid.Set(2, 3, level.TileWall)

	tests := []struct {
		name      string
		start     Coord
		goal      Coord
		expectMin int
		expectMax int
	}{
		{
			name:      "path around wall",
			start:     Coord{X: 0, Y: 2},
			goal:      Coord{X: 4, Y: 2},
			expectMin: 7, // Must go around, not through
			expectMax: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := FindPathCoord(*grid, tt.start, tt.goal)

			if len(path) == 0 {
				t.Errorf("Expected a path, got empty")
				return
			}

			if len(path) < tt.expectMin || len(path) > tt.expectMax {
				t.Errorf("Expected path length between %d and %d, got %d",
					tt.expectMin, tt.expectMax, len(path))
			}

			// Verify path starts at start and ends at goal
			if path[0].X != tt.start.X || path[0].Y != tt.start.Y {
				t.Errorf("Path should start at (%d,%d), starts at (%d,%d)",
					tt.start.X, tt.start.Y, path[0].X, path[0].Y)
			}

			last := path[len(path)-1]
			if last.X != tt.goal.X || last.Y != tt.goal.Y {
				t.Errorf("Path should end at (%d,%d), ends at (%d,%d)",
					tt.goal.X, tt.goal.Y, last.X, last.Y)
			}

			// Verify path doesn't go through walls
			for _, coord := range path {
				if !grid.IsWalkable(coord.X, coord.Y) {
					t.Errorf("Path goes through non-walkable tile at (%d,%d)", coord.X, coord.Y)
				}
			}
		})
	}
}

func TestFindPathCoord_NoPath(t *testing.T) {
	// Create a grid with complete wall blocking - divide map in half
	// W W W W W
	// W W W W W
	// . . . . .
	// . . . . .
	// . . . . .
	grid := level.NewTileMap(5, 5)
	for x := 0; x < 5; x++ {
		grid.Set(x, 0, level.TileWall)
		grid.Set(x, 1, level.TileWall)
	}

	tests := []struct {
		name  string
		start Coord
		goal  Coord
	}{
		{
			name:  "completely blocked",
			start: Coord{X: 2, Y: 3},
			goal:  Coord{X: 2, Y: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := FindPathCoord(*grid, tt.start, tt.goal)

			if len(path) != 0 {
				t.Errorf("Expected no path, got path with length %d", len(path))
			}
		})
	}
}

func TestFindPathCoord_SamePosition(t *testing.T) {
	grid := level.NewTileMap(5, 5)
	start := Coord{X: 2, Y: 2}
	goal := Coord{X: 2, Y: 2}

	path := FindPathCoord(*grid, start, goal)

	if len(path) != 1 {
		t.Errorf("Expected path of length 1 for same position, got %d", len(path))
	}

	if len(path) > 0 && (path[0].X != start.X || path[0].Y != start.Y) {
		t.Errorf("Path should contain start position (%d,%d), got (%d,%d)",
			start.X, start.Y, path[0].X, path[0].Y)
	}
}

func TestFindPathCoord_OutOfBounds(t *testing.T) {
	grid := level.NewTileMap(5, 5)

	tests := []struct {
		name  string
		start Coord
		goal  Coord
	}{
		{
			name:  "start out of bounds negative",
			start: Coord{X: -1, Y: 2},
			goal:  Coord{X: 2, Y: 2},
		},
		{
			name:  "goal out of bounds high",
			start: Coord{X: 2, Y: 2},
			goal:  Coord{X: 10, Y: 2},
		},
		{
			name:  "both out of bounds",
			start: Coord{X: -1, Y: -1},
			goal:  Coord{X: 10, Y: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := FindPathCoord(*grid, tt.start, tt.goal)

			if len(path) != 0 {
				t.Errorf("Expected empty path for out of bounds, got length %d", len(path))
			}
		})
	}
}

func TestFindPathCoord_NonWalkableStartOrGoal(t *testing.T) {
	grid := level.NewTileMap(5, 5)
	grid.Set(0, 0, level.TileWall)
	grid.Set(4, 4, level.TileWall)

	tests := []struct {
		name  string
		start Coord
		goal  Coord
	}{
		{
			name:  "start on wall",
			start: Coord{X: 0, Y: 0},
			goal:  Coord{X: 2, Y: 2},
		},
		{
			name:  "goal on wall",
			start: Coord{X: 2, Y: 2},
			goal:  Coord{X: 4, Y: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := FindPathCoord(*grid, tt.start, tt.goal)

			if len(path) != 0 {
				t.Errorf("Expected empty path for non-walkable start/goal, got length %d", len(path))
			}
		})
	}
}

func TestFindPathCoord_ThroughDoors(t *testing.T) {
	// Doors should be walkable
	grid := level.NewTileMap(5, 5)
	grid.Set(2, 1, level.TileWall)
	grid.Set(2, 2, level.TileDoor)
	grid.Set(2, 3, level.TileWall)

	start := Coord{X: 0, Y: 2}
	goal := Coord{X: 4, Y: 2}

	path := FindPathCoord(*grid, start, goal)

	if len(path) == 0 {
		t.Errorf("Expected path through door, got empty path")
		return
	}

	// Verify path goes through door
	foundDoor := false
	for _, coord := range path {
		if coord.X == 2 && coord.Y == 2 {
			foundDoor = true
			break
		}
	}

	if !foundDoor {
		t.Errorf("Path should go through door at (2,2)")
	}
}

func TestFindPathCoord_ThroughSecrets(t *testing.T) {
	// Secret tiles should be walkable
	grid := level.NewTileMap(5, 5)
	grid.Set(2, 1, level.TileWall)
	grid.Set(2, 2, level.TileSecret)
	grid.Set(2, 3, level.TileWall)

	start := Coord{X: 0, Y: 2}
	goal := Coord{X: 4, Y: 2}

	path := FindPathCoord(*grid, start, goal)

	if len(path) == 0 {
		t.Errorf("Expected path through secret, got empty path")
		return
	}
}

func TestFindPathCoord_NoDiagonals(t *testing.T) {
	// Verify path uses only 4-directional movement
	grid := level.NewTileMap(5, 5)
	start := Coord{X: 0, Y: 0}
	goal := Coord{X: 2, Y: 2}

	path := FindPathCoord(*grid, start, goal)

	if len(path) == 0 {
		t.Errorf("Expected a path, got empty")
		return
	}

	// Check each step is only 1 tile in X or Y, not both
	for i := 1; i < len(path); i++ {
		prev := path[i-1]
		curr := path[i]
		dx := absInt(curr.X - prev.X)
		dy := absInt(curr.Y - prev.Y)

		// Should move 1 tile in exactly one direction
		if dx+dy != 1 {
			t.Errorf("Diagonal or invalid move detected from (%d,%d) to (%d,%d)",
				prev.X, prev.Y, curr.X, curr.Y)
		}
	}
}

func TestFindPathCoord_ManhattanHeuristic(t *testing.T) {
	// Test that Manhattan distance is used (not Euclidean)
	// by checking path on open grid is shortest Manhattan path
	grid := level.NewTileMap(10, 10)
	start := Coord{X: 0, Y: 0}
	goal := Coord{X: 5, Y: 5}

	path := FindPathCoord(*grid, start, goal)

	// Manhattan distance = |5-0| + |5-0| = 10
	// Shortest path should be 10 + 1 (start tile) = 11
	expectedLen := 11

	if len(path) != expectedLen {
		t.Errorf("Expected shortest Manhattan path length %d, got %d", expectedLen, len(path))
	}
}

func TestManhattanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        Coord
		b        Coord
		expected float64
	}{
		{
			name:     "same point",
			a:        Coord{X: 0, Y: 0},
			b:        Coord{X: 0, Y: 0},
			expected: 0,
		},
		{
			name:     "horizontal",
			a:        Coord{X: 0, Y: 0},
			b:        Coord{X: 5, Y: 0},
			expected: 5,
		},
		{
			name:     "vertical",
			a:        Coord{X: 0, Y: 0},
			b:        Coord{X: 0, Y: 5},
			expected: 5,
		},
		{
			name:     "diagonal",
			a:        Coord{X: 0, Y: 0},
			b:        Coord{X: 3, Y: 4},
			expected: 7,
		},
		{
			name:     "negative coords",
			a:        Coord{X: -2, Y: -3},
			b:        Coord{X: 2, Y: 3},
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manhattanDistance(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("manhattanDistance(%v, %v) = %f, want %f",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{input: 5, expected: 5},
		{input: -5, expected: 5},
		{input: 0, expected: 0},
		{input: 100, expected: 100},
		{input: -100, expected: 100},
	}

	for _, tt := range tests {
		result := absInt(tt.input)
		if result != tt.expected {
			t.Errorf("absInt(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func BenchmarkFindPath_Open(b *testing.B) {
	grid := level.NewTileMap(30, 30)
	start := Coord{X: 0, Y: 0}
	goal := Coord{X: 29, Y: 29}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindPathCoord(*grid, start, goal)
	}
}

func BenchmarkFindPath_Maze(b *testing.B) {
	grid := level.NewTileMap(30, 30)
	// Create a simple maze pattern
	for y := 0; y < 30; y++ {
		for x := 0; x < 30; x++ {
			if x%4 == 2 && y%4 != 1 {
				grid.Set(x, y, level.TileWall)
			}
		}
	}

	start := Coord{X: 0, Y: 0}
	goal := Coord{X: 29, Y: 29}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FindPathCoord(*grid, start, goal)
	}
}

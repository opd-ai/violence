package ai

import (
	"testing"

	"github.com/opd-ai/violence/pkg/level"
)

func TestFindCoverTiles(t *testing.T) {
	tests := []struct {
		name      string
		setupGrid func() level.TileMap
		threatPos Coord
		wantCount int                    // Expected number of cover positions
		validate  func([]CoverTile) bool // Additional validation function
	}{
		{
			name: "simple wall provides cover",
			setupGrid: func() level.TileMap {
				// .....
				// .###.
				// .C T.
				// .....
				tm := level.NewTileMap(5, 4)
				tm.Set(1, 1, level.TileWall)
				tm.Set(2, 1, level.TileWall)
				tm.Set(3, 1, level.TileWall)
				return *tm
			},
			threatPos: Coord{X: 3, Y: 2},
			wantCount: 4, // Multiple positions adjacent to wall provide cover
			validate: func(covers []CoverTile) bool {
				// All found covers should be adjacent to walls
				for _, c := range covers {
					if c.Score <= 0 || c.Score > 1 {
						return false
					}
				}
				return true
			},
		},
		{
			name: "no cover in open area",
			setupGrid: func() level.TileMap {
				// .....
				// .....
				// ..T..
				// .....
				tm := level.NewTileMap(5, 4)
				return *tm
			},
			threatPos: Coord{X: 2, Y: 2},
			wantCount: 0, // No walls, no cover
		},
		{
			name: "multiple cover positions",
			setupGrid: func() level.TileMap {
				// .....
				// .#.#.
				// .CTC.
				// .....
				tm := level.NewTileMap(5, 4)
				tm.Set(1, 1, level.TileWall)
				tm.Set(3, 1, level.TileWall)
				return *tm
			},
			threatPos: Coord{X: 2, Y: 2},
			wantCount: 4, // Multiple positions adjacent to walls
		},
		{
			name: "L-shaped corner",
			setupGrid: func() level.TileMap {
				// #####
				// #...#
				// #.T.#
				// #...#
				// #####
				tm := level.NewTileMap(5, 5)
				// Top wall
				for x := 0; x < 5; x++ {
					tm.Set(x, 0, level.TileWall)
					tm.Set(x, 4, level.TileWall)
				}
				// Side walls
				for y := 0; y < 5; y++ {
					tm.Set(0, y, level.TileWall)
					tm.Set(4, y, level.TileWall)
				}
				return *tm
			},
			threatPos: Coord{X: 2, Y: 2},
			wantCount: 0, // Threat can see all positions in open room
		},
		{
			name: "threat out of bounds",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 5)
				return *tm
			},
			threatPos: Coord{X: 10, Y: 10},
			wantCount: 0,
		},
		{
			name: "cover behind doors",
			setupGrid: func() level.TileMap {
				// .....
				// .#D#.
				// .C T.
				// .....
				tm := level.NewTileMap(5, 4)
				tm.Set(1, 1, level.TileWall)
				tm.Set(2, 1, level.TileDoor)
				tm.Set(3, 1, level.TileWall)
				return *tm
			},
			threatPos: Coord{X: 3, Y: 2},
			wantCount: 2, // Positions adjacent to walls
		},
		{
			name: "diagonal walls",
			setupGrid: func() level.TileMap {
				// #....
				// .#...
				// ..#..
				// T.C#.
				// ....#
				tm := level.NewTileMap(5, 5)
				tm.Set(0, 0, level.TileWall)
				tm.Set(1, 1, level.TileWall)
				tm.Set(2, 2, level.TileWall)
				tm.Set(3, 3, level.TileWall)
				tm.Set(4, 4, level.TileWall)
				return *tm
			},
			threatPos: Coord{X: 0, Y: 3},
			wantCount: 3, // Positions adjacent to diagonal walls with blocked LOS
		},
		{
			name: "large search area",
			setupGrid: func() level.TileMap {
				// Create larger grid with central wall
				tm := level.NewTileMap(50, 50)
				// Vertical wall in center
				for y := 20; y < 30; y++ {
					tm.Set(25, y, level.TileWall)
				}
				return *tm
			},
			threatPos: Coord{X: 30, Y: 25},
			wantCount: 10, // Positions on west side of wall
			validate: func(covers []CoverTile) bool {
				// All cover should be west of wall (x=24)
				for _, c := range covers {
					if c.Position.X != 24 {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := tt.setupGrid()
			got := FindCoverTiles(grid, tt.threatPos)

			if len(got) != tt.wantCount {
				t.Errorf("FindCoverTiles() returned %d covers, want %d", len(got), tt.wantCount)
				t.Logf("Cover positions found: %+v", got)
			}

			// Validate all scores are in valid range
			for _, c := range got {
				if c.Score < 0 || c.Score > 1 {
					t.Errorf("Cover score %f out of range [0,1] at position %+v", c.Score, c.Position)
				}
			}

			// Run custom validation if provided
			if tt.validate != nil && !tt.validate(got) {
				t.Errorf("Custom validation failed for %s", tt.name)
			}
		})
	}
}

func TestIsAdjacentToWall(t *testing.T) {
	tests := []struct {
		name      string
		setupGrid func() level.TileMap
		x, y      int
		want      bool
	}{
		{
			name: "position adjacent to wall north",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				tm.Set(1, 0, level.TileWall)
				return *tm
			},
			x:    1,
			y:    1,
			want: true,
		},
		{
			name: "position adjacent to wall south",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				tm.Set(1, 2, level.TileWall)
				return *tm
			},
			x:    1,
			y:    1,
			want: true,
		},
		{
			name: "position adjacent to wall east",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				tm.Set(2, 1, level.TileWall)
				return *tm
			},
			x:    1,
			y:    1,
			want: true,
		},
		{
			name: "position adjacent to wall west",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				tm.Set(0, 1, level.TileWall)
				return *tm
			},
			x:    1,
			y:    1,
			want: true,
		},
		{
			name: "position not adjacent to wall",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 5)
				tm.Set(0, 0, level.TileWall)
				return *tm
			},
			x:    2,
			y:    2,
			want: false,
		},
		{
			name: "corner position adjacent to diagonal wall",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				tm.Set(2, 2, level.TileWall)
				return *tm
			},
			x:    1,
			y:    1,
			want: false, // Diagonal doesn't count (4-directional only)
		},
		{
			name: "edge of map with wall",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				tm.Set(0, 0, level.TileWall)
				return *tm
			},
			x:    1,
			y:    0,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := tt.setupGrid()
			got := isAdjacentToWall(grid, tt.x, tt.y)
			if got != tt.want {
				t.Errorf("isAdjacentToWall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasLineOfSight(t *testing.T) {
	tests := []struct {
		name      string
		setupGrid func() level.TileMap
		from      Coord
		to        Coord
		want      bool
	}{
		{
			name: "clear horizontal LOS",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 3)
				return *tm
			},
			from: Coord{X: 0, Y: 1},
			to:   Coord{X: 4, Y: 1},
			want: true,
		},
		{
			name: "clear vertical LOS",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 5)
				return *tm
			},
			from: Coord{X: 1, Y: 0},
			to:   Coord{X: 1, Y: 4},
			want: true,
		},
		{
			name: "blocked by wall",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 3)
				tm.Set(2, 1, level.TileWall)
				return *tm
			},
			from: Coord{X: 0, Y: 1},
			to:   Coord{X: 4, Y: 1},
			want: false,
		},
		{
			name: "diagonal LOS clear",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 5)
				return *tm
			},
			from: Coord{X: 0, Y: 0},
			to:   Coord{X: 4, Y: 4},
			want: true,
		},
		{
			name: "diagonal LOS blocked",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 5)
				tm.Set(2, 2, level.TileWall)
				return *tm
			},
			from: Coord{X: 0, Y: 0},
			to:   Coord{X: 4, Y: 4},
			want: false,
		},
		{
			name: "same position",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				return *tm
			},
			from: Coord{X: 1, Y: 1},
			to:   Coord{X: 1, Y: 1},
			want: true,
		},
		{
			name: "LOS through door",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 3)
				tm.Set(2, 1, level.TileDoor)
				return *tm
			},
			from: Coord{X: 0, Y: 1},
			to:   Coord{X: 4, Y: 1},
			want: true, // Doors don't block LOS
		},
		{
			name: "LOS out of bounds",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(3, 3)
				return *tm
			},
			from: Coord{X: 1, Y: 1},
			to:   Coord{X: 10, Y: 10},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := tt.setupGrid()
			got := hasLineOfSight(grid, tt.from, tt.to)
			if got != tt.want {
				t.Errorf("hasLineOfSight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateCoverScore(t *testing.T) {
	tests := []struct {
		name      string
		setupGrid func() level.TileMap
		pos       Coord
		threatPos Coord
		wantRange [2]float64 // [min, max] expected range
	}{
		{
			name: "blocked LOS close distance",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 3)
				tm.Set(2, 1, level.TileWall)
				return *tm
			},
			pos:       Coord{X: 1, Y: 1},
			threatPos: Coord{X: 3, Y: 1},
			wantRange: [2]float64{0.9, 1.0}, // High score: blocked + close
		},
		{
			name: "blocked LOS far distance",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(30, 3)
				tm.Set(15, 1, level.TileWall)
				return *tm
			},
			pos:       Coord{X: 1, Y: 1},
			threatPos: Coord{X: 29, Y: 1},
			wantRange: [2]float64{0.7, 0.8}, // Lower score: blocked but far
		},
		{
			name: "clear LOS",
			setupGrid: func() level.TileMap {
				tm := level.NewTileMap(5, 3)
				return *tm
			},
			pos:       Coord{X: 1, Y: 1},
			threatPos: Coord{X: 3, Y: 1},
			wantRange: [2]float64{0, 0}, // No score: visible to threat
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := tt.setupGrid()
			got := calculateCoverScore(grid, tt.pos, tt.threatPos)
			if got < tt.wantRange[0] || got > tt.wantRange[1] {
				t.Errorf("calculateCoverScore() = %f, want range [%f, %f]", got, tt.wantRange[0], tt.wantRange[1])
			}
		})
	}
}

func TestMaxInt(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{-1, 0, 0},
		{-5, -3, -3},
		{100, 100, 100},
	}

	for _, tt := range tests {
		got := maxInt(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("maxInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		a, b int
		want int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{-1, 0, -1},
		{-5, -3, -5},
		{100, 100, 100},
	}

	for _, tt := range tests {
		got := minInt(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

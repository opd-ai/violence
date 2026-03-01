package level

import "testing"

func TestNewTileMap(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		height     int
		wantWidth  int
		wantHeight int
		wantEmpty  bool
	}{
		{
			name:       "standard size",
			width:      10,
			height:     10,
			wantWidth:  10,
			wantHeight: 10,
			wantEmpty:  false,
		},
		{
			name:       "rectangular map",
			width:      20,
			height:     15,
			wantWidth:  20,
			wantHeight: 15,
			wantEmpty:  false,
		},
		{
			name:       "zero width",
			width:      0,
			height:     10,
			wantWidth:  0,
			wantHeight: 0,
			wantEmpty:  true,
		},
		{
			name:       "zero height",
			width:      10,
			height:     0,
			wantWidth:  0,
			wantHeight: 0,
			wantEmpty:  true,
		},
		{
			name:       "negative dimensions",
			width:      -5,
			height:     -5,
			wantWidth:  0,
			wantHeight: 0,
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTileMap(tt.width, tt.height)

			if tm.Width != tt.wantWidth {
				t.Errorf("Width = %d, want %d", tm.Width, tt.wantWidth)
			}
			if tm.Height != tt.wantHeight {
				t.Errorf("Height = %d, want %d", tm.Height, tt.wantHeight)
			}

			if tt.wantEmpty {
				if len(tm.Tiles) != 0 {
					t.Errorf("expected empty tiles slice, got length %d", len(tm.Tiles))
				}
			} else {
				if len(tm.Tiles) != tt.height {
					t.Errorf("Tiles rows = %d, want %d", len(tm.Tiles), tt.height)
				}
				for row := 0; row < tt.height; row++ {
					if len(tm.Tiles[row]) != tt.width {
						t.Errorf("Tiles[%d] cols = %d, want %d", row, len(tm.Tiles[row]), tt.width)
					}
				}
			}
		})
	}
}

func TestNewTileMap_InitializedToEmpty(t *testing.T) {
	tm := NewTileMap(5, 5)
	for y := 0; y < tm.Height; y++ {
		for x := 0; x < tm.Width; x++ {
			if tm.Tiles[y][x] != TileEmpty {
				t.Errorf("Tiles[%d][%d] = %v, want TileEmpty", y, x, tm.Tiles[y][x])
			}
		}
	}
}

func TestTileMap_Get(t *testing.T) {
	tm := NewTileMap(5, 5)
	tm.Tiles[2][3] = TileWall
	tm.Tiles[1][1] = TileDoor
	tm.Tiles[4][4] = TileSecret

	tests := []struct {
		name string
		x, y int
		want TileType
	}{
		{
			name: "wall tile",
			x:    3,
			y:    2,
			want: TileWall,
		},
		{
			name: "door tile",
			x:    1,
			y:    1,
			want: TileDoor,
		},
		{
			name: "secret tile",
			x:    4,
			y:    4,
			want: TileSecret,
		},
		{
			name: "empty tile",
			x:    0,
			y:    0,
			want: TileEmpty,
		},
		{
			name: "out of bounds negative x",
			x:    -1,
			y:    2,
			want: TileEmpty,
		},
		{
			name: "out of bounds negative y",
			x:    2,
			y:    -1,
			want: TileEmpty,
		},
		{
			name: "out of bounds x too large",
			x:    10,
			y:    2,
			want: TileEmpty,
		},
		{
			name: "out of bounds y too large",
			x:    2,
			y:    10,
			want: TileEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tm.Get(tt.x, tt.y)
			if got != tt.want {
				t.Errorf("Get(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestTileMap_Set(t *testing.T) {
	tests := []struct {
		name      string
		x, y      int
		tileType  TileType
		shouldSet bool
	}{
		{
			name:      "set wall in bounds",
			x:         3,
			y:         2,
			tileType:  TileWall,
			shouldSet: true,
		},
		{
			name:      "set door in bounds",
			x:         1,
			y:         1,
			tileType:  TileDoor,
			shouldSet: true,
		},
		{
			name:      "set at edge",
			x:         4,
			y:         4,
			tileType:  TileSecret,
			shouldSet: true,
		},
		{
			name:      "set out of bounds negative x",
			x:         -1,
			y:         2,
			tileType:  TileWall,
			shouldSet: false,
		},
		{
			name:      "set out of bounds x too large",
			x:         10,
			y:         2,
			tileType:  TileWall,
			shouldSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tm := NewTileMap(5, 5)
			tm.Set(tt.x, tt.y, tt.tileType)

			if tt.shouldSet {
				got := tm.Get(tt.x, tt.y)
				if got != tt.tileType {
					t.Errorf("after Set(%d, %d, %v), Get() = %v", tt.x, tt.y, tt.tileType, got)
				}
			}
		})
	}
}

func TestTileMap_IsWalkable(t *testing.T) {
	tm := NewTileMap(5, 5)
	tm.Set(0, 0, TileEmpty)
	tm.Set(1, 1, TileWall)
	tm.Set(2, 2, TileDoor)
	tm.Set(3, 3, TileSecret)

	tests := []struct {
		name string
		x, y int
		want bool
	}{
		{
			name: "empty tile is walkable",
			x:    0,
			y:    0,
			want: true,
		},
		{
			name: "wall is not walkable",
			x:    1,
			y:    1,
			want: false,
		},
		{
			name: "door is walkable",
			x:    2,
			y:    2,
			want: true,
		},
		{
			name: "secret is walkable",
			x:    3,
			y:    3,
			want: true,
		},
		{
			name: "out of bounds returns true (TileEmpty)",
			x:    -1,
			y:    0,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tm.IsWalkable(tt.x, tt.y)
			if got != tt.want {
				t.Errorf("IsWalkable(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestTileMap_InBounds(t *testing.T) {
	tm := NewTileMap(10, 8)

	tests := []struct {
		name string
		x, y int
		want bool
	}{
		{
			name: "origin is in bounds",
			x:    0,
			y:    0,
			want: true,
		},
		{
			name: "center is in bounds",
			x:    5,
			y:    4,
			want: true,
		},
		{
			name: "max coordinates in bounds",
			x:    9,
			y:    7,
			want: true,
		},
		{
			name: "negative x out of bounds",
			x:    -1,
			y:    4,
			want: false,
		},
		{
			name: "negative y out of bounds",
			x:    5,
			y:    -1,
			want: false,
		},
		{
			name: "x too large",
			x:    10,
			y:    4,
			want: false,
		},
		{
			name: "y too large",
			x:    5,
			y:    8,
			want: false,
		},
		{
			name: "both out of bounds",
			x:    -5,
			y:    -5,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tm.InBounds(tt.x, tt.y)
			if got != tt.want {
				t.Errorf("InBounds(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestTileType_Constants(t *testing.T) {
	// Ensure constants have expected values for serialization compatibility
	if TileEmpty != 0 {
		t.Errorf("TileEmpty = %d, want 0", TileEmpty)
	}
	if TileWall != 1 {
		t.Errorf("TileWall = %d, want 1", TileWall)
	}
	if TileDoor != 2 {
		t.Errorf("TileDoor = %d, want 2", TileDoor)
	}
	if TileSecret != 3 {
		t.Errorf("TileSecret = %d, want 3", TileSecret)
	}
}

func BenchmarkTileMap_Get(b *testing.B) {
	tm := NewTileMap(100, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.Get(50, 50)
	}
}

func BenchmarkTileMap_Set(b *testing.B) {
	tm := NewTileMap(100, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.Set(50, 50, TileWall)
	}
}

func BenchmarkTileMap_IsWalkable(b *testing.B) {
	tm := NewTileMap(100, 100)
	tm.Set(50, 50, TileWall)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.IsWalkable(50, 50)
	}
}

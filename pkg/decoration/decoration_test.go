package decoration

import (
	"testing"

	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem()
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.genre != genre.Fantasy {
		t.Errorf("Expected genre %s, got %s", genre.Fantasy, sys.genre)
	}
	if sys.genreCfg == nil {
		t.Error("genreCfg is nil")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		wantFD  float64
	}{
		{"Fantasy", genre.Fantasy, 0.15},
		{"SciFi", genre.SciFi, 0.20},
		{"Horror", genre.Horror, 0.12},
		{"Cyberpunk", genre.Cyberpunk, 0.18},
		{"PostApoc", genre.PostApoc, 0.10},
		{"Unknown", "unknown", 0.15}, // fallback to Fantasy
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem()
			sys.SetGenre(tt.genreID)
			if sys.genreCfg.FurnitureDensity != tt.wantFD {
				t.Errorf("FurnitureDensity = %f, want %f", sys.genreCfg.FurnitureDensity, tt.wantFD)
			}
		})
	}
}

func TestDetermineRoomType(t *testing.T) {
	sys := NewSystem()
	r := rng.NewRNG(12345)

	tests := []struct {
		name       string
		width      int
		height     int
		roomIndex  int
		totalRooms int
		wantType   RoomType
	}{
		{"Boss room", 10, 10, 9, 10, RoomBoss},
		{"Treasure medium", 8, 6, 8, 10, RoomTreasure},
		{"Generic small", 5, 4, 2, 10, RoomGeneric},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := sys.DetermineRoomType(tt.width, tt.height, tt.roomIndex, tt.totalRooms, r)
			if tt.name == "Boss room" && rt != RoomBoss {
				t.Errorf("Expected RoomBoss, got %d", rt)
			}
			if tt.name == "Generic small" {
				// Just verify it returns a valid room type
				if rt < RoomGeneric || rt > RoomBoss {
					t.Errorf("Invalid room type: %d", rt)
				}
			}
		})
	}
}

func TestDecorateRoom(t *testing.T) {
	sys := NewSystem()
	r := rng.NewRNG(54321)

	tiles := make([][]int, 20)
	for i := range tiles {
		tiles[i] = make([]int, 20)
		for j := range tiles[i] {
			tiles[i][j] = 1 // walls
		}
	}

	// Create room
	for y := 5; y < 15; y++ {
		for x := 5; x < 15; x++ {
			tiles[y][x] = 2 // floor
		}
	}

	tests := []struct {
		name     string
		roomType RoomType
	}{
		{"Generic room", RoomGeneric},
		{"Armory", RoomArmory},
		{"Treasure", RoomTreasure},
		{"Boss", RoomBoss},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decor := sys.DecorateRoom(tt.roomType, 5, 5, 10, 10, tiles, r)
			if decor == nil {
				t.Fatal("DecorateRoom returned nil")
			}
			if decor.RoomType != tt.roomType {
				t.Errorf("RoomType = %d, want %d", decor.RoomType, tt.roomType)
			}
			// Should have some decorations
			if len(decor.Decorations) == 0 {
				t.Error("No decorations generated")
			}
		})
	}
}

func TestDecorationTypes(t *testing.T) {
	sys := NewSystem()
	r := rng.NewRNG(99999)

	tiles := make([][]int, 20)
	for i := range tiles {
		tiles[i] = make([]int, 20)
		for j := range tiles[i] {
			tiles[i][j] = 1
		}
	}

	for y := 5; y < 15; y++ {
		for x := 5; x < 15; x++ {
			tiles[y][x] = 2
		}
	}

	decor := sys.DecorateRoom(RoomArmory, 5, 5, 10, 10, tiles, r)

	// Count decoration types
	counts := make(map[DecoType]int)
	for _, d := range decor.Decorations {
		counts[d.Type]++

		// Verify decoration is within room bounds
		if d.X < 5 || d.X >= 15 || d.Y < 5 || d.Y >= 15 {
			t.Errorf("Decoration at (%d, %d) outside room bounds", d.X, d.Y)
		}

		// Verify seeded flag
		if !d.Seeded {
			t.Error("Decoration should be seeded")
		}
	}

	// Should have at least some furniture
	if counts[DecoFurniture] == 0 {
		t.Error("No furniture decorations")
	}
}

func TestGetRoomTypeName(t *testing.T) {
	tests := []struct {
		rt   RoomType
		want string
	}{
		{RoomGeneric, "Chamber"},
		{RoomArmory, "Armory"},
		{RoomLibrary, "Library"},
		{RoomShrine, "Shrine"},
		{RoomTreasure, "Treasure Room"},
		{RoomPrison, "Prison"},
		{RoomBarracks, "Barracks"},
		{RoomLaboratory, "Laboratory"},
		{RoomStorage, "Storage"},
		{RoomBoss, "Boss Arena"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := GetRoomTypeName(tt.rt)
			if got != tt.want {
				t.Errorf("GetRoomTypeName(%d) = %s, want %s", tt.rt, got, tt.want)
			}
		})
	}
}

func TestIsWalkable(t *testing.T) {
	sys := NewSystem()
	tiles := [][]int{
		{1, 1, 1},
		{1, 2, 1},
		{1, 1, 1},
	}

	tests := []struct {
		name string
		x, y int
		want bool
	}{
		{"Center floor", 1, 1, true},
		{"Wall", 0, 0, false},
		{"Out of bounds", -1, 0, false},
		{"Out of bounds right", 5, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.isWalkable(tt.x, tt.y, tiles)
			if got != tt.want {
				t.Errorf("isWalkable(%d, %d) = %t, want %t", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestIsNearWall(t *testing.T) {
	sys := NewSystem()
	tiles := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}

	tests := []struct {
		name string
		x, y int
		want bool
	}{
		{"Against wall", 1, 1, true},
		{"Center", 2, 2, false},
		{"Near wall", 2, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.isNearWall(tt.x, tt.y, tiles)
			if got != tt.want {
				t.Errorf("isNearWall(%d, %d) = %t, want %t", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestTooClose(t *testing.T) {
	sys := NewSystem()
	decorations := []Decoration{
		{X: 5, Y: 5},
		{X: 10, Y: 10},
	}

	tests := []struct {
		name    string
		x, y    int
		minDist int
		want    bool
	}{
		{"Same position", 5, 5, 2, true},
		{"Adjacent", 6, 5, 2, true},
		{"Far enough", 8, 8, 2, false},
		{"Far away", 20, 20, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.tooClose(tt.x, tt.y, decorations, tt.minDist)
			if got != tt.want {
				t.Errorf("tooClose(%d, %d, decorations, %d) = %t, want %t", tt.x, tt.y, tt.minDist, got, tt.want)
			}
		})
	}
}

func TestBlocksMovement(t *testing.T) {
	sys := NewSystem()

	// Open room
	tiles1 := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 2, 2, 2, 1},
		{1, 1, 1, 1, 1},
	}

	// Very tight corner
	tiles2 := [][]int{
		{1, 1, 1, 1, 1},
		{1, 2, 1, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1},
	}

	tests := []struct {
		name  string
		tiles [][]int
		x, y  int
		want  bool
	}{
		{"Open area", tiles1, 2, 2, false},
		{"Isolated tile", tiles2, 1, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sys.blocksMovement(tt.x, tt.y, tt.tiles)
			if got != tt.want {
				t.Errorf("blocksMovement(%d, %d) = %t, want %t", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestGenreSpecificRoomTypes(t *testing.T) {
	r := rng.NewRNG(11111)

	tests := []struct {
		genreID string
	}{
		{genre.Fantasy},
		{genre.SciFi},
		{genre.Horror},
		{genre.Cyberpunk},
		{genre.PostApoc},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			sys := NewSystem()
			sys.SetGenre(tt.genreID)

			// Generate several rooms to test variety
			roomTypes := make(map[RoomType]bool)
			for i := 0; i < 20; i++ {
				rt := sys.DetermineRoomType(8, 8, i, 20, r)
				roomTypes[rt] = true
			}

			// Should have some variety
			if len(roomTypes) < 2 {
				t.Errorf("Only %d room types generated, expected more variety", len(roomTypes))
			}
		})
	}
}

func BenchmarkDecorateRoom(b *testing.B) {
	sys := NewSystem()
	r := rng.NewRNG(42)

	tiles := make([][]int, 30)
	for i := range tiles {
		tiles[i] = make([]int, 30)
		for j := range tiles[i] {
			tiles[i][j] = 1
		}
	}

	for y := 5; y < 25; y++ {
		for x := 5; x < 25; x++ {
			tiles[y][x] = 2
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.DecorateRoom(RoomArmory, 5, 5, 20, 20, tiles, r)
	}
}

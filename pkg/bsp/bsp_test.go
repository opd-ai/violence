package bsp

import (
	"testing"

	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
)

func TestNewGenerator(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"small", 32, 32},
		{"medium", 64, 64},
		{"large", 128, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(12345)
			g := NewGenerator(tt.width, tt.height, r)
			if g.Width != tt.width {
				t.Errorf("Width = %d, want %d", g.Width, tt.width)
			}
			if g.Height != tt.height {
				t.Errorf("Height = %d, want %d", g.Height, tt.height)
			}
			if g.MinSize != 6 {
				t.Errorf("MinSize = %d, want 6", g.MinSize)
			}
			if g.MaxSize != 12 {
				t.Errorf("MaxSize = %d, want 12", g.MaxSize)
			}
		})
	}
}

func TestGenerateDeterministic(t *testing.T) {
	tests := []struct {
		name string
		seed uint64
	}{
		{"seed1", 12345},
		{"seed2", 67890},
		{"seed3", 11111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r1 := rng.NewRNG(tt.seed)
			g1 := NewGenerator(64, 64, r1)
			_, tiles1 := g1.Generate()

			r2 := rng.NewRNG(tt.seed)
			g2 := NewGenerator(64, 64, r2)
			_, tiles2 := g2.Generate()

			if !tilesEqual(tiles1, tiles2) {
				t.Error("Same seed produced different maps")
			}
		})
	}
}

func TestGenerateDifferentSeeds(t *testing.T) {
	r1 := rng.NewRNG(12345)
	g1 := NewGenerator(64, 64, r1)
	_, tiles1 := g1.Generate()

	r2 := rng.NewRNG(67890)
	g2 := NewGenerator(64, 64, r2)
	_, tiles2 := g2.Generate()

	if tilesEqual(tiles1, tiles2) {
		t.Error("Different seeds produced identical maps")
	}
}

func TestGenerateRoomCount(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		seed   uint64
	}{
		{"small", 32, 32, 12345},
		{"medium", 64, 64, 67890},
		{"large", 128, 128, 11111},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(tt.seed)
			g := NewGenerator(tt.width, tt.height, r)
			root, _ := g.Generate()

			rooms := countRooms(root)
			if rooms < 2 {
				t.Errorf("Too few rooms: %d", rooms)
			}
			if rooms > 100 {
				t.Errorf("Too many rooms: %d (expected <= 100)", rooms)
			}
		})
	}
}

func TestGenerateConnectivity(t *testing.T) {
	r := rng.NewRNG(12345)
	g := NewGenerator(64, 64, r)
	_, tiles := g.Generate()

	// Find first floor tile
	startX, startY := -1, -1
	for y := 0; y < 64 && startX == -1; y++ {
		for x := 0; x < 64; x++ {
			if tiles[y][x] == TileFloor || tiles[y][x] == TileDoor {
				startX, startY = x, y
				break
			}
		}
	}

	if startX == -1 {
		t.Fatal("No floor tiles found")
	}

	// Flood fill to count reachable tiles
	visited := make([][]bool, 64)
	for i := range visited {
		visited[i] = make([]bool, 64)
	}

	reachable := floodFill(tiles, visited, startX, startY)
	totalFloor := countFloorTiles(tiles)

	// At least 80% of floor tiles should be reachable
	if float64(reachable)/float64(totalFloor) < 0.8 {
		t.Errorf("Connectivity too low: %d/%d reachable (%.1f%%)",
			reachable, totalFloor, 100.0*float64(reachable)/float64(totalFloor))
	}
}

func TestGenerateDoorPlacement(t *testing.T) {
	r := rng.NewRNG(12345)
	g := NewGenerator(64, 64, r)
	_, tiles := g.Generate()

	doorCount := 0
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if tiles[y][x] == TileDoor {
				doorCount++
			}
		}
	}

	if doorCount < 1 {
		t.Error("No doors placed")
	}
	if doorCount > 50 {
		t.Errorf("Too many doors: %d (expected <= 50)", doorCount)
	}
}

func TestGenerateSecretPlacement(t *testing.T) {
	r := rng.NewRNG(12345)
	g := NewGenerator(64, 64, r)
	_, tiles := g.Generate()

	secretCount := 0
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if tiles[y][x] == TileSecret {
				secretCount++
			}
		}
	}

	// Secrets are optional, but should be rare
	if secretCount > 10 {
		t.Errorf("Too many secrets: %d", secretCount)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name          string
		genreID       string
		expectedWall  int
		expectedFloor int
	}{
		{"fantasy", genre.Fantasy, TileWallStone, TileFloorStone},
		{"scifi", genre.SciFi, TileWallHull, TileFloorHull},
		{"horror", genre.Horror, TileWallPlaster, TileFloorWood},
		{"cyberpunk", genre.Cyberpunk, TileWallConcrete, TileFloorConcrete},
		{"postapoc", genre.PostApoc, TileWallRust, TileFloorDirt},
		{"unknown", "unknown", TileWall, TileFloor}, // Fallback to generic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(12345)
			g := NewGenerator(64, 64, r)
			g.SetGenre(tt.genreID)
			if g.genre != tt.genreID {
				t.Errorf("genre = %s, want %s", g.genre, tt.genreID)
			}
			if g.wallTile != tt.expectedWall {
				t.Errorf("wallTile = %d, want %d", g.wallTile, tt.expectedWall)
			}
			if g.floorTile != tt.expectedFloor {
				t.Errorf("floorTile = %d, want %d", g.floorTile, tt.expectedFloor)
			}
		})
	}
}

// TestGenreTileGeneration verifies generated maps use genre-specific tiles.
func TestGenreTileGeneration(t *testing.T) {
	tests := []struct {
		name          string
		genreID       string
		expectedWall  int
		expectedFloor int
	}{
		{"fantasy_tiles", genre.Fantasy, TileWallStone, TileFloorStone},
		{"scifi_tiles", genre.SciFi, TileWallHull, TileFloorHull},
		{"horror_tiles", genre.Horror, TileWallPlaster, TileFloorWood},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rng.NewRNG(54321)
			g := NewGenerator(32, 32, r)
			g.SetGenre(tt.genreID)
			_, tiles := g.Generate()

			// Verify at least some tiles use genre-specific wall type
			hasGenreWall := false
			hasGenreFloor := false

			for y := range tiles {
				for x := range tiles[y] {
					if tiles[y][x] == tt.expectedWall {
						hasGenreWall = true
					}
					if tiles[y][x] == tt.expectedFloor {
						hasGenreFloor = true
					}
				}
			}

			if !hasGenreWall {
				t.Errorf("Generated map has no genre-specific wall tiles (expected %d)", tt.expectedWall)
			}
			if !hasGenreFloor {
				t.Errorf("Generated map has no genre-specific floor tiles (expected %d)", tt.expectedFloor)
			}
		})
	}
}

func TestSplitSmallNode(t *testing.T) {
	r := rng.NewRNG(12345)
	g := NewGenerator(64, 64, r)

	// Too small to split
	node := &Node{X: 0, Y: 0, W: 8, H: 8}
	result := g.split(node, 0)

	if result {
		t.Error("Expected split to fail for small node")
	}
	if node.Left != nil || node.Right != nil {
		t.Error("Small node should not have children")
	}
}

func TestSplitLargeNode(t *testing.T) {
	r := rng.NewRNG(12345)
	g := NewGenerator(64, 64, r)

	node := &Node{X: 0, Y: 0, W: 64, H: 64}
	result := g.split(node, 0)

	if !result {
		t.Error("Expected split to succeed for large node")
	}
	if node.Left == nil || node.Right == nil {
		t.Error("Large node should have children")
	}
}

func TestSplitMaxDepth(t *testing.T) {
	r := rng.NewRNG(12345)
	g := NewGenerator(1024, 1024, r)

	node := &Node{X: 0, Y: 0, W: 1024, H: 1024}
	result := g.split(node, 11) // Beyond max depth

	if result {
		t.Error("Expected split to fail at max depth")
	}
}

func TestCorridorCarving(t *testing.T) {
	r := rng.NewRNG(12345)
	g := NewGenerator(64, 64, r)

	tiles := make([][]int, 64)
	for y := range tiles {
		tiles[y] = make([]int, 64)
		for x := range tiles[y] {
			tiles[y][x] = TileWall
		}
	}

	g.carveCorridor(10, 10, 20, 10, tiles)
	g.carveCorridor(20, 10, 20, 20, tiles)

	// Check horizontal corridor
	for x := 10; x <= 20; x++ {
		if tiles[10][x] != TileFloor {
			t.Errorf("Horizontal corridor not carved at (%d, 10)", x)
		}
	}

	// Check vertical corridor
	for y := 10; y <= 20; y++ {
		if tiles[y][20] != TileFloor {
			t.Errorf("Vertical corridor not carved at (20, %d)", y)
		}
	}
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		name    string
		a, b    int
		wantMin int
		wantMax int
	}{
		{"positive", 5, 10, 5, 10},
		{"reverse", 10, 5, 5, 10},
		{"equal", 7, 7, 7, 7},
		{"negative", -5, 3, -5, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := min(tt.a, tt.b); got != tt.wantMin {
				t.Errorf("min(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.wantMin)
			}
			if got := max(tt.a, tt.b); got != tt.wantMax {
				t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.wantMax)
			}
		})
	}
}

// Helper functions

func tilesEqual(t1, t2 [][]int) bool {
	if len(t1) != len(t2) {
		return false
	}
	for y := range t1 {
		if len(t1[y]) != len(t2[y]) {
			return false
		}
		for x := range t1[y] {
			if t1[y][x] != t2[y][x] {
				return false
			}
		}
	}
	return true
}

func countRooms(n *Node) int {
	if n == nil {
		return 0
	}
	if n.Room != nil {
		return 1
	}
	return countRooms(n.Left) + countRooms(n.Right)
}

func floodFill(tiles [][]int, visited [][]bool, x, y int) int {
	if y < 0 || y >= len(tiles) || x < 0 || x >= len(tiles[0]) {
		return 0
	}
	if visited[y][x] {
		return 0
	}
	if tiles[y][x] == TileWall || tiles[y][x] == TileSecret {
		return 0
	}

	visited[y][x] = true
	count := 1

	count += floodFill(tiles, visited, x+1, y)
	count += floodFill(tiles, visited, x-1, y)
	count += floodFill(tiles, visited, x, y+1)
	count += floodFill(tiles, visited, x, y-1)

	return count
}

func countFloorTiles(tiles [][]int) int {
	count := 0
	for y := range tiles {
		for x := range tiles[y] {
			if tiles[y][x] == TileFloor || tiles[y][x] == TileDoor {
				count++
			}
		}
	}
	return count
}

// Benchmarks

func BenchmarkGenerate(b *testing.B) {
	r := rng.NewRNG(12345)
	g := NewGenerator(64, 64, r)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Seed(12345)
		g.Generate()
	}
}

func BenchmarkGenerateLarge(b *testing.B) {
	r := rng.NewRNG(12345)
	g := NewGenerator(128, 128, r)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Seed(12345)
		g.Generate()
	}
}

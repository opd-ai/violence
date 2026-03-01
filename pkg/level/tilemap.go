// Package level provides level generation and tile-based map structures.
package level

// TileType represents a type of tile in the game world.
type TileType uint8

// Tile type constants.
const (
	TileEmpty TileType = iota
	TileWall
	TileDoor
	TileSecret
)

// TileMap represents a 2D grid of tiles using row-major indexing.
// Access tiles via: tilemap.Tiles[row][col]
type TileMap struct {
	Tiles  [][]TileType
	Width  int
	Height int
}

// NewTileMap creates a new TileMap with the specified dimensions.
// All tiles are initialized to TileEmpty.
func NewTileMap(width, height int) *TileMap {
	if width <= 0 || height <= 0 {
		return &TileMap{
			Tiles:  [][]TileType{},
			Width:  0,
			Height: 0,
		}
	}

	tiles := make([][]TileType, height)
	for row := range tiles {
		tiles[row] = make([]TileType, width)
	}

	return &TileMap{
		Tiles:  tiles,
		Width:  width,
		Height: height,
	}
}

// Get returns the tile at the specified coordinates.
// Returns TileEmpty if coordinates are out of bounds.
func (tm *TileMap) Get(x, y int) TileType {
	if x < 0 || y < 0 || y >= tm.Height || x >= tm.Width {
		return TileEmpty
	}
	return tm.Tiles[y][x]
}

// Set sets the tile at the specified coordinates.
// Does nothing if coordinates are out of bounds.
func (tm *TileMap) Set(x, y int, tile TileType) {
	if x < 0 || y < 0 || y >= tm.Height || x >= tm.Width {
		return
	}
	tm.Tiles[y][x] = tile
}

// IsWalkable returns true if the tile at (x, y) can be walked through.
// Walls are not walkable; empty tiles, doors, and secrets are walkable.
func (tm *TileMap) IsWalkable(x, y int) bool {
	tile := tm.Get(x, y)
	return tile != TileWall
}

// InBounds returns true if the coordinates are within the map boundaries.
func (tm *TileMap) InBounds(x, y int) bool {
	return x >= 0 && y >= 0 && x < tm.Width && y < tm.Height
}

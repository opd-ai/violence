// Package bsp provides BSP-based procedural level generation.
package bsp

import (
	"github.com/opd-ai/violence/pkg/procgen/genre"
	"github.com/opd-ai/violence/pkg/rng"
)

const (
	TileEmpty  = 0
	TileWall   = 1
	TileFloor  = 2
	TileDoor   = 3
	TileSecret = 4
)

// Node represents a BSP tree node used during level generation.
type Node struct {
	X, Y, W, H  int
	Left, Right *Node
	Room        *Room // Leaf nodes have rooms
}

// Room represents a rectangular room within a BSP node.
type Room struct {
	X, Y, W, H int
}

// Generator produces levels using binary space partitioning.
type Generator struct {
	Width     int
	Height    int
	MinSize   int
	MaxSize   int
	rng       *rng.RNG
	genre     string
	wallTile  int
	floorTile int
}

// GeneratorConfig holds BSP generation parameters.
type GeneratorConfig struct {
	MinRoomSize int
	MaxRoomSize int
}

// NewGenerator creates a BSP generator for the given dimensions.
func NewGenerator(width, height int, r *rng.RNG) *Generator {
	return &Generator{
		Width:     width,
		Height:    height,
		MinSize:   6,
		MaxSize:   12,
		rng:       r,
		genre:     genre.Fantasy,
		wallTile:  TileWall,
		floorTile: TileFloor,
	}
}

// SetGenre configures level generation parameters for a genre.
func (g *Generator) SetGenre(genreID string) {
	g.genre = genreID
	// Genre-specific wall/floor tiles can be set here
	// For now, all genres use the same tile constants
}

// Generate produces a BSP tree and tile map.
func (g *Generator) Generate() (*Node, [][]int) {
	root := &Node{X: 0, Y: 0, W: g.Width, H: g.Height}
	g.split(root, 0)

	tiles := make([][]int, g.Height)
	for y := range tiles {
		tiles[y] = make([]int, g.Width)
		for x := range tiles[y] {
			tiles[y][x] = g.wallTile
		}
	}

	g.createRooms(root, tiles)
	g.createCorridors(root, tiles)
	g.placeDoors(root, tiles)
	g.placeSecrets(root, tiles)

	return root, tiles
}

// split recursively partitions space into smaller nodes.
func (g *Generator) split(n *Node, depth int) bool {
	if depth > 10 { // Prevent infinite recursion
		return false
	}

	// Stop splitting if too small
	if n.W < g.MinSize*2 || n.H < g.MinSize*2 {
		return false
	}

	// Decide split direction
	horizontal := g.rng.Intn(2) == 0
	if n.W > n.H && float64(n.W)/float64(n.H) >= 1.25 {
		horizontal = false
	} else if n.H > n.W && float64(n.H)/float64(n.W) >= 1.25 {
		horizontal = true
	}

	if horizontal {
		// Split horizontally
		maxSplit := n.H - g.MinSize*2
		if maxSplit <= 0 {
			return false
		}
		splitPos := g.MinSize + g.rng.Intn(maxSplit)
		n.Left = &Node{X: n.X, Y: n.Y, W: n.W, H: splitPos}
		n.Right = &Node{X: n.X, Y: n.Y + splitPos, W: n.W, H: n.H - splitPos}
	} else {
		// Split vertically
		maxSplit := n.W - g.MinSize*2
		if maxSplit <= 0 {
			return false
		}
		splitPos := g.MinSize + g.rng.Intn(maxSplit)
		n.Left = &Node{X: n.X, Y: n.Y, W: splitPos, H: n.H}
		n.Right = &Node{X: n.X + splitPos, Y: n.Y, W: n.W - splitPos, H: n.H}
	}

	g.split(n.Left, depth+1)
	g.split(n.Right, depth+1)
	return true
}

// createRooms carves rooms in leaf nodes.
func (g *Generator) createRooms(n *Node, tiles [][]int) {
	if n.Left == nil && n.Right == nil {
		// Leaf node: create a room
		maxW := min(n.W-2, g.MaxSize)
		maxH := min(n.H-2, g.MaxSize)

		if maxW < g.MinSize || maxH < g.MinSize {
			return // Node too small for a room
		}

		wRange := maxW - g.MinSize + 1
		hRange := maxH - g.MinSize + 1

		w := g.MinSize
		h := g.MinSize
		if wRange > 1 {
			w += g.rng.Intn(wRange)
		}
		if hRange > 1 {
			h += g.rng.Intn(hRange)
		}

		xRange := n.W - w - 1
		yRange := n.H - h - 1

		x := n.X + 1
		y := n.Y + 1
		if xRange > 1 {
			x += g.rng.Intn(xRange)
		}
		if yRange > 1 {
			y += g.rng.Intn(yRange)
		}

		n.Room = &Room{X: x, Y: y, W: w, H: h}

		for dy := 0; dy < h; dy++ {
			for dx := 0; dx < w; dx++ {
				if y+dy >= 0 && y+dy < g.Height && x+dx >= 0 && x+dx < g.Width {
					tiles[y+dy][x+dx] = g.floorTile
				}
			}
		}
		return
	}

	if n.Left != nil {
		g.createRooms(n.Left, tiles)
	}
	if n.Right != nil {
		g.createRooms(n.Right, tiles)
	}
}

// createCorridors connects sibling rooms with L-shaped or straight corridors.
func (g *Generator) createCorridors(n *Node, tiles [][]int) {
	if n.Left == nil || n.Right == nil {
		return
	}

	r1 := g.getRandomRoom(n.Left)
	r2 := g.getRandomRoom(n.Right)

	if r1 == nil || r2 == nil {
		return
	}

	// Center points of rooms
	x1 := r1.X + r1.W/2
	y1 := r1.Y + r1.H/2
	x2 := r2.X + r2.W/2
	y2 := r2.Y + r2.H/2

	// Carve L-shaped corridor
	if g.rng.Intn(2) == 0 {
		g.carveCorridor(x1, y1, x2, y1, tiles)
		g.carveCorridor(x2, y1, x2, y2, tiles)
	} else {
		g.carveCorridor(x1, y1, x1, y2, tiles)
		g.carveCorridor(x1, y2, x2, y2, tiles)
	}

	g.createCorridors(n.Left, tiles)
	g.createCorridors(n.Right, tiles)
}

// getRandomRoom returns a random leaf room from a subtree.
func (g *Generator) getRandomRoom(n *Node) *Room {
	if n == nil {
		return nil
	}
	if n.Room != nil {
		return n.Room
	}

	rooms := g.collectRooms(n)
	if len(rooms) == 0 {
		return nil
	}
	return rooms[g.rng.Intn(len(rooms))]
}

// collectRooms gathers all rooms in a subtree.
func (g *Generator) collectRooms(n *Node) []*Room {
	if n == nil {
		return nil
	}
	if n.Room != nil {
		return []*Room{n.Room}
	}

	var rooms []*Room
	rooms = append(rooms, g.collectRooms(n.Left)...)
	rooms = append(rooms, g.collectRooms(n.Right)...)
	return rooms
}

// carveCorridor carves a straight corridor between two points.
func (g *Generator) carveCorridor(x1, y1, x2, y2 int, tiles [][]int) {
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		if x >= 0 && x < g.Width && y1 >= 0 && y1 < g.Height {
			tiles[y1][x] = g.floorTile
		}
	}
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		if x2 >= 0 && x2 < g.Width && y >= 0 && y < g.Height {
			tiles[y][x2] = g.floorTile
		}
	}
}

// placeDoors inserts doors at chokepoints.
func (g *Generator) placeDoors(n *Node, tiles [][]int) {
	if n == nil {
		return
	}

	// Place doors at corridor-room junctions
	for y := 1; y < g.Height-1; y++ {
		for x := 1; x < g.Width-1; x++ {
			if tiles[y][x] == g.floorTile {
				// Check if this is a chokepoint (floor surrounded by walls on opposite sides)
				if (tiles[y-1][x] == g.wallTile && tiles[y+1][x] == g.wallTile &&
					tiles[y][x-1] == g.floorTile && tiles[y][x+1] == g.floorTile) ||
					(tiles[y][x-1] == g.wallTile && tiles[y][x+1] == g.wallTile &&
						tiles[y-1][x] == g.floorTile && tiles[y+1][x] == g.floorTile) {
					if g.rng.Intn(100) < 30 { // 30% chance
						tiles[y][x] = TileDoor
					}
				}
			}
		}
	}
}

// placeSecrets inserts secret walls in dead ends.
func (g *Generator) placeSecrets(n *Node, tiles [][]int) {
	if n == nil {
		return
	}

	// Find dead ends (floor tiles with 3 wall neighbors)
	for y := 1; y < g.Height-1; y++ {
		for x := 1; x < g.Width-1; x++ {
			if tiles[y][x] == g.floorTile {
				wallCount := 0
				if tiles[y-1][x] == g.wallTile {
					wallCount++
				}
				if tiles[y+1][x] == g.wallTile {
					wallCount++
				}
				if tiles[y][x-1] == g.wallTile {
					wallCount++
				}
				if tiles[y][x+1] == g.wallTile {
					wallCount++
				}

				if wallCount == 3 && g.rng.Intn(100) < 15 { // 15% chance
					// Place secret on one of the walls
					if tiles[y-1][x] == g.wallTile && g.rng.Intn(2) == 0 {
						tiles[y-1][x] = TileSecret
					} else if tiles[y+1][x] == g.wallTile && g.rng.Intn(2) == 0 {
						tiles[y+1][x] = TileSecret
					} else if tiles[y][x-1] == g.wallTile && g.rng.Intn(2) == 0 {
						tiles[y][x-1] = TileSecret
					} else if tiles[y][x+1] == g.wallTile {
						tiles[y][x+1] = TileSecret
					}
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

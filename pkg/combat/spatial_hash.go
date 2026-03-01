// Package combat handles damage calculation and combat events.
package combat

import (
	"math"
)

// Entity represents a spatial entity with position and bounds.
type Entity struct {
	ID     uint64
	X, Y   float64
	Radius float64
}

// SpatialHash provides O(1) broadphase collision detection for projectiles.
// Grid cells map to buckets containing entities in that cell.
type SpatialHash struct {
	cellSize float64
	grid     map[int64]map[int64][]Entity
}

// NewSpatialHash creates a spatial hash with the specified cell size.
// cellSize should typically match average entity radius × 2.
func NewSpatialHash(cellSize float64) *SpatialHash {
	return &SpatialHash{
		cellSize: cellSize,
		grid:     make(map[int64]map[int64][]Entity),
	}
}

// Clear removes all entities from the spatial hash.
func (sh *SpatialHash) Clear() {
	sh.grid = make(map[int64]map[int64][]Entity)
}

// Insert adds an entity to the spatial hash.
// Entities spanning multiple cells are added to all relevant cells.
func (sh *SpatialHash) Insert(e Entity) {
	minCellX := sh.cellCoord(e.X - e.Radius)
	maxCellX := sh.cellCoord(e.X + e.Radius)
	minCellY := sh.cellCoord(e.Y - e.Radius)
	maxCellY := sh.cellCoord(e.Y + e.Radius)

	for cx := minCellX; cx <= maxCellX; cx++ {
		for cy := minCellY; cy <= maxCellY; cy++ {
			if sh.grid[cx] == nil {
				sh.grid[cx] = make(map[int64][]Entity)
			}
			sh.grid[cx][cy] = append(sh.grid[cx][cy], e)
		}
	}
}

// Query returns all entities within radius of the query point.
// Returns a slice of entity IDs that may collide (broadphase).
func (sh *SpatialHash) Query(x, y, radius float64) []uint64 {
	minCellX := sh.cellCoord(x - radius)
	maxCellX := sh.cellCoord(x + radius)
	minCellY := sh.cellCoord(y - radius)
	maxCellY := sh.cellCoord(y + radius)

	seen := make(map[uint64]bool)
	var results []uint64

	for cx := minCellX; cx <= maxCellX; cx++ {
		for cy := minCellY; cy <= maxCellY; cy++ {
			if sh.grid[cx] != nil {
				for _, e := range sh.grid[cx][cy] {
					if !seen[e.ID] {
						dx := e.X - x
						dy := e.Y - y
						dist := math.Sqrt(dx*dx + dy*dy)
						if dist <= radius+e.Radius {
							results = append(results, e.ID)
							seen[e.ID] = true
						}
					}
				}
			}
		}
	}

	return results
}

// QueryEntities returns full entity data for entities within radius.
// Useful when position/radius information is needed.
func (sh *SpatialHash) QueryEntities(x, y, radius float64) []Entity {
	minCellX := sh.cellCoord(x - radius)
	maxCellX := sh.cellCoord(x + radius)
	minCellY := sh.cellCoord(y - radius)
	maxCellY := sh.cellCoord(y + radius)

	seen := make(map[uint64]bool)
	var results []Entity

	for cx := minCellX; cx <= maxCellX; cx++ {
		for cy := minCellY; cy <= maxCellY; cy++ {
			if sh.grid[cx] != nil {
				for _, e := range sh.grid[cx][cy] {
					if !seen[e.ID] {
						dx := e.X - x
						dy := e.Y - y
						dist := math.Sqrt(dx*dx + dy*dy)
						if dist <= radius+e.Radius {
							results = append(results, e)
							seen[e.ID] = true
						}
					}
				}
			}
		}
	}

	return results
}

// cellCoord converts world coordinate to grid cell coordinate.
func (sh *SpatialHash) cellCoord(worldCoord float64) int64 {
	return int64(math.Floor(worldCoord / sh.cellSize))
}

// CellCount returns the number of occupied cells in the hash.
func (sh *SpatialHash) CellCount() int {
	count := 0
	for _, row := range sh.grid {
		count += len(row)
	}
	return count
}

// EntityCount returns the total number of entity insertions (may include duplicates across cells).
func (sh *SpatialHash) EntityCount() int {
	count := 0
	for _, row := range sh.grid {
		for _, cell := range row {
			count += len(cell)
		}
	}
	return count
}

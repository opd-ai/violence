// Package spatial provides grid-based spatial indexing for fast proximity queries in the ECS.
package spatial

import (
	"math"
	"sync"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// Grid provides O(1) entity lookup within spatial regions.
// Replaces linear iteration for proximity queries.
type Grid struct {
	cellSize  float64
	cells     map[int64]map[int64][]engine.Entity
	entityPos map[engine.Entity]cellCoord
	mu        sync.RWMutex
	logger    *logrus.Entry
}

type cellCoord struct {
	x, y int64
}

// NewGrid creates a spatial grid with the specified cell size.
// cellSize should be tuned to typical query radius (2-4x recommended).
func NewGrid(cellSize float64) *Grid {
	return &Grid{
		cellSize:  cellSize,
		cells:     make(map[int64]map[int64][]engine.Entity),
		entityPos: make(map[engine.Entity]cellCoord),
		logger: logrus.WithFields(logrus.Fields{
			"system_name": "spatial",
		}),
	}
}

// Insert adds an entity at the given position.
func (g *Grid) Insert(e engine.Entity, x, y float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	cx := g.cellCoord(x)
	cy := g.cellCoord(y)

	if g.cells[cx] == nil {
		g.cells[cx] = make(map[int64][]engine.Entity)
	}
	g.cells[cx][cy] = append(g.cells[cx][cy], e)
	g.entityPos[e] = cellCoord{cx, cy}
}

// Update moves an entity to a new position.
// If the entity hasn't moved to a new cell, this is a no-op.
func (g *Grid) Update(e engine.Entity, x, y float64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	newCX := g.cellCoord(x)
	newCY := g.cellCoord(y)

	oldPos, exists := g.entityPos[e]
	if !exists {
		// Entity not in grid, insert it
		if g.cells[newCX] == nil {
			g.cells[newCX] = make(map[int64][]engine.Entity)
		}
		g.cells[newCX][newCY] = append(g.cells[newCX][newCY], e)
		g.entityPos[e] = cellCoord{newCX, newCY}
		return
	}

	// Same cell, no update needed
	if oldPos.x == newCX && oldPos.y == newCY {
		return
	}

	// Remove from old cell
	g.removeFromCell(e, oldPos.x, oldPos.y)

	// Add to new cell
	if g.cells[newCX] == nil {
		g.cells[newCX] = make(map[int64][]engine.Entity)
	}
	g.cells[newCX][newCY] = append(g.cells[newCX][newCY], e)
	g.entityPos[e] = cellCoord{newCX, newCY}
}

// Remove removes an entity from the grid.
func (g *Grid) Remove(e engine.Entity) {
	g.mu.Lock()
	defer g.mu.Unlock()

	pos, exists := g.entityPos[e]
	if !exists {
		return
	}

	g.removeFromCell(e, pos.x, pos.y)
	delete(g.entityPos, e)
}

// removeFromCell removes entity from the specified cell (caller must hold lock).
func (g *Grid) removeFromCell(e engine.Entity, cx, cy int64) {
	if g.cells[cx] == nil {
		return
	}

	cell := g.cells[cx][cy]
	for i, entity := range cell {
		if entity == e {
			// Remove by swapping with last element
			cell[i] = cell[len(cell)-1]
			g.cells[cx][cy] = cell[:len(cell)-1]
			break
		}
	}

	// Clean up empty cells
	if len(g.cells[cx][cy]) == 0 {
		delete(g.cells[cx], cy)
		if len(g.cells[cx]) == 0 {
			delete(g.cells, cx)
		}
	}
}

// QueryRadius returns all entities within the given radius of (x, y).
// This is the primary fast-path for proximity queries.
func (g *Grid) QueryRadius(x, y, radius float64) []engine.Entity {
	g.mu.RLock()
	defer g.mu.RUnlock()

	minCX := g.cellCoord(x - radius)
	maxCX := g.cellCoord(x + radius)
	minCY := g.cellCoord(y - radius)
	maxCY := g.cellCoord(y + radius)

	seen := make(map[engine.Entity]bool)
	var results []engine.Entity

	for cx := minCX; cx <= maxCX; cx++ {
		if g.cells[cx] == nil {
			continue
		}
		for cy := minCY; cy <= maxCY; cy++ {
			for _, e := range g.cells[cx][cy] {
				if seen[e] {
					continue
				}
				seen[e] = true
				results = append(results, e)
			}
		}
	}

	return results
}

// QueryRadiusFiltered returns entities within radius, filtered by distance check.
// Use this when you need exact circular proximity (QueryRadius returns cell-bounded results).
func (g *Grid) QueryRadiusFiltered(x, y, radius float64, positions map[engine.Entity]*engine.Position) []engine.Entity {
	g.mu.RLock()
	defer g.mu.RUnlock()

	minCX := g.cellCoord(x - radius)
	maxCX := g.cellCoord(x + radius)
	minCY := g.cellCoord(y - radius)
	maxCY := g.cellCoord(y + radius)

	seen := make(map[engine.Entity]bool)
	var results []engine.Entity

	radiusSq := radius * radius

	for cx := minCX; cx <= maxCX; cx++ {
		if g.cells[cx] == nil {
			continue
		}
		for cy := minCY; cy <= maxCY; cy++ {
			for _, e := range g.cells[cx][cy] {
				if seen[e] {
					continue
				}
				seen[e] = true

				// Distance check
				pos, ok := positions[e]
				if !ok {
					continue
				}
				dx := pos.X - x
				dy := pos.Y - y
				distSq := dx*dx + dy*dy
				if distSq <= radiusSq {
					results = append(results, e)
				}
			}
		}
	}

	return results
}

// QueryBounds returns all entities within the axis-aligned bounding box.
func (g *Grid) QueryBounds(minX, minY, maxX, maxY float64) []engine.Entity {
	g.mu.RLock()
	defer g.mu.RUnlock()

	minCX := g.cellCoord(minX)
	maxCX := g.cellCoord(maxX)
	minCY := g.cellCoord(minY)
	maxCY := g.cellCoord(maxY)

	seen := make(map[engine.Entity]bool)
	var results []engine.Entity

	for cx := minCX; cx <= maxCX; cx++ {
		if g.cells[cx] == nil {
			continue
		}
		for cy := minCY; cy <= maxCY; cy++ {
			for _, e := range g.cells[cx][cy] {
				if !seen[e] {
					seen[e] = true
					results = append(results, e)
				}
			}
		}
	}

	return results
}

// Clear removes all entities from the grid.
func (g *Grid) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.cells = make(map[int64]map[int64][]engine.Entity)
	g.entityPos = make(map[engine.Entity]cellCoord)
}

// Count returns the total number of entities in the grid.
func (g *Grid) Count() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.entityPos)
}

// CellCount returns the number of occupied cells.
func (g *Grid) CellCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, row := range g.cells {
		count += len(row)
	}
	return count
}

// cellCoord converts world coordinate to grid cell coordinate.
func (g *Grid) cellCoord(worldCoord float64) int64 {
	return int64(math.Floor(worldCoord / g.cellSize))
}

// GetCellSize returns the grid's cell size.
func (g *Grid) GetCellSize() float64 {
	return g.cellSize
}

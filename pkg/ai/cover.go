// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"github.com/opd-ai/violence/pkg/level"
)

// CoverTile represents a position that provides cover from a threat.
type CoverTile struct {
	Position Coord   // Position of the cover tile
	Score    float64 // Cover quality score (higher is better, 0-1 range)
}

// FindCoverTiles analyzes the grid to find positions that provide cover from a threat.
// Cover positions are walkable tiles adjacent to walls that block line of sight to the threat.
// Returns a slice of CoverTile scored by quality (0-1, with 1 being best cover).
// Score considers: threat visibility (blocked LOS = higher score), distance from threat (closer = higher score).
func FindCoverTiles(grid level.TileMap, threatPos Coord) []CoverTile {
	// Validate threat position
	if !grid.InBounds(threatPos.X, threatPos.Y) {
		return []CoverTile{}
	}

	covers := []CoverTile{}
	visited := make(map[Coord]bool)

	// Search radius around threat (limit to reasonable area)
	searchRadius := 20
	minX := maxInt(0, threatPos.X-searchRadius)
	maxX := minInt(grid.Width-1, threatPos.X+searchRadius)
	minY := maxInt(0, threatPos.Y-searchRadius)
	maxY := minInt(grid.Height-1, threatPos.Y+searchRadius)

	// Scan grid for potential cover positions
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			pos := Coord{X: x, Y: y}

			// Skip if not walkable or already visited
			if !grid.IsWalkable(x, y) || visited[pos] {
				continue
			}
			visited[pos] = true

			// Check if position is adjacent to a wall
			if !isAdjacentToWall(grid, x, y) {
				continue
			}

			// Calculate cover score
			score := calculateCoverScore(grid, pos, threatPos)
			if score > 0 {
				covers = append(covers, CoverTile{Position: pos, Score: score})
			}
		}
	}

	return covers
}

// isAdjacentToWall checks if position (x, y) is adjacent to at least one wall tile.
// Uses 4-directional adjacency (no diagonals).
func isAdjacentToWall(grid level.TileMap, x, y int) bool {
	neighbors := []Coord{
		{X: x + 1, Y: y},
		{X: x - 1, Y: y},
		{X: x, Y: y + 1},
		{X: x, Y: y - 1},
	}

	for _, n := range neighbors {
		if grid.InBounds(n.X, n.Y) && grid.Get(n.X, n.Y) == level.TileWall {
			return true
		}
	}
	return false
}

// calculateCoverScore computes cover quality from position to threat.
// Score ranges from 0 (no cover) to 1 (perfect cover).
// Factors: line of sight blocking (0.5), distance from threat (0.5).
func calculateCoverScore(grid level.TileMap, pos, threatPos Coord) float64 {
	// Check if direct line of sight to threat is blocked
	losBlocked := !hasLineOfSight(grid, pos, threatPos)
	if !losBlocked {
		return 0 // No cover if threat can see this position
	}

	// Distance factor: closer positions are better (more aggressive positioning)
	dist := manhattanDistance(pos, threatPos)
	maxDist := 20.0
	distScore := 1.0 - (dist / maxDist)
	if distScore < 0 {
		distScore = 0
	}

	// Weighted score: LOS blocking is critical (0.7), distance preference (0.3)
	score := 0.7 + (0.3 * distScore)
	return score
}

// hasLineOfSight checks if there's a clear line of sight between two positions.
// Uses Bresenham's line algorithm to trace ray through grid.
// Returns true if LOS is clear (no walls blocking), false otherwise.
func hasLineOfSight(grid level.TileMap, from, to Coord) bool {
	// Bresenham's line algorithm
	x0, y0 := from.X, from.Y
	x1, y1 := to.X, to.Y

	dx := absInt(x1 - x0)
	dy := absInt(y1 - y0)

	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}

	err := dx - dy
	x, y := x0, y0

	for {
		// Check current position (skip start and end points)
		if (x != from.X || y != from.Y) && (x != to.X || y != to.Y) {
			if !grid.InBounds(x, y) {
				return false
			}
			if grid.Get(x, y) == level.TileWall {
				return false
			}
		}

		// Reached destination
		if x == x1 && y == y1 {
			break
		}

		// Bresenham step
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}

	return true
}

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

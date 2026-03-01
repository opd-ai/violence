// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"github.com/opd-ai/violence/pkg/level"
)

// Coord represents a 2D coordinate in tile space.
type Coord struct {
	X, Y int
}

// astarNode represents an A* pathfinding node.
type astarNode struct {
	coord  Coord
	g, h   float64
	parent *astarNode
}

func (n *astarNode) f() float64 {
	return n.g + n.h
}

// FindPathCoord uses A* algorithm to find a path from start to goal on a TileMap.
// Returns a slice of coordinates representing the path, or an empty slice if no path exists.
// Uses Manhattan distance heuristic and disables diagonal movement.
func FindPathCoord(grid level.TileMap, start, goal Coord) []Coord {
	// Validate inputs
	if !grid.InBounds(start.X, start.Y) || !grid.InBounds(goal.X, goal.Y) {
		return []Coord{}
	}
	if !grid.IsWalkable(start.X, start.Y) || !grid.IsWalkable(goal.X, goal.Y) {
		return []Coord{}
	}
	if start.X == goal.X && start.Y == goal.Y {
		return []Coord{start}
	}

	// A* initialization
	openSet := []*astarNode{{coord: start, g: 0, h: manhattanDistance(start, goal)}}
	closedSet := make(map[Coord]bool)
	maxIter := 1000

	for iter := 0; iter < maxIter && len(openSet) > 0; iter++ {
		// Find node with lowest f score
		current := openSet[0]
		currentIdx := 0
		for i, n := range openSet {
			if n.f() < current.f() {
				current = n
				currentIdx = i
			}
		}

		// Remove current from open set
		openSet = append(openSet[:currentIdx], openSet[currentIdx+1:]...)

		// Check if goal reached
		if current.coord.X == goal.X && current.coord.Y == goal.Y {
			return reconstructPathCoord(current)
		}

		closedSet[current.coord] = true

		// Explore neighbors (4-directional, no diagonals)
		neighbors := []Coord{
			{X: current.coord.X + 1, Y: current.coord.Y},
			{X: current.coord.X - 1, Y: current.coord.Y},
			{X: current.coord.X, Y: current.coord.Y + 1},
			{X: current.coord.X, Y: current.coord.Y - 1},
		}

		for _, neighborCoord := range neighbors {
			// Skip if out of bounds or not walkable
			if !grid.InBounds(neighborCoord.X, neighborCoord.Y) {
				continue
			}
			if !grid.IsWalkable(neighborCoord.X, neighborCoord.Y) {
				continue
			}
			if closedSet[neighborCoord] {
				continue
			}

			// Calculate scores
			g := current.g + 1
			h := manhattanDistance(neighborCoord, goal)
			neighbor := &astarNode{coord: neighborCoord, g: g, h: h, parent: current}

			// Check if already in open set with better score
			found := false
			for i, n := range openSet {
				if n.coord.X == neighborCoord.X && n.coord.Y == neighborCoord.Y {
					if g < n.g {
						openSet[i] = neighbor
					}
					found = true
					break
				}
			}
			if !found {
				openSet = append(openSet, neighbor)
			}
		}
	}

	// No path found
	return []Coord{}
}

// manhattanDistance computes Manhattan distance heuristic.
func manhattanDistance(a, b Coord) float64 {
	dx := float64(absInt(a.X - b.X))
	dy := float64(absInt(a.Y - b.Y))
	return dx + dy
}

// absInt returns absolute value of an integer.
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// reconstructPathCoord builds path from goal node back to start.
func reconstructPathCoord(n *astarNode) []Coord {
	path := []Coord{}
	for current := n; current != nil; current = current.parent {
		path = append([]Coord{current.coord}, path...)
	}
	return path
}

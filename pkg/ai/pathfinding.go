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
func FindPathCoord(grid level.TileMap, start, goal Coord) []Coord {
	if !validateCoordPathInput(grid, start, goal) {
		return []Coord{}
	}

	openSet := []*astarNode{{coord: start, g: 0, h: manhattanDistance(start, goal)}}
	closedSet := make(map[Coord]bool)

	path := findAStarPathCoord(openSet, closedSet, grid, goal)
	if path != nil {
		return path
	}

	return []Coord{}
}

// validateCoordPathInput validates pathfinding inputs for coordinate-based search.
func validateCoordPathInput(grid level.TileMap, start, goal Coord) bool {
	if !grid.InBounds(start.X, start.Y) || !grid.InBounds(goal.X, goal.Y) {
		return false
	}
	if !grid.IsWalkable(start.X, start.Y) || !grid.IsWalkable(goal.X, goal.Y) {
		return false
	}
	if start.X == goal.X && start.Y == goal.Y {
		return false
	}
	return true
}

// findAStarPathCoord performs A* pathfinding with coordinate nodes.
func findAStarPathCoord(openSet []*astarNode, closedSet map[Coord]bool, grid level.TileMap, goal Coord) []Coord {
	maxIter := 1000
	for iter := 0; iter < maxIter && len(openSet) > 0; iter++ {
		current, currentIdx := findLowestFNodeCoord(openSet)
		openSet = append(openSet[:currentIdx], openSet[currentIdx+1:]...)

		if current.coord.X == goal.X && current.coord.Y == goal.Y {
			return reconstructPathCoord(current)
		}

		closedSet[current.coord] = true
		openSet = expandCoordNode(current, openSet, closedSet, grid, goal)
	}
	return nil
}

// findLowestFNodeCoord finds the node with lowest f score.
func findLowestFNodeCoord(openSet []*astarNode) (*astarNode, int) {
	current := openSet[0]
	currentIdx := 0
	for i, n := range openSet {
		if n.f() < current.f() {
			current = n
			currentIdx = i
		}
	}
	return current, currentIdx
}

// expandCoordNode explores neighbors of the current node.
func expandCoordNode(current *astarNode, openSet []*astarNode, closedSet map[Coord]bool, grid level.TileMap, goal Coord) []*astarNode {
	neighbors := []Coord{
		{X: current.coord.X + 1, Y: current.coord.Y},
		{X: current.coord.X - 1, Y: current.coord.Y},
		{X: current.coord.X, Y: current.coord.Y + 1},
		{X: current.coord.X, Y: current.coord.Y - 1},
	}

	for _, neighborCoord := range neighbors {
		if isValidCoordNeighbor(neighborCoord, closedSet, grid) {
			openSet = addOrUpdateCoordNeighbor(neighborCoord, current, openSet, goal)
		}
	}
	return openSet
}

// isValidCoordNeighbor checks if a coordinate neighbor is valid for pathfinding.
func isValidCoordNeighbor(coord Coord, closedSet map[Coord]bool, grid level.TileMap) bool {
	if !grid.InBounds(coord.X, coord.Y) {
		return false
	}
	if !grid.IsWalkable(coord.X, coord.Y) {
		return false
	}
	if closedSet[coord] {
		return false
	}
	return true
}

// addOrUpdateCoordNeighbor adds or updates a neighbor in the open set.
func addOrUpdateCoordNeighbor(neighborCoord Coord, current *astarNode, openSet []*astarNode, goal Coord) []*astarNode {
	g := current.g + 1
	h := manhattanDistance(neighborCoord, goal)
	neighbor := &astarNode{coord: neighborCoord, g: g, h: h, parent: current}

	for i, n := range openSet {
		if n.coord.X == neighborCoord.X && n.coord.Y == neighborCoord.Y {
			if g < n.g {
				openSet[i] = neighbor
			}
			return openSet
		}
	}
	return append(openSet, neighbor)
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

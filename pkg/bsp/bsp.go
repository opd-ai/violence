// Package bsp provides BSP-based procedural level generation.
package bsp

// Node represents a BSP tree node used during level generation.
type Node struct {
	X, Y, W, H int
	Left, Right *Node
}

// Generator produces levels using binary space partitioning.
type Generator struct {
	Width  int
	Height int
}

// NewGenerator creates a BSP generator for the given dimensions.
func NewGenerator(width, height int) *Generator {
	return &Generator{Width: width, Height: height}
}

// Generate produces a BSP tree and returns the root node.
func (g *Generator) Generate() *Node {
	return &Node{X: 0, Y: 0, W: g.Width, H: g.Height}
}

// SetGenre configures level generation parameters for a genre.
func SetGenre(genreID string) {}

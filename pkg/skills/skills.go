// Package skills provides skill tree management.
package skills

// Node represents a single skill tree node.
type Node struct {
	ID       string
	Name     string
	Children []string
}

// Tree holds a skill tree structure.
type Tree struct {
	Nodes     map[string]Node
	Allocated map[string]bool
}

// NewTree creates an empty skill tree.
func NewTree() *Tree {
	return &Tree{
		Nodes:     make(map[string]Node),
		Allocated: make(map[string]bool),
	}
}

// Allocate unlocks a skill node by ID.
func (t *Tree) Allocate(nodeID string) bool {
	t.Allocated[nodeID] = true
	return true
}

// GetBonus returns the cumulative bonus from allocated skills.
func (t *Tree) GetBonus() float64 {
	return 0
}

// SetGenre configures the skill tree for a genre.
func SetGenre(genreID string) {}

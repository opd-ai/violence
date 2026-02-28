// Package skills provides skill tree management.
package skills

import (
	"fmt"
	"sync"
)

// NodeType defines the type of skill node.
type NodeType string

const (
	NodeTypeCombat   NodeType = "combat"
	NodeTypeSurvival NodeType = "survival"
	NodeTypeTech     NodeType = "tech"
)

// Node represents a single skill tree node.
type Node struct {
	ID          string
	Name        string
	Description string
	Type        NodeType
	Requires    []string // IDs of prerequisite nodes
	BonusType   string
	BonusValue  float64
	Cost        int
}

// Tree holds a skill tree structure.
type Tree struct {
	Nodes     map[string]*Node
	Allocated map[string]bool
	Points    int
	mu        sync.RWMutex
}

// NewTree creates an empty skill tree.
func NewTree() *Tree {
	return &Tree{
		Nodes:     make(map[string]*Node),
		Allocated: make(map[string]bool),
		Points:    0,
	}
}

// AddNode adds a node to the skill tree.
func (t *Tree) AddNode(node *Node) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Nodes[node.ID] = node
}

// AddPoints adds skill points to spend.
func (t *Tree) AddPoints(amount int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Points += amount
}

// GetPoints returns available skill points.
func (t *Tree) GetPoints() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Points
}

// Allocate unlocks a skill node by ID.
// Returns true if successful, false if prerequisites not met or insufficient points.
func (t *Tree) Allocate(nodeID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check if already allocated
	if t.Allocated[nodeID] {
		return false
	}

	// Check if node exists
	node, ok := t.Nodes[nodeID]
	if !ok {
		return false
	}

	// Check prerequisites
	for _, reqID := range node.Requires {
		if !t.Allocated[reqID] {
			return false
		}
	}

	// Check points
	if t.Points < node.Cost {
		return false
	}

	// Allocate
	t.Allocated[nodeID] = true
	t.Points -= node.Cost
	return true
}

// IsAllocated checks if a node is allocated.
func (t *Tree) IsAllocated(nodeID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.Allocated[nodeID]
}

// GetNode retrieves a node by ID.
func (t *Tree) GetNode(nodeID string) (*Node, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node, ok := t.Nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("node not found: %s", nodeID)
	}
	return node, nil
}

// GetAllocatedNodes returns all allocated nodes.
func (t *Tree) GetAllocatedNodes() []*Node {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make([]*Node, 0)
	for id := range t.Allocated {
		if node, ok := t.Nodes[id]; ok {
			result = append(result, node)
		}
	}
	return result
}

// GetBonus returns the cumulative bonus of a specific type from allocated skills.
func (t *Tree) GetBonus(bonusType string) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var total float64
	for id := range t.Allocated {
		if node, ok := t.Nodes[id]; ok {
			if node.BonusType == bonusType {
				total += node.BonusValue
			}
		}
	}
	return total
}

// GetAllBonuses returns all bonus types and their cumulative values.
func (t *Tree) GetAllBonuses() map[string]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	bonuses := make(map[string]float64)
	for id := range t.Allocated {
		if node, ok := t.Nodes[id]; ok {
			bonuses[node.BonusType] += node.BonusValue
		}
	}
	return bonuses
}

// Reset deallocates all nodes and refunds points.
func (t *Tree) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Refund points
	for id := range t.Allocated {
		if node, ok := t.Nodes[id]; ok {
			t.Points += node.Cost
		}
	}

	// Clear allocations
	t.Allocated = make(map[string]bool)
}

// SetGenre configures the skill tree for a genre.
func SetGenre(genreID string) {}

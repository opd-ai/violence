// Package skills provides skill tree management with three trees: Combat, Survival, and Tech.
// Each tree has 5 nodes with prerequisites forming a directed graph.
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

// Manager manages all three skill trees: Combat, Survival, and Tech.
type Manager struct {
	trees map[string]*Tree
	mu    sync.RWMutex
}

// NewManager creates a manager with all three skill trees pre-configured.
func NewManager() *Manager {
	m := &Manager{
		trees: make(map[string]*Tree),
	}

	// Initialize the three trees
	m.trees["combat"] = NewTree()
	m.trees["survival"] = NewTree()
	m.trees["tech"] = NewTree()

	// Configure Combat tree (damage, reload, accuracy)
	m.initializeCombatTree()

	// Configure Survival tree (health, armor, stamina)
	m.initializeSurvivalTree()

	// Configure Tech tree (hacking, stealth, detection)
	m.initializeTechTree()

	return m
}

// initializeCombatTree sets up the Combat skill tree with 5 nodes.
func (m *Manager) initializeCombatTree() {
	tree := m.trees["combat"]

	tree.AddNode(&Node{
		ID:          "combat_dmg_1",
		Name:        "Weapon Training",
		Description: "Increases weapon damage by 10%",
		Type:        NodeTypeCombat,
		Requires:    []string{},
		BonusType:   "damage",
		BonusValue:  0.10,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "combat_dmg_2",
		Name:        "Deadly Force",
		Description: "Increases weapon damage by 15%",
		Type:        NodeTypeCombat,
		Requires:    []string{"combat_dmg_1"},
		BonusType:   "damage",
		BonusValue:  0.15,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "combat_reload_1",
		Name:        "Quick Hands",
		Description: "Decreases reload time by 15%",
		Type:        NodeTypeCombat,
		Requires:    []string{"combat_dmg_1"},
		BonusType:   "reload_speed",
		BonusValue:  0.15,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "combat_accuracy_1",
		Name:        "Steady Aim",
		Description: "Increases weapon accuracy by 10%",
		Type:        NodeTypeCombat,
		Requires:    []string{"combat_reload_1"},
		BonusType:   "accuracy",
		BonusValue:  0.10,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "combat_master",
		Name:        "Combat Mastery",
		Description: "Increases all combat stats by 5%",
		Type:        NodeTypeCombat,
		Requires:    []string{"combat_dmg_2", "combat_accuracy_1"},
		BonusType:   "combat_all",
		BonusValue:  0.05,
		Cost:        2,
	})
}

// initializeSurvivalTree sets up the Survival skill tree with 5 nodes.
func (m *Manager) initializeSurvivalTree() {
	tree := m.trees["survival"]

	tree.AddNode(&Node{
		ID:          "survival_health_1",
		Name:        "Vitality",
		Description: "Increases max health by 20%",
		Type:        NodeTypeSurvival,
		Requires:    []string{},
		BonusType:   "max_health",
		BonusValue:  0.20,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "survival_armor_1",
		Name:        "Reinforced Plating",
		Description: "Increases armor by 15%",
		Type:        NodeTypeSurvival,
		Requires:    []string{"survival_health_1"},
		BonusType:   "armor",
		BonusValue:  0.15,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "survival_stamina_1",
		Name:        "Endurance",
		Description: "Increases stamina by 25%",
		Type:        NodeTypeSurvival,
		Requires:    []string{"survival_health_1"},
		BonusType:   "stamina",
		BonusValue:  0.25,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "survival_regen",
		Name:        "Regeneration",
		Description: "Increases health regeneration by 50%",
		Type:        NodeTypeSurvival,
		Requires:    []string{"survival_armor_1", "survival_stamina_1"},
		BonusType:   "health_regen",
		BonusValue:  0.50,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "survival_master",
		Name:        "Survival Mastery",
		Description: "Increases all survival stats by 10%",
		Type:        NodeTypeSurvival,
		Requires:    []string{"survival_regen"},
		BonusType:   "survival_all",
		BonusValue:  0.10,
		Cost:        2,
	})
}

// initializeTechTree sets up the Tech skill tree with 5 nodes.
func (m *Manager) initializeTechTree() {
	tree := m.trees["tech"]

	tree.AddNode(&Node{
		ID:          "tech_hack_1",
		Name:        "Basic Hacking",
		Description: "Reduces hacking difficulty by 20%",
		Type:        NodeTypeTech,
		Requires:    []string{},
		BonusType:   "hacking",
		BonusValue:  0.20,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "tech_stealth_1",
		Name:        "Silent Movement",
		Description: "Increases stealth by 25%",
		Type:        NodeTypeTech,
		Requires:    []string{"tech_hack_1"},
		BonusType:   "stealth",
		BonusValue:  0.25,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "tech_detect_1",
		Name:        "Enhanced Sensors",
		Description: "Increases detection range by 30%",
		Type:        NodeTypeTech,
		Requires:    []string{"tech_hack_1"},
		BonusType:   "detection",
		BonusValue:  0.30,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "tech_advanced",
		Name:        "Advanced Tech",
		Description: "Improves all tech interactions by 15%",
		Type:        NodeTypeTech,
		Requires:    []string{"tech_stealth_1", "tech_detect_1"},
		BonusType:   "tech_all",
		BonusValue:  0.15,
		Cost:        1,
	})

	tree.AddNode(&Node{
		ID:          "tech_master",
		Name:        "Tech Mastery",
		Description: "Unlocks master hacker abilities",
		Type:        NodeTypeTech,
		Requires:    []string{"tech_advanced"},
		BonusType:   "tech_master",
		BonusValue:  0.50,
		Cost:        2,
	})
}

// AllocatePoint allocates a skill point in the specified tree and node.
// Returns an error if the tree doesn't exist, prerequisites aren't met, or insufficient points.
func (m *Manager) AllocatePoint(treeID, nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tree, ok := m.trees[treeID]
	if !ok {
		return fmt.Errorf("tree not found: %s", treeID)
	}

	success := tree.Allocate(nodeID)
	if !success {
		return fmt.Errorf("failed to allocate node %s: check prerequisites and points", nodeID)
	}

	return nil
}

// GetTree returns a skill tree by ID.
func (m *Manager) GetTree(treeID string) (*Tree, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tree, ok := m.trees[treeID]
	if !ok {
		return nil, fmt.Errorf("tree not found: %s", treeID)
	}
	return tree, nil
}

// AddPoints adds skill points to all trees.
func (m *Manager) AddPoints(amount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tree := range m.trees {
		tree.AddPoints(amount)
	}
}

// GetModifier returns the cumulative modifier for a specific stat across all trees.
// Modifiers are additive (e.g., 0.10 + 0.15 = 0.25 = 25% bonus).
func (m *Manager) GetModifier(stat string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var total float64

	// Check all trees for matching bonuses
	for _, tree := range m.trees {
		total += tree.GetBonus(stat)

		// Handle "all" bonuses (combat_all, survival_all, tech_all)
		if stat == "damage" || stat == "reload_speed" || stat == "accuracy" {
			total += tree.GetBonus("combat_all")
		} else if stat == "max_health" || stat == "armor" || stat == "stamina" || stat == "health_regen" {
			total += tree.GetBonus("survival_all")
		} else if stat == "hacking" || stat == "stealth" || stat == "detection" {
			total += tree.GetBonus("tech_all")
		}
	}

	return total
}

// GetAllModifiers returns all active stat modifiers.
func (m *Manager) GetAllModifiers() map[string]float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	modifiers := make(map[string]float64)

	for _, tree := range m.trees {
		bonuses := tree.GetAllBonuses()
		for bonusType, value := range bonuses {
			modifiers[bonusType] += value
		}
	}

	return modifiers
}

// IsNodeAllocated checks if a node is allocated in any tree.
func (m *Manager) IsNodeAllocated(treeID, nodeID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tree, ok := m.trees[treeID]
	if !ok {
		return false
	}

	return tree.IsAllocated(nodeID)
}

// Reset resets all skill trees, refunding all points.
func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tree := range m.trees {
		tree.Reset()
	}
}

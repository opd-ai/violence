package skills

import (
	"testing"
)

func TestNewTree(t *testing.T) {
	tree := NewTree()
	if tree == nil {
		t.Fatal("NewTree returned nil")
	}
	if tree.Nodes == nil {
		t.Fatal("Nodes map not initialized")
	}
	if tree.Allocated == nil {
		t.Fatal("Allocated map not initialized")
	}
	if tree.Points != 0 {
		t.Fatal("Points should start at 0")
	}
}

func TestTree_AddNode(t *testing.T) {
	tree := NewTree()
	node := &Node{
		ID:         "node1",
		Name:       "Test Node",
		Type:       NodeTypeCombat,
		BonusType:  "damage",
		BonusValue: 10,
		Cost:       1,
	}
	tree.AddNode(node)

	retrieved, err := tree.GetNode("node1")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if retrieved.Name != "Test Node" {
		t.Fatalf("wrong node: got name %q", retrieved.Name)
	}
}

func TestTree_AddPoints(t *testing.T) {
	tree := NewTree()
	tree.AddPoints(5)

	if tree.GetPoints() != 5 {
		t.Fatalf("wrong points: got %d", tree.GetPoints())
	}

	tree.AddPoints(3)
	if tree.GetPoints() != 8 {
		t.Fatalf("wrong points after second add: got %d", tree.GetPoints())
	}
}

func TestTree_Allocate(t *testing.T) {
	tree := NewTree()
	node := &Node{ID: "node1", Cost: 1, BonusType: "damage", BonusValue: 10}
	tree.AddNode(node)
	tree.AddPoints(2)

	success := tree.Allocate("node1")
	if !success {
		t.Fatal("allocation failed")
	}
	if !tree.IsAllocated("node1") {
		t.Fatal("node not marked allocated")
	}
	if tree.GetPoints() != 1 {
		t.Fatalf("wrong points after allocation: got %d", tree.GetPoints())
	}
}

func TestTree_AllocateInsufficientPoints(t *testing.T) {
	tree := NewTree()
	node := &Node{ID: "node1", Cost: 5, BonusType: "damage", BonusValue: 10}
	tree.AddNode(node)
	tree.AddPoints(2)

	success := tree.Allocate("node1")
	if success {
		t.Fatal("should fail with insufficient points")
	}
	if tree.IsAllocated("node1") {
		t.Fatal("node should not be allocated")
	}
}

func TestTree_AllocatePrerequisites(t *testing.T) {
	tree := NewTree()
	node1 := &Node{ID: "node1", Cost: 1, BonusType: "damage", BonusValue: 5}
	node2 := &Node{ID: "node2", Cost: 1, Requires: []string{"node1"}, BonusType: "damage", BonusValue: 10}
	tree.AddNode(node1)
	tree.AddNode(node2)
	tree.AddPoints(5)

	// Try to allocate node2 without node1
	success := tree.Allocate("node2")
	if success {
		t.Fatal("should fail without prerequisites")
	}

	// Allocate node1 first
	tree.Allocate("node1")

	// Now allocate node2
	success = tree.Allocate("node2")
	if !success {
		t.Fatal("should succeed with prerequisites")
	}
}

func TestTree_AllocateTwice(t *testing.T) {
	tree := NewTree()
	node := &Node{ID: "node1", Cost: 1, BonusType: "damage", BonusValue: 10}
	tree.AddNode(node)
	tree.AddPoints(5)

	tree.Allocate("node1")
	success := tree.Allocate("node1")
	if success {
		t.Fatal("should not allocate twice")
	}
	if tree.GetPoints() != 4 {
		t.Fatalf("points should only be spent once: got %d", tree.GetPoints())
	}
}

func TestTree_AllocateNonexistent(t *testing.T) {
	tree := NewTree()
	tree.AddPoints(5)

	success := tree.Allocate("nonexistent")
	if success {
		t.Fatal("should fail for nonexistent node")
	}
}

func TestTree_GetNode(t *testing.T) {
	tree := NewTree()
	node := &Node{ID: "node1", Name: "Test"}
	tree.AddNode(node)

	retrieved, err := tree.GetNode("node1")
	if err != nil {
		t.Fatalf("GetNode failed: %v", err)
	}
	if retrieved.Name != "Test" {
		t.Fatalf("wrong node: got %q", retrieved.Name)
	}
}

func TestTree_GetNodeNotFound(t *testing.T) {
	tree := NewTree()
	_, err := tree.GetNode("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent node")
	}
}

func TestTree_GetAllocatedNodes(t *testing.T) {
	tree := NewTree()
	node1 := &Node{ID: "node1", Cost: 1, BonusType: "damage", BonusValue: 5}
	node2 := &Node{ID: "node2", Cost: 1, BonusType: "armor", BonusValue: 10}
	tree.AddNode(node1)
	tree.AddNode(node2)
	tree.AddPoints(5)

	tree.Allocate("node1")
	tree.Allocate("node2")

	allocated := tree.GetAllocatedNodes()
	if len(allocated) != 2 {
		t.Fatalf("expected 2 allocated nodes, got %d", len(allocated))
	}
}

func TestTree_GetBonus(t *testing.T) {
	tree := NewTree()
	node1 := &Node{ID: "node1", Cost: 1, BonusType: "damage", BonusValue: 5}
	node2 := &Node{ID: "node2", Cost: 1, BonusType: "damage", BonusValue: 10}
	node3 := &Node{ID: "node3", Cost: 1, BonusType: "armor", BonusValue: 20}
	tree.AddNode(node1)
	tree.AddNode(node2)
	tree.AddNode(node3)
	tree.AddPoints(5)

	tree.Allocate("node1")
	tree.Allocate("node2")
	tree.Allocate("node3")

	damageBonus := tree.GetBonus("damage")
	if damageBonus != 15 {
		t.Fatalf("wrong damage bonus: got %f", damageBonus)
	}

	armorBonus := tree.GetBonus("armor")
	if armorBonus != 20 {
		t.Fatalf("wrong armor bonus: got %f", armorBonus)
	}
}

func TestTree_GetAllBonuses(t *testing.T) {
	tree := NewTree()
	node1 := &Node{ID: "node1", Cost: 1, BonusType: "damage", BonusValue: 5}
	node2 := &Node{ID: "node2", Cost: 1, BonusType: "armor", BonusValue: 10}
	tree.AddNode(node1)
	tree.AddNode(node2)
	tree.AddPoints(5)

	tree.Allocate("node1")
	tree.Allocate("node2")

	bonuses := tree.GetAllBonuses()
	if len(bonuses) != 2 {
		t.Fatalf("expected 2 bonus types, got %d", len(bonuses))
	}
	if bonuses["damage"] != 5 {
		t.Fatalf("wrong damage bonus: got %f", bonuses["damage"])
	}
	if bonuses["armor"] != 10 {
		t.Fatalf("wrong armor bonus: got %f", bonuses["armor"])
	}
}

func TestTree_Reset(t *testing.T) {
	tree := NewTree()
	node1 := &Node{ID: "node1", Cost: 2, BonusType: "damage", BonusValue: 5}
	node2 := &Node{ID: "node2", Cost: 3, BonusType: "armor", BonusValue: 10}
	tree.AddNode(node1)
	tree.AddNode(node2)
	tree.AddPoints(10)

	tree.Allocate("node1")
	tree.Allocate("node2")

	// Points should be 10 - 2 - 3 = 5
	if tree.GetPoints() != 5 {
		t.Fatalf("wrong points before reset: got %d", tree.GetPoints())
	}

	tree.Reset()

	// Points should be refunded: 5 + 2 + 3 = 10
	if tree.GetPoints() != 10 {
		t.Fatalf("wrong points after reset: got %d", tree.GetPoints())
	}
	if tree.IsAllocated("node1") {
		t.Fatal("node1 should be deallocated")
	}
	if tree.IsAllocated("node2") {
		t.Fatal("node2 should be deallocated")
	}
}

func TestNodeTypes(t *testing.T) {
	// Test all node types exist
	types := []NodeType{NodeTypeCombat, NodeTypeSurvival, NodeTypeTech}
	for _, nt := range types {
		if string(nt) == "" {
			t.Fatalf("node type %v has empty string", nt)
		}
	}
}

func TestSetGenre(t *testing.T) {
	// Should not panic
	SetGenre("fantasy")
	SetGenre("scifi")
	SetGenre("")
}

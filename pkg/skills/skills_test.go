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

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.trees == nil {
		t.Fatal("trees map not initialized")
	}

	// Check all three trees exist
	trees := []string{"combat", "survival", "tech"}
	for _, treeID := range trees {
		tree, err := m.GetTree(treeID)
		if err != nil {
			t.Fatalf("tree %s not found: %v", treeID, err)
		}
		if tree == nil {
			t.Fatalf("tree %s is nil", treeID)
		}
	}
}

func TestManager_CombatTreeStructure(t *testing.T) {
	m := NewManager()
	tree, _ := m.GetTree("combat")

	// Verify all 5 combat nodes exist
	expectedNodes := []string{
		"combat_dmg_1",
		"combat_dmg_2",
		"combat_reload_1",
		"combat_accuracy_1",
		"combat_master",
	}

	for _, nodeID := range expectedNodes {
		node, err := tree.GetNode(nodeID)
		if err != nil {
			t.Fatalf("combat node %s not found: %v", nodeID, err)
		}
		if node.Type != NodeTypeCombat {
			t.Fatalf("node %s has wrong type: %s", nodeID, node.Type)
		}
	}

	// Verify prerequisites
	dmg2, _ := tree.GetNode("combat_dmg_2")
	if len(dmg2.Requires) != 1 || dmg2.Requires[0] != "combat_dmg_1" {
		t.Fatalf("combat_dmg_2 has wrong prerequisites: %v", dmg2.Requires)
	}

	master, _ := tree.GetNode("combat_master")
	if len(master.Requires) != 2 {
		t.Fatalf("combat_master should have 2 prerequisites, got %d", len(master.Requires))
	}
}

func TestManager_SurvivalTreeStructure(t *testing.T) {
	m := NewManager()
	tree, _ := m.GetTree("survival")

	// Verify all 5 survival nodes exist
	expectedNodes := []string{
		"survival_health_1",
		"survival_armor_1",
		"survival_stamina_1",
		"survival_regen",
		"survival_master",
	}

	for _, nodeID := range expectedNodes {
		node, err := tree.GetNode(nodeID)
		if err != nil {
			t.Fatalf("survival node %s not found: %v", nodeID, err)
		}
		if node.Type != NodeTypeSurvival {
			t.Fatalf("node %s has wrong type: %s", nodeID, node.Type)
		}
	}
}

func TestManager_TechTreeStructure(t *testing.T) {
	m := NewManager()
	tree, _ := m.GetTree("tech")

	// Verify all 5 tech nodes exist
	expectedNodes := []string{
		"tech_hack_1",
		"tech_stealth_1",
		"tech_detect_1",
		"tech_advanced",
		"tech_master",
	}

	for _, nodeID := range expectedNodes {
		node, err := tree.GetNode(nodeID)
		if err != nil {
			t.Fatalf("tech node %s not found: %v", nodeID, err)
		}
		if node.Type != NodeTypeTech {
			t.Fatalf("node %s has wrong type: %s", nodeID, node.Type)
		}
	}
}

func TestManager_AllocatePoint(t *testing.T) {
	tests := []struct {
		name      string
		treeID    string
		nodeID    string
		points    int
		wantError bool
	}{
		{
			name:      "valid allocation in combat tree",
			treeID:    "combat",
			nodeID:    "combat_dmg_1",
			points:    1,
			wantError: false,
		},
		{
			name:      "invalid tree",
			treeID:    "invalid",
			nodeID:    "combat_dmg_1",
			points:    1,
			wantError: true,
		},
		{
			name:      "insufficient points",
			treeID:    "combat",
			nodeID:    "combat_dmg_1",
			points:    0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager()
			m.AddPoints(tt.points)

			err := m.AllocatePoint(tt.treeID, tt.nodeID)
			if (err != nil) != tt.wantError {
				t.Fatalf("AllocatePoint() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestManager_AllocatePointWithPrerequisites(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	// Try to allocate combat_dmg_2 without combat_dmg_1
	err := m.AllocatePoint("combat", "combat_dmg_2")
	if err == nil {
		t.Fatal("should fail without prerequisites")
	}

	// Allocate combat_dmg_1 first
	err = m.AllocatePoint("combat", "combat_dmg_1")
	if err != nil {
		t.Fatalf("failed to allocate combat_dmg_1: %v", err)
	}

	// Now allocate combat_dmg_2
	err = m.AllocatePoint("combat", "combat_dmg_2")
	if err != nil {
		t.Fatalf("failed to allocate combat_dmg_2 with prerequisites: %v", err)
	}
}

func TestManager_GetModifier(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	// Allocate combat_dmg_1 (0.10 damage bonus)
	m.AllocatePoint("combat", "combat_dmg_1")

	modifier := m.GetModifier("damage")
	if modifier != 0.10 {
		t.Fatalf("wrong damage modifier: got %f, want 0.10", modifier)
	}

	// Allocate combat_dmg_2 (0.15 damage bonus)
	m.AllocatePoint("combat", "combat_dmg_2")

	modifier = m.GetModifier("damage")
	expected := 0.10 + 0.15
	if modifier != expected {
		t.Fatalf("wrong cumulative damage modifier: got %f, want %f", modifier, expected)
	}
}

func TestManager_GetModifierWithMasteryBonus(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	// Allocate full combat tree to get combat_master
	m.AllocatePoint("combat", "combat_dmg_1")
	m.AllocatePoint("combat", "combat_dmg_2")
	m.AllocatePoint("combat", "combat_reload_1")
	m.AllocatePoint("combat", "combat_accuracy_1")
	m.AllocatePoint("combat", "combat_master")

	// combat_master gives +5% to all combat stats
	damageModifier := m.GetModifier("damage")
	// 0.10 (dmg_1) + 0.15 (dmg_2) + 0.05 (master) = 0.30
	expected := 0.30
	if damageModifier != expected {
		t.Fatalf("damage modifier with mastery: got %f, want %f", damageModifier, expected)
	}

	reloadModifier := m.GetModifier("reload_speed")
	// 0.15 (reload_1) + 0.05 (master) = 0.20
	expected = 0.20
	if reloadModifier != expected {
		t.Fatalf("reload modifier with mastery: got %f, want %f", reloadModifier, expected)
	}
}

func TestManager_GetModifierAcrossTrees(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	// Allocate in both combat and survival trees
	m.AllocatePoint("combat", "combat_dmg_1")
	m.AllocatePoint("survival", "survival_health_1")

	damageModifier := m.GetModifier("damage")
	if damageModifier != 0.10 {
		t.Fatalf("damage modifier: got %f, want 0.10", damageModifier)
	}

	healthModifier := m.GetModifier("max_health")
	if healthModifier != 0.20 {
		t.Fatalf("health modifier: got %f, want 0.20", healthModifier)
	}
}

func TestManager_GetAllModifiers(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	m.AllocatePoint("combat", "combat_dmg_1")
	m.AllocatePoint("survival", "survival_health_1")
	m.AllocatePoint("tech", "tech_hack_1")

	modifiers := m.GetAllModifiers()

	if len(modifiers) != 3 {
		t.Fatalf("expected 3 modifiers, got %d", len(modifiers))
	}

	if modifiers["damage"] != 0.10 {
		t.Fatalf("wrong damage modifier: %f", modifiers["damage"])
	}
	if modifiers["max_health"] != 0.20 {
		t.Fatalf("wrong health modifier: %f", modifiers["max_health"])
	}
	if modifiers["hacking"] != 0.20 {
		t.Fatalf("wrong hacking modifier: %f", modifiers["hacking"])
	}
}

func TestManager_IsNodeAllocated(t *testing.T) {
	m := NewManager()
	m.AddPoints(5)

	if m.IsNodeAllocated("combat", "combat_dmg_1") {
		t.Fatal("node should not be allocated yet")
	}

	m.AllocatePoint("combat", "combat_dmg_1")

	if !m.IsNodeAllocated("combat", "combat_dmg_1") {
		t.Fatal("node should be allocated")
	}

	// Check invalid tree
	if m.IsNodeAllocated("invalid", "combat_dmg_1") {
		t.Fatal("invalid tree should return false")
	}
}

func TestManager_Reset(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	m.AllocatePoint("combat", "combat_dmg_1")
	m.AllocatePoint("survival", "survival_health_1")

	// Verify allocations
	if !m.IsNodeAllocated("combat", "combat_dmg_1") {
		t.Fatal("combat node should be allocated")
	}
	if !m.IsNodeAllocated("survival", "survival_health_1") {
		t.Fatal("survival node should be allocated")
	}

	// Reset
	m.Reset()

	// Verify all nodes are deallocated
	if m.IsNodeAllocated("combat", "combat_dmg_1") {
		t.Fatal("combat node should be deallocated")
	}
	if m.IsNodeAllocated("survival", "survival_health_1") {
		t.Fatal("survival node should be deallocated")
	}

	// Verify points were refunded
	combatTree, _ := m.GetTree("combat")
	if combatTree.GetPoints() != 10 {
		t.Fatalf("combat tree should have 10 points, got %d", combatTree.GetPoints())
	}
}

func TestManager_AddPoints(t *testing.T) {
	m := NewManager()

	m.AddPoints(5)

	// All trees should have the points
	for _, treeID := range []string{"combat", "survival", "tech"} {
		tree, _ := m.GetTree(treeID)
		if tree.GetPoints() != 5 {
			t.Fatalf("%s tree should have 5 points, got %d", treeID, tree.GetPoints())
		}
	}
}

func TestManager_GetTreeNotFound(t *testing.T) {
	m := NewManager()

	_, err := m.GetTree("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent tree")
	}
}

func TestManager_ComplexPrerequisitePath(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	// Try to allocate survival_master (requires survival_regen)
	err := m.AllocatePoint("survival", "survival_master")
	if err == nil {
		t.Fatal("should fail without full prerequisite chain")
	}

	// Build the prerequisite chain
	m.AllocatePoint("survival", "survival_health_1")
	m.AllocatePoint("survival", "survival_armor_1")
	m.AllocatePoint("survival", "survival_stamina_1")
	m.AllocatePoint("survival", "survival_regen")

	// Now should succeed
	err = m.AllocatePoint("survival", "survival_master")
	if err != nil {
		t.Fatalf("should succeed with full prerequisite chain: %v", err)
	}
}

func TestManager_TreeIndependence(t *testing.T) {
	m := NewManager()
	m.AddPoints(10)

	// Allocate in combat tree
	m.AllocatePoint("combat", "combat_dmg_1")

	combatTree, _ := m.GetTree("combat")
	survivalTree, _ := m.GetTree("survival")

	// Combat tree should have spent points
	if combatTree.GetPoints() != 9 {
		t.Fatalf("combat tree should have 9 points, got %d", combatTree.GetPoints())
	}

	// Survival tree should still have all points
	if survivalTree.GetPoints() != 10 {
		t.Fatalf("survival tree should have 10 points, got %d", survivalTree.GetPoints())
	}
}

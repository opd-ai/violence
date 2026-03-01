// Package ai implements enemy artificial intelligence behaviors.
package ai

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestEnemyRoleComponent_Type(t *testing.T) {
	c := &EnemyRoleComponent{}
	if c.Type() != "EnemyRoleComponent" {
		t.Errorf("Type() = %v, want EnemyRoleComponent", c.Type())
	}
}

func TestSquadTacticsComponent_Type(t *testing.T) {
	c := &SquadTacticsComponent{}
	if c.Type() != "SquadTacticsComponent" {
		t.Errorf("Type() = %v, want SquadTacticsComponent", c.Type())
	}
}

func TestPositionComponent_Type(t *testing.T) {
	c := &PositionComponent{}
	if c.Type() != "PositionComponent" {
		t.Errorf("Type() = %v, want PositionComponent", c.Type())
	}
}

func TestHealthComponent_Type(t *testing.T) {
	c := &HealthComponent{}
	if c.Type() != "HealthComponent" {
		t.Errorf("Type() = %v, want HealthComponent", c.Type())
	}
}

func TestTargetComponent_Type(t *testing.T) {
	c := &TargetComponent{}
	if c.Type() != "TargetComponent" {
		t.Errorf("Type() = %v, want TargetComponent", c.Type())
	}
}

func TestNewRoleBasedAISystem(t *testing.T) {
	sys := NewRoleBasedAISystem()
	if sys == nil {
		t.Fatal("NewRoleBasedAISystem returned nil")
	}
	if sys.squads == nil {
		t.Error("squads map should be initialized")
	}
	if sys.gameTime != 0.0 {
		t.Error("gameTime should start at 0")
	}
}

func TestRoleBasedAISystem_Update(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create a tank enemy with position and health
	e := w.AddEntity()
	w.AddComponent(e, &EnemyRoleComponent{
		Role:   RoleTank,
		Config: GetRoleConfig(RoleTank, "fantasy"),
	})
	w.AddComponent(e, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(e, &HealthComponent{Current: 100, Max: 100})

	// Should not panic
	sys.Update(w)

	if sys.gameTime <= 0 {
		t.Error("gameTime should increase after Update")
	}
}

func TestRoleBasedAISystem_UpdateWithSquad(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create squad members
	e1 := w.AddEntity()
	w.AddComponent(e1, &EnemyRoleComponent{
		Role:   RoleTank,
		Config: GetRoleConfig(RoleTank, "fantasy"),
	})
	w.AddComponent(e1, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(e1, &HealthComponent{Current: 100, Max: 100})
	w.AddComponent(e1, &SquadTacticsComponent{
		SquadID:  "alpha",
		IsLeader: true,
	})

	e2 := w.AddEntity()
	w.AddComponent(e2, &EnemyRoleComponent{
		Role:   RoleRanged,
		Config: GetRoleConfig(RoleRanged, "fantasy"),
	})
	w.AddComponent(e2, &PositionComponent{X: 150, Y: 100})
	w.AddComponent(e2, &HealthComponent{Current: 80, Max: 100})
	w.AddComponent(e2, &SquadTacticsComponent{
		SquadID:  "alpha",
		IsLeader: false,
	})

	// Run update multiple times to trigger squad update
	for i := 0; i < 20; i++ {
		sys.Update(w)
	}

	// Verify squad was created
	if _, exists := sys.squads["alpha"]; !exists {
		t.Error("Squad 'alpha' should exist after update")
	}

	squad := sys.squads["alpha"]
	if len(squad.Members) != 2 {
		t.Errorf("Squad should have 2 members, got %v", len(squad.Members))
	}
}

func TestRoleBasedAISystem_UpdateWithTarget(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create target entity
	target := w.AddEntity()
	w.AddComponent(target, &PositionComponent{X: 200, Y: 200})

	// Create enemy with target
	e := w.AddEntity()
	w.AddComponent(e, &EnemyRoleComponent{
		Role:   RoleRanged,
		Config: GetRoleConfig(RoleRanged, "fantasy"),
	})
	w.AddComponent(e, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(e, &HealthComponent{Current: 50, Max: 100})
	w.AddComponent(e, &TargetComponent{
		TargetID: target,
		LastSeen: 0.0,
	})

	// Should not panic
	sys.Update(w)
}

func TestRoleBasedAISystem_MultipleRoles(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	target := w.AddEntity()
	w.AddComponent(target, &PositionComponent{X: 500, Y: 500})

	roles := []EnemyRole{RoleTank, RoleRanged, RoleHealer, RoleAmbusher, RoleScout}

	for _, role := range roles {
		e := w.AddEntity()
		w.AddComponent(e, &EnemyRoleComponent{
			Role:   role,
			Config: GetRoleConfig(role, "fantasy"),
		})
		w.AddComponent(e, &PositionComponent{X: 100, Y: 100})
		w.AddComponent(e, &HealthComponent{Current: 100, Max: 100})
		w.AddComponent(e, &TargetComponent{
			TargetID: target,
			LastSeen: 0.0,
		})
	}

	// Should handle all roles without panic
	sys.Update(w)
}

func TestRoleBasedAISystem_HealerBehavior(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create wounded squad member
	wounded := w.AddEntity()
	w.AddComponent(wounded, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(wounded, &HealthComponent{Current: 30, Max: 100})
	w.AddComponent(wounded, &SquadTacticsComponent{SquadID: "alpha"})

	// Create healer
	healer := w.AddEntity()
	w.AddComponent(healer, &EnemyRoleComponent{
		Role:   RoleHealer,
		Config: GetRoleConfig(RoleHealer, "fantasy"),
	})
	w.AddComponent(healer, &PositionComponent{X: 120, Y: 100})
	w.AddComponent(healer, &HealthComponent{Current: 100, Max: 100})
	w.AddComponent(healer, &SquadTacticsComponent{SquadID: "alpha"})

	target := w.AddEntity()
	w.AddComponent(target, &PositionComponent{X: 300, Y: 300})
	w.AddComponent(healer, &TargetComponent{TargetID: target})

	// Update to trigger squad formation
	for i := 0; i < 20; i++ {
		sys.Update(w)
	}

	// Squad should exist
	if _, exists := sys.squads["alpha"]; !exists {
		t.Error("Squad should be created")
	}
}

func TestRoleBasedAISystem_ScoutAlerts(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create scout with squad
	scout := w.AddEntity()
	w.AddComponent(scout, &EnemyRoleComponent{
		Role:   RoleScout,
		Config: GetRoleConfig(RoleScout, "fantasy"),
	})
	w.AddComponent(scout, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(scout, &HealthComponent{Current: 100, Max: 100})
	w.AddComponent(scout, &SquadTacticsComponent{SquadID: "bravo"})

	// Create target within sight range
	target := w.AddEntity()
	w.AddComponent(target, &PositionComponent{X: 200, Y: 100}) // 100 units away
	w.AddComponent(scout, &TargetComponent{TargetID: target})

	// Update multiple times
	for i := 0; i < 20; i++ {
		sys.Update(w)
	}

	// Squad should exist and have raised alert
	if squad, exists := sys.squads["bravo"]; exists {
		if squad.AlertLevel <= 0 {
			t.Error("Scout should raise squad alert level")
		}
	}
}

func TestRoleBasedAISystem_AmbusherBehavior(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	target := w.AddEntity()
	// Position target at ambusher's preferred range (60 units)
	w.AddComponent(target, &PositionComponent{X: 160, Y: 100})

	ambusher := w.AddEntity()
	cfg := GetRoleConfig(RoleAmbusher, "fantasy")
	w.AddComponent(ambusher, &EnemyRoleComponent{
		Role:   RoleAmbusher,
		Config: cfg,
	})
	w.AddComponent(ambusher, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(ambusher, &HealthComponent{Current: 100, Max: 100})
	w.AddComponent(ambusher, &TargetComponent{TargetID: target})

	// Verify preferred range is achievable
	if cfg.PreferredRange < 50 || cfg.PreferredRange > 70 {
		t.Logf("Ambusher preferred range: %v", cfg.PreferredRange)
	}

	// Should not panic when running ambusher behavior
	for i := 0; i < 10; i++ {
		sys.Update(w)
	}
}

func TestRoleBasedAISystem_FindVisibleTargets(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create non-zero entity IDs for targets
	_ = w.AddEntity()        // 0
	target1 := w.AddEntity() // 1
	target2 := w.AddEntity() // 2

	e1 := w.AddEntity()
	w.AddComponent(e1, &TargetComponent{TargetID: target1})

	e2 := w.AddEntity()
	w.AddComponent(e2, &TargetComponent{TargetID: target2})

	e3 := w.AddEntity()
	w.AddComponent(e3, &TargetComponent{TargetID: target1}) // Duplicate

	members := []engine.Entity{e1, e2, e3}
	targets := sys.findVisibleTargets(w, members)

	// Should have 2 unique targets
	if len(targets) != 2 {
		t.Errorf("Expected 2 unique targets, got %v", len(targets))
	}
}

func TestRoleBasedAISystem_NoTargetNoBehavior(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create enemy without target
	e := w.AddEntity()
	w.AddComponent(e, &EnemyRoleComponent{
		Role:   RoleTank,
		Config: GetRoleConfig(RoleTank, "fantasy"),
	})
	w.AddComponent(e, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(e, &HealthComponent{Current: 100, Max: 100})

	// Should not panic
	sys.Update(w)
}

func TestRoleBasedAISystem_Integration(t *testing.T) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	// Create dummy entity to ensure player has non-zero ID
	_ = w.AddEntity()

	// Create player target
	player := w.AddEntity()
	w.AddComponent(player, &PositionComponent{X: 500, Y: 500})

	// Create diverse enemy squad
	squadID := "charlie"

	// Tank
	tank := w.AddEntity()
	w.AddComponent(tank, &EnemyRoleComponent{Role: RoleTank, Config: GetRoleConfig(RoleTank, "fantasy")})
	w.AddComponent(tank, &PositionComponent{X: 100, Y: 100})
	w.AddComponent(tank, &HealthComponent{Current: 150, Max: 150})
	w.AddComponent(tank, &SquadTacticsComponent{SquadID: squadID, IsLeader: true})
	w.AddComponent(tank, &TargetComponent{TargetID: player})

	// Ranged
	ranged := w.AddEntity()
	w.AddComponent(ranged, &EnemyRoleComponent{Role: RoleRanged, Config: GetRoleConfig(RoleRanged, "fantasy")})
	w.AddComponent(ranged, &PositionComponent{X: 120, Y: 120})
	w.AddComponent(ranged, &HealthComponent{Current: 80, Max: 80})
	w.AddComponent(ranged, &SquadTacticsComponent{SquadID: squadID})
	w.AddComponent(ranged, &TargetComponent{TargetID: player})

	// Healer
	healer := w.AddEntity()
	w.AddComponent(healer, &EnemyRoleComponent{Role: RoleHealer, Config: GetRoleConfig(RoleHealer, "fantasy")})
	w.AddComponent(healer, &PositionComponent{X: 80, Y: 100})
	w.AddComponent(healer, &HealthComponent{Current: 70, Max: 70})
	w.AddComponent(healer, &SquadTacticsComponent{SquadID: squadID})
	w.AddComponent(healer, &TargetComponent{TargetID: player})

	// Run simulation - need enough updates to trigger squad update timer (0.2s intervals)
	// At ~0.016s per update, need at least 13 updates, run more to be safe
	for i := 0; i < 50; i++ {
		sys.Update(w)
	}

	// Verify squad exists and has correct membership
	if squad, exists := sys.squads[squadID]; exists {
		if len(squad.Members) != 3 {
			t.Errorf("Squad should have 3 members, got %v", len(squad.Members))
		}
		if squad.FocusTargetID == "" {
			t.Error("Squad should have selected a focus target")
		}
		if squad.AlertLevel <= 0 {
			t.Error("Squad should be on alert")
		}
	} else {
		t.Error("Squad should exist")
	}
}

func BenchmarkRoleBasedAISystem_Update(b *testing.B) {
	sys := NewRoleBasedAISystem()
	w := engine.NewWorld()

	player := w.AddEntity()
	w.AddComponent(player, &PositionComponent{X: 500, Y: 500})

	// Create 20 enemies across 4 squads
	for squadIdx := 0; squadIdx < 4; squadIdx++ {
		squadID := string(rune('A' + squadIdx))
		for i := 0; i < 5; i++ {
			e := w.AddEntity()
			role := EnemyRole(i % 5)
			w.AddComponent(e, &EnemyRoleComponent{Role: role, Config: GetRoleConfig(role, "fantasy")})
			w.AddComponent(e, &PositionComponent{X: float64(i * 50), Y: float64(squadIdx * 50)})
			w.AddComponent(e, &HealthComponent{Current: 100, Max: 100})
			w.AddComponent(e, &SquadTacticsComponent{SquadID: squadID})
			w.AddComponent(e, &TargetComponent{TargetID: player})
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w)
	}
}

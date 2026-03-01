package ai

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/combat"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
)

func TestNewTelegraphBehaviorTree(t *testing.T) {
	tree := NewTelegraphBehaviorTree()

	if tree == nil {
		t.Fatal("expected non-nil behavior tree")
	}

	if tree.Root == nil {
		t.Fatal("expected non-nil root node")
	}
}

func TestTelegraphBehaviorTree_Integration(t *testing.T) {
	w := engine.NewWorld()
	sys := combat.NewTelegraphSystem("fantasy", 42)

	// Setup enemy entity
	enemy := w.AddEntity()
	w.AddComponent(enemy, &engine.Position{X: 100, Y: 100})
	w.AddComponent(enemy, &engine.Health{Current: 100, Max: 100})
	w.AddArchetypeComponent(enemy, engine.ComponentIDPosition)
	w.AddArchetypeComponent(enemy, engine.ComponentIDHealth)

	telegraph := &combat.TelegraphComponent{
		Phase: combat.PhaseInactive,
	}
	w.AddComponent(enemy, telegraph)

	agent := &Agent{
		X:                  100,
		Y:                  100,
		Health:             100,
		MaxHealth:          100,
		AttackRange:        60,
		RetreatHealthRatio: 0.25,
	}

	tileMap := make([][]int, 20)
	for i := range tileMap {
		tileMap[i] = make([]int, 20)
	}

	ctx := &TelegraphAttackContext{
		Context: &Context{
			TileMap: tileMap,
			PlayerX: 130,
			PlayerY: 100,
			RNG:     rng.NewRNG(42),
		},
		World:           w,
		TelegraphSystem: sys,
	}

	tree := NewTelegraphBehaviorTree()

	t.Run("uses telegraph when available", func(t *testing.T) {
		tree.Tick(agent, ctx.Context)

		// Should have initiated telegraph attack
		if telegraph.Phase == combat.PhaseInactive {
			t.Error("expected telegraph to be initiated")
		}
		if agent.State != StateAttack {
			t.Errorf("expected attack state, got %v", agent.State)
		}
	})

	t.Run("retreats when low health", func(t *testing.T) {
		telegraph.Phase = combat.PhaseInactive
		agent.Health = 20
		agent.State = StateIdle

		tree.Tick(agent, ctx.Context)

		if agent.State != StateRetreat {
			t.Errorf("expected retreat state when low health, got %v", agent.State)
		}
	})
}

func TestFindEntityByAgent(t *testing.T) {
	w := engine.NewWorld()

	// Create multiple entities
	e1 := w.AddEntity()
	w.AddComponent(e1, &engine.Position{X: 100, Y: 100})
	w.AddArchetypeComponent(e1, engine.ComponentIDPosition)

	e2 := w.AddEntity()
	w.AddComponent(e2, &engine.Position{X: 200, Y: 200})
	w.AddArchetypeComponent(e2, engine.ComponentIDPosition)

	t.Run("finds matching entity", func(t *testing.T) {
		agent := &Agent{X: 100, Y: 100}

		entity := findEntityByAgent(w, agent)

		if entity != e1 {
			t.Errorf("expected entity %d, got %d", e1, entity)
		}
	})

	t.Run("finds second entity", func(t *testing.T) {
		agent := &Agent{X: 200, Y: 200}

		entity := findEntityByAgent(w, agent)

		if entity != e2 {
			t.Errorf("expected entity %d, got %d", e2, entity)
		}
	})

	t.Run("returns zero for no match", func(t *testing.T) {
		agent := &Agent{X: 500, Y: 500}

		entity := findEntityByAgent(w, agent)

		if entity != 0 {
			t.Errorf("expected entity 0 for no match, got %d", entity)
		}
	})

	t.Run("handles close positions", func(t *testing.T) {
		agent := &Agent{X: 100.05, Y: 100.05}

		entity := findEntityByAgent(w, agent)

		if entity != e1 {
			t.Error("should match within tolerance")
		}
	})
}

func TestAddTelegraphToAgent(t *testing.T) {
	w := engine.NewWorld()

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 100, Y: 100})
	w.AddArchetypeComponent(entity, engine.ComponentIDPosition)

	AddTelegraphToAgent(w, entity, "fantasy", 42)

	// Verify component was added
	tComp, ok := w.GetComponent(entity, reflect.TypeOf(&combat.TelegraphComponent{}))
	if !ok {
		t.Fatal("telegraph component not added")
	}

	telegraph := tComp.(*combat.TelegraphComponent)

	if telegraph.Phase != combat.PhaseInactive {
		t.Error("expected inactive phase initially")
	}

	if telegraph.Seed != 42 {
		t.Errorf("expected seed 42, got %d", telegraph.Seed)
	}
}

package ai

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/opd-ai/violence/pkg/telegraph"
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
	sys := telegraph.NewSystem("fantasy", 42)

	// Setup enemy entity
	enemy := w.AddEntity()
	w.AddComponent(enemy, &engine.Position{X: 100, Y: 100})
	w.AddComponent(enemy, &engine.Health{Current: 100, Max: 100})
	w.AddArchetypeComponent(enemy, engine.ComponentIDPosition)
	w.AddArchetypeComponent(enemy, engine.ComponentIDHealth)

	telegraphComp := &telegraph.Component{
		Active: false,
	}
	w.AddComponent(enemy, telegraphComp)

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

	baseCtx := &Context{
		TileMap: tileMap,
		PlayerX: 130,
		PlayerY: 100,
		RNG:     rng.NewRNG(42),
	}

	ctx := &TelegraphAttackContext{
		Context:         baseCtx,
		World:           w,
		TelegraphSystem: sys,
	}
	// Set back-reference so actionTelegraphAttack can recover ctx via ctx.Context.Extension.
	baseCtx.Extension = ctx

	tree := NewTelegraphBehaviorTree()

	t.Run("tree creation", func(t *testing.T) {
		if tree == nil || tree.Root == nil {
			t.Fatal("tree should have a valid root")
		}
	})

	t.Run("can initiate telegraph", func(t *testing.T) {
		// Manually test the action function
		status := actionTelegraphAttack(agent, ctx.Context)

		if status != StatusSuccess {
			t.Errorf("expected success status, got %v", status)
		}

		// Should have initiated telegraph
		if !telegraphComp.Active {
			t.Error("expected telegraph to be initiated")
		}
		if agent.State != StateAttack {
			t.Errorf("expected attack state, got %v", agent.State)
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
	tComp, ok := w.GetComponent(entity, reflect.TypeOf(&telegraph.Component{}))
	if !ok {
		t.Fatal("telegraph component not added")
	}

	telegraphComp := tComp.(*telegraph.Component)

	if telegraphComp.Active {
		t.Error("expected inactive initially")
	}

	if telegraphComp.TelegraphTime <= 0 {
		t.Errorf("expected positive telegraph time, got %f", telegraphComp.TelegraphTime)
	}
}

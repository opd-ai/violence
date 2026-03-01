// Package ai - Telegraph attack integration
package ai

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/combat"
	"github.com/opd-ai/violence/pkg/engine"
)

// TelegraphAttackContext extends Context with telegraph system access.
type TelegraphAttackContext struct {
	*Context
	World           *engine.World
	TelegraphSystem *combat.TelegraphSystem
}

// checkCanTelegraphAttack checks if entity can initiate a telegraph attack.
func checkCanTelegraphAttack(agent *Agent, ctx *Context) bool {
	// Check if we have a TelegraphAttackContext
	if tCtx, ok := interface{}(ctx).(*TelegraphAttackContext); ok {
		// Find entity by agent ID
		entity := findEntityByAgent(tCtx.World, agent)
		if entity == 0 {
			return false
		}

		// Check if telegraph system allows attack
		return tCtx.TelegraphSystem.CanAttack(tCtx.World, entity)
	}
	return false
}

// actionTelegraphAttack initiates a telegraph attack toward the player.
func actionTelegraphAttack(agent *Agent, ctx *Context) NodeStatus {
	tCtx, ok := interface{}(ctx).(*TelegraphAttackContext)
	if !ok {
		return StatusFailure
	}

	entity := findEntityByAgent(tCtx.World, agent)
	if entity == 0 {
		return StatusFailure
	}

	// Initiate attack toward player
	success := tCtx.TelegraphSystem.InitiateAttack(tCtx.World, entity, ctx.PlayerX, ctx.PlayerY)
	if !success {
		return StatusFailure
	}

	agent.State = StateAttack

	// Face the player
	dx := ctx.PlayerX - agent.X
	dy := ctx.PlayerY - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > 0.01 {
		agent.DirX = dx / dist
		agent.DirY = dy / dist
	}

	return StatusSuccess
}

// NewTelegraphBehaviorTree creates a behavior tree that uses telegraph attacks.
func NewTelegraphBehaviorTree() *BehaviorTree {
	// Enhanced tree with telegraph attack priority
	root := NewSelector(
		// Retreat if low health
		NewSequence(
			NewCondition(checkLowHealth),
			NewAction(actionRetreat),
		),
		// Telegraph attack if player in sight, in range, and can attack
		NewSequence(
			NewCondition(checkCanSeePlayer),
			NewCondition(checkInAttackRange),
			NewCondition(checkCanTelegraphAttack),
			NewAction(actionTelegraphAttack),
		),
		// Fallback to regular attack if telegraph is on cooldown
		NewSequence(
			NewCondition(checkCanSeePlayer),
			NewCondition(checkInAttackRange),
			NewAction(actionAttack),
		),
		// Strafe if player in sight but can't attack
		NewSequence(
			NewCondition(checkCanSeePlayer),
			NewAction(actionStrafe),
		),
		// Chase if player visible
		NewSequence(
			NewCondition(checkCanSeePlayer),
			NewAction(actionChase),
		),
		// Investigate gunshot
		NewSequence(
			NewCondition(checkHeardGunshot),
			NewAction(actionAlert),
		),
		// Patrol
		NewAction(actionPatrol),
	)
	return &BehaviorTree{Root: root}
}

// findEntityByAgent locates the engine.Entity corresponding to an AI agent.
// This is a helper for linking the legacy AI system to ECS.
func findEntityByAgent(w *engine.World, agent *Agent) engine.Entity {
	// Query all entities with position
	it := w.QueryWithBitmask(engine.ComponentIDPosition)

	posType := reflect.TypeOf(&engine.Position{})

	for it.Next() {
		e := it.Entity()

		posComp, ok := w.GetComponent(e, posType)
		if !ok {
			continue
		}

		pos := posComp.(*engine.Position)

		// Match by position (assumes agent.X/Y are kept in sync)
		dx := pos.X - agent.X
		dy := pos.Y - agent.Y
		if math.Abs(dx) < 0.1 && math.Abs(dy) < 0.1 {
			return e
		}
	}

	return 0
}

// AddTelegraphToAgent adds a telegraph component to an entity for an AI agent.
func AddTelegraphToAgent(w *engine.World, entity engine.Entity, genreID string, seed int64) {
	// Create telegraph component with default inactive state
	telegraph := &combat.TelegraphComponent{
		Phase: combat.PhaseInactive,
		Seed:  seed,
	}

	w.AddComponent(entity, telegraph)
}

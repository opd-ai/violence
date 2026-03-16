// Package ai - Telegraph attack integration
package ai

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/telegraph"
)

// TelegraphAttackContext extends Context with telegraph system access.
type TelegraphAttackContext struct {
	*Context
	World           *engine.World
	TelegraphSystem *telegraph.System
}

// checkCanTelegraphAttack checks if entity can initiate a telegraph attack.
func checkCanTelegraphAttack(agent *Agent, ctx *Context) bool {
	// Check if we have a TelegraphAttackContext via the Extension back-reference
	tCtx, ok := ctx.Extension.(*TelegraphAttackContext)
	if !ok {
		return false
	}
	// Find entity by agent ID
	entity := findEntityByAgent(tCtx.World, agent)
	if entity == 0 {
		return false
	}

	// Check if telegraph component exists and is not active
	telegraphType := reflect.TypeOf(&telegraph.Component{})
	comp, ok := tCtx.World.GetComponent(entity, telegraphType)
	if !ok {
		return true // No component yet, can create one
	}

	tc := comp.(*telegraph.Component)
	return !tc.Active // Can attack if not already telegraphing
}

// actionTelegraphAttack initiates a telegraph attack toward the player.
func actionTelegraphAttack(agent *Agent, ctx *Context) NodeStatus {
	tCtx, ok := ctx.Extension.(*TelegraphAttackContext)
	if !ok {
		return StatusFailure
	}

	entity := findEntityByAgent(tCtx.World, agent)
	if entity == 0 {
		return StatusFailure
	}

	// Determine attack type based on distance and agent role
	dx := ctx.PlayerX - agent.X
	dy := ctx.PlayerY - agent.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	attackType := "melee"
	duration := 0.8

	if dist > 3.0 {
		attackType = "ranged"
		duration = 1.0
	} else if dist < 1.5 {
		attackType = "charge"
		duration = 0.5
	}

	// Initiate attack telegraph
	tCtx.TelegraphSystem.StartTelegraph(tCtx.World, entity, attackType, duration)

	agent.State = StateAttack

	// Face the player
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
	telegraphComp := &telegraph.Component{
		Active:        false,
		TelegraphTime: 1.0,
		AttackType:    "melee",
	}

	w.AddComponent(entity, telegraphComp)
}

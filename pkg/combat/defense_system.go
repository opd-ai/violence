// Package combat - Defense system for processing defensive maneuvers
package combat

import (
	"math"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// DefenseSystem processes defensive maneuver components.
type DefenseSystem struct {
	genreID string
	logger  *logrus.Entry
}

// NewDefenseSystem creates a defense processing system.
func NewDefenseSystem(genreID string) *DefenseSystem {
	return &DefenseSystem{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "defense",
			"genre":  genreID,
		}),
	}
}

// Update processes all defense components.
func (s *DefenseSystem) Update(w *engine.World) {
	posType := reflect.TypeOf(&engine.Position{})
	defenseType := reflect.TypeOf(&DefenseComponent{})

	entities := w.Query(posType, defenseType)

	for _, e := range entities {
		posComp, _ := w.GetComponent(e, posType)
		pos := posComp.(*engine.Position)

		defComp, _ := w.GetComponent(e, defenseType)
		defense := defComp.(*DefenseComponent)

		s.updateDefense(w, e, pos, defense)
	}
}

func (s *DefenseSystem) updateDefense(w *engine.World, entity engine.Entity, pos *engine.Position, def *DefenseComponent) {
	deltaTime := 1.0 / 60.0

	// Regenerate stamina when not actively defending
	if def.State == DefenseInactive {
		def.StaminaCurrent = math.Min(def.StaminaMax, def.StaminaCurrent+def.StaminaRegen*deltaTime)
	}

	// Update cooldown
	if def.CooldownTimer > 0 {
		def.CooldownTimer -= deltaTime
		if def.CooldownTimer < 0 {
			def.CooldownTimer = 0
		}
	}

	// Update buff timers
	if def.PerfectParryTimer > 0 {
		def.PerfectParryTimer -= deltaTime
		if def.PerfectParryTimer <= 0 {
			def.PerfectParryBuff = false
		}
	}

	if def.BackstabTimer > 0 {
		def.BackstabTimer -= deltaTime
		if def.BackstabTimer <= 0 {
			def.BackstabWindow = false
		}
	}

	// Update i-frame timer
	if def.DodgeIFrameTimer > 0 {
		def.DodgeIFrameTimer -= deltaTime
	}

	// Process defense state machine
	switch def.State {
	case DefenseInactive:
		// Waiting for input
		return

	case DefenseWindup:
		def.StateTimer -= deltaTime
		if def.StateTimer <= 0 {
			s.activateDefense(w, entity, pos, def)
		}

	case DefenseActive:
		def.StateTimer -= deltaTime
		s.processActiveDefense(w, entity, pos, def, deltaTime)

		if def.StateTimer <= 0 {
			s.enterRecovery(def)
		}

	case DefenseRecovery:
		def.StateTimer -= deltaTime
		if def.StateTimer <= 0 {
			def.State = DefenseInactive
			def.Type = DefenseNone
		}
	}
}

func (s *DefenseSystem) activateDefense(w *engine.World, entity engine.Entity, pos *engine.Position, def *DefenseComponent) {
	def.State = DefenseActive

	switch def.Type {
	case DefenseDodge:
		def.StateTimer = def.DodgeIFrames + 0.1
		def.DodgeIFrameTimer = def.DodgeIFrames

		// Apply dodge velocity
		velocityType := reflect.TypeOf(&engine.Velocity{})
		if w.HasComponent(entity, velocityType) {
			velComp, _ := w.GetComponent(entity, velocityType)
			vel := velComp.(*engine.Velocity)

			vel.DX = def.DodgeVelocityX * def.DodgeDistance / def.DodgeIFrames
			vel.DY = def.DodgeVelocityY * def.DodgeDistance / def.DodgeIFrames
		}

		s.logger.WithFields(logrus.Fields{
			"entity":    entity,
			"direction": def.DodgeDirection,
		}).Debug("dodge activated")

	case DefenseParry:
		def.StateTimer = def.ParryWindowEnd
		s.logger.WithFields(logrus.Fields{
			"entity": entity,
			"window": def.ParryWindowEnd,
		}).Debug("parry activated")

	case DefenseBlock:
		def.StateTimer = 10.0
		s.logger.WithFields(logrus.Fields{
			"entity": entity,
		}).Debug("block activated")
	}
}

func (s *DefenseSystem) processActiveDefense(w *engine.World, entity engine.Entity, pos *engine.Position, def *DefenseComponent, deltaTime float64) {
	switch def.Type {
	case DefenseDodge:
		// Dodge velocity handled by movement system
		// Just track i-frames

	case DefenseParry:
		// Parry window active - telegraph system checks this

	case DefenseBlock:
		// Drain stamina while blocking
		def.StaminaCurrent -= def.BlockStaminaDrain * deltaTime
		if def.StaminaCurrent < 0 {
			def.StaminaCurrent = 0
			s.enterRecovery(def)
		}
	}
}

func (s *DefenseSystem) enterRecovery(def *DefenseComponent) {
	def.State = DefenseRecovery
	def.StateTimer = 0.2

	switch def.Type {
	case DefenseDodge:
		def.CooldownTimer = def.DodgeCooldown
	case DefenseParry:
		def.CooldownTimer = def.ParryCooldown
	case DefenseBlock:
		def.CooldownTimer = def.BlockCooldown
	}
}

// InitiateDodge starts a dodge maneuver.
func (s *DefenseSystem) InitiateDodge(def *DefenseComponent, dirX, dirY float64, genreID string) bool {
	if !def.CanDodge() {
		return false
	}

	preset := GetDefensePreset(genreID)
	def.StaminaCurrent -= preset.DodgeCost

	def.Type = DefenseDodge
	def.State = DefenseWindup
	def.StateTimer = 0.05

	// Normalize direction
	mag := math.Sqrt(dirX*dirX + dirY*dirY)
	if mag > 0 {
		def.DodgeVelocityX = dirX / mag
		def.DodgeVelocityY = dirY / mag
	} else {
		// Default backward dodge
		def.DodgeVelocityX = 0
		def.DodgeVelocityY = -1
	}
	def.DodgeDirection = math.Atan2(def.DodgeVelocityY, def.DodgeVelocityX)

	return true
}

// InitiateParry starts a parry maneuver.
func (s *DefenseSystem) InitiateParry(def *DefenseComponent, genreID string) bool {
	if !def.CanParry() {
		return false
	}

	preset := GetDefensePreset(genreID)
	def.StaminaCurrent -= preset.ParryCost

	def.Type = DefenseParry
	def.State = DefenseWindup
	def.StateTimer = 0.05

	return true
}

// InitiateBlock starts blocking.
func (s *DefenseSystem) InitiateBlock(def *DefenseComponent, genreID string) bool {
	if !def.CanBlock() {
		return false
	}

	preset := GetDefensePreset(genreID)
	def.StaminaCurrent -= preset.BlockCost

	def.Type = DefenseBlock
	def.State = DefenseWindup
	def.StateTimer = 0.05

	return true
}

// CancelBlock stops an active block.
func (s *DefenseSystem) CancelBlock(def *DefenseComponent) {
	if def.Type == DefenseBlock {
		s.enterRecovery(def)
	}
}

// ProcessIncomingDamage modifies damage based on active defense.
// Returns modified damage and whether attack was fully negated.
func ProcessIncomingDamage(def *DefenseComponent, damage, attackerAngle, defenderFacing float64) (float64, bool) {
	// Check invulnerability frames
	if def.IsInvulnerable() {
		return 0, true
	}

	// Check parry
	if def.IsParrying() {
		// Perfect parry negates damage and applies buff
		if def.IsPerfectParryWindow() {
			def.PerfectParryBuff = true
			def.PerfectParryTimer = 2.0
			def.BackstabWindow = true
			def.BackstabTimer = 1.5
			return 0, true
		}
		// Regular parry reduces damage
		return damage * 0.3, false
	}

	// Check block
	if def.IsBlocking() {
		// Check if attack is within block arc
		angleDiff := math.Abs(normalizeAngle(attackerAngle - defenderFacing))
		if angleDiff <= def.BlockArc/2 {
			return damage * (1.0 - def.BlockDamageReduction), false
		}
	}

	return damage, false
}

// GetCounterDamageMultiplier returns damage multiplier for counter-attacks.
func GetCounterDamageMultiplier(def *DefenseComponent) float64 {
	mul := 1.0
	if def.PerfectParryBuff {
		mul *= def.ParryCounterMul
	}
	if def.BackstabWindow {
		mul *= 1.5
	}
	return mul
}

// Package combat - Combo system for managing attack chains
//
// The combo system enables timing-based weapon attack chains that reward skillful play.
// Players and enemies can execute multi-hit combos with escalating damage and effects.
//
// Usage:
//   - System automatically registers via main.go initialization
//   - Call InitiateCombo(world, entity, weaponType) to start a combo chain
//   - Subsequent attack inputs within timing windows advance the combo
//   - GetComboMultiplier returns current damage/speed/range/knockback modifiers
//   - Combos break if timing windows expire or entity is staggered
//
// Integration:
//   - Weapon attacks should check IsInCombo and apply GetComboMultiplier to stats
//   - Hit effects (particles, screen shake) scale with combo step
//   - Enemy AI can break player combos by landing hits (via BreakEntityCombo)
package combat

import (
	"math"
	"math/rand"
	"reflect"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
	"github.com/sirupsen/logrus"
)

// ComboSystem manages weapon combo chains and timing.
type ComboSystem struct {
	genreID string
	rng     *rng.RNG
	seed    int64
	chains  []ComboChain
	logger  *logrus.Entry
}

// NewComboSystem creates a combo system for the specified genre.
func NewComboSystem(genreID string, seed int64) *ComboSystem {
	r := rng.NewRNG(uint64(seed))
	stdRng := rand.New(rand.NewSource(seed))
	return &ComboSystem{
		genreID: genreID,
		rng:     r,
		seed:    seed,
		chains:  DefaultChains(genreID, stdRng),
		logger: logrus.WithFields(logrus.Fields{
			"system": "combo",
			"genre":  genreID,
		}),
	}
}

// Update processes all combo components, tracking timing windows and advancing chains.
func (s *ComboSystem) Update(w *engine.World) {
	deltaTime := 1.0 / 60.0 // 60 FPS

	it := w.QueryWithBitmask(engine.ComponentIDPosition)

	posType := reflect.TypeOf(&engine.Position{})
	comboType := reflect.TypeOf(&ComboComponent{})

	for it.Next() {
		e := it.Entity()

		comboComp, hasCombo := w.GetComponent(e, comboType)
		if !hasCombo {
			continue
		}

		combo := comboComp.(*ComboComponent)
		pos, _ := w.GetComponent(e, posType)
		position := pos.(*engine.Position)

		s.updateCombo(w, e, combo, position, deltaTime)
	}
}

func (s *ComboSystem) updateCombo(w *engine.World, entity engine.Entity, combo *ComboComponent, pos *engine.Position, deltaTime float64) {
	if combo.State == ComboStateNone {
		// Idle state, waiting for first attack input
		return
	}

	if combo.State == ComboStateBroken {
		// Combo was broken, wait a frame then reset
		combo.TimeSinceHit += deltaTime
		if combo.TimeSinceHit > 0.1 {
			ResetCombo(combo)
		}
		return
	}

	// Active combo state - track timing window
	combo.TimeSinceHit += deltaTime

	// Find the combo chain
	var chain *ComboChain
	for i := range s.chains {
		if s.chains[i].ID == combo.ChainID {
			chain = &s.chains[i]
			break
		}
	}

	if chain == nil {
		s.logger.WithField("chain_id", combo.ChainID).Warn("combo chain not found")
		ResetCombo(combo)
		return
	}

	// Check if we're still within a valid step
	if combo.CurrentStep >= len(chain.Steps) {
		// Combo complete
		s.logger.WithFields(logrus.Fields{
			"entity": entity,
			"hits":   combo.TotalHits,
			"damage": combo.TotalDamage,
		}).Info("combo completed")
		ResetCombo(combo)
		return
	}

	step := chain.Steps[combo.CurrentStep]

	// Check if timing window has expired
	if combo.TimeSinceHit > step.WindowEnd {
		// Window expired, break combo
		s.logger.WithFields(logrus.Fields{
			"entity":    entity,
			"step":      combo.CurrentStep,
			"step_name": step.Name,
			"time":      combo.TimeSinceHit,
		}).Debug("combo window expired")
		BreakCombo(combo)
		return
	}

	// Check if input was buffered and window is open
	if combo.InputBuffer && combo.TimeSinceHit >= step.WindowStart {
		// Advance to next step
		s.executeComboStep(w, entity, combo, chain, pos)
		AdvanceCombo(combo, chain)
	}
}

func (s *ComboSystem) executeComboStep(w *engine.World, entity engine.Entity, combo *ComboComponent, chain *ComboChain, pos *engine.Position) {
	if combo.CurrentStep >= len(chain.Steps) {
		return
	}

	step := chain.Steps[combo.CurrentStep]

	s.logger.WithFields(logrus.Fields{
		"entity":    entity,
		"step":      combo.CurrentStep + 1,
		"step_name": step.Name,
		"damage":    step.DamageMul,
	}).Info("combo step executed")

	// Apply screen shake effect (stored for renderer to consume)
	if step.ScreenShake > 0 {
		// Add or update shake component
		shakeType := reflect.TypeOf(&ScreenShakeComponent{})
		shakeComp, hasShake := w.GetComponent(entity, shakeType)
		if hasShake {
			shake := shakeComp.(*ScreenShakeComponent)
			shake.Intensity = math.Max(shake.Intensity, step.ScreenShake)
			shake.Duration = 0.15
		} else {
			w.AddComponent(entity, &ScreenShakeComponent{
				Intensity: step.ScreenShake,
				Duration:  0.15,
			})
		}
	}

	// Update combo tracking
	combo.TotalHits++
}

// InitiateCombo starts a new combo chain for an entity.
func (s *ComboSystem) InitiateCombo(w *engine.World, entity engine.Entity, weaponType string) bool {
	comboType := reflect.TypeOf(&ComboComponent{})

	comboComp, hasCombo := w.GetComponent(entity, comboType)
	if !hasCombo {
		// Create new combo component
		combo := &ComboComponent{
			State:       ComboStateNone,
			CurrentStep: 0,
		}

		// Select appropriate chain
		stdRng := rand.New(rand.NewSource(s.seed))
		chain := SelectChain(s.chains, weaponType, stdRng)
		if chain == nil {
			s.logger.WithField("weapon_type", weaponType).Warn("no combo chain found")
			return false
		}

		combo.ChainID = chain.ID
		w.AddComponent(entity, combo)
		comboComp = combo
	}

	combo := comboComp.(*ComboComponent)

	// If combo is already active, buffer the input
	if combo.State == ComboStateActive {
		combo.InputBuffer = true
		return true
	}

	// Start new combo
	stdRng := rand.New(rand.NewSource(s.seed))
	chain := SelectChain(s.chains, weaponType, stdRng)
	if chain == nil {
		return false
	}

	combo.ChainID = chain.ID
	combo.State = ComboStateActive
	combo.CurrentStep = 0
	combo.TimeSinceHit = 0
	combo.InputBuffer = false
	combo.TotalHits = 0
	combo.TotalDamage = 0

	s.logger.WithFields(logrus.Fields{
		"entity":   entity,
		"chain":    chain.Name,
		"chain_id": chain.ID,
	}).Info("combo initiated")

	// Execute first step immediately
	posType := reflect.TypeOf(&engine.Position{})
	posComp, _ := w.GetComponent(entity, posType)
	pos := posComp.(*engine.Position)

	s.executeComboStep(w, entity, combo, chain, pos)

	return true
}

// GetComboMultiplier returns the damage/speed multipliers for the current combo step.
func (s *ComboSystem) GetComboMultiplier(w *engine.World, entity engine.Entity) (damageMul, speedMul, rangeMul, knockbackMul float64) {
	comboType := reflect.TypeOf(&ComboComponent{})

	comboComp, hasCombo := w.GetComponent(entity, comboType)
	if !hasCombo {
		return 1.0, 1.0, 1.0, 1.0
	}

	combo := comboComp.(*ComboComponent)
	if combo.State != ComboStateActive {
		return 1.0, 1.0, 1.0, 1.0
	}

	// Find chain
	var chain *ComboChain
	for i := range s.chains {
		if s.chains[i].ID == combo.ChainID {
			chain = &s.chains[i]
			break
		}
	}

	if chain == nil || combo.CurrentStep >= len(chain.Steps) {
		return 1.0, 1.0, 1.0, 1.0
	}

	step := chain.Steps[combo.CurrentStep]
	return step.DamageMul, step.SpeedMul, step.RangeMul, step.KnockbackMul
}

// IsInCombo returns true if the entity is currently in an active combo.
func (s *ComboSystem) IsInCombo(w *engine.World, entity engine.Entity) bool {
	comboType := reflect.TypeOf(&ComboComponent{})

	comboComp, hasCombo := w.GetComponent(entity, comboType)
	if !hasCombo {
		return false
	}

	combo := comboComp.(*ComboComponent)
	return combo.State == ComboStateActive
}

// GetComboStep returns the current combo step number (0-indexed).
func (s *ComboSystem) GetComboStep(w *engine.World, entity engine.Entity) int {
	comboType := reflect.TypeOf(&ComboComponent{})

	comboComp, hasCombo := w.GetComponent(entity, comboType)
	if !hasCombo {
		return -1
	}

	combo := comboComp.(*ComboComponent)
	if combo.State != ComboStateActive {
		return -1
	}

	return combo.CurrentStep
}

// BreakEntityCombo forcibly breaks an entity's combo (e.g., on hit/stagger).
func (s *ComboSystem) BreakEntityCombo(w *engine.World, entity engine.Entity) {
	comboType := reflect.TypeOf(&ComboComponent{})

	comboComp, hasCombo := w.GetComponent(entity, comboType)
	if !hasCombo {
		return
	}

	combo := comboComp.(*ComboComponent)
	if combo.State == ComboStateActive {
		s.logger.WithFields(logrus.Fields{
			"entity": entity,
			"step":   combo.CurrentStep,
		}).Debug("combo broken externally")
		BreakCombo(combo)
	}
}

// ScreenShakeComponent stores screen shake effect data.
type ScreenShakeComponent struct {
	Intensity float64
	Duration  float64
}

// Type implements Component interface.
func (s *ScreenShakeComponent) Type() string {
	return "screen_shake"
}

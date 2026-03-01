// Package combat - Tests for combo system
package combat

import (
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestComboComponent(t *testing.T) {
	combo := &ComboComponent{
		ChainID:     "test_chain",
		CurrentStep: 0,
		State:       ComboStateNone,
	}

	if combo.Type() != "combo" {
		t.Errorf("expected type 'combo', got %s", combo.Type())
	}
}

func TestDefaultChains(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			chains := DefaultChains(genre, nil)
			if len(chains) == 0 {
				t.Errorf("expected chains for genre %s, got none", genre)
			}

			for _, chain := range chains {
				if chain.ID == "" {
					t.Error("chain missing ID")
				}
				if chain.Name == "" {
					t.Error("chain missing name")
				}
				if len(chain.Steps) == 0 {
					t.Errorf("chain %s has no steps", chain.ID)
				}

				// Validate steps
				for i, step := range chain.Steps {
					if step.Name == "" {
						t.Errorf("step %d in chain %s missing name", i, chain.ID)
					}
					if step.DamageMul <= 0 {
						t.Errorf("step %s has invalid damage multiplier: %f", step.Name, step.DamageMul)
					}
					if step.WindowEnd <= step.WindowStart {
						t.Errorf("step %s has invalid timing window: [%f, %f]", step.Name, step.WindowStart, step.WindowEnd)
					}
					if step.StaggerChance < 0 || step.StaggerChance > 1 {
						t.Errorf("step %s has invalid stagger chance: %f", step.Name, step.StaggerChance)
					}
				}
			}
		})
	}
}

func TestSelectChain(t *testing.T) {
	chains := []ComboChain{
		{ID: "melee1", WeaponType: "melee"},
		{ID: "melee2", WeaponType: "melee"},
		{ID: "ranged1", WeaponType: "ranged"},
	}

	// Test selecting melee chain
	chain := SelectChain(chains, "melee", nil)
	if chain == nil {
		t.Fatal("expected melee chain, got nil")
	}
	if chain.WeaponType != "melee" {
		t.Errorf("expected melee weapon type, got %s", chain.WeaponType)
	}

	// Test selecting ranged chain
	chain = SelectChain(chains, "ranged", nil)
	if chain == nil {
		t.Fatal("expected ranged chain, got nil")
	}
	if chain.WeaponType != "ranged" {
		t.Errorf("expected ranged weapon type, got %s", chain.WeaponType)
	}

	// Test fallback for non-existent type
	chain = SelectChain(chains, "energy", nil)
	if chain == nil {
		t.Error("expected fallback chain, got nil")
	}

	// Test empty chains
	chain = SelectChain([]ComboChain{}, "melee", nil)
	if chain != nil {
		t.Error("expected nil for empty chains")
	}
}

func TestResetCombo(t *testing.T) {
	combo := &ComboComponent{
		State:        ComboStateActive,
		CurrentStep:  2,
		TimeSinceHit: 1.5,
		InputBuffer:  true,
	}

	ResetCombo(combo)

	if combo.State != ComboStateNone {
		t.Errorf("expected state None, got %d", combo.State)
	}
	if combo.CurrentStep != 0 {
		t.Errorf("expected step 0, got %d", combo.CurrentStep)
	}
	if combo.TimeSinceHit != 0 {
		t.Errorf("expected time 0, got %f", combo.TimeSinceHit)
	}
	if combo.InputBuffer {
		t.Error("expected input buffer cleared")
	}
}

func TestAdvanceCombo(t *testing.T) {
	chain := &ComboChain{
		Steps: []ComboStep{
			{Name: "Step1"},
			{Name: "Step2"},
			{Name: "Step3"},
		},
	}

	combo := &ComboComponent{
		State:       ComboStateActive,
		CurrentStep: 0,
	}

	// Advance to step 1
	advanced := AdvanceCombo(combo, chain)
	if !advanced {
		t.Error("expected combo to advance")
	}
	if combo.CurrentStep != 1 {
		t.Errorf("expected step 1, got %d", combo.CurrentStep)
	}
	if combo.State != ComboStateActive {
		t.Errorf("expected state Active, got %d", combo.State)
	}

	// Advance to step 2
	advanced = AdvanceCombo(combo, chain)
	if !advanced {
		t.Error("expected combo to advance")
	}
	if combo.CurrentStep != 2 {
		t.Errorf("expected step 2, got %d", combo.CurrentStep)
	}

	// Try to advance beyond chain length
	advanced = AdvanceCombo(combo, chain)
	if advanced {
		t.Error("expected combo to not advance beyond chain length")
	}
	if combo.State != ComboStateNone {
		t.Errorf("expected state None after completion, got %d", combo.State)
	}
	if combo.CurrentStep != 0 {
		t.Errorf("expected step reset to 0, got %d", combo.CurrentStep)
	}
}

func TestBreakCombo(t *testing.T) {
	combo := &ComboComponent{
		State:        ComboStateActive,
		CurrentStep:  3,
		HighestCombo: 2,
	}

	BreakCombo(combo)

	if combo.State != ComboStateBroken {
		t.Errorf("expected state Broken, got %d", combo.State)
	}
	if combo.TimeSinceHit != 0 {
		t.Errorf("expected time reset, got %f", combo.TimeSinceHit)
	}
	if combo.HighestCombo != 3 {
		t.Errorf("expected highest combo 3, got %d", combo.HighestCombo)
	}
}

func TestComboSystemCreation(t *testing.T) {
	sys := NewComboSystem("fantasy", 42)
	if sys == nil {
		t.Fatal("expected combo system, got nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("expected genre 'fantasy', got %s", sys.genreID)
	}
	if len(sys.chains) == 0 {
		t.Error("expected chains to be initialized")
	}
}

func TestComboSystemInitiate(t *testing.T) {
	w := engine.NewWorld()
	sys := NewComboSystem("fantasy", 42)

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})
	w.SetArchetype(entity, engine.ComponentIDPosition)

	// Initiate combo
	success := sys.InitiateCombo(w, entity, "melee")
	if !success {
		t.Fatal("expected combo initiation to succeed")
	}

	// Check combo component was added
	comboType := reflect.TypeOf(&ComboComponent{})
	comboComp, hasCombo := w.GetComponent(entity, comboType)
	if !hasCombo {
		t.Fatal("expected combo component to be added")
	}

	combo := comboComp.(*ComboComponent)
	if combo.State != ComboStateActive {
		t.Errorf("expected active state, got %d", combo.State)
	}
	if combo.ChainID == "" {
		t.Error("expected chain ID to be set")
	}
}

func TestComboSystemUpdate(t *testing.T) {
	w := engine.NewWorld()
	sys := NewComboSystem("fantasy", 42)

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})
	w.AddComponent(entity, &ComboComponent{
		ChainID:     "sword_basic",
		State:       ComboStateActive,
		CurrentStep: 0,
	})
	w.SetArchetype(entity, engine.ComponentIDPosition)

	// Update (uses internal 60 FPS timing)
	sys.Update(w)

	comboType := reflect.TypeOf(&ComboComponent{})
	comboComp, _ := w.GetComponent(entity, comboType)
	combo := comboComp.(*ComboComponent)

	// Time should advance
	if combo.TimeSinceHit < 0.01 {
		t.Errorf("expected time to advance, got %f", combo.TimeSinceHit)
	}
}

func TestComboSystemGetMultiplier(t *testing.T) {
	w := engine.NewWorld()
	sys := NewComboSystem("fantasy", 42)

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})

	// No combo - should return 1.0 multipliers
	dmg, spd, rng, kb := sys.GetComboMultiplier(w, entity)
	if dmg != 1.0 || spd != 1.0 || rng != 1.0 || kb != 1.0 {
		t.Errorf("expected default multipliers, got dmg=%f spd=%f rng=%f kb=%f", dmg, spd, rng, kb)
	}

	// Add active combo
	sys.InitiateCombo(w, entity, "melee")
	dmg, spd, rng, kb = sys.GetComboMultiplier(w, entity)

	// Should have some multiplier from first step
	if dmg == 1.0 && spd == 1.0 && rng == 1.0 && kb == 1.0 {
		t.Error("expected combo multipliers to differ from defaults")
	}
}

func TestComboSystemIsInCombo(t *testing.T) {
	w := engine.NewWorld()
	sys := NewComboSystem("fantasy", 42)

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})

	// Not in combo initially
	if sys.IsInCombo(w, entity) {
		t.Error("expected entity to not be in combo")
	}

	// Initiate combo
	sys.InitiateCombo(w, entity, "melee")

	if !sys.IsInCombo(w, entity) {
		t.Error("expected entity to be in combo")
	}
}

func TestComboSystemBreak(t *testing.T) {
	w := engine.NewWorld()
	sys := NewComboSystem("fantasy", 42)

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})

	// Initiate and break
	sys.InitiateCombo(w, entity, "melee")
	sys.BreakEntityCombo(w, entity)

	comboType := reflect.TypeOf(&ComboComponent{})
	comboComp, _ := w.GetComponent(entity, comboType)
	combo := comboComp.(*ComboComponent)

	if combo.State != ComboStateBroken {
		t.Errorf("expected broken state, got %d", combo.State)
	}
}

func TestComboSystemGetStep(t *testing.T) {
	w := engine.NewWorld()
	sys := NewComboSystem("fantasy", 42)

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})

	// No combo
	step := sys.GetComboStep(w, entity)
	if step != -1 {
		t.Errorf("expected -1 for no combo, got %d", step)
	}

	// With combo
	sys.InitiateCombo(w, entity, "melee")
	step = sys.GetComboStep(w, entity)
	if step != 0 {
		t.Errorf("expected step 0, got %d", step)
	}
}

func TestScreenShakeComponent(t *testing.T) {
	shake := &ScreenShakeComponent{
		Intensity: 0.05,
		Duration:  0.2,
	}

	if shake.Type() != "screen_shake" {
		t.Errorf("expected type 'screen_shake', got %s", shake.Type())
	}
}

func TestComboChainStepProgression(t *testing.T) {
	chains := DefaultChains("fantasy", nil)
	if len(chains) == 0 {
		t.Fatal("expected at least one chain")
	}

	chain := &chains[0]

	// Test that damage and effects generally increase through combo
	for i := 1; i < len(chain.Steps); i++ {
		prev := chain.Steps[i-1]
		curr := chain.Steps[i]

		// Later steps should generally have higher damage
		if curr.DamageMul < prev.DamageMul {
			t.Logf("step %d damage (%f) less than step %d (%f) - acceptable for variation",
				i, curr.DamageMul, i-1, prev.DamageMul)
		}

		// Later steps should have larger effects
		if curr.ParticleCount < prev.ParticleCount {
			t.Logf("step %d particles (%d) less than step %d (%d)",
				i, curr.ParticleCount, i-1, prev.ParticleCount)
		}
	}
}

func TestComboWindowExpiry(t *testing.T) {
	w := engine.NewWorld()
	sys := NewComboSystem("fantasy", 42)

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})
	w.SetArchetype(entity, engine.ComponentIDPosition)

	// Start combo
	sys.InitiateCombo(w, entity, "melee")

	comboType := reflect.TypeOf(&ComboComponent{})

	// Update many times to expire window
	for i := 0; i < 100; i++ {
		sys.Update(w)
	}

	comboComp, _ := w.GetComponent(entity, comboType)
	combo := comboComp.(*ComboComponent)

	// Combo should have broken due to expired window
	if combo.State == ComboStateActive {
		t.Error("expected combo to break after window expiry")
	}
}

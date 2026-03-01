package combat

import (
	"math"
	"testing"
)

func TestDefensePresets(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset := GetDefensePreset(genre)

			// All values should be positive
			if preset.DodgeDistance <= 0 {
				t.Errorf("DodgeDistance should be positive, got %f", preset.DodgeDistance)
			}
			if preset.DodgeIFrames <= 0 {
				t.Errorf("DodgeIFrames should be positive, got %f", preset.DodgeIFrames)
			}
			if preset.StaminaMax <= 0 {
				t.Errorf("StaminaMax should be positive, got %f", preset.StaminaMax)
			}
			if preset.StaminaRegen <= 0 {
				t.Errorf("StaminaRegen should be positive, got %f", preset.StaminaRegen)
			}

			// Damage reduction should be between 0 and 1
			if preset.BlockReduction < 0 || preset.BlockReduction > 1 {
				t.Errorf("BlockReduction should be in [0,1], got %f", preset.BlockReduction)
			}

			// Parry window should be reasonable
			if preset.ParryWindow <= 0 || preset.ParryWindow > 1.0 {
				t.Errorf("ParryWindow should be in (0,1], got %f", preset.ParryWindow)
			}
			if preset.ParryPerfect > preset.ParryWindow {
				t.Errorf("ParryPerfect (%f) should be <= ParryWindow (%f)", preset.ParryPerfect, preset.ParryWindow)
			}
		})
	}
}

func TestNewDefenseComponent(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			def := NewDefenseComponent(genre)

			if def.State != DefenseInactive {
				t.Errorf("initial state should be inactive, got %v", def.State)
			}
			if def.Type != DefenseNone {
				t.Errorf("initial type should be none, got %v", def.Type)
			}
			if def.StaminaCurrent != def.StaminaMax {
				t.Errorf("stamina should start at max, got %f / %f", def.StaminaCurrent, def.StaminaMax)
			}
		})
	}
}

func TestCanDodge(t *testing.T) {
	def := NewDefenseComponent("fantasy")

	// Should be able to dodge with full stamina
	if !def.CanDodge() {
		t.Error("should be able to dodge with full stamina")
	}

	// Cannot dodge with insufficient stamina
	def.StaminaCurrent = 5.0
	if def.CanDodge() {
		t.Error("should not be able to dodge with low stamina")
	}

	// Cannot dodge on cooldown
	def.StaminaCurrent = def.StaminaMax
	def.CooldownTimer = 0.5
	if def.CanDodge() {
		t.Error("should not be able to dodge on cooldown")
	}

	// Cannot dodge while already defending
	def.CooldownTimer = 0
	def.State = DefenseActive
	if def.CanDodge() {
		t.Error("should not be able to dodge while in active state")
	}
}

func TestCanParry(t *testing.T) {
	def := NewDefenseComponent("fantasy")

	// Should be able to parry with full stamina
	if !def.CanParry() {
		t.Error("should be able to parry with full stamina")
	}

	// Cannot parry with insufficient stamina
	def.StaminaCurrent = 5.0
	if def.CanParry() {
		t.Error("should not be able to parry with low stamina")
	}
}

func TestIsInvulnerable(t *testing.T) {
	def := NewDefenseComponent("fantasy")

	// Not invulnerable by default
	if def.IsInvulnerable() {
		t.Error("should not be invulnerable initially")
	}

	// Invulnerable during dodge i-frames
	def.Type = DefenseDodge
	def.DodgeIFrameTimer = 0.2
	if !def.IsInvulnerable() {
		t.Error("should be invulnerable during dodge i-frames")
	}

	// Not invulnerable after i-frames expire
	def.DodgeIFrameTimer = 0
	if def.IsInvulnerable() {
		t.Error("should not be invulnerable after i-frames expire")
	}
}

func TestIsBlocking(t *testing.T) {
	def := NewDefenseComponent("fantasy")

	// Not blocking initially
	if def.IsBlocking() {
		t.Error("should not be blocking initially")
	}

	// Blocking when active
	def.Type = DefenseBlock
	def.State = DefenseActive
	if !def.IsBlocking() {
		t.Error("should be blocking when in active block state")
	}

	// Not blocking during recovery
	def.State = DefenseRecovery
	if def.IsBlocking() {
		t.Error("should not be blocking during recovery")
	}
}

func TestIsParrying(t *testing.T) {
	def := NewDefenseComponent("fantasy")

	// Parrying when active
	def.Type = DefenseParry
	def.State = DefenseActive
	if !def.IsParrying() {
		t.Error("should be parrying when in active parry state")
	}
}

func TestIsPerfectParryWindow(t *testing.T) {
	def := NewDefenseComponent("fantasy")
	def.Type = DefenseParry
	def.State = DefenseActive
	def.ParryWindowEnd = 0.3
	def.ParryPerfectWindow = 0.1

	// Not in perfect window at start (timer high)
	def.StateTimer = 0.3
	if def.IsPerfectParryWindow() {
		t.Error("should not be in perfect window at start")
	}

	// Not in perfect window mid-way
	def.StateTimer = 0.15
	if def.IsPerfectParryWindow() {
		t.Error("should not be in perfect window mid-way")
	}

	// In perfect window near end (timer low)
	def.StateTimer = 0.08
	if !def.IsPerfectParryWindow() {
		t.Error("should be in perfect window near end")
	}
}

func TestProcessIncomingDamage(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*DefenseComponent)
		damage         float64
		attackerAngle  float64
		defenderFacing float64
		wantDamage     float64
		wantNegated    bool
		tolerance      float64
	}{
		{
			name: "no defense",
			setup: func(d *DefenseComponent) {
				d.State = DefenseInactive
			},
			damage:      100.0,
			wantDamage:  100.0,
			wantNegated: false,
			tolerance:   0.01,
		},
		{
			name: "invulnerable during dodge",
			setup: func(d *DefenseComponent) {
				d.Type = DefenseDodge
				d.DodgeIFrameTimer = 0.2
			},
			damage:      100.0,
			wantDamage:  0.0,
			wantNegated: true,
			tolerance:   0.01,
		},
		{
			name: "perfect parry",
			setup: func(d *DefenseComponent) {
				d.Type = DefenseParry
				d.State = DefenseActive
				d.ParryWindowEnd = 0.3
				d.ParryPerfectWindow = 0.1
				d.StateTimer = 0.08
			},
			damage:      100.0,
			wantDamage:  0.0,
			wantNegated: true,
			tolerance:   0.01,
		},
		{
			name: "regular parry",
			setup: func(d *DefenseComponent) {
				d.Type = DefenseParry
				d.State = DefenseActive
				d.ParryWindowEnd = 0.3
				d.ParryPerfectWindow = 0.1
				d.StateTimer = 0.15
			},
			damage:      100.0,
			wantDamage:  30.0,
			wantNegated: false,
			tolerance:   0.01,
		},
		{
			name: "block from front",
			setup: func(d *DefenseComponent) {
				d.Type = DefenseBlock
				d.State = DefenseActive
				d.BlockDamageReduction = 0.7
				d.BlockArc = math.Pi * 0.75
			},
			damage:         100.0,
			attackerAngle:  0.0,
			defenderFacing: 0.0,
			wantDamage:     30.0,
			wantNegated:    false,
			tolerance:      0.01,
		},
		{
			name: "block from behind",
			setup: func(d *DefenseComponent) {
				d.Type = DefenseBlock
				d.State = DefenseActive
				d.BlockDamageReduction = 0.7
				d.BlockArc = math.Pi * 0.75
			},
			damage:         100.0,
			attackerAngle:  math.Pi,
			defenderFacing: 0.0,
			wantDamage:     100.0,
			wantNegated:    false,
			tolerance:      0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := NewDefenseComponent("fantasy")
			tt.setup(def)

			gotDamage, gotNegated := ProcessIncomingDamage(def, tt.damage, tt.attackerAngle, tt.defenderFacing)

			if math.Abs(gotDamage-tt.wantDamage) > tt.tolerance {
				t.Errorf("damage = %f, want %f", gotDamage, tt.wantDamage)
			}
			if gotNegated != tt.wantNegated {
				t.Errorf("negated = %v, want %v", gotNegated, tt.wantNegated)
			}
		})
	}
}

func TestGetCounterDamageMultiplier(t *testing.T) {
	def := NewDefenseComponent("fantasy")

	// Base multiplier
	if mul := GetCounterDamageMultiplier(def); mul != 1.0 {
		t.Errorf("base multiplier should be 1.0, got %f", mul)
	}

	// Perfect parry buff
	def.PerfectParryBuff = true
	def.ParryCounterMul = 2.0
	if mul := GetCounterDamageMultiplier(def); mul != 2.0 {
		t.Errorf("perfect parry multiplier should be 2.0, got %f", mul)
	}

	// Backstab window
	def.PerfectParryBuff = false
	def.BackstabWindow = true
	if mul := GetCounterDamageMultiplier(def); mul != 1.5 {
		t.Errorf("backstab multiplier should be 1.5, got %f", mul)
	}

	// Both buffs
	def.PerfectParryBuff = true
	def.BackstabWindow = true
	if mul := GetCounterDamageMultiplier(def); mul != 3.0 {
		t.Errorf("combined multiplier should be 3.0, got %f", mul)
	}
}

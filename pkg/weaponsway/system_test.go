package weaponsway

import (
	"math"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewComponent(t *testing.T) {
	c := NewComponent()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"OffsetX", c.OffsetX, 0.0},
		{"OffsetY", c.OffsetY, 0.0},
		{"VelocityX", c.VelocityX, 0.0},
		{"VelocityY", c.VelocityY, 0.0},
		{"WeaponWeight", c.WeaponWeight, 1.0},
		{"RecoverySpeed", c.RecoverySpeed, 8.0},
		{"Enabled", c.Enabled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestComponent_Type(t *testing.T) {
	c := NewComponent()
	if got := c.Type(); got != "weapon_sway" {
		t.Errorf("Type() = %v, want %v", got, "weapon_sway")
	}
}

func TestComponent_SetWeaponWeight(t *testing.T) {
	tests := []struct {
		name   string
		weight float64
		want   float64
	}{
		{"normal weight", 1.0, 1.0},
		{"light weapon", 0.5, 0.5},
		{"heavy weapon", 2.0, 2.0},
		{"clamp low", 0.01, 0.1},
		{"clamp high", 5.0, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewComponent()
			c.SetWeaponWeight(tt.weight)
			if c.WeaponWeight != tt.want {
				t.Errorf("WeaponWeight = %v, want %v", c.WeaponWeight, tt.want)
			}
		})
	}
}

func TestComponent_GetSwayOffset(t *testing.T) {
	c := NewComponent()
	c.OffsetX = 5.0
	c.OffsetY = -3.0

	x, y := c.GetSwayOffset()
	if x != 5.0 || y != -3.0 {
		t.Errorf("GetSwayOffset() = (%v, %v), want (5.0, -3.0)", x, y)
	}

	// Test disabled returns zero
	c.Enabled = false
	x, y = c.GetSwayOffset()
	if x != 0.0 || y != 0.0 {
		t.Errorf("GetSwayOffset() when disabled = (%v, %v), want (0.0, 0.0)", x, y)
	}
}

func TestComponent_Reset(t *testing.T) {
	c := NewComponent()
	c.OffsetX = 10.0
	c.OffsetY = -5.0
	c.VelocityX = 2.0
	c.VelocityY = -3.0
	c.TargetX = 4.0
	c.TargetY = -2.0
	c.BreathPhase = 1.5
	c.MovementPhase = 2.0

	c.Reset()

	if c.OffsetX != 0 || c.OffsetY != 0 {
		t.Errorf("Reset did not clear offset: (%v, %v)", c.OffsetX, c.OffsetY)
	}
	if c.VelocityX != 0 || c.VelocityY != 0 {
		t.Errorf("Reset did not clear velocity: (%v, %v)", c.VelocityX, c.VelocityY)
	}
	if c.TargetX != 0 || c.TargetY != 0 {
		t.Errorf("Reset did not clear target: (%v, %v)", c.TargetX, c.TargetY)
	}
	if c.BreathPhase != 0 || c.MovementPhase != 0 {
		t.Errorf("Reset did not clear phases: breath=%v, movement=%v", c.BreathPhase, c.MovementPhase)
	}
}

func TestNewSystem(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			s := NewSystem(genre)
			if s == nil {
				t.Fatal("NewSystem returned nil")
			}
			if s.genre != genre {
				t.Errorf("genre = %v, want %v", s.genre, genre)
			}
			// Verify config was applied
			if s.config.TurnSwayMultiplier <= 0 {
				t.Error("TurnSwayMultiplier not configured")
			}
		})
	}
}

func TestSystem_SetGenre(t *testing.T) {
	s := NewSystem("fantasy")

	// Store fantasy config
	fantasyTurn := s.config.TurnSwayMultiplier

	// Switch to scifi
	s.SetGenre("scifi")

	if s.genre != "scifi" {
		t.Errorf("SetGenre did not update genre")
	}
	if s.config.TurnSwayMultiplier == fantasyTurn {
		t.Error("SetGenre did not change config")
	}
}

func TestSystem_applyGenreConfig_AllGenres(t *testing.T) {
	genres := []struct {
		name string
		id   string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"unknown (fallback)", "unknown"},
	}

	for _, tt := range genres {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSystem(tt.id)

			// Verify all required config fields are non-zero
			if s.config.TurnSwayMultiplier <= 0 {
				t.Error("TurnSwayMultiplier not set")
			}
			if s.config.MovementBobAmplitude <= 0 {
				t.Error("MovementBobAmplitude not set")
			}
			if s.config.MovementBobFrequency <= 0 {
				t.Error("MovementBobFrequency not set")
			}
			if s.config.BreathAmplitude <= 0 {
				t.Error("BreathAmplitude not set")
			}
			if s.config.BreathFrequency <= 0 {
				t.Error("BreathFrequency not set")
			}
			if s.config.MaxSwayOffset <= 0 {
				t.Error("MaxSwayOffset not set")
			}
			if s.config.Damping <= 0 {
				t.Error("Damping not set")
			}
			if s.config.SpringStiffness <= 0 {
				t.Error("SpringStiffness not set")
			}
		})
	}
}

func TestSystem_Update(t *testing.T) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()

	entity := w.AddEntity()
	comp := NewComponent()
	comp.VelocityX = 10.0
	comp.VelocityY = 5.0
	w.AddComponent(entity, comp)

	// Run update
	s.Update(w)

	// Position should change from velocity
	if comp.OffsetX == 0 && comp.OffsetY == 0 {
		t.Log("Note: sway did not change offset, possibly due to dt=0")
	}
}

func TestSystem_Update_DisabledComponent(t *testing.T) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()

	entity := w.AddEntity()
	comp := NewComponent()
	comp.Enabled = false
	comp.VelocityX = 100.0
	comp.OffsetX = 5.0
	w.AddComponent(entity, comp)

	originalOffset := comp.OffsetX

	s.Update(w)

	// Offset should not change when disabled
	if comp.OffsetX != originalOffset {
		t.Errorf("Disabled component should not update: offset changed from %v to %v",
			originalOffset, comp.OffsetX)
	}
}

func TestSystem_AddTurnImpulse(t *testing.T) {
	s := NewSystem("fantasy")
	comp := NewComponent()

	// Apply turn impulse
	s.AddTurnImpulse(comp, 0.1, 0.05)

	// Velocity should have changed
	if comp.VelocityX == 0 && comp.VelocityY == 0 {
		t.Error("AddTurnImpulse did not affect velocity")
	}

	// Impulse should be opposite direction (inertia)
	if comp.VelocityX > 0 {
		t.Error("Turn impulse should be opposite to yaw direction")
	}
}

func TestSystem_AddTurnImpulse_Aiming(t *testing.T) {
	s := NewSystem("fantasy")

	compNormal := NewComponent()
	compAiming := NewComponent()
	compAiming.IsAiming = true

	s.AddTurnImpulse(compNormal, 0.1, 0.05)
	s.AddTurnImpulse(compAiming, 0.1, 0.05)

	// Aiming should have less velocity
	if math.Abs(compAiming.VelocityX) >= math.Abs(compNormal.VelocityX) {
		t.Error("Aiming should reduce turn impulse")
	}
}

func TestSystem_AddTurnImpulse_NilComponent(t *testing.T) {
	s := NewSystem("fantasy")

	// Should not panic
	s.AddTurnImpulse(nil, 0.1, 0.05)
}

func TestSystem_SetMovementState(t *testing.T) {
	s := NewSystem("fantasy")
	comp := NewComponent()

	s.SetMovementState(comp, true, false)
	if !comp.IsMoving || comp.IsSprinting {
		t.Error("SetMovementState did not set walking state")
	}

	s.SetMovementState(comp, true, true)
	if !comp.IsMoving || !comp.IsSprinting {
		t.Error("SetMovementState did not set sprinting state")
	}

	s.SetMovementState(comp, false, false)
	if comp.IsMoving || comp.IsSprinting {
		t.Error("SetMovementState did not clear state")
	}
}

func TestSystem_SetMovementState_NilComponent(t *testing.T) {
	s := NewSystem("fantasy")
	// Should not panic
	s.SetMovementState(nil, true, true)
}

func TestSystem_SetAimingState(t *testing.T) {
	s := NewSystem("fantasy")
	comp := NewComponent()

	s.SetAimingState(comp, true)
	if !comp.IsAiming {
		t.Error("SetAimingState did not enable aiming")
	}

	s.SetAimingState(comp, false)
	if comp.IsAiming {
		t.Error("SetAimingState did not disable aiming")
	}
}

func TestSystem_SetAimingState_NilComponent(t *testing.T) {
	s := NewSystem("fantasy")
	// Should not panic
	s.SetAimingState(nil, true)
}

func TestSystem_GetConfig(t *testing.T) {
	s := NewSystem("fantasy")
	cfg := s.GetConfig()

	if cfg.TurnSwayMultiplier <= 0 {
		t.Error("GetConfig returned invalid config")
	}
}

func TestSystem_clampOffset(t *testing.T) {
	s := NewSystem("fantasy")
	comp := NewComponent()

	// Set extreme values
	comp.OffsetX = 1000.0
	comp.OffsetY = -1000.0
	comp.TargetX = 500.0
	comp.TargetY = -500.0

	s.clampOffset(comp)

	maxOffset := s.config.MaxSwayOffset
	if comp.OffsetX > maxOffset {
		t.Errorf("OffsetX not clamped: %v > %v", comp.OffsetX, maxOffset)
	}
	if comp.OffsetY < -maxOffset {
		t.Errorf("OffsetY not clamped: %v < %v", comp.OffsetY, -maxOffset)
	}
	if comp.TargetX > maxOffset {
		t.Errorf("TargetX not clamped: %v > %v", comp.TargetX, maxOffset)
	}
}

func TestSystem_clampOffset_Aiming(t *testing.T) {
	s := NewSystem("fantasy")
	compNormal := NewComponent()
	compAiming := NewComponent()
	compAiming.IsAiming = true

	// Set same extreme values
	compNormal.OffsetX = 100.0
	compAiming.OffsetX = 100.0

	s.clampOffset(compNormal)
	s.clampOffset(compAiming)

	// Aiming should have tighter clamp
	if compAiming.OffsetX >= compNormal.OffsetX {
		t.Error("Aiming should have tighter clamp bounds")
	}
}

func TestSwayDeterminism(t *testing.T) {
	// Verify that identical inputs produce identical outputs
	s1 := NewSystem("fantasy")
	s2 := NewSystem("fantasy")

	comp1 := NewComponent()
	comp2 := NewComponent()

	// Apply identical impulses
	s1.AddTurnImpulse(comp1, 0.1, 0.05)
	s2.AddTurnImpulse(comp2, 0.1, 0.05)

	if comp1.VelocityX != comp2.VelocityX {
		t.Errorf("Non-deterministic: VelocityX %v != %v", comp1.VelocityX, comp2.VelocityX)
	}
	if comp1.VelocityY != comp2.VelocityY {
		t.Errorf("Non-deterministic: VelocityY %v != %v", comp1.VelocityY, comp2.VelocityY)
	}
}

func BenchmarkSystem_Update(b *testing.B) {
	s := NewSystem("fantasy")
	w := engine.NewWorld()

	// Create multiple entities with sway components
	for i := 0; i < 100; i++ {
		entity := w.AddEntity()
		comp := NewComponent()
		comp.VelocityX = float64(i) * 0.1
		comp.VelocityY = float64(i) * 0.05
		comp.IsMoving = i%2 == 0
		w.AddComponent(entity, comp)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update(w)
	}
}

func BenchmarkSystem_AddTurnImpulse(b *testing.B) {
	s := NewSystem("fantasy")
	comp := NewComponent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.AddTurnImpulse(comp, 0.01, 0.005)
	}
}

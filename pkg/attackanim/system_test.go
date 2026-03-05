package attackanim

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestComponent_Type(t *testing.T) {
	comp := &Component{}
	if comp.Type() != "AttackAnimation" {
		t.Errorf("expected Type() = 'AttackAnimation', got '%s'", comp.Type())
	}
}

func TestComponent_IsAnimating(t *testing.T) {
	tests := []struct {
		name     string
		state    AttackState
		expected bool
	}{
		{"idle", StateIdle, false},
		{"windup", StateWindup, true},
		{"strike", StateStrike, true},
		{"recovery", StateRecovery, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &Component{AttackState: tt.state}
			if comp.IsAnimating() != tt.expected {
				t.Errorf("expected IsAnimating() = %v, got %v", tt.expected, comp.IsAnimating())
			}
		})
	}
}

func TestComponent_AnimationProgress(t *testing.T) {
	tests := []struct {
		name      string
		state     AttackState
		stateTime float64
		duration  float64
		expected  float64
	}{
		{"windup half", StateWindup, 0.1, 0.2, 0.5},
		{"strike complete", StateStrike, 0.15, 0.15, 1.0},
		{"recovery start", StateRecovery, 0.0, 0.25, 0.0},
		{"idle", StateIdle, 1.0, 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &Component{
				AttackState: tt.state,
				StateTime:   tt.stateTime,
			}

			switch tt.state {
			case StateWindup:
				comp.WindupTime = tt.duration
			case StateStrike:
				comp.StrikeTime = tt.duration
			case StateRecovery:
				comp.RecoveryTime = tt.duration
			}

			progress := comp.AnimationProgress()
			if progress != tt.expected {
				t.Errorf("expected progress %v, got %v", tt.expected, progress)
			}
		})
	}
}

func TestSystem_StartAttack(t *testing.T) {
	tests := []struct {
		name          string
		animationType string
		intensity     float64
	}{
		{"melee slash", "melee_slash", 1.0},
		{"overhead smash", "overhead_smash", 1.2},
		{"lunge", "lunge", 0.8},
		{"ranged charge", "ranged_charge", 1.0},
		{"spin attack", "spin_attack", 1.0},
		{"quick jab", "quick_jab", 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewSystem("fantasy")
			world := engine.NewWorld()
			entity := world.AddEntity()

			system.StartAttack(world, entity, tt.animationType, 1.0, 0.0, tt.intensity)

			comp := system.getComponent(world, entity)
			if comp == nil {
				t.Fatal("component not created")
			}

			if comp.AnimationType != tt.animationType {
				t.Errorf("expected animation type %s, got %s", tt.animationType, comp.AnimationType)
			}

			if comp.AttackState != StateWindup {
				t.Errorf("expected state Windup, got %v", comp.AttackState)
			}

			if comp.StateTime != 0 {
				t.Errorf("expected StateTime 0, got %v", comp.StateTime)
			}

			if comp.WindupTime == 0 || comp.StrikeTime == 0 || comp.RecoveryTime == 0 {
				t.Error("animation timings not set")
			}
		})
	}
}

func TestSystem_Update_StateTransitions(t *testing.T) {
	system := NewSystem("fantasy")
	world := engine.NewWorld()
	entity := world.AddEntity()

	system.StartAttack(world, entity, "melee_slash", 1.0, 0.0, 1.0)
	comp := system.getComponent(world, entity)

	// Manually advance state time to trigger transitions
	comp.StateTime = comp.WindupTime
	system.Update(world)
	if comp.AttackState != StateStrike {
		t.Errorf("expected transition to Strike, got %v", comp.AttackState)
	}

	comp.StateTime = comp.StrikeTime
	system.Update(world)
	if comp.AttackState != StateRecovery {
		t.Errorf("expected transition to Recovery, got %v", comp.AttackState)
	}

	comp.StateTime = comp.RecoveryTime
	system.Update(world)
	if comp.AttackState != StateIdle {
		t.Errorf("expected transition to Idle, got %v", comp.AttackState)
	}
}

func TestSystem_StopAttack(t *testing.T) {
	system := NewSystem("fantasy")
	world := engine.NewWorld()
	entity := world.AddEntity()

	system.StartAttack(world, entity, "melee_slash", 1.0, 0.0, 1.0)
	comp := system.getComponent(world, entity)

	if comp.AttackState == StateIdle {
		t.Fatal("attack should be active")
	}

	system.StopAttack(world, entity)

	if comp.AttackState != StateIdle {
		t.Errorf("expected state Idle after stop, got %v", comp.AttackState)
	}

	if comp.OffsetX != 0 || comp.OffsetY != 0 || comp.RotationAngle != 0 {
		t.Error("visual parameters not reset after stop")
	}
}

func TestSystem_GetAnimationParams(t *testing.T) {
	system := NewSystem("fantasy")
	world := engine.NewWorld()
	entity := world.AddEntity()

	// No animation
	_, _, _, _, animating := system.GetAnimationParams(world, entity)
	if animating {
		t.Error("expected no animation initially")
	}

	// Start animation
	system.StartAttack(world, entity, "melee_slash", 1.0, 0.0, 1.0)
	_, _, _, _, animating = system.GetAnimationParams(world, entity)
	if !animating {
		t.Error("expected animation to be active")
	}
}

func TestSystem_VisualUpdates(t *testing.T) {
	tests := []struct {
		name          string
		animationType string
		state         AttackState
	}{
		{"slash windup", "melee_slash", StateWindup},
		{"smash strike", "overhead_smash", StateStrike},
		{"lunge recovery", "lunge", StateRecovery},
		{"spin windup", "spin_attack", StateWindup},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system := NewSystem("fantasy")
			world := engine.NewWorld()
			entity := world.AddEntity()

			system.StartAttack(world, entity, tt.animationType, 1.0, 0.0, 1.0)
			comp := system.getComponent(world, entity)

			// Manually set to desired state for testing
			comp.AttackState = tt.state
			comp.StateTime = 0

			// Update to generate visuals
			system.Update(world)

			// Just verify update doesn't crash and parameters are set
			offsetX, offsetY, rotation, squash, _ := system.GetAnimationParams(world, entity)
			_ = offsetX
			_ = offsetY
			_ = rotation

			if squash <= 0 {
				t.Errorf("invalid squash value: %v", squash)
			}
		})
	}
}

func BenchmarkSystem_Update(b *testing.B) {
	system := NewSystem("fantasy")
	world := engine.NewWorld()

	for i := 0; i < 100; i++ {
		entity := world.AddEntity()
		if i%2 == 0 {
			system.StartAttack(world, entity, "melee_slash", 1.0, 0.0, 1.0)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(world)
	}
}

func BenchmarkSystem_GetAnimationParams(b *testing.B) {
	system := NewSystem("fantasy")
	world := engine.NewWorld()
	entity := world.AddEntity()
	system.StartAttack(world, entity, "melee_slash", 1.0, 0.0, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.GetAnimationParams(world, entity)
	}
}

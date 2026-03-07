package motion

import (
	"math"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestComponent_Type(t *testing.T) {
	comp := &Component{}
	if got := comp.Type(); got != "motion" {
		t.Errorf("Type() = %v, want motion", got)
	}
}

func TestSystem_UpdateBreathing(t *testing.T) {
	sys := NewSystem()

	tests := []struct {
		name         string
		frequency    float64
		deltaTime    float64
		iterations   int
		wantPhaseMin float64
		wantPhaseMax float64
	}{
		{
			name:         "breathing advances phase",
			frequency:    1.0,
			deltaTime:    0.1,
			iterations:   10,
			wantPhaseMin: 0.5,
			wantPhaseMax: 2 * math.Pi,
		},
		{
			name:         "zero frequency no change",
			frequency:    0.0,
			deltaTime:    0.1,
			iterations:   10,
			wantPhaseMin: 0.0,
			wantPhaseMax: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			motion := &Component{
				BreathFrequency: tt.frequency,
				BreathPhase:     0,
			}

			for i := 0; i < tt.iterations; i++ {
				sys.updateBreathing(motion, tt.deltaTime)
			}

			if motion.BreathPhase < tt.wantPhaseMin || motion.BreathPhase > tt.wantPhaseMax {
				t.Errorf("BreathPhase = %v, want in range [%v, %v]",
					motion.BreathPhase, tt.wantPhaseMin, tt.wantPhaseMax)
			}
		})
	}
}

func TestSystem_ApplyEasing(t *testing.T) {
	sys := NewSystem()

	tests := []struct {
		name        string
		initialVelX float64
		targetVelX  float64
		mass        float64
		easeRate    float64
		iterations  int
		maxDiff     float64
	}{
		{
			name:        "converges toward target velocity",
			initialVelX: 0.0,
			targetVelX:  10.0,
			mass:        1.0,
			easeRate:    5.0,
			iterations:  100,
			maxDiff:     5.0, // Should get at least halfway
		},
		{
			name:        "heavier mass slower convergence",
			initialVelX: 0.0,
			targetVelX:  10.0,
			mass:        4.0,
			easeRate:    5.0,
			iterations:  50,
			maxDiff:     9.0, // Should move slowly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			motion := &Component{
				EasedVelocityX: tt.initialVelX,
				Mass:           tt.mass,
				EaseRate:       tt.easeRate,
			}
			vel := &engine.Velocity{DX: tt.targetVelX}

			for i := 0; i < tt.iterations; i++ {
				// Set target each iteration (simulating system behavior)
				motion.TargetVelocityX = tt.targetVelX
				sys.applyEasing(motion, vel, 1.0/60.0)
			}

			diff := math.Abs(motion.EasedVelocityX - tt.targetVelX)

			if diff > tt.maxDiff {
				t.Errorf("Diff = %v, expected <= %v", diff, tt.maxDiff)
			}

			// Should have moved toward target
			if math.Abs(motion.EasedVelocityX-tt.initialVelX) < 0.1 {
				t.Error("Easing did not move velocity at all")
			}
		})
	}
}

func TestSystem_EaseInOut(t *testing.T) {
	sys := NewSystem()

	tests := []struct {
		name       string
		current    float64
		target     float64
		speed      float64
		wantCloser bool
	}{
		{
			name:       "moves toward target",
			current:    0.0,
			target:     10.0,
			speed:      0.1,
			wantCloser: true,
		},
		{
			name:       "already at target stays",
			current:    5.0,
			target:     5.0,
			speed:      0.1,
			wantCloser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.easeInOut(tt.current, tt.target, tt.speed)

			initialDiff := math.Abs(tt.target - tt.current)
			resultDiff := math.Abs(tt.target - result)

			if tt.wantCloser {
				if resultDiff >= initialDiff {
					t.Errorf("easeInOut did not move closer: initial diff %v, result diff %v",
						initialDiff, resultDiff)
				}
			} else {
				if result != tt.target {
					t.Errorf("easeInOut() = %v, want %v", result, tt.target)
				}
			}
		})
	}
}

func TestSystem_UpdateSquashStretch(t *testing.T) {
	sys := NewSystem()
	motion := &Component{
		SquashX:    0.7,
		SquashY:    1.3,
		ImpactTime: 0.0,
	}

	// Run for several frames
	for i := 0; i < 20; i++ {
		sys.updateSquashStretch(motion, 1.0/60.0)
	}

	// Should recover toward 1.0
	if math.Abs(motion.SquashX-1.0) > 0.1 {
		t.Errorf("SquashX = %v, expected close to 1.0", motion.SquashX)
	}
	if math.Abs(motion.SquashY-1.0) > 0.1 {
		t.Errorf("SquashY = %v, expected close to 1.0", motion.SquashY)
	}

	// ImpactTime should increase
	if motion.ImpactTime < 0.2 {
		t.Errorf("ImpactTime = %v, expected >= 0.2", motion.ImpactTime)
	}
}

func TestSystem_UpdateSecondaryMotion(t *testing.T) {
	sys := NewSystem()

	motion := &Component{
		TrailLength:    3,
		TrailStiffness: 0.5,
	}

	pos := &engine.Position{X: 0, Y: 0}

	// Initialize trail
	sys.updateSecondaryMotion(motion, pos, 1.0/60.0)

	if len(motion.TrailOffsetX) != 3 {
		t.Fatalf("Trail not initialized, got length %d", len(motion.TrailOffsetX))
	}

	// Move entity
	pos.X = 10
	pos.Y = 5

	for i := 0; i < 10; i++ {
		sys.updateSecondaryMotion(motion, pos, 1.0/60.0)
	}

	// Trail should lag behind
	if motion.TrailOffsetX[motion.TrailLength-1] >= pos.X {
		t.Errorf("Trail end should lag behind entity position")
	}
}

func TestSystem_TriggerImpact(t *testing.T) {
	sys := NewSystem()

	tests := []struct {
		name              string
		velocityMagnitude float64
		wantSquashY       string
	}{
		{
			name:              "high velocity strong squash",
			velocityMagnitude: 100.0,
			wantSquashY:       "less_than_one",
		},
		{
			name:              "low velocity mild squash",
			velocityMagnitude: 20.0,
			wantSquashY:       "close_to_one",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			motion := &Component{
				SquashX:    1.0,
				SquashY:    1.0,
				ImpactTime: 0.2,
			}

			sys.TriggerImpact(motion, tt.velocityMagnitude)

			switch tt.wantSquashY {
			case "less_than_one":
				if motion.SquashY >= 1.0 {
					t.Errorf("SquashY = %v, expected < 1.0 for high impact", motion.SquashY)
				}
			case "close_to_one":
				if motion.SquashY < 0.9 {
					t.Errorf("SquashY = %v, expected close to 1.0 for low impact", motion.SquashY)
				}
			}

			// SquashX should expand to preserve volume
			if motion.SquashX <= 1.0 {
				t.Errorf("SquashX = %v, expected > 1.0 to preserve volume", motion.SquashX)
			}
		})
	}
}

func TestSystem_GetBreathOffset(t *testing.T) {
	sys := NewSystem()

	motion := &Component{
		BreathPhase:     0,
		BreathAmplitude: 2.0,
	}

	// At phase 0, sin(0) = 0
	offset := sys.GetBreathOffset(motion)
	if math.Abs(offset) > 0.01 {
		t.Errorf("GetBreathOffset at phase 0 = %v, want ~0", offset)
	}

	// At phase π/2, sin(π/2) = 1, offset should be amplitude
	motion.BreathPhase = math.Pi / 2
	offset = sys.GetBreathOffset(motion)
	if math.Abs(offset-2.0) > 0.01 {
		t.Errorf("GetBreathOffset at phase π/2 = %v, want ~2.0", offset)
	}
}

func TestSystem_GetSquashStretch(t *testing.T) {
	sys := NewSystem()
	motion := &Component{
		SquashX: 1.2,
		SquashY: 0.8,
	}

	x, y := sys.GetSquashStretch(motion)
	if x != 1.2 {
		t.Errorf("GetSquashStretch X = %v, want 1.2", x)
	}
	if y != 0.8 {
		t.Errorf("GetSquashStretch Y = %v, want 0.8", y)
	}
}

func TestSystem_GetTrailSegment(t *testing.T) {
	sys := NewSystem()
	motion := &Component{
		TrailOffsetX: []float64{10, 20, 30},
		TrailOffsetY: []float64{15, 25, 35},
	}

	tests := []struct {
		name      string
		index     int
		wantX     float64
		wantY     float64
		wantValid bool
	}{
		{
			name:      "valid index 0",
			index:     0,
			wantX:     10,
			wantY:     15,
			wantValid: true,
		},
		{
			name:      "valid index 2",
			index:     2,
			wantX:     30,
			wantY:     35,
			wantValid: true,
		},
		{
			name:      "invalid negative index",
			index:     -1,
			wantValid: false,
		},
		{
			name:      "invalid out of bounds",
			index:     3,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, valid := sys.GetTrailSegment(motion, tt.index)

			if valid != tt.wantValid {
				t.Errorf("GetTrailSegment valid = %v, want %v", valid, tt.wantValid)
			}

			if tt.wantValid {
				if x != tt.wantX {
					t.Errorf("GetTrailSegment X = %v, want %v", x, tt.wantX)
				}
				if y != tt.wantY {
					t.Errorf("GetTrailSegment Y = %v, want %v", y, tt.wantY)
				}
			}
		})
	}
}

func TestInitializeMotion(t *testing.T) {
	tests := []struct {
		name     string
		mass     float64
		hasTrail bool
	}{
		{
			name:     "light entity with trail",
			mass:     1.0,
			hasTrail: true,
		},
		{
			name:     "heavy entity no trail",
			mass:     10.0,
			hasTrail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := InitializeMotion(tt.mass, tt.hasTrail)

			if comp.Mass != tt.mass {
				t.Errorf("Mass = %v, want %v", comp.Mass, tt.mass)
			}

			if tt.hasTrail {
				if comp.TrailLength <= 0 {
					t.Errorf("TrailLength = %v, want > 0", comp.TrailLength)
				}
			} else {
				if comp.TrailLength != 0 {
					t.Errorf("TrailLength = %v, want 0", comp.TrailLength)
				}
			}

			// Check defaults
			if comp.SquashX != 1.0 || comp.SquashY != 1.0 {
				t.Errorf("Squash not initialized to 1.0")
			}
			if comp.EaseRate <= 0 {
				t.Errorf("EaseRate = %v, want > 0", comp.EaseRate)
			}
			if comp.BreathFrequency <= 0 {
				t.Errorf("BreathFrequency = %v, want > 0", comp.BreathFrequency)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		min   float64
		max   float64
		want  float64
	}{
		{
			name:  "below min",
			value: 5.0,
			min:   10.0,
			max:   20.0,
			want:  10.0,
		},
		{
			name:  "above max",
			value: 25.0,
			min:   10.0,
			max:   20.0,
			want:  20.0,
		},
		{
			name:  "within range",
			value: 15.0,
			min:   10.0,
			max:   20.0,
			want:  15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.value, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%v, %v, %v) = %v, want %v",
					tt.value, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestSystem_Update_Integration(t *testing.T) {
	w := engine.NewWorld()
	sys := NewSystem()

	// Create entity with position, velocity, and motion
	e := w.AddEntity()
	w.AddComponent(e, &engine.Position{X: 0, Y: 0})
	w.AddComponent(e, &engine.Velocity{DX: 5.0, DY: 3.0})
	w.AddComponent(e, InitializeMotion(2.0, true))

	// Run several updates
	for i := 0; i < 30; i++ {
		sys.Update(w)
	}

	// Verify motion component was updated
	motionComp, found := w.GetComponent(e, reflect.TypeOf(&Component{}))
	if !found {
		t.Fatal("Motion component not found after update")
	}

	motion := motionComp.(*Component)

	// Breath phase should have advanced
	if motion.BreathPhase == 0 {
		t.Error("BreathPhase not updated")
	}

	// Trail should be initialized
	if len(motion.TrailOffsetX) == 0 {
		t.Error("Trail not initialized")
	}

	// Impact time should increase
	if motion.ImpactTime <= 0 {
		t.Error("ImpactTime not updated")
	}
}

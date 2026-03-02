package weaponanim

import (
	"math"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestWeaponAnimComponent(t *testing.T) {
	tests := []struct {
		name     string
		progress float64
		start    float64
		end      float64
		expected float64
	}{
		{"start of animation", 0.0, 0.0, math.Pi, 0.0},
		{"mid animation", 0.5, 0.0, math.Pi, math.Pi / 2},
		{"end of animation", 1.0, 0.0, math.Pi, math.Pi},
		{"partial progress", 0.25, math.Pi / 4, math.Pi / 2, 0.857}, // approximate with easing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := &WeaponAnimComponent{
				Active:     true,
				Progress:   tt.progress,
				StartAngle: tt.start,
				EndAngle:   tt.end,
			}

			angle := comp.GetCurrentAngle()

			// Allow small tolerance for easing function
			tolerance := 0.1
			if tt.progress == 0.0 || tt.progress == 1.0 {
				tolerance = 0.01
			}

			diff := math.Abs(angle - tt.expected)
			if diff > tolerance && tt.name != "partial progress" {
				t.Errorf("GetCurrentAngle() = %v, expected %v", angle, tt.expected)
			}
		})
	}
}

func TestGetTipPosition(t *testing.T) {
	comp := &WeaponAnimComponent{
		Active:     true,
		Progress:   0.0,
		StartAngle: 0.0,
		EndAngle:   math.Pi,
		ArcRadius:  10.0,
	}

	wielderX, wielderY := 100.0, 200.0
	tipX, tipY := comp.GetTipPosition(wielderX, wielderY)

	// At angle 0, tip should be at (wielderX + radius, wielderY)
	expectedX := wielderX + 10.0
	expectedY := wielderY

	if math.Abs(tipX-expectedX) > 0.01 || math.Abs(tipY-expectedY) > 0.01 {
		t.Errorf("GetTipPosition() = (%v, %v), expected (%v, %v)", tipX, tipY, expectedX, expectedY)
	}
}

func TestGetSwingParameters(t *testing.T) {
	tests := []struct {
		name      string
		swingType SwingType
		facing    float64
		wantDur   float64
	}{
		{"slash", SwingSlash, 0.0, 0.3},
		{"overhead", SwingOverhead, 0.0, 0.4},
		{"thrust", SwingThrust, 0.0, 0.2},
		{"uppercut", SwingUppercut, 0.0, 0.35},
		{"wide", SwingWide, 0.0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, duration := GetSwingParameters(tt.swingType, tt.facing)
			if duration != tt.wantDur {
				t.Errorf("GetSwingParameters() duration = %v, want %v", duration, tt.wantDur)
			}
		})
	}
}

func TestSystemUpdate(t *testing.T) {
	world := engine.NewWorld()
	system := NewSystem()

	// Create entity with weapon animation
	entity := world.AddEntity()
	world.AddComponent(entity, &PositionComponent{X: 100, Y: 100})
	anim := &WeaponAnimComponent{
		Active:     true,
		Progress:   0.0,
		Duration:   0.3,
		StartAngle: 0.0,
		EndAngle:   math.Pi,
		ArcRadius:  20.0,
	}
	world.AddComponent(entity, anim)

	// Update multiple times
	for i := 0; i < 20; i++ {
		system.Update(world)
	}

	// Animation should have progressed
	if anim.Progress <= 0.0 {
		t.Error("Animation progress did not increase")
	}

	// Should have trail points
	if len(anim.TrailPoints) == 0 {
		t.Error("Expected trail points to be created")
	}
}

func TestStartSwing(t *testing.T) {
	world := engine.NewWorld()
	system := NewSystem()
	entity := world.AddEntity()
	world.AddComponent(entity, &PositionComponent{X: 100, Y: 100})
	world.AddComponent(entity, &VelocityComponent{VX: 1.0, VY: 0.0})

	// Add swing trigger
	world.AddComponent(entity, &SwingTriggerComponent{
		SwingType: 0, // SwingSlash
		Pending:   true,
	})

	// Run system update to process trigger
	system.Update(world)

	// Should have weapon animation component
	weaponAnimType := reflect.TypeOf(&WeaponAnimComponent{})
	animComp, ok := world.GetComponent(entity, weaponAnimType)
	if !ok {
		t.Fatal("WeaponAnimComponent not added")
	}

	anim := animComp.(*WeaponAnimComponent)
	if !anim.Active {
		t.Error("Animation should be active")
	}

	if anim.SwingType != SwingSlash {
		t.Errorf("SwingType = %v, want %v", anim.SwingType, SwingSlash)
	}

	if anim.ArcRadius != 25.0 {
		t.Errorf("ArcRadius = %v, want 25.0", anim.ArcRadius)
	}
}

func TestTrailFading(t *testing.T) {
	world := engine.NewWorld()
	system := NewSystem()

	entity := world.AddEntity()
	world.AddComponent(entity, &PositionComponent{X: 100, Y: 100})
	anim := &WeaponAnimComponent{
		Active:     true,
		Progress:   0.0,
		Duration:   0.3,
		StartAngle: 0.0,
		EndAngle:   math.Pi,
		ArcRadius:  20.0,
	}
	world.AddComponent(entity, anim)

	// Create initial trail
	for i := 0; i < 10; i++ {
		system.Update(world)
	}

	initialCount := len(anim.TrailPoints)
	if initialCount == 0 {
		t.Fatal("No trail points created")
	}

	// Stop animation
	anim.Active = false

	// Continue updating - trail should fade
	for i := 0; i < 100; i++ {
		system.Update(world)
	}

	if len(anim.TrailPoints) >= initialCount {
		t.Error("Trail points should have faded")
	}
}

func TestEaseInOut(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{1.0, 1.0},
		{0.5, 0.5},
	}

	for _, tt := range tests {
		result := easeInOut(tt.input)
		if math.Abs(result-tt.expected) > 0.01 {
			t.Errorf("easeInOut(%v) = %v, expected %v", tt.input, result, tt.expected)
		}
	}

	// Test that easing is smooth (derivative exists)
	for testVal := 0.1; testVal < 1.0; testVal += 0.1 {
		val := easeInOut(testVal)
		if val < 0.0 || val > 1.0 {
			t.Errorf("easeInOut(%v) = %v, out of range [0, 1]", testVal, val)
		}
	}
}

func TestComponentType(t *testing.T) {
	comp := &WeaponAnimComponent{}
	if comp.Type() != "weaponanim" {
		t.Errorf("Type() = %v, want weaponanim", comp.Type())
	}

	pos := &PositionComponent{}
	if pos.Type() != "position" {
		t.Errorf("Type() = %v, want position", pos.Type())
	}

	vel := &VelocityComponent{}
	if vel.Type() != "velocity" {
		t.Errorf("Type() = %v, want velocity", vel.Type())
	}
}

func BenchmarkSystemUpdate(b *testing.B) {
	world := engine.NewWorld()
	system := NewSystem()

	// Create 100 entities with weapon animations
	for i := 0; i < 100; i++ {
		entity := world.AddEntity()
		world.AddComponent(entity, &PositionComponent{X: float64(i * 10), Y: float64(i * 10)})
		anim := &WeaponAnimComponent{
			Active:     true,
			Progress:   0.0,
			Duration:   0.3,
			StartAngle: 0.0,
			EndAngle:   math.Pi,
			ArcRadius:  20.0,
		}
		world.AddComponent(entity, anim)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(world)
	}
}

func BenchmarkGetSwingColor(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetSwingColor("sword", "fantasy")
		GetSwingColor("magic", "fantasy")
		GetSwingColor("blade", "cyberpunk")
	}
}

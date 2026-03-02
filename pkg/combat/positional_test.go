// Package combat - Positional advantage tests
package combat

import (
	"math"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestPositionalComponent(t *testing.T) {
	comp := &PositionalComponent{
		FacingAngle: 0,
		Height:      1.0,
	}

	if comp.Type() != "PositionalComponent" {
		t.Errorf("wrong component type: %s", comp.Type())
	}

	// Test facing update
	comp.SetFacingFromDirection(1.0, 0.0)
	if math.Abs(comp.FacingAngle-0) > 0.01 {
		t.Errorf("facing should be 0 radians (right), got %f", comp.FacingAngle)
	}

	comp.SetFacingFromDirection(0.0, 1.0)
	if math.Abs(comp.FacingAngle-math.Pi/2) > 0.01 {
		t.Errorf("facing should be π/2 radians (down), got %f", comp.FacingAngle)
	}

	comp.SetFacingFromDirection(-1.0, 0.0)
	if math.Abs(comp.FacingAngle-math.Pi) > 0.01 && math.Abs(comp.FacingAngle+math.Pi) > 0.01 {
		t.Errorf("facing should be ±π radians (left), got %f", comp.FacingAngle)
	}

	// Test facing vector
	comp.FacingAngle = 0
	fx, fy := comp.GetFacingVector()
	if math.Abs(fx-1.0) > 0.01 || math.Abs(fy-0.0) > 0.01 {
		t.Errorf("facing vector should be (1, 0), got (%f, %f)", fx, fy)
	}
}

func TestGetPositionalConfig(t *testing.T) {
	tests := []struct {
		genre    string
		minMulti float64
	}{
		{"fantasy", 1.5},
		{"scifi", 1.4},
		{"horror", 1.6},
		{"cyberpunk", 1.5},
		{"unknown", 1.5}, // defaults to fantasy
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			cfg := GetPositionalConfig(tt.genre)
			if cfg.FlankMultiplier < tt.minMulti {
				t.Errorf("flank multiplier too low: %f", cfg.FlankMultiplier)
			}
			if cfg.BackstabMultiplier <= cfg.FlankMultiplier {
				t.Errorf("backstab should be stronger than flank")
			}
			if cfg.BackstabAngle <= 0 || cfg.BackstabAngle > math.Pi {
				t.Errorf("backstab angle out of range: %f", cfg.BackstabAngle)
			}
		})
	}
}

func TestCalculatePositionalAdvantage(t *testing.T) {
	cfg := GetPositionalConfig("fantasy")

	tests := []struct {
		name           string
		attackerX      float64
		attackerY      float64
		targetX        float64
		targetY        float64
		targetFacing   float64
		expectedAdv    PositionalAdvantage
		expectedMinMul float64
	}{
		{
			name:           "frontal attack",
			attackerX:      10,
			attackerY:      0,
			targetX:        0,
			targetY:        0,
			targetFacing:   0, // facing right, attacker on right
			expectedAdv:    AdvantageFrontal,
			expectedMinMul: 1.0,
		},
		{
			name:           "backstab attack",
			attackerX:      -10,
			attackerY:      0,
			targetX:        0,
			targetY:        0,
			targetFacing:   0, // facing right, attacker behind (left)
			expectedAdv:    AdvantageBackstab,
			expectedMinMul: 2.0,
		},
		{
			name:           "flank from above",
			attackerX:      0,
			attackerY:      10,
			targetX:        0,
			targetY:        0,
			targetFacing:   0, // facing right, attacker on side (below)
			expectedAdv:    AdvantageFlank,
			expectedMinMul: 1.5,
		},
		{
			name:           "flank from below",
			attackerX:      0,
			attackerY:      -10,
			targetX:        0,
			targetY:        0,
			targetFacing:   0, // facing right, attacker on side (above)
			expectedAdv:    AdvantageFlank,
			expectedMinMul: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetPos := &PositionalComponent{
				FacingAngle: tt.targetFacing,
				Height:      0,
			}

			adv, mul := CalculatePositionalAdvantage(
				tt.attackerX, tt.attackerY,
				tt.targetX, tt.targetY,
				nil, targetPos,
				cfg,
			)

			if adv != tt.expectedAdv {
				t.Errorf("expected advantage %v, got %v", tt.expectedAdv, adv)
			}

			if mul < tt.expectedMinMul-0.01 {
				t.Errorf("expected multiplier >= %f, got %f", tt.expectedMinMul, mul)
			}
		})
	}
}

func TestElevationAdvantage(t *testing.T) {
	cfg := GetPositionalConfig("fantasy")

	attackerPos := &PositionalComponent{
		FacingAngle: 0,
		Height:      5.0, // elevated
	}

	targetPos := &PositionalComponent{
		FacingAngle: 0,
		Height:      0.0, // ground level
	}

	// Frontal attack with elevation
	adv, mul := CalculatePositionalAdvantage(
		10, 0, 0, 0,
		attackerPos, targetPos,
		cfg,
	)

	if mul < 1.3-0.01 {
		t.Errorf("elevation should provide bonus, got multiplier %f", mul)
	}

	// Should still show frontal/elevation as primary advantage
	if adv != AdvantageElevation && adv != AdvantageFrontal {
		t.Errorf("expected elevation or frontal advantage, got %v", adv)
	}
}

func TestBackstabWithElevation(t *testing.T) {
	cfg := GetPositionalConfig("fantasy")

	attackerPos := &PositionalComponent{
		FacingAngle: math.Pi, // facing left
		Height:      5.0,
	}

	targetPos := &PositionalComponent{
		FacingAngle: 0, // facing right
		Height:      0.0,
	}

	// Attack from behind with elevation
	adv, mul := CalculatePositionalAdvantage(
		-10, 0, 0, 0,
		attackerPos, targetPos,
		cfg,
	)

	if adv != AdvantageBackstab {
		t.Errorf("should be backstab, got %v", adv)
	}

	// Should get both backstab AND elevation multipliers
	expectedMin := cfg.BackstabMultiplier * cfg.ElevationMultiplier
	if mul < expectedMin-0.01 {
		t.Errorf("expected multiplier >= %f (backstab+elevation), got %f", expectedMin, mul)
	}
}

func TestPositionalSystem(t *testing.T) {
	w := engine.NewWorld()
	sys := NewPositionalSystem("fantasy")

	// Create entities
	attacker := w.AddEntity()
	target := w.AddEntity()

	// Add positions
	w.AddComponent(attacker, &engine.Position{X: 10, Y: 0})
	w.AddComponent(target, &engine.Position{X: 0, Y: 0})

	// Add positional components
	sys.AddPositionalComponent(w, attacker, 0, 0)
	sys.AddPositionalComponent(w, target, 0, 0)

	// Test frontal attack
	mul := sys.CalculateDamageMultiplier(w, attacker, target)
	if mul < 1.0-0.01 || mul > 1.0+0.01 {
		t.Errorf("frontal attack should be 1.0x, got %f", mul)
	}

	// Reposition for backstab
	if comp, ok := w.GetComponent(attacker, reflect.TypeOf(&engine.Position{})); ok {
		pos := comp.(*engine.Position)
		pos.X = -10
	}

	mul = sys.CalculateDamageMultiplier(w, attacker, target)
	if mul < 2.0-0.01 {
		t.Errorf("backstab should be >= 2.0x, got %f", mul)
	}
}

func TestUpdateFacingFromVelocity(t *testing.T) {
	w := engine.NewWorld()
	sys := NewPositionalSystem("fantasy")

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Position{X: 0, Y: 0})
	w.AddComponent(entity, &engine.Velocity{DX: 1.0, DY: 0.0})

	sys.AddPositionalComponent(w, entity, math.Pi, 0) // Start facing left

	// Update should change facing to match velocity (right)
	sys.Update(w)

	if comp, ok := w.GetComponent(entity, reflect.TypeOf(&PositionalComponent{})); ok {
		pc := comp.(*PositionalComponent)
		if math.Abs(pc.FacingAngle-0) > 0.1 {
			t.Errorf("facing should update to 0 (right), got %f", pc.FacingAngle)
		}
	} else {
		t.Error("positional component not found")
	}
}

func TestIsBackstabAngle(t *testing.T) {
	sys := NewPositionalSystem("fantasy")

	// Target facing right (0)
	targetFacing := 0.0

	tests := []struct {
		name       string
		attackerX  float64
		attackerY  float64
		isBackstab bool
	}{
		{"from front", 10, 0, false},
		{"from behind", -10, 0, true},
		{"from above", 0, 10, false},
		{"from below", 0, -10, false},
		{"diagonal behind", -10, 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.IsBackstabAngle(tt.attackerX, tt.attackerY, 0, 0, targetFacing)
			if result != tt.isBackstab {
				t.Errorf("expected backstab=%v, got %v", tt.isBackstab, result)
			}
		})
	}
}

func TestIsFlankAngle(t *testing.T) {
	sys := NewPositionalSystem("fantasy")

	targetFacing := 0.0 // facing right

	tests := []struct {
		name      string
		attackerX float64
		attackerY float64
		isFlank   bool
	}{
		{"from front", 10, 0, false},
		{"from behind", -10, 0, false},
		{"from side (below)", 0, 10, true},
		{"from side (above)", 0, -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.IsFlankAngle(tt.attackerX, tt.attackerY, 0, 0, targetFacing)
			if result != tt.isFlank {
				t.Errorf("expected flank=%v, got %v", tt.isFlank, result)
			}
		})
	}
}

func TestApplyPositionalDamage(t *testing.T) {
	tests := []struct {
		baseDamage float64
		advantage  PositionalAdvantage
		multiplier float64
		expected   float64
	}{
		{100, AdvantageFrontal, 1.0, 100},
		{100, AdvantageFlank, 1.5, 150},
		{100, AdvantageBackstab, 2.0, 200},
		{50, AdvantageElevation, 1.3, 65},
	}

	for _, tt := range tests {
		result := ApplyPositionalDamage(tt.baseDamage, tt.advantage, tt.multiplier)
		if math.Abs(result-tt.expected) > 0.01 {
			t.Errorf("ApplyPositionalDamage(%f, %v, %f) = %f, expected %f",
				tt.baseDamage, tt.advantage, tt.multiplier, result, tt.expected)
		}
	}
}

func TestPositionalSystemSetGenre(t *testing.T) {
	sys := NewPositionalSystem("fantasy")

	originalBackstab := sys.config.BackstabMultiplier

	sys.SetGenre("horror")

	if sys.config.BackstabMultiplier == originalBackstab {
		t.Error("genre change should update config")
	}

	if sys.genreID != "horror" {
		t.Errorf("genre should be horror, got %s", sys.genreID)
	}
}

// Benchmarks

func BenchmarkCalculatePositionalAdvantage(b *testing.B) {
	cfg := GetPositionalConfig("fantasy")
	targetPos := &PositionalComponent{
		FacingAngle: 0,
		Height:      0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculatePositionalAdvantage(
			10, 0, 0, 0,
			nil, targetPos,
			cfg,
		)
	}
}

func BenchmarkPositionalSystemUpdate(b *testing.B) {
	w := engine.NewWorld()
	sys := NewPositionalSystem("fantasy")

	// Create test entities
	for i := 0; i < 100; i++ {
		e := w.AddEntity()
		w.AddComponent(e, &engine.Position{X: float64(i), Y: float64(i)})
		w.AddComponent(e, &engine.Velocity{DX: 1.0, DY: 0.0})
		sys.AddPositionalComponent(w, e, 0, 0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(w)
	}
}

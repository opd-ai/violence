package combat

import (
	"math"
	"reflect"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
	"github.com/opd-ai/violence/pkg/rng"
)

func TestTelegraphComponent(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		comp := &TelegraphComponent{}
		if comp.Phase != PhaseInactive {
			t.Errorf("expected inactive phase, got %v", comp.Phase)
		}
		if comp.HasHit {
			t.Error("expected HasHit to be false initially")
		}
	})

	t.Run("pattern assignment", func(t *testing.T) {
		pattern := AttackPattern{
			Name:       "Test",
			Shape:      ShapeCone,
			Range:      50,
			Angle:      math.Pi / 4,
			WindupTime: 0.5,
		}

		comp := &TelegraphComponent{
			Pattern: pattern,
		}

		if comp.Pattern.Name != "Test" {
			t.Errorf("expected pattern name 'Test', got %s", comp.Pattern.Name)
		}
		if comp.Pattern.Range != 50 {
			t.Errorf("expected range 50, got %f", comp.Pattern.Range)
		}
	})
}

func TestDefaultPatterns(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			r := rng.NewRNG(42)
			patterns := DefaultPatterns(genre, r)

			if len(patterns) == 0 {
				t.Fatal("expected at least one pattern")
			}

			// Verify all patterns have required fields
			for _, p := range patterns {
				if p.Name == "" {
					t.Error("pattern missing name")
				}
				if p.Range <= 0 {
					t.Errorf("pattern %s has invalid range: %f", p.Name, p.Range)
				}
				if p.WindupTime <= 0 {
					t.Errorf("pattern %s has invalid windup time: %f", p.Name, p.WindupTime)
				}
				if p.ActiveTime <= 0 {
					t.Errorf("pattern %s has invalid active time: %f", p.Name, p.ActiveTime)
				}
				if p.Damage <= 0 {
					t.Errorf("pattern %s has invalid damage: %f", p.Name, p.Damage)
				}
			}
		})
	}
}

func TestSelectPattern(t *testing.T) {
	r := rng.NewRNG(42)
	patterns := []AttackPattern{
		{Name: "Short", Range: 30},
		{Name: "Medium", Range: 60},
		{Name: "Long", Range: 120},
	}

	tests := []struct {
		name     string
		distance float64
		maxRange float64 // Expect pattern with range >= this
	}{
		{"close range", 20, 30},
		{"medium range", 50, 60},
		{"long range", 100, 120},
		{"out of range", 200, 120}, // Should return longest
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selected := SelectPattern(patterns, tt.distance, r)
			if selected.Range < tt.maxRange {
				t.Errorf("expected range >= %f, got %f", tt.maxRange, selected.Range)
			}
		})
	}
}

func TestTelegraphSystem_Update(t *testing.T) {
	w := engine.NewWorld()
	sys := NewTelegraphSystem("fantasy", 42)

	// Create enemy with telegraph
	enemy := w.AddEntity()
	w.AddComponent(enemy, &engine.Position{X: 100, Y: 100})
	w.AddComponent(enemy, &engine.Health{Current: 100, Max: 100})
	w.AddArchetypeComponent(enemy, engine.ComponentIDPosition)
	w.AddArchetypeComponent(enemy, engine.ComponentIDHealth)

	pattern := AttackPattern{
		Name:         "Test Attack",
		Shape:        ShapeCone,
		Range:        50,
		Angle:        math.Pi / 3,
		WindupTime:   0.5,
		ActiveTime:   0.2,
		CooldownTime: 1.0,
		Damage:       10,
	}

	telegraph := &TelegraphComponent{
		Pattern:    pattern,
		Phase:      PhaseInactive,
		DirectionX: 1,
		DirectionY: 0,
	}
	w.AddComponent(enemy, telegraph)

	t.Run("inactive phase does nothing", func(t *testing.T) {
		sys.Update(w, 0.1)
		if telegraph.Phase != PhaseInactive {
			t.Errorf("expected phase to remain inactive, got %v", telegraph.Phase)
		}
	})

	t.Run("windup phase progresses", func(t *testing.T) {
		telegraph.Phase = PhaseWindup
		telegraph.PhaseTimer = 0.5

		sys.Update(w, 0.1)

		if telegraph.PhaseTimer >= 0.5 {
			t.Error("expected timer to decrease")
		}
		if telegraph.Phase != PhaseWindup {
			t.Error("expected to remain in windup")
		}
	})

	t.Run("windup transitions to active", func(t *testing.T) {
		telegraph.Phase = PhaseWindup
		telegraph.PhaseTimer = 0.05

		sys.Update(w, 0.1)

		if telegraph.Phase != PhaseActive {
			t.Errorf("expected active phase, got %v", telegraph.Phase)
		}
		if telegraph.PhaseTimer <= 0 {
			t.Error("expected active timer to be set")
		}
	})

	t.Run("active phase deals damage once", func(t *testing.T) {
		// Create target in attack cone
		target := w.AddEntity()
		w.AddComponent(target, &engine.Position{X: 130, Y: 100})
		w.AddComponent(target, &engine.Health{Current: 50, Max: 50})
		w.AddArchetypeComponent(target, engine.ComponentIDPosition)
		w.AddArchetypeComponent(target, engine.ComponentIDHealth)

		telegraph.Phase = PhaseActive
		telegraph.PhaseTimer = 0.2
		telegraph.HasHit = false

		sys.Update(w, 0.05)

		// Check target took damage
		healthComp, _ := w.GetComponent(target, reflect.TypeOf(&engine.Health{}))
		health := healthComp.(*engine.Health)
		if health.Current >= 50 {
			t.Errorf("expected health < 50, got %d", health.Current)
		}

		if !telegraph.HasHit {
			t.Error("expected HasHit to be true")
		}

		// Second update should not deal damage again
		initialHealth := health.Current
		sys.Update(w, 0.05)
		if health.Current != initialHealth {
			t.Error("damage should only be dealt once during active phase")
		}
	})

	t.Run("active transitions to cooldown", func(t *testing.T) {
		telegraph.Phase = PhaseActive
		telegraph.PhaseTimer = 0.05
		telegraph.HasHit = true

		sys.Update(w, 0.1)

		if telegraph.Phase != PhaseCooldown {
			t.Errorf("expected cooldown phase, got %v", telegraph.Phase)
		}
	})

	t.Run("cooldown transitions to inactive", func(t *testing.T) {
		telegraph.Phase = PhaseCooldown
		telegraph.PhaseTimer = 0.05

		sys.Update(w, 0.1)

		if telegraph.Phase != PhaseInactive {
			t.Errorf("expected inactive phase, got %v", telegraph.Phase)
		}
	})

	t.Run("dead entity doesn't attack", func(t *testing.T) {
		healthComp, _ := w.GetComponent(enemy, reflect.TypeOf(&engine.Health{}))
		health := healthComp.(*engine.Health)
		health.Current = 0

		telegraph.Phase = PhaseWindup
		telegraph.PhaseTimer = 0.5

		sys.Update(w, 0.1)

		if telegraph.Phase != PhaseInactive {
			t.Error("dead entity should not attack")
		}
	})
}

func TestTelegraphSystem_InitiateAttack(t *testing.T) {
	w := engine.NewWorld()
	sys := NewTelegraphSystem("fantasy", 42)

	enemy := w.AddEntity()
	w.AddComponent(enemy, &engine.Position{X: 100, Y: 100})
	w.AddArchetypeComponent(enemy, engine.ComponentIDPosition)

	telegraph := &TelegraphComponent{
		Phase: PhaseInactive,
	}
	w.AddComponent(enemy, telegraph)

	t.Run("successful initiation", func(t *testing.T) {
		success := sys.InitiateAttack(w, enemy, 150, 100)

		if !success {
			t.Error("expected attack to initiate successfully")
		}
		if telegraph.Phase != PhaseWindup {
			t.Errorf("expected windup phase, got %v", telegraph.Phase)
		}
		if telegraph.PhaseTimer <= 0 {
			t.Error("expected positive timer")
		}

		// Direction should be normalized
		dist := math.Sqrt(telegraph.DirectionX*telegraph.DirectionX + telegraph.DirectionY*telegraph.DirectionY)
		if math.Abs(dist-1.0) > 0.001 {
			t.Errorf("expected normalized direction, got length %f", dist)
		}
	})

	t.Run("cannot initiate when busy", func(t *testing.T) {
		telegraph.Phase = PhaseWindup

		success := sys.InitiateAttack(w, enemy, 200, 100)

		if success {
			t.Error("should not be able to initiate when already attacking")
		}
	})

	t.Run("no telegraph component", func(t *testing.T) {
		noTelegraph := w.AddEntity()
		w.AddComponent(noTelegraph, &engine.Position{X: 0, Y: 0})
		w.AddArchetypeComponent(noTelegraph, engine.ComponentIDPosition)

		success := sys.InitiateAttack(w, noTelegraph, 10, 10)

		if success {
			t.Error("should fail when entity has no telegraph component")
		}
	})
}

func TestTelegraphSystem_CanAttack(t *testing.T) {
	w := engine.NewWorld()
	sys := NewTelegraphSystem("fantasy", 42)

	enemy := w.AddEntity()
	telegraph := &TelegraphComponent{Phase: PhaseInactive}
	w.AddComponent(enemy, telegraph)

	tests := []struct {
		phase     TelegraphPhase
		canAttack bool
	}{
		{PhaseInactive, true},
		{PhaseWindup, false},
		{PhaseActive, false},
		{PhaseCooldown, false},
	}

	for _, tt := range tests {
		t.Run(tt.phase.String(), func(t *testing.T) {
			telegraph.Phase = tt.phase

			can := sys.CanAttack(w, enemy)
			if can != tt.canAttack {
				t.Errorf("expected CanAttack=%v for phase %v, got %v", tt.canAttack, tt.phase, can)
			}
		})
	}
}

func TestTelegraphSystem_GetAttackProgress(t *testing.T) {
	w := engine.NewWorld()
	sys := NewTelegraphSystem("fantasy", 42)

	enemy := w.AddEntity()
	pattern := AttackPattern{WindupTime: 1.0, CooldownTime: 2.0}
	telegraph := &TelegraphComponent{Pattern: pattern}
	w.AddComponent(enemy, telegraph)

	tests := []struct {
		name        string
		phase       TelegraphPhase
		timer       float64
		expectedMin float64
		expectedMax float64
	}{
		{"inactive", PhaseInactive, 0, -1, -1},
		{"windup start", PhaseWindup, 1.0, 0, 0.1},
		{"windup mid", PhaseWindup, 0.5, 0.4, 0.6},
		{"windup end", PhaseWindup, 0.1, 0.8, 1.0},
		{"active", PhaseActive, 0.1, 1.0, 1.0},
		{"cooldown start", PhaseCooldown, 2.0, 0, 0.1},
		{"cooldown mid", PhaseCooldown, 1.0, 0.4, 0.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			telegraph.Phase = tt.phase
			telegraph.PhaseTimer = tt.timer

			progress := sys.GetAttackProgress(w, enemy)

			if progress < tt.expectedMin || progress > tt.expectedMax {
				t.Errorf("expected progress in [%f, %f], got %f", tt.expectedMin, tt.expectedMax, progress)
			}
		})
	}
}

func TestIsInAttackArea_Cone(t *testing.T) {
	sys := NewTelegraphSystem("fantasy", 42)

	attackerPos := &engine.Position{X: 100, Y: 100}

	telegraph := &TelegraphComponent{
		Pattern: AttackPattern{
			Shape: ShapeCone,
			Range: 50,
			Angle: math.Pi / 3, // 60 degrees
		},
		DirectionX: 1,
		DirectionY: 0,
	}

	tests := []struct {
		name     string
		targetX  float64
		targetY  float64
		expected bool
	}{
		{"in range and angle", 130, 100, true},
		{"in range, slight angle", 130, 110, true},
		{"in range, outside angle", 130, 150, false},
		{"out of range", 200, 100, false},
		{"behind attacker", 50, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetPos := &engine.Position{X: tt.targetX, Y: tt.targetY}
			result := sys.isInAttackArea(attackerPos, targetPos, telegraph)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInAttackArea_Circle(t *testing.T) {
	sys := NewTelegraphSystem("fantasy", 42)

	attackerPos := &engine.Position{X: 100, Y: 100}

	telegraph := &TelegraphComponent{
		Pattern: AttackPattern{
			Shape: ShapeCircle,
			Range: 50,
		},
	}

	tests := []struct {
		name     string
		targetX  float64
		targetY  float64
		expected bool
	}{
		{"at center", 100, 100, true},
		{"within range", 130, 100, true},
		{"on edge", 150, 100, true},
		{"just outside", 151, 100, false},
		{"far away", 200, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetPos := &engine.Position{X: tt.targetX, Y: tt.targetY}
			result := sys.isInAttackArea(attackerPos, targetPos, telegraph)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInAttackArea_Line(t *testing.T) {
	sys := NewTelegraphSystem("fantasy", 42)

	attackerPos := &engine.Position{X: 100, Y: 100}

	telegraph := &TelegraphComponent{
		Pattern: AttackPattern{
			Shape: ShapeLine,
			Range: 100,
			Width: 20,
		},
		DirectionX: 1,
		DirectionY: 0,
	}

	tests := []struct {
		name     string
		targetX  float64
		targetY  float64
		expected bool
	}{
		{"on line", 150, 100, true},
		{"within width", 150, 105, true},
		{"outside width", 150, 120, false},
		{"beyond range", 250, 100, false},
		{"behind", 50, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetPos := &engine.Position{X: tt.targetX, Y: tt.targetY}
			result := sys.isInAttackArea(attackerPos, targetPos, telegraph)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInAttackArea_Ring(t *testing.T) {
	sys := NewTelegraphSystem("fantasy", 42)

	attackerPos := &engine.Position{X: 100, Y: 100}

	telegraph := &TelegraphComponent{
		Pattern: AttackPattern{
			Shape: ShapeRing,
			Range: 50,
			Width: 10,
		},
	}

	tests := []struct {
		name     string
		targetX  float64
		targetY  float64
		expected bool
	}{
		{"in ring", 145, 100, true},
		{"on outer edge", 150, 100, true},
		{"on inner edge", 140.3, 100, true},
		{"inside ring", 110, 100, false},
		{"outside ring", 160, 100, false},
		{"at center", 100, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetPos := &engine.Position{X: tt.targetX, Y: tt.targetY}
			result := sys.isInAttackArea(attackerPos, targetPos, telegraph)

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTelegraphColors(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			warning, active := TelegraphColors(genre, PhaseWindup)

			wr, wg, wb, wa := warning.RGBA()
			ar, ag, ab, aa := active.RGBA()

			// Both colors should have some opacity
			if wa == 0 {
				t.Error("warning color has zero alpha")
			}
			if aa == 0 {
				t.Error("active color has zero alpha")
			}

			// Colors should be different
			if wr == ar && wg == ag && wb == ab {
				t.Error("warning and active colors should differ")
			}
		})
	}
}

// Helper for TelegraphPhase.String() for test output
func (p TelegraphPhase) String() string {
	switch p {
	case PhaseInactive:
		return "Inactive"
	case PhaseWindup:
		return "Windup"
	case PhaseActive:
		return "Active"
	case PhaseCooldown:
		return "Cooldown"
	default:
		return "Unknown"
	}
}

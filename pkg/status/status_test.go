package status

import (
	"reflect"
	"testing"
	"time"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.effects == nil {
		t.Error("Registry effects map not initialized")
	}
	// Should have default fantasy effects
	if len(r.effects) == 0 {
		t.Error("Registry should have default effects")
	}
}

func TestEffect(t *testing.T) {
	tests := []struct {
		name          string
		effectName    string
		duration      time.Duration
		damagePerTick float64
	}{
		{"poison", "Poisoned", 5 * time.Second, 2.0},
		{"burning", "Burning", 3 * time.Second, 5.0},
		{"bleeding", "Bleeding", 10 * time.Second, 1.0},
		{"irradiated", "Irradiated", 15 * time.Second, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effect := Effect{
				Name:          tt.effectName,
				Duration:      tt.duration,
				DamagePerTick: tt.damagePerTick,
			}

			if effect.Name != tt.effectName {
				t.Errorf("Name: expected %s, got %s", tt.effectName, effect.Name)
			}
			if effect.Duration != tt.duration {
				t.Errorf("Duration: expected %v, got %v", tt.duration, effect.Duration)
			}
			if effect.DamagePerTick != tt.damagePerTick {
				t.Errorf("DamagePerTick: expected %f, got %f", tt.damagePerTick, effect.DamagePerTick)
			}
		})
	}
}

func TestApply(t *testing.T) {
	r := NewRegistry()

	// Apply should not panic (deprecated method)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("Apply panicked: %v", rec)
		}
	}()

	r.Apply("poison")
	r.Apply("burning")
	r.Apply("bleeding")
}

func TestApplyToEntity(t *testing.T) {
	r := NewRegistry()
	w := engine.NewWorld()

	// Create entity with health
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Health{Current: 100, Max: 100})

	// Apply effect
	r.ApplyToEntity(w, entity, "poisoned")

	// Verify status component exists
	statusType := reflect.TypeOf(&StatusComponent{})
	comp, ok := w.GetComponent(entity, statusType)
	if !ok {
		t.Fatal("StatusComponent not added to entity")
	}

	statusComp := comp.(*StatusComponent)
	if len(statusComp.ActiveEffects) != 1 {
		t.Errorf("Expected 1 active effect, got %d", len(statusComp.ActiveEffects))
	}

	if statusComp.ActiveEffects[0].EffectName != "poisoned" {
		t.Errorf("Expected 'poisoned', got '%s'", statusComp.ActiveEffects[0].EffectName)
	}
}

func TestNonStackableEffects(t *testing.T) {
	r := NewRegistry()
	w := engine.NewWorld()

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Health{Current: 100, Max: 100})

	// Apply poisoned twice
	r.ApplyToEntity(w, entity, "poisoned")
	r.ApplyToEntity(w, entity, "poisoned")

	statusType := reflect.TypeOf(&StatusComponent{})
	comp, _ := w.GetComponent(entity, statusType)
	statusComp := comp.(*StatusComponent)

	// Should still only have one instance
	if len(statusComp.ActiveEffects) != 1 {
		t.Errorf("Non-stackable effect should not stack, got %d instances", len(statusComp.ActiveEffects))
	}
}

func TestStackableEffects(t *testing.T) {
	r := NewRegistry()
	w := engine.NewWorld()

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Health{Current: 100, Max: 100})

	// Bleeding is stackable in fantasy genre
	r.ApplyToEntity(w, entity, "bleeding")
	r.ApplyToEntity(w, entity, "bleeding")

	statusType := reflect.TypeOf(&StatusComponent{})
	comp, _ := w.GetComponent(entity, statusType)
	statusComp := comp.(*StatusComponent)

	// Should have multiple instances
	if len(statusComp.ActiveEffects) != 2 {
		t.Errorf("Stackable effect should stack, got %d instances", len(statusComp.ActiveEffects))
	}
}

func TestTick(t *testing.T) {
	r := NewRegistry()

	// Tick should not panic on empty registry (deprecated)
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("Tick panicked: %v", rec)
		}
	}()

	r.Tick()
}

func TestSystemUpdate(t *testing.T) {
	r := NewRegistry()
	s := NewSystem(r)
	w := engine.NewWorld()

	// Create entity with health and poison effect
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Health{Current: 100, Max: 100})
	r.ApplyToEntity(w, entity, "poisoned")

	// Wait for tick interval
	time.Sleep(1100 * time.Millisecond)

	// Run system update
	s.Update(w)

	// Check health decreased
	healthType := reflect.TypeOf(&engine.Health{})
	comp, _ := w.GetComponent(entity, healthType)
	health := comp.(*engine.Health)

	if health.Current >= 100 {
		t.Errorf("Poison should have dealt damage, health: %d", health.Current)
	}
}

func TestHealingEffect(t *testing.T) {
	r := NewRegistry()
	s := NewSystem(r)
	w := engine.NewWorld()

	// Create entity with low health
	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Health{Current: 50, Max: 100})
	r.ApplyToEntity(w, entity, "regeneration")

	initialHealth := 50

	// Wait for tick interval
	time.Sleep(1100 * time.Millisecond)

	// Run system update
	s.Update(w)

	// Check health increased
	healthType := reflect.TypeOf(&engine.Health{})
	comp, _ := w.GetComponent(entity, healthType)
	health := comp.(*engine.Health)

	if health.Current <= initialHealth {
		t.Errorf("Regeneration should have healed, health: %d", health.Current)
	}
}

func TestEffectExpiration(t *testing.T) {
	r := NewRegistry()
	s := NewSystem(r)
	w := engine.NewWorld()

	entity := w.AddEntity()
	w.AddComponent(entity, &engine.Health{Current: 100, Max: 100})

	// Apply short-duration stun
	r.ApplyToEntity(w, entity, "stunned")

	// Simulate many frames (~2.5 seconds at 60fps)
	for i := 0; i < 150; i++ {
		s.Update(w)
		time.Sleep(16 * time.Millisecond)
	}

	// Effect should have expired
	statusType := reflect.TypeOf(&StatusComponent{})
	comp, _ := w.GetComponent(entity, statusType)
	statusComp := comp.(*StatusComponent)

	if len(statusComp.ActiveEffects) != 0 {
		t.Errorf("Effect should have expired, got %d active effects", len(statusComp.ActiveEffects))
	}
}

func TestGetSpeedMultiplier(t *testing.T) {
	r := NewRegistry()
	w := engine.NewWorld()

	entity := w.AddEntity()

	// No effects - should be 1.0
	speed := GetSpeedMultiplier(w, entity)
	if speed != 1.0 {
		t.Errorf("No effects: expected speed 1.0, got %f", speed)
	}

	// Apply slow effect
	r.ApplyToEntity(w, entity, "slowed")
	speed = GetSpeedMultiplier(w, entity)
	if speed >= 1.0 {
		t.Errorf("Slowed: expected speed < 1.0, got %f", speed)
	}
}

func TestIsStunned(t *testing.T) {
	r := NewRegistry()
	w := engine.NewWorld()

	entity := w.AddEntity()

	// Not stunned initially
	if IsStunned(w, entity) {
		t.Error("Entity should not be stunned initially")
	}

	// Apply stun
	r.ApplyToEntity(w, entity, "stunned")
	if !IsStunned(w, entity) {
		t.Error("Entity should be stunned after applying stun effect")
	}
}

func TestSetGenre(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			defer func() {
				if rec := recover(); rec != nil {
					t.Errorf("SetGenre(%q) panicked: %v", genre, rec)
				}
			}()
			SetGenre(genre)
		})
	}
}

func TestGenreSpecificEffects(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			r := NewRegistry()
			r.loadDefaultEffects(genre)

			if len(r.effects) == 0 {
				t.Errorf("Genre %s should have effects loaded", genre)
			}
		})
	}
}

func TestEffectDuration(t *testing.T) {
	durations := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		10 * time.Second,
		1 * time.Minute,
	}

	for _, dur := range durations {
		t.Run(dur.String(), func(t *testing.T) {
			effect := Effect{
				Name:          "Test",
				Duration:      dur,
				DamagePerTick: 1.0,
			}
			if effect.Duration != dur {
				t.Errorf("Duration mismatch: expected %v, got %v", dur, effect.Duration)
			}
		})
	}
}

func TestEffectZeroDamage(t *testing.T) {
	effect := Effect{
		Name:          "Slow",
		Duration:      5 * time.Second,
		DamagePerTick: 0.0,
	}

	if effect.DamagePerTick != 0.0 {
		t.Errorf("Zero damage effect should have 0.0 damage, got %f", effect.DamagePerTick)
	}
}

func TestEffectNegativeDamage(t *testing.T) {
	// Negative damage could represent healing
	effect := Effect{
		Name:          "Regeneration",
		Duration:      10 * time.Second,
		DamagePerTick: -2.0,
	}

	if effect.DamagePerTick != -2.0 {
		t.Errorf("Healing effect should have negative damage, got %f", effect.DamagePerTick)
	}
}

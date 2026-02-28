package status

import (
	"testing"
	"time"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if r.effects == nil {
		t.Error("Registry effects map not initialized")
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

	// Apply should not panic
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("Apply panicked: %v", rec)
		}
	}()

	r.Apply("poison")
	r.Apply("burning")
	r.Apply("bleeding")
}

func TestTick(t *testing.T) {
	r := NewRegistry()

	// Tick should not panic on empty registry
	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("Tick panicked: %v", rec)
		}
	}()

	r.Tick()

	// Apply some effects and tick
	r.Apply("poison")
	r.Apply("burning")
	r.Tick()
	r.Tick()
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

func TestMultipleEffects(t *testing.T) {
	r := NewRegistry()

	// Apply multiple effects
	effects := []string{"poison", "burning", "bleeding", "irradiated", "stunned"}
	for _, eff := range effects {
		r.Apply(eff)
	}

	// Should handle multiple ticks
	for i := 0; i < 10; i++ {
		r.Tick()
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

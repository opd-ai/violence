// Package combat handles damage calculation and combat events.
package combat

import (
	"math/rand"
	"testing"
)

func TestCreateBossPhases(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))

	tests := []struct {
		name          string
		genre         string
		minPhases     int
		maxPhases     int
		expectEnraged bool
	}{
		{"fantasy", "fantasy", 3, 3, true},
		{"scifi", "scifi", 3, 3, true},
		{"horror", "horror", 4, 4, true},
		{"cyberpunk", "cyberpunk", 3, 3, true},
		{"default", "unknown", 3, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phases := CreateBossPhases(tt.genre, rng)

			if len(phases) < tt.minPhases || len(phases) > tt.maxPhases {
				t.Errorf("expected %d-%d phases, got %d", tt.minPhases, tt.maxPhases, len(phases))
			}

			// Verify phase progression
			for i := 0; i < len(phases)-1; i++ {
				if phases[i].HealthThreshold <= phases[i+1].HealthThreshold {
					t.Errorf("phase %d threshold (%f) should be > phase %d threshold (%f)",
						i, phases[i].HealthThreshold, i+1, phases[i+1].HealthThreshold)
				}

				if phases[i].DamageMultiplier > phases[i+1].DamageMultiplier {
					t.Errorf("phase %d damage (%f) should increase over phase %d (%f)",
						i, phases[i].DamageMultiplier, i+1, phases[i+1].DamageMultiplier)
				}
			}

			// Verify enraged state in final phase
			if tt.expectEnraged {
				finalPhase := phases[len(phases)-1]
				if !finalPhase.Enraged {
					t.Error("final phase should be enraged")
				}
			}

			// Verify all phases have abilities
			for i, phase := range phases {
				if len(phase.AbilitySet) == 0 {
					t.Errorf("phase %d has no abilities", i)
				}
			}
		})
	}
}

func TestBossPhaseComponent_ShouldTransition(t *testing.T) {
	boss := &BossPhaseComponent{
		CurrentPhase:       0,
		TransitionCooldown: 1.0,
		Phases: []PhaseTransition{
			{PhaseID: 0, HealthThreshold: 1.0},
			{PhaseID: 1, HealthThreshold: 0.66},
			{PhaseID: 2, HealthThreshold: 0.33},
		},
	}

	tests := []struct {
		name          string
		currentHealth float64
		maxHealth     float64
		gameTime      float64
		setup         func()
		expect        bool
	}{
		{
			name:          "above_threshold",
			currentHealth: 700,
			maxHealth:     1000,
			gameTime:      5.0,
			setup:         func() {},
			expect:        false,
		},
		{
			name:          "below_threshold",
			currentHealth: 600,
			maxHealth:     1000,
			gameTime:      5.0,
			setup:         func() {},
			expect:        true,
		},
		{
			name:          "during_transition",
			currentHealth: 600,
			maxHealth:     1000,
			gameTime:      5.0,
			setup: func() {
				boss.IsTransitioning = true
			},
			expect: false,
		},
		{
			name:          "cooldown_active",
			currentHealth: 600,
			maxHealth:     1000,
			gameTime:      5.0,
			setup: func() {
				boss.IsTransitioning = false
				boss.LastTransitionTime = 4.5
			},
			expect: false,
		},
		{
			name:          "no_more_phases",
			currentHealth: 100,
			maxHealth:     1000,
			gameTime:      10.0,
			setup: func() {
				boss.IsTransitioning = false
				boss.LastTransitionTime = 0
				boss.CurrentPhase = 2
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			boss.CurrentPhase = 0
			boss.IsTransitioning = false
			boss.LastTransitionTime = 0

			tt.setup()

			result := boss.ShouldTransition(tt.currentHealth, tt.maxHealth, tt.gameTime)
			if result != tt.expect {
				t.Errorf("expected %v, got %v", tt.expect, result)
			}
		})
	}
}

func TestBossPhaseComponent_UpdateTransition(t *testing.T) {
	boss := &BossPhaseComponent{
		CurrentPhase: 0,
		Phases: []PhaseTransition{
			{PhaseID: 0, HealthThreshold: 1.0},
			{PhaseID: 1, HealthThreshold: 0.5},
		},
		IsTransitioning:    true,
		TransitionProgress: 0.0,
		PhaseChangeCount:   0,
	}

	// Update partway through transition
	completed := boss.UpdateTransition(0.2)
	if completed {
		t.Error("transition should not be complete after 0.2s")
	}
	if boss.CurrentPhase != 0 {
		t.Error("phase should not change during transition")
	}
	if boss.TransitionProgress <= 0.0 {
		t.Error("transition progress should advance")
	}

	// Complete transition
	completed = boss.UpdateTransition(0.5)
	if !completed {
		t.Error("transition should complete after 0.7s total")
	}
	if boss.CurrentPhase != 1 {
		t.Error("phase should advance on completion")
	}
	if boss.IsTransitioning {
		t.Error("transition flag should clear on completion")
	}
	if boss.PhaseChangeCount != 1 {
		t.Errorf("expected 1 phase change, got %d", boss.PhaseChangeCount)
	}
}

func TestBossPhaseComponent_GetCurrentPhaseData(t *testing.T) {
	boss := &BossPhaseComponent{
		CurrentPhase: 1,
		Phases: []PhaseTransition{
			{PhaseID: 0, DamageMultiplier: 1.0},
			{PhaseID: 1, DamageMultiplier: 1.5},
			{PhaseID: 2, DamageMultiplier: 2.0},
		},
	}

	phase := boss.GetCurrentPhaseData()
	if phase == nil {
		t.Fatal("phase should not be nil")
	}
	if phase.PhaseID != 1 {
		t.Errorf("expected phase 1, got %d", phase.PhaseID)
	}
	if phase.DamageMultiplier != 1.5 {
		t.Errorf("expected damage 1.5, got %f", phase.DamageMultiplier)
	}

	// Test invalid phase
	boss.CurrentPhase = 99
	phase = boss.GetCurrentPhaseData()
	if phase != nil {
		t.Error("invalid phase should return nil")
	}
}

func TestBossPhaseComponent_StartTransition(t *testing.T) {
	boss := &BossPhaseComponent{
		IsTransitioning:    false,
		TransitionProgress: 0.5,
		LastTransitionTime: 0.0,
	}

	boss.StartTransition(10.0)

	if !boss.IsTransitioning {
		t.Error("transition flag should be set")
	}
	if boss.TransitionProgress != 0.0 {
		t.Error("progress should reset to 0")
	}
	if boss.LastTransitionTime != 10.0 {
		t.Error("last transition time should be updated")
	}
}

func TestPhaseTransitionScaling(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "default"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			phases := CreateBossPhases(genre, rng)

			// Verify damage scaling increases
			prevDamage := 0.0
			for i, phase := range phases {
				if phase.DamageMultiplier < prevDamage {
					t.Errorf("phase %d damage (%f) should not decrease from phase %d (%f)",
						i, phase.DamageMultiplier, i-1, prevDamage)
				}
				prevDamage = phase.DamageMultiplier
			}

			// Verify health thresholds decrease
			prevThreshold := 2.0
			for i, phase := range phases {
				if phase.HealthThreshold > prevThreshold {
					t.Errorf("phase %d threshold (%f) should not increase from phase %d (%f)",
						i, phase.HealthThreshold, i-1, prevThreshold)
				}
				prevThreshold = phase.HealthThreshold
			}

			// Verify ability count increases or stays same
			prevAbilities := 0
			for i, phase := range phases {
				if len(phase.AbilitySet) < prevAbilities {
					t.Errorf("phase %d abilities (%d) should not decrease from phase %d (%d)",
						i, len(phase.AbilitySet), i-1, prevAbilities)
				}
				prevAbilities = len(phase.AbilitySet)
			}
		})
	}
}

func TestFantasyBossPhases(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	phases := createFantasyBossPhases(rng)

	if len(phases) != 3 {
		t.Errorf("expected 3 fantasy phases, got %d", len(phases))
	}

	// Verify whirlwind added in phase 2
	found := false
	for _, ability := range phases[1].AbilitySet {
		if ability == "whirlwind" {
			found = true
		}
	}
	if !found {
		t.Error("phase 1 should include whirlwind ability")
	}

	// Verify enraged in final phase
	if !phases[2].Enraged {
		t.Error("final fantasy phase should be enraged")
	}
}

func TestHorrorBossPhases(t *testing.T) {
	rng := rand.New(rand.NewSource(456))
	phases := createHorrorBossPhases(rng)

	if len(phases) != 4 {
		t.Errorf("expected 4 horror phases, got %d", len(phases))
	}

	// Verify slow start
	if phases[0].SpeedMultiplier != 0.8 {
		t.Errorf("horror phase 0 should be slow, got %f", phases[0].SpeedMultiplier)
	}

	// Verify escalation
	if phases[3].SpeedMultiplier <= phases[0].SpeedMultiplier {
		t.Error("horror final phase should be faster than initial")
	}

	// Verify spawn minions mechanic
	found := false
	for _, ability := range phases[1].AbilitySet {
		if ability == "spawn_minions" {
			found = true
		}
	}
	if !found {
		t.Error("horror phase 1 should spawn minions")
	}
}

func BenchmarkBossPhaseTransition(b *testing.B) {
	boss := &BossPhaseComponent{
		CurrentPhase: 0,
		Phases: []PhaseTransition{
			{PhaseID: 0, HealthThreshold: 1.0},
			{PhaseID: 1, HealthThreshold: 0.5},
			{PhaseID: 2, HealthThreshold: 0.25},
		},
		TransitionCooldown: 1.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		boss.ShouldTransition(400, 1000, float64(i)*0.016)
	}
}

func BenchmarkCreateBossPhases(b *testing.B) {
	rng := rand.New(rand.NewSource(999))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateBossPhases("fantasy", rng)
	}
}

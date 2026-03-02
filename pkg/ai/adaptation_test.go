package ai

import (
	"testing"
)

func TestNewPlayerBehaviorProfile(t *testing.T) {
	profile := NewPlayerBehaviorProfile()
	if profile == nil {
		t.Fatal("NewPlayerBehaviorProfile returned nil")
	}
	if profile.AverageEngagementRange != 10.0 {
		t.Errorf("Expected default range 10.0, got %f", profile.AverageEngagementRange)
	}
	if profile.ObservationCount != 0 {
		t.Errorf("Expected 0 observations, got %d", profile.ObservationCount)
	}
}

func TestRecordObservation(t *testing.T) {
	tests := []struct {
		name             string
		observations     []PlayerTactic
		expectedDominant PlayerTactic
	}{
		{
			name:             "melee dominant",
			observations:     []PlayerTactic{TacticRushMelee, TacticRushMelee, TacticRushMelee, TacticKiteRanged},
			expectedDominant: TacticRushMelee,
		},
		{
			name:             "ranged dominant",
			observations:     []PlayerTactic{TacticKiteRanged, TacticKiteRanged, TacticKiteRanged, TacticRushMelee},
			expectedDominant: TacticKiteRanged,
		},
		{
			name:             "explosives dominant",
			observations:     []PlayerTactic{TacticExplosives, TacticExplosives, TacticExplosives, TacticRushMelee},
			expectedDominant: TacticExplosives,
		},
		{
			name:             "insufficient data",
			observations:     []PlayerTactic{TacticRushMelee},
			expectedDominant: TacticUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := NewPlayerBehaviorProfile()
			for i, tactic := range tt.observations {
				obs := TacticObservation{
					Tactic:     tactic,
					Timestamp:  float64(i),
					Confidence: 0.8,
				}
				profile.RecordObservation(obs)
			}

			dominant := profile.GetDominantTactic()
			if dominant != tt.expectedDominant {
				t.Errorf("Expected dominant tactic %v, got %v", tt.expectedDominant, dominant)
			}
		})
	}
}

func TestUpdateRange(t *testing.T) {
	profile := NewPlayerBehaviorProfile()

	// First observation should set the range
	profile.ObservationCount = 1
	profile.UpdateRange(5.0)
	if profile.AverageEngagementRange != 5.0 {
		t.Errorf("Expected range 5.0, got %f", profile.AverageEngagementRange)
	}

	// Subsequent observations should be averaged
	profile.ObservationCount = 2
	profile.UpdateRange(15.0)

	// Should be between 5 and 15 due to exponential moving average
	if profile.AverageEngagementRange <= 5.0 || profile.AverageEngagementRange >= 15.0 {
		t.Errorf("Expected range between 5.0 and 15.0, got %f", profile.AverageEngagementRange)
	}
}

func TestRecordFlankAttempt(t *testing.T) {
	profile := NewPlayerBehaviorProfile()
	if profile.FlankingAttempts != 0 {
		t.Errorf("Expected 0 flanking attempts, got %d", profile.FlankingAttempts)
	}

	profile.RecordFlankAttempt()
	if profile.FlankingAttempts != 1 {
		t.Errorf("Expected 1 flanking attempt, got %d", profile.FlankingAttempts)
	}

	profile.RecordFlankAttempt()
	profile.RecordFlankAttempt()
	if profile.FlankingAttempts != 3 {
		t.Errorf("Expected 3 flanking attempts, got %d", profile.FlankingAttempts)
	}
}

func TestGetDominantTacticCoverBased(t *testing.T) {
	profile := NewPlayerBehaviorProfile()
	profile.ObservationCount = 10

	// Simulate cover usage
	for i := 0; i < 8; i++ {
		profile.RecordObservation(TacticObservation{
			Tactic:     TacticCoverBased,
			Timestamp:  float64(i),
			Confidence: 0.8,
		})
	}

	dominant := profile.GetDominantTactic()
	if dominant != TacticCoverBased {
		t.Errorf("Expected TacticCoverBased, got %v", dominant)
	}
}

func TestGetDominantTacticHitAndRun(t *testing.T) {
	profile := NewPlayerBehaviorProfile()
	profile.ObservationCount = 15
	profile.HitAndRunCount = 8

	dominant := profile.GetDominantTactic()
	if dominant != TacticHitAndRun {
		t.Errorf("Expected TacticHitAndRun, got %v", dominant)
	}
}

func TestComputeAdaptation(t *testing.T) {
	tests := []struct {
		name            string
		setupProfile    func(*PlayerBehaviorProfile)
		checkAdaptation func(*testing.T, AIAdaptation)
	}{
		{
			name: "counter melee rusher",
			setupProfile: func(p *PlayerBehaviorProfile) {
				p.ObservationCount = 10
				p.MeleeFrequency = 0.9
				p.AverageEngagementRange = 3.0
			},
			checkAdaptation: func(t *testing.T, a AIAdaptation) {
				if a.PreferredRangeMultiplier <= 1.0 {
					t.Error("Expected increased range multiplier for melee counter")
				}
				if a.DodgeFrequency <= 0.5 {
					t.Error("Expected increased dodge frequency for melee counter")
				}
			},
		},
		{
			name: "counter ranged kiter",
			setupProfile: func(p *PlayerBehaviorProfile) {
				p.ObservationCount = 10
				p.RangedFrequency = 0.9
				p.AverageEngagementRange = 15.0
			},
			checkAdaptation: func(t *testing.T, a AIAdaptation) {
				if a.PursuitAggression <= 0.7 {
					t.Error("Expected high pursuit aggression for kiter counter")
				}
				if a.FocusFirePriority <= 0.7 {
					t.Error("Expected high focus fire for kiter counter")
				}
			},
		},
		{
			name: "counter cover user",
			setupProfile: func(p *PlayerBehaviorProfile) {
				p.ObservationCount = 10
				p.PrefersCover = 0.8
			},
			checkAdaptation: func(t *testing.T, a AIAdaptation) {
				if a.FlankingPriority <= 0.8 {
					t.Error("Expected high flanking priority for cover counter")
				}
				if a.AlertnessModifier <= 1.0 {
					t.Error("Expected increased alertness for cover counter")
				}
			},
		},
		{
			name: "counter stealth",
			setupProfile: func(p *PlayerBehaviorProfile) {
				p.ObservationCount = 10
				p.StealthFrequency = 0.9
			},
			checkAdaptation: func(t *testing.T, a AIAdaptation) {
				if a.AlertnessModifier <= 1.3 {
					t.Error("Expected high alertness modifier for stealth counter")
				}
				if a.FocusFirePriority <= 0.8 {
					t.Error("Expected high focus fire for stealth counter")
				}
			},
		},
		{
			name: "counter explosives",
			setupProfile: func(p *PlayerBehaviorProfile) {
				p.ObservationCount = 10
				p.ExplosiveFrequency = 0.9
			},
			checkAdaptation: func(t *testing.T, a AIAdaptation) {
				if a.SpreadFormation <= 0.8 {
					t.Error("Expected high spread formation for explosive counter")
				}
				if a.DodgeFrequency <= 0.6 {
					t.Error("Expected high dodge frequency for explosive counter")
				}
			},
		},
		{
			name: "counter hit-and-run",
			setupProfile: func(p *PlayerBehaviorProfile) {
				p.ObservationCount = 15
				p.HitAndRunCount = 8
			},
			checkAdaptation: func(t *testing.T, a AIAdaptation) {
				if a.PursuitAggression <= 0.8 {
					t.Error("Expected high pursuit for hit-and-run counter")
				}
				if a.RetreatThreshold >= 0.3 {
					t.Error("Expected low retreat threshold for hit-and-run counter")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := NewPlayerBehaviorProfile()
			tt.setupProfile(profile)

			adaptation := ComputeAdaptation(profile)
			tt.checkAdaptation(t, adaptation)
		})
	}
}

func TestAdaptationRangeTuning(t *testing.T) {
	// Close range tuning
	profile := NewPlayerBehaviorProfile()
	profile.ObservationCount = 10
	profile.AverageEngagementRange = 3.0

	adaptation := ComputeAdaptation(profile)
	if adaptation.PreferredRangeMultiplier <= 1.0 {
		t.Error("Expected range multiplier increase for close engagements")
	}
	if adaptation.SpreadFormation <= 0.5 {
		t.Error("Expected spread formation increase for close engagements")
	}

	// Long range tuning
	profile.AverageEngagementRange = 20.0
	adaptation = ComputeAdaptation(profile)
	if adaptation.PreferredRangeMultiplier >= 1.0 {
		t.Error("Expected range multiplier decrease for long engagements")
	}
	if adaptation.PursuitAggression <= 0.5 {
		t.Error("Expected pursuit aggression increase for long engagements")
	}
}

func TestAdaptationFlankingTuning(t *testing.T) {
	profile := NewPlayerBehaviorProfile()
	profile.ObservationCount = 10
	profile.FlankingAttempts = 15

	adaptation := ComputeAdaptation(profile)
	if adaptation.AlertnessModifier <= 1.0 {
		t.Error("Expected alertness increase when player flanks frequently")
	}
	if adaptation.SpreadFormation <= 0.5 {
		t.Error("Expected spread formation increase when player flanks frequently")
	}
}

func TestLerp(t *testing.T) {
	tests := []struct {
		a, b, t, expected float64
	}{
		{0.0, 10.0, 0.0, 0.0},
		{0.0, 10.0, 1.0, 10.0},
		{0.0, 10.0, 0.5, 5.0},
		{5.0, 15.0, 0.25, 7.5},
	}

	for _, tt := range tests {
		result := lerp(tt.a, tt.b, tt.t)
		if result != tt.expected {
			t.Errorf("lerp(%f, %f, %f) = %f, expected %f", tt.a, tt.b, tt.t, result, tt.expected)
		}
	}
}

func TestAdaptationComponentType(t *testing.T) {
	comp := &AdaptationComponent{}
	if comp.Type() != "AdaptationComponent" {
		t.Errorf("Expected type 'AdaptationComponent', got '%s'", comp.Type())
	}
}

func TestPlayerProfileComponentType(t *testing.T) {
	comp := &PlayerProfileComponent{}
	if comp.Type() != "PlayerProfileComponent" {
		t.Errorf("Expected type 'PlayerProfileComponent', got '%s'", comp.Type())
	}
}

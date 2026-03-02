// Package combat - Integration test for positional system
package combat

import (
	"math"
	"testing"
)

func TestPositionalSystemIntegration(t *testing.T) {
	// Test that all genre configs are valid
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			cfg := GetPositionalConfig(genre)

			// Validate backstab is stronger than flank
			if cfg.BackstabMultiplier <= cfg.FlankMultiplier {
				t.Errorf("%s: backstab multiplier (%f) should be > flank multiplier (%f)",
					genre, cfg.BackstabMultiplier, cfg.FlankMultiplier)
			}

			// Validate angle thresholds are reasonable
			if cfg.BackstabAngle <= 0 || cfg.BackstabAngle > math.Pi {
				t.Errorf("%s: backstab angle (%f) out of valid range",
					genre, cfg.BackstabAngle)
			}

			if cfg.FlankAngle <= 0 || cfg.FlankAngle > math.Pi {
				t.Errorf("%s: flank angle (%f) out of valid range",
					genre, cfg.FlankAngle)
			}

			// Validate elevation makes sense
			if cfg.ElevationThreshold <= 0 {
				t.Errorf("%s: elevation threshold (%f) should be positive",
					genre, cfg.ElevationThreshold)
			}

			if cfg.ElevationMultiplier < 1.0 {
				t.Errorf("%s: elevation multiplier (%f) should be >= 1.0",
					genre, cfg.ElevationMultiplier)
			}
		})
	}
}

func TestPositionalCalculations(t *testing.T) {
	cfg := GetPositionalConfig("fantasy")

	// Create target facing right (0 radians)
	target := &PositionalComponent{
		FacingAngle: 0,
		Height:      0,
	}

	tests := []struct {
		name           string
		attackerX      float64
		attackerY      float64
		expectedAdv    PositionalAdvantage
		shouldMultiply bool
	}{
		{
			name:           "frontal",
			attackerX:      10,
			attackerY:      0,
			expectedAdv:    AdvantageFrontal,
			shouldMultiply: false,
		},
		{
			name:           "backstab",
			attackerX:      -10,
			attackerY:      0,
			expectedAdv:    AdvantageBackstab,
			shouldMultiply: true,
		},
		{
			name:           "flank_top",
			attackerX:      0,
			attackerY:      -10,
			expectedAdv:    AdvantageFlank,
			shouldMultiply: true,
		},
		{
			name:           "flank_bottom",
			attackerX:      0,
			attackerY:      10,
			expectedAdv:    AdvantageFlank,
			shouldMultiply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adv, mul := CalculatePositionalAdvantage(
				tt.attackerX, tt.attackerY,
				0, 0,
				nil, target,
				cfg,
			)

			if adv != tt.expectedAdv {
				t.Errorf("expected advantage %v, got %v", tt.expectedAdv, adv)
			}

			if tt.shouldMultiply && mul <= 1.0 {
				t.Errorf("expected multiplier > 1.0, got %f", mul)
			}

			if !tt.shouldMultiply && mul != 1.0 {
				t.Errorf("expected multiplier = 1.0, got %f", mul)
			}
		})
	}
}

func TestPositionalSystemCreation(t *testing.T) {
	sys := NewPositionalSystem("fantasy")

	if sys == nil {
		t.Fatal("NewPositionalSystem returned nil")
	}

	if sys.genreID != "fantasy" {
		t.Errorf("expected genre fantasy, got %s", sys.genreID)
	}

	// Test genre change
	sys.SetGenre("horror")
	if sys.genreID != "horror" {
		t.Errorf("SetGenre failed, expected horror, got %s", sys.genreID)
	}
}

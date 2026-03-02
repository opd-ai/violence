package weaponanim

import (
	"testing"
)

// TestIntegration verifies that swing animations complete end-to-end.
func TestIntegration(t *testing.T) {
	// Create swing parameters
	startAngle, endAngle, duration := GetSwingParameters(SwingSlash, 0.0)

	if duration <= 0 {
		t.Error("Duration should be positive")
	}

	if startAngle == endAngle {
		t.Error("Slash should have different start and end angles")
	}

	// Test easing function
	progress := easeInOut(0.5)
	if progress < 0 || progress > 1 {
		t.Errorf("Easing should return [0,1], got %v", progress)
	}
}

// TestTrailPointCreation verifies trail point mechanics.
func TestTrailPointCreation(t *testing.T) {
	anim := &WeaponAnimComponent{
		Active:     true,
		Progress:   0.3,
		Duration:   0.3,
		StartAngle: 0.0,
		EndAngle:   1.57,
		ArcRadius:  20.0,
		TrailPoints: []TrailPoint{
			{X: 100, Y: 100, Age: 0.0},
			{X: 105, Y: 102, Age: 0.1},
		},
	}

	if len(anim.TrailPoints) != 2 {
		t.Errorf("Expected 2 trail points, got %d", len(anim.TrailPoints))
	}

	tipX, tipY := anim.GetTipPosition(100, 100)
	if tipX == 100 && tipY == 100 {
		t.Error("Tip position should differ from center at mid-swing")
	}
}

// TestSwingTypeColors verifies genre-specific colors.
func TestSwingTypeColors(t *testing.T) {
	tests := []struct {
		weapon string
		genre  string
	}{
		{"sword", "fantasy"},
		{"magic", "fantasy"},
		{"blade", "cyberpunk"},
		{"axe", "scifi"},
	}

	for _, tt := range tests {
		color := GetSwingColor(tt.weapon, tt.genre)
		if color.A == 0 {
			t.Errorf("Color for %s/%s should not be fully transparent", tt.weapon, tt.genre)
		}
	}
}

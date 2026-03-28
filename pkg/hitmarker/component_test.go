package hitmarker

import (
	"image/color"
	"testing"
)

func TestComponentType(t *testing.T) {
	comp := NewComponent()
	if comp.Type() != "hitmarker" {
		t.Errorf("Type() = %q, want %q", comp.Type(), "hitmarker")
	}
}

func TestNewComponent(t *testing.T) {
	comp := NewComponent()

	if comp.Active {
		t.Error("New component should not be active")
	}
	if comp.HitType != HitNormal {
		t.Errorf("HitType = %v, want HitNormal", comp.HitType)
	}
	if comp.Duration != 0.25 {
		t.Errorf("Duration = %f, want 0.25", comp.Duration)
	}
	if comp.Alpha != 1.0 {
		t.Errorf("Alpha = %f, want 1.0", comp.Alpha)
	}
	if comp.Scale != 1.0 {
		t.Errorf("Scale = %f, want 1.0", comp.Scale)
	}
}

func TestTrigger(t *testing.T) {
	tests := []struct {
		name       string
		hitType    HitType
		damage     int
		wantActive bool
		wantMinDur float64
		wantMaxDur float64
		wantMinInt float64
	}{
		{
			name:       "normal_hit",
			hitType:    HitNormal,
			damage:     25,
			wantActive: true,
			wantMinDur: 0.2,
			wantMaxDur: 0.3,
			wantMinInt: 0.9,
		},
		{
			name:       "critical_hit",
			hitType:    HitCritical,
			damage:     75,
			wantActive: true,
			wantMinDur: 0.3,
			wantMaxDur: 0.4,
			wantMinInt: 1.4,
		},
		{
			name:       "kill_hit",
			hitType:    HitKill,
			damage:     150,
			wantActive: true,
			wantMinDur: 0.4,
			wantMaxDur: 0.6,
			wantMinInt: 3.0,
		},
		{
			name:       "headshot",
			hitType:    HitHeadshot,
			damage:     100,
			wantActive: true,
			wantMinDur: 0.35,
			wantMaxDur: 0.45,
			wantMinInt: 2.5,
		},
		{
			name:       "weakpoint",
			hitType:    HitWeakpoint,
			damage:     50,
			wantActive: true,
			wantMinDur: 0.25,
			wantMaxDur: 0.35,
			wantMinInt: 1.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent()
			comp.Trigger(tt.hitType, tt.damage, 160, 100)

			if comp.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", comp.Active, tt.wantActive)
			}
			if comp.Duration < tt.wantMinDur || comp.Duration > tt.wantMaxDur {
				t.Errorf("Duration = %f, want between %f and %f", comp.Duration, tt.wantMinDur, tt.wantMaxDur)
			}
			if comp.Intensity < tt.wantMinInt {
				t.Errorf("Intensity = %f, want >= %f", comp.Intensity, tt.wantMinInt)
			}
			if comp.Age != 0 {
				t.Errorf("Age = %f, want 0", comp.Age)
			}
			if comp.ScreenX != 160 {
				t.Errorf("ScreenX = %f, want 160", comp.ScreenX)
			}
			if comp.ScreenY != 100 {
				t.Errorf("ScreenY = %f, want 100", comp.ScreenY)
			}
		})
	}
}

func TestReset(t *testing.T) {
	comp := NewComponent()
	comp.Trigger(HitKill, 200, 160, 100)

	// Verify triggered state
	if !comp.Active {
		t.Error("Component should be active after trigger")
	}

	comp.Reset()

	if comp.Active {
		t.Error("Component should not be active after reset")
	}
	if comp.Age != 0 {
		t.Errorf("Age = %f, want 0 after reset", comp.Age)
	}
	if comp.Scale != 1.0 {
		t.Errorf("Scale = %f, want 1.0 after reset", comp.Scale)
	}
	if comp.Alpha != 1.0 {
		t.Errorf("Alpha = %f, want 1.0 after reset", comp.Alpha)
	}
	if comp.Intensity != 1.0 {
		t.Errorf("Intensity = %f, want 1.0 after reset", comp.Intensity)
	}
}

func TestDamageIntensityScaling(t *testing.T) {
	tests := []struct {
		damage     int
		wantMinInt float64
	}{
		{10, 1.0},
		{50, 1.0},
		{51, 1.5},
		{100, 1.5},
		{101, 2.0},
		{200, 2.0},
	}

	for _, tt := range tests {
		comp := NewComponent()
		comp.Trigger(HitNormal, tt.damage, 0, 0)

		if comp.Intensity < tt.wantMinInt {
			t.Errorf("damage=%d: Intensity = %f, want >= %f", tt.damage, comp.Intensity, tt.wantMinInt)
		}
	}
}

func TestHitTypeConstants(t *testing.T) {
	// Verify hit types are distinct
	types := []HitType{HitNormal, HitCritical, HitKill, HitHeadshot, HitWeakpoint}
	seen := make(map[HitType]bool)

	for _, ht := range types {
		if seen[ht] {
			t.Errorf("HitType %v is duplicated", ht)
		}
		seen[ht] = true
	}
}

func TestColorAssignment(t *testing.T) {
	comp := NewComponent()

	// Default color should be white
	if comp.Color != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("Default color = %v, want white", comp.Color)
	}

	// Secondary color should be slightly darker
	if comp.SecondaryColor.R >= comp.Color.R {
		t.Error("SecondaryColor should be darker than primary")
	}
}

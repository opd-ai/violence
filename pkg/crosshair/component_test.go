package crosshair

import (
	"math"
	"testing"
)

func TestNewComponent(t *testing.T) {
	comp := NewComponent()

	if comp.WeaponType != "melee" {
		t.Errorf("Expected default weapon type 'melee', got '%s'", comp.WeaponType)
	}

	if !comp.Visible {
		t.Error("Expected crosshair to be visible by default")
	}

	if comp.Range != 3.0 {
		t.Errorf("Expected default range 3.0, got %f", comp.Range)
	}

	if comp.Scale != 1.0 {
		t.Errorf("Expected default scale 1.0, got %f", comp.Scale)
	}

	if comp.ColorA != 0.8 {
		t.Errorf("Expected default alpha 0.8, got %f", comp.ColorA)
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent()
	expected := "crosshair.Component"

	if comp.Type() != expected {
		t.Errorf("Expected type '%s', got '%s'", expected, comp.Type())
	}
}

func TestAimDirection(t *testing.T) {
	tests := []struct {
		name     string
		aimX     float64
		aimY     float64
		wantNorm bool
	}{
		{
			name:     "normalized east",
			aimX:     1.0,
			aimY:     0.0,
			wantNorm: true,
		},
		{
			name:     "normalized north",
			aimX:     0.0,
			aimY:     -1.0,
			wantNorm: true,
		},
		{
			name:     "normalized diagonal",
			aimX:     math.Sqrt(0.5),
			aimY:     math.Sqrt(0.5),
			wantNorm: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent()
			comp.AimX = tt.aimX
			comp.AimY = tt.aimY

			magnitude := math.Sqrt(comp.AimX*comp.AimX + comp.AimY*comp.AimY)
			isNormalized := math.Abs(magnitude-1.0) < 0.01

			if tt.wantNorm && !isNormalized {
				t.Errorf("Expected normalized vector, got magnitude %f", magnitude)
			}
		})
	}
}

func TestWeaponTypes(t *testing.T) {
	weaponTypes := []string{"melee", "ranged", "magic"}

	for _, wt := range weaponTypes {
		t.Run(wt, func(t *testing.T) {
			comp := NewComponent()
			comp.WeaponType = wt

			if comp.WeaponType != wt {
				t.Errorf("Expected weapon type '%s', got '%s'", wt, comp.WeaponType)
			}
		})
	}
}

func TestVisibilityToggle(t *testing.T) {
	comp := NewComponent()

	if !comp.Visible {
		t.Error("Expected initial visibility to be true")
	}

	comp.Visible = false
	if comp.Visible {
		t.Error("Expected visibility to be false after toggle")
	}
}

func TestColorValues(t *testing.T) {
	comp := NewComponent()

	// Test setting custom color
	comp.ColorR = 1.0
	comp.ColorG = 0.5
	comp.ColorB = 0.0
	comp.ColorA = 1.0

	if comp.ColorR != 1.0 || comp.ColorG != 0.5 || comp.ColorB != 0.0 || comp.ColorA != 1.0 {
		t.Errorf("Color values not set correctly: R=%f G=%f B=%f A=%f",
			comp.ColorR, comp.ColorG, comp.ColorB, comp.ColorA)
	}
}

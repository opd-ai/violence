package flicker

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	comp := NewComponent("torch", 12345)

	if comp.LightType != "torch" {
		t.Errorf("LightType = %q, want 'torch'", comp.LightType)
	}
	if comp.Seed != 12345 {
		t.Errorf("Seed = %d, want 12345", comp.Seed)
	}
	if !comp.Enabled {
		t.Error("Enabled = false, want true")
	}
	if comp.CurrentIntensity != 1.0 {
		t.Errorf("CurrentIntensity = %f, want 1.0", comp.CurrentIntensity)
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent("torch", 12345)
	if comp.Type() != "flicker" {
		t.Errorf("Type() = %q, want 'flicker'", comp.Type())
	}
}

func TestComponentInitialize(t *testing.T) {
	sys := NewSystem("fantasy")
	comp := NewComponent("torch", 12345)

	comp.Initialize(sys, 1.0, 0.6, 0.2)

	if comp.Params.LightType != "torch" {
		t.Errorf("Params.LightType = %q, want 'torch'", comp.Params.LightType)
	}
	if comp.Params.Seed != 12345 {
		t.Errorf("Params.Seed = %d, want 12345", comp.Params.Seed)
	}
	if comp.Params.BaseR != 1.0 || comp.Params.BaseG != 0.6 || comp.Params.BaseB != 0.2 {
		t.Error("Params base colors not set correctly")
	}
}

func TestComponentSetEnabled(t *testing.T) {
	comp := NewComponent("torch", 12345)

	if !comp.Enabled {
		t.Error("Initially Enabled should be true")
	}

	comp.SetEnabled(false)
	if comp.Enabled {
		t.Error("After SetEnabled(false), Enabled should be false")
	}

	comp.SetEnabled(true)
	if !comp.Enabled {
		t.Error("After SetEnabled(true), Enabled should be true")
	}
}

func TestComponentGetIntensityMultiplier(t *testing.T) {
	comp := NewComponent("torch", 12345)

	// When enabled, returns CurrentIntensity
	comp.CurrentIntensity = 0.8
	if got := comp.GetIntensityMultiplier(); got != 0.8 {
		t.Errorf("GetIntensityMultiplier() = %f, want 0.8", got)
	}

	// When disabled, returns 1.0
	comp.SetEnabled(false)
	if got := comp.GetIntensityMultiplier(); got != 1.0 {
		t.Errorf("GetIntensityMultiplier() when disabled = %f, want 1.0", got)
	}
}

func TestComponentGetColorModulation(t *testing.T) {
	comp := NewComponent("torch", 12345)

	// When enabled, returns current colors
	comp.CurrentR = 0.9
	comp.CurrentG = 0.5
	comp.CurrentB = 0.2
	r, g, b := comp.GetColorModulation()
	if r != 0.9 || g != 0.5 || b != 0.2 {
		t.Errorf("GetColorModulation() = (%f, %f, %f), want (0.9, 0.5, 0.2)", r, g, b)
	}

	// When disabled, returns 1.0 for all
	comp.SetEnabled(false)
	r, g, b = comp.GetColorModulation()
	if r != 1.0 || g != 1.0 || b != 1.0 {
		t.Errorf("GetColorModulation() when disabled = (%f, %f, %f), want (1.0, 1.0, 1.0)", r, g, b)
	}
}

func TestFlickerParamsType(t *testing.T) {
	params := FlickerParams{}
	if params.Type() != "flicker_params" {
		t.Errorf("FlickerParams.Type() = %q, want 'flicker_params'", params.Type())
	}
}

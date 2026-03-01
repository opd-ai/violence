package lighting

import (
	"testing"
)

func TestNewLightComponent(t *testing.T) {
	preset := LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.6,
		B:         0.2,
		Flicker:   true,
	}

	lc := NewLightComponent(preset, 42)
	if lc == nil {
		t.Fatal("NewLightComponent returned nil")
	}

	if lc.Type() != "Light" {
		t.Errorf("expected Type() = 'Light', got %s", lc.Type())
	}

	if !lc.Enabled {
		t.Error("light should be enabled by default")
	}

	if lc.Radius != 5.0 {
		t.Errorf("expected Radius = 5.0, got %f", lc.Radius)
	}

	if lc.Intensity != 0.8 {
		t.Errorf("expected Intensity = 0.8, got %f", lc.Intensity)
	}

	if lc.Lifetime != 0 {
		t.Errorf("expected Lifetime = 0 (infinite), got %f", lc.Lifetime)
	}
}

func TestNewAttachedLight(t *testing.T) {
	preset := LightPreset{
		Name:      "torch",
		Radius:    5.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.6,
		B:         0.2,
		Flicker:   true,
	}

	lc := NewAttachedLight(preset, 42, 1.0, 2.0)
	if lc == nil {
		t.Fatal("NewAttachedLight returned nil")
	}

	if !lc.AttachedToEntity {
		t.Error("attached light should have AttachedToEntity = true")
	}

	if lc.OffsetX != 1.0 {
		t.Errorf("expected OffsetX = 1.0, got %f", lc.OffsetX)
	}

	if lc.OffsetY != 2.0 {
		t.Errorf("expected OffsetY = 2.0, got %f", lc.OffsetY)
	}
}

func TestNewTemporaryLight(t *testing.T) {
	preset := LightPreset{
		Name:      "explosion",
		Radius:    10.0,
		Intensity: 1.0,
		R:         1.0,
		G:         0.5,
		B:         0.1,
		Flicker:   false,
	}

	lifetime := 2.0
	lc := NewTemporaryLight(preset, 42, lifetime)
	if lc == nil {
		t.Fatal("NewTemporaryLight returned nil")
	}

	if lc.Lifetime != lifetime {
		t.Errorf("expected Lifetime = %f, got %f", lifetime, lc.Lifetime)
	}

	if lc.FadeInDuration == 0 {
		t.Error("temporary light should have fade in duration")
	}

	if lc.FadeOutDuration == 0 {
		t.Error("temporary light should have fade out duration")
	}
}

func TestNewPulsingLight(t *testing.T) {
	preset := LightPreset{
		Name:      "alarm",
		Radius:    4.0,
		Intensity: 0.8,
		R:         1.0,
		G:         0.2,
		B:         0.2,
		Flicker:   false,
	}

	lc := NewPulsingLight(preset, 42, 2.0)
	if lc == nil {
		t.Fatal("NewPulsingLight returned nil")
	}

	if !lc.Pulsing {
		t.Error("pulsing light should have Pulsing = true")
	}

	if lc.PulseSpeed != 2.0 {
		t.Errorf("expected PulseSpeed = 2.0, got %f", lc.PulseSpeed)
	}
}

func TestLightComponentType(t *testing.T) {
	lc := &LightComponent{}
	if lc.Type() != "Light" {
		t.Errorf("expected Type() = 'Light', got %s", lc.Type())
	}
}

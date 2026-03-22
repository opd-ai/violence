package muzzleflash

import (
	"image/color"
	"testing"
)

func TestComponentType(t *testing.T) {
	c := NewComponent()
	expected := "muzzleflash.Component"
	if c.Type() != expected {
		t.Errorf("Type() = %q, want %q", c.Type(), expected)
	}
}

func TestNewComponent(t *testing.T) {
	c := NewComponent()

	if c.ActiveFlashes == nil {
		t.Error("ActiveFlashes should be initialized")
	}
	if len(c.ActiveFlashes) != 0 {
		t.Error("ActiveFlashes should start empty")
	}
	if c.MaxFlashes != 8 {
		t.Errorf("MaxFlashes = %d, want 8", c.MaxFlashes)
	}
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		flashType string
		wantRays  int
	}{
		{"bullet", 5},
		{"plasma", 0},
		{"energy", 4},
		{"fire", 7},
		{"magic", 6},
		{"shotgun", 8},
		{"laser", 0},
		{"unknown", 5}, // Should default to bullet
	}

	for _, tt := range tests {
		t.Run(tt.flashType, func(t *testing.T) {
			profile := GetProfile(tt.flashType)
			if profile.RayCount != tt.wantRays {
				t.Errorf("GetProfile(%q).RayCount = %d, want %d", tt.flashType, profile.RayCount, tt.wantRays)
			}
			if profile.Duration <= 0 {
				t.Errorf("GetProfile(%q).Duration = %f, want > 0", tt.flashType, profile.Duration)
			}
			if profile.BaseSize <= 0 {
				t.Errorf("GetProfile(%q).BaseSize = %f, want > 0", tt.flashType, profile.BaseSize)
			}
		})
	}
}

func TestFlashProfileColors(t *testing.T) {
	tests := []struct {
		flashType string
	}{
		{"bullet"},
		{"plasma"},
		{"energy"},
		{"fire"},
		{"magic"},
		{"shotgun"},
		{"laser"},
	}

	for _, tt := range tests {
		t.Run(tt.flashType, func(t *testing.T) {
			profile := GetProfile(tt.flashType)

			// Primary color should have visible alpha
			if profile.PrimaryColor.A == 0 {
				t.Errorf("GetProfile(%q).PrimaryColor.A = 0, want > 0", tt.flashType)
			}

			// Secondary color should have visible alpha
			if profile.SecondaryColor.A == 0 {
				t.Errorf("GetProfile(%q).SecondaryColor.A = 0, want > 0", tt.flashType)
			}

			// Core brightness should be positive
			if profile.CoreBrightness <= 0 {
				t.Errorf("GetProfile(%q).CoreBrightness = %f, want > 0", tt.flashType, profile.CoreBrightness)
			}
		})
	}
}

func TestDefaultProfilesComplete(t *testing.T) {
	expectedTypes := []string{"bullet", "plasma", "energy", "fire", "magic", "shotgun", "laser"}

	for _, flashType := range expectedTypes {
		t.Run(flashType, func(t *testing.T) {
			profile, ok := DefaultProfiles[flashType]
			if !ok {
				t.Errorf("DefaultProfiles missing %q", flashType)
				return
			}

			if profile.Duration <= 0 {
				t.Errorf("Profile %q has invalid Duration: %f", flashType, profile.Duration)
			}
			if profile.BaseSize <= 0 {
				t.Errorf("Profile %q has invalid BaseSize: %f", flashType, profile.BaseSize)
			}
			if profile.LightRadius <= 0 {
				t.Errorf("Profile %q has invalid LightRadius: %f", flashType, profile.LightRadius)
			}
		})
	}
}

func TestFlashStruct(t *testing.T) {
	flash := &Flash{
		X:              10.0,
		Y:              20.0,
		Angle:          1.5,
		Age:            0.02,
		Duration:       0.06,
		FlashType:      "bullet",
		Intensity:      1.2,
		PrimaryColor:   color.RGBA{R: 255, G: 220, B: 150, A: 255},
		SecondaryColor: color.RGBA{R: 255, G: 180, B: 80, A: 200},
		Scale:          1.0,
		EmitsLight:     true,
		LightIntensity: 1.5,
		LightRadius:    3.0,
	}

	if flash.X != 10.0 {
		t.Errorf("Flash.X = %f, want 10.0", flash.X)
	}
	if flash.Y != 20.0 {
		t.Errorf("Flash.Y = %f, want 20.0", flash.Y)
	}
	if flash.FlashType != "bullet" {
		t.Errorf("Flash.FlashType = %q, want %q", flash.FlashType, "bullet")
	}
	if !flash.EmitsLight {
		t.Error("Flash.EmitsLight should be true")
	}
}

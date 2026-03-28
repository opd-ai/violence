package reloadbar

import (
	"testing"
)

func TestNewComponent(t *testing.T) {
	c := NewComponent()

	if c == nil {
		t.Fatal("NewComponent returned nil")
	}

	if c.IsReloading {
		t.Error("New component should not be reloading")
	}

	if c.Progress != 0.0 {
		t.Errorf("Expected progress 0.0, got %f", c.Progress)
	}

	if c.FadeAlpha != 0.0 {
		t.Errorf("Expected FadeAlpha 0.0, got %f", c.FadeAlpha)
	}
}

func TestComponentType(t *testing.T) {
	c := NewComponent()

	if c.Type() != "reloadbar" {
		t.Errorf("Expected type 'reloadbar', got '%s'", c.Type())
	}
}

func TestStartReload(t *testing.T) {
	c := NewComponent()

	c.StartReload(2.0, "Shotgun", 8)

	if !c.IsReloading {
		t.Error("Component should be reloading after StartReload")
	}

	if c.TotalDuration != 2.0 {
		t.Errorf("Expected duration 2.0, got %f", c.TotalDuration)
	}

	if c.WeaponName != "Shotgun" {
		t.Errorf("Expected weapon name 'Shotgun', got '%s'", c.WeaponName)
	}

	if c.ClipSize != 8 {
		t.Errorf("Expected clip size 8, got %d", c.ClipSize)
	}

	if c.Progress != 0.0 {
		t.Errorf("Expected initial progress 0.0, got %f", c.Progress)
	}
}

func TestUpdateProgress(t *testing.T) {
	tests := []struct {
		name             string
		totalDuration    float64
		deltaTime        float64
		updateCount      int
		expectedProgress float64
		expectedComplete bool
	}{
		{
			name:             "half progress",
			totalDuration:    1.0,
			deltaTime:        0.5,
			updateCount:      1,
			expectedProgress: 0.5,
			expectedComplete: false,
		},
		{
			name:             "full progress",
			totalDuration:    1.0,
			deltaTime:        0.25,
			updateCount:      4,
			expectedProgress: 1.0,
			expectedComplete: true,
		},
		{
			name:             "over progress clamped",
			totalDuration:    1.0,
			deltaTime:        2.0,
			updateCount:      1,
			expectedProgress: 1.0,
			expectedComplete: true,
		},
		{
			name:             "quarter progress",
			totalDuration:    2.0,
			deltaTime:        0.5,
			updateCount:      1,
			expectedProgress: 0.25,
			expectedComplete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewComponent()
			c.StartReload(tt.totalDuration, "TestWeapon", 10)

			for i := 0; i < tt.updateCount; i++ {
				c.UpdateProgress(tt.deltaTime)
			}

			if c.Progress != tt.expectedProgress {
				t.Errorf("Expected progress %f, got %f", tt.expectedProgress, c.Progress)
			}

			if c.IsReloading == tt.expectedComplete {
				t.Errorf("Expected complete=%v, got IsReloading=%v", tt.expectedComplete, c.IsReloading)
			}
		})
	}
}

func TestCancelReload(t *testing.T) {
	c := NewComponent()
	c.StartReload(2.0, "Pistol", 12)
	c.UpdateProgress(0.5)

	if c.Progress == 0.0 {
		t.Error("Progress should be non-zero after update")
	}

	c.CancelReload()

	if c.IsReloading {
		t.Error("Should not be reloading after cancel")
	}

	if c.Progress != 0.0 {
		t.Errorf("Progress should be reset to 0.0, got %f", c.Progress)
	}

	if c.ElapsedTime != 0.0 {
		t.Errorf("ElapsedTime should be reset to 0.0, got %f", c.ElapsedTime)
	}
}

func TestCompleteReload(t *testing.T) {
	c := NewComponent()
	c.StartReload(1.0, "Rifle", 30)
	c.CompleteReload()

	if c.IsReloading {
		t.Error("Should not be reloading after complete")
	}

	if c.Progress != 1.0 {
		t.Errorf("Progress should be 1.0 after complete, got %f", c.Progress)
	}

	if c.AmmoCount != 30 {
		t.Errorf("AmmoCount should equal ClipSize (30), got %d", c.AmmoCount)
	}
}

func TestUpdateProgressNotReloading(t *testing.T) {
	c := NewComponent()

	// Should not panic or change state when not reloading
	c.UpdateProgress(0.5)

	if c.Progress != 0.0 {
		t.Errorf("Progress should remain 0.0 when not reloading, got %f", c.Progress)
	}
}

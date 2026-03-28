package focusring

import (
	"testing"
)

func TestFocusableElement_DefaultValues(t *testing.T) {
	elem := &FocusableElement{
		ID: "test",
	}

	if elem.ID != "test" {
		t.Error("ID should be set")
	}
	if elem.X != 0 || elem.Y != 0 {
		t.Error("Default position should be (0, 0)")
	}
	if elem.Width != 0 || elem.Height != 0 {
		t.Error("Default size should be (0, 0)")
	}
	if elem.TabIndex != 0 {
		t.Error("Default TabIndex should be 0")
	}
	if elem.Enabled != false {
		t.Error("Default Enabled should be false (AddFocusable enables it)")
	}
}

func TestFocusableElement_Callbacks(t *testing.T) {
	focusCalled := false
	blurCalled := false
	activateCalled := false

	elem := &FocusableElement{
		ID:         "test",
		OnFocus:    func() { focusCalled = true },
		OnBlur:     func() { blurCalled = true },
		OnActivate: func() { activateCalled = true },
	}

	elem.OnFocus()
	if !focusCalled {
		t.Error("OnFocus callback should be callable")
	}

	elem.OnBlur()
	if !blurCalled {
		t.Error("OnBlur callback should be callable")
	}

	elem.OnActivate()
	if !activateCalled {
		t.Error("OnActivate callback should be callable")
	}
}

func TestFocusRingConfig_Fields(t *testing.T) {
	config := FocusRingConfig{
		RingThickness:   3.0,
		GlowRadius:      8.0,
		CornerRadius:    6.0,
		PulseSpeed:      0.1,
		PulseIntensity:  0.5,
		TransitionSpeed: 0.2,
	}

	if config.RingThickness != 3.0 {
		t.Error("RingThickness should be set")
	}
	if config.GlowRadius != 8.0 {
		t.Error("GlowRadius should be set")
	}
	if config.CornerRadius != 6.0 {
		t.Error("CornerRadius should be set")
	}
	if config.PulseSpeed != 0.1 {
		t.Error("PulseSpeed should be set")
	}
	if config.PulseIntensity != 0.5 {
		t.Error("PulseIntensity should be set")
	}
	if config.TransitionSpeed != 0.2 {
		t.Error("TransitionSpeed should be set")
	}
}

func TestFocusState_Fields(t *testing.T) {
	state := FocusState{
		FocusedID:          "btn1",
		CurrentX:           10.0,
		CurrentY:           20.0,
		CurrentW:           100.0,
		CurrentH:           40.0,
		TargetX:            15.0,
		TargetY:            25.0,
		TargetW:            110.0,
		TargetH:            45.0,
		PulsePhase:         1.5,
		TransitionProgress: 0.75,
		Visible:            true,
	}

	if state.FocusedID != "btn1" {
		t.Error("FocusedID should be set")
	}
	if state.CurrentX != 10.0 || state.CurrentY != 20.0 {
		t.Error("Current position should be set")
	}
	if state.TargetX != 15.0 || state.TargetY != 25.0 {
		t.Error("Target position should be set")
	}
	if !state.Visible {
		t.Error("Visible should be true")
	}
}

func TestGenrePreset_AllFields(t *testing.T) {
	preset := GenrePreset{
		Name:           "test_genre",
		PulseSpeed:     0.15,
		PulseIntensity: 0.4,
	}

	if preset.Name != "test_genre" {
		t.Error("Name should be set")
	}
	if preset.PulseSpeed != 0.15 {
		t.Error("PulseSpeed should be set")
	}
	if preset.PulseIntensity != 0.4 {
		t.Error("PulseIntensity should be set")
	}
}

func TestDefaultGenrePresets_Colors(t *testing.T) {
	presets := DefaultGenrePresets()

	// Fantasy should be gold
	fantasy := presets["fantasy"]
	if fantasy.RingColor.R != 255 || fantasy.RingColor.G != 215 || fantasy.RingColor.B != 0 {
		t.Error("Fantasy ring color should be gold")
	}

	// Scifi should be cyan
	scifi := presets["scifi"]
	if scifi.RingColor.R != 0 || scifi.RingColor.G != 255 || scifi.RingColor.B != 255 {
		t.Error("Scifi ring color should be cyan")
	}

	// Horror should be blood red
	horror := presets["horror"]
	if horror.RingColor.R != 180 || horror.RingColor.G != 30 || horror.RingColor.B != 30 {
		t.Error("Horror ring color should be blood red")
	}

	// Cyberpunk should be magenta
	cyberpunk := presets["cyberpunk"]
	if cyberpunk.RingColor.R != 255 || cyberpunk.RingColor.G != 0 || cyberpunk.RingColor.B != 255 {
		t.Error("Cyberpunk ring color should be magenta")
	}

	// Postapoc should be orange
	postapoc := presets["postapoc"]
	if postapoc.RingColor.R != 255 || postapoc.RingColor.G != 140 || postapoc.RingColor.B != 0 {
		t.Error("Postapoc ring color should be orange")
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	config := DefaultConfig()

	// Verify sensible defaults
	if config.RingThickness < 1.0 || config.RingThickness > 5.0 {
		t.Errorf("RingThickness %f outside reasonable range", config.RingThickness)
	}
	if config.GlowRadius < 2.0 || config.GlowRadius > 20.0 {
		t.Errorf("GlowRadius %f outside reasonable range", config.GlowRadius)
	}
	if config.PulseSpeed <= 0 || config.PulseSpeed > 1.0 {
		t.Errorf("PulseSpeed %f outside reasonable range", config.PulseSpeed)
	}
	if config.TransitionSpeed <= 0 || config.TransitionSpeed > 1.0 {
		t.Errorf("TransitionSpeed %f outside reasonable range", config.TransitionSpeed)
	}
}

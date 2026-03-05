package camerafx

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID, 12345)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genre != tt.genreID {
				t.Errorf("expected genre %s, got %s", tt.genreID, sys.genre)
			}
			if sys.component == nil {
				t.Fatal("component not initialized")
			}
		})
	}
}

func TestShakeTrigger(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.TriggerShake(Shake.Medium())

	if sys.component.ShakeIntensity == 0 {
		t.Error("shake intensity should be non-zero after trigger")
	}

	if sys.component.ShakeDecay == 0 {
		t.Error("shake decay should be non-zero after trigger")
	}
}

func TestShakeUpdate(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.TriggerShake(Shake.Light())
	initialIntensity := sys.component.ShakeIntensity

	sys.Update(0.1)

	if sys.component.ShakeIntensity >= initialIntensity {
		t.Error("shake intensity should decay over time")
	}
}

func TestShakeDecay(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.TriggerShake(Shake.Tiny())

	for i := 0; i < 100; i++ {
		sys.Update(0.1)
	}

	if sys.component.ShakeIntensity != 0 {
		t.Error("shake should fully decay to zero")
	}

	if sys.component.ShakeOffsetX != 0 || sys.component.ShakeOffsetY != 0 {
		t.Error("shake offset should be zero when intensity is zero")
	}
}

func TestFlashTrigger(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	r, g, b, a := Flash.Red()
	sys.TriggerFlash(r, g, b, a)

	if sys.component.FlashAlpha == 0 {
		t.Error("flash alpha should be non-zero after trigger")
	}

	if sys.component.FlashR != r || sys.component.FlashG != g || sys.component.FlashB != b {
		t.Error("flash color not set correctly")
	}
}

func TestFlashDecay(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	r, g, b, a := Flash.White()
	sys.TriggerFlash(r, g, b, a)
	initialAlpha := sys.component.FlashAlpha

	sys.Update(0.1)

	if sys.component.FlashAlpha >= initialAlpha {
		t.Error("flash should decay over time")
	}
}

func TestZoom(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.SetZoom(1.5)

	if sys.component.ZoomTarget != 1.5 {
		t.Errorf("expected zoom target 1.5, got %f", sys.component.ZoomTarget)
	}

	for i := 0; i < 10; i++ {
		sys.Update(0.1)
	}

	if sys.component.ZoomCurrent < 1.4 {
		t.Errorf("zoom should converge to target, got %f", sys.component.ZoomCurrent)
	}
}

func TestZoomClamping(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.SetZoom(10.0)
	if sys.component.ZoomTarget > 2.0 {
		t.Error("zoom should be clamped to maximum 2.0")
	}

	sys.SetZoom(0.1)
	if sys.component.ZoomTarget < 0.5 {
		t.Error("zoom should be clamped to minimum 0.5")
	}
}

func TestChromaticAberration(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.TriggerChromatic(0.5)

	if sys.component.ChromaticAberr == 0 {
		t.Error("chromatic aberration should be non-zero after trigger")
	}

	initialChroma := sys.component.ChromaticAberr

	sys.Update(0.1)

	if sys.component.ChromaticAberr >= initialChroma {
		t.Error("chromatic aberration should decay over time")
	}
}

func TestGenreScaling(t *testing.T) {
	tests := []struct {
		genre      string
		expectMore string
		expectLess string
	}{
		{"horror", "shake", "flash"},
		{"cyberpunk", "flash", "shake"},
		{"scifi", "chroma", "shake"},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewSystem(tt.genre, 12345)

			intensity := 5.0
			sys.TriggerShake(intensity)
			shake := sys.component.ShakeIntensity

			r, g, b, a := Flash.White()
			sys.TriggerFlash(r, g, b, a)
			flash := sys.component.FlashAlpha

			sys.TriggerChromatic(0.5)
			chroma := sys.component.ChromaticAberr

			if shake < 0 || flash < 0 || chroma < 0 {
				t.Error("effect values should be non-negative")
			}
		})
	}
}

func TestGetters(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	sys.TriggerShake(Shake.Medium())
	sys.Update(0.016) // Update to generate offset
	x, y := sys.GetShakeOffset()
	if x == 0 && y == 0 {
		t.Error("shake offset should be non-zero after update")
	}

	r, g, b, a := Flash.Red()
	sys.TriggerFlash(r, g, b, a)
	fr, fg, fb, fa := sys.GetFlashColor()
	if fa == 0 {
		t.Error("flash alpha should be non-zero")
	}
	if fr != r || fg != g || fb != b {
		t.Error("flash color mismatch")
	}

	sys.SetZoom(1.3)
	zoom := sys.GetZoom()
	if zoom < 0.5 || zoom > 2.0 {
		t.Error("zoom out of valid range")
	}

	sys.TriggerChromatic(0.4)
	chroma := sys.GetChromaticAberration()
	if chroma == 0 {
		t.Error("chromatic aberration should be non-zero")
	}

	vignette := sys.GetVignette()
	if vignette < 0 || vignette > 1 {
		t.Error("vignette out of valid range")
	}
}

func TestPresets(t *testing.T) {
	shakeValues := []float64{
		Shake.Tiny(),
		Shake.Light(),
		Shake.Medium(),
		Shake.Heavy(),
		Shake.Massive(),
		Shake.Cataclysmic(),
	}

	for i := 0; i < len(shakeValues)-1; i++ {
		if shakeValues[i] >= shakeValues[i+1] {
			t.Error("shake presets should be in ascending order")
		}
	}

	flashColors := []struct {
		name string
		fn   func() (float64, float64, float64, float64)
	}{
		{"white", Flash.White},
		{"red", Flash.Red},
		{"orange", Flash.Orange},
		{"blue", Flash.Blue},
		{"green", Flash.Green},
		{"purple", Flash.Purple},
		{"yellow", Flash.Yellow},
	}

	for _, fc := range flashColors {
		r, g, b, a := fc.fn()
		if r < 0 || r > 1 || g < 0 || g > 1 || b < 0 || b > 1 || a < 0 || a > 1 {
			t.Errorf("flash color %s has invalid component values", fc.name)
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem("fantasy", 12345)
	sys.TriggerShake(Shake.Medium())
	r, g, bl, a := Flash.White()
	sys.TriggerFlash(r, g, bl, a)
	sys.TriggerChromatic(0.5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(0.016)
	}
}

func BenchmarkTriggerShake(b *testing.B) {
	sys := NewSystem("fantasy", 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.TriggerShake(Shake.Medium())
	}
}

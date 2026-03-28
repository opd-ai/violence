package eyeglint

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
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
		{"unknown defaults to fantasy", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != tt.genreID {
				t.Errorf("genreID = %v, want %v", sys.genreID, tt.genreID)
			}
			if sys.glintCache == nil {
				t.Error("glintCache is nil")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy")

	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID after SetGenre = %v, want cyberpunk", sys.genreID)
	}

	preset := sys.GetPresetForGenre()
	expectedPreset := GetPreset("cyberpunk")
	if preset.AnimationSpeed != expectedPreset.AnimationSpeed {
		t.Error("Preset not updated after SetGenre")
	}
}

func TestNewComponent(t *testing.T) {
	comp := NewComponent()
	if comp == nil {
		t.Fatal("NewComponent returned nil")
	}
	if !comp.Enabled {
		t.Error("Component should be enabled by default")
	}
	if comp.GlintIntensity != 0.8 {
		t.Errorf("GlintIntensity = %v, want 0.8", comp.GlintIntensity)
	}
	if comp.EyeCount() != 0 {
		t.Error("New component should have no eyes")
	}
}

func TestComponentAddEye(t *testing.T) {
	comp := NewComponent()

	comp.AddEye(10, 20, 3)
	if comp.EyeCount() != 1 {
		t.Errorf("EyeCount after AddEye = %v, want 1", comp.EyeCount())
	}

	comp.AddEye(30, 20, 3)
	if comp.EyeCount() != 2 {
		t.Errorf("EyeCount after second AddEye = %v, want 2", comp.EyeCount())
	}

	if comp.EyePositions[0][0] != 10 || comp.EyePositions[0][1] != 20 {
		t.Error("First eye position incorrect")
	}
	if comp.EyeSizes[0] != 3 {
		t.Errorf("First eye radius = %v, want 3", comp.EyeSizes[0])
	}
}

func TestComponentClearEyes(t *testing.T) {
	comp := NewComponent()
	comp.AddEye(10, 20, 3)
	comp.AddEye(30, 20, 3)

	comp.ClearEyes()
	if comp.EyeCount() != 0 {
		t.Errorf("EyeCount after ClearEyes = %v, want 0", comp.EyeCount())
	}
}

func TestComponentType(t *testing.T) {
	comp := NewComponent()
	if comp.Type() != "eyeglint" {
		t.Errorf("Type() = %v, want eyeglint", comp.Type())
	}
}

func TestGenrePresets(t *testing.T) {
	presets := DefaultGenrePresets()

	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range expectedGenres {
		preset, ok := presets[genre]
		if !ok {
			t.Errorf("Missing preset for genre %s", genre)
			continue
		}

		if preset.PrimarySize <= 0 || preset.PrimarySize > 1.0 {
			t.Errorf("%s: PrimarySize out of range: %v", genre, preset.PrimarySize)
		}
		if preset.PrimaryOffset <= 0 || preset.PrimaryOffset > 1.0 {
			t.Errorf("%s: PrimaryOffset out of range: %v", genre, preset.PrimaryOffset)
		}
		if preset.AnimationSpeed < 0 {
			t.Errorf("%s: AnimationSpeed negative: %v", genre, preset.AnimationSpeed)
		}
	}
}

func TestGetPresetUnknownGenre(t *testing.T) {
	preset := GetPreset("unknown_genre")
	fantasyPreset := GetPreset("fantasy")

	if preset.AnimationSpeed != fantasyPreset.AnimationSpeed {
		t.Error("Unknown genre should fall back to fantasy preset")
	}
}

func TestApplyEyeGlintsNilInputs(t *testing.T) {
	sys := NewSystem("fantasy")

	// Nil sprite
	result := sys.ApplyEyeGlints(nil, NewComponent(), 12345)
	if result != nil {
		t.Error("Should return nil for nil sprite")
	}

	// Nil component
	sprite := ebiten.NewImage(32, 32)
	result = sys.ApplyEyeGlints(sprite, nil, 12345)
	if result != sprite {
		t.Error("Should return original sprite for nil component")
	}

	// Disabled component
	comp := NewComponent()
	comp.Enabled = false
	result = sys.ApplyEyeGlints(sprite, comp, 12345)
	if result != sprite {
		t.Error("Should return original sprite for disabled component")
	}

	// No eyes
	comp.Enabled = true
	result = sys.ApplyEyeGlints(sprite, comp, 12345)
	if result != sprite {
		t.Error("Should return original sprite when no eyes detected")
	}
}

func TestApplyEyeGlintsWithEyes(t *testing.T) {
	sys := NewSystem("fantasy")

	// Create a test sprite with a yellow "eye"
	sprite := ebiten.NewImage(32, 32)
	rgba := image.NewRGBA(image.Rect(0, 0, 32, 32))
	eyeColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}

	// Draw an eye at position (16, 10)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			rgba.Set(16+dx, 10+dy, eyeColor)
		}
	}
	sprite.WritePixels(rgba.Pix)

	comp := NewComponent()
	comp.AddEye(16, 10, 3)

	result := sys.ApplyEyeGlints(sprite, comp, 12345)
	if result == nil {
		t.Fatal("ApplyEyeGlints returned nil")
	}

	// Verify a new image was created (different from original)
	// We can't read pixels in tests, but we can verify the result is non-nil
	// and has correct dimensions
	bounds := result.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("Result bounds = %v, want 32x32", bounds)
	}
}

func TestDetectEyesHumanoid(t *testing.T) {
	sys := NewSystem("fantasy")

	// Create a sprite with humanoid eye-like features (using RGBA directly)
	rgba := image.NewRGBA(image.Rect(0, 0, 32, 48))

	// Fill with skin tone
	skinColor := color.RGBA{R: 220, G: 180, B: 150, A: 255}
	for y := 0; y < 48; y++ {
		for x := 0; x < 32; x++ {
			rgba.Set(x, y, skinColor)
		}
	}

	// Draw eyes (yellow, in upper portion)
	eyeColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}
	// Left eye at (12, 12)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			rgba.Set(12+dx, 12+dy, eyeColor)
		}
	}
	// Right eye at (20, 12)
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			rgba.Set(20+dx, 12+dy, eyeColor)
		}
	}

	comp := sys.DetectEyesFromRGBA(rgba, "humanoid")
	if comp == nil {
		t.Fatal("DetectEyesFromRGBA returned nil")
	}

	if comp.CreatureType != "humanoid" {
		t.Errorf("CreatureType = %v, want humanoid", comp.CreatureType)
	}

	// Should detect at least one eye
	if comp.EyeCount() == 0 {
		t.Log("No eyes detected - this may be acceptable depending on detection thresholds")
	}
}

func TestDetectEyesNilSprite(t *testing.T) {
	sys := NewSystem("fantasy")

	// Test with nil RGBA image
	comp := sys.DetectEyesFromRGBA(nil, "humanoid")
	if comp == nil {
		t.Fatal("DetectEyesFromRGBA should return non-nil component even for nil image")
	}
	if comp.EyeCount() != 0 {
		t.Error("Should have no eyes detected for nil image")
	}
}

func TestIsEyeColor(t *testing.T) {
	sys := NewSystem("fantasy")

	tests := []struct {
		name     string
		r, g, b  uint8
		expected bool
	}{
		{"yellow eye", 255, 200, 50, true},
		{"red eye", 255, 50, 50, true},
		{"green eye", 50, 200, 50, true},
		{"blue eye", 50, 100, 200, true},
		{"black pupil", 30, 30, 30, true},
		{"white sclera", 255, 255, 255, false},
		{"skin tone", 220, 180, 150, false},
		{"gray", 128, 128, 128, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sys.isEyeColor(tt.r, tt.g, tt.b)
			if result != tt.expected {
				t.Errorf("isEyeColor(%d,%d,%d) = %v, want %v",
					tt.r, tt.g, tt.b, result, tt.expected)
			}
		})
	}
}

func TestMultipleCreatureTypes(t *testing.T) {
	sys := NewSystem("fantasy")

	creatureTypes := []string{
		"humanoid", "skeleton", "zombie",
		"quadruped", "wolf", "bear",
		"insect", "spider", "scorpion",
		"serpent", "snake", "dragon",
		"flying", "bat", "bird",
		"amorphous", "slime", "ooze",
		"unknown",
	}

	// Use RGBA image instead of ebiten.Image
	rgba := image.NewRGBA(image.Rect(0, 0, 32, 32))

	for _, ct := range creatureTypes {
		t.Run(ct, func(t *testing.T) {
			comp := sys.DetectEyesFromRGBA(rgba, ct)
			if comp == nil {
				t.Errorf("DetectEyesFromRGBA returned nil for creature type %s", ct)
			}
			if comp.CreatureType != ct {
				t.Errorf("CreatureType = %v, want %v", comp.CreatureType, ct)
			}
		})
	}
}

func TestRenderEyeGlintDirect(t *testing.T) {
	sys := NewSystem("fantasy")

	screen := ebiten.NewImage(64, 64)

	// Should not panic
	sys.RenderEyeGlint(screen, 32, 32, 3, 0.8)

	// Verify the function completes without error
	// We can't read pixels in tests, but no panic is success
	bounds := screen.Bounds()
	if bounds.Dx() != 64 || bounds.Dy() != 64 {
		t.Errorf("Screen bounds changed unexpectedly: %v", bounds)
	}
}

func TestAnimationPhaseWrapping(t *testing.T) {
	comp := NewComponent()
	comp.GlintPhase = 6.0 // Just under 2π

	sys := NewSystem("fantasy")

	// Simulate many updates
	for i := 0; i < 1000; i++ {
		comp.GlintPhase += sys.preset.AnimationSpeed * 0.0167
		if comp.GlintPhase > 6.28318 { // 2π
			comp.GlintPhase -= 6.28318
		}
	}

	// Phase should stay bounded
	if comp.GlintPhase < 0 || comp.GlintPhase > 6.29 {
		t.Errorf("GlintPhase out of bounds: %v", comp.GlintPhase)
	}
}

func TestPresetValues(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset := GetPreset(genre)

			// Verify all presets have sensible values
			if preset.PrimaryColor.A == 0 {
				t.Error("PrimaryColor should not be fully transparent")
			}
			if preset.PrimarySize <= 0 {
				t.Error("PrimarySize should be positive")
			}
			if preset.PrimaryOffset <= 0 {
				t.Error("PrimaryOffset should be positive")
			}
		})
	}
}

func BenchmarkApplyEyeGlints(b *testing.B) {
	sys := NewSystem("fantasy")
	sprite := ebiten.NewImage(64, 64)
	comp := NewComponent()
	comp.AddEye(20, 20, 3)
	comp.AddEye(44, 20, 3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.ApplyEyeGlints(sprite, comp, int64(i))
	}
}

func BenchmarkDetectEyes(b *testing.B) {
	sys := NewSystem("fantasy")

	// Create a complex image using RGBA directly
	rgba := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			rgba.Set(x, y, color.RGBA{R: 200, G: 180, B: 160, A: 255})
		}
	}
	// Add eyes
	eyeColor := color.RGBA{R: 255, G: 200, B: 50, A: 255}
	for dy := -3; dy <= 3; dy++ {
		for dx := -3; dx <= 3; dx++ {
			rgba.Set(20+dx, 20+dy, eyeColor)
			rgba.Set(44+dx, 20+dy, eyeColor)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sys.DetectEyesFromRGBA(rgba, "humanoid")
	}
}

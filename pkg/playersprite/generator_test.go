package playersprite

import (
	"image/color"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("fantasy")
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.genreID != "fantasy" {
		t.Errorf("genreID = %v, want fantasy", g.genreID)
	}
	if len(g.templates) == 0 {
		t.Error("templates should be initialized")
	}
}

func TestGenerator_SetGenre(t *testing.T) {
	tests := []struct {
		genre           string
		expectedClasses []string
	}{
		{"fantasy", []string{"warrior", "mage", "rogue"}},
		{"scifi", []string{"soldier", "hacker", "cyborg"}},
		{"cyberpunk", []string{"soldier", "hacker", "cyborg"}},
		{"horror", []string{"survivor", "occultist"}},
		{"postapoc", []string{"scavenger"}},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			g := NewGenerator(tt.genre)
			for _, class := range tt.expectedClasses {
				if _, ok := g.templates[class]; !ok {
					t.Errorf("Expected template for class %s in genre %s", class, tt.genre)
				}
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	tests := []struct {
		name   string
		genre  string
		class  string
		weapon string
		armor  string
	}{
		{"Fantasy Warrior", "fantasy", "warrior", "sword", "heavy"},
		{"Fantasy Mage", "fantasy", "mage", "staff", "light"},
		{"Fantasy Rogue", "fantasy", "rogue", "dagger", "medium"},
		{"SciFi Soldier", "scifi", "soldier", "gun", "heavy"},
		{"Cyberpunk Hacker", "cyberpunk", "hacker", "pistol", "light"},
		{"Horror Survivor", "horror", "survivor", "blade", "medium"},
		{"Postapoc Scavenger", "postapoc", "scavenger", "gun", "medium"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.genre)
			img := g.Generate(tt.class, 12345, AnimIdle, 0, tt.weapon, tt.armor)

			if img == nil {
				t.Error("Generate returned nil image")
			}

			bounds := img.Bounds()
			if bounds.Dx() != 48 || bounds.Dy() != 48 {
				t.Errorf("Image size = %dx%d, want 48x48", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestGenerator_GenerateAnimationStates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	g := NewGenerator("fantasy")

	states := []AnimationState{
		AnimIdle,
		AnimWalk,
		AnimAttack,
		AnimHurt,
		AnimDeath,
		AnimDodge,
		AnimCast,
	}

	for _, state := range states {
		t.Run(string(rune(state)), func(t *testing.T) {
			img := g.Generate("warrior", 42, state, 0, "sword", "heavy")
			if img == nil {
				t.Errorf("Generate failed for state %v", state)
			}
		})
	}
}

func TestGenerator_GenerateAnimationFrames(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	g := NewGenerator("fantasy")

	for frame := 0; frame < 8; frame++ {
		t.Run(string(rune(frame)), func(t *testing.T) {
			img := g.Generate("mage", 99, AnimWalk, frame, "staff", "light")
			if img == nil {
				t.Errorf("Generate failed for frame %d", frame)
			}
		})
	}
}

func TestGenerator_GenerateWithUnknownClass(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	g := NewGenerator("fantasy")
	img := g.Generate("unknown_class", 42, AnimIdle, 0, "sword", "heavy")

	if img == nil {
		t.Error("Generate should not return nil for unknown class")
	}
}

func TestGenerator_DeterministicGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	g1 := NewGenerator("fantasy")
	g2 := NewGenerator("fantasy")

	img1 := g1.Generate("warrior", 12345, AnimIdle, 0, "sword", "heavy")
	img2 := g2.Generate("warrior", 12345, AnimIdle, 0, "sword", "heavy")

	// Compare image sizes
	if img1.Bounds() != img2.Bounds() {
		t.Error("Same seed should produce same image bounds")
	}
}

func TestDarken(t *testing.T) {
	c := color.RGBA{R: 100, G: 150, B: 200, A: 255}
	result := darken(c, 0.5)

	if result.R != 50 || result.G != 75 || result.B != 100 {
		t.Errorf("darken() = {%d, %d, %d}, want {50, 75, 100}", result.R, result.G, result.B)
	}
	if result.A != 255 {
		t.Errorf("darken() alpha = %d, want 255", result.A)
	}
}

func TestLighten(t *testing.T) {
	c := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	result := lighten(c, 2.0)

	if result.R != 200 || result.G != 200 || result.B != 200 {
		t.Errorf("lighten() = {%d, %d, %d}, want {200, 200, 200}", result.R, result.G, result.B)
	}
	if result.A != 255 {
		t.Errorf("lighten() alpha = %d, want 255", result.A)
	}
}

func TestLighten_Clamp(t *testing.T) {
	c := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	result := lighten(c, 2.0)

	// Should clamp at 255
	if result.R != 255 || result.G != 255 || result.B != 255 {
		t.Errorf("lighten() = {%d, %d, %d}, want {255, 255, 255}", result.R, result.G, result.B)
	}
}

func TestBodyProps(t *testing.T) {
	props := BodyProps{
		HeadSize:   0.22,
		TorsoWidth: 0.35,
		ArmLength:  0.38,
		LegLength:  0.40,
	}

	if props.HeadSize != 0.22 {
		t.Errorf("HeadSize = %v, want 0.22", props.HeadSize)
	}
	if props.TorsoWidth != 0.35 {
		t.Errorf("TorsoWidth = %v, want 0.35", props.TorsoWidth)
	}
	if props.ArmLength != 0.38 {
		t.Errorf("ArmLength = %v, want 0.38", props.ArmLength)
	}
	if props.LegLength != 0.40 {
		t.Errorf("LegLength = %v, want 0.40", props.LegLength)
	}
}

func TestClassTemplate(t *testing.T) {
	template := ClassTemplate{
		BaseColor:   color.RGBA{120, 100, 80, 255},
		AccentColor: color.RGBA{180, 150, 100, 255},
		SkinTone:    color.RGBA{220, 180, 150, 255},
		HairColor:   color.RGBA{80, 60, 40, 255},
		BodyProportions: BodyProps{
			HeadSize:   0.22,
			TorsoWidth: 0.35,
			ArmLength:  0.38,
			LegLength:  0.40,
		},
	}

	if template.BaseColor.R != 120 {
		t.Errorf("BaseColor.R = %d, want 120", template.BaseColor.R)
	}
	if template.BodyProportions.HeadSize != 0.22 {
		t.Errorf("BodyProportions.HeadSize = %v, want 0.22", template.BodyProportions.HeadSize)
	}
}

func TestGenerator_GenerateAllWeaponTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	g := NewGenerator("fantasy")

	weapons := []string{"sword", "blade", "staff", "wand", "gun", "pistol", "dagger", ""}

	for _, weapon := range weapons {
		t.Run(weapon, func(t *testing.T) {
			img := g.Generate("warrior", 42, AnimAttack, 0, weapon, "heavy")
			if img == nil {
				t.Errorf("Generate failed for weapon %s", weapon)
			}
		})
	}
}

func TestGenerator_GenerateAllArmorTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that requires graphics context")
	}

	g := NewGenerator("fantasy")

	armors := []string{"heavy", "medium", "light", ""}

	for _, armor := range armors {
		t.Run(armor, func(t *testing.T) {
			img := g.Generate("warrior", 42, AnimIdle, 0, "sword", armor)
			if img == nil {
				t.Errorf("Generate failed for armor %s", armor)
			}
		})
	}
}

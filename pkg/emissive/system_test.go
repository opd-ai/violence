package emissive

import (
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestComponentType(t *testing.T) {
	comp := NewComponent(TypeFlame, color.RGBA{R: 255, G: 180, B: 80, A: 255})
	if got := comp.Type(); got != "emissive.Component" {
		t.Errorf("Component.Type() = %q, want %q", got, "emissive.Component")
	}
}

func TestNewComponent(t *testing.T) {
	tests := []struct {
		name      string
		glowType  GlowType
		wantColor color.RGBA
	}{
		{"flame", TypeFlame, color.RGBA{R: 255, G: 100, B: 50, A: 255}},
		{"magic", TypeMagic, color.RGBA{R: 100, G: 100, B: 255, A: 255}},
		{"projectile", TypeProjectile, color.RGBA{R: 255, G: 255, B: 0, A: 255}},
		{"eye", TypeEye, color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"neon", TypeNeon, color.RGBA{R: 0, G: 255, B: 255, A: 255}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewComponent(tt.glowType, tt.wantColor)
			if comp.GlowType != tt.glowType {
				t.Errorf("GlowType = %v, want %v", comp.GlowType, tt.glowType)
			}
			if comp.Color != tt.wantColor {
				t.Errorf("Color = %v, want %v", comp.Color, tt.wantColor)
			}
			if comp.Intensity != 1.0 {
				t.Errorf("Intensity = %f, want 1.0", comp.Intensity)
			}
			if !comp.Enabled {
				t.Error("Component should be enabled by default")
			}
		})
	}
}

func TestNewFlameGlow(t *testing.T) {
	comp := NewFlameGlow()
	if comp.GlowType != TypeFlame {
		t.Errorf("Type = %v, want TypeFlame", comp.GlowType)
	}
	if comp.Color.R < 200 {
		t.Errorf("Flame glow should be warm (high red), got R=%d", comp.Color.R)
	}
	if comp.PulseSpeed == 0 {
		t.Error("Flame glow should have pulse speed for flicker")
	}
}

func TestNewMagicGlow(t *testing.T) {
	magicColor := color.RGBA{R: 150, G: 100, B: 255, A: 255}
	comp := NewMagicGlow(magicColor)
	if comp.GlowType != TypeMagic {
		t.Errorf("Type = %v, want TypeMagic", comp.GlowType)
	}
	if comp.Color != magicColor {
		t.Errorf("Color = %v, want %v", comp.Color, magicColor)
	}
	if comp.Intensity < 1.0 {
		t.Errorf("Magic glow should have elevated intensity, got %f", comp.Intensity)
	}
}

func TestNewProjectileGlow(t *testing.T) {
	projColor := color.RGBA{R: 255, G: 200, B: 0, A: 255}
	comp := NewProjectileGlow(projColor)
	if comp.GlowType != TypeProjectile {
		t.Errorf("Type = %v, want TypeProjectile", comp.GlowType)
	}
	if comp.Radius > 16 {
		t.Errorf("Projectile glow should be compact, got radius %f", comp.Radius)
	}
}

func TestNewEyeGlow(t *testing.T) {
	eyeColor := color.RGBA{R: 255, G: 50, B: 50, A: 255}
	comp := NewEyeGlow(eyeColor)
	if comp.GlowType != TypeEye {
		t.Errorf("Type = %v, want TypeEye", comp.GlowType)
	}
	if comp.Radius > 12 {
		t.Errorf("Eye glow should be small, got radius %f", comp.Radius)
	}
}

func TestNewNeonGlow(t *testing.T) {
	neonColor := color.RGBA{R: 0, G: 255, B: 255, A: 255}
	comp := NewNeonGlow(neonColor)
	if comp.GlowType != TypeNeon {
		t.Errorf("Type = %v, want TypeNeon", comp.GlowType)
	}
	if comp.PulseSpeed != 0 {
		t.Errorf("Neon glow should be steady (no pulse), got %f", comp.PulseSpeed)
	}
}

func TestSystemCreation(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			sys := NewSystem(genre, 12345)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.genreID != genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, genre)
			}
			preset := sys.GetPreset()
			if preset.IntensityMult == 0 {
				t.Error("Preset IntensityMult should not be zero")
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	initialPreset := sys.GetPreset()

	sys.SetGenre("cyberpunk")
	newPreset := sys.GetPreset()

	if sys.genreID != "cyberpunk" {
		t.Errorf("genreID = %q, want cyberpunk", sys.genreID)
	}
	if newPreset.FlameColor == initialPreset.FlameColor {
		t.Error("Preset should change when genre changes")
	}
}

func TestSetScreenSize(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	sys.SetScreenSize(640, 480)

	if sys.screenW != 640 || sys.screenH != 480 {
		t.Errorf("Screen size = %dx%d, want 640x480", sys.screenW, sys.screenH)
	}
}

func TestSetCamera(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	sys.SetCamera(100.0, 200.0)

	if sys.cameraX != 100.0 || sys.cameraY != 200.0 {
		t.Errorf("Camera = (%f, %f), want (100, 200)", sys.cameraX, sys.cameraY)
	}
}

func TestUpdate(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()

	entity := world.AddEntity()
	comp := NewFlameGlow()
	world.AddComponent(entity, comp)
	world.AddComponent(entity, &engine.Position{X: 50, Y: 50})

	initialPhase := comp.PulsePhase
	sys.Update(world)

	if comp.PulsePhase == initialPhase {
		t.Error("PulsePhase should change after update for flame glow")
	}
}

func TestUpdateScreenPosition(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	sys.SetCamera(100.0, 100.0)
	world := engine.NewWorld()

	entity := world.AddEntity()
	comp := NewFlameGlow()
	world.AddComponent(entity, comp)
	world.AddComponent(entity, &engine.Position{X: 150, Y: 150})

	sys.Update(world)

	expectedScreenX := 50.0
	expectedScreenY := 50.0
	if comp.ScreenX != expectedScreenX || comp.ScreenY != expectedScreenY {
		t.Errorf("ScreenPos = (%f, %f), want (%f, %f)",
			comp.ScreenX, comp.ScreenY, expectedScreenX, expectedScreenY)
	}
}

func TestUpdateDisabled(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()

	entity := world.AddEntity()
	comp := NewFlameGlow()
	comp.Enabled = false
	world.AddComponent(entity, comp)
	world.AddComponent(entity, &engine.Position{X: 50, Y: 50})

	initialPhase := comp.PulsePhase
	sys.Update(world)

	if comp.PulsePhase != initialPhase {
		t.Error("Disabled component should not update")
	}
}

func TestCreateGlowSprite(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp := NewFlameGlow()

	sprite := sys.createGlowSprite(comp, 16)
	if sprite == nil {
		t.Fatal("createGlowSprite returned nil")
	}

	bounds := sprite.Bounds()
	if bounds.Dx() != 32 || bounds.Dy() != 32 {
		t.Errorf("Sprite size = %dx%d, want 32x32", bounds.Dx(), bounds.Dy())
	}
}

func TestGlowSpriteCache(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp := NewFlameGlow()

	sprite1 := sys.getOrCreateGlowSprite(comp, 16)
	sprite2 := sys.getOrCreateGlowSprite(comp, 16)

	if sprite1 != sprite2 {
		t.Error("Cache should return same sprite for identical parameters")
	}
}

func TestGlowSpriteCacheDifferentParams(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp1 := NewFlameGlow()
	comp2 := NewMagicGlow(color.RGBA{R: 100, G: 100, B: 255, A: 255})

	sprite1 := sys.getOrCreateGlowSprite(comp1, 16)
	sprite2 := sys.getOrCreateGlowSprite(comp2, 16)

	if sprite1 == sprite2 {
		t.Error("Different glow types should produce different sprites")
	}
}

func TestClearCache(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp := NewFlameGlow()

	sys.getOrCreateGlowSprite(comp, 16)
	sys.ClearCache()

	sys.cacheMu.RLock()
	cacheLen := len(sys.cache)
	sys.cacheMu.RUnlock()

	if cacheLen != 0 {
		t.Errorf("Cache length after clear = %d, want 0", cacheLen)
	}
}

func TestModulatedIntensityFlame(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp := NewFlameGlow()
	comp.PulsePhase = 0

	baseIntensity := 1.0
	intensities := make([]float64, 10)

	for i := 0; i < 10; i++ {
		comp.PulsePhase = float64(i) * 0.5
		intensities[i] = sys.getModulatedIntensity(comp, baseIntensity)
	}

	allSame := true
	for i := 1; i < 10; i++ {
		if intensities[i] != intensities[0] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("Flame glow intensity should vary with phase")
	}
}

func TestModulatedIntensityNeon(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	comp := NewNeonGlow(color.RGBA{R: 255, G: 0, B: 255, A: 255})

	baseIntensity := 1.0
	intensities := make([]float64, 10)

	for i := 0; i < 10; i++ {
		comp.PulsePhase = float64(i) * 0.5
		intensities[i] = sys.getModulatedIntensity(comp, baseIntensity)
	}

	allSame := true
	for i := 1; i < 10; i++ {
		if intensities[i] != intensities[0] {
			allSame = false
			break
		}
	}
	if !allSame {
		t.Error("Neon glow should have steady intensity (no pulse)")
	}
}

func TestGenrePresetValues(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset := genrePresets[genre]

			if preset.IntensityMult <= 0 {
				t.Errorf("%s IntensityMult = %f, should be > 0", genre, preset.IntensityMult)
			}
			if preset.RadiusMult <= 0 {
				t.Errorf("%s RadiusMult = %f, should be > 0", genre, preset.RadiusMult)
			}
			if preset.FlameColor.A != 255 {
				t.Errorf("%s FlameColor.A = %d, want 255", genre, preset.FlameColor.A)
			}
		})
	}
}

func TestGlowTypesComplete(t *testing.T) {
	types := []GlowType{
		TypeFlame,
		TypeMagic,
		TypeProjectile,
		TypeEye,
		TypeNeon,
		TypeRadioactive,
		TypeElectric,
	}

	for _, glowType := range types {
		comp := NewComponent(glowType, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		if comp.GlowType != glowType {
			t.Errorf("NewComponent(%d) produced Type = %d", glowType, comp.GlowType)
		}
	}
}

func TestDistanceScaling(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()

	entity := world.AddEntity()
	comp := NewFlameGlow()
	comp.Radius = 20.0
	world.AddComponent(entity, comp)
	world.AddComponent(entity, &engine.Position{X: 500, Y: 500})

	sys.SetCamera(0, 0)
	sys.Update(world)

	if comp.Distance < 100 {
		t.Errorf("Distance = %f, expected > 100", comp.Distance)
	}
}

func BenchmarkCreateGlowSprite(b *testing.B) {
	sys := NewSystem("fantasy", 12345)
	comp := NewFlameGlow()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.createGlowSprite(comp, 16)
	}
}

func BenchmarkGetOrCreateGlowSprite(b *testing.B) {
	sys := NewSystem("fantasy", 12345)
	comp := NewFlameGlow()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.getOrCreateGlowSprite(comp, 16)
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()

	for i := 0; i < 100; i++ {
		entity := world.AddEntity()
		world.AddComponent(entity, NewFlameGlow())
		world.AddComponent(entity, &engine.Position{X: float64(i * 10), Y: float64(i * 10)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(world)
	}
}

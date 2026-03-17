//go:build !headless

package impactburst

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// createTestImage creates a minimal test image for rendering tests.
func createTestImage(w, h int) *ebiten.Image {
	return ebiten.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, w, h)))
}

func TestNewRenderer(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			r := NewRenderer(genre)
			if r == nil {
				t.Fatal("NewRenderer returned nil")
			}
			if r.genreID != genre {
				t.Errorf("expected genre %s, got %s", genre, r.genreID)
			}
		})
	}
}

func TestRendererSetGenre(t *testing.T) {
	r := NewRenderer("fantasy")

	r.SetGenre("cyberpunk")

	if r.genreID != "cyberpunk" {
		t.Errorf("expected genre cyberpunk, got %s", r.genreID)
	}
}

func TestRenderEmptyImpacts(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	// Should not panic with empty slice
	r.Render(screen, []Impact{}, 0, 0, 320, 240)
}

func TestRenderSingleImpact(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(0, 0, 0, ImpactMelee, MaterialFlesh, 1.0)

	impacts := sys.GetGlobalImpacts()

	// Should not panic
	r.Render(screen, impacts, 0, 0, 320, 240)
}

func TestRenderMultipleImpacts(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)

	// Spawn various impact types
	sys.SpawnImpact(0, 0, 0, ImpactMelee, MaterialFlesh, 1.0)
	sys.SpawnImpact(5, 5, 1.0, ImpactProjectile, MaterialMetal, 1.5)
	sys.SpawnImpact(-5, 3, 2.0, ImpactExplosion, MaterialStone, 2.0)
	sys.SpawnImpact(2, -4, 0.5, ImpactMagic, MaterialEnergy, 0.8)
	sys.SpawnImpact(-3, -3, 3.0, ImpactCritical, MaterialFlesh, 1.2)

	impacts := sys.GetGlobalImpacts()

	// Should not panic
	r.Render(screen, impacts, 0, 0, 320, 240)
}

func TestRenderWithCameraOffset(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(50, 50, 0, ImpactMelee, MaterialFlesh, 1.0)

	impacts := sys.GetGlobalImpacts()

	// Render with various camera offsets
	r.Render(screen, impacts, 0, 0, 320, 240)
	r.Render(screen, impacts, 50, 50, 320, 240)
	r.Render(screen, impacts, -100, -100, 320, 240)
	r.Render(screen, impacts, 1000, 1000, 320, 240) // Should be off-screen
}

func TestRenderOffScreenImpact(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	// Create impact far off-screen
	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(10000, 10000, 0, ImpactMelee, MaterialFlesh, 1.0)

	impacts := sys.GetGlobalImpacts()

	// Should not panic, impact should be culled
	r.Render(screen, impacts, 0, 0, 320, 240)
}

func TestRenderAtVariousLifetimes(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(0, 0, 0, ImpactExplosion, MaterialMetal, 1.0)

	// Render at different ages
	for frame := 0; frame < 100; frame++ {
		sys.Update(nil)
		impacts := sys.GetGlobalImpacts()
		if len(impacts) > 0 {
			r.Render(screen, impacts, 0, 0, 320, 240)
		}
	}
}

func TestRenderAllMaterialTypes(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	materials := []MaterialType{
		MaterialFlesh,
		MaterialMetal,
		MaterialStone,
		MaterialWood,
		MaterialEnergy,
		MaterialEthereal,
	}

	for _, mat := range materials {
		t.Run(materialName(mat), func(t *testing.T) {
			sys := NewSystem("fantasy", 12345)
			sys.SpawnImpact(0, 0, 0, ImpactMelee, mat, 1.0)

			impacts := sys.GetGlobalImpacts()
			r.Render(screen, impacts, 0, 0, 320, 240)
		})
	}
}

func TestRenderAllImpactTypes(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	types := []ImpactType{
		ImpactMelee,
		ImpactProjectile,
		ImpactExplosion,
		ImpactMagic,
		ImpactCritical,
		ImpactBlock,
		ImpactDeath,
	}

	for _, typ := range types {
		t.Run(impactTypeName(typ), func(t *testing.T) {
			sys := NewSystem("fantasy", 12345)
			sys.SpawnImpact(0, 0, 0, typ, MaterialFlesh, 1.0)

			impacts := sys.GetGlobalImpacts()
			r.Render(screen, impacts, 0, 0, 320, 240)
		})
	}
}

func TestRenderAllGenres(t *testing.T) {
	screen := createTestImage(320, 240)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			r := NewRenderer(genre)
			sys := NewSystem(genre, 12345)

			sys.SpawnImpact(0, 0, 0, ImpactCritical, MaterialMetal, 1.5)
			sys.SpawnImpact(5, 5, 1, ImpactMagic, MaterialEnergy, 1.0)

			impacts := sys.GetGlobalImpacts()
			r.Render(screen, impacts, 0, 0, 320, 240)
		})
	}
}

func TestRenderToWorld(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(0, 0, 0, ImpactMelee, MaterialFlesh, 1.0)

	impact := &sys.GetGlobalImpacts()[0]

	// Various transform values
	transforms := []struct {
		x, y float64
	}{
		{0, 1.0},
		{0.5, 2.0},
		{-0.5, 1.5},
		{0, 0.5},
	}

	for _, tr := range transforms {
		r.RenderToWorld(screen, impact, tr.x, tr.y, 320, 240)
	}
}

func TestRenderToWorldBehindCamera(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(0, 0, 0, ImpactMelee, MaterialFlesh, 1.0)

	impact := &sys.GetGlobalImpacts()[0]

	// Negative Y (behind camera) should not render
	r.RenderToWorld(screen, impact, 0, -1.0, 320, 240)
	r.RenderToWorld(screen, impact, 0, 0.05, 320, 240)
}

func TestRenderGlowEffect(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	// Critical and magic have glow
	sys.SpawnImpact(0, 0, 0, ImpactCritical, MaterialEnergy, 1.0)

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 {
		t.Fatal("no impacts")
	}

	// Set glow intensity manually
	impacts[0].GlowIntensity = 1.0

	r.Render(screen, impacts, 0, 0, 320, 240)
}

func TestRenderShockwaveEffect(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(0, 0, 0, ImpactExplosion, MaterialMetal, 1.0)

	// Update to expand shockwave
	for i := 0; i < 10; i++ {
		sys.Update(nil)
	}

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 {
		t.Skip("impact expired")
	}

	if impacts[0].ShockwaveRadius <= 0 {
		t.Error("shockwave should have expanded")
	}

	r.Render(screen, impacts, 0, 0, 320, 240)
}

func TestRenderFlashEffect(t *testing.T) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(0, 0, 0, ImpactMelee, MaterialFlesh, 1.0)

	impacts := sys.GetGlobalImpacts()
	if len(impacts) == 0 {
		t.Fatal("no impacts")
	}

	// New impact should have flash
	impacts[0].FlashAlpha = 1.0

	r.Render(screen, impacts, 0, 0, 320, 240)
}

func TestDarkenColor(t *testing.T) {
	c := color.RGBA{R: 200, G: 200, B: 200, A: 255}

	darkened := darkenColor(c, 0.5)

	if darkened.R >= c.R || darkened.G >= c.G || darkened.B >= c.B {
		t.Error("darkened color should have lower RGB values")
	}

	if darkened.A != c.A {
		t.Error("alpha should be unchanged")
	}

	// Exact value check
	expected := color.RGBA{R: 100, G: 100, B: 100, A: 255}
	if darkened.R != expected.R || darkened.G != expected.G || darkened.B != expected.B {
		t.Errorf("expected %v, got %v", expected, darkened)
	}
}

func TestGetProfileForImpact(t *testing.T) {
	impact := &Impact{
		Type:     ImpactMelee,
		Material: MaterialFlesh,
	}

	profile := getProfileForImpact(impact, "fantasy")

	if profile.PrimaryColor.A == 0 {
		t.Error("profile primary color should have alpha")
	}
	if profile.Duration <= 0 {
		t.Error("profile should have positive duration")
	}
}

func BenchmarkRender(b *testing.B) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	for i := 0; i < 20; i++ {
		sys.SpawnImpact(float64(i-10), float64(i-10), float64(i)*0.3, ImpactMelee, MaterialFlesh, 1.0)
	}

	impacts := sys.GetGlobalImpacts()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Render(screen, impacts, 0, 0, 320, 240)
	}
}

func BenchmarkRenderToWorld(b *testing.B) {
	r := NewRenderer("fantasy")
	screen := createTestImage(320, 240)

	sys := NewSystem("fantasy", 12345)
	sys.SpawnImpact(0, 0, 0, ImpactCritical, MaterialMetal, 1.5)

	impact := &sys.GetGlobalImpacts()[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RenderToWorld(screen, impact, 0.5, 2.0, 320, 240)
	}
}

// Helper functions for test names
func materialName(m MaterialType) string {
	names := []string{"flesh", "metal", "stone", "wood", "energy", "ethereal"}
	if int(m) < len(names) {
		return names[m]
	}
	return "unknown"
}

func impactTypeName(t ImpactType) string {
	names := []string{"melee", "projectile", "explosion", "magic", "critical", "block", "death"}
	if int(t) < len(names) {
		return names[t]
	}
	return "unknown"
}

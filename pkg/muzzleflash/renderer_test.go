package muzzleflash

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewRenderer(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	renderer := NewRenderer(sys, "fantasy")

	if renderer == nil {
		t.Fatal("NewRenderer returned nil")
	}
	if renderer.system != sys {
		t.Error("renderer.system not set correctly")
	}
	if renderer.genreID != "fantasy" {
		t.Errorf("genreID = %q, want %q", renderer.genreID, "fantasy")
	}
	if renderer.logger == nil {
		t.Error("logger should be initialized")
	}
}

func TestRendererSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	renderer := NewRenderer(sys, "fantasy")

	renderer.SetGenre("scifi")
	if renderer.genreID != "scifi" {
		t.Errorf("genreID = %q, want %q", renderer.genreID, "scifi")
	}
}

func TestRenderNoFlashes(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	renderer := NewRenderer(sys, "fantasy")
	world := engine.NewWorld()

	// Should not panic with no flashes
	img := ebiten.NewImage(320, 200)
	defer img.Dispose()

	renderer.Render(img, world, 0, 0, 320, 200)
}

func TestRenderWithFlashes(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	renderer := NewRenderer(sys, "fantasy")
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 5.0, 5.0, 0, "bullet", 1.0)

	img := ebiten.NewImage(320, 200)
	defer img.Dispose()

	// Should not panic
	renderer.Render(img, world, 0, 0, 320, 200)
}

func TestRenderSingle(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	renderer := NewRenderer(sys, "fantasy")

	img := ebiten.NewImage(320, 200)
	defer img.Dispose()

	// Test each flash type
	flashTypes := []string{"bullet", "plasma", "energy", "fire", "magic", "shotgun", "laser"}

	for _, flashType := range flashTypes {
		t.Run(flashType, func(t *testing.T) {
			// Should not panic for any progress value
			renderer.RenderSingle(img, 160, 100, flashType, 0.0, 1.0)
			renderer.RenderSingle(img, 160, 100, flashType, 0.5, 1.0)
			renderer.RenderSingle(img, 160, 100, flashType, 1.0, 1.0)
		})
	}
}

func TestRenderDataGeneration(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 0, 0, 0, "bullet", 1.0)

	flashes := sys.GetActiveFlashes(world, entity)
	if len(flashes) == 0 {
		t.Fatal("No flash created")
	}

	data := sys.PrepareRenderData(flashes[0], 0, 0, 320, 200)

	// Verify render data structure
	if data.ScreenX != 160 { // Center of 320
		t.Errorf("ScreenX = %f, want 160", data.ScreenX)
	}
	if data.ScreenY != 100 { // Center of 200
		t.Errorf("ScreenY = %f, want 100", data.ScreenY)
	}

	if len(data.RayAngles) == 0 {
		// Bullet should have rays
		profile := GetProfile("bullet")
		if profile.RayCount > 0 {
			t.Error("RayAngles should be populated for bullet type")
		}
	}
}

func TestRenderOffScreen(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	renderer := NewRenderer(sys, "fantasy")
	world := engine.NewWorld()
	entity := world.AddEntity()

	// Spawn flash far off-screen
	sys.SpawnFlash(world, entity, 1000.0, 1000.0, 0, "bullet", 1.0)

	img := ebiten.NewImage(320, 200)
	defer img.Dispose()

	// Should skip rendering for off-screen flash without panic
	renderer.Render(img, world, 0, 0, 320, 200)
}

func TestRenderAlphaProgression(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 0, 0, 0, "bullet", 1.0)

	flashes := sys.GetActiveFlashes(world, entity)
	if len(flashes) == 0 {
		t.Fatal("No flash created")
	}

	// Test alpha at different ages
	testAges := []float64{0.0, 0.01, 0.03, 0.05}
	lastAlpha := 2.0 // Start high

	for _, age := range testAges {
		flashes[0].Age = age
		data := sys.PrepareRenderData(flashes[0], 0, 0, 320, 200)

		if data.Alpha > lastAlpha && age > 0 {
			t.Errorf("Alpha should decrease over time: at age %f got alpha %f > previous %f",
				age, data.Alpha, lastAlpha)
		}
		lastAlpha = data.Alpha
	}
}

func TestFlashRenderDataColors(t *testing.T) {
	data := &FlashRenderData{
		ScreenX:        100,
		ScreenY:        100,
		Scale:          1.0,
		Progress:       0.0,
		Alpha:          1.0,
		CoreRadius:     5.0,
		OuterRadius:    10.0,
		RayLength:      15.0,
		RayAngles:      []float64{0, 1.57, 3.14, 4.71},
		PrimaryColor:   color.RGBA{R: 255, G: 200, B: 150, A: 255},
		SecondaryColor: color.RGBA{R: 255, G: 180, B: 80, A: 200},
		Profile:        GetProfile("bullet"),
	}

	if data.PrimaryColor.A == 0 {
		t.Error("Primary color should have alpha")
	}
	if data.SecondaryColor.A == 0 {
		t.Error("Secondary color should have alpha")
	}
}

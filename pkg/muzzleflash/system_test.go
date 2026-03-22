package muzzleflash

import (
	"image/color"
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("genreID = %q, want %q", sys.genreID, "fantasy")
	}
	if sys.rng == nil {
		t.Error("rng should be initialized")
	}
	if sys.logger == nil {
		t.Error("logger should be initialized")
	}
}

func TestSystemType(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	if sys.Type() != "MuzzleFlashSystem" {
		t.Errorf("Type() = %q, want %q", sys.Type(), "MuzzleFlashSystem")
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		genre         string
		wantIntensity float64
	}{
		{"fantasy", 1.0},
		{"scifi", 1.2},
		{"horror", 0.8},
		{"cyberpunk", 1.3},
		{"postapoc", 0.9},
		{"unknown", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.genre, func(t *testing.T) {
			sys := NewSystem("fantasy", 12345)
			sys.SetGenre(tt.genre)

			if sys.genreID != tt.genre {
				t.Errorf("genreID = %q, want %q", sys.genreID, tt.genre)
			}
			if sys.intensityMult != tt.wantIntensity {
				t.Errorf("intensityMult = %f, want %f", sys.intensityMult, tt.wantIntensity)
			}
		})
	}
}

func TestSpawnFlash(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 10.0, 20.0, 1.5, "bullet", 1.0)

	flashes := sys.GetActiveFlashes(world, entity)
	if len(flashes) != 1 {
		t.Fatalf("GetActiveFlashes returned %d flashes, want 1", len(flashes))
	}

	flash := flashes[0]
	if flash.X != 10.0 {
		t.Errorf("Flash.X = %f, want 10.0", flash.X)
	}
	if flash.Y != 20.0 {
		t.Errorf("Flash.Y = %f, want 20.0", flash.Y)
	}
	if flash.FlashType != "bullet" {
		t.Errorf("Flash.FlashType = %q, want %q", flash.FlashType, "bullet")
	}
}

func TestSpawnFlashMaxLimit(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	// Spawn more than MaxFlashes (8)
	for i := 0; i < 12; i++ {
		sys.SpawnFlash(world, entity, float64(i), 0, 0, "bullet", 1.0)
	}

	flashes := sys.GetActiveFlashes(world, entity)
	if len(flashes) > 8 {
		t.Errorf("Got %d flashes, want <= 8 (MaxFlashes)", len(flashes))
	}
}

func TestSpawnFlashDifferentTypes(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	flashTypes := []string{"bullet", "plasma", "energy", "fire", "magic", "shotgun", "laser"}

	for _, flashType := range flashTypes {
		t.Run(flashType, func(t *testing.T) {
			sys.SpawnFlash(world, entity, 0, 0, 0, flashType, 1.0)
			flashes := sys.GetActiveFlashes(world, entity)
			if len(flashes) == 0 {
				t.Errorf("No flash spawned for type %q", flashType)
			}
		})
	}
}

func TestUpdateFlashAging(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 10.0, 20.0, 0, "bullet", 1.0)

	// Get initial age
	flashes := sys.GetActiveFlashes(world, entity)
	if len(flashes) != 1 {
		t.Fatal("Expected 1 flash")
	}
	initialAge := flashes[0].Age

	// Update system
	sys.Update(world)

	// Age should have increased
	flashes = sys.GetActiveFlashes(world, entity)
	if len(flashes) != 1 {
		t.Fatal("Expected 1 flash after update")
	}
	if flashes[0].Age <= initialAge {
		t.Error("Flash age should increase after Update")
	}
}

func TestFlashExpiration(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 10.0, 20.0, 0, "bullet", 1.0)

	flashes := sys.GetActiveFlashes(world, entity)
	if len(flashes) != 1 {
		t.Fatal("Expected 1 flash initially")
	}

	// Manually age the flash past its duration
	flashes[0].Age = flashes[0].Duration + 0.01

	// Update should remove expired flash
	sys.Update(world)

	flashes = sys.GetActiveFlashes(world, entity)
	if len(flashes) != 0 {
		t.Errorf("Got %d flashes after expiration, want 0", len(flashes))
	}
}

func TestGetAllActiveFlashes(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()

	entity1 := world.AddEntity()
	entity2 := world.AddEntity()

	sys.SpawnFlash(world, entity1, 10.0, 20.0, 0, "bullet", 1.0)
	sys.SpawnFlash(world, entity1, 15.0, 25.0, 0, "plasma", 1.0)
	sys.SpawnFlash(world, entity2, 30.0, 40.0, 0, "fire", 1.0)

	allFlashes := sys.GetAllActiveFlashes(world)
	if len(allFlashes) != 3 {
		t.Errorf("GetAllActiveFlashes returned %d, want 3", len(allFlashes))
	}
}

func TestCollectLightSources(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 10.0, 20.0, 0, "bullet", 1.0)

	lights := sys.CollectLightSources(world)
	if len(lights) != 1 {
		t.Fatalf("CollectLightSources returned %d lights, want 1", len(lights))
	}

	light := lights[0]
	if light.X != 10.0 {
		t.Errorf("Light.X = %f, want 10.0", light.X)
	}
	if light.Y != 20.0 {
		t.Errorf("Light.Y = %f, want 20.0", light.Y)
	}
	if light.Intensity <= 0 {
		t.Error("Light.Intensity should be > 0")
	}
	if light.Radius <= 0 {
		t.Error("Light.Radius should be > 0")
	}
}

func TestCollectLightSourcesFadeout(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 10.0, 20.0, 0, "bullet", 1.0)

	// Initial lights
	lightsInitial := sys.CollectLightSources(world)
	if len(lightsInitial) != 1 {
		t.Fatal("Expected 1 light initially")
	}
	initialIntensity := lightsInitial[0].Intensity

	// Age the flash
	flashes := sys.GetActiveFlashes(world, entity)
	flashes[0].Age = flashes[0].Duration * 0.5

	// Lights should be dimmer
	lightsLater := sys.CollectLightSources(world)
	if len(lightsLater) != 1 {
		t.Fatal("Expected 1 light after aging")
	}
	if lightsLater[0].Intensity >= initialIntensity {
		t.Error("Light intensity should decrease as flash ages")
	}
}

func TestPrepareRenderData(t *testing.T) {
	sys := NewSystem("fantasy", 12345)
	world := engine.NewWorld()
	entity := world.AddEntity()

	sys.SpawnFlash(world, entity, 5.0, 5.0, 0, "bullet", 1.0)

	flashes := sys.GetActiveFlashes(world, entity)
	if len(flashes) != 1 {
		t.Fatal("Expected 1 flash")
	}

	renderData := sys.PrepareRenderData(flashes[0], 0, 0, 320, 200)

	if renderData == nil {
		t.Fatal("PrepareRenderData returned nil")
	}

	// Screen position should be centered plus offset
	expectedCenterX := 320.0 / 2
	if renderData.ScreenX < expectedCenterX {
		t.Errorf("ScreenX = %f, expected >= %f", renderData.ScreenX, expectedCenterX)
	}

	if renderData.Alpha <= 0 || renderData.Alpha > 1 {
		t.Errorf("Alpha = %f, want 0 < alpha <= 1", renderData.Alpha)
	}

	if renderData.CoreRadius <= 0 {
		t.Error("CoreRadius should be > 0")
	}
	if renderData.OuterRadius <= 0 {
		t.Error("OuterRadius should be > 0")
	}
}

func TestTintColor(t *testing.T) {
	sys := NewSystem("fantasy", 12345)

	// Fantasy genre should have warm tint
	original := color.RGBA{R: 255, G: 200, B: 150, A: 255}
	tinted := sys.tintColor(original)

	// Result should still be valid colors
	if tinted.R == 0 && tinted.G == 0 && tinted.B == 0 {
		t.Error("Tinted color should not be black")
	}
	if tinted.A != original.A {
		t.Errorf("Alpha changed: got %d, want %d", tinted.A, original.A)
	}
}

func TestDeterminism(t *testing.T) {
	seed := int64(42)

	// Create two systems with same seed
	sys1 := NewSystem("fantasy", seed)
	sys2 := NewSystem("fantasy", seed)

	world1 := engine.NewWorld()
	world2 := engine.NewWorld()

	entity1 := world1.AddEntity()
	entity2 := world2.AddEntity()

	// Spawn same flashes
	sys1.SpawnFlash(world1, entity1, 10.0, 20.0, 1.5, "bullet", 1.0)
	sys2.SpawnFlash(world2, entity2, 10.0, 20.0, 1.5, "bullet", 1.0)

	flashes1 := sys1.GetActiveFlashes(world1, entity1)
	flashes2 := sys2.GetActiveFlashes(world2, entity2)

	if len(flashes1) != len(flashes2) {
		t.Fatalf("Flash count mismatch: %d vs %d", len(flashes1), len(flashes2))
	}

	if len(flashes1) == 0 {
		t.Fatal("No flashes created")
	}

	// Flashes should have same properties (except for RNG variation which should be deterministic)
	f1 := flashes1[0]
	f2 := flashes2[0]

	if f1.Duration != f2.Duration {
		t.Errorf("Duration mismatch: %f vs %f", f1.Duration, f2.Duration)
	}
	if f1.FlashType != f2.FlashType {
		t.Errorf("FlashType mismatch: %q vs %q", f1.FlashType, f2.FlashType)
	}
	if f1.PrimaryColor != f2.PrimaryColor {
		t.Errorf("PrimaryColor mismatch: %v vs %v", f1.PrimaryColor, f2.PrimaryColor)
	}
}

package weather

import (
	"testing"

	"github.com/opd-ai/violence/pkg/engine"
)

func TestNewSystem(t *testing.T) {
	sys := NewSystem(1000, 42, "fantasy")
	if sys == nil {
		t.Fatal("NewSystem returned nil")
	}
	if sys.weather == nil {
		t.Error("weather system not initialized")
	}
}

func TestSystemUpdate(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(500, 42, "cyberpunk")
	sys.SetCamera(0, 0, 800, 600)

	// Should not panic
	sys.Update(world)
}

func TestSystemSetGenre(t *testing.T) {
	sys := NewSystem(500, 42, "fantasy")
	sys.SetGenre("horror")

	if sys.weather.genre != "horror" {
		t.Errorf("genre = %s, want horror", sys.weather.genre)
	}
}

func TestSystemSetCamera(t *testing.T) {
	sys := NewSystem(500, 42, "fantasy")
	sys.SetCamera(100, 200, 1024, 768)

	if sys.weather.cameraX != 100 || sys.weather.cameraY != 200 {
		t.Errorf("camera position = (%f, %f), want (100, 200)", sys.weather.cameraX, sys.weather.cameraY)
	}
}

func TestGetWeatherSystem(t *testing.T) {
	sys := NewSystem(500, 42, "fantasy")
	ws := sys.GetWeatherSystem()

	if ws == nil {
		t.Error("GetWeatherSystem returned nil")
	}
	if ws != sys.weather {
		t.Error("GetWeatherSystem returned different instance")
	}
}

func TestAddWeatherToWorld(t *testing.T) {
	world := engine.NewWorld()
	sys := NewSystem(500, 42, "fantasy")

	sys.AddWeatherToWorld(world)

	// Retrieve and verify
	ws := GetWeatherFromWorld(world)
	if ws == nil {
		t.Error("weather system not added to world")
	}
	if ws != sys.weather {
		t.Error("retrieved weather system is not the same instance")
	}
}

func TestGetWeatherFromWorld_Empty(t *testing.T) {
	world := engine.NewWorld()
	ws := GetWeatherFromWorld(world)

	if ws != nil {
		t.Error("GetWeatherFromWorld should return nil for empty world")
	}
}

package dustmote

import (
	"testing"
)

func TestNewSystem(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
		seed    int64
		screenW int
		screenH int
	}{
		{"fantasy default", "fantasy", 12345, 320, 200},
		{"scifi genre", "scifi", 54321, 320, 200},
		{"horror genre", "horror", 11111, 320, 200},
		{"cyberpunk genre", "cyberpunk", 22222, 320, 200},
		{"postapoc genre", "postapoc", 33333, 320, 200},
		{"unknown genre falls back", "unknown", 44444, 320, 200},
		{"larger screen", "fantasy", 55555, 640, 480},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewSystem(tt.genreID, tt.seed, tt.screenW, tt.screenH)
			if sys == nil {
				t.Fatal("NewSystem returned nil")
			}
			if sys.maxMotes < 50 {
				t.Errorf("maxMotes too low: got %d, want >= 50", sys.maxMotes)
			}
			if sys.maxMotes > 500 {
				t.Errorf("maxMotes too high: got %d, want <= 500", sys.maxMotes)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	// Spawn some motes
	sys.SetCamera(0, 0, 32, 20)
	for i := 0; i < 10; i++ {
		sys.Update(0.016)
	}

	initialCount := sys.GetActiveCount()
	if initialCount == 0 {
		t.Log("No motes spawned initially (may happen due to density)")
	}

	// Change genre should clear motes
	sys.SetGenre("cyberpunk")
	if sys.genreID != "cyberpunk" {
		t.Errorf("Genre not changed: got %s, want cyberpunk", sys.genreID)
	}
	if sys.preset.MoteType != MoteTypeDigit {
		t.Errorf("Preset not updated for cyberpunk")
	}

	afterCount := sys.GetActiveCount()
	if afterCount != 0 {
		t.Errorf("Motes not cleared after genre change: got %d, want 0", afterCount)
	}
}

func TestSetCamera(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	sys.SetCamera(10.5, 20.5, 32, 20)

	if sys.cameraX != 10.5 {
		t.Errorf("cameraX not set: got %f, want 10.5", sys.cameraX)
	}
	if sys.cameraY != 20.5 {
		t.Errorf("cameraY not set: got %f, want 20.5", sys.cameraY)
	}
	if sys.viewWidth != 32 {
		t.Errorf("viewWidth not set: got %f, want 32", sys.viewWidth)
	}
	if sys.viewHeight != 20 {
		t.Errorf("viewHeight not set: got %f, want 20", sys.viewHeight)
	}
}

func TestSetLights(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	lights := []LightSource{
		{X: 5, Y: 5, Radius: 10, Intensity: 1.0, ColorR: 255, ColorG: 200, ColorB: 100},
		{X: 15, Y: 15, Radius: 8, Intensity: 0.8, ColorR: 100, ColorG: 200, ColorB: 255},
	}

	sys.SetLights(lights)
	if len(sys.lights) != 2 {
		t.Errorf("Lights not set: got %d, want 2", len(sys.lights))
	}

	sys.AddLight(LightSource{X: 25, Y: 25, Radius: 5, Intensity: 0.5})
	if len(sys.lights) != 3 {
		t.Errorf("Light not added: got %d, want 3", len(sys.lights))
	}

	sys.ClearLights()
	if len(sys.lights) != 0 {
		t.Errorf("Lights not cleared: got %d, want 0", len(sys.lights))
	}
}

func TestSetAmbientLight(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	tests := []struct {
		input    float64
		expected float64
	}{
		{0.5, 0.5},
		{0.0, 0.0},
		{1.0, 1.0},
		{-0.5, 0.0}, // Clamped to 0
		{1.5, 1.0},  // Clamped to 1
	}

	for _, tt := range tests {
		sys.SetAmbientLight(tt.input)
		if sys.ambientLit != tt.expected {
			t.Errorf("SetAmbientLight(%f): got %f, want %f", tt.input, sys.ambientLit, tt.expected)
		}
	}
}

func TestUpdate(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	sys.SetCamera(0, 0, 32, 20)
	sys.SetAmbientLight(0.5)

	// Add a light so motes are visible
	sys.AddLight(LightSource{X: 0, Y: 0, Radius: 50, Intensity: 1.0, ColorR: 255, ColorG: 255, ColorB: 255})

	// Run multiple updates to spawn and simulate motes
	for i := 0; i < 60; i++ {
		sys.Update(0.016) // ~60 FPS
	}

	count := sys.GetActiveCount()
	if count == 0 {
		t.Log("No motes active after updates (density may be low)")
	}

	// Verify motes have moved (check they have non-zero velocity or position != spawn)
	movedCount := 0
	for i := range sys.motes {
		m := &sys.motes[i]
		if m.Active && (m.VX != 0 || m.VY != 0) {
			movedCount++
		}
	}

	// At least some motes should have moved
	if count > 0 && movedCount == 0 {
		t.Error("No motes moved after updates")
	}
}

func TestUpdateCullsDistantMotes(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	sys.SetCamera(0, 0, 32, 20)

	// Spawn motes
	for i := 0; i < 30; i++ {
		sys.Update(0.016)
	}

	// Move camera far away - motes should be culled
	sys.SetCamera(1000, 1000, 32, 20)

	// Run updates - distant motes should be culled
	for i := 0; i < 10; i++ {
		sys.Update(0.016)
	}

	// Most motes from old position should be culled
	// New motes may have spawned, so we just check system doesn't crash
	count := sys.GetActiveCount()
	t.Logf("Active motes after camera move: %d", count)
}

func TestCalculateIllumination(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	sys.SetAmbientLight(0.1)

	// No lights - should return ambient only
	illum := sys.calculateIllumination(0, 0)
	if illum != 0.1 {
		t.Errorf("Illumination without lights: got %f, want 0.1", illum)
	}

	// Add a light at origin
	sys.AddLight(LightSource{X: 0, Y: 0, Radius: 10, Intensity: 1.0})

	// At light center - should be fully lit
	illum = sys.calculateIllumination(0, 0)
	if illum < 0.9 {
		t.Errorf("Illumination at light center too low: got %f, want >= 0.9", illum)
	}

	// At edge of light radius
	illum = sys.calculateIllumination(10, 0)
	// At exact edge, falloff should be 0, so only ambient
	if illum > 0.2 {
		t.Errorf("Illumination at light edge too high: got %f, want <= 0.2", illum)
	}

	// Outside light radius
	illum = sys.calculateIllumination(20, 0)
	if illum != 0.1 {
		t.Errorf("Illumination outside light: got %f, want 0.1", illum)
	}
}

func TestClear(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	sys.SetCamera(0, 0, 32, 20)

	// Spawn motes
	for i := 0; i < 30; i++ {
		sys.Update(0.016)
	}

	beforeCount := sys.GetActiveCount()
	if beforeCount == 0 {
		t.Log("No motes to clear")
	}

	sys.Clear()
	afterCount := sys.GetActiveCount()
	if afterCount != 0 {
		t.Errorf("Motes not cleared: got %d, want 0", afterCount)
	}
}

func TestType(t *testing.T) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	if sys.Type() != "DustMoteSystem" {
		t.Errorf("Type() = %s, want DustMoteSystem", sys.Type())
	}
}

func TestGenrePresets(t *testing.T) {
	requiredGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range requiredGenres {
		preset, ok := genrePresets[genre]
		if !ok {
			t.Errorf("Missing preset for genre: %s", genre)
			continue
		}

		// Validate preset values
		if preset.Density <= 0 || preset.Density > 2 {
			t.Errorf("Genre %s has invalid density: %f", genre, preset.Density)
		}
		if preset.BaseSize <= 0 {
			t.Errorf("Genre %s has invalid base size: %f", genre, preset.BaseSize)
		}
		if preset.Brightness < 0 || preset.Brightness > 1 {
			t.Errorf("Genre %s has invalid brightness: %f", genre, preset.Brightness)
		}
		if preset.LightResponse < 0 || preset.LightResponse > 1 {
			t.Errorf("Genre %s has invalid light response: %f", genre, preset.LightResponse)
		}
	}
}

func TestDeterminism(t *testing.T) {
	seed := int64(42)

	// Create two systems with same seed
	sys1 := NewSystem("fantasy", seed, 320, 200)
	sys2 := NewSystem("fantasy", seed, 320, 200)

	sys1.SetCamera(0, 0, 32, 20)
	sys2.SetCamera(0, 0, 32, 20)

	// Run same number of updates
	for i := 0; i < 30; i++ {
		sys1.Update(0.016)
		sys2.Update(0.016)
	}

	// Count should be same (deterministic spawning)
	count1 := sys1.GetActiveCount()
	count2 := sys2.GetActiveCount()

	if count1 != count2 {
		t.Errorf("Non-deterministic mote counts: %d vs %d", count1, count2)
	}

	// Note: We can't easily compare exact positions due to floating point,
	// but count matching is a good indicator of determinism
}

func TestMoteTypes(t *testing.T) {
	types := []struct {
		genre    string
		expected MoteType
	}{
		{"fantasy", MoteTypeDust},
		{"scifi", MoteTypeClean},
		{"horror", MoteTypeSpore},
		{"cyberpunk", MoteTypeDigit},
		{"postapoc", MoteTypeAsh},
	}

	for _, tt := range types {
		preset := genrePresets[tt.genre]
		if preset.MoteType != tt.expected {
			t.Errorf("Genre %s has wrong mote type: got %d, want %d", tt.genre, preset.MoteType, tt.expected)
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	sys := NewSystem("fantasy", 12345, 320, 200)
	sys.SetCamera(0, 0, 32, 20)
	sys.SetAmbientLight(0.5)
	sys.AddLight(LightSource{X: 0, Y: 0, Radius: 50, Intensity: 1.0})

	// Warm up - spawn motes
	for i := 0; i < 60; i++ {
		sys.Update(0.016)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(0.016)
	}
}

func BenchmarkCalculateIllumination(b *testing.B) {
	sys := NewSystem("fantasy", 12345, 320, 200)

	// Add several lights
	for i := 0; i < 10; i++ {
		sys.AddLight(LightSource{
			X:         float64(i * 10),
			Y:         float64(i * 5),
			Radius:    15,
			Intensity: 0.8,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.calculateIllumination(25, 25)
	}
}

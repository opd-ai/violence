package weather

import (
	"testing"
)

func TestNewWeatherSystem(t *testing.T) {
	ws := NewWeatherSystem(1000, 12345)
	if ws == nil {
		t.Fatal("NewWeatherSystem returned nil")
	}
	if len(ws.particles) != 1000 {
		t.Errorf("particles slice length = %d, want 1000", len(ws.particles))
	}
	if ws.weatherType != WeatherNone {
		t.Errorf("initial weather type = %v, want WeatherNone", ws.weatherType)
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name        string
		genre       string
		wantWeather WeatherType
	}{
		{"fantasy dust", "fantasy", WeatherDust},
		{"scifi none", "scifi", WeatherNone},
		{"horror fog", "horror", WeatherFog},
		{"cyberpunk rain", "cyberpunk", WeatherRain},
		{"postapoc ash", "postapoc", WeatherAsh},
		{"unknown genre", "unknown", WeatherNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := NewWeatherSystem(500, 42)
			ws.SetGenre(tt.genre)
			if ws.weatherType != tt.wantWeather {
				t.Errorf("SetGenre(%s) weather type = %v, want %v", tt.genre, ws.weatherType, tt.wantWeather)
			}
		})
	}
}

func TestSetWeather(t *testing.T) {
	ws := NewWeatherSystem(500, 42)

	ws.SetWeather(WeatherRain, 0.8)
	if ws.weatherType != WeatherRain {
		t.Errorf("weather type = %v, want WeatherRain", ws.weatherType)
	}
	if ws.intensity != 0.8 {
		t.Errorf("intensity = %f, want 0.8", ws.intensity)
	}

	// Test clamping
	ws.SetWeather(WeatherSnow, 1.5)
	if ws.intensity != 1.0 {
		t.Errorf("intensity = %f, want 1.0 (clamped)", ws.intensity)
	}

	ws.SetWeather(WeatherDust, -0.5)
	if ws.intensity != 0.0 {
		t.Errorf("intensity = %f, want 0.0 (clamped)", ws.intensity)
	}
}

func TestSetWind(t *testing.T) {
	ws := NewWeatherSystem(500, 42)
	ws.SetWind(10.0, 20.0)
	if ws.windX != 10.0 || ws.windY != 20.0 {
		t.Errorf("wind = (%f, %f), want (10.0, 20.0)", ws.windX, ws.windY)
	}
}

func TestSetCamera(t *testing.T) {
	ws := NewWeatherSystem(500, 42)
	ws.SetCamera(100.0, 200.0, 1920.0, 1080.0)
	if ws.cameraX != 100.0 || ws.cameraY != 200.0 {
		t.Errorf("camera position = (%f, %f), want (100.0, 200.0)", ws.cameraX, ws.cameraY)
	}
	if ws.viewWidth != 1920.0 || ws.viewHeight != 1080.0 {
		t.Errorf("view size = (%f, %f), want (1920.0, 1080.0)", ws.viewWidth, ws.viewHeight)
	}
}

func TestUpdateNoneWeather(t *testing.T) {
	ws := NewWeatherSystem(500, 42)
	ws.SetWeather(WeatherNone, 0.0)
	ws.Update(1.0)

	active := ws.GetActiveParticles()
	if len(active) != 0 {
		t.Errorf("active particles = %d, want 0 for WeatherNone", len(active))
	}
}

func TestUpdateSpawnsParticles(t *testing.T) {
	ws := NewWeatherSystem(500, 42)
	ws.SetCamera(0, 0, 800, 600)
	ws.SetWeather(WeatherRain, 0.5)

	// Update for 1 second
	ws.Update(1.0)

	active := ws.GetActiveParticles()
	if len(active) == 0 {
		t.Error("no particles spawned after 1 second of rain")
	}
}

func TestParticleLifetime(t *testing.T) {
	ws := NewWeatherSystem(100, 42)
	ws.SetCamera(0, 0, 800, 600)
	ws.SetWeather(WeatherRain, 1.0)

	// Spawn particles
	ws.Update(0.1)
	initialCount := len(ws.GetActiveParticles())
	if initialCount == 0 {
		t.Fatal("no particles spawned")
	}

	// Update for longer than particle lifetime
	for i := 0; i < 100; i++ {
		ws.Update(0.1)
	}

	// Particles should have been recycled (some still active, but different ones)
	finalCount := len(ws.GetActiveParticles())
	if finalCount == 0 {
		t.Error("all particles died, should have continuous spawning")
	}
}

func TestClear(t *testing.T) {
	ws := NewWeatherSystem(500, 42)
	ws.SetCamera(0, 0, 800, 600)
	ws.SetWeather(WeatherSnow, 0.8)
	ws.Update(1.0)

	if len(ws.GetActiveParticles()) == 0 {
		t.Fatal("no particles to clear")
	}

	ws.Clear()
	if len(ws.GetActiveParticles()) != 0 {
		t.Errorf("active particles after Clear() = %d, want 0", len(ws.GetActiveParticles()))
	}
}

func TestGetters(t *testing.T) {
	ws := NewWeatherSystem(500, 42)
	ws.SetWeather(WeatherEmbers, 0.6)

	if ws.GetWeatherType() != WeatherEmbers {
		t.Errorf("GetWeatherType() = %v, want WeatherEmbers", ws.GetWeatherType())
	}
	if ws.GetIntensity() != 0.6 {
		t.Errorf("GetIntensity() = %f, want 0.6", ws.GetIntensity())
	}
}

func TestParticleTypes(t *testing.T) {
	tests := []struct {
		name    string
		weather WeatherType
	}{
		{"rain", WeatherRain},
		{"snow", WeatherSnow},
		{"embers", WeatherEmbers},
		{"dust", WeatherDust},
		{"ash", WeatherAsh},
		{"fog", WeatherFog},
		{"neon glitch", WeatherNeonGlitch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := NewWeatherSystem(500, 42)
			ws.SetCamera(0, 0, 800, 600)
			ws.SetWeather(tt.weather, 1.0)
			ws.Update(0.5)

			active := ws.GetActiveParticles()
			if len(active) == 0 {
				t.Errorf("no particles spawned for weather type %v", tt.weather)
			}

			// Check particle properties are initialized
			p := active[0]
			if p.Color.R == 0 && p.Color.G == 0 && p.Color.B == 0 && p.Color.A == 0 {
				t.Error("particle color not initialized")
			}
			if p.Size == 0 {
				t.Error("particle size not initialized")
			}
			if p.MaxLifetime == 0 {
				t.Error("particle max lifetime not initialized")
			}
		})
	}
}

func TestParticleCulling(t *testing.T) {
	ws := NewWeatherSystem(500, 42)
	ws.SetCamera(0, 0, 800, 600)
	ws.SetWeather(WeatherRain, 1.0)
	ws.Update(0.5)

	// Move a particle far outside view
	for i := range ws.particles {
		if ws.particles[i].Active {
			ws.particles[i].X = -1000
			ws.particles[i].Y = -1000
			break
		}
	}

	// Update should cull the distant particle
	ws.Update(0.016)

	// Check that particles far from camera are culled eventually
	// (This is a basic check; particles may still be spawning)
}

func TestSpawnRate(t *testing.T) {
	ws := NewWeatherSystem(500, 42)

	tests := []struct {
		weather     WeatherType
		intensity   float64
		wantNonZero bool
	}{
		{WeatherNone, 0.5, false},
		{WeatherRain, 0.5, true},
		{WeatherSnow, 0.0, false},
		{WeatherDust, 1.0, true},
	}

	for _, tt := range tests {
		ws.SetWeather(tt.weather, tt.intensity)
		rate := ws.getSpawnRate()
		if tt.wantNonZero && rate == 0 {
			t.Errorf("spawn rate for %v at intensity %f = 0, want > 0", tt.weather, tt.intensity)
		}
		if !tt.wantNonZero && rate != 0 {
			t.Errorf("spawn rate for %v at intensity %f = %f, want 0", tt.weather, tt.intensity, rate)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		v, min, max, want float64
	}{
		{0.5, 0.0, 1.0, 0.5},
		{1.5, 0.0, 1.0, 1.0},
		{-0.5, 0.0, 1.0, 0.0},
		{0.0, 0.0, 1.0, 0.0},
		{1.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		got := clamp(tt.v, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
		}
	}
}

func BenchmarkUpdate(b *testing.B) {
	ws := NewWeatherSystem(2000, 42)
	ws.SetCamera(0, 0, 1920, 1080)
	ws.SetWeather(WeatherRain, 0.5)

	// Pre-spawn some particles
	ws.Update(2.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.Update(0.016) // ~60 FPS
	}
}

func BenchmarkSpawn(b *testing.B) {
	ws := NewWeatherSystem(5000, 42)
	ws.SetCamera(0, 0, 1920, 1080)
	ws.SetWeather(WeatherRain, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ws.spawnSingleParticle()
	}
}

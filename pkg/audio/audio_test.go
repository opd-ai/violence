package audio

import (
	"math"
	"testing"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
	if engine.sfxPlayers == nil {
		t.Error("sfxPlayers map not initialized")
	}
	if engine.intensity != 0.5 {
		t.Errorf("expected default intensity 0.5, got %f", engine.intensity)
	}
}

func TestPlayMusic(t *testing.T) {
	tests := []struct {
		name       string
		trackName  string
		intensity  float64
		wantLayers int
	}{
		{"low intensity", "combat", 0.2, 1},
		{"medium intensity", "combat", 0.5, 2},
		{"high intensity", "combat", 0.8, 3},
		{"max intensity", "combat", 1.0, 4},
		{"zero intensity", "combat", 0.0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			err := engine.PlayMusic(tt.trackName, tt.intensity)
			if err != nil {
				t.Fatalf("PlayMusic failed: %v", err)
			}
			if len(engine.musicLayers) < 1 {
				t.Error("expected at least 1 music layer")
			}
			if engine.intensity != tt.intensity {
				t.Errorf("expected intensity %f, got %f", tt.intensity, engine.intensity)
			}
		})
	}
}

func TestSetIntensity(t *testing.T) {
	tests := []struct {
		name     string
		initial  float64
		newValue float64
		expected float64
	}{
		{"increase intensity", 0.3, 0.7, 0.7},
		{"decrease intensity", 0.8, 0.4, 0.4},
		{"clamp high", 0.5, 1.5, 1.0},
		{"clamp low", 0.5, -0.2, 0.0},
		{"zero to max", 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			_ = engine.PlayMusic("test", tt.initial)
			engine.SetIntensity(tt.newValue)

			engine.mu.RLock()
			actual := engine.intensity
			engine.mu.RUnlock()

			if actual != tt.expected {
				t.Errorf("expected intensity %f, got %f", tt.expected, actual)
			}
		})
	}
}

func TestPlaySFX(t *testing.T) {
	tests := []struct {
		name string
		sfx  string
		x    float64
		y    float64
	}{
		{"gunshot near", "gunshot", 1.0, 1.0},
		{"footstep far", "footstep", 10.0, 10.0},
		{"door open", "door", 0.0, 0.0},
		{"pickup close", "pickup", 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			engine.SetListenerPosition(0.0, 0.0)
			err := engine.PlaySFX(tt.sfx, tt.x, tt.y)
			if err != nil {
				t.Fatalf("PlaySFX failed: %v", err)
			}
		})
	}
}

func TestSetListenerPosition(t *testing.T) {
	tests := []struct {
		name string
		x    float64
		y    float64
	}{
		{"origin", 0.0, 0.0},
		{"positive quadrant", 5.5, 7.2},
		{"negative quadrant", -3.0, -4.0},
		{"mixed", 10.0, -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			engine.SetListenerPosition(tt.x, tt.y)

			engine.mu.RLock()
			x, y := engine.listenerX, engine.listenerY
			engine.mu.RUnlock()

			if x != tt.x || y != tt.y {
				t.Errorf("expected position (%f, %f), got (%f, %f)", tt.x, tt.y, x, y)
			}
		})
	}
}

func TestSetGenre(t *testing.T) {
	tests := []struct {
		name    string
		genreID string
	}{
		{"fantasy", "fantasy"},
		{"scifi", "scifi"},
		{"horror", "horror"},
		{"cyberpunk", "cyberpunk"},
		{"postapoc", "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			engine.SetGenre(tt.genreID)

			engine.mu.RLock()
			genre := engine.genreID
			engine.mu.RUnlock()

			if genre != tt.genreID {
				t.Errorf("expected genre %s, got %s", tt.genreID, genre)
			}
		})
	}
}

func TestCalculateLayerVolume(t *testing.T) {
	tests := []struct {
		name      string
		layer     int
		intensity float64
		wantMin   float64
		wantMax   float64
	}{
		{"layer 1 off", 1, 0.2, 0.0, 0.1},
		{"layer 1 mid", 1, 0.45, 0.3, 0.7},
		{"layer 1 full", 1, 0.7, 0.9, 1.0},
		{"layer 2 off", 2, 0.4, 0.0, 0.1},
		{"layer 2 mid", 2, 0.65, 0.3, 0.7},
		{"layer 2 full", 2, 0.9, 0.9, 1.0},
		{"layer 3 off", 3, 0.6, 0.0, 0.1},
		{"layer 3 mid", 3, 0.85, 0.3, 0.7},
		{"layer 3 full", 3, 1.0, 0.9, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			volume := engine.calculateLayerVolume(tt.layer, tt.intensity)
			if volume < tt.wantMin || volume > tt.wantMax {
				t.Errorf("layer %d at intensity %f: volume %f not in range [%f, %f]",
					tt.layer, tt.intensity, volume, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateVolume(t *testing.T) {
	tests := []struct {
		name     string
		distance float64
		wantMin  float64
		wantMax  float64
	}{
		{"very close", 0.05, 0.9, 1.0},
		{"close", 1.0, 0.8, 1.0},
		{"medium", 5.0, 0.2, 0.5},
		{"far", 10.0, 0.0, 0.2},
		{"very far", 20.0, 0.0, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			volume := engine.calculateVolume(tt.distance)
			if volume < 0.0 || volume > 1.0 {
				t.Errorf("volume %f out of valid range [0.0, 1.0]", volume)
			}
			if volume < tt.wantMin || volume > tt.wantMax {
				t.Errorf("distance %f: volume %f not in expected range [%f, %f]",
					tt.distance, volume, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculatePan(t *testing.T) {
	tests := []struct {
		name    string
		dx      float64
		wantMin float64
		wantMax float64
	}{
		{"far left", -20.0, -1.0, -0.9},
		{"left", -5.0, -0.6, -0.4},
		{"center", 0.0, -0.1, 0.1},
		{"right", 5.0, 0.4, 0.6},
		{"far right", 20.0, 0.9, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			pan := engine.calculatePan(tt.dx)
			if pan < -1.0 || pan > 1.0 {
				t.Errorf("pan %f out of valid range [-1.0, 1.0]", pan)
			}
			if pan < tt.wantMin || pan > tt.wantMax {
				t.Errorf("dx %f: pan %f not in expected range [%f, %f]",
					tt.dx, pan, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name string
		v    float64
		min  float64
		max  float64
		want float64
	}{
		{"below min", -1.0, 0.0, 1.0, 0.0},
		{"at min", 0.0, 0.0, 1.0, 0.0},
		{"in range", 0.5, 0.0, 1.0, 0.5},
		{"at max", 1.0, 0.0, 1.0, 1.0},
		{"above max", 2.0, 0.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.v, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clamp(%f, %f, %f) = %f, want %f", tt.v, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

func TestSmoothstep(t *testing.T) {
	tests := []struct {
		name  string
		edge0 float64
		edge1 float64
		x     float64
		want  float64
	}{
		{"before edge0", 0.0, 1.0, -0.5, 0.0},
		{"at edge0", 0.0, 1.0, 0.0, 0.0},
		{"middle", 0.0, 1.0, 0.5, 0.5},
		{"at edge1", 0.0, 1.0, 1.0, 1.0},
		{"after edge1", 0.0, 1.0, 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := smoothstep(tt.edge0, tt.edge1, tt.x)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("smoothstep(%f, %f, %f) = %f, want %f", tt.edge0, tt.edge1, tt.x, got, tt.want)
			}
		})
	}
}

func TestSmoothstepSmooth(t *testing.T) {
	// Verify smoothstep produces smooth interpolation
	prev := smoothstep(0.0, 1.0, 0.0)
	for x := 0.1; x <= 1.0; x += 0.1 {
		curr := smoothstep(0.0, 1.0, x)
		if curr < prev {
			t.Errorf("smoothstep not monotonically increasing at x=%f", x)
		}
		prev = curr
	}
}

func TestGenerateSilence(t *testing.T) {
	samples := sampleRate
	data := generateSilence(samples)
	if len(data) < 44 {
		t.Fatal("generated data too short for WAV header")
	}
	// Check WAV header
	if string(data[0:4]) != "RIFF" {
		t.Error("missing RIFF header")
	}
	if string(data[8:12]) != "WAVE" {
		t.Error("missing WAVE header")
	}
}

func TestGenerateBlip(t *testing.T) {
	samples := sampleRate / 10
	data := generateBlip(samples)
	if len(data) < 44 {
		t.Fatal("generated data too short for WAV header")
	}
	// Check WAV header
	if string(data[0:4]) != "RIFF" {
		t.Error("missing RIFF header")
	}
	if string(data[8:12]) != "WAVE" {
		t.Error("missing WAVE header")
	}
}

func TestVolumeAttenuationFormula(t *testing.T) {
	// Test that volume decreases with distance
	engine := NewEngine()
	vol1 := engine.calculateVolume(1.0)
	vol2 := engine.calculateVolume(2.0)
	vol3 := engine.calculateVolume(5.0)

	if vol1 <= vol2 {
		t.Error("volume should decrease with distance")
	}
	if vol2 <= vol3 {
		t.Error("volume should decrease with distance")
	}
}

func TestPanCalculation(t *testing.T) {
	// Test that pan changes correctly with position
	engine := NewEngine()
	panLeft := engine.calculatePan(-10.0)
	panCenter := engine.calculatePan(0.0)
	panRight := engine.calculatePan(10.0)

	if panLeft >= panCenter {
		t.Error("left position should have negative pan")
	}
	if panRight <= panCenter {
		t.Error("right position should have positive pan")
	}
	if math.Abs(panCenter) > 0.1 {
		t.Error("center position should have near-zero pan")
	}
}

func TestGenreSwap(t *testing.T) {
	engine := NewEngine()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		engine.SetGenre(genre)
		engine.mu.RLock()
		current := engine.genreID
		engine.mu.RUnlock()

		if current != genre {
			t.Errorf("genre not set correctly: expected %s, got %s", genre, current)
		}
	}
}

func BenchmarkPlayMusic(b *testing.B) {
	engine := NewEngine()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.PlayMusic("combat", 0.5)
	}
}

func BenchmarkPlaySFX(b *testing.B) {
	engine := NewEngine()
	engine.SetListenerPosition(0.0, 0.0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.PlaySFX("gunshot", 5.0, 5.0)
	}
}

func BenchmarkCalculateVolume(b *testing.B) {
	engine := NewEngine()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.calculateVolume(float64(i % 20))
	}
}

func BenchmarkCalculatePan(b *testing.B) {
	engine := NewEngine()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.calculatePan(float64(i%20) - 10.0)
	}
}

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

func TestProceduralMusicGeneration(t *testing.T) {
	tests := []struct {
		name    string
		track   string
		layer   int
		genreID string
		minSize int
	}{
		{"fantasy base layer", "combat", 0, "fantasy", 1000},
		{"scifi layer 1", "ambient", 1, "scifi", 1000},
		{"horror layer 2", "boss", 2, "horror", 1000},
		{"cyberpunk layer 3", "stealth", 3, "cyberpunk", 1000},
		{"postapoc base", "explore", 0, "postapoc", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			engine.SetGenre(tt.genreID)

			data := engine.getMusicData(tt.track, tt.layer)
			if data == nil {
				t.Fatal("getMusicData returned nil")
			}
			if len(data) < tt.minSize {
				t.Errorf("music data too small: got %d bytes, want >= %d", len(data), tt.minSize)
			}

			// Verify WAV header
			if len(data) < 44 {
				t.Fatal("data too small for WAV header")
			}
			if string(data[0:4]) != "RIFF" {
				t.Error("missing RIFF header")
			}
			if string(data[8:12]) != "WAVE" {
				t.Error("missing WAVE header")
			}
		})
	}
}

func TestProceduralMusicDeterminism(t *testing.T) {
	engine := NewEngine()
	engine.SetGenre("fantasy")

	data1 := engine.getMusicData("combat", 0)
	data2 := engine.getMusicData("combat", 0)

	if len(data1) != len(data2) {
		t.Errorf("music data length mismatch: %d vs %d", len(data1), len(data2))
	}

	if len(data1) > 0 && len(data2) > 0 {
		match := true
		for i := 0; i < len(data1) && i < len(data2); i++ {
			if data1[i] != data2[i] {
				match = false
				break
			}
		}
		if !match {
			t.Error("music data not deterministic for same input")
		}
	}
}

func TestProceduralMusicGenreVariety(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	results := make(map[string][]byte)

	for _, genre := range genres {
		engine := NewEngine()
		engine.SetGenre(genre)
		results[genre] = engine.getMusicData("test", 0)
	}

	// Verify different genres produce different output
	foundDifferences := 0
	for i, g1 := range genres {
		for j, g2 := range genres {
			if i >= j {
				continue
			}

			d1 := results[g1]
			d2 := results[g2]

			if len(d1) == 0 || len(d2) == 0 {
				continue
			}

			// Compare PCM data starting after WAV header
			// Skip first 1000 bytes to avoid header
			// Check last 80% of file for differences
			start := 1000
			maxLen := len(d1)
			if len(d2) < maxLen {
				maxLen = len(d2)
			}

			if maxLen <= start {
				t.Errorf("music files too small: %d bytes", maxLen)
				continue
			}

			// Count differences
			diffCount := 0
			checkLen := maxLen - start
			for i := start; i < maxLen; i++ {
				if d1[i] != d2[i] {
					diffCount++
				}
			}

			// At least 5% of samples should be different between genres
			diffRatio := float64(diffCount) / float64(checkLen)
			if diffRatio > 0.05 {
				foundDifferences++
			} else {
				t.Logf("genres %s and %s have only %.2f%% differences", g1, g2, diffRatio*100)
			}
		}
	}

	// We should find differences in at least some genre pairs
	if foundDifferences == 0 {
		t.Error("no significant differences found between any genre pairs")
	}
}

func TestProceduralSFXGeneration(t *testing.T) {
	tests := []struct {
		name    string
		sfxName string
		minSize int
	}{
		{"gunshot", "gunshot", 1000},
		{"footstep", "footstep", 1000},
		{"door open", "door_open", 1000},
		{"explosion", "explosion", 1000},
		{"pickup item", "pickup_item", 1000},
		{"pain sound", "pain", 1000},
		{"reload", "reload", 1000},
		{"pistol fire", "pistol_fire", 1000},
		{"walk left", "walk_left", 1000},
		{"unknown sfx", "unknown_sfx", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine()
			data := engine.getSFXData(tt.sfxName)

			if data == nil {
				t.Fatal("getSFXData returned nil")
			}
			if len(data) < tt.minSize {
				t.Errorf("SFX data too small: got %d bytes, want >= %d", len(data), tt.minSize)
			}

			// Verify WAV header
			if len(data) < 44 {
				t.Fatal("data too small for WAV header")
			}
			if string(data[0:4]) != "RIFF" {
				t.Error("missing RIFF header")
			}
			if string(data[8:12]) != "WAVE" {
				t.Error("missing WAVE header")
			}
		})
	}
}

func TestProceduralSFXDeterminism(t *testing.T) {
	engine := NewEngine()

	sfxNames := []string{"gunshot", "footstep", "door_open", "explosion"}

	for _, name := range sfxNames {
		t.Run(name, func(t *testing.T) {
			data1 := engine.getSFXData(name)
			data2 := engine.getSFXData(name)

			if len(data1) != len(data2) {
				t.Errorf("SFX data length mismatch: %d vs %d", len(data1), len(data2))
			}

			if len(data1) > 0 && len(data2) > 0 {
				match := true
				for i := 0; i < len(data1) && i < len(data2); i++ {
					if data1[i] != data2[i] {
						match = false
						break
					}
				}
				if !match {
					t.Errorf("SFX data not deterministic for %s", name)
				}
			}
		})
	}
}

func TestProceduralSFXVariety(t *testing.T) {
	engine := NewEngine()
	sfxNames := []string{"gunshot", "footstep", "door_open", "explosion", "pickup"}

	results := make(map[string][]byte)
	for _, name := range sfxNames {
		results[name] = engine.getSFXData(name)
	}

	// Verify different SFX produce different output
	for i, name1 := range sfxNames {
		for j, name2 := range sfxNames {
			if i >= j {
				continue
			}

			d1 := results[name1]
			d2 := results[name2]

			if len(d1) == 0 || len(d2) == 0 {
				continue
			}

			// Compare a sample of the data
			sampleStart := 50
			sampleEnd := 200
			minLen := len(d1)
			if len(d2) < minLen {
				minLen = len(d2)
			}
			if minLen < sampleEnd {
				sampleEnd = minLen
			}
			if sampleStart >= sampleEnd {
				continue
			}

			identical := true
			for i := sampleStart; i < sampleEnd; i++ {
				if d1[i] != d2[i] {
					identical = false
					break
				}
			}

			if identical {
				t.Errorf("SFX %s and %s produced identical data", name1, name2)
			}
		}
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"simple", "test"},
		{"combat", "combat"},
		{"gunshot", "gunshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h1 := hashString(tt.input)
			h2 := hashString(tt.input)

			if h1 != h2 {
				t.Error("hashString not deterministic")
			}

			// Different inputs should produce different hashes
			if tt.input != "" {
				diff := hashString(tt.input + "x")
				if diff == h1 {
					t.Error("different inputs produced same hash")
				}
			}
		})
	}
}

func TestMidiToFreq(t *testing.T) {
	tests := []struct {
		midi int
		freq float64
	}{
		{69, 440.0},  // A4
		{60, 261.63}, // C4 (approximate)
		{57, 220.0},  // A3
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.midi)), func(t *testing.T) {
			freq := midiToFreq(tt.midi)
			tolerance := 0.5
			if math.Abs(freq-tt.freq) > tolerance {
				t.Errorf("midiToFreq(%d) = %.2f, want %.2f", tt.midi, freq, tt.freq)
			}
		})
	}
}

func TestADSREnvelope(t *testing.T) {
	totalSamples := 1000

	// Test attack phase
	val := adsrEnvelope(10, totalSamples, 0.1, 0.1, 0.7, 0.2)
	if val <= 0 || val > 1 {
		t.Errorf("attack phase value out of range: %f", val)
	}

	// Test sustain phase
	val = adsrEnvelope(500, totalSamples, 0.1, 0.1, 0.7, 0.2)
	if math.Abs(val-0.7) > 0.05 {
		t.Errorf("sustain phase value = %f, want 0.7", val)
	}

	// Test release phase
	val = adsrEnvelope(950, totalSamples, 0.1, 0.1, 0.7, 0.2)
	if val <= 0 || val >= 0.7 {
		t.Errorf("release phase value out of range: %f", val)
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		substrs  []string
		expected bool
	}{
		{"match single", "gunshot", []string{"gun", "fire"}, true},
		{"match multiple", "pistol_fire", []string{"gun", "pistol"}, true},
		{"no match", "footstep", []string{"gun", "door"}, false},
		{"exact match", "door", []string{"door"}, true},
		{"partial match", "door_open", []string{"door", "close"}, true},
		{"empty string", "", []string{"test"}, false},
		{"empty substrs", "test", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAny(tt.str, tt.substrs...)
			if result != tt.expected {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tt.str, tt.substrs, result, tt.expected)
			}
		})
	}
}

func TestLocalRNG(t *testing.T) {
	seed := uint64(12345)
	rng1 := newLocalRNG(seed)
	rng2 := newLocalRNG(seed)

	// Test determinism
	for i := 0; i < 10; i++ {
		v1 := rng1.Float64()
		v2 := rng2.Float64()

		if v1 != v2 {
			t.Errorf("RNG not deterministic at iteration %d", i)
		}

		if v1 < 0 || v1 >= 1 {
			t.Errorf("Float64() out of range: %f", v1)
		}
	}

	// Test Intn
	rng3 := newLocalRNG(seed)
	for i := 0; i < 10; i++ {
		v := rng3.Intn(100)
		if v < 0 || v >= 100 {
			t.Errorf("Intn(100) out of range: %d", v)
		}
	}
}

func BenchmarkGenerateMusic(b *testing.B) {
	seed := uint64(12345)
	samples := sampleRate * 2
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateMusic(seed, samples, "fantasy", 0)
	}
}

func BenchmarkGenerateSFX(b *testing.B) {
	seed := uint64(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		generateSFX(seed, "gunshot")
	}
}

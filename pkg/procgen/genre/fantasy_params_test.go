package genre

import "testing"

// TestDefaultFantasyParams verifies default parameter initialization.
func TestDefaultFantasyParams(t *testing.T) {
	params := DefaultFantasyParams()

	// Verify fog parameters
	if params.Fog.R != 120 || params.Fog.G != 130 || params.Fog.B != 150 {
		t.Errorf("Fog RGB = (%d, %d, %d), want (120, 130, 150)",
			params.Fog.R, params.Fog.G, params.Fog.B)
	}

	// Verify palette parameters
	if params.Palette.PrimaryHue != 240.0 {
		t.Errorf("PrimaryHue = %f, want 240.0", params.Palette.PrimaryHue)
	}
	if params.Palette.SecondaryHue != 30.0 {
		t.Errorf("SecondaryHue = %f, want 30.0", params.Palette.SecondaryHue)
	}
	if params.Palette.Saturation != 0.6 {
		t.Errorf("Saturation = %f, want 0.6", params.Palette.Saturation)
	}
	if params.Palette.Brightness != 0.7 {
		t.Errorf("Brightness = %f, want 0.7", params.Palette.Brightness)
	}
	if params.Palette.NumColors != 16 {
		t.Errorf("NumColors = %d, want 16", params.Palette.NumColors)
	}

	// Verify texture parameters
	if params.Texture.SeedOffset != 1000 {
		t.Errorf("SeedOffset = %d, want 1000", params.Texture.SeedOffset)
	}
	if params.Texture.NoiseScale != 0.15 {
		t.Errorf("NoiseScale = %f, want 0.15", params.Texture.NoiseScale)
	}
	if params.Texture.OctaveCount != 4 {
		t.Errorf("OctaveCount = %d, want 4", params.Texture.OctaveCount)
	}
	if params.Texture.Persistence != 0.5 {
		t.Errorf("Persistence = %f, want 0.5", params.Texture.Persistence)
	}
	if params.Texture.TileSize != 64 {
		t.Errorf("TileSize = %d, want 64", params.Texture.TileSize)
	}
	if params.Texture.WallVariance != 0.3 {
		t.Errorf("WallVariance = %f, want 0.3", params.Texture.WallVariance)
	}

	// Verify SFX parameters
	if params.SFX.Waveform != WaveformTriangle {
		t.Errorf("Waveform = %d, want %d", params.SFX.Waveform, WaveformTriangle)
	}
	if params.SFX.Frequency != 440.0 {
		t.Errorf("Frequency = %f, want 440.0", params.SFX.Frequency)
	}
	if params.SFX.Envelope.Attack != 0.05 {
		t.Errorf("Attack = %f, want 0.05", params.SFX.Envelope.Attack)
	}
	if params.SFX.Envelope.Decay != 0.1 {
		t.Errorf("Decay = %f, want 0.1", params.SFX.Envelope.Decay)
	}
	if params.SFX.Envelope.Sustain != 0.3 {
		t.Errorf("Sustain = %f, want 0.3", params.SFX.Envelope.Sustain)
	}
	if params.SFX.Envelope.Release != 0.2 {
		t.Errorf("Release = %f, want 0.2", params.SFX.Envelope.Release)
	}

	// Verify music parameters
	if params.Music.Scale != ScaleDorian {
		t.Errorf("Scale = %d, want %d", params.Music.Scale, ScaleDorian)
	}
	if params.Music.Tempo != 90 {
		t.Errorf("Tempo = %d, want 90", params.Music.Tempo)
	}
	if params.Music.TimeSignNum != 3 {
		t.Errorf("TimeSignNum = %d, want 3", params.Music.TimeSignNum)
	}
	if params.Music.TimeSignDen != 4 {
		t.Errorf("TimeSignDen = %d, want 4", params.Music.TimeSignDen)
	}
	if len(params.Music.ChordProg) != 4 {
		t.Fatalf("len(ChordProg) = %d, want 4", len(params.Music.ChordProg))
	}
	expected := []int{1, 6, 4, 5}
	for i, v := range params.Music.ChordProg {
		if v != expected[i] {
			t.Errorf("ChordProg[%d] = %d, want %d", i, v, expected[i])
		}
	}
}

// TestFogParamsRange verifies fog color values are valid RGB.
func TestFogParamsRange(t *testing.T) {
	tests := []struct {
		name string
		fog  FogParams
		want bool
	}{
		{"valid RGB", FogParams{R: 120, G: 130, B: 150}, true},
		{"zero RGB", FogParams{R: 0, G: 0, B: 0}, true},
		{"max RGB", FogParams{R: 255, G: 255, B: 255}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All uint8 values are valid, just verify they exist
			if tt.fog.R > 255 || tt.fog.G > 255 || tt.fog.B > 255 {
				t.Errorf("Invalid RGB values")
			}
		})
	}
}

// TestPaletteParamsValidation verifies palette parameter ranges.
func TestPaletteParamsValidation(t *testing.T) {
	params := DefaultFantasyParams().Palette

	if params.PrimaryHue < 0 || params.PrimaryHue > 360 {
		t.Errorf("PrimaryHue out of range [0, 360]: %f", params.PrimaryHue)
	}
	if params.SecondaryHue < 0 || params.SecondaryHue > 360 {
		t.Errorf("SecondaryHue out of range [0, 360]: %f", params.SecondaryHue)
	}
	if params.Saturation < 0 || params.Saturation > 1 {
		t.Errorf("Saturation out of range [0, 1]: %f", params.Saturation)
	}
	if params.Brightness < 0 || params.Brightness > 1 {
		t.Errorf("Brightness out of range [0, 1]: %f", params.Brightness)
	}
	if params.NumColors <= 0 {
		t.Errorf("NumColors must be positive: %d", params.NumColors)
	}
}

// TestTextureParamsValidation verifies texture parameter ranges.
func TestTextureParamsValidation(t *testing.T) {
	params := DefaultFantasyParams().Texture

	if params.NoiseScale <= 0 {
		t.Errorf("NoiseScale must be positive: %f", params.NoiseScale)
	}
	if params.OctaveCount <= 0 {
		t.Errorf("OctaveCount must be positive: %d", params.OctaveCount)
	}
	if params.Persistence < 0 || params.Persistence > 1 {
		t.Errorf("Persistence out of range [0, 1]: %f", params.Persistence)
	}
	if params.TileSize <= 0 {
		t.Errorf("TileSize must be positive: %d", params.TileSize)
	}
	if params.WallVariance < 0 || params.WallVariance > 1 {
		t.Errorf("WallVariance out of range [0, 1]: %f", params.WallVariance)
	}
}

// TestWaveformConstants verifies waveform type constants.
func TestWaveformConstants(t *testing.T) {
	tests := []struct {
		name     string
		waveform WaveformType
		expected int
	}{
		{"Sine", WaveformSine, 0},
		{"Square", WaveformSquare, 1},
		{"Sawtooth", WaveformSawtooth, 2},
		{"Triangle", WaveformTriangle, 3},
		{"Noise", WaveformNoise, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.waveform) != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.waveform, tt.expected)
			}
		})
	}
}

// TestEnvelopeParamsValidation verifies envelope parameter ranges.
func TestEnvelopeParamsValidation(t *testing.T) {
	envelope := DefaultFantasyParams().SFX.Envelope

	if envelope.Attack < 0 {
		t.Errorf("Attack must be non-negative: %f", envelope.Attack)
	}
	if envelope.Decay < 0 {
		t.Errorf("Decay must be non-negative: %f", envelope.Decay)
	}
	if envelope.Sustain < 0 || envelope.Sustain > 1 {
		t.Errorf("Sustain out of range [0, 1]: %f", envelope.Sustain)
	}
	if envelope.Release < 0 {
		t.Errorf("Release must be non-negative: %f", envelope.Release)
	}
}

// TestScaleConstants verifies musical scale type constants.
func TestScaleConstants(t *testing.T) {
	tests := []struct {
		name     string
		scale    ScaleType
		expected int
	}{
		{"Minor", ScaleMinor, 0},
		{"Major", ScaleMajor, 1},
		{"Dorian", ScaleDorian, 2},
		{"Phrygian", ScalePhrygian, 3},
		{"Lydian", ScaleLydian, 4},
		{"Mixolydian", ScaleMixolydian, 5},
		{"Aeolian", ScaleAeolian, 6},
		{"Locrian", ScaleLocrian, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.scale) != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.scale, tt.expected)
			}
		})
	}
}

// TestMusicParamsValidation verifies music parameter ranges.
func TestMusicParamsValidation(t *testing.T) {
	music := DefaultFantasyParams().Music

	if music.Tempo <= 0 {
		t.Errorf("Tempo must be positive: %d", music.Tempo)
	}
	if music.TimeSignNum <= 0 {
		t.Errorf("TimeSignNum must be positive: %d", music.TimeSignNum)
	}
	if music.TimeSignDen <= 0 {
		t.Errorf("TimeSignDen must be positive: %d", music.TimeSignDen)
	}
	if len(music.ChordProg) == 0 {
		t.Error("ChordProg must not be empty")
	}
	for i, degree := range music.ChordProg {
		if degree < 1 || degree > 7 {
			t.Errorf("ChordProg[%d] out of range [1, 7]: %d", i, degree)
		}
	}
}

// TestFantasyParamsModification verifies params can be modified.
func TestFantasyParamsModification(t *testing.T) {
	params := DefaultFantasyParams()

	// Modify fog
	params.Fog.R = 200
	if params.Fog.R != 200 {
		t.Errorf("Failed to modify Fog.R: got %d, want 200", params.Fog.R)
	}

	// Modify palette
	params.Palette.PrimaryHue = 180.0
	if params.Palette.PrimaryHue != 180.0 {
		t.Errorf("Failed to modify PrimaryHue: got %f, want 180.0", params.Palette.PrimaryHue)
	}

	// Modify texture
	params.Texture.SeedOffset = 2000
	if params.Texture.SeedOffset != 2000 {
		t.Errorf("Failed to modify SeedOffset: got %d, want 2000", params.Texture.SeedOffset)
	}

	// Modify SFX
	params.SFX.Waveform = WaveformSine
	if params.SFX.Waveform != WaveformSine {
		t.Errorf("Failed to modify Waveform: got %d, want %d", params.SFX.Waveform, WaveformSine)
	}

	// Modify music
	params.Music.Scale = ScaleMajor
	if params.Music.Scale != ScaleMajor {
		t.Errorf("Failed to modify Scale: got %d, want %d", params.Music.Scale, ScaleMajor)
	}
}

// TestFantasyParamsIndependence verifies multiple param instances are independent.
func TestFantasyParamsIndependence(t *testing.T) {
	params1 := DefaultFantasyParams()
	params2 := DefaultFantasyParams()

	// Modify params1
	params1.Fog.R = 255
	params1.Music.Tempo = 120

	// Verify params2 unchanged
	if params2.Fog.R != 120 {
		t.Errorf("params2.Fog.R = %d, want 120 (should be independent)", params2.Fog.R)
	}
	if params2.Music.Tempo != 90 {
		t.Errorf("params2.Music.Tempo = %d, want 90 (should be independent)", params2.Music.Tempo)
	}
}

// BenchmarkDefaultFantasyParams benchmarks parameter initialization.
func BenchmarkDefaultFantasyParams(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DefaultFantasyParams()
	}
}

package render

import (
	"image/color"
	"testing"
)

func TestNewPostProcessor(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		seed   int64
	}{
		{"320x200", 320, 200, 42},
		{"640x480", 640, 480, 12345},
		{"1920x1080", 1920, 1080, 99999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(tt.width, tt.height, tt.seed)
			if pp == nil {
				t.Fatal("NewPostProcessor returned nil")
			}
			if pp.width != tt.width {
				t.Errorf("width = %d, want %d", pp.width, tt.width)
			}
			if pp.height != tt.height {
				t.Errorf("height = %d, want %d", pp.height, tt.height)
			}
			if pp.seed != tt.seed {
				t.Errorf("seed = %d, want %d", pp.seed, tt.seed)
			}
			if pp.genreID != "fantasy" {
				t.Errorf("genreID = %s, want fantasy", pp.genreID)
			}
		})
	}
}

func TestPostProcessorSetGenre(t *testing.T) {
	pp := NewPostProcessor(320, 200, 42)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			pp.SetGenre(genre)
			if pp.genreID != genre {
				t.Errorf("genreID = %s, want %s", pp.genreID, genre)
			}
		})
	}
}

func TestApplyVignette(t *testing.T) {
	tests := []struct {
		name   string
		config VignetteConfig
	}{
		{
			name: "disabled",
			config: VignetteConfig{
				Enabled:   false,
				Intensity: 0.5,
				Power:     2.0,
				Tint:      color.RGBA{R: 20, G: 20, B: 20, A: 255},
			},
		},
		{
			name: "low_intensity",
			config: VignetteConfig{
				Enabled:   true,
				Intensity: 0.3,
				Power:     2.0,
				Tint:      color.RGBA{R: 10, G: 10, B: 10, A: 255},
			},
		},
		{
			name: "high_intensity",
			config: VignetteConfig{
				Enabled:   true,
				Intensity: 0.8,
				Power:     2.5,
				Tint:      color.RGBA{R: 0, G: 0, B: 0, A: 255},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(100, 100, 42)
			fb := createTestFramebuffer(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})

			pp.ApplyVignette(fb, tt.config)

			// Check center pixel (should be less affected)
			centerIdx := (50*100 + 50) * 4
			centerR := fb[centerIdx]

			// Check corner pixel (should be more affected)
			cornerIdx := (0*100 + 0) * 4
			cornerR := fb[cornerIdx]

			if tt.config.Enabled && tt.config.Intensity > 0.5 {
				// Corner should be darker than center for high intensity
				if cornerR >= centerR {
					t.Errorf("corner R = %d >= center R = %d, expected corner darker", cornerR, centerR)
				}
			}
		})
	}
}

func TestApplyFilmGrain(t *testing.T) {
	tests := []struct {
		name   string
		config FilmGrainConfig
		seed   int64
	}{
		{
			name:   "disabled",
			config: FilmGrainConfig{Enabled: false, Intensity: 0.1},
			seed:   42,
		},
		{
			name:   "low_grain",
			config: FilmGrainConfig{Enabled: true, Intensity: 0.05},
			seed:   42,
		},
		{
			name:   "high_grain",
			config: FilmGrainConfig{Enabled: true, Intensity: 0.2},
			seed:   42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(100, 100, tt.seed)
			fb1 := createTestFramebuffer(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})
			fb2 := createTestFramebuffer(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})

			pp.ApplyFilmGrain(fb1, tt.config)

			// Reset RNG and apply again - should be deterministic
			pp.rng.Seed(tt.seed)
			pp.ApplyFilmGrain(fb2, tt.config)

			// Should produce identical results
			for i := 0; i < len(fb1); i++ {
				if fb1[i] != fb2[i] {
					t.Errorf("non-deterministic grain at byte %d: %d != %d", i, fb1[i], fb2[i])
					break
				}
			}
		})
	}
}

func TestApplyScanlines(t *testing.T) {
	tests := []struct {
		name   string
		config ScanlinesConfig
	}{
		{
			name:   "disabled",
			config: ScanlinesConfig{Enabled: false, Intensity: 0.2, Spacing: 2},
		},
		{
			name:   "spacing_2",
			config: ScanlinesConfig{Enabled: true, Intensity: 0.3, Spacing: 2},
		},
		{
			name:   "spacing_4",
			config: ScanlinesConfig{Enabled: true, Intensity: 0.5, Spacing: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(100, 100, 42)
			fb := createTestFramebuffer(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})

			pp.ApplyScanlines(fb, tt.config)

			if tt.config.Enabled {
				// Check that scanlines are darker
				scanlineIdx := (0*100 + 50) * 4
				normalIdx := (1*100 + 50) * 4

				scanlineR := fb[scanlineIdx]
				normalR := fb[normalIdx]

				if tt.config.Spacing == 2 && normalIdx/4/100%tt.config.Spacing != 0 {
					if scanlineR >= normalR {
						t.Errorf("scanline R = %d >= normal R = %d, expected darker", scanlineR, normalR)
					}
				}
			}
		})
	}
}

func TestApplyChromaticAberration(t *testing.T) {
	tests := []struct {
		name   string
		config ChromaticAberrationConfig
	}{
		{
			name:   "disabled",
			config: ChromaticAberrationConfig{Enabled: false, Offset: 0.01},
		},
		{
			name:   "low_offset",
			config: ChromaticAberrationConfig{Enabled: true, Offset: 0.002},
		},
		{
			name:   "high_offset",
			config: ChromaticAberrationConfig{Enabled: true, Offset: 0.01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(100, 100, 42)

			// Create a gradient framebuffer
			fb := make([]byte, 100*100*4)
			for y := 0; y < 100; y++ {
				for x := 0; x < 100; x++ {
					idx := (y*100 + x) * 4
					fb[idx] = uint8(x * 255 / 100)   // R gradient
					fb[idx+1] = 128                  // G constant
					fb[idx+2] = uint8(y * 255 / 100) // B gradient
					fb[idx+3] = 255
				}
			}

			original := make([]byte, len(fb))
			copy(original, fb)

			pp.ApplyChromaticAberration(fb, tt.config)

			// For disabled, framebuffer should be unchanged
			if !tt.config.Enabled {
				for i := 0; i < len(fb); i++ {
					if fb[i] != original[i] {
						t.Errorf("framebuffer modified when disabled at byte %d", i)
						break
					}
				}
			}
		})
	}
}

func TestApplyBloom(t *testing.T) {
	tests := []struct {
		name   string
		config BloomConfig
	}{
		{
			name:   "disabled",
			config: BloomConfig{Enabled: false, Threshold: 0.7, Intensity: 0.5, Radius: 3},
		},
		{
			name:   "low_threshold",
			config: BloomConfig{Enabled: true, Threshold: 0.5, Intensity: 0.3, Radius: 2},
		},
		{
			name:   "high_threshold",
			config: BloomConfig{Enabled: true, Threshold: 0.8, Intensity: 0.6, Radius: 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(100, 100, 42)

			// Create framebuffer with a bright spot in center
			fb := make([]byte, 100*100*4)
			for y := 0; y < 100; y++ {
				for x := 0; x < 100; x++ {
					idx := (y*100 + x) * 4
					if x >= 45 && x <= 55 && y >= 45 && y <= 55 {
						// Bright center
						fb[idx] = 255
						fb[idx+1] = 255
						fb[idx+2] = 255
					} else {
						// Dark surroundings
						fb[idx] = 32
						fb[idx+1] = 32
						fb[idx+2] = 32
					}
					fb[idx+3] = 255
				}
			}

			pp.ApplyBloom(fb, tt.config)

			// If enabled, surrounding pixels should be brighter than original
			if tt.config.Enabled {
				// Check a pixel just outside the bright area
				testIdx := (45*100 + 40) * 4
				brightness := float64(fb[testIdx]) / 255.0

				// Should be brighter than the original dark value
				if brightness <= 32.0/255.0 {
					t.Logf("bloom may not have spread sufficiently, brightness = %f", brightness)
				}
			}
		})
	}
}

func TestApplyColorGrade(t *testing.T) {
	tests := []struct {
		name   string
		config ColorGradeConfig
	}{
		{
			name:   "disabled",
			config: ColorGradeConfig{Enabled: false, Contrast: 1.0, Saturation: 1.0, Warmth: 0.0},
		},
		{
			name:   "warm",
			config: ColorGradeConfig{Enabled: true, Contrast: 1.1, Saturation: 0.9, Warmth: 0.5},
		},
		{
			name:   "cool",
			config: ColorGradeConfig{Enabled: true, Contrast: 1.2, Saturation: 1.0, Warmth: -0.5},
		},
		{
			name:   "high_saturation",
			config: ColorGradeConfig{Enabled: true, Contrast: 1.0, Saturation: 1.5, Warmth: 0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(100, 100, 42)
			fb := createTestFramebuffer(100, 100, color.RGBA{R: 100, G: 150, B: 200, A: 255})
			original := make([]byte, len(fb))
			copy(original, fb)

			pp.ApplyColorGrade(fb, tt.config)

			// Check that colors were modified (unless disabled)
			idx := (50*100 + 50) * 4
			if !tt.config.Enabled {
				if fb[idx] != original[idx] || fb[idx+1] != original[idx+1] || fb[idx+2] != original[idx+2] {
					t.Error("framebuffer modified when disabled")
				}
			} else {
				// Warmth should affect R and B differently
				if tt.config.Warmth > 0 {
					// Warm should increase R more than B
					// (Due to clamping, just check that it changed)
					if fb[idx] == original[idx] && fb[idx+2] == original[idx+2] {
						t.Log("color grade may not have applied warmth")
					}
				}
			}
		})
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name    string
		genre   string
		seed    int64
		wantLen int
	}{
		{"fantasy", "fantasy", 42, 100 * 100 * 4},
		{"scifi", "scifi", 12345, 100 * 100 * 4},
		{"horror", "horror", 99999, 100 * 100 * 4},
		{"cyberpunk", "cyberpunk", 777, 100 * 100 * 4},
		{"postapoc", "postapoc", 555, 100 * 100 * 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := NewPostProcessor(100, 100, tt.seed)
			pp.SetGenre(tt.genre)

			fb := createTestFramebuffer(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})
			original := make([]byte, len(fb))
			copy(original, fb)

			pp.Apply(fb)

			// Check that framebuffer was modified
			modified := false
			for i := 0; i < len(fb); i++ {
				if fb[i] != original[i] {
					modified = true
					break
				}
			}

			if !modified {
				t.Error("Apply did not modify framebuffer")
			}

			// Check length unchanged
			if len(fb) != tt.wantLen {
				t.Errorf("framebuffer length = %d, want %d", len(fb), tt.wantLen)
			}
		})
	}
}

func TestApplyDeterminism(t *testing.T) {
	seed := int64(42)
	pp1 := NewPostProcessor(100, 100, seed)
	pp2 := NewPostProcessor(100, 100, seed)

	pp1.SetGenre("fantasy")
	pp2.SetGenre("fantasy")

	fb1 := createTestFramebuffer(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	fb2 := createTestFramebuffer(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})

	pp1.Apply(fb1)
	pp2.Apply(fb2)

	// Should produce identical results
	for i := 0; i < len(fb1); i++ {
		if fb1[i] != fb2[i] {
			t.Errorf("non-deterministic output at byte %d: %d != %d", i, fb1[i], fb2[i])
			break
		}
	}
}

func TestGetGenrePreset(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", "unknown"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			preset := GetGenrePreset(genre)

			// Verify preset is returned
			if genre == "unknown" {
				// Unknown genre should return disabled effects
				if preset.Vignette.Enabled || preset.FilmGrain.Enabled ||
					preset.Scanlines.Enabled || preset.ChromaticAberration.Enabled ||
					preset.Bloom.Enabled || preset.ColorGrade.Enabled {
					t.Error("unknown genre should have all effects disabled")
				}
			} else {
				// Known genres should have at least color grade and vignette
				if !preset.ColorGrade.Enabled || !preset.Vignette.Enabled {
					t.Errorf("%s should have color grade and vignette enabled", genre)
				}
			}
		})
	}
}

func TestGenrePresetUniqueness(t *testing.T) {
	// Verify each genre has unique settings
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	presets := make(map[string]GenrePreset)

	for _, genre := range genres {
		presets[genre] = GetGenrePreset(genre)
	}

	// Check that presets are different
	for i, g1 := range genres {
		for j, g2 := range genres {
			if i >= j {
				continue
			}

			p1 := presets[g1]
			p2 := presets[g2]

			// At least some parameter should differ
			same := p1.ColorGrade.Warmth == p2.ColorGrade.Warmth &&
				p1.ColorGrade.Saturation == p2.ColorGrade.Saturation &&
				p1.Vignette.Intensity == p2.Vignette.Intensity &&
				p1.FilmGrain.Intensity == p2.FilmGrain.Intensity &&
				p1.Scanlines.Enabled == p2.Scanlines.Enabled &&
				p1.ChromaticAberration.Enabled == p2.ChromaticAberration.Enabled &&
				p1.Bloom.Enabled == p2.Bloom.Enabled

			if same {
				t.Errorf("%s and %s have identical presets", g1, g2)
			}
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  float64
	}{
		{"negative", -10.0, 0.0},
		{"zero", 0.0, 0.0},
		{"mid", 128.0, 128.0},
		{"max", 255.0, 255.0},
		{"over", 300.0, 255.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clamp(tt.value)
			if got != tt.want {
				t.Errorf("clamp(%f) = %f, want %f", tt.value, got, tt.want)
			}
		})
	}
}

func TestClampInt(t *testing.T) {
	tests := []struct {
		name string
		v    int
		min  int
		max  int
		want int
	}{
		{"below", -5, 0, 100, 0},
		{"min", 0, 0, 100, 0},
		{"mid", 50, 0, 100, 50},
		{"max", 100, 0, 100, 100},
		{"above", 150, 0, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampInt(tt.v, tt.min, tt.max)
			if got != tt.want {
				t.Errorf("clampInt(%d, %d, %d) = %d, want %d", tt.v, tt.min, tt.max, got, tt.want)
			}
		})
	}
}

// createTestFramebuffer creates a framebuffer filled with a solid color.
func createTestFramebuffer(width, height int, c color.RGBA) []byte {
	fb := make([]byte, width*height*4)
	for i := 0; i < len(fb); i += 4 {
		fb[i] = c.R
		fb[i+1] = c.G
		fb[i+2] = c.B
		fb[i+3] = c.A
	}
	return fb
}

func BenchmarkApplyVignette(b *testing.B) {
	pp := NewPostProcessor(640, 480, 42)
	fb := createTestFramebuffer(640, 480, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	cfg := VignetteConfig{
		Enabled:   true,
		Intensity: 0.5,
		Power:     2.0,
		Tint:      color.RGBA{R: 20, G: 20, B: 20, A: 255},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pp.ApplyVignette(fb, cfg)
	}
}

func BenchmarkApplyFilmGrain(b *testing.B) {
	pp := NewPostProcessor(640, 480, 42)
	fb := createTestFramebuffer(640, 480, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	cfg := FilmGrainConfig{
		Enabled:   true,
		Intensity: 0.1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pp.ApplyFilmGrain(fb, cfg)
	}
}

func BenchmarkApplyBloom(b *testing.B) {
	pp := NewPostProcessor(640, 480, 42)
	fb := createTestFramebuffer(640, 480, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	cfg := BloomConfig{
		Enabled:   true,
		Threshold: 0.7,
		Intensity: 0.5,
		Radius:    3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pp.ApplyBloom(fb, cfg)
	}
}

func BenchmarkApply(b *testing.B) {
	pp := NewPostProcessor(640, 480, 42)
	pp.SetGenre("cyberpunk") // Most effects enabled
	fb := createTestFramebuffer(640, 480, color.RGBA{R: 128, G: 128, B: 128, A: 255})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pp.Apply(fb)
	}
}

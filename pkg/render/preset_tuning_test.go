package render

import (
	"testing"
)

// TestGenrePresetVisualBalance validates that genre presets maintain visual balance
// without obscuring gameplay. This implements Step 23 of PLAN.md.
func TestGenrePresetVisualBalance(t *testing.T) {
	tests := []struct {
		genreID       string
		maxVignette   float64 // Maximum acceptable vignette intensity
		maxGrain      float64 // Maximum acceptable grain intensity
		minContrast   float64 // Minimum contrast for visibility
		maxContrast   float64 // Maximum contrast before clipping
		minSaturation float64 // Minimum saturation to preserve color info
	}{
		{
			genreID:       "fantasy",
			maxVignette:   0.6, // Warm, atmospheric but not too dark
			maxGrain:      0.12,
			minContrast:   0.9,
			maxContrast:   1.3,
			minSaturation: 0.7, // Preserve color for readability
		},
		{
			genreID:       "scifi",
			maxVignette:   0.5, // Clean, technological aesthetic
			maxGrain:      0.08,
			minContrast:   1.0,
			maxContrast:   1.4,
			minSaturation: 0.8,
		},
		{
			genreID:       "horror",
			maxVignette:   0.8, // Dark but gameplay-visible
			maxGrain:      0.18,
			minContrast:   1.1, // High contrast for tension
			maxContrast:   1.5,
			minSaturation: 0.5, // Desaturation acceptable for mood
		},
		{
			genreID:       "cyberpunk",
			maxVignette:   0.5, // Bright neon focus
			maxGrain:      0.1,
			minContrast:   1.0,
			maxContrast:   1.3,
			minSaturation: 1.2, // High saturation for neon
		},
		{
			genreID:       "postapoc",
			maxVignette:   0.6, // Dusty, worn atmosphere
			maxGrain:      0.2, // Heavy grain acceptable for dust
			minContrast:   0.8, // Lower contrast for washed-out look
			maxContrast:   1.2,
			minSaturation: 0.6, // Faded colors
		},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			preset := GetGenrePreset(tt.genreID)

			// Validate vignette doesn't obscure too much
			if preset.Vignette.Enabled && preset.Vignette.Intensity > tt.maxVignette {
				t.Errorf("%s vignette too intense: %.2f > %.2f (may obscure gameplay)",
					tt.genreID, preset.Vignette.Intensity, tt.maxVignette)
			}

			// Validate film grain is subtle
			if preset.FilmGrain.Enabled && preset.FilmGrain.Intensity > tt.maxGrain {
				t.Errorf("%s grain too heavy: %.2f > %.2f (may obscure details)",
					tt.genreID, preset.FilmGrain.Intensity, tt.maxGrain)
			}

			// Validate contrast preserves visibility
			if preset.ColorGrade.Enabled {
				if preset.ColorGrade.Contrast < tt.minContrast {
					t.Errorf("%s contrast too low: %.2f < %.2f (reduces visibility)",
						tt.genreID, preset.ColorGrade.Contrast, tt.minContrast)
				}
				if preset.ColorGrade.Contrast > tt.maxContrast {
					t.Errorf("%s contrast too high: %.2f > %.2f (may cause clipping)",
						tt.genreID, preset.ColorGrade.Contrast, tt.maxContrast)
				}
			}

			// Validate saturation preserves color information
			if preset.ColorGrade.Enabled && preset.ColorGrade.Saturation < tt.minSaturation {
				t.Errorf("%s saturation too low: %.2f < %.2f (reduces color readability)",
					tt.genreID, preset.ColorGrade.Saturation, tt.minSaturation)
			}
		})
	}
}

// TestScanlinesReadability ensures scanlines don't interfere with gameplay visibility.
func TestScanlinesReadability(t *testing.T) {
	preset := GetGenrePreset("scifi")

	if preset.Scanlines.Enabled {
		// Scanlines should be subtle and spaced appropriately
		if preset.Scanlines.Intensity > 0.25 {
			t.Errorf("scanlines too dark: %.2f > 0.25", preset.Scanlines.Intensity)
		}
		if preset.Scanlines.Spacing < 2 {
			t.Errorf("scanlines too dense: %d < 2 (minimum spacing)", preset.Scanlines.Spacing)
		}
	}
}

// TestChromaticAberrationSubtle ensures aberration doesn't cause eye strain.
func TestChromaticAberrationSubtle(t *testing.T) {
	genres := []string{"scifi", "cyberpunk"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			preset := GetGenrePreset(genreID)

			if preset.ChromaticAberration.Enabled {
				// Aberration should be subtle (< 0.5% of screen width)
				if preset.ChromaticAberration.Offset > 0.005 {
					t.Errorf("%s aberration too strong: %.4f > 0.005 (may cause eye strain)",
						genreID, preset.ChromaticAberration.Offset)
				}
			}
		})
	}
}

// TestBloomPerformance ensures bloom parameters won't tank performance.
func TestBloomPerformance(t *testing.T) {
	preset := GetGenrePreset("cyberpunk")

	if preset.Bloom.Enabled {
		// Bloom radius affects performance (box blur is O(n²) per pixel)
		// At 320x200, radius > 5 would be ~25x cost per pixel
		if preset.Bloom.Radius > 5 {
			t.Errorf("bloom radius too large: %d > 5 (performance concern at 320x200)",
				preset.Bloom.Radius)
		}

		// Threshold should be high enough to limit bloom to bright areas
		if preset.Bloom.Threshold < 0.6 {
			t.Errorf("bloom threshold too low: %.2f < 0.6 (blooms too much, impacts performance)",
				preset.Bloom.Threshold)
		}

		// Intensity should be noticeable but not overwhelming
		if preset.Bloom.Intensity > 1.0 {
			t.Errorf("bloom intensity too high: %.2f > 1.0 (overpowers scene)",
				preset.Bloom.Intensity)
		}
	}
}

// TestStaticBurstFrequency ensures horror static doesn't trigger too often.
func TestStaticBurstFrequency(t *testing.T) {
	preset := GetGenrePreset("horror")

	if preset.StaticBurst.Enabled {
		// At 30 FPS, 1% chance = ~once every 3 seconds
		// At 60 FPS, 1% chance = ~once every 1.5 seconds
		// Maximum 2% to avoid annoyance
		if preset.StaticBurst.Probability > 0.02 {
			t.Errorf("static burst too frequent: %.3f > 0.02 (2%% max)",
				preset.StaticBurst.Probability)
		}

		// Duration should be brief (2-5 frames at 30 FPS = 66-166ms)
		if preset.StaticBurst.Duration > 5 {
			t.Errorf("static burst too long: %d > 5 frames", preset.StaticBurst.Duration)
		}
		if preset.StaticBurst.Duration < 1 {
			t.Errorf("static burst too short: %d < 1 frame", preset.StaticBurst.Duration)
		}

		// Intensity should be noticeable but not blinding
		if preset.StaticBurst.Intensity > 0.9 {
			t.Errorf("static burst too intense: %.2f > 0.9", preset.StaticBurst.Intensity)
		}
	}
}

// TestFilmScratchesDensity ensures scratches don't clutter the screen.
func TestFilmScratchesDensity(t *testing.T) {
	preset := GetGenrePreset("postapoc")

	if preset.FilmScratches.Enabled {
		// At 320 width, 2% = ~6 scratches per frame
		// Maximum 5% to avoid visual clutter
		if preset.FilmScratches.Density > 0.05 {
			t.Errorf("film scratches too dense: %.3f > 0.05 (5%% max)",
				preset.FilmScratches.Density)
		}

		// Length should be partial screen (not full vertical lines)
		if preset.FilmScratches.Length > 0.8 {
			t.Errorf("film scratches too long: %.2f > 0.8 (80%% screen height max)",
				preset.FilmScratches.Length)
		}
		if preset.FilmScratches.Length < 0.2 {
			t.Errorf("film scratches too short: %.2f < 0.2 (barely visible)",
				preset.FilmScratches.Length)
		}
	}
}

// TestWarmthBalance ensures warmth doesn't oversaturate or desaturate colors.
func TestWarmthBalance(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			preset := GetGenrePreset(genreID)

			if preset.ColorGrade.Enabled {
				// Warmth should be subtle (-0.5 to +0.5 range)
				if preset.ColorGrade.Warmth < -0.5 || preset.ColorGrade.Warmth > 0.5 {
					t.Errorf("%s warmth out of bounds: %.2f (should be in [-0.5, 0.5])",
						genreID, preset.ColorGrade.Warmth)
				}
			}
		})
	}
}

// TestVignetteTintNotBlack ensures vignette tints preserve some color.
func TestVignetteTintNotBlack(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genreID := range genres {
		t.Run(genreID, func(t *testing.T) {
			preset := GetGenrePreset(genreID)

			if preset.Vignette.Enabled {
				tint := preset.Vignette.Tint
				// At least one channel should be > 0 (not pure black)
				if tint.R == 0 && tint.G == 0 && tint.B == 0 {
					t.Errorf("%s vignette tint is pure black (should have some color)", genreID)
				}

				// No channel should be too bright (defeats vignette purpose)
				if tint.R > 100 || tint.G > 100 || tint.B > 100 {
					t.Errorf("%s vignette tint too bright: RGB(%d,%d,%d)",
						genreID, tint.R, tint.G, tint.B)
				}
			}
		})
	}
}

// TestGenreDistinctiveness validates that genres have measurably different presets.
func TestGenreDistinctiveness(t *testing.T) {
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	presets := make(map[string]GenrePreset)

	for _, genreID := range genres {
		presets[genreID] = GetGenrePreset(genreID)
	}

	// Fantasy should be warm
	if presets["fantasy"].ColorGrade.Warmth <= 0 {
		t.Errorf("fantasy should have warm color grade (got %.2f)",
			presets["fantasy"].ColorGrade.Warmth)
	}

	// Scifi should be cool
	if presets["scifi"].ColorGrade.Warmth >= 0 {
		t.Errorf("scifi should have cool color grade (got %.2f)",
			presets["scifi"].ColorGrade.Warmth)
	}

	// Horror should have heavy vignette
	if presets["horror"].Vignette.Intensity < 0.6 {
		t.Errorf("horror vignette should be intense (got %.2f)",
			presets["horror"].Vignette.Intensity)
	}

	// Cyberpunk should have bloom
	if !presets["cyberpunk"].Bloom.Enabled {
		t.Error("cyberpunk should have bloom enabled")
	}

	// Postapoc should have film scratches
	if !presets["postapoc"].FilmScratches.Enabled {
		t.Error("postapoc should have film scratches enabled")
	}

	// Only horror should have static burst
	for _, genreID := range genres {
		hasStatic := presets[genreID].StaticBurst.Enabled
		if genreID == "horror" && !hasStatic {
			t.Error("horror should have static burst enabled")
		}
		if genreID != "horror" && hasStatic {
			t.Errorf("%s should not have static burst (horror-only)", genreID)
		}
	}
}

// Package render provides post-processing effects for the rendering pipeline.
package render

import (
	"image/color"
	"math"
	"math/rand"
)

// PostProcessor applies post-processing effects to a framebuffer.
type PostProcessor struct {
	width            int
	height           int
	seed             int64
	rng              *rand.Rand
	genreID          string
	staticBurstTimer int // Frame counter for static burst timing
}

// NewPostProcessor creates a post-processor for the given dimensions.
func NewPostProcessor(width, height int, seed int64) *PostProcessor {
	return &PostProcessor{
		width:   width,
		height:  height,
		seed:    seed,
		rng:     rand.New(rand.NewSource(seed)),
		genreID: "fantasy",
	}
}

// SetGenre configures the post-processor for a genre.
func (p *PostProcessor) SetGenre(genreID string) {
	p.genreID = genreID
}

// Apply applies the genre-specific post-processing effects to a framebuffer.
// The framebuffer is modified in place (RGBA format: width*height*4 bytes).
func (p *PostProcessor) Apply(framebuffer []byte) {
	preset := GetGenrePreset(p.genreID)

	// Reset RNG for deterministic effects
	p.rng = rand.New(rand.NewSource(p.seed))

	// Apply effects in order: Color Grade → Vignette → Film Grain → Scanlines → Chromatic Aberration → Bloom → Static Burst → Film Scratches
	if preset.ColorGrade.Enabled {
		p.ApplyColorGrade(framebuffer, preset.ColorGrade)
	}
	if preset.Vignette.Enabled {
		p.ApplyVignette(framebuffer, preset.Vignette)
	}
	if preset.FilmGrain.Enabled {
		p.ApplyFilmGrain(framebuffer, preset.FilmGrain)
	}
	if preset.Scanlines.Enabled {
		p.ApplyScanlines(framebuffer, preset.Scanlines)
	}
	if preset.ChromaticAberration.Enabled {
		p.ApplyChromaticAberration(framebuffer, preset.ChromaticAberration)
	}
	if preset.Bloom.Enabled {
		p.ApplyBloom(framebuffer, preset.Bloom)
	}
	if preset.StaticBurst.Enabled {
		p.ApplyStaticBurst(framebuffer, preset.StaticBurst)
	}
	if preset.FilmScratches.Enabled {
		p.ApplyFilmScratches(framebuffer, preset.FilmScratches)
	}
}

// ApplyVignette darkens edges of the screen with configurable tint.
func (p *PostProcessor) ApplyVignette(framebuffer []byte, cfg VignetteConfig) {
	centerX := float64(p.width) / 2.0
	centerY := float64(p.height) / 2.0
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			// Normalized distance (0 at center, 1 at corners)
			normDist := dist / maxDist

			// Apply power curve for vignette falloff
			vignette := 1.0 - math.Pow(normDist, cfg.Power)
			vignette = cfg.Intensity*vignette + (1.0 - cfg.Intensity)

			idx := (y*p.width + x) * 4
			r := float64(framebuffer[idx]) / 255.0
			g := float64(framebuffer[idx+1]) / 255.0
			b := float64(framebuffer[idx+2]) / 255.0

			// Apply vignette multiplier and tint
			r = r*vignette + float64(cfg.Tint.R)/255.0*(1.0-vignette)
			g = g*vignette + float64(cfg.Tint.G)/255.0*(1.0-vignette)
			b = b*vignette + float64(cfg.Tint.B)/255.0*(1.0-vignette)

			framebuffer[idx] = uint8(clamp(r * 255.0))
			framebuffer[idx+1] = uint8(clamp(g * 255.0))
			framebuffer[idx+2] = uint8(clamp(b * 255.0))
		}
	}
}

// ApplyFilmGrain adds noise to the image.
func (p *PostProcessor) ApplyFilmGrain(framebuffer []byte, cfg FilmGrainConfig) {
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			idx := (y*p.width + x) * 4

			// Generate deterministic noise based on position
			noise := (p.rng.Float64()*2.0 - 1.0) * cfg.Intensity

			r := float64(framebuffer[idx])/255.0 + noise
			g := float64(framebuffer[idx+1])/255.0 + noise
			b := float64(framebuffer[idx+2])/255.0 + noise

			framebuffer[idx] = uint8(clamp(r * 255.0))
			framebuffer[idx+1] = uint8(clamp(g * 255.0))
			framebuffer[idx+2] = uint8(clamp(b * 255.0))
		}
	}
}

// ApplyScanlines adds horizontal scanline effect.
func (p *PostProcessor) ApplyScanlines(framebuffer []byte, cfg ScanlinesConfig) {
	for y := 0; y < p.height; y++ {
		// Every nth line is darkened
		if y%cfg.Spacing == 0 {
			for x := 0; x < p.width; x++ {
				idx := (y*p.width + x) * 4

				r := float64(framebuffer[idx]) / 255.0 * (1.0 - cfg.Intensity)
				g := float64(framebuffer[idx+1]) / 255.0 * (1.0 - cfg.Intensity)
				b := float64(framebuffer[idx+2]) / 255.0 * (1.0 - cfg.Intensity)

				framebuffer[idx] = uint8(clamp(r * 255.0))
				framebuffer[idx+1] = uint8(clamp(g * 255.0))
				framebuffer[idx+2] = uint8(clamp(b * 255.0))
			}
		}
	}
}

// ApplyChromaticAberration separates RGB channels at screen edges.
func (p *PostProcessor) ApplyChromaticAberration(framebuffer []byte, cfg ChromaticAberrationConfig) {
	// Create a copy for reading original values
	original := make([]byte, len(framebuffer))
	copy(original, framebuffer)

	centerX := float64(p.width) / 2.0

	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			dx := float64(x) - centerX

			// Offset increases with distance from center
			offsetR := int(dx * cfg.Offset)
			offsetB := -int(dx * cfg.Offset)

			idx := (y*p.width + x) * 4

			// Red channel shifted
			xR := clampInt(x+offsetR, 0, p.width-1)
			idxR := (y*p.width + xR) * 4
			framebuffer[idx] = original[idxR]

			// Green unchanged
			framebuffer[idx+1] = original[idx+1]

			// Blue channel shifted opposite direction
			xB := clampInt(x+offsetB, 0, p.width-1)
			idxB := (y*p.width + xB) * 4
			framebuffer[idx+2] = original[idxB+2]
		}
	}
}

// ApplyBloom brightens and spreads bright areas.
func (p *PostProcessor) ApplyBloom(framebuffer []byte, cfg BloomConfig) {
	// Extract bright pixels (threshold)
	brightPixels := make([]float64, len(framebuffer))
	for i := 0; i < len(framebuffer); i += 4 {
		r := float64(framebuffer[i]) / 255.0
		g := float64(framebuffer[i+1]) / 255.0
		b := float64(framebuffer[i+2]) / 255.0

		// Luminance
		luma := r*0.299 + g*0.587 + b*0.114

		if luma > cfg.Threshold {
			excess := luma - cfg.Threshold
			brightPixels[i] = r * excess
			brightPixels[i+1] = g * excess
			brightPixels[i+2] = b * excess
		}
	}

	// Simple box blur on bright pixels
	blurred := make([]float64, len(brightPixels))
	radius := cfg.Radius
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			idx := (y*p.width + x) * 4

			var sumR, sumG, sumB float64
			count := 0

			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx := clampInt(x+dx, 0, p.width-1)
					ny := clampInt(y+dy, 0, p.height-1)
					nidx := (ny*p.width + nx) * 4

					sumR += brightPixels[nidx]
					sumG += brightPixels[nidx+1]
					sumB += brightPixels[nidx+2]
					count++
				}
			}

			if count > 0 {
				blurred[idx] = sumR / float64(count)
				blurred[idx+1] = sumG / float64(count)
				blurred[idx+2] = sumB / float64(count)
			}
		}
	}

	// Add bloom back to framebuffer
	for i := 0; i < len(framebuffer); i += 4 {
		r := float64(framebuffer[i])/255.0 + blurred[i]*cfg.Intensity
		g := float64(framebuffer[i+1])/255.0 + blurred[i+1]*cfg.Intensity
		b := float64(framebuffer[i+2])/255.0 + blurred[i+2]*cfg.Intensity

		framebuffer[i] = uint8(clamp(r * 255.0))
		framebuffer[i+1] = uint8(clamp(g * 255.0))
		framebuffer[i+2] = uint8(clamp(b * 255.0))
	}
}

// ApplyColorGrade adjusts color balance, saturation, and contrast.
func (p *PostProcessor) ApplyColorGrade(framebuffer []byte, cfg ColorGradeConfig) {
	for y := 0; y < p.height; y++ {
		for x := 0; x < p.width; x++ {
			idx := (y*p.width + x) * 4

			r := float64(framebuffer[idx]) / 255.0
			g := float64(framebuffer[idx+1]) / 255.0
			b := float64(framebuffer[idx+2]) / 255.0

			// Apply contrast (centered around 0.5)
			r = (r-0.5)*cfg.Contrast + 0.5
			g = (g-0.5)*cfg.Contrast + 0.5
			b = (b-0.5)*cfg.Contrast + 0.5

			// Apply saturation
			luma := r*0.299 + g*0.587 + b*0.114
			r = luma + (r-luma)*cfg.Saturation
			g = luma + (g-luma)*cfg.Saturation
			b = luma + (b-luma)*cfg.Saturation

			// Apply warmth (shift toward orange/blue)
			if cfg.Warmth > 0 {
				r += cfg.Warmth * 0.1
				g += cfg.Warmth * 0.05
			} else {
				b += -cfg.Warmth * 0.1
			}

			framebuffer[idx] = uint8(clamp(r * 255.0))
			framebuffer[idx+1] = uint8(clamp(g * 255.0))
			framebuffer[idx+2] = uint8(clamp(b * 255.0))
		}
	}
}

// GenrePreset defines post-processing configuration for a genre.
type GenrePreset struct {
	Vignette            VignetteConfig
	FilmGrain           FilmGrainConfig
	Scanlines           ScanlinesConfig
	ChromaticAberration ChromaticAberrationConfig
	Bloom               BloomConfig
	ColorGrade          ColorGradeConfig
	StaticBurst         StaticBurstConfig
	FilmScratches       FilmScratchesConfig
}

// VignetteConfig controls vignette effect.
type VignetteConfig struct {
	Enabled   bool
	Intensity float64    // 0.0-1.0
	Power     float64    // Falloff curve exponent
	Tint      color.RGBA // Edge color
}

// FilmGrainConfig controls film grain effect.
type FilmGrainConfig struct {
	Enabled   bool
	Intensity float64 // 0.0-1.0
}

// ScanlinesConfig controls scanline effect.
type ScanlinesConfig struct {
	Enabled   bool
	Intensity float64 // 0.0-1.0
	Spacing   int     // Pixels between scanlines
}

// ChromaticAberrationConfig controls RGB channel separation.
type ChromaticAberrationConfig struct {
	Enabled bool
	Offset  float64 // Pixel offset multiplier
}

// BloomConfig controls bloom effect.
type BloomConfig struct {
	Enabled   bool
	Threshold float64 // Luminance threshold (0.0-1.0)
	Intensity float64 // Bloom strength
	Radius    int     // Blur radius in pixels
}

// ColorGradeConfig controls color grading.
type ColorGradeConfig struct {
	Enabled    bool
	Contrast   float64 // 0.5-2.0 (1.0 = neutral)
	Saturation float64 // 0.0-2.0 (1.0 = neutral)
	Warmth     float64 // -1.0 to 1.0 (0 = neutral, + = warm, - = cool)
}

// StaticBurstConfig controls horror static burst effect.
type StaticBurstConfig struct {
	Enabled     bool
	Probability float64 // Per-frame chance (0.0-1.0)
	Intensity   float64 // Noise strength (0.0-1.0)
	Duration    int     // Frames to persist
}

// FilmScratchesConfig controls film scratch effect.
type FilmScratchesConfig struct {
	Enabled bool
	Density float64 // Scratches per screen width (e.g., 0.05 = 5% of width)
	Length  float64 // Scratch length as fraction of height (0.0-1.0)
}

// ApplyStaticBurst overlays noise on the screen for horror genre.
func (p *PostProcessor) ApplyStaticBurst(framebuffer []byte, cfg StaticBurstConfig) {
	// Check if we should trigger a new burst
	if p.staticBurstTimer <= 0 {
		// Random chance to trigger burst
		if p.rng.Float64() < cfg.Probability {
			p.staticBurstTimer = cfg.Duration
		}
	}

	// Apply static if burst is active
	if p.staticBurstTimer > 0 {
		p.staticBurstTimer--

		for y := 0; y < p.height; y++ {
			for x := 0; x < p.width; x++ {
				idx := (y*p.width + x) * 4

				// Generate strong random noise
				noise := p.rng.Float64() * cfg.Intensity

				r := float64(framebuffer[idx])/255.0*(1.0-noise) + noise
				g := float64(framebuffer[idx+1])/255.0*(1.0-noise) + noise
				b := float64(framebuffer[idx+2])/255.0*(1.0-noise) + noise

				framebuffer[idx] = uint8(clamp(r * 255.0))
				framebuffer[idx+1] = uint8(clamp(g * 255.0))
				framebuffer[idx+2] = uint8(clamp(b * 255.0))
			}
		}
	}
}

// ApplyFilmScratches draws vertical scratch lines for postapoc genre.
func (p *PostProcessor) ApplyFilmScratches(framebuffer []byte, cfg FilmScratchesConfig) {
	scratchCount := int(float64(p.width) * cfg.Density)
	scratchLength := int(float64(p.height) * cfg.Length)

	for i := 0; i < scratchCount; i++ {
		// Random X position
		x := p.rng.Intn(p.width)

		// Random start Y position
		startY := p.rng.Intn(p.height - scratchLength)
		if startY < 0 {
			startY = 0
		}

		// Random scratch brightness (dim scratches)
		brightness := uint8(100 + p.rng.Intn(100))

		// Draw vertical line
		for y := startY; y < startY+scratchLength && y < p.height; y++ {
			idx := (y*p.width + x) * 4

			// Brighten pixel (additive)
			r := uint8(clampInt(int(framebuffer[idx])+int(brightness), 0, 255))
			g := uint8(clampInt(int(framebuffer[idx+1])+int(brightness), 0, 255))
			b := uint8(clampInt(int(framebuffer[idx+2])+int(brightness), 0, 255))

			framebuffer[idx] = r
			framebuffer[idx+1] = g
			framebuffer[idx+2] = b
		}
	}
}

// GetGenrePreset returns the post-processing preset for a genre.
func GetGenrePreset(genreID string) GenrePreset {
	switch genreID {
	case "fantasy":
		// Warm sepia with film grain
		return GenrePreset{
			ColorGrade: ColorGradeConfig{
				Enabled:    true,
				Contrast:   1.1,
				Saturation: 0.9,
				Warmth:     0.3, // Warm sepia
			},
			Vignette: VignetteConfig{
				Enabled:   true,
				Intensity: 0.4,
				Power:     2.0,
				Tint:      color.RGBA{R: 40, G: 30, B: 20, A: 255},
			},
			FilmGrain: FilmGrainConfig{
				Enabled:   true,
				Intensity: 0.08,
			},
			Scanlines: ScanlinesConfig{
				Enabled: false,
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled: false,
			},
			Bloom: BloomConfig{
				Enabled: false,
			},
			StaticBurst: StaticBurstConfig{
				Enabled: false,
			},
			FilmScratches: FilmScratchesConfig{
				Enabled: false,
			},
		}

	case "scifi":
		// Cold blue with scanlines and chromatic aberration
		return GenrePreset{
			ColorGrade: ColorGradeConfig{
				Enabled:    true,
				Contrast:   1.2,
				Saturation: 1.0,
				Warmth:     -0.3, // Cool blue
			},
			Vignette: VignetteConfig{
				Enabled:   true,
				Intensity: 0.3,
				Power:     2.5,
				Tint:      color.RGBA{R: 10, G: 20, B: 40, A: 255},
			},
			FilmGrain: FilmGrainConfig{
				Enabled:   true,
				Intensity: 0.05,
			},
			Scanlines: ScanlinesConfig{
				Enabled:   true,
				Intensity: 0.15,
				Spacing:   2,
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled: true,
				Offset:  0.002,
			},
			Bloom: BloomConfig{
				Enabled: false,
			},
			StaticBurst: StaticBurstConfig{
				Enabled: false,
			},
			FilmScratches: FilmScratchesConfig{
				Enabled: false,
			},
		}

	case "horror":
		// Heavy dark with green desaturate
		return GenrePreset{
			ColorGrade: ColorGradeConfig{
				Enabled:    true,
				Contrast:   1.3,
				Saturation: 0.6,
				Warmth:     -0.1, // Slight green tint
			},
			Vignette: VignetteConfig{
				Enabled:   true,
				Intensity: 0.7, // Heavy vignette
				Power:     1.8,
				Tint:      color.RGBA{R: 5, G: 10, B: 5, A: 255},
			},
			FilmGrain: FilmGrainConfig{
				Enabled:   true,
				Intensity: 0.12,
			},
			Scanlines: ScanlinesConfig{
				Enabled: false,
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled: false,
			},
			Bloom: BloomConfig{
				Enabled: false,
			},
			StaticBurst: StaticBurstConfig{
				Enabled:     true,
				Probability: 0.01, // 1% chance per frame
				Intensity:   0.8,  // Strong noise
				Duration:    3,    // 3 frames

			},
			FilmScratches: FilmScratchesConfig{
				Enabled: false,
			},
		}

	case "cyberpunk":
		// Magenta/cyan with neon bloom
		return GenrePreset{
			ColorGrade: ColorGradeConfig{
				Enabled:    true,
				Contrast:   1.15,
				Saturation: 1.4, // High saturation
				Warmth:     0.0,
			},
			Vignette: VignetteConfig{
				Enabled:   true,
				Intensity: 0.35,
				Power:     2.2,
				Tint:      color.RGBA{R: 30, G: 10, B: 40, A: 255},
			},
			FilmGrain: FilmGrainConfig{
				Enabled:   true,
				Intensity: 0.06,
			},
			Scanlines: ScanlinesConfig{
				Enabled: false,
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled: true,
				Offset:  0.003,
			},
			Bloom: BloomConfig{
				Enabled:   true,
				Threshold: 0.7,
				Intensity: 0.5,
				Radius:    3,
			},
			StaticBurst: StaticBurstConfig{
				Enabled: false,
			},
			FilmScratches: FilmScratchesConfig{
				Enabled: false,
			},
		}

	case "postapoc":
		// Washed orange with dust/scratches
		return GenrePreset{
			ColorGrade: ColorGradeConfig{
				Enabled:    true,
				Contrast:   0.95,
				Saturation: 0.8,
				Warmth:     0.4, // Orange/dust tint
			},
			Vignette: VignetteConfig{
				Enabled:   true,
				Intensity: 0.5,
				Power:     2.0,
				Tint:      color.RGBA{R: 50, G: 40, B: 25, A: 255},
			},
			FilmGrain: FilmGrainConfig{
				Enabled:   true,
				Intensity: 0.15, // Heavy grain for dust
			},
			Scanlines: ScanlinesConfig{
				Enabled: false,
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled: false,
			},
			Bloom: BloomConfig{
				Enabled: false,
			},
			StaticBurst: StaticBurstConfig{
				Enabled: false,
			},
			FilmScratches: FilmScratchesConfig{
				Enabled: true,
				Density: 0.02, // 2% of width
				Length:  0.6,  // 60% of height

			},
		}

	default:
		// No effects
		return GenrePreset{
			ColorGrade:          ColorGradeConfig{Enabled: false},
			Vignette:            VignetteConfig{Enabled: false},
			FilmGrain:           FilmGrainConfig{Enabled: false},
			Scanlines:           ScanlinesConfig{Enabled: false},
			ChromaticAberration: ChromaticAberrationConfig{Enabled: false},
			Bloom:               BloomConfig{Enabled: false},
			StaticBurst:         StaticBurstConfig{Enabled: false},
			FilmScratches:       FilmScratchesConfig{Enabled: false},
		}
	}
}

// clamp restricts a value to [0, 255].
func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// clampInt restricts an integer to [min, max].
func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

package heatdistort

import (
	"image"
	"math"
	"sync"

	"github.com/sirupsen/logrus"
)

// GenrePreset defines heat distortion behavior for each genre.
type GenrePreset struct {
	// BaseIntensity scales all heat distortion intensity.
	BaseIntensity float64

	// WaveFrequency controls animation speed (Hz).
	WaveFrequency float64

	// WaveAmplitude controls maximum pixel displacement.
	WaveAmplitude float64

	// VerticalBias biases distortion upward (heat rises).
	VerticalBias float64

	// FalloffExponent controls distance falloff curve.
	FalloffExponent float64

	// JitterAmount adds irregular motion for organic feel.
	JitterAmount float64

	// TintStrength controls how much heat tints distorted pixels.
	TintStrength float64
}

var genrePresets = map[string]GenrePreset{
	"fantasy": {
		BaseIntensity:   0.7,
		WaveFrequency:   3.5,
		WaveAmplitude:   4.0,
		VerticalBias:    0.7,
		FalloffExponent: 2.0,
		JitterAmount:    0.15,
		TintStrength:    0.2,
	},
	"scifi": {
		BaseIntensity:   0.5,
		WaveFrequency:   5.0,
		WaveAmplitude:   2.5,
		VerticalBias:    0.4,
		FalloffExponent: 2.5,
		JitterAmount:    0.05,
		TintStrength:    0.15,
	},
	"horror": {
		BaseIntensity:   0.85,
		WaveFrequency:   2.5,
		WaveAmplitude:   5.0,
		VerticalBias:    0.8,
		FalloffExponent: 1.5,
		JitterAmount:    0.3, // Erratic for unsettling effect
		TintStrength:    0.25,
	},
	"cyberpunk": {
		BaseIntensity:   0.6,
		WaveFrequency:   6.0,
		WaveAmplitude:   3.0,
		VerticalBias:    0.5,
		FalloffExponent: 3.0, // Sharp falloff
		JitterAmount:    0.1,
		TintStrength:    0.3, // Strong neon tinting
	},
	"postapoc": {
		BaseIntensity:   0.9,
		WaveFrequency:   3.0,
		WaveAmplitude:   6.0,
		VerticalBias:    0.75,
		FalloffExponent: 1.8,
		JitterAmount:    0.2,
		TintStrength:    0.35,
	},
}

// System manages heat distortion effects across all heat sources.
type System struct {
	mu              sync.RWMutex
	genre           string
	preset          GenrePreset
	enabled         bool
	width           int
	height          int
	time            float64
	sources         []*sourceData
	pixelsPerUnit   float64
	jitterSeed      float64
	sinCache        []float64
	sinCacheSize    int
	logger          *logrus.Entry
	workBuffer      []byte
	workBufferWidth int
}

// sourceData holds processed heat source information for rendering.
type sourceData struct {
	ScreenX, ScreenY    float64
	RadiusPx            float64
	Intensity           float64
	Phase               float64
	TintR, TintG, TintB float64
}

// NewSystem creates a heat distortion system for the specified genre.
func NewSystem(genreID string, screenWidth, screenHeight int) *System {
	s := &System{
		genre:         genreID,
		enabled:       true,
		width:         screenWidth,
		height:        screenHeight,
		pixelsPerUnit: 32.0,
		sources:       make([]*sourceData, 0, 32),
		sinCacheSize:  360,
		sinCache:      make([]float64, 360),
		logger: logrus.WithFields(logrus.Fields{
			"system": "heatdistort",
		}),
	}

	// Pre-compute sine lookup table
	for i := 0; i < s.sinCacheSize; i++ {
		angle := float64(i) * 2.0 * math.Pi / float64(s.sinCacheSize)
		s.sinCache[i] = math.Sin(angle)
	}

	s.applyGenrePreset(genreID)
	return s
}

// applyGenrePreset configures the system for a specific genre.
func (s *System) applyGenrePreset(genreID string) {
	preset, ok := genrePresets[genreID]
	if !ok {
		s.logger.Warnf("unknown genre %s, using fantasy defaults", genreID)
		preset = genrePresets["fantasy"]
	}
	s.preset = preset
	s.logger.Debugf("applied genre preset: intensity=%.2f, freq=%.2f, amp=%.2f",
		preset.BaseIntensity, preset.WaveFrequency, preset.WaveAmplitude)
}

// SetGenre updates the heat distortion configuration for a new genre.
func (s *System) SetGenre(genreID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.genre = genreID
	s.applyGenrePreset(genreID)
}

// SetEnabled toggles the heat distortion system.
func (s *System) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled
}

// IsEnabled returns whether heat distortion is currently active.
func (s *System) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// SetPixelsPerUnit configures the world-to-screen scaling.
func (s *System) SetPixelsPerUnit(ppu float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pixelsPerUnit = ppu
}

// SetScreenSize updates the screen dimensions.
func (s *System) SetScreenSize(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.width = width
	s.height = height
}

// Update advances the animation time and updates jitter seed.
func (s *System) Update(deltaTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.time += deltaTime
	s.jitterSeed += deltaTime * 17.31 // Prime-ish number for pseudo-random variation
	if s.jitterSeed > 1000 {
		s.jitterSeed -= 1000
	}
}

// ClearSources resets the heat source list for the new frame.
func (s *System) ClearSources() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sources = s.sources[:0]
}

// AddSource registers a heat source for the current frame.
func (s *System) AddSource(screenX, screenY, worldRadius, intensity, phase, tintR, tintG, tintB float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled || intensity <= 0 {
		return
	}

	radiusPx := worldRadius * s.pixelsPerUnit
	if radiusPx < 1 {
		return
	}

	s.sources = append(s.sources, &sourceData{
		ScreenX:   screenX,
		ScreenY:   screenY,
		RadiusPx:  radiusPx,
		Intensity: intensity * s.preset.BaseIntensity,
		Phase:     phase,
		TintR:     tintR,
		TintG:     tintG,
		TintB:     tintB,
	})
}

// AddSourceFromComponent adds a heat source from an ECS component.
func (s *System) AddSourceFromComponent(comp *Component) {
	if !comp.IsActive() {
		return
	}
	s.AddSource(
		comp.ScreenX, comp.ScreenY,
		comp.Radius, comp.Intensity, comp.WavePhase,
		comp.TintR, comp.TintG, comp.TintB,
	)
}

// Apply applies heat distortion to a framebuffer.
// The framebuffer must be in RGBA format (4 bytes per pixel).
func (s *System) Apply(framebuffer []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled || len(s.sources) == 0 {
		return
	}

	// Ensure work buffer exists
	requiredSize := len(framebuffer)
	if len(s.workBuffer) < requiredSize {
		s.workBuffer = make([]byte, requiredSize)
	}
	copy(s.workBuffer, framebuffer)
	s.workBufferWidth = s.width

	// Process each pixel that might be affected by heat distortion
	for _, src := range s.sources {
		s.applySourceDistortion(framebuffer, src)
	}
}

// applySourceDistortion applies distortion from a single heat source.
func (s *System) applySourceDistortion(framebuffer []byte, src *sourceData) {
	// Calculate bounding box for this source's influence
	minX := int(math.Max(0, src.ScreenX-src.RadiusPx-s.preset.WaveAmplitude))
	maxX := int(math.Min(float64(s.width-1), src.ScreenX+src.RadiusPx+s.preset.WaveAmplitude))
	minY := int(math.Max(0, src.ScreenY-src.RadiusPx-s.preset.WaveAmplitude))
	maxY := int(math.Min(float64(s.height-1), src.ScreenY+src.RadiusPx+s.preset.WaveAmplitude))

	if minX >= maxX || minY >= maxY {
		return
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			s.distortPixel(framebuffer, x, y, src)
		}
	}
}

// applySourceDistortionToBuffer applies distortion to output buffer reading from work buffer.
func (s *System) applySourceDistortionToBuffer(output []byte, src *sourceData) {
	// Calculate bounding box for this source's influence
	minX := int(math.Max(0, src.ScreenX-src.RadiusPx-s.preset.WaveAmplitude))
	maxX := int(math.Min(float64(s.width-1), src.ScreenX+src.RadiusPx+s.preset.WaveAmplitude))
	minY := int(math.Max(0, src.ScreenY-src.RadiusPx-s.preset.WaveAmplitude))
	maxY := int(math.Min(float64(s.height-1), src.ScreenY+src.RadiusPx+s.preset.WaveAmplitude))

	if minX >= maxX || minY >= maxY {
		return
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			s.distortPixelToBuffer(output, x, y, src)
		}
	}
}

// distortPixel applies heat distortion to a single pixel.
func (s *System) distortPixel(framebuffer []byte, x, y int, src *sourceData) {
	dx := float64(x) - src.ScreenX
	dy := float64(y) - src.ScreenY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > src.RadiusPx {
		return
	}

	// Calculate falloff (1.0 at center, 0.0 at edge)
	normalizedDist := dist / src.RadiusPx
	falloff := math.Pow(1.0-normalizedDist, s.preset.FalloffExponent)

	// Calculate wave displacement
	waveInput := src.Phase + s.time*s.preset.WaveFrequency + float64(y)*0.05 + float64(x)*0.02
	wave := s.fastSin(waveInput)

	// Add secondary wave for more organic motion
	wave2 := s.fastSin(waveInput*1.7 + 0.5)
	combinedWave := wave*0.7 + wave2*0.3

	// Add jitter for irregular motion
	jitter := s.fastSin(s.jitterSeed+float64(x)*0.1+float64(y)*0.13) * s.preset.JitterAmount

	// Calculate displacement
	amplitude := s.preset.WaveAmplitude * falloff * src.Intensity
	dispX := combinedWave * amplitude * (1.0 - s.preset.VerticalBias)
	dispY := (combinedWave + jitter) * amplitude * s.preset.VerticalBias * -1.0 // Negative = upward

	// Sample from displaced position
	srcX := clampInt(x+int(dispX), 0, s.width-1)
	srcY := clampInt(y+int(dispY), 0, s.height-1)

	srcIdx := (srcY*s.width + srcX) * 4
	dstIdx := (y*s.width + x) * 4

	if srcIdx+3 >= len(s.workBuffer) || dstIdx+3 >= len(framebuffer) {
		return
	}

	// Read source pixel
	r := float64(s.workBuffer[srcIdx])
	g := float64(s.workBuffer[srcIdx+1])
	b := float64(s.workBuffer[srcIdx+2])

	// Apply heat tinting based on distance and intensity
	tintStrength := s.preset.TintStrength * falloff * src.Intensity
	r = r*(1.0-tintStrength) + r*src.TintR*tintStrength
	g = g*(1.0-tintStrength) + g*src.TintG*tintStrength
	b = b*(1.0-tintStrength) + b*src.TintB*tintStrength

	// Write distorted pixel
	framebuffer[dstIdx] = uint8(clampFloat(r, 0, 255))
	framebuffer[dstIdx+1] = uint8(clampFloat(g, 0, 255))
	framebuffer[dstIdx+2] = uint8(clampFloat(b, 0, 255))
	// Alpha unchanged
}

// distortPixelToBuffer applies heat distortion reading from workBuffer and writing to output.
func (s *System) distortPixelToBuffer(output []byte, x, y int, src *sourceData) {
	dx := float64(x) - src.ScreenX
	dy := float64(y) - src.ScreenY
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > src.RadiusPx {
		return
	}

	// Calculate falloff (1.0 at center, 0.0 at edge)
	normalizedDist := dist / src.RadiusPx
	falloff := math.Pow(1.0-normalizedDist, s.preset.FalloffExponent)

	// Calculate wave displacement
	waveInput := src.Phase + s.time*s.preset.WaveFrequency + float64(y)*0.05 + float64(x)*0.02
	wave := s.fastSin(waveInput)

	// Add secondary wave for more organic motion
	wave2 := s.fastSin(waveInput*1.7 + 0.5)
	combinedWave := wave*0.7 + wave2*0.3

	// Add jitter for irregular motion
	jitter := s.fastSin(s.jitterSeed+float64(x)*0.1+float64(y)*0.13) * s.preset.JitterAmount

	// Calculate displacement
	amplitude := s.preset.WaveAmplitude * falloff * src.Intensity
	dispX := combinedWave * amplitude * (1.0 - s.preset.VerticalBias)
	dispY := (combinedWave + jitter) * amplitude * s.preset.VerticalBias * -1.0 // Negative = upward

	// Sample from displaced position in work buffer
	srcX := clampInt(x+int(dispX), 0, s.width-1)
	srcY := clampInt(y+int(dispY), 0, s.height-1)

	srcIdx := (srcY*s.width + srcX) * 4
	dstIdx := (y*s.width + x) * 4

	if srcIdx+3 >= len(s.workBuffer) || dstIdx+3 >= len(output) {
		return
	}

	// Read source pixel from work buffer
	r := float64(s.workBuffer[srcIdx])
	g := float64(s.workBuffer[srcIdx+1])
	b := float64(s.workBuffer[srcIdx+2])

	// Apply heat tinting based on distance and intensity
	tintStrength := s.preset.TintStrength * falloff * src.Intensity
	r = r*(1.0-tintStrength) + r*src.TintR*tintStrength
	g = g*(1.0-tintStrength) + g*src.TintG*tintStrength
	b = b*(1.0-tintStrength) + b*src.TintB*tintStrength

	// Write distorted pixel to output
	output[dstIdx] = uint8(clampFloat(r, 0, 255))
	output[dstIdx+1] = uint8(clampFloat(g, 0, 255))
	output[dstIdx+2] = uint8(clampFloat(b, 0, 255))
	// Alpha unchanged
}

// fastSin returns a cached sine value for the given angle.
func (s *System) fastSin(angle float64) float64 {
	// Normalize angle to [0, 2*pi)
	normalizedAngle := math.Mod(angle, 2.0*math.Pi)
	if normalizedAngle < 0 {
		normalizedAngle += 2.0 * math.Pi
	}

	// Map to cache index
	index := int(normalizedAngle / (2.0 * math.Pi) * float64(s.sinCacheSize))
	if index >= s.sinCacheSize {
		index = s.sinCacheSize - 1
	}

	return s.sinCache[index]
}

// GetSourceCount returns the number of active heat sources.
func (s *System) GetSourceCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sources)
}

// GetPreset returns the current genre preset (for testing/inspection).
func (s *System) GetPreset() GenrePreset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.preset
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

// clampFloat restricts a float64 to [min, max].
func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// RenderDistortionPreview generates a debug visualization of the heat distortion.
// Useful for testing and tuning the effect.
func (s *System) RenderDistortionPreview(width, height int) *image.RGBA {
	s.mu.RLock()
	defer s.mu.RUnlock()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with gray
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 4
			img.Pix[idx] = 128
			img.Pix[idx+1] = 128
			img.Pix[idx+2] = 128
			img.Pix[idx+3] = 255
		}
	}

	// Visualize heat source influence
	for _, src := range s.sources {
		minX := int(math.Max(0, src.ScreenX-src.RadiusPx))
		maxX := int(math.Min(float64(width-1), src.ScreenX+src.RadiusPx))
		minY := int(math.Max(0, src.ScreenY-src.RadiusPx))
		maxY := int(math.Min(float64(height-1), src.ScreenY+src.RadiusPx))

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				dx := float64(x) - src.ScreenX
				dy := float64(y) - src.ScreenY
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist <= src.RadiusPx {
					falloff := 1.0 - dist/src.RadiusPx
					intensity := falloff * src.Intensity

					idx := (y*width + x) * 4
					img.Pix[idx] = uint8(math.Min(255, 128+intensity*127))
					img.Pix[idx+1] = uint8(math.Max(0, 128-intensity*50))
					img.Pix[idx+2] = uint8(math.Max(0, 128-intensity*80))
				}
			}
		}
	}

	return img
}

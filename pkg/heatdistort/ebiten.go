package heatdistort

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// ApplyToImage applies heat distortion directly to an ebiten.Image.
// This reads the image pixels, applies distortion, and writes back.
func (s *System) ApplyToImage(screen *ebiten.Image, sources []Component) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled || len(sources) == 0 {
		return
	}

	// Clear and re-add sources
	s.sources = s.sources[:0]
	for i := range sources {
		if !sources[i].IsActive() {
			continue
		}
		s.sources = append(s.sources, &sourceData{
			ScreenX:   sources[i].ScreenX,
			ScreenY:   sources[i].ScreenY,
			RadiusPx:  sources[i].Radius,
			Intensity: sources[i].Intensity * s.preset.BaseIntensity,
			Phase:     sources[i].WavePhase,
			TintR:     sources[i].TintR,
			TintG:     sources[i].TintG,
			TintB:     sources[i].TintB,
		})
	}

	if len(s.sources) == 0 {
		return
	}

	// Get screen dimensions
	bounds := screen.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width != s.width || height != s.height {
		s.width = width
		s.height = height
	}

	// Read pixels from screen
	requiredSize := width * height * 4
	if len(s.workBuffer) < requiredSize {
		s.workBuffer = make([]byte, requiredSize)
	}

	// Read pixels (source buffer for sampling)
	screen.ReadPixels(s.workBuffer)
	s.workBufferWidth = width

	// Create output buffer
	output := make([]byte, requiredSize)
	copy(output, s.workBuffer)

	// Apply distortion for each source
	for _, src := range s.sources {
		s.applySourceDistortionToBuffer(output, src)
	}

	// Write distorted pixels back to screen
	screen.WritePixels(output)
}

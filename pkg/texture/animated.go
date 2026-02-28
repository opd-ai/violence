// Package texture provides procedurally generated texture atlas.
package texture

import (
	"image"
	"image/color"
	"math"

	"github.com/opd-ai/violence/pkg/rng"
)

// AnimatedTexture represents a multi-frame animated texture sequence.
// Frames are procedurally generated and deterministically selected based on game tick.
type AnimatedTexture struct {
	frames []image.Image
	fps    int
}

// NewAnimatedTexture creates an animated texture with the given frame count and FPS.
// Each frame is procedurally generated using a deterministic seed.
func NewAnimatedTexture(frameCount, fps int) *AnimatedTexture {
	return &AnimatedTexture{
		frames: make([]image.Image, frameCount),
		fps:    fps,
	}
}

// GetFrame returns the appropriate frame for the given tick.
// Uses deterministic frame selection: (tick / fps) % frameCount.
func (at *AnimatedTexture) GetFrame(tick int) image.Image {
	if len(at.frames) == 0 {
		return nil
	}
	frameIndex := (tick / at.fps) % len(at.frames)
	return at.frames[frameIndex]
}

// SetFrame assigns a generated image to a specific frame index.
func (at *AnimatedTexture) SetFrame(index int, img image.Image) {
	if index >= 0 && index < len(at.frames) {
		at.frames[index] = img
	}
}

// FrameCount returns the number of frames in the animation.
func (at *AnimatedTexture) FrameCount() int {
	return len(at.frames)
}

// GenerateAnimated creates an animated texture and adds it to the atlas.
// Pattern determines the animation type: "flicker_torch", "blink_panel", "drip_water".
func (a *Atlas) GenerateAnimated(name string, size, frameCount, fps int, pattern string) error {
	anim := NewAnimatedTexture(frameCount, fps)

	for i := 0; i < frameCount; i++ {
		frameSeed := a.seed ^ hashString(name) ^ uint64(i)
		r := rng.NewRNG(frameSeed)
		img := image.NewRGBA(image.Rect(0, 0, size, size))

		switch pattern {
		case "flicker_torch":
			a.generateFlickerTorchFrame(img, r, i, frameCount)
		case "blink_panel":
			a.generateBlinkPanelFrame(img, r, i, frameCount)
		case "drip_water":
			a.generateDripWaterFrame(img, r, i, frameCount)
		case "neon_pulse":
			a.generateNeonPulseFrame(img, r, i, frameCount)
		case "radiation_glow":
			a.generateRadiationGlowFrame(img, r, i, frameCount)
		default:
			a.generateFlickerTorchFrame(img, r, i, frameCount)
		}

		anim.SetFrame(i, img)
	}

	a.mu.Lock()
	a.animated[name] = anim
	a.mu.Unlock()
	return nil
}

// GetAnimatedFrame retrieves a specific frame from an animated texture.
// Returns nil if the texture doesn't exist or isn't animated.
func (a *Atlas) GetAnimatedFrame(name string, tick int) (image.Image, bool) {
	a.mu.RLock()
	anim, ok := a.animated[name]
	a.mu.RUnlock()
	if !ok {
		return nil, false
	}

	return anim.GetFrame(tick), true
}

// generateFlickerTorchFrame creates a fantasy torch flame frame.
// Brightness varies by frame to simulate flickering fire.
func (a *Atlas) generateFlickerTorchFrame(img *image.RGBA, r *rng.RNG, frame, totalFrames int) {
	bounds := img.Bounds()

	// Flickering intensity based on sine wave with noise
	t := float64(frame) / float64(totalFrames)
	flicker := 0.7 + 0.3*math.Sin(t*math.Pi*2.0) + r.Float64()*0.1 - 0.05

	baseR, baseG, baseB := 255, 180, 80 // Warm torch colors

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Radial distance from center for flame shape
			cx, cy := float64(bounds.Max.X)/2, float64(bounds.Max.Y)/2
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx+dy*dy) / (float64(bounds.Max.X) / 2)

			// Flame intensity falls off with distance
			intensity := (1.0 - dist) * flicker
			if intensity < 0 {
				intensity = 0
			}

			// Add noise for organic flame look
			noise := a.perlinNoise(float64(x)/8.0, float64(y)/8.0+t*10, r) * 0.2

			finalIntensity := math.Max(0, math.Min(1.0, intensity+noise))

			c := color.RGBA{
				R: clampUint8(float64(baseR) * finalIntensity),
				G: clampUint8(float64(baseG) * finalIntensity),
				B: clampUint8(float64(baseB) * finalIntensity),
				A: 255,
			}
			img.Set(x, y, c)
		}
	}
}

// generateBlinkPanelFrame creates a sci-fi control panel blinking frame.
// Panel segments turn on/off in sequence.
func (a *Atlas) generateBlinkPanelFrame(img *image.RGBA, r *rng.RNG, frame, totalFrames int) {
	bounds := img.Bounds()

	// Which panel is active this frame
	activePanel := frame % 4

	baseR, baseG, baseB := 80, 180, 255 // Cool sci-fi blue

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Divide into 4 quadrants
			quadX := x / (bounds.Max.X / 2)
			quadY := y / (bounds.Max.Y / 2)
			panelID := quadY*2 + quadX

			intensity := 0.2 // Dim default
			if panelID == activePanel {
				intensity = 1.0 // Bright when active
			}

			// Add subtle noise
			noise := a.perlinNoise(float64(x)/16.0, float64(y)/16.0, r) * 0.1

			finalIntensity := math.Max(0, math.Min(1.0, intensity+noise))

			c := color.RGBA{
				R: clampUint8(float64(baseR) * finalIntensity),
				G: clampUint8(float64(baseG) * finalIntensity),
				B: clampUint8(float64(baseB) * finalIntensity),
				A: 255,
			}
			img.Set(x, y, c)
		}
	}
}

// generateDripWaterFrame creates a horror dripping water effect frame.
// Water droplets appear at random vertical positions that descend over frames.
func (a *Atlas) generateDripWaterFrame(img *image.RGBA, r *rng.RNG, frame, totalFrames int) {
	bounds := img.Bounds()

	// Base dark wall
	baseColor := color.RGBA{R: 60, G: 65, B: 60, A: 255}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := baseColor

			// Add base texture noise
			noise := a.perlinNoise(float64(x)/12.0, float64(y)/12.0, r) * 0.3
			c = a.applyNoise(c, noise)

			// Drip positions are deterministic based on x coordinate
			if x%8 == 0 {
				// Calculate drip position for this frame
				dripCycle := float64(frame%totalFrames) / float64(totalFrames)
				dripY := int(dripCycle * float64(bounds.Max.Y))

				// Add highlight for water droplet
				if y == dripY || y == dripY+1 {
					// Bright water droplet
					c = color.RGBA{R: 120, G: 140, B: 160, A: 255}
				} else if y > dripY-4 && y < dripY {
					// Wet trail above droplet
					wetness := 1.0 - float64(dripY-y)/4.0
					c.R = clampUint8(float64(c.R) + wetness*30)
					c.G = clampUint8(float64(c.G) + wetness*40)
					c.B = clampUint8(float64(c.B) + wetness*50)
				}
			}

			img.Set(x, y, c)
		}
	}
}

// generateNeonPulseFrame creates a cyberpunk neon pulsing effect.
// Neon colors pulse with sine wave intensity.
func (a *Atlas) generateNeonPulseFrame(img *image.RGBA, r *rng.RNG, frame, totalFrames int) {
	bounds := img.Bounds()

	// Pulsing intensity based on sine wave
	t := float64(frame) / float64(totalFrames)
	pulse := 0.5 + 0.5*math.Sin(t*math.Pi*2.0)

	baseR, baseG, baseB := 255, 0, 255 // Magenta neon

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Horizontal neon bars
			barHeight := bounds.Max.Y / 4
			barIndex := y / barHeight

			intensity := pulse
			if barIndex%2 == 1 {
				// Alternate bars pulse opposite phase
				intensity = 1.0 - pulse
			}

			// Add scan line effect
			if y%2 == 0 {
				intensity *= 0.8
			}

			// Add subtle noise
			noise := a.perlinNoise(float64(x)/20.0, float64(y)/20.0, r) * 0.1

			finalIntensity := math.Max(0.2, math.Min(1.0, intensity+noise))

			c := color.RGBA{
				R: clampUint8(float64(baseR) * finalIntensity),
				G: clampUint8(float64(baseG) * finalIntensity),
				B: clampUint8(float64(baseB) * finalIntensity),
				A: 255,
			}
			img.Set(x, y, c)
		}
	}
}

// generateRadiationGlowFrame creates a post-apocalyptic radiation glow effect.
// Green/yellow glow pulses and shimmers.
func (a *Atlas) generateRadiationGlowFrame(img *image.RGBA, r *rng.RNG, frame, totalFrames int) {
	bounds := img.Bounds()

	// Pulsing glow based on frame
	t := float64(frame) / float64(totalFrames)
	glow := 0.6 + 0.4*math.Sin(t*math.Pi*2.0)

	baseR, baseG, baseB := 200, 255, 50 // Radioactive green-yellow

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Radial glow from center
			cx, cy := float64(bounds.Max.X)/2, float64(bounds.Max.Y)/2
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx+dy*dy) / (float64(bounds.Max.X) / 2)

			// Glow is stronger at center
			intensity := (1.0 - dist*0.7) * glow

			// Add organic shimmer
			noise := a.perlinNoise(float64(x)/10.0+t*5, float64(y)/10.0, r) * 0.3

			finalIntensity := math.Max(0.1, math.Min(1.0, intensity+noise))

			c := color.RGBA{
				R: clampUint8(float64(baseR) * finalIntensity),
				G: clampUint8(float64(baseG) * finalIntensity),
				B: clampUint8(float64(baseB) * finalIntensity),
				A: 255,
			}
			img.Set(x, y, c)
		}
	}
}

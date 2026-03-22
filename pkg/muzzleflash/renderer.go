package muzzleflash

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// Renderer handles drawing muzzle flash effects to screen.
type Renderer struct {
	system  *System
	genreID string
	logger  *logrus.Entry
}

// NewRenderer creates a muzzle flash renderer.
func NewRenderer(system *System, genreID string) *Renderer {
	return &Renderer{
		system:  system,
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"renderer": "muzzleflash",
			"genre":    genreID,
		}),
	}
}

// SetGenre updates the genre-specific rendering style.
func (r *Renderer) SetGenre(genreID string) {
	r.genreID = genreID
	r.logger = r.logger.WithField("genre", genreID)
}

// Render draws all active muzzle flashes.
func (r *Renderer) Render(screen *ebiten.Image, w *engine.World, cameraX, cameraY float64, screenWidth, screenHeight int) {
	flashes := r.system.GetAllActiveFlashes(w)
	if len(flashes) == 0 {
		return
	}

	for _, flash := range flashes {
		renderData := r.system.PrepareRenderData(flash, cameraX, cameraY, screenWidth, screenHeight)

		// Skip if off-screen
		margin := float64(50)
		if renderData.ScreenX < -margin || renderData.ScreenX > float64(screenWidth)+margin ||
			renderData.ScreenY < -margin || renderData.ScreenY > float64(screenHeight)+margin {
			continue
		}

		r.renderFlash(screen, renderData)
	}
}

// renderFlash draws a single muzzle flash effect.
func (r *Renderer) renderFlash(screen *ebiten.Image, data *FlashRenderData) {
	if data.Alpha <= 0.01 {
		return
	}

	// Draw layers from back to front for proper blending
	r.renderOuterGlow(screen, data)
	r.renderRays(screen, data)
	r.renderCore(screen, data)
	r.renderHotspot(screen, data)
}

// renderOuterGlow draws the soft outer glow.
func (r *Renderer) renderOuterGlow(screen *ebiten.Image, data *FlashRenderData) {
	// Multiple concentric circles for gradient effect
	layers := 4
	for i := layers; i >= 1; i-- {
		layerProgress := float64(i) / float64(layers)
		radius := data.OuterRadius * (0.5 + layerProgress*0.5)
		alpha := data.Alpha * (1.0 - layerProgress*0.7) * 0.6

		glowColor := color.RGBA{
			R: data.SecondaryColor.R,
			G: data.SecondaryColor.G,
			B: data.SecondaryColor.B,
			A: uint8(alpha * 255),
		}

		if glowColor.A > 2 {
			vector.DrawFilledCircle(
				screen,
				float32(data.ScreenX),
				float32(data.ScreenY),
				float32(radius),
				glowColor,
				false,
			)
		}
	}
}

// renderRays draws the spike/ray effects emanating from the flash.
func (r *Renderer) renderRays(screen *ebiten.Image, data *FlashRenderData) {
	if len(data.RayAngles) == 0 {
		return
	}

	rayAlpha := data.Alpha * 0.9
	if rayAlpha < 0.05 {
		return
	}

	// Ray thickness varies by flash type
	rayThickness := float32(2.0 * data.Scale)
	if data.Profile.RayCount > 6 {
		rayThickness = float32(1.5 * data.Scale)
	}

	// Rays have gradient from core to tip
	for _, angle := range data.RayAngles {
		// Ray length varies slightly for organic look
		rayLen := data.RayLength * (0.8 + 0.4*math.Abs(math.Sin(angle*2)))

		startX := data.ScreenX + math.Cos(angle)*data.CoreRadius
		startY := data.ScreenY + math.Sin(angle)*data.CoreRadius
		endX := data.ScreenX + math.Cos(angle)*rayLen
		endY := data.ScreenY + math.Sin(angle)*rayLen

		// Draw ray with primary color
		rayColor := color.RGBA{
			R: data.PrimaryColor.R,
			G: data.PrimaryColor.G,
			B: data.PrimaryColor.B,
			A: uint8(rayAlpha * 255),
		}

		vector.StrokeLine(
			screen,
			float32(startX), float32(startY),
			float32(endX), float32(endY),
			rayThickness,
			rayColor,
			false,
		)

		// Add highlight at ray tip
		tipAlpha := rayAlpha * 0.5
		tipColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(tipAlpha * 255)}
		if tipColor.A > 5 {
			vector.DrawFilledCircle(
				screen,
				float32(endX),
				float32(endY),
				float32(rayThickness*0.8),
				tipColor,
				false,
			)
		}
	}
}

// renderCore draws the bright central flash.
func (r *Renderer) renderCore(screen *ebiten.Image, data *FlashRenderData) {
	// Bright core with primary color
	coreAlpha := data.Alpha * data.Profile.CoreBrightness
	if coreAlpha > 1.0 {
		coreAlpha = 1.0
	}

	// Outer core edge (primary color)
	coreColor := color.RGBA{
		R: data.PrimaryColor.R,
		G: data.PrimaryColor.G,
		B: data.PrimaryColor.B,
		A: uint8(coreAlpha * 255),
	}

	if coreColor.A > 2 {
		vector.DrawFilledCircle(
			screen,
			float32(data.ScreenX),
			float32(data.ScreenY),
			float32(data.CoreRadius),
			coreColor,
			false,
		)
	}
}

// renderHotspot draws the bright white center.
func (r *Renderer) renderHotspot(screen *ebiten.Image, data *FlashRenderData) {
	// Pure white center for maximum brightness
	hotspotAlpha := data.Alpha * data.Profile.CoreBrightness
	if hotspotAlpha > 1.0 {
		hotspotAlpha = 1.0
	}

	hotspotRadius := data.CoreRadius * 0.5
	hotspotColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(hotspotAlpha * 255)}

	if hotspotColor.A > 5 {
		vector.DrawFilledCircle(
			screen,
			float32(data.ScreenX),
			float32(data.ScreenY),
			float32(hotspotRadius),
			hotspotColor,
			false,
		)
	}

	// Inner core ring for definition
	if hotspotAlpha > 0.3 {
		ringColor := color.RGBA{R: 255, G: 255, B: 255, A: uint8(hotspotAlpha * 180)}
		vector.StrokeCircle(
			screen,
			float32(data.ScreenX),
			float32(data.ScreenY),
			float32(hotspotRadius*1.3),
			1.0,
			ringColor,
			false,
		)
	}
}

// RenderSingle draws a single flash at a specific position (for preview/testing).
func (r *Renderer) RenderSingle(screen *ebiten.Image, screenX, screenY float64, flashType string, progress, intensity float64) {
	profile := GetProfile(flashType)

	alpha := 1.0 - progress*progress
	sizeProgress := 1.0
	if progress < 0.3 {
		sizeProgress = progress / 0.3
	} else {
		sizeProgress = 1.0 - (progress-0.3)/0.7
	}
	sizeProgress = math.Max(0.3, sizeProgress)

	baseSize := profile.BaseSize * intensity * sizeProgress
	coreRadius := baseSize * 0.3
	outerRadius := baseSize * 0.8
	rayLength := baseSize * 1.5

	// Generate ray angles
	var rayAngles []float64
	if profile.RayCount > 0 {
		rayAngles = make([]float64, profile.RayCount)
		for i := 0; i < profile.RayCount; i++ {
			rayAngles[i] = float64(i) * 2 * math.Pi / float64(profile.RayCount)
		}
	}

	data := &FlashRenderData{
		ScreenX:        screenX,
		ScreenY:        screenY,
		Scale:          1.0,
		Progress:       progress,
		Alpha:          alpha,
		CoreRadius:     coreRadius,
		OuterRadius:    outerRadius,
		RayLength:      rayLength,
		RayAngles:      rayAngles,
		PrimaryColor:   profile.PrimaryColor,
		SecondaryColor: profile.SecondaryColor,
		Profile:        profile,
	}

	r.renderFlash(screen, data)
}

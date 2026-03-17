package impactburst

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Renderer handles drawing impact burst effects to the screen.
type Renderer struct {
	genreID string
}

// NewRenderer creates a new impact burst renderer.
func NewRenderer(genreID string) *Renderer {
	return &Renderer{genreID: genreID}
}

// SetGenre updates the genre-specific rendering style.
func (r *Renderer) SetGenre(genreID string) {
	r.genreID = genreID
}

// Render draws all impacts in the provided list to the screen.
func (r *Renderer) Render(screen *ebiten.Image, impacts []Impact, cameraX, cameraY float64, screenWidth, screenHeight int) {
	for i := range impacts {
		r.renderImpact(screen, &impacts[i], cameraX, cameraY, screenWidth, screenHeight)
	}
}

// renderImpact draws a single impact effect with all its visual components.
func (r *Renderer) renderImpact(screen *ebiten.Image, imp *Impact, cameraX, cameraY float64, screenWidth, screenHeight int) {
	// Transform world position to screen space
	relX := imp.X - cameraX
	relY := imp.Y - cameraY

	// Simple orthographic projection (scaled by 10 for typical game scale)
	screenX := float64(screenWidth/2) + relX*10.0
	screenY := float64(screenHeight/2) + relY*10.0

	// Early exit if off-screen
	margin := 100.0
	if screenX < -margin || screenX > float64(screenWidth)+margin ||
		screenY < -margin || screenY > float64(screenHeight)+margin {
		return
	}

	// Render layers from back to front
	r.renderGlow(screen, imp, screenX, screenY)
	r.renderShockwave(screen, imp, screenX, screenY)
	r.renderDebris(screen, imp, screenX, screenY)
	r.renderFlash(screen, imp, screenX, screenY)
}

// renderGlow draws the soft bloom glow effect behind the impact.
func (r *Renderer) renderGlow(screen *ebiten.Image, imp *Impact, screenX, screenY float64) {
	if imp.GlowIntensity <= 0 {
		return
	}

	profile := getProfileForImpact(imp, r.genreID)
	if !profile.HasGlow {
		return
	}

	// Multiple concentric circles for soft glow effect
	layers := 5
	baseRadius := profile.ShockwaveMaxRadius * imp.Intensity * 0.8

	for i := 0; i < layers; i++ {
		layerProgress := float64(i) / float64(layers-1)
		radius := baseRadius * (0.2 + layerProgress*0.8)

		// Alpha decreases toward edges (inverse square for realistic falloff)
		alphaFactor := (1.0 - layerProgress*layerProgress) * imp.GlowIntensity

		glowColor := profile.GlowColor
		glowColor.A = uint8(float64(glowColor.A) * alphaFactor * 0.5)

		if glowColor.A > 0 {
			vector.DrawFilledCircle(screen,
				float32(screenX), float32(screenY),
				float32(radius),
				glowColor,
				false)
		}
	}
}

// renderShockwave draws expanding shockwave rings with inverse-square falloff.
func (r *Renderer) renderShockwave(screen *ebiten.Image, imp *Impact, screenX, screenY float64) {
	if imp.ShockwaveAlpha <= 0 || imp.ShockwaveRadius <= 0 {
		return
	}

	profile := getProfileForImpact(imp, r.genreID)
	if !profile.HasShockwave {
		return
	}

	// Draw multiple concentric rings for richer visual effect
	for ring := 0; ring < profile.ShockwaveRings; ring++ {
		ringProgress := float64(ring) / float64(profile.ShockwaveRings)
		radius := imp.ShockwaveRadius * (0.5 + ringProgress*0.5)

		// Each ring slightly different alpha for depth
		alphaFactor := imp.ShockwaveAlpha * (1.0 - ringProgress*0.3)

		shockColor := profile.ShockwaveColor
		shockColor.A = uint8(float64(shockColor.A) * alphaFactor)

		if shockColor.A > 2 {
			// Main ring
			width := profile.ShockwaveWidth * (1.0 - ringProgress*0.5)
			vector.StrokeCircle(screen,
				float32(screenX), float32(screenY),
				float32(radius),
				float32(width),
				shockColor,
				false)

			// Inner glow for the ring (brighter core)
			innerColor := brightenColor(shockColor, 1.3)
			innerColor.A = uint8(float64(innerColor.A) * 0.6)
			if innerColor.A > 2 {
				vector.StrokeCircle(screen,
					float32(screenX), float32(screenY),
					float32(radius-width*0.3),
					float32(width*0.5),
					innerColor,
					false)
			}
		}
	}
}

// renderDebris draws all debris particles as shaded shapes.
func (r *Renderer) renderDebris(screen *ebiten.Image, imp *Impact, screenX, screenY float64) {
	for i := range imp.Debris {
		debris := &imp.Debris[i]
		if debris.Age >= debris.MaxAge {
			continue
		}

		// Calculate fade based on lifetime
		progress := debris.Age / debris.MaxAge
		alphaFactor := 1.0 - progress*progress // Ease-out fade

		// Calculate screen position
		debrisScreenX := screenX + debris.X
		debrisScreenY := screenY + debris.Y

		debrisColor := debris.Color
		debrisColor.A = uint8(float64(debrisColor.A) * alphaFactor)

		if debrisColor.A < 2 {
			continue
		}

		if debris.IsChunk {
			// Render chunk as a rotated polygon with shading
			r.renderChunk(screen, debrisScreenX, debrisScreenY, debris, debrisColor)
		} else {
			// Render particle as a shaded circle with highlight
			r.renderParticle(screen, debrisScreenX, debrisScreenY, debris, debrisColor)
		}
	}
}

// renderChunk draws a larger debris chunk with rotation and 3D-like shading.
func (r *Renderer) renderChunk(screen *ebiten.Image, x, y float64, debris *DebrisParticle, col color.RGBA) {
	size := debris.Size
	rot := debris.Rotation

	// Create a simple rotated square/diamond shape
	// Four corners of a square, rotated
	corners := make([]struct{ x, y float32 }, 4)
	for i := 0; i < 4; i++ {
		angle := rot + float64(i)*math.Pi/2 + math.Pi/4
		corners[i].x = float32(x + math.Cos(angle)*size)
		corners[i].y = float32(y + math.Sin(angle)*size)
	}

	// Draw the chunk with three tonal values for 3D effect
	// Shadow side (bottom-right quadrant)
	shadowColor := darkenColor(col, 0.6)
	// Midtone (main body)
	// Highlight side (top-left quadrant)
	highlightColor := brightenColor(col, 1.3)

	// Draw as filled shape by drawing triangles
	// Main body color
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(size*0.8), col, false)

	// Shadow edge (offset down-right)
	shadowColor.A = uint8(float64(shadowColor.A) * 0.7)
	vector.DrawFilledCircle(screen, float32(x+size*0.2), float32(y+size*0.2), float32(size*0.5), shadowColor, false)

	// Highlight edge (offset up-left)
	highlightColor.A = uint8(float64(highlightColor.A) * 0.5)
	vector.DrawFilledCircle(screen, float32(x-size*0.15), float32(y-size*0.15), float32(size*0.3), highlightColor, false)
}

// renderParticle draws a small debris particle with highlight for depth.
func (r *Renderer) renderParticle(screen *ebiten.Image, x, y float64, debris *DebrisParticle, col color.RGBA) {
	size := debris.Size

	// Main particle body
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(size), col, false)

	// Highlight spot (offset toward light source - assume top-left)
	highlightColor := brightenColor(col, 1.5)
	highlightColor.A = uint8(float64(highlightColor.A) * 0.6)
	highlightOffset := size * 0.3
	vector.DrawFilledCircle(screen,
		float32(x-highlightOffset), float32(y-highlightOffset),
		float32(size*0.4),
		highlightColor,
		false)
}

// renderFlash draws the initial bright flash at impact point.
func (r *Renderer) renderFlash(screen *ebiten.Image, imp *Impact, screenX, screenY float64) {
	if imp.FlashAlpha <= 0 {
		return
	}

	profile := getProfileForImpact(imp, r.genreID)

	// Central bright flash
	flashColor := brightenColor(profile.PrimaryColor, 2.0)
	flashColor.A = uint8(255.0 * imp.FlashAlpha)

	flashRadius := profile.ShockwaveMaxRadius * imp.Intensity * 0.3 * (1.0 + imp.FlashAlpha)

	if flashColor.A > 2 {
		vector.DrawFilledCircle(screen,
			float32(screenX), float32(screenY),
			float32(flashRadius),
			flashColor,
			false)

		// White hot center
		whiteFlash := color.RGBA{R: 255, G: 255, B: 255, A: uint8(200.0 * imp.FlashAlpha)}
		vector.DrawFilledCircle(screen,
			float32(screenX), float32(screenY),
			float32(flashRadius*0.4),
			whiteFlash,
			false)
	}
}

// RenderToWorld renders an impact at a specific world position with camera transformation.
// This is used for raycaster-style games where world-to-screen mapping is complex.
func (r *Renderer) RenderToWorld(screen *ebiten.Image, imp *Impact, transformX, transformY float64, screenWidth, screenHeight int) {
	if transformY <= 0.1 {
		return // Behind camera
	}

	// Calculate screen position using perspective projection
	screenX := (float64(screenWidth) / 2.0) * (1.0 + transformX/transformY)
	screenY := float64(screenHeight)/2.0 - float64(screenHeight)/(transformY*4.0)

	// Scale effects based on distance
	distanceScale := 1.0 / (transformY * 0.5)
	distanceScale = clamp(distanceScale, 0.1, 2.0)

	r.renderImpactWithScale(screen, imp, screenX, screenY, distanceScale)
}

// renderImpactWithScale draws an impact with a distance-based scale factor.
func (r *Renderer) renderImpactWithScale(screen *ebiten.Image, imp *Impact, screenX, screenY, scale float64) {
	// Render layers from back to front with scaling
	r.renderGlowScaled(screen, imp, screenX, screenY, scale)
	r.renderShockwaveScaled(screen, imp, screenX, screenY, scale)
	r.renderDebrisScaled(screen, imp, screenX, screenY, scale)
	r.renderFlashScaled(screen, imp, screenX, screenY, scale)
}

func (r *Renderer) renderGlowScaled(screen *ebiten.Image, imp *Impact, screenX, screenY, scale float64) {
	if imp.GlowIntensity <= 0 {
		return
	}

	profile := getProfileForImpact(imp, r.genreID)
	if !profile.HasGlow {
		return
	}

	layers := 4
	baseRadius := profile.ShockwaveMaxRadius * imp.Intensity * 0.8 * scale

	for i := 0; i < layers; i++ {
		layerProgress := float64(i) / float64(layers-1)
		radius := baseRadius * (0.2 + layerProgress*0.8)

		alphaFactor := (1.0 - layerProgress*layerProgress) * imp.GlowIntensity

		glowColor := profile.GlowColor
		glowColor.A = uint8(float64(glowColor.A) * alphaFactor * 0.4)

		if glowColor.A > 0 && radius > 0.5 {
			vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(radius), glowColor, false)
		}
	}
}

func (r *Renderer) renderShockwaveScaled(screen *ebiten.Image, imp *Impact, screenX, screenY, scale float64) {
	if imp.ShockwaveAlpha <= 0 || imp.ShockwaveRadius <= 0 {
		return
	}

	profile := getProfileForImpact(imp, r.genreID)
	if !profile.HasShockwave {
		return
	}

	for ring := 0; ring < profile.ShockwaveRings; ring++ {
		ringProgress := float64(ring) / float64(profile.ShockwaveRings)
		radius := imp.ShockwaveRadius * (0.5 + ringProgress*0.5) * scale

		alphaFactor := imp.ShockwaveAlpha * (1.0 - ringProgress*0.3)

		shockColor := profile.ShockwaveColor
		shockColor.A = uint8(float64(shockColor.A) * alphaFactor)

		if shockColor.A > 2 && radius > 1.0 {
			width := profile.ShockwaveWidth * scale * (1.0 - ringProgress*0.5)
			if width < 0.5 {
				width = 0.5
			}
			vector.StrokeCircle(screen, float32(screenX), float32(screenY), float32(radius), float32(width), shockColor, false)
		}
	}
}

func (r *Renderer) renderDebrisScaled(screen *ebiten.Image, imp *Impact, screenX, screenY, scale float64) {
	for i := range imp.Debris {
		debris := &imp.Debris[i]
		if debris.Age >= debris.MaxAge {
			continue
		}

		progress := debris.Age / debris.MaxAge
		alphaFactor := 1.0 - progress*progress

		debrisScreenX := screenX + debris.X*scale*10.0
		debrisScreenY := screenY + debris.Y*scale*10.0

		debrisColor := debris.Color
		debrisColor.A = uint8(float64(debrisColor.A) * alphaFactor)

		if debrisColor.A < 2 {
			continue
		}

		scaledSize := debris.Size * scale
		if scaledSize < 0.5 {
			scaledSize = 0.5
		}

		if debris.IsChunk {
			// Simplified chunk at distance
			vector.DrawFilledCircle(screen, float32(debrisScreenX), float32(debrisScreenY), float32(scaledSize), debrisColor, false)
			highlightColor := brightenColor(debrisColor, 1.3)
			highlightColor.A = uint8(float64(highlightColor.A) * 0.5)
			vector.DrawFilledCircle(screen, float32(debrisScreenX-scaledSize*0.2), float32(debrisScreenY-scaledSize*0.2), float32(scaledSize*0.4), highlightColor, false)
		} else {
			vector.DrawFilledCircle(screen, float32(debrisScreenX), float32(debrisScreenY), float32(scaledSize), debrisColor, false)
		}
	}
}

func (r *Renderer) renderFlashScaled(screen *ebiten.Image, imp *Impact, screenX, screenY, scale float64) {
	if imp.FlashAlpha <= 0 {
		return
	}

	profile := getProfileForImpact(imp, r.genreID)

	flashColor := brightenColor(profile.PrimaryColor, 2.0)
	flashColor.A = uint8(255.0 * imp.FlashAlpha)

	flashRadius := profile.ShockwaveMaxRadius * imp.Intensity * 0.3 * (1.0 + imp.FlashAlpha) * scale

	if flashColor.A > 2 && flashRadius > 0.5 {
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(flashRadius), flashColor, false)

		whiteFlash := color.RGBA{R: 255, G: 255, B: 255, A: uint8(200.0 * imp.FlashAlpha)}
		vector.DrawFilledCircle(screen, float32(screenX), float32(screenY), float32(flashRadius*0.4), whiteFlash, false)
	}
}

// getProfileForImpact retrieves the appropriate profile for an impact.
func getProfileForImpact(imp *Impact, genreID string) ImpactProfile {
	// Create a temporary system to get the profile
	// In production, this would use a cached lookup
	s := &System{genreID: genreID, profiles: make(map[profileKey]ImpactProfile)}
	return s.buildProfile(imp.Type, imp.Material)
}

func darkenColor(c color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(c.R) * factor),
		G: uint8(float64(c.G) * factor),
		B: uint8(float64(c.B) * factor),
		A: c.A,
	}
}

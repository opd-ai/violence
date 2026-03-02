package weaponanim

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Renderer draws weapon swing animations and trails.
type Renderer struct {
	genre string
}

// NewRenderer creates a weapon animation renderer.
func NewRenderer() *Renderer {
	return &Renderer{
		genre: "fantasy",
	}
}

// SetGenre configures genre-specific rendering.
func (r *Renderer) SetGenre(genreID string) {
	r.genre = genreID
}

// DrawSwing renders the weapon swing arc and trail.
func (r *Renderer) DrawSwing(screen *ebiten.Image, anim *WeaponAnimComponent, centerX, centerY, camX, camY float64) {
	if !anim.Active && len(anim.TrailPoints) == 0 {
		return
	}

	screenX := float32(centerX - camX)
	screenY := float32(centerY - camY)

	// Draw motion trail
	r.drawTrail(screen, anim, screenX, screenY)

	// Draw current weapon position arc
	if anim.Active {
		r.drawArc(screen, anim, screenX, screenY)
	}
}

// drawTrail renders the fading motion trail.
func (r *Renderer) drawTrail(screen *ebiten.Image, anim *WeaponAnimComponent, centerX, centerY float32) {
	if len(anim.TrailPoints) < 2 {
		return
	}

	for i := 0; i < len(anim.TrailPoints)-1; i++ {
		p1 := anim.TrailPoints[i]
		p2 := anim.TrailPoints[i+1]

		// Trail points are already in world space, convert to screen space
		x1 := float32(p1.X) - float32(centerX)
		y1 := float32(p1.Y) - float32(centerY)
		x2 := float32(p2.X) - float32(centerX)
		y2 := float32(p2.Y) - float32(centerY)

		// Fade based on age
		alpha := uint8((1.0 - p1.Age) * float64(anim.Color.A))
		width := float32(anim.Width * (1.0 - p1.Age*0.7))

		trailColor := color.RGBA{
			R: anim.Color.R,
			G: anim.Color.G,
			B: anim.Color.B,
			A: alpha,
		}

		// Draw line segment with varying width
		vector.StrokeLine(screen, x1, y1, x2, y2, width, trailColor, false)
	}
}

// drawArc renders the current weapon arc with glow.
func (r *Renderer) drawArc(screen *ebiten.Image, anim *WeaponAnimComponent, centerX, centerY float32) {
	// Draw outer glow
	glowColor := color.RGBA{
		R: anim.Color.R,
		G: anim.Color.G,
		B: anim.Color.B,
		A: anim.Color.A / 3,
	}

	angle := anim.GetCurrentAngle()
	tipX := centerX + float32(math.Cos(angle)*anim.ArcRadius)
	tipY := centerY + float32(math.Sin(angle)*anim.ArcRadius)

	// Glow
	vector.StrokeLine(screen, centerX, centerY, tipX, tipY, float32(anim.Width+4), glowColor, false)

	// Core
	vector.StrokeLine(screen, centerX, centerY, tipX, tipY, float32(anim.Width), anim.Color, false)

	// Weapon tip highlight
	highlightColor := color.RGBA{R: 255, G: 255, B: 255, A: 200}
	vector.DrawFilledCircle(screen, tipX, tipY, float32(anim.Width), highlightColor, false)
}

// GetSwingColor returns the trail color for a weapon type and genre.
func GetSwingColor(weaponType, genre string) color.RGBA {
	switch genre {
	case "scifi":
		return color.RGBA{R: 100, G: 200, B: 255, A: 200}
	case "cyberpunk":
		return color.RGBA{R: 255, G: 0, B: 255, A: 200}
	case "horror":
		return color.RGBA{R: 150, G: 0, B: 0, A: 180}
	case "postapoc":
		return color.RGBA{R: 180, G: 140, B: 80, A: 160}
	default: // fantasy
		if weaponType == "magic" {
			return color.RGBA{R: 150, G: 100, B: 255, A: 220}
		}
		return color.RGBA{R: 200, G: 200, B: 220, A: 200}
	}
}

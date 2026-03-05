// Package particle provides enhanced particle rendering with varied shapes and effects.
package particle

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
)

// ParticleShape defines the visual representation of a particle.
type ParticleShape int

const (
	ShapeCircle ParticleShape = iota
	ShapeSquare
	ShapeDiamond
	ShapeStar
	ShapeLine
	ShapeGlow
	ShapeSpark
	ShapeSmoke
)

// RenderSystem provides enhanced particle rendering with multiple shapes and effects.
type RenderSystem struct {
	glowCache    map[int]*ebiten.Image
	smokeCache   map[int]*ebiten.Image
	sparkCache   map[int]*ebiten.Image
	maxCacheSize int
}

// NewRenderSystem creates an enhanced particle renderer.
func NewRenderSystem() *RenderSystem {
	return &RenderSystem{
		glowCache:    make(map[int]*ebiten.Image),
		smokeCache:   make(map[int]*ebiten.Image),
		sparkCache:   make(map[int]*ebiten.Image),
		maxCacheSize: 20,
	}
}

// DetermineShape selects particle shape based on particle properties.
func (rs *RenderSystem) DetermineShape(p *Particle, genreID string) ParticleShape {
	// Use particle properties to infer type
	speed := math.Sqrt(p.VX*p.VX + p.VY*p.VY)
	hasVertical := math.Abs(p.VZ) > 0.1

	// Fast moving particles with vertical component = sparks
	if speed > 80 && hasVertical {
		return ShapeSpark
	}

	// Slow moving with upward velocity = smoke
	if speed < 20 && p.VZ < -5 {
		return ShapeSmoke
	}

	// Medium speed = glow (explosions, magic)
	if speed > 40 && speed <= 80 {
		return ShapeGlow
	}

	// Very bright = star (muzzle flash)
	if p.R > 200 && p.G > 180 && p.B < 150 {
		return ShapeStar
	}

	// Red particles = blood (diamond shape)
	if p.R > 150 && p.G < 50 && p.B < 50 {
		return ShapeDiamond
	}

	// Fast directional = line (trails)
	if speed > 60 {
		return ShapeLine
	}

	// Default to circle
	return ShapeCircle
}

// RenderParticle draws a single particle with the appropriate shape.
func (rs *RenderSystem) RenderParticle(screen *ebiten.Image, p *Particle, screenX, screenY float32, genreID string) {
	if !p.Active || p.Life <= 0 {
		return
	}

	// Calculate fade alpha based on lifetime
	lifeFrac := float32(p.Life / p.MaxLife)
	alpha := float32(p.A) / 255.0 * lifeFrac

	shape := rs.DetermineShape(p, genreID)
	size := float32(p.Size)

	switch shape {
	case ShapeCircle:
		rs.drawCircle(screen, screenX, screenY, size, p.R, p.G, p.B, alpha)
	case ShapeSquare:
		rs.drawSquare(screen, screenX, screenY, size, p.R, p.G, p.B, alpha)
	case ShapeDiamond:
		rs.drawDiamond(screen, screenX, screenY, size, p.R, p.G, p.B, alpha)
	case ShapeStar:
		rs.drawStar(screen, screenX, screenY, size, p.R, p.G, p.B, alpha)
	case ShapeLine:
		rs.drawLine(screen, screenX, screenY, float32(p.VX), float32(p.VY), size, p.R, p.G, p.B, alpha)
	case ShapeGlow:
		rs.drawGlow(screen, screenX, screenY, size, p.R, p.G, p.B, alpha)
	case ShapeSpark:
		rs.drawSpark(screen, screenX, screenY, float32(p.VX), float32(p.VY), size, p.R, p.G, p.B, alpha)
	case ShapeSmoke:
		rs.drawSmoke(screen, screenX, screenY, size, lifeFrac, p.R, p.G, p.B, alpha)
	}
}

// drawCircle renders a filled circle particle.
func (rs *RenderSystem) drawCircle(screen *ebiten.Image, x, y, radius float32, r, g, b uint8, alpha float32) {
	c := color.RGBA{R: r, G: g, B: b, A: uint8(alpha * 255)}
	vector.DrawFilledCircle(screen, x, y, radius, c, false)
}

// drawSquare renders a filled square particle.
func (rs *RenderSystem) drawSquare(screen *ebiten.Image, x, y, size float32, r, g, b uint8, alpha float32) {
	c := color.RGBA{R: r, G: g, B: b, A: uint8(alpha * 255)}
	vector.DrawFilledRect(screen, x-size/2, y-size/2, size, size, c, false)
}

// drawDiamond renders a diamond-shaped particle (rotated square).
func (rs *RenderSystem) drawDiamond(screen *ebiten.Image, x, y, size float32, r, g, b uint8, alpha float32) {
	c := color.RGBA{R: r, G: g, B: b, A: uint8(alpha * 255)}
	halfSize := size / 2

	// Draw diamond as 4 triangles from center
	points := []float32{
		x, y - halfSize,
		x + halfSize, y,
		x, y + halfSize,
		x - halfSize, y,
	}

	// Draw filled polygon by drawing two triangles
	vector.StrokeLine(screen, points[0], points[1], points[2], points[3], size*0.4, c, false)
	vector.StrokeLine(screen, points[4], points[5], points[6], points[7], size*0.4, c, false)
}

// drawStar renders a star-shaped particle (4-pointed).
func (rs *RenderSystem) drawStar(screen *ebiten.Image, x, y, size float32, r, g, b uint8, alpha float32) {
	c := color.RGBA{R: r, G: g, B: b, A: uint8(alpha * 255)}

	// Draw as cross with bright center
	vector.StrokeLine(screen, x-size, y, x+size, y, size*0.3, c, false)
	vector.StrokeLine(screen, x, y-size, x, y+size, size*0.3, c, false)

	// Bright center point
	vector.DrawFilledCircle(screen, x, y, size*0.4, c, false)
}

// drawLine renders a motion-trail line particle.
func (rs *RenderSystem) drawLine(screen *ebiten.Image, x, y, vx, vy, size float32, r, g, b uint8, alpha float32) {
	c := color.RGBA{R: r, G: g, B: b, A: uint8(alpha * 255)}

	// Trail extends opposite to velocity
	length := float32(math.Min(float64(size)*2, math.Sqrt(float64(vx*vx+vy*vy))*0.15))
	if length < 1 {
		length = size
	}

	angle := float32(math.Atan2(float64(vy), float64(vx)))
	dx := float32(math.Cos(float64(angle))) * length
	dy := float32(math.Sin(float64(angle))) * length

	vector.StrokeLine(screen, x-dx, y-dy, x, y, size*0.5, c, false)
}

// drawGlow renders a particle with radial gradient glow effect.
func (rs *RenderSystem) drawGlow(screen *ebiten.Image, x, y, size float32, r, g, b uint8, alpha float32) {
	// Draw layered circles for glow effect
	for i := 3; i >= 0; i-- {
		radius := size * (1.0 + float32(i)*0.3)
		layerAlpha := alpha / float32(i+1)
		c := color.RGBA{R: r, G: g, B: b, A: uint8(layerAlpha * 255)}
		vector.DrawFilledCircle(screen, x, y, radius, c, false)
	}
}

// drawSpark renders an elongated spark particle with directional streak.
func (rs *RenderSystem) drawSpark(screen *ebiten.Image, x, y, vx, vy, size float32, r, g, b uint8, alpha float32) {
	// Spark is a bright line in direction of motion
	c := color.RGBA{R: r, G: g, B: b, A: uint8(alpha * 255)}
	cDim := color.RGBA{R: r / 2, G: g / 2, B: b / 2, A: uint8(alpha * 128)}

	angle := float32(math.Atan2(float64(vy), float64(vx)))
	length := size * 3

	dx := float32(math.Cos(float64(angle))) * length
	dy := float32(math.Sin(float64(angle))) * length

	// Draw dimmer outer trail
	vector.StrokeLine(screen, x-dx, y-dy, x+dx*0.3, y+dy*0.3, size*0.8, cDim, false)

	// Draw bright core
	vector.StrokeLine(screen, x-dx*0.5, y-dy*0.5, x, y, size*0.5, c, false)

	// Bright head
	vector.DrawFilledCircle(screen, x, y, size*0.6, c, false)
}

// drawSmoke renders a soft, expanding smoke particle.
func (rs *RenderSystem) drawSmoke(screen *ebiten.Image, x, y, size, lifeFrac float32, r, g, b uint8, alpha float32) {
	// Smoke expands and fades over lifetime
	expandedSize := size * (1.0 + (1.0-lifeFrac)*1.5)

	// Draw soft smoke as layered semi-transparent circles
	for i := 2; i >= 0; i-- {
		radius := expandedSize * (1.0 + float32(i)*0.2)
		layerAlpha := alpha / float32(i+2)
		c := color.RGBA{R: r, G: g, B: b, A: uint8(layerAlpha * 200)}
		vector.DrawFilledCircle(screen, x, y, radius, c, false)
	}
}

// getGlowTexture returns a cached radial gradient texture for glow effects.
func (rs *RenderSystem) getGlowTexture(size int) *ebiten.Image {
	if cached, ok := rs.glowCache[size]; ok {
		return cached
	}

	// Limit cache growth
	if len(rs.glowCache) >= rs.maxCacheSize {
		// Clear cache
		rs.glowCache = make(map[int]*ebiten.Image)
	}

	// Generate radial gradient
	img := ebiten.NewImage(size, size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - float64(size)/2
			dy := float64(y) - float64(size)/2
			dist := math.Sqrt(dx*dx + dy*dy)
			radius := float64(size) / 2

			if dist < radius {
				// Smooth falloff
				alpha := uint8((1.0 - dist/radius) * 255)
				img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: alpha})
			}
		}
	}

	rs.glowCache[size] = img
	return img
}

// ClearCache clears all cached textures.
func (rs *RenderSystem) ClearCache() {
	rs.glowCache = make(map[int]*ebiten.Image)
	rs.smokeCache = make(map[int]*ebiten.Image)
	rs.sparkCache = make(map[int]*ebiten.Image)
}

// Component holds particle shape metadata for an entity.
type Component struct {
	PreferredShape ParticleShape
	GenreID        string
}

// Type returns the component type identifier.
func (c *Component) Type() string {
	return "ParticleRenderer"
}

// NewComponent creates a particle renderer component.
func NewComponent(preferredShape ParticleShape, genreID string) *Component {
	return &Component{
		PreferredShape: preferredShape,
		GenreID:        genreID,
	}
}

// RendererSystem wraps the render system for ECS integration.
type RendererSystem struct {
	renderer *RenderSystem
}

// NewRendererSystem creates a new particle rendering system.
func NewRendererSystem() *RendererSystem {
	return &RendererSystem{
		renderer: NewRenderSystem(),
	}
}

// GetRenderer returns the underlying render system.
func (s *RendererSystem) GetRenderer() *RenderSystem {
	return s.renderer
}

// Type returns the system type identifier.
func (s *RendererSystem) Type() string {
	return "ParticleRenderer"
}

// Update processes particle rendering components (no-op, rendering is done separately).
func (s *RendererSystem) Update(w *engine.World) {
	// Rendering is done in draw phase, not update
}

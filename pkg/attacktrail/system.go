package attacktrail

import (
	"image/color"
	"math"
	"math/rand"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System handles attack trail rendering and lifecycle.
type System struct {
	genreID string
	logger  *logrus.Entry
	world   *engine.World
}

// NewSystem creates an attack trail system.
func NewSystem(genreID string) *System {
	return &System{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "attacktrail",
			"genre":  genreID,
		}),
	}
}

// Update processes all trail components and updates trail state.
func (s *System) Update(w *engine.World) {
	s.world = w
	deltaTime := 0.016 // Assume 60 FPS

	// Query all entities with TrailComponent
	entities := w.Query(reflect.TypeOf(&TrailComponent{}))

	for _, entityID := range entities {
		if trailComp := s.getTrailComponent(entityID); trailComp != nil {
			trailComp.Update(deltaTime)
		}
	}
}

// Render draws all active attack trails to the screen.
func (s *System) Render(screen *ebiten.Image, w *engine.World, cameraX, cameraY float64) {
	s.world = w

	// Query all entities with TrailComponent
	entities := w.Query(reflect.TypeOf(&TrailComponent{}))

	for _, entityID := range entities {
		if trailComp := s.getTrailComponent(entityID); trailComp != nil {
			for _, trail := range trailComp.Trails {
				s.renderTrail(screen, trail, cameraX, cameraY)
			}
		}
	}
}

// renderTrail draws a single trail based on its type.
func (s *System) renderTrail(screen *ebiten.Image, trail *Trail, cameraX, cameraY float64) {
	if trail.Intensity <= 0 {
		return
	}

	// Convert world coordinates to screen coordinates
	screenX := float32(trail.StartX - cameraX)
	screenY := float32(trail.StartY - cameraY)

	// Apply intensity to alpha
	alpha := uint8(float64(trail.Color.A) * trail.Intensity)
	trailColor := color.RGBA{
		R: trail.Color.R,
		G: trail.Color.G,
		B: trail.Color.B,
		A: alpha,
	}

	switch trail.Type {
	case TrailSlash:
		s.renderSlashTrail(screen, trail, screenX, screenY, trailColor)
	case TrailThrust:
		s.renderThrustTrail(screen, trail, screenX, screenY, trailColor)
	case TrailCleave:
		s.renderCleaveTrail(screen, trail, screenX, screenY, trailColor)
	case TrailSmash:
		s.renderSmashTrail(screen, trail, screenX, screenY, trailColor)
	case TrailSpin:
		s.renderSpinTrail(screen, trail, screenX, screenY, trailColor)
	case TrailProjectile:
		s.renderProjectileTrail(screen, trail, screenX, screenY, trailColor)
	}
}

// renderSlashTrail draws a curved arc (sword swing).
func (s *System) renderSlashTrail(screen *ebiten.Image, trail *Trail, screenX, screenY float32, c color.RGBA) {
	segments := 16
	halfArc := trail.Arc / 2

	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments-1)
		angle := trail.Angle - halfArc + t*trail.Arc

		// Calculate arc position
		distance := trail.Range * (0.5 + 0.5*t) // Taper from center
		x1 := screenX + float32(math.Cos(angle)*distance)
		y1 := screenY + float32(math.Sin(angle)*distance)

		// Next segment
		if i < segments-1 {
			t2 := float64(i+1) / float64(segments-1)
			angle2 := trail.Angle - halfArc + t2*trail.Arc
			distance2 := trail.Range * (0.5 + 0.5*t2)
			x2 := screenX + float32(math.Cos(angle2)*distance2)
			y2 := screenY + float32(math.Sin(angle2)*distance2)

			// Fade along the arc
			segmentAlpha := uint8(float64(c.A) * (1.0 - t*0.5))
			segmentColor := color.RGBA{R: c.R, G: c.G, B: c.B, A: segmentAlpha}

			vector.StrokeLine(screen, x1, y1, x2, y2, float32(trail.Width), segmentColor, false)
		}
	}
}

// renderThrustTrail draws a linear piercing trail (spear).
func (s *System) renderThrustTrail(screen *ebiten.Image, trail *Trail, screenX, screenY float32, c color.RGBA) {
	endX := screenX + float32(math.Cos(trail.Angle)*trail.Range)
	endY := screenY + float32(math.Sin(trail.Angle)*trail.Range)

	// Main thrust line
	vector.StrokeLine(screen, screenX, screenY, endX, endY, float32(trail.Width), c, false)

	// Add glow effect at tip
	tipGlow := color.RGBA{R: 255, G: 255, B: 255, A: uint8(float64(c.A) * 0.6)}
	tipRadius := float32(trail.Width * 1.5)
	vector.DrawFilledCircle(screen, endX, endY, tipRadius, tipGlow, false)
}

// renderCleaveTrail draws a wide sweeping arc (greatsword).
func (s *System) renderCleaveTrail(screen *ebiten.Image, trail *Trail, screenX, screenY float32, c color.RGBA) {
	segments := 24
	halfArc := trail.Arc / 2

	// Draw multiple layers for thickness
	layers := 3
	for layer := 0; layer < layers; layer++ {
		layerAlpha := uint8(float64(c.A) * (1.0 - float64(layer)*0.3))
		layerColor := color.RGBA{R: c.R, G: c.G, B: c.B, A: layerAlpha}
		layerOffset := float64(layer) * (trail.Width * 0.3)

		for i := 0; i < segments; i++ {
			t := float64(i) / float64(segments-1)
			angle := trail.Angle - halfArc + t*trail.Arc

			distance := trail.Range - layerOffset
			x1 := screenX + float32(math.Cos(angle)*distance)
			y1 := screenY + float32(math.Sin(angle)*distance)

			if i < segments-1 {
				t2 := float64(i+1) / float64(segments-1)
				angle2 := trail.Angle - halfArc + t2*trail.Arc
				x2 := screenX + float32(math.Cos(angle2)*distance)
				y2 := screenY + float32(math.Sin(angle2)*distance)

				vector.StrokeLine(screen, x1, y1, x2, y2, float32(trail.Width*1.5), layerColor, false)
			}
		}
	}
}

// renderSmashTrail draws a radial impact burst (hammer).
func (s *System) renderSmashTrail(screen *ebiten.Image, trail *Trail, screenX, screenY float32, c color.RGBA) {
	// Draw expanding ring
	radius := float32(trail.Range * (trail.Age / trail.MaxAge))
	ringWidth := float32(trail.Width * 2)

	// Outer ring
	outerAlpha := uint8(float64(c.A) * 0.8)
	vector.StrokeCircle(screen, screenX, screenY, radius, ringWidth, color.RGBA{R: c.R, G: c.G, B: c.B, A: outerAlpha}, false)

	// Inner flash
	if trail.Age < trail.MaxAge*0.3 {
		flashAlpha := uint8(float64(c.A) * (1.0 - trail.Age/(trail.MaxAge*0.3)))
		flashColor := color.RGBA{R: 255, G: 255, B: 255, A: flashAlpha}
		vector.DrawFilledCircle(screen, screenX, screenY, radius*0.5, flashColor, false)
	}
}

// renderSpinTrail draws a full-circle rotation (staff).
func (s *System) renderSpinTrail(screen *ebiten.Image, trail *Trail, screenX, screenY float32, c color.RGBA) {
	segments := 32
	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments)
		angle := t * 2 * math.Pi

		x1 := screenX + float32(math.Cos(angle)*trail.Range)
		y1 := screenY + float32(math.Sin(angle)*trail.Range)

		t2 := float64(i+1) / float64(segments)
		angle2 := t2 * 2 * math.Pi
		x2 := screenX + float32(math.Cos(angle2)*trail.Range)
		y2 := screenY + float32(math.Sin(angle2)*trail.Range)

		// Fade based on rotation progress
		segmentAlpha := uint8(float64(c.A) * (1.0 - t*0.7))
		segmentColor := color.RGBA{R: c.R, G: c.G, B: c.B, A: segmentAlpha}

		vector.StrokeLine(screen, x1, y1, x2, y2, float32(trail.Width), segmentColor, false)
	}
}

// renderProjectileTrail draws a streak following a projectile.
func (s *System) renderProjectileTrail(screen *ebiten.Image, trail *Trail, screenX, screenY float32, c color.RGBA) {
	// Draw from start to end with taper
	endX := screenX + float32(math.Cos(trail.Angle)*trail.Range)
	endY := screenY + float32(math.Sin(trail.Angle)*trail.Range)

	// Main streak
	vector.StrokeLine(screen, screenX, screenY, endX, endY, float32(trail.Width), c, false)

	// Fading segments for motion blur
	if len(trail.Segments) > 1 {
		for i := 0; i < len(trail.Segments)-1; i++ {
			seg1 := trail.Segments[i]
			seg2 := trail.Segments[i+1]

			sx1 := float32(seg1.X - float64(screenX-screenX))
			sy1 := float32(seg1.Y - float64(screenY-screenY))
			sx2 := float32(seg2.X - float64(screenX-screenX))
			sy2 := float32(seg2.Y - float64(screenY-screenY))

			segAlpha := uint8(float64(c.A) * seg1.Intensity * 0.5)
			segColor := color.RGBA{R: c.R, G: c.G, B: c.B, A: segAlpha}

			vector.StrokeLine(screen, sx1, sy1, sx2, sy2, float32(trail.Width*0.7), segColor, false)
		}
	}
}

// CreateSlashTrail generates a slash trail for melee weapons.
func CreateSlashTrail(x, y, angle, range_, arc, width float64, weaponColor color.RGBA) *Trail {
	return &Trail{
		Type:      TrailSlash,
		StartX:    x,
		StartY:    y,
		Angle:     angle,
		Arc:       arc,
		Range:     range_,
		Width:     width,
		Color:     weaponColor,
		Intensity: 1.0,
		MaxAge:    0.15,
		FadeStart: 0.05,
	}
}

// CreateThrustTrail generates a thrust trail for stabbing weapons.
func CreateThrustTrail(x, y, angle, range_, width float64, weaponColor color.RGBA) *Trail {
	return &Trail{
		Type:      TrailThrust,
		StartX:    x,
		StartY:    y,
		Angle:     angle,
		Range:     range_,
		Width:     width,
		Color:     weaponColor,
		Intensity: 1.0,
		MaxAge:    0.12,
		FadeStart: 0.04,
	}
}

// CreateCleaveTrail generates a heavy cleave trail for two-handed weapons.
func CreateCleaveTrail(x, y, angle, range_, arc, width float64, weaponColor color.RGBA) *Trail {
	return &Trail{
		Type:      TrailCleave,
		StartX:    x,
		StartY:    y,
		Angle:     angle,
		Arc:       arc,
		Range:     range_,
		Width:     width,
		Color:     weaponColor,
		Intensity: 1.0,
		MaxAge:    0.2,
		FadeStart: 0.08,
	}
}

// CreateSmashTrail generates an impact trail for blunt weapons.
func CreateSmashTrail(x, y, range_, width float64, weaponColor color.RGBA) *Trail {
	return &Trail{
		Type:      TrailSmash,
		StartX:    x,
		StartY:    y,
		Range:     range_,
		Width:     width,
		Color:     weaponColor,
		Intensity: 1.0,
		MaxAge:    0.25,
		FadeStart: 0.1,
	}
}

// CreateSpinTrail generates a spinning trail for rotating weapons.
func CreateSpinTrail(x, y, range_, width float64, weaponColor color.RGBA) *Trail {
	return &Trail{
		Type:      TrailSpin,
		StartX:    x,
		StartY:    y,
		Range:     range_,
		Width:     width,
		Color:     weaponColor,
		Intensity: 1.0,
		MaxAge:    0.3,
		FadeStart: 0.15,
	}
}

// GetWeaponTrailColor returns a genre-appropriate trail color for a weapon type.
func (s *System) GetWeaponTrailColor(weaponName string, rng *rand.Rand) color.RGBA {
	switch s.genreID {
	case "fantasy":
		return color.RGBA{R: 200, G: 220, B: 255, A: 180}
	case "scifi":
		colors := []color.RGBA{
			{R: 100, G: 200, B: 255, A: 200}, // Plasma blue
			{R: 255, G: 100, B: 100, A: 200}, // Laser red
			{R: 100, G: 255, B: 150, A: 200}, // Energy green
		}
		if rng == nil {
			return colors[0]
		}
		return colors[rng.Intn(len(colors))]
	case "horror":
		return color.RGBA{R: 140, G: 20, B: 20, A: 160}
	case "cyberpunk":
		colors := []color.RGBA{
			{R: 255, G: 0, B: 255, A: 200}, // Neon magenta
			{R: 0, G: 255, B: 255, A: 200}, // Neon cyan
			{R: 255, G: 255, B: 0, A: 200}, // Neon yellow
		}
		if rng == nil {
			return colors[0]
		}
		return colors[rng.Intn(len(colors))]
	default:
		return color.RGBA{R: 220, G: 220, B: 240, A: 180}
	}
}

// getTrailComponent retrieves the trail component from an entity.
func (s *System) getTrailComponent(entityID engine.Entity) *TrailComponent {
	if s.world == nil {
		return nil
	}

	comp, found := s.world.GetComponent(entityID, reflect.TypeOf(&TrailComponent{}))
	if !found {
		return nil
	}

	if tc, ok := comp.(*TrailComponent); ok {
		return tc
	}
	return nil
}

package telegraph

import (
	"image/color"
	"math"
	"math/rand"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages attack telegraph updates and rendering.
type System struct {
	genre  string
	logger *logrus.Entry
	rng    *rand.Rand

	// Rendering buffer (lazy initialized on first render)
	telegraphLayer *ebiten.Image
}

// NewSystem creates an attack telegraph system.
func NewSystem(genreID string, seed int64) *System {
	return &System{
		genre: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system": "telegraph",
			"genre":  genreID,
		}),
		rng:            rand.New(rand.NewSource(seed)),
		telegraphLayer: nil, // Lazy init on first render
	}
}

// Update processes all telegraph components each frame.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime

	telegraphType := reflect.TypeOf(&Component{})
	positionType := reflect.TypeOf(&PositionComponent{})

	entities := w.Query(telegraphType)

	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, telegraphType)
		if !ok {
			continue
		}

		telegraph, ok := comp.(*Component)
		if !ok || !telegraph.Active {
			continue
		}

		// Update position from entity
		if posComp, found := w.GetComponent(entity, positionType); found {
			if pos, ok := posComp.(*PositionComponent); ok {
				telegraph.X = pos.X
				telegraph.Y = pos.Y
			}
		}

		// Advance charge progress
		telegraph.ChargeProgress += deltaTime / telegraph.TelegraphTime

		// Pulsing alpha effect
		pulseSpeed := 2.0 + telegraph.ChargeProgress*4.0 // Speed up as charge builds
		telegraph.IndicatorAlpha = 0.3 + 0.7*math.Abs(math.Sin(telegraph.ChargeProgress*math.Pi*pulseSpeed))

		// Increase indicator size as charge builds
		baseRadius := 16.0
		if telegraph.AttackType == "aoe" {
			baseRadius = 32.0
		} else if telegraph.AttackType == "charge" {
			baseRadius = 20.0
		}
		telegraph.IndicatorRadius = baseRadius * (0.5 + 0.5*telegraph.ChargeProgress)

		// Complete telegraph when fully charged
		if telegraph.ChargeProgress >= 1.0 {
			telegraph.Active = false
			telegraph.ChargeProgress = 0.0

			// Signal attack execution (system integration point)
			s.logger.Debugf("telegraph complete for entity %v, type %s", entity, telegraph.AttackType)
		}
	}
}

// Render draws all active telegraph indicators.
func (s *System) Render(screen *ebiten.Image, w *engine.World, cameraX, cameraY float64) {
	// Lazy init rendering buffer
	if s.telegraphLayer == nil {
		s.telegraphLayer = ebiten.NewImage(800, 600)
	}

	s.telegraphLayer.Clear()

	telegraphType := reflect.TypeOf(&Component{})
	entities := w.Query(telegraphType)

	for _, entity := range entities {
		comp, ok := w.GetComponent(entity, telegraphType)
		if !ok {
			continue
		}

		telegraph, ok := comp.(*Component)
		if !ok || !telegraph.Active {
			continue
		}

		// Convert world to screen coordinates
		screenX := float32((telegraph.X - cameraX) * 32.0)
		screenY := float32((telegraph.Y - cameraY) * 32.0)

		// Draw telegraph indicator based on attack type
		switch telegraph.AttackType {
		case "melee":
			s.drawMeleeTelegraph(screenX, screenY, telegraph)
		case "ranged":
			s.drawRangedTelegraph(screenX, screenY, telegraph)
		case "aoe":
			s.drawAoETelegraph(screenX, screenY, telegraph)
		case "charge":
			s.drawChargeTelegraph(screenX, screenY, telegraph)
		default:
			s.drawGenericTelegraph(screenX, screenY, telegraph)
		}
	}

	// Composite telegraph layer onto screen
	opts := &ebiten.DrawImageOptions{}
	opts.Blend = ebiten.BlendLighter // Additive blend for glow effect
	screen.DrawImage(s.telegraphLayer, opts)
}

// drawMeleeTelegraph renders a sweeping arc indicator for melee attacks.
func (s *System) drawMeleeTelegraph(x, y float32, t *Component) {
	radius := float32(t.IndicatorRadius)
	progress := float32(t.ChargeProgress)

	// Sweeping arc
	startAngle := float32(-math.Pi / 4)
	endAngle := startAngle + progress*float32(math.Pi)/2.0

	// Draw arc segments
	segments := 16
	for i := 0; i < segments; i++ {
		angle := float64(startAngle + float32(i)*(endAngle-startAngle)/float32(segments-1))
		nextAngle := float64(startAngle + float32(i+1)*(endAngle-startAngle)/float32(segments-1))

		x1 := x + float32(math.Cos(angle))*radius
		y1 := y + float32(math.Sin(angle))*radius
		x2 := x + float32(math.Cos(nextAngle))*radius
		y2 := y + float32(math.Sin(nextAngle))*radius

		alpha := uint8(t.IndicatorAlpha * 255 * float64(i) / float64(segments))
		c := color.RGBA{t.PrimaryColor.R, t.PrimaryColor.G, t.PrimaryColor.B, alpha}

		vector.StrokeLine(s.telegraphLayer, x, y, x1, y1, 2.0+progress*2.0, c, false)
		vector.StrokeLine(s.telegraphLayer, x1, y1, x2, y2, 2.0+progress*2.0, c, false)
	}
}

// drawRangedTelegraph renders a directional arrow indicator for ranged attacks.
func (s *System) drawRangedTelegraph(x, y float32, t *Component) {
	length := float32(t.IndicatorRadius * 2.0)
	progress := float32(t.ChargeProgress)

	// Arrow points in direction of attack (simplified - points down)
	endX := x
	endY := y + length*progress

	alpha := uint8(t.IndicatorAlpha * 255)
	c := color.RGBA{t.PrimaryColor.R, t.PrimaryColor.G, t.PrimaryColor.B, alpha}

	// Arrow shaft
	vector.StrokeLine(s.telegraphLayer, x, y, endX, endY, 3.0, c, false)

	// Arrowhead
	if progress > 0.5 {
		headSize := float32(6.0 * progress)
		vector.StrokeLine(s.telegraphLayer, endX, endY, endX-headSize, endY-headSize, 2.0, c, false)
		vector.StrokeLine(s.telegraphLayer, endX, endY, endX+headSize, endY-headSize, 2.0, c, false)
	}
}

// drawAoETelegraph renders a growing circle for area-of-effect attacks.
func (s *System) drawAoETelegraph(x, y float32, t *Component) {
	radius := float32(t.IndicatorRadius)
	progress := float32(t.ChargeProgress)

	// Outer ring
	alpha := uint8(t.IndicatorAlpha * 255)
	c := color.RGBA{t.PrimaryColor.R, t.PrimaryColor.G, t.PrimaryColor.B, alpha}
	vector.StrokeCircle(s.telegraphLayer, x, y, radius, 3.0+progress*2.0, c, false)

	// Inner fill (pulsing)
	fillAlpha := uint8(t.IndicatorAlpha * 100 * float64(progress))
	fillColor := color.RGBA{t.SecondaryColor.R, t.SecondaryColor.G, t.SecondaryColor.B, fillAlpha}
	vector.DrawFilledCircle(s.telegraphLayer, x, y, radius*progress, fillColor, false)

	// Cross-hairs
	if progress > 0.7 {
		crossSize := radius * 0.5
		vector.StrokeLine(s.telegraphLayer, x-crossSize, y, x+crossSize, y, 1.5, c, false)
		vector.StrokeLine(s.telegraphLayer, x, y-crossSize, x, y+crossSize, 1.5, c, false)
	}
}

// drawChargeTelegraph renders a directional charge indicator.
func (s *System) drawChargeTelegraph(x, y float32, t *Component) {
	radius := float32(t.IndicatorRadius)
	progress := float32(t.ChargeProgress)

	// Motion lines
	alpha := uint8(t.IndicatorAlpha * 255)
	c := color.RGBA{t.PrimaryColor.R, t.PrimaryColor.G, t.PrimaryColor.B, alpha}

	for i := 0; i < 5; i++ {
		offset := float32(i-2) * 8.0
		lineLength := radius * progress * (1.0 + float32(i)*0.2)

		vector.StrokeLine(s.telegraphLayer,
			x+offset, y-lineLength,
			x+offset, y,
			2.0-float32(i)*0.3,
			c, false)
	}

	// Center glow
	glowAlpha := uint8(t.IndicatorAlpha * 180 * float64(progress))
	glowColor := color.RGBA{t.SecondaryColor.R, t.SecondaryColor.G, t.SecondaryColor.B, glowAlpha}
	vector.DrawFilledCircle(s.telegraphLayer, x, y, radius*0.3, glowColor, false)
}

// drawGenericTelegraph renders a simple pulsing circle.
func (s *System) drawGenericTelegraph(x, y float32, t *Component) {
	radius := float32(t.IndicatorRadius)
	alpha := uint8(t.IndicatorAlpha * 255)
	c := color.RGBA{t.PrimaryColor.R, t.PrimaryColor.G, t.PrimaryColor.B, alpha}

	vector.StrokeCircle(s.telegraphLayer, x, y, radius, 2.0, c, false)
}

// StartTelegraph initializes a telegraph for an entity.
func (s *System) StartTelegraph(w *engine.World, entity engine.Entity, attackType string, duration float64) {
	telegraphType := reflect.TypeOf(&Component{})

	comp, ok := w.GetComponent(entity, telegraphType)
	var telegraph *Component

	if ok {
		telegraph, _ = comp.(*Component)
	} else {
		// Create new component
		telegraph = &Component{}
		w.AddComponent(entity, telegraph)
	}

	// Configure telegraph based on attack type and genre
	telegraph.Active = true
	telegraph.ChargeProgress = 0.0
	telegraph.TelegraphTime = duration
	telegraph.AttackType = attackType

	// Set colors based on genre and attack type
	telegraph.PrimaryColor, telegraph.SecondaryColor = s.getColorScheme(attackType)

	// Initial settings
	telegraph.IndicatorRadius = 16.0
	telegraph.IndicatorAlpha = 0.5
	telegraph.EmitParticles = true
	telegraph.ParticleCount = 3
	telegraph.ParticleSpread = 8.0
}

// getColorScheme returns primary and secondary colors based on attack type and genre.
func (s *System) getColorScheme(attackType string) (color.RGBA, color.RGBA) {
	switch s.genre {
	case "fantasy":
		switch attackType {
		case "melee":
			return color.RGBA{220, 180, 40, 255}, color.RGBA{255, 220, 80, 255}
		case "ranged":
			return color.RGBA{40, 180, 220, 255}, color.RGBA{80, 220, 255, 255}
		case "aoe":
			return color.RGBA{220, 40, 40, 255}, color.RGBA{255, 100, 100, 255}
		case "charge":
			return color.RGBA{180, 40, 220, 255}, color.RGBA{220, 100, 255, 255}
		}
	case "scifi":
		switch attackType {
		case "melee":
			return color.RGBA{40, 220, 220, 255}, color.RGBA{80, 255, 255, 255}
		case "ranged":
			return color.RGBA{220, 40, 120, 255}, color.RGBA{255, 80, 160, 255}
		case "aoe":
			return color.RGBA{220, 120, 40, 255}, color.RGBA{255, 160, 80, 255}
		case "charge":
			return color.RGBA{120, 40, 220, 255}, color.RGBA{160, 80, 255, 255}
		}
	case "horror":
		switch attackType {
		case "melee":
			return color.RGBA{120, 20, 20, 255}, color.RGBA{180, 40, 40, 255}
		case "ranged":
			return color.RGBA{20, 120, 60, 255}, color.RGBA{40, 180, 100, 255}
		case "aoe":
			return color.RGBA{100, 20, 120, 255}, color.RGBA{150, 40, 180, 255}
		case "charge":
			return color.RGBA{120, 100, 20, 255}, color.RGBA{180, 150, 40, 255}
		}
	case "cyberpunk":
		switch attackType {
		case "melee":
			return color.RGBA{255, 0, 120, 255}, color.RGBA{255, 80, 180, 255}
		case "ranged":
			return color.RGBA{0, 255, 220, 255}, color.RGBA{80, 255, 255, 255}
		case "aoe":
			return color.RGBA{220, 0, 255, 255}, color.RGBA{255, 80, 255, 255}
		case "charge":
			return color.RGBA{255, 220, 0, 255}, color.RGBA{255, 255, 80, 255}
		}
	}

	// Default
	return color.RGBA{220, 180, 40, 255}, color.RGBA{255, 220, 80, 255}
}

// PositionComponent for system reference.
type PositionComponent struct {
	X, Y float64
}

// Type implements Component interface.
func (p *PositionComponent) Type() string {
	return "Position"
}

package threat

import (
	"image/color"
	"math"
	"reflect"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/violence/pkg/common"
	"github.com/opd-ai/violence/pkg/engine"
	"github.com/sirupsen/logrus"
)

// System manages threat indicators for information hierarchy.
type System struct {
	genreID string
	style   GenreStyle
	logger  *logrus.Entry

	// Off-screen threat indicators
	offscreenIndicators []OffscreenIndicator
	maxOffscreen        int

	// Screen dimensions for positioning
	screenWidth  int
	screenHeight int

	// Player position for distance calculations
	playerX, playerY float64

	// Global time for animations
	time float64
}

// NewSystem creates a threat indicator system.
func NewSystem(genreID string) *System {
	s := &System{
		genreID: genreID,
		logger: logrus.WithFields(logrus.Fields{
			"system":  "threat",
			"package": "threat",
		}),
		offscreenIndicators: make([]OffscreenIndicator, 0, 8),
		maxOffscreen:        8,
		screenWidth:         320,
		screenHeight:        200,
	}
	s.applyGenreStyle()
	return s
}

// SetGenre updates genre-specific visual style.
func (s *System) SetGenre(genreID string) {
	s.genreID = genreID
	s.applyGenreStyle()
}

// SetScreenSize updates screen dimensions for indicator positioning.
func (s *System) SetScreenSize(width, height int) {
	s.screenWidth = width
	s.screenHeight = height
}

// SetPlayerPosition updates player position for threat calculations.
func (s *System) SetPlayerPosition(x, y float64) {
	s.playerX = x
	s.playerY = y
}

// applyGenreStyle sets visual parameters based on genre.
func (s *System) applyGenreStyle() {
	switch s.genreID {
	case "cyberpunk":
		s.style = GenreStyle{
			PrimaryColor:    color.RGBA{R: 255, G: 0, B: 180, A: 255},
			SecondaryColor:  color.RGBA{R: 0, G: 255, B: 255, A: 200},
			PulseSpeed:      8.0,
			BorderThickness: 2.0,
			GlowRadius:      6.0,
			ArrowSize:       12.0,
			EdgePadding:     8.0,
		}
	case "horror":
		s.style = GenreStyle{
			PrimaryColor:    color.RGBA{R: 180, G: 0, B: 0, A: 255},
			SecondaryColor:  color.RGBA{R: 100, G: 0, B: 0, A: 180},
			PulseSpeed:      3.0,
			BorderThickness: 2.5,
			GlowRadius:      8.0,
			ArrowSize:       14.0,
			EdgePadding:     6.0,
		}
	case "scifi":
		s.style = GenreStyle{
			PrimaryColor:    color.RGBA{R: 0, G: 200, B: 255, A: 255},
			SecondaryColor:  color.RGBA{R: 100, G: 150, B: 255, A: 200},
			PulseSpeed:      6.0,
			BorderThickness: 1.5,
			GlowRadius:      5.0,
			ArrowSize:       10.0,
			EdgePadding:     10.0,
		}
	case "postapoc":
		s.style = GenreStyle{
			PrimaryColor:    color.RGBA{R: 255, G: 120, B: 0, A: 255},
			SecondaryColor:  color.RGBA{R: 180, G: 80, B: 0, A: 180},
			PulseSpeed:      4.0,
			BorderThickness: 2.0,
			GlowRadius:      6.0,
			ArrowSize:       12.0,
			EdgePadding:     8.0,
		}
	default: // fantasy
		s.style = GenreStyle{
			PrimaryColor:    color.RGBA{R: 255, G: 180, B: 0, A: 255},
			SecondaryColor:  color.RGBA{R: 255, G: 100, B: 50, A: 200},
			PulseSpeed:      5.0,
			BorderThickness: 2.0,
			GlowRadius:      6.0,
			ArrowSize:       12.0,
			EdgePadding:     8.0,
		}
	}
}

// Update processes threat component states.
func (s *System) Update(w *engine.World) {
	deltaTime := common.DeltaTime
	s.time += deltaTime

	compType := reflect.TypeOf(&Component{})
	entities := w.Query(compType)

	for _, eid := range entities {
		comp, ok := w.GetComponent(eid, compType)
		if !ok {
			continue
		}
		tc := comp.(*Component)

		// Decay threat over time
		if tc.ThreatDecay > 0 {
			tc.ThreatDecay -= deltaTime
			if tc.ThreatDecay <= 0 {
				tc.ThreatDecay = 0
				s.decreaseThreatLevel(tc)
			}
		}

		// Update pulse animation
		tc.PulsePhase += s.style.PulseSpeed * deltaTime
		if tc.PulsePhase > 2*math.Pi {
			tc.PulsePhase -= 2 * math.Pi
		}

		// Calculate border alpha based on threat level
		targetAlpha := s.threatLevelToAlpha(tc.ThreatLevel)
		tc.BorderAlpha = lerp(tc.BorderAlpha, targetAlpha, deltaTime*5.0)

		// Update windup progress
		if tc.AttackWindup {
			tc.WindupProgress += deltaTime * 2.0
			if tc.WindupProgress > 1.0 {
				tc.WindupProgress = 1.0
			}
		} else {
			tc.WindupProgress = lerp(tc.WindupProgress, 0, deltaTime*4.0)
		}
	}

	// Update off-screen indicators
	s.updateOffscreenIndicators(deltaTime)
}

// decreaseThreatLevel steps down the threat level by one.
func (s *System) decreaseThreatLevel(tc *Component) {
	if tc.IsBoss && tc.ThreatLevel <= ThreatMedium {
		return // Bosses stay at minimum medium
	}
	if tc.ThreatLevel > ThreatNone {
		tc.ThreatLevel--
	}
}

// threatLevelToAlpha converts threat level to target alpha.
func (s *System) threatLevelToAlpha(level ThreatLevel) float64 {
	switch level {
	case ThreatCritical:
		return 1.0
	case ThreatHigh:
		return 0.85
	case ThreatMedium:
		return 0.6
	case ThreatLow:
		return 0.35
	default:
		return 0.0
	}
}

// updateOffscreenIndicators fades and removes stale indicators.
func (s *System) updateOffscreenIndicators(deltaTime float64) {
	active := s.offscreenIndicators[:0]
	for i := range s.offscreenIndicators {
		ind := &s.offscreenIndicators[i]
		ind.PulsePhase += s.style.PulseSpeed * deltaTime
		if ind.PulsePhase > 2*math.Pi {
			ind.PulsePhase -= 2 * math.Pi
		}
		ind.Alpha -= deltaTime * 0.5 // Fade over 2 seconds
		if ind.Alpha > 0 {
			active = append(active, *ind)
		}
	}
	s.offscreenIndicators = active
}

// MarkThreat registers that an entity has become a threat to the player.
// Call this when an enemy deals damage or begins an attack.
func (s *System) MarkThreat(w *engine.World, entity engine.Entity, level ThreatLevel, duration float64) {
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if !ok {
		// Add component if not present
		tc := NewComponent()
		tc.ThreatLevel = level
		tc.ThreatDecay = duration
		w.AddComponent(entity, tc)
		return
	}

	tc := comp.(*Component)
	// Only increase threat level, don't decrease
	if level > tc.ThreatLevel {
		tc.ThreatLevel = level
	}
	// Reset decay timer
	if duration > tc.ThreatDecay {
		tc.ThreatDecay = duration
	}
	tc.LastDamageTime = s.time
}

// MarkAttackWindup indicates an entity is winding up an attack (for telegraph).
func (s *System) MarkAttackWindup(w *engine.World, entity engine.Entity, isWindup bool) {
	compType := reflect.TypeOf(&Component{})
	comp, ok := w.GetComponent(entity, compType)
	if !ok {
		return
	}
	tc := comp.(*Component)
	tc.AttackWindup = isWindup
	if !isWindup {
		tc.WindupProgress = 0
	}
}

// AddOffscreenThreat adds a directional indicator for an off-screen threat.
// angle is in radians (0 = right, π/2 = up, π = left, 3π/2 = down).
func (s *System) AddOffscreenThreat(angle, distance float64, level ThreatLevel) {
	if len(s.offscreenIndicators) >= s.maxOffscreen {
		// Replace oldest (lowest alpha)
		minIdx := 0
		minAlpha := s.offscreenIndicators[0].Alpha
		for i := 1; i < len(s.offscreenIndicators); i++ {
			if s.offscreenIndicators[i].Alpha < minAlpha {
				minAlpha = s.offscreenIndicators[i].Alpha
				minIdx = i
			}
		}
		s.offscreenIndicators[minIdx] = OffscreenIndicator{
			Angle:       angle,
			Distance:    distance,
			ThreatLevel: level,
			Alpha:       1.0,
			Color:       s.style.PrimaryColor,
			PulsePhase:  0,
		}
		return
	}

	s.offscreenIndicators = append(s.offscreenIndicators, OffscreenIndicator{
		Angle:       angle,
		Distance:    distance,
		ThreatLevel: level,
		Alpha:       1.0,
		Color:       s.style.PrimaryColor,
		PulsePhase:  0,
	})
}

// Render draws threat indicators for entities and off-screen threats.
func (s *System) Render(screen *ebiten.Image, w *engine.World, camX, camY float64) {
	s.renderOffscreenIndicators(screen)
	s.renderEntityThreatBorders(screen, w, camX, camY)
}

// renderOffscreenIndicators draws directional arrows at screen edges.
func (s *System) renderOffscreenIndicators(screen *ebiten.Image) {
	if len(s.offscreenIndicators) == 0 {
		return
	}

	centerX := float32(s.screenWidth) / 2
	centerY := float32(s.screenHeight) / 2
	padding := s.style.EdgePadding

	for i := range s.offscreenIndicators {
		ind := &s.offscreenIndicators[i]
		if ind.Alpha < 0.05 {
			continue
		}

		// Calculate position on screen edge
		angle := ind.Angle
		cos := float32(math.Cos(angle))
		sin := float32(math.Sin(angle))

		// Find edge intersection
		var edgeX, edgeY float32
		halfW := float32(s.screenWidth)/2 - padding
		halfH := float32(s.screenHeight)/2 - padding

		// Check which edge we hit first
		if math.Abs(float64(cos)) > 0.001 {
			t := halfW / float32(math.Abs(float64(cos)))
			testY := t * sin
			if math.Abs(float64(testY)) <= float64(halfH) {
				if cos > 0 {
					edgeX = centerX + halfW
				} else {
					edgeX = centerX - halfW
				}
				edgeY = centerY - testY
			} else {
				// Hit top/bottom edge
				t = halfH / float32(math.Abs(float64(sin)))
				edgeX = centerX + t*cos
				if sin > 0 {
					edgeY = centerY - halfH
				} else {
					edgeY = centerY + halfH
				}
			}
		} else {
			// Vertical
			edgeX = centerX
			if sin > 0 {
				edgeY = centerY - halfH
			} else {
				edgeY = centerY + halfH
			}
		}

		// Scale based on threat level and pulse
		baseSize := s.style.ArrowSize * (0.8 + 0.2*float32(ind.ThreatLevel)/float32(ThreatCritical))
		pulseScale := float32(1.0 + 0.15*math.Sin(ind.PulsePhase))
		size := baseSize * pulseScale

		// Draw arrow pointing toward threat
		s.drawThreatArrow(screen, edgeX, edgeY, angle, size, ind.Alpha, ind.ThreatLevel)
	}
}

// drawThreatArrow draws a triangular threat indicator.
func (s *System) drawThreatArrow(screen *ebiten.Image, x, y float32, angle float64, size float32, alpha float64, level ThreatLevel) {
	// Arrow points inward (toward center), so we reverse the angle
	inwardAngle := angle + math.Pi

	// Calculate triangle vertices
	cos := float32(math.Cos(inwardAngle))
	sin := float32(math.Sin(inwardAngle))

	// Tip of arrow (pointing inward)
	tipX := x + cos*size
	tipY := y - sin*size

	// Base vertices (perpendicular to direction)
	perpCos := float32(math.Cos(inwardAngle + math.Pi/2))
	perpSin := float32(math.Sin(inwardAngle + math.Pi/2))
	baseOffset := size * 0.6

	base1X := x + perpCos*baseOffset
	base1Y := y - perpSin*baseOffset
	base2X := x - perpCos*baseOffset
	base2Y := y + perpSin*baseOffset

	// Color with alpha
	col := s.style.PrimaryColor
	col.A = uint8(float64(col.A) * alpha)

	// Draw filled triangle using path
	var path vector.Path
	path.MoveTo(tipX, tipY)
	path.LineTo(base1X, base1Y)
	path.LineTo(base2X, base2Y)
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(col.R) / 255.0
		vs[i].ColorG = float32(col.G) / 255.0
		vs[i].ColorB = float32(col.B) / 255.0
		vs[i].ColorA = float32(col.A) / 255.0
	}

	screen.DrawTriangles(vs, is, emptyImage(), &ebiten.DrawTrianglesOptions{})

	// Draw pulsing outline for high threats
	if level >= ThreatHigh {
		outlineCol := s.style.SecondaryColor
		outlineCol.A = uint8(float64(outlineCol.A) * alpha * 0.7)
		vector.StrokeLine(screen, tipX, tipY, base1X, base1Y, 1.5, outlineCol, false)
		vector.StrokeLine(screen, base1X, base1Y, base2X, base2Y, 1.5, outlineCol, false)
		vector.StrokeLine(screen, base2X, base2Y, tipX, tipY, 1.5, outlineCol, false)
	}
}

// renderEntityThreatBorders draws pulsing borders around threatening entities.
func (s *System) renderEntityThreatBorders(screen *ebiten.Image, w *engine.World, camX, camY float64) {
	compType := reflect.TypeOf(&Component{})
	posType := reflect.TypeOf(&engine.Position{})

	entities := w.Query(compType, posType)

	for _, eid := range entities {
		tc, ok := w.GetComponent(eid, compType)
		if !ok {
			continue
		}
		threatComp := tc.(*Component)

		if threatComp.BorderAlpha < 0.05 {
			continue
		}

		posComp, ok := w.GetComponent(eid, posType)
		if !ok {
			continue
		}
		pos := posComp.(*engine.Position)

		// Convert world to screen coordinates
		screenX, screenY, visible := s.worldToScreen(pos.X, pos.Y, camX, camY)
		if !visible {
			continue
		}

		// Draw threat border
		s.drawEntityThreatBorder(screen, screenX, screenY, threatComp)
	}
}

// worldToScreen converts world coordinates to screen coordinates.
// Returns screen position and whether the entity is visible.
func (s *System) worldToScreen(worldX, worldY, camX, camY float64) (float32, float32, bool) {
	// Simple orthographic projection (for raycaster this would be different)
	// This is a placeholder - the actual game uses raycasting
	dx := worldX - camX
	dy := worldY - camY
	dist := math.Sqrt(dx*dx + dy*dy)

	// Check if within view distance
	if dist > 20 {
		return 0, 0, false
	}

	// Convert to screen space (simplified)
	screenX := float32(s.screenWidth/2) + float32(dx*20)
	screenY := float32(s.screenHeight/2) + float32(dy*20)

	// Check screen bounds
	margin := float32(30)
	if screenX < -margin || screenX > float32(s.screenWidth)+margin ||
		screenY < -margin || screenY > float32(s.screenHeight)+margin {
		return 0, 0, false
	}

	return screenX, screenY, true
}

// drawEntityThreatBorder draws the pulsing threat border around an entity.
func (s *System) drawEntityThreatBorder(screen *ebiten.Image, x, y float32, tc *Component) {
	// Base size scales with threat level
	baseSize := float32(16) + float32(tc.ThreatLevel)*4

	// Pulse effect
	pulseScale := float32(1.0 + 0.1*math.Sin(tc.PulsePhase))
	size := baseSize * pulseScale

	// Calculate alpha
	alpha := tc.BorderAlpha * (0.7 + 0.3*math.Sin(tc.PulsePhase))

	// Get colors
	col := s.style.PrimaryColor
	col.A = uint8(float64(col.A) * alpha)

	thickness := s.style.BorderThickness

	// Draw rectangular border
	halfSize := size / 2
	left := x - halfSize
	right := x + halfSize
	top := y - halfSize
	bottom := y + halfSize

	// Top edge
	vector.StrokeLine(screen, left, top, right, top, thickness, col, false)
	// Bottom edge
	vector.StrokeLine(screen, left, bottom, right, bottom, thickness, col, false)
	// Left edge
	vector.StrokeLine(screen, left, top, left, bottom, thickness, col, false)
	// Right edge
	vector.StrokeLine(screen, right, top, right, bottom, thickness, col, false)

	// Corner highlights for high threats
	if tc.ThreatLevel >= ThreatHigh {
		cornerSize := size * 0.2
		cornerCol := s.style.SecondaryColor
		cornerCol.A = uint8(float64(cornerCol.A) * alpha)

		// Top-left corner
		vector.StrokeLine(screen, left, top, left+cornerSize, top, thickness*1.5, cornerCol, false)
		vector.StrokeLine(screen, left, top, left, top+cornerSize, thickness*1.5, cornerCol, false)

		// Top-right corner
		vector.StrokeLine(screen, right, top, right-cornerSize, top, thickness*1.5, cornerCol, false)
		vector.StrokeLine(screen, right, top, right, top+cornerSize, thickness*1.5, cornerCol, false)

		// Bottom-left corner
		vector.StrokeLine(screen, left, bottom, left+cornerSize, bottom, thickness*1.5, cornerCol, false)
		vector.StrokeLine(screen, left, bottom, left, bottom-cornerSize, thickness*1.5, cornerCol, false)

		// Bottom-right corner
		vector.StrokeLine(screen, right, bottom, right-cornerSize, bottom, thickness*1.5, cornerCol, false)
		vector.StrokeLine(screen, right, bottom, right, bottom-cornerSize, thickness*1.5, cornerCol, false)
	}

	// Attack windup indicator (filling circle)
	if tc.WindupProgress > 0.01 {
		s.drawWindupIndicator(screen, x, y-halfSize-8, tc.WindupProgress)
	}
}

// drawWindupIndicator draws a filling arc showing attack telegraph.
func (s *System) drawWindupIndicator(screen *ebiten.Image, x, y float32, progress float64) {
	radius := float32(6)
	col := s.style.SecondaryColor
	col.A = 200

	// Background circle
	bgCol := color.RGBA{R: 40, G: 40, B: 40, A: 180}
	drawCircle(screen, x, y, radius, bgCol)

	// Progress arc
	if progress > 0 {
		drawArc(screen, x, y, radius-1, progress, col)
	}
}

// drawCircle draws a filled circle.
func drawCircle(screen *ebiten.Image, x, y, radius float32, col color.RGBA) {
	const segments = 16
	var path vector.Path
	for i := 0; i <= segments; i++ {
		angle := float64(i) * 2 * math.Pi / float64(segments)
		px := x + radius*float32(math.Cos(angle))
		py := y + radius*float32(math.Sin(angle))
		if i == 0 {
			path.MoveTo(px, py)
		} else {
			path.LineTo(px, py)
		}
	}
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(col.R) / 255.0
		vs[i].ColorG = float32(col.G) / 255.0
		vs[i].ColorB = float32(col.B) / 255.0
		vs[i].ColorA = float32(col.A) / 255.0
	}
	screen.DrawTriangles(vs, is, emptyImage(), &ebiten.DrawTrianglesOptions{})
}

// drawArc draws a filled arc (pie slice) for progress indication.
func drawArc(screen *ebiten.Image, x, y, radius float32, progress float64, col color.RGBA) {
	if progress <= 0 {
		return
	}
	if progress > 1 {
		progress = 1
	}

	// Start from top (-π/2) and go clockwise
	startAngle := -math.Pi / 2
	endAngle := startAngle + progress*2*math.Pi

	const segments = 16
	segmentCount := int(float64(segments) * progress)
	if segmentCount < 2 {
		segmentCount = 2
	}

	var path vector.Path
	path.MoveTo(x, y) // Center

	for i := 0; i <= segmentCount; i++ {
		t := float64(i) / float64(segmentCount)
		angle := startAngle + t*(endAngle-startAngle)
		px := x + radius*float32(math.Cos(angle))
		py := y + radius*float32(math.Sin(angle))
		path.LineTo(px, py)
	}
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	for i := range vs {
		vs[i].ColorR = float32(col.R) / 255.0
		vs[i].ColorG = float32(col.G) / 255.0
		vs[i].ColorB = float32(col.B) / 255.0
		vs[i].ColorA = float32(col.A) / 255.0
	}
	screen.DrawTriangles(vs, is, emptyImage(), &ebiten.DrawTrianglesOptions{})
}

// lerp performs linear interpolation.
func lerp(a, b, t float64) float64 {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return a + (b-a)*t
}

// emptyImage returns a 1x1 white image for drawing triangles.
var emptyImg *ebiten.Image

func emptyImage() *ebiten.Image {
	if emptyImg == nil {
		emptyImg = ebiten.NewImage(1, 1)
		emptyImg.Fill(color.White)
	}
	return emptyImg
}

// GetStyle returns the current genre style (for testing/debugging).
func (s *System) GetStyle() GenreStyle {
	return s.style
}

// GetOffscreenIndicatorCount returns the number of active off-screen indicators.
func (s *System) GetOffscreenIndicatorCount() int {
	return len(s.offscreenIndicators)
}

// ClearOffscreenIndicators removes all off-screen indicators.
func (s *System) ClearOffscreenIndicators() {
	s.offscreenIndicators = s.offscreenIndicators[:0]
}
